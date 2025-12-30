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


bool ocl_generic_blake3_generator(const OclDevice &device, const Algorithm &algorithm, OclThreads &threads)
{
    if (algorithm.family() != Algorithm::BLAKE3) {
        return false;
    }

    // Blake3 is compute-bound with minimal memory requirements
    // Each work item only needs ~200 bytes for state

    // Calculate intensity based on compute units
    // Blake3 is very parallel - maximize GPU utilization
    uint32_t intensity = device.computeUnits() * 1024 * 8;

    // Cap intensity based on available memory (very permissive for Blake3)
    const size_t freeMem = device.freeMemSize();
    const size_t memPerThread = 256;  // ~256 bytes per work item
    const uint32_t maxByMem = static_cast<uint32_t>(freeMem / memPerThread);
    intensity = std::min(intensity, maxByMem);

    // Round down to work group size multiple
    uint32_t worksize = 256;

    // NVIDIA cards often perform better with smaller work groups
    if (device.vendorId() == OCL_VENDOR_NVIDIA) {
        worksize = 128;
    }

    intensity = (intensity / worksize) * worksize;

    // Minimum intensity
    if (intensity < worksize * 256) {
        intensity = worksize * 256;
    }

    // Maximum intensity (avoid GPU timeout)
    intensity = std::min(intensity, static_cast<uint32_t>(1 << 24));

    threads.add(OclThread(device.index(), intensity, worksize, 1));

    return true;
}


} // namespace xmrig
