<script lang="ts">
	import { apiFetch } from '$lib/api';
	import type { Node } from '$lib/stores.svelte';
	import { serverVersions } from '$lib/stores.svelte';
	import { createEventDispatcher } from 'svelte';
	import { formatBytes } from '$lib/utils';
	import InstanceRow from './InstanceRow.svelte';
	import Dropdown from './Dropdown.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';
	import { compareVersions } from '$lib/semver';
	import { slide, fade } from 'svelte/transition';

	let {
		nodes = [],
		highlightNewNodeId = null
	}: { nodes?: Node[]; highlightNewNodeId?: number | null } = $props();

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
			const res = await apiFetch(`/api/nodes/${id}/instances`);
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
				return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
			case 'Degraded':
			case 'Unresponsive':
				return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
			case 'Offline':
				return 'bg-black/40 text-slate-600 border-white/5';
			default:
				return 'bg-black/40 text-slate-600 border-white/5';
		}
	}

	function getInstancePercent(node: Node) {
		return node.max_instances > 0 ? (node.current_instances / node.max_instances) * 100 : 0;
	}

	function deleteNode(id: number) {
		if (confirm(`Are you sure you want to delete Node #${id}? This cannot be undone.`)) {
			dispatch('deleteNodeRequest', id);
		}
	}
</script>

