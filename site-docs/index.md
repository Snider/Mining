# Mining Dashboard

<div style="text-align: center; margin: 2rem 0;">
<img src="assets/screenshots/dashboard.png" alt="Mining Dashboard" style="max-width: 100%; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);">
</div>

**Mining Dashboard** is a powerful, open-source multi-miner management system that lets you control XMRig, TT-Miner, and other mining software from a single, beautiful web interface.

!!! tip "Built with Claude Code"
    This entire project—backend, frontend, documentation, and tests—was developed using [Claude Code](https://claude.ai/code), Anthropic's AI-powered development assistant. See [how Claude helped build this](about/claude.md).

## Key Features

<div class="grid cards" markdown>

-   :material-lightning-bolt:{ .lg .middle } **Real-time Monitoring**

    ---

    Live hashrate graphs, share statistics, and performance metrics updated every 5 seconds

-   :material-cog:{ .lg .middle } **Multi-Miner Support**

    ---

    Control XMRig (CPU) and TT-Miner (GPU) from a unified interface. Easy to extend for additional miners

-   :material-console:{ .lg .middle } **Console Access**

    ---

    Full console output with ANSI color support and stdin input for miner commands

-   :material-server-network:{ .lg .middle } **P2P Multi-Node**

    ---

    Control remote mining rigs via encrypted WebSocket connections without cloud dependencies

-   :material-database:{ .lg .middle } **Historical Data**

    ---

    SQLite-backed hashrate history with configurable retention (5m to 24h views)

-   :material-api:{ .lg .middle } **REST API**

    ---

    Full REST API with Swagger documentation for automation and integration

</div>

## Quick Start

```bash
# Clone the repository
git clone https://github.com/Snider/Mining.git
cd Mining

# Build the CLI
make build

# Start the server
./miner-ctrl serve
```

Then open [http://localhost:9090](http://localhost:9090) in your browser!

## Screenshots

<div class="grid" markdown>

![Dashboard](assets/screenshots/dashboard.png){ loading=lazy }

![Profiles](assets/screenshots/profiles.png){ loading=lazy }

![Console](assets/screenshots/console.png){ loading=lazy }

![Nodes](assets/screenshots/nodes.png){ loading=lazy }

</div>

## Architecture

The Mining Dashboard consists of:

- **Go Backend** - REST API server with miner process management
- **Angular Frontend** - Modern, responsive web interface
- **SQLite Database** - Persistent hashrate history storage
- **P2P Network** - Encrypted node-to-node communication

```mermaid
graph TB
    subgraph "Web Browser"
        UI[Angular Dashboard]
    end

    subgraph "Mining Server"
        API[REST API :9090]
        MGR[Miner Manager]
        DB[(SQLite DB)]
        P2P[P2P Transport :9091]
    end

    subgraph "Mining Processes"
        XMR[XMRig]
        TTM[TT-Miner]
    end

    subgraph "Remote Nodes"
        RN1[Worker Node 1]
        RN2[Worker Node 2]
    end

    UI --> API
    API --> MGR
    MGR --> XMR
    MGR --> TTM
    MGR --> DB
    P2P --> RN1
    P2P --> RN2
```

## Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/Snider/Mining/issues)
- **Documentation**: You're reading it!

## License

This project is open source. See the repository for license details.
