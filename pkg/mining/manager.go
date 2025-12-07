package mining

import (
	"fmt"
	"log"
	"net"
	"strconv"
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
func NewManager() *Manager {
	m := &Manager{
		miners:    make(map[string]Miner),
		stopChan:  make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
	m.startStatsCollection()
	return m
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

// StartMiner starts a new miner with the given configuration.
func (m *Manager) StartMiner(minerType string, config *Config) (Miner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Prevent nil pointer panic if request body is empty
	if config == nil {
		config = &Config{}
	}

	var miner Miner
	switch strings.ToLower(minerType) {
	case "xmrig":
		miner = NewXMRigMiner()
	default:
		return nil, fmt.Errorf("unsupported miner type: %s", minerType)
	}

	instanceName := miner.GetName()
	if config.Algo != "" {
		instanceName = fmt.Sprintf("%s-%s", instanceName, config.Algo)
	} else {
		instanceName = fmt.Sprintf("%s-%d", instanceName, time.Now().UnixNano()%1000)
	}

	if _, exists := m.miners[instanceName]; exists {
		return nil, fmt.Errorf("a miner with a similar configuration is already running: %s", instanceName)
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

	if err := miner.Start(config); err != nil {
		return nil, err
	}

	m.miners[instanceName] = miner

	logMessage := fmt.Sprintf("CryptoCurrency Miner started: %s (Binary: %s)", miner.GetName(), miner.GetBinaryPath())
	logToSyslog(logMessage)

	return miner, nil
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
	for _, miner := range m.miners {
		minersToCollect = append(minersToCollect, miner)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, miner := range minersToCollect {
		stats, err := miner.GetStats()
		if err != nil {
			log.Printf("Error getting stats for miner %s: %v\n", miner.GetName(), err)
			continue
		}
		miner.AddHashratePoint(HashratePoint{
			Timestamp: now,
			Hashrate:  stats.Hashrate,
		})
		miner.ReduceHashrateHistory(now)
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

// Stop stops the manager and its background goroutines.
func (m *Manager) Stop() {
	close(m.stopChan)
	m.waitGroup.Wait()
}

// Helper to convert port to string for net.JoinHostPort
func portToString(port int) string {
	return strconv.Itoa(port)
}
