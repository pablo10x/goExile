<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { config, restartRequired, notifications, siteSettings } from '$lib/stores.svelte';
	import type { ServerConfig } from '$lib/stores.svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { 
		ChevronLeft, 
		RefreshCw, 
		Save, 
		Lock, 
		AlertTriangle, 
		CheckCircle2, 
		Info,
		Settings as SettingsIcon,
		Shield,
		Cpu,
		Terminal,
		Search,
		X,
		Copy
	} from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let categoryConfigs = $state<ServerConfig[]>([]);
	let saving = $state(false);
	let searchQuery = $state('');
	let pendingChanges = $state<Map<string, string>>(new Map());

	let category = $derived(page.params.category || '');
	const categoryTitles: Record<string, string> = {
		system: 'System Configuration',
		node: 'Node Configuration',
		security: 'Security Configuration',
		aesthetic: 'Visual Calibration'
	};

	let filteredConfigs = $derived.by(() => {
		if (!searchQuery.trim()) return categoryConfigs;
		const q = searchQuery.toLowerCase();
		return categoryConfigs.filter(c => 
			c.key.toLowerCase().includes(q) || 
			c.description?.toLowerCase().includes(q) ||
			c.value.toLowerCase().includes(q)
		);
	});

	async function loadCategoryConfig() {
		if (!category) return;
		try {
			loading = true;
			error = null;

			const response = await fetch(`/api/config/category/${category}`);
			if (!response.ok) {
				throw new Error(`Failed to load ${category} configuration`);
			}

			const configData = await response.json();
			categoryConfigs = configData;
		} catch (e: any) {
			error = e.message;
			notifications.add({ type: 'error', message: 'Uplink Failure', details: e.message });
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

			if (failed.length > 0) throw new Error(`Failed to save ${failed.length} items`);

			notifications.add({ type: 'success', message: 'PROTOCOL_UPDATE_COMMITTED' });
			pendingChanges = new Map();
			await loadCategoryConfig();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Commit Failed', details: e.message });
		} finally {
			saving = false;
		}
	}

	function copyToClipboard(value: string) {
		navigator.clipboard.writeText(value);
		notifications.add({ type: 'success', message: 'Value copied to buffer' });
	}

	onMount(() => {
		if (category) loadCategoryConfig();
	});
</script>

