# Mining Pool Integration Guide

This guide provides comprehensive information about mining pool selection, configuration, and integration with the Mining Platform.

## Overview

Mining pools allow miners to combine their computational power and share rewards. Choosing the right pool is crucial for optimizing your mining profitability and experience.

## Recommended Pools by Coin

### Monero (XMR)

| Pool | URL | Port | Fee | Min Payout | Notes |
|------|-----|------|-----|-----------|-------|
| **SupportXMR** | pool.supportxmr.com | 3333 | 0.6% | 0.003 XMR | Best for beginners, no registration |
| **P2Pool** | p2pool.io | 3333 | 0% | 0.0 XMR | Decentralized, instant payouts |
| **Nanopool** | xmr-eu1.nanopool.org | 14433 | 1.0% | 0.003 XMR | Global network, mobile app |
| **MoneroOcean** | gulf.moneroocean.stream | 10128 | 1.0% | 0.003 XMR | Multi-algo, auto-switching |
| **WoolyPooly** | xmr.woolypooly.com | 3333 | 0.5% | 0.003 XMR | Low fees, merged mining |

### Ethereum Classic (ETC)

| Pool | URL | Port | Fee | Min Payout | Notes |
|------|-----|------|-----|-----------|-------|
| **WoolyPooly** | etc.woolypooly.com | 3333 | 0.5% | 0.01 ETC | Reliable, low fees |
| **Nanopool** | etc-eu1.nanopool.org | 19999 | 1.0% | 0.01 ETC | Established, global |
| **2Miners** | etc.2miners.com | 1010 | 1.0% | 0.01 ETC | PPLNS, no registration |
| **Ethermine** | etc.ethermine.org | 4444 | 1.0% | 0.01 ETC | High performance |

### Ravencoin (RVN)

| Pool | URL | Port | Fee | Min Payout | Notes |
|------|-----|------|-----|-----------|-------|
| **WoolyPooly** | rvn.woolypooly.com | 3333 | 0.5% | 5 RVN | Best overall |
| **Flypool** | rvn.flypool.org | 3333 | 1.0% | 5 RVN | High uptime |
| **2Miners** | rvn.2miners.com | 6060 | 1.0% | 5 RVN | PPLNS rewards |
| **Ravenminer** | ravenminer.com | 3333 | 0.5% | 5 RVN | Community pool |

### Zano (ZANO)

| Pool | URL | Port | Fee | Min Payout | Notes |
|------|-----|------|-----|-----------|-------|
| **WoolyPooly** | zano.woolypooly.com | 3333 | 1.0% | 0.5 ZANO | Primary pool |
| **ZanoPool** | pool.zano.org | 11555 | 1.0% | 0.5 ZANO | Official pool |

## Pool Configuration

### Connection String Format

Most pools use this standard format:

```
protocol://hostname:port
```

**Examples:**
```
stratum+tcp://pool.supportxmr.com:3333
stratum+ssl://pool.supportxmr.com:3334
```

### Authentication Format

Standard authentication uses:

```
Username: WALLET_ADDRESS.WORKER_NAME
Password: x (or empty)
```

**Example:**
```json
{
  "pool": "stratum+tcp://pool.supportxmr.com:3333",
  "wallet": "4ABC1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890AB",
  "username": "4ABC1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890AB.miner1",
  "password": "x"
}
```

## Port Configuration

### Standard Port Mapping

Most pools follow this convention:

```
3333 = Standard (auto difficulty)     → Try this first
4444 = Medium difficulty
5555 = High difficulty                 → Use for powerful miners
6666 = Very high difficulty

For TLS/SSL, add 1 to the port:
3334 = Standard over TLS
4445 = Medium over TLS
5556 = High over TLS
```

### Difficulty Selection

Choose a port based on your hashrate:

**Monero (RandomX):**
- 3333 (auto): Any hashrate
- 4444: > 5 KH/s
- 5555: > 20 KH/s

**Ethereum Classic (ETChash):**
- 3333 (auto): Any hashrate
- 4444: > 50 MH/s
- 5555: > 200 MH/s

## Regional Servers

For best performance, choose a server close to your location:

### Nanopool Regions

```
Europe:      xmr-eu1.nanopool.org
             etc-eu1.nanopool.org

US East:     xmr-us-east1.nanopool.org
             etc-us-east1.nanopool.org

US West:     xmr-us-west1.nanopool.org
             etc-us-west1.nanopool.org

Asia:        xmr-asia1.nanopool.org
             etc-asia1.nanopool.org
```

