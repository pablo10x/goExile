<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import {
		notifications,
		siteSettings,
		backgroundConfig,
		theme
	} from '$lib/stores';
	import Button from '$lib/components/Button.svelte';
	import {
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
		Menu,
		Code2,
		LayoutDashboard,
		CloudRain,
		Waves,
		Wind,
		Activity,
		Settings as SettingsIcon
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
	let activeTab = $state<'master' | 'nodes' | 'firebase'>('master');
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
			icon: SettingsIcon,
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

	// Node Configuration Sections -> Node Configuration Sections
	let nodeSections = $state<ConfigSection[]>([
		{
			id: 'defaults',
			title: 'Default Settings',
			description: 'Default values for new nodes',
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

	let filteredNodeSections = $derived.by(() => {
		if (!searchQuery.trim()) return nodeSections;
		const query = searchQuery.toLowerCase();
		return nodeSections
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
		nodeSections = nodeSections.map((s) => ({ ...s, items: [] }));

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
			} else if (config.category === 'node') {
				if (config.key.includes('port')) {
					nodeSections.find((s) => s.id === 'ports')?.items.push(config);
				} else if (
					config.key.includes('max') ||
					config.key.includes('limit') ||
					config.key.includes('memory') ||
					config.key.includes('cpu')
				) {
					nodeSections.find((s) => s.id === 'limits')?.items.push(config);
				} else if (config.key.includes('update') || config.key.includes('auto')) {
					nodeSections.find((s) => s.id === 'updates')?.items.push(config);
				} else {
					nodeSections.find((s) => s.id === 'defaults')?.items.push(config);
				}
			}
		}
		masterSections = [...masterSections];
		nodeSections = [...nodeSections];
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

	onMount(async () => {
		await loadConfig();
		await loadFirebaseStatus();
	});
</script>

<div class="relative z-10 w-full space-y-10 pb-32 md:pb-12">
	<!-- Header -->
	<div class="space-y-10">
		<div class="flex flex-col xl:flex-row xl:items-center justify-between gap-8 mb-10 border-l-4 border-rust pl-6 sm:pl-10 py-2 bg-[var(--header-bg)]/60 backdrop-blur-xl industrial-frame">
			<div class="flex items-center gap-6">
				<div
					class="p-4 bg-rust/10 border border-rust/30 shadow-2xl industrial-frame"
				>
					<SettingsIcon class="w-10 h-10 text-rust-light" />
				</div>
				<div>
					<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter leading-none">
						CONFIGURATION
					</h1>
					<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-[0.3em] font-black mt-2">
						System Parameters & Environment Control
					</p>
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
						<span class="font-jetbrains text-[11px] font-black text-rust-light uppercase tracking-[0.2em]" >{pendingChangeCount} PENDING CHANGES</span>
					</div>
					<Button
						onclick={discardChanges}
						variant="secondary"
						size="md"
					>
						Discard
					</Button>
					<Button
						onclick={saveChanges}
						disabled={saving}
						variant="primary"
						size="md"
						loading={saving}
					>
						SAVE CHANGES
					</Button>
				{:else}
					<Button
						onclick={loadConfig}
						disabled={loading}
						variant="secondary"
						size="md"
						loading={loading}
						icon="ph:arrows-clockwise-bold"
					>
						Reload
					</Button>
				{/if}
			</div>
		</div>

		<!-- Tab Navigation -->
		<div
			class="flex items-center p-1.5 bg-[var(--header-bg)]/80 border border-stone-800 backdrop-blur-xl overflow-x-auto no-scrollbar industrial-frame shadow-2xl"
		>
			<button
				onclick={() => (activeTab = 'master')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'master'
					? 'bg-rust text-white shadow-lg'
					: 'text-text-dim hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">MASTER SERVER</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">CORE CONFIG</span>
			</button>
			<button
				onclick={() => (activeTab = 'nodes')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'nodes'
					? 'bg-rust text-white shadow-lg'
					: 'text-text-dim hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">Node_Fleet</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">DEFAULT SETTINGS</span>
			</button>
			<button
				onclick={() => (activeTab = 'firebase')}
				class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'firebase'
					? 'bg-rust text-white shadow-lg'
					: 'text-text-dim hover:text-white hover:bg-stone-900'}"
			>
				<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">REMOTE CONFIG</span>
				<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">FIREBASE SYNC</span>
			</button>
		</div>

		<!-- Search Bar -->
		<div class="relative group">
			<Search
				class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-text-dim group-focus-within:text-rust transition-colors"
			/>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="SEARCH PARAMETERS..."
				class="w-full pl-14 pr-10 py-4 bg-stone-950 border border-stone-800 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
			/>
			{#if searchQuery}
				<button
					onclick={() => (searchQuery = '')}
					class="absolute right-5 top-1/2 -translate-y-1/2 text-text-dim hover:text-white"
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
				<span class="text-text-dim font-mono text-sm uppercase">Loading...</span>
			</div>
		</div>
	{:else}
		<div class="space-y-6">
			{#if activeTab === 'master' || activeTab === 'nodes'}
				<div class="space-y-8" transition:fade={{ duration: 200 }}>
					{#each activeTab === 'master' ? filteredMasterSections : filteredNodeSections as section (section.id)}
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
										<p class="text-[9px] font-mono text-text-dim uppercase tracking-widest mt-0.5">
											{section.description}
										</p>
									</div>
								</div>
								<div class="flex items-center gap-4">
									<span class="tactical-code text-text-dim hidden sm:inline">{section.items.length} PARAMETERS</span>
									<ChevronDown class="w-5 h-5 text-text-dim transition-transform duration-300 {isExpanded ? 'rotate-180 text-rust' : ''}" />
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
																<span class="text-[7px] font-black bg-stone-800 text-text-dim px-2 py-0.5 border border-stone-700 uppercase">ReadOnly</span>
															{/if}
															{#if item.requires_restart}
																<span class="text-[7px] font-black bg-amber-950/30 text-warning px-2 py-0.5 border border-amber-900/30 uppercase">Restart Required</span>
															{/if}
														</div>
													</div>
													<p class="text-[10px] font-mono text-text-dim uppercase tracking-tight leading-relaxed max-w-2xl">{item.description}</p>
													
													<div class="mt-4">
														{#if item.type === 'bool'}
															<button 
																onclick={() => handleValueChange(item.key, (pendingChanges.get(item.key) ?? item.value) === 'true' ? 'false' : 'true', item.value)}
																disabled={item.is_read_only}
																class="flex items-center gap-3 px-4 py-2 rounded-none border-2 transition-all {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-rust/20 border-rust text-white' : 'bg-stone-950 border-stone-800 text-text-dim'}"
															>
																<div class="w-2 h-2 {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-rust shadow-[0_0_8px_var(--color-rust)]' : 'bg-stone-800'}"></div>
																<span class="font-black text-[10px] uppercase tracking-[0.2em]">{(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'ENABLED' : 'DISABLED'}</span>
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
																			class="absolute right-3 top-1/2 -translate-y-1/2 text-text-dim hover:text-rust transition-colors"
																		>
																			{#if showValue}<EyeOff class="w-4 h-4"/>{:else}<Eye class="w-4 h-4"/>{/if}
																		</button>
																	{/if}
																</div>
																<button 
																	onclick={() => copyToClipboard(pendingChanges.get(item.key) ?? item.value)} 
																	class="p-2.5 bg-stone-800 text-text-dim hover:text-white hover:bg-rust transition-all border border-stone-700"
																	title="Copy"
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
									<p class="text-xs text-text-dim font-mono italic">Synchronize remote client parameters</p>
								</div>
							</div>
							{#if firebaseConnected}
								<div class="px-4 py-2 bg-success/10 border border-emerald-500/30 rounded-xl flex items-center gap-2">
									<div class="w-2 h-2 bg-success rounded-full animate-pulse shadow-[0_0_10px_#10b981]"></div>
									<span class="text-success font-black text-xs uppercase tracking-widest">CONNECTED</span>
								</div>
							{:else}
								<div class="px-4 py-2 bg-stone-800 border border-white/5 rounded-xl flex items-center gap-2">
									<div class="w-2 h-2 bg-stone-600 rounded-full"></div>
									<span class="text-text-dim font-black text-xs uppercase tracking-widest">OFFLINE</span>
								</div>
							{/if}
						</div>

						{#if firebaseConnected}
							<div class="bg-[var(--card-bg)] backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden shadow-2xl mt-8">
								<div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center justify-between bg-black/20">
									<div class="flex items-center gap-3">
										<FileJson class="w-5 h-5 text-orange-500" />
										<h3 class="text-lg font-bold text-slate-100 font-heading tracking-widest uppercase">Parameter Buffer</h3>
									</div>
									<div class="flex gap-2">
										<button onclick={syncFirebaseConfig} class="px-4 py-2 bg-stone-800 hover:bg-stone-700 text-white rounded-lg text-[10px] font-black uppercase tracking-widest transition-all">Sync</button>
										<button onclick={() => openFirebaseModal('create')} class="px-4 py-2 bg-orange-600 hover:bg-orange-500 text-white rounded-lg text-[10px] font-black uppercase tracking-widest transition-all">Add Node</button>
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
															<span class="text-[8px] bg-stone-800 text-text-dim px-1 py-0.5 rounded">{config.valueType}</span>
														</div>
														<p class="text-[10px] text-text-dim italic mb-2 truncate uppercase tracking-tight">{config.description || 'NO_META_DATA'}</p>
														<code class="text-[10px] text-stone-400 bg-stone-900/50 px-2 py-1 rounded block truncate font-mono">
															{config.value}
														</code>
													</div>
													<div class="flex gap-1">
														<button onclick={() => openFirebaseModal('edit', config)} class="p-1.5 text-text-dim hover:text-white transition-colors"><Edit3 class="w-3.5 h-3.5"/></button>
														<button onclick={() => deleteFirebaseParameter(config.key)} class="p-1.5 text-text-dim hover:text-danger transition-colors"><Trash2 class="w-3.5 h-3.5"/></button>
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
					{firebaseModalMode === 'create' ? 'Add Parameter' : 'Modify Parameter'}
				</h3>
				<button
					onclick={closeFirebaseModal}
					class="p-2 text-text-dim hover:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 space-y-6 max-h-[70vh] overflow-y-auto bg-[var(--terminal-bg)] font-mono">
				<div>
					<label for="fbKey" class="block text-[10px] font-bold text-text-dim uppercase tracking-widest mb-2">Identifier</label>
					<input
						id="fbKey"
						type="text"
						bind:value={firebaseForm.key}
						disabled={firebaseModalMode === 'edit'}
						placeholder="PARAMETER_NAME"
						class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all placeholder:text-stone-800"
					/>
				</div>

				<div>
					<label for="fbType" class="block text-[10px] font-bold text-text-dim uppercase tracking-widest mb-2">Value Type</label>
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
					<label for="fbValue" class="block text-[10px] font-bold text-text-dim uppercase tracking-widest mb-2">Value</label>
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
							placeholder="Enter value..."
							class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all"
						/>
					{/if}
				</div>

				<div>
					<label for="fbDesc" class="block text-[10px] font-bold text-text-dim uppercase tracking-widest mb-2">Description</label>
										<textarea
											id="fbDesc"
											bind:value={firebaseForm.description}
											rows={2}
											placeholder="Purpose of this parameter..."
											class="w-full bg-black border border-white/5 px-4 py-3 text-slate-200 focus:border-orange-500 outline-none transition-all resize-none"
										></textarea>				</div>
			</div>

			<!-- Footer -->
			<div class="px-8 py-6 bg-black border-t border-white/5 flex items-center justify-end gap-4">
				<Button 
					onclick={closeFirebaseModal}
					variant="ghost"
					size="md"
				>
					Cancel
				</Button>
				<Button
					onclick={saveFirebaseParameter}
					disabled={firebaseSaving || !firebaseForm.key.trim()}
					variant="primary"
					size="md"
					loading={firebaseSaving}
				>
					SAVE PARAMETER
				</Button>
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
</style>