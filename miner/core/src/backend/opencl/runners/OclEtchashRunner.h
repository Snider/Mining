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

#ifndef XMRIG_OCLETCHASHRUNNER_H
#define XMRIG_OCLETCHASHRUNNER_H


#include "backend/opencl/runners/OclBaseRunner.h"

#include <mutex>

namespace xmrig {


class Etchash_CalculateDAGKernel;


class OclEtchashRunner : public OclBaseRunner
{
public:
    XMRIG_DISABLE_COPY_MOVE_DEFAULT(OclEtchashRunner)

    OclEtchashRunner(size_t index, const OclLaunchData &data);
    ~OclEtchashRunner() override;

protected:
    void run(uint32_t nonce, uint32_t nonce_offset, uint32_t *hashOutput) override;
    void set(const Job &job, uint8_t *blob) override;
    void build() override;
    void init() override;
    void jobEarlyNotification(const Job& job) override;
    uint32_t processedHashes() const override { return m_intensity - m_skippedHashes; }

private:
    uint8_t* m_blob = nullptr;
    uint32_t m_skippedHashes = 0;

    uint32_t m_blockHeight = 0;
    uint32_t m_epoch = 0xFFFFFFFFUL;

    cl_mem m_lightCache = nullptr;
    size_t m_lightCacheSize = 0;
    size_t m_lightCacheCapacity = 0;

    cl_mem m_dag = nullptr;
    size_t m_dagCapacity = 0;

    Etchash_CalculateDAGKernel* m_calculateDagKernel = nullptr;

    cl_kernel m_searchKernel = nullptr;
    cl_program m_searchProgram = nullptr;

    size_t m_workGroupSize = 128;
    size_t m_dagWorkGroupSize = 64;

    cl_command_queue m_controlQueue = nullptr;
    cl_mem m_stop = nullptr;

    bool m_isETC = true;  // true for ETChash, false for Ethash
};


} /* namespace xmrig */


#endif // XMRIG_OCLETCHASHRUNNER_H
