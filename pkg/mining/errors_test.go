package mining

import (
	"errors"
	"net/http"
	"testing"
)

func TestMiningError_Error(t *testing.T) {
	err := NewMiningError(ErrCodeMinerNotFound, "miner not found")
	expected := "MINER_NOT_FOUND: miner not found"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestMiningError_ErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewMiningError(ErrCodeStartFailed, "failed to start").WithCause(cause)

	// Should include cause in error message
	if err.Cause != cause {
		t.Error("Cause was not set")
	}

	// Should be unwrappable
	if errors.Unwrap(err) != cause {
		t.Error("Unwrap did not return cause")
	}
}

func TestMiningError_WithDetails(t *testing.T) {
	err := NewMiningError(ErrCodeInvalidConfig, "invalid config").
		WithDetails("port must be between 1024 and 65535")

	if err.Details != "port must be between 1024 and 65535" {
		t.Errorf("Details not set correctly: %s", err.Details)
	}
}

func TestMiningError_WithSuggestion(t *testing.T) {
	err := NewMiningError(ErrCodeConnectionFailed, "connection failed").
		WithSuggestion("check your network")

	if err.Suggestion != "check your network" {
		t.Errorf("Suggestion not set correctly: %s", err.Suggestion)
	}
}

func TestMiningError_StatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *MiningError
		expected int
	}{
		{"default", NewMiningError("TEST", "test"), http.StatusInternalServerError},
		{"not found", ErrMinerNotFound("test"), http.StatusNotFound},
		{"conflict", ErrMinerExists("test"), http.StatusConflict},
		{"bad request", ErrInvalidConfig("bad"), http.StatusBadRequest},
		{"service unavailable", ErrConnectionFailed("pool"), http.StatusServiceUnavailable},
		{"timeout", ErrTimeout("operation"), http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, tt.err.StatusCode())
			}
		})
	}
}

func TestMiningError_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       *MiningError
		retryable bool
	}{
		{"not found", ErrMinerNotFound("test"), false},
		{"exists", ErrMinerExists("test"), false},
		{"invalid config", ErrInvalidConfig("bad"), false},
		{"install failed", ErrInstallFailed("xmrig"), true},
		{"start failed", ErrStartFailed("test"), true},
		{"connection failed", ErrConnectionFailed("pool"), true},
		{"timeout", ErrTimeout("operation"), true},
		{"database error", ErrDatabaseError("query"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.IsRetryable() != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, tt.err.IsRetryable())
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *MiningError
		code string
	}{
		{"ErrMinerNotFound", ErrMinerNotFound("test"), ErrCodeMinerNotFound},
		{"ErrMinerExists", ErrMinerExists("test"), ErrCodeMinerExists},
		{"ErrMinerNotRunning", ErrMinerNotRunning("test"), ErrCodeMinerNotRunning},
		{"ErrInstallFailed", ErrInstallFailed("xmrig"), ErrCodeInstallFailed},
		{"ErrStartFailed", ErrStartFailed("test"), ErrCodeStartFailed},
		{"ErrStopFailed", ErrStopFailed("test"), ErrCodeStopFailed},
		{"ErrInvalidConfig", ErrInvalidConfig("bad port"), ErrCodeInvalidConfig},
		{"ErrUnsupportedMiner", ErrUnsupportedMiner("unknown"), ErrCodeUnsupportedMiner},
		{"ErrConnectionFailed", ErrConnectionFailed("pool:3333"), ErrCodeConnectionFailed},
		{"ErrTimeout", ErrTimeout("GetStats"), ErrCodeTimeout},
		{"ErrDatabaseError", ErrDatabaseError("insert"), ErrCodeDatabaseError},
		{"ErrProfileNotFound", ErrProfileNotFound("abc123"), ErrCodeProfileNotFound},
		{"ErrProfileExists", ErrProfileExists("My Profile"), ErrCodeProfileExists},
		{"ErrInternal", ErrInternal("unexpected error"), ErrCodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, tt.err.Code)
			}
			if tt.err.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}

func TestMiningError_Chaining(t *testing.T) {
	cause := errors.New("network timeout")
	err := ErrConnectionFailed("pool:3333").
		WithCause(cause).
		WithDetails("timeout after 30s").
		WithSuggestion("check firewall settings")

	if err.Code != ErrCodeConnectionFailed {
		t.Errorf("Code changed: %s", err.Code)
	}
	if err.Cause != cause {
		t.Error("Cause not set")
	}
	if err.Details != "timeout after 30s" {
		t.Errorf("Details not set: %s", err.Details)
	}
	if err.Suggestion != "check firewall settings" {
		t.Errorf("Suggestion not set: %s", err.Suggestion)
	}
}
