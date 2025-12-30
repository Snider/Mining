package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// MiningService exposes mining functionality to the Wails frontend.
type MiningService struct {
	manager     *mining.Manager
	profileMgr  *mining.ProfileManager
	settingsMgr *mining.SettingsManager
}

// NewMiningService creates a new mining service with an initialized manager.
func NewMiningService() *MiningService {
	manager := mining.NewManager()
	profileMgr, _ := mining.NewProfileManager()
	settingsMgr, _ := mining.NewSettingsManager()
	return &MiningService{
		manager:     manager,
		profileMgr:  profileMgr,
		settingsMgr: settingsMgr,
	}
}

// SystemInfo represents system information for the frontend.
type SystemInfo struct {
	Platform string             `json:"platform"`
	CPU      string             `json:"cpu"`
	Cores    int                `json:"cores"`
	MemoryGB int                `json:"memory_gb"`
	Miners   []MinerInstallInfo `json:"installed_miners_info"`
}

// MinerInstallInfo represents installed miner information.
type MinerInstallInfo struct {
	MinerType   string `json:"miner_type"`
	IsInstalled bool   `json:"is_installed"`
	Version     string `json:"version"`
	Path        string `json:"path"`
}

// MinerStatus represents a running miner's status.
type MinerStatus struct {
	Name      string                     `json:"name"`
	Running   bool                       `json:"running"`
	MinerType string                     `json:"miner_type"`
	Stats     *mining.PerformanceMetrics `json:"stats,omitempty"`
}

// Profile represents a mining profile for the frontend.
type Profile struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	MinerType string                 `json:"minerType"`
	Config    map[string]interface{} `json:"config"`
}

// GetSystemInfo returns system information and installed miners.
func (s *MiningService) GetSystemInfo() (*SystemInfo, error) {
	cpuInfo, _ := cpu.Info()
	cpuName := "Unknown"
	if len(cpuInfo) > 0 {
		cpuName = cpuInfo[0].ModelName
	}

	memInfo, _ := mem.VirtualMemory()
	memGB := 0
	if memInfo != nil {
		memGB = int(memInfo.Total / 1024 / 1024 / 1024)
	}

	miners := []MinerInstallInfo{}
	// Check installation for each miner type by creating temporary instances
	for _, minerType := range []string{"xmrig", "tt-miner"} {
		var miner mining.Miner
		switch minerType {
		case "xmrig":
			miner = mining.NewXMRigMiner()
		case "tt-miner":
			miner = mining.NewTTMiner()
		}
		if miner != nil {
			details, err := miner.CheckInstallation()
			if err == nil && details != nil {
				miners = append(miners, MinerInstallInfo{
					MinerType:   minerType,
					IsInstalled: details.IsInstalled,
					Version:     details.Version,
					Path:        details.Path,
				})
			} else {
				miners = append(miners, MinerInstallInfo{
					MinerType:   minerType,
					IsInstalled: false,
				})
			}
		}
	}

	return &SystemInfo{
		Platform: runtime.GOOS,
		CPU:      cpuName,
		Cores:    runtime.NumCPU(),
		MemoryGB: memGB,
		Miners:   miners,
	}, nil
}

// ListMiners returns all running miners.
func (s *MiningService) ListMiners() []MinerStatus {
	miners := s.manager.ListMiners()
	result := make([]MinerStatus, len(miners))
	for i, m := range miners {
		stats, _ := m.GetStats()
		result[i] = MinerStatus{
			Name:      m.GetName(),
			Running:   true, // If it's in the list, it's running
			MinerType: getMinerType(m),
			Stats:     stats,
		}
	}
	return result
}

// getMinerType extracts the miner type from a miner instance.
func getMinerType(m mining.Miner) string {
	name := m.GetName()
	if strings.HasPrefix(name, "xmrig") {
		return "xmrig"
	}
	if strings.HasPrefix(name, "tt-miner") || strings.HasPrefix(name, "ttminer") {
		return "tt-miner"
	}
	return "unknown"
}

// StartMiner starts a miner with the given configuration.
func (s *MiningService) StartMiner(minerType string, config *mining.Config) (string, error) {
	miner, err := s.manager.StartMiner(minerType, config)
	if err != nil {
		return "", err
	}
	return miner.GetName(), nil
}

// StartMinerFromProfile starts a miner using a saved profile.
func (s *MiningService) StartMinerFromProfile(profileID string) (string, error) {
	if s.profileMgr == nil {
		return "", fmt.Errorf("profile manager not initialized")
	}
	profile, ok := s.profileMgr.GetProfile(profileID)
	if !ok {
		return "", fmt.Errorf("profile not found: %s", profileID)
	}

	// Convert RawConfig to *Config
	var config mining.Config
	if profile.Config != nil {
		if err := json.Unmarshal(profile.Config, &config); err != nil {
			return "", fmt.Errorf("failed to parse profile config: %w", err)
		}
	}

	miner, err := s.manager.StartMiner(profile.MinerType, &config)
	if err != nil {
		return "", err
	}
	return miner.GetName(), nil
}

// StopMiner stops a running miner by name.
func (s *MiningService) StopMiner(name string) error {
	return s.manager.StopMiner(name)
}

// GetMinerStats returns stats for a specific miner.
func (s *MiningService) GetMinerStats(name string) (*mining.PerformanceMetrics, error) {
	miner, err := s.manager.GetMiner(name)
	if err != nil {
		return nil, err
	}
	return miner.GetStats()
}

