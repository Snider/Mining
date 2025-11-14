package mining

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// NewXMRigMiner creates a new XMRig miner instance with default settings.
// This is the entry point for creating a new XMRig miner that can be managed
// by the Manager. The returned miner is ready to be installed and started.
//
// Example:
//
//	// Create a new XMRig miner
//	xmrigMiner := mining.NewXMRigMiner()
//
//	// Now you can use the miner to perform actions like
//	// installing, starting, and stopping.
func NewXMRigMiner() *XMRigMiner {
	return &XMRigMiner{
		Name:    "xmrig",
		Version: "latest",
		URL:     "https://github.com/xmrig/xmrig/releases",
		API: &API{
			Enabled:    true,
			ListenHost: "127.0.0.1",
			ListenPort: 9000,
		},
		HashrateHistory:       make([]HashratePoint, 0),
		LowResHashrateHistory: make([]HashratePoint, 0),
		LastLowResAggregation: time.Now(),
	}
}

// GetName returns the name of the miner.
func (m *XMRigMiner) GetName() string {
	return m.Name
}

// GetPath returns the base installation directory for the XMRig miner.
// This is the directory where different versions of the miner are stored.
func (m *XMRigMiner) GetPath() string {
	dataPath, err := xdg.DataFile("lethean-desktop/miners/xmrig")
	if err != nil {
		return ""
	}
	return dataPath
}

// GetBinaryPath returns the full path to the miner's executable file.
// This path is set after a successful installation or check of the installation status.
func (m *XMRigMiner) GetBinaryPath() string {
	return m.MinerBinary
}

// GetLatestVersion fetches the latest version of XMRig from the GitHub API.
// It returns the version string (e.g., "v6.18.0") or an error if the
// version could not be retrieved.
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

// Install downloads and installs the latest version of the XMRig miner.
// It determines the correct release for the current operating system,
// downloads it, and extracts it to the appropriate installation directory.
func (m *XMRigMiner) Install() error {
	version, err := m.GetLatestVersion()
	if err != nil {
		return err
	}
	m.Version = version

	// Construct the download URL
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

	// Create a temporary file to download the release to
	tmpfile, err := os.CreateTemp("", "xmrig-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	// Download the release
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download release: unexpected status code %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpfile, resp.Body); err != nil {
		return err
	}

	// The base installation path (e.g., .../miners/xmrig)
	baseInstallPath := m.GetPath()

	// Create the base installation directory if it doesn't exist
	if err := os.MkdirAll(baseInstallPath, 0755); err != nil {
		return err
	}

	// Extract the release
	if strings.HasSuffix(url, ".zip") {
		err = m.unzip(tmpfile.Name(), baseInstallPath)
	} else {
		err = m.untar(tmpfile.Name(), baseInstallPath)
	}
	if err != nil {
		return fmt.Errorf("failed to extract miner: %w", err)
	}

	// After extraction, call CheckInstallation to populate m.Path and m.MinerBinary correctly
	_, err = m.CheckInstallation()
	if err != nil {
		return fmt.Errorf("failed to verify installation after extraction: %w", err)
	}

	return nil
}

// Uninstall removes all files related to the XMRig miner.
// This is a destructive operation that will remove the entire installation
// directory of the miner.
func (m *XMRigMiner) Uninstall() error {
	return os.RemoveAll(m.GetPath())
}

// findMinerBinary searches for the miner's executable file.
// It first checks the standard installation path, and if not found, falls
// back to searching the system's PATH. This allows for both managed
// installations and pre-existing installations to be used.
func (m *XMRigMiner) findMinerBinary() (string, error) {
	executableName := "xmrig"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}

	// 1. Check the standard installation directory first
	baseInstallPath := m.GetPath()
	if _, err := os.Stat(baseInstallPath); err == nil {
		files, err := os.ReadDir(baseInstallPath)
		if err == nil {
			for _, f := range files {
				if f.IsDir() && strings.HasPrefix(f.Name(), "xmrig-") {
					versionedPath := filepath.Join(baseInstallPath, f.Name())
					fullPath := filepath.Join(versionedPath, executableName)
					if _, err := os.Stat(fullPath); err == nil {
						log.Printf("Found miner binary at standard path: %s", fullPath)
						return fullPath, nil
					}
				}
			}
		}
	}

	// 2. Fallback to searching the system PATH
	path, err := exec.LookPath(executableName)
	if err == nil {
		log.Printf("Found miner binary in system PATH: %s", path)
		return path, nil
	}

	return "", errors.New("miner executable not found in standard directory or system PATH")
}

