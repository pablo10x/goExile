<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';

    export let isOpen: boolean = false;
    export let title: string = 'Confirm Action';
    export let message: string = 'Are you sure you want to proceed?';
    export let confirmText: string = 'Confirm';
    export let cancelText: string = 'Cancel';
    export let isCritical: boolean = false; // If true, use red colors

    export let onConfirm: () => Promise<void>;

    const dispatch = createEventDispatcher();

    let loading = false;
    let error: string | null = null;

    async function handleConfirm() {
        loading = true;
        error = null;
        try {
            await onConfirm();
            dispatch('success');
            close();
        } catch (e: any) {
            error = e.message || 'An unexpected error occurred.';
        } finally {
            loading = false;
        }
    }

    function close() {
        isOpen = false;
        error = null;
        loading = false;
        dispatch('close');
    }
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop -->
        <div 
            class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" 
            on:click={close}
            role="button"
            tabindex="0"
            aria-label="Close dialog"
        ></div>

        <!-- Modal -->
        <div 
            class="relative w-full max-w-md bg-slate-900 border border-slate-700 rounded-xl shadow-2xl overflow-hidden"
            transition:scale={{ start: 0.95, duration: 200 }}
        >
            <!-- Content -->
            <div class="p-6">
                <h3 class="text-xl font-bold text-slate-100 mb-2">{title}</h3>
                <p class="text-slate-400 text-sm leading-relaxed">{message}</p>

                {#if error}
                    <div class="mt-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg flex items-start gap-3" transition:fade>
                        <span class="text-red-400 text-lg">⚠️</span>
                        <p class="text-red-400 text-sm font-medium">{error}</p>
                    </div>
                {/if}
            </div>

            <!-- Actions -->
            <div class="px-6 py-4 bg-slate-800/50 border-t border-slate-700 flex justify-end gap-3">
                {#if loading}
                    <div class="flex items-center gap-3 text-slate-400 px-4 py-2">
                        <div class="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
                        <span class="text-sm font-medium">Processing...</span>
                    </div>
                {:else}
                    <button 
                        on:click={close}
                        class="px-4 py-2 text-sm font-semibold text-slate-300 hover:text-white hover:bg-slate-700/50 rounded-lg transition-colors"
                    >
                        {cancelText}
                    </button>
                    <button 
                        on:click={handleConfirm}
                        class={`px-4 py-2 text-sm font-semibold text-white rounded-lg shadow-lg transition-all transform active:scale-95 ${
                            isCritical 
                            ? 'bg-red-600 hover:bg-red-500 shadow-red-900/20' 
                            : 'bg-blue-600 hover:bg-blue-500 shadow-blue-900/20'
                        }`}
                    >
                        {confirmText}
                    </button>
                {/if}
            </div>
        </div>
    </div>
{/if}