# Code Review Findings - Miner Proxy Enterprise Audit

**Generated:** 2025-12-31
**Reviewed by:** 8 Parallel Opus Code Reviewers
**Confidence Threshold:** 80%+ (all reported issues)

---

## Review Domains

- [x] Domain 1: Entry Point & App Lifecycle - 5 issues
- [x] Domain 2: Core Controller & Configuration - 5 issues
- [x] Domain 3: Proxy Core (Connection Handling) - 8 issues
- [x] Domain 4: Event System & Statistics - 5 issues
- [x] Domain 5: Splitter System (Nonce Management) - 6 issues
- [x] Domain 6: Stratum Protocol - 9 issues
- [x] Domain 7: HTTP/API Layer - 10 issues
- [x] Domain 8: TLS & Security - 6 issues

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Entry Point & App Lifecycle | 1 | 2 | 2 | 5 |
| Core Controller & Config | 2 | 2 | 1 | 5 |
| Proxy Core | 2 | 4 | 2 | 8 |
| Event System & Stats | 2 | 3 | 0 | 5 |
| Splitter System | 2 | 4 | 0 | 6 |
| Stratum Protocol | 3 | 6 | 0 | 9 |
| HTTP/API Layer | 4 | 4 | 2 | 10 |
| TLS & Security | 2 | 2 | 2 | 6 |
| **TOTAL** | **18** | **27** | **9** | **54** |

---

## Critical Issues

### CRIT-001: Static Shared Send Buffer Race Condition
- **File:** `src/proxy/Miner.cpp:59,146-147`
- **Domain:** Proxy Core
- **Confidence:** 100%

Multiple miner instances share a single static send buffer `m_sendBuf[16384]`. With 100K+ concurrent miners, threads overwrite each other's data, causing corrupted job data to be sent to miners.

**Fix:** Make `m_sendBuf` a per-instance member variable (not static).

---

### CRIT-002: Static Event Buffer Race Condition
- **File:** `src/proxy/events/Event.h:52`
- **Domain:** Proxy Core / Event System
- **Confidence:** 100%

All events share a single 4KB static buffer for placement new. Concurrent events overwrite each other's memory, causing use-after-free and memory corruption under load.

**Fix:** Use heap allocation or thread-local storage for event objects.

---

### CRIT-003: API Exposes Miner Passwords in Clear Text
- **File:** `src/api/v1/ApiRouter.cpp:145,160`
- **Domain:** HTTP/API Layer
- **Confidence:** 100%

The `/1/miners` endpoint exposes miner passwords in API responses without redaction.

**Fix:** Never expose passwords in API responses. Replace with redacted values.

---

### CRIT-004: Cipher Configuration Failure Returns Success
- **File:** `src/base/net/tls/TlsContext.cpp:165-174`
- **Domain:** TLS & Security
- **Confidence:** 100%

`setCiphers()` returns `true` even when cipher configuration fails, silently falling back to weak default ciphers.

**Fix:** Return `false` on failure.

---

### CRIT-005: Double-Free Vulnerability in Controller
- **File:** `src/core/Controller.cpp:45,73-74`
- **Domain:** Entry Point & Configuration
- **Confidence:** 95%

`m_proxy` is deleted in both destructor and `stop()` method. If `stop()` is called before destructor, double-free occurs.

**Fix:** Add null check before deletion or delete only in destructor.

---

### CRIT-006: Missing JSON Type Validation in BindHost
- **File:** `src/proxy/BindHost.cpp:62-73`
- **Domain:** Core Controller
- **Confidence:** 95%

Constructor calls `GetString()`, `GetUint()`, `GetBool()` without checking if members exist or have correct type. Crashes on malformed JSON.

**Fix:** Add `HasMember()` and type checks before accessing JSON values.

---

### CRIT-007: Permissive CORS Configuration
- **File:** `src/base/net/http/HttpApiResponse.cpp:53-55`
- **Domain:** HTTP/API Layer
- **Confidence:** 95%

