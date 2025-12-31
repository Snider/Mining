# API Overview

The Mining Platform provides a comprehensive RESTful API for managing cryptocurrency miners programmatically.

## Base URL

All API endpoints are prefixed with the configured namespace (default: `/api/v1/mining`):

```
http://localhost:8080/api/v1/mining
```

You can customize the namespace when starting the server:

```bash
miner-ctrl serve --namespace /custom/path
```

## Swagger Documentation

Interactive API documentation is available via Swagger UI when the server is running:

```
http://localhost:8080/api/v1/mining/swagger/index.html
```

The Swagger specification is also available in multiple formats:
- **JSON**: `/docs/swagger.json`
- **YAML**: `/docs/swagger.yaml`

## Authentication

Currently, the API does not require authentication. This is suitable for local/trusted networks.

For production deployments, consider:
- Running behind a reverse proxy with authentication
- Implementing API key authentication
- Using OAuth2/JWT for multi-user scenarios

## API Versioning

The API follows semantic versioning:

- **Current Version**: v1
- **Endpoint Format**: `/api/v{version}/mining/{endpoint}`
- **Backward Compatibility**: Maintained within major versions

## Content Type

All requests and responses use JSON:

```
Content-Type: application/json
```

## Response Format

All API responses follow a consistent structure:

### Success Response

```json
{
  "success": true,
  "data": {
    // Response data here
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Additional error context
    }
  }
}
```

## HTTP Status Codes

The API uses standard HTTP status codes:

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 204 | No Content | Request successful, no content to return |
| 400 | Bad Request | Invalid request parameters |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists or conflicting state |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily unavailable |

## Rate Limiting

Currently, there is no rate limiting implemented. For production use, consider:

- Implementing rate limits at the reverse proxy level
- Using nginx `limit_req` module
- Implementing application-level rate limiting

## CORS

Cross-Origin Resource Sharing (CORS) is enabled by default for all origins. To restrict:

```bash
# Start server with CORS restrictions (future feature)
miner-ctrl serve --cors-origin "https://example.com"
```

## WebSocket Support

Real-time updates are available via WebSocket connection:

```
ws://localhost:8080/api/v1/mining/ws
```

The WebSocket provides:
- Real-time miner statistics
- Live hashrate updates
- Event notifications (miner start/stop/crash)

Example JavaScript client:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/mining/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
};
```

## Pagination

List endpoints support pagination via query parameters:

```
GET /api/v1/mining/miners?page=1&limit=10
```

Response includes pagination metadata:

```json
{
  "success": true,
  "data": {
    "items": [ /* ... */ ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 42,
      "pages": 5
    }
  }
}
```

## Filtering and Sorting

List endpoints support filtering and sorting:

```
GET /api/v1/mining/miners?status=running&sort=hashrate&order=desc
```

Common parameters:
- `status`: Filter by status (running, stopped, error)
- `type`: Filter by miner type (xmrig, etc.)
- `sort`: Sort field (name, hashrate, uptime)
- `order`: Sort order (asc, desc)

## Date and Time Format

All timestamps use ISO 8601 format with UTC timezone:

```
2025-12-31T23:59:59Z
```

## Field Naming

API fields use camelCase naming:

```json
{
  "minerName": "xmrig",
  "hashRate": 4520,
  "acceptedShares": 42
}
```

## Data Models

### SystemInfo

System information and installed miners.

```json
{
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
```

### Config

Miner configuration object.

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

### PerformanceMetrics

Real-time miner statistics.

```json
{
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
  },
  "gpu": [
    {
      "id": 0,
      "name": "NVIDIA GeForce RTX 3080",
      "hashrate": 95234.5,
      "temperature": 68.5,
      "fanSpeed": 75,
      "powerUsage": 220
    }
  ]
}
```

### HashratePoint

Historical hashrate data point.

```json
{
  "timestamp": "2025-12-31T12:00:00Z",
  "hashrate": 4520.5,
  "resolution": "10s"
}
```

### MiningProfile

Saved mining configuration.

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "XMR - SupportXMR",
  "description": "Monero mining on SupportXMR pool",
  "minerType": "xmrig",
  "config": { /* Config object */ },
  "createdAt": "2025-12-31T10:00:00Z",
  "updatedAt": "2025-12-31T12:00:00Z"
}
```

## Error Codes

Common error codes returned by the API:

| Code | Description |
|------|-------------|
| `MINER_NOT_FOUND` | Specified miner not found |
| `MINER_ALREADY_RUNNING` | Miner is already running |
| `MINER_NOT_INSTALLED` | Miner software not installed |
| `INVALID_CONFIG` | Invalid configuration provided |
| `POOL_UNREACHABLE` | Cannot connect to mining pool |
| `PROFILE_NOT_FOUND` | Mining profile not found |
| `PROFILE_ALREADY_EXISTS` | Profile with same name exists |
| `INVALID_WALLET` | Invalid wallet address format |
| `INSUFFICIENT_RESOURCES` | System resources insufficient |

## Examples

See the [Endpoints](endpoints.md) page for detailed examples of each API endpoint.

## Client Libraries

### JavaScript/TypeScript

```javascript
const BASE_URL = 'http://localhost:8080/api/v1/mining';

async function startMiner(config) {
  const response = await fetch(`${BASE_URL}/miners/xmrig`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  });
  return response.json();
}

async function getMiners() {
  const response = await fetch(`${BASE_URL}/miners`);
  return response.json();
}
```

### Python

```python
import requests

BASE_URL = 'http://localhost:8080/api/v1/mining'

def start_miner(config):
    response = requests.post(
        f'{BASE_URL}/miners/xmrig',
        json=config
    )
    return response.json()

def get_miners():
    response = requests.get(f'{BASE_URL}/miners')
    return response.json()
```

### Go

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
)

const baseURL = "http://localhost:8080/api/v1/mining"

func startMiner(config map[string]interface{}) error {
    data, _ := json.Marshal(config)
    resp, err := http.Post(
        baseURL+"/miners/xmrig",
        "application/json",
        bytes.NewBuffer(data),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}
```

### cURL

```bash
# Start miner
curl -X POST http://localhost:8080/api/v1/mining/miners/xmrig \
  -H "Content-Type: application/json" \
  -d '{"pool":"stratum+tcp://pool.supportxmr.com:3333","wallet":"YOUR_WALLET","algo":"rx/0"}'

# Get miners
curl http://localhost:8080/api/v1/mining/miners

# Get stats
curl http://localhost:8080/api/v1/mining/miners/xmrig/stats
```

## Next Steps

- Browse the [API Endpoints](endpoints.md) for detailed documentation
- Try the interactive [Swagger UI](http://localhost:8080/api/v1/mining/swagger/index.html)
- See the [Development Guide](../development/index.md) for contributing
