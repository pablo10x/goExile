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
			if (!res.ok) {
				activeInstances[id] = [];
				return; // Exit early if not ok to avoid parsing error for no content
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
				<div
					class="p-4 border-b border-slate-700/50 flex items-center justify-between"
					onclick={() => toggleRow(spawner.id)}
					onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleRow(spawner.id); }}
					role="button"
					tabindex="0"
				>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-slate-800 rounded-lg border border-slate-700">
							<Server class="w-5 h-5 text-slate-400" />
						</div>
						<div>
							<div class="flex items-center gap-2">
								<span class="font-bold text-slate-100">{spawner.name || spawner.region}</span>
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
								<Settings class="w-3.5 h-3.5" />
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
								                                <div class="flex justify-end pr-2">
								                                    <Dropdown label="Actions">
								                                        {#snippet children()}
								                                            <button
								                                                onclick={() => {
								                                                    dispatchBulkAction('start', spawner.id);
								                                                }}
								                                                class="w-full text-left px-4 py-2 text-sm text-emerald-400 hover:bg-emerald-500/10"
								                                                >Start All</button
								                                            >
								                                            <button
								                                                onclick={() => {
								                                                    dispatchBulkAction('stop', spawner.id);
								                                                }}
								                                                class="w-full text-left px-4 py-2 text-sm text-rose-400 hover:bg-rose-500/10"
								                                                >Stop All</button
								                                            >
								                                            <button
								                                                onclick={() => {
								                                                    dispatchBulkAction('restart', spawner.id);
								                                                }}
								                                                class="w-full text-left px-4 py-2 text-sm text-blue-400 hover:bg-blue-500/10"
								                                                >Restart All</button
								                                            >
								                                            <div class="border-t border-slate-800 my-1"></div>
								                                            <button
								                                                onclick={() => {
								                                                    dispatchBulkAction('update', spawner.id);
								                                                }}
								                                                class="w-full text-left px-4 py-2 text-sm text-slate-300 hover:bg-slate-800"
								                                                >Update All</button
								                                            >
								                                        {/snippet}
								                                    </Dropdown>
								                                </div>							</div>

							{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
								<div class="p-4 text-center text-slate-500 text-xs bg-slate-800/30 rounded-lg">Loading...</div>
							{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
								<div class="p-4 text-center text-slate-500 text-xs bg-slate-800/30 rounded-lg border border-slate-700/50">No active instances.</div>
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
					</div>
				{/if}
			</div>
		{/each}
		
		{#if spawners.length === 0}
			<div class="p-8 text-center bg-slate-800/20 rounded-xl border border-slate-700/50 border-dashed">
				<Box class="w-10 h-10 text-slate-600 mx-auto mb-3" />
				<p class="text-slate-400 text-sm">No spawners registered.</p>
			</div>
		{/if}
	</div>

	<!-- Desktop View (Table) -->
	<div class="hidden md:block overflow-x-auto">
		<table class="w-full text-sm">
			<thead class="border-b border-slate-700">
				<tr>
					<th class="w-8"></th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">ID</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Name</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Region</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Address</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Instances</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Game Version</th>
					<th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider text-right">Actions</th>
				</tr>
			</thead>
			<tbody class="text-slate-300">
				{#each spawners as spawner}
					<tr
						class="border-t border-slate-700 hover:bg-slate-800/50 transition cursor-pointer {spawner.id === highlightNewSpawnerId ? 'animate-highlight-new-spawner' : ''}"
						onclick={() => toggleRow(spawner.id)}
						onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleRow(spawner.id); }}
						role="button"
						tabindex="0"
					>
						<td class="px-4 py-3 text-center text-slate-500">
							<span class="inline-block transition-transform duration-200 {expandedRows.has(spawner.id) ? 'rotate-90' : ''}">▶</span>
						</td>
						<td class="px-4 py-3 text-slate-400">#{spawner.id}</td>
						<td class="px-4 py-3 font-semibold text-slate-100">{spawner.name || spawner.region}</td>
						<td class="px-4 py-3 font-semibold text-slate-100">{spawner.region}</td>
						<td class="px-4 py-3 font-mono text-slate-400">{spawner.host}:{spawner.port}</td>
						<td class="px-4 py-3">
							<span class={`inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-semibold ${getStatusClass(spawner.status)}`}>
								<span class="w-2 h-2 rounded-full bg-current"></span>
								{spawner.status}
							</span>
						</td>
						<td class="px-4 py-3">
							<div class="text-sm mb-2">{spawner.current_instances} / {spawner.max_instances}</div>
							<div class="w-full h-1.5 bg-slate-700 rounded-full overflow-hidden">
								<div class="h-full bg-gradient-to-r from-blue-500 to-purple-500" style={`width: ${getInstancePercent(spawner)}%`}></div>
							</div>
						</td>
						<td class="px-4 py-3 font-mono text-slate-400 text-xs">
							{spawner.game_version || 'N/A'}
						</td>
						<td class="px-4 py-3 text-right space-x-2" onclick={(e) => e.stopPropagation()}>
							{#if spawner.status === 'Offline'}
								<button onclick={() => deleteSpawner(spawner.id)} class="inline-flex items-center gap-2 px-3 py-1 bg-red-900/30 hover:bg-red-900/50 text-red-400 border border-red-900/50 rounded text-xs font-semibold transition-colors">
									<Trash2 class="w-3.5 h-3.5" />
									Delete
								</button>
							{:else}
								<a href="/spawners/{spawner.id}" class="inline-flex items-center gap-2 px-3 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded text-xs font-semibold transition-colors">
									<svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
									Manage
								</a>
								<button onclick={() => dispatch('viewLogs', spawner.id)} class="px-3 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded text-xs font-semibold transition-colors">Logs</button>
								<button onclick={() => dispatch('spawn', spawner.id)} disabled={spawner.status !== 'Online'} class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-semibold transition-colors shadow-lg shadow-blue-900/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-slate-700">Spawn</button>
							{/if}
						</td>
					</tr>
					{#if expandedRows.has(spawner.id)}
						<tr class="bg-slate-900/50 border-b border-slate-700">
							<td colspan="9" class="px-4 py-4"> <!-- Colspan updated from 7 to 9 -->
								<div class="grid grid-cols-1 md:grid-cols-3 gap-6 text-xs mb-6">
									<!-- CPU -->
									<div class="bg-slate-800 rounded p-3 border border-slate-700">
										<div class="text-slate-400 uppercase font-semibold mb-1">CPU Usage</div>
										<div class="text-lg font-bold text-slate-100">{spawner.cpu_usage ? spawner.cpu_usage.toFixed(1) : 0}%</div>
										<div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
											<div class="h-full bg-blue-500" style="width: {spawner.cpu_usage || 0}%"></div>
										</div>
									</div>
									<!-- Memory -->
									<div class="bg-slate-800 rounded p-3 border border-slate-700">
										<div class="text-slate-400 uppercase font-semibold mb-1">Memory Usage</div>
										<div class="text-lg font-bold text-slate-100">{formatBytes(spawner.mem_used || 0)} / {formatBytes(spawner.mem_total || 0)}</div>
										<div class="text-slate-500 mt-1">Free: {formatBytes((spawner.mem_total || 0) - (spawner.mem_used || 0))}</div>
										<div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
											<div class="h-full bg-purple-500" style="width: {spawner.mem_total ? (spawner.mem_used / spawner.mem_total) * 100 : 0}%"></div>
										</div>
									</div>
									<!-- Disk -->
									<div class="bg-slate-800 rounded p-3 border border-slate-700">
										<div class="text-slate-400 uppercase font-semibold mb-1">Disk Usage</div>
										<div class="text-lg font-bold text-slate-100">{formatBytes(spawner.disk_used || 0)} / {formatBytes(spawner.disk_total || 0)}</div>
										<div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
											<div class="h-full bg-orange-500" style="width: {spawner.disk_total ? (spawner.disk_used / spawner.disk_total) * 100 : 0}%"></div>
										</div>
									</div>
								</div>

								<!-- Active Instances -->
								<div class="space-y-2">
									<div class="flex justify-between items-center mb-2 px-1">
										<h4 class="text-xs font-bold text-slate-300 uppercase tracking-wider">Active Instances</h4>
										<div class="flex items-center gap-2">
											                                                                                        <Dropdown label="Bulk Actions">
											                                                                                         {#snippet children()}
											                                                                                         <button onclick={() => { dispatchBulkAction('start', spawner.id); }} class="w-full text-left px-4 py-2 text-sm text-emerald-400 hover:bg-emerald-500/10">Start All</button>
											                                                                                         <button onclick={() => { dispatchBulkAction('stop', spawner.id); }} class="w-full text-left px-4 py-2 text-sm text-yellow-400 hover:bg-yellow-500/10">Stop All</button>
											                                                                                         <button onclick={() => { dispatchBulkAction('restart', spawner.id); }} class="w-full text-left px-4 py-2 text-sm text-blue-400 hover:bg-blue-500/10">Restart All</button>
											                                                                                         <button onclick={() => { dispatchBulkAction('update', spawner.id); }} class="w-full text-left px-4 py-2 text-sm text-purple-400 hover:bg-purple-500/10 border-t border-slate-800">Update All</button>
											                                                                                         {/snippet}
											                                                                                        </Dropdown>											<div class="w-px h-4 bg-slate-700 mx-1"></div>
											{#if loadingInstances[spawner.id]}
												<div class="w-3 h-3 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
											{/if}
											<button onclick={() => fetchInstances(spawner.id)} class="text-xs text-blue-400 hover:text-blue-300">Refresh</button>
										</div>
									</div>

									{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
										<div class="p-4 text-center text-slate-500 text-xs">Loading instances...</div>
									{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
										<div class="p-4 text-center text-slate-500 text-xs bg-slate-800/30 rounded border border-slate-700/50">No active instances.</div>
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
				{#if spawners.length === 0}
					<tr>
						<td colspan="9" class="px-4 py-12 text-center text-slate-500 bg-slate-800/20 rounded-b-lg">
							<div class="flex flex-col items-center gap-2">
								<span class="text-2xl">📭</span>
								<span>No spawners registered yet.</span>
							</div>
						</td>
					</tr>
				{/if}
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