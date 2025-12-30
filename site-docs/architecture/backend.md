# Backend Architecture

The Go backend provides miner management, REST API, and P2P networking.

## Package Structure

```
pkg/
├── mining/           # Core mining functionality
│   ├── mining.go     # Interfaces and types
│   ├── miner.go      # BaseMiner shared logic
│   ├── manager.go    # Miner lifecycle management
│   ├── service.go    # REST API endpoints
│   ├── xmrig*.go     # XMRig implementation
│   ├── ttminer*.go   # TT-Miner implementation
│   ├── profile_manager.go
│   └── config_manager.go
├── node/             # P2P networking
│   ├── identity.go   # Node identity (X25519)
│   ├── peer.go       # Peer registry
│   ├── transport.go  # WebSocket transport
│   ├── message.go    # Protocol messages
│   ├── controller.go # Remote operations
│   └── worker.go     # Message handlers
└── database/         # Persistence
    └── sqlite.go     # SQLite operations
```

## Core Interfaces

### Miner Interface

```go
type Miner interface {
    Install() error
    Uninstall() error
    Start(cfg *Config) error
    Stop() error
    GetStats() (*PerformanceMetrics, error)
    GetConfig() *Config
    GetBinaryPath() string
    GetDataPath() string
    GetVersion() (string, error)
    IsInstalled() bool
    IsRunning() bool
    GetMinerType() string
    GetName() string
}
```

### Manager Interface

```go
type ManagerInterface interface {
    StartMiner(minerType string, cfg *Config) (string, error)
    StopMiner(name string) error
    GetMinerStats(name string) (*PerformanceMetrics, error)
    ListMiners() []MinerStatus
    InstallMiner(minerType string) error
    UninstallMiner(minerType string) error
}
```

## REST API Routes

```go
// System
GET  /info                     # System information

// Miners
GET  /miners                   # List running miners
GET  /miners/available         # List installable miners
DELETE /miners/:name           # Stop miner
GET  /miners/:name/stats       # Get miner stats
GET  /miners/:name/logs        # Get miner logs
POST /miners/:name/stdin       # Send stdin command
GET  /miners/:name/hashrate-history

// Installation
POST   /miners/:type/install
DELETE /miners/:type/uninstall

// Profiles
GET    /profiles
POST   /profiles
PUT    /profiles/:id
DELETE /profiles/:id
POST   /profiles/:id/start

// History
GET /history/miners
GET /history/miners/:name
GET /history/miners/:name/hashrate

// P2P
GET  /node/info
GET  /peers
POST /peers
DELETE /peers/:id
POST /peers/:id/ping

// Remote
GET  /remote/stats
GET  /remote/:peerId/stats
POST /remote/:peerId/start
POST /remote/:peerId/stop
```

## Miner Implementations

### XMRig

- Downloads from GitHub releases
- Generates `config.json` for each run
- Polls HTTP API (port 8080) for stats
- Parses stdout for hashrate if API unavailable

### TT-Miner

- Downloads from GitHub releases
- Uses command-line arguments
- Parses stdout for stats (no HTTP API)
- Supports CUDA/OpenCL GPU mining

## Stats Collection

```go
// Every 10 seconds
func (m *Manager) collectStats() {
    for _, miner := range m.miners {
        stats, err := miner.GetStats()
        if err != nil {
            continue
        }

        // Store in memory (high-res, 5 min)
        m.hashrateHistory[name].AddPoint(stats.Hashrate)

        // Store in SQLite (low-res, 30 days)
        m.db.InsertHashrate(name, stats.Hashrate, time.Now())
    }
}
```

## P2P Architecture

### Node Identity

```go
type NodeIdentity struct {
    ID        string    // Derived from public key
    Name      string    // Human-friendly name
    PublicKey string    // X25519 base64
    Role      NodeRole  // controller|worker|dual
}
```

### Message Protocol

```go
type Message struct {
    ID        string          // UUID
    Type      MessageType     // handshake, ping, get_stats, etc.
    From      string          // Sender node ID
    To        string          // Recipient node ID
    Timestamp time.Time
    Payload   json.RawMessage
}
```

### WebSocket Transport

- Listens on port 9091 by default
- Binary frames with JSON messages
- Automatic reconnection handling
- Ping/pong keepalive

## Error Handling

All API endpoints return consistent error format:

```json
{
    "error": "Error message here"
}
```

HTTP status codes:
- `200` - Success
- `400` - Bad request
- `404` - Not found
- `500` - Server error
