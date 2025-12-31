# Comprehensive Code Review: 109 Findings

> **Generated:** December 31, 2025
> **Reviewed by:** 8 Opus 4.5 Domain-Specialized Agents
> **Commit:** d533164 (post-hardening baseline)

This document captures all 109 findings from a comprehensive 8-domain code review. Each finding includes severity, file locations, and actionable remediation steps.

---

## Summary Table

| Domain | Findings | Critical | High | Medium | Low |
|--------|----------|----------|------|--------|-----|
| Security | 8 | 0 | 0 | 4 | 4 |
| Concurrency | 9 | 0 | 1 | 5 | 3 |
| Performance | 12 | 0 | 2 | 6 | 4 |
| Resilience | 17 | 0 | 3 | 8 | 6 |
| Testing | 12 | 3 | 5 | 3 | 1 |
| API Design | 16 | 0 | 2 | 8 | 6 |
| Architecture | 14 | 0 | 2 | 7 | 5 |
| P2P Network | 21 | 4 | 4 | 8 | 5 |
| **Total** | **109** | **7** | **19** | **49** | **34** |

---

## Priority 1: Critical Issues (Must Fix Immediately)

### P2P-CRIT-1: Unrestricted Peer Auto-Registration (DoS Vector)
- **File:** `pkg/node/peer_registry.go`
- **Issue:** Any node can register as a peer without authentication, enabling DoS attacks
- **Fix:** Implement peer allowlist or require cryptographic proof before registration
- **Impact:** Network can be flooded with malicious peer registrations

### P2P-CRIT-2: No Message Size Limits (Memory Exhaustion)
- **File:** `pkg/node/transport.go`
- **Issue:** Incoming messages have no size cap, allowing memory exhaustion attacks
- **Fix:** Add `MaxMessageSize` config (e.g., 1MB) and reject oversized messages
- **Impact:** Single malicious peer can crash nodes via large message payloads

### P2P-CRIT-3: Connection Limit Bypass During Handshake
- **File:** `pkg/node/transport.go`
- **Issue:** Connection count checked after handshake, allowing limit bypass
- **Fix:** Check connection count BEFORE accepting WebSocket upgrade
- **Impact:** Node can be overwhelmed with connections during handshake phase

### P2P-CRIT-4: Challenge-Response Auth Not Implemented
- **File:** `pkg/node/transport.go`, `pkg/node/handshake.go`
- **Issue:** Peer identity claimed but not cryptographically verified
- **Fix:** Implement challenge-response using X25519 keypairs during handshake
- **Impact:** Peers can impersonate other nodes

### TEST-CRIT-1: No Tests for auth.go (Security-Critical)
- **File:** `pkg/mining/auth.go` (missing `auth_test.go`)
- **Issue:** Authentication code has zero test coverage
- **Fix:** Create `auth_test.go` with tests for BasicAuth, DigestAuth, nonce management
- **Impact:** Security regressions can ship undetected

### TEST-CRIT-2: No Tests for profile_manager.go
- **File:** `pkg/mining/profile_manager.go` (missing tests)
- **Issue:** Profile persistence logic untested
- **Fix:** Create `profile_manager_test.go` covering CRUD operations
- **Impact:** Profile corruption/loss bugs can ship undetected

### TEST-CRIT-3: No Tests for ttminer.go
- **File:** `pkg/mining/ttminer.go` (missing tests)
- **Issue:** TTMiner implementation completely untested
- **Fix:** Create `ttminer_test.go` with startup/config/stats tests
- **Impact:** TTMiner regressions shipped without detection

---

## Priority 2: High Severity Issues

### CONC-HIGH-1: Race Condition in wsClient.miners Map
- **File:** `pkg/mining/events.go`
- **Severity:** HIGH
- **Issue:** `wsClient.miners` map accessed without synchronization from multiple goroutines
- **Fix:** Add `sync.RWMutex` to protect map access, or use `sync.Map`
- **Impact:** Can cause panics under concurrent access

