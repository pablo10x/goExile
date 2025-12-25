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
			bg: 'bg-rust/10',
			iconBg: 'bg-rust',
			glow: 'shadow-[0_0_20px_rgba(120,53,15,0.2)]',
			accent: 'bg-rust'
		},
		emerald: {
			border: 'border-emerald-900/30',
			text: 'text-emerald-400',
			bg: 'bg-emerald-950/10',
			iconBg: 'bg-emerald-600',
			glow: 'shadow-[0_0_20px_rgba(16,185,129,0.2)]',
			accent: 'bg-emerald-500'
		},
		orange: {
			border: 'border-orange-900/30',
			text: 'text-orange-400',
			bg: 'bg-orange-950/10',
			iconBg: 'bg-orange-600',
			glow: 'shadow-[0_0_20px_rgba(249,115,22,0.2)]',
			accent: 'bg-orange-500'
		},
		red: {
			border: 'border-red-900/30',
			text: 'text-red-400',
			bg: 'bg-red-950/10',
			iconBg: 'bg-red-600',
			glow: 'shadow-[0_0_20px_rgba(239,68,68,0.2)]',
			accent: 'bg-red-500'
		},
		purple: {
			border: 'border-purple-900/30',
			text: 'text-purple-400',
			bg: 'bg-purple-950/10',
			iconBg: 'bg-purple-600',
			glow: 'shadow-[0_0_20px_rgba(168,85,247,0.2)]',
			accent: 'bg-purple-500'
		}
	};

	let colors = $derived(colorMap[color as ColorKey] || colorMap.blue);
</script>

<div

	class={`group relative overflow-hidden transition-all duration-300 bg-stone-900/30 border-2 ${colors.border} ${$siteSettings.aesthetic.industrial_styling ? 'rounded-none' : 'rounded-xl'} hover:border-rust/60 hover:translate-y-[-2px] shadow-[2px_2px_0px_rgba(0,0,0,0.3)] hover:shadow-[4px_4px_0px_rgba(120,53,15,0.1)] cursor-pointer font-primary`}

>

	<!-- Status Bar Accent -->

	<div class={`absolute top-0 left-0 w-full h-0.5 ${colors.accent} opacity-20 group-hover:opacity-100 transition-opacity duration-500`}></div>



	<!-- Tactical Background Pattern -->

	<div class="absolute inset-0 opacity-[0.02] pointer-events-none" style="background-image: radial-gradient(circle at 2px 2px, white 1px, transparent 0); background-size: 32px 32px;"></div>



	<div class="relative z-10 p-5">

		<div class="flex items-center justify-between mb-4">

			<div class="flex flex-col">

				<span class="text-[8px] font-mono uppercase tracking-[0.2em] text-stone-500 mb-0.5">Telemetry_Signal</span>

				<span class="text-[10px] font-bold uppercase tracking-wider text-stone-300 font-primary"

					>{title}</span

				>

			</div>

			{#if Icon}

				{@const CardIcon = Icon}

				<div

					class={`p-1.5 rounded-none border border-white/5 ${colors.iconBg} bg-opacity-20 backdrop-blur-md transition-all duration-500 group-hover:bg-opacity-100`}

				>

					<CardIcon class="w-3.5 h-3.5 text-white" />

				</div>

			{/if}

		</div>



		<div class="space-y-2">

			<div class={`text-3xl font-bold text-white tabular-nums tracking-tight font-primary`}>

				{value}

			</div>

			{#if subValue}

				<div class={`text-[9px] font-mono leading-relaxed uppercase tracking-widest ${subValueClass} opacity-80`}>

					{@html sanitizedSubValue}

				</div>

			{/if}

		</div>

	</div>



	<!-- Terminal ID Corner -->

	<div class="absolute bottom-0 right-0 p-1 opacity-5 group-hover:opacity-20 transition-opacity">

		<span class="font-mono text-[6px] text-white">CH_{title.substring(0,3).toUpperCase()}</span>

	</div>

</div>