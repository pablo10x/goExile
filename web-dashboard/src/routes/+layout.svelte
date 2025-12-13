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
        <div class="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
            <div class="relative">
                <div class="animate-spin rounded-full h-16 w-16 border-4 border-slate-700 border-t-indigo-500 shadow-2xl"></div>
                <div class="absolute inset-0 rounded-full bg-gradient-to-r from-indigo-500/20 to-purple-500/20 blur-xl animate-pulse"></div>
            </div>
        </div>
    {:else}
    {#if $isAuthenticated && page.url.pathname !== '/login'}
        <div class="flex h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 text-slate-100 overflow-hidden">
            <!-- Modern Sidebar -->
            <aside class="relative w-72 bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950 border-r border-slate-700/50 flex flex-col shrink-0 overflow-hidden shadow-2xl z-20 backdrop-blur-xl">

                <!-- Enhanced Background Layers -->
                <div class="absolute inset-0 z-0">
                    <!-- Base gradient -->
                    <div class="absolute inset-0 bg-gradient-to-b from-slate-950 to-slate-900"></div>

                    <!-- Animated geometric shapes -->
                    <div class="absolute top-0 left-0 w-32 h-32 bg-gradient-to-br from-indigo-600/10 to-purple-600/10 rounded-full blur-3xl animate-pulse"></div>
                    <div class="absolute bottom-0 right-0 w-40 h-40 bg-gradient-to-tl from-cyan-600/10 to-blue-600/10 rounded-full blur-3xl animate-pulse" style="animation-delay: 2s;"></div>

                    <!-- Subtle grid pattern -->
                    <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMjAiIGN5PSIyMCIgcj0iMSIgZmlsbD0icmdiYSgyNTUsMjU1LDI1NSwwLjAzKSIvPjwvc3ZnPg==')] opacity-30"></div>

                    <!-- Glass morphism overlay -->
                    <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm border-r border-white/10"></div>
                </div>

                <!-- Content Container -->
                <div class="relative z-10 flex flex-col h-full">
                    <!-- Modern Header -->
                    <div class="p-8 border-b border-slate-700/50 bg-gradient-to-r from-slate-900/80 to-slate-800/80 backdrop-blur-md">
                        <div class="flex items-center gap-4">
                            <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                                </svg>
                            </div>
                            <div>
                                <h1 class="text-2xl font-bold bg-gradient-to-r from-indigo-400 via-purple-400 to-cyan-400 bg-clip-text text-transparent">GoExile</h1>
                                <p class="text-xs text-slate-400 font-medium tracking-wider uppercase">Control Center</p>
                            </div>
                        </div>
                    </div>

                    <!-- Enhanced Navigation -->
                    <nav class="flex-1 p-6 space-y-3 overflow-y-auto">
                        <a href="/"
                            class="group relative flex items-center gap-4 px-5 py-4 rounded-2xl transition-all duration-300 overflow-hidden
                            {isRouteActive('/') ?
                                'bg-gradient-to-r from-indigo-600/20 to-purple-600/20 text-indigo-300 shadow-lg shadow-indigo-500/10 border border-indigo-500/30 backdrop-blur-sm' :
                                'hover:bg-slate-800/50 text-slate-400 hover:text-slate-200 border border-transparent hover:border-slate-600/30'
                            }"
                        >
                            {#if isRouteActive('/')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-indigo-400 to-purple-400 rounded-r-full shadow-lg"></div>
                                <div class="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 rounded-2xl"></div>
                            {/if}
                            <div class="relative z-10 w-6 h-6 flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-all duration-300 group-hover:scale-110 {isRouteActive('/') ? 'text-indigo-400' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <rect x="3" y="3" width="7" height="7"></rect>
                                    <rect x="14" y="3" width="7" height="7"></rect>
                                    <rect x="14" y="14" width="7" height="7"></rect>
                                    <rect x="3" y="14" width="7" height="7"></rect>
                                </svg>
                            </div>
                            <span class="relative z-10 font-semibold tracking-wide">Dashboard</span>
                            {#if isRouteActive('/')}
                                <div class="ml-auto w-2 h-2 bg-indigo-400 rounded-full animate-pulse"></div>
                            {/if}
                        </a>

                        <a href="/performance" 
                            class="flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group relative overflow-hidden
                            {isRouteActive('/performance') ? 'bg-blue-600/20 text-blue-300 shadow-lg shadow-blue-900/20 border border-blue-500/30' : 'hover:bg-white/5 text-slate-400 hover:text-slate-200 hover:shadow-md border border-transparent'}"
                        >
                            {#if isRouteActive('/performance')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-blue-400 rounded-r shadow-[0_0_10px_rgba(96,165,250,0.8)]"></div>
                            {/if}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform group-hover:scale-110 group-hover:rotate-3 duration-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
                            </svg>
                            <span class="font-medium tracking-wide">Performance</span>
                        </a>

                        <a href="/config"  
                            class="flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group relative overflow-hidden
                            {isRouteActive('/config') || isRouteActive('/config/') ? 'bg-blue-600/20 text-blue-300 shadow-lg shadow-blue-900/20 border border-blue-500/30' : 'hover:bg-white/5 text-slate-400 hover:text-slate-200 hover:shadow-md border border-transparent'}"
                        >
                            {#if isRouteActive('/config') || isRouteActive('/config/')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-indigo-400 to-purple-400 rounded-r-full shadow-lg"></div>
                                <div class="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 rounded-2xl"></div>
                            {/if}
                            <div class="relative z-10 w-6 h-6 flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-all duration-300 group-hover:scale-110 group-hover:rotate-12 {isRouteActive('/config') || isRouteActive('/config/') ? 'text-indigo-400' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <circle cx="12" cy="12" r="3"></circle>
                                    <path d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"></path>
                                </svg>
                            </div>
                            <span class="relative z-10 font-semibold tracking-wide">Configuration</span>
                            {#if isRouteActive('/config') || isRouteActive('/config/')}
                                <div class="ml-auto w-2 h-2 bg-indigo-400 rounded-full animate-pulse"></div>
                            {/if}
                        </a>

                        <a href="/server"
                            class="group relative flex items-center gap-4 px-5 py-4 rounded-2xl transition-all duration-300 overflow-hidden
                            {isRouteActive('/server') ?
                                'bg-gradient-to-r from-indigo-600/20 to-purple-600/20 text-indigo-300 shadow-lg shadow-indigo-500/10 border border-indigo-500/30 backdrop-blur-sm' :
                                'hover:bg-slate-800/50 text-slate-400 hover:text-slate-200 border border-transparent hover:border-slate-600/30'
                            }"
                        >
                            {#if isRouteActive('/server')}
                                <div class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-indigo-400 to-purple-400 rounded-r-full shadow-lg"></div>
                                <div class="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 rounded-2xl"></div>
                            {/if}
                            <div class="relative z-10 w-6 h-6 flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-all duration-300 group-hover:scale-110 group-hover:-rotate-12 {isRouteActive('/server') ? 'text-indigo-400' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                    <polyline points="7 10 12 15 17 10"></polyline>
                                    <line x1="12" y1="15" x2="12" y2="3"></line>
                                </svg>
                            </div>
                            <span class="relative z-10 font-semibold tracking-wide">Server Files</span>
                            {#if isRouteActive('/server')}
                                <div class="ml-auto w-2 h-2 bg-indigo-400 rounded-full animate-pulse"></div>
                            {/if}
                        </a>
                    </nav>

                    <!-- Modern Logout Section -->
                    <div class="p-6 border-t border-slate-700/50 bg-gradient-to-t from-slate-900/50 to-transparent backdrop-blur-sm">
                        <button
                            onclick={logout}
                            class="w-full group relative flex items-center justify-center gap-3 px-6 py-4 rounded-2xl text-slate-400 hover:text-white border border-slate-600/30 hover:border-red-500/50 transition-all duration-300 overflow-hidden bg-gradient-to-r hover:from-red-600/10 hover:to-red-500/10"
                        >
                            <div class="absolute inset-0 bg-gradient-to-r from-red-600/0 via-red-600/5 to-red-600/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform group-hover:-translate-x-1 relative z-10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                                <polyline points="16 17 21 12 16 7"></polyline>
                                <line x1="21" y1="12" x2="9" y2="12"></line>
                            </svg>
                            <span class="font-semibold tracking-wide relative z-10">Sign Out</span>
                        </button>
                    </div>
                </div>
            </aside>

            <!-- Enhanced Main Content -->
            <main class="flex-1 overflow-auto bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative">
                <!-- Subtle background pattern -->
                <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMSIgZmlsbD0icmdiYSgxMDQsIDExNCwgMTIyLCAwLjAzKSIvPjwvc3ZnPg==')] opacity-20"></div>

                <div class="relative max-w-7xl mx-auto px-8 py-8 min-h-full">
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