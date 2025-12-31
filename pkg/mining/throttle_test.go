package mining

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

// TestCPUThrottleSingleMiner tests that a single miner respects CPU throttle settings
func TestCPUThrottleSingleMiner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CPU throttle test in short mode")
	}

	miner := NewXMRigMiner()
	details, err := miner.CheckInstallation()
	if err != nil || !details.IsInstalled {
		t.Skip("XMRig not installed, skipping throttle test")
	}

	// Use the manager to start miner (handles API port assignment)
	manager := NewManager()
	defer manager.Stop()

	// Configure miner to use only 10% of CPU
	config := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		CPUMaxThreadsHint: 10, // 10% CPU usage
		Algo:              "rx/0",
	}

	minerInstance, err := manager.StartMiner("xmrig", config)
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}
	t.Logf("Started miner: %s", minerInstance.GetName())

	// Let miner warm up
	time.Sleep(15 * time.Second)

	// Measure CPU usage
	avgCPU := measureCPUUsage(t, 10*time.Second)

	t.Logf("Configured: 10%% CPU, Measured: %.1f%% CPU", avgCPU)

	// Allow 15% margin (10% target + 5% tolerance)
	if avgCPU > 25 {
		t.Errorf("CPU usage %.1f%% exceeds expected ~10%% (with tolerance)", avgCPU)
	}

	manager.StopMiner(minerInstance.GetName())
}

// TestCPUThrottleDualMiners tests that two miners together respect combined CPU limits
func TestCPUThrottleDualMiners(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CPU throttle test in short mode")
	}

	miner1 := NewXMRigMiner()
	details, err := miner1.CheckInstallation()
	if err != nil || !details.IsInstalled {
		t.Skip("XMRig not installed, skipping throttle test")
	}

	manager := NewManager()
	defer manager.Stop()

	// Start first miner at 10% CPU with RandomX
	config1 := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		CPUMaxThreadsHint: 10,
		Algo:              "rx/0",
	}

	miner1Instance, err := manager.StartMiner("xmrig", config1)
	if err != nil {
		t.Fatalf("Failed to start first miner: %v", err)
	}
	t.Logf("Started miner 1: %s", miner1Instance.GetName())

	// Start second miner at 10% CPU with different algo
	config2 := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:5555",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		CPUMaxThreadsHint: 10,
		Algo:              "gr", // GhostRider algo
	}

	miner2Instance, err := manager.StartMiner("xmrig", config2)
	if err != nil {
		t.Fatalf("Failed to start second miner: %v", err)
	}
	t.Logf("Started miner 2: %s", miner2Instance.GetName())

	// Let miners warm up
	time.Sleep(20 * time.Second)

	// Verify both miners are running
	miners := manager.ListMiners()
	if len(miners) != 2 {
		t.Fatalf("Expected 2 miners running, got %d", len(miners))
	}

	// Measure combined CPU usage
	avgCPU := measureCPUUsage(t, 15*time.Second)

	t.Logf("Configured: 2x10%% CPU, Measured: %.1f%% CPU", avgCPU)

	// Combined should be ~20% with tolerance
	if avgCPU > 40 {
		t.Errorf("Combined CPU usage %.1f%% exceeds expected ~20%% (with tolerance)", avgCPU)
	}

	// Clean up
	manager.StopMiner(miner1Instance.GetName())
	manager.StopMiner(miner2Instance.GetName())
}

// TestCPUThrottleThreadCount tests thread-based CPU limiting
func TestCPUThrottleThreadCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CPU throttle test in short mode")
	}

	miner := NewXMRigMiner()
	details, err := miner.CheckInstallation()
	if err != nil || !details.IsInstalled {
		t.Skip("XMRig not installed, skipping throttle test")
	}

	// Use the manager to start miner (handles API port assignment)
	manager := NewManager()
	defer manager.Stop()

	numCPU := runtime.NumCPU()
	targetThreads := 1 // Use only 1 thread
	expectedMaxCPU := float64(100) / float64(numCPU) * float64(targetThreads) * 1.5 // 50% tolerance

	config := &Config{
		Pool:    "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:  "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		Threads: targetThreads,
		Algo:    "rx/0",
	}

	minerInstance, err := manager.StartMiner("xmrig", config)
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}
	t.Logf("Started miner: %s", minerInstance.GetName())
	defer manager.StopMiner(minerInstance.GetName())

	// Let miner warm up
	time.Sleep(15 * time.Second)

	avgCPU := measureCPUUsage(t, 10*time.Second)

	t.Logf("CPUs: %d, Threads: %d, Expected max: %.1f%%, Measured: %.1f%%",
		numCPU, targetThreads, expectedMaxCPU, avgCPU)

	if avgCPU > expectedMaxCPU {
		t.Errorf("CPU usage %.1f%% exceeds expected max %.1f%% for %d thread(s)",
			avgCPU, expectedMaxCPU, targetThreads)
	}
}

