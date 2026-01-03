<script lang="ts">
	import { Settings, Save, RefreshCw, AlertCircle, Search } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';

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

<div class="h-full flex flex-col bg-transparent">
	<div
		class="p-6 border-b border-slate-800 bg-slate-900/40 backdrop-blur-md flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-indigo-500/10 border border-indigo-500/20 rounded-xl">
				<Settings class="w-6 h-6 text-indigo-400" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-white uppercase tracking-tighter italic">KERNEL_TUNING_MODULE</h2>
				<p class="font-jetbrains text-[10px] text-slate-500 uppercase tracking-widest mt-1 italic font-bold">Tune database parameters (postgresql.conf)</p>
			</div>
		</div>
		<div class="flex gap-3">
			{#if Object.keys(pendingChanges).length > 0}
				<button
					onclick={saveChanges}
					disabled={loading}
					class="px-6 py-2.5 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 animate-pulse transition-all active:translate-y-px rounded-xl"
				>
					<Save class="w-4 h-4" /> Commit_Changes
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar -->
	<div
		class="px-6 py-4 bg-slate-900/20 border-b border-slate-800 flex justify-between items-center"
	>
		<div class="relative w-80">
			<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-600" />
			<input
				type="text"
				bind:value={filter}
				placeholder="FILTER_PARAMS..."
				class="w-full bg-slate-950/40 border border-slate-800 py-2.5 pl-10 pr-4 font-jetbrains text-[10px] text-slate-300 placeholder:text-slate-700 focus:border-indigo-500 outline-none uppercase tracking-widest rounded-xl"
			/>
		</div>
		<div class="font-jetbrains text-[9px] font-bold text-amber-500/60 flex items-center gap-3 uppercase tracking-widest italic">
			<AlertCircle class="w-3.5 h-3.5" />
			<span>WARNING: Changes trigger global core reload</span>
		</div>
	</div>

	<div class="flex-1 overflow-auto custom-scrollbar bg-transparent">
		{#if loading && config.length === 0}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(99,102,241,0.4)]"
				></div>
			</div>
		{:else}
			<table class="w-full text-left font-jetbrains text-[11px] border-collapse">
				<thead
					class="bg-slate-950/50 text-slate-500 sticky top-0 z-10 border-b border-slate-800 backdrop-blur-md"
				>
					<tr>
						<th class="px-6 py-4 font-bold uppercase tracking-widest w-1/3 border-r border-slate-800/30 italic"
							>Parameter_Key</th
						>
						<th class="px-6 py-4 font-bold uppercase tracking-widest w-1/4 border-r border-slate-800/30 italic">Target_Value</th>
						<th class="px-6 py-4 font-bold uppercase tracking-widest italic">Protocol_Description</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-800/50">
					{#each filteredConfig as cfg}
						{@const isModified = pendingChanges[cfg.name] !== undefined}
						<tr
							class="hover:bg-indigo-500/5 transition-colors {isModified ? 'bg-indigo-500/5' : ''}"
						>
							<td class="px-6 py-4 font-bold text-indigo-400/80 uppercase tracking-tighter border-r border-slate-800/20 italic">
								{cfg.name}
							</td>
							<td class="px-6 py-4 border-r border-slate-800/20">
								<input
									type="text"
									value={isModified ? pendingChanges[cfg.name] : cfg.setting}
									oninput={(e) => updateValue(cfg.name, e.currentTarget.value)}
									class="w-full bg-slate-950/40 border {isModified ? 'border-indigo-500' : 'border-slate-800'} py-1.5 px-3 text-slate-200 font-jetbrains text-[10px] focus:border-indigo-500 outline-none rounded-lg"
								/>
							</td>
							<td class="px-6 py-4 text-slate-500 font-medium uppercase tracking-tight leading-relaxed text-[10px]">
								{cfg.description}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>
