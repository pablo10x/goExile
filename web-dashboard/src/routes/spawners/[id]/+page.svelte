<script lang="ts">
    import { page } from '$app/stores';
    import { onMount, onDestroy } from 'svelte';
    import { formatBytes, formatUptime } from '$lib/utils';
    import InstanceRow from '$lib/components/InstanceRow.svelte';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
    import LogViewerModal from '$lib/components/LogViewerModal.svelte';
    import Dropdown from '$lib/components/Dropdown.svelte';
    import { serverVersions } from '$lib/stores';
    import { compareVersions } from '$lib/semver';
    import PlayersChart from '$lib/components/PlayersChart.svelte';
    import { Server, Cpu, HardDrive, MemoryStick, List, Plus } from 'lucide-svelte';

    const spawnerId = parseInt($page.params.id || '0');

    let spawner: any = null;
    let instances: any[] = [];
    let isLoading = true;
    let error: string | null = null;
    let refreshInterval: any;

    // Dummy Data for Chart
    const chartData = Array.from({ length: 24 }, (_, i) => {
        const time = new Date().getTime() - (23 - i) * 3600000;
        return {
            timestamp: time,
            count: Math.floor(Math.max(0, 50 + Math.sin(i / 3) * 30 + (Math.random() - 0.5) * 20))
        };
    });

    // Instance Action State
    let isInstanceActionDialogOpen = false;
    let instanceActionType: 'start' | 'stop' | 'delete' | 'update' | 'rename' | 'restart' | 'bulk_stop' | 'bulk_restart' | 'bulk_update' | 'bulk_start' | 'update_spawner_build' | null = null;
    let instanceActionInstanceId: string | null = null;
    let instanceActionNewID: string | null = null;
    let instanceActionBulkIds: string[] = [];
    let instanceActionDialogTitle = '';
    let instanceActionDialogMessage = '';
    let instanceActionConfirmText = '';
    
    // Progress State for ConfirmDialog
    let actionProgress: number | null = null;
    let actionStatusMessage: string | null = null;

    // Spawn Dialog
    let isSpawnDialogOpen = false;

    // Console & Logs
    let isConsoleOpen = false;
    let consoleInstanceId: string | null = null;
    let isLogViewerOpen = false;

    $: activeVersion = $serverVersions.find(v => v.is_active);

    async function fetchSpawnerData() {
        try {
            const res = await fetch(`/api/spawners/${spawnerId}`);
            if (!res.ok) {
                if (res.status === 404) throw new Error('Spawner not found');
                throw new Error('Failed to load spawner details');
            }
            spawner = await res.json();

            const instRes = await fetch(`/api/spawners/${spawnerId}/instances`);
            if (instRes.ok) {
                const data = await instRes.json();
                const list = data.instances || data || [];
                // Sort by ID to prevent jumping
                instances = list.sort((a: any, b: any) => a.id.localeCompare(b.id));
            }
        } catch (e: any) {
            error = e.message;
        } finally {
            isLoading = false;
        }
    }

    onMount(() => {
        fetchSpawnerData();
        refreshInterval = setInterval(fetchSpawnerData, 5000); // Poll every 5s
        
        // Fetch versions if not already loaded
        if ($serverVersions.length === 0) {
            fetch('/api/versions').then(r => r.json()).then(v => serverVersions.set(v)).catch(console.error);
        }
    });

    onDestroy(() => {
        if (refreshInterval) clearInterval(refreshInterval);
    });

    function getStatusClass(status: string) {
        if (status === 'Updating') return 'bg-orange-500/10 text-orange-400 border-orange-500/20 animate-pulse';
        return status === 'active' 
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
            : 'bg-red-500/10 text-red-400 border-red-500/20';
    }

    // --- Action Handlers ---

    function openSpawnDialog() {
        isSpawnDialogOpen = true;
    }

    async function executeSpawn() {
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/spawn`, { method: 'POST' });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || `Server returned ${res.status}`);
            }
            const instance = await res.json();
            consoleInstanceId = instance.id;
            isConsoleOpen = true;
            fetchSpawnerData();
        } catch (e: any) {
            alert(e.message);
        }
        isSpawnDialogOpen = false;
    }

    function openUpdateSpawnerBuildDialog() {
        instanceActionType = 'update_spawner_build';
        instanceActionDialogTitle = 'Update Spawner Build';
        instanceActionDialogMessage = `Are you sure you want Spawner #${spawnerId} to download the latest game server build? This might take a while.`;
        instanceActionConfirmText = 'Update Build';
        isInstanceActionDialogOpen = true;
    }

    function openInstanceActionDialog(type: any, instanceId: string, extra?: any) {
        instanceActionType = type;
        instanceActionInstanceId = instanceId;
        
        if (type === 'start') {
            instanceActionDialogTitle = 'Start Instance';
            instanceActionDialogMessage = `Are you sure you want to start instance "${instanceId}"?`;
            instanceActionConfirmText = 'Start';
        } else if (type === 'stop') {
            instanceActionDialogTitle = 'Stop Instance';
            instanceActionDialogMessage = `Are you sure you want to stop instance "${instanceId}"?`;
            instanceActionConfirmText = 'Stop';
        } else if (type === 'restart') {
            instanceActionDialogTitle = 'Restart Instance';
            instanceActionDialogMessage = `Are you sure you want to restart instance "${instanceId}"?`;
            instanceActionConfirmText = 'Restart';
        } else if (type === 'delete') {
            instanceActionDialogTitle = 'Delete Instance';
            instanceActionDialogMessage = `Are you sure you want to PERMANENTLY DELETE instance "${instanceId}"?`;
            instanceActionConfirmText = 'Delete';
        } else if (type === 'update') {
            instanceActionDialogTitle = 'Update Instance';
            instanceActionDialogMessage = `Are you sure you want to update instance "${instanceId}"?`;
            instanceActionConfirmText = 'Update';
        } else if (type === 'rename') {
            instanceActionNewID = extra;
            instanceActionDialogTitle = 'Rename Instance';
            instanceActionDialogMessage = `Rename "${instanceId}" to "${extra}"?`;
            instanceActionConfirmText = 'Rename';
        }

        isInstanceActionDialogOpen = true;
    }

    function dispatchBulkAction(action: 'start' | 'stop' | 'restart' | 'update') {
        let targetInstances = [];
        if (action === 'start') targetInstances = instances.filter(i => i.status !== 'Running' && i.status !== 'Provisioning');
        else if (action === 'stop') targetInstances = instances.filter(i => i.status === 'Running');
        else if (action === 'restart') targetInstances = instances.filter(i => i.status === 'Running');
        else if (action === 'update') targetInstances = instances.filter(i => activeVersion && i.version !== activeVersion.version);

        if (targetInstances.length === 0) return;

        instanceActionType = `bulk_${action}` as any;
        instanceActionBulkIds = targetInstances.map(i => i.id);
        
        const actionName = action.charAt(0).toUpperCase() + action.slice(1);
        instanceActionDialogTitle = `${actionName} All Instances`;
        instanceActionDialogMessage = `Are you sure you want to ${action} ${targetInstances.length} instances?`;
        instanceActionConfirmText = `${actionName} All`;
        isInstanceActionDialogOpen = true;
    }

    async function executeInstanceAction() {
        try {
            let res: Response | any = null;
            actionProgress = null;
            actionStatusMessage = null;

            if (instanceActionType === 'update_spawner_build') {
                actionStatusMessage = "Requesting spawner update...";
                res = await fetch(`/api/spawners/${spawnerId}/update-template`, { method: 'POST' });
            } else if (instanceActionType?.startsWith('bulk_')) {
                const action = instanceActionType.replace('bulk_', '');
                let failureCount = 0;
                
                // Initialize progress
                actionProgress = 0;
                const total = instanceActionBulkIds.length;
                
                // Execute sequentially or in small batches to show progress
                for (let i = 0; i < total; i++) {
                    const id = instanceActionBulkIds[i];
                    actionStatusMessage = `Processing ${id} (${i + 1}/${total})...`;
                    
                    try {
                        let url = '';
                        if (action === 'restart') {
                             await fetch(`/api/spawners/${spawnerId}/instances/${id}/stop`, { method: 'POST' });
                             // Small delay between stop and start?
                             url = `/api/spawners/${spawnerId}/instances/${id}/start`;
                        } else {
                            url = `/api/spawners/${spawnerId}/instances/${id}/${action}`;
                        }
                        const r = await fetch(url, { method: 'POST' });
                        if (!r.ok) throw new Error('Failed');
                    } catch { 
                        failureCount++; 
                    }

                    // Update progress
                    actionProgress = ((i + 1) / total) * 100;
                }
                
                if (failureCount > 0) alert(`${failureCount} actions failed.`);
                res = { ok: true, text: () => Promise.resolve('') } as any;
                
                // Brief delay to show 100%
                await new Promise(r => setTimeout(r, 500));
            } else {
                // Single instance action
                actionStatusMessage = "Executing action...";
                let url = `/api/spawners/${spawnerId}/instances/${instanceActionInstanceId}`;
                let method = 'POST';
                
                if (instanceActionType === 'delete') {
                    method = 'DELETE';
                    res = await fetch(url, { method });
                } else if (instanceActionType === 'rename') {
                    url += '/rename';
                    res = await fetch(url, {
                        method,
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ new_id: instanceActionNewID })
                    });
                } else if (instanceActionType === 'restart') {
                     actionStatusMessage = "Stopping instance...";
                     await fetch(url + '/stop', { method: 'POST' });
                     actionStatusMessage = "Starting instance...";
                     url += '/start';
                     res = await fetch(url, { method: 'POST' });
                } else {
                    url += `/${instanceActionType}`;
                    res = await fetch(url, { method });
                }
            }

            if (res && !res.ok) {
                 const txt = await res.text();
                 throw new Error(txt || 'Action failed');
            }
            
            fetchSpawnerData();
        } catch (e: any) {
            alert(e.message); // The dialog catches errors too, but alert is fine for fallback
            throw e; // Re-throw so dialog stays open or handles error
        } finally {
            // Reset progress state if we are closing manually, 
            // but ConfirmDialog might close itself on success.
            // We set them to null just in case.
            actionProgress = null;
            actionStatusMessage = null;
        }
        isInstanceActionDialogOpen = false;
    }

    function handleTail(e: CustomEvent) {
        consoleInstanceId = e.detail.instanceId;
        isConsoleOpen = true;
    }

