package node

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	t.Run("BasicMessage", func(t *testing.T) {
		msg, err := NewMessage(MsgPing, "sender-id", "receiver-id", nil)
		if err != nil {
			t.Fatalf("failed to create message: %v", err)
		}

		if msg.Type != MsgPing {
			t.Errorf("expected type MsgPing, got %s", msg.Type)
		}

		if msg.From != "sender-id" {
			t.Errorf("expected from 'sender-id', got '%s'", msg.From)
		}

		if msg.To != "receiver-id" {
			t.Errorf("expected to 'receiver-id', got '%s'", msg.To)
		}

		if msg.ID == "" {
			t.Error("message ID should not be empty")
		}

		if msg.Timestamp.IsZero() {
			t.Error("timestamp should be set")
		}
	})

	t.Run("MessageWithPayload", func(t *testing.T) {
		payload := PingPayload{
			SentAt: time.Now().UnixMilli(),
		}

		msg, err := NewMessage(MsgPing, "sender", "receiver", payload)
		if err != nil {
			t.Fatalf("failed to create message: %v", err)
		}

		if msg.Payload == nil {
			t.Error("payload should not be nil")
		}

		var parsed PingPayload
		err = msg.ParsePayload(&parsed)
		if err != nil {
			t.Fatalf("failed to parse payload: %v", err)
		}

		if parsed.SentAt != payload.SentAt {
			t.Errorf("expected SentAt %d, got %d", payload.SentAt, parsed.SentAt)
		}
	})
}

func TestMessageReply(t *testing.T) {
	original, _ := NewMessage(MsgPing, "sender", "receiver", PingPayload{SentAt: 12345})

	reply, err := original.Reply(MsgPong, PongPayload{
		SentAt:     12345,
		ReceivedAt: 12350,
	})

	if err != nil {
		t.Fatalf("failed to create reply: %v", err)
	}

	if reply.ReplyTo != original.ID {
		t.Errorf("reply should reference original message ID")
	}

	if reply.From != original.To {
		t.Error("reply From should be original To")
	}

	if reply.To != original.From {
		t.Error("reply To should be original From")
	}

	if reply.Type != MsgPong {
		t.Errorf("expected type MsgPong, got %s", reply.Type)
	}
}

func TestParsePayload(t *testing.T) {
	t.Run("ValidPayload", func(t *testing.T) {
		payload := StartMinerPayload{
			ProfileID: "test-profile",
		}

		msg, _ := NewMessage(MsgStartMiner, "ctrl", "worker", payload)

		var parsed StartMinerPayload
		err := msg.ParsePayload(&parsed)
		if err != nil {
			t.Fatalf("failed to parse payload: %v", err)
		}

		if parsed.ProfileID != "test-profile" {
			t.Errorf("expected ProfileID 'test-profile', got '%s'", parsed.ProfileID)
		}
	})

	t.Run("NilPayload", func(t *testing.T) {
		msg, _ := NewMessage(MsgGetStats, "ctrl", "worker", nil)

		var parsed StatsPayload
		err := msg.ParsePayload(&parsed)
		if err != nil {
			t.Errorf("parsing nil payload should not error: %v", err)
		}
	})

	t.Run("ComplexPayload", func(t *testing.T) {
		stats := StatsPayload{
			NodeID:   "node-123",
			NodeName: "Test Node",
			Miners: []MinerStatsItem{
				{
					Name:      "xmrig-1",
					Type:      "xmrig",
					Hashrate:  1234.56,
					Shares:    100,
					Rejected:  2,
					Uptime:    3600,
					Pool:      "pool.example.com:3333",
					Algorithm: "RandomX",
				},
			},
			Uptime: 86400,
		}

		msg, _ := NewMessage(MsgStats, "worker", "ctrl", stats)

		var parsed StatsPayload
		err := msg.ParsePayload(&parsed)
		if err != nil {
			t.Fatalf("failed to parse stats payload: %v", err)
		}

		if parsed.NodeID != "node-123" {
			t.Errorf("expected NodeID 'node-123', got '%s'", parsed.NodeID)
		}

		if len(parsed.Miners) != 1 {
			t.Fatalf("expected 1 miner, got %d", len(parsed.Miners))
		}

		if parsed.Miners[0].Hashrate != 1234.56 {
			t.Errorf("expected hashrate 1234.56, got %f", parsed.Miners[0].Hashrate)
		}
	})
}

func TestNewErrorMessage(t *testing.T) {
	errMsg, err := NewErrorMessage("sender", "receiver", ErrCodeOperationFailed, "something went wrong", "original-msg-id")
	if err != nil {
		t.Fatalf("failed to create error message: %v", err)
	}

	if errMsg.Type != MsgError {
		t.Errorf("expected type MsgError, got %s", errMsg.Type)
	}

	if errMsg.ReplyTo != "original-msg-id" {
		t.Errorf("expected ReplyTo 'original-msg-id', got '%s'", errMsg.ReplyTo)
	}

	var errPayload ErrorPayload
	err = errMsg.ParsePayload(&errPayload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if errPayload.Code != ErrCodeOperationFailed {
		t.Errorf("expected code %d, got %d", ErrCodeOperationFailed, errPayload.Code)
	}

	if errPayload.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", errPayload.Message)
	}
}

func TestMessageSerialization(t *testing.T) {
	original, _ := NewMessage(MsgStartMiner, "ctrl", "worker", StartMinerPayload{
		ProfileID: "my-profile",
	})

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to serialize message: %v", err)
	}

	// Deserialize
	var restored Message
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("failed to deserialize message: %v", err)
	}

	if restored.ID != original.ID {
		t.Error("ID mismatch after serialization")
	}

	if restored.Type != original.Type {
		t.Error("Type mismatch after serialization")
	}

	if restored.From != original.From {
		t.Error("From mismatch after serialization")
	}

	var payload StartMinerPayload
	err = restored.ParsePayload(&payload)
	if err != nil {
		t.Fatalf("failed to parse restored payload: %v", err)
	}

	if payload.ProfileID != "my-profile" {
		t.Errorf("expected ProfileID 'my-profile', got '%s'", payload.ProfileID)
	}
}

func TestMessageTypes(t *testing.T) {
	types := []MessageType{
		MsgHandshake,
		MsgHandshakeAck,
		MsgPing,
		MsgPong,
		MsgDisconnect,
		MsgGetStats,
		MsgStats,
		MsgStartMiner,
		MsgStopMiner,
		MsgMinerAck,
		MsgDeploy,
		MsgDeployAck,
		MsgGetLogs,
		MsgLogs,
		MsgError,
	}

	for _, msgType := range types {
		t.Run(string(msgType), func(t *testing.T) {
			msg, err := NewMessage(msgType, "from", "to", nil)
			if err != nil {
				t.Fatalf("failed to create message of type %s: %v", msgType, err)
			}

			if msg.Type != msgType {
				t.Errorf("expected type %s, got %s", msgType, msg.Type)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	codes := map[int]string{
		ErrCodeUnknown:         "Unknown",
		ErrCodeInvalidMessage:  "InvalidMessage",
		ErrCodeUnauthorized:    "Unauthorized",
		ErrCodeNotFound:        "NotFound",
		ErrCodeOperationFailed: "OperationFailed",
		ErrCodeTimeout:         "Timeout",
	}

	for code, name := range codes {
		t.Run(name, func(t *testing.T) {
			if code < 1000 || code > 1999 {
				t.Errorf("error code %d should be in 1000-1999 range", code)
			}
		})
	}
}
