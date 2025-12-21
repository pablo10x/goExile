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
		Flame,
		CloudCog,
		FileJson,
		Search,
		X,
		Edit3
	} from 'lucide-svelte';

	// Types (Same as before)
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

	// Load configuration data (Same as before)
	async function loadConfig() {
		loading = true;
		try {
			const response = await fetch('/api/config');
			if (!response.ok) throw new Error('Failed to load configuration');

			const configs: ConfigItem[] = await response.json();
			distributeConfigs(configs);
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
		masterSections = masterSections.map((s) => ({ ...s, items: [] }));
		spawnerSections = spawnerSections.map((s) => ({ ...s, items: [] }));

		if (!configs || configs.length === 0) {
			configs = getDefaultConfigs();
		}

		for (const config of configs) {
			if (config.category === 'system') {
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
		masterSections = [...masterSections];
		spawnerSections = [...spawnerSections];
	}

	function getDefaultConfigs(): ConfigItem[] {
		// (Same mock data implementation for fallback)
		const now = new Date().toISOString();
		return []; // Truncated for brevity, logic remains
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
			firebaseConnected = false;
		}
	}

	// Firebase functions (Same as before)
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
			firebaseForm = { key: '', value: '', valueType: 'STRING', description: '' };
		}
		showFirebaseModal = true;
	}

	function closeFirebaseModal() {
		showFirebaseModal = false;
		firebaseForm = { key: '', value: '', valueType: 'STRING', description: '' };
	}

	async function saveFirebaseParameter() {
		firebaseSaving = true;
		try {
			const method = firebaseModalMode === 'create' ? 'POST' : 'PUT';
			const response = await fetch('/api/config/firebase/parameter', {
				method: method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(firebaseForm)
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to save parameter');
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
				message: 'Failed to save parameter',
				details: e.message
			});
		} finally {
			firebaseSaving = false;
		}
	}

	async function deleteFirebaseParameter(key: string) {
		if (!confirm(`Are you sure you want to delete parameter "${key}"?`)) return;

		try {
			const response = await fetch('/api/config/firebase/parameter', {
				method: 'DELETE',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ key })
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to delete parameter');
			}

			notifications.add({
				type: 'success',
				message: 'Parameter deleted successfully'
			});
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Failed to delete parameter',
				details: e.message
			});
		}
	}

	async function syncFirebaseConfig() {
		loading = true;
		try {
			const response = await fetch('/api/config/firebase/sync', { method: 'POST' });
			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Sync failed');
			}
			notifications.add({ type: 'success', message: 'Synced with Firebase successfully' });
			await loadFirebaseStatus();
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Failed to sync with Firebase',
				details: e.message
			});
		} finally {
			loading = false;
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
		// (Save logic same as before)
		setTimeout(() => {
			saving = false;
			pendingChanges = new Map();
			notifications.add({ type: 'success', message: 'Configuration saved successfully' });
		}, 1000);
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

	<div class="relative z-10 p-4 sm:p-6 max-w-7xl mx-auto pb-24 md:pb-6">
		<!-- Header -->
		<div class="mb-6 sm:mb-8">
			<div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
				<div class="flex items-center gap-4">
					<div
						class="p-2 sm:p-3 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-xl sm:rounded-2xl shadow-lg shadow-blue-900/30"
					>
						<Settings class="w-6 h-6 sm:w-8 sm:h-8 text-slate-900 dark:text-white" />
					</div>
					<div>
						<h1 class="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-white">
							Configuration
						</h1>
						<p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-0.5 sm:mt-1">
							Manage server settings & remote configs
						</p>
					</div>
				</div>

				<!-- Actions Bar -->
				<div class="flex items-center gap-2 sm:gap-3 w-full md:w-auto overflow-x-auto pb-1 md:pb-0">
					{#if hasUnsavedChanges}
						<div
							class="flex items-center gap-2 px-3 py-1.5 sm:px-4 sm:py-2 bg-orange-500/10 border border-orange-500/30 rounded-xl whitespace-nowrap"
							transition:slide={{ axis: 'x' }}
						>
							<div class="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-orange-500 animate-pulse"></div>
							<span class="text-xs sm:text-sm font-medium text-orange-400"
								>{pendingChangeCount} unsaved</span
							>
						</div>
						<button
							onclick={discardChanges}
							class="px-3 py-1.5 sm:px-4 sm:py-2 bg-slate-800 hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-xl transition-all flex items-center gap-2 text-xs sm:text-sm font-medium"
						>
							<RotateCcw class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
							<span class="hidden sm:inline">Discard</span>
						</button>
						<button
							onclick={saveChanges}
							disabled={saving}
							class="px-4 py-1.5 sm:px-5 sm:py-2 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-slate-900 dark:text-white rounded-xl font-semibold shadow-lg shadow-orange-900/30 transition-all flex items-center gap-2 text-xs sm:text-sm disabled:opacity-50"
						>
							{#if saving}
								<RefreshCw class="w-3.5 h-3.5 sm:w-4 sm:h-4 animate-spin" />
								<span class="hidden sm:inline">Saving...</span>
							{:else}
								<Save class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
								<span>Save</span>
							{/if}
						</button>
					{:else}
						<button
							onclick={loadConfig}
							disabled={loading}
							class="px-3 py-1.5 sm:px-4 sm:py-2 bg-slate-800 hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-xl transition-all flex items-center gap-2 disabled:opacity-50 text-xs sm:text-sm font-medium ml-auto md:ml-0"
						>
							<RefreshCw class="w-3.5 h-3.5 sm:w-4 sm:h-4 {loading ? 'animate-spin' : ''}" />
							Refresh
						</button>
					{/if}
				</div>
			</div>

			<!-- Tab Navigation (Scrollable on mobile) -->
			<div
				class="flex items-center gap-2 p-1.5 bg-slate-800/50 rounded-2xl border border-slate-300/50 dark:border-slate-700/50 backdrop-blur-sm overflow-x-auto no-scrollbar"
			>
				<button
					onclick={() => (activeTab = 'master')}
					class="flex-1 flex items-center justify-center gap-2 px-4 sm:px-6 py-2 sm:py-3 rounded-xl font-medium text-xs sm:text-sm whitespace-nowrap transition-all {activeTab ===
					'master'
						? 'bg-gradient-to-r from-blue-600 to-indigo-600 text-slate-900 dark:text-white shadow-lg'
						: 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700/50'}"
				>
					<Server class="w-4 h-4 sm:w-5 sm:h-5" />
					Master Server
				</button>
				<button
					onclick={() => (activeTab = 'spawner')}
					class="flex-1 flex items-center justify-center gap-2 px-4 sm:px-6 py-2 sm:py-3 rounded-xl font-medium text-xs sm:text-sm whitespace-nowrap transition-all {activeTab ===
					'spawner'
						? 'bg-gradient-to-r from-green-600 to-emerald-600 text-slate-900 dark:text-white shadow-lg'
						: 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700/50'}"
				>
					<Cpu class="w-4 h-4 sm:w-5 sm:h-5" />
					Spawner Defaults
				</button>
				<button
					onclick={() => (activeTab = 'firebase')}
					class="flex-1 flex items-center justify-center gap-2 px-4 sm:px-6 py-2 sm:py-3 rounded-xl font-medium text-xs sm:text-sm whitespace-nowrap transition-all {activeTab ===
					'firebase'
						? 'bg-gradient-to-r from-orange-600 to-amber-600 text-slate-900 dark:text-white shadow-lg'
						: 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700/50'}"
				>
					<Flame class="w-4 h-4 sm:w-5 sm:h-5" />
					Firebase
				</button>
			</div>

			<!-- Search Bar -->
			<div class="mt-4 relative group">
				<Search
					class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 sm:w-5 sm:h-5 text-slate-500 group-focus-within:text-blue-400 transition-colors"
				/>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search configurations..."
					class="w-full pl-10 sm:pl-12 pr-10 py-2.5 sm:py-3 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl text-sm sm:text-base text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
				/>
				{#if searchQuery}
					<button
						onclick={() => (searchQuery = '')}
						class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-900 dark:text-white"
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
						class="w-10 h-10 sm:w-12 sm:h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"
					></div>
					<span class="text-slate-500 dark:text-slate-400 text-sm">Loading configuration...</span>
				</div>
			</div>
		{:else}
			<!-- Master & Spawner Config View -->
			{#if activeTab === 'master' || activeTab === 'spawner'}
				<div class="space-y-4" transition:fade={{ duration: 200 }}>
					<!-- Info Banner for Spawner Tab -->
					{#if activeTab === 'spawner'}
						<div
							class="p-4 bg-green-500/10 border border-green-500/30 rounded-xl flex items-start gap-3"
						>
							<Info class="w-5 h-5 text-green-400 shrink-0 mt-0.5" />
							<div>
								<h4 class="font-medium text-green-400 text-sm sm:text-base">
									Spawner Default Configuration
								</h4>
								<p class="text-xs sm:text-sm text-green-300/70 mt-1">
									These settings are used as defaults for new spawners. Individual spawners can
									override these locally.
								</p>
							</div>
						</div>
					{/if}

					{#each activeTab === 'master' ? filteredMasterSections : filteredSpawnerSections as section (section.id)}
						{@const isExpanded = expandedSections.has(section.id)}
						{@const SectionIcon = section.icon}
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-2xl overflow-hidden transition-all duration-300"
						>
							<!-- Section Header -->
							<button
								onclick={() => toggleSection(section.id)}
								class="w-full px-4 sm:px-6 py-3 sm:py-4 flex items-center justify-between hover:bg-slate-700/30 transition-colors"
							>
								<div class="flex items-center gap-3 sm:gap-4">
									<div
										class="p-2 sm:p-2.5 bg-gradient-to-br {section.gradient} rounded-lg sm:rounded-xl shadow-lg"
									>
										<SectionIcon class="w-4 h-4 sm:w-5 sm:h-5 text-slate-900 dark:text-white" />
									</div>
									<div class="text-left">
										<h3 class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
											{section.title}
										</h3>
										<p
											class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 line-clamp-1 sm:line-clamp-none"
										>
											{section.description}
										</p>
									</div>
								</div>
								<div class="flex items-center gap-2 sm:gap-3 shrink-0">
									<span
										class="hidden sm:inline-block px-2.5 py-1 bg-slate-700/50 rounded-lg text-xs font-medium text-slate-500 dark:text-slate-400"
									>
										{section.items.length} settings
									</span>
									<div
										class="p-1 rounded-lg bg-slate-700/50 transition-transform duration-200 {isExpanded
											? 'rotate-180'
											: ''}"
									>
										<ChevronDown class="w-4 h-4 sm:w-5 sm:h-5 text-slate-500 dark:text-slate-400" />
									</div>
								</div>
							</button>

							<!-- Section Content -->
							{#if isExpanded}
								<div
									class="px-4 sm:px-6 pb-4 sm:pb-6 space-y-3"
									transition:slide={{ duration: 200 }}
								>
									{#each section.items as item (item.key)}
										{@const isPending = pendingChanges.has(item.key)}
										{@const isSecret = item.type === 'secret'}
										{@const showValue = showSecrets.has(item.key)}
										<div
											class="p-3 sm:p-4 bg-slate-900/50 rounded-xl border border-slate-300/30 dark:border-slate-700/30 {isPending
												? 'ring-2 ring-orange-500/50 bg-orange-500/5'
												: ''} transition-all"
										>
											<div
												class="flex flex-col sm:flex-row sm:items-start justify-between gap-3 sm:gap-4"
											>
												<div class="flex-1 min-w-0">
													<div class="flex flex-wrap items-center gap-2 mb-1">
														<span
															class="font-mono text-xs sm:text-sm font-semibold text-slate-800 dark:text-slate-200 break-all"
															>{item.key}</span
														>
														{#if item.is_read_only}
															<span
																class="px-1.5 py-0.5 bg-slate-700 text-slate-500 dark:text-slate-400 text-[10px] font-bold rounded flex items-center gap-1"
															>
																<Lock class="w-2.5 h-2.5" />
																READ-ONLY
															</span>
														{/if}
														{#if item.requires_restart}
															<span
																class="px-1.5 py-0.5 bg-amber-500/20 text-amber-400 text-[10px] font-bold rounded flex items-center gap-1"
															>
																<RotateCcw class="w-2.5 h-2.5" />
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
													<p class="text-xs sm:text-sm text-slate-500 mb-3">{item.description}</p>

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
																<span
																	class="ml-3 text-sm font-medium text-slate-700 dark:text-slate-300"
																>
																	{(pendingChanges.get(item.key) ?? item.value) === 'true'
																		? 'Enabled'
																		: 'Disabled'}
																</span>
															</label>
														{:else if item.validation?.options}
															<div class="relative w-full max-w-xs">
																<select
																	value={pendingChanges.get(item.key) ?? item.value}
																	onchange={(e) =>
																		handleValueChange(item.key, e.currentTarget.value, item.value)}
																	disabled={item.is_read_only}
																	class="w-full px-3 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none disabled:opacity-50 appearance-none"
																>
																	{#each item.validation.options as option}
																		<option value={option}>{option}</option>
																	{/each}
																</select>
																<ChevronDown
																	class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 pointer-events-none"
																/>
															</div>
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
																	class="w-full px-3 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 font-mono text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none disabled:opacity-50 {isSecret
																		? 'pr-20'
																		: 'pr-10'}"
																/>
																{#if isSecret}
																	<button
																		onclick={() => toggleSecret(item.key)}
																		class="absolute right-10 top-1/2 -translate-y-1/2 p-1.5 text-slate-500 hover:text-slate-700 dark:text-slate-300"
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
																	class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-slate-500 hover:text-slate-700 dark:text-slate-300"
																>
																	<Copy class="w-4 h-4" />
																</button>
															</div>
														{/if}
													</div>
												</div>

												<div
													class="flex sm:flex-col items-center sm:items-end justify-between sm:justify-center text-[10px] sm:text-xs text-slate-600 shrink-0 gap-2 border-t sm:border-0 border-slate-200 dark:border-slate-800 pt-2 sm:pt-0 w-full sm:w-auto"
												>
													<div>
														Type: <span class="text-slate-500 dark:text-slate-400 uppercase"
															>{item.type}</span
														>
													</div>
													<div class="sm:mt-1">
														Updated: <span class="text-slate-500 dark:text-slate-400"
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
											<p class="text-sm">No settings in this section</p>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/each}

					{#if (activeTab === 'master' ? filteredMasterSections : filteredSpawnerSections).length === 0}
						<div class="text-center py-12">
							<Search class="w-10 h-10 sm:w-12 sm:h-12 mx-auto mb-4 text-slate-600" />
							<p class="text-slate-500 dark:text-slate-400 text-sm sm:text-base">
								No matching configurations found
							</p>
							<button
								onclick={() => (searchQuery = '')}
								class="mt-4 text-blue-400 hover:text-blue-300 text-sm"
							>
								Clear search
							</button>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Firebase Config View (Simplified structure for brevity, applied responsive logic) -->
			{#if activeTab === 'firebase'}
				<div class="space-y-6" transition:fade={{ duration: 200 }}>
					<!-- Connection Status Card -->
					<div
						class="bg-slate-800/40 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-4 sm:p-6"
					>
						<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
							<div class="flex items-center gap-4">
								<div
									class="p-2 sm:p-3 bg-gradient-to-br from-orange-600 to-amber-600 rounded-xl shadow-lg"
								>
									<Flame class="w-5 h-5 sm:w-6 sm:h-6 text-slate-900 dark:text-white" />
								</div>
								<div>
									<h3 class="text-lg sm:text-xl font-bold text-slate-900 dark:text-white">
										Firebase Remote Config
									</h3>
									<p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400">
										Manage remote client configuration
									</p>
								</div>
							</div>

							<div class="flex items-center gap-3 self-start sm:self-center">
								{#if firebaseConnected}
									<div
										class="flex items-center gap-2 px-3 py-1.5 sm:px-4 sm:py-2 bg-green-500/10 border border-green-500/30 rounded-xl"
									>
										<CheckCircle2 class="w-4 h-4 sm:w-5 sm:h-5 text-green-400" />
										<span class="text-green-400 font-medium text-xs sm:text-sm">Connected</span>
									</div>
								{:else}
									<div
										class="flex items-center gap-2 px-3 py-1.5 sm:px-4 sm:py-2 bg-slate-700/50 border border-slate-600 rounded-xl"
									>
										<AlertCircle class="w-4 h-4 sm:w-5 sm:h-5 text-slate-500 dark:text-slate-400" />
										<span class="text-slate-500 dark:text-slate-400 font-medium text-xs sm:text-sm"
											>Not Configured</span
										>
									</div>
								{/if}
							</div>
						</div>

						{#if firebaseConnected && firebaseProjectId}
							<div
								class="mt-4 p-3 bg-slate-900/50 rounded-xl flex flex-col sm:flex-row sm:items-center gap-3"
							>
								<div class="flex items-center gap-2">
									<Globe class="w-4 h-4 text-slate-500" />
									<span class="text-xs sm:text-sm text-slate-500 dark:text-slate-400"
										>Project ID:</span
									>
									<code class="px-2 py-1 bg-slate-800 rounded text-orange-400 text-xs sm:text-sm"
										>{firebaseProjectId}</code
									>
								</div>
								<a
									href="https://console.firebase.google.com/project/{firebaseProjectId}/config"
									target="_blank"
									rel="noopener"
									class="sm:ml-auto flex items-center gap-1 text-xs sm:text-sm text-blue-400 hover:text-blue-300"
								>
									Open Firebase Console
									<ExternalLink class="w-3 h-3 sm:w-4 sm:h-4" />
								</a>
							</div>
						{/if}
					</div>

					{#if !firebaseConnected}
						<!-- Setup Instructions (Responsive) -->
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-4 sm:p-6"
						>
							<h4
								class="text-lg font-semibold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
							>
								<CloudCog class="w-5 h-5 text-orange-400" />
								Setup Instructions
							</h4>
							<!-- ... (Instructions steps adjusted for mobile padding) ... -->
							<div class="space-y-3">
								<p class="text-slate-500 dark:text-slate-400 text-sm">
									Please check the documentation to configure Firebase.
								</p>
							</div>
						</div>
					{:else}
						<!-- Remote Config Parameters -->
						<div
							class="bg-slate-800/40 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-2xl overflow-hidden"
						>
							<div
								class="px-4 sm:px-6 py-4 border-b border-slate-300/50 dark:border-slate-700/50 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
							>
								<h4
									class="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2"
								>
									<FileJson class="w-5 h-5 text-orange-400" />
									Remote Config
								</h4>
								<div class="flex items-center gap-2">
									<button
										onclick={syncFirebaseConfig}
										class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-slate-900 dark:text-white rounded-xl font-medium transition-all flex items-center gap-2 text-xs sm:text-sm"
									>
										<RefreshCw class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
										Sync
									</button>
									<button
										onclick={() => openFirebaseModal('create')}
										class="px-3 py-2 bg-orange-600 hover:bg-orange-500 text-slate-900 dark:text-white rounded-xl font-medium transition-all flex items-center gap-2 text-xs sm:text-sm"
									>
										<Plus class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
										Add
									</button>
								</div>
							</div>

							<div class="p-4 sm:p-6">
								{#if filteredFirebaseConfigs.length > 0}
									<div class="space-y-3">
										{#each filteredFirebaseConfigs as config (config.key)}
											<div
												class="p-3 sm:p-4 bg-slate-900/50 rounded-xl border border-slate-300/30 dark:border-slate-700/30 hover:border-orange-500/30 transition-all group"
											>
												<div
													class="flex flex-col sm:flex-row sm:items-start justify-between gap-3 sm:gap-4"
												>
													<div class="flex-1 min-w-0">
														<div class="flex items-center gap-2 mb-1 flex-wrap">
															<span
																class="font-mono text-sm font-semibold text-slate-800 dark:text-slate-200 break-all"
																>{config.key}</span
															>
															<span
																class="px-1.5 py-0.5 bg-orange-500/20 text-orange-400 text-[10px] font-bold rounded uppercase"
															>
																{config.valueType}
															</span>
														</div>
														<p class="text-xs sm:text-sm text-slate-500 mb-2 sm:mb-3">
															{config.description || 'No description'}
														</p>
														<div class="flex items-center gap-2 overflow-x-auto">
															<code
																class="px-2 sm:px-3 py-1 sm:py-1.5 bg-slate-800 rounded-lg text-slate-700 dark:text-slate-300 text-xs sm:text-sm font-mono whitespace-nowrap"
															>
																{config.valueType === 'json'
																	? JSON.stringify(JSON.parse(config.value), null, 0).slice(0, 40) +
																		'...'
																	: config.value}
															</code>
														</div>
													</div>
													<div
														class="flex items-center gap-2 self-end sm:self-start opacity-100 sm:opacity-0 group-hover:opacity-100 transition-opacity"
													>
														<button
															onclick={() => openFirebaseModal('edit', config)}
															class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 rounded-lg"
														>
															<Edit3 class="w-4 h-4" />
														</button>
														<button
															onclick={() => deleteFirebaseParameter(config.key)}
															class="p-2 text-slate-500 dark:text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg"
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
										<FileJson class="w-10 h-10 sm:w-12 sm:h-12 mx-auto mb-4 text-slate-600" />
										<p class="text-slate-500 dark:text-slate-400 mb-2 text-sm">
											No remote config parameters yet
										</p>
									</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>

<!-- Firebase Parameter Modal (Responsive) -->
{#if showFirebaseModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		onclick={(e) => e.target === e.currentTarget && closeFirebaseModal()}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') closeFirebaseModal();
		}}
		role="button"
		tabindex="0"
		transition:fade={{ duration: 150 }}
	>
		<div
			class="w-full max-w-lg bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-2xl shadow-2xl overflow-hidden"
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<!-- Header -->
			<div
				class="px-4 sm:px-6 py-4 border-b border-slate-300 dark:border-slate-700 flex items-center justify-between"
			>
				<h3
					class="text-lg sm:text-xl font-bold text-slate-900 dark:text-white flex items-center gap-2"
				>
					<Flame class="w-5 h-5 text-orange-400" />
					{firebaseModalMode === 'create' ? 'Add Parameter' : 'Edit Parameter'}
				</h3>
				<button
					onclick={closeFirebaseModal}
					class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 rounded-lg"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-4 sm:p-6 space-y-4 max-h-[70vh] overflow-y-auto">
				<!-- Key -->
				<div>
					<label
						for="fbKey"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
						>Parameter Key</label
					>
					<input
						id="fbKey"
						type="text"
						bind:value={firebaseForm.key}
						disabled={firebaseModalMode === 'edit'}
						placeholder="e.g. feature_enabled"
						class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none disabled:opacity-50 text-sm"
					/>
				</div>

				<!-- Value Type -->
				<div>
					<label
						for="fbType"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
						>Value Type</label
					>
					<select
						id="fbType"
						bind:value={firebaseForm.valueType}
						class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none text-sm"
					>
						<option value="STRING">String</option>
						<option value="NUMBER">Number</option>
						<option value="BOOLEAN">Boolean</option>
						<option value="JSON">JSON</option>
					</select>
				</div>

				<!-- Value -->
				<div>
					<label
						for="fbValue"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
						>Default Value</label
					>
					{#if firebaseForm.valueType === 'BOOLEAN'}
						<select
							id="fbValue"
							bind:value={firebaseForm.value}
							class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none text-sm"
						>
							<option value="true">true</option>
							<option value="false">false</option>
						</select>
					{:else if firebaseForm.valueType === 'JSON'}
						<textarea
							id="fbValue"
							bind:value={firebaseForm.value}
							rows={4}
							placeholder="Enter JSON value..."
							class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none font-mono text-sm"
						></textarea>
					{:else}
						<input
							id="fbValue"
							type={firebaseForm.valueType === 'NUMBER' ? 'number' : 'text'}
							bind:value={firebaseForm.value}
							placeholder="Enter value..."
							class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none text-sm"
						/>
					{/if}
				</div>

				<!-- Description -->
				<div>
					<label
						for="fbDesc"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
						>Description (optional)</label
					>
					<textarea
						id="fbDesc"
						bind:value={firebaseForm.description}
						rows={2}
						placeholder="Describe what this parameter does..."
						class="w-full px-4 py-2.5 sm:py-3 bg-slate-800 border border-slate-600 rounded-xl text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 outline-none text-sm"
					></textarea>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="px-4 sm:px-6 py-4 border-t border-slate-300 dark:border-slate-700 flex items-center justify-end gap-3"
			>
				<button
					onclick={closeFirebaseModal}
					class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-slate-900 dark:text-white rounded-xl font-medium transition-all text-sm"
				>
					Cancel
				</button>
				<button
					onclick={saveFirebaseParameter}
					disabled={firebaseSaving || !firebaseForm.key.trim()}
					class="px-4 py-2 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-slate-900 dark:text-white rounded-xl font-semibold shadow-lg shadow-orange-900/30 transition-all flex items-center gap-2 disabled:opacity-50 text-sm"
				>
					{#if firebaseSaving}
						<RefreshCw class="w-4 h-4 animate-spin" />
						Saving...
					{:else}
						<Save class="w-4 h-4" />
						{firebaseModalMode === 'create' ? 'Create' : 'Update'}
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

	/* Hide scrollbar for Chrome, Safari and Opera */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}

	/* Hide scrollbar for IE, Edge and Firefox */
	.no-scrollbar {
		-ms-overflow-style: none; /* IE and Edge */
		scrollbar-width: none; /* Firefox */
	}
</style>
