# Parallel Code Review with Claude Code

A reproducible pattern for running multiple Opus code reviewers in parallel across different domains of a codebase.

---

## Overview

This technique spawns 6-10 specialized code review agents simultaneously, each focused on a specific domain. Results are consolidated into a single TODO.md with prioritized findings.

**Best for:**
- Large C/C++/Go/Rust codebases
- Security audits
- Pre-release quality gates
- Technical debt assessment

---

## Step 1: Define Review Domains

Analyze your codebase structure and identify 6-10 logical domains. Each domain should be:
- Self-contained enough for independent review
- Small enough to review thoroughly (5-20 key files)
- Aligned with architectural boundaries

### Example Domain Breakdown (C++ Miner)

```
1. Entry Point & App Lifecycle     -> src/App.cpp, src/xmrig.cpp
2. Core Controller & Miner         -> src/core/
3. CPU Backend                      -> src/backend/cpu/, src/backend/common/
4. GPU Backends                     -> src/backend/opencl/, src/backend/cuda/
5. Crypto Algorithms                -> src/crypto/
6. Network & Stratum                -> src/base/net/stratum/, src/net/
7. HTTP REST API                    -> src/base/api/, src/base/net/http/
8. Hardware Access                  -> src/hw/, src/base/kernel/
```

---

## Step 2: Create Output File

Create a skeleton TODO.md to track progress:

```markdown
# Code Review Findings - [Project Name]

Generated: [DATE]

## Review Domains

- [ ] Domain 1
- [ ] Domain 2
...

## Critical Issues
_Pending review..._

## High Priority Issues
_Pending review..._

## Medium Priority Issues
_Pending review..._
```

---

## Step 3: Launch Parallel Reviewers

Use this prompt template for each domain. Launch ALL domains simultaneously in a single message with multiple Task tool calls.

### Reviewer Prompt Template

```
You are reviewing the [LANGUAGE] [PROJECT] for enterprise quality. Focus on:

**Domain: [DOMAIN NAME]**
- `path/to/file1.cpp` - description
- `path/to/file2.cpp` - description
- `path/to/directory/` - description

Look for:
1. Memory leaks, resource management issues
2. Thread safety and race conditions
3. Error handling gaps
4. Null pointer dereferences
5. Security vulnerabilities
6. Input validation issues

Report your findings in a structured format with:
- File path and line number
- Issue severity (CRITICAL/HIGH/MEDIUM/LOW)
- Confidence percentage (only report issues with 80%+ confidence)
- Description of the problem
- Suggested fix

Work from: /absolute/path/to/project
```

### Launch Command Pattern

```
Use Task tool with:
- subagent_type: "feature-dev:code-reviewer"
- run_in_background: true
- description: "Review [Domain Name]"
- prompt: [Template above filled in]

Launch ALL domains in ONE message to run in parallel.
```

---

## Step 4: Collect Results

After launching, wait for all agents to complete:

```
Use TaskOutput tool with:
- task_id: [agent_id from launch]
- block: true
- timeout: 120000
```

Collect all results in parallel once agents start completing.

---

## Step 5: Consolidate Findings

Structure the final TODO.md with this format:

```markdown
# Code Review Findings - [Project] Enterprise Audit

**Generated:** YYYY-MM-DD
**Reviewed by:** N Parallel Opus Code Reviewers

---

## Summary

| Domain | Critical | High | Medium | Total |
|--------|----------|------|--------|-------|
| Domain 1 | X | Y | Z | N |
| Domain 2 | X | Y | Z | N |
| **TOTAL** | **X** | **Y** | **Z** | **N** |

---

## Critical Issues

### CRIT-001: [Short Title]
- **File:** `path/to/file.cpp:LINE`
- **Domain:** [Domain Name]
- **Confidence:** XX%

[Description of the issue]

**Fix:** [Suggested fix]

---

[Repeat for each critical issue]

## High Priority Issues

### HIGH-001: [Short Title]
- **File:** `path/to/file.cpp:LINE`
- **Domain:** [Domain Name]
- **Confidence:** XX%

[Description]

---

## Medium Priority Issues

[Same format]

---

## Recommended Priority Order

### Immediate (Security Critical)
1. CRIT-XXX: [title]
2. CRIT-XXX: [title]

### This Week (Data Integrity)
3. CRIT-XXX: [title]
4. HIGH-XXX: [title]

### Next Sprint (Stability)
5. HIGH-XXX: [title]

### Backlog (Quality)
- MED-XXX items

---

## Review Completion Status

- [x] Domain 1 - N issues found
- [x] Domain 2 - N issues found
- [ ] Domain 3 - Review incomplete

**Total Issues Identified: N**
```

---

## Domain-Specific Prompts

### For C/C++ Projects

```
Look for:
1. Memory leaks, resource management issues (RAII violations)
2. Buffer overflows, bounds checking
3. Thread safety and race conditions
4. Use-after-free, double-free
5. Null pointer dereferences
6. Integer overflow/underflow
7. Format string vulnerabilities
8. Uninitialized variables
```

### For Go Projects

```
Look for:
1. Goroutine leaks
2. Race conditions (run with -race)
3. Nil pointer dereferences
4. Error handling gaps (ignored errors)
5. Context cancellation issues
6. Channel deadlocks
7. Slice/map concurrent access
8. Resource cleanup (defer patterns)
```

### For Network/API Code

```
Look for:
1. Buffer overflows in protocol parsing
2. TLS/SSL configuration issues
3. Input validation vulnerabilities
4. Authentication/authorization gaps
5. Timing attacks in comparisons
6. Connection/request limits (DoS)
7. CORS misconfigurations
8. Information disclosure
```

### For Crypto Code

```
Look for:
1. Side-channel vulnerabilities
2. Weak random number generation
3. Key/secret exposure in logs
4. Timing attacks
5. Buffer overflows in crypto ops
6. Integer overflow in calculations
7. Proper constant-time operations
8. Key lifecycle management
```

---

## Tips for Best Results

1. **Be specific about file paths** - Give reviewers exact paths to focus on
2. **Set confidence threshold** - Only report 80%+ confidence issues
3. **Include context** - Mention the project type, language, and any special patterns
4. **Limit scope** - 5-20 files per domain is ideal
5. **Run in parallel** - Launch all agents in one message for efficiency
6. **Use background mode** - `run_in_background: true` allows parallel execution
7. **Consolidate immediately** - Write findings while context is fresh

---

## Example Invocation

```
"Spin up Opus code reviewers to analyze this codebase for enterprise quality.
Create a TODO.md with findings organized by severity."
```

This triggers:
1. Domain identification from project structure
2. Parallel agent launch (6-10 reviewers)
3. Result collection
4. Consolidated TODO.md generation

---

## Metrics

Typical results for a medium-sized project (50-100k LOC):

- **Time:** 3-5 minutes for full parallel review
- **Issues found:** 30-60 total
- **Critical:** 5-15 issues
- **High:** 15-25 issues
- **False positive rate:** ~10-15% (filtered by confidence threshold)