import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { tick } from 'svelte'; // Import tick from svelte
import LogViewer from './LogViewer.svelte';

global.fetch = vi.fn() as any;

describe('LogViewer Component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

    it('displays loading state initially', async () => {
        (global.fetch as any).mockReturnValue(new Promise(() => {})); // Never resolve
        render(LogViewer, { spawnerId: 1 });
        await tick(); // Force Svelte to flush updates
        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('displays logs when fetch is successful', async () => {
        (global.fetch as any).mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ logs: 'Server started successfully.' }),
        });

        render(LogViewer, { spawnerId: 1 });
        await tick(); // Force Svelte to flush updates

        await waitFor(() => {
            expect(screen.getByText('Server started successfully.')).toBeInTheDocument();
        });
    });

    it('displays error message on fetch failure', async () => {
        (global.fetch as any).mockResolvedValue({
            ok: false,
            statusText: 'Internal Server Error',
        });

        render(LogViewer, { spawnerId: 1 });
        await tick(); // Force Svelte to flush updates

        await waitFor(() => {
            expect(screen.getByText('Failed to fetch logs: Internal Server Error')).toBeInTheDocument();
        });
    });
});
