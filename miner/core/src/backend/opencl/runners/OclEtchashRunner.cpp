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

#include <stdexcept>


#include "backend/opencl/runners/OclEtchashRunner.h"
#include "backend/common/Tags.h"
#include "3rdparty/libethash/ethash_internal.h"
#include "3rdparty/libethash/data_sizes.h"
#include "backend/opencl/kernels/etchash/Etchash_CalculateDAGKernel.h"
#include "backend/opencl/OclLaunchData.h"
#include "backend/opencl/wrappers/OclError.h"
#include "backend/opencl/wrappers/OclLib.h"
#include "base/io/log/Log.h"
#include "base/io/log/Tags.h"
#include "base/net/stratum/Job.h"
#include "base/tools/Chrono.h"
#include "crypto/common/VirtualMemory.h"
#include "crypto/etchash/ETCCache.h"
#include "crypto/etchash/ETChash.h"
#include "backend/opencl/cl/etchash/etchash_cl.h"
#include "backend/opencl/cl/etchash/etchash_dag_cl.h"


namespace xmrig {


// ETChash uses 30000 blocks per epoch (pre-ECIP-1099)
// Post ECIP-1099 (block 11700000): epoch = 390 + (block - 11700000) / 60000
constexpr uint32_t EPOCH_LENGTH = 30000;
constexpr uint32_t ECIP1099_BLOCK = 11700000;
constexpr uint32_t ECIP1099_EPOCH = 390;
constexpr size_t BLOB_SIZE = 32;  // Header hash size


static uint32_t calculateEpoch(uint32_t height, bool isETC)
{
    if (!isETC) {
        // Standard Ethash epoch calculation
        return height / EPOCH_LENGTH;
    }

    // ECIP-1099 epoch calculation for ETC
    if (height < ECIP1099_BLOCK) {
        return height / EPOCH_LENGTH;
    }
    return ECIP1099_EPOCH + (height - ECIP1099_BLOCK) / 60000;
}


OclEtchashRunner::OclEtchashRunner(size_t index, const OclLaunchData &data) : OclBaseRunner(index, data)
{
    // Determine if this is ETC or ETH
    m_isETC = (data.algorithm.id() == Algorithm::ETCHASH_ETC);

    switch (data.thread.worksize())
    {
    case 64:
    case 128:
    case 256:
        m_workGroupSize = data.thread.worksize();
        break;
    }

    if (data.device.vendorId() == OclVendor::OCL_VENDOR_NVIDIA) {
        m_options += " -DPLATFORM=OPENCL_PLATFORM_NVIDIA";
        m_dagWorkGroupSize = 32;
    }
}


OclEtchashRunner::~OclEtchashRunner()
{
    OclLib::release(m_lightCache);
    OclLib::release(m_dag);

    delete m_calculateDagKernel;

    OclLib::release(m_searchKernel);
    OclLib::release(m_searchProgram);

    OclLib::release(m_controlQueue);
    OclLib::release(m_stop);
}


void OclEtchashRunner::run(uint32_t nonce, uint32_t /*nonce_offset*/, uint32_t *hashOutput)
{
    const size_t local_work_size = m_workGroupSize;
    const size_t global_work_offset = nonce;
    const size_t global_work_size = m_intensity - (m_intensity % m_workGroupSize);

    enqueueWriteBuffer(m_input, CL_FALSE, 0, BLOB_SIZE, m_blob);

    const uint32_t zero[2] = {};
    enqueueWriteBuffer(m_output, CL_FALSE, 0, sizeof(uint32_t), zero);
    enqueueWriteBuffer(m_stop, CL_FALSE, 0, sizeof(uint32_t) * 2, zero);

    m_skippedHashes = 0;

    const cl_int ret = OclLib::enqueueNDRangeKernel(m_queue, m_searchKernel, 1, &global_work_offset, &global_work_size, &local_work_size, 0, nullptr, nullptr);
    if (ret != CL_SUCCESS) {
        LOG_ERR("%s" RED(" error ") RED_BOLD("%s") RED(" when calling ") RED_BOLD("clEnqueueNDRangeKernel") RED(" for kernel ") RED_BOLD("ethash_search"),
            ocl_tag(), OclError::toString(ret));

        throw std::runtime_error(OclError::toString(ret));
    }

    uint32_t stop[2] = {};
    enqueueReadBuffer(m_stop, CL_FALSE, 0, sizeof(stop), stop);

    uint32_t output[16] = {};
    enqueueReadBuffer(m_output, CL_TRUE, 0, sizeof(output), output);

    m_skippedHashes = stop[1] * m_workGroupSize;

    if (output[0] > 15) {
        output[0] = 15;
    }

    hashOutput[0xFF] = output[0];
    memcpy(hashOutput, output + 1, output[0] * sizeof(uint32_t));
}


void OclEtchashRunner::set(const Job &job, uint8_t *blob)
{
    m_blockHeight = static_cast<uint32_t>(job.height());

    const uint32_t epoch = calculateEpoch(m_blockHeight, m_isETC);

    const uint64_t dag_size = ETCCache::dagSize(epoch);
    if (dag_size > m_dagCapacity) {
        OclLib::release(m_dag);

        m_dagCapacity = VirtualMemory::align(dag_size, 16 * 1024 * 1024);
        m_dag = OclLib::createBuffer(m_ctx, CL_MEM_READ_WRITE, m_dagCapacity);
    }

    if (epoch != m_epoch) {
        m_epoch = epoch;

        {
            std::lock_guard<std::mutex> lock(ETCCache::s_cacheMutex);

            ETCCache& cache = m_isETC ? ETCCache::s_etcCache : ETCCache::s_ethCache;
            cache.init(epoch, m_isETC);

            if (cache.size() > m_lightCacheCapacity) {
                OclLib::release(m_lightCache);

                m_lightCacheCapacity = VirtualMemory::align(cache.size());
                m_lightCache = OclLib::createBuffer(m_ctx, CL_MEM_READ_ONLY, m_lightCacheCapacity);
            }

            m_lightCacheSize = cache.size();
            enqueueWriteBuffer(m_lightCache, CL_TRUE, 0, m_lightCacheSize, cache.data());
        }

        const uint64_t start_ms = Chrono::steadyMSecs();

        const uint32_t dag_words = dag_size / sizeof(node);
        m_calculateDagKernel->setArgs(0, m_lightCache, m_dag, dag_words, m_lightCacheSize / sizeof(node));

        constexpr uint32_t N = 1 << 18;

        for (uint32_t start = 0; start < dag_words; start += N) {
            m_calculateDagKernel->setArg(0, sizeof(start), &start);
            m_calculateDagKernel->enqueue(m_queue, N, m_dagWorkGroupSize);
        }

        OclLib::finish(m_queue);

        const char* algoName = m_isETC ? "ETChash" : "Ethash";
        LOG_INFO("%s " CYAN("%s") " DAG for epoch " WHITE_BOLD("%u") " calculated " BLACK_BOLD("(%" PRIu64 "ms)"), Tags::opencl(), algoName, epoch, Chrono::steadyMSecs() - start_ms);
    }

    // Update search kernel arguments
    const uint64_t target = job.target();

    OclLib::setKernelArg(m_searchKernel, 0, sizeof(cl_mem), &m_dag);
    OclLib::setKernelArg(m_searchKernel, 1, sizeof(cl_mem), &m_input);
    OclLib::setKernelArg(m_searchKernel, 2, sizeof(target), &target);

    const uint32_t hack_false = 0;
    OclLib::setKernelArg(m_searchKernel, 3, sizeof(hack_false), &hack_false);
    OclLib::setKernelArg(m_searchKernel, 4, sizeof(cl_mem), &m_output);
    OclLib::setKernelArg(m_searchKernel, 5, sizeof(cl_mem), &m_stop);

    m_blob = blob;
    enqueueWriteBuffer(m_input, CL_TRUE, 0, BLOB_SIZE, m_blob);
}


void OclEtchashRunner::jobEarlyNotification(const Job&)
{
    const uint32_t one = 1;
    const cl_int ret = OclLib::enqueueWriteBuffer(m_controlQueue, m_stop, CL_TRUE, 0, sizeof(one), &one, 0, nullptr, nullptr);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }
}


