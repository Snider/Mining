// Package mining provides core functionality for miner management
package mining

import (
	"fmt"
	"time"
)

// Miner represents a mining instance
type Miner struct {
	ID        string
	Name      string
	Status    string
	StartTime time.Time
	HashRate  float64
}

// MinerConfig holds configuration for a miner
type MinerConfig struct {
	Name      string
	Algorithm string
	Pool      string
	Wallet    string
}

// Manager handles miner lifecycle and operations
type Manager struct {
	miners map[string]*Miner
}

// NewManager creates a new miner manager
func NewManager() *Manager {
	return &Manager{
		miners: make(map[string]*Miner),
	}
}

// StartMiner starts a new miner with the given configuration
func (m *Manager) StartMiner(config MinerConfig) (*Miner, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("miner name is required")
	}

	miner := &Miner{
		ID:        generateID(),
		Name:      config.Name,
		Status:    "running",
		StartTime: time.Now(),
		HashRate:  0.0,
	}

	m.miners[miner.ID] = miner
	return miner, nil
}

// StopMiner stops a running miner
func (m *Manager) StopMiner(id string) error {
	miner, exists := m.miners[id]
	if !exists {
		return fmt.Errorf("miner not found: %s", id)
	}

	miner.Status = "stopped"
	return nil
}

// GetMiner retrieves a miner by ID
func (m *Manager) GetMiner(id string) (*Miner, error) {
	miner, exists := m.miners[id]
	if !exists {
		return nil, fmt.Errorf("miner not found: %s", id)
	}
	return miner, nil
}

// ListMiners returns all miners
func (m *Manager) ListMiners() []*Miner {
	miners := make([]*Miner, 0, len(m.miners))
	for _, miner := range m.miners {
		miners = append(miners, miner)
	}
	return miners
}

// UpdateHashRate updates the hash rate for a miner
func (m *Manager) UpdateHashRate(id string, hashRate float64) error {
	miner, exists := m.miners[id]
	if !exists {
		return fmt.Errorf("miner not found: %s", id)
	}

	miner.HashRate = hashRate
	return nil
}

// generateID generates a unique ID for a miner
func generateID() string {
	return fmt.Sprintf("miner-%d", time.Now().UnixNano())
}

// GetVersion returns the package version
func GetVersion() string {
	return "0.1.0"
}
