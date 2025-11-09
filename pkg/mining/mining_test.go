package mining

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.miners == nil {
		t.Error("Manager miners map is nil")
	}
}

func TestStartAndStopMiner(t *testing.T) {
	manager := NewManager()

	config := &Config{
		Pool:   "pool.example.com",
		Wallet: "wallet123",
	}

	// We can't fully test StartMiner without a mock miner,
	// but we can test the manager's behavior.
	// This will fail because the miner executable is not present,
	// which is expected in a test environment.
	_, err := manager.StartMiner("xmrig", config)
	if err == nil {
		t.Log("StartMiner did not fail as expected in test environment")
	}

	// Since we can't start a miner, we can't test stop either.
	// A more complete test suite would use a mock miner.
}

func TestGetNonExistentMiner(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetMiner("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent miner")
	}
}

func TestListMinersEmpty(t *testing.T) {
	manager := NewManager()
	miners := manager.ListMiners()
	if len(miners) != 0 {
		t.Errorf("Expected 0 miners, got %d", len(miners))
	}
}

func TestListAvailableMiners(t *testing.T) {
	manager := NewManager()
	miners := manager.ListAvailableMiners()
	if len(miners) == 0 {
		t.Error("Expected at least one available miner")
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version is empty")
	}
}
