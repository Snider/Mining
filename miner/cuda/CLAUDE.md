# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xmrig-cuda is a CUDA plugin for XMRig cryptocurrency miner, providing NVIDIA GPU acceleration. It compiles to a shared library (`libxmrig-cuda.so` on Linux, `xmrig-cuda.dll` on Windows) that XMRig loads at runtime.

## Build Commands

```bash
# Standard build
mkdir build && cd build
cmake ..
cmake --build .

# Build with specific CUDA architectures (semicolon-separated compute capabilities)
cmake -DCUDA_ARCH="60;75;86" ..

# Build with specific algorithm selection
cmake -DWITH_RANDOMX=ON -DWITH_KAWPOW=ON -DWITH_CN_R=ON ..

# Disable Driver API (disables CN-R and KawPow which require NVRTC)
cmake -DWITH_DRIVER_API=OFF ..

# Custom CUDA toolkit path
cmake -DCUDA_TOOLKIT_ROOT_DIR=/usr/local/cuda-12 ..

# Debug build with register usage and intermediate files
cmake -DCUDA_SHOW_REGISTER=ON -DCUDA_KEEP_FILES=ON ..
```

**Output:** `libxmrig-cuda.so` (Linux) or `xmrig-cuda.dll` + NVRTC DLLs (Windows)

## Architecture

### Plugin Interface (`src/xmrig-cuda.h`)

C-linkage API that XMRig calls. Key functions:
- `alloc()`/`release()` - GPU context lifecycle
- `deviceInfo()`/`deviceInit()` - GPU detection and initialization
- `setJob()` - Configure algorithm and job data
- `cnHash()` - CryptoNight family hashing
- `rxPrepare()`/`rxHash()` - RandomX hashing
- `kawPowPrepare_v2()`/`kawPowHash()` - KawPow hashing

### GPU Context (`nvid_ctx`)

Per-GPU state including device properties, memory pointers, and algorithm-specific resources. Defined in `src/cryptonight.h`.

### Algorithm Implementations

| Algorithm | Source Files | CMake Option |
|-----------|--------------|--------------|
| CryptoNight variants | `cuda_core.cu`, `cuda_extra.cu`, `CryptonightR.cu` | `WITH_CN_*` options |
| RandomX | `src/RandomX/` (per-coin subdirs) | `WITH_RANDOMX` |
| KawPow | `src/KawPow/raven/` | `WITH_KAWPOW` |

CN-R and KawPow use CUDA Driver API for runtime kernel compilation via NVRTC.

### CUDA Architecture Support

Automatically configured based on CUDA toolkit version:
- CUDA 8.x: Fermi (20), Kepler (30), Maxwell (50), Pascal (60)
- CUDA 9.x: Kepler+, adds Volta (70)
- CUDA 10.x: Kepler+, adds Turing (75)
- CUDA 11.x: Maxwell+ (Kepler partial), adds Ampere (80, 86, 87)
- CUDA 11.8+: Adds Ada/Hopper (89, 90)

## Key CMake Options

| Option | Default | Description |
|--------|---------|-------------|
| `WITH_DRIVER_API` | ON | Required for CN-R and KawPow (needs NVRTC) |
| `WITH_RANDOMX` | ON* | RandomX algorithms (*OFF if CUDA < 9.0) |
| `WITH_KAWPOW` | ON* | KawPow/RavenCoin (*requires Driver API) |
| `WITH_CN_R` | ON | CryptoNight-R (*requires Driver API) |
| `CUDA_ARCH` | auto | GPU compute capabilities to target |
| `XMRIG_LARGEGRID` | ON | Support >128 CUDA blocks |

## File Layout

```
src/
├── xmrig-cuda.cpp     # Main plugin implementation
├── xmrig-cuda.h       # C API header
├── cuda_core.cu       # Core CUDA kernels
├── cuda_extra.cu      # CUDA utilities
├── CryptonightR.cu    # CN-R with runtime compilation
├── crypto/            # Algorithm definitions
├── RandomX/           # RandomX per-coin implementations
├── KawPow/            # KawPow with DAG support
└── 3rdparty/cub/      # NVIDIA CUB library
cmake/
├── CUDA.cmake         # CUDA toolchain and arch detection
├── CUDA-Version.cmake # CUDA version detection
└── flags.cmake        # Compiler flags
```

## Platform Notes

- **macOS:** Not supported (NVIDIA dropped CUDA support)
- **Windows:** Build auto-copies `nvrtc64*.dll` dependencies
- **Linux:** Requires CUDA toolkit and nvidia driver
