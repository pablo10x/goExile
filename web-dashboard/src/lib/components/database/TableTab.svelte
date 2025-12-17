<script lang="ts">
	import { onMount } from 'svelte';
    import DataGrid from './DataGrid.svelte';
    import TableDesigner from './TableDesigner.svelte';
    import RowEditorModal from './RowEditorModal.svelte';
    import { notifications } from '$lib/stores';
    import { Settings, Table, RefreshCw } from 'lucide-svelte';

	let { schema, table } = $props<{ schema: string; table: string }>();

    let view = $state<'data' | 'structure'>('data');
    let rows = $state<any[]>([]);
    let columns = $state<any[]>([]);
    let loading = $state(false);
    let limit = $state(50);
    let offset = $state(0);

    // Row Editor State
    let isRowEditorOpen = $state(false);
    let editingRowData = $state<any>(null); // null for add, object for edit

    async function loadData() {
        loading = true;
        try {
            // Load Columns first to know structure
            const colRes = await fetch(`/api/database/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`);
            if (colRes.ok) {
                const cols = await colRes.json();
                if (cols) {
                    // Map to grid format
                    columns = cols.map((c: any) => ({ name: c.name, type: c.type, nullable: c.nullable === 'YES', isPk: false })); // PK info missing from API currently?
                } else {
                    columns = [];
                }
            }

            // Load Rows
            const res = await fetch(`/api/database/table/${table}?schema=${encodeURIComponent(schema)}&limit=${limit}&offset=${offset}`);
            if (res.ok) {
                const data = await res.json();
                rows = data.data || [];
            }
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to load table data', details: e.message });
        } finally {
            loading = false;
        }
    }

    async function saveRow(id: any, changes: any) {
        try {
            const res = await fetch(`/api/database/table/${table}/${id}?schema=${encodeURIComponent(schema)}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(changes)
            });
            if (!res.ok) throw new Error('Update failed');
            notifications.add({ type: 'success', message: 'Row updated' });
            // Update local state properly? DataGrid does optimistic, but we should confirm.
            // loadData(); // Reloading might be disruptive.
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Save failed', details: e.message });
        }
    }

    async function deleteRows(ids: any[]) {
        try {
            for (const id of ids) {
                await fetch(`/api/database/table/${table}/${id}?schema=${encodeURIComponent(schema)}`, { method: 'DELETE' });
            }
            notifications.add({ type: 'success', message: `Deleted ${ids.length} rows` });
            loadData();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Delete failed', details: e.message });
        }
    }

    function addRow() {
        editingRowData = null;
        isRowEditorOpen = true;
    }

    async function handleRowSave(data: any) {
        try {
            const res = await fetch(`/api/database/table/${table}?schema=${encodeURIComponent(schema)}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Insert failed');
            }
            notifications.add({ type: 'success', message: 'Row added successfully' });
            loadData();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to add row', details: e.message });
            throw e; // Keep modal open if error? The modal handles close on success usually, but here we throw so it catches in modal?
            // Actually RowEditorModal catches errors if passed as promise? No, it awaits.
            // Let's rely on modal staying open or closing.
            // Wait, RowEditorModal implementation: 
            // try { await onSave(payload); onClose(); }
            // So if I throw, it stays open. Good.
        }
    }

    onMount(() => {
        loadData();
    });
</script>

<div class="flex flex-col h-full">
    <!-- Tab Toolbar (Sub-nav for table) -->
    <div class="flex items-center justify-between bg-slate-950 border-b border-slate-800 px-2">
        <div class="flex gap-1">
            <button 
                onclick={() => view = 'data'}
                class="px-4 py-2 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 {view === 'data' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-500 hover:text-slate-300'}"
            >
                <Table class="w-3.5 h-3.5" /> Data
            </button>
            <button 
                onclick={() => view = 'structure'}
                class="px-4 py-2 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 {view === 'structure' ? 'border-purple-500 text-purple-400' : 'border-transparent text-slate-500 hover:text-slate-300'}"
            >
                <Settings class="w-3.5 h-3.5" /> Structure
            </button>
        </div>
        <button 
            onclick={loadData}
            class="p-1.5 text-slate-500 hover:text-white rounded hover:bg-slate-800 transition-colors"
            title="Refresh Data"
        >
            <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
        </button>
    </div>

    <div class="flex-1 overflow-hidden">
        {#if view === 'data'}
            <DataGrid 
                data={rows} 
                {columns} 
                isLoading={loading}
                onSort={(col, dir) => { /* TODO: Backend sort */ }}
                onSaveRow={saveRow}
                onDeleteRows={deleteRows}
                onAddRow={addRow}
            />
        {:else}
            <TableDesigner 
                {schema} 
                {table} 
                {columns} 
                onClose={() => view = 'data'} 
                onReload={loadData} 
            />
        {/if}
    </div>

    <RowEditorModal
        bind:isOpen={isRowEditorOpen}
        {schema}
        {table}
        {columns}
        rowData={editingRowData}
        onClose={() => isRowEditorOpen = false}
        onSave={handleRowSave}
    />
</div>
