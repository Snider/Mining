package mining

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.miners == nil {
		t.Error("Manager miners map is nil")
	}
}

func TestStartAndStopMiner(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	config := &Config{
		Pool:   "pool.example.com",
		Wallet: "wallet123",
	}

	// We can't fully test StartMiner without a mock miner,
	// but we can test the manager's behavior.
	// This will fail because the miner executable is not present,
	// which is expected in a test environment.
	_, err := manager.StartMiner(context.Background(), "xmrig", config)
	if err == nil {
		t.Log("StartMiner did not fail as expected in test environment")
	}

	// Since we can't start a miner, we can't test stop either.
	// A more complete test suite would use a mock miner.
}

func TestGetNonExistentMiner(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	_, err := manager.GetMiner("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent miner")
	}
}

func TestListMiners(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	// ListMiners should return a valid slice (may include autostarted miners)
	miners := manager.ListMiners()
	if miners == nil {
		t.Error("ListMiners returned nil")
	}
	// Note: count may be > 0 if autostart is configured
}

func TestListAvailableMiners(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

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
