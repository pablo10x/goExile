// Example usage of NodeErrorWrapper

import {
	NodeErrorWrapper,
	NodeErrorType,
	ErrorCodes,
	withErrorHandler,
	handleUIError
} from '$lib/utils/node-errors';

// Example 1: Creating specific error types
try {
	// Some node operation that fails
	throw new Error('Connection timeout');
} catch (error) {
	const wrappedError = NodeErrorWrapper.network(
		ErrorCodes.NETWORK_TIMEOUT,
		'Failed to connect to node',
		{ nodeId: 'node-123', retryCount: 3 },
		{ nodeId: 'node-123', operation: 'connect' }
	);

	// Log the error
	wrappedError.log();

	// Get user-friendly message
	console.log(wrappedError.getUserMessage()); // "The operation timed out. The node may be busy or unresponsive."

	// Get technical details for debugging
	console.log(wrappedError.getTechnicalDetails());
}

// Example 2: Wrapping existing errors
async function connectToNode(nodeId: string) {
	try {
		const response = await fetch(`/api/nodes/${nodeId}/connect`);
		if (!response.ok) {
			throw new Error(`HTTP ${response.status}`);
		}
		return response.json();
	} catch (error) {
		// Wrap the network error
		const wrappedError = NodeErrorWrapper.wrap(error as Error, NodeErrorType.NETWORK, {
			nodeId,
			operation: 'connect'
		});

		wrappedError.log();
		throw wrappedError;
	}
}

// Example 3: Using error handler for async operations
async function spawnInstance(nodeId: string, config: any) {
	return withErrorHandler(
		async () => {
			// Validate input
			if (!config.name) {
				throw NodeErrorWrapper.validation(
					ErrorCodes.VALIDATION_MISSING_REQUIRED,
					'Instance name is required',
					{ field: 'name' },
					{ nodeId, operation: 'spawn' }
				);
			}

			const response = await fetch(`/api/nodes/${nodeId}/spawn`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(config)
			});

			if (!response.ok) {
				if (response.status === 403) {
					throw NodeErrorWrapper.permission(
						ErrorCodes.PERMISSION_DENIED,
						'Insufficient permissions to spawn instances',
						{},
						{ nodeId, operation: 'spawn' }
					);
				}
				throw NodeErrorWrapper.network(
					ErrorCodes.NETWORK_CONNECTION_FAILED,
					'Failed to spawn instance',
					{ status: response.status },
					{ nodeId, operation: 'spawn' }
				);
			}

			return response.json();
		},
		{ nodeId, operation: 'spawn' }
	);
}

// Example 4: UI error handling
function handleNodeError(error: unknown, context?: any) {
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
function createErrorResponse(error: NodeErrorWrapper) {
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