// TestMinerResourceIsolation tests that miners don't interfere with each other
func TestMinerResourceIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource isolation test in short mode")
	}

	miner := NewXMRigMiner()
	details, err := miner.CheckInstallation()
	if err != nil || !details.IsInstalled {
		t.Skip("XMRig not installed, skipping test")
	}

	manager := NewManager()
	defer manager.Stop()

	// Start first miner
	config1 := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:3333",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		CPUMaxThreadsHint: 25,
		Algo:              "rx/0",
	}

	miner1, err := manager.StartMiner("xmrig", config1)
	if err != nil {
		t.Fatalf("Failed to start miner 1: %v", err)
	}

	time.Sleep(10 * time.Second)

	// Get baseline hashrate for miner 1 alone
	stats1Alone, err := miner1.GetStats(context.Background())
	if err != nil {
		t.Logf("Warning: couldn't get stats for miner 1: %v", err)
	}
	baselineHashrate := 0
	if stats1Alone != nil {
		baselineHashrate = stats1Alone.Hashrate
	}

	// Start second miner
	config2 := &Config{
		Pool:              "stratum+tcp://pool.supportxmr.com:5555",
		Wallet:            "44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3XjrpDtQGv7SqSsaBYBb98uNbr2VBBEt7f2wfn3RVGQBEP3A",
		CPUMaxThreadsHint: 25,
		Algo:              "gr",
	}

	miner2, err := manager.StartMiner("xmrig", config2)
	if err != nil {
		t.Fatalf("Failed to start miner 2: %v", err)
	}

	time.Sleep(15 * time.Second)

	// Check both miners are running and producing hashrate
	stats1, err := miner1.GetStats(context.Background())
	if err != nil {
		t.Logf("Warning: couldn't get stats for miner 1: %v", err)
	}
	stats2, err := miner2.GetStats(context.Background())
	if err != nil {
		t.Logf("Warning: couldn't get stats for miner 2: %v", err)
	}

	t.Logf("Miner 1 baseline: %d H/s, with miner 2: %d H/s", baselineHashrate, getHashrate(stats1))
	t.Logf("Miner 2 hashrate: %d H/s", getHashrate(stats2))

	// Both miners should be producing some hashrate
	if stats1 != nil && stats1.Hashrate == 0 {
		t.Error("Miner 1 has zero hashrate")
	}
	if stats2 != nil && stats2.Hashrate == 0 {
		t.Error("Miner 2 has zero hashrate")
	}

	// Clean up
	manager.StopMiner(miner1.GetName())
	manager.StopMiner(miner2.GetName())
}

// measureCPUUsage measures average CPU usage over a duration
func measureCPUUsage(t *testing.T, duration time.Duration) float64 {
	t.Helper()

	samples := int(duration.Seconds())
	if samples < 1 {
		samples = 1
	}

	var totalCPU float64
	for i := 0; i < samples; i++ {
		percentages, err := cpu.Percent(time.Second, false)
		if err != nil {
			t.Logf("Warning: failed to get CPU percentage: %v", err)
			continue
		}
		if len(percentages) > 0 {
			totalCPU += percentages[0]
		}
	}

	return totalCPU / float64(samples)
}

// measureProcessCPU measures CPU usage of a specific process
func measureProcessCPU(t *testing.T, pid int32, duration time.Duration) float64 {
	t.Helper()

	proc, err := process.NewProcess(pid)
	if err != nil {
		t.Logf("Warning: failed to get process: %v", err)
		return 0
	}

	samples := int(duration.Seconds())
	if samples < 1 {
		samples = 1
	}

	var totalCPU float64
	for i := 0; i < samples; i++ {
		pct, err := proc.CPUPercent()
		if err != nil {
			continue
		}
		totalCPU += pct
		time.Sleep(time.Second)
	}

	return totalCPU / float64(samples)
}

func getHashrate(stats *PerformanceMetrics) int {
	if stats == nil {
		return 0
	}
	return stats.Hashrate
}
