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
	import { serverVersions, notifications } from '$lib/stores.svelte';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from '$lib/components/PlayersChart.svelte';
	import { Server, Cpu, HardDrive, MemoryStick, List, Plus, ArrowLeft, AlertCircle, Activity, ShieldAlert, ShieldCheck } from 'lucide-svelte';

	const nodeId = parseInt($page.params.id || '0');

	let node = $state<any>(null);
	let instances = $state<any[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let refreshInterval: any;
	let togglingDrain = $state(false);

	// Dummy Data for Chart
	const chartData = Array.from({ length: 24 }, (_, i) => {
		const time = new Date().getTime() - (23 - i) * 3600000;
		return {
			timestamp: time,
			count: Math.floor(Math.max(0, 50 + Math.sin(i / 3) * 30 + (Math.random() - 0.5) * 20))
		};
	});

	// Instance Action State
	let isInstanceActionDialogOpen = $state(false);
	let instanceActionType = $state<
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
		| 'update_node_build'
		| null
	>(null);
	let instanceActionInstanceId = $state<string | null>(null);
	let instanceActionNewID = $state<string | null>(null);
	let instanceActionBulkIds = $state<string[]>([]);
	let instanceActionDialogTitle = $state('');
	let instanceActionDialogMessage = $state('');
	let instanceActionConfirmText = $state('');

	// Progress State for ConfirmDialog
	let actionProgress = $state<number | null>(null);
	let actionStatusMessage = $state<string | null>(null);

	// Spawn Dialog
	let isSpawnDialogOpen = $state(false);

	// Console & Logs
	let isConsoleOpen = $state(false);
	let consoleInstanceId = $state<string | null>(null);
	let isLogViewerOpen = $state(false);

	let activeVersion = $derived($serverVersions.find((v) => v.is_active));

	async function fetchNodeData() {
		try {
			const res = await fetch(`/api/nodes/${nodeId}`);
			if (!res.ok) {
				if (res.status === 404) throw new Error('Node not found');
				throw new Error('Failed to load node details');
			}
			node = await res.json();

			const instRes = await fetch(`/api/nodes/${nodeId}/instances`);
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

	async function toggleDrainMode() {
		if (!node) return;
		togglingDrain = true;
		try {
			const res = await fetch(`/api/nodes/${nodeId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					region: node.region,
					max_instances: node.max_instances,
					is_draining: !node.is_draining
				})
			});
			if (res.ok) {
				notifications.add({
					type: 'success',
					message: `Drain Mode ${!node.is_draining ? 'enabled' : 'disabled'}`,
					details: !node.is_draining ? 'No new instances will be spawned on this node.' : 'Node is now accepting new instances.'
				});
				await fetchNodeData();
			} else {
				throw new Error('Failed to update node settings');
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: e.message });
		} finally {
			togglingDrain = false;
		}
	}

	onMount(() => {
		fetchNodeData();
		refreshInterval = setInterval(fetchNodeData, 5000); // Poll every 5s

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
			const res = await fetch(`/api/nodes/${nodeId}/spawn`, { method: 'POST' });
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || `Server returned ${res.status}`);
			}
			const instance = await res.json();
			consoleInstanceId = instance.id;
			isConsoleOpen = true;
			fetchNodeData();
		} catch (e: any) {
			alert(e.message);
		}
		isSpawnDialogOpen = false;
	}

	function openUpdateNodeBuildDialog() {
		instanceActionType = 'update_node_build';
		instanceActionDialogTitle = 'Update Node Build';
		instanceActionDialogMessage = `Are you sure you want Node #${nodeId} to download the latest game server build? This might take a while.`;
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

			if (instanceActionType === 'update_node_build') {
				actionStatusMessage = 'Requesting node update...';
				res = await fetch(`/api/nodes/${nodeId}/update-template`, { method: 'POST' });
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
							await fetch(`/api/nodes/${nodeId}/instances/${id}/stop`, { method: 'POST' });
							// Small delay between stop and start?
							url = `/api/nodes/${nodeId}/instances/${id}/start`;
						} else {
							url = `/api/nodes/${nodeId}/instances/${id}/${action}`;
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
				let url = `/api/nodes/${nodeId}/instances/${instanceActionInstanceId}`;
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

			fetchNodeData();
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
					Node_<span class="text-rust">#{nodeId}</span>
				</h1>
				{#if node}
					<div class="flex items-center gap-4 mt-3">
						<div
							class={`px-3 py-1 font-jetbrains font-bold text-[10px] uppercase flex items-center gap-2.5 border ${getStatusClass(node.status)}`}
						>
							<span class={`w-1.5 h-1.5 rounded-full ${node.status === 'Online' || node.status === 'active' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span>
							{node.status}
						</div>
						<div class="w-px h-4 bg-stone-800"></div>
						{#if node.is_draining}
							<div class="flex items-center gap-2 px-3 py-1 bg-amber-500/10 text-amber-500 border border-amber-500/20 text-[10px] font-black uppercase tracking-widest">
								<ShieldAlert class="w-3.5 h-3.5 animate-pulse" />
								DRAINING
							</div>
						{/if}
						<span class="text-[10px] font-jetbrains font-bold text-stone-600 uppercase tracking-widest">Binary_Rev: v{node.game_version || '0.0.0'}</span>
					</div>
				{/if}
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-4">
			{#if node}
				<button
					onclick={toggleDrainMode}
					disabled={togglingDrain}
					class={`px-8 py-3 font-heading font-black text-[11px] uppercase tracking-widest transition-all border shadow-xl whitespace-nowrap active:tranneutral-y-px flex items-center gap-3 ${node.is_draining ? 'bg-emerald-950/20 text-emerald-500 border-emerald-900/30 hover:bg-emerald-600 hover:text-white' : 'bg-amber-950/20 text-amber-500 border-amber-900/30 hover:bg-amber-600 hover:text-white'}`}
				>
					{#if node.is_draining}
						<ShieldCheck class="w-4 h-4" />
						Resume_Node
					{:else}
						<ShieldAlert class="w-4 h-4" />
						Drain_Node
					{/if}
				</button>
			{/if}

			<button
				onclick={() => (isLogViewerOpen = true)}
				class="px-8 py-3 bg-stone-900 hover:bg-white hover:text-black text-stone-400 font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800 active:tranneutral-y-px shadow-xl"
			>
				Console_Output
			</button>

			{#if node && activeVersion}
				{@const cmp = compareVersions(activeVersion.version, node.game_version || '0.0.0')}
				{#if cmp !== 0}
					<button
						onclick={() => openUpdateNodeBuildDialog()}
						disabled={node.status === 'Updating'}
						class={`px-8 py-3 font-heading font-black text-[11px] uppercase tracking-widest transition-all border shadow-xl whitespace-nowrap active:tranneutral-y-px ${node.status === 'Updating' ? 'bg-stone-800 text-stone-600 border-stone-700 cursor-not-allowed' : cmp > 0 ? 'bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-400 shadow-emerald-900/20' : 'bg-rust hover:bg-rust-light text-white border-rust-light shadow-rust/20'}`}
					>
						{node.status === 'Updating' ? 'Synchronizing...' : cmp > 0 ? 'Apply_Patch' : 'Revert_Rev'}
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

	{#if isLoading && !node}
		<div class="flex flex-col items-center justify-center h-96 gap-6">
			<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
			<p class="font-heading font-black text-[12px] text-rust animate-pulse uppercase tracking-[0.5em]">Establishing_Uplink...</p>
		</div>
	{:else if error}
		<div class="p-12 text-center bg-red-950/10 border border-red-900/30 industrial-frame shadow-2xl">
			<AlertCircle class="w-16 h-16 text-red-600 mx-auto mb-6 animate-pulse" />
			<h2 class="text-2xl font-heading font-black text-red-500 mb-3 uppercase tracking-widest">Terminal_Connection_Fault</h2>
			<p class="font-jetbrains text-stone-500 font-bold uppercase tracking-tight">{error}</p>
			<button class="mt-10 px-10 py-3 bg-red-600 hover:bg-red-500 text-white font-heading font-black text-[11px] uppercase tracking-widest transition-all shadow-lg" onclick={fetchNodeData}>Retry_Protocol</button>
		</div>
	{:else if node}
		<!-- Node Stats -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-10">
			<StatsCard
				title="Active_Sub_Nodes"
				value={`${node.current_instances} / ${node.max_instances}`}
				subValue={`CAPACITY_LOAD: ${node.max_instances > 0 ? ((node.current_instances / node.max_instances) * 100).toFixed(0) : 0}%`}
				Icon={Server}
				color="cyan"
			/>
			<StatsCard
				title="Core_Utilization"
				value={`${node.cpu_usage?.toFixed(1) || 0}%`}
				Icon={Cpu}
				color="purple"
			/>
			<StatsCard
				title="Volatile_Memory"
				value={formatBytes(node.mem_used || 0)}
				subValue={`REGISTRY_CAP: ${formatBytes(node.mem_total || 0)}`}
				Icon={MemoryStick}
				color="emerald"
			/>
			<StatsCard
				title="Storage_Array"
				value={formatBytes(node.disk_used || 0)}
				subValue={`ARRAY_TOTAL: ${formatBytes(node.disk_total || 0)}`}
				Icon={HardDrive}
				color="orange"
			/>
		</div>

		<!-- Instances Section -->
		<div
			class="modern-industrial-card glass-panel !rounded-none overflow-hidden relative shadow-2xl"
		>
			{#if node.status === 'Updating'}
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
						class="flex-1 lg:flex-none flex items-center justify-center gap-3 px-10 py-3.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-xs uppercase tracking-[0.2em] transition-all shadow-xl shadow-rust/30 active:tranneutral-y-px industrial-frame"
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
								{nodeId}
								{instance}
								on:tail={handleTail}
								on:start={(e: any) => openInstanceActionDialog('start', e.detail.instanceId)}
								on:stop={(e: any) => openInstanceActionDialog('stop', e.detail.instanceId)}
								on:restart={(e: any) => openInstanceActionDialog('restart', e.detail.instanceId)}
								on:delete={(e: any) => openInstanceActionDialog('delete', e.detail.instanceId)}
								on:update={(e: any) => openInstanceActionDialog('update', e.detail.instanceId)}
								on:rename={(e: any) =>
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
{#if node}
	<LogViewer {nodeId} isOpen={isLogViewerOpen} onClose={() => (isLogViewerOpen = false)} />
{/if}

<!-- Instance Console Modal -->
<InstanceManagerModal
	bind:isOpen={isConsoleOpen}
	{nodeId}
	instanceId={consoleInstanceId}
	onClose={() => (isConsoleOpen = false)}
	memTotal={node?.mem_total || 0}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isSpawnDialogOpen}
	title="Spawn New Instance"
	message={`Are you sure you want to spawn a new game server instance on Node #${nodeId}?`}
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
