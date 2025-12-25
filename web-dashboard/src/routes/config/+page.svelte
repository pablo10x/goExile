<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import {
		notifications,
		backgroundConfig,
		siteSettings,
		theme
	} from '$lib/stores';
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
		Edit3,
		Palette,
		Monitor,
		Code2,
		LayoutDashboard,
		CloudRain,
		Waves,
		Wind,
		Activity
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
	let activeTab = $state<'master' | 'spawner' | 'firebase' | 'aesthetic'>('master');
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
			gradient: 'from-rust to-rust-light',
			expanded: true,
			items: []
		},
		{
			id: 'network',
			title: 'Network & Connectivity',
			description: 'Ports, hosts, and connection settings',
			icon: Network,
			color: 'cyan',
			gradient: 'from-orange-600 to-amber-600',
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
			gradient: 'from-amber-700 to-orange-800',
			expanded: false,
			items: []
		},
		{
			id: 'performance',
			title: 'Performance & Limits',
			description: 'Resource limits and performance tuning',
			icon: Zap,
			color: 'amber',
			gradient: 'from-rust to-rust-light',
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
			gradient: 'from-stone-700 to-stone-800',
			expanded: true,
			items: []
		},
		{
			id: 'limits',
			title: 'Resource Limits',
			description: 'Instance limits and resource allocation',
			icon: HardDrive,
			color: 'orange',
			gradient: 'from-rust to-orange-700',
			expanded: true,
			items: []
		},
		{
			id: 'ports',
			title: 'Port Configuration',
			description: 'Game server port ranges',
			icon: Network,
			color: 'teal',
			gradient: 'from-rust-light to-rust',
			expanded: false,
			items: []
		},
		{
			id: 'updates',
			title: 'Auto-Update Settings',
			description: 'Automatic update behavior',
			icon: RefreshCw,
			color: 'indigo',
			gradient: 'from-stone-600 to-rust',
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
			if (!response.ok) throw new Error('Failed to fetch config');
			const configs = await response.json();
			distributeConfigs(configs);
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

	function loadFirebaseStatus() {
		return fetch('/api/config/firebase/status').then(response => {
			if (response.ok) {
				return response.json();
			}
		}).then((status) => {
			firebaseConnected = status.connected;
			firebaseProjectId = status.project_id || '';
			if (status.configs) {
				firebaseConfigs = status.configs;
			}
		}).catch(() => {
			firebaseConnected = false;
		});
	}

	// Firebase functions
	function openFirebaseModal(mode: 'create' | 'edit', config?: FirebaseConfig) {
		firebaseModalMode = mode;
		if (mode === 'edit' && config) {
			firebaseEditingKey = config.key;
			firebaseForm = {
				key: config.key,
				value: config.value,
				valueType: config.valueType.toUpperCase() as any,
				description: config.description || ''
			};
		} else {
			firebaseEditingKey = '';
			firebaseForm = { key: '', value: '', valueType: 'STRING' as any, description: '' };
		}
		showFirebaseModal = true;
	}

	function closeFirebaseModal() {
		showFirebaseModal = false;
		firebaseForm = { key: '', value: '', valueType: 'STRING' as any, description: '' };
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
		
		try {
			const promises = [];
			for (const [key, value] of pendingChanges.entries()) {
				promises.push(
					fetch(`/api/config/${key}`, {
						method: 'PUT',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ value })
					})
				);
			}

			const results = await Promise.all(promises);
			const failed = results.filter(r => !r.ok);

			if (failed.length > 0) {
				throw new Error(`Failed to save ${failed.length} items`);
			}

			notifications.add({ type: 'success', message: 'Configuration saved successfully' });
			pendingChanges = new Map();
			await loadConfig(); // Reload to refresh timestamps etc
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Save failed',
				details: e.message
			});
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

	onMount(async () => {
		await loadConfig();
		await loadFirebaseStatus();
	});
</script>

