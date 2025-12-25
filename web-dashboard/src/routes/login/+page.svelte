<script lang="ts">
	import { goto } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores';
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { Shield, Lock, User, ChevronRight, Activity, Terminal } from 'lucide-svelte';

	let email = $state('admin@example.com');
	let password = $state('admin123');
	let error = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let formShake = $state(false);
	let isSubmitting = $state(false);

	// Terminal simulation data
	let terminalLines = $state<string[]>([]);
	const bootSequence = [
		'[BOOT] Initializing Exile_OS Kernel...',
		'[SYS] Mounting encrypted partitions...',
		'[NET] Opening secure bridge to Master_Registry...',
		'[SEC] Firewall rules injected via RedEye...',
		'[OK] Environment stable. Awaiting credentials.'
	];

	onMount(() => {
		mounted = true;
		
		// Run boot sequence simulation
		let lineIndex = 0;
		const interval = setInterval(() => {
			if (lineIndex < bootSequence.length) {
				terminalLines = [...terminalLines, bootSequence[lineIndex]];
				lineIndex++;
			} else {
				clearInterval(interval);
			}
		}, 400);

		return () => clearInterval(interval);
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

			// Artificial delay for industrial feel
			await new Promise(r => setTimeout(r, 1200));

			if (response.ok) {
				const data = await response.json();
				if (data.next_step === 'totp') {
					goto('/login/2fa');
				} else {
					isAuthenticated.set(true);
					goto('/');
				}
			} else {
				formShake = true;
				error = 'CREDENTIAL_MISMATCH: ACCESS_DENIED';
				setTimeout(() => (formShake = false), 600);
				password = '';
			}
		} catch (e) {
			formShake = true;
			error = 'REGISTRY_UNREACHABLE: CONNECTION_FAULT';
			setTimeout(() => (formShake = false), 600);
		} finally {
			loading = false;
			isSubmitting = false;
		}
	}
</script>

