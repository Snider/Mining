# Quick Start Guide

Get up and running with Mining Platform in just a few minutes.

## Prerequisites

Ensure you have completed the [Installation](index.md) steps and have `miner-ctrl` installed.

## Step 1: Install Mining Software

First, install the miner software you want to use. For this guide, we'll use XMRig for Monero mining:

```bash
miner-ctrl install xmrig
```

This will download and install XMRig to `~/.local/share/lethean-desktop/miners/xmrig/`.

## Step 2: Start the Mining Service

Start the Mining Platform API server:

```bash
miner-ctrl serve --host localhost --port 9090
```

This starts:
- REST API server on `http://localhost:9090`
- Swagger UI at `http://localhost:9090/api/v1/mining/swagger/index.html`
- Interactive shell for quick commands

## Step 3: Configure Your First Miner

You can configure mining in two ways:

### Option A: Using the CLI

Create a configuration file `xmr-config.json`:

```json
{
  "pool": "stratum+tcp://pool.supportxmr.com:3333",
  "wallet": "YOUR_MONERO_WALLET_ADDRESS",
  "algo": "rx/0",
  "threads": 4
}
```

Start mining:

```bash
miner-ctrl start xmrig --config xmr-config.json
```

### Option B: Using the API

Send a POST request to start mining:

```bash
curl -X POST http://localhost:9090/api/v1/mining/miners/xmrig \
  -H "Content-Type: application/json" \
  -d '{
    "pool": "stratum+tcp://pool.supportxmr.com:3333",
    "wallet": "YOUR_MONERO_WALLET_ADDRESS",
    "algo": "rx/0",
    "threads": 4
  }'
```

## Step 4: Monitor Your Miner

### Check Status

```bash
# List running miners
miner-ctrl list

# Get detailed statistics
miner-ctrl status xmrig
```

### View in Dashboard

Open your browser to `http://localhost:9090` to access the web dashboard, where you can see:

- Real-time hashrate
- Accepted/rejected shares
- Uptime and performance metrics
- Temperature and power usage (if supported)

## Step 5: Save Your Configuration as a Profile

Save your mining configuration for easy reuse:

```bash
curl -X POST http://localhost:9090/api/v1/mining/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "XMR Mining - SupportXMR",
    "minerType": "xmrig",
    "config": {
      "pool": "stratum+tcp://pool.supportxmr.com:3333",
      "wallet": "YOUR_MONERO_WALLET_ADDRESS",
      "algo": "rx/0",
      "threads": 4
    }
  }'
```

Profiles are saved to `~/.config/lethean-desktop/mining_profiles.json`.

## Common Mining Configurations

### Monero (XMR) - CPU Mining

```json
{
  "pool": "stratum+tcp://pool.supportxmr.com:3333",
  "wallet": "YOUR_XMR_WALLET",
  "algo": "rx/0",
  "threads": 4,
  "cpuPriority": 3
}
```

### Ethereum Classic (ETC) - GPU Mining

```json
{
  "pool": "stratum+tcp://etc.woolypooly.com:3333",
  "wallet": "YOUR_ETC_WALLET",
  "algo": "etchash",
  "cuda": {
    "enabled": true,
    "devices": [0, 1]
  }
}
```

### Ravencoin (RVN) - GPU Mining

```json
{
  "pool": "stratum+tcp://rvn.woolypooly.com:3333",
  "wallet": "YOUR_RVN_WALLET",
  "algo": "kawpow",
  "opencl": {
    "enabled": true,
    "devices": [0]
  }
}
```

## Stopping a Miner

```bash
# Via CLI
miner-ctrl stop xmrig

# Via API
curl -X DELETE http://localhost:9090/api/v1/mining/miners/xmrig
```

## Updating Mining Software

Keep your mining software up to date:

```bash
# Check for updates
miner-ctrl update

# Update a specific miner
miner-ctrl install xmrig --force
```

## Desktop Application Quick Start

If you're using the desktop application instead of the CLI:

1. Launch the Mining Dashboard app
2. Click "Install Miner" and select XMRig
3. Go to "Setup Wizard" to configure your first miner
4. Enter your pool URL and wallet address
5. Click "Start Mining"

The desktop app provides the same functionality as the CLI with a graphical interface.

## Pool Recommendations

For beginners, we recommend these pools:

### Monero (XMR)

- **SupportXMR**: `pool.supportxmr.com:3333` (0.6% fee, no registration)
- **P2Pool**: `p2pool.io:3333` (0% fee, decentralized)
- **Nanopool**: `xmr-eu1.nanopool.org:14433` (1.0% fee, mobile app)

### Ethereum Classic (ETC)

- **WoolyPooly**: `etc.woolypooly.com:3333` (0.5% fee)
- **Nanopool**: `etc-eu1.nanopool.org:19999` (1.0% fee)

### Ravencoin (RVN)

- **WoolyPooly**: `rvn.woolypooly.com:3333` (0.5% fee)
- **Flypool**: `rvn.flypool.org:3333` (1.0% fee)

See the [Pool Integration Guide](../reference/pools.md) for comprehensive pool information.

## Next Steps

Now that you're mining:

1. Learn all [CLI commands](../user-guide/cli.md)
2. Explore the [Web Dashboard](../user-guide/web-dashboard.md)
3. Configure [multiple profiles](../user-guide/desktop-app.md) for different coins
4. Read about [pool selection](../reference/pools.md) to optimize your earnings
5. Review the [API documentation](../api/endpoints.md) to integrate with your own apps

## Troubleshooting

### Miner Won't Start

Check the installation:

```bash
miner-ctrl doctor
```

This will verify all installed miners and show any issues.

### Low Hashrate

- Ensure your CPU isn't being throttled due to high temperatures
- Adjust the `threads` parameter (try half your CPU cores)
- Set appropriate `cpuPriority` (1-5, with 5 being highest)

### Connection Refused

Verify the pool is reachable:

```bash
telnet pool.supportxmr.com 3333
```

If the connection fails, try a different pool or port.

### Shares Being Rejected

- Verify your wallet address is correct
- Check that you're using the right algorithm for the pool
- Ensure your miner software is up to date

For more help, see the full [API documentation](../api/index.md) or visit our [GitHub Issues](https://github.com/Snider/Mining/issues).
