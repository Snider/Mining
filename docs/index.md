# Mining Platform Documentation

Welcome to the Mining Platform documentation. This is a modern, modular cryptocurrency mining management platform with GPU support, RESTful API, and cross-platform desktop application.

## Overview

Mining Platform provides a comprehensive solution for managing cryptocurrency mining operations across multiple algorithms and hardware configurations. Whether you're mining Monero with your CPU, Ethereum Classic with your GPU, or running dual mining operations, Mining Platform gives you the tools to manage it all.

## Key Features

- **Multi-Algorithm Support**: Mine CPU and GPU across RandomX, KawPow, ETChash, ProgPowZ, Blake3, and CryptoNight algorithms
- **Dual Mining**: Run CPU and GPU mining simultaneously with separate pool configurations
- **Profile Management**: Save and quickly switch between mining configurations
- **Real-time Monitoring**: Live hashrate, shares, and performance metrics
- **RESTful API**: Full control via HTTP endpoints with Swagger documentation
- **Web Dashboard**: Embeddable Angular web component for any application
- **Desktop Application**: Native cross-platform app built with Wails v3
- **Mobile Responsive**: Touch-friendly UI optimized for all devices

## Supported Algorithms

| Algorithm | Coin | CPU | GPU (OpenCL) | GPU (CUDA) |
|-----------|------|-----|--------------|------------|
| RandomX | Monero (XMR) | ✅ | ✅ | ✅ |
| KawPow | Ravencoin (RVN) | ❌ | ✅ | ✅ |
| ETChash | Ethereum Classic (ETC) | ❌ | ✅ | ✅ |
| ProgPowZ | Zano (ZANO) | ❌ | ✅ | ✅ |
| Blake3 | Decred (DCR) | ✅ | ✅ | ✅ |
| CryptoNight | Various | ✅ | ✅ | ✅ |

## Quick Links

- **[Getting Started](getting-started/index.md)**: Installation and setup guide
- **[User Guide](user-guide/cli.md)**: Learn how to use the CLI, web dashboard, and desktop app
- **[API Reference](api/index.md)**: RESTful API documentation
- **[Development Guide](development/index.md)**: Contributing and building from source
- **[Pool Integration](reference/pools.md)**: Mining pool configuration and recommendations

## Architecture

The platform consists of three main components:

1. **Core Go Backend** (`pkg/mining/`): Manages miner lifecycle, configuration, and statistics
2. **Web Dashboard** (`ui/`): Angular-based web component for monitoring and control
3. **Desktop Application** (`cmd/desktop/`): Native app with embedded web dashboard

## Managed Mining Software

Mining Platform handles installation and configuration of popular mining software:

- **XMRig**: High-performance CPU/GPU miner for RandomX and CryptoNight
- **T-Rex**: NVIDIA GPU miner for KawPow, Ethash, and more
- **lolMiner**: AMD/NVIDIA GPU miner for Ethash, Beam, Equihash
- **TT-Miner**: NVIDIA GPU miner for Ethash, KawPow, Autolykos2

## Community and Support

- **GitHub**: [Snider/Mining](https://github.com/Snider/Mining)
- **Issue Tracker**: [Report bugs or request features](https://github.com/Snider/Mining/issues)
- **License**: EUPL-1.2

## Next Steps

New to Mining Platform? Start with our [Installation Guide](getting-started/index.md) to get up and running in minutes.

Already installed? Check out the [Quick Start Guide](getting-started/quick-start.md) to begin mining.
