<script lang="ts">
	import { X, Info, CheckCircle, AlertTriangle, XCircle } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { quintOut, backOut, elasticOut } from 'svelte/easing';
	import type { Notification } from '$lib/stores';

	export let notification: Notification;
	export let onDismiss: (id: string) => void;

	$: progress = 100;
	let timer: ReturnType<typeof setTimeout>;
	let interval: ReturnType<typeof setInterval>;
	let startTime: number;
	let isHovered = false;

	// Generate random particles
	const particles = Array.from({ length: 20 }, (_, i) => ({
		id: i,
		x: Math.random() * 100,
		y: Math.random() * 100,
		size: Math.random() * 4 + 2,
		duration: Math.random() * 3 + 2,
		delay: Math.random() * 2
	}));

	function startTimer() {
		if (notification.timeout && notification.timeout > 0) {
			startTime = Date.now();
			timer = setTimeout(() => onDismiss(notification.id), notification.timeout);
			interval = setInterval(() => {
				progress = 100 - ((Date.now() - startTime) / notification.timeout!) * 100;
				if (progress <= 0) clearInterval(interval);
			}, 50);
		}
	}

	function pauseTimer() {
		clearTimeout(timer);
		clearInterval(interval);
		isHovered = true;
	}

	function resumeTimer() {
		isHovered = false;
		startTimer();
	}

	$: {
		if (notification && notification.timeout) {
			startTimer();
		}
	}

	// Clean up on component destroy
	import { onDestroy } from 'svelte';
	onDestroy(() => {
		clearTimeout(timer);
		clearInterval(interval);
	});

	// Custom slide in from right with bounce
	function slideInBounce(node: HTMLElement, params: { duration?: number } = {}) {
		const { duration = 600 } = params;
		return {
			duration,
			css: (t: number) => {
				const eased = backOut(t);
				return `
					transform: translateX(${(1 - eased) * 100}%) scale(${0.8 + eased * 0.2});
					opacity: ${eased};
				`;
			}
		};
	}

	// Custom exit animation
	function slideOutScale(node: HTMLElement, params: { duration?: number } = {}) {
		const { duration = 300 } = params;
		return {
			duration,
			css: (t: number) => {
				const eased = quintOut(1 - t);
				return `
					transform: translateX(${eased * 50}%) scale(${t * 0.9});
					opacity: ${t};
				`;
			}
		};
	}
</script>

<div
	class={`notification-card relative w-full max-w-sm rounded-xl overflow-hidden shadow-2xl p-4 flex items-start gap-3 transition-all duration-300 border backdrop-blur-md
	${notification.type === 'success' ? 'bg-emerald-900/40 border-emerald-400/50 shadow-emerald-500/20' : ''}
	${notification.type === 'error' ? 'bg-red-900/40 border-red-400/50 shadow-red-500/20' : ''}
	${notification.type === 'info' ? 'bg-rust/40 border-rust-light/50 shadow-rust-light/20' : ''}
	${notification.type === 'warning' ? 'bg-orange-900/40 border-orange-400/50 shadow-orange-500/20' : ''}
	${isHovered ? 'scale-105 shadow-2xl' : 'scale-100'}`}
	onmouseover={pauseTimer}
	onmouseout={resumeTimer}
	onfocus={pauseTimer}
	onblur={resumeTimer}
	role="alertdialog"
	tabindex="-1"
	in:slideInBounce
	out:slideOutScale
