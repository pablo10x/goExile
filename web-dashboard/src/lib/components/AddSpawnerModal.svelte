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
		Sparkles
	} from 'lucide-svelte';

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
			class="modal-container relative w-full max-w-xl bg-slate-900/50 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden ring-1 ring-white/10 backdrop-blur-xl"
			transition:modalScale
		>
			<!-- Animated gradient backgrounds -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				<!-- Primary gradient blob -->
				<div
					class="gradient-blob blob-1 bg-gradient-to-br from-emerald-500/30 via-cyan-600/20 to-transparent"
				></div>

				<!-- Secondary gradient blob -->
				<div
					class="gradient-blob blob-2 bg-gradient-to-tl from-blue-400/25 via-emerald-500/15 to-transparent"
				></div>

				<!-- Tertiary gradient blob -->
				<div
					class="gradient-blob blob-3 bg-gradient-to-tr from-teal-600/20 via-cyan-400/10 to-transparent"
				></div>
			</div>

			<!-- Particles -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				{#each particles as particle (particle.id)}
					<div
						class="particle bg-emerald-400/60"
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
			<div class="absolute inset-0 opacity-15 pointer-events-none gradient-overlay"></div>

			<!-- Header Pattern -->
			<div
				class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-emerald-500 via-cyan-500 to-emerald-500 opacity-80 shimmer-bar"
			></div>

			<!-- Close Button -->
			<button
				onclick={close}
				class="absolute top-4 right-4 z-20 p-2 rounded-lg bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-700/50 transition-all"
				aria-label="Close"
			>
				<X class="w-5 h-5" />
			</button>

			<!-- Content -->
			<div class="p-6 sm:p-8 relative z-10">
				<!-- Header -->
				<div class="flex items-center gap-4 mb-6">
					<div
						class="icon-wrapper flex-shrink-0 p-3 rounded-xl bg-gradient-to-br from-emerald-500/20 to-cyan-500/20 border border-emerald-500/30 backdrop-blur-sm"
					>
						<div class="animate-icon-pop">
							<Server class="w-7 h-7 text-emerald-400" />
						</div>
					</div>
					<div>
						<h2 class="text-2xl font-bold text-slate-100 tracking-tight slide-in-text">
							Add New Spawner
						</h2>
						<p class="text-slate-500 dark:text-slate-400 text-sm mt-0.5 slide-in-text-delayed">
							Generate an enrollment key to register a new spawner
						</p>
					</div>
				</div>

				{#if loading}
					<!-- Loading State -->
					<div class="flex flex-col items-center justify-center py-12" transition:fade>
						<div class="relative">
							<div class="w-20 h-20 border-4 border-slate-300 dark:border-slate-700 rounded-full animate-pulse"></div>
							<div
								class="absolute inset-0 w-20 h-20 border-4 border-emerald-500 border-t-transparent rounded-full animate-spin"
							></div>
							<Key class="absolute inset-0 m-auto w-8 h-8 text-emerald-400 animate-pulse" />
						</div>
						<p class="mt-6 text-slate-500 dark:text-slate-400 font-medium">Generating enrollment key...</p>
					</div>
				{:else if error}
					<!-- Error State -->
					<div
						class="flex flex-col items-center justify-center py-8 error-bounce"
						transition:scale={{ start: 0.95, duration: 300, easing: cubicOut }}
					>
						<div class="p-4 rounded-full bg-red-500/10 border border-red-500/30 mb-4">
							<AlertCircle class="w-10 h-10 text-red-400 animate-shake-error" />
						</div>
						<h3 class="text-lg font-semibold text-red-400 mb-2">Failed to Generate Key</h3>
						<p class="text-slate-500 dark:text-slate-400 text-sm text-center max-w-sm mb-6">{error}</p>
						<button
							onclick={generateKey}
							class="px-5 py-2.5 text-sm font-semibold text-slate-900 dark:text-white bg-gradient-to-r from-emerald-600 to-emerald-500 hover:from-emerald-500 hover:to-emerald-400 rounded-lg shadow-lg transition-all transform hover:-translate-y-0.5 active:scale-95"
						>
							Try Again
						</button>
					</div>
				{:else if isEnrolled && enrolledSpawner}
					<!-- Spawner Enrolled Success State -->
					<div
						class="space-y-6 success-state"
						transition:scale={{ start: 0.9, duration: 400, easing: backOut }}
					>
						<!-- Success Icon with Celebration -->
						<div class="flex flex-col items-center justify-center py-4">
							<div class="relative">
								<!-- Sparkle effects -->
								<div class="absolute -inset-8 flex items-center justify-center">
									<Sparkles
										class="absolute -top-6 -left-4 w-5 h-5 text-yellow-400 animate-sparkle-1"
									/>
									<Sparkles
										class="absolute -top-4 -right-6 w-4 h-4 text-emerald-400 animate-sparkle-2"
									/>
									<Sparkles
										class="absolute -bottom-4 -left-6 w-4 h-4 text-cyan-400 animate-sparkle-3"
									/>
									<Sparkles
										class="absolute -bottom-6 -right-4 w-5 h-5 text-emerald-300 animate-sparkle-4"
									/>
								</div>

								<!-- Success circle -->
								<div
									class="relative w-24 h-24 rounded-full bg-gradient-to-br from-emerald-500/30 to-cyan-500/30 border-2 border-emerald-500/50 flex items-center justify-center success-circle"
								>
									<CheckCircle2 class="w-12 h-12 text-emerald-400 animate-check-pop" />
								</div>

								<!-- Ripple effect -->
								<div
									class="absolute inset-0 rounded-full border-2 border-emerald-400/50 animate-ripple"
								></div>
								<div
									class="absolute inset-0 rounded-full border-2 border-emerald-400/30 animate-ripple-delayed"
								></div>
							</div>

							<h3 class="mt-6 text-xl font-bold text-emerald-400 animate-text-slide">
								Spawner Enrolled Successfully!
							</h3>
							<p class="text-slate-500 dark:text-slate-400 text-sm mt-1 animate-text-slide-delayed">
								The spawner is now connected to your master server
							</p>
						</div>

						<!-- Spawner Info Card -->
						<div
							class="spawner-card relative overflow-hidden bg-slate-800/70 border border-emerald-500/30 rounded-xl p-5 backdrop-blur-sm"
						>
							<div
								class="absolute inset-0 bg-gradient-to-r from-emerald-500/5 via-cyan-500/5 to-emerald-500/5"
							></div>

							<div class="relative flex items-center gap-4">
								<div class="p-3 rounded-xl bg-emerald-500/20 border border-emerald-500/30">
									<Server class="w-8 h-8 text-emerald-400" />
								</div>
								<div class="flex-1">
									<div class="flex items-center gap-2 mb-1">
										<span class="text-lg font-bold text-slate-100">
											{enrolledSpawner.region || 'Spawner'}
										</span>
										<span
											class="px-2 py-0.5 text-xs font-semibold bg-emerald-500/20 text-emerald-400 rounded-full border border-emerald-500/30"
										>
											ID: {enrolledSpawner.id}
										</span>
									</div>
									<div class="text-sm text-slate-500 dark:text-slate-400 font-mono">
										{enrolledSpawner.host}:{enrolledSpawner.port}
									</div>
								</div>
								<div
									class="flex items-center gap-1 px-3 py-1.5 bg-emerald-500/20 rounded-lg border border-emerald-500/30"
								>
									<div class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div>
									<span class="text-xs font-semibold text-emerald-400">Online</span>
								</div>
							</div>
						</div>

						<!-- Action Buttons -->
						<div class="flex justify-center gap-3 pt-2">
							<button
								onclick={generateKey}
								class="px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-emerald-400 bg-slate-800/50 hover:bg-slate-800 border border-slate-300/50 dark:border-slate-700/50 hover:border-emerald-500/30 rounded-lg transition-all flex items-center gap-2"
							>
								<Key class="w-4 h-4" />
								Add Another Spawner
							</button>
							<button
								onclick={close}
								class="px-5 py-2.5 text-sm font-semibold text-slate-900 dark:text-white bg-gradient-to-r from-emerald-600 to-cyan-600 hover:from-emerald-500 hover:to-cyan-500 rounded-lg shadow-lg transition-all transform hover:-translate-y-0.5 active:scale-95"
							>
								Done
							</button>
						</div>
					</div>
				{:else if enrollmentKey}
					<!-- Waiting for Enrollment State -->
					<div class="space-y-6" transition:fade>
						<!-- Status Indicator -->
						<div class="flex items-center justify-center gap-2 py-2">
							<div
								class="flex items-center gap-2 px-4 py-2 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-full"
							>
								<div class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></div>
								<span class="text-sm font-medium text-slate-700 dark:text-slate-300"
									>Waiting for spawner to connect...</span
								>
							</div>
						</div>

						<!-- Countdown Timer -->
						<div class="flex items-center justify-center">
							<div class="relative">
								<!-- Circular Progress -->
								<svg class="w-28 h-28 -rotate-90 transform">
									<!-- Background circle -->
									<circle
										cx="56"
										cy="56"
										r="50"
										fill="none"
										stroke="currentColor"
										stroke-width="6"
										class="text-slate-700/50"
									/>
									<!-- Progress circle -->
									<circle
										cx="56"
										cy="56"
										r="50"
										fill="none"
										stroke="url(#timerGradient)"
										stroke-width="6"
										stroke-linecap="round"
										stroke-dasharray={314}
										stroke-dashoffset={314 - (314 * progressPercent) / 100}
										class="transition-all duration-1000 ease-linear {isExpiringSoon
											? 'animate-pulse'
											: ''}"
									/>
									<defs>
										<linearGradient id="timerGradient" x1="0%" y1="0%" x2="100%" y2="100%">
											<stop offset="0%" stop-color={isExpiringSoon ? '#ef4444' : '#10b981'} />
											<stop offset="100%" stop-color={isExpiringSoon ? '#f97316' : '#06b6d4'} />
										</linearGradient>
									</defs>
								</svg>
								<!-- Time Display -->
								<div class="absolute inset-0 flex flex-col items-center justify-center">
									<Clock
										class="w-5 h-5 mb-1 {isExpiringSoon ? 'text-red-400' : 'text-emerald-400'}"
									/>
									<span
										class="text-2xl font-bold font-mono {isExpiringSoon
											? 'text-red-400'
											: 'text-slate-100'}"
									>
										{formatTime(remainingSeconds)}
									</span>
								</div>
							</div>
						</div>

						<!-- Key Display -->
						<div class="key-container relative">
							<div
								class="absolute inset-0 bg-gradient-to-r from-emerald-500/10 via-cyan-500/10 to-emerald-500/10 rounded-xl blur-xl"
							></div>
							<div
								class="relative bg-slate-800/70 border border-slate-600/50 rounded-xl p-4 backdrop-blur-sm"
							>
								<div class="flex items-center gap-2 mb-2">
									<Key class="w-4 h-4 text-emerald-400" />
									<span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
										>Enrollment Key</span
									>
								</div>
								<div class="font-mono text-lg text-emerald-300 break-all tracking-wider">
									{enrollmentKey}
								</div>
							</div>
						</div>

						<!-- Command Display -->
						<div>
							<div class="flex items-center gap-2 mb-2">
								<Terminal class="w-4 h-4 text-slate-500 dark:text-slate-400" />
								<span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
									>Run this command on your spawner</span
								>
							</div>
							<div
								class="group relative bg-white/70 dark:bg-slate-950/70 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-4 font-mono text-sm overflow-x-auto"
							>
								<code class="text-cyan-300">./spawner</code>
								<code class="text-slate-500 dark:text-slate-400"> -key </code>
								<code class="text-emerald-400">{enrollmentKey}</code>

								<!-- Copy Button -->
								<button
									onclick={copyToClipboard}
									class="absolute top-2 right-2 p-2 rounded-lg bg-slate-800/80 border border-slate-300/50 dark:border-slate-700/50 text-slate-500 dark:text-slate-400 hover:text-emerald-400 hover:border-emerald-500/50 hover:bg-slate-700/80 transition-all opacity-0 group-hover:opacity-100 focus:opacity-100"
									aria-label="Copy command"
								>
									{#if copied}
										<Check class="w-4 h-4 text-emerald-400" />
									{:else}
										<Copy class="w-4 h-4" />
									{/if}
								</button>
							</div>
							{#if copied}
								<p class="text-xs text-emerald-400 mt-2 flex items-center gap-1" transition:fade>
									<Check class="w-3 h-3" />
									Copied to clipboard!
								</p>
							{/if}
						</div>

						<!-- Instructions -->
						<div class="bg-slate-800/30 border border-slate-300/30 dark:border-slate-700/30 rounded-xl p-4">
							<h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">Instructions</h4>
							<ol class="text-xs text-slate-500 dark:text-slate-400 space-y-1.5 list-decimal list-inside">
								<li>Copy the command above</li>
								<li>Run it on the machine where your spawner is installed</li>
								<li>The spawner will automatically register with the master server</li>
								<li>The API key will be saved for future connections</li>
							</ol>
						</div>

						<!-- Regenerate Button -->
						<div class="flex justify-center pt-2">
							<button
								onclick={generateKey}
								class="px-4 py-2 text-sm font-medium text-slate-500 dark:text-slate-400 hover:text-emerald-400 hover:bg-slate-800/50 rounded-lg transition-all flex items-center gap-2"
							>
								<Key class="w-4 h-4" />
								Generate New Key
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

	.blob-3 {
		width: 250px;
		height: 250px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		animation: float3 10s ease-in-out infinite;
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

	.animate-ripple-delayed {
		animation: ripple 2s ease-out 0.5s infinite;
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
