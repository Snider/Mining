package mining

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	HighResolutionDuration = 5 * time.Minute
	HighResolutionInterval = 10 * time.Second
	LowResolutionInterval  = 1 * time.Minute
	LowResHistoryRetention = 24 * time.Hour
)

// Miner defines the standard interface for a cryptocurrency miner.
// The interface is logically grouped into focused capabilities:
//
// Lifecycle - Installation and process management:
//   - Install, Uninstall, Start, Stop
//
// Stats - Performance metrics collection:
//   - GetStats
//
// Info - Miner identification and installation details:
//   - GetType, GetName, GetPath, GetBinaryPath, CheckInstallation, GetLatestVersion
//
// History - Hashrate history management:
//   - GetHashrateHistory, AddHashratePoint, ReduceHashrateHistory
//
// IO - Interactive input/output:
//   - GetLogs, WriteStdin
type Miner interface {
	// Lifecycle operations
	Install() error
	Uninstall() error
	Start(config *Config) error
	Stop() error

	// Stats operations
	GetStats(ctx context.Context) (*PerformanceMetrics, error)

	// Info operations
	GetType() string // Returns miner type identifier (e.g., "xmrig", "tt-miner")
	GetName() string
	GetPath() string
	GetBinaryPath() string
	CheckInstallation() (*InstallationDetails, error)
	GetLatestVersion() (string, error)

	// History operations
	GetHashrateHistory() []HashratePoint
	AddHashratePoint(point HashratePoint)
	ReduceHashrateHistory(now time.Time)

	// IO operations
	GetLogs() []string
	WriteStdin(input string) error
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
	// GPU-specific options (for XMRig dual CPU+GPU mining)
	GPUEnabled   bool   `json:"gpuEnabled,omitempty"`   // Enable GPU mining
	GPUPool      string `json:"gpuPool,omitempty"`      // Separate pool for GPU (can differ from CPU)
	GPUWallet    string `json:"gpuWallet,omitempty"`    // Wallet for GPU pool (defaults to main Wallet)
	GPUAlgo      string `json:"gpuAlgo,omitempty"`      // Algorithm for GPU (e.g., "kawpow", "ethash")
	GPUPassword  string `json:"gpuPassword,omitempty"`  // Password for GPU pool
	GPUIntensity int    `json:"gpuIntensity,omitempty"` // GPU mining intensity (0-100)
	GPUThreads   int    `json:"gpuThreads,omitempty"`   // GPU threads per card
	Devices      string `json:"devices,omitempty"`      // GPU device selection (e.g., "0,1,2")
	OpenCL       bool   `json:"opencl,omitempty"`       // Enable OpenCL (AMD/Intel GPUs)
	CUDA         bool   `json:"cuda,omitempty"`         // Enable CUDA (NVIDIA GPUs)
	Intensity    int    `json:"intensity,omitempty"`    // Mining intensity for GPU miners
	CLIArgs      string `json:"cliArgs,omitempty"`      // Additional CLI arguments
}

// Validate checks the Config for common errors and security issues.
// Returns nil if valid, otherwise returns a descriptive error.
func (c *Config) Validate() error {
	// Pool URL validation
	if c.Pool != "" {
		// Block shell metacharacters in pool URL
		if containsShellChars(c.Pool) {
			return fmt.Errorf("pool URL contains invalid characters")
		}
	}

	// Wallet validation (basic alphanumeric + special chars allowed in addresses)
	if c.Wallet != "" {
		if containsShellChars(c.Wallet) {
			return fmt.Errorf("wallet address contains invalid characters")
		}
		// Most wallet addresses are 40-128 chars
		if len(c.Wallet) > 256 {
			return fmt.Errorf("wallet address too long (max 256 chars)")
		}
	}

	// Thread count validation
	if c.Threads < 0 {
		return fmt.Errorf("threads cannot be negative")
	}
	if c.Threads > 1024 {
		return fmt.Errorf("threads value too high (max 1024)")
	}

	// Algorithm validation (alphanumeric, dash, slash)
	if c.Algo != "" {
		if !isValidAlgo(c.Algo) {
			return fmt.Errorf("algorithm name contains invalid characters")
		}
	}

	// Intensity validation
	if c.Intensity < 0 || c.Intensity > 100 {
		return fmt.Errorf("intensity must be between 0 and 100")
	}
	if c.GPUIntensity < 0 || c.GPUIntensity > 100 {
		return fmt.Errorf("GPU intensity must be between 0 and 100")
	}

	// Donate level validation
	if c.DonateLevel < 0 || c.DonateLevel > 100 {
		return fmt.Errorf("donate level must be between 0 and 100")
	}

	// CLIArgs validation - check for shell metacharacters
	if c.CLIArgs != "" {
		if containsShellChars(c.CLIArgs) {
			return fmt.Errorf("CLI arguments contain invalid characters")
		}
		// Limit length to prevent abuse
		if len(c.CLIArgs) > 1024 {
			return fmt.Errorf("CLI arguments too long (max 1024 chars)")
		}
	}

	return nil
}

