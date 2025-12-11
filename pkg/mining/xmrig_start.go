package mining

import (
	"encoding/json"
	"errors"
	"fmt"
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
		// Use the centralized helper to get the config path
		configPath, err := getXMRigConfigPath()
		if err != nil {
			return fmt.Errorf("could not determine config file path: %w", err)
		}
		m.ConfigPath = configPath
		if _, err := os.Stat(m.ConfigPath); os.IsNotExist(err) {
			return errors.New("config file does not exist and no pool/wallet provided to create one")
		}
	}

	args := []string{"-c", "\"" + m.ConfigPath + "\""}

	if m.API != nil && m.API.Enabled {
		args = append(args, "--http-host", m.API.ListenHost, "--http-port", fmt.Sprintf("%d", m.API.ListenPort))
	}

	addCliArgs(config, &args)

	log.Printf("Executing XMRig command: %s %s", m.MinerBinary, strings.Join(args, " "))

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
	// Use the centralized helper to get the config path
	configPath, err := getXMRigConfigPath()
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

	c := map[string]interface{}{
		"api": map[string]interface{}{
			"enabled":    m.API != nil && m.API.Enabled,
			"listen":     apiListen,
			"restricted": true,
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
