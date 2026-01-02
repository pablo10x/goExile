import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ fetch, cookies, url }) => {
	const session = cookies.get('session');
	const isLoginPath = url.pathname.startsWith('/login');
	
	if (!session) {
		if (!isLoginPath) {
			throw redirect(303, '/login');
		}
		return { isAuthenticated: false };
	}

	try {
		const statsRes = await fetch('/api/stats');
		
		const isValid = statsRes.ok;
		let stats = null;

		if (isValid) {
			stats = await statsRes.json();
		}
		
		if (!isValid && !isLoginPath) {
			throw redirect(303, '/login');
		}

		return {
			isAuthenticated: isValid,
			stats: stats
		};
	} catch (e) {
		if (!isLoginPath) {
			throw redirect(303, '/login');
		}
		return {
			isAuthenticated: false
		};
	}
};