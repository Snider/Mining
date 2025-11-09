.PHONY: all build test clean install run demo help lint fmt vet docs install-swag dev package

# Variables
BINARY_NAME=miner-cli
MAIN_PACKAGE=./cmd/mining
GO=go
GOFLAGS=-v

all: test build

# Build the CLI binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/amd64/linux/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GO) build -o dist/arm64/linux/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/amd64/darwin/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GO) build -o dist/arm64/darwin/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/amd64/windows/$(BINARY_NAME).exe $(MAIN_PACKAGE)

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests and build for all platforms
test-release: test build-all
	@echo "Test release successful"

# Run tests with coverage report
coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run demo
demo:
	@echo "Running demo..."
	$(GO) run main.go

# Run the CLI
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html
	$(GO) clean

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linters
lint: fmt vet
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Generate Swagger documentation
docs:
	@echo "Generating Swagger documentation..."
	swag init -g ./cmd/mining/main.go

# Install the swag CLI
install-swag:
	@echo "Installing swag CLI..."
	go install github.com/swaggo/swag/cmd/swag@latest
# Install the swag CLI
install-goreleaser:
	@echo "Installing go release..."
	go install github.com/goreleaser/goreleaser/v2@latest

# Create local packages using goreleaser
package:
	@echo "Creating local packages with GoReleaser..."
	goreleaser release --snapshot --clean

# Development workflow
dev: tidy docs build
	@echo "Starting development server..."
	./$(BINARY_NAME) serve --host localhost --port 9090 --namespace /api/v1/mining

# Help
help:
	@echo "Available targets:"
	@echo "  all         - Run tests and build"
	@echo "  build       - Build the CLI binary"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  install     - Install the binary"
	@echo "  test        - Run tests"
	@echo "  coverage    - Run tests with coverage report"
	@echo "  demo        - Run the demo"
	@echo "  run         - Build and run the CLI"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run linters"
	@echo "  tidy        - Tidy dependencies"
	@echo "  deps        - Download dependencies"
	@echo "  docs        - Generate Swagger documentation"
	@echo "  install-swag- Install the swag CLI"
	@echo "  package     - Create local distribution packages using GoReleaser"
	@echo "  dev         - Start the development server with docs and build"
	@echo "  help        - Show this help message"
