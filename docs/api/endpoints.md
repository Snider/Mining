# API Endpoints

Complete reference for all Mining Platform API endpoints.

## System Endpoints

### GET /info

Retrieve system information and installed miners.

**Response:**

```json
{
  "success": true,
  "data": {
    "os": "linux",
    "arch": "amd64",
    "goVersion": "go1.24.0",
    "totalMemory": 16777216,
    "installedMiners": [
      {
        "type": "xmrig",
        "version": "6.21.0",
        "installed": true,
        "path": "/home/user/.local/share/lethean-desktop/miners/xmrig"
      }
    ]
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/mining/info
```

### POST /doctor

Perform a live diagnostic check on all miners and system configuration.

**Response:**

```json
{
  "success": true,
  "data": {
    "miners": [
      {
        "type": "xmrig",
        "installed": true,
        "version": "6.21.0",
        "status": "ok",
        "issues": []
      }
    ],
    "gpu": {
      "opencl": {
        "available": true,
        "devices": [
          {
            "id": 0,
            "name": "NVIDIA GeForce RTX 3080",
            "memory": 10737418240
          }
        ]
      },
      "cuda": {
        "available": true,
        "version": "12.0",
        "devices": [
          {
            "id": 0,
            "name": "NVIDIA GeForce RTX 3080",
            "computeCapability": "8.6"
          }
        ]
      }
    },
    "recommendations": [
      "System configured correctly",
      "All miners up to date"
    ]
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/mining/doctor
```

### POST /update

Check for updates to installed miners.

**Query Parameters:**
- `check_only` (bool): Only check, don't install updates
- `all` (bool): Update all miners

**Response:**

```json
{
  "success": true,
  "data": {
    "updates": [
      {
        "miner": "xmrig",
        "currentVersion": "6.21.0",
        "latestVersion": "6.21.1",
        "updateAvailable": true
      }
    ]
  }
}
```

**Example:**

```bash
# Check for updates
curl -X POST "http://localhost:8080/api/v1/mining/update?check_only=true"

# Update all
curl -X POST "http://localhost:8080/api/v1/mining/update?all=true"
```

## Miner Management

### GET /miners

List all currently running miners.

**Query Parameters:**
- `status` (string): Filter by status (running, stopped, error)
- `type` (string): Filter by miner type

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "name": "xmrig",
      "type": "xmrig",
      "status": "running",
      "pid": 12345,
      "startedAt": "2025-12-31T10:00:00Z",
      "config": {
        "pool": "stratum+tcp://pool.supportxmr.com:3333",
        "wallet": "YOUR_WALLET_ADDRESS",
        "algo": "rx/0"
      }
    }
  ]
}
```

**Example:**

```bash
# List all miners
curl http://localhost:8080/api/v1/mining/miners

# Filter by status
curl "http://localhost:8080/api/v1/mining/miners?status=running"
```

### GET /miners/available

List all available miner types that can be installed.

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "type": "xmrig",
      "name": "XMRig",
      "description": "High-performance CPU/GPU miner for RandomX and CryptoNight algorithms",
      "algorithms": ["rx/0", "rx/wow", "cn/r", "cn/0"],
      "cpuSupport": true,
      "gpuSupport": true,
      "latestVersion": "6.21.1",
      "installed": true,
      "installedVersion": "6.21.0"
    }
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/mining/miners/available
```

### POST /miners/:miner_type

Start a new miner instance.

**Path Parameters:**
- `miner_type`: Type of miner to start (e.g., `xmrig`)

**Request Body:**

```json
{
  "pool": "stratum+tcp://pool.supportxmr.com:3333",
  "wallet": "YOUR_WALLET_ADDRESS",
  "algo": "rx/0",
  "threads": 4,
  "cpuPriority": 3,
  "cuda": {
    "enabled": false,
    "devices": []
  },
  "opencl": {
    "enabled": false,
    "devices": []
  }
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "name": "xmrig",
    "type": "xmrig",
    "status": "starting",
    "pid": 12345,
    "message": "Miner started successfully"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/mining/miners/xmrig \
  -H "Content-Type: application/json" \
  -d '{
    "pool": "stratum+tcp://pool.supportxmr.com:3333",
    "wallet": "YOUR_WALLET_ADDRESS",
    "algo": "rx/0",
    "threads": 4
  }'
```

