# Future Ideas

This document captures ideas for future enhancements identified during code review and architecture analysis.

## Remote Monitoring Bot

**Priority:** High
**Effort:** Medium

Create a Telegram or Discord bot for remote monitoring of mining operations.

### Features
- Real-time hashrate alerts (drop below threshold)
- Share accepted/rejected notifications
- Daily summary reports
- Remote start/stop commands
- Multi-node aggregated stats

### Implementation Notes
- Use existing EventHub WebSocket infrastructure
- Bot subscribes to miner events and forwards to chat
- Store bot token in settings (encrypted)
- Rate limit notifications to prevent spam

---

## Pool Auto-Discovery

**Priority:** Medium
**Effort:** Low

Add pool auto-discovery with a community-maintained `pools.json` file.

### Features
- Curated list of pools per algorithm/coin
- Pool health/latency checking
- Automatic failover suggestions
- Community contributions via PR

### Implementation Notes
- Host `pools.json` on GitHub (or embed in binary)
- Include: name, url, ports, fees, minimum payout, regions
- UI dropdown to select from known pools
- Validate pool connectivity before saving

### Example Structure
```json
{
  "monero": [
    {
      "name": "SupportXMR",
      "url": "pool.supportxmr.com",
      "ports": {"stratum": 3333, "ssl": 443},
      "fee": 0.6,
      "minPayout": 0.1
    }
  ]
}
```

---

## Profitability Calculator

**Priority:** Medium
**Effort:** Medium

Add real-time profitability calculations using CoinGecko API.

### Features
- Fetch current coin prices (XMR, ETH, RVN, etc.)
- Calculate daily/weekly/monthly earnings based on hashrate
- Factor in electricity costs (user-configurable)
- Compare profitability across algorithms
- Historical profitability charts

### Implementation Notes
- CoinGecko free tier: 10-50 calls/minute
- Cache prices for 5 minutes to reduce API calls
- Store electricity rate in settings ($/kWh)
- Formula: `(hashrate / network_hashrate) * block_reward * price - electricity_cost`

### API Endpoints
- `GET /api/v1/mining/profitability` - Current estimates
- `GET /api/v1/mining/profitability/history` - Historical data

---

## One-Click Deploy Templates

**Priority:** Low
**Effort:** Medium

Create deployment templates for popular self-hosting platforms.

### Platforms
- **Unraid** - Community Applications template
- **Proxmox** - LXC/VM template with cloud-init
- **DigitalOcean** - 1-Click Droplet image
- **Docker Compose** - Production-ready compose file
- **Kubernetes** - Helm chart

### Template Contents
- Pre-configured environment variables
- Volume mounts for persistence
- Health checks
- Resource limits
- Auto-update configuration

### Files to Create
```
deploy/
├── docker-compose.prod.yml
├── unraid/
│   └── mining-dashboard.xml
├── proxmox/
│   └── mining-dashboard.yaml
├── kubernetes/
│   └── helm/
└── digitalocean/
    └── marketplace.yaml
```

---

## Community Visibility (Manual Tasks)

### Submit to Awesome Lists
- [ ] [awesome-monero](https://github.com/monero-ecosystem/awesome-monero)
- [ ] [awesome-selfhosted](https://github.com/awesome-selfhosted/awesome-selfhosted)
- [ ] [awesome-crypto](https://github.com/coinpride/CryptoList)

### GitHub Repository Optimization
- [ ] Add topic tags: `mining`, `monero`, `xmrig`, `cryptocurrency`, `dashboard`, `self-hosted`, `golang`, `angular`
- [ ] Add social preview image
- [ ] Create demo GIF for README showcasing the dashboard UI
- [ ] Create GitHub Discussions for community Q&A
- [ ] Add "Used By" section in README

---

## Advanced API Authentication

**Priority:** Medium
**Effort:** Medium

Expand beyond basic/digest auth with more robust authentication options.

### Current Implementation
- HTTP Basic and Digest authentication (implemented)
- Enabled via environment variables: `MINING_API_AUTH`, `MINING_API_USER`, `MINING_API_PASS`

### Future Options

#### JWT Tokens
- Stateless authentication with expiring tokens
- Refresh token support
- Scoped permissions (read-only, admin, etc.)

#### API Keys
- Generate/revoke API keys from dashboard
- Per-key permissions and rate limits
- Key rotation support

#### OAuth2/OIDC Integration
- Support external identity providers (Google, GitHub, Keycloak)
- SSO for enterprise deployments
- Useful for multi-user mining farms

#### mTLS (Mutual TLS)
- Certificate-based client authentication
- Strongest security for production deployments
- No passwords to manage

### Implementation Notes
- Store credentials/keys in encrypted config file
- Add `/api/v1/auth/token` endpoint for JWT issuance
- Consider using `golang-jwt/jwt` for JWT implementation
- Add audit logging for authentication events

---

## Additional Ideas

### GPU Temperature Monitoring
- Read GPU temps via NVML (NVIDIA) or ROCm (AMD)
- Alert on thermal throttling
- Auto-pause mining on overtemp

### Mining Schedule
- Time-based mining schedules
- Pause during peak electricity hours
- Resume when rates are lower

### Multi-Algorithm Auto-Switching
- Monitor profitability across algorithms
- Automatically switch to most profitable
- Configurable switch threshold (prevent thrashing)

### Web Terminal
- Embedded terminal in dashboard
- Direct access to miner stdin/stdout
- Real-time log streaming with search/filter
