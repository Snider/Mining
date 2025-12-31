# Development Guide

Welcome to the Mining Platform development guide. This documentation will help you set up your development environment and contribute to the project.

## Prerequisites

### Required Tools

- **Go**: Version 1.24 or higher
- **Node.js**: Version 20 or higher (for UI development)
- **npm**: Version 10 or higher
- **Make**: For build automation
- **Git**: For version control

### Optional Tools

- **CMake**: Version 3.21+ (for building miner core with GPU support)
- **OpenCL SDK**: For AMD GPU development
- **CUDA Toolkit**: For NVIDIA GPU development
- **golangci-lint**: For code linting
- **swag**: For generating Swagger documentation

### Install Development Tools

**Go Tools:**
```bash
# Install swag for Swagger generation
go install github.com/swaggo/swag/cmd/swag@latest

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Node.js Tools:**
```bash
cd ui
npm install
```

## Getting the Source Code

Clone the repository:

```bash
git clone https://github.com/Snider/Mining.git
cd Mining
```

## Project Structure

```
Mining/
├── cmd/
│   ├── mining/              # CLI application
│   │   ├── main.go          # Entry point
│   │   └── cmd/             # Cobra commands
│   └── desktop/             # Desktop application
│       └── mining-desktop/  # Wails app
├── pkg/mining/              # Core Go package
│   ├── mining.go            # Interfaces and types
│   ├── manager.go           # Miner lifecycle management
│   ├── service.go           # REST API (Gin)
│   ├── xmrig.go             # XMRig implementation
│   ├── xmrig_start.go       # XMRig startup logic
│   ├── xmrig_stats.go       # XMRig statistics parsing
│   ├── profile_manager.go   # Profile persistence
│   └── config_manager.go    # Config management
├── miner/core/              # Modified XMRig
│   └── src/
│       ├── backend/         # Mining backends
│       │   ├── opencl/      # OpenCL (AMD/NVIDIA)
│       │   └── cuda/        # CUDA (NVIDIA)
│       └── crypto/          # Algorithm implementations
│           ├── etchash/     # Ethereum Classic
│           └── progpowz/    # Zano
├── ui/                      # Angular web dashboard
│   ├── src/
│   │   ├── app/
│   │   │   ├── components/  # Reusable components
│   │   │   ├── pages/       # Route pages
│   │   │   └── services/    # API services
│   │   └── environments/    # Environment configs
│   ├── e2e/                 # Playwright E2E tests
│   └── package.json
├── docs/                    # Documentation
├── Makefile                 # Build automation
└── README.md
```

## Building the Project

### Backend (Go)

Build the CLI binary:

```bash
make build
```

The binary will be created as `miner-ctrl` in the current directory.

For cross-platform builds:

```bash
make build-all
```

Binaries will be in `dist/` directory for Linux, macOS, and Windows.

### Frontend (Angular)

Build the web dashboard:

```bash
cd ui
npm install
npm run build
```

Output will be in `ui/dist/browser/` as `mbe-mining-dashboard.js`.

For development with hot reload:

```bash
cd ui
npm run start
```

This starts a development server on `http://localhost:4200`.

### Desktop Application

Build the Wails desktop app:

```bash
cd cmd/desktop/mining-desktop
npm install
wails3 build
```

Binary will be in `cmd/desktop/mining-desktop/bin/`.

For development mode with hot reload:

```bash
cd cmd/desktop/mining-desktop
wails3 dev
```

### Miner Core (with GPU support)

Build the modified XMRig with GPU support:

```bash
cd miner/core
mkdir build && cd build

# Configure with OpenCL and CUDA
cmake .. -DWITH_OPENCL=ON -DWITH_CUDA=ON

# Build
make -j$(nproc)
```

Binary will be in `miner/core/build/xmrig`.

## Running Tests

### Go Tests

Run all tests:

```bash
make test
```

Run with race detection and coverage:

```bash
make test-release
```

Generate coverage report:

```bash
make coverage
```

Opens an HTML coverage report in your browser.

Run specific tests:

```bash
go test -v ./pkg/mining/... -run TestName
```

### Frontend Tests

Run Angular unit tests:

```bash
cd ui
npm test
```

This runs Karma/Jasmine tests (36 specs).

Run E2E tests with Playwright:

```bash
cd ui
npm run e2e
```

Or run specific test suites:

```bash
# API tests only (no browser)
make e2e-api

# UI tests only
make e2e-ui

# Interactive UI mode
make e2e
```

## Code Quality

### Linting

Format Go code:

```bash
make fmt
```

Run Go linters:

```bash
make lint
```

Format TypeScript/Angular code:

```bash
cd ui
npm run lint
npm run lint:fix
```

### Code Style

**Go:**
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions focused and small
- Write descriptive variable names
- Add comments for exported functions

