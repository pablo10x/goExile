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
	class="flex items-center w-full bg-[#0a0a0a] border-b border-stone-800 overflow-x-auto no-scrollbar"
>
	{#each tabs as tab (tab.id)}
		<div
			animate:flip={{ duration: 200 }}
			transition:fade={{ duration: 150 }}
			class="group relative flex items-center gap-2 px-6 py-4 cursor-pointer border-r border-stone-800/50 select-none transition-all min-w-[140px] max-w-[240px] {activeTabId ===
			tab.id
				? 'bg-[#050505] text-rust border-t-2 border-t-rust'
				: 'text-stone-500 hover:bg-[#050505]/50 hover:text-stone-300 border-t-2 border-t-transparent'}"
			onclick={() => onSelect(tab.id)}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Enter' && onSelect(tab.id)}
		>
			{#if tab.type === 'table'}
				<Table class="w-3.5 h-3.5 shrink-0" />
			{:else if tab.type === 'sql'}
				<FileCode class="w-3.5 h-3.5 shrink-0 text-amber-500" />
			{:else}
				<Database class="w-3.5 h-3.5 shrink-0" />
			{/if}

			<span class="truncate text-[10px] font-black font-heading tracking-widest uppercase flex-1">{tab.label}</span>

			<button
				onclick={(e) => {
					e.stopPropagation();
					onClose(tab.id);
				}}
				class="p-1 rounded-none opacity-0 group-hover:opacity-100 hover:bg-stone-800 text-stone-600 hover:text-red-500 transition-all"
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
