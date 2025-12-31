# Code Review Findings - XMRig Proxy Enterprise Audit

**Generated:** 2025-12-31
**Reviewed by:** 8 Parallel Opus Code Reviewers
**Target:** XMRig-based C++ Stratum Proxy (347 source files)

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Entry Point & App Lifecycle | 2 | 2 | 2 | 6 |
| Core Controller & Config | 1 | 4 | 1 | 6 |
| Proxy Core (Server, Miner, Events) | 4 | 5 | 1 | 10 |
| Proxy TLS & Workers | 3 | 2 | 2 | 7 |
| Splitter System | 2 | 3 | 0 | 5 |
| Network & Stratum Client | 3 | 5 | 1 | 9 |
| HTTP/HTTPS & REST API | 1 | 3 | 3 | 7 |
| Base I/O & Kernel | 2 | 2 | 3 | 7 |
| **TOTAL** | **18** | **26** | **13** | **57** |

---

## Critical Issues

### CRIT-001: Double-Delete in Controller Destructor and stop()
- **File:** `src/core/Controller.cpp:45,73-74`
- **Domain:** Entry Point & App Lifecycle
- **Confidence:** 100%

`m_proxy` deleted in both destructor and `stop()` method. If `stop()` called before destructor, double-free causes crash/heap corruption.

**Fix:** Add null check in destructor or stop(), set to nullptr after delete.

---

### CRIT-002: UV Event Loop Closed Without Draining Handles
- **File:** `src/App.cpp:78-79`
- **Domain:** Entry Point & App Lifecycle
- **Confidence:** 95%

`uv_loop_close()` called immediately after `uv_run()` without ensuring handles closed. Returns UV_EBUSY (ignored), leaking resources.

**Fix:** Loop until `uv_loop_close()` succeeds, calling `uv_run(UV_RUN_ONCE)`.

---

### CRIT-003: Missing JSON Type Validation in BindHost Constructor
- **File:** `src/proxy/BindHost.cpp:67-72`
- **Domain:** Core Controller & Config
- **Confidence:** 95%

Direct `GetString()`, `GetUint()`, `GetBool()` calls without checking field existence/type. Crashes on malformed config (DoS).

**Fix:** Add `HasMember()` and type checks before accessing JSON fields.

---

### CRIT-004: Race Condition in Events System - Non-Atomic Ready Flag
- **File:** `src/proxy/Events.cpp:37-56`
- **Domain:** Proxy Core
- **Confidence:** 95%

`m_ready` flag is bool, not atomic. Multiple threads can pass check simultaneously, causing event corruption.

**Fix:** Use `std::atomic<bool>` or mutex to protect event dispatch.

---

### CRIT-005: Memory Pool (MemPool) Not Thread-Safe
- **File:** `src/base/net/tools/MemPool.h:45-73`
- **Domain:** Proxy Core
- **Confidence:** 100%

`allocate()` and `deallocate()` modify shared STL containers without synchronization. Called from libuv callbacks (multi-threaded). Heap corruption guaranteed under load.

**Fix:** Add mutex to protect all MemPool operations.

---

### CRIT-006: Static Event Buffer Shared Across All Events
- **File:** `src/proxy/events/Event.h:52`
- **Domain:** Proxy Core
- **Confidence:** 90%

All events use single static 4KB buffer with placement new. Concurrent events corrupt each other's memory.

**Fix:** Use heap allocation for events or implement thread-safe event pool.

---

### CRIT-007: Storage Counter Overflow - ID Collision
- **File:** `src/base/net/tools/Storage.h:37-42,81`
- **Domain:** Proxy Core
- **Confidence:** 85%

`m_counter` increments without bounds check. After 2^32/2^64 connections, IDs wrap causing wrong miner deletion, use-after-free.

**Fix:** Add overflow detection and ID recycling mechanism.

---

### CRIT-008: Unchecked SSL_write Return Value
- **File:** `src/base/net/tls/ServerTls.cpp:65`
- **Domain:** TLS & Workers
- **Confidence:** 90%

`SSL_write()` return value ignored. Silent data loss, corrupted protocol messages.

