<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { stats, spawners } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import SpawnerTable from '$lib/components/SpawnerTable.svelte';
    import Drawer from '$lib/components/Drawer.svelte';
    import LogViewer from '$lib/components/LogViewer.svelte';
    import { formatBytes, formatUptime } from '$lib/utils';

    let eventSource: EventSource | null = null;
    let isConnected = false;
    let connectionStatus = 'Connecting...';
    
    // Log Viewer State
    let isLogDrawerOpen = false;
    let selectedSpawnerId: number | null = null;

    function connectSSE() {
        if (eventSource) eventSource.close();

        eventSource = new EventSource('/events');

        eventSource.onopen = () => {
            isConnected = true;
            connectionStatus = 'Live (SSE)';
        };

        eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'stats') {
                    stats.set(data.payload);
                } else if (data.type === 'spawners') {
                    // Ensure payload is an array (handle map or array from server)
                    const list = Array.isArray(data.payload) ? data.payload : Object.values(data.payload);
                    spawners.set(list);
                }
            } catch (e) {
                console.error('SSE Parse Error', e);
            }
        };

        eventSource.onerror = () => {
            isConnected = false;
            connectionStatus = 'Reconnecting...';
            // EventSource auto-reconnects, but we update UI
        };
    }

    // Fallback initial fetch
    async function initialFetch() {
        try {
            const [statsRes, spawnersRes] = await Promise.all([
                fetch('/api/stats'),
                fetch('/api/spawners')
            ]);
            if (statsRes.ok) stats.set(await statsRes.json());
            if (spawnersRes.ok) spawners.set(await spawnersRes.json());
        } catch (e) {
            console.error('Initial fetch failed', e);
        }
    }

    onMount(() => {
        initialFetch();
        connectSSE();
    });

    onDestroy(() => {
        if (eventSource) eventSource.close();
    });

    async function handleSpawn(event: CustomEvent<number>) {
        const id = event.detail;
        if (!confirm(`Are you sure you want to spawn an instance on Spawner #${id}?`)) return;
        
        try {
            const res = await fetch(`/api/spawners/${id}/spawn`, { method: 'POST' });
            if (res.ok) {
                // Optimistic update or wait for SSE? SSE is fast enough.
                alert('Spawn request sent!');
            } else {
                const err = await res.json();
                alert(`Failed: ${err.error || 'Unknown error'}`);
            }
        } catch (e) {
            alert('Error sending spawn request');
        }
    }

    function handleViewLogs(event: CustomEvent<number>) {
        selectedSpawnerId = event.detail;
        isLogDrawerOpen = true;
    }
</script>

<div class="flex justify-between items-center mb-6">
    <div>
        <h1 class="text-2xl font-bold text-slate-50">Dashboard</h1>
        <div class="flex items-center gap-2 mt-1">
            <div class={`w-2 h-2 rounded-full ${isConnected ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></div>
            <span class="text-xs font-mono text-slate-400">{connectionStatus}</span>
        </div>
    </div>
    <div class="text-slate-500 text-sm">
        {new Date().toLocaleDateString()}
    </div>
</div>

<!-- Stats Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
    <StatsCard 
        title="⏱️ Uptime" 
        value={formatUptime($stats.uptime)} 
        borderColorClass="border-t-blue-500"
    />
    <StatsCard 
        title="🖥️ Active Spawners" 
        value={$stats.active_spawners} 
        borderColorClass="border-t-emerald-500"
        valueGradientClass="from-emerald-400 to-cyan-400"
    />
    <StatsCard 
        title="📡 Requests" 
        value={$stats.total_requests} 
        borderColorClass="border-t-blue-500"
    />
    <a href="/errors" class="block transition hover:scale-[1.02]">
        <StatsCard 
            title="⚠️ Errors" 
            value={$stats.total_errors} 
            borderColorClass="border-t-red-500"
            valueGradientClass="from-red-400 to-pink-400"
        />
    </a>
    <StatsCard 
        title="💾 Memory" 
        value={formatBytes($stats.memory_usage)} 
        borderColorClass="border-t-purple-500"
        valueGradientClass="from-purple-400 to-indigo-400"
    />
</div>

<!-- Secondary Stats Row (Network & DB) -->
<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
    <StatsCard 
        title="⬆️ Tx / ⬇️ Rx" 
        value="" 
        subValue={`<span class="text-orange-400">${formatBytes($stats.bytes_sent)}</span> / <span class="text-cyan-400">${formatBytes($stats.bytes_received)}</span>`}
        borderColorClass="border-t-orange-500"
    />
    <StatsCard 
        title="🗄️ Database" 
        value={$stats.db_connected ? '✓ Connected' : '✗ Disconnected'} 
        borderColorClass={$stats.db_connected ? "border-t-emerald-500" : "border-t-red-500"}
        valueGradientClass={$stats.db_connected ? "from-emerald-400 to-cyan-400" : "from-red-400 to-orange-400"}
    />
</div>

<!-- Spawners Section -->
<div class="card bg-slate-800/30 border border-slate-700/50">
    <div class="border-b border-slate-700 px-6 py-4 flex justify-between items-center">
        <h2 class="text-xl font-bold text-slate-50">📦 Registered Spawners</h2>
        <span class="text-xs text-slate-500 uppercase tracking-widest font-semibold">Real-time</span>
    </div>
    <div class="p-0">
        <SpawnerTable 
            spawners={$spawners} 
            on:spawn={handleSpawn} 
            on:viewLogs={handleViewLogs}
        />
    </div>
</div>

<!-- Log Drawer -->
<Drawer 
    isOpen={isLogDrawerOpen} 
    onClose={() => isLogDrawerOpen = false} 
    title={`Spawner #${selectedSpawnerId} Logs`}
>
    {#if selectedSpawnerId}
        <LogViewer spawnerId={selectedSpawnerId} />
    {/if}
</Drawer>