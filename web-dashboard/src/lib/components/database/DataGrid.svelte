<script lang="ts">
	import {
		ArrowUp,
		ArrowDown,
		Trash2,
		Plus,
		Filter,
		CheckSquare,
		Square,
		RefreshCw,
		Search,
		X,
		ChevronLeft,
		ChevronRight,
		Edit3,
		Copy,
		MoreVertical,
		Save,
		Undo2
	} from 'lucide-svelte';
	import { fade, slide } from 'svelte/transition';
	import ConfirmModal from './ConfirmModal.svelte';

	let {
		data = [],
		columns = [],
		isLoading = false,
		primaryKeyColumn = 'id',
		onSort,
		onSaveRow,
		onDeleteRows,
		onAddRow,
		onRefresh
	} = $props<{
		data: any[];
		columns: any[];
		isLoading: boolean;
		primaryKeyColumn?: string;
		onSort: (col: string, dir: 'asc' | 'desc') => void;
		onSaveRow: (id: any, changes: any) => Promise<void>;
		onDeleteRows: (ids: any[]) => Promise<void>;
		onAddRow: () => void;
		onRefresh?: () => void;
	}>();

	let selectedIds = $state<Set<any>>(new Set());
	let sortCol = $state<string | null>(null);
	let sortDir = $state<'asc' | 'desc'>('asc');
	let editingCell = $state<{ rowKey: any; col: string } | null>(null);
	let editValue = $state('');
	let searchQuery = $state('');
	let showSearch = $state(false);

	// Pending changes: Map<rowKey, Map<colName, newValue>>
	let pendingChanges = $state<Map<any, Map<string, any>>>(new Map());

	// Confirmation modal state
	let showDeleteConfirm = $state(false);
	let deleteLoading = $state(false);
	let pendingDeleteIds = $state<any[]>([]);

	// Apply changes loading state
	let applyingChanges = $state(false);

	// Get the row's unique identifier (primary key value)
	function getRowKey(row: any): any {
		return row[primaryKeyColumn];
	}

	// Check if a cell has pending changes
	function hasPendingChange(rowKey: any, col: string): boolean {
		const rowChanges = pendingChanges.get(rowKey);
		return rowChanges?.has(col) ?? false;
	}

	// Get pending value for a cell (or original if no pending change)
	function getCellValue(row: any, col: string): any {
		const rowKey = getRowKey(row);
		const rowChanges = pendingChanges.get(rowKey);
		if (rowChanges?.has(col)) {
			return rowChanges.get(col);
		}
		return row[col];
	}

	// Count total pending changes
	let pendingChangeCount = $derived.by(() => {
		let count = 0;
		pendingChanges.forEach((colChanges) => {
			count += colChanges.size;
		});
		return count;
	});

	// Filtered data based on search
	let filteredData = $derived.by(() => {
		if (!searchQuery.trim()) return data;
		const query = searchQuery.toLowerCase();
		return data.filter((row: any) =>
			columns.some((col: any) => {
				const val = getCellValue(row, col.name);
				return val !== null && String(val).toLowerCase().includes(query);
			})
		);
	});

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
		if (selectedIds.size === filteredData.length && filteredData.length > 0) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set(filteredData.map((r: any) => getRowKey(r)));
		}
	}

	function toggleSelect(rowKey: any) {
		if (selectedIds.has(rowKey)) {
			selectedIds.delete(rowKey);
		} else {
			selectedIds.add(rowKey);
		}
		selectedIds = new Set(selectedIds);
	}

	function startEdit(row: any, col: string) {
		const pkCol = columns.find((c: any) => c.isPk)?.name || primaryKeyColumn;
		if (col === pkCol) return; // Don't edit primary key
		const rowKey = getRowKey(row);
		editingCell = { rowKey, col };
		editValue = getCellValue(row, col) === null ? '' : String(getCellValue(row, col));
	}

	function saveEdit() {
		if (!editingCell) return;
		const { rowKey, col } = editingCell;

		const row = data.find((r: any) => getRowKey(r) === rowKey);
		if (row) {
			const originalValue = row[col];
			const newValue = editValue === '' ? null : editValue;

			// Check if value actually changed from original
			if (String(originalValue ?? '') !== String(newValue ?? '')) {
				// Stage the change
				let rowChanges = pendingChanges.get(rowKey);
				if (!rowChanges) {
					rowChanges = new Map();
					pendingChanges.set(rowKey, rowChanges);
				}
				rowChanges.set(col, newValue);
				// Trigger reactivity
				pendingChanges = new Map(pendingChanges);
			} else {
				// If value reverted to original, remove from pending
				const rowChanges = pendingChanges.get(rowKey);
				if (rowChanges) {
					rowChanges.delete(col);
					if (rowChanges.size === 0) {
						pendingChanges.delete(rowKey);
					}
					pendingChanges = new Map(pendingChanges);
				}
			}
		}
		editingCell = null;
	}

	function cancelEdit() {
		editingCell = null;
		editValue = '';
	}

	async function applyChanges() {
		if (pendingChangeCount === 0) return;

		applyingChanges = true;
		const errors: string[] = [];

		try {
			for (const [rowKey, colChanges] of pendingChanges) {
				const changes: Record<string, any> = {};
				colChanges.forEach((value, col) => {
					changes[col] = value;
				});

				try {
					await onSaveRow(rowKey, changes);
				} catch (e: any) {
					errors.push(`Row ${rowKey}: ${e.message || 'Update failed'}`);
				}
			}

			if (errors.length === 0) {
				// Clear all pending changes on success
				pendingChanges = new Map();
			} else {
				// Keep changes that failed
				console.error('Some updates failed:', errors);
			}
		} finally {
			applyingChanges = false;
		}
	}

	function discardChanges() {
		pendingChanges = new Map();
	}

	function discardRowChanges(rowKey: any) {
		pendingChanges.delete(rowKey);
		pendingChanges = new Map(pendingChanges);
	}

	function requestDelete() {
		if (selectedIds.size === 0) return;
		pendingDeleteIds = Array.from(selectedIds);
		showDeleteConfirm = true;
	}

	async function confirmDelete() {
		deleteLoading = true;
		try {
			await onDeleteRows(pendingDeleteIds);
			// Remove any pending changes for deleted rows
			pendingDeleteIds.forEach((id) => pendingChanges.delete(id));
			pendingChanges = new Map(pendingChanges);
			selectedIds = new Set();
			pendingDeleteIds = [];
		} finally {
			deleteLoading = false;
		}
	}

	function copyValue(value: any) {
		navigator.clipboard.writeText(String(value ?? ''));
	}

	function getDisplayValue(value: any): string {
		if (value === null || value === undefined) return 'NULL';
		if (typeof value === 'object') return JSON.stringify(value);
		return String(value);
	}

	function getCellStyle(value: any, isModified: boolean): string {
		if (isModified) return 'text-warning font-bold';
		if (value === null || value === undefined) return 'text-text-dim italic';
		if (typeof value === 'number') return 'text-rust-light font-jetbrains';
		if (typeof value === 'boolean') return value ? 'text-success' : 'text-danger';
		return 'text-stone-300';
	}
