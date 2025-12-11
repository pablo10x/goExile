<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { stats, spawners, serverVersions } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import SpawnerTable from '$lib/components/SpawnerTable.svelte';
    import Drawer from '$lib/components/Drawer.svelte';
    import LogViewer from '$lib/components/LogViewer.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import InstanceConsoleModal from '$lib/components/InstanceConsoleModal.svelte';
    import { formatBytes, formatUptime } from '$lib/utils';

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
<InstanceConsoleModal
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