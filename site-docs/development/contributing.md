# Contributing

Contributions are welcome! Here's how to get involved.

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 20+
- Make

### Clone and Build

```bash
git clone https://github.com/yourusername/Mining.git
cd Mining

# Build backend
make build

# Build frontend
cd ui && npm install && ng build
```

## Code Style

### Go

- Run `make lint` before committing
- Follow standard Go conventions
- Use meaningful variable names
- Add comments for exported functions

### TypeScript/Angular

- Use standalone components
- Follow Angular style guide
- Use TypeScript strict mode

## Testing

### Backend Tests

```bash
make test                       # All tests
go test -v ./pkg/mining/...     # Specific package
go test -run TestName ./...     # Single test
```

### E2E Tests

```bash
cd ui
npm run e2e                     # All E2E tests
npm run e2e:api                 # API tests only
npm run e2e:ui                  # Interactive mode
```

## Pull Request Guidelines

1. **One feature per PR** - Keep changes focused
2. **Write tests** - Add tests for new functionality
3. **Update docs** - Update relevant documentation
4. **Describe changes** - Clear PR description
5. **Pass CI** - All tests must pass

## Adding a New Miner

To add support for a new miner:

1. Create `pkg/mining/newminer.go`
2. Implement the `Miner` interface
3. Register in `manager.go`
4. Add UI support if needed
5. Write tests
6. Document the miner

Example structure:

```go
type NewMiner struct {
    *BaseMiner
    // miner-specific fields
}

func NewNewMiner() *NewMiner {
    return &NewMiner{
        BaseMiner: NewBaseMiner("newminer", "newminer"),
    }
}

func (m *NewMiner) Start(cfg *Config) error {
    // Implementation
}

func (m *NewMiner) GetStats() (*PerformanceMetrics, error) {
    // Implementation
}
```

## Reporting Issues

When reporting bugs:

1. Check existing issues first
2. Include system information
3. Provide steps to reproduce
4. Include relevant logs
5. Attach screenshots if UI-related

## License

By contributing, you agree that your contributions will be licensed under the project's license.
