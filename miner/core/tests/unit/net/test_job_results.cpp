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
#include "net/JobResult.h"
#include "net/JobResults.h"
#include "base/net/stratum/Job.h"
#include "net/interfaces/IJobResultListener.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

// Mock listener for testing
class MockJobResultListener : public IJobResultListener {
public:
    MockJobResultListener() : submitCount(0), acceptedCount(0), rejectedCount(0) {}

    void onJobResult(const JobResult& result) override {
        submitCount++;
        lastResult = result;
    }

    void onResultAccepted(IClient* client, int64_t ms, const char* error) override {
        if (error == nullptr) {
            acceptedCount++;
        } else {
            rejectedCount++;
        }
    }

    int submitCount;
    int acceptedCount;
    int rejectedCount;
    JobResult lastResult;
};

class JobResultsTest : public ::testing::Test {
protected:
    void SetUp() override {
        listener = new MockJobResultListener();
    }

    void TearDown() override {
        JobResults::stop();
        delete listener;
        listener = nullptr;
    }

    MockJobResultListener* listener;
};

// Test JobResult construction
TEST_F(JobResultsTest, JobResultConstruction) {
    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-1");

    uint32_t testNonce = 0x12345678;
    uint8_t testResult[32] = {0};

    JobResult result(job, testNonce, testResult);

    EXPECT_STREQ(result.jobId, "test-job-1");
    EXPECT_EQ(result.nonce, testNonce);
    EXPECT_EQ(result.algorithm, Algorithm::RX_0);
}

// Test JobResult data integrity
TEST_F(JobResultsTest, JobResultDataIntegrity) {
    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-2");
    job.setDiff(100000);

    uint32_t testNonce = 0xABCDEF00;
    uint8_t testResult[32];

    // Fill with test pattern
    for (int i = 0; i < 32; i++) {
        testResult[i] = static_cast<uint8_t>(i);
    }

    JobResult result(job, testNonce, testResult);

    // Verify data
    EXPECT_STREQ(result.jobId, "test-job-2");
    EXPECT_EQ(result.nonce, testNonce);
    EXPECT_EQ(result.diff, 100000);

    // Verify result hash
    for (int i = 0; i < 32; i++) {
        EXPECT_EQ(result.result[i], static_cast<uint8_t>(i))
            << "Result byte " << i << " mismatch";
    }
}

// Test basic job submission
TEST_F(JobResultsTest, BasicSubmission) {
    JobResults::setListener(listener, true);

    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-3");

    uint32_t nonce = 0x11111111;
    uint8_t result[32] = {0};

    JobResults::submit(job, nonce, result);

    // Give some time for async processing
    // Note: In real tests, you'd want proper synchronization
}

// Test client ID propagation
TEST_F(JobResultsTest, ClientIdPropagation) {
    const char* testClientId = "test-client-123";

    Job job(false, Algorithm::RX_0, testClientId);
    job.setId("test-job-4");

    uint32_t nonce = 0x22222222;
    uint8_t result[32] = {0};

    JobResult jobResult(job, nonce, result);

    EXPECT_STREQ(jobResult.clientId, testClientId);
}

// Test backend ID assignment
TEST_F(JobResultsTest, BackendIdAssignment) {
    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-5");
    job.setBackend(Job::CPU);

    uint32_t nonce = 0x33333333;
    uint8_t result[32] = {0};

    JobResult jobResult(job, nonce, result);

    EXPECT_EQ(jobResult.backend, Job::CPU);
}

// Test difficulty tracking
TEST_F(JobResultsTest, DifficultyTracking) {
    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-6");

    uint64_t testDiff = 500000;
    job.setDiff(testDiff);

    uint32_t nonce = 0x44444444;
    uint8_t result[32] = {0};

    JobResult jobResult(job, nonce, result);

    EXPECT_EQ(jobResult.diff, testDiff);
}

// Test algorithm preservation
TEST_F(JobResultsTest, AlgorithmPreservation) {
    Algorithm::Id testAlgo = Algorithm::RX_WOW;

    Job job(false, testAlgo, "test-client");
    job.setId("test-job-7");

    uint32_t nonce = 0x55555555;
    uint8_t result[32] = {0};

    JobResult jobResult(job, nonce, result);

    EXPECT_EQ(jobResult.algorithm, testAlgo);
}

// Test multiple submissions
TEST_F(JobResultsTest, MultipleSubmissions) {
    JobResults::setListener(listener, true);

    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-multi");

    uint8_t result[32] = {0};

    // Submit multiple results
    for (uint32_t i = 0; i < 5; i++) {
        JobResults::submit(job, 0x10000000 + i, result);
    }

    // Verify listener was called (would need proper async handling in production)
    // Test structure is here for documentation
}

// Test result hash uniqueness
TEST_F(JobResultsTest, ResultHashUniqueness) {
    Job job(false, Algorithm::RX_0, "test-client");
    job.setId("test-job-8");

    uint32_t nonce1 = 0x66666666;
    uint32_t nonce2 = 0x77777777;

    uint8_t result1[32];
    uint8_t result2[32];

    // Fill with different patterns
    for (int i = 0; i < 32; i++) {
        result1[i] = static_cast<uint8_t>(i);
        result2[i] = static_cast<uint8_t>(i + 1);
    }

    JobResult jr1(job, nonce1, result1);
    JobResult jr2(job, nonce2, result2);

    // Verify different nonces
    EXPECT_NE(jr1.nonce, jr2.nonce);

    // Verify different results
    EXPECT_NE(0, memcmp(jr1.result, jr2.result, 32));
}

} // namespace xmrig
