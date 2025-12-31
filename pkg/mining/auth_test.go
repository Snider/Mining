package mining

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestDefaultAuthConfig(t *testing.T) {
	cfg := DefaultAuthConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.Username != "" {
		t.Error("expected Username to be empty by default")
	}
	if cfg.Password != "" {
		t.Error("expected Password to be empty by default")
	}
	if cfg.Realm != "Mining API" {
		t.Errorf("expected Realm to be 'Mining API', got %s", cfg.Realm)
	}
	if cfg.NonceExpiry != 5*time.Minute {
		t.Errorf("expected NonceExpiry to be 5 minutes, got %v", cfg.NonceExpiry)
	}
}

func TestAuthConfigFromEnv(t *testing.T) {
	// Save original env
	origAuth := os.Getenv("MINING_API_AUTH")
	origUser := os.Getenv("MINING_API_USER")
	origPass := os.Getenv("MINING_API_PASS")
	origRealm := os.Getenv("MINING_API_REALM")
	defer func() {
		os.Setenv("MINING_API_AUTH", origAuth)
		os.Setenv("MINING_API_USER", origUser)
		os.Setenv("MINING_API_PASS", origPass)
		os.Setenv("MINING_API_REALM", origRealm)
	}()

	t.Run("auth disabled by default", func(t *testing.T) {
		os.Setenv("MINING_API_AUTH", "")
		cfg := AuthConfigFromEnv()
		if cfg.Enabled {
			t.Error("expected Enabled to be false when env not set")
		}
	})

	t.Run("auth enabled with valid credentials", func(t *testing.T) {
		os.Setenv("MINING_API_AUTH", "true")
		os.Setenv("MINING_API_USER", "testuser")
		os.Setenv("MINING_API_PASS", "testpass")

		cfg := AuthConfigFromEnv()
		if !cfg.Enabled {
			t.Error("expected Enabled to be true")
		}
		if cfg.Username != "testuser" {
			t.Errorf("expected Username 'testuser', got %s", cfg.Username)
		}
		if cfg.Password != "testpass" {
			t.Errorf("expected Password 'testpass', got %s", cfg.Password)
		}
	})

	t.Run("auth disabled if credentials missing", func(t *testing.T) {
		os.Setenv("MINING_API_AUTH", "true")
		os.Setenv("MINING_API_USER", "")
		os.Setenv("MINING_API_PASS", "")

		cfg := AuthConfigFromEnv()
		if cfg.Enabled {
			t.Error("expected Enabled to be false when credentials missing")
		}
	})

	t.Run("custom realm", func(t *testing.T) {
		os.Setenv("MINING_API_AUTH", "")
		os.Setenv("MINING_API_REALM", "Custom Realm")

		cfg := AuthConfigFromEnv()
		if cfg.Realm != "Custom Realm" {
			t.Errorf("expected Realm 'Custom Realm', got %s", cfg.Realm)
		}
	})
}

func TestNewDigestAuth(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: time.Second,
	}

	da := NewDigestAuth(cfg)
	if da == nil {
		t.Fatal("expected non-nil DigestAuth")
	}

	// Cleanup
	da.Stop()
}

func TestDigestAuthStop(t *testing.T) {
	cfg := DefaultAuthConfig()
	da := NewDigestAuth(cfg)

	// Should not panic when called multiple times
	da.Stop()
	da.Stop()
	da.Stop()
}

func TestMiddlewareAuthDisabled(t *testing.T) {
	cfg := AuthConfig{Enabled: false}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", w.Body.String())
	}
}

func TestMiddlewareNoAuth(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth == "" {
		t.Error("expected WWW-Authenticate header")
	}
	if !authTestContains(wwwAuth, "Digest") {
		t.Error("expected Digest challenge in WWW-Authenticate")
	}
	if !authTestContains(wwwAuth, `realm="Test"`) {
		t.Error("expected realm in WWW-Authenticate")
	}
}

func TestMiddlewareBasicAuthValid(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareBasicAuthInvalid(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	testCases := []struct {
		name     string
		user     string
		password string
	}{
		{"wrong user", "wronguser", "pass"},
		{"wrong password", "user", "wrongpass"},
		{"both wrong", "wronguser", "wrongpass"},
		{"empty user", "", "pass"},
		{"empty password", "user", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetBasicAuth(tc.user, tc.password)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", w.Code)
			}
		})
	}
}

func TestMiddlewareDigestAuthValid(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "testuser",
		Password:    "testpass",
		Realm:       "Test Realm",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// First request to get nonce
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 to get nonce, got %d", w.Code)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	params := parseDigestParams(wwwAuth[7:]) // Skip "Digest "
	nonce := params["nonce"]

	if nonce == "" {
		t.Fatal("nonce not found in challenge")
	}

	// Build digest auth response
	uri := "/test"
	nc := "00000001"
	cnonce := "abc123"
	qop := "auth"

	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", cfg.Username, cfg.Realm, cfg.Password))
	ha2 := md5Hash(fmt.Sprintf("GET:%s", uri))
	response := md5Hash(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2))

	authHeader := fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", qop=%s, nc=%s, cnonce="%s", response="%s"`,
		cfg.Username, cfg.Realm, nonce, uri, qop, nc, cnonce, response,
	)

	// Second request with digest auth
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", authHeader)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d; body: %s", w2.Code, w2.Body.String())
	}
}

