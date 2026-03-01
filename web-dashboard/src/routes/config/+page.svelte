<script lang="ts">
import { apiFetch } from "$lib/api";
	import { onMount } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import {
		notifications,
		siteSettings,
		backgroundConfig,
		theme
	} from '$lib/stores.svelte';
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
    import PageHeader from '$lib/components/theme/PageHeader.svelte';
    import Card from '$lib/components/theme/Card.svelte';

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
	}

	interface ConfigSection {
		id: string;
		title: string;
		description: string;
		icon: any;
		items: ConfigItem[];
	}

	interface FirebaseConfig {
		key: string;
		value: string;
		valueType: 'string' | 'number' | 'boolean' | 'json';
		description: string;
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

	let masterSections = $state<ConfigSection[]>([
		{ id: 'general', title: 'General Settings', description: 'Core server identification and behavior', icon: SettingsIcon, items: [] },
		{ id: 'network', title: 'Network & Connectivity', description: 'Ports, hosts, and connection settings', icon: Network, items: [] },
		{ id: 'security', title: 'Security & Authentication', description: 'API keys, tokens, and access control', icon: Shield, items: [] },
		{ id: 'database', title: 'Database Configuration', description: 'Database connections and pooling', icon: Database, items: [] },
		{ id: 'performance', title: 'Performance & Limits', description: 'Resource limits and performance tuning', icon: Zap, items: [] }
	]);

	let nodeSections = $state<ConfigSection[]>([
		{ id: 'defaults', title: 'Templates', description: 'Default values for new infrastructure', icon: Cpu, items: [] },
		{ id: 'limits', title: 'Resource Allocation', description: 'Instance limits and resource bounds', icon: HardDrive, items: [] },
		{ id: 'ports', title: 'Port Ranges', description: 'Game server network port pools', icon: Network, items: [] },
		{ id: 'updates', title: 'Software Updates', description: 'Automatic binary update behavior', icon: RefreshCw, items: [] }
	]);

	let firebaseConfigs = $state<FirebaseConfig[]>([]);
	let firebaseConnected = $state(false);
	let firebaseProjectId = $state('');

	let pendingChangeCount = $derived(pendingChanges.size);
	let hasUnsavedChanges = $derived(pendingChangeCount > 0);

	let filteredMasterSections = $derived.by(() => {
		if (!searchQuery.trim()) return masterSections;
		const query = searchQuery.toLowerCase();
		return masterSections
			.map(s => ({ ...s, items: s.items.filter(i => i.key.toLowerCase().includes(query) || i.description.toLowerCase().includes(query) || i.value.toLowerCase().includes(query)) }))
			.filter(s => s.items.length > 0);
	});

	let filteredNodeSections = $derived.by(() => {
		if (!searchQuery.trim()) return nodeSections;
		const query = searchQuery.toLowerCase();
		return nodeSections
			.map(s => ({ ...s, items: s.items.filter(i => i.key.toLowerCase().includes(query) || i.description.toLowerCase().includes(query) || i.value.toLowerCase().includes(query)) }))
			.filter(s => s.items.length > 0);
	});

	let filteredFirebaseConfigs = $derived.by(() => {
		if (!searchQuery.trim()) return firebaseConfigs;
		const query = searchQuery.toLowerCase();
		return firebaseConfigs.filter(c => c.key.toLowerCase().includes(query) || c.description.toLowerCase().includes(query));
	});

	async function loadConfig() {
		loading = true;
		try {
			const res = await apiFetch('/api/config');
			if (!res.ok) throw new Error('Failed to fetch config');
			distributeConfigs(await res.json());
		} catch (e: any) { notifications.add({ type: 'error', message: 'Config Error', details: e.message }); }
		finally { loading = false; }
	}

	function distributeConfigs(configs: ConfigItem[]) {
		masterSections = masterSections.map(s => ({ ...s, items: [] }));
		nodeSections = nodeSections.map(s => ({ ...s, items: [] }));
		for (const c of configs) {
			if (c.category === 'system') {
				if (c.key.includes('port') || c.key.includes('host') || c.key.includes('url')) masterSections.find(s => s.id === 'network')?.items.push(c);
				else if (c.key.includes('key') || c.key.includes('secret') || c.key.includes('auth') || c.key.includes('token')) masterSections.find(s => s.id === 'security')?.items.push(c);
				else if (c.key.includes('db') || c.key.includes('database') || c.key.includes('pool')) masterSections.find(s => s.id === 'database')?.items.push(c);
				else if (c.key.includes('max') || c.key.includes('limit') || c.key.includes('timeout') || c.key.includes('ttl')) masterSections.find(s => s.id === 'performance')?.items.push(c);
				else masterSections.find(s => s.id === 'general')?.items.push(c);
			} else if (c.category === 'node') {
				if (c.key.includes('port')) nodeSections.find(s => s.id === 'ports')?.items.push(c);
				else if (c.key.includes('max') || c.key.includes('limit') || c.key.includes('memory') || c.key.includes('cpu')) nodeSections.find(s => s.id === 'limits')?.items.push(c);
				else if (c.key.includes('update') || c.key.includes('auto')) nodeSections.find(s => s.id === 'updates')?.items.push(c);
				else nodeSections.find(s => s.id === 'defaults')?.items.push(c);
			}
		}
	}

	async function loadFirebaseStatus() {
		try {
			const res = await apiFetch('/api/config/firebase/status');
			if (res.ok) {
				const status = await res.json();
				firebaseConnected = status.connected;
				firebaseProjectId = status.project_id || '';
				if (status.configs) firebaseConfigs = status.configs;
			}
		} catch { firebaseConnected = false; }
	}

	function openFirebaseModal(mode: 'create' | 'edit', config?: FirebaseConfig) {
		firebaseModalMode = mode;
		if (mode === 'edit' && config) {
			firebaseEditingKey = config.key;
			firebaseForm = { key: config.key, value: config.value, valueType: config.valueType.toUpperCase() as any, description: config.description || '' };
		} else {
			firebaseEditingKey = '';
			firebaseForm = { key: '', value: '', valueType: 'STRING' as any, description: '' };
		}
		showFirebaseModal = true;
	}

	async function saveFirebaseParameter() {
		firebaseSaving = true;
		try {
			const res = await apiFetch('/api/config/firebase/parameter', {
				method: firebaseModalMode === 'create' ? 'POST' : 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(firebaseForm)
			});
			if (!res.ok) throw new Error((await res.json()).error);
			notifications.add({ type: 'success', message: 'Parameter updated' });
			showFirebaseModal = false;
			await loadFirebaseStatus();
		} catch (e: any) { notifications.add({ type: 'error', message: 'Save failed', details: e.message }); }
		finally { firebaseSaving = false; }
	}

	async function deleteFirebaseParameter(key: string) {
		if (!confirm(`Delete parameter "${key}"?`)) return;
		try {
			const res = await apiFetch('/api/config/firebase/parameter', { method: 'DELETE', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key }) });
			if (res.ok) { notifications.add({ type: 'success', message: 'Deleted' }); await loadFirebaseStatus(); }
		} catch (e: any) { notifications.add({ type: 'error', message: 'Delete failed', details: e.message }); }
	}

	async function syncFirebaseConfig() {
		loading = true;
		try {
			const res = await apiFetch('/api/config/firebase/sync', { method: 'POST' });
			if (res.ok) { notifications.add({ type: 'success', message: 'Sync complete' }); await loadFirebaseStatus(); }
		} catch (e: any) { notifications.add({ type: 'error', message: 'Sync failed', details: e.message }); }
		finally { loading = false; }
	}

	function handleValueChange(key: string, value: string, original: string) {
		if (value !== original) pendingChanges.set(key, value);
		else pendingChanges.delete(key);
		pendingChanges = new Map(pendingChanges);
	}

	async function saveChanges() {
		saving = true;
		try {
			const ps = [];
			for (const [k, v] of pendingChanges.entries()) ps.push(apiFetch(`/api/config/${k}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ value: v }) }));
			const rs = await Promise.all(ps);
			if (rs.every(r => r.ok)) { notifications.add({ type: 'success', message: 'Saved' }); pendingChanges = new Map(); await loadConfig(); }
			else throw new Error('Some items failed to save');
		} catch (e: any) { notifications.add({ type: 'error', message: 'Error', details: e.message }); }
		finally { saving = false; }
	}

	onMount(() => { loadConfig(); loadFirebaseStatus(); });
</script>

<div class="space-y-10 font-sans">
	<PageHeader title="Settings" subtitle="System Configuration & Environment" icon="ph:sliders-bold">
		{#snippet actions()}
			<div class="flex items-center gap-4">
				{#if hasUnsavedChanges}
					<div class="flex items-center gap-3 px-4 py-2 bg-sky-500/10 border border-sky-500/20 rounded-xl" transition:fade>
						<div class="w-2 h-2 bg-sky-500 rounded-full animate-pulse shadow-lg"></div>
						<span class="text-xs font-bold text-sky-400 uppercase tracking-wider">{pendingChangeCount} Unsaved Changes</span>
					</div>
					<Button onclick={() => { pendingChanges = new Map(); loadConfig(); }} variant="secondary" size="md">Discard</Button>
					<Button onclick={saveChanges} loading={saving} variant="primary" size="md">Save All</Button>
				{:else}
					<Button onclick={loadConfig} loading={loading} variant="secondary" size="md" icon="ph:arrows-clockwise-bold">Reload</Button>
				{/if}
			</div>
		{/snippet}
	</PageHeader>

	<!-- Tabs -->
	<div class="flex bg-slate-900/50 p-1 rounded-2xl border border-white/5 backdrop-blur-md">
		{#each [
			{ id: 'master', label: 'Master Server', sub: 'Main Settings' },
			{ id: 'nodes', label: 'Infrastructure', sub: 'Templates' },
			{ id: 'firebase', label: 'Remote Config', sub: 'External Sync' }
		] as tab}
			<button onclick={() => activeTab = tab.id as any} class="flex-1 flex flex-col items-center py-3 rounded-xl transition-all {activeTab === tab.id ? 'bg-sky-500 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300 hover:bg-white/5'}">
				<span class="text-xs font-bold uppercase tracking-wider">{tab.label}</span>
				<span class="text-[9px] font-semibold opacity-60 uppercase">{tab.sub}</span>
			</button>
		{/each}
	</div>

	<!-- Search -->
	<div class="relative group">
		<Search class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-sky-400 transition-colors" />
		<input type="text" bind:value={searchQuery} placeholder="Search settings..." class="w-full pl-14 pr-6 py-4 bg-slate-900/50 border border-white/5 rounded-2xl text-white text-sm focus:border-sky-500/50 outline-none transition-all shadow-inner backdrop-blur-sm" />
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-20 gap-4" transition:fade>
			<div class="w-10 h-10 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"></div>
			<span class="text-slate-500 text-xs font-bold uppercase tracking-widest">Loading Settings...</span>
		</div>
	{:else}
		<div class="space-y-6">
			{#if activeTab === 'master' || activeTab === 'nodes'}
				<div class="space-y-8">
					{#each (activeTab === 'master' ? filteredMasterSections : filteredNodeSections) as section (section.id)}
						<Card title={section.title} subtitle={section.description} icon={section.icon} class="overflow-hidden">
							{#snippet actions()}
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest bg-white/5 px-3 py-1 rounded-lg">{section.items.length} Settings</span>
							{/snippet}
							<div class="p-6 space-y-4">
								{#each section.items as item (item.key)}
									{@const isPending = pendingChanges.has(item.key)}
									{@const isSecret = item.type === 'secret'}
									<div class="p-5 bg-slate-900/40 rounded-2xl border {isPending ? 'border-sky-500/30 bg-sky-500/5 shadow-inner' : 'border-white/5'} transition-all group">
										<div class="flex flex-col sm:flex-row justify-between gap-6">
											<div class="flex-1">
												<div class="flex items-center gap-3 mb-2 flex-wrap">
													<span class="text-xs font-bold text-sky-400 uppercase tracking-wide">{item.key}</span>
													{#if item.is_read_only}<span class="text-[9px] font-bold bg-slate-800 text-slate-500 px-2 py-0.5 rounded uppercase border border-white/5">Read Only</span>{/if}
													{#if item.requires_restart}<span class="text-[9px] font-bold bg-amber-500/10 text-amber-500 px-2 py-0.5 rounded uppercase border border-amber-500/20">Restart Required</span>{/if}
												</div>
												<p class="text-xs text-slate-500 leading-relaxed font-medium">{item.description}</p>
												
												<div class="mt-4">
													{#if item.type === 'bool'}
														<button onclick={() => handleValueChange(item.key, (pendingChanges.get(item.key) ?? item.value) === 'true' ? 'false' : 'true', item.value)} disabled={item.is_read_only} class="flex items-center gap-3 px-4 py-2 rounded-xl border transition-all {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : 'bg-slate-950 border-white/5 text-slate-500'}">
															<div class="w-2 h-2 rounded-full {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-emerald-500 shadow-lg' : 'bg-slate-800'}"></div>
															<span class="font-bold text-[10px] uppercase tracking-widest">{(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'Enabled' : 'Disabled'}</span>
														</button>
													{:else}
														<div class="flex items-center gap-2 max-w-xl">
															<div class="relative flex-1">
																<input type={isSecret && !showSecrets.has(item.key) ? 'password' : 'text'} value={pendingChanges.get(item.key) ?? item.value} oninput={e => handleValueChange(item.key, e.currentTarget.value, item.value)} disabled={item.is_read_only} class="w-full bg-slate-950 border border-white/5 rounded-xl text-white font-mono text-xs px-4 py-2.5 focus:border-sky-500/50 outline-none transition-all disabled:opacity-30" />
																{#if isSecret}<button onclick={() => { if(showSecrets.has(item.key)) showSecrets.delete(item.key); else showSecrets.add(item.key); showSecrets = new Set(showSecrets); }} class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-sky-400">{#if showSecrets.has(item.key)}<EyeOff size={16}/>{:else}<Eye size={16}/>{/if}</button>{/if}
															</div>
															<button onclick={() => { navigator.clipboard.writeText(pendingChanges.get(item.key) ?? item.value); notifications.add({ type: 'success', message: 'Copied' }); }} class="p-2.5 bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 rounded-xl border border-white/5 transition-all"><Copy size={16} /></button>
														</div>
													{/if}
												</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</Card>
					{/each}
				</div>
			{/if}

			{#if activeTab === 'firebase'}
				<div class="space-y-6">
					<Card title="Firebase Remote Config" subtitle="Synchronize remote client parameters" icon={Flame}>
						{#snippet actions()}
							{#if firebaseConnected}<div class="flex items-center gap-2 px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-emerald-400 text-[10px] font-bold uppercase"><div class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse"></div>Connected</div>
							{:else}<div class="flex items-center gap-2 px-3 py-1 bg-slate-800 border border-white/5 rounded-lg text-slate-500 text-[10px] font-bold uppercase"><div class="w-1.5 h-1.5 bg-slate-600 rounded-full"></div>Offline</div>{/if}
						{/snippet}
						<div class="p-6">
							{#if firebaseConnected}
								<div class="space-y-6">
									<div class="flex items-center justify-between border-b border-white/5 pb-4">
										<div class="flex items-center gap-3"><FileJson class="w-5 h-5 text-sky-400" /><h3 class="text-sm font-bold text-white uppercase tracking-wider">Remote Parameters</h3></div>
										<div class="flex gap-2"><Button onclick={syncFirebaseConfig} variant="secondary" size="sm">Sync All</Button><Button onclick={() => openFirebaseModal('create')} variant="primary" size="sm">Add Parameter</Button></div>
									</div>
									<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
										{#each filteredFirebaseConfigs as config (config.key)}
											<div class="p-5 bg-slate-900/40 rounded-2xl border border-white/5 hover:border-sky-500/30 transition-all group">
												<div class="flex items-start justify-between gap-4">
													<div class="min-w-0">
														<div class="flex items-center gap-2 mb-1"><span class="text-xs font-bold text-sky-400 font-mono uppercase">{config.key}</span><span class="text-[9px] bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded font-bold">{config.valueType}</span></div>
														<p class="text-[10px] text-slate-500 font-medium mb-3 truncate uppercase tracking-tight">{config.description || 'No description'}</p>
														<code class="text-[10px] text-slate-400 bg-slate-950 px-3 py-2 rounded-lg block truncate font-mono border border-white/5 shadow-inner">{config.value}</code>
													</div>
													<div class="flex gap-1"><Button onclick={() => openFirebaseModal('edit', config)} variant="ghost" size="xs" icon="ph:pencil-bold" /><Button onclick={() => deleteFirebaseParameter(config.key)} variant="ghost" size="xs" icon="ph:trash-bold" class="text-slate-500 hover:text-rose-500" /></div>
												</div>
											</div>
										{/each}
									</div>
								</div>
							{:else}
								<div class="text-center py-12 text-slate-500"><CloudCog class="w-12 h-12 mx-auto opacity-20 mb-4" /><p class="text-sm font-bold uppercase tracking-widest">Firebase integration not configured</p></div>
							{/if}
						</div>
					</Card>
				</div>
			{/if}
		</div>
	{/if}
</div>

{#if showFirebaseModal}
	<div class="fixed inset-0 z-[500] flex items-center justify-center p-4 bg-slate-950/90 backdrop-blur-md" onclick={e => e.target === e.currentTarget && (showFirebaseModal = false)} role="button" tabindex="0" onkeydown={null} transition:fade>
		<div class="w-full max-w-lg bg-slate-900 border border-white/10 rounded-3xl shadow-2xl overflow-hidden" transition:scale>
			<div class="px-8 py-6 border-b border-white/5 flex items-center justify-between"><h3 class="text-xl font-bold text-white tracking-tight flex items-center gap-3"><Flame class="text-sky-400" />{firebaseModalMode === 'create' ? 'New Parameter' : 'Edit Parameter'}</h3><button onclick={() => showFirebaseModal = false} class="p-2 text-slate-500 hover:text-white transition-colors"><X /></button></div>
			<div class="p-8 space-y-6 max-h-[70vh] overflow-y-auto no-scrollbar">
				<div><label for="fbKey" class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Identifier</label><input id="fbKey" type="text" bind:value={firebaseForm.key} disabled={firebaseModalMode === 'edit'} placeholder="e.g. MAINTENANCE_MODE" class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm focus:border-sky-500/50 outline-none" /></div>
				<div><label for="fbType" class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Value Type</label><select id="fbType" bind:value={firebaseForm.valueType} class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm focus:border-sky-500/50 outline-none appearance-none"><option value="STRING">STRING</option><option value="NUMBER">NUMBER</option><option value="BOOLEAN">BOOLEAN</option><option value="JSON">JSON</option></select></div>
				<div><label for="fbValue" class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Value</label>{#if firebaseForm.valueType === 'BOOLEAN'}<select id="fbValue" bind:value={firebaseForm.value} class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm focus:border-sky-500/50 outline-none"><option value="true">TRUE</option><option value="false">FALSE</option></select>{:else if firebaseForm.valueType === 'JSON'}<textarea id="fbValue" bind:value={firebaseForm.value} rows={4} placeholder={`{ "active": true }`} class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm font-mono focus:border-sky-500/50 outline-none resize-none"></textarea>{:else}<input id="fbValue" type={firebaseForm.valueType === 'NUMBER' ? 'number' : 'text'} bind:value={firebaseForm.value} placeholder="Enter value..." class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm focus:border-sky-500/50 outline-none" />{/if}</div>
				<div><label for="fbDesc" class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Description</label><textarea id="fbDesc" bind:value={firebaseForm.description} rows={2} placeholder="Brief description of this setting..." class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-white text-sm focus:border-sky-500/50 outline-none resize-none"></textarea></div>
			</div>
			<div class="px-8 py-6 bg-slate-950 border-t border-white/5 flex items-center justify-end gap-4"><Button onclick={() => showFirebaseModal = false} variant="ghost" size="md">Cancel</Button><Button onclick={saveFirebaseParameter} disabled={firebaseSaving || !firebaseForm.key.trim()} variant="primary" size="md" loading={firebaseSaving}>Save Parameter</Button></div>
		</div>
	</div>
{/if}

<style>
	.no-scrollbar::-webkit-scrollbar { display: none; }
</style>