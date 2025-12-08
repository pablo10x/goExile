<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { stats, spawners } from '$lib/stores';

    let interval: ReturnType<typeof setInterval>;

    async function fetchData() {
        try {
            const [statsRes, spawnersRes] = await Promise.all([
                fetch('/api/stats'),
                fetch('/api/spawners')
            ]);

            if (statsRes.ok) {
                stats.set(await statsRes.json());
            }
            if (spawnersRes.ok) {
                const data = await spawnersRes.json();
                spawners.set(Object.values(data)); 
            }
        } catch (e) {
            console.error('Failed to fetch data', e);
        }
    }

    onMount(() => {
        fetchData();
        interval = setInterval(fetchData, 2000);
    });

    onDestroy(() => {
        clearInterval(interval);
    });

    async function spawnInstance(spawnerId: number) {
        if (!confirm('Are you sure you want to spawn an instance?')) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/spawn`, { method: 'POST' });
            if (res.ok) {
                fetchData();
                alert('Spawn request sent!');
            } else {
                alert('Failed to spawn instance');
            }
        } catch (e) {
            alert('Error spawning instance');
        }
    }

    function formatBytes(bytes: number) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function formatUptime(ms: number) {
        if (!ms) return '0s';
        const s = Math.floor(ms / 1000);
        const m = Math.floor(s / 60);
        const h = Math.floor(m / 60);
        return `${h}h ${m % 60}m ${s % 60}s`;
    }
</script>

<div class="flex justify-between items-center mb-4 text-sm">
    <div class="text-slate-400">Last updated: {new Date().toLocaleTimeString()}</div>
    <div class="text-slate-500 font-mono text-xs">Live Update (2s)</div>
</div>

<!-- Stats Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
    <div class="stat-card border-t-2 border-t-blue-500">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">⏱️ Uptime</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            {formatUptime($stats.uptime)}
        </div>
    </div>
    <div class="stat-card border-t-2 border-t-emerald-500">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">🖥️ Active Spawners</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-emerald-400 to-cyan-400 bg-clip-text text-transparent">
            {$stats.active_spawners}
        </div>
    </div>
    <div class="stat-card border-t-2 border-t-blue-500">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">📡 Requests</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            {$stats.total_requests}
        </div>
    </div>
    <!-- Error card link could go here -->
    <a href="/errors" class="stat-card border-t-2 border-t-red-500 hover:bg-slate-800 transition cursor-pointer">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">⚠️ Errors</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-red-400 to-pink-400 bg-clip-text text-transparent">
            {$stats.total_errors}
        </div>
    </a>
    
    <div class="stat-card border-t-2 border-t-purple-500">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">💾 Memory</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-purple-400 to-indigo-400 bg-clip-text text-transparent">
            {formatBytes($stats.memory_usage)}
        </div>
    </div>
    
    <div class="stat-card border-t-2 border-t-orange-500">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">⬆️ Tx / ⬇️ Rx</div>
        <div class="text-lg font-bold text-slate-100">
            <span class="text-orange-400">{formatBytes($stats.bytes_sent)}</span> / 
            <span class="text-cyan-400">{formatBytes($stats.bytes_received)}</span>
        </div>
    </div>

    <div class="stat-card border-t-2 {$stats.db_connected ? 'border-t-emerald-500' : 'border-t-red-500'}">
        <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">🗄️ Database</div>
        <div class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            {$stats.db_connected ? '✓ Connected' : '✗ Disconnected'}
        </div>
    </div>
</div>

<!-- Spawners Section -->
<div class="card">
    <div class="border-b border-slate-700 px-8 py-6">
        <h2 class="text-2xl font-bold text-slate-50">📦 Registered Spawners</h2>
    </div>

    <div class="p-8">
        <div class="overflow-x-auto">
            <table class="w-full text-sm">
                <thead class="border-b border-slate-700">
                    <tr>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">ID</th>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Region</th>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Address</th>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Instances</th>
                        <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-slate-300">
                    {#each $spawners as spawner}
                        {@const statusClass = spawner.status === 'active' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}
                        {@const instancePercent = spawner.max_instances > 0 ? (spawner.current_instances / spawner.max_instances) * 100 : 0}
                        
                        <tr class="border-t border-slate-700 hover:bg-slate-800/50 transition">
                            <td class="px-4 py-3 text-slate-400">#{spawner.id}</td>
                            <td class="px-4 py-3 font-semibold text-slate-100">{spawner.region}</td>
                            <td class="px-4 py-3 font-mono text-slate-400">{spawner.host}:{spawner.port}</td>
                            <td class="px-4 py-3">
                                <span class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-semibold {statusClass}">
                                    <span class="w-2 h-2 rounded-full bg-current"></span>
                                    {spawner.status}
                                </span>
                            </td>
                            <td class="px-4 py-3">
                                <div class="text-sm mb-2">{spawner.current_instances} / {spawner.max_instances}</div>
                                <div class="w-full h-1.5 bg-slate-700 rounded-full overflow-hidden">
                                    <div class="h-full bg-gradient-to-r from-blue-500 to-purple-500" style="width: {instancePercent}%"></div>
                                </div>
                            </td>
                            <td class="px-4 py-3">
                                <button 
                                    onclick={() => spawnInstance(spawner.id)}
                                    class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-semibold"
                                >
                                    Spawn
                                </button>
                            </td>
                        </tr>
                    {/each}
                    {#if $spawners.length === 0}
                        <tr>
                            <td colspan="6" class="px-4 py-8 text-center text-slate-500">
                                No spawners registered.
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>