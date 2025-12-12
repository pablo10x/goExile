<script lang="ts">
    import { onMount } from 'svelte';
    import { config } from '$lib/stores';
    import StatsCard from '$lib/components/StatsCard.svelte';
    
    let loading = $state(true);
    let error = $state<string | null>(null);
    
    // Configuration categories
    const categories = [
        { 
            name: 'system', 
            label: 'System', 
            description: 'Core server settings',
            icon: 'M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z',
            color: 'blue'
        },
        { 
            name: 'spawner', 
            label: 'Spawners', 
            description: 'Spawner management defaults',
            icon: 'M19 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.11 0 2-.9 2-2V5c0-1.1-.89-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z',
            color: 'green'
        },
        { 
            name: 'security', 
            label: 'Security', 
            description: 'Authentication and access control',
            icon: 'M12,1L3,5V11C3,16.55 6.84,21.74 12,23C17.16,21.74 21,16.55 21,11V5L12,1M12,7C13.4,7 14.8,8.6 14.8,10V11.5C15.4,11.5 16,12.1 16,12.7V16.2C16,16.8 15.4,17.4 14.8,17.4H9.2C8.6,17.4 8,16.8 8,16.2V12.7C8,12.1 8.6,11.5 9.2,11.5V10C9.2,8.6 10.6,7 12,7M12,8.2C11.2,8.2 10.5,8.7 10.5,10V11.5H13.5V10C13.5,8.7 12.8,8.2 12,8.2Z',
            color: 'red'
        }
    ];
    
    async function loadConfig() {
        try {
            loading = true;
            error = null;
            
            const response = await fetch('/api/config');
            if (!response.ok) {
                throw new Error(`Failed to load configuration: ${response.statusText}`);
            }
            
            const configData = await response.json();
            config.set(configData);
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load configuration';
            console.error('Failed to load configuration:', e);
        } finally {
            loading = false;
        }
    }
    
    function getConfigCount(category: string) {
        if (!$config) return 0;
        return $config.filter((c: any) => c.category === category).length;
    }
    
    function getCategoryColor(category: string) {
        const cat = categories.find(c => c.name === category);
        return cat ? cat.color : 'gray';
    }
    
    onMount(() => {
        loadConfig();
    });
</script>

<div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold text-slate-100 mb-2">Configuration</h1>
            <p class="text-slate-400">Manage server settings, spawner defaults, and security options</p>
        </div>
        <button 
            onclick={loadConfig}
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors duration-200 flex items-center gap-2"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M23 4v6h-6"></path>
                <path d="M1 20v-6h6"></path>
                <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            Refresh
        </button>
    </div>

    <!-- Error State -->
    {#if error}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-300">
            <div class="flex items-center gap-3">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <span>{error}</span>
            </div>
        </div>
    {/if}

    <!-- Loading State -->
    {#if loading}
        <div class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
    {:else}
        <!-- Configuration Categories Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {#each categories as category}
                <a 
                    href="/config/{category.name}"
                    class="group block"
                >
                    <div class="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6 hover:border-{category.color}-500/50 transition-all duration-300 group-hover:shadow-lg group-hover:shadow-{category.color}-900/20">
                        <div class="flex items-start justify-between mb-4">
                            <div class="p-2 bg-{category.color}-500/10 rounded-lg">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-{category.color}-400" viewBox="0 0 24 24" fill="currentColor">
                                    <path d={category.icon} />
                                </svg>
                            </div>
                            <div class="text-2xl font-bold text-slate-100">
                                {getConfigCount(category.name)}
                            </div>
                        </div>
                        <h3 class="text-lg font-semibold text-slate-100 mb-1">{category.label}</h3>
                        <p class="text-sm text-slate-400">{category.description}</p>
                    </div>
                </a>
            {/each}
        </div>

        <!-- Recent Configuration Changes -->
        {#if $config && $config.length > 0}
            <div class="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
                <h2 class="text-xl font-semibold text-slate-100 mb-4">All Configuration Settings</h2>
                <div class="space-y-3">
                    {#each $config as configItem (configItem.key)}
                        <div class="flex items-center justify-between p-3 bg-slate-900/50 rounded-lg border border-slate-700">
                            <div class="flex items-center gap-3">
                                <div class="w-2 h-2 rounded-full bg-{getCategoryColor(configItem.category)}-500"></div>
                                <div>
                                    <div class="font-medium text-slate-200">{configItem.key}</div>
                                    <div class="text-sm text-slate-400">{configItem.description}</div>
                                </div>
                            </div>
                            <div class="flex items-center gap-3">
                                <span class="text-slate-300 font-mono text-sm">{configItem.value}</span>
                                {#if configItem.is_read_only}
                                    <span class="px-2 py-1 bg-slate-700 text-slate-300 text-xs rounded">Read-only</span>
                                {/if}
                                {#if configItem.requires_restart}
                                    <span class="px-2 py-1 bg-orange-500/20 text-orange-300 text-xs rounded">Restart Required</span>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    {/if}
</div>