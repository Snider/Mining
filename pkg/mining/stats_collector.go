package mining

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StatsCollector defines the interface for collecting miner statistics.
// This allows different miner types to implement their own stats collection logic
// while sharing common HTTP fetching infrastructure.
type StatsCollector interface {
	// CollectStats fetches and returns performance metrics from the miner.
	CollectStats(ctx context.Context) (*PerformanceMetrics, error)
}

// HTTPStatsConfig holds configuration for HTTP-based stats collection.
type HTTPStatsConfig struct {
	Host     string
	Port     int
	Endpoint string // e.g., "/2/summary" for XMRig, "/summary" for TT-Miner
}

// FetchJSONStats performs an HTTP GET request and decodes the JSON response.
// This is a common helper for HTTP-based miner stats collection.
// The caller must provide the target struct to decode into.
func FetchJSONStats[T any](ctx context.Context, config HTTPStatsConfig, target *T) error {
	if config.Port == 0 {
		return fmt.Errorf("API port is zero")
	}

	url := fmt.Sprintf("http://%s:%d%s", config.Host, config.Port, config.Endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // Drain body to allow connection reuse
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// MinerTypeRegistry provides a central registry of known miner types.
// This can be used for validation and discovery of available miners.
var MinerTypeRegistry = map[string]string{
	MinerTypeXMRig:     "XMRig - CPU/GPU miner for RandomX, KawPow, CryptoNight",
	MinerTypeTTMiner:   "TT-Miner - NVIDIA GPU miner for Ethash, KawPow, ProgPow",
	MinerTypeSimulated: "Simulated - Mock miner for testing and development",
}

// IsKnownMinerType returns true if the given type is a registered miner type.
func IsKnownMinerType(minerType string) bool {
	_, exists := MinerTypeRegistry[minerType]
	return exists
}
