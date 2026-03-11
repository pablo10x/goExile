<script lang="ts">
	import { apiFetch } from '$lib/api';
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
			const res = await apiFetch('/api/instances');
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
			<div class="w-1 h-4 bg-sky-500 shadow-[0_0_8px_rgba(14,165,233,0.5)] rounded-full"></div>
			<h3 class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em]">
				Deployment Overview
			</h3>
		</div>
		<span class="text-[8px] font-bold text-slate-600 uppercase tracking-widest"
			>{instances.length} Instances Running</span
		>
	</div>

	{#if loading && instances.length === 0}
		<div class="flex gap-4 overflow-x-auto no-scrollbar pb-4">
			{#each Array(4) as _}
				<div
					class="min-w-[280px] h-32 bg-black/20 border border-white/5 rounded-2xl animate-pulse"
				></div>
			{/each}
		</div>
	{:else if instances.length === 0}
		<div
			class="p-8 border-2 border-dashed border-white/5 rounded-3xl text-center opacity-30 bg-black/40"
		>
			<p class="text-[10px] font-bold text-slate-500 uppercase tracking-widest font-sans">
				No active sessions detected
			</p>
		</div>
	{:else}
		<div class="flex gap-4 overflow-x-auto no-scrollbar pb-4 -mx-2 px-2 font-sans">
			{#each instances as inst (inst.id)}
				<div
					class="min-w-[280px] group relative bg-black/40 border border-white/5 p-5 rounded-2xl transition-all duration-300 hover:border-sky-500/30 hover:bg-slate-900/60 shadow-xl"
					in:fly={{ x: 20, duration: 400 }}
				>
					<div class="relative z-10 flex justify-between items-start mb-4">
						<div class="flex items-center gap-3">
							<div
								class="w-10 h-10 bg-black border border-white/10 rounded-xl flex items-center justify-center text-sky-400 group-hover:border-sky-500/30 transition-colors shadow-inner"
							>
								<Icon name="ph:cpu-bold" size="1.2rem" />
							</div>
							<div class="flex flex-col">
								<span class="text-sm font-bold text-white tracking-tight truncate w-32 font-heading italic uppercase"
									>{inst.id.split('-').pop()}</span
								>
								<span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider"
									>{inst.node_name}</span
								>
							</div>
						</div>
						<div class="flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
							<button
								onclick={() => dispatch('tail', { nodeId: inst.node_id, instanceId: inst.id })}
								class="p-2 text-slate-400 hover:text-white hover:bg-white/5 rounded-lg transition-all"
								title="Terminal"
							>
								<Icon name="ph:terminal-window-bold" size="1rem" />
							</button>
							<button
								onclick={() => dispatch('stop', { nodeId: inst.node_id, instanceId: inst.id })}
								class="p-2 text-rose-500/60 hover:text-white hover:bg-rose-500 rounded-lg transition-all"
								title="Stop"
							>
								<Icon name="ph:stop-bold" size="1rem" />
							</button>
						</div>
					</div>

					<div class="relative z-10 flex items-end justify-between mt-2">
						<div class="space-y-1">
							<span class="block text-[8px] font-bold text-slate-500 uppercase tracking-widest"
								>Uptime</span
							>
							<span class="text-xs font-mono font-bold text-sky-400/80 tabular-nums"
								>{formatUptime(inst.uptime * 1000)}</span
							>
						</div>
						<div class="flex flex-col items-end gap-1">
							<span class="text-[8px] font-bold text-slate-500 uppercase tracking-widest"
								>Players</span
							>
							<div class="flex items-center gap-2">
								<span class="text-xs font-mono font-bold text-slate-200 tabular-nums"
									>{inst.player_count}</span
								>
								<div
									class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"
								></div>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
