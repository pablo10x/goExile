<script lang="ts">
	import type { ComponentType } from 'svelte';

	export let title: string;
	export let value: string | number;
	export let Icon: ComponentType | null = null;
	export let subValue: string = '';
	export let subValueClass: string = 'text-slate-400';
	export let color: 'blue' | 'emerald' | 'orange' | 'red' | 'purple' = 'blue';

	// Sanitize HTML to prevent XSS - only allow safe formatting tags
	function sanitizeHtml(html: string): string {
		if (!html) return '';

		// First, escape all HTML
		const escaped = html
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&#039;');

		// Then, selectively allow safe tags for formatting
		// Allow: <br>, <b>, <strong>, <i>, <em>, <span class="...">
		return (
			escaped
				.replace(/&lt;br\s*\/?&gt;/gi, '<br>')
				.replace(/&lt;b&gt;/gi, '<b>')
				.replace(/&lt;\/b&gt;/gi, '</b>')
				.replace(/&lt;strong&gt;/gi, '<strong>')
				.replace(/&lt;\/strong&gt;/gi, '</strong>')
				.replace(/&lt;i&gt;/gi, '<i>')
				.replace(/&lt;\/i&gt;/gi, '</i>')
				.replace(/&lt;em&gt;/gi, '<em>')
				.replace(/&lt;\/em&gt;/gi, '</em>')
				// Allow span with only class attribute (no onclick, style, etc.)
				.replace(/&lt;span class=&quot;([a-zA-Z0-9\s\-_]+)&quot;&gt;/gi, '<span class="$1">')
				.replace(/&lt;\/span&gt;/gi, '</span>')
		);
	}

	$: sanitizedSubValue = sanitizeHtml(subValue);

	const colorMap = {
		blue: {
			border: 'border-blue-500/30',
			text: 'text-blue-300',
			bg: 'bg-gradient-to-br from-blue-500/20 to-blue-600/10',
			iconBg: 'bg-gradient-to-br from-blue-500 to-blue-600',
			glow: 'shadow-blue-500/20'
		},
		emerald: {
			border: 'border-emerald-500/30',
			text: 'text-emerald-300',
			bg: 'bg-gradient-to-br from-emerald-500/20 to-emerald-600/10',
			iconBg: 'bg-gradient-to-br from-emerald-500 to-emerald-600',
			glow: 'shadow-emerald-500/20'
		},
		orange: {
			border: 'border-orange-500/30',
			text: 'text-orange-300',
			bg: 'bg-gradient-to-br from-orange-500/20 to-orange-600/10',
			iconBg: 'bg-gradient-to-br from-orange-500 to-orange-600',
			glow: 'shadow-orange-500/20'
		},
		red: {
			border: 'border-red-500/30',
			text: 'text-red-300',
			bg: 'bg-gradient-to-br from-red-500/20 to-red-600/10',
			iconBg: 'bg-gradient-to-br from-red-500 to-red-600',
			glow: 'shadow-red-500/20'
		},
		purple: {
			border: 'border-purple-500/30',
			text: 'text-purple-300',
			bg: 'bg-gradient-to-br from-purple-500/20 to-purple-600/10',
			iconBg: 'bg-gradient-to-br from-purple-500 to-purple-600',
			glow: 'shadow-purple-500/20'
		}
	};

	$: colors = colorMap[color] || colorMap.blue;
</script>

<div
	class={`group relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800/60 to-slate-900/60 border ${colors.border} backdrop-blur-sm transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl ${colors.glow} cursor-pointer`}
>
	<!-- Background gradient overlay -->
	<div
		class={`absolute inset-0 ${colors.bg} opacity-0 group-hover:opacity-100 transition-opacity duration-300`}
	></div>

	<!-- Subtle animated border -->
	<div
		class="absolute inset-0 rounded-2xl bg-gradient-to-r from-transparent via-white/5 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"
	></div>

	<div class="relative z-10 p-6">
		<div class="flex items-center justify-between mb-4">
			<span class="text-slate-400 text-xs font-bold uppercase tracking-wider">{title}</span>
			{#if Icon}
				<div
					class={`p-2.5 rounded-xl ${colors.iconBg} shadow-lg ${colors.glow} transition-transform duration-300 group-hover:scale-110`}
				>
					<svelte:component this={Icon} class="w-5 h-5 text-white" />
				</div>
			{/if}
		</div>

		<div class="space-y-2">
			<div class="text-3xl font-bold text-slate-100 tabular-nums tracking-tight">
				{value}
			</div>
			{#if subValue}
				<div class={`text-sm leading-relaxed ${subValueClass}`}>
					{@html sanitizedSubValue}
				</div>
			{/if}
		</div>
	</div>
</div>
