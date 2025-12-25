<script lang="ts">
	import type { Spawner } from '$lib/stores';
	import { serverVersions, siteSettings } from '$lib/stores';
	import { createEventDispatcher } from 'svelte';
	import { formatBytes } from '$lib/utils';
	import InstanceRow from './InstanceRow.svelte';
	import Dropdown from './Dropdown.svelte';
	import { compareVersions } from '$lib/semver';
	import {
		Trash2,
		Server,
		Activity,
		HardDrive,
		Cpu,
		Zap,
		Box,
		Terminal,
		Settings,
		ChevronRight,
		ChevronDown,
		RefreshCw,
		Network,
		Radio,
		List,
		Plus,
		Globe
	} from 'lucide-svelte';
	import { slide, fade } from 'svelte/transition';

	export let spawners: Spawner[] = [];
	export let highlightNewSpawnerId: number | null = null;

	let expandedRows: Set<number> = new Set();
	let activeInstances: Record<number, any[]> = {};
	let loadingInstances: Record<number, boolean> = {};

	$: activeVersion = ($serverVersions || []).find((v) => v.is_active);

	const dispatch = createEventDispatcher();

	export function refreshSpawner(id: number) {
		if (expandedRows.has(id)) {
			fetchInstances(id);
		}
	}

	function dispatchBulkAction(action: 'start' | 'stop' | 'restart' | 'update', spawnerId: number) {
		const instances = activeInstances[spawnerId] || [];
		let targetInstances = [];

		if (action === 'start') {
			targetInstances = instances.filter(
				(i) => i.status !== 'Running' && i.status !== 'Provisioning'
			);
		} else if (action === 'stop') {
			targetInstances = instances.filter((i) => i.status === 'Running');
		} else if (action === 'restart') {
			targetInstances = instances.filter((i) => i.status === 'Running');
		} else if (action === 'update') {
			if (!activeVersion) {
				console.error('Bulk Update: No active version found.');
				return;
			}
			targetInstances = instances.filter(
				(i) => !i.version || compareVersions(activeVersion.version, i.version) > 0
			);
		}

		if (targetInstances.length === 0) return;

		dispatch('bulkInstanceActionRequest', {
			action,
			spawnerId,
			instanceIds: targetInstances.map((i) => i.id)
		});
	}

	function toggleRow(id: number) {
		if (expandedRows.has(id)) {
			expandedRows.delete(id);
		} else {
			expandedRows.add(id);
			fetchInstances(id);
		}
		expandedRows = expandedRows;
	}

	async function fetchInstances(id: number) {
		loadingInstances[id] = true;
		try {
			const res = await fetch(`/api/spawners/${id}/instances`);
			if (!res.ok) {
				activeInstances[id] = [];
				return;
			}
			const data = await res.json();
			activeInstances[id] = data.instances || data || [];
			activeInstances = activeInstances;
		} catch (e) {
			console.error('Failed to fetch instances', e);
			activeInstances[id] = [];
		} finally {
			loadingInstances[id] = false;
		}
	}

	function getStatusClass(status: string) {
		switch (status) {
			case 'Online':
				return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
			case 'Degraded':
				return 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30';
			case 'Unresponsive':
				return 'bg-orange-600/10 text-orange-400 border-orange-600/30';
			case 'Offline':
				return 'bg-stone-900/40 text-stone-600 border-stone-800';
			default:
				return 'bg-stone-900/40 text-stone-600 border-stone-800';
		}
	}

	function getInstancePercent(spawner: Spawner) {
		return spawner.max_instances > 0
			? (spawner.current_instances / spawner.max_instances) * 100
			: 0;
	}

	function deleteSpawner(id: number) {
		if (confirm(`Are you sure you want to delete Spawner #${id}? This cannot be undone.`)) {
			dispatch('deleteSpawnerRequest', id);
		}
	}
</script>

