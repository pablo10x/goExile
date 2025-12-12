import { writable, type Writable } from 'svelte/store';
import type { ResourceStats, ResourceHistory, PeakResourceStats, ResourceMetricsState } from '$lib/types/resource-metrics';

export function useResourceMetrics(spawnerId: number, instanceId: number | string) {
    const stats = writable<ResourceStats | null>(null);
    const history = writable<ResourceHistory[]>([]);
    const peakStats = writable<PeakResourceStats>({ peakCpu: 0, peakMemory: 0, peakDisk: 0 });
    const memTotalStore = writable<number>(0);
    const loading = writable<boolean>(false);
    const error = writable<string | null>(null);

    let statsInterval: number | null = null;
    let historyInterval: number | null = null;
    let memTotal = 0;

    async function fetchStats() {
        try {
            const response = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
            if (!response.ok) throw new Error(`Failed to fetch stats: ${response.statusText}`);
            
            const data = await response.json();
            stats.set(data);
            
            // Calculate peak stats
            peakStats.update(peak => {
                const newPeak = { ...peak };
                if (data.cpu_percent > newPeak.peakCpu) newPeak.peakCpu = data.cpu_percent;
                if (data.memory_usage > newPeak.peakMemory) newPeak.peakMemory = data.memory_usage;
                if (data.disk_usage > newPeak.peakDisk) newPeak.peakDisk = data.disk_usage;
                return newPeak;
            });

            // Inject real-time data into history for instances < 12 hours old
            if (data.uptime < 43200) {
                const nowStr = new Date().toISOString();
                const memPct = memTotal > 0 ? (data.memory_usage / memTotal) * 100 : 0;
                history.update(h => {
                    const newHistory = [...h, {
                        timestamp: nowStr,
                        cpu: data.cpu_percent,
                        memory_percent: memPct
                    }];
                    // Prevent memory leaks - keep last 2000 points
                    if (newHistory.length > 2000) newHistory.shift();
                    return newHistory;
                });
            }
        } catch (e) {
            error.set(e instanceof Error ? e.message : 'Failed to fetch stats');
        }
    }

    async function fetchHistory() {
        try {
            loading.set(true);
            const response = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats/history`);
            if (!response.ok) throw new Error(`Failed to fetch history: ${response.statusText}`);
            
            const rawData = await response.json();
            const data = Array.isArray(rawData) ? rawData : (rawData.history || []);
            history.set(data);
            
            // Calculate initial peak stats from history
            if (data && data.length > 0) {
                const peakCpu = Math.max(...data.map((h: ResourceHistory) => h.cpu));
                // Peak memory from history is percent, but we want bytes. 
                // We can't recover bytes accurately without total.
                // So we'll rely on fetchStats to populate peak bytes, or accept 0 initially.
                // peakStats.set({ peakCpu, peakMemory: 0, peakDisk: 0 });
                
                // Store memory total if available (assuming API might provide it or we infer from somewhere)
                // For now, assume data[0] MIGHT have meta info, but standard ResourceHistory doesn't.
                // If we can't get it, we keep 0.
            }
        } catch (e) {
            error.set(e instanceof Error ? e.message : 'Failed to fetch history');
        } finally {
            loading.set(false);
        }
    }

    function startPolling() {
        // Start real-time stats polling (every 2 seconds)
        statsInterval = setInterval(fetchStats, 2000) as unknown as number;
        
        // Start history polling (every 60 seconds)
        historyInterval = setInterval(fetchHistory, 60000) as unknown as number;
        
        // Initial fetch
        fetchStats();
        fetchHistory();
    }

    function stopPolling() {
        if (statsInterval) {
            clearInterval(statsInterval);
            statsInterval = null;
        }
        if (historyInterval) {
            clearInterval(historyInterval);
            historyInterval = null;
        }
    }

    function refresh() {
        fetchStats();
        fetchHistory();
    }

    return {
        stats,
        history,
        peakStats,
        memTotal: memTotalStore,
        loading,
        error,
        startPolling,
        stopPolling,
        refresh
    };
}

// Utility function for getting top resource consumers
export async function getTopResourceConsumers(limit: number = 10, resourceType: 'cpu' | 'memory' | 'disk' = 'cpu'): Promise<any[]> {
    try {
        // This would need a new API endpoint for system-wide resource consumption
        // For now, we'll use existing spawner data and sort by resource usage
        const response = await fetch('/api/spawners');
        if (!response.ok) throw new Error('Failed to fetch spawners');
        
        const spawnersData = await response.json();
        const spawners = Array.isArray(spawnersData) ? spawnersData : Object.values(spawnersData);
        const allInstances: any[] = [];
        
        // Collect all instances from all spawners
        for (const spawner of spawners) {
            const instancesResponse = await fetch(`/api/spawners/${spawner.id}/instances`);
            if (instancesResponse.ok) {
                const instances = await instancesResponse.json();
                for (const instance of instances) {
                    allInstances.push({
                        ...instance,
                        spawnerId: spawner.id,
                        region: spawner.region,
                        host: spawner.host
                    });
                }
            }
        }
        
        // Sort by specified resource type
        const sorted = allInstances.sort((a, b) => {
            const aValue = resourceType === 'cpu' ? a.cpu_percent : 
                          resourceType === 'memory' ? a.memory_usage : 
                          a.disk_usage;
            const bValue = resourceType === 'cpu' ? b.cpu_percent : 
                          resourceType === 'memory' ? b.memory_usage : 
                          b.disk_usage;
            return bValue - aValue;
        });
        
        return sorted.slice(0, limit);
    } catch (e) {
        console.error('Failed to get top resource consumers:', e);
        return [];
    }
}