### RESIL-HIGH-1: Missing recover() in Stats Collection Goroutines
- **File:** `pkg/mining/manager.go` (lines 544-632)
- **Severity:** HIGH
- **Issue:** Background stats collection has no panic recovery
- **Fix:** Add `defer func() { if r := recover(); r != nil { ... } }()` to goroutines
- **Impact:** Panic in stats collection crashes entire service

### RESIL-HIGH-2: Profile Manager Init Failure Blocks Entire Service
- **File:** `pkg/mining/service.go` (NewService)
- **Severity:** HIGH
- **Issue:** ProfileManager failure in NewService() prevents service startup
- **Fix:** Make ProfileManager optional, log warning but continue with degraded mode
- **Impact:** Corrupted profile file makes entire application unusable

### RESIL-HIGH-3: GitHub API Calls Without Circuit Breaker
- **File:** `pkg/mining/xmrig.go` (GetLatestVersion)
- **Severity:** HIGH
- **Issue:** GitHub API rate limits or outages cascade to service degradation
- **Fix:** Implement circuit breaker pattern with fallback to cached version
- **Impact:** GitHub outage blocks miner installation/updates

### PERF-HIGH-1: No Connection Pooling for HTTP Client
- **File:** `pkg/mining/miner.go`, `pkg/mining/xmrig.go`
- **Severity:** HIGH
- **Issue:** HTTP client may create new connections per request
- **Fix:** Use shared `http.Client` with configured transport and connection pool
- **Impact:** Unnecessary TCP overhead, potential connection exhaustion

### PERF-HIGH-2: JSON Encoding Without Buffer Pool
- **File:** `pkg/mining/events.go`, `pkg/mining/service.go`
- **Severity:** HIGH
- **Issue:** JSON marshaling allocates new buffers per operation
- **Fix:** Use `sync.Pool` for JSON encoder buffers
- **Impact:** GC pressure under high message throughput

### API-HIGH-1: Inconsistent Error Response Format
- **File:** `pkg/mining/service.go`, `pkg/mining/node_service.go`
- **Severity:** HIGH
- **Issue:** Some endpoints return `{"error": "..."}`, others return `{"code": "...", "message": "..."}`
- **Fix:** Standardize all errors to APIError struct format
- **Impact:** Client code cannot reliably parse error responses

### API-HIGH-2: Missing Input Validation on Critical Endpoints
- **File:** `pkg/mining/service.go` (handleStartMiner)
- **Severity:** HIGH
- **Issue:** Miner config accepts arbitrary values without validation
- **Fix:** Add validation for pool URLs, wallet addresses, algorithm values
- **Impact:** Malformed configs can cause unexpected behavior

### TEST-HIGH-1: No Integration Tests for WebSocket Events
- **File:** `pkg/mining/events.go`
- **Severity:** HIGH
- **Issue:** WebSocket event broadcasting untested
- **Fix:** Create integration test with mock WebSocket clients
- **Impact:** Event delivery bugs undetected

### TEST-HIGH-2: No End-to-End Tests for P2P Communication
- **File:** `pkg/node/*.go`
- **Severity:** HIGH
- **Issue:** P2P message exchange not tested end-to-end
- **Fix:** Create tests with two nodes exchanging messages
- **Impact:** Protocol bugs ship undetected

### TEST-HIGH-3: No Tests for Miner Installation Flow
- **File:** `pkg/mining/miner.go` (InstallMiner)
- **Severity:** HIGH
- **Issue:** Download/extract/verify flow untested
- **Fix:** Create tests with mock HTTP server serving test binaries
- **Impact:** Installation failures not caught in CI

### TEST-HIGH-4: No Stress/Load Tests
- **File:** N/A
- **Severity:** HIGH
- **Issue:** No tests for behavior under concurrent load
- **Fix:** Add benchmark tests simulating multiple miners/connections
- **Impact:** Performance regressions undetected

### TEST-HIGH-5: No Tests for Database Migrations
- **File:** `pkg/database/database.go`
- **Severity:** HIGH
- **Issue:** Schema creation untested, no migration tests
- **Fix:** Test Initialize() with fresh DB and existing DB scenarios
- **Impact:** Database schema bugs can corrupt user data

