# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Standard CMake build
mkdir build && cd build
cmake ..
make

# Optional build flags
cmake -DWITH_TLS=ON ..       # Enable OpenSSL/TLS support
cmake -DWITH_LIBPNG=ON ..    # Enable LibPNG for better performance
cmake -DWITH_DEBUG_LOG=ON .. # Enable debug logging
```

### Dependencies
- libuv (required)
- OpenSSL (optional, for TLS)
- libpng (optional, for performance)

## Architecture Overview

This is a C++ tool that generates heatmap visualizations of nonce distributions from cryptocurrency blockchain data. It connects to a daemon via JSON-RPC to fetch block headers and renders the nonce data as a PNG heatmap.

### Core Components (`src/`)

- **`App`**: Main application class implementing console, signal, and network event listeners. Orchestrates the sync→render workflow.

- **`Network`**: Handles JSON-RPC communication with the daemon. Fetches block headers using `get_last_block_header` and `get_block_headers_range` methods. Manages concurrent range requests and stores nonce data.

- **`Nonces`**: Persistence layer for nonce data. Loads/saves nonce arrays to JSON files.

- **`Heatmap`**: Renders nonce distribution as PNG using the heatmap library. Calculates image dimensions and generates metadata JSON.

- **`Config`**: JSON configuration parsing. Extends `ConfigFile` base class.

- **`Daemon`**: Connection settings (host, port, TLS, concurrency, timeouts).

- **`Job`**: Heatmap rendering parameters (dimensions, radius, nonce range, output file).

### Base Library (`src/base/`)

Shared infrastructure from XMRig:
- `io/`: Console, signals, JSON handling, logging
- `net/`: HTTP client implementation
- `kernel/`: Process management, config file handling, interfaces
- `tools/`: String utilities

### Third-Party (`src/3rdparty/`)

- `heatmap/`: C heatmap rendering library with Spectral colorscheme
- `http-parser/`: HTTP parsing for RPC responses

## Configuration

Uses `config.json` in working directory. Key options:
- `nonces_file`: Storage for fetched nonce data
- `offline`: Run without daemon connection (requires existing nonces file)
- `daemon`: Connection settings (host, port, tls)
- `heatmap`: Output settings (name, height, radius, block range)

Override working directory with `--data-dir` or `-d`, config file with `--config` or `-c`.
