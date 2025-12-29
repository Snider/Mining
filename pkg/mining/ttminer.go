package mining

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TTMiner represents a TT-Miner (GPU miner), embedding the BaseMiner for common functionality.
type TTMiner struct {
	BaseMiner
	FullStats *TTMinerSummary `json:"full_stats,omitempty"`
}

// TTMinerSummary represents the stats response from TT-Miner API
type TTMinerSummary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Uptime  int    `json:"uptime"`
	Algo    string `json:"algo"`
	GPUs    []struct {
		Name      string  `json:"name"`
		ID        int     `json:"id"`
		Hashrate  float64 `json:"hashrate"`
		Temp      int     `json:"temp"`
		Fan       int     `json:"fan"`
		Power     int     `json:"power"`
		Accepted  int     `json:"accepted"`
		Rejected  int     `json:"rejected"`
		Intensity float64 `json:"intensity"`
	} `json:"gpus"`
	Results struct {
		SharesGood  int `json:"shares_good"`
		SharesTotal int `json:"shares_total"`
		AvgTime     int `json:"avg_time"`
	} `json:"results"`
	Connection struct {
		Pool string `json:"pool"`
		Ping int    `json:"ping"`
		Diff int    `json:"diff"`
	} `json:"connection"`
	Hashrate struct {
		Total   []float64 `json:"total"`
		Highest float64   `json:"highest"`
	} `json:"hashrate"`
}

// NewTTMiner creates a new TT-Miner instance with default settings.
func NewTTMiner() *TTMiner {
	return &TTMiner{
		BaseMiner: BaseMiner{
			Name:           "tt-miner",
			ExecutableName: "TT-Miner",
			Version:        "latest",
			URL:            "https://github.com/TrailingStop/TT-Miner-release",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
				ListenPort: 4068, // TT-Miner default port
			},
			HashrateHistory:       make([]HashratePoint, 0),
			LowResHashrateHistory: make([]HashratePoint, 0),
			LastLowResAggregation: time.Now(),
			LogBuffer:             NewLogBuffer(500), // Keep last 500 lines
		},
	}
}

// getTTMinerConfigPath returns the platform-specific path for the tt-miner config file.
func getTTMinerConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "lethean-desktop", "tt-miner.json"), nil
}

// GetLatestVersion fetches the latest version of TT-Miner from the GitHub API.
func (m *TTMiner) GetLatestVersion() (string, error) {
	resp, err := httpClient.Get("https://api.github.com/repos/TrailingStop/TT-Miner-release/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get latest release: unexpected status code %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

// Install determines the correct download URL for the latest version of TT-Miner
// and then calls the generic InstallFromURL method on the BaseMiner.
func (m *TTMiner) Install() error {
	version, err := m.GetLatestVersion()
	if err != nil {
		return err
	}
	m.Version = version

	var url string
	switch runtime.GOOS {
	case "windows":
		// Windows version - uses .zip
		url = fmt.Sprintf("https://github.com/TrailingStop/TT-Miner-release/releases/download/%s/TT-Miner-%s.zip", version, version)
	case "linux":
		// Linux version - uses .tar.gz
		url = fmt.Sprintf("https://github.com/TrailingStop/TT-Miner-release/releases/download/%s/TT-Miner-%s.tar.gz", version, version)
	default:
		return errors.New("TT-Miner is only available for Windows and Linux (requires CUDA)")
	}

	if err := m.InstallFromURL(url); err != nil {
		return err
	}

	// After installation, verify it.
	_, err = m.CheckInstallation()
	if err != nil {
		return fmt.Errorf("failed to verify installation after extraction: %w", err)
	}

	return nil
}

// Uninstall removes all files related to the TT-Miner, including its specific config file.
func (m *TTMiner) Uninstall() error {
	// Remove the specific tt-miner config file
	configPath, err := getTTMinerConfigPath()
	if err == nil {
		os.Remove(configPath) // Ignore error if it doesn't exist
	}

	// Call the base uninstall method to remove the installation directory
	return m.BaseMiner.Uninstall()
}

// CheckInstallation verifies if the TT-Miner is installed correctly.
func (m *TTMiner) CheckInstallation() (*InstallationDetails, error) {
	binaryPath, err := m.findMinerBinary()
	if err != nil {
		return &InstallationDetails{IsInstalled: false}, err
	}

	m.MinerBinary = binaryPath
	m.Path = filepath.Dir(binaryPath)

	// TT-Miner uses --version to check version
	cmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		m.Version = "Unknown (could not run executable)"
	} else {
		// Parse version from output
		output := strings.TrimSpace(out.String())
		fields := strings.Fields(output)
		if len(fields) >= 2 {
			m.Version = fields[1]
		} else if len(fields) >= 1 {
			m.Version = fields[0]
		} else {
			m.Version = "Unknown (could not parse version)"
		}
	}

	// Get the config path using the helper
	configPath, err := getTTMinerConfigPath()
	if err != nil {
		configPath = "Error: Could not determine config path"
	}

	return &InstallationDetails{
		IsInstalled: true,
		MinerBinary: m.MinerBinary,
		Path:        m.Path,
		Version:     m.Version,
		ConfigPath:  configPath,
	}, nil
}
