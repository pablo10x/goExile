<script lang="ts">
	import { Save, X } from 'lucide-svelte';
	import { slide } from 'svelte/transition';

	let { 
        isOpen = $bindable(false), 
        schema, 
        table, 
        columns = [], 
        rowData = null, // If null, it's 'Add Mode'. If set, 'Edit Mode' (future use)
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

    // Initialize form data when opening
    $effect(() => {
        if (isOpen) {
            formData = {};
            if (rowData) {
                formData = { ...rowData };
            } else {
                // Set defaults?
                columns.forEach((col: any) => {
                    if (col.name !== 'id') { // Skip auto-increment usually
                        formData[col.name] = null;
                    }
                });
            }
        }
    });

	async function handleSubmit() {
		loading = true;
        // Filter out nulls/undefined for insertion if we want DB defaults to trigger?
        // Or send explicitly null.
        // For simplicity, we send keys that have values or explicit user input.
        
        const payload: Record<string, any> = {};
        for (const col of columns) {
            if (col.name === 'id' && !rowData) continue; // Skip ID on insert
            
            let val = formData[col.name];
            
            // Basic type coercion
            if (col.type.startsWith('int') || col.type === 'numeric' || col.type === 'real') {
                if (val !== null && val !== '') val = Number(val);
            } else if (col.type === 'boolean') {
                // boolean usually handled by checkbox binding
            }
            
            if (val !== undefined && val !== '') {
                payload[col.name] = val;
            }
        }

		try {
			await onSave(payload);
			onClose();
		} finally {
			loading = false;
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 bg-black/50 backdrop-blur-sm z-[60] flex items-center justify-center p-4"
		transition:slide={{ duration: 100 }}
	>
		<div class="bg-slate-900 border border-slate-700 rounded-xl w-full max-w-lg shadow-2xl flex flex-col max-h-[90vh]">
			<div class="p-4 border-b border-slate-700 flex justify-between items-center bg-slate-950">
				<h3 class="text-lg font-bold text-slate-100">
                    {rowData ? 'Edit Row' : 'Add Row'} 
                    <span class="text-slate-500 text-sm font-normal">in {schema}.{table}</span>
                </h3>
				<button onclick={onClose} class="text-slate-400 hover:text-white">
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-6 overflow-y-auto space-y-4">
                {#each columns as col}
                    {#if col.name !== 'id'} <!-- Skip ID for now, assume serial/uuid -->
                        <div>
                            <label class="block text-xs font-bold text-slate-400 uppercase mb-1">
                                {col.name} 
                                <span class="text-slate-600 font-normal normal-case">({col.type})</span>
                            </label>
                            
                            {#if col.type === 'boolean'}
                                <select 
                                    bind:value={formData[col.name]}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 outline-none focus:border-blue-500"
                                >
                                    <option value={null}>NULL</option>
                                    <option value={true}>TRUE</option>
                                    <option value={false}>FALSE</option>
                                </select>
                            {:else if col.type.includes('text') || col.type.includes('char') || col.type === 'uuid'}
                                <input
                                    type="text"
                                    bind:value={formData[col.name]}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 outline-none focus:border-blue-500"
                                    placeholder="NULL"
                                />
                            {:else if col.type.includes('int') || col.type.includes('float') || col.type.includes('numeric')}
                                <input
                                    type="number"
                                    bind:value={formData[col.name]}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 outline-none focus:border-blue-500"
                                    placeholder="NULL"
                                />
                            {:else if col.type.includes('json')}
                                <textarea
                                    bind:value={formData[col.name]}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 outline-none focus:border-blue-500 font-mono text-xs"
                                    rows="3"
                                    placeholder="{`{}`}"
                                ></textarea>
                            {:else}
                                <!-- Fallback -->
                                <input
                                    type="text"
                                    bind:value={formData[col.name]}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 outline-none focus:border-blue-500"
                                />
                            /
                            {/if}
                        </div>
                    {/if}
                {/each}
			</div>

			<div class="p-4 border-t border-slate-700 bg-slate-950 flex justify-end gap-3">
				<button
					onclick={onClose}
					class="px-4 py-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={loading}
					class="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-bold flex items-center gap-2 shadow-lg disabled:opacity-50"
				>
					{#if loading}
						<div class="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin"></div>
					{:else}
						<Save class="w-4 h-4" />
					{/if}
					Save
				</button>
			</div>
		</div>
	</div>
{/if}
