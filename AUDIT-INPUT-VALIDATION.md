# Security Audit: Input Validation

This document outlines the findings of a security audit focused on input validation and sanitization within the mining application.

## Input Entry Points Inventory

### API Endpoints

The primary entry points for untrusted input are the API handlers defined in `pkg/mining/service.go`. The following handlers process user-controllable data from URL path parameters, query strings, and request bodies:

-   **System & Miner Management:**
    -   `POST /miners/:miner_name/install`: `miner_name` from path.
    -   `DELETE /miners/:miner_name/uninstall`: `miner_name` from path.
    -   `DELETE /miners/:miner_name`: `miner_name` from path.
    -   `POST /miners/:miner_name/stdin`: `miner_name` from path and JSON body (`input`).

-   **Statistics & History:**
    -   `GET /miners/:miner_name/stats`: `miner_name` from path.
    -   `GET /miners/:miner_name/hashrate-history`: `miner_name` from path.
    -   `GET /miners/:miner_name/logs`: `miner_name` from path.
    -   `GET /history/miners/:miner_name`: `miner_name` from path.
    -   `GET /history/miners/:miner_name/hashrate`: `miner_name` from path, `since` and `until` from query.

-   **Profiles:**
    -   `POST /profiles`: JSON body (`MiningProfile`).
    -   `GET /profiles/:id`: `id` from path.
    -   `PUT /profiles/:id`: `id` from path and JSON body (`MiningProfile`).
    -   `DELETE /profiles/:id`: `id` from path.
    -   `POST /profiles/:id/start`: `id` from path.

### WebSocket Events

The WebSocket endpoint provides another significant entry point for untrusted input:

-   **`GET /ws/events`**: Establishes a WebSocket connection. While the primary flow is server-to-client, the initial handshake and any client-to-server messages must be considered untrusted input. The `wsUpgrader` in `pkg/mining/service.go` has an origin check, which is a good security measure.

## Validation Gaps Found

The `Config.Validate()` method in `pkg/mining/mining.go` provides a solid baseline for input validation but has several gaps:

### Strengths

-   **Core Fields Validated**: The most critical fields for command-line construction (`Pool`, `Wallet`, `Algo`, `CLIArgs`) have validation checks.
-   **Denylist for Shell Characters**: The `containsShellChars` function attempts to block a wide range of characters that could be used for shell injection.
-   **Range Checks**: Numeric fields like `Threads`, `Intensity`, and `DonateLevel` are correctly checked to ensure they fall within a sane range.
-   **Allowlist for Algorithm**: The `isValidAlgo` function uses a strict allowlist for the `Algo` field, which is a security best practice.

### Weaknesses and Gaps

-   **Incomplete Field Coverage**: A significant number of fields in the `Config` struct are not validated at all. An attacker could potentially abuse these fields if they are used in command-line arguments or other sensitive operations in the future. Unvalidated fields include:
    -   `Coin`
    -   `Password`
    -   `UserPass`
    -   `Proxy`
    -   `RigID`
    -   `LogFile` (potential for path traversal)
    -   `CPUAffinity`
    -   `Devices`
    -   Many others.

-   **Denylist Approach**: The primary validation mechanism, `containsShellChars`, relies on a denylist of dangerous characters. This approach is inherently brittle because it is impossible to foresee all possible malicious inputs. A determined attacker might find ways to bypass the filter using alternative encodings or unlisted characters. An allowlist approach, accepting only known-good characters, is much safer.

-   **No Path Traversal Protection**: The `LogFile` field is not validated. An attacker could provide a value like `../../../../etc/passwd` to attempt to write files in arbitrary locations on the filesystem.

-   **Inconsistent Numeric Validation**: While some numeric fields are validated, others like `Retries`, `RetryPause`, `CPUPriority`, etc., are not checked for negative values or reasonable upper bounds.

## Injection Vectors Discovered

The primary injection vector discovered is through the `Config.CLIArgs` field, which is used to pass additional command-line arguments to the miner executables.

### XMRig Miner (`pkg/mining/xmrig_start.go`)

-   **Unused in `xmrig_start.go`**: The `addCliArgs` function in `xmrig_start.go` does not actually use the `CLIArgs` field. It constructs arguments from other validated fields. This is good, but the presence of the field in the `Config` struct is misleading and could be used in the future, creating a vulnerability if not handled carefully.

### TT-Miner (`pkg/mining/ttminer_start.go`)

-   **Direct Command Injection via `CLIArgs`**: The `addTTMinerCliArgs` function directly appends the contents of `Config.CLIArgs` to the command-line arguments. Although it uses a denylist-based `isValidCLIArg` function to filter out some dangerous characters, this approach is not foolproof.

    -   **Vulnerability**: An attacker can bypass the filter by crafting a malicious string that is not on the denylist but is still interpreted by the shell. For example, if a new shell feature or a different shell is used on the system, the denylist may become ineffective.

    -   **Example**: While the current filter blocks most common injection techniques, an attacker could still pass arguments that might cause unexpected behavior in the miner, such as `--algo some-exploitable-algo`, if the miner itself has vulnerabilities in how it parses certain arguments.

