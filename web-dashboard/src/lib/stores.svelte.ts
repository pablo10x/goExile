import { writable, get } from 'svelte/store';
export * from './types/models';
import type { Node, Note, Todo, ServerConfig } from './types/models';

// Svelte 5 Runes Store implementation for maximum performance
class SystemState {
	isAuthenticated = $state(false);
	userEmail = $state('');
	isConnected = $state(false);
	connectionStatus = $state('Connecting...');
	restartRequired = $state(false);
	showQuickActions = $state(true);
	lowPowerMode = $state(false);

	stats = $state({
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

	nodes = $state<Node[]>([]);
	serverVersions = $state<any[]>([]);
	config = $state<ServerConfig[]>([]);
	notes = $state<Note[]>([]);
	todos = $state<Todo[]>([]);

	siteSettings = $state({
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

	backgroundConfig = $state({
		global_type: 'digital_stream',
		show_global_background: true,
		show_smoke: true,
		show_rain: true,
		show_clouds: true,
		show_vignette: true,
		show_navbar_particles: true
	});

	constructor() {
		// Initialize from localStorage if available
		if (typeof window !== 'undefined') {
			this.loadPersistedState();
		}
	}

	loadPersistedState() {
		const keys = ['theme', 'siteSettings', 'backgroundConfig'];
		keys.forEach(key => {
			const saved = localStorage.getItem(key);
			if (saved) {
				try {
					const parsed = JSON.parse(saved);
					// @ts-ignore
					Object.assign(this[key], parsed);
				} catch (e) {}
			}
		});
	}

	async saveSetting(key: string, value: any) {
		if (typeof window !== 'undefined') {
			localStorage.setItem(key, JSON.stringify(value));
		}
		try {
			await fetch(`/api/config/${key}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ value: JSON.stringify(value) })
			});
		} catch (e) {}
	}
}

// Global state instance
export const sysState = new SystemState();

// Backward compatibility wrappers for stores (bridging writable to runes)
// This allows existing components to work while we transition
function bridge<T>(getter: () => T, setter: (val: T) => void) {
	const { subscribe, set, update } = writable<T>(getter());
	return {
		subscribe,
		set: (val: T) => { set(val); setter(val); },
		update: (fn: (v: T) => T) => {
			const current = getter();
			const next = fn(current);
			set(next);
			setter(next);
		}
	};
}

export const isAuthenticated = bridge(() => sysState.isAuthenticated, (v) => sysState.isAuthenticated = v);
export const userEmail = bridge(() => sysState.userEmail, (v) => sysState.userEmail = v);
export const isConnected = bridge(() => sysState.isConnected, (v) => sysState.isConnected = v);
export const connectionStatus = bridge(() => sysState.connectionStatus, (v) => sysState.connectionStatus = v);
export const restartRequired = bridge(() => sysState.restartRequired, (v) => sysState.restartRequired = v);
export const showQuickActions = bridge(() => sysState.showQuickActions, (v) => sysState.showQuickActions = v);
export const lowPowerMode = bridge(() => sysState.lowPowerMode, (v) => sysState.lowPowerMode = v);
export const stats = bridge(() => sysState.stats, (v) => sysState.stats = v);
export const nodes = bridge(() => sysState.nodes, (v) => sysState.nodes = v);
export const serverVersions = bridge(() => sysState.serverVersions, (v) => sysState.serverVersions = v);
export const config = bridge(() => sysState.config, (v) => sysState.config = v);
export const notes = bridge(() => sysState.notes, (v) => sysState.notes = v);
export const todos = bridge(() => sysState.todos, (v) => sysState.todos = v);
export const siteSettings = bridge(() => sysState.siteSettings, (v) => sysState.siteSettings = v);
export const backgroundConfig = bridge(() => sysState.backgroundConfig, (v) => sysState.backgroundConfig = v);
export const theme = writable<'light' | 'dark'>('dark');

/**
 * Load all settings from the database and initialize stores
 */
export async function loadAllSettings() {
	try {
        const response = await fetch('/api/config');
        if (response.status === 401) return;
        if (!response.ok) return;
        const configs: any[] = await response.json();

		configs.forEach((cfg) => {
			try {
				if (cfg.key === 'site.settings') {
                    const val = JSON.parse(cfg.value);
                    sysState.siteSettings = { ...sysState.siteSettings, ...val };
                    siteSettings.set(sysState.siteSettings);
				}
			} catch (e) {
				console.error(`Error parsing config ${cfg.key}:`, e);
			}
		});
	} catch (e) {
		console.error('Failed to load settings from database:', e);
	}
}

// Notification store remains writable for now as it's event-based
function createNotificationStore() {
	const { subscribe, update, set } = writable<any[]>([]);
	const history = writable<any[]>([]);

	return {
		subscribe,
		history: { subscribe: history.subscribe },
		add: (n: any) => {
			const id = crypto.randomUUID();
			const notification = { ...n, id, timestamp: Date.now(), timeout: n.timeout ?? 5000 };
			update((notifications) => [...notifications, notification]);
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
		clearPermanentHistory: () => {
			history.set([]);
		}
	};
}

export const notifications = createNotificationStore();
