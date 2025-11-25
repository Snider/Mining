# Mining Architecture Guide

This document provides an overview of the architecture of the Mining project.

## High-Level Overview

The project is structured as a modular Go application. It consists of:
1.  **Core Library (`pkg/mining`)**: Contains the business logic for miner management.
2.  **CLI (`cmd/mining`)**: A command-line interface built with Cobra.
3.  **REST API**: A Gin-based web server exposed via the `serve` command.
4.  **Frontend**: An Angular-based dashboard for monitoring miners.

## Core Components

### Manager Interface
The core of the system is the `ManagerInterface` defined in `pkg/mining/manager_interface.go`. This interface abstracts the operations of starting, stopping, and monitoring miners.

```go
type ManagerInterface interface {
    StartMiner(minerType string, config *Config) (Miner, error)
    StopMiner(name string) error
    GetMiner(name string) (Miner, error)
    ListMiners() []Miner
    ListAvailableMiners() []AvailableMiner
    GetMinerHashrateHistory(name string) ([]HashratePoint, error)
    Stop()
}
```

This abstraction allows for:
- Easier testing (mocking the manager).
- Pluggable implementations (e.g., supporting different miner backends).

### Miner Interface
Each specific miner (e.g., XMRig) implements the `Miner` interface. This interface defines how to interact with the underlying miner executable process, parse its output, and control it.

The `XMRigMiner` implementation (`pkg/mining/xmrig.go`) handles:
- Downloading and verifying the miner binary.
- Generating configuration files.
- Executing the binary.
- Parsing stdout/stderr/API for stats.

### Service Layer
The `Service` struct (`pkg/mining/service.go`) wraps the `Manager` and exposes its functionality via a HTTP API using the Gin framework. It handles:
- Route registration.
- Request validation/binding.
- Response formatting.
- Swagger documentation generation.

## Directory Structure

- `cmd/mining/`: Entry point for the application.
- `pkg/mining/`: Core library code.
- `docs/`: Documentation and Swagger files.
- `ui/`: Source code for the Angular frontend.
- `dist/`: Build artifacts (binaries).

## Data Flow

1.  **User Action**: User issues a command via CLI or calls an API endpoint.
2.  **Service/CMD Layer**: The request is validated and passed to the `Manager`.
3.  **Manager Layer**: The manager looks up the appropriate `Miner` implementation.
4.  **Miner Layer**: The miner instance interacts with the OS (filesystem, processes).
5.  **Feedback**: Status and stats are returned up the stack to the user.
