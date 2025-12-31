package mining

import (
	"fmt"
	"net/http"
)

// Error codes for the mining package
const (
	ErrCodeMinerNotFound      = "MINER_NOT_FOUND"
	ErrCodeMinerExists        = "MINER_EXISTS"
	ErrCodeMinerNotRunning    = "MINER_NOT_RUNNING"
	ErrCodeInstallFailed      = "INSTALL_FAILED"
	ErrCodeStartFailed        = "START_FAILED"
	ErrCodeStopFailed         = "STOP_FAILED"
	ErrCodeInvalidConfig      = "INVALID_CONFIG"
	ErrCodeInvalidInput       = "INVALID_INPUT"
	ErrCodeUnsupportedMiner   = "UNSUPPORTED_MINER"
	ErrCodeNotSupported       = "NOT_SUPPORTED"
	ErrCodeConnectionFailed   = "CONNECTION_FAILED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeProfileNotFound    = "PROFILE_NOT_FOUND"
	ErrCodeProfileExists      = "PROFILE_EXISTS"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeInternal           = "INTERNAL_ERROR" // Alias for consistency
)

// MiningError is a structured error type for the mining package
type MiningError struct {
	Code       string // Machine-readable error code
	Message    string // Human-readable message
	Details    string // Technical details (for debugging)
	Suggestion string // What to do next
	Retryable  bool   // Can the client retry?
	HTTPStatus int    // HTTP status code to return
	Cause      error  // Underlying error
}

// Error implements the error interface
func (e *MiningError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *MiningError) Unwrap() error {
	return e.Cause
}

// WithCause adds an underlying error
func (e *MiningError) WithCause(err error) *MiningError {
	e.Cause = err
	return e
}

// WithDetails adds technical details
func (e *MiningError) WithDetails(details string) *MiningError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion for the user
func (e *MiningError) WithSuggestion(suggestion string) *MiningError {
	e.Suggestion = suggestion
	return e
}

// IsRetryable returns whether the error is retryable
func (e *MiningError) IsRetryable() bool {
	return e.Retryable
}

// StatusCode returns the HTTP status code for this error
func (e *MiningError) StatusCode() int {
	if e.HTTPStatus == 0 {
		return http.StatusInternalServerError
	}
	return e.HTTPStatus
}

// NewMiningError creates a new MiningError
func NewMiningError(code, message string) *MiningError {
	return &MiningError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// Predefined error constructors for common errors

// ErrMinerNotFound creates a miner not found error
func ErrMinerNotFound(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeMinerNotFound,
		Message:    fmt.Sprintf("miner '%s' not found", name),
		Suggestion: "Check that the miner name is correct and that it is running",
		Retryable:  false,
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrMinerExists creates a miner already exists error
func ErrMinerExists(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeMinerExists,
		Message:    fmt.Sprintf("miner '%s' is already running", name),
		Suggestion: "Stop the existing miner first or use a different configuration",
		Retryable:  false,
		HTTPStatus: http.StatusConflict,
	}
}

// ErrMinerNotRunning creates a miner not running error
func ErrMinerNotRunning(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeMinerNotRunning,
		Message:    fmt.Sprintf("miner '%s' is not running", name),
		Suggestion: "Start the miner first before performing this operation",
		Retryable:  false,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrInstallFailed creates an installation failed error
func ErrInstallFailed(minerType string) *MiningError {
	return &MiningError{
		Code:       ErrCodeInstallFailed,
		Message:    fmt.Sprintf("failed to install %s", minerType),
		Suggestion: "Check your internet connection and try again",
		Retryable:  true,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ErrStartFailed creates a start failed error
func ErrStartFailed(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeStartFailed,
		Message:    fmt.Sprintf("failed to start miner '%s'", name),
		Suggestion: "Check the miner configuration and logs for details",
		Retryable:  true,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ErrStopFailed creates a stop failed error
func ErrStopFailed(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeStopFailed,
		Message:    fmt.Sprintf("failed to stop miner '%s'", name),
		Suggestion: "The miner process may need to be terminated manually",
		Retryable:  true,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ErrInvalidConfig creates an invalid configuration error
func ErrInvalidConfig(reason string) *MiningError {
	return &MiningError{
		Code:       ErrCodeInvalidConfig,
		Message:    fmt.Sprintf("invalid configuration: %s", reason),
		Suggestion: "Review the configuration and ensure all required fields are provided",
		Retryable:  false,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrUnsupportedMiner creates an unsupported miner type error
func ErrUnsupportedMiner(minerType string) *MiningError {
	return &MiningError{
		Code:       ErrCodeUnsupportedMiner,
		Message:    fmt.Sprintf("unsupported miner type: %s", minerType),
		Suggestion: "Use one of the supported miner types: xmrig, tt-miner",
		Retryable:  false,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrConnectionFailed creates a connection failed error
func ErrConnectionFailed(target string) *MiningError {
	return &MiningError{
		Code:       ErrCodeConnectionFailed,
		Message:    fmt.Sprintf("failed to connect to %s", target),
		Suggestion: "Check network connectivity and try again",
		Retryable:  true,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// ErrTimeout creates a timeout error
func ErrTimeout(operation string) *MiningError {
	return &MiningError{
		Code:       ErrCodeTimeout,
		Message:    fmt.Sprintf("operation timed out: %s", operation),
		Suggestion: "The operation is taking longer than expected, try again later",
		Retryable:  true,
		HTTPStatus: http.StatusGatewayTimeout,
	}
}

// ErrDatabaseError creates a database error
func ErrDatabaseError(operation string) *MiningError {
	return &MiningError{
		Code:       ErrCodeDatabaseError,
		Message:    fmt.Sprintf("database error during %s", operation),
		Suggestion: "This may be a temporary issue, try again",
		Retryable:  true,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ErrProfileNotFound creates a profile not found error
func ErrProfileNotFound(id string) *MiningError {
	return &MiningError{
		Code:       ErrCodeProfileNotFound,
		Message:    fmt.Sprintf("profile '%s' not found", id),
		Suggestion: "Check that the profile ID is correct",
		Retryable:  false,
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrProfileExists creates a profile already exists error
func ErrProfileExists(name string) *MiningError {
	return &MiningError{
		Code:       ErrCodeProfileExists,
		Message:    fmt.Sprintf("profile '%s' already exists", name),
		Suggestion: "Use a different name or update the existing profile",
		Retryable:  false,
		HTTPStatus: http.StatusConflict,
	}
}

// ErrInternal creates a generic internal error
func ErrInternal(message string) *MiningError {
	return &MiningError{
		Code:       ErrCodeInternalError,
		Message:    message,
		Suggestion: "Please report this issue if it persists",
		Retryable:  true,
		HTTPStatus: http.StatusInternalServerError,
	}
}
