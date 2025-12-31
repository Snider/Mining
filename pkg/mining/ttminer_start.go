package mining

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Snider/Mining/pkg/logging"
)

// Start launches the TT-Miner with the given configuration.
func (m *TTMiner) Start(config *Config) error {
	// Check installation BEFORE acquiring lock (CheckInstallation takes its own locks)
	m.mu.RLock()
	needsInstallCheck := m.MinerBinary == ""
	m.mu.RUnlock()

	if needsInstallCheck {
		if _, err := m.CheckInstallation(); err != nil {
			return err // Propagate the detailed error from CheckInstallation
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Running {
		return errors.New("miner is already running")
	}

	if m.API != nil && config.HTTPPort != 0 {
		m.API.ListenPort = config.HTTPPort
	} else if m.API != nil && m.API.ListenPort == 0 {
		return errors.New("miner API port not assigned")
	}

	// Build command line arguments for TT-Miner
	args := m.buildArgs(config)

	logging.Info("executing TT-Miner command", logging.Fields{"binary": m.MinerBinary, "args": strings.Join(args, " ")})

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
		stdinPipe.Close()
		return fmt.Errorf("failed to start TT-Miner: %w", err)
	}

	m.Running = true

	// Capture cmd locally to avoid race with Stop()
	cmd := m.cmd
	go func() {
		// Use a channel to detect if Wait() completes
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		// Wait with timeout to prevent goroutine leak on zombie processes
		var err error
		select {
		case err = <-done:
			// Normal exit
		case <-time.After(5 * time.Minute):
			// Process didn't exit after 5 minutes - force cleanup
			logging.Warn("TT-Miner process wait timeout, forcing cleanup")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			err = <-done // Wait for the inner goroutine to finish
		}

		m.mu.Lock()
		// Only clear if this is still the same command (not restarted)
		if m.cmd == cmd {
			m.Running = false
			m.cmd = nil
		}
		m.mu.Unlock()
		if err != nil {
			logging.Debug("TT-Miner exited with error", logging.Fields{"error": err})
		} else {
			logging.Debug("TT-Miner exited normally")
		}
	}()

	return nil
}

// buildArgs constructs the command line arguments for TT-Miner
func (m *TTMiner) buildArgs(config *Config) []string {
	var args []string

	// Pool configuration
	if config.Pool != "" {
		args = append(args, "-P", config.Pool)
	}

	// Wallet/user configuration
	if config.Wallet != "" {
		args = append(args, "-u", config.Wallet)
	}

	// Password
	if config.Password != "" {
		args = append(args, "-p", config.Password)
	} else {
		args = append(args, "-p", "x")
	}

	// Algorithm selection
	if config.Algo != "" {
		args = append(args, "-a", config.Algo)
	}

	// API binding for stats collection
	if m.API != nil && m.API.Enabled {
		args = append(args, "-b", fmt.Sprintf("%s:%d", m.API.ListenHost, m.API.ListenPort))
	}

	// GPU device selection (if specified)
	if config.Devices != "" {
		args = append(args, "-d", config.Devices)
	}

	// Intensity (if specified)
	if config.Intensity > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", config.Intensity))
	}

	// Additional CLI arguments
	addTTMinerCliArgs(config, &args)

	return args
}

// addTTMinerCliArgs adds any additional CLI arguments from config
func addTTMinerCliArgs(config *Config, args *[]string) {
	// Add any extra arguments passed via CLIArgs
	if config.CLIArgs != "" {
		extraArgs := strings.Fields(config.CLIArgs)
		for _, arg := range extraArgs {
			// Skip potentially dangerous arguments
			if isValidCLIArg(arg) {
				*args = append(*args, arg)
			} else {
				logging.Warn("skipping invalid CLI argument", logging.Fields{"arg": arg})
			}
		}
	}
}

// isValidCLIArg validates CLI arguments to prevent injection or dangerous patterns
func isValidCLIArg(arg string) bool {
	// Block shell metacharacters and dangerous patterns
	dangerousPatterns := []string{";", "|", "&", "`", "$", "(", ")", "{", "}", "<", ">", "\n", "\r"}
	for _, p := range dangerousPatterns {
		if strings.Contains(arg, p) {
			return false
		}
	}
	// Block arguments that could override security-related settings
	blockedArgs := []string{"--api-access-token", "--api-worker-id"}
	lowerArg := strings.ToLower(arg)
	for _, blocked := range blockedArgs {
		if strings.HasPrefix(lowerArg, blocked) {
			return false
		}
	}
	return true
}
