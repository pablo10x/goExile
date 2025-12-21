<script lang="ts">
	import { Plus, Trash2, X, Save } from 'lucide-svelte';
	import { slide } from 'svelte/transition';

	let {
		isOpen = $bindable(false),
		schema = 'public',
		onClose,
		onSave
	} = $props<{
		isOpen: boolean;
		schema: string;
		onClose: () => void;
		onSave: (data: { name: string; columns: string[] }) => Promise<void>;
	}>();

	let tableName = $state('');
	let columns = $state<{ name: string; type: string; constraints: string[] }[]>([
		{ name: 'id', type: 'SERIAL', constraints: ['PRIMARY KEY'] }
	]);
	let loading = $state(false);

	const dataTypes = ['SERIAL', 'INTEGER', 'TEXT', 'BOOLEAN', 'TIMESTAMP', 'JSONB', 'UUID'];
	const constraints = ['PRIMARY KEY', 'NOT NULL', 'UNIQUE', 'DEFAULT 0', 'DEFAULT NOW()'];

	function addColumn() {
		columns = [...columns, { name: '', type: 'TEXT', constraints: [] }];
	}

	function removeColumn(index: number) {
		columns = columns.filter((_, i) => i !== index);
	}

	function toggleConstraint(index: number, constraint: string) {
		const col = columns[index];
		if (col.constraints.includes(constraint)) {
			col.constraints = col.constraints.filter((c) => c !== constraint);
		} else {
			col.constraints = [...col.constraints, constraint];
		}
		columns[index] = col; // Trigger reactivity
	}

	async function handleSave() {
		if (!tableName.trim()) return;
		loading = true;

		// Build column definitions
		const colDefs = columns.map((col) => {
			const parts = [col.name, col.type, ...col.constraints];
			return parts.join(' ');
		});

		try {
			await onSave({ name: tableName, columns: colDefs });
			onClose();
			// Reset form
			tableName = '';
			columns = [{ name: 'id', type: 'SERIAL', constraints: ['PRIMARY KEY'] }];
		} finally {
			loading = false;
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
		transition:slide={{ duration: 100 }}
	>
		<div
			class="bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl w-full max-w-2xl shadow-2xl flex flex-col max-h-[90vh]"
		>
			<div
				class="p-4 border-b border-slate-300 dark:border-slate-700 flex justify-between items-center"
			>
				<h3 class="text-lg font-bold text-slate-100">
					Create Table in <span class="text-blue-400 font-mono">{schema}</span>
				</h3>
				<button
					onclick={onClose}
					class="text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-6 overflow-y-auto space-y-6">
				<div>
					<label
						for="tableName"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5"
						>Table Name</label
					>
					<input
						type="text"
						id="tableName"
						bind:value={tableName}
						class="w-full bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 outline-none focus:border-blue-500"
						placeholder="e.g. users"
					/>
				</div>

				<div class="space-y-3">
					<div class="flex justify-between items-center">
						<h4
							class="text-sm font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
						>
							Columns
						</h4>
						<button
							onclick={addColumn}
							class="text-xs flex items-center gap-1 text-emerald-400 hover:text-emerald-300 font-bold"
						>
							<Plus class="w-3 h-3" /> Add Column
						</button>
					</div>

					<div class="space-y-2">
						{#each columns as col, i}
							<div
								class="flex flex-col gap-2 p-3 bg-white/50 dark:bg-slate-950/50 border border-slate-200 dark:border-slate-800 rounded-lg group"
							>
								<div class="flex gap-2 items-start">
									<div class="flex flex-col gap-1 w-full">
										<label for="columnName-{i}" class="text-xs text-slate-500 dark:text-slate-400"
											>Name</label
										>
										<input
											type="text"
											id="columnName-{i}"
											bind:value={col.name}
											placeholder="Column Name"
											class="flex-1 bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm text-slate-800 dark:text-slate-200 outline-none focus:border-blue-500"
										/>
									</div>
									<div class="flex flex-col gap-1 w-full">
										<label for="columnType-{i}" class="text-xs text-slate-500 dark:text-slate-400"
											>Type</label
										>
										<select
											id="columnType-{i}"
											bind:value={col.type}
											class="bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm text-blue-400 outline-none focus:border-blue-500 font-mono"
										>
											{#each dataTypes as type}
												<option value={type}>{type}</option>
											{/each}
										</select>
									</div>
									<button
										onclick={() => removeColumn(i)}
										class="p-1.5 text-slate-500 hover:text-red-400 rounded"
										title="Remove Column"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								</div>

								<div class="flex flex-wrap gap-2">
									{#each constraints as c}
										<button
											onclick={() => toggleConstraint(i, c)}
											class="px-2 py-0.5 rounded text-xs border transition-colors {col.constraints.includes(
												c
											)
												? 'bg-blue-500/20 text-blue-300 border-blue-500/40'
												: 'bg-slate-900 text-slate-500 border-slate-300 dark:border-slate-700 hover:border-slate-500'}"
										>
											{c}
										</button>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<div
				class="p-4 border-t border-slate-300 dark:border-slate-700 bg-slate-900/50 flex justify-end gap-3"
			>
				<button
					onclick={onClose}
					class="px-4 py-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleSave}
					disabled={loading || !tableName.trim()}
					class="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 dark:text-white rounded-lg font-bold flex items-center gap-2 shadow-lg disabled:opacity-50"
				>
					{#if loading}
						<div
							class="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin"
						></div>
					{:else}
						<Save class="w-4 h-4" />
					{/if}
					Create Table
				</button>
			</div>
		</div>
	</div>
{/if}
