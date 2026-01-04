<script lang="ts">
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
			const response = await fetch('/api/auth/login', {
				method: 'POST',
				body: formData,
				headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
			});

			if (response.ok) {
				const data = await response.json();
				if (data.next_step === 'totp') {
					await goto('/login/2fa');
				} else {
					isAuthenticated.set(true);
					await invalidateAll();
					await goto('/');
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

{#if mounted}
<div class="min-h-screen w-full bg-black text-stone-300 font-jetbrains flex items-center justify-center p-4 sm:p-8 relative overflow-hidden">
	<!-- Cinematic Background -->
	<div class="absolute inset-0 bg-grid opacity-[0.03]"></div>
	<div class="absolute top-1/2 left-1/2 -tranneutral-x-1/2 -tranneutral-y-1/2 w-[800px] h-[800px] bg-rust/5 rounded-full blur-[150px] pointer-events-none opacity-50"></div>
	<div class="fixed inset-0 pointer-events-none z-[100] bg-vignette opacity-60"></div>

	<!-- Login Chassis -->
	<div 
		class="w-full max-w-md relative z-10 modern-industrial-card glass-panel p-8 sm:p-12 !rounded-none shadow-[0_0_100px_rgba(0,0,0,0.8)] border-stone-800 transition-all duration-500" 
		class:animate-shake={formShake}
	>
		<!-- Tactical Corners -->
		<div class="corner-tl"></div>
		<div class="corner-tr"></div>
		<div class="corner-bl"></div>
		<div class="corner-br"></div>

		<div class="text-center mb-12 relative z-10">
			<div class="inline-block p-5 bg-rust/5 border border-rust/20 industrial-frame mb-6 group">
				<Cpu class="w-12 h-12 text-rust group-hover:scale-110 transition-transform duration-500" />
			</div>
			<div class="space-y-2">
				<div class="flex items-center justify-center gap-3">
					<div class="w-2 h-2 bg-rust animate-pulse shadow-[0_0_10px_var(--color-rust)]"></div>
					<span class="text-[10px] font-black text-text-dim uppercase tracking-[0.4em]">Secure_Uplink_v4.2</span>
				</div>
				<h1 class="text-5xl font-heading font-black text-white uppercase tracking-tighter leading-none">
					EXILE_<span class="text-rust">CORE</span>
				</h1>
				<p class="text-[10px] font-black text-text-dim uppercase tracking-widest italic pt-2">Neural Identity Verification Required</p>
			</div>
		</div>

		<form onsubmit={handleLogin} class="space-y-8 relative z-10">
			<div class="space-y-3">
				<label for="email" class="text-[9px] font-black text-text-dim uppercase tracking-[0.3em] ml-1">Identity_Signature</label>
				<div class="relative group">
					<User class="absolute left-4 top-1/2 -tranneutral-y-1/2 w-4 h-4 text-stone-700 group-focus-within:text-rust transition-colors" />
					<input
						type="email"
						id="email"
						bind:value={email}
						required
						placeholder="OPERATOR@SYSTEM.NODE"
						class="w-full pl-12 pr-4 py-4 bg-stone-950/50 border border-stone-800 text-white placeholder:text-stone-800 outline-none focus:border-rust transition-all duration-300 font-jetbrains text-sm industrial-frame"
					/>
				</div>
			</div>

			<div class="space-y-3">
				<label for="password" class="text-[9px] font-black text-text-dim uppercase tracking-[0.3em] ml-1">Access_Cipher</label>
				<div class="relative group">
					<Lock class="absolute left-4 top-1/2 -tranneutral-y-1/2 w-4 h-4 text-stone-700 group-focus-within:text-rust transition-colors" />
					<input
						type="password"
						id="password"
						bind:value={password}
						required
						placeholder="••••••••••••"
						class="w-full pl-12 pr-4 py-4 bg-stone-950/50 border border-stone-800 text-white placeholder:text-stone-800 outline-none focus:border-rust transition-all duration-300 font-jetbrains text-sm industrial-frame"
					/>
				</div>
			</div>

			{#if error}
			<div in:fade class="bg-red-950/20 border border-red-900/50 text-danger text-[10px] font-black p-4 flex items-center gap-4 uppercase tracking-widest shadow-inner industrial-frame">
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
				class="!py-5"
			>
				Authorize Access
			</Button>
		</form>

		<div class="text-center mt-12 space-y-2 relative z-10">
			<div class="h-px w-full bg-gradient-to-r from-transparent via-stone-800 to-transparent mb-6"></div>
			<p class="text-[8px] text-stone-700 font-mono uppercase tracking-[0.5em]">Node: {`main-registry-${randomHex()}`}</p>
			<p class="text-[8px] text-stone-800 font-mono uppercase tracking-[0.3em]">Authorized Personnel Only</p>
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