### MoneroOcean Regions

```
US:          gulf.moneroocean.stream
Europe:      eu.moneroocean.stream
Asia:        asia.moneroocean.stream
```

## Pool Selection Criteria

### For Beginners

Choose pools with:
- Low minimum payout
- No registration required
- Good documentation
- Active community support
- Stable uptime

**Recommended:**
1. SupportXMR (XMR)
2. WoolyPooly (ETC, RVN)
3. Nanopool (all coins)

### For Advanced Users

Consider pools with:
- Lower fees (0.5% or less)
- Advanced features
- API access
- Custom configurations

**Recommended:**
1. P2Pool (XMR) - Decentralized
2. WoolyPooly (all coins) - Low fees
3. SupportXMR (XMR) - Open source

### For Privacy-Focused Mining

Prioritize:
- No registration required
- Decentralized pools
- No personal information collection

**Recommended:**
1. P2Pool (XMR) - Fully decentralized
2. SupportXMR (XMR) - No KYC
3. Any pool without registration

## Configuration Examples

### Monero on SupportXMR

```json
{
  "pool": "stratum+tcp://pool.supportxmr.com:3333",
  "wallet": "YOUR_XMR_WALLET_ADDRESS",
  "algo": "rx/0",
  "threads": 4,
  "cpuPriority": 3
}
```

### Ethereum Classic on WoolyPooly

```json
{
  "pool": "stratum+tcp://etc.woolypooly.com:3333",
  "wallet": "YOUR_ETC_WALLET_ADDRESS",
  "algo": "etchash",
  "cuda": {
    "enabled": true,
    "devices": [0, 1]
  }
}
```

### Ravencoin on Flypool

```json
{
  "pool": "stratum+tcp://rvn.flypool.org:3333",
  "wallet": "YOUR_RVN_WALLET_ADDRESS",
  "algo": "kawpow",
  "opencl": {
    "enabled": true,
    "devices": [0]
  }
}
```

### Dual Mining (CPU + GPU)

Mine Monero on CPU and Ethereum Classic on GPU:

```json
{
  "pools": [
    {
      "pool": "stratum+tcp://pool.supportxmr.com:3333",
      "wallet": "YOUR_XMR_WALLET",
      "algo": "rx/0",
      "threads": 4
    },
    {
      "pool": "stratum+tcp://etc.woolypooly.com:3333",
      "wallet": "YOUR_ETC_WALLET",
      "algo": "etchash",
      "cuda": {
        "enabled": true,
        "devices": [0]
      }
    }
  ]
}
```

## Fee Comparison

### Monero Pools

| Pool | Fee | Min Payout | Est. Monthly Earnings (1 KH/s) |
|------|-----|-----------|-------------------------------|
| P2Pool | 0% | 0.0 XMR | 100% |
| WoolyPooly | 0.5% | 0.003 XMR | 99.5% |
| SupportXMR | 0.6% | 0.003 XMR | 99.4% |
| Nanopool | 1.0% | 0.003 XMR | 99.0% |
| MoneroOcean | 1.0% | 0.003 XMR | 99.0% |

**Impact:** Fee difference of 0.5% = ~$0.50/month at $100/month earnings

### Ethereum Classic Pools

| Pool | Fee | Min Payout | Est. Monthly Earnings (100 MH/s) |
|------|-----|-----------|----------------------------------|
| WoolyPooly | 0.5% | 0.01 ETC | 99.5% |
| 2Miners | 1.0% | 0.01 ETC | 99.0% |
| Nanopool | 1.0% | 0.01 ETC | 99.0% |
| Ethermine | 1.0% | 0.01 ETC | 99.0% |

## Wallet Address Validation

### Monero (XMR)

Valid XMR addresses:
- **Length:** 95 characters
- **Prefix:** 4 (mainnet) or 8 (testnet)
- **Format:** Base58

Example:
```
4ABC1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890AB1234567890ABCDEF1234567890ABC
```

### Ethereum Classic (ETC)

Valid ETC addresses:
- **Length:** 42 characters (including 0x)
- **Prefix:** 0x
- **Format:** Hexadecimal

Example:
```
0x1234567890123456789012345678901234567890
```

### Ravencoin (RVN)