### ARCH-HIGH-1: Global Database State
- **File:** `pkg/database/database.go`
- **Severity:** HIGH
- **Issue:** Package-level `var db *sql.DB` creates tight coupling
- **Fix:** Create Database interface, use dependency injection
- **Impact:** Hard to test, prevents database backend swapping

### ARCH-HIGH-2: Manager Violates Single Responsibility
- **File:** `pkg/mining/manager.go`
- **Severity:** HIGH
- **Issue:** Manager handles lifecycle, stats, config, persistence
- **Fix:** Extract StatsCollector, ConfigRepository as separate concerns
- **Impact:** Large file (700+ lines), hard to maintain

### P2P-HIGH-1: No Peer Scoring/Reputation System
- **File:** `pkg/node/peer_registry.go`
- **Severity:** HIGH
- **Issue:** All peers treated equally regardless of behavior
- **Fix:** Implement scoring based on response time, errors, uptime
- **Impact:** Misbehaving peers not penalized

### P2P-HIGH-2: No Message Deduplication
- **File:** `pkg/node/transport.go`
- **Severity:** HIGH
- **Issue:** Duplicate messages processed repeatedly
- **Fix:** Track message IDs with TTL cache, reject duplicates
- **Impact:** Amplification attacks possible

### P2P-HIGH-3: Handshake Timeout Too Long
- **File:** `pkg/node/transport.go`
- **Severity:** HIGH
- **Issue:** Default handshake timeout allows resource exhaustion
- **Fix:** Reduce handshake timeout to 5-10 seconds
- **Impact:** Slow-loris style attacks possible

### P2P-HIGH-4: No Rate Limiting Per Peer
- **File:** `pkg/node/transport.go`
- **Severity:** HIGH
- **Issue:** Single peer can flood node with messages
- **Fix:** Implement per-peer message rate limiting
- **Impact:** Single peer can degrade performance for all

---

## Priority 3: Medium Severity Issues

### SEC-MED-1: Timing Attack in Password Comparison
- **File:** `pkg/mining/auth.go`
- **Issue:** Password comparison may not be constant-time in all paths
- **Fix:** Ensure all password comparisons use `subtle.ConstantTimeCompare`

### SEC-MED-2: Nonce Entropy Could Be Improved
- **File:** `pkg/mining/auth.go`
- **Issue:** Nonce generation uses crypto/rand but format could be stronger
- **Fix:** Consider using UUIDv4 or longer nonce values

### SEC-MED-3: No CSRF Protection on State-Changing Endpoints
- **File:** `pkg/mining/service.go`
- **Issue:** POST/PUT/DELETE endpoints lack CSRF tokens
- **Fix:** Add CSRF middleware for non-API browser access

### SEC-MED-4: API Keys Stored in Plaintext
- **File:** `pkg/mining/settings_manager.go`
- **Issue:** Pool API keys stored unencrypted
- **Fix:** Encrypt sensitive fields using system keyring or derived key

### CONC-MED-1: Potential Deadlock in Manager.Stop()
- **File:** `pkg/mining/manager.go`
- **Issue:** Stop() acquires locks in different order than other methods
- **Fix:** Audit lock ordering, document expected lock acquisition order

### CONC-MED-2: Channel Close Race in Events
- **File:** `pkg/mining/events.go`
- **Issue:** Event channel close can race with sends
- **Fix:** Use done channel pattern or atomic state flag

### CONC-MED-3: Stats Collection Without Context Deadline
- **File:** `pkg/mining/manager.go` (lines 588-594)
- **Issue:** Stats timeout doesn't propagate to database writes
- **Fix:** Pass context to database operations

### CONC-MED-4: RWMutex Downgrade Pattern Missing
- **File:** `pkg/mining/manager.go`
- **Issue:** Some operations hold write lock when read lock sufficient
- **Fix:** Downgrade to RLock where possible

### CONC-MED-5: Event Hub Broadcast Blocking
- **File:** `pkg/mining/events.go`
- **Issue:** Slow client can block broadcasts to all clients
- **Fix:** Use buffered channels or drop messages for slow clients

### PERF-MED-1: SQL Queries Missing Indexes
- **File:** `pkg/database/hashrate.go`
- **Issue:** Query by miner_name without index on frequent queries
- **Fix:** Add indexes: `CREATE INDEX idx_miner_name ON hashrate_points(miner_name)`

