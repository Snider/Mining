# Dashboard

The Dashboard is the main monitoring view for your mining operations.

![Dashboard](../assets/screenshots/dashboard.png)

## Stats Bar

The top stats bar shows aggregate statistics across all running miners:

| Stat | Description |
|------|-------------|
| **Hashrate** | Combined hashrate from all miners |
| **Shares** | Total accepted shares / rejected |
| **Uptime** | Longest running miner uptime |
| **Pool** | Connected pool(s) |
| **Avg Diff** | Average difficulty per accepted share |
| **Workers** | Number of active mining processes |

## Hashrate Chart

The main chart displays hashrate over time with configurable time ranges:

- **5m** - Last 5 minutes (high resolution)
- **15m** - Last 15 minutes
- **30m** - Last 30 minutes
- **1h** - Last hour
- **3h** - Last 3 hours
- **6h** - Last 6 hours
- **12h** - Last 12 hours
- **24h** - Last 24 hours

### Chart Features

- **Multi-miner support** - Each miner shows as a separate line
- **Color coding** - Consistent colors per miner
- **Hover tooltips** - Exact values on hover
- **Auto-refresh** - Updates every 5 seconds

## Quick Stats Cards

Below the chart, four cards show key metrics:

| Card | Description |
|------|-------------|
| **Peak Hashrate** | Highest hashrate achieved |
| **Efficiency** | Share acceptance rate (%) |
| **Avg. Share Time** | Average time between shares |
| **Difficulty** | Current pool difficulty |

## Worker Selector

The dropdown in the top-right allows filtering:

- **All Workers** - Show combined stats from all miners
- **Individual miner** - Focus on a single miner's stats

When a single miner is selected:
- Stats show only that miner's data
- Chart shows only that miner's history
- Console defaults to that miner

## Data Sources

### Live Data
- Polled every **5 seconds** from miner APIs
- Stored in memory with 5-minute high-resolution window

### Historical Data
- Stored in SQLite database
- 1-minute resolution after initial 5-minute window
- Configurable retention (default 30 days)

## API Endpoints

The dashboard data comes from these endpoints:

```
GET /api/v1/mining/miners
GET /api/v1/mining/miners/{name}/stats
GET /api/v1/mining/history/miners/{name}/hashrate
```

See [API Reference](../api/endpoints.md) for details.
