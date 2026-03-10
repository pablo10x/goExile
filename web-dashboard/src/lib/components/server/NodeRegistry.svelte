<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { createEventDispatcher } from 'svelte';
	import Icon from '../theme/Icon.svelte';
	import Button from '../Button.svelte';
	import { formatBytes, formatUptime } from '$lib/utils';

	const dispatch = createEventDispatcher();

	interface Instance {
		id: string;
		node_id: number;
		node_name: string;
		status: string;
		port: number;
		player_count: number;
		cpu_usage: number;
		mem_used: number;
		uptime: number;
		version: string;
	}

	let instances = $state<Instance[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let refreshInterval: any;

	async function fetchInstances() {
		try {
			const res = await apiFetch('/api/instances');
			if (res.ok) {
				const data = await res.json();
				let flat: Instance[] = [];
				data.forEach((node: any) => {
					if (node.instances) {
						node.instances.forEach((inst: any) => {
							flat.push({
								...inst,
								node_id: node.node_id,
								node_name: node.node_name
							});
						});
					}
				});
				instances = flat;
			}
		} catch (e) {
			console.error('Instance fetch failed', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchInstances();
		refreshInterval = setInterval(fetchInstances, 5000);
	});

	onDestroy(() => {
		clearInterval(refreshInterval);
	});

	let filteredInstances = $derived(
		instances.filter(
			(i) =>
				i.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
				i.node_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				i.status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	function getStatusClass(status: string) {
		switch (status) {
			case 'Running':
				return 'border-emerald-500/20 bg-emerald-500/[0.02]';
			case 'Provisioning':
				return 'border-amber-500/20 bg-amber-500/[0.02]';
			case 'Error':
				return 'border-rose-500/20 bg-rose-500/[0.02]';
			default:
				return 'border-white/5 bg-white/[0.02]';
		}
	}
</script>

<div class="space-y-6 font-sans">
	<!-- Search & Filters -->
	<div
		class="flex flex-col md:flex-row gap-4 justify-between items-center bg-black/20 p-6 border border-white/5 rounded-2xl backdrop-blur-md"
	>
		<div class="relative flex-1 max-w-md w-full group">
			<Icon
				name="ph:magnifying-glass-bold"
				class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500 group-focus-within:text-sky-400 transition-colors"
				size="1.1rem"
			/>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search by ID, Node, or Status..."
				class="w-full bg-black/40 border border-white/5 pl-12 pr-4 py-3 text-sm font-medium text-white focus:border-sky-500/50 outline-none transition-all rounded-xl shadow-inner"
			/>
		</div>
		<div class="flex items-center gap-4">
			<div
				class="px-4 py-2 bg-black/40 border border-white/5 rounded-xl hidden lg:block shadow-inner"
			>
				<span class="text-xs font-bold text-slate-400 uppercase tracking-wider"
					>{instances.length} Instances Running</span
				>
			</div>
			<Button
				onclick={fetchInstances}
				variant="secondary"
				size="md"
				class="!p-3"
				{loading}
				icon="ph:arrows-clockwise-bold"
			/>
		</div>
	</div>

	{#if loading && instances.length === 0}
		<div class="py-20 flex flex-col items-center justify-center gap-4">
			<div
				class="w-10 h-10 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"
			></div>
			<p class="text-xs font-bold text-slate-500 uppercase tracking-widest animate-pulse">
				Loading instances...
			</p>
		</div>
	{:else if filteredInstances.length === 0}
		<div
			class="py-20 border-2 border-dashed border-white/5 rounded-3xl flex flex-col items-center justify-center opacity-40 bg-black/20"
		>
			<Icon name="ph:cube-transparent-bold" size="3rem" class="text-slate-700 mb-4" />
			<p class="text-xs font-bold text-slate-500 uppercase tracking-widest">
				No matching instances found
			</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4 pb-12">
			{#each filteredInstances as instance (instance.id)}
				<div
					class="group relative flex flex-col lg:flex-row items-center gap-8 p-6 bg-black/40 border border-white/5 transition-all duration-300 hover:border-sky-500/30 hover:bg-slate-900/60 rounded-2xl {getStatusClass(
						instance.status
					)}"
					transition:scale={{ duration: 200, start: 0.98, easing: cubicOut }}
				>
					<!-- Identity Section -->
					<div class="flex items-center gap-5 w-full lg:w-80 shrink-0">
						<div
							class="w-14 h-14 bg-black border border-white/10 rounded-2xl flex items-center justify-center shrink-0 shadow-lg group-hover:border-sky-500/30 transition-colors"
						>
							<Icon
								name="ph:cpu-bold"
								size="1.5rem"
								class={instance.status === 'Running' ? 'text-sky-400' : 'text-slate-700'}
							/>
						</div>
						<div class="min-w-0">
							<div class="flex items-center gap-3 mb-1.5">
								<span class="text-lg font-bold text-white tracking-tight truncate font-heading italic uppercase"
									>{instance.id.split('-').pop()}</span
								>
								{#if instance.status === 'Running'}
									<div
										class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]"
									></div>
								{/if}
							</div>
							<div class="flex items-center gap-2">
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider"
									>Node:</span
								>
								<span class="text-xs font-bold text-sky-500/70 truncate">{instance.node_name}</span>
							</div>
						</div>
					</div>

					<!-- Metrics HUD -->
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-10 flex-1 w-full">
						<div class="space-y-3">
							<div
								class="flex justify-between items-center text-[10px] font-bold text-slate-500 uppercase tracking-wider"
							>
								<span>CPU Usage</span>
								<span class="text-slate-300 tabular-nums">{instance.cpu_usage?.toFixed(1)}%</span>
							</div>
							<div
								class="h-1.5 bg-black rounded-full overflow-hidden border border-white/5 shadow-inner"
							>
								<div
									class="h-full bg-sky-500 transition-all duration-1000"
									style="width: {instance.cpu_usage}%"
								></div>
							</div>
						</div>
						<div class="space-y-3">
							<div
								class="flex justify-between items-center text-[10px] font-bold text-slate-500 uppercase tracking-wider"
							>
								<span>Memory</span>
								<span class="text-slate-300 tabular-nums">{formatBytes(instance.mem_used)}</span>
							</div>
							<div
								class="h-1.5 bg-black rounded-full overflow-hidden border border-white/5 shadow-inner"
							>
								<div
									class="h-full bg-sky-500/60 transition-all duration-1000"
									style="width: 40%"
								></div>
							</div>
						</div>
						<div class="space-y-2">
							<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider block"
								>Users Online</span
							>
							<div class="flex items-center gap-3">
								<span class="text-sm font-bold text-white tabular-nums"
									>{instance.player_count}</span
								>
								<div class="flex gap-1">
									{#each Array(5) as _, i}
										<div
											class="w-1.5 h-4 rounded-sm {i < instance.player_count / 20
												? 'bg-sky-500'
												: 'bg-slate-800'}"
										></div>
									{/each}
								</div>
							</div>
						</div>
						<div class="space-y-2 hidden sm:block">
							<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider block"
								>Uptime</span
							>
							<span class="text-xs font-bold text-slate-300 uppercase tracking-wide"
								>{formatUptime(instance.uptime * 1000)}</span
							>
						</div>
					</div>

					<!-- Actions -->
					<div
						class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all translate-x-4 group-hover:translate-x-0"
					>
						<Button
							onclick={() =>
								dispatch('tail', { nodeId: instance.node_id, instanceId: instance.id })}
							variant="secondary"
							size="sm"
							icon="ph:terminal-window-bold"
							class="!p-3 !rounded-xl"
							title="Open Logs"
						/>

						<div class="w-px h-8 bg-white/5 mx-2"></div>

						{#if instance.status !== 'Running'}
							<Button
								onclick={() =>
									dispatch('start', { nodeId: instance.node_id, instanceId: instance.id })}
								variant="primary"
								size="sm"
								icon="ph:play-bold"
								class="!p-3 !rounded-xl"
								title="Start Instance"
							/>
						{:else}
							<Button
								onclick={() =>
									dispatch('stop', { nodeId: instance.node_id, instanceId: instance.id })}
								variant="secondary"
								size="sm"
								icon="ph:stop-bold"
								class="!p-3 !rounded-xl !text-rose-400 hover:!bg-rose-500/10"
								title="Stop Instance"
							/>
							<Button
								onclick={() =>
									dispatch('restart', { nodeId: instance.node_id, instanceId: instance.id })}
								variant="secondary"
								size="sm"
								icon="ph:arrows-clockwise-bold"
								class="!p-3 !rounded-xl"
								title="Restart Instance"
							/>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
</style>
