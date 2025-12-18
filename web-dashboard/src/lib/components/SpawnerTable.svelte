<script lang="ts">
	import type { Spawner } from '$lib/stores';
	import { serverVersions } from '$lib/stores';
	import { createEventDispatcher } from 'svelte';
	import { formatBytes } from '$lib/utils';
	import InstanceRow from './InstanceRow.svelte';
	import Dropdown from './Dropdown.svelte';
	import { compareVersions } from '$lib/semver';
	import { Trash2, MoreHorizontal, Server, Activity, HardDrive, Cpu, Zap, Box, Terminal, Settings } from 'lucide-svelte';
	import { slide } from 'svelte/transition';

	export let spawners: Spawner[] = [];
	export let highlightNewSpawnerId: number | null = null;

	let expandedRows: Set<number> = new Set();
	let activeInstances: Record<number, any[]> = {};
	let loadingInstances: Record<number, boolean> = {};

	$: activeVersion = ($serverVersions || []).find((v) => v.is_active);

	function getOutdatedCount(spawnerId: number) {
		if (!activeInstances[spawnerId] || !activeVersion) return 0;
		return activeInstances[spawnerId].filter(
			(i) => i.version && compareVersions(activeVersion.version, i.version) > 0
		).length;
	}

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
			if (res.ok) {
				const data = await res.json();
				activeInstances[id] = data.instances || data || [];
				activeInstances = activeInstances;
			} else {
				activeInstances[id] = [];
				activeInstances = activeInstances;
			}
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
				return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
			case 'Degraded':
				return 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/20';
			case 'Unresponsive':
				return 'bg-orange-500/10 text-orange-400 border border-orange-500/20';
			case 'Offline':
				return 'bg-slate-700/30 text-slate-400 border border-slate-700/50';
			default:
				return 'bg-slate-700/30 text-slate-400';
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

