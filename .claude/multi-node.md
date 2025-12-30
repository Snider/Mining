Multi-Node P2P Mining Management Plan

Overview

Add secure peer-to-peer communication between Mining CLI instances, enabling control of remote mining rigs without commercial mining OS
dependencies.

Libraries

- Borg (github.com/Snider/Borg) - Encryption & packaging toolkit
    - pkg/smsg - SMSG encrypted messaging (ChaCha20-Poly1305)
    - pkg/stmf - X25519 keypairs for node identity
    - pkg/tim - Terminal Isolation Matrix for deployment bundles
- Poindexter (github.com/Snider/Poindexter) - KD-tree peer selection
    - Multi-dimensional ranking by ping/hops/geo/score
    - Optimal peer routing

 ---
Architecture Overview

┌─────────────────────────────────────────────────────────────────┐
│                      CONTROLLER NODE                             │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │ NodeManager │  │ PeerRegistry │  │ Poindexter KD-Tree     │  │
│  │ (identity)  │  │ (known peers)│  │ (peer selection)       │  │
│  └──────┬──────┘  └──────┬───────┘  └────────────────────────┘  │
│         │                │                                       │
│  ┌──────┴────────────────┴───────────────────────────────────┐  │
│  │                    MessageRouter                           │  │
│  │  SMSG encrypt/decrypt  |  Command dispatch  |  Response    │  │
│  └──────────────────────────┬────────────────────────────────┘  │
│                             │ TCP/TLS                            │
└─────────────────────────────┼───────────────────────────────────┘
│
┌─────────────────────┼─────────────────────┐
│                     │                     │
▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  WORKER NODE  │     │  WORKER NODE  │     │  WORKER NODE  │
│  rig-alpha    │     │  rig-beta     │     │  rig-gamma    │
│  ────────────│     │  ────────────│     │  ────────────│
│  XMRig       │     │  TT-Miner    │     │  XMRig       │
│  12.5 kH/s   │     │  45.2 MH/s   │     │  11.8 kH/s   │
└───────────────┘     └───────────────┘     └───────────────┘

 ---
Phase 1: Node Identity System

1.1 Data Structures

File: pkg/node/identity.go
type NodeIdentity struct {
ID         string    `json:"id"`          // Derived from public key (first 16 bytes hex)
Name       string    `json:"name"`        // Human-friendly name
PublicKey  string    `json:"publicKey"`   // X25519 base64
CreatedAt  time.Time `json:"createdAt"`
Role       NodeRole  `json:"role"`        // controller | worker
}

type NodeRole string
const (
RoleController NodeRole = "controller"  // Manages remote workers
RoleWorker     NodeRole = "worker"      // Receives commands, runs miners
RoleDual       NodeRole = "dual"        // Both controller AND worker (default)
)

// Dual mode: Node can control remote peers AND run local miners
// - Can receive commands from other controllers
// - Can send commands to worker peers
// - Runs its own miners locally

File: pkg/node/manager.go
type NodeManager struct {
identity   *NodeIdentity
privateKey []byte              // Never serialized to JSON
keyPath    string              // ~/.local/share/lethean-desktop/node/private.key
configPath string              // ~/.config/lethean-desktop/node.json
mu         sync.RWMutex
}

// Key methods:
func NewNodeManager() (*NodeManager, error)          // Load or generate identity
func (n *NodeManager) GenerateIdentity(name string, role NodeRole) error
func (n *NodeManager) GetIdentity() *NodeIdentity
func (n *NodeManager) Sign(data []byte) ([]byte, error)
func (n *NodeManager) DeriveSharedSecret(peerPubKey []byte) ([]byte, error)

1.2 Storage Layout

~/.config/lethean-desktop/
├── node.json                    # Public identity (ID, name, pubkey, role)
└── peers.json                   # Registered peers

~/.local/share/lethean-desktop/node/
└── private.key                  # X25519 private key (0600 permissions)

 ---
Phase 2: Peer Registry

2.1 Data Structures

