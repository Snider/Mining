# Mining API Documentation

The Mining project provides a comprehensive RESTful API for managing cryptocurrency miners. This API is served by the `miner-ctrl serve` command.

## Swagger Documentation

The project includes automatically generated Swagger (OpenAPI) documentation.

When you run the service (e.g., `miner-ctrl serve`), the Swagger UI is available at:

```
http://<host>:<port>/api/v1/mining/swagger/index.html
```

(Default: `http://localhost:8080/api/v1/mining/swagger/index.html`)

You can also find the raw Swagger files in this directory:
- [swagger.json](swagger.json)
- [swagger.yaml](swagger.yaml)

## API Endpoints Summary

All endpoints are prefixed with `/api/v1/mining` (or the configured `--namespace`).

### System & Health

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/info` | Retrieves cached installation details for all miners and system info. |
| `POST` | `/doctor` | Performs a live check on all available miners to verify installation status. |
| `POST` | `/update` | Checks if any installed miners have a new version available. |

### Miner Management

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/miners` | List all currently running miners. |
| `GET` | `/miners/available` | List all miner types supported by the system. |
| `POST` | `/miners/:miner_type` | Start a new miner instance. Requires a JSON config body. |
| `DELETE` | `/miners/:miner_name` | Stop a running miner instance. |
| `POST` | `/miners/:miner_type/install` | Install or update a specific miner binary. |
| `DELETE` | `/miners/:miner_type/uninstall` | Uninstall a specific miner and remove its files. |
| `GET` | `/miners/:miner_name/stats` | Get real-time statistics (hashrate, shares, etc.) for a running miner. |
| `GET` | `/miners/:miner_name/hashrate-history` | Get historical hashrate data. |

## Data Models

### SystemInfo
Contains OS, Architecture, Go Version, total RAM, and a list of installed miners.

### Miner Configuration (Config)
A comprehensive object containing settings for the miner, such as:
- `pool`: Mining pool address.
- `wallet`: Wallet address.
- `cpuPriority`: CPU priority settings.
- `threads`: Number of threads to use.
- `algo`: Algorithm to mine.

### PerformanceMetrics
Real-time stats from the miner:
- `hashrate`: Current hashrate in H/s.
- `shares`: Total shares submitted.
- `rejected`: Rejected shares.
- `uptime`: Time since start in seconds.
