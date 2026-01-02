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
			orange: { text: 'text-rust-light', bg: 'bg-rust', border: 'border-rust/20' },
			blue: { text: 'text-cyan-400', bg: 'bg-cyan-500', border: 'border-cyan-500/20' },
			green: { text: 'text-emerald-400', bg: 'bg-emerald-500', border: 'border-emerald-500/20' },
			red: { text: 'text-red-400', bg: 'bg-red-500', border: 'border-red-500/20' },
			purple: { text: 'text-purple-400', bg: 'bg-purple-500', border: 'border-purple-500/20' },
			teal: { text: 'text-teal-400', bg: 'bg-teal-500', border: 'border-teal-500/20' }
		};
		return map[color as keyof typeof map] || map.blue;
	}

	const c = $derived(getColors());
	const pct = $derived(Math.min((current / (peak || 1)) * 100, 100));
</script>

<div
	class="modern-industrial-card glass-panel p-6 relative overflow-hidden group hover:border-rust/40 transition-all duration-500 shadow-xl industrial-sharp"
>
	<div class="flex justify-between items-start mb-5 relative z-10">
		<div>
			<div class="text-[9px] font-jetbrains font-black text-stone-500 uppercase tracking-[0.3em] mb-1.5 flex items-center gap-2">
				<div class="w-1 h-1 rounded-full {c.bg} shadow-lg shadow-current"></div>
				{title}
			</div>
			<div
				class="text-3xl font-heading font-black text-white tabular-nums tracking-tighter"
			>
				{displayCurrent.toFixed(1)}<span class="text-xs text-stone-600 ml-1 font-jetbrains font-bold uppercase">{unit}</span>
			</div>
		</div>
		<div
			class="p-2.5 rounded-none border border-stone-800 bg-stone-950 text-[10px] font-black font-jetbrains text-stone-500 group-hover:scale-110 group-hover:text-rust transition-all shadow-inner industrial-sharp"
		>
			{icon}
		</div>
	</div>

	<!-- Progress Bar -->
	<div class="h-1 w-full bg-stone-950 border border-stone-800/50 rounded-none overflow-hidden mb-4 p-0 relative z-10 shadow-inner">
		<div
			class="absolute top-0 left-0 h-full {c.bg} transition-all duration-700 ease-out"
			style="width: {pct}%; box-shadow: 0 0 15px currentColor;"
		></div>
	</div>

	<!-- Footer -->
	<div class="flex items-center justify-between text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-widest relative z-10">
		<span class="flex items-center gap-2">MAX_DETECTION: <span class="text-stone-400">{peak.toFixed(1)} {unit}</span></span>
		<div class={`px-2 py-0.5 border ${trend === 'up' ? 'text-red-500 border-red-500/20 bg-red-500/5' : 'text-emerald-500 border-emerald-500/20 bg-emerald-500/5'}`}>
			{pct.toFixed(0)}%_LOAD
		</div>
	</div>
</div>