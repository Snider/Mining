if (WITH_BLAKE3DCR)
    add_definitions(/DXMRIG_ALGO_BLAKE3DCR)

    list(APPEND HEADERS_CRYPTO
        src/crypto/blake3dcr/Blake3DCR.h
    )

    list(APPEND SOURCES_CRYPTO
        src/crypto/blake3dcr/Blake3DCR.cpp
    )

    # Add Blake3 library
    add_subdirectory(src/3rdparty/blake3)
    set(BLAKE3_LIBRARY blake3)
else()
    remove_definitions(/DXMRIG_ALGO_BLAKE3DCR)
endif()
