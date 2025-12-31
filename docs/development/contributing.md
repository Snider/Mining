# Contributing Guidelines

Thank you for considering contributing to the Mining Platform! This document provides guidelines for contributing to the project.

## Code of Conduct

We expect all contributors to:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

## How to Contribute

### Reporting Bugs

Before creating a bug report:

1. Check the [existing issues](https://github.com/Snider/Mining/issues) to avoid duplicates
2. Collect relevant information (OS, Go version, logs, etc.)
3. Create a minimal reproducible example if possible

When creating a bug report, include:

- **Title**: Clear, descriptive summary
- **Description**: Detailed explanation of the issue
- **Steps to Reproduce**: Numbered list of steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**:
  - OS and version
  - Go version
  - Mining Platform version
  - Miner software versions
- **Logs**: Relevant log output or error messages
- **Screenshots**: If applicable

**Example:**

```markdown
### Bug: XMRig miner fails to start on Ubuntu 22.04

**Description:**
When attempting to start XMRig through the API, the miner process starts but immediately exits with code 1.

**Steps to Reproduce:**
1. Install Mining Platform v1.0.0
2. Install XMRig via `miner-ctrl install xmrig`
3. Start miner with: `POST /api/v1/mining/miners/xmrig`
4. Check miner status

**Expected Behavior:**
Miner should start and begin mining.

**Actual Behavior:**
Miner process exits immediately with error code 1.

**Environment:**
- OS: Ubuntu 22.04 LTS
- Go: 1.24.0
- Mining Platform: v1.0.0
- XMRig: 6.21.0

**Logs:**
```
[ERROR] Failed to start miner: process exited with code 1
[DEBUG] XMRig output: FAILED TO ALLOCATE MEMORY
```
```

### Requesting Features

Feature requests are welcome! Before submitting:

1. Check if the feature already exists or is planned
2. Search existing feature requests
3. Consider if it fits the project scope

When requesting a feature, include:

- **Use Case**: Why is this feature needed?
- **Description**: What should the feature do?
- **Alternatives**: Have you considered other solutions?
- **Examples**: How would it work?

### Submitting Pull Requests

1. **Fork the Repository**

   ```bash
   git clone https://github.com/YOUR_USERNAME/Mining.git
   cd Mining
   git remote add upstream https://github.com/Snider/Mining.git
   ```

2. **Create a Branch**

   ```bash
   git checkout -b feature/my-feature
   ```

   Branch naming convention:
   - `feature/` - New features
   - `fix/` - Bug fixes
   - `docs/` - Documentation changes
   - `refactor/` - Code refactoring
   - `test/` - Test improvements

3. **Make Your Changes**

   - Write clean, readable code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation
   - Keep commits focused and atomic

4. **Test Your Changes**

   ```bash
   # Run Go tests
   make test
   make lint

   # Run frontend tests
   cd ui && npm test
   cd ui && npm run e2e
   ```

5. **Commit Your Changes**

   Follow [Conventional Commits](https://www.conventionalcommits.org/):

   ```
   type(scope): description

   [optional body]

   [optional footer]
   ```

   **Types:**
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation
   - `style`: Formatting
   - `refactor`: Code restructuring
   - `test`: Tests
   - `chore`: Maintenance

   **Examples:**
   ```bash
   git commit -m "feat(api): Add profile management endpoints"
   git commit -m "fix(miner): Fix XMRig hashrate calculation"
   git commit -m "docs(readme): Update installation instructions"
   ```

6. **Push to Your Fork**

   ```bash
   git push origin feature/my-feature
   ```

7. **Create a Pull Request**

   - Go to the GitHub repository
   - Click "New Pull Request"
   - Select your branch
   - Fill in the PR template
   - Link related issues

### Pull Request Guidelines

A good pull request:

- **Focused**: Addresses a single concern
- **Tested**: Includes tests for new code
- **Documented**: Updates relevant documentation
- **Reviewed**: Self-reviewed before submission
- **Linked**: References related issues

**PR Template:**

```markdown
## Description
Brief description of what this PR does.

## Related Issues
Fixes #123
Relates to #456

## Changes
- Added X functionality
- Fixed Y bug
- Updated Z documentation

## Testing
- [ ] Go tests pass
- [ ] Frontend tests pass
- [ ] E2E tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Commits are atomic and well-described
```

## Development Setup

See the [Development Guide](index.md) for detailed setup instructions.

Quick start:

```bash
# Clone
git clone https://github.com/YOUR_USERNAME/Mining.git
cd Mining

# Install dependencies
go mod download
cd ui && npm install

# Run tests
make test
cd ui && npm test

# Start dev environment
make dev
```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write descriptive variable and function names
- Add comments for exported symbols
- Keep functions small and focused
- Handle errors explicitly

**Good Example:**

```go
// GetMinerStats retrieves real-time statistics from a running miner.
// Returns an error if the miner is not running or stats cannot be fetched.
func (m *Manager) GetMinerStats(name string) (*PerformanceMetrics, error) {
    m.mu.RLock()
    miner, exists := m.miners[name]
    m.mu.RUnlock()

    if !exists {
        return nil, fmt.Errorf("miner %s not found", name)
    }

    stats, err := miner.GetStats()
    if err != nil {
        return nil, fmt.Errorf("failed to get stats: %w", err)
    }

    return stats, nil
}
```

### TypeScript/Angular Code Style

- Follow [Angular Style Guide](https://angular.io/guide/styleguide)
- Use TypeScript strict mode
- Prefer interfaces over types for objects
- Use RxJS operators properly
- Clean up subscriptions in `ngOnDestroy`
- Write unit tests for components and services

**Good Example:**

```typescript
@Component({
  selector: 'app-miner-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './miner-card.component.html',
  styleUrls: ['./miner-card.component.scss']
})
export class MinerCardComponent implements OnInit, OnDestroy {
  @Input() minerId!: string;
  miner$!: Observable<Miner>;

  private destroy$ = new Subject<void>();

  constructor(private minerService: MinerService) {}

  ngOnInit(): void {
    this.miner$ = this.minerService.getMiner(this.minerId).pipe(
      takeUntil(this.destroy$)
    );
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
```

### Documentation Standards

- Document all public APIs
- Include examples in documentation
- Keep documentation up-to-date with code
- Use clear, concise language
- Add diagrams for complex concepts

## Testing Requirements

### Required Tests

All PRs must include appropriate tests:

**Go:**
- Unit tests for new functions
- Integration tests for complex flows
- Table-driven tests where applicable
- Error case coverage

**TypeScript/Angular:**
- Unit tests for components and services
- E2E tests for user flows
- Test edge cases and error states

### Running Tests

```bash
# Go tests
make test                    # All tests
go test -v ./pkg/mining/...  # Specific package
go test -run TestName        # Specific test

# Frontend tests
cd ui
npm test                     # Unit tests
npm run test:coverage        # With coverage
npm run e2e                  # E2E tests
```

### Test Coverage

- Aim for >80% coverage for new code
- Don't decrease overall coverage
- Focus on critical paths

Check coverage:

```bash
make coverage    # Opens HTML report
```

## Documentation

### API Documentation

API changes require Swagger annotation updates:

```go
// @Summary Get miner statistics
// @Description Returns real-time performance metrics for a running miner
// @Tags miners
// @Accept json
// @Produce json
// @Param name path string true "Miner name"
// @Success 200 {object} PerformanceMetrics
// @Failure 404 {object} ErrorResponse
// @Router /miners/{name}/stats [get]
func (s *Service) handleGetStats(c *gin.Context) {
    // Implementation
}
```

Generate docs:

```bash
make docs
```

### User Documentation

Update relevant docs in `docs/`:

- Getting started guides
- API references
- User guides
- Architecture docs

## Review Process

### What to Expect

1. **Automated Checks**: CI runs tests and linters
2. **Code Review**: Maintainers review your code
3. **Feedback**: You may be asked to make changes
4. **Approval**: Once approved, PR will be merged

### Responding to Feedback

- Be open to suggestions
- Ask questions if unclear
- Make requested changes promptly
- Explain your reasoning when necessary
- Keep discussions professional and constructive

## Release Process

Releases are handled by maintainers:

1. Update `CHANGELOG.md`
2. Create a version tag
3. GitHub Actions builds and releases

Contributors don't need to worry about releases unless they're maintainers.

## Getting Help

If you need help:

- **Documentation**: Check the `docs/` folder first
- **Issues**: Search existing issues
- **Discussions**: Use GitHub Discussions for questions
- **Discord**: Join our community server (if available)

## Recognition

Contributors are recognized in:

- `CONTRIBUTORS.md` file
- Release notes
- GitHub contributors page

Thank you for contributing to Mining Platform!

## Quick Links

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Development Guide](index.md)
- [Architecture Guide](architecture.md)
- [API Documentation](../api/index.md)
- [Issue Tracker](https://github.com/Snider/Mining/issues)
- [Discussions](https://github.com/Snider/Mining/discussions)
