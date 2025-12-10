package mining

import (
	"errors"
)

// RawConfig is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can be used to delay JSON decoding or precompute a JSON encoding.
// We define it as []byte (like json.RawMessage) to avoid swagger parsing issues with the json package.
type RawConfig []byte

// MiningProfile represents a saved configuration for running a specific miner.
// It decouples the UI from the underlying miner's specific config structure.
type MiningProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MinerType string    `json:"minerType"`                   // e.g., "xmrig", "ttminer"
	Config    RawConfig `json:"config" swaggertype:"object"` // The raw JSON config for the specific miner
}

// MarshalJSON returns m as the JSON encoding of m.
func (m RawConfig) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawConfig) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("RawConfig: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}
