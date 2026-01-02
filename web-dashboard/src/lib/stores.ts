import { writable } from 'svelte/store';

export const isAuthenticated = writable(false);
export const userEmail = writable('');

export const stats = writable({
	uptime: 0,
	active_nodes: 0,
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

export interface Node {
	id: number;
	name: string;
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

export const nodes = writable<Node[]>([]);

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

export interface Note {
	id: number;
	title: string;
	content: string;
	color: string; // e.g., 'yellow', 'blue', 'green'
	status: 'normal' | 'warn' | 'critical'; // Added for note status
	rotation: number;
	created_at: string;
	updated_at: string; // Added for consistency
}

export const notes = writable<Note[]>([]);

export interface Todo {
	id: number;
	parent_id?: number;
	content: string;
	done: boolean;
	in_progress: boolean;
	created_at: string;
	deadline?: string;
	sub_tasks?: Todo[];
	comments?: TodoComment[];
}

export interface TodoComment {
	id: number;
	todo_id: number;
	content: string;
	author?: string;
	created_at: string;
}

export const todos = writable<Todo[]>([]);

export interface Notification {
	id: string;
	type: 'success' | 'error' | 'info' | 'warning';
	message: string;
	details?: string;
	timeout: number;
	timestamp?: number;
}

export type NotificationInput = Omit<Notification, 'id' | 'timeout' | 'timestamp'> & { timeout?: number };

export const notifications = createNotificationStore();

export const isConnected = writable(false);
export const connectionStatus = writable('Connecting...');

// Helper for deep merging objects
function deepMerge(target: any, source: any) {
	if (!source) return target;
	const result = { ...target };
	for (const key in source) {
		if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
			result[key] = deepMerge(target[key] || {}, source[key]);
		} else {
			result[key] = source[key];
		}
	}
	return result;
}

// Helper for database-backed stores
function createDatabaseStore<T>(key: string, initialValue: T) {
	let storedValue;
	if (typeof window !== 'undefined') {
		const json = localStorage.getItem(key);
		if (json) {
			try {
				const parsed = JSON.parse(json);
				storedValue = deepMerge(initialValue, parsed);
			} catch (e) {
				console.error(`Error parsing persistent store ${key}:`, e);
			}
		}
	}

	const { subscribe, set, update } = writable<T>(storedValue || initialValue);
	let saveTimeout: ReturnType<typeof setTimeout> | null = null;

	async function saveToDB(value: T) {
		// Use debouncing to prevent spamming the server
		if (saveTimeout) clearTimeout(saveTimeout);
		
		saveTimeout = setTimeout(async () => {
			try {
				const response = await fetch(`/api/config/${key}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ value: JSON.stringify(value) })
				});
				if (response.status === 401) return;
			} catch (e) {
				// Network error
			}
		}, 500); // 500ms debounce
	}

	return {
		subscribe,
		set: (value: T) => {
			if (typeof window !== 'undefined') {
				localStorage.setItem(key, JSON.stringify(value));
			}
			set(value);
			saveToDB(value);
		},
		update: (fn: (value: T) => T) => {
			update((oldValue) => {
				const newValue = fn(oldValue);
				if (typeof window !== 'undefined') {
					localStorage.setItem(key, JSON.stringify(newValue));
				}
				saveToDB(newValue);
				return newValue;
			});
		},
		// Initialize from DB (Immediate, no debounce needed)
		init: (value: T) => {
			const merged = deepMerge(initialValue, value);
			if (typeof window !== 'undefined') {
				localStorage.setItem(key, JSON.stringify(merged));
			}
			set(merged);
		}
	};
}

export const theme = writable<'light' | 'dark'>('dark');

export const siteSettings = writable({
	site_name: 'EXILE',
	version_tag: 'v0.9.4-PROTOTYPE',
	dashboard: {
		show_topology: true,
		show_stats_cards: true,
		show_traffic_card: true,
		show_db_card: true,
		show_nodes_table: true,
		compact_mode: false
	},
	site_notice: {
		enabled: false,
		message: 'SYSTEM MAINTENANCE SCHEDULED FOR 0200 HOURS',
		type: 'info'
	}
});

/**
 * Load all settings from the database and initialize stores
 */
export async function loadAllSettings() {
	try {
        const response = await fetch('/api/config');
        if (response.status === 401) return;
        if (!response.ok) return;
        const configs: ServerConfig[] = await response.json();

		configs.forEach((cfg) => {
			try {
				if (cfg.key === 'site.settings') {
                    const val = JSON.parse(cfg.value);
                    siteSettings.update(s => ({...s, ...val}));
				}
			} catch (e) {
				console.error(`Error parsing config ${cfg.key}:`, e);
			}
		});
	} catch (e) {
		console.error('Failed to load settings from database:', e);
	}
}

function createNotificationStore() {
	const { subscribe, update } = writable<Notification[]>([]);
	const history = writable<Notification[]>([]);

	return {
		subscribe,
		history: { subscribe: history.subscribe },
		add: (n: NotificationInput) => {
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
			update(() => []);
		},
		clearPermanentHistory: () => {
			history.set([]);
		}
	};
}

export const showQuickActions = writable(true);
export const backgroundConfig = writable({
    global_type: 'digital_stream',
    show_global_background: true,
    show_smoke: true,
    show_rain: true,
    show_clouds: true,
    show_vignette: true,
    show_navbar_particles: true
});