# Testing Guide for Miner Proxy

This document provides comprehensive guidance on testing the miner-proxy project.

## Quick Start

```bash
# Build and run all tests
mkdir build && cd build
cmake ..
make -j$(nproc)
cd tests
make -j$(nproc)
./unit_tests
./integration_tests
```

## Test Framework

We use [Google Test](https://github.com/google/googletest) (gtest) version 1.14.0, automatically fetched via CMake's FetchContent.

## Test Organization

### Unit Tests (`tests/unit/`)

Test individual components in isolation:

| Test File | Component | Status |
|-----------|-----------|--------|
| `test_counters.cpp` | Share counters, miner count | ✅ Complete |
| `test_custom_diff.cpp` | Per-user difficulty | ✅ Complete |
| `test_error.cpp` | Error code mapping | ✅ Complete |
| `test_worker.cpp` | Worker tracking | ✅ Complete |
| `test_login.cpp` | Login validation | 🚧 Needs mocks |
| `test_nonce_mapper.cpp` | Nonce transformation | 🚧 Needs mocks |

### Integration Tests (`tests/integration/`)

Test component interactions:

| Test File | Purpose | Status |
|-----------|---------|--------|
| `test_splitter_nicehash.cpp` | NiceHash splitter flow | 🚧 Needs mocks |
| `test_splitter_simple.cpp` | Simple splitter mode | 🚧 Needs mocks |
| `test_event_system.cpp` | Event dispatching | 🚧 Needs setup |

### Test Utilities (`tests/utils/`)

- **`test_helpers.h/cpp`** - Base test fixtures, mock objects, test data generators
- **`ProxyTestBase`** - Common test fixture with temp file management
- **`TestDataGenerator`** - Generate valid stratum JSON (login, job, submit)
- **`MockController`** - Placeholder for Controller mocking

## Running Tests

### All Tests

```bash
cd build
ctest --output-on-failure
```

### Specific Test Suite

```bash
./tests/unit_tests --gtest_filter=CountersTest.*
./tests/unit_tests --gtest_filter=CustomDiffTest.*
```

### Single Test

```bash
./tests/unit_tests --gtest_filter=CountersTest.InitialStateIsZero
```

### List Available Tests

```bash
./tests/unit_tests --gtest_list_tests
```

### Verbose Output

```bash
./tests/unit_tests --gtest_verbose
```

### Repeat Tests

```bash
./tests/unit_tests --gtest_repeat=100
```

### Shuffle Test Order

```bash
./tests/unit_tests --gtest_shuffle
```

## Code Coverage

### Generate Coverage Report

```bash
# Configure with coverage flags
mkdir build-coverage && cd build-coverage
cmake .. -DCMAKE_BUILD_TYPE=Debug \
  -DCMAKE_CXX_FLAGS="--coverage" \
  -DCMAKE_C_FLAGS="--coverage"

# Build and run tests
make -j$(nproc)
cd tests && make -j$(nproc)
./unit_tests
./integration_tests

# Generate HTML report
gcovr --root ../.. --exclude ../../tests \
  --html --html-details -o coverage.html

# Open in browser
xdg-open coverage.html
```

### Coverage Tools

- **gcovr** - Recommended, included in most Linux distros
- **lcov** - Alternative tool with HTML output
- **Codecov** - CI integration (automatic via GitHub Actions)

## CI/CD Integration

Tests run automatically on:
- Every push to `main` or `develop`
- All pull requests
- Manual workflow dispatch

Platforms tested:
- **Linux** (Ubuntu latest) - Primary platform
- **macOS** (latest) - Secondary platform
- **Windows** (latest) - Compatibility check

See `.github/workflows/test.yml` for configuration.

## Writing New Tests

### Unit Test Template

```cpp
#include <gtest/gtest.h>
#include "proxy/YourComponent.h"
#include "../utils/test_helpers.h"

using namespace xmrig;
using namespace xmrig::test;

class YourComponentTest : public ProxyTestBase {
protected:
    void SetUp() override {
        ProxyTestBase::SetUp();
        component = new YourComponent();
    }

    void TearDown() override {
        delete component;
        ProxyTestBase::TearDown();
    }

    YourComponent* component = nullptr;
};

TEST_F(YourComponentTest, DescriptiveTestName) {
    // Arrange
    int input = 42;

    // Act
    int result = component->process(input);

    // Assert
    EXPECT_EQ(result, 84);
    ASSERT_NE(component, nullptr);
}
```

### Integration Test Template

```cpp
#include <gtest/gtest.h>
#include "../utils/test_helpers.h"

using namespace xmrig;
using namespace xmrig::test;

class SystemIntegrationTest : public ProxyTestBase {
protected:
    void SetUp() override {
        ProxyTestBase::SetUp();
        // Setup multiple components
    }
};

TEST_F(SystemIntegrationTest, EndToEndFlow) {
    // Test complete workflow through multiple components
    SUCCEED();
}
```

## Test Best Practices

### Naming Conventions

- **Test files**: `test_<component>.cpp`
- **Test fixtures**: `<Component>Test`
- **Test cases**: `<DescriptiveAction><ExpectedResult>`

Examples:
- `TEST_F(CountersTest, IncrementMinerCount)`
- `TEST_F(CustomDiffTest, OverwriteExistingDifficulty)`

### Assertions

Use appropriate assertions:
- `EXPECT_*` - Continues test on failure (preferred for multiple checks)
- `ASSERT_*` - Stops test on failure (use for critical checks)

Common assertions:
```cpp
EXPECT_EQ(a, b)         // a == b
EXPECT_NE(a, b)         // a != b
EXPECT_LT(a, b)         // a < b
EXPECT_LE(a, b)         // a <= b
EXPECT_GT(a, b)         // a > b
EXPECT_GE(a, b)         // a >= b
EXPECT_TRUE(condition)
EXPECT_FALSE(condition)
EXPECT_STREQ(s1, s2)    // C strings
```

### Test Organization

1. **Arrange** - Set up test data and conditions
2. **Act** - Execute the code under test
3. **Assert** - Verify the results

### Test Independence

- Each test should be independent
- Use `SetUp()` and `TearDown()` for initialization/cleanup
- Don't rely on test execution order

## Debugging Tests

### Run with GDB

```bash
gdb --args ./tests/unit_tests --gtest_filter=CountersTest.InitialStateIsZero
(gdb) break CountersTest_InitialStateIsZero_Test::TestBody
(gdb) run
```

### Valgrind Memory Check

```bash
valgrind --leak-check=full --show-leak-kinds=all \
  ./tests/unit_tests --gtest_filter=CountersTest.*
```

### Address Sanitizer

```bash
# Configure with ASAN
cmake .. -DCMAKE_BUILD_TYPE=Debug \
  -DCMAKE_CXX_FLAGS="-fsanitize=address -fno-omit-frame-pointer"

# Build and run
make -j$(nproc)
./tests/unit_tests
```

## Future Enhancements

### Priority 1 - Mocking Infrastructure
- Create mock implementations for Controller, IClient, IStrategy
- Complete NonceMapper and Login integration tests
- Mock libuv networking for full E2E tests

### Priority 2 - Protocol Testing
- Stratum protocol request/response validation
- Invalid input handling (fuzzing)
- Protocol extension tests (NiceHash, algorithm negotiation)

### Priority 3 - Performance Testing
- Load tests with 100K+ simulated miners
- Nonce mapping throughput benchmarks
- Memory usage profiling under load
- Connection handling stress tests

### Priority 4 - Additional Coverage
- TLS/SSL connection tests
- Configuration loading/validation
- HTTP API endpoint tests
- Log output validation
- Platform-specific code (Windows/macOS/Linux)

## Troubleshooting

### Tests Won't Build

```bash
# Clean build
rm -rf build
mkdir build && cd build
cmake .. -DBUILD_TESTS=ON
make clean
make -j$(nproc)
```

### Google Test Not Found

Google Test is automatically downloaded via CMake FetchContent. If you have network issues:

```bash
# Download manually
git clone https://github.com/google/googletest.git
cd googletest && mkdir build && cd build
cmake .. && make && sudo make install
```

### Link Errors

Ensure all dependencies are installed:
```bash
# Ubuntu/Debian
sudo apt-get install libuv1-dev libssl-dev

# macOS
brew install libuv openssl

# Check CMake finds them
cmake .. -DWITH_TLS=ON -DWITH_HTTP=ON
```

## Contributing

When adding new features:
1. Write tests first (TDD approach recommended)
2. Ensure all existing tests pass
3. Add integration tests for new subsystems
4. Update this documentation
5. Verify CI passes on all platforms

## Resources

- [Google Test Documentation](https://google.github.io/googletest/)
- [Google Mock Documentation](https://google.github.io/googletest/gmock_for_dummies.html)
- [CMake Testing](https://cmake.org/cmake/help/latest/manual/ctest.1.html)
- [gcovr Documentation](https://gcovr.com/)
