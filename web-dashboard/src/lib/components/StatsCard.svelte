<script lang="ts">
	import type { ComponentType } from 'svelte';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores';
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
			border: 'border-rust/30',
			text: 'text-rust-light',
			bg: 'bg-rust/5',
			iconBg: 'bg-rust',
			glow: 'shadow-rust/10',
			accent: 'bg-rust'
		},
		emerald: {
			border: 'border-[var(--color-success)]/20',
			text: 'text-[var(--color-success)]',
			bg: 'bg-[var(--color-success)]/5',
			iconBg: 'bg-[var(--color-success)]',
			glow: 'shadow-[var(--color-success)]/10',
			accent: 'bg-[var(--color-success)]'
		},
		orange: {
			border: 'border-[var(--color-warning)]/20',
			text: 'text-[var(--color-warning)]',
			bg: 'bg-[var(--color-warning)]/5',
			iconBg: 'bg-[var(--color-warning)]',
			glow: 'shadow-[var(--color-warning)]/10',
			accent: 'bg-[var(--color-warning)]'
		},
		red: {
			border: 'border-[var(--color-danger)]/20',
			text: 'text-[var(--color-danger)]',
			bg: 'bg-[var(--color-danger)]/5',
			iconBg: 'bg-[var(--color-danger)]',
			glow: 'shadow-[var(--color-danger)]/10',
			accent: 'bg-[var(--color-danger)]'
		},
		purple: {
			border: 'border-purple-500/20',
			text: 'text-purple-400',
			bg: 'bg-purple-500/5',
			iconBg: 'bg-purple-600',
			glow: 'shadow-purple-500/10',
			accent: 'bg-purple-500'
		},
		cyan: {
			border: 'border-[var(--color-info)]/20',
			text: 'text-[var(--color-info)]',
			bg: 'bg-[var(--color-info)]/5',
			iconBg: 'bg-[var(--color-info)]',
			glow: 'shadow-[var(--color-info)]/10',
			accent: 'bg-[var(--color-info)]'
		}
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.rust);
	let isHovered = $state(false);
</script>

<div
	class={`group modern-industrial-card tactical-border ${!$siteSettings.aesthetic.industrial_styling ? 'rounded-2xl' : ''} cursor-pointer font-primary`}
	class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
	onmouseenter={() => isHovered = true}
	onmouseleave={() => isHovered = false}
	role="button"
	tabindex="0"
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isHovered = !isHovered)}
>
	<!-- Tactical Corners -->
	<div class="corner-tl"></div>
	<div class="corner-tr"></div>
	<div class="corner-bl"></div>
	<div class="corner-br"></div>

	<!-- Status Bar Accent -->
	<div class={`absolute top-0 left-0 w-full h-0.5 ${colors.accent} opacity-30 group-hover:opacity-100 transition-opacity duration-500`}></div>

	<!-- Glass Overlay -->
	<div class="absolute inset-0 bg-black/40 backdrop-blur-sm opacity-50"></div>

	<!-- Tactical Background Pattern -->
	<div class="absolute inset-0 opacity-[0.03] pointer-events-none" style="background-image: radial-gradient(circle at 2px 2px, var(--color-rust) 1px, transparent 0); background-size: 24px 24px;"></div>

	<!-- Hover Intelligence Overlay -->
	<CardHoverOverlay active={isHovered} />

	<div class="relative z-10" class:p-6={!$siteSettings.dashboard.compact_mode} class:p-4={$siteSettings.dashboard.compact_mode}>
		<div class="flex items-center justify-between" class:mb-5={!$siteSettings.dashboard.compact_mode} class:mb-3={$siteSettings.dashboard.compact_mode}>
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-2">
					<div class={`w-1 h-1 rounded-full ${colors.accent} animate-pulse`}></div>
					<span class="text-[9px] font-jetbrains font-black uppercase tracking-[0.2em]" style="color: var(--text-dim)">SIGNAL_STREAM</span>
				</div>
				<span class="text-[11px] font-heading font-black uppercase tracking-widest text-stone-300"
					>{title}</span
				>
			</div>

			{#if iconName || Icon}
				<div
					class={`p-2.5 bg-opacity-20 backdrop-blur-xl transition-all duration-500 group-hover:bg-opacity-100 group-hover:scale-110 shadow-lg ${colors.iconBg}`}
					class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
					class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
					class:!p-1.5={$siteSettings.dashboard.compact_mode}
				>
					{#if iconName}
						<IconComponent name={iconName} size="1rem" class="text-white drop-shadow-md" />
					{:else if Icon}
						{@const CardIcon = Icon}
						<CardIcon class="w-4 h-4 text-white drop-shadow-md" />
					{/if}
				</div>
			{/if}
		</div>

		<div class="space-y-3">
			<div class={`text-4xl font-heading font-black text-white tabular-nums tracking-tighter drop-shadow-sm`} class:!text-2xl={$siteSettings.dashboard.compact_mode}>
				{value}
			</div>
			{#if subValue}
				<div class={`text-[10px] font-jetbrains font-bold leading-relaxed uppercase tracking-wider ${subValueClass} bg-black/30 p-2 backdrop-blur-md`}
					 class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
					 class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
					 class:!text-[8px]={$siteSettings.dashboard.compact_mode}
					 class:!p-1={$siteSettings.dashboard.compact_mode}
				>
					{@html sanitizedSubValue}
				</div>
			{/if}
		</div>
	</div>

	<!-- Terminal ID Corner -->
	<div class="absolute bottom-2 right-3 opacity-10 group-hover:opacity-40 transition-opacity font-jetbrains text-[7px] text-stone-500">
		CORE_SEC::{title.substring(0,3).toUpperCase()}
	</div>
</div>