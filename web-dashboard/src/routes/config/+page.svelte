<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { notifications } from '$lib/stores';
	import {
		Settings,
		Server,
		Cpu,
		Cloud,
		Shield,
		RefreshCw,
		Save,
		ChevronRight,
		ChevronDown,
		AlertCircle,
		CheckCircle2,
		Info,
		Lock,
		Unlock,
		RotateCcw,
		Plus,
		Trash2,
		Eye,
		EyeOff,
		Copy,
		ExternalLink,
		Zap,
		Database,
		Globe,
		Key,
		Clock,
		HardDrive,
		Network,
		Bell,
		Flame,
		CloudCog,
		FileJson,
		Search,
		Filter,
		X,
		Edit3,
		MoreVertical
	} from 'lucide-svelte';

	// Types
	interface ConfigItem {
		key: string;
		value: string;
		type: 'string' | 'int' | 'bool' | 'duration' | 'secret' | 'json' | 'url';
		category: string;
		description: string;
		is_read_only: boolean;
		requires_restart: boolean;
		updated_at: string;
		updated_by: string;
		validation?: {
			min?: number;
			max?: number;
			pattern?: string;
			options?: string[];
		};
	}

	interface ConfigSection {
		id: string;
		title: string;
		description: string;
		icon: any;
		color: string;
		gradient: string;
		items: ConfigItem[];
		expanded: boolean;
	}

	interface FirebaseConfig {
		key: string;
		value: string;
		valueType: 'string' | 'number' | 'boolean' | 'json';
		description: string;
		conditions?: string[];
		defaultValue: string;
		updated_at: string;
	}

	// State
	let loading = $state(true);
	let saving = $state(false);
	let activeTab = $state<'master' | 'spawner' | 'firebase'>('master');
	let searchQuery = $state('');
	let showSecrets = $state<Set<string>>(new Set());
	let pendingChanges = $state<Map<string, string>>(new Map());
	let expandedSections = $state<Set<string>>(new Set(['general', 'network', 'defaults', 'limits']));

	// Firebase modal state
	let showFirebaseModal = $state(false);
	let firebaseModalMode = $state<'create' | 'edit'>('create');
	let firebaseEditingKey = $state('');
	let firebaseForm = $state({
		key: '',
		value: '',
		valueType: 'STRING',
		description: ''
	});
	let firebaseSaving = $state(false);

	// Master Server Configuration Sections
	let masterSections = $state<ConfigSection[]>([
		{
			id: 'general',
			title: 'General Settings',
			description: 'Core server identification and behavior',
			icon: Settings,
			color: 'blue',
			gradient: 'from-blue-600 to-indigo-600',
			expanded: true,
			items: []
		},
		{
			id: 'network',
			title: 'Network & Connectivity',
			description: 'Ports, hosts, and connection settings',
			icon: Network,
			color: 'cyan',
			gradient: 'from-cyan-600 to-teal-600',
			expanded: true,
			items: []
		},
		{
			id: 'security',
			title: 'Security & Authentication',
			description: 'API keys, tokens, and access control',
			icon: Shield,
			color: 'red',
			gradient: 'from-red-600 to-rose-600',
			expanded: false,
			items: []
		},
		{
			id: 'database',
			title: 'Database Configuration',
			description: 'Database connections and pooling',
			icon: Database,
			color: 'purple',
			gradient: 'from-purple-600 to-violet-600',
			expanded: false,
			items: []
		},
		{
			id: 'performance',
			title: 'Performance & Limits',
			description: 'Resource limits and performance tuning',
			icon: Zap,
			color: 'amber',
			gradient: 'from-amber-600 to-orange-600',
			expanded: false,
			items: []
		}
	]);

	// Spawner Configuration Sections
	let spawnerSections = $state<ConfigSection[]>([
		{
			id: 'defaults',
			title: 'Default Settings',
			description: 'Default values for new spawners',
			icon: Cpu,
			color: 'green',
			gradient: 'from-green-600 to-emerald-600',
			expanded: true,
			items: []
		},
		{
			id: 'limits',
			title: 'Resource Limits',
			description: 'Instance limits and resource allocation',
			icon: HardDrive,
			color: 'orange',
			gradient: 'from-orange-600 to-red-600',
			expanded: true,
			items: []
		},
		{
			id: 'ports',
			title: 'Port Configuration',
			description: 'Game server port ranges',
			icon: Network,
			color: 'teal',
			gradient: 'from-teal-600 to-cyan-600',
			expanded: false,
			items: []
		},
		{
			id: 'updates',
			title: 'Auto-Update Settings',
			description: 'Automatic update behavior',
			icon: RefreshCw,
			color: 'indigo',
			gradient: 'from-indigo-600 to-blue-600',
			expanded: false,
			items: []
		}
	]);

	// Firebase Remote Config
	let firebaseConfigs = $state<FirebaseConfig[]>([]);
	let firebaseConnected = $state(false);
	let firebaseProjectId = $state('');

	// Derived values
	let pendingChangeCount = $derived(pendingChanges.size);
	let hasUnsavedChanges = $derived(pendingChangeCount > 0);

	let filteredMasterSections = $derived.by(() => {
		if (!searchQuery.trim()) return masterSections;
		const query = searchQuery.toLowerCase();
		return masterSections
			.map((section) => ({
				...section,
				items: section.items.filter(
					(item) =>
						item.key.toLowerCase().includes(query) ||
						item.description.toLowerCase().includes(query) ||
						item.value.toLowerCase().includes(query)
				)
			}))
			.filter((section) => section.items.length > 0);
	});

	let filteredSpawnerSections = $derived.by(() => {
		if (!searchQuery.trim()) return spawnerSections;
		const query = searchQuery.toLowerCase();
		return spawnerSections
			.map((section) => ({
				...section,
				items: section.items.filter(
					(item) =>
						item.key.toLowerCase().includes(query) ||
						item.description.toLowerCase().includes(query) ||
						item.value.toLowerCase().includes(query)
				)
			}))
			.filter((section) => section.items.length > 0);
	});

	let filteredFirebaseConfigs = $derived.by(() => {
		if (!searchQuery.trim()) return firebaseConfigs;
		const query = searchQuery.toLowerCase();
		return firebaseConfigs.filter(
			(config) =>
				config.key.toLowerCase().includes(query) || config.description.toLowerCase().includes(query)
		);
	});

	// Load configuration data
	async function loadConfig() {
		loading = true;
		try {
			const response = await fetch('/api/config');
			if (!response.ok) throw new Error('Failed to load configuration');

			const configs: ConfigItem[] = await response.json();

			// Distribute configs to sections based on category and key patterns
			distributeConfigs(configs);

			// Load Firebase config status
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Failed to load configuration',
				details: e.message
			});
		} finally {
			loading = false;
		}
	}

	function distributeConfigs(configs: ConfigItem[]) {
		// Reset all section items
		masterSections.forEach((s) => (s.items = []));
		spawnerSections.forEach((s) => (s.items = []));

		// Add default configs if empty (for demo/initial setup)
		if (!configs || configs.length === 0) {
			configs = getDefaultConfigs();
		}

		for (const config of configs) {
			if (config.category === 'system') {
				// Distribute to master sections
				if (
					config.key.includes('port') ||
					config.key.includes('host') ||
					config.key.includes('url')
				) {
					masterSections.find((s) => s.id === 'network')?.items.push(config);
				} else if (
					config.key.includes('key') ||
					config.key.includes('secret') ||
					config.key.includes('auth') ||
					config.key.includes('token')
				) {
					masterSections.find((s) => s.id === 'security')?.items.push(config);
				} else if (
					config.key.includes('db') ||
					config.key.includes('database') ||
					config.key.includes('pool')
				) {
					masterSections.find((s) => s.id === 'database')?.items.push(config);
				} else if (
					config.key.includes('max') ||
					config.key.includes('limit') ||
					config.key.includes('timeout') ||
					config.key.includes('ttl')
				) {
					masterSections.find((s) => s.id === 'performance')?.items.push(config);
				} else {
					masterSections.find((s) => s.id === 'general')?.items.push(config);
				}
			} else if (config.category === 'spawner') {
				// Distribute to spawner sections
				if (config.key.includes('port')) {
					spawnerSections.find((s) => s.id === 'ports')?.items.push(config);
				} else if (
					config.key.includes('max') ||
					config.key.includes('limit') ||
					config.key.includes('memory') ||
					config.key.includes('cpu')
				) {
					spawnerSections.find((s) => s.id === 'limits')?.items.push(config);
				} else if (config.key.includes('update') || config.key.includes('auto')) {
					spawnerSections.find((s) => s.id === 'updates')?.items.push(config);
				} else {
					spawnerSections.find((s) => s.id === 'defaults')?.items.push(config);
				}
			}
		}

		// Trigger reactivity
		masterSections = [...masterSections];
		spawnerSections = [...spawnerSections];
	}

	function getDefaultConfigs(): ConfigItem[] {
		const now = new Date().toISOString();
		return [
			// Master Server - General
			{
				key: 'server_name',
				value: 'Exile Master Server',
				type: 'string',
				category: 'system',
				description: 'Display name for the master server',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'environment',
				value: 'development',
				type: 'string',
				category: 'system',
				description: 'Current environment (development, staging, production)',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system',
				validation: { options: ['development', 'staging', 'production'] }
			},
			{
				key: 'log_level',
				value: 'info',
				type: 'string',
				category: 'system',
				description: 'Logging verbosity level',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system',
				validation: { options: ['debug', 'info', 'warn', 'error'] }
			},
			{
				key: 'enable_metrics',
				value: 'true',
				type: 'bool',
				category: 'system',
				description: 'Enable Prometheus metrics endpoint',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Master Server - Network
			{
				key: 'server_host',
				value: '0.0.0.0',
				type: 'string',
				category: 'system',
				description: 'Host address to bind the server',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'server_port',
				value: '8081',
				type: 'int',
				category: 'system',
				description: 'Port for the HTTP API server',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1, max: 65535 }
			},
			{
				key: 'websocket_port',
				value: '8082',
				type: 'int',
				category: 'system',
				description: 'Port for WebSocket connections',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1, max: 65535 }
			},
			{
				key: 'public_url',
				value: 'http://localhost:8081',
				type: 'url',
				category: 'system',
				description: 'Public URL for external access',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'allowed_origins',
				value: '*',
				type: 'string',
				category: 'system',
				description: 'CORS allowed origins (comma-separated)',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Master Server - Security
			{
				key: 'master_api_key',
				value: 'change-me-in-production',
				type: 'secret',
				category: 'system',
				description: 'API key for spawner authentication',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'jwt_secret',
				value: '',
				type: 'secret',
				category: 'system',
				description: 'Secret key for JWT token signing',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'session_timeout',
				value: '24h',
				type: 'duration',
				category: 'system',
				description: 'Dashboard session timeout duration',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'enable_2fa',
				value: 'true',
				type: 'bool',
				category: 'system',
				description: 'Require 2FA for admin login',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Master Server - Database
			{
				key: 'db_driver',
				value: 'pgx',
				type: 'string',
				category: 'system',
				description: 'Database driver (sqlite, pgx)',
				is_read_only: true,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'db_max_connections',
				value: '25',
				type: 'int',
				category: 'system',
				description: 'Maximum database connections',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1, max: 100 }
			},
			{
				key: 'db_idle_connections',
				value: '5',
				type: 'int',
				category: 'system',
				description: 'Idle database connections to keep',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 0, max: 50 }
			},

			// Master Server - Performance
			{
				key: 'max_body_size',
				value: '1048576',
				type: 'int',
				category: 'system',
				description: 'Maximum request body size (bytes)',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'request_timeout',
				value: '30s',
				type: 'duration',
				category: 'system',
				description: 'HTTP request timeout',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_ttl',
				value: '60s',
				type: 'duration',
				category: 'system',
				description: 'Spawner heartbeat timeout',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'cleanup_interval',
				value: '30s',
				type: 'duration',
				category: 'system',
				description: 'Interval for cleanup routines',
				is_read_only: false,
				requires_restart: true,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'rate_limit_requests',
				value: '100',
				type: 'int',
				category: 'system',
				description: 'Rate limit requests per minute',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Spawner - Defaults
			{
				key: 'spawner_default_region',
				value: 'us-east',
				type: 'string',
				category: 'spawner',
				description: 'Default region for new spawners',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_heartbeat_interval',
				value: '10s',
				type: 'duration',
				category: 'spawner',
				description: 'Heartbeat interval for spawners',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_game_binary',
				value: 'server.exe',
				type: 'string',
				category: 'spawner',
				description: 'Game server binary filename',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_game_args',
				value: '-dedicated -log',
				type: 'string',
				category: 'spawner',
				description: 'Default game server launch arguments',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Spawner - Limits
			{
				key: 'spawner_max_instances',
				value: '10',
				type: 'int',
				category: 'spawner',
				description: 'Maximum instances per spawner',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1, max: 100 }
			},
			{
				key: 'spawner_max_memory_mb',
				value: '16384',
				type: 'int',
				category: 'spawner',
				description: 'Maximum memory per spawner (MB)',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_max_cpu_percent',
				value: '80',
				type: 'int',
				category: 'spawner',
				description: 'Maximum CPU usage threshold (%)',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 10, max: 100 }
			},
			{
				key: 'instance_memory_limit_mb',
				value: '2048',
				type: 'int',
				category: 'spawner',
				description: 'Memory limit per instance (MB)',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},

			// Spawner - Ports
			{
				key: 'spawner_min_game_port',
				value: '7777',
				type: 'int',
				category: 'spawner',
				description: 'Minimum port for game servers',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1024, max: 65535 }
			},
			{
				key: 'spawner_max_game_port',
				value: '8000',
				type: 'int',
				category: 'spawner',
				description: 'Maximum port for game servers',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system',
				validation: { min: 1024, max: 65535 }
			},

			// Spawner - Updates
			{
				key: 'spawner_auto_update',
				value: 'true',
				type: 'bool',
				category: 'spawner',
				description: 'Enable automatic game server updates',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_update_check_interval',
				value: '5m',
				type: 'duration',
				category: 'spawner',
				description: 'How often to check for updates',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			},
			{
				key: 'spawner_update_restart_delay',
				value: '30s',
				type: 'duration',
				category: 'spawner',
				description: 'Delay before restarting after update',
				is_read_only: false,
				requires_restart: false,
				updated_at: now,
				updated_by: 'system'
			}
		];
	}

	async function loadFirebaseStatus() {
		try {
			const response = await fetch('/api/config/firebase/status');
			if (response.ok) {
				const status = await response.json();
				firebaseConnected = status.connected;
				firebaseProjectId = status.project_id || '';
				if (status.configs) {
					firebaseConfigs = status.configs;
				}
			}
		} catch (e) {
			// Firebase not configured
			firebaseConnected = false;
		}
	}

	function openFirebaseModal(mode: 'create' | 'edit', config?: FirebaseConfig) {
		firebaseModalMode = mode;
		if (mode === 'edit' && config) {
			firebaseEditingKey = config.key;
			firebaseForm = {
				key: config.key,
				value: config.value,
				valueType: config.valueType.toUpperCase(),
				description: config.description || ''
			};
		} else {
			firebaseEditingKey = '';
			firebaseForm = {
				key: '',
				value: '',
				valueType: 'STRING',
				description: ''
			};
		}
		showFirebaseModal = true;
	}

	function closeFirebaseModal() {
		showFirebaseModal = false;
		firebaseForm = { key: '', value: '', valueType: 'STRING', description: '' };
	}

	async function saveFirebaseParameter() {
		if (!firebaseForm.key.trim()) {
			notifications.add({ type: 'error', message: 'Parameter key is required' });
			return;
		}

		firebaseSaving = true;
		try {
			const endpoint = '/api/config/firebase/parameter';
			const method = firebaseModalMode === 'create' ? 'POST' : 'PUT';

			const response = await fetch(endpoint, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					key: firebaseForm.key,
					value: firebaseForm.value,
					valueType: firebaseForm.valueType,
					description: firebaseForm.description
				})
			});

			if (!response.ok) {
				const err = await response.json();
				throw new Error(err.error || 'Failed to save parameter');
			}

			notifications.add({
				type: 'success',
				message: `Parameter ${firebaseModalMode === 'create' ? 'created' : 'updated'} successfully`
			});

			closeFirebaseModal();
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: e.message || 'Failed to save parameter'
			});
		} finally {
			firebaseSaving = false;
		}
	}

	async function deleteFirebaseParameter(key: string) {
		if (!confirm(`Are you sure you want to delete "${key}"?`)) return;

		try {
			const response = await fetch('/api/config/firebase/parameter', {
				method: 'DELETE',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ key })
			});

			if (!response.ok) {
				const err = await response.json();
				throw new Error(err.error || 'Failed to delete parameter');
			}

			notifications.add({ type: 'success', message: 'Parameter deleted successfully' });
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({ type: 'error', message: e.message || 'Failed to delete parameter' });
		}
	}

	async function syncFirebaseConfig() {
		try {
			const response = await fetch('/api/config/firebase/sync', { method: 'POST' });
			if (!response.ok) {
				const err = await response.json();
				throw new Error(err.error || 'Sync failed');
			}
			notifications.add({ type: 'success', message: 'Configuration synced from Firebase' });
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({ type: 'error', message: e.message || 'Sync failed' });
		}
	}

	function handleValueChange(key: string, value: string, originalValue: string) {
		if (value !== originalValue) {
			pendingChanges.set(key, value);
		} else {
			pendingChanges.delete(key);
		}
		pendingChanges = new Map(pendingChanges);
	}

	function toggleSecret(key: string) {
		if (showSecrets.has(key)) {
			showSecrets.delete(key);
		} else {
			showSecrets.add(key);
		}
		showSecrets = new Set(showSecrets);
	}

	function toggleSection(sectionId: string) {
		if (expandedSections.has(sectionId)) {
			expandedSections.delete(sectionId);
		} else {
			expandedSections.add(sectionId);
		}
		expandedSections = new Set(expandedSections);
	}

	async function saveChanges() {
		if (pendingChanges.size === 0) return;

		saving = true;
		let successCount = 0;
		let errorCount = 0;
		let requiresRestart = false;

		try {
			for (const [key, value] of pendingChanges) {
				try {
					const response = await fetch(`/api/config/${key}`, {
						method: 'PUT',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ value })
					});

					if (response.ok) {
						successCount++;
						// Check if this config requires restart
						const allConfigs = [...masterSections, ...spawnerSections].flatMap((s) => s.items);
						const config = allConfigs.find((c) => c.key === key);
						if (config?.requires_restart) {
							requiresRestart = true;
						}
					} else {
						errorCount++;
					}
				} catch {
					errorCount++;
				}
			}

			if (successCount > 0) {
				notifications.add({
					type: 'success',
					message: `Saved ${successCount} configuration${successCount > 1 ? 's' : ''}`,
					details: requiresRestart ? 'Some changes require a server restart' : undefined
				});
				pendingChanges = new Map();
				await loadConfig();
			}

			if (errorCount > 0) {
				notifications.add({
					type: 'error',
					message: `Failed to save ${errorCount} configuration${errorCount > 1 ? 's' : ''}`
				});
			}
		} finally {
			saving = false;
		}
	}

	function discardChanges() {
		pendingChanges = new Map();
		loadConfig();
	}

	function copyToClipboard(value: string) {
		navigator.clipboard.writeText(value);
		notifications.add({ type: 'success', message: 'Copied to clipboard' });
	}

	function getInputType(configType: string): string {
		switch (configType) {
			case 'int':
				return 'number';
			case 'secret':
				return 'password';
			case 'url':
				return 'url';
			default:
				return 'text';
		}
	}

	onMount(() => {
		loadConfig();
	});
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
	<!-- Animated Background -->
	<div class="fixed inset-0 overflow-hidden pointer-events-none">
		<div
			class="absolute -top-40 -right-40 w-80 h-80 bg-blue-500/10 rounded-full blur-3xl animate-pulse"
		></div>
		<div
			class="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500/10 rounded-full blur-3xl animate-pulse"
			style="animation-delay: 1s;"
		></div>
	</div>

	<div class="relative z-10 p-6 max-w-7xl mx-auto">
		<!-- Header -->
		<div class="mb-8">
			<div class="flex items-center justify-between mb-6">
				<div class="flex items-center gap-4">
					<div
						class="p-3 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-2xl shadow-lg shadow-blue-900/30"
					>
						<Settings class="w-8 h-8 text-white" />
					</div>
					<div>
						<h1 class="text-3xl font-bold text-white">Configuration</h1>
						<p class="text-slate-400 mt-1">
							Manage server settings, spawner defaults, and remote configs
						</p>
					</div>
				</div>

				<div class="flex items-center gap-3">
					{#if hasUnsavedChanges}
						<div
							class="flex items-center gap-2 px-4 py-2 bg-orange-500/10 border border-orange-500/30 rounded-xl"
							transition:slide={{ axis: 'x' }}
						>
							<div class="w-2 h-2 rounded-full bg-orange-500 animate-pulse"></div>
							<span class="text-sm font-medium text-orange-400">{pendingChangeCount} unsaved</span>
						</div>
						<button
							onclick={discardChanges}
							class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition-all flex items-center gap-2"
						>
							<RotateCcw class="w-4 h-4" />
							Discard
						</button>
						<button
							onclick={saveChanges}
							disabled={saving}
							class="px-5 py-2 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-white rounded-xl font-semibold shadow-lg shadow-orange-900/30 transition-all flex items-center gap-2 disabled:opacity-50"
						>
							{#if saving}
								<RefreshCw class="w-4 h-4 animate-spin" />
								Saving...
							{:else}
								<Save class="w-4 h-4" />
								Save Changes
							{/if}
						</button>
					{:else}
						<button
							onclick={loadConfig}
							disabled={loading}
							class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition-all flex items-center gap-2 disabled:opacity-50"
						>
							<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
							Refresh
						</button>
					{/if}
				</div>
			</div>

			<!-- Tab Navigation -->
			<div
				class="flex items-center gap-2 p-1.5 bg-slate-800/50 rounded-2xl border border-slate-700/50 backdrop-blur-sm"
			>
				<button
					onclick={() => (activeTab = 'master')}
					class="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-medium transition-all {activeTab ===
					'master'
						? 'bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow-lg'
						: 'text-slate-400 hover:text-white hover:bg-slate-700/50'}"
				>
					<Server class="w-5 h-5" />
					Master Server
				</button>
				<button
					onclick={() => (activeTab = 'spawner')}
					class="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-medium transition-all {activeTab ===
					'spawner'
						? 'bg-gradient-to-r from-green-600 to-emerald-600 text-white shadow-lg'
						: 'text-slate-400 hover:text-white hover:bg-slate-700/50'}"
				>
					<Cpu class="w-5 h-5" />
					Spawner Defaults
				</button>
				<button
					onclick={() => (activeTab = 'firebase')}
					class="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-medium transition-all {activeTab ===
					'firebase'
						? 'bg-gradient-to-r from-orange-600 to-amber-600 text-white shadow-lg'
						: 'text-slate-400 hover:text-white hover:bg-slate-700/50'}"
				>
					<Flame class="w-5 h-5" />
					Firebase Remote Config
				</button>
			</div>

			<!-- Search Bar -->
			<div class="mt-4 relative">
				<Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500" />
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search configurations..."
					class="w-full pl-12 pr-10 py-3 bg-slate-800/50 border border-slate-700/50 rounded-xl text-slate-200 placeholder:text-slate-500 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
				/>
				{#if searchQuery}
					<button
						onclick={() => (searchQuery = '')}
						class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
					>
						<X class="w-4 h-4" />
					</button>
				{/if}
			</div>
		</div>

		<!-- Loading State -->
		{#if loading}
			<div class="flex items-center justify-center py-20" transition:fade>
				<div class="flex flex-col items-center gap-4">
					<div
						class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"
					></div>
					<span class="text-slate-400">Loading configuration...</span>
				</div>
			</div>
		{:else}
			<!-- Master Server Config -->
			{#if activeTab === 'master'}
				<div class="space-y-4" transition:fade={{ duration: 200 }}>
					{#each filteredMasterSections as section (section.id)}
						{@const isExpanded = expandedSections.has(section.id)}
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden"
						>
							<!-- Section Header -->
							<button
								onclick={() => toggleSection(section.id)}
								class="w-full px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors"
							>
								<div class="flex items-center gap-4">
									<div class="p-2.5 bg-gradient-to-br {section.gradient} rounded-xl shadow-lg">
										{#if section.id === 'general'}
											<Settings class="w-5 h-5 text-white" />
										{:else if section.id === 'network'}
											<Network class="w-5 h-5 text-white" />
										{:else if section.id === 'security'}
											<Shield class="w-5 h-5 text-white" />
										{:else if section.id === 'database'}
											<Database class="w-5 h-5 text-white" />
										{:else if section.id === 'performance'}
											<Zap class="w-5 h-5 text-white" />
										{/if}
									</div>
									<div class="text-left">
										<h3 class="text-lg font-semibold text-white">{section.title}</h3>
										<p class="text-sm text-slate-400">{section.description}</p>
									</div>
								</div>
								<div class="flex items-center gap-3">
									<span
										class="px-2.5 py-1 bg-slate-700/50 rounded-lg text-xs font-medium text-slate-400"
									>
										{section.items.length} settings
									</span>
									<div class="p-1 rounded-lg bg-slate-700/50">
										{#if isExpanded}
											<ChevronDown class="w-5 h-5 text-slate-400" />
										{:else}
											<ChevronRight class="w-5 h-5 text-slate-400" />
										{/if}
									</div>
								</div>
							</button>

							<!-- Section Content -->
							{#if isExpanded}
								<div class="px-6 pb-6 space-y-3" transition:slide={{ duration: 200 }}>
									{#each section.items as item (item.key)}
										{@const isPending = pendingChanges.has(item.key)}
										{@const isSecret = item.type === 'secret'}
										{@const showValue = showSecrets.has(item.key)}
										<div
											class="p-4 bg-slate-900/50 rounded-xl border border-slate-700/30 {isPending
												? 'ring-2 ring-orange-500/50 bg-orange-500/5'
												: ''} transition-all"
										>
											<div class="flex items-start justify-between gap-4">
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-2 mb-1">
														<span class="font-mono text-sm font-semibold text-slate-200"
															>{item.key}</span
														>
														{#if item.is_read_only}
															<span
																class="px-1.5 py-0.5 bg-slate-700 text-slate-400 text-[10px] font-bold rounded flex items-center gap-1"
															>
																<Lock class="w-3 h-3" />
																READ-ONLY
															</span>
														{/if}
														{#if item.requires_restart}
															<span
																class="px-1.5 py-0.5 bg-amber-500/20 text-amber-400 text-[10px] font-bold rounded flex items-center gap-1"
															>
																<RotateCcw class="w-3 h-3" />
																RESTART
															</span>
														{/if}
														{#if isPending}
															<span
																class="px-1.5 py-0.5 bg-orange-500/20 text-orange-400 text-[10px] font-bold rounded animate-pulse"
															>
																MODIFIED
															</span>
														{/if}
													</div>
													<p class="text-sm text-slate-500 mb-3">{item.description}</p>

													<!-- Input Field -->
													<div class="flex items-center gap-2">
														{#if item.type === 'bool'}
															<label class="relative inline-flex items-center cursor-pointer">
																<input
																	type="checkbox"
																	checked={(pendingChanges.get(item.key) ?? item.value) === 'true'}
																	onchange={(e) =>
																		handleValueChange(
																			item.key,
																			e.currentTarget.checked ? 'true' : 'false',
																			item.value
																		)}
																	disabled={item.is_read_only}
																	class="sr-only peer"
																/>
																<div
																	class="w-11 h-6 bg-slate-700 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600 peer-disabled:opacity-50"
																></div>
																<span class="ml-3 text-sm font-medium text-slate-300">
																	{(pendingChanges.get(item.key) ?? item.value) === 'true'
																		? 'Enabled'
																		: 'Disabled'}
																</span>
															</label>
														{:else if item.validation?.options}
															<select
																value={pendingChanges.get(item.key) ?? item.value}
																onchange={(e) =>
																	handleValueChange(item.key, e.currentTarget.value, item.value)}
																disabled={item.is_read_only}
																class="flex-1 max-w-xs px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-slate-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none disabled:opacity-50"
															>
																{#each item.validation.options as option}
																	<option value={option}>{option}</option>
																{/each}
															</select>
														{:else}
															<div class="flex-1 max-w-md relative">
																<input
																	type={isSecret && !showValue
																		? 'password'
																		: getInputType(item.type)}
																	value={pendingChanges.get(item.key) ?? item.value}
																	oninput={(e) =>
																		handleValueChange(item.key, e.currentTarget.value, item.value)}
																	disabled={item.is_read_only}
																	min={item.validation?.min}
																	max={item.validation?.max}
																	class="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-slate-200 font-mono text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none disabled:opacity-50 {isSecret
																		? 'pr-20'
																		: 'pr-10'}"
																/>
																{#if isSecret}
																	<button
																		onclick={() => toggleSecret(item.key)}
																		class="absolute right-10 top-1/2 -translate-y-1/2 p-1 text-slate-500 hover:text-slate-300"
																	>
																		{#if showValue}
																			<EyeOff class="w-4 h-4" />
																		{:else}
																			<Eye class="w-4 h-4" />
																		{/if}
																	</button>
																{/if}
																<button
																	onclick={() =>
																		copyToClipboard(pendingChanges.get(item.key) ?? item.value)}
																	class="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-slate-500 hover:text-slate-300"
																>
																	<Copy class="w-4 h-4" />
																</button>
															</div>
														{/if}
													</div>
												</div>

												<div class="text-right text-xs text-slate-600 shrink-0">
													<div>Type: <span class="text-slate-400">{item.type}</span></div>
													<div class="mt-1">
														Updated: <span class="text-slate-400"
															>{new Date(item.updated_at).toLocaleDateString()}</span
														>
													</div>
												</div>
											</div>
										</div>
									{/each}

									{#if section.items.length === 0}
										<div class="text-center py-8 text-slate-500">
											<Info class="w-8 h-8 mx-auto mb-2 opacity-50" />
											<p>No settings in this section</p>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/each}

					{#if filteredMasterSections.length === 0}
						<div class="text-center py-12">
							<Search class="w-12 h-12 mx-auto mb-4 text-slate-600" />
							<p class="text-slate-400">No matching configurations found</p>
							<button
								onclick={() => (searchQuery = '')}
								class="mt-4 text-blue-400 hover:text-blue-300"
							>
								Clear search
							</button>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Spawner Config -->
			{#if activeTab === 'spawner'}
				<div class="space-y-4" transition:fade={{ duration: 200 }}>
					<!-- Info Banner -->
					<div
						class="p-4 bg-green-500/10 border border-green-500/30 rounded-xl flex items-start gap-3"
					>
						<Info class="w-5 h-5 text-green-400 shrink-0 mt-0.5" />
						<div>
							<h4 class="font-medium text-green-400">Spawner Default Configuration</h4>
							<p class="text-sm text-green-300/70 mt-1">
								These settings are used as defaults when new spawners register with the master
								server. Individual spawners can override these values in their local configuration.
							</p>
						</div>
					</div>

					{#each filteredSpawnerSections as section (section.id)}
						{@const isExpanded = expandedSections.has(section.id)}
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden"
						>
							<!-- Section Header -->
							<button
								onclick={() => toggleSection(section.id)}
								class="w-full px-6 py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors"
							>
								<div class="flex items-center gap-4">
									<div class="p-2.5 bg-gradient-to-br {section.gradient} rounded-xl shadow-lg">
										{#if section.id === 'defaults'}
											<Cpu class="w-5 h-5 text-white" />
										{:else if section.id === 'limits'}
											<HardDrive class="w-5 h-5 text-white" />
										{:else if section.id === 'ports'}
											<Network class="w-5 h-5 text-white" />
										{:else if section.id === 'updates'}
											<RefreshCw class="w-5 h-5 text-white" />
										{/if}
									</div>
									<div class="text-left">
										<h3 class="text-lg font-semibold text-white">{section.title}</h3>
										<p class="text-sm text-slate-400">{section.description}</p>
									</div>
								</div>
								<div class="flex items-center gap-3">
									<span
										class="px-2.5 py-1 bg-slate-700/50 rounded-lg text-xs font-medium text-slate-400"
									>
										{section.items.length} settings
									</span>
									<div class="p-1 rounded-lg bg-slate-700/50">
										{#if isExpanded}
											<ChevronDown class="w-5 h-5 text-slate-400" />
										{:else}
											<ChevronRight class="w-5 h-5 text-slate-400" />
										{/if}
									</div>
								</div>
							</button>

							<!-- Section Content -->
							{#if isExpanded}
								<div class="px-6 pb-6 space-y-3" transition:slide={{ duration: 200 }}>
									{#each section.items as item (item.key)}
										{@const isPending = pendingChanges.has(item.key)}
										<div
											class="p-4 bg-slate-900/50 rounded-xl border border-slate-700/30 {isPending
												? 'ring-2 ring-orange-500/50 bg-orange-500/5'
												: ''} transition-all"
										>
											<div class="flex items-start justify-between gap-4">
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-2 mb-1">
														<span class="font-mono text-sm font-semibold text-slate-200"
															>{item.key}</span
														>
														{#if item.requires_restart}
															<span
																class="px-1.5 py-0.5 bg-amber-500/20 text-amber-400 text-[10px] font-bold rounded flex items-center gap-1"
															>
																<RotateCcw class="w-3 h-3" />
																RESTART
															</span>
														{/if}
														{#if isPending}
															<span
																class="px-1.5 py-0.5 bg-orange-500/20 text-orange-400 text-[10px] font-bold rounded animate-pulse"
															>
																MODIFIED
															</span>
														{/if}
													</div>
													<p class="text-sm text-slate-500 mb-3">{item.description}</p>

													<!-- Input Field -->
													<div class="flex items-center gap-2">
														{#if item.type === 'bool'}
															<label class="relative inline-flex items-center cursor-pointer">
																<input
																	type="checkbox"
																	checked={(pendingChanges.get(item.key) ?? item.value) === 'true'}
																	onchange={(e) =>
																		handleValueChange(
																			item.key,
																			e.currentTarget.checked ? 'true' : 'false',
																			item.value
																		)}
																	class="sr-only peer"
																/>
																<div
																	class="w-11 h-6 bg-slate-700 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-green-500/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"
																></div>
																<span class="ml-3 text-sm font-medium text-slate-300">
																	{(pendingChanges.get(item.key) ?? item.value) === 'true'
																		? 'Enabled'
																		: 'Disabled'}
																</span>
															</label>
														{:else}
															<div class="flex-1 max-w-md relative">
																<input
																	type={getInputType(item.type)}
																	value={pendingChanges.get(item.key) ?? item.value}
																	oninput={(e) =>
																		handleValueChange(item.key, e.currentTarget.value, item.value)}
																	min={item.validation?.min}
																	max={item.validation?.max}
																	class="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-slate-200 font-mono text-sm focus:border-green-500 focus:ring-2 focus:ring-green-500/20 outline-none pr-10"
																/>
																<button
																	onclick={() =>
																		copyToClipboard(pendingChanges.get(item.key) ?? item.value)}
																	class="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-slate-500 hover:text-slate-300"
																>
																	<Copy class="w-4 h-4" />
																</button>
															</div>
														{/if}
													</div>
												</div>

												<div class="text-right text-xs text-slate-600 shrink-0">
													<div>Type: <span class="text-slate-400">{item.type}</span></div>
												</div>
											</div>
										</div>
									{/each}

									{#if section.items.length === 0}
										<div class="text-center py-8 text-slate-500">
											<Info class="w-8 h-8 mx-auto mb-2 opacity-50" />
											<p>No settings in this section</p>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<!-- Firebase Remote Config -->
			{#if activeTab === 'firebase'}
				<div class="space-y-6" transition:fade={{ duration: 200 }}>
					<!-- Connection Status Card -->
					<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-4">
								<div
									class="p-3 bg-gradient-to-br from-orange-600 to-amber-600 rounded-xl shadow-lg"
								>
									<Flame class="w-6 h-6 text-white" />
								</div>
								<div>
									<h3 class="text-xl font-bold text-white">Firebase Remote Config</h3>
									<p class="text-slate-400">Manage remote configuration for game clients</p>
								</div>
							</div>

							<div class="flex items-center gap-3">
								{#if firebaseConnected}
									<div
										class="flex items-center gap-2 px-4 py-2 bg-green-500/10 border border-green-500/30 rounded-xl"
									>
										<CheckCircle2 class="w-5 h-5 text-green-400" />
										<span class="text-green-400 font-medium">Connected</span>
									</div>
								{:else}
									<div
										class="flex items-center gap-2 px-4 py-2 bg-slate-700/50 border border-slate-600 rounded-xl"
									>
										<AlertCircle class="w-5 h-5 text-slate-400" />
										<span class="text-slate-400 font-medium">Not Configured</span>
									</div>
								{/if}
							</div>
						</div>

						{#if firebaseConnected && firebaseProjectId}
							<div class="mt-4 p-3 bg-slate-900/50 rounded-xl flex items-center gap-3">
								<Globe class="w-5 h-5 text-slate-500" />
								<span class="text-sm text-slate-400">Project ID:</span>
								<code class="px-2 py-1 bg-slate-800 rounded text-orange-400 text-sm"
									>{firebaseProjectId}</code
								>
								<a
									href="https://console.firebase.google.com/project/{firebaseProjectId}/config"
									target="_blank"
									rel="noopener"
									class="ml-auto flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
								>
									Open Firebase Console
									<ExternalLink class="w-4 h-4" />
								</a>
							</div>
						{/if}
					</div>

					{#if !firebaseConnected}
						<!-- Setup Instructions -->
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
						>
							<h4 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
								<CloudCog class="w-5 h-5 text-orange-400" />
								Setup Firebase Remote Config
							</h4>

							<div class="space-y-4">
								<div class="flex items-start gap-4 p-4 bg-slate-900/50 rounded-xl">
									<div
										class="w-8 h-8 bg-orange-500/20 rounded-full flex items-center justify-center shrink-0"
									>
										<span class="text-orange-400 font-bold">1</span>
									</div>
									<div>
										<h5 class="font-medium text-slate-200">Create Firebase Project</h5>
										<p class="text-sm text-slate-400 mt-1">
											Go to Firebase Console and create a new project or select an existing one.
										</p>
									</div>
								</div>

								<div class="flex items-start gap-4 p-4 bg-slate-900/50 rounded-xl">
									<div
										class="w-8 h-8 bg-orange-500/20 rounded-full flex items-center justify-center shrink-0"
									>
										<span class="text-orange-400 font-bold">2</span>
									</div>
									<div>
										<h5 class="font-medium text-slate-200">Download Service Account Key</h5>
										<p class="text-sm text-slate-400 mt-1">
											Navigate to Project Settings → Service Accounts → Generate new private key.
										</p>
									</div>
								</div>

								<div class="flex items-start gap-4 p-4 bg-slate-900/50 rounded-xl">
									<div
										class="w-8 h-8 bg-orange-500/20 rounded-full flex items-center justify-center shrink-0"
									>
										<span class="text-orange-400 font-bold">3</span>
									</div>
									<div>
										<h5 class="font-medium text-slate-200">Configure Environment</h5>
										<p class="text-sm text-slate-400 mt-1">
											Set the <code class="px-1 py-0.5 bg-slate-800 rounded text-orange-400 text-xs"
												>GOOGLE_APPLICATION_CREDENTIALS</code
											> environment variable to the path of your service account key file.
										</p>
									</div>
								</div>

								<div class="flex items-start gap-4 p-4 bg-slate-900/50 rounded-xl">
									<div
										class="w-8 h-8 bg-orange-500/20 rounded-full flex items-center justify-center shrink-0"
									>
										<span class="text-orange-400 font-bold">4</span>
									</div>
									<div>
										<h5 class="font-medium text-slate-200">Set Project ID</h5>
										<p class="text-sm text-slate-400 mt-1">
											Add <code class="px-1 py-0.5 bg-slate-800 rounded text-orange-400 text-xs"
												>FIREBASE_PROJECT_ID</code
											> to your environment variables.
										</p>
									</div>
								</div>
							</div>

							<div
								class="mt-6 p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl flex items-start gap-3"
							>
								<AlertCircle class="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
								<div>
									<p class="text-sm text-amber-300">
										After configuring, restart the master server to connect to Firebase.
									</p>
								</div>
							</div>
						</div>
					{:else}
						<!-- Remote Config Parameters -->
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden"
						>
							<div class="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
								<h4 class="text-lg font-semibold text-white flex items-center gap-2">
									<FileJson class="w-5 h-5 text-orange-400" />
									Remote Config Parameters
								</h4>
								<div class="flex items-center gap-2">
									<button
										onclick={syncFirebaseConfig}
										class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-xl font-medium transition-all flex items-center gap-2"
									>
										<RefreshCw class="w-4 h-4" />
										Sync
									</button>
									<button
										onclick={() => openFirebaseModal('create')}
										class="px-4 py-2 bg-orange-600 hover:bg-orange-500 text-white rounded-xl font-medium transition-all flex items-center gap-2"
									>
										<Plus class="w-4 h-4" />
										Add Parameter
									</button>
								</div>
							</div>

							<div class="p-6">
								{#if filteredFirebaseConfigs.length > 0}
									<div class="space-y-3">
										{#each filteredFirebaseConfigs as config (config.key)}
											<div
												class="p-4 bg-slate-900/50 rounded-xl border border-slate-700/30 hover:border-orange-500/30 transition-all group"
											>
												<div class="flex items-start justify-between gap-4">
													<div class="flex-1">
														<div class="flex items-center gap-2 mb-1">
															<span class="font-mono text-sm font-semibold text-slate-200"
																>{config.key}</span
															>
															<span
																class="px-1.5 py-0.5 bg-orange-500/20 text-orange-400 text-[10px] font-bold rounded uppercase"
															>
																{config.valueType}
															</span>
														</div>
														<p class="text-sm text-slate-500 mb-3">
															{config.description || 'No description'}
														</p>
														<div class="flex items-center gap-2">
															<code
																class="px-3 py-1.5 bg-slate-800 rounded-lg text-slate-300 text-sm font-mono"
															>
																{config.valueType === 'json'
																	? JSON.stringify(JSON.parse(config.value), null, 0).slice(0, 50) +
																		'...'
																	: config.value}
															</code>
														</div>
													</div>
													<div
														class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
													>
														<button
															onclick={() => openFirebaseModal('edit', config)}
															class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg"
														>
															<Edit3 class="w-4 h-4" />
														</button>
														<button
															onclick={() => deleteFirebaseParameter(config.key)}
															class="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg"
														>
															<Trash2 class="w-4 h-4" />
														</button>
													</div>
												</div>
											</div>
										{/each}
									</div>
								{:else}
									<div class="text-center py-12">
										<FileJson class="w-12 h-12 mx-auto mb-4 text-slate-600" />
										<p class="text-slate-400 mb-2">No remote config parameters yet</p>
										<p class="text-sm text-slate-500">
											Add parameters to configure your game clients remotely
										</p>
									</div>
								{/if}
							</div>
						</div>

						<!-- Publish Changes -->
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
						>
							<div class="flex items-center justify-between">
								<div>
									<h4 class="text-lg font-semibold text-white">Publish Changes</h4>
									<p class="text-slate-400 text-sm mt-1">
										Push your configuration changes to all connected clients
									</p>
								</div>
								<button
									class="px-6 py-2.5 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-white rounded-xl font-semibold shadow-lg shadow-orange-900/30 transition-all flex items-center gap-2"
								>
									<Cloud class="w-5 h-5" />
									Publish to Firebase
								</button>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>

<!-- Firebase Parameter Modal -->
{#if showFirebaseModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		onclick={(e) => e.target === e.currentTarget && closeFirebaseModal()}
		transition:fade={{ duration: 150 }}
	>
		<div
			class="w-full max-w-lg bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl overflow-hidden"
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<!-- Modal Header -->
			<div class="px-6 py-4 border-b border-slate-700 flex items-center justify-between">
				<h3 class="text-xl font-bold text-white flex items-center gap-2">
					<Flame class="w-5 h-5 text-orange-400" />
					{firebaseModalMode === 'create' ? 'Add Parameter' : 'Edit Parameter'}
				</h3>
				<button
					onclick={closeFirebaseModal}
					class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Modal Body -->
			<div class="p-6 space-y-4">
				<!-- Key -->
				<div>
					<label class="block text-sm font-medium text-slate-300 mb-2">Parameter Key</label>
					<input
						type="text"
						bind:value={firebaseForm.key}
						disabled={firebaseModalMode === 'edit'}
						placeholder="e.g. feature_enabled"
						class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none disabled:opacity-50"
					/>
					<p class="text-xs text-slate-500 mt-1">Alphanumeric characters and underscores only</p>
				</div>

				<!-- Value Type -->
				<div>
					<label class="block text-sm font-medium text-slate-300 mb-2">Value Type</label>
					<select
						bind:value={firebaseForm.valueType}
						class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none"
					>
						<option value="STRING">String</option>
						<option value="NUMBER">Number</option>
						<option value="BOOLEAN">Boolean</option>
						<option value="JSON">JSON</option>
					</select>
				</div>

				<!-- Value -->
				<div>
					<label class="block text-sm font-medium text-slate-300 mb-2">Default Value</label>
					{#if firebaseForm.valueType === 'BOOLEAN'}
						<select
							bind:value={firebaseForm.value}
							class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none"
						>
							<option value="true">true</option>
							<option value="false">false</option>
						</select>
					{:else if firebaseForm.valueType === 'JSON'}
						<textarea
							bind:value={firebaseForm.value}
							rows={4}
							placeholder="Enter JSON value..."
							class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none font-mono text-sm"
						></textarea>
					{:else}
						<input
							type={firebaseForm.valueType === 'NUMBER' ? 'number' : 'text'}
							bind:value={firebaseForm.value}
							placeholder="Enter value..."
							class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none"
						/>
					{/if}
				</div>

				<!-- Description -->
				<div>
					<label class="block text-sm font-medium text-slate-300 mb-2">Description (optional)</label
					>
					<textarea
						bind:value={firebaseForm.description}
						rows={2}
						placeholder="Describe what this parameter does..."
						class="w-full px-4 py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none"
					></textarea>
				</div>
			</div>

			<!-- Modal Footer -->
			<div class="px-6 py-4 border-t border-slate-700 flex items-center justify-end gap-3">
				<button
					onclick={closeFirebaseModal}
					class="px-5 py-2.5 bg-slate-700 hover:bg-slate-600 text-white rounded-xl font-medium transition-all"
				>
					Cancel
				</button>
				<button
					onclick={saveFirebaseParameter}
					disabled={firebaseSaving || !firebaseForm.key.trim()}
					class="px-5 py-2.5 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-white rounded-xl font-semibold shadow-lg shadow-orange-900/30 transition-all flex items-center gap-2 disabled:opacity-50"
				>
					{#if firebaseSaving}
						<RefreshCw class="w-4 h-4 animate-spin" />
						Saving...
					{:else}
						<Save class="w-4 h-4" />
						{firebaseModalMode === 'create' ? 'Create Parameter' : 'Update Parameter'}
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(body) {
		background: rgb(2, 6, 23);
	}
</style>