`Access-Control-Allow-Origin: *` allows any website to make authenticated API requests. Combined with password exposure (CRIT-003), enables remote credential theft.

**Fix:** Remove wildcard CORS or implement origin validation.

---

### CRIT-008: No Certificate Verification on Client Side
- **File:** `src/base/net/stratum/Tls.cpp:38-48`
- **Domain:** TLS & Security
- **Confidence:** 95%

Client-side TLS contexts don't call `SSL_CTX_set_verify()`. Vulnerable to MITM attacks when no fingerprint is configured.

**Fix:** Add `SSL_CTX_set_verify(m_ctx, SSL_VERIFY_PEER, nullptr)`.

---

### CRIT-009: Use-After-Free via Dangling JSON Reference
- **File:** `src/proxy/events/LoginEvent.h:52,62`
- **Domain:** Event System
- **Confidence:** 95%

`LoginEvent` stores reference to stack-allocated `rapidjson::Document` that goes out of scope. Fragile design that breaks with any asynchronous processing.

**Fix:** Store a copy of JSON params instead of a reference.

---

### CRIT-010: Null Pointer Dereference in ExtraNonceSplitter
- **File:** `src/proxy/splitters/extra_nonce/ExtraNonceSplitter.cpp:66,73,79,etc`
- **Domain:** Splitter System
- **Confidence:** 95%

`m_upstream` pointer used without null checks throughout class. Segfault if methods called before `connect()`.

**Fix:** Add null checks before all `m_upstream` usage.

---

### CRIT-011: Buffer Overflow in Client::parseResponse()
- **File:** `src/base/net/stratum/Client.cpp:815`
- **Domain:** Stratum Protocol
- **Confidence:** 90%

Unchecked access to error["message"] without validating it exists or is a string.

**Fix:** Add `HasMember()` and `IsString()` checks.

---

### CRIT-012: Buffer Overflow in Job::setBlob()
- **File:** `src/base/net/stratum/Job.cpp:73-74`
- **Domain:** Stratum Protocol
- **Confidence:** 90%

Buffer size validation can be bypassed if algorithm type changes after check but before copy.

**Fix:** Use compile-time bounds checking.

---

### CRIT-013: Non-Thread-Safe Event System
- **File:** `src/proxy/Events.cpp:39-54`
- **Domain:** Event System
- **Confidence:** 90%

Global boolean `m_ready` flag for reentrancy is not thread-safe. Multiple threads can corrupt event state.

**Fix:** Use `std::atomic<bool>` with compare-and-swap or mutex protection.

---

### CRIT-014: Missing Request Body Size Limits (DoS)
- **File:** `src/base/net/http/HttpContext.cpp:259-264`
- **Domain:** HTTP/API Layer
- **Confidence:** 90%

HTTP body parsing has no size limits, allowing unbounded memory allocation. Attacker can exhaust server memory.

**Fix:** Add `MAX_HTTP_BODY_SIZE` constant (e.g., 1MB) and check before appending.

---

### CRIT-015: Improper libuv Event Loop Cleanup
- **File:** `src/App.cpp:78-79`
- **Domain:** Entry Point
- **Confidence:** 90%

`uv_loop_close()` called without ensuring all handles are properly closed. Resource leaks and potential crashes.

**Fix:** Run loop again to process pending callbacks, use `uv_walk()` for cleanup.

---

### CRIT-016: Use-After-Free in Client::onClose()
- **File:** `src/base/net/stratum/Client.cpp:1003-1011`
- **Domain:** Stratum Protocol
- **Confidence:** 88%

Static callback retrieves client pointer but no guarantee object is still valid during callback.

**Fix:** Use reference counting for Client lifecycle management.

---

### CRIT-017: Header Injection via Unbounded Headers
- **File:** `src/base/net/http/HttpContext.cpp:194-225`
- **Domain:** HTTP/API Layer
- **Confidence:** 85%

