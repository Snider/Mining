package mining

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// setupTestManager creates a new Manager and a dummy executable for tests.
// It also temporarily modifies the PATH to include the dummy executable's directory.
func setupTestManager(t *testing.T) *Manager {
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

	return NewManager()
}

// TestStartMiner tests the StartMiner function
func TestStartMiner(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
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

	// Case 2: Attempt to start an unsupported miner
	_, err = m.StartMiner("unsupported", config)
	if err == nil {
		t.Error("Expected an error when starting an unsupported miner, but got nil")
	}

	// Case 3: Attempt to start a duplicate miner
	_, err = m.StartMiner("xmrig", config)
	if err == nil {
		t.Error("Expected an error when starting a duplicate miner, but got nil")
	}
}

// TestStopMiner tests the StopMiner function
func TestStopMiner(t *testing.T) {
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

	// Case 2: Attempt to stop a non-existent miner
	err = m.StopMiner("nonexistent")
	if err == nil {
		t.Error("Expected an error when stopping a non-existent miner, but got nil")
	}
}
