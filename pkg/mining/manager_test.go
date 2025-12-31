package mining

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// setupTestManager creates a new Manager and a dummy executable for tests.
// It also temporarily modifies the PATH to include the dummy executable's directory.
func setupTestManager(t *testing.T) *Manager {
	dummyDir := t.TempDir()
	executableName := "miner"
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
func TestStartMiner_Good(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	// Case 1: Successfully start a supported miner
	miner, err := m.StartMiner(context.Background(), "xmrig", config)
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
	_, err := m.StartMiner(context.Background(), "unsupported", config)
	if err == nil {
		t.Error("Expected an error when starting an unsupported miner, but got nil")
	}
}

func TestStartMiner_Ugly(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Use an algorithm to get consistent instance naming (xmrig-test_algo)
	// Without algo, each start gets a random suffix and won't be detected as duplicate
	config := &Config{
		HTTPPort: 9001, // Use a different port to avoid conflict
		Pool:     "test:1234",
		Wallet:   "testwallet",
		Algo:     "test_algo", // Consistent algo = consistent instance name
	}
	// Case 1: Successfully start a supported miner
	_, err := m.StartMiner(context.Background(), "xmrig", config)
	if err != nil {
		t.Fatalf("Expected to start miner, but got error: %v", err)
	}
	// Case 3: Attempt to start a duplicate miner (same algo = same instance name)
	_, err = m.StartMiner(context.Background(), "xmrig", config)
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
	miner, _ := m.StartMiner(context.Background(), "xmrig", config)
	err := m.StopMiner(context.Background(), miner.GetName())
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
	err := m.StopMiner(context.Background(), "nonexistent")
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
	startedMiner, _ := m.StartMiner(context.Background(), "xmrig", config)
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

	// Get initial count (may include autostarted miners from config)
	initialMiners := m.ListMiners()
	initialCount := len(initialMiners)

	// Case 2: List miners after starting one - should have one more
	config := &Config{
		HTTPPort: 9004,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}
	_, _ = m.StartMiner(context.Background(), "xmrig", config)
	miners := m.ListMiners()
	if len(miners) != initialCount+1 {
		t.Errorf("Expected %d miners (initial %d + 1), but got %d", initialCount+1, initialCount, len(miners))
	}
}

// TestManagerStop_Idempotent tests that Stop() can be called multiple times safely
func TestManagerStop_Idempotent(t *testing.T) {
	m := setupTestManager(t)

	// Start a miner
	config := &Config{
		HTTPPort: 9010,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}
	_, _ = m.StartMiner(context.Background(), "xmrig", config)

	// Call Stop() multiple times - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() panicked: %v", r)
		}
	}()

	m.Stop()
	m.Stop()
	m.Stop()

	// If we got here without panicking, the test passes
}

// TestStartMiner_CancelledContext tests that StartMiner respects context cancellation
func TestStartMiner_CancelledContext(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := &Config{
		HTTPPort: 9011,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	_, err := m.StartMiner(ctx, "xmrig", config)
	if err == nil {
		t.Error("Expected error when starting miner with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// TestStopMiner_CancelledContext tests that StopMiner respects context cancellation
func TestStopMiner_CancelledContext(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := m.StopMiner(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when stopping miner with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// TestManagerEventHub tests that SetEventHub works correctly
func TestManagerEventHub(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	eventHub := NewEventHub()
	go eventHub.Run()
	defer eventHub.Stop()

	m.SetEventHub(eventHub)

	// Get initial miner count (may have autostarted miners)
	initialCount := len(m.ListMiners())

	// Start a miner - should emit events
	config := &Config{
		HTTPPort: 9012,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}

	_, err := m.StartMiner(context.Background(), "xmrig", config)
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}

	// Give time for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify miner count increased by 1
	miners := m.ListMiners()
	if len(miners) != initialCount+1 {
		t.Errorf("Expected %d miners, got %d", initialCount+1, len(miners))
	}
}

// TestManagerShutdownTimeout tests the graceful shutdown timeout
func TestManagerShutdownTimeout(t *testing.T) {
	m := setupTestManager(t)

	// Start a miner
	config := &Config{
		HTTPPort: 9013,
		Pool:     "test:1234",
		Wallet:   "testwallet",
	}
	_, _ = m.StartMiner(context.Background(), "xmrig", config)

	// Stop should complete within a reasonable time
	done := make(chan struct{})
	go func() {
		m.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success - stopped in time
	case <-time.After(15 * time.Second):
		t.Error("Manager.Stop() took too long - possible shutdown issue")
	}
}
