<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { createEventDispatcher } from 'svelte';
	import Icon from '../theme/Icon.svelte';
	import Button from '../Button.svelte';
	import { formatBytes, formatUptime } from '$lib/utils';

	const dispatch = createEventDispatcher();

	interface FleetInstance {
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

	let instances = $state<FleetInstance[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let refreshInterval: any;

	async function fetchFleet() {
		try {
			const res = await fetch('/api/instances');
			if (res.ok) {
				const data = await res.json();
				// Flatten the nested node results
				let flat: FleetInstance[] = [];
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
			console.error('Fleet fetch failed', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchFleet();
		refreshInterval = setInterval(fetchFleet, 5000);
	});

	onDestroy(() => {
		clearInterval(refreshInterval);
	});

	let filteredInstances = $derived(
		instances.filter(i => 
			i.id.toLowerCase().includes(searchQuery.toLowerCase()) || 
			i.node_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			i.status.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	function getStatusAura(status: string) {
		switch (status) {
			case 'Running': return 'shadow-[0_0_15px_rgba(16,185,129,0.15)] border-emerald-500/30';
			case 'Provisioning': return 'shadow-[0_0_15px_rgba(249,115,22,0.15)] border-amber-500/30';
			case 'Error': return 'shadow-[0_0_15px_rgba(239,68,68,0.15)] border-red-500/30';
			default: return 'border-slate-800';
		}
	}
</script>

<div class="space-y-6">
	<!-- Control Bar -->
	<div class="flex flex-col md:flex-row gap-4 justify-between items-center bg-slate-900/40 p-4 border border-slate-800 rounded-xl backdrop-blur-md">
		<div class="relative flex-1 max-w-md w-full">
			<Icon name="ph:magnifying-glass-bold" class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size="1.1rem" />
			<input 
				type="text" 
				bind:value={searchQuery}
				placeholder="OMNI_SEARCH: INSTANCE_ID, NODE, OR STATUS..."
				class="w-full bg-slate-950/50 border border-slate-800 pl-12 pr-4 py-2.5 text-xs font-jetbrains font-bold text-white focus:border-indigo-500 outline-none transition-all uppercase tracking-widest rounded-lg shadow-inner"
			/>
		</div>
		<div class="flex items-center gap-4">
			<div class="px-4 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-lg hidden lg:block">
				<span class="text-[10px] font-black text-indigo-400 uppercase tracking-widest">{instances.length} ACTIVE_ENTITIES</span>
			</div>
			<button 
				onclick={fetchFleet}
				class="p-2.5 bg-slate-800 border border-slate-700 hover:border-indigo-500 text-slate-400 hover:text-white transition-all rounded-lg"
			>
				<Icon name="ph:arrows-clockwise-bold" class={loading ? 'animate-spin' : ''} size="1.1rem" />
			</button>
		</div>
	</div>

	{#if loading && instances.length === 0}
		<div class="py-20 flex flex-col items-center justify-center gap-6">
			<div class="w-12 h-12 border-2 border-indigo-500 border-t-transparent rounded-none animate-spin"></div>
			<p class="font-heading font-black text-xs text-slate-500 uppercase tracking-[0.3em] animate-pulse">Synchronizing Fleet Stream...</p>
		</div>
	{:else if filteredInstances.length === 0}
		<div class="py-20 border-2 border-dashed border-slate-800 rounded-2xl flex flex-col items-center justify-center opacity-40">
			<Icon name="ph:cube-transparent-bold" size="3rem" class="text-slate-700 mb-4" />
			<p class="font-jetbrains font-black text-xs text-slate-500 uppercase tracking-widest">No matching tactical signatures found</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-3">
			{#each filteredInstances as instance (instance.id)}
				<div 
					class="group relative flex flex-col lg:flex-row items-center gap-6 p-4 bg-slate-900/40 border transition-all duration-500 hover:bg-slate-800/60 rounded-xl {getStatusAura(instance.status)}"
					transition:scale={{ duration: 200, start: 0.98, easing: cubicOut }}
				>
					<!-- Identity Section -->
					<div class="flex items-center gap-5 w-full lg:w-72 shrink-0">
						<div class="w-12 h-12 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-center shrink-0 shadow-xl group-hover:border-indigo-500/50 transition-colors">
							<Icon name="ph:cpu-bold" size="1.25rem" class={instance.status === 'Running' ? 'text-indigo-400' : 'text-slate-700'} />
						</div>
						<div class="min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<span class="font-heading font-black text-sm text-white uppercase tracking-tighter truncate">{instance.id.split('-').pop()}</span>
								{#if instance.status === 'Running'}
									<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_#10b981]"></div>
								{/if}
							</div>
							<div class="flex items-center gap-2">
								<span class="text-[8px] font-black text-slate-600 uppercase tracking-widest">NODE:</span>
								<span class="text-[9px] font-bold text-indigo-400/70 uppercase tracking-tight truncate">{instance.node_name}</span>
							</div>
						</div>
					</div>

					<!-- Metrics HUD -->
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-8 flex-1 w-full">
						<div class="space-y-2">
							<div class="flex justify-between items-center text-[8px] font-black text-slate-600 uppercase tracking-widest">
								<span>CPU_LOAD</span>
								<span class="text-slate-400">{instance.cpu_usage?.toFixed(1)}%</span>
							</div>
							<div class="h-1 bg-slate-950 rounded-full overflow-hidden border border-slate-800/50">
								<div class="h-full bg-indigo-500 transition-all duration-1000" style="width: {instance.cpu_usage}%"></div>
							</div>
						</div>
						<div class="space-y-2">
							<div class="flex justify-between items-center text-[8px] font-black text-slate-600 uppercase tracking-widest">
								<span>MEM_SYNC</span>
								<span class="text-slate-400">{formatBytes(instance.mem_used)}</span>
							</div>
							<div class="h-1 bg-slate-950 rounded-full overflow-hidden border border-slate-800/50">
								<div class="h-full bg-indigo-400/60 transition-all duration-1000" style="width: 40%"></div>
							</div>
						</div>
						<div class="space-y-1">
							<span class="text-[8px] font-black text-slate-600 uppercase tracking-widest block">PLAYERS</span>
							<div class="flex items-center gap-2">
								<span class="text-xs font-jetbrains font-black text-white tabular-nums">{instance.player_count}</span>
								<div class="flex gap-0.5">
									{#each Array(5) as _, i}
										<div class="w-1 h-3 rounded-full {i < (instance.player_count / 20) ? 'bg-indigo-500' : 'bg-slate-800'}"></div>
									{/each}
								</div>
							</div>
						</div>
						<div class="space-y-1 hidden sm:block">
							<span class="text-[8px] font-black text-slate-600 uppercase tracking-widest block">UPTIME</span>
							<span class="text-[10px] font-jetbrains font-bold text-slate-400 uppercase">{formatUptime(instance.uptime * 1000)}</span>
						</div>
					</div>

					<!-- Quick-Action Blade -->
					<div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all translate-x-4 group-hover:translate-x-0">
						<button 
							onclick={() => dispatch('tail', { nodeId: instance.node_id, instanceId: instance.id })}
							class="p-2.5 bg-slate-800/80 border border-slate-700 hover:border-indigo-500 text-slate-400 hover:text-white transition-all rounded-lg shadow-xl"
							title="Open Terminal"
						>
							<Icon name="ph:terminal-window-bold" size="1.1rem" />
						</button>
						
						<div class="w-px h-6 bg-slate-800 mx-1"></div>

						{#if instance.status !== 'Running'}
							<button 
								onclick={() => dispatch('start', { nodeId: instance.node_id, instanceId: instance.id })}
								class="p-2.5 bg-emerald-500/10 border border-emerald-500/20 hover:bg-emerald-500 hover:text-white text-emerald-500 transition-all rounded-lg shadow-xl"
								title="Start Instance"
							>
								<Icon name="ph:play-bold" size="1.1rem" />
							</button>
						{:else}
							<button 
								onclick={() => dispatch('stop', { nodeId: instance.node_id, instanceId: instance.id })}
								class="p-2.5 bg-rose-500/10 border border-rose-500/20 hover:bg-rose-500 hover:text-white text-rose-500 transition-all rounded-lg shadow-xl"
								title="Stop Instance"
							>
								<Icon name="ph:stop-bold" size="1.1rem" />
							</button>
							<button 
								onclick={() => dispatch('restart', { nodeId: instance.node_id, instanceId: instance.id })}
								class="p-2.5 bg-amber-500/10 border border-amber-500/20 hover:bg-amber-500 hover:text-white text-amber-500 transition-all rounded-lg shadow-xl"
								title="Restart Instance"
							>
								<Icon name="ph:arrows-clockwise-bold" size="1.1rem" />
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.font-heading { font-family: 'Inter', sans-serif; }
	.font-jetbrains { font-family: 'JetBrains Mono', monospace; }
</style>
