package mining

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// statsTimeout is the timeout for stats HTTP requests (shorter than general timeout)
const statsTimeout = 5 * time.Second

// GetStats retrieves the performance statistics from the running XMRig miner.
func (m *XMRigMiner) GetStats(ctx context.Context) (*PerformanceMetrics, error) {
	// Read state under RLock, then release before HTTP call
	m.mu.RLock()
	if !m.Running {
		m.mu.RUnlock()
		return nil, errors.New("miner is not running")
	}
	if m.API == nil || m.API.ListenPort == 0 {
		m.mu.RUnlock()
		return nil, errors.New("miner API not configured or port is zero")
	}
	host := m.API.ListenHost
	port := m.API.ListenPort
	m.mu.RUnlock()

	// Create request with context and timeout
	reqCtx, cancel := context.WithTimeout(ctx, statsTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", fmt.Sprintf("http://%s:%d/2/summary", host, port), nil)
	if err != nil {
		return nil, err
	}

	// HTTP call outside the lock to avoid blocking other operations
	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // Drain body to allow connection reuse
		return nil, fmt.Errorf("failed to get stats: unexpected status code %d", resp.StatusCode)
	}

	var summary XMRigSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	// Store the full summary in the miner struct (requires lock)
	m.mu.Lock()
	m.FullStats = &summary
	m.mu.Unlock()

	var hashrate int
	if len(summary.Hashrate.Total) > 0 {
		hashrate = int(summary.Hashrate.Total[0])
	}

	// Calculate average difficulty per accepted share
	var avgDifficulty int
	if summary.Results.SharesGood > 0 {
		avgDifficulty = summary.Results.HashesTotal / summary.Results.SharesGood
	}

	return &PerformanceMetrics{
		Hashrate:      hashrate,
		Shares:        summary.Results.SharesGood,
		Rejected:      summary.Results.SharesTotal - summary.Results.SharesGood,
		Uptime:        summary.Uptime,
		Algorithm:     summary.Algo,
		AvgDifficulty: avgDifficulty,
		DiffCurrent:   summary.Results.DiffCurrent,
	}, nil
}
