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
<div class="min-h-screen w-full flex bg-black text-slate-300 font-sans overflow-hidden">
    
    <!-- Left Side: Visuals -->
    <div class="hidden lg:flex w-1/2 relative items-center justify-center bg-slate-900 overflow-hidden">
        <!-- Animated Background Elements -->
        <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMSIgY3k9IjEiIHI9IjEiIGZpbGw9InJnYmEoMjU1LDI1NSwyNTUsMC4xKSIvPjwvc3ZnPg==')] opacity-20"></div>
        <div class="absolute top-[-20%] left-[-20%] w-[80%] h-[80%] bg-blue-600/20 rounded-full blur-[120px] animate-pulse"></div>
        <div class="absolute bottom-[-20%] right-[-20%] w-[80%] h-[80%] bg-purple-600/20 rounded-full blur-[120px] animate-pulse" style="animation-delay: 2s;"></div>
        
        <div class="relative z-10 p-12 text-center" in:fly={{ y: 20, duration: 1000, delay: 200 }}>
            <div class="mb-6 inline-flex p-4 rounded-2xl bg-white/5 border border-white/10 backdrop-blur-lg shadow-2xl">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"></path></svg>
            </div>
            <h1 class="text-5xl font-bold text-white mb-4 tracking-tight">GoExile</h1>
            <p class="text-xl text-slate-400 max-w-md mx-auto">Advanced Game Server Management Registry</p>
        </div>
    </div>

    <!-- Right Side: Login Form -->
    <div class="w-full lg:w-1/2 flex items-center justify-center p-8 bg-slate-950 relative">
        <!-- Decoration -->
        <div class="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/10 to-transparent blur-3xl"></div>

        <div class="w-full max-w-sm" in:fly={{ x: 20, duration: 800 }}>
            <div class="mb-10">
                <h2 class="text-3xl font-bold text-white mb-2">Welcome Back</h2>
                <p class="text-slate-500">Please authenticate to continue.</p>
            </div>

            <form onsubmit={handleLogin} class="space-y-6">
                <div class="space-y-2">
                    <label for="email" class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Email Address</label>
                    <div class="relative group">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
                        </div>
                        <input 
                            type="email" 
                            id="email" 
                            bind:value={email}
                            required
                            class="w-full pl-12 pr-4 py-3 bg-slate-900 border border-slate-800 rounded-lg text-slate-200 placeholder-slate-600 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all"
                            placeholder="name@example.com"
                        />
                    </div>
                </div>

                <div class="space-y-2">
                    <label for="password" class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Password</label>
                    <div class="relative group">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-blue-400 transition-colors">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                        </div>
                        <input 
                            type="password" 
                            id="password" 
                            bind:value={password}
                            required
                            class="w-full pl-12 pr-4 py-3 bg-slate-900 border border-slate-800 rounded-lg text-slate-200 placeholder-slate-600 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all"
                            placeholder="••••••••"
                        />
                    </div>
                </div>

                <button 
                    type="submit" 
                    disabled={loading}
                    class="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-3.5 rounded-lg shadow-lg shadow-blue-900/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 group"
                >
                    {#if loading}
                        <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    {:else}
                        <span>Secure Login</span>
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
                    {/if}
                </button>
            </form>
        </div>
    </div>
</div>
{/if}
