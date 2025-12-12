<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { stats, spawners, serverVersions } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import SpawnerTable from '$lib/components/SpawnerTable.svelte';
    import Drawer from '$lib/components/Drawer.svelte';
    import LogViewer from '$lib/components/LogViewer.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
    import TopResourceConsumers from '$lib/components/TopResourceConsumers.svelte';
    import { formatBytes, formatUptime } from '$lib/utils';
    import { Clock, Server, Activity, AlertCircle, Database, Network } from 'lucide-svelte';

    let eventSource: EventSource | null = null;
    let isConnected = false;
    let connectionStatus = 'Connecting...';
    
    // Log Viewer State
    let isLogDrawerOpen = false;
    let selectedSpawnerId: number | null = null;

    // Instance Console State
    let isConsoleOpen = false;
    let consoleSpawnerId: number | null = null;
    let consoleInstanceId: string | null = null;

    // Spawn Dialog State
    let isSpawnDialogOpen = false;
    let spawnTargetId: number | null = null;

    // Instance Action Dialog State (Start/Stop)
    let isInstanceActionDialogOpen = false;
    let instanceActionType: 'start' | 'stop' | 'delete' | 'update' | 'rename' | 'restart' | 'bulk_stop' | 'bulk_restart' | 'bulk_update' | 'bulk_start' | 'update_spawner_build' | null = null;
    let instanceActionSpawnerId: number | null = null;
    let instanceActionInstanceId: string | null = null;
    let instanceActionNewID: string | null = null;
    let instanceActionBulkIds: string[] = []; // New state for bulk IDs
    let instanceActionDialogTitle = '';
    let instanceActionDialogMessage = '';
    let instanceActionConfirmText = '';

    // Animation state for new spawners
    let previousSpawnerIds: Set<number> = new Set();
    let highlightNewSpawnerId: number | null = null;
    let highlightTimeout: ReturnType<typeof setTimeout> | null = null;

    let spawnerTableComponent: any;

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
                    const list: any[] = Array.isArray(data.payload) ? data.payload : Object.values(data.payload);
                    
                    // Detect new spawners for animation
                    const currentSpawnerIds = new Set<number>(list.map((s: any) => s.id));
                    for (const spawner of list) {
                        if (!previousSpawnerIds.has(spawner.id)) {
                            highlightNewSpawnerId = spawner.id;
                            if (highlightTimeout) clearTimeout(highlightTimeout);
                            highlightTimeout = setTimeout(() => highlightNewSpawnerId = null, 5000); // Highlight for 5 seconds
                            break; // Highlight only the first new one found
                        }
                    }
                    previousSpawnerIds = currentSpawnerIds;
                    
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
            const [statsRes, spawnersRes, versionsRes] = await Promise.all([
                fetch('/api/stats'),
                fetch('/api/spawners'),
                fetch('/api/versions')
            ]);
            if (statsRes.ok) stats.set(await statsRes.json());
            if (versionsRes.ok) serverVersions.set(await versionsRes.json());
            if (spawnersRes.ok) {
                const list: any[] = await spawnersRes.json();
                
                // Initialize previousSpawnerIds on initial fetch
                previousSpawnerIds = new Set<number>(list.map((s: any) => s.id));

                spawners.set(list);
            }
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

    function openSpawnDialog(event: CustomEvent<number>) {
        spawnTargetId = event.detail;
        isSpawnDialogOpen = true;
    }

    function openStartInstanceDialog(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        instanceActionType = 'start';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.instanceId;
        instanceActionDialogTitle = 'Start Instance';
        instanceActionDialogMessage = `Are you sure you want to start instance "${event.detail.instanceId}" on Spawner #${event.detail.spawnerId}?`;
        instanceActionConfirmText = 'Start Instance';
        isInstanceActionDialogOpen = true;
    }

    function openRestartInstanceDialog(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        instanceActionType = 'restart';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.instanceId;
        instanceActionDialogTitle = 'Restart Instance';
        instanceActionDialogMessage = `Are you sure you want to restart instance "${event.detail.instanceId}"?`;
        instanceActionConfirmText = 'Restart';
        isInstanceActionDialogOpen = true;
    }

    function openStopInstanceDialog(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        instanceActionType = 'stop';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.instanceId;
        instanceActionDialogTitle = 'Stop Instance';
        instanceActionDialogMessage = `Are you sure you want to stop instance "${event.detail.instanceId}" on Spawner #${event.detail.spawnerId}?`;
        instanceActionConfirmText = 'Stop Instance';
        isInstanceActionDialogOpen = true;
    }

    function openDeleteInstanceDialog(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        instanceActionType = 'delete';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.instanceId;
        instanceActionDialogTitle = 'Delete Instance';
        instanceActionDialogMessage = `Are you sure you want to PERMANENTLY DELETE instance "${event.detail.instanceId}" from Spawner #${event.detail.spawnerId}? ALL FILES WILL BE LOST.`;
        instanceActionConfirmText = 'Delete Instance';
        isInstanceActionDialogOpen = true;
    }

    function openUpdateInstanceDialog(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        instanceActionType = 'update';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.instanceId;
        instanceActionDialogTitle = 'Update Instance';
        instanceActionDialogMessage = `Are you sure you want to reinstall game server files for instance "${event.detail.instanceId}"? This will stop the server if it's running.`;
        instanceActionConfirmText = 'Update Instance';
        isInstanceActionDialogOpen = true;
    }

    function openRenameInstanceDialog(event: CustomEvent<{ spawnerId: number, oldId: string, newId: string }>) {
        instanceActionType = 'rename';
        instanceActionSpawnerId = event.detail.spawnerId;
        instanceActionInstanceId = event.detail.oldId;
        instanceActionNewID = event.detail.newId;
        instanceActionDialogTitle = 'Rename Instance';
        instanceActionDialogMessage = `Are you sure you want to rename instance "${event.detail.oldId}" to "${event.detail.newId}"? This will stop the server if it's running.`;
        instanceActionConfirmText = 'Rename Instance';
        isInstanceActionDialogOpen = true;
    }

    function openUpdateSpawnerBuildDialog(event: CustomEvent<number>) {
        instanceActionType = 'update_spawner_build';
        instanceActionSpawnerId = event.detail;
        instanceActionDialogTitle = 'Update Spawner Build';
        instanceActionDialogMessage = `Are you sure you want Spawner #${instanceActionSpawnerId} to download the latest game server build? This might take a while.`;
        instanceActionConfirmText = 'Update Build';
        isInstanceActionDialogOpen = true;
    }

    function openBulkActionDialog(event: CustomEvent<{ action: 'stop' | 'restart' | 'update', spawnerId: number, instanceIds: string[] }>) {
        const { action, spawnerId, instanceIds } = event.detail;
        console.log('openBulkActionDialog:', action, spawnerId, instanceIds);
        instanceActionType = `bulk_${action}` as any;
        instanceActionSpawnerId = spawnerId;
        instanceActionBulkIds = instanceIds;
        
        const actionName = action.charAt(0).toUpperCase() + action.slice(1);
        instanceActionDialogTitle = `${actionName} All Instances`;
        instanceActionDialogMessage = `Are you sure you want to ${action} ${instanceIds.length} instances on Spawner #${spawnerId}?`;
        instanceActionConfirmText = `${actionName} All`;
        isInstanceActionDialogOpen = true;
    }

    async function executeSpawn() {
        if (!spawnTargetId) return;
        
        const res = await fetch(`/api/spawners/${spawnTargetId}/spawn`, { method: 'POST' });
        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || `Server returned ${res.status}`);
        }
        
        // Open console for the new instance
        const instance = await res.json();
        consoleSpawnerId = spawnTargetId;
        consoleInstanceId = instance.id;
        isConsoleOpen = true;
    }

    async function executeInstanceAction() {
        console.log('executeInstanceAction called', { spawnerId: instanceActionSpawnerId, instanceId: instanceActionInstanceId, type: instanceActionType });
        if (!instanceActionSpawnerId || !instanceActionType) {
            console.warn('executeInstanceAction returned early: missing spawnerId or type');
            return;
        }
        // instanceActionInstanceId is only required for single instance actions
        if (!instanceActionInstanceId && !instanceActionType.startsWith('bulk_')) {
            console.warn('executeInstanceAction returned early: missing instanceId for non-bulk action');
            return;
        }

        let res: Response;
        try {
            if (instanceActionType === 'start') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/start`, { method: 'POST' });
            } else if (instanceActionType === 'stop') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/stop`, { method: 'POST' });
            } else if (instanceActionType === 'delete') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}`, { method: 'DELETE' });
            } else if (instanceActionType === 'update') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/update`, { method: 'POST' });
            } else if (instanceActionType === 'rename') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/rename`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ new_id: instanceActionNewID })
                });
            } else if (instanceActionType === 'restart') {
                // Stop first
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/stop`, { method: 'POST' });
                if (!res.ok) throw new Error('Failed to stop instance during restart');
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${instanceActionInstanceId}/start`, { method: 'POST' });
            } else if (instanceActionType === 'update_spawner_build') {
                res = await fetch(`/api/spawners/${instanceActionSpawnerId}/update-template`, { method: 'POST' });
            } else if (instanceActionType === 'bulk_stop' || instanceActionType === 'bulk_restart' || instanceActionType === 'bulk_update' || instanceActionType === 'bulk_start') {
                console.log(`Executing bulk action: ${instanceActionType} for spawner ${instanceActionSpawnerId} on instances:`, instanceActionBulkIds);
                
                let failureCount = 0;
                // Execute in parallel
                const promises = instanceActionBulkIds.map(async (id) => {
                    try {
                        let resUrl = '';
                        let resMethod = 'POST';
                        if (instanceActionType === 'bulk_stop') {
                            resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/stop`;
                        } else if (instanceActionType === 'bulk_start') {
                            resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/start`;
                        } else if (instanceActionType === 'bulk_restart') {
                            console.log(`Bulk Restart: Stopping instance ${id}`);
                            await fetch(`/api/spawners/${instanceActionSpawnerId}/instances/${id}/stop`, { method: 'POST' });
                            resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/start`;
                        } else if (instanceActionType === 'bulk_update') {
                            resUrl = `/api/spawners/${instanceActionSpawnerId}/instances/${id}/update`;
                        }
                        
                        console.log(`Bulk action: ${instanceActionType} -> fetching ${resUrl} with method ${resMethod}`);
                        const response = await fetch(resUrl, { method: resMethod });
                        if (!response.ok) {
                            const errJson = await response.json();
                            throw new Error(errJson.error || `Server returned ${response.status}`);
                        }
                    } catch (e: any) {
                        console.error(`Failed bulk action on instance ${id}: ${e.message}`, e);
                        failureCount++;
                    }
                });
                await Promise.all(promises);
                console.log('All bulk actions dispatched.');
                
                if (failureCount > 0) {
                    alert(`${failureCount} actions failed. Check browser console for details.`);
                }

                // Return fake success response since individual errors are logged
                res = { ok: true } as Response;
            } else {
                throw new Error('Unknown action type');
            }

            if (!res.ok) {
                let errorBody = `Server returned ${res.status}`;
                try {
                    const errJson = await res.json();
                    errorBody = errJson.error || JSON.stringify(errJson);
                } catch (jsonErr) {
                    // If parsing as JSON fails, try to get plain text or use a default message
                    try {
                        const errText = await res.text();
                        if (errText) errorBody = errText;
                    } catch (textErr) {
                        // Fallback if even text is unreadable
                    }
                }
                throw new Error(errorBody);
            }
            // Refresh spawners and instances after successful action
            await initialFetch();
            if (instanceActionSpawnerId) {
                spawnerTableComponent?.refreshSpawner(instanceActionSpawnerId);
            }
        } catch (e: any) {
            console.error(`Error ${instanceActionType}ing instance:`, e);
            alert(`Failed to ${instanceActionType} instance: ${e.message || 'Unknown error'}`);
        }
        isInstanceActionDialogOpen = false; // Close dialog on completion
    }


    function handleViewLogs(event: CustomEvent<number>) {
        selectedSpawnerId = event.detail;
        isLogDrawerOpen = true;
    }

    function handleTail(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        consoleSpawnerId = event.detail.spawnerId;
        consoleInstanceId = event.detail.instanceId;
        isConsoleOpen = true;
    }