Valid RVN addresses:
- **Length:** 26-35 characters
- **Prefix:** R
- **Format:** Base58

Example:
```
RAbC123456789aBcDeF123456789XyZ
```

## Testing Pool Connectivity

### Using telnet

```bash
telnet pool.supportxmr.com 3333
```

If successful, you'll see a connection established message.

### Using nc (netcat)

```bash
nc -zv pool.supportxmr.com 3333
```

### Using the Mining Platform

```bash
# Via API
curl -X POST http://localhost:8080/api/v1/mining/test-pool \
  -H "Content-Type: application/json" \
  -d '{
    "pool": "stratum+tcp://pool.supportxmr.com:3333"
  }'
```

## Troubleshooting

### Connection Refused

**Possible causes:**
- Pool is down
- Port is blocked by firewall
- Incorrect hostname

**Solutions:**
1. Try TLS port (add 1 to port number)
2. Check pool website for status
3. Try alternative pool
4. Check firewall settings

### High Rejected Shares

**Possible causes:**
- Network latency
- Incorrect algorithm
- Outdated miner software

**Solutions:**
1. Switch to closer regional server
2. Verify algorithm matches pool
3. Update miner software
4. Try different difficulty port

### Very Low Hashrate

**Possible causes:**
- Incorrect thread count
- CPU throttling
- System resource constraints

**Solutions:**
1. Adjust thread count (try half of CPU cores)
2. Check CPU temperature
3. Close other applications
4. Increase CPU priority

### No Payouts

**Possible causes:**
- Minimum payout not reached
- Incorrect wallet address
- Pool payment schedule

**Solutions:**
1. Check pool dashboard for balance
2. Verify wallet address is correct
3. Review pool's payout policy
4. Contact pool support

## Advanced Features

### Pool Failover

Configure backup pools:

```json
{
  "pools": [
    {
      "pool": "stratum+tcp://pool.supportxmr.com:3333",
      "wallet": "YOUR_WALLET",
      "algo": "rx/0"
    },
    {
      "pool": "stratum+tcp://xmr-eu1.nanopool.org:14433",
      "wallet": "YOUR_WALLET",
      "algo": "rx/0",
      "failover": true
    }
  ]
}
```

### TLS/SSL Encryption

Use encrypted connection:

```json
{
  "pool": "stratum+ssl://pool.supportxmr.com:3334",
  "wallet": "YOUR_WALLET",
  "algo": "rx/0",
  "tls": {
    "enabled": true,
    "fingerprint": "optional_pool_certificate_fingerprint"
  }
}
```

### Nicehash Support

For Nicehash-compatible pools:

```json
{
  "pool": "stratum+tcp://randomxmonero.auto.nicehash.com:9200",
  "wallet": "YOUR_NICEHASH_BTC_ADDRESS",
  "algo": "rx/0",
  "nicehash": true
}
```

## Pool Database

The Mining Platform includes a comprehensive pool database at:

```
/docs/xmr-pools-database.json
```

This database contains:
- 10+ major mining pools
- 60+ port configurations
- Regional server variants
- Fee structures
- Minimum payouts
- Reliability scores

Load in your application:

```typescript
import poolDatabase from './xmr-pools-database.json';

const supportxmr = poolDatabase.pools.find(p => p.id === 'supportxmr');
console.log(`${supportxmr.name} - ${supportxmr.fee_percent}% fee`);
```

## Best Practices

1. **Start with recommended pools**: Use established pools with good reputation
2. **Monitor performance**: Track hashrate and accepted shares
3. **Use regional servers**: Choose servers close to your location
4. **Enable TLS when possible**: For enhanced security
5. **Configure failover**: Have backup pools configured
6. **Check pool stats regularly**: Monitor your balance and payouts
7. **Join pool community**: Discord, Telegram, or forums
8. **Read pool documentation**: Understand specific pool features
9. **Test before committing**: Mine for a day before large deployments
10. **Update regularly**: Keep miner software up to date

## Resources

- [Pool Research Documentation](../00-START-HERE.md)
- [Pool Integration Guide](../pool-integration-guide.md)
- [Quick Reference](../QUICK-REFERENCE.md)
- [XMR Pool Database](../xmr-pools-database.json)

## Next Steps

- Try the [Quick Start Guide](../getting-started/quick-start.md) to begin mining
- Read about [Algorithms](algorithms.md) supported by the platform
- Explore the [API Documentation](../api/endpoints.md) for automation
