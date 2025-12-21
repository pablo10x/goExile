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
		if (isModified) return 'text-orange-300 font-medium';
		if (value === null || value === undefined) return 'text-slate-500 italic';
		if (typeof value === 'number') return 'text-emerald-400 font-mono';
		if (typeof value === 'boolean') return value ? 'text-green-400' : 'text-red-400';
		return 'text-slate-700 dark:text-slate-300';
	}
</script>

<div class="flex flex-col h-full bg-slate-900">
	<!-- Enhanced Toolbar -->
	<div class="border-b border-slate-200 dark:border-slate-800 bg-slate-900/80 backdrop-blur-sm sticky top-0 z-20">
		<div class="p-3 flex items-center justify-between gap-4">
			<!-- Left Actions -->
			<div class="flex items-center gap-2">
				<!-- Pending Changes Indicator -->
				{#if pendingChangeCount > 0}
					<div
						class="flex items-center gap-2 px-3 py-1.5 bg-orange-500/10 border border-orange-500/30 rounded-lg"
						transition:slide={{ axis: 'x', duration: 150 }}
					>
						<div class="w-2 h-2 rounded-full bg-orange-500 animate-pulse"></div>
						<span class="text-sm font-medium text-orange-400">
							{pendingChangeCount} unsaved {pendingChangeCount === 1 ? 'change' : 'changes'}
						</span>
						<button
							onclick={discardChanges}
							class="p-0.5 hover:bg-orange-500/20 rounded"
							title="Discard all changes"
						>
							<X class="w-3.5 h-3.5 text-orange-400" />
						</button>
					</div>

					<!-- Apply Changes Button -->
					<button
						onclick={applyChanges}
						disabled={applyingChanges}
						class="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-slate-900 dark:text-white rounded-lg text-sm font-bold shadow-lg shadow-orange-900/25 transition-all hover:shadow-orange-900/40 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{#if applyingChanges}
							<RefreshCw class="w-4 h-4 animate-spin" />
							<span>Applying...</span>
						{:else}
							<Save class="w-4 h-4" />
							<span>Apply Changes</span>
						{/if}
					</button>

					<div class="h-6 w-px bg-slate-700 mx-1"></div>
				{/if}

				<!-- Selection Info -->
				{#if selectedIds.size > 0}
					<div
						class="flex items-center gap-2 px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-lg"
						transition:slide={{ axis: 'x', duration: 150 }}
					>
						<span class="text-sm font-medium text-blue-400">
							{selectedIds.size} selected
						</span>
						<button
							onclick={() => (selectedIds = new Set())}
							class="p-0.5 hover:bg-blue-500/20 rounded"
						>
							<X class="w-3.5 h-3.5 text-blue-400" />
						</button>
					</div>
				{/if}

				<!-- Delete Button -->
				<button
					disabled={selectedIds.size === 0}
					onclick={requestDelete}
					class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all {selectedIds.size >
					0
						? 'bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20'
						: 'text-slate-600 cursor-not-allowed bg-slate-800/50'}"
				>
					<Trash2 class="w-4 h-4" />
					<span class="hidden sm:inline">Delete</span>
				</button>

				<div class="h-6 w-px bg-slate-700 mx-1 hidden sm:block"></div>

				<!-- Search Toggle -->
				<button
					onclick={() => {
						showSearch = !showSearch;
						if (!showSearch) searchQuery = '';
					}}
					class="p-2 rounded-lg transition-colors {showSearch
						? 'bg-indigo-500/10 text-indigo-400'
						: 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800'}"
				>
					<Search class="w-4 h-4" />
				</button>

				<!-- Filter Button (placeholder) -->
				<button
					class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800 rounded-lg transition-colors"
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
						class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800 rounded-lg transition-colors disabled:opacity-50"
						title="Refresh"
					>
						<RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
					</button>
				{/if}

				<button
					onclick={onAddRow}
					class="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-slate-900 dark:text-white rounded-lg text-sm font-bold shadow-lg shadow-blue-900/25 transition-all hover:shadow-blue-900/40 hover:scale-[1.02] active:scale-[0.98]"
				>
					<Plus class="w-4 h-4" />
					<span>Add Row</span>
				</button>
			</div>
		</div>

		<!-- Search Bar -->
		{#if showSearch}
			<div class="px-3 pb-3" transition:slide={{ duration: 150 }}>
				<div class="relative">
					<Search
						class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none"
					/>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search in all columns..."
						class="w-full bg-slate-800/50 border border-slate-300 dark:border-slate-700 rounded-lg py-2 pl-10 pr-10 text-sm text-slate-800 dark:text-slate-200 placeholder:text-slate-500 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/50 outline-none transition-all"
					/>
					{#if searchQuery}
						<button
							onclick={() => (searchQuery = '')}
							class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-900 dark:text-white"
						>
							<X class="w-4 h-4" />
						</button>
					{/if}
				</div>
				{#if searchQuery && filteredData.length !== data.length}
					<p class="text-xs text-slate-500 mt-2 px-1">
						Showing {filteredData.length} of {data.length} rows
					</p>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Grid Container -->
	<div class="flex-1 overflow-auto relative">
		{#if isLoading}
			<div
				class="absolute inset-0 bg-slate-900/70 backdrop-blur-sm z-30 flex items-center justify-center"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex flex-col items-center gap-3">
					<div
						class="w-10 h-10 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"
					></div>
					<span class="text-sm text-slate-500 dark:text-slate-400">Loading data...</span>
				</div>
			</div>
		{/if}

		<table class="w-full text-left text-sm border-collapse">
			<thead class="bg-white/90 dark:bg-slate-950/90 backdrop-blur-sm text-slate-500 dark:text-slate-400 sticky top-0 z-10">
				<tr>
					<!-- Checkbox Column -->
					<th class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 w-12">
						<button
							onclick={toggleSelectAll}
							class="flex items-center justify-center w-5 h-5 hover:text-blue-400 transition-colors"
						>
							{#if selectedIds.size > 0 && selectedIds.size === filteredData.length}
								<CheckSquare class="w-[18px] h-[18px] text-blue-500" />
							{:else if selectedIds.size > 0}
								<div class="w-[18px] h-[18px] bg-blue-500 rounded flex items-center justify-center">
									<div class="w-2.5 h-0.5 bg-white rounded-full"></div>
								</div>
							{:else}
								<Square class="w-[18px] h-[18px]" />
							{/if}
						</button>
					</th>

					<!-- Data Columns -->
					{#each columns as col (col.name)}
						<th
							class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 font-medium whitespace-nowrap cursor-pointer hover:bg-slate-800/50 transition-colors group"
							onclick={() => handleSort(col.name)}
						>
							<div class="flex items-center gap-2">
								<!-- Primary Key Badge -->
								{#if col.name === primaryKeyColumn || col.isPk}
									<span
										class="px-1.5 py-0.5 text-[10px] font-bold bg-amber-500/20 text-amber-400 rounded"
									>
										PK
									</span>
								{/if}

								<span class="text-slate-700 dark:text-slate-300">{col.name}</span>

								<!-- Sort Indicator -->
								<div class="flex items-center">
									{#if sortCol === col.name}
										{#if sortDir === 'asc'}
											<ArrowUp class="w-3.5 h-3.5 text-indigo-400" />
										{:else}
											<ArrowDown class="w-3.5 h-3.5 text-indigo-400" />
										{/if}
									{:else}
										<ArrowUp
											class="w-3.5 h-3.5 opacity-0 group-hover:opacity-30 transition-opacity"
										/>
									{/if}
								</div>
							</div>

							<!-- Column Type -->
							<div class="text-[10px] text-slate-600 font-normal font-mono mt-0.5">
								{col.type || 'unknown'}
							</div>
						</th>
					{/each}

					<!-- Actions Column -->
					<th class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 w-16"></th>
				</tr>
			</thead>

			<tbody class="divide-y divide-slate-800/50">
				{#each filteredData as row (getRowKey(row))}
					{@const rowKey = getRowKey(row)}
					{@const rowHasChanges = pendingChanges.has(rowKey)}
					<tr
						class="group transition-colors {rowHasChanges
							? 'bg-orange-500/5'
							: selectedIds.has(rowKey)
								? 'bg-blue-500/5'
								: 'hover:bg-slate-800/30'}"
					>
						<!-- Checkbox -->
						<td class="px-4 py-2.5">
							<button
								onclick={() => toggleSelect(rowKey)}
								class="flex items-center justify-center w-5 h-5 text-slate-500 hover:text-blue-400 transition-colors"
							>
								{#if selectedIds.has(rowKey)}
									<CheckSquare class="w-[18px] h-[18px] text-blue-500" />
								{:else}
									<Square class="w-[18px] h-[18px]" />
								{/if}
							</button>
						</td>

						<!-- Data Cells -->
						{#each columns as col (col.name)}
							{@const isModified = hasPendingChange(rowKey, col.name)}
							{@const cellValue = getCellValue(row, col.name)}
							{@const isPkCol = col.name === primaryKeyColumn || col.isPk}
							<td
								class="px-4 py-2.5 max-w-xs relative {isModified
									? 'bg-orange-500/10 border-l-2 border-orange-500'
									: ''}"
								onclick={() => !isPkCol && startEdit(row, col.name)}
							>
								{#if editingCell?.rowKey === rowKey && editingCell?.col === col.name}
									<!-- Edit Mode -->
									<div class="absolute inset-0 z-20 flex items-center px-2">
										<input
											type="text"
											bind:value={editValue}
											onblur={saveEdit}
											onkeydown={(e) => {
												if (e.key === 'Enter') saveEdit();
												if (e.key === 'Escape') cancelEdit();
											}}
											class="w-full h-full bg-slate-800 text-slate-900 dark:text-white px-3 py-1 outline-none border-2 border-indigo-500 rounded-md text-sm"
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
												class="px-1 py-0.5 text-[9px] font-bold bg-orange-500/20 text-orange-400 rounded shrink-0"
											>
												MODIFIED
											</span>
										{/if}

										<!-- Copy Button (on hover) -->
										{#if cellValue !== null}
											<button
												onclick={(e) => {
													e.stopPropagation();
													copyValue(cellValue);
												}}
												class="opacity-0 group-hover/cell:opacity-100 p-1 hover:bg-slate-700 rounded transition-all shrink-0"
												title="Copy value"
											>
												<Copy class="w-3 h-3 text-slate-500" />
											</button>
										{/if}
									</div>
								{/if}
							</td>
						{/each}

						<!-- Row Actions -->
						<td class="px-4 py-2.5">
							<div
								class="flex items-center gap-1 {rowHasChanges
									? 'opacity-100'
									: 'opacity-0 group-hover:opacity-100'} transition-opacity"
							>
								{#if rowHasChanges}
									<button
										onclick={() => discardRowChanges(rowKey)}
										class="p-1.5 text-orange-400 hover:text-orange-300 hover:bg-orange-500/10 rounded transition-colors"
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
									class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded transition-colors"
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
						<td colspan={columns.length + 2} class="py-16">
							<div class="flex flex-col items-center justify-center text-slate-500">
								{#if searchQuery}
									<Search class="w-12 h-12 mb-4 opacity-20" />
									<p class="text-lg font-medium">No matching rows</p>
									<p class="text-sm mt-1">Try adjusting your search query</p>
									<button
										onclick={() => (searchQuery = '')}
										class="mt-4 px-4 py-2 bg-slate-800 hover:bg-slate-700 rounded-lg text-sm text-slate-700 dark:text-slate-300 transition-colors"
									>
										Clear Search
									</button>
								{:else}
									<div
										class="w-16 h-16 mb-4 rounded-full bg-slate-800 flex items-center justify-center"
									>
										<Plus class="w-8 h-8 opacity-40" />
									</div>
									<p class="text-lg font-medium">No rows yet</p>
									<p class="text-sm mt-1">Add your first row to get started</p>
									<button
										onclick={onAddRow}
										class="mt-4 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm text-slate-900 dark:text-white font-medium transition-colors flex items-center gap-2"
									>
										<Plus class="w-4 h-4" />
										Add Row
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
	<div class="px-4 py-2.5 bg-white dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
		<div class="flex items-center gap-4 text-xs text-slate-500">
			<span class="flex items-center gap-1.5">
				<div class="w-2 h-2 rounded-full bg-emerald-500"></div>
				{filteredData.length}
				{filteredData.length === 1 ? 'row' : 'rows'}
				{#if searchQuery}
					<span class="text-slate-600">(filtered)</span>
				{/if}
			</span>

			{#if selectedIds.size > 0}
				<span class="text-blue-400">
					{selectedIds.size} selected
				</span>
			{/if}

			{#if pendingChangeCount > 0}
				<span class="text-orange-400 flex items-center gap-1.5">
					<div class="w-2 h-2 rounded-full bg-orange-500 animate-pulse"></div>
					{pendingChangeCount} pending {pendingChangeCount === 1 ? 'change' : 'changes'}
				</span>
			{/if}
		</div>

		<div class="flex items-center gap-2 text-xs text-slate-600">
			<span>Click cell to edit</span>
			<span>•</span>
			<span>Enter to confirm</span>
			<span>•</span>
			<span>Esc to cancel</span>
			{#if pendingChangeCount > 0}
				<span>•</span>
				<span class="text-orange-400">Apply to save</span>
			{/if}
		</div>
	</div>
</div>

<!-- Delete Confirmation Modal -->
<ConfirmModal
	bind:isOpen={showDeleteConfirm}
	title={pendingDeleteIds.length === 1 ? 'Delete Row' : `Delete ${pendingDeleteIds.length} Rows`}
	message={pendingDeleteIds.length === 1
		? 'Are you sure you want to delete this row? This action cannot be undone.'
		: `Are you sure you want to delete these ${pendingDeleteIds.length} rows? This action cannot be undone.`}
	confirmText="Delete"
	cancelText="Cancel"
	variant="danger"
	loading={deleteLoading}
	onConfirm={confirmDelete}
	onCancel={() => (pendingDeleteIds = [])}
/>
