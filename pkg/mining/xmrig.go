package mining

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// XMRigMiner represents an XMRig miner, embedding the BaseMiner for common functionality.
type XMRigMiner struct {
	BaseMiner
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
		},
	}
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
		url = fmt.Sprintf("https://github.com/xmrig/xmrig/releases/download/%s/xmrig-%s-msvc-win64.zip", version, strings.TrimPrefix(version, "v"))
	case "linux":
		url = fmt.Sprintf("https://github.com/xmrig/xmrig/releases/download/%s/xmrig-%s-linux-x64.tar.gz", version, strings.TrimPrefix(version, "v"))
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