func TestMiddlewareDigestAuthInvalidNonce(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Try with a fake nonce that was never issued
	authHeader := `Digest username="user", realm="Test", nonce="fakenonce123", uri="/test", qop=auth, nc=00000001, cnonce="abc", response="xxx"`
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid nonce, got %d", w.Code)
	}
}

func TestMiddlewareDigestAuthExpiredNonce(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 50 * time.Millisecond, // Very short for testing
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Get a valid nonce
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	wwwAuth := w.Header().Get("WWW-Authenticate")
	params := parseDigestParams(wwwAuth[7:])
	nonce := params["nonce"]

	// Wait for nonce to expire
	time.Sleep(100 * time.Millisecond)

	// Try to use expired nonce
	uri := "/test"
	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", cfg.Username, cfg.Realm, cfg.Password))
	ha2 := md5Hash(fmt.Sprintf("GET:%s", uri))
	response := md5Hash(fmt.Sprintf("%s:%s:%s", ha1, nonce, ha2))

	authHeader := fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		cfg.Username, cfg.Realm, nonce, uri, response,
	)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", authHeader)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired nonce, got %d", w2.Code)
	}
}

func TestParseDigestParams(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "basic params",
			input: `username="john", realm="test"`,
			expected: map[string]string{
				"username": "john",
				"realm":    "test",
			},
		},
		{
			name:  "params with spaces",
			input: `  username = "john" ,   realm = "test"  `,
			expected: map[string]string{
				"username": "john",
				"realm":    "test",
			},
		},
		{
			name:  "unquoted values",
			input: `qop=auth, nc=00000001`,
			expected: map[string]string{
				"qop": "auth",
				"nc":  "00000001",
			},
		},
		{
			name:  "full digest header",
			input: `username="user", realm="Test", nonce="abc123", uri="/api", qop=auth, nc=00000001, cnonce="xyz", response="hash"`,
			expected: map[string]string{
				"username": "user",
				"realm":    "Test",
				"nonce":    "abc123",
				"uri":      "/api",
				"qop":      "auth",
				"nc":       "00000001",
				"cnonce":   "xyz",
				"response": "hash",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDigestParams(tc.input)
			for key, expectedVal := range tc.expected {
				if result[key] != expectedVal {
					t.Errorf("key %s: expected %s, got %s", key, expectedVal, result[key])
				}
			}
		})
	}
}

func TestMd5Hash(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"user:realm:password", func() string {
			h := md5.Sum([]byte("user:realm:password"))
			return hex.EncodeToString(h[:])
		}()},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := md5Hash(tc.input)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestNonceGeneration(t *testing.T) {
	cfg := DefaultAuthConfig()
	da := NewDigestAuth(cfg)
	defer da.Stop()

	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce := da.generateNonce()
		if len(nonce) != 32 { // 16 bytes = 32 hex chars
			t.Errorf("expected nonce length 32, got %d", len(nonce))
		}
		if nonces[nonce] {
			t.Error("duplicate nonce generated")
		}
		nonces[nonce] = true
	}
}

func TestOpaqueGeneration(t *testing.T) {
	cfg := AuthConfig{Realm: "TestRealm"}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	opaque1 := da.generateOpaque()
	opaque2 := da.generateOpaque()

	// Same realm should produce same opaque
	if opaque1 != opaque2 {
		t.Error("opaque should be consistent for same realm")
	}

	// Should be MD5 of realm
	expected := md5Hash("TestRealm")
	if opaque1 != expected {
		t.Errorf("expected opaque %s, got %s", expected, opaque1)
	}
}

func TestNonceCleanup(t *testing.T) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 50 * time.Millisecond,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	// Store a nonce
	nonce := da.generateNonce()
	da.nonces.Store(nonce, time.Now())

	// Verify it exists
	if _, ok := da.nonces.Load(nonce); !ok {
		t.Error("nonce should exist immediately after storing")
	}

	// Wait for cleanup (2x expiry to be safe)
	time.Sleep(150 * time.Millisecond)

	// Verify it was cleaned up
	if _, ok := da.nonces.Load(nonce); ok {
		t.Error("expired nonce should have been cleaned up")
	}
}

// Helper function
func authTestContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkMd5Hash(b *testing.B) {
	input := "user:realm:password"
	for i := 0; i < b.N; i++ {
		md5Hash(input)
	}
}

func BenchmarkNonceGeneration(b *testing.B) {
	cfg := DefaultAuthConfig()
	da := NewDigestAuth(cfg)
	defer da.Stop()

	for i := 0; i < b.N; i++ {
		da.generateNonce()
	}
}

func BenchmarkBasicAuthValidation(b *testing.B) {
	cfg := AuthConfig{
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
		Realm:       "Test",
		NonceExpiry: 5 * time.Minute,
	}
	da := NewDigestAuth(cfg)
	defer da.Stop()

	router := gin.New()
	router.Use(da.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("user:pass")))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
