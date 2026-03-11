<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { fade, scale, fly, slide } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { serverVersions } from '$lib/stores.svelte';
	import Console from './Console.svelte';
	import ResourceMetricsPanel from './ResourceMetricsPanel.svelte';
	import ConfirmDialog from './ConfirmDialog.svelte';
	import LogViewer from './LogViewer.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';

	interface Props {
		isOpen: boolean;
		nodeId: number | null;
		instanceId: string | null;
		onClose: () => void;
		memTotal?: number;
	}

	let {
		isOpen = $bindable(false),
		nodeId = null,
		instanceId = null,
		onClose,
		memTotal = 0
	}: Props = $props();

	// Component State
	let logs = $state<string[]>([]);
	let stats = $state({
		cpu_percent: 0,
		memory_usage: 0,
		disk_usage: 0,
		status: 'Unknown',
		uptime: 0
	});

	type TabType = 'console' | 'metrics' | 'backups' | 'history' | 'node_logs';
	let activeTab = $state<TabType>('console');

	const tabs: { id: TabType; label: string }[] = [
		{ id: 'console', label: 'Access Console' },
		{ id: 'metrics', label: 'Performance' },
		{ id: 'backups', label: 'Backups' },
		{ id: 'history', label: 'Activity' },
		{ id: 'node_logs', label: 'Host Logs' }
	];

	// Data States
	let backups = $state<any[]>([]);
	let historyLogs = $state<any[]>([]);
	let isLoadingData = $state(false);

	// Confirm Dialog State
	let isConfirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let confirmBtnText = $state('Confirm');
	let isCriticalAction = $state(false);
	let pendingBackupAction = $state<() => Promise<void>>(async () => {});

	// Polling Intervals
	let statsInterval: ReturnType<typeof setInterval> | null = null;
	let logsInterval: ReturnType<typeof setInterval> | null = null;

	// Deployment Logic
	let isDeploying = $state(false);
	let deploymentStep = $state(0);
	const deploymentSteps = [
		'Allocating resources...',
		'Downloading build files...',
		'Configuring environment...',
		'Starting process...'
	];

	// Derived State
	let activeVersion = $derived($serverVersions.find((v) => v.is_active));
	let memoryPercent = $derived(
		memTotal ? Math.min(100, (stats.memory_usage / (memTotal * 1024 * 1024)) * 100) : 0
	);

	function getBackupVersion(filename: string): string | null {
		const match = filename.match(/_v(.*?)\.zip$/);
		return match ? match[1] : null;
	}

	// API Functions
	async function fetchBackups() {
		if (!nodeId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await apiFetch(`/api/nodes/${nodeId}/instances/${instanceId}/backups`);
			if (res.ok) {
				const data = await res.json();
				backups = (data.backups || []).sort(
					(a: any, b: any) => new Date(b.date).getTime() - new Date(a.date).getTime()
				);
			}
		} catch (e) {
			console.error(e);
		} finally {
			isLoadingData = false;
		}
	}

	async function fetchHistoryLogs() {
		if (!nodeId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await apiFetch(`/api/nodes/${nodeId}/instances/${instanceId}/history`);
			if (res.ok) historyLogs = await res.json();
		} catch (e) {
			console.error(e);
		} finally {
			isLoadingData = false;
		}
	}

	async function fetchInstanceLogs() {
		if (!nodeId || !instanceId) return;
		try {
			const res = await apiFetch(`/api/nodes/${nodeId}/instances/${instanceId}/logs`);
			if (res.ok) {
				const data = await res.json();
				if (data.logs) logs = data.logs.split('\n');
			}
		} catch (e) {
			console.error('Log fetch error:', e);
		}
	}

	async function fetchStats() {
		if (!nodeId || !instanceId) return;
		try {
			const res = await apiFetch(`/api/nodes/${nodeId}/instances/${instanceId}/stats`);
			if (res.ok) stats = { ...stats, ...(await res.json()) };
		} catch (e) {
			console.error('Stats fetch error:', e);
		}
	}

	function startPolling() {
		stopPolling();
		logs = [];
		fetchStats();
		fetchInstanceLogs();
		statsInterval = setInterval(fetchStats, 2000);
		logsInterval = setInterval(() => {
			if (activeTab === 'console') fetchInstanceLogs();
		}, 2000);
	}

	function stopPolling() {
		if (statsInterval) clearInterval(statsInterval);
		if (logsInterval) clearInterval(logsInterval);
		statsInterval = null;
		logsInterval = null;
	}

	type BackupAction = 'create' | 'restore' | 'delete';
	function handleBackupAction(action: BackupAction, filename: string | undefined = undefined) {
		if (!nodeId || !instanceId) return;
		confirmTitle =
			action === 'create'
				? 'Create Backup'
				: action === 'restore'
					? 'Restore Backup'
					: 'Delete Backup';
		confirmMessage =
			action === 'create'
				? 'Are you sure you want to create a new backup?'
				: action === 'restore'
					? `Are you sure you want to restore ${filename}? Current data will be overwritten.`
					: `Permanently delete backup ${filename}?`;
		confirmBtnText = action === 'create' ? 'Confirm' : action === 'restore' ? 'Restore' : 'Delete';
		isCriticalAction = action !== 'create';

		pendingBackupAction = async () => {
			let url = `/api/nodes/${nodeId}/instances/${instanceId}/backup`;
			if (action === 'restore') url = `/api/nodes/${nodeId}/instances/${instanceId}/restore`;
			else if (action === 'delete')
				url = `/api/nodes/${nodeId}/instances/${instanceId}/backup/delete`;

			const res = await apiFetch(url, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: filename ? JSON.stringify({ filename }) : null
			});
			if (res.ok) {
				if (action !== 'restore') await fetchBackups();
				else alert('Backup restored successfully.');
			} else {
				const err = await res.json();
				alert(`Action failed: ${err.error || 'Unknown error'}`);
			}
		};
		isConfirmOpen = true;
	}

	function triggerAction(action: string) {
		confirmTitle = `${action.charAt(0).toUpperCase() + action.slice(1)} Instance`;
		confirmMessage = `Are you sure you want to ${action} instance ${instanceId}?`;
		confirmBtnText = action === 'delete' ? 'Delete' : 'Confirm';
		isCriticalAction = action === 'delete' || action === 'stop';
		pendingBackupAction = async () => {
			await apiFetch(`/api/nodes/${nodeId}/instances/${instanceId}/${action}`, { method: 'POST' });
		};
		isConfirmOpen = true;
	}

	function close() {
		stopPolling();
		onClose();
	}

	$effect(() => {
		if (isOpen && nodeId !== null && instanceId) {
			startPolling();
		} else {
			stopPolling();
			if (!isOpen) activeTab = 'console';
		}
	});

	$effect(() => {
		isDeploying = stats.status === 'Provisioning';
	});

	let deployTimer: ReturnType<typeof setInterval> | undefined;
	$effect(() => {
		if (isDeploying) {
			deployTimer = setInterval(() => {
				if (deploymentStep < deploymentSteps.length - 1) deploymentStep++;
			}, 2000);
		} else {
			clearInterval(deployTimer);
			deploymentStep = 0;
		}
		return () => clearInterval(deployTimer);
	});

	$effect(() => {
		if (isOpen && activeTab === 'backups' && nodeId && instanceId) fetchBackups();
	});

	$effect(() => {
		if (isOpen && activeTab === 'history' && nodeId && instanceId) fetchHistoryLogs();
	});
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop -->
		<div
			class="absolute inset-0"
			onclick={close}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Escape' && close()}
			aria-label="Close"
		></div>

		<!-- Modal Window -->
		<div
			class="relative w-full max-w-7xl h-full sm:h-[90vh] flex flex-col md:flex-row bg-slate-900 border border-white/10 rounded-3xl shadow-2xl overflow-hidden font-sans"
			transition:scale={{ start: 0.98, duration: 200, easing: cubicOut }}
		>
			<!-- Sidebar -->
			<div
				class="w-full md:w-72 bg-black/20 border-b md:border-b-0 md:border-r border-white/5 flex flex-col shrink-0"
			>
				<div
					class="p-6 border-b border-white/5 bg-black/20 flex justify-between items-center md:block"
				>
					<h3 class="text-xs font-bold text-white truncate tracking-widest uppercase italic font-heading">
						{instanceId}
					</h3>
					<div class="md:mt-4">
						<div
							class="inline-flex items-center gap-2 px-3 py-1 bg-black/40 border border-white/5 rounded-lg text-[10px] font-bold uppercase tracking-wide {stats.status ===
							'Running'
								? 'text-emerald-400'
								: 'text-rose-400'}"
						>
							<div
								class="w-1.5 h-1.5 rounded-full {stats.status === 'Running'
									? 'bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]'
									: 'bg-rose-500'}"
							></div>
							{stats.status}
						</div>
					</div>
				</div>

				<div class="flex-1 overflow-y-auto p-6 space-y-10 hidden md:block">
					<div>
						<div class="text-[10px] font-bold uppercase tracking-widest mb-2 text-slate-500">
							Instance Uptime
						</div>
						<div class="text-2xl font-bold text-sky-400 tabular-nums">
							{formatUptime((stats.uptime || 0) * 1000)}
						</div>
					</div>

					<div class="space-y-8">
						<div class="space-y-2">
							<div class="flex justify-between text-[10px] font-bold uppercase tracking-widest">
								<span class="text-slate-500">CPU Usage</span>
								<span class="text-white">{stats.cpu_percent?.toFixed(1)}%</span>
							</div>
							<div
								class="h-1.5 bg-black rounded-full overflow-hidden border border-white/5 shadow-inner"
							>
								<div
									class="h-full bg-sky-500 transition-all duration-500"
									style="width: {stats.cpu_percent}%"
								></div>
							</div>
						</div>
						<div class="space-y-2">
							<div class="flex justify-between text-[10px] font-bold uppercase tracking-widest">
								<span class="text-slate-500">Memory Usage</span>
								<span class="text-white">{formatBytes(stats.memory_usage)}</span>
							</div>
							<div
								class="h-1.5 bg-black rounded-full overflow-hidden border border-white/5 shadow-inner"
							>
								<div
									class="h-full bg-indigo-500/60 transition-all duration-500"
									style="width: {memoryPercent}%"
								></div>
							</div>
						</div>
					</div>

					{#if isDeploying}
						<div class="pt-8 border-t border-white/5" transition:slide>
							<div
								class="text-[10px] font-bold text-sky-400 uppercase tracking-widest mb-6 flex items-center gap-3"
							>
								<div class="w-2 h-2 rounded-full bg-sky-500 animate-pulse"></div>
								Starting...
							</div>
							<div class="space-y-4">
								{#each deploymentSteps as step, i}
									<div
										class="text-[11px] font-medium flex items-center gap-4 transition-colors {i <=
										deploymentStep
											? 'text-slate-300'
											: 'text-slate-600'}"
									>
										<div
											class="w-1.5 h-1.5 rounded-full {i < deploymentStep
												? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]'
												: i === deploymentStep
													? 'bg-sky-500 animate-pulse'
													: 'bg-slate-800'}"
										></div>
										<span class={i < deploymentStep ? 'opacity-40' : ''}>{step}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<div class="p-4 border-t border-white/5 bg-black/40 shrink-0">
					<div class="grid grid-cols-2 gap-2">
						<div class="col-span-2">
							<Button
								onclick={() => triggerAction('start')}
								disabled={stats.status === 'Running' || isDeploying}
								variant="primary"
								size="md"
								block={true}
								class="!rounded-xl shadow-lg shadow-sky-500/10">Start Instance</Button
							>
						</div>
						<Button
							onclick={() => triggerAction('restart')}
							disabled={stats.status !== 'Running'}
							variant="secondary"
							size="md"
							class="!rounded-xl">Restart</Button
						>
						<Button
							onclick={() => triggerAction('stop')}
							disabled={stats.status !== 'Running'}
							variant="danger"
							size="md"
							class="!rounded-xl">Stop</Button
						>
					</div>
				</div>
			</div>

			<!-- Main Area -->
			<div class="flex-1 flex flex-col min-w-0 bg-transparent overflow-hidden">
				<div
					class="flex border-b border-white/5 bg-black/40 overflow-x-auto no-scrollbar shrink-0"
				>
					{#each tabs as tab}
						<button
							onclick={() => (activeTab = tab.id)}
							class="px-8 py-5 text-[10px] font-bold uppercase tracking-widest transition-all border-b-2 whitespace-nowrap {activeTab ===
							tab.id
								? 'text-sky-400 border-sky-500 bg-sky-500/5'
								: 'text-slate-500 border-transparent hover:text-slate-300'}"
						>
							{tab.label}
						</button>
					{/each}
				</div>

				<div class="flex-1 relative overflow-hidden flex flex-col bg-black/20">
					{#if activeTab === 'console'}
						<div class="flex-1 p-6 min-h-0" in:fade={{ duration: 150 }}>
							<Console {logs} title={`instance@${instanceId}:~`} />
						</div>
					{:else if activeTab === 'metrics'}
						<div class="flex-1 p-10 overflow-y-auto custom-scrollbar" in:fade={{ duration: 150 }}>
							{#if nodeId !== null && instanceId}
								<ResourceMetricsPanel {nodeId} {instanceId} {memTotal} height={450} />
							{/if}
						</div>
					{:else if activeTab === 'node_logs'}
						<div class="flex-1 min-h-0" in:fade={{ duration: 150 }}>
							{#if nodeId !== null}
								<LogViewer {nodeId} {isOpen} embedded={true} />
							{/if}
						</div>
					{:else if activeTab === 'backups'}
						<div class="flex-1 p-10 overflow-y-auto custom-scrollbar" in:fade={{ duration: 150 }}>
							<div
								class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-10 border-b border-white/5 pb-6"
							>
								<div>
									<h4 class="text-sm font-bold text-white uppercase tracking-widest font-heading italic">
										Instance Backups
									</h4>
									<p class="text-xs text-slate-500 mt-2 font-medium">
										Manage and restore instance snapshots
									</p>
								</div>
								<Button
									onclick={() => handleBackupAction('create')}
									variant="primary"
									size="md"
									icon="ph:plus-bold"
									class="!rounded-xl shadow-lg shadow-sky-500/10">Create Backup</Button
								>
							</div>

							{#if isLoadingData}
								<div class="flex justify-center py-20">
									<div
										class="w-8 h-8 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin shadow-[0_0_15px_rgba(14,165,233,0.3)]"
									></div>
								</div>
							{:else if backups.length === 0}
								<div
									class="text-center py-20 text-[11px] font-bold text-slate-600 uppercase tracking-widest border-2 border-dashed border-white/5 rounded-2xl bg-black/20"
								>
									No backups found for this instance
								</div>
							{:else}
								<div class="space-y-3">
									{#each backups as backup}
										<div
											class="flex flex-col sm:flex-row items-start sm:items-center justify-between p-6 bg-black/40 border border-white/5 rounded-2xl hover:border-sky-500/30 transition-all group gap-4"
										>
											<div class="flex-1 min-w-0">
												<div class="text-xs font-mono font-bold text-slate-200 truncate">
													{backup.filename}
												</div>
												<div
													class="text-[10px] mt-2 font-medium uppercase tracking-wider flex items-center gap-4 text-slate-500"
												>
													<span>{new Date(backup.date).toLocaleString()}</span>
													<span class="w-1 h-1 bg-slate-800 rounded-full"></span>
													<span class="text-sky-500/70">{formatBytes(backup.size)}</span>
													{#if getBackupVersion(backup.filename)}
														<span
															class="px-2 py-0.5 bg-sky-500/10 text-sky-400 border border-sky-500/20 rounded"
															>v{getBackupVersion(backup.filename)}</span
														>
													{/if}
												</div>
											</div>
											<div class="flex gap-3 w-full sm:w-auto">
												<Button
													onclick={() => handleBackupAction('restore', backup.filename)}
													variant="secondary"
													size="sm"
													class="!rounded-lg flex-1 sm:flex-initial">Restore</Button
												>
												<button
													onclick={() => handleBackupAction('delete', backup.filename)}
													class="p-2 text-slate-600 hover:text-rose-400 transition-colors"
													aria-label="Delete"
												>
													<Icon name="ph:trash-bold" size="1.25rem" />
												</button>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					{:else if activeTab === 'history'}
						<div
							class="flex-1 min-h-0 overflow-y-auto custom-scrollbar"
							in:fade={{ duration: 150 }}
						>
							<table class="w-full text-left border-collapse">
								<thead class="sticky top-0 bg-black border-b border-white/5 z-10">
									<tr>
										<th
											class="px-8 py-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest"
											>Activity</th
										>
										<th
											class="px-8 py-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest text-center"
											>Status</th
										>
										<th
											class="px-8 py-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest text-right"
											>Timestamp</th
										>
									</tr>
								</thead>
								<tbody class="divide-y divide-white/5">
									{#each historyLogs as log}
										<tr class="hover:bg-white/[0.02] transition-colors group text-sm font-medium">
											<td class="px-8 py-5 text-slate-300 uppercase tracking-tight text-[11px]"
												>{log.action}</td
											>
											<td class="px-8 py-5 text-center">
												<span
													class="px-2.5 py-1 text-[9px] font-bold uppercase rounded-md {log.status ===
													'success'
														? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
														: 'bg-rose-500/10 text-rose-400 border border-rose-500/20'}"
												>
													{log.status}
												</span>
											</td>
											<td class="px-8 py-5 text-slate-500 text-right font-mono text-[10px]">
												{new Date(log.timestamp).toLocaleTimeString()}
											</td>
										</tr>
									{:else}
										<tr>
											<td
												colspan="3"
												class="px-8 py-24 text-center text-[11px] font-bold text-slate-600 uppercase tracking-widest"
											>
												No activity logs found for this instance
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>

			<button
				onclick={close}
				class="absolute top-4 right-4 sm:top-6 sm:right-6 p-2 text-slate-500 hover:text-white hover:bg-white/5 rounded-xl transition-all z-50"
				aria-label="Close"
			>
				<Icon name="ph:x-bold" size="1.5rem" />
			</button>
		</div>
	</div>

	<ConfirmDialog
		bind:isOpen={isConfirmOpen}
		title={confirmTitle}
		message={confirmMessage}
		confirmText={confirmBtnText}
		isCritical={isCriticalAction}
		onConfirm={pendingBackupAction}
	/>
{/if}

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
		height: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
