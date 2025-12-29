package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

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
func SaveMinersConfig(cfg *MinersConfig) error {
	configPath, err := GetMinersConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine miners config path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal miners config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write miners config file: %w", err)
	}
	return nil
}
