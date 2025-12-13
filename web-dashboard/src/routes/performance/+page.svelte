<script lang="ts">
    import TopResourceConsumers from '$lib/components/TopResourceConsumers.svelte';
    import { Zap, Activity, ShieldCheck } from 'lucide-svelte';
</script>

<!-- Content without extra container - let layout handle the background -->
<div class="space-y-6">
    <div class="p-6">
        <!-- Page Content -->
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
                <div class="px-4 py-2 bg-slate-700/60 backdrop-blur border border-slate-600 rounded-lg flex items-center gap-3">
                    <Activity class="w-5 h-5 text-blue-400" />
                    <div class="text-sm">
                        <span class="text-slate-400">Status:</span>
                        <span class="text-emerald-400 font-semibold ml-1">Optimal</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- CPU Consumers -->
            <div class="transform hover:scale-[1.02] transition-transform duration-300">
                <div class="bg-slate-800/60 backdrop-blur-sm border border-slate-700 rounded-xl overflow-hidden">
                    <TopResourceConsumers limit={10} resourceType="cpu" />
                </div>
            </div>

            <!-- Memory Consumers -->
            <div class="transform hover:scale-[1.02] transition-transform duration-300">
                <div class="bg-slate-800/60 backdrop-blur-sm border border-slate-700 rounded-xl overflow-hidden">
                    <TopResourceConsumers limit={10} resourceType="memory" />
                </div>
            </div>
        </div>
        
        <div class="mt-8">
            <!-- Disk Consumers -->
            <div class="transform hover:scale-[1.02] transition-transform duration-300">
                <div class="bg-slate-800/60 backdrop-blur-sm border border-slate-700 rounded-xl overflow-hidden">
                    <TopResourceConsumers limit={5} resourceType="disk" />
                </div>
            </div>
        </div>
    </div>
</div>
