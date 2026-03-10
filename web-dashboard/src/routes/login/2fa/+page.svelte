<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { goto, invalidateAll } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores.svelte';
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { Shield, Smartphone, Mail, ArrowLeft } from 'lucide-svelte';

	let code = $state('');
	let emailCode = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let shake = $state(false);
	let isSubmitting = $state(false);
	let showEmailSection = $state(false);

	onMount(() => {
		mounted = true;
	});

	// Sanitize inputs (digits only)
	$effect(() => {
		code = code.replace(/\D/g, '');
		emailCode = emailCode.replace(/\D/g, '');
	});

	// Auto-verify effect for TOTP code
	$effect(() => {
		if (code.length === 6 && !loading && !isSubmitting && !showEmailSection) {
			handleVerify();
		}
	});

	// Auto-verify effect for Email code
	$effect(() => {
		if (emailCode.length === 6 && !loading && !isSubmitting && showEmailSection) {
			handleEmailVerify();
		}
	});

	async function handleVerify(e?: Event) {
		if (e) e.preventDefault();
		if (isSubmitting) return;

		isSubmitting = true;
		loading = true;

		const formData = new URLSearchParams();
		formData.append('code', code);

		try {
			const response = await apiFetch('/api/auth/2fa', {
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
					code = '';
				} else {
					isAuthenticated.set(true);
					await invalidateAll();
					await goto('/');
				}
			} else {
				triggerShake();
				code = '';
				if (response.status === 401) {
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
			const response = await apiFetch('/api/auth/email', {
				method: 'POST',
				body: formData,
				headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
			});

			if (response.status === 429) {
				goto('/login');
				return;
			}

			if (response.ok) {
				isAuthenticated.set(true);
				await invalidateAll();
				await goto('/');
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
		setTimeout(() => (shake = false), 500);
	}
</script>

<svelte:head>
	<title>Verification | Exile</title>
</svelte:head>

{#if mounted}
	<div
		class="min-h-screen w-full flex flex-col items-center justify-center p-6 relative overflow-hidden animate-gradient-mesh"
	>
		<div class="w-full max-w-[420px] relative z-10" in:fade={{ duration: 400 }}>
			<!-- Header -->
			<div class="text-center mb-8 animate-reveal" style="animation-delay: 0.1s">
				<div class="inline-flex relative mb-4">
					<div class="absolute inset-0 bg-sky-500/20 blur-2xl rounded-full"></div>
					<div
						class="relative w-14 h-14 rounded-2xl bg-slate-950 border border-white/10 flex items-center justify-center shadow-2xl"
					>
						{#if showEmailSection}
							<Mail class="w-7 h-7 text-sky-400" />
						{:else}
							<Smartphone class="w-7 h-7 text-sky-400" />
						{/if}
					</div>
				</div>
				<h1 class="text-3xl font-bold text-white tracking-tight mb-2">
					Two-Factor <span class="text-sky-500">Auth</span>
				</h1>
				<p class="text-sm text-slate-400 font-medium">
					{#if showEmailSection}
						Enter the code sent to your email
					{:else}
						Enter the code from your authenticator app
					{/if}
				</p>
			</div>

			<!-- Verification Card -->
			<div
				class="modern-card animate-reveal"
				style="animation-delay: 0.2s"
				class:animate-shake={shake}
			>
				<div class="p-8 sm:p-10">
					<form onsubmit={showEmailSection ? handleEmailVerify : handleVerify} class="space-y-8">
						<div class="relative group">
							{#if showEmailSection}
								<input
									type="text"
									bind:value={emailCode}
									maxlength="6"
									inputmode="numeric"
									autocomplete="one-time-code"
									placeholder="000000"
									class="w-full bg-slate-950/50 border border-white/5 px-4 py-6 text-center text-5xl font-mono tracking-[0.4em] text-white focus:border-sky-500/40 focus:ring-4 focus:ring-sky-500/5 outline-none transition-all placeholder:text-slate-800 rounded-2xl"
									disabled={loading}
									autofocus
								/>
							{:else}
								<input
									type="text"
									bind:value={code}
									maxlength="6"
									inputmode="numeric"
									autocomplete="one-time-code"
									placeholder="000000"
									class="w-full bg-slate-950/50 border border-white/5 px-4 py-6 text-center text-5xl font-mono tracking-[0.4em] text-white focus:border-sky-500/40 focus:ring-4 focus:ring-sky-500/5 outline-none transition-all placeholder:text-slate-800 rounded-2xl"
									disabled={loading}
									autofocus
								/>
							{/if}

							{#if loading}
								<div class="absolute right-6 top-1/2 -translate-y-1/2">
									<div
										class="w-6 h-6 border-2 border-sky-500 border-t-transparent rounded-full animate-spin"
									></div>
								</div>
							{/if}
						</div>

						{#if showEmailSection}
							<div
								in:slide={{ axis: 'y', duration: 400 }}
								class="flex items-center justify-center gap-2 text-xs font-bold text-emerald-500 bg-emerald-500/10 py-3 rounded-xl border border-emerald-500/20 uppercase tracking-wider"
							>
								<Shield class="w-4 h-4" />
								<span>Identity Verified</span>
							</div>
						{/if}

						<div class="text-center pt-4 border-t border-white/5">
							<a
								href="/login"
								class="inline-flex items-center gap-2 text-xs font-bold text-slate-500 hover:text-white uppercase tracking-widest transition-colors"
							>
								<ArrowLeft class="w-3.5 h-3.5" />
								Cancel Verification
							</a>
						</div>
					</form>
				</div>
			</div>

			<!-- Footer Info -->
			<div class="mt-8 text-center animate-reveal" style="animation-delay: 0.3s">
				<p class="text-[10px] font-bold text-slate-600 uppercase tracking-[0.3em]">
					Secure Encryption Active
				</p>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(body) {
		background-color: black;
		overflow: hidden;
	}

	@keyframes shake {
		0%,
		100% {
			transform: translateX(0);
		}
		10%,
		30%,
		50%,
		70%,
		90% {
			transform: translateX(-4px);
		}
		20%,
		40%,
		60%,
		80% {
			transform: translateX(4px);
		}
	}
	.animate-shake {
		animation: shake 0.4s ease-in-out;
	}

	@keyframes reveal-up {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-reveal {
		animation: reveal-up 1s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
		opacity: 0;
	}
</style>
