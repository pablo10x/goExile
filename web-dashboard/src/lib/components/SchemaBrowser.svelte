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
	class="bg-[var(--header-bg)] border-r border-stone-800 flex flex-col h-full transition-all duration-300 ease-in-out {isSidebarOpen
		? 'w-full'
		: 'w-16 bg-transparent border-none'}"
>
	<!-- Header -->
	<div
		class="p-4 border-b border-stone-800 flex flex-col gap-4 relative shrink-0"
	>
		<div class="flex justify-between items-center h-10">
			{#if isSidebarOpen}
				<div
					class="flex items-center gap-3 text-slate-100 animate-in fade-in slide-in-from-left-2"
				>
					<div class="p-2 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
						<Database class="w-4 h-4 text-rust-light" />
					</div>
					<span class="font-heading font-black text-[11px] uppercase tracking-[0.2em]">EXPLORER</span>
				</div>
				<div class="flex gap-1">
					<button
						onclick={onCreateSchema}
						class="p-2 hover:bg-rust/10 text-stone-500 hover:text-rust transition-all"
						title="New Schema"
					>
						<Plus class="w-4 h-4" />
					</button>
					<button
						onclick={onToggleSidebar}
						class="p-2 hover:bg-stone-800 text-stone-500 hover:text-white transition-all"
						title="Collapse"
					>
						<ChevronLeft class="w-4 h-4" />
					</button>
				</div>
			{:else}
				<button
					onclick={onToggleSidebar}
					class="w-10 h-10 flex items-center justify-center bg-stone-900 border border-stone-800 text-stone-500 hover:text-rust transition-all mx-auto industrial-frame"
					title="Expand"
				>
					<ChevronRight class="w-5 h-5" />
				</button>
			{/if}
		</div>

		{#if isSidebarOpen}
			<div class="relative group">
				<Search
					class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-stone-600 group-focus-within:text-rust transition-colors"
				/>
				<input
					type="text"
					bind:value={filterText}
					placeholder="FILTER..."
					class="w-full bg-stone-950 border border-stone-800 py-2.5 pl-10 pr-8 font-jetbrains text-[10px] text-stone-300 placeholder:text-stone-700 focus:border-rust outline-none transition-all uppercase tracking-widest"
				/>
				{#if filterText}
					<button
						onclick={() => (filterText = '')}
						class="absolute right-2 top-1/2 -translate-y-1/2 text-stone-600 hover:text-white p-0.5"
					>
						<X class="w-3.5 h-3.5" />
					</button>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Scrollable Tree -->
	{#if isSidebarOpen}
		<div class="flex-1 overflow-y-auto overflow-x-hidden p-3 space-y-1 custom-scrollbar">
			{#if filteredData.length === 0}
				<div
					class="flex flex-col items-center justify-center h-32 text-stone-700 text-[10px] font-black uppercase tracking-widest animate-in fade-in"
				>
					<Database class="w-8 h-8 mb-3 opacity-10" />
					<span>Zero_Results</span>
				</div>
			{/if}

			{#each filteredData as schema (schema.name)}
				<div class="group/schema" transition:slide|local={{ duration: 200, easing: quintOut }}>
					<!-- Schema Header -->
					<div
						class="flex items-center gap-3 px-3 py-2 cursor-pointer select-none transition-all hover:bg-stone-900/50 group/row border border-transparent hover:border-stone-800/50"
						onclick={() => onToggleSchema(schema.name)}
						onkeydown={(e) => {
							if (e.key === 'Enter' || e.key === ' ') onToggleSchema(schema.name);
						}}
						role="button"
						tabindex="0"
					>
						<div
							class="w-4 h-4 flex items-center justify-center text-stone-600 transition-transform duration-200 {schema.isOpen
								? 'rotate-90 text-rust'
								: ''}"
						>
							<ChevronRight class="w-3.5 h-3.5" />
						</div>

						{#if schema.isOpen}
							<FolderOpen class="w-4 h-4 text-rust" />
						{:else}
							<Folder
								class="w-4 h-4 text-stone-600 group-hover/row:text-stone-400 transition-colors"
							/>
						{/if}

						<span
							class="truncate font-heading font-bold text-[11px] tracking-widest uppercase {schema.isOpen
								? 'text-rust-light'
								: 'text-stone-500'}"
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
								class="p-1 hover:bg-rust/10 text-stone-600 hover:text-rust transition-colors"
								title="Create Table"
							>
								<Plus class="w-3.5 h-3.5" />
							</button>
							<button
								onclick={(e) => {
									e.stopPropagation();
									onDropSchema(schema.name);
								}}
								class="p-1 hover:bg-red-500/10 text-stone-600 hover:text-red-500 transition-colors"
								title="Drop Schema"
							>
								<Trash2 class="w-3.5 h-3.5" />
							</button>
						</div>
					</div>

					<!-- Tables List -->
					{#if schema.isOpen}
						<div
							class="ml-4 pl-4 border-l border-stone-800/50 space-y-1 mt-1 mb-2"
							transition:slide|local={{ duration: 150 }}
						>
							{#if schema.tables.length === 0}
								<div class="px-3 py-2 text-[9px] font-jetbrains font-bold text-stone-700 italic uppercase tracking-widest">Sector_Empty</div>
							{/if}

							{#each schema.tables as table (table)}
								<div
									class="flex items-center gap-3 px-3 py-2 cursor-pointer select-none transition-all group/table
                                    {selectedSchema === schema.name && selectedTable === table
										? 'bg-rust/10 text-rust border-l-2 border-rust'
										: 'text-stone-500 hover:bg-stone-900 hover:text-stone-200'}"
									onclick={() => onSelectTable(schema.name, table)}
									onkeydown={(e) => {
										if (e.key === 'Enter' || e.key === ' ') onSelectTable(schema.name, table);
									}}
									role="button"
									tabindex="0"
								>
									<Table
										class="w-3.5 h-3.5 {selectedSchema === schema.name && selectedTable === table
											? 'text-rust'
											: 'text-stone-700'}"
									/>
									<span class="truncate font-jetbrains text-[10px] font-black uppercase tracking-widest flex-1">{table}</span>

									<!-- Table Actions -->
									<div
										class="flex gap-1 opacity-0 group-hover/table:opacity-100 transition-opacity"
									>
										<button
											onclick={(e) => {
												e.stopPropagation();
												onEditTable?.(schema.name, table);
											}}
											class="p-1 hover:bg-stone-800 text-stone-600 hover:text-rust-light transition-all"
											title="Edit Structure"
										>
											<Pencil class="w-3 h-3" />
										</button>
										<button
											onclick={(e) => {
												e.stopPropagation();
												onDropTable(schema.name, table);
											}}
											class="p-1 hover:bg-red-500/10 text-stone-600 hover:text-red-500 transition-all"
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
		background: #2a2a2a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}
</style>
