# Config unit tests

# Configuration parsing tests
add_executable(test_config
    test_config.cpp
)

target_link_libraries(test_config
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_config)
