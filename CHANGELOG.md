# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Structured logging package with configurable log levels
- Rate limiter with automatic cleanup for API protection
- X-Request-ID middleware for request tracing
- Structured API error responses with error codes and suggestions
- MinerFactory pattern for centralized miner instantiation
- StatsCollector pattern for parallel stats collection
- Context propagation throughout the codebase
- WebSocket event system for real-time updates
- Simulation mode for UI development and testing
- Mermaid architecture diagrams in documentation

### Changed
- Optimized `collectMinerStats()` for parallel execution
- Replaced `log.Printf` with structured logging throughout
- Improved hashrate history with two-tier storage (high-res and low-res)
- Enhanced shutdown handling with proper cleanup

### Fixed
- Race conditions in concurrent database access
- Memory leaks in hashrate history retention
- Context cancellation propagation to database operations

## [0.0.9] - 2025-12-11

### Added
- Enhanced dashboard layout with responsive stats bar
- Setup wizard for first-time configuration
- Admin panel for miner management
- Profile management with multiple miner support
- Live hashrate visualization with Highcharts
- Comprehensive docstrings throughout the mining package
- CI/CD matrix testing and conditional releases

### Changed
- Refactored profile selection to support multiple miners
- Improved UI layout and accessibility
- Enhanced mining configuration management

### Fixed
- UI build and server configuration issues

## [0.0.8] - 2025-11-09

### Added
- Web dashboard (`mbe-mining-dashboard.js`) included in release binaries
- Interactive web interface for miner-cli

## [0.0.7] - 2025-11-09

### Fixed
- Windows build compatibility

## [0.0.6] - 2025-11-09

### Added
- Initial public release
- XMRig miner support
- TT-Miner (GPU) support
- RESTful API with Swagger documentation
- CLI with interactive shell
- Miner autostart configuration
- Hashrate history tracking

[Unreleased]: https://github.com/Snider/Mining/compare/v0.0.9...HEAD
[0.0.9]: https://github.com/Snider/Mining/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/Snider/Mining/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/Snider/Mining/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/Snider/Mining/releases/tag/v0.0.6
