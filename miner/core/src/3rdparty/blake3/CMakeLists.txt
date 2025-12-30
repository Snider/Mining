cmake_minimum_required(VERSION 3.5)
project(blake3 C)

set(BLAKE3_SOURCES
    blake3.c
)

set(BLAKE3_HEADERS
    blake3.h
)

add_library(blake3 STATIC ${BLAKE3_SOURCES} ${BLAKE3_HEADERS})

target_include_directories(blake3 PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
