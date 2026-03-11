<script lang="ts">
	import { apiFetch } from '$lib/api';
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
		aesthetic: 'Interface Settings'
	};

	let filteredConfigs = $derived.by(() => {
		if (!searchQuery.trim()) return categoryConfigs;
		const q = searchQuery.toLowerCase();
		return categoryConfigs.filter(
			(c) =>
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

			const response = await apiFetch(`/api/config/category/${category}`);
			if (!response.ok) {
				throw new Error(`Failed to load ${category} configuration`);
			}

			const configData = await response.json();
			categoryConfigs = configData;
		} catch (e: any) {
			error = e.message;
			notifications.add({ type: 'error', message: 'Connection Error', details: e.message });
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
					apiFetch(`/api/config/${key}`, {
						method: 'PUT',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ value })
					})
				);
			}

			const results = await Promise.all(promises);
			const failed = results.filter((r) => !r.ok);

			if (failed.length > 0) throw new Error(`Failed to save ${failed.length} items`);

			notifications.add({ type: 'success', message: 'Changes successfully applied' });
			pendingChanges = new Map();
			await loadCategoryConfig();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Save Failed', details: e.message });
		} finally {
			saving = false;
		}
	}

	function copyToClipboard(value: string) {
		navigator.clipboard.writeText(value);
		notifications.add({ type: 'success', message: 'Value copied to clipboard' });
	}

	onMount(() => {
		if (category) loadCategoryConfig();
	});
</script>

