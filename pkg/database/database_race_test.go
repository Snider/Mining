package database

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// setupRaceTestDB creates a fresh database for race testing
func setupRaceTestDB(t *testing.T) func() {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "race_test.db")

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

// TestConcurrentHashrateInserts verifies that concurrent inserts
// don't cause race conditions
func TestConcurrentHashrateInserts(t *testing.T) {
	cleanup := setupRaceTestDB(t)
	defer cleanup()

	var wg sync.WaitGroup

	// 10 goroutines inserting points concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(minerIndex int) {
			defer wg.Done()
			minerName := "miner" + string(rune('A'+minerIndex))
			minerType := "xmrig"

			for j := 0; j < 100; j++ {
				point := HashratePoint{
					Timestamp: time.Now().Add(time.Duration(-j) * time.Second),
					Hashrate:  1000 + minerIndex*100 + j,
				}
				err := InsertHashratePoint(minerName, minerType, point, ResolutionHigh)
				if err != nil {
					t.Errorf("Insert error for %s: %v", minerName, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify data was inserted
	for i := 0; i < 10; i++ {
		minerName := "miner" + string(rune('A'+i))
		history, err := GetHashrateHistory(minerName, ResolutionHigh, time.Now().Add(-2*time.Minute), time.Now())
		if err != nil {
			t.Errorf("Failed to get history for %s: %v", minerName, err)
		}
		if len(history) == 0 {
			t.Errorf("Expected history for %s, got none", minerName)
		}
	}
}

// TestConcurrentInsertAndQuery verifies that concurrent reads and writes
// don't cause race conditions
func TestConcurrentInsertAndQuery(t *testing.T) {
	cleanup := setupRaceTestDB(t)
	defer cleanup()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				point := HashratePoint{
					Timestamp: time.Now(),
					Hashrate:  1000 + i,
				}
				InsertHashratePoint("concurrent-test", "xmrig", point, ResolutionHigh)
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// Multiple reader goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				select {
				case <-stop:
					return
				default:
					GetHashrateHistory("concurrent-test", ResolutionHigh, time.Now().Add(-time.Hour), time.Now())
					time.Sleep(2 * time.Millisecond)
				}
			}
		}()
	}

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentInsertAndCleanup verifies that cleanup doesn't race
// with ongoing inserts
func TestConcurrentInsertAndCleanup(t *testing.T) {
	cleanup := setupRaceTestDB(t)
	defer cleanup()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Continuous inserts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				// Insert some old data and some new data
				oldPoint := HashratePoint{
					Timestamp: time.Now().AddDate(0, 0, -10), // 10 days old
					Hashrate:  500 + i,
				}
				InsertHashratePoint("cleanup-test", "xmrig", oldPoint, ResolutionHigh)

				newPoint := HashratePoint{
					Timestamp: time.Now(),
					Hashrate:  1000 + i,
				}
				InsertHashratePoint("cleanup-test", "xmrig", newPoint, ResolutionHigh)
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// Periodic cleanup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			select {
			case <-stop:
				return
			default:
				Cleanup(7) // 7 day retention
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	// Let it run
	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentStats verifies that GetHashrateStats can be called
// concurrently without race conditions
func TestConcurrentStats(t *testing.T) {
	cleanup := setupRaceTestDB(t)
	defer cleanup()

	// Insert some test data
	minerName := "stats-test"
	for i := 0; i < 100; i++ {
		point := HashratePoint{
			Timestamp: time.Now().Add(time.Duration(-i) * time.Second),
			Hashrate:  1000 + i*10,
		}
		InsertHashratePoint(minerName, "xmrig", point, ResolutionHigh)
	}

	var wg sync.WaitGroup

	// Multiple goroutines querying stats
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				stats, err := GetHashrateStats(minerName)
				if err != nil {
					t.Errorf("Stats error: %v", err)
				}
				if stats != nil && stats.TotalPoints == 0 {
					// This is fine, data might be in flux
				}
			}
		}()
	}

	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentGetAllStats verifies that GetAllMinerStats can be called
// concurrently without race conditions
func TestConcurrentGetAllStats(t *testing.T) {
	cleanup := setupRaceTestDB(t)
	defer cleanup()

	// Insert data for multiple miners
	for m := 0; m < 5; m++ {
		minerName := "all-stats-" + string(rune('A'+m))
		for i := 0; i < 50; i++ {
			point := HashratePoint{
				Timestamp: time.Now().Add(time.Duration(-i) * time.Second),
				Hashrate:  1000 + m*100 + i,
			}
			InsertHashratePoint(minerName, "xmrig", point, ResolutionHigh)
		}
	}

	var wg sync.WaitGroup

	// Multiple goroutines querying all stats
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				_, err := GetAllMinerStats()
				if err != nil {
					t.Errorf("GetAllMinerStats error: %v", err)
				}
			}
		}()
	}

	// Concurrent inserts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			point := HashratePoint{
				Timestamp: time.Now(),
				Hashrate:  2000 + i,
			}
			InsertHashratePoint("all-stats-new", "xmrig", point, ResolutionHigh)
		}
	}()

	wg.Wait()

	// Test passes if no race detector warnings
}
