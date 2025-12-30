# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is a **binary distribution repository** containing precompiled static libraries and headers for XMRig cryptocurrency mining software. There is no source code to build, test, or lint - only precompiled binaries.

### Included Dependencies

- **libuv** 1.51.0 - Asynchronous I/O library
- **OpenSSL** 3.0.16 - Cryptography and SSL/TLS
- **hwloc** 2.12.1 - Hardware topology library

## Directory Structure

```
<compiler>/<arch>/
├── include/    # Header files (hwloc.h, uv.h, openssl/)
└── lib/        # Static libraries (.a for Unix, .lib for Windows)
```

**Supported platforms:**
- `gcc/x64`, `gcc/x86` - MSYS2/MinGW
- `clang/arm64` - Clang ARM64
- `msvc2015/x64`, `msvc2015/x86` - Visual Studio 2015
- `msvc2017/x64`, `msvc2017/x86` - Visual Studio 2017
- `msvc2019/x64` - Visual Studio 2019
- `msvc2022/x64`, `msvc2022/arm64` - Visual Studio 2022

## Usage with XMRig

When building XMRig from source, point CMake to the appropriate directory:

```bash
# Visual Studio 2022 x64
cmake .. -G "Visual Studio 17 2022" -A x64 -DXMRIG_DEPS=c:\xmrig-deps\msvc2022\x64

# Visual Studio 2019 x64
cmake .. -G "Visual Studio 16 2019" -A x64 -DXMRIG_DEPS=c:\xmrig-deps\msvc2019\x64

# MSYS2 x64
cmake .. -G "Unix Makefiles" -DXMRIG_DEPS=c:/xmrig-deps/gcc/x64
```
