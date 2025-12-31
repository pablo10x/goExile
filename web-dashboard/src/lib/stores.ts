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
	timeout?: number;
	timestamp?: number;
}

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

export const theme = createDatabaseStore<'light' | 'dark'>('site.theme', 'dark');

const defaultBackgroundConfig = {
	show_global_background: true,
	show_smoke: true,
	show_rain: true,
	show_clouds: true,
	show_vignette: true,
	show_navbar_particles: true,
	rain_opacity: 0.2,
	clouds_opacity: 0.3,
	global_type: 'digital_stream' as 'digital_stream' | 'circuit_grid' | 'neon_pulse' | 'noise_static' | 'glass_refraction' | 'cyber_scan' | 'vector_wave' | 'none',
	card_hover_effect: true,
	settings: {
		digital_stream: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		circuit_grid: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		neon_pulse: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		noise_static: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		glass_refraction: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		cyber_scan: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 },
		vector_wave: { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 }
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
		bg_color: '#18181bff', /* Lighter default (was #0f172a) */
		primary_color: '#f59e0bff',
		secondary_color: '#334155ff',
		card_bg_color: '#27272aff', /* Lighter default (was #1e293b) */
		hover_color: '#3f3f46ff',
		text_color_primary: '#f1f5f9ff',
		text_color_secondary: '#888888',
		text_color_dim: '#666666ff',
		border_color: '#3f3f46',
		header_bg_color: '#000000cc',
		sidebar_bg_color: '#18181bcc',
		scrollbar_thumb_color: '#3f3f46',
		scrollbar_track_color: '#00000033',
		success_color: '#10b981ff',
		warning_color: '#f59e0bff',
		danger_color: '#ef4444ff',
		info_color: '#06b6d4ff',
		terminal_bg_color: '#050505ff',
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
		theme_preset: 'modern_industrial',
		icon_pack: 'lucide' as 'lucide' | 'mdi' | 'ph' | 'ri' | 'tabler',
		icon_stroke: 2,
		// Advanced Atmospheric
		glitch_intensity: 0.05,
		chromatic_aberration: 0.02,
		flicker_intensity: 0.01,
		scanline_type: 'subtle' as 'subtle' | 'heavy' | 'rgb' | 'none',
		noise_type: 'analog' as 'asfalt' | 'analog' | 'grain' | 'none',
		// Advanced Geometry
		corner_style: 'clipped' as 'clipped' | 'rounded' | 'sharp',
		border_opacity: 0.3,
		grid_opacity: 0.05,
		// Micro-Typography
		letter_spacing: 0.05, // em
		line_height: 1.5,
		text_transform: 'uppercase' as 'uppercase' | 'none',
		heading_weight: '900',
		font_weight_base: '400',
		font_weight_mono: '400',
		text_glow_intensity: 0.5,
		paragraph_spacing: 1.0, // em
		// Kinetic Physics (Animations)
		global_transition_speed: 300, // ms
		hover_scale_factor: 1.02,
		button_press_depth: 2, // px
		ui_animation_intensity: 1.0,
		spin_velocity: 1.0,
		scanline_speed_hz: 1.0,
		modal_entry_speed: 300, // ms
		glow_pulse_depth: 0.5,
		// Signal Dynamics (New Tab)
		typing_speed: 1.0,
		glitch_frequency: 1.0,
		header_pulse_speed: 1.0,
		navbar_entry_delay: 100, // ms
		progress_fill_speed: 1.0,
		global_animation_scale: 1.0,
		// Background & Hover Dynamics
		bg_anim_speed: 1.0,
		bg_anim_opacity: 0.2,
		card_hover_intensity: 0.5,
		card_hover_data_speed: 1.0,
		card_hover_spark_density: 8
	},
	performance: {
		high_quality_smoke: false,
		particle_density: 0.5,
		low_power_mode: false,
		disable_expensive_animations: false
	},
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
export async function loadAllSettings(prefetchedConfigs?: ServerConfig[]) {
	try {
		let configs: ServerConfig[];
		
		if (prefetchedConfigs) {
			configs = prefetchedConfigs;
		} else {
			const response = await fetch('/api/config');
			if (response.status === 401) return; // Silent return if not logged in
			if (!response.ok) return;
			configs = await response.json();
		}

		configs.forEach((cfg) => {
			try {
				if (cfg.key === 'site.theme') {
					theme.init(JSON.parse(cfg.value));
				} else if (cfg.key === 'site.settings') {
					siteSettings.init(JSON.parse(cfg.value));
				} else if (cfg.key === 'site.background_config') {
					backgroundConfig.init(JSON.parse(cfg.value));
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
			update(() => []);
		},
		clearPermanentHistory: () => {
			history.set([]);
		}
	};
}

export const showQuickActions = writable(true);
