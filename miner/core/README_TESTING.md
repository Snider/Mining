# Testing Guide

This document describes the testing infrastructure for the miner project.

## Overview

The project uses Google Test framework for unit, integration, and benchmark tests. Tests are automatically built when `BUILD_TESTS=ON` is set.

## Building Tests

```bash
mkdir build && cd build
cmake .. -DBUILD_TESTS=ON -DCMAKE_BUILD_TYPE=Debug
cmake --build .
```

## Running Tests

### Run all tests
```bash
cd build
ctest --output-on-failure
```

### Run specific test suite
```bash
# Run only crypto tests
ctest -R crypto --output-on-failure

# Run only network tests
ctest -R net --output-on-failure

# Run only integration tests
ctest -R integration --output-on-failure

# Run only benchmark tests
ctest -R benchmark --output-on-failure
```

### Run individual test executable
```bash
cd build
./tests/unit/crypto/test_cryptonight
./tests/unit/crypto/test_randomx_benchmark
./tests/unit/net/test_stratum
```

## Test Structure

```
tests/
├── unit/                       # Unit tests
│   ├── crypto/                 # Cryptographic algorithm tests
│   │   ├── test_cryptonight.cpp
│   │   ├── test_randomx_benchmark.cpp
│   │   └── test_memory_pool.cpp
│   ├── backend/                # Backend tests
│   │   └── test_cpu_worker.cpp
│   ├── net/                    # Network protocol tests
│   │   ├── test_stratum.cpp
│   │   └── test_job_results.cpp
│   └── config/                 # Configuration tests
│       └── test_config.cpp
├── integration/                # Integration tests
│   └── test_mining_integration.cpp
└── benchmark/                  # Performance tests
    └── test_performance.cpp
```

## Test Coverage Areas

### Crypto Tests
- **test_cryptonight.cpp**: CryptoNight algorithm validation using test vectors
- **test_randomx_benchmark.cpp**: RandomX hash validation against known benchmarks
- **test_memory_pool.cpp**: Memory allocation and management

### Backend Tests
- **test_cpu_worker.cpp**: Hashrate calculation, algorithm handling

### Network Tests
- **test_stratum.cpp**: Pool URL parsing, authentication, protocol handling
- **test_job_results.cpp**: Job result creation and submission

### Config Tests
- **test_config.cpp**: JSON parsing, configuration validation

### Integration Tests
- **test_mining_integration.cpp**: End-to-end mining flow, algorithm switching

### Benchmark Tests
- **test_performance.cpp**: Performance regression detection, throughput measurement

## CI/CD Integration

Tests run automatically on:
- Every push to `main` or `develop` branches
- Every pull request
- Nightly at 2 AM UTC (includes extended benchmarks)

Platforms tested:
- Linux (Ubuntu) - GCC and Clang
- Windows (MSVC)
- macOS (Apple Clang)

## Code Coverage

Coverage is tracked on Linux Debug builds:

```bash
cmake .. -DCMAKE_BUILD_TYPE=Debug -DBUILD_TESTS=ON \
         -DCMAKE_CXX_FLAGS="--coverage" -DCMAKE_C_FLAGS="--coverage"
cmake --build .
ctest
lcov --capture --directory . --output-file coverage.info
lcov --remove coverage.info '/usr/*' '*/tests/*' '*/3rdparty/*' --output-file coverage.info
genhtml coverage.info --output-directory coverage_html
```

## Writing New Tests

### Unit Test Template

```cpp
#include <gtest/gtest.h>
#include "your/header.h"

namespace xmrig {

class YourTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Setup code
    }

    void TearDown() override {
        // Cleanup code
    }
};

TEST_F(YourTest, TestName) {
    EXPECT_EQ(expected, actual);
    ASSERT_TRUE(condition);
}

} // namespace xmrig
```

### Adding Test to CMake

Edit `tests/unit/<category>/CMakeLists.txt`:

```cmake
add_executable(test_your_feature
    test_your_feature.cpp
)

target_link_libraries(test_your_feature
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_your_feature)
```

## Best Practices

1. **Test Names**: Use descriptive names that explain what is being tested
2. **Isolation**: Each test should be independent and not rely on other tests
3. **Fast Tests**: Keep unit tests fast (< 1 second each)
4. **Assertions**: Use `EXPECT_*` for non-fatal, `ASSERT_*` for fatal assertions
5. **Test Data**: Use existing test vectors from `*_test.h` files when available
6. **Coverage**: Aim for at least 80% code coverage for critical paths

## Debugging Tests

### Run test with verbose output
```bash
cd build
./tests/unit/crypto/test_cryptonight --gtest_filter="*" --gtest_verbose
```

### Run test under GDB
```bash
gdb --args ./tests/unit/crypto/test_cryptonight
```

### Run single test case
```bash
./tests/unit/crypto/test_cryptonight --gtest_filter="CryptoNightTest.ValidateCryptoNightR"
```

## Performance Testing

Benchmark tests measure:
- Hash computation time
- Memory allocation performance
- Context creation overhead
- Throughput under load

Run performance tests separately:
```bash
ctest -R performance --output-on-failure
```

## Continuous Integration

GitHub Actions workflow (`.github/workflows/test.yml`) runs:
- Debug and Release builds
- Multiple compilers (GCC, Clang, MSVC)
- Code coverage analysis
- Nightly benchmark runs

## Known Issues

- GPU tests (CUDA/OpenCL) require hardware and are disabled in CI
- Some tests may be slow in Debug builds due to unoptimized crypto code
- Coverage may be incomplete for platform-specific code

## Contributing

When adding new features:
1. Write tests first (TDD approach recommended)
2. Ensure all existing tests pass
3. Add tests for edge cases and error conditions
4. Update this documentation if adding new test categories