### PERF-MED-2: Logger Creates Allocations Per Call
- **File:** `pkg/logging/logger.go`
- **Issue:** Fields map allocated on every log call
- **Fix:** Use pre-allocated field pools or structured logging library

### PERF-MED-3: Config File Read On Every Access
- **File:** `pkg/mining/config_manager.go`
- **Issue:** Config read from disk on each access
- **Fix:** Cache config in memory, reload on file change (fsnotify)

### PERF-MED-4: HTTP Response Body Not Drained Consistently
- **File:** `pkg/mining/xmrig.go`, `pkg/mining/xmrig_stats.go`
- **Issue:** Error paths don't drain response body
- **Fix:** Always `io.Copy(io.Discard, resp.Body)` before close on errors

### PERF-MED-5: No Database Connection Pooling Tuning
- **File:** `pkg/database/database.go`
- **Issue:** Default SQLite connection pool settings
- **Fix:** Configure `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`

### PERF-MED-6: JSON Unmarshal Into Interface{}
- **File:** `pkg/node/controller.go`
- **Issue:** `json.Unmarshal` into `interface{}` prevents optimization
- **Fix:** Use typed structs for all message payloads

### RESIL-MED-1: No Retry for Failed Database Writes
- **File:** `pkg/mining/manager.go`
- **Issue:** Single database write failure loses data
- **Fix:** Implement retry with exponential backoff for DB writes

### RESIL-MED-2: No Graceful Degradation for Missing Miners
- **File:** `pkg/mining/manager.go`
- **Issue:** Missing miner binary fails hard
- **Fix:** Return degraded status, offer installation prompt

### RESIL-MED-3: WebSocket Reconnection Not Automatic
- **File:** `pkg/mining/events.go`
- **Issue:** Disconnected clients not automatically reconnected
- **Fix:** Implement client-side reconnection with backoff (UI concern)

### RESIL-MED-4: No Health Check Endpoint
- **File:** `pkg/mining/service.go`
- **Issue:** No `/health` or `/ready` endpoints for orchestration
- **Fix:** Add health check with component status reporting

### RESIL-MED-5: Transport Failure Doesn't Notify Peers
- **File:** `pkg/node/transport.go`
- **Issue:** Node shutdown doesn't send disconnect to peers
- **Fix:** Send graceful shutdown message before closing connections

### RESIL-MED-6: No Watchdog for Background Tasks
- **File:** `pkg/mining/manager.go`
- **Issue:** No monitoring of background goroutine health
- **Fix:** Implement supervisor pattern with restart capability

### RESIL-MED-7: Config Corruption Recovery Missing
- **File:** `pkg/mining/config_manager.go`
- **Issue:** Corrupted JSON file fails silently or crashes
- **Fix:** Implement backup/restore with validation

### RESIL-MED-8: No Request Timeout Middleware
- **File:** `pkg/mining/service.go`
- **Issue:** Long-running requests not bounded
- **Fix:** Add timeout middleware (e.g., 30s default)

### API-MED-1: Missing Pagination on List Endpoints
- **File:** `pkg/mining/service.go` (handleListMiners, handleListProfiles)
- **Issue:** All results returned at once
- **Fix:** Add `?limit=N&offset=M` query parameters

### API-MED-2: No HATEOAS Links in Responses
- **File:** `pkg/mining/service.go`
- **Issue:** Clients must construct URLs manually
- **Fix:** Add `_links` object with related resource URLs

### API-MED-3: PUT Should Return 404 for Missing Resources
- **File:** `pkg/mining/service.go` (handleUpdateProfile)
- **Issue:** PUT on non-existent profile creates it (should be POST)
- **Fix:** Return 404 if profile doesn't exist, use POST for creation

### API-MED-4: DELETE Not Idempotent
- **File:** `pkg/mining/service.go` (handleDeleteProfile)
- **Issue:** DELETE on missing resource returns error
- **Fix:** Return 204 No Content for already-deleted resources

