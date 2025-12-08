import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import Page from './+page.svelte';

describe('Login Page', () => {
    it('renders login form', () => {
        render(Page);
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
