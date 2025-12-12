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

	// Create a script that prints version and exits
	var script []byte
	if runtime.GOOS == "windows" {
		script = []byte("@echo off\necho XMRig 6.24.0\n")
	} else {
		script = []byte("#!/bin/sh\necho 'XMRig 6.24.0'\n")
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
	t.Skip("Skipping test that runs miner process as per request")
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
	t.Skip("Skipping test that runs miner process")
}

// TestStopMiner tests the StopMiner function
func TestStopMiner_Good(t *testing.T) {
	t.Skip("Skipping test that runs miner process")
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

	// Case 1: Get an existing miner (manually injected)
	miner := NewXMRigMiner()
	// Set name to match what StartMiner would produce usually ("xmrig")
	// Since we inject it, we can use the default name or set one.
	miner.Name = "xmrig-test"
	m.mu.Lock()
	m.miners["xmrig-test"] = miner
	m.mu.Unlock()

	retrievedMiner, err := m.GetMiner("xmrig-test")
	if err != nil {
		t.Fatalf("Expected to get miner, but got error: %v", err)
	}
	if retrievedMiner.GetName() != "xmrig-test" {
		t.Errorf("Expected to get miner 'xmrig-test', but got %s", retrievedMiner.GetName())
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

	// Case 2: List miners when not empty (manually injected)
	miner := NewXMRigMiner()
	miner.Name = "xmrig-test"
	m.mu.Lock()
	m.miners["xmrig-test"] = miner
	m.mu.Unlock()

	miners = m.ListMiners()
	if len(miners) != 1 {
		t.Errorf("Expected 1 miner, but got %d", len(miners))
	}
}
