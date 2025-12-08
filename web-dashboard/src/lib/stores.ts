import { writable } from 'svelte/store';

export const isAuthenticated = writable(false);
export const userEmail = writable('');

export const stats = writable({
    uptime: 0,
    active_spawners: 0,
    total_requests: 0,
    total_errors: 0,
    db_connected: false,
    memory_usage: 0,
    bytes_sent: 0,
    bytes_received: 0
});

export interface Spawner {
    id: number;
    region: string;
    host: string;
    port: number;
    status: string;
    current_instances: number;
    max_instances: number;
}

export const spawners = writable<Spawner[]>([]);

export interface ServerVersion {
    id: number;
    filename: string;
    comment: string;
    uploaded_at: string; // ISO string from JSON
    is_active: boolean;
}

export const serverVersions = writable<ServerVersion[]>([]);