<div class="w-full space-y-6 font-sans">
	{#each nodes as node (node.id)}
		{@const isExpanded = expandedRows.has(node.id)}
		<div
			class="bg-slate-900/40 border border-white/5 transition-all duration-500 group {node.id ===
			highlightNewNodeId
				? 'animate-highlight-new-node border-sky-500'
				: 'hover:border-sky-500/30'} shadow-2xl overflow-hidden rounded-3xl"
			onmouseenter={() => (hoveredRows[node.id] = true)}
			onmouseleave={() => (hoveredRows[node.id] = false)}
			role="article"
		>
			<!-- Header / Identity -->
			<div
				class="p-6 md:p-8 flex flex-col md:flex-row md:items-center justify-between gap-8 relative bg-black/20"
			>
				<div class="flex items-center gap-6 flex-1 min-w-0">
					<button
						class="w-16 h-16 bg-black border border-white/10 flex items-center justify-center group-hover:border-sky-500 transition-all duration-500 rounded-2xl shadow-xl shrink-0 cursor-pointer"
						onclick={() => toggleRow(node.id)}
						aria-expanded={isExpanded}
					>
						<Icon
							name="ph:cpu-bold"
							size="2rem"
							class={node.status === 'Online' ? 'text-sky-400' : 'text-slate-700'}
						/>
					</button>

					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-4 flex-wrap mb-2">
							<h3>
								<button
									class="text-2xl md:text-3xl font-bold text-white tracking-tight group-hover:text-sky-400 transition-colors cursor-pointer bg-transparent border-none p-0 text-left font-heading italic uppercase"
									onclick={() => toggleRow(node.id)}
								>
									{node.name || node.region}
								</button>
							</h3>
							<span
								class="text-[9px] bg-black border border-white/10 text-slate-500 px-3 py-1 font-bold uppercase tracking-widest rounded-lg"
								>ID: {node.id.toString().padStart(3, '0')}</span
							>
							<div
								class={`px-3 py-1 font-bold text-[9px] uppercase flex items-center gap-2 border rounded-lg ${getStatusClass(node.status)}`}
							>
								<div
									class={`w-1.5 h-1.5 rounded-full ${node.status === 'Online' ? 'bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'bg-slate-600'}`}
								></div>
								{node.status}
							</div>
						</div>
						<div
							class="flex flex-wrap items-center gap-6 text-[10px] font-bold text-slate-500 uppercase tracking-wider"
						>
							<span class="flex items-center gap-2"
								><Icon name="ph:globe-bold" size="0.875rem" class="text-slate-600" />
								{node.region}</span
							>
							<span class="flex items-center gap-2"
								><Icon name="ph:network-bold" size="0.875rem" class="text-slate-600" />
								{node.host}:{node.port}</span
							>
							<span class="flex items-center gap-2 text-sky-500/60"
								><Icon name="ph:info-bold" size="0.875rem" /> v{node.game_version || '0.0.0'}</span
							>
						</div>
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="flex items-center gap-3 shrink-0">
					{#if node.status !== 'Offline'}
						<div
							class="flex items-center gap-1 bg-black/40 p-1 border border-white/5 rounded-xl mr-2"
						>
							<button
								onclick={() => dispatch('updateNodeBuild', node.id)}
								class="p-2 text-emerald-400 hover:bg-emerald-500/10 transition-all rounded-lg"
								title="Update Game Build"
							>
								<Icon name="ph:arrow-down-to-line-bold" size="1rem" />
							</button>
						</div>
						<Button
							href={`/nodes/${node.id}`}
							variant="secondary"
							size="sm"
							icon="ph:gear-bold"
							class="!rounded-xl"
						>
							Manage
						</Button>
						<Button
							onclick={() => dispatch('spawn', node.id)}
							disabled={node.status !== 'Online'}
							variant="primary"
							size="sm"
							icon="ph:plus-bold"
							class="!rounded-xl shadow-lg shadow-sky-500/10"
						>
							Provision Instance
						</Button>
					{:else}
						<Button
							onclick={() => deleteNode(node.id)}
							variant="danger"
							size="sm"
							icon="ph:trash-bold"
							class="!rounded-xl"
						>
							Remove
						</Button>
					{/if}

					<div class="w-[1px] h-10 bg-white/5 mx-2 hidden md:block"></div>

					<button
						onclick={() => toggleRow(node.id)}
						class="p-3 text-slate-600 hover:text-sky-400 transition-all hidden md:block"
					>
						<Icon
							name="ph:caret-down-bold"
							size="1.5rem"
							class="transition-transform duration-500 {isExpanded
								? 'rotate-180 text-sky-400'
								: ''}"
						/>
					</button>
				</div>
			</div>

			<!-- Quick Telemetry Row -->
			<div
				class="px-8 py-4 bg-black/40 border-t border-white/5 grid grid-cols-1 md:grid-cols-4 gap-8"
			>
				<div class="flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span class="text-[9px] font-bold uppercase tracking-widest text-slate-500"
							>Memory Load</span
						>
						<span class="text-[10px] font-bold text-sky-400 tabular-nums"
							>{node.current_instances} / {node.max_instances}</span
						>
					</div>
					<div
						class="h-1.5 bg-black border border-white/5 rounded-full overflow-hidden shadow-inner"
					>
						<div
							class="h-full bg-sky-500 transition-all duration-1000 ease-out shadow-[0_0_10px_rgba(14,165,233,0.3)]"
							style={`width: ${getInstancePercent(node)}%`}
						></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span
							class="text-[9px] font-bold uppercase tracking-widest flex items-center gap-2 text-slate-500"
						>
							<Icon name="ph:cpu-bold" size="0.75rem" class="text-slate-700" />
							CPU Usage
						</span>
						<span class="text-[10px] font-bold text-slate-400 tabular-nums"
							>{node.cpu_usage?.toFixed(1)}%</span
						>
					</div>
					<div
						class="h-1.5 bg-black border border-white/5 rounded-full overflow-hidden shadow-inner"
					>
						<div
							class="h-full bg-slate-700 transition-all duration-1000 ease-out"
							style={`width: ${node.cpu_usage}%`}
						></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span
							class="text-[9px] font-bold uppercase tracking-widest flex items-center gap-2 text-slate-500"
						>
							<Icon name="ph:activity-bold" size="0.75rem" class="text-slate-700" />
							RAM Usage
						</span>
						<span class="text-[10px] font-bold text-slate-400 tabular-nums"
							>{node.mem_total ? ((node.mem_used / node.mem_total) * 100).toFixed(1) : 0}%</span
						>
					</div>
					<div
						class="h-1.5 bg-black border border-white/5 rounded-full overflow-hidden shadow-inner"
					>
						<div
							class="h-full bg-slate-700 transition-all duration-1000 ease-out"
							style={`width: ${node.mem_total ? (node.mem_used / node.mem_total) * 100 : 0}%`}
						></div>
					</div>
				</div>

				<div class="hidden md:flex flex-col justify-center">
					<div class="flex justify-between items-center mb-2">
						<span
							class="text-[9px] font-bold uppercase tracking-widest flex items-center gap-2 text-slate-500"
						>
							<Icon name="ph:hard-drive-bold" size="0.75rem" class="text-slate-700" />
							Storage
						</span>
						<span class="text-[10px] font-bold text-slate-400 tabular-nums"
							>{node.disk_total ? ((node.disk_used / node.disk_total) * 100).toFixed(1) : 0}%</span
						>
					</div>
					<div
						class="h-1.5 bg-black border border-white/5 rounded-full overflow-hidden shadow-inner"
					>
						<div
							class="h-full bg-slate-700 transition-all duration-1000 ease-out"
							style={`width: ${node.disk_total ? (node.disk_used / node.disk_total) * 100 : 0}%`}
						></div>
					</div>
				</div>

				<button
					class="md:hidden w-full py-2 text-[9px] font-bold text-sky-500 uppercase tracking-widest"
					onclick={() => toggleRow(node.id)}
				>
					{isExpanded ? 'Collapse Stats' : 'View Stats'}
				</button>
			</div>

			<!-- Expanded Section -->
			{#if isExpanded}
				<div
					class="border-t border-white/5 bg-black/20 p-8 space-y-10"
					transition:slide={{ duration: 300 }}
				>
					<!-- Metrics Grid -->
					<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
						<div
							class="bg-black/40 border border-white/5 p-6 rounded-2xl shadow-xl group/m hover:border-sky-500/30 transition-all backdrop-blur-md shadow-inner"
						>
							<div class="flex justify-between items-center mb-4">
								<span
									class="text-[10px] font-bold text-slate-500 uppercase tracking-widest group-hover/m:text-sky-400 transition-colors"
									>CPU LOAD</span
								>
								<Icon name="ph:cpu-bold" size="1rem" class="text-slate-700" />
							</div>
							<div class="text-4xl font-bold text-white tracking-tight tabular-nums mb-4">
								{node.cpu_usage ? node.cpu_usage?.toFixed(1) : 0}%
							</div>
							<div
								class="h-1.5 bg-black border border-white/5 rounded-full shadow-inner overflow-hidden"
							>
								<div
									class="h-full bg-sky-500 shadow-[0_0_10px_rgba(14,165,233,0.4)]"
									style="width: {node.cpu_usage || 0}%"
								></div>
							</div>
						</div>

						<div
							class="bg-black/40 border border-white/5 p-6 rounded-2xl shadow-xl group/m hover:border-sky-500/30 transition-all backdrop-blur-md shadow-inner"
						>
							<div class="flex justify-between items-center mb-4">
								<span
									class="text-[10px] font-bold text-slate-500 uppercase tracking-widest group-hover/m:text-sky-400 transition-colors"
									>MEMORY USAGE</span
								>
								<Icon name="ph:activity-bold" size="1rem" class="text-slate-700" />
							</div>
							<div class="text-2xl font-bold text-white tracking-tight mb-2 tabular-nums">
								{formatBytes(node.mem_used || 0)}
								<span class="text-sm text-slate-600 font-medium"
									>/ {formatBytes(node.mem_total || 0)}</span
								>
							</div>
							<div
								class="h-1.5 bg-black border border-white/5 rounded-full shadow-inner overflow-hidden"
							>
								<div
									class="h-full bg-sky-500/60"
									style="width: {node.mem_total ? (node.mem_used / node.mem_total) * 100 : 0}%"
								></div>
							</div>
						</div>

						<div
							class="bg-black/40 border border-white/5 p-6 rounded-2xl shadow-xl group/m hover:border-sky-500/30 transition-all backdrop-blur-md shadow-inner"
						>
							<div class="flex justify-between items-center mb-4">
								<span
									class="text-[10px] font-bold text-slate-500 uppercase tracking-widest group-hover/m:text-sky-400 transition-colors"
									>DISK USAGE</span
								>
								<Icon name="ph:hard-drive-bold" size="1rem" class="text-slate-700" />
							</div>
							<div class="text-2xl font-bold text-white tracking-tight mb-2 tabular-nums">
								{formatBytes(node.disk_used || 0)}
								<span class="text-sm text-slate-600 font-medium"
									>/ {formatBytes(node.disk_total || 0)}</span
								>
							</div>
							<div
								class="h-1.5 bg-black border border-white/5 rounded-full shadow-inner overflow-hidden"
							>
								<div
									class="h-full bg-slate-700"
									style="width: {node.disk_total ? (node.disk_used / node.disk_total) * 100 : 0}%"
								></div>
							</div>
						</div>
					</div>

					<!-- Instance Management -->
					<div class="space-y-6">
						<div class="flex justify-between items-center px-2 border-b border-white/5 pb-4">
							<div class="flex items-center gap-4">
								<div class="p-2 bg-sky-500/5 rounded-xl border border-sky-500/20">
									<Icon name="ph:list-bullets-bold" size="1.25rem" class="text-sky-400" />
								</div>
								<h4 class="text-sm font-bold text-white uppercase tracking-widest font-heading italic">
									Active Instances
								</h4>
							</div>
							<div class="flex items-center gap-6">
								<Dropdown label="Actions">
									{#snippet children()}
										<button
											onclick={() => dispatchBulkAction('start', node.id)}
											class="w-full text-left px-6 py-3 text-[10px] font-bold uppercase text-emerald-400 hover:bg-emerald-500/10 tracking-widest"
											>Start All</button
										>
										<button
											onclick={() => dispatchBulkAction('stop', node.id)}
											class="w-full text-left px-6 py-3 text-[10px] font-bold uppercase text-rose-400 hover:bg-rose-500/10 tracking-widest"
											>Stop All</button
										>
										<button
											onclick={() => dispatchBulkAction('restart', node.id)}
											class="w-full text-left px-6 py-3 text-[10px] font-bold uppercase text-sky-400 hover:bg-sky-500/10 tracking-widest"
											>Restart All</button
										>
										<button
											onclick={() => dispatchBulkAction('update', node.id)}
											class="w-full text-left px-6 py-3 text-[10px] font-bold uppercase text-amber-400 hover:bg-amber-500/10 border-t border-white/5 tracking-widest"
											>Update All</button
										>
									{/snippet}
								</Dropdown>
								<div class="w-px h-8 bg-white/5"></div>
								<button
									onclick={() => fetchInstances(node.id)}
									disabled={loadingInstances[node.id]}
									class="flex items-center gap-2 font-bold text-[10px] text-slate-500 hover:text-white transition-all uppercase tracking-widest disabled:opacity-30"
								>
									{#if loadingInstances[node.id]}
										<Icon name="ph:arrows-clockwise-bold" size="0.875rem" class="animate-spin" />
									{/if}
									Refresh
								</button>
							</div>
						</div>

						{#if loadingInstances[node.id] && (!activeInstances[node.id] || activeInstances[node.id].length === 0)}
							<div class="py-20 flex flex-col items-center justify-center gap-6">
								<div
									class="w-12 h-12 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"
								></div>
								<p
									class="font-bold text-[10px] text-slate-500 uppercase tracking-widest animate-pulse"
								>
									Loading instances...
								</p>
							</div>
						{:else if !activeInstances[node.id] || activeInstances[node.id].length === 0}
							<div
								class="py-20 text-center opacity-30 border-2 border-slate-800 border-dashed rounded-3xl bg-black/20"
							>
								<Icon name="ph:cube-bold" size="3rem" class="text-slate-800 mx-auto mb-4" />
								<p class="text-slate-600 font-bold text-[10px] uppercase tracking-widest">
									No instances found
								</p>
							</div>
						{:else}
							<div
								class={`grid grid-cols-1 gap-3 ${loadingInstances[node.id] ? 'opacity-50 pointer-events-none' : ''}`}
							>
								{#each activeInstances[node.id] as instance (instance.id)}
									<InstanceRow
										nodeId={node.id}
										{instance}
										on:tail={(e: any) => dispatch('tail', e.detail)}
										on:start={(e: any) => dispatch('startInstanceRequest', e.detail)}
										on:stop={(e: any) => dispatch('stopInstanceRequest', e.detail)}
										on:restart={(e: any) => dispatch('restartInstanceRequest', e.detail)}
										on:update={(e: any) => dispatch('updateInstanceRequest', e.detail)}
										on:rename={(e: any) => dispatch('renameInstanceRequest', e.detail)}
										on:delete={(e: any) => dispatch('deleteInstanceRequest', e.detail)}
									/>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="py-40 text-center opacity-40" in:fade>
			<div
				class="inline-block p-8 bg-slate-900/40 border-2 border-dashed border-slate-800 rounded-3xl mb-8"
			>
				<Icon name="ph:server-bold" size="4rem" class="text-slate-800" />
			</div>
			<h3 class="font-bold text-2xl text-slate-700 uppercase tracking-widest mb-3">
				No Nodes Active
			</h3>
			<p class="text-[11px] font-bold text-slate-600 uppercase tracking-widest">
				Waiting for nodes to synchronize with controller.
			</p>
		</div>
	{/each}
</div>

<style lang="ts">
	@keyframes highlight-new-node {
		0% {
			background-color: rgba(14, 165, 233, 0.2);
		}
		50% {
			background-color: rgba(14, 165, 233, 0.2);
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
		background: #1e293b;
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #334155;
	}
</style>
