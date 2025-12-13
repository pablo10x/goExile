<script lang="ts">
    import { onMount, createEventDispatcher, tick } from 'svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import { Search, Download, Copy, RefreshCw, X, ChevronDown, ChevronUp, Clock, BarChart3 } from 'lucide-svelte';
    
    export let spawnerId: number;

    type LogLevel = 'DBG' | 'INF' | 'WRN' | 'ERR' | 'FTL' | 'PANIC';
    type LogTab = 'all' | 'info' | 'warn' | 'error';

    interface ParsedLogEntry {
        id: number;
        time: string;
        level: LogLevel;
        message: string;
        attributes: string;
        originalLine: string;
        expanded?: boolean;
        timestamp: number;
    }

    interface LogStats {
        total: number;
        debug: number;
        info: number;
        warn: number;
        error: number;
        fatal: number;
        panic: number;
    }

    let logsRaw: string = '';
    let parsedLogs: ParsedLogEntry[] = [];
    let filteredLogs: ParsedLogEntry[] = [];
    let loading: boolean = true;
    let error: string = '';
    let selectedTab: LogTab = 'all';
    let searchTerm: string = '';
    let caseSensitive: boolean = false;
    let regexSearch: boolean = false;
    let autoRefresh: boolean = false;
    let refreshInterval: number = 5000; // 5 seconds
    let refreshTimer: ReturnType<typeof setInterval>;
    let showStats: boolean = false;
    let logStats: LogStats = {
        total: 0,
        debug: 0,
        info: 0,
        warn: 0,
        error: 0,
        fatal: 0,
        panic: 0
    };

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmAction: () => Promise<void> = async () => {};

    // Search state
    let searchResults: number[] = [];
    let currentSearchIndex = -1;
    let totalSearchResults = 0;

    const dispatch = createEventDispatcher();

    // Regex to parse slog log format: time=... level=... msg=... ...
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
            const timestamp = new Date(match[1]).getTime();
            return {
                id: index,
                time: new Date(match[1]).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                level: level,
                message: match[3],
                attributes: match[4],
                originalLine: line,
                timestamp,
                expanded: false
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
            originalLine: line,
            timestamp: Date.now(),
            expanded: false
        };
    }

    function calculateStats(logs: ParsedLogEntry[]): LogStats {
        const stats = { ...logStats };
        stats.total = logs.length;
        stats.debug = logs.filter(l => l.level === 'DBG').length;
        stats.info = logs.filter(l => l.level === 'INF').length;
        stats.warn = logs.filter(l => l.level === 'WRN').length;
        stats.error = logs.filter(l => l.level === 'ERR').length;
        stats.fatal = logs.filter(l => l.level === 'FTL').length;
        stats.panic = logs.filter(l => l.level === 'PANIC').length;
        return stats;
    }

    function filterLogs(): ParsedLogEntry[] {
        let filtered = parsedLogs;

        // Filter by log level
        if (selectedTab !== 'all') {
            filtered = filtered.filter(entry => {
                if (selectedTab === 'info') return entry.level === 'INF' || entry.level === 'DBG';
                if (selectedTab === 'warn') return entry.level === 'WRN';
                if (selectedTab === 'error') return entry.level === 'ERR' || entry.level === 'FTL' || entry.level === 'PANIC';
                return true;
            });
        }

        // Apply search filter
        if (searchTerm.trim()) {
            const searchRegex = createSearchRegex();
            filtered = filtered.filter(entry => {
                const searchText = `${entry.time} ${entry.level} ${entry.message} ${entry.attributes}`;
                return searchRegex.test(searchText);
            });
        }

        return filtered;
    }

    function createSearchRegex(): RegExp {
        if (!searchTerm.trim()) return /./;
        
        try {
            let pattern = searchTerm;
            if (!regexSearch) {
                pattern = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            }
            const flags = caseSensitive ? 'g' : 'gi';
            return new RegExp(pattern, flags);
        } catch (e) {
            // Invalid regex, fall back to string search
            return new RegExp(searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), caseSensitive ? 'g' : 'gi');
        }
    }

    function highlightSearchTerm(text: string): string {
        if (!searchTerm.trim()) return text;
        
        const regex = createSearchRegex();
        return text.replace(regex, match => `<mark class="bg-yellow-300 text-black px-0.5 rounded">${match}</mark>`);
    }

    function performSearch() {
        if (!searchTerm.trim()) {
            searchResults = [];
            currentSearchIndex = -1;
            totalSearchResults = 0;
            return;
        }

        const regex = createSearchRegex();
        searchResults = [];
        
        filteredLogs.forEach((entry, index) => {
            const searchText = `${entry.time} ${entry.level} ${entry.message} ${entry.attributes}`;
            if (regex.test(searchText)) {
                searchResults.push(index);
            }
        });

        totalSearchResults = searchResults.length;
        currentSearchIndex = totalSearchResults > 0 ? 0 : -1;
    }

    function navigateSearch(direction: 'next' | 'prev') {
        if (searchResults.length === 0) return;
        
        if (direction === 'next') {
            currentSearchIndex = (currentSearchIndex + 1) % searchResults.length;
        } else {
            currentSearchIndex = currentSearchIndex <= 0 ? searchResults.length - 1 : currentSearchIndex - 1;
        }
        
        const targetIndex = searchResults[currentSearchIndex];
        const targetElement = document.querySelector(`[data-log-index="${targetIndex}"]`);
        if (targetElement) {
            targetElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
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
                filteredLogs = filterLogs();
                logStats = calculateStats(parsedLogs);
                performSearch();
            } else {
                error = `Failed to fetch logs: ${res.statusText}`;
                parsedLogs = [];
                filteredLogs = [];
            }
        } catch (e: any) {
            error = `Error: ${e.message}`;
            parsedLogs = [];
            filteredLogs = [];
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
                filteredLogs = [];
                logStats = calculateStats([]);
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

    function exportLogs(format: 'json' | 'txt') {
        if (format === 'json') {
            const data = JSON.stringify(filteredLogs, null, 2);
            downloadFile(data, `logs_${spawnerId}_${new Date().toISOString().slice(0, 19)}.json`, 'application/json');
        } else {
            const text = filteredLogs.map(log => log.originalLine).join('\n');
            downloadFile(text, `logs_${spawnerId}_${new Date().toISOString().slice(0, 19)}.txt`, 'text/plain');
        }
    }

    function copyToClipboard() {
        const text = filteredLogs.map(log => log.originalLine).join('\n');
        navigator.clipboard.writeText(text).then(() => {
            // Show success feedback
            dispatch('notification', { message: 'Logs copied to clipboard', type: 'success' });
        }).catch(err => {
            console.error('Failed to copy logs:', err);
        });
    }

    function downloadFile(content: string, filename: string, mimeType: string) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function toggleAutoRefresh() {
        if (refreshTimer) {
            clearInterval(refreshTimer);
            refreshTimer = null as any;
        }
        
        if (autoRefresh) {
            refreshTimer = setInterval(fetchLogs, refreshInterval);
        }
    }

    function getLevelClass(level: LogLevel) {
        switch (level) {
            case 'DBG': return 'bg-slate-700 text-slate-300 border border-slate-600';
            case 'INF': return 'bg-blue-900/40 text-blue-400 border border-blue-500/20';
            case 'WRN': return 'bg-yellow-900/40 text-yellow-400 border border-yellow-500/20';
            case 'ERR': return 'bg-red-900/40 text-red-400 border border-red-500/20';
            case 'FTL': 
            case 'PANIC': return 'bg-red-600 text-white border border-red-500 animate-pulse';
            default: return 'bg-slate-700 text-slate-300 border border-slate-600';
        }
    }

    function getLevelIcon(level: LogLevel) {
        switch (level) {
            case 'DBG': return '🔍';
            case 'INF': return 'ℹ️';
            case 'WRN': return '⚠️';
            case 'ERR': return '❌';
            case 'FTL': return '💀';
            case 'PANIC': return '🚨';
            default: return '📝';
        }
    }

    function toggleLogEntry(entry: ParsedLogEntry) {
        entry.expanded = !entry.expanded;
    }

    // Keyboard shortcuts
    function handleKeydown(e: KeyboardEvent) {
        if (e.ctrlKey || e.metaKey) {
            switch (e.key) {
                case 'f':
                    e.preventDefault();
                    document.getElementById('log-search')?.focus();
                    break;
                case 'c':
                    if (e.shiftKey) {
                        e.preventDefault();
                        copyToClipboard();
                    }
                    break;
            }
        } else if (e.key === '/') {
            e.preventDefault();
            document.getElementById('log-search')?.focus();
        } else if (e.key === 'n' && !e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            navigateSearch('next');
        } else if (e.key === 'p' && !e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            navigateSearch('prev');
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
        document.addEventListener('keydown', handleKeydown);
        
        return () => {
            document.removeEventListener('keydown', handleKeydown);
            if (refreshTimer) {
                clearInterval(refreshTimer);
            }
        };
    });

    // Reactive updates
    $: filteredLogs = filterLogs();
    $: if (filteredLogs) {
        performSearch();
        logStats = calculateStats(parsedLogs);
    }
    $: toggleAutoRefresh();

    // Auto-scroll to bottom on new logs (only if user is at bottom)
    let shouldAutoScroll = true;
    let logContainer: HTMLElement;

    function handleScroll() {
        if (logContainer) {
            const { scrollTop, scrollHeight, clientHeight } = logContainer;
            shouldAutoScroll = scrollTop + clientHeight >= scrollHeight - 100;
        }
    }

    async function scrollToBottom() {
        if (logContainer && shouldAutoScroll) {
            await tick();
            logContainer.scrollTop = logContainer.scrollHeight;
        }
    }

    $: if (filteredLogs) {
        scrollToBottom();
    }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="h-full flex flex-col bg-slate-950">
    <!-- Enhanced Header Controls -->
    <div class="flex flex-col gap-4 p-4 border-b border-slate-800">
        <!-- Top Row: Tabs and Actions -->
        <div class="flex flex-wrap justify-between items-start gap-4">
            <!-- Log Level Tabs with Counts -->
            <div class="flex gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-800">
                <button 
                    class="px-3 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1 {selectedTab === 'all' ? 'bg-slate-700 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}" 
                    onclick={() => selectedTab = 'all'}
                >
                    All <span class="bg-slate-600 px-1.5 rounded-full text-xs">{logStats.total}</span>
                </button>
                <button 
                    class="px-3 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1 {selectedTab === 'info' ? 'bg-blue-600/20 text-blue-400 shadow-sm' : 'text-slate-400 hover:text-blue-400 hover:bg-slate-800'}" 
                    onclick={() => selectedTab = 'info'}
                >
                    Info <span class="bg-blue-600/30 px-1.5 rounded-full text-xs">{logStats.info + logStats.debug}</span>
                </button>
                <button 
                    class="px-3 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1 {selectedTab === 'warn' ? 'bg-yellow-600/20 text-yellow-400 shadow-sm' : 'text-slate-400 hover:text-yellow-400 hover:bg-slate-800'}" 
                    onclick={() => selectedTab = 'warn'}
                >
                    Warnings <span class="bg-yellow-600/30 px-1.5 rounded-full text-xs">{logStats.warn}</span>
                </button>
                <button 
                    class="px-3 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1 {selectedTab === 'error' ? 'bg-red-600/20 text-red-400 shadow-sm' : 'text-slate-400 hover:text-red-400 hover:bg-slate-800'}" 
                    onclick={() => selectedTab = 'error'}
                >
                    Errors <span class="bg-red-600/30 px-1.5 rounded-full text-xs">{logStats.error + logStats.fatal + logStats.panic}</span>
                </button>
            </div>

            <!-- Action Buttons -->
            <div class="flex gap-2">
                <button 
                    onclick={() => showStats = !showStats}
                    class="px-3 py-1.5 bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 rounded-md text-xs font-semibold transition-colors flex items-center gap-2"
                    title="Toggle Statistics"
                >
                    <BarChart3 class="w-3.5 h-3.5" />
                    Stats
                </button>
                <button 
                    onclick={copyToClipboard} 
                    disabled={loading || filteredLogs.length === 0}
                    class="px-3 py-1.5 bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 rounded-md text-xs font-semibold transition-colors disabled:opacity-50 flex items-center gap-2"
                    title="Copy to Clipboard (Ctrl+Shift+C)"
                >
                    <Copy class="w-3.5 h-3.5" />
                    Copy
                </button>
                <div class="relative group">
                    <button 
                        class="px-3 py-1.5 bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 rounded-md text-xs font-semibold transition-colors flex items-center gap-2"
                        title="Export Logs"
                    >
                        <Download class="w-3.5 h-3.5" />
                        Export
                    </button>
                    <div class="absolute right-0 mt-1 w-32 bg-slate-800 border border-slate-700 rounded-md shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-10">
                        <button 
                            onclick={() => exportLogs('json')}
                            class="block w-full text-left px-3 py-2 text-xs text-slate-300 hover:bg-slate-700 rounded-t-md"
                        >
                            Export as JSON
                        </button>
                        <button 
                            onclick={() => exportLogs('txt')}
                            class="block w-full text-left px-3 py-2 text-xs text-slate-300 hover:bg-slate-700 rounded-b-md"
                        >
                            Export as Text
                        </button>
                    </div>
                </div>
                <button 
                    onclick={requestClearLogs} 
                    disabled={loading}
                    class="px-3 py-1.5 bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 rounded-md text-xs font-semibold transition-colors disabled:opacity-50 flex items-center gap-2"
                >
                    <X class="w-3.5 h-3.5" />
                    Clear
                </button>
                <button 
                    onclick={fetchLogs} 
                    disabled={loading}
                    class="px-3 py-1.5 bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 rounded-md text-xs font-semibold transition-colors disabled:opacity-50 flex items-center gap-2"
                >
                    <RefreshCw class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" />
                    Refresh
                </button>
            </div>
        </div>

        <!-- Search and Controls Row -->
        <div class="flex flex-wrap gap-4 items-center">
            <!-- Enhanced Search Bar -->
            <div class="flex-1 min-w-64">
                <div class="relative">
                    <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-slate-400" />
                    <input
                        id="log-search"
                        type="text"
                        bind:value={searchTerm}
                        placeholder="Search logs... (Ctrl+F or /)"
                        class="w-full pl-10 pr-20 py-2 bg-slate-900 border border-slate-700 rounded-md text-sm text-slate-300 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <div class="absolute right-2 top-1/2 transform -translate-y-1/2 flex items-center gap-1">
                        {#if totalSearchResults > 0}
                            <span class="text-xs text-slate-400 bg-slate-800 px-2 py-1 rounded">
                                {currentSearchIndex + 1}/{totalSearchResults}
                            </span>
                        {/if}
                        <div class="flex gap-1">
                            <button
                                onclick={() => navigateSearch('prev')}
                                disabled={totalSearchResults === 0}
                                class="p-1 text-slate-400 hover:text-slate-200 disabled:opacity-50"
                                title="Previous (P)"
                            >
                                <ChevronUp class="w-3 h-3" />
                            </button>
                            <button
                                onclick={() => navigateSearch('next')}
                                disabled={totalSearchResults === 0}
                                class="p-1 text-slate-400 hover:text-slate-200 disabled:opacity-50"
                                title="Next (N)"
                            >
                                <ChevronDown class="w-3 h-3" />
                            </button>
                        </div>
                    </div>
                </div>
                
                <!-- Search Options -->
                <div class="flex gap-4 mt-2">
                    <label class="flex items-center gap-2 text-xs text-slate-400">
                        <input
                            type="checkbox"
                            bind:checked={caseSensitive}
                            class="rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                        />
                        Case Sensitive
                    </label>
                    <label class="flex items-center gap-2 text-xs text-slate-400">
                        <input
                            type="checkbox"
                            bind:checked={regexSearch}
                            class="rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                        />
                        Regex
                    </label>
                </div>
            </div>

            <!-- Auto Refresh Controls -->
            <div class="flex items-center gap-3">
                <label class="flex items-center gap-2 text-xs text-slate-400">
                    <input
                        type="checkbox"
                        bind:checked={autoRefresh}
                        class="rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                    />
                    Auto Refresh
                </label>
                {#if autoRefresh}
                    <select
                        bind:value={refreshInterval}
                        class="bg-slate-900 border border-slate-700 rounded-md text-xs text-slate-300 px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value={1000}>1s</option>
                        <option value={5000}>5s</option>
                        <option value={10000}>10s</option>
                        <option value={30000}>30s</option>
                    </select>
                {/if}
            </div>
        </div>
    </div>

    <!-- Statistics Panel -->
    {#if showStats}
        <div class="bg-slate-900/50 border-b border-slate-800 p-4">
            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4 text-xs">
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Total</div>
                    <div class="text-lg font-bold text-slate-200">{logStats.total}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Debug</div>
                    <div class="text-lg font-bold text-slate-400">{logStats.debug}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Info</div>
                    <div class="text-lg font-bold text-blue-400">{logStats.info}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Warnings</div>
                    <div class="text-lg font-bold text-yellow-400">{logStats.warn}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Errors</div>
                    <div class="text-lg font-bold text-red-400">{logStats.error}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Fatal</div>
                    <div class="text-lg font-bold text-red-500">{logStats.fatal}</div>
                </div>
                <div class="bg-slate-800 p-3 rounded-lg border border-slate-700">
                    <div class="text-slate-400 mb-1">Panic</div>
                    <div class="text-lg font-bold text-red-600">{logStats.panic}</div>
                </div>
            </div>
        </div>
    {/if}

    <!-- Logs Container -->
    <div class="flex-1 bg-slate-950 overflow-hidden relative">
        {#if loading && !logsRaw}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-500 gap-4 bg-slate-950/80 backdrop-blur-sm">
                <div class="w-10 h-10 border-3 border-slate-600 border-t-blue-500 rounded-full animate-spin"></div>
                <span class="text-base font-medium text-slate-400">Loading logs...</span>
            </div>
        {:else if error}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-red-400 p-4 text-center">
                <div class="w-10 h-10 bg-red-500/10 rounded-full flex items-center justify-center mb-3">
                    <X class="w-5 h-5" />
                </div>
                <div class="text-sm font-medium">{error}</div>
            </div>
        {:else if filteredLogs.length === 0}
            <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-600 gap-2">
                <div class="w-16 h-16 bg-slate-800 rounded-full flex items-center justify-center mb-3">
                    <Search class="w-8 h-8" />
                </div>
                <span class="text-sm font-medium">No logs found</span>
                {#if searchTerm}
                    <span class="text-xs text-slate-500">Try adjusting your search or filters</span>
                {:else}
                    <span class="text-xs text-slate-500">Logs will appear here when available</span>
                {/if}
            </div>
        {:else}
            <div 
                bind:this={logContainer} 
                class="h-full overflow-y-auto scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-slate-900/50"
                on:scroll={handleScroll}
            >
                <div class="space-y-0.5 p-2">
                    {#each filteredLogs as entry (entry.id)}
                        <div 
                            data-log-index={entry.id}
                            class="group flex items-start gap-3 p-1.5 rounded hover:bg-slate-900/80 transition-colors text-xs font-mono leading-tight border-b border-slate-900/50 {searchResults.includes(entry.id) ? 'bg-yellow-900/20' : ''}"
                        >
                            <!-- Time -->
                            <div class="text-slate-500 whitespace-nowrap shrink-0 w-16 text-right select-none font-mono">
                                {entry.time}
                            </div>
                            
                            <!-- Level Badge -->
                            <div class="shrink-0">
                                <span class={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold tracking-wider ${getLevelClass(entry.level)}`}>
                                    <span>{getLevelIcon(entry.level)}</span>
                                    {entry.level}
                                </span>
                            </div>

                            <!-- Message & Attributes -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-start gap-2">
                                    <div class="flex-1">
                                        <span class="text-slate-300 break-words">
                                            {@html highlightSearchTerm(entry.message)}
                                        </span>
                                        {#if entry.attributes}
                                            <span class="text-slate-500 ml-2 italic break-words">
                                                {@html highlightSearchTerm(entry.attributes)}
                                            </span>
                                        {/if}
                                    </div>
                                    {#if entry.attributes}
                                        <button
                                            onclick={() => toggleLogEntry(entry)}
                                            class="p-1 text-slate-500 hover:text-slate-300 transition-colors"
                                            title="Toggle details"
                                        >
                                            {#if entry.expanded}
                                                <ChevronUp class="w-3 h-3" />
                                            {:else}
                                                <ChevronDown class="w-3 h-3" />
                                            {/if}
                                        </button>
                                    {/if}
                                </div>
                                
                                <!-- Expanded Details -->
                                {#if entry.expanded && entry.attributes}
                                    <div class="mt-2 p-2 bg-slate-900/50 rounded border border-slate-800 text-xs">
                                        <div class="text-slate-400 mb-1">Full Attributes:</div>
                                        <div class="text-slate-300 font-mono break-all">
                                            {@html highlightSearchTerm(entry.attributes)}
                                        </div>
                                        <div class="text-slate-400 mt-2">
                                            <div class="flex items-center gap-2">
                                                <Clock class="w-3 h-3" />
                                                <span>Timestamp: {new Date(entry.timestamp).toISOString()}</span>
                                            </div>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
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

<style>
    :global(mark) {
        background-color: rgb(253 224 71);
        color: rgb(0 0 0);
        padding: 1px 4px;
        border-radius: 2px;
    }
    
    /* Custom scrollbar styling */
    .scrollbar-thin {
        scrollbar-width: thin;
    }
    
    .scrollbar-thumb-slate-700::-webkit-scrollbar-thumb {
        background-color: rgb(51 65 85);
        border-radius: 4px;
    }
    
    .scrollbar-track-slate-900\/50::-webkit-scrollbar-track {
        background-color: rgba(15 23 42 / 0.5);
        border-radius: 4px;
    }
</style>