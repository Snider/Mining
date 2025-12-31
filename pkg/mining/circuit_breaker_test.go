package mining

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreakerDefaultConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()

	if cfg.FailureThreshold != 3 {
		t.Errorf("expected FailureThreshold 3, got %d", cfg.FailureThreshold)
	}
	if cfg.ResetTimeout != 30*time.Second {
		t.Errorf("expected ResetTimeout 30s, got %v", cfg.ResetTimeout)
	}
	if cfg.SuccessThreshold != 1 {
		t.Errorf("expected SuccessThreshold 1, got %d", cfg.SuccessThreshold)
	}
}

func TestCircuitBreakerStateString(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("state %d: expected %s, got %s", tt.state, tt.expected, got)
		}
	}
}

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	if cb.State() != CircuitClosed {
		t.Error("expected initial state to be closed")
	}

	// Successful execution
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
	if cb.State() != CircuitClosed {
		t.Error("state should still be closed after success")
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     time.Minute,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	testErr := errors.New("test error")

	// First failure
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, testErr
	})
	if err != testErr {
		t.Errorf("expected test error, got %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Error("should still be closed after 1 failure")
	}

	// Second failure - should open circuit
	_, err = cb.Execute(func() (interface{}, error) {
		return nil, testErr
	})
	if err != testErr {
		t.Errorf("expected test error, got %v", err)
	}
	if cb.State() != CircuitOpen {
		t.Error("should be open after 2 failures")
	}
}

func TestCircuitBreakerRejectsWhenOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     time.Hour, // Long timeout to keep circuit open
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	// Open the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	if cb.State() != CircuitOpen {
		t.Fatal("circuit should be open")
	}

	// Next request should be rejected
	called := false
	_, err := cb.Execute(func() (interface{}, error) {
		called = true
		return "should not run", nil
	})

	if called {
		t.Error("function should not have been called when circuit is open")
	}
	if err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreakerTransitionsToHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	// Open the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	if cb.State() != CircuitOpen {
		t.Fatal("circuit should be open")
	}

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Next request should transition to half-open and execute
	result, err := cb.Execute(func() (interface{}, error) {
		return "probe success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "probe success" {
		t.Errorf("expected 'probe success', got %v", result)
	}
	if cb.State() != CircuitClosed {
		t.Error("should be closed after successful probe")
	}
}

func TestCircuitBreakerHalfOpenFailureReopens(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	// Open the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Probe fails
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("probe failed")
	})

	if cb.State() != CircuitOpen {
		t.Error("should be open after probe failure")
	}
}

func TestCircuitBreakerCaching(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     time.Hour,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	// Successful call - caches result
	result, err := cb.Execute(func() (interface{}, error) {
		return "cached value", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "cached value" {
		t.Fatalf("expected 'cached value', got %v", result)
	}

	// Open the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	// Should return cached value when circuit is open
	result, err = cb.Execute(func() (interface{}, error) {
		return "should not run", nil
	})

	if err != nil {
		t.Errorf("expected cached result, got error: %v", err)
	}
	if result != "cached value" {
		t.Errorf("expected 'cached value', got %v", result)
	}
}

func TestCircuitBreakerGetCached(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	// No cache initially
	_, ok := cb.GetCached()
	if ok {
		t.Error("expected no cached value initially")
	}

	// Cache a value
	cb.Execute(func() (interface{}, error) {
		return "test value", nil
	})

	cached, ok := cb.GetCached()
	if !ok {
		t.Error("expected cached value")
	}
	if cached != "test value" {
		t.Errorf("expected 'test value', got %v", cached)
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     time.Hour,
		SuccessThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)

	// Open the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	if cb.State() != CircuitOpen {
		t.Fatal("circuit should be open")
	}

	// Manual reset
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Error("circuit should be closed after reset")
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cb.Execute(func() (interface{}, error) {
				if n%3 == 0 {
					return nil, errors.New("fail")
				}
				return "success", nil
			})
		}(i)
	}
	wg.Wait()

	// Just verify no panics occurred
	_ = cb.State()
}

func TestGetGitHubCircuitBreaker(t *testing.T) {
	cb1 := getGitHubCircuitBreaker()
	cb2 := getGitHubCircuitBreaker()

	if cb1 != cb2 {
		t.Error("expected singleton circuit breaker")
	}

	if cb1.name != "github-api" {
		t.Errorf("expected name 'github-api', got %s", cb1.name)
	}
}

// Benchmark tests
func BenchmarkCircuitBreakerExecute(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(func() (interface{}, error) {
			return "result", nil
		})
	}
}

func BenchmarkCircuitBreakerConcurrent(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() (interface{}, error) {
				return "result", nil
			})
		}
	})
}
