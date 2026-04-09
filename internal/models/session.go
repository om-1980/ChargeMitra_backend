package models

import "time"

type SessionStatus string

const (
	SessionRequested  SessionStatus = "requested"
	SessionAuthorized SessionStatus = "authorized"
	SessionInProgress SessionStatus = "in_progress"
	SessionCompleted  SessionStatus = "completed"
	SessionStopped    SessionStatus = "stopped"
	SessionFailed     SessionStatus = "failed"
	SessionCancelled  SessionStatus = "cancelled"
)

type ChargingSession struct {
	ID         string        `json:"id" db:"id"`
	UserID     *string       `json:"user_id,omitempty" db:"user_id"`
	ChargerID  string        `json:"charger_id" db:"charger_id"`
	StationID  string        `json:"station_id" db:"station_id"`
	Status     SessionStatus `json:"status" db:"status"`
	MeterStart *float64      `json:"meter_start,omitempty" db:"meter_start"`
	MeterStop  *float64      `json:"meter_stop,omitempty" db:"meter_stop"`
	EnergyKwh  float64       `json:"energy_kwh" db:"energy_kwh"`
	Amount     float64       `json:"amount" db:"amount"`
	StartedAt  *time.Time    `json:"started_at,omitempty" db:"started_at"`
	EndedAt    *time.Time    `json:"ended_at,omitempty" db:"ended_at"`
	StopReason *string       `json:"stop_reason,omitempty" db:"stop_reason"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}