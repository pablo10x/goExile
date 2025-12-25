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
		MoreHorizontal,
		Server,
		Activity,
		HardDrive,
		Cpu,
		Zap,
		Box,
		Terminal,
		Settings,
		ChevronRight,
		ChevronDown
	} from 'lucide-svelte';
	import { slide } from 'svelte/transition';

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
				return 'bg-emerald-500 text-black border-emerald-400';
			case 'Degraded':
				return 'bg-yellow-500 text-black border-yellow-400';
			case 'Unresponsive':
				return 'bg-orange-600 text-white border-orange-500';
			case 'Offline':
				return 'bg-stone-800 text-stone-500 border-stone-700';
			default:
				return 'bg-stone-800 text-stone-500 border-stone-700';
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
	<div class="md:hidden space-y-6 px-2 py-4">
		{#each spawners as spawner}
			<div
				class="bg-black/60 border-2 border-stone-800 rounded-none overflow-hidden {spawner.id ===
				highlightNewSpawnerId && $siteSettings.aesthetic.animations_enabled
					? 'border-rust shadow-[0_0_20px_rgba(120,53,15,0.3)]'
					: ''}"
			>
				<!-- Card Header -->
				<div
					class="p-4 border-b-2 border-stone-800 flex items-center justify-between bg-stone-900/40"
					onclick={() => toggleRow(spawner.id)}
					onkeydown={(e) => {
						if (e.key === 'Enter' || e.key === ' ') toggleRow(spawner.id);
					}}
					role="button"
					tabindex="0"
				>
					<div class="flex items-center gap-4">
						<div class="p-2 bg-stone-800 border border-stone-700">
							<Server class="w-5 h-5 text-rust" />
						</div>
						<div>
							<div class="flex items-center gap-2">
								<span class="font-black text-white uppercase tracking-tighter text-sm">{spawner.name || spawner.region}</span>
								<span class="font-mono text-[10px] text-rust-light">#{spawner.id.toString().padStart(3, '0')}</span>
							</div>
							<div class="tactical-code text-stone-500 mt-0.5">
								{spawner.host}:{spawner.port}
							</div>
						</div>
					</div>
					<div class="flex flex-col items-end gap-1">
						<span
							class={`text-[9px] px-2 py-0.5 font-black uppercase border-l-2 ${getStatusClass(spawner.status)}`}
						>
							{spawner.status}
						</span>
					</div>
				</div>

				<!-- Card Body -->
				<div class="p-4 space-y-5">
					<!-- Instances Progress -->
					<div>
						<div class="flex justify-between items-center mb-2">
							<span class="tactical-code text-stone-500">Node_Load</span>
							<span class="font-jetbrains text-[10px] font-bold text-rust-light"
								>{spawner.current_instances} / {spawner.max_instances}</span
							>
						</div>
						<div class="w-full h-1.5 bg-stone-900 border border-stone-800 p-[1px]">
							<div
								class="h-full bg-rust shadow-[0_0_10px_rgba(120,53,15,0.4)] transition-all duration-500"
								style={`width: ${getInstancePercent(spawner)}%`}
							></div>
						</div>
					</div>

					<!-- Actions Grid -->
					<div class="grid grid-cols-2 gap-2">
						{#if spawner.status === 'Offline'}
							<button
								onclick={() => deleteSpawner(spawner.id)}
								class="col-span-2 py-2.5 bg-red-950/20 text-red-500 border border-red-900/50 text-[10px] font-black uppercase hover:bg-red-600 hover:text-black transition-all"
							>
								Decommission_Offline_Node
							</button>
						{:else}
							<button
								onclick={() => dispatch('viewLogs', spawner.id)}
								class="py-2.5 bg-stone-800 text-stone-300 border border-stone-700 text-[10px] font-black uppercase hover:bg-white hover:text-black transition-all"
							>
								Terminal
							</button>
							<button
								onclick={() => dispatch('spawn', spawner.id)}
								disabled={spawner.status !== 'Online'}
								class="py-2.5 bg-rust text-white border border-rust-light text-[10px] font-black uppercase hover:bg-rust-light transition-all disabled:opacity-30 disabled:grayscale"
							>
								Execute
							</button>
						{/if}
					</div>
				</div>

				<!-- Expand Button -->
				<button
					class="w-full py-3 flex items-center justify-center gap-3 text-[10px] font-black uppercase tracking-widest text-rust-light hover:text-white hover:bg-rust/10 border-t-2 border-stone-800 transition-all"
					onclick={() => toggleRow(spawner.id)}
				>
					<span>{expandedRows.has(spawner.id) ? 'Defragment_View' : 'Access_Detailed_Node_Report'}</span>
					<div
						class="transform transition-transform {expandedRows.has(spawner.id)
							? 'rotate-180'
							: ''}"
					>
						<ChevronDown class="w-3.5 h-3.5" />
					</div>
				</button>

				<!-- Expanded Details (Mobile) -->
				{#if expandedRows.has(spawner.id)}
					<div
						class="border-t-2 border-stone-800 bg-stone-950/50 p-4 space-y-6"
						transition:slide
					>
						<div class="grid grid-cols-1 gap-3">
							<div class="brutalist-card p-3 border-l-4 border-l-rust">
								<div class="flex justify-between items-center mb-1">
									<span class="tactical-code text-stone-500 text-[9px]">Processor_Load</span>
									<Cpu class="w-3 h-3 text-rust" />
								</div>
								<div class="text-xl font-black military-label">
									{spawner.cpu_usage?.toFixed(1) || 0}%
								</div>
								<div class="w-full h-1 bg-stone-900 mt-2 border border-stone-800">
									<div class="h-full bg-rust" style="width: {spawner.cpu_usage || 0}%"></div>
								</div>
							</div>
							
							<div class="grid grid-cols-2 gap-3">
								<div class="brutalist-card p-3 border-l-4 border-l-rust">
									<div class="flex justify-between items-center mb-1">
										<span class="tactical-code text-stone-500 text-[9px]">RAM</span>
										<Zap class="w-3 h-3 text-rust" />
									</div>
									<div class="text-sm font-black military-label">
										{(((spawner.mem_used || 0) / (spawner.mem_total || 1)) * 100).toFixed(0)}%
									</div>
								</div>
								<div class="brutalist-card p-3 border-l-4 border-l-rust">
									<div class="flex justify-between items-center mb-1">
										<span class="tactical-code text-stone-500 text-[9px]">DISK</span>
										<HardDrive class="w-3 h-3 text-rust" />
									</div>
									<div class="text-sm font-black military-label">
										{(((spawner.disk_used || 0) / (spawner.disk_total || 1)) * 100).toFixed(0)}%
									</div>
								</div>
							</div>
						</div>

						<div class="space-y-4">
							<div class="flex items-center justify-between border-b border-stone-800 pb-2">
								<h4 class="text-[10px] font-black military-label uppercase tracking-widest">Sub_Nodes</h4>
								<Dropdown label="CMD">
									{#snippet children()}
										<button onclick={() => dispatchBulkAction('start', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-emerald-400 hover:bg-emerald-500/10">Execute_All</button>
										<button onclick={() => dispatchBulkAction('stop', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-rose-400 hover:bg-rose-500/10">Kill_All</button>
										<div class="border-t border-stone-800 my-1"></div>
										<button onclick={() => dispatchBulkAction('update', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-rust-light hover:bg-white hover:text-black">Patch_Stream</button>
									{/snippet}
								</Dropdown>
							</div>

							{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
								<div class="p-8 text-center tactical-code text-stone-600 bg-stone-900/40 border-2 border-dashed border-stone-800">Buffer_Sync...</div>
							{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
								<div class="p-8 text-center tactical-code text-stone-600 bg-stone-900/40 border-2 border-dashed border-stone-800">NULL_REPORT</div>
							{:else}
								<div class={loadingInstances[spawner.id] ? 'opacity-50 pointer-events-none' : ''}>
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
			<div class="p-12 text-center bg-stone-950 border-2 border-dashed border-stone-800">
				<Box class="w-10 h-10 text-stone-800 mx-auto mb-4" />
				<p class="tactical-code text-stone-600">No_Nodes_Registered_In_Registry</p>
			</div>
		{/if}
	</div>

	<!-- Desktop View (Table) -->
	<div class="hidden md:block overflow-x-auto">
		<table class="w-full text-sm border-separate border-spacing-y-2 px-4">
			<thead>
				<tr class="tactical-code text-stone-500">
					<th class="w-8"></th>
					<th class="text-left px-4 py-4 font-black uppercase">Ref_ID</th>
					<th class="text-left px-4 py-4 font-black uppercase">Tactical_Name</th>
					<th class="text-left px-4 py-4 font-black uppercase">Region_Code</th>
					<th class="text-left px-4 py-4 font-black uppercase">Network_ADDR</th>
					<th class="text-left px-4 py-4 font-black uppercase">Operational_Status</th>
					<th class="text-left px-4 py-4 font-black uppercase">Buffer_Load</th>
					<th class="text-left px-4 py-4 font-black uppercase">Ver</th>
					<th class="text-right px-4 py-4 font-black uppercase">CMD</th>
				</tr>
			</thead>
			<tbody class="text-slate-300">
				{#each spawners as spawner}
					<tr
						class="industrial-frame bg-stone-900/40 hover:bg-rust/5 transition-all cursor-pointer group {spawner.id ===
						highlightNewSpawnerId && $siteSettings.aesthetic.animations_enabled
							? 'animate-highlight-new-spawner'
							: ''}"
						onclick={() => toggleRow(spawner.id)}
						onkeydown={(e) => {
							if (e.key === 'Enter' || e.key === ' ') toggleRow(spawner.id);
						}}
						role="button"
						tabindex="0"
					>
						<td class="px-4 py-5 text-center text-stone-600">
							<div class="transition-transform duration-300 {expandedRows.has(spawner.id) ? 'rotate-90 text-rust' : ''}">
								<ChevronRight class="w-4 h-4" />
							</div>
						</td>
						<td class="px-4 py-5 font-jetbrains font-bold text-rust-light">[{spawner.id.toString().padStart(3, '0')}]</td>
						<td class="px-4 py-5 font-heading font-black tracking-tight text-white uppercase">{spawner.name || spawner.region}</td>
						<td class="px-4 py-5">
							<span class="px-2 py-0.5 bg-stone-800 text-[10px] font-black uppercase border border-stone-700">{spawner.region}</span>
						</td>
						<td class="px-4 py-5 font-mono text-[11px] text-stone-500">{spawner.host}:{spawner.port}</td>
						<td class="px-4 py-5">
							<span class={`inline-flex items-center gap-2 px-3 py-1 rounded-none text-[10px] font-black uppercase border-l-4 ${getStatusClass(spawner.status)}`}>
								{spawner.status}
							</span>
						</td>
						<td class="px-4 py-5 w-48">
							<div class="flex items-center justify-between gap-3">
								<div class="flex-1 h-1.5 bg-stone-800 rounded-none overflow-hidden border border-stone-700 p-[1px]">
									<div class="h-full bg-rust shadow-[0_0_10px_rgba(120,53,15,0.5)] transition-all duration-1000" style={`width: ${getInstancePercent(spawner)}%`}></div>
								</div>
								<span class="font-jetbrains text-[10px] font-bold text-rust-light">{spawner.current_instances}/{spawner.max_instances}</span>
							</div>
						</td>
						<td class="px-4 py-5 font-jetbrains text-stone-600 text-[10px]">{spawner.game_version || '??'}</td>
												<td class="px-4 py-5 text-right space-x-2" onclick={(e) => e.stopPropagation()}>
													{#if spawner.status === 'Offline'}
														<button
															onclick={() => deleteSpawner(spawner.id)}
															class="px-4 py-1.5 bg-red-950/20 text-red-500 border border-red-900/50 rounded-none text-[10px] font-black uppercase hover:bg-red-600 hover:text-black transition-all"
														>
															Decommission_Node
														</button>
							{:else}
								<button onclick={() => dispatch('viewLogs', spawner.id)} class="px-4 py-1.5 bg-stone-800 text-stone-300 border border-stone-700 rounded-none text-[10px] font-black uppercase hover:bg-white hover:text-black transition-all">Terminal</button>
								<button onclick={() => dispatch('spawn', spawner.id)} disabled={spawner.status !== 'Online'} class="px-4 py-1.5 bg-rust text-white border border-rust-light rounded-none text-[10px] font-black uppercase hover:bg-rust-light transition-all shadow-[4px_4px_0px_rgba(120,53,15,0.3)] disabled:opacity-30 disabled:grayscale">Execute</button>
							{/if}
						</td>
					</tr>
					{#if expandedRows.has(spawner.id)}
						<tr class="bg-black/40 border-b-2 border-stone-800">
							<td colspan="9" class="px-8 py-8">
								<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
									<div class="brutalist-card p-4 border-l-4 border-l-rust">
										<div class="flex justify-between items-center mb-2">
											<span class="tactical-code text-stone-500">Core_Utilization</span>
											<Cpu class="w-3.5 h-3.5 text-rust" />
										</div>
										<div class="text-3xl font-black military-label mb-2">{spawner.cpu_usage ? spawner.cpu_usage?.toFixed(1) : 0}%</div>
										<div class="w-full h-1 bg-stone-900 rounded-none overflow-hidden border border-stone-800">
											<div class="h-full bg-rust" style="width: {spawner.cpu_usage || 0}%"></div>
										</div>
									</div>
									<div class="brutalist-card p-4 border-l-4 border-l-rust">
										<div class="flex justify-between items-center mb-2">
											<span class="tactical-code text-stone-500">Volatile_Memory</span>
											<Zap class="w-3.5 h-3.5 text-rust" />
										</div>
										<div class="text-xl font-black military-label mb-1">{formatBytes(spawner.mem_used || 0)} <span class="text-xs text-stone-600">/ {formatBytes(spawner.mem_total || 0)}</span></div>
										<div class="text-[10px] font-mono text-stone-500 mb-2">AVAILABLE: {formatBytes((spawner.mem_total || 0) - (spawner.mem_used || 0))}</div>
										<div class="w-full h-1 bg-stone-900 rounded-none overflow-hidden border border-stone-800">
											<div class="h-full bg-rust" style="width: {spawner.mem_total ? (spawner.mem_used / spawner.mem_total) * 100 : 0}%"></div>
										</div>
									</div>
									<div class="brutalist-card p-4 border-l-4 border-l-rust">
										<div class="flex justify-between items-center mb-2">
											<span class="tactical-code text-stone-500">Storage_Array</span>
											<HardDrive class="w-3.5 h-3.5 text-rust" />
										</div>
										<div class="text-xl font-black military-label mb-1">{formatBytes(spawner.disk_used || 0)} <span class="text-xs text-stone-600">/ {formatBytes(spawner.disk_total || 0)}</span></div>
										<div class="text-[10px] font-mono text-stone-500 mb-2">I/O_STREAM: ACTIVE</div>
										<div class="w-full h-1 bg-stone-900 rounded-none overflow-hidden border border-stone-800">
											<div class="h-full bg-rust" style="width: {spawner.disk_total ? (spawner.disk_used / spawner.disk_total) * 100 : 0}%"></div>
										</div>
									</div>
								</div>

								<div class="space-y-4">
									<div class="flex justify-between items-center mb-4 px-1 border-b border-stone-800 pb-2">
										<div class="flex items-center gap-3">
											<Box class="w-4 h-4 text-rust" />
											<h4 class="text-sm font-black military-label uppercase tracking-widest">Sub_Node_Instances</h4>
										</div>
										<div class="flex items-center gap-4">
											<Dropdown label="System_CMD">
												{#snippet children()}
													<button onclick={() => dispatchBulkAction('start', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-emerald-400 hover:bg-emerald-500/10">Execute_All</button>
													<button onclick={() => dispatchBulkAction('stop', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-yellow-400 hover:bg-yellow-500/10">Terminate_All</button>
													<button onclick={() => dispatchBulkAction('restart', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-rust-light hover:bg-rust/10">Reboot_All</button>
													<button onclick={() => dispatchBulkAction('update', spawner.id)} class="w-full text-left px-4 py-2 text-[10px] font-black uppercase text-purple-400 hover:bg-purple-500/10 border-t border-stone-800">Patch_All</button>
												{/snippet}
											</Dropdown>
											<div class="w-px h-6 bg-stone-800"></div>
											{#if loadingInstances[spawner.id]}
												<div class="w-4 h-4 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
											{/if}
											<button onclick={() => fetchInstances(spawner.id)} class="tactical-code text-rust-light hover:text-white transition-colors">Recalibrate_Stream</button>
										</div>
									</div>

									{#if loadingInstances[spawner.id] && (!activeInstances[spawner.id] || activeInstances[spawner.id].length === 0)}
										<div class="p-12 text-center tactical-code text-stone-600 bg-stone-900/20 border-2 border-dashed border-stone-800">Synchronizing_Instance_Buffer...</div>
									{:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
										<div class="p-12 text-center tactical-code text-stone-600 bg-stone-900/20 border-2 border-dashed border-stone-800">NULL_INSTANCE_REPORTED</div>
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
						<td colspan="9" class="px-4 py-20 text-center bg-stone-950 border-2 border-dashed border-stone-800">
							<div class="flex flex-col items-center gap-4">
								<Box class="w-12 h-12 text-stone-800" />
								<p class="tactical-code text-stone-600 text-lg">Empty_Registry_Reported</p>
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
</style>