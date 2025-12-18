import '@testing-library/jest-dom';
import { vi } from 'vitest';

vi.mock('$app/navigation', () => ({
	goto: vi.fn(),
	invalidate: vi.fn(),
	invalidateAll: vi.fn(),
	preloadData: vi.fn(),
	preloadCode: vi.fn(),
	beforeNavigate: vi.fn(),
	afterNavigate: vi.fn()
}));

vi.mock('$app/stores', () => ({
	getStores: vi.fn(),
	navigating: { subscribe: vi.fn() },
	page: { subscribe: vi.fn() },
	updated: { subscribe: vi.fn() }
}));

// Mock for Svelte 5 $app/state if used
vi.mock('$app/state', () => ({
	page: {
		url: new URL('http://localhost'),
		params: {},
		route: { id: null },
		status: 200,
		error: null,
		data: {},
		form: null
	}
}));

// Polyfill for element.animate in jsdom environment
if (typeof Element !== 'undefined' && !Element.prototype.animate) {
	Element.prototype.animate = () => ({
		// Mock the return value to prevent further errors if methods are called on it
		cancel: () => {},
		pause: () => {},
		play: () => {},
		finish: () => {},
		reverse: () => {},
		commitStyles: () => {},
		add: () => {},
		// Add other properties that might be accessed, e.g., onfinish
		onfinish: null,
		// Mock a Promise-like object for `finished`
		finished: Promise.resolve(this as any)
	}) as any;
}