HTTP headers accumulated without size limits, allowing memory exhaustion and potential injection.

**Fix:** Add `MAX_HEADER_SIZE` constant (e.g., 8KB) and check.

---

### CRIT-018: Out-of-Bounds Vector Access in NonceSplitter
- **File:** `src/proxy/splitters/nicehash/NonceSplitter.cpp:199,209`
- **Domain:** Splitter System
- **Confidence:** 85%

Direct vector indexing without bounds checking during garbage collection. Race with concurrent access.

**Fix:** Add bounds check: `if (mapperId >= m_upstreams.size()) return;`

---

## High Priority Issues

### HIGH-001: Race Condition in Miner ID Generation
- **File:** `src/proxy/Miner.cpp:58`
- **Domain:** Proxy Core
- **Confidence:** 95%

Static `nextId` incremented without synchronization. Duplicate miner IDs with 100K+ connections.

**Fix:** Use `std::atomic<int64_t>`.

---

### HIGH-002: Private Key Files Without Secure Permissions
- **File:** `src/base/net/tls/TlsGen.cpp:126-149`
- **Domain:** TLS & Security
- **Confidence:** 95%

Private keys written without restrictive permissions. May be world-readable.

**Fix:** Use `umask(0077)` and `chmod(m_certKey, 0600)`.

---

### HIGH-003: Global Counter Race Conditions
- **File:** `src/proxy/Counters.h:42-57`
- **Domain:** Proxy Core / Event System
- **Confidence:** 90%

Global counters modified without synchronization. Incorrect statistics under load.

**Fix:** Use `std::atomic<uint64_t>` for all counters.

---

### HIGH-004: Unsafe strtol Usage Without Error Checking
- **File:** `src/core/config/ConfigTransform.cpp:85`, `src/proxy/BindHost.cpp:136,158`
- **Domain:** Core Controller
- **Confidence:** 90%

`strtol()` used without checking for conversion errors, overflow, or invalid input.

**Fix:** Check `errno == ERANGE`, `endptr`, and value bounds.

---

### HIGH-005: Weak TLS Configuration Allows Deprecated Protocols
- **File:** `src/base/net/tls/TlsContext.cpp:271-293`
- **Domain:** HTTP/API Layer
- **Confidence:** 90%

TLSv1.0 and TLSv1.1 can be enabled. Known vulnerabilities (BEAST, POODLE).

**Fix:** Always disable TLSv1.0 and TLSv1.1 regardless of config.

---

### HIGH-006: No Rate Limiting on Authentication Attempts
- **File:** `src/base/api/Httpd.cpp:136-175`
- **Domain:** HTTP/API Layer
- **Confidence:** 90%

Unlimited brute-force attempts allowed on authentication.

**Fix:** Implement rate limiting per IP address.

---

### HIGH-007: Integer Overflow in ExtraNonceStorage
- **File:** `src/proxy/splitters/extra_nonce/ExtraNonceStorage.cpp:37,99`
- **Domain:** Splitter System
- **Confidence:** 90%

Unbounded increment of `m_extraNonce` (int64_t). Overflow causes nonce collisions.

**Fix:** Add overflow check and wrap-around handling.

---

### HIGH-008: Unsafe Signal Handler Logging
- **File:** `src/base/io/Signals.cpp:61-88`
- **Domain:** Entry Point
- **Confidence:** 85%

Signal handler calls `LOG_WARN()` which may not be async-signal-safe. Potential deadlock.

**Fix:** Use lockless signal-safe logging or defer to event loop.

---

### HIGH-009: Unbounded Memory Growth in Statistics
- **File:** `src/proxy/Stats.cpp:138`, `src/proxy/StatsData.h:96`
- **Domain:** Event System
- **Confidence:** 100%

`m_data.latency` vector grows without bounds. Memory exhaustion over time.

**Fix:** Implement rolling window with maximum size.

---