### DELETE /miners/:miner_name

Stop a running miner instance.

**Path Parameters:**
- `miner_name`: Name of the miner to stop

**Response:**

```json
{
  "success": true,
  "data": {
    "message": "Miner stopped successfully",
    "name": "xmrig",
    "stoppedAt": "2025-12-31T12:00:00Z"
  }
}
```

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/mining/miners/xmrig
```

### POST /miners/:miner_type/install

Install or update a specific miner.

**Path Parameters:**
- `miner_type`: Type of miner to install

**Query Parameters:**
- `force` (bool): Force reinstall even if already installed
- `version` (string): Install specific version

**Response:**

```json
{
  "success": true,
  "data": {
    "miner": "xmrig",
    "version": "6.21.1",
    "path": "/home/user/.local/share/lethean-desktop/miners/xmrig",
    "message": "Miner installed successfully"
  }
}
```

**Example:**

```bash
# Install latest version
curl -X POST http://localhost:8080/api/v1/mining/miners/xmrig/install

# Install specific version
curl -X POST "http://localhost:8080/api/v1/mining/miners/xmrig/install?version=6.21.0"

# Force reinstall
curl -X POST "http://localhost:8080/api/v1/mining/miners/xmrig/install?force=true"
```

### DELETE /miners/:miner_type/uninstall

Uninstall a miner and remove its files.

**Path Parameters:**
- `miner_type`: Type of miner to uninstall

**Response:**

```json
{
  "success": true,
  "data": {
    "message": "Miner uninstalled successfully",
    "miner": "xmrig"
  }
}
```

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/mining/miners/xmrig/uninstall
```

### GET /miners/:miner_name/stats

Get real-time statistics for a running miner.

**Path Parameters:**
- `miner_name`: Name of the running miner

**Response:**

```json
{
  "success": true,
  "data": {
    "hashrate": 4520.5,
    "hashrateAvg": 4485.2,
    "shares": {
      "accepted": 42,
      "rejected": 0,
      "invalid": 0
    },
    "uptime": 8215,
    "connection": {
      "pool": "pool.supportxmr.com:3333",
      "uptime": 8215,
      "ping": 45,
      "failures": 0
    },
    "cpu": {
      "usage": 95.5,
      "temperature": 65.2
    }
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/mining/miners/xmrig/stats
```

### GET /miners/:miner_name/hashrate-history

Get historical hashrate data for a miner.

**Path Parameters:**
- `miner_name`: Name of the miner

**Query Parameters:**
- `resolution` (string): Data resolution (`10s`, `1m`, `5m`, `1h`)
- `duration` (string): Time duration (`5m`, `1h`, `24h`)

**Response:**

```json
{
  "success": true,
  "data": {
    "points": [
      {
        "timestamp": "2025-12-31T12:00:00Z",
        "hashrate": 4520.5
      },
      {
        "timestamp": "2025-12-31T12:00:10Z",
        "hashrate": 4535.2
      }
    ],
    "resolution": "10s",
    "duration": "5m"
  }
}
```

**Example:**

```bash
# Get last 5 minutes at 10s resolution
curl "http://localhost:8080/api/v1/mining/miners/xmrig/hashrate-history?resolution=10s&duration=5m"

# Get last 24 hours at 1m resolution
curl "http://localhost:8080/api/v1/mining/miners/xmrig/hashrate-history?resolution=1m&duration=24h"
```

## Profile Management

### GET /profiles

List all saved mining profiles.

