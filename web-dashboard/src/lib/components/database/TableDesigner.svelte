<script lang="ts">
	import { Plus, Trash2, Save, X, Key, Type, Settings } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { notifications } from '$lib/stores.svelte';
	import { autofocus } from '$lib/actions';

	let {
		schema,
		table,
		columns = [],
		onClose,
		onReload
	} = $props<{
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

	const dataTypes = [
		'SERIAL',
		'INTEGER',
		'BIGINT',
		'TEXT',
		'VARCHAR',
		'BOOLEAN',
		'TIMESTAMP',
		'TIMESTAMPTZ',
		'DATE',
		'JSONB',
		'UUID',
		'FLOAT',
		'DOUBLE PRECISION'
	];

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
		if (!confirm(`Changing type to ${newType} might fail if data is incompatible. Continue?`))
			return;
		await performAlter({ action: 'alter_column_type', column: colName, type: newType });
	}

	async function toggleNullable(colName: string, currentNullable: string) {
		// currentNullable is "YES" or "NO"
		const notNull = currentNullable === 'YES'; // Toggle: if currently yes (nullable), we want not null
		await performAlter({ action: 'alter_column_nullable', column: colName, not_null: notNull });
	}

	async function changeDefault(colName: string, currentDefault: string) {
		const newVal = prompt(
			"Enter new default value (SQL expression, e.g. '0' or 'now()'). Leave empty to drop default.",
			currentDefault
		);
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

<div class="h-full flex flex-col bg-black/20">
	<div
		class="p-4 border-b border-stone-800 flex justify-between items-center bg-black/40 backdrop-blur-md"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-yellow-500/10 border border-yellow-500/20 rounded-none industrial-sharp">
				<Settings class="w-5 h-5 text-yellow-400" />
			</div>
			<div>
				<h2 class="text-base font-heading font-black text-slate-100 flex items-center gap-3 tracking-widest uppercase italic">
					Logic_Schema: <span class="text-yellow-500 font-mono text-xs">{schema}.{table}</span>
				</h2>
				<p class="text-[9px] text-stone-600 font-bold uppercase tracking-widest italic">Manage sectors, types, and primary constraints</p>
			</div>
		</div>
		<button
			onclick={onClose}
			class="p-2 hover:bg-stone-800 rounded-none text-stone-600 hover:text-white transition-all"
		>
			<X class="w-5 h-5" />
		</button>
	</div>

	<div class="flex-1 overflow-auto p-8 bg-black/20">
		<div
			class="bg-slate-900/40 border border-stone-800 rounded-none overflow-hidden shadow-2xl industrial-sharp backdrop-blur-sm"
		>
			<table class="w-full text-left font-jetbrains">
				<thead
					class="bg-black text-stone-500 uppercase text-[9px] font-black tracking-[0.2em] italic"
				>
					<tr>
						<th class="px-8 py-4 border-b border-stone-800">Field_Identity</th>
						<th class="px-8 py-4 border-b border-stone-800">Type_Protocol</th>
						<th class="px-8 py-4 border-b border-stone-800 text-center"
							>Nullable_State</th
						>
						<th class="px-8 py-4 border-b border-stone-800">Default_Buffer</th>
						<th class="px-8 py-4 border-b border-stone-800 text-right"
							>Purge_Op</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-stone-800/50">
					{#each columns as col}
						<tr class="hover:bg-yellow-500/5 transition-colors group">
							<td class="px-8 py-4 font-black text-[11px] text-stone-300 tracking-widest">
								{#if editingColName === col.name}
									<input
										type="text"
										bind:value={newColNameVal}
										class="bg-black border border-yellow-500 rounded-none px-3 py-1 text-white w-full outline-none industrial-sharp"
										onkeydown={(e) => e.key === 'Enter' && renameColumn(col.name)}
										onblur={() => renameColumn(col.name)}
										use:autofocus
									/>
								{:else}
									<button
										onclick={() => startRename(col)}
										class="hover:text-yellow-400 hover:underline decoration-dashed underline-offset-4 decoration-stone-700 uppercase"
									>
										{col.name}
									</button>
									{#if col.constraints?.includes('PRIMARY KEY') || col.name === 'id'}
										<Key class="w-3.5 h-3.5 inline ml-3 text-yellow-500 animate-pulse" />
									{/if}
								{/if}
							</td>
							<td class="px-8 py-4">
								<div class="relative group/type">
									<select
										value={col.type.toUpperCase()}
										onchange={(e) => changeType(col.name, e.currentTarget.value)}
										class="bg-transparent border-none text-emerald-500 font-black text-[10px] cursor-pointer outline-none appearance-none hover:text-emerald-300 w-full tracking-widest"
									>
										{#each dataTypes as t}
											<option value={t} class="bg-stone-900 text-stone-300"
												>{t}</option
											>
										{/each}
										<!-- Keep original if not in list -->
										{#if !dataTypes.includes(col.type.toUpperCase())}
											<option value={col.type.toUpperCase()}>{col.type.toUpperCase()}</option>
										{/if}
									</select>
								</div>
							</td>
							<td class="px-8 py-4 text-center">
								<button
									onclick={() => toggleNullable(col.name, col.nullable)}
									class="px-3 py-1 rounded-none text-[9px] font-black border transition-all uppercase tracking-widest {col.nullable ===
									'YES'
										? 'border-stone-800 text-stone-600 hover:border-stone-600'
										: 'border-yellow-500/30 bg-yellow-500/10 text-yellow-400'}"
								>
									{col.nullable === 'YES' ? 'NULL' : 'NOT NULL'}
								</button>
							</td>
							<td class="px-8 py-4 font-black text-[10px] text-stone-600 tracking-widest">
								<button
									onclick={() => changeDefault(col.name, col.default)}
									class="hover:text-white truncate max-w-[150px] block transition-colors italic"
									title={col.default || 'No Default'}
								>
									{col.default || 'NULL_PTR'}
								</button>
							</td>
							<td class="px-8 py-4 text-right">
								<button
									onclick={() => dropColumn(col.name)}
									class="p-2 text-stone-700 hover:text-red-500 hover:bg-red-500/10 rounded-none transition-colors opacity-0 group-hover:opacity-100"
									title="Drop Column"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							</td>
						</tr>
					{/each}
					<tr class="bg-black/40 border-t border-stone-800 border-dashed">
						<td colspan="5" class="px-8 py-4 text-center">
							<span class="text-stone-700 text-[9px] font-black uppercase tracking-[0.3em] italic"
								>Neural_Field_Modification_Buffer_v1.0</span
							>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</div>
