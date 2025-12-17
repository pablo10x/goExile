<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { stats, spawners, notifications, isConnected, connectionStatus } from '$lib/stores';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import SpawnerTable from '$lib/components/SpawnerTable.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import DatabaseDetailModal from '$lib/components/DatabaseDetailModal.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { Clock, Server, Activity, AlertCircle, Database, Network } from 'lucide-svelte';

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
	let selectedSpawnerId = $state<number | null>(null);

	// Instance Console State
	let isConsoleOpen = $state(false);
	let consoleSpawnerId = $state<number | null>(null);
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
		| 'update_spawner_build'
		| null
	>(null);
	let instanceActionSpawnerId = $state<number | null>(null);
	let instanceActionInstanceId = $state<string | null>(null);
	let instanceActionNewID = $state<string | null>(null);
	let instanceActionBulkIds = $state<string[]>([]);
	let instanceActionDialogTitle = $state('');
	let instanceActionDialogMessage = $state('');
	let instanceActionConfirmText = $state('');
	let actionProgress = $state<number | null>(null);
	let actionStatusMessage = $state<string | null>(null);

	// Animation state for new spawners
	let previousSpawnerIds = new Set<number>(); // Non-reactive to prevent loop
	let highlightNewSpawnerId = $state<number | null>(null);
	let highlightTimeout: ReturnType<typeof setTimeout> | null = null;

	let spawnerTableComponent = $state<any>(null);

	// Spawner Deletion State
	let isSpawnerDeleteDialogOpen = $state(false);
	let spawnerToDeleteId = $state<number | null>(null);

	// Reactivity for new spawners animation
	$effect(() => {
		const currentSpawnerIds = new Set<number>($spawners.map((s: any) => s.id));
		for (const spawner of $spawners) {
			if (!previousSpawnerIds.has(spawner.id) && previousSpawnerIds.size > 0) {
				highlightNewSpawnerId = spawner.id;
				if (highlightTimeout) clearTimeout(highlightTimeout);
				highlightTimeout = setTimeout(() => (highlightNewSpawnerId = null), 5000);
				break;
			}
		}
		previousSpawnerIds = currentSpawnerIds;
	});

	onMount(() => {
		// Mouse tracking for parallax effects
		const handleMouseMove = (e: MouseEvent) => {
			mouseX = (e.clientX / window.innerWidth - 0.5) * 2;
			mouseY = (e.clientY / window.innerHeight - 0.5) * 2;
		};

		window.addEventListener('mousemove', handleMouseMove);

		// Trigger animations after component mounts
		setTimeout(() => {
			isLoaded = true;
			setTimeout(() => {
				animateStats = true;
			}, 200);
		}, 100);

		return () => {
			window.removeEventListener('mousemove', handleMouseMove);
		};
	});

	// ... (Dialog functions)
	function openSpawnDialog(event: CustomEvent<number>) {
		spawnTargetId = event.detail;
		isSpawnDialogOpen = true;
	}

	function openStartInstanceDialog(event: CustomEvent<{ spawnerId: number; instanceId: string }>) {
		instanceActionType = 'start';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Start Instance';
		instanceActionDialogMessage = `Are you sure you want to start instance "${event.detail.instanceId}" on Spawner #${event.detail.spawnerId}?`;
		instanceActionConfirmText = 'Start Instance';
		isInstanceActionDialogOpen = true;
	}

	function openRestartInstanceDialog(
		event: CustomEvent<{ spawnerId: number; instanceId: string }>
	) {
		instanceActionType = 'restart';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Restart Instance';
		instanceActionDialogMessage = `Are you sure you want to restart instance "${event.detail.instanceId}"?`;
		instanceActionConfirmText = 'Restart';
		isInstanceActionDialogOpen = true;
	}

	function openStopInstanceDialog(event: CustomEvent<{ spawnerId: number; instanceId: string }>) {
		instanceActionType = 'stop';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Stop Instance';
		instanceActionDialogMessage = `Are you sure you want to stop instance "${event.detail.instanceId}" on Spawner #${event.detail.spawnerId}?`;
		instanceActionConfirmText = 'Stop Instance';
		isInstanceActionDialogOpen = true;
	}

	function openDeleteInstanceDialog(event: CustomEvent<{ spawnerId: number; instanceId: string }>) {
		instanceActionType = 'delete';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Delete Instance';
		instanceActionDialogMessage = `Are you sure you want to PERMANENTLY DELETE instance "${event.detail.instanceId}" from Spawner #${event.detail.spawnerId}? ALL FILES WILL BE LOST.`;
		instanceActionConfirmText = 'Delete Instance';
		isInstanceActionDialogOpen = true;
	}

	function openUpdateInstanceDialog(event: CustomEvent<{ spawnerId: number; instanceId: string }>) {
		instanceActionType = 'update';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.instanceId;
		instanceActionDialogTitle = 'Update Instance';
		instanceActionDialogMessage = `Are you sure you want to reinstall game server files for instance "${event.detail.instanceId}"? This will stop the server if it's running.`;
		instanceActionConfirmText = 'Update Instance';
		isInstanceActionDialogOpen = true;
	}

	function openRenameInstanceDialog(
		event: CustomEvent<{ spawnerId: number; oldId: string; newId: string }>
	) {
		instanceActionType = 'rename';
		instanceActionSpawnerId = event.detail.spawnerId;
		instanceActionInstanceId = event.detail.oldId;
		instanceActionNewID = event.detail.newId;
		instanceActionDialogTitle = 'Rename Instance';
		instanceActionDialogMessage = `Are you sure you want to rename instance "${event.detail.oldId}" to "${event.detail.newId}"? This will stop the server if it's running.`;
		instanceActionConfirmText = 'Rename Instance';
		isInstanceActionDialogOpen = true;
	}

	function openUpdateSpawnerBuildDialog(event: CustomEvent<number>) {
		instanceActionType = 'update_spawner_build';
		instanceActionSpawnerId = event.detail;
		instanceActionDialogTitle = 'Update Spawner Build';
		instanceActionDialogMessage = `Are you sure you want Spawner #${instanceActionSpawnerId} to download the latest game server build? This might take a while.`;
		instanceActionConfirmText = 'Update Build';
		isInstanceActionDialogOpen = true;
	}

	function openBulkActionDialog(
		event: CustomEvent<{
			action: 'stop' | 'restart' | 'update';
			spawnerId: number;
			instanceIds: string[];
		}>
	) {
		const { action, spawnerId, instanceIds } = event.detail;
		instanceActionType = `bulk_${action}` as any;
		instanceActionSpawnerId = spawnerId;
		instanceActionBulkIds = instanceIds;

		const actionName = action.charAt(0).toUpperCase() + action.slice(1);
		instanceActionDialogTitle = `${actionName} All Instances`;
		instanceActionDialogMessage = `Are you sure you want to ${action} ${instanceIds.length} instances on Spawner #${spawnerId}?`;
		instanceActionConfirmText = `${actionName} All`;
		isInstanceActionDialogOpen = true;
	}

	function openDeleteSpawnerDialog(event: CustomEvent<number>) {
		spawnerToDeleteId = event.detail;
		isSpawnerDeleteDialogOpen = true;
	}

	async function executeDeleteSpawner() {
		if (!spawnerToDeleteId) return;
		try {
			const res = await fetch(`/api/spawners/${spawnerToDeleteId}`, { method: 'DELETE' });
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				throw new Error(err.error || 'Failed to delete spawner');
			}
		} catch (e: any) {
			console.error(e);
			notifications.add({ type: 'error', message: `Failed to delete spawner: ${e.message}` });
		}
		isSpawnerDeleteDialogOpen = false;
	}

	async function executeSpawn() {
		if (!spawnTargetId) return;

		const res = await fetch(`/api/spawners/${spawnTargetId}/spawn`, { method: 'POST' });
		if (!res.ok) {
			const err = await res.json();
			throw new Error(err.error || `Server returned ${res.status}`);
		}

		// Open console for the new instance
		const instance = await res.json();
		consoleSpawnerId = spawnTargetId;
		consoleInstanceId = instance.id;
		isConsoleOpen = true;
	}

	async function executeInstanceAction() {
		if (!instanceActionSpawnerId || !instanceActionType) return;

		let res: Response;
		try {
			if (instanceActionType === 'start') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/start`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'stop') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/stop`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'delete') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}`,
					{ method: 'DELETE' }
				);
			} else if (instanceActionType === 'update') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/update`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'rename') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/rename`,
					{
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ new_id: instanceActionNewID })
					}
				);
			} else if (instanceActionType === 'restart') {
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/stop`,
					{ method: 'POST' }
				);
				res = await fetch(
					`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/start`,
					{ method: 'POST' }
				);
			} else if (instanceActionType === 'update_spawner_build') {
				res = await fetch(`/api/spawners/${instanceActionSpawnerId}/update-template`, {
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
							resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/stop`;
						} else if (instanceActionType === 'bulk_start') {
							resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/start`;
						} else if (instanceActionType === 'bulk_restart') {
							await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${id}/stop`, {
								method: 'POST'
							});
							resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/start`;
						} else if (instanceActionType === 'bulk_update') {
							resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/update`;
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
			if (instanceActionSpawnerId) {
				spawnerTableComponent?.refreshSpawner(instanceActionSpawnerId);
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
		selectedSpawnerId = event.detail;
		isLogViewerOpen = true;
	}

	function handleTail(event: CustomEvent<{ spawnerId: number; instanceId: string }>) {
		consoleSpawnerId = event.detail.spawnerId;
		consoleInstanceId = event.detail.instanceId;
		isConsoleOpen = true;
	}
</script>

<div class="flex justify-between items-center mb-6 relative">
	<div
		class="transform transition-all duration-700 {isLoaded
			? 'translate-x-0 opacity-100'
			: '-translate-x-4 opacity-0'}"
	>
		<h1
			class="text-3xl font-bold text-slate-50 bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent animate-pulse"
		>
			Dashboard
		</h1>
		<div class="flex items-center gap-2 mt-1">
			<div
				class={`w-2 h-2 rounded-full ${$isConnected ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'} shadow-lg`}
			></div>
			<span class="text-xs font-mono text-slate-400 backdrop-blur-sm">{$connectionStatus}</span>
		</div>
	</div>

	<div class="flex items-center gap-4">
		<NotificationBell />
		<div
			class="text-slate-500 text-sm transform transition-all duration-700 delay-100 {isLoaded
				? 'translate-x-0 opacity-100'
				: 'translate-x-4 opacity-0'}"
		>
			{new Date().toLocaleDateString()}
		</div>
	</div>
</div>

<!-- Stats Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
	<div
		class="transform transition-all duration-700 hover:scale-105 {animateStats
			? 'translate-y-0 opacity-100'
			: 'translate-y-8 opacity-0'}"
		style="animation-delay: 0.1s;"
	>
		<StatsCard title="Uptime" value={formatUptime($stats.uptime)} Icon={Clock} color="blue" />
	</div>
	<div
		class="transform transition-all duration-700 hover:scale-105 {animateStats
			? 'translate-y-0 opacity-100'
			: 'translate-y-8 opacity-0'}"
		style="animation-delay: 0.2s;"
	>
		<StatsCard
			title="Active Spawners"
			value={$stats.active_spawners}
			Icon={Server}
			color="emerald"
		/>
	</div>
	<div
		class="transform transition-all duration-700 hover:scale-105 {animateStats
			? 'translate-y-0 opacity-100'
			: 'translate-y-8 opacity-0'}"
		style="animation-delay: 0.3s;"
	>
		<StatsCard
			title="Total Requests"
			value={$stats.total_requests}
			Icon={Activity}
			color="purple"
		/>
	</div>
	<a
		href="/errors"
		class="block transition-all duration-700 hover:scale-[1.05] hover:rotate-1 {animateStats
			? 'translate-y-0 opacity-100'
			: 'translate-y-8 opacity-0'}"
		style="animation-delay: 0.4s;"
	>
		<StatsCard title="Total Errors" value={$stats.total_errors} Icon={AlertCircle} color="red" />
	</a>
