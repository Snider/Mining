# Changelog

All notable changes to this project.

## [Unreleased]

### Added
- MkDocs documentation site with Material theme
- Screenshots of all UI pages
- Comprehensive API and CLI documentation

---

## [0.3.0] - 2024

### Added
- **Multi-Node P2P System**
  - Node identity with X25519 keypairs
  - Peer registry and management
  - WebSocket transport layer
  - Remote miner control (start/stop/stats/logs)
  - CLI commands: `node`, `peer`, `remote`
  - REST API endpoints for P2P operations

- **SQLite Persistence**
  - Hashrate history storage (30-day retention)
  - Historical data API endpoints
  - Time-range queries

- **Dashboard Enhancements**
  - Stats bar with all key metrics
  - Time range selector for charts
  - Worker dropdown for multi-miner support
  - Avg difficulty per share display

- **Console Improvements**
  - Stdin command support (h, p, r, s, c)
  - ANSI color rendering
  - Auto-scroll toggle

### Changed
- Refactored miner interface for better extensibility
- Improved stats collection with background goroutines
- Enhanced error handling throughout

---

## [0.2.0] - 2024

### Added
- **TT-Miner Support**
  - GPU mining with NVIDIA CUDA
  - Automatic installation from GitHub
  - Stats parsing from stdout

- **Profile Management**
  - Create, edit, delete mining profiles
  - Start miners from saved profiles
  - Autostart configuration

- **Console View**
  - Live miner output streaming
  - Base64-encoded log retrieval
  - Clear and auto-scroll controls

### Changed
- Improved XMRig stats collection
- Better process lifecycle management
- Enhanced UI responsiveness

---

## [0.1.0] - 2024

### Added
- **Initial Release**
- XMRig miner support
  - Automatic installation
  - Config generation
  - Stats via HTTP API
- REST API with Gin framework
- Angular dashboard
  - Real-time hashrate display
  - Miner control (start/stop)
  - System information
- CLI with Cobra
  - `serve` command
  - `start`, `stop`, `status`
  - `install`, `uninstall`
  - `doctor` health check
- Swagger API documentation

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 0.3.0 | 2024 | P2P multi-node, SQLite, enhanced dashboard |
| 0.2.0 | 2024 | TT-Miner, profiles, console |
| 0.1.0 | 2024 | Initial release with XMRig |

## Roadmap

Future planned features:

- [ ] Automatic peer discovery
- [ ] Fleet-wide statistics aggregation
- [ ] Mobile-responsive improvements
- [ ] Notification system (email, webhooks)
- [ ] Mining pool profit comparison
- [ ] Additional miner support (GMiner, lolMiner)
