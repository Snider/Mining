package mining

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
)

// setupTestManager creates a new Manager and a dummy executable for tests.
// It also temporarily modifies the PATH to include the dummy executable's directory.
func setupTestManager(t *testing.T) *Manager {
	// Isolate config directory for this test
	tempConfigDir := t.TempDir()

	// Backup original xdg paths
	origConfigHome := xdg.ConfigHome
	origDataHome := xdg.DataHome
	origConfigDirs := xdg.ConfigDirs

	// Set new paths
	xdg.ConfigHome = tempConfigDir
	xdg.DataHome = tempConfigDir
	xdg.ConfigDirs = []string{tempConfigDir}

	// Restore on cleanup
	t.Cleanup(func() {
		xdg.ConfigHome = origConfigHome
		xdg.DataHome = origDataHome
		xdg.ConfigDirs = origConfigDirs
	})

	dummyDir := t.TempDir()
	executableName := "xmrig"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	dummyPath := filepath.Join(dummyDir, executableName)

	// Create a script that does nothing but exit, to simulate the miner executable
	var script []byte
	if runtime.GOOS == "windows" {
		script = []byte("@echo off\r\nexit 0")
	} else {
		script = []byte("#!/bin/sh\nexit 0")
	}

	if err := os.WriteFile(dummyPath, script, 0755); err != nil {
		t.Fatalf("Failed to create dummy miner executable: %v", err)
	}

	// Prepend the dummy directory to the PATH
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() {
		os.Setenv("PATH", originalPath)
	})
	os.Setenv("PATH", dummyDir+string(os.PathListSeparator)+originalPath)

	m := NewManager()
	// Clear any autostarted miners to ensure clean state
	m.mu.Lock()
	for name, miner := range m.miners {
		_ = miner.Stop()
		delete(m.miners, name)
	}
	m.mu.Unlock()

	return m
}

// TestStartMiner tests the StartMiner function
func TestStartMiner_Good(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
		Algo:     "rx/0", // Use Algo to ensure deterministic naming for duplicate check
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	// Case 1: Successfully start a supported miner
	miner, err := m.StartMiner("xmrig", config)
	if err != nil {
		t.Fatalf("Expected to start miner, but got error: %v", err)
	}
	if miner == nil {
		t.Fatal("Expected miner to be non-nil, but it was")
	}
	if _, exists := m.miners[miner.GetName()]; !exists {
		t.Errorf("Miner %s was not added to the manager's list", miner.GetName())
	}
}

func TestStartMiner_Bad(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	// Case 2: Attempt to start an unsupported miner
	_, err := m.StartMiner("unsupported", config)
	if err == nil {
		t.Error("Expected an error when starting an unsupported miner, but got nil")
	}
}

func TestStartMiner_Ugly(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
		Algo:     "rx/0", // Use Algo to ensure deterministic naming for duplicate check
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}
	// Case 1: Successfully start a supported miner
	_, err := m.StartMiner("xmrig", config)
	if err != nil {
		t.Fatalf("Expected to start miner, but got error: %v", err)
	}
	// Case 3: Attempt to start a duplicate miner
	_, err = m.StartMiner("xmrig", config)
	if err == nil {
		t.Error("Expected an error when starting a duplicate miner, but got nil")
	}
}

// TestStopMiner tests the StopMiner function
func TestStopMiner_Good(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9002,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	// Case 1: Stop a running miner
	miner, _ := m.StartMiner("xmrig", config)
	err := m.StopMiner(miner.GetName())
	if err != nil {
		t.Fatalf("Expected to stop miner, but got error: %v", err)
	}
	if _, exists := m.miners[miner.GetName()]; exists {
		t.Errorf("Miner %s was not removed from the manager's list", miner.GetName())
	}
}

func TestStopMiner_Bad(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Case 2: Attempt to stop a non-existent miner
	err := m.StopMiner("nonexistent")
	if err == nil {
		t.Error("Expected an error when stopping a non-existent miner, but got nil")
	}
}

// TestGetMiner tests the GetMiner function
func TestGetMiner_Good(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9003,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	// Case 1: Get an existing miner
	startedMiner, _ := m.StartMiner("xmrig", config)
	retrievedMiner, err := m.GetMiner(startedMiner.GetName())
	if err != nil {
		t.Fatalf("Expected to get miner, but got error: %v", err)
	}
	if retrievedMiner.GetName() != startedMiner.GetName() {
		t.Errorf("Expected to get miner %s, but got %s", startedMiner.GetName(), retrievedMiner.GetName())
	}
}

func TestGetMiner_Bad(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Case 2: Attempt to get a non-existent miner
	_, err := m.GetMiner("nonexistent")
	if err == nil {
		t.Error("Expected an error when getting a non-existent miner, but got nil")
	}
}

// TestListMiners tests the ListMiners function
func TestListMiners_Good(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Case 1: List miners when empty
	miners := m.ListMiners()
	if len(miners) != 0 {
		t.Errorf("Expected 0 miners, but got %d", len(miners))
	}

	// Case 2: List miners when not empty
	config := &Config{
		HTTPPort: 9004,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}
	_, _ = m.StartMiner("xmrig", config)
	miners = m.ListMiners()
	if len(miners) != 1 {
		t.Errorf("Expected 1 miner, but got %d", len(miners))
	}
}
