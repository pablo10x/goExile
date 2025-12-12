<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import { cubicOut } from 'svelte/easing';

    export let isOpen: boolean = false;
    export let title: string = 'Confirm Action';
    export let message: string = 'Are you sure you want to proceed?';
    export let confirmText: string = 'Confirm';
    export let cancelText: string = 'Cancel';
    export let isCritical: boolean = false; // If true, use red colors
    export let progress: number | null = null; // 0-100, if set shows progress bar
    export let statusMessage: string | null = null; // Message during loading/progress

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
            // We do NOT close automatically if progress mode was engaged and finished? 
            // Actually, usually the parent will close it, or we close it here.
            // But if there's a progress bar, the parent might want to show "Done" before closing.
            // For now, let's keep the close behavior but allow delay if needed?
            // simpler: keep standard behavior. Parent can control isOpen if they want.
            close();
        } catch (e: any) {
            error = e.message || 'An unexpected error occurred.';
            loading = false; // Stop loading so user can try again or cancel
        }
    }

    function close() {
        if (loading && progress !== null) return; // Prevent closing if bulk action is in progress?
        isOpen = false;
        error = null;
        loading = false;
        dispatch('close');
    }

    // Dynamic icon based on criticality
    $: icon = isCritical 
        ? '<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-red-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>'
        : '<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';

</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" transition:fade={{ duration: 200 }}>
        <!-- Backdrop with blur -->
                    <div 
                        class="absolute inset-0 bg-slate-950/60 backdrop-blur-md" 
                        onclick={!loading ? close : undefined}
                        role="button"
                        tabindex="0"
                        aria-label="Close dialog"
                    ></div>
        <!-- Modal Container -->
        <div 
            class="relative w-full max-w-lg bg-slate-900/90 border border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden ring-1 ring-white/10"
            transition:scale={{ start: 0.95, duration: 300, easing: cubicOut }}
        >
            <!-- Header Pattern -->
            <div class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r {isCritical ? 'from-red-500 via-orange-500 to-red-500' : 'from-blue-500 via-cyan-500 to-blue-500'} opacity-80"></div>

            <!-- Content -->
            <div class="p-6 sm:p-8">
                <div class="flex items-start gap-4">
                    <div class="flex-shrink-0 p-3 rounded-full bg-slate-800/50 border border-slate-700/50">
                        {@html icon}
                    </div>
                    <div class="flex-1">
                        <h3 class="text-xl font-bold text-slate-100 mb-2 tracking-tight">{title}</h3>
                        <div class="text-slate-400 text-sm leading-relaxed space-y-2">
                             {#if loading && statusMessage}
                                <p class="animate-pulse text-slate-300 font-medium">{statusMessage}</p>
                             {:else}
                                <p>{message}</p>
                             {/if}
                        </div>
                    </div>
                </div>

                {#if error}
                    <div class="mt-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl flex items-start gap-3 animate-in fade-in slide-in-from-top-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                        <div>
                            <h4 class="text-sm font-bold text-red-400">Error</h4>
                            <p class="text-red-300/80 text-xs mt-0.5">{error}</p>
                        </div>
                    </div>
                {/if}

                {#if loading && progress !== null}
                    <!-- Progress Bar -->
                    <div class="mt-8 space-y-2" transition:fade>
                        <div class="flex justify-between text-xs font-semibold text-slate-400 uppercase tracking-wider">
                            <span>Progress</span>
                            <span>{Math.round(progress)}%</span>
                        </div>
                        <div class="w-full h-2 bg-slate-800 rounded-full overflow-hidden border border-slate-700/50">
                            <div 
                                class="h-full bg-gradient-to-r {isCritical ? 'from-red-500 to-orange-500' : 'from-blue-500 to-cyan-500'} transition-all duration-300 ease-out relative"
                                style="width: {progress}%"
                            >
                                <div class="absolute inset-0 bg-white/20 animate-pulse"></div>
                            </div>
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Actions -->
            <div class="px-6 py-5 bg-slate-950/30 border-t border-slate-800/50 flex justify-end gap-3">
                {#if loading && progress !== null}
                     <!-- Locked state during progress -->
                     <button disabled class="px-4 py-2 text-sm font-medium text-slate-500 cursor-not-allowed">
                        Please wait...
                     </button>
                {:else if loading}
                    <button disabled class="flex items-center gap-2 px-6 py-2 text-sm font-semibold text-slate-400 bg-slate-800 rounded-lg cursor-not-allowed">
                        <svg class="animate-spin h-4 w-4 text-slate-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        <span>Processing</span>
                    </button>
                {:else}
                    <button 
                        onclick={close}
                        class="px-5 py-2 text-sm font-semibold text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-all"
                    >
                        {cancelText}
                    </button>
                    <button 
                        onclick={handleConfirm}
                        class={`px-5 py-2 text-sm font-semibold text-white rounded-lg shadow-lg hover:shadow-xl transition-all transform active:scale-95 flex items-center gap-2 ${
                            isCritical 
                            ? 'bg-gradient-to-r from-red-600 to-red-500 hover:from-red-500 hover:to-red-400 shadow-red-900/20' 
                            : 'bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 shadow-blue-900/20'
                        }`}
                    >
                        {confirmText}
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 opacity-80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path><path d="M12 5l7 7-7 7"></path></svg>
                    </button>
                {/if}
            </div>
        </div>
    </div>
{/if}