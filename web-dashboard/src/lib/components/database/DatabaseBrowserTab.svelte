<script lang="ts">
	import SchemaBrowser from '$lib/components/SchemaBrowser.svelte';
	import TableCreatorModal from '$lib/components/TableCreatorModal.svelte';
	import ColumnManagerModal from '$lib/components/ColumnManagerModal.svelte';
	import SchemaCreatorModal from '$lib/components/database/SchemaCreatorModal.svelte';
	import ConfirmModal from '$lib/components/database/ConfirmModal.svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';
	import { Database, RefreshCw, FolderTree, Table, Layers, Plus } from 'lucide-svelte';

	let { onSelectTable } = $props<{
		onSelectTable: (schema: string, table: string) => void;
	}>();

	// Internal State for Browser
	let schemas = $state<string[]>([]);
	let schemaTables = $state<Record<string, string[]>>({});
	let tableColumns = $state<Record<string, any[]>>({});
	let expandedSchemas = $state<Set<string>>(new Set());
	let expandedTables = $state<Set<string>>(new Set());
	let isSidebarOpen = $state(true);
	let isRefreshing = $state(false);

	// Modal States
	let isTableCreatorOpen = $state(false);
	let creatorSchema = $state('public');

	let isSchemaCreatorOpen = $state(false);

	let isColumnManagerOpen = $state(false);
	let columnManagerTarget = $state({ schema: '', table: '' });

	// Confirmation Modal States
	let showDeleteSchemaConfirm = $state(false);
	let deleteSchemaLoading = $state(false);
	let pendingDeleteSchema = $state('');

	let showDeleteTableConfirm = $state(false);
	let deleteTableLoading = $state(false);
	let pendingDeleteTable = $state({ schema: '', table: '' });

	// Stats
	let totalTables = $derived(Object.values(schemaTables).flat().length);
	let totalSchemas = $derived(schemas.length);

	async function loadDatabaseMap() {
		isRefreshing = true;
		try {
			// Fetch all tables at once
			const res = await fetch('/api/database/all-tables');
			if (res.ok) {
				const allTables: { schema: string; table: string }[] = await res.json();

				// Process into map
				const newSchemaTables: Record<string, string[]> = {};
				const newSchemasSet = new Set<string>();

				allTables.forEach((row) => {
					newSchemasSet.add(row.schema);
					if (!newSchemaTables[row.schema]) newSchemaTables[row.schema] = [];
					newSchemaTables[row.schema].push(row.table);
				});

				// Fetch schemas list to get empty schemas too
				const schemaRes = await fetch('/api/database/schemas');
				if (schemaRes.ok) {
					const schemaList: string[] = await schemaRes.json();
					schemaList.forEach((s) => newSchemasSet.add(s));
				}

				schemas = Array.from(newSchemasSet).sort();
				schemaTables = newSchemaTables;
			}
		} catch (e) {
			console.error(e);
			notifications.add({ type: 'error', message: 'Failed to load database structure' });
		} finally {
			isRefreshing = false;
		}
	}

	function toggleSchema(schema: string) {
		if (expandedSchemas.has(schema)) expandedSchemas.delete(schema);
		else {
			expandedSchemas.add(schema);
		}
		expandedSchemas = new Set(expandedSchemas);
	}

	function toggleTable(schema: string, table: string) {
		const key = `${schema}.${table}`;
		if (expandedTables.has(key)) {
			expandedTables.delete(key);
		} else {
			expandedTables.add(key);
		}
		expandedTables = new Set(expandedTables);
	}

	async function handleSchemaCreate(data: { name: string; owner: string }) {
		if (!data.name.trim()) return;
		try {
			const res = await fetch('/api/database/schemas', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
			if (res.ok) {
				notifications.add({
					type: 'success',
					message: `Schema '${data.name}' created successfully`
				});
				loadDatabaseMap();
			} else {
				const err = await res.json();
				throw new Error(err.error);
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to create schema', details: e.message });
		}
	}

	function requestDeleteSchema(name: string) {
		pendingDeleteSchema = name;
		showDeleteSchemaConfirm = true;
	}

	async function confirmDeleteSchema() {
		deleteSchemaLoading = true;
		try {
			const res = await fetch(`/api/database/schemas/${pendingDeleteSchema}`, { method: 'DELETE' });
			if (res.ok) {
				notifications.add({ type: 'success', message: `Schema '${pendingDeleteSchema}' deleted` });
				delete schemaTables[pendingDeleteSchema];
				loadDatabaseMap();
			} else {
				const err = await res.json();
				throw new Error(err.error);
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to delete schema', details: e.message });
		} finally {
			deleteSchemaLoading = false;
			pendingDeleteSchema = '';
		}
	}

	// Table Operations
	function openTableCreator(schema: string) {
		creatorSchema = schema;
		isTableCreatorOpen = true;
	}

	async function handleTableCreate(data: { name: string; columns: string[] }) {
		try {
			const res = await fetch('/api/database/tables/create', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ schema: creatorSchema, ...data })
			});
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error);
			}
			notifications.add({ type: 'success', message: `Table '${data.name}' created successfully` });
			loadDatabaseMap();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to create table', details: e.message });
			throw e;
		}
	}

	function requestDeleteTable(schema: string, table: string) {
		pendingDeleteTable = { schema, table };
		showDeleteTableConfirm = true;
	}

	async function confirmDeleteTable() {
		deleteTableLoading = true;
		try {
			const res = await fetch(
				`/api/database/tables/${pendingDeleteTable.table}?schema=${encodeURIComponent(pendingDeleteTable.schema)}`,
				{ method: 'DELETE' }
			);
			if (res.ok) {
				notifications.add({
					type: 'success',
					message: `Table '${pendingDeleteTable.table}' deleted`
				});
				loadDatabaseMap();
			} else {
				const err = await res.json();
				throw new Error(err.error);
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to delete table', details: e.message });
		} finally {
			deleteTableLoading = false;
			pendingDeleteTable = { schema: '', table: '' };
		}
	}

	// Column Operations
	function openColumnManager(schema: string, table: string) {
		columnManagerTarget = { schema, table };
		isColumnManagerOpen = true;
	}

	async function handleAddColumn(data: { column: string; type: string }) {
		try {
			const res = await fetch(`/api/database/tables/${columnManagerTarget.table}/alter`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					schema: columnManagerTarget.schema,
					action: 'add_column',
					column: data.column,
					type: data.type
				})
			});
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error);
			}
			notifications.add({ type: 'success', message: `Column '${data.column}' added successfully` });
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to add column', details: e.message });
			throw e;
		}
	}

	function refresh() {
		loadDatabaseMap();
	}

	onMount(() => {
		loadDatabaseMap();
	});
