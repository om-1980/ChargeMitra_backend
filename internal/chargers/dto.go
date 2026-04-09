package chargers

type CreateChargerRequest struct {
	StationID       string  `json:"station_id" binding:"required,uuid"`
	OCPPID          string  `json:"ocpp_id" binding:"required,min=2,max=100"`
	Vendor          *string `json:"vendor,omitempty"`
	Model           *string `json:"model,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
	ConnectorCount  *int    `json:"connector_count,omitempty"`
}

type ChargerResponse struct {
	ID              string  `json:"id"`
	StationID       string  `json:"station_id"`
	OCPPID          string  `json:"ocpp_id"`
	Vendor          *string `json:"vendor,omitempty"`
	Model           *string `json:"model,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
	ConnectorCount  int     `json:"connector_count"`
	Status          string  `json:"status"`
	LastSeenAt      *string `json:"last_seen_at,omitempty"`
	IsActive        bool    `json:"is_active"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}