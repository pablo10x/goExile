<script lang="ts">
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
			return 'text-red-400 font-bold drop-shadow-[0_0_2px_rgba(248,113,113,0.5)]';
		if (line.includes('Warning') || line.includes('Warn')) return 'text-amber-400';
		if (line.includes('Success') || line.includes('Done') || line.includes('Ready')) return 'text-emerald-400 font-bold';
		if (line.includes('Info') || line.includes('unity')) return 'text-blue-300';
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
	class="relative flex flex-col h-full bg-[#0a0a0a] rounded-lg border border-slate-800 overflow-hidden font-mono text-xs shadow-2xl group"
>
	<!-- CRT Overlay Effects -->
	<div class="absolute inset-0 pointer-events-none z-20 overflow-hidden rounded-lg">
		<div class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[length:100%_2px,3px_100%] opacity-20"></div>
		<div class="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-5"></div>
	</div>

	<!-- Terminal Header -->
	<div
		class="relative z-30 px-4 py-2 bg-[#111] border-b border-slate-800 flex justify-between items-center select-none"
	>
		<div class="flex items-center gap-3">
			<div class="flex gap-1.5 opacity-70">
				<div class="w-2.5 h-2.5 rounded-full bg-red-500/20 border border-red-500/50 shadow-[0_0_5px_rgba(239,68,68,0.3)]"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-amber-500/20 border border-amber-500/50 shadow-[0_0_5px_rgba(245,158,11,0.3)]"></div>
				<div class="w-2.5 h-2.5 rounded-full bg-emerald-500/20 border border-emerald-500/50 shadow-[0_0_5px_rgba(16,185,129,0.3)]"></div>
			</div>
			<div class="flex items-center gap-2">
				<svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				<span class="text-slate-400 font-bold tracking-wider text-[11px] uppercase font-mono"
					>{title}</span
				>
			</div>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={copyLogs}
				class="flex items-center gap-1.5 px-2 py-1 rounded bg-slate-800/50 hover:bg-slate-700/50 border border-slate-700 hover:border-slate-600 text-[10px] text-slate-400 hover:text-slate-200 transition-all"
				title="Copy all logs"
			>
				{#if isCopied}
					<svg class="w-3 h-3 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					<span class="text-emerald-400">COPIED</span>
				{:else}
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
					</svg>
					<span>COPY</span>
				{/if}
			</button>
			{#if !autoScroll}
				<button
					onclick={() => {
						autoScroll = true;
					}}
					class="flex items-center gap-1.5 px-2 py-1 rounded bg-amber-500/10 text-amber-500 border border-amber-500/20 hover:bg-amber-500/20 transition-all text-[10px] font-bold animate-pulse"
				>
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
					</svg>
					RESUME
				</button>
			{/if}
		</div>
	</div>

	<!-- Terminal Output -->
	<div
		bind:this={container}
		onscroll={handleScroll}
		class="relative z-10 flex-1 overflow-y-auto p-4 space-y-0.5 custom-scrollbar bg-[#0a0a0a]"
	>
		{#each logs as line}
			<div class={`break-all whitespace-pre-wrap leading-relaxed ${colorize(line)} font-medium opacity-90 hover:opacity-100 transition-opacity`}>
				<span class="opacity-30 mr-2 select-none text-slate-600">$</span>{line}
			</div>
		{/each}
		{#if logs.length === 0}
			<div class="flex flex-col items-center justify-center h-full text-slate-700 gap-2 opacity-50">
				<div class="w-2 h-2 rounded-full bg-slate-700 animate-ping"></div>
				<span class="text-[10px] uppercase tracking-widest">Signal Lost / Waiting for data...</span>
			</div>
		{/if}
		
		<!-- Blinking Cursor at the end -->
		<div class="mt-1 flex items-center gap-2 opacity-50">
			<div class="w-1.5 h-4 bg-slate-500 animate-pulse"></div>
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 10px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: #0a0a0a;
		border-left: 1px solid #1a1a1a;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #262626;
		border: 2px solid #0a0a0a;
		border-radius: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #404040;
	}
</style>
