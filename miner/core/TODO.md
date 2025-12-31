# Code Review Findings - C++ Miner Core Enterprise Audit

**Generated:** 2025-12-31
**Reviewed by:** 8 Parallel Opus Code Reviewers
**Target:** XMRig-based C++ Miner Core

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Entry Point & App Lifecycle | 2 | 4 | 0 | 6 |
| Core Controller & Miner | 2 | 4 | 0 | 6 |
| CPU Backend | 3 | 3 | 3 | 9 |
| GPU Backends (OpenCL/CUDA) | 1 | 5 | 1 | 7 |
| Network & Stratum | 3 | 4 | 0 | 7 |
| HTTP REST API | 3 | 3 | 2 | 8 |
| **TOTAL** | **14** | **23** | **6** | **43** |

---

## Critical Issues

### CRIT-001: UV Event Loop Resource Leak on Failure
- **File:** `src/App.cpp:89-90`
- **Domain:** Entry Point & App Lifecycle
- **Confidence:** 95%

`uv_loop_close()` is called unconditionally without checking if handles are still active. Returns `UV_EBUSY` and leaks file descriptors, memory on every shutdown.

**Fix:** Add `uv_walk()` to close remaining handles before `uv_loop_close()`.

---

### CRIT-002: Process Object Lifetime Issue After Fork
- **File:** `src/xmrig.cpp:28,34` + `src/App_unix.cpp:42`
- **Domain:** Entry Point & App Lifecycle
- **Confidence:** 90%

`Process` object created on stack, pointer passed to `App`. After `fork()`, both parent and child have references to same memory with different meanings. Potential double-free/undefined behavior.

**Fix:** Use `unique_ptr` for clear ownership or ensure parent blocks until child stabilizes.

---

### CRIT-003: Use-After-Free in Controller::stop()
- **File:** `src/core/Controller.cpp:75-83`
- **Domain:** Core Controller
- **Confidence:** 90%

Network destroyed before Miner stopped. Miner may submit job results to destroyed Network.

```cpp
void Controller::stop() {
    m_network.reset();  // WRONG: Network gone first
    m_miner->stop();    // Accesses network!
}
```

**Fix:** Stop miner first, then destroy network.

---

### CRIT-004: Null Pointer Dereference in Controller::execCommand()
- **File:** `src/core/Controller.cpp:102-106`
- **Domain:** Core Controller
- **Confidence:** 95%

`miner()` and `network()` use assertions disabled in release builds. `m_miner` only initialized in `start()`, not `init()`. Early `execCommand()` calls crash.

**Fix:** Add runtime null checks before dereferencing.

---

### CRIT-005: Race Condition in NUMAMemoryPool::getOrCreate()
- **File:** `src/crypto/common/NUMAMemoryPool.cpp:88-97`
- **Domain:** CPU Backend
- **Confidence:** 95%

Check-then-act race: multiple threads can check `m_map.count(node)`, all see missing, all create new `MemoryPool` instances. Memory leaks + corruption of `std::map`.

**Fix:** Add mutex protection around entire check-insert operation.

---

### CRIT-006: Race Condition in MemoryPool::get()
- **File:** `src/crypto/common/MemoryPool.cpp:70-84`
- **Domain:** CPU Backend
- **Confidence:** 92%

`m_offset` and `m_refs` modified without synchronization. Multiple workers can receive overlapping memory regions.

**Fix:** Add mutex or use atomic operations.

---

### CRIT-007: cn_heavyZen3Memory Global Memory Leak
- **File:** `src/backend/cpu/CpuWorker.cpp:64,91-96,120-124`
- **Domain:** CPU Backend
- **Confidence:** 88%

Global `cn_heavyZen3Memory` allocated once, never freed. Algorithm changes leave gigabytes of huge pages allocated.

**Fix:** Add reference counting or smart pointer for shared CN_HEAVY memory.

---

### CRIT-008: Memory Leak on OpenCL Program Build Failure
- **File:** `src/backend/opencl/OclCache.cpp:51-54`
- **Domain:** GPU Backends
- **Confidence:** 95%

