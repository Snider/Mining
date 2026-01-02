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
#include "crypto/common/MemoryPool.h"
#include "crypto/common/VirtualMemory.h"
#include "base/crypto/Algorithm.h"

namespace xmrig {

class MemoryPoolTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Initialize with reasonable test size
    }

    void TearDown() override {
        // Cleanup handled by smart pointers
    }
};

// Test basic memory pool allocation
TEST_F(MemoryPoolTest, BasicAllocation) {
    MemoryPool pool;

    auto mem1 = pool.get(1024, 0);
    ASSERT_NE(mem1, nullptr) << "Failed to allocate memory from pool";

    auto mem2 = pool.get(1024, 0);
    ASSERT_NE(mem2, nullptr) << "Failed to allocate second memory from pool";

    // Verify different allocations
    EXPECT_NE(mem1, mem2) << "Pool returned same memory twice";
}

// Test memory pool reuse
TEST_F(MemoryPoolTest, MemoryReuse) {
    MemoryPool pool;

    auto mem1 = pool.get(1024, 0);
    ASSERT_NE(mem1, nullptr);

    uint8_t* ptr1 = mem1->scratchpad();

    // Release memory back to pool
    pool.release(mem1);

    // Get memory again - should reuse
    auto mem2 = pool.get(1024, 0);
    ASSERT_NE(mem2, nullptr);

    uint8_t* ptr2 = mem2->scratchpad();

    // Should be the same underlying memory
    EXPECT_EQ(ptr1, ptr2) << "Pool should reuse released memory";
}

// Test VirtualMemory allocation
TEST_F(MemoryPoolTest, VirtualMemoryAllocation) {
    const size_t size = 2 * 1024 * 1024; // 2 MB

    auto vm = new VirtualMemory(size, false, false, false, 0);
    ASSERT_NE(vm, nullptr) << "Failed to allocate VirtualMemory";

    EXPECT_GE(vm->size(), size) << "Allocated size should be at least requested size";
    EXPECT_NE(vm->scratchpad(), nullptr) << "Scratchpad pointer should not be null";

    // Write and read test
    uint8_t* ptr = vm->scratchpad();
    ptr[0] = 0x42;
    ptr[size - 1] = 0x24;

    EXPECT_EQ(ptr[0], 0x42) << "Memory should be readable/writable";
    EXPECT_EQ(ptr[size - 1], 0x24) << "Memory should be readable/writable at end";

    delete vm;
}

// Test alignment
TEST_F(MemoryPoolTest, MemoryAlignment) {
    const size_t size = 1024;

    auto vm = new VirtualMemory(size, false, false, false, 0);
    ASSERT_NE(vm, nullptr);

    uintptr_t addr = reinterpret_cast<uintptr_t>(vm->scratchpad());

    // Memory should be aligned to at least 16 bytes for crypto operations
    EXPECT_EQ(addr % 16, 0) << "Memory should be 16-byte aligned";

    delete vm;
}

// Test huge pages info
TEST_F(MemoryPoolTest, HugePagesInfo) {
    // Just verify we can query huge pages info without crashing
    VirtualMemory::init(0, 0);

    // Should not crash
    SUCCEED();
}

// Test multiple pool instances
TEST_F(MemoryPoolTest, MultiplePoolInstances) {
    MemoryPool pool1;
    MemoryPool pool2;

    auto mem1 = pool1.get(1024, 0);
    auto mem2 = pool2.get(1024, 0);

    ASSERT_NE(mem1, nullptr);
    ASSERT_NE(mem2, nullptr);

    // Different pools should give different memory
    EXPECT_NE(mem1, mem2) << "Different pools should allocate different memory";
}

// Test zero-size allocation handling
TEST_F(MemoryPoolTest, ZeroSizeAllocation) {
    MemoryPool pool;

    // Should handle gracefully (likely return nullptr or throw)
    auto mem = pool.get(0, 0);

    // Test passes if we don't crash - behavior may vary
    SUCCEED();
}

// Test large allocation
TEST_F(MemoryPoolTest, LargeAllocation) {
    const size_t largeSize = 256 * 1024 * 1024; // 256 MB

    // This might fail on systems with limited memory, but shouldn't crash
    auto vm = new VirtualMemory(largeSize, false, false, false, 0);

    if (vm != nullptr && vm->scratchpad() != nullptr) {
        EXPECT_GE(vm->size(), largeSize);
        delete vm;
    }

    // Test passes if we don't crash
    SUCCEED();
}

} // namespace xmrig
