package mining

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/Snider/Mining/docs"
	"github.com/Snider/Mining/pkg/logging"
	"github.com/adrg/xdg"
	ginmcp "github.com/ckanthony/gin-mcp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/swaggo/swag"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Service encapsulates the gin-gonic router and the mining manager.
type Service struct {
	Manager             ManagerInterface
	ProfileManager      *ProfileManager
	NodeService         *NodeService
	EventHub            *EventHub
	Router              *gin.Engine
	Server              *http.Server
	DisplayAddr         string
	SwaggerInstanceName string
	APIBasePath         string
	SwaggerUIPath       string
	rateLimiter         *RateLimiter
	auth                *DigestAuth
	mcpServer           *ginmcp.GinMCP
}

// APIError represents a structured error response for the API
type APIError struct {
	Code       string `json:"code"`                 // Machine-readable error code
	Message    string `json:"message"`              // Human-readable message
	Details    string `json:"details,omitempty"`    // Technical details (for debugging)
	Suggestion string `json:"suggestion,omitempty"` // What to do next
	Retryable  bool   `json:"retryable"`            // Can the client retry?
}

// Error codes are defined in errors.go

// respondWithError sends a structured error response
func respondWithError(c *gin.Context, status int, code string, message string, details string) {
	apiErr := APIError{
		Code:      code,
		Message:   message,
		Details:   details,
		Retryable: isRetryableError(status),
	}

	// Add suggestions based on error code
	switch code {
	case ErrCodeMinerNotFound:
		apiErr.Suggestion = "Check the miner name or install the miner first"
	case ErrCodeProfileNotFound:
		apiErr.Suggestion = "Create a new profile or check the profile ID"
	case ErrCodeInstallFailed:
		apiErr.Suggestion = "Check your internet connection and try again"
	case ErrCodeStartFailed:
		apiErr.Suggestion = "Check the miner configuration and logs"
	case ErrCodeInvalidInput:
		apiErr.Suggestion = "Verify the request body matches the expected format"
	case ErrCodeServiceUnavailable:
		apiErr.Suggestion = "The service is temporarily unavailable, try again later"
		apiErr.Retryable = true
	}

	c.JSON(status, apiErr)
}

// respondWithMiningError sends a structured error response from a MiningError.
// This allows using pre-built error constructors from errors.go.
func respondWithMiningError(c *gin.Context, err *MiningError) {
	details := ""
	if err.Cause != nil {
		details = err.Cause.Error()
	}
	if err.Details != "" {
		if details != "" {
			details += "; "
		}
		details += err.Details
	}

	apiErr := APIError{
		Code:       err.Code,
		Message:    err.Message,
		Details:    details,
		Suggestion: err.Suggestion,
		Retryable:  err.Retryable,
	}

	c.JSON(err.StatusCode(), apiErr)
}

// isRetryableError determines if an error status code is retryable
func isRetryableError(status int) bool {
	return status == http.StatusServiceUnavailable ||
		status == http.StatusTooManyRequests ||
		status == http.StatusGatewayTimeout
}

// requestIDMiddleware adds a unique request ID to each request for tracing
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use existing request ID from header if provided, otherwise generate one
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set in context for use by handlers
		c.Set("requestID", requestID)

		// Set in response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID creates a unique request ID using timestamp and random bytes
func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = base64.StdEncoding.Decode(b, []byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("%d-%x", time.Now().UnixMilli(), b[:4])
}

// getRequestID extracts the request ID from gin context
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("requestID"); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// logWithRequestID logs a message with request ID correlation
func logWithRequestID(c *gin.Context, level string, message string, fields logging.Fields) {
	if fields == nil {
		fields = logging.Fields{}
	}
	if reqID := getRequestID(c); reqID != "" {
		fields["request_id"] = reqID
	}
	switch level {
	case "error":
		logging.Error(message, fields)
	case "warn":
		logging.Warn(message, fields)
	case "info":
		logging.Info(message, fields)
	default:
		logging.Debug(message, fields)
	}
}

