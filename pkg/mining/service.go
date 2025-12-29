package mining

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/Snider/Mining/docs"
	"github.com/adrg/xdg"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	Router              *gin.Engine
	Server              *http.Server
	DisplayAddr         string
	SwaggerInstanceName string
	APIBasePath         string
	SwaggerUIPath       string
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
		return nil, fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	// Initialize node service (optional - only fails if XDG paths are broken)
	nodeService, err := NewNodeService()
	if err != nil {
		log.Printf("Warning: failed to initialize node service: %v", err)
		// Continue without node service - P2P features will be unavailable
	}

	return &Service{
		Manager:        manager,
		ProfileManager: profileManager,
		NodeService:    nodeService,
		Server: &http.Server{
			Addr: listenAddr,
		},
		DisplayAddr:         displayAddr,
		SwaggerInstanceName: instanceName,
		APIBasePath:         apiBasePath,
		SwaggerUIPath:       swaggerUIPath,
	}, nil
}

// InitRouter initializes the Gin router and sets up all routes without starting an HTTP server.
// Use this when embedding the mining service in another application (e.g., Wails).
// After calling InitRouter, you can use the Router field directly as an http.Handler.
func (s *Service) InitRouter() {
	s.Router = gin.Default()
	s.Router.Use(cors.Default())
	s.SetupRoutes()
}

// ServiceStartup initializes the router and starts the HTTP server.
// For embedding without a standalone server, use InitRouter() instead.
func (s *Service) ServiceStartup(ctx context.Context) error {
	s.InitRouter()
	s.Server.Handler = s.Router

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s: %v\n", s.Server.Addr, err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Manager.Stop()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Server.Shutdown(ctxShutdown); err != nil {
			log.Fatalf("server shutdown failed: %+v", err)
		}
	}()

	return nil
}

// SetupRoutes configures all API routes on the Gin router.
// This is called automatically by ServiceStartup, but can also be called
// manually after InitRouter for embedding in other applications.
func (s *Service) SetupRoutes() {
	apiGroup := s.Router.Group(s.APIBasePath)
	{
		apiGroup.GET("/info", s.handleGetInfo)
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
		var miner Miner
		switch availableMiner.Name {
		case "xmrig":
			miner = NewXMRigMiner()
		default:
			continue
		}
		details, _ := miner.CheckInstallation()
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

	if err := os.WriteFile(configPath, data, 0644); err != nil {
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
		var miner Miner
		switch availableMiner.Name {
		case "xmrig":
			miner = NewXMRigMiner()
		default:
			continue
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
	if err := s.Manager.UninstallMiner(minerType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.updateInstallationCache(); err != nil {
		log.Printf("Warning: failed to update cache after uninstall: %v", err)
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
	var miner Miner
	switch minerType {
	case "xmrig":
		miner = NewXMRigMiner()
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown miner type"})
		return
	}

	if err := miner.Install(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.updateInstallationCache(); err != nil {
		log.Printf("Warning: failed to update cache after install: %v", err)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	var config Config
	if err := json.Unmarshal(profile.Config, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse profile config", "details": err.Error()})
		return
	}

	miner, err := s.Manager.StartMiner(profile.MinerType, &config)
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
	if err := s.Manager.StopMiner(minerName); err != nil {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "miner not found"})
		return
	}
	stats, err := miner.GetStats()
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
// @Description Get the captured stdout/stderr output from a running miner
// @Tags miners
// @Produce  json
// @Param miner_name path string true "Miner Name"
// @Success 200 {array} string
// @Router /miners/{miner_name}/logs [get]
func (s *Service) handleGetMinerLogs(c *gin.Context) {
	minerName := c.Param("miner_name")
	miner, err := s.Manager.GetMiner(minerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "miner not found"})
		return
	}
	logs := miner.GetLogs()
	c.JSON(http.StatusOK, logs)
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
// @Router /profiles [post]
func (s *Service) handleCreateProfile(c *gin.Context) {
	var profile MiningProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdProfile, err := s.ProfileManager.CreateProfile(&profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile", "details": err.Error()})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
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
// @Router /profiles/{id} [put]
func (s *Service) handleUpdateProfile(c *gin.Context) {
	profileID := c.Param("id")
	var profile MiningProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile.ID = profileID

	if err := s.ProfileManager.UpdateProfile(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// handleDeleteProfile godoc
// @Summary Delete a mining profile
// @Description Delete a mining profile by its ID
// @Tags profiles
// @Produce  json
// @Param id path string true "Profile ID"
// @Success 200 {object} map[string]string
// @Router /profiles/{id} [delete]
func (s *Service) handleDeleteProfile(c *gin.Context) {
	profileID := c.Param("id")
	if err := s.ProfileManager.DeleteProfile(profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete profile", "details": err.Error()})
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
