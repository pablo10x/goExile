<script lang="ts">
	import { apiFetch } from '$lib/api';
	import type { ComponentType } from 'svelte';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores.svelte';
	import IconComponent from '$lib/components/theme/Icon.svelte';

	type ColorKey = 'rust' | 'emerald' | 'orange' | 'red' | 'purple' | 'cyan' | 'indigo' | 'sky';

	let {
		title,
		value,
		Icon = null,
		iconName = '',
		subValue = '',
		subValueClass = 'font-mono text-slate-500',
		color = 'sky'
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

	const colorMap: Record<ColorKey, { icon: string }> = {
		rust: { icon: 'text-sky-400' },
		emerald: { icon: 'text-emerald-400' },
		orange: { icon: 'text-amber-400' },
		red: { icon: 'text-rose-400' },
		purple: { icon: 'text-indigo-400' },
		cyan: { icon: 'text-cyan-400' },
		indigo: { icon: 'text-indigo-400' },
		sky: { icon: 'text-sky-400' }
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.sky);
	let isHovered = $state(false);
</script>

<div
	class={`group hover:scale-[1.02] cursor-pointer relative overflow-hidden premium-card font-sans`}
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
	role="button"
	tabindex="0"
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isHovered = !isHovered)}
>
	<div
		class="relative z-10"
		class:p-6={!$siteSettings.dashboard.compact_mode}
		class:p-5={$siteSettings.dashboard.compact_mode}
	>
		<div class="flex items-center justify-between mb-4">
			<div class="flex flex-col gap-1">
				<span
					class="text-[11px] font-bold uppercase tracking-wider text-slate-500 transition-colors group-hover:text-slate-400"
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
				<div
					class="pt-2 text-[10px] font-medium text-slate-500 leading-relaxed border-t border-white/5"
				>
					{@html sanitizedSubValue}
				</div>
			{/if}
		</div>
	</div>
</div>
