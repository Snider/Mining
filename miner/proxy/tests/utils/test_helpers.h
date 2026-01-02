/* XMRig
 * Copyright 2016-2021 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
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

#ifndef XMRIG_TEST_HELPERS_H
#define XMRIG_TEST_HELPERS_H

#include <gtest/gtest.h>
#include <string>
#include <vector>

namespace xmrig {
namespace test {

/**
 * Test fixture base class that provides common utilities
 */
class ProxyTestBase : public ::testing::Test {
protected:
    void SetUp() override;
    void TearDown() override;

    // Helper to create temporary test files
    std::string createTempFile(const std::string& content);

    // Helper to clean up temp files
    void cleanupTempFiles();

private:
    std::vector<std::string> m_tempFiles;
};

/**
 * Mock Controller for testing components that depend on it
 */
class MockController {
public:
    MockController() = default;
    virtual ~MockController() = default;

    // Add mock methods as needed
};

/**
 * Test data generators
 */
class TestDataGenerator {
public:
    // Generate valid stratum job JSON
    static std::string generateJobJson(const std::string& jobId = "test_job");

    // Generate valid login request JSON
    static std::string generateLoginJson(const std::string& user = "testuser",
                                         const std::string& pass = "x",
                                         const std::string& agent = "test/1.0");

    // Generate valid submit request JSON
    static std::string generateSubmitJson(const std::string& jobId = "test_job",
                                          uint32_t nonce = 0x12345678,
                                          const std::string& result = "0123456789abcdef");
};

/**
 * Assertion helpers for crypto mining specific checks
 */
void AssertNonceValid(uint32_t nonce);
void AssertHashValid(const std::string& hash);
void AssertJobValid(const std::string& jobJson);

} // namespace test
} // namespace xmrig

#endif /* XMRIG_TEST_HELPERS_H */