### Path Traversal in Config File Creation

-   **Vulnerability**: The `getXMRigConfigPath` function in `xmrig.go` uses the `instanceName` to construct a config file path. The `instanceName` is derived from the user-provided `config.Algo`. While the `instanceNameRegex` in `manager.go` sanitizes the algorithm name, it still allows forward slashes (`/`).

-   **Example**: If an attacker provides a crafted `algo` like `../../../../tmp/myconfig`, the `instanceNameRegex` will not sanitize it, and the application could write a config file to an arbitrary location. This could be used to overwrite critical files or place malicious configuration files in sensitive locations.

## Remediation Recommendations

To address the identified vulnerabilities, the following remediation actions are recommended:

### 1. Strengthen `Config.Validate()` with an Allowlist Approach

Instead of relying on a denylist of dangerous characters, the validation should be updated to use a strict allowlist of known-good characters for each field.

**Code Example (`pkg/mining/mining.go`):**
\`\`\`go
// isValidInput checks if a string contains only allowed characters.
// This should be used for fields like Wallet, Password, Pool, etc.
func isValidInput(s string, allowedChars string) bool {
    for _, r := range s {
        if !strings.ContainsRune(allowedChars, r) {
            return false
        }
    }
    return true
}

// In Config.Validate():
func (c *Config) Validate() error {
    // Example for Wallet field
    if c.Wallet != "" {
        // Allow alphanumeric, plus common address characters like '-' and '_'
        allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
        if !isValidInput(c.Wallet, allowedChars) {
            return fmt.Errorf("wallet address contains invalid characters")
        }
    }

    // Apply similar allowlist validation to all other string fields.
    // ...

    return nil
}
\`\`\`

### 2. Sanitize File Paths to Prevent Path Traversal

Sanitize any user-controllable input that is used to construct file paths. The `filepath.Clean` function and checks to ensure the path stays within an expected directory are essential.

**Code Example (`pkg/mining/manager.go`):**
\`\`\`go
import "path/filepath"

// In Manager.StartMiner():
// ...
instanceName := miner.GetName()
if config.Algo != "" {
    // Sanitize algo to prevent directory traversal
    sanitizedAlgo := instanceNameRegex.ReplaceAllString(config.Algo, "_")
    // Also, explicitly remove any path-related characters that the regex might miss
    sanitizedAlgo = strings.ReplaceAll(sanitizedAlgo, "/", "")
    sanitizedAlgo = strings.ReplaceAll(sanitizedAlgo, "..", "")
    instanceName = fmt.Sprintf("%s-%s", instanceName, sanitizedAlgo)
}
// ...
\`\`\`

### 3. Avoid Passing Raw CLI Arguments to `exec.Command`

The `CLIArgs` field is inherently dangerous. If it must be supported, it should be parsed and validated argument by argument, rather than being passed directly to the shell.

**Code Example (`pkg/mining/ttminer_start.go`):**
\`\`\`go
// In addTTMinerCliArgs():
func addTTMinerCliArgs(config *Config, args *[]string) {
    if config.CLIArgs != "" {
        // A safer approach is to define a list of allowed arguments
        allowedArgs := map[string]bool{
            "--list-devices": true,
            "--no-watchdog":  true,
            // Add other safe, non-sensitive arguments here
        }

        extraArgs := strings.Fields(config.CLIArgs)
        for _, arg := range extraArgs {
            if allowedArgs[arg] {
                *args = append(*args, arg)
            } else {
                logging.Warn("skipping potentially unsafe CLI argument", logging.Fields{"arg": arg})
            }
        }
    }
}
\`\`\`

### 4. Expand Validation Coverage in `Config.Validate()`

All fields in the `Config` struct should have some form of validation. For string fields, this should be allowlist-based character validation. For numeric fields, this should be range checking.

**Code Example (`pkg/mining/mining.go`):**
\`\`\`go
// In Config.Validate():
// ...
    // Example for LogFile
    if c.LogFile != "" {
        // Basic validation: ensure it's just a filename, not a path
        if strings.Contains(c.LogFile, "/") || strings.Contains(c.LogFile, "\\") {
            return fmt.Errorf("LogFile cannot be a path")
        }
        // Use an allowlist for the filename itself
        allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."
        if !isValidInput(c.LogFile, allowedChars) {
            return fmt.Errorf("LogFile contains invalid characters")
        }
    }

    // Example for CPUPriority
    if c.CPUPriority < 0 || c.CPUPriority > 5 {
        return fmt.Errorf("CPUPriority must be between 0 and 5")
    }
// ...
\`\`\`
