# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Standard build
mkdir build && cd build
cmake ..
cmake --build . --config Release

# Build with specific features
cmake .. -DCMAKE_BUILD_TYPE=Release \
         -DWITH_HWLOC=ON \
         -DWITH_OPENCL=ON \
         -DWITH_CUDA=ON \
         -DWITH_HTTP=ON \
         -DWITH_TLS=ON

# Debug build
cmake .. -DCMAKE_BUILD_TYPE=Debug -DWITH_DEBUG_LOG=ON

# Disable specific algorithms to reduce binary size
cmake .. -DWITH_RANDOMX=OFF -DWITH_ARGON2=OFF -DWITH_KAWPOW=OFF

# Static build
cmake .. -DBUILD_STATIC=ON
```

## Key CMake Options

| Option | Default | Description |
|--------|---------|-------------|
| `WITH_HWLOC` | ON | Hardware topology support (recommended) |
| `WITH_RANDOMX` | ON | RandomX algorithms (Monero) |
| `WITH_ARGON2` | ON | Argon2 algorithms |
| `WITH_KAWPOW` | ON | KawPow algorithm (GPU only) |
| `WITH_GHOSTRIDER` | ON | GhostRider algorithm |
| `WITH_OPENCL` | ON | AMD GPU backend |
| `WITH_CUDA` | ON | NVIDIA GPU backend (requires external plugin) |
| `WITH_HTTP` | ON | HTTP API and solo mining |
| `WITH_TLS` | ON | SSL/TLS encrypted connections |
| `WITH_ASM` | ON | Assembly optimizations |
| `WITH_MSR` | ON | MSR mod for CPU tuning |
| `WITH_BENCHMARK` | ON | Built-in RandomX benchmark |

## Architecture Overview

XMRig is a C++ cryptocurrency miner supporting RandomX, KawPow, CryptoNight, and GhostRider algorithms across CPU, OpenCL (AMD), and CUDA (NVIDIA) backends.

### Source Structure (`src/`)

```
src/
├── xmrig.cpp              # Entry point: main() -> Process -> Entry -> App
├── App.h/cpp              # Application lifecycle, signal/console handling
├── core/
│   ├── Controller.h/cpp   # Main controller: manages backends, network, config
│   ├── Miner.h/cpp        # Mining orchestration, job distribution
│   └── config/            # JSON configuration parsing and validation
├── backend/
│   ├── cpu/               # CPU mining: threads, workers, platform-specific code
│   ├── opencl/            # AMD GPU: kernels, runners, OpenCL wrappers
│   └── cuda/              # NVIDIA GPU: runners, CUDA wrappers
├── crypto/
│   ├── cn/                # CryptoNight variants (x86/ARM/RISC-V implementations)
│   ├── randomx/           # RandomX with JIT compiler
│   ├── kawpow/            # KawPow (GPU-optimized)
│   ├── ghostrider/        # GhostRider algorithm
│   ├── argon2/            # Argon2 variants
│   └── common/            # Huge pages, NUMA memory pools, virtual memory
├── base/
│   ├── api/               # HTTP REST API server
│   ├── net/
│   │   ├── stratum/       # Pool protocol: Client, Job, strategies
│   │   └── http/          # HTTP client/server
│   ├── kernel/            # Process, signals, entry points
│   └── io/                # Logging, JSON utilities
├── net/
│   ├── Network.h/cpp      # Pool connection management
│   ├── JobResults.h/cpp   # Share submission queue
│   └── strategies/        # Donation strategy
└── hw/                    # Hardware: MSR access, DMI reading
```

### Key Interfaces

- **`IBackend`** - Mining backend abstraction (CPU, OpenCL, CUDA)
- **`IConsoleListener`** - Interactive console commands (h=hashrate, p=pause, r=resume)
- **`ISignalListener`** - SIGTERM/SIGINT handling for graceful shutdown
- **`IRxListener`** - RandomX dataset initialization events

### Configuration

JSON config file is the primary configuration method. Key sections:
- `pools[]` - Mining pool connections with algorithm/coin settings
- `cpu` - Thread configuration, huge pages, affinity
- `opencl`/`cuda` - GPU backend settings
- `randomx` - Dataset mode, NUMA, MSR tuning
- `http` - API server settings

Runtime config changes via HTTP API: `PUT /1/config`

### Memory Management

- **Huge pages**: 2MB and 1GB page support for RandomX performance
- **NUMA-aware**: Allocates memory on correct NUMA node via hwloc
- **Memory pools**: Reusable scratchpad memory for mining threads

### Build Dependencies

- **Required**: CMake 3.10+, C++11 compiler, libuv
- **Recommended**: hwloc (hardware topology), OpenSSL (TLS)
- **GPU**: OpenCL SDK (AMD), CUDA plugin (NVIDIA - external)

Build scripts in `scripts/`:
- `build_deps.sh` - Compile all dependencies
- `build.hwloc.sh`, `build.openssl.sh` - Individual dependency builds
- `randomx_boost.sh` - CPU-specific RandomX tuning
- `enable_1gb_pages.sh` - Linux huge page setup

### Platform-Specific Code

Platform variants use suffix naming:
- `*_unix.cpp` / `*_win.cpp` - OS-specific implementations
- `CryptoNight_x86.h` / `CryptoNight_arm.h` - Architecture-specific crypto

Supported: Linux, Windows, macOS, FreeBSD, OpenBSD, Haiku, Android, iOS
Architectures: x86-64, x86, ARMv7, ARMv8, RISC-V

## HTTP API

When built with `-DWITH_HTTP=ON`:
- `GET /1/summary` - Miner statistics
- `GET /1/threads` - Per-thread details
- `GET /1/config` - Current configuration (requires access token)
- `PUT /1/config` - Update configuration at runtime
