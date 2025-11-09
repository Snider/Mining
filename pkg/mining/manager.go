package mining

import (
	"fmt"
	"strings"
)

// Manager handles miner lifecycle and operations
type Manager struct {
	miners map[string]Miner
}

// NewManager creates a new miner manager
func NewManager() *Manager {
	return &Manager{
		miners: make(map[string]Miner),
	}
}

// StartMiner starts a new miner with the given configuration
func (m *Manager) StartMiner(minerType string, config *Config) (Miner, error) {
	var miner Miner
	switch strings.ToLower(minerType) {
	case "xmrig":
		miner = NewXMRigMiner()
	default:
		return nil, fmt.Errorf("unsupported miner type: %s", minerType)
	}

	if _, exists := m.miners[miner.GetName()]; exists {
		return nil, fmt.Errorf("miner already started: %s", miner.GetName())
	}

	if err := miner.Start(config); err != nil {
		return nil, err
	}

	m.miners[miner.GetName()] = miner
	return miner, nil
}

// StopMiner stops a running miner
func (m *Manager) StopMiner(name string) error {
	miner, exists := m.miners[name]
	if !exists {
		return fmt.Errorf("miner not found: %s", name)
	}

	if err := miner.Stop(); err != nil {
		return err
	}

	delete(m.miners, name)
	return nil
}

// GetMiner retrieves a miner by ID
func (m *Manager) GetMiner(name string) (Miner, error) {
	miner, exists := m.miners[name]
	if !exists {
		return nil, fmt.Errorf("miner not found: %s", name)
	}
	return miner, nil
}

// ListMiners returns all miners
func (m *Manager) ListMiners() []Miner {
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
