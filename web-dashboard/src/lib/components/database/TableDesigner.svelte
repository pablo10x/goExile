<script lang="ts">
	import { Plus, Trash2, Save, X, Key, Type, Settings } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
    import { notifications } from '$lib/stores';

	let { schema, table, columns = [], onClose, onReload } = $props<{
		schema: string;
		table: string;
		columns: any[];
		onClose: () => void;
		onReload: () => void;
	}>();

	let loading = $state(false);
	
	// Column Modification State
	let editingColName = $state<string | null>(null);
	let newColNameVal = $state('');

	const dataTypes = ['SERIAL', 'INTEGER', 'BIGINT', 'TEXT', 'VARCHAR', 'BOOLEAN', 'TIMESTAMP', 'TIMESTAMPTZ', 'DATE', 'JSONB', 'UUID', 'FLOAT', 'DOUBLE PRECISION'];

	async function dropColumn(colName: string) {
		if (!confirm(`Are you sure you want to drop column '${colName}'? Data will be lost.`)) return;
		await performAlter({ action: 'drop_column', column: colName });
	}

	async function renameColumn(oldName: string) {
		if (!newColNameVal.trim() || newColNameVal === oldName) {
            editingColName = null;
            return;
        }
		await performAlter({ action: 'rename_column', column: oldName, new_name: newColNameVal });
        editingColName = null;
	}

	async function changeType(colName: string, newType: string) {
		if (!confirm(`Changing type to ${newType} might fail if data is incompatible. Continue?`)) return;
		await performAlter({ action: 'alter_column_type', column: colName, type: newType });
	}

    async function toggleNullable(colName: string, currentNullable: string) {
        // currentNullable is "YES" or "NO"
        const notNull = currentNullable === "YES"; // Toggle: if currently yes (nullable), we want not null
        await performAlter({ action: 'alter_column_nullable', column: colName, not_null: notNull });
    }

    async function changeDefault(colName: string, currentDefault: string) {
        const newVal = prompt("Enter new default value (SQL expression, e.g. '0' or 'now()'). Leave empty to drop default.", currentDefault);
        if (newVal === null) return;
        await performAlter({ action: 'alter_column_default', column: colName, default: newVal });
    }

	async function performAlter(payload: any) {
		loading = true;
		try {
			const res = await fetch(`/api/database/tables/${table}/alter`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ schema, ...payload })
			});
			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Operation failed');
			}
			notifications.add({ type: 'success', message: 'Table modified successfully' });
			onReload();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Modification failed', details: e.message });
		} finally {
			loading = false;
		}
	}

    function startRename(col: any) {
        editingColName = col.name;
        newColNameVal = col.name;
    }
</script>

<div class="h-full flex flex-col bg-slate-900">
	<div class="p-4 border-b border-slate-700 flex justify-between items-center bg-slate-950">
		<div class="flex items-center gap-3">
            <div class="p-2 bg-blue-500/10 rounded-lg">
                <Settings class="w-5 h-5 text-blue-400" />
            </div>
			<div>
                <h2 class="text-lg font-bold text-slate-100 flex items-center gap-2">
                    Table Structure: <span class="font-mono text-blue-400">{schema}.{table}</span>
                </h2>
                <p class="text-xs text-slate-500">Manage columns, types, and constraints</p>
            </div>
		</div>
		<button onclick={onClose} class="p-2 hover:bg-slate-800 rounded-lg text-slate-400 hover:text-white transition-colors">
			<X class="w-5 h-5" />
		</button>
	</div>

	<div class="flex-1 overflow-auto p-6">
		<div class="bg-slate-950/50 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
			<table class="w-full text-left text-sm">
				<thead class="bg-slate-900 text-slate-400 uppercase text-xs font-bold tracking-wider">
					<tr>
						<th class="px-6 py-4 border-b border-slate-800">Name</th>
						<th class="px-6 py-4 border-b border-slate-800">Type</th>
						<th class="px-6 py-4 border-b border-slate-800 text-center">Nullable</th>
						<th class="px-6 py-4 border-b border-slate-800">Default</th>
						<th class="px-6 py-4 border-b border-slate-800 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-800/50">
					{#each columns as col}
						<tr class="hover:bg-slate-800/30 transition-colors group">
							<td class="px-6 py-3 font-mono text-slate-200">
                                {#if editingColName === col.name}
                                    <input 
                                        type="text" 
                                        bind:value={newColNameVal} 
                                        class="bg-slate-900 border border-blue-500 rounded px-2 py-1 text-white w-full outline-none"
                                        onkeydown={(e) => e.key === 'Enter' && renameColumn(col.name)}
                                        onblur={() => renameColumn(col.name)}
                                        autofocus
                                    />
                                {:else}
                                    <button onclick={() => startRename(col)} class="hover:text-blue-400 hover:underline decoration-dashed underline-offset-4 decoration-slate-600">
                                        {col.name}
                                    </button>
                                    {#if col.constraints?.includes('PRIMARY KEY') || col.name === 'id'} 
                                        <Key class="w-3 h-3 inline ml-2 text-amber-400" />
                                    {/if}
                                {/if}
                            </td>
							<td class="px-6 py-3">
                                <div class="relative group/type">
                                    <select 
                                        value={col.type.toUpperCase()} 
                                        onchange={(e) => changeType(col.name, e.currentTarget.value)}
                                        class="bg-transparent border-none text-emerald-400 font-mono cursor-pointer outline-none appearance-none hover:text-emerald-300 w-full"
                                    >
                                        {#each dataTypes as t}
                                            <option value={t} class="bg-slate-900 text-slate-300">{t}</option>
                                        {/each}
                                        <!-- Keep original if not in list -->
                                        {#if !dataTypes.includes(col.type.toUpperCase())}
                                            <option value={col.type.toUpperCase()}>{col.type.toUpperCase()}</option>
                                        {/if}
                                    </select>
                                </div>
                            </td>
							<td class="px-6 py-3 text-center">
                                <button 
                                    onclick={() => toggleNullable(col.name, col.nullable)}
                                    class="px-2 py-1 rounded text-xs font-bold border transition-all {col.nullable === 'YES' ? 'border-slate-700 text-slate-500 hover:border-slate-500' : 'border-purple-500/30 bg-purple-500/10 text-purple-400'}"
                                >
                                    {col.nullable === 'YES' ? 'NULL' : 'NOT NULL'}
                                </button>
                            </td>
							<td class="px-6 py-3 font-mono text-xs text-slate-400">
                                <button onclick={() => changeDefault(col.name, col.default)} class="hover:text-white truncate max-w-[150px] block" title={col.default || 'No Default'}>
                                    {col.default || '-'}
                                </button>
                            </td>
							<td class="px-6 py-3 text-right">
								<button
									onclick={() => dropColumn(col.name)}
									class="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                                    title="Drop Column"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							</td>
						</tr>
					{/each}
                    <tr class="bg-slate-900/30 border-t border-slate-800 border-dashed">
                        <td colspan="5" class="px-6 py-3 text-center">
                            <!-- Helper text or quick add -->
                            <span class="text-slate-600 text-xs">To add a column, use the 'Add Column' button (Coming soon to inline)</span>
                        </td>
                    </tr>
				</tbody>
			</table>
		</div>
	</div>
</div>
