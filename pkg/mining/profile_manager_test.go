package mining

import (
	"encoding/json"
	"testing"

	"github.com/adrg/xdg"
)

func TestProfileManager(t *testing.T) {
	// Isolate config directory for this test
	tempConfigDir := t.TempDir()
	origConfigHome := xdg.ConfigHome
	xdg.ConfigHome = tempConfigDir
	t.Cleanup(func() {
		xdg.ConfigHome = origConfigHome
	})

	pm, err := NewProfileManager()
	if err != nil {
		t.Fatalf("Failed to create ProfileManager: %v", err)
	}

	// Create
	config := Config{Wallet: "test"}
	configBytes, _ := json.Marshal(config)
	profile := &MiningProfile{
		Name:      "Test Profile",
		MinerType: "xmrig",
		Config:    RawConfig(configBytes),
	}

	created, err := pm.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}
	if created.ID == "" {
		t.Error("Created profile has empty ID")
	}

	// Get
	got, exists := pm.GetProfile(created.ID)
	if !exists {
		t.Error("Failed to get profile")
	}
	if got.Name != "Test Profile" {
		t.Errorf("Expected name 'Test Profile', got '%s'", got.Name)
	}

	// List
	all := pm.GetAllProfiles()
	if len(all) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(all))
	}

	// Update
	created.Name = "Updated Profile"
	err = pm.UpdateProfile(created)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}
	got, _ = pm.GetProfile(created.ID)
	if got.Name != "Updated Profile" {
		t.Errorf("Expected name 'Updated Profile', got '%s'", got.Name)
	}

	// Delete
	err = pm.DeleteProfile(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}
	_, exists = pm.GetProfile(created.ID)
	if exists {
		t.Error("Profile should have been deleted")
	}
}
