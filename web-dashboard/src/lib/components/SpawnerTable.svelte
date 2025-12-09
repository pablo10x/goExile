<script lang="ts">
    import type { Spawner } from '$lib/stores';
    import { createEventDispatcher } from 'svelte';
    import { formatBytes } from '$lib/utils';

    export let spawners: Spawner[] = [];
    
    let expandedRows: Set<number> = new Set();
    let activeInstances: Record<number, any[]> = {};
    let loadingInstances: Record<number, boolean> = {};

    const dispatch = createEventDispatcher();

    function toggleRow(id: number) {
        if (expandedRows.has(id)) {
            expandedRows.delete(id);
        } else {
            expandedRows.add(id);
            fetchInstances(id);
        }
        expandedRows = expandedRows; // Trigger reactivity
    }

    async function fetchInstances(id: number) {
        loadingInstances[id] = true;
        try {
            const res = await fetch(`/api/spawners/${id}/instances`);
            if (res.ok) {
                const data = await res.json();
                // handle { instances: [...] } wrapper if present
                activeInstances[id] = data.instances || data || []; 
            } else {
                activeInstances[id] = [];
            }
        } catch (e) {
            console.error('Failed to fetch instances', e);
            activeInstances[id] = [];
        } finally {
            loadingInstances[id] = false;
        }
    }

    async function stopInstance(spawnerId: number, instanceId: string) {
        if (!confirm(`Stop instance ${instanceId}?`)) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}`, { method: 'DELETE' });
            if (res.ok) {
                fetchInstances(spawnerId); // Refresh list
            } else {
                alert('Failed to stop instance');
            }
        } catch (e) {
            alert('Error stopping instance');
        }
    }

    function getStatusClass(status: string) {
        return status === 'active' 
            ? 'bg-emerald-500/10 text-emerald-400' 
            : 'bg-red-500/10 text-red-400';
    }

    function getInstancePercent(spawner: Spawner) {
        return spawner.max_instances > 0 
            ? (spawner.current_instances / spawner.max_instances) * 100 
            : 0;
    }
</script>

<div class="overflow-x-auto">
    <table class="w-full text-sm">
        <thead class="border-b border-slate-700">
            <tr>
                <th class="w-8"></th> <!-- Caret column -->
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">ID</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Region</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Address</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">Instances</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider text-right">Actions</th>
            </tr>
        </thead>
        <tbody class="text-slate-300">
            {#each spawners as spawner}
                <tr class="border-t border-slate-700 hover:bg-slate-800/50 transition cursor-pointer" onclick={() => toggleRow(spawner.id)}>
                    <td class="px-4 py-3 text-center text-slate-500">
                        <span class="inline-block transition-transform duration-200 {expandedRows.has(spawner.id) ? 'rotate-90' : ''}">▶</span>
                    </td>
                    <td class="px-4 py-3 text-slate-400">#{spawner.id}</td>
                    <td class="px-4 py-3 font-semibold text-slate-100">{spawner.region}</td>
                    <td class="px-4 py-3 font-mono text-slate-400">{spawner.host}:{spawner.port}</td>
                    <td class="px-4 py-3">
                        <span class={`inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-semibold ${getStatusClass(spawner.status)}`}>
                            <span class="w-2 h-2 rounded-full bg-current"></span>
                            {spawner.status}
                        </span>
                    </td>
                    <td class="px-4 py-3">
                        <div class="text-sm mb-2">{spawner.current_instances} / {spawner.max_instances}</div>
                        <div class="w-full h-1.5 bg-slate-700 rounded-full overflow-hidden">
                            <div class="h-full bg-gradient-to-r from-blue-500 to-purple-500" style={`width: ${getInstancePercent(spawner)}%`}></div>
                        </div>
                    </td>
                    <td class="px-4 py-3 text-right space-x-2" onclick={(e) => e.stopPropagation()}>
                        <button 
                            onclick={() => dispatch('viewLogs', spawner.id)}
                            class="px-3 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded text-xs font-semibold transition-colors"
                        >
                            Logs
                        </button>
                        <button 
                            onclick={() => dispatch('spawn', spawner.id)}
                            class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-semibold transition-colors shadow-lg shadow-blue-900/20"
                        >
                            Spawn
                        </button>
                    </td>
                </tr>
                <!-- Details Row -->
                {#if expandedRows.has(spawner.id)}
                    <tr class="bg-slate-900/50 border-b border-slate-700">
                        <td colspan="7" class="px-4 py-4">
                            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 text-xs">
                                <!-- CPU -->
                                <div class="bg-slate-800 rounded p-3 border border-slate-700">
                                    <div class="text-slate-400 uppercase font-semibold mb-1">CPU Usage</div>
                                    <div class="text-lg font-bold text-slate-100">{spawner.cpu_usage ? spawner.cpu_usage.toFixed(1) : 0}%</div>
                                    <div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
                                        <div class="h-full bg-blue-500" style="width: {spawner.cpu_usage || 0}%"></div>
                                    </div>
                                </div>
                                <!-- Memory -->
                                <div class="bg-slate-800 rounded p-3 border border-slate-700">
                                    <div class="text-slate-400 uppercase font-semibold mb-1">Memory Usage</div>
                                    <div class="text-lg font-bold text-slate-100">
                                        {formatBytes(spawner.mem_used || 0)} / {formatBytes(spawner.mem_total || 0)}
                                    </div>
                                    <div class="text-slate-500 mt-1">
                                        Free: {formatBytes((spawner.mem_total || 0) - (spawner.mem_used || 0))}
                                    </div>
                                    <div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
                                        <div class="h-full bg-purple-500" style="width: {spawner.mem_total ? (spawner.mem_used / spawner.mem_total) * 100 : 0}%"></div>
                                    </div>
                                </div>
                                <!-- Disk -->
                                <div class="bg-slate-800 rounded p-3 border border-slate-700">
                                    <div class="text-slate-400 uppercase font-semibold mb-1">Disk Usage</div>
                                    <div class="text-lg font-bold text-slate-100">
                                        {formatBytes(spawner.disk_used || 0)} / {formatBytes(spawner.disk_total || 0)}
                                    </div>
                                    <div class="w-full h-1 bg-slate-700 rounded-full mt-2 overflow-hidden">
                                        <div class="h-full bg-orange-500" style="width: {spawner.disk_total ? (spawner.disk_used / spawner.disk_total) * 100 : 0}%"></div>
                                    </div>
                                </div>
                            </div>

                            <!-- Active Instances -->
                            <div class="mt-6 bg-slate-800 rounded border border-slate-700 overflow-hidden">
                                <div class="px-4 py-2 bg-slate-700/50 border-b border-slate-700 flex justify-between items-center">
                                    <h4 class="text-xs font-bold text-slate-300 uppercase tracking-wider">Active Instances</h4>
                                    <button onclick={() => fetchInstances(spawner.id)} class="text-xs text-blue-400 hover:text-blue-300">Refresh</button>
                                </div>
                                {#if loadingInstances[spawner.id]}
                                    <div class="p-4 text-center text-slate-500 text-xs">Loading instances...</div>
                                {:else if !activeInstances[spawner.id] || activeInstances[spawner.id].length === 0}
                                    <div class="p-4 text-center text-slate-500 text-xs">No active instances found.</div>
                                {:else}
                                    <table class="w-full text-xs text-left">
                                        <thead class="bg-slate-700/20 text-slate-400">
                                            <tr>
                                                <th class="px-4 py-2">ID</th>
                                                <th class="px-4 py-2">Port</th>
                                                <th class="px-4 py-2">PID</th>
                                                <th class="px-4 py-2 text-right">Action</th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-slate-700/50">
                                            {#each activeInstances[spawner.id] as instance}
                                                <tr class="hover:bg-slate-700/10">
                                                    <td class="px-4 py-2 text-slate-300 font-mono">{instance.id}</td>
                                                    <td class="px-4 py-2 text-slate-300 font-mono">{instance.port}</td>
                                                    <td class="px-4 py-2 text-slate-400 font-mono">{instance.pid || '-'}</td>
                                                    <td class="px-4 py-2 text-right">
                                                        <button 
                                                            onclick={() => stopInstance(spawner.id, instance.id)}
                                                            class="text-red-400 hover:text-red-300 font-semibold px-2 py-1 rounded hover:bg-red-400/10 transition-colors"
                                                        >
                                                            Stop
                                                        </button>
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                {/if}
                            </div>
                        </td>
                    </tr>
                {/if}
            {/each}
            {#if spawners.length === 0}
                <tr>
                    <td colspan="7" class="px-4 py-12 text-center text-slate-500 bg-slate-800/20 rounded-b-lg">
                        <div class="flex flex-col items-center gap-2">
                            <span class="text-2xl">📭</span>
                            <span>No spawners registered yet.</span>
                        </div>
                    </td>
                </tr>
            {/if}
        </tbody>
    </table>
</div>