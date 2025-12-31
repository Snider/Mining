# Code Review Findings - Mining Project

**Generated:** 2025-12-31
**Reviewed by:** 4 Parallel Code Reviewers (2 Opus, 2 Sonnet)

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Core Mining (pkg/mining/) | 0 | 2 | 3 | 5 |
| P2P Networking (pkg/node/) | 1 | 0 | 0 | 1 |
| CLI Commands (cmd/mining/) | 3 | 3 | 2 | 8 |
| Angular Frontend (ui/src/app/) | 1 | 1 | 1 | 3 |
| **TOTAL** | **5** | **6** | **6** | **17** |

---

## Critical Issues

### CRIT-001: Path Traversal in Tar Extraction (Zip Slip)
- **File:** `pkg/node/bundle.go:268`
- **Domain:** P2P Networking
- **Confidence:** 95%

The `extractTarball` function uses `filepath.Join(destDir, hdr.Name)` without validating the path stays within destination. Malicious tar archives can write files anywhere on the filesystem.

**Attack Vector:** A remote peer could craft a malicious miner bundle with path traversal entries like `../../../etc/cron.d/malicious`.

**Fix:**
```go
cleanName := filepath.Clean(hdr.Name)
if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
    return "", fmt.Errorf("invalid tar entry: %s", hdr.Name)
}
path := filepath.Join(destDir, cleanName)
if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(destDir)+string(os.PathSeparator)) {
    return "", fmt.Errorf("path escape attempt: %s", hdr.Name)
}
```

---

### CRIT-002: XSS Vulnerability in Console ANSI-to-HTML
- **File:** `ui/src/app/pages/console/console.component.ts:501-549`
- **Domain:** Angular Frontend
- **Confidence:** 95%

The `ansiToHtml()` method bypasses Angular XSS protection using `bypassSecurityTrustHtml()` while constructing HTML with inline styles from ANSI escape sequences. Malicious log output could inject scripts.

**Fix:** Use CSS classes instead of inline styles, validate ANSI codes against whitelist.

---

### CRIT-003: Resource Leak in `node serve` Command
- **File:** `cmd/mining/cmd/node.go:114-161`
- **Domain:** CLI Commands
- **Confidence:** 95%

The `nodeServeCmd` uses `select {}` to block forever without signal handling. Transport connections and goroutines leak on Ctrl+C.

**Fix:** Add signal handling and call `transport.Stop()` on shutdown.

---

### CRIT-004: Path Traversal in `doctor` Command
- **File:** `cmd/mining/cmd/doctor.go:49-55`
- **Domain:** CLI Commands
- **Confidence:** 90%

Reads arbitrary files via manipulated signpost file (`~/.installed-miners`).

**Fix:** Validate that `configPath` is within expected directories using `filepath.Clean()` and prefix checking.

---

### CRIT-005: Path Traversal in `update` Command
- **File:** `cmd/mining/cmd/update.go:33-39`
- **Domain:** CLI Commands
- **Confidence:** 90%

Same vulnerability as CRIT-004.

---

## High Priority Issues

### HIGH-001: Race Condition in `requestTimeoutMiddleware`
- **File:** `pkg/mining/service.go:313-350`
- **Domain:** Core Mining
- **Confidence:** 85%

Goroutine calls `c.Next()` while timeout handler may also write to response. Gin's Context is not thread-safe for concurrent writes.

**Fix:** Use mutex or atomic flag to coordinate response writing.

---

### HIGH-002: Missing Rollback in `UpdateProfile`
- **File:** `pkg/mining/profile_manager.go:123-133`
- **Domain:** Core Mining
- **Confidence:** 82%

If `saveProfiles()` fails, in-memory state is already modified. Unlike `CreateProfile`, `UpdateProfile` has no rollback logic.

**Fix:** Store old profile before update, restore on save failure.

---

### HIGH-003: Type Confusion in `update` Command
- **File:** `cmd/mining/cmd/update.go:44-47`
- **Domain:** CLI Commands
- **Confidence:** 85%

Unmarshals cache as `[]*mining.InstallationDetails` but `doctor` command saves as `mining.SystemInfo`.

**Fix:** Use consistent types between commands.

---

