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
		class="fixed inset-0 bg-black/60 backdrop-blur-md z-50 flex items-center justify-center p-4"
		transition:slide={{ duration: 100 }}
	>
		<div
			class="bg-neutral-900/80 backdrop-blur-2xl border border-stone-800 rounded-none w-full max-w-md shadow-2xl flex flex-col industrial-frame"
		>
			<div
				class="p-6 border-b border-stone-800 flex justify-between items-center bg-neutral-950/40"
			>
				<h3 class="text-xl font-heading font-black text-neutral-100 uppercase tracking-tighter">
					Alter_Schema: <span class="text-rust font-jetbrains text-xs ml-3 uppercase tracking-widest">{table}</span>
				</h3>
				<button
					onclick={onClose}
					class="text-stone-600 hover:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-8 space-y-8">
				<div>
					<label
						for="column-name"
						class="block font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest mb-3"
						>Field_Identity</label
					>
					<input
						id="column-name"
						type="text"
						bind:value={columnName}
						class="w-full bg-stone-950 border border-stone-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none uppercase tracking-widest transition-all"
						placeholder="e.g. CORE_STATUS"
					/>
				</div>
				<div>
					<label
						for="column-type"
						class="block font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest mb-3"
						>Data_Type_Protocol</label
					>
					<select
						id="column-type"
						bind:value={columnType}
						class="w-full bg-stone-950 border border-stone-800 py-3 px-4 text-rust font-jetbrains text-xs focus:border-rust outline-none appearance-none cursor-pointer uppercase tracking-widest"
					>
						{#each dataTypes as type}
							<option value={type}>{type}</option>
						{/each}
					</select>
				</div>
			</div>

			<div
				class="p-6 border-t border-stone-800 bg-neutral-950/40 flex justify-end gap-4"
			>
				<button
					onclick={onClose}
					class="px-6 py-3 font-heading font-black text-[10px] text-stone-500 hover:text-white transition-all uppercase tracking-widest"
				>
					Abort_Sequence
				</button>
				<button
					onclick={handleSave}
					disabled={loading || !columnName.trim()}
					class="px-8 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 disabled:opacity-20 transition-all active:tranneutral-y-px"
				>
					{#if loading}
						<div
							class="w-4 h-4 border-2 border-white/30 border-t-white animate-spin"
						></div>
						SYNCING...
					{:else}
						<Save class="w-4 h-4" />
						Append_Field
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
