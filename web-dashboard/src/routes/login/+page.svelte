<script lang="ts">
    import { goto } from '$app/navigation';
    import { isAuthenticated } from '$lib/stores';
    import { onMount } from 'svelte';
    import { fade, fly, slide } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';

    let email = 'admin@example.com';
    let password = 'admin123';
    let error = '';
    let loading = false;
    let mounted = false;

    onMount(() => {
        mounted = true;
    });

    async function handleLogin(event: Event) {
        event.preventDefault();
        loading = true;
        error = '';

        const start = Date.now();
        const formData = new URLSearchParams();
        formData.append('email', email);
        formData.append('password', password);

        try {
            const response = await fetch('/login', {
                method: 'POST',
                body: formData,
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                }
            });

            const elapsed = Date.now() - start;
            if (elapsed < 800) await new Promise(r => setTimeout(r, 800 - elapsed));

            if (response.redirected && response.url.includes('/login')) {
                 // Silent failure as per requirement (no error message to client?)
                 // User said "don't show any error messages to client".
                 // So we just reset the form or do nothing?
                 // That's bad UX, but requested.
                 // I'll just clear password and look innocent.
                 password = '';
                 // error = 'Invalid credentials'; // Suppressed
            } else if (response.ok || response.redirected) {
                isAuthenticated.set(true);
                goto('/');
            } else {
                // error = 'Login failed'; // Suppressed
                password = '';
            }
        } catch (e) {
            // error = 'An unexpected error occurred'; // Suppressed
        } finally {
            loading = false;
        }
    }
</script>

