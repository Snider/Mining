# Benchmark tests

# Performance regression tests
add_executable(test_performance
    test_performance.cpp
)

target_link_libraries(test_performance
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_performance)
