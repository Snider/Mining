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

	"github.com/adrg/xdg"
)

// XMRigMiner represents an XMRig miner, embedding the BaseMiner for common functionality.
type XMRigMiner struct {
	BaseMiner
	FullStats *XMRigSummary `json:"full_stats,omitempty"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// NewXMRigMiner creates a new XMRig miner instance with default settings.
func NewXMRigMiner() *XMRigMiner {
	return &XMRigMiner{
		BaseMiner: BaseMiner{
			Name:           "xmrig",
			ExecutableName: "xmrig",
			Version:        "latest",
			URL:            "https://github.com/xmrig/xmrig/releases",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
			},
			HashrateHistory:       make([]HashratePoint, 0),
			LowResHashrateHistory: make([]HashratePoint, 0),
			LastLowResAggregation: time.Now(),
			LogBuffer:             NewLogBuffer(500), // Keep last 500 lines
		},
	}
}

// getXMRigConfigPath returns the platform-specific path for the xmrig.json file.
// If instanceName is provided, it creates an instance-specific config file.
// This is a variable so it can be overridden in tests.
var getXMRigConfigPath = func(instanceName string) (string, error) {
	configFileName := "xmrig.json"
	if instanceName != "" && instanceName != "xmrig" {
		// Use instance-specific config file (e.g., xmrig-78.json)
		configFileName = instanceName + ".json"
	}

	path, err := xdg.ConfigFile("lethean-desktop/" + configFileName)
	if err != nil {
		// Fallback for non-XDG environments or when XDG variables are not set
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", homeErr
		}
		return filepath.Join(homeDir, ".config", "lethean-desktop", configFileName), nil
	}
	return path, nil
}

// GetLatestVersion fetches the latest version of XMRig from the GitHub API.
func (m *XMRigMiner) GetLatestVersion() (string, error) {
	resp, err := httpClient.Get("https://api.github.com/repos/xmrig/xmrig/releases/latest")
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

// Install determines the correct download URL for the latest version of XMRig
// and then calls the generic InstallFromURL method on the BaseMiner.
func (m *XMRigMiner) Install() error {
	version, err := m.GetLatestVersion()
	if err != nil {
		return err
	}
	m.Version = version

	var url string
	switch runtime.GOOS {
	case "windows":
		url = fmt.Sprintf("https://github.com/xmrig/xmrig/releases/download/%s/xmrig-%s-windows-x64.zip", version, strings.TrimPrefix(version, "v"))
	case "linux":
		url = fmt.Sprintf("https://github.com/xmrig/xmrig/releases/download/%s/xmrig-%s-linux-static-x64.tar.gz", version, strings.TrimPrefix(version, "v"))
	case "darwin":
		url = fmt.Sprintf("https://github.com/xmrig/xmrig/releases/download/%s/xmrig-%s-macos-x64.tar.gz", version, strings.TrimPrefix(version, "v"))
	default:
		return errors.New("unsupported operating system")
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

// Uninstall removes all files related to the XMRig miner, including its specific config file.
func (m *XMRigMiner) Uninstall() error {
	// Remove the instance-specific config file
	configPath, err := getXMRigConfigPath(m.Name)
	if err == nil {
		os.Remove(configPath) // Ignore error if it doesn't exist
	}

	// Call the base uninstall method to remove the installation directory
	return m.BaseMiner.Uninstall()
}

// CheckInstallation verifies if the XMRig miner is installed correctly.
func (m *XMRigMiner) CheckInstallation() (*InstallationDetails, error) {
	binaryPath, err := m.findMinerBinary()
	if err != nil {
		return &InstallationDetails{IsInstalled: false}, err
	}

	m.MinerBinary = binaryPath
	m.Path = filepath.Dir(binaryPath)

	cmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		m.Version = "Unknown (could not run executable)"
	} else {
		fields := strings.Fields(out.String())
		if len(fields) >= 2 {
			m.Version = fields[1]
		} else {
			m.Version = "Unknown (could not parse version)"
		}
	}

	// Get the config path using the helper (use instance name if set)
	configPath, err := getXMRigConfigPath(m.Name)
	if err != nil {
		// Log the error but don't fail CheckInstallation if config path can't be determined
		configPath = "Error: Could not determine config path"
	}

	return &InstallationDetails{
		IsInstalled: true,
		MinerBinary: m.MinerBinary,
		Path:        m.Path,
		Version:     m.Version,
		ConfigPath:  configPath, // Include the config path
	}, nil
}
