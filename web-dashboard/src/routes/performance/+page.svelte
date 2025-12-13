<script lang="ts">
    import TopResourceConsumers from '$lib/components/TopResourceConsumers.svelte';
    import { Zap, Activity, ShieldCheck } from 'lucide-svelte';
</script>

<div class="relative min-h-[calc(100vh-4rem)] overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/50 shadow-2xl">
    <!-- Custom Background: Speed & Reliability Theme -->
    <div class="absolute inset-0 z-0">
        <!-- Base Dark Layer -->
        <div class="absolute inset-0 bg-slate-950"></div>
        
        <!-- Speed Lines / Grid -->
        <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] bg-[size:50px_50px] [mask-image:radial-gradient(ellipse_at_center,black_40%,transparent_80%)]"></div>
        
        <!-- Moving glowing orbs for "Speed" -->
        <div class="absolute top-0 left-1/4 w-96 h-96 bg-cyan-500/20 rounded-full blur-[100px] animate-pulse"></div>
        <div class="absolute bottom-0 right-1/4 w-96 h-96 bg-blue-600/20 rounded-full blur-[100px] animate-pulse animation-delay-2000"></div>
        
        <!-- Reliability Shield Hexagon Pattern overlay -->
        <div class="absolute inset-0 opacity-[0.15]" 
             style="background-image: url('data:image/svg+xml,%3Csvg width=\'24\' height=\'40\' viewBox=\'0 0 24 40\' xmlns=\'http://www.w3.org/2000/svg\'%3E%3Cpath d=\'M0 40c5.523 0 10-4.477 10-10V10c0-5.523-4.477-10-10-10s-10 4.477-10 10v20c0 5.523 4.477 10 10 10z\' fill=\'%234f46e5\' fill-opacity=\'0.4\' fill-rule=\'evenodd\'%2F%3E%3C/svg%3E');">
        </div>
    </div>

    <!-- Content -->
    <div class="relative z-10 p-8">
        <div class="flex items-center justify-between mb-8">
            <div>
                <h1 class="text-3xl font-bold text-white flex items-center gap-3">
                    <Zap class="w-8 h-8 text-yellow-400" />
                    System Performance
                </h1>
                <p class="text-slate-400 mt-2 flex items-center gap-2">
                    <ShieldCheck class="w-4 h-4 text-emerald-400" />
                    Monitoring system speed, resource reliability, and instance health.
                </p>
            </div>
            
            <div class="flex gap-4">
                <div class="px-4 py-2 bg-slate-800/80 backdrop-blur border border-slate-700 rounded-lg flex items-center gap-3">
                    <Activity class="w-5 h-5 text-blue-400" />
                    <div class="text-sm">
                        <span class="text-slate-400">Status:</span>
                        <span class="text-emerald-400 font-semibold ml-1">Optimal</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- CPU Consumers -->
            <div class="transform hover:scale-[1.01] transition-transform duration-300">
                <TopResourceConsumers limit={10} resourceType="cpu" />
            </div>

            <!-- Memory Consumers -->
            <div class="transform hover:scale-[1.01] transition-transform duration-300">
                <TopResourceConsumers limit={10} resourceType="memory" />
            </div>
        </div>
        
        <div class="mt-8 grid grid-cols-1">
             <!-- Disk Consumers -->
             <div class="transform hover:scale-[1.01] transition-transform duration-300">
                <TopResourceConsumers limit={5} resourceType="disk" />
            </div>
        </div>
    </div>
</div>
