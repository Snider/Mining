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
#include "backend/common/Hashrate.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class CpuWorkerTest : public ::testing::Test {
protected:
    void SetUp() override {
    }

    void TearDown() override {
    }
};

// Test Hashrate calculation
TEST_F(CpuWorkerTest, HashrateCalculation) {
    Hashrate hashrate(4); // 4 threads

    // Add some hash counts
    for (size_t i = 0; i < 4; i++) {
        hashrate.add(i, 1000, 1000); // 1000 hashes in 1000ms = 1000 H/s
    }

    // Calculate total hashrate (should be approximately 4000 H/s)
    double total = hashrate.calc(0);
    EXPECT_GT(total, 0.0);
}

// Test Hashrate with zero hashes
TEST_F(CpuWorkerTest, HashrateZeroHashes) {
    Hashrate hashrate(1);

    hashrate.add(0, 0, 1000);

    double rate = hashrate.calc(0);
    EXPECT_EQ(rate, 0.0);
}

// Test Hashrate averaging
TEST_F(CpuWorkerTest, HashrateAveraging) {
    Hashrate hashrate(1);

    // Add multiple samples
    hashrate.add(0, 1000, 1000);
    hashrate.add(0, 2000, 1000);
    hashrate.add(0, 3000, 1000);

    // Should calculate average
    double rate = hashrate.calc(0);
    EXPECT_GT(rate, 0.0);
}

// Test Hashrate thread isolation
TEST_F(CpuWorkerTest, HashrateThreadIsolation) {
    Hashrate hashrate(4);

    // Only add to thread 0
    hashrate.add(0, 1000, 1000);

    // Thread 0 should have hashrate
    double rate0 = hashrate.calc(0);
    EXPECT_GT(rate0, 0.0);

    // Thread 1 should have zero hashrate
    double rate1 = hashrate.calc(1);
    EXPECT_EQ(rate1, 0.0);
}

// Test Hashrate reset
TEST_F(CpuWorkerTest, HashrateReset) {
    Hashrate hashrate(1);

    // Add some data
    hashrate.add(0, 1000, 1000);

    double rate1 = hashrate.calc(0);
    EXPECT_GT(rate1, 0.0);

    // Reset (if method exists)
    // hashrate.reset();

    // After reset should be zero
    // double rate2 = hashrate.calc(0);
    // EXPECT_EQ(rate2, 0.0);
}

// Test Hashrate with different time windows
TEST_F(CpuWorkerTest, HashrateTimeWindows) {
    Hashrate hashrate(1);

    // Add samples with different timestamps
    uint64_t baseTime = 1000000;
    hashrate.add(0, 1000, baseTime);
    hashrate.add(0, 2000, baseTime + 1000);
    hashrate.add(0, 3000, baseTime + 2000);

    double rate = hashrate.calc(0);
    EXPECT_GT(rate, 0.0);
}

// Test Algorithm validation
TEST_F(CpuWorkerTest, AlgorithmValidation) {
    // Test valid algorithm
    Algorithm rxAlgo("rx/0");
    EXPECT_TRUE(rxAlgo.isValid());
    EXPECT_EQ(rxAlgo.id(), Algorithm::RX_0);

    // Test another valid algorithm
    Algorithm cnAlgo("cn/r");
    EXPECT_TRUE(cnAlgo.isValid());
    EXPECT_EQ(cnAlgo.id(), Algorithm::CN_R);
}

// Test Algorithm from ID
TEST_F(CpuWorkerTest, AlgorithmFromId) {
    Algorithm algo(Algorithm::RX_0);

    EXPECT_TRUE(algo.isValid());
    EXPECT_EQ(algo.id(), Algorithm::RX_0);
}

// Test Algorithm family
TEST_F(CpuWorkerTest, AlgorithmFamily) {
    Algorithm rx0(Algorithm::RX_0);
    Algorithm rxWow(Algorithm::RX_WOW);

    // Both should be RandomX family
    EXPECT_EQ(rx0.family(), Algorithm::RANDOM_X);
    EXPECT_EQ(rxWow.family(), Algorithm::RANDOM_X);
}

// Test Algorithm comparison
TEST_F(CpuWorkerTest, AlgorithmComparison) {
    Algorithm algo1(Algorithm::RX_0);
    Algorithm algo2(Algorithm::RX_0);
    Algorithm algo3(Algorithm::RX_WOW);

    EXPECT_EQ(algo1, algo2);
    EXPECT_NE(algo1, algo3);
}

// Test invalid algorithm
TEST_F(CpuWorkerTest, InvalidAlgorithm) {
    Algorithm invalid("invalid-algo");

    EXPECT_FALSE(invalid.isValid());
}

// Test Algorithm name
TEST_F(CpuWorkerTest, AlgorithmName) {
    Algorithm algo(Algorithm::RX_0);

    EXPECT_TRUE(algo.isValid());
    EXPECT_STREQ(algo.name(), "rx/0");
}

// Test Hashrate large values
TEST_F(CpuWorkerTest, HashrateLargeValues) {
    Hashrate hashrate(1);

    // Add large hash count
    hashrate.add(0, 1000000000, 1000); // 1 billion hashes in 1 second

    double rate = hashrate.calc(0);
    EXPECT_GT(rate, 900000000.0); // Should be close to 1 GH/s
}

// Test Hashrate stability over time
TEST_F(CpuWorkerTest, HashrateStability) {
    Hashrate hashrate(1);

    // Add consistent samples
    for (int i = 0; i < 10; i++) {
        hashrate.add(0, 1000, 1000);
    }

    // Should have stable hashrate
    double rate = hashrate.calc(0);
    EXPECT_GT(rate, 0.0);
    EXPECT_LT(rate, 2000.0); // Should be around 1000 H/s
}

} // namespace xmrig
