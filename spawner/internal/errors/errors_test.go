package errors

import (
	"testing"
)

func TestSpawnerError(t *testing.T) {
	// Test basic error creation
	err := New(ErrCodeProcessStart, "Failed to start process")
	if err.Code != ErrCodeProcessStart {
		t.Errorf("Expected code %s, got %s", ErrCodeProcessStart, err.Code)
	}
	if err.Message != "Failed to start process" {
		t.Errorf("Expected message 'Failed to start process', got %s", err.Message)
	}

	// Test error wrapping
	originalErr := New(ErrCodeFileOperation, "File not found")
	wrappedErr := Wrap(originalErr, ErrCodeProcessStart, "start_instance")
	if wrappedErr.Cause != originalErr {
		t.Error("Cause not preserved in wrapped error")
	}
	if wrappedErr.Operation != "start_instance" {
		t.Errorf("Expected operation 'start_instance', got %s", wrappedErr.Operation)
	}

	// Test context addition
	err.WithContext("instance_id", "test-123").
		WithContext("port", 8080)
	if err.Context["instance_id"] != "test-123" {
		t.Error("Context not properly added")
	}
	if err.Context["port"] != 8080 {
		t.Error("Context not properly added")
	}

	// Test error code checking
	if !IsErrorCode(err, ErrCodeProcessStart) {
		t.Error("Error code checking failed")
	}
	if IsErrorCode(err, ErrCodeFileOperation) {
		t.Error("Error code checking should have failed")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test ProcessStartError
	err := ProcessStartError("spawn_instance", &testError{"test error"})
	if !IsErrorCode(err, ErrCodeProcessStart) {
		t.Error("ProcessStartError should have correct error code")
	}
	if err.Context["error_type"] != "process_start" {
		t.Error("ProcessStartError should have correct error type in context")
	}

	// Test PortAllocationError
	portErr := PortAllocationError(8000, 9000, &testError{"no ports available"})
	if !IsErrorCode(portErr, ErrCodePortAllocation) {
		t.Error("PortAllocationError should have correct error code")
	}
	if portErr.Context["min_port"] != 8000 {
		t.Error("PortAllocationError should have correct min_port in context")
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
