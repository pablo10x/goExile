import { browser } from '$app/environment';

// Detect if we are running in a native environment (Tauri)
// Tauri v2 uses 'tauri://localhost' on Windows
export const isNative = browser && (
    (window as any).__TAURI_INTERNALS__ !== undefined ||
    window.location.protocol === 'tauri:' ||
    window.location.protocol === 'asset:'
);

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

// Determine base URL
// In development, we usually proxy through Vite, but in native we need the full URL
export const API_BASE = isNative 
    ? 'http://localhost:8081' 
    : '';

/**
 * Universal fetch wrapper that handles:
 * 1. Base URL prepending for native apps
 * 2. Absolute URL handling
 */
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
    const url = path.startsWith('http') ? path : `${API_BASE}${path.startsWith('/') ? '' : '/'}${path}`;
    
    // Get stored session token
    const token = browser ? localStorage.getItem('exile_session') : null;

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
                localStorage.setItem('exile_session', data.session);
            }
        } catch (e) {}
    }

    return response;
}
