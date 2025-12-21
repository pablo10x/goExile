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

	let logs = $state<SystemLog[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let limit = 50;
	let offset = $state(0);
	let category = $state<'All' | 'Internal' | 'Spawner' | 'Security'>('All');
    let selectedLog = $state<SystemLog | null>(null);
    let counts = $state<Record<string, number>>({});
    let selectedIds = $state(new Set<number>());

    const categories = ['All', 'Internal', 'Spawner', 'Security'];

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
            console.error("Failed to fetch counts:", e);
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
            selectedIds = new Set(logs.map(l => l.id));
        }
    }

    async function deleteSelected() {
        const idsToDelete = Array.from(selectedIds);
        if (idsToDelete.length === 0) return;
        
        // Optimistic UI update
        const originalLogs = [...logs];
        logs = logs.filter(l => !selectedIds.has(l.id));
        selectedIds = new Set();

        // Perform deletions in parallel
        try {
            await Promise.all(idsToDelete.map(id => fetch(`/api/logs/${id}`, { method: 'DELETE' })));
            
            // Update counts approximately
            if (counts[category]) counts[category] = Math.max(0, counts[category] - idsToDelete.length);
            if (counts['All']) counts['All'] = Math.max(0, counts['All'] - idsToDelete.length);
            
            fetchCounts(); // accurate count sync
        } catch (e) {
            console.error("Failed to delete selected:", e);
            logs = originalLogs; // Revert on failure
            fetchLogs();
        }
    }

    async function deleteLog(e: MouseEvent, id: number) {
        e.stopPropagation();
        // Removed confirmation as requested
        
        try {
            const res = await fetch(`/api/logs/${id}`, { method: 'DELETE' });
            if (res.ok) {
                // Optimistically remove from UI to feel instant
                logs = logs.filter(l => l.id !== id);
                if (selectedIds.has(id)) {
                    const newSet = new Set(selectedIds);
                    newSet.delete(id);
                    selectedIds = newSet;
                }

                if (counts[category]) counts[category]--;
                if (counts['All']) counts['All']--;
                
                // Only fetch counts to keep numbers accurate, but don't refetch logs
                // as it interrupts the exit animation
                fetchCounts();
            }
        } catch (e) {
            console.error("Failed to delete log:", e);
        }
    }

    async function clearLogs() {
        if (!confirm('Are you sure you want to clear ALL logs? This cannot be undone.')) return;
        
        try {
            const res = await fetch('/api/logs', { method: 'DELETE' });
            if (res.ok) {
                fetchLogs();
                fetchCounts();
            }
        } catch (e) {
            console.error("Failed to clear logs:", e);
        }
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
        switch(level) {
            case 'ERROR': return 'text-red-400';
            case 'FATAL': return 'text-rose-500 font-bold';
            case 'WARN': return 'text-orange-400';
            default: return 'text-blue-400';
        }
    }

	onMount(() => {
        fetchLogs();
        fetchCounts();
    });
</script>

