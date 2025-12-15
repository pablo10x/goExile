<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import type { ResourceMetricsPanelProps } from '$lib/types/resource-metrics';
    import { useResourceMetrics } from '$lib/composables/useResourceMetrics';
    import ResourceStatsCard from '$lib/components/ResourceStatsCard.svelte';
    import ResourceHistoryChart from '$lib/components/ResourceHistoryChart.svelte';
    import { RefreshCw, AlertTriangle, Cpu, HardDrive, Database } from 'lucide-svelte';
    
    let { 
        spawnerId, 
        instanceId, 
        memTotal = 0,
        diskTotal = 0,
        height = 250, 
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

<div class="space-y-6">
    <!-- Header -->
    {#if showTitle}
        <div class="flex items-center justify-between">
            <div>
                <h3 class="text-lg font-bold text-slate-100 flex items-center gap-2">
                    Resource Usage
                    {#if $loading}
                        <div class="w-3 h-3 rounded-full border-2 border-slate-500 border-t-blue-500 animate-spin"></div>
                    {/if}
                </h3>
                <p class="text-xs text-slate-500 font-mono mt-0.5">
                    Instance {instanceId}
                </p>
            </div>
            
            <div class="flex items-center gap-2">
                <div class="flex bg-slate-900/50 rounded-lg p-0.5 border border-slate-800">
                    {#each timeRanges as range}
                        <button 
                            class="px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider rounded transition-all {timeRange === range ? 'bg-slate-700 text-white shadow-sm' : 'text-slate-500 hover:text-slate-300'}"
                            onclick={() => timeRange = range}
                        >
                            {range}
                        </button>
                    {/each}
                </div>
                <button 
                    onclick={refresh}
                    disabled={$loading}
                    class="p-1.5 text-slate-500 hover:text-white hover:bg-slate-800 rounded-lg transition-colors disabled:opacity-50"
                    title="Refresh"
                >
                    <RefreshCw class="w-4 h-4 {$loading ? 'animate-spin' : ''}" />
                </button>
            </div>
        </div>
    {/if}
    
    <!-- Error State -->
    {#if $error}
        <div class="bg-red-500/10 border border-red-500/20 rounded-lg p-4 flex items-start gap-3">
            <AlertTriangle class="w-5 h-5 text-red-400 shrink-0" />
            <div class="text-sm text-red-300">
                <span class="font-bold">Error fetching metrics:</span> {$error}
            </div>
        </div>
    {/if}
    
    {#if !$error}
        <!-- Stats Cards Grid -->
        {#if $stats}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <ResourceStatsCard
                    title="CPU"
                    current={$stats.cpu_percent}
                    peak={$peakStats.peakCpu}
                    unit="%"
                    icon="⚡"
                    color="orange"
                    trend={$stats.cpu_percent > $peakStats.peakCpu * 0.9 ? 'up' : 'stable'}
                />
                
                <ResourceStatsCard
                    title="Memory"
                    current={$stats.memory_usage / 1024 / 1024}
                    peak={$peakStats.peakMemory / 1024 / 1024}
                    unit="MB"
                    icon="🧠"
                    color="blue"
                    trend={$stats.memory_usage > $peakStats.peakMemory * 0.9 ? 'up' : 'stable'}
                />
                
                <ResourceStatsCard
                    title="Disk"
                    current={$stats.disk_usage / 1024 / 1024 / 1024}
                    peak={$peakStats.peakDisk / 1024 / 1024 / 1024}
                    unit="GB"
                    icon="💾"
                    color="green"
                    trend={$stats.disk_usage > $peakStats.peakDisk * 0.9 ? 'up' : 'stable'}
                />
            </div>
        {:else if $loading}
             <!-- Skeleton Loading -->
             <div class="grid grid-cols-1 md:grid-cols-3 gap-4 animate-pulse">
                {#each [1, 2, 3] as _}
                    <div class="h-32 bg-slate-800/50 rounded-xl border border-slate-700/50"></div>
                {/each}
             </div>
        {/if}
        
        <!-- Resource History Chart -->
        {#if !compact}
            <div class="bg-slate-900/30 rounded-xl border border-slate-800 p-1 relative overflow-hidden group">
                <div class="absolute top-3 left-4 text-xs font-bold text-slate-500 uppercase tracking-widest z-10 pointer-events-none">
                    Performance Timeline
                </div>
                <ResourceHistoryChart 
                    data={filteredHistory}
                    height={height}
                />
            </div>
        {/if}
    {/if}
</div>