### HIGH-010: Potential Use-After-Free in CloseEvent
- **File:** `src/proxy/Miner.cpp:555-556,571-572`
- **Domain:** Event System
- **Confidence:** 85%

CloseEvent dispatched with Miner pointer, then immediately deleted. Use-after-free if listener stores pointer.

**Fix:** Document lifetime guarantees or use shared_ptr.

---

### HIGH-011: Memory Leak in BindHost Parsing
- **File:** `src/proxy/BindHost.cpp:108,132,154`
- **Domain:** Core Controller
- **Confidence:** 85%

Raw buffer allocation with `new char[]` assigned to String. Fragile ownership pattern.

**Fix:** Use direct String construction.

---

### HIGH-012: Unvalidated Buffer Access in LineReader
- **File:** `src/base/net/tools/LineReader.cpp:59-61`
- **Domain:** Stratum Protocol
- **Confidence:** 85%

Silent truncation when line exceeds 64KB. No error reported to caller.

**Fix:** Return error status and log the issue.

---

### HIGH-013: Integer Overflow in Job::setTarget()
- **File:** `src/base/net/stratum/Job.cpp:122`
- **Domain:** Stratum Protocol
- **Confidence:** 85%

Division by zero possible if raw target value is 0.

**Fix:** Add validation for zero values.

---

### HIGH-014: Weak Random Number Generation
- **File:** `src/base/tools/Cvt.cpp:67-68,285-296`
- **Domain:** TLS & Security
- **Confidence:** 85%

Uses `std::mt19937` (not cryptographically secure) for key generation when libsodium not available.

**Fix:** Use OpenSSL's `RAND_bytes()` instead.

---

### HIGH-015: Insufficient TLS Certificate Validation
- **File:** `src/base/net/https/HttpsClient.cpp:142-162`
- **Domain:** HTTP/API Layer
- **Confidence:** 85%

Only fingerprint checking, no hostname/chain/expiration validation.

**Fix:** Add proper certificate chain and hostname validation.

---

### HIGH-016: Nonce Space Exhaustion Not Handled
- **File:** `src/proxy/splitters/nicehash/NonceStorage.cpp:45-62`
- **Domain:** Splitter System
- **Confidence:** 85%

When all 256 nonce slots exhausted, creates new upstream indefinitely. No limit on upstream growth.

**Fix:** Add maximum upstream limit and reject miners when full.

---

### HIGH-017: Unbounded Memory Growth in Client Results
- **File:** `src/base/net/stratum/Client.cpp:237,239`
- **Domain:** Stratum Protocol
- **Confidence:** 85%

`m_results` map grows if responses never arrive. DoS via memory exhaustion.

**Fix:** Implement timeout-based cleanup in `tick()`.

---

### HIGH-018: Missing NULL Check in JSON Parsing
- **File:** `src/proxy/Miner.cpp:355`
- **Domain:** Proxy Core
- **Confidence:** 85%

`doc["method"].GetString()` called without checking if field exists. DoS via crash.

**Fix:** Add `HasMember("method")` and `IsString()` checks.

---

### HIGH-019: Race Condition in Client State Machine
- **File:** `src/base/net/stratum/Client.cpp:334-354`
- **Domain:** Stratum Protocol
- **Confidence:** 82%

TOCTOU vulnerability: `m_socket` checked for null then used without synchronization.

**Fix:** Use proper mutex synchronization for state transitions.

---

### HIGH-020: Use-After-Free Risk in SimpleSplitter::tick()
- **File:** `src/proxy/splitters/simple/SimpleSplitter.cpp:103-123`
- **Domain:** Splitter System
- **Confidence:** 80%

Iterating over map while simultaneously deleting entries. Iterator invalidation risk.

**Fix:** Collect keys to delete, then delete in separate loop.

---

### HIGH-021: Missing Null Check in SimpleMapper::submit()
- **File:** `src/proxy/splitters/simple/SimpleSplitter.cpp:242-251`
- **Domain:** Splitter System
- **Confidence:** 80%

