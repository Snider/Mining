# Lethean Miner Suite

[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/license/eupl-1-2)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows%20%7C%20freebsd-lightgrey.svg)](https://github.com/letheanVPN/Mining/releases)

High-performance cryptocurrency mining tools. These standalone C++ programs can be used independently or managed through the Mining Platform GUI.

## Components

| Component | Description | Binary |
|-----------|-------------|--------|
| [**core**](core/) | CPU/GPU miner with full algorithm support | `miner` |
| [**proxy**](proxy/) | Stratum proxy for mining farms (100K+ connections) | `miner-proxy` |
| [**cuda**](cuda/) | CUDA plugin for NVIDIA GPUs | `libminer-cuda.so` |
| [**config**](config/) | Configuration generator tool | `miner-config` |
| [**workers**](workers/) | Worker management utilities | `miner-workers` |
| [**heatmap**](heatmap/) | Hardware temperature visualization | `miner-heatmap` |

## Supported Algorithms

### CPU Mining
| Algorithm | Coins |
|-----------|-------|
| RandomX | Monero (XMR), Lethean (LTHN), Wownero (WOW) |
| CryptoNight | Various CN variants |
| GhostRider | Raptoreum (RTM) |
| Argon2 | Chukwa, Ninja |

### GPU Mining (OpenCL/CUDA)
| Algorithm | Coins |
|-----------|-------|
| RandomX | Monero, Lethean |
| KawPow | Ravencoin (RVN), Neoxa |
| ETChash | Ethereum Classic (ETC) |
| ProgPowZ | Zano (ZANO) |
| Blake3 | Decred (DCR) |

## Quick Start

### Download Pre-built Binaries

Download from [Releases](https://github.com/letheanVPN/Mining/releases):
- `miner-linux-x64.tar.gz` - Linux x86_64
- `miner-linux-arm64.tar.gz` - Linux ARM64
- `miner-macos-x64.tar.gz` - macOS Intel
- `miner-macos-arm64.tar.gz` - macOS Apple Silicon
- `miner-windows-x64.zip` - Windows x64

### Run the Miner

```bash
# Basic CPU mining
./miner -o pool.example.com:3333 -u YOUR_WALLET -p x

# With config file (recommended)
./miner -c config.json

# CPU + GPU mining
./miner -c config.json --opencl --cuda

# Show help
./miner --help
```

### Run the Proxy

```bash
# Start proxy for mining farm
./miner-proxy -o pool.example.com:3333 -u YOUR_WALLET -b 0.0.0.0:3333

# With config file
./miner-proxy -c proxy-config.json
```

## Building from Source

### Prerequisites

**All Platforms:**
- CMake 3.10+
- C++11 compatible compiler
- libuv
- OpenSSL (for TLS support)

**Linux:**
```bash
sudo apt-get install build-essential cmake libuv1-dev libssl-dev libhwloc-dev
```

**macOS:**
```bash
brew install cmake libuv openssl hwloc
```

**Windows:**
- Visual Studio 2019+ with C++ workload
- vcpkg for dependencies

### Build Miner Core

```bash
cd core
mkdir build && cd build

# Standard build
cmake ..
cmake --build . --config Release -j$(nproc)

# With GPU support
cmake .. -DWITH_OPENCL=ON -DWITH_CUDA=ON

# Static build (portable)
cmake .. -DBUILD_STATIC=ON

# Minimal build (RandomX only)
cmake .. -DWITH_ARGON2=OFF -DWITH_KAWPOW=OFF -DWITH_GHOSTRIDER=OFF
```

### Build Proxy

```bash
cd proxy
mkdir build && cd build
cmake ..
cmake --build . --config Release -j$(nproc)
```

### Build All Components

From the repository root:

```bash
make build-miner          # Build miner core
make build-miner-proxy    # Build proxy
make build-miner-all      # Build all components
```

## Configuration

### Miner Config (config.json)

```json
{
    "autosave": true,
    "cpu": true,
    "opencl": false,
    "cuda": false,
    "pools": [
        {
            "url": "stratum+tcp://pool.example.com:3333",
            "user": "YOUR_WALLET",
            "pass": "x",
            "keepalive": true,
            "tls": false
        }
    ],
    "http": {
        "enabled": true,
        "host": "127.0.0.1",
        "port": 8080,
        "access-token": null
    }
}
```

### Proxy Config (proxy-config.json)

```json
{
    "mode": "nicehash",
    "pools": [
        {
            "url": "stratum+tcp://pool.example.com:3333",
            "user": "YOUR_WALLET",
            "pass": "x"
        }
    ],
    "bind": [
        {
            "host": "0.0.0.0",
            "port": 3333
        }
    ],
    "http": {
        "enabled": true,
        "host": "127.0.0.1",
        "port": 8081
    }
}
```

## HTTP API

Both miner and proxy expose HTTP APIs for monitoring and control.

### Miner API (default: http://127.0.0.1:8080)

| Endpoint | Description |
|----------|-------------|
| `GET /1/summary` | Mining statistics |
| `GET /1/threads` | Per-thread hashrates |
| `GET /1/config` | Current configuration |
| `PUT /1/config` | Update configuration |

### Proxy API (default: http://127.0.0.1:8081)

| Endpoint | Description |
|----------|-------------|
| `GET /1/summary` | Proxy statistics |
| `GET /1/workers` | Connected workers |
| `GET /1/config` | Current configuration |

## Performance Tuning

### CPU Mining

```bash
# Enable huge pages (Linux)
sudo sysctl -w vm.nr_hugepages=1280

# Or permanent (add to /etc/sysctl.conf)
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf

# Enable 1GB pages (better performance)
sudo ./scripts/enable_1gb_pages.sh
```

### GPU Mining

```bash
# AMD GPUs - increase virtual memory
# Add to /etc/security/limits.conf:
# * soft memlock unlimited
# * hard memlock unlimited

# NVIDIA GPUs - optimize power
nvidia-smi -pl 120  # Set power limit
```

## Testing

```bash
# Run miner tests
cd core/build
ctest --output-on-failure

# Run proxy tests
cd proxy/build
./tests/unit_tests
./tests/integration_tests
```

## Directory Structure

```
miner/
├── core/               # Main miner (CPU/GPU)
│   ├── src/
│   │   ├── backend/    # CPU, OpenCL, CUDA backends
│   │   ├── crypto/     # Algorithm implementations
│   │   ├── base/       # Network, I/O, utilities
│   │   └── core/       # Configuration, controller
│   ├── scripts/        # Build and setup scripts
│   └── CMakeLists.txt
├── proxy/              # Stratum proxy
│   ├── src/
│   │   ├── proxy/      # Proxy core (splitters, events)
│   │   └── base/       # Shared base code
│   ├── tests/          # Unit and integration tests
│   └── CMakeLists.txt
├── cuda/               # CUDA plugin
├── config/             # Config generator
├── workers/            # Worker utilities
├── heatmap/            # Temperature visualization
├── deps/               # Dependency build scripts
└── README.md           # This file
```

## License

Copyright (c) 2025 Lethean <https://lethean.io>

Licensed under the European Union Public License 1.2 (EUPL-1.2).
See [LICENSE](../LICENSE) for details.

## Related Projects

- [Mining Platform](../) - GUI management platform
- [Lethean](https://lethean.io) - Lethean Network
