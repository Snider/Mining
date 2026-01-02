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
#include "net/JobResult.h"
#include "crypto/cn/CnHash.h"
#include "crypto/cn/CnCtx.h"
#include "crypto/cn/CryptoNight_test.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class MiningIntegrationTest : public ::testing::Test {
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

    CnCtx *ctx = nullptr;
};

// Test complete job creation and result submission flow
TEST_F(MiningIntegrationTest, JobToResultFlow) {
    // Create a job
    Job job(false, Algorithm::CN_R, "integration-test-client");
    job.setId("test-job-integration-1");
    job.setDiff(100000);
    job.setHeight(1806260);

    EXPECT_TRUE(job.algorithm().isValid());
    EXPECT_STREQ(job.id(), "test-job-integration-1");

    // Simulate mining (hash computation)
    const auto& input = cn_r_test_input[0];
    uint8_t output[32];

    CnHash::fn(Algorithm::CN_R, input.data, input.size, output, &ctx, input.height);

    // Create result
    JobResult result(job, 0x12345678, output);

    EXPECT_STREQ(result.jobId, "test-job-integration-1");
    EXPECT_EQ(result.algorithm, Algorithm::CN_R);
    EXPECT_EQ(result.diff, 100000);
}

// Test algorithm switching
TEST_F(MiningIntegrationTest, AlgorithmSwitching) {
    // Start with RX_0
    Algorithm algo1(Algorithm::RX_0);
    EXPECT_EQ(algo1.id(), Algorithm::RX_0);

    // Switch to CN_R
    Algorithm algo2(Algorithm::CN_R);
    EXPECT_EQ(algo2.id(), Algorithm::CN_R);

    // Create jobs with different algorithms
    Job job1(false, algo1, "client1");
    Job job2(false, algo2, "client2");

    EXPECT_EQ(job1.algorithm(), algo1);
    EXPECT_EQ(job2.algorithm(), algo2);
}

// Test multiple job handling
TEST_F(MiningIntegrationTest, MultipleJobHandling) {
    std::vector<Job> jobs;

    // Create multiple jobs
    for (int i = 0; i < 5; i++) {
        Job job(false, Algorithm::CN_R, "multi-client");
        job.setId((std::string("job-") + std::to_string(i)).c_str());
        job.setDiff(100000 + i * 10000);
        jobs.push_back(job);
    }

    EXPECT_EQ(jobs.size(), 5);

    // Verify each job is unique
    for (size_t i = 0; i < jobs.size(); i++) {
        EXPECT_EQ(jobs[i].diff(), 100000 + i * 10000);
    }
}

// Test hash validation cycle
TEST_F(MiningIntegrationTest, HashValidationCycle) {
    // Use test vectors for validation
    const auto& input = cn_r_test_input[0];
    const uint8_t* expectedHash = test_output_r;

    uint8_t computedHash[32];

    // Compute hash
    CnHash::fn(Algorithm::CN_R, input.data, input.size, computedHash, &ctx, input.height);

    // Validate
    EXPECT_EQ(0, memcmp(computedHash, expectedHash, 32))
        << "Computed hash should match test vector";

    // Create job result with validated hash
    Job job(false, Algorithm::CN_R, "validation-client");
    job.setId("validation-job");
    job.setHeight(input.height);

    JobResult result(job, 0xDEADBEEF, computedHash);

    // Verify result integrity
    EXPECT_EQ(0, memcmp(result.result, expectedHash, 32));
}

// Test backend type propagation
TEST_F(MiningIntegrationTest, BackendTypePropagation) {
    Job cpuJob(false, Algorithm::RX_0, "cpu-client");
    cpuJob.setBackend(Job::CPU);

    EXPECT_EQ(cpuJob.backend(), Job::CPU);

    uint8_t dummyHash[32] = {0};
    JobResult cpuResult(cpuJob, 0x11111111, dummyHash);

    EXPECT_EQ(cpuResult.backend, Job::CPU);

#ifdef XMRIG_FEATURE_OPENCL
    Job oclJob(false, Algorithm::RX_0, "ocl-client");
    oclJob.setBackend(Job::OPENCL);

    EXPECT_EQ(oclJob.backend(), Job::OPENCL);

    JobResult oclResult(oclJob, 0x22222222, dummyHash);
    EXPECT_EQ(oclResult.backend, Job::OPENCL);
#endif

#ifdef XMRIG_FEATURE_CUDA
    Job cudaJob(false, Algorithm::RX_0, "cuda-client");
    cudaJob.setBackend(Job::CUDA);

    EXPECT_EQ(cudaJob.backend(), Job::CUDA);

    JobResult cudaResult(cudaJob, 0x33333333, dummyHash);
    EXPECT_EQ(cudaResult.backend, Job::CUDA);
#endif
}

// Test difficulty scaling
TEST_F(MiningIntegrationTest, DifficultyScaling) {
    std::vector<uint64_t> difficulties = {
        1000,
        10000,
        100000,
        1000000,
        10000000
    };

    for (auto diff : difficulties) {
        Job job(false, Algorithm::RX_0, "diff-test");
        job.setDiff(diff);

        EXPECT_EQ(job.diff(), diff);

        uint8_t dummyHash[32] = {0};
        JobResult result(job, 0xAAAAAAAA, dummyHash);

        EXPECT_EQ(result.diff, diff);
    }
}

// Test client ID tracking through mining cycle
TEST_F(MiningIntegrationTest, ClientIdTracking) {
    const char* clientIds[] = {
        "pool1-client",
        "pool2-client",
        "pool3-client"
    };

    for (const char* clientId : clientIds) {
        Job job(false, Algorithm::RX_0, clientId);
        EXPECT_STREQ(job.clientId(), clientId);

        uint8_t dummyHash[32] = {0};
        JobResult result(job, 0xBBBBBBBB, dummyHash);

        EXPECT_STREQ(result.clientId, clientId);
    }
}

// Test empty job handling
TEST_F(MiningIntegrationTest, EmptyJobHandling) {
    Job emptyJob(false, Algorithm::INVALID, "");

    EXPECT_FALSE(emptyJob.algorithm().isValid());
    EXPECT_FALSE(emptyJob.isValid());
}

// Test nonce uniqueness in results
TEST_F(MiningIntegrationTest, NonceUniqueness) {
    Job job(false, Algorithm::RX_0, "nonce-test");
    job.setId("nonce-job");

    uint8_t dummyHash[32] = {0};
    std::vector<uint32_t> nonces = {
        0x00000001,
        0x00000002,
        0xFFFFFFFF,
        0x12345678,
        0xDEADBEEF
    };

    for (auto nonce : nonces) {
        JobResult result(job, nonce, dummyHash);
        EXPECT_EQ(result.nonce, nonce);
    }
}

// Test algorithm family consistency
TEST_F(MiningIntegrationTest, AlgorithmFamilyConsistency) {
    // RandomX family
    Algorithm rx0(Algorithm::RX_0);
    Algorithm rxWow(Algorithm::RX_WOW);

    EXPECT_EQ(rx0.family(), Algorithm::RANDOM_X);
    EXPECT_EQ(rxWow.family(), Algorithm::RANDOM_X);
    EXPECT_EQ(rx0.family(), rxWow.family());

    // CryptoNight family
    Algorithm cnR(Algorithm::CN_R);
    EXPECT_EQ(cnR.family(), Algorithm::CN);
}

} // namespace xmrig
