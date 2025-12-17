import { render, screen, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte'; // Import tick from svelte
import { describe, it, expect } from 'vitest';
import Page from './+page.svelte';

describe('Login Page', () => {
    it('renders login form', async () => {
        render(Page);
        await tick(); // Force Svelte to flush updates
        await Promise.resolve(); // Allow Svelte's reactivity to settle
        await waitFor(() => {
            const heading = screen.getByText('GoExile Admin');
            expect(heading).toBeInTheDocument();

            const emailInput = screen.getByLabelText('Email');
            expect(emailInput).toBeInTheDocument();

            const passwordInput = screen.getByLabelText('Password');
            expect(passwordInput).toBeInTheDocument();

            const loginButton = screen.getByRole('button', { name: /login/i });
            expect(loginButton).toBeInTheDocument();
        });
    });
});
