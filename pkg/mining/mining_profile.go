package mining

// MiningProfile represents a saved configuration for a specific mining setup.
// This allows users to define and switch between different miners, pools,
// and wallets without re-entering information.
type MiningProfile struct {
	Name   string `json:"name"`   // A user-defined name for the profile, e.g., "My XMR Rig"
	Pool   string `json:"pool"`   // The mining pool address
	Wallet string `json:"wallet"` // The wallet address
	Miner  string `json:"miner"`  // The type of miner, e.g., "xmrig"
	// This can be expanded later to include the full *Config for advanced options
}
