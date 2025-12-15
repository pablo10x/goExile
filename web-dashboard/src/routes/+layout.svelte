<script lang="ts">
    import '../app.css';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/state';
    import { isAuthenticated } from '$lib/stores';

    let { children } = $props();
    let isChecking = $state(true);
    
    // Animation states
    let sidebarLoaded = $state(false);
    let mouseX = $state(0);
    let mouseY = $state(0);
    let hoveredItem = $state(-1);

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
        
        // Trigger sidebar animations after mount
        setTimeout(() => {
            sidebarLoaded = true;
        }, 300);
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
        <div class="flex h-screen text-slate-300 overflow-hidden relative">
            <!-- Sidebar with 3D Background -->
            <aside class="relative w-64 bg-slate-950 border-r border-slate-800 flex flex-col shrink-0 overflow-hidden shadow-2xl z-20">
                
    <!-- Glass Surface -->
                <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-[1px] border-r border-white/5"></div>

              <!-- Content Container -->
                <div class="relative z-10 flex flex-col h-full">
                    <div class="p-6 border-b border-white/5 bg-white/5 backdrop-blur-md transform transition-all duration-700 {sidebarLoaded ? 'translate-y-0 opacity-100' : '-translate-y-4 opacity-0'}">
                        <div class="flex items-center gap-3">
                            <div class="w-3 h-3 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full animate-pulse shadow-lg shadow-blue-500/50"></div>
                            <h1 class="text-2xl font-bold text-slate-50 -tracking-wide bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent filter drop-shadow-lg animate-gradient bg-size-200">GoExile</h1>
                        </div>
                        <p class="text-[10px] text-blue-200/60 font-mono mt-1 tracking-widest animate-pulse">REGISTRY CONTROL</p>
                    </div>

                  <nav class="flex-1 p-4 space-y-2 overflow-y-auto">
                        <!-- Dashboard -->
                        <a 
                            href="/" 
                            class="nav-link"
                            class:nav-active={isRouteActive('/')}
                            onmouseenter={() => hoveredItem = 0}
                            onmouseleave={() => hoveredItem = -1}
                            style="animation-delay: 0.1s;"
                        >
                            {#if isRouteActive('/')}
                                <div class="nav-indicator"></div>
                            {/if}
                            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="3" y="3" width="7" height="7"></rect>
                                <rect x="14" y="3" width="7" height="7"></rect>
                                <rect x="14" y="14" width="7" height="7"></rect>
                                <rect x="3" y="14" width="7" height="7"></rect>
                            </svg>
                            <span>Dashboard</span>
                        </a>

                        <!-- Performance -->
                        <a 
                            href="/performance" 
                            class="nav-link"
                            class:nav-active={isRouteActive('/performance')}
                            onmouseenter={() => hoveredItem = 1}
                            onmouseleave={() => hoveredItem = -1}
                            style="animation-delay: 0.2s;"
                        >
                            {#if isRouteActive('/performance')}
                                <div class="nav-indicator"></div>
                            {/if}
                            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
                            </svg>
                            <span>Performance</span>
                        </a>

                        <!-- Configuration -->
                        <a 
                            href="/config"  
                            class="nav-link"
                            class:nav-active={isRouteActive('/config') || isRouteActive('/config/')}
                            onmouseenter={() => hoveredItem = 2}
                            onmouseleave={() => hoveredItem = -1}
                            style="animation-delay: 0.3s;"
                        >
                            {#if isRouteActive('/config') || isRouteActive('/config/')}
                                <div class="nav-indicator"></div>
                            {/if}
                            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="3"></circle>
                                <path d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"></path>
                            </svg>
                            <span>Configuration</span>
                        </a>

                        <!-- Server Files -->
                        <a 
                            href="/server" 
                            class="nav-link"
                            class:nav-active={isRouteActive('/server')}
                            onmouseenter={() => hoveredItem = 3}
                            onmouseleave={() => hoveredItem = -1}
                            style="animation-delay: 0.4s;"
                        >
                            {#if isRouteActive('/server')}
                                <div class="nav-indicator"></div>
                            {/if}
                            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                <polyline points="7 10 12 15 17 10"></polyline>
                                <line x1="12" y1="15" x2="12" y2="3"></line>
                            </svg>
                            <span>Server Files</span>
                        </a>

                        <!-- Database -->
                        <a 
                            href="/database" 
                            class="nav-link"
                            class:nav-active={isRouteActive('/database')}
                            onmouseenter={() => hoveredItem = 4}
                            onmouseleave={() => hoveredItem = -1}
                            style="animation-delay: 0.5s;"
                        >
                            {#if isRouteActive('/database')}
                                <div class="nav-indicator"></div>
                            {/if}
                            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
                                <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
                                <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
                            </svg>
                            <span>Databases</span>
                        </a>
                    </nav>

                <div class="p-4 border-t border-white/5 bg-black/20 backdrop-blur-md transform transition-all duration-700 {sidebarLoaded ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'}" style="animation-delay: 0.6s;">
                        <button 
                            onclick={logout} 
                            onmouseenter={() => hoveredItem = 5}
                            onmouseleave={() => hoveredItem = -1}
                            class="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg text-slate-400 hover:text-white border border-transparent transition-all duration-500 group relative overflow-hidden backdrop-blur-sm
                            {hoveredItem === 5 ? 'bg-gradient-to-r from-red-500/20 to-red-600/20 border-red-500/40 shadow-lg shadow-red-500/20 scale-[1.02]' : 'hover:bg-red-500/10 hover:border-red-500/30'}"
                        >
                            <!-- Animated Background -->
                            <div class="absolute inset-0 bg-gradient-to-r from-red-600/0 via-red-600/15 to-red-600/0 translate-x-[-100%] {hoveredItem === 5 ? 'translate-x-[100%]' : ''} transition-transform duration-700"></div>
                            
                            <!-- Warning Pulse on Hover -->
                            <div class="absolute inset-0 {hoveredItem === 5 ? 'animate-pulse' : ''} bg-red-500/5 rounded-lg"></div>
                            
                            <div class="relative z-10 flex items-center justify-center gap-2">
                                <div class="w-4 h-4 flex items-center justify-center {hoveredItem === 5 ? 'animate-bounce' : ''}">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 transition-all duration-500 group-hover:-translate-x-1 group-hover:scale-110" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                                        <polyline points="16 17 21 12 16 7"></polyline>
                                        <line x1="21" y1="12" x2="9" y2="12"></line>
                                    </svg>
                                </div>
                                <span class="text-sm font-medium transition-all duration-300 group-hover:translate-x-1">Logout</span>
                            </div>
                        </button>
                    </div>
      </div>
            </aside>

            <!-- Main Content -->
            <main class="flex-1 overflow-auto relative">
                 <!-- Top shadow/fade for content scrolling under -->
                <div class="max-w-7xl mx-auto px-6 sm:px-8 py-8 min-h-full">
                    {@render children()}
                </div>
            </main>
        </div>
        
        <!-- Spectacular Background for authenticated pages -->
        {#if $isAuthenticated && !isRouteActive('/login')}
            <div class="fixed inset-0 -z-50 overflow-hidden">
                <!-- Animated Gradient Background -->
                <div class="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 animate-gradient-shift"></div>
                
                <!-- Moving Nebula Clouds -->
                <div class="absolute inset-0">
                    <div class="absolute top-0 left-0 w-[60%] h-[40%] bg-slate-700/20 rounded-full blur-[150px] animate-blob" style="animation-delay: 0s;"></div>
                    <div class="absolute top-1/4 right-0 w-[50%] h-[35%] bg-gray-600/20 rounded-full blur-[150px] animate-blob" style="animation-delay: 3s;"></div>
                    <div class="absolute bottom-0 left-1/3 w-[45%] h-[30%] bg-zinc-700/20 rounded-full blur-[120px] animate-blob" style="animation-delay: 5s;"></div>
                </div>
                
                <!-- Grid Pattern Overlay -->
                <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPgogICAgPGcgZmlsbD0iIzY0NzQ4YiIgZmlsbC1vcGFjaXR5PSIwLjAzIj4KICAgICAgPHBhdGggZD0iTTM2IDM0di00aC0ydjRoLTR2Mmg0djRoMnYtNGg0di0yaC00em0wLTMwVjBoLTJ2NGgtNHYyaDR2NGgyVjZoNFY0aC00ek02IDM0di00SDR2NEgwdjJoNHY0aDJ2LTRoNHYtMkg2ek02IDRWMEg0djRIMHZoaDR2NGgyVjZoNFY0SDZ6Ii8+CiAgICA8L2c+CiAgPC9nPgo8L3N2Zz4=')] opacity-[0.03]"></div>
                
                <!-- Vignette Effect -->
                <div class="absolute inset-0 bg-radial-gradient from-transparent via-black/20 to-black/60"></div>
            </div>
            
            <!-- Floating Elements -->
            <div class="fixed inset-0 pointer-events-none z-5">
                <div class="absolute top-20 left-20 w-16 h-16 border-2 border-slate-400/20 rounded-lg animate-float" style="animation-delay: 0s;"></div>
                <div class="absolute top-40 right-32 w-12 h-12 border-2 border-gray-400/20 rounded-full animate-pulse" style="animation-delay: 1s;"></div>
                <div class="absolute bottom-32 left-1/3 w-20 h-20 border-2 border-zinc-400/20 transform rotate-45 animate-spin" style="animation-delay: 2s; animation-duration: 15s;"></div>
            </div>
        {/if}
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
    
    @keyframes gradient-shift {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
    }
    
    @keyframes float {
        0%, 100% { transform: translateY(0px); }
        50% { transform: translateY(-10px); }
    }
    
    @keyframes glow-pulse {
        0%, 100% { opacity: 0.3; }
        50% { opacity: 0.8; }
    }
    
    @keyframes slide-in-nav {
        from {
            opacity: 0;
            transform: translateX(-30px);
        }
        to {
            opacity: 1;
            transform: translateX(0);
        }
    }
    
    @keyframes neon-flicker {
        0%, 100% { opacity: 1; }
        41.99% { opacity: 1; }
        42% { opacity: 0.8; }
        43% { opacity: 1; }
        47.99% { opacity: 1; }
        48% { opacity: 0.9; }
        49% { opacity: 1; }
    }
    
    .animate-blob {
        animation: blob 10s infinite;
    }
    
    .animation-delay-4000 {
        animation-delay: 4s;
    }
    
    .animate-gradient {
        background-size: 200% 200%;
        animation: gradient-shift 3s ease-in-out infinite;
    }
    
    .animate-float {
        animation: float 3s ease-in-out infinite;
    }
    
    .animate-glow-pulse {
        animation: glow-pulse 2s ease-in-out infinite;
    }
    
    .animate-slide-nav {
        animation: slide-in-nav 0.5s ease-out forwards;
    }
    
    .animate-neon-flicker {
        animation: neon-flicker 2s infinite;
    }
    
    .bg-size-200 {
        background-size: 200% 200%;
    }
    
    /* Enhanced hover effects */
    .hover-lift:hover {
        transform: translateY(-2px) scale(1.02);
        transition: all 0.3s ease;
    }
    
    .nav-item-active {
        position: relative;
        overflow: hidden;
    }
    
    .nav-item-active::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.1), transparent);
        animation: slide-shine 3s infinite;
    }
    
    @keyframes slide-shine {
        0% { transform: translateX(-100%); }
        100% { transform: translateX(100%); }
    }
    
    /* Sidebar glow effects */
    .sidebar-glow {
        box-shadow: 
            inset 0 0 30px rgba(59, 130, 246, 0.1),
            0 0 20px rgba(59, 130, 246, 0.2);
    }
    
    /* Background Animations */
    @keyframes gradient-shift {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
    }
    
    @keyframes blob {
        0% { transform: translate(0px, 0px) scale(1); }
        33% { transform: translate(30px, -50px) scale(1.1); }
        66% { transform: translate(-20px, 20px) scale(0.9); }
        100% { transform: translate(0px, 0px) scale(1); }
    }
    
    @keyframes float {
        0%, 100% { transform: translateY(0px) rotate(0deg); }
        50% { transform: translateY(-20px) rotate(5deg); }
    }
    
    .animate-gradient-shift {
        background-size: 200% 200%;
        animation: gradient-shift 8s ease infinite;
    }
    
    .animate-blob {
        animation: blob 7s infinite;
    }
    
    .animate-float {
        animation: float 3s ease-in-out infinite;
    }
    
    .bg-radial-gradient {
        background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.2) 50%, rgba(0,0,0,0.6) 100%);
    }
    
    /* Reusable Page Container */
    .page-container {
        /* Let pages handle their own styling - this is just for global page wrapper if needed */
    }
    
    /* Navigation Links */
    .nav-link {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.75rem 1rem;
        border-radius: 0.75rem;
        transition: all 0.5s;
        position: relative;
        overflow: hidden;
        backdrop-filter: blur(8px);
        border: 1px solid transparent;
        color: rgb(148, 163, 184);
        opacity: 1;
        transform: translateX(0);
    }
    
    .nav-link:hover {
        background: linear-gradient(to right, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
        color: rgb(226, 232, 240);
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        border-color: rgba(255, 255, 255, 0.1);
        transform: scale(1.01);
    }
    
    .nav-link.nav-active {
        background: linear-gradient(to right, rgba(37, 99, 235, 0.3), rgba(37, 99, 235, 0.2));
        color: rgb(147, 197, 253);
        box-shadow: 0 25px 50px -12px rgba(37, 99, 235, 0.25);
        border-color: rgba(37, 99, 235, 0.4);
        transform: scale(1.02);
    }
    
    .nav-icon {
        width: 1.25rem;
        height: 1.25rem;
        transition: all 0.5s;
    }
    
    .nav-link:hover .nav-icon {
        transform: scale(1.25);
    }
    
    .nav-link:nth-child(1):hover .nav-icon {
        transform: scale(1.25) rotate(12deg);
    }
    
    .nav-link:nth-child(2):hover .nav-icon {
        transform: scale(1.25) rotate(180deg);
    }
    
    .nav-link:nth-child(3):hover .nav-icon {
        transform: scale(1.25) rotate(90deg);
    }
    
    .nav-link:nth-child(4):hover .nav-icon {
        transform: scale(1.25) translateY(-0.125rem);
    }
    
    .nav-link span {
        font-weight: 500;
        letter-spacing: 0.025em;
        transition: all 0.3s;
    }
    
    .nav-link:hover span {
        transform: translateX(0.25rem);
    }
    
    .nav-indicator {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        width: 0.375rem;
        background: linear-gradient(to bottom, rgb(96, 165, 250), rgb(34, 211, 238));
        border-radius: 0 0.25rem 0.25rem 0;
        box-shadow: 0 0 10px rgba(96, 165, 250, 0.5);
        animation: pulse 2s infinite;
    }
</style>