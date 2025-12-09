<script lang="ts">
    import type { Spawner } from '$lib/stores';
    import { createEventDispatcher } from 'svelte';

    export let spawners: Spawner[] = [];

    const dispatch = createEventDispatcher();

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
                <tr class="border-t border-slate-700 hover:bg-slate-800/50 transition group">
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
                    <td class="px-4 py-3 text-right space-x-2">
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
            {/each}
            {#if spawners.length === 0}
                <tr>
                    <td colspan="6" class="px-4 py-12 text-center text-slate-500 bg-slate-800/20 rounded-b-lg">
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