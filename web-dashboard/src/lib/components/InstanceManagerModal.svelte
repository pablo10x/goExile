<script lang="ts">
	import { fade, scale, fly, slide } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { serverVersions } from '$lib/stores';
	import Terminal from './Terminal.svelte';
	import ResourceMetricsPanel from './ResourceMetricsPanel.svelte';
	import ConfirmDialog from './ConfirmDialog.svelte';
	import LogViewer from './LogViewer.svelte';

	interface Props {
		isOpen: boolean;
		spawnerId: number | null;
		instanceId: string | null;
		onClose: () => void;
		memTotal?: number;
	}

	let {
		isOpen = $bindable(false),
		spawnerId = null,
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
		{ id: 'console', label: 'Terminal' },
		{ id: 'metrics', label: 'Metrics' },
		{ id: 'backups', label: 'Archives' },
		{ id: 'history', label: 'Protocol Logs' },
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

	// Provisioning Logic
	let isProvisioning = $state(false);
	let provisioningStep = $state(0);
	const provisioningSteps = [
		'Allocating resources...', 
		'Downloading game files...', 
		'Configuring environment...', 
		'Starting process...'
	];

	// Derived State
	let activeVersion = $derived($serverVersions.find((v) => v.is_active));
	let memoryPercent = $derived(memTotal ? Math.min(100, (stats.memory_usage / (memTotal * 1024 * 1024)) * 100) : 0);

	function getBackupVersion(filename: string): string | null {
		const match = filename.match(/_v(.*?)\.zip$/);
		return match ? match[1] : null;
	}

	function isOutdated(filename: string): boolean {
		if (!activeVersion) return false;
		const version = getBackupVersion(filename);
		return version ? version !== activeVersion.version : false;
	}

	// Effects
	$effect(() => {
		if (isOpen && spawnerId !== null && instanceId) {
			startPolling();
		} else {
			stopPolling();
			if (!isOpen) activeTab = 'console';
		}
	});

	$effect(() => {
		isProvisioning = stats.status === 'Provisioning';
	});

	let provTimer: ReturnType<typeof setInterval> | undefined;
	$effect(() => {
		if (isProvisioning) {
			provTimer = setInterval(() => {
				if (provisioningStep < provisioningSteps.length - 1) provisioningStep++;
			}, 2000);
		} else {
			clearInterval(provTimer);
			provisioningStep = 0;
		}
		return () => clearInterval(provTimer);
	});

	$effect(() => {
		if (isOpen && activeTab === 'backups' && spawnerId && instanceId) fetchBackups();
	});

	$effect(() => {
		if (isOpen && activeTab === 'history' && spawnerId && instanceId) fetchHistoryLogs();
	});

	// API Functions
	async function fetchBackups() {
		if (!spawnerId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/backups`);
			if (res.ok) {
				const data = await res.json();
				backups = (data.backups || []).sort((a: any, b: any) => new Date(b.date).getTime() - new Date(a.date).getTime());
			}
		} catch (e) { console.error(e); } finally { isLoadingData = false; }
	}

	async function fetchHistoryLogs() {
		if (!spawnerId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/history`);
			if (res.ok) historyLogs = await res.json();
		} catch (e) { console.error(e); } finally { isLoadingData = false; }
	}

	async function fetchInstanceLogs() {
		if (!spawnerId || !instanceId) return;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`);
			if (res.ok) {
				const data = await res.json();
				if (data.logs) logs = data.logs.split('\n');
			}
		} catch (e) { console.error('Log fetch error:', e); }
	}

	async function fetchStats() {
		if (!spawnerId || !instanceId) return;
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
			if (res.ok) stats = { ...stats, ...(await res.json()) };
		} catch (e) { console.error('Stats fetch error:', e); }
	}

	function startPolling() {
		stopPolling();
		logs = [];
		fetchStats();
		fetchInstanceLogs();
		statsInterval = setInterval(fetchStats, 2000);
		logsInterval = setInterval(() => { if (activeTab === 'console') fetchInstanceLogs(); }, 2000);
	}

	function stopPolling() {
		if (statsInterval) clearInterval(statsInterval);
		if (logsInterval) clearInterval(logsInterval);
		statsInterval = null;
		logsInterval = null;
	}

	type BackupAction = 'create' | 'restore' | 'delete';
	function handleBackupAction(action: BackupAction, filename: string | undefined = undefined) {
		if (!spawnerId || !instanceId) return;
		confirmTitle = action === 'create' ? 'Create Backup' : action === 'restore' ? 'Restore Backup' : 'Delete Backup';
		confirmMessage = action === 'create' 
			? 'Create a new backup? Node must be stopped.' 
			: action === 'restore' 
				? `Restore ${filename}? WARNING: Overwrites all current data!` 
				: `Permanently delete ${filename}?`;
		confirmBtnText = action === 'create' ? 'Start' : action === 'restore' ? 'Restore' : 'Delete';
		isCriticalAction = action !== 'create';

		pendingBackupAction = async () => {
			let url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup`;
			if (action === 'restore') url = `/api/spawners/${spawnerId}/instances/${instanceId}/restore`;
			else if (action === 'delete') url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup/delete`;

			const res = await fetch(url, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: filename ? JSON.stringify({ filename }) : null
			});
			if (res.ok) {
				if (action !== 'restore') await fetchBackups();
				else alert('Restore completed.');
			} else {
				const err = await res.json();
				alert(`Action failed: ${err.error || 'Unknown error'}`);
			}
		};
		isConfirmOpen = true;
	}

	function triggerAction(action: string) {
		confirmTitle = `${action.toUpperCase()} NODE`;
		confirmMessage = `Initiate ${action} sequence for instance ${instanceId}?`;
		confirmBtnText = action === 'delete' ? 'Terminate' : 'Confirm';
		isCriticalAction = action === 'delete' || action === 'stop';
		pendingBackupAction = async () => {
			await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/${action}`, { method: 'POST' });
		};
		isConfirmOpen = true;
	}

	function close() {
		stopPolling();
		onClose();
	}