<div class="relative z-10 w-full space-y-10 pb-32 font-sans">
	<!-- Header -->
	<div
		class="flex flex-col xl:flex-row xl:items-center justify-between gap-8 border-l-4 border-sky-500 pl-6 sm:pl-10 py-4 bg-slate-900/60 backdrop-blur-xl rounded-2xl border border-white/5 shadow-2xl"
	>
		<div class="flex items-center gap-6">
			<a
				href="/config"
				class="p-3 bg-slate-950 border border-white/5 hover:border-sky-500 rounded-xl transition-all group"
			>
				<ChevronLeft class="w-6 h-6 text-slate-500 group-hover:text-white" />
			</a>
			<div>
				<div class="flex items-center gap-3 mb-1">
					<span
						class="bg-sky-500 text-white px-2 py-0.5 text-[8px] font-black uppercase tracking-widest rounded"
						>Category: {category}</span
					>
					<div class="w-px h-3 bg-white/10"></div>
					<span class="text-slate-500 text-[8px] font-black uppercase tracking-widest"
						>Module: System_Config</span
					>
				</div>
				<h1
					class="text-3xl sm:text-4xl font-black text-white uppercase tracking-tighter leading-none"
				>
					{categoryTitles[category] || category.toUpperCase() + '_SETTINGS'}
				</h1>
			</div>
		</div>

		<div class="flex items-center gap-4">
			{#if pendingChanges.size > 0}
				<div
					class="flex items-center gap-4 px-5 py-3 bg-sky-500/10 border border-sky-500/30 rounded-xl"
					transition:scale
				>
					<div class="w-2 h-2 bg-sky-500 animate-pulse rounded-full"></div>
					<span class="font-black text-[10px] text-sky-400 uppercase tracking-widest"
						>{pendingChanges.size} PENDING</span
					>
				</div>
				<button
					onclick={saveChanges}
					disabled={saving}
					class="px-8 py-3 bg-sky-500 hover:bg-sky-400 text-white font-black text-[11px] uppercase tracking-widest shadow-xl shadow-sky-500/20 rounded-xl transition-all active:translate-y-px"
				>
					{saving ? 'SAVING...' : 'APPLY CHANGES'}
				</button>
			{:else}
				<button
					onclick={loadCategoryConfig}
					disabled={loading}
					class="px-8 py-3 bg-slate-950 hover:bg-white hover:text-black text-slate-400 font-black text-[11px] uppercase tracking-widest transition-all border border-white/5 rounded-xl active:translate-y-px"
				>
					<RefreshCw class="w-4 h-4 inline mr-3 {loading ? 'animate-spin' : ''}" />
					REFRESH DATA
				</button>
			{/if}
		</div>
	</div>

	<!-- Search & Filters -->
	<div class="relative group">
		<Search
			class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-sky-500 transition-colors"
		/>
		<input
			type="text"
			bind:value={searchQuery}
			placeholder="FILTER PARAMETERS..."
			class="w-full pl-14 pr-10 py-4 bg-slate-950 border border-white/5 text-slate-200 font-sans text-xs focus:border-sky-500 outline-none transition-all uppercase tracking-widest rounded-2xl shadow-inner"
		/>
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-32 gap-6" transition:fade>
			<div
				class="w-16 h-16 border-2 border-sky-500 border-t-transparent rounded-full animate-spin"
			></div>
			<span class="text-slate-500 font-black uppercase tracking-[0.4em] animate-pulse"
				>Loading Configuration...</span
			>
		</div>
	{:else if filteredConfigs.length === 0}
		<div
			class="py-32 text-center bg-slate-900/20 border-2 border-dashed border-white/5 rounded-3xl opacity-40"
		>
			<Terminal class="w-12 h-12 text-slate-800 mx-auto mb-6" />
			<p class="text-slate-600 font-black uppercase tracking-widest">
				No parameters found
			</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4">
			{#each filteredConfigs as item (item.key)}
				{@const isPending = pendingChanges.has(item.key)}
				<div
					class="bg-slate-900/40 backdrop-blur-md border border-white/5 rounded-2xl group {isPending
						? 'border-sky-500/40 bg-sky-500/5'
						: 'hover:border-white/10'} transition-all"
				>
					<div class="p-6 sm:p-8 flex flex-col md:flex-row md:items-center justify-between gap-8">
						<div class="flex-1 space-y-3">
							<div class="flex items-center gap-4">
								<h3 class="text-lg font-black text-white uppercase tracking-tight">{item.key}</h3>
								<div class="flex gap-1">
									{#if item.is_read_only}<span
											class="px-2 py-0.5 bg-slate-800 text-slate-500 text-[7px] font-black border border-white/5 uppercase rounded"
											>Read Only</span
										>{/if}
									{#if item.requires_restart}<span
											class="px-2 py-0.5 bg-amber-500/10 text-amber-500 text-[7px] font-black border border-amber-500/30 uppercase rounded"
											>Restart Required</span
										>{/if}
								</div>
							</div>
							<p class="text-[10px] text-slate-400 font-bold uppercase leading-relaxed max-w-3xl">
								{item.description}
							</p>
						</div>

						<div class="w-full md:w-96 flex items-center gap-3">
							{#if item.type === 'bool'}
								<button
									onclick={() =>
										handleValueChange(
											item.key,
											(pendingChanges.get(item.key) ?? item.value) === 'true' ? 'false' : 'true',
											item.value
										)}
									disabled={item.is_read_only}
									class="flex-1 flex items-center justify-between px-6 py-3 border-2 rounded-xl transition-all {(pendingChanges.get(
										item.key
									) ?? item.value) === 'true'
										? 'bg-emerald-500/10 border-emerald-500 text-emerald-500 shadow-lg shadow-emerald-500/10'
										: 'bg-slate-950 border-white/5 text-slate-500'}"
								>
									<span class="font-black text-[10px] uppercase tracking-[0.2em]"
										>{(pendingChanges.get(item.key) ?? item.value) === 'true'
											? 'ACTIVE'
											: 'DISABLED'}</span
									>
									<div
										class="w-2 h-2 rounded-full {(pendingChanges.get(item.key) ?? item.value) === 'true'
											? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)] animate-pulse'
											: 'bg-slate-800'}"
									></div>
								</button>
							{:else}
								<div class="relative flex-1 group/input">
									<input
										type="text"
										value={pendingChanges.get(item.key) ?? item.value}
										oninput={(e) => handleValueChange(item.key, e.currentTarget.value, item.value)}
										disabled={item.is_read_only}
										class="w-full bg-slate-950 border border-white/5 rounded-xl focus:border-sky-500 text-white font-mono text-xs px-4 py-3 transition-all disabled:opacity-30 shadow-inner outline-none"
									/>
									<button
										onclick={() => copyToClipboard(pendingChanges.get(item.key) ?? item.value)}
										class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-sky-500 opacity-0 group-hover/input:opacity-100 transition-all"
									>
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
