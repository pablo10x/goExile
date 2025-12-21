<script lang="ts">
	import { onMount, afterUpdate } from 'svelte';

	export let logs: string[] = [];
	export let title: string = 'Terminal';

	let container: HTMLElement;
	let autoScroll = true;

	// Scroll to bottom when logs update, if autoScroll is enabled
	afterUpdate(() => {
		if (autoScroll && container) {
			container.scrollTop = container.scrollHeight;
		}
	});

	function handleScroll() {
		if (!container) return;
		const { scrollTop, scrollHeight, clientHeight } = container;
		// If user scrolls up (more than 20px from bottom), disable auto-scroll
		autoScroll = scrollHeight - scrollTop - clientHeight < 20;
	}

	function colorize(line: string) {
		if (line.includes('Error') || line.includes('Exception') || line.includes('Failed'))
			return 'text-red-400';
		if (line.includes('Warning') || line.includes('Warn')) return 'text-yellow-400';
		if (line.includes('Success') || line.includes('Done')) return 'text-emerald-400';
		if (line.includes('Info') || line.includes('unity')) return 'text-blue-300';
		return 'text-slate-700 dark:text-slate-300';
	}
</script>

<div
	class="flex flex-col h-full bg-[#0d1117] rounded-lg border border-slate-200 dark:border-slate-800 overflow-hidden font-mono text-xs shadow-inner"
>
	<!-- Terminal Header -->
	<div
		class="px-4 py-2 bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center select-none"
	>
		<div class="flex items-center gap-2">
			<div class="flex gap-1.5">
				<div class="w-2.5 h-2.5 rounded-full bg-red-500/20 border border-red-500/50"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-yellow-500/20 border border-yellow-500/50"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-emerald-500/20 border border-emerald-500/50"></div>
			</div>
			<span class="text-slate-500 ml-2 font-semibold tracking-wide text-[10px] uppercase"
				>{title}</span
			>
		</div>
		{#if !autoScroll}
			<button
				on:click={() => {
					autoScroll = true;
				}}
				class="text-[10px] bg-blue-500/10 text-blue-400 px-2 py-0.5 rounded border border-blue-500/20 hover:bg-blue-500/20 transition-colors"
			>
				Resume Scroll
			</button>
		{/if}
	</div>

	<!-- Terminal Output -->
	<div
		bind:this={container}
		on:scroll={handleScroll}
		class="flex-1 overflow-y-auto p-4 space-y-0.5 scrollbar-thin scrollbar-track-transparent scrollbar-thumb-slate-700"
	>
		{#each logs as line}
			<div class={`break-all whitespace-pre-wrap leading-relaxed ${colorize(line)} font-medium`}>
				<span class="opacity-30 mr-2 select-none">$</span>{line}
			</div>
		{/each}
		{#if logs.length === 0}
			<div class="text-slate-600 italic">Waiting for output...</div>
		{/if}
		<!-- Blinking Cursor -->
		<div class="h-4 w-2 bg-slate-500 animate-pulse mt-1 inline-block"></div>
	</div>
</div>
