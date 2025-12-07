package mining

import (
	"time"
)

const (
	HighResolutionDuration = 5 * time.Minute
	HighResolutionInterval = 10 * time.Second
	LowResolutionInterval  = 1 * time.Minute
	LowResHistoryRetention = 24 * time.Hour
)

// Miner defines the standard interface for a cryptocurrency miner.
type Miner interface {
	Install() error
	Uninstall() error
	Start(config *Config) error
	Stop() error
	GetStats() (*PerformanceMetrics, error)
	GetName() string
	GetPath() string
	GetBinaryPath() string
	CheckInstallation() (*InstallationDetails, error)
	GetLatestVersion() (string, error)
	GetHashrateHistory() []HashratePoint
	AddHashratePoint(point HashratePoint)
	ReduceHashrateHistory(now time.Time)
}

// InstallationDetails contains information about an installed miner.
type InstallationDetails struct {
	IsInstalled bool   `json:"is_installed"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	MinerBinary string `json:"miner_binary"`
	ConfigPath  string `json:"config_path,omitempty"` // Add path to the miner-specific config
}

// SystemInfo provides general system and miner installation information.
type SystemInfo struct {
	Timestamp           time.Time              `json:"timestamp"`
	OS                  string                 `json:"os"`
	Architecture        string                 `json:"architecture"`
	GoVersion           string                 `json:"go_version"`
	AvailableCPUCores   int                    `json:"available_cpu_cores"`
	TotalSystemRAMGB    float64                `json:"total_system_ram_gb"`
	InstalledMinersInfo []*InstallationDetails `json:"installed_miners_info"`
}

// Config represents the configuration for a miner.
type Config struct {
	Miner             string `json:"miner"`
	Pool              string `json:"pool"`
	Wallet            string `json:"wallet"`
	Threads           int    `json:"threads"`
	TLS               bool   `json:"tls"`
	HugePages         bool   `json:"hugePages"`
	Algo              string `json:"algo,omitempty"`
	Coin              string `json:"coin,omitempty"`
	Password          string `json:"password,omitempty"`
	UserPass          string `json:"userPass,omitempty"`
	Proxy             string `json:"proxy,omitempty"`
	Keepalive         bool   `json:"keepalive,omitempty"`
	Nicehash          bool   `json:"nicehash,omitempty"`
	RigID             string `json:"rigId,omitempty"`
	TLSSingerprint    string `json:"tlsFingerprint,omitempty"`
	Retries           int    `json:"retries,omitempty"`
	RetryPause        int    `json:"retryPause,omitempty"`
	UserAgent         string `json:"userAgent,omitempty"`
	DonateLevel       int    `json:"donateLevel,omitempty"`
	DonateOverProxy   bool   `json:"donateOverProxy,omitempty"`
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
	APIWorkerID       string `json:"apiWorkerId,omitempty"`
	APIID             string `json:"apiId,omitempty"`
	HTTPHost          string `json:"httpHost,omitempty"`
	HTTPPort          int    `json:"httpPort,omitempty"`
	HTTPAccessToken   string `json:"httpAccessToken,omitempty"`
	HTTPNoRestricted  bool   `json:"httpNoRestricted,omitempty"`
	Syslog            bool   `json:"syslog,omitempty"`
	LogFile           string `json:"logFile,omitempty"`
	PrintTime         int    `json:"printTime,omitempty"`
	HealthPrintTime   int    `json:"healthPrintTime,omitempty"`
	NoColor           bool   `json:"noColor,omitempty"`
	Verbose           bool   `json:"verbose,omitempty"`
	LogOutput         bool   `json:"logOutput,omitempty"`
	Background        bool   `json:"background,omitempty"`
	Title             string `json:"title,omitempty"`
	NoTitle           bool   `json:"noTitle,omitempty"`
	PauseOnBattery    bool   `json:"pauseOnBattery,omitempty"`
	PauseOnActive     int    `json:"pauseOnActive,omitempty"`
	Stress            bool   `json:"stress,omitempty"`
	Bench             string `json:"bench,omitempty"`
	Submit            bool   `json:"submit,omitempty"`
	Verify            string `json:"verify,omitempty"`
	Seed              string `json:"seed,omitempty"`
	Hash              string `json:"hash,omitempty"`
	NoDMI             bool   `json:"noDMI,omitempty"`
}

// PerformanceMetrics represents the performance metrics for a miner.
type PerformanceMetrics struct {
	Hashrate  int                    `json:"hashrate"`
	Shares    int                    `json:"shares"`
	Rejected  int                    `json:"rejected"`
	Uptime    int                    `json:"uptime"`
	LastShare int64                  `json:"lastShare"`
	Algorithm string                 `json:"algorithm"`
	ExtraData map[string]interface{} `json:"extraData,omitempty"`
}

// HashratePoint represents a single hashrate measurement at a specific time.
type HashratePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Hashrate  int       `json:"hashrate"`
}

// API represents the miner's API configuration.
type API struct {
	Enabled    bool   `json:"enabled"`
	ListenHost string `json:"listenHost"`
	ListenPort int    `json:"listenPort"`
}

// XMRigSummary represents the summary of an XMRig miner's performance.
type XMRigSummary struct {
	Hashrate struct {
		Total []float64 `json:"total"`
	} `json:"hashrate"`
	Results struct {
		SharesGood  uint64 `json:"shares_good"`
		SharesTotal uint64 `json:"shares_total"`
	} `json:"results"`
	Uptime    uint64 `json:"uptime"`
	Algorithm string `json:"algorithm"`
}

// AvailableMiner represents a miner that is available for use.
type AvailableMiner struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
