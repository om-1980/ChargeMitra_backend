package ocppcore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	db    *pgxpool.Pool
	redis *redis.Client
	hub   *Hub
}

func NewService(db *pgxpool.Pool, redisClient *redis.Client, hub *Hub) *Service {
	return &Service{
		db:    db,
		redis: redisClient,
		hub:   hub,
	}
}

func (s *Service) ValidateCharger(ocppID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM chargers WHERE ocpp_id = $1 AND is_active = true
		)
	`, ocppID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("charger not registered")
	}

	return nil
}

func (s *Service) MarkOnline(ocppID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		UPDATE chargers
		SET status = 'available',
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID)
	if err != nil {
		return err
	}

	if s.redis != nil {
		key := fmt.Sprintf("charger:%s:online", ocppID)
		_ = s.redis.Set(ctx, key, "1", 24*time.Hour).Err()
	}

	return nil
}

func (s *Service) MarkOffline(ocppID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		UPDATE chargers
		SET status = 'offline',
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID)
	if err != nil {
		return err
	}

	if s.redis != nil {
		key := fmt.Sprintf("charger:%s:online", ocppID)
		_ = s.redis.Del(ctx, key).Err()
	}

	return nil
}

func (s *Service) TouchHeartbeat(ocppID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		UPDATE chargers
		SET last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID)
	return err
}

func (s *Service) LogIncomingMessage(ocppID string, raw []byte) error {
	if s.redis == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload := IncomingMessage{
		OCPPID:     ocppID,
		Message:    string(raw),
		ReceivedAt: time.Now(),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("ocpp:logs:%s", ocppID)
	return s.redis.LPush(ctx, key, b).Err()
}

func (s *Service) GetChargerIDByOCPPID(ocppID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chargerID string
	err := s.db.QueryRow(ctx, `SELECT id FROM chargers WHERE ocpp_id = $1`, ocppID).Scan(&chargerID)
	return chargerID, err
}

func (s *Service) GetStationIDByOCPPID(ocppID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stationID string
	err := s.db.QueryRow(ctx, `SELECT station_id FROM chargers WHERE ocpp_id = $1`, ocppID).Scan(&stationID)
	return stationID, err
}

func (s *Service) Hub() *Hub {
	return s.hub
}

func (s *Service) SaveOCPPEvent(ocppID, direction, action, messageID string, payload interface{}, rawMessage string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chargerID *string
	var foundChargerID string

	err := s.db.QueryRow(ctx, `
		SELECT id
		FROM chargers
		WHERE ocpp_id = $1
		LIMIT 1
	`, ocppID).Scan(&foundChargerID)
	if err == nil {
		chargerID = &foundChargerID
	}

	var payloadJSON []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadJSON = b
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO ocpp_events (
			charger_id,
			ocpp_id,
			direction,
			action,
			message_id,
			payload,
			raw_message,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
	`,
		chargerID,
		ocppID,
		direction,
		nullIfEmpty(action),
		nullIfEmpty(messageID),
		payloadJSON,
		nullIfEmpty(rawMessage),
	)
	return err
}

func (s *Service) ListOCPPEventsAdvanced(
	filters map[string]string,
	limit, offset int,
) ([]map[string]interface{}, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, ocpp_id, direction, action, message_id, payload, raw_message, created_at
		FROM ocpp_events
		WHERE 1=1
	`
	args := []interface{}{}
	i := 1

	if v := filters["ocpp_id"]; v != "" {
		query += fmt.Sprintf(" AND ocpp_id = $%d", i)
		args = append(args, v)
		i++
	}

	if v := filters["direction"]; v != "" {
		query += fmt.Sprintf(" AND direction = $%d", i)
		args = append(args, v)
		i++
	}

	if v := filters["action"]; v != "" {
		query += fmt.Sprintf(" AND action ILIKE $%d", i)
		args = append(args, "%"+v+"%")
		i++
	}

	if v := filters["message_id"]; v != "" {
		query += fmt.Sprintf(" AND message_id = $%d", i)
		args = append(args, v)
		i++
	}

	if v := filters["from"]; v != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", i)
		args = append(args, v)
		i++
	}

	if v := filters["to"]; v != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", i)
		args = append(args, v)
		i++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var id, ocppID, direction string
		var action, messageID *string
		var payloadBytes []byte
		var rawMessage *string
		var createdAt time.Time

		err := rows.Scan(
			&id,
			&ocppID,
			&direction,
			&action,
			&messageID,
			&payloadBytes,
			&rawMessage,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		var payload interface{}
		if payloadBytes != nil {
			_ = json.Unmarshal(payloadBytes, &payload)
		}

		results = append(results, map[string]interface{}{
			"id":          id,
			"ocpp_id":     ocppID,
			"direction":   direction,
			"action":      action,
			"message_id":  messageID,
			"payload":     payload,
			"raw_message": rawMessage,
			"created_at":  createdAt,
		})
	}

	return results, nil
}

func (s *Service) GetOCPPSummary(ocppID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int
	var incoming int
	var outgoing int

	err := s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE direction='incoming'),
			COUNT(*) FILTER (WHERE direction='outgoing')
		FROM ocpp_events
		WHERE ocpp_id = $1
	`, ocppID).Scan(&total, &incoming, &outgoing)
	if err != nil {
		return nil, err
	}

	var lastBoot, lastHeartbeat *time.Time

	_ = s.db.QueryRow(ctx, `
		SELECT created_at FROM ocpp_events 
		WHERE ocpp_id=$1 AND action='BootNotification'
		ORDER BY created_at DESC LIMIT 1
	`, ocppID).Scan(&lastBoot)

	_ = s.db.QueryRow(ctx, `
		SELECT created_at FROM ocpp_events 
		WHERE ocpp_id=$1 AND action='Heartbeat'
		ORDER BY created_at DESC LIMIT 1
	`, ocppID).Scan(&lastHeartbeat)

	return map[string]interface{}{
		"ocpp_id":        ocppID,
		"total_events":   total,
		"incoming":       incoming,
		"outgoing":       outgoing,
		"last_boot":      lastBoot,
		"last_heartbeat": lastHeartbeat,
	}, nil
}