</script>

<div class="h-full flex flex-col bg-transparent">
	<!-- Header -->
	<div class="shrink-0 border-b border-neutral-800 bg-neutral-900/40 backdrop-blur-md">
		<!-- Title Bar -->
		<div class="px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<div class="p-2.5 bg-indigo-500/10 border border-indigo-500/20 rounded-xl shadow-sm">
					<FolderTree class="w-5 h-5 text-indigo-400" />
				</div>
				<div>
					<h2 class="text-sm font-heading font-black text-white tracking-widest uppercase italic">Schema_Explorer</h2>
					<p class="text-[9px] text-neutral-500 mt-1 font-bold uppercase tracking-widest italic">Database Topology Mapping</p>
				</div>
			</div>

			<div class="flex items-center gap-2">
				<button
					onclick={() => (isSchemaCreatorOpen = true)}
					class="p-2 text-neutral-500 hover:text-indigo-400 hover:bg-indigo-500/10 transition-all rounded-xl"
					title="New Schema"
				>
					<Plus class="w-5 h-5" />
				</button>
				<div class="w-px h-6 bg-neutral-800 mx-1"></div>
				<button
					onclick={refresh}
					disabled={isRefreshing}
					class="p-2 text-neutral-500 hover:text-white hover:bg-neutral-800 transition-all rounded-xl disabled:opacity-50"
					title="Refresh"
				>
					<RefreshCw class="w-5 h-5 {isRefreshing ? 'animate-spin' : ''}" />
				</button>
			</div>
		</div>

		<!-- Stats Bar -->
		<div class="px-6 pb-4 flex items-center gap-3">
			<div class="flex items-center gap-3 px-4 py-1.5 bg-neutral-950/40 border border-neutral-800 rounded-xl">
				<Layers class="w-3.5 h-3.5 text-indigo-400" />
				<span class="text-[9px] font-bold text-neutral-400 uppercase tracking-widest italic">{totalSchemas} Sectors</span>
			</div>
			<div class="flex items-center gap-3 px-4 py-1.5 bg-neutral-950/40 border border-neutral-800 rounded-xl">
				<Table class="w-3.5 h-3.5 text-indigo-400" />
				<span class="text-[9px] font-bold text-neutral-400 uppercase tracking-widest italic">{totalTables} Tables</span>
			</div>
		</div>
	</div>

	<!-- Browser Content -->
	<div class="flex-1 overflow-hidden">
		<SchemaBrowser
			{schemas}
			{schemaTables}
			{tableColumns}
			selectedSchema={null}
			selectedTable={null}
			{expandedSchemas}
			{expandedTables}
			{isSidebarOpen}
			onToggleSchema={toggleSchema}
			onToggleTable={toggleTable}
			{onSelectTable}
			onCreateSchema={() => (isSchemaCreatorOpen = true)}
			onDropSchema={requestDeleteSchema}
			onCreateTable={openTableCreator}
			onDropTable={requestDeleteTable}
			onEditTable={openColumnManager}
			onToggleSidebar={() => (isSidebarOpen = !isSidebarOpen)}
		/>
	</div>