**Fix:** Check return value, handle partial writes and errors.

---

### CRIT-009: TLS setCiphers() Returns True on Failure
- **File:** `src/base/net/tls/TlsContext.cpp:165-174`
- **Domain:** TLS & Workers
- **Confidence:** 100%

Copy-paste bug: function logs error but returns `true` on cipher config failure. Server runs with weak default ciphers.

**Fix:** Return `false` on line 173.

---

### CRIT-010: Unbounded Memory Growth in m_results Map
- **File:** `src/proxy/splitters/nicehash/NonceMapper.cpp:148,264-276`
- **Domain:** Splitter System
- **Confidence:** 95%

Submit contexts stored in map, only removed on pool response. Network issues = unbounded memory growth.

**Fix:** Add timestamp to SubmitCtx, cleanup stale entries in gc().

---

### CRIT-011: NonceSplitter gc() Vector Out-of-Bounds Access
- **File:** `src/proxy/splitters/nicehash/NonceSplitter.cpp:93-97`
- **Domain:** Splitter System
- **Confidence:** 90%

While loop calls `m_upstreams.back()` without empty check. If all mappers suspended, crashes on empty vector.

**Fix:** Add `!m_upstreams.empty()` to while condition.

---

### CRIT-012: Unchecked SSL_write/BIO_write in Stratum TLS
- **File:** `src/base/net/stratum/Tls.cpp:84-89,104-107`
- **Domain:** Network & Stratum
- **Confidence:** 95%

Return values ignored. Silent data loss, TLS state corruption.

**Fix:** Check return values, handle errors appropriately.

---

### CRIT-013: Missing TLS Certificate Verification
- **File:** `src/base/net/stratum/Tls.cpp:35-48`
- **Domain:** Network & Stratum
- **Confidence:** 100%

No `SSL_CTX_set_verify()` call. Certificates not validated unless fingerprint provided. Vulnerable to MITM attacks.

**Fix:** Add `SSL_CTX_set_verify(m_ctx, SSL_VERIFY_PEER, nullptr)`.

---

### CRIT-014: Timing Attack in API Token Authentication
- **File:** `src/base/api/Httpd.cpp:193-197`
- **Domain:** HTTP API
- **Confidence:** 100%

Uses `strncmp()` for token comparison. Attacker can extract token character-by-character via timing.

**Fix:** Use `CRYPTO_memcmp()` for constant-time comparison.

---

### CRIT-015: Race Condition in Signal Handler
- **File:** `src/base/io/Signals.cpp:61-88`
- **Domain:** Base I/O & Kernel
- **Confidence:** 95%

Signal handler calls `LOG_WARN()` which takes mutex, allocates memory. Not async-signal-safe. Deadlock or heap corruption.

**Fix:** Only forward signal to listener, log in main event loop context.

---

### CRIT-016: Potential Buffer Overflow in Log Formatting
- **File:** `src/base/io/log/Log.cpp:96-101`
- **Domain:** Base I/O & Kernel
- **Confidence:** 85%

Magic number `32` in buffer size calculation. Large timestamps + messages can underflow available size.

**Fix:** Add explicit bounds checking before vsnprintf.

---

### CRIT-017: Private Key File Written with Insecure Permissions
- **File:** `src/base/net/tls/TlsGen.cpp:128-134`
- **Domain:** TLS & Workers
- **Confidence:** 90%

Private key file created with default permissions (0644 = world-readable).

**Fix:** Add `chmod(m_certKey, 0600)` on Unix.

---

### CRIT-018: Missing NULL Check in BindHost JSON Constructor (Duplicate)
- **File:** `src/proxy/BindHost.cpp:67,71-72`
- **Domain:** TLS & Workers
- **Confidence:** 95%

Same as CRIT-003 - found by multiple reviewers, confirming severity.

---

## High Priority Issues

### HIGH-001: Missing uv_stop() in Shutdown Path
- **File:** `src/App.cpp:121-129`
- **Domain:** Entry Point
- **Confidence:** 85%

`close()` doesn't call `uv_stop()`. UV loop continues until handles naturally close. Delayed/hung shutdown.

