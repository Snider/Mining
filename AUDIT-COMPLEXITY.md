# Code Complexity and Maintainability Audit

This document analyzes the code quality of the codebase, identifies maintainability issues, and provides recommendations for improvement. The audit focuses on cyclomatic and cognitive complexity, code duplication, and other maintainability metrics.

## 1. God Class: `Manager`

### Finding
The `Manager` struct in `pkg/mining/manager.go` is a "God Class" that violates the Single Responsibility Principle. It handles multiple, unrelated responsibilities, including:
- Miner lifecycle management (`StartMiner`, `StopMiner`)
- Configuration management (`syncMinersConfig`, `updateMinerConfig`)
- Database interactions (`initDatabase`, `startDBCleanup`)
- Statistics collection (`startStatsCollection`, `collectMinerStats`)

This centralization of concerns makes the `Manager` class difficult to understand, test, and maintain. The presence of multiple mutexes (`mu`, `eventHubMu`) to prevent deadlocks is a clear indicator of its high cognitive complexity.

### Recommendation
Refactor the `Manager` class into smaller, more focused components, each with a single responsibility.

- **`MinerRegistry`**: Manages the lifecycle of miner instances.
- **`StatsCollector`**: Gathers and aggregates statistics from miners.
- **`ConfigService`**: Handles loading, saving, and updating miner configurations.
- **`DBManager`**: Manages all database-related operations.

This separation of concerns will improve modularity, reduce complexity, and make the system easier to reason about and test.

## 2. Code Duplication: Miner Installation

### Finding
The `Install` and `CheckInstallation` methods in `pkg/mining/xmrig.go` and `pkg/mining/ttminer.go` contain nearly identical logic for downloading, extracting, and verifying miner installations. This copy-paste pattern violates the DRY (Don't Repeat Yourself) principle and creates a significant maintenance burden. Any change to the installation process must be manually duplicated across all miner implementations.

### Recommendation
Refactor the duplicated logic into the `BaseMiner` struct using the **Template Method Pattern**. The base struct will define the skeleton of the installation algorithm, while subclasses will override specific steps (like providing the download URL format) that vary between miners.

#### Example
The `BaseMiner` can provide a generic `Install` method that relies on a new, unexported method, `getDownloadURL`, which each miner implementation must provide.

**`pkg/mining/miner.go` (BaseMiner)**
```go
// Install orchestrates the download and extraction process.
func (b *BaseMiner) Install() error {
	version, err := b.GetLatestVersion()
	if err != nil {
		return err
	}
	b.Version = version

	url, err := b.getDownloadURL(version)
	if err != nil {
		return err
	}

	return b.InstallFromURL(url)
}

// getDownloadURL is a template method to be implemented by subclasses.
func (b *BaseMiner) getDownloadURL(version string) (string, error) {
	// This will be overridden by specific miner types
	return "", errors.New("getDownloadURL not implemented")
}
```

**`pkg/mining/xmrig.go` (XMRigMiner)**
```go
// getDownloadURL implements the template method for XMRig.
func (m *XMRigMiner) getDownloadURL(version string) (string, error) {
	v := strings.TrimPrefix(version, "v")
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("https://.../xmrig-%s-windows-x64.zip", v), nil
	case "linux":
		return fmt.Sprintf("https://.../xmrig-%s-linux-static-x64.tar.gz", v), nil
	default:
		return "", errors.New("unsupported OS")
	}
}
```

## 3. Long and Complex Methods

### Finding
Several methods in the codebase are overly long and have high cognitive complexity, making them difficult to read, understand, and maintain.

- **`manager.StartMiner`**: This method is responsible for creating, configuring, and starting a miner. It mixes validation, port finding, instance name generation, and state management, making it hard to follow.
- **`manager.collectMinerStats`**: This function orchestrates the parallel collection of stats, but the logic for handling timeouts, retries, and database persistence is deeply nested.
- **`miner.ReduceHashrateHistory`**: The logic for aggregating high-resolution hashrate data into a low-resolution format is convoluted and hard to reason about.

### Recommendation
Apply the **Extract Method** refactoring to break down these long methods into smaller, well-named functions, each with a single, clear purpose.

#### Example: Refactoring `manager.StartMiner`
The `StartMiner` method could be refactored into several smaller helper functions.

**`pkg/mining/manager.go` (Original `StartMiner`)**
```go
func (m *Manager) StartMiner(ctx context.Context, minerType string, config *Config) (Miner, error) {
    // ... (20+ lines of setup, validation, port finding)

    // ... (10+ lines of miner-specific configuration)

    // ... (10+ lines of starting and saving logic)
}
```

**`pkg/mining/manager.go` (Refactored `StartMiner`)**
```go
func (m *Manager) StartMiner(ctx context.Context, minerType string, config *Config) (Miner, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }

    instanceName, err := m.generateInstanceName(minerType, config)
    if err != nil {
        return nil, err
    }

    miner, err := m.configureMiner(minerType, instanceName, config)
    if err != nil {
        return nil, err
    }

    if err := m.launchAndRegisterMiner(miner, config); err != nil {
        return nil, err
    }

    return miner, nil
}
```
