/* XMRig
 * Copyright (c) 2018-2021 SChernykh   <https://github.com/SChernykh>
 * Copyright (c) 2016-2021 XMRig       <https://github.com/xmrig>, <support@xmrig.com>
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

#include "base/io/Console.h"
#include "base/kernel/interfaces/IConsoleListener.h"
#include "base/tools/Handle.h"


xmrig::Console::Console(IConsoleListener *listener)
    : m_listener(listener)
{
    if (!isSupported()) {
        return;
    }

    m_tty = new uv_tty_t;
    m_tty->data = this;
    uv_tty_init(uv_default_loop(), m_tty, 0, 1);

    if (!uv_is_readable(reinterpret_cast<uv_stream_t*>(m_tty))) {
        // SECURITY: Clean up allocated handle to prevent memory leak
        Handle::close(m_tty);
        m_tty = nullptr;
        return;
    }

    uv_tty_set_mode(m_tty, UV_TTY_MODE_RAW);
    uv_read_start(reinterpret_cast<uv_stream_t*>(m_tty), Console::onAllocBuffer, Console::onRead);
}


xmrig::Console::~Console()
{
    // SECURITY: Check return value - failure could leave terminal in raw mode
    const int rc = uv_tty_reset_mode();
    if (rc < 0) {
        // Note: Can't use LOG here as it may already be destroyed
        // Best effort: write to stderr directly
        fprintf(stderr, "Warning: uv_tty_reset_mode() failed: %s\n", uv_strerror(rc));
    }

    Handle::close(m_tty);
}


bool xmrig::Console::isSupported()
{
    const uv_handle_type type = uv_guess_handle(0);
    return type == UV_TTY || type == UV_NAMED_PIPE;
}


void xmrig::Console::onAllocBuffer(uv_handle_t *handle, size_t, uv_buf_t *buf)
{
    auto console = static_cast<Console*>(handle->data);
    buf->len  = 1;
    buf->base = console->m_buf;
}


void xmrig::Console::onRead(uv_stream_t *stream, ssize_t nread, const uv_buf_t *buf)
{
    if (nread < 0) {
        // SECURITY: Check if already closing to prevent double-close
        // Handle::close() checks uv_is_closing() but explicit check is clearer
        uv_handle_t *handle = reinterpret_cast<uv_handle_t*>(stream);
        if (!uv_is_closing(handle)) {
            uv_close(handle, nullptr);
        }
        return;
    }

    if (nread == 1) {
        static_cast<Console*>(stream->data)->m_listener->onConsoleCommand(buf->base[0]);
    }
}
