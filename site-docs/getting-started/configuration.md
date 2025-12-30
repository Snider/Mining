# Configuration

The Mining Dashboard uses XDG base directories for configuration and data storage.

## Directory Structure

```
~/.config/lethean-desktop/
├── miners.json              # Autostart and general settings
├── mining_profiles.json     # Saved mining profiles
├── node.json               # P2P node identity
└── peers.json              # P2P peer registry

~/.local/share/lethean-desktop/
├── miners/                  # Installed miner binaries
│   ├── xmrig/
│   │   └── xmrig-6.25.0/
│   └── tt-miner/
├── mining.db               # SQLite hashrate history
└── node/
    └── private.key         # P2P private key
```

## miners.json

Controls autostart and database settings:

```json
{
  "miners": [
    {
      "minerType": "xmrig",
      "autostart": true,
      "config": {
        "pool": "pool.supportxmr.com:3333",
        "wallet": "4xxx...",
        "tls": true
      }
    }
  ],
  "database": {
    "enabled": true,
    "retentionDays": 30
  }
}
```

### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `miners[].minerType` | string | - | Miner type (xmrig, tt-miner) |
| `miners[].autostart` | bool | false | Start on server launch |
| `miners[].config` | object | - | Miner configuration |
| `database.enabled` | bool | true | Enable SQLite persistence |
| `database.retentionDays` | int | 30 | Days to keep history |

## mining_profiles.json

Stores saved mining profiles:

```json
[
  {
    "id": "abc123",
    "name": "My XMR Pool",
    "minerType": "xmrig",
    "config": {
      "pool": "pool.supportxmr.com:3333",
      "wallet": "4xxx...",
      "tls": true,
      "hugePages": true,
      "threads": 0
    }
  }
]
```

### Profile Config Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `pool` | string | - | Pool address (host:port) |
| `wallet` | string | - | Wallet address |
| `password` | string | "x" | Pool password |
| `tls` | bool | false | Enable TLS encryption |
| `hugePages` | bool | true | Enable huge pages (Linux) |
| `threads` | int | 0 | CPU threads (0=auto) |
| `devices` | string | "" | GPU devices (tt-miner) |
| `algo` | string | "" | Algorithm override |
| `intensity` | int | 0 | Mining intensity (GPU) |
| `cliArgs` | string | "" | Extra CLI arguments |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MINING_API_PORT` | 9090 | REST API port |
| `MINING_P2P_PORT` | 9091 | P2P WebSocket port |
| `XDG_CONFIG_HOME` | ~/.config | Config directory |
| `XDG_DATA_HOME` | ~/.local/share | Data directory |

## Command Line Flags

```bash
./miner-ctrl serve --help

Flags:
  -p, --port int      API port (default 9090)
  -n, --namespace     API namespace (default /api/v1/mining)
      --no-autostart  Disable autostart
```

## Database Settings

The SQLite database stores hashrate history for graphing:

- **Location**: `~/.local/share/lethean-desktop/mining.db`
- **Default retention**: 30 days
- **Polling interval**: 10 seconds (high-res), 1 minute (low-res)

To disable database persistence:

```json
{
  "database": {
    "enabled": false
  }
}
```

## Node Identity (P2P)

The P2P node identity is auto-generated on first run:

```bash
# Initialize with custom name
./miner-ctrl node init --name "my-rig" --role worker
```

See [P2P Multi-Node](../features/p2p-multinode.md) for more details.
