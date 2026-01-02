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
#include "crypto/cn/CryptoNight_test.h"
#include "crypto/cn/CnHash.h"
#include "crypto/cn/CnCtx.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class CryptoNightTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Allocate memory for crypto context
        ctx = CnCtx::create(1);
    }

    void TearDown() override {
        // Clean up
        if (ctx) {
            CnCtx::release(ctx, 1);
            ctx = nullptr;
        }
    }

    CnCtx *ctx = nullptr;
};

// Test CryptoNight-R hash validation using test vectors
TEST_F(CryptoNightTest, ValidateCryptoNightR) {
    uint8_t output[32];

    for (size_t i = 0; i < sizeof(cn_r_test_input) / sizeof(cn_r_test_input[0]); i++) {
        const auto& input = cn_r_test_input[i];
        const uint8_t* expected = test_output_r + (i * 32);

        // Hash the test input
        CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);

        // Compare with expected output
        EXPECT_EQ(0, memcmp(output, expected, 32))
            << "Hash mismatch for CryptoNight-R at height " << input.height
            << " (test case " << i << ")";
    }
}

// Test basic input/output behavior
TEST_F(CryptoNightTest, BasicHashComputation) {
    uint8_t output1[32];
    uint8_t output2[32];

    const uint8_t* input = test_input;
    const size_t size = 76;

    // Hash the same input twice
    CnHash::fn(Algorithm::CN_R, input, size, output1, &ctx, 1806260);
    CnHash::fn(Algorithm::CN_R, input, size, output2, &ctx, 1806260);

    // Should produce identical outputs
    EXPECT_EQ(0, memcmp(output1, output2, 32))
        << "Identical inputs should produce identical outputs";
}

// Test that different heights produce different hashes (CryptoNight-R is height-dependent)
TEST_F(CryptoNightTest, HeightDependency) {
    uint8_t output1[32];
    uint8_t output2[32];

    const uint8_t* input = cn_r_test_input[0].data;
    const size_t size = cn_r_test_input[0].size;

    // Hash at different heights
    CnHash::fn(Algorithm::CN_R, input, size, output1, &ctx, 1806260);
    CnHash::fn(Algorithm::CN_R, input, size, output2, &ctx, 1806261);

    // Should produce different outputs due to height dependency
    EXPECT_NE(0, memcmp(output1, output2, 32))
        << "Different heights should produce different hashes for CryptoNight-R";
}

// Test empty input handling
TEST_F(CryptoNightTest, EmptyInput) {
    uint8_t output[32];
    uint8_t empty_input[1] = {0};

    // Should not crash with empty/minimal input
    EXPECT_NO_THROW({
        CnHash::fn(Algorithm::CN_R, empty_input, 0, output, &ctx, 1806260);
    });
}

// Test output buffer isolation
TEST_F(CryptoNightTest, OutputIsolation) {
    uint8_t output1[32];
    uint8_t output2[32];

    memset(output1, 0xAA, 32);
    memset(output2, 0xBB, 32);

    const uint8_t* input = cn_r_test_input[0].data;
    const size_t size = cn_r_test_input[0].size;

    CnHash::fn(Algorithm::CN_R, input, size, output1, &ctx, 1806260);
    CnHash::fn(Algorithm::CN_R, input, size, output2, &ctx, 1806260);

    // Both should have the same hash
    EXPECT_EQ(0, memcmp(output1, output2, 32))
        << "Separate output buffers should not affect hash computation";
}

} // namespace xmrig
