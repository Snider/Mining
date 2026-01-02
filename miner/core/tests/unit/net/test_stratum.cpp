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
#include "base/net/stratum/Job.h"
#include "base/net/stratum/Pool.h"
#include "base/crypto/Algorithm.h"
#include "3rdparty/rapidjson/document.h"

namespace xmrig {

class StratumTest : public ::testing::Test {
protected:
    void SetUp() override {
    }

    void TearDown() override {
    }
};

// Test Job construction and basic properties
TEST_F(StratumTest, JobConstruction) {
    Job job(false, Algorithm::RX_0, "test-client");

    EXPECT_FALSE(job.isValid()) << "Empty job should not be valid";
    EXPECT_EQ(job.algorithm(), Algorithm::RX_0);
    EXPECT_EQ(job.size(), 0) << "Empty job should have size 0";
}

// Test Job ID handling
TEST_F(StratumTest, JobIdHandling) {
    Job job(false, Algorithm::RX_0, "test-client");

    const char* testId = "test-job-123";
    job.setId(testId);

    EXPECT_STREQ(job.id(), testId);
}

// Test Pool URL parsing
TEST_F(StratumTest, PoolUrlParsing) {
    Pool pool("pool.example.com:3333");

    EXPECT_STREQ(pool.host(), "pool.example.com");
    EXPECT_EQ(pool.port(), 3333);
}

// Test Pool URL with protocol
TEST_F(StratumTest, PoolUrlWithProtocol) {
    Pool pool("stratum+tcp://pool.example.com:3333");

    EXPECT_STREQ(pool.host(), "pool.example.com");
    EXPECT_EQ(pool.port(), 3333);
}

// Test Pool SSL URL parsing
TEST_F(StratumTest, PoolSslUrl) {
    Pool pool("stratum+ssl://secure.pool.com:443");

    EXPECT_STREQ(pool.host(), "secure.pool.com");
    EXPECT_EQ(pool.port(), 443);
    EXPECT_TRUE(pool.isTLS());
}

// Test Pool with user/pass
TEST_F(StratumTest, PoolAuthentication) {
    Pool pool("pool.example.com:3333");
    pool.setUser("wallet123");
    pool.setPassword("x");

    EXPECT_STREQ(pool.user(), "wallet123");
    EXPECT_STREQ(pool.password(), "x");
}

// Test Pool algorithm setting
TEST_F(StratumTest, PoolAlgorithm) {
    Pool pool("pool.example.com:3333");
    pool.setAlgo(Algorithm::RX_0);

    EXPECT_EQ(pool.algorithm(), Algorithm::RX_0);
}

// Test Job size calculation
TEST_F(StratumTest, JobSize) {
    Job job(false, Algorithm::RX_0, "test-client");

    // Job size depends on blob data
    // Empty job should have size 0
    EXPECT_EQ(job.size(), 0);
}

// Test Job difficulty
TEST_F(StratumTest, JobDifficulty) {
    Job job(false, Algorithm::RX_0, "test-client");

    uint64_t testDiff = 100000;
    job.setDiff(testDiff);

    EXPECT_EQ(job.diff(), testDiff);
}

// Test Job height
TEST_F(StratumTest, JobHeight) {
    Job job(false, Algorithm::RX_0, "test-client");

    uint64_t testHeight = 1234567;
    job.setHeight(testHeight);

    EXPECT_EQ(job.height(), testHeight);
}

// Test Pool keepalive setting
TEST_F(StratumTest, PoolKeepalive) {
    Pool pool("pool.example.com:3333");

    pool.setKeepaliveTimeout(60);
    EXPECT_EQ(pool.keepAliveTimeout(), 60);
}

// Test invalid pool URL
TEST_F(StratumTest, InvalidPoolUrl) {
    Pool pool("");

    EXPECT_TRUE(pool.host() == nullptr || strlen(pool.host()) == 0);
}

// Test pool equality
TEST_F(StratumTest, PoolEquality) {
    Pool pool1("pool.example.com:3333");
    Pool pool2("pool.example.com:3333");

    pool1.setUser("user1");
    pool2.setUser("user1");

    // Pools with same host, port, and user should be considered equal
    EXPECT_STREQ(pool1.host(), pool2.host());
    EXPECT_EQ(pool1.port(), pool2.port());
    EXPECT_STREQ(pool1.user(), pool2.user());
}

// Test pool fingerprint (for TLS)
TEST_F(StratumTest, PoolFingerprint) {
    Pool pool("stratum+ssl://secure.pool.com:443");

    const char* testFp = "AA:BB:CC:DD:EE:FF";
    pool.setFingerprint(testFp);

    EXPECT_STREQ(pool.fingerprint(), testFp);
}

} // namespace xmrig
