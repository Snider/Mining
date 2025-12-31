package mining

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestXMRigDualMiningConfig(t *testing.T) {
	// Create a temp directory for the config
	tmpDir := t.TempDir()

	miner := &XMRigMiner{
		BaseMiner: BaseMiner{
			Name: "xmrig-dual-test",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
				ListenPort: 12345,
			},
		},
	}

	// Temporarily override config path
	origGetPath := getXMRigConfigPath
	getXMRigConfigPath = func(name string) (string, error) {
		return filepath.Join(tmpDir, name+".json"), nil
	}
	defer func() { getXMRigConfigPath = origGetPath }()

	// Config with CPU mining rx/0 and GPU mining kawpow on different pools
	config := &Config{
		// CPU config
		Pool:              "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:            "cpu_wallet_address",
		Algo:              "rx/0",
		CPUMaxThreadsHint: 50,

		// GPU config - separate pool and algo
		// MUST specify Devices explicitly - no auto-picking!
		GPUEnabled: true,
		GPUPool:    "stratum+tcp://ravencoin.pool.com:3333",
		GPUWallet:  "gpu_wallet_address",
		GPUAlgo:    "kawpow",
		CUDA:       true, // NVIDIA
		OpenCL:     false,
		Devices:    "0", // Explicit device selection required
	}

	err := miner.createConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Read and parse the generated config
	data, err := os.ReadFile(miner.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var generatedConfig map[string]interface{}
	if err := json.Unmarshal(data, &generatedConfig); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify pools
	pools, ok := generatedConfig["pools"].([]interface{})
	if !ok {
		t.Fatal("pools not found in config")
	}
	if len(pools) != 2 {
		t.Errorf("Expected 2 pools (CPU + GPU), got %d", len(pools))
	}

	// Verify CPU pool
	cpuPool := pools[0].(map[string]interface{})
	if cpuPool["url"] != "stratum+tcp://pool.supportxmr.com:3333" {
		t.Errorf("CPU pool URL mismatch: %v", cpuPool["url"])
	}
	if cpuPool["user"] != "cpu_wallet_address" {
		t.Errorf("CPU wallet mismatch: %v", cpuPool["user"])
	}
	if cpuPool["algo"] != "rx/0" {
		t.Errorf("CPU algo mismatch: %v", cpuPool["algo"])
	}

	// Verify GPU pool
	gpuPool := pools[1].(map[string]interface{})
	if gpuPool["url"] != "stratum+tcp://ravencoin.pool.com:3333" {
		t.Errorf("GPU pool URL mismatch: %v", gpuPool["url"])
	}
	if gpuPool["user"] != "gpu_wallet_address" {
		t.Errorf("GPU wallet mismatch: %v", gpuPool["user"])
	}
	if gpuPool["algo"] != "kawpow" {
		t.Errorf("GPU algo mismatch: %v", gpuPool["algo"])
	}

	// Verify CUDA enabled, OpenCL disabled
	cuda := generatedConfig["cuda"].(map[string]interface{})
	if cuda["enabled"] != true {
		t.Error("CUDA should be enabled")
	}

	opencl := generatedConfig["opencl"].(map[string]interface{})
	if opencl["enabled"] != false {
		t.Error("OpenCL should be disabled")
	}

	// Verify CPU config
	cpu := generatedConfig["cpu"].(map[string]interface{})
	if cpu["enabled"] != true {
		t.Error("CPU should be enabled")
	}
	if cpu["max-threads-hint"] != float64(50) {
		t.Errorf("CPU max-threads-hint mismatch: %v", cpu["max-threads-hint"])
	}

	t.Logf("Generated dual-mining config:\n%s", string(data))
}

func TestXMRigGPUOnlyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	miner := &XMRigMiner{
		BaseMiner: BaseMiner{
			Name: "xmrig-gpu-only",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
				ListenPort: 12346,
			},
		},
	}

	origGetPath := getXMRigConfigPath
	getXMRigConfigPath = func(name string) (string, error) {
		return filepath.Join(tmpDir, name+".json"), nil
	}
	defer func() { getXMRigConfigPath = origGetPath }()

	// GPU-only config using same pool for simplicity
	// MUST specify Devices explicitly - no auto-picking!
	config := &Config{
		Pool:       "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:     "test_wallet",
		Algo:       "rx/0",
		NoCPU:      true, // Disable CPU
		GPUEnabled: true,
		OpenCL:     true,  // AMD GPU
		CUDA:       true,  // Also NVIDIA
		Devices:    "0,1", // Explicit device selection required
	}

	err := miner.createConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	data, err := os.ReadFile(miner.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var generatedConfig map[string]interface{}
	json.Unmarshal(data, &generatedConfig)

	// Both GPU backends should be enabled
	cuda := generatedConfig["cuda"].(map[string]interface{})
	opencl := generatedConfig["opencl"].(map[string]interface{})

	if cuda["enabled"] != true {
		t.Error("CUDA should be enabled")
	}
	if opencl["enabled"] != true {
		t.Error("OpenCL should be enabled")
	}

	t.Logf("Generated GPU config:\n%s", string(data))
}

func TestXMRigCPUOnlyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	miner := &XMRigMiner{
		BaseMiner: BaseMiner{
			Name: "xmrig-cpu-only",
			API: &API{
				Enabled:    true,
				ListenHost: "127.0.0.1",
				ListenPort: 12347,
			},
		},
	}

	origGetPath := getXMRigConfigPath
	getXMRigConfigPath = func(name string) (string, error) {
		return filepath.Join(tmpDir, name+".json"), nil
	}
	defer func() { getXMRigConfigPath = origGetPath }()

	// CPU-only config (GPUEnabled defaults to false)
	config := &Config{
		Pool:   "stratum+tcp://pool.supportxmr.com:3333",
		Wallet: "test_wallet",
		Algo:   "rx/0",
	}

	err := miner.createConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	data, err := os.ReadFile(miner.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var generatedConfig map[string]interface{}
	json.Unmarshal(data, &generatedConfig)

	// GPU backends should be disabled
	cuda := generatedConfig["cuda"].(map[string]interface{})
	opencl := generatedConfig["opencl"].(map[string]interface{})

	if cuda["enabled"] != false {
		t.Error("CUDA should be disabled for CPU-only config")
	}
	if opencl["enabled"] != false {
		t.Error("OpenCL should be disabled for CPU-only config")
	}

	// Should only have 1 pool
	pools := generatedConfig["pools"].([]interface{})
	if len(pools) != 1 {
		t.Errorf("Expected 1 pool for CPU-only, got %d", len(pools))
	}

	t.Logf("Generated CPU-only config:\n%s", string(data))
}
