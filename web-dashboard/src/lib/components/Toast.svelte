<script lang="ts">
    import { fade, fly, slide } from 'svelte/transition';
    import { CheckCircle, AlertCircle, Info, XCircle, X, ChevronDown } from 'lucide-svelte';
    import { notifications } from '$lib/stores';
    import type { Notification } from '$lib/stores';

    export let notification: Notification;

    let isHovered = false;
    let showDetails = false;

    const icons = {
        success: CheckCircle,
        error: XCircle,
        warning: AlertCircle,
        info: Info
    };

    const colors = {
        success: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
        error: 'bg-red-500/10 border-red-500/20 text-red-400',
        warning: 'bg-orange-500/10 border-orange-500/20 text-orange-400',
        info: 'bg-blue-500/10 border-blue-500/20 text-blue-400'
    };
    
    const progressColors = {
        success: 'bg-emerald-500',
        error: 'bg-red-500',
        warning: 'bg-orange-500',
        info: 'bg-blue-500'
    };

    function handleMouseEnter() {
        isHovered = true;
    }

    function handleMouseLeave() {
        isHovered = false;
    }
</script>

<div 
    class="pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg border backdrop-blur-md shadow-lg transition-all duration-300 {colors[notification.type]} relative group"
    transition:fly={{ y: 20, duration: 300 }}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    role="alert"
>
    <div class="p-4">
        <div class="flex items-start gap-3">
            <div class="flex-shrink-0 mt-0.5">
                <svelte:component this={icons[notification.type]} class="h-5 w-5" />
            </div>
            <div class="flex-1 w-0">
                <p class="text-sm font-medium leading-5 opacity-90">{notification.message}</p>
                {#if notification.details}
                    <button 
                        class="mt-1 text-xs underline opacity-70 hover:opacity-100 flex items-center gap-1"
                        on:click={() => showDetails = !showDetails}
                    >
                        {showDetails ? 'Hide details' : 'Show details'}
                        <ChevronDown class="w-3 h-3 transition-transform {showDetails ? 'rotate-180' : ''}" />
                    </button>
                {/if}
            </div>
            <div class="ml-4 flex flex-shrink-0">
                <button
                    class="inline-flex rounded-md text-current opacity-50 hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-slate-900"
                    on:click={() => notifications.remove(notification.id)}
                >
                    <span class="sr-only">Close</span>
                    <X class="h-4 w-4" />
                </button>
            </div>
        </div>
        
        {#if showDetails && notification.details}
            <div class="mt-2 text-xs opacity-80 bg-black/20 p-2 rounded overflow-x-auto whitespace-pre-wrap font-mono" transition:slide>
                {notification.details}
            </div>
        {/if}
    </div>
    
    <!-- Progress Bar -->
    {#if notification.timeout && notification.timeout > 0}
         <div 
            class="absolute bottom-0 left-0 h-1 {progressColors[notification.type]} opacity-50"
            style="width: 100%; animation: shrink {notification.timeout}ms linear forwards; animation-play-state: {isHovered ? 'paused' : 'running'};"
        ></div>
    {/if}
</div>

<style>
    @keyframes shrink {
        from { width: 100%; }
        to { width: 0%; }
    }
</style>