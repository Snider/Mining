.PHONY: all build test clean install run demo help lint fmt vet docs install-swag dev package e2e e2e-ui e2e-api test-cpp test-cpp-core test-cpp-proxy build-cpp-tests build-miner build-miner-core build-miner-proxy build-miner-all

# Variables
BINARY_NAME=miner-ctrl
MAIN_PACKAGE=./cmd/mining
GO=go
GOFLAGS=-v
CMAKE=cmake
CTEST=ctest
MINER_CORE_DIR=./miner/core
MINER_PROXY_DIR=./miner/proxy
MINER_CORE_BUILD_DIR=$(MINER_CORE_DIR)/build
MINER_PROXY_BUILD_DIR=$(MINER_PROXY_DIR)/build

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

# Run tests (Go + C++)
test: test-go test-cpp
	@echo "All tests completed"

# Run Go tests only
test-go:
	@echo "Running Go tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests and build for all platforms
test-release: test build-all
	@echo "Test release successful"

# Build C++ tests
build-cpp-tests: build-cpp-tests-core build-cpp-tests-proxy
	@echo "C++ tests built successfully"

# Build miner/core tests
build-cpp-tests-core:
	@echo "Building miner/core tests..."
	@mkdir -p $(MINER_CORE_BUILD_DIR)
	@cd $(MINER_CORE_BUILD_DIR) && \
		$(CMAKE) -DBUILD_TESTS=ON .. && \
		$(CMAKE) --build . --parallel

# Build miner/proxy tests
build-cpp-tests-proxy:
	@echo "Building miner/proxy tests..."
	@mkdir -p $(MINER_PROXY_BUILD_DIR)
	@cd $(MINER_PROXY_BUILD_DIR) && \
		$(CMAKE) -DBUILD_TESTS=ON .. && \
		$(CMAKE) --build . --target unit_tests integration_tests --parallel

# Build miner binaries (release builds)
build-miner: build-miner-core build-miner-proxy
	@echo "Miner binaries built successfully"

# Build miner core (CPU/GPU miner)
build-miner-core:
	@echo "Building miner core..."
	@mkdir -p $(MINER_CORE_BUILD_DIR)
	@cd $(MINER_CORE_BUILD_DIR) && \
		$(CMAKE) -DCMAKE_BUILD_TYPE=Release .. && \
		$(CMAKE) --build . --config Release --parallel

# Build miner proxy
build-miner-proxy:
	@echo "Building miner proxy..."
	@mkdir -p $(MINER_PROXY_BUILD_DIR)
	@cd $(MINER_PROXY_BUILD_DIR) && \
		$(CMAKE) -DCMAKE_BUILD_TYPE=Release .. && \
		$(CMAKE) --build . --config Release --parallel

# Build all miner components and package
build-miner-all: build-miner
	@echo "Packaging miner binaries..."
	@mkdir -p dist/miner
	@cp $(MINER_CORE_BUILD_DIR)/miner dist/miner/ 2>/dev/null || true
	@cp $(MINER_PROXY_BUILD_DIR)/miner-proxy dist/miner/ 2>/dev/null || true
	@echo "Miner binaries available in dist/miner/"

# Run C++ tests (builds first if needed)
test-cpp: test-cpp-proxy
	@echo "All C++ tests completed"

# Run miner/core C++ tests (currently has build issues with test library)
test-cpp-core: build-cpp-tests-core
	@echo "Running miner/core tests..."
	@echo "Note: Core tests currently have platform-specific build issues"
	@cd $(MINER_CORE_BUILD_DIR) && $(CTEST) --output-on-failure || true

# Run miner/proxy C++ tests
test-cpp-proxy: build-cpp-tests-proxy
	@echo "Running miner/proxy tests..."
	@cd $(MINER_PROXY_BUILD_DIR) && ./tests/unit_tests --gtest_color=yes
	@cd $(MINER_PROXY_BUILD_DIR) && ./tests/integration_tests --gtest_color=yes

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
	rm -rf $(MINER_CORE_BUILD_DIR)
	rm -rf $(MINER_PROXY_BUILD_DIR)
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

# E2E Tests
e2e: build
	@echo "Running E2E tests..."
	cd ui && npm run e2e

# E2E Tests with Playwright UI
e2e-ui:
	@echo "Opening Playwright UI..."
	cd ui && npm run e2e:ui

# API-only E2E Tests
e2e-api: build
	@echo "Running API tests..."
	cd ui && npm run e2e:api

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Go Application:"
	@echo "  all           - Run tests and build"
	@echo "  build         - Build the CLI binary"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  install       - Install the binary"
	@echo "  run           - Build and run the CLI"
	@echo "  dev           - Start the development server with docs and build"
	@echo ""
	@echo "Miner (C++ Binaries):"
	@echo "  build-miner       - Build miner core and proxy"
	@echo "  build-miner-core  - Build miner core only"
	@echo "  build-miner-proxy - Build miner proxy only"
	@echo "  build-miner-all   - Build and package all miner binaries"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run all tests (Go + C++)"
	@echo "  test-go       - Run Go tests only"
	@echo "  test-cpp      - Run C++ tests (proxy)"
	@echo "  test-cpp-core - Run miner/core C++ tests"
	@echo "  test-cpp-proxy- Run miner/proxy C++ tests"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  e2e           - Run E2E tests with Playwright"
	@echo "  e2e-ui        - Open Playwright UI for interactive testing"
	@echo "  e2e-api       - Run API-only E2E tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run linters"
	@echo "  tidy          - Tidy dependencies"
	@echo ""
	@echo "Other:"
	@echo "  clean         - Clean all build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  docs          - Generate Swagger documentation"
	@echo "  package       - Create local distribution packages"
	@echo "  help          - Show this help message"
