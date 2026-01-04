<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Icon from './theme/Icon.svelte';
	import { siteSettings } from '$lib/stores.svelte';

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
		title?: string;
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
		title = '',
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
	const baseClasses = "inline-flex items-center justify-center transition-all focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed active:tranneutral-y-[2px] relative overflow-hidden group font-mono uppercase tracking-widest font-black";

	// Size classes
	const sizeClasses = {
		xs: "text-[8px] px-2.5 py-1.5 gap-1.5",
		sm: "text-[9px] px-4 py-2 gap-2",
		md: "text-[10px] px-6 py-3 gap-2.5",
		lg: "text-[11px] px-8 py-4 gap-3"
	};

	// Industrial tech palette
	const defaultColors = {
		primary: "bg-rust text-white hover:bg-rust-light shadow-xl shadow-rust/20 border-2 border-rust/30",
		secondary: "bg-neutral-950 text-neutral-400 border-2 border-neutral-800 hover:text-white hover:border-neutral-600 hover:bg-neutral-900",
		danger: "bg-red-600/10 text-red-500 border-2 border-red-600/30 hover:bg-red-600 hover:text-white shadow-xl shadow-red-600/10",
		success: "bg-emerald-600/10 text-emerald-500 border-2 border-emerald-600/30 hover:bg-emerald-600 hover:text-white shadow-xl shadow-emerald-600/10",
		warning: "bg-amber-600/10 text-amber-500 border-2 border-amber-600/30 hover:bg-amber-600 hover:text-white shadow-xl shadow-amber-600/10",
		ghost: "bg-transparent text-neutral-500 hover:text-rust-light hover:bg-rust/5 border-2 border-transparent",
		outline: "bg-transparent text-neutral-400 border-2 border-neutral-800 hover:border-rust/50 hover:text-white"
	};

	// Loading spinner size
	const spinnerSizes = {
		xs: "0.75rem",
		sm: "0.875rem",
		md: "1rem",
		lg: "1.25rem"
	};
</script>

{#if href}
	<a
		{href}
		{title}
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-sm"
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
		{title}
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-sm"
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
</style>
