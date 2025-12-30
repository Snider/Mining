# REST API

The Mining Dashboard exposes a RESTful API for programmatic access.

## Base URL

```
http://localhost:9090/api/v1/mining
```

## Authentication

Currently, the API does not require authentication. It's designed to run on a local network.

!!! warning "Security"
    Do not expose port 9090 to the public internet without additional security measures.

## Response Format

All responses are JSON:

```json
{
  "data": { ... },
  "error": null
}
```

Error responses:
```json
{
  "error": "Error message here"
}
```

## Swagger Documentation

Interactive API docs are available at:

```
http://localhost:9090/api/v1/mining/swagger/index.html
```

## Common Headers

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `Accept` | `application/json` |

## HTTP Methods

| Method | Usage |
|--------|-------|
| `GET` | Retrieve resources |
| `POST` | Create resources or trigger actions |
| `PUT` | Update resources |
| `DELETE` | Remove resources |

## Error Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `400` | Bad request (invalid input) |
| `404` | Resource not found |
| `500` | Internal server error |

## Rate Limiting

No rate limiting is currently implemented.

## Example: List Running Miners

```bash
curl http://localhost:9090/api/v1/mining/miners
```

Response:
```json
[
  {
    "name": "xmrig-123",
    "running": true,
    "full_stats": {
      "hashrate": {
        "total": [1234.5, 1230.2, 1228.8]
      },
      "results": {
        "shares_good": 42,
        "shares_total": 43
      }
    }
  }
]
```

## Example: Start a Miner

```bash
curl -X POST http://localhost:9090/api/v1/mining/profiles/abc123/start
```

Response:
```json
{
  "name": "xmrig-456",
  "message": "Miner started successfully"
}
```

## Example: Stop a Miner

```bash
curl -X DELETE http://localhost:9090/api/v1/mining/miners/xmrig-456
```

Response:
```json
{
  "message": "Miner stopped"
}
```

## WebSocket (Future)

A WebSocket endpoint for real-time updates is planned:
```
ws://localhost:9090/api/v1/mining/ws
```
