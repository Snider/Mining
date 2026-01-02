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
#include "base/crypto/Algorithm.h"
#include "3rdparty/rapidjson/document.h"
#include "3rdparty/rapidjson/error/en.h"

namespace xmrig {

class ConfigTest : public ::testing::Test {
protected:
    void SetUp() override {
    }

    void TearDown() override {
    }

    // Helper to parse JSON
    bool parseJson(const char* json, rapidjson::Document& doc) {
        doc.Parse(json);
        return !doc.HasParseError();
    }
};

// Test JSON parsing - valid config
TEST_F(ConfigTest, ValidJsonParsing) {
    const char* validJson = R"({
        "algo": "rx/0",
        "pool": "pool.example.com:3333",
        "user": "wallet123",
        "pass": "x"
    })";

    rapidjson::Document doc;
    EXPECT_TRUE(parseJson(validJson, doc));
    EXPECT_TRUE(doc.IsObject());
}

// Test JSON parsing - invalid JSON
TEST_F(ConfigTest, InvalidJsonParsing) {
    const char* invalidJson = R"({
        "algo": "rx/0",
        "pool": "pool.example.com:3333"
        "user": "wallet123"
    })";

    rapidjson::Document doc;
    EXPECT_FALSE(parseJson(invalidJson, doc));
}

// Test algorithm parsing
TEST_F(ConfigTest, AlgorithmParsing) {
    const char* testJson = R"({"algo": "rx/0"})";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    if (doc.HasMember("algo") && doc["algo"].IsString()) {
        Algorithm algo(doc["algo"].GetString());
        EXPECT_TRUE(algo.isValid());
        EXPECT_EQ(algo.id(), Algorithm::RX_0);
    }
}

// Test multiple pool configuration
TEST_F(ConfigTest, MultiplePoolsParsing) {
    const char* testJson = R"({
        "pools": [
            {"url": "pool1.example.com:3333", "user": "wallet1"},
            {"url": "pool2.example.com:3333", "user": "wallet2"}
        ]
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("pools"));
    EXPECT_TRUE(doc["pools"].IsArray());
    EXPECT_EQ(doc["pools"].Size(), 2);
}

// Test CPU configuration
TEST_F(ConfigTest, CpuConfigParsing) {
    const char* testJson = R"({
        "cpu": {
            "enabled": true,
            "max-threads-hint": 50,
            "priority": 5
        }
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("cpu"));
    EXPECT_TRUE(doc["cpu"].IsObject());

    if (doc["cpu"].HasMember("enabled")) {
        EXPECT_TRUE(doc["cpu"]["enabled"].IsBool());
        EXPECT_TRUE(doc["cpu"]["enabled"].GetBool());
    }

    if (doc["cpu"].HasMember("max-threads-hint")) {
        EXPECT_TRUE(doc["cpu"]["max-threads-hint"].IsInt());
        EXPECT_EQ(doc["cpu"]["max-threads-hint"].GetInt(), 50);
    }
}

// Test OpenCL configuration
TEST_F(ConfigTest, OpenCLConfigParsing) {
    const char* testJson = R"({
        "opencl": {
            "enabled": true,
            "platform": 0
        }
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("opencl"));
    EXPECT_TRUE(doc["opencl"].IsObject());

    if (doc["opencl"].HasMember("enabled")) {
        EXPECT_TRUE(doc["opencl"]["enabled"].IsBool());
    }
}

// Test CUDA configuration
TEST_F(ConfigTest, CudaConfigParsing) {
    const char* testJson = R"({
        "cuda": {
            "enabled": true,
            "loader": "xmrig-cuda.dll"
        }
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("cuda"));
    EXPECT_TRUE(doc["cuda"].IsObject());

    if (doc["cuda"].HasMember("loader")) {
        EXPECT_TRUE(doc["cuda"]["loader"].IsString());
        EXPECT_STREQ(doc["cuda"]["loader"].GetString(), "xmrig-cuda.dll");
    }
}

