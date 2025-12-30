# Building from Source

Complete guide to building the Mining Dashboard from source.

## Prerequisites

### Required

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Backend compilation |
| Node.js | 20+ | Frontend build |
| npm | 10+ | Package management |
| Make | any | Build automation |

### Optional

| Tool | Purpose |
|------|---------|
| golangci-lint | Code linting |
| swag | Swagger doc generation |
| Docker | Containerized builds |

## Quick Build

```bash
# Clone repository
git clone https://github.com/Snider/Mining.git
cd Mining

# Build everything
make build

# Output: ./miner-ctrl
```

## Backend Build

### Standard Build

```bash
make build
# or
go build -o miner-ctrl ./cmd/mining
```

### With Version Info

```bash
VERSION=1.0.0
go build -ldflags "-X main.version=$VERSION" -o miner-ctrl ./cmd/mining
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all

# Or manually:
GOOS=linux GOARCH=amd64 go build -o dist/miner-ctrl-linux-amd64 ./cmd/mining
GOOS=darwin GOARCH=amd64 go build -o dist/miner-ctrl-darwin-amd64 ./cmd/mining
GOOS=windows GOARCH=amd64 go build -o dist/miner-ctrl-windows-amd64.exe ./cmd/mining
```

## Frontend Build

```bash
cd ui

# Install dependencies
npm install

# Development build
ng build

# Production build
ng build --configuration production

# Output: ui/dist/
```

## Generate Documentation

### Swagger Docs

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate
make docs
# or
swag init -g ./cmd/mining/main.go
```

### MkDocs Site

```bash
# Create virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Install MkDocs
pip install mkdocs-material mkdocs-glightbox

# Serve locally
mkdocs serve

# Build static site
mkdocs build
```

## Running Tests

### Unit Tests

```bash
# All tests with coverage
make test

# Specific package
go test -v ./pkg/mining/...

# With race detection
go test -race ./...
```

### E2E Tests

```bash
cd ui

# Install Playwright
npx playwright install

# Run all E2E tests
npm run e2e

# API tests only (faster)
npm run e2e:api

# Interactive UI mode
npm run e2e:ui
```

## Development Server

```bash
# Start backend + frontend
make dev

# Or separately:
# Terminal 1: Backend
./miner-ctrl serve

# Terminal 2: Frontend
cd ui && ng serve
```

Access:
- Frontend: http://localhost:4200
- Backend API: http://localhost:9090/api/v1/mining
- Swagger UI: http://localhost:9090/api/v1/mining/swagger/index.html

## Docker Build

### Single Binary

```bash
docker build -t mining-cli .
docker run -p 9090:9090 mining-cli serve
```

### Multi-Node Setup

```bash
docker-compose -f docker-compose.p2p.yml up
```

## Release Build

```bash
# Using GoReleaser
make package

# Creates:
# - dist/miner-ctrl_linux_amd64.tar.gz
# - dist/miner-ctrl_darwin_amd64.tar.gz
# - dist/miner-ctrl_windows_amd64.zip
```

## Troubleshooting

### CGO Issues

SQLite requires CGO. If you get errors:

```bash
# Enable CGO
CGO_ENABLED=1 go build ./cmd/mining
```

### Node Modules

If frontend build fails:

```bash
cd ui
rm -rf node_modules package-lock.json
npm install
```

### Swagger Generation

If swagger fails:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g ./cmd/mining/main.go
```
