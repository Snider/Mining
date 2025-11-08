# Mining

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![GoDoc](https://pkg.go.dev/badge/github.com/Snider/Mining.svg)](https://pkg.go.dev/github.com/Snider/Mining)
[![Go Report Card](https://goreportcard.com/badge/github.com/Snider/Mining)](https://goreportcard.com/report/github.com/Snider/Mining)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/license/eupl-1-2)
[![Release](https://img.shields.io/github/release/Snider/Mining.svg)](https://github.com/Snider/Mining/releases)
[![codecov](https://codecov.io/gh/Snider/Mining/branch/main/graph/badge.svg)](https://codecov.io/gh/Snider/Mining)

GoLang Miner management with RESTful control - A modern, modular package for managing cryptocurrency miners.

## Overview

Mining is a Go package designed to provide comprehensive miner management capabilities. It can be used both as a standalone CLI tool and as a module/plugin in other Go projects. The package offers:

- **Miner Lifecycle Management**: Start, stop, and monitor miners
- **Status Tracking**: Real-time status and hash rate monitoring
- **CLI Interface**: Easy-to-use command-line interface built with Cobra
- **Modular Design**: Import as a package in your own projects
- **RESTful Ready**: Designed for integration with RESTful control systems

## Features

- ✅ Start and stop miners programmatically
- ✅ Monitor miner status and performance
- ✅ Track hash rates
- ✅ List all active miners
- ✅ CLI for easy management
- ✅ Designed as a reusable Go module
- ✅ Comprehensive test coverage
- ✅ Standards-compliant configuration (CodeRabbit, GoReleaser)

## Installation

### As a CLI Tool

```bash
go install github.com/Snider/Mining/cmd/mining@latest
```

### As a Go Module

```bash
go get github.com/Snider/Mining
```

## Usage

### CLI Usage

```bash
# Start a miner
mining start --name bitcoin-miner-1 --algorithm sha256 --pool pool.bitcoin.com

# List all miners
mining list

# Get miner status
mining status <miner-id>

# Stop a miner
mining stop <miner-id>

# Show version
mining --version

# Show help
mining --help
```

### As a Go Package

```go
package main

import (
    "fmt"
    "github.com/Snider/Mining/pkg/mining"
)

func main() {
    // Create a manager
    manager := mining.NewManager()

    // Start a miner
    config := mining.MinerConfig{
        Name:      "my-miner",
        Algorithm: "sha256",
        Pool:      "pool.example.com",
        Wallet:    "your-wallet-address",
    }

    miner, err := manager.StartMiner(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Started miner: %s\n", miner.ID)

    // Update hash rate
    manager.UpdateHashRate(miner.ID, 150.5)

    // List all miners
    miners := manager.ListMiners()
    for _, m := range miners {
        fmt.Printf("%s: %s (%.2f H/s)\n", m.ID, m.Name, m.HashRate)
    }

    // Stop the miner
    manager.StopMiner(miner.ID)
}
```

## Development

### Prerequisites

- Go 1.24 or higher
- Make (optional, for using Makefile targets)

### Build

```bash
# Build the CLI
go build -o mining ./cmd/mining

# Run demo
go run main.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Project Structure

```
.
├── cmd/
│   └── mining/          # CLI application
│       ├── main.go      # CLI entry point
│       └── cmd/         # Cobra commands
├── pkg/
│   └── mining/          # Core mining package
│       ├── mining.go    # Main package code
│       └── mining_test.go
├── main.go              # Demo/development main
├── .coderabbit.yaml     # CodeRabbit configuration
├── .goreleaser.yaml     # GoReleaser configuration
├── .gitignore
├── go.mod
├── LICENSE
└── README.md
```

## Configuration

### CodeRabbit

The project uses CodeRabbit for automated code reviews. Configuration is in `.coderabbit.yaml`.

### GoReleaser

Releases are managed with GoReleaser. Configuration is in `.goreleaser.yaml`. To create a release:

```bash
# Tag a version
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# GoReleaser will automatically build and publish
```

## API Reference

### Types

#### `Manager`
Main manager for miner operations.

#### `Miner`
Represents a mining instance with ID, name, status, and performance metrics.

#### `MinerConfig`
Configuration for starting a new miner.

### Functions

- `NewManager() *Manager` - Create a new miner manager
- `StartMiner(config MinerConfig) (*Miner, error)` - Start a new miner
- `StopMiner(id string) error` - Stop a running miner
- `GetMiner(id string) (*Miner, error)` - Get miner by ID
- `ListMiners() []*Miner` - List all miners
- `UpdateHashRate(id string, hashRate float64) error` - Update miner hash rate

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the EUPL-1.2 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI functionality
- Configured for [CodeRabbit](https://coderabbit.ai) automated reviews
- Releases managed with [GoReleaser](https://goreleaser.com)

## Support

For issues, questions, or contributions, please open an issue on GitHub.
