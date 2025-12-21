<script lang="ts">
	import { Plus, Trash2, X, Save } from 'lucide-svelte';
	import { slide } from 'svelte/transition';

	let {
		isOpen = $bindable(false),
		schema,
		table,
		onClose,
		onSave
	} = $props<{
		isOpen: boolean;
		schema: string;
		table: string;
		onClose: () => void;
		onSave: (data: { column: string; type: string }) => Promise<void>;
	}>();

	let columnName = $state('');
	let columnType = $state('TEXT');
	let loading = $state(false);

	const dataTypes = ['SERIAL', 'INTEGER', 'TEXT', 'BOOLEAN', 'TIMESTAMP', 'JSONB', 'UUID'];

	async function handleSave() {
		if (!columnName.trim()) return;
		loading = true;
		try {
			await onSave({ column: columnName, type: columnType });
			onClose();
			columnName = '';
			columnType = 'TEXT';
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
			class="bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl w-full max-w-md shadow-2xl flex flex-col"
		>
			<div
				class="p-4 border-b border-slate-300 dark:border-slate-700 flex justify-between items-center"
			>
				<h3 class="text-lg font-bold text-slate-100">
					Add Column to <span class="text-blue-400 font-mono">{table}</span>
				</h3>
				<button
					onclick={onClose}
					class="text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-6 space-y-4">
				<div>
					<label
						for="column-name"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5"
						>Column Name</label
					>
					<input
						id="column-name"
						type="text"
						bind:value={columnName}
						class="w-full bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 outline-none focus:border-blue-500"
						placeholder="e.g. status"
					/>
				</div>
				<div>
					<label
						for="column-type"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5"
						>Data Type</label
					>
					<select
						id="column-type"
						bind:value={columnType}
						class="w-full bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 outline-none focus:border-blue-500"
					>
						{#each dataTypes as type}
							<option value={type}>{type}</option>
						{/each}
					</select>
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
					disabled={loading || !columnName.trim()}
					class="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 dark:text-white rounded-lg font-bold flex items-center gap-2 shadow-lg disabled:opacity-50"
				>
					{#if loading}
						<div
							class="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin"
						></div>
					{:else}
						<Save class="w-4 h-4" />
					{/if}
					Add Column
				</button>
			</div>
		</div>
	</div>
{/if}