File: pkg/node/peer.go
type Peer struct {
ID         string    `json:"id"`
Name       string    `json:"name"`
PublicKey  string    `json:"publicKey"`
Address    string    `json:"address"`      // host:port
Role       NodeRole  `json:"role"`
AddedAt    time.Time `json:"addedAt"`
LastSeen   time.Time `json:"lastSeen"`

     // Poindexter metrics (updated dynamically)
     PingMS     float64   `json:"pingMs"`
     Hops       int       `json:"hops"`
     GeoKM      float64   `json:"geoKm"`
     Score      float64   `json:"score"`        // Reliability score 0-100
}

type PeerRegistry struct {
peers    map[string]*Peer
kdTree   *poindexter.KDTree[*Peer]  // For optimal selection
path     string
mu       sync.RWMutex
}

2.2 Key Methods

func (r *PeerRegistry) AddPeer(peer *Peer) error
func (r *PeerRegistry) RemovePeer(id string) error
func (r *PeerRegistry) GetPeer(id string) *Peer
func (r *PeerRegistry) ListPeers() []*Peer
func (r *PeerRegistry) UpdateMetrics(id string, ping, geo float64, hops int)
func (r *PeerRegistry) SelectOptimalPeer() *Peer              // Poindexter query
func (r *PeerRegistry) SelectNearestPeers(n int) []*Peer      // k-NN query

 ---
Phase 3: Message Protocol

3.1 Message Types

File: pkg/node/message.go
type MessageType string
const (
MsgHandshake    MessageType = "handshake"     // Initial key exchange
MsgPing         MessageType = "ping"          // Health check
MsgPong         MessageType = "pong"
MsgGetStats     MessageType = "get_stats"     // Request miner stats
MsgStats        MessageType = "stats"         // Stats response
MsgStartMiner   MessageType = "start_miner"   // Start mining command
MsgStopMiner    MessageType = "stop_miner"    // Stop mining command
MsgDeploy       MessageType = "deploy"        // Deploy config/bundle
MsgDeployAck    MessageType = "deploy_ack"
MsgGetLogs      MessageType = "get_logs"      // Request console logs
MsgLogs         MessageType = "logs"          // Logs response
MsgError        MessageType = "error"
)

type Message struct {
ID        string          `json:"id"`         // UUID
Type      MessageType     `json:"type"`
From      string          `json:"from"`       // Sender node ID
To        string          `json:"to"`         // Recipient node ID
Timestamp time.Time       `json:"ts"`
Payload   json.RawMessage `json:"payload"`
Signature []byte          `json:"sig"`        // Ed25519 signature
}

3.2 Payload Types

// Handshake
type HandshakePayload struct {
Identity  NodeIdentity `json:"identity"`
Challenge []byte       `json:"challenge"`   // Random bytes for auth
}

// Start Miner
type StartMinerPayload struct {
ProfileID string  `json:"profileId"`
Config    *Config `json:"config,omitempty"` // Override profile config
}

// Stats Response
type StatsPayload struct {
Miners []MinerStats `json:"miners"`
}

// Deploy (STIM bundle)
type DeployPayload struct {
BundleType string `json:"type"`      // "profile" | "miner" | "full"
Data       []byte `json:"data"`      // STIM-encrypted bundle
Checksum   string `json:"checksum"`
}

 ---
Phase 4: Transport Layer (WebSocket + SMSG)

4.1 Connection Manager

File: pkg/node/transport.go
type TransportConfig struct {
ListenAddr   string        // ":9091" default
WSPath       string        // "/ws" - WebSocket endpoint path
TLSCertPath  string        // Optional TLS for wss://
TLSKeyPath   string
MaxConns     int
PingInterval time.Duration // WebSocket keepalive
PongTimeout  time.Duration
}

type Transport struct {
config     TransportConfig
server     *http.Server
upgrader   websocket.Upgrader  // gorilla/websocket
conns      map[string]*PeerConnection
node       *NodeManager
registry   *PeerRegistry
handler    MessageHandler
mu         sync.RWMutex
}

