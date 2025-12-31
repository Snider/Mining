package mining

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SimulatedMiner is a mock miner that generates realistic-looking stats for UI testing.
type SimulatedMiner struct {
	// Exported fields for JSON serialization
	Name             string           `json:"name"`
	Version          string           `json:"version"`
	URL              string           `json:"url"`
	Path             string           `json:"path"`
	MinerBinary      string           `json:"miner_binary"`
	Running          bool             `json:"running"`
	Algorithm        string           `json:"algorithm"`
	HashrateHistory  []HashratePoint  `json:"hashrateHistory"`
	LowResHistory    []HashratePoint  `json:"lowResHashrateHistory"`
	Stats            *PerformanceMetrics `json:"stats,omitempty"`
	FullStats        *XMRigSummary       `json:"full_stats,omitempty"` // XMRig-compatible format for UI

	// Internal fields (not exported)
	baseHashrate     int
	peakHashrate     int
	variance         float64
	startTime        time.Time
	shares           int
	rejected         int
	logs             []string
	mu               sync.RWMutex
	stopChan         chan struct{}
	poolName         string
	difficultyBase   int
}

// SimulatedMinerConfig holds configuration for creating a simulated miner.
type SimulatedMinerConfig struct {
	Name         string  // Miner instance name (e.g., "sim-xmrig-001")
	Algorithm    string  // Algorithm name (e.g., "rx/0", "kawpow", "ethash")
	BaseHashrate int     // Base hashrate in H/s
	Variance     float64 // Variance as percentage (0.0-0.2 for 20% variance)
	PoolName     string  // Simulated pool name
	Difficulty   int     // Base difficulty
}

// NewSimulatedMiner creates a new simulated miner instance.
func NewSimulatedMiner(config SimulatedMinerConfig) *SimulatedMiner {
	if config.Variance <= 0 {
		config.Variance = 0.1 // Default 10% variance
	}
	if config.PoolName == "" {
		config.PoolName = "sim-pool.example.com:3333"
	}
	if config.Difficulty <= 0 {
		config.Difficulty = 10000
	}

	return &SimulatedMiner{
		Name:            config.Name,
		Version:         "1.0.0-simulated",
		URL:             "https://github.com/simulated/miner",
		Path:            "/simulated/miner",
		MinerBinary:     "/simulated/miner/sim-miner",
		Algorithm:       config.Algorithm,
		HashrateHistory: make([]HashratePoint, 0),
		LowResHistory:   make([]HashratePoint, 0),
		baseHashrate:    config.BaseHashrate,
		variance:        config.Variance,
		poolName:        config.PoolName,
		difficultyBase:  config.Difficulty,
		logs:            make([]string, 0),
	}
}

// Install is a no-op for simulated miners.
func (m *SimulatedMiner) Install() error {
	return nil
}

// Uninstall is a no-op for simulated miners.
func (m *SimulatedMiner) Uninstall() error {
	return nil
}

// Start begins the simulated mining process.
func (m *SimulatedMiner) Start(config *Config) error {
	m.mu.Lock()
	if m.Running {
		m.mu.Unlock()
		return fmt.Errorf("simulated miner %s is already running", m.Name)
	}

	m.Running = true
	m.startTime = time.Now()
	m.shares = 0
	m.rejected = 0
	m.stopChan = make(chan struct{})
	m.HashrateHistory = make([]HashratePoint, 0)
	m.LowResHistory = make([]HashratePoint, 0)
	m.logs = []string{
		fmt.Sprintf("[%s] Simulated miner starting...", time.Now().Format("15:04:05")),
		fmt.Sprintf("[%s] Connecting to %s", time.Now().Format("15:04:05"), m.poolName),
		fmt.Sprintf("[%s] Pool connected, algorithm: %s", time.Now().Format("15:04:05"), m.Algorithm),
	}
	m.mu.Unlock()

	// Start background simulation
	go m.runSimulation()

	return nil
}

// Stop stops the simulated miner.
func (m *SimulatedMiner) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running {
		return fmt.Errorf("simulated miner %s is not running", m.Name)
	}

	close(m.stopChan)
	m.Running = false
	m.logs = append(m.logs, fmt.Sprintf("[%s] Miner stopped", time.Now().Format("15:04:05")))

	return nil
}

// runSimulation runs the background simulation loop.
func (m *SimulatedMiner) runSimulation() {
	ticker := time.NewTicker(HighResolutionInterval)
	defer ticker.Stop()

	shareTicker := time.NewTicker(time.Duration(5+rand.Intn(10)) * time.Second)
	defer shareTicker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.updateHashrate()
		case <-shareTicker.C:
			m.simulateShare()
			// Randomize next share time
			shareTicker.Reset(time.Duration(5+rand.Intn(15)) * time.Second)
		}
	}
}