**Query Parameters:**
- `miner_type` (string): Filter by miner type

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "XMR - SupportXMR",
      "description": "Monero mining on SupportXMR pool",
      "minerType": "xmrig",
      "config": {
        "pool": "stratum+tcp://pool.supportxmr.com:3333",
        "wallet": "YOUR_WALLET_ADDRESS",
        "algo": "rx/0"
      },
      "createdAt": "2025-12-31T10:00:00Z",
      "updatedAt": "2025-12-31T12:00:00Z"
    }
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/mining/profiles
```

### GET /profiles/:id

Get a specific mining profile by ID.

**Path Parameters:**
- `id`: Profile UUID

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "XMR - SupportXMR",
    "minerType": "xmrig",
    "config": { /* ... */ }
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/mining/profiles/550e8400-e29b-41d4-a716-446655440000
```

### POST /profiles

Create a new mining profile.

**Request Body:**

```json
{
  "name": "XMR - SupportXMR",
  "description": "Monero mining on SupportXMR pool",
  "minerType": "xmrig",
  "config": {
    "pool": "stratum+tcp://pool.supportxmr.com:3333",
    "wallet": "YOUR_WALLET_ADDRESS",
    "algo": "rx/0",
    "threads": 4
  }
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "XMR - SupportXMR",
    "message": "Profile created successfully"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/mining/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "XMR - SupportXMR",
    "minerType": "xmrig",
    "config": {
      "pool": "stratum+tcp://pool.supportxmr.com:3333",
      "wallet": "YOUR_WALLET_ADDRESS",
      "algo": "rx/0"
    }
  }'
```

### PUT /profiles/:id

Update an existing mining profile.

**Path Parameters:**
- `id`: Profile UUID

**Request Body:** Same as POST /profiles

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Profile updated successfully"
  }
}
```

**Example:**

```bash
curl -X PUT http://localhost:8080/api/v1/mining/profiles/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "XMR - Updated",
    "config": { /* ... */ }
  }'
```

### DELETE /profiles/:id

Delete a mining profile.

**Path Parameters:**
- `id`: Profile UUID

**Response:**

```json
{
  "success": true,
  "data": {
    "message": "Profile deleted successfully",
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/mining/profiles/550e8400-e29b-41d4-a716-446655440000
```

### POST /profiles/:id/start

Start mining using a saved profile.

**Path Parameters:**
- `id`: Profile UUID

**Response:**

```json
{
  "success": true,
  "data": {
    "minerName": "xmrig",
    "profile": "XMR - SupportXMR",
    "status": "starting"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/mining/profiles/550e8400-e29b-41d4-a716-446655440000/start
```

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
  "error": {
    "code": "MINER_NOT_FOUND",
    "message": "Miner 'xmrig' is not currently running",
    "details": {
      "minerName": "xmrig",
      "suggestion": "Start the miner first using POST /miners/xmrig"
    }
  }
}
```

Common error scenarios:

### 400 Bad Request

```json
{
  "success": false,
  "error": {
    "code": "INVALID_CONFIG",
    "message": "Invalid configuration provided",
    "details": {
      "field": "wallet",
      "issue": "Wallet address must be 95 characters"
    }
  }
}
```

### 404 Not Found

```json
{
  "success": false,
  "error": {
    "code": "MINER_NOT_FOUND",
    "message": "Miner 'xmrig' not found"
  }
}
```

### 409 Conflict

```json
{
  "success": false,
  "error": {
    "code": "MINER_ALREADY_RUNNING",
    "message": "Miner 'xmrig' is already running",
    "details": {
      "pid": 12345,
      "startedAt": "2025-12-31T10:00:00Z"
    }
  }
}
```

### 500 Internal Server Error

```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "details": {
      "requestId": "req_123456"
    }
  }
}
```

## Next Steps

- Try the [Swagger UI](http://localhost:8080/api/v1/mining/swagger/index.html) for interactive testing
- See [API Overview](index.md) for authentication and general information
- Check the [Development Guide](../development/index.md) for contributing
