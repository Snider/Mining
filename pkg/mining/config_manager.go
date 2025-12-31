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

// defaultDatabaseConfig returns the default database configuration.
func defaultDatabaseConfig() DatabaseConfig {
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

// getMinersConfigPath returns the path to the miners configuration file.
func getMinersConfigPath() (string, error) {
	return xdg.ConfigFile("lethean-desktop/miners/config.json")
}

// LoadMinersConfig loads the miners configuration from the file system.
func LoadMinersConfig() (*MinersConfig, error) {
	configMu.RLock()
	defer configMu.RUnlock()

	configPath, err := getMinersConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not determine miners config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config with defaults if file doesn't exist
			return &MinersConfig{
				Miners:   []MinerAutostartConfig{},
				Database: defaultDatabaseConfig(),
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
		cfg.Database = defaultDatabaseConfig()
	}

	return &cfg, nil
}

// SaveMinersConfig saves the miners configuration to the file system.
// Uses atomic write pattern: write to temp file, then rename.
func SaveMinersConfig(cfg *MinersConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	configPath, err := getMinersConfigPath()
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

	return AtomicWriteFile(configPath, data, 0600)
}

// UpdateMinersConfig atomically loads, modifies, and saves the miners config.
// This prevents race conditions in read-modify-write operations.
func UpdateMinersConfig(fn func(*MinersConfig) error) error {
	configMu.Lock()
	defer configMu.Unlock()

	configPath, err := getMinersConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine miners config path: %w", err)
	}

	// Load current config
	var cfg MinersConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = MinersConfig{
				Miners:   []MinerAutostartConfig{},
				Database: defaultDatabaseConfig(),
			}
		} else {
			return fmt.Errorf("failed to read miners config file: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to unmarshal miners config: %w", err)
		}
		if cfg.Database.RetentionDays == 0 {
			cfg.Database = defaultDatabaseConfig()
		}
	}

	// Apply the modification
	if err := fn(&cfg); err != nil {
		return err
	}

	// Save atomically
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal miners config: %w", err)
	}

	return AtomicWriteFile(configPath, newData, 0600)
}
