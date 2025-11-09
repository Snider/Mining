package mining

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Manager handles miner lifecycle and operations
type Manager struct {
	miners    map[string]Miner
	mu        sync.RWMutex // Mutex to protect the miners map
	stopChan  chan struct{}
	waitGroup sync.WaitGroup
}

// NewManager creates a new miner manager
func NewManager() *Manager {
	m := &Manager{
		miners:    make(map[string]Miner),
		stopChan:  make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
	m.startStatsCollection()
	return m
}

// StartMiner starts a new miner with the given configuration
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

// StopMiner stops a running miner
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

// GetMiner retrieves a miner by ID
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

// ListMiners returns all miners
func (m *Manager) ListMiners() []Miner {
	m.mu.RLock()
	defer m.mu.RUnlock()

	miners := make([]Miner, 0, len(m.miners))
	for _, miner := range m.miners {
		miners = append(miners, miner)
	}
	return miners
}

// ListAvailableMiners returns a list of available miners
func (m *Manager) ListAvailableMiners() []AvailableMiner {
	return []AvailableMiner{
		{
			Name:        "xmrig",
			Description: "XMRig is a high performance, open source, cross platform RandomX, KawPow, CryptoNight and AstroBWT CPU/GPU miner and RandomX benchmark.",
		},
	}
}

// startStatsCollection starts a goroutine to periodically collect stats from active miners
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

// collectMinerStats iterates through active miners and collects their stats
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

// GetMinerHashrateHistory returns the hashrate history for a specific miner
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

// Stop stops the manager and its background goroutines
func (m *Manager) Stop() {
	close(m.stopChan)
	m.waitGroup.Wait() // Wait for the stats collection goroutine to finish
}
