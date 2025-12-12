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
    cpu_usage: number;
    mem_used: number;
    mem_total: number;
    disk_used: number;
    disk_total: number;
    game_version: string;
}

export const spawners = writable<Spawner[]>([]);

export interface ServerVersion {
    id: number;
    filename: string;
    version: string;
    comment: string;
    uploaded_at: string; // ISO string from JSON
    is_active: boolean;
}

export const serverVersions = writable<ServerVersion[]>([]);

export interface ServerConfig {
    id: number;
    key: string;
    value: string;
    type: string;
    category: string;
    description: string;
    is_read_only: boolean;
    requires_restart: boolean;
    updated_at: string; // ISO string from JSON
    updated_by: string;
}

export const config = writable<ServerConfig[]>([]);
