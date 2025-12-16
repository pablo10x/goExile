<script lang="ts">
    import { onMount } from 'svelte';
    import { spawners } from '$lib/stores';
    import { Server, Activity, Cpu } from 'lucide-svelte';
    import { fade, scale } from 'svelte/transition';

    // Animation state
    let center = { x: 400, y: 300 };
    let radius = 200;
    
    // Track heartbeats for pulse animation
    let lastHeartbeats = $state<Record<number, number>>({});
    let pulsingSpawners = $state<Set<number>>(new Set());

    $effect(() => {
        $spawners.forEach(s => {
            const lastTime = new Date(s.last_seen || 0).getTime();
            if (lastHeartbeats[s.id] && lastTime > lastHeartbeats[s.id]) {
                // New heartbeat detected
                triggerPulse(s.id);
            }
            lastHeartbeats[s.id] = lastTime;
        });
    });

    function triggerPulse(id: number) {
        pulsingSpawners.add(id);
        pulsingSpawners = new Set(pulsingSpawners);
        setTimeout(() => {
            pulsingSpawners.delete(id);
            pulsingSpawners = new Set(pulsingSpawners);
        }, 1000);
    }

    function getPosition(index: number, total: number) {
        if (total === 0) return center;
        const angle = (index / total) * 2 * Math.PI - Math.PI / 2; // Start from top
        return {
            x: center.x + radius * Math.cos(angle),
            y: center.y + radius * Math.sin(angle)
        };
    }
</script>