</div>

<!-- Secondary Stats & Resources -->
<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
	<div class="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-6 h-full">
		<div
			class="transform transition-all duration-700 hover:scale-105 hover:-translate-y-1 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.5s;"
		>
			<StatsCard
				title="Network Traffic"
				value=""
				subValue={`<span class="text-orange-400">↑ ${formatBytes($stats.bytes_sent)}</span> <span class="text-slate-600 mx-2">|</span> <span class="text-cyan-400">↓ ${formatBytes($stats.bytes_received)}</span>`}
				Icon={Network}
				color="orange"
			/>
		</div>
		<div
			class="transform transition-all duration-700 hover:scale-105 hover:-translate-y-1 {animateStats
				? 'translate-y-0 opacity-100'
				: 'translate-y-8 opacity-0'}"
			style="animation-delay: 0.6s;"
			onmouseenter={handleDBMouseEnter}
			onmouseleave={handleDBMouseLeave}
			onmousemove={handleDBMouseMove}
			role="tooltip"
		>
			<StatsCard
				title="Database Status"
				value={$stats.db_connected ? 'Connected' : 'Disconnected'}
				subValue={$stats.db_connected
					? `<span class="text-emerald-400">${$stats.db_open_connections} Open</span> <span class="text-slate-600 mx-2">|</span> <span class="text-blue-400">${$stats.db_in_use} In Use</span>`
					: ''}
				Icon={Database}
				color={$stats.db_connected ? 'emerald' : 'red'}
			/>
		</div>
	</div>
