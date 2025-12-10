<script lang="ts">
    import { onMount, afterUpdate } from 'svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    
    export let spawnerId: number;

    type LogLevel = 'DBG' | 'INF' | 'WRN' | 'ERR' | 'FTL' | 'PANIC';
    type LogTab = 'all' | 'info' | 'warn' | 'error';

    interface ParsedLogEntry {
        id: number;
        time: string;
        level: LogLevel;
        message: string;
        attributes: string; // The remaining key=value pairs
        originalLine: string;
    }

    let logsRaw: string = '';
    let parsedLogs: ParsedLogEntry[] = [];
    let loading: boolean = true;
    let error: string = '';
    let selectedTab: LogTab = 'all';

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmAction: () => Promise<void> = async () => {};

    // Regex to parse slog log format: time=... level=... msg=... ...
    const logPattern = /time="([^"]+)" level=(DBG|INF|WRN|ERR|FTL|PANIC) msg="([^"]+)"\s*(.*)/;

    function parseLogLine(line: string, index: number): ParsedLogEntry {
        const match = line.match(logPattern);
        if (match) {
            return {
                id: index,
                time: new Date(match[1]).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                level: match[2] as LogLevel,
                message: match[3],
                attributes: match[4],
                originalLine: line
            };
        }
        // Fallback
        return {
            id: index,
            time: '-',
            level: 'INF',
            message: line,
            attributes: '',
            originalLine: line
        };
    }

    function filterLogs(): ParsedLogEntry[] {
        if (selectedTab === 'all') {
            return parsedLogs;
        }
        return parsedLogs.filter(entry => {
            if (selectedTab === 'info') return entry.level === 'INF' || entry.level === 'DBG';
            if (selectedTab === 'warn') return entry.level === 'WRN';
            if (selectedTab === 'error') return entry.level === 'ERR' || entry.level === 'FTL' || entry.level === 'PANIC';
            return true;
        });
    }

    async function fetchLogs() {
        loading = true;
        error = '';
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/logs`);
            if (res.ok) {
                const data = await res.json();
                logsRaw = data.logs || '';
                if (!logsRaw) {
                    parsedLogs = [];
                } else {
                    parsedLogs = logsRaw.split('\n')
                        .filter(line => line.trim().length > 0)
                        .map((line, i) => parseLogLine(line, i));
                }
            } else {
                error = `Failed to fetch logs: ${res.statusText}`;
                parsedLogs = [];
            }
        } catch (e: any) {
            error = `Error: ${e.message}`;
            parsedLogs = [];
        } finally {
            loading = false;
        }
    }

    function requestClearLogs() {
        confirmAction = async () => await clearLogs();
        isConfirmOpen = true;
    }

    async function clearLogs() {
        loading = true;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/logs`, { method: 'DELETE' });
            if (res.ok) {
                logsRaw = '';
                parsedLogs = [];
                await fetchLogs();
            } else {
                alert('Failed to clear logs');
            }
        } catch (e: any) {
            alert(`Error: ${e.message}`);
        } finally {
            loading = false;
        }
    }

    function getLevelClass(level: LogLevel) {
        switch (level) {
            case 'DBG': return 'bg-slate-700 text-slate-300';
            case 'INF': return 'bg-blue-900/40 text-blue-400 border border-blue-500/20';
            case 'WRN': return 'bg-yellow-900/40 text-yellow-400 border border-yellow-500/20';
            case 'ERR': return 'bg-red-900/40 text-red-400 border border-red-500/20';
            case 'FTL': 
            case 'PANIC': return 'bg-red-600 text-white animate-pulse';
            default: return 'bg-slate-700 text-slate-300';
        }
    }

    onMount(() => {
        fetchLogs();
    });

    let logContainer: HTMLElement;
    afterUpdate(() => {
        if (logContainer) {
            logContainer.scrollTop = logContainer.scrollHeight;
        }
    });
</script>

