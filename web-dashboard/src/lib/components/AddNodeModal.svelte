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
	import { siteSettings } from '$lib/stores.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';

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

	// Registration form state
	let region = $state('');
	let maxInstances = $state(10);
	let registrationLoading = $state(false);

	// Node enrollment status
	let enrollmentStatus = $state<'active' | 'pending' | 'approved' | 'expired'>('active');
	let enrolledNode = $state<{
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

	// Progress percentage for countdown circle
	let progressPercent = $derived(enrollmentKey ? (remainingSeconds / 120) * 100 : 0);
	let isExpiringSoon = $derived(
		remainingSeconds <= 30 && remainingSeconds > 0 && (enrollmentStatus === 'active' || enrollmentStatus === 'pending')
	);
	let isEnrolled = $derived(enrollmentStatus === 'approved' && enrolledNode !== null);

	// Determine Master URL for CLI instructions
	let masterUrl = $derived.by(() => {
		if (typeof window === 'undefined') return '';
		if (window.location.port === '3001') {
			return `${window.location.protocol}//${window.location.hostname}:8081`;
		}
		return window.location.origin;
	});

	let initialCheckDone = $state(false);

	async function checkExistingOrGenerate() {
		initialCheckDone = true;
		loading = true;
		try {
			const res = await fetch('/api/enrollment/keys');
			if (res.ok) {
				const keys = (await res.json()) || [];
				if (!Array.isArray(keys)) {
					// Fallback if API response is unexpected
					generateKey();
					return;
				}

				// 1. Prioritize PENDING keys (waiting for approval)
				const pendingKey = keys.find((k: any) => k.status === 'pending');
				if (pendingKey) {
					enrollmentKey = pendingKey.key;
					enrollmentStatus = 'pending';
					enrolledNode = pendingKey.node_info;
					expiresAt = new Date(pendingKey.expires_at);

					// Initialize form defaults
					if (enrolledNode?.host) {
						region = enrolledNode.host.split('.')[0].toUpperCase();
					}

					startCountdown();
					startStatusPolling();
					loading = false;
					return;
				}

				// 2. Fallback to ACTIVE keys if they have > 30s remaining
				const activeKey = keys.find((k: any) => k.status === 'active');
				if (activeKey) {
					const exp = new Date(activeKey.expires_at);
					if (exp.getTime() - Date.now() > 30000) {
						enrollmentKey = activeKey.key;
						enrollmentStatus = 'active';
						expiresAt = exp;
						startCountdown();
						startStatusPolling();
						loading = false;
						return;
					}
				}
			}
		} catch (e) {
			console.error('Failed to list keys', e);
		}

		// 3. Generate new key if no usable keys found
		generateKey();
	}

	async function generateKey() {
		loading = true;
		error = null;
		enrollmentKey = null;
		enrollmentStatus = 'active';

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
			enrolledNode = null;

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
			if (expiresAt && (enrollmentStatus === 'active' || enrollmentStatus === 'pending')) {
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
			if (!enrollmentKey || enrollmentStatus === 'approved' || enrollmentStatus === 'expired') {
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

					if (data.status === 'used' || data.status === 'approved') {
						enrollmentStatus = 'approved';
						enrolledNode = data.node_info;
						stopPolling();
					} else if (data.status === 'pending') {
						if (enrollmentStatus !== 'pending') {
							enrollmentStatus = 'pending';
							// Initial form defaults from detected host if possible
							if (data.node_info?.host) {
								region = data.node_info.host.split('.')[0].toUpperCase();
							}
						}
						enrolledNode = data.node_info;
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

	async function registerNode() {
		if (!enrollmentKey || !region) return;
		registrationLoading = true;
		error = null;

		try {
			const response = await fetch('/api/enrollment/approve', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					key: enrollmentKey,
					region: region,
					max_instances: maxInstances
				})
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to approve registration');
			}

			// Polling will detect the status change to approved
		} catch (e: any) {
			error = e.message || 'Failed to register node';
		} finally {
			registrationLoading = false;
		}
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

		// Command is now simplified as per requirement
		const command = `./node -m ${masterUrl} -key ${enrollmentKey}`;
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
		enrollmentStatus = 'active';
		enrolledNode = null;
		region = '';
		maxInstances = 10;
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

	// Check for existing keys or generate on open
	$effect(() => {
		if (isOpen) {
			if (!enrollmentKey && !loading && !initialCheckDone) {
				checkExistingOrGenerate();
			}
		} else {
			initialCheckDone = false;
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
			class="absolute inset-0 bg-neutral-950/60 backdrop-blur-md cursor-default"
			onclick={close}
			aria-label="Close modal"
		></button>

		<!-- Modal Container -->
		<div
			class="modal-container relative w-full max-w-xl bg-neutral-900/80 border border-stone-800 rounded-none shadow-2xl overflow-hidden backdrop-blur-2xl industrial-sharp"
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
				<Icon name="ph:x-bold" size="1.25rem" />
			</button>

			<!-- Content -->
			<div class="p-8 sm:p-10 relative z-10">
				<!-- Header -->
				<div class="flex items-center gap-5 mb-8">
					<div
						class="icon-wrapper flex-shrink-0 p-4 rounded-none bg-rust/10 border border-rust/30 shadow-lg industrial-sharp"
					>
						<div class="animate-icon-pop">
							<Icon name="cpu" size="2rem" class="text-rust-light" />
						</div>
					</div>
					<div>
						<h2 class="text-3xl font-heading font-black text-white tracking-tighter uppercase slide-in-text">
							Register_Node
						</h2>
						<p class="font-jetbrains text-[10px] uppercase tracking-widest mt-1 slide-in-text-delayed text-stone-500">
							Authorize new node cluster integration
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
								<Icon name="ph:key-bold" size="2.5rem" class="text-rust animate-pulse shadow-rust/50 shadow-lg" />
							</div>
						</div>
						<p class="mt-8 font-heading font-black text-[11px] uppercase tracking-[0.2em] text-stone-500">
							AUTHORIZING_ENROLLMENT...
						</p>
					</div>
				{:else if error}
					<!-- Error State -->
					<div
						class="flex flex-col items-center justify-center py-10 error-bounce"
						transition:scale={{ start: 0.95, duration: 300, easing: cubicOut }}
					>
						<div class="p-6 rounded-none bg-red-500/5 border border-red-500/20 mb-6 shadow-2xl industrial-sharp">
							<Icon name="alert" size="3.5rem" class="text-red-500 animate-shake-error" />
						</div>
						<h3 class="text-xl font-heading font-black text-red-500 mb-2 uppercase tracking-tighter">PROTO_FAULT_0x04</h3>
						<p class="font-jetbrains text-[11px] text-center max-w-sm mb-8 uppercase leading-relaxed font-bold text-stone-500">
							{error}
						</p>
						<Button
							onclick={generateKey}
							variant="danger"
							size="md"
						>
							Retry_Sequence
						</Button>
					</div>
				{:else if isEnrolled && enrolledNode}
					<!-- Node Enrolled Success State -->
					<div
						class="space-y-8 success-state"
						transition:scale={{ start: 0.9, duration: 400, easing: backOut }}
					>
						<!-- Success Icon with Celebration -->
						<div class="flex flex-col items-center justify-center py-4">
							<div class="relative">
								<!-- Sparkle effects -->
								<div class="absolute -inset-10 flex items-center justify-center">
									<Icon name="ph:sparkle-bold" size="1.5rem" class="absolute -top-8 -left-6 text-rust animate-sparkle-1" />
									<Icon name="ph:sparkle-bold" size="1.25rem" class="absolute -top-6 -right-8 text-rust-light animate-sparkle-2" />
									<Icon name="ph:sparkle-bold" size="1.25rem" class="absolute -bottom-6 -left-8 text-rust animate-sparkle-3" />
									<Icon name="ph:sparkle-bold" size="1.5rem" class="absolute -bottom-8 -right-6 text-rust-light animate-sparkle-4" />
								</div>

								<!-- Success circle -->
								<div
									class="relative w-28 h-28 rounded-none bg-rust/10 border-2 border-rust/40 flex items-center justify-center success-circle shadow-rust/20 shadow-2xl industrial-sharp"
								>
									<Icon name="ph:check-circle-bold" size="3.5rem" class="text-rust animate-check-pop" />
								</div>

								<!-- Ripple effect -->
								<div
									class="absolute inset-0 border-2 border-rust/30 animate-ripple"
								></div>
							</div>

							<h3 class="mt-8 text-2xl font-heading font-black text-white uppercase tracking-tighter animate-text-slide">
								NODE_AUTHORIZED
							</h3>
							<p class="font-jetbrains text-[10px] uppercase tracking-widest mt-2 animate-text-slide-delayed text-stone-500">
								Sector registry updated // Handshake complete
							</p>
						</div>

						<!-- Node Info Card -->
						<div
							class="node-card modern-industrial-card bg-neutral-950/40 backdrop-blur-md border border-neutral-800 p-6"
						>
							<div class="relative flex items-center gap-6">
								<div class="p-4 rounded-none bg-rust/10 border border-rust/30 industrial-sharp"
								>
									<Icon name="cpu" size="2.5rem" class="text-rust-light" />
								</div>
								<div class="flex-1">
									<div class="flex items-center gap-3 mb-2">
										<span class="text-xl font-heading font-black text-white uppercase tracking-tighter">
											{enrolledNode.region || 'CORE_NODE'}
										</span>
										<span
											class="px-2 py-0.5 font-jetbrains text-[9px] font-black bg-rust text-white uppercase tracking-widest shadow-lg shadow-rust/20"
										>
											ID_{enrolledNode.id}
										</span>
									</div>
									<div class="font-jetbrains text-[11px] font-bold tracking-tight text-stone-500">
										INTERFACE: {enrolledNode.host}:{enrolledNode.port}
									</div>
								</div>
								<div
									class="flex items-center gap-2 px-4 py-2 bg-[var(--color-success)]/10 border border-[var(--color-success)]/30 shadow-[0_0_15px_rgba(16,185,129,0.1)]"
								>
									<div class="w-2 h-2 rounded-full bg-[var(--color-success)] animate-pulse shadow-[var(--color-success)]/50 shadow-lg"></div>
									<span class="font-heading font-black text-[10px] text-success uppercase tracking-widest">ONLINE</span>
								</div>
							</div>
						</div>

						<!-- Action Buttons -->
						<div class="flex justify-center gap-4 pt-4">
							<Button
								onclick={generateKey}
								variant="secondary"
								size="md"
								icon="ph:key-bold"
							>
								Append_Node
							</Button>
							<Button
								onclick={close}
								variant="primary"
								size="md"
							>
								Finalize
							</Button>
						</div>
					</div>
				{:else if enrollmentStatus === 'pending' && enrolledNode}
					<!-- Configuration Form State (Node found, waiting for details) -->
					<div class="space-y-6" transition:fade>
						<div class="flex items-center gap-4 p-4 bg-rust/5 border border-rust/20">
							<Icon name="ph:broadcast-bold" size="1.5rem" class="text-rust animate-pulse" />
							<div>
								<p class="font-heading font-black text-[10px] text-rust uppercase tracking-widest">SIGNAL_DETECTED</p>
								<p class="font-jetbrains text-[11px] text-white uppercase font-bold">NODE AT {enrolledNode.host}:{enrolledNode.port}</p>
							</div>
						</div>

						<div class="space-y-4">
							<div class="space-y-2">
								<label for="region" class="font-heading font-black text-[10px] text-stone-500 uppercase tracking-widest block ml-1">Assign_Sector (Region)</label>
								<input
									id="region"
									type="text"
									bind:value={region}
									placeholder="e.g. US-EAST-1"
									class="w-full bg-stone-950 border border-stone-800 p-4 font-jetbrains text-white focus:outline-none focus:border-rust transition-all uppercase placeholder:text-stone-700"
								/>
							</div>

							<div class="space-y-2">
								<label for="maxInstances" class="font-heading font-black text-[10px] text-stone-500 uppercase tracking-widest block ml-1">Instance_Capacity</label>
								<div class="flex items-center gap-4">
									<input
										id="maxInstances"
										type="range"
										min="1"
										max="100"
										bind:value={maxInstances}
										class="flex-1 accent-rust"
									/>
									<span class="w-12 text-center font-jetbrains font-black text-rust bg-stone-950 border border-stone-800 py-2">
										{maxInstances}
									</span>
								</div>
							</div>
						</div>

						<div class="pt-4">
							<Button
								onclick={registerNode}
								disabled={!region || registrationLoading}
								loading={registrationLoading}
								variant="primary"
								size="lg"
								block={true}
								icon="ph:check-bold"
							>
								Complete_Registration
							</Button>
						</div>
					</div>
				{:else if enrollmentKey}
					<!-- Waiting for Enrollment State -->
					<div class="space-y-8" transition:fade>
						<!-- Status Indicator -->
						<div class="flex items-center justify-center gap-2 py-2">
							<div
								class="flex items-center gap-3 px-5 py-2.5 bg-stone-900/50 border border-stone-800 shadow-inner"
							>
								<div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse shadow-amber-500/50 shadow-lg"></div>
								<span class="font-jetbrains text-[10px] font-black uppercase tracking-widest text-stone-500"
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
											<stop offset="0%" stop-color={isExpiringSoon ? 'var(--color-danger)' : 'var(--primary-color)'} />
											<stop offset="100%" stop-color={isExpiringSoon ? 'var(--color-danger)' : 'var(--primary-color)'} stop-opacity="0.5" />
										</linearGradient>
									</defs>
								</svg>
								<!-- Time Display -->
								<div class="absolute inset-0 flex flex-col items-center justify-center">
									<Icon name="clock" size="1.5rem" class="mb-1.5 {isExpiringSoon ? 'text-red-500' : 'text-rust'}" />
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
								class="relative bg-stone-950 border border-stone-800 p-6 shadow-inner industrial-sharp"
							>
								<div class="flex items-center gap-3 mb-3">
									<Icon name="ph:key-bold" size="1rem" class="text-rust" />
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
								<Icon name="ph:terminal-window-bold" size="1rem" class="text-stone-600" />
								<span
									class="font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest"
									>EXECUTE_CLI_DIRECTIVE</span
								>
							</div>
							<div class="group relative bg-stone-950 border border-stone-800 p-5 font-jetbrains text-xs overflow-hidden shadow-inner">
								<div class="flex items-center gap-2">
									<code class="text-rust">./node</code>
									<code class="text-stone-600"> -m </code>
									<code class="text-white font-black">{masterUrl}</code>
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
										<Icon name="ph:check-bold" size="1.25rem" />
									{:else}
										<Icon name="ph:copy-bold" size="1.25rem" />
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
							class="bg-stone-900/40 border border-stone-800 p-5 industrial-sharp"
						>
							<h4 class="font-heading font-black text-[10px] text-stone-400 uppercase tracking-[0.2em] mb-3">
								OPERATIONAL_STEPS
							</h4>
							<ol
								class="font-jetbrains text-[10px] text-stone-500 space-y-2 list-none uppercase tracking-tight"
							>
								<li class="flex gap-3"><span class="text-rust font-black">[01]</span> Copy raw directive from buffer</li>
								<li class="flex gap-3"><span class="text-rust font-black">[02]</span> Initialize on target node architecture</li>
								<li class="flex gap-3"><span class="text-rust font-black">[03]</span> Node will handshake via encrypted channel</li>
								<li class="flex gap-3"><span class="text-rust font-black">[04]</span> Return here to complete configuration</li>
							</ol>
						</div>

						<!-- Regenerate Button -->
						<div class="flex justify-center pt-2">
							<button
								onclick={generateKey}
								class="font-jetbrains text-[10px] font-black text-stone-600 hover:text-rust transition-all flex items-center gap-2 uppercase tracking-widest"
							>
								<Icon name="ph:arrows-clockwise-bold" size="0.875rem" />
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

	.node-card {
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