// WebSocket upgrader for the events endpoint
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost origins only
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // No origin header (non-browser clients)
		}
		// Parse the origin URL properly to prevent bypass attacks
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := u.Hostname()
		// Only allow exact localhost matches
		return host == "localhost" || host == "127.0.0.1" || host == "::1" ||
			host == "wails.localhost"
	},
}

// NewService creates a new mining service
func NewService(manager ManagerInterface, listenAddr string, displayAddr string, swaggerNamespace string) (*Service, error) {
	apiBasePath := "/" + strings.Trim(swaggerNamespace, "/")
	swaggerUIPath := apiBasePath + "/swagger"

	docs.SwaggerInfo.Title = "Mining Module API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = displayAddr
	docs.SwaggerInfo.BasePath = apiBasePath
	instanceName := "swagger_" + strings.ReplaceAll(strings.Trim(swaggerNamespace, "/"), "/", "_")
	swag.Register(instanceName, docs.SwaggerInfo)

	profileManager, err := NewProfileManager()
	if err != nil {
		logging.Warn("failed to initialize profile manager", logging.Fields{"error": err})
		// Continue without profile manager - profile features will be degraded
		// Create a minimal in-memory profile manager as fallback
		profileManager = &ProfileManager{
			profiles: make(map[string]*MiningProfile),
		}
	}

	// Initialize node service (optional - only fails if XDG paths are broken)
	nodeService, err := NewNodeService()
	if err != nil {
		logging.Warn("failed to initialize node service", logging.Fields{"error": err})
		// Continue without node service - P2P features will be unavailable
	}

	// Initialize event hub for WebSocket real-time updates
	eventHub := NewEventHub()
	go eventHub.Run()

	// Wire up event hub to manager for miner events
	if mgr, ok := manager.(*Manager); ok {
		mgr.SetEventHub(eventHub)
	}

	// Set up state provider for WebSocket state sync on reconnect
	eventHub.SetStateProvider(func() interface{} {
		miners := manager.ListMiners()
		if len(miners) == 0 {
			return nil
		}
		// Return current state of all miners
		state := make([]map[string]interface{}, 0, len(miners))
		for _, miner := range miners {
			stats, _ := miner.GetStats(context.Background())
			minerState := map[string]interface{}{
				"name":   miner.GetName(),
				"status": "running",
			}
			if stats != nil {
				minerState["hashrate"] = stats.Hashrate
				minerState["shares"] = stats.Shares
				minerState["rejected"] = stats.Rejected
				minerState["uptime"] = stats.Uptime
			}
			state = append(state, minerState)
		}
		return map[string]interface{}{
			"miners": state,
		}
	})

	// Initialize authentication from environment
	authConfig := AuthConfigFromEnv()
	var auth *DigestAuth
	if authConfig.Enabled {
		auth = NewDigestAuth(authConfig)
		logging.Info("API authentication enabled", logging.Fields{"realm": authConfig.Realm})
	}

	return &Service{
		Manager:        manager,
		ProfileManager: profileManager,
		NodeService:    nodeService,
		EventHub:       eventHub,
		Server: &http.Server{
			Addr:              listenAddr,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		},
		DisplayAddr:         displayAddr,
		SwaggerInstanceName: instanceName,
		APIBasePath:         apiBasePath,
		SwaggerUIPath:       swaggerUIPath,
		auth:                auth,
	}, nil
}

// InitRouter initializes the Gin router and sets up all routes without starting an HTTP server.
// Use this when embedding the mining service in another application (e.g., Wails).
// After calling InitRouter, you can use the Router field directly as an http.Handler.
func (s *Service) InitRouter() {
	s.Router = gin.Default()

	// Extract port safely from server address for CORS
	serverPort := "9090" // default fallback
	if s.Server.Addr != "" {
		if _, port, err := net.SplitHostPort(s.Server.Addr); err == nil && port != "" {
			serverPort = port
		}
	}

	// Configure CORS to only allow local origins
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"http://localhost:4200", // Angular dev server
			"http://127.0.0.1:4200",
			"http://localhost:9090", // Default API port
			"http://127.0.0.1:9090",
			"http://localhost:" + serverPort,
			"http://127.0.0.1:" + serverPort,
			"http://wails.localhost", // Wails desktop app (uses localhost origin)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	s.Router.Use(cors.New(corsConfig))

	// Add request body size limit middleware (1MB max)
	s.Router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1MB
		c.Next()
	})

	// Add X-Request-ID middleware for request tracing
	s.Router.Use(requestIDMiddleware())

	// Add rate limiting (10 requests/second with burst of 20)
	s.rateLimiter = NewRateLimiter(10, 20)
	s.Router.Use(s.rateLimiter.Middleware())

	s.SetupRoutes()
}

