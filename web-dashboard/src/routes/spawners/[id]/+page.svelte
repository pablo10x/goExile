<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { formatBytes, formatUptime } from '$lib/utils';
	import InstanceRow from '$lib/components/InstanceRow.svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import Dropdown from '$lib/components/Dropdown.svelte';
	import { serverVersions } from '$lib/stores';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from '$lib/components/PlayersChart.svelte';
	import { Server, Cpu, HardDrive, MemoryStick, List, Plus, ArrowLeft, AlertCircle, Activity } from 'lucide-svelte';

	const spawnerId = parseInt($page.params.id || '0');

	let spawner: any = null;
	let instances: any[] = [];
	let isLoading = true;
	let error: string | null = null;
	let refreshInterval: any;

	// Dummy Data for Chart
	const chartData = Array.from({ length: 24 }, (_, i) => {
		const time = new Date().getTime() - (23 - i) * 3600000;
		return {
			timestamp: time,
			count: Math.floor(Math.max(0, 50 + Math.sin(i / 3) * 30 + (Math.random() - 0.5) * 20))
		};
	});

	// Instance Action State
	let isInstanceActionDialogOpen = false;
	let instanceActionType:
		| 'start'
		| 'stop'
		| 'delete'
		| 'update'
		| 'rename'
		| 'restart'
		| 'bulk_stop'
		| 'bulk_restart'
		| 'bulk_update'
		| 'bulk_start'
		| 'update_spawner_build'
		| null = null;
	let instanceActionInstanceId: string | null = null;
	let instanceActionNewID: string | null = null;
	let instanceActionBulkIds: string[] = [];
	let instanceActionDialogTitle = '';
	let instanceActionDialogMessage = '';
	let instanceActionConfirmText = '';

	// Progress State for ConfirmDialog
	let actionProgress: number | null = null;
	let actionStatusMessage: string | null = null;

	// Spawn Dialog
	let isSpawnDialogOpen = false;

	// Console & Logs
	let isConsoleOpen = false;
	let consoleInstanceId: string | null = null;
	let isLogViewerOpen = false;

	$: activeVersion = $serverVersions.find((v) => v.is_active);

	async function fetchSpawnerData() {
		try {
			const res = await fetch(`/api/spawners/${spawnerId}`);
			if (!res.ok) {
				if (res.status === 404) throw new Error('Spawner not found');
				throw new Error('Failed to load spawner details');
			}
			spawner = await res.json();

			const instRes = await fetch(`/api/spawners/${spawnerId}/instances`);
			if (instRes.ok) {
				const data = await instRes.json();
				const list = data.instances || data || [];
				// Sort by ID to prevent jumping
				instances = list.sort((a: any, b: any) => a.id.localeCompare(b.id));
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		fetchSpawnerData();
		refreshInterval = setInterval(fetchSpawnerData, 5000); // Poll every 5s

		// Fetch versions if not already loaded
		if ($serverVersions.length === 0) {
			fetch('/api/versions')
				.then((r) => r.json())
				.then((v) => serverVersions.set(v))
				.catch(console.error);
		}
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});

	function getStatusClass(status: string) {
		if (status === 'Updating')
			return 'bg-orange-500/10 text-orange-400 border-orange-500/20 animate-pulse';
		return status === 'active' || status === 'Online' // Handle both cases just in case
			? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
			: 'bg-red-500/10 text-red-400 border-red-500/20';
	}

	// --- Action Handlers ---

	function openSpawnDialog() {
		isSpawnDialogOpen = true;
	}

	async function executeSpawn() {
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/spawn`, { method: 'POST' });
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || `Server returned ${res.status}`);
			}
			const instance = await res.json();
			consoleInstanceId = instance.id;
			isConsoleOpen = true;
			fetchSpawnerData();
		} catch (e: any) {
			alert(e.message);
		}
		isSpawnDialogOpen = false;
	}

	function openUpdateSpawnerBuildDialog() {
		instanceActionType = 'update_spawner_build';
		instanceActionDialogTitle = 'Update Spawner Build';
		instanceActionDialogMessage = `Are you sure you want Spawner #${spawnerId} to download the latest game server build? This might take a while.`;
		instanceActionConfirmText = 'Update Build';
		isInstanceActionDialogOpen = true;
	}

	function openInstanceActionDialog(type: any, instanceId: string, extra?: any) {
		instanceActionType = type;
		instanceActionInstanceId = instanceId;

		if (type === 'start') {
			instanceActionDialogTitle = 'Start Instance';
			instanceActionDialogMessage = `Are you sure you want to start instance "${instanceId}"?`;
			instanceActionConfirmText = 'Start';
		} else if (type === 'stop') {
			instanceActionDialogTitle = 'Stop Instance';
			instanceActionDialogMessage = `Are you sure you want to stop instance "${instanceId}"?`;
			instanceActionConfirmText = 'Stop';
		} else if (type === 'restart') {
			instanceActionDialogTitle = 'Restart Instance';
			instanceActionDialogMessage = `Are you sure you want to restart instance "${instanceId}"?`;
			instanceActionConfirmText = 'Restart';
		} else if (type === 'delete') {
			instanceActionDialogTitle = 'Delete Instance';
			instanceActionDialogMessage = `Are you sure you want to PERMANENTLY DELETE instance "${instanceId}"?`;
			instanceActionConfirmText = 'Delete';
		} else if (type === 'update') {
			instanceActionDialogTitle = 'Update Instance';
			instanceActionDialogMessage = `Are you sure you want to update instance "${instanceId}"?`;
			instanceActionConfirmText = 'Update';
		} else if (type === 'rename') {
			instanceActionNewID = extra;
			instanceActionDialogTitle = 'Rename Instance';
			instanceActionDialogMessage = `Rename "${instanceId}" to "${extra}"?`;
			instanceActionConfirmText = 'Rename';
		}

		isInstanceActionDialogOpen = true;
	}

	function dispatchBulkAction(action: 'start' | 'stop' | 'restart' | 'update') {
		let targetInstances = [];
		if (action === 'start')
			targetInstances = instances.filter(
				(i) => i.status !== 'Running' && i.status !== 'Provisioning'
			);
		else if (action === 'stop') targetInstances = instances.filter((i) => i.status === 'Running');
		else if (action === 'restart')
			targetInstances = instances.filter((i) => i.status === 'Running');
		else if (action === 'update')
			targetInstances = instances.filter(
				(i) => activeVersion && i.version !== activeVersion.version
			);

		if (targetInstances.length === 0) return;

		instanceActionType = `bulk_${action}` as any;
		instanceActionBulkIds = targetInstances.map((i) => i.id);

		const actionName = action.charAt(0).toUpperCase() + action.slice(1);
		instanceActionDialogTitle = `${actionName} All Instances`;
		instanceActionDialogMessage = `Are you sure you want to ${action} ${targetInstances.length} instances?`;
		instanceActionConfirmText = `${actionName} All`;
		isInstanceActionDialogOpen = true;
	}

	async function executeInstanceAction() {
		try {
			let res: Response | any = null;
			actionProgress = null;
			actionStatusMessage = null;

			if (instanceActionType === 'update_spawner_build') {
				actionStatusMessage = 'Requesting spawner update...';
				res = await fetch(`/api/spawners/${spawnerId}/update-template`, { method: 'POST' });
			} else if (instanceActionType?.startsWith('bulk_')) {
				const action = instanceActionType.replace('bulk_', '');
				let failureCount = 0;

				// Initialize progress
				actionProgress = 0;
				const total = instanceActionBulkIds.length;

				// Execute sequentially or in small batches to show progress
				for (let i = 0; i < total; i++) {
					const id = instanceActionBulkIds[i];
					actionStatusMessage = `Processing ${id} (${i + 1}/${total})...`;

					try {
						let url = '';
						if (action === 'restart') {
							await fetch(`/api/spawners/${spawnerId}/instances/${id}/stop`, { method: 'POST' });
							// Small delay between stop and start?
							url = `/api/spawners/${spawnerId}/instances/${id}/start`;
						} else {
							url = `/api/spawners/${spawnerId}/instances/${id}/${action}`;
						}
						const r = await fetch(url, { method: 'POST' });
						if (!r.ok) throw new Error('Failed');
					} catch {
						failureCount++;
					}

					// Update progress
					actionProgress = ((i + 1) / total) * 100;
				}

				if (failureCount > 0) alert(`${failureCount} actions failed.`);
				res = { ok: true, text: () => Promise.resolve('') } as any;

				// Brief delay to show 100%
				await new Promise((r) => setTimeout(r, 500));
			} else {
				// Single instance action
				actionStatusMessage = 'Executing action...';
				let url = `/api/spawners/${spawnerId}/instances/${instanceActionInstanceId}`;
				let method = 'POST';

				if (instanceActionType === 'delete') {
					method = 'DELETE';
					res = await fetch(url, { method });
				} else if (instanceActionType === 'rename') {
					url += '/rename';
					res = await fetch(url, {
						method,
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ new_id: instanceActionNewID })
					});
				} else if (instanceActionType === 'restart') {
					actionStatusMessage = 'Stopping instance...';
					await fetch(url + '/stop', { method: 'POST' });
					actionStatusMessage = 'Starting instance...';
					url += '/start';
					res = await fetch(url, { method: 'POST' });
				} else {
					url += `/${instanceActionType}`;
					res = await fetch(url, { method });
				}
			}

			if (res && !res.ok) {
				const txt = await res.text();
				throw new Error(txt || 'Action failed');
			}

			fetchSpawnerData();
		} catch (e: any) {
			alert(e.message); // The dialog catches errors too, but alert is fine for fallback
			throw e; // Re-throw so dialog stays open or handles error
		} finally {
			// Reset progress state if we are closing manually,
			// but ConfirmDialog might close itself on success.
			// We set them to null just in case.
			actionProgress = null;
			actionStatusMessage = null;
		}
		isInstanceActionDialogOpen = false;
	}

	function handleTail(e: CustomEvent) {
		consoleInstanceId = e.detail.instanceId;
		isConsoleOpen = true;
	}
