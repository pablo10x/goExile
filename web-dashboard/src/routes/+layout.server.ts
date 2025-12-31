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
		// Use the proxied fetch which will include cookies
		const [statsRes, configRes] = await Promise.all([
			fetch('/api/stats'),
			fetch('/api/config')
		]);
		
		const isValid = statsRes.ok;
		let stats = null;
		let config = null;

		if (isValid) {
			stats = await statsRes.json();
			if (configRes.ok) {
				config = await configRes.json();
			}
		}
		
		if (!isValid && !isLoginPath) {
			throw redirect(303, '/login');
		}

		return {
			isAuthenticated: isValid,
			stats: stats,
			config: config
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
