# Memory and Resource Management Audit

This audit examines the application's memory and resource management based on a review of the codebase, with a focus on `pkg/mining/manager.go`, `pkg/mining/service.go`, and `pkg/database/database.go`.

## 1. Goroutine Leak Analysis

The application uses several long-running goroutines for background tasks. Overall, goroutine lifecycle management is robust, but there are minor areas for improvement.

### Findings:

- **Stats Collection (`manager.go`):** The `startStatsCollection` goroutine runs in a `for` loop with a `time.Ticker`. It reliably terminates when the `stopChan` is closed during `Manager.Stop()`.
- **Database Cleanup (`manager.go`):** The `startDBCleanup` goroutine also uses a `time.Ticker` and correctly listens for the `stopChan` signal, ensuring it exits cleanly.
- **WebSocket Event Hub (`service.go`):** The `EventHub.Run` method is launched as a goroutine and manages client connections. It terminates when its internal `quit` channel is closed, which is triggered by the `EventHub.Stop()` method.

### Recommendations:

- **No major issues found.** The use of `stopChan` and `sync.WaitGroup` in `Manager` provides a solid foundation for graceful shutdowns.

## 2. Memory Leak Analysis

The primary areas of concern for memory leaks are in-memory data structures that could grow indefinitely.

### Findings:

- **`Manager.miners` Map:** The `miners` map in the `Manager` struct stores active miner processes. Entries are added in `StartMiner` and removed in `StopMiner` and `UninstallMiner`. If a miner process were to crash or become unresponsive without `StopMiner` being called, its entry would persist in the map, causing a minor memory leak.
- **In-Memory Hashrate History:** Each miner maintains an in-memory `HashrateHistory`. The `ReduceHashrateHistory` method is called periodically to trim this data, preventing unbounded growth. This is a good practice.
- **Request Body Size Limit:** The `service.go` file correctly implements a 1MB request body size limit, which helps prevent memory exhaustion from large API requests.

### Recommendations:

- **Implement a health check for miners.** A periodic health check could detect unresponsive miner processes and trigger their removal from the `miners` map, preventing memory leaks from orphaned entries.

## 3. Database Resource Management

The application uses an SQLite database for persisting historical data.

### Findings:

- **Connection Pooling:** The `database.go` file configures the connection pool with `SetMaxOpenConns(1)`. This is appropriate for SQLite's single-writer model and prevents connection-related issues.
- **`hashrate_history` Cleanup:** The `Cleanup` function in `database.go` correctly removes old records from the `hashrate_history` table based on the configured retention period.
- **`miner_sessions` Table:** The `miner_sessions` table tracks miner uptime but has no corresponding cleanup mechanism. This table will grow indefinitely, leading to a gradual increase in database size and a potential performance degradation over time.

### Recommendations:

- **Add a cleanup mechanism for `miner_sessions`.** Extend the `Cleanup` function to also remove old records from the `miner_sessions` table based on the retention period.

## 4. File Handle and Process Management

The application manages external miner processes, which requires careful handling of file descriptors and process handles.

### Findings:

- **Process Lifecycle:** The `Stop` method on miner implementations (`xmrig.go`, `ttminer.go`) is responsible for terminating the `exec.Cmd` process. This appears to be handled correctly.
- **I/O Pipes:** The miner's `stdout`, `stderr`, and `stdin` pipes are created and managed. The code does not show any obvious leaks of these file handles.

### Recommendations:

- **No major issues found.** The process management logic appears to be sound.

## 5. Network Connection Handling

The application's API server and WebSocket endpoint are critical areas for resource management.

### Findings:

- **HTTP Server Timeouts:** The `service.go` file correctly configures `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` for the HTTP server, which is a best practice for preventing slow client attacks and connection exhaustion.
- **WebSocket Connections:** The `wsUpgrader` has a `CheckOrigin` function that restricts connections to `localhost` origins, providing a layer of security. The `EventHub` manages the lifecycle of WebSocket connections.

### Recommendations:

- **No major issues found.** The network connection handling is well-configured.
