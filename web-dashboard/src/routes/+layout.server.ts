import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
    const session = cookies.get('session');
    return {
        isAuthenticated: !!session
    };
};
