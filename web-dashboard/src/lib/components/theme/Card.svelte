<script lang="ts">
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
    import Icon from './Icon.svelte';

	let { 
        title = '', 
        children, 
        actions,
        icon = '',
        class: className = '',
        subtitle = ''
    } = $props<{
        title?: string;
        children: Snippet;
        actions?: Snippet;
        icon?: string;
        class?: string;
        subtitle?: string;
    }>();
</script>

<div
	class="bg-slate-900/40 border border-slate-800/50 rounded-3xl overflow-hidden shadow-2xl backdrop-blur-xl group {className} hover:border-indigo-500/30 transition-all duration-500"
    transition:fade={{ duration: 300 }}
>
    <!-- Subtle Inner Glow -->
    <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/[0.02] to-transparent pointer-events-none"></div>

    {#if title || icon || actions}
        <div class="px-8 py-7 border-b border-slate-800/50 bg-slate-950/20 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 relative z-10">
            <div class="flex items-center gap-5">
                {#if icon}
                    <div class="p-3 bg-indigo-500/10 border border-indigo-500/20 rounded-2xl shadow-inner group-hover:bg-indigo-500/20 group-hover:scale-105 transition-all duration-500">
                        <Icon name={icon} size="1.35rem" class="text-indigo-400 group-hover:text-indigo-300 transition-colors" />
                    </div>
                {/if}
                <div class="flex flex-col">
                    {#if title}
                        <h2 class="text-xl font-heading font-black text-white tracking-tight uppercase group-hover:text-indigo-100 transition-colors">{title}</h2>
                    {/if}
                    {#if subtitle}
                        <div class="flex items-center gap-2 mt-1.5">
                            <div class="w-1 h-1 rounded-full bg-indigo-500/40"></div>
                            <span class="text-[9px] font-bold text-slate-500 uppercase tracking-[0.2em]">{subtitle}</span>
                        </div>
                    {/if}
                </div>
            </div>
            
            {#if actions}
                <div class="flex items-center gap-4 w-full sm:w-auto justify-end relative z-20">
                    {@render actions()}
                </div>
            {/if}
        </div>
    {/if}

    <div class="p-0 relative z-10">
        {@render children()}
    </div>
</div>
