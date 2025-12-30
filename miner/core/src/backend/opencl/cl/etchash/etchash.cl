/* Miner
 * Copyright (c) 2025 Lethean
 *
 *   ETChash/Ethash OpenCL mining kernel
 *   Based on various open-source Ethash implementations
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

//
// ETChash/Ethash mining kernel
// Standard Ethash algorithm: DAG lookup + Keccak
//

#define FNV_PRIME 0x01000193
#define ETHASH_ACCESSES 64
#define ETHASH_MIX_BYTES 128
#define ETHASH_HASH_BYTES 64

#ifndef GROUP_SIZE
#define GROUP_SIZE 128
#endif

#ifndef PROGPOW_DAG_ELEMENTS
#define PROGPOW_DAG_ELEMENTS 0
#endif

__constant uint2 const Keccak_f1600_RC[24] = {
    (uint2)(0x00000001, 0x00000000),
    (uint2)(0x00008082, 0x00000000),
    (uint2)(0x0000808a, 0x80000000),
    (uint2)(0x80008000, 0x80000000),
    (uint2)(0x0000808b, 0x00000000),
    (uint2)(0x80000001, 0x00000000),
    (uint2)(0x80008081, 0x80000000),
    (uint2)(0x00008009, 0x80000000),
    (uint2)(0x0000008a, 0x00000000),
    (uint2)(0x00000088, 0x00000000),
    (uint2)(0x80008009, 0x00000000),
    (uint2)(0x8000000a, 0x00000000),
    (uint2)(0x8000808b, 0x00000000),
    (uint2)(0x0000008b, 0x80000000),
    (uint2)(0x00008089, 0x80000000),
    (uint2)(0x00008003, 0x80000000),
    (uint2)(0x00008002, 0x80000000),
    (uint2)(0x00000080, 0x80000000),
    (uint2)(0x0000800a, 0x00000000),
    (uint2)(0x8000000a, 0x80000000),
    (uint2)(0x80008081, 0x80000000),
    (uint2)(0x00008080, 0x80000000),
    (uint2)(0x80000001, 0x00000000),
    (uint2)(0x80008008, 0x80000000),
};

#if PLATFORM == OPENCL_PLATFORM_NVIDIA && COMPUTE >= 35
static uint2 ROL2(const uint2 a, const int offset)
{
    uint2 result;
    if (offset >= 32)
    {
        asm("shf.l.wrap.b32 %0, %1, %2, %3;" : "=r"(result.x) : "r"(a.x), "r"(a.y), "r"(offset));
        asm("shf.l.wrap.b32 %0, %1, %2, %3;" : "=r"(result.y) : "r"(a.y), "r"(a.x), "r"(offset));
    }
    else
    {
        asm("shf.l.wrap.b32 %0, %1, %2, %3;" : "=r"(result.x) : "r"(a.y), "r"(a.x), "r"(offset));
        asm("shf.l.wrap.b32 %0, %1, %2, %3;" : "=r"(result.y) : "r"(a.x), "r"(a.y), "r"(offset));
    }
    return result;
}
#elif defined(cl_amd_media_ops)
#pragma OPENCL EXTENSION cl_amd_media_ops : enable
static uint2 ROL2(const uint2 vv, const int r)
{
    if (r <= 32)
    {
        return amd_bitalign((vv).xy, (vv).yx, 32 - r);
    }
    else
    {
        return amd_bitalign((vv).yx, (vv).xy, 64 - r);
    }
}
#else
static uint2 ROL2(const uint2 v, const int n)
{
    uint2 result;
    if (n <= 32)
    {
        result.y = ((v.y << (n)) | (v.x >> (32 - n)));
        result.x = ((v.x << (n)) | (v.y >> (32 - n)));
    }
    else
    {
        result.y = ((v.x << (n - 32)) | (v.y >> (64 - n)));
        result.x = ((v.y << (n - 32)) | (v.x >> (64 - n)));
    }
    return result;
}
#endif

static void chi(uint2* a, const uint n, const uint2* t)
{
    a[n + 0] = bitselect(t[n + 0] ^ t[n + 2], t[n + 0], t[n + 1]);
    a[n + 1] = bitselect(t[n + 1] ^ t[n + 3], t[n + 1], t[n + 2]);
    a[n + 2] = bitselect(t[n + 2] ^ t[n + 4], t[n + 2], t[n + 3]);
    a[n + 3] = bitselect(t[n + 3] ^ t[n + 0], t[n + 3], t[n + 4]);
    a[n + 4] = bitselect(t[n + 4] ^ t[n + 1], t[n + 4], t[n + 0]);
}

static void keccak_f1600_round(uint2* a, uint r)
{
    uint2 t[25];
    uint2 u;

    // Theta
    t[0] = a[0] ^ a[5] ^ a[10] ^ a[15] ^ a[20];
    t[1] = a[1] ^ a[6] ^ a[11] ^ a[16] ^ a[21];
    t[2] = a[2] ^ a[7] ^ a[12] ^ a[17] ^ a[22];
    t[3] = a[3] ^ a[8] ^ a[13] ^ a[18] ^ a[23];
    t[4] = a[4] ^ a[9] ^ a[14] ^ a[19] ^ a[24];
    u = t[4] ^ ROL2(t[1], 1);
    a[0] ^= u;
    a[5] ^= u;
    a[10] ^= u;
    a[15] ^= u;
    a[20] ^= u;
    u = t[0] ^ ROL2(t[2], 1);
    a[1] ^= u;
    a[6] ^= u;
    a[11] ^= u;
    a[16] ^= u;
    a[21] ^= u;
    u = t[1] ^ ROL2(t[3], 1);
    a[2] ^= u;
    a[7] ^= u;
    a[12] ^= u;
    a[17] ^= u;
    a[22] ^= u;
    u = t[2] ^ ROL2(t[4], 1);
    a[3] ^= u;
    a[8] ^= u;
    a[13] ^= u;
    a[18] ^= u;
    a[23] ^= u;
    u = t[3] ^ ROL2(t[0], 1);
    a[4] ^= u;
    a[9] ^= u;
    a[14] ^= u;
    a[19] ^= u;
    a[24] ^= u;

    // Rho Pi
    t[0] = a[0];
    t[10] = ROL2(a[1], 1);
    t[20] = ROL2(a[2], 62);
    t[5] = ROL2(a[3], 28);
    t[15] = ROL2(a[4], 27);

    t[16] = ROL2(a[5], 36);
    t[1] = ROL2(a[6], 44);
    t[11] = ROL2(a[7], 6);
    t[21] = ROL2(a[8], 55);
    t[6] = ROL2(a[9], 20);

    t[7] = ROL2(a[10], 3);
    t[17] = ROL2(a[11], 10);
    t[2] = ROL2(a[12], 43);
    t[12] = ROL2(a[13], 25);
    t[22] = ROL2(a[14], 39);

    t[23] = ROL2(a[15], 41);
    t[8] = ROL2(a[16], 45);
    t[18] = ROL2(a[17], 15);
    t[3] = ROL2(a[18], 21);
    t[13] = ROL2(a[19], 8);

    t[14] = ROL2(a[20], 18);
    t[24] = ROL2(a[21], 2);
    t[9] = ROL2(a[22], 61);
    t[19] = ROL2(a[23], 56);
    t[4] = ROL2(a[24], 14);

    // Chi
    chi(a, 0, t);

    // Iota
    a[0] ^= Keccak_f1600_RC[r];

    chi(a, 5, t);
    chi(a, 10, t);
    chi(a, 15, t);
    chi(a, 20, t);
}

static void keccak_f1600(uint2* a)
{
    for (uint r = 0; r < 24; ++r)
    {
        keccak_f1600_round(a, r);
    }
}

static uint fnv(uint x, uint y)
{
    return x * FNV_PRIME ^ y;
}

static uint4 fnv4(uint4 x, uint4 y)
{
    return x * FNV_PRIME ^ y;
}

typedef union
{
    uint words[64 / sizeof(uint)];
    uint2 uint2s[64 / sizeof(uint2)];
    uint4 uint4s[64 / sizeof(uint4)];
} hash64_t;

typedef union
{
    uint words[128 / sizeof(uint)];
    uint4 uint4s[128 / sizeof(uint4)];
} hash128_t;

typedef union
{
    uint words[200 / sizeof(uint)];
    uint2 uint2s[200 / sizeof(uint2)];
} hash200_t;

// Keccak-256 final hash (first 4 uint2s = 256 bits)
static void keccak_f256(uint2* state)
{
    // Pad for Keccak-256
    for (uint i = 4; i < 25; ++i)
    {
        state[i] = (uint2)(0, 0);
    }
    state[4].x = 0x00000001;
    state[8].y = 0x80000000;
    keccak_f1600(state);
}

__kernel void ethash_search(
    __global uint4 const* g_dag,
    __global uint const* g_header,
    ulong target,
    uint hack_false,
    __global uint* results,
    __global uint* stop)
{
    if (*stop)
        return;

    const uint gid = get_global_id(0);

    // Initialize state with header (32 bytes = 8 words)
    hash200_t state;
    for (uint i = 0; i < 25; ++i)
        state.uint2s[i] = (uint2)(0, 0);

    // Load header hash (32 bytes)
    state.words[0] = g_header[0];
    state.words[1] = g_header[1];
    state.words[2] = g_header[2];
    state.words[3] = g_header[3];
    state.words[4] = g_header[4];
    state.words[5] = g_header[5];
    state.words[6] = g_header[6];
    state.words[7] = g_header[7];

    // Add nonce (8 bytes)
    state.words[8] = gid;
    state.words[9] = 0;

    // Keccak-512 padding
    state.words[10] = 0x00000001;
    state.uint2s[8].y = 0x80000000;

    // Keccak-512 to get seed
    keccak_f1600(state.uint2s);

    // Initialize mix (128 bytes = 32 words)
    uint mix[32];
    for (uint i = 0; i < 16; ++i)
    {
        mix[i] = state.words[i % 16];
        mix[i + 16] = state.words[i % 16];
    }

    // DAG accesses
    const uint dag_elements = PROGPOW_DAG_ELEMENTS;
    for (uint i = 0; i < ETHASH_ACCESSES; ++i)
    {
        uint p = fnv(i ^ state.words[0], mix[i % 32]) % dag_elements;

        // Load 128 bytes from DAG (2 * 64 bytes)
        uint4 dag_data[8];
        for (uint j = 0; j < 8; ++j)
        {
            dag_data[j] = g_dag[p * 2 + j / 4];
        }

        // FNV mix
        for (uint j = 0; j < 32; ++j)
        {
            mix[j] = fnv(mix[j], ((uint*)dag_data)[j]);
        }
    }

    // Compress mix to 32 bytes (8 words)
    uint cmix[8];
    for (uint i = 0; i < 8; ++i)
    {
        cmix[i] = fnv(fnv(fnv(mix[i*4], mix[i*4+1]), mix[i*4+2]), mix[i*4+3]);
    }

    // Final Keccak-256
    hash200_t final_state;
    for (uint i = 0; i < 25; ++i)
        final_state.uint2s[i] = (uint2)(0, 0);

    // Copy seed state (first 8 words) and mix (8 words)
    for (uint i = 0; i < 8; ++i)
    {
        final_state.words[i] = state.words[i];
    }
    for (uint i = 0; i < 8; ++i)
    {
        final_state.words[8 + i] = cmix[i];
    }

    // Keccak-256 padding
    final_state.words[16] = 0x00000001;
    final_state.uint2s[8].y = 0x80000000;

    keccak_f1600(final_state.uint2s);

    // Check against target (compare first 8 bytes / 64 bits)
    ulong result = as_ulong(as_uchar8((ulong)final_state.words[0] | ((ulong)final_state.words[1] << 32)).s76543210);

    if (result <= target)
    {
        *stop = 1;
        const uint k = atomic_inc(results) + 1;
        if (k <= 15)
            results[k] = gid;
    }
}
