<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Keyboard, Mouse } from 'lucide-svelte';
	import Icon from './theme/Icon.svelte';

	let { isOpen = $bindable(false) } = $props<{ isOpen: boolean }>();

	const shortcuts = [
		{ key: 'Ctrl + K', desc: 'Open Command Palette (Omni-Search)', cat: 'Global' },
		{ key: 'Esc', desc: 'Abort / Close Modals', cat: 'Global' },
		{ key: 'G then D', desc: 'Jump to Dashboard', cat: 'Navigation' },
		{ key: 'G then L', desc: 'Jump to Kernel Logs', cat: 'Navigation' },
		{ key: 'G then P', desc: 'Jump to Performance', cat: 'Navigation' },
		{ key: 'G then N', desc: 'Jump to Node Fleet', cat: 'Navigation' },
		{ key: 'G then T', desc: 'Jump to Theme Lab', cat: 'Navigation' },
	];

	function close() { isOpen = false; }
</script>

{#if isOpen}
	<div 
		class="fixed inset-0 z-[600] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={close}
	>
		<div 
			class="w-full max-w-md bg-neutral-900 border border-neutral-800 shadow-2xl overflow-hidden industrial-sharp"
			transition:scale={{ start: 0.95, duration: 200, easing: cubicOut }}
			onclick={e => e.stopPropagation()}
		>
			<div class="px-6 py-4 border-b border-neutral-800 bg-neutral-950/40 flex justify-between items-center">
				<div class="flex items-center gap-3">
					<Keyboard class="w-4 h-4 text-indigo-400" />
					<span class="text-[10px] font-black text-white uppercase tracking-widest">Tactical_Shortcuts</span>
				</div>
				<button onclick={close} class="text-neutral-500 hover:text-white transition-all">
					<X class="w-4 h-4" />
				</button>
			</div>

			<div class="p-6 space-y-6">
				{#each ['Global', 'Navigation'] as category}
					<div class="space-y-3">
						<span class="text-[8px] font-black text-neutral-600 uppercase tracking-[0.3em] block border-b border-neutral-800/50 pb-1">{category}</span>
						<div class="space-y-2">
							{#each shortcuts.filter(s => s.cat === category) as s}
								<div class="flex justify-between items-center group">
									<span class="text-[10px] font-bold text-neutral-400 group-hover:text-neutral-200 transition-colors uppercase tracking-tight">{s.desc}</span>
									<div class="flex gap-1">
										{#each s.key.split(' ') as k}
											{#if k === '+' || k === 'then'}
												<span class="text-[8px] text-neutral-700 self-center px-1">{k}</span>
											{:else}
												<kbd class="px-2 py-1 bg-neutral-950 border border-neutral-800 text-[9px] font-black text-indigo-400 min-w-[24px] text-center rounded shadow-sm">{k}</kbd>
											{/if}
										{/each}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>

			<div class="px-6 py-3 bg-neutral-950/40 border-t border-neutral-800 flex justify-center">
				<span class="text-[8px] font-black text-neutral-700 uppercase tracking-widest italic">Acknowledge_Uplink_Protocols</span>
			</div>
		</div>
	</div>
{/if}
