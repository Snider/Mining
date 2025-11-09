package mining

import (
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Miner is the interface for a miner
type Miner interface {
	Install() error
	Uninstall() error
	Start(config *Config) error
	Stop() error
	GetStats() (*PerformanceMetrics, error)
	GetName() string
	GetPath() string
	CheckInstallation() (*InstallationDetails, error)
	GetLatestVersion() (string, error)
}

// InstallationDetails contains information about an installed miner
type InstallationDetails struct {
	IsInstalled bool   `json:"is_installed"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	MinerBinary string `json:"miner_binary"`
}

// SystemInfo provides general system and miner installation information
type SystemInfo struct {
	Timestamp           time.Time              `json:"timestamp"`
	OS                  string                 `json:"os"`
	Architecture        string                 `json:"architecture"`
	GoVersion           string                 `json:"go_version"`
	AvailableCPUCores   int                    `json:"available_cpu_cores"`
	TotalSystemRAMGB    float64                `json:"total_system_ram_gb"`
	InstalledMinersInfo []*InstallationDetails `json:"installed_miners_info"`
}

type Service struct {
	Manager             *Manager
	Router              *gin.Engine
	Server              *http.Server
	DisplayAddr         string // The address to display in messages (e.g., 127.0.0.1:8080)
	SwaggerInstanceName string
	APIBasePath         string // The base path for all API routes (e.g., /api/v1/mining)
	SwaggerUIPath       string // The path where Swagger UI assets are served (e.g., /api/v1/mining/swagger)
}

// Config represents the config for a miner, including XMRig specific options
type Config struct {
	Miner     string `json:"miner"`
	Pool      string `json:"pool"`
	Wallet    string `json:"wallet"`
	Threads   int    `json:"threads"`
	TLS       bool   `json:"tls"`
	HugePages bool   `json:"hugePages"`

	// Network options
	Algo            string `json:"algo,omitempty"`
	Coin            string `json:"coin,omitempty"`
	Password        string `json:"password,omitempty"` // Corresponds to -p, not --userpass
	UserPass        string `json:"userPass,omitempty"` // Corresponds to -O
	Proxy           string `json:"proxy,omitempty"`
	Keepalive       bool   `json:"keepalive,omitempty"`
	Nicehash        bool   `json:"nicehash,omitempty"`
	RigID           string `json:"rigId,omitempty"`
	TLSSingerprint  string `json:"tlsFingerprint,omitempty"`
	Retries         int    `json:"retries,omitempty"`
	RetryPause      int    `json:"retryPause,omitempty"`
	UserAgent       string `json:"userAgent,omitempty"`
	DonateLevel     int    `json:"donateLevel,omitempty"`
	DonateOverProxy bool   `json:"donateOverProxy,omitempty"`

	// CPU backend options
	NoCPU             bool   `json:"noCpu,omitempty"`
	CPUAffinity       string `json:"cpuAffinity,omitempty"`
	AV                int    `json:"av,omitempty"`
	CPUPriority       int    `json:"cpuPriority,omitempty"`
	CPUMaxThreadsHint int    `json:"cpuMaxThreadsHint,omitempty"`
	CPUMemoryPool     int    `json:"cpuMemoryPool,omitempty"`
	CPUNoYield        bool   `json:"cpuNoYield,omitempty"`
	HugepageSize      int    `json:"hugepageSize,omitempty"`
	HugePagesJIT      bool   `json:"hugePagesJIT,omitempty"`
	ASM               string `json:"asm,omitempty"`
	Argon2Impl        string `json:"argon2Impl,omitempty"`
	RandomXInit       int    `json:"randomXInit,omitempty"`
	RandomXNoNUMA     bool   `json:"randomXNoNuma,omitempty"`
	RandomXMode       string `json:"randomXMode,omitempty"`
	RandomX1GBPages   bool   `json:"randomX1GBPages,omitempty"`
	RandomXWrmsr      string `json:"randomXWrmsr,omitempty"`
	RandomXNoRdmsr    bool   `json:"randomXNoRdmsr,omitempty"`
	RandomXCacheQoS   bool   `json:"randomXCacheQoS,omitempty"`

	// API options (can be overridden or supplemented here)
	APIWorkerID      string `json:"apiWorkerId,omitempty"`
	APIID            string `json:"apiId,omitempty"`
	HTTPHost         string `json:"httpHost,omitempty"`
	HTTPPort         int    `json:"httpPort,omitempty"`
	HTTPAccessToken  string `json:"httpAccessToken,omitempty"`
	HTTPNoRestricted bool   `json:"httpNoRestricted,omitempty"`

	// Logging options
	Syslog          bool   `json:"syslog,omitempty"`
	LogFile         string `json:"logFile,omitempty"`
	PrintTime       int    `json:"printTime,omitempty"`
	HealthPrintTime int    `json:"healthPrintTime,omitempty"`
	NoColor         bool   `json:"noColor,omitempty"`
	Verbose         bool   `json:"verbose,omitempty"`

	// Misc options
	Background     bool   `json:"background,omitempty"`
	Title          string `json:"title,omitempty"`
	NoTitle        bool   `json:"noTitle,omitempty"`
	PauseOnBattery bool   `json:"pauseOnBattery,omitempty"`
	PauseOnActive  int    `json:"pauseOnActive,omitempty"`
	Stress         bool   `json:"stress,omitempty"`
	Bench          string `json:"bench,omitempty"`
	Submit         bool   `json:"submit,omitempty"`
	Verify         string `json:"verify,omitempty"`
	Seed           string `json:"seed,omitempty"`
	Hash           string `json:"hash,omitempty"`
	NoDMI          bool   `json:"noDMI,omitempty"`
}

// PerformanceMetrics represents the performance metrics for a miner
type PerformanceMetrics struct {
	Hashrate  int                    `json:"hashrate"`
	Shares    int                    `json:"shares"`
	Rejected  int                    `json:"rejected"`
	Uptime    int                    `json:"uptime"`
	LastShare int64                  `json:"lastShare"`
	Algorithm string                 `json:"algorithm"`
	ExtraData map[string]interface{} `json:"extraData,omitempty"`
}

// History represents the history of a miner
type History struct {
	Miner   string               `json:"miner"`
	Stats   []PerformanceMetrics `json:"stats"`
	Updated int64                `json:"updated"`
}

// XMRigMiner represents an XMRig miner
type XMRigMiner struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	URL           string `json:"url"`
	Path          string `json:"path"`         // This will now be the versioned folder path
	MinerBinary   string `json:"miner_binary"` // New field for the full path to the miner executable
	Running       bool   `json:"running"`
	LastHeartbeat int64  `json:"lastHeartbeat"`
	ConfigPath    string `json:"configPath"`
	API           *API   `json:"api"`
	mu            sync.Mutex
	cmd           *exec.Cmd `json:"-"`
}

// API represents the XMRig API configuration
type API struct {
	Enabled    bool   `json:"enabled"`
	ListenHost string `json:"listenHost"`
	ListenPort int    `json:"listenPort"`
}

// XMRigSummary represents the summary from the XMRig API
type XMRigSummary struct {
	Hashrate struct {
		Total []float64 `json:"total"`
	} `json:"hashrate"`
	Results struct {
		SharesGood  uint64 `json:"shares_good"`
		SharesTotal uint64 `json:"shares_total"`
	} `json:"results"`
	Uptime    uint64 `json:"uptime"`
	Algorithm string `json:"algo"`
}

// AvailableMiner represents a miner that is available to be started
type AvailableMiner struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
