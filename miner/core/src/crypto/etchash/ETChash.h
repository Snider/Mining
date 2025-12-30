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

#ifndef XMRIG_ETCHASH_H
#define XMRIG_ETCHASH_H


#include <cstdint>


namespace xmrig
{


class ETCCache;


class ETChash
{
public:
    // ETChash constants
    // ECIP-1099: After epoch 390, epoch length changes to 60000 blocks
    static constexpr uint32_t EPOCH_LENGTH_OLD = 30000;  // Before ECIP-1099
    static constexpr uint32_t EPOCH_LENGTH_NEW = 60000;  // After ECIP-1099
    static constexpr uint32_t ECIP1099_ACTIVATION_EPOCH = 390;
    static constexpr uint32_t ECIP1099_ACTIVATION_BLOCK = 11700000;  // 390 * 30000

    // Ethash core constants (shared with libethash)
    static constexpr uint32_t MIX_BYTES = 128;
    static constexpr uint32_t HASH_BYTES = 64;
    static constexpr uint32_t DATASET_PARENTS = 256;
    static constexpr uint32_t CACHE_ROUNDS = 3;
    static constexpr uint32_t ACCESSES = 64;

    // Calculate epoch from block number (accounts for ECIP-1099)
    static uint32_t epoch(uint64_t blockNumber);

    // Calculate block number at start of epoch
    static uint64_t epochStartBlock(uint32_t epoch);

    // Calculate hash
    static void calculate(const ETCCache& cache, uint64_t blockNumber,
                         const uint8_t (&headerHash)[32], uint64_t nonce,
                         uint8_t (&output)[32], uint8_t (&mixHash)[32]);
};


// Ethash class - identical to ETChash but with standard epoch length
class Ethash
{
public:
    static constexpr uint32_t EPOCH_LENGTH = 30000;

    static uint32_t epoch(uint64_t blockNumber) { return static_cast<uint32_t>(blockNumber / EPOCH_LENGTH); }
    static uint64_t epochStartBlock(uint32_t epoch) { return static_cast<uint64_t>(epoch) * EPOCH_LENGTH; }

    static void calculate(const ETCCache& cache, uint64_t blockNumber,
                         const uint8_t (&headerHash)[32], uint64_t nonce,
                         uint8_t (&output)[32], uint8_t (&mixHash)[32]);
};


} // namespace xmrig


#endif // XMRIG_ETCHASH_H