// CheckInstallation verifies if the XMRig miner is installed correctly.
// It returns details about the installation, such as whether it is installed,
// its version, and the path to the executable. This method also updates the
// miner's internal state with the installation details.
func (m *XMRigMiner) CheckInstallation() (*InstallationDetails, error) {
	details := &InstallationDetails{}

	binaryPath, err := m.findMinerBinary()
	if err != nil {
		details.IsInstalled = false
		return details, nil // Return not-installed, but no error
	}

	details.IsInstalled = true
	details.MinerBinary = binaryPath
	details.Path = filepath.Dir(binaryPath) // The directory containing the executable

	// Try to get the version from the executable
	cmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		details.Version = "Unknown (could not run executable)"
	} else {
		// XMRig version output is typically "XMRig 6.18.0"
		fields := strings.Fields(out.String())
		if len(fields) >= 2 {
			details.Version = fields[1]
		} else {
			details.Version = "Unknown (could not parse version)"
		}
	}

	// Update the XMRigMiner struct's fields
	m.Path = details.Path
	m.MinerBinary = details.MinerBinary
	m.Version = details.Version // Keep the miner's version in sync

	return details, nil
}

// Start launches the XMRig miner with the specified configuration.
// It creates a configuration file, constructs the necessary command-line
// arguments, and starts the miner process in the background.
//
// Example:
//
//	// Create a new XMRig miner and a configuration
//	xmrigMiner := mining.NewXMRigMiner()
//	config := &mining.Config{
//		Pool:    "your-pool-address",
//		Wallet:  "your-wallet-address",
//		Threads: 4,
//	}
//
//	// Start the miner
//	err := xmrigMiner.Start(config)
//	if err != nil {
//		log.Fatalf("Failed to start miner: %v", err)
//	}
//
//	// Stop the miner when you are done
//	defer xmrigMiner.Stop()
func (m *XMRigMiner) Start(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Running {
		return errors.New("miner is already running")
	}

	// Ensure MinerBinary is set before starting
	if m.MinerBinary == "" {
		// Re-check installation to populate MinerBinary if it's not set
		_, err := m.CheckInstallation()
		if err != nil {
			return fmt.Errorf("failed to verify miner installation before starting: %w", err)
		}
		if m.MinerBinary == "" {
			return errors.New("miner executable path not found")
		}
	}

	if _, err := os.Stat(m.MinerBinary); os.IsNotExist(err) {
		return fmt.Errorf("xmrig executable not found at %s", m.MinerBinary)
	}

	// Create the config file (this handles pool, wallet, threads, hugepages, tls, and API settings)
	if err := m.createConfig(config); err != nil {
		return err
	}

	// Arguments for XMRig
	args := []string{
		"-c", m.ConfigPath, // Always use the generated config file
	}

	// Dynamically add command-line arguments based on the Config struct
	// Network options
	// Pool and Wallet are primarily handled by the config file, but CLI can override
	if config.Pool != "" {
		args = append(args, "-o", config.Pool)
	}
	if config.Wallet != "" {
		args = append(args, "-u", config.Wallet)
	}
	if config.Algo != "" {
		args = append(args, "-a", config.Algo)
	}
	if config.Coin != "" {
		args = append(args, "--coin", config.Coin)
	}
	if config.Password != "" {
		args = append(args, "-p", config.Password)
	}
	if config.UserPass != "" {
		args = append(args, "-O", config.UserPass)
	}
	if config.Proxy != "" {
		args = append(args, "-x", config.Proxy)
	}
	if config.Keepalive {
		args = append(args, "-k")
	}
	if config.Nicehash {
		args = append(args, "--nicehash")
	}
	if config.RigID != "" {
		args = append(args, "--rig-id", config.RigID)
	}
	// TLS is handled by config file, but --tls-fingerprint is a CLI option
	//if config.TLS { // If TLS is true in config, ensure --tls is passed if not already in config file
	args = append(args, "--tls")
	//}
	if config.TLSSingerprint != "" {
		args = append(args, "--tls-fingerprint", config.TLSSingerprint)
	}
	if config.Retries != 0 {
		args = append(args, "-r", fmt.Sprintf("%d", config.Retries))
	}
	if config.RetryPause != 0 {
		args = append(args, "-R", fmt.Sprintf("%d", config.RetryPause))
	}
	if config.UserAgent != "" {
		args = append(args, "--user-agent", config.UserAgent)
	}
	if config.DonateLevel != 0 {
		args = append(args, "--donate-level", fmt.Sprintf("%d", config.DonateLevel))
	}
	if config.DonateOverProxy {
		args = append(args, "--donate-over-proxy")
	}

	// CPU backend options
	if config.NoCPU {
		args = append(args, "--no-cpu")
	}
	// Threads is handled by config file, but can be overridden by CLI
	if config.Threads != 0 { // This will override the config file setting if provided
		args = append(args, "-t", fmt.Sprintf("%d", config.Threads))
	}
	if config.CPUAffinity != "" {
		args = append(args, "--cpu-affinity", config.CPUAffinity)
	}
	if config.AV != 0 {
		args = append(args, "-v", fmt.Sprintf("%d", config.AV))
	}
	if config.CPUPriority != 0 {
		args = append(args, "--cpu-priority", fmt.Sprintf("%d", config.CPUPriority))
	}
	if config.CPUMaxThreadsHint != 0 {
		args = append(args, "--cpu-max-threads-hint", fmt.Sprintf("%d", config.CPUMaxThreadsHint))
	}
	if config.CPUMemoryPool != 0 {
		args = append(args, "--cpu-memory-pool", fmt.Sprintf("%d", config.CPUMemoryPool))
	}
	if config.CPUNoYield {
		args = append(args, "--cpu-no-yield")
	}
	if !config.HugePages { // If HugePages is explicitly false in config, add --no-huge-pages
		args = append(args, "--no-huge-pages")
	}
	if config.HugepageSize != 0 {
		args = append(args, "--hugepage-size", fmt.Sprintf("%d", config.HugepageSize))
	}
	if config.HugePagesJIT {
		args = append(args, "--huge-pages-jit")
	}
	if config.ASM != "" {
		args = append(args, "--asm", config.ASM)
	}
	if config.Argon2Impl != "" {
		args = append(args, "--argon2-impl", config.Argon2Impl)
	}
	if config.RandomXInit != 0 {
		args = append(args, "--randomx-init", fmt.Sprintf("%d", config.RandomXInit))
	}
	if config.RandomXNoNUMA {
		args = append(args, "--randomx-no-numa")
	}
	if config.RandomXMode != "" {
		args = append(args, "--randomx-mode", config.RandomXMode)
	}
	if config.RandomX1GBPages {
		args = append(args, "--randomx-1gb-pages")
	}
	if config.RandomXWrmsr != "" {
		args = append(args, "--randomx-wrmsr", config.RandomXWrmsr)
	}
	if config.RandomXNoRdmsr {
		args = append(args, "--randomx-no-rdmsr")
	}
	if config.RandomXCacheQoS {
		args = append(args, "--randomx-cache-qos")
	}

	// API options (CLI options override config file and m.API defaults)
	if m.API.Enabled {
		if config.APIWorkerID != "" {
			args = append(args, "--api-worker-id", config.APIWorkerID)
		}
		if config.APIID != "" {
			args = append(args, "--api-id", config.APIID)
		}
		if config.HTTPHost != "" {
			args = append(args, "--http-host", config.HTTPHost)
		} else {
			args = append(args, "--http-host", m.API.ListenHost)
		}
		if config.HTTPPort != 0 {
			args = append(args, "--http-port", fmt.Sprintf("%d", config.HTTPPort))
		} else {
			args = append(args, "--http-port", fmt.Sprintf("%d", m.API.ListenPort))
		}
		if config.HTTPAccessToken != "" {
			args = append(args, "--http-access-token", config.HTTPAccessToken)
		}
		if config.HTTPNoRestricted {
			args = append(args, "--http-no-restricted")
		}
	}

	// Logging options
	if config.Syslog {
		args = append(args, "-S")
	}
	if config.LogFile != "" {
		args = append(args, "-l", config.LogFile)
	}
	if config.PrintTime != 0 {
		args = append(args, "--print-time", fmt.Sprintf("%d", config.PrintTime))
	}
	if config.HealthPrintTime != 0 {
		args = append(args, "--health-print-time", fmt.Sprintf("%d", config.HealthPrintTime))
	}
	if config.NoColor {
		args = append(args, "--no-color")
	}
	if config.Verbose {
		args = append(args, "--verbose")
	}

	// Misc options
	if config.Background {
		args = append(args, "-B")
	}
	if config.Title != "" {
		args = append(args, "--title", config.Title)
	}
	if config.NoTitle {
		args = append(args, "--no-title")
	}
	if config.PauseOnBattery {
		args = append(args, "--pause-on-battery")
	}
	if config.PauseOnActive != 0 {
		args = append(args, "--pause-on-active", fmt.Sprintf("%d", config.PauseOnActive))
	}
	if config.Stress {
		args = append(args, "--stress")
	}
	if config.Bench != "" {
		args = append(args, "--bench", config.Bench)
	}
	if config.Submit {
		args = append(args, "--submit")
	}
	if config.Verify != "" {
		args = append(args, "--verify", config.Verify)
	}
	if config.Seed != "" {
		args = append(args, "--seed", config.Seed)
	}
	if config.Hash != "" {
		args = append(args, "--hash", config.Hash)
	}
	if config.NoDMI {
		args = append(args, "--no-dmi")
	}

	fmt.Fprintf(os.Stderr, "Executing XMRig command: %s %s\n", m.MinerBinary, strings.Join(args, " "))

	m.cmd = exec.Command(m.MinerBinary, args...)

	if config.LogOutput {
		m.cmd.Stdout = os.Stdout
		m.cmd.Stderr = os.Stderr
	}

	if err := m.cmd.Start(); err != nil {
		return err
	}

	m.Running = true

	go func() {
		m.cmd.Wait()
		m.mu.Lock()
		m.Running = false
		m.cmd = nil
		m.mu.Unlock()
	}()

	return nil
}