{#if mounted}
<div class="min-h-screen w-full flex bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 text-slate-100 font-sans overflow-hidden">

    <!-- Left Side: Enhanced Visuals -->
    <div class="hidden lg:flex w-1/2 relative items-center justify-center overflow-hidden">
        <!-- Modern Background Layers -->
        <div class="absolute inset-0 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950"></div>

        <!-- Animated geometric shapes -->
        <div class="absolute top-0 left-0 w-96 h-96 bg-gradient-to-br from-indigo-600/20 to-purple-600/10 rounded-full blur-3xl animate-pulse"></div>
        <div class="absolute bottom-0 right-0 w-80 h-80 bg-gradient-to-tl from-cyan-600/20 to-blue-600/10 rounded-full blur-3xl animate-pulse" style="animation-delay: 3s;"></div>
        <div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-64 h-64 bg-gradient-to-r from-emerald-600/10 to-teal-600/10 rounded-full blur-2xl animate-pulse" style="animation-delay: 6s;"></div>

        <!-- Subtle grid pattern -->
        <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMSIgZmlsbD0icmdiYSgxMDQsIDExNCwgMTIyLCAwLjAzKSIvPjwvc3ZnPg==')] opacity-30"></div>

        <!-- Glass morphism overlay -->
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"></div>

        <div class="relative z-10 p-12 text-center" in:fly={{ y: 20, duration: 1000, delay: 200 }}>
            <div class="mb-8 inline-flex p-6 rounded-3xl bg-gradient-to-br from-white/10 to-white/5 border border-white/20 backdrop-blur-xl shadow-2xl">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-20 h-20 text-indigo-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                </svg>
            </div>
            <h1 class="text-6xl font-bold bg-gradient-to-r from-indigo-400 via-purple-400 to-cyan-400 bg-clip-text text-transparent mb-6 tracking-tight">GoExile</h1>
            <p class="text-xl text-slate-300 max-w-lg mx-auto leading-relaxed">Advanced Game Server Management & Registry Platform</p>
            <div class="mt-8 flex justify-center">
                <div class="flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-500/10 border border-indigo-500/20 backdrop-blur-sm">
                    <div class="w-2 h-2 bg-indigo-400 rounded-full animate-pulse"></div>
                    <span class="text-sm font-medium text-indigo-300">System Online</span>
                </div>
            </div>
        </div>
    </div>

    <!-- Right Side: Enhanced Login Form -->
    <div class="w-full lg:w-1/2 flex items-center justify-center p-8 relative">
        <!-- Background decorations -->
        <div class="absolute top-0 right-0 w-96 h-96 bg-gradient-to-br from-indigo-500/10 to-purple-500/5 blur-3xl"></div>
        <div class="absolute bottom-0 left-0 w-80 h-80 bg-gradient-to-tr from-cyan-500/10 to-blue-500/5 blur-3xl"></div>

        <!-- Subtle grid pattern -->
        <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMjAiIGN5PSIyMCIgcj0iMSIgZmlsbD0icmdiYSgxMDQsIDExNCwgMTIyLCAwLjAyKSIvPjwvc3ZnPg==')] opacity-20"></div>

        <div class="w-full max-w-md relative z-10" in:fly={{ x: 20, duration: 800 }}>
            <!-- Form Container -->
            <div class="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl border border-slate-700/50 rounded-3xl p-8 shadow-2xl">
                <div class="mb-8 text-center">
                    <div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl mb-6 shadow-lg shadow-indigo-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                        </svg>
                    </div>
                    <h2 class="text-3xl font-bold bg-gradient-to-r from-slate-100 to-slate-200 bg-clip-text text-transparent mb-3">Welcome Back</h2>
                    <p class="text-slate-400 text-sm leading-relaxed">Access your control center with secure authentication</p>
                </div>

                <form onsubmit={handleLogin} class="space-y-6">
                    <div class="space-y-3">
                        <label for="email" class="text-xs font-bold text-slate-300 uppercase tracking-wider">Email Address</label>
                        <div class="relative group">
                            <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-slate-500 group-focus-within:text-indigo-400 transition-colors duration-300">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                                    <polyline points="22,6 12,13 2,6"></polyline>
                                </svg>
                            </div>
                            <input
                                type="email"
                                id="email"
                                bind:value={email}
                                required
                                class="w-full pl-14 pr-5 py-4 bg-slate-900/50 border border-slate-700/50 rounded-xl text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 focus:ring-2 focus:ring-indigo-500/20 transition-all duration-300 backdrop-blur-sm"
                                placeholder="admin@goexile.com"
                            />
                        </div>
                    </div>

                    <div class="space-y-3">
                        <label for="password" class="text-xs font-bold text-slate-300 uppercase tracking-wider">Password</label>
                        <div class="relative group">
                            <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-slate-500 group-focus-within:text-indigo-400 transition-colors duration-300">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                                </svg>
                            </div>
                            <input
                                type="password"
                                id="password"
                                bind:value={password}
                                required
                                class="w-full pl-14 pr-5 py-4 bg-slate-900/50 border border-slate-700/50 rounded-xl text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 focus:ring-2 focus:ring-indigo-500/20 transition-all duration-300 backdrop-blur-sm"
                                placeholder="••••••••••••"
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        class="w-full group relative overflow-hidden bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-bold py-4 rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
                    >
                        <div class="absolute inset-0 bg-gradient-to-r from-indigo-400 to-purple-400 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                        {#if loading}
                            <div class="relative z-10 w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                            <span class="relative z-10 font-semibold">Authenticating...</span>
                        {:else}
                            <span class="relative z-10 font-semibold">Access Control Center</span>
                            <svg xmlns="http://www.w3.org/2000/svg" class="relative z-10 w-5 h-5 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <line x1="5" y1="12" x2="19" y2="12"></line>
                                <polyline points="12 5 19 12 12 19"></polyline>
                            </svg>
                        {/if}
                    </button>
                </form>

                <!-- Security notice -->
                <div class="mt-6 text-center">
                    <p class="text-xs text-slate-500">🔒 Secure SSL encrypted connection</p>
                </div>
            </div>
        </div>
    </div>
</div>
{/if}