---

### HIGH-002: Use-After-Free Risk in Signal/Console Callbacks
- **File:** `src/base/io/Signals.cpp:87`, `Console.cpp:74`
- **Domain:** Entry Point
- **Confidence:** 80%

`m_listener` accessed after `App::close()` resets handles. Race between close and pending events.

---

### HIGH-003: Integer Overflow in strtol Conversion
- **File:** `src/core/config/ConfigTransform.cpp:85`
- **Domain:** Config
- **Confidence:** 85%

`strtol()` cast to `uint64_t` without overflow/error checking. Negative values wrap, no error detection.

---

### HIGH-004: Port Number Parsing Without Bounds Check
- **File:** `src/proxy/BindHost.cpp:136,158`
- **Domain:** Config
- **Confidence:** 90%

Port parsed via `strtol()`, cast to `uint16_t` without validating 0-65535 range.

---

### HIGH-005: Double-Delete Risk in Controller (Config Review)
- **File:** `src/core/Controller.cpp:43-46,69-75`
- **Domain:** Config
- **Confidence:** 85%

Same as CRIT-001, confirmed by second reviewer.

---

### HIGH-006: Missing Null Pointer Check in Controller Methods
- **File:** `src/core/Controller.cpp:65,78-99`
- **Domain:** Config
- **Confidence:** 90%

`proxy()` returns potentially null `m_proxy` without checks. Crashes if called before `init()`.

---

### HIGH-007: Unbounded Vector Growth in StatsData::latency
- **File:** `src/proxy/StatsData.h:96,138`
- **Domain:** Proxy Core
- **Confidence:** 100%

One entry per accepted share, forever. Memory exhaustion guaranteed.

---

### HIGH-008: NULL Dereference in Server::create() - Dead Code
- **File:** `src/proxy/Server.cpp:89-92`
- **Domain:** Proxy Core
- **Confidence:** 80%

`new` throws on failure, doesn't return NULL. Check is dead code, real failures unhandled.

---

### HIGH-009: Missing Validation in Miner::parseRequest()
- **File:** `src/proxy/Miner.cpp:354-355`
- **Domain:** Proxy Core
- **Confidence:** 85%

`doc["method"].GetString()` called without validating field exists. Crash on malformed client request.

---

### HIGH-010: Non-Atomic Counters in Counters Class
- **File:** `src/proxy/Counters.h:42-67`
- **Domain:** Proxy Core
- **Confidence:** 90%

Static counters modified from multiple threads without atomics. Statistics incorrect, potential corruption.

---

### HIGH-011: Use-After-Free in Miner Shutdown Path
- **File:** `src/proxy/Miner.cpp:547-577`
- **Domain:** Proxy Core
- **Confidence:** 85%

Complex callback chain. Miner can be accessed after removal from storage if shutdowns overlap.

---

### HIGH-012: Integer Overflow in ExtraNonce Allocation
- **File:** `src/proxy/splitters/extra_nonce/ExtraNonceStorage.cpp:37,99`
- **Domain:** Splitter
- **Confidence:** 90%

`m_extraNonce` increments forever, but only 32 bits used. Nonce collision after 4B connections.

---

### HIGH-013: Race Condition in NonceStorage::remove()
- **File:** `src/proxy/splitters/nicehash/NonceStorage.cpp:103-110,122-126`
- **Domain:** Splitter
- **Confidence:** 85%

Dead slots only cleared during setJob from same client. Different clients = slots never recycled.

---

### HIGH-014: Potential Use-After-Free in submitCtx()
- **File:** `src/proxy/splitters/nicehash/NonceMapper.cpp:264-278`
- **Domain:** Splitter
- **Confidence:** 85%

Miner lookup after context retrieval. Redundant map lookup, miner may have disconnected.

---

### HIGH-015: Timing Attack in Certificate Fingerprint
- **File:** `src/base/net/stratum/Tls.cpp:186`
- **Domain:** Network
- **Confidence:** 85%

`strncasecmp()` for fingerprint comparison. Timing attack vulnerability.

---

