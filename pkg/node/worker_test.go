package node

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestEnv sets up a temporary environment for testing and returns cleanup function
func setupTestEnv(t *testing.T) func() {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	return func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_DATA_HOME")
	}
}

func TestNewWorker(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	if worker == nil {
		t.Fatal("NewWorker returned nil")
	}
	if worker.node != nm {
		t.Error("worker.node not set correctly")
	}
	if worker.transport != transport {
		t.Error("worker.transport not set correctly")
	}
}

func TestWorker_SetMinerManager(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	mockManager := &mockMinerManager{}
	worker.SetMinerManager(mockManager)

	if worker.minerManager != mockManager {
		t.Error("minerManager not set correctly")
	}
}

func TestWorker_SetProfileManager(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	mockProfile := &mockProfileManager{}
	worker.SetProfileManager(mockProfile)

	if worker.profileManager != mockProfile {
		t.Error("profileManager not set correctly")
	}
}

func TestWorker_HandlePing(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a ping message
	identity := nm.GetIdentity()
	pingPayload := PingPayload{SentAt: time.Now().UnixMilli()}
	pingMsg, err := NewMessage(MsgPing, "sender-id", identity.ID, pingPayload)
	if err != nil {
		t.Fatalf("failed to create ping message: %v", err)
	}

	// Call handlePing directly
	response, err := worker.handlePing(pingMsg)
	if err != nil {
		t.Fatalf("handlePing returned error: %v", err)
	}

	if response == nil {
		t.Fatal("handlePing returned nil response")
	}

	if response.Type != MsgPong {
		t.Errorf("expected response type %s, got %s", MsgPong, response.Type)
	}

	var pong PongPayload
	if err := response.ParsePayload(&pong); err != nil {
		t.Fatalf("failed to parse pong payload: %v", err)
	}

	if pong.SentAt != pingPayload.SentAt {
		t.Errorf("pong SentAt mismatch: expected %d, got %d", pingPayload.SentAt, pong.SentAt)
	}

	if pong.ReceivedAt == 0 {
		t.Error("pong ReceivedAt not set")
	}
}

func TestWorker_HandleGetStats(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a get_stats message
	identity := nm.GetIdentity()
	msg, err := NewMessage(MsgGetStats, "sender-id", identity.ID, nil)
	if err != nil {
		t.Fatalf("failed to create get_stats message: %v", err)
	}

	// Call handleGetStats directly (without miner manager)
	response, err := worker.handleGetStats(msg)
	if err != nil {
		t.Fatalf("handleGetStats returned error: %v", err)
	}

	if response == nil {
		t.Fatal("handleGetStats returned nil response")
	}

	if response.Type != MsgStats {
		t.Errorf("expected response type %s, got %s", MsgStats, response.Type)
	}

	var stats StatsPayload
	if err := response.ParsePayload(&stats); err != nil {
		t.Fatalf("failed to parse stats payload: %v", err)
	}

	if stats.NodeID != identity.ID {
		t.Errorf("stats NodeID mismatch: expected %s, got %s", identity.ID, stats.NodeID)
	}

	if stats.NodeName != identity.Name {
		t.Errorf("stats NodeName mismatch: expected %s, got %s", identity.Name, stats.NodeName)
	}
}

func TestWorker_HandleStartMiner_NoManager(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a start_miner message
	identity := nm.GetIdentity()
	payload := StartMinerPayload{ProfileID: "test-profile"}
	msg, err := NewMessage(MsgStartMiner, "sender-id", identity.ID, payload)
	if err != nil {
		t.Fatalf("failed to create start_miner message: %v", err)
	}

	// Without miner manager, should return error
	_, err = worker.handleStartMiner(msg)
	if err == nil {
		t.Error("expected error when miner manager is nil")
	}
}

func TestWorker_HandleStopMiner_NoManager(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a stop_miner message
	identity := nm.GetIdentity()
	payload := StopMinerPayload{MinerName: "test-miner"}
	msg, err := NewMessage(MsgStopMiner, "sender-id", identity.ID, payload)
	if err != nil {
		t.Fatalf("failed to create stop_miner message: %v", err)
	}

	// Without miner manager, should return error
	_, err = worker.handleStopMiner(msg)
	if err == nil {
		t.Error("expected error when miner manager is nil")
	}
}

