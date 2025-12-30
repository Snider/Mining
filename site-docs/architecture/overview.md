# Architecture Overview

The Mining Dashboard follows a modular architecture with clear separation of concerns.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Web Browser                              │
│                   Angular UI (4200)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP/REST
┌─────────────────────▼───────────────────────────────────────┐
│                   Go Backend (9090)                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  REST API (Gin)                       │  │
│  │            /api/v1/mining/*                          │  │
│  └──────────────────────────────────────────────────────┘  │
│                          │                                  │
│  ┌──────────────────────▼──────────────────────────────┐   │
│  │                 Mining Manager                       │   │
│  │  - Process lifecycle     - Stats collection         │   │
│  │  - Profile management    - Hashrate history         │   │
│  └──────────────────────────────────────────────────────┘  │
│                          │                                  │
│  ┌───────────┬───────────┴────────────┬────────────────┐   │
│  │  XMRig    │      TT-Miner          │    Future...   │   │
│  │  Adapter  │      Adapter           │    Miners      │   │
│  └───────────┴────────────────────────┴────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
     ┌─────────┐    ┌─────────┐    ┌─────────┐
     │ XMRig   │    │TT-Miner │    │ SQLite  │
     │ Process │    │ Process │    │   DB    │
     └─────────┘    └─────────┘    └─────────┘
```

## Component Responsibilities

### Frontend (Angular)

| Component | Purpose |
|-----------|---------|
| Dashboard | Real-time hashrate display, stats bar |
| Profiles | CRUD for mining configurations |
| Console | Live miner output with ANSI colors |
| Workers | Running miner instances |
| Nodes | P2P peer management |

### Backend (Go)

| Package | Purpose |
|---------|---------|
| `pkg/mining` | Core miner management, API service |
| `pkg/node` | P2P networking, identity, transport |
| `pkg/database` | SQLite persistence layer |
| `cmd/mining` | CLI commands via Cobra |

## Data Flow

### Starting a Miner

```
1. UI: POST /api/v1/mining/profiles/{id}/start
2. Service: Validates profile, calls Manager.StartMiner()
3. Manager: Creates miner instance (XMRig/TT-Miner)
4. Miner: Generates config, spawns process
5. Manager: Starts stats collection goroutine
6. Response: Returns miner name to UI
```

### Stats Collection

```
Every 10 seconds:
1. Manager iterates running miners
2. Each miner adapter polls stats (HTTP API or stdout parsing)
3. Stats stored in memory + SQLite
4. UI polls /api/v1/mining/miners for updates
```

## Storage

### Configuration Files

```
~/.config/lethean-desktop/
├── mining_profiles.json    # Saved profiles
├── miners.json            # Autostart config
├── node.json              # P2P identity
└── peers.json             # Known peers
```

### Data Files

```
~/.local/share/lethean-desktop/
├── miners/                # Installed miner binaries
│   ├── xmrig/
│   └── tt-miner/
├── node/
│   └── private.key        # X25519 private key
└── mining.db              # SQLite database
```

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Interface-based miners | Easy to add new miner types |
| Gorilla WebSocket | P2P transport with good browser support |
| SQLite | Zero-config persistence, embedded |
| Gin framework | Fast, widely used Go HTTP framework |
| Angular standalone | Modern, tree-shakable components |
