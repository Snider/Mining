package database

import (
	"context"
	"fmt"
	"log"
	"time"
)

// dbOperationTimeout is the maximum time for database operations
const dbOperationTimeout = 30 * time.Second

// parseSQLiteTimestamp parses timestamp strings from SQLite which may use various formats.
// Logs a warning if parsing fails and returns zero time.
func parseSQLiteTimestamp(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	// Try common SQLite timestamp formats
	formats := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	log.Printf("Warning: failed to parse timestamp '%s' from database", s)
	return time.Time{}
}

// Resolution indicates the data resolution type
type Resolution string

const (
	ResolutionHigh Resolution = "high" // 10-second intervals
	ResolutionLow  Resolution = "low"  // 1-minute averages
)

// HashratePoint represents a single hashrate measurement
type HashratePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Hashrate  int       `json:"hashrate"`
}

// InsertHashratePoint stores a hashrate measurement in the database
func InsertHashratePoint(minerName, minerType string, point HashratePoint, resolution Resolution) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil // DB not enabled, silently skip
	}

	_, err := db.Exec(`
		INSERT INTO hashrate_history (miner_name, miner_type, timestamp, hashrate, resolution)
		VALUES (?, ?, ?, ?, ?)
	`, minerName, minerType, point.Timestamp, point.Hashrate, string(resolution))

	return err
}

// InsertHashratePoints stores multiple hashrate measurements in a single transaction
func InsertHashratePoints(minerName, minerType string, points []HashratePoint, resolution Resolution) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	if len(points) == 0 {
		return nil
	}

	// Use context with timeout to prevent hanging on locked database
	ctx, cancel := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO hashrate_history (miner_name, miner_type, timestamp, hashrate, resolution)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, point := range points {
		_, err := stmt.ExecContext(ctx, minerName, minerType, point.Timestamp, point.Hashrate, string(resolution))
		if err != nil {
			return fmt.Errorf("failed to insert point: %w", err)
		}
	}

	return tx.Commit()
}

// GetHashrateHistory retrieves hashrate history for a miner within a time range
func GetHashrateHistory(minerName string, resolution Resolution, since, until time.Time) ([]HashratePoint, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT timestamp, hashrate
		FROM hashrate_history
		WHERE miner_name = ?
		  AND resolution = ?
		  AND timestamp >= ?
		  AND timestamp <= ?
		ORDER BY timestamp ASC
	`, minerName, string(resolution), since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query hashrate history: %w", err)
	}
	defer rows.Close()

	var points []HashratePoint
	for rows.Next() {
		var point HashratePoint
		if err := rows.Scan(&point.Timestamp, &point.Hashrate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		points = append(points, point)
	}

	return points, rows.Err()
}

// GetLatestHashrate retrieves the most recent hashrate for a miner
func GetLatestHashrate(minerName string) (*HashratePoint, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	var point HashratePoint
	err := db.QueryRow(`
		SELECT timestamp, hashrate
		FROM hashrate_history
		WHERE miner_name = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, minerName).Scan(&point.Timestamp, &point.Hashrate)

	if err != nil {
		return nil, nil // Not found is not an error
	}

	return &point, nil
}

// GetAverageHashrate calculates the average hashrate for a miner in a time range
func GetAverageHashrate(minerName string, since, until time.Time) (int, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return 0, nil
	}

	var avg float64
	err := db.QueryRow(`
		SELECT COALESCE(AVG(hashrate), 0)
		FROM hashrate_history
		WHERE miner_name = ?
		  AND timestamp >= ?
		  AND timestamp <= ?
	`, minerName, since, until).Scan(&avg)

	return int(avg), err
}

// GetMaxHashrate retrieves the maximum hashrate for a miner in a time range
func GetMaxHashrate(minerName string, since, until time.Time) (int, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return 0, nil
	}

	var max int
	err := db.QueryRow(`
		SELECT COALESCE(MAX(hashrate), 0)
		FROM hashrate_history
		WHERE miner_name = ?
		  AND timestamp >= ?
		  AND timestamp <= ?
	`, minerName, since, until).Scan(&max)

	return max, err
}

// GetHashrateStats retrieves aggregated stats for a miner
type HashrateStats struct {
	MinerName   string    `json:"minerName"`
	TotalPoints int       `json:"totalPoints"`
	AverageRate int       `json:"averageRate"`
	MaxRate     int       `json:"maxRate"`
	MinRate     int       `json:"minRate"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
}

func GetHashrateStats(minerName string) (*HashrateStats, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	// First check if there are any rows for this miner
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM hashrate_history WHERE miner_name = ?`, minerName).Scan(&count)
	if err != nil {
		return nil, err
	}

	// No data for this miner
	if count == 0 {
		return nil, nil
	}

	var stats HashrateStats
	stats.MinerName = minerName

	// SQLite returns timestamps as strings, so scan them as strings first
	var firstSeenStr, lastSeenStr string
	err = db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(AVG(hashrate), 0),
			COALESCE(MAX(hashrate), 0),
			COALESCE(MIN(hashrate), 0),
			MIN(timestamp),
			MAX(timestamp)
		FROM hashrate_history
		WHERE miner_name = ?
	`, minerName).Scan(
		&stats.TotalPoints,
		&stats.AverageRate,
		&stats.MaxRate,
		&stats.MinRate,
		&firstSeenStr,
		&lastSeenStr,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps using helper that logs errors
	stats.FirstSeen = parseSQLiteTimestamp(firstSeenStr)
	stats.LastSeen = parseSQLiteTimestamp(lastSeenStr)

	return &stats, nil
}

// GetAllMinerStats retrieves stats for all miners
func GetAllMinerStats() ([]HashrateStats, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT
			miner_name,
			COUNT(*),
			COALESCE(AVG(hashrate), 0),
			COALESCE(MAX(hashrate), 0),
			COALESCE(MIN(hashrate), 0),
			MIN(timestamp),
			MAX(timestamp)
		FROM hashrate_history
		GROUP BY miner_name
		ORDER BY miner_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allStats []HashrateStats
	for rows.Next() {
		var stats HashrateStats
		var firstSeenStr, lastSeenStr string
		if err := rows.Scan(
			&stats.MinerName,
			&stats.TotalPoints,
			&stats.AverageRate,
			&stats.MaxRate,
			&stats.MinRate,
			&firstSeenStr,
			&lastSeenStr,
		); err != nil {
			return nil, err
		}
		// Parse timestamps using helper that logs errors
		stats.FirstSeen = parseSQLiteTimestamp(firstSeenStr)
		stats.LastSeen = parseSQLiteTimestamp(lastSeenStr)
		allStats = append(allStats, stats)
	}

	return allStats, rows.Err()
}

// CleanupOldData removes hashrate data older than the specified duration
func CleanupOldData(resolution Resolution, maxAge time.Duration) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	cutoff := time.Now().Add(-maxAge)
	_, err := db.Exec(`
		DELETE FROM hashrate_history
		WHERE resolution = ?
		  AND timestamp < ?
	`, string(resolution), cutoff)

	return err
}
