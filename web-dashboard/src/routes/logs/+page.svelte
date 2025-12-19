<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { 
        AlertTriangle, 
        Shield, 
        Server, 
        Activity, 
        Search, 
        RefreshCw, 
        ChevronLeft, 
        ChevronRight, 
        Info 
    } from 'lucide-svelte';
	import type { SystemLog } from '$lib/types/logs';

	let logs = $state<SystemLog[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let limit = 50;
	let offset = $state(0);
	let category = $state<'All' | 'Internal' | 'Spawner' | 'Security'>('All');
    let selectedLog = $state<SystemLog | null>(null);

    const categories = ['All', 'Internal', 'Spawner', 'Security'];

	async function fetchLogs() {
		loading = true;
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

	onMount(fetchLogs);
</script>

<div class="p-6 h-full flex flex-col overflow-hidden">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6 shrink-0">
        <div>
            <h1 class="text-2xl font-bold text-white flex items-center gap-2">
                <Activity class="w-6 h-6 text-red-400" />
                System Logs
            </h1>
            <p class="text-slate-400 text-sm mt-1">Monitor system-wide errors and events</p>
        </div>
        <div class="flex gap-2">
            <button onclick={fetchLogs} class="p-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg transition-colors">
                <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
            </button>
        </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-2 mb-4 overflow-x-auto shrink-0 pb-2">
        {#each categories as cat}
            <button
                onclick={() => changeCategory(cat)}
                class="px-4 py-2 rounded-lg text-sm font-medium transition-all border border-transparent {category === cat ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 hover:border-slate-600'}"
            >
                {cat}
            </button>
        {/each}
    </div>

    <!-- Table -->
    <div class="flex-1 bg-slate-900/50 border border-slate-800 rounded-xl overflow-hidden flex flex-col">
        <div class="overflow-x-auto overflow-y-auto flex-1 custom-scrollbar">
            <table class="w-full text-left text-sm text-slate-400">
                <thead class="bg-slate-900 text-slate-200 sticky top-0 z-10">
                    <tr>
                        <th class="px-4 py-3 font-semibold">Time</th>
                        <th class="px-4 py-3 font-semibold">Level</th>
                        <th class="px-4 py-3 font-semibold">Category</th>
                        <th class="px-4 py-3 font-semibold">Message</th>
                        <th class="px-4 py-3 font-semibold">Path</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-800">
                    {#if loading && logs.length === 0}
                        <tr><td colspan="5" class="p-8 text-center">Loading...</td></tr>
                    {:else if logs.length === 0}
                        <tr><td colspan="5" class="p-8 text-center">No logs found</td></tr>
                    {:else}
                        {#each logs as log}
                            <tr class="hover:bg-slate-800/50 transition-colors cursor-pointer group" onclick={() => selectedLog = log}>
                                <td class="px-4 py-3 whitespace-nowrap font-mono text-xs">{new Date(log.timestamp).toLocaleString()}</td>
                                <td class="px-4 py-3 font-bold {getLevelColor(log.level)}">{log.level}</td>
                                <td class="px-4 py-3">
                                    <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800 text-xs border border-slate-700">
                                        {log.category}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-slate-300 max-w-md truncate" title={log.message}>{log.message}</td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-500">{log.path || '-'}</td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>

        <!-- Pagination -->
        <div class="p-3 border-t border-slate-800 bg-slate-900 flex justify-between items-center shrink-0">
            <span class="text-xs text-slate-500">
                Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
            </span>
            <div class="flex gap-2">
                <button onclick={prevPage} disabled={offset === 0} class="p-1.5 rounded hover:bg-slate-800 disabled:opacity-50 text-slate-400 hover:text-white transition-colors">
                    <ChevronLeft class="w-4 h-4" />
                </button>
                <button onclick={nextPage} disabled={offset + limit >= total} class="p-1.5 rounded hover:bg-slate-800 disabled:opacity-50 text-slate-400 hover:text-white transition-colors">
                    <ChevronRight class="w-4 h-4" />
                </button>
            </div>
        </div>
    </div>

    <!-- Detail Modal -->
    {#if selectedLog}
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" transition:fade onclick={() => selectedLog = null}>
            <div class="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
                <div class="p-6 border-b border-slate-800 flex justify-between items-start">
                    <div>
                        <h3 class="text-lg font-bold text-white flex items-center gap-2">
                            <span class={getLevelColor(selectedLog.level)}>{selectedLog.level}</span>
                            Log Details
                        </h3>
                        <p class="text-slate-400 text-sm mt-1">{new Date(selectedLog.timestamp).toLocaleString()}</p>
                    </div>
                    <button onclick={() => selectedLog = null} class="text-slate-500 hover:text-white">✕</button>
                </div>
                <div class="p-6 overflow-y-auto space-y-4">
                    <div class="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Category</span>
                            <span class="text-slate-300">{selectedLog.category}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Source</span>
                            <span class="text-slate-300">{selectedLog.source}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Path</span>
                            <span class="font-mono text-slate-300">{selectedLog.path || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Method</span>
                            <span class="font-mono text-slate-300">{selectedLog.method || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Client IP</span>
                            <span class="font-mono text-slate-300">{selectedLog.client_ip || '-'}</span>
                        </div>
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider">Log ID</span>
                            <span class="font-mono text-slate-300">#{selectedLog.id}</span>
                        </div>
                    </div>

                    <div>
                        <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Message</span>
                        <div class="bg-slate-950 p-3 rounded-lg border border-slate-800 text-slate-300 font-mono text-sm whitespace-pre-wrap">
                            {selectedLog.message}
                        </div>
                    </div>

                    {#if selectedLog.details}
                        <div>
                            <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Details</span>
                            <div class="bg-slate-950 p-3 rounded-lg border border-slate-800 text-slate-300 font-mono text-sm whitespace-pre-wrap overflow-x-auto">
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
