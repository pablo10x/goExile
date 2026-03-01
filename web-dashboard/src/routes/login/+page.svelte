<script lang="ts">
import { apiFetch } from "$lib/api";
	import { goto, invalidateAll } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores.svelte';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Shield, Lock, User, ChevronRight, Activity, Cpu } from 'lucide-svelte';
	import Button from '$lib/components/Button.svelte';

	let email = $state('admin@example.com');
	let password = $state('admin123');
	let error = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let formShake = $state(false);
	let isSubmitting = $state(false);

	const randomHex = (len = 6) => `0x${Math.floor(Math.random() * Math.pow(16, len)).toString(16).padStart(len, '0')}`;

	onMount(() => {
		mounted = true;
	});

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
				console.log('Login OK:', data);
				if (data.session) {
					localStorage.setItem('exile_session', data.session);
				}
				console.log('Session Token in Storage:', localStorage.getItem('exile_session'));
				
				if (data.next_step === 'totp') {
					await goto('/login/2fa');
				} else {
					isAuthenticated.set(true);
					await invalidateAll();
					await goto('/');
				}
			} else {
				console.error('Login Failed:', response.status, response.statusText, await response.text());
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

{#if mounted}
<div class="min-h-screen w-full bg-black text-slate-300 font-sans flex items-center justify-center p-4 relative overflow-hidden">
	
	<!-- Background mesh should show through from layout, but here we add a subtle glow -->
	<div class="fixed inset-0 z-[-1] animate-gradient-mesh opacity-100"></div>
	<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-sky-500/5 rounded-full blur-[150px] pointer-events-none opacity-50"></div>

	<!-- Login Card -->
	<div 
		class="w-full max-w-md relative z-10 modern-card p-10 sm:p-12 rounded-[3rem] bg-slate-900/80 backdrop-blur-3xl border border-white/10 shadow-2xl transition-all duration-500" 
		class:animate-shake={formShake}
	>
		<div class="text-center mb-10 relative z-10">
			<div class="inline-block p-4 bg-sky-500/10 border border-sky-500/20 rounded-2xl mb-6 group">
				<Cpu class="w-10 h-10 text-sky-400 group-hover:scale-110 transition-transform duration-500" />
			</div>
			<div class="space-y-2">
				<div class="flex items-center justify-center gap-2">
					<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]"></div>
					<span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Admin Authentication</span>
				</div>
				<h1 class="text-5xl font-bold text-white tracking-tight">
					Exile<span class="text-sky-400">OS</span>
				</h1>
				<p class="text-sm font-medium text-slate-400 pt-2">Authorized access required</p>
			</div>
		</div>

		<form onsubmit={handleLogin} class="space-y-6 relative z-10">
			<div class="space-y-2">
				<label for="email" class="text-xs font-bold text-slate-500 uppercase tracking-wider ml-1">Email Address</label>
				<div class="relative group">
					<User class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-600 group-focus-within:text-sky-400 transition-colors" />
					<input
						type="email"
						id="email"
						bind:value={email}
						required
						placeholder="admin@example.com"
						class="w-full pl-12 pr-4 py-4 bg-slate-950 border border-white/5 rounded-xl text-white placeholder:text-slate-700 outline-none focus:border-sky-500/50 focus:ring-1 focus:ring-sky-500/20 transition-all font-medium text-sm"
					/>
				</div>
			</div>

			<div class="space-y-2">
				<label for="password" class="text-xs font-bold text-slate-500 uppercase tracking-wider ml-1">Password</label>
				<div class="relative group">
					<Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-600 group-focus-within:text-sky-400 transition-colors" />
					<input
						type="password"
						id="password"
						bind:value={password}
						required
						placeholder="••••••••••••"
						class="w-full pl-12 pr-4 py-4 bg-slate-950 border border-white/5 rounded-xl text-white placeholder:text-slate-700 outline-none focus:border-sky-500/50 focus:ring-1 focus:ring-sky-500/20 transition-all font-medium text-sm"
					/>
				</div>
			</div>

			{#if error}
			<div in:fade class="bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-bold p-4 rounded-xl flex items-center gap-3">
				<Shield class="w-4 h-4 shrink-0" />
				<span>{error}</span>
			</div>
			{/if}

			<Button
				type="submit"
				disabled={loading || isSubmitting}
				loading={loading}
				variant="primary"
				size="lg"
				block={true}
				class="!py-4 !rounded-xl text-base"
			>
				Sign In
			</Button>
		</form>

		<div class="text-center mt-10 space-y-1 relative z-10 border-t border-white/5 pt-8">
			<p class="text-[10px] text-slate-600 font-medium uppercase tracking-widest">Master Terminal Registry</p>
			<p class="text-[10px] text-slate-700 font-medium uppercase tracking-widest">Unauthorized Access Prohibited</p>
		</div>
	</div>
</div>
{/if}

<style>
	@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Teko:wght@700&display=swap');

	:global(body) {
		font-family: 'JetBrains Mono', monospace;
		background-color: black;
		margin: 0;
		padding: 0;
	}

	h1 {
		font-family: 'Teko', sans-serif;
	}

	.bg-grid {
		background-image:
			linear-gradient(rgba(255, 255, 255, 0.07) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255, 255, 255, 0.07) 1px, transparent 1px);
		background-size: 30px 30px;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		10%, 30%, 50%, 70%, 90% { transform: translateX(-6px); }
		20%, 40%, 60%, 80% { transform: translateX(6px); }
	}
	.animate-shake {
		animation: shake 0.4s ease-in-out;
	}
</style>