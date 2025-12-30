package mining

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsManager_DefaultSettings(t *testing.T) {
	defaults := DefaultSettings()

	if defaults.Window.Width != 1400 {
		t.Errorf("Expected default width 1400, got %d", defaults.Window.Width)
	}
	if defaults.Window.Height != 900 {
		t.Errorf("Expected default height 900, got %d", defaults.Window.Height)
	}
	if defaults.MinerDefaults.CPUMaxThreadsHint != 50 {
		t.Errorf("Expected default CPU hint 50, got %d", defaults.MinerDefaults.CPUMaxThreadsHint)
	}
	if defaults.MinerDefaults.CPUThrottleThreshold != 80 {
		t.Errorf("Expected default throttle threshold 80, got %d", defaults.MinerDefaults.CPUThrottleThreshold)
	}
	if !defaults.PauseOnBattery {
		t.Error("Expected PauseOnBattery to be true by default")
	}
}

func TestSettingsManager_SaveAndLoad(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Create settings manager with custom path
	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	// Modify settings
	sm.settings.Window.Width = 1920
	sm.settings.Window.Height = 1080
	sm.settings.StartOnBoot = true
	sm.settings.AutostartMiners = true
	sm.settings.CPUThrottlePercent = 50

	// Save
	err := sm.Save()
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Create new manager and load
	sm2 := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}
	err = sm2.Load()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify loaded values
	if sm2.settings.Window.Width != 1920 {
		t.Errorf("Expected width 1920, got %d", sm2.settings.Window.Width)
	}
	if sm2.settings.Window.Height != 1080 {
		t.Errorf("Expected height 1080, got %d", sm2.settings.Window.Height)
	}
	if !sm2.settings.StartOnBoot {
		t.Error("Expected StartOnBoot to be true")
	}
	if !sm2.settings.AutostartMiners {
		t.Error("Expected AutostartMiners to be true")
	}
	if sm2.settings.CPUThrottlePercent != 50 {
		t.Errorf("Expected CPUThrottlePercent 50, got %d", sm2.settings.CPUThrottlePercent)
	}
}

func TestSettingsManager_UpdateWindowState(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	err := sm.UpdateWindowState(100, 200, 800, 600, false)
	if err != nil {
		t.Fatalf("Failed to update window state: %v", err)
	}

	state := sm.GetWindowState()
	if state.X != 100 {
		t.Errorf("Expected X 100, got %d", state.X)
	}
	if state.Y != 200 {
		t.Errorf("Expected Y 200, got %d", state.Y)
	}
	if state.Width != 800 {
		t.Errorf("Expected Width 800, got %d", state.Width)
	}
	if state.Height != 600 {
		t.Errorf("Expected Height 600, got %d", state.Height)
	}
}

func TestSettingsManager_SetCPUThrottle(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	// Test enabling throttle
	err := sm.SetCPUThrottle(true, 30)
	if err != nil {
		t.Fatalf("Failed to set CPU throttle: %v", err)
	}

	settings := sm.Get()
	if !settings.EnableCPUThrottle {
		t.Error("Expected EnableCPUThrottle to be true")
	}
	if settings.CPUThrottlePercent != 30 {
		t.Errorf("Expected CPUThrottlePercent 30, got %d", settings.CPUThrottlePercent)
	}

	// Test invalid percentage (should be ignored)
	err = sm.SetCPUThrottle(true, 150)
	if err != nil {
		t.Fatalf("Failed to set CPU throttle: %v", err)
	}
	settings = sm.Get()
	if settings.CPUThrottlePercent != 30 { // Should remain unchanged
		t.Errorf("Expected CPUThrottlePercent to remain 30, got %d", settings.CPUThrottlePercent)
	}
}

func TestSettingsManager_SetMinerDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	defaults := MinerDefaults{
		DefaultPool:          "stratum+tcp://pool.example.com:3333",
		DefaultWallet:        "wallet123",
		DefaultAlgorithm:     "rx/0",
		CPUMaxThreadsHint:    25,
		CPUThrottleThreshold: 90,
	}

	err := sm.SetMinerDefaults(defaults)
	if err != nil {
		t.Fatalf("Failed to set miner defaults: %v", err)
	}

	settings := sm.Get()
	if settings.MinerDefaults.DefaultPool != "stratum+tcp://pool.example.com:3333" {
		t.Errorf("Expected pool to be set, got %s", settings.MinerDefaults.DefaultPool)
	}
	if settings.MinerDefaults.CPUMaxThreadsHint != 25 {
		t.Errorf("Expected CPUMaxThreadsHint 25, got %d", settings.MinerDefaults.CPUMaxThreadsHint)
	}
}

func TestSettingsManager_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	sm := &SettingsManager{
		settings:     DefaultSettings(),
		settingsPath: settingsPath,
	}

	// Concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				_ = sm.Get()
				sm.UpdateWindowState(n*10, n*10, 800+n, 600+n, false)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	state := sm.GetWindowState()
	if state.Width < 800 || state.Width > 900 {
		t.Errorf("Unexpected width after concurrent access: %d", state.Width)
	}
}
