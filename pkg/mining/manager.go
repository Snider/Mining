package mining

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Manager handles the lifecycle and operations of multiple miners.
// It provides a centralized way to start, stop, and manage different miner
// instances, while also collecting and exposing their performance data.
// The Manager is safe for concurrent use.
type Manager struct {
	miners    map[string]Miner
	mu        sync.RWMutex
	stopChan  chan struct{}
	waitGroup sync.WaitGroup
}

var _ ManagerInterface = (*Manager)(nil)

// NewManager creates a new miner manager.
// It initializes the manager and starts a background goroutine for periodic
// statistics collection from the miners.
//
// Example:
//
//	// Create a new manager
//	manager := mining.NewManager()
//	defer manager.Stop()
//
//	// Now you can use the manager to start and stop miners
func NewManager() *Manager {
	m := &Manager{
		miners:    make(map[string]Miner),
		stopChan:  make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
	m.startStatsCollection()
	return m
}

// StartMiner starts a new miner with the given configuration.
// It takes the miner type and a configuration object, and returns the
// created miner instance or an error if the miner could not be started.
//
// Example:
//
//	// Create a new manager
//	manager := mining.NewManager()
//	defer manager.Stop()
//
//	// Create a new configuration for the XMRig miner
//	config := &mining.Config{
//		Miner:   "xmrig",
//		Pool:    "your-pool-address",
//		Wallet:  "your-wallet-address",
//		Threads: 4,
//		TLS:     true,
//	}
//
//	// Start the miner
//	miner, err := manager.StartMiner("xmrig", config)
//	if err != nil {
//		log.Fatalf("Failed to start miner: %v", err)
//	}
//
//	// Stop the miner when you are done
//	defer manager.StopMiner(miner.GetName())
func (m *Manager) StartMiner(minerType string, config *Config) (Miner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var miner Miner
	switch strings.ToLower(minerType) {
	case "xmrig":
		miner = NewXMRigMiner()
	default:
		return nil, fmt.Errorf("unsupported miner type: %s", minerType)
	}

	// Ensure the miner's internal name is used for map key
	minerKey := miner.GetName()
	if _, exists := m.miners[minerKey]; exists {
		return nil, fmt.Errorf("miner already started: %s", minerKey)
	}

	if err := miner.Start(config); err != nil {
		return nil, err
	}

	m.miners[minerKey] = miner

	// Log to syslog (or standard log on Windows)
	logMessage := fmt.Sprintf("CryptoCurrency Miner started: %s (Binary: %s)", miner.GetName(), miner.GetBinaryPath())
	logToSyslog(logMessage)

	return miner, nil
}

// StopMiner stops a running miner.
// It takes the name of the miner to be stopped and returns an error if the
// miner could not be stopped.
func (m *Manager) StopMiner(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	minerKey := strings.ToLower(name) // Normalize input name to lowercase
	miner, exists := m.miners[minerKey]
	if !exists {
		return fmt.Errorf("miner not found: %s", name)
	}

	if err := miner.Stop(); err != nil {
		return err
	}

	delete(m.miners, minerKey)
	return nil
}

// GetMiner retrieves a running miner by its name.
// It returns the miner instance or an error if the miner is not found.
func (m *Manager) GetMiner(name string) (Miner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	minerKey := strings.ToLower(name) // Normalize input name to lowercase
	miner, exists := m.miners[minerKey]
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
// This provides a way to discover the types of miners supported by the manager.
func (m *Manager) ListAvailableMiners() []AvailableMiner {
	return []AvailableMiner{
		{
			Name:        "xmrig",
			Description: "XMRig is a high performance, open source, cross platform RandomX, KawPow, CryptoNight and AstroBWT CPU/GPU miner and RandomX benchmark.",
		},
	}
}

// startStatsCollection starts a goroutine to periodically collect stats from active miners.
func (m *Manager) startStatsCollection() {
	m.waitGroup.Add(1)
	go func() {
		defer m.waitGroup.Done()
		ticker := time.NewTicker(HighResolutionInterval) // Collect stats every 10 seconds
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
	for _, miner := range m.miners {
		minersToCollect = append(minersToCollect, miner)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, miner := range minersToCollect {
		stats, err := miner.GetStats()
		if err != nil {
			// Log the error but don't stop the collection for other miners
			log.Printf("Error getting stats for miner %s: %v\n", miner.GetName(), err)
			continue
		}
		miner.AddHashratePoint(HashratePoint{
			Timestamp: now,
			Hashrate:  stats.Hashrate,
		})
		miner.ReduceHashrateHistory(now) // Call the reducer
	}
}

// GetMinerHashrateHistory returns the hashrate history for a specific miner.
// It takes the name of the miner and returns a slice of hashrate points
// or an error if the miner is not found.
func (m *Manager) GetMinerHashrateHistory(name string) ([]HashratePoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	minerKey := strings.ToLower(name)
	miner, exists := m.miners[minerKey]
	if !exists {
		return nil, fmt.Errorf("miner not found: %s", name)
	}
	return miner.GetHashrateHistory(), nil
}

// Stop stops the manager and its background goroutines.
// It should be called when the manager is no longer needed to ensure a
// graceful shutdown of the statistics collection goroutine.
func (m *Manager) Stop() {
	close(m.stopChan)
	m.waitGroup.Wait() // Wait for the stats collection goroutine to finish
}
