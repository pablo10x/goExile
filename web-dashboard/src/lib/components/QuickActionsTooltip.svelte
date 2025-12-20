<script lang="ts">
	import { scale, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { portal } from '$lib/actions';

	interface Action {
		label: string;
		icon?: any; // Component type
		onClick: () => void;
		variant?: 'default' | 'danger' | 'warning' | 'success';
		disabled?: boolean;
	}

	let { 
		children, 
		actions = [], 
		placement = 'right',
		title = '',
		enabled = true,
		class: className = ''
	} = $props<{
		children: any;
		actions: Action[];
		placement?: 'right' | 'top' | 'bottom' | 'left';
		title?: string;
		enabled?: boolean;
		class?: string;
	}>();

	let visible = $state(false);
	let timeout: any;
	let triggerEl: HTMLElement;
	let coords = $state({ x: 0, y: 0 });

	function updatePosition() {
		if (!triggerEl) return;
		const rect = triggerEl.getBoundingClientRect();
		const offset = 12;

		switch (placement) {
			case 'right':
				coords = { x: rect.right + offset, y: rect.top };
				break;
			case 'left':
				coords = { x: rect.left - offset, y: rect.top };
				break;
			case 'bottom':
				coords = { x: rect.left + rect.width / 2, y: rect.bottom + offset };
				break;
			case 'top':
				coords = { x: rect.left + rect.width / 2, y: rect.top - offset };
				break;
			default:
				coords = { x: rect.right + offset, y: rect.top };
		}
	}

	function show() {
		if (!enabled) return;
		clearTimeout(timeout);
		updatePosition();
		visible = true;
	}

	function hide() {
		timeout = setTimeout(() => {
			visible = false;
		}, 150);
	}

	function getPlacementClass(p: string) {
		switch (p) {
			case 'right': return 'origin-top-left';
			case 'left': return 'origin-top-right -translate-x-full';
			case 'bottom': return 'origin-top -translate-x-1/2';
			case 'top': return 'origin-bottom -translate-x-1/2 -translate-y-full';
			default: return 'origin-top-left';
		}
	}

	function getVariantClass(v: string | undefined) {
		switch (v) {
			case 'danger': return 'text-red-400 hover:bg-red-500/10 hover:text-red-300';
			case 'warning': return 'text-yellow-400 hover:bg-yellow-500/10 hover:text-yellow-300';
			case 'success': return 'text-emerald-400 hover:bg-emerald-500/10 hover:text-emerald-300';
			default: return 'text-slate-300 hover:bg-slate-700/50 hover:text-white';
		}
	}
</script>

<div 
	class="relative inline-flex {className}" 
	role="group" 
	onmouseenter={show} 
	onmouseleave={hide}
	bind:this={triggerEl}
>
	{@render children()}

	{#if visible}
		<div
			use:portal
			transition:scale={{ duration: 200, start: 0.9, opacity: 0, easing: quintOut }}
			class="fixed z-[9999] w-48 bg-slate-900/95 backdrop-blur-xl border border-slate-700/50 shadow-2xl shadow-black/50 rounded-xl overflow-hidden p-1.5 {getPlacementClass(placement)}"
			style="top: {coords.y}px; left: {coords.x}px;"
			onmouseenter={show} 
			onmouseleave={hide}
			role="group"
		>
			{#if title}
				<div class="px-3 py-2 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-700/50 mb-1 flex items-center gap-2">
					<span class="w-1 h-1 bg-blue-500 rounded-full"></span>
					{title}
				</div>
			{/if}
			
			<div class="flex flex-col gap-0.5">
				{#each actions as action}
					<button
						onclick={(e) => {
							e.stopPropagation();
							action.onClick();
							visible = false;
						}}
						disabled={action.disabled}
						class="w-full text-left flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed group {getVariantClass(action.variant)}"
					>
						{#if action.icon}
							<action.icon class="w-4 h-4 transition-transform group-hover:scale-110" />
						{/if}
						<span>{action.label}</span>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
