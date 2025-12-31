package node

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Protocol version constants
const (
	// ProtocolVersion is the current protocol version
	ProtocolVersion = "1.0"
	// MinProtocolVersion is the minimum supported version
	MinProtocolVersion = "1.0"
)

// SupportedProtocolVersions lists all protocol versions this node supports.
// Used for version negotiation during handshake.
var SupportedProtocolVersions = []string{"1.0"}

// IsProtocolVersionSupported checks if a given version is supported.
func IsProtocolVersionSupported(version string) bool {
	for _, v := range SupportedProtocolVersions {
		if v == version {
			return true
		}
	}
	return false
}

// MessageType defines the type of P2P message.
type MessageType string

const (
	// Connection lifecycle
	MsgHandshake    MessageType = "handshake"
	MsgHandshakeAck MessageType = "handshake_ack"
	MsgPing         MessageType = "ping"
	MsgPong         MessageType = "pong"
	MsgDisconnect   MessageType = "disconnect"

	// Miner operations
	MsgGetStats   MessageType = "get_stats"
	MsgStats      MessageType = "stats"
	MsgStartMiner MessageType = "start_miner"
	MsgStopMiner  MessageType = "stop_miner"
	MsgMinerAck   MessageType = "miner_ack"

	// Deployment
	MsgDeploy    MessageType = "deploy"
	MsgDeployAck MessageType = "deploy_ack"

	// Logs
	MsgGetLogs MessageType = "get_logs"
	MsgLogs    MessageType = "logs"

	// Error response
	MsgError MessageType = "error"
)

// Message represents a P2P message between nodes.
type Message struct {
	ID        string          `json:"id"` // UUID
	Type      MessageType     `json:"type"`
	From      string          `json:"from"` // Sender node ID
	To        string          `json:"to"`   // Recipient node ID (empty for broadcast)
	Timestamp time.Time       `json:"ts"`
	Payload   json.RawMessage `json:"payload"`
	ReplyTo   string          `json:"replyTo,omitempty"` // ID of message being replied to
}

// NewMessage creates a new message with a generated ID and timestamp.
func NewMessage(msgType MessageType, from, to string, payload interface{}) (*Message, error) {
	var payloadBytes json.RawMessage
	if payload != nil {
		data, err := MarshalJSON(payload)
		if err != nil {
			return nil, err
		}
		payloadBytes = data
	}

	return &Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		From:      from,
		To:        to,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}

// Reply creates a reply message to this message.
func (m *Message) Reply(msgType MessageType, payload interface{}) (*Message, error) {
	reply, err := NewMessage(msgType, m.To, m.From, payload)
	if err != nil {
		return nil, err
	}
	reply.ReplyTo = m.ID
	return reply, nil
}

// ParsePayload unmarshals the payload into the given struct.
func (m *Message) ParsePayload(v interface{}) error {
	if m.Payload == nil {
		return nil
	}
	return json.Unmarshal(m.Payload, v)
}

// --- Payload Types ---

// HandshakePayload is sent during connection establishment.
type HandshakePayload struct {
	Identity  NodeIdentity `json:"identity"`
	Challenge []byte       `json:"challenge,omitempty"` // Random bytes for auth
	Version   string       `json:"version"`             // Protocol version
}

// HandshakeAckPayload is the response to a handshake.
type HandshakeAckPayload struct {
	Identity          NodeIdentity `json:"identity"`
	ChallengeResponse []byte       `json:"challengeResponse,omitempty"`
	Accepted          bool         `json:"accepted"`
	Reason            string       `json:"reason,omitempty"` // If not accepted
}

// PingPayload for keepalive/latency measurement.
type PingPayload struct {
	SentAt int64 `json:"sentAt"` // Unix timestamp in milliseconds
}

// PongPayload response to ping.
type PongPayload struct {
	SentAt     int64 `json:"sentAt"`     // Echo of ping's sentAt
	ReceivedAt int64 `json:"receivedAt"` // When ping was received
}

// StartMinerPayload requests starting a miner.
type StartMinerPayload struct {
	MinerType string          `json:"minerType"` // Required: miner type (e.g., "xmrig", "tt-miner")
	ProfileID string          `json:"profileId,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"` // Override profile config
}

// StopMinerPayload requests stopping a miner.
type StopMinerPayload struct {
	MinerName string `json:"minerName"`
}

// MinerAckPayload acknowledges a miner start/stop operation.
type MinerAckPayload struct {
	Success   bool   `json:"success"`
	MinerName string `json:"minerName,omitempty"`
	Error     string `json:"error,omitempty"`
}

// MinerStatsItem represents stats for a single miner.
type MinerStatsItem struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Hashrate   float64 `json:"hashrate"`
	Shares     int     `json:"shares"`
	Rejected   int     `json:"rejected"`
	Uptime     int     `json:"uptime"` // Seconds
	Pool       string  `json:"pool"`
	Algorithm  string  `json:"algorithm"`
	CPUThreads int     `json:"cpuThreads,omitempty"`
}

// StatsPayload contains miner statistics.
type StatsPayload struct {
	NodeID   string           `json:"nodeId"`
	NodeName string           `json:"nodeName"`
	Miners   []MinerStatsItem `json:"miners"`
	Uptime   int64            `json:"uptime"` // Node uptime in seconds
}

// GetLogsPayload requests console logs from a miner.
type GetLogsPayload struct {
	MinerName string `json:"minerName"`
	Lines     int    `json:"lines"`           // Number of lines to fetch
	Since     int64  `json:"since,omitempty"` // Unix timestamp, logs after this time
}

// LogsPayload contains console log lines.
type LogsPayload struct {
	MinerName string   `json:"minerName"`
	Lines     []string `json:"lines"`
	HasMore   bool     `json:"hasMore"` // More logs available
}

// DeployPayload contains a deployment bundle.
type DeployPayload struct {
	BundleType string `json:"type"`     // "profile" | "miner" | "full"
	Data       []byte `json:"data"`     // STIM-encrypted bundle
	Checksum   string `json:"checksum"` // SHA-256 of Data
	Name       string `json:"name"`     // Profile or miner name
}

// DeployAckPayload acknowledges a deployment.
type DeployAckPayload struct {
	Success bool   `json:"success"`
	Name    string `json:"name,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorPayload contains error information.
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Common error codes
const (
	ErrCodeUnknown         = 1000
	ErrCodeInvalidMessage  = 1001
	ErrCodeUnauthorized    = 1002
	ErrCodeNotFound        = 1003
	ErrCodeOperationFailed = 1004
	ErrCodeTimeout         = 1005
)

// NewErrorMessage creates an error response message.
func NewErrorMessage(from, to string, code int, message string, replyTo string) (*Message, error) {
	msg, err := NewMessage(MsgError, from, to, ErrorPayload{
		Code:    code,
		Message: message,
	})
	if err != nil {
		return nil, err
	}
	msg.ReplyTo = replyTo
	return msg, nil
}
