# Historical Data

The Mining Dashboard stores hashrate history in an SQLite database for charting and analysis.

## Database Location

```
~/.local/share/lethean-desktop/mining.db
```

## Data Retention

| Time Range | Resolution | Retention |
|------------|------------|-----------|
| 0-5 minutes | 10 seconds | In-memory only |
| 5 min - 24 hours | 1 minute | SQLite |
| 24 hours+ | 1 minute | Configurable (default 30 days) |

## Configuring Retention

In `~/.config/lethean-desktop/miners.json`:

```json
{
  "database": {
    "enabled": true,
    "retentionDays": 30
  }
}
```

### Disable Database

```json
{
  "database": {
    "enabled": false
  }
}
```

When disabled, only in-memory data is available (last 5 minutes).

## Time Range Selection

The dashboard supports viewing different time windows:

| Label | Minutes | Use Case |
|-------|---------|----------|
| 5m | 5 | Real-time monitoring |
| 15m | 15 | Short-term trends |
| 30m | 30 | Recent performance |
| 45m | 45 | Extended recent |
| 1h | 60 | Last hour |
| 3h | 180 | Morning/afternoon |
| 6h | 360 | Half day |
| 12h | 720 | Full shift |
| 24h | 1440 | Full day |

## Data Schema

```sql
CREATE TABLE hashrate_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    miner_name TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    hashrate INTEGER NOT NULL
);

CREATE INDEX idx_miner_time ON hashrate_history(miner_name, timestamp);
```

## API Endpoints

### Get Historical Hashrate

```bash
curl "http://localhost:9090/api/v1/mining/history/miners/xmrig-123/hashrate?since=2024-01-15T00:00:00Z&until=2024-01-15T23:59:59Z"
```

Response:
```json
[
  {"timestamp": "2024-01-15T10:30:00Z", "hashrate": 1234},
  {"timestamp": "2024-01-15T10:31:00Z", "hashrate": 1256},
  {"timestamp": "2024-01-15T10:32:00Z", "hashrate": 1248}
]
```

### Get All Miners History

```bash
curl "http://localhost:9090/api/v1/mining/history/miners?since=2024-01-15T00:00:00Z"
```

### Get Miner Stats Summary

```bash
curl "http://localhost:9090/api/v1/mining/history/miners/xmrig-123?since=2024-01-15T00:00:00Z"
```

Response includes:
- Average hashrate
- Peak hashrate
- Total shares
- Time period

## Data Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Miner     │────▶│   Manager   │────▶│   SQLite    │
│   (XMRig)   │     │  (polling)  │     │    DB       │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       │   Stats every     │   Store every     │
       │   10 seconds      │   1 minute        │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  In-Memory  │◀────│   Service   │◀────│  Frontend   │
│   Buffer    │     │   (API)     │     │   (Charts)  │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Automatic Cleanup

The database automatically purges old data:

1. Runs on startup
2. Runs every 24 hours
3. Deletes records older than `retentionDays`

## Manual Database Access

```bash
sqlite3 ~/.local/share/lethean-desktop/mining.db

# View recent records
SELECT * FROM hashrate_history
ORDER BY timestamp DESC
LIMIT 10;

# Get average hashrate for a miner
SELECT AVG(hashrate) FROM hashrate_history
WHERE miner_name = 'xmrig-123'
AND timestamp > datetime('now', '-1 hour');

# Database size
SELECT page_count * page_size as size
FROM pragma_page_count(), pragma_page_size();
```

## Exporting Data

```bash
sqlite3 ~/.local/share/lethean-desktop/mining.db \
  ".mode csv" \
  ".headers on" \
  "SELECT * FROM hashrate_history WHERE timestamp > datetime('now', '-24 hours');" \
  > hashrate_export.csv
```
