<script lang="ts">
    import { fade, scale } from 'svelte/transition';
    import { X, Globe, Save, RotateCcw } from 'lucide-svelte';
    import Button from './Button.svelte';
    import { onMount } from 'svelte';
    import { isNative } from '$lib/api';

    let { isOpen = $bindable(), onClose } = $props();

    let serverUrl = $state('http://localhost:8081');
    let isSaving = $state(false);

    onMount(() => {
        if (typeof window !== 'undefined') {
            const stored = localStorage.getItem('server_url');
            if (stored) serverUrl = stored;
        }
    });

    function save() {
        isSaving = true;
        if (typeof window !== 'undefined') {
            localStorage.setItem('server_url', serverUrl);
            setTimeout(() => {
                window.location.reload();
            }, 500);
        }
    }

    function reset() {
        serverUrl = 'http://localhost:8081';
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') {
            isOpen = false;
            onClose?.();
        }
    }
</script>

{#if isOpen}
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div 
        class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-slate-950/90 backdrop-blur-md"
        onclick={e => e.target === e.currentTarget && (isOpen = false)}
        onkeydown={handleKeydown}
        transition:fade={{ duration: 200 }}
        role="dialog"
    >
        <div 
            class="w-full max-w-md bg-slate-900 border border-white/10 rounded-3xl shadow-2xl overflow-hidden"
            transition:scale={{ duration: 200, start: 0.95 }}
        >
            <div class="px-8 py-6 border-b border-white/5 flex items-center justify-between">
                <div class="flex items-center gap-3">
                    <div class="p-2 bg-sky-500/10 rounded-xl">
                        <Globe class="w-5 h-5 text-sky-400" />
                    </div>
                    <div>
                        <h3 class="text-lg font-bold text-white tracking-tight">Connection Settings</h3>
                        <p class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Master Server Configuration</p>
                    </div>
                </div>
                <button 
                    onclick={() => { isOpen = false; onClose?.(); }}
                    class="p-2 text-slate-500 hover:text-white transition-colors"
                >
                    <X size={20} />
                </button>
            </div>

            <div class="p-8 space-y-6">
                <div class="p-4 bg-sky-500/5 border border-sky-500/10 rounded-2xl">
                    <p class="text-xs text-sky-400/80 leading-relaxed">
                        Configure the primary connection endpoint for the Exile Dashboard. 
                        {#if isNative}
                            As a desktop application, you can connect to any reachable server on your network.
                        {:else}
                            Changes will take effect after the application reloads.
                        {/if}
                    </p>
                </div>

                <div class="space-y-2">
                    <label for="serverUrl" class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">
                        Endpoint URL
                    </label>
                    <div class="relative group">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                            <Globe class="h-4 w-4 text-slate-500 group-focus-within:text-sky-400 transition-colors" />
                        </div>
                        <input
                            id="serverUrl"
                            type="text"
                            bind:value={serverUrl}
                            placeholder="http://localhost:8081"
                            class="w-full bg-slate-950 border border-white/5 rounded-2xl pl-11 pr-4 py-3 text-white text-sm font-mono focus:border-sky-500/50 outline-none transition-all shadow-inner"
                        />
                    </div>
                </div>

                <div class="flex items-center gap-2 text-[10px] text-slate-500 font-bold uppercase tracking-widest bg-slate-950/50 p-3 rounded-xl border border-white/5">
                    <div class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></div>
                    <span>Interface reload required after saving</span>
                </div>
            </div>

            <div class="px-8 py-6 bg-slate-950 border-t border-white/5 flex items-center justify-between gap-4">
                <Button onclick={reset} variant="ghost" size="md" icon="ph:arrows-counter-clockwise-bold">
                    Default
                </Button>
                <div class="flex gap-3">
                    <Button onclick={() => { isOpen = false; onClose?.(); }} variant="secondary" size="md">
                        Cancel
                    </Button>
                    <Button onclick={save} loading={isSaving} variant="primary" size="md">
                        Save & Sync
                    </Button>
                </div>
            </div>
        </div>
    </div>
{/if}
