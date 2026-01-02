<script lang="ts">
	import { slide } from 'svelte/transition';
	import type { ComponentType, Snippet } from 'svelte';
	import { ChevronDown } from 'lucide-svelte';

	let {
		label = 'Directives',
		Icon = null,
		children
	} = $props<{
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
		const target = event.target as Element;
		if (isOpen && target && !target.closest('.dropdown-container')) {
			close();
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="relative dropdown-container">
	<button
		onclick={toggle}
		class="flex items-center gap-3 px-5 py-2.5 bg-stone-950 text-stone-400 hover:text-white border border-stone-800 hover:border-rust/50 transition-all duration-300 font-heading font-black text-[10px] uppercase tracking-widest active:translate-y-px shadow-lg industrial-sharp"
	>
		{#if Icon}
			{@const DropdownIcon = Icon}
			<DropdownIcon class="w-3.5 h-3.5" />
		{/if}
		{label}
		<ChevronDown class="w-3.5 h-3.5 text-stone-600 transition-transform duration-300 {isOpen ? 'rotate-180 text-rust' : ''}" />
	</button>

	{#if isOpen}
		<div
			transition:slide={{ duration: 200 }}
			class="absolute right-0 mt-2 w-56 bg-black/80 border border-stone-800 shadow-[0_0_50px_rgba(0,0,0,0.8)] z-50 overflow-hidden industrial-sharp"
		>
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
			<div class="relative z-10 flex flex-col divide-y divide-stone-900">
				{#if children}
					{@render children()}
				{/if}
			</div>
		</div>
	{/if}
</div>