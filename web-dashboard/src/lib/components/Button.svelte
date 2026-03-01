<script lang="ts">
import { apiFetch } from "$lib/api";
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
	const baseClasses = "btn-premium active:scale-95 transition-all duration-300 group tracking-tight";

	// Size classes
	const sizeClasses = {
		xs: "text-[10px] px-3 py-1.5 gap-1.5",
		sm: "text-xs px-4 py-2 gap-2",
		md: "text-sm px-6 py-3 gap-2.5",
		lg: "text-base px-8 py-4 gap-3"
	};

	// Modern professional palette
	const defaultColors = {
		primary: "btn-premium-primary",
		secondary: "btn-premium-secondary",
		danger: "bg-rose-500/10 text-rose-400 border border-rose-500/20 hover:bg-rose-500 hover:text-white shadow-lg shadow-rose-500/10",
		success: "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500 hover:text-white shadow-lg shadow-emerald-500/10",
		warning: "bg-amber-500/10 text-amber-400 border border-amber-500/20 hover:bg-amber-500 hover:text-white shadow-lg shadow-amber-500/10",
		ghost: "bg-transparent text-slate-400 hover:text-slate-100 hover:bg-white/5",
		outline: "bg-transparent text-slate-300 border border-white/10 hover:bg-white/5 hover:border-white/20"
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
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-xl"
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
		class="{baseClasses} {sizeClasses[size]} {className} {defaultColors[variant]} rounded-xl"
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
