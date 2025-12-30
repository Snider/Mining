if (WITH_ETCHASH)
    add_definitions(/DXMRIG_ALGO_ETCHASH)

    list(APPEND HEADERS_CRYPTO
        src/crypto/etchash/ETChash.h
        src/crypto/etchash/ETCCache.h
    )

    list(APPEND SOURCES_CRYPTO
        src/crypto/etchash/ETChash.cpp
        src/crypto/etchash/ETCCache.cpp
    )

    # ETChash uses the same libethash library as KawPow
    if (NOT WITH_KAWPOW)
        add_subdirectory(src/3rdparty/libethash)
        set(ETHASH_LIBRARY ethash)
    endif()
else()
    remove_definitions(/DXMRIG_ALGO_ETCHASH)
endif()
