<script lang="ts">
    import { CheckSquare, Square, Trash2 } from 'lucide-svelte';
    import { slide } from 'svelte/transition';
    import type { Todo } from '$lib/stores';

    let { todo, onToggle, onDelete } = $props<{
        todo: Todo;
        onToggle: (todo: Todo) => void;
        onDelete: (id: number) => void;
    }>();
</script>

<div
    class="flex items-center gap-3 p-3 rounded-xl transition-all group border border-transparent {todo.done
        ? 'bg-slate-800/30 opacity-60'
        : 'bg-slate-800/80 hover:bg-slate-700/80 hover:border-slate-600'}"
    transition:slide|local
>
    <button
        onclick={() => onToggle(todo)}
        class="text-slate-400 hover:text-emerald-400 transition-colors"
    >
        {#if todo.done}
            <CheckSquare class="w-5 h-5 text-emerald-500" />
        {:else}
            <Square class="w-5 h-5" />
        {/if}
    </button>
    <span class="flex-1 text-sm {todo.done ? 'line-through text-slate-500' : 'text-slate-200'}">
        {todo.content}
    </span>
    <button
        onclick={() => onDelete(todo.id)}
        class="text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity p-1"
    >
        <Trash2 class="w-4 h-4" />
    </button>
</div>
