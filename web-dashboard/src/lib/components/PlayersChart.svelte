<script lang="ts">
	import { onMount } from 'svelte';

	let { data = [], height = 200, color = 'var(--color-rust)' }: { data?: { timestamp: number; count: number }[]; height?: number; color?: string } = $props();

	let width = $state(0);
	let container: HTMLDivElement;

	// Tooltip state
	let hoveredIndex = $state<number | null>(null);
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	onMount(() => {
		const resizeObserver = new ResizeObserver((entries) => {
			if (entries[0]) {
				width = entries[0].contentRect.width;
			}
		});
		if (container) resizeObserver.observe(container);
		return () => resizeObserver.disconnect();
	});

	let maxVal = $derived(Math.max(...data.map((d) => d.count), 10)); // Minimum max of 10 for scale
	let minVal = 0;

	let points = $derived(data
		.map((d, i) => {
			const x = (i / (data.length - 1)) * width;
			const y = height - ((d.count - minVal) / (maxVal - minVal)) * (height * 0.7) - height * 0.15; // Padding
			return `${x},${y}`;
		})
		.join(' '));

	let areaPoints = $derived(`${points} ${width},${height} 0,${height}`);

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
		return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="w-full relative font-jetbrains group"
	bind:this={container}
	style="height: {height}px"
	onmousemove={handleMouseMove}
	onmouseleave={handleMouseLeave}
>
	{#if width > 0 && data.length > 1}
		<svg {width} {height} class="overflow-visible" preserveAspectRatio="none">
			<defs>
				<linearGradient id="playerChartGradient" x1="0" x2="0" y1="0" y2="1">
					<stop offset="0%" stop-color={color} stop-opacity="0.3" />
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
					stroke="#262626"
					stroke-width="1"
					stroke-dasharray="4 4"
					stroke-opacity="0.5"
				/>
			{/each}

			<!-- Area Fill -->
			<path d="M{data[0] ? `0,${height}` : ''} {areaPoints}" fill="url(#playerChartGradient)" />

			<!-- Line -->
			<path
				d="M{points}"
				fill="none"
				stroke={color}
				stroke-width="2"
				vector-effect="non-scaling-stroke"
				stroke-opacity="0.8"
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
					stroke="var(--color-rust)"
					stroke-opacity="0.3"
					stroke-width="1"
				/>

				<circle
					cx={x}
					cy={y}
					r="4"
					fill={color}
					stroke="white"
					stroke-width="2"
					class="pointer-events-none shadow-lg shadow-current"
				/>
			{/if}
		</svg>

		<!-- Tooltip -->
		{#if hoveredIndex !== null}
			{@const d = data[hoveredIndex]}
			<div
				class="absolute z-10 pointer-events-none transform -tranneutral-x-1/2 -tranneutral-y-full mb-4 bg-stone-950/90 backdrop-blur-md border border-stone-800 px-4 py-3 shadow-2xl text-center min-w-[120px]"
				style="left: {tooltipX}px; top: 0;"
			>
				<div class="text-[9px] text-stone-500 font-black mb-2 uppercase tracking-widest border-b border-stone-800 pb-1">
					{formatTime(d.timestamp)}
				</div>
				<div
					class="text-[10px] font-black text-stone-300 flex items-center justify-between gap-4 uppercase"
				>
					<span class="text-rust">POPULATION</span>
					<span class="tabular-nums text-white">{d.count} UNITS</span>
				</div>
			</div>
		{/if}

		<!-- X-Axis Labels (Simple: First, Middle, Last) -->
		<div
			class="absolute bottom-0 left-0 right-0 flex justify-between text-[8px] font-black text-stone-600 px-1 uppercase tracking-widest"
		>
			<span>{formatTime(data[0].timestamp)}</span>
			<span>{formatTime(data[Math.floor(data.length / 2)].timestamp)}</span>
			<span>{formatTime(data[data.length - 1].timestamp)}</span>
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center h-full text-stone-700 gap-3 opacity-40">
			<div class="w-8 h-8 border border-dashed border-stone-800 flex items-center justify-center">
				<div class="w-1.5 h-1.5 bg-stone-800 animate-ping"></div>
			</div>
			<span class="text-[10px] font-black uppercase tracking-widest">Signal_Lost // Empty_Buffer</span>
		</div>
	{/if}
</div>