</script>

<div class="flex flex-col h-full bg-[var(--terminal-bg)]">
		<!-- Enhanced Toolbar -->
		<div
			class="border-b border-stone-800 bg-[var(--header-bg)] sticky top-0 z-20"
		>
			<div class="px-4 py-3 flex items-center justify-between gap-4">
				<!-- Left Actions -->
				<div class="flex items-center gap-2">
					<!-- Pending Changes Indicator -->
					{#if pendingChangeCount > 0}
						<div
							class="flex items-center gap-3 px-3 py-2 bg-warning/10 border border-warning/30"
							transition:slide={{ axis: 'x', duration: 150 }}
						>
							<div class="w-2 h-2 rounded-none bg-warning animate-pulse"></div>
							<span class="font-jetbrains text-[10px] font-black text-warning uppercase tracking-widest">
								{pendingChangeCount} UNSAVED_MODS
							</span>
							<button
								onclick={discardChanges}
								class="p-0.5 hover:bg-warning/20 text-warning transition-colors"
								title="Discard all changes"
							>
								<X class="w-3.5 h-3.5" />
							</button>
						</div>
	
						<!-- Apply Changes Button -->
						<button
							onclick={applyChanges}
							disabled={applyingChanges}
							class="flex items-center gap-2 px-4 py-2 bg-rust hover:bg-rust-light text-white font-heading font-black text-[10px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all disabled:opacity-20"
						>
							{#if applyingChanges}
								<RefreshCw class="w-3.5 h-3.5 animate-spin" />
								<span>Processing...</span>
							{:else}
								<Save class="w-3.5 h-3.5" />
								<span>Commit_Changes</span>
							{/if}
						</button>
	
						<div class="h-6 w-px bg-stone-800 mx-1"></div>
					{/if}
	
					<!-- Selection Info -->
					{#if selectedIds.size > 0}
						<div
							class="flex items-center gap-3 px-3 py-2 bg-rust/10 border border-rust/20"
							transition:slide={{ axis: 'x', duration: 150 }}
						>
							<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-widest">
								{selectedIds.size} TARGETS
							</span>
							<button
								onclick={() => (selectedIds = new Set())}
								class="p-0.5 hover:bg-rust/20 text-rust transition-colors"
							>
								<X class="w-3.5 h-3.5" />
							</button>
						</div>
					{/if}
	
					<!-- Delete Button -->
					<button
						disabled={selectedIds.size === 0}
						onclick={requestDelete}
						class="flex items-center gap-2 px-3 py-2 transition-all {selectedIds.size >
						0
							? 'bg-danger/10 text-danger hover:bg-danger/20 border border-danger/20'
							: 'text-text-dim cursor-not-allowed bg-stone-900/50 border border-stone-800'}"
					>
						<Trash2 class="w-4 h-4" />
						<span class="font-heading text-[10px] font-black uppercase tracking-widest hidden sm:inline">Terminate</span>
					</button>
	
					<div class="h-6 w-px bg-stone-800 mx-1 hidden sm:block"></div>
	
					<!-- Search Toggle -->
					<button
						onclick={() => {
							showSearch = !showSearch;
							if (!showSearch) searchQuery = '';
						}}
						class="p-2 transition-colors {showSearch
							? 'bg-rust/10 text-rust'
							: 'text-text-dim hover:text-stone-300 hover:bg-stone-900/50'}"
					>
						<Search class="w-4 h-4" />
					</button>
	
					<!-- Filter Button (placeholder) -->
					<button
						class="p-2 text-text-dim hover:text-stone-300 hover:bg-stone-900/50 transition-colors"
					>
						<Filter class="w-4 h-4" />
					</button>
				</div>
	
				<!-- Right Actions -->
				<div class="flex items-center gap-2">
					{#if onRefresh}
						<button
							onclick={onRefresh}
							disabled={isLoading}
							class="p-2 text-text-dim hover:text-stone-300 hover:bg-stone-900/50 transition-colors disabled:opacity-20"
							title="Refresh"
						>
							<RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
						</button>
					{/if}
	
					<button
						onclick={onAddRow}
						class="flex items-center gap-2 px-4 py-2 bg-stone-100 text-stone-950 hover:bg-white font-heading font-black text-[10px] uppercase tracking-widest shadow-lg transition-all active:translate-y-px"
					>
						<Plus class="w-4 h-4" />
						<span>Insert_Row</span>
					</button>
				</div>
			</div>

		<!-- Search Bar -->
		{#if showSearch}
			<div class="px-4 pb-4" transition:slide={{ duration: 150 }}>
				<div class="relative">
					<Search
						class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-text-dim pointer-events-none"
					/>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="FILTER_SUBJECTS_IN_ALL_CHANNELS..."
						class="w-full bg-stone-950 border border-stone-800 py-3 pl-12 pr-12 font-jetbrains text-[11px] text-stone-200 placeholder:text-stone-700 focus:border-rust outline-none transition-all uppercase tracking-widest"
					/>
					{#if searchQuery}
						<button
							onclick={() => (searchQuery = '')}
							class="absolute right-4 top-1/2 -translate-y-1/2 text-text-dim hover:text-white"
						>
							<X class="w-4 h-4" />
						</button>
					{/if}
				</div>
				{#if searchQuery && filteredData.length !== data.length}
					<p class="font-jetbrains text-[9px] text-text-dim mt-2 px-1 uppercase tracking-widest">
						Filtering: {filteredData.length} // TOTAL: {data.length} Subjects
					</p>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Grid Container -->
	<div class="flex-1 overflow-auto relative custom-scrollbar">
		{#if isLoading}
			<div
				class="absolute inset-0 bg-[var(--terminal-bg)]/80 backdrop-blur-sm z-30 flex items-center justify-center"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex flex-col items-center gap-4">
					<div
						class="w-10 h-10 border-4 border-rust border-t-transparent rounded-none animate-spin"
					></div>
					<span class="font-heading text-[11px] font-black text-text-dim uppercase tracking-[0.2em]">Synchronizing_Data...</span>
				</div>
			</div>
		{/if}

		<table class="w-full text-left text-[11px] border-collapse font-jetbrains">
			<thead
				class="bg-[var(--header-bg)] text-text-dim sticky top-0 z-10 border-b border-stone-800"
			>
				<tr>
					<!-- Checkbox Column -->
					<th class="px-5 py-4 w-14">
						<button
							onclick={toggleSelectAll}
							class="flex items-center justify-center w-5 h-5 hover:text-rust transition-colors"
						>
							{#if selectedIds.size > 0 && selectedIds.size === filteredData.length}
								<CheckSquare class="w-5 h-5 text-rust" />
							{:else if selectedIds.size > 0}
								<div class="w-5 h-5 bg-rust rounded-none flex items-center justify-center shadow-[0_0_10px_rgba(120,53,15,0.4)]">
									<div class="w-2.5 h-0.5 bg-white"></div>
								</div>
							{:else}
								<Square class="w-5 h-5 text-stone-700" />
							{/if}
						</button>
					</th>

					<!-- Data Columns -->
					{#each columns as col (col.name)}
						<th
							class="px-5 py-4 font-black uppercase tracking-widest whitespace-nowrap cursor-pointer hover:bg-stone-900/50 transition-colors group border-r border-stone-800/30"
							onclick={() => handleSort(col.name)}
						>
							<div class="flex items-center gap-3">
								<!-- Primary Key Badge -->
								{#if col.name === primaryKeyColumn || col.isPk}
									<span
										class="px-1.5 py-0.5 text-[9px] font-black bg-rust/10 text-rust border border-rust/30"
									>
										PK
									</span>
								{/if}

								<span class="text-stone-400 group-hover:text-white transition-colors">{col.name}</span>

								<!-- Sort Indicator -->
								<div class="flex items-center">
									{#if sortCol === col.name}
										{#if sortDir === 'asc'}
											<ArrowUp class="w-3.5 h-3.5 text-rust" />
										{:else}
											<ArrowDown class="w-3.5 h-3.5 text-rust" />
										{/if}
									{:else}
										<ArrowUp
											class="w-3.5 h-3.5 opacity-0 group-hover:opacity-30 transition-opacity"
										/>
									{/if}
								</div>
							</div>

							<!-- Column Type -->
							<div class="text-[9px] text-text-dim font-bold mt-1 uppercase tracking-tighter">
								{col.type || 'PROTOCOL_UNKNOWN'}
							</div>
						</th>
					{/each}

					<!-- Actions Column -->
					<th class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 w-16"></th>
				</tr>
			</thead>

			<tbody class="divide-y divide-stone-900">
				{#each filteredData as row (getRowKey(row))}
					{@const rowKey = getRowKey(row)}
					{@const rowHasChanges = pendingChanges.has(rowKey)}
					<tr
						class="group transition-colors {rowHasChanges
							? 'bg-warning/5'
							: selectedIds.has(rowKey)
								? 'bg-rust/5'
								: 'hover:bg-stone-900/30'}"
					>
						<!-- Checkbox -->
						<td class="px-5 py-3">
							<button
								onclick={() => toggleSelect(rowKey)}
								class="flex items-center justify-center w-5 h-5 text-stone-700 hover:text-rust transition-colors"
							>
								{#if selectedIds.has(rowKey)}
									<CheckSquare class="w-5 h-5 text-rust" />
								{:else}
									<Square class="w-5 h-5" />
								{/if}
							</button>
						</td>

						<!-- Data Cells -->
						{#each columns as col (col.name)}
							{@const isModified = hasPendingChange(rowKey, col.name)}
							{@const cellValue = getCellValue(row, col.name)}
							{@const isPkCol = col.name === primaryKeyColumn || col.isPk}
							<td
								class="px-5 py-3 max-w-xs relative border-r border-stone-800/20 {isModified
									? 'bg-warning/10 border-l-2 border-warning'
									: ''}"
								onclick={() => !isPkCol && startEdit(row, col.name)}
							>
								{#if editingCell?.rowKey === rowKey && editingCell?.col === col.name}
									<!-- Edit Mode -->
									<div class="absolute inset-0 z-20 flex items-center px-2 bg-stone-900">
										<input
											type="text"
											bind:value={editValue}
											onblur={saveEdit}
											onkeydown={(e) => {
												if (e.key === 'Enter') saveEdit();
												if (e.key === 'Escape') cancelEdit();
											}}
											class="w-full h-8 bg-stone-950 text-white px-3 outline-none border border-rust font-jetbrains text-[11px]"
										/>
									</div>
								{:else}
									<!-- Display Mode -->
									<div
										class="flex items-center gap-3 group/cell {isPkCol
											? 'cursor-default'
											: 'cursor-text'}"
										title={getDisplayValue(cellValue)}
									>
										<span class="truncate {getCellStyle(cellValue, isModified)}">
											{getDisplayValue(cellValue)}
										</span>

										<!-- Modified indicator -->
										{#if isModified}
											<span
												class="px-1.5 py-0.5 text-[8px] font-black bg-warning text-stone-950 rounded-none shrink-0 tracking-widest"
											>
												MOD
											</span>
										{/if}

										<!-- Copy Button (on hover) -->
										{#if cellValue !== null}
											<button
												onclick={(e) => {
													e.stopPropagation();
													copyValue(cellValue);
												}}
												class="opacity-0 group-hover/cell:opacity-100 p-1 hover:bg-stone-800 text-text-dim hover:text-rust transition-all shrink-0"
												title="Copy value"
											>
												<Copy class="w-3 h-3" />
											</button>
										{/if}
									</div>
								{/if}
							</td>
						{/each}

						<!-- Row Actions -->
						<td class="px-5 py-3">
							<div
								class="flex items-center gap-2 {rowHasChanges
									? 'opacity-100'
									: 'opacity-0 group-hover:opacity-100'} transition-opacity justify-end"
							>
								{#if rowHasChanges}
									<button
										onclick={() => discardRowChanges(rowKey)}
										class="p-1.5 text-warning hover:text-warning hover:bg-warning/10 transition-colors"
										title="Discard row changes"
									>
										<Undo2 class="w-4 h-4" />
									</button>
								{/if}
								<button
									onclick={() => {
										pendingDeleteIds = [rowKey];
										showDeleteConfirm = true;
									}}
									class="p-1.5 text-text-dim hover:text-danger hover:bg-danger/10 transition-colors"
									title="Delete row"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							</div>
						</td>
					</tr>
				{/each}

				<!-- Empty State -->
				{#if filteredData.length === 0 && !isLoading}
					<tr>
						<td colspan={columns.length + 2} class="py-24">
							<div class="flex flex-col items-center justify-center text-text-dim">
								{#if searchQuery}
									<Search class="w-16 h-16 mb-6 opacity-10" />
									<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">No matching Subjects found</p>
									<p class="font-jetbrains text-[9px] mt-2 uppercase">Neural filters returned zero results</p>
									<button
										onclick={() => (searchQuery = '')}
										class="mt-8 px-6 py-3 bg-stone-900 border border-stone-800 hover:border-rust hover:text-rust text-[10px] font-heading font-black uppercase tracking-widest transition-all"
									>
										Reset_Filters
									</button>
								{:else}
									<div
										class="w-20 h-20 mb-6 border border-dashed border-stone-800 flex items-center justify-center industrial-frame"
									>
										<Plus class="w-8 h-8 opacity-20" />
									</div>
									<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">Database_Empty</p>
									<p class="font-jetbrains text-[9px] mt-2 uppercase">No data entities exist in this sector</p>
									<button
										onclick={onAddRow}
										class="mt-8 px-8 py-3 bg-rust hover:bg-rust-light text-white text-[10px] font-heading font-black uppercase tracking-widest transition-all shadow-lg shadow-rust/20"
									>
										Initialize_First_Row
									</button>
								{/if}
							</div>
						</td>
					</tr>
				{/if}
			</tbody>
		</table>
	</div>

	<!-- Enhanced Footer -->
	<div
		class="px-6 py-4 bg-[var(--header-bg)] border-t border-stone-800 flex items-center justify-between"
	>
		<div class="flex items-center gap-6 text-[10px] font-jetbrains font-bold uppercase tracking-widest">
			<span class="flex items-center gap-2">
				<div class="w-2 h-2 bg-success shadow-[0_0_8px_rgba(16,185,129,0.4)]"></div>
				<span class="text-stone-400">{filteredData.length} Subjects_Detected</span>
				{#if searchQuery}
					<span class="text-rust/60 ml-2">Filtered_View</span>
				{/if}
			</span>

			{#if selectedIds.size > 0}
				<span class="text-rust">
					{selectedIds.size} TARGETS_LOCKED
				</span>
			{/if}

			{#if pendingChangeCount > 0}
				<span class="text-warning flex items-center gap-2">
					<div class="w-2 h-2 bg-warning animate-pulse"></div>
					{pendingChangeCount} PENDING_MODS
				</span>
			{/if}
		</div>

		<div class="flex items-center gap-4 text-[9px] font-jetbrains font-bold text-text-dim uppercase tracking-widest hidden lg:flex">
			<span>Click_Cell_To_Edit</span>
			<span class="w-1 h-1 bg-stone-800"></span>
			<span>Enter_To_Commit</span>
			<span class="w-1 h-1 bg-stone-800"></span>
			<span>Esc_To_Abort</span>
			{#if pendingChangeCount > 0}
				<span class="w-1 h-1 bg-stone-800"></span>
				<span class="text-warning/60 animate-pulse">Execute_Apply_To_Save</span>
			{/if}
		</div>
	</div>
</div>

<ConfirmModal
	bind:isOpen={showDeleteConfirm}
	title={pendingDeleteIds.length === 1 ? 'Terminate_Subject' : `Purge_${pendingDeleteIds.length}_Subjects`}
	message={pendingDeleteIds.length === 1
		? 'Are you certain you wish to terminate this data subject? This operation is irreversible.'
		: `Are you certain you wish to purge these ${pendingDeleteIds.length} data subjects? This operation is irreversible.`}
	confirmText="Purge_Data"
	cancelText="Abort"
	variant="danger"
	loading={deleteLoading}
	onConfirm={confirmDelete}
	onCancel={() => (pendingDeleteIds = [])}
/>
