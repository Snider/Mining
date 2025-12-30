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


#include "crypto/blake3dcr/Blake3DCR.h"
#include "3rdparty/blake3/blake3.h"

#include <cstring>


namespace xmrig {


void Blake3DCR::hash(const uint8_t* header, size_t header_len, uint8_t* output)
{
    blake3_hash(header, header_len, output);
}


void Blake3DCR::calculate(const uint8_t* header, uint32_t nonce, uint8_t (&output)[32])
{
    // Copy header and insert nonce at the correct offset
    uint8_t work[BLOCK_HEADER_SIZE];
    memcpy(work, header, BLOCK_HEADER_SIZE);

    // Insert nonce (little-endian) at offset 140
    work[NONCE_OFFSET]     = static_cast<uint8_t>(nonce);
    work[NONCE_OFFSET + 1] = static_cast<uint8_t>(nonce >> 8);
    work[NONCE_OFFSET + 2] = static_cast<uint8_t>(nonce >> 16);
    work[NONCE_OFFSET + 3] = static_cast<uint8_t>(nonce >> 24);

    // Calculate Blake3 hash
    blake3_hash(work, BLOCK_HEADER_SIZE, output);
}


bool Blake3DCR::checkDifficulty(const uint8_t* hash, const uint8_t* target)
{
    // Compare hash with target (both are big-endian 256-bit numbers)
    // Hash must be less than or equal to target
    for (int i = 0; i < 32; i++) {
        if (hash[i] < target[i]) {
            return true;
        }
        if (hash[i] > target[i]) {
            return false;
        }
    }
    return true;
}


} // namespace xmrig
