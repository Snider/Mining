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


#include "backend/opencl/runners/OclBlake3Runner.h"
#include "backend/common/Tags.h"
#include "backend/opencl/OclLaunchData.h"
#include "base/io/log/Tags.h"
#include "backend/opencl/wrappers/OclError.h"
#include "backend/opencl/wrappers/OclLib.h"
#include "base/io/log/Log.h"
#include "base/net/stratum/Job.h"
#include "backend/opencl/cl/blake3/blake3_cl.h"


namespace xmrig {


// Decred block header size
constexpr size_t BLOCK_HEADER_SIZE = 180;


OclBlake3Runner::OclBlake3Runner(size_t index, const OclLaunchData &data) : OclBaseRunner(index, data)
{
    switch (data.thread.worksize())
    {
    case 64:
    case 128:
    case 256:
    case 512:
        m_workGroupSize = data.thread.worksize();
        break;
    }

    if (data.device.vendorId() == OclVendor::OCL_VENDOR_NVIDIA) {
        m_options += " -DPLATFORM=OPENCL_PLATFORM_NVIDIA";
    }
}


OclBlake3Runner::~OclBlake3Runner()
{
    OclLib::release(m_searchKernel);
    OclLib::release(m_searchProgram);
    OclLib::release(m_controlQueue);
    OclLib::release(m_stop);
}


void OclBlake3Runner::run(uint32_t nonce, uint32_t /*nonce_offset*/, uint32_t *hashOutput)
{
    const size_t local_work_size = m_workGroupSize;
    const size_t global_work_offset = nonce;
    const size_t global_work_size = m_intensity - (m_intensity % m_workGroupSize);

    // Upload block header
    enqueueWriteBuffer(m_input, CL_FALSE, 0, m_blobSize, m_blob);

    // Clear results and stop flag
    const uint32_t zero[2] = {};
    enqueueWriteBuffer(m_output, CL_FALSE, 0, sizeof(uint32_t), zero);
    enqueueWriteBuffer(m_stop, CL_FALSE, 0, sizeof(uint32_t), zero);

    // Run search kernel
    const cl_int ret = OclLib::enqueueNDRangeKernel(m_queue, m_searchKernel, 1, &global_work_offset, &global_work_size, &local_work_size, 0, nullptr, nullptr);
    if (ret != CL_SUCCESS) {
        LOG_ERR("%s" RED(" error ") RED_BOLD("%s") RED(" when calling ") RED_BOLD("clEnqueueNDRangeKernel") RED(" for kernel ") RED_BOLD("blake3_search"),
            ocl_tag(), OclError::toString(ret));

        throw std::runtime_error(OclError::toString(ret));
    }

    // Read results
    uint32_t output[16] = {};
    enqueueReadBuffer(m_output, CL_TRUE, 0, sizeof(output), output);

    if (output[0] > 15) {
        output[0] = 15;
    }

    hashOutput[0xFF] = output[0];
    memcpy(hashOutput, output + 1, output[0] * sizeof(uint32_t));
}


void OclBlake3Runner::set(const Job &job, uint8_t *blob)
{
    m_blob = blob;
    m_blobSize = job.size();

    if (m_blobSize > BLOCK_HEADER_SIZE) {
        m_blobSize = BLOCK_HEADER_SIZE;
    }

    // Update kernel arguments
    const uint64_t target = job.target();

    OclLib::setKernelArg(m_searchKernel, 0, sizeof(cl_mem), &m_input);
    OclLib::setKernelArg(m_searchKernel, 1, sizeof(target), &target);
    OclLib::setKernelArg(m_searchKernel, 2, sizeof(cl_mem), &m_output);
    OclLib::setKernelArg(m_searchKernel, 3, sizeof(cl_mem), &m_stop);

    enqueueWriteBuffer(m_input, CL_TRUE, 0, m_blobSize, m_blob);
}


void OclBlake3Runner::jobEarlyNotification(const Job&)
{
    const uint32_t one = 1;
    const cl_int ret = OclLib::enqueueWriteBuffer(m_controlQueue, m_stop, CL_TRUE, 0, sizeof(one), &one, 0, nullptr, nullptr);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }
}


void OclBlake3Runner::build()
{
    OclBaseRunner::build();

    // Build search kernel
    cl_int ret = 0;
    const char* source = blake3_cl;
    m_searchProgram = OclLib::createProgramWithSource(m_ctx, 1, &source, nullptr, &ret);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }

    std::string options = m_options;
    options += " -DGROUP_SIZE=" + std::to_string(m_workGroupSize);

    cl_device_id device = data().device.id();
    ret = OclLib::buildProgram(m_searchProgram, 1, &device, options.c_str());
    if (ret != CL_SUCCESS) {
        LOG_ERR("%s" RED(" Blake3 kernel build failed: %s"), ocl_tag(), OclLib::getProgramBuildLog(m_searchProgram, device).data());
        throw std::runtime_error(OclError::toString(ret));
    }

    m_searchKernel = OclLib::createKernel(m_searchProgram, "blake3_search", &ret);
    if (ret != CL_SUCCESS) {
        throw std::runtime_error(OclError::toString(ret));
    }

    LOG_INFO("%s " CYAN("Blake3") " OpenCL kernel compiled", Tags::opencl());
}


void OclBlake3Runner::init()
{
    OclBaseRunner::init();

    m_controlQueue = OclLib::createCommandQueue(m_ctx, data().device.id());
    m_stop = OclLib::createBuffer(m_ctx, CL_MEM_READ_ONLY, sizeof(uint32_t));
}


} // namespace xmrig
