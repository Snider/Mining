# Installation

This guide will help you install Mining Platform on your system.

## System Requirements

### Minimum Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go**: Version 1.24 or higher (for building from source)
- **RAM**: 2GB minimum, 4GB recommended
- **Storage**: 1GB free space

### For GPU Mining

- **OpenCL SDK**: For AMD GPU support
- **CUDA Toolkit**: For NVIDIA GPU support
- **GPU Drivers**: Latest drivers for your GPU

### For Development

- **Node.js**: Version 20 or higher
- **CMake**: Version 3.21 or higher
- **Make**: For build automation

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/Snider/Mining/releases).

#### Linux

```bash
# Download the binary
wget https://github.com/Snider/Mining/releases/latest/download/miner-ctrl-linux-amd64

# Make it executable
chmod +x miner-ctrl-linux-amd64

# Move to PATH
sudo mv miner-ctrl-linux-amd64 /usr/local/bin/miner-ctrl
```

#### macOS

```bash
# Download the binary
curl -L -o miner-ctrl https://github.com/Snider/Mining/releases/latest/download/miner-ctrl-darwin-amd64

# Make it executable
chmod +x miner-ctrl

# Move to PATH
sudo mv miner-ctrl /usr/local/bin/
```

#### Windows

1. Download `miner-ctrl-windows-amd64.exe` from the releases page
2. Rename to `miner-ctrl.exe`
3. Add the directory to your PATH or run from the download location

### Method 2: Install via Go

If you have Go installed, you can install directly:

```bash
go install github.com/Snider/Mining/cmd/mining@latest
```

The binary will be installed to `$GOPATH/bin/mining` (typically `~/go/bin/mining`).

### Method 3: Build from Source

For the latest development version or if you want to contribute:

```bash
# Clone the repository
git clone https://github.com/Snider/Mining.git
cd Mining

# Build the CLI
make build

# The binary will be in the current directory as 'miner-ctrl'
```

## Desktop Application

### Install Pre-built Desktop App

Download the desktop application for your platform:

- **Linux**: `mining-dashboard-linux-amd64` (or `.deb`/`.rpm` packages)
- **macOS**: `mining-dashboard.app` (DMG installer)
- **Windows**: `mining-dashboard-setup.exe` (installer)

### Build Desktop App from Source

```bash
cd cmd/desktop/mining-desktop

# Install dependencies
npm install

# Build for current platform
wails3 build

# Binary will be in: bin/mining-dashboard
```

## Verify Installation

After installation, verify it's working:

```bash
# Check version
miner-ctrl --version

# Show help
miner-ctrl --help
```

You should see output similar to:

```
Mining Platform v1.0.0
A modern cryptocurrency mining management platform
```

## Configuration

### XDG Base Directories

Mining Platform follows XDG Base Directory specifications:

- **Config**: `~/.config/lethean-desktop/`
- **Data**: `~/.local/share/lethean-desktop/miners/`
- **Profiles**: `~/.config/lethean-desktop/mining_profiles.json`

### First Run Setup

On first run, Mining Platform will create the necessary directories automatically. No manual configuration is required.

## Installing Mining Software

Mining Platform can automatically install the mining software it manages:

```bash
# Install XMRig
miner-ctrl install xmrig

# Check installation status
miner-ctrl doctor
```

See the [CLI Guide](../user-guide/cli.md) for more commands.

## Next Steps

Now that you have Mining Platform installed:

1. Follow the [Quick Start Guide](quick-start.md) to begin mining
2. Read the [CLI Guide](../user-guide/cli.md) to learn the commands
3. Explore the [Web Dashboard](../user-guide/web-dashboard.md) for a visual interface

## Troubleshooting

### Permission Errors (Linux/macOS)

If you get permission errors when running commands, ensure the binary is executable:

```bash
chmod +x miner-ctrl
```

### Command Not Found

If the `miner-ctrl` command is not found, ensure it's in your PATH:

```bash
# For Go install
export PATH=$PATH:$GOPATH/bin

# Or use the full path
~/go/bin/mining --help
```

### GPU Mining Not Working

Ensure you have the appropriate SDK installed:

- **AMD GPUs**: Install OpenCL SDK and drivers
- **NVIDIA GPUs**: Install CUDA Toolkit and drivers

Check GPU detection:

```bash
miner-ctrl doctor
```

This will show which GPUs are detected and available for mining.
