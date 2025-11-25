# Mining CLI Documentation

The `miner-cli` is the command-line interface for the Mining project. It allows you to manage miners directly from the terminal or start a REST API server.

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
miner-cli serve [flags]
```

**Flags:**
- `--host`: Host to listen on (default "0.0.0.0")
- `-p, --port`: Port to listen on (default 8080)
- `-n, --namespace`: API namespace for the swagger UI (default "/")

### `start`
Start a new miner.

**Usage:**
```bash
miner-cli start [miner-type] [flags]
```

### `stop`
Stop a running miner.

**Usage:**
```bash
miner-cli stop [miner-name]
```

### `status`
Get status of a running miner.

**Usage:**
```bash
miner-cli status [miner-name]
```

### `list`
List running and available miners.

**Usage:**
```bash
miner-cli list
```

### `install`
Install or update a miner.

**Usage:**
```bash
miner-cli install [miner-type]
```

### `uninstall`
Uninstall a miner.

**Usage:**
```bash
miner-cli uninstall [miner-type]
```

### `update`
Check for updates to installed miners.

**Usage:**
```bash
miner-cli update
```

### `doctor`
Check and refresh the status of installed miners.

**Usage:**
```bash
miner-cli doctor
```

### `completion`
Generate the autocompletion script for the specified shell (bash, zsh, fish, powershell).

**Usage:**
```bash
miner-cli completion [shell]
```
