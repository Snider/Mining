# Multi-Miner Support

The Mining Dashboard supports running multiple miners simultaneously with unified monitoring.

## Supported Miners

### XMRig (CPU Mining)

| Feature | Details |
|---------|---------|
| **Type** | CPU miner |
| **Algorithms** | RandomX, CryptoNight, AstroBWT |
| **Coins** | Monero (XMR), Loki, Arweave, etc. |
| **API** | HTTP JSON API |
| **Download** | Automatic from GitHub releases |

### TT-Miner (GPU Mining)

| Feature | Details |
|---------|---------|
| **Type** | NVIDIA GPU miner |
| **Algorithms** | Ethash, KawPow, ProgPow, ZelHash |
| **Coins** | Various PoW coins |
| **API** | HTTP JSON API |
| **Requirements** | NVIDIA CUDA GPU |

## Running Multiple Instances

You can run multiple miners of the same or different types:

```
┌─────────────────────────────────────────┐
│           Mining Dashboard              │
├─────────────────────────────────────────┤
│  xmrig-456    │  12.5 kH/s  │ Running  │
│  xmrig-789    │  11.8 kH/s  │ Running  │
│  tt-miner-123 │  45.2 MH/s  │ Running  │
└─────────────────────────────────────────┘
```

### Unique Instance Names

Each miner gets a unique name based on:
- Miner type (xmrig, tt-miner)
- Unique suffix (timestamp or random)

Example: `xmrig-1704067200`

### Separate Configurations

Each instance can have:
- Different pool
- Different wallet
- Different resource allocation

## Workers Page

The Workers page shows all running miners:

![Workers](../assets/screenshots/workers.png)

### Starting Multiple Miners

1. Go to **Workers** page
2. Select a profile from the dropdown
3. Click **Start**
4. Repeat with different profiles

### Managing Workers

Each worker card shows:
- **Name** - Instance name
- **Hashrate** - Current mining speed
- **Status** - Running/Stopped
- **Stop** button - Terminate the miner

## Resource Allocation

### CPU Miners (XMRig)

Control CPU usage via profile settings:

```json
{
  "threads": 4,  // Use 4 threads
  "hugePages": true
}
```

Run multiple instances with different thread counts:
- Instance 1: 4 threads
- Instance 2: 2 threads

### GPU Miners (TT-Miner)

Control GPU selection via profile settings:

```json
{
  "devices": "0,1"  // Use GPU 0 and 1
}
```

Run multiple instances on different GPUs:
- Instance 1: GPU 0
- Instance 2: GPU 1,2

## Aggregated Statistics

The dashboard shows combined stats:

| Stat | Calculation |
|------|-------------|
| **Total Hashrate** | Sum of all miners |
| **Total Shares** | Sum of all accepted shares |
| **Total Rejected** | Sum of all rejected shares |
| **Efficiency** | (Total - Rejected) / Total |

## API Access

### List All Miners
```bash
curl http://localhost:9090/api/v1/mining/miners
```

### Start a New Instance
```bash
curl -X POST http://localhost:9090/api/v1/mining/profiles/{id}/start
```

### Stop a Specific Instance
```bash
curl -X DELETE http://localhost:9090/api/v1/mining/miners/xmrig-456
```

## Best Practices

1. **Don't oversubscribe resources** - Leave some CPU/GPU headroom
2. **Use different pools** - Spread risk across multiple pools
3. **Monitor temperatures** - Watch for thermal throttling
4. **Name profiles clearly** - "GPU0-ETH", "CPU-XMR", etc.