`operator[]` on std::map inserts nullptr if key doesn't exist. Silent map corruption.

**Fix:** Use `find()` instead of `operator[]`.

---

### HIGH-022: Missing Bounds Check in EthStratum Height Parsing
- **File:** `src/base/net/stratum/EthStratumClient.cpp:287-300`
- **Domain:** Stratum Protocol
- **Confidence:** 80%

Buffer pointer arithmetic without bounds checking. Out-of-bounds access possible.

**Fix:** Add explicit bounds checking for `p + 2` and `p + 4` offsets.

---

### HIGH-023: Null Pointer Dereference in Client::parse()
- **File:** `src/base/net/stratum/Client.cpp:689-691`
- **Domain:** Stratum Protocol
- **Confidence:** 83%

`id.GetInt64()` called without checking `id.IsInt64()` first.

**Fix:** Add type checking before accessing JSON values.

---

### HIGH-024: Authentication Token Timing Attack
- **File:** `src/base/api/Httpd.cpp:178-198`
- **Domain:** HTTP/API Layer
- **Confidence:** 80%

Bearer token comparison uses `strncmp` which is vulnerable to timing attacks.

**Fix:** Use constant-time comparison function.

---

### HIGH-025: Windows Handle Leak
- **File:** `src/App_win.cpp:44-45`
- **Domain:** Entry Point
- **Confidence:** 80%

`CloseHandle()` called on standard handle from `GetStdHandle()`. Standard handles shouldn't be closed.

**Fix:** Remove `CloseHandle()` call for standard handles.

---

### HIGH-026: Storage Counter Overflow
- **File:** `src/base/net/tools/Storage.h:37-42`
- **Domain:** Proxy Core
- **Confidence:** 80%

Storage counter can overflow after many connections. ID collisions and use-after-free.

**Fix:** Check for overflow or implement ID recycling.

---

### HIGH-027: Unsafe new Miner Check
- **File:** `src/proxy/Server.cpp:89-92`
- **Domain:** Proxy Core
- **Confidence:** 80%

`if (!miner)` check after `new Miner()` is useless - `new` throws on failure.

**Fix:** Use `new (std::nothrow)` or catch `std::bad_alloc`.

---

## Medium Priority Issues

### MED-001: Weak Custom Diff Validation
- **File:** `src/core/config/Config.cpp:160-165`
- **Domain:** Core Controller
- **Confidence:** 85%

Validation uses platform-dependent `INT_MAX` for `uint64_t` parameter.

**Fix:** Use explicit maximum value or `UINT64_MAX`.

---

### MED-002: Missing Security Headers
- **File:** `src/base/net/http/HttpApiResponse.cpp:49-81`
- **Domain:** HTTP/API Layer
- **Confidence:** 85%

Missing `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy`.

**Fix:** Add standard security headers.

---

### MED-003: Missing TLS Hardening Options
- **File:** `src/base/net/tls/TlsContext.cpp:152-161`
- **Domain:** TLS & Security
- **Confidence:** 90%

Missing `SSL_OP_NO_COMPRESSION` (CRIME attack), `SSL_OP_NO_RENEGOTIATION`.

**Fix:** Add all recommended TLS hardening options.

---

### MED-004: Missing Error Checking on fork()
- **File:** `src/App_unix.cpp:44-49`
- **Domain:** Entry Point
- **Confidence:** 85%

`fork()` failure doesn't log error message, making debugging difficult.

**Fix:** Log error with `strerror(errno)`.

---

### MED-005: Potential Alignment Issues in Keccak
- **File:** `src/base/crypto/keccak.cpp:183,196,201`
- **Domain:** TLS & Security
- **Confidence:** 80%

C-style cast of `uint8_t*` to `uint64_t*` without alignment guarantees. UB on ARM/SPARC.

**Fix:** Use `memcpy` for unaligned access or require aligned input.

---

