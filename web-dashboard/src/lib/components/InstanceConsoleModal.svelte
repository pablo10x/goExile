<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { formatBytes, formatUptime } from '$lib/utils';
    import Terminal from './Terminal.svelte';
    import ResourceHistoryChart from './ResourceHistoryChart.svelte';

    export let isOpen: boolean = false;
    export let spawnerId: number | null = null;
    export let instanceId: string | null = null;
    export let onClose: () => void;
    export let memTotal: number = 0;

    let logs: string[] = [];
    let eventSource: EventSource | null = null;
    
    let stats = {
        cpu_percent: 0,
        memory_usage: 0,
        disk_usage: 0,
        status: 'Unknown',
        uptime: 0
    };
    let history: any[] = [];
    let activeTab: 'console' | 'metrics' = 'console';

    let statsInterval: ReturnType<typeof setInterval> | null = null;
    let historyInterval: ReturnType<typeof setInterval> | null = null;
    
    // Provisioning State
    let isProvisioning = false;
    let provisioningStep = 0;
    const provisioningSteps = [
        'Allocating resources...',
        'Downloading game files...',
        'Configuring environment...',
        'Starting process...'
    ];

    $: if (isOpen && spawnerId && instanceId) {
        startStreaming();
        // Reset provisioning animation if just opened
        provisioningStep = 0;
    } else {
        stopStreaming();
        activeTab = 'console'; // Reset tab
    }

    // Provisioning simulation logic
    $: if (stats.status === 'Provisioning') {
         isProvisioning = true;
    } else {
         isProvisioning = false;
    }

    let provTimer: any;
    $: if (isProvisioning) {
        clearInterval(provTimer);
        provTimer = setInterval(() => {
            if (provisioningStep < provisioningSteps.length - 1) provisioningStep++;
        }, 2000);
    } else {
        clearInterval(provTimer);
    }

    // History Polling
    $: if (isOpen && activeTab === 'metrics') {
        fetchHistory();
        clearInterval(historyInterval!);
        historyInterval = setInterval(fetchHistory, 60000);
    } else {
        clearInterval(historyInterval!);
    }

    function startStreaming() {
        stopStreaming(); 
        logs = [];
        provisioningStep = 0;

        if (!spawnerId || !instanceId) return;

        // Connect to SSE
        eventSource = new EventSource(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`);
        
        eventSource.onopen = () => {
            logs = [...logs, `>>> Connected to ${instanceId} log stream.`];
        };

        eventSource.addEventListener('log', (event) => {
            try {
                const line = JSON.parse(event.data);
                if (line) logs = [...logs, line.trimEnd()];
            } catch (e) {
                 logs = [...logs, event.data.trimEnd()];
            }
        });

        eventSource.onerror = () => {
            eventSource?.close();
        };

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
        if (historyInterval) {
            clearInterval(historyInterval);
            historyInterval = null;
        }
        logs = [];
    }

    async function fetchStats() {
        if (!spawnerId || !instanceId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
            if (res.ok) {
                const data = await res.json();
                stats = { ...stats, ...data };

                // Realtime injection for short uptime instances
                if (activeTab === 'metrics' && stats.uptime < 43200 && memTotal > 0) {
                    const nowStr = new Date().toISOString();
                    const memPct = (stats.memory_usage / memTotal) * 100;
                    history = [...history, {
                        timestamp: nowStr,
                        cpu: stats.cpu_percent,
                        memory_percent: memPct
                    }];
                    // Keep history tidy in frontend to avoid memory leak if left open for days
                    if (history.length > 2000) history.shift();
                }
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function fetchHistory() {
        if (!spawnerId || !instanceId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats/history`);
            if (res.ok) {
                const data = await res.json();
                history = data.history || [];
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function handleAction(action: string) {
        if (!confirm(`Are you sure you want to ${action.toUpperCase()} this instance?`)) return;
        try {
            await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/${action}`, { method: 'POST' });
        } catch (e) {
            alert(`Failed: ${e}`);
        }
    }

    function close() {
        stopStreaming();
        onClose();
    }
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop -->
        <div 
            class="absolute inset-0 bg-slate-950/90 backdrop-blur-md"
            on:click={close}
            role="button"
            tabindex="0"
            aria-label="Close console"
        ></div>

        <!-- Modal Window -->
        <div 
            class="relative w-full max-w-6xl h-[85vh] flex bg-slate-900 border border-slate-700 rounded-xl shadow-2xl overflow-hidden"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <!-- Sidebar (Stats & Controls) -->
            <div class="w-72 bg-slate-950/50 border-r border-slate-800 flex flex-col shrink-0">
                <div class="p-6 border-b border-slate-800">
                    <h3 class="text-lg font-bold text-slate-100 break-all">{instanceId}</h3>
                    <div class="flex items-center gap-2 mt-2">
                        <span class={`w-2.5 h-2.5 rounded-full ${stats.status === 'Running' ? 'bg-emerald-500 animate-pulse' : (stats.status === 'Provisioning' ? 'bg-blue-500 animate-bounce' : 'bg-red-500')}`}></span>
                        <span class="text-sm font-mono text-slate-400">{stats.status || 'Unknown'}</span>
                    </div>
                </div>

                <div class="p-6 space-y-8 flex-1 overflow-y-auto">
                    <!-- Uptime -->
                    <div>
                        <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1">Uptime</div>
                        <div class="text-3xl font-mono text-slate-200 tracking-tight">{formatUptime((stats.uptime || 0) * 1000)}</div>
                    </div>

                    <!-- Resources -->
                    <div class="space-y-5">
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>CPU Usage</span>
                                <span>{stats.cpu_percent.toFixed(1)}%</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-blue-500 transition-all duration-500 ease-out" style="width: {stats.cpu_percent}%"></div>
                            </div>
                        </div>
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>Memory Usage</span>
                                <span>{formatBytes(stats.memory_usage)}</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-purple-500 transition-all duration-500 ease-out" style="width: {Math.min(100, (stats.memory_usage / (1024*1024*1024*4))*100)}%"></div> 
                            </div>
                        </div>
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>Disk Space</span>
                                <span>{formatBytes(stats.disk_usage)}</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-orange-500 transition-all duration-500 ease-out" style="width: 10%"></div> 
                            </div>
                        </div>
                    </div>

                    <!-- Provisioning Steps (Only if Provisioning) -->
                    {#if isProvisioning}
                        <div class="pt-6 border-t border-slate-800 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            <div class="text-[10px] font-bold text-blue-400 uppercase tracking-wider mb-4 flex items-center gap-2">
                                <div class="w-2 h-2 rounded-full border border-blue-400 border-t-transparent animate-spin"></div>
                                Provisioning
                            </div>
                            <div class="space-y-4">
                                {#each provisioningSteps as step, i}
                                    <div class="flex items-center gap-3 text-xs transition-all duration-300">
                                        {#if i < provisioningStep}
                                            <div class="w-5 h-5 rounded-full bg-emerald-500/20 flex items-center justify-center text-emerald-500 border border-emerald-500/30">✓</div>
                                            <span class="text-slate-500 line-through decoration-slate-700">{step}</span>
                                        {:else if i === provisioningStep}
                                            <div class="w-5 h-5 rounded-full border-2 border-blue-500 border-t-transparent animate-spin"></div>
                                            <span class="text-blue-200 font-semibold">{step}</span>
                                        {:else}
                                            <div class="w-5 h-5 rounded-full border border-slate-800"></div>
                                            <span class="text-slate-700">{step}</span>
                                        {/if}
                                    </div>
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Actions -->
                <div class="p-4 border-t border-slate-800 grid grid-cols-2 gap-3 bg-slate-900/50">
                    <button 
                        on:click={() => handleAction('restart')}
                        disabled={stats.status !== 'Running'}
                        class="px-4 py-2.5 bg-slate-800 hover:bg-blue-600/20 hover:text-blue-400 text-slate-400 rounded-lg text-xs font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-700 hover:border-blue-500/30"
                    >
                        Restart
                    </button>
                    <button 
                        on:click={() => handleAction('stop')}
                        disabled={stats.status !== 'Running'}
                        class="px-4 py-2.5 bg-slate-800 hover:bg-red-600/20 hover:text-red-400 text-slate-400 rounded-lg text-xs font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-700 hover:border-red-500/30"
                    >
                        Stop
                    </button>
                </div>
            </div>

            <!-- Main Area -->
            <div class="flex-1 flex flex-col min-w-0 bg-black relative">
                <!-- Tabs -->
                <div class="flex border-b border-slate-800 bg-slate-900 shrink-0">
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='console' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        on:click={() => activeTab = 'console'}
                    >
                        Console
                    </button>
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='metrics' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        on:click={() => activeTab = 'metrics'}
                    >
                        Metrics
                    </button>
                </div>

                <div class="flex-1 relative overflow-hidden">
                    {#if activeTab === 'console'}
                        <div class="absolute inset-0 p-0">
                            <!-- Live Indicator -->
                            <div class="absolute top-4 right-4 z-10">
                                <div class="px-3 py-1 rounded-full bg-slate-800/80 backdrop-blur border border-slate-700 text-[10px] font-mono text-slate-400 flex items-center gap-2">
                                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                                    Live
                                </div>
                            </div>
                            <Terminal logs={logs} title={`root@${instanceId}:~`} />
                        </div>
                    {:else}
                        <div class="p-8 h-full overflow-y-auto bg-slate-950/50">
                            <div class="mb-8">
                                <h3 class="text-lg font-bold text-slate-200 mb-2">Resource Usage (24h)</h3>
                                <p class="text-sm text-slate-500 mb-6">Historical CPU and Memory utilization for this instance.</p>
                                
                                <div class="h-72 bg-slate-900/50 border border-slate-800 rounded-xl p-4 shadow-inner relative">
                                    {#if history.length > 0}
                                        <ResourceHistoryChart data={history} height={280} />
                                    {:else}
                                        <div class="absolute inset-0 flex items-center justify-center text-slate-600">
                                            Collecting data...
                                        </div>
                                    {/if}
                                </div>
                            </div>
                            
                            <div class="grid grid-cols-2 gap-4">
                                <div class="bg-slate-900 border border-slate-800 p-4 rounded-lg">
                                    <div class="text-xs text-slate-500 uppercase font-bold mb-1">Peak CPU</div>
                                    <div class="text-xl text-red-400 font-mono">
                                        {Math.max(...history.map(h => h.cpu), 0).toFixed(1)}%
                                    </div>
                                </div>
                                <div class="bg-slate-900 border border-slate-800 p-4 rounded-lg">
                                    <div class="text-xs text-slate-500 uppercase font-bold mb-1">Peak Memory</div>
                                    <div class="text-xl text-blue-400 font-mono">
                                        {Math.max(...history.map(h => h.memory_percent), 0).toFixed(1)}%
                                    </div>
                                </div>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
            
            <!-- Close Button (Absolute) -->
            <button 
                on:click={close}
                class="absolute top-3 right-3 p-2 text-slate-500 hover:text-white transition-colors z-20"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
        </div>
    </div>
{/if}