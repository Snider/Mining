package mining

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Snider/Mining/pkg/database"
)

func setupContainerTestEnv(t *testing.T) func() {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	return func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_DATA_HOME")
	}
}

func TestNewContainer(t *testing.T) {
	config := DefaultContainerConfig()
	container := NewContainer(config)

	if container == nil {
		t.Fatal("NewContainer returned nil")
	}

	if container.IsInitialized() {
		t.Error("Container should not be initialized before Initialize() is called")
	}
}

func TestDefaultContainerConfig(t *testing.T) {
	config := DefaultContainerConfig()

	if !config.Database.Enabled {
		t.Error("Database should be enabled by default")
	}

	if config.Database.RetentionDays != 30 {
		t.Errorf("Expected 30 retention days, got %d", config.Database.RetentionDays)
	}

	if config.ListenAddr != ":9090" {
		t.Errorf("Expected :9090, got %s", config.ListenAddr)
	}

	if config.SimulationMode {
		t.Error("SimulationMode should be false by default")
	}
}

func TestContainer_Initialize(t *testing.T) {
	cleanup := setupContainerTestEnv(t)
	defer cleanup()

	config := DefaultContainerConfig()
	config.Database.Enabled = true
	config.Database.Path = filepath.Join(t.TempDir(), "test.db")
	config.SimulationMode = true // Use simulation mode for faster tests

	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !container.IsInitialized() {
		t.Error("Container should be initialized after Initialize()")
	}

	// Verify services are available
	if container.Manager() == nil {
		t.Error("Manager should not be nil after initialization")
	}

	if container.ProfileManager() == nil {
		t.Error("ProfileManager should not be nil after initialization")
	}

	if container.EventHub() == nil {
		t.Error("EventHub should not be nil after initialization")
	}

	if container.HashrateStore() == nil {
		t.Error("HashrateStore should not be nil after initialization")
	}

	// Cleanup
	if err := container.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestContainer_InitializeTwice(t *testing.T) {
	cleanup := setupContainerTestEnv(t)
	defer cleanup()

	config := DefaultContainerConfig()
	config.Database.Enabled = false
	config.SimulationMode = true

	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		t.Fatalf("First Initialize failed: %v", err)
	}

	// Second initialization should fail
	if err := container.Initialize(ctx); err == nil {
		t.Error("Second Initialize should fail")
	}

	container.Shutdown(ctx)
}

func TestContainer_DatabaseDisabled(t *testing.T) {
	cleanup := setupContainerTestEnv(t)
	defer cleanup()

	config := DefaultContainerConfig()
	config.Database.Enabled = false
	config.SimulationMode = true

	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Should use NopStore when database is disabled
	store := container.HashrateStore()
	if store == nil {
		t.Fatal("HashrateStore should not be nil")
	}

	// NopStore should accept inserts without error
	point := database.HashratePoint{
		Timestamp: time.Now(),
		Hashrate:  1000,
	}
	if err := store.InsertHashratePoint(nil, "test", "xmrig", point, database.ResolutionHigh); err != nil {
		t.Errorf("NopStore insert should not fail: %v", err)
	}

	container.Shutdown(ctx)
}

func TestContainer_SetHashrateStore(t *testing.T) {
	cleanup := setupContainerTestEnv(t)
	defer cleanup()

	config := DefaultContainerConfig()
	config.Database.Enabled = false
	config.SimulationMode = true

	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Inject custom store
	customStore := database.NopStore()
	container.SetHashrateStore(customStore)

	if container.HashrateStore() != customStore {
		t.Error("SetHashrateStore should update the store")
	}

	container.Shutdown(ctx)
}

func TestContainer_StartWithoutInitialize(t *testing.T) {
	config := DefaultContainerConfig()
	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Start(ctx); err == nil {
		t.Error("Start should fail if Initialize was not called")
	}
}

func TestContainer_ShutdownWithoutInitialize(t *testing.T) {
	config := DefaultContainerConfig()
	container := NewContainer(config)
	ctx := context.Background()

	// Shutdown on uninitialized container should not error
	if err := container.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown on uninitialized container should not error: %v", err)
	}
}

func TestContainer_ShutdownChannel(t *testing.T) {
	cleanup := setupContainerTestEnv(t)
	defer cleanup()

	config := DefaultContainerConfig()
	config.Database.Enabled = false
	config.SimulationMode = true

	container := NewContainer(config)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	shutdownCh := container.ShutdownCh()

	// Channel should be open before shutdown
	select {
	case <-shutdownCh:
		t.Error("ShutdownCh should not be closed before Shutdown()")
	default:
		// Expected
	}

	if err := container.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Channel should be closed after shutdown
	select {
	case <-shutdownCh:
		// Expected
	case <-time.After(time.Second):
		t.Error("ShutdownCh should be closed after Shutdown()")
	}
}
