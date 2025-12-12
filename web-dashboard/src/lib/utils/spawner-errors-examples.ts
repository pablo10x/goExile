// Example usage of SpawnerErrorWrapper

import { SpawnerErrorWrapper, SpawnerErrorType, ErrorCodes, withErrorHandler, handleUIError } from '$lib/utils/spawner-errors';

// Example 1: Creating specific error types
try {
  // Some spawner operation that fails
  throw new Error('Connection timeout');
} catch (error) {
  const wrappedError = SpawnerErrorWrapper.network(
    ErrorCodes.NETWORK_TIMEOUT,
    'Failed to connect to spawner',
    { spawnerId: 'spawner-123', retryCount: 3 },
    { spawnerId: 'spawner-123', operation: 'connect' }
  );

  // Log the error
  wrappedError.log();

  // Get user-friendly message
  console.log(wrappedError.getUserMessage()); // "The operation timed out. The spawner may be busy or unresponsive."

  // Get technical details for debugging
  console.log(wrappedError.getTechnicalDetails());
}

// Example 2: Wrapping existing errors
async function connectToSpawner(spawnerId: string) {
  try {
    const response = await fetch(`/api/spawners/${spawnerId}/connect`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
  } catch (error) {
    // Wrap the network error
    const wrappedError = SpawnerErrorWrapper.wrap(
      error as Error,
      SpawnerErrorType.NETWORK,
      { spawnerId, operation: 'connect' }
    );

    wrappedError.log();
    throw wrappedError;
  }
}

// Example 3: Using error handler for async operations
async function spawnInstance(spawnerId: string, config: any) {
  return withErrorHandler(async () => {
    // Validate input
    if (!config.name) {
      throw SpawnerErrorWrapper.validation(
        ErrorCodes.VALIDATION_MISSING_REQUIRED,
        'Instance name is required',
        { field: 'name' },
        { spawnerId, operation: 'spawn' }
      );
    }

    const response = await fetch(`/api/spawners/${spawnerId}/spawn`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    });

    if (!response.ok) {
      if (response.status === 403) {
        throw SpawnerErrorWrapper.permission(
          ErrorCodes.PERMISSION_DENIED,
          'Insufficient permissions to spawn instances',
          {},
          { spawnerId, operation: 'spawn' }
        );
      }
      throw SpawnerErrorWrapper.network(
        ErrorCodes.NETWORK_CONNECTION_FAILED,
        'Failed to spawn instance',
        { status: response.status },
        { spawnerId, operation: 'spawn' }
      );
    }

    return response.json();
  }, { spawnerId, operation: 'spawn' });
}

// Example 4: UI error handling
function handleSpawnerError(error: unknown, context?: any) {
  const wrappedError = handleUIError(error, context);
  const errorData = wrappedError.getError();

  // Show user-friendly message (using your UI framework)
  // showToast(wrappedError.getUserMessage(), {
  //   type: errorData.severity === 'CRITICAL' ? 'error' : 'warning'
  // });

  // Log technical details
  wrappedError.log();

  return wrappedError;
}

// Example 5: Error response for API
function createErrorResponse(error: SpawnerErrorWrapper) {
  const errorData = error.getError();
  return {
    success: false,
    error: {
      code: errorData.code,
      message: error.getUserMessage(),
      type: errorData.type,
      severity: errorData.severity,
      ...(process.env.NODE_ENV === 'development' && {
        details: error.getTechnicalDetails()
      })
    }
  };
}