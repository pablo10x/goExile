<script lang="ts">
    import { slide } from 'svelte/transition';
    import { ChevronRight, ChevronDown, Folder, Table, Columns, Database } from 'lucide-svelte';

    export let label: string;
    export let type: 'schema' | 'table' | 'column' | 'root';
    export let isOpen: boolean = false;
    export let isLoading: boolean = false;
    export let isSelected: boolean = false;
    export let onToggle: () => void = () => {};
    export let onSelect: () => void = () => {};
</script>

<div class="select-none">
    <div 
        class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors text-sm {isSelected ? 'bg-blue-600 text-white' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}"
        on:click={(e) => { e.stopPropagation(); onSelect(); if (type !== 'column') onToggle(); }}
        role="button"
        tabindex="0"
        on:keydown={(e) => e.key === 'Enter' && onToggle()}
    >
        {#if type !== 'column'}
            <div class="w-4 h-4 flex items-center justify-center transition-transform duration-200 {isOpen ? 'rotate-90' : ''}">
                <ChevronRight class="w-3.5 h-3.5" />
            </div>
        {:else}
            <div class="w-4 h-4"></div> <!-- Spacer -->
        {/if}

        {#if type === 'root'}
            <Database class="w-4 h-4 text-purple-400" />
        {:else if type === 'schema'}
            <Folder class="w-4 h-4 text-yellow-400" />
        {:else if type === 'table'}
            <Table class="w-4 h-4 text-blue-400" />
        {:else}
            <Columns class="w-3.5 h-3.5 text-emerald-400" />
        {/if}

        <span class="truncate font-mono">{label}</span>
        
        {#if isLoading}
            <div class="ml-auto w-3 h-3 rounded-full border-2 border-slate-500 border-t-transparent animate-spin"></div>
        {/if}
    </div>

    {#if isOpen}
        <div class="ml-4 pl-2 border-l border-slate-700/50" transition:slide={{ duration: 200 }}>
            <slot />
        </div>
    {/if}
</div>