{#if mounted}
	<!-- Industrial Background -->
	<div class="fixed inset-0 -z-50 bg-stone-950 overflow-hidden">
		<!-- CRT Scanline Effect -->
		<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.03] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
		
		<!-- Static Noise -->
		<div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/asfalt-dark.png')] opacity-[0.02] mix-blend-overlay"></div>

		<!-- Neural Blobs (Subtle) -->
		<div class="absolute inset-0 overflow-hidden opacity-20 pointer-events-none">
			<div class="absolute -top-1/4 -left-1/4 w-1/2 h-1/2 bg-rust/10 rounded-full blur-[120px] animate-blob"></div>
			<div class="absolute -bottom-1/4 -right-1/4 w-1/2 h-1/2 bg-stone-800/20 rounded-full blur-[120px] animate-blob" style="animation-delay: 3s"></div>
		</div>

		<!-- Animated Grid -->
		<div class="absolute inset-0 opacity-[0.03]" style="background-image: linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px); background-size: 50px 50px;"></div>
	</div>

	<div class="min-h-screen w-full flex items-center justify-center p-4 relative z-10 font-jetbrains">
		<!-- Brutalist Container -->
		<div
			class="w-full max-w-5xl bg-stone-900 border-2 border-stone-800 shadow-[12px_12px_0px_rgba(0,0,0,0.4)] flex flex-col md:flex-row overflow-hidden {formShake ? 'animate-shake' : ''}"
		>
			<!-- Left Panel: System Status -->
			<div class="w-full md:w-1/2 bg-black p-8 md:p-12 border-b-2 md:border-b-0 md:border-r-2 border-stone-800 flex flex-col justify-between relative overflow-hidden group">
				<!-- Glitch Decoration -->
				<div class="absolute top-0 left-0 w-full h-1 bg-rust opacity-30 group-hover:opacity-100 transition-opacity"></div>
				
				<div class="relative z-10">
					<div class="flex items-center gap-3 mb-8">
						<div class="p-2 bg-rust border border-rust-light shadow-[0_0_15px_rgba(120,53,15,0.4)]">
							<Terminal class="w-6 h-6 text-white" />
						</div>
						<div>
							<h1 class="text-3xl font-black military-label text-white uppercase tracking-tighter">EXILE_OS</h1>
							<span class="text-[8px] font-mono text-stone-600 tracking-[0.4em]">REGISTRY_INTERFACE_V1.0</span>
						</div>
					</div>

					<!-- Boot Sequence Display -->
					<div class="space-y-2 mt-12 bg-stone-950/50 p-4 border border-stone-800 font-mono text-[10px]">
						{#each terminalLines as line}
							<div class="flex gap-2" transition:fade>
								<span class="text-rust">>></span>
								<span class="text-stone-400 uppercase tracking-widest">{line}</span>
							</div>
						{/each}
						{#if terminalLines.length < bootSequence.length}
							<div class="flex gap-2">
								<span class="text-rust">>></span>
								<span class="w-2 h-4 bg-rust-light animate-pulse"></span>
							</div>
						{/if}
					</div>
				</div>

				<div class="mt-12 space-y-4">
					<div class="flex items-center gap-4 tactical-code text-stone-600">
						<Activity class="w-4 h-4 text-emerald-500" />
						<span>CLUSTER_HEALTH: OPTIMAL</span>
					</div>
					<div class="flex items-center gap-4 tactical-code text-stone-600">
						<Shield class="w-4 h-4 text-rust" />
						<span>SEC_PROTOCOL: REDEYE_ACTIVE</span>
					</div>
				</div>
			</div>

			<!-- Right Panel: Authentication -->
			<div class="w-full md:w-1/2 p-8 md:p-12 flex flex-col justify-center bg-stone-900/50 relative">
				<div class="mb-10">
					<h2 class="text-2xl font-black text-white uppercase tracking-widest mb-2 italic">IDENT_REQUIRED</h2>
					<p class="tactical-code text-stone-500 italic">Enter credentials to initialize secure session_</p>
				</div>

				<form onsubmit={handleLogin} class="space-y-8">
					<!-- Email -->
					<div class="space-y-2 group">
						<label for="email" class="tactical-code text-stone-400 group-focus-within:text-rust-light transition-colors font-bold">Registry_Identifier</label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-stone-600 group-focus-within:text-rust-light transition-colors">
								<User class="w-4 h-4" />
							</div>
							<input
								type="email"
								id="email"
								bind:value={email}
								required
								class="w-full pl-12 pr-4 py-4 bg-black border-2 border-stone-800 text-white text-xs focus:border-rust outline-none transition-all placeholder:text-stone-800"
								placeholder="OPERATOR@EXILE.SYS"
							/>
						</div>
					</div>

					<!-- Password -->
					<div class="space-y-2 group">
						<label for="password" class="tactical-code text-stone-400 group-focus-within:text-rust-light transition-colors font-bold">Access_Encryption_Key</label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-stone-600 group-focus-within:text-rust-light transition-colors">
								<Lock class="w-4 h-4" />
							</div>
							<input
								type="password"
								id="password"
								bind:value={password}
								required
								class="w-full pl-12 pr-4 py-4 bg-black border-2 border-stone-800 text-white text-xs focus:border-rust outline-none transition-all placeholder:text-stone-800"
								placeholder="****************"
							/>
						</div>
					</div>

					{#if error}
						<div class="p-3 bg-red-950/20 border-l-4 border-red-600 text-red-500 text-[10px] font-black uppercase italic animate-flicker">
							>> {error}
						</div>
					{/if}

					<!-- Action -->
					<button
						type="submit"
						disabled={loading}
						class="w-full py-5 bg-rust hover:bg-rust-light text-white font-black text-xs uppercase tracking-[0.3em] shadow-[6px_6px_0px_rgba(0,0,0,0.3)] transition-all active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 disabled:grayscale flex items-center justify-center gap-4 group"
					>
						{#if loading}
							<Activity class="w-5 h-5 animate-spin" />
							<span>SYNCHRONIZING...</span>
						{:else}
							<span>INITIALIZE_CORE</span>
							<ChevronRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
						{/if}
					</button>
				</form>

				<!-- System ID Footer -->
				<div class="mt-12 pt-6 border-t border-stone-800 flex justify-between items-center opacity-30">
					<span class="text-[7px] font-mono text-stone-500 uppercase tracking-widest">Node_Ref: 0x4F2A_EX</span>
					<span class="text-[7px] font-mono text-stone-500 uppercase tracking-widest">Enc_Type: AES_256_RSA</span>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes blob {
		0% { transform: translate(0px, 0px) scale(1); }
		33% { transform: translate(30px, -50px) scale(1.1); }
		66% { transform: translate(-20px, 20px) scale(0.9); }
		100% { transform: translate(0px, 0px) scale(1); }
	}

	.animate-blob {
		animation: blob 10s infinite alternate ease-in-out;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		10%, 30%, 50%, 70%, 90% { transform: translateX(-5px); }
		20%, 40%, 60%, 80% { transform: translateX(5px); }
	}

	.animate-shake {
		animation: shake 0.5s ease-in-out;
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.7; }
		55% { opacity: 0.9; }
		60% { opacity: 0.6; }
	}

	.animate-flicker {
		animation: flicker 0.2s infinite;
	}
</style>