<div class="relative w-full h-[600px] bg-slate-900/50 rounded-2xl border border-slate-800 overflow-hidden flex items-center justify-center shadow-inner">
    <!-- Grid Background -->
    <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHZpZXdCb3g9IjAgMCA0MCA0MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cGF0aCBkPSJNMCAwaDQwdjQwSDB6IiBmaWxsPSJub25lIi8+CiAgPHBhdGggZD0iTTAgNDBoNDBNNDAgMHY0MCIgc3Ryb2tlPSJyZ2JhKDI1NSwgMjU1LCAyNTUsIDAuMDMpIiBzdHJva2Utd2lkdGg9IjEiLz4KPC9zdmc+')] opacity-20"></div>

    <svg class="w-full h-full pointer-events-none absolute inset-0">
        <defs>
            <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
                <feGaussianBlur stdDeviation="5" result="blur" />
                <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
            <marker id="arrow" markerWidth="10" markerHeight="10" refX="20" refY="3" orient="auto" markerUnits="strokeWidth">
                <path d="M0,0 L0,6 L9,3 z" fill="#475569" />
            </marker>
        </defs>

        <!-- Connections -->
        {#each $spawners as spawner, i}
            {@const pos = getPosition(i, $spawners.length)}
            {@const isActive = spawner.status !== 'offline'}
            
            <!-- Connection Line -->
            <line 
                x1={pos.x} 
                y1={pos.y} 
                x2={center.x} 
                y2={center.y} 
                stroke={isActive ? '#64748b' : '#7f1d1d'} 
                stroke-width={isActive ? 2 : 1}
                stroke-dasharray={isActive ? "0" : "5,5"}
                opacity="0.5"
            />

            <!-- Pulse Packet (Spark) -->
            {#if pulsingSpawners.has(spawner.id) && isActive}
                <g filter="url(#glow)">
                    <!-- Spark circles -->
                    <circle r="3" fill="#10b981">
                        <animateMotion 
                            dur="1s" 
                            repeatCount="1"
                            path={`M${pos.x},${pos.y} L${center.x},${center.y}`}
                            fill="freeze"
                            keyPoints="0;1"
                            keyTimes="0;1"
                            calcMode="linear"
                        />
                    </circle>
                    <!-- Trail Effect -->
                     <circle r="1.5" fill="#059669" opacity="0.6">
                        <animateMotion 
                            dur="1s" 
                            repeatCount="1"
                            path={`M${pos.x},${pos.y} L${center.x},${center.y}`}
                            fill="freeze"
                            begin="0.05s"
                            keyPoints="0;1"
                            keyTimes="0;1"
                            calcMode="linear"
                        />
                    </circle>
                    <circle r="1" fill="#059669" opacity="0.3">
                        <animateMotion 
                            dur="1s" 
                            repeatCount="1"
                            path={`M${pos.x},${pos.y} L${center.x},${center.y}`}
                            fill="freeze"
                            begin="0.1s"
                            keyPoints="0;1"
                            keyTimes="0;1"
                            calcMode="linear"
                        />
                    </circle>
                </g>
            {/if}

            <!-- Sync Animation (Continuous Flow) -->
            {#if isActive}
                <circle r="2" fill="white" opacity="0.7">
                    <animateMotion 
                        dur="3s" 
                        repeatCount="indefinite"
                        path={`M${pos.x},${pos.y} L${center.x},${center.y} L${pos.x},${pos.y}`}
                        rotate="auto"
                    />
                </circle>
            {/if}
        {/each}
    </svg>

    <!-- Master Node (Center) -->
    <div 
        class="absolute z-20 flex flex-col items-center justify-center w-32 h-32 rounded-full bg-slate-900 border-4 border-blue-500 shadow-[0_0_50px_rgba(59,130,246,0.3)] animate-pulse-slow"
        style="top: {center.y - 64}px; left: {center.x - 64}px;"
    >
        <Server class="w-10 h-10 text-blue-400 mb-1" />
        <span class="text-xs font-bold text-blue-200">MASTER</span>
        <span class="text-[10px] text-blue-400/60 font-mono mt-1">ONLINE</span>
        
        <!-- Orbital Rings around Master -->
        <div class="absolute inset-0 rounded-full border border-blue-500/20 animate-ping-slow"></div>
    </div>

    <!-- Spawner Nodes -->
    {#each $spawners as spawner, i}
        {@const pos = getPosition(i, $spawners.length)}
        {@const isActive = spawner.status !== 'offline'}
        
        <div 
            class="absolute z-10 flex flex-col items-center group cursor-pointer transition-all duration-300 hover:scale-110"
            style="top: {pos.y - 24}px; left: {pos.x - 24}px;"
            transition:scale
        >
            <div class={`
                relative w-12 h-12 rounded-full flex items-center justify-center border-2 shadow-lg backdrop-blur-md
                ${isActive ? 'bg-slate-800/80 border-slate-500 shadow-slate-500/20' : 'bg-red-900/20 border-red-800 shadow-red-900/20'}
                ${pulsingSpawners.has(spawner.id) ? 'scale-110 border-emerald-400 shadow-emerald-500/50' : ''}
                transition-all duration-300
            `}>
                <Cpu class={`w-6 h-6 ${isActive ? 'text-slate-300' : 'text-red-500'}`} />
                
                {#if isActive}
                    <div class="absolute -bottom-1 -right-1 w-3 h-3 bg-emerald-500 rounded-full border-2 border-slate-900 animate-pulse"></div>
                {/if}
            </div>
            
            <div class="absolute top-14 flex flex-col items-center bg-slate-900/90 px-3 py-1 rounded-lg border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-30 pointer-events-none">
                <span class="text-xs font-bold text-white">Spawner #{spawner.id}</span>
                <span class="text-[10px] text-slate-400">{spawner.region}</span>
                <span class={`text-[10px] font-mono ${isActive ? 'text-emerald-400' : 'text-red-400'}`}>{spawner.status}</span>
            </div>
        </div>
    {/each}
</div>

<style>
    @keyframes pulse-slow {
        0%, 100% { box-shadow: 0 0 30px rgba(59, 130, 246, 0.2); }
        50% { box-shadow: 0 0 60px rgba(59, 130, 246, 0.4); }
    }
    .animate-pulse-slow {
        animation: pulse-slow 4s infinite;
    }
    .animate-ping-slow {
        animation: ping 3s cubic-bezier(0, 0, 0.2, 1) infinite;
    }
</style>