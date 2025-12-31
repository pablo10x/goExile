<script lang="ts">
	import { siteSettings } from '$lib/stores';
	import { X, ChevronRight, Terminal } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	export let isOpen: boolean;
	export let title: string;
	export let onClose: () => void;
</script>

{#if isOpen}
	<div class="fixed inset-0 z-[150] flex justify-end">
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/80 backdrop-blur-sm transition-opacity"
			onclick={onClose}
			onkeydown={(e) => e.key === 'Escape' && onClose()}
			role="button"
			tabindex="0"
			aria-label="Close drawer"
			transition:fade={{ duration: 300 }}
		></div>

		<!-- Drawer Content -->
		<div
			class="relative w-full max-w-2xl bg-[var(--terminal-bg)] border-l border-zinc-800 shadow-2xl h-full flex flex-col transform transition-transform duration-300 ease-out glass-panel"
			class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
			class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
			transition:fly={{ x: 100, duration: 400, easing: cubicOut }}
		>
			<!-- CRT Overlay -->
			{#if $siteSettings.aesthetic.crt_effect}
				<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.03] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
			{/if}

			<!-- Header -->
			<div
				class="flex items-center justify-between px-8 py-6 border-b border-zinc-800 bg-[var(--header-bg)] relative z-20"
			>
				<div class="flex items-center gap-4">
					<div class="p-2 bg-rust/10 border border-rust/30 rounded-none shadow-lg shadow-rust/10">
						<Terminal class="w-5 h-5 text-rust" />
					</div>
					<div>
						<h2 class="text-xl font-heading font-black text-white uppercase tracking-tighter">{title}</h2>
						<div class="flex items-center gap-2 mt-1">
							<ChevronRight class="w-3 h-3 text-rust" />
							<span class="font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-widest">DRAWER_PANEL_ACTIVE</span>
						</div>
					</div>
				</div>
				<button
					onclick={onClose}
					class="p-2 text-text-dim hover:text-white hover:bg-red-950/30 hover:border-red-500/30 border border-transparent transition-all rounded-none group"
				>
					<X class="w-6 h-6 group-hover:rotate-90 transition-transform duration-300" />
				</button>
			</div>

			<!-- Content Area -->
			<div class="flex-1 overflow-y-auto p-8 relative bg-[var(--terminal-bg)] custom-scrollbar z-10">
				<!-- Background Grid -->
				<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
				
				<div class="relative z-10">
					<slot />
				</div>
			</div>

			<!-- Footer Decoration -->
			<div class="h-2 bg-stone-900 border-t border-zinc-800 flex items-center px-4 gap-2">
				<div class="w-16 h-0.5 bg-rust/40"></div>
				<div class="w-4 h-0.5 bg-rust/20"></div>
				<div class="flex-1"></div>
				<div class="w-8 h-0.5 bg-stone-700"></div>
			</div>
		</div>
	</div>
{/if}

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: var(--terminal-bg);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #222;
		border: 1px solid #111;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}
</style>