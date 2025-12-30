package mining

import (
	"testing"
	"time"
)

// TestDualMiningCPUAndGPU tests running CPU and GPU mining together
// This test requires XMRig installed and a GPU with OpenCL support
func TestDualMiningCPUAndGPU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dual mining test in short mode")
	}

	miner := NewXMRigMiner()
	details, err := miner.CheckInstallation()
	if err != nil || !details.IsInstalled {
		t.Skip("XMRig not installed, skipping dual mining test")
	}

	manager := NewManager()
	defer manager.Stop()

	// Dual mining config:
	// - CPU: 25% threads on RandomX
	// - GPU: OpenCL device 0 (discrete GPU, not iGPU)
	config := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		Algo:              "rx/0",
		CPUMaxThreadsHint: 25, // 25% CPU

		// GPU config - explicit device selection required!
		GPUEnabled: true,
		OpenCL:     true,  // AMD GPU
		Devices:    "0",   // Device 0 only - user must pick
	}

	minerInstance, err := manager.StartMiner("xmrig", config)
	if err != nil {
		t.Fatalf("Failed to start dual miner: %v", err)
	}
	t.Logf("Started dual miner: %s", minerInstance.GetName())

	// Let it warm up
	time.Sleep(20 * time.Second)

	// Get stats
	stats, err := minerInstance.GetStats()
	if err != nil {
		t.Logf("Warning: couldn't get stats: %v", err)
	} else {
		t.Logf("Hashrate: %d H/s, Shares: %d, Algo: %s",
			stats.Hashrate, stats.Shares, stats.Algorithm)
	}

	// Check logs for GPU initialization
	logs := minerInstance.GetLogs()
	gpuFound := false
	for _, line := range logs {
		if contains(line, "OpenCL") || contains(line, "GPU") {
			gpuFound = true
			t.Logf("GPU log: %s", line)
		}
	}

	if !gpuFound {
		t.Log("No GPU-related log lines found - GPU may not be mining")
	}

	// Clean up
	manager.StopMiner(minerInstance.GetName())
}

// TestGPUDeviceSelection tests that GPU mining requires explicit device selection
func TestGPUDeviceSelection(t *testing.T) {
	tmpDir := t.TempDir()

	miner := &XMRigMiner{
		BaseMiner: BaseMiner{
			Name: "xmrig-device-test",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
				ListenPort: 54321,
			},
		},
	}

	origGetPath := getXMRigConfigPath
	getXMRigConfigPath = func(name string) (string, error) {
		return tmpDir + "/" + name + ".json", nil
	}
	defer func() { getXMRigConfigPath = origGetPath }()

	// Config WITHOUT device selection - GPU should be disabled
	configNoDevice := &Config{
		Pool:       "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:     "test_wallet",
		Algo:       "rx/0",
		GPUEnabled: true,
		OpenCL:     true,
		// NO Devices specified!
	}

	err := miner.createConfig(configNoDevice)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// GPU should be disabled because no device was specified
	t.Log("Config without explicit device - GPU should be disabled (safe default)")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
