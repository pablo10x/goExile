<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { stats, spawners, serverVersions } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import SpawnerTable from '$lib/components/SpawnerTable.svelte';
    import Drawer from '$lib/components/Drawer.svelte';
    import LogViewer from '$lib/components/LogViewer.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
    import { formatBytes, formatUptime } from '$lib/utils';
    import { Clock, Server, Activity, AlertCircle, Database, Network } from 'lucide-svelte';
    
    // Animation states
    let isLoaded = false;
    let animateStats = false;
    let mouseX = 0;
    let mouseY = 0;
    let particles = Array.from({ length: 50 }, (_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        size: Math.random() * 4 + 1,
        speedX: (Math.random() - 0.5) * 0.5,
        speedY: (Math.random() - 0.5) * 0.5
    }));

    let eventSource: EventSource | null = null;
    let isConnected = false;
    let connectionStatus = 'Connecting...';
    
    // Log Viewer State
    let isLogViewerOpen = false;
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

    // Spawner Deletion State
    let isSpawnerDeleteDialogOpen = false;
    let spawnerToDeleteId: number | null = null;

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
        
        // Mouse tracking for parallax effects
        const handleMouseMove = (e: MouseEvent) => {
            mouseX = (e.clientX / window.innerWidth - 0.5) * 2;
            mouseY = (e.clientY / window.innerHeight - 0.5) * 2;
        };
        
        window.addEventListener('mousemove', handleMouseMove);
        
        // Particle animation
        const animateParticles = () => {
            particles = particles.map(particle => ({
                ...particle,
                x: (particle.x + particle.speedX + 100) % 100,
                y: (particle.y + particle.speedY + 100) % 100
            }));
        };
        
        const interval = setInterval(animateParticles, 50);
        
        // Trigger animations after component mounts
        setTimeout(() => {
            isLoaded = true;
            setTimeout(() => {
                animateStats = true;
            }, 200);
        }, 100);
        
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            clearInterval(interval);
        };
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

    function openDeleteSpawnerDialog(event: CustomEvent<number>) {
        spawnerToDeleteId = event.detail;
        isSpawnerDeleteDialogOpen = true;
    }

    async function executeDeleteSpawner() {
        if (!spawnerToDeleteId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerToDeleteId}`, { method: 'DELETE' });
            if (!res.ok) {
                const err = await res.json().catch(() => ({}));
                throw new Error(err.error || 'Failed to delete spawner');
            }
            await initialFetch();
        } catch (e: any) {
            console.error(e);
            alert(`Failed to delete spawner: ${e.message}`);
        }
        isSpawnerDeleteDialogOpen = false;
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
        isLogViewerOpen = true;
    }

    function handleTail(event: CustomEvent<{ spawnerId: number, instanceId: string }>) {
        consoleSpawnerId = event.detail.spawnerId;
        consoleInstanceId = event.detail.instanceId;
        isConsoleOpen = true;
    }
</script>


        
        <!-- Floating Particles -->
        <div class="absolute inset-0">
            {#each Array(15) as _, i}
                <div 
                    class="absolute w-1 h-1 bg-blue-400 rounded-full opacity-30 animate-pulse"
                    style="top: {Math.random() * 100}%; left: {Math.random() * 100}%; animation-delay: {i * 0.5}s; animation-duration: {3 + Math.random() * 4}s;"
                ></div>
            {/each}
        </div>
        
        <!-- Gradient Overlay -->
        <div class="absolute inset-0 bg-gradient-to-t from-slate-900/80 via-transparent to-slate-900/60"></div>
   


<div class="flex justify-between items-center mb-6">
    <div class="transform transition-all duration-700 {isLoaded ? 'translate-x-0 opacity-100' : '-translate-x-4 opacity-0'}">
        <h1 class="text-3xl font-bold text-slate-50 bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent animate-pulse">Dashboard</h1>
        <div class="flex items-center gap-2 mt-1">
            <div class={`w-2 h-2 rounded-full ${isConnected ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'} shadow-lg`}></div>
            <span class="text-xs font-mono text-slate-400 backdrop-blur-sm">{connectionStatus}</span>
        </div>
    </div>
    <div class="text-slate-500 text-sm transform transition-all duration-700 delay-100 {isLoaded ? 'translate-x-0 opacity-100' : 'translate-x-4 opacity-0'}">
        {new Date().toLocaleDateString()}
    </div>
</div>

<!-- Stats Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
    <div class="transform transition-all duration-700 hover:scale-105 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.1s;">
        <StatsCard 
            title="Uptime" 
            value={formatUptime($stats.uptime)} 
            Icon={Clock}
            color="blue"
        />
    </div>
    <div class="transform transition-all duration-700 hover:scale-105 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.2s;">
        <StatsCard 
            title="Active Spawners" 
            value={$stats.active_spawners} 
            Icon={Server}
            color="emerald"
        />
    </div>
    <div class="transform transition-all duration-700 hover:scale-105 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.3s;">
        <StatsCard 
            title="Total Requests" 
            value={$stats.total_requests} 
            Icon={Activity}
            color="purple"
        />
    </div>
    <a href="/errors" class="block transition-all duration-700 hover:scale-[1.05] hover:rotate-1 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.4s;">
        <StatsCard 
            title="Total Errors" 
            value={$stats.total_errors} 
            Icon={AlertCircle}
            color="red"
        />
    </a>
</div>

<!-- Secondary Stats & Resources -->
<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
    <div class="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-6 h-full">
        <div class="transform transition-all duration-700 hover:scale-105 hover:-translate-y-1 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.5s;">
            <StatsCard 
                title="Network Traffic" 
                value="" 
                subValue={`<span class="text-orange-400">↑ ${formatBytes($stats.bytes_sent)}</span> <span class="text-slate-600 mx-2">|</span> <span class="text-cyan-400">↓ ${formatBytes($stats.bytes_received)}</span>`}
                Icon={Network}
                color="orange"
            />
        </div>
        <div class="transform transition-all duration-700 hover:scale-105 hover:-translate-y-1 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.6s;">
            <StatsCard 
                title="Database Status" 
                value={$stats.db_connected ? 'Connected' : 'Disconnected'} 
                Icon={Database}
                color={$stats.db_connected ? "emerald" : "red"}
            />
        </div>
    </div>
</div>

<!-- Spawners Section -->
<div class="card bg-slate-800/60 backdrop-blur-sm border border-slate-700/50 rounded-xl overflow-hidden transform transition-all duration-700 hover:scale-[1.01] hover:border-blue-500/20 hover:shadow-2xl hover:shadow-blue-500/5 {animateStats ? 'translate-y-0 opacity-100' : 'translate-y-12 opacity-0'}" style="animation-delay: 0.7s;">
    <div class="border-b border-slate-700/50 px-6 py-4 flex justify-between items-center bg-gradient-to-r from-slate-800/80 to-slate-800/60 backdrop-blur-sm">
        <div class="flex items-center gap-3">
            <div class="w-3 h-3 bg-blue-500 rounded-full animate-pulse shadow-lg shadow-blue-500/50"></div>
            <h2 class="text-xl font-bold text-slate-50">📦 Registered Spawners</h2>
        </div>
        <span class="text-xs text-slate-500 uppercase tracking-widest font-semibold bg-slate-900/80 px-3 py-1 rounded-full border border-slate-700">Real-time</span>
    </div>
    <div class="p-0 bg-slate-900/40">
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
            on:deleteSpawnerRequest={openDeleteSpawnerDialog}
            on:tail={handleTail}
            highlightNewSpawnerId={highlightNewSpawnerId}
        />
    </div>
</div>

<!-- Log Drawer -->
{#if selectedSpawnerId}
    <LogViewer
        spawnerId={selectedSpawnerId}
        isOpen={isLogViewerOpen}
        onClose={() => isLogViewerOpen = false}
    />
{/if}

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

<!-- Spawner Deletion Confirmation Dialog -->
<ConfirmDialog
    bind:isOpen={isSpawnerDeleteDialogOpen}
    title="Delete Spawner"
    message={`Are you sure you want to delete Spawner #${spawnerToDeleteId}? This will remove it from the registry. If it is still running, it might re-register.`}
    confirmText="Delete Spawner"
    isCritical={true}
    onConfirm={executeDeleteSpawner}
/>

<!-- Instance Action Confirmation Dialog (Start/Stop) -->
<ConfirmDialog
    bind:isOpen={isInstanceActionDialogOpen}
    title={instanceActionDialogTitle}
    message={instanceActionDialogMessage}
    confirmText={instanceActionConfirmText}
    onConfirm={executeInstanceAction}
/>



<style>
    @keyframes float {
        0%, 100% { transform: translateY(0px) rotate(0deg); }
        25% { transform: translateY(-20px) rotate(1deg); }
        50% { transform: translateY(0px) rotate(0deg); }
        75% { transform: translateY(-10px) rotate(-1deg); }
    }
    
    @keyframes pulse-glow {
        0%, 100% { 
            box-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
        }
        50% { 
            box-shadow: 0 0 40px rgba(59, 130, 246, 0.6);
        }
    }
    
    @keyframes slide-in-fade {
        from {
            opacity: 0;
            transform: translateY(30px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    
    @keyframes gradient-shift {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
    }
    
    @keyframes blob {
        0% { transform: translate(0px, 0px) scale(1); }
        33% { transform: translate(30px, -50px) scale(1.1); }
        66% { transform: translate(-20px, 20px) scale(0.9); }
        100% { transform: translate(0px, 0px) scale(1); }
    }
    
    .animate-float {
        animation: float 6s ease-in-out infinite;
    }
    
    .animate-pulse-glow {
        animation: pulse-glow 2s ease-in-out infinite;
    }
    
    .animate-slide-fade {
        animation: slide-in-fade 0.6s ease-out forwards;
    }
    
    .animate-gradient {
        background-size: 200% 200%;
        animation: gradient-shift 3s ease-in-out infinite;
    }
    
    .animate-gradient-shift {
        background-size: 200% 200%;
        animation: gradient-shift 8s ease infinite;
    }
    
    .animate-blob {
        animation: blob 7s infinite;
    }
    
    /* Enhanced hover effects */
    .hover-glow:hover {
        box-shadow: 0 0 30px rgba(59, 130, 246, 0.4);
        transform: translateY(-2px) scale(1.02);
        transition: all 0.3s ease;
    }
    
    /* Parallax effect on scroll */
    .parallax-bg {
        will-change: transform;
        transform: translateZ(0);
    }
    
    .bg-radial-gradient {
        background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.2) 50%, rgba(0,0,0,0.6) 100%);
    }
</style>