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
	timeout?: number;
	timestamp?: number;
}

export const notifications = createNotificationStore();

export const isConnected = writable(false);
export const connectionStatus = writable('Connecting...');

// Helper for database-backed stores
function createDatabaseStore<T>(key: string, initialValue: T) {
	let storedValue;
	if (typeof window !== 'undefined') {
		const json = localStorage.getItem(key);
		if (json) {
			try {
				storedValue = JSON.parse(json);
			} catch (e) {
				console.error(`Error parsing persistent store ${key}:`, e);
			}
		}
	}

	const { subscribe, set, update } = writable<T>(storedValue || initialValue);

	async function saveToDB(value: T) {
		try {
			const response = await fetch(`/api/config/${key}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ value: JSON.stringify(value) })
			});
			if (response.status === 401) return; // Ignore unauthorized, expected during login
		} catch (e) {
			// Only log real network errors, not expected 401s
		}
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
		// Initialize from DB
		init: (value: T) => {
			if (typeof window !== 'undefined') {
				localStorage.setItem(key, JSON.stringify(value));
			}
			set(value);
		}
	};
}

export const theme = createDatabaseStore<'light' | 'dark'>('site.theme', 'dark');

const defaultBackgroundConfig = {
	show_smoke: true,
	show_rain: true,
	show_clouds: true,
	show_vignette: true,
	show_navbar_particles: true,
	rain_opacity: 0.2,
	clouds_opacity: 0.3,
	global_type: 'architecture' as 'architecture' | 'tactical_grid' | 'neural_network' | 'data_flow' | 'digital_horizon' | 'cyber_ocean' | 'static_void' | 'particle_network' | 'none',
	settings: {
		architecture: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		tactical_grid: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		neural_network: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		data_flow: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		digital_horizon: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		cyber_ocean: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		static_void: { intensity: 0.5, speed: 1, density: 1, color: '#f97316' },
		particle_network: { intensity: 0.5, speed: 1, density: 1, color: '#f97316', size: 1.0 }
	}
};

export const backgroundConfig = createDatabaseStore('site.background_config', defaultBackgroundConfig);

export const siteSettings = createDatabaseStore('site.settings', {
	site_name: 'EXILE',
	version_tag: 'v0.9.4-PROTOTYPE',
	aesthetic: {
		crt_effect: false,
		scanlines_opacity: 0.02,
		noise_opacity: 0.02,
		industrial_styling: true,
		glassmorphism: true,
		glow_effects: true,
		animations_enabled: true,
		topology_blobs: true,
		card_alpha: 0.8,
		backdrop_blur: 12,
		accent_color: '#f59e0bff',
		sidebar_alpha: 0.9,
		bg_opacity: 1.0,
		bg_color: '#0f172aff',
		primary_color: '#f59e0bff',
		secondary_color: '#334155ff',
		card_bg_color: '#1e293bff',
		hover_color: '#334155ff',
		text_color_primary: '#f1f5f9ff',
		text_color_secondary: '#888888',
		border_color: '#1e293b',
		font_primary: 'Inter',
		font_header: 'Kanit',
		font_body: 'Inter',
		font_mono: 'JetBrains Mono',
		border_radius_sm: 2,
		border_radius_md: 4,
		border_radius_lg: 8,
		glass_strength: 0.6, // opacity
		modal_animation: 'scale',
		card_border_width: 1,
		card_shadow_size: 4,
		scanline_speed: 4,
		scanline_density: 2,
		noise_intensity: 0.02,
		terminal_line_height: 1.5,
		panic_mode: false,
		industrial_border_color: '#334155ff',
		crt_curve: false,
		vignette_intensity: 0.2,
		border_glow_intensity: 0.2,
		text_glow: false,
		reduced_motion: false,
		mobile_optimized: true,
		sidebar_width: 260,
		card_glow_color: '#f59e0bff',
		font_size_base: 14,
		theme_preset: 'modern_industrial'
	},
	performance: {
		high_quality_smoke: false,
		particle_density: 0.5,
		low_power_mode: false,
		disable_expensive_animations: false
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
		if (response.status === 401) return; // Silent return if not logged in
		if (!response.ok) return;
		const configs: ServerConfig[] = await response.json();

		configs.forEach((cfg) => {
			try {
				if (cfg.key === 'site.theme') {
					theme.init(JSON.parse(cfg.value));
				} else if (cfg.key === 'site.settings') {
					siteSettings.init(JSON.parse(cfg.value));
				} else if (cfg.key === 'site.background_config') {
					const loaded = JSON.parse(cfg.value);
					// Deep merge settings to ensure new engines are present
					const mergedSettings = { ...defaultBackgroundConfig.settings, ...(loaded.settings || {}) };
					const merged = { ...defaultBackgroundConfig, ...loaded, settings: mergedSettings };
					backgroundConfig.init(merged);
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

export const showQuickActions = writable(true);
