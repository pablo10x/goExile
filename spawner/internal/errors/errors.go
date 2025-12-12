package errors

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeNetwork       ErrorType = "NETWORK"
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypePermission    ErrorType = "PERMISSION"
	ErrorTypeResource      ErrorType = "RESOURCE"
	ErrorTypeConfiguration ErrorType = "CONFIGURATION"
	ErrorTypeTimeout       ErrorType = "TIMEOUT"
	ErrorTypeUnknown       ErrorType = "UNKNOWN"
)

// ErrorSeverity represents the severity level of errors
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// SpawnerError represents a structured error with context
type SpawnerError struct {
	Type     ErrorType              `json:"type"`
	Severity ErrorSeverity          `json:"severity"`
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Context  *ErrorContext          `json:"context,omitempty"`
	Cause    error                  `json:"-"`
	Stack    string                 `json:"stack,omitempty"`
}

// ErrorContext provides additional context about where and when the error occurred
type ErrorContext struct {
	SpawnerID  string    `json:"spawnerId,omitempty"`
	InstanceID string    `json:"instanceId,omitempty"`
	Operation  string    `json:"operation,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	UserID     string    `json:"userId,omitempty"`
}

// ErrorCodes defines standard error codes
var ErrorCodes = struct {
	// Network errors
	NetworkConnectionFailed string
	NetworkTimeout          string
	NetworkUnauthorized     string

	// Validation errors
	ValidationInvalidInput    string
	ValidationMissingRequired string
	ValidationInvalidFormat   string

	// Permission errors
	PermissionDenied       string
	PermissionInsufficient string

	// Resource errors
	ResourceNotFound  string
	ResourceExhausted string
	ResourceConflict  string

	// Configuration errors
	ConfigInvalid string
	ConfigMissing string

	// Timeout errors
	TimeoutExceeded string
}{
	NetworkConnectionFailed: "NETWORK_001",
	NetworkTimeout:          "NETWORK_002",
	NetworkUnauthorized:     "NETWORK_003",

	ValidationInvalidInput:    "VALIDATION_001",
	ValidationMissingRequired: "VALIDATION_002",
	ValidationInvalidFormat:   "VALIDATION_003",

	PermissionDenied:       "PERMISSION_001",
	PermissionInsufficient: "PERMISSION_002",

	ResourceNotFound:  "RESOURCE_001",
	ResourceExhausted: "RESOURCE_002",
	ResourceConflict:  "RESOURCE_003",

	ConfigInvalid: "CONFIG_001",
	ConfigMissing: "CONFIG_002",

	TimeoutExceeded: "TIMEOUT_001",
}

// NewSpawnerError creates a new structured error
func NewSpawnerError(errType ErrorType, severity ErrorSeverity, code, message string, details map[string]interface{}, context *ErrorContext, cause error) *SpawnerError {
	if context == nil {
		context = &ErrorContext{
			Timestamp: time.Now(),
		}
	} else if context.Timestamp.IsZero() {
		context.Timestamp = time.Now()
	}

	return &SpawnerError{
		Type:     errType,
		Severity: severity,
		Code:     code,
		Message:  message,
		Details:  details,
		Context:  context,
		Cause:    cause,
	}
}

// Error implements the error interface
func (e *SpawnerError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// GetUserMessage returns a user-friendly error message
func (e *SpawnerError) GetUserMessage() string {
	switch e.Code {
	case ErrorCodes.NetworkConnectionFailed:
		return "Unable to connect to the spawner. Please check your network connection and try again."
	case ErrorCodes.NetworkTimeout:
		return "The operation timed out. The spawner may be busy or unresponsive."
	case ErrorCodes.PermissionDenied:
		return "You do not have permission to perform this operation."
	case ErrorCodes.ResourceNotFound:
		return "The requested resource was not found."
	case ErrorCodes.ResourceExhausted:
		return "System resources are exhausted. Please try again later."
	case ErrorCodes.ValidationInvalidInput:
		return "The provided input is invalid. Please check your data and try again."
	default:
		return e.Message
	}
}

// GetTechnicalDetails returns detailed error information for logging/debugging
func (e *SpawnerError) GetTechnicalDetails() map[string]interface{} {
	details := map[string]interface{}{
		"type":     e.Type,
		"severity": e.Severity,
		"code":     e.Code,
		"message":  e.Message,
	}

	if e.Details != nil {
		details["details"] = e.Details
	}

	if e.Context != nil {
		details["context"] = e.Context
	}

	if e.Stack != "" {
		details["stack"] = e.Stack
	}

	if e.Cause != nil {
		details["cause"] = e.Cause.Error()
	}

	return details
}

// Log logs the error with appropriate level
func (e *SpawnerError) Log(logger *slog.Logger) {
	details := e.GetTechnicalDetails()

	switch e.Severity {
	case ErrorSeverityCritical:
		logger.Error("CRITICAL ERROR", "error", details)
	case ErrorSeverityHigh:
		logger.Error("HIGH SEVERITY ERROR", "error", details)
	case ErrorSeverityMedium:
		logger.Warn("MEDIUM SEVERITY ERROR", "error", details)
	case ErrorSeverityLow:
		logger.Info("LOW SEVERITY ERROR", "error", details)
	}
}

// ToJSON converts the error to JSON for API responses
func (e *SpawnerError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ToHTTPResponse converts the error to an HTTP response
func (e *SpawnerError) ToHTTPResponse() (int, map[string]interface{}) {
	statusCode := http.StatusInternalServerError

	// Map error types to HTTP status codes
	switch e.Type {
	case ErrorTypeValidation:
		statusCode = http.StatusBadRequest
	case ErrorTypePermission:
		statusCode = http.StatusForbidden
	case ErrorTypeResource:
		if e.Code == ErrorCodes.ResourceNotFound {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusConflict
		}
	case ErrorTypeNetwork:
		statusCode = http.StatusBadGateway
	case ErrorTypeTimeout:
		statusCode = http.StatusGatewayTimeout
	}

	return statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"code":     e.Code,
			"message":  e.GetUserMessage(),
			"type":     e.Type,
			"severity": e.Severity,
		},
	}
}

// Factory functions for common error types

// Network creates a network error
func Network(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypeNetwork, ErrorSeverityHigh, code, message, details, context, nil)
}

// Validation creates a validation error
func Validation(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypeValidation, ErrorSeverityMedium, code, message, details, context, nil)
}

// Permission creates a permission error
func Permission(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypePermission, ErrorSeverityHigh, code, message, details, context, nil)
}

// Resource creates a resource error
func Resource(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypeResource, ErrorSeverityHigh, code, message, details, context, nil)
}

// Config creates a configuration error
func Config(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypeConfiguration, ErrorSeverityMedium, code, message, details, context, nil)
}

// Timeout creates a timeout error
func Timeout(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError {
	return NewSpawnerError(ErrorTypeTimeout, ErrorSeverityMedium, code, message, details, context, nil)
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, context *ErrorContext) *SpawnerError {
	if spawnerErr, ok := err.(*SpawnerError); ok {
		return spawnerErr
	}

	return NewSpawnerError(errType, ErrorSeverityMedium, "WRAPPED_ERROR", err.Error(), nil, context, err)
}

// HandleError is a utility function for handling errors in HTTP handlers
func HandleError(err error, logger *slog.Logger, context *ErrorContext) *SpawnerError {
	var spawnerErr *SpawnerError

	if se, ok := err.(*SpawnerError); ok {
		spawnerErr = se
	} else {
		spawnerErr = Wrap(err, ErrorTypeUnknown, context)
	}

	spawnerErr.Log(logger)
	return spawnerErr
}
