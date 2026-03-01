<script lang="ts">
import { apiFetch } from "$lib/api";
	import type { ComponentType } from 'svelte';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores.svelte';
	import IconComponent from '$lib/components/theme/Icon.svelte';

	type ColorKey = 'rust' | 'emerald' | 'orange' | 'red' | 'purple' | 'cyan';

	let {
		title,
		value,
		Icon = null,
		iconName = '',
		subValue = '',
		subValueClass = 'tactical-code text-stone-500',
		color = 'rust'
	} = $props<{
		title: string;
		value: string | number;
		Icon?: ComponentType | null;
		iconName?: string;
		subValue?: string;
		subValueClass?: string;
		color?: ColorKey;
	}>();

	let sanitizedSubValue = $derived(
		DOMPurify.sanitize(subValue, {
			ALLOWED_TAGS: ['br', 'b', 'strong', 'i', 'em', 'span'],
			ALLOWED_ATTR: ['class']
		})
	);

	const colorMap: Record<
		ColorKey,
		{ variant: string; icon: string }
	> = {
		rust: { variant: 'card-vibrant-blue', icon: 'text-sky-400' },
		emerald: { variant: 'card-vibrant-emerald', icon: 'text-emerald-400' },
		orange: { variant: 'card-vibrant-amber', icon: 'text-amber-400' },
		red: { variant: 'card-vibrant-rose', icon: 'text-rose-400' },
		purple: { variant: 'card-vibrant-blue', icon: 'text-indigo-400' },
		cyan: { variant: 'card-vibrant-blue', icon: 'text-cyan-400' }
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.rust);
	let isHovered = $state(false);
</script>

<div
	class={`group modern-card hover:scale-[1.02] cursor-pointer relative overflow-hidden bg-slate-800/40 border border-white/5 rounded-2xl backdrop-blur-md shadow-lg hover:border-sky-500/30 transition-all duration-300`}
	onmouseenter={() => isHovered = true}
	onmouseleave={() => isHovered = false}
	role="button"
	tabindex="0"
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isHovered = !isHovered)}
>
	<div class="relative z-10" class:p-6={!$siteSettings.dashboard.compact_mode} class:p-5={$siteSettings.dashboard.compact_mode}>
		<div class="flex items-center justify-between mb-4">
			<div class="flex flex-col gap-1">
				<span class="text-[11px] font-bold uppercase tracking-wider text-slate-500 transition-colors group-hover:text-slate-400"
					>{title}</span
				>
			</div>

			{#if iconName || Icon}
				<div
					class={`p-2.5 rounded-xl bg-slate-900/50 border border-white/5 transition-all duration-300 group-hover:scale-110`}
				>
					{#if iconName}
						<IconComponent name={iconName} size="1.1rem" class={colors.icon} />
					{:else if Icon}
						{@const CardIcon = Icon}
						<CardIcon class={`w-4 h-4 ${colors.icon}`} />
					{/if}
				</div>
			{/if}
		</div>

		<div class="space-y-2">
			<div class={`text-3xl font-sans font-semibold text-slate-100 tracking-tight leading-none`}>
				{value}
			</div>
			{#if subValue}
				<div class="pt-2 text-[10px] font-medium text-slate-500 leading-relaxed border-t border-white/5">
					{@html sanitizedSubValue}
				</div>
			{/if}
		</div>
	</div>
</div>