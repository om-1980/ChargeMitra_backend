package ws

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/om-1980/ChargeMitra_backend/internal/ocppcore"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
)

const (
	readWait       = 180 * time.Second
	writeWait      = 10 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 1024 * 1024
)

type Server struct {
	log      *logger.Logger
	service  *ocppcore.Service
	upgrader websocket.Upgrader
}

func NewServer(log *logger.Logger, service *ocppcore.Service) *Server {
	return &Server{
		log:     log,
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) HandleChargePoint(c *gin.Context) {
	ocppID := strings.TrimSpace(c.Param("ocppId"))
	if ocppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ocppId is required"})
		return
	}

	if err := s.service.ValidateCharger(ocppID); err != nil {
		s.log.Errorf("charger validation failed for %s: %v", ocppID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "charger not registered"})
		return
	}

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.Errorf("websocket upgrade failed for %s: %v", ocppID, err)
		return
	}

	client := ocppcore.NewClient(ocppID, conn, c.Request.RemoteAddr)
	s.service.Hub().Add(client)

	if err := s.service.MarkOnline(ocppID); err != nil {
		s.log.Errorf("failed to mark charger online %s: %v", ocppID, err)
	}

	s.log.Infof("charger connected: ocpp_id=%s remote=%s", ocppID, c.Request.RemoteAddr)

	done := make(chan struct{})
	stopPinger := func() {
		select {
		case <-done:
		default:
			close(done)
		}
	}

	defer func() {
		stopPinger()
		s.service.Hub().Remove(ocppID)
		_ = s.service.MarkOffline(ocppID)
		_ = writeCloseFrame(conn)
		_ = client.Close()
		s.log.Infof("charger disconnected: ocpp_id=%s", ocppID)
	}()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(readWait))

	conn.SetPongHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readWait))
		if err := s.service.TouchHeartbeat(ocppID); err != nil {
			s.log.Errorf("failed to update heartbeat on pong for %s: %v", ocppID, err)
		}
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readWait))
		if err := s.service.TouchHeartbeat(ocppID); err != nil {
			s.log.Errorf("failed to update heartbeat on ping for %s: %v", ocppID, err)
		}

		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeWait))
	})

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(writeWait)); err != nil {
					s.log.Errorf("failed to send ping to %s: %v", ocppID, err)
					_ = conn.Close()
					return
				}
			}
		}
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			stopPinger()

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				s.log.Errorf("read failed for %s: %v", ocppID, err)
			} else {
				s.log.Infof("connection closed for %s: %v", ocppID, err)
			}
			break
		}

		_ = conn.SetReadDeadline(time.Now().Add(readWait))

		if messageType != websocket.TextMessage {
			s.log.Errorf("unsupported websocket message type for %s: %d", ocppID, messageType)
			continue
		}

		if err := s.service.TouchHeartbeat(ocppID); err != nil {
			s.log.Errorf("failed to update heartbeat for %s: %v", ocppID, err)
		}

		if err := s.service.LogIncomingMessage(ocppID, message); err != nil {
			s.log.Errorf("failed to log incoming message for %s: %v", ocppID, err)
		}

		s.log.Infof("incoming OCPP frame from %s: %s", ocppID, string(message))

		msg, err := ocppcore.ParseOCPPMessage(message)
		if err != nil {
			_ = s.service.SaveOCPPEvent(
				ocppID,
				"incoming",
				"PARSE_ERROR",
				"",
				nil,
				string(message),
			)

			s.log.Errorf("failed to parse OCPP frame from %s: %v", ocppID, err)

			errorFrame, buildErr := ocppcore.BuildCallError(
				"",
				"FormationViolation",
				"invalid OCPP frame",
				nil,
			)
			if buildErr == nil {
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				_ = client.WriteMessage(websocket.TextMessage, errorFrame)
			}
			continue
		}

		switch msg.MessageType {
		case ocppcore.MessageTypeCall:
			var parsedPayload interface{}
			_ = json.Unmarshal(msg.Payload, &parsedPayload)

			_ = s.service.SaveOCPPEvent(
				ocppID,
				"incoming",
				msg.Action,
				msg.MessageID,
				parsedPayload,
				string(message),
			)

			call := &ocppcore.OCPPCall{
				MessageType: msg.MessageType,
				MessageID:   msg.MessageID,
				Action:      msg.Action,
				Payload:     msg.Payload,
			}

			responseFrame, err := s.service.HandleCall(ocppID, call)
			if err != nil {
				s.log.Errorf("failed to handle OCPP action %s for %s: %v", call.Action, ocppID, err)

				errorFrame, buildErr := ocppcore.BuildCallError(
					call.MessageID,
					"InternalError",
					"failed to process request",
					nil,
				)
				if buildErr == nil {
					_ = s.service.SaveOCPPEvent(
						ocppID,
						"outgoing",
						call.Action+".error",
						call.MessageID,
						nil,
						string(errorFrame),
					)

					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					_ = client.WriteMessage(websocket.TextMessage, errorFrame)
				}
				continue
			}

			_ = s.service.SaveOCPPEvent(
				ocppID,
				"outgoing",
				call.Action+".conf",
				call.MessageID,
				nil,
				string(responseFrame),
			)

			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.WriteMessage(websocket.TextMessage, responseFrame); err != nil {
				s.log.Errorf("failed to write OCPP response for %s: %v", ocppID, err)
				stopPinger()
				break
			}

		case ocppcore.MessageTypeCallResult:
			var resultPayload interface{}
			_ = json.Unmarshal(msg.Payload, &resultPayload)

			resolvedAction := s.service.ResolveResponseAction(ocppID, msg.MessageID, "CALLRESULT")

			_ = s.service.SaveOCPPEvent(
				ocppID,
				"incoming",
				resolvedAction,
				msg.MessageID,
				resultPayload,
				string(message),
			)

			s.log.Infof("incoming CALLRESULT from %s: message_id=%s resolved_action=%s",
				ocppID, msg.MessageID, resolvedAction)

		case ocppcore.MessageTypeCallError:
			payload := map[string]interface{}{
				"error_code":        msg.ErrorCode,
				"error_description": msg.ErrorDescription,
				"details":           msg.ErrorDetails,
			}

			resolvedAction := s.service.ResolveErrorAction(ocppID, msg.MessageID, "CALLERROR")

			_ = s.service.SaveOCPPEvent(
				ocppID,
				"incoming",
				resolvedAction,
				msg.MessageID,
				payload,
				string(message),
			)

			s.log.Errorf("incoming CALLERROR from %s: message_id=%s resolved_action=%s error_code=%s description=%s",
				ocppID, msg.MessageID, resolvedAction, msg.ErrorCode, msg.ErrorDescription)
		}
	}
}

func writeCloseFrame(conn *websocket.Conn) error {
	_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing connection"),
		time.Now().Add(writeWait),
	)
}