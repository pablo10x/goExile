// Error types for spawner operations
export enum SpawnerErrorType {
	NETWORK = 'NETWORK',
	VALIDATION = 'VALIDATION',
	PERMISSION = 'PERMISSION',
	RESOURCE = 'RESOURCE',
	CONFIGURATION = 'CONFIGURATION',
	TIMEOUT = 'TIMEOUT',
	UNKNOWN = 'UNKNOWN'
}

// Error severity levels
export enum ErrorSeverity {
	LOW = 'LOW',
	MEDIUM = 'MEDIUM',
	HIGH = 'HIGH',
	CRITICAL = 'CRITICAL'
}

// Structured error interface
export interface SpawnerError {
	type: SpawnerErrorType;
	severity: ErrorSeverity;
	code: string;
	message: string;
	details?: Record<string, any>;
	context?: {
		spawnerId?: string;
		instanceId?: string;
		operation?: string;
		timestamp: string;
		userId?: string;
	};
	cause?: Error;
	stack?: string;
}

// Context type for input (all fields optional)
export type SpawnerErrorContextInput = Partial<SpawnerError['context']>;

// Error code definitions
export const ErrorCodes = {
	// Network errors
	NETWORK_CONNECTION_FAILED: 'NETWORK_001',
	NETWORK_TIMEOUT: 'NETWORK_002',
	NETWORK_UNAUTHORIZED: 'NETWORK_003',

	// Validation errors
	VALIDATION_INVALID_INPUT: 'VALIDATION_001',
	VALIDATION_MISSING_REQUIRED: 'VALIDATION_002',
	VALIDATION_INVALID_FORMAT: 'VALIDATION_003',

	// Permission errors
	PERMISSION_DENIED: 'PERMISSION_001',
	PERMISSION_INSUFFICIENT: 'PERMISSION_002',

	// Resource errors
	RESOURCE_NOT_FOUND: 'RESOURCE_001',
	RESOURCE_EXHAUSTED: 'RESOURCE_002',
	RESOURCE_CONFLICT: 'RESOURCE_003',

	// Configuration errors
	CONFIG_INVALID: 'CONFIG_001',
	CONFIG_MISSING: 'CONFIG_002',

	// Timeout errors
	TIMEOUT_EXCEEDED: 'TIMEOUT_001'
} as const;

/**
 * SpawnerErrorWrapper - Comprehensive error handling for spawner operations
 */
export class SpawnerErrorWrapper {
	private error: SpawnerError;

	constructor(error: Partial<SpawnerError>) {
		this.error = {
			type: error.type || SpawnerErrorType.UNKNOWN,
			severity: error.severity || ErrorSeverity.MEDIUM,
			code: error.code || 'UNKNOWN_ERROR',
			message: error.message || 'An unknown error occurred',
			details: error.details,
			context: error.context
				? ({
						...error.context,
						timestamp: (error.context as any).timestamp || new Date().toISOString()
					} as SpawnerError['context'])
				: undefined,
			cause: error.cause,
			stack: error.stack || error.cause?.stack
		};

		// Basic validation
		this.validateError();
	}

	private validateError(): void {
		if (!this.error.code || typeof this.error.code !== 'string') {
			console.warn('Invalid error code:', this.error.code);
		}
		if (!this.error.message || typeof this.error.message !== 'string') {
			console.warn('Invalid error message:', this.error.message);
		}
		if (this.error.context?.timestamp && isNaN(Date.parse(this.error.context.timestamp))) {
			console.warn('Invalid timestamp format:', this.error.context.timestamp);
		}
	}

