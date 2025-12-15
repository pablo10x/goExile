<script lang="ts">
    import { onMount } from 'svelte';
    import type { ResourceHistory } from '$lib/types/resource-metrics';
    
    let { data = [], height = 200 }: { data?: ResourceHistory[], height?: number } = $props();
    
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
        const resizeObserver = new ResizeObserver(entries => {
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
            minute: '2-digit' 
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
        return chartData.map((d, i) => {
            const x = (i / (chartData.length - 1)) * width;
            const val = d.cpu;
            const y = height - (val / 100) * (height * 0.8) - (height * 0.1); 
            return `${x.toFixed(1)},${y.toFixed(1)}`; // Limit precision
        }).join(' ');
    });

    let memPath = $derived.by(() => {
        if (chartData.length === 0 || width === 0) return '';
        return chartData.map((d, i) => {
            const x = (i / (chartData.length - 1)) * width;
            const val = d.memory_percent;
            const y = height - (val / 100) * (height * 0.8) - (height * 0.1);
            return `${x.toFixed(1)},${y.toFixed(1)}`; // Limit precision
        }).join(' ');
    });

    let cpuAreaPath = $derived(cpuPath ? `M0,${height} ${cpuPath} L${width},${height} Z` : '');
    let memAreaPath = $derived(memPath ? `M0,${height} ${memPath} L${width},${height} Z` : '');

    function getY(val: number) {
        return height - (val / 100) * (height * 0.8) - (height * 0.1);
    }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div 
    class="w-full relative font-sans group" 
    bind:this={container}
    style="height: {height}px"
    onmousemove={handleMouseMove}
    onmouseleave={handleMouseLeave}
>
    {#if chartData.length > 0 && width > 0}
        <svg {width} {height} class="overflow-visible" preserveAspectRatio="none">
            <defs>
                <linearGradient id="cpuGradient" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stop-color="#f97316" stop-opacity="0.2" />
                    <stop offset="100%" stop-color="#f97316" stop-opacity="0" />
                </linearGradient>
                <linearGradient id="memGradient" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.2" />
                    <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
                </linearGradient>
            </defs>

            <!-- Grid Lines (Horizontal) -->
            {#each [0, 25, 50, 75, 100] as tick}
                <line 
                    x1="0" 
                    y1={getY(tick)} 
                    x2={width} 
                    y2={getY(tick)} 
                    stroke="#334155" 
                    stroke-width="1" 
                    stroke-dasharray="4"
                    stroke-opacity="0.3"
                />
            {/each}

            <!-- Areas -->
            <path d={cpuAreaPath} fill="url(#cpuGradient)" />
            <path d={memAreaPath} fill="url(#memGradient)" />

            <!-- Lines -->
            <path d={`M${cpuPath}`} fill="none" stroke="#f97316" stroke-width="2" vector-effect="non-scaling-stroke" />
            <path d={`M${memPath}`} fill="none" stroke="#3b82f6" stroke-width="2" vector-effect="non-scaling-stroke" />

            <!-- Highlighted Point -->
            {#if hoveredIndex !== null}
                {@const d = chartData[hoveredIndex]}
                {@const yCpu = getY(d.cpu)}
                {@const yMem = getY(d.memory_percent)}
                
                <!-- Vertical Line -->
                <line x1={tooltipX} y1={0} x2={tooltipX} y2={height} stroke="white" stroke-opacity="0.1" stroke-width="1" />
                
                <!-- CPU Dot -->
                <circle cx={tooltipX} cy={yCpu} r="4" fill="#f97316" stroke="white" stroke-width="2" class="pointer-events-none" />
                
                <!-- Mem Dot -->
                <circle cx={tooltipX} cy={yMem} r="4" fill="#3b82f6" stroke="white" stroke-width="2" class="pointer-events-none" />
            {/if}
        </svg>

        <!-- Tooltip -->
        {#if hoveredIndex !== null}
            {@const d = chartData[hoveredIndex]}
            <div 
                class="absolute z-10 pointer-events-none transform -translate-x-1/2 -translate-y-full mb-2 bg-slate-800/90 backdrop-blur border border-slate-600 rounded px-3 py-2 shadow-xl text-center min-w-[100px]"
                style="left: {tooltipX}px; top: 0;"
            >
                <div class="text-xs text-slate-400 font-mono mb-1">{formatTime(d.timestamp)}</div>
                <div class="flex flex-col gap-1">
                    <div class="text-xs font-bold text-white flex items-center justify-between gap-3">
                        <span class="text-orange-400">CPU</span>
                        <span>{d.cpu.toFixed(1)}%</span>
                    </div>
                    <div class="text-xs font-bold text-white flex items-center justify-between gap-3">
                        <span class="text-blue-400">MEM</span>
                        <span>{d.memory_percent.toFixed(1)}%</span>
                    </div>
                </div>
            </div>
        {/if}

        <!-- X-Axis Labels -->
        <div class="absolute bottom-0 left-0 right-0 flex justify-between text-[10px] text-slate-500 px-1 font-mono">
            <span>{formatTime(chartData[0].timestamp)}</span>
            {#if chartData.length > 2}
                <span>{formatTime(chartData[Math.floor(chartData.length/2)].timestamp)}</span>
            {/if}
            <span>{formatTime(chartData[chartData.length-1].timestamp)}</span>
        </div>
    {:else}
        <div class="flex items-center justify-center h-full text-slate-500 text-sm">
            Waiting for data...
        </div>
    {/if}
</div>
