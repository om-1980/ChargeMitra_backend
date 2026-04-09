package ocppcore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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