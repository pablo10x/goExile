<script lang="ts">
	import type { ResourceProgressBarProps } from '$lib/types/resource-metrics';

	let {
		label,
		value,
		max,
		color,
		animated = true,
		showThreshold = false,
		threshold = 80,
		height = 'h-1.5'
	}: ResourceProgressBarProps = $props();

	let displayValue = $state(0);
	let animatedWidth = $state(0);

	// Animate value changes
	$effect(() => {
		if (animated) {
			const targetWidth = (value / max) * 100;
			if (value !== displayValue || animatedWidth !== targetWidth) {
				animateValue(displayValue, value, (val) => (displayValue = val));
				animateWidth(animatedWidth, targetWidth, (width) => (animatedWidth = width));
			}
		} else {
			displayValue = value;
			animatedWidth = (value / max) * 100;
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

	function animateWidth(start: number, end: number, callback: (val: number) => void) {
		const duration = 500;
		const startTime = Date.now();

		function update() {
			const elapsed = Date.now() - startTime;
			const progress = Math.min(elapsed / duration, 1);
			const easeProgress = 1 - Math.pow(1 - progress, 3);
			const width = start + (end - start) * easeProgress;
			callback(width);
			if (progress < 1) requestAnimationFrame(update);
		}
		requestAnimationFrame(update);
	}

	function getColorClasses() {
		const colorMap = {
			orange: 'bg-rust',
			blue: 'bg-cyan-500',
			green: 'bg-emerald-500',
			red: 'bg-red-500',
			purple: 'bg-purple-500',
			teal: 'bg-teal-500'
		};
		return colorMap[color as keyof typeof colorMap] || colorMap.blue;
	}

	function getThresholdColor() {
		const percentage = (value / max) * 100;
		if (percentage >= threshold) return 'text-red-500';
		if (percentage >= threshold * 0.8) return 'text-amber-500';
		return 'text-emerald-500';
	}

	function getProgressPercentage() {
		return Math.min((value / max) * 100, 100);
	}
</script>

<div class="space-y-3 font-jetbrains">
	<!-- Label and Value -->
	<div class="flex items-center justify-between">
		<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest">{label}</span>
		<div class="flex items-center gap-3">
			<span class="text-xs font-black text-white tabular-nums tracking-tighter">{displayValue.toFixed(1)}</span>
			{#if showThreshold}
				<span class="{getThresholdColor()} text-[9px] font-black border border-current/20 px-1.5 py-0.5 bg-black/40">
					{getProgressPercentage().toFixed(0)}%_LOAD
				</span>
			{/if}
		</div>
	</div>

	<!-- Progress Bar -->
	<div class="{height} bg-stone-900 border border-stone-800 rounded-none overflow-hidden relative shadow-inner">
		<!-- Progress fill -->
		<div class="absolute top-0 left-0 h-full {getColorClasses()} transition-all duration-500 ease-out shadow-lg shadow-current/20" style="width: {animatedWidth}%">
			<!-- Animated pulse effect -->
			<div class="absolute inset-0 bg-white/10 animate-pulse"></div>
		</div>

		<!-- Threshold indicator -->
		{#if showThreshold}
			<div class="absolute top-0 h-full w-[1px] bg-red-500/40 z-10" style="left: {threshold}%"></div>
		{/if}
	</div>
</div>