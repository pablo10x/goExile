<script lang="ts">
	import { slide } from 'svelte/transition';
	import { ChevronRight, Folder, Table, Columns, Database, RefreshCw } from 'lucide-svelte';

	let {
		label,
		type,
		isOpen = false,
		isLoading = false,
		isSelected = false,
		onToggle = () => {},
		onSelect = () => {}
	} = $props<{
		label: string;
		type: 'schema' | 'table' | 'column' | 'root';
		isOpen?: boolean;
		isLoading?: boolean;
		isSelected?: boolean;
		onToggle?: () => void;
		onSelect?: () => void;
	}>();
</script>

<div class="select-none font-jetbrains">
	<div
		class="flex items-center gap-3 px-3 py-2 rounded-none cursor-pointer transition-all text-[11px] font-bold uppercase tracking-tight
		{isSelected
			? 'bg-rust text-white shadow-lg shadow-rust/20'
			: 'text-stone-500 hover:text-white hover:bg-rust/10'}"
		onclick={(e) => {
			e.stopPropagation();
			onSelect();
			if (type !== 'column') onToggle();
		}}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Enter' && onToggle()}
	>
		{#if type !== 'column'}
			<div
				class="w-4 h-4 flex items-center justify-center transition-transform duration-300 {isOpen
					? 'rotate-90 text-rust'
					: ''}"
			>
				<ChevronRight class="w-3.5 h-3.5" />
			</div>
		{:else}
			<div class="w-4 h-4 flex items-center justify-center opacity-20">
				<div class="w-1 h-1 bg-stone-500 rounded-full"></div>
			</div>
		{/if}

		<div class="shrink-0">
			{#if type === 'root'}
				<Database class="w-3.5 h-3.5 {isSelected ? 'text-white' : 'text-purple-500'}" />
			{:else if type === 'schema'}
				<Folder class="w-3.5 h-3.5 {isSelected ? 'text-white' : 'text-amber-500'}" />
			{:else if type === 'table'}
				<Table class="w-3.5 h-3.5 {isSelected ? 'text-white' : 'text-cyan-500'}" />
			{:else}
				<Columns class="w-3.5 h-3.5 {isSelected ? 'text-white' : 'text-emerald-500'}" />
			{/if}
		</div>

		<span class="truncate">{label}</span>

		{#if isLoading}
			<RefreshCw class="ml-auto w-3 h-3 text-rust animate-spin" />
		{/if}
	</div>

	{#if isOpen}
		<div
			class="ml-5 pl-3 border-l border-stone-800/50 space-y-0.5 mt-0.5"
			transition:slide={{ duration: 200 }}
		>
			<slot />
		</div>
	{/if}
</div>