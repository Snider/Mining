package mining

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetStats retrieves performance metrics from the TT-Miner API.
func (m *TTMiner) GetStats() (*PerformanceMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running {
		return nil, errors.New("miner is not running")
	}
	if m.API == nil || m.API.ListenPort == 0 {
		return nil, errors.New("miner API not configured or port is zero")
	}

	// TT-Miner API endpoint - try the summary endpoint
	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/summary", m.API.ListenHost, m.API.ListenPort))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stats: unexpected status code %d", resp.StatusCode)
	}

	var summary TTMinerSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	// Store the full summary in the miner struct
	m.FullStats = &summary

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
