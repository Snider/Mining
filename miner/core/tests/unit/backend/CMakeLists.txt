# Backend unit tests

# CPU worker tests
add_executable(test_cpu_worker
    test_cpu_worker.cpp
)

target_link_libraries(test_cpu_worker
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_cpu_worker)
