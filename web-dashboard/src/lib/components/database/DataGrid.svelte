<script lang="ts">
	import { ArrowUp, ArrowDown, Save, Trash2, Plus, Filter, MoreHorizontal, CheckSquare, Square } from 'lucide-svelte';
	import { fade } from 'svelte/transition';

	let { 
        data = [], 
        columns = [], // Expects { name, type, nullable, isPk }
        isLoading = false,
        onSort,
        onSaveRow,
        onDeleteRows,
        onAddRow
    } = $props<{
        data: any[];
        columns: any[];
        isLoading: boolean;
        onSort: (col: string, dir: 'asc' | 'desc') => void;
        onSaveRow: (id: any, changes: any) => Promise<void>;
        onDeleteRows: (ids: any[]) => Promise<void>;
        onAddRow: () => void;
    }>();

    let selectedIds = $state<Set<any>>(new Set());
    let sortCol = $state<string | null>(null);
    let sortDir = $state<'asc' | 'desc'>('asc');
    let editingCell = $state<{ id: any, col: string } | null>(null);
    let editValue = $state('');
    let pendingChanges = $state<Record<string, any>>({}); // id -> { col: val }

    function handleSort(col: string) {
        if (sortCol === col) {
            sortDir = sortDir === 'asc' ? 'desc' : 'asc';
        } else {
            sortCol = col;
            sortDir = 'asc';
        }
        onSort(sortCol, sortDir);
    }

    function toggleSelectAll() {
        if (selectedIds.size === data.length && data.length > 0) {
            selectedIds = new Set();
        } else {
            selectedIds = new Set(data.map((r: any) => r.id));
        }
    }

    function toggleSelect(id: any) {
        if (selectedIds.has(id)) {
            selectedIds.delete(id);
        } else {
            selectedIds.add(id);
        }
        selectedIds = new Set(selectedIds); // Trigger update
    }

    function startEdit(row: any, col: string) {
        // if (row.id === undefined) return; // Can't edit without ID?
        editingCell = { id: row.id, col };
        editValue = String(row[col] ?? '');
    }

    async function saveEdit() {
        if (!editingCell) return;
        const { id, col } = editingCell;
        
        // Optimistic update visual
        const row = data.find((r: any) => r.id === id);
        if (row && String(row[col]) !== editValue) {
             // Store pending change to commit later or immediately?
             // For "production ready", immediate save on blur/enter is good, or batch.
             // Let's do immediate for simplicity + robust feel.
             await onSaveRow(id, { [col]: editValue });
        }
        editingCell = null;
    }

    async function deleteSelected() {
        if (selectedIds.size === 0) return;
        if (!confirm(`Delete ${selectedIds.size} rows?`)) return;
        await onDeleteRows(Array.from(selectedIds));
        selectedIds = new Set();
    }
</script>

