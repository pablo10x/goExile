<script lang="ts">
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
	import Icon from './Icon.svelte';
    import { isConnected, connectionStatus } from '$lib/stores.svelte';

	let { 
        title, 
        subtitle = '', 
        actions,
        icon = ''
    } = $props<{
        title: string;
        subtitle?: string;
        actions?: Snippet;
        icon?: string;
    }>();
</script>

<div class="flex flex-col lg:flex-row lg:justify-between lg:items-center mb-10 sm:mb-12 relative gap-6 sm:gap-8">
	<div
		class="transform transition-all duration-700"
        transition:fade={{ duration: 400 }}
	>
		<div class="flex items-center gap-4 mb-3">
			<div class="h-0.5 w-8 sm:w-12 bg-indigo-600"></div>
			<span class="text-[10px] font-bold text-indigo-500 uppercase tracking-[0.3em]">{subtitle || 'System Module'}</span>
		</div>
		<h1
			class="text-4xl sm:text-5xl lg:text-6xl font-heading font-black text-white tracking-tight uppercase leading-none flex items-center gap-6"
		>
            {#if icon}
                <div class="p-4 bg-indigo-500/10 rounded-3xl border border-indigo-500/20 shadow-2xl shadow-indigo-500/10">
                    <Icon name={icon} size="2.5rem" class="text-indigo-400" />
                </div>
            {/if}
			{title}
		</h1>
		<div class="flex flex-wrap items-center gap-3 sm:gap-4 mt-6">
			<div
				class={`px-4 py-1.5 font-bold text-[9px] sm:text-[10px] uppercase rounded-full flex items-center gap-2.5 ${$isConnected ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-lg shadow-emerald-500/5' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}
			>
				<span class={`w-1.5 h-1.5 rounded-full ${$isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`}></span>
				<span class="truncate max-w-[120px] sm:max-w-none">{$connectionStatus}</span>
			</div>
			<div class="w-px h-4 bg-slate-800 hidden sm:block"></div>
			<span class="text-[9px] sm:text-[10px] font-bold text-slate-500 uppercase tracking-widest italic">Uplink Verified</span>
		</div>
	</div>

	<div class="flex items-center justify-between sm:justify-end gap-4 sm:gap-6 w-full lg:w-auto relative z-20">
		{#if actions}
            {@render actions()}
        {/if}
	</div>
</div>