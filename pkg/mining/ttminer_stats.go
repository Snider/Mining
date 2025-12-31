package mining

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// GetStats retrieves performance metrics from the TT-Miner API.
func (m *TTMiner) GetStats(ctx context.Context) (*PerformanceMetrics, error) {
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

	req, err := http.NewRequestWithContext(reqCtx, "GET", fmt.Sprintf("http://%s:%d/summary", host, port), nil)
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

	var summary TTMinerSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	// Store the full summary in the miner struct (requires lock)
	m.mu.Lock()
	m.FullStats = &summary
	m.mu.Unlock()

	// Calculate total hashrate from all GPUs
	var totalHashrate float64
	if len(summary.Hashrate.Total) > 0 {
		totalHashrate = summary.Hashrate.Total[0]
	} else {
		// Sum individual GPU hashrates
		for _, gpu := range summary.GPUs {
			totalHashrate += gpu.Hashrate
		}
	}

	// For TT-Miner, we use the connection difficulty as both current and avg
	// since TT-Miner doesn't expose per-share difficulty data
	diffCurrent := summary.Connection.Diff

	return &PerformanceMetrics{
		Hashrate:      int(totalHashrate),
		Shares:        summary.Results.SharesGood,
		Rejected:      summary.Results.SharesTotal - summary.Results.SharesGood,
		Uptime:        summary.Uptime,
		Algorithm:     summary.Algo,
		AvgDifficulty: diffCurrent, // Use pool diff as approximation
		DiffCurrent:   diffCurrent,
	}, nil
}
