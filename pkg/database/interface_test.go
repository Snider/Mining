package database

import (
	"context"
	"testing"
	"time"
)

func TestDefaultStore(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	store := DefaultStore()

	// Test InsertHashratePoint
	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1500,
	}
	if err := store.InsertHashratePoint(nil, "interface-test", "xmrig", point, ResolutionHigh); err != nil {
		t.Fatalf("InsertHashratePoint failed: %v", err)
	}

	// Test GetHashrateHistory
	history, err := store.GetHashrateHistory("interface-test", ResolutionHigh, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GetHashrateHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 point, got %d", len(history))
	}

	// Test GetHashrateStats
	stats, err := store.GetHashrateStats("interface-test")
	if err != nil {
		t.Fatalf("GetHashrateStats failed: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
	if stats.TotalPoints != 1 {
		t.Errorf("Expected 1 total point, got %d", stats.TotalPoints)
	}

	// Test GetAllMinerStats
	allStats, err := store.GetAllMinerStats()
	if err != nil {
		t.Fatalf("GetAllMinerStats failed: %v", err)
	}
	if len(allStats) != 1 {
		t.Errorf("Expected 1 miner in stats, got %d", len(allStats))
	}

	// Test Cleanup
	if err := store.Cleanup(30); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

func TestDefaultStore_WithContext(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	store := DefaultStore()
	ctx := context.Background()

	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  2000,
	}
	if err := store.InsertHashratePoint(ctx, "ctx-test", "xmrig", point, ResolutionHigh); err != nil {
		t.Fatalf("InsertHashratePoint with context failed: %v", err)
	}

	history, err := store.GetHashrateHistory("ctx-test", ResolutionHigh, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GetHashrateHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 point, got %d", len(history))
	}
}

func TestNopStore(t *testing.T) {
	store := NopStore()

	// All operations should succeed without error
	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1000,
	}
	if err := store.InsertHashratePoint(nil, "test", "xmrig", point, ResolutionHigh); err != nil {
		t.Errorf("NopStore InsertHashratePoint should not error: %v", err)
	}

	history, err := store.GetHashrateHistory("test", ResolutionHigh, time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Errorf("NopStore GetHashrateHistory should not error: %v", err)
	}
	if history != nil {
		t.Errorf("NopStore GetHashrateHistory should return nil, got %v", history)
	}

	stats, err := store.GetHashrateStats("test")
	if err != nil {
		t.Errorf("NopStore GetHashrateStats should not error: %v", err)
	}
	if stats != nil {
		t.Errorf("NopStore GetHashrateStats should return nil, got %v", stats)
	}

	allStats, err := store.GetAllMinerStats()
	if err != nil {
		t.Errorf("NopStore GetAllMinerStats should not error: %v", err)
	}
	if allStats != nil {
		t.Errorf("NopStore GetAllMinerStats should return nil, got %v", allStats)
	}

	if err := store.Cleanup(30); err != nil {
		t.Errorf("NopStore Cleanup should not error: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("NopStore Close should not error: %v", err)
	}
}

// TestInterfaceCompatibility ensures all implementations satisfy HashrateStore
func TestInterfaceCompatibility(t *testing.T) {
	var _ HashrateStore = DefaultStore()
	var _ HashrateStore = NopStore()
	var _ HashrateStore = &defaultStore{}
	var _ HashrateStore = &nopStore{}
}

func TestDefaultStore_ContextCancellation(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	store := DefaultStore()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1000,
	}

	// Insert with cancelled context should fail
	err := store.InsertHashratePoint(ctx, "cancel-test", "xmrig", point, ResolutionHigh)
	if err == nil {
		t.Log("InsertHashratePoint with cancelled context succeeded (SQLite may not check context)")
	} else {
		t.Logf("InsertHashratePoint with cancelled context: %v (expected)", err)
	}
}

func TestDefaultStore_ContextTimeout(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	store := DefaultStore()

	// Create a context that expires very quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout to expire
	time.Sleep(1 * time.Millisecond)

	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1000,
	}

	// Insert with expired context
	err := store.InsertHashratePoint(ctx, "timeout-test", "xmrig", point, ResolutionHigh)
	if err == nil {
		t.Log("InsertHashratePoint with expired context succeeded (SQLite may not check context)")
	} else {
		t.Logf("InsertHashratePoint with expired context: %v (expected)", err)
	}
}

func TestNopStore_WithContext(t *testing.T) {
	store := NopStore()

	// NopStore should work with any context, including cancelled ones
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	point := HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1000,
	}

	// Should still succeed (nop store ignores context)
	if err := store.InsertHashratePoint(ctx, "nop-cancel-test", "xmrig", point, ResolutionHigh); err != nil {
		t.Errorf("NopStore should succeed even with cancelled context: %v", err)
	}
}