### API-MED-5: No Request ID in Responses
- **File:** `pkg/mining/service.go`
- **Issue:** Hard to correlate requests with logs
- **Fix:** Return X-Request-ID header in all responses

### API-MED-6: Version Not in URL Path
- **File:** `pkg/mining/service.go`
- **Issue:** API versioning only via base path
- **Fix:** Document versioning strategy, consider Accept header versioning

### API-MED-7: No Cache Headers
- **File:** `pkg/mining/service.go`
- **Issue:** Static-ish resources (miner list) not cacheable
- **Fix:** Add Cache-Control headers for appropriate endpoints

### API-MED-8: Missing Content-Type Validation
- **File:** `pkg/mining/service.go`
- **Issue:** JSON endpoints don't validate Content-Type header
- **Fix:** Require `Content-Type: application/json` for POST/PUT

### ARCH-MED-1: No Interface for Miner Configuration
- **File:** `pkg/mining/mining.go`
- **Issue:** Config struct tightly coupled to XMRig fields
- **Fix:** Create ConfigBuilder interface for miner-specific configs

### ARCH-MED-2: Event Types as Strings
- **File:** `pkg/mining/events.go`
- **Issue:** Event types are magic strings
- **Fix:** Use typed constants or enums

### ARCH-MED-3: Circular Import Risk
- **File:** `pkg/mining/`, `pkg/node/`
- **Issue:** Service.go imports node, node imports mining types
- **Fix:** Extract shared types to `pkg/types/` package

### ARCH-MED-4: No Plugin Architecture for Miners
- **File:** `pkg/mining/`
- **Issue:** Adding new miner requires modifying manager.go
- **Fix:** Implement miner registry with auto-discovery

### ARCH-MED-5: Settings Scattered Across Multiple Managers
- **File:** `pkg/mining/config_manager.go`, `pkg/mining/settings_manager.go`, `pkg/mining/profile_manager.go`
- **Issue:** Three different config file managers
- **Fix:** Unify into single ConfigRepository with namespaces

### ARCH-MED-6: BaseMiner Has Too Many Responsibilities
- **File:** `pkg/mining/miner.go`
- **Issue:** BaseMiner handles download, extract, process, stats
- **Fix:** Extract Downloader, Extractor as separate services

### ARCH-MED-7: Missing Factory Pattern for Service Creation
- **File:** `pkg/mining/service.go`
- **Issue:** NewService() directly instantiates all dependencies
- **Fix:** Use factory/builder pattern for testable construction

### P2P-MED-1: No Message Versioning
- **File:** `pkg/node/messages.go`
- **Issue:** No protocol version negotiation
- **Fix:** Add version field to handshake, reject incompatible versions

### P2P-MED-2: Peer Discovery Not Implemented
- **File:** `pkg/node/`
- **Issue:** Peers must be manually added
- **Fix:** Implement mDNS/DHT peer discovery for local networks

### P2P-MED-3: No Encryption for Message Payloads
- **File:** `pkg/node/transport.go`
- **Issue:** Relying on WSS only, no end-to-end encryption
- **Fix:** Encrypt payloads with session key from handshake

### P2P-MED-4: Connection State Machine Incomplete
- **File:** `pkg/node/transport.go`
- **Issue:** Connection states (connecting, handshaking, connected) informal
- **Fix:** Implement explicit state machine with transitions

### P2P-MED-5: No Keepalive/Heartbeat
- **File:** `pkg/node/transport.go`
- **Issue:** Dead connections not detected until send fails
- **Fix:** Implement periodic ping/pong heartbeat

### P2P-MED-6: Broadcast Doesn't Exclude Sender
- **File:** `pkg/node/controller.go`
- **Issue:** Broadcast messages may echo back to originator
- **Fix:** Filter sender from broadcast targets

### P2P-MED-7: No Message Priority Queuing
- **File:** `pkg/node/transport.go`
- **Issue:** All messages treated equally
- **Fix:** Implement priority queues (control > stats > logs)

### P2P-MED-8: Missing Graceful Reconnection
- **File:** `pkg/node/transport.go`
- **Issue:** Disconnected peers not automatically reconnected
- **Fix:** Implement reconnection with exponential backoff

