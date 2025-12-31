package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
)

// configMu protects concurrent access to config file operations
var configMu sync.RWMutex

// MinerAutostartConfig represents the configuration for a single miner's autostart settings.
type MinerAutostartConfig struct {
	MinerType string  `json:"minerType"`
	Autostart bool    `json:"autostart"`
	Config    *Config `json:"config,omitempty"` // Store the last used config
}

// DatabaseConfig holds configuration for SQLite database persistence.
type DatabaseConfig struct {
	// Enabled determines if database persistence is active (default: true)
	Enabled bool `json:"enabled"`
	// RetentionDays is how long to keep historical data (default: 30)
	RetentionDays int `json:"retentionDays,omitempty"`
}

// DefaultDatabaseConfig returns the default database configuration.
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Enabled:       true,
		RetentionDays: 30,
	}
}

// MinersConfig represents the overall configuration for all miners, including autostart settings.
type MinersConfig struct {
	Miners   []MinerAutostartConfig `json:"miners"`
	Database DatabaseConfig         `json:"database"`
}

// GetMinersConfigPath returns the path to the miners configuration file.
func GetMinersConfigPath() (string, error) {
	return xdg.ConfigFile("lethean-desktop/miners/config.json")
}

// LoadMinersConfig loads the miners configuration from the file system.
func LoadMinersConfig() (*MinersConfig, error) {
	configMu.RLock()
	defer configMu.RUnlock()

	configPath, err := GetMinersConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not determine miners config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config with defaults if file doesn't exist
			return &MinersConfig{
				Miners:   []MinerAutostartConfig{},
				Database: DefaultDatabaseConfig(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read miners config file: %w", err)
	}

	var cfg MinersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal miners config: %w", err)
	}

	// Apply default database config if not set (for backwards compatibility)
	if cfg.Database.RetentionDays == 0 {
		cfg.Database = DefaultDatabaseConfig()
	}

	return &cfg, nil
}

// SaveMinersConfig saves the miners configuration to the file system.
// Uses atomic write pattern: write to temp file, then rename.
func SaveMinersConfig(cfg *MinersConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	configPath, err := GetMinersConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine miners config path: %w", err)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal miners config: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpFile, err := os.CreateTemp(dir, "miners-config-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0600); err != nil {
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true
	return nil
}
