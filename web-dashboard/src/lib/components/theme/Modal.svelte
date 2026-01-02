<script lang="ts">
	import { portal } from '../../actions/portal';
    import type { Snippet } from 'svelte';
	import { siteSettings } from '$lib/stores';
	import { fade, scale, slide } from 'svelte/transition';
	import { backOut, cubicOut } from 'svelte/easing';
	import Button from '../Button.svelte';

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

	// Dynamic transition picker
	function getTransition(node: HTMLElement) {
		const type = $siteSettings.aesthetic.modal_animation || 'scale';
		if (!$siteSettings.aesthetic.animations_enabled) return { duration: 0 };
		
		switch (type) {
			case 'slide': return slide(node, { duration: 400, easing: cubicOut });
			case 'fade': return fade(node, { duration: 300 });
			default: return scale(node, { duration: 400, easing: backOut, start: 0.9 });
		}
	}
</script>

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div 
		use:portal
		class="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-black/80 {$siteSettings.aesthetic.glassmorphism ? 'backdrop-blur-sm' : ''}"
		onclick={close}
		transition:fade={{ duration: 200 }}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<div 
			class="industrial-modal w-full {maxWidth} overflow-hidden shadow-2xl"
			style="border-radius: var(--radius-lg); border: var(--card-border-width) solid var(--border-color);"
			onclick={e => e.stopPropagation()}
			transition:getTransition
			role="document"
		>
			<!-- Header -->
			<div class="flex items-center justify-between bg-black/40 backdrop-blur-md px-6 py-4 border-b border-stone-800">
				<h3 id="modal-title" class="font-heading text-xl text-rust-light tracking-widest uppercase">{title}</h3>
				<button 
					onclick={close}
					class="text-text-dim hover:text-white transition-colors p-1"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
				</button>
			</div>

			<!-- Content -->
			<div class="p-6 md:p-8 max-h-[70vh] overflow-y-auto bg-stone-950 relative custom-scrollbar">
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
				<div class="bg-black/20 px-6 py-4 border-t border-stone-800 text-right">
					{#if footer}
						{@render footer()}
					{:else}
						<Button 
							onclick={close} 
							variant="primary"
							size="sm"
						>
							ACKNOWLEDGE
						</Button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}