package ocppcore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func (s *Service) HandleCall(ocppID string, call *OCPPCall) ([]byte, error) {
	switch call.Action {
	case "BootNotification":
		return s.handleBootNotification(ocppID, call)
	case "Heartbeat":
		return s.handleHeartbeat(ocppID, call)
	case "StatusNotification":
		return s.handleStatusNotification(ocppID, call)
	case "Authorize":
		return s.handleAuthorize(ocppID, call)
	case "StartTransaction":
		return s.handleStartTransaction(ocppID, call)
	case "MeterValues":
		return s.handleMeterValues(ocppID, call)
	case "StopTransaction":
		return s.handleStopTransaction(ocppID, call)
	default:
		return BuildCallError(
			call.MessageID,
			"NotSupported",
			fmt.Sprintf("action %s is not supported", call.Action),
			nil,
		)
	}
}

func (s *Service) handleBootNotification(ocppID string, call *OCPPCall) ([]byte, error) {
	var req BootNotificationRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid BootNotification payload", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		UPDATE chargers
		SET vendor = COALESCE(NULLIF($2, ''), vendor),
		    model = COALESCE(NULLIF($3, ''), model),
		    firmware_version = COALESCE(NULLIF($4, ''), firmware_version),
		    status = 'available',
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`,
		ocppID,
		strings.TrimSpace(req.ChargePointVendor),
		strings.TrimSpace(req.ChargePointModel),
		strings.TrimSpace(req.FirmwareVersion),
	)
	if err != nil {
		return BuildCallError(call.MessageID, "InternalError", "failed to update charger boot info", nil)
	}

	resp := BootNotificationResponse{
		Status:      "Accepted",
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
		Interval:    300,
	}

	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleHeartbeat(ocppID string, call *OCPPCall) ([]byte, error) {
	if err := s.TouchHeartbeat(ocppID); err != nil {
		return BuildCallError(call.MessageID, "InternalError", "failed to update heartbeat", nil)
	}

	resp := HeartbeatResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
	}

	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleStatusNotification(ocppID string, call *OCPPCall) ([]byte, error) {
	var req StatusNotificationRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid StatusNotification payload", nil)
	}

	internalStatus := mapOCPPStatusToInternal(req.Status)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		UPDATE chargers
		SET status = $2,
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID, internalStatus)
	if err != nil {
		return BuildCallError(call.MessageID, "InternalError", "failed to update charger status", nil)
	}

	resp := map[string]interface{}{}
	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleAuthorize(_ string, call *OCPPCall) ([]byte, error) {
	var req AuthorizeRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid Authorize payload", nil)
	}

	status := "Invalid"
	if strings.TrimSpace(req.IDTag) != "" {
		status = "Accepted"
	}

	resp := AuthorizeResponse{
		IDTagInfo: IDTagInfo{
			Status: status,
		},
	}

	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleStartTransaction(ocppID string, call *OCPPCall) ([]byte, error) {
	var req StartTransactionRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid StartTransaction payload", nil)
	}

	log.Printf("[OCPP] StartTransaction received ocpp_id=%s connector_id=%d idTag=%s meter_start=%d",
		ocppID, req.ConnectorID, req.IDTag, req.MeterStart)

	chargerID, err := s.GetChargerIDByOCPPID(ocppID)
	if err != nil {
		return BuildCallError(call.MessageID, "InternalError", "charger lookup failed", nil)
	}

	stationID, err := s.GetStationIDByOCPPID(ocppID)
	if err != nil {
		return BuildCallError(call.MessageID, "InternalError", "station lookup failed", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var userID sql.NullString
	if strings.TrimSpace(req.IDTag) != "" {
		_ = s.db.QueryRow(ctx, `
			SELECT id
			FROM users
			WHERE LOWER(email) = LOWER($1) AND is_active = true
			LIMIT 1
		`, strings.TrimSpace(req.IDTag)).Scan(&userID)
	}

	txTime := time.Now().UTC()
	if strings.TrimSpace(req.Timestamp) != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			txTime = parsed.UTC()
		}
	}

	var sessionID string
	var ocppTransactionID int64

	err = s.db.QueryRow(ctx, `
		INSERT INTO charging_sessions (
			user_id,
			charger_id,
			station_id,
			connector_id,
			id_tag,
			status,
			meter_start,
			energy_kwh,
			amount,
			started_at,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,'in_progress',$6,0,0,$7,NOW(),NOW())
		RETURNING id, ocpp_transaction_id
	`,
		nullStringValue(userID),
		chargerID,
		stationID,
		req.ConnectorID,
		nullIfEmpty(req.IDTag),
		float64(req.MeterStart),
		txTime,
	).Scan(&sessionID, &ocppTransactionID)
	if err != nil {
		log.Printf("[OCPP] StartTransaction insert failed ocpp_id=%s err=%v", ocppID, err)
		return BuildCallError(call.MessageID, "InternalError", "failed to create charging session", nil)
	}

	log.Printf("[OCPP] StartTransaction session created ocpp_id=%s session_id=%s ocpp_transaction_id=%d",
		ocppID, sessionID, ocppTransactionID)

	_, _ = s.db.Exec(ctx, `
		UPDATE chargers
		SET status = 'charging',
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID)

	resp := StartTransactionResponse{
		TransactionID: ocppTransactionID,
		IDTagInfo: IDTagInfo{
			Status: "Accepted",
		},
	}

	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleMeterValues(ocppID string, call *OCPPCall) ([]byte, error) {
	var req MeterValuesRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid MeterValues payload", nil)
	}

	log.Printf("[OCPP] MeterValues received ocpp_id=%s transaction_id=%d connector_id=%d entries=%d",
		ocppID, req.TransactionID, req.ConnectorID, len(req.MeterValue))

	chargerID, err := s.GetChargerIDByOCPPID(ocppID)
	if err != nil {
		return BuildCallError(call.MessageID, "InternalError", "charger lookup failed", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sessionID string
	var meterStart sql.NullFloat64

	err = s.db.QueryRow(ctx, `
		SELECT id, meter_start
		FROM charging_sessions
		WHERE ocpp_transaction_id = $1
		LIMIT 1
	`, req.TransactionID).Scan(&sessionID, &meterStart)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[OCPP] MeterValues session not found ocpp_id=%s transaction_id=%d", ocppID, req.TransactionID)
			return BuildCallError(call.MessageID, "PropertyConstraintViolation", "transaction not found", nil)
		}
		return BuildCallError(call.MessageID, "InternalError", "failed to load session", nil)
	}

	var latestWh float64

	for _, mv := range req.MeterValue {
		sampledAt := time.Now().UTC()
		if strings.TrimSpace(mv.Timestamp) != "" {
			if parsed, err := time.Parse(time.RFC3339, mv.Timestamp); err == nil {
				sampledAt = parsed.UTC()
			}
		}

		for _, sv := range mv.SampledValue {
			val, err := strconv.ParseFloat(strings.TrimSpace(sv.Value), 64)
			if err != nil {
				continue
			}

			measurand := strings.TrimSpace(sv.Measurand)
			if measurand == "" {
				measurand = "Energy.Active.Import.Register"
			}

			unit := strings.TrimSpace(sv.Unit)
			if unit == "" {
				unit = "Wh"
			}

			_, err = s.db.Exec(ctx, `
				INSERT INTO meter_values (
					session_id,
					charger_id,
					ocpp_transaction_id,
					sampled_at,
					measurand,
					value,
					unit,
					context
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			`,
				sessionID,
				chargerID,
				req.TransactionID,
				sampledAt,
				measurand,
				val,
				unit,
				nullIfEmpty(sv.Context),
			)
			if err != nil {
				log.Printf("[OCPP] MeterValues insert failed ocpp_id=%s transaction_id=%d err=%v",
					ocppID, req.TransactionID, err)
				return BuildCallError(call.MessageID, "InternalError", "failed to store meter values", nil)
			}

			if strings.EqualFold(measurand, "Energy.Active.Import.Register") {
				if strings.EqualFold(unit, "kWh") {
					latestWh = val * 1000
				} else {
					latestWh = val
				}
			}
		}
	}

	if latestWh > 0 && meterStart.Valid {
		energyKWh := (latestWh - meterStart.Float64) / 1000
		if energyKWh < 0 {
			energyKWh = 0
		}

		_, _ = s.db.Exec(ctx, `
			UPDATE charging_sessions
			SET meter_stop = $2,
			    energy_kwh = $3,
			    updated_at = NOW()
			WHERE id = $1
		`, sessionID, latestWh, energyKWh)
	}

	resp := map[string]interface{}{}
	return BuildCallResult(call.MessageID, resp)
}

