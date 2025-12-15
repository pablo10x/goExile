<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade, scale, fly } from 'svelte/transition';
    import { formatBytes, formatUptime } from '$lib/utils';
    import { serverVersions } from '$lib/stores';
    import Terminal from './Terminal.svelte';
    import ResourceHistoryChart from './ResourceHistoryChart.svelte';
    import ResourceMetricsPanel from './ResourceMetricsPanel.svelte';
    import ConfirmDialog from './ConfirmDialog.svelte';
    import LogViewer from './LogViewer.svelte'; // Import LogViewer for Node Logs

    export let isOpen: boolean = false;
    export let spawnerId: number | null = null;
    export let instanceId: string | null = null;
    export let onClose: () => void;
    export let memTotal: number = 0;

    let logs: string[] = [];
    
    let stats = {
        cpu_percent: 0,
        memory_usage: 0,
        disk_usage: 0,
        status: 'Unknown',
        uptime: 0
    };
    let activeTab: 'console' | 'metrics' | 'backups' | 'history' | 'node_logs' = 'console';
    
    // New State for Backups and History
    let backups: any[] = [];
    let historyLogs: any[] = [];
    let isLoadingData = false;

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmTitle = '';
    let confirmMessage = '';
    let confirmBtnText = 'Confirm';
    let isCriticalAction = false;
    let pendingBackupAction: () => Promise<void> = async () => {};

    let statsInterval: ReturnType<typeof setInterval> | null = null;
    let logsInterval: ReturnType<typeof setInterval> | null = null;
    
    // Provisioning State
    let isProvisioning = false;
    let provisioningStep = 0;
    const provisioningSteps = [
        'Allocating resources...',
        'Downloading game files...',
        'Configuring environment...',
        'Starting process...'
    ];

    $: activeVersion = $serverVersions.find(v => v.is_active);

    function getBackupVersion(filename: string): string | null {
        const match = filename.match(/_v(.*?)\.zip$/);
        return match ? match[1] : null;
    }

    function isOutdated(filename: string): boolean {
        if (!activeVersion) return false;
        const version = getBackupVersion(filename);
        if (!version) return false; 
        return version !== activeVersion.version;
    }

    $: if (isOpen && spawnerId !== null && instanceId) {
        startPolling();
        // Reset provisioning animation if just opened
        provisioningStep = 0;
    } else {
        stopPolling();
        activeTab = 'console'; // Reset tab
    }

    // Provisioning simulation logic
    $: if (stats.status === 'Provisioning') {
         isProvisioning = true;
    } else {
         isProvisioning = false;
    }

    let provTimer: any;
    $: if (isProvisioning) {
        clearInterval(provTimer);
        provTimer = setInterval(() => {
            if (provisioningStep < provisioningSteps.length - 1) provisioningStep++;
        }, 2000);
    } else {
        clearInterval(provTimer);
    }

    $: if (isOpen && activeTab === 'backups' && spawnerId && instanceId) {
        fetchBackups();
    }

    $: if (isOpen && activeTab === 'history' && spawnerId && instanceId) {
        fetchHistoryLogs();
    }

    async function fetchBackups() {
        if (!spawnerId || !instanceId) return;
        isLoadingData = true;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/backups`);
            if (res.ok) {
                const data = await res.json();
                backups = (data.backups || []).sort((a: any, b: any) => new Date(b.date).getTime() - new Date(a.date).getTime());
            }
        } catch (e) {
            console.error(e);
        } finally {
            isLoadingData = false;
        }
    }

    async function fetchHistoryLogs() {
        if (!spawnerId || !instanceId) return;
        isLoadingData = true;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/history`);
            if (res.ok) {
                historyLogs = await res.json();
            } else {
                console.error('Failed to fetch history:', res.status, res.statusText);
                historyLogs = [];
            }
        } catch (e) {
            console.error('Error fetching history:', e);
            historyLogs = [];
        } finally {
            isLoadingData = false;
        }
    }

    function handleBackupAction(action: 'create' | 'restore' | 'delete', filename?: string) {
        if (!spawnerId || !instanceId) return;
        
        if (action === 'create') {
            confirmTitle = 'Create Backup';
            confirmMessage = 'Are you sure you want to create a new backup? The server must be stopped.';
            confirmBtnText = 'Start Backup';
            isCriticalAction = false;
        } else if (action === 'restore') {
            confirmTitle = 'Restore Backup';
            confirmMessage = `Are you sure you want to restore "${filename}"?\n\n⚠️ WARNING: This will overwrite all current game files!`;
            confirmBtnText = 'Restore Files';
            isCriticalAction = true;
        } else if (action === 'delete') {
            confirmTitle = 'Delete Backup';
            confirmMessage = `Are you sure you want to PERMANENTLY delete "${filename}"?`;
            confirmBtnText = 'Delete Backup';
            isCriticalAction = true;
        }

        pendingBackupAction = async () => {
            let url = '';
            let body = null;
            if (action === 'create') url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup`;
            else if (action === 'restore') {
                url = `/api/spawners/${spawnerId}/instances/${instanceId}/restore`;
                body = JSON.stringify({ filename });
            } else if (action === 'delete') {
                url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup/delete`;
                body = JSON.stringify({ filename });
            }

            const res = await fetch(url, {
                method: 'POST',
                headers: body ? { 'Content-Type': 'application/json' } : undefined,
                body
            });

            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Action failed');
            }

            if (action !== 'restore') await fetchBackups();
            else alert('Backup restored successfully.');
        };

        isConfirmOpen = true;
    }

    async function fetchInstanceLogs() {
        if (!spawnerId || !instanceId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/logs`);
            if (res.ok) {
                const data = await res.json();
                if (data.logs) {
                    // Split logs by newline
                    logs = data.logs.split('\n');
                }
            }
        } catch (e) {
            console.error('Failed to fetch logs:', e);
        }
    }

    function startPolling() {
        stopPolling();
        logs = [];
        fetchStats();
        fetchInstanceLogs(); // Initial fetch
        
        statsInterval = setInterval(fetchStats, 2000);
        logsInterval = setInterval(() => {
            if (activeTab === 'console') {
                fetchInstanceLogs();
            }
        }, 2000); // Poll logs every 2s
    }

    function stopPolling() {
        if (statsInterval) {
            clearInterval(statsInterval);
            statsInterval = null;
        }
        if (logsInterval) {
            clearInterval(logsInterval);
            logsInterval = null;
        }
        logs = [];
    }

    async function fetchStats() {
        if (spawnerId === null || !instanceId) return;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/stats`);
            if (res.ok) {
                const data = await res.json();
                stats = { ...stats, ...data };
            }
        } catch (e) {
            console.error(e);
        }
    }

    function triggerAction(action: string) {
        confirmTitle = `${action.charAt(0).toUpperCase() + action.slice(1)} Instance`;
        confirmMessage = `Are you sure you want to ${action.toUpperCase()} this instance?`;
        confirmBtnText = action === 'delete' ? 'Delete' : 'Confirm';
        isCriticalAction = action === 'delete' || action === 'stop';
        
        pendingBackupAction = async () => {
             await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/${action}`, { method: 'POST' });
        };
        isConfirmOpen = true;
    }

    function close() {
        stopPolling();
        onClose();
    }
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop -->
                    <div 
                        class="absolute inset-0 bg-slate-950/90 backdrop-blur-md"
                        onclick={close}
                        role="button"
                        tabindex="0"
                        aria-label="Close console"
                    ></div>
        <!-- Modal Window -->
        <div 
            class="relative w-full max-w-6xl h-[85vh] flex bg-slate-900 border border-slate-700 rounded-xl shadow-2xl overflow-hidden"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <!-- Sidebar (Stats & Controls) -->
            <div class="w-72 bg-slate-950/50 border-r border-slate-800 flex flex-col shrink-0">
                <div class="p-6 border-b border-slate-800">
                    <h3 class="text-lg font-bold text-slate-100 break-all">{instanceId}</h3>
                    <div class="flex items-center gap-2 mt-2">
                        <span class={`w-2.5 h-2.5 rounded-full ${stats.status === 'Running' ? 'bg-emerald-500 animate-pulse' : (stats.status === 'Provisioning' ? 'bg-blue-500 animate-bounce' : 'bg-red-500')}`}></span>
                        <span class="text-sm font-mono text-slate-400">{stats.status || 'Unknown'}</span>
                    </div>
                </div>

                <div class="p-6 space-y-8 flex-1 overflow-y-auto">
                    <!-- Uptime -->
                    <div>
                        <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1">Uptime</div>
                        <div class="text-3xl font-mono text-slate-200 tracking-tight">{formatUptime((stats.uptime || 0) * 1000)}</div>
                    </div>

                    <!-- Resources -->
                    <div class="space-y-5">
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>CPU Usage</span>
                                <span>{stats.cpu_percent.toFixed(1)}%</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-blue-500 transition-all duration-500 ease-out" style="width: {stats.cpu_percent}%"></div>
                            </div>
                        </div>
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>Memory Usage</span>
                                <span>{formatBytes(stats.memory_usage)}</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-purple-500 transition-all duration-500 ease-out" style="width: {Math.min(100, (stats.memory_usage / (1024*1024*1024*4))*100)}%"></div> 
                            </div>
                        </div>
                        <div>
                            <div class="flex justify-between text-xs text-slate-400 mb-1.5">
                                <span>Disk Space</span>
                                <span>{formatBytes(stats.disk_usage)}</span>
                            </div>
                            <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                                <div class="h-full bg-orange-500 transition-all duration-500 ease-out" style="width: 10%"></div> 
                            </div>
                        </div>
                    </div>

                    <!-- Provisioning Steps (Only if Provisioning) -->
                    {#if isProvisioning}
                        <div class="pt-6 border-t border-slate-800 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            <div class="text-[10px] font-bold text-blue-400 uppercase tracking-wider mb-4 flex items-center gap-2">
                                <div class="w-2 h-2 rounded-full border border-blue-400 border-t-transparent animate-spin"></div>
                                Provisioning
                            </div>
                            <div class="space-y-4">
                                {#each provisioningSteps as step, i}
                                    <div class="flex items-center gap-3 text-xs transition-all duration-300">
                                        {#if i < provisioningStep}
                                            <div class="w-5 h-5 rounded-full bg-emerald-500/20 flex items-center justify-center text-emerald-500 border border-emerald-500/30">✓</div>
                                            <span class="text-slate-500 line-through decoration-slate-700">{step}</span>
                                        {:else if i === provisioningStep}
                                            <div class="w-5 h-5 rounded-full border-2 border-blue-500 border-t-transparent animate-spin"></div>
                                            <span class="text-blue-200 font-semibold">{step}</span>
                                        {:else}
                                            <div class="w-5 h-5 rounded-full border border-slate-800"></div>
                                            <span class="text-slate-700">{step}</span>
                                        {/if}
                                    </div>
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Actions -->
                <div class="p-4 border-t border-slate-800 grid grid-cols-2 gap-3 bg-slate-900/50">
                    <button 
                        onclick={() => triggerAction('start')}
                        disabled={stats.status === 'Running' || stats.status === 'Provisioning'}
                        class="col-span-2 px-4 py-2.5 bg-slate-800 hover:bg-emerald-600/20 hover:text-emerald-400 text-slate-400 rounded-lg text-xs font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-700 hover:border-emerald-500/30"
                    >
                        Start
                    </button>
                    <button 
                        onclick={() => triggerAction('restart')}
                        disabled={stats.status !== 'Running'}
                        class="px-4 py-2.5 bg-slate-800 hover:bg-blue-600/20 hover:text-blue-400 text-slate-400 rounded-lg text-xs font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-700 hover:border-blue-500/30"
                    >
                        Restart
                    </button>
                    <button 
                        onclick={() => triggerAction('stop')}
                        disabled={stats.status !== 'Running'}
                        class="px-4 py-2.5 bg-slate-800 hover:bg-red-600/20 hover:text-red-400 text-slate-400 rounded-lg text-xs font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed border border-slate-700 hover:border-red-500/30"
                    >
                        Stop
                    </button>
                </div>
            </div>

            <!-- Main Area -->
            <div class="flex-1 flex flex-col min-w-0 bg-black relative">
                <!-- Tabs -->
                <div class="flex border-b border-slate-800 bg-slate-900 shrink-0">
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='console' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        onclick={() => activeTab = 'console'}
                    >
                        Console
                    </button>
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='metrics' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        onclick={() => activeTab = 'metrics'}
                    >
                        Metrics
                    </button>
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='backups' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        onclick={() => activeTab = 'backups'}
                    >
                        Backups
                    </button>
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='history' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        onclick={() => activeTab = 'history'}
                    >
                        History
                    </button>
                    <button 
                        class={`px-6 py-3 text-xs font-bold uppercase tracking-wider transition-colors ${activeTab==='node_logs' ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800/50' : 'text-slate-500 hover:text-slate-300 hover:bg-slate-800/30'}`} 
                        onclick={() => activeTab = 'node_logs'}
                    >
                        Node Logs
                    </button>
                </div>

                <div class="flex-1 relative overflow-hidden">
                    {#if activeTab === 'console'}
                        <div class="absolute inset-0 p-0">
                            <!-- Live Indicator -->
                            <div class="absolute top-4 right-4 z-10">
                                <div class="px-3 py-1 rounded-full bg-slate-800/80 backdrop-blur border border-slate-700 text-[10px] font-mono text-slate-400 flex items-center gap-2">
                                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                                    Live
                                </div>
                            </div>
                            <Terminal logs={logs} title={`root@${instanceId}:~`} />
                        </div>
                    {:else if activeTab === 'metrics'}
                        <div class="p-8 h-full overflow-y-auto bg-slate-950/50">
                            {#if spawnerId !== null && instanceId}
                                <ResourceMetricsPanel 
                                    spawnerId={spawnerId} 
                                    instanceId={instanceId} 
                                    memTotal={memTotal}
                                    height={300}
                                />
                            {:else}
                                <div class="flex items-center justify-center h-full text-slate-500">
                                    No instance selected
                                </div>
                            {/if}
                        </div>
                    {:else if activeTab === 'node_logs'}
                        <div class="h-full relative">
                            {#if spawnerId !== null}
                                <LogViewer spawnerId={spawnerId} isOpen={true} embedded={true} />
                            {/if}
                        </div>
                    {:else if activeTab === 'backups'}
                        <div class="p-8 h-full overflow-y-auto bg-slate-950/50">
                            <div class="flex justify-between items-center mb-6">
                                <h3 class="text-lg font-bold text-slate-200">Instance Backups</h3>
                                <button 
                                    onclick={() => handleBackupAction('create')}
                                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-bold transition-colors shadow-lg shadow-blue-900/20"
                                >
                                    Create Backup
                                </button>
                            </div>
                            
                            {#if isLoadingData}
                                <div class="text-center py-12 text-slate-500">Loading...</div>
                            {:else if backups.length === 0}
                                <div class="text-center py-12 text-slate-500 border-2 border-dashed border-slate-800 rounded-xl">
                                    No backups found.
                                </div>
                            {:else}
                                <div class="space-y-3">
                                    {#each backups as backup, i (backup.filename)}
                                        {@const outdated = isOutdated(backup.filename)}
                                        {@const version = getBackupVersion(backup.filename)}
                                        <div 
                                            class={`relative flex items-center justify-between p-4 rounded-xl border transition-all group ${outdated ? 'bg-orange-500/5 border-orange-500/20 hover:border-orange-500/40' : 'bg-slate-900/50 border-slate-800 hover:border-slate-700'}`}
                                            in:fly={{ y: 20, duration: 300, delay: 50 * i }}
                                        >
                                            <div class="flex items-center gap-4">
                                                <div class={`p-2 rounded-lg flex-shrink-0 ${outdated ? 'bg-orange-500/10 text-orange-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                                                </div>
                                                <div>
                                                    <div class="text-sm font-bold text-slate-200 font-mono truncate max-w-xs">{backup.filename}</div>
                                                    <div class="flex items-center gap-3 mt-1 text-xs text-slate-400">
                                                        <span>{new Date(backup.date).toLocaleString()}</span>
                                                        <span>{formatBytes(backup.size)}</span>
                                                        {#if version}
                                                            <span class={`px-1.5 py-0.5 rounded font-mono font-bold text-[10px] ${outdated ? 'bg-orange-500/20 text-orange-400 border border-orange-500/30' : 'bg-slate-700 text-slate-300'}`}>
                                                                v{version}
                                                            </span>
                                                        {/if}
                                                    </div>
                                                </div>
                                            </div>
                                            <div class="flex gap-2">
                                                <button 
                                                    onclick={() => handleBackupAction('restore', backup.filename)}
                                                    class="px-3 py-1.5 bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-400 rounded text-xs font-semibold transition-colors"
                                                >
                                                    Restore
                                                </button>
                                                <button 
                                                    onclick={() => handleBackupAction('delete', backup.filename)}
                                                    class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded transition-colors"
                                                    title="Delete"
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                                                </button>
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    {:else if activeTab === 'history'}
                        <div class="p-0 h-full overflow-y-auto bg-slate-950/50">
                            <table class="w-full text-left border-collapse">
                                <thead class="bg-slate-900 border-b border-slate-800 sticky top-0">
                                    <tr>
                                        <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider">Action</th>
                                        <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider">Status</th>
                                        <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider">Time</th>
                                        <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider">Details</th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-slate-800/50">
                                    {#each (historyLogs || []) as log}
                                        <tr class="hover:bg-slate-900/30 transition-colors">
                                            <td class="px-6 py-3">
                                                <span class="font-mono text-sm text-slate-300 bg-slate-800/50 px-2 py-0.5 rounded border border-slate-700/50">
                                                    {log.action}
                                                </span>
                                            </td>
                                            <td class="px-6 py-3">
                                                <span class={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${log.status === 'success' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'}`}>
                                                    {log.status}
                                                </span>
                                            </td>
                                            <td class="px-6 py-3 text-sm text-slate-400 font-mono">
                                                {new Date(log.timestamp).toLocaleString()}
                                            </td>
                                            <td class="px-6 py-3 text-sm text-slate-500 truncate max-w-xs" title={log.details}>
                                                {log.details || '-'}
                                            </td>
                                        </tr>
                                    {/each}
                                     {#if (historyLogs || []).length === 0 && !isLoadingData}
                                        <tr>
                                            <td colspan="4" class="px-6 py-8 text-center text-slate-500 italic">No history found</td>
                                        </tr>
                                    {/if}
                                </tbody>
                            </table>
                        </div>
                    {/if}
                </div>
            </div>
            
            <!-- Close Button (Absolute) -->
            <button 
                onclick={close}
                class="absolute top-3 right-3 p-2 text-slate-500 hover:text-white transition-colors z-20"
                aria-label="Close"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
        </div>
    </div>

    <ConfirmDialog
        bind:isOpen={isConfirmOpen}
        title={confirmTitle}
        message={confirmMessage}
        confirmText={confirmBtnText}
        isCritical={isCriticalAction}
        onConfirm={pendingBackupAction}
    />
{/if}