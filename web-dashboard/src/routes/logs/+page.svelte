<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide, fly, scale } from 'svelte/transition';
	import {
		AlertTriangle,
		Shield,
		Server,
		Activity,
		Search,
		RefreshCw,
		ChevronLeft,
		ChevronRight,
		Info,
		Trash2,
		Check,
		X
	} from 'lucide-svelte';
	import type { SystemLog } from '$lib/types/logs';
	import Icon from '$lib/components/theme/Icon.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

	let logs = $state<SystemLog[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let limit = 50;
	let offset = $state(0);
	let category = $state<'All' | 'Internal' | 'Node' | 'Security'>('All');
	let selectedLog = $state<SystemLog | null>(null);
	let counts = $state<Record<string, number>>({});
	let selectedIds = $state(new Set<number>());

	// Confirmation States
	let isConfirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let isCriticalAction = $state(false);
	let pendingAction = $state<() => Promise<void>>(async () => {});

	const categories = ['All', 'Internal', 'Node', 'Security'];

	async function fetchLogs() {
		loading = true;
		selectedIds = new Set(); // Clear selection on refresh/filter
		try {
			const query = new URLSearchParams({
				limit: limit.toString(),
				offset: offset.toString(),
				category: category === 'All' ? '' : category
			});
			const res = await fetch(`/api/logs?${query}`);
			if (res.ok) {
				const data = await res.json();
				logs = data.logs;
				total = data.total;
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function fetchCounts() {
		try {
			const res = await fetch('/api/logs/counts');
			if (res.ok) {
				counts = await res.json();
				// console.log("Log counts:", counts);
			}
		} catch (e) {
			console.error('Failed to fetch counts:', e);
		}
	}

	function toggleSelection(e: MouseEvent, id: number) {
		e.stopPropagation();
		const newSet = new Set(selectedIds);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		selectedIds = newSet;
	}

	function toggleSelectAll() {
		if (selectedIds.size === logs.length && logs.length > 0) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set(logs.map((l) => l.id));
		}
	}

	function requestDeleteSelected() {
		confirmTitle = 'Purge Selected Signals';
		confirmMessage = `Initiate localized wipe for ${selectedIds.size} encrypted signals? This action is irreversible.`;
		isCriticalAction = true;
		pendingAction = async () => {
			const idsToDelete = Array.from(selectedIds);
			const originalLogs = [...logs];
			logs = logs.filter((l) => !selectedIds.has(l.id));
			selectedIds = new Set();

			try {
				await Promise.all(idsToDelete.map((id) => fetch(`/api/logs/${id}`, { method: 'DELETE' })));
				fetchCounts();
			} catch (e) {
				logs = originalLogs;
				fetchLogs();
				throw e;
			}
		};
		isConfirmOpen = true;
	}

	async function deleteLog(e: MouseEvent, id: number) {
		e.stopPropagation();
		try {
			const res = await fetch(`/api/logs/${id}`, { method: 'DELETE' });
			if (res.ok) {
				logs = logs.filter((l) => l.id !== id);
				if (selectedIds.has(id)) {
					const newSet = new Set(selectedIds);
					newSet.delete(id);
					selectedIds = newSet;
				}
				fetchCounts();
			}
		} catch (e) {
			console.error('Failed to delete log:', e);
		}
	}

	function requestClearLogs() {
		confirmTitle = 'Global Buffer Wipe';
		confirmMessage = 'Executing root-level protocol to clear ALL system logs. System history will be permanently erased.';
		isCriticalAction = true;
		pendingAction = async () => {
			const res = await fetch('/api/logs', { method: 'DELETE' });
			if (res.ok) {
				fetchLogs();
				fetchCounts();
			} else {
				throw new Error('CLEAR_OP_FAILED');
			}
		};
		isConfirmOpen = true;
	}

	function changeCategory(c: string) {
		category = c as any;
		offset = 0;
		fetchLogs();
	}

	function nextPage() {
		if (offset + limit < total) {
			offset += limit;
			fetchLogs();
		}
	}

	function prevPage() {
		if (offset > 0) {
			offset = Math.max(0, offset - limit);
			fetchLogs();
		}
	}

	function getLevelColor(level: string) {
		switch (level) {
			case 'ERROR':
				return 'text-danger';
			case 'FATAL':
				return 'text-danger font-bold';
			case 'WARN':
				return 'text-warning';
			default:
				return 'text-rust-light';
		}
	}

	onMount(() => {
		fetchLogs();
		fetchCounts();
	});
</script>

<div class="w-full h-[calc(100vh-140px)] md:h-[calc(100vh-160px)] flex flex-col overflow-hidden relative border border-stone-800 bg-[var(--terminal-bg)] shadow-2xl industrial-frame">
	<!-- Header -->
	<div class="flex flex-col md:flex-row justify-between items-start md:items-center p-8 border-b border-stone-800 bg-[var(--header-bg)] gap-6 shrink-0 relative z-10">
		<div class="flex items-center gap-6">
			<div class="p-4 bg-rust/10 border border-rust/30 industrial-frame shadow-lg">
				<Icon name="activity" size="2rem" class="text-rust-light" />
			</div>
			<div>
				<h1 class="text-3xl font-heading font-black text-white uppercase tracking-tighter">
					SYSTEM_LOG_INTERCEPT
				</h1>
				<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest font-black mt-1">
					Monitor system-wide anomalies and kernel events
				</p>
			</div>
		</div>
		<div class="flex flex-wrap gap-3 items-center w-full md:w-auto">
			{#if selectedIds.size > 0}
				<button
					onclick={requestDeleteSelected}
					transition:scale={{ duration: 200, start: 0.9 }}
					class="flex-1 md:flex-none px-6 py-3 bg-danger hover:bg-red-500 text-white font-heading font-black text-[10px] uppercase tracking-widest shadow-lg shadow-red-900/20 active:translate-y-px transition-all"
				>
					<Icon name="ph:trash-bold" size="1rem" class="inline mr-2" />
					Purge_Selected ({selectedIds.size})
				</button>
			{/if}

			<button
				onclick={requestClearLogs}
				class="flex-1 md:flex-none px-6 py-3 bg-stone-900 border border-stone-800 text-text-dim hover:text-danger hover:border-red-500/30 font-heading font-black text-[10px] uppercase tracking-widest transition-all active:translate-y-px"
			>
				<Icon name="ph:trash-bold" size="1rem" class="inline mr-2" />
				Global_Wipe
			</button>
			<button
				onclick={() => {
					fetchLogs();
					fetchCounts();
				}}
				class="p-3 bg-stone-900 border border-stone-800 text-text-dim hover:text-rust transition-all shadow-xl active:translate-y-px"
			>
				<Icon name="ph:arrows-clockwise-bold" size="1.25rem" class="{loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Filters -->
	<div class="flex gap-2 p-6 bg-stone-950 border-b border-stone-800 overflow-x-auto shrink-0 no-scrollbar relative z-10 shadow-inner">
		{#each categories as cat}
			<button
				onclick={() => changeCategory(cat)}
				class="px-6 py-2.5 font-heading font-black text-[10px] uppercase tracking-widest transition-all border {category ===
				cat
					? 'bg-rust text-white border-rust shadow-lg shadow-rust/20'
					: 'bg-stone-900 text-text-dim border-stone-800 hover:text-stone-300 hover:border-stone-700'}"
			>
				{cat} <span class="ml-2 opacity-40 font-jetbrains">[{counts[cat] || 0}]</span>
			</button>
		{/each}
	</div>

	<!-- Table Area -->
	<div class="flex-1 flex flex-col min-h-0 bg-black/20 relative">
		<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
		
		<div class="overflow-x-auto overflow-y-auto flex-1 custom-scrollbar relative z-10">
			<table class="w-full text-left font-jetbrains text-[11px] border-collapse">
				<thead class="bg-[var(--header-bg)] text-text-dim sticky top-0 z-10 border-b border-stone-800 shadow-md">
					<tr class="uppercase font-black tracking-widest">
						<th class="px-6 py-4 w-16 text-center border-r border-stone-800/30">
							<button
								onclick={toggleSelectAll}
								class="w-5 h-5 mx-auto border-2 border-stone-800 flex items-center justify-center transition-all {selectedIds.size ===
									logs.length && logs.length > 0
																										? 'bg-rust border-rust shadow-[0_0_10px_rgba(249,115,22,0.3)]'
																										: 'hover:border-stone-600'}"							>
								{#if selectedIds.size === logs.length && logs.length > 0}
									<Icon name="ph:check-bold" size="0.875rem" class="text-white" />
								{/if}
							</button>
						</th>
						<th class="px-6 py-4 border-r border-stone-800/30">Timestamp</th>
						<th class="px-6 py-4 border-r border-stone-800/30">Integrity</th>
						<th class="px-6 py-4 border-r border-stone-800/30">Sector</th>
						<th class="px-6 py-4 border-r border-stone-800/30">Signal_Message</th>
						<th class="px-6 py-4 border-r border-stone-800/30">Kernel_Path</th>
						<th class="px-6 py-4 w-16"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-stone-900">
					{#if loading && logs.length === 0}
						<tr>
							<td colspan="7" class="py-32">
								<div class="flex flex-col items-center justify-center gap-4">
									<div class="w-10 h-10 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg"></div>
									<span class="font-heading font-black text-[11px] text-text-dim uppercase tracking-[0.4em] animate-pulse">Syncing_Active_Logs...</span>
								</div>
							</td>
						</tr>
					{:else if logs.length === 0}
						<tr>
							<td colspan="7" class="py-32">
								<div class="flex flex-col items-center justify-center text-stone-800 gap-4">
									<div class="p-6 border border-dashed border-stone-800 industrial-frame">
										<Icon name="activity" size="3rem" class="opacity-10" />
									</div>
									<span class="font-jetbrains text-[10px] font-black uppercase tracking-[0.3em]">Null_Archive_Reported</span>
								</div>
							</td>
						</tr>
					{:else}
						{#each logs as log (log.id)}
							<tr
								transition:fly={{ x: 20, duration: 300 }}
								class="hover:bg-rust/5 transition-all cursor-pointer group {selectedIds.has(log.id) ? 'bg-rust/10' : ''}"								onclick={() => (selectedLog = log)}
							>
								<td class="px-6 py-4 text-center border-r border-stone-800/20" onclick={(e) => toggleSelection(e, log.id)}>
									<div
										class="w-5 h-5 mx-auto border border-stone-800 flex items-center justify-center transition-all relative overflow-hidden {selectedIds.has(
											log.id
										)
																																		? 'bg-rust border-rust shadow-[0_0_8px_rgba(249,115,22,0.2)]'
																																		: 'group-hover:border-stone-600 bg-stone-950 shadow-inner'}"									>
										{#if selectedIds.has(log.id)}
											<div transition:scale={{ duration: 200, start: 0.5 }}>
												<Icon name="ph:check-bold" size="0.875rem" class="text-white" />
											</div>
										{/if}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap font-bold tabular-nums text-text-dim group-hover:text-stone-300 transition-colors border-r border-stone-800/20"
									>{log.timestamp ? new Date(log.timestamp).toLocaleString([], { hour12: false }) : 'UNKNOWN_T'}</td
								>
								<td class="px-6 py-4 font-black tracking-tighter border-r border-stone-800/20 {getLevelColor(log.level)}">{log.level}</td>
								<td class="px-6 py-4 border-r border-stone-800/20">
									<span
										class="px-2 py-0.5 font-black text-[9px] bg-stone-900 border border-stone-800 text-text-dim uppercase tracking-widest"
									>
										{log.category}
									</span>
								</td>
								<td
									class="px-6 py-4 text-stone-400 group-hover:text-white transition-colors max-w-xl truncate uppercase font-bold tracking-tight border-r border-stone-800/20"
									title={log.message}>{log.message}</td
								>
								<td class="px-6 py-4 font-bold text-text-dim uppercase tracking-tighter border-r border-stone-800/20">{log.path || 'GLOBAL_CORE'}</td>
								<td class="px-6 py-4 text-right">
									<button
										onclick={(e) => deleteLog(e, log.id)}
										class="p-2 text-stone-700 hover:text-danger hover:bg-red-500/10 transition-all opacity-0 group-hover:opacity-100"
										title="Delete Log"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<!-- Pagination Footer -->
		<div
			class="p-6 border-t border-stone-800 bg-[var(--header-bg)] flex flex-col sm:flex-row justify-between items-center gap-6 shrink-0 relative z-10"
		>
			<div class="flex items-center gap-6">
				<span class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest">
					Showing <span class="text-rust">{offset + 1}-{Math.min(offset + limit, total)}</span> // Total <span class="text-white">{total}</span> Signals_Mapped
				</span>
				{#if selectedIds.size > 0}
					<div class="w-px h-4 bg-stone-800"></div>
					<span class="font-jetbrains text-[10px] font-black text-danger uppercase tracking-widest animate-pulse">
						{selectedIds.size} Targets_Locked
					</span>
				{/if}
			</div>
			
			<div class="flex gap-2">
				<button
					onclick={prevPage}
					disabled={offset === 0}
					class="p-3 bg-stone-950 border border-stone-800 hover:border-rust hover:text-rust disabled:opacity-20 text-text-dim transition-all active:translate-x-px shadow-lg"
				>
					<Icon name="ph:caret-left-bold" size="1.25rem" />
				</button>
				<button
					onclick={nextPage}
					disabled={offset + limit >= total}
					class="p-3 bg-stone-950 border border-stone-800 hover:border-rust hover:text-rust disabled:opacity-20 text-text-dim transition-all active:translate-x-px shadow-lg"
				>
					<Icon name="ph:caret-right-bold" size="1.25rem" />
				</button>
			</div>
		</div>
	</div>

	<!-- Detail Modal -->
	{#if selectedLog}
		<div
			class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/90 backdrop-blur-xl"
			transition:fade={{ duration: 150 }}
			onclick={() => (selectedLog = null)}
			role="button"
			tabindex="0"
			onkeydown={(e) => {
				if (e.key === 'Escape') selectedLog = null;
			}}
		>
			<div
				class="bg-[var(--terminal-bg)] border border-stone-800 rounded-none shadow-[0_0_100px_rgba(0,0,0,0.8)] w-full max-w-3xl max-h-[85vh] flex flex-col overflow-hidden industrial-frame"
				onclick={(e) => e.stopPropagation()}
			>
				<div
					class="p-8 border-b border-stone-800 bg-[var(--header-bg)] flex justify-between items-start"
				>
					<div class="flex items-center gap-6">
						<div class="p-3 bg-stone-950 border border-stone-800 industrial-frame">
							<div class={`font-black text-2xl tracking-tighter ${getLevelColor(selectedLog.level)}`}>{selectedLog.level}</div>
						</div>
						<div>
							<h3 class="text-xl font-heading font-black text-white uppercase tracking-tighter flex items-center gap-3">
								INTERCEPT_DETAIL_READOUT
							</h3>
							<p class="font-jetbrains text-[10px] text-text-dim font-bold mt-1 uppercase tracking-widest">
								Captured: {selectedLog.timestamp ? new Date(selectedLog.timestamp).toLocaleString([], { hour12: false }) : 'N/A'}
							</p>
						</div>
					</div>
					<button
						onclick={() => (selectedLog = null)}
						class="p-2 text-text-dim hover:text-white transition-all hover:rotate-90">
						<Icon name="ph:x-bold" size="1.5rem" />
					</button>
				</div>
				<div class="p-10 overflow-y-auto space-y-10 custom-scrollbar bg-[var(--terminal-bg)] relative">
					<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
					
					<div class="grid grid-cols-2 md:grid-cols-3 gap-8 relative z-10">
						{#each [
							{ label: 'Signal_Category', val: selectedLog.category },
							{ label: 'Origin_Source', val: selectedLog.source },
							{ label: 'Kernel_Vector', val: selectedLog.path || 'SYSTEM_CORE' },
							{ label: 'Access_Method', val: selectedLog.method || 'INTERNAL' },
							{ label: 'Remote_Address', val: selectedLog.client_ip || 'DECENTRALIZED' },
							{ label: 'LOG_UID', val: `#${selectedLog.id}` }
						] as meta}
							<div class="space-y-2 bg-stone-900/40 border border-stone-800 p-4 industrial-frame">
								<span class="block font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-widest">{meta.label}</span>
								<span class="font-jetbrains text-[11px] font-black text-stone-200 uppercase tracking-tight break-all">{meta.val}</span>
							</div>
						{/each}
					</div>

					<div class="space-y-4 relative z-10">
						<span class="block font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-[0.3em]">SIGNAL_MESSAGE_RAW</span>
						<div
							class="bg-stone-950 p-6 border border-stone-800 text-stone-300 font-jetbrains text-xs whitespace-pre-wrap leading-relaxed uppercase tracking-wide shadow-inner"
						>
							{selectedLog.message}
						</div>
					</div>

					{#if selectedLog.details}
						<div class="space-y-4 relative z-10">
							<span class="block font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-[0.3em]">EXTENDED_TELEMETRY_DATA</span>
							<div
								class="bg-stone-950 p-6 border border-stone-800 text-stone-400 font-jetbrains text-[11px] whitespace-pre-wrap overflow-x-auto shadow-inner leading-relaxed"
							>
								{selectedLog.details}
							</div>
						</div>
					{/if}
				</div>
				
				<div class="p-8 bg-[var(--header-bg)] border-t border-stone-800 flex justify-end">
					<button 
						onclick={() => (selectedLog = null)}
						class="px-10 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all active:translate-y-px flex items-center gap-3"
					>
						<Icon name="ph:check-bold" size="1rem" />
						Acknowledge_Signal
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<ConfirmDialog
	bind:isOpen={isConfirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	isCritical={isCriticalAction}
	onConfirm={pendingAction}
/>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: rgba(15, 23, 42, 0.3);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(71, 85, 105, 0.5);
		border-radius: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(71, 85, 105, 0.7);
	}
</style>