type PeerConnection struct {
Peer         *Peer
Conn         *websocket.Conn
SharedSecret []byte          // Derived via X25519 ECDH, used for SMSG
LastActivity time.Time
writeMu      sync.Mutex      // Serialize WebSocket writes
}

4.2 WebSocket Protocol

Client connects: ws://host:9091/ws
wss://host:9091/ws (with TLS)

Each WebSocket message is:
┌────────────────────────────────────────────────────┐
│  Binary frame containing SMSG-encrypted payload    │
│  (JSON Message struct inside after decryption)     │
└────────────────────────────────────────────────────┘

Benefits of WebSocket over raw TCP:
- Better firewall/NAT traversal
- Built-in framing (no need for length prefixes)
- HTTP upgrade allows future reverse-proxy support
- Easy browser integration for web dashboard

4.3 Key Methods

func (t *Transport) Start() error                            // Start WS server
func (t *Transport) Stop() error                             // Graceful shutdown
func (t *Transport) Connect(peer *Peer) (*PeerConnection, error)  // Dial peer
func (t *Transport) Send(peerID string, msg *Message) error  // SMSG encrypt + send
func (t *Transport) Broadcast(msg *Message) error            // Send to all peers
func (t *Transport) OnMessage(handler MessageHandler)        // Register handler

// WebSocket handlers
func (t *Transport) handleWSUpgrade(w http.ResponseWriter, r *http.Request)
func (t *Transport) handleConnection(conn *websocket.Conn)
func (t *Transport) readLoop(pc *PeerConnection)
func (t *Transport) keepalive(pc *PeerConnection)            // Ping/pong

 ---
Phase 5: Command Handlers

5.1 Controller Commands

File: pkg/node/controller.go
type Controller struct {
node       *NodeManager
peers      *PeerRegistry
transport  *Transport
manager    *Manager        // Local miner manager
}

// Remote operations
func (c *Controller) StartRemoteMiner(peerID, profileID string) error
func (c *Controller) StopRemoteMiner(peerID, minerName string) error
func (c *Controller) GetRemoteStats(peerID string) (*StatsPayload, error)
func (c *Controller) GetRemoteLogs(peerID, minerName string, lines int) ([]string, error)
func (c *Controller) DeployProfile(peerID string, profile *MiningProfile) error
func (c *Controller) DeployMinerBundle(peerID string, minerType string) error

// Aggregation
func (c *Controller) GetAllStats() map[string]*StatsPayload
func (c *Controller) GetTotalHashrate() float64

5.2 Worker Handlers

File: pkg/node/worker.go
type Worker struct {
node      *NodeManager
transport *Transport
manager   *Manager
}

func (w *Worker) HandleMessage(msg *Message) (*Message, error)
func (w *Worker) handleGetStats(msg *Message) (*Message, error)
func (w *Worker) handleStartMiner(msg *Message) (*Message, error)
func (w *Worker) handleStopMiner(msg *Message) (*Message, error)
func (w *Worker) handleDeploy(msg *Message) (*Message, error)
func (w *Worker) handleGetLogs(msg *Message) (*Message, error)

 ---
Phase 6: CLI Commands

6.1 Node Management

File: cmd/mining/cmd/node.go
// miner-cli node init --name "rig-alpha" --role worker
// miner-cli node init --name "control-center" --role controller
var nodeInitCmd = &cobra.Command{
Use:   "init",
Short: "Initialize node identity",
}

// miner-cli node info
var nodeInfoCmd = &cobra.Command{
Use:   "info",
Short: "Show node identity and status",
}

// miner-cli node serve --listen :9091
var nodeServeCmd = &cobra.Command{
Use:   "serve",
Short: "Start P2P server for remote connections",
}

6.2 Peer Management

File: cmd/mining/cmd/peer.go
// miner-cli peer add --address 192.168.1.100:9091 --name "rig-alpha"
var peerAddCmd = &cobra.Command{
Use:   "add",
Short: "Add a peer node (initiates handshake)",
}

// miner-cli peer list
var peerListCmd = &cobra.Command{
Use:   "list",
Short: "List registered peers with status",
}

