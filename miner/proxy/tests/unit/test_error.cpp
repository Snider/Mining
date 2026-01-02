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

#include <gtest/gtest.h>
#include "proxy/Error.h"

using namespace xmrig;

TEST(ErrorTest, ErrorMessages) {
    EXPECT_NE(Error::toString(Error::NoError), nullptr);
    EXPECT_NE(Error::toString(Error::BadGateway), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidJobId), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidMethod), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidNonce), nullptr);
}

TEST(ErrorTest, InvalidErrorCode) {
    // Test boundary conditions
    const char* msg = Error::toString(999);
    EXPECT_NE(msg, nullptr);  // Should return something, not crash
}

TEST(ErrorTest, AllErrorCodesHaveMessages) {
    // Ensure all defined error codes have non-null messages
    EXPECT_NE(Error::toString(Error::NoError), nullptr);
    EXPECT_NE(Error::toString(Error::BadGateway), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidJobId), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidMethod), nullptr);
    EXPECT_NE(Error::toString(Error::InvalidNonce), nullptr);
    EXPECT_NE(Error::toString(Error::LowDifficulty), nullptr);
    EXPECT_NE(Error::toString(Error::Unauthenticated), nullptr);
    EXPECT_NE(Error::toString(Error::IncompatibleAlgorithm), nullptr);
    EXPECT_NE(Error::toString(Error::IncorrectAlgorithm), nullptr);
    EXPECT_NE(Error::toString(Error::Forbidden), nullptr);
    EXPECT_NE(Error::toString(Error::RouteNotFound), nullptr);
}

TEST(ErrorTest, ErrorMessagesAreNotEmpty) {
    // Most errors should have non-empty messages
    EXPECT_GT(strlen(Error::toString(Error::BadGateway)), 0);
    EXPECT_GT(strlen(Error::toString(Error::InvalidJobId)), 0);
    EXPECT_GT(strlen(Error::toString(Error::InvalidNonce)), 0);
}