<div class="h-full flex flex-col">
    <!-- Header Controls -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-4">
        <div class="flex gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-800">
            <button 
                class="px-3 py-1.5 text-xs font-medium rounded-md transition-all {selectedTab === 'all' ? 'bg-slate-700 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}" 
                onclick={() => selectedTab = 'all'}
            >All</button>
            <button 
                class="px-3 py-1.5 text-xs font-medium rounded-md transition-all {selectedTab === 'info' ? 'bg-blue-600/20 text-blue-400 shadow-sm' : 'text-slate-400 hover:text-blue-400 hover:bg-slate-800'}" 
                onclick={() => selectedTab = 'info'}
            >Info</button>
            <button 
                class="px-3 py-1.5 text-xs font-medium rounded-md transition-all {selectedTab === 'warn' ? 'bg-yellow-600/20 text-yellow-400 shadow-sm' : 'text-slate-400 hover:text-yellow-400 hover:bg-slate-800'}" 
                onclick={() => selectedTab = 'warn'}
            >Warnings</button>
            <button 
                class="px-3 py-1.5 text-xs font-medium rounded-md transition-all {selectedTab === 'error' ? 'bg-red-600/20 text-red-400 shadow-sm' : 'text-slate-400 hover:text-red-400 hover:bg-slate-800'}" 
                onclick={() => selectedTab = 'error'}
            >Errors</button>
        </div>

        <div class="flex gap-2">
            <button 
                onclick={requestClearLogs} 
                disabled={loading}
                class="px-3 py-1.5 bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 rounded-md text-xs font-semibold transition-colors disabled:opacity-50 flex items-center gap-2"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
                Clear
            </button>
            <button 
                onclick={fetchLogs} 
                disabled={loading}
                class="px-3 py-1.5 bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 rounded-md text-xs font-semibold transition-colors disabled:opacity-50 flex items-center gap-2"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6"></path><path d="M2.5 22v-6h6"></path><path d="M2 11.5a10 10 0 0 1 18.8-4.3L21.5 8"></path><path d="M22 12.5a10 10 0 0 1-18.8 4.2L2.5 16"></path></svg>
                Refresh
            </button>
        </div>
    </div>

    <!-- Logs Container -->
    <div class="flex-1 bg-slate-950 rounded-xl border border-slate-800 overflow-hidden flex flex-col relative shadow-inner">
        {#if loading && !logsRaw}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-500 gap-3">
                <div class="w-8 h-8 border-2 border-slate-600 border-t-blue-500 rounded-full animate-spin"></div>
                <span class="text-sm font-medium">Loading logs...</span>
            </div>
        {:else if error}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-red-400 p-4 text-center">
                <div class="w-10 h-10 bg-red-500/10 rounded-full flex items-center justify-center mb-3">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                </div>
                {error}
            </div>
        {:else}
            <!-- Log List -->
            <div 
                bind:this={logContainer} 
                class="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-slate-900/50 p-2 space-y-0.5"
            >
                {#each filterLogs() as entry (entry.id)}
                    <div class="group flex items-start gap-3 p-1.5 rounded hover:bg-slate-900/80 transition-colors text-xs font-mono leading-tight">
                        <!-- Time -->
                        <div class="text-slate-500 whitespace-nowrap shrink-0 w-16 text-right select-none">
                            {entry.time}
                        </div>
                        
                        <!-- Level -->
                        <div class="shrink-0">
                            <span class={`inline-block px-1.5 py-0.5 rounded text-[10px] font-bold tracking-wider ${getLevelClass(entry.level)}`}>
                                {entry.level}
                            </span>
                        </div>

                        <!-- Message & Attributes -->
                        <div class="flex-1 min-w-0 break-all">
                            <span class="text-slate-300">{entry.message}</span>
                            {#if entry.attributes}
                                <span class="text-slate-500 ml-2 italic">{entry.attributes}</span>
                            {/if}
                        </div>
                    </div>
                {/each}
                {#if filterLogs().length === 0 && !loading}
                    <div class="flex flex-col items-center justify-center h-full text-slate-600 gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 opacity-50" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="9" y1="15" x2="15" y2="15"></line></svg>
                        <span>No logs found.</span>
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title="Clear Spawner Logs"
    message="Are you sure you want to clear these logs? This action cannot be undone."
    confirmText="Clear Logs"
    isCritical={true}
    onConfirm={confirmAction}
/>