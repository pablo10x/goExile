<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, scale, fly, slide } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { serverVersions } from '$lib/stores';
	import Terminal from './Terminal.svelte';
	import ResourceHistoryChart from './ResourceHistoryChart.svelte';
	import ResourceMetricsPanel from './ResourceMetricsPanel.svelte';
	import ConfirmDialog from './ConfirmDialog.svelte';
	import LogViewer from './LogViewer.svelte';

	export let isOpen: boolean = false;
	export let spawnerId: number | null = null;
	export let instanceId: string | null = null;
	export let onClose: () => void;
	export let memTotal: number = 0;

	let logs: string[] = [];

	let stats = {
		cpu_percent: 0,
		memory_usage: 0,
		disk_usage: 0,
		status: 'Unknown',
		uptime: 0
	};
	let activeTab: 'console' | 'metrics' | 'backups' | 'history' | 'node_logs' = 'console';

	interface TabItem {
		id: 'console' | 'metrics' | 'backups' | 'history' | 'node_logs';
		label: string;
		icon: string;
	}

	// New State for Backups and History
	let backups: any[] = [];
	let historyLogs: any[] = [];
	let isLoadingData = false;

	// Confirm Dialog State
	let isConfirmOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let confirmBtnText = 'Confirm';
	let isCriticalAction = false;
	let pendingBackupAction: () => Promise<void> = async () => {};

	let statsInterval: ReturnType<typeof setInterval> | null = null;
	let logsInterval: ReturnType<typeof setInterval> | null = null;

	// Provisioning State
	let isProvisioning = false;
	let provisioningStep = 0;
	const provisioningSteps = [
		'Allocating resources...',
		'Downloading game files...',
		'Configuring environment...',
		'Starting process...'
	];

	$: activeVersion = $serverVersions.find((v) => v.is_active);

	function getBackupVersion(filename: string): string | null {
		const match = filename.match(/_v(.*?)\.zip$/);
		return match ? match[1] : null;
	}

	function isOutdated(filename: string): boolean {
		if (!activeVersion) return false;
		const version = getBackupVersion(filename);
		if (!version) return false;
		return version !== activeVersion.version;
	}

	$: if (isOpen && spawnerId !== null && instanceId) {
		startPolling();
		provisioningStep = 0;
	} else {
		stopPolling();
		activeTab = 'console';
	}

	$: if (stats.status === 'Provisioning') {
		isProvisioning = true;
	} else {
		isProvisioning = false;
	}

	let provTimer: any;
	$: if (isProvisioning) {
		clearInterval(provTimer);
		provTimer = setInterval(() => {
			if (provisioningStep < provisioningSteps.length - 1) provisioningStep++;
		}, 2000);
	} else {
		clearInterval(provTimer);
	}

	$: if (isOpen && activeTab === 'backups' && spawnerId && instanceId) {
		fetchBackups();
	}

	$: if (isOpen && activeTab === 'history' && spawnerId && instanceId) {
		fetchHistoryLogs();
	}

	async function fetchBackups() {
		if (!spawnerId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/backups`);
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
		if (!spawnerId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/history`);
			if (res.ok) {
				historyLogs = await res.json();
			} else {
				console.error('Failed to fetch history:', res.status, res.statusText);
				historyLogs = [];
			}
		} catch (e) {
			console.error('Error fetching history:', e);
			historyLogs = [];
		} finally {
			isLoadingData = false;
		}
	}

	function handleBackupAction(action: 'create' | 'restore' | 'delete', filename?: string) {
		if (!spawnerId || !instanceId) return;

		if (action === 'create') {
			confirmTitle = 'Create Backup';
			confirmMessage = 'Are you sure you want to create a new backup? The server must be stopped.';
			confirmBtnText = 'Start Backup';
			isCriticalAction = false;
		} else if (action === 'restore') {
			confirmTitle = 'Restore Backup';
			confirmMessage = `Are you sure you want to restore "${filename}"?\n\n⚠️ WARNING: This will overwrite all current game files!`;
			confirmBtnText = 'Restore Files';
			isCriticalAction = true;
		} else if (action === 'delete') {
			confirmTitle = 'Delete Backup';
			confirmMessage = `Are you sure you want to PERMANENTLY delete "${filename}"?`;
			confirmBtnText = 'Delete Backup';
			isCriticalAction = true;
		}

		pendingBackupAction = async () => {
			let url = '';
			let body = null;
			if (action === 'create') url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup`;
			else if (action === 'restore') {
				url = `/api/spawners/${spawnerId}/instances/${instanceId}/restore`;
				body = JSON.stringify({ filename });
			} else if (action === 'delete') {
				url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup/delete`;
				body = JSON.stringify({ filename });
			}

			const res = await fetch(url, {
				method: 'POST',
				headers: body ? { 'Content-Type': 'application/json' } : undefined,
				body
			});

			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Action failed');
			}

			if (action !== 'restore') await fetchBackups();
			else alert('Backup restored successfully.');
		};

		isConfirmOpen = true;
	}

	async function fetchInstanceLogs() {
		if (!spawnerId || !instanceId) return;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`);
			if (res.ok) {
				const data = await res.json();
				if (data.logs) {
					logs = data.logs.split('\n');
				}
			}
		} catch (e) {
			console.error('Failed to fetch logs:', e);
		}
	}

	function startPolling() {
		stopPolling();
		logs = [];
		fetchStats();
		fetchInstanceLogs();

		statsInterval = setInterval(fetchStats, 2000);
		logsInterval = setInterval(() => {
			if (activeTab === 'console') {
				fetchInstanceLogs();
			}
		}, 2000);
	}

	function stopPolling() {
		if (statsInterval) {
			clearInterval(statsInterval);
			statsInterval = null;
		}
		if (logsInterval) {
			clearInterval(logsInterval);
			logsInterval = null;
		}
		logs = [];
	}

	async function fetchStats() {
		if (spawnerId === null || !instanceId) return;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
			if (res.ok) {
				const data = await res.json();
				stats = { ...stats, ...data };
			}
		} catch (e) {
			console.error(e);
		}
	}

	function triggerAction(action: string) {
		confirmTitle = `${action.charAt(0).toUpperCase() + action.slice(1)} Instance`;
		confirmMessage = `Are you sure you want to ${action.toUpperCase()} this instance?`;
		confirmBtnText = action === 'delete' ? 'Delete' : 'Confirm';
		isCriticalAction = action === 'delete' || action === 'stop';

		pendingBackupAction = async () => {
			await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/${action}`, {
				method: 'POST'
			});
		};
		isConfirmOpen = true;
	}

	function close() {
		stopPolling();
		onClose();
	}

	// Calculate memory percentage for display
	$: memoryPercent = memTotal
		? Math.min(100, (stats.memory_usage / (memTotal * 1024 * 1024)) * 100)
		: 10;
	$: diskPercent = 15; // Placeholder
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
		transition:fade={{ duration: 250, easing: cubicOut }}
	>
		<!-- Enhanced Backdrop with animated gradient -->
		<div
			class="absolute inset-0 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 backdrop-blur-xl"
			onclick={close}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && close()}
			role="button"
			tabindex="0"
			aria-label="Close console"
			style="background-image: radial-gradient(circle at 20% 50%, rgba(59, 130, 246, 0.05) 0%, transparent 50%), radial-gradient(circle at 80% 80%, rgba(139, 92, 246, 0.05) 0%, transparent 50%);"
		></div>

		<!-- Enhanced Modal Window with glassmorphism -->
		<div
			class="relative w-full max-w-7xl h-[90vh] flex bg-slate-900/40 backdrop-blur-2xl border border-slate-300/50 dark:border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden"
			transition:scale={{ start: 0.9, duration: 300, easing: elasticOut }}
			style="box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.05) inset;"
		>
			<!-- Enhanced Sidebar -->
			<div
				class="w-80 bg-gradient-to-b from-slate-950/90 to-slate-900/90 backdrop-blur-xl border-r border-slate-300/50 dark:border-slate-700/50 flex flex-col shrink-0 relative overflow-hidden"
			>
				<!-- Ambient glow effect -->
				<div
					class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-transparent to-purple-500/5 pointer-events-none"
				></div>

				<!-- Header -->
				<div
					class="relative p-6 border-b border-slate-300/50 dark:border-slate-700/50 bg-slate-900/50"
				>
					<div class="flex items-start justify-between mb-3">
						<h3
							class="text-xl font-bold text-slate-100 break-all leading-tight bg-gradient-to-r from-slate-100 to-slate-300 bg-clip-text text-transparent"
						>
							{instanceId}
						</h3>
						<div class="flex items-center gap-2 ml-3">
							<span
								class={`relative flex h-3 w-3 ${stats.status === 'Running' ? 'animate-pulse' : ''}`}
							>
								{#if stats.status === 'Running'}
									<span
										class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
									></span>
									<span
										class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500 shadow-lg shadow-emerald-500/50"
									></span>
								{:else if stats.status === 'Provisioning'}
									<span
										class="relative inline-flex rounded-full h-3 w-3 bg-blue-500 shadow-lg shadow-blue-500/50 animate-bounce"
									></span>
								{:else}
									<span
										class="relative inline-flex rounded-full h-3 w-3 bg-red-500 shadow-lg shadow-red-500/50"
									></span>
								{/if}
							</span>
						</div>
					</div>
					<div
						class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-800/50 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50"
					>
						<span class="text-xs font-semibold font-mono text-slate-700 dark:text-slate-300"
							>{stats.status || 'Unknown'}</span
						>
					</div>
				</div>

				<div class="relative p-6 space-y-6 flex-1 overflow-y-auto custom-scrollbar">
					<!-- Uptime Card -->
					<div class="relative group">
						<div
							class="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-xl blur-xl group-hover:blur-2xl transition-all"
						></div>
						<div
							class="relative bg-slate-900/50 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-4 hover:border-slate-600/50 transition-all"
						>
							<div
								class="text-[10px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2"
							>
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								Uptime
							</div>
							<div
								class="text-4xl font-bold font-mono bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent tracking-tight"
							>
								{formatUptime((stats.uptime || 0) * 1000)}
							</div>
						</div>
					</div>

					<!-- Enhanced Resource Metrics -->
					<div class="space-y-4">
						<!-- CPU -->
						<div class="relative group">
							<div
								class="absolute inset-0 bg-gradient-to-r from-orange-500/10 to-red-500/10 rounded-xl blur-lg group-hover:blur-xl transition-all"
							></div>
							<div
								class="relative bg-slate-900/50 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-4 hover:border-orange-500/30 transition-all"
							>
								<div class="flex items-center justify-between mb-3">
									<div class="flex items-center gap-2">
										<div class="p-2 rounded-lg bg-orange-500/10 text-orange-400">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M13 10V3L4 14h7v7l9-11h-7z"
												/>
											</svg>
										</div>
										<span
											class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider"
											>CPU Usage</span
										>
									</div>
									<span class="text-lg font-bold font-mono text-orange-400"
										>{stats.cpu_percent?.toFixed(1)}%</span
									>
								</div>
								<div
									class="relative w-full h-2 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-orange-500 to-red-500 transition-all duration-700 ease-out rounded-full shadow-lg shadow-orange-500/50"
										style="width: {stats.cpu_percent}%; box-shadow: 0 0 20px rgba(249, 115, 22, 0.4);"
									></div>
								</div>
							</div>
						</div>

						<!-- Memory -->
						<div class="relative group">
							<div
								class="absolute inset-0 bg-gradient-to-r from-purple-500/10 to-pink-500/10 rounded-xl blur-lg group-hover:blur-xl transition-all"
							></div>
							<div
								class="relative bg-slate-900/50 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-4 hover:border-purple-500/30 transition-all"
							>
								<div class="flex items-center justify-between mb-3">
									<div class="flex items-center gap-2">
										<div class="p-2 rounded-lg bg-purple-500/10 text-purple-400">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
												/>
											</svg>
										</div>
										<span
											class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider"
											>Memory</span
										>
									</div>
									<span class="text-lg font-bold font-mono text-purple-400"
										>{formatBytes(stats.memory_usage)}</span
									>
								</div>
								<div
									class="relative w-full h-2 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-purple-500 to-pink-500 transition-all duration-700 ease-out rounded-full shadow-lg shadow-purple-500/50"
										style="width: {memoryPercent}%; box-shadow: 0 0 20px rgba(168, 85, 247, 0.4);"
									></div>
								</div>
							</div>
						</div>

						<!-- Disk -->
						<div class="relative group">
							<div
								class="absolute inset-0 bg-gradient-to-r from-emerald-500/10 to-teal-500/10 rounded-xl blur-lg group-hover:blur-xl transition-all"
							></div>
							<div
								class="relative bg-slate-900/50 backdrop-blur-sm border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-4 hover:border-emerald-500/30 transition-all"
							>
								<div class="flex items-center justify-between mb-3">
									<div class="flex items-center gap-2">
										<div class="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"
												/>
											</svg>
										</div>
										<span
											class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider"
											>Disk Space</span
										>
									</div>
									<span class="text-lg font-bold font-mono text-emerald-400"
										>{formatBytes(stats.disk_usage)}</span
									>
								</div>
								<div
									class="relative w-full h-2 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-emerald-500 to-teal-500 transition-all duration-700 ease-out rounded-full shadow-lg shadow-emerald-500/50"
										style="width: {diskPercent}%; box-shadow: 0 0 20px rgba(16, 185, 129, 0.4);"
									></div>
								</div>
							</div>
						</div>
					</div>

					<!-- Enhanced Provisioning Steps -->
					{#if isProvisioning}
						<div
							class="relative pt-6 border-t border-slate-300/50 dark:border-slate-700/50"
							in:fly={{ y: 20, duration: 500, easing: cubicOut }}
						>
							<div
								class="absolute inset-0 bg-gradient-to-r from-blue-500/5 to-cyan-500/5 rounded-xl blur-xl"
							></div>
							<div class="relative">
								<div
									class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-4 flex items-center gap-2"
								>
									<div class="relative">
										<div
											class="w-4 h-4 rounded-full border-2 border-blue-400 border-t-transparent animate-spin"
										></div>
										<div class="absolute inset-0 w-4 h-4 rounded-full bg-blue-400/20 blur-sm"></div>
									</div>
									Provisioning Instance
								</div>
								<div class="space-y-3">
									{#each provisioningSteps as step, i}
										<div
											class="flex items-center gap-3 text-xs transition-all duration-300"
											in:fly={{ x: -20, duration: 300, delay: i * 100 }}
										>
											{#if i < provisioningStep}
												<div class="relative flex-shrink-0">
													<div
														class="w-6 h-6 rounded-full bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center text-slate-900 dark:text-white shadow-lg shadow-emerald-500/50"
													>
														<svg
															class="w-3 h-3"
															fill="none"
															stroke="currentColor"
															viewBox="0 0 24 24"
														>
															<path
																stroke-linecap="round"
																stroke-linejoin="round"
																stroke-width="3"
																d="M5 13l4 4L19 7"
															/>
														</svg>
													</div>
												</div>
												<span class="text-slate-500 line-through decoration-slate-700 font-medium"
													>{step}</span
												>
											{:else if i === provisioningStep}
												<div class="relative flex-shrink-0">
													<div
														class="w-6 h-6 rounded-full border-2 border-blue-500 border-t-transparent animate-spin"
													></div>
													<div
														class="absolute inset-0 w-6 h-6 rounded-full bg-blue-500/20 blur-sm animate-pulse"
													></div>
												</div>
												<span class="text-blue-200 font-semibold animate-pulse">{step}</span>
											{:else}
												<div
													class="w-6 h-6 rounded-full border-2 border-slate-200 dark:border-slate-800 flex-shrink-0"
												></div>
												<span class="text-slate-600 font-medium">{step}</span>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Enhanced Action Buttons -->
				<div
					class="relative p-4 border-t border-slate-300/50 dark:border-slate-700/50 bg-slate-900/70 backdrop-blur-sm"
				>
					<div class="grid grid-cols-2 gap-2">
						<button
							onclick={() => triggerAction('start')}
							disabled={stats.status === 'Running' || stats.status === 'Provisioning'}
							class="col-span-2 relative group px-4 py-3 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-slate-900 dark:text-white rounded-xl text-sm font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed disabled:from-slate-800 disabled:to-slate-800 shadow-lg hover:shadow-emerald-500/50 disabled:shadow-none overflow-hidden"
						>
							<span class="relative z-10 flex items-center justify-center gap-2">
								<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
									<path d="M8 5v14l11-7z" />
								</svg>
								Start Instance
							</span>
							<div
								class="absolute inset-0 bg-gradient-to-r from-emerald-400 to-teal-400 opacity-0 group-hover:opacity-20 transition-opacity"
							></div>
						</button>
						<button
							onclick={() => triggerAction('restart')}
							disabled={stats.status !== 'Running'}
							class="relative group px-4 py-2.5 bg-slate-800/80 hover:bg-gradient-to-r hover:from-blue-600/20 hover:to-cyan-600/20 text-slate-700 dark:text-slate-300 hover:text-blue-400 rounded-xl text-sm font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-300/50 dark:border-slate-700/50 hover:border-blue-500/50 overflow-hidden"
						>
							<span class="relative z-10 flex items-center justify-center gap-2">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
									/>
								</svg>
								Restart
							</span>
						</button>
						<button
							onclick={() => triggerAction('stop')}
							disabled={stats.status !== 'Running'}
							class="relative group px-4 py-2.5 bg-slate-800/80 hover:bg-gradient-to-r hover:from-red-600/20 hover:to-rose-600/20 text-slate-700 dark:text-slate-300 hover:text-red-400 rounded-xl text-sm font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-300/50 dark:border-slate-700/50 hover:border-red-500/50 overflow-hidden"
						>
							<span class="relative z-10 flex items-center justify-center gap-2">
								<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
									<path d="M6 6h12v12H6z" />
								</svg>
								Stop
							</span>
						</button>
					</div>
				</div>
			</div>

			<!-- Enhanced Main Area -->
			<div class="flex-1 flex flex-col min-w-0 relative overflow-hidden">
				<!-- Ambient background effects -->
				<div
					class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-transparent to-purple-500/5 pointer-events-none"
				></div>

				<!-- Enhanced Tabs -->
				<div
					class="relative flex border-b border-slate-300/50 dark:border-slate-700/50 bg-slate-900/40 backdrop-blur-xl shrink-0 z-10"
				>
					{#each [{ id: 'console', label: 'Console', icon: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' }, { id: 'metrics', label: 'Metrics', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' }, { id: 'backups', label: 'Backups', icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4' }, { id: 'history', label: 'History', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' }, { id: 'node_logs', label: 'Node Logs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' }] as TabItem[] as tab: TabItem}
						<button
							class="relative px-6 py-4 text-xs font-bold uppercase tracking-wider transition-all group {activeTab ===
							tab.id
								? 'text-blue-400'
								: 'text-slate-500 hover:text-slate-700 dark:text-slate-300'}"
							onclick={() => (activeTab = tab.id)}
						>
							<span class="relative z-10 flex items-center gap-2">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d={tab.icon}
									/>
								</svg>
								{tab.label}
							</span>
							{#if activeTab === tab.id}
								<div
									class="absolute bottom-0 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500 to-cyan-500 shadow-lg shadow-blue-500/50"
									transition:slide={{ duration: 200 }}
								></div>
								<div
									class="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-cyan-500/10 rounded-t-lg"
								></div>
							{/if}
						</button>
					{/each}
				</div>

				<div class="relative flex-1 overflow-hidden">
					{#if activeTab === 'console'}
						<div class="absolute inset-0 p-0">
							<!-- Enhanced Live Indicator -->
							<div class="absolute top-4 right-4 z-10">
								<div class="relative">
									<div
										class="absolute inset-0 bg-emerald-500/20 blur-xl rounded-full animate-pulse"
									></div>
									<div
										class="relative px-4 py-2 rounded-full bg-slate-900/80 backdrop-blur-xl border border-emerald-500/30 text-[11px] font-bold font-mono text-emerald-400 flex items-center gap-2 shadow-lg"
									>
										<span class="relative flex h-2 w-2">
											<span
												class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
											></span>
											<span
												class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500 shadow-lg shadow-emerald-500/50"
											></span>
										</span>
										LIVE
									</div>
								</div>
							</div>
							<Terminal {logs} title={`root@${instanceId}:~`} />
						</div>
					{:else if activeTab === 'metrics'}
						<div
							class="p-8 h-full overflow-y-auto custom-scrollbar bg-gradient-to-br from-slate-950/50 to-slate-900/50"
						>
							{#if spawnerId !== null && instanceId}
								{#key instanceId}
									<ResourceMetricsPanel {spawnerId} {instanceId} {memTotal} height={300} />
								{/key}
							{:else}
								<div class="flex items-center justify-center h-full">
									<div class="text-center">
										<div
											class="w-16 h-16 mx-auto mb-4 rounded-full bg-slate-800/50 flex items-center justify-center"
										>
											<svg
												class="w-8 h-8 text-slate-600"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
												/>
											</svg>
										</div>
										<p class="text-slate-500 font-medium">No instance selected</p>
									</div>
								</div>
							{/if}
						</div>
					{:else if activeTab === 'node_logs'}
						<div class="h-full relative">
							{#if spawnerId !== null}
								<LogViewer {spawnerId} isOpen={true} embedded={true} />
							{/if}
						</div>
					{:else if activeTab === 'backups'}
						<div
							class="p-8 h-full overflow-y-auto custom-scrollbar bg-gradient-to-br from-slate-950/50 to-slate-900/50"
						>
							<div class="flex justify-between items-center mb-8">
								<div>
									<h3
										class="text-2xl font-bold bg-gradient-to-r from-slate-100 to-slate-300 bg-clip-text text-transparent"
									>
										Instance Backups
									</h3>
									<p class="text-sm text-slate-500 mt-1">
										Manage and restore your instance backups
									</p>
								</div>
								<button
									onclick={() => handleBackupAction('create')}
									class="relative group px-6 py-3 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-slate-900 dark:text-white rounded-xl text-sm font-bold transition-all shadow-lg hover:shadow-blue-500/50 overflow-hidden"
								>
									<span class="relative z-10 flex items-center gap-2">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 4v16m8-8H4"
											/>
										</svg>
										Create Backup
									</span>
									<div
										class="absolute inset-0 bg-gradient-to-r from-blue-400 to-cyan-400 opacity-0 group-hover:opacity-20 transition-opacity"
									></div>
								</button>
							</div>

							{#if isLoadingData}
								<div class="flex items-center justify-center py-20">
									<div class="relative">
										<div
											class="w-12 h-12 rounded-full border-4 border-slate-200 dark:border-slate-800 border-t-blue-500 animate-spin"
										></div>
										<div
											class="absolute inset-0 w-12 h-12 rounded-full bg-blue-500/20 blur-lg animate-pulse"
										></div>
									</div>
								</div>
							{:else if backups.length === 0}
								<div class="relative group">
									<div
										class="absolute inset-0 bg-gradient-to-r from-slate-800/20 to-slate-700/20 rounded-2xl blur-xl"
									></div>
									<div
										class="relative text-center py-20 border-2 border-dashed border-slate-300/50 dark:border-slate-700/50 rounded-2xl bg-slate-900/30 backdrop-blur-sm hover:border-slate-600/50 transition-all"
									>
										<div
											class="w-16 h-16 mx-auto mb-4 rounded-full bg-slate-800/50 flex items-center justify-center"
										>
											<svg
												class="w-8 h-8 text-slate-600"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
												/>
											</svg>
										</div>
										<p class="text-slate-500 font-medium text-lg">No backups found</p>
										<p class="text-slate-600 text-sm mt-2">
											Create your first backup to get started
										</p>
									</div>
								</div>
							{:else}
								<div class="space-y-3">
									{#each backups as backup, i (backup.filename)}
										{@const outdated = isOutdated(backup.filename)}
										{@const version = getBackupVersion(backup.filename)}
										<div
											class="relative group"
											in:fly={{ y: 20, duration: 400, delay: 50 * i, easing: cubicOut }}
										>
											<div
												class="absolute inset-0 bg-gradient-to-r {outdated
													? 'from-orange-500/10 to-red-500/10'
													: 'from-emerald-500/10 to-teal-500/10'} rounded-xl blur-lg opacity-0 group-hover:opacity-100 transition-opacity"
											></div>
											<div
												class="relative flex items-center justify-between p-5 rounded-xl border transition-all {outdated
													? 'bg-orange-500/5 border-orange-500/20 hover:border-orange-500/40'
													: 'bg-slate-900/50 border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600/50'} backdrop-blur-sm"
											>
												<div class="flex items-center gap-4 flex-1 min-w-0">
													<div class="relative flex-shrink-0">
														<div
															class="absolute inset-0 {outdated
																? 'bg-orange-500/20'
																: 'bg-emerald-500/20'} blur-lg rounded-xl"
														></div>
														<div
															class="relative p-3 rounded-xl {outdated
																? 'bg-orange-500/10 text-orange-400'
																: 'bg-emerald-500/10 text-emerald-400'} border {outdated
																? 'border-orange-500/30'
																: 'border-emerald-500/30'}"
														>
															<svg
																class="w-6 h-6"
																fill="none"
																stroke="currentColor"
																viewBox="0 0 24 24"
															>
																<path
																	stroke-linecap="round"
																	stroke-linejoin="round"
																	stroke-width="2"
																	d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
																/>
															</svg>
														</div>
													</div>
													<div class="flex-1 min-w-0">
														<div
															class="text-sm font-bold text-slate-800 dark:text-slate-200 font-mono truncate mb-1"
														>
															{backup.filename}
														</div>
														<div
															class="flex items-center gap-3 flex-wrap text-xs text-slate-500 dark:text-slate-400"
														>
															<span class="flex items-center gap-1">
																<svg
																	class="w-3 h-3"
																	fill="none"
																	stroke="currentColor"
																	viewBox="0 0 24 24"
																>
																	<path
																		stroke-linecap="round"
																		stroke-linejoin="round"
																		stroke-width="2"
																		d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
																	/>
																</svg>
																{new Date(backup.date).toLocaleString()}
															</span>
															<span class="flex items-center gap-1">
																<svg
																	class="w-3 h-3"
																	fill="none"
																	stroke="currentColor"
																	viewBox="0 0 24 24"
																>
																	<path
																		stroke-linecap="round"
																		stroke-linejoin="round"
																		stroke-width="2"
																		d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
																	/>
																</svg>
																{formatBytes(backup.size)}
															</span>
															{#if version}
																<span
																	class="px-2 py-1 rounded-full font-mono font-bold text-[10px] {outdated
																		? 'bg-orange-500/20 text-orange-400 border border-orange-500/30'
																		: 'bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-600'}"
																>
																	v{version}
																</span>
															{/if}
														</div>
													</div>
												</div>
												<div class="flex gap-2 ml-4 flex-shrink-0">
													<button
														onclick={() => handleBackupAction('restore', backup.filename)}
														class="relative group/btn px-4 py-2 bg-slate-800/80 hover:bg-gradient-to-r hover:from-blue-600 hover:to-cyan-600 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:text-white rounded-lg text-xs font-semibold transition-all border border-slate-300/50 dark:border-slate-700/50 hover:border-blue-500/50 overflow-hidden"
													>
														<span class="relative z-10 flex items-center gap-1">
															<svg
																class="w-3 h-3"
																fill="none"
																stroke="currentColor"
																viewBox="0 0 24 24"
															>
																<path
																	stroke-linecap="round"
																	stroke-linejoin="round"
																	stroke-width="2"
																	d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
																/>
															</svg>
															Restore
														</span>
													</button>
													<button
														onclick={() => handleBackupAction('delete', backup.filename)}
														class="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors border border-transparent hover:border-red-500/30"
														title="Delete"
													>
														<svg
															class="w-4 h-4"
															fill="none"
															stroke="currentColor"
															viewBox="0 0 24 24"
														>
															<path
																stroke-linecap="round"
																stroke-linejoin="round"
																stroke-width="2"
																d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
															/>
														</svg>
													</button>
												</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					{:else if activeTab === 'history'}
						<div
							class="h-full overflow-y-auto custom-scrollbar bg-gradient-to-br from-slate-950/50 to-slate-900/50"
						>
							<table class="w-full text-left border-collapse">
								<thead
									class="sticky top-0 z-10 bg-slate-900/90 backdrop-blur-xl border-b border-slate-300/50 dark:border-slate-700/50"
								>
									<tr>
										<th
											class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
											>Action</th
										>
										<th
											class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
											>Status</th
										>
										<th
											class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
											>Time</th
										>
										<th
											class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
											>Details</th
										>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-800/30">
									{#each historyLogs || [] as log, i}
										<tr
											class="group hover:bg-slate-800/30 transition-all"
											in:fly={{ x: -20, duration: 300, delay: i * 50 }}
										>
											<td class="px-6 py-4">
												<span
													class="inline-flex items-center font-mono text-sm text-slate-700 dark:text-slate-300 bg-slate-800/50 px-3 py-1.5 rounded-lg border border-slate-300/50 dark:border-slate-700/50 group-hover:border-slate-600/50 transition-colors"
												>
													{log.action}
												</span>
											</td>
											<td class="px-6 py-4">
												<span
													class="inline-flex items-center px-3 py-1 rounded-lg text-xs font-bold border {log.status ===
													'success'
														? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
														: 'bg-red-500/10 text-red-400 border-red-500/20'}"
												>
													{#if log.status === 'success'}
														<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 24 24">
															<path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
														</svg>
													{:else}
														<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 24 24">
															<path
																d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
															/>
														</svg>
													{/if}
													{log.status}
												</span>
											</td>
											<td class="px-6 py-4 text-sm text-slate-500 dark:text-slate-400 font-mono">
												{new Date(log.timestamp).toLocaleString()}
											</td>
											<td
												class="px-6 py-4 text-sm text-slate-500 truncate max-w-xs"
												title={log.details}
											>
												{log.details || '-'}
											</td>
										</tr>
									{/each}
									{#if (historyLogs || []).length === 0 && !isLoadingData}
										<tr>
											<td colspan="4" class="px-6 py-20 text-center">
												<div class="flex flex-col items-center">
													<div
														class="w-16 h-16 mb-4 rounded-full bg-slate-800/50 flex items-center justify-center"
													>
														<svg
															class="w-8 h-8 text-slate-600"
															fill="none"
															stroke="currentColor"
															viewBox="0 0 24 24"
														>
															<path
																stroke-linecap="round"
																stroke-linejoin="round"
																stroke-width="2"
																d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
															/>
														</svg>
													</div>
													<p class="text-slate-500 font-medium text-lg">No history found</p>
													<p class="text-slate-600 text-sm mt-2">
														Instance actions will appear here
													</p>
												</div>
											</td>
										</tr>
									{/if}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>

			<!-- Enhanced Close Button -->
			<button
				onclick={close}
				class="absolute top-4 right-4 p-2.5 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800/80 backdrop-blur-xl rounded-xl transition-all z-30 border border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600 group"
				aria-label="Close"
			>
				<svg
					class="w-5 h-5 group-hover:rotate-90 transition-transform duration-300"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
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
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: rgba(15, 23, 42, 0.3);
		border-radius: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(71, 85, 105, 0.5);
		border-radius: 4px;
		transition: background 0.2s;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(71, 85, 105, 0.7);
	}

	@keyframes shimmer {
		0% {
			background-position: -200% center;
		}
		100% {
			background-position: 200% center;
		}
	}
</style>
