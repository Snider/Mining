if (WITH_PROGPOWZ)
    add_definitions(/DXMRIG_ALGO_PROGPOWZ)

    list(APPEND HEADERS_CRYPTO
        src/crypto/progpowz/ProgPowZHash.h
        src/crypto/progpowz/ProgPowZCache.h
    )

    list(APPEND SOURCES_CRYPTO
        src/crypto/progpowz/ProgPowZHash.cpp
        src/crypto/progpowz/ProgPowZCache.cpp
    )

    # ProgPowZ uses the same libethash library as KawPow and ETChash
    if (NOT WITH_KAWPOW AND NOT WITH_ETCHASH)
        add_subdirectory(src/3rdparty/libethash)
        set(ETHASH_LIBRARY ethash)
    endif()
else()
    remove_definitions(/DXMRIG_ALGO_PROGPOWZ)
endif()
