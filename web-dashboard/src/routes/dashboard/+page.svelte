	<script lang="ts">
		import { onMount, onDestroy } from 'svelte';
		import { stats, nodes, notifications, isConnected, connectionStatus, siteSettings } from '$lib/stores';	import StatsCard from '$lib/components/StatsCard.svelte';
	import NodeTable from '$lib/components/NodeTable.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import DatabaseDetailModal from '$lib/components/DatabaseDetailModal.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import SystemTopology from '$lib/components/SystemTopology.svelte';
	import AddNodeModal from '$lib/components/AddNodeModal.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { Clock, Server, Activity, AlertCircle, Database, Network, Plus } from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';

	// Animation states
	let isLoaded = $state(false);
	let animateStats = $state(false);
	let mouseX = $state(0);
	let mouseY = $state(0);

	// DB Hover State
	let isDBHovered = $state(false);
	let dbHoverX = $state(0);
	let dbHoverY = $state(0);

	function handleDBMouseEnter(e: MouseEvent) {
		isDBHovered = true;
		updateDBHoverPosition(e);
	}

	function handleDBMouseLeave() {
		isDBHovered = false;
	}

	function handleDBMouseMove(e: MouseEvent) {
		updateDBHoverPosition(e);
	}

	function updateDBHoverPosition(e: MouseEvent) {
		dbHoverX = e.clientX;
		dbHoverY = e.clientY;
	}

	// Log Viewer State
	let isLogViewerOpen = $state(false);
	let selectedNodeId = $state<number | null>(null);

	// Instance Console State
	let isConsoleOpen = $state(false);
	let consoleNodeId = $state<number | null>(null);
	let consoleInstanceId = $state<string | null>(null);

	// Spawn Dialog State
	let isSpawnDialogOpen = $state(false);
	let spawnTargetId = $state<number | null>(null);

	// Instance Action Dialog State (Start/Stop)
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
	let instanceActionNodeId = $state<number | null>(null);
	let instanceActionInstanceId = $state<string | null>(null);
	let instanceActionNewID = $state<string | null>(null);
	let instanceActionBulkIds = $state<string[]>([]);
	let instanceActionDialogTitle = $state('');
	let instanceActionDialogMessage = $state('');
	let instanceActionConfirmText = $state('');
	let actionProgress = $state<number | null>(null);
	let actionStatusMessage = $state<string | null>(null);

	// Animation state for new nodes
	let previousNodeIds = new Set<number>(); // Non-reactive to prevent loop
	let highlightNewNodeId = $state<number | null>(null);
	let highlightTimeout: ReturnType<typeof setTimeout> | null = null;

	let nodeTableComponent = $state<any>(null);

	// Node Deletion State
	let isNodeDeleteDialogOpen = $state(false);
	let nodeToDeleteId = $state<number | null>(null);

	// Add Node State
	let showAddNodeModal = $state(false);

	// Reactivity for new nodes animation
	$effect(() => {
		const currentNodeIds = new Set<number>($nodes.map((s: any) => s.id));
		for (const node of $nodes) {
			if (!previousNodeIds.has(node.id) && previousNodeIds.size > 0) {
				highlightNewNodeId = node.id;
				if (highlightTimeout) clearTimeout(highlightTimeout);
				highlightTimeout = setTimeout(() => (highlightNewNodeId = null), 5000);
				break;
			}
		}
		previousNodeIds = currentNodeIds;
	});

	onMount(() => {
		// Mouse tracking for parallax effects
		const handleMouseMove = (e: MouseEvent) => {
			mouseX = (e.clientX / window.innerWidth - 0.5) * 2;
			mouseY = (e.clientY / window.innerHeight - 0.5) * 2;
		};

		window.addEventListener('mousemove', handleMouseMove, { passive: true });

		// Trigger animations using RAF for maximum smoothness
		requestAnimationFrame(() => {
			isLoaded = true;
			requestAnimationFrame(() => {
				animateStats = true;
			});
		});

		return () => {
			window.removeEventListener('mousemove', handleMouseMove);
		};
	});

	// ... (Dialog functions)
	function openSpawnDialog(event: CustomEvent<number>) {
		spawnTargetId = event.detail;
		isSpawnDialogOpen = true;
	}

	function openStartInstanceDialog(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		instanceActionType = 'start';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Start Instance';
		instanceActionDialogMessage = `Are you sure you want to start instance "${event.detail.instanceId}" on Node #${event.detail.nodeId}?`;
		instanceActionConfirmText = 'Start Instance';
		isInstanceActionDialogOpen = true;
	}

	function openRestartInstanceDialog(
		event: CustomEvent<{ nodeId: number; instanceId: string }>
	) {
		instanceActionType = 'restart';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Restart Instance';
		instanceActionDialogMessage = `Are you sure you want to restart instance "${event.detail.instanceId}"?`;
		instanceActionConfirmText = 'Restart';
		isInstanceActionDialogOpen = true;
	}

	function openStopInstanceDialog(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		instanceActionType = 'stop';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Stop Instance';
		instanceActionDialogMessage = `Are you sure you want to stop instance "${event.detail.instanceId}" on Node #${event.detail.nodeId}?`;
		instanceActionConfirmText = 'Stop Instance';
		isInstanceActionDialogOpen = true;
	}

	function openDeleteInstanceDialog(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		instanceActionType = 'delete';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Delete Instance';
		instanceActionDialogMessage = `Are you sure you want to PERMANENTLY DELETE instance "${event.detail.instanceId}" from Node #${event.detail.nodeId}? ALL FILES WILL BE LOST.`;
		instanceActionConfirmText = 'Delete Instance';
		isInstanceActionDialogOpen = true;
	}

	function openUpdateInstanceDialog(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		instanceActionType = 'update';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Update Instance';
		instanceActionDialogMessage = `Are you sure you want to reinstall game server files for instance "${event.detail.instanceId}"? This will stop the server if it's running.`;
		instanceActionConfirmText = 'Update Instance';
		isInstanceActionDialogOpen = true;
	}

	function openRenameInstanceDialog(
		event: CustomEvent<{ nodeId: number; oldId: string; newId: string }>
	) {
		instanceActionType = 'rename';
		instanceActionNodeId = event.detail.nodeId;
		instanceActionInstanceId = event.detail.oldId;
		instanceActionNewID = event.detail.newId;
		instanceActionDialogTitle = 'Rename Instance';
		instanceActionDialogMessage = `Are you sure you want to rename instance "${event.detail.oldId}" to "${event.detail.newId}"? This will stop the server if it's running.`;
		instanceActionConfirmText = 'Rename Instance';
		isInstanceActionDialogOpen = true;
	}

	function openUpdateNodeBuildDialog(event: CustomEvent<number>) {
		instanceActionType = 'update_node_build';
		instanceActionNodeId = event.detail;
		instanceActionDialogTitle = 'Update Node Build';
		instanceActionDialogMessage = `Are you sure you want Node #${instanceActionNodeId} to download the latest game server build? This might take a while.`;
		instanceActionConfirmText = 'Update Build';
		isInstanceActionDialogOpen = true;
	}

	function openBulkActionDialog(
		event: CustomEvent<{
			action: 'stop' | 'restart' | 'update';
			nodeId: number;
			instanceIds: string[];
		}>
	) {
		const { action, nodeId, instanceIds } = event.detail;
		instanceActionType = `bulk_${action}` as any;
		instanceActionNodeId = nodeId;
		instanceActionBulkIds = instanceIds;

		const actionName = action.charAt(0).toUpperCase() + action.slice(1);
		instanceActionDialogTitle = `${actionName} All Instances`;
		instanceActionDialogMessage = `Are you sure you want to ${action} ${instanceIds.length} instances on Node #${nodeId}?`;
		instanceActionConfirmText = `${actionName} All`;
		isInstanceActionDialogOpen = true;
	}

	function openDeleteNodeDialog(event: CustomEvent<number>) {
		nodeToDeleteId = event.detail;
		isNodeDeleteDialogOpen = true;
	}

	async function executeDeleteNode() {
		if (!nodeToDeleteId) return;
		try {
			const res = await fetch(`/api/nodes/${nodeToDeleteId}`, { method: 'DELETE' });
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error || 'Failed to delete node');
			}
			notifications.add({
				type: 'success',
				message: `Node #${nodeToDeleteId} deleted successfully.`
			});
		} catch (e: any) {
			console.error(e);
			notifications.add({ type: 'error', message: `Failed to delete node: ${e.message}` });
		}
		isNodeDeleteDialogOpen = false;
	}

	async function executeSpawn() {
		if (!spawnTargetId) return;

		const res = await fetch(`/api/nodes/${spawnTargetId}/spawn`, { method: 'POST' });
		if (!res.ok) {
			const err = await res.json();
			throw new Error(err.error || `Server returned ${res.status}`);
		}

		// Open console for the new instance
		const instance = await res.json();
		consoleNodeId = spawnTargetId;
		consoleInstanceId = instance.id;
		isConsoleOpen = true;
	}

	async function executeInstanceAction() {
		if (!instanceActionNodeId || !instanceActionType) return;

		let res: Response;
		try {
			if (instanceActionType === 'start') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/start`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'stop') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/stop`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'delete') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}`,
					{ method: 'DELETE' }
				);
			} else if (instanceActionType === 'update') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/update`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'rename') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/rename`,
					{
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ new_id: instanceActionNewID })
					}
				);
			} else if (instanceActionType === 'restart') {
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/stop`,
					{ method: 'POST' }
				);
				res = await fetch(
					`/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}/start`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'update_node_build') {
				res = await fetch(`/api/nodes/${instanceActionNodeId}/update-template`, {
					method: 'POST'
				});
			} else if (
				instanceActionType === 'bulk_stop' ||
				instanceActionType === 'bulk_restart' ||
				instanceActionType === 'bulk_update' ||
				instanceActionType === 'bulk_start'
			) {
				const promises = instanceActionBulkIds.map(async (id) => {
					try {
						let resUrl = '';
						if (instanceActionType === 'bulk_stop') {
							resUrl = `/api/nodes/${instanceActionNodeId}/instances/${id}/stop`;
						} else if (instanceActionType === 'bulk_start') {
							resUrl = `/api/nodes/${instanceActionNodeId}/instances/${id}/start`;
						} else if (instanceActionType === 'bulk_restart') {
							await fetch(`/api/nodes/${instanceActionNodeId}/instances/${id}/stop`, {
								method: 'POST'
							});
							resUrl = `/api/nodes/${instanceActionNodeId}/instances/${id}/start`;
						} else if (instanceActionType === 'bulk_update') {
							resUrl = `/api/nodes/${instanceActionNodeId}/instances/${id}/update`;
						}
						await fetch(resUrl, { method: 'POST' });
					} catch (e) {
						console.error(e);
					}
				});
				await Promise.all(promises);
				res = { ok: true } as Response;
			} else {
				throw new Error('Unknown action type');
			}

			if (!res.ok) {
				const errJson = await res.json().catch(() => ({}));
				throw new Error(errJson.error || `Server returned ${res.status}`);
			}
			if (instanceActionNodeId) {
				nodeTableComponent?.refreshNode(instanceActionNodeId);
			}
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: `Failed to ${instanceActionType} instance: ${e.message || 'Unknown error'}`
			});
		}
		isInstanceActionDialogOpen = false;
	}

	function handleViewLogs(event: CustomEvent<number>) {
		selectedNodeId = event.detail;
		isLogViewerOpen = true;
	}

	function handleTail(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		consoleNodeId = event.detail.nodeId;
		consoleInstanceId = event.detail.instanceId;
		isConsoleOpen = true;
	}