</script>

<div class="container mx-auto max-w-7xl">
    <div class="mb-6 flex items-center gap-4">
        <a href="/" aria-label="Back to Dashboard" class="p-2 rounded-lg bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        </a>
        <h1 class="text-2xl font-bold text-slate-50">Spawner #{spawnerId}</h1>
        {#if spawner}
            <div class="flex items-center gap-2">
                <span class={`px-3 py-1 rounded-full text-xs font-semibold border ${getStatusClass(spawner.status)}`}>
                    {spawner.status}
                </span>
                <span class="px-3 py-1 rounded-full text-xs font-semibold bg-slate-700 text-slate-300 border border-slate-600 font-mono" title="Current Game Server Version">
                    v{spawner.game_version || 'Unknown'}
                </span>
            </div>
        {/if}
        <div class="ml-auto flex items-center gap-3">
            <button 
                onclick={() => isLogViewerOpen = true}
                class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-sm font-semibold transition-colors border border-slate-700"
            >
                View Logs
            </button>
            
            {#if spawner && activeVersion}
                {@const cmp = compareVersions(activeVersion.version, spawner.game_version || '0.0.0')}
                {#if cmp !== 0}
                    <button 
                        onclick={() => openUpdateSpawnerBuildDialog()}
                        disabled={spawner.status === 'Updating'}
                        class={`px-4 py-2 rounded-lg text-sm font-semibold transition-colors border shadow-lg ${spawner.status === 'Updating' ? 'bg-slate-700 text-slate-500 border-slate-600 cursor-not-allowed' : (cmp > 0 ? 'bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-500 shadow-emerald-900/20' : 'bg-orange-600 hover:bg-orange-500 text-white border-orange-500 shadow-orange-900/20')}`}
                        title={spawner.status === 'Updating' ? 'Update in progress...' : (cmp > 0 ? `Upgrade to ${activeVersion.version}` : `Downgrade to ${activeVersion.version}`)}
                    >
                        {spawner.status === 'Updating' ? 'Updating...' : (cmp > 0 ? 'Update Build' : 'Downgrade')}
                    </button>
                {:else}
                    <button 
                        disabled
                        class="px-4 py-2 bg-slate-800/50 text-slate-500 rounded-lg text-sm font-semibold border border-slate-700/50 cursor-not-allowed"
                    >
                        Build Up-to-Date
                    </button>
                {/if}
            {/if}
        </div>
    </div>

    {#if isLoading && !spawner}
        <div class="flex items-center justify-center h-64 text-slate-500">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-slate-500 mr-3"></div>
            Loading Spawner Details...
        </div>
    {:else if error}
        <div class="p-8 text-center text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg">
            <h2 class="text-lg font-bold mb-2">Error</h2>
            <p>{error}</p>
        </div>
    {:else if spawner}
        <!-- Spawner Stats -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <StatsCard 
                title="Active Instances" 
                value={`${spawner.current_instances} / ${spawner.max_instances}`} 
                subValue={spawner.max_instances > 0 ? `${((spawner.current_instances / spawner.max_instances) * 100).toFixed(1)}% Capacity` : ''}
                subValueClass="text-slate-400 text-xs mt-1"
                Icon={Server}
                color="blue"
            />
            <StatsCard 
                title="CPU Usage" 
                value={`${spawner.cpu_usage?.toFixed(1) || 0}%`} 
                Icon={Cpu}
                color="purple"
            />
            <StatsCard 
                title="Memory Usage" 
                value={formatBytes(spawner.mem_used || 0)} 
                subValue={`of ${formatBytes(spawner.mem_total || 0)}`}
                subValueClass="text-slate-400 text-xs mt-1"
                Icon={MemoryStick}
                color="emerald"
            />
            <StatsCard 
                title="Disk Usage" 
                value={formatBytes(spawner.disk_used || 0)} 
                subValue={`of ${formatBytes(spawner.disk_total || 0)}`}
                subValueClass="text-slate-400 text-xs mt-1"
                Icon={HardDrive}
                color="orange"
            />
        </div>

        <!-- Instances Section -->
        <div class="bg-slate-800/30 border border-slate-700/50 rounded-xl overflow-hidden backdrop-blur-sm relative">
            {#if spawner.status === 'Updating'}
                <div class="absolute inset-0 z-50 bg-slate-900/80 backdrop-blur-sm flex flex-col items-center justify-center text-slate-300">
                    <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
                    <h3 class="text-xl font-bold text-white">System Updating</h3>
                    <p class="text-sm">Downloading game files. Actions are disabled.</p>
                </div>
            {/if}

            <div class="p-6 border-b border-slate-700/50 flex flex-wrap items-center justify-between gap-4">
                <div>
                    <h2 class="text-xl font-bold text-slate-50">Game Instances</h2>
                    <p class="text-sm text-slate-400 mt-1">Manage game servers running on this spawner.</p>
                </div>
                
                <div class="flex items-center gap-3">
                    <Dropdown label="Bulk Actions" Icon={List}>
                        <div slot="default" let:close>
                            <button onclick={() => { dispatchBulkAction('start'); close(); }} class="w-full text-left px-4 py-2 text-sm text-emerald-400 hover:bg-emerald-500/10">Start All</button>
                            <button onclick={() => { dispatchBulkAction('stop'); close(); }} class="w-full text-left px-4 py-2 text-sm text-yellow-400 hover:bg-yellow-500/10">Stop All</button>
                            <button onclick={() => { dispatchBulkAction('restart'); close(); }} class="w-full text-left px-4 py-2 text-sm text-blue-400 hover:bg-blue-500/10">Restart All</button>
                            <button onclick={() => { dispatchBulkAction('update'); close(); }} class="w-full text-left px-4 py-2 text-sm text-purple-400 hover:bg-purple-500/10 border-t border-slate-700">Update Outdated</button>
                        </div>
                    </Dropdown>

                    <button 
                        onclick={openSpawnDialog}
                        class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-semibold transition-all shadow-lg shadow-blue-900/20"
                    >
                        <Plus class="w-4 h-4" />
                        Spawn Instance
                    </button>
                </div>
            </div>

            <div class="p-6">
                <!-- Players Chart -->
                <div class="mb-8">
                    <div class="flex justify-between items-end mb-4 px-1">
                        <div>
                            <h3 class="text-sm font-bold text-slate-300 uppercase tracking-wider">Active Players (24h)</h3>
                            <p class="text-xs text-slate-500 mt-1">Real-time player concurrency across all instances.</p>
                        </div>
                        <div class="text-2xl font-bold text-emerald-400">{chartData[chartData.length-1].count}</div>
                    </div>
                    <div class="h-48 bg-slate-900/50 rounded-xl border border-slate-700/50 overflow-hidden relative shadow-inner">
                        <PlayersChart data={chartData} height={192} />
                    </div>
                </div>

                {#if instances.length === 0}
                    <div class="text-center py-12 text-slate-500 bg-slate-800/50 rounded-lg border border-slate-700/50 border-dashed">
                        <p class="text-lg mb-2">No instances running</p>
                        <p class="text-sm">Click "Spawn Instance" to start a new game server.</p>
                    </div>
                {:else}
                    <div class="space-y-3">
                        {#each instances as instance (instance.id)}
                            <InstanceRow 
                                spawnerId={spawnerId}
                                instance={instance}
                                on:tail={handleTail}
                                on:start={(e) => openInstanceActionDialog('start', e.detail.instanceId)}
                                on:stop={(e) => openInstanceActionDialog('stop', e.detail.instanceId)}
                                on:restart={(e) => openInstanceActionDialog('restart', e.detail.instanceId)}
                                on:delete={(e) => openInstanceActionDialog('delete', e.detail.instanceId)}
                                on:update={(e) => openInstanceActionDialog('update', e.detail.instanceId)}
                                on:rename={(e) => openInstanceActionDialog('rename', e.detail.oldId, e.detail.newId)}
                            />
                        {/each}
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>

<!-- Log Viewer Modal -->
{#if spawner}
    <LogViewer
        spawnerId={spawnerId}
        isOpen={isLogViewerOpen}
        onClose={() => isLogViewerOpen = false}
    />
{/if}

<!-- Instance Console Modal -->
<InstanceManagerModal
    bind:isOpen={isConsoleOpen}
    spawnerId={spawnerId}
    instanceId={consoleInstanceId}
    onClose={() => isConsoleOpen = false}
    memTotal={spawner?.mem_total || 0}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
    bind:isOpen={isSpawnDialogOpen}
    title="Spawn New Instance"
    message={`Are you sure you want to spawn a new game server instance on Spawner #${spawnerId}?`}
    confirmText="Spawn Server"
    onConfirm={executeSpawn}
/>

<!-- Instance Action Confirmation Dialog -->
<ConfirmDialog
    bind:isOpen={isInstanceActionDialogOpen}
    title={instanceActionDialogTitle}
    message={instanceActionDialogMessage}
    confirmText={instanceActionConfirmText}
    onConfirm={executeInstanceAction}
    progress={actionProgress}
    statusMessage={actionStatusMessage}
/>
