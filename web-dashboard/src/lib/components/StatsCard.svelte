<script lang="ts">
    export let title: string;
    export let value: string | number;
    export let borderColorClass: string = 'border-t-blue-500';
    export let valueGradientClass: string = 'from-blue-400 to-purple-400';
    export let subValue: string = '';
    export let subValueClass: string = 'text-slate-100';
</script>

<div class={`stat-card border-t-2 ${borderColorClass} bg-slate-800/50 p-6 rounded-lg shadow-lg`}>
    <div class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">{title}</div>
    {#if subValue}
        <div class={`text-lg font-bold ${subValueClass}`}>
            {@html subValue}
        </div>
    {:else}
        <div class={`text-2xl font-bold bg-gradient-to-r ${valueGradientClass} bg-clip-text text-transparent`}>
            {value}
        </div>
    {/if}
</div>