<script lang="ts">
	import { Settings, Save, RefreshCw, AlertCircle, Search } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores';

	let config = $state<any[]>([]);
	let loading = $state(false);
	let filter = $state('');
	let pendingChanges = $state<Record<string, string>>({});

	let filteredConfig = $derived(
		config.filter(
			(c) =>
				c.name.toLowerCase().includes(filter.toLowerCase()) ||
				(c.description && c.description.toLowerCase().includes(filter.toLowerCase()))
		)
	);

	async function loadConfig() {
		loading = true;
		try {
			const res = await fetch('/api/database/config');
			if (res.ok) {
				config = await res.json();
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to load config', details: e.message });
		} finally {
			loading = false;
		}
	}

	function updateValue(name: string, value: string) {
		pendingChanges[name] = value;
	}

	async function saveChanges() {
		if (Object.keys(pendingChanges).length === 0) return;

		loading = true;
		try {
			for (const [name, value] of Object.entries(pendingChanges)) {
				await fetch('/api/database/config', {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ name, value })
				});
			}

			// Reload
			const restart = await fetch('/api/database/config/restart', { method: 'POST' });
			if (!restart.ok) throw new Error('Failed to reload config');

			notifications.add({ type: 'success', message: 'Configuration applied & reloaded' });
			pendingChanges = {};
			loadConfig();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Save failed', details: e.message });
			loading = false;
		}
	}

	onMount(() => {
		loadConfig();
	});
</script>

<div class="h-full flex flex-col bg-[var(--terminal-bg)]">
	<div
		class="p-6 border-b border-stone-800 bg-[var(--header-bg)] flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
				<Settings class="w-6 h-6 text-rust-light" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-slate-100 uppercase tracking-tighter">KERNEL_TUNING_MODULE</h2>
				<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest mt-1">Tune database parameters (postgresql.conf)</p>
			</div>
		</div>
		<div class="flex gap-3">
			{#if Object.keys(pendingChanges).length > 0}
				<button
					onclick={saveChanges}
					disabled={loading}
					class="px-6 py-2.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 animate-pulse transition-all active:translate-y-px"
				>
					<Save class="w-4 h-4" /> Commit_Changes
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar -->
	<div
		class="px-6 py-4 bg-[var(--header-bg)]/50 border-b border-stone-800 flex justify-between items-center"
	>
		<div class="relative w-80">
			<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-dim" />
			<input
				type="text"
				bind:value={filter}
				placeholder="FILTER_PARAMS..."
				class="w-full bg-stone-950 border border-stone-800 py-2.5 pl-10 pr-4 font-jetbrains text-[10px] text-stone-300 placeholder:text-stone-700 focus:border-rust outline-none uppercase tracking-widest"
			/>
		</div>
		<div class="font-jetbrains text-[9px] font-black text-warning/60 flex items-center gap-3 uppercase tracking-widest">
			<AlertCircle class="w-3.5 h-3.5" />
			<span>WARNING: Changes trigger global core reload</span>
		</div>
	</div>

	<div class="flex-1 overflow-auto custom-scrollbar">
		{#if loading && config.length === 0}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-rust border-t-transparent rounded-none animate-spin"
				></div>
			</div>
		{:else}
			<table class="w-full text-left font-jetbrains text-[11px] border-collapse">
				<thead
					class="bg-[var(--header-bg)] text-text-dim sticky top-0 z-10 border-b border-stone-800"
				>
					<tr>
						<th class="px-6 py-4 font-black uppercase tracking-widest w-1/3 border-r border-stone-800/30"
							>Parameter_Key</th
						>
						<th class="px-6 py-4 font-black uppercase tracking-widest w-1/4 border-r border-stone-800/30">Target_Value</th>
						<th class="px-6 py-4 font-black uppercase tracking-widest">Protocol_Description</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-stone-900">
					{#each filteredConfig as cfg}
						{@const isModified = pendingChanges[cfg.name] !== undefined}
						<tr
							class="hover:bg-rust/5 transition-colors {isModified ? 'bg-amber-500/5' : ''}"
						>
							<td class="px-6 py-4 font-bold text-rust-light uppercase tracking-tight border-r border-stone-800/20">
								{cfg.name}
							</td>
							<td class="px-6 py-4 border-r border-stone-800/20">
								<input
									type="text"
									value={isModified ? pendingChanges[cfg.name] : cfg.setting}
									oninput={(e) => updateValue(cfg.name, e.currentTarget.value)}
									class="w-full bg-stone-950 border {isModified ? 'border-amber-500' : 'border-stone-800'} py-1.5 px-3 text-stone-200 font-jetbrains text-[10px] focus:border-rust outline-none"
								/>
							</td>
							<td class="px-6 py-4 text-text-dim font-medium leading-relaxed">
								{cfg.description}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>
