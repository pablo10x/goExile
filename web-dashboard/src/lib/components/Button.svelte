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
	const baseClasses = "inline-flex items-center justify-center transition-all focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed active:translate-y-px relative overflow-hidden group";

	// Size classes
	const sizeClasses = {
		xs: "text-[9px] px-3 py-1.5 gap-1.5",
		sm: "text-[10px] px-4 py-2 gap-2",
		md: "text-[11px] px-6 py-3 gap-2.5",
		lg: "text-xs px-8 py-4 gap-3"
	};

	// Deep Slate tech palette
	const defaultColors = {
		primary: "bg-indigo-600 text-white hover:bg-indigo-500 shadow-lg shadow-indigo-500/20 border border-indigo-500/20",
		secondary: "bg-slate-800/60 text-slate-300 border border-slate-700 hover:text-white hover:border-slate-600 hover:bg-slate-800",
		danger: "bg-red-500/10 text-red-500 border border-red-500/20 hover:bg-red-600 hover:text-white shadow-lg shadow-red-500/10",
		success: "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 hover:bg-emerald-600 hover:text-white shadow-lg shadow-emerald-500/10",
		warning: "bg-amber-500/10 text-amber-500 border border-amber-500/20 hover:bg-amber-600 hover:text-white shadow-lg shadow-amber-500/10",
		ghost: "bg-transparent text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/5 border border-transparent",
		outline: "bg-transparent text-slate-300 border border-slate-700 hover:border-indigo-500/50 hover:text-indigo-400"
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
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-lg"
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
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-lg"
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