<div class="w-full space-y-6">
	{#each spawners as spawner (spawner.id)}
		{@const isExpanded = expandedRows.has(spawner.id)}
		<div
			class="modern-industrial-card glass-panel !rounded-none transition-all duration-500 group {spawner.id ===
			highlightNewSpawnerId && $siteSettings.aesthetic.animations_enabled
				? 'animate-highlight-new-spawner border-rust'
				: 'hover:border-rust/30'} shadow-2xl overflow-hidden"
		>
			<!-- Header / Identity -->
			<div class="p-6 md:p-8 flex flex-col md:flex-row md:items-center justify-between gap-8 bg-stone-950/40 relative">
				<div class="flex items-center gap-6 flex-1 min-w-0">
					<div 
						class="w-16 h-16 bg-stone-950 border border-stone-800 flex items-center justify-center group-hover:border-rust transition-all duration-500 industrial-frame shadow-xl shrink-0"
						onclick={() => toggleRow(spawner.id)}
						role="button"
						tabindex="0"
						onkeydown={(e) => e.key === 'Enter' && toggleRow(spawner.id)}
					>
						<Radio class="w-8 h-8 {spawner.status === 'Online' ? 'text-rust' : 'text-stone-700'}" />
						<div class="absolute bottom-0 left-0 w-full h-1 {spawner.status === 'Online' ? 'bg-rust' : 'bg-stone-800'}"></div>
					</div>

					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-4 flex-wrap mb-2">
							<h3 
								class="text-2xl md:text-3xl font-heading font-black italic uppercase text-white tracking-tighter group-hover:text-rust transition-colors cursor-pointer"
								onclick={() => toggleRow(spawner.id)}
								role="button"
								tabindex="0"
								onkeydown={(e) => e.key === 'Enter' && toggleRow(spawner.id)}
							>
								{spawner.name || spawner.region}
							</h3>
							<span class="text-[9px] bg-stone-900 border border-stone-800 text-stone-500 px-3 py-1 font-black uppercase tracking-widest">UNIT_{spawner.id.toString().padStart(3, '0')}</span>
							<div
								class={`px-3 py-1 font-jetbrains font-bold text-[9px] uppercase flex items-center gap-2 border ${getStatusClass(spawner.status)}`}
							>
								<div class={`w-1.5 h-1.5 rounded-full ${spawner.status === 'Online' ? 'bg-emerald-500 animate-pulse' : 'bg-stone-600'}`}></div>
								{spawner.status}
							</div>
						</div>
						<div class="flex flex-wrap items-center gap-6 font-jetbrains text-[10px] font-bold text-stone-600 uppercase tracking-widest italic">
							<span class="flex items-center gap-2"><Globe class="w-3.5 h-3.5" /> {spawner.region}</span>
							<span class="flex items-center gap-2"><Network class="w-3.5 h-3.5" /> {spawner.host}:{spawner.port}</span>
							<span class="flex items-center gap-2 text-rust-light opacity-60"><Terminal class="w-3.5 h-3.5" /> v{spawner.game_version || '0.0.0'}</span>
						</div>
					</div>
				</div>

				<!-- Quick Actions - ALWAYS ON TOP/RIGHT -->
				<div class="flex items-center gap-3 shrink-0">
					{#if spawner.status !== 'Offline'}
						<a 
							href={`/spawners/${spawner.id}`}
							class="px-6 py-2.5 bg-stone-900 text-stone-400 border border-stone-800 font-heading font-black text-[10px] uppercase tracking-widest hover:bg-white hover:text-black transition-all active:scale-95 shadow-lg flex items-center gap-2"
						>
							<Settings class="w-3.5 h-3.5" />
							Manager
						</a>
						<button 
							onclick={() => dispatch('spawn', spawner.id)} 
							disabled={spawner.status !== 'Online'} 
							class="px-6 py-2.5 bg-rust text-white border border-rust-light font-heading font-black text-[10px] uppercase tracking-widest hover:bg-rust-light transition-all shadow-lg shadow-rust/30 disabled:opacity-20 disabled:grayscale active:scale-95 flex items-center gap-2"
						>
							<Plus class="w-3.5 h-3.5" />
							Spawn
						</button>
					{:else}
						<button 
							onclick={() => deleteSpawner(spawner.id)} 
							class="px-6 py-2.5 bg-red-950/20 text-red-500 border border-red-900/40 font-heading font-black text-[10px] uppercase tracking-widest hover:bg-red-600 hover:text-white transition-all shadow-lg shadow-red-900/10 flex items-center gap-2"
						>
							<Trash2 class="w-3.5 h-3.5" />
							Delete
						</button>
					{/if}
					
					<div class="w-[1px] h-10 bg-stone-800 mx-2 hidden md:block"></div>
					
					<button 
						onclick={() => toggleRow(spawner.id)}
						class="p-3 text-stone-600 hover:text-rust transition-all hidden md:block"
					>
						<ChevronDown class="w-6 h-6 transition-transform duration-500 {isExpanded ? 'rotate-180 text-rust' : ''}" />
					</button>
				</div>
			</div>

			<!-- Quick Telemetry Row -->
			<div class="px-8 py-4 bg-stone-900/20 border-t border-stone-800/50 grid grid-cols-1 md:grid-cols-4 gap-8">
				<div class="flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest italic">Node Capacity</span>
						<span class="text-[10px] font-black text-rust-light tabular-nums">{spawner.current_instances} / {spawner.max_instances}</span>
					</div>
					<div class="h-1.5 bg-stone-950 border border-stone-800 p-0 overflow-hidden shadow-inner">
						<div class="h-full bg-rust transition-all duration-1000 ease-out" style={`width: ${getInstancePercent(spawner)}%`}></div>
					</div>
				</div>

				{#each [
					{ label: 'CPU Load', val: spawner.cpu_usage, icon: Cpu },
					{ label: 'RAM Load', val: spawner.mem_total ? (spawner.mem_used / spawner.mem_total) * 100 : 0, icon: Zap },
					{ label: 'Disk IO', val: spawner.disk_total ? (spawner.disk_used / spawner.disk_total) * 100 : 0, icon: HardDrive }
				] as metric}
					<div class="hidden md:flex flex-col justify-center">
						<div class="flex justify-between items-center mb-2">
							<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest italic flex items-center gap-2">
								<metric.icon class="w-3 h-3 text-stone-700" />
								{metric.label}
							</span>
							<span class="text-[10px] font-black text-stone-400 tabular-nums">{metric.val?.toFixed(1)}%</span>
						</div>
						<div class="h-1 bg-stone-950 border border-stone-800/50 p-0 overflow-hidden">
							<div class="h-full bg-stone-700 transition-all duration-1000 ease-out" style={`width: ${metric.val}%`}></div>
						</div>
					</div>
				{/each}
				
				<button 
					class="md:hidden w-full py-2 text-[9px] font-black text-rust uppercase tracking-[0.2em]"
					onclick={() => toggleRow(spawner.id)}
				>
					{isExpanded ? 'Collapse Telemetry' : 'View Expanded Telemetry'}
				</button>
			</div>

			<!-- Expanded Section -->
			{#if isExpanded}
				<div 
					class="border-t border-stone-800 bg-[#050505]/60 p-8 space-y-10"
					transition:slide={{ duration: 300 }}
				>
					<!-- Metrics Grid -->
					<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">Processor Load</span>
								<Cpu class="w-4 h-4 text-stone-700" />
							</div>
							<div class="text-4xl font-heading font-black text-white tracking-tighter tabular-nums mb-4">{spawner.cpu_usage ? spawner.cpu_usage?.toFixed(1) : 0}%</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-rust shadow-[0_0_10px_rgba(249,115,22,0.4)]" style="width: {spawner.cpu_usage || 0}%"></div>
							</div>
						</div>

						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">Memory Allocation</span>
								<Zap class="w-4 h-4 text-stone-700" />
							</div>
							<div class="text-2xl font-heading font-black text-white tracking-tighter mb-2 tabular-nums">
								{formatBytes(spawner.mem_used || 0)} 
								<span class="text-sm text-stone-600 font-jetbrains">/ {formatBytes(spawner.mem_total || 0)}</span>
							</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-rust-light" style="width: {spawner.mem_total ? (spawner.mem_used / spawner.mem_total) * 100 : 0}%"></div>
							</div>
						</div>

						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">Disk Storage</span>
								<HardDrive class="w-4 h-4 text-stone-700" />
							</div>
							<div class="text-2xl font-heading font-black text-white tracking-tighter mb-2 tabular-nums">
								{formatBytes(spawner.disk_used || 0)} 
								<span class="text-sm text-stone-600 font-jetbrains">/ {formatBytes(spawner.disk_total || 0)}</span>
							</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-stone-600" style="width: {spawner.disk_total ? (spawner.disk_used / spawner.disk_total) * 100 : 0}%"></div>
							</div>
						</div>
					</div>

					<!-- Instance Management -->
					<div class="space-y-6">
						<div class="flex justify-between items-center px-2 border-b border-stone-800 pb-4">
							<div class="flex items-center gap-4">
								<div class="p-2 bg-rust/5 industrial-frame">
									<List class="w-5 h-5 text-rust-light" />
								</div>
								<h4 class="text-sm font-heading font-black text-white uppercase tracking-widest">Active Logic Clusters</h4>
							</div>
							<div class="flex items-center gap-6">
								<Dropdown label="Bulk Operations">
									{#snippet children()}
										<button onclick={() => dispatchBulkAction('start', spawner.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-emerald-400 hover:bg-emerald-500/10 tracking-widest">Execute All</button>
										<button onclick={() => dispatchBulkAction('stop', spawner.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-rose-400 hover:bg-rose-500/10 tracking-widest">Terminate All</button>
										<button onclick={() => dispatchBulkAction('restart', spawner.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-rust hover:bg-rust/10 tracking-widest">Reboot All</button>
										<button onclick={() => dispatchBulkAction('update', spawner.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-amber-400 hover:bg-amber-500/10 border-t border-stone-800 tracking-widest">Update All</button>
									{/snippet}
								</Dropdown>
								<div class="w-px h-8 bg-stone-800"></div>
								<button 
									onclick={() => fetchInstances(spawner.id)} 
									disabled={loadingInstances[spawner.id]}
									class="flex items-center gap-2 font-jetbrains text-[10px] font-black text-rust-light hover:text-white transition-all uppercase tracking-widest disabled:opacity-30"
								>
									{#if loadingInstances[spawner.id]}
										<RefreshCw class="w-3.5 h-3.5 animate-spin" />
									{/if}
									Refresh
								</button>
							</div>
						</div>

						{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
							<div class="py-20 flex flex-col items-center gap-6">
								<div class="w-12 h-12 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
								<p class="font-heading font-black text-[10px] text-stone-600 uppercase tracking-widest animate-pulse">Synchronizing Instances...</p>
							</div>
						{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
							<div class="py-20 text-center opacity-30 border-2 border-stone-800 border-dashed industrial-frame bg-stone-900/20">
								<Box class="w-12 h-12 text-stone-800 mx-auto mb-4" />
								<p class="text-stone-600 font-jetbrains font-black text-[10px] uppercase tracking-widest">No Active Nodes Mapped</p>
							</div>
						{:else}
							<div class={`grid grid-cols-1 gap-3 ${loadingInstances[spawner.id] ? 'opacity-50 pointer-events-none' : ''}`}>
								{#each activeInstances[spawner.id] as instance (instance.id)}
									<InstanceRow
										spawnerId={spawner.id}
										{instance}
										on:tail={(e) => dispatch('tail', e.detail)}
										on:start={(e) => dispatch('startInstanceRequest', e.detail)}
										on:stop={(e) => dispatch('stopInstanceRequest', e.detail)}
										on:restart={(e) => dispatch('restartInstanceRequest', e.detail)}
										on:update={(e) => dispatch('updateInstanceRequest', e.detail)}
										on:rename={(e) => dispatch('renameInstanceRequest', e.detail)}
										on:delete={(e) => dispatch('deleteInstanceRequest', e.detail)}
									/>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="py-40 text-center opacity-40">
			<div class="inline-block p-8 bg-stone-900/40 border border-dashed border-stone-800 industrial-frame mb-8">
				<Server class="w-16 h-16 text-stone-800" />
			</div>
			<h3 class="font-heading font-black text-2xl text-stone-700 uppercase tracking-[0.4em] mb-3">Registry Offline</h3>
			<p class="font-jetbrains text-[11px] font-bold text-stone-600 uppercase tracking-widest">Waiting for nodes to synchronize with master core.</p>
		</div>
	{/each}
</div>

<style lang="ts">
	@keyframes highlight-new-spawner {
		0% {
			background-color: var(--color-rust);
		}
		50% {
			background-color: var(--color-rust);
		}
		100% {
			background-color: transparent;
		}
	}

	.animate-highlight-new-spawner {
		animation: highlight-new-spawner 5s ease-out forwards;
	}

	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}
</style>