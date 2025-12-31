/* XMRig
 * Copyright 2010      Jeff Garzik <jgarzik@pobox.com>
 * Copyright 2012-2014 pooler      <pooler@litecoinpool.org>
 * Copyright 2014      Lucas Jones <https://github.com/lucasjones>
 * Copyright 2014-2016 Wolf9466    <https://github.com/OhGodAPet>
 * Copyright 2016      Jay D Dee   <jayddee246@gmail.com>
 * Copyright 2016-2017 XMRig       <support@xmrig.com>
 *
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

#ifndef __COUNTERS_H__
#define __COUNTERS_H__


#include <atomic>
#include <cstdint>


// THREAD SAFETY FIX: All counters are now atomic to prevent race conditions
// with 100K+ concurrent miner connections
class Counters
{
public:
    static inline void reset()
    {
        m_added.store(0, std::memory_order_relaxed);
        m_removed.store(0, std::memory_order_relaxed);
        accepted.store(0, std::memory_order_relaxed);
    }


    static inline void add()
    {
        uint64_t current = m_miners.fetch_add(1, std::memory_order_relaxed) + 1;
        m_added.fetch_add(1, std::memory_order_relaxed);

        // Thread-safe max update using compare-and-swap
        uint64_t maxVal = m_maxMiners.load(std::memory_order_relaxed);
        while (current > maxVal && !m_maxMiners.compare_exchange_weak(maxVal, current, std::memory_order_relaxed)) {
            // maxVal is updated by compare_exchange_weak on failure
        }
    }


    static inline void remove()
    {
        m_miners.fetch_sub(1, std::memory_order_relaxed);
        m_removed.fetch_add(1, std::memory_order_relaxed);
    }


    static inline uint32_t added()     { return m_added.load(std::memory_order_relaxed); }
    static inline uint32_t removed()   { return m_removed.load(std::memory_order_relaxed); }
    static inline uint64_t maxMiners() { return m_maxMiners.load(std::memory_order_relaxed); }
    static inline uint64_t miners()    { return m_miners.load(std::memory_order_relaxed); }

    static std::atomic<uint64_t> accepted;
    static std::atomic<uint64_t> connections;
    static std::atomic<uint64_t> expired;

private:
    static std::atomic<uint32_t> m_added;
    static std::atomic<uint32_t> m_removed;
    static std::atomic<uint64_t> m_maxMiners;
    static std::atomic<uint64_t> m_miners;
};

#endif /* __COUNTERS_H__ */
