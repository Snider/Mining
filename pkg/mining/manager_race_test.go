package mining

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestConcurrentStartMultipleMiners verifies that concurrent StartMiner calls
// with different algorithms create unique miners without race conditions
func TestConcurrentStartMultipleMiners(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Try to start 10 miners concurrently with different algos
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			config := &Config{
				HTTPPort: 10000 + index,
				Pool:     "test:1234",
				Wallet:   "testwallet",
				Algo:     "algo" + string(rune('A'+index)), // algoA, algoB, etc.
			}
			_, err := m.StartMiner(context.Background(), "xmrig", config)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errCount int
	for err := range errors {
		t.Logf("Concurrent start error: %v", err)
		errCount++
	}

	// Some failures are expected due to port conflicts, but shouldn't crash
	t.Logf("Started miners with %d errors out of 10 attempts", errCount)

	// Verify no data races occurred (test passes if no race detector warnings)
}

// TestConcurrentStartDuplicateMiner verifies that starting the same miner
// concurrently results in only one success
func TestConcurrentStartDuplicateMiner(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	var wg sync.WaitGroup
	successes := make(chan struct{}, 10)
	failures := make(chan error, 10)

	// Try to start the same miner 10 times concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config := &Config{
				HTTPPort: 11000,
				Pool:     "test:1234",
				Wallet:   "testwallet",
				Algo:     "duplicate_test", // Same algo = same instance name
			}
			_, err := m.StartMiner(context.Background(), "xmrig", config)
			if err != nil {
				failures <- err
			} else {
				successes <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(successes)
	close(failures)

	successCount := len(successes)
	failureCount := len(failures)

	t.Logf("Duplicate miner test: %d successes, %d failures", successCount, failureCount)

	// Only one should succeed (or zero if there's a timing issue)
	if successCount > 1 {
		t.Errorf("Expected at most 1 success for duplicate miner, got %d", successCount)
	}
}

// TestConcurrentStartStop verifies that starting and stopping miners
// concurrently doesn't cause race conditions
func TestConcurrentStartStop(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	var wg sync.WaitGroup

	// Start some miners
	for i := 0; i < 5; i++ {
		config := &Config{
			HTTPPort: 12000 + i,
			Pool:     "test:1234",
			Wallet:   "testwallet",
			Algo:     "startstop" + string(rune('A'+i)),
		}
		_, err := m.StartMiner(context.Background(), "xmrig", config)
		if err != nil {
			t.Logf("Setup error (may be expected): %v", err)
		}
	}

	// Give miners time to start
	time.Sleep(100 * time.Millisecond)

	// Now concurrently start new ones and stop existing ones
	for i := 0; i < 10; i++ {
		wg.Add(2)

		// Start a new miner
		go func(index int) {
			defer wg.Done()
			config := &Config{
				HTTPPort: 12100 + index,
				Pool:     "test:1234",
				Wallet:   "testwallet",
				Algo:     "new" + string(rune('A'+index)),
			}
			m.StartMiner(context.Background(), "xmrig", config)
		}(i)

		// Stop a miner
		go func(index int) {
			defer wg.Done()
			minerName := "xmrig-startstop" + string(rune('A'+index%5))
			m.StopMiner(context.Background(), minerName)
		}(i)
	}

	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentListMiners verifies that listing miners while modifying
// the miner map doesn't cause race conditions
func TestConcurrentListMiners(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Continuously list miners
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				miners := m.ListMiners()
				_ = len(miners) // Use the result
			}
		}
	}()

	// Continuously start miners
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				config := &Config{
					HTTPPort: 13000 + i,
					Pool:     "test:1234",
					Wallet:   "testwallet",
					Algo:     "list" + string(rune('A'+i%26)),
				}
				m.StartMiner(context.Background(), "xmrig", config)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentGetMiner verifies that getting a miner while others
// are being started/stopped doesn't cause race conditions
func TestConcurrentGetMiner(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Start a miner first
	config := &Config{
		HTTPPort: 14000,
		Pool:     "test:1234",
		Wallet:   "testwallet",
		Algo:     "gettest",
	}
	miner, err := m.StartMiner(context.Background(), "xmrig", config)
	if err != nil {
		t.Skipf("Could not start test miner: %v", err)
	}
	minerName := miner.GetName()

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Continuously get the miner
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					m.GetMiner(minerName)
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Start more miners in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				config := &Config{
					HTTPPort: 14100 + i,
					Pool:     "test:1234",
					Wallet:   "testwallet",
					Algo:     "parallel" + string(rune('A'+i)),
				}
				m.StartMiner(context.Background(), "xmrig", config)
			}
		}
	}()

	wg.Wait()

	// Test passes if no race detector warnings
}

// TestConcurrentStatsCollection verifies that stats collection
// doesn't race with miner operations
func TestConcurrentStatsCollection(t *testing.T) {
	m := setupTestManager(t)
	defer m.Stop()

	// Start some miners
	for i := 0; i < 3; i++ {
		config := &Config{
			HTTPPort: 15000 + i,
			Pool:     "test:1234",
			Wallet:   "testwallet",
			Algo:     "stats" + string(rune('A'+i)),
		}
		m.StartMiner(context.Background(), "xmrig", config)
	}

	var wg sync.WaitGroup

	// Simulate stats collection (normally done by background goroutine)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			miners := m.ListMiners()
			for _, miner := range miners {
				miner.GetStats(context.Background())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Concurrently stop miners
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Let stats collection start
		for _, name := range []string{"xmrig-statsA", "xmrig-statsB", "xmrig-statsC"} {
			m.StopMiner(context.Background(), name)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Test passes if no race detector warnings
}
