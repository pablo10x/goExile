<script lang="ts">
	import { apiFetch } from '$lib/api';
	import type { Snippet } from 'svelte';
	import { fade } from 'svelte/transition';
	import Icon from './Icon.svelte';

	let {
		title = '',
		children,
		actions,
		icon = '',
		class: className = '',
		subtitle = ''
	} = $props<{
		title?: string;
		children: Snippet;
		actions?: Snippet;
		icon?: string;
		class?: string;
		subtitle?: string;
	}>();
</script>

<div
	class="premium-card group {className} relative"
	transition:fade={{ duration: 300 }}
>
	{#if title || icon || actions}
		<div
			class="px-6 py-5 border-b border-white/5 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 relative z-10"
		>
			<div class="flex items-center gap-4">
				{#if icon}
					<div
						class="p-2 bg-sky-500/10 border border-sky-500/20 rounded-xl shadow-sm transition-all duration-300 group-hover:border-sky-500/40"
					>
						<Icon name={icon} size="1.2rem" class="text-sky-400" />
					</div>
				{/if}
				<div class="flex flex-col">
					{#if title}
						<h2 class="text-lg font-bold text-white tracking-tight leading-none uppercase italic font-heading">{title}</h2>
					{/if}
					{#if subtitle}
						<div class="flex items-center mt-1.5">
							<span class="text-xs font-medium text-slate-400 tracking-wide uppercase font-sans">{subtitle}</span>
						</div>
					{/if}
				</div>
			</div>

			{#if actions}
				<div class="flex items-center gap-3 w-full sm:w-auto justify-end relative z-20">
					{@render actions()}
				</div>
			{/if}
		</div>
	{/if}

	<div class="p-0 relative z-10 bg-transparent">
		{@render children()}
	</div>
</div>