</script>

{#snippet statsSkeleton()}
	<div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-6 mb-10">
		{#each Array(4) as _}
			<div class="modern-industrial-card p-6 h-32 flex flex-col justify-between">
				<Skeleton width="40%" height="0.5rem" />
				<Skeleton width="70%" height="1.5rem" />
				<Skeleton width="100%" height="0.5rem" />
			</div>
		{/each}
	</div>
{/snippet}

<div class="flex flex-col lg:flex-row lg:justify-between lg:items-center mb-8 sm:mb-12 relative gap-6 sm:gap-8" class:!mb-6={$siteSettings.dashboard.compact_mode}>
	<div
		class="transform transition-all duration-700 {isLoaded
			? 'translate-x-0 opacity-100'
			: '-translate-x-4 opacity-0'}"
	>
		<div class="flex items-center gap-4 mb-2">
			<div class="h-0.5 w-6 sm:w-10 bg-rust"></div>
			<span class="font-jetbrains text-[8px] sm:text-[10px] font-black text-rust uppercase tracking-[0.3em]">System Overview // Controller</span>
		</div>
		<h1
			class="text-3xl sm:text-5xl lg:text-6xl font-heading font-black text-white tracking-tighter uppercase leading-none"
			class:!text-2xl={$siteSettings.dashboard.compact_mode}
		>
			<span class="text-rust">EXILE</span>_CONTROLLER
		</h1>
		<div class="flex flex-wrap items-center gap-3 sm:gap-4 mt-3">
			<div
				class={`px-2 sm:px-3 py-1 font-jetbrains font-bold text-[8px] sm:text-[10px] uppercase flex items-center gap-2.5 ${$isConnected ? 'bg-success/5 text-success border border-emerald-500/20' : 'bg-red-500/5 text-danger border border-red-500/20 shadow-red-900/10'}`}
			>
				<span class={`w-1.5 h-1.5 rounded-full ${$isConnected ? 'bg-success animate-pulse shadow-emerald-500/50 shadow-lg' : 'bg-danger'}`}></span>
				<Icon name={$isConnected ? 'ph:check-circle-bold' : 'ph:warning-octagon-bold'} size="0.75rem" />
				<span class="truncate max-w-[120px] sm:max-w-none">{$connectionStatus}</span>
			</div>
			<div class="w-px h-3 sm:h-4 bg-stone-800 hidden sm:block"></div>
			<span class="text-[8px] sm:text-[10px] font-jetbrains font-bold text-text-dim uppercase tracking-widest">Version 0.9.4</span>
		</div>
	</div>

	<div class="flex items-center justify-between sm:justify-end gap-3 sm:gap-6 w-full lg:w-auto">
		<!-- Add Node Button -->
		<button
			onclick={() => (showAddNodeModal = true)}
			class="group relative flex items-center justify-center gap-3 sm:gap-4 px-4 sm:px-8 py-3 sm:py-4 bg-white text-black font-black font-heading text-[10px] sm:text-xs rounded-none border-2 border-white hover:bg-rust hover:text-white hover:border-rust transition-all duration-500 shadow-2xl active:translate-y-px flex-1 sm:flex-initial {isLoaded
				? 'translate-y-0 opacity-100'
				: 'translate-y-4 opacity-0'}"
			class:!py-2={$siteSettings.dashboard.compact_mode}
		>
			<Icon name="ph:plus-bold" size="1.1rem" />
			<span class="uppercase tracking-[0.2em]">Add Node</span>
			<div class="absolute -bottom-1 -right-1 w-full h-full border-r-2 border-b-2 border-rust/30 group-hover:border-white/30 transition-colors"></div>
		</button>

		<div class="relative group">
			<NotificationBell />
		</div>
	</div>
