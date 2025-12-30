# Mining CLI Documentation

The `miner-ctrl` is the command-line interface for the Mining project. It allows you to manage miners directly from the terminal or start a REST API server.

## Installation

```bash
go install github.com/Snider/Mining/cmd/mining@latest
```

## Global Flags

- `--config string`: Config file (default is $HOME/.mining.yaml)
- `--help`: Help for the command

## Commands

### `serve`
Starts the mining service and interactive shell.

**Usage:**
```bash
miner-ctrl serve [flags]
```

**Flags:**
- `--host`: Host to listen on (default "0.0.0.0")
- `-p, --port`: Port to listen on (default 8080)
- `-n, --namespace`: API namespace for the swagger UI (default "/")

### `start`
Start a new miner.

**Usage:**
```bash
miner-ctrl start [miner-type] [flags]
```

### `stop`
Stop a running miner.

**Usage:**
```bash
miner-ctrl stop [miner-name]
```

### `status`
Get status of a running miner.

**Usage:**
```bash
miner-ctrl status [miner-name]
```

### `list`
List running and available miners.

**Usage:**
```bash
miner-ctrl list
```

### `install`
Install or update a miner.

**Usage:**
```bash
miner-ctrl install [miner-type]
```

### `uninstall`
Uninstall a miner.

**Usage:**
```bash
miner-ctrl uninstall [miner-type]
```

### `update`
Check for updates to installed miners.

**Usage:**
```bash
miner-ctrl update
```

### `doctor`
Check and refresh the status of installed miners.

**Usage:**
```bash
miner-ctrl doctor
```

### `completion`
Generate the autocompletion script for the specified shell (bash, zsh, fish, powershell).

**Usage:**
```bash
miner-ctrl completion [shell]
```
