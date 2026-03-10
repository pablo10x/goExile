import { apiFetch } from '$lib/api';
import { isAuthenticated } from '$lib/stores.svelte';
import { get } from 'svelte/store';
import type { LayoutLoad } from './$types';

export const prerender = false;
export const ssr = false;

export const load: LayoutLoad = async ({ url }) => {
	const isLoginPath = url.pathname.startsWith('/login');

	// In SPA mode, we can check localStorage directly
	const token = typeof window !== 'undefined' ? localStorage.getItem('exile_session') : null;

	if (!token) {
		return { isAuthenticated: false };
	}

	// If we have a token, we might want to verify it if not already authenticated
	// But to keep it fast, we can just return true and let the layout handle the actual verification
	// Or we can do a quick check here.
	try {
		// If we are already authenticated in the store, don't bother re-checking on every route change
		if (get(isAuthenticated)) {
			return { isAuthenticated: true };
		}

		// Otherwise, verify the token
		const res = await apiFetch('/api/stats');
		return { isAuthenticated: res.ok };
	} catch (e) {
		return { isAuthenticated: false };
	}
};
