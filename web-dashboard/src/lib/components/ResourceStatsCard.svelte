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
    let displayPeak = $state(0);
    
    // Animate number changes
    $effect(() => {
        if (animated) {
            const targetCurrent = current;
            if (targetCurrent !== displayCurrent) {
                animateValue(displayCurrent, targetCurrent, (val) => displayCurrent = val);
            }
            
            const targetPeak = peak;
            if (targetPeak !== displayPeak) {
                animateValue(displayPeak, targetPeak, (val) => displayPeak = val);
            }
        } else {
            displayCurrent = current;
            displayPeak = peak;
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
    
    function getColorClasses() {
        const baseClasses = 'relative overflow-hidden';
        const colorMap = {
            orange: 'from-orange-500 to-red-500',
            blue: 'from-blue-500 to-purple-500',
            green: 'from-green-500 to-teal-500',
            red: 'from-red-500 to-pink-500',
            purple: 'from-purple-500 to-indigo-500',
            teal: 'from-teal-500 to-cyan-500'
        };
        
        return `${baseClasses} bg-gradient-to-br ${colorMap[color] || colorMap.blue}`;
    }
    
    function getTrendIcon() {
        if (!trend) return '';
        
        const trendMap = {
            up: '↗️',
            down: '↘️',
            stable: '→'
        };
        
        return trendMap[trend];
    }
    
    function getTrendColor() {
        if (!trend) return 'text-slate-400';
        
        const colorMap = {
            up: 'text-red-400',
            down: 'text-green-400',
            stable: 'text-slate-400'
        };
        
        return colorMap[trend];
    }
    
    function getProgressPercentage() {
        return Math.min((current / peak) * 100, 100);
    }
</script>

<div class="group relative bg-slate-800/50 backdrop-blur-md border border-slate-700/50 rounded-xl p-6 hover:border-slate-600/50 transition-all duration-300 hover:shadow-lg hover:shadow-slate-900/20">
    <!-- Background Gradient -->
    <div class="absolute inset-0 bg-gradient-to-br from-slate-900/20 to-slate-800/10 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
    
    <!-- Content -->
    <div class="relative z-10">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
                <div class="p-2 bg-slate-900/50 rounded-lg border border-slate-700/50">
                    <div class="text-xl">{icon}</div>
                </div>
                <h3 class="text-slate-200 font-semibold tracking-wide">{title}</h3>
            </div>
            {#if trend}
                <div class="flex items-center gap-1 text-sm {getTrendColor()}">
                    <span class="text-lg">{getTrendIcon()}</span>
                    <span class="font-medium">{Math.abs(Math.round((current - peak) / peak * 100))}%</span>
                </div>
            {/if}
        </div>
        
        <!-- Current Value -->
        <div class="mb-3">
            <div class="text-3xl font-bold text-slate-100 tabular-nums">
                {displayCurrent.toFixed(1)}{unit}
            </div>
        </div>
        
        <!-- Progress Bar -->
        <div class="mb-3">
            <div class="h-2 bg-slate-700/50 rounded-full overflow-hidden">
                <div 
                    class="h-full {getColorClasses()} rounded-full transition-all duration-500 ease-out"
                    style="width: {getProgressPercentage()}%"
                ></div>
            </div>
        </div>
        
        <!-- Peak Value -->
        <div class="flex items-center justify-between text-sm text-slate-400">
            <span>Peak: {displayPeak.toFixed(1)}{unit}</span>
            <span>{getProgressPercentage().toFixed(0)}% of peak</span>
        </div>
    </div>
    
    <!-- Hover Effect Border -->
    <div class="absolute inset-0 rounded-xl border border-transparent group-hover:border-slate-600/30 transition-all duration-300 pointer-events-none"></div>
</div>