<script lang="ts">
	import {
		Save,
		X,
		Plus,
		Database,
		Sparkles,
		AlertCircle,
		Check,
		Type,
		Hash,
		Calendar,
		ToggleLeft,
		FileJson,
		Loader2
	} from 'lucide-svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut, backOut, elasticOut } from 'svelte/easing';
	import { untrack } from 'svelte';
	import Button from '../Button.svelte';

	let {
		isOpen = $bindable(false),
		schema,
		table,
		columns = [],
		rowData = null,
		onClose,
		onSave
	} = $props<{
		isOpen: boolean;
		schema: string;
		table: string;
		columns: any[];
		rowData?: any;
		onClose: () => void;
		onSave: (data: any) => Promise<void>;
	}>();

	let formData = $state<Record<string, any>>({});
	let loading = $state(false);
	let lastOpenState = $state(false);
	let error = $state<string | null>(null);
	let successFields = $state<Set<string>>(new Set());
	let focusedField = $state<string | null>(null);
	let modifiedFields = $state<Set<string>>(new Set());

	// Derived values
	let editableColumns = $derived(columns.filter((c: any) => c.name !== 'id'));
	let modifiedCount = $derived(modifiedFields.size);
	let isEditing = $derived(rowData !== null);

	// Initialize form data when opening
	$effect(() => {
		const currentIsOpen = isOpen;

		if (currentIsOpen && !lastOpenState) {
			untrack(() => {
				const newFormData: Record<string, any> = {};
				if (rowData) {
					Object.assign(newFormData, rowData);
				} else {
					for (const col of columns) {
						if (col.name !== 'id') {
							newFormData[col.name] = null;
						}
					}
				}
				formData = newFormData;
				modifiedFields = new Set();
				error = null;
				successFields = new Set();
			});
		}

		lastOpenState = currentIsOpen;
	});

	function getFieldIcon(type: string) {
		if (!type) return Type;
		const t = type.toLowerCase();
		if (
			t.includes('int') ||
			t.includes('float') ||
			t.includes('numeric') ||
			t.includes('decimal') ||
			t === 'real'
		)
			return Hash;
		if (t.includes('bool')) return ToggleLeft;
		if (t.includes('json')) return FileJson;
		if (t.includes('timestamp') || t.includes('date') || t.includes('time')) return Calendar;
		return Type;
	}

	function getFieldColor(type: string) {
		if (!type)
			return {
				bg: 'bg-slate-500/20',
				text: 'text-text-dim dark:text-text-dim',
				border: 'border-slate-500/30'
			};
		const t = type.toLowerCase();
		if (
			t.includes('int') ||
			t.includes('float') ||
			t.includes('numeric') ||
			t.includes('decimal') ||
			t === 'real'
		) {
			return { bg: 'bg-rust/20', text: 'text-rust-light', border: 'border-rust/30' };
		}
		if (t.includes('bool')) {
			return { bg: 'bg-stone-800', text: 'text-stone-400', border: 'border-stone-700' };
		}
		if (t.includes('json')) {
			return { bg: 'bg-amber-500/20', text: 'text-warning', border: 'border-amber-500/30' };
		}
		if (t.includes('timestamp') || t.includes('date') || t.includes('time')) {
			return { bg: 'bg-success/20', text: 'text-success', border: 'border-emerald-500/30' };
		}
		if (t.includes('text') || t.includes('char') || t === 'uuid') {
			return { bg: 'bg-rust/10', text: 'text-rust-light', border: 'border-rust/20' };
		}
		return {
			bg: 'bg-slate-500/20',
			text: 'text-text-dim dark:text-text-dim',
			border: 'border-slate-500/30'
		};
	}

	function handleFieldChange(colName: string) {
		modifiedFields.add(colName);
		modifiedFields = new Set(modifiedFields);
	}

	async function handleSubmit() {
		loading = true;
		error = null;

		const payload: Record<string, any> = {};
		for (const col of columns) {
			if (col.name === 'id' && !rowData) continue;

			let val = formData[col.name];

			if (
				col.type &&
				(col.type.startsWith('int') || col.type === 'numeric' || col.type === 'real')
			) {
				if (val !== null && val !== '' && val !== undefined) val = Number(val);
			}

			if (val !== undefined && val !== '') {
				payload[col.name] = val;
			}
		}

		try {
			await onSave(payload);
			for (const key of Object.keys(payload)) {
				successFields.add(key);
			}
			successFields = new Set(successFields);

			setTimeout(() => {
				onClose();
			}, 300);
		} catch (e: any) {
			error = e.message || 'An unexpected error occurred.';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && !loading) {
			onClose();
		}
		if (e.key === 'Enter' && e.ctrlKey) {
			handleSubmit();
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget && !loading) {
			onClose();
		}
	}

	function getContainerClass(isSuccess: boolean, isFocused: boolean, isModified: boolean): string {
		if (isSuccess) return 'border-emerald-500/50 bg-success/5';
		if (isFocused) return 'border-rust/50 bg-rust/5';
		if (isModified) return 'border-amber-500/30 bg-amber-500/5';
		return 'border-slate-300/50 dark:border-slate-700/50 bg-slate-950/40 hover:border-slate-600/50';
	}

	function getBoolButtonClass(formValue: any, optionValue: any, optionColor: string): string {
		if (formValue === optionValue) {
			if (optionColor === 'emerald')
				return 'bg-success/20 text-success border border-emerald-500/50 shadow-lg shadow-emerald-500/10';
			if (optionColor === 'red')
				return 'bg-red-500/20 text-danger border border-red-500/50 shadow-lg shadow-red-500/10';
			return 'bg-slate-600/30 text-slate-700 dark:text-slate-300 border border-slate-500/50';
		}
		return 'bg-slate-800/50 text-text-dim border border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600/50 hover:text-text-dim dark:text-text-dim';
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center p-4 sm:p-6"
		transition:fade={{ duration: 200 }}
		onclick={handleBackdropClick}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') handleBackdropClick(e as unknown as MouseEvent);
		}}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		tabindex="-1"
	>
		<div
			class="absolute inset-0 bg-black/60 backdrop-blur-md"
			transition:fade={{ duration: 300 }}
		></div>

		<div class="absolute inset-0 overflow-hidden pointer-events-none">
			{#each Array(15) as _, i}
				<div
					class="absolute w-1 h-1 bg-rust/30 rounded-full animate-float"
					style="left: {Math.random() * 100}%; top: {Math.random() *
						100}%; animation-delay: {Math.random() * 5}s; animation-duration: {3 +
						Math.random() * 4}s;"
				></div>
			{/each}
		</div>

		<div
			class="relative w-full max-w-2xl max-h-[90vh] flex flex-col"
			transition:fly={{ y: 30, duration: 400, easing: backOut }}
		>
			<div
				class="absolute -inset-1 bg-gradient-to-r from-rust/20 via-stone-600/20 to-orange-600/20 rounded-2xl blur-xl opacity-75"
				transition:scale={{ start: 0.8, duration: 500, easing: elasticOut }}
			></div>

			<div
				class="relative bg-slate-900/80 backdrop-blur-2xl border border-stone-800 rounded-2xl shadow-2xl overflow-hidden"
			>
				<div class="relative overflow-hidden">
					<div
						class="absolute inset-0 bg-gradient-to-r from-rust/10 via-stone-600/10 to-orange-600/10"
					></div>
					<div
						class="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent -translate-x-full animate-shimmer"
					></div>

					<div class="relative p-6 flex items-start justify-between gap-4">
						<div class="flex items-start gap-4">
							<div
								class="relative p-3 rounded-xl bg-gradient-to-br from-stone-800 to-stone-900 border border-stone-700"
								transition:scale={{ start: 0.5, duration: 400, delay: 100, easing: backOut }}
							>
								{#if isEditing}
									<Database class="w-6 h-6 text-rust-light" />
								{:else}
									<Plus class="w-6 h-6 text-rust" />
								{/if}
								<div
									class="absolute inset-0 rounded-xl border-2 border-rust animate-ping opacity-20"
								></div>
							</div>

							<div>
								<h2
									id="modal-title"
									class="text-xl font-heading font-black text-white flex items-center gap-2 tracking-widest uppercase"
								>
									{#if isEditing}
										Edit Row
									{:else}
										<span class="flex items-center gap-2">
											Add New Row
											<Sparkles class="w-4 h-4 text-rust animate-pulse" />
										</span>
									{/if}
								</h2>
								<p class="text-sm text-text-dim dark:text-text-dim mt-1 flex items-center gap-2">
									<span class="px-2 py-0.5 bg-slate-800 rounded-md font-mono text-xs"
										>{schema}.{table}</span
									>
									<span class="text-slate-600">•</span>
									<span>{editableColumns.length} fields</span>
									{#if modifiedCount > 0}
										<span
											class="px-2 py-0.5 bg-rust/20 text-rust-light rounded-full text-[9px] font-black uppercase tracking-widest"
											transition:scale={{ start: 0.8, duration: 200 }}
										>
											{modifiedCount} modified
										</span>
									{/if}
								</p>
							</div>
						</div>

						<button
							onclick={onClose}
							disabled={loading}
							class="p-2 text-text-dim dark:text-text-dim hover:text-slate-900 dark:text-white hover:bg-slate-800/80 rounded-xl transition-all duration-200 hover:scale-105 active:scale-95 disabled:opacity-50"
							aria-label="Close modal"
						>
							<X class="w-5 h-5" />
						</button>
					</div>
				</div>

				{#if error}
					<div
						class="mx-6 mb-4 p-4 bg-red-500/10 border border-red-500/30 rounded-xl flex items-start gap-3"
						transition:fly={{ y: -10, duration: 300 }}
					>
						<AlertCircle class="w-5 h-5 text-danger flex-shrink-0 mt-0.5" />
						<div class="flex-1">
							<p class="text-danger font-medium text-sm">Error saving row</p>
							<p class="text-red-300/80 text-sm mt-1">{error}</p>
						</div>
						<button
							onclick={() => (error = null)}
							class="text-danger/60 hover:text-danger transition-colors"
						>
							<X class="w-4 h-4" />
						</button>
					</div>
				{/if}

				<div class="p-6 pt-2 overflow-y-auto max-h-[calc(90vh-220px)] custom-scrollbar">
					{#if editableColumns.length === 0}
						<div class="text-center py-12">
							<div
								class="w-16 h-16 mx-auto mb-4 rounded-full bg-slate-800/50 flex items-center justify-center"
							>
								<Database class="w-8 h-8 text-slate-600" />
							</div>
							<p class="text-text-dim dark:text-text-dim font-medium">No editable columns</p>
							<p class="text-text-dim text-sm mt-1">
								This table has no columns that can be edited
							</p>
						</div>
					{:else}
						<div class="grid gap-4">
							{#each editableColumns as col, index (col.name)}
								{@const colors = getFieldColor(col.type)}
								{@const FieldIcon = getFieldIcon(col.type)}
								{@const isModified = modifiedFields.has(col.name)}
								{@const isSuccess = successFields.has(col.name)}
								{@const isFocused = focusedField === col.name}

								<div
									class="group relative"
									transition:fly={{ y: 20, duration: 300, delay: index * 50, easing: cubicOut }}
								>
									<div
										class="relative rounded-xl border transition-all duration-300 {getContainerClass(
											isSuccess,
											isFocused,
											isModified
										)}"
									>
										<div class="flex items-center justify-between px-4 pt-3 pb-1">
											<label
												for="field-{col.name}"
												class="flex items-center gap-2 text-sm font-medium transition-colors {isFocused
													? 'text-info'
													: 'text-slate-700 dark:text-slate-300'}"
											>
												<div class="p-1 rounded-md {colors.bg}">
													<FieldIcon class="w-3.5 h-3.5 {colors.text}" />
												</div>
												<span>{col.name}</span>
											</label>

											<div class="flex items-center gap-2">
												<span
													class="px-2 py-0.5 text-xs rounded-md {colors.bg} {colors.text} font-mono"
												>
													{col.type || 'unknown'}
												</span>

												{#if col.nullable === 'YES'}
													<span
														class="px-1.5 py-0.5 text-xs rounded-md bg-slate-700/50 text-text-dim"
														>nullable</span
													>
												{/if}

												{#if isSuccess}
													<div transition:scale={{ start: 0.5, duration: 300 }}>
														<Check class="w-4 h-4 text-success" />
													</div>
												{:else if isModified}
													<div
														class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"
														transition:scale={{ start: 0.5, duration: 200 }}
													></div>
												{/if}
											</div>
										</div>

										<div class="px-4 pb-3">
											{#if col.type === 'boolean'}
												<div class="flex items-center gap-3 py-2">
													{#each [{ value: null, label: 'NULL', color: 'slate' }, { value: true, label: 'TRUE', color: 'emerald' }, { value: false, label: 'FALSE', color: 'red' }] as option}
														<button
															type="button"
															onclick={() => {
																formData[col.name] = option.value;
																handleFieldChange(col.name);
															}}
															class="flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-200 {getBoolButtonClass(
																formData[col.name],
																option.value,
																option.color
															)}"
														>
															{option.label}
														</button>
													{/each}
												</div>
											{:else if col.type && col.type.includes('json')}
												<textarea
													id="field-{col.name}"
													bind:value={formData[col.name]}
													oninput={() => handleFieldChange(col.name)}
													onfocus={() => (focusedField = col.name)}
													onblur={() => (focusedField = null)}
													class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300/50 dark:border-slate-700/50 rounded-lg px-4 py-3 text-slate-800 dark:text-slate-200 font-mono text-sm outline-none transition-all duration-200 focus:border-rust/50 focus:ring-2 focus:ring-rust/20 resize-none"
													rows="4"
													placeholder={'{ "key": "value" }'}
													spellcheck="false"
												></textarea>
											{:else if col.type && (col.type.includes('timestamp') || col.type.includes('date'))}
												<input
													id="field-{col.name}"
													type="datetime-local"
													bind:value={formData[col.name]}
													oninput={() => handleFieldChange(col.name)}
													onfocus={() => (focusedField = col.name)}
													onblur={() => (focusedField = null)}
													class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300/50 dark:border-slate-700/50 rounded-lg px-4 py-3 text-slate-800 dark:text-slate-200 outline-none transition-all duration-200 focus:border-rust/50 focus:ring-2 focus:ring-rust/20"
												/>
											{:else if col.type && (col.type.includes('int') || col.type.includes('float') || col.type.includes('numeric') || col.type.includes('decimal') || col.type === 'real')}
												<input
													id="field-{col.name}"
													type="number"
													bind:value={formData[col.name]}
													oninput={() => handleFieldChange(col.name)}
													onfocus={() => (focusedField = col.name)}
													onblur={() => (focusedField = null)}
													class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300/50 dark:border-slate-700/50 rounded-lg px-4 py-3 text-slate-800 dark:text-slate-200 outline-none transition-all duration-200 focus:border-rust/50 focus:ring-2 focus:ring-rust/20"
													placeholder="Enter a number..."
													step="any"
												/>
											{:else if col.type && (col.type.includes('text') || col.type === 'character varying') && !col.type.includes('[]')}
												<textarea
													id="field-{col.name}"
													bind:value={formData[col.name]}
													oninput={() => handleFieldChange(col.name)}
													onfocus={() => (focusedField = col.name)}
													onblur={() => (focusedField = null)}
													class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300/50 dark:border-slate-700/50 rounded-lg px-4 py-3 text-slate-800 dark:text-slate-200 outline-none transition-all duration-200 focus:border-rust/50 focus:ring-2 focus:ring-rust/20 resize-none"
													rows="2"
													placeholder="Enter text..."
												></textarea>
											{:else}
												<input
													id="field-{col.name}"
													type="text"
													bind:value={formData[col.name]}
													oninput={() => handleFieldChange(col.name)}
													onfocus={() => (focusedField = col.name)}
													onblur={() => (focusedField = null)}
													class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300/50 dark:border-slate-700/50 rounded-lg px-4 py-3 text-slate-800 dark:text-slate-200 outline-none transition-all duration-200 focus:border-rust/50 focus:ring-2 focus:ring-rust/20"
													placeholder="Enter value..."
												/>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<div
					class="relative border-t border-slate-300/50 dark:border-slate-700/50 bg-slate-900/80 backdrop-blur-sm p-4"
				>
					<div
						class="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-slate-600/50 to-transparent"
					></div>

					<div class="flex items-center justify-between gap-4">
						<div class="hidden sm:flex items-center gap-2 text-xs text-text-dim">
							<kbd
								class="px-2 py-1 bg-slate-800 rounded border border-slate-300 dark:border-slate-700 font-mono"
								>Ctrl</kbd
							>
							<span>+</span>
							<kbd
								class="px-2 py-1 bg-slate-800 rounded border border-slate-300 dark:border-slate-700 font-mono"
								>Enter</kbd
							>
							<span>to save</span>
						</div>

						<div class="flex items-center gap-3 ml-auto">
							<Button
								onclick={onClose}
								disabled={loading}
								variant="ghost"
								size="md"
							>
								Cancel
							</Button>

							<Button
								onclick={handleSubmit}
								disabled={loading || editableColumns.length === 0}
								variant="primary"
								size="md"
								loading={loading}
								icon={isEditing ? 'save' : 'plus'}
							>
								{isEditing ? 'Update Row' : 'Add Row'}
							</Button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes float {
		0%,
		100% {
			transform: translateY(0) translateX(0);
			opacity: 0.3;
		}
		50% {
			transform: translateY(-20px) translateX(10px);
			opacity: 0.6;
		}
	}

	@keyframes shimmer {
		100% {
			transform: translateX(100%);
		}
	}

	.animate-float {
		animation: float 3s ease-in-out infinite;
	}

	.animate-shimmer {
		animation: shimmer 2s ease-in-out infinite;
	}

	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #334155;
		border-radius: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569;
	}
</style>