</script>

<div class="w-full space-y-10 pb-32 md:pb-12">
	<!-- Header -->
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-8 mb-10">
		<div class="flex items-center gap-6">
			<a
				href="/dashboard"
				aria-label="Back to Dashboard"
				class="p-4 bg-stone-900 border border-stone-800 text-stone-500 hover:text-white hover:border-rust transition-all industrial-frame shadow-xl"
			>
				<ArrowLeft class="w-6 h-6" />
			</a>
			<div>
				<div class="flex items-center gap-3 mb-1">
					<div class="h-0.5 w-8 bg-rust"></div>
					<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-[0.3em]">Node_Telemetry_Bridge</span>
				</div>
				<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter">
					Spawner_<span class="text-rust">#{spawnerId}</span>
				</h1>
				{#if spawner}
					<div class="flex items-center gap-4 mt-3">
						<div
							class={`px-3 py-1 font-jetbrains font-bold text-[10px] uppercase flex items-center gap-2.5 border ${getStatusClass(spawner.status)}`}
						>
							<span class={`w-1.5 h-1.5 rounded-full ${spawner.status === 'Online' || spawner.status === 'active' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span>
							{spawner.status}
						</div>
						<div class="w-px h-4 bg-stone-800"></div>
						<span class="text-[10px] font-jetbrains font-bold text-stone-600 uppercase tracking-widest">Binary_Rev: v{spawner.game_version || '0.0.0'}</span>
					</div>
				{/if}
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-4">
			<button
				onclick={() => (isLogViewerOpen = true)}
				class="px-8 py-3 bg-stone-900 hover:bg-white hover:text-black text-stone-400 font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800 active:translate-y-px shadow-xl"
			>
				Console_Output
			</button>

			{#if spawner && activeVersion}
				{@const cmp = compareVersions(activeVersion.version, spawner.game_version || '0.0.0')}
				{#if cmp !== 0}
					<button
						onclick={() => openUpdateSpawnerBuildDialog()}
						disabled={spawner.status === 'Updating'}
						class={`px-8 py-3 font-heading font-black text-[11px] uppercase tracking-widest transition-all border shadow-xl whitespace-nowrap active:translate-y-px ${spawner.status === 'Updating' ? 'bg-stone-800 text-stone-600 border-stone-700 cursor-not-allowed' : cmp > 0 ? 'bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-400 shadow-emerald-900/20' : 'bg-rust hover:bg-rust-light text-white border-rust-light shadow-rust/20'}`}
					>
						{spawner.status === 'Updating' ? 'Synchronizing...' : cmp > 0 ? 'Apply_Patch' : 'Revert_Rev'}
					</button>
				{:else}
					<div
						class="px-8 py-3 bg-stone-950 text-stone-700 font-heading font-black text-[11px] border border-stone-900 uppercase tracking-widest shadow-inner cursor-default"
					>
						Rev_Synchronized
					</div>
				{/if}
			{/if}
		</div>
	</div>

	{#if isLoading && !spawner}
		<div class="flex flex-col items-center justify-center h-96 gap-6">
			<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
			<p class="font-heading font-black text-[12px] text-rust animate-pulse uppercase tracking-[0.5em]">Establishing_Uplink...</p>
		</div>
	{:else if error}
		<div class="p-12 text-center bg-red-950/10 border border-red-900/30 industrial-frame shadow-2xl">
			<AlertCircle class="w-16 h-16 text-red-600 mx-auto mb-6 animate-pulse" />
			<h2 class="text-2xl font-heading font-black text-red-500 mb-3 uppercase tracking-widest">Terminal_Connection_Fault</h2>
			<p class="font-jetbrains text-stone-500 font-bold uppercase tracking-tight">{error}</p>
			<button class="mt-10 px-10 py-3 bg-red-600 hover:bg-red-500 text-white font-heading font-black text-[11px] uppercase tracking-widest transition-all shadow-lg" onclick={fetchSpawnerData}>Retry_Protocol</button>
		</div>
	{:else if spawner}
		<!-- Spawner Stats -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-10">
			<StatsCard
				title="Active_Sub_Nodes"
				value={`${spawner.current_instances} / ${spawner.max_instances}`}
				subValue={`CAPACITY_LOAD: ${spawner.max_instances > 0 ? ((spawner.current_instances / spawner.max_instances) * 100).toFixed(0) : 0}%`}
				Icon={Server}
				color="cyan"
			/>
			<StatsCard
				title="Core_Utilization"
				value={`${spawner.cpu_usage?.toFixed(1) || 0}%`}
				Icon={Cpu}
				color="purple"
			/>
			<StatsCard
				title="Volatile_Memory"
				value={formatBytes(spawner.mem_used || 0)}
				subValue={`REGISTRY_CAP: ${formatBytes(spawner.mem_total || 0)}`}
				Icon={MemoryStick}
				color="emerald"
			/>
			<StatsCard
				title="Storage_Array"
				value={formatBytes(spawner.disk_used || 0)}
				subValue={`ARRAY_TOTAL: ${formatBytes(spawner.disk_total || 0)}`}
				Icon={HardDrive}
				color="orange"
			/>
		</div>

		<!-- Instances Section -->
		<div
			class="modern-industrial-card glass-panel !rounded-none overflow-hidden relative shadow-2xl"
		>
			{#if spawner.status === 'Updating'}
				<div
					class="absolute inset-0 z-50 bg-black/80 backdrop-blur-md flex flex-col items-center justify-center text-center p-10"
				>
					<div
						class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin mb-8 shadow-lg shadow-rust/30"
					></div>
					<h3 class="text-2xl font-heading font-black text-white uppercase tracking-[0.2em] mb-3">
						Synchronizing_Core
					</h3>
					<p class="font-jetbrains text-stone-500 text-[11px] font-black uppercase tracking-widest">Downloading deployment package. System locked.</p>
				</div>
			{/if}

			<div
				class="p-8 border-b border-stone-800 flex flex-col lg:flex-row lg:items-center justify-between gap-8 bg-stone-950/40"
			>
				<div>
					<div class="flex items-center gap-4 mb-2">
						<div class="p-2.5 bg-rust/5 border border-rust/20 industrial-frame">
							<List class="w-5 h-5 text-rust-light" />
						</div>
						<h2 class="text-2xl font-heading font-black text-white uppercase tracking-tighter">Sub_Logic_Clusters</h2>
					</div>
					<p class="text-[10px] font-jetbrains font-bold text-stone-600 uppercase tracking-widest ml-12">
						Manage active instance buffers on this node.
					</p>
				</div>

				<div class="flex items-center gap-4 w-full lg:w-auto">
					<Dropdown label="CLUSTER_DIRECTIVES" Icon={List}>
						{#snippet children()}
							<button
								onclick={() => {
									dispatchBulkAction('start');
								}}
								class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains text-emerald-400 hover:bg-emerald-500/10 uppercase tracking-widest"
								>Execute_All</button
							>
							<button
								onclick={() => {
									dispatchBulkAction('stop');
								}}
								class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains text-rose-400 hover:bg-rose-500/10 uppercase tracking-widest"
								>Terminate_All</button
							>
							<button
								onclick={() => {
									dispatchBulkAction('restart');
								}}
								class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains text-rust hover:bg-rust/10 uppercase tracking-widest"
								>Reboot_All</button
							>
							<button
								onclick={() => {
									dispatchBulkAction('update');
								}}
								class="w-full text-left px-6 py-3 text-[10px] font-black font-jetbrains text-amber-400 hover:bg-amber-500/10 border-t border-stone-800 uppercase tracking-widest"
								>Patch_Buffer</button
							>
						{/snippet}
					</Dropdown>

					<button
						onclick={openSpawnDialog}
						class="flex-1 lg:flex-none flex items-center justify-center gap-3 px-10 py-3.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-xs uppercase tracking-[0.2em] transition-all shadow-xl shadow-rust/30 active:translate-y-px industrial-frame"
					>
						<Plus class="w-5 h-5" />
						Spawn
					</button>
				</div>
			</div>

			<div class="p-8 space-y-10 bg-black/20">
				<!-- Players Chart -->
				<div class="space-y-6">
					<div class="flex justify-between items-end px-2">
						<div class="space-y-2">
							<div class="flex items-center gap-3 font-jetbrains text-[9px] font-black text-stone-600 uppercase tracking-[0.4em] italic">
								<Activity class="w-3.5 h-3.5 text-rust" />
								TELEMETRY_STREAM_24H
							</div>
							<h3 class="text-sm font-heading font-black text-stone-300 uppercase tracking-widest">
								Active_Client_Load
							</h3>
						</div>
						<div class="text-3xl font-heading font-black text-rust-light tabular-nums tracking-tighter">
							{chartData[chartData.length - 1].count} <span class="text-[10px] font-jetbrains text-stone-700 tracking-widest">UNITS</span>
						</div>
					</div>
					<div
						class="h-48 sm:h-64 bg-stone-950 border border-stone-800 industrial-frame overflow-hidden relative shadow-inner p-6"
					>
						<div class="absolute inset-0 bg-[url('/grid.svg')] opacity-[0.02] pointer-events-none"></div>
						<PlayersChart data={chartData} height={200} color="var(--color-rust)" />
					</div>
				</div>

				{#if instances.length === 0}
					<div
						class="text-center py-24 text-stone-700 bg-stone-900/20 border-2 border-stone-800 border-dashed industrial-frame"
					>
						<Server class="w-16 h-16 mx-auto mb-6 opacity-20 text-stone-600" />
						<p class="text-xl font-heading font-black uppercase tracking-[0.3em] mb-2">Registry_Vacant</p>
						<p class="font-jetbrains text-[10px] font-bold uppercase tracking-widest">Initialize a new logic buffer to begin synchronization.</p>
					</div>
				{:else}
					<div class="space-y-4">
						{#each instances as instance (instance.id)}
							<InstanceRow
								{spawnerId}
								{instance}
								on:tail={handleTail}
								on:start={(e) => openInstanceActionDialog('start', e.detail.instanceId)}
								on:stop={(e) => openInstanceActionDialog('stop', e.detail.instanceId)}
								on:restart={(e) => openInstanceActionDialog('restart', e.detail.instanceId)}
								on:delete={(e) => openInstanceActionDialog('delete', e.detail.instanceId)}
								on:update={(e) => openInstanceActionDialog('update', e.detail.instanceId)}
								on:rename={(e) =>
									openInstanceActionDialog('rename', e.detail.oldId, e.detail.newId)}
							/>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Log Viewer Modal -->
{#if spawner}
	<LogViewer {spawnerId} isOpen={isLogViewerOpen} onClose={() => (isLogViewerOpen = false)} />
{/if}

<!-- Instance Console Modal -->
<InstanceManagerModal
	bind:isOpen={isConsoleOpen}
	{spawnerId}
	instanceId={consoleInstanceId}
	onClose={() => (isConsoleOpen = false)}
	memTotal={spawner?.mem_total || 0}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isSpawnDialogOpen}
	title="Spawn New Instance"
	message={`Are you sure you want to spawn a new game server instance on Spawner #${spawnerId}?`}
	confirmText="Spawn Server"
	onConfirm={executeSpawn}
/>

<!-- Instance Action Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isInstanceActionDialogOpen}
	title={instanceActionDialogTitle}
	message={instanceActionDialogMessage}
	confirmText={instanceActionConfirmText}
	onConfirm={executeInstanceAction}
	progress={actionProgress}
	statusMessage={actionStatusMessage}
/>
