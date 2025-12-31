/* XMRig
 * Copyright 2010      Jeff Garzik <jgarzik@pobox.com>
 * Copyright 2012-2014 pooler      <pooler@litecoinpool.org>
 * Copyright 2014      Lucas Jones <https://github.com/lucasjones>
 * Copyright 2014-2016 Wolf9466    <https://github.com/OhGodAPet>
 * Copyright 2016      Jay D Dee   <jayddee246@gmail.com>
 * Copyright 2017-2018 XMR-Stak    <https://github.com/fireice-uk>, <https://github.com/psychocrypt>
 * Copyright 2018      Lee Clagett <https://github.com/vtnerd>
 * Copyright 2018-2019 SChernykh   <https://github.com/SChernykh>
 * Copyright 2018-2019 tevador     <tevador@gmail.com>
 * Copyright 2016-2019 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
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


#include "crypto/common/MemoryPool.h"
#include "crypto/common/VirtualMemory.h"


#include <cassert>


namespace xmrig {


constexpr size_t pageSize = 2 * 1024 * 1024;


} // namespace xmrig


xmrig::MemoryPool::MemoryPool(size_t size, bool hugePages, uint32_t node)
{
    if (!size) {
        return;
    }

    constexpr size_t alignment = 1 << 24;

    m_memory = new VirtualMemory(size * pageSize + alignment, hugePages, false, false, node, VirtualMemory::kDefaultHugePageSize);

    m_alignOffset = (alignment - (((size_t)m_memory->scratchpad()) % alignment)) % alignment;
}


xmrig::MemoryPool::~MemoryPool()
{
    delete m_memory;
}


bool xmrig::MemoryPool::isHugePages(uint32_t) const
{
    return m_memory && m_memory->isHugePages();
}


uint8_t *xmrig::MemoryPool::get(size_t size, uint32_t)
{
    assert(!(size % pageSize));

    // SECURITY: Check for integer overflow before subtraction to prevent underflow
    // The subtraction (m_memory->size() - m_offset - m_alignOffset) can wrap around
    // if (m_offset + m_alignOffset) > m_memory->size()
    if (!m_memory) {
        return nullptr;
    }

    const size_t totalSize = m_memory->size();

    // Check for overflow: ensure m_alignOffset doesn't exceed total size
    if (m_alignOffset > totalSize) {
        return nullptr;
    }

    // Check for overflow: ensure m_offset doesn't exceed remaining size
    if (m_offset > totalSize - m_alignOffset) {
        return nullptr;
    }

    // Now safe to compute remaining size
    const size_t remaining = totalSize - m_alignOffset - m_offset;
    if (remaining < size) {
        return nullptr;
    }

    uint8_t *out = m_memory->scratchpad() + m_alignOffset + m_offset;

    m_offset += size;
    ++m_refs;

    return out;
}


void xmrig::MemoryPool::release(uint32_t)
{
    assert(m_refs > 0);

    if (m_refs > 0) {
        --m_refs;
    }

    if (m_refs == 0) {
        m_offset = 0;
    }
}
