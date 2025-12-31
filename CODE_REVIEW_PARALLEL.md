# Code Review Findings - Mining Project Enterprise Audit

**Generated:** 2025-12-31
**Reviewed by:** 4 Parallel Code Reviewers (2 Opus, 2 Sonnet)

---

## Review Domains

- [x] Domain 1: Core Mining (`pkg/mining/`) - Opus
- [x] Domain 2: P2P Networking (`pkg/node/`) - Opus
- [x] Domain 3: CLI Commands (`cmd/`) - Sonnet
- [x] Domain 4: Angular Frontend (`ui/`) - Sonnet

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Core Mining | 0 | 3 | 2 | 5 |
| P2P Networking | 2 | 3 | 0 | 5 |
| CLI Commands | 2 | 2 | 0 | 4 |
| Angular Frontend | 2 | 3 | 0 | 5 |
| **TOTAL** | **6** | **11** | **2** | **19** |

---

## Critical Issues

### CRIT-001: Panic from Short Public Key in peer.go
- **File:** `pkg/node/peer.go:159,167`
- **Domain:** P2P Networking
- **Confidence:** 95%

The `AllowPublicKey` and `RevokePublicKey` functions access `publicKey[:16]` for logging without validating length. An attacker providing a short public key will cause a panic.

```go
logging.Debug("public key added to allowlist", logging.Fields{"key": publicKey[:16] + "..."})
```

**Fix:** Add length check before string slicing:
```go
keyPreview := publicKey
if len(publicKey) > 16 {
    keyPreview = publicKey[:16] + "..."
}
```

---

### CRIT-002: Panic from Short Public Key in transport.go
- **File:** `pkg/node/transport.go:470`
- **Domain:** P2P Networking
- **Confidence:** 95%

During handshake rejection logging, `payload.Identity.PublicKey[:16]` is accessed without length validation. Malicious peers can crash the transport.

**Fix:** Use same safe string prefix function as CRIT-001.

---

### CRIT-003: Race Condition on Global Variables in node.go
- **File:** `cmd/mining/cmd/node.go:14-17,236-258`
- **Domain:** CLI Commands
- **Confidence:** 95%

Global variables `nodeManager` and `peerRegistry` are initialized with a check-then-act pattern without synchronization, causing race conditions.

```go
func getNodeManager() (*node.NodeManager, error) {
    if nodeManager == nil {  // RACE
        nodeManager, err = node.NewNodeManager()  // Multiple initializations possible
    }
    return nodeManager, nil
}
```

**Fix:** Use `sync.Once` for thread-safe lazy initialization:
```go
var nodeManagerOnce sync.Once
func getNodeManager() (*node.NodeManager, error) {
    nodeManagerOnce.Do(func() {
        nodeManager, nodeManagerErr = node.NewNodeManager()
    })
    return nodeManager, nodeManagerErr
}
```

---

### CRIT-004: Race Condition on Global Variables in remote.go
- **File:** `cmd/mining/cmd/remote.go:12-15,323-351`
- **Domain:** CLI Commands
- **Confidence:** 95%

Same check-then-act race condition on `controller` and `transport` global variables.

**Fix:** Use `sync.Once` pattern.

---

### CRIT-005: XSS via bypassSecurityTrustHtml in Console
- **File:** `ui/src/app/pages/console/console.component.ts:534-575`
- **Domain:** Angular Frontend
- **Confidence:** 85%

The `ansiToHtml()` method uses `DomSanitizer.bypassSecurityTrustHtml()` to render ANSI-formatted log output. A compromised miner or pool could inject malicious payloads.

**Fix:** Remove `bypassSecurityTrustHtml()`, use property binding with pre-sanitized class names, or use a security-audited ANSI library.

---

### CRIT-006: Missing Input Validation on HTTP Endpoints
- **File:** `ui/src/app/miner.service.ts:352-356`, `ui/src/app/node.service.ts:220-247`
- **Domain:** Angular Frontend
- **Confidence:** 90%

Multiple HTTP requests pass user-controlled data directly to backend without client-side validation, exposing to command injection via `sendStdin()`, path traversal via `minerName`, and SSRF via peer addresses.

**Fix:** Add validation for `minerName` (whitelist alphanumeric + hyphens), sanitize `input` in `sendStdin()`, validate peer addresses format.

---

## High Priority Issues

### HIGH-001: TTMiner Goroutine Leak
- **File:** `pkg/mining/ttminer_start.go:75-108`
- **Domain:** Core Mining
- **Confidence:** 85%

In TTMiner `Start()`, the inner goroutine that calls `cmd.Wait()` can leak if process kill timeout occurs but Wait() never returns.

**Fix:** Add secondary timeout for inner goroutine like XMRig implementation.

---

### HIGH-002: Request Timeout Middleware Race
- **File:** `pkg/mining/service.go:339-357`
- **Domain:** Core Mining
- **Confidence:** 82%

The `requestTimeoutMiddleware` spawns a goroutine that continues running after timeout, potentially writing to aborted response.

**Fix:** Use request context cancellation or document handlers must check `c.IsAborted()`.

---

### HIGH-003: Peer Registry AllowPublicKey Index Panic
- **File:** `pkg/node/peer.go:159,167`
- **Domain:** Core Mining
- **Confidence:** 88%

Same issue as CRIT-001 (duplicate finding from different reviewer).

---

### HIGH-004: Unbounded Tar File Extraction
- **File:** `pkg/node/bundle.go:314`
- **Domain:** P2P Networking
- **Confidence:** 85%

`extractTarball` uses `io.Copy(f, tr)` without limiting file size, allowing decompression bombs.

