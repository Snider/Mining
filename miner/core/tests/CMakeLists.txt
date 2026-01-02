# Test suite for miner project

# Create a library with common test utilities and miner components
add_library(miner_test_lib STATIC
    ${SOURCES}
    ${SOURCES_OS}
    ${SOURCES_CRYPTO}
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