// Stop terminates the XMRig miner process.
// It sends a kill signal to the running miner process.
func (m *XMRigMiner) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running || m.cmd == nil {
		return errors.New("miner is not running")
	}

	return m.cmd.Process.Kill()
}

// GetStats retrieves the performance statistics from the running XMRig miner.
// It queries the miner's API and returns a PerformanceMetrics struct
// containing the hashrate, share counts, and uptime.
func (m *XMRigMiner) GetStats() (*PerformanceMetrics, error) {
	m.mu.Lock()
	running := m.Running
	m.mu.Unlock()

	if !running {
		return nil, errors.New("miner is not running")
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

// GetHashrateHistory returns the combined high-resolution and low-resolution hashrate history.
// This provides a complete view of the miner's performance over time.
func (m *XMRigMiner) GetHashrateHistory() []HashratePoint {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Combine low-res and high-res history
	combinedHistory := make([]HashratePoint, 0, len(m.LowResHashrateHistory)+len(m.HashrateHistory))
	combinedHistory = append(combinedHistory, m.LowResHashrateHistory...)
	combinedHistory = append(combinedHistory, m.HashrateHistory...)

	return combinedHistory
}

// AddHashratePoint adds a new hashrate measurement to the high-resolution history.
// This method is called periodically by the Manager to record the miner's performance.
func (m *XMRigMiner) AddHashratePoint(point HashratePoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HashrateHistory = append(m.HashrateHistory, point)
}

// GetHighResHistoryLength returns the number of data points in the high-resolution hashrate history.
func (m *XMRigMiner) GetHighResHistoryLength() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.HashrateHistory)
}

