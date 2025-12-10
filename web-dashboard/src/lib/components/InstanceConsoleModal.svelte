<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { formatBytes } from '$lib/utils';

    export let isOpen: boolean = false;
    export let spawnerId: number | null = null;
    export let instanceId: string | null = null;
    export let onClose: () => void;

    let logs: string[] = [];
    let logContainer: HTMLElement;
    let eventSource: EventSource | null = null;
    
    let stats = {
        cpu_percent: 0,
        memory_usage: 0,
        disk_usage: 0
    };
    let statsInterval: ReturnType<typeof setInterval> | null = null;

    $: if (isOpen && spawnerId && instanceId) {
        startStreaming();
    } else {
        stopStreaming();
    }

    function startStreaming() {
        stopStreaming(); // Ensure clean start
        logs = ['Connecting to console stream...'];

        // Connect to SSE for Logs
        eventSource = new EventSource(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`);
        
        eventSource.onopen = () => {
            logs = [...logs, 'Connected. Tailing logs...'];
        };

        eventSource.addEventListener('log', (event) => {
            try {
                // The server sends the raw line
                const line = JSON.parse(event.data); // Or just event.data if it's plain text, but Gin SSEEvent marshals to string? 
                // Gin's c.SSEvent("log", line) sends: event: log\ndata: <line>\n\n
                // If <line> is a string, browser receives it as string.
                if (line) {
                    logs = [...logs, line.trimEnd()];
                    scrollToBottom();
                }
            } catch (e) {
                 // Fallback if it's just raw text
                 logs = [...logs, event.data.trimEnd()];
                 scrollToBottom();
            }
        });

        eventSource.onerror = () => {
            logs = [...logs, 'Connection lost. Reconnecting...'];
            eventSource?.close();
            // Optional: retry logic, but EventSource usually retries auto. 
            // However, we might want to manually handle it if it fails permanently.
        };

        // Start Stats Polling
        fetchStats();
        statsInterval = setInterval(fetchStats, 2000);
    }

    function stopStreaming() {
        if (eventSource) {
            eventSource.close();
            eventSource = null;
        }
        if (statsInterval) {
            clearInterval(statsInterval);
            statsInterval = null;
        }
        logs = [];
    }

    async function fetchStats() {
        if (!spawnerId || !instanceId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
            if (res.ok) {
                stats = await res.json();
            }
        } catch (e) {
            console.error('Failed to fetch instance stats', e);
        }
    }

    function scrollToBottom() {
        if (logContainer) {
            setTimeout(() => {
                logContainer.scrollTop = logContainer.scrollHeight;
            }, 0);
        }
    }

    function close() {
        stopStreaming();
        onClose();
    }

    async function clearLogs() {
        if (!confirm('Clear console logs?')) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`, { method: 'DELETE' });
            if (res.ok) {
                logs = [];
            }
        } catch (e) {
            console.error('Failed to clear logs', e);
        }
    }

    onDestroy(() => {
        stopStreaming();
    });
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop -->
        <div 
            class="absolute inset-0 bg-slate-950/90 backdrop-blur-sm"
            on:click={close}
            role="button"
            tabindex="0"
            aria-label="Close console"
        ></div>

        <!-- Modal Window -->
        <div 
            class="relative w-full max-w-4xl h-[80vh] flex flex-col bg-slate-900 border border-slate-700 rounded-xl shadow-2xl overflow-hidden"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <!-- Header -->
            <div class="px-6 py-4 bg-slate-800 border-b border-slate-700 flex justify-between items-center shrink-0">
                <div class="flex items-center gap-4">
                    <h3 class="text-lg font-bold text-slate-100 flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>
                        Instance Console
                        <span class="px-2 py-0.5 rounded-full bg-slate-700 text-xs font-mono text-slate-300 ml-2">{instanceId}</span>
                    </h3>
                </div>
                <div class="flex items-center gap-4">
                    <button 
                        on:click={clearLogs}
                        class="text-xs font-semibold text-slate-400 hover:text-red-400 transition-colors uppercase tracking-wider"
                    >
                        Clear Output
                    </button>
                    <button 
                        on:click={close}
                        class="text-slate-400 hover:text-white transition-colors"
                        aria-label="Close"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Stats Bar -->
            <div class="px-6 py-2 bg-slate-950/50 border-b border-slate-800 flex gap-6 text-xs font-mono text-slate-400 shrink-0">
                <div class="flex items-center gap-2">
                    <span class="text-slate-500 uppercase font-bold">CPU:</span>
                    <span class={stats.cpu_percent > 80 ? 'text-red-400' : 'text-slate-200'}>{stats.cpu_percent.toFixed(1)}%</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-slate-500 uppercase font-bold">MEM:</span>
                    <span class="text-slate-200">{formatBytes(stats.memory_usage)}</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-slate-500 uppercase font-bold">DISK:</span>
                    <span class="text-slate-200">{formatBytes(stats.disk_usage)}</span>
                </div>
            </div>

            <!-- Console Output -->
            <div class="flex-1 bg-black p-4 overflow-hidden relative group">
                <div 
                    bind:this={logContainer}
                    class="h-full overflow-y-auto font-mono text-xs text-slate-300 space-y-1 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent pr-2"
                >
                    {#each logs as line, i}
                        <div class="break-all whitespace-pre-wrap hover:bg-slate-900/50 px-1 rounded">{line}</div>
                    {/each}
                    {#if logs.length === 0}
                        <div class="text-slate-600 italic">Waiting for output...</div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
{/if}