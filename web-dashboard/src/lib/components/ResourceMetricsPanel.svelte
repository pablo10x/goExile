<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import type { ResourceMetricsPanelProps } from '$lib/types/resource-metrics';
    import { useResourceMetrics } from '$lib/composables/useResourceMetrics';
    import ResourceStatsCard from '$lib/components/ResourceStatsCard.svelte';
    import ResourceProgressBar from '$lib/components/ResourceProgressBar.svelte';
    import ResourceHistoryChart from '$lib/components/ResourceHistoryChart.svelte';
    
    let { 
        spawnerId, 
        instanceId, 
        memTotal = 0,
        diskTotal = 0,
        height = 300, 
        compact = false, 
        showTitle = true, 
        autoRefresh = true 
    }: ResourceMetricsPanelProps = $props();
    
    const {
        stats,
        history,
        peakStats,
        memTotal: fetchedMemTotal,
        loading,
        error,
        startPolling,
        stopPolling,
        refresh
    } = useResourceMetrics(spawnerId, instanceId);

    // Use passed total or fetched total
    let effectiveMemTotal = $derived(memTotal || $fetchedMemTotal || 0);
    
    let timeRange = $state('24h');
    const timeRanges = ['1h', '5h', '12h', '24h'];
    
    let filteredHistory = $derived($history.filter(point => {
        if (timeRange === '24h') return true;
        const pointTime = new Date(point.timestamp).getTime();
        const now = Date.now();
        const hours = parseInt(timeRange);
        const cutoff = now - (hours * 60 * 60 * 1000);
        return pointTime >= cutoff;
    }));
    
    onMount(() => {
        if (autoRefresh) {
            startPolling();
        }
    });
    
    onDestroy(() => {
        stopPolling();
    });
</script>

<div class="bg-slate-800/50 backdrop-blur-md border border-slate-700/50 rounded-xl p-6">
    <!-- Header -->
    {#if showTitle}
        <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
                <div class="p-2 bg-slate-900/50 rounded-lg border border-slate-700/50">
                    <div class="text-xl">📊</div>
                </div>
                <div>
                    <h3 class="text-lg font-semibold text-slate-200">Resource Usage</h3>
                    <p class="text-sm text-slate-400">
                        Instance #{instanceId} • Spawner #{spawnerId}
                    </p>
                </div>
            </div>
            
            <div class="flex items-center gap-2">
                <button 
                    onclick={refresh}
                    disabled={$loading}
                    class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    title="Refresh"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M23 4v6h-6"></path>
                        <path d="M1 20v6h6"></path>
                        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
                    </svg>
                </button>
            </div>
        </div>
    {/if}
    
    <!-- Error State -->
    {#if $error}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-300 mb-6">
            <div class="flex items-center gap-3">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <span>{$error}</span>
            </div>
        </div>
    {/if}
    
    <!-- Loading State -->
    {#if $loading}
        <div class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
    {:else if !$error}
        <!-- Stats Cards -->
        {#if $stats}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <ResourceStatsCard
                    title="CPU Usage"
                    current={$stats.cpu_percent}
                    peak={$peakStats.peakCpu}
                    unit="%"
                    icon="🖥️"
                    color="orange"
                    trend={$stats.cpu_percent > $peakStats.peakCpu * 0.9 ? 'up' : 'stable'}
                />
                
                <ResourceStatsCard
                    title="Memory Usage"
                    current={$stats.memory_usage / 1024 / 1024}
                    peak={$peakStats.peakMemory / 1024 / 1024}
                    unit="MB"
                    icon="💾"
                    color="blue"
                    trend={$stats.memory_usage > $peakStats.peakMemory * 0.9 ? 'up' : 'stable'}
                />
                
                <ResourceStatsCard
                    title="Disk Usage"
                    current={$stats.disk_usage / 1024 / 1024 / 1024}
                    peak={$peakStats.peakDisk / 1024 / 1024 / 1024}
                    unit="GB"
                    icon="💿"
                    color="green"
                    trend={$stats.disk_usage > $peakStats.peakDisk * 0.9 ? 'up' : 'stable'}
                />
            </div>
        {/if}
        
        <!-- Progress Bars -->
        {#if $stats}
            <div class="space-y-4 mb-6">
                <ResourceProgressBar
                    label="CPU Usage"
                    value={$stats.cpu_percent}
                    max={100}
                    color="orange"
                    showThreshold={true}
                    threshold={80}
                />
                
                <ResourceProgressBar
                    label="Memory Usage"
                    value={$stats.memory_usage}
                    max={effectiveMemTotal || $stats.memory_usage * 1.2} 
                    color="blue"
                    showThreshold={true}
                    threshold={85}
                />
                
                <ResourceProgressBar
                    label="Disk Usage"
                    value={$stats.disk_usage}
                    max={diskTotal || $stats.disk_usage * 1.2}
                    color="green"
                    showThreshold={true}
                    threshold={90}
                />
            </div>
        {/if}
        
        <!-- Resource History Chart -->
        {#if !compact}
            <div class="mb-6">
                <div class="flex items-center justify-between mb-4">
                    <h4 class="text-lg font-semibold text-slate-200">Resource History ({timeRange})</h4>
                    <div class="flex bg-slate-900/50 rounded-lg p-1 border border-slate-700/50">
                        {#each timeRanges as range}
                            <button 
                                class="px-3 py-1 text-xs font-medium rounded transition-all {timeRange === range ? 'bg-slate-700 text-white shadow-sm' : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800/50'}"
                                onclick={() => timeRange = range}
                            >
                                {range}
                            </button>
                        {/each}
                    </div>
                </div>
                <div class="bg-slate-900/50 rounded-lg p-4">
                    <ResourceHistoryChart 
                        data={filteredHistory}
                        height={height}
                    />
                </div>
            </div>
        {/if}
    {/if}
</div>