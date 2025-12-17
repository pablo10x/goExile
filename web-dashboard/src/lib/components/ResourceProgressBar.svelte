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
		height = 'h-2'
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
		const duration = 500; // 500ms animation
		const startTime = Date.now();

		function update() {
			const elapsed = Date.now() - startTime;
			const progress = Math.min(elapsed / duration, 1);
			const easeProgress = 1 - Math.pow(1 - progress, 3); // Ease out cubic
			const value = start + (end - start) * easeProgress;

			callback(value);

			if (progress < 1) {
				requestAnimationFrame(update);
			}
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

			if (progress < 1) {
				requestAnimationFrame(update);
			}
		}

		requestAnimationFrame(update);
	}

	function getColorClasses() {
		const baseClasses =
			'relative overflow-hidden rounded-full transition-all duration-500 ease-out';
		const colorMap = {
			orange: 'from-orange-500 to-red-500',
			blue: 'from-blue-500 to-purple-500',
			green: 'from-green-500 to-teal-500',
			red: 'from-red-500 to-pink-500',
			purple: 'from-purple-500 to-indigo-500',
			teal: 'from-teal-500 to-cyan-500'
		};

		return `${baseClasses} bg-gradient-to-r ${colorMap[color] || colorMap.blue}`;
	}

	function getThresholdColor() {
		const percentage = (value / max) * 100;
		if (percentage >= threshold) {
			return 'text-red-400';
		} else if (percentage >= threshold * 0.8) {
			return 'text-yellow-400';
		}
		return 'text-green-400';
	}

	function getProgressPercentage() {
		return Math.min((value / max) * 100, 100);
	}
</script>

<div class="space-y-2">
	<!-- Label and Value -->
	<div class="flex items-center justify-between text-sm">
		<span class="text-slate-300 font-medium">{label}</span>
		<div class="flex items-center gap-2">
			<span class="text-slate-100 font-mono tabular-nums">{displayValue.toFixed(1)}</span>
			{#if showThreshold}
				<span class="{getThresholdColor()} font-medium">
					{getProgressPercentage().toFixed(0)}%
				</span>
			{/if}
		</div>
	</div>

	<!-- Progress Bar -->
	<div class="{height} bg-slate-700/50 rounded-full overflow-hidden relative">
		<!-- Background gradient -->
		<div class="absolute inset-0 bg-gradient-to-r from-slate-800/50 to-slate-700/30"></div>

		<!-- Progress fill -->
		<div class="{height} {getColorClasses()} relative" style="width: {animatedWidth}%">
			<!-- Animated shimmer effect -->
			<div
				class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-pulse"
			></div>
		</div>

		<!-- Threshold indicator -->
		{#if showThreshold}
			<div class="absolute top-0 h-full w-0.5 bg-red-500/80" style="left: {threshold}%"></div>
		{/if}
	</div>
</div>
