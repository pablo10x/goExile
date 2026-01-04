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
	class="bg-black/40 border-2 border-neutral-800/50 rounded-none overflow-hidden shadow-2xl backdrop-blur-xl group {className} hover:border-rust/30 transition-all duration-500 relative"
    transition:fade={{ duration: 300 }}
>
    <!-- Tactical Corner Brackets -->
    <div class="corner-bracket-tl"></div>
    <div class="corner-bracket-br"></div>

    <!-- Subtle Inner Glow -->
    <div class="absolute inset-0 bg-gradient-to-br from-rust/[0.02] to-transparent pointer-events-none"></div>

    {#if title || icon || actions}
        <div class="px-8 py-6 border-b-2 border-neutral-800/50 bg-neutral-950/20 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 relative z-10">
            <div class="flex items-center gap-5">
                {#if icon}
                    <div class="p-2.5 bg-rust/10 border border-rust/30 rounded-none shadow-inner group-hover:bg-rust/20 transition-all duration-500">
                        <Icon name={icon} size="1.2rem" class="text-rust-light group-hover:text-white transition-colors" />
                    </div>
                {/if}
                <div class="flex flex-col">
                    {#if title}
                        <h2 class="text-lg font-heading font-black text-white tracking-widest uppercase italic group-hover:text-rust-light transition-colors">{title}</h2>
                    {/if}
                    {#if subtitle}
                        <div class="flex items-center gap-3 mt-1.5">
                            <span class="text-[8px] font-mono font-black text-neutral-600 uppercase tracking-[0.3em]">{subtitle}</span>
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

    <div class="p-0 relative z-10 bg-transparent">
        {@render children()}
    </div>
</div>
