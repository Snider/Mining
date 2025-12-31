package mining

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter provides token bucket rate limiting per IP address
type RateLimiter struct {
	requestsPerSecond int
	burst             int
	clients           map[string]*rateLimitClient
	mu                sync.RWMutex
	stopChan          chan struct{}
	stopped           bool
}

type rateLimitClient struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter with the specified limits
func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	rl := &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
		clients:           make(map[string]*rateLimitClient),
		stopChan:          make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop removes stale clients periodically
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopChan:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

// cleanup removes clients that haven't made requests in 5 minutes
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, c := range rl.clients {
		if time.Since(c.lastCheck) > 5*time.Minute {
			delete(rl.clients, ip)
		}
	}
}

// Stop stops the rate limiter's cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.stopped {
		close(rl.stopChan)
		rl.stopped = true
	}
}

// Middleware returns a Gin middleware handler for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		cl, exists := rl.clients[ip]
		if !exists {
			cl = &rateLimitClient{tokens: float64(rl.burst), lastCheck: time.Now()}
			rl.clients[ip] = cl
		}

		// Token bucket algorithm
		now := time.Now()
		elapsed := now.Sub(cl.lastCheck).Seconds()
		cl.tokens += elapsed * float64(rl.requestsPerSecond)
		if cl.tokens > float64(rl.burst) {
			cl.tokens = float64(rl.burst)
		}
		cl.lastCheck = now

		if cl.tokens < 1 {
			rl.mu.Unlock()
			respondWithError(c, http.StatusTooManyRequests, "RATE_LIMITED",
				"too many requests", "rate limit exceeded")
			c.Abort()
			return
		}

		cl.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}

// ClientCount returns the number of tracked clients (for testing/monitoring)
func (rl *RateLimiter) ClientCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.clients)
}
