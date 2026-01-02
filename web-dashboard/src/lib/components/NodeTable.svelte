<script lang="ts">
	import type { Node } from '$lib/stores';
	import { serverVersions } from '$lib/stores';
	import { createEventDispatcher } from 'svelte';
	import { formatBytes } from '$lib/utils';
	import InstanceRow from './InstanceRow.svelte';
	import Dropdown from './Dropdown.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';
	import CardHoverOverlay from './theme/CardHoverOverlay.svelte';
	import { compareVersions } from '$lib/semver';
	import { slide } from 'svelte/transition';

	let { nodes = [], highlightNewNodeId = null }: { nodes?: Node[], highlightNewNodeId?: number | null } = $props();

	let expandedRows = $state(new Set<number>());
	let hoveredRows = $state<Record<number, boolean>>({});
	let activeInstances = $state<Record<number, any[]>>({});
	let loadingInstances = $state<Record<number, boolean>>({});

	const activeVersion = $derived(($serverVersions || []).find((v) => v.is_active));

	const dispatch = createEventDispatcher();

	export function refreshNode(id: number) {
		if (expandedRows.has(id)) {
			fetchInstances(id);
		}
	}

	function dispatchBulkAction(action: 'start' | 'stop' | 'restart' | 'update', nodeId: number) {
		const instances = activeInstances[nodeId] || [];
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
			nodeId,
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
	}

	async function fetchInstances(id: number) {
		loadingInstances[id] = true;
		try {
			const res = await fetch(`/api/nodes/${id}/instances`);
			if (!res.ok) {
				activeInstances[id] = [];
				return;
			}
			const data = await res.json();
			activeInstances[id] = data.instances || data || [];
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
				return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/30';
			case 'Degraded':
			case 'Unresponsive':
				return 'bg-amber-500/10 text-amber-500 border-amber-500/30';
			case 'Offline':
				return 'bg-stone-900/40 text-stone-600 border-stone-800';
			default:
				return 'bg-stone-900/40 text-stone-600 border-stone-800';
		}
	}

	function getInstancePercent(node: Node) {
		return node.max_instances > 0
			? (node.current_instances / node.max_instances) * 100
			: 0;
	}

	function deleteNode(id: number) {
		if (confirm(`Are you sure you want to delete Node #${id}? This cannot be undone.`)) {
			dispatch('deleteNodeRequest', id);
		}
	}
</script>

