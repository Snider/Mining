package database

import (
	"fmt"
	"time"
)

// MinerSession represents a mining session
type MinerSession struct {
	ID              int64      `json:"id"`
	MinerName       string     `json:"minerName"`
	MinerType       string     `json:"minerType"`
	StartedAt       time.Time  `json:"startedAt"`
	StoppedAt       *time.Time `json:"stoppedAt,omitempty"`
	TotalShares     int        `json:"totalShares"`
	RejectedShares  int        `json:"rejectedShares"`
	AverageHashrate int        `json:"averageHashrate"`
}

// StartSession records the start of a new mining session
func StartSession(minerName, minerType string) (int64, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return 0, nil
	}

	result, err := db.Exec(`
		INSERT INTO miner_sessions (miner_name, miner_type, started_at)
		VALUES (?, ?, ?)
	`, minerName, minerType, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to start session: %w", err)
	}

	return result.LastInsertId()
}

// EndSession marks a session as complete with final stats
func EndSession(sessionID int64, totalShares, rejectedShares, averageHashrate int) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	_, err := db.Exec(`
		UPDATE miner_sessions
		SET stopped_at = ?,
		    total_shares = ?,
		    rejected_shares = ?,
		    average_hashrate = ?
		WHERE id = ?
	`, time.Now(), totalShares, rejectedShares, averageHashrate, sessionID)

	return err
}

// EndSessionByName marks the most recent session for a miner as complete
func EndSessionByName(minerName string, totalShares, rejectedShares, averageHashrate int) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	_, err := db.Exec(`
		UPDATE miner_sessions
		SET stopped_at = ?,
		    total_shares = ?,
		    rejected_shares = ?,
		    average_hashrate = ?
		WHERE miner_name = ?
		  AND stopped_at IS NULL
	`, time.Now(), totalShares, rejectedShares, averageHashrate, minerName)

	return err
}

// GetSession retrieves a session by ID
func GetSession(sessionID int64) (*MinerSession, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	var session MinerSession
	var stoppedAt *time.Time

	err := db.QueryRow(`
		SELECT id, miner_name, miner_type, started_at, stopped_at,
		       total_shares, rejected_shares, average_hashrate
		FROM miner_sessions
		WHERE id = ?
	`, sessionID).Scan(
		&session.ID,
		&session.MinerName,
		&session.MinerType,
		&session.StartedAt,
		&stoppedAt,
		&session.TotalShares,
		&session.RejectedShares,
		&session.AverageHashrate,
	)
	if err != nil {
		return nil, err
	}

	session.StoppedAt = stoppedAt
	return &session, nil
}

// GetActiveSessions retrieves all currently active (non-stopped) sessions
func GetActiveSessions() ([]MinerSession, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT id, miner_name, miner_type, started_at, stopped_at,
		       total_shares, rejected_shares, average_hashrate
		FROM miner_sessions
		WHERE stopped_at IS NULL
		ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []MinerSession
	for rows.Next() {
		var session MinerSession
		var stoppedAt *time.Time
		if err := rows.Scan(
			&session.ID,
			&session.MinerName,
			&session.MinerType,
			&session.StartedAt,
			&stoppedAt,
			&session.TotalShares,
			&session.RejectedShares,
			&session.AverageHashrate,
		); err != nil {
			return nil, err
		}
		session.StoppedAt = stoppedAt
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetRecentSessions retrieves the most recent sessions for a miner
func GetRecentSessions(minerName string, limit int) ([]MinerSession, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT id, miner_name, miner_type, started_at, stopped_at,
		       total_shares, rejected_shares, average_hashrate
		FROM miner_sessions
		WHERE miner_name = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, minerName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []MinerSession
	for rows.Next() {
		var session MinerSession
		var stoppedAt *time.Time
		if err := rows.Scan(
			&session.ID,
			&session.MinerName,
			&session.MinerType,
			&session.StartedAt,
			&stoppedAt,
			&session.TotalShares,
			&session.RejectedShares,
			&session.AverageHashrate,
		); err != nil {
			return nil, err
		}
		session.StoppedAt = stoppedAt
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetSessionStats retrieves aggregated session statistics for a miner
type SessionStats struct {
	MinerName      string        `json:"minerName"`
	TotalSessions  int           `json:"totalSessions"`
	TotalUptime    time.Duration `json:"totalUptime"`
	TotalShares    int           `json:"totalShares"`
	TotalRejected  int           `json:"totalRejected"`
	AvgSessionTime time.Duration `json:"avgSessionTime"`
	AvgHashrate    int           `json:"avgHashrate"`
	LastSessionAt  time.Time     `json:"lastSessionAt"`
}

func GetSessionStats(minerName string) (*SessionStats, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil, nil
	}

	var stats SessionStats
	stats.MinerName = minerName

	// Get basic aggregates
	err := db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(total_shares), 0),
			COALESCE(SUM(rejected_shares), 0),
			COALESCE(AVG(average_hashrate), 0),
			MAX(started_at)
		FROM miner_sessions
		WHERE miner_name = ?
		  AND stopped_at IS NOT NULL
	`, minerName).Scan(
		&stats.TotalSessions,
		&stats.TotalShares,
		&stats.TotalRejected,
		&stats.AvgHashrate,
		&stats.LastSessionAt,
	)
	if err != nil {
		return nil, err
	}

	// Calculate total uptime
	rows, err := db.Query(`
		SELECT started_at, stopped_at
		FROM miner_sessions
		WHERE miner_name = ?
		  AND stopped_at IS NOT NULL
	`, minerName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalSeconds int64
	for rows.Next() {
		var started, stopped time.Time
		if err := rows.Scan(&started, &stopped); err != nil {
			return nil, err
		}
		totalSeconds += int64(stopped.Sub(started).Seconds())
	}

	stats.TotalUptime = time.Duration(totalSeconds) * time.Second
	if stats.TotalSessions > 0 {
		stats.AvgSessionTime = stats.TotalUptime / time.Duration(stats.TotalSessions)
	}

	return &stats, nil
}