// GetMinerLogs returns log lines for a specific miner.
func (s *MiningService) GetMinerLogs(name string) ([]string, error) {
	miner, err := s.manager.GetMiner(name)
	if err != nil {
		return nil, err
	}
	return miner.GetLogs(), nil
}

// InstallMiner installs a miner of the given type.
func (s *MiningService) InstallMiner(minerType string) error {
	var miner mining.Miner
	switch minerType {
	case "xmrig":
		miner = mining.NewXMRigMiner()
	case "tt-miner":
		miner = mining.NewTTMiner()
	default:
		return fmt.Errorf("unsupported miner type: %s", minerType)
	}
	return miner.Install()
}

// UninstallMiner uninstalls a miner of the given type.
func (s *MiningService) UninstallMiner(minerType string) error {
	return s.manager.UninstallMiner(minerType)
}

// GetProfiles returns all saved mining profiles.
func (s *MiningService) GetProfiles() ([]Profile, error) {
	if s.profileMgr == nil {
		return []Profile{}, nil
	}
	profiles := s.profileMgr.GetAllProfiles()

	result := make([]Profile, len(profiles))
	for i, p := range profiles {
		// Convert RawConfig to map for frontend
		var configMap map[string]interface{}
		if p.Config != nil {
			json.Unmarshal(p.Config, &configMap)
		}
		result[i] = Profile{
			ID:        p.ID,
			Name:      p.Name,
			MinerType: p.MinerType,
			Config:    configMap,
		}
	}
	return result, nil
}

// CreateProfile creates a new mining profile.
func (s *MiningService) CreateProfile(name, minerType string, config map[string]interface{}) (*Profile, error) {
	if s.profileMgr == nil {
		return nil, fmt.Errorf("profile manager not initialized")
	}

	// Convert map to RawConfig (JSON bytes)
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	newProfile := &mining.MiningProfile{
		Name:      name,
		MinerType: minerType,
		Config:    mining.RawConfig(configBytes),
	}

	profile, err := s.profileMgr.CreateProfile(newProfile)
	if err != nil {
		return nil, err
	}

	return &Profile{
		ID:        profile.ID,
		Name:      profile.Name,
		MinerType: profile.MinerType,
		Config:    config,
	}, nil
}

// DeleteProfile deletes a profile by ID.
func (s *MiningService) DeleteProfile(id string) error {
	if s.profileMgr == nil {
		return nil
	}
	return s.profileMgr.DeleteProfile(id)
}

// GetHashrateHistory returns hashrate history for a miner.
func (s *MiningService) GetHashrateHistory(name string) []mining.HashratePoint {
	history, _ := s.manager.GetMinerHashrateHistory(name)
	return history
}

// SendStdin sends input to a miner's stdin.
func (s *MiningService) SendStdin(name, input string) error {
	miner, err := s.manager.GetMiner(name)
	if err != nil {
		return err
	}
	return miner.WriteStdin(input)
}

// Shutdown gracefully shuts down all miners.
func (s *MiningService) Shutdown() {
	s.manager.Stop()
}

// === Settings Methods ===

// GetSettings returns the current app settings
func (s *MiningService) GetSettings() (*mining.AppSettings, error) {
	if s.settingsMgr == nil {
		return mining.DefaultSettings(), nil
	}
	return s.settingsMgr.Get(), nil
}

// SaveSettings saves the app settings
func (s *MiningService) SaveSettings(settings *mining.AppSettings) error {
	if s.settingsMgr == nil {
		return fmt.Errorf("settings manager not initialized")
	}
	return s.settingsMgr.Update(func(s *mining.AppSettings) {
		*s = *settings
	})
}

// SaveWindowState saves the window position and size
func (s *MiningService) SaveWindowState(x, y, width, height int, maximized bool) error {
	if s.settingsMgr == nil {
		return nil
	}
	return s.settingsMgr.UpdateWindowState(x, y, width, height, maximized)
}

// WindowState represents window position and size for the frontend
type WindowState struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

// GetWindowState returns the saved window state
func (s *MiningService) GetWindowState() *WindowState {
	if s.settingsMgr == nil {
		return &WindowState{Width: 1400, Height: 900}
	}
	state := s.settingsMgr.GetWindowState()
	return &WindowState{
		X:         state.X,
		Y:         state.Y,
		Width:     state.Width,
		Height:    state.Height,
		Maximized: state.Maximized,
	}
}

// SetStartOnBoot enables/disables start on system boot
func (s *MiningService) SetStartOnBoot(enabled bool) error {
	if s.settingsMgr == nil {
		return nil
	}
	return s.settingsMgr.SetStartOnBoot(enabled)
}

// SetAutostartMiners enables/disables automatic miner start
func (s *MiningService) SetAutostartMiners(enabled bool) error {
	if s.settingsMgr == nil {
		return nil
	}
	return s.settingsMgr.SetAutostartMiners(enabled)
}

// SetCPUThrottle configures CPU throttling settings
func (s *MiningService) SetCPUThrottle(enabled bool, maxPercent int) error {
	if s.settingsMgr == nil {
		return nil
	}
	return s.settingsMgr.SetCPUThrottle(enabled, maxPercent)
}

// SetMinerDefaults updates default miner configuration
func (s *MiningService) SetMinerDefaults(defaults mining.MinerDefaults) error {
	if s.settingsMgr == nil {
		return nil
	}
	return s.settingsMgr.SetMinerDefaults(defaults)
}
