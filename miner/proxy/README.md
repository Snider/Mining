# Miner Proxy

High-performance stratum protocol proxy for cryptocurrency mining farms. Efficiently manages 100K+ miner connections while maintaining minimal pool-side connections through nonce splitting.

## Features

- Handle 100K+ concurrent miner connections
- Reduce pool connections (100,000 miners → ~400 pool connections)
- NiceHash compatibility mode
- TLS/SSL support for secure connections
- HTTP API for monitoring
- Low memory footprint (~1GB RAM for 100K connections)

## Quick Start

### Download

Pre-built binaries are available from [Releases](https://github.com/letheanVPN/Mining/releases).

### Usage

```bash
# Basic usage
./miner-proxy -o pool.example.com:3333 -u YOUR_WALLET -b 0.0.0.0:3333

# With config file (recommended)
./miner-proxy -c config.json

# Test configuration
./miner-proxy --dry-run -c config.json

# Show all options
./miner-proxy --help
```

### Command Line Options

```
Network:
  -o, --url=URL                 URL of mining server
  -a, --algo=ALGO               mining algorithm
  -u, --user=USERNAME           username for mining server
  -p, --pass=PASSWORD           password for mining server
  -k, --keepalive               send keepalive packets
      --tls                     enable SSL/TLS support

Proxy:
  -b, --bind=ADDR               bind to specified address (e.g., "0.0.0.0:3333")
  -m, --mode=MODE               proxy mode: nicehash (default) or simple
      --custom-diff=N           override pool difficulty
      --access-password=P       password to restrict proxy access

API:
      --http-host=HOST          bind host for HTTP API (default: 127.0.0.1)
      --http-port=N             bind port for HTTP API
      --http-access-token=T     access token for HTTP API

TLS:
      --tls-bind=ADDR           bind with TLS enabled
      --tls-cert=FILE           TLS certificate file (PEM)
      --tls-cert-key=FILE       TLS private key file (PEM)

Logging:
  -l, --log-file=FILE           log all output to file
  -A, --access-log-file=FILE    log worker access to file
      --verbose                 verbose output

Misc:
  -c, --config=FILE             load JSON configuration file
  -B, --background              run in background
  -V, --version                 show version
  -h, --help                    show help
```

## Configuration

### JSON Config (config.json)

```json
{
    "mode": "nicehash",
    "pools": [
        {
            "url": "stratum+tcp://pool.example.com:3333",
            "user": "YOUR_WALLET",
            "pass": "x",
            "keepalive": true
        }
    ],
    "bind": [
        {
            "host": "0.0.0.0",
            "port": 3333
        },
        {
            "host": "0.0.0.0",
            "port": 3334,
            "tls": true
        }
    ],
    "http": {
        "enabled": true,
        "host": "127.0.0.1",
        "port": 8081,
        "access-token": "your-secret-token"
    },
    "tls": {
        "cert": "/path/to/cert.pem",
        "cert-key": "/path/to/key.pem"
    },
    "access-password": null,
    "workers": true,
    "verbose": false
}
```

### Proxy Modes

**NiceHash Mode** (default): Full nonce splitting for maximum efficiency
- Best for large farms with many workers
- Each worker gets unique nonce space
- Maximum reduction in pool connections

**Simple Mode**: Direct passthrough with shared connections
- Simpler setup
- Workers share pool connections
- Good for smaller setups

## Building from Source

### Dependencies

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install build-essential cmake libuv1-dev libssl-dev
```

**macOS:**
```bash
brew install cmake libuv openssl
```

### Build

```bash
mkdir build && cd build
cmake ..
cmake --build . --config Release

# With debug logging
cmake .. -DWITH_DEBUG_LOG=ON
```

### CMake Options

| Option | Default | Description |
|--------|---------|-------------|
| `WITH_TLS` | ON | SSL/TLS support |
| `WITH_HTTP` | ON | HTTP API |
| `WITH_DEBUG_LOG` | OFF | Debug logging |
| `BUILD_TESTS` | ON | Build unit tests |

## HTTP API

| Endpoint | Description |
|----------|-------------|
| `GET /1/summary` | Proxy statistics |
| `GET /1/workers` | Connected workers list |
| `GET /1/config` | Current configuration |

Example:
```bash
curl http://127.0.0.1:8081/1/summary
```

## High Connection Setup (Linux)

For 1000+ connections, increase file descriptor limits:

```bash
# /etc/security/limits.conf
* soft nofile 1000000
* hard nofile 1000000

# /etc/sysctl.conf
fs.file-max = 1000000
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
```

Then apply:
```bash
sudo sysctl -p
```

## Testing

```bash
cd build

# Run all tests
ctest --output-on-failure

# Run specific test suites
./tests/unit_tests
./tests/integration_tests

# Run with verbose output
./tests/unit_tests --gtest_verbose
```

## License

Copyright (c) 2025 Lethean <https://lethean.io>

Licensed under the European Union Public License 1.2 (EUPL-1.2).
