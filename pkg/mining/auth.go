package mining

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/gin-gonic/gin"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Enabled determines if authentication is required
	Enabled bool
	// Username for basic/digest auth
	Username string
	// Password for basic/digest auth
	Password string
	// Realm for digest auth
	Realm string
	// NonceExpiry is how long a nonce is valid
	NonceExpiry time.Duration
}

// DefaultAuthConfig returns the default auth configuration.
// Auth is disabled by default for local development.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Enabled:     false,
		Username:    "",
		Password:    "",
		Realm:       "Mining API",
		NonceExpiry: 5 * time.Minute,
	}
}

// AuthConfigFromEnv creates auth config from environment variables.
// Set MINING_API_AUTH=true to enable, MINING_API_USER and MINING_API_PASS for credentials.
func AuthConfigFromEnv() AuthConfig {
	config := DefaultAuthConfig()

	if os.Getenv("MINING_API_AUTH") == "true" {
		config.Enabled = true
		config.Username = os.Getenv("MINING_API_USER")
		config.Password = os.Getenv("MINING_API_PASS")

		if config.Username == "" || config.Password == "" {
			logging.Warn("API auth enabled but credentials not set", logging.Fields{
				"hint": "Set MINING_API_USER and MINING_API_PASS environment variables",
			})
			config.Enabled = false
		}
	}

	if realm := os.Getenv("MINING_API_REALM"); realm != "" {
		config.Realm = realm
	}

	return config
}

// DigestAuth implements HTTP Digest Authentication middleware
type DigestAuth struct {
	config   AuthConfig
	nonces   sync.Map // map[string]time.Time for nonce expiry tracking
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewDigestAuth creates a new digest auth middleware
func NewDigestAuth(config AuthConfig) *DigestAuth {
	da := &DigestAuth{
		config:   config,
		stopChan: make(chan struct{}),
	}
	// Start nonce cleanup goroutine
	go da.cleanupNonces()
	return da
}

// Stop gracefully shuts down the DigestAuth, stopping the cleanup goroutine.
// Safe to call multiple times.
func (da *DigestAuth) Stop() {
	da.stopOnce.Do(func() {
		close(da.stopChan)
	})
}

// Middleware returns a Gin middleware that enforces digest authentication
func (da *DigestAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !da.config.Enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			da.sendChallenge(c)
			return
		}

		// Try digest auth first
		if strings.HasPrefix(authHeader, "Digest ") {
			if da.validateDigest(c, authHeader) {
				c.Next()
				return
			}
			da.sendChallenge(c)
			return
		}

		// Fall back to basic auth
		if strings.HasPrefix(authHeader, "Basic ") {
			if da.validateBasic(c, authHeader) {
				c.Next()
				return
			}
		}

		da.sendChallenge(c)
	}
}

// sendChallenge sends a 401 response with digest auth challenge
func (da *DigestAuth) sendChallenge(c *gin.Context) {
	nonce := da.generateNonce()
	da.nonces.Store(nonce, time.Now())

	challenge := fmt.Sprintf(
		`Digest realm="%s", qop="auth", nonce="%s", opaque="%s"`,
		da.config.Realm,
		nonce,
		da.generateOpaque(),
	)

	c.Header("WWW-Authenticate", challenge)
	c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
		Code:       "AUTH_REQUIRED",
		Message:    "Authentication required",
		Suggestion: "Provide valid credentials using Digest or Basic authentication",
	})
}

// validateDigest validates a digest auth header
func (da *DigestAuth) validateDigest(c *gin.Context, authHeader string) bool {
	params := parseDigestParams(authHeader[7:]) // Skip "Digest "

	nonce := params["nonce"]
	if nonce == "" {
		return false
	}

	// Check nonce validity
	if storedTime, ok := da.nonces.Load(nonce); ok {
		if time.Since(storedTime.(time.Time)) > da.config.NonceExpiry {
			da.nonces.Delete(nonce)
			return false
		}
	} else {
		return false
	}

	// Validate username with constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(params["username"]), []byte(da.config.Username)) != 1 {
		return false
	}

	// Calculate expected response
	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", da.config.Username, da.config.Realm, da.config.Password))
	ha2 := md5Hash(fmt.Sprintf("%s:%s", c.Request.Method, params["uri"]))

	var expectedResponse string
	if params["qop"] == "auth" {
		expectedResponse = md5Hash(fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			ha1, nonce, params["nc"], params["cnonce"], params["qop"], ha2))
	} else {
		expectedResponse = md5Hash(fmt.Sprintf("%s:%s:%s", ha1, nonce, ha2))
	}

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(expectedResponse), []byte(params["response"])) == 1
}

// validateBasic validates a basic auth header
func (da *DigestAuth) validateBasic(c *gin.Context, authHeader string) bool {
	// Gin has built-in basic auth, but we do manual validation for consistency
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(da.config.Username)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(da.config.Password)) == 1

	return userMatch && passMatch
}

// generateNonce creates a cryptographically random nonce
func (da *DigestAuth) generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateOpaque creates an opaque value
func (da *DigestAuth) generateOpaque() string {
	return md5Hash(da.config.Realm)
}

// cleanupNonces removes expired nonces periodically
func (da *DigestAuth) cleanupNonces() {
	interval := da.config.NonceExpiry
	if interval <= 0 {
		interval = 5 * time.Minute // Default if not set
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-da.stopChan:
			return
		case <-ticker.C:
			now := time.Now()
			da.nonces.Range(func(key, value interface{}) bool {
				if now.Sub(value.(time.Time)) > da.config.NonceExpiry {
					da.nonces.Delete(key)
				}
				return true
			})
		}
	}
}

// parseDigestParams parses the parameters from a digest auth header
func parseDigestParams(header string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(header, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		value := strings.TrimSpace(part[idx+1:])
		// Remove quotes
		value = strings.Trim(value, `"`)
		params[key] = value
	}

	return params
}

// md5Hash returns the MD5 hash of a string as a hex string
func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
