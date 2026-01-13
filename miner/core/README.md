# Miner

High-performance, cross-platform CPU/GPU cryptocurrency miner supporting RandomX, KawPow, CryptoNight, GhostRider, ETChash, ProgPowZ, and Blake3 algorithms.

## Features

### Mining Backends
- **CPU** (x86/x64/ARMv7/ARMv8/RISC-V)
- **OpenCL** for AMD GPUs
- **CUDA** for NVIDIA GPUs via external [CUDA plugin](../cuda/)

### Supported Algorithms

| Algorithm | Variants | CPU | GPU |
|-----------|----------|-----|-----|
| RandomX | rx/0, rx/wow, rx/arq, rx/graft, rx/sfx, rx/keva | Yes | Yes |
| CryptoNight | cn/0-2, cn-lite, cn-heavy, cn-pico | Yes | Yes |
| GhostRider | gr | Yes | No |
| Argon2 | chukwa, chukwa2, ninja | Yes | No |
| KawPow | kawpow | No | Yes |
| ETChash | etchash, ethash | No | Yes |
| ProgPowZ | progpowz | No | Yes |
| Blake3 | blake3 | Yes | Yes |

## Quick Start

### Download

Pre-built binaries are available from [Releases](https://github.com/letheanVPN/Mining/releases).

### Usage

```bash
# Basic CPU mining to a pool
./miner -o pool.example.com:3333 -u YOUR_WALLET -p x

# With JSON config (recommended)
./miner -c config.json

# Enable GPU mining
./miner -c config.json --opencl --cuda

# Benchmark mode
./miner --bench=1M

# Show all options
./miner --help
```

### Configuration

The recommended way to configure the miner is via JSON config file:

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
            "keepalive": true
        }
    ]
}
```

## Building from Source

### Dependencies

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install git build-essential cmake libuv1-dev libssl-dev libhwloc-dev
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install git cmake gcc gcc-c++ libuv-devel openssl-devel hwloc-devel
```

**macOS:**
```bash
brew install cmake libuv openssl hwloc
```

**Windows:**
- Visual Studio 2019 or later
- CMake 3.10+
- vcpkg for dependencies

### Build Commands

```bash
mkdir build && cd build

# Standard build
cmake ..
cmake --build . --config Release

# With GPU support
cmake .. -DWITH_OPENCL=ON -DWITH_CUDA=ON

# Static binary
cmake .. -DBUILD_STATIC=ON

# Debug build
cmake .. -DCMAKE_BUILD_TYPE=Debug -DWITH_DEBUG_LOG=ON

# Minimal build (reduce binary size)
cmake .. -DWITH_KAWPOW=OFF -DWITH_GHOSTRIDER=OFF -DWITH_ARGON2=OFF
```

### CMake Options

| Option | Default | Description |
|--------|---------|-------------|
| `WITH_HWLOC` | ON | Hardware topology support |
| `WITH_RANDOMX` | ON | RandomX algorithms |
| `WITH_ARGON2` | ON | Argon2 algorithms |
| `WITH_KAWPOW` | ON | KawPow (GPU only) |
| `WITH_ETCHASH` | ON | ETChash/Ethash (GPU only) |
| `WITH_PROGPOWZ` | ON | ProgPowZ (GPU only) |
| `WITH_BLAKE3DCR` | ON | Blake3 for Decred |
| `WITH_GHOSTRIDER` | ON | GhostRider algorithm |
| `WITH_OPENCL` | ON | AMD GPU support |
| `WITH_CUDA` | ON | NVIDIA GPU support |
| `WITH_HTTP` | ON | HTTP API |
| `WITH_TLS` | ON | SSL/TLS support |
| `WITH_ASM` | ON | Assembly optimizations |
| `WITH_MSR` | ON | MSR mod for CPU tuning |
| `BUILD_STATIC` | OFF | Static binary |
| `BUILD_TESTS` | OFF | Build unit tests |

## HTTP API

When built with `-DWITH_HTTP=ON`, the miner exposes a REST API:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/1/summary` | GET | Mining statistics |
| `/1/threads` | GET | Per-thread details |
| `/1/config` | GET | Current configuration |
| `/1/config` | PUT | Update configuration |

Example:
```bash
curl http://127.0.0.1:8080/1/summary
```

## Performance Optimization

### Huge Pages (Linux)

```bash
# Temporary
sudo sysctl -w vm.nr_hugepages=1280

# Permanent
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf

# 1GB pages (best performance)
sudo ./scripts/enable_1gb_pages.sh
```

### MSR Mod (Intel/AMD CPUs)

The miner can automatically apply MSR tweaks for better RandomX performance. Requires root/admin privileges.

```bash
sudo ./miner -c config.json
```

## Testing

```bash
# Build with tests
cmake .. -DBUILD_TESTS=ON
cmake --build .

# Run tests
ctest --output-on-failure
```

## License

Copyright (c) 2025 Lethean <https://lethean.io>

Licensed under the European Union Public License 1.2 (EUPL-1.2).
