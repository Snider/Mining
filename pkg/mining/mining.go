package mining

import (
	"os/exec"
	"sync"
	"time"
)

const (
	// HighResolutionDuration is the duration for which hashrate data is kept at high resolution (10s intervals)
	HighResolutionDuration = 5 * time.Minute
	// HighResolutionInterval is the interval at which hashrate data is collected for high resolution
	HighResolutionInterval = 10 * time.Second
	// LowResolutionInterval is the interval for aggregated hashrate data (1m averages)
	LowResolutionInterval = 1 * time.Minute
	// LowResHistoryRetention is the duration for which low-resolution hashrate data is retained
	LowResHistoryRetention = 24 * time.Hour // Example: keep 24 hours of 1-minute averages
)

// Miner defines the standard interface for a cryptocurrency miner.
// This interface abstracts the core functionalities of a miner, such as installation,
// starting, stopping, and statistics retrieval, allowing for different miner
// implementations to be used interchangeably.
type Miner interface {
	// Install handles the setup and installation of the miner software.
	// This may include downloading binaries, creating configuration files,
	// and setting up necessary permissions.
	Install() error

	// Uninstall removes the miner software and any related configuration files.
	Uninstall() error
	Start(config *Config) error
	Stop() error
	GetStats() (*PerformanceMetrics, error)
	GetName() string
	GetPath() string
	GetBinaryPath() string

	// CheckInstallation verifies if the miner is installed correctly and returns
	// details about the installation, such as the version and path.
	CheckInstallation() (*InstallationDetails, error)

	// GetLatestVersion retrieves the latest available version of the miner software.
	GetLatestVersion() (string, error)

	// GetHashrateHistory returns the recent hashrate history of the miner.
	GetHashrateHistory() []HashratePoint

	// AddHashratePoint adds a new hashrate data point to the miner's history.
	AddHashratePoint(point HashratePoint)

	// ReduceHashrateHistory processes the raw hashrate data, potentially
	// aggregating high-resolution data into a lower-resolution format for
	// long-term storage.
	ReduceHashrateHistory(now time.Time)
}

// InstallationDetails contains information about an installed miner.
// It provides a standard structure for reporting the status of a miner's
// installation, including whether it's present, its version, and its location.
type InstallationDetails struct {
	// IsInstalled is true if the miner is installed, false otherwise.
	IsInstalled bool `json:"is_installed"`
	// Version is the detected version of the installed miner.
	Version string `json:"version"`
	// Path is the installation path of the miner.
	Path string `json:"path"`
	// MinerBinary is the name of the miner's executable file.
	MinerBinary string `json:"miner_binary"`
}

// SystemInfo provides general system and miner installation information.
// This struct aggregates various details about the system's environment,
// such as operating system, architecture, and available resources, as well
// as information about installed miners.
type SystemInfo struct {
	// Timestamp is the time when the system information was collected.
	Timestamp time.Time `json:"timestamp"`
	// OS is the operating system of the host.
	OS string `json:"os"`
	// Architecture is the system's hardware architecture (e.g., amd64, arm64).
	Architecture string `json:"architecture"`
	// GoVersion is the version of the Go runtime.
	GoVersion string `json:"go_version"`
	// AvailableCPUCores is the number of available CPU cores.
	AvailableCPUCores int `json:"available_cpu_cores"`
	// TotalSystemRAMGB is the total system RAM in gigabytes.
	TotalSystemRAMGB float64 `json:"total_system_ram_gb"`
	// InstalledMinersInfo is a slice containing details of all installed miners.
	InstalledMinersInfo []*InstallationDetails `json:"installed_miners_info"`
}

// Config represents the configuration for a miner.
// This struct includes general mining parameters as well as specific options
// for different miner implementations like XMRig. It is designed to be

