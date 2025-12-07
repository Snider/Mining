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
	"github.com/shirou/gopsutil/v4/mem" // Import mem for memory stats
	"github.com/swaggo/swag"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Service encapsulates the gin-gonic router and the mining manager.
type Service struct {
	Manager             ManagerInterface
	Router              *gin.Engine
	Server              *http.Server
	DisplayAddr         string
	SwaggerInstanceName string
	APIBasePath         string
	SwaggerUIPath       string
}

// NewService creates a new mining service
func NewService(manager ManagerInterface, listenAddr string, displayAddr string, swaggerNamespace string) *Service {
	apiBasePath := "/" + strings.Trim(swaggerNamespace, "/")
	swaggerUIPath := apiBasePath + "/swagger"

	docs.SwaggerInfo.Title = "Mining Module API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = displayAddr
	docs.SwaggerInfo.BasePath = apiBasePath
	instanceName := "swagger_" + strings.ReplaceAll(strings.Trim(swaggerNamespace, "/"), "/", "_")
	swag.Register(instanceName, docs.SwaggerInfo)

	return &Service{
		Manager: manager,
		Server: &http.Server{
			Addr: listenAddr,
		},
		DisplayAddr:         displayAddr,
		SwaggerInstanceName: instanceName,
		APIBasePath:         apiBasePath,
		SwaggerUIPath:       swaggerUIPath,
	}
}

func (s *Service) ServiceStartup(ctx context.Context) error {
	s.Router = gin.Default()
	s.Router.Use(cors.Default())
	s.setupRoutes()
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

func (s *Service) setupRoutes() {
	apiGroup := s.Router.Group(s.APIBasePath)
	{
		apiGroup.GET("/info", s.handleGetInfo)
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
			minersGroup.GET("/:miner_name/hashrate-history", s.handleGetMinerHashrateHistory)
		}
	}

	s.Router.StaticFile("/component/mining-dashboard.js", "./ui/dist/ui/mbe-mining-dashboard.js")

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
	configDir, err := xdg.ConfigFile("lethean-desktop/miners")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get config directory"})
		return
	}
	configPath := filepath.Join(configDir, "config.json")

	cacheBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cache file not found, run setup"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read cache file"})
		return
	}

	var systemInfo SystemInfo
	if err := json.Unmarshal(cacheBytes, &systemInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not parse cache file"})
		return
	}

	systemInfo.Timestamp = time.Now()
	vMem, err := mem.VirtualMemory()
	if err == nil {
		systemInfo.TotalSystemRAMGB = float64(vMem.Total) / (1024 * 1024 * 1024)
	}

	c.JSON(http.StatusOK, systemInfo)
}

// updateInstallationCache performs a live check and updates the cache file.
func (s *Service) updateInstallationCache() (*SystemInfo, error) {
	var allDetails []*InstallationDetails
	for _, availableMiner := range s.Manager.ListAvailableMiners() {
		var miner Miner
		switch availableMiner.Name {
		case "xmrig":
			miner = NewXMRigMiner()
		default:
			continue
		}
		details, _ := miner.CheckInstallation()
		allDetails = append(allDetails, details)
	}

	systemInfo := &SystemInfo{
		Timestamp:           time.Now(),
		OS:                  runtime.GOOS,
		Architecture:        runtime.GOARCH,
		GoVersion:           runtime.Version(),
		AvailableCPUCores:   runtime.NumCPU(),
		InstalledMinersInfo: allDetails,
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
