<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores.svelte';
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
			const response = await fetch('/api/auth/2fa', {
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
			const response = await fetch('/api/auth/email', {
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

{#if mounted}
	<!-- Industrial Background -->
	<div class="fixed inset-0 -z-50 bg-stone-950 overflow-hidden">
		<!-- CRT Scanline Effect -->
		<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.03] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
		
		<!-- Static Noise -->
		<div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/asfalt-dark.png')] opacity-[0.02] mix-blend-overlay"></div>

		<div
			class="absolute inset-0 bg-gradient-to-tr from-stone-950 via-[#0a0a0a] to-[#121212] opacity-80"
		></div>
		
		<!-- Animated Grid -->
		<div class="absolute inset-0 opacity-[0.03]" style="background-image: linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px); background-size: 50px 50px;"></div>
	</div>

	<div class="min-h-screen flex items-center justify-center p-6 font-jetbrains" in:fade={{ duration: 300 }}>
		<div class="w-full max-w-[420px] relative">
			<!-- Card -->
			<div
				class="bg-stone-900 border-2 border-stone-800 shadow-[12px_12px_0px_rgba(0,0,0,0.4)] overflow-hidden relative group industrial-frame"
			>
				<!-- Status Highlight -->
				<div
					class="absolute top-0 left-0 w-full h-1 bg-rust opacity-40 group-hover:opacity-100 transition-opacity"
				></div>

				<div class="p-10 pt-12 text-center bg-black/40">
					<!-- Icon -->
					<div
						class="w-20 h-20 mx-auto bg-stone-950 border border-stone-800 flex items-center justify-center mb-8 shadow-2xl industrial-frame transition-all group-hover:border-rust/50"
					>
						{#if showEmailSection}
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-8 h-8 text-rust"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="1.5"
								><rect width="20" height="16" x="2" y="4" rx="0" /><path
									d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"
								/></svg
							>
						{:else}
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-8 h-8 text-rust"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="1.5"
								><rect width="18" height="11" x="3" y="11" rx="0" ry="0" /><path
									d="M7 11V7a5 5 0 0 1 10 0v4"
								/></svg
							>
						{/if}
					</div>

					<h1 class="text-2xl font-black text-white mb-2 uppercase tracking-tighter italic">Auth_Verify</h1>
					<p class="text-[10px] text-stone-500 font-bold uppercase tracking-widest italic">
						{#if showEmailSection}
							Secondary_Vector: Email_Auth
						{:else}
							Primary_Vector: TOTP_Buffer
						{/if}
					</p>
				</div>

				<div class="px-10 pb-12 pt-8">
					<div class="relative" class:animate-shake={shake}>
						{#if !showEmailSection}
							<div class="space-y-8">
								<div class="relative group/input">
									<!-- svelte-ignore a11y_autofocus -->
									<input
										type="text"
										bind:value={code}
										maxlength="6"
										inputmode="numeric"
										autocomplete="one-time-code"
										class="w-full bg-black border-2 border-stone-800 px-4 py-5 text-center text-4xl font-mono tracking-[0.4em] text-white focus:border-rust outline-none transition-all placeholder:text-stone-900 shadow-inner"
										placeholder="000000"
										disabled={loading}
										autofocus
									/>
									{#if loading}
										<div class="absolute right-4 top-1/2 -tranneutral-y-1/2">
											<div
												class="w-5 h-5 border-2 border-rust border-t-transparent rounded-none animate-spin"
											></div>
										</div>
									{/if}
								</div>
							</div>
						{:else}
							<div in:slide={{ axis: 'y', duration: 400 }} class="space-y-8">
								<div
									class="flex items-center justify-center gap-3 text-[10px] font-black text-emerald-400 bg-emerald-950/10 py-3 border border-emerald-900/30 uppercase tracking-widest shadow-inner"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="w-4 h-4"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="3"><polyline points="20 6 9 17 4 12" /></svg
									>
									<span>Primary_Link_Stable</span>
								</div>

								<div class="relative group/input">
									<!-- svelte-ignore a11y_autofocus -->
									<input
										type="text"
										bind:value={emailCode}
										maxlength="6"
										inputmode="numeric"
										class="w-full bg-black border-2 border-stone-800 px-4 py-5 text-center text-4xl font-mono tracking-[0.4em] text-white focus:border-rust outline-none transition-all placeholder:text-stone-900 shadow-inner"
										placeholder="000000"
										disabled={loading}
										autofocus
									/>
									{#if loading}
										<div class="absolute right-4 top-1/2 -tranneutral-y-1/2">
											<div
												class="w-5 h-5 border-2 border-rust border-t-transparent rounded-none animate-spin"
											></div>
										</div>
									{/if}
								</div>
							</div>
						{/if}
					</div>

					<div class="mt-10 text-center border-t border-stone-800 pt-8">
						<a
							href="/login"
							class="text-[10px] font-black text-stone-600 hover:text-white uppercase tracking-[0.3em] transition-all italic"
						>
							[Abord_Sequence]
						</a>
					</div>
				</div>
			</div>

			<div
				class="text-center mt-8 text-[8px] text-stone-700 uppercase tracking-[0.5em] font-black opacity-40 italic"
			>
				Neural_Encryption: Active_GCM_256
			</div>
		</div>
	</div>
{/if}

<style>
	.animate-shake {
		animation: shake 0.4s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
	}
	@keyframes shake {
		10%,
		90% {
			transform: translate3d(-1px, 0, 0);
		}
		20%,
		80% {
			transform: translate3d(2px, 0, 0);
		}
		30%,
		50%,
		70% {
			transform: translate3d(-4px, 0, 0);
		}
		40%,
		60% {
			transform: translate3d(4px, 0, 0);
		}
	}
</style>
