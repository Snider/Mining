/* Miner
 * Copyright (c) 2025 Lethean
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

#ifndef XMRIG_ETC_CACHE_H
#define XMRIG_ETC_CACHE_H


#include "base/tools/Object.h"
#include <mutex>
#include <vector>
#include <cstdint>


namespace xmrig
{


class VirtualMemory;


class ETCCache
{
public:
    // Ethash cache item size = 64 bytes (HASH_BYTES)
    static constexpr size_t HASH_BYTES = 64;

    XMRIG_DISABLE_COPY_MOVE(ETCCache)

    ETCCache();
    ~ETCCache();

    // Initialize cache for given epoch
    bool init(uint32_t epoch, bool isETC = true);

    // Access cache data
    void* data() const;
    size_t size() const { return m_size; }
    uint32_t epoch() const { return m_epoch; }
    bool isETC() const { return m_isETC; }

    // Calculate cache and DAG sizes for epoch
    static uint64_t cacheSize(uint32_t epoch);
    static uint64_t dagSize(uint32_t epoch);

    // Get seed hash for epoch
    static void seedHash(uint32_t epoch, uint8_t (&seed)[32]);

    // Singleton instances
    static std::mutex s_cacheMutex;
    static ETCCache s_etcCache;   // For ETC (ETChash)
    static ETCCache s_ethCache;   // For ETH (Ethash)

private:
    VirtualMemory* m_memory = nullptr;
    size_t m_size = 0;
    uint32_t m_epoch = 0xFFFFFFFFUL;
    bool m_isETC = true;
};


} /* namespace xmrig */


#endif /* XMRIG_ETC_CACHE_H */