<div class="relative z-10 w-full space-y-10 pb-32 font-jetbrains">
	<!-- Header -->
	<div class="flex flex-col xl:flex-row xl:items-center justify-between gap-8 border-l-4 border-rust pl-6 sm:pl-10 py-4 bg-[var(--header-bg)]/60 backdrop-blur-xl industrial-frame shadow-2xl">
		<div class="flex items-center gap-6">
			<a href="/config" class="p-3 bg-stone-900 border border-stone-800 hover:border-rust transition-all group">
				<ChevronLeft class="w-6 h-6 text-stone-500 group-hover:text-white" />
			</a>
			<div>
				<div class="flex items-center gap-3 mb-1">
					<span class="bg-rust text-white px-2 py-0.5 text-[8px] font-black uppercase tracking-widest">Category: {category}</span>
					<div class="w-px h-3 bg-stone-800"></div>
					<span class="text-text-dim text-[8px] font-black uppercase tracking-widest">Sector: Config_Node</span>
				</div>
				<h1 class="text-3xl sm:text-4xl font-heading font-black text-white uppercase tracking-tighter leading-none">
					{categoryTitles[category] || category.toUpperCase() + '_PARAMS'}
				</h1>
			</div>
		</div>

		<div class="flex items-center gap-4">
			{#if pendingChanges.size > 0}
				<div class="flex items-center gap-4 px-5 py-3 bg-rust/10 border border-rust/30 industrial-frame" transition:scale>
					<div class="w-2 h-2 bg-rust animate-pulse"></div>
					<span class="font-black text-[10px] text-rust-light uppercase tracking-widest">{pendingChanges.size} PENDING</span>
				</div>
				<button onclick={saveChanges} disabled={saving} class="px-8 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-xl shadow-rust/20 transition-all active:tranneutral-y-px">
					{saving ? 'SYNCING...' : 'COMMIT CHANGES'}
				</button>
			{:else}
				<button onclick={loadCategoryConfig} disabled={loading} class="px-8 py-3 bg-stone-950 hover:bg-white hover:text-black text-text-dim font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800 active:tranneutral-y-px">
					<RefreshCw class="w-4 h-4 inline mr-3 {loading ? 'animate-spin' : ''}" />
					Reload_Buffer
				</button>
			{/if}
		</div>
	</div>

	<!-- Search & Filters -->
	<div class="relative group">
		<Search class="absolute left-5 top-1/2 -tranneutral-y-1/2 w-5 h-5 text-text-dim group-focus-within:text-rust transition-colors" />
		<input type="text" bind:value={searchQuery} placeholder="FILTER PARAMETERS..." class="w-full pl-14 pr-10 py-4 bg-stone-950 border border-stone-800 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner" />
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-32 gap-6" transition:fade>
			<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
			<span class="text-text-dim font-black uppercase tracking-[0.4em] animate-pulse">Synchronizing_With_Mainframe...</span>
		</div>
	{:else if filteredConfigs.length === 0}
		<div class="py-32 text-center modern-industrial-card glass-panel border-dashed border-stone-800 !bg-transparent opacity-40">
			<Terminal class="w-12 h-12 text-stone-800 mx-auto mb-6" />
			<p class="text-stone-600 font-black uppercase tracking-widest">No Parameters Located In Buffer</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4">
			{#each filteredConfigs as item (item.key)}
				{@const isPending = pendingChanges.has(item.key)}
				<div class="modern-industrial-card glass-panel group !rounded-none border-stone-800 {isPending ? 'border-rust/40 bg-rust/5' : ''}">
					<div class="p-6 sm:p-8 flex flex-col md:flex-row md:items-center justify-between gap-8">
						<div class="flex-1 space-y-3">
							<div class="flex items-center gap-4">
								<h3 class="text-lg font-black text-white uppercase tracking-tight">{item.key}</h3>
								<div class="flex gap-1">
									{#if item.is_read_only}<span class="px-2 py-0.5 bg-stone-800 text-text-dim text-[7px] font-black border border-stone-700 uppercase">ReadOnly</span>{/if}
									{#if item.requires_restart}<span class="px-2 py-0.5 bg-warning/10 text-warning text-[7px] font-black border border-warning/30 uppercase">Restart Required</span>{/if}
								</div>
							</div>
							<p class="text-[10px] text-text-dim font-bold uppercase leading-relaxed max-w-3xl">{item.description}</p>
						</div>

						<div class="w-full md:w-96 flex items-center gap-3">
							{#if item.type === 'bool'}
								<button 
									onclick={() => handleValueChange(item.key, (pendingChanges.get(item.key) ?? item.value) === 'true' ? 'false' : 'true', item.value)}
									disabled={item.is_read_only}
									class="flex-1 flex items-center justify-between px-6 py-3 border-2 transition-all {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-success/10 border-success text-success' : 'bg-stone-950 border-stone-800 text-text-dim'}"
								>
									<span class="font-black text-[10px] uppercase tracking-[0.2em]">{(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'ACTIVE' : 'DISABLED'}</span>
									<div class="w-2 h-2 {(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'bg-success shadow-[0_0_10px_var(--color-success)] animate-pulse' : 'bg-stone-800'}"></div>
								</button>
							{:else}
								<div class="relative flex-1 group/input">
									<input 
										type="text" 
										value={pendingChanges.get(item.key) ?? item.value}
										oninput={e => handleValueChange(item.key, e.currentTarget.value, item.value)}
										disabled={item.is_read_only}
										class="w-full bg-black border border-stone-800 focus:border-rust text-white font-mono text-xs px-4 py-3 transition-all disabled:opacity-30 shadow-inner"
									/>
									<button onclick={() => copyToClipboard(pendingChanges.get(item.key) ?? item.value)} class="absolute right-3 top-1/2 -tranneutral-y-1/2 text-text-dim hover:text-rust opacity-0 group-hover/input:opacity-100 transition-all">
										<Copy class="w-4 h-4" />
									</button>
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>