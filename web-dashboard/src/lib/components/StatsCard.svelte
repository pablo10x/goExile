<script lang="ts">
    import type { ComponentType } from 'svelte';

    export let title: string;
    export let value: string | number;
    export let Icon: ComponentType | null = null;
    export let subValue: string = '';
    export let subValueClass: string = 'text-slate-400';
    export let color: 'blue' | 'emerald' | 'orange' | 'red' | 'purple' = 'blue';
    
    const colorMap = {
        blue: { border: 'border-blue-500/20', text: 'text-blue-400', bg: 'bg-blue-500/10' },
        emerald: { border: 'border-emerald-500/20', text: 'text-emerald-400', bg: 'bg-emerald-500/10' },
        orange: { border: 'border-orange-500/20', text: 'text-orange-400', bg: 'bg-orange-500/10' },
        red: { border: 'border-red-500/20', text: 'text-red-400', bg: 'bg-red-500/10' },
        purple: { border: 'border-purple-500/20', text: 'text-purple-400', bg: 'bg-purple-500/10' }
    };
    
    $: colors = colorMap[color] || colorMap.blue;
</script>

<div class={`relative p-5 rounded-xl bg-slate-900/50 border ${colors.border} backdrop-blur-sm transition-all hover:bg-slate-900/80`}>
    <div class="flex items-center justify-between mb-3">
        <span class="text-slate-400 text-xs font-bold uppercase tracking-wider">{title}</span>
        {#if Icon}
            <div class={`p-2 rounded-lg ${colors.bg}`}>
                <svelte:component this={Icon} class={`w-4 h-4 ${colors.text}`} />
            </div>
        {/if}
    </div>
    
    <div class="flex items-baseline gap-2">
        <div class="text-2xl font-bold text-slate-100 tabular-nums tracking-tight">
            {value}
        </div>
        {#if subValue}
            <div class={`text-xs ${subValueClass}`}>
                {@html subValue}
            </div>
        {/if}
    </div>
</div>