	/**
	 * Create a network error
	 */
	static network(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.NETWORK,
			severity: ErrorSeverity.HIGH,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Create a validation error
	 */
	static validation(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.VALIDATION,
			severity: ErrorSeverity.MEDIUM,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Create a permission error
	 */
	static permission(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.PERMISSION,
			severity: ErrorSeverity.HIGH,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Create a resource error
	 */
	static resource(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.RESOURCE,
			severity: ErrorSeverity.HIGH,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Create a configuration error
	 */
	static config(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.CONFIGURATION,
			severity: ErrorSeverity.MEDIUM,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Create a timeout error
	 */
	static timeout(
		code: string,
		message: string,
		details?: Record<string, any>,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type: SpawnerErrorType.TIMEOUT,
			severity: ErrorSeverity.MEDIUM,
			code,
			message,
			details,
			context: context as any
		});
	}

	/**
	 * Wrap an existing error
	 */
	static wrap(
		error: Error,
		type: SpawnerErrorType = SpawnerErrorType.UNKNOWN,
		context?: SpawnerErrorContextInput
	): SpawnerErrorWrapper {
		return new SpawnerErrorWrapper({
			type,
			severity: ErrorSeverity.MEDIUM,
			code: 'WRAPPED_ERROR',
			message: error.message,
			context: context as any,
			cause: error,
			stack: error.stack
		});
	}

	/**
	 * Get the structured error
	 */
	getError(): SpawnerError {
		return { ...this.error };
	}

	/**
	 * Get user-friendly error message
	 */
	getUserMessage(): string {
		const { type, code, message } = this.error;

		// Provide user-friendly messages based on error type and code
		switch (code) {
			case ErrorCodes.NETWORK_CONNECTION_FAILED:
				return 'Unable to connect to the spawner. Please check your network connection and try again.';
			case ErrorCodes.NETWORK_TIMEOUT:
				return 'The operation timed out. The spawner may be busy or unresponsive.';
			case ErrorCodes.PERMISSION_DENIED:
				return 'You do not have permission to perform this operation.';
			case ErrorCodes.RESOURCE_NOT_FOUND:
				return 'The requested resource was not found.';
			case ErrorCodes.RESOURCE_EXHAUSTED:
				return 'System resources are exhausted. Please try again later.';
			case ErrorCodes.VALIDATION_INVALID_INPUT:
				return 'The provided input is invalid. Please check your data and try again.';
			default:
				return message;
		}
	}

	/**
	 * Get technical details for logging/debugging
	 */
	getTechnicalDetails(): Record<string, any> {
		return {
			type: this.error.type,
			severity: this.error.severity,
			code: this.error.code,
			message: this.error.message,
			details: this.error.details,
			context: this.error.context,
			stack: this.error.stack,
			cause: this.error.cause?.message
		};
	}

	/**
	 * Log the error with appropriate level
	 */
	log(): void {
		const details = this.getTechnicalDetails();

		switch (this.error.severity) {
			case ErrorSeverity.CRITICAL:
				console.error('CRITICAL ERROR:', details);
				break;
			case ErrorSeverity.HIGH:
				console.error('HIGH SEVERITY ERROR:', details);
				break;
			case ErrorSeverity.MEDIUM:
				console.warn('MEDIUM SEVERITY ERROR:', details);
				break;
			case ErrorSeverity.LOW:
				console.info('LOW SEVERITY ERROR:', details);
				break;
		}
	}

	/**
	 * Convert to JSON for API responses
	 */
	toJSON(): SpawnerError {
		return this.getError();
	}

	/**
	 * Convert to string representation
	 */
	toString(): string {
		return `[${this.error.type}:${this.error.code}] ${this.error.message}`;
	}
}

/**
 * Error handler utility for async operations
 */
export async function withErrorHandler<T>(
	operation: () => Promise<T>,
	context?: Partial<SpawnerError['context']>
): Promise<T> {
	try {
		return await operation();
	} catch (error) {
		if (error instanceof SpawnerErrorWrapper) {
			throw error;
		}

		// Wrap unknown errors
		const wrappedError = SpawnerErrorWrapper.wrap(
			error instanceof Error ? error : new Error(String(error)),
			SpawnerErrorType.UNKNOWN,
			context
		);

		wrappedError.log();
		throw wrappedError;
	}
}

/**
 * Error boundary helper for UI components
 */
export function handleUIError(
	error: unknown,
	context?: Partial<SpawnerError['context']>
): SpawnerErrorWrapper {
	if (error instanceof SpawnerErrorWrapper) {
		return error;
	}

	const wrappedError = SpawnerErrorWrapper.wrap(
		error instanceof Error ? error : new Error(String(error)),
		SpawnerErrorType.UNKNOWN,
		context
	);

	wrappedError.log();
	return wrappedError;
}
