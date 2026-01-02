# Crypto unit tests

# CryptoNight hash tests
add_executable(test_cryptonight
    test_cryptonight.cpp
)

target_link_libraries(test_cryptonight
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_cryptonight)

# RandomX benchmark tests
if (WITH_RANDOMX)
    add_executable(test_randomx_benchmark
        test_randomx_benchmark.cpp
    )

    target_link_libraries(test_randomx_benchmark
        miner_test_lib
        gtest_main
    )

    gtest_discover_tests(test_randomx_benchmark)
endif()

# Memory pool tests
add_executable(test_memory_pool
    test_memory_pool.cpp
)

target_link_libraries(test_memory_pool
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_memory_pool)
