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

// MinersConfig represents the overall configuration for all miners, including autostart settings.
type MinersConfig struct {
	Miners []MinerAutostartConfig `json:"miners"`
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
			return &MinersConfig{Miners: []MinerAutostartConfig{}}, nil // Return empty config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read miners config file: %w", err)
	}

	var cfg MinersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal miners config: %w", err)
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
