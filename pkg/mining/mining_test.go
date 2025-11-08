package mining

import (
	"testing"
	"time"
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

func TestStartMiner(t *testing.T) {
	manager := NewManager()

	config := MinerConfig{
		Name:      "test-miner",
		Algorithm: "sha256",
		Pool:      "pool.example.com",
		Wallet:    "wallet123",
	}

	miner, err := manager.StartMiner(config)
	if err != nil {
		t.Fatalf("StartMiner failed: %v", err)
	}

	if miner.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, miner.Name)
	}
	if miner.Status != "running" {
		t.Errorf("Expected status 'running', got %s", miner.Status)
	}
	if miner.ID == "" {
		t.Error("Miner ID is empty")
	}
}

func TestStartMinerWithoutName(t *testing.T) {
	manager := NewManager()

	config := MinerConfig{
		Algorithm: "sha256",
	}

	_, err := manager.StartMiner(config)
	if err == nil {
		t.Error("Expected error for miner without name")
	}
}

func TestStopMiner(t *testing.T) {
	manager := NewManager()

	config := MinerConfig{Name: "test-miner"}
	miner, _ := manager.StartMiner(config)

	err := manager.StopMiner(miner.ID)
	if err != nil {
		t.Fatalf("StopMiner failed: %v", err)
	}

	if miner.Status != "stopped" {
		t.Errorf("Expected status 'stopped', got %s", miner.Status)
	}
}

func TestStopNonExistentMiner(t *testing.T) {
	manager := NewManager()

	err := manager.StopMiner("non-existent")
	if err == nil {
		t.Error("Expected error for stopping non-existent miner")
	}
}

func TestGetMiner(t *testing.T) {
	manager := NewManager()

	config := MinerConfig{Name: "test-miner"}
	startedMiner, _ := manager.StartMiner(config)

	retrievedMiner, err := manager.GetMiner(startedMiner.ID)
	if err != nil {
		t.Fatalf("GetMiner failed: %v", err)
	}

	if retrievedMiner.ID != startedMiner.ID {
		t.Errorf("Expected ID %s, got %s", startedMiner.ID, retrievedMiner.ID)
	}
}

func TestGetNonExistentMiner(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetMiner("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent miner")
	}
}

func TestListMiners(t *testing.T) {
	manager := NewManager()

	// Start multiple miners
	for i := 0; i < 3; i++ {
		config := MinerConfig{Name: "test-miner"}
		_, _ = manager.StartMiner(config)
		time.Sleep(time.Millisecond) // Ensure unique IDs
	}

	miners := manager.ListMiners()
	if len(miners) != 3 {
		t.Errorf("Expected 3 miners, got %d", len(miners))
	}
}

func TestUpdateHashRate(t *testing.T) {
	manager := NewManager()

	config := MinerConfig{Name: "test-miner"}
	miner, _ := manager.StartMiner(config)

	newHashRate := 123.45
	err := manager.UpdateHashRate(miner.ID, newHashRate)
	if err != nil {
		t.Fatalf("UpdateHashRate failed: %v", err)
	}

	if miner.HashRate != newHashRate {
		t.Errorf("Expected hash rate %f, got %f", newHashRate, miner.HashRate)
	}
}

func TestUpdateHashRateNonExistent(t *testing.T) {
	manager := NewManager()

	err := manager.UpdateHashRate("non-existent", 100.0)
	if err == nil {
		t.Error("Expected error for updating non-existent miner")
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version is empty")
	}
}
