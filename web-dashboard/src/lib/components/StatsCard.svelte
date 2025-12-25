<script lang="ts">
	import type { ComponentType } from 'svelte';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores';

	type ColorKey = 'blue' | 'emerald' | 'orange' | 'red' | 'purple';

	let {
		title,
		value,
		Icon = null,
		subValue = '',
		subValueClass = 'tactical-code text-stone-500',
		color = 'blue'
	} = $props<{
		title: string;
		value: string | number;
		Icon?: ComponentType | null;
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
		blue: {
			border: 'border-rust/30',
			text: 'text-rust-light',
			bg: 'bg-rust/5',
			iconBg: 'bg-rust',
			glow: 'shadow-rust/10',
			accent: 'bg-rust'
		},
		emerald: {
			border: 'border-emerald-500/20',
			text: 'text-emerald-400',
			bg: 'bg-emerald-500/5',
			iconBg: 'bg-emerald-600',
			glow: 'shadow-emerald-500/10',
			accent: 'bg-emerald-500'
		},
		orange: {
			border: 'border-orange-500/20',
			text: 'text-orange-400',
			bg: 'bg-orange-500/5',
			iconBg: 'bg-orange-600',
			glow: 'shadow-orange-500/10',
			accent: 'bg-orange-500'
		},
		red: {
			border: 'border-red-500/20',
			text: 'text-red-400',
			bg: 'bg-red-500/5',
			iconBg: 'bg-red-600',
			glow: 'shadow-red-500/10',
			accent: 'bg-red-500'
		},
		purple: {
			border: 'border-purple-500/20',
			text: 'text-purple-400',
			bg: 'bg-purple-500/5',
			iconBg: 'bg-purple-600',
			glow: 'shadow-purple-500/10',
			accent: 'bg-purple-500'
		}
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.blue);
</script>

<div
	class={`group modern-industrial-card ${$siteSettings.aesthetic.industrial_styling ? 'rounded-none' : 'rounded-2xl'} cursor-pointer font-primary`}
>
	<!-- Status Bar Accent -->
	<div class={`absolute top-0 left-0 w-full h-0.5 ${colors.accent} opacity-30 group-hover:opacity-100 transition-opacity duration-500`}></div>

	<!-- Glass Overlay -->
	<div class="absolute inset-0 bg-black/40 backdrop-blur-sm opacity-50"></div>

	<!-- Tactical Background Pattern -->
	<div class="absolute inset-0 opacity-[0.03] pointer-events-none" style="background-image: radial-gradient(circle at 2px 2px, var(--color-rust) 1px, transparent 0); background-size: 24px 24px;"></div>

	<div class="relative z-10 p-6">
		<div class="flex items-center justify-between mb-5">
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-2">
					<div class={`w-1 h-1 rounded-full ${colors.accent} animate-pulse`}></div>
					<span class="text-[9px] font-jetbrains font-black uppercase tracking-[0.2em] text-stone-500">SIGNAL_STREAM</span>
				</div>
				<span class="text-[11px] font-heading font-black uppercase tracking-widest text-stone-300"
					>{title}</span
				>
			</div>

			{#if Icon}
				{@const CardIcon = Icon}
				<div
					class={`p-2.5 rounded-none border border-white/10 ${colors.iconBg} bg-opacity-20 backdrop-blur-xl transition-all duration-500 group-hover:bg-opacity-100 group-hover:scale-110 shadow-lg industrial-frame`}
				>
					<CardIcon class="w-4 h-4 text-white drop-shadow-md" />
				</div>
			{/if}
		</div>

		<div class="space-y-3">
			<div class={`text-4xl font-heading font-black text-white tabular-nums tracking-tighter drop-shadow-sm`}>
				{value}
			</div>
			{#if subValue}
				<div class={`text-[10px] font-jetbrains font-bold leading-relaxed uppercase tracking-wider ${subValueClass} bg-black/30 p-2 border-l-2 ${colors.border.replace('/20', '/40')} backdrop-blur-md`}>
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