# CLI User Guide

The `miner-ctrl` command-line interface provides complete control over Mining Platform from your terminal.

## Installation

See the [Installation Guide](../getting-started/index.md) for installation instructions.

## Global Flags

These flags work with any command:

- `--config string`: Config file path (default: `$HOME/.mining.yaml`)
- `--help`: Show help for the command

## Commands

### serve

Start the mining service and interactive shell.

**Usage:**
```bash
miner-ctrl serve [flags]
```

**Flags:**
- `--host string`: Host to listen on (default: `0.0.0.0`)
- `-p, --port int`: Port to listen on (default: `8080`)
- `-n, --namespace string`: API namespace (default: `/api/v1/mining`)

**Examples:**

```bash
# Start on localhost:9090
miner-ctrl serve --host localhost --port 9090

# Start with custom namespace
miner-ctrl serve --namespace /mining/v2

# Start with interactive shell
miner-ctrl serve
```

When the server is running, you can access:
- REST API: `http://localhost:8080/api/v1/mining/`
- Swagger UI: `http://localhost:8080/api/v1/mining/swagger/index.html`

### start

Start a new miner instance.

**Usage:**
```bash
miner-ctrl start [miner-type] [flags]
```

**Arguments:**
- `miner-type`: Type of miner to start (e.g., `xmrig`)

**Flags:**
- `--config string`: Path to configuration file
- `--pool string`: Mining pool URL
- `--wallet string`: Wallet address
- `--algo string`: Mining algorithm
- `--threads int`: Number of threads to use

**Examples:**

```bash
# Start with config file
miner-ctrl start xmrig --config xmr-config.json

# Start with inline parameters
miner-ctrl start xmrig \
  --pool stratum+tcp://pool.supportxmr.com:3333 \
  --wallet YOUR_WALLET_ADDRESS \
  --algo rx/0 \
  --threads 4

# Start GPU mining
miner-ctrl start xmrig \
  --pool stratum+tcp://etc.woolypooly.com:3333 \
  --wallet YOUR_ETC_WALLET \
  --algo etchash \
  --cuda
```

### stop

Stop a running miner instance.

**Usage:**
```bash
miner-ctrl stop [miner-name]
```

**Examples:**

```bash
# Stop a miner by name
miner-ctrl stop xmrig

# Stop all miners
miner-ctrl stop --all
```

### status

Get the status and statistics of a running miner.

**Usage:**
```bash
miner-ctrl status [miner-name]
```

**Output includes:**
- Hashrate (current and average)
- Accepted/rejected shares
- Uptime
- Temperature (if supported)
- Power usage (if supported)

**Examples:**

```bash
# Get status of specific miner
miner-ctrl status xmrig

# Get JSON output for scripting
miner-ctrl status xmrig --json
```

### list

List running miners and available miner types.

**Usage:**
```bash
miner-ctrl list [flags]
```

**Flags:**
- `--running`: Show only running miners
- `--available`: Show only available miner types

**Examples:**

```bash
# List all running miners
miner-ctrl list

# Show available miner types
miner-ctrl list --available

# Detailed output
miner-ctrl list --verbose
```

### install

Install or update a miner binary.

**Usage:**
```bash
miner-ctrl install [miner-type] [flags]
```

**Flags:**
- `--force`: Force reinstall even if already installed
- `--version string`: Install specific version

**Examples:**

```bash
# Install XMRig
miner-ctrl install xmrig

# Install specific version
miner-ctrl install xmrig --version 6.21.0

# Force reinstall
miner-ctrl install xmrig --force
```

Miners are installed to `~/.local/share/lethean-desktop/miners/[miner-type]/`.

### uninstall

Uninstall a miner and remove its files.

**Usage:**
```bash
miner-ctrl uninstall [miner-type]
```

**Examples:**

```bash
# Uninstall XMRig
miner-ctrl uninstall xmrig

# Uninstall with confirmation
miner-ctrl uninstall xmrig --confirm
```

### update

Check for updates to installed miners.

**Usage:**
```bash
miner-ctrl update [flags]
```

**Flags:**
- `--check-only`: Only check for updates, don't install
- `--all`: Update all miners with available updates

**Examples:**

```bash
# Check for updates
miner-ctrl update --check-only

# Update all miners
miner-ctrl update --all

# Update specific miner
miner-ctrl update xmrig
```

### doctor

Check the status of all installed miners and system configuration.

**Usage:**
```bash
miner-ctrl doctor
```

**Output includes:**
- Installed miners and versions
- GPU detection (OpenCL/CUDA)
- System resources
- Configuration file locations
- Potential issues and recommendations

**Examples:**

```bash
# Run full diagnostic
miner-ctrl doctor

# Output to file for troubleshooting
miner-ctrl doctor > system-info.txt
```

### completion

