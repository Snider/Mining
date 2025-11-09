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
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/mem" // Import mem for memory stats
	"github.com/swaggo/swag"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewService creates a new mining service
func NewService(manager *Manager, listenAddr string, displayAddr string, swaggerNamespace string) *Service {
	apiBasePath := "/" + strings.Trim(swaggerNamespace, "/")
	swaggerUIPath := apiBasePath + "/swagger" // Serve Swagger UI under a distinct sub-path

	// Dynamically configure Swagger at runtime
	docs.SwaggerInfo.Title = "Mining Module API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = displayAddr // Use the displayable address for Swagger UI
	docs.SwaggerInfo.BasePath = apiBasePath
	// Use a unique instance name to avoid conflicts in a multi-module environment
	instanceName := "swagger_" + strings.ReplaceAll(strings.Trim(swaggerNamespace, "/"), "/", "_")
	swag.Register(instanceName, docs.SwaggerInfo)

	return &Service{
		Manager: manager,
		Server: &http.Server{
			Addr: listenAddr, // Server listens on this address
		},
		DisplayAddr:         displayAddr, // Store displayable address for messages
		SwaggerInstanceName: instanceName,
		APIBasePath:         apiBasePath,
		SwaggerUIPath:       swaggerUIPath,
	}
}

func (s *Service) ServiceStartup(ctx context.Context) error {
	s.Router = gin.Default()
	s.setupRoutes()
	s.Server.Handler = s.Router

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s: %v\n", s.Server.Addr, err)
		}
	}()

	go func() {
		<-ctx.Done()
		// Stop the manager's background goroutines
		s.Manager.Stop()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Server.Shutdown(ctxShutdown); err != nil {
			log.Fatalf("server shutdown failed: %+v", err)
		}
	}()

	return nil
}

func (s *Service) setupRoutes() {
	// All API routes are now relative to the service's APIBasePath
	apiGroup := s.Router.Group(s.APIBasePath)
	{
		apiGroup.GET("/info", s.handleGetInfo) // New GET endpoint for cached info
		apiGroup.POST("/doctor", s.handleDoctor)
		apiGroup.POST("/update", s.handleUpdateCheck)

		minersGroup := apiGroup.Group("/miners")
		{
			minersGroup.GET("", s.handleListMiners)
			minersGroup.GET("/available", s.handleListAvailableMiners)
			minersGroup.POST("/:miner_name", s.handleStartMiner)
			minersGroup.POST("/:miner_name/install", s.handleInstallMiner)
			minersGroup.DELETE("/:miner_name/uninstall", s.handleUninstallMiner)
			minersGroup.DELETE("/:miner_name", s.handleStopMiner)
			minersGroup.GET("/:miner_name/stats", s.handleGetMinerStats)
			minersGroup.GET("/:miner_name/hashrate-history", s.handleGetMinerHashrateHistory) // New endpoint
		}
	}

	// New route to serve the custom HTML element bundle
	// This path now points to the output of the Angular project within the 'ui' directory
	s.Router.StaticFile("/component/mining-dashboard.js", "./ui/dist/ui/main.js")

	// Register Swagger UI route under a distinct sub-path to avoid conflicts
	swaggerURL := ginSwagger.URL(fmt.Sprintf("http://%s%s/doc.json", s.DisplayAddr, s.SwaggerUIPath))
	s.Router.GET(s.SwaggerUIPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL, ginSwagger.InstanceName(s.SwaggerInstanceName)))
}

// handleGetInfo godoc
// @Summary Get cached miner installation information
// @Description Retrieves the last cached installation details for all miners, along with system information.
// @Tags system
// @Produce  json
// @Success 200 {object} SystemInfo
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /info [get]
func (s *Service) handleGetInfo(c *gin.Context) {
	systemInfo := SystemInfo{
		Timestamp:         time.Now(),
		OS:                runtime.GOOS,
		Architecture:      runtime.GOARCH,
		GoVersion:         runtime.Version(),
		AvailableCPUCores: runtime.NumCPU(),
	}

	// Get total system RAM
	vMem, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Warning: Failed to get virtual memory info: %v", err)
		systemInfo.TotalSystemRAMGB = 0.0 // Default to 0 on error
	} else {
		// Convert bytes to GB
		systemInfo.TotalSystemRAMGB = float64(vMem.Total) / (1024 * 1024 * 1024)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get home directory"})
		return
	}
	signpostPath := filepath.Join(homeDir, ".installed-miners")

	configPathBytes, err := os.ReadFile(signpostPath)
	if err != nil {
		// If signpost or cache doesn't exist, return SystemInfo with empty miner details
		systemInfo.InstalledMinersInfo = []*InstallationDetails{}
		c.JSON(http.StatusOK, systemInfo)
		return
	}
	configPath := string(configPathBytes)

	cacheBytes, err := os.ReadFile(configPath)
	if err != nil {
		// If cache file is missing, return SystemInfo with empty miner details
		systemInfo.InstalledMinersInfo = []*InstallationDetails{}
		c.JSON(http.StatusOK, systemInfo)
		return
	}

	var cachedDetails []*InstallationDetails
	if err := json.Unmarshal(cacheBytes, &cachedDetails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not parse cache file"})
		return
	}

	// Filter for only installed miners
	var installedOnly []*InstallationDetails
	for _, detail := range cachedDetails {
		if detail.IsInstalled {
			installedOnly = append(installedOnly, detail)
		}
	}
	systemInfo.InstalledMinersInfo = installedOnly

	c.JSON(http.StatusOK, systemInfo)
}

// handleDoctor godoc
// @Summary Check miner installations
// @Description Performs a live check on all available miners to verify their installation status, version, and path.
// @Tags system
// @Produce  json
// @Success 200 {array} InstallationDetails
// @Router /doctor [post]
func (s *Service) handleDoctor(c *gin.Context) {
	var allDetails []*InstallationDetails
	for _, availableMiner := range s.Manager.ListAvailableMiners() {
		var miner Miner
		switch availableMiner.Name {
		case "xmrig":
			miner = NewXMRigMiner()
		default:
			continue
		}
		details, err := miner.CheckInstallation()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check " + miner.GetName(), "details": err.Error()})
			return
		}
		allDetails = append(allDetails, details)
	}
	c.JSON(http.StatusOK, allDetails)
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
	var miner Miner
	switch minerType {
	case "xmrig":
		miner = NewXMRigMiner()
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown miner type"})
		return
	}
	if err := miner.Uninstall(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": miner.GetName() + " uninstalled successfully."})
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

	details, err := miner.CheckInstallation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify installation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "installed", "version": details.Version, "path": details.Path})
}

// handleStartMiner godoc
// @Summary Start a new miner
// @Description Start a new miner with the given configuration
// @Tags miners
// @Accept  json
// @Produce  json
// @Param miner_type path string true "Miner Type"
// @Param config body Config true "Miner Configuration"
// @Success 200 {object} XMRigMiner
// @Router /miners/{miner_type} [post]
func (s *Service) handleStartMiner(c *gin.Context) {
	minerType := c.Param("miner_name")
	var config Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	miner, err := s.Manager.StartMiner(minerType, &config)
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
