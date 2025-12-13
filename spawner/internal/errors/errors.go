package errors

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// Error codes for categorization
const (
	ErrCodeProcessStart   = "PROCESS_START_FAILED"
	ErrCodeProcessStop    = "PROCESS_STOP_FAILED"
	ErrCodePortAllocation = "PORT_ALLOCATION_FAILED"
	ErrCodeFileOperation  = "FILE_OPERATION_FAILED"
	ErrCodeNetwork        = "NETWORK_ERROR"
	ErrCodeConfig         = "CONFIG_ERROR"
	ErrCodeTimeout        = "TIMEOUT_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodePermission     = "PERMISSION_ERROR"
	ErrCodeResource       = "RESOURCE_ERROR"
)

// SpawnerError represents a structured error with context
type SpawnerError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Operation string         `json:"operation"`
	Context   map[string]any `json:"context,omitempty"`
	Stack     []string       `json:"stack,omitempty"`
	Cause     error          `json:"-"`
}

// Error implements the error interface
func (e *SpawnerError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Code, e.Message))
	if e.Operation != "" {
		sb.WriteString(fmt.Sprintf(" (operation: %s)", e.Operation))
	}
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(": %v", e.Cause))
	}
	return sb.String()
}

// Unwrap returns the underlying cause
func (e *SpawnerError) Unwrap() error {
	return e.Cause
}

// New creates a new SpawnerError with the given code and message
func New(code, message string) *SpawnerError {
	return &SpawnerError{
		Code:    code,
		Message: message,
		Stack:   captureStack(),
	}
}

// Newf creates a new SpawnerError with formatted message
func Newf(code, format string, args ...any) *SpawnerError {
	return &SpawnerError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Stack:   captureStack(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, operation string) *SpawnerError {
	if err == nil {
		return nil
	}

	se := &SpawnerError{
		Code:      code,
		Message:   err.Error(),
		Operation: operation,
		Stack:     captureStack(),
		Cause:     err,
	}

	// If the underlying error is already a SpawnerError, preserve its context
	if underlying, ok := err.(*SpawnerError); ok {
		se.Context = underlying.Context
		if se.Code == "" {
			se.Code = underlying.Code
		}
	}

	return se
}

// Wrapf wraps an existing error with formatted operation context
func Wrapf(err error, code, format string, args ...any) *SpawnerError {
	if err == nil {
		return nil
	}

	return &SpawnerError{
		Code:      code,
		Message:   err.Error(),
		Operation: fmt.Sprintf(format, args...),
		Stack:     captureStack(),
		Cause:     err,
	}
}

// WithContext adds context to the error
func (e *SpawnerError) WithContext(key string, value any) *SpawnerError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context values
func (e *SpawnerError) WithContextMap(ctx map[string]any) *SpawnerError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	for k, v := range ctx {
		e.Context[k] = v
	}
	return e
}

// LogAttrs returns slog attributes for structured logging
func (e *SpawnerError) LogAttrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("error_code", e.Code),
		slog.String("error_message", e.Message),
	}

	if e.Operation != "" {
		attrs = append(attrs, slog.String("operation", e.Operation))
	}

	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}

	// Add context attributes
	for k, v := range e.Context {
		attrs = append(attrs, slog.Any(k, v))
	}

	// Add stack trace if available
	if len(e.Stack) > 0 {
		attrs = append(attrs, slog.String("stack", strings.Join(e.Stack, "\n")))
	}

	return attrs
}

// captureStack captures the current goroutine's stack trace
func captureStack() []string {
	const maxFrames = 10
	var stack []string

	for i := 2; i < maxFrames+2; i++ { // Skip captureStack and the calling function
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// Create a readable stack frame
		frame := fmt.Sprintf("%s:%d %s",
			file,
			line,
			fn.Name())
		stack = append(stack, frame)
	}

	return stack
}

// IsErrorCode checks if an error matches a specific error code
func IsErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}

	if se, ok := err.(*SpawnerError); ok {
		return se.Code == code
	}

	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if err == nil {
		return ""
	}

	if se, ok := err.(*SpawnerError); ok {
		return se.Code
	}

	return "UNKNOWN"
}

// Helper functions for common error scenarios

// ProcessStartError creates a process start failure error
func ProcessStartError(operation string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeProcessStart, operation).
		WithContext("error_type", "process_start")
}

// ProcessStopError creates a process stop failure error
func ProcessStopError(operation string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeProcessStop, operation).
		WithContext("error_type", "process_stop")
}

// PortAllocationError creates a port allocation failure error
func PortAllocationError(minPort, maxPort int, cause error) *SpawnerError {
	return Wrap(cause, ErrCodePortAllocation, "find_available_port").
		WithContext("error_type", "port_allocation").
		WithContext("min_port", minPort).
		WithContext("max_port", maxPort)
}

// FileOperationError creates a file operation failure error
func FileOperationError(operation, path string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeFileOperation, operation).
		WithContext("error_type", "file_operation").
		WithContext("file_path", path)
}

// NetworkError creates a network-related error
func NetworkError(operation string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeNetwork, operation).
		WithContext("error_type", "network")
}

// ConfigError creates a configuration error
func ConfigError(field, value string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeConfig, "config_validation").
		WithContext("error_type", "config").
		WithContext("config_field", field).
		WithContext("config_value", value)
}

// TimeoutError creates a timeout error
func TimeoutError(operation string, timeout string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeTimeout, operation).
		WithContext("error_type", "timeout").
		WithContext("timeout_duration", timeout)
}

// ValidationError creates a validation error
func ValidationError(field, value string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeValidation, "validation").
		WithContext("error_type", "validation").
		WithContext("field", field).
		WithContext("value", value)
}

// PermissionError creates a permission error
func PermissionError(operation, resource string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodePermission, operation).
		WithContext("error_type", "permission").
		WithContext("resource", resource)
}

// ResourceError creates a resource-related error
func ResourceError(resourceType, resourceID string, cause error) *SpawnerError {
	return Wrap(cause, ErrCodeResource, "resource_operation").
		WithContext("error_type", "resource").
		WithContext("resource_type", resourceType).
		WithContext("resource_id", resourceID)
}
