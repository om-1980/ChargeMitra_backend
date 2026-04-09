package chargers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(req CreateChargerRequest) (*ChargerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check station exists
	var stationExists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM stations WHERE id = $1 AND is_active = true
		)
	`, req.StationID).Scan(&stationExists)
	if err != nil {
		return nil, err
	}
	if !stationExists {
		return nil, errors.New("station not found")
	}

	// Check ocpp id uniqueness
	var chargerExists bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM chargers WHERE LOWER(ocpp_id) = LOWER($1)
		)
	`, strings.TrimSpace(req.OCPPID)).Scan(&chargerExists)
	if err != nil {
		return nil, err
	}
	if chargerExists {
		return nil, errors.New("charger with this ocpp_id already exists")
	}

	connectorCount := 1
	if req.ConnectorCount != nil && *req.ConnectorCount > 0 {
		connectorCount = *req.ConnectorCount
	}

	var charger ChargerResponse
	var lastSeenAt *time.Time
	var createdAt time.Time
	var updatedAt time.Time

	err = s.db.QueryRow(ctx, `
		INSERT INTO chargers (
			station_id, ocpp_id, vendor, model, firmware_version, connector_count
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, station_id, ocpp_id, vendor, model, firmware_version,
		          connector_count, status, last_seen_at, is_active, created_at, updated_at
	`,
		req.StationID,
		strings.TrimSpace(req.OCPPID),
		req.Vendor,
		req.Model,
		req.FirmwareVersion,
		connectorCount,
	).Scan(
		&charger.ID,
		&charger.StationID,
		&charger.OCPPID,
		&charger.Vendor,
		&charger.Model,
		&charger.FirmwareVersion,
		&charger.ConnectorCount,
		&charger.Status,
		&lastSeenAt,
		&charger.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if lastSeenAt != nil {
		formatted := lastSeenAt.Format(time.RFC3339)
		charger.LastSeenAt = &formatted
	}

	charger.CreatedAt = createdAt.Format(time.RFC3339)
	charger.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &charger, nil
}

func (s *Service) List() ([]ChargerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, station_id, ocpp_id, vendor, model, firmware_version,
		       connector_count, status, last_seen_at, is_active, created_at, updated_at
		FROM chargers
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chargers := make([]ChargerResponse, 0)

	for rows.Next() {
		var charger ChargerResponse
		var lastSeenAt *time.Time
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(
			&charger.ID,
			&charger.StationID,
			&charger.OCPPID,
			&charger.Vendor,
			&charger.Model,
			&charger.FirmwareVersion,
			&charger.ConnectorCount,
			&charger.Status,
			&lastSeenAt,
			&charger.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastSeenAt != nil {
			formatted := lastSeenAt.Format(time.RFC3339)
			charger.LastSeenAt = &formatted
		}

		charger.CreatedAt = createdAt.Format(time.RFC3339)
		charger.UpdatedAt = updatedAt.Format(time.RFC3339)

		chargers = append(chargers, charger)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chargers, nil
}

func (s *Service) ListByStation(stationID string) ([]ChargerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, station_id, ocpp_id, vendor, model, firmware_version,
		       connector_count, status, last_seen_at, is_active, created_at, updated_at
		FROM chargers
		WHERE station_id = $1
		ORDER BY created_at DESC
	`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chargers := make([]ChargerResponse, 0)

	for rows.Next() {
		var charger ChargerResponse
		var lastSeenAt *time.Time
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(
			&charger.ID,
			&charger.StationID,
			&charger.OCPPID,
			&charger.Vendor,
			&charger.Model,
			&charger.FirmwareVersion,
			&charger.ConnectorCount,
			&charger.Status,
			&lastSeenAt,
			&charger.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastSeenAt != nil {
			formatted := lastSeenAt.Format(time.RFC3339)
			charger.LastSeenAt = &formatted
		}

		charger.CreatedAt = createdAt.Format(time.RFC3339)
		charger.UpdatedAt = updatedAt.Format(time.RFC3339)

		chargers = append(chargers, charger)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chargers, nil
}