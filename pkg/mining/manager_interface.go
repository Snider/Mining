package mining

// ManagerInterface defines the interface for a miner manager.
// This interface abstracts the core functionalities of a miner manager,
// allowing for different implementations to be used interchangeably. It provides
// a standard way to manage the lifecycle of miners and retrieve their data.
type ManagerInterface interface {
	// StartMiner starts a new miner with the given configuration.
	// It takes the miner type and a configuration object, and returns the
	// created miner instance or an error if the miner could not be started.
	StartMiner(minerType string, config *Config) (Miner, error)

	// StopMiner stops a running miner.
	// It takes the name of the miner to be stopped and returns an error if the
	// miner could not be stopped.
	StopMiner(name string) error

	// GetMiner retrieves a running miner by its name.
	// It returns the miner instance or an error if the miner is not found.
	GetMiner(name string) (Miner, error)

	// ListMiners returns a slice of all running miners.
	ListMiners() []Miner

	// ListAvailableMiners returns a list of available miners that can be started.
	// This provides a way to discover the types of miners supported by the manager.
	ListAvailableMiners() []AvailableMiner

	// GetMinerHashrateHistory returns the hashrate history for a specific miner.
	// It takes the name of the miner and returns a slice of hashrate points
	// or an error if the miner is not found.
	GetMinerHashrateHistory(name string) ([]HashratePoint, error)

	// Stop stops the manager and its background goroutines.
	// It should be called when the manager is no longer needed to ensure a
	// graceful shutdown of any background processes.
	Stop()
}
