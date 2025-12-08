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