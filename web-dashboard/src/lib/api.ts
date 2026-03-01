import { browser } from '$app/environment';

// Detect if we are running in a native environment (Tauri)
export const isNative = typeof window !== 'undefined' && (
    (window as any).__TAURI_INTERNALS__ !== undefined ||
    window.location.protocol === 'tauri:' ||
    window.location.hostname === 'tauri.localhost' ||
    window.location.protocol === 'asset:'
);

// Determine base URL
// In native (Tauri), we prioritize the user-configured URL from localStorage.
// Fallback to localhost for dev/initial setup.
export const getApiBase = () => {
    if (typeof window !== 'undefined') {
        const stored = localStorage.getItem('server_url');
        if (stored) return stored.replace(/\/$/, ''); // Remove trailing slash
    }
    return (isNative || !import.meta.env.DEV) ? 'http://localhost:8081' : '';
};

export const API_BASE = getApiBase();

/**
 * Sends a native system notification if running in Tauri
 */
export async function notify(title: string, body: string) {
    if (!isNative) return;
    
    try {
        const { isPermissionGranted, requestPermission, sendNotification } = await import('@tauri-apps/plugin-notification');
        
        let permissionGranted = await isPermissionGranted();
        if (!permissionGranted) {
            const permission = await requestPermission();
            permissionGranted = permission === 'granted';
        }
        
        if (permissionGranted) {
            sendNotification({ title, body });
        }
    } catch (e) {
        console.error('Failed to send native notification:', e);
    }
}

/**
 * Universal fetch wrapper that handles:
 * 1. Base URL prepending for native apps
 * 2. Absolute URL handling
 */
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
    const base = getApiBase();
    const url = path.startsWith('http') ? path : `${base}${path.startsWith('/') ? '' : '/'}${path}`;
    
    // Get stored session token
    const token = (typeof window !== 'undefined') ? localStorage.getItem('exile_session') : null;

    if (path.includes('/api/auth/login') === false) {
        console.log(`[API] Fetching ${url} | Token attached: ${!!token} (${token ? token.substring(0, 5) + '...' : 'null'})`);
        if (typeof window !== 'undefined') {
            console.log(`[API] All LocalStorage keys:`, Object.keys(localStorage));
        }
    }

    const options: RequestInit = {
        ...init,
        headers: {
            ...init?.headers,
            // Add Bearer token for native app auth
            ...(token ? { 'Authorization': `Bearer ${token}` } : {})
        },
        credentials: 'include'
    };

    const response = await fetch(url, options);

    // If this was a login request and successful, extract the session token
    if (path.includes('/api/auth/login') && response.ok) {
        // We clone to allow the caller to also read the body
        const cloned = response.clone();
        try {
            const data = await cloned.json();
            if (data.session) {
                console.log(`[API] Login successful, saving token: ${data.session.substring(0, 5)}...`);
                localStorage.setItem('exile_session', data.session);
            }
        } catch (e) {
            console.error(`[API] Failed to parse login response:`, e);
        }
    }

    return response;
}