// containsShellChars checks for shell metacharacters that could enable injection
func containsShellChars(s string) bool {
	dangerous := []string{";", "|", "&", "`", "$", "(", ")", "{", "}", "<", ">", "\n", "\r", "\\", "'", "\"", "!"}
	for _, d := range dangerous {
		if strings.Contains(s, d) {
			return true
		}
	}
	return false
}

// isValidAlgo checks if an algorithm name contains only valid characters
func isValidAlgo(algo string) bool {
	for _, r := range algo {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '/' || r == '_') {
			return false
		}
	}
	return true
}

// PerformanceMetrics represents the performance metrics for a miner.
type PerformanceMetrics struct {
	Hashrate      int                    `json:"hashrate"`
	Shares        int                    `json:"shares"`
	Rejected      int                    `json:"rejected"`
	Uptime        int                    `json:"uptime"`
	LastShare     int64                  `json:"lastShare"`
	Algorithm     string                 `json:"algorithm"`
	AvgDifficulty int                    `json:"avgDifficulty"` // Average difficulty per accepted share (HashesTotal/SharesGood)
	DiffCurrent   int                    `json:"diffCurrent"`   // Current job difficulty from pool
	ExtraData     map[string]interface{} `json:"extraData,omitempty"`
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

// XMRigSummary represents the full JSON response from the XMRig API.
type XMRigSummary struct {
	ID         string `json:"id"`
	WorkerID   string `json:"worker_id"`
	Uptime     int    `json:"uptime"`
	Restricted bool   `json:"restricted"`
	Resources  struct {
		Memory struct {
			Free              int64 `json:"free"`
			Total             int64 `json:"total"`
			ResidentSetMemory int64 `json:"resident_set_memory"`
		} `json:"memory"`
		LoadAverage         []float64 `json:"load_average"`
		HardwareConcurrency int       `json:"hardware_concurrency"`
	} `json:"resources"`
	Features []string `json:"features"`
	Results  struct {
		DiffCurrent int   `json:"diff_current"`
		SharesGood  int   `json:"shares_good"`
		SharesTotal int   `json:"shares_total"`
		AvgTime     int   `json:"avg_time"`
		AvgTimeMS   int   `json:"avg_time_ms"`
		HashesTotal int   `json:"hashes_total"`
		Best        []int `json:"best"`
	} `json:"results"`
	Algo       string `json:"algo"`
	Connection struct {
		Pool           string `json:"pool"`
		IP             string `json:"ip"`
		Uptime         int    `json:"uptime"`
		UptimeMS       int    `json:"uptime_ms"`
		Ping           int    `json:"ping"`
		Failures       int    `json:"failures"`
		TLS            string `json:"tls"`
		TLSFingerprint string `json:"tls-fingerprint"`
		Algo           string `json:"algo"`
		Diff           int    `json:"diff"`
		Accepted       int    `json:"accepted"`
		Rejected       int    `json:"rejected"`
		AvgTime        int    `json:"avg_time"`
		AvgTimeMS      int    `json:"avg_time_ms"`
		HashesTotal    int    `json:"hashes_total"`
	} `json:"connection"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	UA      string `json:"ua"`
	CPU     struct {
		Brand    string   `json:"brand"`
		Family   int      `json:"family"`
		Model    int      `json:"model"`
		Stepping int      `json:"stepping"`
		ProcInfo int      `json:"proc_info"`
		AES      bool     `json:"aes"`
		AVX2     bool     `json:"avx2"`
		X64      bool     `json:"x64"`
		Is64Bit  bool     `json:"64_bit"`
		L2       int      `json:"l2"`
		L3       int      `json:"l3"`
		Cores    int      `json:"cores"`
		Threads  int      `json:"threads"`
		Packages int      `json:"packages"`
		Nodes    int      `json:"nodes"`
		Backend  string   `json:"backend"`
		MSR      string   `json:"msr"`
		Assembly string   `json:"assembly"`
		Arch     string   `json:"arch"`
		Flags    []string `json:"flags"`
	} `json:"cpu"`
	DonateLevel int      `json:"donate_level"`
	Paused      bool     `json:"paused"`
	Algorithms  []string `json:"algorithms"`
	Hashrate    struct {
		Total   []float64 `json:"total"`
		Highest float64   `json:"highest"`
	} `json:"hashrate"`
	Hugepages []int `json:"hugepages"`
}

// AvailableMiner represents a miner that is available for use.
type AvailableMiner struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
