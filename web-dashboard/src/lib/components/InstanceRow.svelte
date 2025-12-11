<script lang="ts">
            import { createEventDispatcher } from 'svelte';
            import { slide } from 'svelte/transition';
            import { serverVersions } from '$lib/stores';
            import { compareVersions } from '$lib/semver';
            import PlayersChart from './PlayersChart.svelte';
            import BackupModal from './BackupModal.svelte';
        
            export let spawnerId: number;
            export let instance: any;
        
            let expanded = false;
            let renameValue = instance.id;
            let chartData: any[] = [];
            let isBackupModalOpen = false;
        
            const dispatch = createEventDispatcher();    
        // Derived state for version checking
        $: activeVersion = $serverVersions.find(v => v.is_active);
        $: versionDiff = (activeVersion && instance.version) ? compareVersions(activeVersion.version, instance.version) : 0;
        $: isOutdated = versionDiff > 0; // Only true if active is NEWER than instance
    
        function toggle() {
            expanded = !expanded;
            if (expanded) {
                renameValue = instance.id;
                if (chartData.length === 0) {
                    chartData = Array.from({ length: 24 }, (_, i) => {
                        const time = new Date().getTime() - (23 - i) * 3600000; // Last 24h
                        return {
                            timestamp: time,
                            count: Math.floor(Math.random() * 30) // Random 0-30 players
                        };
                    });
                }
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
    >
        <div class="text-slate-500 transform transition-transform duration-200 {expanded ? 'rotate-90' : ''}">
            ▶
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

                <!-- svelte-ignore a11y_click_events_have_key_events -->

                <!-- svelte-ignore a11y_no_static_element_interactions -->

                <div class="flex items-center gap-1 ml-auto pl-4 border-l border-slate-700/50" onclick={(e) => e.stopPropagation()}>

                    <button 

                        onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}

                        class="p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"

                        title="Console"

                    >

                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>

                    </button>

                    <button 

                        onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}

                        disabled={instance.status === 'Running'}

                        class="p-1.5 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                        title="Start"

                    >

                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>

                    </button>

                    <button 

                        onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}

                        disabled={instance.status !== 'Running'}

                        class="p-1.5 text-yellow-400 hover:text-yellow-300 hover:bg-yellow-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                        title="Stop"

                    >

                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg>

                    </button>

                    <button 

                        onclick={() => dispatch('restart', { spawnerId, instanceId: instance.id })}

                        disabled={instance.status !== 'Running'}

                        class="p-1.5 text-blue-400 hover:text-blue-300 hover:bg-blue-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"

                        title="Restart"

                    >

                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M23 4v6h-6"></path><path d="M1 20v6h6"></path><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>

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
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>
                    Console
                </button>

                <button 
                    onclick={() => isBackupModalOpen = true}
                    class="btn-toolbar bg-slate-700 hover:bg-slate-600 text-slate-200"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v13a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                    Backups
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
                    <div class="text-sm font-mono text-blue-400">{chartData[chartData.length-1]?.count || 0} active</div>
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
    <BackupModal 
        bind:isOpen={isBackupModalOpen} 
        spawnerId={spawnerId} 
        instanceId={instance.id} 
    />
</div>

<style lang="postcss">
    .btn-toolbar {
        @apply flex items-center gap-2 px-3 py-1.5 rounded text-xs font-semibold transition-all;
    }
</style>