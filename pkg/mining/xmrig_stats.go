package mining

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetStats retrieves the performance statistics from the running XMRig miner.
func (m *XMRigMiner) GetStats() (*PerformanceMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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

	var hashrate int
	if len(summary.Hashrate.Total) > 0 {
		hashrate = int(summary.Hashrate.Total[0])
	}

	return &PerformanceMetrics{
		Hashrate:  hashrate,
		Shares:    int(summary.Results.SharesGood),
		Rejected:  int(summary.Results.SharesTotal - summary.Results.SharesGood),
		Uptime:    int(summary.Uptime),
		Algorithm: summary.Algorithm,
	}, nil
}
