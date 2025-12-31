package mining

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetStats retrieves the performance statistics from the running XMRig miner.
func (m *XMRigMiner) GetStats() (*PerformanceMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running {
		return nil, errors.New("miner is not running")
	}
	if m.API == nil || m.API.ListenPort == 0 {
		return nil, errors.New("miner API not configured or port is zero")
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/2/summary", m.API.ListenHost, m.API.ListenPort))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stats: unexpected status code %d", resp.StatusCode)
	}

	var summary XMRigSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	// Store the full summary in the miner struct
	m.FullStats = &summary

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