// Stop gracefully stops the service and cleans up resources
func (s *Service) Stop() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
	if s.EventHub != nil {
		s.EventHub.Stop()
	}
	if s.auth != nil {
		s.auth.Stop()
	}
	if s.NodeService != nil {
		if err := s.NodeService.StopTransport(); err != nil {
			logging.Warn("failed to stop node service transport", logging.Fields{"error": err})
		}
	}
}

// ServiceStartup initializes the router and starts the HTTP server.
// For embedding without a standalone server, use InitRouter() instead.
func (s *Service) ServiceStartup(ctx context.Context) error {
	s.InitRouter()
	s.Server.Handler = s.Router

	// Channel to capture server startup errors
	errChan := make(chan error, 1)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("server error", logging.Fields{"addr": s.Server.Addr, "error": err})
			errChan <- err
		}
		close(errChan) // Prevent goroutine leak
	}()

	go func() {
		<-ctx.Done()
		s.Stop() // Clean up service resources (auth, event hub, node service)
		s.Manager.Stop()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Server.Shutdown(ctxShutdown); err != nil {
			logging.Error("server shutdown error", logging.Fields{"error": err})
		}
	}()

	// Verify server is actually listening by attempting to connect
	maxRetries := 50 // 50 * 100ms = 5 seconds max
	for i := 0; i < maxRetries; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("failed to start server: %w", err)
			}
			return nil // Channel closed without error means server shut down
		default:
			// Try to connect to verify server is listening
			conn, err := net.DialTimeout("tcp", s.Server.Addr, 50*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil // Server is ready
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("server failed to start listening on %s within timeout", s.Server.Addr)
}