func TestWorker_HandleGetLogs_NoManager(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a get_logs message
	identity := nm.GetIdentity()
	payload := GetLogsPayload{MinerName: "test-miner", Lines: 100}
	msg, err := NewMessage(MsgGetLogs, "sender-id", identity.ID, payload)
	if err != nil {
		t.Fatalf("failed to create get_logs message: %v", err)
	}

	// Without miner manager, should return error
	_, err = worker.handleGetLogs(msg)
	if err == nil {
		t.Error("expected error when miner manager is nil")
	}
}

func TestWorker_HandleDeploy_Profile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a deploy message for profile
	identity := nm.GetIdentity()
	payload := DeployPayload{
		BundleType: "profile",
		Data:       []byte(`{"id": "test", "name": "Test Profile"}`),
		Name:       "test-profile",
	}
	msg, err := NewMessage(MsgDeploy, "sender-id", identity.ID, payload)
	if err != nil {
		t.Fatalf("failed to create deploy message: %v", err)
	}

	// Without profile manager, should return error
	_, err = worker.handleDeploy(msg)
	if err == nil {
		t.Error("expected error when profile manager is nil")
	}
}

func TestWorker_HandleDeploy_UnknownType(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	nm, err := NewNodeManager()
	if err != nil {
		t.Fatalf("failed to create node manager: %v", err)
	}
	if err := nm.GenerateIdentity("test-worker", RoleWorker); err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	pr, err := NewPeerRegistryWithPath(t.TempDir() + "/peers.json")
	if err != nil {
		t.Fatalf("failed to create peer registry: %v", err)
	}

	transport := NewTransport(nm, pr, DefaultTransportConfig())
	worker := NewWorker(nm, transport)

	// Create a deploy message with unknown type
	identity := nm.GetIdentity()
	payload := DeployPayload{
		BundleType: "unknown",
		Data:       []byte(`{}`),
		Name:       "test",
	}
	msg, err := NewMessage(MsgDeploy, "sender-id", identity.ID, payload)
	if err != nil {
		t.Fatalf("failed to create deploy message: %v", err)
	}

	_, err = worker.handleDeploy(msg)
	if err == nil {
		t.Error("expected error for unknown bundle type")
	}
}

func TestConvertMinerStats(t *testing.T) {
	tests := []struct {
		name     string
		rawStats interface{}
		wantHash float64
	}{
		{
			name: "MapWithHashrate",
			rawStats: map[string]interface{}{
				"hashrate":  100.5,
				"shares":    10,
				"rejected":  2,
				"uptime":    3600,
				"pool":      "test-pool",
				"algorithm": "rx/0",
			},
			wantHash: 100.5,
		},
		{
			name:     "EmptyMap",
			rawStats: map[string]interface{}{},
			wantHash: 0,
		},
		{
			name:     "NonMap",
			rawStats: "not a map",
			wantHash: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinerInstance{name: "test", minerType: "xmrig"}
			result := convertMinerStats(mock, tt.rawStats)

			if result.Name != "test" {
				t.Errorf("expected name 'test', got '%s'", result.Name)
			}
			if result.Hashrate != tt.wantHash {
				t.Errorf("expected hashrate %f, got %f", tt.wantHash, result.Hashrate)
			}
		})
	}
}

// Mock implementations for testing

type mockMinerManager struct {
	miners []MinerInstance
}

func (m *mockMinerManager) StartMiner(minerType string, config interface{}) (MinerInstance, error) {
	return nil, nil
}

func (m *mockMinerManager) StopMiner(name string) error {
	return nil
}

func (m *mockMinerManager) ListMiners() []MinerInstance {
	return m.miners
}

func (m *mockMinerManager) GetMiner(name string) (MinerInstance, error) {
	for _, miner := range m.miners {
		if miner.GetName() == name {
			return miner, nil
		}
	}
	return nil, nil
}

type mockMinerInstance struct {
	name      string
	minerType string
	stats     interface{}
}

func (m *mockMinerInstance) GetName() string                      { return m.name }
func (m *mockMinerInstance) GetType() string                      { return m.minerType }
func (m *mockMinerInstance) GetStats() (interface{}, error)       { return m.stats, nil }
func (m *mockMinerInstance) GetConsoleHistory(lines int) []string { return []string{} }

type mockProfileManager struct{}

func (m *mockProfileManager) GetProfile(id string) (interface{}, error) {
	return nil, nil
}

func (m *mockProfileManager) SaveProfile(profile interface{}) error {
	return nil
}
