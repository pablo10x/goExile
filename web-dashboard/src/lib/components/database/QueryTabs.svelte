<script lang="ts">
	import { X, Database, Table, FileCode, Layers, Shield, HardDrive, Settings, Code2 } from 'lucide-svelte';
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

	function getIcon(type: string) {
		switch (type) {
			case 'table': return Table;
			case 'sql': return FileCode;
			case 'browser': return Layers;
			case 'roles': return Shield;
			case 'backups': return HardDrive;
			case 'config': return Settings;
			case 'functions': return Code2;
			default: return Database;
		}
	}
</script>

<div
	class="flex items-center w-full bg-neutral-900/40 border-b border-neutral-800 overflow-x-auto no-scrollbar backdrop-blur-md rounded-t-2xl"
>
	{#each tabs as tab (tab.id)}
		{@const TabIcon = getIcon(tab.type)}
		<div
			animate:flip={{ duration: 200 }}
			transition:fade={{ duration: 150 }}
			class="group relative flex items-center gap-3 px-6 py-4 cursor-pointer border-r border-neutral-800 select-none transition-all min-w-[150px] max-w-[220px]
			{activeTabId === tab.id
				? 'bg-indigo-500/10 text-indigo-400 border-b-2 border-b-indigo-500'
				: 'text-neutral-500 hover:bg-neutral-800/50 hover:text-neutral-300 border-b-2 border-b-transparent'}"
			onclick={() => onSelect(tab.id)}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Enter' && onSelect(tab.id)}
		>
			<TabIcon class="w-4 h-4 shrink-0 {activeTabId === tab.id ? 'text-indigo-400' : 'text-neutral-600 group-hover:text-neutral-400'}" />

			<span class="truncate text-[10px] font-bold uppercase tracking-widest flex-1">{tab.label}</span>

			{#if tabs.length > 1}
				<button
					onclick={(e) => {
						e.stopPropagation();
						onClose(tab.id);
					}}
					class="p-1 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-neutral-800 text-neutral-600 hover:text-red-400 transition-all"
				>
					<X class="w-3 h-3" />
				</button>
			{/if}
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