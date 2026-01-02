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
#include "proxy/Counters.h"

// Note: Counters is a static class, no instantiation needed

class CountersTest : public ::testing::Test {
protected:
    void SetUp() override {
        Counters::reset();
    }

    void TearDown() override {
        Counters::reset();
    }
};

TEST_F(CountersTest, InitialStateIsZero) {
    EXPECT_EQ(Counters::miners(), 0);
    EXPECT_EQ(Counters::accepted.load(std::memory_order_relaxed), 0);
    EXPECT_EQ(Counters::added(), 0);
    EXPECT_EQ(Counters::removed(), 0);
}

TEST_F(CountersTest, IncrementMinerCount) {
    Counters::add();
    EXPECT_EQ(Counters::miners(), 1);

    Counters::add();
    EXPECT_EQ(Counters::miners(), 2);
}

TEST_F(CountersTest, DecrementMinerCount) {
    // Ensure we start from clean state
    uint64_t initial = Counters::miners();

    Counters::add();
    Counters::add();
    Counters::add();
    EXPECT_EQ(Counters::miners(), initial + 3);

    Counters::remove();
    EXPECT_EQ(Counters::miners(), initial + 2);

    Counters::remove();
    EXPECT_EQ(Counters::miners(), initial + 1);
}

TEST_F(CountersTest, AcceptedSharesIncrement) {
    Counters::accepted.fetch_add(1, std::memory_order_relaxed);
    EXPECT_EQ(Counters::accepted.load(std::memory_order_relaxed), 1);

    Counters::accepted.fetch_add(1, std::memory_order_relaxed);
    EXPECT_EQ(Counters::accepted.load(std::memory_order_relaxed), 2);
}

TEST_F(CountersTest, MixedOperations) {
    uint64_t initialMiners = Counters::miners();
    uint32_t initialAdded = Counters::added();
    uint32_t initialRemoved = Counters::removed();

    Counters::add();
    Counters::add();
    Counters::accepted.fetch_add(3, std::memory_order_relaxed);

    EXPECT_EQ(Counters::miners(), initialMiners + 2);
    EXPECT_EQ(Counters::accepted.load(std::memory_order_relaxed), 3);
    EXPECT_EQ(Counters::added(), initialAdded + 2);

    Counters::remove();
    EXPECT_EQ(Counters::miners(), initialMiners + 1);
    EXPECT_EQ(Counters::removed(), initialRemoved + 1);
}

TEST_F(CountersTest, MaxMinersTracking) {
    uint64_t initialMax = Counters::maxMiners();
    uint64_t initialMiners = Counters::miners();

    Counters::add();
    Counters::add();
    Counters::add();

    uint64_t newMax = Counters::maxMiners();
    EXPECT_GE(newMax, initialMiners + 3);  // Max should be at least current count

    Counters::remove();
    // Max should not decrease
    EXPECT_EQ(Counters::maxMiners(), newMax);
}

TEST_F(CountersTest, AddedRemovedCounters) {
    EXPECT_EQ(Counters::added(), 0);
    EXPECT_EQ(Counters::removed(), 0);

    Counters::add();
    EXPECT_EQ(Counters::added(), 1);

    Counters::add();
    EXPECT_EQ(Counters::added(), 2);

    Counters::remove();
    EXPECT_EQ(Counters::removed(), 1);
}
