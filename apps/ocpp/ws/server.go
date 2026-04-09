package ws

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/om-1980/ChargeMitra_backend/internal/ocppcore"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
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

	defer func() {
		s.service.Hub().Remove(ocppID)
		_ = s.service.MarkOffline(ocppID)
		_ = client.Close()
		s.log.Infof("charger disconnected: ocpp_id=%s", ocppID)
	}()

	conn.SetReadLimit(1024 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return s.service.TouchHeartbeat(ocppID)
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			s.log.Errorf("read failed for %s: %v", ocppID, err)
			break
		}

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

		call, err := ocppcore.ParseOCPPCall(message)
		if err != nil {
			s.log.Errorf("failed to parse OCPP frame from %s: %v", ocppID, err)

			errorFrame, buildErr := ocppcore.BuildCallError(
				"",
				"FormationViolation",
				"invalid OCPP CALL frame",
				nil,
			)
			if buildErr == nil {
				_ = client.WriteMessage(websocket.TextMessage, errorFrame)
			}
			continue
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
				_ = client.WriteMessage(websocket.TextMessage, errorFrame)
			}
			continue
		}

		if err := client.WriteMessage(websocket.TextMessage, responseFrame); err != nil {
			s.log.Errorf("failed to write OCPP response for %s: %v", ocppID, err)
			break
		}
	}
}