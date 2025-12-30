# API Endpoints

Complete reference for all REST API endpoints.

## System

### Get System Information

```http
GET /api/v1/mining/info
```

Returns system and miner installation details.

**Response:**
```json
{
  "platform": "linux",
  "cpu": "AMD Ryzen 9 5950X",
  "cores": 32,
  "memory_gb": 64,
  "installed_miners_info": [
    {
      "is_installed": true,
      "version": "6.25.0",
      "path": "/home/user/.local/share/lethean-desktop/miners/xmrig",
      "miner_binary": "xmrig"
    }
  ]
}
```

---

## Miners

### List Running Miners

```http
GET /api/v1/mining/miners
```

Returns all currently running miner instances.

**Response:**
```json
[
  {
    "name": "xmrig-123",
    "running": true,
    "full_stats": { ... }
  }
]
```

### List Available Miners

```http
GET /api/v1/mining/miners/available
```

Returns miners that can be installed.

**Response:**
```json
[
  {
    "name": "xmrig",
    "description": "XMRig CPU/GPU miner"
  },
  {
    "name": "tt-miner",
    "description": "TT-Miner NVIDIA GPU miner"
  }
]
```

### Stop a Miner

```http
DELETE /api/v1/mining/miners/{miner_name}
```

Stops a running miner instance.

**Parameters:**
- `miner_name` (path) - Name of the miner instance (e.g., "xmrig-123")

**Response:**
```json
{"message": "Miner stopped"}
```

### Get Miner Stats

```http
GET /api/v1/mining/miners/{miner_name}/stats
```

Returns performance metrics for a specific miner.

**Response:**
```json
{
  "hashrate": 1234,
  "shares": 42,
  "rejected": 1,
  "uptime": 3600,
  "algorithm": "rx/0",
  "avgDifficulty": 100000,
  "diffCurrent": 100000
}
```

### Get Miner Logs

```http
GET /api/v1/mining/miners/{miner_name}/logs
```

Returns base64-encoded log lines.

**Response:**
```json
[
  "W1hNUmlnXSBzcGVlZCAxMHMvNjBzLzE1bQ==",
  "W1hNUmlnXSBhY2NlcHRlZA=="
]
```

### Send Stdin Command

```http
POST /api/v1/mining/miners/{miner_name}/stdin
```

Sends input to the miner's stdin.

**Request:**
```json
{"input": "h"}
```

**Response:**
```json
{"status": "sent", "input": "h"}
```

### Get Hashrate History

```http
GET /api/v1/mining/miners/{miner_name}/hashrate-history
```

Returns in-memory hashrate history (last 5 minutes).

**Response:**
```json
[
  {"timestamp": "2024-01-15T10:30:00Z", "hashrate": 1234},
  {"timestamp": "2024-01-15T10:30:10Z", "hashrate": 1256}
]
```

---

## Installation

### Install Miner

```http
POST /api/v1/mining/miners/{miner_type}/install
```

Downloads and installs a miner.

**Parameters:**
- `miner_type` (path) - Type of miner ("xmrig" or "tt-miner")

**Response:**
```json
{"message": "Miner installed successfully"}
```

### Uninstall Miner

```http
DELETE /api/v1/mining/miners/{miner_type}/uninstall
```

Removes an installed miner.

**Response:**
```json
{"message": "Miner uninstalled"}
```

---

## Profiles

### List Profiles

```http
GET /api/v1/mining/profiles
```

Returns all saved mining profiles.

**Response:**
```json
[
  {
    "id": "abc123",
    "name": "My Profile",
    "minerType": "xmrig",
    "config": {
      "pool": "pool.example.com:3333",
      "wallet": "4xxx..."
    }
  }
]
```

### Create Profile

```http
POST /api/v1/mining/profiles
```

Creates a new mining profile.

**Request:**
```json
{
  "name": "My Profile",
  "minerType": "xmrig",
  "config": {
    "pool": "pool.example.com:3333",
    "wallet": "4xxx...",
    "tls": true
  }
}
```

**Response:**
```json
{
  "id": "abc123",
  "name": "My Profile",
  ...
}
```

### Update Profile

```http
PUT /api/v1/mining/profiles/{id}
```

Updates an existing profile.

### Delete Profile

```http
DELETE /api/v1/mining/profiles/{id}
```

Removes a profile.

### Start Miner from Profile

```http
POST /api/v1/mining/profiles/{id}/start
```

Starts a miner using the profile configuration.

**Response:**
```json
{
  "name": "xmrig-456",
  "message": "Miner started"
}
```

---

## Historical Data

### Get All Miners History

```http
GET /api/v1/mining/history/miners?since={timestamp}&until={timestamp}
```

**Parameters:**
- `since` (query) - Start time (ISO 8601)
- `until` (query) - End time (ISO 8601)

### Get Miner Historical Stats

```http
GET /api/v1/mining/history/miners/{miner_name}?since={timestamp}
```

### Get Miner Historical Hashrate

```http
GET /api/v1/mining/history/miners/{miner_name}/hashrate?since={timestamp}&until={timestamp}
```

**Response:**
```json
[
  {"timestamp": "2024-01-15T10:30:00Z", "hashrate": 1234},
  {"timestamp": "2024-01-15T10:31:00Z", "hashrate": 1256}
]
```

---

## P2P / Nodes

### Get Node Info

```http
GET /api/v1/mining/node/info
```

Returns local node identity.

### List Peers

```http
GET /api/v1/mining/peers
```

### Add Peer

```http
POST /api/v1/mining/peers
```

**Request:**
```json
{
  "name": "rig-alpha",
  "address": "192.168.1.100:9091"
}
```

### Remove Peer

```http
DELETE /api/v1/mining/peers/{id}
```

### Ping Peer

```http
POST /api/v1/mining/peers/{id}/ping
```

---

## Remote Operations

### Get All Remote Stats

```http
GET /api/v1/mining/remote/stats
```

### Get Peer Stats

```http
GET /api/v1/mining/remote/{peerId}/stats
```

### Start Remote Miner

```http
POST /api/v1/mining/remote/{peerId}/start
```

**Request:**
```json
{"profileId": "abc123"}
```

### Stop Remote Miner

```http
POST /api/v1/mining/remote/{peerId}/stop
```

**Request:**
```json
{"minerName": "xmrig-123"}
```

### Get Remote Logs

```http
GET /api/v1/mining/remote/{peerId}/logs/{minerName}
```
