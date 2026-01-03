<script lang="ts">
	import { portal } from '../../actions/portal';
    import type { Snippet } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
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

	function getTransition(node: HTMLElement) {
		return scale(node, { duration: 400, easing: backOut, start: 0.9 });
	}
</script>

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div 
		use:portal
		class="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md"
		onclick={close}
		transition:fade={{ duration: 200 }}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<div 
			class="w-full {maxWidth} overflow-hidden shadow-2xl border border-slate-800 rounded-3xl bg-slate-900/90 backdrop-blur-2xl"
			onclick={e => e.stopPropagation()}
			transition:getTransition
			role="document"
		>
			<!-- Header -->
			<div class="flex items-center justify-between bg-slate-950/40 px-8 py-6 border-b border-slate-800">
				<h3 id="modal-title" class="font-heading text-2xl text-white font-black tracking-tight uppercase italic">{title}</h3>
				<button 
					onclick={close}
					class="text-slate-500 hover:text-white transition-all p-2 rounded-lg hover:bg-white/5"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
				</button>
			</div>

			<!-- Content -->
			<div class="p-8 md:p-10 max-h-[75vh] overflow-y-auto relative custom-scrollbar">
				<div class="relative z-10">
					{@render children()}
				</div>
			</div>

			<!-- Footer -->
			{#if !hideFooter}
				<div class="bg-slate-950/20 px-8 py-6 border-t border-slate-800 flex justify-end gap-4">
					{#if footer}
						{@render footer()}
					{:else}
						<Button 
							onclick={close} 
							variant="secondary"
							size="md"
						>
							CLOSE
						</Button>
						<Button 
							onclick={close} 
							variant="primary"
							size="md"
						>
							ACKNOWLEDGE
						</Button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}