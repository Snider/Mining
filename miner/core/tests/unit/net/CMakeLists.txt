# Network unit tests

# Stratum protocol tests
add_executable(test_stratum
    test_stratum.cpp
)

target_link_libraries(test_stratum
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_stratum)

# Job results tests
add_executable(test_job_results
    test_job_results.cpp
)

target_link_libraries(test_job_results
    miner_test_lib
    gtest_main
)

gtest_discover_tests(test_job_results)
