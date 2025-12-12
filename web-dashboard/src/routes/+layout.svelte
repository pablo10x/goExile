<script lang="ts">
    import '../app.css';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/state';
    import { isAuthenticated } from '$lib/stores';

    let { children } = $props();
    let isChecking = $state(true);

    async function checkAuth() {
        try {
            const res = await fetch('/api/stats');
            if (res.ok) {
                isAuthenticated.set(true);
            } else {
                isAuthenticated.set(false);
                if (window.location.pathname !== '/login') {
                    goto('/login');
                }
            }
        } catch (e) {
            isAuthenticated.set(false);
            if (window.location.pathname !== '/login') {
                goto('/login');
            }
        } finally {
            isChecking = false;
        }
    }

    onMount(() => {
        checkAuth();
    });

    async function logout() {
        await fetch('/logout');
        isAuthenticated.set(false);
        goto('/login');
    }

    function isRouteActive(path: string) {
        return page.url.pathname === path;
    }
</script>

{#if isChecking}
    <div class="flex items-center justify-center min-h-screen bg-slate-900 text-slate-300">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
{:else}
    {#if $isAuthenticated && page.url.pathname !== '/login'}
        <div class="flex h-screen bg-slate-950 text-slate-300 overflow-hidden">
            <!-- Sidebar with 3D Background -->
            <aside class="relative w-64 bg-slate-950 border-r border-slate-800 flex flex-col shrink-0 overflow-hidden shadow-2xl z-20">
                
                <!-- Animated 3D Background Layers -->
                <div class="absolute inset-0 z-0">
                    <!-- Deep Space Base -->
                    <div class="absolute inset-0 bg-slate-950"></div>
                    
                    <!-- Moving Gradient Orbs -->
                    <div class="absolute top-[-20%] left-[-20%] w-[80%] h-[50%] rounded-full bg-blue-600/30 blur-[80px] animate-blob mix-blend-screen"></div>
                    <div class="absolute bottom-[-10%] right-[-20%] w-[80%] h-[50%] rounded-full bg-purple-600/30 blur-[80px] animate-blob animation-delay-4000 mix-blend-screen"></div>
                    
                    <!-- Grid Texture Overlay -->
                    <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAiIGhlaWdodD0iMjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMSIgY3k9IjEiIHI9IjEiIGZpbGw9InJnYmEoMjU1LDI1NSwyNTUsMC4wNSkiLz48L3N2Zz4=')] opacity-20"></div>
                    
                    <!-- Glass Surface -->
                    <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-[1px] border-r border-white/5"></div>
                </div>

                <!-- Content Container -->
                <div class="relative z-10 flex flex-col h-full">
                    <div class="p-6 border-b border-white/5 bg-white/5 backdrop-blur-md">
                        <h1 class="text-2xl font-bold text-slate-50 -tracking-wide bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent filter drop-shadow-lg">GoExile</h1>
                        <p class="text-[10px] text-blue-200/60 font-mono mt-1 tracking-widest">REGISTRY CONTROL</p>
                    </div>

                    <nav class="flex-1 p-4 space-y-2 overflow-y-auto">
                        <a href="/" 
                            class="flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group relative overflow-hidden
                            {isRouteActive('/') ? 'bg-blue-600/20 text-blue-300 shadow-lg shadow-blue-900/20 border border-blue-500/30' : 'hover:bg-white/5 text-slate-400 hover:text-slate-200 hover:shadow-md border border-transparent'}"
                        >
                            {#if isRouteActive('/')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-blue-400 rounded-r shadow-[0_0_10px_rgba(96,165,250,0.8)]"></div>
                            {/if}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform group-hover:scale-110 group-hover:rotate-3 duration-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <rect x="3" y="3" width="7" height="7"></rect>
                                <rect x="14" y="3" width="7" height="7"></rect>
                                <rect x="14" y="14" width="7" height="7"></rect>
                                <rect x="3" y="14" width="7" height="7"></rect>
                            </svg>
                            <span class="font-medium tracking-wide">Dashboard</span>
                        </a>

                        <a href="/config" 
                            class="flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group relative overflow-hidden
                            {isRouteActive('/config') || isRouteActive('/config/') ? 'bg-blue-600/20 text-blue-300 shadow-lg shadow-blue-900/20 border border-blue-500/30' : 'hover:bg-white/5 text-slate-400 hover:text-slate-200 hover:shadow-md border border-transparent'}"
                        >
                            {#if isRouteActive('/config') || isRouteActive('/config/')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-blue-400 rounded-r shadow-[0_0_10px_rgba(96,165,250,0.8)]"></div>
                            {/if}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform group-hover:scale-110 group-hover:rotate-3 duration-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <circle cx="12" cy="12" r="3"></circle>
                                <path d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"></path>
                            </svg>
                            <span class="font-medium tracking-wide">Configuration</span>
                        </a>

                        <a href="/server" 
                            class="flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group relative overflow-hidden
                            {isRouteActive('/server') ? 'bg-blue-600/20 text-blue-300 shadow-lg shadow-blue-900/20 border border-blue-500/30' : 'hover:bg-white/5 text-slate-400 hover:text-slate-200 hover:shadow-md border border-transparent'}"
                        >
                            {#if isRouteActive('/server')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-blue-400 rounded-r shadow-[0_0_10px_rgba(96,165,250,0.8)]"></div>
                            {/if}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform group-hover:scale-110 group-hover:-rotate-3 duration-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                <polyline points="7 10 12 15 17 10"></polyline>
                                <line x1="12" y1="15" x2="12" y2="3"></line>
                            </svg>
                            <span class="font-medium tracking-wide">Server Files</span>
                        </a>
                    </nav>

                    <div class="p-4 border-t border-white/5 bg-black/20 backdrop-blur-md">
                        <button 
                            onclick={logout} 
                            class="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg text-slate-400 hover:text-white hover:bg-red-500/20 hover:border-red-500/30 border border-transparent transition-all duration-300 group relative overflow-hidden"
                        >
                            <div class="absolute inset-0 bg-gradient-to-r from-red-600/0 via-red-600/10 to-red-600/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 transition-transform group-hover:-translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                                <polyline points="16 17 21 12 16 7"></polyline>
                                <line x1="21" y1="12" x2="9" y2="12"></line>
                            </svg>
                            <span class="text-sm font-medium">Logout</span>
                        </button>
                    </div>
                </div>
            </aside>

            <!-- Main Content -->
            <main class="flex-1 overflow-auto bg-gradient-to-br from-slate-950 to-slate-900 relative">
                 <!-- Top shadow/fade for content scrolling under -->
                <div class="max-w-7xl mx-auto px-6 sm:px-8 py-8 min-h-full">
                    {@render children()}
                </div>
            </main>
        </div>
    {:else}
        {@render children()}
    {/if}
{/if}

<style>
    @keyframes blob {
        0% { transform: translate(0px, 0px) scale(1); }
        33% { transform: translate(30px, -20px) scale(1.1); }
        66% { transform: translate(-20px, 20px) scale(0.9); }
        100% { transform: translate(0px, 0px) scale(1); }
    }
    .animate-blob {
        animation: blob 10s infinite;
    }
    .animation-delay-4000 {
        animation-delay: 4s;
    }
</style>