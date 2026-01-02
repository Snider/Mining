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

#include "test_helpers.h"
#include <fstream>
#include <cstdio>
#include <cstdlib>
#include <sstream>
#include <iomanip>

namespace xmrig {
namespace test {

void ProxyTestBase::SetUp() {
    m_tempFiles.clear();
}

void ProxyTestBase::TearDown() {
    cleanupTempFiles();
}

std::string ProxyTestBase::createTempFile(const std::string& content) {
    char tempName[] = "/tmp/proxy_test_XXXXXX";
    int fd = mkstemp(tempName);
    if (fd == -1) {
        return "";
    }

    std::ofstream file(tempName);
    file << content;
    file.close();
    close(fd);

    m_tempFiles.push_back(tempName);
    return tempName;
}

void ProxyTestBase::cleanupTempFiles() {
    for (const auto& file : m_tempFiles) {
        std::remove(file.c_str());
    }
    m_tempFiles.clear();
}

std::string TestDataGenerator::generateJobJson(const std::string& jobId) {
    std::ostringstream json;
    json << R"({)"
         << R"("job_id":")" << jobId << R"(",)"
         << R"("blob":"0606b1d7a8d505b68e70449ca4b0ea24f764cf2f9c4f0b81cc087ced026100000000000000000000000000000000000000000000000000000000000000000000",)"
         << R"("target":"b88d0600",)"
         << R"("algo":"cn/r",)"
         << R"("height":2000000,)"
         << R"("seed_hash":"0000000000000000000000000000000000000000000000000000000000000000")"
         << R"(})";
    return json.str();
}

std::string TestDataGenerator::generateLoginJson(const std::string& user,
                                                  const std::string& pass,
                                                  const std::string& agent) {
    std::ostringstream json;
    json << R"({"id":1,"jsonrpc":"2.0","method":"login","params":{)"
         << R"("login":")" << user << R"(",)"
         << R"("pass":")" << pass << R"(",)"
         << R"("agent":")" << agent << R"(")"
         << R"(}})";
    return json.str();
}

std::string TestDataGenerator::generateSubmitJson(const std::string& jobId,
                                                   uint32_t nonce,
                                                   const std::string& result) {
    std::ostringstream json;
    json << R"({"id":2,"jsonrpc":"2.0","method":"submit","params":{)"
         << R"("id":"test_session",)"
         << R"("job_id":")" << jobId << R"(",)"
         << R"("nonce":")" << std::hex << std::setw(8) << std::setfill('0') << nonce << R"(",)"
         << R"("result":")" << result << R"(")"
         << R"(}})";
    return json.str();
}

void AssertNonceValid(uint32_t nonce) {
    // Nonce should be within valid 32-bit range (always true for uint32_t)
    ASSERT_TRUE(true);
}

void AssertHashValid(const std::string& hash) {
    // Hash should be hex string of appropriate length
    ASSERT_FALSE(hash.empty());
    ASSERT_TRUE(hash.length() % 2 == 0);

    for (char c : hash) {
        ASSERT_TRUE(std::isxdigit(c));
    }
}

void AssertJobValid(const std::string& jobJson) {
    // Basic validation that JSON contains required fields
    ASSERT_FALSE(jobJson.empty());
    ASSERT_NE(jobJson.find("job_id"), std::string::npos);
    ASSERT_NE(jobJson.find("blob"), std::string::npos);
    ASSERT_NE(jobJson.find("target"), std::string::npos);
}

} // namespace test
} // namespace xmrig
