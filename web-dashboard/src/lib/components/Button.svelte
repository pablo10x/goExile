<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Icon from './theme/Icon.svelte';
	import { siteSettings } from '$lib/stores';

	interface Props {
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline' | 'success' | 'warning';
		size?: 'xs' | 'sm' | 'md' | 'lg';
		disabled?: boolean;
		loading?: boolean;
		icon?: string;
		block?: boolean;
		type?: 'button' | 'submit' | 'reset';
		href?: string;
		class?: string;
		children?: any;
		onclick?: (e: MouseEvent) => void;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		icon = '',
		block = false,
		type = 'button',
		href = '',
		class: className = '',
		children,
		onclick
	}: Props = $props();

	const dispatch = createEventDispatcher();

	function handleClick(e: MouseEvent) {
		if (href) return; // Let the link handle it
		if (!disabled && !loading && onclick) {
			onclick(e);
		}
		if (!disabled && !loading) {
			dispatch('click', e);
		}
	}

	// Base classes
	const baseClasses = "inline-flex items-center justify-center transition-all focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed active:translate-y-px relative overflow-hidden group";

	// Size classes
	const sizeClasses = {
		xs: "text-[9px] px-3 py-1.5 gap-1.5",
		sm: "text-[10px] px-4 py-2 gap-2",
		md: "text-[11px] px-6 py-3 gap-2.5",
		lg: "text-xs px-8 py-4 gap-3"
	};

	// Default fallback colors (industrial theme) if store is missing or during load
	const defaultColors = {
		primary: "bg-rust text-white hover:bg-rust-light shadow-lg shadow-rust/20 border border-transparent",
		secondary: "bg-stone-900 text-stone-400 border border-stone-800 hover:text-white hover:border-rust/50 hover:bg-stone-800",
		danger: "bg-red-950/20 text-red-500 border border-red-900/40 hover:bg-red-600 hover:text-white shadow-lg shadow-red-900/10",
		success: "bg-emerald-950/20 text-emerald-500 border border-emerald-900/40 hover:bg-emerald-600 hover:text-white shadow-lg shadow-emerald-900/10",
		warning: "bg-amber-950/20 text-amber-500 border border-amber-900/40 hover:bg-amber-600 hover:text-white shadow-lg shadow-amber-900/10",
		ghost: "bg-transparent text-stone-500 hover:text-rust hover:bg-rust/5 border border-transparent",
		outline: "bg-transparent text-stone-400 border border-stone-700 hover:border-rust hover:text-rust"
	};

	// Loading spinner size
	const spinnerSizes = {
		xs: "0.75rem",
		sm: "0.875rem",
		md: "1rem",
		lg: "1.25rem"
	};

	// Dynamic Style Generation
	// We use a derived state to construct the style string
	let buttonStyle = $derived.by(() => {
		const btnConfig = $siteSettings.aesthetic.buttons;
		if (!btnConfig) return ''; // Fallback to classes

		// Only apply dynamic styles for variants that are configured
		const vConfig = (btnConfig as any)[variant];
		if (!vConfig) return '';

		return `
			border-radius: ${btnConfig.border_radius}px;
			font-weight: ${btnConfig.font_weight};
			text-transform: ${btnConfig.text_transform};
			background-color: ${vConfig.bg_color};
			color: ${vConfig.text_color};
			border-color: ${vConfig.border_color};
			--hover-bg: ${vConfig.hover_bg};
		`;
	});

	// We need to handle hover state manually if we use inline styles for background
	// Or we can use CSS variables. Let's use CSS variables for cleanliness.
</script>

{#if href}
	<a
		{href}
		class="{baseClasses} {sizeClasses[size]} {className} {buttonStyle ? 'custom-btn' : defaultColors[variant]}"
		style={buttonStyle}
		onclick={handleClick}
	>
		{#if icon}
			<Icon name={icon} size={spinnerSizes[size]} />
		{/if}
		
		{@render children?.()}
	</a>
{:else}
	<button
		{type}
		class="{baseClasses} {sizeClasses[size]} {className} {buttonStyle ? 'custom-btn' : defaultColors[variant]}"
		style={buttonStyle}
		{disabled}
		aria-disabled={disabled || loading}
		onclick={handleClick}
	>
		{#if loading}
			<Icon name="ph:arrows-clockwise-bold" size={spinnerSizes[size]} class="animate-spin" />
		{:else if icon}
			<Icon name={icon} size={spinnerSizes[size]} />
		{/if}
		
		{@render children?.()}
	</button>
{/if}

<style>
	.custom-btn {
		/* Use CSS variables set in style attribute */
		border-width: 1px;
		border-style: solid;
	}
	.custom-btn:hover:not(:disabled) {
		background-color: var(--hover-bg) !important;
	}
</style>