// updateHashrate generates a new hashrate value with realistic variation.
func (m *SimulatedMiner) updateHashrate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate hashrate with variance and smooth transitions
	now := time.Now()
	uptime := now.Sub(m.startTime).Seconds()

	// Ramp up period (first 30 seconds)
	rampFactor := math.Min(1.0, uptime/30.0)

	// Add some sine wave variation for realistic fluctuation
	sineVariation := math.Sin(uptime/10) * 0.05

	// Random noise
	noise := (rand.Float64() - 0.5) * 2 * m.variance

	// Calculate final hashrate
	hashrate := int(float64(m.baseHashrate) * rampFactor * (1.0 + sineVariation + noise))
	if hashrate < 0 {
		hashrate = 0
	}

	point := HashratePoint{
		Timestamp: now,
		Hashrate:  hashrate,
	}

	m.HashrateHistory = append(m.HashrateHistory, point)

	// Track peak hashrate
	if hashrate > m.peakHashrate {
		m.peakHashrate = hashrate
	}

	// Update stats for JSON serialization
	uptimeInt := int(uptime)
	diffCurrent := m.difficultyBase + rand.Intn(m.difficultyBase/2)

	m.Stats = &PerformanceMetrics{
		Hashrate:      hashrate,
		Shares:        m.shares,
		Rejected:      m.rejected,
		Uptime:        uptimeInt,
		Algorithm:     m.Algorithm,
		AvgDifficulty: m.difficultyBase,
		DiffCurrent:   diffCurrent,
	}

	// Update XMRig-compatible full_stats for UI
	m.FullStats = &XMRigSummary{
		ID:       m.Name,
		WorkerID: m.Name,
		Uptime:   uptimeInt,
		Algo:     m.Algorithm,
		Version:  m.Version,
	}
	m.FullStats.Hashrate.Total = []float64{float64(hashrate)}
	m.FullStats.Hashrate.Highest = float64(m.peakHashrate)
	m.FullStats.Results.SharesGood = m.shares
	m.FullStats.Results.SharesTotal = m.shares + m.rejected
	m.FullStats.Results.DiffCurrent = diffCurrent
	m.FullStats.Results.AvgTime = 15 + rand.Intn(10) // Simulated avg share time
	m.FullStats.Results.HashesTotal = m.shares * diffCurrent
	m.FullStats.Connection.Pool = m.poolName
	m.FullStats.Connection.Uptime = uptimeInt
	m.FullStats.Connection.Diff = diffCurrent
	m.FullStats.Connection.Accepted = m.shares
	m.FullStats.Connection.Rejected = m.rejected
	m.FullStats.Connection.Algo = m.Algorithm
	m.FullStats.Connection.Ping = 50 + rand.Intn(50)

	// Trim high-res history to last 5 minutes
	cutoff := now.Add(-HighResolutionDuration)
	for len(m.HashrateHistory) > 0 && m.HashrateHistory[0].Timestamp.Before(cutoff) {
		m.HashrateHistory = m.HashrateHistory[1:]
	}
}

// simulateShare simulates finding a share.
func (m *SimulatedMiner) simulateShare() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 2% chance of rejected share
	if rand.Float64() < 0.02 {
		m.rejected++
		m.logs = append(m.logs, fmt.Sprintf("[%s] Share rejected (stale)", time.Now().Format("15:04:05")))
	} else {
		m.shares++
		diff := m.difficultyBase + rand.Intn(m.difficultyBase/2)
		m.logs = append(m.logs, fmt.Sprintf("[%s] Share accepted (%d/%d) diff %d", time.Now().Format("15:04:05"), m.shares, m.rejected, diff))
	}

	// Keep last 100 log lines
	if len(m.logs) > 100 {
		m.logs = m.logs[len(m.logs)-100:]
	}
}

// GetStats returns current performance metrics.
func (m *SimulatedMiner) GetStats(ctx context.Context) (*PerformanceMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.Running {
		return nil, fmt.Errorf("simulated miner %s is not running", m.Name)
	}

	// Calculate current hashrate from recent history
	var hashrate int
	if len(m.HashrateHistory) > 0 {
		hashrate = m.HashrateHistory[len(m.HashrateHistory)-1].Hashrate
	}

	uptime := int(time.Since(m.startTime).Seconds())

	// Calculate average difficulty
	avgDiff := m.difficultyBase
	if m.shares > 0 {
		avgDiff = m.difficultyBase + rand.Intn(m.difficultyBase/4)
	}

	return &PerformanceMetrics{
		Hashrate:      hashrate,
		Shares:        m.shares,
		Rejected:      m.rejected,
		Uptime:        uptime,
		LastShare:     time.Now().Unix() - int64(rand.Intn(30)),
		Algorithm:     m.Algorithm,
		AvgDifficulty: avgDiff,
		DiffCurrent:   m.difficultyBase + rand.Intn(m.difficultyBase/2),
		ExtraData: map[string]interface{}{
			"pool":      m.poolName,
			"simulated": true,
		},
	}, nil
}

