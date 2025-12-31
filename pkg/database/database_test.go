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
		if err := InsertHashratePoint(minerName, minerType, p, ResolutionHigh); err != nil {
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
		if err := InsertHashratePoint(minerName, minerType, p, ResolutionHigh); err != nil {
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
