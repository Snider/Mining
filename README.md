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

### CLI Commands

The `miner-cli` provides the following commands:

```
miner-cli completion  Generate the autocompletion script for the specified shell
miner-cli doctor      Check and refresh the status of installed miners
miner-cli help        Help about any command
miner-cli install     Install or update a miner
miner-cli list        List running and available miners
miner-cli serve       Start the mining service and interactive shell
miner-cli start       Start a new miner
miner-cli status      Get status of a running miner
miner-cli stop        Stop a running miner
miner-cli uninstall   Uninstall a miner
miner-cli update      Check for updates to installed miners
```

For more details on any command, use `miner-cli [command] --help`.

### RESTful API Endpoints

When running the `miner-cli serve` command, the following RESTful API endpoints are exposed (default base path `/api/v1/mining`):

- `GET /api/v1/mining/info` - Get cached miner installation information and system details.
- `POST /api/v1/mining/doctor` - Perform a live check on all available miners to verify their installation status, version, and path.
- `POST /api/v1/mining/update` - Check if any installed miners have a new version available for download.
- `GET /api/v1/mining/miners` - Get a list of all running miners.
- `GET /api/v1/mining/miners/available` - Get a list of all available miners.
- `POST /api/v1/mining/miners/:miner_name` - Start a new miner with the given configuration.
- `POST /api/v1/mining/miners/:miner_name/install` - Install a new miner or update an existing one.
- `DELETE /api/v1/mining/miners/:miner_name/uninstall` - Remove all files for a specific miner.
- `DELETE /api/v1/mining/miners/:miner_name` - Stop a running miner by its name.
- `GET /api/v1/mining/miners/:miner_name/stats` - Get statistics for a running miner.
- `GET /api/v1/mining/swagger/*any` - Serve Swagger UI for API documentation.

Swagger documentation is typically available at `http://<host>:<port>/api/v1/mining/swagger/index.html`.

## Development

### Prerequisites

- Go 1.24 or higher
- Make (optional, for using Makefile targets)

### Build

```bash
# Build the CLI
go build -o miner-cli ./cmd/mining

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
