<script lang="ts">
    import './layout.css';
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
</script>

{#if isChecking}
    <div class="flex items-center justify-center min-h-screen bg-slate-900 text-slate-300">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
{:else}
    {#if $isAuthenticated && page.url.pathname !== '/login'}
        <div class="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-slate-300">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <!-- Header -->
                <header class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-6 mb-12">
                    <div>
                        <h1 class="text-4xl font-bold text-slate-50 -tracking-wide mb-2">GoExile Registry</h1>
                        <p class="text-slate-400">Real-Time Monitoring Dashboard</p>
                    </div>
                    <div class="flex gap-3 items-center flex-col sm:flex-row">
                         <!-- Navigation -->
                        <nav class="flex gap-2 mr-4">
                            <a href="/" class="btn-secondary text-xs {page.url.pathname === '/' ? 'bg-slate-300 text-slate-900' : ''}">
                                Dashboard
                            </a>
                            <a href="/server" class="btn-secondary text-xs {page.url.pathname === '/server' ? 'bg-slate-300 text-slate-900' : ''}">
                                Server Files
                            </a>
                        </nav>
                        
                        <button onclick={logout} class="btn-primary gap-2">
                            <span>Logout</span>
                        </button>
                    </div>
                </header>

                <main>
                    {@render children()}
                </main>
            </div>
        </div>
    {:else}
        {@render children()}
    {/if}
{/if}