<script lang="ts">
	import SchemaBrowser from '$lib/components/SchemaBrowser.svelte';
	import TableCreatorModal from '$lib/components/TableCreatorModal.svelte';
	import ColumnManagerModal from '$lib/components/ColumnManagerModal.svelte';
    import SchemaCreatorModal from '$lib/components/database/SchemaCreatorModal.svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores';

	let { 
		onSelectTable, 
	} = $props<{
		onSelectTable: (schema: string, table: string) => void;
	}>();

	// Internal State for Browser
	let schemas = $state<string[]>([]);
	let schemaTables = $state<Record<string, string[]>>({});
	let tableColumns = $state<Record<string, any[]>>({});
	let expandedSchemas = $state<Set<string>>(new Set());
	let expandedTables = $state<Set<string>>(new Set());
	let isSidebarOpen = $state(true);

    // Modal States
    let isTableCreatorOpen = $state(false);
    let creatorSchema = $state('public');
    
    let isSchemaCreatorOpen = $state(false);

    let isColumnManagerOpen = $state(false);
    let columnManagerTarget = $state({ schema: '', table: '' });

    async function loadDatabaseMap() {
		try {
            // Fetch all tables at once
			const res = await fetch('/api/database/all-tables');
			if (res.ok) {
				const allTables: {schema: string, table: string}[] = await res.json();
                
                // Process into map
                const newSchemaTables: Record<string, string[]> = {};
                const newSchemasSet = new Set<string>();

                allTables.forEach(row => {
                    newSchemasSet.add(row.schema);
                    if (!newSchemaTables[row.schema]) newSchemaTables[row.schema] = [];
                    newSchemaTables[row.schema].push(row.table);
                });

                // Fetch schemas list to get empty schemas too (optional, but good for completeness)
                const schemaRes = await fetch('/api/database/schemas');
                if (schemaRes.ok) {
                    const schemaList: string[] = await schemaRes.json();
                    schemaList.forEach(s => newSchemasSet.add(s));
                }

                schemas = Array.from(newSchemasSet).sort();
                schemaTables = newSchemaTables;
			}
		} catch (e) { console.error(e); }
	}

	async function loadSchemas() {
        // Redundant with loadDatabaseMap but keeping for initial check if needed, 
        // strictly speaking loadDatabaseMap does it all.
        // Let's rely on loadDatabaseMap.
        loadDatabaseMap();
	}

    // loadTables removed to prevent overwriting with potentially stale data from the per-schema endpoint

	function toggleSchema(schema: string) {
		if (expandedSchemas.has(schema)) expandedSchemas.delete(schema);
		else {
			expandedSchemas.add(schema);
            // No fetch needed, data is already loaded via global map
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
				notifications.add({ type: 'success', message: `Schema '${data.name}' created.` });
				loadDatabaseMap();
			} else {
                const err = await res.json();
                throw new Error(err.error);
            }
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to create schema', details: e.message });
		}
	}

    async function deleteSchema(name: string) {
		if (!confirm(`Delete schema '${name}'?`)) return;
		try {
			const res = await fetch(`/api/database/schemas/${name}`, { method: 'DELETE' });
			if (res.ok) {
				notifications.add({ type: 'success', message: `Schema '${name}' deleted.` });
				delete schemaTables[name];
				loadDatabaseMap();
			} else {
                const err = await res.json();
                throw new Error(err.error);
            }
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to delete schema', details: e.message });
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
            notifications.add({ type: 'success', message: `Table '${data.name}' created.` });
            loadDatabaseMap();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to create table', details: e.message });
            throw e; 
        }
    }

    async function deleteTable(schema: string, table: string) {
        if (!confirm(`Delete table '${schema}.${table}'?`)) return;
        try {
            const res = await fetch(`/api/database/tables/${table}?schema=${encodeURIComponent(schema)}`, { method: 'DELETE' });
            if (res.ok) {
                notifications.add({ type: 'success', message: `Table '${table}' deleted.` });
                loadDatabaseMap();
            } else {
                const err = await res.json();
                throw new Error(err.error);
            }
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Failed to delete table', details: e.message });
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
                notifications.add({ type: 'success', message: `Column added.` });
            } catch (e: any) {
                notifications.add({ type: 'error', message: 'Failed to add column', details: e.message });
                throw e;
            }
        }
    
        function refresh() {
            loadDatabaseMap(); // Use the global loader
        }
    
    	onMount(() => {
    		loadDatabaseMap(); // Initial load
    	});
    </script>
    
    <div class="h-full bg-slate-900 border-r border-slate-800 flex flex-col">
        <div class="p-2 border-b border-slate-800 flex justify-end">
            <button onclick={refresh} class="p-1 text-slate-500 hover:text-white rounded" title="Refresh Browser">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path><path d="M3 3v5h5"></path><path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"></path><path d="M16 16h5v5"></path></svg>
            </button>
        </div>
    	<SchemaBrowser
    		{schemas}		{schemaTables}
		{tableColumns}
		selectedSchema={null}
		selectedTable={null}
		{expandedSchemas}
		{expandedTables}
		{isSidebarOpen}
		onToggleSchema={toggleSchema}
		onToggleTable={toggleTable}
		onSelectTable={onSelectTable}
		onCreateSchema={() => isSchemaCreatorOpen = true}
		onDropSchema={deleteSchema}
		onCreateTable={openTableCreator}
		onDropTable={deleteTable}
        onEditTable={openColumnManager}
		onToggleSidebar={() => isSidebarOpen = !isSidebarOpen}
	/>
</div>

<SchemaCreatorModal
    bind:isOpen={isSchemaCreatorOpen}
    onClose={() => isSchemaCreatorOpen = false}
    onSave={handleSchemaCreate}
/>

<TableCreatorModal 
    bind:isOpen={isTableCreatorOpen} 
    schema={creatorSchema} 
    onClose={() => isTableCreatorOpen = false} 
    onSave={handleTableCreate} 
/>

<ColumnManagerModal
    bind:isOpen={isColumnManagerOpen}
    schema={columnManagerTarget.schema}
    table={columnManagerTarget.table}
    onClose={() => isColumnManagerOpen = false}
    onSave={handleAddColumn}
/>