**Fix:**
```go
const maxFileSize = 100 * 1024 * 1024
limitedReader := io.LimitReader(tr, min(hdr.Size, maxFileSize))
io.Copy(f, limitedReader)
```

---

### HIGH-005: Unvalidated Lines Parameter (DoS)
- **File:** `pkg/node/worker.go:266-276`
- **Domain:** P2P Networking
- **Confidence:** 82%

`handleGetLogs` passes `Lines` parameter without validation, allowing memory exhaustion.

**Fix:** Add validation: `if payload.Lines > 10000 { payload.Lines = 10000 }`

---

### HIGH-006: Missing TLS Configuration Hardening
- **File:** `pkg/node/transport.go:206-216`
- **Domain:** P2P Networking
- **Confidence:** 80%

TLS uses default configuration without minimum version or cipher suite restrictions.

**Fix:** Add TLS config with `MinVersion: tls.VersionTLS12` and restricted cipher suites.

---

### HIGH-007: Missing Input Validation on Pool/Wallet
- **File:** `cmd/mining/cmd/serve.go:95-112`
- **Domain:** CLI Commands
- **Confidence:** 85%

Interactive shell accepts pool/wallet without format validation.

**Fix:** Validate pool URL prefix (stratum+tcp:// or stratum+ssl://), length limits.

---

### HIGH-008: Incomplete Signal Handling
- **File:** `cmd/mining/cmd/node.go:162-176`
- **Domain:** CLI Commands
- **Confidence:** 82%

Missing SIGHUP handling, no force cleanup if Stop() fails.

**Fix:** Add SIGHUP to signal handling, implement forced cleanup on Stop() failure.

---

### HIGH-009: Insecure WebSocket Message Handling
- **File:** `ui/src/app/websocket.service.ts:155-168`
- **Domain:** Angular Frontend
- **Confidence:** 82%

WebSocket messages parsed without validation or type guards.

**Fix:** Validate message structure, implement type guards, validate event types against whitelist.

---

### HIGH-010: Memory Leaks from Unsubscribed Observables
- **File:** `ui/src/app/pages/profiles/profiles.component.ts`, `workers.component.ts`
- **Domain:** Angular Frontend
- **Confidence:** 85%

Components subscribe to observables in event handlers without proper cleanup.

**Fix:** Use `takeUntil(destroy$)` pattern, implement `OnDestroy`.

---

### HIGH-011: Error Information Disclosure
- **File:** `ui/src/app/pages/profiles/profiles.component.ts:590-593`, `setup-wizard.component.ts:43-52`
- **Domain:** Angular Frontend
- **Confidence:** 80%

Error handlers display detailed error messages exposing internal API structure.

**Fix:** Create generic error messages, log details only in dev mode.

---

## Medium Priority Issues

### MED-001: Profile Manager DeleteProfile Missing Rollback
- **File:** `pkg/mining/profile_manager.go:146-156`
- **Domain:** Core Mining
- **Confidence:** 80%

If `saveProfiles()` fails after deletion, in-memory and on-disk state become inconsistent.

**Fix:** Store reference to deleted profile and restore on save failure.

---

### MED-002: Config Validation Missing for CLIArgs
- **File:** `pkg/mining/mining.go:162-213`
- **Domain:** Core Mining
- **Confidence:** 83%

`Config.Validate()` doesn't validate `CLIArgs` field for shell characters.

**Fix:** Add CLIArgs validation in Config.Validate().

---

## Recommended Priority Order

### Immediate (Crash Prevention)
1. CRIT-001: Panic from short public key in peer.go
2. CRIT-002: Panic from short public key in transport.go
3. CRIT-003: Race condition in node.go
4. CRIT-004: Race condition in remote.go

### This Week (Security Critical)
5. CRIT-005: XSS via bypassSecurityTrustHtml
6. CRIT-006: Missing input validation
7. HIGH-004: Unbounded tar extraction
8. HIGH-006: Missing TLS hardening

### Next Sprint (Stability)
9. HIGH-001: TTMiner goroutine leak
10. HIGH-002: Timeout middleware race
11. HIGH-005: Unvalidated Lines parameter
12. HIGH-007: Pool/wallet validation
13. HIGH-008: Signal handling
14. HIGH-009: WebSocket validation
15. HIGH-010: Memory leaks
16. HIGH-011: Error disclosure

### Backlog (Quality)
17. MED-001: Profile manager rollback
18. MED-002: CLIArgs validation

---

## Positive Findings (Good Practices)

The codebase demonstrates several enterprise-quality patterns:

**Core Mining:**
- Proper mutex usage with separate read/write locks
- Panic recovery in goroutines
- Graceful shutdown with `sync.Once`
- Atomic writes for file operations
- Input validation with shell character blocking

**P2P Networking:**
- Constant-time comparison with `hmac.Equal`
- Path traversal protection in tar extraction
- Symlinks/hard links blocked
- Message deduplication
- Per-peer rate limiting

**CLI Commands:**
- Proper argument separation (no shell execution)
- Path validation in doctor.go
- Instance name sanitization with regex

**Angular Frontend:**
- No dynamic code execution patterns
- No localStorage/sessionStorage usage
- Type-safe HTTP client
- ShadowDOM encapsulation

---

## Review Completion Status

- [x] Core Mining (`pkg/mining/`) - 5 issues found
- [x] P2P Networking (`pkg/node/`) - 5 issues found
- [x] CLI Commands (`cmd/`) - 4 issues found
- [x] Angular Frontend (`ui/`) - 5 issues found

**Total Issues Identified: 19**
