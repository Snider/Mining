/*
 * BLAKE3 reference implementation
 * Based on https://github.com/BLAKE3-team/BLAKE3
 *
 * This is a minimal portable C implementation for Decred mining.
 */

#include "blake3.h"
#include <string.h>

// Initial vector
static const uint32_t IV[8] = {
    BLAKE3_IV_0, BLAKE3_IV_1, BLAKE3_IV_2, BLAKE3_IV_3,
    BLAKE3_IV_4, BLAKE3_IV_5, BLAKE3_IV_6, BLAKE3_IV_7
};

// Message schedule permutation
static const uint8_t MSG_SCHEDULE[7][16] = {
    {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
    {2, 6, 3, 10, 7, 0, 4, 13, 1, 11, 12, 5, 9, 14, 15, 8},
    {3, 4, 10, 12, 13, 2, 7, 14, 6, 5, 9, 0, 11, 15, 8, 1},
    {10, 7, 12, 9, 14, 3, 13, 15, 4, 0, 11, 2, 5, 8, 1, 6},
    {12, 13, 9, 11, 15, 10, 14, 8, 7, 2, 5, 3, 0, 1, 6, 4},
    {9, 14, 11, 5, 8, 12, 15, 1, 13, 3, 0, 10, 2, 6, 4, 7},
    {11, 15, 5, 0, 1, 9, 8, 6, 14, 10, 2, 12, 3, 4, 7, 13},
};

static inline uint32_t load32_le(const uint8_t *src) {
    return ((uint32_t)src[0]) | ((uint32_t)src[1] << 8) |
           ((uint32_t)src[2] << 16) | ((uint32_t)src[3] << 24);
}

static inline void store32_le(uint8_t *dst, uint32_t w) {
    dst[0] = (uint8_t)(w);
    dst[1] = (uint8_t)(w >> 8);
    dst[2] = (uint8_t)(w >> 16);
    dst[3] = (uint8_t)(w >> 24);
}

static inline uint32_t rotr32(uint32_t w, unsigned c) {
    return (w >> c) | (w << (32 - c));
}

// Quarter round
static inline void g(uint32_t *state, size_t a, size_t b, size_t c, size_t d,
                     uint32_t x, uint32_t y) {
    state[a] = state[a] + state[b] + x;
    state[d] = rotr32(state[d] ^ state[a], 16);
    state[c] = state[c] + state[d];
    state[b] = rotr32(state[b] ^ state[c], 12);
    state[a] = state[a] + state[b] + y;
    state[d] = rotr32(state[d] ^ state[a], 8);
    state[c] = state[c] + state[d];
    state[b] = rotr32(state[b] ^ state[c], 7);
}

static inline void round_fn(uint32_t state[16], const uint32_t *msg, size_t round) {
    const uint8_t *schedule = MSG_SCHEDULE[round];
    g(state, 0, 4, 8, 12, msg[schedule[0]], msg[schedule[1]]);
    g(state, 1, 5, 9, 13, msg[schedule[2]], msg[schedule[3]]);
    g(state, 2, 6, 10, 14, msg[schedule[4]], msg[schedule[5]]);
    g(state, 3, 7, 11, 15, msg[schedule[6]], msg[schedule[7]]);
    g(state, 0, 5, 10, 15, msg[schedule[8]], msg[schedule[9]]);
    g(state, 1, 6, 11, 12, msg[schedule[10]], msg[schedule[11]]);
    g(state, 2, 7, 8, 13, msg[schedule[12]], msg[schedule[13]]);
    g(state, 3, 4, 9, 14, msg[schedule[14]], msg[schedule[15]]);
}

static void compress_pre(uint32_t state[16], const uint32_t cv[8],
                         const uint8_t block[BLAKE3_BLOCK_LEN],
                         uint8_t block_len, uint64_t counter, uint8_t flags) {
    uint32_t msg[16];
    for (size_t i = 0; i < 16; i++) {
        msg[i] = load32_le(block + 4 * i);
    }

    state[0] = cv[0];
    state[1] = cv[1];
    state[2] = cv[2];
    state[3] = cv[3];
    state[4] = cv[4];
    state[5] = cv[5];
    state[6] = cv[6];
    state[7] = cv[7];
    state[8] = IV[0];
    state[9] = IV[1];
    state[10] = IV[2];
    state[11] = IV[3];
    state[12] = (uint32_t)counter;
    state[13] = (uint32_t)(counter >> 32);
    state[14] = block_len;
    state[15] = flags;

    for (size_t round = 0; round < 7; round++) {
        round_fn(state, msg, round);
    }
}

static void compress_in_place(uint32_t cv[8], const uint8_t block[BLAKE3_BLOCK_LEN],
                              uint8_t block_len, uint64_t counter, uint8_t flags) {
    uint32_t state[16];
    compress_pre(state, cv, block, block_len, counter, flags);
    cv[0] = state[0] ^ state[8];
    cv[1] = state[1] ^ state[9];
    cv[2] = state[2] ^ state[10];
    cv[3] = state[3] ^ state[11];
    cv[4] = state[4] ^ state[12];
    cv[5] = state[5] ^ state[13];
    cv[6] = state[6] ^ state[14];
    cv[7] = state[7] ^ state[15];
}

static void compress_xof(const uint32_t cv[8], const uint8_t block[BLAKE3_BLOCK_LEN],
                         uint8_t block_len, uint64_t counter, uint8_t flags,
                         uint8_t out[64]) {
    uint32_t state[16];
    compress_pre(state, cv, block, block_len, counter, flags);
    for (size_t i = 0; i < 8; i++) {
        store32_le(&out[i * 4], state[i] ^ state[i + 8]);
    }
    for (size_t i = 8; i < 16; i++) {
        store32_le(&out[i * 4], state[i] ^ cv[i - 8]);
    }
}

// Chunk state functions
static void chunk_state_init(blake3_chunk_state *self, const uint32_t key[8], uint8_t flags) {
    memcpy(self->cv, key, BLAKE3_KEY_LEN);
    self->chunk_counter = 0;
    memset(self->buf, 0, BLAKE3_BLOCK_LEN);
    self->buf_len = 0;
    self->blocks_compressed = 0;
    self->flags = flags;
}

static size_t chunk_state_len(const blake3_chunk_state *self) {
    return (BLAKE3_BLOCK_LEN * (size_t)self->blocks_compressed) + (size_t)self->buf_len;
}

static size_t chunk_state_fill_buf(blake3_chunk_state *self, const uint8_t *input, size_t input_len) {
    size_t take = BLAKE3_BLOCK_LEN - (size_t)self->buf_len;
    if (take > input_len) {
        take = input_len;
    }
    memcpy(self->buf + self->buf_len, input, take);
    self->buf_len += (uint8_t)take;
    return take;
}

static uint8_t chunk_state_maybe_start_flag(const blake3_chunk_state *self) {
    if (self->blocks_compressed == 0) {
        return CHUNK_START;
    }
    return 0;
}

static void chunk_state_update(blake3_chunk_state *self, const uint8_t *input, size_t input_len) {
    if (self->buf_len > 0) {
        size_t take = chunk_state_fill_buf(self, input, input_len);
        input += take;
        input_len -= take;
        if (input_len > 0) {
            compress_in_place(self->cv, self->buf, BLAKE3_BLOCK_LEN,
                              self->chunk_counter, self->flags | chunk_state_maybe_start_flag(self));
            self->blocks_compressed++;
            self->buf_len = 0;
            memset(self->buf, 0, BLAKE3_BLOCK_LEN);
        }
    }

    while (input_len > BLAKE3_BLOCK_LEN) {
        compress_in_place(self->cv, input, BLAKE3_BLOCK_LEN,
                          self->chunk_counter, self->flags | chunk_state_maybe_start_flag(self));
        self->blocks_compressed++;
        input += BLAKE3_BLOCK_LEN;
        input_len -= BLAKE3_BLOCK_LEN;
    }

    size_t take = chunk_state_fill_buf(self, input, input_len);
    (void)take;
}

static void chunk_state_output(const blake3_chunk_state *self, uint8_t out[BLAKE3_OUT_LEN]) {
    uint8_t block_flags = self->flags | chunk_state_maybe_start_flag(self) | CHUNK_END;
    uint8_t wide_buf[64];
    compress_xof(self->cv, self->buf, self->buf_len, self->chunk_counter, block_flags, wide_buf);
    memcpy(out, wide_buf, BLAKE3_OUT_LEN);
}

// Hasher functions
void blake3_hasher_init(blake3_hasher *self) {
    memcpy(self->key, IV, BLAKE3_KEY_LEN);
    chunk_state_init(&self->chunk, IV, 0);
    self->cv_stack_len = 0;
}

void blake3_hasher_init_keyed(blake3_hasher *self, const uint8_t key[BLAKE3_KEY_LEN]) {
    uint32_t key_words[8];
    for (size_t i = 0; i < 8; i++) {
        key_words[i] = load32_le(key + 4 * i);
    }
    memcpy(self->key, key_words, BLAKE3_KEY_LEN);
    chunk_state_init(&self->chunk, key_words, KEYED_HASH);
    self->cv_stack_len = 0;
}

void blake3_hasher_init_derive_key(blake3_hasher *self, const char *context) {
    blake3_hasher_init_derive_key_raw(self, context, strlen(context));
}

void blake3_hasher_init_derive_key_raw(blake3_hasher *self, const void *context, size_t context_len) {
    blake3_hasher context_hasher;
    memcpy(context_hasher.key, IV, BLAKE3_KEY_LEN);
    chunk_state_init(&context_hasher.chunk, IV, DERIVE_KEY_CONTEXT);
    context_hasher.cv_stack_len = 0;
    blake3_hasher_update(&context_hasher, context, context_len);
    uint8_t context_key[BLAKE3_KEY_LEN];
    blake3_hasher_finalize(&context_hasher, context_key, BLAKE3_KEY_LEN);
    uint32_t context_key_words[8];
    for (size_t i = 0; i < 8; i++) {
        context_key_words[i] = load32_le(context_key + 4 * i);
    }
    memcpy(self->key, context_key_words, BLAKE3_KEY_LEN);
    chunk_state_init(&self->chunk, context_key_words, DERIVE_KEY_MATERIAL);
    self->cv_stack_len = 0;
}

static void hasher_merge_cv_stack(blake3_hasher *self, uint64_t total_len) {
    size_t post_merge_stack_len = (size_t)__builtin_popcountll(total_len);
    while (self->cv_stack_len > post_merge_stack_len) {
        uint8_t *parent_block = &self->cv_stack[(self->cv_stack_len - 2) * BLAKE3_OUT_LEN];
        uint32_t parent_cv[8];
        memcpy(parent_cv, self->key, BLAKE3_KEY_LEN);
        compress_in_place(parent_cv, parent_block, BLAKE3_BLOCK_LEN, 0, self->chunk.flags | PARENT);
        self->cv_stack_len--;
        for (size_t i = 0; i < 8; i++) {
            store32_le(&self->cv_stack[(self->cv_stack_len - 1) * BLAKE3_OUT_LEN + 4 * i], parent_cv[i]);
        }
    }
}

static void hasher_push_cv(blake3_hasher *self, uint8_t new_cv[BLAKE3_OUT_LEN], uint64_t chunk_counter) {
    hasher_merge_cv_stack(self, chunk_counter);
    memcpy(&self->cv_stack[self->cv_stack_len * BLAKE3_OUT_LEN], new_cv, BLAKE3_OUT_LEN);
    self->cv_stack_len++;
}

void blake3_hasher_update(blake3_hasher *self, const void *input, size_t input_len) {
    const uint8_t *input_bytes = (const uint8_t *)input;

    while (input_len > 0) {
        if (chunk_state_len(&self->chunk) == BLAKE3_CHUNK_LEN) {
            uint8_t chunk_cv[32];
            chunk_state_output(&self->chunk, chunk_cv);
            uint64_t total_chunks = self->chunk.chunk_counter + 1;
            hasher_push_cv(self, chunk_cv, total_chunks);
            chunk_state_init(&self->chunk, self->key, self->chunk.flags);
            self->chunk.chunk_counter = total_chunks;
        }

        size_t want = BLAKE3_CHUNK_LEN - chunk_state_len(&self->chunk);
        size_t take = want < input_len ? want : input_len;
        chunk_state_update(&self->chunk, input_bytes, take);
        input_bytes += take;
        input_len -= take;
    }
}

void blake3_hasher_finalize(const blake3_hasher *self, uint8_t *out, size_t out_len) {
    blake3_hasher_finalize_seek(self, 0, out, out_len);
}

void blake3_hasher_finalize_seek(const blake3_hasher *self, uint64_t seek, uint8_t *out, size_t out_len) {
    if (out_len == 0) {
        return;
    }

    // If the subtree stack is empty, the output is in the chunk state
    if (self->cv_stack_len == 0) {
        uint8_t block_flags = self->chunk.flags | chunk_state_maybe_start_flag(&self->chunk) | CHUNK_END | ROOT;
        uint64_t output_block_counter = seek / 64;
        size_t offset_within_block = seek % 64;
        uint8_t wide_buf[64];
        compress_xof(self->chunk.cv, self->chunk.buf, self->chunk.buf_len,
                     output_block_counter, block_flags, wide_buf);
        size_t available_bytes = 64 - offset_within_block;
        size_t copy_len = out_len < available_bytes ? out_len : available_bytes;
        memcpy(out, wide_buf + offset_within_block, copy_len);
        out += copy_len;
        out_len -= copy_len;
        output_block_counter++;

        while (out_len > 0) {
            compress_xof(self->chunk.cv, self->chunk.buf, self->chunk.buf_len,
                         output_block_counter, block_flags, wide_buf);
            copy_len = out_len < 64 ? out_len : 64;
            memcpy(out, wide_buf, copy_len);
            out += copy_len;
            out_len -= copy_len;
            output_block_counter++;
        }
        return;
    }

    // Otherwise merge the CV stack and finalize
    uint8_t chunk_cv[32];
    chunk_state_output(&self->chunk, chunk_cv);

    uint8_t parent_block[BLAKE3_BLOCK_LEN];
    uint32_t parent_cv[8];
    size_t num_cvs = self->cv_stack_len;
    memcpy(parent_block + BLAKE3_OUT_LEN, chunk_cv, BLAKE3_OUT_LEN);

    while (num_cvs > 0) {
        memcpy(parent_block, &self->cv_stack[(num_cvs - 1) * BLAKE3_OUT_LEN], BLAKE3_OUT_LEN);
        memcpy(parent_cv, self->key, BLAKE3_KEY_LEN);
        uint8_t flags = self->chunk.flags | PARENT;
        if (num_cvs == 1) {
            flags |= ROOT;
        }
        compress_in_place(parent_cv, parent_block, BLAKE3_BLOCK_LEN, 0, flags);
        num_cvs--;
        for (size_t i = 0; i < 8; i++) {
            store32_le(&parent_block[BLAKE3_OUT_LEN + 4 * i], parent_cv[i]);
        }
    }

    memcpy(out, parent_block + BLAKE3_OUT_LEN, out_len < BLAKE3_OUT_LEN ? out_len : BLAKE3_OUT_LEN);
}

void blake3_hasher_reset(blake3_hasher *self) {
    chunk_state_init(&self->chunk, self->key, self->chunk.flags & (KEYED_HASH | DERIVE_KEY_MATERIAL));
    self->cv_stack_len = 0;
}

// Simple one-shot hash for mining
void blake3_hash(const void *input, size_t input_len, uint8_t out[BLAKE3_OUT_LEN]) {
    blake3_hasher hasher;
    blake3_hasher_init(&hasher);
    blake3_hasher_update(&hasher, input, input_len);
    blake3_hasher_finalize(&hasher, out, BLAKE3_OUT_LEN);
}