When `buildProgram()` fails after `createProgramWithSource()` succeeds, `cl_program` is released. But if `createProgramWithSource` returns nullptr edge case, subsequent code dereferences nullptr.

**Fix:** Add null check after createProgramWithSource.

---

### CRIT-009: Buffer Overflow Silent Truncation in LineReader
- **File:** `src/base/net/tools/LineReader.cpp:57-71`
- **Domain:** Network & Stratum
- **Confidence:** 90%

When data exceeds 64KB buffer, silently drops data. Leads to protocol desync, missed commands, DoS.

**Fix:** Log error and close connection on overflow.

---

### CRIT-010: Missing Null Check in Client::parseResponse()
- **File:** `src/base/net/stratum/Client.cpp:814-815`
- **Domain:** Network & Stratum
- **Confidence:** 85%

`error["message"].GetString()` called without checking if field exists. Potential segfault.

**Fix:** Use `Json::getString()` safe getter with fallback.

---

### CRIT-011: Race Condition in Client Socket Cleanup
- **File:** `src/base/net/stratum/Client.cpp:643-659`
- **Domain:** Network & Stratum
- **Confidence:** 82%

`delete m_socket` in `onClose()` while network callback may still be executing. Use-after-free.

**Fix:** Call `uv_read_stop()` before deleting socket.

---

### CRIT-012: Timing Attack in Token Authentication
- **File:** `src/base/api/Httpd.cpp:197`
- **Domain:** HTTP API
- **Confidence:** 100%

Uses `strncmp()` for token comparison - vulnerable to timing attacks. Attacker can extract API token character by character.

**Fix:** Use constant-time comparison function.

---

### CRIT-013: Overly Permissive CORS Configuration
- **File:** `src/base/net/http/HttpApiResponse.cpp:53-55`
- **Domain:** HTTP API
- **Confidence:** 95%

`Access-Control-Allow-Origin: *` allows any website to control miner via CSRF attacks.

