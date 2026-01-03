package ueps

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ParsedPacket holds the verified data
type ParsedPacket struct {
	Header  UEPSHeader
	Payload []byte
}

// ReadAndVerify reads a UEPS frame from the stream and validates the HMAC.
// It consumes the stream up to the end of the packet.
func ReadAndVerify(r *bufio.Reader, sharedSecret []byte) (*ParsedPacket, error) {
	// Buffer to reconstruct the data for HMAC verification
	// We have to "record" what we read to verify the signature later.
	var signedData bytes.Buffer
	header := UEPSHeader{}
	var signature []byte
	var payload []byte

	// Loop through TLVs until we hit Payload (0xFF) or EOF
	for {
		// 1. Read Tag
		tag, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		// 2. Handle Payload Tag (0xFF) - The Exit Condition
		if tag == TagPayload {
			// Stop recording signedData here (HMAC covers headers + payload, but logic splits)
			// Actually, wait. The HMAC covers (Headers + Payload).
			// We need to read the payload to verify.
			
			// For this implementation, we read until EOF or a specific delimiter?
			// In a TCP stream, we need a length. 
			// If you are using standard TCP, you typically prefix the WHOLE frame with 
			// a 4-byte length. Assuming you handle that framing *before* calling this.
			
			// Reading the rest as payload:
			remaining, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			payload = remaining
			
			// Add 0xFF and payload to the buffer for signature check?
			// NO. In MarshalAndSign:
			// mac.Write(buf.Bytes()) // Headers
			// mac.Write(p.Payload)   // Data
			// It did NOT write the 0xFF tag into the HMAC.
			
			break // Exit loop
		}

		// 3. Read Length (Standard TLV)
		lengthByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		length := int(lengthByte)

		// 4. Read Value
		value := make([]byte, length)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}

		// Store for processing
		switch tag {
		case TagVersion:
			header.Version = value[0]
			// Reconstruct signed data: Tag + Len + Val
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		case TagCurrentLay:
			header.CurrentLayer = value[0]
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		case TagTargetLay:
			header.TargetLayer = value[0]
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		case TagIntent:
			header.IntentID = value[0]
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		case TagThreatScore:
			header.ThreatScore = binary.BigEndian.Uint16(value)
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		case TagHMAC:
			signature = value
			// We do NOT add the HMAC itself to signedData
		default:
			// Unknown tag (future proofing), verify it but ignore semantics
			signedData.WriteByte(tag)
			signedData.WriteByte(byte(length))
			signedData.Write(value)
		}
	}

	if len(signature) == 0 {
		return nil, errors.New("UEPS packet missing HMAC signature")
	}

	// 5. Verify HMAC
	// Reconstruct: Headers (signedData) + Payload
	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(signedData.Bytes())
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(signature, expectedMAC) {
		// Log this. This is a Threat Event.
		// "Axiom Violation: Integrity Check Failed"
		return nil, fmt.Errorf("integrity violation: HMAC mismatch (ThreatScore +100)")
	}

	return &ParsedPacket{
		Header:  header,
		Payload: payload,
	}, nil
}
