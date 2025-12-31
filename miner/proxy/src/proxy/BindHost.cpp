/* XMRig
 * Copyright 2010      Jeff Garzik <jgarzik@pobox.com>
 * Copyright 2012-2014 pooler      <pooler@litecoinpool.org>
 * Copyright 2014      Lucas Jones <https://github.com/lucasjones>
 * Copyright 2014-2016 Wolf9466    <https://github.com/OhGodAPet>
 * Copyright 2016      Jay D Dee   <jayddee246@gmail.com>
 * Copyright 2017-2018 XMR-Stak    <https://github.com/fireice-uk>, <https://github.com/psychocrypt>
 * Copyright 2018-2020 SChernykh   <https://github.com/SChernykh>
 * Copyright 2016-2020 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
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


#include <assert.h>
#include <string.h>
#include <stdlib.h>


#include "proxy/BindHost.h"
#include "3rdparty/rapidjson/document.h"


xmrig::BindHost::BindHost(const char *addr) :
    m_tls(false),
    m_version(0),
    m_port(0)
{
    if (!addr || strlen(addr) < 5) {
        return;
    }

    if (addr[0] == '[') {
        parseIPv6(addr);
        return;
    }

    parseIPv4(addr);
}


xmrig::BindHost::BindHost(const char *host, uint16_t port, int version) :
    m_tls(false),
    m_version(version),
    m_port(port),
    m_host(host)
{
}


xmrig::BindHost::BindHost(const rapidjson::Value &object) :
    m_tls(false),
    m_version(0),
    m_port(0)
{
    // SECURITY FIX: Validate JSON structure before accessing members
    if (!object.IsObject()) {
        return;
    }

    // Validate "host" field exists and is a string
    if (!object.HasMember("host") || !object["host"].IsString()) {
        return;
    }

    if (!parseHost(object["host"].GetString())) {
        return;
    }

    // Validate "port" field exists and is a number
    if (object.HasMember("port") && object["port"].IsUint()) {
        const uint32_t port = object["port"].GetUint();
        if (port <= 65535) {
            m_port = static_cast<uint16_t>(port);
        }
    }

    // Validate "tls" field exists and is a boolean
    if (object.HasMember("tls") && object["tls"].IsBool()) {
        m_tls = object["tls"].GetBool();
    }
}


rapidjson::Value xmrig::BindHost::toJSON(rapidjson::Document &doc) const
{
    using namespace rapidjson;

    auto &allocator = doc.GetAllocator();

    Value obj(kObjectType);

    obj.AddMember("host", StringRef(host()), allocator);
    obj.AddMember("port", port(), allocator);
    obj.AddMember("tls",  isTLS(), allocator);

    return obj;
}


bool xmrig::BindHost::parseHost(const char *host)
{
    assert(host != nullptr && strlen(host) >= 2);
    m_version = 0;

    if (host == nullptr || strlen(host) < 2) {
        return false;
    }

    if (host[0] == '[') {
        const char *end = strchr(host, ']');
        if (!end) {
            return false;
        }

        // SECURITY FIX (HIGH-011): Use String copy constructor instead of raw new[]
        const size_t size = end - host - 1;  // Length of content between brackets
        m_version = 6;
        m_host    = String(host + 1, size);
    }
    else {
        m_version = strchr(host, ':') != nullptr ? 6 : 4;
        m_host    = host;
    }

    return m_version > 0;
}


void xmrig::BindHost::parseIPv4(const char *addr)
{
    const char *port = strchr(addr, ':');
    if (!port) {
        return;
    }

    // SECURITY FIX (HIGH-011): Use String copy constructor instead of raw new[]
    m_version = 4;
    const size_t size = port - addr;  // Length of host part
    m_host = String(addr, size);
    port++;  // Move past ':'

    // SECURITY FIX: Validate strtol result
    char *endptr = nullptr;
    const long portVal = strtol(port, &endptr, 10);
    if (endptr != port && portVal > 0 && portVal <= 65535) {
        m_port = static_cast<uint16_t>(portVal);
    }
}


void xmrig::BindHost::parseIPv6(const char *addr)
{
    const char *end = strchr(addr, ']');
    if (!end) {
        return;
    }

    const char *port = strchr(end, ':');
    if (!port) {
        return;
    }

    // SECURITY FIX (HIGH-011): Use String copy constructor instead of raw new[]
    m_version = 6;
    const size_t size = end - addr - 1;  // Length of content between brackets
    m_host = String(addr + 1, size);

    // SECURITY FIX: Validate strtol result
    char *endptr = nullptr;
    const long portVal = strtol(port + 1, &endptr, 10);
    if (endptr != (port + 1) && portVal > 0 && portVal <= 65535) {
        m_port = static_cast<uint16_t>(portVal);
    }
}
