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


#include <cinttypes>

#include "crypto/etchash/ETCCache.h"
#include "3rdparty/libethash/data_sizes.h"
#include "3rdparty/libethash/ethash_internal.h"
#include "3rdparty/libethash/ethash.h"
#include "base/io/log/Log.h"
#include "base/io/log/Tags.h"
#include "base/tools/Chrono.h"
#include "crypto/common/VirtualMemory.h"


namespace xmrig {


std::mutex ETCCache::s_cacheMutex;
ETCCache ETCCache::s_etcCache;
ETCCache ETCCache::s_ethCache;


ETCCache::ETCCache()
{
}


ETCCache::~ETCCache()
{
    delete m_memory;
}


bool ETCCache::init(uint32_t epoch, bool isETC)
{
    if (epoch >= sizeof(cache_sizes) / sizeof(cache_sizes[0])) {
        return false;
    }

    if (m_epoch == epoch && m_isETC == isETC) {
        return true;
    }

    const uint64_t start_ms = Chrono::steadyMSecs();

    const size_t size = cache_sizes[epoch];
    if (!m_memory || m_memory->size() < size) {
        delete m_memory;
        m_memory = new VirtualMemory(size, false, false, false);
    }

    // Calculate seed hash for this epoch
    uint8_t seed[32];
    seedHash(epoch, seed);

    ethash_h256_t seedhash;
    memcpy(seedhash.b, seed, 32);

    ethash_compute_cache_nodes(m_memory->raw(), size, &seedhash);

    m_size = size;
    m_epoch = epoch;
    m_isETC = isETC;

    const char* algoName = isETC ? "ETChash" : "Ethash";
    LOG_INFO("%s " YELLOW("%s") " light cache for epoch " WHITE_BOLD("%u") " calculated " BLACK_BOLD("(%" PRIu64 "ms)"),
             Tags::miner(), algoName, epoch, Chrono::steadyMSecs() - start_ms);

    return true;
}


void* ETCCache::data() const
{
    return m_memory ? m_memory->raw() : nullptr;
}


uint64_t ETCCache::cacheSize(uint32_t epoch)
{
    if (epoch >= sizeof(cache_sizes) / sizeof(cache_sizes[0])) {
        return 0;
    }

    return cache_sizes[epoch];
}


uint64_t ETCCache::dagSize(uint32_t epoch)
{
    if (epoch >= sizeof(dag_sizes) / sizeof(dag_sizes[0])) {
        return 0;
    }

    return dag_sizes[epoch];
}


void ETCCache::seedHash(uint32_t epoch, uint8_t (&seed)[32])
{
    // Seed hash starts as zeros
    memset(seed, 0, 32);

    // Each epoch, seed = keccak256(previous_seed)
    for (uint32_t i = 0; i < epoch; ++i) {
        ethash_h256_t hash;
        memcpy(hash.b, seed, 32);
        hash = ethash_get_seedhash(i + 1);
        memcpy(seed, hash.b, 32);
    }

    // Actually just use libethash's function directly
    ethash_h256_t hash = ethash_get_seedhash(epoch);
    memcpy(seed, hash.b, 32);
}


} // namespace xmrig
