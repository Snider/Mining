/* Miner
 * Copyright (c) 2025 Lethean
 *
 *   Based on XMRig KawPow OpenCL implementation
 *   Copyright 2018-2021 SChernykh   <https://github.com/SChernykh>
 *   Copyright 2016-2021 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
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

#ifdef cl_clang_storage_class_specifiers
#pragma OPENCL EXTENSION cl_clang_storage_class_specifiers : enable
#endif

#ifndef GROUP_SIZE
#define GROUP_SIZE 256
#endif
#define GROUP_SHARE (GROUP_SIZE / 16)

typedef unsigned int       uint32_t;
typedef unsigned long      uint64_t;
#define ROTL32(x, n) rotate((x), (uint32_t)(n))
#define ROTR32(x, n) rotate((x), (uint32_t)(32-n))

// ProgPowZ algorithm constants (differs from KawPow)
#define PROGPOW_LANES           16
#define PROGPOW_REGS            32
#define PROGPOW_DAG_LOADS       4
#define PROGPOW_CACHE_WORDS     4096
#define PROGPOW_CNT_DAG         64
#define PROGPOW_CNT_CACHE       12      // KawPow uses 11
#define PROGPOW_CNT_MATH        20      // KawPow uses 18

#define OPENCL_PLATFORM_UNKNOWN 0
#define OPENCL_PLATFORM_NVIDIA 1
#define OPENCL_PLATFORM_AMD 2
#define OPENCL_PLATFORM_CLOVER 3

#ifndef MAX_OUTPUTS
#define MAX_OUTPUTS 63U
#endif

#ifndef PLATFORM
#ifdef cl_amd_media_ops
#define PLATFORM OPENCL_PLATFORM_AMD
#else
#define PLATFORM OPENCL_PLATFORM_UNKNOWN
#endif
#endif

#define HASHES_PER_GROUP (GROUP_SIZE / PROGPOW_LANES)

#define FNV_PRIME 0x1000193
#define FNV_OFFSET_BASIS 0x811c9dc5
