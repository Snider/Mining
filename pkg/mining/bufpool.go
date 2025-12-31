package mining

import (
	"bytes"
	"encoding/json"
	"sync"
)

// bufferPool provides reusable byte buffers for JSON encoding.
// This reduces allocation overhead in hot paths like WebSocket event serialization.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// getBuffer retrieves a buffer from the pool.
func getBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// putBuffer returns a buffer to the pool.
func putBuffer(buf *bytes.Buffer) {
	// Don't pool buffers that grew too large (>64KB)
	if buf.Cap() <= 65536 {
		bufferPool.Put(buf)
	}
}

// MarshalJSON encodes a value to JSON using a pooled buffer.
// Returns a copy of the encoded bytes (safe to use after the function returns).
func MarshalJSON(v interface{}) ([]byte, error) {
	buf := getBuffer()
	defer putBuffer(buf)

	enc := json.NewEncoder(buf)
	// Don't escape HTML characters (matches json.Marshal behavior for these use cases)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	// json.Encoder.Encode adds a newline; remove it to match json.Marshal
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	// Return a copy since the buffer will be reused
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}
