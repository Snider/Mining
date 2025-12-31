package mining

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/database"
)

// sanitizeInstanceName ensures the instance name only contains safe characters.
var instanceNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_/-]`)

// ManagerInterface defines the contract for a miner manager.
type ManagerInterface interface {
	StartMiner(minerType string, config *Config) (Miner, error)
	StopMiner(name string) error
	GetMiner(name string) (Miner, error)
	ListMiners() []Miner
	ListAvailableMiners() []AvailableMiner
	GetMinerHashrateHistory(name string) ([]HashratePoint, error)
	UninstallMiner(minerType string) error
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

// initDatabase initializes the SQLite database based on config.
func (m *Manager) initDatabase() {
	cfg, err := LoadMinersConfig()
	if err != nil {
		log.Printf("Warning: could not load config for database init: %v", err)
		return
	}

	m.dbEnabled = cfg.Database.Enabled
	m.dbRetention = cfg.Database.RetentionDays
	if m.dbRetention == 0 {
		m.dbRetention = 30
	}

	if !m.dbEnabled {
		log.Println("Database persistence is disabled")
		return
	}

	dbCfg := database.Config{
		Enabled:       true,
		RetentionDays: m.dbRetention,
	}

	if err := database.Initialize(dbCfg); err != nil {
		log.Printf("Warning: failed to initialize database: %v", err)
		m.dbEnabled = false
		return
	}

	log.Printf("Database persistence enabled (retention: %d days)", m.dbRetention)

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
			log.Printf("Warning: database cleanup failed: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := database.Cleanup(m.dbRetention); err != nil {
					log.Printf("Warning: database cleanup failed: %v", err)
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
		log.Printf("Warning: could not load miners config for sync: %v", err)
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
			log.Printf("Added default config for missing miner: %s", availableMiner.Name)
		}
	}

	if configUpdated {
		if err := SaveMinersConfig(cfg); err != nil {
			log.Printf("Warning: failed to save updated miners config: %v", err)
		}
	}
}

// autostartMiners loads the miners config and starts any miners marked for autostart.
func (m *Manager) autostartMiners() {
	cfg, err := LoadMinersConfig()
	if err != nil {
		log.Printf("Warning: could not load miners config for autostart: %v", err)
		return
	}

	for _, minerCfg := range cfg.Miners {
		if minerCfg.Autostart && minerCfg.Config != nil {
			log.Printf("Autostarting miner: %s", minerCfg.MinerType)
			if _, err := m.StartMiner(minerCfg.MinerType, minerCfg.Config); err != nil {
				log.Printf("Failed to autostart miner %s: %v", minerCfg.MinerType, err)
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
func (m *Manager) StartMiner(minerType string, config *Config) (Miner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		config = &Config{}
	}

	var miner Miner
	switch strings.ToLower(minerType) {
	case "xmrig":
		miner = NewXMRigMiner()
	case "tt-miner", "ttminer":
		miner = NewTTMiner()
	default:
		return nil, fmt.Errorf("unsupported miner type: %s", minerType)
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

	if err := miner.Start(config); err != nil {
		return nil, err
	}

	m.miners[instanceName] = miner

	if err := m.updateMinerConfig(minerType, true, config); err != nil {
		log.Printf("Warning: failed to save miner config for autostart: %v", err)
	}

	logMessage := fmt.Sprintf("CryptoCurrency Miner started: %s (Binary: %s)", miner.GetName(), miner.GetBinaryPath())
	logToSyslog(logMessage)

	return miner, nil
}

// UninstallMiner stops, uninstalls, and removes a miner's configuration.
func (m *Manager) UninstallMiner(minerType string) error {
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
			log.Printf("Warning: failed to stop running miner %s during uninstall: %v", minersToDelete[i], err)
		}
	}

	var miner Miner
	switch strings.ToLower(minerType) {
	case "xmrig":
		miner = NewXMRigMiner()
	case "tt-miner", "ttminer":
		miner = NewTTMiner()
	default:
		return fmt.Errorf("unsupported miner type: %s", minerType)
	}

	if err := miner.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall miner files: %w", err)
	}

	cfg, err := LoadMinersConfig()
	if err != nil {
		return fmt.Errorf("failed to load miners config to update uninstall status: %w", err)
	}

	var updatedMiners []MinerAutostartConfig
	for _, minerCfg := range cfg.Miners {
		if !strings.EqualFold(minerCfg.MinerType, minerType) {
			updatedMiners = append(updatedMiners, minerCfg)
		}
	}
	cfg.Miners = updatedMiners

	return SaveMinersConfig(cfg)
}

// updateMinerConfig saves the autostart and last-used config for a miner.
func (m *Manager) updateMinerConfig(minerType string, autostart bool, config *Config) error {
	cfg, err := LoadMinersConfig()
	if err != nil {
		return err
	}

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

	return SaveMinersConfig(cfg)
}

// StopMiner stops a running miner.
func (m *Manager) StopMiner(name string) error {
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

	if err := miner.Stop(); err != nil {
		return err
	}

	delete(m.miners, name)
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

// collectMinerStats iterates through active miners and collects their stats.
func (m *Manager) collectMinerStats() {
	m.mu.RLock()
	minersToCollect := make([]Miner, 0, len(m.miners))
	minerTypes := make(map[string]string)
	for name, miner := range m.miners {
		minersToCollect = append(minersToCollect, miner)
		// Determine miner type from name prefix
		if strings.HasPrefix(name, "xmrig") {
			minerTypes[name] = "xmrig"
		} else if strings.HasPrefix(name, "tt-miner") || strings.HasPrefix(name, "ttminer") {
			minerTypes[name] = "tt-miner"
		} else {
			minerTypes[name] = "unknown"
		}
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, miner := range minersToCollect {
		minerName := miner.GetName()

		// Verify miner still exists before collecting stats
		m.mu.RLock()
		_, stillExists := m.miners[minerName]
		m.mu.RUnlock()
		if !stillExists {
			continue // Miner was removed, skip it
		}

		stats, err := miner.GetStats(context.Background())
		if err != nil {
			log.Printf("Error getting stats for miner %s: %v\n", minerName, err)
			continue
		}

		point := HashratePoint{
			Timestamp: now,
			Hashrate:  stats.Hashrate,
		}

		// Add to in-memory history (rolling window)
		miner.AddHashratePoint(point)
		miner.ReduceHashrateHistory(now)

		// Persist to database if enabled
		if m.dbEnabled {
			minerType := minerTypes[minerName]
			dbPoint := database.HashratePoint{
				Timestamp: point.Timestamp,
				Hashrate:  point.Hashrate,
			}
			if err := database.InsertHashratePoint(minerName, minerType, dbPoint, database.ResolutionHigh); err != nil {
				log.Printf("Warning: failed to persist hashrate for %s: %v", minerName, err)
			}
		}
	}
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

// Stop stops all running miners, background goroutines, and closes resources.
// Safe to call multiple times - subsequent calls are no-ops.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		// Stop all running miners first
		m.mu.Lock()
		for name, miner := range m.miners {
			if err := miner.Stop(); err != nil {
				log.Printf("Warning: failed to stop miner %s: %v", name, err)
			}
		}
		m.mu.Unlock()

		close(m.stopChan)
		m.waitGroup.Wait()

		// Close the database
		if m.dbEnabled {
			if err := database.Close(); err != nil {
				log.Printf("Warning: failed to close database: %v", err)
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
