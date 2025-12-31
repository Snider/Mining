# Architecture

This document provides a detailed overview of the Mining Platform architecture, design decisions, and component interactions.

## High-Level Architecture

The Mining Platform follows a modular, layered architecture:

```
┌─────────────────────────────────────────────────────────┐
│                    User Interfaces                       │
│  ┌──────────┐  ┌───────────────┐  ┌─────────────────┐  │
│  │   CLI    │  │ Web Dashboard │  │ Desktop App     │  │
│  │  (Cobra) │  │   (Angular)   │  │  (Wails+Angular)│  │
│  └────┬─────┘  └───────┬───────┘  └────────┬────────┘  │
└───────┼────────────────┼───────────────────┼───────────┘
        │                │                   │
        └────────────────┼───────────────────┘
                         │
┌────────────────────────┼───────────────────────────────┐
│                 REST API Layer (Gin)                    │
│                  /api/v1/mining/*                       │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────┼───────────────────────────────┐
│              Core Business Logic (pkg/mining)           │
│  ┌──────────────┐  ┌─────────────┐  ┌───────────────┐ │
│  │   Manager    │  │   Miner     │  │   Profile     │ │
│  │  Interface   │  │ Implemen-   │  │   Manager     │ │
│  │              │  │   tations   │  │               │ │
│  └──────┬───────┘  └──────┬──────┘  └───────┬───────┘ │
└─────────┼──────────────────┼─────────────────┼─────────┘
          │                  │                 │
┌─────────┼──────────────────┼─────────────────┼─────────┐
│    System Layer (OS, Filesystem, Processes)             │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Mining Software (XMRig, T-Rex, lolMiner, etc.)  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Core Components

### Manager Interface

The `ManagerInterface` is the central abstraction for miner lifecycle management.

**Location:** `pkg/mining/manager.go`

**Purpose:**
- Provides a contract for miner operations
- Enables testing through mocking
- Supports multiple miner implementations
- Manages running miners in-memory

**Interface Definition:**

```go
type ManagerInterface interface {
    StartMiner(minerType string, config *Config) (Miner, error)
    StopMiner(name string) error
    GetMiner(name string) (Miner, error)
    ListMiners() []Miner
    ListAvailableMiners() []AvailableMiner
    GetMinerHashrateHistory(name string) ([]HashratePoint, error)
    Stop()
}
```

**Implementation Details:**
- Maintains a `map[string]Miner` for running miners
- Automatically collects statistics every 10 seconds
- Supports autostart from configuration
- Thread-safe with mutex locks

### Miner Interface

The `Miner` interface defines the contract for all miner implementations.

**Location:** `pkg/mining/mining.go`

**Interface Definition:**

```go
type Miner interface {
    GetName() string
    GetStats() (*PerformanceMetrics, error)
    Stop() error
    IsRunning() bool
    GetConfig() *Config
    GetHashrateHistory() []HashratePoint
}
```

**Implementations:**
- **XMRigMiner**: CPU/GPU mining for RandomX and CryptoNight
- **TRexMiner**: NVIDIA GPU mining for KawPow, Ethash (future)
- **LolMiner**: AMD/NVIDIA mining for Ethash, Beam (future)

### BaseMiner

Provides shared functionality for all miner implementations.

**Location:** `pkg/mining/miner.go`

**Features:**
- Binary discovery and installation
- Archive extraction (tar.gz, tar.xz, zip)
- Download from URLs with progress tracking
- Hashrate history management
- XDG directory compliance

**Key Methods:**

```go
func (m *BaseMiner) InstallFromURL(url string) error
func (m *BaseMiner) FindBinary(name string) (string, error)
func (m *BaseMiner) AddHashratePoint(hashrate float64)
func (m *BaseMiner) GetHashrateHistory() []HashratePoint
```

**Hashrate History:**
- High-resolution: 10-second intervals, 5-minute retention
- Low-resolution: 1-minute averages, 24-hour retention
- Automatically manages data retention

### XMRig Implementation

Complete implementation for XMRig miner.

**Files:**
- `pkg/mining/xmrig.go`: Core implementation
- `pkg/mining/xmrig_start.go`: Startup logic
- `pkg/mining/xmrig_stats.go`: Statistics parsing

**Architecture:**

```
┌─────────────────────────────────────────────┐
│           XMRigMiner                        │
│  ┌─────────────────────────────────────┐   │
│  │        BaseMiner (embedded)         │   │
│  │  - Binary management                │   │
│  │  - Hashrate history                 │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  Start() → Generate config.json            │
│         → Execute xmrig binary             │
│         → Capture stdout/stderr            │
│         → Monitor process                  │
│                                             │
│  GetStats() → Poll HTTP API                │
│            → Parse JSON response           │
│            → Return PerformanceMetrics     │
│                                             │
│  Stop() → Send SIGTERM                     │
│        → Wait for graceful shutdown        │
│        → Force kill if needed              │
└─────────────────────────────────────────────┘
```

**Configuration Generation:**
- Creates JSON config file
- Supports CPU and GPU mining
- Handles pool authentication
- Manages algorithm selection
- Configures API endpoint for stats

**Statistics Collection:**
- Polls XMRig HTTP API (default: `http://127.0.0.1:44321/1/summary`)
- Parses JSON response
- Extracts hashrate, shares, connection info
- Updates hashrate history