// GetName returns the miner's name.
func (m *SimulatedMiner) GetName() string {
	return m.Name
}

// GetPath returns a simulated path.
func (m *SimulatedMiner) GetPath() string {
	return m.Path
}

// GetBinaryPath returns a simulated binary path.
func (m *SimulatedMiner) GetBinaryPath() string {
	return m.MinerBinary
}

// CheckInstallation returns simulated installation details.
func (m *SimulatedMiner) CheckInstallation() (*InstallationDetails, error) {
	return &InstallationDetails{
		IsInstalled: true,
		Version:     "1.0.0-simulated",
		Path:        "/simulated/miner",
		MinerBinary: "simulated-miner",
		ConfigPath:  "/simulated/config.json",
	}, nil
}

// GetLatestVersion returns a simulated version.
func (m *SimulatedMiner) GetLatestVersion() (string, error) {
	return "1.0.0-simulated", nil
}

// GetHashrateHistory returns the hashrate history.
func (m *SimulatedMiner) GetHashrateHistory() []HashratePoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]HashratePoint, len(m.HashrateHistory))
	copy(result, m.HashrateHistory)
	return result
}

// AddHashratePoint adds a point to the history.
func (m *SimulatedMiner) AddHashratePoint(point HashratePoint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HashrateHistory = append(m.HashrateHistory, point)
}

// ReduceHashrateHistory reduces the history (called by manager).
func (m *SimulatedMiner) ReduceHashrateHistory(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Move old high-res points to low-res
	cutoff := now.Add(-HighResolutionDuration)
	var toMove []HashratePoint

	newHistory := make([]HashratePoint, 0)
	for _, point := range m.HashrateHistory {
		if point.Timestamp.Before(cutoff) {
			toMove = append(toMove, point)
		} else {
			newHistory = append(newHistory, point)
		}
	}
	m.HashrateHistory = newHistory

	// Average the old points and add to low-res
	if len(toMove) > 0 {
		var sum int
		for _, p := range toMove {
			sum += p.Hashrate
		}
		avg := sum / len(toMove)
		m.LowResHistory = append(m.LowResHistory, HashratePoint{
			Timestamp: toMove[len(toMove)-1].Timestamp,
			Hashrate:  avg,
		})
	}

	// Trim low-res history
	lowResCutoff := now.Add(-LowResHistoryRetention)
	newLowRes := make([]HashratePoint, 0)
	for _, point := range m.LowResHistory {
		if !point.Timestamp.Before(lowResCutoff) {
			newLowRes = append(newLowRes, point)
		}
	}
	m.LowResHistory = newLowRes
}

// GetLogs returns the simulated logs.
func (m *SimulatedMiner) GetLogs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, len(m.logs))
	copy(result, m.logs)
	return result
}

// WriteStdin simulates stdin input.
func (m *SimulatedMiner) WriteStdin(input string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running {
		return fmt.Errorf("simulated miner %s is not running", m.Name)
	}

	m.logs = append(m.logs, fmt.Sprintf("[%s] stdin: %s", time.Now().Format("15:04:05"), input))
	return nil
}

// SimulatedMinerPresets provides common presets for simulated miners.
var SimulatedMinerPresets = map[string]SimulatedMinerConfig{
	"cpu-low": {
		Algorithm:    "rx/0",
		BaseHashrate: 500,
		Variance:     0.15,
		PoolName:     "pool.hashvault.pro:443",
		Difficulty:   50000,
	},
	"cpu-medium": {
		Algorithm:    "rx/0",
		BaseHashrate: 5000,
		Variance:     0.10,
		PoolName:     "pool.hashvault.pro:443",
		Difficulty:   100000,
	},
	"cpu-high": {
		Algorithm:    "rx/0",
		BaseHashrate: 15000,
		Variance:     0.08,
		PoolName:     "pool.hashvault.pro:443",
		Difficulty:   200000,
	},
	"gpu-ethash": {
		Algorithm:    "ethash",
		BaseHashrate: 30000000, // 30 MH/s
		Variance:     0.05,
		PoolName:     "eth.2miners.com:2020",
		Difficulty:   4000000000,
	},
	"gpu-kawpow": {
		Algorithm:    "kawpow",
		BaseHashrate: 15000000, // 15 MH/s
		Variance:     0.06,
		PoolName:     "rvn.2miners.com:6060",
		Difficulty:   1000000000,
	},
}
