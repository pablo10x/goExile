<script lang="ts">
	import { apiFetch, checkConnection } from '$lib/api';
	import { goto, invalidateAll } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores.svelte';
	import { onMount } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import { Shield, Lock, Mail, Cpu, Sliders } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';
	import ConnectionModal from '$lib/components/ConnectionModal.svelte';
	import AppBackground from '$lib/components/theme/AppBackground.svelte';

	let email = $state('admin@example.com');
	let password = $state('admin123');
	let error = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let formShake = $state(false);
	let isSubmitting = $state(false);
	let serverStatus = $state<'checking' | 'online' | 'offline'>('checking');
	let isConnectionModalOpen = $state(false);

	onMount(() => {
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
				error = 'Invalid email or password';
				formShake = true;
				setTimeout(() => (formShake = false), 500);
			}
		} catch (e) {
			error = 'Could not connect to the authentication server';
			formShake = true;
			setTimeout(() => (formShake = false), 500);
		} finally {
			loading = false;
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>Sign In | Exile</title>
</svelte:head>

{#if mounted}
	<div
		class="min-h-screen w-full flex flex-col items-center justify-center p-6 relative overflow-hidden bg-transparent font-sans"
	>
		<AppBackground />

		<!-- Connection Status Pill -->
		<div class="fixed top-8 z-50">
			<button
				class="flex items-center gap-2.5 px-4 py-2 rounded-full bg-slate-900/40 border border-white/10 backdrop-blur-md hover:border-sky-500/30 transition-all group animate-reveal shadow-xl shadow-black/40"
				onclick={() => (isConnectionModalOpen = true)}
			>
				<div class="relative flex items-center justify-center">
					{#if serverStatus === 'checking'}
						<div class="w-2 h-2 rounded-full bg-sky-500 animate-pulse"></div>
					{:else if serverStatus === 'online'}
						<div
							class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]"
						></div>
					{:else}
						<div
							class="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.6)] animate-pulse"
						></div>
					{/if}
				</div>
				<span class="text-[10px] font-bold text-slate-300 uppercase tracking-widest">
					{serverStatus === 'online'
						? 'System Online'
						: serverStatus === 'offline'
							? 'System Offline'
							: 'Connecting...'}
				</span>
				<Sliders class="w-3 h-3 text-slate-500 group-hover:text-sky-400 transition-colors" />
			</button>
		</div>

		<!-- Main Content Container -->
		<div class="w-full max-w-[420px] relative z-10">
			<!-- Header -->
			<div class="text-center mb-10 animate-reveal" style="animation-delay: 0.1s">
				<div class="inline-flex relative mb-6">
					<div class="absolute inset-0 bg-sky-500/20 blur-2xl rounded-full"></div>
					<div
						class="relative w-16 h-16 rounded-3xl bg-slate-950 border border-white/10 flex items-center justify-center shadow-2xl"
					>
						<Cpu class="w-8 h-8 text-sky-400" />
					</div>
				</div>
				<h1 class="text-4xl font-bold text-white tracking-tighter mb-2 font-heading italic uppercase">
					Exile <span class="text-sky-400">Admin</span>
				</h1>
				<p class="text-xs font-bold text-slate-500 uppercase tracking-[0.2em]">Management Interface Login</p>
			</div>

			<!-- Login Card -->
			<div
				class="premium-card animate-reveal overflow-visible"
				style="animation-delay: 0.2s"
				class:animate-shake={formShake}
			>
				<div class="p-8 sm:p-10">
					<form onsubmit={handleLogin} class="space-y-6">
						<div class="space-y-2">
							<label
								for="email"
								class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
								>Email Address</label
							>
							<div class="relative group">
								<div
									class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-sky-500 transition-colors"
								>
									<Mail class="w-4 h-4" />
								</div>
								<input
									type="email"
									id="email"
									bind:value={email}
									required
									placeholder="admin@exile.io"
									class="glass-input !pl-12 !py-4"
								/>
							</div>
						</div>

						<div class="space-y-2">
							<label
								for="password"
								class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
								>Security Password</label
							>
							<div class="relative group">
								<div
									class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-sky-500 transition-colors"
								>
									<Lock class="w-4 h-4" />
								</div>
								<input
									type="password"
									id="password"
									bind:value={password}
									required
									placeholder="••••••••••••"
									class="glass-input !pl-12 !py-4"
								/>
							</div>
						</div>

						{#if error}
							<div
								in:slide={{ duration: 300 }}
								class="bg-rose-500/10 border border-rose-500/20 text-rose-400 text-[11px] font-bold p-4 rounded-xl flex items-center gap-3 shadow-lg shadow-rose-500/5 uppercase tracking-wide"
							>
								<Shield class="w-4 h-4 shrink-0" />
								<span>{error}</span>
							</div>
						{/if}

						<div class="pt-4">
							<Button
								type="submit"
								disabled={loading || isSubmitting}
								{loading}
								variant="primary"
								size="lg"
								block={true}
								class="!py-4 !rounded-2xl text-xs font-black uppercase tracking-widest shadow-2xl shadow-sky-500/20 hover:shadow-sky-500/40 active:scale-[0.98] transition-all"
							>
								Sign In to System
							</Button>
						</div>
					</form>
				</div>
			</div>

			<!-- Footer Info -->
			<div class="mt-10 text-center animate-reveal" style="animation-delay: 0.3s">
				<p class="text-[10px] font-bold text-slate-600 uppercase tracking-[0.3em]">
					&copy; 2026 Exile Infrastructure
				</p>
			</div>
		</div>
	</div>

	<ConnectionModal
		bind:isOpen={isConnectionModalOpen}
		onClose={() => (isConnectionModalOpen = false)}
	/>
{/if}

<style>
	:global(body) {
		background-color: #020617;
		overflow: hidden;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		10%, 30%, 50%, 70%, 90% { transform: translateX(-4px); }
		20%, 40%, 60%, 80% { transform: translateX(4px); }
	}
	.animate-shake {
		animation: shake 0.4s ease-in-out;
	}

	@keyframes reveal-up {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-reveal {
		animation: reveal-up 1s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
		opacity: 0;
	}
</style>