>
	<!-- Animated gradient backgrounds -->
	<div class="absolute inset-0 pointer-events-none overflow-hidden">
		<!-- Primary gradient blob -->
		<div
			class={`gradient-blob blob-1
			${notification.type === 'success' ? 'bg-gradient-to-br from-emerald-500/30 via-emerald-600/20 to-transparent' : ''}
			${notification.type === 'error' ? 'bg-gradient-to-br from-red-500/30 via-red-600/20 to-transparent' : ''}
			${notification.type === 'info' ? 'bg-gradient-to-br from-blue-500/30 via-blue-600/20 to-transparent' : ''}
			${notification.type === 'warning' ? 'bg-gradient-to-br from-orange-500/30 via-orange-600/20 to-transparent' : ''}`}
		></div>

		<!-- Secondary gradient blob -->
		<div
			class={`gradient-blob blob-2
			${notification.type === 'success' ? 'bg-gradient-to-tl from-emerald-400/25 via-emerald-500/15 to-transparent' : ''}
			${notification.type === 'error' ? 'bg-gradient-to-tl from-red-400/25 via-red-500/15 to-transparent' : ''}
			${notification.type === 'info' ? 'bg-gradient-to-tl from-blue-400/25 via-blue-500/15 to-transparent' : ''}
			${notification.type === 'warning' ? 'bg-gradient-to-tl from-orange-400/25 via-orange-500/15 to-transparent' : ''}`}
		></div>

		<!-- Tertiary gradient blob -->
		<div
			class={`gradient-blob blob-3
			${notification.type === 'success' ? 'bg-gradient-to-tr from-emerald-600/20 via-emerald-400/10 to-transparent' : ''}
			${notification.type === 'error' ? 'bg-gradient-to-tr from-red-600/20 via-red-400/10 to-transparent' : ''}
			${notification.type === 'info' ? 'bg-gradient-to-tr from-blue-600/20 via-blue-400/10 to-transparent' : ''}
			${notification.type === 'warning' ? 'bg-gradient-to-tr from-orange-600/20 via-orange-400/10 to-transparent' : ''}`}
		></div>
	</div>

	<!-- Particles -->
	<div class="absolute inset-0 pointer-events-none overflow-hidden">
		{#each particles as particle (particle.id)}
			<div
				class={`particle
				${notification.type === 'success' ? 'bg-emerald-400/60' : ''}
				${notification.type === 'error' ? 'bg-red-400/60' : ''}
				${notification.type === 'info' ? 'bg-blue-400/60' : ''}
				${notification.type === 'warning' ? 'bg-orange-400/60' : ''}`}
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
	<div class="absolute inset-0 opacity-20 pointer-events-none gradient-overlay"></div>

	<!-- Icon with pulse animation -->
	<div class="flex-shrink-0 icon-container relative z-10">
		{#if notification.type === 'success'}
			<CheckCircle class="w-6 h-6 text-emerald-300 animate-icon" />
		{:else if notification.type === 'error'}
			<XCircle class="w-6 h-6 text-red-300 animate-icon shake" />
		{:else if notification.type === 'info'}
			<Info class="w-6 h-6 text-blue-300 animate-icon" />
		{:else if notification.type === 'warning'}
			<AlertTriangle class="w-6 h-6 text-orange-300 animate-icon pulse-warning" />
		{/if}
	</div>

	<!-- Content -->
	<div class="flex-1 relative z-10">
		<p class="font-semibold text-slate-900 dark:text-white text-sm slide-in-text drop-shadow-lg">
			{notification.message}
		</p>
		{#if notification.details}
			<p class="text-xs text-slate-800 dark:text-slate-200 mt-1 slide-in-text-delayed drop-shadow">
				{notification.details}
			</p>
		{/if}
	</div>

	<!-- Dismiss Button with hover effect -->
	<button
		onclick={() => onDismiss(notification.id)}
		class="dismiss-button flex-shrink-0 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:text-white transition-all duration-200 p-1 rounded-full hover:bg-white/20 hover:rotate-90 relative z-10"
	>
		<X class="w-4 h-4" />
	</button>

	<!-- Animated Progress Bar -->
	{#if notification.timeout && notification.timeout > 0}
		<div class="absolute bottom-0 left-0 right-0 h-1 bg-black/30">
			<div
				class={`h-full transition-all duration-100 progress-bar
				${notification.type === 'success' ? 'bg-gradient-to-r from-emerald-400 to-emerald-300' : ''}
				${notification.type === 'error' ? 'bg-gradient-to-r from-red-400 to-red-300' : ''}
				${notification.type === 'info' ? 'bg-gradient-to-r from-blue-400 to-blue-300' : ''}
				${notification.type === 'warning' ? 'bg-gradient-to-r from-orange-400 to-orange-300' : ''}`}
				style="width: {progress}%;"
			></div>
		</div>
	{/if}
</div>

<style>
	.notification-card {
		transform-origin: right center;
		will-change: transform, opacity;
	}

	/* Animated gradient blobs */
	.gradient-blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(40px);
		opacity: 0.8;
	}

	.blob-1 {
		width: 200px;
		height: 200px;
		top: -50px;
		right: -50px;
		animation: float1 8s ease-in-out infinite;
	}

	.blob-2 {
		width: 180px;
		height: 180px;
		bottom: -40px;
		left: -40px;
		animation: float2 10s ease-in-out infinite;
	}

	.blob-3 {
		width: 150px;
		height: 150px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		animation: float3 7s ease-in-out infinite;
	}

	@keyframes float1 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(-20px, 20px) scale(1.1);
		}
		66% {
			transform: translate(10px, -15px) scale(0.9);
		}
	}

	@keyframes float2 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(15px, -25px) scale(0.95);
		}
		66% {
			transform: translate(-20px, 10px) scale(1.05);
		}
	}

	@keyframes float3 {
		0%,
		100% {
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			transform: translate(-50%, -50%) scale(1.15);
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
			transform: translateY(-30px) translateX(10px) scale(1);
			opacity: 0.8;
		}
		90% {
			opacity: 0.3;
		}
		100% {
			transform: translateY(-60px) translateX(-5px) scale(0);
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
		animation: shimmer 3s ease-in-out infinite;
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

	/* Icon animations */
	.animate-icon {
		animation: iconPop 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
		filter: drop-shadow(0 0 8px currentColor);
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

	.shake {
		animation:
			iconPop 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55),
			shake 0.5s ease-in-out 0.5s;
	}

	@keyframes shake {
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

	.pulse-warning {
		animation:
			iconPop 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55),
			pulse 2s ease-in-out 0.5s infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			transform: scale(1);
			opacity: 1;
		}
		50% {
			transform: scale(1.1);
			opacity: 0.8;
		}
	}

	/* Text slide-in animations */
	.slide-in-text {
		animation: slideInText 0.4s ease-out 0.2s both;
	}

	.slide-in-text-delayed {
		animation: slideInText 0.4s ease-out 0.3s both;
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

	/* Progress bar animation */
	.progress-bar {
		box-shadow: 0 0 10px currentColor;
		animation: progressGlow 2s ease-in-out infinite;
	}

	@keyframes progressGlow {
		0%,
		100% {
			filter: brightness(1);
		}
		50% {
			filter: brightness(1.3);
		}
	}

	/* Dismiss button animation */
	.dismiss-button {
		transition: all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
	}

	.dismiss-button:hover {
		transform: rotate(90deg) scale(1.1);
	}

	.dismiss-button:active {
		transform: rotate(90deg) scale(0.9);
	}

	/* Hover state for entire card */
	.notification-card:hover {
		transform: scale(1.02) translateX(-5px);
	}

	.notification-card:hover .gradient-overlay {
		animation-duration: 1.5s;
	}

	.notification-card:hover .gradient-blob {
		animation-duration: 4s;
	}

	.notification-card:hover .particle {
		animation-duration: 1.5s;
	}
</style>
