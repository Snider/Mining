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

#ifndef XMRIG_OCLBLAKE3RUNNER_H
#define XMRIG_OCLBLAKE3RUNNER_H


#include "backend/opencl/runners/OclBaseRunner.h"


namespace xmrig {


class OclBlake3Runner : public OclBaseRunner
{
public:
    XMRIG_DISABLE_COPY_MOVE_DEFAULT(OclBlake3Runner)

    OclBlake3Runner(size_t index, const OclLaunchData &data);
    ~OclBlake3Runner() override;

protected:
    void run(uint32_t nonce, uint32_t nonce_offset, uint32_t *hashOutput) override;
    void set(const Job &job, uint8_t *blob) override;
    void build() override;
    void init() override;
    void jobEarlyNotification(const Job& job) override;
    uint32_t processedHashes() const override { return m_intensity; }

private:
    uint8_t* m_blob = nullptr;
    size_t m_blobSize = 0;

    cl_kernel m_searchKernel = nullptr;
    cl_program m_searchProgram = nullptr;

    size_t m_workGroupSize = 256;

    cl_command_queue m_controlQueue = nullptr;
    cl_mem m_stop = nullptr;
};


} /* namespace xmrig */


#endif // XMRIG_OCLBLAKE3RUNNER_H