### Service Layer (REST API)

Exposes the Manager functionality via HTTP endpoints.

**Location:** `pkg/mining/service.go`

**Framework:** Gin Web Framework

**Features:**
- RESTful API design
- Swagger documentation
- CORS support
- JSON request/response
- Error handling middleware
- Route grouping

**Route Organization:**

```
/api/v1/mining
├── /info                    # GET - System info
├── /doctor                  # POST - Diagnostics
├── /update                  # POST - Check updates
├── /miners
│   ├── /                    # GET - List miners
│   ├── /available           # GET - Available types
│   ├── /:miner_type         # POST - Start miner
│   ├── /:miner_name         # DELETE - Stop miner
│   ├── /:miner_name/stats   # GET - Get statistics
│   ├── /:miner_type/install # POST - Install miner
│   └── /:miner_type/uninstall # DELETE - Uninstall
└── /profiles
    ├── /                    # GET - List profiles
    ├── /                    # POST - Create profile
    ├── /:id                 # GET - Get profile
    ├── /:id                 # PUT - Update profile
    ├── /:id                 # DELETE - Delete profile
    └── /:id/start           # POST - Start from profile
```

**Middleware Stack:**
1. Logger
2. Recovery (panic handler)
3. CORS
4. Request validation
5. Response formatter

### Profile Manager

Manages saved mining configurations.

**Location:** `pkg/mining/profile_manager.go`

**Storage:** JSON file at `~/.config/lethean-desktop/mining_profiles.json`

**Data Structure:**

```go
type MiningProfile struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    MinerType   string    `json:"minerType"`
    Config      *Config   `json:"config"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

**Features:**
- CRUD operations for profiles
- UUID-based profile IDs
- Atomic file writes
- Import/export support
- Validation

### Config Manager

Handles autostart and last-used configurations.

**Location:** `pkg/mining/config_manager.go`

**Storage:** JSON file at `~/.config/lethean-desktop/mining_config.json`

**Features:**
- Autostart configuration
- Last-used miner configs
- Preference storage
- Default settings

## Data Flow

### Starting a Miner

```
User Request
    │
    ├─→ CLI: miner-ctrl start xmrig --config config.json
    ├─→ API: POST /api/v1/mining/miners/xmrig
    └─→ Desktop: profileManager.start(profileId)
    │
    ▼
Service Layer (service.go)
    │
    ├─→ Validate request
    ├─→ Parse configuration
    └─→ Call manager.StartMiner()
    │
    ▼
Manager (manager.go)
    │
    ├─→ Check if miner already running
    ├─→ Validate configuration
    └─→ Create miner instance
    │
    ▼
Miner Implementation (xmrig.go)
    │
    ├─→ Generate config.json
    ├─→ Find/verify binary
    ├─→ Execute miner process
    ├─→ Capture output streams
    └─→ Start statistics collection
    │
    ▼
Manager
    │
    ├─→ Store miner in running map
    └─→ Return miner instance
    │
    ▼
Service Layer
    │
    ├─→ Format response
    └─→ Return HTTP 200/201
```

### Collecting Statistics

```
Background goroutine (every 10s)
    │
    ▼
For each running miner:
    │
    ├─→ miner.GetStats()
    │       │
    │       ├─→ Poll HTTP API
    │       ├─→ Parse JSON
    │       └─→ Return PerformanceMetrics
    │
    ├─→ Extract hashrate
    └─→ miner.AddHashratePoint(hashrate)
            │
            ├─→ Store in high-res buffer
            ├─→ Update low-res averages
            └─→ Prune old data
```

### Retrieving Statistics (API Request)

```
Client: GET /api/v1/mining/miners/xmrig/stats
    │
    ▼
Service Layer
    │
    ├─→ Extract miner name
    └─→ Call manager.GetMiner(name)
    │
    ▼
Manager
    │
    ├─→ Lookup in miners map
    └─→ Return miner instance
    │
    ▼
Service Layer
    │
    ├─→ Call miner.GetStats()
    ├─→ Format response
    └─→ Return JSON
```

## Frontend Architecture (Angular)

### Component Hierarchy

```
AppComponent
├── DashboardPage
│   ├── MinerStatusCard
│   ├── HashrateCard
│   ├── SharesCard
│   └── EarningsCard
├── MinersPage
│   ├── RunningMinersList
│   │   └── MinerCard
│   └── AvailableMinersList
│       └── MinerInstallCard
├── ProfilesPage
│   ├── ProfileList
│   │   └── ProfileCard
│   └── ProfileEditor
├── StatisticsPage
│   ├── HashrateChart
│   ├── SharesChart
│   └── TimeRangeSelector
├── PoolsPage
│   ├── RecommendedPools
│   └── CustomPoolForm
├── AdminPage
│   ├── SystemInfo
│   ├── MinerManagement
│   └── DiagnosticsPanel
└── SettingsPage
    ├── GeneralSettings
    ├── NotificationSettings
    └── AdvancedSettings
```

