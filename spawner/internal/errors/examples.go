// Example usage of the Go SpawnerError package

package main

import (
	"log/slog"
	"net/http"
	"spawner/internal/errors"

	"github.com/gin-gonic/gin"
)

func exampleHandler(c *gin.Context, logger *slog.Logger) {
	// Example 1: Creating specific errors
	id := c.Param("id")
	if id == "" {
		err := errors.Validation(
			errors.ErrorCodes.ValidationMissingRequired,
			"Instance ID is required",
			map[string]interface{}{"field": "id"},
			&errors.ErrorContext{
				Operation: "get_instance",
				SpawnerID: "spawner-123",
			},
		)
		handleError(c, err, logger)
		return
	}

	// Example 2: Wrapping existing errors
	instance, err := getInstance(id)
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeResource, &errors.ErrorContext{
			InstanceID: id,
			Operation:  "get_instance",
			SpawnerID:  "spawner-123",
		})
		handleError(c, wrappedErr, logger)
		return
	}

	c.JSON(http.StatusOK, instance)
}

func handleError(c *gin.Context, err *errors.SpawnerError, logger *slog.Logger) {
	statusCode, response := err.ToHTTPResponse()
	c.JSON(statusCode, response)
}

// Example 3: Resource error
func handleResourceError(c *gin.Context, logger *slog.Logger, resourceID string) {
	err := errors.Resource(
		errors.ErrorCodes.ResourceNotFound,
		"Instance not found",
		map[string]interface{}{
			"resourceId":   resourceID,
			"resourceType": "instance",
		},
		&errors.ErrorContext{
			InstanceID: resourceID,
			Operation:  "find_instance",
			SpawnerID:  "spawner-123",
		},
	)
	handleError(c, err, logger)
}

// Example 4: Network error
func handleNetworkError(c *gin.Context, logger *slog.Logger, target string) {
	err := errors.Network(
		errors.ErrorCodes.NetworkConnectionFailed,
		"Failed to connect to target",
		map[string]interface{}{
			"target":  target,
			"timeout": "30s",
		},
		&errors.ErrorContext{
			Operation: "network_connect",
			SpawnerID: "spawner-123",
		},
	)
	handleError(c, err, logger)
}

// Example 5: Permission error
func handlePermissionError(c *gin.Context, logger *slog.Logger, userID string) {
	err := errors.Permission(
		errors.ErrorCodes.PermissionDenied,
		"Insufficient permissions",
		map[string]interface{}{
			"requiredRole": "admin",
			"userRole":     "user",
		},
		&errors.ErrorContext{
			Operation: "admin_operation",
			SpawnerID: "spawner-123",
			UserID:    userID,
		},
	)
	handleError(c, err, logger)
}

// Example 6: Configuration error
func handleConfigError(c *gin.Context, logger *slog.Logger, configKey string) {
	err := errors.Config(
		errors.ErrorCodes.ConfigMissing,
		"Required configuration missing",
		map[string]interface{}{
			"configKey":  configKey,
			"configFile": "spawner.yaml",
		},
		&errors.ErrorContext{
			Operation: "load_config",
			SpawnerID: "spawner-123",
		},
	)
	handleError(c, err, logger)
}

// Example 7: Timeout error
func handleTimeoutError(c *gin.Context, logger *slog.Logger, operation string) {
	err := errors.Timeout(
		errors.ErrorCodes.TimeoutExceeded,
		"Operation timed out",
		map[string]interface{}{
			"timeout":   "30s",
			"operation": operation,
		},
		&errors.ErrorContext{
			Operation: operation,
			SpawnerID: "spawner-123",
		},
	)
	handleError(c, err, logger)
}

// Mock functions for examples
func getInstance(id string) (interface{}, error) {
	return nil, errors.New("instance not found")
}
