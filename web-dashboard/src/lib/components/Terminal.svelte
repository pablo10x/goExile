<script lang="ts">
import { apiFetch } from "$lib/api";
	import { TerminalSquare, Check, Copy, ChevronDown } from 'lucide-svelte';

	interface Props {
		logs?: string[];
		title?: string;
	}

	let { logs = [], title = 'Terminal' }: Props = $props();

	let container = $state<HTMLElement | null>(null);
	let autoScroll = $state(true);
	let isCopied = $state(false);

	// Auto-scroll effect
	$effect(() => {
		if (logs && autoScroll && container) {
			// Use requestAnimationFrame to ensure DOM has updated
			requestAnimationFrame(() => {
				if (container) {
					container.scrollTop = container.scrollHeight;
				}
			});
		}
	});

	function handleScroll() {
		if (!container) return;
		const { scrollTop, scrollHeight, clientHeight } = container;
		// If user scrolls up (more than 20px from bottom), disable auto-scroll
		const isAtBottom = scrollHeight - scrollTop - clientHeight < 20;
		if (autoScroll !== isAtBottom) {
			autoScroll = isAtBottom;
		}
	}

	function colorize(line: string) {
		if (line.includes('Error') || line.includes('Exception') || line.includes('Failed') || line.includes('CRITICAL'))
			return 'text-rose-400 font-semibold';
		if (line.includes('Warning') || line.includes('Warn')) return 'text-amber-400 font-semibold';
		if (line.includes('Success') || line.includes('Done') || line.includes('Ready')) return 'text-emerald-400 font-semibold';
		if (line.includes('Info') || line.includes('unity')) return 'text-indigo-300';
		if (line.includes('DEBUG')) return 'text-slate-500';
		return 'text-slate-300';
	}

	function copyLogs() {
		const text = logs.join('\n');
		navigator.clipboard.writeText(text).then(() => {
			isCopied = true;
			setTimeout(() => (isCopied = false), 2000);
		});
	}
</script>

<div
	class="relative flex flex-col h-full bg-slate-950 rounded-xl border border-white/10 shadow-2xl overflow-hidden font-mono text-xs group"
>
	<!-- Terminal Header -->
	<div
		class="relative z-30 px-4 py-2.5 bg-slate-900/80 backdrop-blur-md border-b border-white/5 flex justify-between items-center select-none"
	>
		<div class="flex items-center gap-4">
			<div class="flex gap-1.5 opacity-80">
				<div class="w-2.5 h-2.5 rounded-full bg-rose-500/80"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-amber-500/80"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-emerald-500/80"></div>
			</div>
			<div class="flex items-center gap-2">
				<TerminalSquare class="w-3.5 h-3.5 text-slate-500" />
				<span class="text-slate-400 font-medium tracking-wide text-[11px]"
					>{title}</span
				>
			</div>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={copyLogs}
				class="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-white/5 hover:bg-white/10 border border-white/5 text-[10px] font-medium text-slate-400 hover:text-white transition-all"
				title="Copy logs"
			>
				{#if isCopied}
					<Check class="w-3 h-3 text-emerald-400" />
					<span class="text-emerald-400">Copied</span>
				{:else}
					<Copy class="w-3 h-3" />
					<span>Copy</span>
				{/if}
			</button>
			{#if !autoScroll}
				<button
					onclick={() => {
						autoScroll = true;
					}}
					class="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 hover:bg-indigo-500/20 transition-all text-[10px] font-medium animate-pulse"
				>
					<ChevronDown class="w-3 h-3" />
					Resume
				</button>
			{/if}
		</div>
	</div>

	<!-- Terminal Output -->
	<div
		bind:this={container}
		onscroll={handleScroll}
		class="relative z-10 flex-1 overflow-y-auto p-4 space-y-0.5 custom-scrollbar bg-slate-950"
	>
		{#each logs as line}
			<div class={`break-all whitespace-pre-wrap leading-relaxed ${colorize(line)} flex items-start gap-3`}>
				<span class="opacity-30 select-none text-slate-600">›</span>
				<span class="flex-1">{line}</span>
			</div>
		{/each}
		{#if logs.length === 0}
			<div class="flex flex-col items-center justify-center h-full text-slate-700 gap-3 opacity-60">
				<TerminalSquare class="w-8 h-8 opacity-50" />
				<span class="text-[10px] uppercase tracking-wider font-medium">Awaiting Output stream...</span>
			</div>
		{/if}
		
		<!-- Blinking Cursor -->
		<div class="mt-1 flex items-center gap-2 opacity-50">
			<div class="w-1.5 h-3 bg-indigo-500 animate-pulse"></div>
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #334155;
		border-radius: 99px;
		border: 2px solid transparent;
		background-clip: content-box;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569;
	}
</style>
