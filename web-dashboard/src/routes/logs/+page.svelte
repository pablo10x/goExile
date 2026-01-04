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
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import Button from '$lib/components/Button.svelte';

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
		confirmTitle = 'Purge Selected Logs';
		confirmMessage = `Delete ${selectedIds.size} selected logs? This action is irreversible.`;
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
		confirmTitle = 'Clear All Logs';
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
				return 'text-red-400';
			case 'FATAL':
				return 'text-red-500 font-bold';
			case 'WARN':
				return 'text-amber-400';
			default:
				return 'text-indigo-400';
		}
	}

	onMount(() => {
		fetchLogs();
		fetchCounts();
	});
</script>

<PageHeader 
    title="System Logs" 
    subtitle="Audit & Events" 
    icon="ph:activity-bold"
>
    {#snippet actions()}
        <div class="flex flex-wrap gap-3 items-center">
            {#if selectedIds.size > 0}
                <Button
                    variant="danger"
                    size="md"
                    onclick={requestDeleteSelected}
                    icon="ph:trash-bold"
                >
                    PURGE SELECTED ({selectedIds.size})
                </Button>
            {/if}

            <Button
                variant="secondary"
                size="md"
                onclick={requestClearLogs}
                icon="ph:trash-bold"
            >
                CLEAR ALL
            </Button>
            
            <Button
                variant="secondary"
                size="md"
                onclick={() => {
                    fetchLogs();
                    fetchCounts();
                }}
                loading={loading}
                icon="ph:arrows-clockwise-bold"
            />
        </div>
    {/snippet}
</PageHeader>

<div class="space-y-8">
    <!-- Filters -->
    <div class="flex gap-2 overflow-x-auto no-scrollbar pb-2">
        {#each categories as cat}
            <Button
                variant={category === cat ? 'primary' : 'secondary'}
                size="sm"
                onclick={() => changeCategory(cat)}
                class="whitespace-nowrap"
            >
                {cat} <span class="ml-2 opacity-50 font-mono text-[9px]">[{counts[cat] || 0}]</span>
            </Button>
        {/each}
    </div>

    <Card>
        <div class="overflow-x-auto overflow-y-auto max-h-[65vh] custom-scrollbar">
            <table class="w-full text-left font-jetbrains text-[11px] border-collapse">
                <thead class="bg-neutral-950/40 text-neutral-500 sticky top-0 z-10 border-b border-neutral-800 shadow-sm">
                    <tr class="uppercase font-bold tracking-widest">
                        <th class="px-6 py-5 w-16 text-center">
                            <button
                                onclick={toggleSelectAll}
                                class="w-5 h-5 mx-auto border-2 border-neutral-700 rounded-lg flex items-center justify-center transition-all {selectedIds.size === logs.length && logs.length > 0 ? 'bg-indigo-600 border-indigo-500 shadow-lg shadow-indigo-500/20' : 'hover:border-neutral-500'}"
                            >
                                {#if selectedIds.size === logs.length && logs.length > 0}
                                    <Check class="w-3.5 h-3.5 text-white" />
                                {/if}
                            </button>
                        </th>
                        <th class="px-6 py-5">Timestamp</th>
                        <th class="px-6 py-5">Level</th>
                        <th class="px-6 py-5">Category</th>
                        <th class="px-6 py-5">Message</th>
                        <th class="px-6 py-5">Path</th>
                        <th class="px-6 py-5 w-16"></th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-neutral-800/50">
                    {#if loading && logs.length === 0}
                        <tr>
                            <td colspan="7" class="py-32">
                                <div class="flex flex-col items-center justify-center gap-4">
                                    <div class="w-10 h-10 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin"></div>
                                    <span class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest animate-pulse">Retrieving Logs...</span>
                                </div>
                            </td>
                        </tr>
                    {:else if logs.length === 0}
                        <tr>
                            <td colspan="7" class="py-32">
                                <div class="flex flex-col items-center justify-center text-neutral-700 gap-4">
                                    <div class="p-6 bg-neutral-900/50 rounded-3xl border border-dashed border-neutral-800">
                                        <Info size={40} class="opacity-20 text-indigo-400" />
                                    </div>
                                    <span class="text-[10px] font-bold uppercase tracking-widest">No logs found</span>
                                </div>
                            </td>
                        </tr>
                    {:else}
                        {#each logs as log (log.id)}
                            <tr
                                class="hover:bg-indigo-500/5 transition-all cursor-pointer group {selectedIds.has(log.id) ? 'bg-indigo-500/10' : ''}"
                                onclick={() => (selectedLog = log)}
                            >
                                <td class="px-6 py-4 text-center" onclick={(e) => toggleSelection(e, log.id)}>
                                    <div
                                        class="w-5 h-5 mx-auto border-2 border-neutral-800 rounded-lg flex items-center justify-center transition-all {selectedIds.has(log.id) ? 'bg-indigo-600 border-indigo-500 shadow-md shadow-indigo-500/20' : 'group-hover:border-neutral-600 bg-neutral-950/40'}"
                                    >
                                        {#if selectedIds.has(log.id)}
                                            <Check class="w-3.5 h-3.5 text-white" />
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap font-medium tabular-nums text-neutral-400 group-hover:text-neutral-200 transition-colors">
                                    {log.timestamp ? new Date(log.timestamp).toLocaleString([], { hour12: false }) : 'N/A'}
                                </td>
                                <td class="px-6 py-4 font-black tracking-tight {getLevelColor(log.level)}">{log.level}</td>
                                <td class="px-6 py-4">
                                    <span class="px-2 py-0.5 font-bold text-[9px] bg-neutral-950/40 border border-neutral-800 text-neutral-500 rounded-md uppercase tracking-wider">
                                        {log.category}
                                    </span>
                                </td>
                                <td class="px-6 py-4 text-neutral-300 group-hover:text-white transition-colors max-w-xl truncate font-medium tracking-tight">
                                    {log.message}
                                </td>
                                <td class="px-6 py-4 font-medium text-neutral-500 uppercase tracking-tighter text-[10px]">{log.path || '/'}</td>
                                <td class="px-6 py-4 text-right">
                                    <button
                                        onclick={(e) => deleteLog(e, log.id)}
                                        class="p-2 text-neutral-700 hover:text-red-500 transition-all opacity-0 group-hover:opacity-100"
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

        <div class="p-6 border-t border-neutral-800 bg-neutral-950/20 flex items-center justify-between">
            <div class="flex items-center gap-6">
                <span class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest">
                    Displaying <span class="text-indigo-400 font-black">{offset + 1}-{Math.min(offset + limit, total)}</span> of <span class="text-white font-black">{total}</span>
                </span>
            </div>
            
            <div class="flex gap-2">
                <Button variant="secondary" size="sm" onclick={prevPage} disabled={offset === 0} icon="ph:caret-left-bold" />
                <Button variant="secondary" size="sm" onclick={nextPage} disabled={offset + limit >= total} icon="ph:caret-right-bold" />
            </div>
        </div>
    </Card>
</div>

<!-- Detail Modal -->
{#if selectedLog}
    <div
        class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-neutral-950/80 backdrop-blur-md"
        transition:fade={{ duration: 150 }}
        onclick={() => (selectedLog = null)}
        role="button"
        tabindex="0"
        onkeydown={(e) => {
            if (e.key === 'Escape') selectedLog = null;
        }}
    >
        <div
            class="bg-neutral-900 border border-neutral-800 rounded-3xl shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col overflow-hidden"
            onclick={(e) => e.stopPropagation()}
            onkeydown={(e) => e.stopPropagation()}
            role="document"
            tabindex="0"
        >
            <div
                class="p-8 border-b border-neutral-800 bg-neutral-950/40 flex justify-between items-start"
            >
                <div class="flex items-center gap-6">
                    <div class="p-4 bg-neutral-950 border border-neutral-800 rounded-2xl">
                        <div class={`font-black text-2xl tracking-tighter ${getLevelColor(selectedLog.level)}`}>{selectedLog.level}</div>
                    </div>
                    <div>
                        <h3 class="text-xl font-heading font-black text-white uppercase tracking-tight">
                            Log Detail Readout
                        </h3>
                        <p class="text-[10px] text-neutral-500 font-bold mt-1 uppercase tracking-widest">
                            Captured: {selectedLog.timestamp ? new Date(selectedLog.timestamp).toLocaleString([], { hour12: false }) : 'N/A'}
                        </p>
                    </div>
                </div>
                <button
                    onclick={() => (selectedLog = null)}
                    class="p-2 text-neutral-500 hover:text-white transition-all hover:rotate-90">
                    <X class="w-6 h-6" />
                </button>
            </div>
            <div class="p-10 overflow-y-auto space-y-10 custom-scrollbar bg-neutral-900 relative">
                <div class="grid grid-cols-2 md:grid-cols-3 gap-8 relative z-10">
                    {#each [
                        { label: 'Category', val: selectedLog.category },
                        { label: 'Source', val: selectedLog.source },
                        { label: 'Path', val: selectedLog.path || 'SYSTEM_CORE' },
                        { label: 'Method', val: selectedLog.method || 'INTERNAL' },
                        { label: 'Client IP', val: selectedLog.client_ip || 'INTERNAL' },
                        { label: 'Log ID', val: `#${selectedLog.id}` }
                    ] as meta}
                        <div class="space-y-2 bg-neutral-950/20 border border-neutral-800 p-4 rounded-xl">
                            <span class="block text-[9px] font-bold text-neutral-500 uppercase tracking-widest">{meta.label}</span>
                            <span class="text-[11px] font-bold text-neutral-200 uppercase tracking-tight break-all">{meta.val}</span>
                        </div>
                    {/each}
                </div>

                <div class="space-y-4 relative z-10">
                    <span class="block text-[9px] font-bold text-neutral-500 uppercase tracking-widest">Raw Message</span>
                    <div
                        class="bg-neutral-950 p-6 border border-neutral-800 text-neutral-300 font-mono text-xs whitespace-pre-wrap leading-relaxed rounded-2xl shadow-inner"
                    >
                        {selectedLog.message}
                    </div>
                </div>

                {#if selectedLog.details}
                    <div class="space-y-4 relative z-10">
                        <span class="block text-[9px] font-bold text-neutral-500 uppercase tracking-widest">Extended Data</span>
                        <div
                            class="bg-neutral-950 p-6 border border-neutral-800 text-neutral-400 font-mono text-[11px] whitespace-pre-wrap overflow-x-auto shadow-inner leading-relaxed rounded-2xl"
                        >
                            {selectedLog.details}
                        </div>
                    </div>
                {/if}
            </div>
            
            <div class="p-8 bg-neutral-950/40 border-t border-neutral-800 flex justify-end">
                <Button 
                    onclick={() => (selectedLog = null)}
                    variant="primary"
                    size="md"
                    icon="ph:check-bold"
                >
                    ACKNOWLEDGE
                </Button>
            </div>
        </div>
    </div>
{/if}

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