<div class="flex flex-col h-full bg-slate-900/50">
    <!-- Toolbar -->
    <div class="p-2 border-b border-slate-700 bg-slate-900 flex justify-between items-center">
        <div class="flex items-center gap-2">
            <button 
                disabled={selectedIds.size === 0}
                onclick={deleteSelected}
                class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors {selectedIds.size > 0 ? 'bg-red-500/10 text-red-400 hover:bg-red-500/20' : 'text-slate-600 cursor-not-allowed'}"
            >
                <Trash2 class="w-4 h-4" /> Delete ({selectedIds.size})
            </button>
            <div class="h-4 w-px bg-slate-700 mx-2"></div>
            <button class="text-slate-400 hover:text-white p-1.5 rounded hover:bg-slate-800">
                <Filter class="w-4 h-4" />
            </button>
        </div>
        <div class="flex items-center gap-2">
             <button 
                onclick={onAddRow}
                class="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-bold shadow-lg shadow-blue-900/20 transition-all hover:scale-105"
            >
                <Plus class="w-4 h-4" /> Add Row
            </button>
        </div>
    </div>

    <!-- Grid -->
    <div class="flex-1 overflow-auto relative">
        {#if isLoading}
            <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm z-10 flex items-center justify-center">
                <div class="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
            </div>
        {/if}

        <table class="w-full text-left text-sm border-collapse">
            <thead class="bg-slate-950 text-slate-400 sticky top-0 z-10 shadow-sm">
                <tr>
                    <th class="px-4 py-3 border-b border-slate-800 w-10 text-center">
                        <button onclick={toggleSelectAll} class="hover:text-blue-400">
                            {#if selectedIds.size > 0 && selectedIds.size === data.length}
                                <CheckSquare class="w-4 h-4 text-blue-500" />
                            {:else if selectedIds.size > 0}
                                <div class="w-4 h-4 bg-blue-500 rounded-sm flex items-center justify-center">
                                    <div class="w-2 h-0.5 bg-white"></div>
                                </div>
                            {:else}
                                <Square class="w-4 h-4" />
                            {/if}
                        </button>
                    </th>
                    {#each columns as col}
                        <th 
                            class="px-4 py-3 border-b border-slate-800 font-medium whitespace-nowrap cursor-pointer hover:bg-slate-900 hover:text-white group transition-colors"
                            onclick={() => handleSort(col.name)}
                        >
                            <div class="flex items-center gap-2">
                                {#if col.name === 'id' || col.isPk}
                                    <span class="text-amber-500 text-xs">PK</span>
                                {/if}
                                {col.name}
                                {#if sortCol === col.name}
                                    {#if sortDir === 'asc'}
                                        <ArrowUp class="w-3 h-3 text-blue-400" />
                                    {:else}
                                        <ArrowDown class="w-3 h-3 text-blue-400" />
                                    {/if}
                                {:else}
                                    <ArrowUp class="w-3 h-3 opacity-0 group-hover:opacity-30" />
                                {/if}
                            </div>
                            <div class="text-[10px] text-slate-600 font-normal font-mono mt-0.5">{col.type}</div>
                        </th>
                    {/each}
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/50 bg-slate-900">
                {#each data as row (row.id)}
                    <tr class="hover:bg-slate-800/50 transition-colors {selectedIds.has(row.id) ? 'bg-blue-900/10' : ''}">
                        <td class="px-4 py-2 border-r border-slate-800/30 text-center">
                            <button onclick={() => toggleSelect(row.id)} class="text-slate-500 hover:text-blue-400">
                                {#if selectedIds.has(row.id)}
                                    <CheckSquare class="w-4 h-4 text-blue-500" />
                                {:else}
                                    <Square class="w-4 h-4" />
                                {/if}
                            </button>
                        </td>
                        {#each columns as col}
                            <td 
                                class="px-4 py-2 border-r border-slate-800/30 max-w-xs truncate cursor-text text-slate-300 relative group/cell"
                                onclick={() => startEdit(row, col.name)}
                                title={String(row[col.name])}
                            >
                                {#if editingCell?.id === row.id && editingCell?.col === col.name}
                                    <input 
                                        type="text" 
                                        bind:value={editValue}
                                        onblur={saveEdit}
                                        onkeydown={(e) => e.key === 'Enter' && saveEdit()}
                                        class="absolute inset-0 w-full h-full bg-slate-800 text-white px-4 outline-none border-2 border-blue-500 z-20"
                                        autofocus
                                    />
                                {:else}
                                    <span class="{row[col.name] === null ? 'text-slate-600 italic' : ''}">
                                        {row[col.name] === null ? 'NULL' : String(row[col.name])}
                                    </span>
                                {/if}
                            </td>
                        {/each}
                    </tr>
                {/each}
                {#if data.length === 0 && !isLoading}
                    <tr>
                        <td colspan={columns.length + 1} class="p-8 text-center text-slate-500">
                            No rows found. Add a row to get started.
                        </td>
                    </tr>
                {/if}
            </tbody>
        </table>
    </div>
    
    <!-- Footer / Status -->
    <div class="px-4 py-2 bg-slate-950 border-t border-slate-800 text-xs text-slate-500 flex justify-between">
        <span>{data.length} rows loaded</span>
        <span>Selection: {selectedIds.size}</span>
    </div>
</div>
