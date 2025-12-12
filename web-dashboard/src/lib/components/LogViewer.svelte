<script lang="ts">
    import { onMount, onDestroy, afterUpdate } from 'svelte';
    import { slide } from 'svelte/transition';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import { Search, Trash2, RotateCcw, Download, Copy } from 'lucide-svelte';
    import { stats } from '$lib/stores';
    
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
        jsonData?: any; // Full JSON data for structured logs
    }

    let logsRaw: string = '';
    let parsedLogs: ParsedLogEntry[] = [];
    let loading: boolean = true;
    let error: string = '';
    let selectedTab: LogTab = 'all';
    let searchTerm: string = '';
    let autoRefresh: boolean = false;
    let autoRefreshInterval: any;
    let expandedLogId: number | null = null; // For showing detailed JSON view
    let debugLogged = false; // To prevent repeated debug logs

    // Reactive filtered logs
    $: filteredLogs = (() => {
        let filtered = parsedLogs;

        // Filter by tab
        if (selectedTab !== 'all') {
            filtered = filtered.filter(entry => {
                if (selectedTab === 'info') return entry.level === 'INF' || entry.level === 'DBG';
                if (selectedTab === 'warn') return entry.level === 'WRN';
                if (selectedTab === 'error') return entry.level === 'ERR' || entry.level === 'FTL' || entry.level === 'PANIC';
                return true;
            });
        }

        // Filter by search term
        if (searchTerm.trim()) {
            const term = searchTerm.toLowerCase();
            filtered = filtered.filter(entry => {
                // Search in message
                if (entry.message.toLowerCase().includes(term)) return true;

                // Search in attributes
                if (entry.attributes.toLowerCase().includes(term)) return true;

                // Search in JSON data if available
                if (entry.jsonData) {
                    const jsonString = JSON.stringify(entry.jsonData).toLowerCase();
                    if (jsonString.includes(term)) return true;
                }

                return false;
            });
        }

        return filtered;
    })();

    // Reactive counts for tabs
    $: tabCounts = {
        all: parsedLogs.length,
        info: parsedLogs.filter(entry => entry.level === 'INF' || entry.level === 'DBG').length,
        warn: parsedLogs.filter(entry => entry.level === 'WRN').length,
        error: parsedLogs.filter(entry => entry.level === 'ERR' || entry.level === 'FTL' || entry.level === 'PANIC').length
    };

    // Debug parsed levels (run once when logs are loaded)
    $: if (parsedLogs.length > 0 && !debugLogged) {
        console.log('Parsed log levels:', parsedLogs.map(l => ({ level: l.level, message: l.message.substring(0, 30) })));
        console.log('Tab counts:', tabCounts);
        debugLogged = true;
    }

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmAction: () => Promise<void> = async () => {};

    // Fallback regex for other log formats
    const tagPattern = /^([^;\s]+);?\s*(INFO|WARN|ERROR|DEBUG|FATAL|PANIC)\s+(.+)$/i;
    const logPattern = /time="([^"]+)" level=(DBG|INF|WRN|ERR|FTL|PANIC) msg="([^"]+)"\s*(.*)/;

    function parseLogLine(line: string, index: number): ParsedLogEntry {
        // Try to parse as JSON first
        try {
            const jsonLog = JSON.parse(line.trim());
            if (jsonLog && typeof jsonLog === 'object') {
                // Map full level names to abbreviated ones
                const levelMap: { [key: string]: LogLevel } = {
                    'INFO': 'INF',
                    'WARN': 'WRN',
                    'ERROR': 'ERR',
                    'DEBUG': 'DBG',
                    'FATAL': 'FTL',
                    'PANIC': 'PANIC'
                };

                const rawLevel = jsonLog.level;
                const level = levelMap[rawLevel] || rawLevel as LogLevel || 'INF';

                const time = jsonLog.time ? new Date(jsonLog.time).toLocaleTimeString([], {
                    hour12: false,
                    hour: '2-digit', minute: '2-digit', second: '2-digit'
                }) : '-';

                // Create a readable message from the JSON
                const message = jsonLog.msg || '';
                const attributes = Object.keys(jsonLog)
                    .filter(key => !['time', 'level', 'msg'].includes(key))
                    .map(key => `${key}=${JSON.stringify(jsonLog[key])}`)
                    .join(' ');

                return {
                    id: index,
                    time: time,
                    level: level,
                    message: message,
                    attributes: attributes,
                    originalLine: line,
                    jsonData: jsonLog // Store the full JSON for detailed view
                };
            }
        } catch (e) {
            // Not JSON, continue to other parsing methods
        }

        // Try tag format
        const tagMatch = line.match(tagPattern);
        if (tagMatch) {
            const levelMap: { [key: string]: LogLevel } = {
                'INFO': 'INF', 'WARN': 'WRN', 'ERROR': 'ERR',
                'DEBUG': 'DBG', 'FATAL': 'FTL', 'PANIC': 'PANIC'
            };

            const parsedLevel = tagMatch[2];
            const level = levelMap[parsedLevel] || parsedLevel as LogLevel;

            return {
                id: index,
                time: new Date().toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                level: level,
                message: tagMatch[3],
                attributes: '',
                originalLine: line
            };
        }

        // Try slog format
        const match = line.match(logPattern);
        if (match) {
            const levelMap: { [key: string]: LogLevel } = {
                'INFO': 'INF', 'WARN': 'WRN', 'ERROR': 'ERR',
                'DEBUG': 'DBG', 'FATAL': 'FTL', 'PANIC': 'PANIC'
            };

            const parsedLevel = match[2];
            const level = levelMap[parsedLevel] || parsedLevel as LogLevel;

            return {
                id: index,
                time: new Date(match[1]).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                level: level,
                message: match[3],
                attributes: match[4],
                originalLine: line
            };
        }

        // Fallback
        const levelMatch = line.match(/(INFO|WARN|ERROR|DEBUG|FATAL|PANIC|INF|WRN|ERR|FTL|DBG)/i);
        const levelMap: { [key: string]: LogLevel } = {
            'INFO': 'INF', 'INF': 'INF', 'WARN': 'WRN', 'WRN': 'WRN',
            'ERROR': 'ERR', 'ERR': 'ERR', 'DEBUG': 'DBG', 'DBG': 'DBG',
            'FATAL': 'FTL', 'FTL': 'FTL', 'PANIC': 'PANIC'
        };

        const detectedLevel = levelMatch ? levelMap[levelMatch[1]] || 'INF' : 'INF';

        return {
            id: index,
            time: '-',
            level: detectedLevel,
            message: line,
            attributes: '',
            originalLine: line
        };
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

    async function refreshStats() {
        try {
            const res = await fetch('/api/stats');
            if (res.ok) {
                const statsData = await res.json();
                stats.set(statsData);
            }
        } catch (e) {
            console.error('Failed to refresh stats:', e);
        }
    }

    async function clearLogs() {
        loading = true;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/logs`, { method: 'DELETE' });
            if (res.ok) {
                logsRaw = '';
                parsedLogs = [];
                await fetchLogs();
                // Refresh system stats after clearing logs
                await refreshStats();
            } else {
                alert('Failed to clear logs');
            }
        } catch (e: any) {
            alert(`Error: ${e.message}`);
        } finally {
            loading = false;
        }
    }



    function toggleAutoRefresh() {
        autoRefresh = !autoRefresh;
        if (autoRefresh) {
            autoRefreshInterval = setInterval(fetchLogs, 5000); // Refresh every 5 seconds
        } else {
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
                autoRefreshInterval = null;
            }
        }
    }

    function exportLogs() {
        const logText = filteredLogs.map(entry =>
            entry.jsonData ? JSON.stringify(entry.jsonData) : `${entry.time} ${entry.level} ${entry.message}${entry.attributes ? ' ' + entry.attributes : ''}`
        ).join('\n');
        const blob = new Blob([logText], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `spawner-${spawnerId}-logs.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function copyLogEntry(entry: ParsedLogEntry) {
        const text = entry.jsonData ? JSON.stringify(entry.jsonData, null, 2) : `${entry.time} ${entry.level} ${entry.message}${entry.attributes ? ' ' + entry.attributes : ''}`;
        navigator.clipboard.writeText(text);
    }

    onMount(() => {
        fetchLogs();
    });

    onDestroy(() => {
        if (autoRefreshInterval) {
            clearInterval(autoRefreshInterval);
        }
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
    <div class="flex flex-col gap-6 mb-6">
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
            <!-- Tab Filters -->
            <div class="flex gap-1 p-1.5 bg-slate-900/60 rounded-xl border border-slate-800/60 backdrop-blur-sm shadow-lg">
                <button
                    class="px-4 py-2 text-sm font-semibold rounded-lg transition-all duration-200 flex items-center gap-2 {selectedTab === 'all' ? 'bg-slate-700 text-white shadow-md ring-2 ring-slate-600' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'}"
                    onclick={() => { selectedTab = 'all'; searchTerm = ''; }}
                >
                    All
                    <span class="px-1.5 py-0.5 text-xs bg-slate-600/50 rounded-full min-w-[1.25rem] text-center">
                        {tabCounts.all}
                    </span>
                </button>
                <button
                    class="px-4 py-2 text-sm font-semibold rounded-lg transition-all duration-200 flex items-center gap-2 {selectedTab === 'info' ? 'bg-blue-600/20 text-blue-300 shadow-md ring-2 ring-blue-500/30' : 'text-slate-400 hover:text-blue-300 hover:bg-blue-900/20'}"
                    onclick={() => { selectedTab = 'info'; searchTerm = ''; }}
                >
                    Info
                    <span class="px-1.5 py-0.5 text-xs bg-blue-500/20 text-blue-300 rounded-full min-w-[1.25rem] text-center">
                        {tabCounts.info}
                    </span>
                </button>
                <button
                    class="px-4 py-2 text-sm font-semibold rounded-lg transition-all duration-200 flex items-center gap-2 {selectedTab === 'warn' ? 'bg-yellow-600/20 text-yellow-300 shadow-md ring-2 ring-yellow-500/30' : 'text-slate-400 hover:text-yellow-300 hover:bg-yellow-900/20'}"
                    onclick={() => { selectedTab = 'warn'; searchTerm = ''; }}
                >
                    Warnings
                    <span class="px-1.5 py-0.5 text-xs bg-yellow-500/20 text-yellow-300 rounded-full min-w-[1.25rem] text-center">
                        {tabCounts.warn}
                    </span>
                </button>
                <button
                    class="px-4 py-2 text-sm font-semibold rounded-lg transition-all duration-200 flex items-center gap-2 {selectedTab === 'error' ? 'bg-red-600/20 text-red-300 shadow-md ring-2 ring-red-500/30' : 'text-slate-400 hover:text-red-300 hover:bg-red-900/20'}"
                    onclick={() => { selectedTab = 'error'; searchTerm = ''; }}
                >
                    Errors
                    <span class="px-1.5 py-0.5 text-xs bg-red-500/20 text-red-300 rounded-full min-w-[1.25rem] text-center">
                        {tabCounts.error}
                    </span>
                </button>
            </div>

            <!-- Action Buttons -->
            <div class="flex flex-wrap gap-2">
                <button
                    onclick={exportLogs}
                    disabled={loading || parsedLogs.length === 0}
                    class="px-4 py-2 bg-gradient-to-r from-slate-800 to-slate-700 text-slate-300 hover:from-slate-700 hover:to-slate-600 border border-slate-600/50 rounded-lg text-sm font-semibold transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 shadow-md hover:shadow-lg"
                >
                    <Download class="w-4 h-4" />
                    Export
                </button>
                <button
                    onclick={requestClearLogs}
                    disabled={loading}
                    class="px-4 py-2 bg-gradient-to-r from-red-900/20 to-red-800/20 text-red-400 hover:from-red-800/30 hover:to-red-700/30 border border-red-500/30 rounded-lg text-sm font-semibold transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 shadow-md hover:shadow-lg"
                >
                    <Trash2 class="w-4 h-4" />
                    Clear
                </button>
                <button
                    onclick={fetchLogs}
                    disabled={loading}
                    class="px-4 py-2 bg-gradient-to-r from-slate-800 to-slate-700 text-slate-300 hover:from-slate-700 hover:to-slate-600 border border-slate-600/50 rounded-lg text-sm font-semibold transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 shadow-md hover:shadow-lg"
                >
                    <RotateCcw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
                    Refresh
                </button>
                <button
                    onclick={toggleAutoRefresh}
                    class="px-4 py-2 {autoRefresh ? 'bg-gradient-to-r from-green-600/20 to-emerald-600/20 text-green-300 border-green-500/30 shadow-md ring-2 ring-green-500/20' : 'bg-gradient-to-r from-slate-800 to-slate-700 text-slate-300 hover:from-slate-700 hover:to-slate-600 border-slate-600/50'} border rounded-lg text-sm font-semibold transition-all duration-200 flex items-center gap-2 shadow-md hover:shadow-lg"
                >
                    <RotateCcw class="w-4 h-4 {autoRefresh ? 'animate-spin' : ''}" />
                    Auto
                </button>
            </div>
        </div>

        <!-- Search Input -->
        <div class="relative group">
            <input
                bind:value={searchTerm}
                placeholder="Search logs..."
                class="w-full px-4 py-3 pl-12 bg-slate-900/60 border border-slate-700/60 rounded-xl text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/60 text-sm backdrop-blur-sm shadow-lg transition-all duration-200"
            />
            <Search class="absolute left-4 top-1/2 transform -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-blue-400 transition-colors duration-200" />
            {#if searchTerm}
                <button
                    onclick={() => searchTerm = ''}
                    class="absolute right-4 top-1/2 transform -translate-y-1/2 w-4 h-4 text-slate-500 hover:text-slate-300 transition-colors duration-200"
                    title="Clear search"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
            {/if}
        </div>
    </div>

    <!-- Logs Container -->
    <div class="flex-1 bg-gradient-to-br from-slate-950 to-slate-900 rounded-2xl border border-slate-800/60 overflow-hidden flex flex-col relative shadow-2xl backdrop-blur-sm">
        {#if loading && !logsRaw}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-500 gap-4 bg-slate-950/80 backdrop-blur-sm">
                <div class="w-10 h-10 border-3 border-slate-600 border-t-blue-500 rounded-full animate-spin"></div>
                <span class="text-base font-medium text-slate-400">Loading logs...</span>
            </div>
        {:else if error}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-red-400 p-6 text-center bg-slate-950/80 backdrop-blur-sm">
                <div class="w-12 h-12 bg-red-500/10 rounded-full flex items-center justify-center mb-4 border border-red-500/20">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                </div>
                <span class="text-base font-medium">{error}</span>
            </div>
        {:else}
            <!-- Log List -->
            <div
                bind:this={logContainer}
                class="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-600 scrollbar-track-slate-900/30 p-4 space-y-1"
            >
                {#each filteredLogs as entry (entry.id)}
                    <div class="group rounded-xl border border-slate-700/50 bg-slate-800/20 hover:bg-slate-800/40 transition-all duration-200 overflow-hidden">
                        <!-- Main Log Entry -->
                        <div class="flex items-start gap-4 p-4 text-sm relative">
                            <!-- Expand/Collapse Button -->
                            {#if entry.jsonData}
                                <button
                                    onclick={() => expandedLogId = expandedLogId === entry.id ? null : entry.id}
                                    class="shrink-0 p-1 rounded-md hover:bg-slate-700 text-slate-500 hover:text-slate-300 transition-colors"
                                    title={expandedLogId === entry.id ? "Collapse details" : "Expand details"}
                                >
                                    <svg class="w-4 h-4 transform transition-transform {expandedLogId === entry.id ? 'rotate-90' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <path d="M9 18l6-6-6-6"/>
                                    </svg>
                                </button>
                            {/if}

                            <!-- Time -->
                            <div class="text-slate-400 whitespace-nowrap shrink-0 w-20 text-right select-none font-mono text-xs bg-slate-800/50 px-2 py-1 rounded border border-slate-700/50">
                                {entry.time}
                            </div>

                            <!-- Level Badge -->
                            <div class="shrink-0">
                                <span class={`inline-flex items-center px-2 py-1 text-xs font-bold rounded-md border ${
                                    entry.level === 'ERR' || entry.level === 'FTL' || entry.level === 'PANIC' ? 'bg-red-500/10 text-red-300 border-red-500/30' :
                                    entry.level === 'WRN' ? 'bg-yellow-500/10 text-yellow-300 border-yellow-500/30' :
                                    entry.level === 'INF' || entry.level === 'DBG' ? 'bg-blue-500/10 text-blue-300 border-blue-500/30' :
                                    'bg-slate-500/10 text-slate-300 border-slate-500/30'
                                }`}>
                                    {entry.level}
                                </span>
                            </div>

                            <!-- Message -->
                            <div class="flex-1 min-w-0 break-words text-slate-200 leading-relaxed">
                                <span class="text-slate-100 font-medium">{entry.message}</span>
                                {#if entry.attributes}
                                    <div class="text-slate-400 text-xs mt-1 font-mono">
                                        {entry.attributes}
                                    </div>
                                {/if}
                            </div>

                            <!-- Actions -->
                            <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                <button
                                    onclick={() => copyLogEntry(entry)}
                                    class="p-2 rounded-lg hover:bg-slate-700 text-slate-500 hover:text-slate-200 transition-colors"
                                    title="Copy log entry"
                                >
                                    <Copy class="w-4 h-4" />
                                </button>
                            </div>
                        </div>

                        <!-- Expanded JSON Details -->
                        {#if expandedLogId === entry.id && entry.jsonData}
                            <div class="border-t border-slate-700/50 bg-slate-900/50 p-4" transition:slide={{ duration: 200 }}>
                                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 text-sm">
                                    {#each Object.entries(entry.jsonData) as [key, value]}
                                        <div class="bg-slate-800/50 rounded-lg p-3 border border-slate-700/30">
                                            <div class="text-slate-400 text-xs uppercase tracking-wide font-semibold mb-1">
                                                {key}
                                            </div>
                                            <div class="text-slate-200 font-mono text-sm break-all">
                                                {#if typeof value === 'object'}
                                                    {JSON.stringify(value, null, 2)}
                                                {:else}
                                                    {String(value)}
                                                {/if}
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            </div>
                        {/if}
                    </div>
                {/each}

                <!-- Empty State Messages -->
                {#if filteredLogs.length === 0 && !loading}
                    <div class="flex flex-col items-center justify-center py-16 text-center">
                        <div class="w-20 h-20 bg-slate-800/30 rounded-full flex items-center justify-center border border-slate-700/30 mb-6">
                            {#if selectedTab === 'all'}
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 opacity-50 text-slate-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="9" y1="15" x2="15" y2="15"></line></svg>
                            {:else if selectedTab === 'info'}
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 opacity-50 text-blue-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4"></path><path d="M12 8h.01"></path></svg>
                            {:else if selectedTab === 'warn'}
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 opacity-50 text-yellow-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
                            {:else if selectedTab === 'error'}
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 opacity-50 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                            {/if}
                        </div>

                        {#if selectedTab === 'all'}
                            <h3 class="text-xl font-semibold text-slate-400 mb-2">No logs found</h3>
                            <p class="text-slate-600 max-w-md">There are currently no log entries to display. Logs will appear here as the spawner generates them.</p>
                        {:else if selectedTab === 'info'}
                            <h3 class="text-xl font-semibold text-slate-400 mb-2">No info logs</h3>
                            <p class="text-slate-600 max-w-md">There are no informational or debug log entries at the moment. These typically contain general operational messages.</p>
                        {:else if selectedTab === 'warn'}
                            <h3 class="text-xl font-semibold text-slate-400 mb-2">No warning logs</h3>
                            <p class="text-slate-600 max-w-md">Great! There are no warning messages. Warnings typically indicate potential issues that don't prevent operation.</p>
                        {:else if selectedTab === 'error'}
                            <h3 class="text-xl font-semibold text-slate-400 mb-2">No error logs</h3>
                            <p class="text-slate-600 max-w-md">Excellent! There are no error messages. Errors indicate serious issues that may affect spawner operation.</p>
                        {/if}

                        {#if searchTerm}
                            <div class="mt-4 p-3 bg-slate-800/50 rounded-lg border border-slate-700/50">
                                <p class="text-sm text-slate-500">Try adjusting your search term or clearing the search to see more logs.</p>
                            </div>
                        {/if}
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