// GetLowResHistoryLength returns the number of data points in the low-resolution hashrate history.
func (m *XMRigMiner) GetLowResHistoryLength() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.LowResHashrateHistory)
}

// ReduceHashrateHistory aggregates older high-resolution hashrate data into
// lower-resolution data, and trims the history to a manageable size.
// This method is called periodically by the Manager to maintain the hashrate
// history.
func (m *XMRigMiner) ReduceHashrateHistory(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only aggregate if enough time has passed since the last aggregation
	if !m.LastLowResAggregation.IsZero() && now.Sub(m.LastLowResAggregation) < LowResolutionInterval {
		return
	}

	var pointsToAggregate []HashratePoint
	var newHighResHistory []HashratePoint

	// The cutoff is exclusive: points *at or before* this time are candidates for aggregation.
	// We want to aggregate points that are *strictly older* than HighResolutionDuration ago.
	cutoff := now.Add(-HighResolutionDuration)

	for _, p := range m.HashrateHistory {
		if p.Timestamp.Before(cutoff) {
			pointsToAggregate = append(pointsToAggregate, p)
		} else {
			newHighResHistory = append(newHighResHistory, p)
		}
	}
	m.HashrateHistory = newHighResHistory

	if len(pointsToAggregate) == 0 {
		m.LastLowResAggregation = now
		return
	}

	// Group points by minute and calculate average hashrate
	minuteGroups := make(map[time.Time][]int)
	for _, p := range pointsToAggregate {
		minute := p.Timestamp.Truncate(LowResolutionInterval)
		minuteGroups[minute] = append(minuteGroups[minute], p.Hashrate)
	}

	var newLowResPoints []HashratePoint
	for minute, hashrates := range minuteGroups {
		if len(hashrates) > 0 {
			totalHashrate := 0
			for _, hr := range hashrates {
				totalHashrate += hr
			}
			avgHashrate := totalHashrate / len(hashrates)
			newLowResPoints = append(newLowResPoints, HashratePoint{
				Timestamp: minute,
				Hashrate:  avgHashrate,
			})
		}
	}

	sort.Slice(newLowResPoints, func(i, j int) bool {
		return newLowResPoints[i].Timestamp.Before(newLowResPoints[j].Timestamp)
	})

	m.LowResHashrateHistory = append(m.LowResHashrateHistory, newLowResPoints...)

	// Trim low-resolution history to LowResHistoryRetention
	lowResCutoff := now.Add(-LowResHistoryRetention)
	firstValidLowResIndex := 0
	for i, p := range m.LowResHashrateHistory {
		if p.Timestamp.After(lowResCutoff) || p.Timestamp.Equal(lowResCutoff) {
			firstValidLowResIndex = i
			break
		}
		if i == len(m.LowResHashrateHistory)-1 {
			firstValidLowResIndex = len(m.LowResHashrateHistory)
		}
	}
	m.LowResHashrateHistory = m.LowResHashrateHistory[firstValidLowResIndex:]

	m.LastLowResAggregation = now
}

