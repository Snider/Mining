# Test suite for miner project

# Manually specify source files needed for tests with absolute paths
file(GLOB_RECURSE MINER_SOURCES
    ${CMAKE_SOURCE_DIR}/src/3rdparty/fmt/*.cc
    ${CMAKE_SOURCE_DIR}/src/base/*.cpp
    ${CMAKE_SOURCE_DIR}/src/backend/common/*.cpp
    ${CMAKE_SOURCE_DIR}/src/backend/cpu/*.cpp
    ${CMAKE_SOURCE_DIR}/src/crypto/cn/*.cpp
    ${CMAKE_SOURCE_DIR}/src/crypto/cn/*.c
    ${CMAKE_SOURCE_DIR}/src/crypto/common/*.cpp
)

# Remove test files and platform-specific files not for this architecture
list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_test\\.cpp$")
list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_arm.*")
list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_riscv.*")
if (WIN32)
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_unix\\.cpp$")
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_linux\\.cpp$")
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_mac\\.cpp$")
elseif (APPLE)
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_win\\.cpp$")
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_linux\\.cpp$")
else()
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_win\\.cpp$")
    list(FILTER MINER_SOURCES EXCLUDE REGEX ".*_mac\\.cpp$")
endif()

# Create a library with common test utilities and miner components
add_library(miner_test_lib STATIC
    ${MINER_SOURCES}
)

target_include_directories(miner_test_lib PUBLIC
    ${CMAKE_SOURCE_DIR}/src
    ${CMAKE_SOURCE_DIR}/src/3rdparty
    ${UV_INCLUDE_DIR}
)

target_link_libraries(miner_test_lib PUBLIC
    ${XMRIG_ASM_LIBRARY}
    ${OPENSSL_LIBRARIES}
    ${UV_LIBRARIES}
    ${EXTRA_LIBS}
    ${CPUID_LIB}
    ${ARGON2_LIBRARY}
    ${ETHASH_LIBRARY}
    ${GHOSTRIDER_LIBRARY}
    ${BLAKE3_LIBRARY}
)

# Unit tests
add_subdirectory(unit)

# Integration tests
add_subdirectory(integration)

# Benchmark tests
add_subdirectory(benchmark)
