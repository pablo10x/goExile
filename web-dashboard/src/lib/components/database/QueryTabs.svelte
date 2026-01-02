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
	class="flex items-center w-full bg-slate-900/80 border-b border-slate-800 overflow-x-auto no-scrollbar backdrop-blur-sm"
>
	{#each tabs as tab (tab.id)}
		{@const TabIcon = getIcon(tab.type)}
		<div
			animate:flip={{ duration: 200 }}
			transition:fade={{ duration: 150 }}
			class="group relative flex items-center gap-2.5 px-5 py-3 cursor-pointer border-r border-slate-800/50 select-none transition-all min-w-[140px] max-w-[200px]
			{activeTabId === tab.id
				? 'bg-slate-800/50 text-blue-400 border-b-2 border-b-blue-500'
				: 'text-slate-500 hover:bg-slate-800/30 hover:text-slate-300 border-b-2 border-b-transparent'}"
			onclick={() => onSelect(tab.id)}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Enter' && onSelect(tab.id)}
		>
			<TabIcon class="w-4 h-4 shrink-0 {activeTabId === tab.id ? 'text-blue-400' : 'text-slate-500 group-hover:text-slate-400'}" />

			<span class="truncate text-xs font-medium flex-1">{tab.label}</span>

			{#if tabs.length > 1}
				<button
					onclick={(e) => {
						e.stopPropagation();
						onClose(tab.id);
					}}
					class="p-1 rounded-md opacity-0 group-hover:opacity-100 hover:bg-slate-700/50 text-slate-500 hover:text-red-400 transition-all"
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