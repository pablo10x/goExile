# Spawner Error Wrapper Package

A comprehensive error handling package for spawner operations that provides structured error reporting, user-friendly messages, and detailed logging.

## Features

- **Structured Error Types**: Categorized errors (Network, Validation, Permission, Resource, etc.)
- **Severity Levels**: LOW, MEDIUM, HIGH, CRITICAL
- **Context Information**: Spawner ID, instance ID, operation details, timestamps
- **User-Friendly Messages**: Automatic translation of technical errors to user-readable messages
- **Technical Details**: Full error context for debugging and logging
- **Error Wrapping**: Wrap existing errors with additional context
- **Async Error Handling**: Utilities for handling errors in async operations
- **UI Integration**: Easy integration with UI components for error display

## Installation

This package is included in the spawner utilities. Import from:

```typescript
import { SpawnerErrorWrapper, SpawnerErrorType, ErrorCodes } from '$lib/utils/spawner-errors';
```

## Basic Usage

### Creating Specific Error Types

```typescript
// Network error
const error = SpawnerErrorWrapper.network(
	ErrorCodes.NETWORK_CONNECTION_FAILED,
	'Failed to connect to spawner',
	{ spawnerId: 'spawner-123' },
	{ spawnerId: 'spawner-123', operation: 'connect' }
);

// Validation error
const validationError = SpawnerErrorWrapper.validation(
	ErrorCodes.VALIDATION_MISSING_REQUIRED,
	'Instance name is required',
	{ field: 'name' }
);

// Permission error
const permError = SpawnerErrorWrapper.permission(ErrorCodes.PERMISSION_DENIED, 'Access denied');
```

### Wrapping Existing Errors

```typescript
try {
	await someSpawnerOperation();
} catch (error) {
	const wrappedError = SpawnerErrorWrapper.wrap(error as Error, SpawnerErrorType.NETWORK, {
		spawnerId: 'spawner-123',
		operation: 'spawn'
	});

	// Log with appropriate severity
	wrappedError.log();

	// Get user-friendly message
	console.log(wrappedError.getUserMessage());

	throw wrappedError;
}
```

### Async Error Handling

```typescript
import { withErrorHandler } from '$lib/utils/spawner-errors';

async function spawnInstance(spawnerId: string, config: any) {
	return withErrorHandler(
		async () => {
			const response = await fetch(`/api/spawners/${spawnerId}/spawn`, {
				method: 'POST',
				body: JSON.stringify(config)
			});

			if (!response.ok) {
				throw SpawnerErrorWrapper.network(
					ErrorCodes.NETWORK_CONNECTION_FAILED,
					'Spawn request failed'
				);
			}

			return response.json();
		},
		{ spawnerId, operation: 'spawn' }
	);
}
```

### UI Error Handling

```typescript
import { handleUIError } from '$lib/utils/spawner-errors';

function handleError(error: unknown) {
	const wrappedError = handleUIError(error, { component: 'SpawnerList' });

	// Show user-friendly message in UI
	showNotification(wrappedError.getUserMessage(), {
		type: wrappedError.error.severity === 'CRITICAL' ? 'error' : 'warning'
	});

	// Log technical details
	wrappedError.log();
}
```

## Error Types

- `NETWORK`: Connection, timeout, and HTTP errors
- `VALIDATION`: Input validation and data format errors
- `PERMISSION`: Authorization and access control errors
- `RESOURCE`: Resource not found, exhausted, or conflict errors
- `CONFIGURATION`: Configuration and setup errors
- `TIMEOUT`: Operation timeout errors
- `UNKNOWN`: Unclassified errors

## Severity Levels

- `LOW`: Minor issues that don't affect functionality
- `MEDIUM`: Issues that may cause degraded performance
- `HIGH`: Serious issues that prevent some operations
- `CRITICAL`: System-breaking errors requiring immediate attention

## Error Codes

See `ErrorCodes` object for predefined error codes:

- `NETWORK_001`: Connection failed
- `VALIDATION_001`: Invalid input
- `PERMISSION_001`: Access denied
- `RESOURCE_001`: Not found
- `CONFIG_002`: Configuration missing
- etc.

## API Reference

### SpawnerErrorWrapper Class

#### Constructor

```typescript
new SpawnerErrorWrapper(error: Partial<SpawnerError>)
```

#### Static Factory Methods

- `SpawnerErrorWrapper.network(code, message, details?, context?)`
- `SpawnerErrorWrapper.validation(code, message, details?, context?)`
- `SpawnerErrorWrapper.permission(code, message, details?, context?)`
- `SpawnerErrorWrapper.resource(code, message, details?, context?)`
- `SpawnerErrorWrapper.config(code, message, details?, context?)`
- `SpawnerErrorWrapper.timeout(code, message, details?, context?)`
- `SpawnerErrorWrapper.wrap(error, type?, context?)`

#### Instance Methods

- `getError(): SpawnerError` - Get structured error object
- `getUserMessage(): string` - Get user-friendly message
- `getTechnicalDetails(): object` - Get full technical details
- `log(): void` - Log error with appropriate level
- `toJSON(): SpawnerError` - Convert to JSON
- `toString(): string` - Convert to string representation

### Utility Functions

#### withErrorHandler

```typescript
withErrorHandler<T>(operation: () => Promise<T>, context?: Partial<SpawnerError['context']>): Promise<T>
```

Wraps async operations with automatic error handling and logging.

#### handleUIError

```typescript
handleUIError(error: unknown, context?: Partial<SpawnerError['context']>): SpawnerErrorWrapper
```

Handles errors for UI components, providing user-friendly messages and logging.

## Best Practices

1. **Use Specific Error Types**: Choose the most appropriate error type for better categorization
2. **Provide Context**: Include spawnerId, instanceId, and operation details
3. **Use Error Codes**: Use predefined error codes for consistency
4. **Wrap External Errors**: Always wrap errors from external libraries/APIs
5. **Log Appropriately**: Use the built-in logging methods for consistent logging
6. **User-Friendly Messages**: Use `getUserMessage()` for UI display
7. **Technical Details**: Use `getTechnicalDetails()` for debugging and logging

## Integration with LogViewer

The error wrapper integrates seamlessly with the LogViewer component:

```typescript
// Errors are automatically logged with full context
const error = SpawnerErrorWrapper.resource(
	ErrorCodes.RESOURCE_NOT_FOUND,
	'Spawner not found',
	{ spawnerId },
	{ spawnerId, operation: 'get_status' }
);

error.log(); // Logs with ERROR level and full context
```

## Examples

See `spawner-errors-examples.ts` for comprehensive usage examples.
