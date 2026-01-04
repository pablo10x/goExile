<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { createEventDispatcher } from 'svelte';
	import Icon from './theme/Icon.svelte';
	import { formatUptime } from '$lib/utils';

	const dispatch = createEventDispatcher();

	interface MiniInstance {
		id: string;
		node_id: number;
		node_name: string;
		status: string;
		player_count: number;
		uptime: number;
	}

	let instances = $state<MiniInstance[]>([]);
	let loading = $state(true);
	let interval: any;

	async function fetchRecent() {
		try {
			const res = await fetch('/api/instances');
			if (res.ok) {
				const data = await res.json();
				let active: MiniInstance[] = [];
				data.forEach((node: any) => {
					if (node.instances) {
						node.instances.forEach((inst: any) => {
							if (inst.status === 'Running') {
								active.push({
									...inst,
									node_id: node.node_id,
									node_name: node.node_name
								});
							}
						});
					}
				});
				instances = active.sort((a, b) => b.uptime - a.uptime).slice(0, 8);
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchRecent();
		interval = setInterval(fetchRecent, 5000);
	});

	onDestroy(() => clearInterval(interval));
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between px-2">
		<div class="flex items-center gap-3">
			<div class="w-1 h-4 bg-rust shadow-[0_0_8px_#c2410c]"></div>
			<h3 class="text-[10px] font-mono font-black text-neutral-500 uppercase tracking-[0.3em] italic">Active_Combat_Stream</h3>
		</div>
		<span class="text-[8px] font-black text-neutral-600 uppercase tracking-widest">{instances.length} TARGETS_LOCKED</span>
	</div>

	{#if loading && instances.length === 0}
		<div class="flex gap-4 overflow-x-auto no-scrollbar pb-4">
			{#each Array(4) as _}
				<div class="min-w-[280px] h-32 bg-neutral-900/40 border border-neutral-800 rounded-none animate-pulse"></div>
			{/each}
		</div>
	{:else if instances.length === 0}
		<div class="p-8 border-2 border-dashed border-neutral-800 rounded-none text-center opacity-30 bg-neutral-950/40">
			<p class="text-[10px] font-black text-neutral-600 uppercase tracking-widest">No active tactical signatures detected</p>
		</div>
	{:else}
		<div class="flex gap-4 overflow-x-auto no-scrollbar pb-4 -mx-2 px-2">
			{#each instances as inst (inst.id)}
				<div 
					class="min-w-[280px] group relative bg-neutral-900/60 border-2 border-neutral-800 p-4 rounded-none transition-all duration-500 hover:border-rust/40 hover:bg-neutral-800/80 shadow-2xl"
					in:fly={{ x: 20, duration: 400 }}
				>
					<div class="absolute inset-0 bg-rust/0 group-hover:bg-rust/5 transition-colors duration-500"></div>
					
					<div class="relative z-10 flex justify-between items-start mb-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 bg-black border border-neutral-800 rounded-none flex items-center justify-center text-rust-light group-hover:border-rust/30 transition-colors shadow-inner">
								<Icon name="ph:cube-bold" size="1rem" />
							</div>
							<div class="flex flex-col">
								<span class="text-[11px] font-mono font-black text-white uppercase tracking-tighter truncate w-32">{inst.id.split('-').pop()}</span>
								<span class="text-[7px] font-black text-neutral-600 uppercase tracking-widest italic">{inst.node_name}</span>
							</div>
						</div>
						<div class="flex gap-1">
							<button 
								onclick={() => dispatch('tail', { nodeId: inst.node_id, instanceId: inst.id })}
								class="p-1.5 text-neutral-600 hover:text-white hover:bg-neutral-700 transition-all border border-transparent hover:border-neutral-600"
							>
								<Icon name="ph:terminal-window-bold" size="0.875rem" />
							</button>
							<button 
								onclick={() => dispatch('stop', { nodeId: inst.node_id, instanceId: inst.id })}
								class="p-1.5 text-red-500/60 hover:text-white hover:bg-red-600 transition-all border border-transparent hover:border-red-500/40"
							>
								<Icon name="ph:stop-bold" size="0.875rem" />
							</button>
						</div>
					</div>

					<div class="relative z-10 flex items-end justify-between mt-auto">
						<div class="space-y-1">
							<span class="block text-[7px] font-black text-neutral-600 uppercase tracking-widest">UPTIME</span>
							<span class="text-[10px] font-mono font-black text-rust-light/80 tabular-nums">{formatUptime(inst.uptime * 1000)}</span>
						</div>
						<div class="flex flex-col items-end gap-1">
							<span class="text-[7px] font-black text-neutral-600 uppercase tracking-widest">PAYLOAD</span>
							<div class="flex items-center gap-1.5">
								<span class="text-[10px] font-mono font-black text-neutral-300 tabular-nums">{inst.player_count}</span>
								<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_#10b981]"></div>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.no-scrollbar::-webkit-scrollbar { display: none; }
	.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
