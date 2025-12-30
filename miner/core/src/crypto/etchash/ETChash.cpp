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


#include "crypto/etchash/ETChash.h"
#include "crypto/etchash/ETCCache.h"
#include "3rdparty/libethash/ethash.h"
#include "3rdparty/libethash/ethash_internal.h"
#include "3rdparty/libethash/data_sizes.h"


namespace xmrig {


// ECIP-1099: Calculate epoch from block number for Ethereum Classic
// Before activation: epoch = block / 30000
// After activation:  epoch = 390 + (block - 11700000) / 60000
uint32_t ETChash::epoch(uint64_t blockNumber)
{
    if (blockNumber < ECIP1099_ACTIVATION_BLOCK) {
        return static_cast<uint32_t>(blockNumber / EPOCH_LENGTH_OLD);
    }

    // After ECIP-1099 activation, epoch increases every 60000 blocks
    return ECIP1099_ACTIVATION_EPOCH +
           static_cast<uint32_t>((blockNumber - ECIP1099_ACTIVATION_BLOCK) / EPOCH_LENGTH_NEW);
}


uint64_t ETChash::epochStartBlock(uint32_t epoch)
{
    if (epoch < ECIP1099_ACTIVATION_EPOCH) {
        return static_cast<uint64_t>(epoch) * EPOCH_LENGTH_OLD;
    }

    // After ECIP-1099
    return ECIP1099_ACTIVATION_BLOCK +
           static_cast<uint64_t>(epoch - ECIP1099_ACTIVATION_EPOCH) * EPOCH_LENGTH_NEW;
}


void ETChash::calculate(const ETCCache& cache, uint64_t blockNumber,
                        const uint8_t (&headerHash)[32], uint64_t nonce,
                        uint8_t (&output)[32], uint8_t (&mixHash)[32])
{
    const uint32_t epochNum = cache.epoch();

    // Get DAG size for this epoch
    const uint64_t fullSize = dag_sizes[epochNum];

    // Setup light cache structure for libethash
    ethash_light lightCache;
    lightCache.cache = cache.data();
    lightCache.cache_size = cache.size();
    lightCache.block_number = blockNumber;

    // Calculate fast mod data for optimized DAG item calculation
    lightCache.num_parent_nodes = static_cast<uint32_t>(cache.size() / sizeof(node));

    // Calculate reciprocal, increment, shift for fast modulo
    uint32_t divisor = lightCache.num_parent_nodes;
    if ((divisor & (divisor - 1)) == 0) {
        // Power of 2
        lightCache.reciprocal = 1;
        lightCache.increment = 0;
        uint32_t shift = 0;
        uint32_t temp = divisor;
        while (temp > 1) {
            temp >>= 1;
            shift++;
        }
        lightCache.shift = shift;
    } else {
        // Use fast division algorithm
        uint32_t shift = 31;
        uint32_t temp = divisor;
        while (temp > 0) {
            temp >>= 1;
            if (temp > 0) shift++;
        }
        shift = 63 - (31 - shift);

        const uint64_t N = 1ULL << shift;
        const uint64_t q = N / divisor;
        const uint64_t r = N - q * divisor;

        if (r * 2 < divisor) {
            lightCache.reciprocal = static_cast<uint32_t>(q);
            lightCache.increment = 1;
        } else {
            lightCache.reciprocal = static_cast<uint32_t>(q + 1);
            lightCache.increment = 0;
        }
        lightCache.shift = shift;
    }

    // Convert header hash to libethash format
    ethash_h256_t header;
    memcpy(header.b, headerHash, 32);

    // Compute the Ethash using light client verification
    ethash_return_value_t result = ethash_light_compute_internal(&lightCache, fullSize, header, nonce);

    // Copy results
    memcpy(output, result.result.b, 32);
    memcpy(mixHash, result.mix_hash.b, 32);
}


// Ethash (standard Ethereum) - uses fixed 30000 block epochs
void Ethash::calculate(const ETCCache& cache, uint64_t blockNumber,
                       const uint8_t (&headerHash)[32], uint64_t nonce,
                       uint8_t (&output)[32], uint8_t (&mixHash)[32])
{
    const uint32_t epochNum = cache.epoch();

    // Get DAG size for this epoch
    const uint64_t fullSize = dag_sizes[epochNum];

    // Setup light cache structure for libethash
    ethash_light lightCache;
    lightCache.cache = cache.data();
    lightCache.cache_size = cache.size();
    lightCache.block_number = blockNumber;

    // Calculate fast mod data
    lightCache.num_parent_nodes = static_cast<uint32_t>(cache.size() / sizeof(node));

    uint32_t divisor = lightCache.num_parent_nodes;
    if ((divisor & (divisor - 1)) == 0) {
        lightCache.reciprocal = 1;
        lightCache.increment = 0;
        uint32_t shift = 0;
        uint32_t temp = divisor;
        while (temp > 1) {
            temp >>= 1;
            shift++;
        }
        lightCache.shift = shift;
    } else {
        uint32_t shift = 31;
        uint32_t temp = divisor;
        while (temp > 0) {
            temp >>= 1;
            if (temp > 0) shift++;
        }
        shift = 63 - (31 - shift);

        const uint64_t N = 1ULL << shift;
        const uint64_t q = N / divisor;
        const uint64_t r = N - q * divisor;

        if (r * 2 < divisor) {
            lightCache.reciprocal = static_cast<uint32_t>(q);
            lightCache.increment = 1;
        } else {
            lightCache.reciprocal = static_cast<uint32_t>(q + 1);
            lightCache.increment = 0;
        }
        lightCache.shift = shift;
    }

    ethash_h256_t header;
    memcpy(header.b, headerHash, 32);

    ethash_return_value_t result = ethash_light_compute_internal(&lightCache, fullSize, header, nonce);

    memcpy(output, result.result.b, 32);
    memcpy(mixHash, result.mix_hash.b, 32);
}


} // namespace xmrig
