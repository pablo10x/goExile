import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const session = event.cookies.get('session');
	const url = event.url.pathname;

	// Public routes that don't require authentication
	// We allow /login because it's the login page itself
	// We allow /_app because it serves static assets and JS chunks
	// We allow /api because those are handled by the Go backend proxy and return 401 JSON, not redirect HTML
	if (url.startsWith('/login') || url.startsWith('/_app') || url.startsWith('/api')) {
		return await resolve(event);
	}

	// If no session and trying to access a protected page, redirect to login immediately
	if (!session) {
		throw redirect(303, '/login');
	}

	// If session exists but is not fully authenticated (e.g. during 2FA flow)
	// This part needs to be handled by +layout.server.ts or within API checks
	// For now, if a session exists, we assume +layout.server.ts will validate its step.

	return await resolve(event);
};
