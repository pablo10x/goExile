<script lang="ts">
    import { onMount } from 'svelte';

    export let data: { timestamp: string; cpu: number; memory_percent: number }[] = [];
    export let height = 200;

    let width = 0;
    let container: HTMLDivElement;
    
    // Tooltip state
    let hoveredIndex: number | null = null;
    let tooltipX = 0;

    onMount(() => {
        const resizeObserver = new ResizeObserver(entries => {
            if (entries[0]) {
                width = entries[0].contentRect.width;
            }
        });
        if (container) resizeObserver.observe(container);
        return () => resizeObserver.disconnect();
    });

    // Y-Axis is 0-100%
    const maxVal = 100;
    const minVal = 0;

    function getY(val: number) {
        return height - ((Math.min(100, Math.max(0, val)) - minVal) / (maxVal - minVal)) * (height * 0.7) - (height * 0.15);
    }
    
    $: cpuPoints = data.length > 1 
        ? data.map((d, i) => `${(i / (data.length - 1)) * width},${getY(d.cpu)}`).join(' ')
        : (data.length === 1 ? `0,${getY(data[0].cpu)} ${width},${getY(data[0].cpu)}` : '');

    $: memPoints = data.length > 1 
        ? data.map((d, i) => `${(i / (data.length - 1)) * width},${getY(d.memory_percent)}`).join(' ')
        : (data.length === 1 ? `0,${getY(data[0].memory_percent)} ${width},${getY(data[0].memory_percent)}` : '');

    function handleMouseMove(e: MouseEvent) {
        if (!width || data.length === 0) return;
        const rect = container.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const index = data.length > 1 ? Math.round((x / width) * (data.length - 1)) : 0;
        hoveredIndex = Math.max(0, Math.min(index, data.length - 1));
        tooltipX = data.length > 1 ? (hoveredIndex / (data.length - 1)) * width : width / 2;
    }

    function handleMouseLeave() { hoveredIndex = null; }

    function formatTime(ts: string) {
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
    {#if width > 0 && data.length > 0}
        <svg {width} {height} class="overflow-visible">
            <!-- Grid Lines -->
            {#each [0, 0.25, 0.5, 0.75, 1] as tick}
                <line x1="0" y1={height - (tick * (height * 0.7)) - (height * 0.15)} x2={width} y2={height - (tick * (height * 0.7)) - (height * 0.15)} stroke="#334155" stroke-width="1" stroke-dasharray="4" stroke-opacity="0.5" />
            {/each}

            <!-- CPU Line (Red) -->
            <path d="M{cpuPoints}" fill="none" stroke="#ef4444" stroke-width="2" vector-effect="non-scaling-stroke" />
            
            <!-- Memory Line (Blue) -->
            <path d="M{memPoints}" fill="none" stroke="#3b82f6" stroke-width="2" vector-effect="non-scaling-stroke" />

            <!-- Highlight -->
            {#if hoveredIndex !== null}
                {@const d = data[hoveredIndex]}
                {@const x = data.length > 1 ? (hoveredIndex / (data.length - 1)) * width : width/2}
                
                <line x1={x} y1={0} x2={x} y2={height} stroke="white" stroke-opacity="0.1" stroke-width="1" />
                
                <!-- Dots -->
                <circle cx={x} cy={getY(d.cpu)} r="4" fill="#ef4444" stroke="white" stroke-width="2" />
                <circle cx={x} cy={getY(d.memory_percent)} r="4" fill="#3b82f6" stroke="white" stroke-width="2" />
            {/if}
        </svg>
        
        <!-- Tooltip -->
        {#if hoveredIndex !== null}
            {@const d = data[hoveredIndex]}
            <div class="absolute z-10 pointer-events-none transform -translate-x-1/2 mb-2 bg-slate-800/90 backdrop-blur border border-slate-600 rounded px-3 py-2 shadow-xl text-center min-w-[120px]" style="left: {tooltipX}px; top: 10px;">
                <div class="text-xs text-slate-400 font-mono mb-1">{formatTime(d.timestamp)}</div>
                <div class="flex gap-3 justify-center text-xs font-bold">
                    <div class="text-red-400">CPU: {d.cpu.toFixed(1)}%</div>
                    <div class="text-blue-400">MEM: {d.memory_percent.toFixed(1)}%</div>
                </div>
            </div>
        {/if}

        <!-- X-Axis Labels -->
        <div class="absolute bottom-0 left-0 right-0 flex justify-between text-[10px] text-slate-500 px-1 font-mono">
            <span>{formatTime(data[0].timestamp)}</span>
            {#if data.length > 2}
                <span>{formatTime(data[Math.floor(data.length/2)].timestamp)}</span>
            {/if}
            {#if data.length > 1}
                <span>{formatTime(data[data.length-1].timestamp)}</span>
            {/if}
        </div>
    {/if}
</div>