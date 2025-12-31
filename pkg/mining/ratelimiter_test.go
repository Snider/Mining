package mining

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	defer rl.Stop()

	if rl.requestsPerSecond != 10 {
		t.Errorf("Expected requestsPerSecond 10, got %d", rl.requestsPerSecond)
	}
	if rl.burst != 20 {
		t.Errorf("Expected burst 20, got %d", rl.burst)
	}
}

func TestRateLimiterStop(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	// Stop should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop panicked: %v", r)
		}
	}()

	rl.Stop()

	// Calling Stop again should not panic (idempotent)
	rl.Stop()
}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(10, 5) // 10 req/s, burst of 5
	defer rl.Stop()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// First 5 requests should succeed (burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests, got %d", w.Code)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(10, 2) // 10 req/s, burst of 2
	defer rl.Stop()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Exhaust rate limit for IP1
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// IP1 should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 should be rate limited, got %d", w.Code)
	}

	// IP2 should still be able to make requests
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("IP2 should not be rate limited, got %d", w.Code)
	}
}

func TestRateLimiterClientCount(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Initial count should be 0
	if count := rl.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should have 1 client now
	if count := rl.ClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}

	// Make request from different IP
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should have 2 clients now
	if count := rl.ClientCount(); count != 2 {
		t.Errorf("Expected 2 clients, got %d", count)
	}
}

func TestRateLimiterTokenRefill(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(100, 1) // 100 req/s, burst of 1 (refills quickly)
	defer rl.Stop()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request succeeds
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("First request should succeed, got %d", w.Code)
	}

	// Second request should fail (burst exhausted)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Second request should be rate limited, got %d", w.Code)
	}

	// Wait for token to refill (at 100 req/s, 1 token takes 10ms)
	time.Sleep(20 * time.Millisecond)

	// Third request should succeed (token refilled)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Third request should succeed after refill, got %d", w.Code)
	}
}