</div>

<!-- Spawners Section -->
<div
	class="card bg-slate-800/60 backdrop-blur-sm border border-slate-700/50 rounded-xl overflow-hidden transform transition-all duration-700 hover:scale-[1.01] hover:border-blue-500/20 hover:shadow-2xl hover:shadow-blue-500/5 {animateStats
		? 'translate-y-0 opacity-100'
		: 'translate-y-12 opacity-0'}"
	style="animation-delay: 0.7s;"
>
	<div
		class="border-b border-slate-700/50 px-6 py-4 flex justify-between items-center bg-gradient-to-r from-slate-800/80 to-slate-800/60 backdrop-blur-sm"
	>
		<div class="flex items-center gap-3">
			<div
				class="w-3 h-3 bg-blue-500 rounded-full animate-pulse shadow-lg shadow-blue-500/50"
			></div>
			<h2 class="text-xl font-bold text-slate-50">📦 Registered Spawners</h2>
		</div>
		<span
			class="text-xs text-slate-500 uppercase tracking-widest font-semibold bg-slate-900/80 px-3 py-1 rounded-full border border-slate-700"
			>Real-time</span
		>
	</div>
	<div class="p-0 bg-slate-900/40">
		<SpawnerTable
			bind:this={spawnerTableComponent}
			spawners={$spawners}
			on:spawn={openSpawnDialog}
			on:viewLogs={handleViewLogs}
			on:startInstanceRequest={openStartInstanceDialog}
			on:stopInstanceRequest={openStopInstanceDialog}
			on:restartInstanceRequest={openRestartInstanceDialog}
			on:deleteInstanceRequest={openDeleteInstanceDialog}
			on:updateInstanceRequest={openUpdateInstanceDialog}
			on:renameInstanceRequest={openRenameInstanceDialog}
			on:updateSpawnerBuild={openUpdateSpawnerBuildDialog}
			on:bulkInstanceActionRequest={openBulkActionDialog}
			on:deleteSpawnerRequest={openDeleteSpawnerDialog}
			on:tail={handleTail}
			{highlightNewSpawnerId}
		/>
	</div>
</div>

<!-- Log Drawer -->
{#if selectedSpawnerId}
	<LogViewer
		spawnerId={selectedSpawnerId}
		isOpen={isLogViewerOpen}
		onClose={() => (isLogViewerOpen = false)}
	/>
{/if}

<!-- Instance Console Modal -->
<InstanceManagerModal
	bind:isOpen={isConsoleOpen}
	spawnerId={consoleSpawnerId}
	instanceId={consoleInstanceId}
	onClose={() => (isConsoleOpen = false)}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isSpawnDialogOpen}
	title="Spawn New Instance"
	message={`Are you sure you want to spawn a new game server instance on Spawner #${spawnTargetId}?`}
	confirmText="Spawn Server"
	onConfirm={executeSpawn}
/>

<!-- Spawner Deletion Confirmation Dialog -->
<ConfirmDialog
	bind:isOpen={isSpawnerDeleteDialogOpen}
	title="Delete Spawner"
	message={`Are you sure you want to delete Spawner #${spawnerToDeleteId}? This will remove it from the registry. If it is still running, it might re-register.`}
	confirmText="Delete Spawner"
	isCritical={true}
	onConfirm={executeDeleteSpawner}
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
