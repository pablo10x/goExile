<script lang="ts">
    import { goto } from '$app/navigation';
    import { isAuthenticated } from '$lib/stores';
    import { onMount } from 'svelte';
    import { fade, slide } from 'svelte/transition';

    let code = '';
    let emailCode = '';
    let loading = false;
    let mounted = false;
    let shake = false;
    let isSubmitting = false;
    let showEmailSection = false;

    onMount(() => {
        mounted = true;
    });

    // Sanitize inputs (digits only)
    $: code = code.replace(/\D/g, '');
    $: emailCode = emailCode.replace(/\D/g, '');

    // Auto-verify effect for TOTP code
    $: if (code.length === 6 && !loading && !isSubmitting && !showEmailSection) {
        handleVerify();
    }

    // Auto-verify effect for Email code
    $: if (emailCode.length === 6 && !loading && !isSubmitting && showEmailSection) {
        handleEmailVerify();
    }

    async function handleVerify(e?: Event) {
        if (e) e.preventDefault();
        if (isSubmitting) return;
        
        isSubmitting = true;
        loading = true;

        const formData = new URLSearchParams();
        formData.append('code', code);

        try {
            const response = await fetch('/login/2fa', {
                method: 'POST',
                body: formData,
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
            });

            if (response.status === 429) {
                goto('/login');
                return;
            }

            const data = await response.json().catch(() => ({}));

            if (response.ok) {
                if (data.next_step === 'email') {
                    showEmailSection = true;
                    // Clear code to avoid re-triggering watcher, though !showEmailSection guard handles it
                    code = ''; 
                } else {
                    isAuthenticated.set(true);
                    goto('/');
                }
            } else {
                triggerShake();
                code = '';
                if (response.url.includes('/login') && !response.url.includes('2fa')) {
                    goto('/login');
                }
            }
        } catch {
            triggerShake();
        } finally {
            loading = false;
            isSubmitting = false;
        }
    }

    async function handleEmailVerify(e?: Event) {
        if (e) e.preventDefault();
        if (isSubmitting) return;
        
        isSubmitting = true;
        loading = true;

        const formData = new URLSearchParams();
        formData.append('code', emailCode);

        try {
            const response = await fetch('/login/email', {
                method: 'POST',
                body: formData,
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
            });

            if (response.status === 429) { goto('/login'); return; }

            if (response.ok) {
                isAuthenticated.set(true);
                goto('/');
            } else {
                triggerShake();
                emailCode = '';
            }
        } catch {
            triggerShake();
        } finally {
            loading = false;
            isSubmitting = false;
        }
    }

    function triggerShake() {
        shake = true;
        setTimeout(() => shake = false, 500);
    }
</script>

