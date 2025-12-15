<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { slide } from 'svelte/transition';
    import { serverVersions } from '$lib/stores';
    import { compareVersions } from '$lib/semver';
    import PlayersChart from './PlayersChart.svelte';
    import { ChevronRight, Settings, Play, Square, RotateCw, ArrowDownToLine, Trash2, TerminalSquare } from 'lucide-svelte';

    let { spawnerId, instance }: { spawnerId: number, instance: any } = $props();

    let expanded = $state(false);
    let renameValue = $state(instance.id);
    let chartData = $state<any[]>([]);

    const dispatch = createEventDispatcher();

    // Derived state for version checking
    let activeVersion = $derived($serverVersions.find(v => v.is_active));
    let versionDiff = $derived((activeVersion && instance.version) ? compareVersions(activeVersion.version, instance.version) : 0);
    let isOutdated = $derived(versionDiff > 0); 

    function toggle() {
        expanded = !expanded;
        if (expanded) {
            renameValue = instance.id;
            fetchHistory();
        }
    }

    async function fetchHistory() {
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instance.id}/stats/history`);
            if (res.ok) {
                const data = await res.json();
                if (data.history) {
                    chartData = data.history.map((h: any) => ({
                        timestamp: new Date(h.timestamp).getTime(),
                        count: h.player_count || 0
                    }));
                }
            }
        } catch (e) {
            console.error("Failed to fetch history", e);
        }
    }

    function handleRename() {
        if (renameValue !== instance.id) {
            dispatch('rename', { spawnerId, oldId: instance.id, newId: renameValue });
        }
    }
</script>

<div class="border border-slate-700/50 rounded-lg bg-slate-800/30 overflow-hidden mb-2">
    <!-- Header / Collapsed View -->
    <div 
        class="flex items-center gap-4 px-4 py-3 cursor-pointer hover:bg-slate-700/30 transition-colors"
        onclick={toggle}
        role="button"
        tabindex="0"
        onkeydown={(e) => e.key === 'Enter' && toggle()}
    >
        <div class="text-slate-500 transform transition-transform duration-200 {expanded ? 'rotate-90' : ''}">
            <ChevronRight class="w-5 h-5" />
        </div>
        
        <div class="flex-1 grid grid-cols-12 gap-4 items-center">
            <div class="col-span-3 font-mono text-slate-300 truncate" title={`Exile Gameserver : #${instance.port}`}>
                <span class="hidden md:inline text-slate-500">Exile Gameserver : </span>#{instance.port}
            </div>
            <div class="col-span-2 text-slate-400 text-xs truncate flex items-center gap-2">
                {instance.version || 'Unknown'}
                {#if isOutdated}
                    <span class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-yellow-500/20 text-yellow-400 border border-yellow-500/30 animate-pulse">
                        OUTDATED
                    </span>
                {/if}
            </div>
            <div class="col-span-2 font-mono text-slate-400 text-xs">{instance.port}</div>
            <div class="col-span-2">
                {#if instance.status === 'Running'}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        Running
                    </span>
                {:else if instance.status === 'Provisioning'}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 animate-pulse">
                        Provisioning
                    </span>
                {:else if instance.status === 'Error'}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                        Error
                    </span>
                {:else}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-700 text-slate-400">
                        {instance.status}
                    </span>
                {/if}
            </div>
            <div class="col-span-1 font-mono text-slate-500 text-xs">{instance.pid || '-'}</div>
        </div>

                <!-- Quick Actions -->
                <div 
                    class="flex items-center gap-1 ml-auto pl-4 border-l border-slate-700/50" 
                    role="group" 
                    onclick={(e) => e.stopPropagation()}
                    onkeydown={(e) => e.stopPropagation()}
                >

                    <button 
                        onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
                        class="p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
                        title="Manage"
                    >
                        <Settings class="w-4 h-4" />
                    </button>

                                        <button 

                                            onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}

                                            disabled={instance.status === 'Running'}

                                            class="p-1.5 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                                            title="Start"

                                        >

                                            <Play class="w-4 h-4" />

                                        </button>

                                        <button 

                                            onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}

                                            disabled={instance.status !== 'Running'}

                                            class="p-1.5 text-yellow-400 hover:text-yellow-300 hover:bg-yellow-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                                            title="Stop"

                                        >

                                            <Square class="w-4 h-4" />

                                        </button>

                                        <button 

                                            onclick={() => dispatch('restart', { spawnerId, instanceId: instance.id })}

                                            disabled={instance.status !== 'Running'}

                                            class="p-1.5 text-blue-400 hover:text-blue-300 hover:bg-blue-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                                            title="Restart"

                                        >

                                            <RotateCw class="w-4 h-4" />

                                        </button>

                </div>
    </div>

    <!-- Expanded Details -->
    {#if expanded}
        <div transition:slide={{ duration: 200 }} class="bg-slate-900/50 border-t border-slate-700/50 p-4">
            <!-- Toolbar -->
            <div class="flex flex-wrap gap-2 mb-6 pb-4 border-b border-slate-700/50">
                <button 
                    onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
                    class="btn-toolbar bg-slate-700 hover:bg-slate-600 text-slate-200"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line></svg>
                    Manage
                </button>
                
                <button 
                    onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}
                    disabled={instance.status === 'Running'}
                    class="btn-toolbar bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-600/30 disabled:opacity-50"
                >
                    Start
                </button>

                <button 
                    onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}
                    disabled={instance.status !== 'Running'}
                    class="btn-toolbar bg-yellow-600/20 hover:bg-yellow-600/30 text-yellow-400 border border-yellow-600/30 disabled:opacity-50"
                >
                    Stop
                </button>

                <button 
                    onclick={() => dispatch('update', { spawnerId, instanceId: instance.id })}
                    disabled={versionDiff === 0}
                    class={`btn-toolbar disabled:opacity-50 disabled:cursor-not-allowed ${versionDiff > 0 ? 'bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-600/30' : 'bg-orange-600/20 hover:bg-orange-600/30 text-orange-400 border border-orange-600/30'}`}
                    title={versionDiff > 0 ? `Update to ${activeVersion?.version}` : (versionDiff < 0 ? `Downgrade to ${activeVersion?.version}` : 'Server is up to date')}
                >
                    {versionDiff > 0 ? 'Update' : 'Downgrade'}
                </button>

                <button 
                    onclick={() => dispatch('delete', { spawnerId, instanceId: instance.id })}
                    disabled={instance.status === 'Running'}
                    class="btn-toolbar bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-600/30 disabled:opacity-50 ml-auto"
                >
                    Delete
                </button>
            </div>

            <!-- Stats Chart -->
            <div class="mb-6 bg-slate-950/30 rounded-lg border border-slate-700/50 p-4">
                <div class="flex justify-between items-end mb-2">
                    <h4 class="text-xs font-bold text-slate-400 uppercase tracking-wider">Player Activity (24h)</h4>
                    <div class="text-sm font-mono text-blue-400">{instance.player_count || 0} active</div>
                </div>
                <PlayersChart data={chartData} height={100} color="#3b82f6" />
            </div>

            <!-- Settings Form -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">
                        Game Server Name (ID)
                    </label>
                    <div class="flex gap-2">
                        <input 
                            type="text" 
                            bind:value={renameValue}
                            class="flex-1 bg-slate-800 border border-slate-600 rounded px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-blue-500 transition-colors"
                        />
                        <button 
                            onclick={handleRename}
                            disabled={renameValue === instance.id}
                            class="px-3 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Rename
                        </button>
                    </div>
                </div>

                <div>
                    <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">
                        Assigned Port
                    </label>
                    <input 
                        type="text" 
                        value={instance.port}
                        readonly
                        class="w-full bg-slate-800/50 border border-slate-700 rounded px-3 py-2 text-sm text-slate-400 font-mono cursor-not-allowed focus:outline-none"
                    />
                </div>
            </div>
        </div>
    {/if}
</div>

<style lang="postcss">
    .btn-toolbar {
        @apply flex items-center gap-2 px-3 py-1.5 rounded text-xs font-semibold transition-all;
    }
</style>