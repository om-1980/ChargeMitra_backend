package ocppcore

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeCall       = 2
	MessageTypeCallResult = 3
	MessageTypeCallError  = 4
)

type IncomingMessage struct {
	OCPPID     string    `json:"ocpp_id"`
	Message    string    `json:"message"`
	ReceivedAt time.Time `json:"received_at"`
}

type ChargerConnectionInfo struct {
	OCPPID      string    `json:"ocpp_id"`
	ConnectedAt time.Time `json:"connected_at"`
	RemoteAddr  string    `json:"remote_addr"`
	IsOnline    bool      `json:"is_online"`
}

type OCPPCall struct {
	MessageType int
	MessageID   string
	Action      string
	Payload     json.RawMessage
}

type BootNotificationRequest struct {
	ChargePointVendor string `json:"chargePointVendor"`
	ChargePointModel  string `json:"chargePointModel"`
	ChargePointSerial string `json:"chargePointSerialNumber,omitempty"`
	FirmwareVersion   string `json:"firmwareVersion,omitempty"`
	ChargeBoxSerial   string `json:"chargeBoxSerialNumber,omitempty"`
	Iccid             string `json:"iccid,omitempty"`
	Imsi              string `json:"imsi,omitempty"`
	MeterType         string `json:"meterType,omitempty"`
	MeterSerialNumber string `json:"meterSerialNumber,omitempty"`
}

type BootNotificationResponse struct {
	Status      string `json:"status"`
	CurrentTime string `json:"currentTime"`
	Interval    int    `json:"interval"`
}

type HeartbeatResponse struct {
	CurrentTime string `json:"currentTime"`
}

type StatusNotificationRequest struct {
	ConnectorID int    `json:"connectorId"`
	ErrorCode   string `json:"errorCode"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp,omitempty"`
	Info        string `json:"info,omitempty"`
	VendorID    string `json:"vendorId,omitempty"`
	VendorError string `json:"vendorErrorCode,omitempty"`
}

type AuthorizeRequest struct {
	IDTag string `json:"idTag"`
}

type AuthorizeResponse struct {
	IDTagInfo IDTagInfo `json:"idTagInfo"`
}

type IDTagInfo struct {
	Status      string `json:"status"`
	ExpiryDate  string `json:"expiryDate,omitempty"`
	ParentIDTag string `json:"parentIdTag,omitempty"`
}

type StartTransactionRequest struct {
	ConnectorID int    `json:"connectorId"`
	IDTag       string `json:"idTag"`
	MeterStart  int64  `json:"meterStart"`
	Timestamp   string `json:"timestamp"`
	Reservation int64  `json:"reservationId,omitempty"`
}

type StartTransactionResponse struct {
	TransactionID int64     `json:"transactionId"`
	IDTagInfo     IDTagInfo `json:"idTagInfo"`
}

type MeterValuesRequest struct {
	ConnectorID   int                 `json:"connectorId"`
	TransactionID int64               `json:"transactionId"`
	MeterValue    []MeterValuePayload `json:"meterValue"`
}

type MeterValuePayload struct {
	Timestamp    string           `json:"timestamp"`
	SampledValue []SampledValueVM `json:"sampledValue"`
}

type SampledValueVM struct {
	Value     string `json:"value"`
	Context   string `json:"context,omitempty"`
	Format    string `json:"format,omitempty"`
	Measurand string `json:"measurand,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Location  string `json:"location,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

type StopTransactionRequest struct {
	IDTag         string `json:"idTag,omitempty"`
	MeterStop     int64  `json:"meterStop"`
	Timestamp     string `json:"timestamp"`
	TransactionID int64  `json:"transactionId"`
	Reason        string `json:"reason,omitempty"`
}

type StopTransactionResponse struct {
	IDTagInfo *IDTagInfo `json:"idTagInfo,omitempty"`
}