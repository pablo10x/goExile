<script lang="ts">
	import { Database, Search, Plus, X, Trash2, ChevronRight, Pencil, Table, FolderOpen, Folder, ChevronLeft } from 'lucide-svelte';
	import { slide, fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';
    import TreeItem from '$lib/components/TreeItem.svelte';

	interface SchemaBrowserProps {
		schemas: string[];
		schemaTables: Record<string, string[]>;
		tableColumns: Record<string, any[]>;
		selectedSchema: string | null;
		selectedTable: string | null;
		expandedSchemas: Set<string>;
		expandedTables: Set<string>;
		isSidebarOpen?: boolean;
		onToggleSchema: (schema: string) => void;
		onToggleTable: (schema: string, table: string) => void;
		onSelectTable: (schema: string, table: string) => void;
		onCreateSchema: () => void;
		onCreateTable: (schema: string) => void;
		onDropTable: (schema: string, table: string) => void;
		onDropSchema: (schema: string) => void;
		onEditTable?: (schema: string, table: string) => void;
		onToggleSidebar: () => void;
	}

	let {
		schemas,
		schemaTables,
		selectedSchema,
		selectedTable,
		expandedSchemas,
		isSidebarOpen = true,
		onToggleSchema,
		onSelectTable,
		onCreateSchema,
		onCreateTable,
		onDropTable,
		onDropSchema,
		onEditTable,
		onToggleSidebar
	}: SchemaBrowserProps = $props();

	let filterText = $state('');

	let filteredData = $derived.by(() => {
        const lowerFilter = filterText.toLowerCase();
        return schemas.map(schema => {
            const tables = schemaTables[schema] || [];
            const matchesSchema = schema.toLowerCase().includes(lowerFilter);
            const matchingTables = tables.filter(t => t.toLowerCase().includes(lowerFilter));
            
            if (!matchesSchema && matchingTables.length === 0 && filterText) return null;

            return {
                name: schema,
                tables: filterText ? matchingTables : tables,
                isOpen: expandedSchemas.has(schema) || filterText.length > 0
            };
        }).filter(s => s !== null);
    });
</script>

<div
	class="bg-slate-950 border-r border-slate-800 flex flex-col h-full transition-all duration-300 ease-in-out {isSidebarOpen ? 'w-full' : 'w-16 bg-transparent border-none'}"
>
	<!-- Header -->
	<div class="p-3 border-b border-slate-800/50 flex flex-col gap-3 relative shrink-0">
		<div class="flex justify-between items-center h-8">
			{#if isSidebarOpen}
				<div class="flex items-center gap-2 text-slate-100 font-bold tracking-tight animate-in fade-in slide-in-from-left-2">
                    <div class="p-1.5 bg-indigo-500/10 rounded-lg">
                        <Database class="w-4 h-4 text-indigo-400" />
                    </div>
                    <span>Explorer</span>
                </div>
				<div class="flex gap-1">
					<button
						onclick={onCreateSchema}
						class="p-1.5 hover:bg-emerald-500/10 text-slate-400 hover:text-emerald-400 rounded-lg transition-colors"
						title="New Schema"
					>
						<Plus class="w-4 h-4" />
					</button>
					<button
						onclick={onToggleSidebar}
						class="p-1.5 hover:bg-slate-800 text-slate-400 hover:text-white rounded-lg transition-colors"
						title="Collapse"
					>
						<ChevronLeft class="w-4 h-4" />
					</button>
				</div>
			{:else}
				<button
					onclick={onToggleSidebar}
					class="w-10 h-10 flex items-center justify-center bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded-xl shadow-lg transition-all mx-auto"
					title="Expand"
				>
					<ChevronRight class="w-5 h-5" />
				</button>
			{/if}
		</div>

		{#if isSidebarOpen}
			<div class="relative group">
				<Search class="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
				<input
					type="text"
					bind:value={filterText}
					placeholder="Search..."
					class="w-full bg-slate-900/50 border border-slate-800 rounded-lg py-2 pl-9 pr-8 text-xs text-slate-300 placeholder:text-slate-600 focus:border-indigo-500/50 focus:bg-slate-900 focus:ring-1 focus:ring-indigo-500/20 outline-none transition-all"
				/>
				{#if filterText}
					<button
						onclick={() => (filterText = '')}
						class="absolute right-2 top-2 text-slate-500 hover:text-white p-0.5"
					>
						<X class="w-3.5 h-3.5" />
					</button>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Scrollable Tree -->
	{#if isSidebarOpen}
		<div class="flex-1 overflow-y-auto overflow-x-hidden p-2 space-y-1 custom-scrollbar">
            {#if filteredData.length === 0}
                <div class="flex flex-col items-center justify-center h-32 text-slate-600 text-xs animate-in fade-in">
                    <Database class="w-8 h-8 mb-2 opacity-20" />
                    <span>No schemas found</span>
                </div>
            {/if}

			{#each filteredData as schema (schema.name)}
				<div 
                    class="group/schema"
                    transition:slide|local={{ duration: 200, easing: quintOut }}
                >
                    <!-- Schema Header -->
					<div
                        class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer select-none transition-colors hover:bg-slate-800/50 text-sm group/row"
                        onclick={() => onToggleSchema(schema.name)}
                        onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onToggleSchema(schema.name); }}
                        role="button"
                        tabindex="0"
                    >
                        <div class="w-4 h-4 flex items-center justify-center text-slate-500 transition-transform duration-200 {schema.isOpen ? 'rotate-90 text-indigo-400' : ''}">
                            <ChevronRight class="w-3.5 h-3.5" />
                        </div>
                        
                        {#if schema.isOpen}
                            <FolderOpen class="w-4 h-4 text-indigo-400" />
                        {:else}
                            <Folder class="w-4 h-4 text-slate-500 group-hover/row:text-slate-400" />
                        {/if}

                        <span class="truncate font-medium {schema.isOpen ? 'text-indigo-100' : 'text-slate-400'} flex-1">
                            {schema.name}
                        </span>

                        <!-- Schema Actions (Hover) -->
                        <div class="flex gap-1 opacity-0 group-hover/row:opacity-100 transition-opacity">
                            <button
                                onclick={(e) => {
                                    e.stopPropagation();
                                    onCreateTable(schema.name);
                                }}
                                class="p-1 hover:bg-emerald-500/20 text-slate-500 hover:text-emerald-400 rounded"
                                title="Create Table"
                            >
                                <Plus class="w-3 h-3" />
                            </button>
                            <button
                                onclick={(e) => {
                                    e.stopPropagation();
                                    onDropSchema(schema.name);
                                }}
                                class="p-1 hover:bg-red-500/20 text-slate-500 hover:text-red-400 rounded"
                                title="Drop Schema"
                            >
                                <Trash2 class="w-3 h-3" />
                            </button>
                        </div>
					</div>

                    <!-- Tables List -->
                    {#if schema.isOpen}
                        <div 
                            class="ml-2 pl-2 border-l border-slate-800/50 space-y-0.5 mt-0.5 mb-1"
                            transition:slide|local={{ duration: 150 }}
                        >
                            {#if schema.tables.length === 0}
                                <div class="px-3 py-1 text-[10px] text-slate-600 italic pl-6">
                                    Empty schema
                                </div>
                            {/if}

                            {#each schema.tables as table (table)}
                                <div 
                                    class="flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer select-none transition-all group/table text-xs 
                                    {selectedSchema === schema.name && selectedTable === table 
                                        ? 'bg-indigo-500/10 text-indigo-300' 
                                        : 'text-slate-400 hover:bg-slate-800/30 hover:text-slate-200'}"
                                    onclick={() => onSelectTable(schema.name, table)}
                                    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onSelectTable(schema.name, table); }}
                                    role="button"
                                    tabindex="0"
                                >
                                    <Table class="w-3.5 h-3.5 {selectedSchema === schema.name && selectedTable === table ? 'text-indigo-400' : 'text-slate-600'}" />
                                    <span class="truncate flex-1">{table}</span>

                                    <!-- Table Actions -->
                                    <div class="flex gap-1 opacity-0 group-hover/table:opacity-100 transition-opacity">
                                        <button
                                            onclick={(e) => {
                                                e.stopPropagation();
                                                onEditTable?.(schema.name, table);
                                            }}
                                            class="p-1 hover:bg-blue-500/20 text-slate-500 hover:text-blue-400 rounded"
                                            title="Edit Structure"
                                        >
                                            <Pencil class="w-3 h-3" />
                                        </button>
                                        <button
                                            onclick={(e) => {
                                                e.stopPropagation();
                                                onDropTable(schema.name, table);
                                            }}
                                            class="p-1 hover:bg-red-500/20 text-slate-500 hover:text-red-400 rounded"
                                            title="Drop Table"
                                        >
                                            <Trash2 class="w-3 h-3" />
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
    .custom-scrollbar::-webkit-scrollbar {
        width: 4px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: transparent;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: #334155;
        border-radius: 2px;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: #475569;
    }
</style>