**Fix:** Restrict CORS to trusted origins or remove entirely (localhost doesn't need CORS).

---

### CRIT-014: No HTTP Request Body Size Limit
- **File:** `src/base/net/http/HttpContext.cpp:261`
- **Domain:** HTTP API
- **Confidence:** 90%

HTTP body appended indefinitely. Memory exhaustion DoS via multi-gigabyte POST requests.

**Fix:** Add `MAX_BODY_SIZE` (e.g., 1MB) check in `on_body` callback.

---

## High Priority Issues

### HIGH-001: Static Log Destruction After UV Loop Close
- **File:** `src/App.cpp:123-130`
- **Domain:** Entry Point
- **Confidence:** 85%

`Log::destroy()` called while UV loop still running. Pending callbacks may access destroyed logging system.

---

### HIGH-002: No Error Handling on Fork Failure
- **File:** `src/App_unix.cpp:42-46`
- **Domain:** Entry Point
- **Confidence:** 88%

`fork()` failure returns silently with `rc = 1`. No error message logged.

---

### HIGH-003: Controller Stop Order Wrong
- **File:** `src/core/Controller.cpp:75-82`
- **Domain:** Core Controller
- **Confidence:** 82%

Shutdown order should be: stop miner -> destroy miner -> destroy network.

---

### HIGH-004: Missing Null Check in Console Read Callback
- **File:** `src/base/io/Console.cpp:67-76`
- **Domain:** Entry Point
- **Confidence:** 80%

`stream->data` cast to `Console*` without null check. Use-after-free during shutdown.

---

### HIGH-005: Data Race on Global Mutex in Miner
- **File:** `src/core/Miner.cpp:73,465,572,607`
- **Domain:** Core Controller
- **Confidence:** 85%

Global static mutex protects instance data. Multiple Miner instances would incorrectly share lock. Manual lock/unlock not exception-safe.

---

### HIGH-006: Missing VirtualMemory::init() Error Handling
- **File:** `src/core/Controller.cpp:48-62`
- **Domain:** Core Controller
- **Confidence:** 85%

Huge page allocation failure not detected. Leads to degraded performance or crashes when workers use uninitialized memory pool.

---

### HIGH-007: Memory Leak in Base::reload() on Exception
- **File:** `src/base/kernel/Base.cpp:254-279`
- **Domain:** Core Controller
- **Confidence:** 90%

Raw pointer `new Config()` leaked if `config->save()` throws exception.

---

### HIGH-008: Hashrate::addData() Race Condition
- **File:** `src/backend/common/Hashrate.cpp:185-199`
- **Domain:** CPU Backend
- **Confidence:** 90%

`m_top[index]` read-modify-write is not atomic. Torn reads, incorrect hashrate, potential OOB access.

---

### HIGH-009: CpuBackend Global Mutex for Instance Data
- **File:** `src/backend/cpu/CpuBackend.cpp:64,167-170`
- **Domain:** CPU Backend
- **Confidence:** 85%

Global `static std::mutex` shared across all CpuBackend instances. False contention.

---

### HIGH-010: Missing Alignment Check in CnCtx Creation
- **File:** `src/backend/cpu/CpuWorker.cpp:532-546`
- **Domain:** CPU Backend
- **Confidence:** 82%

No verification that `m_memory->scratchpad() + shift` is properly aligned. No bounds checking.

---

### HIGH-011: Exception Safety Issues in OclKawPowRunner::build()
- **File:** `src/backend/opencl/runners/OclKawPowRunner.cpp:193-198`
- **Domain:** GPU Backends
- **Confidence:** 90%

If `KawPow_CalculateDAGKernel` constructor throws, `m_program` from base class leaks.

---

### HIGH-012: Race Condition in OclSharedData::createBuffer()
- **File:** `src/backend/opencl/runners/tools/OclSharedData.cpp:36-55`
- **Domain:** GPU Backends
- **Confidence:** 85%

Window between checking `!m_buffer` and calling `OclLib::retain()` where buffer could be released.

---

### HIGH-013: Missing Null Check in OclRxJitRunner Kernel Creation
- **File:** `src/backend/opencl/runners/OclRxJitRunner.cpp:66-74`
- **Domain:** GPU Backends
- **Confidence:** 88%

If `loadAsmProgram()` throws after kernel allocation, `m_randomx_jit` leaks.

---

### HIGH-014: Potential Double-Free in OclSharedData::release()
- **File:** `src/backend/opencl/runners/tools/OclSharedData.cpp:133-140`
- **Domain:** GPU Backends
- **Confidence:** 82%

Buffers released without setting to nullptr. Double `release()` call could cause issues.

---

### HIGH-015: Unvalidated DAG Buffer Capacity
- **File:** `src/backend/opencl/runners/OclKawPowRunner.cpp:124-130`
- **Domain:** GPU Backends
- **Confidence:** 83%

No device free memory check before potentially multi-GB allocation.

---

### HIGH-016: TLS Certificate Timing Attack
- **File:** `src/base/net/stratum/Tls.cpp:169-187`
- **Domain:** Network & Stratum
- **Confidence:** 84%

Fingerprint comparison uses `strncasecmp` - should use constant-time comparison.

---

### HIGH-017: SOCKS5 Buffer Size Validation Missing
- **File:** `src/base/net/stratum/Socks5.cpp:29-48`
- **Domain:** Network & Stratum
- **Confidence:** 80%

Accesses `data[0]` and `data[1]` without confirming buffer contains at least 2 bytes.

---

### HIGH-018: Missing TLS Read Loop Timeout
- **File:** `src/base/net/stratum/Tls.cpp:130-136`
- **Domain:** Network & Stratum
- **Confidence:** 82%

`while(SSL_read())` loop unbounded. Static buffer not thread-safe.

---

### HIGH-019: Integer Overflow in Job Target Calculation
- **File:** `src/base/net/stratum/Job.cpp:113-132`
- **Domain:** Network & Stratum
- **Confidence:** 81%

Division by zero if target raw data is all zeros.

---

### HIGH-020: No HTTP Connection Limit
- **Files:** `src/base/net/http/HttpServer.cpp:43-59`, `HttpContext.cpp:74-94`
- **Domain:** HTTP API
- **Confidence:** 85%

Unlimited concurrent connections. Memory exhaustion DoS.

---

### HIGH-021: Missing HTTP Request Timeout
- **File:** `src/base/net/http/HttpContext.cpp`
- **Domain:** HTTP API
- **Confidence:** 90%

No timeout on request processing. Slowloris attack vector.

---

### HIGH-022: Restricted Mode Bypass Review Needed
- **File:** `src/base/api/Httpd.cpp:164-172`
- **Domain:** HTTP API
- **Confidence:** 85%

Restricted mode only blocks non-GET. Some GET endpoints may expose sensitive info.

---

### HIGH-023: Unchecked Backend Pointer in Miner::onTimer()
- **File:** `src/core/Miner.cpp:661-666`
- **Domain:** Core Controller
- **Confidence:** 80%

Backend validity not checked before `printHealth()` call during iteration.

---

## Medium Priority Issues

### MED-001: Workers::stop() Thread Join Order
- **File:** `src/backend/common/Workers.cpp:132-149`
- **Domain:** CPU Backend
- **Confidence:** 80%

Workers signaled to stop, then immediately deleted. No guarantee workers exited mining loops.

---

### MED-002: Thread::~Thread() macOS Resource Leak
- **File:** `src/backend/common/Thread.h:50,68`
- **Domain:** CPU Backend
- **Confidence:** 80%

macOS `pthread_join()` return value not checked.

---

### MED-003: Excessive Hashrate Polling Overhead
- **File:** `src/backend/common/Workers.cpp:79-114`
- **Domain:** CPU Backend
- **Confidence:** 85%

64+ virtual calls per tick for high thread counts. Consider batching.

---

### MED-004: Missing Error Handling in OclKawPowRunner::init()
- **File:** `src/backend/opencl/runners/OclKawPowRunner.cpp:201-207`
- **Domain:** GPU Backends
- **Confidence:** 85%

Partial initialization state left if `createCommandQueue()` or `createBuffer()` throws.

---

### MED-005: Information Disclosure via Error Messages
- **File:** `src/base/api/requests/HttpApiRequest.cpp:120-122`
- **Domain:** HTTP API
- **Confidence:** 80%

Detailed JSON parsing errors returned to clients.

---

### MED-006: Header Injection Potential
- **File:** `src/base/net/http/HttpContext.cpp:281-288`
- **Domain:** HTTP API
- **Confidence:** 80%

No explicit validation of header values for CRLF injection or length limits.

---

## Recommended Priority Order

### Immediate (Security Critical)
1. CRIT-012: Timing attack in API token auth
2. CRIT-013: CORS misconfiguration
3. CRIT-009: Buffer overflow silent truncation
4. CRIT-014: No body size limit

### This Week (Data Integrity)
5. CRIT-005: NUMAMemoryPool race condition
6. CRIT-006: MemoryPool race condition
7. CRIT-003: Controller::stop() order
8. HIGH-005: Miner mutex issues

### Next Sprint (Stability)
9. CRIT-001: UV loop resource leak
10. CRIT-002: Process lifetime after fork
11. HIGH-007: Base::reload() memory leak
12. HIGH-011-015: GPU resource management

### Backlog (Quality)
- All Medium priority items
- Performance optimizations
- Additional logging/monitoring

---

## Review Completion Status

- [x] Entry Point & App Lifecycle - 6 issues found
- [x] Core Controller & Miner - 6 issues found
- [x] CPU Backend - 9 issues found
- [x] GPU Backends (OpenCL/CUDA) - 7 issues found
- [x] Network & Stratum - 7 issues found
- [x] HTTP REST API - 8 issues found
- [ ] Crypto Algorithms - Review incomplete (agent scope confusion)
- [ ] Hardware Access - Review incomplete (agent scope confusion)

**Total Issues Identified: 43**