**TypeScript/Angular:**
- Follow [Angular Style Guide](https://angular.io/guide/styleguide)
- Use TypeScript strict mode
- Prefer composition over inheritance
- Write unit tests for components and services

## Generating Documentation

### Swagger API Docs

Generate Swagger documentation:

```bash
make docs
```

This runs `swag init` and updates `docs/swagger.json` and `docs/swagger.yaml`.

Swagger annotations are in `pkg/mining/service.go`.

### mkdocs Documentation

This documentation is built with mkdocs. To preview locally:

```bash
# Install mkdocs
pip install mkdocs mkdocs-material

# Serve docs locally
mkdocs serve
```

Open `http://127.0.0.1:8000` to view the documentation.

## Development Workflow

### Starting Development Server

Run the full development stack:

```bash
# Terminal 1: Start Go backend
make dev

# Terminal 2: Start Angular dev server
cd ui && npm run start

# Terminal 3: Watch for changes
make watch
```

This provides:
- Backend API on `http://localhost:9090`
- Frontend dev server on `http://localhost:4200`
- Auto-reload on file changes

### Making Changes

1. **Create a branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**
   - Edit source files
   - Add tests
   - Update documentation

3. **Test your changes:**
   ```bash
   make test
   cd ui && npm test
   ```

4. **Lint your code:**
   ```bash
   make lint
   cd ui && npm run lint
   ```

5. **Commit:**
   ```bash
   git add .
   git commit -m "feat: Add my feature"
   ```

6. **Push and create PR:**
   ```bash
   git push origin feature/my-feature
   ```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(api): Add endpoint for profile management
fix(miner): Fix XMRig hashrate calculation
docs(readme): Update installation instructions
```

## Debugging

### Go Backend

Debug with Delve:

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugging
dlv debug ./cmd/mining -- serve --port 9090
```

Or use your IDE's debugger (VS Code, GoLand, etc.).

### Angular Frontend

Debug in browser:

1. Start dev server: `npm run start`
2. Open browser DevTools (F12)
3. Use Sources tab for breakpoints
4. Console for logs and errors

### Desktop App

Debug Wails app:

```bash
cd cmd/desktop/mining-desktop
wails3 dev --devtools
```

This opens the app with Chrome DevTools enabled.

## Testing Guidelines

### Writing Go Tests

```go
func TestStartMiner(t *testing.T) {
    // Arrange
    manager := NewManager()
    config := &Config{
        Pool: "stratum+tcp://pool.test:3333",
        Wallet: "test_wallet",
    }

    // Act
    miner, err := manager.StartMiner("xmrig", config)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, miner)
    assert.Equal(t, "xmrig", miner.GetName())
}
```

### Writing Angular Tests

```typescript
describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [DashboardComponent],
    });
    fixture = TestBed.createComponent(DashboardComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should fetch miners on init', () => {
    spyOn(component.minerService, 'getMiners').and.returnValue(of([]));
    component.ngOnInit();
    expect(component.minerService.getMiners).toHaveBeenCalled();
  });
});
```

## Continuous Integration

The project uses GitHub Actions for CI/CD:

- **Build**: Builds for all platforms on every push
- **Test**: Runs all tests on every PR
- **Lint**: Checks code quality
- **E2E**: Runs Playwright tests
- **Release**: Creates releases on tags

CI configuration is in `.github/workflows/`.

## Common Development Tasks

### Adding a New API Endpoint

1. Add route in `pkg/mining/service.go`:
   ```go
   router.GET("/my-endpoint", s.handleMyEndpoint)
   ```

2. Implement handler:
   ```go
   // @Summary My endpoint
   // @Description Description of what it does
   // @Tags miners
   // @Produce json
   // @Success 200 {object} Response
   // @Router /my-endpoint [get]
   func (s *Service) handleMyEndpoint(c *gin.Context) {
       // Implementation
   }
   ```

3. Generate Swagger docs:
   ```bash
   make docs
   ```

4. Add tests:
   ```go
   func TestHandleMyEndpoint(t *testing.T) {
       // Test implementation
   }
   ```

### Adding a New Angular Component

1. Generate component:
   ```bash
   cd ui
   ng generate component components/my-component
   ```

2. Implement component logic
3. Add styles
4. Write tests
5. Export from module if needed

### Adding a New Miner Implementation

1. Create new file: `pkg/mining/myminer.go`
2. Implement the `Miner` interface
3. Register in manager
4. Add tests
5. Update documentation

See [Architecture Guide](architecture.md) for details.

## Release Process

Releases are handled by GoReleaser:

1. Update `CHANGELOG.md`
2. Tag the release:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically:
   - Build for all platforms
   - Create release artifacts
   - Publish to GitHub Releases

For local testing:

```bash
make package
```

This creates a snapshot release in `dist/`.

## Getting Help

- **Documentation**: Check the docs/ folder
- **API Reference**: Use the Swagger UI
- **Issues**: [GitHub Issues](https://github.com/Snider/Mining/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Snider/Mining/discussions)

## Next Steps

- Read the [Architecture Guide](architecture.md)
- Review the [Contributing Guidelines](contributing.md)
- Explore the [API Documentation](../api/index.md)
- See example code in the test files
