<script lang="ts">
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';

	let canvas: HTMLCanvasElement;
	let context: CanvasRenderingContext2D | null;
	let width = 0;
	let height = 0;
	let animationFrameId: number;

	interface Path {
		points: { x: number; y: number }[];
		length: number;
	}

	interface Electron {
		pathIndex: number;
		distance: number;
		speed: number;
		size: number;
		color: string;
	}

	let paths: Path[] = [];
	let electrons: Electron[] = [];

	function resize() {
		if (canvas) {
			width = window.innerWidth;
			height = window.innerHeight;
			canvas.width = width;
			canvas.height = height;
			generatePaths();
		}
	}

	function generatePaths() {
		paths = [];
		const gridSize = 60;
		const cols = Math.ceil(width / gridSize);
		const rows = Math.ceil(height / gridSize);

		// Generate random circuit paths
		for (let i = 0; i < 40; i++) {
			let currentX = Math.floor(Math.random() * cols) * gridSize;
			let currentY = Math.floor(Math.random() * rows) * gridSize;
			const points = [{ x: currentX, y: currentY }];
			let length = 0;

			const segments = Math.floor(Math.random() * 5) + 3;
			for (let j = 0; j < segments; j++) {
				const direction = Math.floor(Math.random() * 4); // 0: right, 1: down, 2: left, 3: up
				let nextX = currentX;
				let nextY = currentY;

				// Try to move in a cardinal direction
				const moveDist = gridSize * (Math.floor(Math.random() * 3) + 1);

				if (direction === 0) nextX += moveDist;
				else if (direction === 1) nextY += moveDist;
				else if (direction === 2) nextX -= moveDist;
				else if (direction === 3) nextY -= moveDist;

				// Keep within bounds
				nextX = Math.max(0, Math.min(width, nextX));
				nextY = Math.max(0, Math.min(height, nextY));

				if (nextX !== currentX || nextY !== currentY) {
					length += Math.hypot(nextX - currentX, nextY - currentY);
					points.push({ x: nextX, y: nextY });
					currentX = nextX;
					currentY = nextY;
				}
			}
			if (points.length > 1) {
				paths.push({ points, length });
			}
		}
	}

	function spawnElectron() {
		if (paths.length === 0) return;
		const pathIndex = Math.floor(Math.random() * paths.length);
		electrons.push({
			pathIndex,
			distance: 0,
			speed: 2 + Math.random() * 3,
			size: 1.5 + Math.random() * 1.5,
			color: Math.random() > 0.8 ? '#f59e0b' : '#3b82f6' // Blue dominant, occasional amber spark
		});
	}

	function animate() {
		if (!context) return;
		context.clearRect(0, 0, width, height);

		// Draw Paths (Subtle trace)
		context.strokeStyle = 'rgba(59, 130, 246, 0.05)';
		context.lineWidth = 1;
		
		paths.forEach(path => {
			context!.beginPath();
			context!.moveTo(path.points[0].x, path.points[0].y);
			for (let i = 1; i < path.points.length; i++) {
				context!.lineTo(path.points[i].x, path.points[i].y);
			}
			context!.stroke();
			
			// Draw nodes at junctions
			path.points.forEach(p => {
				context!.fillStyle = 'rgba(59, 130, 246, 0.1)';
				context!.beginPath();
				context!.arc(p.x, p.y, 2, 0, Math.PI * 2);
				context!.fill();
			});
		});

		// Spawn new electrons randomly
		if (Math.random() < 0.05) spawnElectron();

		// Update and Draw Electrons
		for (let i = electrons.length - 1; i >= 0; i--) {
			const e = electrons[i];
			e.distance += e.speed;
			const path = paths[e.pathIndex];

			if (e.distance >= path.length) {
				electrons.splice(i, 1);
				continue;
			}

			// Calculate position along path
			let currentDist = 0;
			let pos = { x: 0, y: 0 };
			
			for (let j = 0; j < path.points.length - 1; j++) {
				const p1 = path.points[j];
				const p2 = path.points[j + 1];
				const segLen = Math.hypot(p2.x - p1.x, p2.y - p1.y);
				
				if (currentDist + segLen >= e.distance) {
					const remaining = e.distance - currentDist;
					const ratio = remaining / segLen;
					pos.x = p1.x + (p2.x - p1.x) * ratio;
					pos.y = p1.y + (p2.y - p1.y) * ratio;
					break;
				}
				currentDist += segLen;
			}

			// Draw glow
			context.shadowBlur = 10;
			context.shadowColor = e.color;
			context.fillStyle = e.color;
			context.beginPath();
			context.arc(pos.x, pos.y, e.size, 0, Math.PI * 2);
			context.fill();
			context.shadowBlur = 0;
		}

		animationFrameId = requestAnimationFrame(animate);
	}

	onMount(() => {
		context = canvas.getContext('2d');
		resize();
		window.addEventListener('resize', resize);
		animate();

		return () => {
			window.removeEventListener('resize', resize);
			cancelAnimationFrame(animationFrameId);
		};
	});
</script>

<div class="fixed inset-0 z-[-1] overflow-hidden pointer-events-none bg-[#0f172a]" in:fade={{ duration: 1000 }}>
	<!-- Subtle grid overlay -->
	<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.03]"></div>
	
	<!-- Vignette -->
	<div class="absolute inset-0 bg-[radial-gradient(circle_at_center,transparent_0%,#0f172a_100%)] opacity-80"></div>

	<canvas bind:this={canvas} class="absolute inset-0 block"></canvas>
</div>