<div class="w-full space-y-6">
	{#each nodes as node (node.id)}
		{@const isExpanded = expandedRows.has(node.id)}
		<div
			class="modern-industrial-card glass-panel tactical-border !rounded-none transition-all duration-500 group {node.id ===
			highlightNewNodeId
				? 'animate-highlight-new-node border-rust'
				: 'hover:border-rust/30'} shadow-2xl overflow-hidden"
			onmouseenter={() => hoveredRows[node.id] = true}
			onmouseleave={() => hoveredRows[node.id] = false}
		>
			<!-- Tactical Corners -->
			<div class="corner-tl"></div>
			<div class="corner-tr"></div>
			<div class="corner-bl"></div>
			<div class="corner-br"></div>
			<!-- Hover Intelligence Overlay -->
			<CardHoverOverlay active={hoveredRows[node.id]} />

			<!-- Header / Identity -->
			<div 
				class="p-6 md:p-8 flex flex-col md:flex-row md:items-center justify-between gap-8 relative bg-black/40"
			>
				<div class="flex items-center gap-6 flex-1 min-w-0">
					<div 
						class="w-16 h-16 bg-stone-950 border border-stone-800 flex items-center justify-center group-hover:border-rust transition-all duration-500 industrial-sharp shadow-xl shrink-0"
						onclick={() => toggleRow(node.id)}
						role="button"
						tabindex="0"
						onkeydown={(e) => e.key === 'Enter' && toggleRow(node.id)}
					>
						<Icon name="radio" size="2rem" class="{node.status === 'Online' ? 'text-rust' : 'text-stone-700'}" />
						<div class="absolute bottom-0 left-0 w-full h-1 {node.status === 'Online' ? 'bg-rust' : 'bg-stone-800'}"></div>
					</div>

					<div class="min-w-0 flex-1">
							<div class="flex items-center gap-4 flex-wrap mb-2">
								<h3>
									<button 
										class="text-2xl md:text-3xl font-heading font-black italic uppercase text-white tracking-tighter group-hover:text-rust transition-colors cursor-pointer bg-transparent border-none p-0 text-left"
										onclick={() => toggleRow(node.id)}
									>
										{node.name || node.region}
									</button>
								</h3>
							<span class="text-[9px] bg-stone-900 border border-stone-800 text-stone-500 px-3 py-1 font-black uppercase tracking-widest">Node {node.id.toString().padStart(3, '0')}</span>
							<div
								class={`px-3 py-1 font-jetbrains font-bold text-[9px] uppercase flex items-center gap-2 border ${getStatusClass(node.status)}`}
							>
								<div class={`w-1.5 h-1.5 rounded-full ${node.status === 'Online' ? 'bg-emerald-500 animate-pulse' : 'bg-stone-600'}`}></div>
								{node.status}
							</div>
						</div>
						<div class="flex flex-wrap items-center gap-6 font-jetbrains text-[10px] font-bold text-stone-600 uppercase tracking-widest italic">
							<span class="flex items-center gap-2"><Icon name="ph:globe-bold" size="0.875rem" /> {node.region}</span>
							<span class="flex items-center gap-2"><Icon name="ph:network-bold" size="0.875rem" /> {node.host}:{node.port}</span>
							<span class="flex items-center gap-2 text-rust-light opacity-60"><Icon name="ph:cpu-bold" size="0.875rem" /> v{node.game_version || '0.0.0'}</span>
						</div>
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="flex items-center gap-3 shrink-0">
					{#if node.status !== 'Offline'}
						<div class="flex items-center gap-1 bg-black/20 p-1 border border-stone-800 mr-2">
							<button 
								onclick={() => dispatch('updateNodeBuild', node.id)}
								class="p-2 text-emerald-500 hover:bg-emerald-500/10 transition-all"
								title="Update Game Build"
							>
								<Icon name="ph:arrow-down-to-line-bold" size="1rem" />
							</button>
							<button 
								class="p-2 text-stone-500 hover:bg-stone-500/10 transition-all opacity-30"
								title="Downgrade Build (N/A)"
								disabled
							>
								<Icon name="ph:arrow-counter-clockwise-bold" size="1rem" />
							</button>
						</div>
						<Button 
							href={`/nodes/${node.id}`}
							variant="secondary"
							size="sm"
							icon="ph:gear-bold"
						>
							Manage
						</Button>
						<Button 
							onclick={() => dispatch('spawn', node.id)} 
							disabled={node.status !== 'Online'} 
							variant="primary"
							size="sm"
							icon="ph:plus-bold"
						>
							Spawn
						</Button>
					{:else}
						<Button 
							onclick={() => deleteNode(node.id)} 
							variant="danger"
							size="sm"
							icon="ph:trash-bold"
						>
							Delete
						</Button>
					{/if}
					
					<div class="w-[1px] h-10 bg-stone-800 mx-2 hidden md:block"></div>
					
					<button 
						onclick={() => toggleRow(node.id)}
						class="p-3 text-stone-600 hover:text-rust transition-all hidden md:block"
					>
						<Icon name="ph:caret-down-bold" size="1.5rem" class="transition-transform duration-500 {isExpanded ? 'rotate-180 text-rust' : ''}" />
					</button>
				</div>
			</div>

			<!-- Quick Telemetry Row -->
			<div class="px-8 py-4 bg-stone-900/20 border-t border-stone-800/50 grid grid-cols-1 md:grid-cols-4 gap-8">
				<div class="flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-black uppercase tracking-widest italic text-stone-500">Capacity</span>
						<span class="text-[10px] font-black text-rust-light tabular-nums">{node.current_instances} / {node.max_instances}</span>
					</div>
					<div class="h-1.5 bg-stone-950 border border-stone-800 p-0 overflow-hidden shadow-inner">
						<div class="h-full bg-rust transition-all duration-1000 ease-out" style={`width: ${getInstancePercent(node)}%`}></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-black uppercase tracking-widest italic flex items-center gap-2 text-stone-500">
							<Icon name="ph:cpu-bold" size="0.75rem" class="text-stone-700" />
							CPU Usage
						</span>
						<span class="text-[10px] font-black text-stone-400 tabular-nums">{node.cpu_usage?.toFixed(1)}%</span>
					</div>
					<div class="h-1 bg-stone-950 border border-stone-800/50 p-0 overflow-hidden">
						<div class="h-full bg-stone-700 transition-all duration-1000 ease-out" style={`width: ${node.cpu_usage}%`}></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-black uppercase tracking-widest italic flex items-center gap-2 text-stone-500">
							<Icon name="ph:activity-bold" size="0.75rem" class="text-stone-700" />
							RAM Usage
						</span>
						<span class="text-[10px] font-black text-stone-400 tabular-nums">{node.mem_total ? ((node.mem_used / node.mem_total) * 100).toFixed(1) : 0}%</span>
					</div>
					<div class="h-1 bg-stone-950 border border-stone-800/50 p-0 overflow-hidden">
						<div class="h-full bg-stone-700 transition-all duration-1000 ease-out" style={`width: ${node.mem_total ? (node.mem_used / node.mem_total) * 100 : 0}%`}></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-black uppercase tracking-widest italic flex items-center gap-2 text-stone-500">
							<Icon name="ph:hard-drive-bold" size="0.75rem" class="text-stone-700" />
							Disk IO
						</span>
						<span class="text-[10px] font-black text-stone-400 tabular-nums">{node.disk_total ? ((node.disk_used / node.disk_total) * 100).toFixed(1) : 0}%</span>
					</div>
					<div class="h-1 bg-stone-950 border border-stone-800/50 p-0 overflow-hidden">
						<div class="h-full bg-stone-700 transition-all duration-1000 ease-out" style={`width: ${node.disk_total ? (node.disk_used / node.disk_total) * 100 : 0}%`}></div>
					</div>
				</div>
				
				<button 
					class="md:hidden w-full py-2 text-[9px] font-black text-rust uppercase tracking-[0.2em]"
					onclick={() => toggleRow(node.id)}
				>
					{isExpanded ? 'Collapse Stats' : 'View Stats'}
				</button>
			</div>

			<!-- Expanded Section -->
			{#if isExpanded}
				<div 
					class="border-t border-stone-800 bg-black/60 p-8 space-y-10"
					transition:slide={{ duration: 300 }}
				>
					<!-- Metrics Grid -->
					<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-sharp shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">CPU LOAD</span>
								<Icon name="ph:cpu-bold" size="1rem" class="text-stone-700" />
							</div>
							<div class="text-4xl font-heading font-black text-white tracking-tighter tabular-nums mb-4">{node.cpu_usage ? node.cpu_usage?.toFixed(1) : 0}%</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-rust shadow-[0_0_10px_rgba(249,115,22,0.4)]" style="width: {node.cpu_usage || 0}%"></div>
							</div>
						</div>

						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-sharp shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">RAM USAGE</span>
								<Icon name="ph:activity-bold" size="1rem" class="text-stone-700" />
							</div>
							<div class="text-2xl font-heading font-black text-white tracking-tighter mb-2 tabular-nums">
								{formatBytes(node.mem_used || 0)} 
								<span class="text-sm text-stone-600 font-jetbrains">/ {formatBytes(node.mem_total || 0)}</span>
							</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-rust-light" style="width: {node.mem_total ? (node.mem_used / node.mem_total) * 100 : 0}%"></div>
							</div>
						</div>

						<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-sharp shadow-xl group/m hover:border-rust/30 transition-all">
							<div class="flex justify-between items-center mb-4">
								<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest italic group-hover/m:text-rust transition-colors">DISK USAGE</span>
								<Icon name="ph:hard-drive-bold" size="1rem" class="text-stone-700" />
							</div>
							<div class="text-2xl font-heading font-black text-white tracking-tighter mb-2 tabular-nums">
								{formatBytes(node.disk_used || 0)} 
								<span class="text-sm text-stone-600 font-jetbrains">/ {formatBytes(node.disk_total || 0)}</span>
							</div>
							<div class="h-1.5 bg-stone-950 border border-stone-800 shadow-inner overflow-hidden">
								<div class="h-full bg-stone-600" style="width: {node.disk_total ? (node.disk_used / node.disk_total) * 100 : 0}%"></div>
							</div>
						</div>
					</div>

					<!-- Instance Management -->
					<div class="space-y-6">
						<div class="flex justify-between items-center px-2 border-b border-stone-800 pb-4">
							<div class="flex items-center gap-4">
								<div class="p-2 bg-rust/5 industrial-sharp border border-rust/20">
									<Icon name="ph:list-bullets-bold" size="1.25rem" class="text-rust-light" />
								</div>
								<h4 class="text-sm font-heading font-black text-white uppercase tracking-widest">Active Instances</h4>
							</div>
							<div class="flex items-center gap-6">
								<Dropdown label="Bulk Operations">
									{#snippet children()}
										<button onclick={() => dispatchBulkAction('start', node.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-emerald-400 hover:bg-emerald-500/10 tracking-widest">Start All</button>
										<button onclick={() => dispatchBulkAction('stop', node.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-rose-400 hover:bg-rose-500/10 tracking-widest">Stop All</button>
										<button onclick={() => dispatchBulkAction('restart', node.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-rust hover:bg-rust/10 tracking-widest">Restart All</button>
										<button onclick={() => dispatchBulkAction('update', node.id)} class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains uppercase text-amber-400 hover:bg-amber-500/10 border-t border-stone-800 tracking-widest">Update All</button>
									{/snippet}
								</Dropdown>
								<div class="w-px h-8 bg-stone-800"></div>
								<button 
									onclick={() => fetchInstances(node.id)} 
									disabled={loadingInstances[node.id]}
									class="flex items-center gap-2 font-jetbrains text-[10px] font-black text-rust-light hover:text-white transition-all uppercase tracking-widest disabled:opacity-30"
								>
									{#if loadingInstances[node.id]}
										<Icon name="ph:arrows-clockwise-bold" size="0.875rem" class="animate-spin" />
									{/if}
									Refresh
								</button>
							</div>
						</div>

						{#if loadingInstances[node.id] && (!activeInstances[node.id] || activeInstances[node.id].length === 0)}
							<div class="py-20 flex flex-col items-center gap-6">
								<div class="w-12 h-12 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
								<p class="font-heading font-black text-[10px] text-stone-600 uppercase tracking-widest animate-pulse">Synchronizing Instances...</p>
							</div>
						{:else if !activeInstances[node.id] || activeInstances[node.id].length === 0}
							<div class="py-20 text-center opacity-30 border-2 border-stone-800 border-dashed industrial-sharp bg-stone-900/20">
								<Icon name="ph:cube-bold" size="3rem" class="text-stone-800 mx-auto mb-4" />
								<p class="text-stone-600 font-jetbrains font-black text-[10px] uppercase tracking-widest">No Active Nodes Mapped</p>
							</div>
						{:else}
							<div class={`grid grid-cols-1 gap-3 ${loadingInstances[node.id] ? 'opacity-50 pointer-events-none' : ''}`}>
								{#each activeInstances[node.id] as instance (instance.id)}
									<InstanceRow
										nodeId={node.id}
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
			<div class="inline-block p-8 bg-stone-900/40 border border-dashed border-stone-800 industrial-sharp mb-8">
				<Icon name="ph:server-bold" size="4rem" class="text-stone-800" />
			</div>
			<h3 class="font-heading font-black text-2xl text-stone-700 uppercase tracking-[0.4em] mb-3">No Nodes Active</h3>
			<p class="font-jetbrains text-[11px] font-bold text-stone-600 uppercase tracking-widest">Waiting for nodes to synchronize with controller.</p>
		</div>
	{/each}
</div>

<style lang="ts">
	@keyframes highlight-new-node {
		0% {
			background-color: #c2410c;
		}
		50% {
			background-color: #c2410c;
		}
		100% {
			background-color: transparent;
		}
	}

	.animate-highlight-new-node {
		animation: highlight-new-node 5s ease-out forwards;
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
		background: #c2410c;
	}
</style>