</div>

<!-- Optimized Data Fetching Render -->
{#if !isLoaded}
	{@render statsSkeleton()}
	<div class="space-y-6">
		<Skeleton height="350px" class="opacity-20" />
		<Skeleton height="200px" class="opacity-10" />
	</div>
{:else}
	<!-- Stats Grid -->
	{#if $siteSettings.dashboard.show_stats_cards}
	<div class="grid grid-cols-1 xs:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-8 sm:mb-10" class:!mb-6={$siteSettings.dashboard.compact_mode}>
		<div
			class="transform transition-all duration-700 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.1s;"
		>
			<StatsCard title="UPTIME" value={formatUptime($stats.uptime)} iconName="clock" color="cyan" />
		</div>
		<div
			class="transform transition-all duration-700 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.2s;"
		>
			<StatsCard
				title="ACTIVE NODES"
				value={$stats.active_nodes}
				iconName="server"
				color="emerald"
			/>
		</div>
		<div
			class="transform transition-all duration-700 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.3s;"
		>
			<StatsCard
				title="TOTAL REQUESTS"
				value={$stats.total_requests}
				iconName="activity"
				color="purple"
			/>
		</div>
		<a
			href="/logs"
			class="block transition-all duration-700 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.4s;"
		>
			<StatsCard title="SYSTEM ERRORS" value={$stats.total_errors} iconName="alert" color="red" />
		</a>
	</div>
	{/if}

	<!-- Secondary Stats & Resources -->
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6 mb-8 sm:mb-10" class:!mb-6={$siteSettings.dashboard.compact_mode}>
		{#if $siteSettings.dashboard.show_traffic_card}
		<div
			class="transform transition-all duration-700 hover:scale-[1.01] {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.5s;"
		>
			<StatsCard
				title="TRAFFIC"
				value=""
				subValue={`<span class="text-warning">SENT: ${formatBytes($stats.bytes_sent)}</span> <span class="text-stone-700 mx-2">|</span> <span class="text-rust-light">RECEIVED: ${formatBytes($stats.bytes_received)}</span>`}
				iconName="ri:wifi-line"
				color="orange"
			/>
		</div>
		{/if}
		
		{#if $siteSettings.dashboard.show_db_card}
		<div
			class="transform transition-all duration-700 hover:scale-[1.01] {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.6s;"
			onmouseenter={handleDBMouseEnter}
			onmouseleave={handleDBMouseLeave}
			onmousemove={handleDBMouseMove}
			role="tooltip"
		>
			<StatsCard
				title="DATABASE"
				value={$stats.db_connected ? 'CONNECTED' : 'OFFLINE'}
				subValue={$stats.db_connected
					? `<span class="text-success">CONNS: ${$stats.db_open_connections}</span> <span class="text-stone-700 mx-2">|</span> <span class="text-rust-light">IN USE: ${$stats.db_in_use}</span>`
					: 'RECONNECTING...'}
				iconName="database"
				color={$stats.db_connected ? 'emerald' : 'red'}
			/>
		</div>
		{/if}
	</div>

	<!-- System Topology -->
	{#if $siteSettings.dashboard.show_topology}
	<div
		class="mb-8 sm:mb-10 h-[350px] sm:h-[500px] lg:h-[700px] xl:h-[800px] w-full transform transition-all duration-700 {animateStats
			? 'translate-y-0 opacity-100'
			: 'translate-y-12 opacity-0'}"
		style="animation-delay: 0.65s;"
		class:!h-[300px]={$siteSettings.dashboard.compact_mode}
	>
		<SystemTopology />
	</div>
	{/if}

	<!-- Nodes Section -->
	{#if $siteSettings.dashboard.show_nodes_table}
	<div
		class="modern-industrial-card border-stone-800 rounded-none overflow-hidden transform transition-all duration-700 hover:border-rust/30 shadow-2xl contain-paint"
		style="animation-delay: 0.7s; contain: paint layout;"
	>
		<div
			class="border-b border-stone-800 px-4 sm:px-6 py-4 sm:py-5 flex flex-col sm:flex-row justify-between items-start sm:items-center bg-[var(--card-bg-color)] backdrop-blur-xl gap-4"
			class:!py-3={$siteSettings.dashboard.compact_mode}
		>
			<div class="flex items-center gap-4">
				<div class="p-2 bg-rust/10 border border-rust/30 rounded-none industrial-frame">
					<Icon name="server" size="1.1rem" class="text-rust-light" />
				</div>
				<div>
					<h2 class="text-lg sm:text-xl font-heading font-black text-stone-100 tracking-tighter uppercase">Synchronized_Nodes</h2>
					<div class="flex items-center gap-2 mt-0.5">
						<span class="w-1.5 h-1.5 rounded-full bg-success animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.4)]"></span>
						<span class="text-[8px] sm:text-[9px] font-jetbrains font-bold text-text-dim uppercase tracking-[0.2em]">Active_Registry_Stream</span>
					</div>
				</div>
			</div>
			<div class="flex items-center gap-4 sm:gap-6 w-full sm:w-auto justify-between sm:justify-end">
				<div class="hidden xs:flex flex-col items-end">
					<span class="font-jetbrains text-[8px] sm:text-[9px] font-black text-rust-light uppercase tracking-widest">System_Archive</span>
					<span class="text-[9px] sm:text-[10px] font-heading font-black text-white uppercase tracking-widest mt-0.5">Buffer_Optimized</span>
				</div>
				<div class="h-8 sm:h-10 w-px bg-stone-800/50 hidden xs:block"></div>
				<span
					class="text-[9px] sm:text-[10px] font-heading font-black bg-rust/20 text-rust-light px-4 sm:px-5 py-1.5 sm:py-2 border border-rust/30 uppercase tracking-[0.2em] shadow-inner"
					>Real-time</span>
			</div>
		</div>
		<div class="p-0 bg-[var(--card-bg-color)] backdrop-blur-md relative">
			<div class="absolute inset-0 bg-gradient-to-b from-transparent via-rust/5 to-transparent pointer-events-none opacity-20"></div>
			<NodeTable
				bind:this={nodeTableComponent}
				nodes={$nodes}
				on:spawn={openSpawnDialog}
				on:viewLogs={handleViewLogs}
				on:startInstanceRequest={openStartInstanceDialog}
				on:stopInstanceRequest={openStopInstanceDialog}
				on:restartInstanceRequest={openRestartInstanceDialog}
				on:deleteInstanceRequest={openDeleteInstanceDialog}
				on:updateInstanceRequest={openUpdateInstanceDialog}
				on:renameInstanceRequest={openRenameInstanceDialog}
				on:updateNodeBuild={openUpdateNodeBuildDialog}
				on:bulkInstanceActionRequest={openBulkActionDialog}
				on:deleteNodeRequest={openDeleteNodeDialog}
				on:tail={handleTail}
				highlightNewNodeId={highlightNewNodeId}
			/>
		</div>
	</div>
	{/if}
{/if}

<!-- Log Drawer -->
{#if selectedNodeId}
	<LogViewer
		nodeId={selectedNodeId}
		isOpen={isLogViewerOpen}
		onClose={() => (isLogViewerOpen = false)}
	/>
{/if}

<!-- Instance Console Modal -->
<InstanceManagerModal
	bind:isOpen={isConsoleOpen}
	nodeId={consoleNodeId}
	instanceId={consoleInstanceId}
	onClose={() => (isConsoleOpen = false)}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isSpawnDialogOpen}
	title="Spawn New Instance"
	message={`Are you sure you want to spawn a new game server instance on Node #${spawnTargetId}?`}
	confirmText="Spawn Server"
	onConfirm={executeSpawn}
/>

<!-- Node Deletion Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isNodeDeleteDialogOpen}
	title="Delete Node"
	message={`Are you sure you want to delete Node #${nodeToDeleteId}? This will remove it from the registry. If it is still running, it might re-register.`}
	confirmText="Delete Node"
	isCritical={true}
	onConfirm={executeDeleteNode}
/>

<!-- Instance Action Confirmation Dialog (Start/Stop) -->
<ConfirmDialog
	bind:isOpen={isInstanceActionDialogOpen}
	title={instanceActionDialogTitle}
	message={instanceActionDialogMessage}
	confirmText={instanceActionConfirmText}
	onConfirm={executeInstanceAction}
	progress={actionProgress}
	statusMessage={actionStatusMessage}
/>

<!-- Database Detail Modal -->
<DatabaseDetailModal stats={$stats} isOpen={isDBHovered} x={dbHoverX} y={dbHoverY} />

<!-- Add Node Modal -->
<AddNodeModal bind:isOpen={showAddNodeModal} />

<style>
	@keyframes float {
		0%,
		100% {
			transform: translateY(0px) rotate(0deg);
		}
		25% {
			transform: translateY(-20px) rotate(1deg);
		}
		50% {
			transform: translateY(0px) rotate(0deg);
		}
		75% {
			transform: translateY(-10px) rotate(-1deg);
		}
	}

	@keyframes pulse-glow {
		0%,
		100% {
			box-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
		}
		50% {
			box-shadow: 0 0 40px rgba(59, 130, 246, 0.6);
		}
	}

	@keyframes slide-in-fade {
		from {
			opacity: 0;
			transform: translateY(30px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
	}

	@keyframes blob {
		0% {
			transform: translate(0px, 0px) scale(1);
		}
		33% {
			transform: translate(30px, -20px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
		100% {
			transform: translate(0px, 0px) scale(1);
		}
	}

	.animate-float {
		animation: float 6s ease-in-out infinite;
	}

	.animate-pulse-glow {
		animation: pulse-glow 2s ease-in-out infinite;
	}

	.animate-slide-fade {
		animation: slide-in-fade 0.6s ease-out forwards;
	}

	.animate-gradient {
		background-size: 200% 200%;
		animation: gradient-shift 3s ease-in-out infinite;
	}

	.animate-gradient-shift {
		background-size: 200% 200%;
		animation: gradient-shift 8s ease infinite;
	}

	.animate-blob {
		animation: blob 7s infinite;
	}

	/* Enhanced hover effects */
	.hover-glow:hover {
		box-shadow: 0 0 30px rgba(59, 130, 246, 0.4);
		transform: translateY(-2px) scale(1.02);
		transition: all 0.3s ease;
	}

	/* Parallax effect on scroll */
	.parallax-bg {
		will-change: transform;
		transform: translateZ(0);
	}

	.bg-radial-gradient {
		background: radial-gradient(
			circle at center,
			transparent 0%,
			rgba(0, 0, 0, 0.2) 50%,
			rgba(0, 0, 0, 0.6) 100%
		);
	}
</style>