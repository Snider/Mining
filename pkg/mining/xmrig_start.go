package mining

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Start launches the XMRig miner with the specified configuration.
func (m *XMRigMiner) Start(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Running {
		return errors.New("miner is already running")
	}

	// If the binary path isn't set, run CheckInstallation to find it.
	if m.MinerBinary == "" {
		if _, err := m.CheckInstallation(); err != nil {
			return err // Propagate the detailed error from CheckInstallation
		}
	}

	if m.API != nil && config.HTTPPort != 0 {
		m.API.ListenPort = config.HTTPPort
	} else if m.API != nil && m.API.ListenPort == 0 {
		return errors.New("miner API port not assigned")
	}

	if config.Pool != "" && config.Wallet != "" {
		if err := m.createConfig(config); err != nil {
			return err
		}
	} else {
		// Use the centralized helper to get the instance-specific config path
		configPath, err := getXMRigConfigPath(m.Name)
		if err != nil {
			return fmt.Errorf("could not determine config file path: %w", err)
		}
		m.ConfigPath = configPath
		if _, err := os.Stat(m.ConfigPath); os.IsNotExist(err) {
			return errors.New("config file does not exist and no pool/wallet provided to create one")
		}
	}

	args := []string{"-c", m.ConfigPath}

	if m.API != nil && m.API.Enabled {
		args = append(args, "--http-host", m.API.ListenHost, "--http-port", fmt.Sprintf("%d", m.API.ListenPort))
	}

	addCliArgs(config, &args)

	log.Printf("Executing XMRig command: %s %s", m.MinerBinary, strings.Join(args, " "))

	m.cmd = exec.Command(m.MinerBinary, args...)

	// Create stdin pipe for console commands
	stdinPipe, err := m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	m.stdinPipe = stdinPipe

	// Always capture output to LogBuffer
	if m.LogBuffer != nil {
		m.cmd.Stdout = m.LogBuffer
		m.cmd.Stderr = m.LogBuffer
	}
	// Also output to console if requested
	if config.LogOutput {
		m.cmd.Stdout = io.MultiWriter(m.LogBuffer, os.Stdout)
		m.cmd.Stderr = io.MultiWriter(m.LogBuffer, os.Stderr)
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

// Stop terminates the miner process and cleans up the instance-specific config file.
func (m *XMRigMiner) Stop() error {
	// Call the base Stop to kill the process
	if err := m.BaseMiner.Stop(); err != nil {
		return err
	}

	// Clean up the instance-specific config file
	if m.ConfigPath != "" {
		os.Remove(m.ConfigPath) // Ignore error if it doesn't exist
	}

	return nil
}

// addCliArgs is a helper to append command line arguments based on the config.
func addCliArgs(config *Config, args *[]string) {
	if config.Pool != "" {
		*args = append(*args, "-o", config.Pool)
	}
	if config.Wallet != "" {
		*args = append(*args, "-u", config.Wallet)
	}
	if config.Threads != 0 {
		*args = append(*args, "-t", fmt.Sprintf("%d", config.Threads))
	}
	if !config.HugePages {
		*args = append(*args, "--no-huge-pages")
	}
	if config.TLS {
		*args = append(*args, "--tls")
	}
	*args = append(*args, "--donate-level", "1")
}

// createConfig creates a JSON configuration file for the XMRig miner.
func (m *XMRigMiner) createConfig(config *Config) error {
	// Use the centralized helper to get the instance-specific config path
	configPath, err := getXMRigConfigPath(m.Name)
	if err != nil {
		return err
	}
	m.ConfigPath = configPath

	if err := os.MkdirAll(filepath.Dir(m.ConfigPath), 0755); err != nil {
		return err
	}

	apiListen := "127.0.0.1:0"
	if m.API != nil {
		apiListen = fmt.Sprintf("%s:%d", m.API.ListenHost, m.API.ListenPort)
	}

	cpuConfig := map[string]interface{}{
		"enabled":    true,
		"huge-pages": config.HugePages,
	}

	// Set thread count or max-threads-hint for CPU throttling
	if config.Threads > 0 {
		cpuConfig["threads"] = config.Threads
	}
	if config.CPUMaxThreadsHint > 0 {
		cpuConfig["max-threads-hint"] = config.CPUMaxThreadsHint
	}
	if config.CPUPriority > 0 {
		cpuConfig["priority"] = config.CPUPriority
	}

	// Build pools array - CPU pool first
	cpuPool := map[string]interface{}{
		"url":       config.Pool,
		"user":      config.Wallet,
		"pass":      "x",
		"keepalive": true,
		"tls":       config.TLS,
	}
	// Add algo or coin (coin takes precedence for algorithm auto-detection)
	if config.Coin != "" {
		cpuPool["coin"] = config.Coin
	} else if config.Algo != "" {
		cpuPool["algo"] = config.Algo
	}
	pools := []map[string]interface{}{cpuPool}

	// Add separate GPU pool if configured
	if config.GPUEnabled && config.GPUPool != "" {
		gpuWallet := config.GPUWallet
		if gpuWallet == "" {
			gpuWallet = config.Wallet // Default to main wallet
		}
		gpuPass := config.GPUPassword
		if gpuPass == "" {
			gpuPass = "x"
		}
		gpuPool := map[string]interface{}{
			"url":       config.GPUPool,
			"user":      gpuWallet,
			"pass":      gpuPass,
			"keepalive": true,
		}
		// Add GPU algo (typically etchash, ethash, kawpow, progpowz for GPU mining)
		if config.GPUAlgo != "" {
			gpuPool["algo"] = config.GPUAlgo
		}
		pools = append(pools, gpuPool)
	}

	// Build OpenCL (AMD/Intel GPU) config
	// GPU mining requires explicit device selection - no auto-picking
	openclConfig := map[string]interface{}{
		"enabled": config.GPUEnabled && config.OpenCL && config.Devices != "",
	}
	if config.GPUEnabled && config.OpenCL && config.Devices != "" {
		// User must explicitly specify devices (e.g., "0" or "0,1")
		openclConfig["devices"] = config.Devices
		if config.GPUIntensity > 0 {
			openclConfig["intensity"] = config.GPUIntensity
		}
		if config.GPUThreads > 0 {
			openclConfig["threads"] = config.GPUThreads
		}
	}

	// Build CUDA (NVIDIA GPU) config
	// GPU mining requires explicit device selection - no auto-picking
	cudaConfig := map[string]interface{}{
		"enabled": config.GPUEnabled && config.CUDA && config.Devices != "",
	}
	if config.GPUEnabled && config.CUDA && config.Devices != "" {
		// User must explicitly specify devices (e.g., "0" or "0,1")
		cudaConfig["devices"] = config.Devices
		if config.GPUIntensity > 0 {
			cudaConfig["intensity"] = config.GPUIntensity
		}
		if config.GPUThreads > 0 {
			cudaConfig["threads"] = config.GPUThreads
		}
	}

	c := map[string]interface{}{
		"api": map[string]interface{}{
			"enabled":    m.API != nil && m.API.Enabled,
			"listen":     apiListen,
			"restricted": true,
		},
		"pools":            pools,
		"cpu":              cpuConfig,
		"opencl":           openclConfig,
		"cuda":             cudaConfig,
		"pause-on-active":  config.PauseOnActive,
		"pause-on-battery": config.PauseOnBattery,
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.ConfigPath, data, 0600)
}
