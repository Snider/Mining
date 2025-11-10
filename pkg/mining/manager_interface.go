package mining

// ManagerInterface defines the interface for a miner manager.
type ManagerInterface interface {
	StartMiner(minerType string, config *Config) (Miner, error)
	StopMiner(name string) error
	GetMiner(name string) (Miner, error)
	ListMiners() []Miner
	ListAvailableMiners() []AvailableMiner
	GetMinerHashrateHistory(name string) ([]HashratePoint, error)
	Stop()
}
