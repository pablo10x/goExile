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
	bytes_received: 0,
	db_open_connections: 0,
	db_in_use: 0,
	db_idle: 0,
	db_wait_count: 0,
	db_wait_duration: '0s',
	db_max_lifetime_closed: 0,
	db_max_idle_closed: 0,
	db_size: '',
	db_commits: 0,
	db_rollbacks: 0,
	db_cache_hit: 0,
	db_tup_fetched: 0,
	db_tup_inserted: 0,
	db_tup_updated: 0,
	db_tup_deleted: 0
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
	last_seen?: string;
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

export const restartRequired = writable(false);

export interface Notification {
	id: string;
	type: 'success' | 'error' | 'info' | 'warning';
	message: string;
	details?: string;
	timeout?: number;
	timestamp?: number;
}

export const notifications = createNotificationStore();

export const isConnected = writable(false);
export const connectionStatus = writable('Connecting...');

function createNotificationStore() {
	const { subscribe, update } = writable<Notification[]>([]);
	const history = writable<Notification[]>([]);

	return {
		subscribe,
		history: { subscribe: history.subscribe },
		add: (n: Omit<Notification, 'id'>) => {
			const id = crypto.randomUUID();
			const notification = { ...n, id, timestamp: Date.now(), timeout: n.timeout ?? 5000 };

			// Add to active notifications
			update((notifications) => [...notifications, notification]);

			// Add to history (limit to 50)
			history.update((h) => [notification, ...h].slice(0, 50));

			if (notification.timeout > 0) {
				setTimeout(() => {
					update((notifications) => notifications.filter((n) => n.id !== id));
				}, notification.timeout);
			}
		},
		remove: (id: string) => {
			update((notifications) => notifications.filter((n) => n.id !== id));
		},
		clearHistory: () => {
			history.set([]);
		}
	};
}
