# Integration tests

# Full mining cycle test
add_executable(test_mining_integration
    test_mining_integration.cpp
)

target_link_libraries(test_mining_integration
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_mining_integration)
