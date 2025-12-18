<script lang="ts">
	import { X, Database, Table, FileCode } from 'lucide-svelte';
	import { flip } from 'svelte/animate';
	import { fade, slide } from 'svelte/transition';

	let {
		tabs = [],
		activeTabId,
		onSelect,
		onClose
	} = $props<{
		tabs: {
			id: string;
			label: string;
			type: 'table' | 'sql' | 'info' | 'config' | 'roles' | 'backups' | 'browser' | 'functions';
			icon?: any;
		}[];
		activeTabId: string | null;
		onSelect: (id: string) => void;
		onClose: (id: string) => void;
	}>();
</script>

<div
	class="flex items-center w-full bg-slate-950 border-b border-slate-800 overflow-x-auto no-scrollbar"
>
	{#each tabs as tab (tab.id)}
		<div
			animate:flip={{ duration: 200 }}
			transition:fade={{ duration: 150 }}
			class="group relative flex items-center gap-2 px-4 py-3 cursor-pointer border-r border-slate-800/50 select-none transition-all min-w-[120px] max-w-[200px] {activeTabId ===
			tab.id
				? 'bg-slate-900 text-blue-400 border-t-2 border-t-blue-500'
				: 'text-slate-500 hover:bg-slate-900/50 hover:text-slate-300 border-t-2 border-t-transparent'}"
			onclick={() => onSelect(tab.id)}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Enter' && onSelect(tab.id)}
		>
			{#if tab.type === 'table'}
				<Table class="w-4 h-4 shrink-0" />
			{:else if tab.type === 'sql'}
				<FileCode class="w-4 h-4 shrink-0 text-amber-400" />
			{:else}
				<Database class="w-4 h-4 shrink-0" />
			{/if}

			<span class="truncate text-xs font-medium flex-1">{tab.label}</span>

			<button
				onclick={(e) => {
					e.stopPropagation();
					onClose(tab.id);
				}}
				class="p-0.5 rounded-md opacity-0 group-hover:opacity-100 hover:bg-slate-800 text-slate-500 hover:text-red-400 transition-all"
			>
				<X class="w-3 h-3" />
			</button>
		</div>
	{/each}
</div>

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
