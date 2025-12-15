<script lang="ts">
    import { onMount, tick, createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { cubicOut } from 'svelte/easing';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import { formatBytes } from '$lib/utils';
    import {
        Search, Download, Copy, RefreshCw, X,
        ChevronDown, ChevronUp, Clock, BarChart3,
        AlertTriangle, Info, Bug, XCircle
    } from 'lucide-svelte';

    export let spawnerId: number;
    export let isOpen = false;
    export let onClose: () => void = () => {};
    export let embedded = false;

    type LogLevel = 'DBG' | 'INF' | 'WRN' | 'ERR' | 'FTL' | 'PANIC';
    
    interface ParsedLogEntry {
        id: number;
        time: string;
        level: LogLevel;
        message: string;
        raw: any; // The full parsed JSON object
        originalLine: string;
        timestamp: number;
    }

    type TabId = 'all' | 'info' | 'warn' | 'error';
    interface TabDef {
        id: TabId;
        label: string;
        icon: any;
        color: string;
    }

    const dispatch = createEventDispatcher();

    let logsRaw = '';
    let parsedLogs: ParsedLogEntry[] = [];
    let filteredLogs: ParsedLogEntry[] = [];
    let loading = false;
    let error = '';
    let fileSize = 0;

    let selectedTab: TabId = 'all';
    let searchTerm = '';
    
    // Stats for tabs
    let stats = {
        all: 0,
        info: 0,
        warn: 0,
        error: 0
    };

    let autoRefresh = false;
    let refreshInterval = 5000;
    let refreshTimer: ReturnType<typeof setInterval> | null = null;

    let logContainer: HTMLElement;
    let shouldAutoScroll = true;

    let isConfirmOpen = false;
    
    async function handleClearLogs() {
        try {
            await fetch(`/api/spawners/${spawnerId}/logs`, { method: 'DELETE' });
            parsedLogs = [];
            filteredLogs = [];
            fileSize = 0;
            updateStats();
        } catch (e) {
            console.error(e);
        }
        isConfirmOpen = false;
    }

    let confirmAction = handleClearLogs;

    const tabs: TabDef[] = [
        { id: 'all', label: 'All', icon: BarChart3, color: 'text-slate-400' },
        { id: 'info', label: 'Info', icon: Info, color: 'text-blue-400' },
        { id: 'warn', label: 'Warnings', icon: AlertTriangle, color: 'text-yellow-400' },
        { id: 'error', label: 'Errors', icon: XCircle, color: 'text-red-400' },
    ];

    function parseLogLine(line: string, index: number): ParsedLogEntry {
        try {
            const json = JSON.parse(line);
            const date = new Date(json.time);
            
            // Map slog levels to our types if needed, though usually they match
            // slog default is DEBUG, INFO, WARN, ERROR
            let level: LogLevel = 'INF';
            const l = (json.level || '').toUpperCase();
            if (l === 'DEBUG') level = 'DBG';
            else if (l === 'INFO') level = 'INF';
            else if (l === 'WARN') level = 'WRN';
            else if (l === 'ERROR') level = 'ERR';
            else if (l === 'FATAL') level = 'FTL'; // Not standard slog but good to handle
            else if (l === 'PANIC') level = 'PANIC';

            return {
                id: index,
                time: date.toLocaleTimeString([], { hour12: false }) + '.' + date.getMilliseconds().toString().padStart(3, '0'),
                level: level,
                message: json.msg || json.message || '',
                raw: json,
                originalLine: line,
                timestamp: date.getTime()
            };
        } catch (e) {
            // Fallback for non-JSON lines
            return {
                id: index,
                time: '-',
                level: 'INF',
                message: line,
                raw: {},
                originalLine: line,
                timestamp: Date.now()
            };
        }
    }

    function updateStats() {
        const s = { all: 0, info: 0, warn: 0, error: 0 };
        parsedLogs.forEach(l => {
            s.all++;
            if (l.level === 'WRN') s.warn++;
            else if (['ERR', 'FTL', 'PANIC'].includes(l.level)) s.error++;
            else s.info++;
        });
        stats = s;
    }

    function filterLogs() {
        let out = [...parsedLogs];

        // 1. Tab Filter
        if (selectedTab !== 'all') {
            out = out.filter(l => {
                if (selectedTab === 'info') return ['INF', 'DBG'].includes(l.level);
                if (selectedTab === 'warn') return l.level === 'WRN';
                return ['ERR', 'FTL', 'PANIC'].includes(l.level);
            });
        }

        // 2. Search Filter
        if (searchTerm.trim()) {
            const lowerTerm = searchTerm.toLowerCase();
            out = out.filter(l => 
                l.message.toLowerCase().includes(lowerTerm) || 
                l.time.includes(lowerTerm) ||
                (l.raw.error && String(l.raw.error).toLowerCase().includes(lowerTerm))
            );
        }

        return out;
    }

    function getLevelClass(level: LogLevel) {
        switch (level) {
            case 'DBG': return 'text-slate-500';
            case 'INF': return 'text-blue-400';
            case 'WRN': return 'text-yellow-400';
            case 'ERR': return 'text-red-500 font-bold';
            case 'FTL': return 'text-purple-500 font-bold';
            case 'PANIC': return 'text-purple-600 font-bold bg-purple-950/30';
            default: return 'text-slate-400';
        }
    }

    async function fetchLogs() {
        if (!isOpen) return; // Don't fetch if closed
        loading = parsedLogs.length === 0; // Only show loading on initial fetch
        
        try {
            const r = await fetch(`/api/spawners/${spawnerId}/logs`);
            if (!r.ok) throw new Error('Failed to fetch logs');
            
            const j = await r.json();
            // Handle case where logs might be empty or null
            logsRaw = j.logs || '';
            fileSize = j.size || 0;
            
            parsedLogs = logsRaw
                .split('\n')
                .filter(line => line.trim().length > 0)
                .map(parseLogLine);
            
            updateStats();
            filteredLogs = filterLogs();
        } catch (e: any) {
            error = e.message;
        } finally {
            loading = false;
        }
    }

    function handleScroll() {
        if (!logContainer) return;
        const { scrollTop, scrollHeight, clientHeight } = logContainer;
        // Check if user is near bottom (within 50px)
        shouldAutoScroll = scrollHeight - scrollTop - clientHeight < 50;
    }

    async function scrollBottom() {
        if (shouldAutoScroll && logContainer) {
            await tick();
            logContainer.scrollTop = logContainer.scrollHeight;
        }
    }

    // Reactivity
    $: if (selectedTab || searchTerm) {
        filteredLogs = filterLogs();
        // If searching, maybe disable autoscroll? For now keep it simple.
    }
    
    $: if (filteredLogs.length > 0) {
        scrollBottom();
    }

    // Auto Refresh Logic
    $: if (autoRefresh && isOpen) {
        if (!refreshTimer) {
            refreshTimer = setInterval(fetchLogs, refreshInterval);
        }
    } else {
        if (refreshTimer) {
            clearInterval(refreshTimer);
            refreshTimer = null;
        }
    }

    onMount(() => {
        if (isOpen) fetchLogs();
        return () => {
            if (refreshTimer) clearInterval(refreshTimer);
        };
    });
    
    // Watch for open prop changes to trigger fetch
    $: if (isOpen && parsedLogs.length === 0) {
        fetchLogs();
    }
</script>

{#if isOpen}
    {#if !embedded}
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" transition:fade={{ duration: 200 }}>
            <!-- Backdrop click to close -->
            <div 
                class="absolute inset-0" 
                on:click={onClose}
                on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && onClose()}
                role="button"
                tabindex="0"
                aria-label="Close"
            ></div>

            <div
                class="relative w-full max-w-6xl h-[85vh] bg-slate-900 rounded-xl border border-slate-700 shadow-2xl flex flex-col overflow-hidden"
                transition:scale={{ start: 0.95, duration: 200, easing: cubicOut }}
            >
                <div class="contents">
                    <!-- Shared Content -->
                    {@render content()}
                </div>
            </div>
        </div>
    {:else}
        <div class="h-full flex flex-col bg-slate-900 overflow-hidden">
             {@render content()}
        </div>
    {/if}
{/if}

{#snippet content()}
        <!-- Header -->
        <div class="px-4 py-3 border-b border-slate-700 flex justify-between items-center bg-slate-800/50">
            <div class="flex items-center gap-3">
                <h2 class="text-white font-semibold text-lg flex items-center gap-2">
                    <span class="text-slate-400"># {spawnerId}</span> Console Logs
                    {#if fileSize > 0}
                        <span class="text-xs text-slate-500 font-mono ml-2 border border-slate-700 rounded px-1.5 py-0.5 bg-slate-900/50">{formatBytes(fileSize)}</span>
                    {/if}
                </h2>
                {#if loading && parsedLogs.length > 0}
                    <RefreshCw class="w-4 h-4 text-slate-400 animate-spin" />
                {/if}
            </div>
            
            <div class="flex items-center gap-2">
                 <button 
                    class="p-2 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-colors"
                    title="Refresh Now"
                    on:click={fetchLogs}
                >
                    <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
                </button>
                
                <button 
                    class="p-2 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-colors"
                    title="Clear Logs"
                    on:click={() => isConfirmOpen = true}
                >
                    <XCircle class="w-4 h-4" />
                </button>

                {#if !embedded}
                    <div class="h-6 w-px bg-slate-700 mx-2"></div>

                    <button 
                        on:click={onClose}
                        class="p-2 hover:bg-red-500/20 hover:text-red-400 rounded-lg text-slate-400 transition-colors"
                    >
                        <X class="w-5 h-5" />
                    </button>
                {/if}
            </div>
        </div>

        <!-- Toolbar: Tabs & Search -->
        <div class="px-4 py-3 bg-slate-900 border-b border-slate-800 flex flex-col sm:flex-row gap-4 justify-between items-center">
            
            <!-- Tabs -->
            <div class="flex p-1 bg-slate-950 rounded-lg border border-slate-800">
                {#each tabs as tab}
                    <button
                        class="px-3 py-1.5 rounded-md text-sm font-medium transition-all flex items-center gap-2
                        {selectedTab === tab.id ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-500 hover:text-slate-300'}"
                        on:click={() => selectedTab = tab.id}
                    >
                        <svelte:component this={tab.icon} class="w-3.5 h-3.5" />
                        {tab.label}
                        <span class="ml-1 text-xs opacity-60 bg-slate-900/50 px-1.5 rounded-full">
                            {stats[tab.id]}
                        </span>
                    </button>
                {/each}
            </div>

            <!-- Search & Auto-Refresh -->
            <div class="flex items-center gap-3 w-full sm:w-auto">
                <div class="relative w-full sm:w-64">
                    <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                    <input
                        type="text"
                        placeholder="Search logs..."
                        bind:value={searchTerm}
                        class="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
                    />
                </div>
                
                <label class="flex items-center gap-2 text-xs text-slate-400 cursor-pointer select-none">
                    <input type="checkbox" bind:checked={autoRefresh} class="rounded bg-slate-800 border-slate-700 text-blue-500 focus:ring-0 focus:ring-offset-0" />
                    <span>Auto-scroll</span>
                </label>
            </div>
        </div>

        <!-- Logs Area -->
        <div class="flex-1 relative bg-slate-950 min-h-0">
            {#if loading && parsedLogs.length === 0}
                <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-500 gap-3">
                    <RefreshCw class="w-8 h-8 animate-spin opacity-50" />
                    <span class="text-sm">Loading logs...</span>
                </div>
            {:else if error}
                <div class="absolute inset-0 flex flex-col items-center justify-center text-red-400 gap-2">
                    <AlertTriangle class="w-8 h-8 opacity-50" />
                    <span class="text-sm">{error}</span>
                    <button class="text-xs underline hover:text-red-300" on:click={fetchLogs}>Try Again</button>
                </div>
            {:else if filteredLogs.length === 0}
                <div class="absolute inset-0 flex flex-col items-center justify-center text-slate-600 gap-2">
                    <Search class="w-8 h-8 opacity-20" />
                    <span class="text-sm">No logs found</span>
                </div>
            {:else}
                <div
                    bind:this={logContainer}
                    on:scroll={handleScroll}
                    class="absolute inset-0 overflow-y-auto overflow-x-hidden p-2 font-mono text-xs space-y-0.5 custom-scrollbar"
                >
                    {#each filteredLogs as l (l.id)}
                        <div class="flex items-start gap-3 hover:bg-slate-900/50 px-2 py-1 rounded select-text group transition-colors">
                            <!-- Time -->
                            <span class="text-slate-600 shrink-0 w-24 tabular-nums select-none">{l.time}</span>
                            
                            <!-- Level -->
                            <span class="shrink-0 w-12 font-bold select-none {getLevelClass(l.level)}">
                                {l.level}
                            </span>
                            
                            <!-- Message -->
                            <div class="flex-1 min-w-0 break-words text-slate-300">
                                <span>{l.message}</span>
                                
                                <!-- Structured Context / Error Cause -->
                                {#if l.raw && Object.keys(l.raw).length > 3}
                                    <div class="mt-1 ml-2 text-slate-500 text-[10px] space-y-1 border-l-2 border-slate-800 pl-2 opacity-80 group-hover:opacity-100 transition-opacity">
                                        {#each Object.entries(l.raw) as [k, v]}
                                            {#if !['time', 'level', 'msg', 'message'].includes(k)}
                                                <div class="flex gap-2">
                                                    <span class="text-slate-600">{k}:</span>
                                                    <span class="text-slate-400 font-mono whitespace-pre-wrap">{JSON.stringify(v)}</span>
                                                </div>
                                            {/if}
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/each}
                    
                    {#if shouldAutoScroll}
                         <div class="h-4"></div> <!-- Spacer at bottom -->
                    {/if}
                </div>
            {/if}
        </div>
{/snippet}

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title="Clear Logs"
    message="Are you sure you want to clear all logs? This action cannot be undone."
    onConfirm={confirmAction}
/>

<style>
    /* Custom Scrollbar for the log container */
    .custom-scrollbar::-webkit-scrollbar {
        width: 10px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: #0f172a; /* slate-950 */
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: #334155; /* slate-700 */
        border-radius: 5px;
        border: 2px solid #0f172a;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: #475569; /* slate-600 */
    }
</style>
