# Mining

[![CI](https://github.com/Snider/Mining/actions/workflows/e2e.yml/badge.svg)](https://github.com/Snider/Mining/actions/workflows/e2e.yml)
[![Release](https://img.shields.io/github/release/Snider/Mining.svg)](https://github.com/Snider/Mining/releases)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Angular](https://img.shields.io/badge/angular-20+-DD0031.svg?logo=angular&logoColor=white)](https://angular.io)
[![GoDoc](https://pkg.go.dev/badge/github.com/Snider/Mining.svg)](https://pkg.go.dev/github.com/Snider/Mining)
[![Go Report Card](https://goreportcard.com/badge/github.com/Snider/Mining)](https://goreportcard.com/report/github.com/Snider/Mining)
[![codecov](https://codecov.io/gh/Snider/Mining/branch/main/graph/badge.svg)](https://codecov.io/gh/Snider/Mining)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/license/eupl-1-2)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-lightgrey.svg)](https://github.com/Snider/Mining/releases)
[![Docs](https://img.shields.io/badge/docs-mkdocs-blue.svg)](https://snider.github.io/Mining/)

A modern, modular cryptocurrency mining management platform with GPU support, RESTful API, and cross-platform desktop application.

<img width="834" height="657" alt="Mining Dashboard" src="https://github.com/user-attachments/assets/d4fc4704-819c-4aca-bcd3-ae4af6e25c1b" />

## Features

### Supported Algorithms

| Algorithm | Coin | CPU | GPU (OpenCL) | GPU (CUDA) |
|-----------|------|-----|--------------|------------|
| [RandomX](https://miningpoolstats.stream/monero) | [Monero (XMR)](https://www.getmonero.org/) | ✅ | ✅ | ✅ |
| [KawPow](https://miningpoolstats.stream/ravencoin) | [Ravencoin (RVN)](https://ravencoin.org/) | ❌ | ✅ | ✅ |
| [ETChash](https://miningpoolstats.stream/ethereumclassic) | [Ethereum Classic (ETC)](https://ethereumclassic.org/) | ❌ | ✅ | ✅ |
| [ProgPowZ](https://miningpoolstats.stream/zano) | [Zano (ZANO)](https://zano.org/) | ❌ | ✅ | ✅ |
| [Blake3](https://miningpoolstats.stream/decred) | [Decred (DCR)](https://decred.org/) | ✅ | ✅ | ✅ |
| [CryptoNight](https://miningpoolstats.stream/monero) | Various | ✅ | ✅ | ✅ |

### Core Capabilities

- **Multi-Algorithm Mining**: Support for CPU and GPU mining across multiple algorithms
- **Dual Mining**: Run CPU and GPU mining simultaneously with separate pools
- **Profile Management**: Save and switch between mining configurations
- **Real-time Monitoring**: Live hashrate, shares, and performance metrics
- **RESTful API**: Full control via HTTP endpoints with Swagger documentation
- **Web Dashboard**: Embeddable Angular web component for any application
- **Desktop Application**: Native cross-platform app built with Wails v3
- **Mobile Responsive**: Touch-friendly UI with drawer navigation
- **Simulation Mode**: Test the UI without real mining hardware

### Why Mining Platform?

| Feature | Mining Platform | NiceHash | HiveOS | Manual XMRig |
|---------|:---------------:|:--------:|:------:|:------------:|
| Open Source | ✅ | ❌ | ❌ | ✅ |
| No Fees | ✅ | ❌ (2%+) | ❌ ($3/mo) | ✅ |
| Multi-Algorithm | ✅ | ✅ | ✅ | ❌ |
| GUI Dashboard | ✅ | ✅ | ✅ | ❌ |
| Profile Management | ✅ | ❌ | ✅ | ❌ |
| Dual Mining | ✅ | ❌ | ✅ | ❌ |
| Desktop App | ✅ | ❌ | ❌ | ❌ |
| Embeddable Component | ✅ | ❌ | ❌ | ❌ |
| Self-Hosted | ✅ | ❌ | ❌ | ✅ |
| Simulation Mode | ✅ | ❌ | ❌ | ❌ |

### Mining Software

Manages installation and configuration of:
- **XMRig** - High-performance CPU/GPU miner (RandomX, CryptoNight)
- **T-Rex** - NVIDIA GPU miner (KawPow, Ethash, and more)
- **lolMiner** - AMD/NVIDIA GPU miner (Ethash, Beam, Equihash)
- **TT-Miner** - NVIDIA GPU miner (Ethash, KawPow, Autolykos2)

## Quick Start

### Docker (Fastest)

```bash
# Run with Docker - no dependencies required
docker run -p 9090:9090 ghcr.io/snider/mining:latest

# Access the dashboard at http://localhost:9090
```

### CLI

```bash
# Install
go install github.com/Snider/Mining/cmd/mining@latest

# Start the API server
miner-ctrl serve --host localhost --port 9090

# Or use the interactive shell
miner-ctrl serve
```

### Web Component

```html
<script type="module" src="./mbe-mining-dashboard.js"></script>
<snider-mining api-base-url="http://localhost:9090/api/v1/mining"></snider-mining>
```

### Desktop Application

Download pre-built binaries from [Releases](https://github.com/Snider/Mining/releases) or build from source:

```bash
cd cmd/desktop/mining-desktop
wails3 build
```

## Architecture

```
Mining/
├── cmd/
│   ├── mining/              # CLI application (miner-ctrl)
│   └── desktop/             # Wails desktop app
├── pkg/mining/              # Core Go package
│   ├── mining.go            # Interfaces and types
│   ├── manager.go           # Miner lifecycle management
│   ├── service.go           # RESTful API (Gin)
│   └── profile_manager.go   # Profile persistence
├── miner/                   # Standalone C++ mining tools
│   ├── core/                # CPU/GPU miner binary
│   ├── proxy/               # Stratum proxy for farms
│   ├── cuda/                # CUDA plugin for NVIDIA
│   └── README.md            # Miner documentation
└── ui/                      # Angular 20+ web dashboard
    └── src/app/
        ├── components/      # Reusable UI components
        └── pages/           # Route pages
```

## Standalone Miner Tools

The `miner/` directory contains standalone C++ mining programs that can be used independently without the GUI:

```bash
# Build miner binaries
make build-miner

# Or build individually
make build-miner-core   # CPU/GPU miner
make build-miner-proxy  # Stratum proxy

# Run directly
./miner/core/build/miner -o pool.example.com:3333 -u WALLET -p x
./miner/proxy/build/miner-proxy -o pool.example.com:3333 -b 0.0.0.0:3333
```

Pre-built binaries are available from [Releases](https://github.com/letheanVPN/Mining/releases). See [miner/README.md](miner/README.md) for full documentation.

## API Reference

Base path: `/api/v1/mining`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/info` | System info and installed miners |
| GET | `/miners` | List running miners |
| POST | `/miners/:name` | Start a miner |
| DELETE | `/miners/:name` | Stop a miner |
| GET | `/miners/:name/stats` | Get miner statistics |
| GET | `/profiles` | List saved profiles |
| POST | `/profiles` | Create a profile |
| PUT | `/profiles/:id` | Update a profile |
| DELETE | `/profiles/:id` | Delete a profile |
| POST | `/miners/:name/install` | Install miner software |

Swagger UI: `http://localhost:9090/api/v1/mining/swagger/index.html`

## Development

### Prerequisites

- Go 1.24+
- Node.js 20+ (for UI development)
- CMake 3.21+ (for miner core)
- OpenCL SDK (for GPU support)

### Build Commands

```bash
# Go Backend
make build              # Build CLI binary
make test               # Run all tests (Go + C++)
make dev                # Start dev server on :9090

# Miner (C++ Binaries)
make build-miner        # Build miner and proxy
make build-miner-all    # Build and package to dist/miner/

# Frontend
cd ui
npm install
npm run build           # Build web component
npm test                # Run unit tests

# Desktop
cd cmd/desktop/mining-desktop
wails3 build            # Build native app
```

## Configuration

Mining profiles are stored in `~/.config/lethean-desktop/mining_profiles.json`

Example profile:
```json
{
  "id": "uuid",
  "name": "My XMR Mining",
  "minerType": "xmrig",
  "config": {
    "pool": "stratum+tcp://pool.supportxmr.com:3333",
    "wallet": "YOUR_WALLET_ADDRESS",
    "algo": "rx/0"
  }
}
```

## Contributing

We welcome contributions! Please read our [Code of Conduct](CODE_OF_CONDUCT.md) and [Contributing Guidelines](docs/development/contributing.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTORS.md](CONTRIBUTORS.md) for the list of contributors.

## License

This project is licensed under the EUPL-1.2 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [XMRig](https://github.com/xmrig/xmrig) - High performance miner
- [Wails](https://wails.io) - Desktop application framework
- [Angular](https://angular.io) - Web framework
- [Gin](https://gin-gonic.com) - HTTP web framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
