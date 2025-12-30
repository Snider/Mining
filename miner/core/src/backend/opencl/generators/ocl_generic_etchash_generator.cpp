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


#include "backend/opencl/OclThreads.h"
#include "backend/opencl/wrappers/OclDevice.h"
#include "base/crypto/Algorithm.h"


#include <algorithm>


namespace xmrig {


bool ocl_generic_etchash_generator(const OclDevice &device, const Algorithm &algorithm, OclThreads &threads)
{
    if (algorithm.family() != Algorithm::ETCHASH) {
        return false;
    }

    // ETChash requires at least 3GB free memory for DAG (current epoch ~5GB)
    if (device.freeMemSize() < (3ULL * 1024 * 1024 * 1024)) {
        return false;
    }

    // Calculate intensity based on device memory and compute units
    const size_t freeMem = device.freeMemSize();

    // Reserve memory for DAG (~6GB for later epochs) and light cache
    const size_t dag_size = 6ULL * 1024 * 1024 * 1024; // Max DAG size estimate
    const size_t cache_size = 128 * 1024 * 1024;       // Cache size estimate
    const size_t available = freeMem > (dag_size + cache_size) ? freeMem - dag_size - cache_size : 0;

    // Each work item needs approximately 128 bytes of mix state
    uint32_t intensity = static_cast<uint32_t>(std::min(available / 128, static_cast<size_t>(1U << 24)));

    // Round down to work group size multiple
    intensity = (intensity / 128) * 128;

    // Minimum intensity
    if (intensity < 128 * 1024) {
        intensity = 128 * 1024;
    }

    // Maximum intensity
    intensity = std::min(intensity, static_cast<uint32_t>(1 << 23));

    // Determine optimal work size based on device type
    uint32_t worksize = 128;

    // NVIDIA cards often perform better with smaller work groups
    if (device.vendorId() == OCL_VENDOR_NVIDIA) {
        worksize = 64;
        // Reduce intensity for NVIDIA to avoid timeout
        intensity = std::min(intensity, static_cast<uint32_t>(1 << 20));
    }

    // AMD Navi architecture (gfx10xx) may need adjustment
    if (device.type() >= OclDevice::Navi_10 && device.type() <= OclDevice::Navi_21) {
        intensity = std::min(intensity, static_cast<uint32_t>(1 << 21));
    }

    threads.add(OclThread(device.index(), intensity, worksize, 1));

    return true;
}


} // namespace xmrig