<div class="w-full">
	<!-- Mobile View (Cards) -->
	<div class="md:hidden space-y-4">
		{#each spawners as spawner}
			<div class="bg-slate-800/50 border border-slate-700/50 rounded-xl overflow-hidden {spawner.id === highlightNewSpawnerId ? 'ring-2 ring-blue-500' : ''}">
				<!-- Card Header -->
				<div class="p-4 border-b border-slate-700/50 flex items-center justify-between" onclick={() => toggleRow(spawner.id)}>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-slate-800 rounded-lg border border-slate-700">
							<Server class="w-5 h-5 text-slate-400" />
						</div>
						<div>
							<div class="flex items-center gap-2">
								<span class="font-bold text-slate-100">{spawner.region}</span>
								<span class="text-xs text-slate-500">#{spawner.id}</span>
							</div>
							<div class="text-xs text-slate-400 font-mono mt-0.5">
								{spawner.host}:{spawner.port}
							</div>
						</div>
					</div>
					<div class="flex flex-col items-end gap-1">
						<span class={`text-[10px] px-2 py-0.5 rounded-full font-semibold border ${getStatusClass(spawner.status)}`}>
							{spawner.status}
						</span>
						{#if spawner.game_version}
							<span class="text-[10px] text-slate-500 bg-slate-800 px-1.5 py-0.5 rounded">v{spawner.game_version}</span>
						{/if}
					</div>
				</div>

				<!-- Card Body -->
				<div class="p-4 space-y-4">
					<!-- Instances Progress -->
					<div>
						<div class="flex justify-between text-xs mb-1.5">
							<span class="text-slate-400">Instances</span>
							<span class="text-slate-200 font-medium">{spawner.current_instances} / {spawner.max_instances}</span>
						</div>
						<div class="w-full h-2 bg-slate-700/50 rounded-full overflow-hidden">
							<div
								class="h-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-500"
								style={`width: ${getInstancePercent(spawner)}%`}
							></div>
						</div>
					</div>

					<!-- Actions Grid -->
					<div class="grid grid-cols-3 gap-2">
						{#if spawner.status === 'Offline'}
							<button
								onclick={() => deleteSpawner(spawner.id)}
								class="col-span-3 flex items-center justify-center gap-2 py-2 px-3 bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 rounded-lg text-xs font-medium transition-colors"
							>
								<Trash2 class="w-3.5 h-3.5" />
								Delete Spawner
							</button>
						{:else}
							<a
								href="/spawners/{spawner.id}"
								class="flex items-center justify-center gap-1.5 py-2 px-3 bg-slate-700/50 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-medium transition-colors border border-slate-600/50"
							>
								<Settings class="w-3.5 h-3.5" /> <!-- Use standard icon, import needed if unused -->
								Manage
							</a>
							<button
								onclick={() => dispatch('viewLogs', spawner.id)}
								class="flex items-center justify-center gap-1.5 py-2 px-3 bg-slate-700/50 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-medium transition-colors border border-slate-600/50"
							>
								<Terminal class="w-3.5 h-3.5" />
								Logs
							</button>
							<button
								onclick={() => dispatch('spawn', spawner.id)}
								disabled={spawner.status !== 'Online'}
								class="flex items-center justify-center gap-1.5 py-2 px-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-medium transition-colors disabled:opacity-50 disabled:bg-slate-700"
							>
								<Zap class="w-3.5 h-3.5" />
								Spawn
							</button>
						{/if}
					</div>
				</div>

				<!-- Expand Button -->
				<button 
					class="w-full py-2 flex items-center justify-center gap-1 text-xs text-slate-500 hover:text-slate-300 hover:bg-slate-800/50 border-t border-slate-700/50 transition-colors"
					onclick={() => toggleRow(spawner.id)}
				>
					<span>{expandedRows.has(spawner.id) ? 'Hide Details' : 'Show Details'}</span>
					<div class="transform transition-transform {expandedRows.has(spawner.id) ? 'rotate-180' : ''}">
						<svg width="10" height="6" viewBox="0 0 10 6" fill="none" xmlns="http://www.w3.org/2000/svg">
							<path d="M1 1L5 5L9 1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</div>
				</button>

				<!-- Expanded Details (Mobile) -->
				{#if expandedRows.has(spawner.id)}
					<div class="border-t border-slate-700/50 bg-slate-900/30 p-4" transition:slide>
						<!-- Resource Stats -->
						<div class="grid grid-cols-3 gap-3 mb-6">
							<div class="bg-slate-800/50 p-2.5 rounded-lg border border-slate-700/50">
								<div class="text-[10px] text-slate-500 uppercase font-semibold mb-1">CPU</div>
								<div class="text-sm font-bold text-slate-200">{spawner.cpu_usage?.toFixed(1) || 0}%</div>
								<div class="w-full h-1 bg-slate-700 rounded-full mt-1.5 overflow-hidden">
									<div class="h-full bg-blue-500" style="width: {spawner.cpu_usage || 0}%"></div>
								</div>
							</div>
							<div class="bg-slate-800/50 p-2.5 rounded-lg border border-slate-700/50">
								<div class="text-[10px] text-slate-500 uppercase font-semibold mb-1">MEM</div>
								<div class="text-sm font-bold text-slate-200">{((spawner.mem_used || 0) / (spawner.mem_total || 1) * 100).toFixed(0)}%</div>
								<div class="w-full h-1 bg-slate-700 rounded-full mt-1.5 overflow-hidden">
									<div class="h-full bg-purple-500" style="width: {(spawner.mem_used / spawner.mem_total) * 100 || 0}%"></div>
								</div>
							</div>
							<div class="bg-slate-800/50 p-2.5 rounded-lg border border-slate-700/50">
								<div class="text-[10px] text-slate-500 uppercase font-semibold mb-1">DISK</div>
								<div class="text-sm font-bold text-slate-200">{((spawner.disk_used || 0) / (spawner.disk_total || 1) * 100).toFixed(0)}%</div>
								<div class="w-full h-1 bg-slate-700 rounded-full mt-1.5 overflow-hidden">
									<div class="h-full bg-orange-500" style="width: {(spawner.disk_used / spawner.disk_total) * 100 || 0}%"></div>
								</div>
							</div>
						</div>

						<!-- Instances List -->
						<div>
							<div class="flex items-center justify-between mb-3">
								<h4 class="text-xs font-bold text-slate-300 uppercase tracking-wider">Instances</h4>
								<!-- Mobile Bulk Actions -->
								<Dropdown label="Actions" align="right">
									<div slot="default" let:close>
										<button onclick={() => { dispatchBulkAction('start', spawner.id); close(); }} class="w-full text-left px-4 py-2 text-sm text-emerald-400 hover:bg-emerald-500/10 flex items-center gap-2"><svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg> Start All</button>
										<button onclick={() => { dispatchBulkAction('stop', spawner.id); close(); }} class="w-full text-left px-4 py-2 text-sm text-yellow-400 hover:bg-yellow-500/10 flex items-center gap-2"><svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg> Stop All</button>
										<button onclick={() => { dispatchBulkAction('restart', spawner.id); close(); }} class="w-full text-left px-4 py-2 text-sm text-blue-400 hover:bg-blue-500/10 flex items-center gap-2"><svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M23 4v6h-6"></path><path d="M1 20v6h6"></path><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg> Restart All</button>
										<button onclick={() => { dispatchBulkAction('update', spawner.id); close(); }} class="w-full text-left px-4 py-2 text-sm text-purple-400 hover:bg-purple-500/10 flex items-center gap-2 border-t border-slate-800"><svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg> Update All ({getOutdatedCount(spawner.id)})</button>
									</div>
								</Dropdown>
								<div class="w-px h-4 bg-slate-700 mx-1"></div>
								{#if loadingInstances[spawner.id]}
									<div class="w-3 h-3 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
								{/if}
								<button onclick={() => fetchInstances(spawner.id)} class="text-xs text-blue-400 hover:text-blue-300">Refresh</button>
							</div>
						</div>

						{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
							<div class="p-4 text-center text-slate-500 text-xs">Loading instances...</div>
						{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
							<div class="p-4 text-center text-slate-500 text-xs bg-slate-800/30 rounded border border-slate-700/50">No active instances found.</div>
						{:else}
							<div class={loadingInstances[spawner.id] ? 'opacity-50 pointer-events-none transition-opacity' : 'transition-opacity'}>
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
				</td>
			</tr>
		{/if}
	{/each}
</tbody>
		</table>
	</div>
</div>

<style lang="ts">
	@keyframes highlight-new-spawner {
		0% {
			background-color: theme('colors.blue.900');
		}
		50% {
			background-color: theme('colors.blue.900');
		} /* Hold color */
		100% {
			background-color: transparent;
		}
	}

	.animate-highlight-new-spawner {
		animation: highlight-new-spawner 5s ease-out forwards;
	}
</style>