<div class="p-6 h-full flex flex-col overflow-hidden">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6 shrink-0">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Activity class="w-6 h-6 text-red-400" />
                System Logs
            </h1>
            <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">Monitor system-wide errors and events</p>
        </div>
        <div class="flex gap-2 items-center">
            {#if selectedIds.size > 0}
                <button 
                    onclick={deleteSelected}
                    transition:scale={{ duration: 200, start: 0.9 }}
                    class="p-2 bg-red-600 hover:bg-red-500 text-slate-900 dark:text-white rounded-lg transition-colors flex items-center gap-2 px-4 shadow-lg shadow-red-900/20 mr-2"
                >
                    <Trash2 class="w-4 h-4" />
                    <span class="text-xs font-bold">Delete ({selectedIds.size})</span>
                </button>
            {/if}

            <button onclick={clearLogs} class="p-2 bg-red-900/30 hover:bg-red-900/50 text-red-400 hover:text-red-200 rounded-lg transition-colors flex items-center gap-2 px-3">
                <Trash2 class="w-4 h-4" />
                <span class="text-xs font-semibold">Clear All</span>
            </button>
            <button onclick={() => { fetchLogs(); fetchCounts(); }} class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-900 dark:text-white rounded-lg transition-colors">
                <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
            </button>
        </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-2 mb-4 overflow-x-auto shrink-0 pb-2">
        {#each categories as cat}
            <button
                onclick={() => changeCategory(cat)}
                class="px-4 py-2 rounded-lg text-sm font-medium transition-all border border-transparent {category === cat ? 'bg-blue-600 text-slate-900 dark:text-white shadow-lg shadow-blue-900/20' : 'bg-slate-800 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 hover:border-slate-600'}"
            >
                {cat} <span class="ml-1 opacity-60 text-xs">({counts[cat] || 0})</span>
            </button>
        {/each}
    </div>

    <!-- Table -->
    <div class="flex-1 bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden flex flex-col">
        <div class="overflow-x-auto overflow-y-auto flex-1 custom-scrollbar">
            <table class="w-full text-left text-sm text-slate-500 dark:text-slate-400">
                <thead class="bg-slate-900 text-slate-800 dark:text-slate-200 sticky top-0 z-10">
                    <tr>
                        <th class="px-4 py-3 font-semibold w-10 text-center">
                            <button 
                                onclick={toggleSelectAll}
                                class="w-5 h-5 rounded border border-slate-600 flex items-center justify-center transition-all {selectedIds.size === logs.length && logs.length > 0 ? 'bg-blue-600 border-blue-500' : 'hover:border-slate-400'}"
                            >
                                {#if selectedIds.size === logs.length && logs.length > 0}
                                    <Check class="w-3.5 h-3.5 text-slate-900 dark:text-white" />
                                {/if}
                            </button>
                        </th>
                        <th class="px-4 py-3 font-semibold">Time</th>
                        <th class="px-4 py-3 font-semibold">Level</th>
                        <th class="px-4 py-3 font-semibold">Category</th>
                        <th class="px-4 py-3 font-semibold">Message</th>
                        <th class="px-4 py-3 font-semibold">Path</th>
                        <th class="px-4 py-3 font-semibold w-10"></th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-800">
                    {#if loading && logs.length === 0}
                        <tr><td colspan="7" class="p-8 text-center">Loading...</td></tr>
                    {:else if logs.length === 0}
                        <tr><td colspan="7" class="p-8 text-center">No logs found</td></tr>
                    {:else}
                        {#each logs as log (log.id)}
                            <tr 
                                transition:fly={{ x: 20, duration: 300 }} 
                                class="hover:bg-slate-800/50 transition-colors cursor-pointer group {selectedIds.has(log.id) ? 'bg-blue-900/10' : ''}" 
                                onclick={() => selectedLog = log}
                            >
                                <td class="px-4 py-3 text-center" onclick={(e) => toggleSelection(e, log.id)}>
                                    <div class="w-5 h-5 mx-auto rounded border border-slate-600 flex items-center justify-center transition-all relative overflow-hidden {selectedIds.has(log.id) ? 'bg-blue-600 border-blue-500' : 'group-hover:border-slate-400 bg-slate-800/50'}">
                                        {#if selectedIds.has(log.id)}
                                            <div transition:scale={{ duration: 200, start: 0.5 }}>
                                                <Check class="w-3.5 h-3.5 text-slate-900 dark:text-white" />
                                            </div>
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-4 py-3 whitespace-nowrap font-mono text-xs text-slate-500">{log.timestamp ? new Date(log.timestamp).toLocaleString() : '-'}</td>
                                <td class="px-4 py-3 font-bold {getLevelColor(log.level)}">{log.level}</td>
                                <td class="px-4 py-3">
                                    <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800 text-xs border border-slate-300 dark:border-slate-700">
                                        {log.category}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-slate-700 dark:text-slate-300 max-w-md truncate" title={log.message}>{log.message}</td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-500">{log.path || '-'}</td>
                                <td class="px-4 py-3 text-right">
                                    <button 
                                        onclick={(e) => deleteLog(e, log.id)} 
                                        class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 rounded transition-all opacity-0 group-hover:opacity-100"
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

        <!-- Pagination -->
        <div class="p-3 border-t border-slate-200 dark:border-slate-800 bg-slate-900 flex justify-between items-center shrink-0">
            <span class="text-xs text-slate-500">
                Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
            </span>
            <div class="flex gap-2">
                <button onclick={prevPage} disabled={offset === 0} class="p-1.5 rounded hover:bg-slate-800 disabled:opacity-50 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white transition-colors">
                    <ChevronLeft class="w-4 h-4" />
                </button>
                <button onclick={nextPage} disabled={offset + limit >= total} class="p-1.5 rounded hover:bg-slate-800 disabled:opacity-50 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white transition-colors">
                    <ChevronRight class="w-4 h-4" />
                </button>
            </div>
        </div>
    </div>

    <!-- Detail Modal -->
    {#if selectedLog}
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" transition:fade onclick={() => selectedLog = null} role="button" tabindex="0" onkeydown={(e) => { if (e.key === 'Escape') selectedLog = null; }}>
            <div class="bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
                <div class="p-6 border-b border-slate-200 dark:border-slate-800 flex justify-between items-start">
                    <div>
                        <h3 class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
                            <span class={getLevelColor(selectedLog.level)}>{selectedLog.level}</span>
                            Log Details
                        </h3>
                        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">{selectedLog.timestamp ? new Date(selectedLog.timestamp).toLocaleString() : '-'}</p>
                    </div>
                    <button onclick={() => selectedLog = null} class="text-slate-500 hover:text-slate-900 dark:text-white">✕</button>
                </div>
                <div class="p-6 overflow-y-auto space-y-4">
                    <div class="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Category</span>
                            <span class="text-slate-700 dark:text-slate-300">{selectedLog.category}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Source</span>
                            <span class="text-slate-700 dark:text-slate-300">{selectedLog.source}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Path</span>
                            <span class="font-mono text-slate-700 dark:text-slate-300">{selectedLog.path || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Method</span>
                            <span class="font-mono text-slate-700 dark:text-slate-300">{selectedLog.method || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Client IP</span>
                            <span class="font-mono text-slate-700 dark:text-slate-300">{selectedLog.client_ip || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Log ID</span>
                            <span class="font-mono text-slate-700 dark:text-slate-300">#{selectedLog.id}</span>
                        </div>
                    </div>

                    <div>
                        <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Message</span>
                        <div class="bg-white dark:bg-slate-950 p-3 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 font-mono text-sm whitespace-pre-wrap">
                            {selectedLog.message}
                        </div>
                    </div>

                    {#if selectedLog.details}
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Details</span>
                            <div class="bg-white dark:bg-slate-950 p-3 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 font-mono text-sm whitespace-pre-wrap overflow-x-auto">
                                {selectedLog.details}
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    {/if}
</div>

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
