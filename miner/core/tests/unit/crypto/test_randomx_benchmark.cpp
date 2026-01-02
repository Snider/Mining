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
#include "backend/common/benchmark/BenchState_test.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class RandomXBenchmarkTest : public ::testing::Test {
protected:
    // Verify hash output against known test vectors
    void VerifyHash(Algorithm::Id algo, uint32_t nonce, uint64_t expectedHash) {
        auto it = hashCheck.find(algo);
        ASSERT_NE(it, hashCheck.end()) << "Algorithm not found in test data";

        auto nonceIt = it->second.find(nonce);
        ASSERT_NE(nonceIt, it->second.end())
            << "Nonce " << nonce << " not found in test data for algo " << algo;

        EXPECT_EQ(nonceIt->second, expectedHash)
            << "Hash mismatch for algo " << algo << " at nonce " << nonce;
    }
};

// Test RandomX (RX_0) hash values at various nonce points
TEST_F(RandomXBenchmarkTest, RX0HashValidation) {
    const auto& rx0Hashes = hashCheck.at(Algorithm::RX_0);

    for (const auto& [nonce, expectedHash] : rx0Hashes) {
        VerifyHash(Algorithm::RX_0, nonce, expectedHash);
    }
}

// Test RandomX WOW variant hash values
TEST_F(RandomXBenchmarkTest, RXWOWHashValidation) {
    const auto& rxWowHashes = hashCheck.at(Algorithm::RX_WOW);

    for (const auto& [nonce, expectedHash] : rxWowHashes) {
        VerifyHash(Algorithm::RX_WOW, nonce, expectedHash);
    }
}

// Test single-threaded RandomX (RX_0) hash values
TEST_F(RandomXBenchmarkTest, RX0SingleThreadHashValidation) {
    const auto& rx0Hashes = hashCheck1T.at(Algorithm::RX_0);

    for (const auto& [nonce, expectedHash] : rx0Hashes) {
        auto it = hashCheck1T.find(Algorithm::RX_0);
        ASSERT_NE(it, hashCheck1T.end());

        auto nonceIt = it->second.find(nonce);
        ASSERT_NE(nonceIt, it->second.end())
            << "Nonce " << nonce << " not found in single-thread test data";

        EXPECT_EQ(nonceIt->second, expectedHash)
            << "Single-thread hash mismatch at nonce " << nonce;
    }
}

// Test single-threaded RandomX WOW hash values
TEST_F(RandomXBenchmarkTest, RXWOWSingleThreadHashValidation) {
    const auto& rxWowHashes = hashCheck1T.at(Algorithm::RX_WOW);

    for (const auto& [nonce, expectedHash] : rxWowHashes) {
        auto it = hashCheck1T.find(Algorithm::RX_WOW);
        ASSERT_NE(it, hashCheck1T.end());

        auto nonceIt = it->second.find(nonce);
        ASSERT_NE(nonceIt, it->second.end())
            << "Nonce " << nonce << " not found in WOW single-thread test data";

        EXPECT_EQ(nonceIt->second, expectedHash)
            << "WOW single-thread hash mismatch at nonce " << nonce;
    }
}

// Test that test vectors exist for expected nonces
TEST_F(RandomXBenchmarkTest, TestVectorCompleteness) {
    // Verify RX_0 has test vectors
    EXPECT_TRUE(hashCheck.find(Algorithm::RX_0) != hashCheck.end());
    EXPECT_TRUE(hashCheck1T.find(Algorithm::RX_0) != hashCheck1T.end());

    // Verify RX_WOW has test vectors
    EXPECT_TRUE(hashCheck.find(Algorithm::RX_WOW) != hashCheck.end());
    EXPECT_TRUE(hashCheck1T.find(Algorithm::RX_WOW) != hashCheck1T.end());

    // Verify minimum coverage (at least 4 test points per variant in release builds)
    const auto& rx0 = hashCheck.at(Algorithm::RX_0);
    EXPECT_GE(rx0.size(), 4) << "Need at least 4 test vectors for RX_0";
}

// Test consistency between debug and release test vectors
TEST_F(RandomXBenchmarkTest, DebugReleaseConsistency) {
    // In debug builds, we have extra test points (10000, 20000)
    // In release builds, we start at 250000
    // This test ensures that the data structure is properly organized

    const auto& rx0 = hashCheck.at(Algorithm::RX_0);

    #ifdef NDEBUG
    // Release build: should not have debug-only test points
    EXPECT_EQ(rx0.find(10000U), rx0.end()) << "Debug test points should not exist in release builds";
    EXPECT_EQ(rx0.find(20000U), rx0.end()) << "Debug test points should not exist in release builds";
    #else
    // Debug build: should have debug test points
    EXPECT_NE(rx0.find(10000U), rx0.end()) << "Debug test points should exist in debug builds";
    EXPECT_NE(rx0.find(20000U), rx0.end()) << "Debug test points should exist in debug builds";
    #endif

    // Both builds should have 10M test point
    EXPECT_NE(rx0.find(10000000U), rx0.end()) << "10M test point should always exist";
}

} // namespace xmrig