// flexible and comprehensive, covering a wide range of settings from network
// and CPU configurations to logging and miscellaneous options.
//
// Example:
//
//	// Create a new configuration for the XMRig miner
//	config := &mining.Config{
//		Miner:   "xmrig",
//		Pool:    "your-pool-address",
//		Wallet:  "your-wallet-address",
//		Threads: 4,
//		TLS:     true,
//	}
type Config struct {
	// Miner is the name of the miner to be used (e.g., "xmrig").
	Miner string `json:"miner"`
	// Pool is the address of the mining pool.
	Pool string `json:"pool"`
	// Wallet is the user's wallet address for receiving mining rewards.
	Wallet string `json:"wallet"`
	// Threads is the number of CPU threads to be used for mining.
	Threads int `json:"threads"`
	// TLS indicates whether to use a secure TLS connection to the pool.
	TLS bool `json:"tls"`
	// HugePages enables or disables the use of huge pages for performance optimization.
	HugePages bool `json:"hugePages"`

	// Network options
	// Algo specifies the mining algorithm to be used.
	Algo string `json:"algo,omitempty"`
	// Coin specifies the cryptocurrency to be mined.
	Coin string `json:"coin,omitempty"`
	// Password is the pool password.
	Password string `json:"password,omitempty"`
	// UserPass is the username and password for the pool.
	UserPass string `json:"userPass,omitempty"`
	// Proxy is the address of a proxy to be used for the connection.
	Proxy string `json:"proxy,omitempty"`
	// Keepalive enables or disables the TCP keepalive feature.
	Keepalive bool `json:"keepalive,omitempty"`
	// Nicehash enables or disables Nicehash support.
	Nicehash bool `json:"nicehash,omitempty"`
	// RigID is the identifier of the mining rig.
	RigID string `json:"rigId,omitempty"`
	// TLSSingerprint is the TLS fingerprint of the pool.
	TLSSingerprint string `json:"tlsFingerprint,omitempty"`
	// Retries is the number of times to retry connecting to the pool.
	Retries int `json:"retries,omitempty"`
	// RetryPause is the pause in seconds between connection retries.
	RetryPause int `json:"retryPause,omitempty"`
	// UserAgent is the user agent string to be used for the connection.
	UserAgent string `json:"userAgent,omitempty"`
	// DonateLevel is the donation level to the miner developers.
	DonateLevel int `json:"donateLevel,omitempty"`
	// DonateOverProxy enables or disables donation over a proxy.
	DonateOverProxy bool `json:"donateOverProxy,omitempty"`

	// CPU backend options
	// NoCPU disables the CPU backend.
	NoCPU bool `json:"noCpu,omitempty"`
	// CPUAffinity sets the CPU affinity for the miner.
	CPUAffinity string `json:"cpuAffinity,omitempty"`
	// AV is the algorithm variation.
	AV int `json:"av,omitempty"`
	// CPUPriority is the CPU priority for the miner.
	CPUPriority int `json:"cpuPriority,omitempty"`
	// CPUMaxThreadsHint is the maximum number of threads hint for the CPU.
	CPUMaxThreadsHint int `json:"cpuMaxThreadsHint,omitempty"`
	// CPUMemoryPool is the CPU memory pool size.
	CPUMemoryPool int `json:"cpuMemoryPool,omitempty"`
	// CPUNoYield enables or disables CPU yield.
	CPUNoYield bool `json:"cpuNoYield,omitempty"`
	// HugepageSize is the size of huge pages in kilobytes.
	HugepageSize int `json:"hugepageSize,omitempty"`
	// HugePagesJIT enables or disables huge pages for JIT compiled code.
	HugePagesJIT bool `json:"hugePagesJIT,omitempty"`
	// ASM enables or disables the ASM compiler.
	ASM string `json:"asm,omitempty"`
	// Argon2Impl is the Argon2 implementation.
	Argon2Impl string `json:"argon2Impl,omitempty"`
	// RandomXInit is the RandomX initialization value.
	RandomXInit int `json:"randomXInit,omitempty"`
	// RandomXNoNUMA enables or disables NUMA support for RandomX.
	RandomXNoNUMA bool `json:"randomXNoNuma,omitempty"`
	// RandomXMode is the RandomX mode.
	RandomXMode string `json:"randomXMode,omitempty"`
	// RandomX1GBPages enables or disables 1GB pages for RandomX.
	RandomX1GBPages bool `json:"randomX1GBPages,omitempty"`
	// RandomXWrmsr is the RandomX MSR value.
	RandomXWrmsr string `json:"randomXWrmsr,omitempty"`
	// RandomXNoRdmsr enables or disables MSR reading for RandomX.
	RandomXNoRdmsr bool `json:"randomXNoRdmsr,omitempty"`
	// RandomXCacheQoS enables or disables QoS for the RandomX cache.
	RandomXCacheQoS bool `json:"randomXCacheQoS,omitempty"`

	// API options (can be overridden or supplemented here)
	// APIWorkerID is the worker ID for the API.
	APIWorkerID string `json:"apiWorkerId,omitempty"`
	// APIID is the ID for the API.
	APIID string `json:"apiId,omitempty"`
	// HTTPHost is the host for the HTTP API.
	HTTPHost string `json:"httpHost,omitempty"`
	// HTTPPort is the port for the HTTP API.
	HTTPPort int `json:"httpPort,omitempty"`
	// HTTPAccessToken is the access token for the HTTP API.
	HTTPAccessToken string `json:"httpAccessToken,omitempty"`
	// HTTPNoRestricted enables or disables restricted access to the HTTP API.
	HTTPNoRestricted bool `json:"httpNoRestricted,omitempty"`

	// Logging options
	// Syslog enables or disables logging to the system log.
	Syslog bool `json:"syslog,omitempty"`
	// LogFile is the path to the log file.
	LogFile string `json:"logFile,omitempty"`
	// PrintTime is the interval in seconds for printing performance metrics.
	PrintTime int `json:"printTime,omitempty"`
	// HealthPrintTime is the interval in seconds for printing health metrics.
	HealthPrintTime int `json:"healthPrintTime,omitempty"`
	// NoColor disables color output in the logs.
	NoColor bool `json:"noColor,omitempty"`
	// Verbose enables verbose logging.
	Verbose bool `json:"verbose,omitempty"`
	// LogOutput enables or disables logging of stdout/stderr.
	LogOutput bool `json:"logOutput,omitempty"`

	// Misc options
	// Background runs the miner in the background.
	Background bool `json:"background,omitempty"`
	// Title sets the title of the miner window.
	Title string `json:"title,omitempty"`
	// NoTitle disables the miner window title.
	NoTitle bool `json:"noTitle,omitempty"`
	// PauseOnBattery pauses the miner when the system is on battery power.
	PauseOnBattery bool `json:"pauseOnBattery,omitempty"`
	// PauseOnActive pauses the miner when the user is active.
	PauseOnActive int `json:"pauseOnActive,omitempty"`
	// Stress enables stress testing mode.
	Stress bool `json:"stress,omitempty"`
	// Bench enables benchmark mode.
	Bench string `json:"bench,omitempty"`
	// Submit enables or disables submitting shares.
	Submit bool `json:"submit,omitempty"`
	// Verify enables or disables share verification.
	Verify string `json:"verify,omitempty"`
	// Seed is the seed for the random number generator.
	Seed string `json:"seed,omitempty"`
	// Hash is the hash for the random number generator.
	Hash string `json:"hash,omitempty"`
	// NoDMI disables DMI/SMBIOS probing.
	NoDMI bool `json:"noDMI,omitempty"`
}