// SetupRoutes configures all API routes on the Gin router.
// This is called automatically by ServiceStartup, but can also be called
// manually after InitRouter for embedding in other applications.
func (s *Service) SetupRoutes() {
	apiGroup := s.Router.Group(s.APIBasePath)

	// Health endpoints (no auth required for orchestration/monitoring)
	apiGroup.GET("/health", s.handleHealth)
	apiGroup.GET("/ready", s.handleReady)

	// Apply authentication middleware if enabled
	if s.auth != nil {
		apiGroup.Use(s.auth.Middleware())
	}

	{
		apiGroup.GET("/info", s.handleGetInfo)
		apiGroup.GET("/metrics", s.handleMetrics)
		apiGroup.POST("/doctor", s.handleDoctor)
		apiGroup.POST("/update", s.handleUpdateCheck)

		minersGroup := apiGroup.Group("/miners")
		{
			minersGroup.GET("", s.handleListMiners)
			minersGroup.GET("/available", s.handleListAvailableMiners)
			minersGroup.POST("/:miner_name/install", s.handleInstallMiner)
			minersGroup.DELETE("/:miner_name/uninstall", s.handleUninstallMiner)
			minersGroup.DELETE("/:miner_name", s.handleStopMiner)
			minersGroup.GET("/:miner_name/stats", s.handleGetMinerStats)
			minersGroup.GET("/:miner_name/hashrate-history", s.handleGetMinerHashrateHistory)
			minersGroup.GET("/:miner_name/logs", s.handleGetMinerLogs)
			minersGroup.POST("/:miner_name/stdin", s.handleMinerStdin)
		}

		// Historical data endpoints (database-backed)
		historyGroup := apiGroup.Group("/history")
		{
			historyGroup.GET("/status", s.handleHistoryStatus)
			historyGroup.GET("/miners", s.handleAllMinersHistoricalStats)
			historyGroup.GET("/miners/:miner_name", s.handleMinerHistoricalStats)
			historyGroup.GET("/miners/:miner_name/hashrate", s.handleMinerHistoricalHashrate)
		}

		profilesGroup := apiGroup.Group("/profiles")
		{
			profilesGroup.GET("", s.handleListProfiles)
			profilesGroup.POST("", s.handleCreateProfile)
			profilesGroup.GET("/:id", s.handleGetProfile)
			profilesGroup.PUT("/:id", s.handleUpdateProfile)
			profilesGroup.DELETE("/:id", s.handleDeleteProfile)
			profilesGroup.POST("/:id/start", s.handleStartMinerWithProfile)
		}

		// WebSocket endpoint for real-time events
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("/events", s.handleWebSocketEvents)
		}

		// Add P2P node endpoints if node service is available
		if s.NodeService != nil {
			s.NodeService.SetupRoutes(apiGroup)
		}
	}

	// Serve the embedded web component
	componentFS, err := GetComponentFS()
	if err == nil {
		s.Router.StaticFS("/component", componentFS)
	}

	swaggerURL := ginSwagger.URL(fmt.Sprintf("http://%s%s/doc.json", s.DisplayAddr, s.SwaggerUIPath))
	s.Router.GET(s.SwaggerUIPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL, ginSwagger.InstanceName(s.SwaggerInstanceName)))

	// Initialize MCP server for AI assistant integration
	// This exposes API endpoints as MCP tools for Claude, Cursor, etc.
	s.mcpServer = ginmcp.New(s.Router, &ginmcp.Config{
		Name:        "Mining API",
		Description: "Mining dashboard API exposed via Model Context Protocol (MCP)",
		BaseURL:     fmt.Sprintf("http://%s", s.DisplayAddr),
	})
	s.mcpServer.Mount(s.APIBasePath + "/mcp")
	logging.Info("MCP server enabled", logging.Fields{"endpoint": s.APIBasePath + "/mcp"})
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components,omitempty"`
}

// handleHealth godoc
// @Summary Health check endpoint
// @Description Returns service health status. Used for liveness probes.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (s *Service) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
	})
}