<div class="relative z-10 p-4 sm:p-6 max-w-7xl mx-auto pb-24 md:pb-6">
	<!-- Header -->
	<div class="mb-10">
		<div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
			<div class="flex items-center gap-5">
				<div
					class="p-3 bg-rust border-2 border-rust-light shadow-[0_0_20px_rgba(120,53,15,0.4)]"
				>
					<Settings class="w-8 h-8 text-white" />
				</div>
				<div>
					<div class="flex items-center gap-2 mb-1">
						<div class="h-px w-6 bg-rust"></div>
						<span class="tactical-code text-rust">System_Environment_Bus</span>
					</div>
					<h1 class="text-3xl sm:text-4xl font-black military-label text-white uppercase tracking-tighter">
						CONFIGURATION_CORE
					</h1>
				</div>
			</div>

			<!-- Actions Bar -->
			<div class="flex items-center gap-3">
				{#if hasUnsavedChanges}
					<div
						class="flex items-center gap-3 px-4 py-2 bg-rust/10 border-2 border-rust/30"
						transition:slide={{ axis: 'x' }}
					>
						<div class="w-2 h-2 bg-rust animate-pulse"></div>
						<span class="font-jetbrains text-[10px] font-black text-rust-light uppercase tracking-widest" >{pendingChangeCount}_UNCOMMITTED_CHANGES</span>
					</div>
					<button
						onclick={discardChanges}
						class="px-6 py-2 bg-stone-800 hover:bg-stone-700 text-stone-300 font-black text-[10px] uppercase tracking-widest transition-all border border-stone-700"
					>
						Rollback
					</button>
					<button
						onclick={saveChanges}
						disabled={saving}
						class="px-8 py-2 bg-rust hover:bg-rust-light text-white font-black text-[10px] uppercase tracking-widest shadow-[4px_4px_0px_rgba(120,53,15,0.3)] transition-all disabled:opacity-50"
					>
						{#if saving}
							COMMITTING...
						{:else}
							COMMIT_CHANGES
						{/if}
					</button>
				{:else}
					<button
						onclick={loadConfig}
						disabled={loading}
						class="px-6 py-2 bg-stone-900 hover:bg-white hover:text-black text-stone-400 font-black text-[10px] uppercase tracking-widest transition-all border border-stone-800 ml-auto md:ml-0"
					>
						<RefreshCw class="w-3 h-3 inline mr-2 {loading ? 'animate-spin' : ''}" />
						Recalibrate
					</button>
				{/if}
			</div>
		</div>

		<!-- Tab Navigation -->
		<div
			class="flex items-center p-1 bg-black/60 border-2 border-stone-800 backdrop-blur-md overflow-x-auto no-scrollbar"
		>
			<button
				onclick={() => (activeTab = 'master')}
				class="flex-1 flex flex-col items-center gap-1 px-6 py-3 transition-all {activeTab === 'master'
					? 'bg-rust text-white'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-black text-[11px] uppercase tracking-widest">MASTER_NODE</span>
				<span class="font-mono text-[7px] opacity-50">CORE_RESOURCES</span>
			</button>
			<button
				onclick={() => (activeTab = 'spawner')}
				class="flex-1 flex flex-col items-center gap-1 px-6 py-3 transition-all {activeTab === 'spawner'
					? 'bg-rust text-white'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-black text-[11px] uppercase tracking-widest">SPAWNER_BUS</span>
				<span class="font-mono text-[7px] opacity-50">REGISTRY_DEFAULTS</span>
			</button>
			<button
				onclick={() => (activeTab = 'firebase')}
				class="flex-1 flex flex-col items-center gap-1 px-6 py-3 transition-all {activeTab === 'firebase'
					? 'bg-orange-600 text-white'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-black text-[11px] uppercase tracking-widest">REMOTE_SIGNAL</span>
				<span class="font-mono text-[7px] opacity-50">FIREBASE_SYNC</span>
			</button>
			<button
				onclick={() => (activeTab = 'aesthetic')}
				class="flex-1 flex flex-col items-center gap-1 px-6 py-3 transition-all {activeTab === 'aesthetic'
					? 'bg-stone-100 text-black'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-black text-[11px] uppercase tracking-widest">AESTHETICS</span>
				<span class="font-mono text-[7px] opacity-50">INTERFACE_GEOMETRY</span>
			</button>
		</div>

		<!-- Search Bar -->
		<div class="mt-4 relative group">
			<Search
				class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 sm:w-5 sm:h-5 text-slate-500 group-focus-within:text-rust-light transition-colors"
			/>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Filter identifiers..."
				class="w-full pl-10 sm:pl-12 pr-10 py-2.5 sm:py-3 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl text-sm sm:text-base text-slate-200 placeholder:text-slate-600 focus:border-rust-light focus:ring-2 focus:ring-rust-light/20 outline-none transition-all font-mono"
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
					class="w-10 h-10 sm:w-12 sm:h-12 border-4 border-rust-light border-t-transparent rounded-full animate-spin"
				></div>
				<span class="text-slate-500 font-mono text-sm uppercase">Synchronizing...</span>
			</div>
		</div>
	{:else}
		<div class="space-y-6">
			{#if activeTab === 'master' || activeTab === 'spawner'}
				<div class="space-y-8" transition:fade={{ duration: 200 }}>
					{#each activeTab === 'master' ? filteredMasterSections : filteredSpawnerSections as section (section.id)}
						{@const isExpanded = expandedSections.has(section.id)}
						{@const SectionIcon = section.icon}
						<div
							class="brutalist-card {$siteSettings.aesthetic.industrial_styling ? 'rounded-none border-2' : 'rounded-2xl'} overflow-hidden shadow-2xl transition-all duration-300"
						>
							<button
								onclick={() => toggleSection(section.id)}
								class="w-full px-6 py-5 flex items-center justify-between hover:bg-white/5 transition-colors border-b border-stone-800"
							>
								<div class="flex items-center gap-5">
									<div
										class="p-2.5 bg-rust/20 border border-rust/40 rounded-none shadow-[0_0_15px_rgba(120,53,15,0.2)]"
									>
										<SectionIcon class="w-5 h-5 text-rust-light" />
									</div>
									<div class="text-left">
										<h3 class="text-lg font-black military-label text-white uppercase tracking-widest">
											{section.title}
										</h3>
										<p class="text-[9px] font-mono text-stone-500 uppercase tracking-widest mt-0.5">
											{section.description}
										</p>
									</div>
								</div>
								<div class="flex items-center gap-4">
									<span class="tactical-code text-stone-600 hidden sm:inline">{section.items.length}_IDENTIFIERS_LOADED</span>
									<ChevronDown class="w-5 h-5 text-stone-600 transition-transform duration-300 {isExpanded ? 'rotate-180 text-rust' : ''}" />
								</div>
							</button>

							{#if isExpanded}
								<div class="px-6 py-6 space-y-4 bg-black/20" transition:slide={{ duration: 300 }}>
									{#each section.items as item (item.key)}
										{@const isPending = pendingChanges.has(item.key)}
										{@const isSecret = item.type === 'secret'}
										{@const showValue = showSecrets.has(item.key)}
										<div class="p-5 bg-stone-900/40 border-l-4 {isPending ? 'border-rust bg-rust/5' : 'border-stone-800'} transition-all group">
											<div class="flex flex-col sm:flex-row sm:items-start justify-between gap-6">
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-3 mb-2 flex-wrap">
														<span class="font-jetbrains text-xs font-black text-rust-light uppercase tracking-wider">{item.key}</span>
														<div class="flex gap-1">
															{#if item.is_read_only}
																<span class="text-[7px] font-black bg-stone-800 text-stone-500 px-2 py-0.5 border border-stone-700 uppercase">Protected</span>
															{/if}
															{#if item.requires_restart}
																<span class="text-[7px] font-black bg-amber-950/30 text-amber-500 px-2 py-0.5 border border-amber-900/30 uppercase">Restart_Req</span>
															{/if}
														</div>
													</div>
													<p class="text-[10px] font-mono text-stone-500 uppercase tracking-tight leading-relaxed max-w-2xl">{item.description}</p>
													
													<div class="mt-4">
														{#if item.type === 'bool'}
															<button 
																onclick={() => handleValueChange(item.key, (pendingChanges.get(item.key) ?? item.value) === 'true' ? 'false' : 'true', item.value)}
																disabled={item.is_read_only}
																class="flex items-center gap-3 px-4 py-2 rounded-none border-2 transition-all {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-rust/20 border-rust text-white' : 'bg-stone-950 border-stone-800 text-stone-600'}"
															>
																<div class="w-2 h-2 {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-rust shadow-[0_0_8px_var(--color-rust)]' : 'bg-stone-800'}"></div>
																<span class="font-black text-[10px] uppercase tracking-[0.2em]">{(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'Active_Status' : 'Standby_Protocol'}</span>
															</button>
														{:else}
															<div class="flex items-center gap-2 max-w-xl">
																<div class="relative flex-1">
																	<input
																		type={isSecret && !showValue ? 'password' : 'text'}
																		value={pendingChanges.get(item.key) ?? item.value}
																		oninput={e => handleValueChange(item.key, e.currentTarget.value, item.value)}
																		disabled={item.is_read_only}
																		class="w-full bg-black border-2 border-stone-800 focus:border-rust text-white font-mono text-xs px-4 py-2.5 transition-all disabled:opacity-30"
																	/>
																	{#if isSecret}
																		<button 
																			onclick={() => toggleSecret(item.key)} 
																			class="absolute right-3 top-1/2 -translate-y-1/2 text-stone-600 hover:text-rust transition-colors"
																		>
																			{#if showValue}<EyeOff class="w-4 h-4"/>{:else}<Eye class="w-4 h-4"/>{/if}
																		</button>
																	{/if}
																</div>
																<button 
																	onclick={() => copyToClipboard(pendingChanges.get(item.key) ?? item.value)} 
																	class="p-2.5 bg-stone-800 text-stone-500 hover:text-white hover:bg-rust transition-all border border-stone-700"
																	title="Copy_to_Buffer"
																>
																	<Copy class="w-4 h-4" />
																</button>
															</div>
														{/if}
													</div>
												</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			{#if activeTab === 'firebase'}
				<div class="space-y-6" transition:fade={{ duration: 200 }}>
					<!-- Connection Status Card -->
					<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl p-6 shadow-2xl">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-4">
								<div class="p-3 bg-gradient-to-br from-orange-600 to-amber-600 rounded-xl shadow-lg">
									<Flame class="w-6 h-6 text-white" />
								</div>
								<div>
									<h3 class="text-xl font-bold text-slate-100 font-heading tracking-widest uppercase">Firebase Remote Config</h3>
									<p class="text-xs text-slate-500 font-mono italic">Synchronize remote client parameters</p>
								</div>
							</div>
							{#if firebaseConnected}
								<div class="px-4 py-2 bg-emerald-500/10 border border-emerald-500/30 rounded-xl flex items-center gap-2">
									<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_10px_#10b981]"></div>
									<span class="text-emerald-400 font-black text-xs uppercase tracking-widest">Encrypted_Link</span>
								</div>
							{:else}
								<div class="px-4 py-2 bg-stone-800 border border-white/5 rounded-xl flex items-center gap-2">
									<div class="w-2 h-2 bg-stone-600 rounded-full"></div>
									<span class="text-stone-500 font-black text-xs uppercase tracking-widest">Link_Offline</span>
								</div>
							{/if}
						</div>

						{#if firebaseConnected}
							<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden shadow-2xl">
								<div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center justify-between bg-black/20">
									<div class="flex items-center gap-3">
										<FileJson class="w-5 h-5 text-orange-500" />
										<h3 class="text-lg font-bold text-slate-100 font-heading tracking-widest uppercase">Parameter Buffer</h3>
									</div>
									<div class="flex gap-2">
										<button onclick={syncFirebaseConfig} class="px-4 py-2 bg-stone-800 hover:bg-stone-700 text-white rounded-lg text-[10px] font-black uppercase tracking-widest transition-all">Sync_Core</button>
										<button onclick={() => openFirebaseModal('create')} class="px-4 py-2 bg-orange-600 hover:bg-orange-500 text-white rounded-lg text-[10px] font-black uppercase tracking-widest transition-all">Add_Node</button>
									</div>
								</div>
								<div class="p-6">
									<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
										{#each filteredFirebaseConfigs as config (config.key)}
											<div class="p-4 bg-black/40 rounded-xl border border-white/5 group hover:border-orange-500/30 transition-all">
												<div class="flex items-start justify-between gap-4">
													<div class="min-w-0">
														<div class="flex items-center gap-2 mb-1">
															<span class="text-xs font-bold text-orange-500 font-mono uppercase">{config.key}</span>
														<span class="text-[8px] bg-stone-800 text-stone-500 px-1 py-0.5 rounded">{config.valueType}</span>
													</div>
													<p class="text-[10px] text-slate-500 italic mb-2 truncate uppercase tracking-tight">{config.description || 'NO_META_DATA'}</p>
													<code class="text-[10px] text-stone-400 bg-stone-900/50 px-2 py-1 rounded block truncate font-mono">
														{config.value}
												</code>
													</div>
												<div class="flex gap-1">
													<button onclick={() => openFirebaseModal('edit', config)} class="p-1.5 text-stone-600 hover:text-white transition-colors"><Edit3 class="w-3.5 h-3.5"/></button>
													<button onclick={() => deleteFirebaseParameter(config.key)} class="p-1.5 text-stone-600 hover:text-red-500 transition-colors"><Trash2 class="w-3.5 h-3.5"/></button>
												</div>
											</div>
										</div>
									{/each}
									</div>
																</div>
															</div>
														{/if}
													</div>
												</div>
											{/if}
			{#if activeTab === 'aesthetic'}
				<div class="space-y-6" transition:fade={{ duration: 200 }}>
					<!-- System Aesthetic Section -->
					<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden shadow-2xl">
						<div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center gap-3 bg-black/20">
							<div class="p-2 bg-gradient-to-br from-rust to-rust-light rounded-lg shadow-lg">
								<Monitor class="w-5 h-5 text-white" />
							</div>
							<h3 class="text-xl font-bold text-slate-100 font-heading tracking-widest uppercase">System Interface Aesthetic</h3>
						</div>
					<div class="p-6">
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each [
									{ key: 'crt_effect', label: 'CRT Simulation', desc: 'Cathode-ray visualization', icon: Monitor },
									{ key: 'industrial_styling', label: 'Industrial Geometry', desc: 'Sharp edges & heavy borders', icon: Shield },
									{ key: 'glassmorphism', label: 'Refractive Glass', desc: 'Backdrop occlusion effects', icon: Cloud },
									{ key: 'glow_effects', label: 'Luminous Core', desc: 'Signal & shadow radiance', icon: Zap },
									{ key: 'animations_enabled', label: 'Aggressive Overlays', desc: 'Fluid state synchronization', icon: RefreshCw },
									{ key: 'panic_mode', label: 'Red Alert Protocol', desc: 'Critical state visualization', icon: AlertCircle }
							] as toggle}
								<div class="flex items-center justify-between p-4 bg-black/40 rounded-xl border border-white/5 hover:border-rust-light/30 transition-all">
									<div class="flex items-center gap-3">
										<toggle.icon class="w-4 h-4 text-rust-light" />
										<div>
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">{toggle.label}</h4>
											<p class="text-[9px] text-slate-500 font-mono uppercase">{toggle.desc}</p>
										</div>
									</div>
									<label class="relative inline-flex items-center cursor-pointer">
										<input type="checkbox" checked={($siteSettings.aesthetic as any)[toggle.key]} onchange={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [toggle.key]: e.currentTarget.checked } }))} class="sr-only peer">
										<div class="w-10 h-5 bg-stone-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-rust"></div>
									</label>
								</div>
							{/each}
							</div>

							<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8 pt-8 border-t border-white/5">
								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Wind class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Scanline Intensity</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic.scanlines_opacity || 0) * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="0.2" step="0.01" value={$siteSettings.aesthetic.scanlines_opacity} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, scanlines_opacity: parseFloat(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Activity class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Static Noise</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic.noise_opacity || 0) * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="0.1" step="0.005" value={$siteSettings.aesthetic.noise_opacity} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, noise_opacity: parseFloat(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Palette class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Card Opacity</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic.card_alpha || 0) * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="1" step="0.05" value={$siteSettings.aesthetic.card_alpha} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, card_alpha: parseFloat(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Cloud class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Backdrop Blur</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.backdrop_blur || 0}px</span>
									</div>
									<input type="range" min="0" max="40" step="1" value={$siteSettings.aesthetic.backdrop_blur} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, backdrop_blur: parseInt(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Shield class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Border Weight</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.card_border_width || 0}px</span>
									</div>
									<input type="range" min="0" max="10" step="1" value={$siteSettings.aesthetic.card_border_width} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, card_border_width: parseInt(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Zap class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Shadow Magnitude</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.card_shadow_size || 0}px</span>
									</div>
									<input type="range" min="0" max="50" step="1" value={$siteSettings.aesthetic.card_shadow_size} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, card_shadow_size: parseInt(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<LayoutDashboard class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Sidebar Opacity</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic.sidebar_alpha || 0) * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="1" step="0.05" value={$siteSettings.aesthetic.sidebar_alpha} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, sidebar_alpha: parseFloat(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Monitor class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Global BG Opacity</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic.bg_opacity || 0) * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="1" step="0.05" value={$siteSettings.aesthetic.bg_opacity} oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, bg_opacity: parseFloat(e.currentTarget.value) } }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5 md:col-span-2">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Code2 class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">System Typeface</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.font_primary || 'Inter'}</span>
									</div>
									<div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
										{#each ['Inter', 'Space Grotesk', 'Michroma', 'Orbitron', 'Red Hat Mono', 'Syncopate', 'Kanit', 'JetBrains Mono'] as font}
											<button 
												onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_primary: font } }))}
												class="px-3 py-2 bg-stone-900 border transition-all text-xs {$siteSettings.aesthetic.font_primary === font ? 'border-rust text-rust-light bg-rust/5' : 'border-stone-800 text-stone-500 hover:border-stone-700'}"
												style="font-family: '{font}', sans-serif;"
											>
												{font}
											</button>
										{/each}
									</div>
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5 md:col-span-2">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Monitor class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Global Base Color</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.bg_color}</span>
									</div>
									<div class="flex gap-4 items-center">
										<div class="relative group">
											<input 
												type="color" 
												value={$siteSettings.aesthetic.bg_color || '#050505'} 
												oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, bg_color: e.currentTarget.value } }))} 
												class="w-14 h-14 bg-transparent border-none cursor-pointer appearance-none" 
											/>
											<div class="absolute inset-0 border-2 border-stone-800 pointer-events-none group-hover:border-rust transition-colors"></div>
										</div>
										<div class="grid grid-cols-5 gap-2 flex-1">
											{#each ['#050505', '#0a0a0a', '#121212', '#1c1917', '#0c0a09'] as color}
												<button 
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, bg_color: color } }))}
													class="h-10 border transition-transform hover:scale-105 {$siteSettings.aesthetic.bg_color === color ? 'border-rust shadow-[0_0_10px_rgba(120,53,15,0.3)]' : 'border-white/10'}"
													style="background-color: {color}"
													aria-label="Set background color to {color}"
												></button>
											{/each}
										</div>
									</div>
								</div>

								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5 md:col-span-2">
									<div class="flex justify-between items-center mb-1">
										<div class="flex items-center gap-2">
											<Palette class="w-3 h-3 text-rust-light" />
											<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Primary System Accent</h4>
										</div>
										<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.accent_color}</span>
									</div>
									<div class="flex gap-4 items-center">
										<div class="relative group">
											<input 
												type="color" 
												value={$siteSettings.aesthetic.accent_color} 
												oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, accent_color: e.currentTarget.value } }))} 
												class="w-14 h-14 bg-transparent border-none cursor-pointer appearance-none" 
											/>
											<div class="absolute inset-0 border-2 border-stone-800 pointer-events-none group-hover:border-rust transition-colors"></div>
										</div>
										<div class="grid grid-cols-5 gap-2 flex-1">
											{#each ['#78350f', '#92400e', '#ef4444', '#10b981', '#0ea5e9'] as color}
												<button 
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, accent_color: color } }))}
													class="h-10 border transition-transform hover:scale-105 {$siteSettings.aesthetic.accent_color === color ? 'border-rust shadow-[0_0_10px_rgba(120,53,15,0.3)]' : 'border-white/10'}"
													style="background-color: {color}"
													aria-label="Set accent color to {color}"
												></button>
											{/each}
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Atmospheric Modulation Section -->
					<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden shadow-2xl">
						<div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center gap-3 bg-black/20">
							<div class="p-2 bg-gradient-to-br from-rust to-rust-light rounded-lg shadow-lg">
								<CloudRain class="w-5 h-5 text-white" />
							</div>
							<h3 class="text-xl font-bold text-slate-100 font-heading tracking-widest uppercase">Atmospheric Modulation</h3>
						</div>
					<div class="p-6">
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each [
									{ key: 'show_smoke', label: 'Thermal Exhaust', icon: Wind },
									{ key: 'show_rain', label: 'Acid Rain', icon: CloudRain },
									{ key: 'show_clouds', label: 'Vapor Cover', icon: Cloud },
									{ key: 'show_vignette', label: 'Vignette', icon: Monitor },
									{ key: 'show_navbar_particles', label: 'Ash Fall', icon: Waves }
							] as effect}
								<button 
									onclick={() => backgroundConfig.update((b: any) => ({ ...b, [effect.key]: !b[effect.key] }))}
									class="flex items-center justify-between p-4 rounded-xl border transition-all {($backgroundConfig as any)[effect.key] ? 'bg-rust/10 border-rust/40 text-rust-light' : 'bg-black/40 border-white/5 text-slate-500 hover:border-white/10'}"
								>
									<div class="flex items-center gap-3">
										<effect.icon class="w-4 h-4" />
										<span class="font-bold uppercase text-[10px] tracking-widest">{effect.label}</span>
									</div>
									<div class="w-1.5 h-1.5 rounded-full {($backgroundConfig as any)[effect.key] ? 'bg-rust shadow-[0_0_8px_var(--color-rust)]' : 'bg-stone-800'}"></div>
								</button>
							{/each}
							</div>

							<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8 pt-8 border-t border-white/5">
								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center">
										<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Precipitation Density</h4>
										<span class="text-[10px] font-mono text-rust-light">{($backgroundConfig.rain_opacity * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="1" step="0.05" value={$backgroundConfig.rain_opacity} oninput={e => backgroundConfig.update(b => ({ ...b, rain_opacity: parseFloat(e.currentTarget.value) }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>
								<div class="space-y-3 p-4 bg-black/40 rounded-xl border border-white/5">
									<div class="flex justify-between items-center">
										<h4 class="font-bold text-slate-200 uppercase text-[10px] tracking-widest">Vapor Opacity</h4>
										<span class="text-[10px] font-mono text-rust-light">{($backgroundConfig.clouds_opacity * 100).toFixed(0)}%</span>
									</div>
									<input type="range" min="0" max="1" step="0.05" value={$backgroundConfig.clouds_opacity} oninput={e => backgroundConfig.update(b => ({ ...b, clouds_opacity: parseFloat(e.currentTarget.value) }))} class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" />
								</div>
							</div>
						</div>
					</div>

					<!-- Visual Core Engine Section -->
					<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden shadow-2xl">
						<div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center gap-3 bg-black/20">
							<div class="p-2 bg-gradient-to-br from-amber-600 to-orange-600 rounded-lg shadow-lg">
								<Zap class="w-5 h-5 text-white" />
							</div>
							<h3 class="text-xl font-bold text-slate-100 font-heading tracking-widest uppercase">Visual Core Engine</h3>
						</div>
					<div class="p-6">
						<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
							{#each [
									{ id: 'architecture', label: 'Brutalist', desc: 'Floating masses', icon: Server },
									{ id: 'tactical_grid', label: 'Radar', desc: 'Scanning nodes', icon: Network },
									{ id: 'neural_network', label: 'Neural', desc: 'Plexus mapping', icon: Activity },
									{ id: 'data_flow', label: 'Data Flow', desc: 'Signal stream', icon: FileJson },
									{ id: 'digital_horizon', label: 'Horizon', desc: 'Infinite grid', icon: Globe },
									{ id: 'none', label: 'Minimal', desc: 'Zero noise', icon: Zap }
							] as engine}
								<button 
									onclick={() => backgroundConfig.update(b => ({ ...b, global_type: engine.id as any }))}
									class="flex flex-col items-center gap-2 p-4 rounded-xl border transition-all {$backgroundConfig.global_type === engine.id ? 'bg-rust/20 border-rust shadow-[0_0_20px_rgba(120,53,15,0.3)] text-white' : 'bg-black/40 border-white/5 text-slate-500 hover:bg-slate-800/50'}"
								>
									<div class="flex items-center justify-between w-full">
										<engine.icon class="w-4 h-4 {$backgroundConfig.global_type === engine.id ? 'text-rust-light' : 'text-slate-600'}" />
										{#if $backgroundConfig.global_type === engine.id}
											<div class="w-1.5 h-1.5 bg-rust rounded-full shadow-[0_0_8px_var(--color-rust)] animate-pulse"></div>
										{/if}
									</div>
									<span class="font-bold uppercase text-[10px] tracking-widest"> {engine.label} </span>
									<span class="text-[8px] font-mono opacity-40 uppercase leading-tight">{engine.desc}</span>
								</button>
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}
		</div>
	{/if}
</div>


<!-- Firebase Parameter Modal -->
{#if showFirebaseModal}
	<div
		class="fixed inset-0 z-[500] flex items-center justify-center p-4 bg-black/90 backdrop-blur-sm"
		onclick={(e) => e.target === e.currentTarget && closeFirebaseModal()}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') closeFirebaseModal();
		}}
		role="button"
		tabindex="0"
		transition:fade={{ duration: 150 }}
	>
		<div
			class="w-full max-w-lg bg-black border border-white/10 rounded-xl shadow-2xl overflow-hidden"
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<!-- Header -->
			<div
				class="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-black"
			>
				<h3
					class="text-xl font-bold text-slate-100 font-heading tracking-widest uppercase flex items-center gap-3"
				>
					<Flame class="w-5 h-5 text-orange-500" />
					{firebaseModalMode === 'create' ? 'Append_Parameter' : 'Modify_Node'}
				</h3>
				<button
					onclick={closeFirebaseModal}
					class="p-2 text-stone-600 hover:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 space-y-6 max-h-[70vh] overflow-y-auto bg-[#050505] font-mono">
				<div>
					<label for="fbKey" class="block text-[10px] font-bold text-stone-500 uppercase tracking-widest mb-2">Parameter Identifier</label>
					<input
						id="fbKey"
						type="text"
						bind:value={firebaseForm.key}
						disabled={firebaseModalMode === 'edit'}
						placeholder="NODE_ID_001"
						class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all placeholder:text-stone-800"
					/>
				</div>

				<div>
					<label for="fbType" class="block text-[10px] font-bold text-stone-500 uppercase tracking-widest mb-2">Data Schema</label>
					<select
						id="fbType"
						bind:value={firebaseForm.valueType}
						class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all appearance-none"
					>
						<option value="STRING">STRING</option>
						<option value="NUMBER">NUMBER</option>
						<option value="BOOLEAN">BOOLEAN</option>
						<option value="JSON">JSON</option>
					</select>
				</div>

				<div>
					<label for="fbValue" class="block text-[10px] font-bold text-stone-500 uppercase tracking-widest mb-2">Payload Data</label>
					{#if firebaseForm.valueType === 'BOOLEAN'}
						<select
							id="fbValue"
							bind:value={firebaseForm.value}
							class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all"
						>
							<option value="true">TRUE</option>
							<option value="false">FALSE</option>
						</select>
					{:else if firebaseForm.valueType === 'JSON'}
																		<textarea
																			id="fbValue"
																			bind:value={firebaseForm.value}
																			rows={4}
																			placeholder={`{ "status": "active" }`}
																			class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all resize-none"
																		></textarea>					{:else}
						<input
							id="fbValue"
							type={firebaseForm.valueType === 'NUMBER' ? 'number' : 'text'}
							bind:value={firebaseForm.value}
							placeholder="NULL_PTR"
							class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all"
						/>
					{/if}
				</div>

				<div>
					<label for="fbDesc" class="block text-[10px] font-bold text-stone-500 uppercase tracking-widest mb-2">Signal Description</label>
										<textarea
											id="fbDesc"
											bind:value={firebaseForm.description}
											rows={2}
											placeholder="Protocol purpose..."
											class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all resize-none"
										></textarea>				</div>
			</div>

			<!-- Footer -->
			<div class="px-8 py-6 bg-black border-t border-white/5 flex items-center justify-end gap-4">
				<button onclick={closeFirebaseModal} class="px-6 py-2 text-[11px] font-black text-stone-600 hover:text-white uppercase tracking-widest italic transition-all">
					Abort
				</button>
				<button
					onclick={saveFirebaseParameter}
					disabled={firebaseSaving || !firebaseForm.key.trim()}
					class="px-8 py-3 bg-orange-600 hover:bg-white text-black text-[11px] font-black uppercase tracking-[0.3em] transition-all disabled:opacity-20 shadow-[6px_6px_0px_#000]"
				>
					{#if firebaseSaving}
						SYNCING...
					{:else}
						COMMIT_PROTOCOL
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Hide scrollbar for Chrome, Safari and Opera */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}

	/* Hide scrollbar for IE, Edge and Firefox */
	.no-scrollbar {
		-ms-overflow-style: none; /* IE and Edge */
		scrollbar-width: none; /* Firefox */
	}

	input[type="range"] {
		@apply transition-all duration-300;
	}
	
	input[type="range"]:hover {
		@apply brightness-110;
	}
</style>