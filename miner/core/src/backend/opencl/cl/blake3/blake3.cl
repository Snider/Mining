/* Miner
 * Copyright (c) 2025 Lethean
 *
 *   Blake3 OpenCL mining kernel for Decred
 *   Based on BLAKE3 reference implementation
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 */

//
// Blake3 constants
//

#define BLAKE3_BLOCK_LEN 64
#define BLAKE3_OUT_LEN 32

// Flags
#define CHUNK_START 1
#define CHUNK_END 2
#define ROOT 8

// Initial vector
#define IV_0 0x6A09E667
#define IV_1 0xBB67AE85
#define IV_2 0x3C6EF372
#define IV_3 0xA54FF53A
#define IV_4 0x510E527F
#define IV_5 0x9B05688C
#define IV_6 0x1F83D9AB
#define IV_7 0x5BE0CD19

// Decred block header constants
#define BLOCK_HEADER_SIZE 180
#define NONCE_OFFSET 140

#ifndef GROUP_SIZE
#define GROUP_SIZE 256
#endif

// Message schedule for 7 rounds
__constant uchar MSG_SCHEDULE[7][16] = {
    {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
    {2, 6, 3, 10, 7, 0, 4, 13, 1, 11, 12, 5, 9, 14, 15, 8},
    {3, 4, 10, 12, 13, 2, 7, 14, 6, 5, 9, 0, 11, 15, 8, 1},
    {10, 7, 12, 9, 14, 3, 13, 15, 4, 0, 11, 2, 5, 8, 1, 6},
    {12, 13, 9, 11, 15, 10, 14, 8, 7, 2, 5, 3, 0, 1, 6, 4},
    {9, 14, 11, 5, 8, 12, 15, 1, 13, 3, 0, 10, 2, 6, 4, 7},
    {11, 15, 5, 0, 1, 9, 8, 6, 14, 10, 2, 12, 3, 4, 7, 13},
};

static inline uint rotr32(uint w, uint c)
{
    return (w >> c) | (w << (32 - c));
}

// Quarter round
static inline void g(uint *state, uint a, uint b, uint c, uint d, uint x, uint y)
{
    state[a] = state[a] + state[b] + x;
    state[d] = rotr32(state[d] ^ state[a], 16);
    state[c] = state[c] + state[d];
    state[b] = rotr32(state[b] ^ state[c], 12);
    state[a] = state[a] + state[b] + y;
    state[d] = rotr32(state[d] ^ state[a], 8);
    state[c] = state[c] + state[d];
    state[b] = rotr32(state[b] ^ state[c], 7);
}

static void round_fn(uint *state, const uint *msg, uint round)
{
    __constant uchar *schedule = MSG_SCHEDULE[round];
    g(state, 0, 4, 8, 12, msg[schedule[0]], msg[schedule[1]]);
    g(state, 1, 5, 9, 13, msg[schedule[2]], msg[schedule[3]]);
    g(state, 2, 6, 10, 14, msg[schedule[4]], msg[schedule[5]]);
    g(state, 3, 7, 11, 15, msg[schedule[6]], msg[schedule[7]]);
    g(state, 0, 5, 10, 15, msg[schedule[8]], msg[schedule[9]]);
    g(state, 1, 6, 11, 12, msg[schedule[10]], msg[schedule[11]]);
    g(state, 2, 7, 8, 13, msg[schedule[12]], msg[schedule[13]]);
    g(state, 3, 4, 9, 14, msg[schedule[14]], msg[schedule[15]]);
}

// Compress a single block
static void compress(uint *cv, const uint *msg, uint block_len, ulong counter, uint flags)
{
    uint state[16];

    // Initialize state
    state[0] = cv[0];
    state[1] = cv[1];
    state[2] = cv[2];
    state[3] = cv[3];
    state[4] = cv[4];
    state[5] = cv[5];
    state[6] = cv[6];
    state[7] = cv[7];
    state[8] = IV_0;
    state[9] = IV_1;
    state[10] = IV_2;
    state[11] = IV_3;
    state[12] = (uint)counter;
    state[13] = (uint)(counter >> 32);
    state[14] = block_len;
    state[15] = flags;

    // 7 rounds
    for (uint r = 0; r < 7; r++) {
        round_fn(state, msg, r);
    }

    // Finalize
    cv[0] = state[0] ^ state[8];
    cv[1] = state[1] ^ state[9];
    cv[2] = state[2] ^ state[10];
    cv[3] = state[3] ^ state[11];
    cv[4] = state[4] ^ state[12];
    cv[5] = state[5] ^ state[13];
    cv[6] = state[6] ^ state[14];
    cv[7] = state[7] ^ state[15];
}

// Hash a Decred block header (180 bytes = 3 blocks)
// Block 0: bytes 0-63 (CHUNK_START)
// Block 1: bytes 64-127
// Block 2: bytes 128-179 (52 bytes, CHUNK_END | ROOT)
static void blake3_hash_header(const uint *header, uint nonce, uint *hash)
{
    uint cv[8];
    uint msg[16];

    // Initialize CV with IV
    cv[0] = IV_0;
    cv[1] = IV_1;
    cv[2] = IV_2;
    cv[3] = IV_3;
    cv[4] = IV_4;
    cv[5] = IV_5;
    cv[6] = IV_6;
    cv[7] = IV_7;

    // Block 0: bytes 0-63 (CHUNK_START)
    for (uint i = 0; i < 16; i++) {
        msg[i] = header[i];
    }
    compress(cv, msg, BLAKE3_BLOCK_LEN, 0, CHUNK_START);

    // Block 1: bytes 64-127
    for (uint i = 0; i < 16; i++) {
        msg[i] = header[16 + i];
    }
    compress(cv, msg, BLAKE3_BLOCK_LEN, 0, 0);

    // Block 2: bytes 128-179 (52 bytes with nonce at offset 140-143)
    // Nonce is at byte offset 140, which is word offset 35 (header[35])
    // In block 2, this is word offset 35-32 = 3
    for (uint i = 0; i < 13; i++) {
        msg[i] = header[32 + i];
    }
    // Insert nonce at the correct position (offset 140 = byte 12 in block 2 = word 3)
    msg[3] = nonce;  // Nonce is at bytes 140-143 = word 35 = msg[3] in block 2

    // Zero-pad remaining bytes
    for (uint i = 13; i < 16; i++) {
        msg[i] = 0;
    }

    compress(cv, msg, 52, 0, CHUNK_END | ROOT);

    // Output hash
    for (uint i = 0; i < 8; i++) {
        hash[i] = cv[i];
    }
}

// Compare hash against target (little-endian comparison)
static bool check_target(const uint *hash, ulong target)
{
    // For Decred, compare first 8 bytes of hash against target
    ulong h = ((ulong)hash[1] << 32) | hash[0];
    return h <= target;
}

__kernel void blake3_search(
    __global const uint *g_header,  // 180-byte block header (45 words)
    ulong target,                    // Difficulty target
    __global uint *results,          // Output: found nonces
    __global uint *stop              // Stop flag
)
{
    if (*stop)
        return;

    const uint gid = get_global_id(0);

    // Load header into private memory
    uint header[45];
    for (uint i = 0; i < 45; i++) {
        header[i] = g_header[i];
    }

    // Calculate hash with this nonce
    uint hash[8];
    blake3_hash_header(header, gid, hash);

    // Check against target
    if (check_target(hash, target)) {
        *stop = 1;
        const uint k = atomic_inc(results) + 1;
        if (k <= 15) {
            results[k] = gid;
        }
    }
}
