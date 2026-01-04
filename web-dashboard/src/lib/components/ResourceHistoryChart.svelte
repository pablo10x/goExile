<script lang="ts">
	import { onMount } from 'svelte';
	import type { ResourceHistory } from '$lib/types/resource-metrics';

	let { data = [], height = 200 }: { data?: ResourceHistory[]; height?: number } = $props();

	let width = $state(0);
	let container: HTMLDivElement;

	// Tooltip state
	let hoveredIndex = $state<number | null>(null);
	let tooltipX = $state(0);

	// Optimize: Downsample data if too large
	const MAX_POINTS = 100;

	let chartData = $derived.by(() => {
		if (!data || data.length === 0) return [];
		if (data.length <= MAX_POINTS) return data;

		// Simple downsampling: take every Nth point
		const step = Math.ceil(data.length / MAX_POINTS);
		return data.filter((_, i) => i % step === 0);
	});

	onMount(() => {
		const resizeObserver = new ResizeObserver((entries) => {
			if (entries[0]) {
				width = entries[0].contentRect.width;
			}
		});

		if (container) resizeObserver.observe(container);

		return () => resizeObserver.disconnect();
	});

	function formatTime(ts: string) {
		return new Date(ts).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
	}

	function handleMouseMove(e: MouseEvent) {
		if (!width || chartData.length === 0) return;

		const rect = container.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const index = Math.round((x / width) * (chartData.length - 1));
		hoveredIndex = Math.max(0, Math.min(index, chartData.length - 1));
		tooltipX = (hoveredIndex / (chartData.length - 1)) * width;
	}

	function handleMouseLeave() {
		hoveredIndex = null;
	}

	// Memoize paths using $derived
	let cpuPath = $derived.by(() => {
		if (chartData.length === 0 || width === 0) return '';
		return chartData
			.map((d, i) => {
				const x = (i / (chartData.length - 1)) * width;
				const val = d.cpu;
				const y = height - (val / 100) * (height * 0.8) - height * 0.1;
				return `${x.toFixed(1)},${y.toFixed(1)}`; // Limit precision
			})
			.join(' ');
	});

	let memPath = $derived.by(() => {
		if (chartData.length === 0 || width === 0) return '';
		return chartData
			.map((d, i) => {
				const x = (i / (chartData.length - 1)) * width;
				const val = d.memory_percent;
				const y = height - (val / 100) * (height * 0.8) - height * 0.1;
				return `${x.toFixed(1)},${y.toFixed(1)}`; // Limit precision
			})
			.join(' ');
	});

	let cpuAreaPath = $derived(cpuPath ? `M0,${height} ${cpuPath} L${width},${height} Z` : '');
	let memAreaPath = $derived(memPath ? `M0,${height} ${memPath} L${width},${height} Z` : '');

	function getY(val: number) {
		return height - (val / 100) * (height * 0.8) - height * 0.1;
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
	{#if chartData.length > 0 && width > 0}
		<svg {width} {height} class="overflow-visible" preserveAspectRatio="none">
			<defs>
				<linearGradient id="cpuGradient" x1="0" x2="0" y1="0" y2="1">
					<stop offset="0%" stop-color="#f97316" stop-opacity="0.3" />
					<stop offset="100%" stop-color="#f97316" stop-opacity="0" />
				</linearGradient>
				<linearGradient id="memGradient" x1="0" x2="0" y1="0" y2="1">
					<stop offset="0%" stop-color="#7c2d12" stop-opacity="0.3" />
					<stop offset="100%" stop-color="#7c2d12" stop-opacity="0" />
				</linearGradient>
			</defs>

			<!-- Grid Lines (Horizontal) -->
			{#each [0, 25, 50, 75, 100] as tick}
				<line
					x1="0"
					y1={getY(tick)}
					x2={width}
					y2={getY(tick)}
					stroke="#262626"
					stroke-width="1"
					stroke-dasharray="4 4"
					stroke-opacity="0.5"
				/>
			{/each}

			<!-- Areas -->
			<path d={cpuAreaPath} fill="url(#cpuGradient)" />
			<path d={memAreaPath} fill="url(#memGradient)" />

			<!-- Lines -->
			<path
				d={`M${cpuPath}`}
				fill="none"
				stroke="#f97316"
				stroke-width="2"
				vector-effect="non-scaling-stroke"
				stroke-opacity="0.8"
			/>
			<path
				d={`M${memPath}`}
				fill="none"
				stroke="#7c2d12"
				stroke-width="2"
				vector-effect="non-scaling-stroke"
				stroke-opacity="0.8"
			/>

			<!-- Highlighted Point -->
			{#if hoveredIndex !== null}
				{@const d = chartData[hoveredIndex]}
				{@const yCpu = getY(d.cpu)}
				{@const yMem = getY(d.memory_percent)}

				<!-- Vertical Line -->
				<line
					x1={tooltipX}
					y1={0}
					x2={tooltipX}
					y2={height}
					stroke="var(--color-rust)"
					stroke-opacity="0.3"
					stroke-width="1"
				/>

				<!-- CPU Dot -->
				<circle
					cx={tooltipX}
					cy={yCpu}
					r="4"
					fill="#f97316"
					stroke="white"
					stroke-width="2"
					class="pointer-events-none shadow-lg"
				/>

				<!-- Mem Dot -->
				<circle
					cx={tooltipX}
					cy={yMem}
					r="4"
					fill="#7c2d12"
					stroke="white"
					stroke-width="2"
					class="pointer-events-none shadow-lg"
				/>
			{/if}
		</svg>

		<!-- Tooltip -->
		{#if hoveredIndex !== null}
			{@const d = chartData[hoveredIndex]}
			<div
				class="absolute z-10 pointer-events-none transform -tranneutral-x-1/2 -tranneutral-y-full mb-4 bg-stone-950/90 backdrop-blur-md border border-stone-800 px-4 py-3 shadow-2xl text-center min-w-[120px]"
				style="left: {tooltipX}px; top: 0;"
			>
				<div class="text-[9px] text-stone-500 font-black mb-2 uppercase tracking-widest border-b border-stone-800 pb-1">
					{formatTime(d.timestamp)}
				</div>
				<div class="flex flex-col gap-2">
					<div
						class="text-[10px] font-black text-stone-300 flex items-center justify-between gap-4 uppercase"
					>
						<span class="text-rust">CORE</span>
						<span class="tabular-nums">{d.cpu?.toFixed(1)}%</span>
					</div>
					<div
						class="text-[10px] font-black text-stone-300 flex items-center justify-between gap-4 uppercase"
					>
						<span class="text-rust-dark">MEM</span>
						<span class="tabular-nums">{d.memory_percent?.toFixed(1)}%</span>
					</div>
				</div>
			</div>
		{/if}

		<!-- X-Axis Labels -->
		<div
			class="absolute bottom-0 left-0 right-0 flex justify-between text-[8px] font-black text-stone-600 px-1 uppercase tracking-widest"
		>
			<span>{formatTime(chartData[0].timestamp)}</span>
			{#if chartData.length > 2}
				<span>{formatTime(chartData[Math.floor(chartData.length / 2)].timestamp)}</span>
			{/if}
			<span>{formatTime(chartData[chartData.length - 1].timestamp)}</span>
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center h-full text-stone-700 gap-3 opacity-40">
			<div class="w-8 h-8 border border-dashed border-stone-800 flex items-center justify-center">
				<div class="w-1.5 h-1.5 bg-stone-800 animate-ping"></div>
			</div>
			<span class="text-[10px] font-black uppercase tracking-widest">Awaiting_Temporal_Data</span>
		</div>
	{/if}
</div>