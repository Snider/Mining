package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) func() {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		Enabled:       true,
		Path:          dbPath,
		RetentionDays: 7,
	}

	if err := Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return func() {
		Close()
		os.Remove(dbPath)
	}
}

func TestInitialize(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Database should be initialized
	dbMu.RLock()
	initialized := db != nil
	dbMu.RUnlock()

	if !initialized {
		t.Error("Database should be initialized")
	}
}

func TestInitialize_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}

	if err := Initialize(cfg); err != nil {
		t.Errorf("Initialize with disabled should not error: %v", err)
	}
}

func TestClose(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Close should not error
	if err := Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestHashrateStorage(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Store some hashrate data
	minerName := "test-miner"
	minerType := "xmrig"
	now := time.Now()

	points := []HashratePoint{
		{Timestamp: now.Add(-5 * time.Minute), Hashrate: 1000},
		{Timestamp: now.Add(-4 * time.Minute), Hashrate: 1100},
		{Timestamp: now.Add(-3 * time.Minute), Hashrate: 1200},
	}

	for _, p := range points {
		if err := InsertHashratePoint(nil, minerName, minerType, p, ResolutionHigh); err != nil {
			t.Fatalf("Failed to store hashrate point: %v", err)
		}
	}

	// Retrieve the data
	retrieved, err := GetHashrateHistory(minerName, ResolutionHigh, now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("Failed to get hashrate history: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 points, got %d", len(retrieved))
	}
}

func TestGetHashrateStats(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	minerName := "stats-test-miner"
	minerType := "xmrig"
	now := time.Now()

	// Store some test data
	points := []HashratePoint{
		{Timestamp: now.Add(-2 * time.Minute), Hashrate: 500},
		{Timestamp: now.Add(-1 * time.Minute), Hashrate: 1000},
		{Timestamp: now, Hashrate: 1500},
	}

	for _, p := range points {
		if err := InsertHashratePoint(nil, minerName, minerType, p, ResolutionHigh); err != nil {
			t.Fatalf("Failed to store point: %v", err)
		}
	}

	stats, err := GetHashrateStats(minerName)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalPoints != 3 {
		t.Errorf("Expected 3 total points, got %d", stats.TotalPoints)
	}

	// Average should be (500+1000+1500)/3 = 1000
	if stats.AverageRate != 1000 {
		t.Errorf("Expected average rate 1000, got %d", stats.AverageRate)
	}

	if stats.MaxRate != 1500 {
		t.Errorf("Expected max rate 1500, got %d", stats.MaxRate)
	}

	if stats.MinRate != 500 {
		t.Errorf("Expected min rate 500, got %d", stats.MinRate)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if !cfg.Enabled {
		t.Error("Default config should have Enabled=true")
	}

	if cfg.RetentionDays != 30 {
		t.Errorf("Expected default retention 30, got %d", cfg.RetentionDays)
	}
}

func TestCleanupRetention(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	minerName := "retention-test"
	minerType := "xmrig"
	now := time.Now()

	// Insert data at various ages:
	// - 35 days old (should be deleted with 30-day retention)
	// - 25 days old (should be kept with 30-day retention)
	// - 5 days old (should be kept)
	oldPoint := HashratePoint{
		Timestamp: now.AddDate(0, 0, -35),
		Hashrate:  100,
	}
	midPoint := HashratePoint{
		Timestamp: now.AddDate(0, 0, -25),
		Hashrate:  200,
	}
	newPoint := HashratePoint{
		Timestamp: now.AddDate(0, 0, -5),
		Hashrate:  300,
	}

	// Insert all points
	if err := InsertHashratePoint(nil, minerName, minerType, oldPoint, ResolutionHigh); err != nil {
		t.Fatalf("Failed to insert old point: %v", err)
	}
	if err := InsertHashratePoint(nil, minerName, minerType, midPoint, ResolutionHigh); err != nil {
		t.Fatalf("Failed to insert mid point: %v", err)
	}
	if err := InsertHashratePoint(nil, minerName, minerType, newPoint, ResolutionHigh); err != nil {
		t.Fatalf("Failed to insert new point: %v", err)
	}

	// Verify all 3 points exist
	history, err := GetHashrateHistory(minerName, ResolutionHigh, now.AddDate(0, 0, -40), now)
	if err != nil {
		t.Fatalf("Failed to get history before cleanup: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("Expected 3 points before cleanup, got %d", len(history))
	}

	// Run cleanup with 30-day retention
	if err := Cleanup(30); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify only 2 points remain (35-day old point should be deleted)
	history, err = GetHashrateHistory(minerName, ResolutionHigh, now.AddDate(0, 0, -40), now)
	if err != nil {
		t.Fatalf("Failed to get history after cleanup: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("Expected 2 points after cleanup, got %d", len(history))
	}

	// Verify the remaining points are the mid and new ones
	for _, point := range history {
		if point.Hashrate == 100 {
			t.Error("Old point (100 H/s) should have been deleted")
		}
	}
}

func TestGetHashrateHistoryTimeRange(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	minerName := "timerange-test"
	minerType := "xmrig"
	now := time.Now()

	// Insert points at specific times
	times := []time.Duration{
		-10 * time.Minute,
		-8 * time.Minute,
		-6 * time.Minute,
		-4 * time.Minute,
		-2 * time.Minute,
	}

	for i, offset := range times {
		point := HashratePoint{
			Timestamp: now.Add(offset),
			Hashrate:  1000 + i*100,
		}
		if err := InsertHashratePoint(nil, minerName, minerType, point, ResolutionHigh); err != nil {
			t.Fatalf("Failed to insert point: %v", err)
		}
	}

	// Query for middle range (should get 3 points: -8, -6, -4 minutes)
	since := now.Add(-9 * time.Minute)
	until := now.Add(-3 * time.Minute)
	history, err := GetHashrateHistory(minerName, ResolutionHigh, since, until)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 points in range, got %d", len(history))
	}

	// Query boundary condition - exact timestamp match
	exactSince := now.Add(-6 * time.Minute)
	exactUntil := now.Add(-6 * time.Minute).Add(time.Second)
	history, err = GetHashrateHistory(minerName, ResolutionHigh, exactSince, exactUntil)
	if err != nil {
		t.Fatalf("Failed to get exact history: %v", err)
	}

	// Should get at least 1 point
	if len(history) < 1 {
		t.Error("Expected at least 1 point at exact boundary")
	}
}

func TestMultipleMinerStats(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()

	// Create data for multiple miners
	miners := []struct {
		name      string
		hashrates []int
	}{
		{"miner-A", []int{1000, 1100, 1200}},
		{"miner-B", []int{2000, 2100, 2200}},
		{"miner-C", []int{3000, 3100, 3200}},
	}

	for _, m := range miners {
		for i, hr := range m.hashrates {
			point := HashratePoint{
				Timestamp: now.Add(time.Duration(-i) * time.Minute),
				Hashrate:  hr,
			}
			if err := InsertHashratePoint(nil, m.name, "xmrig", point, ResolutionHigh); err != nil {
				t.Fatalf("Failed to insert point for %s: %v", m.name, err)
			}
		}
	}

	// Get all miner stats
	allStats, err := GetAllMinerStats()
	if err != nil {
		t.Fatalf("Failed to get all stats: %v", err)
	}

	if len(allStats) != 3 {
		t.Errorf("Expected stats for 3 miners, got %d", len(allStats))
	}

	// Verify each miner's stats
	statsMap := make(map[string]HashrateStats)
	for _, s := range allStats {
		statsMap[s.MinerName] = s
	}

	// Check miner-A: avg = (1000+1100+1200)/3 = 1100
	if s, ok := statsMap["miner-A"]; ok {
		if s.AverageRate != 1100 {
			t.Errorf("miner-A: expected avg 1100, got %d", s.AverageRate)
		}
	} else {
		t.Error("miner-A stats not found")
	}

	// Check miner-C: avg = (3000+3100+3200)/3 = 3100
	if s, ok := statsMap["miner-C"]; ok {
		if s.AverageRate != 3100 {
			t.Errorf("miner-C: expected avg 3100, got %d", s.AverageRate)
		}
	} else {
		t.Error("miner-C stats not found")
	}
}

func TestIsInitialized(t *testing.T) {
	// Before initialization
	Close() // Ensure clean state
	if isInitialized() {
		t.Error("Should not be initialized before Initialize()")
	}

	cleanup := setupTestDB(t)
	defer cleanup()

	// After initialization
	if !isInitialized() {
		t.Error("Should be initialized after Initialize()")
	}

	// After close
	Close()
	if isInitialized() {
		t.Error("Should not be initialized after Close()")
	}
}

func TestSchemaCreation(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Verify tables exist by querying sqlite_master
	dbMu.RLock()
	defer dbMu.RUnlock()

	// Check hashrate_history table
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='hashrate_history'").Scan(&tableName)
	if err != nil {
		t.Errorf("hashrate_history table should exist: %v", err)
	}

	// Check miner_sessions table
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='miner_sessions'").Scan(&tableName)
	if err != nil {
		t.Errorf("miner_sessions table should exist: %v", err)
	}

	// Verify indexes exist
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_hashrate_miner_time'").Scan(&indexName)
	if err != nil {
		t.Errorf("idx_hashrate_miner_time index should exist: %v", err)
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_sessions_miner'").Scan(&indexName)
	if err != nil {
		t.Errorf("idx_sessions_miner index should exist: %v", err)
	}
}

func TestReInitializeExistingDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "reinit_test.db")

	cfg := Config{
		Enabled:       true,
		Path:          dbPath,
		RetentionDays: 7,
	}

	// First initialization
	if err := Initialize(cfg); err != nil {
		t.Fatalf("First initialization failed: %v", err)
	}

	// Insert some data
	minerName := "reinit-test-miner"
	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1234,
	}
	if err := InsertHashratePoint(nil, minerName, "xmrig", point, ResolutionHigh); err != nil {
		t.Fatalf("Failed to insert point: %v", err)
	}

	// Close and re-initialize (simulates app restart)
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Re-initialize with same path
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Re-initialization failed: %v", err)
	}
	defer func() {
		Close()
		os.Remove(dbPath)
	}()

	// Verify data persisted
	history, err := GetHashrateHistory(minerName, ResolutionHigh, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to get history after reinit: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 point after reinit, got %d", len(history))
	}

	if len(history) > 0 && history[0].Hashrate != 1234 {
		t.Errorf("Expected hashrate 1234, got %d", history[0].Hashrate)
	}
}

func TestConcurrentDatabaseAccess(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	const numGoroutines = 10
	const numOpsPerGoroutine = 20

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*numOpsPerGoroutine)

	now := time.Now()

	// Launch multiple goroutines doing concurrent reads/writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			minerName := "concurrent-miner-" + string(rune('A'+id))
			for j := 0; j < numOpsPerGoroutine; j++ {
				// Write
				point := HashratePoint{
					Timestamp: now.Add(time.Duration(-j) * time.Second),
					Hashrate:  1000 + j,
				}
				if err := InsertHashratePoint(nil, minerName, "xmrig", point, ResolutionHigh); err != nil {
					errors <- err
				}

				// Read
				_, err := GetHashrateHistory(minerName, ResolutionHigh, now.Add(-time.Hour), now)
				if err != nil {
					errors <- err
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
		errCount++
	}

	if errCount > 0 {
		t.Errorf("Got %d errors during concurrent access", errCount)
	}
}