void xmrig::OclEtchashRunner::build()
{
    OclBaseRunner::build();

    m_calculateDagKernel = new Etchash_CalculateDAGKernel(m_program);

    // Build search kernel
    cl_int ret = 0;
    const char* source = etchash_cl;
    m_searchProgram = OclLib::createProgramWithSource(m_ctx, 1, &source, nullptr, &ret);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }

    // Calculate DAG elements for current epoch (use a reasonable default)
    const uint32_t epoch = 0;  // Will be updated in set()
    const uint64_t dag_elements = dag_sizes[epoch] / 256;

    std::string options = m_options;
    options += " -DPROGPOW_DAG_ELEMENTS=" + std::to_string(dag_elements);
    options += " -DGROUP_SIZE=" + std::to_string(m_workGroupSize);

    cl_device_id device = data().device.id();
    ret = OclLib::buildProgram(m_searchProgram, 1, &device, options.c_str());
    if (ret != CL_SUCCESS) {
        LOG_ERR("%s" RED(" ETChash kernel build failed: %s"), ocl_tag(), OclLib::getProgramBuildLog(m_searchProgram, device).data());
        throw std::runtime_error(OclError::toString(ret));
    }

    m_searchKernel = OclLib::createKernel(m_searchProgram, "ethash_search", &ret);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }
}


void xmrig::OclEtchashRunner::init()
{
    OclBaseRunner::init();

    m_controlQueue = OclLib::createCommandQueue(m_ctx, data().device.id());
    m_stop = OclLib::createBuffer(m_ctx, CL_MEM_READ_ONLY, sizeof(uint32_t) * 2);
}

} // namespace xmrig
