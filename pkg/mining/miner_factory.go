package mining

import (
	"fmt"
	"strings"
	"sync"
)

// MinerConstructor is a function that creates a new miner instance
type MinerConstructor func() Miner

// MinerFactory handles miner instantiation and registration
type MinerFactory struct {
	mu           sync.RWMutex
	constructors map[string]MinerConstructor
	aliases      map[string]string // maps aliases to canonical names
}

// globalFactory is the default factory instance
var globalFactory = NewMinerFactory()

// NewMinerFactory creates a new MinerFactory with default miners registered
func NewMinerFactory() *MinerFactory {
	f := &MinerFactory{
		constructors: make(map[string]MinerConstructor),
		aliases:      make(map[string]string),
	}
	f.registerDefaults()
	return f
}

// registerDefaults registers all built-in miners
func (f *MinerFactory) registerDefaults() {
	// XMRig miner (CPU/GPU RandomX, Cryptonight, etc.)
	f.Register("xmrig", func() Miner { return NewXMRigMiner() })

	// TT-Miner (GPU Kawpow, etc.)
	f.Register("tt-miner", func() Miner { return NewTTMiner() })
	f.RegisterAlias("ttminer", "tt-miner")

	// Simulated miner for testing and development
	f.Register(MinerTypeSimulated, func() Miner {
		return NewSimulatedMiner(SimulatedMinerConfig{
			Name:         "simulated-miner",
			Algorithm:    "rx/0",
			BaseHashrate: 1000,
			Variance:     0.1,
		})
	})
}

// Register adds a miner constructor to the factory
func (f *MinerFactory) Register(name string, constructor MinerConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.constructors[strings.ToLower(name)] = constructor
}

// RegisterAlias adds an alias for an existing miner type
func (f *MinerFactory) RegisterAlias(alias, canonicalName string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aliases[strings.ToLower(alias)] = strings.ToLower(canonicalName)
}

// Create instantiates a miner of the specified type
func (f *MinerFactory) Create(minerType string) (Miner, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	name := strings.ToLower(minerType)

	// Check for alias first
	if canonical, ok := f.aliases[name]; ok {
		name = canonical
	}

	constructor, ok := f.constructors[name]
	if !ok {
		return nil, fmt.Errorf("unsupported miner type: %s", minerType)
	}

	return constructor(), nil
}

// IsSupported checks if a miner type is registered
func (f *MinerFactory) IsSupported(minerType string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	name := strings.ToLower(minerType)

	// Check alias
	if canonical, ok := f.aliases[name]; ok {
		name = canonical
	}

	_, ok := f.constructors[name]
	return ok
}

// ListTypes returns all registered miner type names (excluding aliases)
func (f *MinerFactory) ListTypes() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.constructors))
	for name := range f.constructors {
		types = append(types, name)
	}
	return types
}

// --- Global factory functions for convenience ---

// CreateMiner creates a miner using the global factory
func CreateMiner(minerType string) (Miner, error) {
	return globalFactory.Create(minerType)
}

// IsMinerSupported checks if a miner type is supported using the global factory
func IsMinerSupported(minerType string) bool {
	return globalFactory.IsSupported(minerType)
}

// ListMinerTypes returns all registered miner types from the global factory
func ListMinerTypes() []string {
	return globalFactory.ListTypes()
}

// RegisterMinerType adds a miner constructor to the global factory
func RegisterMinerType(name string, constructor MinerConstructor) {
	globalFactory.Register(name, constructor)
}

// RegisterMinerAlias adds an alias to the global factory
func RegisterMinerAlias(alias, canonicalName string) {
	globalFactory.RegisterAlias(alias, canonicalName)
}
