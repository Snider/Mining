package mining

import (
	"context"
	"errors"
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
	config := HTTPStatsConfig{
		Host:     m.API.ListenHost,
		Port:     m.API.ListenPort,
		Endpoint: "/summary",
	}
	m.mu.RUnlock()

	// Create request with context and timeout
	reqCtx, cancel := context.WithTimeout(ctx, statsTimeout)
	defer cancel()

	// Use the common HTTP stats fetcher
	var summary TTMinerSummary
	if err := FetchJSONStats(reqCtx, config, &summary); err != nil {
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
