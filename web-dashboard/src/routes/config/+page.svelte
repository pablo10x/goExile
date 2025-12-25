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

<div class="relative z-10 w-full space-y-10 pb-32 md:pb-12">
	<!-- Header -->
	<div class="space-y-10">
		<div class="flex flex-col xl:flex-row xl:items-center justify-between gap-8 mb-10">
			<div class="flex items-center gap-6">
				<div
					class="p-4 bg-rust border-2 border-rust-light shadow-[0_0_30px_rgba(120,53,15,0.4)] industrial-frame"
				>
					<Settings class="w-10 h-10 text-white" />
				</div>
				<div>
					<div class="flex items-center gap-3 mb-1">
						<div class="h-0.5 w-8 bg-rust"></div>
						<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-[0.3em]">System_Environment_Bus</span>
					</div>
					<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter">
						CONFIGURATION_CORE
					</h1>
				</div>
			</div>

			<!-- Actions Bar -->
			<div class="flex items-center gap-4">
				{#if hasUnsavedChanges}
					<div
						class="flex items-center gap-4 px-5 py-3 bg-rust/10 border border-rust/30 industrial-frame"
						transition:slide={{ axis: 'x' }}
					>
						<div class="w-2.5 h-2.5 bg-rust animate-pulse shadow-rust/50 shadow-lg"></div>
						<span class="font-jetbrains text-[11px] font-black text-rust-light uppercase tracking-[0.2em]" >{pendingChangeCount}_UNCOMMITTED_CHANGES</span>
					</div>
					<button
						onclick={discardChanges}
						class="px-8 py-3 bg-stone-900 hover:bg-stone-800 text-stone-400 font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800"
					>
						Rollback
					</button>
					<button
						onclick={saveChanges}
						disabled={saving}
						class="px-10 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all disabled:opacity-20 active:translate-y-px"
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
						class="px-8 py-3 bg-stone-950 hover:bg-white hover:text-black text-stone-500 font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800 active:translate-y-px"
					>
						<RefreshCw class="w-4 h-4 inline mr-3 {loading ? 'animate-spin' : ''}" />
						Recalibrate_Registry
					</button>
				{/if}
			</div>
		</div>

		<!-- Tab Navigation -->
		<div
			class="flex items-center p-1.5 bg-[#0a0a0a]/80 border border-stone-800 backdrop-blur-xl overflow-x-auto no-scrollbar industrial-frame shadow-2xl"
		>
			<button
				onclick={() => (activeTab = 'master')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'master'
					? 'bg-rust text-white shadow-lg'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">MASTER_NODE</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">CORE_RESOURCES</span>
			</button>
			<button
				onclick={() => (activeTab = 'spawner')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'spawner'
					? 'bg-rust text-white shadow-lg'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">SPAWNER_BUS</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">REGISTRY_DEFAULTS</span>
			</button>
			<button
				onclick={() => (activeTab = 'firebase')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'firebase'
					? 'bg-orange-600 text-white shadow-lg'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">REMOTE_SIGNAL</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">FIREBASE_SYNC</span>
			</button>
			<button
				onclick={() => (activeTab = 'aesthetic')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'aesthetic'
					? 'bg-stone-100 text-black shadow-lg'
					: 'text-stone-600 hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">AESTHETICS</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">INTERFACE_GEOMETRY</span>
			</button>
		</div>

		<!-- Search Bar -->
		<div class="relative group">
			<Search
				class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-stone-600 group-focus-within:text-rust transition-colors"
			/>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="FILTER_SYSTEM_IDENTIFIERS..."
				class="w-full pl-14 pr-10 py-4 bg-stone-950 border border-stone-800 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
			/>
			{#if searchQuery}
				<button
					onclick={() => (searchQuery = '')}
					class="absolute right-5 top-1/2 -translate-y-1/2 text-stone-600 hover:text-white"
				>
					<X class="w-5 h-5" />
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
				<div class="space-y-8" transition:fade={{ duration: 200 }}>
					<!-- Theme Presets -->
					<div class="modern-industrial-card glass-panel p-8 !rounded-none shadow-2xl">
						<div class="flex items-center gap-4 mb-8 border-b border-stone-800 pb-4">
							<div class="p-2.5 bg-rust/10 border border-rust/30 industrial-frame">
								<Zap class="w-5 h-5 text-rust-light" />
							</div>
							<div>
								<h3 class="text-xl font-heading font-black text-white uppercase tracking-widest">Interface Presets</h3>
								<p class="text-[10px] font-jetbrains text-stone-500 uppercase tracking-widest mt-1">Select a calibrated system aesthetic</p>
							</div>
						</div>

						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
							{#each [
								{ 
									id: 'deep_command', 
									name: 'Deep Command', 
									desc: 'Classic dystopian industrial',
									config: { bg_color: '#050505', accent_color: '#f97316', card_alpha: 0.4, backdrop_blur: 16, industrial_styling: true, crt_effect: true, glow_effects: true }
								},
								{ 
									id: 'slate_minimal', 
									name: 'Slate Minimal', 
									desc: 'Clean tactical interface',
									config: { bg_color: '#0a0a0a', accent_color: '#0ea5e9', card_alpha: 0.6, backdrop_blur: 8, industrial_styling: false, crt_effect: false, glow_effects: false }
								},
								{ 
									id: 'toxic_hazard', 
									name: 'Toxic Hazard', 
									desc: 'High-visibility alert system',
									config: { bg_color: '#050505', accent_color: '#10b981', card_alpha: 0.3, backdrop_blur: 20, industrial_styling: true, crt_effect: true, glow_effects: true }
								},
								{ 
									id: 'red_alert', 
									name: 'Red Alert', 
									desc: 'Critical state visualization',
									config: { bg_color: '#050505', accent_color: '#ef4444', card_alpha: 0.5, backdrop_blur: 12, industrial_styling: true, crt_effect: true, glow_effects: true, panic_mode: true }
								}
							] as preset}
								<button 
									onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, ...preset.config } }))}
									class="flex flex-col p-5 bg-stone-900/40 border border-stone-800 hover:border-rust transition-all group industrial-frame text-left"
								>
									<div class="flex items-center justify-between mb-3">
										<span class="font-heading font-black text-xs text-white uppercase tracking-widest">{preset.name}</span>
										<div class="w-3 h-3 border border-white/20" style="background-color: {preset.config.accent_color}"></div>
									</div>
									<p class="text-[9px] font-jetbrains text-stone-500 uppercase tracking-tight leading-relaxed">{preset.desc}</p>
								</button>
							{/each}
						</div>
					</div>

					<div class="grid grid-cols-1 xl:grid-cols-2 gap-8">
						<!-- System Interface Aesthetic -->
						<div class="modern-industrial-card glass-panel !rounded-none shadow-2xl h-full">
							<div class="px-8 py-6 border-b border-stone-800 flex items-center gap-4 bg-black/20">
								<div class="p-2.5 bg-rust/10 border border-rust/30 industrial-frame">
									<Monitor class="w-5 h-5 text-rust-light" />
								</div>
								<h3 class="text-lg font-heading font-black text-white uppercase tracking-widest">Global Parameters</h3>
							</div>
							
							<div class="p-8 space-y-10">
								<!-- Toggles -->
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
									{#each [
											{ key: 'crt_effect', label: 'CRT Overlay', desc: 'Cathode-ray simulation', icon: Monitor },
											{ key: 'industrial_styling', label: 'Angular Geometry', desc: 'Sharp edges & frames', icon: Shield },
											{ key: 'glassmorphism', label: 'Refractive Glass', desc: 'Backdrop occlusion', icon: Cloud },
											{ key: 'glow_effects', label: 'Luminous Core', desc: 'Signal & shadow radiance', icon: Zap },
											{ key: 'animations_enabled', label: 'State Syncing', desc: 'Aggressive overlays', icon: RefreshCw },
											{ key: 'panic_mode', label: 'Red Alert', desc: 'Critical state logic', icon: AlertCircle }
									] as toggle}
										<div class="flex items-center justify-between p-4 bg-stone-950 border border-stone-800 hover:border-rust/30 transition-all industrial-frame">
											<div class="flex items-center gap-3">
												<toggle.icon class="w-4 h-4 text-rust" />
												<div>
													<h4 class="font-black text-stone-200 uppercase text-[10px] tracking-widest">{toggle.label}</h4>
													<p class="text-[8px] text-stone-600 font-jetbrains uppercase">{toggle.desc}</p>
												</div>
											</div>
											<label class="relative inline-flex items-center cursor-pointer">
												<input type="checkbox" checked={($siteSettings.aesthetic as any)[toggle.key]} onchange={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [toggle.key]: e.currentTarget.checked } }))} class="sr-only peer">
												<div class="w-10 h-5 bg-stone-800 border border-stone-700 peer-focus:outline-none rounded-none peer peer-checked:after:translate-x-5 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[3px] after:left-[3px] after:bg-stone-600 after:rounded-none after:h-3.5 after:w-3.5 after:transition-all peer-checked:bg-rust/20 peer-checked:border-rust"></div>
											</label>
										</div>
									{/each}
								</div>

								<!-- Sliders -->
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-8 pt-6 border-t border-stone-800">
									{#each [
										{ key: 'scanlines_opacity', label: 'Scanline Depth', icon: Wind, max: 0.3, step: 0.01, unit: '%' },
										{ key: 'noise_opacity', label: 'Signal Static', icon: Activity, max: 0.15, step: 0.005, unit: '%' },
										{ key: 'card_alpha', label: 'Panel Density', icon: Palette, max: 1, step: 0.05, unit: '%' },
										{ key: 'backdrop_blur', label: 'Focus Diffusion', icon: Cloud, max: 40, step: 1, unit: 'px' },
										{ key: 'card_border_width', label: 'Frame Weight', icon: Shield, max: 10, step: 1, unit: 'px' },
										{ key: 'card_shadow_size', label: 'Glow Magnitude', icon: Zap, max: 100, step: 1, unit: 'px' }
									] as slider}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<div class="flex items-center gap-3">
													<slider.icon class="w-3.5 h-3.5 text-rust" />
													<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest">{slider.label}</h4>
												</div>
												<span class="text-[10px] font-mono text-rust-light tabular-nums">
													{slider.unit === '%' ? (((($siteSettings.aesthetic as any)[slider.key]) || 0) * 100).toFixed(0) : ($siteSettings.aesthetic as any)[slider.key] || 0}{slider.unit}
												</span>
											</div>
											<div class="relative flex items-center h-2 bg-stone-950 border border-stone-800 shadow-inner">
												<input 
													type="range" 
													min="0" 
													max={slider.max} 
													step={slider.step} 
													value={($siteSettings.aesthetic as any)[slider.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30 pointer-events-none" style="width: {(($siteSettings.aesthetic as any)[slider.key] / slider.max) * 100}%"></div>
											</div>
										</div>
									{/each}
								</div>
							</div>
						</div>

						<!-- Atmospheric & Engine Section -->
						<div class="space-y-8 h-full">
							<!-- Atmosphere -->
							<div class="modern-industrial-card glass-panel !rounded-none shadow-2xl">
								<div class="px-8 py-6 border-b border-stone-800 flex items-center gap-4 bg-black/20">
									<div class="p-2.5 bg-rust/10 border border-rust/30 industrial-frame">
										<CloudRain class="w-5 h-5 text-rust-light" />
									</div>
									<h3 class="text-lg font-heading font-black text-white uppercase tracking-widest">Atmospheric Modulation</h3>
								</div>
								
								<div class="p-8 space-y-8">
									<div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
										{#each [
												{ key: 'show_smoke', label: 'Exhaust', icon: Wind },
												{ key: 'show_rain', label: 'Rain', icon: CloudRain },
												{ key: 'show_clouds', label: 'Cover', icon: Cloud },
												{ key: 'show_vignette', label: 'Vignette', icon: Monitor },
												{ key: 'show_navbar_particles', label: 'Ash Fall', icon: Waves }
										] as effect}
											<button 
												onclick={() => backgroundConfig.update((b: any) => ({ ...b, [effect.key]: !b[effect.key] }))}
												class="flex flex-col items-center gap-3 p-4 bg-stone-950 border transition-all industrial-frame {($backgroundConfig as any)[effect.key] ? 'border-rust text-rust-light shadow-rust/10' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
											>
												<effect.icon class="w-5 h-5" />
												<span class="font-black uppercase text-[9px] tracking-widest">{effect.label}</span>
											</button>
										{/each}
									</div>

									<div class="grid grid-cols-1 sm:grid-cols-2 gap-8 pt-6 border-t border-stone-800">
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest">Precipitation</h4>
												<span class="text-[10px] font-mono text-rust-light">{($backgroundConfig.rain_opacity * 100).toFixed(0)}%</span>
											</div>
											<input type="range" min="0" max="1" step="0.05" value={$backgroundConfig.rain_opacity} oninput={e => backgroundConfig.update(b => ({ ...b, rain_opacity: parseFloat(e.currentTarget.value) }))} class="w-full h-1 bg-stone-950 border border-stone-800 appearance-none cursor-pointer accent-rust" />
										</div>
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest">Vapor</h4>
												<span class="text-[10px] font-mono text-rust-light">{($backgroundConfig.clouds_opacity * 100).toFixed(0)}%</span>
											</div>
											<input type="range" min="0" max="1" step="0.05" value={$backgroundConfig.clouds_opacity} oninput={e => backgroundConfig.update(b => ({ ...b, clouds_opacity: parseFloat(e.currentTarget.value) }))} class="w-full h-1 bg-stone-950 border border-stone-800 appearance-none cursor-pointer accent-rust" />
										</div>
									</div>
								</div>
							</div>

							<!-- Engine -->
							<div class="modern-industrial-card glass-panel !rounded-none shadow-2xl">
								<div class="px-8 py-6 border-b border-stone-800 flex items-center gap-4 bg-black/20">
									<div class="p-2.5 bg-rust/10 border border-rust/30 industrial-frame">
										<Zap class="w-5 h-5 text-rust-light" />
									</div>
									<h3 class="text-lg font-heading font-black text-white uppercase tracking-widest">Visual Core Engine</h3>
								</div>
								
								<div class="p-8">
									<div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
										{#each [
												{ id: 'architecture', label: 'Brutalist', icon: Server },
												{ id: 'tactical_grid', label: 'Radar', icon: Network },
												{ id: 'neural_network', label: 'Neural', icon: Activity },
												{ id: 'data_flow', label: 'Data Flow', icon: FileJson },
												{ id: 'digital_horizon', label: 'Horizon', icon: Globe },
												{ id: 'none', label: 'Minimal', icon: Zap }
										] as engine}
											<button 
												onclick={() => backgroundConfig.update(b => ({ ...b, global_type: engine.id as any }))}
												class="flex flex-col items-center gap-3 p-5 bg-stone-950 border transition-all industrial-frame {$backgroundConfig.global_type === engine.id ? 'border-rust text-white shadow-rust/20' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
											>
												<engine.icon class="w-6 h-6 {$backgroundConfig.global_type === engine.id ? 'text-rust-light' : 'text-stone-700'}" />
												<span class="font-black uppercase text-[10px] tracking-widest">{engine.label}</span>
											</button>
										{/each}
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Advanced Typeface & Colors -->
					<div class="grid grid-cols-1 xl:grid-cols-3 gap-8">
						<div class="modern-industrial-card glass-panel p-8 !rounded-none shadow-2xl">
							<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest mb-6 border-b border-stone-800 pb-4">System Typeface</h4>
							<div class="grid grid-cols-1 gap-2">
								{#each ['Inter', 'Space Grotesk', 'Michroma', 'Orbitron', 'Red Hat Mono', 'Syncopate', 'Kanit', 'JetBrains Mono'] as font}
									<button 
										onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_primary: font } }))}
										class="px-4 py-3 bg-stone-950 border transition-all text-xs text-left industrial-frame {$siteSettings.aesthetic.font_primary === font ? 'border-rust text-rust-light bg-rust/5' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
										style="font-family: '{font}', sans-serif;"
									>
										{font}
									</button>
								{/each}
							</div>
						</div>

						<div class="modern-industrial-card glass-panel p-8 !rounded-none shadow-2xl xl:col-span-2">
							<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest mb-8 border-b border-stone-800 pb-4">Core Color Mapping</h4>
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-12">
								<div class="space-y-6">
									<div class="flex justify-between items-center">
										<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest">Base Color</span>
										<span class="text-[10px] font-mono text-stone-600">{$siteSettings.aesthetic.bg_color}</span>
									</div>
									<div class="flex gap-4 items-center">
										<div class="relative">
											<input 
												type="color" 
												value={$siteSettings.aesthetic.bg_color || '#050505'} 
												oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, bg_color: e.currentTarget.value } }))} 
												class="w-16 h-16 bg-transparent border-none cursor-pointer appearance-none" 
											/>
											<div class="absolute inset-0 border-2 border-stone-800 pointer-events-none group-hover:border-rust transition-colors industrial-frame"></div>
										</div>
										<div class="grid grid-cols-5 gap-2 flex-1">
											{#each ['#050505', '#0a0a0a', '#121212', '#1c1917', '#0c0a09'] as color}
												<button 
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, bg_color: color } }))}
													class="h-10 border transition-all hover:scale-105 industrial-frame {$siteSettings.aesthetic.bg_color === color ? 'border-rust shadow-[0_0_10px_rgba(120,53,15,0.3)]' : 'border-stone-800'}"
													style="background-color: {color}"
													aria-label="Set base color to {color}"
												></button>
											{/each}
										</div>
									</div>
								</div>

								<div class="space-y-6">
									<div class="flex justify-between items-center">
										<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest">Accent Vector</span>
										<span class="text-[10px] font-mono text-stone-600">{$siteSettings.aesthetic.accent_color}</span>
									</div>
									<div class="flex gap-4 items-center">
										<div class="relative">
											<input 
												type="color" 
												value={$siteSettings.aesthetic.accent_color} 
												oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, accent_color: e.currentTarget.value } }))} 
												class="w-16 h-16 bg-transparent border-none cursor-pointer appearance-none" 
											/>
											<div class="absolute inset-0 border-2 border-stone-800 pointer-events-none group-hover:border-rust transition-colors industrial-frame"></div>
										</div>
										<div class="grid grid-cols-5 gap-2 flex-1">
											{#each ['#f97316', '#92400e', '#ef4444', '#10b981', '#0ea5e9'] as color}
												<button 
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, accent_color: color } }))}
													class="h-10 border transition-all hover:scale-105 industrial-frame {$siteSettings.aesthetic.accent_color === color ? 'border-rust shadow-[0_0_10px_rgba(120,53,15,0.3)]' : 'border-stone-800'}"
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