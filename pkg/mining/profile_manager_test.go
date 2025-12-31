package mining

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// setupTestProfileManager creates a ProfileManager with a temp config path.
func setupTestProfileManager(t *testing.T) (*ProfileManager, func()) {
	tmpDir, err := os.MkdirTemp("", "profile-manager-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	configPath := filepath.Join(tmpDir, "mining_profiles.json")

	pm := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: configPath,
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return pm, cleanup
}

func TestProfileManagerCreate(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	profile := &MiningProfile{
		Name:      "Test Profile",
		MinerType: "xmrig",
		Config:    RawConfig(`{"pool": "test.pool.com:3333"}`),
	}

	created, err := pm.CreateProfile(profile)
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	if created.ID == "" {
		t.Error("created profile should have an ID")
	}

	if created.Name != "Test Profile" {
		t.Errorf("expected name 'Test Profile', got '%s'", created.Name)
	}

	// Verify it's stored
	retrieved, exists := pm.GetProfile(created.ID)
	if !exists {
		t.Error("profile should exist after creation")
	}

	if retrieved.Name != created.Name {
		t.Errorf("retrieved name doesn't match: expected '%s', got '%s'", created.Name, retrieved.Name)
	}
}

func TestProfileManagerGet(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	// Get non-existent profile
	_, exists := pm.GetProfile("non-existent-id")
	if exists {
		t.Error("GetProfile should return false for non-existent ID")
	}

	// Create and get
	profile := &MiningProfile{
		Name:      "Get Test",
		MinerType: "xmrig",
	}
	created, _ := pm.CreateProfile(profile)

	retrieved, exists := pm.GetProfile(created.ID)
	if !exists {
		t.Error("GetProfile should return true for existing ID")
	}

	if retrieved.ID != created.ID {
		t.Error("GetProfile returned wrong profile")
	}
}

func TestProfileManagerGetAll(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	// Empty list initially
	profiles := pm.GetAllProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles initially, got %d", len(profiles))
	}

	// Create multiple profiles
	for i := 0; i < 3; i++ {
		pm.CreateProfile(&MiningProfile{
			Name:      "Profile",
			MinerType: "xmrig",
		})
	}

	profiles = pm.GetAllProfiles()
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(profiles))
	}
}

func TestProfileManagerUpdate(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	// Update non-existent profile
	err := pm.UpdateProfile(&MiningProfile{ID: "non-existent"})
	if err == nil {
		t.Error("UpdateProfile should fail for non-existent profile")
	}

	// Create profile
	profile := &MiningProfile{
		Name:      "Original Name",
		MinerType: "xmrig",
	}
	created, _ := pm.CreateProfile(profile)

	// Update it
	created.Name = "Updated Name"
	created.MinerType = "ttminer"
	err = pm.UpdateProfile(created)
	if err != nil {
		t.Fatalf("failed to update profile: %v", err)
	}

	// Verify update
	retrieved, _ := pm.GetProfile(created.ID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", retrieved.Name)
	}
	if retrieved.MinerType != "ttminer" {
		t.Errorf("expected miner type 'ttminer', got '%s'", retrieved.MinerType)
	}
}

func TestProfileManagerDelete(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	// Delete non-existent profile
	err := pm.DeleteProfile("non-existent")
	if err == nil {
		t.Error("DeleteProfile should fail for non-existent profile")
	}

	// Create and delete
	profile := &MiningProfile{
		Name:      "Delete Me",
		MinerType: "xmrig",
	}
	created, _ := pm.CreateProfile(profile)

	err = pm.DeleteProfile(created.ID)
	if err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	// Verify deletion
	_, exists := pm.GetProfile(created.ID)
	if exists {
		t.Error("profile should not exist after deletion")
	}
}

func TestProfileManagerPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profile-persist-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "mining_profiles.json")

	// Create first manager and add profile
	pm1 := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: configPath,
	}

	profile := &MiningProfile{
		Name:      "Persistent Profile",
		MinerType: "xmrig",
		Config:    RawConfig(`{"pool": "persist.pool.com"}`),
	}
	created, err := pm1.CreateProfile(profile)
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	// Create second manager with same path - should load existing profile
	pm2 := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: configPath,
	}
	err = pm2.loadProfiles()
	if err != nil {
		t.Fatalf("failed to load profiles: %v", err)
	}

	// Verify profile persisted
	loaded, exists := pm2.GetProfile(created.ID)
	if !exists {
		t.Fatal("profile should be loaded from file")
	}

	if loaded.Name != "Persistent Profile" {
		t.Errorf("expected name 'Persistent Profile', got '%s'", loaded.Name)
	}
}

func TestProfileManagerConcurrency(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent creates
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pm.CreateProfile(&MiningProfile{
				Name:      "Concurrent Profile",
				MinerType: "xmrig",
			})
		}(i)
	}
	wg.Wait()

	profiles := pm.GetAllProfiles()
	if len(profiles) != numGoroutines {
		t.Errorf("expected %d profiles, got %d", numGoroutines, len(profiles))
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pm.GetAllProfiles()
		}()
	}
	wg.Wait()
}

func TestProfileManagerInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profile-invalid-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "mining_profiles.json")

	// Write invalid JSON
	err = os.WriteFile(configPath, []byte("invalid json{{{"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}

	pm := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: configPath,
	}

	err = pm.loadProfiles()
	if err == nil {
		t.Error("loadProfiles should fail with invalid JSON")
	}
}

func TestProfileManagerFileNotFound(t *testing.T) {
	pm := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: "/non/existent/path/profiles.json",
	}

	err := pm.loadProfiles()
	if err == nil {
		t.Error("loadProfiles should fail when file not found")
	}

	if !os.IsNotExist(err) {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestProfileManagerCreateRollback(t *testing.T) {
	pm := &ProfileManager{
		profiles:   make(map[string]*MiningProfile),
		configPath: "/invalid/path/that/cannot/be/written/profiles.json",
	}

	profile := &MiningProfile{
		Name:      "Rollback Test",
		MinerType: "xmrig",
	}

	_, err := pm.CreateProfile(profile)
	if err == nil {
		t.Error("CreateProfile should fail when save fails")
	}

	// Verify rollback - profile should not be in memory
	profiles := pm.GetAllProfiles()
	if len(profiles) != 0 {
		t.Error("failed create should rollback - no profile should be in memory")
	}
}

func TestProfileManagerConfigWithData(t *testing.T) {
	pm, cleanup := setupTestProfileManager(t)
	defer cleanup()

	config := RawConfig(`{
		"pool": "pool.example.com:3333",
		"wallet": "wallet123",
		"threads": 4,
		"algorithm": "rx/0"
	}`)

	profile := &MiningProfile{
		Name:      "Config Test",
		MinerType: "xmrig",
		Config:    config,
	}

	created, err := pm.CreateProfile(profile)
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	retrieved, _ := pm.GetProfile(created.ID)

	// Parse config to verify
	var parsedConfig map[string]interface{}
	err = json.Unmarshal(retrieved.Config, &parsedConfig)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if parsedConfig["pool"] != "pool.example.com:3333" {
		t.Error("config pool value not preserved")
	}
	if parsedConfig["threads"].(float64) != 4 {
		t.Error("config threads value not preserved")
	}
}
