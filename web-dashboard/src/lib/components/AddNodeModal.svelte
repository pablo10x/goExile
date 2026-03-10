<script lang="ts">
	import { apiFetch } from '$lib/api';
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

	// Progress percentage for countdown circle
	let progressPercent = $derived(enrollmentKey ? (remainingSeconds / 120) * 100 : 0);
	let isExpiringSoon = $derived(
		remainingSeconds <= 30 &&
			remainingSeconds > 0 &&
			(enrollmentStatus === 'active' || enrollmentStatus === 'pending')
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
			const res = await apiFetch('/api/enrollment/keys');
			if (res.ok) {
				const keys = (await res.json()) || [];
				if (!Array.isArray(keys)) {
					generateKey();
					return;
				}

				const pendingKey = keys.find((k: any) => k.status === 'pending');
				if (pendingKey) {
					enrollmentKey = pendingKey.key;
					enrollmentStatus = 'pending';
					enrolledNode = pendingKey.node_info;
					expiresAt = new Date(pendingKey.expires_at);

					if (enrolledNode?.host) {
						region = enrolledNode.host.split('.')[0].toUpperCase();
					}

					startCountdown();
					startStatusPolling();
					loading = false;
					return;
				}

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

		generateKey();
	}

	async function generateKey() {
		loading = true;
		error = null;
		enrollmentKey = null;
		enrollmentStatus = 'active';

		try {
			const response = await apiFetch('/api/enrollment/generate', {
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

		statusPollInterval = setInterval(async () => {
			if (!enrollmentKey || enrollmentStatus === 'approved' || enrollmentStatus === 'expired') {
				return;
			}

			try {
				const response = await apiFetch('/api/enrollment/status', {
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
			const response = await apiFetch('/api/enrollment/approve', {
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

	$effect(() => {
		if (isOpen) {
			if (!enrollmentKey && !loading && !initialCheckDone) {
				checkExistingOrGenerate();
			}
		} else {
			initialCheckDone = false;
		}
	});

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
			class="absolute inset-0 bg-slate-950/60 backdrop-blur-md cursor-default"
			onclick={close}
			aria-label="Close modal"
		></button>

		<!-- Modal Container -->
		<div
			class="relative w-full max-w-xl bg-slate-900/90 border border-white/10 rounded-3xl shadow-2xl overflow-hidden backdrop-blur-2xl"
			transition:modalScale
		>
			<!-- Close Button -->
			<button
				onclick={close}
				class="absolute top-6 right-6 z-20 p-2 rounded-xl bg-slate-950 border border-white/5 text-slate-500 hover:text-white transition-all active:scale-95"
				aria-label="Close"
			>
				<Icon name="ph:x-bold" size="1.25rem" />
			</button>

			<!-- Content -->
			<div class="p-8 sm:p-10 relative z-10">
				<!-- Header -->
				<div class="flex items-center gap-5 mb-8">
					<div
						class="flex-shrink-0 p-4 rounded-2xl bg-sky-500/10 border border-sky-500/20 shadow-lg"
					>
						<Icon name="ph:cpu-bold" size="2rem" class="text-sky-400" />
					</div>
					<div>
						<h2 class="text-3xl font-bold text-white tracking-tight">Provision Node</h2>
						<p class="text-xs font-medium text-slate-500 uppercase tracking-widest mt-1">
							Link a new node to the infrastructure
						</p>
					</div>
				</div>

				{#if loading}
					<!-- Loading State -->
					<div class="flex flex-col items-center justify-center py-16" transition:fade>
						<div class="relative">
							<div class="w-24 h-24 border-4 border-white/5 rounded-full"></div>
							<div
								class="absolute inset-0 w-24 h-24 border-4 border-sky-500 border-t-transparent rounded-full animate-spin"
							></div>
							<div class="absolute inset-0 flex items-center justify-center text-sky-500">
								<Icon name="ph:key-bold" size="2.5rem" class="animate-pulse" />
							</div>
						</div>
						<p class="mt-8 text-xs font-bold uppercase tracking-[0.2em] text-slate-500">
							Authenticating Session...
						</p>
					</div>
				{:else if error}
					<!-- Error State -->
					<div
						class="flex flex-col items-center justify-center py-10"
						transition:scale={{ start: 0.95, duration: 300, easing: cubicOut }}
					>
						<div class="p-6 rounded-2xl bg-rose-500/5 border border-rose-500/20 mb-6 shadow-2xl">
							<Icon name="ph:warning-circle-bold" size="3.5rem" class="text-rose-500" />
						</div>
						<h3 class="text-xl font-bold text-rose-500 mb-2">Provisioning Error</h3>
						<p class="text-sm text-center max-w-sm mb-8 text-slate-400 leading-relaxed font-medium">
							{error}
						</p>
						<Button onclick={generateKey} variant="danger" size="md" class="!rounded-xl">
							Retry Provisioning
						</Button>
					</div>
				{:else if isEnrolled && enrolledNode}
					<!-- Node Enrolled Success State -->
					<div class="space-y-8" transition:scale={{ start: 0.9, duration: 400, easing: backOut }}>
						<div class="flex flex-col items-center justify-center py-4">
							<div class="relative">
								<div
									class="w-28 h-28 rounded-full bg-emerald-500/10 border-2 border-emerald-500/40 flex items-center justify-center shadow-emerald-500/20 shadow-2xl"
								>
									<Icon name="ph:check-circle-bold" size="3.5rem" class="text-emerald-500" />
								</div>
							</div>

							<h3 class="mt-8 text-2xl font-bold text-white tracking-tight">Node Authorized</h3>
							<p class="text-xs font-medium uppercase tracking-widest mt-2 text-slate-500">
								Infrastructure registry updated successfully
							</p>
						</div>

						<!-- Node Info Card -->
						<div class="bg-slate-950/40 backdrop-blur-md border border-white/5 p-6 rounded-2xl">
							<div class="relative flex items-center gap-6">
								<div class="p-4 rounded-xl bg-sky-500/10 border border-sky-500/20">
									<Icon name="ph:cpu-bold" size="2.5rem" class="text-sky-400" />
								</div>
								<div class="flex-1">
									<div class="flex items-center gap-3 mb-2">
										<span class="text-xl font-bold text-white tracking-tight">
											{enrolledNode.region || 'Active Node'}
										</span>
										<span
											class="px-2 py-0.5 text-[10px] font-bold bg-sky-500 text-white uppercase tracking-widest rounded shadow-lg"
										>
											ID {enrolledNode.id}
										</span>
									</div>
									<div class="text-xs font-mono font-medium text-slate-500">
										Endpoint: {enrolledNode.host}:{enrolledNode.port}
									</div>
								</div>
								<div
									class="flex items-center gap-2 px-4 py-2 bg-emerald-500/10 border border-emerald-500/30 rounded-xl"
								>
									<div
										class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]"
									></div>
									<span class="text-[10px] font-bold text-emerald-500 uppercase tracking-widest"
										>Online</span
									>
								</div>
							</div>
						</div>

						<div class="flex justify-center gap-4 pt-4">
							<Button
								onclick={generateKey}
								variant="secondary"
								size="md"
								icon="ph:plus-bold"
								class="!rounded-xl">Add Another</Button
							>
							<Button onclick={close} variant="primary" size="md" class="!rounded-xl"
								>Finalize</Button
							>
						</div>
					</div>
				{:else if enrollmentStatus === 'pending' && enrolledNode}
					<!-- Configuration Form State -->
					<div class="space-y-6" transition:fade>
						<div
							class="flex items-center gap-4 p-5 bg-sky-500/5 border border-sky-500/20 rounded-2xl"
						>
							<Icon name="ph:broadcast-bold" size="1.5rem" class="text-sky-400 animate-pulse" />
							<div>
								<p class="text-[10px] font-bold text-sky-400 uppercase tracking-widest">
									Handshake Detected
								</p>
								<p class="text-xs text-white font-mono uppercase font-bold mt-0.5">
									Link request from {enrolledNode.host}
								</p>
							</div>
						</div>

						<div class="space-y-5">
							<div class="space-y-2">
								<label
									for="region"
									class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block ml-1"
									>Assigned Region</label
								>
								<input
									id="region"
									type="text"
									bind:value={region}
									placeholder="e.g. US-EAST-1"
									class="w-full bg-slate-950 border border-white/10 p-4 rounded-xl text-white focus:outline-none focus:border-sky-500/50 transition-all uppercase placeholder:text-slate-700 font-medium"
								/>
							</div>

							<div class="space-y-2">
								<label
									for="maxInstances"
									class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block ml-1"
									>Instance Capacity ({maxInstances})</label
								>
								<div class="flex items-center gap-4">
									<input
										id="maxInstances"
										type="range"
										min="1"
										max="100"
										bind:value={maxInstances}
										class="flex-1 accent-sky-500"
									/>
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
								class="!rounded-xl"
								icon="ph:check-bold"
							>
								Complete Provisioning
							</Button>
						</div>
					</div>
				{:else if enrollmentKey}
					<!-- Waiting for Enrollment State -->
					<div class="space-y-8" transition:fade>
						<div class="flex items-center justify-center gap-2">
							<div
								class="flex items-center gap-3 px-5 py-2.5 bg-slate-950/50 border border-white/5 rounded-full shadow-inner"
							>
								<div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></div>
								<span class="text-[10px] font-bold uppercase tracking-widest text-slate-500"
									>Awaiting External Link...</span
								>
							</div>
						</div>

						<div class="flex items-center justify-center">
							<div class="relative">
								<svg class="w-32 h-32 -rotate-90 transform">
									<circle
										cx="64"
										cy="64"
										r="60"
										fill="none"
										stroke="currentColor"
										stroke-width="4"
										class="text-white/5"
									/>
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
											<stop offset="0%" stop-color={isExpiringSoon ? '#f43f5e' : '#0ea5e9'} />
											<stop
												offset="100%"
												stop-color={isExpiringSoon ? '#f43f5e' : '#0ea5e9'}
												stop-opacity="0.5"
											/>
										</linearGradient>
									</defs>
								</svg>
								<div class="absolute inset-0 flex flex-col items-center justify-center">
									<Icon
										name="ph:clock-bold"
										size="1.5rem"
										class="mb-1.5 {isExpiringSoon ? 'text-rose-500' : 'text-sky-500'}"
									/>
									<span
										class="text-3xl font-bold {isExpiringSoon
											? 'text-rose-500'
											: 'text-white'} tabular-nums"
									>
										{formatTime(remainingSeconds)}
									</span>
								</div>
							</div>
						</div>

						<div class="relative">
							<div
								class="relative bg-slate-950 border border-white/10 p-6 rounded-2xl shadow-inner text-center"
							>
								<div class="flex items-center justify-center gap-3 mb-3">
									<Icon name="ph:key-bold" size="1rem" class="text-sky-500" />
									<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
										>Enrollment Key</span
									>
								</div>
								<div
									class="text-2xl font-mono font-bold text-white break-all tracking-widest uppercase"
								>
									{enrollmentKey}
								</div>
							</div>
						</div>

						<div class="space-y-3">
							<div class="flex items-center gap-3 ml-1">
								<Icon name="ph:terminal-window-bold" size="1rem" class="text-slate-600" />
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
									>Run setup command</span
								>
							</div>
							<div
								class="group relative bg-slate-950 border border-white/5 p-5 rounded-2xl overflow-hidden shadow-inner"
							>
								<div class="text-xs font-mono">
									<code class="text-sky-400">./node</code>
									<code class="text-slate-600"> -m </code>
									<code class="text-white">{masterUrl}</code>
									<code class="text-slate-600"> -key </code>
									<code class="text-white">{enrollmentKey}</code>
								</div>

								<button
									onclick={copyToClipboard}
									class="absolute top-0 right-0 h-full px-4 bg-white/5 border-l border-white/5 text-slate-500 hover:text-white transition-all opacity-0 group-hover:opacity-100 active:bg-sky-500 active:text-white"
									aria-label="Copy setup command"
								>
									{#if copied}
										<Icon name="ph:check-bold" size="1.25rem" />
									{:else}
										<Icon name="ph:copy-bold" size="1.25rem" />
									{/if}
								</button>
							</div>
							{#if copied}
								<div
									class="text-[9px] font-bold text-emerald-500 mt-2 flex items-center gap-2 uppercase tracking-widest"
									transition:fade
								>
									<div class="w-1 h-1 bg-emerald-500 rounded-full animate-pulse"></div>
									Copied to clipboard
								</div>
							{/if}
						</div>

						<!-- Instructions -->
						<div class="bg-white/[0.02] border border-white/5 p-6 rounded-2xl">
							<h4 class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-4 ml-1">
								Setup Steps
							</h4>
							<ol class="text-[11px] text-slate-500 space-y-3 font-medium">
								<li class="flex gap-3">
									<span class="text-sky-500 font-bold">01</span> Copy the setup command above
								</li>
								<li class="flex gap-3">
									<span class="text-sky-500 font-bold">02</span> Run the command on your target server
								</li>
								<li class="flex gap-3">
									<span class="text-sky-500 font-bold">03</span> Node will connect automatically
								</li>
								<li class="flex gap-3">
									<span class="text-sky-500 font-bold">04</span> Configure node settings in dashboard
								</li>
							</ol>
						</div>

						<div class="flex justify-center pt-2">
							<button
								onclick={generateKey}
								class="text-[10px] font-bold text-slate-600 hover:text-sky-400 transition-all flex items-center gap-2 uppercase tracking-widest"
							>
								<Icon name="ph:arrows-clockwise-bold" size="0.875rem" />
								Refresh Token
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
</style>
