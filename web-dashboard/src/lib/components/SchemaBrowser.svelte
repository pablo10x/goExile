<script lang="ts">
	import {
		Database,
		Search,
		Plus,
		X,
		Trash2,
		ChevronRight,
		Pencil,
		Table,
		FolderOpen,
		Folder,
		ChevronLeft
	} from 'lucide-svelte';
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
		return schemas
			.map((schema) => {
				const tables = schemaTables[schema] || [];
				const matchesSchema = schema.toLowerCase().includes(lowerFilter);
				const matchingTables = tables.filter((t) => t.toLowerCase().includes(lowerFilter));

				if (!matchesSchema && matchingTables.length === 0 && filterText) return null;

				return {
					name: schema,
					tables: filterText ? matchingTables : tables,
					isOpen: expandedSchemas.has(schema) || filterText.length > 0
				};
			})
			.filter((s) => s !== null);
	});
</script>

<div
	class="bg-neutral-900/40 border-r border-neutral-800 flex flex-col h-full transition-all duration-300 ease-in-out backdrop-blur-sm {isSidebarOpen
		? 'w-full'
		: 'w-16 bg-transparent border-none'}"
>
	<!-- Header -->
	<div
		class="p-4 border-b border-neutral-800 flex flex-col gap-4 relative shrink-0"
	>
		<div class="flex justify-between items-center h-10">
			{#if isSidebarOpen}
				<div
					class="flex items-center gap-3 text-neutral-100 animate-in fade-in slide-in-from-left-2"
				>
					<div class="p-2 bg-indigo-500/10 border border-indigo-500/20 rounded-xl">
						<FolderOpen class="w-4 h-4 text-indigo-400" />
					</div>
					<span class="font-heading font-black text-xs uppercase tracking-widest italic">Explorer</span>
				</div>
				<div class="flex gap-1">
					<button
						onclick={onCreateSchema}
						class="p-2 hover:bg-neutral-800 text-neutral-500 hover:text-indigo-400 transition-all rounded-xl"
						title="New Schema"
					>
						<Plus class="w-4 h-4" />
					</button>
					<button
						onclick={onToggleSidebar}
						class="p-2 hover:bg-neutral-800 text-neutral-500 hover:text-white transition-all rounded-xl"
						title="Collapse"
					>
						<ChevronLeft class="w-4 h-4" />
					</button>
				</div>
			{:else}
				<button
					onclick={onToggleSidebar}
					class="w-10 h-10 flex items-center justify-center bg-neutral-900/50 border border-neutral-800 text-neutral-500 hover:text-indigo-400 transition-all mx-auto rounded-xl"
					title="Expand"
				>
					<ChevronRight class="w-5 h-5" />
				</button>
			{/if}
		</div>

		{#if isSidebarOpen}
			<div class="relative group">
				<Search
					class="w-3.5 h-3.5 absolute left-3 top-1/2 -tranneutral-y-1/2 text-neutral-500 group-focus-within:text-indigo-400 transition-colors"
				/>
				<input
					type="text"
					bind:value={filterText}
					placeholder="Filter objects..."
					class="w-full bg-neutral-950/40 border border-neutral-800 py-2 pl-9 pr-8 text-xs font-bold uppercase tracking-widest text-neutral-300 placeholder:text-neutral-700 focus:border-indigo-500/50 outline-none transition-all rounded-xl"
				/>
				{#if filterText}
					<button
						onclick={() => (filterText = '')}
						class="absolute right-2 top-1/2 -tranneutral-y-1/2 text-neutral-500 hover:text-white p-0.5"
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
				<div
					class="flex flex-col items-center justify-center h-32 text-neutral-700 text-[10px] font-bold uppercase tracking-[0.3em] animate-in fade-in"
				>
					<Database class="w-8 h-8 mb-3 opacity-20" />
					<span>No_Matching_Signals</span>
				</div>
			{/if}

			{#each filteredData as schema (schema.name)}
				<div class="group/schema" transition:slide|local={{ duration: 200, easing: quintOut }}>
					<!-- Schema Header -->
					<div
						class="flex items-center gap-3 px-3 py-2 cursor-pointer select-none transition-all hover:bg-indigo-500/5 group/row border border-transparent rounded-xl"
						onclick={() => onToggleSchema(schema.name)}
						onkeydown={(e) => {
							if (e.key === 'Enter' || e.key === ' ') onToggleSchema(schema.name);
						}}
						role="button"
						tabindex="0"
					>
						<div
							class="w-4 h-4 flex items-center justify-center text-neutral-600 transition-transform duration-200 {schema.isOpen
								? 'rotate-90 text-indigo-400'
								: ''}"
						>
							<ChevronRight class="w-3.5 h-3.5" />
						</div>

						{#if schema.isOpen}
							<FolderOpen class="w-4 h-4 text-indigo-400" />
						{:else}
							<Folder
								class="w-4 h-4 text-neutral-600 group-hover/row:text-neutral-400 transition-colors"
							/>
						{/if}

						<span
							class="truncate font-bold text-[10px] uppercase tracking-widest {schema.isOpen
								? 'text-indigo-200'
								: 'text-neutral-500'}"
						>
							{schema.name}
						</span>

						<!-- Schema Actions (Hover) -->
						<div class="flex gap-1 opacity-0 group-hover/row:opacity-100 transition-opacity ml-auto">
							<button
								onclick={(e) => {
									e.stopPropagation();
									onCreateTable(schema.name);
								}}
								class="p-1 hover:bg-indigo-500/10 text-neutral-600 hover:text-indigo-400 transition-colors rounded-lg"
								title="Create Table"
							>
								<Plus class="w-3.5 h-3.5" />
							</button>
							<button
								onclick={(e) => {
									e.stopPropagation();
									onDropSchema(schema.name);
								}}
								class="p-1 hover:bg-red-500/10 text-neutral-600 hover:text-red-400 transition-colors rounded-lg"
								title="Drop Schema"
							>
								<Trash2 class="w-3.5 h-3.5" />
							</button>
						</div>
					</div>

					<!-- Tables List -->
					{#if schema.isOpen}
						<div
							class="ml-3 pl-3 border-l border-neutral-800 space-y-0.5 mt-1 mb-2"
							transition:slide|local={{ duration: 150 }}
						>
							{#if schema.tables.length === 0}
								<div class="px-3 py-2 text-[9px] font-medium text-neutral-700 uppercase italic tracking-widest">No_Tables_Mapped</div>
							{/if}

							{#each schema.tables as table (table)}
								<div
									class="flex items-center gap-3 px-3 py-2 cursor-pointer select-none transition-all group/table rounded-xl
                                    {selectedSchema === schema.name && selectedTable === table
										? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm shadow-indigo-500/5'
										: 'text-neutral-500 hover:bg-neutral-800/50 hover:text-white border border-transparent'}"
									onclick={() => onSelectTable(schema.name, table)}
									onkeydown={(e) => {
										if (e.key === 'Enter' || e.key === ' ') onSelectTable(schema.name, table);
									}}
									role="button"
									tabindex="0"
								>
									<Table
										class="w-3.5 h-3.5 {selectedSchema === schema.name && selectedTable === table
											? 'text-indigo-400'
											: 'text-neutral-700 group-hover/table:text-neutral-500'}"
									/>
									<span class="truncate font-mono text-[10px] font-bold uppercase tracking-tight flex-1">{table}</span>

									<!-- Table Actions -->
									<div
										class="flex gap-1 opacity-0 group-hover/table:opacity-100 transition-opacity"
									>
										<button
											onclick={(e) => {
												e.stopPropagation();
												onEditTable?.(schema.name, table);
											}}
											class="p-1 hover:bg-neutral-800 text-neutral-600 hover:text-indigo-400 transition-all rounded-lg"
											title="Edit Structure"
										>
											<Pencil class="w-3 h-3" />
										</button>
										<button
											onclick={(e) => {
												e.stopPropagation();
												onDropTable(schema.name, table);
											}}
											class="p-1 hover:bg-red-500/10 text-neutral-600 hover:text-red-400 transition-all rounded-lg"
											title="Drop Table"
										>
											<Trash2 class="w-3.5 h-3" />
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