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
		if (isModified) return 'text-yellow-400 font-black';
		if (value === null || value === undefined) return 'text-stone-700 italic';
		if (typeof value === 'number') return 'text-yellow-500/80 font-mono';
		if (typeof value === 'boolean') return value ? 'text-emerald-500' : 'text-red-500';
		return 'text-stone-300';
	}
</script>

<div class="flex flex-col h-full bg-transparent">
		<!-- Enhanced Toolbar -->
		<div
			class="border-b border-slate-800 bg-slate-900/40 backdrop-blur-md sticky top-0 z-20"
		>
			<div class="px-4 py-3 flex items-center justify-between gap-4">
				<!-- Left Actions -->
				<div class="flex items-center gap-2">
					<!-- Pending Changes Indicator -->
					{#if pendingChangeCount > 0}
						<div
							class="flex items-center gap-3 px-3 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-xl"
							transition:slide={{ axis: 'x', duration: 150 }}
						>
							<div class="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-pulse"></div>
							<span class="text-[10px] font-bold text-indigo-400 uppercase tracking-widest">
								{pendingChangeCount} Staged
							</span>
							<button
								onclick={discardChanges}
								class="p-0.5 hover:bg-indigo-500/20 text-indigo-400 transition-colors rounded"
								title="Discard all changes"
							>
								<X class="w-3.5 h-3.5" />
							</button>
						</div>
	
						<!-- Apply Changes Button -->
						<button
							onclick={applyChanges}
							disabled={applyingChanges}
							class="flex items-center gap-3 px-5 py-2 bg-indigo-500 hover:bg-indigo-400 text-white font-bold text-[10px] uppercase tracking-widest shadow-lg shadow-indigo-500/20 transition-all disabled:opacity-50 rounded-xl"
						>
							{#if applyingChanges}
								<RefreshCw class="w-3.5 h-3.5 animate-spin" />
								<span>COMMITTING...</span>
							{:else}
								<Save class="w-3.5 h-3.5" />
								<span>Push Changes</span>
							{/if}
						</button>
	
						<div class="h-6 w-px bg-slate-800 mx-1"></div>
					{/if}
	
					<!-- Selection Info -->
					{#if selectedIds.size > 0}
						<div
							class="flex items-center gap-3 px-3 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-xl"
							transition:slide={{ axis: 'x', duration: 150 }}
						>
							<span class="text-[10px] font-bold text-indigo-400 uppercase tracking-widest">
								{selectedIds.size} Selected
							</span>
							<button
								onclick={() => (selectedIds = new Set())}
								class="p-0.5 hover:bg-indigo-500/20 text-indigo-400 transition-colors rounded"
							>
								<X class="w-3.5 h-3.5" />
							</button>
						</div>
					{/if}
	
					<!-- Delete Button -->
					<button
						disabled={selectedIds.size === 0}
						onclick={requestDelete}
						class="flex items-center gap-2 px-3 py-2 transition-all rounded-xl {selectedIds.size >
						0
							? 'bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20'
							: 'text-slate-600 cursor-not-allowed border border-transparent'}"
					>
						<Trash2 class="w-4 h-4" />
						<span class="text-[10px] font-bold uppercase tracking-widest hidden sm:inline">Purge</span>
					</button>
	
					<div class="h-6 w-px bg-slate-800 mx-1 hidden sm:block"></div>
	
					<!-- Search Toggle -->
					<button
						onclick={() => {
							showSearch = !showSearch;
							if (!showSearch) searchQuery = '';
						}}
						class="p-2 transition-colors rounded-xl {showSearch
							? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
							: 'text-slate-500 hover:text-slate-200 hover:bg-slate-800'}"
					>
						<Search class="w-4 h-4" />
					</button>
	
					<!-- Filter Button (placeholder) -->
					<button
						class="p-2 text-slate-500 hover:text-slate-200 hover:bg-slate-800 transition-colors rounded-xl"
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
							class="p-2 text-slate-500 hover:text-slate-200 hover:bg-slate-800 transition-colors disabled:opacity-20 rounded-xl"
							title="Refresh"
						>
							<RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
						</button>
					{/if}
	
					<button
						onclick={onAddRow}
						class="flex items-center gap-3 px-5 py-2 bg-slate-800 text-white hover:bg-slate-700 font-bold text-[10px] uppercase tracking-widest shadow-lg transition-all active:translate-y-px rounded-xl border border-slate-700"
					>
						<Plus class="w-4 h-4" />
						<span>Initialize Row</span>
					</button>
				</div>
			</div>

		<!-- Search Bar -->
		{#if showSearch}
			<div class="px-4 pb-4" transition:slide={{ duration: 150 }}>
				<div class="relative">
					<Search
						class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none"
					/>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search cluster data..."
						class="w-full bg-slate-950/40 border border-slate-800 py-2.5 pl-10 pr-10 text-xs font-bold uppercase tracking-widest text-slate-300 placeholder:text-slate-700 focus:border-indigo-500/50 outline-none transition-all rounded-xl"
					/>
					{#if searchQuery}
						<button
							onclick={() => (searchQuery = '')}
							class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
						>
							<X class="w-4 h-4" />
						</button>
					{/if}
				</div>
				{#if searchQuery && filteredData.length !== data.length}
					<p class="text-[9px] text-slate-500 mt-2 px-1 font-bold uppercase tracking-widest italic">
						Filtered: {filteredData.length} / {data.length} records mapped
					</p>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Grid Container -->
	<div class="flex-1 overflow-auto relative custom-scrollbar bg-transparent">
		{#if isLoading}
			<div
				class="absolute inset-0 bg-slate-950/60 backdrop-blur-sm z-30 flex items-center justify-center"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex flex-col items-center gap-4">
					<div
						class="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(99,102,241,0.4)]"
					></div>
					<span class="text-[10px] font-bold text-slate-400 uppercase tracking-[0.3em] animate-pulse">Syncing_Data_Stream...</span>
				</div>
			</div>
		{/if}

		<table class="w-full text-left text-xs border-collapse font-jetbrains">
			<thead
				class="bg-slate-900 sticky top-0 z-10 border-b border-slate-800 backdrop-blur-md"
			>
				<tr>
					<!-- Checkbox Column -->
					<th class="px-4 py-3 w-12 text-center">
						<button
							onclick={toggleSelectAll}
							class="flex items-center justify-center w-4 h-4 hover:text-indigo-400 transition-colors mx-auto"
						>
							{#if selectedIds.size > 0 && selectedIds.size === filteredData.length}
								<CheckSquare class="w-4 h-4 text-indigo-400" />
							{:else if selectedIds.size > 0}
								<div class="w-4 h-4 bg-indigo-400 rounded-sm flex items-center justify-center">
									<div class="w-2 h-0.5 bg-slate-950"></div>
								</div>
							{:else}
								<Square class="w-4 h-4 text-slate-700" />
							{/if}
						</button>
					</th>

					<!-- Data Columns -->
					{#each columns as col (col.name)}
						<th
							class="px-4 py-3 font-bold uppercase tracking-widest whitespace-nowrap cursor-pointer hover:bg-slate-800/50 transition-colors group border-r border-slate-800/50"
							onclick={() => handleSort(col.name)}
						>
							<div class="flex items-center gap-3">
								<!-- Primary Key Badge -->
								{#if col.name === primaryKeyColumn || col.isPk}
									<span
										class="px-1.5 py-0.5 text-[8px] font-black bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 rounded-md italic"
									>
										PK
									</span>
								{/if}

								<span class="text-slate-400 group-hover:text-indigo-400 transition-colors">{col.name}</span>

								<!-- Sort Indicator -->
								<div class="flex items-center">
									{#if sortCol === col.name}
										{#if sortDir === 'asc'}
											<ArrowUp class="w-3 h-3 text-indigo-400" />
										{:else}
											<ArrowDown class="w-3 h-3 text-indigo-400" />
										{/if}
									{:else}
										<ArrowUp
											class="w-3 h-3 opacity-0 group-hover:opacity-30 transition-opacity"
										/>
									{/if}
								</div>
							</div>

							<!-- Column Type -->
							<div class="text-[8px] text-slate-600 font-medium mt-1 uppercase tracking-tighter italic">
								{col.type || 'UNKNOWN_TYPE'}
							</div>
						</th>
					{/each}

					<!-- Actions Column -->
					<th class="px-4 py-3 border-b border-slate-800 w-16"></th>
				</tr>
			</thead>

			<tbody class="divide-y divide-slate-800/50">
				{#each filteredData as row (getRowKey(row))}
					{@const rowKey = getRowKey(row)}
					{@const rowHasChanges = pendingChanges.has(rowKey)}
					<tr
						class="group transition-colors {rowHasChanges
							? 'bg-indigo-500/5'
							: selectedIds.has(rowKey)
								? 'bg-indigo-500/5'
								: 'hover:bg-slate-800/20'}"
					>
						<!-- Checkbox -->
						<td class="px-4 py-2.5 text-center">
							<button
								onclick={() => toggleSelect(rowKey)}
								class="flex items-center justify-center w-4 h-4 text-slate-700 hover:text-indigo-400 transition-colors mx-auto"
							>
								{#if selectedIds.has(rowKey)}
									<CheckSquare class="w-4 h-4 text-indigo-400" />
								{:else}
									<Square class="w-4 h-4" />
								{/if}
							</button>
						</td>

						<!-- Data Cells -->
						{#each columns as col (col.name)}
							{@const isModified = hasPendingChange(rowKey, col.name)}
							{@const cellValue = getCellValue(row, col.name)}
							{@const isPkCol = col.name === primaryKeyColumn || col.isPk}
							<td
								class="px-4 py-2.5 max-w-xs relative border-r border-slate-800/30 {isModified
									? 'bg-indigo-500/10 border-l-2 border-l-indigo-500'
									: ''}"
								onclick={() => !isPkCol && startEdit(row, col.name)}
							>
								{#if editingCell?.rowKey === rowKey && editingCell?.col === col.name}
									<!-- Edit Mode -->
									<div class="absolute inset-0 z-20 flex items-center px-2 bg-slate-900">
										<input
											type="text"
											bind:value={editValue}
											onblur={saveEdit}
											onkeydown={(e) => {
												if (e.key === 'Enter') saveEdit();
												if (e.key === 'Escape') cancelEdit();
											}}
											class="w-full h-8 bg-slate-950 text-white px-3 outline-none border border-indigo-500 rounded-lg font-mono text-xs shadow-[0_0_10px_rgba(99,102,241,0.2)]"
											autoFocus
										/>
									</div>
								{:else}
									<!-- Display Mode -->
									<div
										class="flex items-center gap-2 group/cell {isPkCol
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
												class="px-1 py-0.5 text-[8px] font-black bg-indigo-500 text-white rounded-md shrink-0 tracking-widest italic"
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
												class="opacity-0 group-hover/cell:opacity-100 p-1 hover:bg-slate-800 text-slate-600 hover:text-indigo-400 transition-all shrink-0 rounded-lg"
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
						<td class="px-4 py-2.5 text-right">
							<div
								class="flex items-center gap-1 {rowHasChanges
									? 'opacity-100'
									: 'opacity-0 group-hover:opacity-100'} transition-opacity justify-end"
							>
								{#if rowHasChanges}
									<button
										onclick={() => discardRowChanges(rowKey)}
										class="p-1.5 text-indigo-400 hover:bg-indigo-500/10 transition-colors rounded-lg"
										title="Discard row changes"
									>
										<Undo2 class="w-3.5 h-3.5" />
									</button>
								{/if}
								<button
									onclick={() => {
										pendingDeleteIds = [rowKey];
										showDeleteConfirm = true;
									}}
									class="p-1.5 text-slate-600 hover:text-red-400 hover:bg-red-500/10 transition-colors rounded-lg"
									title="Delete row"
								>
									<Trash2 class="w-3.5 h-3.5" />
								</button>
							</div>
						</td>
					</tr>
				{/each}

				<!-- Empty State -->
				{#if filteredData.length === 0 && !isLoading}
					<tr>
						<td colspan={columns.length + 2} class="py-24">
							<div class="flex flex-col items-center justify-center text-slate-700">
								{#if searchQuery}
									<Search class="w-16 h-16 mb-4 opacity-20" />
									<p class="font-bold text-[11px] uppercase tracking-widest">No matching records identified</p>
									<p class="text-[9px] mt-2 font-medium uppercase tracking-widest opacity-50">Adjust query parameters or sector filter</p>
									<button
										onclick={() => (searchQuery = '')}
										class="mt-8 px-6 py-2.5 bg-slate-900 border border-slate-800 hover:border-indigo-500/50 hover:text-indigo-400 text-[10px] font-bold uppercase tracking-[0.2em] transition-all rounded-xl"
									>
										Reset_Query
									</button>
								{:else}
									<div
										class="w-16 h-16 mb-6 border border-dashed border-slate-800 rounded-2xl flex items-center justify-center opacity-30"
									>
										<Plus class="w-8 h-8 opacity-30" />
									</div>
									<p class="font-bold text-[11px] uppercase tracking-widest">Sector buffer is empty</p>
									<p class="text-[9px] mt-2 font-medium uppercase tracking-widest opacity-50">No persistent entities recorded</p>
									<button
										onclick={onAddRow}
										class="mt-8 px-8 py-3 bg-indigo-500 hover:bg-indigo-400 text-white text-[10px] font-bold uppercase tracking-[0.2em] transition-all shadow-xl shadow-indigo-900/20 rounded-xl"
									>
										Initialize_Primary_Row
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
		class="px-6 py-3 bg-slate-900 border-t border-slate-800 flex items-center justify-between"
	>
		<div class="flex items-center gap-8 text-[9px] font-bold uppercase tracking-widest">
			<span class="flex items-center gap-3">
				<div class="w-1.5 h-1.5 bg-indigo-500 rounded-full shadow-[0_0_8px_rgba(99,102,241,0.6)]"></div>
				<span class="text-slate-500">{filteredData.length} ENTITIES_MAPPED</span>
				{#if searchQuery}
					<span class="text-indigo-400 italic">[FILTERED_MODE]</span>
				{/if}
			</span>

			{#if selectedIds.size > 0}
				<span class="text-indigo-400 bg-indigo-500/5 px-3 py-1 border border-indigo-500/20 rounded-lg">
					{selectedIds.size} TARGETS_LOCKED
				</span>
			{/if}

			{#if pendingChangeCount > 0}
				<span class="text-indigo-400 flex items-center gap-3 bg-indigo-500/5 px-3 py-1 border border-indigo-500/20 rounded-lg">
					<div class="w-1.5 h-1.5 bg-indigo-500 rounded-full animate-pulse shadow-[0_0_10px_rgba(99,102,241,0.6)]"></div>
					{pendingChangeCount} STAGED_MUTATIONS
				</span>
			{/if}
		</div>

		<div class="flex items-center gap-6 text-[8px] font-bold text-slate-600 uppercase tracking-[0.3em] hidden lg:flex italic">
			<span>Click_Edit</span>
			<span class="w-1 h-1 bg-slate-800 rounded-full"></span>
			<span>Ctrl+Enter_Commit</span>
			<span class="w-1 h-1 bg-slate-800 rounded-full"></span>
			<span>Esc_Abort</span>
		</div>
	</div>
</div>

<ConfirmModal
	bind:isOpen={showDeleteConfirm}
	title={pendingDeleteIds.length === 1 ? 'Delete Record' : `Delete ${pendingDeleteIds.length} Records`}
	message={pendingDeleteIds.length === 1
		? 'Are you sure you want to delete this record? This action cannot be undone.'
		: `Are you sure you want to delete these ${pendingDeleteIds.length} records? This action cannot be undone.`}
	confirmText="Delete"
	cancelText="Cancel"
	variant="danger"
	loading={deleteLoading}
	onConfirm={confirmDelete}
	onCancel={() => (pendingDeleteIds = [])}
/>