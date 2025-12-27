import { render, screen, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import { describe, it, expect } from 'vitest';
import Page from './+page.svelte';

describe('Login Page', () => {
	it('renders the modern login form correctly', async () => {
		render(Page);
		// Wait for the component to mount and animations to start
		await tick();
		await new Promise(r => setTimeout(r, 100)); // Allow time for {#if mounted}
		
		await waitFor(() => {
			// Check for the main heading
			const heading = screen.getByRole('heading', { name: /Asset Registry/i, level: 1 });
			expect(heading).toBeInTheDocument();

			// Check for input fields by their labels
			const emailInput = screen.getByLabelText(/Operator ID/i);
			expect(emailInput).toBeInTheDocument();
			expect(emailInput).toHaveAttribute('placeholder', 'operator@system.node');

			const passwordInput = screen.getByLabelText(/Access Key/i);
			expect(passwordInput).toBeInTheDocument();
			expect(passwordInput).toHaveAttribute('placeholder', '••••••••••••');

			// Check for the main action button
			const loginButton = screen.getByRole('button', { name: /Authorize Access/i });
			expect(loginButton).toBeInTheDocument();
		});
	});
});
