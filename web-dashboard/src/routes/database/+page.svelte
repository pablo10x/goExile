<script lang="ts">
	import { Database, Search, Plus, X, Trash2, Table, ChevronLeft, ChevronRight, Settings, LayoutDashboard, Terminal, Users, HardDrive, FileText, Activity, Clock } from 'lucide-svelte';
	import SchemaBrowser from '$lib/components/SchemaBrowser.svelte';
	import QueryTabs from '$lib/components/database/QueryTabs.svelte';
    import TableTab from '$lib/components/database/TableTab.svelte';
    import DatabaseBrowserTab from '$lib/components/database/DatabaseBrowserTab.svelte';
    import SQLEditorTab from '$lib/components/database/SQLEditorTab.svelte';
    import RolesTab from '$lib/components/database/RolesTab.svelte';
    import BackupsTab from '$lib/components/database/BackupsTab.svelte';
    import ConfigTab from '$lib/components/database/ConfigTab.svelte';
    import TableCreatorModal from '$lib/components/TableCreatorModal.svelte';
    import ColumnManagerModal from '$lib/components/ColumnManagerModal.svelte';
    import { notifications } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import { formatBytes } from '$lib/utils';

    // Legacy components for Overview/Roles/etc (we'll inline or reuse mostly)
    // Actually, I'll keep the logic for Overview/Roles simple here or extract if needed.
    // For now, I'll inline the Overview and keep it simple.

	// State
	let isSidebarOpen = $state(true);
    let tabs = $state<{ id: string; label: string; type: 'table' | 'sql' | 'info' | 'config' | 'roles' | 'backups' | 'browser'; data?: any }[]>([
        { id: 'overview', label: 'Overview', type: 'info' }
    ]);
    let activeTabId = $state<string>('overview');

    // Overview Data
    let dbStats = $state({ size_bytes: 0, version: '', connections: 0, uptime_seconds: 0 });

    // --- Tab Management ---

    function openTab(id: string, label: string, type: any, data: any = {}) {
        const existing = tabs.find(t => t.id === id);
        if (existing) {
            activeTabId = id;
        } else {
            tabs = [...tabs, { id, label, type, data }];
            activeTabId = id;
        }
    }

    function closeTab(id: string) {
        const idx = tabs.findIndex(t => t.id === id);
        if (idx === -1) return;
        
        const newTabs = tabs.filter(t => t.id !== id);
        tabs = newTabs;

        if (activeTabId === id) {
            // Activate neighbor
            if (newTabs.length > 0) {
                const newIdx = Math.min(idx, newTabs.length - 1);
                activeTabId = newTabs[newIdx].id;
            } else {
                activeTabId = ''; // Or reopen overview
                openTab('overview', 'Overview', 'info');
            }
        }
    }

    function handleSelectTable(schema: string, table: string) {
        openTab(`table:${schema}.${table}`, `${table}`, 'table', { schema, table });
    }

    // --- Initialization ---
    $effect(() => {
        fetch('/api/database/overview').then(r => r.json()).then(d => dbStats = d);
    });

</script>

<div class="flex h-[calc(100vh-64px)] overflow-hidden bg-slate-950 text-slate-200">
    <!-- Sidebar -->
    <div class="flex flex-col border-r border-slate-800 transition-all duration-300 {isSidebarOpen ? 'w-64' : 'w-12'}">
        <!-- System Menu -->
        <div class="p-2 space-y-1 border-b border-slate-800 bg-slate-900/50">
            <button onclick={() => openTab('overview', 'Overview', 'info')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-blue-400" title="Overview">
                <LayoutDashboard class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">Overview</span>{/if}
            </button>
            <button onclick={() => openTab('browser', 'Browser', 'browser')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-indigo-400" title="Database Browser">
                <Database class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">Browser</span>{/if}
            </button>
            <button onclick={() => openTab('sql', 'SQL Editor', 'sql')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-amber-400" title="SQL Editor">
                <Terminal class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">SQL Editor</span>{/if}
            </button>
             <button onclick={() => openTab('roles', 'Roles', 'roles')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-emerald-400" title="Roles">
                <Users class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">Roles</span>{/if}
            </button>
             <button onclick={() => openTab('backups', 'Backups', 'backups')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-orange-400" title="Backups">
                <HardDrive class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">Backups</span>{/if}
            </button>
             <button onclick={() => openTab('config', 'Config', 'config')} class="w-full flex items-center gap-2 p-2 rounded hover:bg-slate-800 text-slate-400 hover:text-purple-400" title="Config">
                <Settings class="w-4 h-4" />
                {#if isSidebarOpen}<span class="text-sm font-medium">Config</span>{/if}
            </button>
        </div>

        <!-- Schema Browser removed from sidebar, now in tab -->
        <div class="flex-1 overflow-hidden bg-slate-950"></div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 bg-slate-900">
        <QueryTabs 
            {tabs} 
            {activeTabId} 
            onSelect={(id) => activeTabId = id} 
            onClose={closeTab} 
        />

        <div class="flex-1 overflow-auto relative">
            {#each tabs as tab (tab.id)}
                <div class="absolute inset-0 bg-slate-900 {activeTabId === tab.id ? 'z-10 block' : 'z-0 hidden'}">
                    {#if tab.type === 'table'}
                        <TableTab schema={tab.data.schema} table={tab.data.table} />
                    {:else if tab.type === 'browser'}
                        <DatabaseBrowserTab 
                            onSelectTable={handleSelectTable}
                        />
                    {:else if tab.id === 'overview'}
                        <!-- Overview Content Inline -->
                        <div class="p-8">
                            <h1 class="text-2xl font-bold text-white mb-6">Database Overview</h1>
                            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                                <StatsCard title="Size" value={formatBytes(dbStats.size_bytes)} Icon={HardDrive} color="blue" />
                                <StatsCard title="Connections" value={dbStats.connections} Icon={Activity} color="emerald" />
                                <StatsCard title="Uptime" value={Math.floor(dbStats.uptime_seconds/3600) + 'h'} Icon={Clock} color="purple" />
                                <StatsCard title="Version" value={dbStats.version.split(' ')[0]} subValue={dbStats.version.split(',')[0]} Icon={Database} color="orange" />
                            </div>
                        </div>
                    {:else if tab.type === 'sql'}
                        <SQLEditorTab />
                    {:else if tab.type === 'roles'}
                        <RolesTab />
                    {:else if tab.type === 'backups'}
                        <BackupsTab />
                    {:else if tab.type === 'config'}
                        <ConfigTab />
                    {:else}
                        <div class="p-8 text-slate-500 flex flex-col items-center justify-center h-full">
                            <span style="display:none">{console.log('Unknown tab type:', tab.type, tab)}</span>
                            <FileText class="w-16 h-16 opacity-20 mb-4" />
                            <p>Feature {tab.label} is currently being upgraded.</p>
                        </div>
                    {/if}
                </div>
            {/each}
        </div>
    </div>
</div>