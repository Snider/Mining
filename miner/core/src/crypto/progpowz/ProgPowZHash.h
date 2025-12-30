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

#ifndef XMRIG_PROGPOWZ_HASH_H
#define XMRIG_PROGPOWZ_HASH_H


#include <cstdint>


namespace xmrig
{


class ProgPowZCache;


class ProgPowZHash
{
public:
    // ProgPowZ uses standard Ethash epoch length (30000 blocks)
    static constexpr uint32_t EPOCH_LENGTH  = 30000;

    // ProgPowZ period - blocks before changing the random program
    // Zano uses 50 (vs 3 for KawPow)
    static constexpr uint32_t PERIOD_LENGTH = 50;

    // ProgPowZ algorithm parameters
    static constexpr int CNT_CACHE          = 12;   // vs 11 for KawPow
    static constexpr int CNT_MATH           = 20;   // vs 18 for KawPow
    static constexpr uint32_t REGS          = 32;
    static constexpr uint32_t LANES         = 16;
    static constexpr uint32_t DAG_LOADS     = 4;
    static constexpr uint32_t CNT_DAG       = 64;
    static constexpr size_t CACHE_BYTES     = 16384;

    static void calculate(const ProgPowZCache& light_cache, uint32_t block_height, const uint8_t (&header_hash)[32], uint64_t nonce, uint32_t (&output)[8], uint32_t (&mix_hash)[8]);
};


} // namespace xmrig


#endif // XMRIG_PROGPOWZ_HASH_H
