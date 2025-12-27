<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { backOut, cubicOut } from 'svelte/easing';
	import {
		Copy,
		Key,
		Server,
		Clock,
		X,
		Check,
		AlertCircle,
		Terminal,
		CheckCircle2,
		Sparkles,
		RefreshCw
	} from 'lucide-svelte';
	import { siteSettings } from '$lib/stores';

	let {
		isOpen = $bindable(false)
	}: {
		isOpen: boolean;
	} = $props();

	let enrollmentKey = $state<string | null>(null);
	let expiresAt = $state<Date | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let copied = $state(false);
	let remainingSeconds = $state(120);
	let countdownInterval: ReturnType<typeof setInterval> | null = null;
	let statusPollInterval: ReturnType<typeof setInterval> | null = null;

	// Spawner enrollment status
	let enrollmentStatus = $state<'pending' | 'used' | 'expired'>('pending');
	let enrolledSpawner = $state<{
		id: number;
		region: string;
		host: string;
		port: number;
	} | null>(null);

	// Generate random particles for background animation
	const particles = Array.from({ length: 30 }, (_, i) => ({
		id: i,
		x: Math.random() * 100,
		y: Math.random() * 100,
		size: Math.random() * 4 + 2,
		duration: Math.random() * 4 + 3,
		delay: Math.random() * 3
	}));

	async function generateKey() {
		loading = true;
		error = null;
		enrollmentKey = null;

		try {
			const response = await fetch('/api/enrollment/generate', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include'
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to generate enrollment key');
			}

			const data = await response.json();
			enrollmentKey = data.key;
			expiresAt = new Date(data.expires_at);
			remainingSeconds = Math.floor((expiresAt.getTime() - Date.now()) / 1000);
			enrollmentStatus = 'pending';
			enrolledSpawner = null;

			// Start countdown and status polling
			startCountdown();
			startStatusPolling();
		} catch (e: any) {
			error = e.message || 'Failed to generate enrollment key';
		} finally {
			loading = false;
		}
	}

	function startCountdown() {
		if (countdownInterval) {
			clearInterval(countdownInterval);
		}

		countdownInterval = setInterval(() => {
			if (expiresAt && enrollmentStatus === 'pending') {
				remainingSeconds = Math.max(0, Math.floor((expiresAt.getTime() - Date.now()) / 1000));
				if (remainingSeconds <= 0) {
					enrollmentStatus = 'expired';
					stopPolling();
				}
			}
		}, 1000);
	}

	function startStatusPolling() {
		if (statusPollInterval) {
			clearInterval(statusPollInterval);
		}

		// Poll every 1.5 seconds
		statusPollInterval = setInterval(async () => {
			if (!enrollmentKey || enrollmentStatus !== 'pending') {
				return;
			}

			try {
				const response = await fetch('/api/enrollment/status', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					credentials: 'include',
					body: JSON.stringify({ key: enrollmentKey })
				});

				if (response.ok) {
					const data = await response.json();

					if (data.used && data.spawner_info) {
						enrollmentStatus = 'used';
						enrolledSpawner = data.spawner_info;
						stopPolling();
					} else if (data.expired) {
						enrollmentStatus = 'expired';
						stopPolling();
					}
				}
			} catch (e) {
				console.error('Failed to poll enrollment status', e);
			}
		}, 1500);
	}

	function stopPolling() {
		if (countdownInterval) {
			clearInterval(countdownInterval);
			countdownInterval = null;
		}
		if (statusPollInterval) {
			clearInterval(statusPollInterval);
			statusPollInterval = null;
		}
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	async function copyToClipboard() {
		if (!enrollmentKey) return;

		const command = `./spawner -key ${enrollmentKey}`;
		try {
			await navigator.clipboard.writeText(command);
			copied = true;
			setTimeout(() => {
				copied = false;
			}, 2000);
		} catch (e) {
			console.error('Failed to copy to clipboard', e);
		}
	}

	function close() {
		stopPolling();
		isOpen = false;
		enrollmentKey = null;
		expiresAt = null;
		error = null;
		loading = false;
		enrollmentStatus = 'pending';
		enrolledSpawner = null;
	}

	// Custom modal entrance animation
	function modalScale(node: HTMLElement, params: { duration?: number } = {}) {
		const { duration = 400 } = params;
		return {
			duration,
			css: (t: number) => {
				const eased = backOut(t);
				return `
					transform: scale(${0.9 + eased * 0.1}) translateY(${(1 - eased) * 20}px);
					opacity: ${eased};
				`;
			}
		};
	}

	// Progress percentage for countdown circle
	let progressPercent = $derived(enrollmentKey ? (remainingSeconds / 120) * 100 : 0);
	let isExpiringSoon = $derived(
		remainingSeconds <= 30 && remainingSeconds > 0 && enrollmentStatus === 'pending'
	);
	let isEnrolled = $derived(enrollmentStatus === 'used' && enrolledSpawner !== null);

	// Generate key on open
	$effect(() => {
		if (isOpen && !enrollmentKey && !loading) {
			generateKey();
		}
	});

	// Cleanup on unmount
	$effect(() => {
		return () => {
			stopPolling();
		};
	});
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop with blur -->
		<button
			class="absolute inset-0 bg-white/70 dark:bg-slate-950/70 backdrop-blur-md cursor-default"
			onclick={close}
			aria-label="Close modal"
		></button>

		<!-- Modal Container -->
		<div
			class="modal-container relative w-full max-w-xl bg-[#050505] border border-stone-800 rounded-none shadow-2xl overflow-hidden backdrop-blur-xl"
			class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
			class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
			transition:modalScale
		>
			<!-- Animated gradient backgrounds -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				<!-- Primary gradient blob -->
				<div
					class="gradient-blob blob-1 bg-gradient-to-br from-rust/20 via-rust-dark/10 to-transparent"
				></div>

				<!-- Secondary gradient blob -->
				<div
					class="gradient-blob blob-2 bg-gradient-to-tl from-stone-800/25 via-rust/10 to-transparent"
				></div>
			</div>

			<!-- Particles -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				{#each particles as particle (particle.id)}
					<div
						class="particle bg-rust/40"
						style="
							left: {particle.x}%;
							top: {particle.y}%;
							width: {particle.size}px;
							height: {particle.size}px;
							animation-duration: {particle.duration}s;
							animation-delay: {particle.delay}s;
						"
					></div>
				{/each}
			</div>

			<!-- Shimmer overlay -->
			<div class="absolute inset-0 opacity-10 pointer-events-none gradient-overlay"></div>

			<!-- Header Pattern -->
			<div
				class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-rust via-rust-light to-rust opacity-80 shimmer-bar"
			></div>

			<!-- Close Button -->
			<button
				onclick={close}
				class="absolute top-6 right-6 z-20 p-2 rounded-none bg-stone-900 border border-stone-800 text-stone-500 hover:text-white transition-all active:scale-95"
				aria-label="Close"
			>
				<X class="w-5 h-5" />
			</button>

			<!-- Content -->
			<div class="p-8 sm:p-10 relative z-10">
				<!-- Header -->
				<div class="flex items-center gap-5 mb-8">
					<div
						class="icon-wrapper flex-shrink-0 p-4 rounded-none bg-rust/10 border border-rust/30 shadow-lg"
						class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
						class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
					>
						<div class="animate-icon-pop">
							<Server class="w-8 h-8 text-rust-light" />
						</div>
					</div>
					<div>
						<h2 class="text-3xl font-heading font-black text-white tracking-tighter uppercase slide-in-text">
							Register_Node
						</h2>
						<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest mt-1 slide-in-text-delayed">
							Generate encrypted enrollment key for new cluster integration
						</p>
					</div>
				</div>

				{#if loading}
					<!-- Loading State -->
					<div class="flex flex-col items-center justify-center py-16" transition:fade>
						<div class="relative">
							<div
								class="w-24 h-24 border-2 border-stone-800 rounded-none animate-pulse"
							></div>
							<div
								class="absolute inset-0 w-24 h-24 border-2 border-rust border-t-transparent rounded-none animate-spin"
							></div>
							<div class="absolute inset-0 flex items-center justify-center">
								<Key class="w-10 h-10 text-rust animate-pulse shadow-rust/50 shadow-lg" />
							</div>
						</div>
						<p class="mt-8 font-heading font-black text-[11px] text-stone-500 uppercase tracking-[0.2em]">
							AUTHORIZING_ENROLLMENT...
						</p>
					</div>
				{:else if error}
					<!-- Error State -->
					<div
						class="flex flex-col items-center justify-center py-10 error-bounce"
						transition:scale={{ start: 0.95, duration: 300, easing: cubicOut }}
					>
						<div class="p-6 rounded-none bg-red-500/5 border border-red-500/20 mb-6 shadow-2xl"
							 class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
							 class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
						>
							<AlertCircle class="w-14 h-14 text-red-500 animate-shake-error" />
						</div>
						<h3 class="text-xl font-heading font-black text-red-500 mb-2 uppercase tracking-tighter">PROTO_FAULT_0x04</h3>
						<p class="font-jetbrains text-[11px] text-stone-500 text-center max-w-sm mb-8 uppercase leading-relaxed font-bold">
							{error}
						</p>
						<button
							onclick={generateKey}
							class="px-10 py-3 bg-red-600 hover:bg-red-500 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-red-900/20 transition-all active:translate-y-px"
						>
							Retry_Sequence
						</button>
					</div>
				{:else if isEnrolled && enrolledSpawner}
					<!-- Spawner Enrolled Success State -->
					<div
						class="space-y-8 success-state"
						transition:scale={{ start: 0.9, duration: 400, easing: backOut }}
					>
						<!-- Success Icon with Celebration -->
						<div class="flex flex-col items-center justify-center py-4">
							<div class="relative">
								<!-- Sparkle effects -->
								<div class="absolute -inset-10 flex items-center justify-center">
									<Sparkles
										class="absolute -top-8 -left-6 w-6 h-6 text-rust animate-sparkle-1"
									/>
									<Sparkles
										class="absolute -top-6 -right-8 w-5 h-5 text-rust-light animate-sparkle-2"
									/>
									<Sparkles
										class="absolute -bottom-6 -left-8 w-5 h-5 text-rust animate-sparkle-3"
									/>
									<Sparkles
										class="absolute -bottom-8 -right-6 w-6 h-6 text-rust-light animate-sparkle-4"
									/>
								</div>

								<!-- Success circle -->
								<div
									class="relative w-28 h-28 rounded-none bg-rust/10 border-2 border-rust/40 flex items-center justify-center success-circle shadow-rust/20 shadow-2xl"
									class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
									class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
								>
									<CheckCircle2 class="w-14 h-12 text-rust animate-check-pop" />
								</div>

								<!-- Ripple effect -->
								<div
									class="absolute inset-0 border-2 border-rust/30 animate-ripple"
								></div>
							</div>

							<h3 class="mt-8 text-2xl font-heading font-black text-white uppercase tracking-tighter animate-text-slide">
								NODE_AUTHORIZED
							</h3>
							<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest mt-2 animate-text-slide-delayed">
								Sector registry updated // Handshake complete
							</p>
						</div>

						<!-- Spawner Info Card -->
						<div
							class="spawner-card modern-industrial-card glass-panel p-6"
						>
							<div class="relative flex items-center gap-6">
								<div class="p-4 rounded-none bg-rust/10 border border-rust/30"
									 class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
									 class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
								>
									<Server class="w-10 h-10 text-rust-light" />
								</div>
								<div class="flex-1">
									<div class="flex items-center gap-3 mb-2">
										<span class="text-xl font-heading font-black text-white uppercase tracking-tighter">
											{enrolledSpawner.region || 'CORE_NODE'}
										</span>
										<span
											class="px-2 py-0.5 font-jetbrains text-[9px] font-black bg-rust text-white uppercase tracking-widest shadow-lg shadow-rust/20"
										>
											ID_{enrolledSpawner.id}
										</span>
									</div>
									<div class="font-jetbrains text-[11px] text-stone-500 font-bold tracking-tight">
										INTERFACE: {enrolledSpawner.host}:{enrolledSpawner.port}
									</div>
								</div>
								<div
									class="flex items-center gap-2 px-4 py-2 bg-emerald-500/10 border border-emerald-500/30 shadow-[0_0_15px_rgba(16,185,129,0.1)]"
								>
									<div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-emerald-500/50 shadow-lg"></div>
									<span class="font-heading font-black text-[10px] text-emerald-400 uppercase tracking-widest">ONLINE</span>
								</div>
							</div>
						</div>

						<!-- Action Buttons -->
						<div class="flex justify-center gap-4 pt-4">
							<button
								onclick={generateKey}
								class="px-6 py-3 font-heading font-black text-[10px] text-stone-400 hover:text-rust bg-stone-900 border border-stone-800 hover:border-rust/40 transition-all flex items-center gap-3 uppercase tracking-widest shadow-lg"
							>
								<Key class="w-4 h-4" />
								Append_Node
							</button>
							<button
								onclick={close}
								class="px-8 py-3 font-heading font-black text-[11px] text-white bg-rust hover:bg-rust-light shadow-lg shadow-rust/20 transition-all uppercase tracking-widest active:translate-y-px"
							>
								Finalize
							</button>
						</div>
					</div>
				{:else if enrollmentKey}
					<!-- Waiting for Enrollment State -->
					<div class="space-y-8" transition:fade>
						<!-- Status Indicator -->
						<div class="flex items-center justify-center gap-2 py-2">
							<div
								class="flex items-center gap-3 px-5 py-2.5 bg-stone-900 border border-stone-800 shadow-inner"
							>
								<div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse shadow-amber-500/50 shadow-lg"></div>
								<span class="font-jetbrains text-[10px] font-black text-stone-400 uppercase tracking-widest"
									>AWAITING_EXTERNAL_HANDSHAKE...</span
								>
							</div>
						</div>

						<!-- Countdown Timer -->
						<div class="flex items-center justify-center">
							<div class="relative">
								<!-- Circular Progress -->
								<svg class="w-32 h-32 -rotate-90 transform">
									<!-- Background circle -->
									<circle
										cx="64"
										cy="64"
										r="60"
										fill="none"
										stroke="currentColor"
										stroke-width="4"
										class="text-stone-900"
									/>
									<!-- Progress circle -->
									<circle
										cx="64"
										cy="64"
										r="60"
										fill="none"
										stroke="url(#timerGradient)"
										stroke-width="4"
										stroke-linecap="round"
										stroke-dasharray={377}
										stroke-dashoffset={377 - (377 * progressPercent) / 100}
										class="transition-all duration-1000 ease-linear {isExpiringSoon
											? 'animate-pulse'
											: ''}"
									/>
									<defs>
										<linearGradient id="timerGradient" x1="0%" y1="0%" x2="100%" y2="100%">
											<stop offset="0%" stop-color={isExpiringSoon ? '#ef4444' : '#f97316'} />
											<stop offset="100%" stop-color={isExpiringSoon ? '#7c2d12' : '#fb923c'} />
										</linearGradient>
									</defs>
								</svg>
								<!-- Time Display -->
								<div class="absolute inset-0 flex flex-col items-center justify-center">
									<Clock
										class="w-6 h-6 mb-1.5 {isExpiringSoon ? 'text-red-500' : 'text-rust'}"
									/>
									<span
										class="text-3xl font-heading font-black {isExpiringSoon
											? 'text-red-500'
											: 'text-white'}"
									>
										{formatTime(remainingSeconds)}
									</span>
								</div>
							</div>
						</div>

						<!-- Key Display -->
						<div class="key-container relative">
							<div
								class="absolute inset-0 bg-rust/5 rounded-none blur-xl"
							></div>
							<div
								class="relative bg-stone-950 border border-stone-800 p-6 shadow-inner"
								class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
								class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
							>
								<div class="flex items-center gap-3 mb-3">
									<Key class="w-4 h-4 text-rust" />
									<span
										class="font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-[0.2em]"
										>ENROLLMENT_KEY_RAW</span
									>
								</div>
								<div class="font-jetbrains text-2xl font-black text-rust-light break-all tracking-[0.1em] uppercase">
									{enrollmentKey}
								</div>
							</div>
						</div>

						<!-- Command Display -->
						<div class="space-y-3">
							<div class="flex items-center gap-3">
								<Terminal class="w-4 h-4 text-stone-600" />
								<span
									class="font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest"
									>EXECUTE_CLI_DIRECTIVE</span
								>
							</div>
							<div class="group relative bg-stone-950 border border-stone-800 p-5 font-jetbrains text-xs overflow-hidden shadow-inner">
								<div class="flex items-center gap-2">
									<code class="text-rust">./spawner</code>
									<code class="text-stone-600"> -key </code>
									<code class="text-white font-black">{enrollmentKey}</code>
								</div>

								<!-- Copy Button -->
								<button
									onclick={copyToClipboard}
									class="absolute top-0 right-0 h-full px-4 bg-stone-900 border-l border-stone-800 text-stone-500 hover:text-rust transition-all opacity-0 group-hover:opacity-100 active:bg-rust active:text-white"
									aria-label="Copy command"
								>
									{#if copied}
										<Check class="w-5 h-5" />
									{:else}
										<Copy class="w-5 h-5" />
									{/if}
								</button>
							</div>
							{#if copied}
								<div class="font-jetbrains text-[9px] font-black text-emerald-500 mt-2 flex items-center gap-2 uppercase tracking-widest" transition:fade>
									<div class="w-1 h-1 bg-emerald-500 animate-pulse"></div>
									Buffered_to_System_Clipboard
								</div>
							{/if}
						</div>

						<!-- Instructions -->
						<div
							class="bg-stone-900/40 border border-stone-800 p-5"
							class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
							class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
						>
							<h4 class="font-heading font-black text-[10px] text-stone-400 uppercase tracking-[0.2em] mb-3">
								OPERATIONAL_STEPS
							</h4>
							<ol
								class="font-jetbrains text-[10px] text-stone-500 space-y-2 list-none uppercase tracking-tight"
							>
								<li class="flex gap-3"><span class="text-rust font-black">[01]</span> Copy raw directive from buffer</li>
								<li class="flex gap-3"><span class="text-rust font-black">[02]</span> Initialize on target node architecture</li>
								<li class="flex gap-3"><span class="text-rust font-black">[03]</span> Spawner will handshake via encrypted channel</li>
								<li class="flex gap-3"><span class="text-rust font-black">[04]</span> Identity persistence will be cached locally</li>
							</ol>
						</div>

						<!-- Regenerate Button -->
						<div class="flex justify-center pt-2">
							<button
								onclick={generateKey}
								class="font-jetbrains text-[10px] font-black text-stone-600 hover:text-rust transition-all flex items-center gap-2 uppercase tracking-widest"
							>
								<RefreshCw class="w-3.5 h-3.5" />
								Cycle_Enrollment_Token
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-container {
		animation: modalEntry 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	@keyframes modalEntry {
		from {
			transform: scale(0.9) translateY(20px);
			opacity: 0;
		}
		to {
			transform: scale(1) translateY(0);
			opacity: 1;
		}
	}

	/* Animated gradient blobs */
	.gradient-blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(60px);
		opacity: 0.7;
	}

	.blob-1 {
		width: 350px;
		height: 350px;
		top: -120px;
		right: -120px;
		animation: float1 12s ease-in-out infinite;
	}

	.blob-2 {
		width: 300px;
		height: 300px;
		bottom: -100px;
		left: -100px;
		animation: float2 15s ease-in-out infinite;
	}

	@keyframes float1 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(-40px, 40px) scale(1.15);
		}
		66% {
			transform: translate(30px, -30px) scale(0.9);
		}
	}

	@keyframes float2 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(35px, -45px) scale(0.95);
		}
		66% {
			transform: translate(-40px, 25px) scale(1.1);
		}
	}

	@keyframes float3 {
		0%,
		100% {
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			transform: translate(-50%, -50%) scale(1.2);
		}
	}

	/* Particles */
	.particle {
		position: absolute;
		border-radius: 50%;
		pointer-events: none;
		animation: particleFloat infinite ease-in-out;
		box-shadow: 0 0 10px currentColor;
		opacity: 0;
	}

	@keyframes particleFloat {
		0%,
		100% {
			transform: translateY(0) translateX(0) scale(0);
			opacity: 0;
		}
		10% {
			opacity: 0.8;
		}
		50% {
			transform: translateY(-50px) translateX(20px) scale(1);
			opacity: 0.5;
		}
		90% {
			opacity: 0.2;
		}
		100% {
			transform: translateY(-100px) translateX(-15px) scale(0);
			opacity: 0;
		}
	}

	/* Shimmer overlay animation */
	.gradient-overlay {
		background: linear-gradient(
			45deg,
			transparent 30%,
			rgba(16, 185, 129, 0.15) 50%,
			transparent 70%
		);
		background-size: 200% 200%;
		animation: shimmer 4s ease-in-out infinite;
	}

	@keyframes shimmer {
		0%,
		100% {
			background-position: 200% 50%;
		}
		50% {
			background-position: -200% 50%;
		}
	}

	/* Header bar shimmer */
	.shimmer-bar {
		animation: shimmerBar 3s ease-in-out infinite;
	}

	@keyframes shimmerBar {
		0%,
		100% {
			opacity: 0.8;
		}
		50% {
			opacity: 1;
		}
	}

	/* Icon animation */
	.animate-icon-pop {
		animation: iconPop 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	@keyframes iconPop {
		0% {
			transform: scale(0) rotate(-180deg);
			opacity: 0;
		}
		60% {
			transform: scale(1.2) rotate(10deg);
		}
		100% {
			transform: scale(1) rotate(0deg);
			opacity: 1;
		}
	}

	.icon-wrapper {
		animation: iconGlow 2s ease-in-out 0.6s infinite;
	}

	@keyframes iconGlow {
		0%,
		100% {
			box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4);
		}
		50% {
			box-shadow: 0 0 25px 5px rgba(16, 185, 129, 0.2);
		}
	}

	/* Text slide-in animations */
	.slide-in-text {
		animation: slideInText 0.5s ease-out 0.2s both;
	}

	.slide-in-text-delayed {
		animation: slideInText 0.5s ease-out 0.35s both;
	}

	@keyframes slideInText {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Error animations */
	.error-bounce {
		animation: errorBounce 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	@keyframes errorBounce {
		0% {
			transform: scale(0.8) translateY(-10px);
			opacity: 0;
		}
		60% {
			transform: scale(1.05);
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}

	/* Key container glow */
	.key-container {
		animation: keyGlow 3s ease-in-out infinite;
	}

	@keyframes keyGlow {
		0%,
		100% {
			filter: brightness(1);
		}
		50% {
			filter: brightness(1.1);
		}
	}

	/* Success state animations */
	.success-state {
		animation: successFadeIn 0.5s ease-out;
	}

	@keyframes successFadeIn {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.success-circle {
		animation: circleGlow 2s ease-in-out infinite;
	}

	@keyframes circleGlow {
		0%,
		100% {
			box-shadow: 0 0 20px rgba(16, 185, 129, 0.3);
		}
		50% {
			box-shadow: 0 0 40px rgba(16, 185, 129, 0.5);
		}
	}

	.animate-ripple {
		animation: ripple 2s ease-out infinite;
	}

	@keyframes ripple {
		0% {
			transform: scale(1);
			opacity: 0.6;
		}
		100% {
			transform: scale(1.8);
			opacity: 0;
		}
	}

	.animate-text-slide {
		animation: textSlide 0.5s ease-out 0.3s both;
	}

	.animate-text-slide-delayed {
		animation: textSlide 0.5s ease-out 0.5s both;
	}

	@keyframes textSlide {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.spawner-card {
		animation: cardSlide 0.5s ease-out 0.4s both;
	}

	@keyframes cardSlide {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
