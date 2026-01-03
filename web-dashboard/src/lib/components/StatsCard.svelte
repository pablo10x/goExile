<script lang="ts">
	import type { ComponentType } from 'svelte';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores.svelte';
	import IconComponent from '$lib/components/theme/Icon.svelte';
	import CardHoverOverlay from '$lib/components/theme/CardHoverOverlay.svelte';

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
		{ border: string; text: string; bg: string; iconBg: string; glow: string; accent: string }
	> = {
		rust: {
			border: 'border-indigo-500/20',
			text: 'text-indigo-600',
			bg: 'bg-indigo-500/5',
			iconBg: 'bg-indigo-600',
			glow: 'shadow-indigo-500/10',
			accent: 'bg-indigo-600'
		},
		emerald: {
			border: 'border-emerald-500/20',
			text: 'text-emerald-600',
			bg: 'bg-emerald-500/5',
			iconBg: 'bg-emerald-600',
			glow: 'shadow-emerald-500/10',
			accent: 'bg-emerald-600'
		},
		orange: {
			border: 'border-orange-500/20',
			text: 'text-orange-600',
			bg: 'bg-orange-500/5',
			iconBg: 'bg-orange-600',
			glow: 'shadow-orange-500/10',
			accent: 'bg-orange-600'
		},
		red: {
			border: 'border-red-500/20',
			text: 'text-red-600',
			bg: 'bg-red-500/5',
			iconBg: 'bg-red-600',
			glow: 'shadow-red-500/10',
			accent: 'bg-red-600'
		},
		purple: {
			border: 'border-purple-500/20',
			text: 'text-purple-600',
			bg: 'bg-purple-500/5',
			iconBg: 'bg-purple-600',
			glow: 'shadow-purple-500/10',
			accent: 'bg-purple-600'
		},
		cyan: {
			border: 'border-cyan-500/20',
			text: 'text-cyan-600',
			bg: 'bg-cyan-500/5',
			iconBg: 'bg-cyan-600',
			glow: 'shadow-cyan-500/10',
			accent: 'bg-cyan-600'
		}
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.rust);
	let isHovered = $state(false);
</script>

<div
	class="group modern-card cursor-pointer font-primary relative hover:shadow-xl hover:shadow-indigo-500/5 hover:-translate-y-1"
	onmouseenter={() => isHovered = true}
	onmouseleave={() => isHovered = false}
	role="button"
	tabindex="0"
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isHovered = !isHovered)}
	style="transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.2s ease, border-color 0.2s ease; contain: content; will-change: transform;"
>
	<!-- Status Bar Accent -->
	<div class={`absolute top-0 left-0 w-full h-1 ${colors.accent} opacity-40 group-hover:opacity-100 transition-opacity duration-500`}></div>

	<!-- Glass Overlay -->
	<div class="absolute inset-0 backdrop-blur-md bg-slate-950/40"></div>

	<div class="relative z-10" class:p-6={!$siteSettings.dashboard.compact_mode} class:p-4={$siteSettings.dashboard.compact_mode}>
		<div class="flex items-center justify-between" class:mb-5={!$siteSettings.dashboard.compact_mode} class:mb-3={$siteSettings.dashboard.compact_mode}>
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-2">
					<div class={`w-1.5 h-1.5 rounded-full ${colors.accent} animate-pulse shadow-[0_0_8px_currentColor]`}></div>
					<span class="text-[9px] font-bold uppercase tracking-widest text-slate-400">System Telemetry</span>
				</div>
				<span class="text-[11px] font-black uppercase tracking-widest text-slate-200"
					>{title}</span
				>
			</div>

			{#if iconName || Icon}
				<div
					class={`p-2.5 bg-opacity-10 transition-all duration-500 group-hover:bg-opacity-20 group-hover:scale-110 shadow-sm ${colors.iconBg} rounded-xl border ${colors.border}`}
				>
					{#if iconName}
						<IconComponent name={iconName} size="1rem" class={`${colors.text} group-hover:text-white transition-colors`} />
					{:else if Icon}
						{@const CardIcon = Icon}
						<CardIcon class={`w-4 h-4 ${colors.text} group-hover:text-white transition-colors`} />
					{/if}
				</div>
			{/if}
		</div>

		<div class="space-y-3">
			<div class={`text-4xl font-heading font-black text-white tabular-nums tracking-tighter shadow-black drop-shadow-lg`}>
				{value}
			</div>
			{#if subValue}
				<div class={`text-[10px] font-bold leading-relaxed uppercase tracking-widest ${subValueClass} bg-slate-900/50 border border-slate-800 p-2 rounded-lg`}
				>
					{@html sanitizedSubValue}
				</div>
			{/if}
		</div>
	</div>
</div>