# Installation

This guide covers installing the Mining Dashboard on your system.

## Prerequisites

- **Go 1.21+** - For building the backend
- **Node.js 18+** - For building the frontend (optional, pre-built included)
- **Git** - For cloning the repository

## Quick Install

### From Source

```bash
# Clone the repository
git clone https://github.com/Snider/Mining.git
cd Mining

# Build the CLI binary
make build

# The binary is now at ./miner-ctrl
./miner-ctrl --help
```

### Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the CLI binary |
| `make build-all` | Build for all platforms (Linux, macOS, Windows) |
| `make test` | Run tests with coverage |
| `make lint` | Run linters |
| `make docs` | Generate Swagger API documentation |
| `make dev` | Start development server |

## Platform-Specific Notes

### Linux

No additional dependencies required. The miner binaries (XMRig, TT-Miner) will be automatically downloaded when you install them through the UI.

```bash
# Optional: Enable huge pages for better XMRig performance
sudo sysctl -w vm.nr_hugepages=1280
```

### macOS

Works out of the box on both Intel and Apple Silicon.

### Windows

Build with:
```powershell
go build -o miner-ctrl.exe ./cmd/mining
```

## Docker

```bash
# Build the image
docker build -t mining-dashboard .

# Run with persistent data
docker run -d \
  -p 9090:9090 \
  -v mining-data:/root/.local/share/lethean-desktop \
  -v mining-config:/root/.config/lethean-desktop \
  mining-dashboard
```

## Verify Installation

```bash
# Check version
./miner-ctrl --version

# Run doctor to check system
./miner-ctrl doctor

# Start the server
./miner-ctrl serve
```

Then open [http://localhost:9090](http://localhost:9090) to access the dashboard.

## Next Steps

- [Quick Start Guide](quickstart.md) - Get mining in 5 minutes
- [Configuration](configuration.md) - Customize your setup
