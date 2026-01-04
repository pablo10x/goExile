<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { stats, nodes, notifications, isConnected, connectionStatus, siteSettings } from '$lib/stores.svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import NodeTable from '$lib/components/NodeTable.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import DatabaseDetailModal from '$lib/components/DatabaseDetailModal.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import SystemTopology from '$lib/components/SystemTopology.svelte';
	import AddNodeModal from '$lib/components/AddNodeModal.svelte';
	import QuickActionHUD from '$lib/components/QuickActionHUD.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { Clock, Server, Activity, AlertCircle, Database, Network, Plus } from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/theme/Card.svelte';

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
			<div class="modern-card p-6 h-32 flex flex-col justify-between">
				<Skeleton width="40%" height="0.5rem" />
				<Skeleton width="70%" height="1.5rem" />
				<Skeleton width="100%" height="0.5rem" />
			</div>
		{/each}
	</div>
{/snippet}

<PageHeader 
    title="Overview" 
    subtitle="System Dashboard" 
    icon="ph:gauge-bold"
>
    {#snippet actions()}
        <div class="flex items-center gap-4">
            <Button
                variant="primary"
                size="lg"
                onclick={() => (showAddNodeModal = true)}
                icon="ph:plus-bold"
            >
                ADD NODE
            </Button>
            <div class="relative group">
                <NotificationBell />
            </div>
        </div>
    {/snippet}
</PageHeader>

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
	<div class="grid grid-cols-1 xs:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-8 sm:mb-10">
		<div
			class="transform transition-all duration-700 {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
			style="animation-delay: 0.1s;"
		>
			<StatsCard title="UPTIME" value={formatUptime($stats.uptime)} iconName="ph:clock-bold" color="cyan" />
		</div>
		<div
			class="transform transition-all duration-700 {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
			style="animation-delay: 0.2s;"
		>
			<StatsCard
				title="ACTIVE NODES"
				value={$stats.active_nodes}
				iconName="ph:server-bold"
				color="emerald"
			/>
		</div>
		<div
			class="transform transition-all duration-700 {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
			style="animation-delay: 0.3s;"
		>
			<StatsCard
				title="TOTAL REQUESTS"
				value={$stats.total_requests}
				iconName="ph:activity-bold"
				color="purple"
			/>
		</div>
		<a
			href="/logs"
			class="block transition-all duration-700 {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
			style="animation-delay: 0.4s;"
		>
			<StatsCard title="SYSTEM ERRORS" value={$stats.total_errors} iconName="ph:warning-octagon-bold" color="red" />
		</a>
	</div>
	{/if}

	<!-- Secondary Stats & Resources -->
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6 mb-8 sm:mb-10">
		{#if $siteSettings.dashboard.show_traffic_card}
		<div
			class="transform transition-all duration-700 hover:scale-[1.01] {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
			style="animation-delay: 0.5s;"
		>
			<StatsCard
				title="TRAFFIC"
				value=""
				subValue={`<span class="text-rust-light">SENT: ${formatBytes($stats.bytes_sent)}</span> <span class="text-neutral-700 mx-2">|</span> <span class="text-rust-light">RECEIVED: ${formatBytes($stats.bytes_received)}</span>`}
				iconName="ph:wifi-high-bold"
				color="orange"
			/>
		</div>
		{/if}
		
		{#if $siteSettings.dashboard.show_db_card}
		<div
			class="transform transition-all duration-700 hover:scale-[1.01] {animateStats
				? 'tranneutral-y-0 opacity-100'
				: 'tranneutral-y-8 opacity-0'}"
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
					? `<span class="text-emerald-400">CONNS: ${$stats.db_open_connections}</span> <span class="text-neutral-700 mx-2">|</span> <span class="text-rust-light">IN USE: ${$stats.db_in_use}</span>`
					: 'RECONNECTING...'}
				iconName="ph:database-bold"
				color={$stats.db_connected ? 'emerald' : 'red'}
			/>
		</div>
		{/if}
	</div>

	<!-- Quick Action HUD -->
	<div 
		class="mb-10 transform transition-all duration-700 {animateStats ? 'tranneutral-y-0 opacity-100' : 'tranneutral-y-8 opacity-0'}"
		style="animation-delay: 0.62s;"
	>
		<QuickActionHUD 
			on:tail={handleTail}
			on:stop={(e: any) => openStopInstanceDialog({ detail: e.detail } as any)}
		/>
	</div>

	<!-- System Topology -->
	{#if $siteSettings.dashboard.show_topology}
	<div
		class="mb-8 sm:mb-10 h-[350px] sm:h-[500px] lg:h-[700px] xl:h-[800px] w-full transform transition-all duration-700 {animateStats
			? 'tranneutral-y-0 opacity-100'
			: 'tranneutral-y-12 opacity-0'}"
		style="animation-delay: 0.65s;"
	>
		<SystemTopology />
	</div>
	{/if}

	<!-- Nodes Section -->
	{#if $siteSettings.dashboard.show_nodes_table}
        <Card 
            title="Node Infrastructure" 
            subtitle="Active Node Stream"
            icon="ph:server-bold"
            class="transform transition-all duration-700 shadow-sm"
        >
            {#snippet actions()}
                <div class="flex items-center gap-4 sm:gap-6 w-full sm:w-auto justify-between sm:justify-end">
                    <div class="hidden xs:flex flex-col items-end text-right">
                        <span class="text-[8px] sm:text-[9px] font-black text-neutral-600 uppercase tracking-widest italic">Inventory_Status</span>
                        <span class="text-[9px] sm:text-[10px] font-black text-white uppercase tracking-widest mt-0.5">SYNCED</span>
                    </div>
                    <div class="h-8 sm:h-10 w-px bg-neutral-800 hidden xs:block"></div>
                    <span
                        class="text-[9px] sm:text-[10px] font-bold bg-rust/10 text-rust-light px-4 sm:px-5 py-1.5 sm:py-2 border border-rust/20 rounded-none uppercase tracking-widest shadow-sm"
                        >Live_Stream</span>
                </div>
            {/snippet}

            <div class="p-0 bg-transparent relative">
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
        </Card>
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
</style>