// createConfig creates a JSON configuration file for the XMRig miner.
// This allows for a consistent and reproducible way to configure the miner,
// based on the provided Config struct.
func (m *XMRigMiner) createConfig(config *Config) error {
	configPath, err := xdg.ConfigFile("lethean-desktop/xmrig.json")
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(homeDir, ".config", "lethean-desktop", "xmrig.json")
	}
	m.ConfigPath = configPath

	if err := os.MkdirAll(filepath.Dir(m.ConfigPath), 0755); err != nil {
		return err
	}

	c := map[string]interface{}{
		"api": map[string]interface{}{
			"enabled":      m.API.Enabled,
			"listen":       fmt.Sprintf("%s:%d", m.API.ListenHost, m.API.ListenPort),
			"access-token": nil,
			"restricted":   true,
		},
		"pools": []map[string]interface{}{
			{
				"url":       config.Pool,
				"user":      config.Wallet,
				"pass":      "x",
				"keepalive": true,
				"tls":       config.TLS,
			},
		},
		"cpu": map[string]interface{}{
			"enabled":    true,
			"threads":    config.Threads,
			"huge-pages": config.HugePages,
		},
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.ConfigPath, data, 0644)
}

// unzip extracts a zip archive to a destination directory.
// This is a helper function used during the installation of the miner.
func (m *XMRigMiner) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// untar extracts a tar.gz archive to a destination directory.
// This is a helper function used during the installation of the miner.
func (m *XMRigMiner) untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		cleanedName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanedName, "..") || strings.HasPrefix(cleanedName, "/") || cleanedName == "." {
			continue
		}

		target := filepath.Join(dest, cleanedName)
		rel, err := filepath.Rel(dest, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
