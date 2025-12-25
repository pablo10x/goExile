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
	class="modern-industrial-card glass-panel p-5 relative overflow-hidden group hover:border-rust/40 transition-all duration-500 shadow-xl"
>
	<div class="flex justify-between items-start mb-4 relative z-10">
		<div>
			<div class="text-[9px] font-jetbrains font-black text-stone-500 uppercase tracking-[0.2em] mb-1">
				{title}
			</div>
			<div
				class="text-3xl font-heading font-black text-white tabular-nums tracking-tighter"
			>
				{displayCurrent.toFixed(1)}<span class="text-xs text-stone-500 ml-1 font-jetbrains">{unit}</span>
			</div>
		</div>
		<div
			class="p-2.5 rounded-none border border-stone-800 bg-stone-900/50 text-xl group-hover:scale-110 group-hover:text-rust transition-all industrial-frame shadow-inner"
		>
			{icon}
		</div>
	</div>

	<!-- Progress Bar -->
	<div class="relative h-1 w-full bg-stone-950 border border-stone-800 rounded-none overflow-hidden mb-3 p-0 relative z-10">
		<div
			class="absolute top-0 left-0 h-full {c.bg.replace('bg-', 'bg-')} transition-all duration-700 ease-out shadow-lg"
			style="width: {pct}%; box-shadow: 0 0 10px rgba(249, 115, 22, 0.2);"
		></div>
	</div>

	<!-- Footer -->
	<div class="flex items-center justify-between text-[9px] font-jetbrains font-bold text-stone-600 uppercase tracking-widest relative z-10">
		<span class="flex items-center gap-2">PEAK_VALUE: <span class="text-stone-400">{peak.toFixed(1)}{unit}</span></span>
		<span class={`px-1.5 py-0.5 border ${trend === 'up' ? 'text-red-500 border-red-500/20 bg-red-500/5' : 'text-emerald-500 border-emerald-500/20 bg-emerald-500/5'}`}>
			{pct.toFixed(0)}%_LOAD
		</span>
	</div>
</div>
