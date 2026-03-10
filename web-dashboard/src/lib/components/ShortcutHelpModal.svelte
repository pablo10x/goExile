<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Keyboard } from 'lucide-svelte';

	let { isOpen = $bindable(false) } = $props<{ isOpen: boolean }>();

	const shortcuts = [
		{ key: 'Ctrl + K', desc: 'Open Quick Search', cat: 'Global' },
		{ key: 'Esc', desc: 'Close Modals', cat: 'Global' },
		{ key: 'G then D', desc: 'Go to Dashboard', cat: 'Navigation' },
		{ key: 'G then L', desc: 'Go to Logs', cat: 'Navigation' },
		{ key: 'G then P', desc: 'Go to Performance', cat: 'Navigation' },
		{ key: 'G then N', desc: 'Go to Nodes', cat: 'Navigation' },
		{ key: 'G then S', desc: 'Go to Settings', cat: 'Navigation' },
	];

	function close() { isOpen = false; }

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			close();
		}
	}
</script>

{#if isOpen}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div 
		class="fixed inset-0 z-[600] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={close}
		onkeydown={handleKeydown}
		role="dialog"
		tabindex="-1"
	>
		<div 
			class="w-full max-w-md bg-neutral-900 border border-neutral-800 shadow-2xl overflow-hidden rounded-xl"
			transition:scale={{ start: 0.95, duration: 200, easing: cubicOut }}
			onclick={e => e.stopPropagation()}
			onkeydown={e => e.stopPropagation()}
			role="document"
		>
			<div class="px-6 py-4 border-b border-neutral-800 bg-neutral-950/40 flex justify-between items-center">
				<div class="flex items-center gap-3">
					<Keyboard class="w-4 h-4 text-indigo-400" />
					<span class="text-[10px] font-bold text-white uppercase tracking-widest">Keyboard Shortcuts</span>
				</div>
				<button onclick={close} class="text-neutral-500 hover:text-white transition-all" aria-label="Close shortcuts">
					<X class="w-4 h-4" />
				</button>
			</div>

			<div class="p-6 space-y-6">
				{#each ['Global', 'Navigation'] as category}
					<div class="space-y-3">
						<span class="text-[8px] font-bold text-neutral-600 uppercase tracking-[0.3em] block border-b border-neutral-800/50 pb-1">{category}</span>
						<div class="space-y-2">
							{#each shortcuts.filter(s => s.cat === category) as s}
								<div class="flex justify-between items-center group">
									<span class="text-[10px] font-bold text-neutral-400 group-hover:text-neutral-200 transition-colors uppercase tracking-tight">{s.desc}</span>
									<div class="flex gap-1">
										{#each s.key.split(' ') as k}
											{#if k === '+' || k === 'then'}
												<span class="text-[8px] text-neutral-700 self-center px-1">{k}</span>
											{:else}
												<kbd class="px-2 py-1 bg-neutral-950 border border-neutral-800 text-[9px] font-bold text-indigo-400 min-w-[24px] text-center rounded shadow-sm">{k}</kbd>
											{/if}
										{/each}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>

			<div class="px-6 py-3 bg-neutral-950/40 border-t border-neutral-800 flex justify-center text-neutral-600 font-medium">
				<span class="text-[8px] uppercase tracking-widest italic">Management System Shortcuts</span>
			</div>
		</div>
	</div>
{/if}
