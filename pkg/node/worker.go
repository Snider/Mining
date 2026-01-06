package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/adrg/xdg"
)

// MinerManager interface for the mining package integration.
// This allows the node package to interact with mining.Manager without import cycles.
type MinerManager interface {
	StartMiner(minerType string, config interface{}) (MinerInstance, error)
	StopMiner(name string) error
	ListMiners() []MinerInstance
	GetMiner(name string) (MinerInstance, error)
}

// MinerInstance represents a running miner for stats collection.
type MinerInstance interface {
	GetName() string
	GetType() string
	GetStats() (interface{}, error)
	GetConsoleHistory(lines int) []string
}

// ProfileManager interface for profile operations.
type ProfileManager interface {
	GetProfile(id string) (interface{}, error)
	SaveProfile(profile interface{}) error
}

// Worker handles incoming messages on a worker node.
type Worker struct {
	node           *NodeManager
	transport      *Transport
	minerManager   MinerManager
	profileManager ProfileManager
	startTime      time.Time
}

// NewWorker creates a new Worker instance.
func NewWorker(node *NodeManager, transport *Transport) *Worker {
	return &Worker{
		node:      node,
		transport: transport,
		startTime: time.Now(),
	}
}

// SetMinerManager sets the miner manager for handling miner operations.
func (w *Worker) SetMinerManager(manager MinerManager) {
	w.minerManager = manager
}

// SetProfileManager sets the profile manager for handling profile operations.
func (w *Worker) SetProfileManager(manager ProfileManager) {
	w.profileManager = manager
}

// HandleMessage processes incoming messages and returns a response.
func (w *Worker) HandleMessage(conn *PeerConnection, msg *Message) {
	var response *Message
	var err error

	switch msg.Type {
	case MsgPing:
		response, err = w.handlePing(msg)
	case MsgGetStats:
		response, err = w.handleGetStats(msg)
	case MsgStartMiner:
		response, err = w.handleStartMiner(msg)
	case MsgStopMiner:
		response, err = w.handleStopMiner(msg)
	case MsgGetLogs:
		response, err = w.handleGetLogs(msg)
	case MsgDeploy:
		response, err = w.handleDeploy(conn, msg)
	default:
		// Unknown message type - ignore or send error
		return
	}

	if err != nil {
		// Send error response
		identity := w.node.GetIdentity()
		if identity != nil {
			errMsg, _ := NewErrorMessage(
				identity.ID,
				msg.From,
				ErrCodeOperationFailed,
				err.Error(),
				msg.ID,
			)
			conn.Send(errMsg)
		}
		return
	}

	if response != nil {
		logging.Debug("sending response", logging.Fields{"type": response.Type, "to": msg.From})
		if err := conn.Send(response); err != nil {
			logging.Error("failed to send response", logging.Fields{"error": err})
		} else {
			logging.Debug("response sent successfully")
		}
	}
}

// handlePing responds to ping requests.
func (w *Worker) handlePing(msg *Message) (*Message, error) {
	var ping PingPayload
	if err := msg.ParsePayload(&ping); err != nil {
		return nil, fmt.Errorf("invalid ping payload: %w", err)
	}

	pong := PongPayload{
		SentAt:     ping.SentAt,
		ReceivedAt: time.Now().UnixMilli(),
	}

	return msg.Reply(MsgPong, pong)
}

// handleGetStats responds with current miner statistics.
func (w *Worker) handleGetStats(msg *Message) (*Message, error) {
	identity := w.node.GetIdentity()
	if identity == nil {
		return nil, fmt.Errorf("node identity not initialized")
	}

	stats := StatsPayload{
		NodeID:   identity.ID,
		NodeName: identity.Name,
		Miners:   []MinerStatsItem{},
		Uptime:   int64(time.Since(w.startTime).Seconds()),
	}

	if w.minerManager != nil {
		miners := w.minerManager.ListMiners()
		for _, miner := range miners {
			minerStats, err := miner.GetStats()
			if err != nil {
				continue
			}

			// Convert to MinerStatsItem - this is a simplified conversion
			// The actual implementation would need to match the mining package's stats structure
			item := convertMinerStats(miner, minerStats)
			stats.Miners = append(stats.Miners, item)
		}
	}

	return msg.Reply(MsgStats, stats)
}

// convertMinerStats converts miner stats to the protocol format.
func convertMinerStats(miner MinerInstance, rawStats interface{}) MinerStatsItem {
	item := MinerStatsItem{
		Name: miner.GetName(),
		Type: miner.GetType(),
	}

	// Try to extract common fields from the stats
	if statsMap, ok := rawStats.(map[string]interface{}); ok {
		if hashrate, ok := statsMap["hashrate"].(float64); ok {
			item.Hashrate = hashrate
		}
		if shares, ok := statsMap["shares"].(int); ok {
			item.Shares = shares
		}
		if rejected, ok := statsMap["rejected"].(int); ok {
			item.Rejected = rejected
		}
		if uptime, ok := statsMap["uptime"].(int); ok {
			item.Uptime = uptime
		}
		if pool, ok := statsMap["pool"].(string); ok {
			item.Pool = pool
		}
		if algorithm, ok := statsMap["algorithm"].(string); ok {
			item.Algorithm = algorithm
		}
	}

	return item
}

