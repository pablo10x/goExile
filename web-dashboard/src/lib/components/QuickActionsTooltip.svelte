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
			case 'right':
				return 'origin-top-left';
			case 'left':
				return 'origin-top-right -translate-x-full';
			case 'bottom':
				return 'origin-top -translate-x-1/2';
			case 'top':
				return 'origin-bottom -translate-x-1/2 -translate-y-full';
			default:
				return 'origin-top-left';
		}
	}

	function getVariantClass(v: string | undefined) {
		switch (v) {
			case 'danger':
				return 'text-red-500 hover:bg-red-600 hover:text-white shadow-lg shadow-red-900/20';
			case 'warning':
				return 'text-amber-500 hover:bg-amber-600 hover:text-white shadow-lg shadow-amber-900/20';
			case 'success':
				return 'text-emerald-500 hover:bg-emerald-600 hover:text-white shadow-lg shadow-emerald-900/20';
			default:
				return 'text-stone-400 hover:bg-white hover:text-black shadow-lg';
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
			class="fixed z-[9999] w-48 bg-black/90 backdrop-blur-xl border border-stone-800 shadow-2xl shadow-black/80 rounded-none overflow-hidden p-1.5 {getPlacementClass(
				placement
			)} industrial-frame"
			style="top: {coords.y}px; left: {coords.x}px;"
			onmouseenter={show}
			onmouseleave={hide}
			role="group"
		>
			{#if title}
				<div
					class="px-3 py-2 text-[10px] font-black text-stone-500 uppercase tracking-[0.2em] border-b border-stone-800 mb-1 flex items-center gap-2"
				>
					<span class="w-1 h-1 bg-rust shadow-[0_0_8px_var(--color-rust)] rounded-full"></span>
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
						class="w-full text-left flex items-center gap-3 px-3 py-2 rounded-none text-xs font-jetbrains font-black uppercase transition-all duration-200 disabled:opacity-20 disabled:cursor-not-allowed group {getVariantClass(
							action.variant
						)}"
					>
						{#if action.icon}
							<action.icon class="w-3.5 h-3.5 transition-transform group-hover:scale-110 group-hover:text-rust" />
						{/if}
						<span class="tracking-widest">{action.label}</span>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
