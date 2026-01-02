/* XMRig
 * Copyright (c) 2025 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

#include <gtest/gtest.h>
#include <chrono>
#include "crypto/cn/CryptoNight_test.h"
#include "crypto/cn/CnHash.h"
#include "crypto/cn/CnCtx.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class PerformanceTest : public ::testing::Test {
protected:
    void SetUp() override {
        ctx = CnCtx::create(1);
    }

    void TearDown() override {
        if (ctx) {
            CnCtx::release(ctx, 1);
            ctx = nullptr;
        }
    }

    // Helper to measure hash computation time
    template<typename Func>
    double MeasureHashTime(Func hashFunc, int iterations = 100) {
        auto start = std::chrono::high_resolution_clock::now();

        for (int i = 0; i < iterations; i++) {
            hashFunc();
        }

        auto end = std::chrono::high_resolution_clock::now();
        std::chrono::duration<double, std::milli> duration = end - start;

        return duration.count() / iterations; // Average time per hash in ms
    }

    CnCtx *ctx = nullptr;
};

// Benchmark CryptoNight-R single hash
TEST_F(PerformanceTest, CryptoNightRSingleHash) {
    const auto& input = cn_r_test_input[0];
    uint8_t output[32];

    auto hashFunc = [&]() {
        CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
    };

    double avgTime = MeasureHashTime(hashFunc, 10); // Use fewer iterations for slow hashes

    // Log performance (actual benchmark should compare against baseline)
    std::cout << "CryptoNight-R average time: " << avgTime << " ms" << std::endl;

    // Performance should be reasonable (this is a loose bound)
    EXPECT_LT(avgTime, 1000.0) << "Hash should complete in less than 1 second";
}

// Benchmark CryptoNight-R with multiple inputs
TEST_F(PerformanceTest, CryptoNightRMultipleInputs) {
    uint8_t output[32];
    const size_t numInputs = sizeof(cn_r_test_input) / sizeof(cn_r_test_input[0]);

    auto start = std::chrono::high_resolution_clock::now();

    for (size_t i = 0; i < numInputs; i++) {
        const auto& input = cn_r_test_input[i];
        CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> duration = end - start;

    double avgTime = duration.count() / numInputs;
    std::cout << "CryptoNight-R average time (" << numInputs << " inputs): "
              << avgTime << " ms" << std::endl;

    EXPECT_LT(avgTime, 1000.0);
}

// Test hash computation throughput
TEST_F(PerformanceTest, HashThroughput) {
    const auto& input = cn_r_test_input[0];
    uint8_t output[32];

    const int iterations = 100;
    auto start = std::chrono::high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double> duration = end - start;

    double hashesPerSecond = iterations / duration.count();
    std::cout << "Throughput: " << hashesPerSecond << " H/s" << std::endl;

    // Should be able to do at least 1 hash per second
    EXPECT_GT(hashesPerSecond, 1.0);
}

// Test memory allocation performance
TEST_F(PerformanceTest, MemoryAllocationPerformance) {
    const size_t size = 2 * 1024 * 1024; // 2 MB
    const int iterations = 100;

    auto start = std::chrono::high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        auto vm = new VirtualMemory(size, false, false, false, 0);
        delete vm;
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> duration = end - start;

    double avgTime = duration.count() / iterations;
    std::cout << "Average allocation time: " << avgTime << " ms" << std::endl;

    // Memory allocation should be reasonably fast
    EXPECT_LT(avgTime, 100.0) << "Memory allocation should be fast";
}

// Test context creation performance
TEST_F(PerformanceTest, ContextCreationPerformance) {
    const int iterations = 100;

    auto start = std::chrono::high_resolution_clock::now();

    for (int i = 0; i < iterations; i++) {
        auto testCtx = CnCtx::create(1);
        CnCtx::release(testCtx, 1);
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> duration = end - start;

    double avgTime = duration.count() / iterations;
    std::cout << "Average context creation time: " << avgTime << " ms" << std::endl;

    EXPECT_LT(avgTime, 100.0) << "Context creation should be fast";
}

// Stress test with rapid job switching
TEST_F(PerformanceTest, RapidJobSwitching) {
    const size_t numInputs = sizeof(cn_r_test_input) / sizeof(cn_r_test_input[0]);
    uint8_t output[32];

    auto start = std::chrono::high_resolution_clock::now();

    // Rapidly switch between different inputs (simulating job changes)
    for (int round = 0; round < 10; round++) {
        for (size_t i = 0; i < numInputs; i++) {
            const auto& input = cn_r_test_input[i];
            CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
        }
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double> duration = end - start;

    std::cout << "Rapid job switching time: " << duration.count() << " s" << std::endl;

    // Should complete in reasonable time
    EXPECT_LT(duration.count(), 300.0) << "Stress test should complete in reasonable time";
}

// Test consistency of performance across runs
TEST_F(PerformanceTest, PerformanceConsistency) {
    const auto& input = cn_r_test_input[0];
    uint8_t output[32];
    const int iterations = 50;

    std::vector<double> timings;

    for (int i = 0; i < 5; i++) {
        auto start = std::chrono::high_resolution_clock::now();

        for (int j = 0; j < iterations; j++) {
            CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
        }

        auto end = std::chrono::high_resolution_clock::now();
        std::chrono::duration<double, std::milli> duration = end - start;
        timings.push_back(duration.count());
    }

    // Calculate variance
    double mean = 0.0;
    for (auto time : timings) {
        mean += time;
    }
    mean /= timings.size();

    double variance = 0.0;
    for (auto time : timings) {
        variance += (time - mean) * (time - mean);
    }
    variance /= timings.size();

    double stddev = std::sqrt(variance);
    double coefficientOfVariation = (stddev / mean) * 100.0;

    std::cout << "Performance coefficient of variation: " << coefficientOfVariation << "%" << std::endl;

    // Performance should be relatively consistent (CV < 20%)
    EXPECT_LT(coefficientOfVariation, 20.0) << "Performance should be consistent across runs";
}

// Test scaling with input size
TEST_F(PerformanceTest, InputSizeScaling) {
    uint8_t output[32];

    // Test different input sizes from cn_r_test_input
    for (size_t i = 0; i < sizeof(cn_r_test_input) / sizeof(cn_r_test_input[0]); i++) {
        const auto& input = cn_r_test_input[i];

        auto start = std::chrono::high_resolution_clock::now();

        for (int j = 0; j < 10; j++) {
            CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);
        }

        auto end = std::chrono::high_resolution_clock::now();
        std::chrono::duration<double, std::milli> duration = end - start;

        std::cout << "Input size " << input.size << " bytes: "
                  << (duration.count() / 10) << " ms average" << std::endl;
    }

    // Test passes if we don't crash and complete in reasonable time
    SUCCEED();
}

} // namespace xmrig
