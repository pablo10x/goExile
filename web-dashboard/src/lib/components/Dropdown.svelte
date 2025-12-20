<script lang="ts">
	import { slide } from 'svelte/transition';
	import type { ComponentType, Snippet } from 'svelte';

	let { label = 'Actions', Icon = null, children } = $props<{
		label?: string;
		Icon?: ComponentType | null;
		children?: Snippet;
	}>();

	let isOpen = $state(false);

	function toggle() {
		isOpen = !isOpen;
	}

	function close() {
		isOpen = false;
	}

	function handleClickOutside(event: MouseEvent) {
		if (isOpen && !(event.target as Element).closest('.dropdown-container')) {
			close();
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="relative dropdown-container">
	<button
		onclick={toggle}
		class="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-medium transition-colors border border-slate-700"
	>
		{#if Icon}
			{@const DropdownIcon = Icon}
			<DropdownIcon class="w-4 h-4" />
		{/if}
		{label}
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="w-3 h-3 transition-transform {isOpen ? 'rotate-180' : ''}"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg
		>
	</button>

	{#if isOpen}
		<div
			transition:slide={{ duration: 150 }}
			class="absolute right-0 mt-2 w-48 bg-slate-900 border border-slate-700 rounded-xl shadow-xl z-50 overflow-hidden"
		>
			{#if children}
				{@render children()}
			{/if}
		</div>
	{/if}
</div>