### TEST-MED-1: Mock Objects Not Standardized
- **File:** Various test files
- **Issue:** Each test creates ad-hoc mocks
- **Fix:** Create `pkg/mocks/` with reusable mock implementations

### TEST-MED-2: No Table-Driven Tests
- **File:** Various test files
- **Issue:** Test cases not parameterized
- **Fix:** Convert to table-driven tests for better coverage

### TEST-MED-3: Test Coverage Not Enforced
- **File:** CI configuration
- **Issue:** No coverage threshold in CI
- **Fix:** Add coverage gate (e.g., fail below 70%)

---

## Priority 4: Low Severity Issues

### SEC-LOW-1: Debug Logging May Expose Sensitive Data
- **File:** `pkg/logging/logger.go`
- **Fix:** Implement field sanitization for debug logs

### SEC-LOW-2: No Rate Limit on Auth Failures
- **File:** `pkg/mining/auth.go`
- **Fix:** Track failed attempts, implement exponential backoff

### SEC-LOW-3: CORS Allows All Origins in Dev Mode
- **File:** `pkg/mining/service.go`
- **Fix:** Restrict CORS origins in production config

### SEC-LOW-4: No Security Headers Middleware
- **File:** `pkg/mining/service.go`
- **Fix:** Add X-Content-Type-Options, X-Frame-Options, etc.

### CONC-LOW-1: Debug Log Counter Not Perfectly Accurate
- **File:** `pkg/node/transport.go`
- **Fix:** Accept approximate counting or use atomic load-modify-store

### CONC-LOW-2: Metrics Histogram Lock Contention
- **File:** `pkg/mining/metrics.go`
- **Fix:** Use sharded histogram or lock-free ring buffer

### CONC-LOW-3: Channel Buffer Sizes Arbitrary
- **File:** Various files
- **Fix:** Document rationale for buffer sizes, tune based on profiling

### PERF-LOW-1: Repeated Type Assertions
- **File:** `pkg/node/controller.go`
- **Fix:** Store typed references after initial assertion

### PERF-LOW-2: String Concatenation in Loops
- **File:** Various files
- **Fix:** Use strings.Builder for concatenation

### PERF-LOW-3: Map Pre-allocation Missing
- **File:** Various files
- **Fix:** Use `make(map[K]V, expectedSize)` where size is known

### PERF-LOW-4: Unnecessary JSON Re-encoding
- **File:** `pkg/node/messages.go`
- **Fix:** Cache encoded messages when broadcasting

### RESIL-LOW-1: Exit Codes Not Semantic
- **File:** `cmd/mining/main.go`
- **Fix:** Define exit codes for different failure modes

### RESIL-LOW-2: No Startup Banner Version Info
- **File:** `cmd/mining/main.go`
- **Fix:** Log version, commit hash, build date on startup

### RESIL-LOW-3: Signal Handling Incomplete
- **File:** `cmd/mining/main.go`
- **Fix:** Handle SIGHUP for config reload

### RESIL-LOW-4: Temp Files Not Cleaned on Crash
- **File:** `pkg/mining/miner.go`
- **Fix:** Use defer for temp file cleanup, implement crash recovery

### RESIL-LOW-5: No Startup Self-Test
- **File:** `pkg/mining/service.go`
- **Fix:** Add startup validation (DB connection, file permissions)

### RESIL-LOW-6: Log Rotation Not Configured
- **File:** `pkg/logging/logger.go`
- **Fix:** Document log rotation setup (logrotate.d)

### API-LOW-1: OPTIONS Response Missing Allow Header
- **File:** `pkg/mining/service.go`
- **Fix:** Include allowed methods in OPTIONS responses

### API-LOW-2: No ETag Support
- **File:** `pkg/mining/service.go`
- **Fix:** Add ETag headers for conditional GET requests

### API-LOW-3: No OpenAPI Examples
- **File:** Swagger annotations
- **Fix:** Add example values to Swagger annotations

### API-LOW-4: Inconsistent Field Naming (camelCase vs snake_case)
- **File:** Various JSON responses
- **Fix:** Standardize on camelCase for JSON

