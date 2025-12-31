package mining

import (
	"errors"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/logging"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed means the circuit is functioning normally
	CircuitClosed CircuitState = iota
	// CircuitOpen means the circuit has tripped and requests are being rejected
	CircuitOpen
	// CircuitHalfOpen means the circuit is testing if the service has recovered
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures before opening the circuit
	FailureThreshold int
	// ResetTimeout is how long to wait before attempting recovery
	ResetTimeout time.Duration
	// SuccessThreshold is the number of successes needed in half-open state to close
	SuccessThreshold int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     30 * time.Second,
		SuccessThreshold: 1,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	config         CircuitBreakerConfig
	state          CircuitState
	failures       int
	successes      int
	lastFailure    time.Time
	mu             sync.RWMutex
	cachedResult   interface{}
	cachedErr      error
	lastCacheTime  time.Time
	cacheDuration  time.Duration
}

// ErrCircuitOpen is returned when the circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:          name,
		config:        config,
		state:         CircuitClosed,
		cacheDuration: 5 * time.Minute, // Cache successful results for 5 minutes
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	// Check if we should allow this request
	if !cb.allowRequest() {
		// Return cached result if available
		cb.mu.RLock()
		if cb.cachedResult != nil && time.Since(cb.lastCacheTime) < cb.cacheDuration {
			result := cb.cachedResult
			cb.mu.RUnlock()
			logging.Debug("circuit breaker returning cached result", logging.Fields{
				"name":  cb.name,
				"state": cb.state.String(),
			})
			return result, nil
		}
		cb.mu.RUnlock()
		return nil, ErrCircuitOpen
	}

	// Execute the function
	result, err := fn()

	// Record the result
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess(result)
	}

	return result, err
}

// allowRequest checks if a request should be allowed through
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailure) > cb.config.ResetTimeout {
			cb.state = CircuitHalfOpen
			cb.successes = 0
			logging.Info("circuit breaker transitioning to half-open", logging.Fields{
				"name": cb.name,
			})
			return true
		}
		return false

	case CircuitHalfOpen:
		// Allow probe requests through
		return true

	default:
		return false
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
			logging.Warn("circuit breaker opened", logging.Fields{
				"name":     cb.name,
				"failures": cb.failures,
			})
		}

	case CircuitHalfOpen:
		// Probe failed, back to open
		cb.state = CircuitOpen
		logging.Warn("circuit breaker probe failed, reopening", logging.Fields{
			"name": cb.name,
		})
	}
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess(result interface{}) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Cache the successful result
	cb.cachedResult = result
	cb.lastCacheTime = time.Now()
	cb.cachedErr = nil

	switch cb.state {
	case CircuitClosed:
		// Reset failure count on success
		cb.failures = 0

	case CircuitHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = CircuitClosed
			cb.failures = 0
			logging.Info("circuit breaker closed after successful probe", logging.Fields{
				"name": cb.name,
			})
		}
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failures = 0
	cb.successes = 0
	logging.Debug("circuit breaker manually reset", logging.Fields{
		"name": cb.name,
	})
}

// GetCached returns the cached result if available
func (cb *CircuitBreaker) GetCached() (interface{}, bool) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.cachedResult != nil && time.Since(cb.lastCacheTime) < cb.cacheDuration {
		return cb.cachedResult, true
	}
	return nil, false
}

// Global circuit breaker for GitHub API
var (
	githubCircuitBreaker     *CircuitBreaker
	githubCircuitBreakerOnce sync.Once
)

// getGitHubCircuitBreaker returns the shared GitHub API circuit breaker
func getGitHubCircuitBreaker() *CircuitBreaker {
	githubCircuitBreakerOnce.Do(func() {
		githubCircuitBreaker = NewCircuitBreaker("github-api", CircuitBreakerConfig{
			FailureThreshold: 3,
			ResetTimeout:     60 * time.Second, // Wait 1 minute before retrying
			SuccessThreshold: 1,
		})
	})
	return githubCircuitBreaker
}
