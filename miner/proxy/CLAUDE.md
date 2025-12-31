# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Standard build (out-of-source recommended)
mkdir build && cd build
cmake ..
make -j$(nproc)

# Build with specific options
cmake .. -DWITH_TLS=ON -DWITH_HTTP=ON
cmake .. -DWITH_DEBUG_LOG=ON          # Enable debug logging
cmake .. -DWITH_GOOGLE_BREAKPAD=ON    # Enable crash reporting

# Clean rebuild
rm -rf build && mkdir build && cd build && cmake .. && make -j$(nproc)

# Run the proxy (after build)
./miner-proxy -c config.json
./miner-proxy --help              # Show all CLI options
./miner-proxy --dry-run           # Test configuration and exit
```

**Build options** (CMakeLists.txt):
- `WITH_TLS` (ON) - OpenSSL/TLS support
- `WITH_HTTP` (ON) - HTTP API support
- `WITH_DEBUG_LOG` (OFF) - Debug logging (enables `APP_DEBUG` define)
- `WITH_ENV_VARS` (ON) - Environment variables in config
- `WITH_GOOGLE_BREAKPAD` (OFF) - Crash reporting

**Dependencies**: CMake 3.10+, libuv, OpenSSL, C++11 compiler

## Architecture Overview

XMRig Proxy is a high-performance CryptoNote stratum protocol proxy that can handle 100K+ miner connections while maintaining minimal pool-side connections through nonce splitting.

### Data Flow

```
Miners (100K+) → [Server] → [Login] → [Miner] → [Splitter] → Pool (few connections)
                    ↑                     ↓
              [TlsContext]          [NonceMapper]
                                         ↓
                              [Events] → [Stats/Workers]
```

### Core Components (`src/`)

**Proxy Module** (`src/proxy/`):
- `Proxy.h/cpp` - Main orchestrator; manages servers, splitters, stats, workers. Runs main tick loop (1s interval), handles garbage collection every 60s
- `Server.h/cpp` - TCP server accepting miner connections (binds to configured addresses)
- `Miner.h/cpp` - Individual miner connection state and protocol handling
- `Miners.h/cpp` - Miner pool management
- `Login.h/cpp` - Stratum authentication
- `Stats.h/cpp`, `StatsData.h` - Performance metrics aggregation

**Event System** (`src/proxy/events/`):
- Events propagate connection lifecycle changes through the proxy
- Event types: `LoginEvent`, `AcceptEvent`, `SubmitEvent`, `CloseEvent`, `ConnectionEvent`
- `IEventListener` interface for subscribing to events
- `Events.cpp` dispatches events to registered listeners

**Splitter System** (`src/proxy/splitters/`) - Handles nonce space partitioning:
- `nicehash/` - Default mode with full nonce splitting (NonceMapper, NonceStorage, NonceSplitter)
- `simple/` - Direct pool connection sharing
- `extra_nonce/` - Solo mining support
- `donate/` - Donation traffic redirection

Each splitter has: Mapper (nonce transformation), Storage (state), Splitter (orchestration)

**Configuration** (`src/core/`):
- `Config.h/cpp` - JSON config parsing via RapidJSON
- `Controller.h/cpp` - Application lifecycle
- `ConfigTransform.cpp` - Config migration

**Base Infrastructure** (`src/base/`):
- `net/` - Network I/O layer built on libuv (stratum clients, DNS, HTTP)
- `io/` - Console, signals, file watchers, JSON parsing
- `crypto/` - Algorithm definitions, keccak, SHA3
- `io/log/` - Logging backends (ConsoleLog, FileLog, SysLog)
- `kernel/` - Platform abstraction, process management, interfaces

**API** (`src/api/v1/`):
- `ApiRouter.h/cpp` - REST API for monitoring (when `WITH_HTTP=ON`)

### Key Interfaces

- `ISplitter` (`src/proxy/interfaces/`) - Splitter abstraction: `connect()`, `tick()`, `gc()`, `upstreams()`
- `IEventListener` - Event handling for connection lifecycle
- `IBaseListener` - Configuration change callbacks via `onConfigChanged()`
- `IClient` / `IClientListener` - Pool client abstraction

### Stratum Protocol

Protocol implementation follows `doc/STRATUM.md`:
- `login` - Miner authorization (returns session ID + first job)
- `job` - Pool pushes new work
- `submit` - Miner submits shares (with nonce transformation in proxy)
- `keepalived` - Connection keepalive

Extensions in `doc/STRATUM_EXT.md`: algorithm negotiation, rig identifiers, NiceHash compatibility.

### Platform-Specific Code

- `App_unix.cpp` - Linux/macOS initialization (signal handling)
- `App_win.cpp` - Windows initialization (console, service support)
- Platform libs: IOKit (macOS), ws2_32/psapi (Windows), pthread/rt (Linux)

### Key Defines

```cpp
XMRIG_PROXY_PROJECT     // Proxy-specific code paths
XMRIG_FORCE_TLS         // TLS enforcement
APP_DEVEL               // Development features (enables printState())
APP_DEBUG               // Debug logging (set via WITH_DEBUG_LOG)
XMRIG_ALGO_RANDOMX      // Algorithm support flags
XMRIG_FEATURE_HTTP      // HTTP API enabled
XMRIG_FEATURE_API       // REST API enabled
```

## Configuration

Default config template: `src/config.json`

Key sections: pools, bind addresses, proxy mode (nicehash/simple/extra_nonce), TLS certificates, HTTP API settings, logging.

Config hot-reload is enabled by default (`"watch": true`).
