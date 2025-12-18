<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';

	export let isOpen: boolean = false;
	export let title: string = 'Confirm Action';
	export let message: string = 'Are you sure you want to proceed?';
	export let confirmText: string = 'Confirm';
	export let cancelText: string = 'Cancel';
	export let isCritical: boolean = false; // If true, use red colors
	export let progress: number | null = null; // 0-100, if set shows progress bar
	export let statusMessage: string | null = null; // Message during loading/progress

	export let onConfirm: () => Promise<void>;

	const dispatch = createEventDispatcher();

	let loading = false;
	let error: string | null = null;

	// Generate random particles
	const particles = Array.from({ length: 25 }, (_, i) => ({
		id: i,
		x: Math.random() * 100,
		y: Math.random() * 100,
		size: Math.random() * 4 + 2,
		duration: Math.random() * 3 + 2,
		delay: Math.random() * 2
	}));

	async function handleConfirm() {
		loading = true;
		error = null;
		try {
			await onConfirm();
			dispatch('success');
			close();
		} catch (e: any) {
			error = e.message || 'An unexpected error occurred.';
			loading = false;
		}
	}

	function close() {
		if (loading && progress !== null) return;
		isOpen = false;
		error = null;
		loading = false;
		dispatch('close');
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

	// Dynamic icon based on criticality
	$: icon = isCritical
		? '<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-red-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>'
		: '<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop with blur and subtle animation -->
		<div
			class="absolute inset-0 bg-slate-950/60 backdrop-blur-md backdrop-animate"
			onclick={!loading ? close : undefined}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && !loading && close()}
			role="button"
			tabindex="0"
			aria-label="Close dialog"
		></div>

		<!-- Modal Container -->
		<div
			class="modal-container relative w-full max-w-lg bg-slate-900/40 border border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden ring-1 ring-white/10 backdrop-blur-xl"
			transition:modalScale
		>
			<!-- Animated gradient backgrounds -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				<!-- Primary gradient blob -->
				<div
					class={`gradient-blob blob-1
					${isCritical ? 'bg-gradient-to-br from-red-500/30 via-red-600/20 to-transparent' : 'bg-gradient-to-br from-blue-500/30 via-blue-600/20 to-transparent'}`}
				></div>

				<!-- Secondary gradient blob -->
				<div
					class={`gradient-blob blob-2
					${isCritical ? 'bg-gradient-to-tl from-red-400/25 via-red-500/15 to-transparent' : 'bg-gradient-to-tl from-blue-400/25 via-blue-500/15 to-transparent'}`}
				></div>

				<!-- Tertiary gradient blob -->
				<div
					class={`gradient-blob blob-3
					${isCritical ? 'bg-gradient-to-tr from-orange-600/20 via-red-400/10 to-transparent' : 'bg-gradient-to-tr from-cyan-600/20 via-blue-400/10 to-transparent'}`}
				></div>
			</div>

			<!-- Particles -->
			<div class="absolute inset-0 pointer-events-none overflow-hidden">
				{#each particles as particle (particle.id)}
					<div
						class={`particle
						${isCritical ? 'bg-red-400/60' : 'bg-blue-400/60'}`}
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
				class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r {isCritical
					? 'from-red-500 via-orange-500 to-red-500'
					: 'from-blue-500 via-cyan-500 to-blue-500'} opacity-80 shimmer-bar"
			></div>

			<!-- Content -->
			<div class="p-6 sm:p-8 relative z-10">
				<div class="flex items-start gap-4">
					<div
						class="icon-wrapper flex-shrink-0 p-3 rounded-full bg-slate-800/50 border border-slate-700/50 backdrop-blur-sm"
					>
						<div class="animate-icon-pop">
							{@html icon}
						</div>
					</div>
					<div class="flex-1">
						<h3 class="text-xl font-bold text-slate-100 mb-2 tracking-tight slide-in-text">
							{title}
						</h3>
						<div class="text-slate-400 text-sm leading-relaxed space-y-2 slide-in-text-delayed">
							{#if loading && statusMessage}
								<p class="animate-pulse text-slate-300 font-medium">{statusMessage}</p>
							{:else}
								<p>{message}</p>
							{/if}
						</div>
						<div class="slide-in-text-more-delayed">
							<slot />
						</div>
					</div>
				</div>

				{#if error}
					<div
						class="mt-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl flex items-start gap-3 error-bounce backdrop-blur-sm"
						transition:scale={{ start: 0.95, duration: 300 }}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5 shake-error"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"
							></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg
						>
						<div>
							<h4 class="text-sm font-bold text-red-400">Error</h4>
							<p class="text-red-300/80 text-xs mt-0.5">{error}</p>
						</div>
					</div>
				{/if}

				{#if loading && progress !== null}
					<!-- Progress Bar -->
					<div class="mt-8 space-y-2" transition:fade>
						<div
							class="flex justify-between text-xs font-semibold text-slate-400 uppercase tracking-wider"
						>
							<span>Progress</span>
							<span class="tabular-nums">{Math.round(progress)}%</span>
						</div>
						<div
							class="w-full h-2 bg-slate-800/80 rounded-full overflow-hidden border border-slate-700/50 backdrop-blur-sm"
						>
							<div
								class="h-full bg-gradient-to-r {isCritical
									? 'from-red-500 to-orange-500'
									: 'from-blue-500 to-cyan-500'} transition-all duration-300 ease-out relative progress-bar-glow"
								style="width: {progress}%"
							>
								<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
								<div class="absolute inset-0 shimmer-progress"></div>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Actions -->
			<div
				class="px-6 py-5 bg-slate-950/30 border-t border-slate-800/50 flex justify-end gap-3 relative z-10 backdrop-blur-sm"
			>
				{#if loading && progress !== null}
					<!-- Locked state during progress -->
					<button disabled class="px-4 py-2 text-sm font-medium text-slate-500 cursor-not-allowed">
						Please wait...
					</button>
				{:else if loading}
					<button
						disabled
						class="flex items-center gap-2 px-6 py-2 text-sm font-semibold text-slate-400 bg-slate-800/80 rounded-lg cursor-not-allowed backdrop-blur-sm"
					>
						<svg
							class="animate-spin h-4 w-4 text-slate-400"
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
						>
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
							></path>
						</svg>
						<span>Processing</span>
					</button>
				{:else}
					<button
						onclick={close}
						class="button-hover px-5 py-2 text-sm font-semibold text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 rounded-lg transition-all backdrop-blur-sm"
					>
						{cancelText}
					</button>
					<button
						onclick={handleConfirm}
						class={`button-confirm px-5 py-2 text-sm font-semibold text-white rounded-lg shadow-lg hover:shadow-xl transition-all transform active:scale-95 flex items-center gap-2 ${
							isCritical
								? 'bg-gradient-to-r from-red-600 to-red-500 hover:from-red-500 hover:to-red-400 shadow-red-900/20'
								: 'bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 shadow-blue-900/20'
						}`}
					>
						{confirmText}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="w-4 h-4 opacity-80 arrow-icon"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"><path d="M5 12h14"></path><path d="M12 5l7 7-7 7"></path></svg
						>
					</button>
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

	/* Backdrop animation */
	.backdrop-animate {
		animation: backdropFade 0.3s ease-out;
	}

	@keyframes backdropFade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	/* Animated gradient blobs */
	.gradient-blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(50px);
		opacity: 0.8;
	}

	.blob-1 {
		width: 300px;
		height: 300px;
		top: -100px;
		right: -100px;
		animation: float1 10s ease-in-out infinite;
	}

	.blob-2 {
		width: 250px;
		height: 250px;
		bottom: -80px;
		left: -80px;
		animation: float2 12s ease-in-out infinite;
	}

	.blob-3 {
		width: 200px;
		height: 200px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		animation: float3 8s ease-in-out infinite;
	}

	@keyframes float1 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(-30px, 30px) scale(1.1);
		}
		66% {
			transform: translate(20px, -20px) scale(0.9);
		}
	}

	@keyframes float2 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(25px, -35px) scale(0.95);
		}
		66% {
			transform: translate(-30px, 15px) scale(1.05);
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
			opacity: 1;
		}
		50% {
			transform: translateY(-40px) translateX(15px) scale(1);
			opacity: 0.8;
		}
		90% {
			opacity: 0.3;
		}
		100% {
			transform: translateY(-80px) translateX(-10px) scale(0);
			opacity: 0;
		}
	}

	/* Shimmer overlay animation */
	.gradient-overlay {
		background: linear-gradient(
			45deg,
			transparent 30%,
			rgba(255, 255, 255, 0.15) 50%,
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
		filter: drop-shadow(0 0 10px currentColor);
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
		animation: iconWrapperPulse 2s ease-in-out 0.6s infinite;
	}

	@keyframes iconWrapperPulse {
		0%,
		100% {
			box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.4);
		}
		50% {
			box-shadow: 0 0 20px 5px rgba(59, 130, 246, 0.1);
		}
	}

	/* Text slide-in animations */
	.slide-in-text {
		animation: slideInText 0.5s ease-out 0.2s both;
	}

	.slide-in-text-delayed {
		animation: slideInText 0.5s ease-out 0.3s both;
	}

	.slide-in-text-more-delayed {
		animation: slideInText 0.5s ease-out 0.4s both;
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

	.shake-error {
		animation: shakeError 0.5s ease-in-out;
	}

	@keyframes shakeError {
		0%,
		100% {
			transform: translateX(0) rotate(0deg);
		}
		25% {
			transform: translateX(-5px) rotate(-5deg);
		}
		75% {
			transform: translateX(5px) rotate(5deg);
		}
	}

	/* Progress bar glow */
	.progress-bar-glow {
		box-shadow: 0 0 20px currentColor;
		animation: progressGlow 2s ease-in-out infinite;
	}

	@keyframes progressGlow {
		0%,
		100% {
			filter: brightness(1);
		}
		50% {
			filter: brightness(1.4);
		}
	}

	.shimmer-progress {
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
		animation: shimmerProgress 1.5s ease-in-out infinite;
	}

	@keyframes shimmerProgress {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	/* Button animations */
	.button-hover {
		transition: all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	.button-hover:hover {
		transform: translateY(-2px);
	}

	.button-hover:active {
		transform: translateY(0);
	}

	.button-confirm {
		position: relative;
		overflow: hidden;
	}

	.button-confirm::before {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1), transparent);
		transform: translateX(-100%);
		transition: transform 0.6s;
	}

	.button-confirm:hover::before {
		transform: translateX(100%);
	}

	.button-confirm:hover {
		transform: translateY(-2px) scale(1.02);
	}

	.button-confirm:active {
		transform: translateY(0) scale(0.98);
	}

	.arrow-icon {
		transition: transform 0.3s ease;
	}

	.button-confirm:hover .arrow-icon {
		transform: translateX(3px);
	}
</style>