### MED-006: Missing Null Pointer Check in Controller
- **File:** `src/core/Controller.cpp:78-99`
- **Domain:** Core Controller
- **Confidence:** 80%

Methods call `proxy()` without null check. Crash if called before `init()` or after `stop()`.

**Fix:** Add assertions or null checks.

---

### MED-007: Potential URL Length Attack
- **File:** `src/base/net/http/HttpContext.cpp:235-239`
- **Domain:** HTTP/API Layer
- **Confidence:** 80%

URL parsing has no length limits. DoS via memory exhaustion.

**Fix:** Add `MAX_URL_LENGTH` constant (e.g., 8KB).

---

### MED-008: Redundant Map Lookup
- **File:** `src/proxy/splitters/nicehash/NonceMapper.cpp:264-276`
- **Domain:** Splitter System
- **Confidence:** 100%

Triple lookup in `m_results` map: `count()`, `at()`, then `find()`.

**Fix:** Use single `find()` call.

---

### MED-009: Self-Signed Certificate Validity Too Long
- **File:** `src/base/net/tls/TlsGen.cpp:113-115`
- **Domain:** TLS & Security
- **Confidence:** 85%

Self-signed certificates valid for 10 years. Industry best practice is 1 year max.

**Fix:** Reduce to 1 year (31536000 seconds).

---

## Recommended Priority Order

### Immediate (Security Critical - Fix Before Production)
1. **CRIT-003**: API Exposes Miner Passwords
2. **CRIT-007**: Permissive CORS Configuration (enables CRIT-003 exploitation)
3. **CRIT-004**: Cipher Configuration Failure Returns Success
4. **CRIT-008**: No Certificate Verification on Client Side
5. **CRIT-001/002**: Static Buffer Race Conditions (data corruption)

### This Week (Data Integrity & Stability)
6. **CRIT-005**: Double-Free in Controller
7. **CRIT-006**: Missing JSON Type Validation
8. **HIGH-001**: Race Condition in Miner ID Generation
9. **HIGH-003**: Global Counter Race Conditions
10. **CRIT-014**: Missing Request Body Size Limits

### Next Sprint (Reliability & Hardening)
11. **HIGH-009**: Unbounded Memory Growth in Statistics
12. **CRIT-010**: Null Pointer Dereference in ExtraNonceSplitter
13. **CRIT-018**: Out-of-Bounds Vector Access
14. **HIGH-004**: Unsafe strtol Usage
15. **HIGH-006**: No Rate Limiting on Auth

### Backlog (Quality & Defense in Depth)
- All remaining HIGH issues
- All MEDIUM issues
- Performance optimizations (MED-008)

---

## Review Completion Status

- [x] Domain 1: Entry Point & App Lifecycle - 5 issues found
- [x] Domain 2: Core Controller & Configuration - 5 issues found
- [x] Domain 3: Proxy Core (Connection Handling) - 8 issues found
- [x] Domain 4: Event System & Statistics - 5 issues found
- [x] Domain 5: Splitter System (Nonce Management) - 6 issues found
- [x] Domain 6: Stratum Protocol - 9 issues found
- [x] Domain 7: HTTP/API Layer - 10 issues found
- [x] Domain 8: TLS & Security - 6 issues found

**Total Issues Identified: 54**
- Critical: 18
- High: 27
- Medium: 9

---

## Key Patterns Identified

1. **Thread Safety Violations**: The codebase assumes single-threaded execution but handles 100K+ concurrent connections. Static shared buffers, non-atomic counters, and missing synchronization throughout.

2. **Missing Input Validation**: JSON parsing lacks type/existence checks. Size limits missing on HTTP bodies, headers, URLs.

3. **Memory Lifecycle Issues**: Use-after-free risks in event system, double-free in Controller, dangling references to stack objects.

4. **Security Configuration Failures**: Weak TLS defaults, permissive CORS, exposed credentials, missing certificate validation.

5. **Resource Exhaustion**: Unbounded memory growth in statistics, results maps, and upstream creation without limits.