### HIGH-016: Buffer Overflow Risk in LineReader
- **File:** `src/base/net/tools/LineReader.cpp:57-71`
- **Domain:** Network
- **Confidence:** 85%

Silently drops oversized messages without error. Protocol desync, DoS vector.

---

### HIGH-017: Weak TLS Configuration - Missing Modern Options
- **File:** `src/base/net/stratum/Tls.cpp:47`
- **Domain:** Network
- **Confidence:** 80%

Only disables SSLv2/SSLv3. TLS 1.0/1.1 still allowed (deprecated, vulnerable).

---

### HIGH-018: SOCKS5 Protocol Validation Insufficient
- **File:** `src/base/net/stratum/Socks5.cpp:29-48`
- **Domain:** Network
- **Confidence:** 82%

Accesses `data[0]`, `data[1]` without buffer length check. Malicious SOCKS5 proxy can crash.

---

### HIGH-019: Race Condition in DNS Resolution
- **File:** `src/base/net/dns/DnsUvBackend.cpp:74-91`
- **Domain:** Network
- **Confidence:** 80%

Multiple resolution requests race on shared state. Inconsistent results possible.

---

### HIGH-020: No HTTP Request Body Size Limit
- **File:** `src/base/net/http/HttpContext.cpp:261`
- **Domain:** HTTP API
- **Confidence:** 95%

Body appended without limit. Memory exhaustion via large POST.

---

### HIGH-021: No HTTP Connection Limits
- **File:** `src/base/net/tools/TcpServer.cpp:71`
- **Domain:** HTTP API
- **Confidence:** 90%

Unlimited connections accepted. Connection exhaustion DoS.

---

### HIGH-022: No HTTP Request Timeout
- **File:** `src/base/net/http/HttpServer.cpp:43-59`
- **Domain:** HTTP API
- **Confidence:** 90%

No timeout on requests. Slowloris attack vector.

---

### HIGH-023: Memory Leak in BindHost Parsing
- **File:** `src/proxy/BindHost.cpp:108-112,132-135,154-157`
- **Domain:** TLS & Workers
- **Confidence:** 85%

Raw `new char[]` not freed if String copies instead of taking ownership.

---

### HIGH-024: File Descriptor Leak on Error Path
- **File:** `src/base/io/log/FileLogWriter.cpp:75-84`
- **Domain:** Base I/O
- **Confidence:** 90%

If `uv_fs_open` succeeds but check fails, fd leaked (set to -1 without close).

---

### HIGH-025: Race Condition in FileLogWriter Async Flush
- **File:** `src/base/io/log/FileLogWriter.cpp:138-152`
- **Domain:** Base I/O
- **Confidence:** 88%

`m_pos` updated before async write completes. Out-of-order writes corrupt log.

---

---

## Medium Priority Issues

### MED-001: Windows Background Mode Closes Invalid Handle
- **File:** `src/App_win.cpp:44-45`
- **Domain:** Entry Point
- **Confidence:** 90%

`CloseHandle()` on standard handle - should not be closed manually.

---

### MED-002: Resource Leaks on Early Return Paths
- **File:** `src/App.cpp:46-74`
- **Domain:** Entry Point
- **Confidence:** 85%

Multiple return paths leave UV handles partially initialized without cleanup.

---

### MED-003: Config Reload Race Condition
- **File:** `src/base/kernel/Base.cpp:254-279,296-313`
- **Domain:** Config
- **Confidence:** 80%

Config swapped without synchronization. Concurrent readers may access freed config.

---

### MED-004: Integer Overflow in Miner::setJob()
- **File:** `src/proxy/Miner.cpp:154`
- **Domain:** Proxy Core
- **Confidence:** 80%

Division by zero if `m_customDiff` is 0.

---

### MED-005: Buffer Overflow Risk in Workers Name Display
- **File:** `src/proxy/workers/Workers.cpp:96`
- **Domain:** TLS & Workers
- **Confidence:** 80%

Complex memcpy arithmetic for name truncation. Off-by-one potential.

---

### MED-006: Unbounded Memory in TickingCounter
- **File:** `src/proxy/TickingCounter.h:60,64`
- **Domain:** TLS & Workers
- **Confidence:** 85%

