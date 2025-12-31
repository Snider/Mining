# Mining Development Guide

This guide is for developers contributing to the Mining project.

## Prerequisites

- **Go**: Version 1.24 or higher.
- **Make**: For running build scripts.
- **Node.js/npm**: For building the frontend (optional).

## Common Tasks

The project uses a `Makefile` to automate common tasks.

### Simulation Mode

For UI development without real mining hardware, use the simulation mode:

```bash
# Start with 3 simulated CPU miners
miner-ctrl simulate --count 3 --preset cpu-high

# Custom hashrate and algorithm
miner-ctrl simulate --count 2 --hashrate 8000 --algorithm rx/0

# Available presets: cpu-low, cpu-medium, cpu-high, gpu-ethash, gpu-kawpow
```

This generates realistic hashrate data with variance, share events, and pool connections for testing the UI.

### Building

Build the CLI binary for the current platform:
```bash
make build
```

Build for all supported platforms (cross-compile):
```bash
make build-all
```

The binaries will be placed in the `dist/` directory.

### Testing

Run all Go tests:
```bash
make test
```

Run tests with race detection and coverage:
```bash
make test-release
```

Generate and view HTML coverage report:
```bash
make coverage
```

### Linting & Formatting

Format code:
```bash
make fmt
```

Run linters (requires `golangci-lint`):
```bash
make lint
```

### Documentation

Generate Swagger documentation from code annotations:
```bash
make docs
```
(Requires `swag` tool: `make install-swag`)

### Release

The project uses GoReleaser for releases.
To create a local snapshot release:
```bash
make package
```

## Project Structure

- **`pkg/mining`**: This is where the core logic resides. If you are adding a new feature, you will likely work here.
- **`cmd/mining`**: If you are adding a new CLI command, look here.
- **`ui`**: Frontend code.

## Contribution Workflow

1.  Fork the repository.
2.  Create a feature branch.
3.  Make your changes.
4.  Ensure tests pass (`make test`).
5.  Submit a Pull Request.

## CodeRabbit

This project uses CodeRabbit for automated code reviews. Please address any feedback provided by the bot on your PR.
