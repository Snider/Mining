package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
)

const settingsFileName = "settings.json"

// WindowState stores the last window position and size
type WindowState struct {
	X      int  `json:"x"`
	Y      int  `json:"y"`
	Width  int  `json:"width"`
	Height int  `json:"height"`
	Maximized bool `json:"maximized"`
}

// MinerDefaults stores default configuration for miners
type MinerDefaults struct {
	DefaultPool          string `json:"defaultPool,omitempty"`
	DefaultWallet        string `json:"defaultWallet,omitempty"`
	DefaultAlgorithm     string `json:"defaultAlgorithm,omitempty"`
	CPUMaxThreadsHint    int    `json:"cpuMaxThreadsHint,omitempty"`    // Default CPU throttle percentage
	CPUThrottleThreshold int    `json:"cpuThrottleThreshold,omitempty"` // Throttle when CPU exceeds this %
}

// AppSettings stores application-wide settings
type AppSettings struct {
	// Window settings
	Window WindowState `json:"window"`

	// Behavior settings
	StartOnBoot       bool `json:"startOnBoot"`
	MinimizeToTray    bool `json:"minimizeToTray"`
	StartMinimized    bool `json:"startMinimized"`
	AutostartMiners   bool `json:"autostartMiners"`
	ShowNotifications bool `json:"showNotifications"`

	// Mining settings
	MinerDefaults          MinerDefaults `json:"minerDefaults"`
	PauseOnBattery         bool          `json:"pauseOnBattery"`
	PauseOnUserActive      bool          `json:"pauseOnUserActive"`
	PauseOnUserActiveDelay int           `json:"pauseOnUserActiveDelay"` // Seconds of inactivity before resuming

	// Performance settings
	EnableCPUThrottle      bool `json:"enableCpuThrottle"`
	CPUThrottlePercent     int  `json:"cpuThrottlePercent"`     // Target max CPU % when throttling
	CPUMonitorInterval     int  `json:"cpuMonitorInterval"`     // Seconds between CPU checks
	AutoThrottleOnHighTemp bool `json:"autoThrottleOnHighTemp"` // Throttle when CPU temp is high

	// Theme
	Theme string `json:"theme"` // "light", "dark", "system"
}

// DefaultSettings returns sensible defaults for app settings
func DefaultSettings() *AppSettings {
	return &AppSettings{
		Window: WindowState{
			Width:  1400,
			Height: 900,
		},
		StartOnBoot:       false,
		MinimizeToTray:    true,
		StartMinimized:    false,
		AutostartMiners:   false,
		ShowNotifications: true,
		MinerDefaults: MinerDefaults{
			CPUMaxThreadsHint:    50, // Default to 50% CPU
			CPUThrottleThreshold: 80, // Throttle if CPU > 80%
		},
		PauseOnBattery:         true,
		PauseOnUserActive:      false,
		PauseOnUserActiveDelay: 60,
		EnableCPUThrottle:      false,
		CPUThrottlePercent:     70,
		CPUMonitorInterval:     5,
		AutoThrottleOnHighTemp: false,
		Theme:                  "system",
	}
}

// SettingsManager handles loading and saving app settings
type SettingsManager struct {
	mu           sync.RWMutex
	settings     *AppSettings
	settingsPath string
}

// NewSettingsManager creates a new settings manager
func NewSettingsManager() (*SettingsManager, error) {
	settingsPath, err := xdg.ConfigFile(filepath.Join("lethean-desktop", settingsFileName))
	if err != nil {
		return nil, fmt.Errorf("could not resolve settings path: %w", err)
	}

	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	if err := sm.Load(); err != nil {
		// If file doesn't exist, use defaults and save them
		if os.IsNotExist(err) {
			if saveErr := sm.Save(); saveErr != nil {
				return nil, fmt.Errorf("could not save default settings: %w", saveErr)
			}
		} else {
			return nil, fmt.Errorf("could not load settings: %w", err)
		}
	}

	return sm, nil
}

// Load reads settings from disk
func (sm *SettingsManager) Load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.settingsPath)
	if err != nil {
		return err
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	sm.settings = &settings
	return nil
}

// Save writes settings to disk
func (sm *SettingsManager) Save() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.settingsPath, data, 0600)
}

// Get returns a copy of the current settings
func (sm *SettingsManager) Get() *AppSettings {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	copy := *sm.settings
	return &copy
}

// Update applies changes to settings and saves
func (sm *SettingsManager) Update(fn func(*AppSettings)) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	fn(sm.settings)

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.settingsPath, data, 0600)
}

// UpdateWindowState saves the current window state
func (sm *SettingsManager) UpdateWindowState(x, y, width, height int, maximized bool) error {
	return sm.Update(func(s *AppSettings) {
		s.Window.X = x
		s.Window.Y = y
		s.Window.Width = width
		s.Window.Height = height
		s.Window.Maximized = maximized
	})
}

// GetWindowState returns the saved window state
func (sm *SettingsManager) GetWindowState() WindowState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.settings.Window
}

// SetStartOnBoot enables/disables start on boot
func (sm *SettingsManager) SetStartOnBoot(enabled bool) error {
	return sm.Update(func(s *AppSettings) {
		s.StartOnBoot = enabled
	})
}

// SetAutostartMiners enables/disables miner autostart
func (sm *SettingsManager) SetAutostartMiners(enabled bool) error {
	return sm.Update(func(s *AppSettings) {
		s.AutostartMiners = enabled
	})
}

// SetCPUThrottle configures CPU throttling
func (sm *SettingsManager) SetCPUThrottle(enabled bool, percent int) error {
	return sm.Update(func(s *AppSettings) {
		s.EnableCPUThrottle = enabled
		if percent > 0 && percent <= 100 {
			s.CPUThrottlePercent = percent
		}
	})
}

// SetMinerDefaults updates default miner configuration
func (sm *SettingsManager) SetMinerDefaults(defaults MinerDefaults) error {
	return sm.Update(func(s *AppSettings) {
		s.MinerDefaults = defaults
	})
}