// handleReady godoc
// @Summary Readiness check endpoint
// @Description Returns service readiness with component status. Used for readiness probes.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Success 503 {object} HealthResponse
// @Router /ready [get]
func (s *Service) handleReady(c *gin.Context) {
	components := make(map[string]string)
	allReady := true

	// Check manager
	if s.Manager != nil {
		components["manager"] = "ready"
	} else {
		components["manager"] = "not initialized"
		allReady = false
	}

	// Check profile manager
	if s.ProfileManager != nil {
		components["profiles"] = "ready"
	} else {
		components["profiles"] = "degraded"
		// Don't fail readiness for degraded profile manager
	}

	// Check event hub
	if s.EventHub != nil {
		components["events"] = "ready"
	} else {
		components["events"] = "not initialized"
		allReady = false
	}

	// Check node service (optional)
	if s.NodeService != nil {
		components["p2p"] = "ready"
	} else {
		components["p2p"] = "disabled"
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allReady {
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, HealthResponse{
		Status:     status,
		Components: components,
	})
}

// handleGetInfo godoc
// @Summary Get live miner installation information
// @Description Retrieves live installation details for all miners, along with system information.
// @Tags system
// @Produce  json
// @Success 200 {object} SystemInfo
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /info [get]
func (s *Service) handleGetInfo(c *gin.Context) {
	systemInfo, err := s.updateInstallationCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get system info", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, systemInfo)
}

// updateInstallationCache performs a live check and updates the cache file.
func (s *Service) updateInstallationCache() (*SystemInfo, error) {
	// Always create a complete SystemInfo object
	systemInfo := &SystemInfo{
		Timestamp:           time.Now(),
		OS:                  runtime.GOOS,
		Architecture:        runtime.GOARCH,
		GoVersion:           runtime.Version(),
		AvailableCPUCores:   runtime.NumCPU(),
		InstalledMinersInfo: []*InstallationDetails{}, // Initialize as empty slice
	}

	vMem, err := mem.VirtualMemory()
	if err == nil {
		systemInfo.TotalSystemRAMGB = float64(vMem.Total) / (1024 * 1024 * 1024)
	}

	for _, availableMiner := range s.Manager.ListAvailableMiners() {
		miner, err := CreateMiner(availableMiner.Name)
		if err != nil {
			continue // Skip unsupported miner types
		}
		details, err := miner.CheckInstallation()
		if err != nil {
			logging.Warn("failed to check installation", logging.Fields{"miner": availableMiner.Name, "error": err})
		}
		systemInfo.InstalledMinersInfo = append(systemInfo.InstalledMinersInfo, details)
	}

	configDir, err := xdg.ConfigFile("lethean-desktop/miners")
	if err != nil {
		return nil, fmt.Errorf("could not get config directory: %w", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(systemInfo, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not marshal cache data: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return nil, fmt.Errorf("could not write cache file: %w", err)
	}

	return systemInfo, nil
}

// handleDoctor godoc
// @Summary Check miner installations
// @Description Performs a live check on all available miners to verify their installation status, version, and path.
// @Tags system
// @Produce  json
// @Success 200 {object} SystemInfo
// @Router /doctor [post]
func (s *Service) handleDoctor(c *gin.Context) {
	systemInfo, err := s.updateInstallationCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cache", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, systemInfo)
}

// handleUpdateCheck godoc
// @Summary Check for miner updates
// @Description Checks if any installed miners have a new version available for download.
// @Tags system
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /update [post]
func (s *Service) handleUpdateCheck(c *gin.Context) {
	updates := make(map[string]string)
	for _, availableMiner := range s.Manager.ListAvailableMiners() {
		miner, err := CreateMiner(availableMiner.Name)
		if err != nil {
			continue // Skip unsupported miner types
		}

		details, err := miner.CheckInstallation()
		if err != nil || !details.IsInstalled {
			continue
		}

		latestVersionStr, err := miner.GetLatestVersion()
		if err != nil {
			continue
		}

		latestVersion, err := semver.NewVersion(latestVersionStr)
		if err != nil {
			continue
		}

		installedVersion, err := semver.NewVersion(details.Version)
		if err != nil {
			continue
		}

		if latestVersion.GreaterThan(installedVersion) {
			updates[miner.GetName()] = latestVersion.String()
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "All miners are up to date."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updates_available": updates})
}

// handleUninstallMiner godoc
// @Summary Uninstall a miner
// @Description Removes all files for a specific miner.
// @Tags miners
// @Produce  json
// @Param miner_type path string true "Miner Type to uninstall"
// @Success 200 {object} map[string]string
// @Router /miners/{miner_type}/uninstall [delete]
func (s *Service) handleUninstallMiner(c *gin.Context) {
	minerType := c.Param("miner_name")
	if err := s.Manager.UninstallMiner(c.Request.Context(), minerType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.updateInstallationCache(); err != nil {
		logging.Warn("failed to update cache after uninstall", logging.Fields{"error": err})
	}
	c.JSON(http.StatusOK, gin.H{"status": minerType + " uninstalled successfully."})
}

// handleListMiners godoc
// @Summary List all running miners
// @Description Get a list of all running miners
// @Tags miners
// @Produce  json
// @Success 200 {array} XMRigMiner
// @Router /miners [get]
func (s *Service) handleListMiners(c *gin.Context) {
	miners := s.Manager.ListMiners()
	c.JSON(http.StatusOK, miners)
}

// handleListAvailableMiners godoc
// @Summary List all available miners
// @Description Get a list of all available miners
// @Tags miners
// @Produce  json
// @Success 200 {array} AvailableMiner
// @Router /miners/available [get]
func (s *Service) handleListAvailableMiners(c *gin.Context) {
	miners := s.Manager.ListAvailableMiners()
	c.JSON(http.StatusOK, miners)
}

// handleInstallMiner godoc
// @Summary Install or update a miner
// @Description Install a new miner or update an existing one.
// @Tags miners
// @Produce  json
// @Param miner_type path string true "Miner Type to install/update"
// @Success 200 {object} map[string]string
// @Router /miners/{miner_type}/install [post]
func (s *Service) handleInstallMiner(c *gin.Context) {
	minerType := c.Param("miner_name")
	miner, err := CreateMiner(minerType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown miner type"})
		return
	}

	if err := miner.Install(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.updateInstallationCache(); err != nil {
		logging.Warn("failed to update cache after install", logging.Fields{"error": err})
	}

	details, err := miner.CheckInstallation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify installation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "installed", "version": details.Version, "path": details.Path})
}

// handleStartMinerWithProfile godoc
// @Summary Start a new miner using a profile
// @Description Start a new miner with the configuration from a saved profile
// @Tags profiles
// @Produce  json
// @Param id path string true "Profile ID"
// @Success 200 {object} XMRigMiner
// @Router /profiles/{id}/start [post]
func (s *Service) handleStartMinerWithProfile(c *gin.Context) {
	profileID := c.Param("id")
	profile, exists := s.ProfileManager.GetProfile(profileID)
	if !exists {
		respondWithMiningError(c, ErrProfileNotFound(profileID))
		return
	}

	var config Config
	if err := json.Unmarshal(profile.Config, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse profile config", "details": err.Error()})
		return
	}

	miner, err := s.Manager.StartMiner(c.Request.Context(), profile.MinerType, &config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, miner)
}

// handleStopMiner godoc
// @Summary Stop a running miner
// @Description Stop a running miner by its name
// @Tags miners
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {object} map[string]string
// @Router /miners/{miner_name} [delete]
func (s *Service) handleStopMiner(c *gin.Context) {
	minerName := c.Param("miner_name")
	if err := s.Manager.StopMiner(c.Request.Context(), minerName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

// handleGetMinerStats godoc
// @Summary Get miner stats
// @Description Get statistics for a running miner
// @Tags miners
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {object} PerformanceMetrics
// @Router /miners/{miner_name}/stats [get]
func (s *Service) handleGetMinerStats(c *gin.Context) {
	minerName := c.Param("miner_name")
	miner, err := s.Manager.GetMiner(minerName)
	if err != nil {
		respondWithMiningError(c, ErrMinerNotFound(minerName).WithCause(err))
		return
	}
	stats, err := miner.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// handleGetMinerHashrateHistory godoc
// @Summary Get miner hashrate history
// @Description Get historical hashrate data for a running miner
// @Tags miners
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {array} HashratePoint
// @Router /miners/{miner_name}/hashrate-history [get]
func (s *Service) handleGetMinerHashrateHistory(c *gin.Context) {
	minerName := c.Param("miner_name")
	history, err := s.Manager.GetMinerHashrateHistory(minerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

// handleGetMinerLogs godoc
// @Summary Get miner log output
// @Description Get the captured stdout/stderr output from a running miner. Log lines are base64 encoded to preserve ANSI escape codes and special characters.
// @Tags miners
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {array} string "Base64 encoded log lines"
// @Router /miners/{miner_name}/logs [get]
func (s *Service) handleGetMinerLogs(c *gin.Context) {
	minerName := c.Param("miner_name")
	miner, err := s.Manager.GetMiner(minerName)
	if err != nil {
		respondWithMiningError(c, ErrMinerNotFound(minerName).WithCause(err))
		return
	}
	logs := miner.GetLogs()
	// Base64 encode each log line to preserve ANSI escape codes and special characters
	encodedLogs := make([]string, len(logs))
	for i, line := range logs {
		encodedLogs[i] = base64.StdEncoding.EncodeToString([]byte(line))
	}
	c.JSON(http.StatusOK, encodedLogs)
}

// StdinInput represents input to send to miner's stdin
type StdinInput struct {
	Input string `json:"input" binding:"required"`
}

// handleMinerStdin godoc
// @Summary Send input to miner stdin
// @Description Send console commands to a running miner's stdin (e.g., 'h' for hashrate, 'p' for pause)
// @Tags miners
// @Accept json
// @Produce json
// @Param miner_name path string true "Miner Name"
// @Param input body StdinInput true "Input to send"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /miners/{miner_name}/stdin [post]
func (s *Service) handleMinerStdin(c *gin.Context) {
	minerName := c.Param("miner_name")
	miner, err := s.Manager.GetMiner(minerName)
	if err != nil {
		respondWithError(c, http.StatusNotFound, ErrCodeMinerNotFound, "miner not found", err.Error())
		return
	}

	var input StdinInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	if err := miner.WriteStdin(input.Input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent", "input": input.Input})
}

// handleListProfiles godoc
// @Summary List all mining profiles
// @Description Get a list of all saved mining profiles
// @Tags profiles
// @Produce  json
// @Success 200 {array} MiningProfile
// @Router /profiles [get]
func (s *Service) handleListProfiles(c *gin.Context) {
	profiles := s.ProfileManager.GetAllProfiles()
	c.JSON(http.StatusOK, profiles)
}

// handleCreateProfile godoc
// @Summary Create a new mining profile
// @Description Create and save a new mining profile
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param profile body MiningProfile true "Mining Profile"
// @Success 201 {object} MiningProfile
// @Failure 400 {object} APIError "Invalid profile data"
// @Router /profiles [post]
func (s *Service) handleCreateProfile(c *gin.Context) {
	var profile MiningProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrCodeInvalidInput, "invalid profile data", err.Error())
		return
	}

	// Validate required fields
	if profile.Name == "" {
		respondWithError(c, http.StatusBadRequest, ErrCodeInvalidInput, "profile name is required", "")
		return
	}
	if profile.MinerType == "" {
		respondWithError(c, http.StatusBadRequest, ErrCodeInvalidInput, "miner type is required", "")
		return
	}

	createdProfile, err := s.ProfileManager.CreateProfile(&profile)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to create profile", err.Error())
		return
	}

	c.JSON(http.StatusCreated, createdProfile)
}

// handleGetProfile godoc
// @Summary Get a specific mining profile
// @Description Get a mining profile by its ID
// @Tags profiles
// @Produce  json
// @Param id path string true "Profile ID"
// @Success 200 {object} MiningProfile
// @Router /profiles/{id} [get]
func (s *Service) handleGetProfile(c *gin.Context) {
	profileID := c.Param("id")
	profile, exists := s.ProfileManager.GetProfile(profileID)
	if !exists {
		respondWithError(c, http.StatusNotFound, ErrCodeProfileNotFound, "profile not found", "")
		return
	}
	c.JSON(http.StatusOK, profile)
}

// handleUpdateProfile godoc
// @Summary Update a mining profile
// @Description Update an existing mining profile
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param id path string true "Profile ID"
// @Param profile body MiningProfile true "Updated Mining Profile"
// @Success 200 {object} MiningProfile
// @Failure 404 {object} APIError "Profile not found"
// @Router /profiles/{id} [put]
func (s *Service) handleUpdateProfile(c *gin.Context) {
	profileID := c.Param("id")
	var profile MiningProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrCodeInvalidInput, "invalid profile data", err.Error())
		return
	}
	profile.ID = profileID

	if err := s.ProfileManager.UpdateProfile(&profile); err != nil {
		// Check if error is "not found"
		if strings.Contains(err.Error(), "not found") {
			respondWithError(c, http.StatusNotFound, ErrCodeProfileNotFound, "profile not found", err.Error())
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to update profile", err.Error())
		return
	}
	c.JSON(http.StatusOK, profile)
}

// handleDeleteProfile godoc
// @Summary Delete a mining profile
// @Description Delete a mining profile by its ID. Idempotent - returns success even if profile doesn't exist.
// @Tags profiles
// @Produce  json
// @Param id path string true "Profile ID"
// @Success 200 {object} map[string]string
// @Router /profiles/{id} [delete]
func (s *Service) handleDeleteProfile(c *gin.Context) {
	profileID := c.Param("id")
	if err := s.ProfileManager.DeleteProfile(profileID); err != nil {
		// Make DELETE idempotent - if profile doesn't exist, still return success
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusOK, gin.H{"status": "profile deleted"})
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to delete profile", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "profile deleted"})
}

// handleHistoryStatus godoc
// @Summary Get database history status
// @Description Get the status of database persistence for historical data
// @Tags history
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /history/status [get]
func (s *Service) handleHistoryStatus(c *gin.Context) {
	if manager, ok := s.Manager.(*Manager); ok {
		c.JSON(http.StatusOK, gin.H{
			"enabled":       manager.IsDatabaseEnabled(),
			"retentionDays": manager.dbRetention,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": false, "error": "manager type not supported"})
}

// handleAllMinersHistoricalStats godoc
// @Summary Get historical stats for all miners
// @Description Get aggregated historical statistics for all miners from the database
// @Tags history
// @Produce  json
// @Success 200 {array} database.HashrateStats
// @Router /history/miners [get]
func (s *Service) handleAllMinersHistoricalStats(c *gin.Context) {
	manager, ok := s.Manager.(*Manager)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manager type not supported"})
		return
	}

	stats, err := manager.GetAllMinerHistoricalStats()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleMinerHistoricalStats godoc
// @Summary Get historical stats for a specific miner
// @Description Get aggregated historical statistics for a specific miner from the database
// @Tags history
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {object} database.HashrateStats
// @Router /history/miners/{miner_name} [get]
func (s *Service) handleMinerHistoricalStats(c *gin.Context) {
	minerName := c.Param("miner_name")
	manager, ok := s.Manager.(*Manager)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manager type not supported"})
		return
	}

	stats, err := manager.GetMinerHistoricalStats(minerName)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no historical data found for miner"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleMinerHistoricalHashrate godoc
// @Summary Get historical hashrate data for a specific miner
// @Description Get detailed historical hashrate data for a specific miner from the database
// @Tags history
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Param since query string false "Start time (RFC3339 format)"
// @Param until query string false "End time (RFC3339 format)"
// @Success 200 {array} HashratePoint
// @Router /history/miners/{miner_name}/hashrate [get]
func (s *Service) handleMinerHistoricalHashrate(c *gin.Context) {
	minerName := c.Param("miner_name")
	manager, ok := s.Manager.(*Manager)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manager type not supported"})
		return
	}

	// Parse time range from query params, default to last 24 hours
	until := time.Now()
	since := until.Add(-24 * time.Hour)

	if sinceStr := c.Query("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = t
		}
	}
	if untilStr := c.Query("until"); untilStr != "" {
		if t, err := time.Parse(time.RFC3339, untilStr); err == nil {
			until = t
		}
	}

	history, err := manager.GetMinerHistoricalHashrate(minerName, since, until)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// handleWebSocketEvents godoc
// @Summary WebSocket endpoint for real-time mining events
// @Description Upgrade to WebSocket for real-time mining stats and events.
// @Description Events include: miner.starting, miner.started, miner.stopping, miner.stopped, miner.stats, miner.error
// @Tags websocket
// @Success 101 {string} string "Switching Protocols"
// @Router /ws/events [get]
func (s *Service) handleWebSocketEvents(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Error("failed to upgrade WebSocket connection", logging.Fields{"error": err})
		return
	}

	logging.Info("new WebSocket connection", logging.Fields{"remote": c.Request.RemoteAddr})
	RecordWSConnection(true)
	if !s.EventHub.ServeWs(conn) {
		RecordWSConnection(false) // Undo increment on rejection
		logging.Warn("WebSocket connection rejected", logging.Fields{"remote": c.Request.RemoteAddr, "reason": "limit reached"})
	}
}

// handleMetrics godoc
// @Summary Get internal metrics
// @Description Returns internal metrics for monitoring and debugging
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (s *Service) handleMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, GetMetricsSnapshot())
}
