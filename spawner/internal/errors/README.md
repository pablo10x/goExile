# Spawner Error Package (Go)

A comprehensive error handling package for the Go spawner service that provides structured error reporting, user-friendly messages, and detailed logging.

## Features

- **Structured Error Types**: Categorized errors (Network, Validation, Permission, Resource, etc.)
- **Severity Levels**: LOW, MEDIUM, HIGH, CRITICAL
- **Context Information**: Spawner ID, instance ID, operation details, timestamps
- **User-Friendly Messages**: Automatic translation of technical errors to user-readable messages
- **Technical Details**: Full error context for debugging and logging
- **Error Wrapping**: Wrap existing errors with additional context
- **HTTP Integration**: Easy conversion to HTTP responses with appropriate status codes
- **JSON Serialization**: Structured error output for APIs

## Installation

This package is included in the spawner internal packages. Import from:

```go
import "spawner/internal/errors"
```

## Basic Usage

### Creating Specific Error Types

```go
// Network error
err := errors.Network(
    errors.ErrorCodes.NetworkConnectionFailed,
    "Failed to connect to spawner",
    map[string]interface{}{"spawnerId": "spawner-123"},
    &errors.ErrorContext{
        SpawnerID: "spawner-123",
        Operation: "connect",
    },
)

// Validation error
validationErr := errors.Validation(
    errors.ErrorCodes.ValidationMissingRequired,
    "Instance name is required",
    map[string]interface{}{"field": "name"},
    &errors.ErrorContext{
        Operation: "spawn_instance",
    },
)

// Permission error
permErr := errors.Permission(
    errors.ErrorCodes.PermissionDenied,
    "Access denied",
    nil,
    &errors.ErrorContext{
        UserID: "user-123",
        Operation: "admin_action",
    },
)
```

### Wrapping Existing Errors

```go
func connectToSpawner(spawnerId string) error {
    resp, err := http.Get(fmt.Sprintf("http://%s/api/health", spawnerId))
    if err != nil {
        // Wrap the network error
        return errors.Wrap(err, errors.ErrorTypeNetwork, &errors.ErrorContext{
            SpawnerID: spawnerId,
            Operation: "health_check",
        })
    }
    return nil
}
```

### HTTP Handler Integration

```go
func (h *Handler) handleError(c *gin.Context, err error, context *errors.ErrorContext) {
    spawnerErr := errors.HandleError(err, h.logger, context)
    statusCode, response := spawnerErr.ToHTTPResponse()
    c.JSON(statusCode, response)
}

func (h *Handler) HandleSpawn(c *gin.Context) {
    // ... validation logic ...
    if !isValidConfig(config) {
        h.handleError(c, errors.Validation(
            errors.ErrorCodes.ValidationInvalidInput,
            "Invalid spawn configuration",
            map[string]interface{}{"config": config},
            &errors.ErrorContext{
                Operation: "spawn",
                SpawnerID: h.config.ID,
            },
        ), &errors.ErrorContext{Operation: "spawn"})
        return
    }

    // ... spawn logic ...
    instance, err := h.manager.SpawnInstance(config)
    if err != nil {
        h.handleError(c, errors.Wrap(err, errors.ErrorTypeResource, &errors.ErrorContext{
            Operation: "spawn",
            SpawnerID: h.config.ID,
        }), &errors.ErrorContext{Operation: "spawn"})
        return
    }

    c.JSON(http.StatusCreated, instance)
}
```

## Error Types

- `ErrorTypeNetwork`: Connection, timeout, and HTTP errors
- `ErrorTypeValidation`: Input validation and data format errors
- `ErrorTypePermission`: Authorization and access control errors
- `ErrorTypeResource`: Resource not found, exhausted, or conflict errors
- `ErrorTypeConfiguration`: Configuration and setup errors
- `ErrorTypeTimeout`: Operation timeout errors
- `ErrorTypeUnknown`: Unclassified errors

## Severity Levels

- `ErrorSeverityLow`: Minor issues that don't affect functionality
- `ErrorSeverityMedium`: Issues that may cause degraded performance
- `ErrorSeverityHigh`: Serious issues that prevent some operations
- `ErrorSeverityCritical`: System-breaking errors requiring immediate attention

## Error Codes

See `ErrorCodes` struct for predefined error codes:

- `NetworkConnectionFailed`: "NETWORK_001"
- `ValidationInvalidInput`: "VALIDATION_001"
- `PermissionDenied`: "PERMISSION_001"
- `ResourceNotFound`: "RESOURCE_001"
- `ConfigMissing`: "CONFIG_002"
- etc.

## API Reference

### SpawnerError Struct

```go
type SpawnerError struct {
    Type       ErrorType
    Severity   ErrorSeverity
    Code       string
    Message    string
    Details    map[string]interface{}
    Context    *ErrorContext
    Cause      error
    Stack      string
}
```

### Methods

- `Error() string`: Implements error interface
- `GetUserMessage() string`: Returns user-friendly message
- `GetTechnicalDetails() map[string]interface{}`: Returns full technical details
- `Log(logger *slog.Logger)`: Logs error with appropriate level
- `ToJSON() ([]byte, error)`: Converts to JSON
- `ToHTTPResponse() (int, map[string]interface{})`: Converts to HTTP response

### Factory Functions

- `Network(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Validation(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Permission(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Resource(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Config(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Timeout(code, message string, details map[string]interface{}, context *ErrorContext) *SpawnerError`
- `Wrap(err error, errType ErrorType, context *ErrorContext) *SpawnerError`

### Utility Functions

- `HandleError(err error, logger *slog.Logger, context *ErrorContext) *SpawnerError`: Processes errors for logging

## HTTP Status Code Mapping

The package automatically maps error types to appropriate HTTP status codes:

- `Validation` → 400 Bad Request
- `Permission` → 403 Forbidden
- `Resource` (not found) → 404 Not Found
- `Resource` (other) → 409 Conflict
- `Network` → 502 Bad Gateway
- `Timeout` → 504 Gateway Timeout
- Others → 500 Internal Server Error

## Best Practices

1. **Use Specific Error Types**: Choose the most appropriate error type for better categorization
2. **Provide Context**: Include spawnerId, instanceId, and operation details
3. **Use Error Codes**: Use predefined error codes for consistency
4. **Wrap External Errors**: Always wrap errors from external libraries/APIs
5. **Log Appropriately**: Use the built-in logging methods for consistent logging
6. **User-Friendly Messages**: Use `GetUserMessage()` for API responses
7. **Technical Details**: Use `GetTechnicalDetails()` for internal logging

## Integration with Frontend

The Go error package works seamlessly with the TypeScript error wrapper:

```go
// Backend returns structured error
spawnerErr := errors.Resource(
    errors.ErrorCodes.ResourceNotFound,
    "Instance not found",
    map[string]interface{}{"instanceId": instanceId},
    &errors.ErrorContext{
        InstanceID: instanceId,
        Operation: "get_instance",
    },
)

statusCode, response := spawnerErr.ToHTTPResponse()
// Returns: 404, {"error": {"code": "RESOURCE_001", "message": "The requested resource was not found", ...}}
```

```typescript
// Frontend receives and handles the error
if (response.status === 404) {
  const error = response.data.error;
  showNotification(error.message, { type: "error" });
}
```

## Examples

See `examples.go` for comprehensive usage examples covering all error types and integration patterns.
