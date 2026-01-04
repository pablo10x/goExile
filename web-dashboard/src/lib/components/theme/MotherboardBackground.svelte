<script lang="ts">
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { lowPowerMode } from '$lib/stores.svelte';

	let canvas: HTMLCanvasElement;
	let context: CanvasRenderingContext2D | null;
	let width = 0;
	let height = 0;
	let animationFrameId: number;
	let isVisible = true;

	interface Particle {
		x: number;
		y: number;
		size: number;
		vx: number;
		vy: number;
		opacity: number;
	}

	let particles: Particle[] = [];
	const particleCount = 45;
	const connectionDistance = 200;
	const connectionDistanceSq = connectionDistance * connectionDistance;

	function resize() {
		if (canvas) {
			width = window.innerWidth;
			height = window.innerHeight;
			canvas.width = width;
			canvas.height = height;
			initParticles();
		}
	}

	function initParticles() {
		particles = [];
		for (let i = 0; i < particleCount; i++) {
			particles.push({
				x: Math.random() * width,
				y: Math.random() * height,
				size: Math.random() * 1.5 + 0.5,
				vx: (Math.random() - 0.5) * 0.25,
				vy: (Math.random() - 0.5) * 0.25,
				opacity: Math.random() * 0.5 + 0.2
			});
		}
	}

	let lastTime = 0;
	const targetFPS = 30;
	const frameInterval = 1000 / targetFPS;

	function animate(time: number) {
		animationFrameId = requestAnimationFrame(animate);

		if (!isVisible || $lowPowerMode) return;

		const deltaTime = time - lastTime;
		if (deltaTime < frameInterval) return;
		lastTime = time - (deltaTime % frameInterval);

		if (!context) return;
		context.clearRect(0, 0, width, height);

		context.fillStyle = 'rgba(194, 65, 12, 0.15)';
		context.lineWidth = 0.5;

		// Update and draw particles
		for (let i = 0; i < particles.length; i++) {
			const p = particles[i];
			p.x += p.vx;
			p.y += p.vy;

			// Wrap around edges
			if (p.x < 0) p.x = width;
			if (p.x > width) p.x = 0;
			if (p.y < 0) p.y = height;
			if (p.y > height) p.y = 0;

			context.beginPath();
			context.arc(p.x, p.y, p.size, 0, Math.PI * 2);
			context.fill();

			// Connections - capped per particle for performance
			let connections = 0;
			for (let j = i + 1; j < particles.length && connections < 5; j++) {
				const p2 = particles[j];
				const dx = p.x - p2.x;
				const dy = p.y - p2.y;
				const distSq = dx * dx + dy * dy;

				if (distSq < connectionDistanceSq) {
					const dist = Math.sqrt(distSq);
					const opacity = (1 - dist / connectionDistance) * 0.12;
					context.strokeStyle = `rgba(194, 65, 12, ${opacity})`;
					context.beginPath();
					context.moveTo(p.x, p.y);
					context.lineTo(p2.x, p2.y);
					context.stroke();
					connections++;
				}
			}
		}
	}

	onMount(() => {
		context = canvas.getContext('2d', { alpha: true });
		resize();
		
		const handleVisibility = () => {
			isVisible = document.visibilityState === 'visible';
		};
		document.addEventListener('visibilitychange', handleVisibility);
		window.addEventListener('resize', resize);
		
		animationFrameId = requestAnimationFrame(animate);

		return () => {
			document.removeEventListener('visibilitychange', handleVisibility);
			window.removeEventListener('resize', resize);
			cancelAnimationFrame(animationFrameId);
		};
	});
</script>

<div class="fixed inset-0 z-[-1] overflow-hidden pointer-events-none bg-[#0a0a0a]" in:fade={{ duration: 1000 }}>
	<!-- Soft ambient glows -->
	<div class="absolute top-[-10%] -left-[10%] w-[40%] h-[40%] bg-rust/5 blur-[120px] rounded-full animate-pulse"></div>
	<div class="absolute bottom-[-10%] -right-[10%] w-[40%] h-[40%] bg-neutral-800/5 blur-[120px] rounded-full animate-pulse" style="animation-delay: 2s;"></div>
	
	<!-- Subtle grid overlay -->
	<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.04]"></div>
	
	<canvas bind:this={canvas} class="absolute inset-0 block"></canvas>
</div>
