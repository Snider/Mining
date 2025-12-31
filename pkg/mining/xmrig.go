package mining

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
)

// XMRigMiner represents an XMRig miner, embedding the BaseMiner for common functionality.
type XMRigMiner struct {
	BaseMiner
	FullStats *XMRigSummary `json:"-"` // Excluded from JSON to prevent race during marshaling
}

var (
	httpClient   = &http.Client{Timeout: 30 * time.Second}
	httpClientMu sync.RWMutex
)

// getHTTPClient returns the HTTP client with proper synchronization
func getHTTPClient() *http.Client {
	httpClientMu.RLock()
	defer httpClientMu.RUnlock()
	return httpClient
}

// setHTTPClient sets the HTTP client (for testing)
func setHTTPClient(client *http.Client) {
	httpClientMu.Lock()
	defer httpClientMu.Unlock()
	httpClient = client
}

// MinerTypeXMRig is the type identifier for XMRig miners.
const MinerTypeXMRig = "xmrig"

// NewXMRigMiner creates a new XMRig miner instance with default settings.
func NewXMRigMiner() *XMRigMiner {
	return &XMRigMiner{
		BaseMiner: BaseMiner{
			Name:           "xmrig",
			MinerType:      MinerTypeXMRig,
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
	resp, err := getHTTPClient().Get("https://api.github.com/repos/xmrig/xmrig/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // Drain body to allow connection reuse
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
// Thread-safe: properly locks before modifying shared fields.
func (m *XMRigMiner) CheckInstallation() (*InstallationDetails, error) {
	binaryPath, err := m.findMinerBinary()
	if err != nil {
		return &InstallationDetails{IsInstalled: false}, err
	}

	// Run version command before acquiring lock (I/O operation)
	cmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	var version string
	if err := cmd.Run(); err != nil {
		version = "Unknown (could not run executable)"
	} else {
		fields := strings.Fields(out.String())
		if len(fields) >= 2 {
			version = fields[1]
		} else {
			version = "Unknown (could not parse version)"
		}
	}

	// Get the config path using the helper (use instance name if set)
	m.mu.RLock()
	instanceName := m.Name
	m.mu.RUnlock()

	configPath, err := getXMRigConfigPath(instanceName)
	if err != nil {
		// Log the error but don't fail CheckInstallation if config path can't be determined
		configPath = "Error: Could not determine config path"
	}

	// Update shared fields under lock
	m.mu.Lock()
	m.MinerBinary = binaryPath
	m.Path = filepath.Dir(binaryPath)
	m.Version = version
	m.mu.Unlock()

	return &InstallationDetails{
		IsInstalled: true,
		MinerBinary: binaryPath,
		Path:        filepath.Dir(binaryPath),
		Version:     version,
		ConfigPath:  configPath,
	}, nil
}
