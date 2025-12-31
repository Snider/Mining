package node

import (
	"fmt"
)

// ProtocolError represents an error from the remote peer.
type ProtocolError struct {
	Code    int
	Message string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("remote error (%d): %s", e.Code, e.Message)
}

// ResponseHandler provides helpers for handling protocol responses.
type ResponseHandler struct{}

// ValidateResponse checks if the response is valid and returns a parsed error if it's an error response.
// It checks:
// 1. If response is nil (returns error)
// 2. If response is an error message (returns ProtocolError)
// 3. If response type matches expected (returns error if not)
func (h *ResponseHandler) ValidateResponse(resp *Message, expectedType MessageType) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	// Check for error response
	if resp.Type == MsgError {
		var errPayload ErrorPayload
		if err := resp.ParsePayload(&errPayload); err != nil {
			return &ProtocolError{Code: ErrCodeUnknown, Message: "unable to parse error response"}
		}
		return &ProtocolError{Code: errPayload.Code, Message: errPayload.Message}
	}

	// Check expected type
	if resp.Type != expectedType {
		return fmt.Errorf("unexpected response type: expected %s, got %s", expectedType, resp.Type)
	}

	return nil
}

// ParseResponse validates the response and parses the payload into the target.
// This combines ValidateResponse and ParsePayload into a single call.
func (h *ResponseHandler) ParseResponse(resp *Message, expectedType MessageType, target interface{}) error {
	if err := h.ValidateResponse(resp, expectedType); err != nil {
		return err
	}

	if target != nil {
		if err := resp.ParsePayload(target); err != nil {
			return fmt.Errorf("failed to parse %s payload: %w", expectedType, err)
		}
	}

	return nil
}

// DefaultResponseHandler is the default response handler instance.
var DefaultResponseHandler = &ResponseHandler{}

// ValidateResponse is a convenience function using the default handler.
func ValidateResponse(resp *Message, expectedType MessageType) error {
	return DefaultResponseHandler.ValidateResponse(resp, expectedType)
}

// ParseResponse is a convenience function using the default handler.
func ParseResponse(resp *Message, expectedType MessageType, target interface{}) error {
	return DefaultResponseHandler.ParseResponse(resp, expectedType, target)
}

// IsProtocolError returns true if the error is a ProtocolError.
func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}

// GetProtocolErrorCode returns the error code if err is a ProtocolError, otherwise returns 0.
func GetProtocolErrorCode(err error) int {
	if pe, ok := err.(*ProtocolError); ok {
		return pe.Code
	}
	return 0
}
