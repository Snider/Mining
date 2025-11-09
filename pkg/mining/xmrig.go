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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// NewXMRigMiner creates a new XMRig miner
func NewXMRigMiner() *XMRigMiner {
	return &XMRigMiner{
		Name:    "xmrig", // Changed to lowercase for consistency
		Version: "latest",
		URL:     "https://github.com/xmrig/xmrig/releases",
		API: &API{
			Enabled:    true,
			ListenHost: "127.0.0.1",
			ListenPort: 9000,
		},
	}
}

// GetName returns the name of the miner
func (m *XMRigMiner) GetName() string {
	return m.Name
}

// GetPath returns the path of the miner
// This now returns the base installation directory for xmrig, not the versioned one.
func (m *XMRigMiner) GetPath() string {
	dataPath, err := xdg.DataFile("lethean-desktop/miners/xmrig")
	if err != nil {
		// Fallback for safety, though it should ideally not fail if Install works.
		return ""
	}
	return dataPath
}

// GetLatestVersion returns the latest version of XMRig
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

// Download and install the latest version of XMRig
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

// Uninstall removes the miner files
func (m *XMRigMiner) Uninstall() error {
	// Uninstall should remove the base path, which contains the versioned folder
	return os.RemoveAll(m.GetPath())
}

// CheckInstallation checks if the miner is installed and returns its details
func (m *XMRigMiner) CheckInstallation() (*InstallationDetails, error) {
	baseInstallPath := m.GetPath()
	details := &InstallationDetails{
		Path: baseInstallPath, // Initialize with base path, will be updated to versioned path
	}

	if _, err := os.Stat(baseInstallPath); os.IsNotExist(err) {
		details.IsInstalled = false
		return details, nil
	}

	// The directory exists, now check for the executable by finding the versioned sub-folder
	files, err := os.ReadDir(baseInstallPath)
	if err != nil {
		return nil, fmt.Errorf("could not read installation directory: %w", err)
	}

	var versionedDir string
	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), "xmrig-") {
			versionedDir = f.Name()
			break
		}
	}

	if versionedDir == "" {
		details.IsInstalled = false // Directory exists but is empty or malformed
		return details, nil
	}

	// Update the Path to be the versioned directory
	details.Path = filepath.Join(baseInstallPath, versionedDir)

	var executableName string
	if runtime.GOOS == "windows" {
		executableName = "xmrig.exe"
	} else {
		executableName = "xmrig"
	}

	executablePath := filepath.Join(details.Path, executableName)
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		details.IsInstalled = false // Versioned folder exists, but no executable
		return details, nil
	}

	details.IsInstalled = true
	details.MinerBinary = executablePath // Set the full path to the miner binary

	// Try to get the version from the executable
	cmd := exec.Command(executablePath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		details.Version = "Unknown (could not run executable)"
		return details, nil
	}

	// XMRig version output is typically "XMRig 6.18.0"
	fields := strings.Fields(out.String())
	if len(fields) >= 2 {
		details.Version = fields[1]
	} else {
		details.Version = "Unknown (could not parse version)"
	}

	// Update the XMRigMiner struct's Path and MinerBinary fields
	m.Path = details.Path
	m.MinerBinary = details.MinerBinary

	return details, nil
}

// Start the miner
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
	if config.TLS { // If TLS is true in config, ensure --tls is passed if not already in config file
		args = append(args, "--tls")
	}
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
	// HugePages is handled by config file, but --no-huge-pages is a CLI option
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
	// The API settings in m.API are used for GetStats, but CLI options can override for starting the miner
	if m.API.Enabled { // Only add API related CLI args if API is generally enabled
		if config.APIWorkerID != "" {
			args = append(args, "--api-worker-id", config.APIWorkerID)
		}
		if config.APIID != "" {
			args = append(args, "--api-id", config.APIID)
		}
		// Prefer config.HTTPHost/Port, fallback to m.API, then to XMRig defaults
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

	m.cmd = exec.Command(m.MinerBinary, args...)
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

// Stop the miner
func (m *XMRigMiner) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running || m.cmd == nil {
		return errors.New("miner is not running")
	}

	// Kill the process. The goroutine in Start() will handle Wait() and state change.
	return m.cmd.Process.Kill()
}

// GetStats returns the stats for the miner
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

func (m *XMRigMiner) createConfig(config *Config) error {
	configPath, err := xdg.ConfigFile("lethean-desktop/xmrig.json")
	if err != nil {
		// Fallback to home directory if XDG is not available
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

	// Create the config
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

	// Write the config to the file
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.ConfigPath, data, 0644)
}

func (m *XMRigMiner) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
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

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

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

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// Sanitize the header name to prevent path traversal
		cleanedName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanedName, "..") || strings.HasPrefix(cleanedName, "/") || cleanedName == "." {
			continue
		}

		target := filepath.Join(dest, cleanedName)
		rel, err := filepath.Rel(dest, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file to wait until all operations have completed.
			f.Close()
		}
	}
}