### API-LOW-5: No Deprecation Headers
- **File:** `pkg/mining/service.go`
- **Fix:** Add Sunset header support for deprecated endpoints

### API-LOW-6: Missing Link Header for Collections
- **File:** `pkg/mining/service.go`
- **Fix:** Add RFC 5988 Link headers for pagination

### ARCH-LOW-1: Package Comments Missing
- **File:** All packages
- **Fix:** Add godoc package comments

### ARCH-LOW-2: Exported Functions Without Godoc
- **File:** Various files
- **Fix:** Add godoc comments to all exported functions

### ARCH-LOW-3: Magic Numbers in Code
- **File:** Various files
- **Fix:** Extract to named constants with documentation

### ARCH-LOW-4: No Makefile Target for Docs
- **File:** `Makefile`
- **Fix:** Add `make godoc` target

### ARCH-LOW-5: Missing Architecture Decision Records
- **File:** `docs/`
- **Fix:** Create `docs/adr/` directory with key decisions

### P2P-LOW-1: Peer List Not Sorted
- **File:** `pkg/node/peer_registry.go`
- **Fix:** Sort by score or name for consistent ordering

### P2P-LOW-2: Debug Messages Verbose
- **File:** `pkg/node/transport.go`
- **Fix:** Add log levels, reduce default verbosity

### P2P-LOW-3: Peer Names Not Validated
- **File:** `pkg/node/peer_registry.go`
- **Fix:** Validate peer names (length, characters)

### P2P-LOW-4: No Connection Metrics Export
- **File:** `pkg/node/transport.go`
- **Fix:** Export Prometheus metrics for connections

### P2P-LOW-5: Message Types Not Documented
- **File:** `pkg/node/messages.go`
- **Fix:** Add godoc with message format examples

### TEST-LOW-1: Test Output Verbose
- **File:** Various test files
- **Fix:** Use t.Log() only for failures

---

## Quick Wins (Implement First)

These changes provide high value with minimal effort:

1. **Add mutex to wsClient.miners map** (CONC-HIGH-1)
   - 5 minutes, prevents panics

2. **Add recover() to background goroutines** (RESIL-HIGH-1)
   - 10 minutes, prevents service crashes

3. **Add message size limit to P2P transport** (P2P-CRIT-2)
   - 15 minutes, prevents memory exhaustion

4. **Check connection count before handshake** (P2P-CRIT-3)
   - 10 minutes, closes DoS vector

5. **Create auth_test.go with basic coverage** (TEST-CRIT-1)
   - 30 minutes, covers security-critical code

6. **Add circuit breaker for GitHub API** (RESIL-HIGH-3)
   - 20 minutes, improves resilience

---

## Implementation Roadmap

### Phase 1: Security Hardening (Week 1)
- P2P-CRIT-1 through P2P-CRIT-4
- TEST-CRIT-1
- CONC-HIGH-1

### Phase 2: Stability (Week 2)
- RESIL-HIGH-1 through RESIL-HIGH-3
- PERF-HIGH-1, PERF-HIGH-2
- All Medium concurrency issues

### Phase 3: API Polish (Week 3)
- API-HIGH-1, API-HIGH-2
- All Medium API issues
- API documentation improvements

### Phase 4: Testing Infrastructure (Week 4)
- TEST-HIGH-1 through TEST-HIGH-5
- TEST-CRIT-2, TEST-CRIT-3
- Coverage gates in CI

### Phase 5: Architecture Cleanup (Ongoing)
- ARCH-HIGH-1, ARCH-HIGH-2
- Interface extractions
- Documentation

---

## Conclusion

This code review represents a comprehensive analysis by 8 specialized AI agents examining security, concurrency, performance, resilience, testing, API design, architecture, and P2P networking domains. The 109 findings range from critical security issues to low-priority improvements.

**Key Statistics:**
- 7 Critical issues (all in P2P/Testing)
- 19 High severity issues
- 49 Medium severity issues
- 34 Low severity improvements

The codebase demonstrates solid fundamentals with comprehensive error handling already in place. These findings represent the difference between "good enough" and "production-hardened" code.

---

*Generated by 8 Opus 4.5 agents as part of human-AI collaborative code review.*
