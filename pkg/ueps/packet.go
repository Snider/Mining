package ueps

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

// TLV Types
const (
	TagVersion     = 0x01
	TagCurrentLay  = 0x02
	TagTargetLay   = 0x03
	TagIntent      = 0x04
	TagThreatScore = 0x05
	TagHMAC        = 0x06 // The Signature
	TagPayload     = 0xFF // The Data
)

// UEPSHeader represents the conscious routing metadata
type UEPSHeader struct {
	Version      uint8  // Default 0x09
	CurrentLayer uint8
	TargetLayer  uint8
	IntentID     uint8  // Semantic Token
	ThreatScore  uint16 // 0-65535
}

// PacketBuilder helps construct a signed UEPS frame
type PacketBuilder struct {
	Header  UEPSHeader
	Payload []byte
}

// NewBuilder creates a packet context for a specific intent
func NewBuilder(intentID uint8, payload []byte) *PacketBuilder {
	return &PacketBuilder{
		Header: UEPSHeader{
			Version:      0x09, // IPv9
			CurrentLayer: 5,    // Application
			TargetLayer:  5,    // Application
			IntentID:     intentID,
			ThreatScore:  0, // Assumed innocent until proven guilty
		},
		Payload: payload,
	}
}

// MarshalAndSign generates the final byte stream using the shared secret
func (p *PacketBuilder) MarshalAndSign(sharedSecret []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. Write Standard Header Tags (0x01 - 0x05)
	// We write these first because they are part of what we sign.
	if err := writeTLV(buf, TagVersion, []byte{p.Header.Version}); err != nil {
		return nil, err
	}
	if err := writeTLV(buf, TagCurrentLay, []byte{p.Header.CurrentLayer}); err != nil {
		return nil, err
	}
	if err := writeTLV(buf, TagTargetLay, []byte{p.Header.TargetLayer}); err != nil {
		return nil, err
	}
	if err := writeTLV(buf, TagIntent, []byte{p.Header.IntentID}); err != nil {
		return nil, err
	}
	
	// Threat Score is uint16, needs binary packing
	tsBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(tsBuf, p.Header.ThreatScore)
	if err := writeTLV(buf, TagThreatScore, tsBuf); err != nil {
		return nil, err
	}

	// 2. Calculate HMAC
	// The signature covers: Existing Header TLVs + The Payload
	// It does NOT cover the HMAC TLV tag itself (obviously)
	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(buf.Bytes()) // The headers so far
	mac.Write(p.Payload)   // The data
	signature := mac.Sum(nil)

	// 3. Write HMAC TLV (0x06)
	// Length is 32 bytes for SHA256
	if err := writeTLV(buf, TagHMAC, signature); err != nil {
		return nil, err
	}

	// 4. Write Payload TLV (0xFF)
	// Note: 0xFF length is variable. For simplicity in this specialized reader,
	// we might handle 0xFF as "read until EOF" or use a varint length.
	// Implementing standard 1-byte length for payload is risky if payload > 255.
	// Assuming your spec allows >255 bytes, we handle 0xFF differently.
	
	buf.WriteByte(TagPayload)
	// We don't write a 1-byte length for payload here assuming stream mode,
	// but if strict TLV, we'd need a multi-byte length protocol.
	// For this snippet, simply appending data:
	buf.Write(p.Payload)

	return buf.Bytes(), nil
}

// Helper to write a simple TLV
func writeTLV(w io.Writer, tag uint8, value []byte) error {
	// Check strict length constraint (1 byte length = max 255 bytes)
	if len(value) > 255 {
		return errors.New("TLV value too large for 1-byte length header")
	}

	if _, err := w.Write([]byte{tag}); err != nil {
		return err
	}
	if _, err := w.Write([]byte{uint8(len(value))}); err != nil {
		return err
	}
	if _, err := w.Write(value); err != nil {
		return err
	}
	return nil
}