// miner-cli peer remove <peer-id>
var peerRemoveCmd = &cobra.Command{
Use:   "remove",
Short: "Remove a peer from registry",
}

// miner-cli peer ping <peer-id>
var peerPingCmd = &cobra.Command{
Use:   "ping",
Short: "Ping a peer and update metrics",
}

6.3 Remote Operations

File: cmd/mining/cmd/remote.go
// miner-cli remote status [peer-id]
// Shows stats from all peers or specific peer
var remoteStatusCmd = &cobra.Command{
Use:   "status",
Short: "Get mining status from remote peers",
}

// miner-cli remote start <peer-id> --profile <profile-id>
var remoteStartCmd = &cobra.Command{
Use:   "start",
Short: "Start miner on remote peer",
}

// miner-cli remote stop <peer-id> [miner-name]
var remoteStopCmd = &cobra.Command{
Use:   "stop",
Short: "Stop miner on remote peer",
}

// miner-cli remote deploy <peer-id> --profile <profile-id>
// miner-cli remote deploy <peer-id> --miner xmrig
var remoteDeployCmd = &cobra.Command{
Use:   "deploy",
Short: "Deploy config or miner bundle to remote peer",
}

// miner-cli remote logs <peer-id> <miner-name> --lines 100
var remoteLogsCmd = &cobra.Command{
Use:   "logs",
Short: "Get console logs from remote miner",
}

 ---
Phase 7: REST API Extensions

7.1 New Endpoints

File: pkg/mining/service.go (additions)
// Node endpoints
nodeGroup := router.Group(s.namespace + "/node")
nodeGroup.GET("/info", s.handleNodeInfo)
nodeGroup.POST("/init", s.handleNodeInit)

// Peer endpoints
peerGroup := router.Group(s.namespace + "/peers")
peerGroup.GET("", s.handleListPeers)
peerGroup.POST("", s.handleAddPeer)
peerGroup.DELETE("/:id", s.handleRemovePeer)
peerGroup.POST("/:id/ping", s.handlePingPeer)

// Remote operations
remoteGroup := router.Group(s.namespace + "/remote")
remoteGroup.GET("/stats", s.handleRemoteStats)           // All peers
remoteGroup.GET("/:peerId/stats", s.handlePeerStats)     // Single peer
remoteGroup.POST("/:peerId/start", s.handleRemoteStart)
remoteGroup.POST("/:peerId/stop", s.handleRemoteStop)
remoteGroup.POST("/:peerId/deploy", s.handleRemoteDeploy)
remoteGroup.GET("/:peerId/logs/:miner", s.handleRemoteLogs)

 ---
Phase 8: Deployment Bundles (TIM/STIM)

8.1 Bundle Creation

File: pkg/node/bundle.go
type BundleType string
const (
BundleProfile BundleType = "profile"  // Just config
BundleMiner   BundleType = "miner"    // Miner binary + config
BundleFull    BundleType = "full"     // Everything
)

func CreateProfileBundle(profile *MiningProfile) (*tim.TerminalIsolationMatrix, error)
func CreateMinerBundle(minerType string, profile *MiningProfile) (*tim.TerminalIsolationMatrix, error)

// Encrypt for transport
func EncryptBundle(t *tim.TerminalIsolationMatrix, password string) ([]byte, error) {
return t.ToSigil(password)  // Returns STIM-encrypted bytes
}

// Decrypt on receipt
func DecryptBundle(data []byte, password string) (*tim.TerminalIsolationMatrix, error) {
return tim.FromSigil(data, password)
}

 ---
Phase 9: UI Integration

9.1 New UI Pages

File: ui/src/app/pages/nodes/nodes.component.ts
- Show local node identity
- List connected peers with status
- Actions: Add peer, remove peer, ping
- View aggregated stats from all nodes

File: ui/src/app/pages/fleet/fleet.component.ts (or extend Workers)
- Fleet-wide view of all miners across all nodes
- Group by node or show flat list
- Remote start/stop actions
- Deploy profiles to remote nodes

9.2 Sidebar Addition