### HIGH-004: Missing Cleanup in `serve` Command
- **File:** `cmd/mining/cmd/serve.go:31-173`
- **Domain:** CLI Commands
- **Confidence:** 85%

No explicit `manager.Stop()` call on shutdown. Relies on implicit service cleanup.

---

### HIGH-005: Scanner Error Not Checked
- **File:** `cmd/mining/cmd/serve.go:72-162`
- **Domain:** CLI Commands
- **Confidence:** 80%

Interactive shell never calls `scanner.Err()` after loop exits.

---

### HIGH-006: Hardcoded HTTP URLs Without TLS
- **Files:** `ui/src/app/miner.service.ts:49`, `node.service.ts:66`, `websocket.service.ts:53`
- **Domain:** Angular Frontend
- **Confidence:** 90%

All API endpoints use `http://localhost` without TLS. Traffic can be intercepted.

**Fix:** Use environment-based config with HTTPS/WSS support.

---

## Medium Priority Issues

### MED-001: Missing `rand.Read` Error Check
- **File:** `pkg/mining/auth.go:209-212`
- **Domain:** Core Mining
- **Confidence:** 88%

`generateNonce()` ignores error from `rand.Read`. Could produce weak nonces.

---

### MED-002: Metrics Race in WebSocket Connection
- **File:** `pkg/mining/service.go:1369-1373`
- **Domain:** Core Mining
- **Confidence:** 80%

`RecordWSConnection(true)` called before connection is accepted. Brief incorrect metrics on rejection.

---

### MED-003: Config Validation Not Called for Profiles
- **File:** `pkg/mining/service.go:978-998`
- **Domain:** Core Mining
- **Confidence:** 82%

`handleStartMinerWithProfile` doesn't call `config.Validate()` after unmarshaling.

---

### MED-004: Weak File Permissions
- **File:** `cmd/mining/cmd/doctor.go:106,115`
- **Domain:** CLI Commands
- **Confidence:** 80%

Cache files created with 0644 (world-readable). Should be 0600.

---

### MED-005: Duplicated Partial ID Matching
- **File:** `cmd/mining/cmd/peer.go:124-131`
- **Domain:** CLI Commands
- **Confidence:** 80%

Partial peer ID matching duplicated across commands. Extract to helper function.

---

### MED-006: innerHTML for Sidebar Icons
- **File:** `ui/src/app/components/sidebar/sidebar.component.ts:64`
- **Domain:** Angular Frontend
- **Confidence:** 85%

Uses `bypassSecurityTrustHtml()` for icons. Currently safe (hardcoded), but fragile.

---

## Review Completion Status

- [x] Domain 1: Core Mining (pkg/mining/) - 5 issues found
- [x] Domain 2: P2P Networking (pkg/node/) - 1 critical issue found
- [x] Domain 3: CLI Commands (cmd/mining/) - 8 issues found
- [x] Domain 4: Angular Frontend (ui/src/app/) - 3 issues found

**Total Issues Identified: 17**

---

## Recommended Priority Order

### Immediate (Security Critical)
1. **CRIT-001:** Path traversal in tar extraction - Remote code execution risk
2. **CRIT-002:** XSS vulnerability in console - Script injection risk
3. **CRIT-003:** Resource leak in node serve - Service stability
4. **CRIT-004/005:** Path traversal in CLI - Arbitrary file read

### This Week (Data Integrity)
5. **HIGH-001:** Race condition in timeout middleware
6. **HIGH-002:** Missing rollback in UpdateProfile
7. **HIGH-003:** Type confusion in update command
8. **HIGH-006:** Hardcoded HTTP URLs

### Next Sprint (Stability)
9. **HIGH-004/005:** Missing cleanup and scanner error checks
10. **MED-001:** rand.Read error check
11. **MED-003:** Config validation for profiles

### Backlog (Quality)
- MED-002, MED-004, MED-005, MED-006

---

## Positive Observations

The codebase demonstrates good practices:
- Proper mutex usage for concurrent access
- `sync.Once` for safe shutdown patterns
- Rate limiting in P2P transport
- Challenge-response auth with constant-time comparison
- Message size limits and deduplication
- Context cancellation handling
- No dynamic code execution or localStorage usage in frontend
