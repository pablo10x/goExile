<script lang="ts">
	import { portal } from '../../actions/portal';
    import type { Snippet } from 'svelte';
	import { siteSettings } from '$lib/stores';

	let { 
        show = $bindable(), 
        title, 
        children, 
        footer, 
        hideFooter = false, 
        maxWidth = 'max-w-2xl',
        onclose = () => {} 
    } = $props<{
        show?: boolean;
        title: string;
        children: Snippet;
        footer?: Snippet;
        hideFooter?: boolean;
        maxWidth?: string;
        onclose?: () => void;
    }>();

	function close() {
		show = false;
		onclose();
	}
</script>

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div 
		use:portal
		class="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-[var(--bg-color)]/90 {$siteSettings.aesthetic.glassmorphism ? 'backdrop-blur-sm' : ''} {$siteSettings.aesthetic.animations_enabled ? 'animate-in fade-in duration-300' : ''}"
		onclick={close}
	>
		<div 
			class="industrial-modal w-full {maxWidth} {$siteSettings.aesthetic.industrial_styling ? 'rounded-sm' : 'rounded-2xl'} overflow-hidden {$siteSettings.aesthetic.animations_enabled ? 'animate-in zoom-in-95 slide-in-from-bottom-4 duration-300' : ''}"
			onclick={e => e.stopPropagation()}
		>
			<!-- Header -->
			<div class="flex items-center justify-between bg-[var(--card-bg)] px-6 py-4 border-b border-stone-800">
				<h3 class="font-heading text-xl text-rust-light tracking-widest">{title}</h3>
				<button 
					onclick={close}
					class="text-stone-500 hover:text-white transition-colors p-1"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
				</button>
			</div>

			<!-- Content -->
			<div class="p-6 md:p-8 max-h-[70vh] overflow-y-auto bg-[var(--bg-color)]/50 relative custom-scrollbar">
				<div class="relative z-10">
					{@render children()}
				</div>
				<!-- Grunge Texture inside modal -->
				{#if $siteSettings.aesthetic.industrial_styling}
					<div class="absolute inset-0 opacity-10 pointer-events-none bg-[url('https://www.transparenttextures.com/patterns/dark-matter.png')]"></div>
				{/if}
			</div>

			<!-- Footer -->
			{#if !hideFooter}
				<div class="bg-[var(--card-bg)] px-6 py-4 border-t border-stone-800 text-right">
					{#if footer}
						{@render footer()}
					{:else}
						<button onclick={close} class="industrial-btn">
							ACKNOWLEDGE
						</button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}