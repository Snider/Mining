# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build the CLI binary
make build                  # Outputs: miner-cli

# Run tests
make test                   # Tests with race detection and coverage
go test -v ./pkg/mining/... # Run tests for specific package
go test -run TestName ./... # Run a single test

# Lint and format
make lint                   # Runs fmt, vet, and golangci-lint
make fmt                    # Format code only

# Generate Swagger docs (required after API changes)
make docs                   # Runs: swag init -g ./cmd/mining/main.go

# Development server (builds, generates docs, starts server)
make dev                    # Starts on localhost:9090

# Build for all platforms
make build-all              # Outputs to dist/

# Create release packages
make package                # Uses GoReleaser snapshot

# E2E Tests (Playwright)
make e2e                    # Run all E2E tests
make e2e-api                # Run API tests only (no browser)
make e2e-ui                 # Open Playwright UI for interactive testing
```

## Architecture Overview

### Go Backend (`pkg/mining/`)

The mining package provides a modular miner management system:

- **`mining.go`**: Core interfaces and types. The `Miner` interface defines the contract all miner implementations must follow (Install, Start, Stop, GetStats, etc.). Also contains `Config` for miner configuration and `PerformanceMetrics` for stats.

- **`miner.go`**: `BaseMiner` struct with shared functionality for all miners (binary discovery, installation from URL, archive extraction, hashrate history management). Uses XDG base directories via `github.com/adrg/xdg`.

- **`manager.go`**: `Manager` handles miner lifecycle. Maintains running miners in a map, supports autostart from config, runs background stats collection every 10 seconds. Key methods: `StartMiner()`, `StopMiner()`, `ListMiners()`.

- **`service.go`**: RESTful API using Gin. `Service` wraps the Manager and exposes HTTP endpoints under configurable namespace (default `/api/v1/mining`). Swagger docs generated via swaggo annotations.

- **`xmrig.go` / `xmrig_start.go` / `xmrig_stats.go`**: XMRig miner implementation. Downloads from GitHub releases, generates config JSON, polls local HTTP API for stats.

- **`profile_manager.go`**: Manages saved mining configurations (profiles). Stored in `~/.config/lethean-desktop/mining_profiles.json`.

- **`config_manager.go`**: Manages autostart settings and last-used configs for miners.

### CLI (`cmd/mining/`)

Cobra-based CLI. Commands in `cmd/mining/cmd/`:
- `serve` - Main command, starts REST API server with interactive shell
- `start/stop/status` - Miner control
- `install/uninstall/update` - Miner installation management
- `doctor` - Check miner installations
- `list` - List running/available miners

### Angular UI (`ui/`)

Angular 20+ frontend that builds to `mbe-mining-dashboard.js` web component. Communicates with the Go backend via the REST API.

```bash
cd ui
ng serve          # Development server on :4200
ng build          # Build to ui/dist/
ng test           # Run unit tests (Karma/Jasmine)
npm run e2e       # Run Playwright E2E tests
```

### E2E Tests (`ui/e2e/`)

Playwright-based E2E tests covering both API and UI:

- **`e2e/api/`**: API tests (no browser) - system endpoints, miners, profiles CRUD
- **`e2e/ui/`**: Browser tests - dashboard, profiles, admin, setup wizard
- **`e2e/page-objects/`**: Page object pattern for UI components
- **`e2e/fixtures/`**: Shared test data

Tests automatically start the Go backend and Angular dev server via `playwright.config.ts` webServer config.

## Key Patterns

- **Interface-based design**: `Miner` and `ManagerInterface` allow different miner implementations
- **XDG directories**: Config in `~/.config/lethean-desktop/`, data in `~/.local/share/lethean-desktop/miners/`
- **Hashrate history**: Two-tier storage - high-res (10s intervals, 5 min retention) and low-res (1 min averages, 24h retention)
- **Syslog integration**: Platform-specific logging via `syslog_unix.go` / `syslog_windows.go`

## API Documentation

When running `make dev`, Swagger UI is available at:
`http://localhost:9090/api/v1/mining/swagger/index.html`
