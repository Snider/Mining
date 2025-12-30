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

#ifndef XMRIG_BLAKE3DCR_H
#define XMRIG_BLAKE3DCR_H


#include <cstdint>
#include <cstddef>


namespace xmrig
{


class Blake3DCR
{
public:
    // Blake3DCR constants for Decred mining
    // Decred block header is 180 bytes
    static constexpr size_t BLOCK_HEADER_SIZE = 180;
    static constexpr size_t HASH_SIZE = 32;

    // Nonce position in Decred block header (bytes 140-143)
    static constexpr size_t NONCE_OFFSET = 140;

    // Calculate Blake3 hash of block header
    static void hash(const uint8_t* header, size_t header_len, uint8_t* output);

    // Mining function: tries nonce and returns hash
    static void calculate(const uint8_t* header, uint32_t nonce, uint8_t (&output)[32]);

    // Check if hash meets difficulty target
    static bool checkDifficulty(const uint8_t* hash, const uint8_t* target);
};


} // namespace xmrig


#endif // XMRIG_BLAKE3DCR_H
