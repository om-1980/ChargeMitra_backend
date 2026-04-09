package stations

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ownerID string, req CreateStationRequest) (*StationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	country := "India"
	if req.Country != nil && *req.Country != "" {
		country = *req.Country
	}

	var station StationResponse
	var createdAt time.Time
	var updatedAt time.Time

	err := s.db.QueryRow(ctx, `
		INSERT INTO stations (
			owner_id, name, address_line1, address_line2,
			city, district, state, country, pincode,
			latitude, longitude
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, owner_id, name, address_line1, address_line2,
		          city, district, state, country, pincode,
		          latitude, longitude, is_active, created_at, updated_at
	`,
		ownerID,
		req.Name,
		req.AddressLine1,
		req.AddressLine2,
		req.City,
		req.District,
		req.State,
		country,
		req.Pincode,
		req.Latitude,
		req.Longitude,
	).Scan(
		&station.ID,
		&station.OwnerID,
		&station.Name,
		&station.AddressLine1,
		&station.AddressLine2,
		&station.City,
		&station.District,
		&station.State,
		&station.Country,
		&station.Pincode,
		&station.Latitude,
		&station.Longitude,
		&station.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	station.CreatedAt = createdAt.Format(time.RFC3339)
	station.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &station, nil
}

func (s *Service) List() ([]StationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, owner_id, name, address_line1, address_line2,
		       city, district, state, country, pincode,
		       latitude, longitude, is_active, created_at, updated_at
		FROM stations
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stations := make([]StationResponse, 0)

	for rows.Next() {
		var station StationResponse
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(
			&station.ID,
			&station.OwnerID,
			&station.Name,
			&station.AddressLine1,
			&station.AddressLine2,
			&station.City,
			&station.District,
			&station.State,
			&station.Country,
			&station.Pincode,
			&station.Latitude,
			&station.Longitude,
			&station.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		station.CreatedAt = createdAt.Format(time.RFC3339)
		station.UpdatedAt = updatedAt.Format(time.RFC3339)

		stations = append(stations, station)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stations, nil
}

func (s *Service) ListByOwner(ownerID string) ([]StationResponse, error) {
	if ownerID == "" {
		return nil, errors.New("owner id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, owner_id, name, address_line1, address_line2,
		       city, district, state, country, pincode,
		       latitude, longitude, is_active, created_at, updated_at
		FROM stations
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stations := make([]StationResponse, 0)

	for rows.Next() {
		var station StationResponse
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(
			&station.ID,
			&station.OwnerID,
			&station.Name,
			&station.AddressLine1,
			&station.AddressLine2,
			&station.City,
			&station.District,
			&station.State,
			&station.Country,
			&station.Pincode,
			&station.Latitude,
			&station.Longitude,
			&station.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		station.CreatedAt = createdAt.Format(time.RFC3339)
		station.UpdatedAt = updatedAt.Format(time.RFC3339)

		stations = append(stations, station)
	}

	return stations, nil
}