</script>

<!-- Modern Header -->
<div class="relative mb-8">
    <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div class="space-y-3">
            <div class="flex items-center gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="9" y1="9" x2="15" y2="15"></line>
                        <line x1="15" y1="9" x2="9" y2="15"></line>
                    </svg>
                </div>
                <div>
                    <h1 class="text-3xl font-bold bg-gradient-to-r from-slate-100 via-slate-200 to-slate-100 bg-clip-text text-transparent">System Dashboard</h1>
                    <p class="text-slate-400 text-sm font-medium">Monitor and manage your game server infrastructure</p>
                </div>
            </div>

            <!-- Connection Status -->
            <div class="flex items-center gap-3">
                <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-800/50 border border-slate-700/50 backdrop-blur-sm">
                    <div class={`w-2 h-2 rounded-full transition-all duration-300 ${isConnected ? 'bg-emerald-400 shadow-lg shadow-emerald-400/50 animate-pulse' : 'bg-red-400 shadow-lg shadow-red-400/50'}`}></div>
                    <span class="text-xs font-semibold text-slate-300">{connectionStatus}</span>
                </div>
                <div class="text-xs text-slate-500 font-mono bg-slate-800/30 px-2 py-1 rounded-md border border-slate-700/30">
                    {new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                </div>
            </div>
        </div>
    </div>
</div>

<!-- Primary Stats Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
    <StatsCard
        title="System Uptime"
        value={formatUptime($stats.uptime)}
        Icon={Clock}
        color="blue"
    />
    <StatsCard
        title="Active Spawners"
        value={$stats.active_spawners}
        Icon={Server}
        color="emerald"
    />
    <StatsCard
        title="Total Requests"
        value={$stats.total_requests}
        Icon={Activity}
        color="purple"
    />
    <a href="/errors" class="block group">
        <StatsCard
            title="System Errors"
            value={$stats.total_errors}
            Icon={AlertCircle}
            color="red"
        />
    </a>
</div>

<!-- Secondary Stats & Resources -->
<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
    <div class="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Network Traffic Card -->
        <div class="group relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800/60 to-slate-900/60 border border-slate-700/50 backdrop-blur-sm p-6 hover:border-slate-600/50 transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-cyan-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
                <div class="flex items-center justify-between mb-4">
                    <div class="w-10 h-10 bg-gradient-to-br from-orange-500 to-orange-600 rounded-xl flex items-center justify-center shadow-lg">
                        <Network class="w-5 h-5 text-white" />
                    </div>
                    <div class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Network</div>
                </div>
                <h3 class="text-lg font-bold text-slate-200 mb-2">Traffic Overview</h3>
                <div class="space-y-2">
                    <div class="flex items-center justify-between text-sm">
                        <span class="text-slate-400">Upload</span>
                        <span class="text-orange-400 font-mono font-semibold">{formatBytes($stats.bytes_sent)}</span>
                    </div>
                    <div class="flex items-center justify-between text-sm">
                        <span class="text-slate-400">Download</span>
                        <span class="text-cyan-400 font-mono font-semibold">{formatBytes($stats.bytes_received)}</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Database Status Card -->
        <div class="group relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800/60 to-slate-900/60 border border-slate-700/50 backdrop-blur-sm p-6 hover:border-slate-600/50 transition-all duration-300">
            <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-red-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div class="relative z-10">
                <div class="flex items-center justify-between mb-4">
                    <div class={`w-10 h-10 rounded-xl flex items-center justify-center shadow-lg ${
                        $stats.db_connected
                            ? 'bg-gradient-to-br from-emerald-500 to-emerald-600'
                            : 'bg-gradient-to-br from-red-500 to-red-600'
                    }`}>
                        <Database class={`w-5 h-5 text-white`} />
                    </div>
                    <div class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Database</div>
                </div>
                <h3 class="text-lg font-bold text-slate-200 mb-2">Connection Status</h3>
                <div class="flex items-center gap-2">
                    <div class={`w-3 h-3 rounded-full ${
                        $stats.db_connected ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'
                    }`}></div>
                    <span class={`text-sm font-semibold ${
                        $stats.db_connected ? 'text-emerald-400' : 'text-red-400'
                    }`}>
                        {$stats.db_connected ? 'Connected' : 'Disconnected'}
                    </span>
                </div>
            </div>
        </div>
    </div>

    <!-- Top Resource Consumers -->
    <div class="lg:col-span-1">
        <TopResourceConsumers limit={5} compact={true} resourceType="cpu" />
    </div>
</div>

<!-- Spawners Management Section -->
<div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-800/40 via-slate-900/40 to-slate-800/40 border border-slate-700/50 backdrop-blur-sm shadow-2xl">
    <!-- Section Header -->
    <div class="relative px-8 py-6 border-b border-slate-700/50 bg-gradient-to-r from-slate-800/60 to-slate-900/60 backdrop-blur-sm">
        <div class="flex items-center justify-between">
            <div class="flex items-center gap-4">
                <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                    </svg>
                </div>
                <div>
                    <h2 class="text-xl font-bold text-slate-100">Spawner Infrastructure</h2>
                    <p class="text-sm text-slate-400">Manage your distributed game server network</p>
                </div>
            </div>
            <div class="flex items-center gap-3">
                <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
                    <div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
                    <span class="text-xs font-semibold text-emerald-400">Live Updates</span>
                </div>
                <div class="text-xs text-slate-500 bg-slate-800/50 px-2 py-1 rounded-md border border-slate-700/50">
                    {$spawners.length} registered
                </div>
            </div>
        </div>
    </div>

    <!-- Spawner Table -->
    <div class="p-0">
        <SpawnerTable
            bind:this={spawnerTableComponent}
            spawners={$spawners}
            on:spawn={openSpawnDialog}
            on:viewLogs={handleViewLogs}
            on:startInstanceRequest={openStartInstanceDialog}
            on:stopInstanceRequest={openStopInstanceDialog}
            on:restartInstanceRequest={openRestartInstanceDialog}
            on:deleteInstanceRequest={openDeleteInstanceDialog}
            on:updateInstanceRequest={openUpdateInstanceDialog}
            on:renameInstanceRequest={openRenameInstanceDialog}
            on:updateSpawnerBuild={openUpdateSpawnerBuildDialog}
            on:bulkInstanceActionRequest={openBulkActionDialog}
            on:tail={handleTail}
            highlightNewSpawnerId={highlightNewSpawnerId}
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

<!-- Instance Console Modal -->
<InstanceManagerModal
    bind:isOpen={isConsoleOpen}
    spawnerId={consoleSpawnerId}
    instanceId={consoleInstanceId}
    onClose={() => isConsoleOpen = false}
/>

<!-- Spawn Confirmation Dialog -->
<ConfirmDialog
    bind:isOpen={isSpawnDialogOpen}
    title="Spawn New Instance"
    message={`Are you sure you want to spawn a new game server instance on Spawner #${spawnTargetId}?`}
    confirmText="Spawn Server"
    onConfirm={executeSpawn}
/>

<!-- Instance Action Confirmation Dialog (Start/Stop) -->
<ConfirmDialog
    bind:isOpen={isInstanceActionDialogOpen}
    title={instanceActionDialogTitle}
    message={instanceActionDialogMessage}
    confirmText={instanceActionConfirmText}
    onConfirm={executeInstanceAction}
/>