// PerformanceMetrics represents the performance metrics for a miner.
// This struct provides a standardized way to report key performance indicators
// such as hashrate, shares, and uptime, allowing for consistent monitoring
// and comparison across different miners.
type PerformanceMetrics struct {
	// Hashrate is the current hashrate of the miner in H/s.
	Hashrate int `json:"hashrate"`
	// Shares is the number of shares submitted by the miner.
	Shares int `json:"shares"`
	// Rejected is the number of rejected shares.
	Rejected int `json:"rejected"`
	// Uptime is the duration the miner has been running, in seconds.
	Uptime int `json:"uptime"`
	// LastShare is the timestamp of the last submitted share.
	LastShare int64 `json:"lastShare"`
	// Algorithm is the mining algorithm currently in use.
	Algorithm string `json:"algorithm"`
	// ExtraData provides a flexible way to include additional, miner-specific
	// performance data that is not covered by the standard fields.
	ExtraData map[string]interface{} `json:"extraData,omitempty"`
}

// History represents the historical performance data for a miner.
// It contains a collection of performance metrics snapshots, allowing for
// the tracking of a miner's performance over time.
type History struct {
	// Miner is the name of the miner.
	Miner string `json:"miner"`
	// Stats is a slice of performance metrics, representing the historical data.
	Stats []PerformanceMetrics `json:"stats"`
	// Updated is the timestamp of the last update to the history.
	Updated int64 `json:"updated"`
}