// handleStartMiner starts a miner with the given profile.
func (w *Worker) handleStartMiner(msg *Message) (*Message, error) {
	if w.minerManager == nil {
		return nil, fmt.Errorf("miner manager not configured")
	}

	var payload StartMinerPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid start miner payload: %w", err)
	}

	// Validate miner type is provided
	if payload.MinerType == "" {
		return nil, fmt.Errorf("miner type is required")
	}

	// Get the config from the profile or use the override
	var config interface{}
	if payload.Config != nil {
		config = payload.Config
	} else if w.profileManager != nil {
		profile, err := w.profileManager.GetProfile(payload.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("profile not found: %s", payload.ProfileID)
		}
		config = profile
	} else {
		return nil, fmt.Errorf("no config provided and no profile manager configured")
	}

	// Start the miner
	miner, err := w.minerManager.StartMiner(payload.MinerType, config)
	if err != nil {
		ack := MinerAckPayload{
			Success: false,
			Error:   err.Error(),
		}
		return msg.Reply(MsgMinerAck, ack)
	}

	ack := MinerAckPayload{
		Success:   true,
		MinerName: miner.GetName(),
	}
	return msg.Reply(MsgMinerAck, ack)
}

// handleStopMiner stops a running miner.
func (w *Worker) handleStopMiner(msg *Message) (*Message, error) {
	if w.minerManager == nil {
		return nil, fmt.Errorf("miner manager not configured")
	}

	var payload StopMinerPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid stop miner payload: %w", err)
	}

	err := w.minerManager.StopMiner(payload.MinerName)
	ack := MinerAckPayload{
		Success:   err == nil,
		MinerName: payload.MinerName,
	}
	if err != nil {
		ack.Error = err.Error()
	}

	return msg.Reply(MsgMinerAck, ack)
}

// handleGetLogs returns console logs from a miner.
func (w *Worker) handleGetLogs(msg *Message) (*Message, error) {
	if w.minerManager == nil {
		return nil, fmt.Errorf("miner manager not configured")
	}

	var payload GetLogsPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid get logs payload: %w", err)
	}

	// Validate and limit the Lines parameter to prevent resource exhaustion
	const maxLogLines = 10000
	if payload.Lines <= 0 || payload.Lines > maxLogLines {
		payload.Lines = maxLogLines
	}

	miner, err := w.minerManager.GetMiner(payload.MinerName)
	if err != nil {
		return nil, fmt.Errorf("miner not found: %s", payload.MinerName)
	}

	lines := miner.GetConsoleHistory(payload.Lines)

	logs := LogsPayload{
		MinerName: payload.MinerName,
		Lines:     lines,
		HasMore:   len(lines) >= payload.Lines,
	}

	return msg.Reply(MsgLogs, logs)
}

// handleDeploy handles deployment of profiles or miner bundles.
func (w *Worker) handleDeploy(conn *PeerConnection, msg *Message) (*Message, error) {
	var payload DeployPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid deploy payload: %w", err)
	}

	// Reconstruct Bundle object from payload
	bundle := &Bundle{
		Type:     BundleType(payload.BundleType),
		Name:     payload.Name,
		Data:     payload.Data,
		Checksum: payload.Checksum,
	}

	// Use shared secret as password (base64 encoded)
	password := ""
	if conn != nil && len(conn.SharedSecret) > 0 {
		password = base64.StdEncoding.EncodeToString(conn.SharedSecret)
	}

	switch bundle.Type {
	case BundleProfile:
		if w.profileManager == nil {
			return nil, fmt.Errorf("profile manager not configured")
		}

		// Decrypt and extract profile data
		profileData, err := ExtractProfileBundle(bundle, password)
		if err != nil {
			return nil, fmt.Errorf("failed to extract profile bundle: %w", err)
		}

		// Unmarshal into interface{} to pass to ProfileManager
		var profile interface{}
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return nil, fmt.Errorf("invalid profile data JSON: %w", err)
		}

		if err := w.profileManager.SaveProfile(profile); err != nil {
			ack := DeployAckPayload{
				Success: false,
				Name:    payload.Name,
				Error:   err.Error(),
			}
			return msg.Reply(MsgDeployAck, ack)
		}

		ack := DeployAckPayload{
			Success: true,
			Name:    payload.Name,
		}
		return msg.Reply(MsgDeployAck, ack)

	case BundleMiner, BundleFull:
		// Determine installation directory
		// We use xdg.DataHome/lethean-desktop/miners/<bundle_name>
		minersDir := filepath.Join(xdg.DataHome, "lethean-desktop", "miners")
		installDir := filepath.Join(minersDir, payload.Name)

		logging.Info("deploying miner bundle", logging.Fields{
			"name": payload.Name,
			"path": installDir,
			"type": payload.BundleType,
		})

		// Extract miner bundle
		minerPath, profileData, err := ExtractMinerBundle(bundle, password, installDir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract miner bundle: %w", err)
		}

		// If the bundle contained a profile config, save it
		if len(profileData) > 0 && w.profileManager != nil {
			var profile interface{}
			if err := json.Unmarshal(profileData, &profile); err != nil {
				logging.Warn("failed to parse profile from miner bundle", logging.Fields{"error": err})
			} else {
				if err := w.profileManager.SaveProfile(profile); err != nil {
					logging.Warn("failed to save profile from miner bundle", logging.Fields{"error": err})
				}
			}
		}

		// Success response
		ack := DeployAckPayload{
			Success: true,
			Name:    payload.Name,
		}

		// Log the installation
		logging.Info("miner bundle installed successfully", logging.Fields{
			"name":       payload.Name,
			"miner_path": minerPath,
		})

		return msg.Reply(MsgDeployAck, ack)

	default:
		return nil, fmt.Errorf("unknown bundle type: %s", payload.BundleType)
	}
}

// RegisterWithTransport registers the worker's message handler with the transport.
func (w *Worker) RegisterWithTransport() {
	w.transport.OnMessage(w.HandleMessage)
}
