package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/google/uuid"
)

const profileConfigFileName = "mining_profiles.json"

// ProfileManager handles CRUD operations for MiningProfiles.
type ProfileManager struct {
	mu         sync.RWMutex
	profiles   map[string]*MiningProfile
	configPath string
}

// NewProfileManager creates and initializes a new ProfileManager.
func NewProfileManager() (*ProfileManager, error) {
	configPath, err := xdg.ConfigFile(filepath.Join("lethean-desktop", profileConfigFileName))
	if err != nil {
		return nil, fmt.Errorf("could not resolve config path: %w", err)
	}

	pm := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: configPath,
	}

	if err := pm.loadProfiles(); err != nil {
		// If the file doesn't exist, that's fine, but any other error is a problem.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not load profiles: %w", err)
		}
	}

	return pm, nil
}

// loadProfiles reads the profiles from the JSON file into memory.
func (pm *ProfileManager) loadProfiles() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		return err
	}

	var profiles []*MiningProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return err
	}

	pm.profiles = make(map[string]*MiningProfile)
	for _, p := range profiles {
		pm.profiles[p.ID] = p
	}

	return nil
}

// saveProfiles writes the current profiles from memory to the JSON file.
// This is an internal method and assumes the caller holds the appropriate lock.
func (pm *ProfileManager) saveProfiles() error {
	profileList := make([]*MiningProfile, 0, len(pm.profiles))
	for _, p := range pm.profiles {
		profileList = append(profileList, p)
	}

	data, err := json.MarshalIndent(profileList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.configPath, data, 0600)
}

// CreateProfile adds a new profile and saves it.
func (pm *ProfileManager) CreateProfile(profile *MiningProfile) (*MiningProfile, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	profile.ID = uuid.New().String()
	pm.profiles[profile.ID] = profile

	if err := pm.saveProfiles(); err != nil {
		// Rollback
		delete(pm.profiles, profile.ID)
		return nil, err
	}

	return profile, nil
}

// GetProfile retrieves a profile by its ID.
func (pm *ProfileManager) GetProfile(id string) (*MiningProfile, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	profile, exists := pm.profiles[id]
	return profile, exists
}

// GetAllProfiles returns a list of all profiles.
func (pm *ProfileManager) GetAllProfiles() []*MiningProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profileList := make([]*MiningProfile, 0, len(pm.profiles))
	for _, p := range pm.profiles {
		profileList = append(profileList, p)
	}
	return profileList
}

// UpdateProfile modifies an existing profile.
func (pm *ProfileManager) UpdateProfile(profile *MiningProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[profile.ID]; !exists {
		return fmt.Errorf("profile with ID %s not found", profile.ID)
	}
	pm.profiles[profile.ID] = profile

	return pm.saveProfiles()
}

// DeleteProfile removes a profile by its ID.
func (pm *ProfileManager) DeleteProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[id]; !exists {
		return fmt.Errorf("profile with ID %s not found", id)
	}
	delete(pm.profiles, id)

	return pm.saveProfiles()
}
