<script lang="ts">
    import { apiFetch, checkConnection, API_BASE } from "$lib/api";
	import { goto, invalidateAll } from '$app/navigation';
	import { isAuthenticated, sysState } from '$lib/stores.svelte';
	import { onMount } from 'svelte';
	import { fade, scale, slide } from 'svelte/transition';
	import { Shield, Lock, User, Cpu, Globe, Sliders, Server, Activity } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';
    import ConnectionModal from '$lib/components/ConnectionModal.svelte';
    import MotherboardBackground from '$lib/components/theme/MotherboardBackground.svelte';

	let email = $state('admin@example.com');
	let password = $state('admin123');
	let error = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let formShake = $state(false);
	let isSubmitting = $state(false);
    let serverStatus = $state<'checking' | 'online' | 'offline'>('checking');
    let isConnectionModalOpen = $state(false);

	onMount(async () => {
		mounted = true;
        checkServer();
        const interval = setInterval(checkServer, 10000);
        return () => clearInterval(interval);
	});

    async function checkServer() {
        const isOnline = await checkConnection();
        serverStatus = isOnline ? 'online' : 'offline';
    }

	async function handleLogin(event: Event) {
		event.preventDefault();
		if (isSubmitting) return;
		isSubmitting = true;
		loading = true;
		error = '';

		const formData = new URLSearchParams();
		formData.append('email', email);
		formData.append('password', password);

		try {
			const response = await apiFetch('/api/auth/login', {
				method: 'POST',
				body: formData,
				headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
			});

			if (response.ok) {
				const data = await response.json();
				if (data.session) {
					localStorage.setItem('exile_session', data.session);
				}
				
				if (data.next_step === 'totp') {
					await goto('/login/2fa');
				} else {
					isAuthenticated.set(true);
					await invalidateAll();
					await goto('/dashboard');
				}
			} else {
				error = 'ACCESS_DENIED :: CREDENTIALS_INVALID';
				formShake = true;
				setTimeout(() => formShake = false, 500);
			}
		} catch (e) {
			error = 'CONNECTION_ERROR :: REGISTRY_UNREACHABLE';
			formShake = true;
			setTimeout(() => formShake = false, 500);
		} finally {
			loading = false;
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
    <title>Login | Exile Dashboard</title>
</svelte:head>

{#if mounted}
<div class="min-h-screen w-full flex flex-col items-center justify-center p-4 sm:p-8 relative overflow-hidden bg-black selection:bg-sky-500/30">
    <MotherboardBackground />

    <!-- Top Status Bar -->
    <div class="fixed top-0 left-0 right-0 h-16 flex items-center justify-between px-6 sm:px-12 z-50 pointer-events-none">
        <div class="flex items-center gap-4 animate-reveal" style="animation-delay: 0.1s">
            <div class="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/20 flex items-center justify-center">
                <span class="text-sky-500 font-bold text-xs">E</span>
            </div>
            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] hidden sm:block">Exile Dashboard // v0.9.4</span>
        </div>

        <button 
            class="pointer-events-auto flex items-center gap-3 px-4 py-2 rounded-xl bg-slate-900/40 border border-white/5 backdrop-blur-md hover:bg-slate-800/60 transition-all group animate-reveal"
            style="animation-delay: 0.2s"
            onclick={() => isConnectionModalOpen = true}
        >
            <div class="relative">
                {#if serverStatus === 'checking'}
                    <div class="w-2 h-2 rounded-full bg-sky-500 animate-pulse"></div>
                {:else if serverStatus === 'online'}
                    <div class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
                {:else}
                    <div class="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)] animate-pulse"></div>
                {/if}
            </div>
            <div class="flex flex-col items-start leading-none">
                <span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest group-hover:text-slate-300 transition-colors">Server Status</span>
                <span class="text-[10px] font-bold {serverStatus === 'online' ? 'text-emerald-500' : serverStatus === 'offline' ? 'text-rose-500' : 'text-sky-500'} uppercase">
                    {serverStatus === 'online' ? 'Connected' : serverStatus === 'offline' ? 'Disconnected' : 'Verifying...'}
                </span>
            </div>
            <Sliders class="w-3 h-3 text-slate-600 group-hover:text-sky-400 transition-colors ml-1" />
        </button>
    </div>

    <!-- Main Content Container -->
    <div class="w-full max-w-[440px] relative z-10 perspective-1000">
        
        <!-- Entrance Logo -->
        <div class="text-center mb-10 animate-reveal" style="animation-delay: 0.3s">
            <div class="inline-flex relative mb-6">
                <div class="absolute inset-0 bg-sky-500/20 blur-2xl rounded-full"></div>
                <div class="relative w-16 h-16 rounded-2xl bg-slate-900 border border-white/10 flex items-center justify-center shadow-2xl">
                    <Cpu class="w-8 h-8 text-sky-400" />
                </div>
            </div>
            <h1 class="text-4xl font-bold text-white tracking-tighter mb-2 font-heading uppercase italic">
                System<span class="text-sky-500">Login</span>
            </h1>
            <p class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.3em]">Management Console Interface</p>
        </div>

        <!-- Login Card -->
        <div 
            class="modern-card p-1 sm:p-1.5 animate-reveal" 
            style="animation-delay: 0.4s"
            class:animate-shake={formShake}
        >
            <div class="bg-slate-950/80 rounded-[2.3rem] p-8 sm:p-10 border border-white/5 relative overflow-hidden group">
                
                <!-- Corner Accents -->
                <div class="absolute top-0 left-0 w-8 h-8 border-t-2 border-l-2 border-sky-500/30 rounded-tl-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
                <div class="absolute top-0 right-0 w-8 h-8 border-t-2 border-r-2 border-sky-500/30 rounded-tr-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
                <div class="absolute bottom-0 left-0 w-8 h-8 border-b-2 border-l-2 border-sky-500/30 rounded-bl-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
                <div class="absolute bottom-0 right-0 w-8 h-8 border-b-2 border-r-2 border-sky-500/30 rounded-br-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>

                <form onsubmit={handleLogin} class="space-y-6 relative z-10">
                    <div class="space-y-2">
                        <div class="flex items-center justify-between ml-1">
                            <label for="email" class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">User Identity</label>
                            <User class="w-3 h-3 text-slate-700" />
                        </div>
                        <div class="relative group/input">
                            <input
                                type="email"
                                id="email"
                                bind:value={email}
                                required
                                placeholder="ADMIN@EXILE.NETWORK"
                                class="w-full px-6 py-4 bg-black border border-white/5 rounded-2xl text-white placeholder:text-slate-800 outline-none focus:border-sky-500/40 focus:ring-1 focus:ring-sky-500/10 transition-all font-mono text-sm tracking-tight"
                            />
                            <div class="absolute bottom-0 left-6 right-6 h-[1px] bg-gradient-to-r from-transparent via-sky-500/0 to-transparent group-focus-within/input:via-sky-500/50 transition-all duration-700"></div>
                        </div>
                    </div>

                    <div class="space-y-2">
                        <div class="flex items-center justify-between ml-1">
                            <label for="password" class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Password</label>
                            <Lock class="w-3 h-3 text-slate-700" />
                        </div>
                        <div class="relative group/input">
                            <input
                                type="password"
                                id="password"
                                bind:value={password}
                                required
                                placeholder="••••••••••••"
                                class="w-full px-6 py-4 bg-black border border-white/5 rounded-2xl text-white placeholder:text-slate-800 outline-none focus:border-sky-500/40 focus:ring-1 focus:ring-sky-500/10 transition-all font-mono text-sm tracking-tight"
                            />
                            <div class="absolute bottom-0 left-6 right-6 h-[1px] bg-gradient-to-r from-transparent via-sky-500/0 to-transparent group-focus-within/input:via-sky-500/50 transition-all duration-700"></div>
                        </div>
                    </div>

                    {#if error}
                        <div in:slide={{ duration: 300 }} class="bg-rose-500/10 border border-rose-500/20 text-rose-400 text-[10px] font-bold p-4 rounded-2xl flex items-center gap-3 animate-shake">
                            <div class="p-1.5 bg-rose-500/20 rounded-lg">
                                <Shield class="w-3.5 h-3.5" />
                            </div>
                            <span class="tracking-widest uppercase">{error}</span>
                        </div>
                    {/if}

                    <div class="pt-2">
                        <Button
                            type="submit"
                            disabled={loading || isSubmitting}
                            loading={loading}
                            variant="primary"
                            size="lg"
                            block={true}
                            class="!py-5 !rounded-2xl text-xs font-bold uppercase tracking-[0.2em] shadow-xl shadow-sky-500/10 hover:shadow-sky-500/20 active:scale-[0.98] transition-all"
                        >
                            Sign In
                        </Button>
                    </div>
                </form>

                <div class="mt-10 pt-8 border-t border-white/5 flex flex-col items-center gap-3">
                    <div class="flex items-center gap-6">
                        <div class="flex flex-col items-center gap-1">
                            <Server class="w-3.5 h-3.5 text-slate-700" />
                            <span class="text-[8px] font-bold text-slate-600 tracking-tighter">SECURE</span>
                        </div>
                        <div class="w-px h-6 bg-white/5"></div>
                        <div class="flex flex-col items-center gap-1">
                            <Shield class="w-3.5 h-3.5 text-slate-700" />
                            <span class="text-[8px] font-bold text-slate-600 tracking-tighter">ENCRYPTED</span>
                        </div>
                        <div class="w-px h-6 bg-white/5"></div>
                        <div class="flex flex-col items-center gap-1">
                            <Activity class="w-3.5 h-3.5 text-slate-700" />
                            <span class="text-[8px] font-bold text-slate-600 tracking-tighter">MONITORED</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Footer Info -->
        <div class="mt-8 text-center space-y-2 animate-reveal" style="animation-delay: 0.6s">
            <p class="text-[10px] font-bold text-slate-600 uppercase tracking-[0.4em]">Proprietary Dashboard Registry</p>
            <div class="flex items-center justify-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full bg-slate-800"></span>
                <span class="text-[8px] font-bold text-slate-700 tracking-widest">UNAUTHORIZED ACCESS IS PROHIBITED</span>
                <span class="w-1.5 h-1.5 rounded-full bg-slate-800"></span>
            </div>
        </div>
    </div>

    <!-- Background Elements -->
    <div class="fixed bottom-0 left-0 w-full h-1/2 bg-gradient-to-t from-sky-500/5 to-transparent pointer-events-none opacity-40"></div>
</div>

<ConnectionModal bind:isOpen={isConnectionModalOpen} />
{/if}

<style>
    :global(body) {
        background-color: black;
        overflow: hidden;
    }

    .perspective-1000 {
        perspective: 1000px;
    }

    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        10%, 30%, 50%, 70%, 90% { transform: translateX(-4px); }
        20%, 40%, 60%, 80% { transform: translateX(4px); }
    }
    .animate-shake {
        animation: shake 0.4s ease-in-out;
    }

    /* Custom reveal animation since it's not global */
    @keyframes reveal-up {
        from { opacity: 0; transform: translateY(30px) scale(0.98); }
        to { opacity: 1; transform: translateY(0) scale(1); }
    }
    .animate-reveal {
        animation: reveal-up 1s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
        opacity: 0;
    }
</style>
