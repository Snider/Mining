package mining

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/database"
	"github.com/Snider/Mining/pkg/logging"
)

// sanitizeInstanceName ensures the instance name only contains safe characters.
var instanceNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_/-]`)

// ManagerInterface defines the contract for a miner manager.
type ManagerInterface interface {
	StartMiner(ctx context.Context, minerType string, config *Config) (Miner, error)
	StopMiner(ctx context.Context, name string) error
	GetMiner(name string) (Miner, error)
	ListMiners() []Miner
	ListAvailableMiners() []AvailableMiner
	GetMinerHashrateHistory(name string) ([]HashratePoint, error)
	UninstallMiner(ctx context.Context, minerType string) error
	Stop()
}

// Manager handles the lifecycle and operations of multiple miners.
type Manager struct {
	miners      map[string]Miner
	mu          sync.RWMutex
	stopChan    chan struct{}
	stopOnce    sync.Once
	waitGroup   sync.WaitGroup
	dbEnabled   bool
	dbRetention int
	eventHub    *EventHub
	eventHubMu  sync.RWMutex // Separate mutex for eventHub to avoid deadlock with main mu
}

// SetEventHub sets the event hub for broadcasting miner events
func (m *Manager) SetEventHub(hub *EventHub) {
	m.eventHubMu.Lock()
	defer m.eventHubMu.Unlock()
	m.eventHub = hub
}

// emitEvent broadcasts an event if an event hub is configured
// Uses separate eventHubMu to avoid deadlock when called while holding m.mu
func (m *Manager) emitEvent(eventType EventType, data interface{}) {
	m.eventHubMu.RLock()
	hub := m.eventHub
	m.eventHubMu.RUnlock()

	if hub != nil {
		hub.Broadcast(NewEvent(eventType, data))
	}
}

var _ ManagerInterface = (*Manager)(nil)

// NewManager creates a new miner manager and autostarts miners based on config.
func NewManager() *Manager {
	m := &Manager{
		miners:    make(map[string]Miner),
		stopChan:  make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
	m.syncMinersConfig() // Ensure config file is populated
	m.initDatabase()
	m.autostartMiners()
	m.startStatsCollection()
	return m
}

// NewManagerForSimulation creates a manager for simulation mode.
// It skips autostarting real miners and config sync, suitable for UI testing.
func NewManagerForSimulation() *Manager {
	m := &Manager{
		miners:    make(map[string]Miner),
		stopChan:  make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
	// Skip syncMinersConfig and autostartMiners for simulation
	m.startStatsCollection()
	return m
}

// initDatabase initializes the SQLite database based on config.
func (m *Manager) initDatabase() {
	cfg, err := LoadMinersConfig()
	if err != nil {
		logging.Warn("could not load config for database init", logging.Fields{"error": err})
		return
	}

	m.dbEnabled = cfg.Database.Enabled
	m.dbRetention = cfg.Database.RetentionDays
	if m.dbRetention == 0 {
		m.dbRetention = 30
	}

	if !m.dbEnabled {
		logging.Debug("database persistence is disabled")
		return
	}

	dbCfg := database.Config{
		Enabled:       true,
		RetentionDays: m.dbRetention,
	}

	if err := database.Initialize(dbCfg); err != nil {
		logging.Warn("failed to initialize database", logging.Fields{"error": err})
		m.dbEnabled = false
		return
	}

	logging.Info("database persistence enabled", logging.Fields{"retention_days": m.dbRetention})

	// Start periodic cleanup
	m.startDBCleanup()
}

// startDBCleanup starts a goroutine that periodically cleans old data.
func (m *Manager) startDBCleanup() {
	m.waitGroup.Add(1)
	go func() {
		defer m.waitGroup.Done()
		// Run cleanup once per hour
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		// Run initial cleanup
		if err := database.Cleanup(m.dbRetention); err != nil {
			logging.Warn("database cleanup failed", logging.Fields{"error": err})
		}

		for {
			select {
			case <-ticker.C:
				if err := database.Cleanup(m.dbRetention); err != nil {
					logging.Warn("database cleanup failed", logging.Fields{"error": err})
				}
			case <-m.stopChan:
				return
			}
		}
	}()
}

// syncMinersConfig ensures the miners.json config file has entries for all available miners.
func (m *Manager) syncMinersConfig() {
	cfg, err := LoadMinersConfig()
	if err != nil {
		logging.Warn("could not load miners config for sync", logging.Fields{"error": err})
		return
	}

	availableMiners := m.ListAvailableMiners()
	configUpdated := false

	for _, availableMiner := range availableMiners {
		found := false
		for _, configuredMiner := range cfg.Miners {
			if strings.EqualFold(configuredMiner.MinerType, availableMiner.Name) {
				found = true
				break
			}
		}
		if !found {
			cfg.Miners = append(cfg.Miners, MinerAutostartConfig{
				MinerType: availableMiner.Name,
				Autostart: false,
				Config:    nil, // No default config
			})
			configUpdated = true
			logging.Info("added default config for missing miner", logging.Fields{"miner": availableMiner.Name})
		}
	}

	if configUpdated {
		if err := SaveMinersConfig(cfg); err != nil {
			logging.Warn("failed to save updated miners config", logging.Fields{"error": err})
		}
	}
}

// autostartMiners loads the miners config and starts any miners marked for autostart.
func (m *Manager) autostartMiners() {
	cfg, err := LoadMinersConfig()
	if err != nil {
		logging.Warn("could not load miners config for autostart", logging.Fields{"error": err})
		return
	}

	for _, minerCfg := range cfg.Miners {
		if minerCfg.Autostart && minerCfg.Config != nil {
			logging.Info("autostarting miner", logging.Fields{"type": minerCfg.MinerType})
			if _, err := m.StartMiner(context.Background(), minerCfg.MinerType, minerCfg.Config); err != nil {
				logging.Error("failed to autostart miner", logging.Fields{"type": minerCfg.MinerType, "error": err})
			}
		}
	}
}

// findAvailablePort finds an available TCP port on the local machine.
func findAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// StartMiner starts a new miner and saves its configuration.
// The context can be used to cancel the operation.
func (m *Manager) StartMiner(ctx context.Context, minerType string, config *Config) (Miner, error) {
	// Check for cancellation before acquiring lock
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		config = &Config{}
	}

	miner, err := CreateMiner(minerType)
	if err != nil {
		return nil, err
	}

	instanceName := miner.GetName()
	if config.Algo != "" {
		// Sanitize algo to prevent directory traversal or invalid filenames
		sanitizedAlgo := instanceNameRegex.ReplaceAllString(config.Algo, "_")
		instanceName = fmt.Sprintf("%s-%s", instanceName, sanitizedAlgo)
	} else {
		instanceName = fmt.Sprintf("%s-%d", instanceName, time.Now().UnixNano()%1000)
	}

	if _, exists := m.miners[instanceName]; exists {
		return nil, fmt.Errorf("a miner with a similar configuration is already running: %s", instanceName)
	}

	// Validate user-provided HTTPPort if specified
	if config.HTTPPort != 0 {
		if config.HTTPPort < 1024 || config.HTTPPort > 65535 {
			return nil, fmt.Errorf("HTTPPort must be between 1024 and 65535, got %d", config.HTTPPort)
		}
	}

	apiPort, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find an available port for the miner API: %w", err)
	}
	if config.HTTPPort == 0 {
		config.HTTPPort = apiPort
	}

	if xmrigMiner, ok := miner.(*XMRigMiner); ok {
		xmrigMiner.Name = instanceName
		if xmrigMiner.API != nil {
			xmrigMiner.API.ListenPort = apiPort
		}
	}
	if ttMiner, ok := miner.(*TTMiner); ok {
		ttMiner.Name = instanceName
		if ttMiner.API != nil {
			ttMiner.API.ListenPort = apiPort
		}
	}

	// Emit starting event before actually starting
	m.emitEvent(EventMinerStarting, MinerEventData{
		Name: instanceName,
	})

	if err := miner.Start(config); err != nil {
		// Emit error event
		m.emitEvent(EventMinerError, MinerEventData{
			Name:  instanceName,
			Error: err.Error(),
		})
		return nil, err
	}

	m.miners[instanceName] = miner

	if err := m.updateMinerConfig(minerType, true, config); err != nil {
		logging.Warn("failed to save miner config for autostart", logging.Fields{"error": err})
	}

	logMessage := fmt.Sprintf("CryptoCurrency Miner started: %s (Binary: %s)", miner.GetName(), miner.GetBinaryPath())
	logToSyslog(logMessage)

	// Emit started event
	m.emitEvent(EventMinerStarted, MinerEventData{
		Name: instanceName,
	})

	return miner, nil
}

// UninstallMiner stops, uninstalls, and removes a miner's configuration.
// The context can be used to cancel the operation.
func (m *Manager) UninstallMiner(ctx context.Context, minerType string) error {
	// Check for cancellation before acquiring lock
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	// Collect miners to stop and delete (can't modify map during iteration)
	minersToDelete := make([]string, 0)
	minersToStop := make([]Miner, 0)
	for name, runningMiner := range m.miners {
		if rm, ok := runningMiner.(*XMRigMiner); ok && strings.EqualFold(rm.ExecutableName, minerType) {
			minersToStop = append(minersToStop, runningMiner)
			minersToDelete = append(minersToDelete, name)
		}
		if rm, ok := runningMiner.(*TTMiner); ok && strings.EqualFold(rm.ExecutableName, minerType) {
			minersToStop = append(minersToStop, runningMiner)
			minersToDelete = append(minersToDelete, name)
		}
	}
	// Delete from map first, then release lock before stopping (Stop may block)
	for _, name := range minersToDelete {
		delete(m.miners, name)
	}
	m.mu.Unlock()

	// Stop miners outside the lock to avoid blocking
	for i, miner := range minersToStop {
		if err := miner.Stop(); err != nil {
			logging.Warn("failed to stop running miner during uninstall", logging.Fields{"miner": minersToDelete[i], "error": err})
		}
	}

	miner, err := CreateMiner(minerType)
	if err != nil {
		return err
	}

	if err := miner.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall miner files: %w", err)
	}

	return UpdateMinersConfig(func(cfg *MinersConfig) error {
		var updatedMiners []MinerAutostartConfig
		for _, minerCfg := range cfg.Miners {
			if !strings.EqualFold(minerCfg.MinerType, minerType) {
				updatedMiners = append(updatedMiners, minerCfg)
			}
		}
		cfg.Miners = updatedMiners
		return nil
	})
}

// updateMinerConfig saves the autostart and last-used config for a miner.
func (m *Manager) updateMinerConfig(minerType string, autostart bool, config *Config) error {
	return UpdateMinersConfig(func(cfg *MinersConfig) error {
		found := false
		for i, minerCfg := range cfg.Miners {
			if strings.EqualFold(minerCfg.MinerType, minerType) {
				cfg.Miners[i].Autostart = autostart
				cfg.Miners[i].Config = config
				found = true
				break
			}
		}

		if !found {
			cfg.Miners = append(cfg.Miners, MinerAutostartConfig{
				MinerType: minerType,
				Autostart: autostart,
				Config:    config,
			})
		}
		return nil
	})
}

// StopMiner stops a running miner and removes it from the manager.
// If the miner is already stopped, it will still be removed from the manager.
// The context can be used to cancel the operation.
func (m *Manager) StopMiner(ctx context.Context, name string) error {
	// Check for cancellation before acquiring lock
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	miner, exists := m.miners[name]
	if !exists {
		for k := range m.miners {
			if strings.HasPrefix(k, name) {
				miner = m.miners[k]
				name = k
				exists = true
				break
			}
		}
	}

	if !exists {
		return fmt.Errorf("miner not found: %s", name)
	}

	// Emit stopping event
	m.emitEvent(EventMinerStopping, MinerEventData{
		Name: name,
	})

	// Try to stop the miner, but always remove it from the map
	// This handles the case where a miner crashed or was killed externally
	stopErr := miner.Stop()

	// Always remove from map - if it's not running, we still want to clean it up
	delete(m.miners, name)

	// Emit stopped event
	reason := "stopped"
	if stopErr != nil && stopErr.Error() != "miner is not running" {
		reason = stopErr.Error()
	}
	m.emitEvent(EventMinerStopped, MinerEventData{
		Name:   name,
		Reason: reason,
	})

	// Only return error if it wasn't just "miner is not running"
	if stopErr != nil && stopErr.Error() != "miner is not running" {
		return stopErr
	}

	return nil
}

// GetMiner retrieves a running miner by its name.
func (m *Manager) GetMiner(name string) (Miner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	miner, exists := m.miners[name]
	if !exists {
		return nil, fmt.Errorf("miner not found: %s", name)
	}
	return miner, nil
}

// ListMiners returns a slice of all running miners.
func (m *Manager) ListMiners() []Miner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	miners := make([]Miner, 0, len(m.miners))
	for _, miner := range m.miners {
		miners = append(miners, miner)
	}
	return miners
}

// RegisterMiner registers an already-started miner with the manager.
// This is useful for simulated miners or externally managed miners.
func (m *Manager) RegisterMiner(miner Miner) error {
	name := miner.GetName()

	m.mu.Lock()
	if _, exists := m.miners[name]; exists {
		m.mu.Unlock()
		return fmt.Errorf("miner %s is already registered", name)
	}
	m.miners[name] = miner
	m.mu.Unlock()

	logging.Info("registered miner", logging.Fields{"name": name})

	// Emit miner started event (outside lock)
	m.emitEvent(EventMinerStarted, map[string]interface{}{
		"name": name,
	})

	return nil
}

// ListAvailableMiners returns a list of available miners that can be started.
func (m *Manager) ListAvailableMiners() []AvailableMiner {
	return []AvailableMiner{
		{
			Name:        "xmrig",
			Description: "XMRig is a high performance, open source, cross platform RandomX, KawPow, CryptoNight and AstroBWT CPU/GPU miner and RandomX benchmark.",
		},
		{
			Name:        "tt-miner",
			Description: "TT-Miner is a high performance NVIDIA GPU miner for various algorithms including Ethash, KawPow, ProgPow, and more. Requires CUDA.",
		},
	}
}

// startStatsCollection starts a goroutine to periodically collect stats from active miners.
func (m *Manager) startStatsCollection() {
	m.waitGroup.Add(1)
	go func() {
		defer m.waitGroup.Done()
		ticker := time.NewTicker(HighResolutionInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.collectMinerStats()
			case <-m.stopChan:
				return
			}
		}
	}()
}

// statsCollectionTimeout is the maximum time to wait for stats from a single miner.
const statsCollectionTimeout = 5 * time.Second

// collectMinerStats iterates through active miners and collects their stats.
// Stats are collected in parallel to reduce overall collection time.
func (m *Manager) collectMinerStats() {
	// Take a snapshot of miners under read lock - minimize lock duration
	m.mu.RLock()
	if len(m.miners) == 0 {
		m.mu.RUnlock()
		return
	}

	type minerInfo struct {
		miner     Miner
		minerType string
	}
	miners := make([]minerInfo, 0, len(m.miners))
	for _, miner := range m.miners {
		// Use the miner's GetType() method for proper type identification
		miners = append(miners, minerInfo{miner: miner, minerType: miner.GetType()})
	}
	dbEnabled := m.dbEnabled // Copy to avoid holding lock
	m.mu.RUnlock()

	now := time.Now()

	// Collect stats from all miners in parallel
	var wg sync.WaitGroup
	for _, mi := range miners {
		wg.Add(1)
		go func(miner Miner, minerType string) {
			defer wg.Done()
			m.collectSingleMinerStats(miner, minerType, now, dbEnabled)
		}(mi.miner, mi.minerType)
	}
	wg.Wait()
}

// collectSingleMinerStats collects stats from a single miner.
// This is called concurrently for each miner.
func (m *Manager) collectSingleMinerStats(miner Miner, minerType string, now time.Time, dbEnabled bool) {
	minerName := miner.GetName()

	// Use context with timeout to prevent hanging on unresponsive miner APIs
	ctx, cancel := context.WithTimeout(context.Background(), statsCollectionTimeout)
	stats, err := miner.GetStats(ctx)
	cancel() // Release context resources immediately

	if err != nil {
		logging.Error("failed to get miner stats", logging.Fields{"miner": minerName, "error": err})
		return
	}

	point := HashratePoint{
		Timestamp: now,
		Hashrate:  stats.Hashrate,
	}

	// Add to in-memory history (rolling window)
	// Note: AddHashratePoint and ReduceHashrateHistory must be thread-safe
	miner.AddHashratePoint(point)
	miner.ReduceHashrateHistory(now)

	// Persist to database if enabled
	if dbEnabled {
		dbPoint := database.HashratePoint{
			Timestamp: point.Timestamp,
			Hashrate:  point.Hashrate,
		}
		// Use nil context to let InsertHashratePoint use its default timeout
		if err := database.InsertHashratePoint(nil, minerName, minerType, dbPoint, database.ResolutionHigh); err != nil {
			logging.Warn("failed to persist hashrate", logging.Fields{"miner": minerName, "error": err})
		}
	}

	// Emit stats event for real-time WebSocket updates
	m.emitEvent(EventMinerStats, MinerStatsData{
		Name:        minerName,
		Hashrate:    stats.Hashrate,
		Shares:      stats.Shares,
		Rejected:    stats.Rejected,
		Uptime:      stats.Uptime,
		Algorithm:   stats.Algorithm,
		DiffCurrent: stats.DiffCurrent,
	})
}

// GetMinerHashrateHistory returns the hashrate history for a specific miner.
func (m *Manager) GetMinerHashrateHistory(name string) ([]HashratePoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	miner, exists := m.miners[name]
	if !exists {
		return nil, fmt.Errorf("miner not found: %s", name)
	}
	return miner.GetHashrateHistory(), nil
}

// ShutdownTimeout is the maximum time to wait for goroutines during shutdown
const ShutdownTimeout = 10 * time.Second

// Stop stops all running miners, background goroutines, and closes resources.
// Safe to call multiple times - subsequent calls are no-ops.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		// Stop all running miners first
		m.mu.Lock()
		for name, miner := range m.miners {
			if err := miner.Stop(); err != nil {
				logging.Warn("failed to stop miner", logging.Fields{"miner": name, "error": err})
			}
		}
		m.mu.Unlock()

		close(m.stopChan)

		// Wait for goroutines with timeout
		done := make(chan struct{})
		go func() {
			m.waitGroup.Wait()
			close(done)
		}()

		select {
		case <-done:
			logging.Info("all goroutines stopped gracefully")
		case <-time.After(ShutdownTimeout):
			logging.Warn("shutdown timeout - some goroutines may not have stopped")
		}

		// Close the database
		if m.dbEnabled {
			if err := database.Close(); err != nil {
				logging.Warn("failed to close database", logging.Fields{"error": err})
			}
		}
	})
}

// GetMinerHistoricalStats returns historical stats from the database for a miner.
func (m *Manager) GetMinerHistoricalStats(minerName string) (*database.HashrateStats, error) {
	if !m.dbEnabled {
		return nil, fmt.Errorf("database persistence is disabled")
	}
	return database.GetHashrateStats(minerName)
}

// GetMinerHistoricalHashrate returns historical hashrate data from the database.
func (m *Manager) GetMinerHistoricalHashrate(minerName string, since, until time.Time) ([]HashratePoint, error) {
	if !m.dbEnabled {
		return nil, fmt.Errorf("database persistence is disabled")
	}

	dbPoints, err := database.GetHashrateHistory(minerName, database.ResolutionHigh, since, until)
	if err != nil {
		return nil, err
	}

	// Convert database points to mining points
	points := make([]HashratePoint, len(dbPoints))
	for i, p := range dbPoints {
		points[i] = HashratePoint{
			Timestamp: p.Timestamp,
			Hashrate:  p.Hashrate,
		}
	}
	return points, nil
}

// GetAllMinerHistoricalStats returns historical stats for all miners from the database.
func (m *Manager) GetAllMinerHistoricalStats() ([]database.HashrateStats, error) {
	if !m.dbEnabled {
		return nil, fmt.Errorf("database persistence is disabled")
	}
	return database.GetAllMinerStats()
}

// IsDatabaseEnabled returns whether database persistence is enabled.
func (m *Manager) IsDatabaseEnabled() bool {
	return m.dbEnabled
}

// Helper to convert port to string for net.JoinHostPort
func portToString(port int) string {
	return strconv.Itoa(port)
}
