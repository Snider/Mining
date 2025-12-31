package database

import (
	"context"
	"time"
)

// HashrateStore defines the interface for hashrate data persistence.
// This interface allows for dependency injection and easier testing.
type HashrateStore interface {
	// InsertHashratePoint stores a hashrate measurement.
	// If ctx is nil, a default timeout will be used.
	InsertHashratePoint(ctx context.Context, minerName, minerType string, point HashratePoint, resolution Resolution) error

	// GetHashrateHistory retrieves hashrate history for a miner within a time range.
	GetHashrateHistory(minerName string, resolution Resolution, since, until time.Time) ([]HashratePoint, error)

	// GetHashrateStats retrieves aggregated statistics for a specific miner.
	GetHashrateStats(minerName string) (*HashrateStats, error)

	// GetAllMinerStats retrieves statistics for all miners.
	GetAllMinerStats() ([]HashrateStats, error)

	// Cleanup removes old data based on retention settings.
	Cleanup(retentionDays int) error

	// Close closes the store and releases resources.
	Close() error
}

// defaultStore implements HashrateStore using the global database connection.
// This provides backward compatibility while allowing interface-based usage.
type defaultStore struct{}

// DefaultStore returns a HashrateStore that uses the global database connection.
// This is useful for gradual migration from package-level functions to interface-based usage.
func DefaultStore() HashrateStore {
	return &defaultStore{}
}

func (s *defaultStore) InsertHashratePoint(ctx context.Context, minerName, minerType string, point HashratePoint, resolution Resolution) error {
	return InsertHashratePoint(ctx, minerName, minerType, point, resolution)
}

func (s *defaultStore) GetHashrateHistory(minerName string, resolution Resolution, since, until time.Time) ([]HashratePoint, error) {
	return GetHashrateHistory(minerName, resolution, since, until)
}

func (s *defaultStore) GetHashrateStats(minerName string) (*HashrateStats, error) {
	return GetHashrateStats(minerName)
}

func (s *defaultStore) GetAllMinerStats() ([]HashrateStats, error) {
	return GetAllMinerStats()
}

func (s *defaultStore) Cleanup(retentionDays int) error {
	return Cleanup(retentionDays)
}

func (s *defaultStore) Close() error {
	return Close()
}

// NopStore returns a HashrateStore that does nothing.
// Useful for testing or when database is disabled.
func NopStore() HashrateStore {
	return &nopStore{}
}

type nopStore struct{}

func (s *nopStore) InsertHashratePoint(ctx context.Context, minerName, minerType string, point HashratePoint, resolution Resolution) error {
	return nil
}

func (s *nopStore) GetHashrateHistory(minerName string, resolution Resolution, since, until time.Time) ([]HashratePoint, error) {
	return nil, nil
}

func (s *nopStore) GetHashrateStats(minerName string) (*HashrateStats, error) {
	return nil, nil
}

func (s *nopStore) GetAllMinerStats() ([]HashrateStats, error) {
	return nil, nil
}

func (s *nopStore) Cleanup(retentionDays int) error {
	return nil
}

func (s *nopStore) Close() error {
	return nil
}
