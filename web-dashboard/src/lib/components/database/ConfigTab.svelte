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

<div class="h-full flex flex-col bg-slate-900">
	<div
		class="p-6 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 flex justify-between items-center"
	>
		<div class="flex items-center gap-3">
			<div class="p-2 bg-purple-500/10 rounded-lg">
				<Settings class="w-6 h-6 text-purple-400" />
			</div>
			<div>
				<h2 class="text-xl font-bold text-slate-100">PostgreSQL Configuration</h2>
				<p class="text-sm text-slate-500">Tune database parameters (postgresql.conf)</p>
			</div>
		</div>
		<div class="flex gap-2">
			{#if Object.keys(pendingChanges).length > 0}
				<button
					onclick={saveChanges}
					disabled={loading}
					class="px-4 py-2 bg-orange-600 hover:bg-orange-500 text-slate-900 dark:text-white rounded-lg font-bold flex items-center gap-2 shadow-lg animate-pulse"
				>
					<Save class="w-4 h-4" /> Apply Changes
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar -->
	<div
		class="px-6 py-3 bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 flex justify-between"
	>
		<div class="relative w-64">
			<Search class="absolute left-2.5 top-2.5 w-4 h-4 text-slate-500" />
			<input
				type="text"
				bind:value={filter}
				placeholder="Filter parameters..."
				class="w-full bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-lg pl-9 pr-3 py-2 text-sm text-slate-800 dark:text-slate-200 focus:border-purple-500 outline-none"
			/>
		</div>
		<div class="text-xs text-slate-500 flex items-center gap-2">
			<AlertCircle class="w-3 h-3" />
			<span>Changes trigger config reload</span>
		</div>
	</div>

	<div class="flex-1 overflow-auto">
		{#if loading && config.length === 0}
			<div class="flex justify-center p-12">
				<div
					class="w-8 h-8 border-4 border-purple-500 border-t-transparent rounded-full animate-spin"
				></div>
			</div>
		{:else}
			<table class="w-full text-left text-sm border-collapse">
				<thead
					class="bg-white dark:bg-slate-950 text-slate-500 dark:text-slate-400 sticky top-0 shadow-sm z-10"
				>
					<tr>
						<th class="px-6 py-3 border-b border-slate-200 dark:border-slate-800 w-1/3"
							>Parameter</th
						>
						<th class="px-6 py-3 border-b border-slate-200 dark:border-slate-800 w-1/3">Value</th>
						<th class="px-6 py-3 border-b border-slate-200 dark:border-slate-800">Description</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-800/50">
					{#each filteredConfig as cfg}
						<tr
							class="hover:bg-slate-800/30 transition-colors {pendingChanges[cfg.name]
								? 'bg-orange-500/5'
								: ''}"
						>
							<td class="px-6 py-3 font-mono text-purple-300 font-medium">
								{cfg.name}
							</td>
							<td class="px-6 py-3">
								<input
									type="text"
									value={pendingChanges[cfg.name] !== undefined
										? pendingChanges[cfg.name]
										: cfg.setting}
									oninput={(e) => updateValue(cfg.name, e.currentTarget.value)}
									class="w-full bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1 text-slate-800 dark:text-slate-200 focus:border-orange-500 outline-none font-mono text-xs"
								/>
							</td>
							<td class="px-6 py-3 text-slate-500 text-xs">
								{cfg.description}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>
