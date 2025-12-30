# P2P Multi-Node

Control multiple mining rigs from a single dashboard using encrypted peer-to-peer communication.

![Nodes Page](../assets/screenshots/nodes.png)

## Overview

The P2P system allows you to:

- **Control remote rigs** without cloud services
- **Aggregate statistics** from all nodes
- **Deploy configurations** to remote workers
- **Secure communication** via encrypted WebSockets

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      CONTROLLER NODE                             │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │ NodeManager │  │ PeerRegistry │  │     Dashboard UI       │  │
│  │ (identity)  │  │ (known peers)│  │                        │  │
│  └──────┬──────┘  └──────┬───────┘  └────────────────────────┘  │
│         │                │                                       │
│  ┌──────┴────────────────┴───────────────────────────────────┐  │
│  │                    WebSocket Transport                     │  │
│  │       SMSG Encryption  |  Message Routing                 │  │
│  └──────────────────────────┬────────────────────────────────┘  │
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
```

## Node Roles

| Role | Description |
|------|-------------|
| **Controller** | Manages remote workers, sends commands |
| **Worker** | Receives commands, runs miners |
| **Dual** | Both controller and worker (default) |

## Setting Up

### 1. Initialize Node Identity

On each machine:

```bash
# Controller node
./miner-ctrl node init --name "control-center" --role controller

# Worker nodes
./miner-ctrl node init --name "rig-alpha" --role worker
```

### 2. Start P2P Server

```bash
# Start with P2P enabled (default port 9091)
./miner-ctrl node serve --listen :9091
```

### 3. Add Peers

From the controller, add worker nodes:

```bash
./miner-ctrl peer add --address 192.168.1.100:9091 --name "rig-alpha"
```

Or via the UI:
1. Go to **Nodes** page
2. Click **Add Peer**
3. Enter the worker's address and name

## Node Identity

Each node has a unique identity:

- **Node ID** - Derived from public key (16 hex characters)
- **Public Key** - X25519 key for encryption
- **Name** - Human-readable name

Identity is stored in:
```
~/.config/lethean-desktop/node.json
~/.local/share/lethean-desktop/node/private.key
```

## Peer Management

### Viewing Peers

The Nodes page shows all registered peers with:

| Column | Description |
|--------|-------------|
| **Peer** | Name and online indicator |
| **Address** | IP:Port |
| **Role** | worker/controller/dual |
| **Ping** | Latency in milliseconds |
| **Score** | Reliability score (0-100) |
| **Last Seen** | Time since last communication |

### Peer Actions

| Action | Description |
|--------|-------------|
| **Ping** | Test connectivity and update metrics |
| **View Stats** | Show miner stats from this peer |
| **Remove** | Delete peer from registry |

## Remote Operations

### Get Remote Stats

```bash
./miner-ctrl remote status rig-alpha
```

### Start Remote Miner

```bash
./miner-ctrl remote start rig-alpha --profile my-profile
```

### Stop Remote Miner

```bash
./miner-ctrl remote stop rig-alpha xmrig-123
```

### Get Remote Logs

```bash
./miner-ctrl remote logs rig-alpha xmrig-123 --lines 100
```

## Security

### Encryption

All communication is encrypted using:

- **X25519** - Key exchange
- **ChaCha20-Poly1305** - Message encryption (SMSG)
- **Message signing** - Ed25519 signatures

### Authentication

- Handshake verifies node identity
- Only registered peers can communicate
- No anonymous connections

### Private Key Protection

- Stored with 0600 permissions
- Never transmitted over network
- Auto-generated on first run

## API Endpoints

### Node Management
```
GET  /api/v1/mining/node/info     # Get local node info
POST /api/v1/mining/node/init     # Initialize node identity
```

### Peer Management
```
GET    /api/v1/mining/peers           # List all peers
POST   /api/v1/mining/peers           # Add a peer
DELETE /api/v1/mining/peers/{id}      # Remove a peer
POST   /api/v1/mining/peers/{id}/ping # Ping a peer
```

### Remote Operations
```
GET  /api/v1/mining/remote/stats              # All peers stats
GET  /api/v1/mining/remote/{peerId}/stats     # Single peer stats
POST /api/v1/mining/remote/{peerId}/start     # Start remote miner
POST /api/v1/mining/remote/{peerId}/stop      # Stop remote miner
GET  /api/v1/mining/remote/{peerId}/logs/{miner} # Get remote logs
```

## CLI Commands

```bash
# Node commands
miner-ctrl node init --name "my-rig" --role worker
miner-ctrl node info
miner-ctrl node serve --listen :9091

# Peer commands
miner-ctrl peer add --address 192.168.1.100:9091 --name "rig"
miner-ctrl peer list
miner-ctrl peer remove <peer-id>
miner-ctrl peer ping <peer-id>

# Remote commands
miner-ctrl remote status [peer-id]
miner-ctrl remote start <peer-id> --profile <profile-id>
miner-ctrl remote stop <peer-id> [miner-name]
miner-ctrl remote logs <peer-id> <miner-name> --lines 100
```

## Network Requirements

- **Port 9091** (default) must be accessible
- TCP WebSocket connections
- Optional TLS for additional security

### Firewall Rules

```bash
# Linux (UFW)
sudo ufw allow 9091/tcp

# Linux (firewalld)
sudo firewall-cmd --permanent --add-port=9091/tcp
sudo firewall-cmd --reload
```
