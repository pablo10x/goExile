<script lang="ts">
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
			return 'text-red-500 font-bold drop-shadow-[0_0_5px_rgba(239,68,68,0.3)]';
		if (line.includes('Warning') || line.includes('Warn')) return 'text-amber-500 font-bold';
		if (line.includes('Success') || line.includes('Done') || line.includes('Ready')) return 'text-emerald-400 font-bold';
		if (line.includes('Info') || line.includes('unity')) return 'text-rust-light font-bold';
		if (line.includes('DEBUG')) return 'text-stone-600';
		return 'text-stone-300';
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
	class="relative flex flex-col h-full bg-[var(--terminal-bg)] rounded-none border border-stone-800 overflow-hidden font-jetbrains text-[11px] shadow-2xl group glass-panel industrial-frame"
>
	<!-- CRT Overlay Effects -->
	<div class="absolute inset-0 pointer-events-none z-20 overflow-hidden">
		<div class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[length:100%_2px,3px_100%] opacity-10"></div>
		<div class="absolute inset-0 bg-gradient-to-b from-white/5 via-transparent to-transparent opacity-5"></div>
	</div>

	<!-- Terminal Header -->
	<div
		class="relative z-30 px-5 py-3 bg-[var(--header-bg)] border-b border-stone-800 flex justify-between items-center select-none"
	>
		<div class="flex items-center gap-4">
			<div class="flex gap-2 opacity-50">
				<div class="w-2.5 h-2.5 rounded-none bg-red-500/40 border border-red-500/60 shadow-[0_0_8px_rgba(239,68,68,0.2)]"></div>
				<div class="w-2.5 h-2.5 rounded-none bg-amber-500/40 border border-amber-500/60 shadow-[0_0_8px_rgba(245,158,11,0.2)]"></div>
				<div class="w-2.5 h-2.5 rounded-none bg-rust/40 border border-rust/60 shadow-[0_0_8px_rgba(249,115,22,0.2)]"></div>
			</div>
			<div class="flex items-center gap-3">
				<TerminalSquare class="w-4 h-4 text-stone-600" />
				<span class="text-stone-400 font-heading font-black tracking-widest text-[10px] uppercase font-mono"
					>{title}</span
				>
			</div>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={copyLogs}
				class="flex items-center gap-2 px-3 py-1.5 rounded-none bg-stone-900 hover:bg-rust hover:text-white border border-stone-800 hover:border-rust text-[9px] font-heading font-black text-stone-500 transition-all uppercase tracking-widest"
				title="Copy all logs"
			>
				{#if isCopied}
					<Check class="w-3.5 h-3.5 text-emerald-400" />
					<span class="text-emerald-400">COPIED</span>
				{:else}
					<Copy class="w-3.5 h-3.5" />
					<span>CLONE_BUFFER</span>
				{/if}
			</button>
			{#if !autoScroll}
				<button
					onclick={() => {
						autoScroll = true;
					}}
					class="flex items-center gap-2 px-3 py-1.5 rounded-none bg-amber-500/10 text-amber-500 border border-amber-500/30 hover:bg-amber-500/20 transition-all text-[9px] font-heading font-black animate-pulse uppercase tracking-widest"
				>
					<ChevronDown class="w-3.5 h-3.5" />
					RESUME_SYNC
				</button>
			{/if}
		</div>
	</div>

	<!-- Terminal Output -->
	<div
		bind:this={container}
		onscroll={handleScroll}
		class="relative z-10 flex-1 overflow-y-auto p-6 space-y-1 custom-scrollbar bg-[var(--terminal-bg)]"
	>
		<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
		
		{#each logs as line}
			<div class={`break-all whitespace-pre-wrap leading-relaxed ${colorize(line)} font-bold uppercase tracking-tight opacity-80 hover:opacity-100 transition-opacity flex items-start gap-3`}>
				<span class="opacity-20 select-none text-stone-600 font-black">>></span>
				<span class="flex-1">{line}</span>
			</div>
		{/each}
		{#if logs.length === 0}
			<div class="flex flex-col items-center justify-center h-full text-stone-800 gap-4 opacity-40">
				<div class="w-12 h-12 border border-dashed border-stone-800 flex items-center justify-center industrial-frame">
					<div class="w-2 h-2 bg-stone-800 animate-ping"></div>
				</div>
				<span class="font-jetbrains text-[10px] font-black uppercase tracking-[0.4em]">Signal_Lost // Awaiting_Neural_Stream</span>
			</div>
		{/if}
		
		<!-- Blinking Cursor at the end -->
		<div class="mt-2 flex items-center gap-2 opacity-30">
			<div class="w-2 h-4 bg-rust shadow-[0_0_10px_rgba(249,115,22,0.5)] animate-pulse"></div>
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 10px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: var(--header-bg);
		border-left: 1px solid #1a1a1a;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #262626;
		border: 2px solid var(--header-bg);
		border-radius: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #404040;
	}
</style>
