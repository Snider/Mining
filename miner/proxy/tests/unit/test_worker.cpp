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
#include "proxy/workers/Worker.h"

using namespace xmrig;

class WorkerTest : public ::testing::Test {
protected:
    void SetUp() override {
        worker = new Worker(1, "test_worker", "127.0.0.1");
    }

    void TearDown() override {
        delete worker;
    }

    Worker* worker = nullptr;
};

TEST_F(WorkerTest, InitialState) {
    EXPECT_STREQ(worker->name(), "test_worker");
    EXPECT_STREQ(worker->ip(), "127.0.0.1");
    EXPECT_EQ(worker->id(), 1);
    EXPECT_EQ(worker->connections(), 1);  // Worker starts with 1 connection
    EXPECT_EQ(worker->accepted(), 0);
    EXPECT_EQ(worker->rejected(), 0);
    EXPECT_EQ(worker->invalid(), 0);
}

TEST_F(WorkerTest, AddConnection) {
    // Worker starts with 1 connection already
    EXPECT_EQ(worker->connections(), 1);

    worker->add("192.168.1.1");
    EXPECT_EQ(worker->connections(), 2);
    EXPECT_STREQ(worker->ip(), "192.168.1.1");

    worker->add("192.168.1.2");
    EXPECT_EQ(worker->connections(), 3);
}

TEST_F(WorkerTest, RemoveConnection) {
    // Worker starts with 1 connection
    worker->add("192.168.1.1");
    worker->add("192.168.1.2");
    worker->add("192.168.1.3");
    EXPECT_EQ(worker->connections(), 4);

    worker->remove();
    EXPECT_EQ(worker->connections(), 3);
}

TEST_F(WorkerTest, AcceptedShares) {
    // add(diff) increments accepted by 1, and adds diff to hashes
    worker->add(1000);
    EXPECT_EQ(worker->accepted(), 1);  // accepted increments by 1
    EXPECT_EQ(worker->hashes(), 1000); // hashes gets the difficulty

    worker->add(5000);
    EXPECT_EQ(worker->accepted(), 2);   // accepted increments by 1 again
    EXPECT_EQ(worker->hashes(), 6000);  // hashes accumulates
}

TEST_F(WorkerTest, RejectedShares) {
    worker->reject(false);  // false = rejected (not invalid)
    EXPECT_EQ(worker->rejected(), 1);
    EXPECT_EQ(worker->invalid(), 0);

    worker->reject(false);
    EXPECT_EQ(worker->rejected(), 2);
}

TEST_F(WorkerTest, InvalidShares) {
    worker->reject(true);  // true = invalid
    EXPECT_EQ(worker->invalid(), 1);
    EXPECT_EQ(worker->rejected(), 0);

    worker->reject(true);
    EXPECT_EQ(worker->invalid(), 2);
}

TEST_F(WorkerTest, MixedOperations) {
    // Worker starts with 1 connection
    worker->add("192.168.1.1");
    worker->add("192.168.1.2");
    worker->add(10000);  // accepted increments by 1, hashes gets 10000
    worker->reject(false);  // rejected
    worker->reject(true);   // invalid

    EXPECT_EQ(worker->connections(), 3);  // 1 initial + 2 added
    EXPECT_EQ(worker->accepted(), 1);     // add() increments by 1
    EXPECT_EQ(worker->hashes(), 10000);   // difficulty goes here
    EXPECT_EQ(worker->rejected(), 1);
    EXPECT_EQ(worker->invalid(), 1);

    worker->remove();
    EXPECT_EQ(worker->connections(), 2);
}

TEST_F(WorkerTest, EmptyWorkerName) {
    Worker emptyWorker(2, "", "10.0.0.1");
    EXPECT_STREQ(emptyWorker.name(), "");
}

TEST_F(WorkerTest, LongWorkerName) {
    std::string longName(1000, 'x');
    Worker longWorker(3, longName, "10.0.0.2");
    EXPECT_STREQ(longWorker.name(), longName.c_str());
}

TEST_F(WorkerTest, HashrateCalculation) {
    // Hashrate requires ticking and time
    EXPECT_EQ(worker->hashrate(10), 0.0);
}
