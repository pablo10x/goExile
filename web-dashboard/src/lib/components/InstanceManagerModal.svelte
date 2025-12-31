<script lang="ts">
	import { fade, scale, fly, slide } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import { formatBytes, formatUptime } from '$lib/utils';
	import { serverVersions } from '$lib/stores';
	import Terminal from './Terminal.svelte';
	import ResourceMetricsPanel from './ResourceMetricsPanel.svelte';
	import ConfirmDialog from './ConfirmDialog.svelte';
	import LogViewer from './LogViewer.svelte';
	import Icon from './theme/Icon.svelte';

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
		if (isOpen && nodeId !== null && instanceId) {
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
		if (isOpen && activeTab === 'backups' && nodeId && instanceId) fetchBackups();
	});

	$effect(() => {
		if (isOpen && activeTab === 'history' && nodeId && instanceId) fetchHistoryLogs();
	});

	// API Functions
	async function fetchBackups() {
		if (!nodeId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/nodes/${nodeId}/instances/${instanceId}/backups`);
			if (res.ok) {
				const data = await res.json();
				backups = (data.backups || []).sort((a: any, b: any) => new Date(b.date).getTime() - new Date(a.date).getTime());
			}
		} catch (e) { console.error(e); } finally { isLoadingData = false; }
	}

	async function fetchHistoryLogs() {
		if (!nodeId || !instanceId) return;
		isLoadingData = true;
		try {
			const res = await fetch(`/api/nodes/${nodeId}/instances/${instanceId}/history`);
			if (res.ok) historyLogs = await res.json();
		} catch (e) { console.error(e); } finally { isLoadingData = false; }
	}

	async function fetchInstanceLogs() {
		if (!nodeId || !instanceId) return;
		try {
			const res = await fetch(`/api/nodes/${nodeId}/instances/${instanceId}/logs`);
			if (res.ok) {
				const data = await res.json();
				if (data.logs) logs = data.logs.split('\n');
			}
		} catch (e) { console.error('Log fetch error:', e); }
	}

	async function fetchStats() {
		if (!nodeId || !instanceId) return;
		try {
			const res = await fetch(`/api/nodes/${nodeId}/instances/${instanceId}/stats`);
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
		if (!nodeId || !instanceId) return;
		confirmTitle = action === 'create' ? 'Create Backup' : action === 'restore' ? 'Restore Backup' : 'Delete Backup';
		confirmMessage = action === 'create' 
			? 'Create a new backup? Node must be stopped.' 
			: action === 'restore' 
				? `Restore ${filename}? WARNING: Overwrites all current data!` 
				: `Permanently delete ${filename}?`;
		confirmBtnText = action === 'create' ? 'Start' : action === 'restore' ? 'Restore' : 'Delete';
		isCriticalAction = action !== 'create';

		pendingBackupAction = async () => {
			let url = `/api/nodes/${nodeId}/instances/${instanceId}/backup`;
			if (action === 'restore') url = `/api/nodes/${nodeId}/instances/${instanceId}/restore`;
			else if (action === 'delete') url = `/api/nodes/${nodeId}/instances/${instanceId}/backup/delete`;

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
			await fetch(`/api/nodes/${nodeId}/instances/${instanceId}/${action}`, { method: 'POST' });
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
			class="relative w-full max-w-7xl h-full sm:h-[90vh] flex flex-col md:flex-row bg-[var(--terminal-bg)] border border-stone-800 rounded-none shadow-[0_0_50px_rgba(0,0,0,0.8)] overflow-hidden industrial-frame"
			transition:scale={{ start: 0.98, duration: 200, easing: cubicOut }}
		>
			<!-- Tactical Sidebar -->
			<div class="w-full md:w-72 bg-[var(--header-bg)] border-b md:border-b-0 md:border-r border-stone-800 flex flex-col shrink-0 max-h-[40vh] md:max-h-full">
				<div class="p-4 sm:p-6 border-b border-stone-800 bg-stone-900/30 flex justify-between items-center md:block">
					<h3 class="text-[10px] sm:text-xs font-heading font-black text-slate-200 truncate tracking-widest uppercase">{instanceId}</h3>
					<div class="md:mt-4 flex items-center gap-2">
						<div class="flex items-center gap-2 px-3 py-1 bg-black border border-stone-800 text-[8px] sm:text-[10px] font-jetbrains font-black uppercase tracking-widest {stats.status === 'Running' ? 'text-emerald-500' : 'text-red-500'}">
							<span class="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full {stats.status === 'Running' ? 'bg-emerald-500 animate-pulse shadow-[0_0_10px_rgba(16,185,129,0.5)]' : 'bg-red-500'}"></span>
							{stats.status}
						</div>
					</div>
				</div>

				<div class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6 sm:space-y-10 custom-scrollbar hidden md:block">
					<!-- Uptime Section -->
					<div>
						<div class="text-[10px] font-heading font-black uppercase tracking-[0.2em] mb-2" style="color: var(--text-dim)">SYSTEM_UPTIME</div>
						<div class="text-xl sm:text-2xl font-heading font-black text-rust drop-shadow-[0_0_15px_rgba(120,53,15,0.4)]">
							{formatUptime((stats.uptime || 0) * 1000)}
						</div>
					</div>

					<!-- Resources -->
					<div class="space-y-8">
						<div class="group">
							<div class="flex justify-between text-[10px] font-jetbrains font-black uppercase mb-2 tracking-widest">
								<span style="color: var(--text-dim)">Core_Load</span>
								<span class="text-rust">{stats.cpu_percent?.toFixed(1)}%</span>
							</div>
							<div class="h-1.5 bg-stone-950 overflow-hidden border border-stone-800/50">
								<div class="h-full bg-rust transition-all duration-500 shadow-[0_0_15px_rgba(120,53,15,0.5)]" style="width: {stats.cpu_percent}%"></div>
							</div>
						</div>
						<div class="group">
							<div class="flex justify-between text-[10px] font-jetbrains font-black uppercase mb-2 tracking-widest">
								<span style="color: var(--text-dim)">Mem_Allocation</span>
								<span class="text-rust-light">{formatBytes(stats.memory_usage)}</span>
							</div>
							<div class="h-1.5 bg-stone-950 overflow-hidden border border-stone-800/50">
								<div class="h-full bg-rust-light transition-all duration-500 shadow-[0_0_15px_rgba(146,64,14,0.5)]" style="width: {memoryPercent}%"></div>
							</div>
						</div>
					</div>

					<!-- Provisioning Monitor -->
					{#if isProvisioning}
						<div class="pt-8 border-t border-stone-800" transition:slide>
							<div class="text-[10px] font-heading font-black text-rust uppercase tracking-widest mb-6 flex items-center gap-3">
								<div class="w-2 h-2 rounded-full bg-rust animate-ping"></div>
								PROVISIONING_SEQ
							</div>
							<div class="space-y-4">
								{#each provisioningSteps as step, i}
									<div class="text-[11px] font-jetbrains font-bold flex items-center gap-4 transition-colors {i <= provisioningStep ? 'text-stone-300' : 'text-stone-700'}">
										<div class="w-1.5 h-1.5 rounded-none {i < provisioningStep ? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.4)]' : i === provisioningStep ? 'bg-rust animate-pulse shadow-[0_0_10px_rgba(120,53,15,0.4)]' : 'bg-stone-800'}"></div>
										<span class={i < provisioningStep ? 'line-through opacity-40' : ''}>{step}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<!-- Quick Controls -->
				<div class="p-3 sm:p-4 border-t border-stone-800 bg-black/60 shrink-0">
					<div class="grid grid-cols-2 gap-2">
						<button 
							onclick={() => triggerAction('start')} 
							disabled={stats.status === 'Running' || isProvisioning} 
							class="col-span-2 py-2 sm:py-3 bg-rust text-white disabled:opacity-20 font-heading font-black text-[9px] sm:text-[11px] uppercase tracking-widest transition-all active:scale-95 shadow-lg shadow-rust/20 hover:bg-rust-light"
						>
							EXECUTE_INIT
						</button>
						<button 
							onclick={() => triggerAction('restart')} 
							disabled={stats.status !== 'Running'} 
							class="py-2 sm:py-2.5 bg-stone-900 border border-stone-800 hover:border-rust/50 hover:text-rust disabled:opacity-20 text-stone-400 font-heading font-black text-[9px] sm:text-[10px] uppercase tracking-widest transition-all"
						>
							REBOOT
						</button>
						<button 
							onclick={() => triggerAction('stop')} 
							disabled={stats.status !== 'Running'} 
							class="py-2 sm:py-2.5 bg-stone-900 border border-stone-800 hover:border-red-500/50 hover:text-red-500 disabled:opacity-20 text-stone-400 font-heading font-black text-[9px] sm:text-[10px] uppercase tracking-widest transition-all"
						>
							ABORT
						</button>
					</div>
				</div>
			</div>

			<!-- Main Terminal/Data Area -->
			<div class="flex-1 flex flex-col min-w-0 bg-[var(--terminal-bg)] overflow-hidden">
				<!-- Navigation Tabs -->
				<div class="flex border-b border-stone-800 bg-[var(--header-bg)] overflow-x-auto no-scrollbar shrink-0">
					{#each tabs as tab}
						<button 
							onclick={() => activeTab = tab.id}
							class="px-4 sm:px-8 py-4 sm:py-5 text-[9px] sm:text-[11px] font-heading font-black uppercase tracking-[0.2em] transition-all border-b-2 whitespace-nowrap {activeTab === tab.id ? 'text-rust border-rust bg-rust/5' : 'text-stone-600 border-transparent hover:text-stone-400'}"
						>
							{tab.label}
						</button>
					{/each}
				</div>

				<!-- Content Viewport -->
				<div class="flex-1 relative overflow-hidden flex flex-col">
					<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
					
					{#if activeTab === 'console'}
						<div class="flex-1 p-3 sm:p-6 min-h-0" in:fade={{ duration: 150 }}>
							<Terminal {logs} title={`root@${instanceId}:~`} />
						</div>
					{:else if activeTab === 'metrics'}
						<div class="flex-1 p-4 sm:p-10 overflow-y-auto custom-scrollbar" in:fade={{ duration: 150 }}>
							{#if nodeId !== null && instanceId}
								<ResourceMetricsPanel {nodeId} {instanceId} {memTotal} height={450} />
							{/if}
						</div>
					{:else if activeTab === 'node_logs'}
						<div class="flex-1 min-h-0" in:fade={{ duration: 150 }}>
							{#if nodeId !== null}
								<LogViewer {nodeId} isOpen={isOpen} embedded={true} />
							{/if}
						</div>
					{:else if activeTab === 'backups'}
						<div class="flex-1 p-4 sm:p-10 overflow-y-auto custom-scrollbar" in:fade={{ duration: 150 }}>
							<div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-10 border-b border-stone-800 pb-6">
								<div>
									<h4 class="text-sm font-heading font-black text-slate-200 uppercase tracking-[0.3em]">Storage_Archives</h4>
									<p class="font-jetbrains text-[10px] text-stone-600 mt-2 uppercase tracking-widest">Node Snapshot Management Protocol</p>
								</div>
								<button 
									onclick={() => handleBackupAction('create')} 
									class="w-full sm:w-auto px-6 py-3 bg-rust text-white font-heading font-black text-[11px] uppercase tracking-widest hover:bg-rust-light transition-all shadow-lg shadow-rust/20"
								>
									GENERATE_SNAPSHOT
								</button>
							</div>
							
							{#if isLoadingData}
								<div class="flex justify-center py-20">
									<div class="w-8 h-8 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
								</div>
							{:else if backups.length === 0}
								<div class="text-center py-20 text-[11px] font-heading font-black text-stone-700 uppercase tracking-[0.3em] border border-dashed border-stone-800 industrial-frame">
									No archives detected in system storage
								</div>
							{:else}
								<div class="space-y-3">
									{#each backups as backup}
										<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 sm:p-6 bg-stone-900/40 border border-stone-800 hover:border-rust/40 transition-all group industrial-frame gap-4">
											<div class="flex-1 min-w-0">
												<div class="text-xs font-jetbrains font-black text-stone-300 truncate tracking-tighter">{backup.filename}</div>
												<div class="text-[9px] sm:text-[10px] mt-2 font-jetbrains font-bold uppercase tracking-widest flex flex-wrap items-center gap-2 sm:gap-4" style="color: var(--text-dim)">
													<span>{new Date(backup.date).toLocaleString()}</span>
													<span class="hidden sm:block w-1.5 h-1.5 bg-stone-800"></span>
													<span class="text-rust/60">{formatBytes(backup.size)}</span>
													{#if getBackupVersion(backup.filename)}
														<span class="px-2 py-0.5 bg-rust/10 text-rust border border-rust/20">v{getBackupVersion(backup.filename)}</span>
													{/if}
												</div>
											</div>
											<div class="flex gap-4 w-full sm:w-auto">
												<button 
													onclick={() => handleBackupAction('restore', backup.filename)} 
													class="flex-1 sm:flex-initial px-5 py-2.5 bg-stone-950 border border-stone-800 text-stone-500 hover:text-rust hover:border-rust/50 font-heading font-black text-[10px] uppercase tracking-widest transition-all"
												>
												RESTORE
												</button>
												<button 
													onclick={() => handleBackupAction('delete', backup.filename)} 
													class="p-2 text-stone-700 hover:text-red-500 transition-colors"
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
						<div class="flex-1 min-h-0 overflow-y-auto custom-scrollbar" in:fade={{ duration: 150 }}>
							<table class="w-full text-left border-collapse font-jetbrains">
								<thead class="sticky top-0 bg-[var(--header-bg)] border-b border-stone-800 z-10">
									<tr>
										<th class="px-4 sm:px-8 py-3 sm:py-5 text-[8px] sm:text-[10px] font-black text-stone-600 uppercase tracking-widest">Protocol Action</th>
										<th class="px-4 sm:px-8 py-3 sm:py-5 text-[8px] sm:text-[10px] font-black text-stone-600 uppercase tracking-widest text-center">Status</th>
										<th class="px-4 sm:px-8 py-3 sm:py-5 text-[8px] sm:text-[10px] font-black text-stone-600 uppercase tracking-widest text-right">Timestamp</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-stone-900">
									{#each historyLogs as log}
										<tr class="hover:bg-rust/5 transition-colors group">
											<td class="px-4 sm:px-8 py-3 sm:py-5 text-[9px] sm:text-[11px] font-bold text-stone-400 uppercase tracking-tighter group-hover:text-rust transition-colors">{log.action}</td>
											<td class="px-4 sm:px-8 py-3 sm:py-5 text-center">
												<span class="px-2 sm:px-3 py-0.5 sm:py-1 text-[8px] sm:text-[10px] font-black uppercase {log.status === 'success' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-red-500/10 text-red-500 border border-red-500/20'}">
													{log.status}
												</span>
											</td>
											<td class="px-4 sm:px-8 py-3 sm:py-5 text-[8px] sm:text-[10px] text-stone-600 text-right font-black">
												{new Date(log.timestamp).toLocaleTimeString()}
											</td>
										</tr>
									{:else}
										<tr>
											<td colspan="3" class="px-8 py-24 text-center text-[11px] font-heading font-black text-stone-700 uppercase tracking-[0.4em]">
												No protocol entries recorded in system logs
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
				class="absolute top-4 right-4 sm:top-6 sm:right-6 p-2 text-stone-600 hover:text-rust hover:bg-rust/10 transition-all z-50 group"
				aria-label="Exit Console"
			>
				<Icon name="ph:x-bold" size="1.5rem" class="group-hover:rotate-90 transition-transform duration-500" />
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