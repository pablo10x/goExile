<script lang="ts">
	import { Plus, Trash2, X, Save } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import Button from './Button.svelte';

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
		class="fixed inset-0 bg-black/60 backdrop-blur-md z-50 flex items-center justify-center p-4"
		transition:slide={{ duration: 100 }}
	>
		<div
			class="bg-slate-900/80 backdrop-blur-2xl border border-stone-800 rounded-none w-full max-w-2xl shadow-2xl flex flex-col max-h-[90vh] industrial-frame"
		>
			<div
				class="p-6 border-b border-stone-800 flex justify-between items-center bg-slate-950/40"
			>
				<h3 class="text-xl font-heading font-black text-slate-100 uppercase tracking-tighter">
					Initialize_Sector in <span class="text-rust font-jetbrains text-xs ml-3 uppercase tracking-widest">{schema}</span>
				</h3>
				<button
					onclick={onClose}
					class="text-stone-600 hover:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-8 overflow-y-auto space-y-8 custom-scrollbar">
				<div>
					<label
						for="tableName"
						class="block font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest mb-3"
						>Table_ID</label
					>
					<input
						type="text"
						id="tableName"
						bind:value={tableName}
						class="w-full bg-stone-950 border border-stone-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none uppercase tracking-widest transition-all"
						placeholder="e.g. USER_REGISTRY"
					/>
				</div>

				<div class="space-y-4">
					<div class="flex justify-between items-center border-b border-stone-800/50 pb-3">
						<h4
							class="font-heading font-black text-xs text-stone-500 uppercase tracking-[0.2em]"
						>
							COLUMN_DEFINITIONS
						</h4>
						<Button
							onclick={addColumn}
							variant="ghost"
							size="xs"
							icon="ph:plus-bold"
							class="text-emerald-500 hover:text-emerald-400"
						>
							Add_Field
						</Button>
					</div>

					<div class="space-y-3">
						{#each columns as col, i}
							<div
								class="flex flex-col gap-4 p-5 bg-slate-950/40 border border-stone-800 rounded-none group industrial-frame"
							>
								<div class="flex gap-4 items-start">
									<div class="flex flex-col gap-2 w-full">
										<label for="columnName-{i}" class="font-jetbrains text-[9px] font-bold text-stone-600 uppercase tracking-widest"
											>Field_ID</label
										>
										<input
											type="text"
											id="columnName-{i}"
											bind:value={col.name}
											placeholder="Identifier"
											class="w-full bg-stone-950 border border-stone-800 py-2 px-3 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none uppercase tracking-widest"
										/>
									</div>
									<div class="flex flex-col gap-2 w-full">
										<label for="columnType-{i}" class="font-jetbrains text-[9px] font-bold text-stone-600 uppercase tracking-widest"
											>Data_Type</label
										>
										<select
											id="columnType-{i}"
											bind:value={col.type}
											class="w-full bg-stone-950 border border-stone-800 py-2 px-3 text-rust font-jetbrains text-xs focus:border-rust outline-none cursor-pointer uppercase tracking-widest"
										>
											{#each dataTypes as type}
												<option value={type}>{type}</option>
											{/each}
										</select>
									</div>
									<button
										onclick={() => removeColumn(i)}
										class="p-2 mt-6 text-stone-700 hover:text-red-500 transition-colors"
										title="Remove Column"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								</div>

								<div class="flex flex-wrap gap-2">
									{#each constraints as c}
										<button
											onclick={() => toggleConstraint(i, c)}
											class="px-2 py-1 font-jetbrains text-[9px] font-black border transition-all uppercase tracking-tighter {col.constraints.includes(
												c
											)
												? 'bg-rust/10 text-rust border-rust/40'
												: 'bg-stone-950 text-stone-600 border-stone-800 hover:border-stone-600'}"
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
				class="p-6 border-t border-stone-800 bg-slate-950/40 flex justify-end gap-4"
			>
				<Button
					onclick={onClose}
					variant="ghost"
					size="md"
				>
					Abort_OP
				</Button>
				<Button
					onclick={handleSave}
					disabled={loading || !tableName.trim()}
					variant="primary"
					size="md"
					loading={loading}
					icon="save"
				>
					Authorize_Init
				</Button>
			</div>
		</div>
	</div>
{/if}
