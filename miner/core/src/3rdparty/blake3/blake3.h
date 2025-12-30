/*
 * BLAKE3 reference implementation
 * Based on https://github.com/BLAKE3-team/BLAKE3
 *
 * This is a minimal implementation for Decred mining.
 * For optimal performance, consider using SIMD-accelerated versions.
 */

#ifndef BLAKE3_H
#define BLAKE3_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define BLAKE3_VERSION_STRING "1.8.2"
#define BLAKE3_KEY_LEN 32
#define BLAKE3_OUT_LEN 32
#define BLAKE3_BLOCK_LEN 64
#define BLAKE3_CHUNK_LEN 1024
#define BLAKE3_MAX_DEPTH 54

// Initial vector (same as BLAKE2s)
#define BLAKE3_IV_0 0x6A09E667UL
#define BLAKE3_IV_1 0xBB67AE85UL
#define BLAKE3_IV_2 0x3C6EF372UL
#define BLAKE3_IV_3 0xA54FF53AUL
#define BLAKE3_IV_4 0x510E527FUL
#define BLAKE3_IV_5 0x9B05688CUL
#define BLAKE3_IV_6 0x1F83D9ABUL
#define BLAKE3_IV_7 0x5BE0CD19UL

// Flags for domain separation
enum blake3_flags {
    CHUNK_START         = 1 << 0,
    CHUNK_END           = 1 << 1,
    PARENT              = 1 << 2,
    ROOT                = 1 << 3,
    KEYED_HASH          = 1 << 4,
    DERIVE_KEY_CONTEXT  = 1 << 5,
    DERIVE_KEY_MATERIAL = 1 << 6,
};

typedef struct {
    uint32_t cv[8];
    uint64_t chunk_counter;
    uint8_t buf[BLAKE3_BLOCK_LEN];
    uint8_t buf_len;
    uint8_t blocks_compressed;
    uint8_t flags;
} blake3_chunk_state;

typedef struct {
    uint32_t key[8];
    blake3_chunk_state chunk;
    uint8_t cv_stack_len;
    uint8_t cv_stack[(BLAKE3_MAX_DEPTH + 1) * BLAKE3_OUT_LEN];
} blake3_hasher;

// Initialize hasher
void blake3_hasher_init(blake3_hasher *self);
void blake3_hasher_init_keyed(blake3_hasher *self, const uint8_t key[BLAKE3_KEY_LEN]);
void blake3_hasher_init_derive_key(blake3_hasher *self, const char *context);
void blake3_hasher_init_derive_key_raw(blake3_hasher *self, const void *context, size_t context_len);

// Update with input data
void blake3_hasher_update(blake3_hasher *self, const void *input, size_t input_len);

// Finalize and output
void blake3_hasher_finalize(const blake3_hasher *self, uint8_t *out, size_t out_len);
void blake3_hasher_finalize_seek(const blake3_hasher *self, uint64_t seek, uint8_t *out, size_t out_len);

// Reset hasher for reuse
void blake3_hasher_reset(blake3_hasher *self);

// Simple one-shot hash function for mining
void blake3_hash(const void *input, size_t input_len, uint8_t out[BLAKE3_OUT_LEN]);

#ifdef __cplusplus
}
#endif

#endif // BLAKE3_H
