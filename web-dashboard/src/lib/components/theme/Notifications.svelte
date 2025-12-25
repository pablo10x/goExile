<script lang="ts">
	import { notifications, siteSettings } from '$lib/stores';
	import { flip } from 'svelte/animate';
	import { fly } from 'svelte/transition';

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
</script>

<div class="fixed top-6 right-6 left-6 sm:left-auto z-[10000] flex flex-col gap-2 pointer-events-none w-auto sm:w-80 lg:w-96">
	{#each $notifications as n (n.id)}
		<div 
			animate:flip={{ duration: 200 }}
			in:terminalTransition={{ duration: $siteSettings.aesthetic.animations_enabled ? 300 : 0 }}
			out:fly={{ x: 40, duration: $siteSettings.aesthetic.animations_enabled ? 300 : 0, opacity: 0 }}
			class="pointer-events-auto bg-[var(--bg-color)]/95 {$siteSettings.aesthetic.glassmorphism ? 'backdrop-blur-md' : ''} border border-stone-800 {$siteSettings.aesthetic.industrial_styling ? 'rounded-sm' : 'rounded-xl'} {$siteSettings.aesthetic.glow_effects ? 'shadow-[0_10px_40px_rgba(0,0,0,0.7)]' : ''} relative overflow-hidden group"
		>
			<!-- Top Status Bar -->
			<div class="flex items-center justify-between px-3 py-1.5 border-b border-stone-800/50 bg-[var(--card-bg)]/30">
				<div class="flex items-center gap-2">
					<div 
						class="w-1.5 h-1.5 rounded-full {n.type === 'success' ? 'bg-green-500' : n.type === 'error' ? 'bg-red-500' : 'bg-rust'} {($siteSettings.aesthetic.glow_effects && $siteSettings.aesthetic.animations_enabled) ? 'animate-pulse' : ''}"
						style={$siteSettings.aesthetic.glow_effects ? `box-shadow: 0 0 8px ${n.type === 'success' ? '#22c55e' : n.type === 'error' ? '#ef4444' : '#92400e'}` : ''}
					></div>
					<span class="text-[8px] font-mono text-stone-500 uppercase tracking-widest">
						SYS_{n.type === 'info' ? 'COMM' : n.type.toUpperCase()}_LOG
					</span>
				</div>
				<span class="text-[7px] font-mono text-stone-700">0x{n.id.slice(0, 8).toUpperCase()}</span>
			</div>

			<div class="p-4 flex items-start gap-3">
				<!-- Content -->
				<div class="flex-1 min-w-0">
					<div class="flex items-baseline gap-2 mb-1">
						<span 
							class="text-[9px] font-mono font-bold {n.type === 'success' ? 'text-green-500' : n.type === 'error' ? 'text-red-500' : 'text-rust'}"
						>
							[{n.type === 'success' ? 'OK' : n.type === 'error' ? 'FAIL' : 'INFO'}]
						</span>
						<span class="text-[7px] font-mono text-stone-600 uppercase">
							{new Date().toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })}
						</span>
					</div>
					<div class="font-mono text-[11px] sm:text-xs text-stone-200 leading-relaxed uppercase tracking-tight break-words relative overflow-hidden">
						<span class="{$siteSettings.aesthetic.animations_enabled ? 'animate-typewriter' : ''} block">{n.message}</span>
						{#if $siteSettings.aesthetic.animations_enabled}
							<span class="inline-block w-1.5 h-3 bg-stone-100/50 animate-cursor ml-1 align-middle"></span>
						{/if}
					</div>
				</div>

				<button 
					onclick={() => notifications.remove(n.id)}
					class="p-1 text-stone-700 hover:text-stone-300 transition-colors group/close"
					aria-label="Dismiss"
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="group-hover/close:rotate-90 transition-transform duration-300"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
				</button>
			</div>

			<!-- Dynamic Timer Bar -->
			<div 
				class="absolute bottom-0 left-0 h-[1px] opacity-30"
				style="
					background: {n.type === 'success' ? '#22c55e' : n.type === 'error' ? '#ef4444' : '#92400e'};
					width: 100%;
					animation: shrink 5s linear forwards;
				"
			></div>

			<!-- Visual Textures -->
			<div class="absolute inset-0 pointer-events-none opacity-5 bg-[repeating-linear-gradient(0deg,transparent,transparent_2px,white_2px,white_4px)]"></div>
			<div class="absolute inset-0 pointer-events-none bg-gradient-to-tr from-stone-950/0 via-transparent to-white/5"></div>
		</div>
	{/each}
</div>

<style>
	@keyframes shrink {
		from { width: 100%; }
		to { width: 0%; }
	}

	@keyframes cursor {
		0%, 100% { opacity: 0; }
		50% { opacity: 1; }
	}

	.animate-cursor {
		animation: cursor 0.8s step-end infinite;
	}

	@keyframes typewriter {
		from { width: 0; }
		to { width: 100%; }
	}

	/* We use a mask-image reveal for a cleaner typewriter feel on text blocks */
	.animate-typewriter {
		mask-image: linear-gradient(to right, white 0%, white var(--progress, 100%), transparent var(--progress, 100%));
		animation: reveal-text 0.5s ease-out forwards;
	}

	@keyframes reveal-text {
		from { opacity: 0; transform: translateX(-4px); }
		to { opacity: 1; transform: translateX(0); }
	}
</style>