<script lang="ts">
	import { onMount } from 'svelte';
	import DataGrid from './DataGrid.svelte';
	import TableDesigner from './TableDesigner.svelte';
	import RowEditorModal from './RowEditorModal.svelte';
	import { notifications } from '$lib/stores';
	import { Settings, Table, RefreshCw, Database, Rows3, Hash } from 'lucide-svelte';

	let { schema, table } = $props<{ schema: string; table: string }>();

	let view = $state<'data' | 'structure'>('data');
	let rows = $state<any[]>([]);
	let columns = $state<any[]>([]);
	let primaryKeyColumn = $state<string>('id');
	let loading = $state(false);
	let limit = $state(50);
	let offset = $state(0);
	let totalCount = $state(0);

	// Row Editor State
	let isRowEditorOpen = $state(false);
	let editingRowData = $state<any>(null);

	async function loadData() {
		loading = true;
		try {
			// Load Columns first to know structure
			const colRes = await fetch(
				`/api/database/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`
			);
			if (colRes.ok) {
				const cols = await colRes.json();
				if (cols) {
					columns = cols.map((c: any) => ({
						name: c.name,
						type: c.type,
						nullable: c.nullable === 'YES',
						isPk: c.is_pk === true
					}));
					// Find the primary key column
					const pkCol = cols.find((c: any) => c.is_pk === true);
					primaryKeyColumn = pkCol?.name || 'id';
				} else {
					columns = [];
					primaryKeyColumn = 'id';
				}
			}

			// Load Rows
			const res = await fetch(
				`/api/database/table/${table}?schema=${encodeURIComponent(schema)}&limit=${limit}&offset=${offset}`
			);
			if (res.ok) {
				const data = await res.json();
				rows = data.data || [];
				totalCount = data.total || rows.length;
			}
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Failed to load table data',
				details: e.message
			});
		} finally {
			loading = false;
		}
	}

	async function saveRow(pkValue: any, changes: any) {
		try {
			const res = await fetch(
				`/api/database/table/${table}/${encodeURIComponent(pkValue)}?schema=${encodeURIComponent(schema)}&pk=${encodeURIComponent(primaryKeyColumn)}`,
				{
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(changes)
				}
			);
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Update failed');
			}
			notifications.add({ type: 'success', message: 'Row updated successfully' });
			loadData();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Save failed', details: e.message });
		}
	}

	async function deleteRows(ids: any[]) {
		try {
			for (const pkValue of ids) {
				const res = await fetch(
					`/api/database/table/${table}/${encodeURIComponent(pkValue)}?schema=${encodeURIComponent(schema)}&pk=${encodeURIComponent(primaryKeyColumn)}`,
					{ method: 'DELETE' }
				);
				if (!res.ok) {
					const err = await res.json();
					throw new Error(err.error || 'Delete failed');
				}
			}
			notifications.add({
				type: 'success',
				message: `Deleted ${ids.length} ${ids.length === 1 ? 'row' : 'rows'}`
			});
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
			throw e;
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<div class="flex flex-col h-full bg-[#050505]">
	<!-- Header -->
	<div class="border-b border-stone-800 bg-[#0a0a0a]">
		<!-- Table Info Bar -->
		<div class="px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
					<Database class="w-5 h-5 text-rust-light" />
				</div>
				<div>
					<h2 class="text-base font-heading font-black text-slate-100 flex items-center gap-3 uppercase tracking-tighter">
						{table}
						<span class="font-jetbrains text-[9px] font-black text-rust bg-rust/10 px-2 py-0.5 border border-rust/20 tracking-widest uppercase">
							{schema}
						</span>
					</h2>
					<div class="flex items-center gap-4 text-[10px] font-jetbrains font-bold text-stone-500 mt-1.5 uppercase tracking-widest">
						<span class="flex items-center gap-1.5">
							<Rows3 class="w-3.5 h-3.5" />
							{totalCount} rows
						</span>
						<span class="flex items-center gap-1.5">
							<Hash class="w-3.5 h-3.5" />
							{columns.length} columns
						</span>
					</div>
				</div>
			</div>

			<button
				onclick={loadData}
				disabled={loading}
				class="p-2.5 text-stone-500 hover:text-rust hover:bg-rust/10 transition-all disabled:opacity-20"
				title="Refresh Data"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>

		<!-- View Tabs -->
		<div class="flex items-center gap-1 px-6">
			<button
				onclick={() => (view = 'data')}
				class="px-6 py-3 text-[11px] font-heading font-black uppercase tracking-[0.2em] border-b-2 transition-all flex items-center gap-3 {view ===
				'data'
					? 'border-rust text-rust bg-rust/5'
					: 'border-transparent text-stone-600 hover:text-stone-300 hover:bg-stone-900/50'}"
			>
				<Table class="w-4 h-4" />
				Data
			</button>
			<button
				onclick={() => (view = 'structure')}
				class="px-6 py-3 text-[11px] font-heading font-black uppercase tracking-[0.2em] border-b-2 transition-all flex items-center gap-3 {view ===
				'structure'
					? 'border-rust text-rust bg-rust/5'
					: 'border-transparent text-stone-600 hover:text-stone-300 hover:bg-stone-900/50'}"
			>
				<Settings class="w-4 h-4" />
				Structure
			</button>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-hidden">
		{#if view === 'data'}
			<DataGrid
				data={rows}
				{columns}
				{primaryKeyColumn}
				isLoading={loading}
				onSort={(col, dir) => {
					/* TODO: Backend sort */
				}}
				onSaveRow={saveRow}
				onDeleteRows={deleteRows}
				onAddRow={addRow}
				onRefresh={loadData}
			/>
		{:else}
			<TableDesigner
				{schema}
				{table}
				{columns}
				onClose={() => (view = 'data')}
				onReload={loadData}
			/>
		{/if}
	</div>

	<!-- Row Editor Modal -->
	<RowEditorModal
		bind:isOpen={isRowEditorOpen}
		{schema}
		{table}
		{columns}
		rowData={editingRowData}
		onClose={() => (isRowEditorOpen = false)}
		onSave={handleRowSave}
	/>
</div>
