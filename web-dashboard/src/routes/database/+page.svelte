<script lang="ts">
    import { onMount } from 'svelte';
    import { Database, Table, Download, RefreshCw, Settings, Save, Upload, Trash2, Activity, HardDrive, Clock, FileText, Play, RotateCcw, Plus, X, Check, Pencil, AlertCircle, Terminal, Users, UserPlus, Key, Folder, Columns } from 'lucide-svelte';
    import { formatBytes } from '$lib/utils';
    import { notifications } from '$lib/stores';
    import TreeItem from '$lib/components/TreeItem.svelte';
    import StatsCard from '$lib/components/StatsCard.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import DatabaseDetailModal from '$lib/components/DatabaseDetailModal.svelte';

    let activeTab = $state('overview');
    let loading = $state(false);
    let error = $state<string | null>(null);

    // Overview Stats
    let dbStats = $state({
        size_bytes: 0,
        version: '',
        connections: 0,
        uptime_seconds: 0
    });

    // Tables & Browser
    // Schema structure: { [schema]: { tables: { [table]: columns[] }, expanded: bool } }
    // But for simplicity, we keep separate lists and expanded sets.
    let schemas = $state<string[]>([]);
    let expandedSchemas = $state<Set<string>>(new Set());
    let schemaTables = $state<Record<string, string[]>>({}); // schema -> tables
    let expandedTables = $state<Set<string>>(new Set()); // schema.table
    let tableColumns = $state<Record<string, any[]>>({}); // schema.table -> columns
    let selectedNode = $state<string | null>(null); // "schema.table" or "schema"

    let selectedSchema = $state('public'); // Default for loading view
    let newSchemaName = $state('');
    let isCreatingSchema = $state(false);
    let tables = $state<string[]>([]);
    let selectedTable = $state<string | null>(null);
    let tableData = $state<any[]>([]);
    let tableTotal = $state(0);
    let tableLimit = $state(50);
    let tableOffset = $state(0);
    let tableFilter = $state(''); // New Filter State
    let loadingTable = $state(false);

    // ... (rest of imports/state)

    async function loadTableData(table: string, offset = 0) {
        selectedTable = table;
        tableOffset = offset;
        loadingTable = true;
        addingRow = false;
        editingCell = null;
        pendingChanges = {}; // Clear pending changes on table switch/reload
        
        let url = `/api/database/table/${table}?schema=${selectedSchema}&limit=${tableLimit}&offset=${offset}`;
        if (tableFilter) {
            url += `&filter=${encodeURIComponent(tableFilter)}`;
        }

        try {
            const res = await fetch(url);
            if (res.ok) {
                const data = await res.json();
                tableData = data.data || [];
                tableTotal = data.total || 0;
            }
        } catch (e: any) {
            notifications.add({ type: 'error', message: `Failed to load data for ${table}`, details: e.message });
        } finally {
            loadingTable = false;
        }
    }

    async function runSQL() {
        if (!sqlQuery.trim()) return;
        sqlLoading = true;
        sqlResults = [];
        try {
            const res = await fetch('/api/database/sql', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: sqlQuery })
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Query failed');
            }
            sqlResults = await res.json();
            notifications.add({ type: 'success', message: 'Query executed successfully' });
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'SQL Execution Failed', details: e.message });
        } finally {
            sqlLoading = false;
        }
    }

    function saveEdit() {
        if (!selectedTable || !editingCell) return;
        
        // Optimistic update local data for display
        const idx = tableData.findIndex(r => r.id === editingCell!.rowId);
        if (idx !== -1) {
            // Track change
            const rowId = String(editingCell!.rowId);
            if (!pendingChanges[rowId]) pendingChanges[rowId] = {};
            pendingChanges[rowId][editingCell!.key] = editValue;
            
            // Update visual table data
            tableData[idx][editingCell!.key] = editValue;
        }
        editingCell = null;
    }

    async function applyChanges() {
        if (!selectedTable) return;
        
        const updates = Object.entries(pendingChanges);
        if (updates.length === 0) return;

        loadingTable = true;
        let successCount = 0;
        let errorCount = 0;
        
        try {
            for (const [rowId, changes] of updates) {
                 // Send updates one by one (or batch if API supported it)
                 // Construct body with ALL changes for that row
                 for (const [key, val] of Object.entries(changes)) {
                     const res = await fetch(`/api/database/table/${selectedTable}/${rowId}`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ [key]: val })
                    });
                    if (!res.ok) {
                        errorCount++;
                        console.error(`Failed to update row ${rowId} col ${key}`);
                    } else {
                        successCount++;
                    }
                 }
            }
            
            if (errorCount > 0) {
                notifications.add({ type: 'warning', message: `Applied ${successCount} updates with ${errorCount} errors.` });
            } else {
                notifications.add({ type: 'success', message: `Successfully applied ${successCount} updates.` });
            }

            // Reload to confirm persistence
            await loadTableData(selectedTable, tableOffset);
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to apply changes', details: e.message });
        } finally {
            loadingTable = false;
        }
    }

    async function saveNewRow() {
        if (!selectedTable) return;
        try {
            const res = await fetch(`/api/database/table/${selectedTable}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newRowData)
            });
            if (!res.ok) throw new Error('Insert failed');
            
            await loadTableData(selectedTable, tableOffset);
            addingRow = false;
            newRowData = {};
            notifications.add({ type: 'success', message: 'Row added successfully' });
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to add row', details: e.message });
        }
    }

    async function deleteRow(id: any) {
        if (!confirm('Are you sure you want to delete this row?')) return;
        if (!selectedTable) return;
        try {
            const res = await fetch(`/api/database/table/${selectedTable}/${id}`, { method: 'DELETE' });
            if (!res.ok) throw new Error('Delete failed');
            await loadTableData(selectedTable, tableOffset);
            notifications.add({ type: 'success', message: 'Row deleted successfully' });
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to delete row', details: e.message });
        }
    }

    function startEdit(row: any, key: string) {
        if (key === 'id') return; // Prevent editing ID
        editingCell = { rowId: row.id, key };
        editValue = String(row[key]);
    }

    async function loadBackups() {
        try {
            const res = await fetch('/api/database/backups');
            if (res.ok) backups = await res.json();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to load backups', details: e.message });
        }
    }

    async function createBackup() {
        try {
            loading = true;
            const res = await fetch('/api/database/backups', { method: 'POST' });
            if (!res.ok) throw new Error('Backup failed');
            await loadBackups();
            notifications.add({ type: 'success', message: 'Backup created successfully' });
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to create backup', details: e.message });
        } finally {
            loading = false;
        }
    }

    async function deleteBackup(filename: string) {
        confirmTitle = 'Delete Backup';
        confirmMessage = `Are you sure you want to delete backup ${filename}?`;
        confirmIsCritical = true;
        confirmAction = async () => {
            try {
                const res = await fetch(`/api/database/backups/${filename}`, { method: 'DELETE' });
                if (!res.ok) throw new Error('Delete failed');
                await loadBackups();
                notifications.add({ type: 'success', message: 'Backup deleted successfully' });
            } catch (e: any) {
                notifications.add({ type: 'error', message: 'Failed to delete backup', details: e.message });
            }
        };
        isConfirmOpen = true;
    }

    async function restoreBackup(filename: string) {
        confirmTitle = 'Restore Backup';
        confirmMessage = `WARNING: This will overwrite the current database with ${filename}. Are you sure?`;
        confirmIsCritical = true;
        confirmAction = async () => {
            loading = true;
            try {
                const res = await fetch('/api/database/restore', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ filename })
                });
                if (!res.ok) throw new Error('Restore failed');
                notifications.add({ type: 'success', message: 'Database restored successfully. Reloading...' });
                setTimeout(() => window.location.reload(), 2000);
            } catch (e: any) {
                notifications.add({ type: 'error', message: 'Failed to restore database', details: e.message });
            } finally {
                loading = false;
            }
        };
        isConfirmOpen = true;
    }

    async function loadConfig() {
        try {
            const res = await fetch('/api/database/config');
            if (res.ok) pgConfig = await res.json();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to load configuration', details: e.message });
        }
    }

    function startConfigEdit(cfg: any) {
        editingConfig = { name: cfg.name };
        editConfigValue = cfg.setting;
    }

    function saveConfigEdit() {
        if (!editingConfig) return;
        
        // Optimistic update
        const idx = pgConfig.findIndex(c => c.name === editingConfig!.name);
        if (idx !== -1) {
            // Track change
            const name = editingConfig!.name;
            pendingConfigChanges[name] = editConfigValue;
            
            // Update visual data
            pgConfig[idx].setting = editConfigValue;
        }
        editingConfig = null;
    }

    async function applyConfigChanges() {
        loading = true;
        try {
            // 1. Apply all changes
            for (const [name, value] of Object.entries(pendingConfigChanges)) {
                const res = await fetch('/api/database/config', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, value })
                });
                if (!res.ok) {
                    const errData = await res.json().catch(() => ({}));
                    throw new Error(errData.error || `Failed to update ${name}: Server returned ${res.status}`);
                }
            }

            // 2. Restart/Reload Postgres
            const restartRes = await fetch('/api/database/config/restart', { method: 'POST' });
            if (!restartRes.ok) throw new Error('Failed to reload configuration');

            notifications.add({ type: 'success', message: 'Configuration updated and reloaded successfully.' });
            pendingConfigChanges = {};
            isConfigConfirmOpen = false;
            await loadConfig(); // Reload from server to be sure
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to apply configuration changes', details: e.message });
        } finally {
            loading = false;
        }
    }

    function refresh() {
        if (activeTab === 'overview') loadOverview();
        if (activeTab === 'tables') {
            loadSchemas();
            loadTables();
        }
        if (activeTab === 'roles') loadRoles();
        if (activeTab === 'backups') loadBackups();
        if (activeTab === 'config') loadConfig();
    }

    $effect(() => {
        refresh();
    });

    function downloadFile(filename: string) {
        window.open(`/api/database/backups/${filename}`, '_blank');
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold text-slate-100 mb-2">Database Management</h1>
            <p class="text-slate-400">PostgreSQL Administration & Maintenance</p>
        </div>
        <button 
            onclick={refresh}
            class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded-lg transition-colors"
        >
            <RefreshCw class={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
        </button>
    </div>

    <!-- Tabs -->
    <div class="border-b border-slate-700">
        <nav class="-mb-px flex space-x-8">
            {#each ['overview', 'tables', 'roles', 'sql', 'backups', 'config'] as tab}
                <button
                    onclick={() => activeTab = tab}
                    class="{activeTab === tab ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm capitalize transition-colors"
                >
                    {tab === 'sql' ? 'SQL Editor' : tab === 'roles' ? 'Role Manager' : tab}
                </button>
            {/each}
        </nav>
    </div>

    {#if activeTab === 'overview'}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <StatsCard title="Database Size" value={formatBytes(dbStats.size_bytes)} Icon={HardDrive} color="blue" />
            <StatsCard title="Active Connections" value={dbStats.connections} Icon={Activity} color="emerald" />
            <StatsCard title="Uptime" value={Math.floor(dbStats.uptime_seconds / 3600) + 'h ' + Math.floor((dbStats.uptime_seconds % 3600) / 60) + 'm'} Icon={Clock} color="purple" />
            <StatsCard title="Version" value={dbStats.version.split(' ')[0] + ' ' + dbStats.version.split(' ')[1]} subValue={dbStats.version.split(',')[0]} Icon={Database} color="orange" />
        </div>
    {/if}

    {#if activeTab === 'tables'}
        <div class="grid grid-cols-1 lg:grid-cols-4 gap-6 h-[600px]">
            <!-- Sidebar -->
            <div class="lg:col-span-1 bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden flex flex-col">
                <div class="p-4 border-b border-slate-700 font-semibold text-slate-300 bg-slate-900/50 flex justify-between items-center">
                    <span class="flex items-center gap-2"><Database class="w-4 h-4 text-purple-400" /> Browser</span>
                    <button onclick={() => isCreatingSchema = true} class="text-emerald-400 hover:text-emerald-300" title="Create Schema">
                        <Plus class="w-4 h-4" />
                    </button>
                </div>
                
                {#if isCreatingSchema}
                    <div class="p-2 border-b border-slate-700 bg-slate-900/30 flex gap-1">
                        <input 
                            type="text" 
                            bind:value={newSchemaName}
                            placeholder="Schema Name"
                            class="w-full bg-slate-900 border border-slate-600 rounded px-2 py-1 text-xs text-slate-200 outline-none"
                            onkeydown={(e) => e.key === 'Enter' && createSchema()}
                        />
                        <button onclick={createSchema} class="p-1 bg-emerald-600 text-white rounded"><Check class="w-3 h-3"/></button>
                        <button onclick={() => isCreatingSchema = false} class="p-1 bg-slate-700 text-white rounded"><X class="w-3 h-3"/></button>
                    </div>
                {/if}

                <div class="flex-1 overflow-y-auto p-2 space-y-0.5">
                    {#each schemas as schema}
                        <TreeItem 
                            label={schema} 
                            type="schema"
                            isOpen={expandedSchemas.has(schema)}
                            onToggle={() => toggleSchema(schema)}
                        >
                            {#if schemaTables[schema]}
                                {#each schemaTables[schema] as table}
                                    <TreeItem 
                                        label={table} 
                                        type="table"
                                        isOpen={expandedTables.has(`${schema}.${table}`)}
                                        isSelected={selectedTable === table && selectedSchema === schema}
                                        onToggle={() => toggleTable(schema, table)}
                                        onSelect={() => toggleTable(schema, table)}
                                    >
                                        {#if tableColumns[`${schema}.${table}`]}
                                            {#each tableColumns[`${schema}.${table}`] as col}
                                                <TreeItem 
                                                    label={`${col.name} (${col.type})`} 
                                                    type="column"
                                                />
                                            {/each}
                                        {/if}
                                    </TreeItem>
                                {/each}
                            {:else}
                                <div class="text-xs text-slate-500 py-1 pl-6">Loading...</div>
                            {/if}
                        </TreeItem>
                    {/each}
                </div>
            </div>

            <!-- Data Grid -->
            <div class="lg:col-span-3 bg-slate-800/50 border border-slate-700 rounded-xl flex flex-col overflow-hidden">
                {#if selectedTable}
                    <div class="p-4 border-b border-slate-700 flex justify-between items-center bg-slate-900/50">
                        <h3 class="font-bold text-slate-200">{selectedTable} <span class="text-slate-500 font-normal ml-2">({tableTotal} rows)</span></h3>
                        <div class="flex gap-2 items-center">
                            <input 
                                type="text" 
                                bind:value={tableFilter} 
                                oninput={() => loadTableData(selectedTable!, 0)}
                                placeholder="Filter data..." 
                                class="px-3 py-1 bg-slate-900 border border-slate-700 rounded text-xs text-slate-200 focus:outline-none focus:border-blue-500"
                            />
                            {#if Object.keys(pendingChanges).length > 0}
                                <button 
                                    onclick={applyChanges}
                                    class="flex items-center gap-1 px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs animate-pulse"
                                >
                                    <Save class="w-3 h-3" /> Apply {Object.keys(pendingChanges).length} Updates
                                </button>
                                <div class="h-4 w-px bg-slate-700 mx-2"></div>
                            {/if}
                            <button 
                                onclick={() => addingRow = true}
                                disabled={addingRow}
                                class="flex items-center gap-1 px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-xs disabled:opacity-50"
                            >
                                <Plus class="w-3 h-3" /> Add Row
                            </button>
                            <div class="h-4 w-px bg-slate-700 mx-2"></div>
                            <button 
                                disabled={tableOffset === 0}
                                onclick={() => loadTableData(selectedTable!, tableOffset - tableLimit)}
                                class="px-3 py-1 bg-slate-700 rounded text-xs disabled:opacity-50"
                            >Prev</button>
                            <span class="text-xs text-slate-400 py-1">{tableOffset + 1} - {Math.min(tableOffset + tableLimit, tableTotal)}</span>
                            <button 
                                disabled={tableOffset + tableLimit >= tableTotal}
                                onclick={() => loadTableData(selectedTable!, tableOffset + tableLimit)}
                                class="px-3 py-1 bg-slate-700 rounded text-xs disabled:opacity-50"
                            >Next</button>
                        </div>
                    </div>
                    <div class="flex-1 overflow-auto">
                        {#if loadingTable}
                            <div class="flex items-center justify-center h-full text-slate-500">Loading...</div>
                        {:else if tableData.length === 0 && !addingRow}
                            <div class="flex items-center justify-center h-full text-slate-500">No data found</div>
                        {:else}
                            <table class="w-full text-left text-xs">
                                <thead class="bg-slate-900/50 text-slate-400 sticky top-0">
                                    <tr>
                                        {#if tableData.length > 0}
                                            {#each Object.keys(tableData[0]) as key}
                                                <th class="px-4 py-3 font-medium border-b border-slate-700">{key}</th>
                                            {/each}
                                            <th class="px-4 py-3 font-medium border-b border-slate-700 w-10"></th>
                                        {:else if addingRow}
                                            <!-- If no data but adding row, need to guess schema or just show generic -->
                                            <th class="px-4 py-3 font-medium border-b border-slate-700">New Row</th>
                                        {/if}
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-slate-700/50">
                                    {#if addingRow}
                                        <tr class="bg-emerald-900/20">
                                            {#if tableData.length > 0}
                                                {#each Object.keys(tableData[0]) as key}
                                                    <td class="px-4 py-2">
                                                        {#if key !== 'id'}
                                                            <input 
                                                                type="text" 
                                                                bind:value={newRowData[key]}
                                                                class="w-full bg-slate-900 border border-slate-600 rounded px-2 py-1 text-slate-200 focus:border-emerald-500 outline-none"
                                                                placeholder={key}
                                                            />
                                                        {:else}
                                                            <span class="text-slate-500 italic">Auto</span>
                                                        {/if}
                                                    </td>
                                                {/each}
                                                <td class="px-4 py-2 flex gap-1">
                                                    <button onclick={saveNewRow} class="p-1 bg-emerald-600 text-white rounded hover:bg-emerald-500"><Check class="w-3 h-3" /></button>
                                                    <button onclick={() => addingRow = false} class="p-1 bg-slate-700 text-white rounded hover:bg-slate-600"><X class="w-3 h-3" /></button>
                                                </td>
                                            {:else}
                                                <td class="p-4 text-center text-slate-500">
                                                    Cannot add row to empty table without schema knowledge yet.
                                                    <button onclick={() => addingRow = false} class="ml-2 text-blue-400">Cancel</button>
                                                </td>
                                            {/if}
                                        </tr>
                                    {/if}

                                    {#each tableData as row}
                                        <tr class="hover:bg-slate-700/30 group {pendingChanges[String(row.id)] ? 'bg-emerald-900/10 border-l-2 border-l-emerald-500' : ''}">
                                            {#each Object.keys(row) as key}
                                                <td 
                                                    class="px-4 py-2 text-slate-300 max-w-xs truncate cursor-pointer hover:bg-slate-700/50 {pendingChanges[String(row.id)]?.[key] !== undefined ? 'text-emerald-300 font-medium' : ''}" 
                                                    title={String(row[key])}
                                                    onclick={() => startEdit(row, key)}
                                                >
                                                    {#if editingCell && editingCell.rowId === row.id && editingCell.key === key}
                                                        <!-- svelte-ignore a11y_autofocus -->
                                                        <input 
                                                            type="text" 
                                                            bind:value={editValue}
                                                            onblur={saveEdit}
                                                            onkeydown={(e) => e.key === 'Enter' && saveEdit()}
                                                            class="w-full bg-slate-900 border border-blue-500 rounded px-1 text-slate-200 outline-none"
                                                            autofocus
                                                        />
                                                    {:else}
                                                        {String(row[key])}
                                                    {/if}
                                                </td>
                                            {/each}
                                            <td class="px-4 py-2 text-right">
                                                <button 
                                                    onclick={() => deleteRow(row.id)}
                                                    class="p-1 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded opacity-0 group-hover:opacity-100 transition-opacity"
                                                >
                                                    <Trash2 class="w-3.5 h-3.5" />
                                                </button>
                                            </td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        {/if}
                    </div>
                {:else}
                    <div class="flex items-center justify-center h-full text-slate-500">Select a table to view data</div>
                {/if}
            </div>
        </div>
    {/if}

    {#if activeTab === 'roles'}
        <div class="space-y-6">
            <div class="flex justify-between items-center">
                <h3 class="text-xl font-bold text-slate-200">Database Roles</h3>
                <button 
                    onclick={() => isCreatingRole = true}
                    disabled={isCreatingRole}
                    class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-bold flex items-center gap-2 shadow-lg disabled:opacity-50"
                >
                    <UserPlus class="w-4 h-4" /> Create Role
                </button>
            </div>

            {#if isCreatingRole}
                <div class="bg-slate-800/80 border border-slate-700 rounded-xl p-6 animate-in slide-in-from-top-4 shadow-2xl">
                    <div class="flex justify-between items-start mb-6">
                        <div>
                            <h4 class="text-lg font-bold text-slate-200">Create New Role</h4>
                            <p class="text-sm text-slate-400">Define privileges and access controls for the new database role.</p>
                        </div>
                        <div class="p-2 bg-emerald-500/10 rounded-lg">
                            <UserPlus class="w-6 h-6 text-emerald-400" />
                        </div>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                        <div class="space-y-4">
                            <h5 class="text-xs font-bold text-slate-500 uppercase tracking-wider border-b border-slate-700 pb-2">Credentials</h5>
                            <div>
                                <label class="block text-xs font-medium text-slate-300 mb-1.5">Role Name</label>
                                <input type="text" bind:value={newRole.name} placeholder="e.g. app_user" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-slate-200 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all" />
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-slate-300 mb-1.5">Password</label>
                                <input type="password" bind:value={newRole.password} placeholder="••••••••" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-slate-200 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all" />
                            </div>
                            <!-- Future: Valid Until / Connection Limit -->
                        </div>

                        <div class="space-y-4">
                            <h5 class="text-xs font-bold text-slate-500 uppercase tracking-wider border-b border-slate-700 pb-2">Privileges</h5>
                            <div class="grid grid-cols-2 gap-3">
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.canLogin} class="rounded bg-slate-900 border-slate-600 text-emerald-500 focus:ring-emerald-500" />
                                    <span class="text-sm text-slate-300">Can Login</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.createDb} class="rounded bg-slate-900 border-slate-600 text-blue-500 focus:ring-blue-500" />
                                    <span class="text-sm text-slate-300">Create Databases</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.createRole} class="rounded bg-slate-900 border-slate-600 text-blue-500 focus:ring-blue-500" />
                                    <span class="text-sm text-slate-300">Create Roles</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.inherit} class="rounded bg-slate-900 border-slate-600 text-blue-500 focus:ring-blue-500" />
                                    <span class="text-sm text-slate-300">Inherit Privileges</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.replication} class="rounded bg-slate-900 border-slate-600 text-pink-500 focus:ring-pink-500" />
                                    <span class="text-sm text-slate-300">Replication</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all">
                                    <input type="checkbox" bind:checked={newRole.bypassRls} class="rounded bg-slate-900 border-slate-600 text-red-500 focus:ring-red-500" />
                                    <span class="text-sm text-slate-300">Bypass RLS</span>
                                </label>
                                <label class="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700/30 cursor-pointer border border-transparent hover:border-slate-700 transition-all col-span-2 bg-amber-900/10 border-amber-900/30">
                                    <input type="checkbox" bind:checked={newRole.isSuperuser} class="rounded bg-slate-900 border-slate-600 text-amber-500 focus:ring-amber-500" />
                                    <span class="text-sm text-amber-200 font-bold">Superuser (Dangerous)</span>
                                </label>
                            </div>
                        </div>
                    </div>

                    <div class="flex justify-end gap-3 border-t border-slate-700/50 pt-4">
                        <button onclick={() => isCreatingRole = false} class="px-4 py-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition-colors">Cancel</button>
                        <button onclick={createRole} class="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-bold shadow-lg shadow-emerald-900/20 flex items-center gap-2">
                            <Check class="w-4 h-4" /> Create Role
                        </button>
                    </div>
                </div>
            {/if}

            <div class="bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden shadow-lg">
                <table class="w-full text-left text-sm">
                    <thead class="bg-slate-900/50 text-slate-400">
                        <tr>
                            <th class="px-6 py-4 font-medium">Role Name</th>
                            <th class="px-6 py-4 font-medium text-center">Attributes</th>
                            <th class="px-6 py-4 font-medium text-center">Member Of</th>
                            <th class="px-6 py-4 font-medium text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-700/50">
                        {#each roles as role}
                            <tr class="hover:bg-slate-700/20 group transition-colors">
                                <td class="px-6 py-4">
                                    <div class="flex items-center gap-3">
                                        <div class="p-2 rounded-lg bg-blue-500/10 text-blue-400">
                                            {#if role.superuser}
                                                <Key class="w-4 h-4 text-amber-400" />
                                            {:else}
                                                <Users class="w-4 h-4" />
                                            {/if}
                                        </div>
                                        <div>
                                            <div class="font-bold text-slate-200">{role.name}</div>
                                            <div class="text-xs text-slate-500">ID: {role.oid || 'N/A'}</div>
                                        </div>
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="flex flex-wrap gap-2 justify-center">
                                        {#if role.superuser} <span class="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 text-xs border border-amber-500/20">Superuser</span> {/if}
                                        {#if role.create_role} <span class="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 text-xs border border-blue-500/20">Create Role</span> {/if}
                                        {#if role.create_db} <span class="px-2 py-0.5 rounded bg-purple-500/10 text-purple-400 text-xs border border-purple-500/20">Create DB</span> {/if}
                                        {#if role.can_login} <span class="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 text-xs border border-emerald-500/20">Login</span> {/if}
                                        {#if role.replication} <span class="px-2 py-0.5 rounded bg-pink-500/10 text-pink-400 text-xs border border-pink-500/20">Replication</span> {/if}
                                        {#if role.bypass_rls} <span class="px-2 py-0.5 rounded bg-red-500/10 text-red-400 text-xs border border-red-500/20">Bypass RLS</span> {/if}
                                        {#if !role.superuser && !role.create_role && !role.create_db && !role.can_login} <span class="text-slate-600 text-xs italic">No special privileges</span> {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4 text-center text-slate-500 text-xs">
                                    -
                                </td>
                                <td class="px-6 py-4 text-right">
                                    <button 
                                        onclick={() => deleteRole(role.name)}
                                        class="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all opacity-0 group-hover:opacity-100"
                                        title="Delete Role"
                                    >
                                        <Trash2 class="w-4 h-4" />
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}

    {#if activeTab === 'sql'}
        <div class="space-y-4 h-[600px] flex flex-col">
            <div class="bg-slate-800/50 border border-slate-700 rounded-xl p-4 flex flex-col gap-4 shadow-lg">
                <div class="flex justify-between items-center">
                    <h3 class="text-lg font-bold text-slate-200 flex items-center gap-2">
                        <Terminal class="w-5 h-5 text-blue-400" /> 
                        Query Editor
                    </h3>
                    <button 
                        onclick={runSQL}
                        disabled={sqlLoading}
                        class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-bold flex items-center gap-2 shadow-lg shadow-emerald-900/20 disabled:opacity-50"
                    >
                        {#if sqlLoading}
                            <RefreshCw class="w-4 h-4 animate-spin" /> Running...
                        {:else}
                            <Play class="w-4 h-4 fill-current" /> Run Query
                        {/if}
                    </button>
                </div>
                <textarea 
                    bind:value={sqlQuery}
                    class="w-full h-32 bg-slate-900 border border-slate-700 rounded-lg p-3 font-mono text-sm text-slate-300 focus:border-blue-500 outline-none resize-none"
                    placeholder="SELECT * FROM table_name WHERE ..."
                ></textarea>
            </div>

            <!-- Results -->
            <div class="flex-1 bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden flex flex-col shadow-inner">
                <div class="p-3 border-b border-slate-700 bg-slate-900/50 text-xs text-slate-400 font-mono">
                    Results: {sqlResults.length} rows
                </div>
                <div class="flex-1 overflow-auto">
                    {#if sqlResults.length > 0}
                        <table class="w-full text-left text-xs">
                            <thead class="bg-slate-900 text-slate-400 sticky top-0">
                                <tr>
                                    {#each Object.keys(sqlResults[0]) as key}
                                        <th class="px-4 py-2 font-medium border-b border-slate-700 whitespace-nowrap">{key}</th>
                                    {/each}
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-slate-700/50">
                                {#each sqlResults as row}
                                    <tr class="hover:bg-slate-700/20">
                                        {#each Object.values(row) as val}
                                            <td class="px-4 py-2 text-slate-300 font-mono whitespace-nowrap max-w-xs truncate" title={String(val)}>{String(val)}</td>
                                        {/each}
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    {:else if !sqlLoading}
                        <div class="flex items-center justify-center h-full text-slate-500">
                            No results to display.
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    {/if}

    {#if activeTab === 'backups'}
        <div class="bg-slate-800/50 border border-slate-700 rounded-xl p-6">
            <div class="flex justify-between items-center mb-6">
                <h3 class="text-lg font-bold text-slate-200">Internal Backups</h3>
                <button onclick={createBackup} class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-bold flex items-center gap-2">
                    <Save class="w-4 h-4" /> Create Backup
                </button>
            </div>
            
            <div class="space-y-3">
                {#each backups as backup}
                    <div class="flex items-center justify-between p-4 bg-slate-900/50 rounded-lg border border-slate-700/50">
                        <div class="flex items-center gap-4">
                            <div class="p-3 bg-blue-500/10 rounded-lg text-blue-400">
                                <FileText class="w-6 h-6" />
                            </div>
                            <div>
                                <div class="font-mono text-slate-200">{backup.name}</div>
                                <div class="text-xs text-slate-500 flex gap-3 mt-1">
                                    <span>{formatBytes(backup.size)}</span>
                                    <span>{new Date(backup.created_at).toLocaleString()}</span>
                                </div>
                            </div>
                        </div>
                        <div class="flex gap-2">
                            <button onclick={() => downloadFile(backup.name)} class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded" title="Download">
                                <Download class="w-4 h-4" />
                            </button>
                            <button onclick={() => restoreBackup(backup.name)} class="p-2 text-orange-400 hover:text-orange-300 hover:bg-orange-500/10 rounded" title="Restore">
                                <RotateCcw class="w-4 h-4" />
                            </button>
                            <button onclick={() => deleteBackup(backup.name)} class="p-2 text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded" title="Delete">
                                <Trash2 class="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                {/each}
                {#if backups.length === 0}
                    <div class="text-center py-8 text-slate-500">No backups found.</div>
                {/if}
            </div>
        </div>
    {/if}

    {#if activeTab === 'config'}
        <div class="bg-slate-800/50 border border-slate-700 rounded-xl overflow-hidden flex flex-col">
            {#if Object.keys(pendingConfigChanges).length > 0}
                <div class="p-4 bg-orange-500/10 border-b border-orange-500/30 flex justify-between items-center">
                    <div class="flex items-center gap-2 text-orange-200">
                        <AlertCircle class="w-5 h-5" />
                        <span class="font-bold">{Object.keys(pendingConfigChanges).length} Pending Changes</span>
                    </div>
                    <button 
                        onclick={() => isConfigConfirmOpen = true}
                        class="px-4 py-2 bg-orange-600 hover:bg-orange-500 text-white rounded-lg font-bold flex items-center gap-2 shadow-lg shadow-orange-900/20"
                    >
                        <Save class="w-4 h-4" /> Confirm & Process
                    </button>
                </div>
            {/if}

            <table class="w-full text-left text-sm">
                <thead class="bg-slate-900/50 text-slate-400">
                    <tr>
                        <th class="px-6 py-4 font-medium">Parameter</th>
                        <th class="px-6 py-4 font-medium">Value</th>
                        <th class="px-6 py-4 font-medium">Description</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-700/50">
                    {#each pgConfig as cfg}
                        <tr class="hover:bg-slate-700/20 group {pendingConfigChanges[cfg.name] !== undefined ? 'bg-orange-900/10 border-l-2 border-l-orange-500' : ''}">
                            <td class="px-6 py-3 font-mono text-blue-400">{cfg.name}</td>
                            <td 
                                class="px-6 py-3 text-slate-300 font-mono max-w-md truncate cursor-pointer hover:bg-slate-700/50 {pendingConfigChanges[cfg.name] !== undefined ? 'text-orange-300 font-bold' : ''}" 
                                title={cfg.setting}
                                onclick={() => startConfigEdit(cfg)}
                            >
                                {#if editingConfig && editingConfig.name === cfg.name}
                                    <!-- svelte-ignore a11y_autofocus -->
                                    <input 
                                        type="text" 
                                        bind:value={editConfigValue}
                                        onblur={saveConfigEdit}
                                        onkeydown={(e) => e.key === 'Enter' && saveConfigEdit()}
                                        class="w-full bg-slate-900 border border-orange-500 rounded px-2 py-1 text-slate-200 outline-none"
                                        autofocus
                                    />
                                {:else}
                                    {cfg.setting}
                                    <Pencil class="w-3 h-3 inline ml-2 opacity-0 group-hover:opacity-50" />
                                {/if}
                            </td>
                            <td class="px-6 py-3 text-slate-500">{cfg.description}</td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
</div>

<!-- Postgres Config Confirmation Dialog -->
<ConfirmDialog
    bind:isOpen={isConfigConfirmOpen}
    title="Confirm Database Configuration Changes"
    message=""
    isCritical={true}
    confirmText="Process & Restart"
    onConfirm={applyConfigChanges}
>
    <div class="mt-4 bg-slate-950 p-4 rounded-lg border border-slate-800 max-h-64 overflow-y-auto">
        <table class="w-full text-sm text-left">
            <thead class="text-slate-500 border-b border-slate-800">
                <tr>
                    <th class="py-2">Parameter</th>
                    <th class="py-2">New Value</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-800">
                {#each Object.entries(pendingConfigChanges) as [key, value]}
                    <tr>
                        <td class="py-2 font-mono text-blue-400">{key}</td>
                        <td class="py-2 font-mono text-emerald-400 break-all">{value}</td>
                    </tr>
                {/each}
            </tbody>
        </table>
        <div class="mt-4 text-orange-400 text-sm flex items-start gap-2">
            <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
            <p>Processing these changes will trigger a PostgreSQL configuration reload (pg_reload_conf). Some settings may require a full server restart to take effect.</p>
        </div>
    </div>
</ConfirmDialog>

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title={confirmTitle}
    message={confirmMessage}
    isCritical={confirmIsCritical}
    onConfirm={confirmAction}
/>