</script>

{#if isOpen}
	<div 
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop -->
		<div 
			class="absolute inset-0 bg-slate-950/90 backdrop-blur-md"
			onclick={close} 
			role="button" 
			tabindex="0" 
			onkeydown={(e) => e.key === 'Escape' && close()}
			aria-label="Close modal"
		></div>

		<!-- Modal Window -->
		<div 
			class="relative w-full max-w-7xl h-[90vh] flex bg-[#0a0a0a] border border-slate-800 rounded-lg shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden"
			transition:scale={{ start: 0.98, duration: 200, easing: cubicOut }}
		>
			<!-- Tactical Sidebar -->
			<div class="w-72 bg-[#0d0d0d] border-r border-slate-800 flex flex-col shrink-0">
				<div class="p-5 border-b border-slate-800 bg-[#111]">
					<h3 class="text-sm font-black text-slate-200 truncate font-mono tracking-tighter uppercase">{instanceId}</h3>
					<div class="mt-3 flex items-center gap-2">
						<div class="flex items-center gap-1.5 px-2 py-1 rounded bg-black border border-slate-800 text-[9px] font-black uppercase tracking-widest {stats.status === 'Running' ? 'text-emerald-500' : 'text-red-500'}">
							<span class="w-1.5 h-1.5 rounded-full {stats.status === 'Running' ? 'bg-emerald-500 animate-pulse shadow-[0_0_5px_rgba(16,185,129,0.5)]' : 'bg-red-500'}"></span>
							{stats.status}
						</div>
					</div>
				</div>

				<div class="flex-1 overflow-y-auto p-5 space-y-8">
					<!-- Uptime Section -->
					<div>
						<div class="text-[9px] font-black text-slate-600 uppercase tracking-[0.2em] mb-2">System Uptime</div>
						<div class="text-xl font-black font-mono text-blue-500 drop-shadow-[0_0_8px_rgba(59,130,246,0.3)]">
							{formatUptime((stats.uptime || 0) * 1000)}
						</div>
					</div>

					<!-- Resources -->
					<div class="space-y-6">
						<div class="group">
							<div class="flex justify-between text-[9px] font-black uppercase mb-2 tracking-widest">
								<span class="text-slate-500">Core Load</span>
								<span class="text-orange-500">{stats.cpu_percent?.toFixed(1)}%</span>
							</div>
							<div class="h-1 bg-slate-900 overflow-hidden border border-slate-800/50">
								<div class="h-full bg-orange-600 transition-all duration-500 shadow-[0_0_10px_rgba(234,88,12,0.4)]" style="width: {stats.cpu_percent}%"></div>
							</div>
						</div>
						<div class="group">
							<div class="flex justify-between text-[9px] font-black uppercase mb-2 tracking-widest">
								<span class="text-slate-500">Memory Allocation</span>
								<span class="text-purple-500">{formatBytes(stats.memory_usage)}</span>
							</div>
							<div class="h-1 bg-slate-900 overflow-hidden border border-slate-800/50">
								<div class="h-full bg-purple-600 transition-all duration-500 shadow-[0_0_10px_rgba(147,51,234,0.4)]" style="width: {memoryPercent}%"></div>
							</div>
						</div>
					</div>

					<!-- Provisioning Monitor -->
					{#if isProvisioning}
						<div class="pt-6 border-t border-slate-800/50" transition:slide>
							<div class="text-[9px] font-black text-blue-400 uppercase tracking-widest mb-4 flex items-center gap-2">
								<div class="w-2 h-2 rounded-full bg-blue-500 animate-ping"></div>
								Provisioning Sequence
							</div>
							<div class="space-y-3">
								{#each provisioningSteps as step, i}
									<div class="text-[10px] font-bold flex items-center gap-3 transition-colors {i <= provisioningStep ? 'text-slate-300' : 'text-slate-700'}">
										<div class="w-1.5 h-1.5 rounded-full {i < provisioningStep ? 'bg-emerald-500 shadow-[0_0_5px_rgba(16,185,129,0.5)]' : i === provisioningStep ? 'bg-blue-500 animate-pulse' : 'bg-slate-800'}"></div>
										<span class={i < provisioningStep ? 'line-through opacity-50' : ''}>{step}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<!-- Quick Controls -->
				<div class="p-4 border-t border-slate-800 bg-black/40">
					<div class="grid grid-cols-2 gap-2">
						<button 
							onclick={() => triggerAction('start')} 
							disabled={stats.status === 'Running' || isProvisioning} 
							class="col-span-2 py-2.5 bg-emerald-600/10 border border-emerald-500/30 hover:bg-emerald-600/20 disabled:opacity-20 text-emerald-500 font-black text-[10px] uppercase tracking-widest transition-all active:scale-95"
						>
							Execute Init
						</button>
						<button 
							onclick={() => triggerAction('restart')} 
							disabled={stats.status !== 'Running'} 
							class="py-2 bg-slate-800/50 border border-slate-700 hover:bg-slate-800 disabled:opacity-20 text-slate-400 font-black text-[9px] uppercase tracking-wider transition-all"
						>
							Reboot
						</button>
						<button 
							onclick={() => triggerAction('stop')} 
							disabled={stats.status !== 'Running'} 
							class="py-2 bg-slate-800/50 border border-slate-700 hover:bg-red-900/20 hover:text-red-400 disabled:opacity-20 text-slate-400 font-black text-[9px] uppercase tracking-wider transition-all"
						>
							Abort
						</button>
					</div>
				</div>
			</div>

			<!-- Main Terminal/Data Area -->
			<div class="flex-1 flex flex-col min-w-0 bg-[#050505]">
				<!-- Navigation Tabs -->
								<div class="flex border-b border-slate-800 bg-[#0d0d0d] overflow-x-auto no-scrollbar">
									{#each tabs as tab}
										<button 
											onclick={() => activeTab = tab.id}
											class="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] transition-all border-b-2 {activeTab === tab.id ? 'text-blue-500 border-blue-500 bg-blue-500/5' : 'text-slate-600 border-transparent hover:text-slate-400'}"
										>
											{tab.label}
										</button>
									{/each}
								</div>

				<!-- Content Viewport -->
				<div class="flex-1 relative overflow-hidden">
					{#if activeTab === 'console'}
						<div class="absolute inset-0 p-4" in:fade={{ duration: 150 }}>
							<Terminal {logs} title={`root@${instanceId}:~`} />
						</div>
					{:else if activeTab === 'metrics'}
						<div class="p-8 h-full overflow-y-auto" in:fade={{ duration: 150 }}>
							{#if spawnerId !== null && instanceId}
								<ResourceMetricsPanel {spawnerId} {instanceId} {memTotal} height={450} />
							{/if}
						</div>
					{:else if activeTab === 'node_logs'}
						<div class="h-full" in:fade={{ duration: 150 }}>
							{#if spawnerId !== null}
								<LogViewer {spawnerId} isOpen={isOpen} embedded={true} />
							{/if}
						</div>
					{:else if activeTab === 'backups'}
						<div class="p-8 h-full overflow-y-auto" in:fade={{ duration: 150 }}>
							<div class="flex justify-between items-center mb-8 border-b border-slate-800 pb-4">
								<div>
									<h4 class="text-xs font-black text-slate-200 uppercase tracking-[0.3em]">Storage Archives</h4>
									<p class="text-[9px] text-slate-600 font-mono mt-1 uppercase">Node snapshot management</p>
								</div>
								<button 
									onclick={() => handleBackupAction('create')} 
									class="px-4 py-2 bg-blue-600 text-black font-black text-[9px] uppercase tracking-widest hover:bg-blue-500 transition-colors active:translate-y-px"
								>
									Generate Snapshot
								</button>
							</div>
							
							{#if isLoadingData}
								<div class="flex justify-center py-20">
									<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
								</div>
							{:else if backups.length === 0}
								<div class="text-center py-20 text-[10px] font-black text-slate-700 uppercase tracking-[0.3em] border border-dashed border-slate-800 rounded">
									No archives detected
								</div>
							{:else}
								<div class="space-y-2">
									{#each backups as backup}
										<div class="flex items-center justify-between p-4 bg-[#0d0d0d] border border-slate-800 hover:border-slate-700 transition-all group">
											<div class="flex-1 min-w-0">
												<div class="text-xs font-black text-slate-300 font-mono truncate tracking-tight">{backup.filename}</div>
												<div class="text-[9px] text-slate-600 mt-1.5 font-black uppercase tracking-widest flex items-center gap-3">
													<span>{new Date(backup.date).toLocaleString()}</span>
													<span class="w-1 h-1 rounded-full bg-slate-800"></span>
													<span>{formatBytes(backup.size)}</span>
													{#if getBackupVersion(backup.filename)}
														<span class="px-1.5 py-0.5 rounded bg-blue-950/30 text-blue-500 border border-blue-900/50">v{getBackupVersion(backup.filename)}</span>
													{/if}
												</div>
											</div>
											<div class="flex gap-3 ml-4">
												<button 
													onclick={() => handleBackupAction('restore', backup.filename)} 
													class="px-4 py-2 bg-slate-900 border border-slate-800 text-slate-400 hover:text-blue-400 hover:border-blue-900/50 font-black text-[9px] uppercase tracking-widest transition-all"
												>
												Restore
												</button>
												<button 
													onclick={() => handleBackupAction('delete', backup.filename)} 
													class="p-2 text-slate-700 hover:text-red-500 transition-colors"
													aria-label="Delete"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
												</button>
											</div>
										</div>
									{/each}
								</div>
						{/if}
						</div>
					{:else if activeTab === 'history'}
						<div class="h-full overflow-y-auto" in:fade={{ duration: 150 }}>
							<table class="w-full text-left border-collapse font-mono">
								<thead class="sticky top-0 bg-[#0d0d0d] border-b border-slate-800 z-10">
									<tr>
										<th class="px-6 py-4 text-[9px] font-black text-slate-600 uppercase tracking-widest">Protocol Action</th>
										<th class="px-6 py-4 text-[9px] font-black text-slate-600 uppercase tracking-widest text-center">Status</th>
										<th class="px-6 py-4 text-[9px] font-black text-slate-600 uppercase tracking-widest text-right">Timestamp</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-900">
									{#each historyLogs as log}
										<tr class="hover:bg-white/[0.02] transition-colors">
											<td class="px-6 py-4 text-[11px] font-bold text-slate-400 uppercase tracking-tight">{log.action}</td>
											<td class="px-6 py-4 text-center">
												<span class="px-2 py-0.5 rounded text-[9px] font-black uppercase {log.status === 'success' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-red-500/10 text-red-500 border border-red-500/20'}">
													{log.status}
												</span>
											</td>
											<td class="px-6 py-4 text-[10px] text-slate-600 text-right font-black">
												{new Date(log.timestamp).toLocaleString()}
											</td>
										</tr>
									{:else}
										<tr>
											<td colspan="3" class="px-6 py-20 text-center text-[10px] font-black text-slate-700 uppercase tracking-[0.3em]">
												No protocol entries recorded
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>

			<!-- Exit Interface -->
			<button 
				onclick={close} 
				class="absolute top-4 right-4 p-2 text-slate-600 hover:text-white hover:bg-white/10 rounded transition-all z-50 group"
				aria-label="Exit Console"
			>
				<svg class="w-5 h-5 group-hover:rotate-90 transition-transform duration-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
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
</style>
