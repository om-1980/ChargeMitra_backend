package ocppcore

import (
	"encoding/json"
	"errors"
	"fmt"
)

func ParseOCPPMessage(raw []byte) (*OCPPMessage, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("invalid json frame: %w", err)
	}

	if len(arr) < 3 {
		return nil, errors.New("invalid ocpp frame length")
	}

	var messageType int
	if err := json.Unmarshal(arr[0], &messageType); err != nil {
		return nil, errors.New("invalid message type")
	}

	var messageID string
	if err := json.Unmarshal(arr[1], &messageID); err != nil {
		return nil, errors.New("invalid message id")
	}

	msg := &OCPPMessage{
		MessageType: messageType,
		MessageID:   messageID,
	}

	switch messageType {
	case MessageTypeCall:
		if len(arr) < 4 {
			return nil, errors.New("invalid ocpp call frame length")
		}

		var action string
		if err := json.Unmarshal(arr[2], &action); err != nil {
			return nil, errors.New("invalid action")
		}

		msg.Action = action
		msg.Payload = arr[3]
		return msg, nil

	case MessageTypeCallResult:
		msg.Action = "CALLRESULT"
		msg.Payload = arr[2]
		return msg, nil

	case MessageTypeCallError:
		if len(arr) < 5 {
			return nil, errors.New("invalid ocpp call error frame length")
		}

		if err := json.Unmarshal(arr[2], &msg.ErrorCode); err != nil {
			return nil, errors.New("invalid error code")
		}
		if err := json.Unmarshal(arr[3], &msg.ErrorDescription); err != nil {
			return nil, errors.New("invalid error description")
		}
		_ = json.Unmarshal(arr[4], &msg.ErrorDetails)
		msg.Action = "CALLERROR"
		return msg, nil

	default:
		return nil, fmt.Errorf("unsupported message type: %d", messageType)
	}
}

func ParseOCPPCall(raw []byte) (*OCPPCall, error) {
	msg, err := ParseOCPPMessage(raw)
	if err != nil {
		return nil, err
	}
	if msg.MessageType != MessageTypeCall {
		return nil, fmt.Errorf("unsupported message type for call parser: %d", msg.MessageType)
	}

	return &OCPPCall{
		MessageType: msg.MessageType,
		MessageID:   msg.MessageID,
		Action:      msg.Action,
		Payload:     msg.Payload,
	}, nil
}

func BuildCall(messageID, action string, payload interface{}) ([]byte, error) {
	frame := []interface{}{
		MessageTypeCall,
		messageID,
		action,
		payload,
	}
	return json.Marshal(frame)
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