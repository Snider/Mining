package mining

import (
	"context"
	"fmt"
	"sync"

	"github.com/Snider/Mining/pkg/database"
	"github.com/Snider/Mining/pkg/logging"
)

// ContainerConfig holds configuration for the service container.
type ContainerConfig struct {
	// Database configuration
	Database database.Config

	// ListenAddr is the address to listen on (e.g., ":9090")
	ListenAddr string

	// DisplayAddr is the address shown in Swagger docs
	DisplayAddr string

	// SwaggerNamespace is the API path prefix
	SwaggerNamespace string

	// SimulationMode enables simulation mode for testing
	SimulationMode bool
}

// DefaultContainerConfig returns sensible defaults for the container.
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		Database: database.Config{
			Enabled:       true,
			RetentionDays: 30,
		},
		ListenAddr:       ":9090",
		DisplayAddr:      "localhost:9090",
		SwaggerNamespace: "/api/v1/mining",
		SimulationMode:   false,
	}
}

// Container manages the lifecycle of all services.
// It provides centralized initialization, dependency injection, and graceful shutdown.
type Container struct {
	config ContainerConfig
	mu     sync.RWMutex

	// Core services
	manager        ManagerInterface
	profileManager *ProfileManager
	nodeService    *NodeService
	eventHub       *EventHub
	service        *Service

	// Database store (interface for testing)
	hashrateStore database.HashrateStore

	// Initialization state
	initialized      bool
	transportStarted bool
	shutdownCh       chan struct{}
}

// NewContainer creates a new service container with the given configuration.
func NewContainer(config ContainerConfig) *Container {
	return &Container{
		config:     config,
		shutdownCh: make(chan struct{}),
	}
}

// Initialize sets up all services in the correct order.
// This should be called before Start().
func (c *Container) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("container already initialized")
	}

	// 1. Initialize database (optional)
	if c.config.Database.Enabled {
		if err := database.Initialize(c.config.Database); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		c.hashrateStore = database.DefaultStore()
		logging.Info("database initialized", logging.Fields{"retention_days": c.config.Database.RetentionDays})
	} else {
		c.hashrateStore = database.NopStore()
		logging.Info("database disabled, using no-op store", nil)
	}

	// 2. Initialize profile manager
	var err error
	c.profileManager, err = NewProfileManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	// 3. Initialize miner manager
	if c.config.SimulationMode {
		c.manager = NewManagerForSimulation()
	} else {
		c.manager = NewManager()
	}

	// 4. Initialize node service (optional - P2P features)
	c.nodeService, err = NewNodeService()
	if err != nil {
		logging.Warn("node service unavailable", logging.Fields{"error": err})
		// Continue without node service - P2P features will be unavailable
	}

	// 5. Initialize event hub for WebSocket
	c.eventHub = NewEventHub()

	// Wire up event hub to manager
	if mgr, ok := c.manager.(*Manager); ok {
		mgr.SetEventHub(c.eventHub)
	}

	c.initialized = true
	logging.Info("service container initialized", nil)
	return nil
}

// Start begins all background services.
func (c *Container) Start(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.initialized {
		return fmt.Errorf("container not initialized")
	}

	// Start event hub
	go c.eventHub.Run()

	// Start node transport if available
	if c.nodeService != nil {
		if err := c.nodeService.StartTransport(); err != nil {
			logging.Warn("failed to start node transport", logging.Fields{"error": err})
		} else {
			c.transportStarted = true
		}
	}

	logging.Info("service container started", nil)
	return nil
}

// Shutdown gracefully stops all services in reverse order.
func (c *Container) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil
	}

	logging.Info("shutting down service container", nil)

	var errs []error

	// 1. Stop service (HTTP server)
	if c.service != nil {
		// Service shutdown is handled externally
	}

	// 2. Stop node transport (only if it was started)
	if c.nodeService != nil && c.transportStarted {
		if err := c.nodeService.StopTransport(); err != nil {
			errs = append(errs, fmt.Errorf("node transport: %w", err))
		}
		c.transportStarted = false
	}

	// 3. Stop event hub
	if c.eventHub != nil {
		c.eventHub.Stop()
	}

	// 4. Stop miner manager
	if mgr, ok := c.manager.(*Manager); ok {
		mgr.Stop()
	}

	// 5. Close database
	if err := database.Close(); err != nil {
		errs = append(errs, fmt.Errorf("database: %w", err))
	}

	c.initialized = false
	close(c.shutdownCh)

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	logging.Info("service container shutdown complete", nil)
	return nil
}

// Manager returns the miner manager.
func (c *Container) Manager() ManagerInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.manager
}

// ProfileManager returns the profile manager.
func (c *Container) ProfileManager() *ProfileManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.profileManager
}

// NodeService returns the node service (may be nil if P2P is unavailable).
func (c *Container) NodeService() *NodeService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeService
}

// EventHub returns the event hub for WebSocket connections.
func (c *Container) EventHub() *EventHub {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eventHub
}

// HashrateStore returns the hashrate store interface.
func (c *Container) HashrateStore() database.HashrateStore {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hashrateStore
}

// SetHashrateStore allows injecting a custom hashrate store (useful for testing).
func (c *Container) SetHashrateStore(store database.HashrateStore) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashrateStore = store
}

// ShutdownCh returns a channel that's closed when shutdown is complete.
func (c *Container) ShutdownCh() <-chan struct{} {
	return c.shutdownCh
}

// IsInitialized returns true if the container has been initialized.
func (c *Container) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}
