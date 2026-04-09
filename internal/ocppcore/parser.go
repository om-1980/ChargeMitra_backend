package ocppcore

import (
	"encoding/json"
	"errors"
	"fmt"
)

func ParseOCPPCall(raw []byte) (*OCPPCall, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("invalid json frame: %w", err)
	}

	if len(arr) < 4 {
		return nil, errors.New("invalid ocpp call frame length")
	}

	var messageType int
	if err := json.Unmarshal(arr[0], &messageType); err != nil {
		return nil, errors.New("invalid message type")
	}

	if messageType != MessageTypeCall {
		return nil, fmt.Errorf("unsupported message type: %d", messageType)
	}

	var messageID string
	if err := json.Unmarshal(arr[1], &messageID); err != nil {
		return nil, errors.New("invalid message id")
	}

	var action string
	if err := json.Unmarshal(arr[2], &action); err != nil {
		return nil, errors.New("invalid action")
	}

	return &OCPPCall{
		MessageType: messageType,
		MessageID:   messageID,
		Action:      action,
		Payload:     arr[3],
	}, nil
}

func BuildCallResult(messageID string, payload interface{}) ([]byte, error) {
	frame := []interface{}{
		MessageTypeCallResult,
		messageID,
		payload,
	}
	return json.Marshal(frame)
}

func BuildCallError(messageID, errorCode, errorDescription string, details map[string]interface{}) ([]byte, error) {
	if details == nil {
		details = map[string]interface{}{}
	}

	frame := []interface{}{
		MessageTypeCallError,
		messageID,
		errorCode,
		errorDescription,
		details,
	}
	return json.Marshal(frame)
}