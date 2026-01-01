<script lang="ts">
	import { notifications, siteSettings } from '$lib/stores';
	import { flip } from 'svelte/animate';
	import { fly } from 'svelte/transition';
	import { CheckCircle2, AlertTriangle, XCircle, Info, X } from 'lucide-svelte';

	function terminalTransition(node: HTMLElement, { duration = 400 }) {
		return {
			duration,
			css: (t: number) => {
				const opacity = t;
				const scaleY = 0.5 + t * 0.5;
				return `
					opacity: ${opacity};
					transform: scaleY(${scaleY});
					transform-origin: top;
					filter: brightness(${0.5 + t * 0.5});
				`;
			}
		};
	}

	function getIcon(type: string) {
		switch (type) {
			case 'success': return CheckCircle2;
			case 'error': return XCircle;
			case 'warning': return AlertTriangle;
			default: return Info;
		}
	}

	function getColorClass(type: string) {
		switch (type) {
			case 'success': return 'text-emerald-500';
			case 'error': return 'text-red-500';
			case 'warning': return 'text-amber-500';
			default: return 'text-blue-400';
		}
	}

	function getBorderClass(type: string) {
		switch (type) {
			case 'success': return 'border-emerald-500/30';
			case 'error': return 'border-red-500/30';
			case 'warning': return 'border-amber-500/30';
			default: return 'border-blue-400/30';
		}
	}
	
	function getBgClass(type: string) {
		switch (type) {
			case 'success': return 'bg-emerald-500/10';
			case 'error': return 'bg-red-500/10';
			case 'warning': return 'bg-amber-500/10';
			default: return 'bg-blue-400/10';
		}
	}
</script>

<div class="fixed top-6 right-6 left-6 sm:left-auto z-[10000] flex flex-col gap-3 pointer-events-none w-auto sm:w-96">
	{#each $notifications as n (n.id)}
		{@const Icon = getIcon(n.type)}
		{@const colorClass = getColorClass(n.type)}
		{@const borderClass = getBorderClass(n.type)}
		{@const bgClass = getBgClass(n.type)}
		
		<div 
			animate:flip={{ duration: 300 }}
			in:terminalTransition={{ duration: $siteSettings.aesthetic.animations_enabled ? 300 : 0 }}
			out:fly={{ x: 100, duration: $siteSettings.aesthetic.animations_enabled ? 300 : 0 }}
			class="pointer-events-auto relative overflow-hidden backdrop-blur-xl bg-[#0a0a0a]/90 border {borderClass} {$siteSettings.aesthetic.industrial_styling ? 'rounded-none' : 'rounded-lg'} shadow-2xl group"
		>
			<!-- Status Bar -->
			<div class="absolute left-0 top-0 bottom-0 w-1 {bgClass.replace('/10', '')}"></div>

			<div class="p-4 pl-5 flex items-start gap-4">
				<!-- Icon -->
				<div class="mt-0.5 shrink-0 {colorClass} drop-shadow-md">
					<Icon size="18" strokeWidth={2.5} />
				</div>

				<!-- Content -->
				<div class="flex-1 min-w-0">
					<div class="flex items-center justify-between gap-2 mb-1">
						<span class="text-[10px] font-black font-jetbrains uppercase tracking-widest {colorClass}">
							SYS_{n.type === 'info' ? 'MSG' : n.type.toUpperCase()}
						</span>
						<span class="text-[8px] font-mono text-stone-600">0x{n.id.slice(0, 4).toUpperCase()}</span>
					</div>
					
					<p class="text-xs font-medium font-sans text-stone-200 leading-relaxed tracking-wide">
						{n.message}
					</p>
					
					{#if n.details}
						<p class="mt-1.5 text-[10px] font-mono text-stone-500 break-all border-l border-stone-800 pl-2">
							{n.details}
						</p>
					{/if}
				</div>

				<!-- Close Button -->
				<button 
					onclick={() => notifications.remove(n.id)}
					class="shrink-0 p-1 text-stone-600 hover:text-white transition-colors"
					aria-label="Dismiss"
				>
					<X size="14" />
				</button>
			</div>

			<!-- Progress Bar -->
			{#if n.timeout > 0}
				<div class="absolute bottom-0 left-1 right-0 h-[2px] bg-stone-900">
					<div 
						class="h-full {colorClass.replace('text-', 'bg-')}"
						style="animation: shrink {n.timeout}ms linear forwards; width: 100%;"
					></div>
				</div>
			{/if}

			<!-- Scanline overlay (optional aesthetic) -->
			{#if $siteSettings.aesthetic.scanline_type !== 'none'}
				<div class="absolute inset-0 pointer-events-none bg-[url('https://www.transparenttextures.com/patterns/dark-matter.png')] opacity-10 mix-blend-overlay"></div>
			{/if}
		</div>
	{/each}
</div>

<style>
	@keyframes shrink {
		from { width: 100%; }
		to { width: 0%; }
	}
</style>