{#if mounted}
    <!-- Background: Grey Modern Texture -->
    <div class="fixed inset-0 -z-50 bg-slate-950 overflow-hidden">
        <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZmlsdGVyIGlkPSJnoiPjxmZVR1cmJ1bGVuY2UgdHlwZT0iZnJhY3RhbE5vaXNlIiBiYXNlRnJlcXVlbmN5PSIwLjY1IiBudW1PY3RhdmVzPSIzIiBzdGl0Y2hUaWxlcz0ic3RpdGNoIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSIgZmlsdGVyPSJ1cmwoI2cpIiBvcGFjaXR5PSIwLjAzIi8+PC9zdmc+')] opacity-40"></div>
        <div class="absolute inset-0 bg-gradient-to-tr from-slate-950 via-slate-900 to-slate-800 opacity-80"></div>
        <!-- Subtle Glow -->
        <div class="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] bg-white/5 blur-[120px] rounded-full"></div>
    </div>

    <div class="min-h-screen flex items-center justify-center p-6" in:fade={{ duration: 300 }}>
        <div class="w-full max-w-[420px] relative">
            
            <!-- Card -->
            <div class="bg-slate-900/50 backdrop-blur-xl border border-slate-800/60 rounded-2xl shadow-2xl overflow-hidden relative group">
                <!-- Top Highlight Line -->
                <div class="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-slate-500/20 to-transparent"></div>

                <div class="p-8 pt-10 text-center">
                    <!-- Icon -->
                    <div class="w-16 h-16 mx-auto bg-slate-800/50 rounded-2xl flex items-center justify-center mb-6 shadow-inner border border-slate-700/50 group-hover:border-slate-600/50 transition-colors">
                        {#if showEmailSection}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7 text-slate-200" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
                        {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7 text-slate-200" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                        {/if}
                    </div>

                    <h1 class="text-2xl font-medium text-slate-100 mb-2 tracking-tight">Authentication</h1>
                    <p class="text-sm text-slate-400 font-light">
                        {#if showEmailSection}
                            Code sent to your email
                        {:else}
                            Enter code from authenticator
                        {/if}
                    </p>
                </div>

                <div class="px-8 pb-10">
                    <div class="relative" class:animate-shake={shake}>
                        {#if !showEmailSection}
                            <div class="space-y-6">
                                <div class="relative group/input">
                                    <!-- svelte-ignore a11y_autofocus -->
                                    <input 
                                        type="text" 
                                        bind:value={code}
                                        maxlength="6"
                                        inputmode="numeric"
                                        autocomplete="one-time-code"
                                        class="w-full bg-slate-950/50 border border-slate-700/50 rounded-xl px-4 py-4 text-center text-3xl font-mono tracking-[0.5em] text-white placeholder:text-slate-700 focus:outline-none focus:border-slate-500 focus:ring-1 focus:ring-slate-500/50 transition-all disabled:opacity-50"
                                        placeholder="······"
                                        disabled={loading}
                                        autofocus
                                    />
                                    {#if loading}
                                        <div class="absolute right-4 top-1/2 -translate-y-1/2">
                                            <div class="w-4 h-4 border-2 border-slate-500 border-t-transparent rounded-full animate-spin"></div>
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        {:else}
                            <div in:slide={{ axis: 'y', duration: 400 }} class="space-y-6">
                                <div class="flex items-center justify-center gap-2 text-xs text-emerald-400 bg-emerald-950/30 py-2 rounded-lg border border-emerald-900/50 mb-4">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
                                    <span>Authenticator verified</span>
                                </div>

                                <div class="relative group/input">
                                    <!-- svelte-ignore a11y_autofocus -->
                                    <input 
                                        type="text" 
                                        bind:value={emailCode}
                                        maxlength="6"
                                        inputmode="numeric"
                                        class="w-full bg-slate-950/50 border border-slate-700/50 rounded-xl px-4 py-4 text-center text-3xl font-mono tracking-[0.5em] text-white placeholder:text-slate-700 focus:outline-none focus:border-slate-500 focus:ring-1 focus:ring-slate-500/50 transition-all disabled:opacity-50"
                                        placeholder="······"
                                        disabled={loading}
                                        autofocus
                                    />
                                    {#if loading}
                                        <div class="absolute right-4 top-1/2 -translate-y-1/2">
                                            <div class="w-4 h-4 border-2 border-slate-500 border-t-transparent rounded-full animate-spin"></div>
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        {/if}
                    </div>

                    <div class="mt-8 text-center">
                        <a href="/login" class="text-xs text-slate-500 hover:text-slate-300 transition-colors">
                            Return to login
                        </a>
                    </div>
                </div>
            </div>
            
            <div class="text-center mt-6 text-[10px] text-slate-600 uppercase tracking-widest font-medium">
                Secured by GoExile
            </div>
        </div>
    </div>
{/if}

<style>
    .animate-shake {
        animation: shake 0.4s cubic-bezier(.36,.07,.19,.97) both;
    }
    @keyframes shake {
        10%, 90% { transform: translate3d(-1px, 0, 0); }
        20%, 80% { transform: translate3d(2px, 0, 0); }
        30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
        40%, 60% { transform: translate3d(4px, 0, 0); }
    }
</style>
