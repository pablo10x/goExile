<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { formatBytes } from '$lib/utils';
    import ConfirmDialog from './ConfirmDialog.svelte';
    import { serverVersions } from '$lib/stores';

    export let isOpen = false;
    export let spawnerId: number;
    export let instanceId: string | null;

    let backups: any[] = [];
    let loading = false;
    let error: string | null = null;
    let processing = false;
    let statusMessage: string | null = null;

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmAction: 'create' | 'restore' | 'delete' = 'create';
    let confirmTarget: string | null = null;

    const dispatch = createEventDispatcher();

    $: if (isOpen && instanceId) {
        fetchBackups();
    }

    $: activeVersion = $serverVersions.find(v => v.is_active);

    function getBackupVersion(filename: string): string | null {
        const match = filename.match(/_v(.*?)\.zip$/);
        return match ? match[1] : null;
    }

    function isOutdated(filename: string): boolean {
        if (!activeVersion) return false;
        const version = getBackupVersion(filename);
        if (!version) return false; // Assume fine if unknown, or maybe false? 
        return version !== activeVersion.version;
    }

    async function fetchBackups() {
        if (!instanceId) return;
        loading = true;
        error = null;
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/backups`);
            if (res.ok) {
                const data = await res.json();
                backups = data.backups || [];
                backups.sort((a: any, b: any) => new Date(b.date).getTime() - new Date(a.date).getTime());
            } else {
                const err = await res.json();
                error = err.error || 'Failed to fetch backups';
            }
        } catch (e: any) {
            error = e.message;
        } finally {
            loading = false;
        }
    }

    async function executeAction() {
        if (!instanceId) return;
        processing = true;
        
        try {
            let url = '';
            let method = 'POST';
            let body = null;

            if (confirmAction === 'create') {
                statusMessage = "Compressing files...";
                url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup`;
            } else if (confirmAction === 'delete') {
                statusMessage = "Deleting backup...";
                url = `/api/spawners/${spawnerId}/instances/${instanceId}/backup/delete`;
                body = JSON.stringify({ filename: confirmTarget });
            } else {
                statusMessage = "Restoring files...";
                url = `/api/spawners/${spawnerId}/instances/${instanceId}/restore`;
                body = JSON.stringify({ filename: confirmTarget });
            }

            const res = await fetch(url, {
                method,
                headers: body ? { 'Content-Type': 'application/json' } : undefined,
                body
            });

            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Action failed');
            }

            if (confirmAction === 'create' || confirmAction === 'delete') {
                await fetchBackups();
            } else {
                dispatch('restored');
                alert('Backup restored successfully.');
                close();
            }
        } catch (e: any) {
            throw e; 
        } finally {
            processing = false;
            statusMessage = null;
            isConfirmOpen = false;
        }
    }

    function formatDate(dateStr: string) {
        return new Date(dateStr).toLocaleString();
    }

    function close() {
        isOpen = false;
    }
    
    function getConfirmTitle() {
        switch(confirmAction) {
            case 'create': return 'Create Backup';
            case 'restore': return 'Restore Backup';
            case 'delete': return 'Delete Backup';
        }
    }

    function getConfirmMessage() {
        switch(confirmAction) {
            case 'create': return 'Are you sure you want to create a new backup? The server must be stopped.';
            case 'restore': return `Are you sure you want to restore "${confirmTarget}"?\n\n⚠️ WARNING: This will overwrite all current game files!`;
            case 'delete': return `Are you sure you want to PERMANENTLY delete "${confirmTarget}"?`;
        }
    }
    
    function getConfirmBtnText() {
        switch(confirmAction) {
            case 'create': return 'Start Backup';
            case 'restore': return 'Restore Files';
            case 'delete': return 'Delete Backup';
        }
    }
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop -->
        <div 
            class="absolute inset-0 bg-slate-950/80 backdrop-blur-md" 
            on:click={close}
            role="button"
            tabindex="0"
            aria-label="Close modal"
        ></div>

        <!-- Modal -->
        <div 
            class="relative w-full max-w-3xl bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <!-- Header -->
            <div class="px-8 py-5 border-b border-slate-700 flex justify-between items-center bg-slate-800/80 backdrop-blur-sm">
                <div>
                    <h3 class="text-xl font-bold text-slate-100 flex items-center gap-3">
                        <div class="p-2 bg-blue-500/10 rounded-lg border border-blue-500/20">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v13a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                        </div>
                        Backup Manager
                    </h3>
                    <p class="text-sm text-slate-400 mt-1 ml-12">Manage snapshots for instance <span class="font-mono text-slate-300">{instanceId}</span></p>
                </div>
                <button on:click={close} class="p-2 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
            </div>

            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-8 bg-slate-900/50">
                {#if loading}
                    <div class="flex flex-col items-center justify-center py-16 gap-4">
                        <div class="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                        <span class="text-slate-500 text-sm">Loading backups...</span>
                    </div>
                {:else if error}
                    <div class="p-6 bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl text-center">
                        <div class="text-2xl mb-2">⚠️</div>
                        <p class="font-medium">{error}</p>
                    </div>
                {:else if backups.length === 0}
                    <div class="text-center py-16 px-8 border-2 border-dashed border-slate-700/50 rounded-2xl bg-slate-800/20">
                        <div class="text-5xl mb-4 opacity-30">📦</div>
                        <h4 class="text-slate-300 font-medium mb-2">No backups found</h4>
                        <p class="text-slate-500 text-sm mb-6">Create a backup to safeguard your game server data.</p>
                        <button 
                            on:click={() => { confirmAction = 'create'; isConfirmOpen = true; }}
                            class="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-semibold transition-all shadow-lg shadow-blue-900/20 hover:scale-[1.02]"
                        >
                            Create First Backup
                        </button>
                    </div>
                {:else}
                    <div class="grid gap-4">
                        {#each backups as backup}
                            {@const outdated = isOutdated(backup.filename)}
                            {@const version = getBackupVersion(backup.filename)}
                            <div class={`relative flex items-center justify-between p-5 rounded-xl border transition-all group ${outdated ? 'bg-orange-500/5 border-orange-500/20 hover:border-orange-500/40' : 'bg-slate-800/40 border-slate-700/50 hover:border-slate-600 hover:bg-slate-800/60'}`}>
                                <div class="flex items-start gap-5">
                                    <div class={`p-3 rounded-xl flex-shrink-0 ${outdated ? 'bg-orange-500/10 text-orange-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                                    </div>
                                    <div>
                                        <div class="font-mono text-slate-200 text-sm font-semibold truncate max-w-xs sm:max-w-md" title={backup.filename}>{backup.filename}</div>
                                        <div class="flex flex-wrap items-center gap-x-4 gap-y-1 mt-2 text-xs text-slate-400">
                                            <span class="flex items-center gap-1.5">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 opacity-70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                                                {formatDate(backup.date)}
                                            </span>
                                            <span class="flex items-center gap-1.5">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 opacity-70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
                                                {formatBytes(backup.size)}
                                            </span>
                                            {#if version}
                                                <span class={`flex items-center gap-1.5 px-2 py-0.5 rounded-full font-mono font-bold ${outdated ? 'bg-orange-500/20 text-orange-400 border border-orange-500/30' : 'bg-slate-700 text-slate-300'}`}>
                                                    v{version}
                                                </span>
                                            {/if}
                                        </div>
                                    </div>
                                </div>
                                <div class="flex items-center gap-2">
                                    <button 
                                        on:click={() => { confirmTarget = backup.filename; confirmAction = 'restore'; isConfirmOpen = true; }}
                                        class="px-4 py-2 bg-slate-700 hover:bg-blue-600 hover:text-white text-slate-300 rounded-lg text-sm font-semibold transition-all shadow-sm active:scale-95"
                                    >
                                        Restore
                                    </button>
                                    <button 
                                        on:click={() => { confirmTarget = backup.filename; confirmAction = 'delete'; isConfirmOpen = true; }}
                                        class="p-2 text-slate-500 hover:bg-red-500/10 hover:text-red-400 rounded-lg transition-colors active:scale-95"
                                        title="Delete Backup"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                                    </button>
                                </div>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>

            <!-- Footer -->
            <div class="px-8 py-5 bg-slate-800/50 border-t border-slate-700 flex justify-end gap-3 backdrop-blur-sm">
                <button 
                    on:click={() => { confirmAction = 'create'; isConfirmOpen = true; }}
                    class="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-sm font-bold transition-all shadow-lg shadow-blue-900/20 hover:shadow-blue-900/40 hover:scale-[1.02] flex items-center gap-2"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v13a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                    Create New Backup
                </button>
            </div>
        </div>
    </div>
{/if}

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title={getConfirmTitle()}
    message={getConfirmMessage()}
    confirmText={getConfirmBtnText()}
    isCritical={confirmAction === 'restore' || confirmAction === 'delete'}
    onConfirm={executeAction}
    statusMessage={statusMessage}
    progress={processing ? 100 : null} 
/>