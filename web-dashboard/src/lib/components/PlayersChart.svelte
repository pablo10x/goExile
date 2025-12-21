<script lang="ts">
	import { onMount } from 'svelte';

	export let data: { timestamp: number; count: number }[] = [];
	export let height = 200;
	export let color = '#10b981'; // emerald-500

	let width = 0;
	let container: HTMLDivElement;

	// Tooltip state
	let hoveredIndex: number | null = null;
	let tooltipX = 0;
	let tooltipY = 0;

	onMount(() => {
		const resizeObserver = new ResizeObserver((entries) => {
			if (entries[0]) {
				width = entries[0].contentRect.width;
			}
		});
		if (container) resizeObserver.observe(container);
		return () => resizeObserver.disconnect();
	});

	$: maxVal = Math.max(...data.map((d) => d.count), 10); // Minimum max of 10 for scale
	$: minVal = 0;

	$: points = data
		.map((d, i) => {
			const x = (i / (data.length - 1)) * width;
			const y = height - ((d.count - minVal) / (maxVal - minVal)) * (height * 0.7) - height * 0.15; // Padding
			return `${x},${y}`;
		})
		.join(' ');

	$: areaPoints = `${points} ${width},${height} 0,${height}`;

	function handleMouseMove(e: MouseEvent) {
		if (!width || data.length === 0) return;

		const rect = container.getBoundingClientRect();
		const x = e.clientX - rect.left;

		// Find nearest point
		const index = Math.round((x / width) * (data.length - 1));
		const safeIndex = Math.max(0, Math.min(index, data.length - 1));

		hoveredIndex = safeIndex;

		// Calculate tooltip position
		tooltipX = (safeIndex / (data.length - 1)) * width;
		// tooltipY depends on value
		const d = data[safeIndex];
		tooltipY = height - ((d.count - minVal) / (maxVal - minVal)) * (height * 0.7) - height * 0.15;
	}

	function handleMouseLeave() {
		hoveredIndex = null;
	}

	function formatTime(ts: number) {
		return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="w-full relative font-sans group"
	bind:this={container}
	style="height: {height}px"
	on:mousemove={handleMouseMove}
	on:mouseleave={handleMouseLeave}
>
	{#if width > 0 && data.length > 1}
		<svg {width} {height} class="overflow-visible">
			<defs>
				<linearGradient id="chartGradient" x1="0" x2="0" y1="0" y2="1">
					<stop offset="0%" stop-color={color} stop-opacity="0.2" />
					<stop offset="100%" stop-color={color} stop-opacity="0" />
				</linearGradient>
			</defs>

			<!-- Grid Lines (Horizontal) -->
			{#each [0, 0.25, 0.5, 0.75, 1] as tick}
				<line
					x1="0"
					y1={height - tick * (height * 0.7) - height * 0.15}
					x2={width}
					y2={height - tick * (height * 0.7) - height * 0.15}
					stroke="#334155"
					stroke-width="1"
					stroke-dasharray="4"
					stroke-opacity="0.5"
				/>
			{/each}

			<!-- Area Fill -->
			<path d="M{data[0] ? `0,${height}` : ''} {areaPoints}" fill="url(#chartGradient)" />

			<!-- Line -->
			<path
				d="M{points}"
				fill="none"
				stroke={color}
				stroke-width="2"
				vector-effect="non-scaling-stroke"
			/>

			<!-- Highlighted Point -->
			{#if hoveredIndex !== null}
				{@const d = data[hoveredIndex]}
				{@const x = (hoveredIndex / (data.length - 1)) * width}
				{@const y =
					height - ((d.count - minVal) / (maxVal - minVal)) * (height * 0.7) - height * 0.15}

				<!-- Vertical Cursor Line -->
				<line
					x1={x}
					y1={0}
					x2={x}
					y2={height}
					stroke="white"
					stroke-opacity="0.1"
					stroke-width="1"
				/>

				<circle
					cx={x}
					cy={y}
					r="4"
					fill={color}
					stroke="white"
					stroke-width="2"
					class="pointer-events-none"
				/>
			{/if}
		</svg>

		<!-- Tooltip -->
		{#if hoveredIndex !== null}
			{@const d = data[hoveredIndex]}
			<div
				class="absolute z-10 pointer-events-none transform -translate-x-1/2 -translate-y-full mb-2 bg-slate-800/90 backdrop-blur border border-slate-600 rounded px-3 py-2 shadow-xl text-center min-w-[80px]"
				style="left: {tooltipX}px; top: {tooltipY}px;"
			>
				<div class="text-xs text-slate-500 dark:text-slate-400 font-mono mb-0.5">{formatTime(d.timestamp)}</div>
				<div class="text-sm font-bold text-slate-900 dark:text-white flex items-center justify-center gap-1">
					<span class="text-emerald-400">●</span>
					{d.count} Players
				</div>
			</div>
		{/if}

		<!-- X-Axis Labels (Simple: First, Middle, Last) -->
		<div
			class="absolute bottom-0 left-0 right-0 flex justify-between text-[10px] text-slate-500 px-1 font-mono"
		>
			<span>{formatTime(data[0].timestamp)}</span>
			<span>{formatTime(data[Math.floor(data.length / 2)].timestamp)}</span>
			<span>{formatTime(data[data.length - 1].timestamp)}</span>
		</div>
	{/if}
</div>