// Test API configuration
TEST_F(ConfigTest, ApiConfigParsing) {
    const char* testJson = R"({
        "api": {
            "enabled": true,
            "port": 8080,
            "access-token": "secret123"
        }
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("api"));
    EXPECT_TRUE(doc["api"].IsObject());

    if (doc["api"].HasMember("port")) {
        EXPECT_TRUE(doc["api"]["port"].IsInt());
        EXPECT_EQ(doc["api"]["port"].GetInt(), 8080);
    }

    if (doc["api"].HasMember("access-token")) {
        EXPECT_TRUE(doc["api"]["access-token"].IsString());
        EXPECT_STREQ(doc["api"]["access-token"].GetString(), "secret123");
    }
}

// Test RandomX configuration
TEST_F(ConfigTest, RandomXConfigParsing) {
    const char* testJson = R"({
        "randomx": {
            "init": -1,
            "mode": "auto",
            "1gb-pages": true,
            "numa": true
        }
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc.HasMember("randomx"));
    EXPECT_TRUE(doc["randomx"].IsObject());

    if (doc["randomx"].HasMember("mode")) {
        EXPECT_TRUE(doc["randomx"]["mode"].IsString());
        EXPECT_STREQ(doc["randomx"]["mode"].GetString(), "auto");
    }

    if (doc["randomx"].HasMember("1gb-pages")) {
        EXPECT_TRUE(doc["randomx"]["1gb-pages"].IsBool());
        EXPECT_TRUE(doc["randomx"]["1gb-pages"].GetBool());
    }
}

// Test logging configuration
TEST_F(ConfigTest, LogConfigParsing) {
    const char* testJson = R"({
        "log-file": "/var/log/miner.log",
        "syslog": true,
        "colors": true
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    if (doc.HasMember("log-file")) {
        EXPECT_TRUE(doc["log-file"].IsString());
        EXPECT_STREQ(doc["log-file"].GetString(), "/var/log/miner.log");
    }

    if (doc.HasMember("syslog")) {
        EXPECT_TRUE(doc["syslog"].IsBool());
    }

    if (doc.HasMember("colors")) {
        EXPECT_TRUE(doc["colors"].IsBool());
    }
}

// Test boolean value validation
TEST_F(ConfigTest, BooleanValidation) {
    const char* testJson = R"({
        "test_true": true,
        "test_false": false
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc["test_true"].IsBool());
    EXPECT_TRUE(doc["test_true"].GetBool());

    EXPECT_TRUE(doc["test_false"].IsBool());
    EXPECT_FALSE(doc["test_false"].GetBool());
}

// Test integer value validation
TEST_F(ConfigTest, IntegerValidation) {
    const char* testJson = R"({
        "positive": 100,
        "negative": -50,
        "zero": 0
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc["positive"].IsInt());
    EXPECT_EQ(doc["positive"].GetInt(), 100);

    EXPECT_TRUE(doc["negative"].IsInt());
    EXPECT_EQ(doc["negative"].GetInt(), -50);

    EXPECT_TRUE(doc["zero"].IsInt());
    EXPECT_EQ(doc["zero"].GetInt(), 0);
}

// Test string value validation
TEST_F(ConfigTest, StringValidation) {
    const char* testJson = R"({
        "empty": "",
        "normal": "test string",
        "special": "test\nwith\ttabs"
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc["empty"].IsString());
    EXPECT_STREQ(doc["empty"].GetString(), "");

    EXPECT_TRUE(doc["normal"].IsString());
    EXPECT_STREQ(doc["normal"].GetString(), "test string");

    EXPECT_TRUE(doc["special"].IsString());
}

// Test array validation
TEST_F(ConfigTest, ArrayValidation) {
    const char* testJson = R"({
        "empty_array": [],
        "int_array": [1, 2, 3],
        "string_array": ["a", "b", "c"]
    })";

    rapidjson::Document doc;
    ASSERT_TRUE(parseJson(testJson, doc));

    EXPECT_TRUE(doc["empty_array"].IsArray());
    EXPECT_EQ(doc["empty_array"].Size(), 0);

    EXPECT_TRUE(doc["int_array"].IsArray());
    EXPECT_EQ(doc["int_array"].Size(), 3);

    EXPECT_TRUE(doc["string_array"].IsArray());
    EXPECT_EQ(doc["string_array"].Size(), 3);
}

} // namespace xmrig