func (s *Service) handleStopTransaction(ocppID string, call *OCPPCall) ([]byte, error) {
	var req StopTransactionRequest
	if err := json.Unmarshal(call.Payload, &req); err != nil {
		return BuildCallError(call.MessageID, "FormationViolation", "invalid StopTransaction payload", nil)
	}

	log.Printf("[OCPP] StopTransaction received ocpp_id=%s transaction_id=%d meter_stop=%d reason=%s",
		ocppID, req.TransactionID, req.MeterStop, req.Reason)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopTime := time.Now().UTC()
	if strings.TrimSpace(req.Timestamp) != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			stopTime = parsed.UTC()
		}
	}

	var sessionID string
	var meterStart sql.NullFloat64

	err := s.db.QueryRow(ctx, `
		SELECT id, meter_start
		FROM charging_sessions
		WHERE ocpp_transaction_id = $1
		LIMIT 1
	`, req.TransactionID).Scan(&sessionID, &meterStart)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[OCPP] StopTransaction session not found ocpp_id=%s transaction_id=%d", ocppID, req.TransactionID)
			return BuildCallError(call.MessageID, "PropertyConstraintViolation", "transaction not found", nil)
		}
		return BuildCallError(call.MessageID, "InternalError", "failed to load session", nil)
	}

	meterStop := float64(req.MeterStop)
	energyKWh := 0.0
	if meterStart.Valid {
		energyKWh = (meterStop - meterStart.Float64) / 1000
		if energyKWh < 0 {
			energyKWh = 0
		}
	}

	amount := energyKWh * 12.0

	_, err = s.db.Exec(ctx, `
		UPDATE charging_sessions
		SET status = 'completed',
		    meter_stop = $2,
		    energy_kwh = $3,
		    amount = $4,
		    ended_at = $5,
		    stop_reason = $6,
		    updated_at = NOW()
		WHERE id = $1
	`, sessionID, meterStop, energyKWh, amount, stopTime, nullIfEmpty(req.Reason))
	if err != nil {
		log.Printf("[OCPP] StopTransaction update failed ocpp_id=%s transaction_id=%d err=%v",
			ocppID, req.TransactionID, err)
		return BuildCallError(call.MessageID, "InternalError", "failed to stop transaction", nil)
	}

	_, _ = s.db.Exec(ctx, `
		UPDATE chargers
		SET status = 'available',
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE ocpp_id = $1
	`, ocppID)

	resp := StopTransactionResponse{}
	return BuildCallResult(call.MessageID, resp)
}

func mapOCPPStatusToInternal(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "available":
		return "available"
	case "preparing":
		return "preparing"
	case "charging":
		return "charging"
	case "finishing":
		return "finishing"
	case "faulted":
		return "faulted"
	case "unavailable":
		return "unavailable"
	default:
		return "offline"
	}
}

func nullIfEmpty(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.TrimSpace(v)
}

func nullStringValue(v sql.NullString) interface{} {
	if v.Valid {
		return v.String
	}
	return nil
}