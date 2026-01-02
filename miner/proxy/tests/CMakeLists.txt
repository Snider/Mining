# Test configuration for miner-proxy
cmake_minimum_required(VERSION 3.10)

# Fetch Google Test
include(FetchContent)
FetchContent_Declare(
  googletest
  GIT_REPOSITORY https://github.com/google/googletest.git
  GIT_TAG        v1.14.0
)

# For Windows: Prevent overriding the parent project's compiler/linker settings
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

enable_testing()

# Include directories from main project
include_directories(${CMAKE_SOURCE_DIR}/src)
include_directories(${CMAKE_SOURCE_DIR}/src/3rdparty)
include_directories(${UV_INCLUDE_DIR})

# Test utility library
add_library(test_utils STATIC
    utils/test_helpers.cpp
)

target_link_libraries(test_utils
    GTest::gtest
    ${UV_LIBRARIES}
)

# Link against proxy source files needed by tests
set(TEST_PROXY_SOURCES
    ${CMAKE_SOURCE_DIR}/src/proxy/Counters.cpp
    ${CMAKE_SOURCE_DIR}/src/proxy/Error.cpp
    ${CMAKE_SOURCE_DIR}/src/proxy/workers/Worker.cpp
)

# Unit tests
add_executable(unit_tests
    unit/test_nonce_mapper.cpp
    unit/test_login.cpp
    unit/test_counters.cpp
    unit/test_custom_diff.cpp
    unit/test_worker.cpp
    unit/test_error.cpp
    ${TEST_PROXY_SOURCES}
)

target_link_libraries(unit_tests
    GTest::gtest_main
    test_utils
    ${UV_LIBRARIES}
    ${OPENSSL_LIBRARIES}
    ${EXTRA_LIBS}
)

# Integration tests
add_executable(integration_tests
    integration/test_splitter_nicehash.cpp
    integration/test_splitter_simple.cpp
    integration/test_event_system.cpp
)

target_link_libraries(integration_tests
    GTest::gtest_main
    test_utils
    ${UV_LIBRARIES}
    ${OPENSSL_LIBRARIES}
    ${EXTRA_LIBS}
)

# Register tests with CTest
include(GoogleTest)
gtest_discover_tests(unit_tests)
gtest_discover_tests(integration_tests)

# Add convenience target to run all tests
add_custom_target(run_tests
    COMMAND ${CMAKE_CTEST_COMMAND} --output-on-failure
    DEPENDS unit_tests integration_tests
)
