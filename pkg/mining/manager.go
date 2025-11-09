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

	// Ensure the miner's internal name is used for map key
	minerKey := miner.GetName()
	if _, exists := m.miners[minerKey]; exists {
		return nil, fmt.Errorf("miner already started: %s", minerKey)
	}

	if err := miner.Start(config); err != nil {
		return nil, err
	}

	m.miners[minerKey] = miner
	return miner, nil
}

// StopMiner stops a running miner
func (m *Manager) StopMiner(name string) error {
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
	minerKey := strings.ToLower(name) // Normalize input name to lowercase
	miner, exists := m.miners[minerKey]
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
