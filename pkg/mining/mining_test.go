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
	t.Skip("Skipping test that attempts to run miner process")
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