func (s *Service) RemoteStartTransaction(ocppID, idTag string, connectorID *int) (string, error) {
	if s.hub == nil {
		return "", fmt.Errorf("hub not available")
	}

	client, ok := s.hub.Get(ocppID)
	if !ok || client == nil || client.IsClosed() {
		return "", fmt.Errorf("charger not connected")
	}

	messageID := fmt.Sprintf("rstart-%d", time.Now().UnixNano())

	payload := map[string]interface{}{
		"idTag": idTag,
	}
	if connectorID != nil {
		payload["connectorId"] = *connectorID
	}

	data, err := BuildCall(messageID, "RemoteStartTransaction", payload)
	if err != nil {
		return "", err
	}

	if err := s.SaveOCPPEvent(ocppID, "outgoing", "RemoteStartTransaction.req", messageID, payload, string(data)); err != nil {
		return "", err
	}

	if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
		if s.hub != nil {
			s.hub.Remove(ocppID)
		}
		_ = s.MarkOffline(ocppID)
		return "", fmt.Errorf("failed to send remote start: %w", err)
	}

	return messageID, nil
}

func (s *Service) RemoteStopTransaction(ocppID string, transactionID int64) (string, error) {
	if s.hub == nil {
		return "", fmt.Errorf("hub not available")
	}

	client, ok := s.hub.Get(ocppID)
	if !ok || client == nil || client.IsClosed() {
		return "", fmt.Errorf("charger not connected")
	}

	messageID := fmt.Sprintf("rstop-%d", time.Now().UnixNano())

	payload := map[string]interface{}{
		"transactionId": transactionID,
	}

	data, err := BuildCall(messageID, "RemoteStopTransaction", payload)
	if err != nil {
		return "", err
	}

	if err := s.SaveOCPPEvent(ocppID, "outgoing", "RemoteStopTransaction.req", messageID, payload, string(data)); err != nil {
		return "", err
	}

	if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
		if s.hub != nil {
			s.hub.Remove(ocppID)
		}
		_ = s.MarkOffline(ocppID)
		return "", fmt.Errorf("failed to send remote stop: %w", err)
	}

	return messageID, nil
}

func (s *Service) ResolveResponseAction(ocppID, messageID, fallback string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var action string
	err := s.db.QueryRow(ctx, `
		SELECT action
		FROM ocpp_events
		WHERE ocpp_id = $1
		  AND message_id = $2
		  AND direction = 'outgoing'
		ORDER BY created_at DESC
		LIMIT 1
	`, ocppID, messageID).Scan(&action)
	if err != nil {
		return fallback
	}

	if strings.HasSuffix(action, ".req") {
		return strings.TrimSuffix(action, ".req") + ".conf"
	}

	if strings.HasSuffix(action, ".error") {
		return action
	}

	if strings.HasSuffix(action, ".conf") {
		return action
	}

	return action + ".conf"
}

func (s *Service) ResolveErrorAction(ocppID, messageID, fallback string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var action string
	err := s.db.QueryRow(ctx, `
		SELECT action
		FROM ocpp_events
		WHERE ocpp_id = $1
		  AND message_id = $2
		  AND direction = 'outgoing'
		ORDER BY created_at DESC
		LIMIT 1
	`, ocppID, messageID).Scan(&action)
	if err != nil {
		return fallback
	}

	if strings.HasSuffix(action, ".req") {
		return strings.TrimSuffix(action, ".req") + ".error"
	}

	if strings.HasSuffix(action, ".conf") {
		return strings.TrimSuffix(action, ".conf") + ".error"
	}

	if strings.HasSuffix(action, ".error") {
		return action
	}

	return action + ".error"
}