Generate shell completion scripts.

**Usage:**
```bash
miner-ctrl completion [shell]
```

**Supported shells:**
- `bash`
- `zsh`
- `fish`
- `powershell`

**Examples:**

```bash
# Bash completion
miner-ctrl completion bash > /etc/bash_completion.d/miner-ctrl

# Zsh completion
miner-ctrl completion zsh > "${fpath[1]}/_miner-ctrl"

# Fish completion
miner-ctrl completion fish > ~/.config/fish/completions/miner-ctrl.fish
```

## Configuration File

Create a config file at `~/.mining.yaml`:

```yaml
server:
  host: localhost
  port: 9090
  namespace: /api/v1/mining

defaults:
  miner: xmrig
  threads: 4
  cpuPriority: 3

profiles:
  - name: XMR
    pool: stratum+tcp://pool.supportxmr.com:3333
    wallet: YOUR_XMR_WALLET
    algo: rx/0

  - name: ETC
    pool: stratum+tcp://etc.woolypooly.com:3333
    wallet: YOUR_ETC_WALLET
    algo: etchash
```

Load a profile:

```bash
miner-ctrl start --profile XMR
```

## Interactive Shell

When you run `miner-ctrl serve` without backgrounding it, you get an interactive shell:

```
Mining Platform v1.0.0
API Server running on http://localhost:9090
Type 'help' for available commands

mining> help
Available commands:
  list      - List running miners
  start     - Start a miner
  stop      - Stop a miner
  status    - Get miner status
  profiles  - Manage profiles
  exit      - Exit shell

mining> list
Running miners:
  xmrig - Hashrate: 4520 H/s, Shares: 42/0

mining> status xmrig
Miner: xmrig
Status: Running
Hashrate: 4520 H/s
Accepted Shares: 42
Rejected Shares: 0
Uptime: 2h 15m
```

## Output Formats

Most commands support multiple output formats:

```bash
# Human-readable (default)
miner-ctrl list

# JSON for scripting
miner-ctrl list --output json

# YAML
miner-ctrl status xmrig --output yaml

# Table format
miner-ctrl list --output table
```

## Environment Variables

Configure behavior with environment variables:

```bash
# Set config file location
export MINING_CONFIG=~/.config/mining.yaml

# Set data directory
export MINING_DATA_DIR=~/.local/share/mining

# Set log level
export MINING_LOG_LEVEL=debug

# Start server
miner-ctrl serve
```

## Examples

### Basic Mining Workflow

```bash
# 1. Install miner
miner-ctrl install xmrig

# 2. Verify installation
miner-ctrl doctor

# 3. Start mining
miner-ctrl start xmrig \
  --pool stratum+tcp://pool.supportxmr.com:3333 \
  --wallet YOUR_WALLET \
  --algo rx/0

# 4. Monitor
miner-ctrl status xmrig

# 5. Stop when done
miner-ctrl stop xmrig
```

### Dual Mining (CPU + GPU)

```bash
# Start CPU mining for Monero
miner-ctrl start xmrig-cpu \
  --pool stratum+tcp://pool.supportxmr.com:3333 \
  --wallet YOUR_XMR_WALLET \
  --algo rx/0 \
  --threads 4

# Start GPU mining for Ethereum Classic
miner-ctrl start xmrig-gpu \
  --pool stratum+tcp://etc.woolypooly.com:3333 \
  --wallet YOUR_ETC_WALLET \
  --algo etchash \
  --cuda

# Monitor both
miner-ctrl list
```

### Automated Monitoring

```bash
# Create a monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
while true; do
  clear
  echo "=== Mining Status ==="
  miner-ctrl status xmrig --output json | jq '.hashrate, .shares'
  sleep 10
done
EOF

chmod +x monitor.sh
./monitor.sh
```

## Troubleshooting

### Command Not Found

Ensure `miner-ctrl` is in your PATH:

```bash
which miner-ctrl
# If not found, add to PATH or use full path
```

### Permission Denied

On Linux, you may need to run with appropriate permissions:

```bash
# Grant execute permission
chmod +x miner-ctrl

# Or run with sudo if needed (not recommended)
sudo miner-ctrl doctor
```

### Miner Won't Start

Check logs for errors:

```bash
# Enable debug logging
miner-ctrl --log-level debug start xmrig ...

# Check system journal (Linux)
journalctl -u mining-service
```

### Port Already in Use

If port 8080 is already in use:

```bash
# Use a different port
miner-ctrl serve --port 9090
```

## Next Steps

- Explore the [Web Dashboard](web-dashboard.md) for visual management
- Try the [Desktop Application](desktop-app.md) for a native experience
- Read the [API Documentation](../api/endpoints.md) for integration
- See [Pool Integration Guide](../reference/pools.md) for pool recommendations
