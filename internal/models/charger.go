package models

import "time"

type ChargerStatus string

const (
	ChargerAvailable   ChargerStatus = "available"
	ChargerPreparing   ChargerStatus = "preparing"
	ChargerCharging    ChargerStatus = "charging"
	ChargerFinishing   ChargerStatus = "finishing"
	ChargerFaulted     ChargerStatus = "faulted"
	ChargerOffline     ChargerStatus = "offline"
	ChargerUnavailable ChargerStatus = "unavailable"
)

type Charger struct {
	ID              string        `json:"id" db:"id"`
	StationID       string        `json:"station_id" db:"station_id"`
	OCPPID          string        `json:"ocpp_id" db:"ocpp_id"`
	Vendor          *string       `json:"vendor,omitempty" db:"vendor"`
	Model           *string       `json:"model,omitempty" db:"model"`
	FirmwareVersion *string       `json:"firmware_version,omitempty" db:"firmware_version"`
	ConnectorCount  int           `json:"connector_count" db:"connector_count"`
	Status          ChargerStatus `json:"status" db:"status"`
	LastSeenAt      *time.Time    `json:"last_seen_at,omitempty" db:"last_seen_at"`
	IsActive        bool          `json:"is_active" db:"is_active"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}