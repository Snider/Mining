# CLI Commands

Complete reference for the `miner-cli` command-line interface.

## Global Flags

```bash
miner-cli [command] [flags]
```

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for command |
| `--version`, `-v` | Show version |

---

## serve

Start the REST API server and web dashboard.

```bash
miner-cli serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port`, `-p` | 9090 | API server port |
| `--namespace`, `-n` | /api/v1/mining | API namespace |
| `--no-autostart` | false | Disable miner autostart |

**Examples:**

```bash
# Start with defaults
miner-cli serve

# Custom port
miner-cli serve --port 8080

# Disable autostart
miner-cli serve --no-autostart
```

---

## start

Start a miner with a specific configuration.

```bash
miner-cli start <miner-type> [flags]
```

| Flag | Description |
|------|-------------|
| `--pool`, `-o` | Pool address |
| `--wallet`, `-u` | Wallet address |
| `--threads`, `-t` | CPU threads |
| `--tls` | Enable TLS |
| `--profile` | Use profile by name/ID |

**Examples:**

```bash
# Start XMRig with pool and wallet
miner-cli start xmrig --pool pool.example.com:3333 --wallet 4xxx...

# Start using a profile
miner-cli start --profile "My Profile"

# Start TT-Miner on specific GPUs
miner-cli start tt-miner --pool pool.example.com:4444 --devices 0,1
```

---

## stop

Stop a running miner.

```bash
miner-cli stop <miner-name>
```

**Examples:**

```bash
# Stop a specific miner
miner-cli stop xmrig-123

# Stop all miners
miner-cli stop --all
```

---

## status

Show status of running miners.

```bash
miner-cli status [miner-name]
```

**Examples:**

```bash
# Show all miners
miner-cli status

# Show specific miner
miner-cli status xmrig-123
```

**Output:**
```
NAME          HASHRATE    SHARES    UPTIME    POOL
xmrig-123     1.23 kH/s   42/43     1h 23m    pool.example.com
```

---

## list

List available or running miners.

```bash
miner-cli list [flags]
```

| Flag | Description |
|------|-------------|
| `--available` | Show available miners |
| `--running` | Show running miners |
| `--installed` | Show installed miners |

---

## install

Install a miner.

```bash
miner-cli install <miner-type>
```

**Examples:**

```bash
miner-cli install xmrig
miner-cli install tt-miner
```

---

## uninstall

Uninstall a miner.

```bash
miner-cli uninstall <miner-type>
```

---

## update

Update a miner to the latest version.

```bash
miner-cli update <miner-type>
```

---

## doctor

Check system health and miner installations.

```bash
miner-cli doctor
```

**Output:**
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

## node

P2P node management commands.

### node init

Initialize node identity.

```bash
miner-cli node init [flags]
```

| Flag | Description |
|------|-------------|
| `--name` | Node name |
| `--role` | Role (controller/worker/dual) |

### node info

Show node information.

```bash
miner-cli node info
```

### node serve

Start P2P server.

```bash
miner-cli node serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | :9091 | Listen address |

---

## peer

Peer management commands.

### peer add

Add a peer node.

```bash
miner-cli peer add [flags]
```

| Flag | Description |
|------|-------------|
| `--address` | Peer address (host:port) |
| `--name` | Peer name |

### peer list

List registered peers.

```bash
miner-cli peer list
```

### peer remove

Remove a peer.

```bash
miner-cli peer remove <peer-id>
```

### peer ping

Ping a peer.

```bash
miner-cli peer ping <peer-id>
```

---

## remote

Remote miner operations.

### remote status

Get stats from remote peers.

```bash
miner-cli remote status [peer-id]
```

### remote start

Start miner on remote peer.

```bash
miner-cli remote start <peer-id> --profile <profile-id>
```

### remote stop

Stop miner on remote peer.

```bash
miner-cli remote stop <peer-id> [miner-name]
```

### remote logs

Get logs from remote miner.

```bash
miner-cli remote logs <peer-id> <miner-name> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--lines`, `-n` | 100 | Number of lines |

---

## profile

Profile management commands.

### profile list

List all profiles.

```bash
miner-cli profile list
```

### profile create

Create a new profile.

```bash
miner-cli profile create [flags]
```

### profile delete

Delete a profile.

```bash
miner-cli profile delete <profile-id>
```