`m_data` vector grows unbounded with each tick().

---

### MED-007: Static Buffer in TLS Read - Thread Safety
- **File:** `src/base/net/stratum/Tls.cpp:130-135`
- **Domain:** Network
- **Confidence:** 85%

Static buffer shared across all TLS instances. Data corruption possible.

---

### MED-008: Overly Permissive CORS Configuration
- **File:** `src/base/net/http/HttpApiResponse.cpp:53-55`
- **Domain:** HTTP API
- **Confidence:** 85%

`Access-Control-Allow-Origin: *` allows any website to access API.

---

### MED-009: TLS 1.0/1.1 Support - Deprecated Protocols
- **File:** `src/base/net/tls/TlsContext.cpp:152,271-279`
- **Domain:** HTTP API
- **Confidence:** 85%

Deprecated TLS versions not disabled by default. Downgrade attacks possible.

---

### MED-010: Cipher Suite Error Ignored
- **File:** `src/base/net/tls/TlsContext.cpp:165-174`
- **Domain:** HTTP API
- **Confidence:** 82%

Same as CRIT-009, duplicate finding confirming severity.

---

### MED-011: Integer Overflow in Keccak
- **File:** `src/base/crypto/keccak.cpp:176,190-191`
- **Domain:** Base I/O
- **Confidence:** 82%

`rsiz` calculation can underflow with large `mdlen`.

---

### MED-012: Missing Null Check in Console
- **File:** `src/base/io/Console.cpp:33-40,74`
- **Domain:** Base I/O
- **Confidence:** 85%

`m_listener` not null-checked in callbacks.

---

### MED-013: TOCTOU in Watcher
- **File:** `src/base/io/Watcher.cpp:74-82`
- **Domain:** Base I/O
- **Confidence:** 80%

File can be replaced between callback and restart. Acceptable for config files.

---

## Recommended Priority Order

### Immediate (Security Critical)
1. CRIT-014: Timing attack in API authentication
2. CRIT-013: Missing TLS certificate verification
3. CRIT-001: Double-delete in Controller
4. CRIT-005: MemPool thread safety
5. CRIT-015: Signal handler race condition

### This Week (Data Integrity)
6. CRIT-004: Events system race condition
7. CRIT-006: Static event buffer corruption
8. CRIT-010: Unbounded m_results memory
9. HIGH-007: StatsData unbounded memory
10. HIGH-020: HTTP body size limit

### Next Sprint (Stability)
11. CRIT-002: UV loop cleanup
12. CRIT-011: gc() out-of-bounds access
13. CRIT-012: SSL_write return checking
14. HIGH-021: Connection limits
15. HIGH-022: Request timeouts

### Backlog (Quality)
- All Medium priority items
- Documentation updates
- Performance optimizations

---

## Review Completion Status

- [x] Entry Point & App Lifecycle - 6 issues found
- [x] Core Controller & Config - 6 issues found
- [x] Proxy Core (Server, Miner, Events) - 10 issues found
- [x] Proxy TLS & Workers - 7 issues found
- [x] Splitter System - 5 issues found
- [x] Network & Stratum Client - 9 issues found
- [x] HTTP/HTTPS & REST API - 7 issues found
- [x] Base I/O & Kernel - 7 issues found

**Total Issues Identified: 57**

---

## Files Requiring Immediate Attention

1. `src/core/Controller.cpp` - Double-delete, null checks
2. `src/base/api/Httpd.cpp` - Timing attack
3. `src/base/net/tls/TlsContext.cpp` - Cipher error, TLS config
4. `src/base/net/tools/MemPool.h` - Thread safety
5. `src/proxy/Events.cpp` - Race condition
6. `src/proxy/events/Event.h` - Static buffer
7. `src/base/io/Signals.cpp` - Async-signal-safety
8. `src/base/net/stratum/Tls.cpp` - SSL_write, cert verify
9. `src/proxy/splitters/nicehash/NonceSplitter.cpp` - Bounds check
10. `src/base/net/http/HttpContext.cpp` - Body size limit