Add "Nodes" or "Fleet" navigation item to sidebar between Workers and Graphs.

 ---
Implementation Order

Sprint 1: Node Identity & Peer Registry

1. Create pkg/node/identity.go - NodeIdentity, NodeManager
2. Create pkg/node/peer.go - Peer, PeerRegistry
3. Add STMF dependency (github.com/Snider/Borg)
4. Implement key generation and storage
5. Add node init and node info CLI commands

Sprint 2: Transport Layer

1. Create pkg/node/message.go - Message types and payloads
2. Create pkg/node/transport.go - TCP transport with SMSG encryption
3. Implement handshake protocol
4. Add node serve CLI command
5. Add peer add and peer list CLI commands

Sprint 3: Remote Operations

1. Create pkg/node/controller.go - Controller operations
2. Create pkg/node/worker.go - Worker message handlers
3. Integrate with existing Manager for local operations
4. Add remote status/start/stop/logs CLI commands

Sprint 4: Poindexter Integration & Deployment

1. Add Poindexter dependency
2. Integrate KD-tree peer selection
3. Create pkg/node/bundle.go - TIM/STIM bundles
4. Add remote deploy CLI command
5. Add peer metrics (ping, geo, score)

Sprint 5: REST API & UI

1. Add node/peer REST endpoints to service.go
2. Add remote operation REST endpoints
3. Create Nodes UI page
4. Update Workers page for fleet view
5. Add node status to stats panel

 ---
Critical Files

New Files

pkg/node/
├── identity.go     # NodeIdentity, NodeManager
├── peer.go         # Peer, PeerRegistry
├── message.go      # Message types and protocol
├── transport.go    # TCP transport with SMSG
├── controller.go   # Controller operations
├── worker.go       # Worker message handlers
└── bundle.go       # TIM/STIM deployment bundles

cmd/mining/cmd/
├── node.go         # node init/info/serve commands
├── peer.go         # peer add/list/remove/ping commands
└── remote.go       # remote status/start/stop/deploy/logs commands

ui/src/app/pages/nodes/
└── nodes.component.ts  # Node management UI

Modified Files

go.mod                       # Add Borg, Poindexter deps
pkg/mining/service.go        # Add node/peer/remote REST endpoints
pkg/mining/manager.go        # Integrate with node transport
cmd/mining/cmd/root.go       # Register node/peer/remote commands
ui/src/app/components/sidebar/sidebar.component.ts  # Add Nodes nav

 ---
Security Considerations

1. Private key storage: 0600 permissions, never in JSON
2. Shared secrets: Derived per-peer via X25519 ECDH, used for SMSG
3. Message signing: All messages signed with sender's private key
4. TLS option: Support TLS for transport (optional, SMSG provides encryption)
5. Peer verification: Handshake verifies identity before accepting commands
6. Command authorization: Workers only accept commands from registered controllers

 ---
Design Decisions Summary

| Decision       | Choice                   | Rationale                                                              |
|----------------|--------------------------|------------------------------------------------------------------------|
| Discovery      | Manual only              | Simpler, more secure - explicit peer registration                      |
| Transport      | WebSocket + SMSG         | Better firewall traversal, built-in framing, browser-friendly          |
| Node Mode      | Dual (default)           | Maximum flexibility - each node controls remotes AND runs local miners |
| Encryption     | SMSG (ChaCha20-Poly1305) | Uses Borg library, password-derived keys via ECDH                      |
| Identity       | X25519 keypairs (STMF)   | Standard, fast, 32-byte keys                                           |
| Peer Selection | Poindexter KD-tree       | Multi-factor optimization (ping, hops, geo, score)                     |
| Deployment     | TIM/STIM bundles         | Encrypted container bundles for miner+config deployment                |

 ---
Dependencies to Add

// go.mod additions
require (
github.com/Snider/Borg v0.x.x           // SMSG, STMF, TIM encryption
github.com/Snider/Poindexter v0.x.x     // KD-tree peer selection
github.com/gorilla/websocket v1.5.x     // WebSocket transport
)
