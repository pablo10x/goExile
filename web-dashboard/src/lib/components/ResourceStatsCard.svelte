<script lang="ts">
	import type { ResourceStatsCardProps } from '$lib/types/resource-metrics';

	let {
		title,
		current,
		peak,
		unit,
		icon,
		color,
		trend,
		animated = true
	}: ResourceStatsCardProps = $props();

	let displayCurrent = $state(0);

	// Animate number changes
	$effect(() => {
		if (animated) {
			const targetCurrent = current;
			if (targetCurrent !== displayCurrent) {
				animateValue(displayCurrent, targetCurrent, (val) => (displayCurrent = val));
			}
		} else {
			displayCurrent = current;
		}
	});

	function animateValue(start: number, end: number, callback: (val: number) => void) {
		const duration = 500;
		const startTime = Date.now();

		function update() {
			const elapsed = Date.now() - startTime;
			const progress = Math.min(elapsed / duration, 1);
			const easeProgress = 1 - Math.pow(1 - progress, 3);
			const value = start + (end - start) * easeProgress;
			callback(value);
			if (progress < 1) requestAnimationFrame(update);
		}
		requestAnimationFrame(update);
	}

	function getColors() {
		const map = {
			orange: { text: 'text-orange-400', bg: 'bg-orange-500', border: 'border-orange-500/20' },
			blue: { text: 'text-blue-400', bg: 'bg-blue-500', border: 'border-blue-500/20' },
			green: { text: 'text-emerald-400', bg: 'bg-emerald-500', border: 'border-emerald-500/20' },
			red: { text: 'text-red-400', bg: 'bg-red-500', border: 'border-red-500/20' },
			purple: { text: 'text-purple-400', bg: 'bg-purple-500', border: 'border-purple-500/20' },
			teal: { text: 'text-teal-400', bg: 'bg-teal-500', border: 'border-teal-500/20' }
		};
		return map[color] || map.blue;
	}

	const c = $derived(getColors());
	const pct = $derived(Math.min((current / (peak || 1)) * 100, 100));
</script>

<div
	class="bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-xl p-4 relative overflow-hidden group hover:border-slate-300 dark:border-slate-700 transition-colors"
>
	<div class="flex justify-between items-start mb-3">
		<div>
			<div class="text-[10px] uppercase font-bold text-slate-500 tracking-wider mb-0.5">
				{title}
			</div>
			<div
				class="text-2xl font-mono font-medium text-slate-800 dark:text-slate-200 tabular-nums tracking-tight"
			>
				{displayCurrent.toFixed(1)}<span class="text-sm text-slate-500 ml-0.5">{unit}</span>
			</div>
		</div>
		<div
			class="p-2 rounded-lg bg-slate-800/50 text-lg opacity-80 group-hover:scale-110 transition-transform"
		>
			{icon}
		</div>
	</div>

	<!-- Progress Bar -->
	<div class="relative h-1.5 w-full bg-slate-800 rounded-full overflow-hidden mb-2">
		<div
			class="absolute top-0 left-0 h-full {c.bg} transition-all duration-500 ease-out"
			style="width: {pct}%"
		></div>
	</div>

	<!-- Footer -->
	<div class="flex items-center justify-between text-[10px] text-slate-500 font-mono">
		<span>PEAK: {peak.toFixed(1)}{unit}</span>
		<span class={trend === 'up' ? c.text : 'text-slate-600'}>
			{pct.toFixed(0)}%
		</span>
	</div>
</div>
