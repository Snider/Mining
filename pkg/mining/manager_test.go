package mining

import (
	"testing"
)

// TestManager_StartStopMultipleMiners tests starting and stopping multiple miners.
func TestManager_StartStopMultipleMiners(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	configs := []*Config{
		{Pool: "pool1", Wallet: "wallet1"},
	}

	minerNames := []string{"xmrig"}

	for i, config := range configs {
		// Since we can't start a real miner in the test, we'll just check that the manager doesn't crash.
		// A more complete test would involve a mock miner.
		_, err := manager.StartMiner(minerNames[i], config)
		if err == nil {
			t.Errorf("Expected error when starting miner without executable")
		}
	}
}

// TestManager_collectMinerStats tests the stat collection logic.
func TestManager_collectMinerStats(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	// Since we can't start a real miner, we can't fully test this.
	// A more complete test would involve a mock miner that can be added to the manager.
	manager.collectMinerStats()
}

// TestManager_GetMinerHashrateHistory tests getting hashrate history.
func TestManager_GetMinerHashrateHistory(t *testing.T) {
	manager := NewManager()
	defer manager.Stop()

	_, err := manager.GetMinerHashrateHistory("non-existent")
	if err == nil {
		t.Error("Expected error for getting hashrate history for non-existent miner")
	}
}
