<script lang="ts">
    import { fade, fly, slide } from 'svelte/transition';
    import LogViewer from './LogViewer.svelte';

    export let isOpen: boolean = false;
    export let spawnerId: number | null = null;
    export let onClose: () => void;
</script>

{#if isOpen}
    <!-- Animated Backdrop -->
    <div
        class="fixed inset-0 z-50 bg-black/70 backdrop-blur-lg flex items-center justify-center p-6"
        transition:fade={{ duration: 300 }}
        onclick={onClose}
        onkeydown={(e) => e.key === 'Escape' && onClose()}
        role="button"
        tabindex="0"
        aria-label="Close log viewer"
    >
        <!-- Floating particles effect -->
        <div class="absolute inset-0 overflow-hidden pointer-events-none">
            {#each Array(20) as _, i}
                <div
                    class="absolute w-1 h-1 bg-blue-400/20 rounded-full animate-pulse"
                    style="left: {Math.random() * 100}%; top: {Math.random() * 100}%; animation-delay: {Math.random() * 3}s; animation-duration: {2 + Math.random() * 2}s;"
                ></div>
            {/each}
        </div>

        <!-- Modal Content -->
        <div
            class="relative w-full max-w-7xl h-[92vh] bg-gradient-to-br from-slate-900 via-slate-900 to-slate-950 border border-slate-700/60 rounded-3xl shadow-2xl overflow-hidden flex flex-col backdrop-blur-xl"
            transition:fly={{ y: 20, duration: 400, easing: (t) => t * (2 - t) }}
            onclick={(e) => e.stopPropagation()}
        >
            <!-- Animated Header -->
            <div
                class="flex items-center justify-between px-8 py-6 border-b border-slate-700/50 bg-gradient-to-r from-slate-800/40 to-slate-900/40 backdrop-blur-md"
                transition:slide={{ duration: 300, delay: 100 }}
            >
                <div class="flex items-center gap-4">
                    <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="9" y1="15" x2="15" y2="15"></line></svg>
                    </div>
                    <div>
                        <h2 class="text-2xl font-bold text-slate-100">Spawner #{spawnerId} Logs</h2>
                        <p class="text-sm text-slate-400 mt-1">Real-time log monitoring and filtering</p>
                    </div>
                </div>
                <button
                    onclick={onClose}
                    class="p-3 text-slate-400 hover:text-white transition-all duration-200 rounded-xl hover:bg-slate-700/50 hover:scale-110 hover:rotate-90"
                    aria-label="Close modal"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
            </div>

            <!-- Log Viewer Content -->
            <div
                class="flex-1 overflow-hidden"
                transition:fade={{ duration: 300, delay: 200 }}
            >
                {#if spawnerId}
                    <LogViewer {spawnerId} />
                {/if}
            </div>
        </div>
    </div>
{/if}