// HashratePoint represents a single hashrate measurement at a specific time.
// This struct is used to build a time-series history of a miner's hashrate,
// which is essential for performance analysis and visualization.
type HashratePoint struct {
	// Timestamp is the time at which the hashrate was measured.
	Timestamp time.Time `json:"timestamp"`
	// Hashrate is the measured hashrate in H/s.
	Hashrate int `json:"hashrate"`
}

// XMRigMiner represents an XMRig miner, encapsulating its configuration,
// state, and operational details. This struct provides a comprehensive
// representation of an XMRig miner instance, including its identity,
// connection details, and performance history.
type XMRigMiner struct {
	// Name is the name of the miner.
	Name string `json:"name"`
	// Version is the version of the XMRig miner.
	Version string `json:"version"`
	// URL is the download URL for the XMRig miner.
	URL string `json:"url"`
	// Path is the installation path of the miner.
	Path string `json:"path"`
	// MinerBinary is the full path to the miner's executable file.
	MinerBinary string `json:"miner_binary"`
	// Running indicates whether the miner is currently running.
	Running bool `json:"running"`
	// LastHeartbeat is the timestamp of the last heartbeat from the miner.
	LastHeartbeat int64 `json:"lastHeartbeat"`
	// ConfigPath is the path to the miner's configuration file.
	ConfigPath string `json:"configPath"`
	// API provides access to the miner's API for statistics and control.
	API *API `json:"api"`
	// mu is a mutex to protect against concurrent access to the miner's state.
	mu sync.Mutex
	// cmd is the command used to execute the miner process.
	cmd *exec.Cmd `json:"-"`
	// HashrateHistory is a slice of high-resolution hashrate data points.
	HashrateHistory []HashratePoint `json:"hashrateHistory"`
	// LowResHashrateHistory is a slice of low-resolution hashrate data points.
	LowResHashrateHistory []HashratePoint `json:"lowResHashrateHistory"`
	// LastLowResAggregation is the timestamp of the last low-resolution aggregation.
	LastLowResAggregation time.Time `json:"-"`
}

// API represents the XMRig API configuration.
// It specifies the details needed to connect to the miner's API,
// enabling programmatic monitoring and control.
type API struct {
	// Enabled indicates whether the API is enabled.
	Enabled bool `json:"enabled"`
	// ListenHost is the host on which the API is listening.
	ListenHost string `json:"listenHost"`
	// ListenPort is the port on which the API is listening.
	ListenPort int `json:"listenPort"`
}

// XMRigSummary represents the summary of an XMRig miner's performance,
// as retrieved from its API. This struct provides a structured way to
// access key performance indicators from the miner's API.
type XMRigSummary struct {
	// Hashrate contains the hashrate data from the API.
	Hashrate struct {
		Total []float64 `json:"total"`
	} `json:"hashrate"`
	// Results contains the share statistics from the API.
	Results struct {
		SharesGood  uint64 `json:"shares_good"`
		SharesTotal uint64 `json:"shares_total"`
	} `json:"results"`
	// Uptime is the duration the miner has been running, in seconds.
	Uptime uint64 `json:"uptime"`
	// Algorithm is the mining algorithm currently in use.
	Algorithm string `json:"algorithm"`
}

// AvailableMiner represents a miner that is available for use.
// It provides a simple way to list and describe the miners that can be
// started and managed by the system.
type AvailableMiner struct {
	// Name is the name of the available miner.
	Name string `json:"name"`
	// Description is a brief description of the miner.
	Description string `json:"description"`
}