### Services

**MinerService**
- API communication
- Miner lifecycle management
- Statistics fetching

**ProfileService**
- Profile CRUD operations
- Profile storage
- Import/export

**WebSocketService**
- Real-time updates
- Event notifications
- Connection management

**ThemeService**
- Theme switching
- Preference persistence

### State Management

The application uses RxJS for state management:

- Services emit observables
- Components subscribe to updates
- Automatic cleanup on destroy
- Centralized error handling

## Desktop Application (Wails)

### Architecture

```
┌─────────────────────────────────────────┐
│         Go Backend (main.go)            │
│  ┌───────────────────────────────────┐  │
│  │     MiningService                 │  │
│  │  - Wraps pkg/mining.Manager       │  │
│  │  - Exposes methods to frontend    │  │
│  └────────────┬──────────────────────┘  │
└───────────────┼─────────────────────────┘
                │ Wails Bindings
┌───────────────┼─────────────────────────┐
│         TypeScript Frontend             │
│  ┌────────────┴──────────────────────┐  │
│  │   Angular Application             │  │
│  │ (Embedded from ui/dist/browser)   │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

**MiningService** (`miningservice.go`):
- Binds Go methods to frontend
- Handles lifecycle events
- Manages application state
- Provides system tray integration

**Auto-generated Bindings** (`frontend/bindings/`):
- TypeScript definitions
- Type-safe Go method calls
- Event system integration

## Modified XMRig Core

### OpenCL Backend

**Location:** `miner/core/src/backend/opencl/`

**Supported Algorithms:**
- ETChash (Ethereum Classic)
- ProgPowZ (Zano)
- KawPow (Ravencoin)

**Key Files:**
- `cl/etchash/`: OpenCL kernels for ETChash
- `cl/progpowz/`: OpenCL kernels for ProgPowZ
- `runners/OclEtchashRunner.*`: ETChash GPU runner
- `runners/OclProgPowZRunner.*`: ProgPowZ GPU runner

### Algorithm Implementations

**Location:** `miner/core/src/crypto/`

**ETChash:**
- `ETCCache.cpp/h`: DAG cache management
- Ethash variant optimized for ETC

**ProgPowZ:**
- Custom ProgPow variant for Zano
- Period-based algorithm rotation

### Build System

CMake-based build with conditional compilation:

```cmake
option(WITH_OPENCL "Enable OpenCL backend" ON)
option(WITH_CUDA "Enable CUDA backend" ON)
```

Automatically detects:
- OpenCL SDK
- CUDA Toolkit
- GPU capabilities

## Security Considerations

### API Security

- No authentication by default (local use)
- Consider reverse proxy for production
- CORS enabled for web component
- Input validation on all endpoints

### File System

- XDG Base Directory compliance
- Restricted file permissions (0644 for config, 0755 for binaries)
- Atomic file writes for configs
- Safe path handling (no path traversal)

### Process Management

- Graceful shutdown (SIGTERM → SIGKILL)
- Process isolation
- Resource limits (if configured)
- Log rotation

## Performance Optimizations

### Hashrate History

- Two-tier storage (high-res + low-res)
- Automatic data pruning
- In-memory ring buffers
- Minimal memory footprint

### Statistics Collection

- Background goroutine (non-blocking)
- Cached HTTP requests
- JSON parsing optimization
- Error resilience

### Frontend

- Lazy loading of routes
- Virtual scrolling for large lists
- Chart data decimation
- Debounced API calls

## Extensibility

### Adding New Miners

1. Implement `Miner` interface
2. Extend `BaseMiner` for common functionality
3. Register in `Manager`
4. Add API endpoints if needed

### Adding New Algorithms

1. Add algorithm support to miner core
2. Update configuration structs
3. Add validation rules
4. Update UI selectors

### Adding New Frontends

The REST API is frontend-agnostic:
- Mobile apps (React Native, Flutter)
- CLI tools
- Third-party integrations
- Monitoring dashboards

## Deployment Patterns

### Single User (Local)

```
User Machine
├── miner-ctrl (CLI/API server)
├── Browser (accessing localhost:8080)
└── Mining software (XMRig, etc.)
```

### Multi-User (Server)

```
Server
├── miner-ctrl (API server on 0.0.0.0:8080)
└── Mining software

Reverse Proxy (nginx)
├── HTTPS termination
├── Authentication
└── Rate limiting

Clients
├── Web browsers
├── Mobile apps
└── API clients
```

### Desktop (Standalone)

```
Single Binary (Wails)
├── Embedded API server
├── Embedded frontend
└── Integrated mining software
```

## Next Steps

- Review [Development Guide](index.md) for setup instructions
- Read [Contributing Guidelines](contributing.md) for contribution process
- See [API Documentation](../api/index.md) for endpoint details
