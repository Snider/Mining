# CLI Examples

Practical examples for common tasks.

## Quick Start

### Start Mining with XMRig

```bash
# Install XMRig
miner-cli install xmrig

# Start mining with a profile
miner-cli start --profile "My Profile"

# Or with direct parameters
miner-cli start xmrig --pool pool.example.com:3333 --wallet 4xxx...
```

### Monitor Mining Status

```bash
# Check status of all miners
miner-cli status

# Watch status continuously
watch -n 5 miner-cli status
```

---

## Profile Management

### Create and Use Profiles

```bash
# List existing profiles
miner-cli profile list

# Start a miner from profile
miner-cli start --profile "Monero Mining"
```

---

## Multi-Node Operations

### Set Up a Controller Node

```bash
# Initialize as controller
miner-cli node init --name "control-center" --role controller

# Start the P2P server
miner-cli node serve --listen :9091
```

### Set Up a Worker Node

```bash
# Initialize as worker
miner-cli node init --name "rig-alpha" --role worker

# Start accepting connections
miner-cli node serve --listen :9091
```

### Connect and Manage Peers

```bash
# On controller: add a worker
miner-cli peer add --address 192.168.1.100:9091 --name "rig-alpha"

# List all peers
miner-cli peer list

# Ping a peer
miner-cli peer ping abc123
```

### Remote Mining Commands

```bash
# Get stats from all remote miners
miner-cli remote status

# Start miner on remote peer
miner-cli remote start abc123 --profile "My Profile"

# Stop miner on remote peer
miner-cli remote stop abc123 xmrig-456

# Get logs from remote miner
miner-cli remote logs abc123 xmrig-456 --lines 50
```

---

## Server Operations

### Run the Dashboard

```bash
# Start with defaults (port 9090)
miner-cli serve

# Custom port
miner-cli serve --port 8080

# Disable autostart
miner-cli serve --no-autostart
```

### System Health Check

```bash
# Run diagnostics
miner-cli doctor
```

Output:
```
System Check
============
Platform: linux
CPU: AMD Ryzen 9 5950X
Cores: 32
Memory: 64 GB

Miner Status
============
✓ xmrig v6.25.0 installed
✗ tt-miner not installed

Recommendations
===============
- Enable huge pages for better performance
```

---

## Scripting Examples

### Bash Script: Auto-restart on Failure

```bash
#!/bin/bash
PROFILE="My Profile"

while true; do
    miner-cli start --profile "$PROFILE"
    sleep 10

    # Check if still running
    if ! miner-cli status | grep -q "running"; then
        echo "Miner stopped, restarting..."
        continue
    fi

    sleep 60
done
```

### Monitor Hashrate via API

```bash
#!/bin/bash
while true; do
    curl -s http://localhost:9090/api/v1/mining/miners | \
        jq -r '.[] | "\(.name): \(.full_stats.hashrate.total[0] // 0) H/s"'
    sleep 10
done
```

---

## Docker Examples

### Run with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  mining-dashboard:
    image: mining-cli:latest
    command: serve --port 9090
    ports:
      - "9090:9090"
    volumes:
      - mining-data:/root/.local/share/lethean-desktop
      - mining-config:/root/.config/lethean-desktop

volumes:
  mining-data:
  mining-config:
```

```bash
docker-compose up -d
```

### Multi-Node Docker Setup

```bash
# Start controller
docker run -d --name controller \
    -p 9090:9090 -p 9091:9091 \
    mining-cli node serve

# Start workers
docker run -d --name worker1 \
    -p 9092:9091 \
    mining-cli node serve
```
