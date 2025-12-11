package mining

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// MockRoundTripper is a mock implementation of http.RoundTripper for testing.
type MockRoundTripper func(req *http.Request) *http.Response

// RoundTrip executes a single HTTP transaction, returning a Response for the given Request.
func (f MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn MockRoundTripper) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

// helper function to create a temporary directory for testing
func tempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNewXMRigMiner_Good(t *testing.T) {
	miner := NewXMRigMiner()
	if miner == nil {
		t.Fatal("NewXMRigMiner returned nil")
	}
	if miner.Name != "xmrig" {
		t.Errorf("Expected miner name to be 'xmrig', got '%s'", miner.Name)
	}
	if miner.Version != "latest" {
		t.Errorf("Expected miner version to be 'latest', got '%s'", miner.Version)
	}
	if !miner.API.Enabled {
		t.Error("Expected API to be enabled by default")
	}
}

func TestXMRigMiner_GetName_Good(t *testing.T) {
	miner := NewXMRigMiner()
	if name := miner.GetName(); name != "xmrig" {
		t.Errorf("Expected GetName() to return 'xmrig', got '%s'", name)
	}
}

func TestXMRigMiner_GetLatestVersion_Good(t *testing.T) {
	originalClient := httpClient
	httpClient = newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() != "https://api.github.com/repos/xmrig/xmrig/releases/latest" {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("Not Found")),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"tag_name": "v6.18.0"}`)),
			Header:     make(http.Header),
		}
	})
	defer func() { httpClient = originalClient }()

	miner := NewXMRigMiner()
	version, err := miner.GetLatestVersion()
	if err != nil {
		t.Fatalf("GetLatestVersion() returned an error: %v", err)
	}
	if version != "v6.18.0" {
		t.Errorf("Expected version 'v6.18.0', got '%s'", version)
	}
}

func TestXMRigMiner_GetLatestVersion_Bad(t *testing.T) {
	originalClient := httpClient
	httpClient = newTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not Found")),
			Header:     make(http.Header),
		}
	})
	defer func() { httpClient = originalClient }()

	miner := NewXMRigMiner()
	_, err := miner.GetLatestVersion()
	if err == nil {
		t.Fatalf("GetLatestVersion() did not return an error")
	}
}

func TestXMRigMiner_Start_Stop_Good(t *testing.T) {
	// Create a temporary directory for the dummy executable
	tmpDir := t.TempDir()
	dummyExePath := filepath.Join(tmpDir, "xmrig")
	if runtime.GOOS == "windows" {
		dummyExePath += ".bat"
		// Create a dummy batch file for Windows
		if err := os.WriteFile(dummyExePath, []byte("@echo off\n"), 0755); err != nil {
			t.Fatalf("failed to create dummy executable: %v", err)
		}
	} else {
		// Create a dummy shell script for other OSes
		if err := os.WriteFile(dummyExePath, []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatalf("failed to create dummy executable: %v", err)
		}
	}

	miner := NewXMRigMiner()
	miner.MinerBinary = dummyExePath
	miner.API.ListenPort = 12345 // Set a port for testing

	config := &Config{
		Pool:   "test:1234",
		Wallet: "testwallet",
	}

	err := miner.Start(config)
	if err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	if !miner.Running {
		t.Fatal("Miner is not running after Start()")
	}

	err = miner.Stop()
	if err != nil {
		// On some systems, stopping a process that quickly exits can error. We log but don't fail.
		t.Logf("Stop() returned an error (often benign in tests): %v", err)
	}

	// Give a moment for the process to be marked as not running
	time.Sleep(100 * time.Millisecond)

	miner.mu.Lock()
	if miner.Running {
		miner.mu.Unlock()
		t.Fatal("Miner is still running after Stop()")
	}
	miner.mu.Unlock()
}

func TestXMRigMiner_Start_Stop_Bad(t *testing.T) {
	miner := NewXMRigMiner()
	miner.MinerBinary = "nonexistent"

	config := &Config{
		Pool:   "test:1234",
		Wallet: "testwallet",
	}

	err := miner.Start(config)
	if err == nil {
		t.Fatalf("Start() did not return an error")
	}
}

func TestXMRigMiner_GetStats_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		summary := XMRigSummary{
			Hashrate: struct {
				Total   []float64 `json:"total"`
				Highest float64   `json:"highest"`
			}{Total: []float64{123.45}},
			Results: struct {
				DiffCurrent int   `json:"diff_current"`
				SharesGood  int   `json:"shares_good"`
				SharesTotal int   `json:"shares_total"`
				AvgTime     int   `json:"avg_time"`
				AvgTimeMS   int   `json:"avg_time_ms"`
				HashesTotal int   `json:"hashes_total"`
				Best        []int `json:"best"`
			}{SharesGood: 10, SharesTotal: 12},
			Uptime: 600,
			Algo:   "rx/0",
		}
		json.NewEncoder(w).Encode(summary)
	}))
	defer server.Close()

	originalHTTPClient := httpClient
	httpClient = server.Client()
	defer func() { httpClient = originalHTTPClient }()

	miner := NewXMRigMiner()
	miner.Running = true // Mock running state
	miner.API.ListenHost = strings.TrimPrefix(server.URL, "http://")
	miner.API.ListenHost, miner.API.ListenPort = server.Listener.Addr().String(), 0
	parts := strings.Split(server.Listener.Addr().String(), ":")
	miner.API.ListenHost = parts[0]
	fmt.Sscanf(parts[1], "%d", &miner.API.ListenPort)

	stats, err := miner.GetStats()
	if err != nil {
		t.Fatalf("GetStats() returned an error: %v", err)
	}
	if stats.Hashrate != 123 {
		t.Errorf("Expected hashrate 123, got %d", stats.Hashrate)
	}
	if stats.Shares != 10 {
		t.Errorf("Expected 10 shares, got %d", stats.Shares)
	}
	if stats.Rejected != 2 {
		t.Errorf("Expected 2 rejected shares, got %d", stats.Rejected)
	}
	if stats.Uptime != 600 {
		t.Errorf("Expected uptime 600, got %d", stats.Uptime)
	}
	if stats.Algorithm != "rx/0" {
		t.Errorf("Expected algorithm 'rx/0', got '%s'", stats.Algorithm)
	}
}

func TestXMRigMiner_GetStats_Bad(t *testing.T) {
	// Don't start a server, so the API call will fail
	miner := NewXMRigMiner()
	miner.Running = true // Mock running state
	miner.API.ListenHost = "127.0.0.1"
	miner.API.ListenPort = 9999 // A port that is unlikely to be in use

	_, err := miner.GetStats()
	if err == nil {
		t.Fatalf("GetStats() did not return an error")
	}
}

func TestXMRigMiner_HashrateHistory_Good(t *testing.T) {
	miner := NewXMRigMiner()
	now := time.Now()

	// Add high-resolution points
	for i := 0; i < 10; i++ {
		miner.AddHashratePoint(HashratePoint{Timestamp: now.Add(time.Duration(i) * time.Second), Hashrate: 100 + i})
	}

	history := miner.GetHashrateHistory()
	if len(history) != 10 {
		t.Fatalf("Expected 10 hashrate points, got %d", len(history))
	}

	// Test ReduceHashrateHistory
	// Move time forward to make some points eligible for reduction
	future := now.Add(HighResolutionDuration + 30*time.Second)
	miner.ReduceHashrateHistory(future)

	// After reduction, high-res history should be smaller
	if len(miner.HashrateHistory) >= 10 {
		t.Errorf("High-res history not reduced, size: %d", len(miner.HashrateHistory))
	}
	if len(miner.LowResHashrateHistory) == 0 {
		t.Error("Low-res history not populated")
	}

	combinedHistory := miner.GetHashrateHistory()
	if len(combinedHistory) == 0 {
		t.Error("GetHashrateHistory returned empty slice after reduction")
	}
}
