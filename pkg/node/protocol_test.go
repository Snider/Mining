package node

import (
	"fmt"
	"testing"
)

func TestResponseHandler_ValidateResponse(t *testing.T) {
	handler := &ResponseHandler{}

	t.Run("NilResponse", func(t *testing.T) {
		err := handler.ValidateResponse(nil, MsgStats)
		if err == nil {
			t.Error("Expected error for nil response")
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		errMsg, _ := NewErrorMessage("sender", "receiver", ErrCodeOperationFailed, "operation failed", "")
		err := handler.ValidateResponse(errMsg, MsgStats)
		if err == nil {
			t.Fatal("Expected error for error response")
		}

		if !IsProtocolError(err) {
			t.Errorf("Expected ProtocolError, got %T", err)
		}

		if GetProtocolErrorCode(err) != ErrCodeOperationFailed {
			t.Errorf("Expected code %d, got %d", ErrCodeOperationFailed, GetProtocolErrorCode(err))
		}
	})

	t.Run("WrongType", func(t *testing.T) {
		msg, _ := NewMessage(MsgPong, "sender", "receiver", nil)
		err := handler.ValidateResponse(msg, MsgStats)
		if err == nil {
			t.Error("Expected error for wrong type")
		}
		if IsProtocolError(err) {
			t.Error("Should not be a ProtocolError for type mismatch")
		}
	})

	t.Run("ValidResponse", func(t *testing.T) {
		msg, _ := NewMessage(MsgStats, "sender", "receiver", StatsPayload{NodeID: "test"})
		err := handler.ValidateResponse(msg, MsgStats)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestResponseHandler_ParseResponse(t *testing.T) {
	handler := &ResponseHandler{}

	t.Run("ParseStats", func(t *testing.T) {
		payload := StatsPayload{
			NodeID:   "node-123",
			NodeName: "Test Node",
			Uptime:   3600,
		}
		msg, _ := NewMessage(MsgStats, "sender", "receiver", payload)

		var parsed StatsPayload
		err := handler.ParseResponse(msg, MsgStats, &parsed)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if parsed.NodeID != "node-123" {
			t.Errorf("Expected NodeID 'node-123', got '%s'", parsed.NodeID)
		}
		if parsed.Uptime != 3600 {
			t.Errorf("Expected Uptime 3600, got %d", parsed.Uptime)
		}
	})

	t.Run("ParseMinerAck", func(t *testing.T) {
		payload := MinerAckPayload{
			Success:   true,
			MinerName: "xmrig-1",
		}
		msg, _ := NewMessage(MsgMinerAck, "sender", "receiver", payload)

		var parsed MinerAckPayload
		err := handler.ParseResponse(msg, MsgMinerAck, &parsed)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !parsed.Success {
			t.Error("Expected Success to be true")
		}
		if parsed.MinerName != "xmrig-1" {
			t.Errorf("Expected MinerName 'xmrig-1', got '%s'", parsed.MinerName)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		errMsg, _ := NewErrorMessage("sender", "receiver", ErrCodeNotFound, "not found", "")

		var parsed StatsPayload
		err := handler.ParseResponse(errMsg, MsgStats, &parsed)
		if err == nil {
			t.Error("Expected error for error response")
		}
		if !IsProtocolError(err) {
			t.Errorf("Expected ProtocolError, got %T", err)
		}
	})

	t.Run("NilTarget", func(t *testing.T) {
		msg, _ := NewMessage(MsgPong, "sender", "receiver", nil)
		err := handler.ParseResponse(msg, MsgPong, nil)
		if err != nil {
			t.Errorf("Unexpected error with nil target: %v", err)
		}
	})
}

func TestProtocolError(t *testing.T) {
	err := &ProtocolError{Code: 1001, Message: "test error"}

	if err.Error() != "remote error (1001): test error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	if !IsProtocolError(err) {
		t.Error("IsProtocolError should return true")
	}

	if GetProtocolErrorCode(err) != 1001 {
		t.Errorf("Expected code 1001, got %d", GetProtocolErrorCode(err))
	}
}

func TestConvenienceFunctions(t *testing.T) {
	msg, _ := NewMessage(MsgStats, "sender", "receiver", StatsPayload{NodeID: "test"})

	// Test ValidateResponse
	if err := ValidateResponse(msg, MsgStats); err != nil {
		t.Errorf("ValidateResponse failed: %v", err)
	}

	// Test ParseResponse
	var parsed StatsPayload
	if err := ParseResponse(msg, MsgStats, &parsed); err != nil {
		t.Errorf("ParseResponse failed: %v", err)
	}
	if parsed.NodeID != "test" {
		t.Errorf("Expected NodeID 'test', got '%s'", parsed.NodeID)
	}
}

func TestGetProtocolErrorCode_NonProtocolError(t *testing.T) {
	err := fmt.Errorf("regular error")
	if GetProtocolErrorCode(err) != 0 {
		t.Error("Expected 0 for non-ProtocolError")
	}
}