</div>

<!-- Schema Creator Modal -->
<SchemaCreatorModal
	bind:isOpen={isSchemaCreatorOpen}
	onClose={() => (isSchemaCreatorOpen = false)}
	onSave={handleSchemaCreate}
/>

<!-- Table Creator Modal -->
<TableCreatorModal
	bind:isOpen={isTableCreatorOpen}
	schema={creatorSchema}
	onClose={() => (isTableCreatorOpen = false)}
	onSave={handleTableCreate}
/>

<!-- Column Manager Modal -->
<ColumnManagerModal
	bind:isOpen={isColumnManagerOpen}
	schema={columnManagerTarget.schema}
	table={columnManagerTarget.table}
	onClose={() => (isColumnManagerOpen = false)}
	onSave={handleAddColumn}
/>

<!-- Delete Schema Confirmation -->
<ConfirmModal
	bind:isOpen={showDeleteSchemaConfirm}
	title="Delete Schema"
	message={`Are you sure you want to delete the schema '${pendingDeleteSchema}'? This will permanently delete all tables and data within this schema. This action cannot be undone.`}
	confirmText="Delete Schema"
	cancelText="Cancel"
	variant="danger"
	loading={deleteSchemaLoading}
	onConfirm={confirmDeleteSchema}
	onCancel={() => (pendingDeleteSchema = '')}
/>

<!-- Delete Table Confirmation -->
<ConfirmModal
	bind:isOpen={showDeleteTableConfirm}
	title="Delete Table"
	message={`Are you sure you want to delete the table '${pendingDeleteTable.schema}.${pendingDeleteTable.table}'? All data in this table will be permanently lost. This action cannot be undone.`}
	confirmText="Delete Table"
	cancelText="Cancel"
	variant="danger"
	loading={deleteTableLoading}
	onConfirm={confirmDeleteTable}
	onCancel={() => (pendingDeleteTable = { schema: '', table: '' })}
/>