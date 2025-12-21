<script lang="ts">
	import { onMount } from 'svelte';
	import { spawners } from '$lib/stores';
	import { Server, Activity, Cpu, Skull, Database, User } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';

	// Animation state
	let center = $state({ x: 300, y: 300 }); // Master node position
	let spawnerRadius = $state(250); // Radius for spawner placement around the master

	// RedEye and Database node positions (relative to master)
	let redeyePos = $derived({ x: center.x - 220, y: center.y + 120 });
	let databasePos = $derived({ x: center.x + 220, y: center.y + 120 });

	// Track heartbeats for pulse animation
	let lastHeartbeats: Record<number, number> = $state({});
	let pulsingSpawners: Set<number> = $state(new Set());
	let masterReceiving = $state(false); // Track when master receives spark
	let masterIgniteTimeout: ReturnType<typeof setTimeout> | null = null;

	// Track spawner removal animations
	let removingSpawners: Set<number> = $state(new Set());
	let previousSpawners: Map<number, any> = new Map();

	let hoveredSpawnerId: number | null = $state(null);
	let containerElement: HTMLDivElement;
	let svgDimensions = $state({ width: 600, height: 600 });

	// Zoom state
	let zoom = $state(1);
	let panOffset = $state({ x: 0, y: 0 });
	const MIN_ZOOM = 0.5;
	const MAX_ZOOM = 2.5;

	// Zoom Control (only double-click enables)
	let canZoom = $state(false);

	// Drag state
	let isDragging = $state(false);
	let dragStart = $state({ x: 0, y: 0 });
	let lastPanOffset = $state({ x: 0, y: 0 });

	function enableZoomImmediately() {
		canZoom = true;
		if (containerElement && !isDragging) containerElement.style.cursor = 'grab';
	}

	// Background particles
	const backgroundParticles = Array.from({ length: 30 }, (_, i) => ({
		id: i,
		x: Math.random() * 100,
		y: Math.random() * 100,
		size: Math.random() * 3 + 1,
		duration: Math.random() * 4 + 3,
		delay: Math.random() * 3
	}));

	// RedEye Interception State
	let showInterception = $state(false);
	let interceptionSeed = $state(0); // Used to trigger re-calculation of random values
	let interceptionInterval: ReturnType<typeof setInterval>;

	// Derived interception data to ensure everything stays in sync with Master center
	let interceptionData = $derived.by(() => {
		const s = interceptionSeed;
		const offset = 140;
		const rx = ((Math.sin(s * 123.45) + 1) / 2 - 0.5) * offset * 2;
		const ry = ((Math.cos(s * 678.9) + 1) / 2 - 0.5) * offset * 2;

		const tx = center.x + rx;
		const ty = center.y + ry;

		const midX1 = (redeyePos.x + tx) / 2;
		const midY1 = (redeyePos.y + ty) / 2;
		const dx = tx - redeyePos.x;
		const dy = ty - redeyePos.y;
		const dist = Math.sqrt(dx * dx + dy * dy);
		const px = -dy / dist;
		const py = dx / dist;

		// Missile paths
		const path1 = `M${redeyePos.x},${redeyePos.y} Q${midX1 + px * 80},${midY1 + py * 80} ${tx},${ty}`;
		const path2 = `M${redeyePos.x},${redeyePos.y} Q${midX1 - px * 60},${midY1 - py * 60} ${tx},${ty}`;

		return {
			target: { x: tx, y: ty },
			path1,
			path2,
			angle: Math.atan2(dy, dx) * (180 / Math.PI)
		};
	});

	onMount(() => {
		interceptionInterval = setInterval(() => {
			interceptionSeed = Math.random();
			showInterception = true;
			setTimeout(() => (showInterception = false), 4000);
		}, 8000);

		if (containerElement) {
			const updateDimensions = () => {
				const rect = containerElement.getBoundingClientRect();
				svgDimensions = { width: rect.width, height: rect.height };
				center = { x: rect.width / 2, y: rect.height / 2 };
				spawnerRadius = Math.min(rect.width, rect.height) * 0.35;
			};

			updateDimensions();
			const resizeObserver = new ResizeObserver(updateDimensions);
			resizeObserver.observe(containerElement);

			// Add wheel event listener for zoom
			const handleWheel = (e: WheelEvent) => {
				if (!canZoom) return; // Prevent zoom if not enabled

				e.preventDefault();

				// Zoom factor
				const zoomSensitivity = 0.001;
				const delta = -e.deltaY * zoomSensitivity;
				const newZoom = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, zoom + delta));

				zoom = newZoom;
			};

			// Add drag event handlers
			const handleMouseDown = (e: MouseEvent) => {
				isDragging = true;
				dragStart = { x: e.clientX, y: e.clientY };
				lastPanOffset = { ...panOffset };
				containerElement.style.cursor = 'grabbing';
			};

			const handleMouseMove = (e: MouseEvent) => {
				if (!isDragging) return;

				const deltaX = e.clientX - dragStart.x;
				const deltaY = e.clientY - dragStart.y;

				panOffset = {
					x: lastPanOffset.x + deltaX,
					y: lastPanOffset.y + deltaY
				};
			};

			const handleMouseUp = () => {
				isDragging = false;
				containerElement.style.cursor = canZoom ? 'grab' : 'default';
			};

			const handleMouseLeave = () => {
				canZoom = false; // Disable zoom when mouse leaves
				if (isDragging) {
					isDragging = false;
				}
				if (containerElement) containerElement.style.cursor = 'default';
			};

			const handleDblClick = () => {
				enableZoomImmediately();
			};

			containerElement.addEventListener('wheel', handleWheel, { passive: false });
			containerElement.addEventListener('mousedown', handleMouseDown);
			containerElement.addEventListener('mouseleave', handleMouseLeave);
			containerElement.addEventListener('dblclick', handleDblClick);
			document.addEventListener('mousemove', handleMouseMove);
			document.addEventListener('mouseup', handleMouseUp);

			return () => {
				resizeObserver.disconnect();
				if (containerElement) {
					containerElement.removeEventListener('wheel', handleWheel);
					containerElement.removeEventListener('mousedown', handleMouseDown);
					containerElement.removeEventListener('mouseleave', handleMouseLeave);
					containerElement.removeEventListener('dblclick', handleDblClick);
				}
				document.removeEventListener('mousemove', handleMouseMove);
				document.removeEventListener('mouseup', handleMouseUp);
			};
		}
	});

	// Watch for spawner removal
	$effect(() => {
		const currentIds = new Set($spawners.map((s) => s.id));

		// Check for removed spawners
		previousSpawners.forEach((spawner, id) => {
			if (!currentIds.has(id) && !removingSpawners.has(id)) {
				// Spawner was removed, trigger shatter animation
				removingSpawners.add(id);
				removingSpawners = new Set(removingSpawners);

				// Clean up after animation completes
				setTimeout(() => {
					removingSpawners.delete(id);
					removingSpawners = new Set(removingSpawners);
					previousSpawners.delete(id);
				}, 1500);
			}
		});

		// Update previous spawners map
		$spawners.forEach((s) => {
			previousSpawners.set(s.id, { ...s });
		});
	});

	// Watch for heartbeat changes
	$effect(() => {
		$spawners.forEach((s) => {
			const lastTime = new Date(s.last_seen || 0).getTime();
			if (lastHeartbeats[s.id] && lastTime > lastHeartbeats[s.id]) {
				// New heartbeat detected
				triggerPulse(s.id);
			}
			lastHeartbeats[s.id] = lastTime;
		});
	});

	function triggerPulse(id: number) {
		pulsingSpawners.add(id);
		pulsingSpawners = new Set(pulsingSpawners);

		// Trigger master ignition after spark travel time (2.5s)
		setTimeout(() => {
			triggerMasterIgnite();
		}, 2500);

		setTimeout(() => {
			pulsingSpawners.delete(id);
			pulsingSpawners = new Set(pulsingSpawners);
		}, 1000);
	}

	function triggerMasterIgnite() {
		masterReceiving = true;

		// Clear any existing timeout
		if (masterIgniteTimeout) {
			clearTimeout(masterIgniteTimeout);
		}

		// Reset after animation completes
		masterIgniteTimeout = setTimeout(() => {
			masterReceiving = false;
		}, 800);
	}

	function getPosition(index: number, total: number) {
		if (total === 0) return { x: center.x + spawnerRadius, y: center.y };

		// Arrange spawners in a circle around the master
		// Offset by -Math.PI / 2 to start at the top
		const angle = (index / total) * 2 * Math.PI - Math.PI / 2;

		return {
			x: center.x + spawnerRadius * Math.cos(angle),
			y: center.y + spawnerRadius * Math.sin(angle)
		};
	}

	function getConnectionPath(startX: number, startY: number, endX: number, endY: number) {
		// Calculate control point for a slight curve
		const midX = (startX + endX) / 2;
		const midY = (startY + endY) / 2;

		// Add slight perpendicular offset for curve
		const dx = endX - startX;
		const dy = endY - startY;
		const distance = Math.sqrt(dx * dx + dy * dy);
		const perpX = -dy / distance;
		const perpY = dx / distance;
		const curveOffset = 15;

		const controlX = midX + perpX * curveOffset;
		const controlY = midY + perpY * curveOffset;

		return `M${startX},${startY} Q${controlX},${controlY} ${endX},${endY}`;
	}

	function getBusPath(startX: number, startY: number, endX: number, endY: number) {
		// Create a right-angled "circuit board" style path
		// Horizontal then Vertical
		const midY = (startY + endY) / 2;
		return `M${startX},${startY} L${startX},${midY} L${endX},${midY} L${endX},${endY}`;
	}

	function getStatusColor(spawner: any, isHovered: boolean) {
		const isActive = spawner.status === 'online' || spawner.status === 'Online';

		if (!isActive) return { stroke: '#ef4444', glow: 'rgba(239, 68, 68, 0.3)' };
		if (isHovered) return { stroke: '#38bdf8', glow: 'rgba(56, 189, 248, 0.5)' };
		return { stroke: '#64748b', glow: 'rgba(100, 116, 139, 0.2)' };
	}

	// Generate random shatter pieces for a spawner
	function generateShatterPieces(spawnerIndex: number) {
		const pieceCount = 12;
		return Array.from({ length: pieceCount }, (_, i) => {
			const angle = (i / pieceCount) * Math.PI * 2;
			const distance = 40 + Math.random() * 60;
			const rotation = Math.random() * 360;
			const size = 4 + Math.random() * 2;

			return {
				id: i,
				dx: Math.cos(angle) * distance,
				dy: Math.sin(angle) * distance,
				rotation,
				size,
				delay: Math.random() * 0.1
			};
		});
	}
</script>

<div
	bind:this={containerElement}
	class="flex absolute w-full h-full bg-slate-250 rounded-2xl overflow-hidden items-center justify-center shadow-2xl transition-colors duration-300"
>
	<!-- Animated gradient blobs in background -->

	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<!-- Primary gradient blob -->
		<div
			class="gradient-blob blob-1 bg-gradient-to-br from-blue-30000/15 via-cyan-600/10 to-transparent"
		></div>

		<!-- Secondary gradient blob -->
		<div
			class="gradient-blob blob-2 bg-gradient-to-tl from-purple-600/12 via-blue-600/8 to-transparent"
		></div>

		<!-- Tertiary gradient blob -->
		<div
			class="gradient-blob blob-3 bg-gradient-to-tr from-cyan-600/10 via-blue-500/8 to-transparent"
		></div>

		<!-- Quaternary blob for more depth -->
		<div
			class="gradient-blob blob-4 bg-gradient-to-bl from-indigo-600/12 via-blue-700/6 to-transparent"
		></div>
	</div>

	<!-- Background particles -->
	<div class="absolute inset-0 pointer-events-none overflow-hidden">
		{#each backgroundParticles as particle (particle.id)}
			<div
				class="background-particle bg-blue-400/40"
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

	<!-- Animated Grid Background -->
	<div
		class="absolute inset-0 opacity-80"
		style="background-image:
			linear-gradient(rgba(56, 189, 248, 0.03) 1px, transparent 1px),
			linear-gradient(90deg, rgba(56, 189, 248, 0.03) 1px, transparent 1px);
			background-size: 50px 50px;
			animation: gridMove 5s linear infinite; "
	></div>

	<!-- Radial gradient overlay for depth -->
	<div
		class="absolute inset-0 bg-radial-gradient opacity-10"
		style="background: radial-gradient(circle at 90% 50%, transparent 0%, rgba(0, 0, 0, 0.4) 100%);"
	></div>

	<!-- Shimmer overlay -->

	<!-- Zoomable and pannable content wrapper -->
	<div
		class="absolute inset-0 transition-transform duration-600 ease-out pointer-events-none"
		style="transform: scale({zoom}) translate({panOffset.x / zoom}px, {panOffset.y /
			zoom}px); transform-origin: center center;"
	>
		<div class="pointer-events-auto">
			<svg
				class="w-full h-full pointer-events-none absolute inset-0"
				viewBox="0 0 {svgDimensions.width} {svgDimensions.height}"
			>
				<defs>
					<!-- Enhanced glow filter -->
					<filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
						<feGaussianBlur stdDeviation="4" result="blur" />
						<feComposite in="SourceGraphic" in2="blur" operator="over" />
					</filter>

					<!-- Stronger glow for pulses -->
					<filter id="strongGlow" x="-100%" y="-100%" width="300%" height="300%">
						<feGaussianBlur stdDeviation="8" result="blur1" />
						<feGaussianBlur stdDeviation="4" result="blur2" />
						<feMerge>
							<feMergeNode in="blur1" />
							<feMergeNode in="blur2" />
							<feMergeNode in="SourceGraphic" />
						</feMerge>
					</filter>

					<marker
						id="arrow"
						markerWidth="10"
						markerHeight="10"
						refX="8"
						refY="3"
						orient="auto"
						markerUnits="strokeWidth"
					>
						<path d="M0,0 L0,6 L9,3 z" fill="#475569" />
					</marker>

					<!-- Gradient for connections -->
					<linearGradient id="connectionGradient" x1="0%" y1="0%" x2="100%" y2="0%">
						<stop offset="0%" style="stop-color:#64748b;stop-opacity:0.3" />
						<stop offset="50%" style="stop-color:#64748b;stop-opacity:0.6" />
						<stop offset="100%" style="stop-color:#64748b;stop-opacity:0.3" />
					</linearGradient>

					<!-- Missile Gradients -->
					<linearGradient id="missileGradient1" x1="0%" y1="0%" x2="100%" y2="0%">
						<stop offset="0%" stop-color="#ef4444" stop-opacity="0" />
						<stop offset="100%" stop-color="#fb923c" stop-opacity="1" />
					</linearGradient>
					<linearGradient id="missileGradient2" x1="0%" y1="0%" x2="100%" y2="0%">
						<stop offset="0%" stop-color="#ef4444" stop-opacity="0" />
						<stop offset="100%" stop-color="#ec4899" stop-opacity="1" />
					</linearGradient>

					<linearGradient id="masterBorderGradient" x1="0%" y1="0%" x2="0%" y2="100%">
						<stop offset="0%" class="master-border-top" />
						<stop offset="100%" class="master-border-bottom" />
					</linearGradient>

					<linearGradient id="dataStreamGradient" x1="0%" y1="0%" x2="0%" y2="100%">
						<stop offset="0%" stop-color="#10b981" stop-opacity="0" />
						<stop offset="50%" stop-color="#34d399" stop-opacity="1" />
						<stop offset="100%" stop-color="#10b981" stop-opacity="0" />
					</linearGradient>

					<filter id="masterGlow">
						<feGaussianBlur stdDeviation="3" result="blur" />
						<feComposite in="SourceGraphic" in2="blur" operator="over" />
					</filter>
				</defs>

				<!-- RedEye Connection -->
				<path
					d={getBusPath(redeyePos.x, redeyePos.y, center.x, center.y)}
					stroke="#ef4444"
					stroke-width="3"
					opacity="0.1"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<path
					d={getBusPath(redeyePos.x, redeyePos.y, center.x, center.y)}
					stroke="#ef4444"
					stroke-width="2"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
					class="animate-energy-surge"
				/>
				<path
					d={getBusPath(redeyePos.x, redeyePos.y, center.x, center.y)}
					stroke="#ff0000"
					stroke-width="1"
					opacity="0.5"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
					class="animate-redeye-flicker"
				/>

				<!-- Database Connection -->
				<path
					d={getBusPath(databasePos.x, databasePos.y, center.x, center.y)}
					stroke="#10b981"
					stroke-width="3"
					opacity="0.4"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<path
					d={getBusPath(databasePos.x, databasePos.y, center.x, center.y)}
					stroke="#10b981"
					stroke-width="1"
					opacity="0.8"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<animate
						attributeName="stroke-dasharray"
						values="0, 100; 100, 0"
						dur="4s"
						repeatCount="indefinite"
					/>
				</path>

				<!-- RedEye Interception (Global SVG) -->
				{#if showInterception}
					<g filter="url(#strongGlow)" style="pointer-events: none;">
						<!-- Threat Reticle -->
						<g
							class="animate-threat-appear"
							style="transform-box: fill-box; transform-origin: center;"
						>
							<path
								d="M {interceptionData.target.x - 20} {interceptionData.target.y - 10} L {interceptionData.target.x - 20} {interceptionData.target.y - 20} L {interceptionData.target.x - 10} {interceptionData.target.y - 20}"
								fill="none"
								stroke="#ef4444"
								stroke-width="2"
							/>
							<path
								d="M {interceptionData.target.x + 10} {interceptionData.target.y - 20} L {interceptionData.target.x + 20} {interceptionData.target.y - 20} L {interceptionData.target.x + 20} {interceptionData.target.y - 10}"
								fill="none"
								stroke="#ef4444"
								stroke-width="2"
							/>
							<path
								d="M {interceptionData.target.x + 20} {interceptionData.target.y + 10} L {interceptionData.target.x + 20} {interceptionData.target.y + 20} L {interceptionData.target.x + 10} {interceptionData.target.y + 20}"
								fill="none"
								stroke="#ef4444"
								stroke-width="2"
							/>
							<path
								d="M {interceptionData.target.x - 10} {interceptionData.target.y + 20} L {interceptionData.target.x - 20} {interceptionData.target.y + 20} L {interceptionData.target.x - 20} {interceptionData.target.y + 10}"
								fill="none"
								stroke="#ef4444"
								stroke-width="2"
							/>

							<!-- Target Attacker (User Icon) - Destructible -->
							<foreignObject
								x={interceptionData.target.x - 12}
								y={interceptionData.target.y - 12}
								width="24"
								height="24"
							>
								<div class="flex items-center justify-center w-full h-full text-red-500">
									<User class="w-full h-full">
										<animate attributeName="opacity" values="1;1;0" dur="1s" repeatCount="1" begin="0.8s" fill="freeze" />
										<animateTransform attributeName="transform" type="scale" values="1;1.2;0" dur="1s" repeatCount="1" begin="0.8s" fill="freeze" />
									</User>
								</div>
							</foreignObject>
						</g>

						<!-- Launch Flare at RedEye -->
						<g transform="translate({redeyePos.x}, {redeyePos.y})">
							<circle r="10" fill="#ef4444" opacity="0">
								<animate
									attributeName="opacity"
									values="0;1;0"
									dur="0.3s"
									repeatCount="1"
									fill="freeze"
								/>
								<animate attributeName="r" values="5;20" dur="0.3s" repeatCount="1" fill="freeze" />
							</circle>
						</g>

						<!-- Rocket 1 -->
						<g class="rocket-ship">
							<!-- Smoke Trail -->
							<path
								d={interceptionData.path1}
								fill="none"
								stroke="#64748b"
								stroke-width="2"
								opacity="0.4"
								stroke-dasharray="0, 1000"
							>
								<animate attributeName="stroke-dasharray" from="0, 1000" to="1000, 0" dur="0.8s" fill="freeze" />
								<animate attributeName="opacity" values="0.4;0" dur="1.2s" fill="freeze" />
							</path>

							<!-- The Rocket -->
							<g>
								<animateMotion dur="0.8s" repeatCount="1" path={interceptionData.path1} fill="freeze" rotate="auto" />
								<!-- Body -->
								<path d="M 10 0 L -6 -4 L -4 0 L -6 4 Z" fill="#334155" stroke="#ef4444" stroke-width="0.5" />
								<!-- Nose -->
								<path d="M 10 0 L 4 -2 L 4 2 Z" fill="#ef4444" />
								<!-- Engine Flame -->
								<path d="M -6 0 L -12 -3 L -10 0 L -12 3 Z" fill="#fb923c" class="missile-flame" />
								<path d="M -6 0 L -9 -1.5 L -8 0 L -9 1.5 Z" fill="#ffffff" class="missile-flame" />
							</g>
						</g>

						<!-- Rocket 2 -->
						<g class="rocket-ship">
							<path
								d={interceptionData.path2}
								fill="none"
								stroke="#64748b"
								stroke-width="2"
								opacity="0.4"
								stroke-dasharray="0, 1000"
							>
								<animate attributeName="stroke-dasharray" from="0, 1000" to="1000, 0" dur="0.8s" fill="freeze" begin="0.1s" />
								<animate attributeName="opacity" values="0.4;0" dur="1.2s" fill="freeze" begin="0.1s" />
							</path>

							<g>
								<animateMotion dur="0.8s" repeatCount="1" path={interceptionData.path2} fill="freeze" begin="0.1s" rotate="auto" />
								<!-- Body -->
								<path d="M 10 0 L -6 -4 L -4 0 L -6 4 Z" fill="#334155" stroke="#ef4444" stroke-width="0.5" />
								<!-- Nose -->
								<path d="M 10 0 L 4 -2 L 4 2 Z" fill="#ef4444" />
								<!-- Engine Flame -->
								<path d="M -6 0 L -12 -3 L -10 0 L -12 3 Z" fill="#fb923c" class="missile-flame" />
								<path d="M -6 0 L -9 -1.5 L -8 0 L -9 1.5 Z" fill="#ffffff" class="missile-flame" />
							</g>
						</g>

						<!-- Impact Blast - Refined -->
						<g transform="translate({interceptionData.target.x}, {interceptionData.target.y})">
							<!-- Core Flash -->
							<circle r="30" fill="#ffffff" opacity="0">
								<animate
									attributeName="opacity"
									values="0;1;0"
									dur="0.4s"
									begin="0.75s"
									fill="freeze"
								/>
								<animate attributeName="r" values="0;50" dur="0.4s" begin="0.75s" fill="freeze" />
							</circle>

							<!-- Explosion Cloud -->
							<circle r="40" fill="#fb923c" opacity="0">
								<animate attributeName="opacity" values="0;0.8;0" dur="0.8s" begin="0.8s" fill="freeze" />
								<animate attributeName="r" values="10;60" dur="0.8s" begin="0.8s" fill="freeze" />
							</circle>

							<!-- Secondary Shockwave -->
							<circle
								r="50"
								fill="none"
								stroke="#ef4444"
								stroke-width="2"
								class="animate-refined-blast"
							>
								<animate
									attributeName="opacity"
									values="0;1;0"
									dur="1.5s"
									repeatCount="1"
									begin="0.8s"
								/>
							</circle>

							<!-- Shatter Fragments (The "Attacker" breaking apart) -->
							{#each Array(12) as _, i}
								<rect width="4" height="4" fill="#ef4444">
									<animate
										attributeName="x"
										from="0"
										to={Math.cos(i * 30 * (Math.PI / 180)) * 80}
										dur="1s"
										begin="0.85s"
										fill="freeze"
									/>
									<animate
										attributeName="y"
										from="0"
										to={Math.sin(i * 30 * (Math.PI / 180)) * 80}
										dur="1s"
										begin="0.85s"
										fill="freeze"
									/>
									<animateTransform attributeName="transform" type="rotate" from="0" to={Math.random() * 360} dur="1s" begin="0.85s" />
									<animate attributeName="opacity" values="1;0" dur="1s" begin="0.85s" fill="freeze" />
								</rect>
							{/each}
						</g>
					</g>
				{/if}

				<!-- Connections -->
				{#each $spawners as spawner, i (spawner.id)}
					{@const pos = getPosition(i, $spawners.length)}
					{@const isActive = spawner.status === 'online' || spawner.status === 'Online'}
					{@const connectionPathD = getConnectionPath(pos.x, pos.y, center.x, center.y)}
					{@const colors = getStatusColor(spawner, hoveredSpawnerId === spawner.id)}

					<!-- Connection Line with curve -->
					<path
						id={`connection-${spawner.id}`}
						d={connectionPathD}
						stroke={colors.stroke}
						stroke-width={hoveredSpawnerId === spawner.id ? 3 : 2}
						stroke-dasharray={isActive ? '25, 5' : '40, 5'}
						opacity={hoveredSpawnerId === spawner.id ? 0.8 : 0.5}
						fill="none"
						filter={hoveredSpawnerId === spawner.id ? 'url(#glow)' : ''}
						style="transition: all 0.1s ease;"
					>
						{#if isActive}
							<animate
								attributeName="stroke-dashoffset"
								from="0"
								to="-15"
								dur="2.5s"
								repeatCount="indefinite"
							/>
						{:else}
							<animate
								attributeName="stroke-dashoffset"
								from="0"
								to="-10"
								dur="2s"
								repeatCount="indefinite"
							/>
						{/if}
					</path>

					<!-- Pulse Packet (Enhanced Spark) -->
					{#if pulsingSpawners.has(spawner.id) && isActive}
						<g filter="url(#strongGlow)">
							<!-- Main spark -->
							<circle r="4" fill="#10b971">
								<animateMotion dur="1s" repeatCount="1" path={connectionPathD} fill="freeze" />
								<animate attributeName="r" values="4;5;4" dur="0.3s" repeatCount="3" />
							</circle>
							<!-- Trail Effect -->
							<circle r="6" fill="#34d399" opacity="1.7">
								<animateMotion
									dur="0.5s"
									repeatCount="1"
									path={connectionPathD}
									fill="freeze"
									begin="0.05s"
								/>
							</circle>
							<circle r="2" fill="#6ee7b7" opacity="0.5">
								<animateMotion
									dur="0.5s"
									repeatCount="1"
									path={connectionPathD}
									fill="freeze"
									begin="0.1s"
								/>
							</circle>
						</g>
					{/if}

					<!-- Fire spark with trails -->
					{#if isActive}
						<g filter="url(#strongGlow)">
							<!-- Main fire spark (orange-yellow core) -->
							<circle r="3" fill="#ff6b35">
								<animateMotion dur="2.5s" repeatCount="indefinite" path={connectionPathD} />
								<animate
									attributeName="fill"
									values="#ff6b35;#ffd93d;#ff6b35"
									dur="8.4s"
									repeatCount="indefinite"
								/>
								<animate attributeName="r" values="3;3.5;3" dur="0.3s" repeatCount="indefinite" />
							</circle>

							<!-- Inner glow (bright yellow) -->
							<circle r="2" fill="#ffeb3b">
								<animateMotion dur="2.7s" repeatCount="indefinite" path={connectionPathD} />
								<animate
									attributeName="opacity"
									values="0.9;1;0.9"
									dur="8.2s"
									repeatCount="indefinite"
								/>
							</circle>

							<!-- Trail particles (orange to red gradient) -->
							<circle r="2.5" fill="#ff8c42" opacity="0.8">
								<animateMotion
									dur="2.5s"
									repeatCount="indefinite"
									path={connectionPathD}
									begin="0.1s"
								/>
								<animate
									attributeName="r"
									values="2.5;1.5;0.5"
									dur="0.8s"
									repeatCount="indefinite"
								/>
								<animate
									attributeName="opacity"
									values="0.8;0.4;0"
									dur="0.8s"
									repeatCount="indefinite"
								/>
							</circle>

							<circle r="2" fill="#ff6b35" opacity="0.7">
								<animateMotion
									dur="2.5s"
									repeatCount="indefinite"
									path={connectionPathD}
									begin="0.2s"
								/>
								<animate attributeName="r" values="2;1.2;0.3" dur="0.7s" repeatCount="indefinite" />
								<animate
									attributeName="opacity"
									values="0.7;0.3;0"
									dur="0.7s"
									repeatCount="indefinite"
								/>
							</circle>

							<circle r="1.5" fill="#d64933" opacity="0.6">
								<animateMotion
									dur="2.5s"
									repeatCount="indefinite"
									path={connectionPathD}
									begin="0.3s"
								/>
								<animate
									attributeName="r"
									values="1.5;0.8;0.2"
									dur="0.6s"
									repeatCount="indefinite"
								/>
								<animate
									attributeName="opacity"
									values="0.6;0.2;0"
									dur="0.6s"
									repeatCount="indefinite"
								/>
							</circle>

							<!-- Ember particles (small red dots) -->
							<circle r="1" fill="#d64933" opacity="0.5">
								<animateMotion
									dur="2.5s"
									repeatCount="indefinite"
									path={connectionPathD}
									begin="0.4s"
								/>
								<animate attributeName="r" values="1;0.5;0.1" dur="0.5s" repeatCount="indefinite" />
								<animate
									attributeName="opacity"
									values="0.5;0.1;0"
									dur="0.5s"
									repeatCount="indefinite"
								/>
							</circle>
						</g>
					{/if}
				{/each}
			</svg>

			<!-- Master Node (Center) -->
			<div class="absolute z-20" style="top: {center.y - 65}px; left: {center.x - 55}px;">
				<!-- Modern shield/badge container - smaller size -->
				<div class="relative w-[110px] h-[130px] flex items-center justify-center">
					<!-- Background shape -->
					<svg class="absolute inset-0 w-full h-full" viewBox="0 0 110 130">
						<!-- Main shield shape with sophisticated curves - TRANSPARENT FILL -->
						<path
							d="M 55 8
						L 95 25
						L 95 60
						Q 95 80, 83 95
						L 55 118
						L 27 95
						Q 15 80, 15 60
						L 15 25
						Z"
							fill="none"
							stroke="url(#masterBorderGradient)"
							class={`transition-all duration-500 master-shield ${
								masterReceiving ? 'master-shield-ignite' : ''
							}`}
							stroke-width="2.5"
							filter="url(#masterGlow)"
						/>

						<!-- Inner accent lines -->
						<path
							d="M 55 18
						L 83 30
						L 83 58
						Q 83 72, 75 82
						L 55 100
						L 35 82
						Q 27 72, 27 58
						L 27 30
						Z"
							fill="none"
							class={`transition-all duration-500 ${
								masterReceiving ? 'stroke-emerald-500/30' : 'stroke-blue-500/20'
							}`}
							stroke-width="1"
						/>

						<!-- Center vertical accent -->
						<line
							x1="55"
							y1="18"
							x2="55"
							y2="100"
							class={`transition-all duration-500 ${
								masterReceiving ? 'stroke-emerald-400/40' : 'stroke-blue-400/25'
							}`}
							stroke-width="0.8"
						/>
					</svg>

					<!-- Content layer -->
					<div class="relative z-10 flex flex-col items-center justify-center pt-3">
						<!-- Icon stack - just the main server icon, no badges -->
						<div class="relative mb-2">
							<!-- Main server rack icon only -->
							<Server
								class={`w-9 h-9 transition-all duration-500 ${
									masterReceiving ? 'text-emerald-400 master-icon-ignite' : 'text-blue-400'
								}`}
								style="filter: drop-shadow(0 0 8px currentColor);"
							/>
						</div>

						<!-- Labels -->
						<div class="flex flex-col items-center gap-0.5">
							<span
								class={` text-[8px] font-bold tracking-widest transition-colors duration-500 ${
									masterReceiving ? 'text-emerald-200' : 'text-blue-200'
								}`}>MASTER</span
							>
							<div class="flex items-center gap-1">
								<div
									class={`w-1.5 h-1.5 rounded-full transition-colors duration-800  ${
										masterReceiving ? 'bg-emerald-400 animate-pulse' : 'bg-emerald-200'
									}`}
								></div>
								<!--<span class="text-[6px] text-emerald-400 font-mono font-semibold">ONLINE</span>-->
							</div>
						</div>
					</div>

					<!-- Green particles flying in when receiving data -->
					{#if masterReceiving}
						<!-- Particles from top -->
						<div
							class="absolute -top-16 left-1/2 -translate-x-1/2 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-top shadow-[0_0_10px_rgba(52,211,153,0.8)]"
						></div>
						<div
							class="absolute -top-14 left-1/3 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-top"
							style="animation-delay: 0.05s;"
						></div>
						<div
							class="absolute -top-14 right-1/3 w-1.5 h-1.5 rounded-full bg-emerald-300 animate-particle-fly-top"
							style="animation-delay: 0.1s;"
						></div>

						<!-- Particles from sides -->
						<div
							class="absolute top-1/4 -left-16 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-left shadow-[0_0_10px_rgba(74,222,128,0.8)]"
						></div>
						<div
							class="absolute top-1/3 -left-14 w-1.5 h-1.5 rounded-full bg-slate-300-400 animate-particle-fly-left"
							style="animation-delay: 0.07s;"
						></div>
						<div
							class="absolute top-1/4 -right-16 w-2 h-2 rounded-full bg-emerald-500 animate-particle-fly-right shadow-[0_0_10px_rgba(16,185,129,0.8)]"
						></div>
						<div
							class="absolute top-1/3 -right-14 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-right"
							style="animation-delay: 0.08s;"
						></div>

						<!-- Particles from bottom -->
						<div
							class="absolute -bottom-12 left-1/2 -translate-x-1/2 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-bottom shadow-[0_0_10px_rgba(52,211,153,0.8)]"
						></div>
						<div
							class="absolute -bottom-10 left-1/3 w-1.5 h-1.5 rounded-full bg-green-300 animate-particle-fly-bottom"
							style="animation-delay: 0.06s;"
						></div>

						<!-- Diagonal particles -->
						<div
							class="absolute -top-12 -left-12 w-1.5 h-1.5 rounded-full bg-emerald-300 animate-particle-fly-diagonal-tl"
							style="animation-delay: 0.04s;"
						></div>
						<div
							class="absolute -top-12 -right-12 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-diagonal-tr"
							style="animation-delay: 0.03s;"
						></div>
						<div
							class="absolute -bottom-8 -left-10 w-1.5 h-1.5 rounded-full bg-emerald-400 animate-particle-fly-diagonal-bl"
							style="animation-delay: 0.09s;"
						></div>
						<div
							class="absolute -bottom-8 -right-10 w-1.5 h-1.5 rounded-full bg-green-300 animate-particle-fly-diagonal-br"
							style="animation-delay: 0.05s;"
						></div>
					{/if}
				</div>
			</div>

			<!-- RedEye Node (Cyber Sentinel) -->
			<div
				class="absolute z-30 flex flex-col items-center group cursor-pointer transition-all duration-300 hover:scale-110"
				style="top: {redeyePos.y - 45}px; left: {redeyePos.x - 45}px;"
			>
				<div class="relative w-[90px] h-[90px] flex items-center justify-center">
					<!-- NEW Cyber Frame -->
					<svg class="absolute inset-0 w-full h-full pointer-events-none" viewBox="0 0 100 100">
						<!-- Main Frame Outer -->
						<path
							d="M 25 5 L 75 5 L 95 25 L 95 75 L 75 95 L 25 95 L 5 75 L 5 25 Z"
							fill="rgba(15, 23, 42, 0.6)"
							stroke="#ef4444"
							stroke-width="1.5"
							stroke-opacity="0.4"
							class="backdrop-blur-xl"
						/>
						<!-- Corner Brackets -->
						<path d="M 25 5 L 10 5 L 5 20" fill="none" stroke="#ef4444" stroke-width="2.5" class="animate-pulse" />
						<path d="M 75 5 L 90 5 L 95 20" fill="none" stroke="#ef4444" stroke-width="2.5" class="animate-pulse" />
						<path d="M 5 80 L 10 95 L 25 95" fill="none" stroke="#ef4444" stroke-width="2.5" class="animate-pulse" />
						<path d="M 95 80 L 90 95 L 75 95" fill="none" stroke="#ef4444" stroke-width="2.5" class="animate-pulse" />

						<!-- Internal decorative crosshair -->
						<circle cx="50" cy="50" r="40" fill="none" stroke="#ef4444" stroke-width="0.5" stroke-dasharray="2 4" opacity="0.3" />
					</svg>

					<!-- Rotating Outer Ring -->
					<svg class="absolute inset-0 w-full h-full animate-spin-slow opacity-40">
						<circle
							cx="50"
							cy="50"
							r="42"
							fill="none"
							stroke="#ef4444"
							stroke-width="1"
							stroke-dasharray="10 20"
						/>
					</svg>

					<!-- The Cyber Eye -->
					<div
						class="relative z-10 w-16 h-16 flex items-center justify-center filter drop-shadow-[0_0_15px_rgba(239,68,68,0.8)]"
					>
						<svg viewBox="0 0 100 100" class="w-full h-full">
							<!-- Scanner Laser -->
							<line
								x1="10"
								y1="50"
								x2="90"
								y2="50"
								stroke="#ff0000"
								stroke-width="1"
								class="animate-scan-laser"
							/>

							<!-- Outer Shell -->
							<circle
								cx="50"
								cy="50"
								r="45"
								fill="none"
								stroke="#ef4444"
								stroke-width="1.5"
								opacity="0.4"
								stroke-dasharray="15,5"
							/>

							<!-- Main Lens -->
							<path
								d="M 10 50 Q 50 5 90 50 Q 50 95 10 50 Z"
								fill="#000"
								stroke="#ef4444"
								stroke-width="2.5"
							/>

							<!-- Iris Array -->
							<g class={showInterception ? 'animate-pulse' : ''} style="animation-duration: 0.5s">
								<circle
									cx="50"
									cy="50"
									r="22"
									fill="none"
									stroke="#ef4444"
									stroke-width="1"
									opacity="0.6"
								/>
								<circle cx="50" cy="50" r="15" fill="#450a0a" />
								<circle
									cx="50"
									cy="50"
									r="10"
									fill={showInterception ? '#ff0000' : '#7f1d1d'}
									class="transition-colors duration-200"
								/>
							</g>

							<!-- Pupil Details -->
							<rect x="49" y="40" width="2" height="20" fill="#ffffff" opacity="0.8" rx="1" />
						</svg>
					</div>

					<!-- Sentinel Status -->
					<div
						class="absolute -bottom-4 px-3 py-0.5 bg-red-950/80 border border-red-500/50 rounded-full text-[8px] font-black text-red-100 tracking-widest uppercase shadow-lg z-40"
					>
						{showInterception ? 'ENGAGED' : 'SCANNING'}
					</div>
				</div>
				<div class="mt-8 flex flex-col items-center">
					<span class="text-[10px] font-black text-red-500 tracking-[0.4em] uppercase opacity-90"
						>REDEYE_CORE</span
					>
				</div>
			</div>

			<!-- Database Node (Quantum Persistence Core) -->
			<div
				class="absolute z-20 flex flex-col items-center group cursor-pointer transition-all duration-300 hover:scale-110"
				style="top: {databasePos.y - 45}px; left: {databasePos.x - 45}px;"
			>
				<div class="relative w-[90px] h-[90px] flex items-center justify-center">
					<!-- NEW Cyber Frame (Emerald) -->
					<svg class="absolute inset-0 w-full h-full pointer-events-none" viewBox="0 0 100 100">
						<!-- Main Frame Outer -->
						<path
							d="M 25 5 L 75 5 L 95 25 L 95 75 L 75 95 L 25 95 L 5 75 L 5 25 Z"
							fill="rgba(6, 78, 59, 0.4)"
							stroke="#10b981"
							stroke-width="1.5"
							stroke-opacity="0.4"
							class="backdrop-blur-xl"
						/>
						<!-- Corner Brackets -->
						<path d="M 25 5 L 10 5 L 5 20" fill="none" stroke="#10b981" stroke-width="2.5" class="animate-pulse" />
						<path d="M 75 5 L 90 5 L 95 20" fill="none" stroke="#10b981" stroke-width="2.5" class="animate-pulse" />
						<path d="M 5 80 L 10 95 L 25 95" fill="none" stroke="#10b981" stroke-width="2.5" class="animate-pulse" />
						<path d="M 95 80 L 90 95 L 75 95" fill="none" stroke="#10b981" stroke-width="2.5" class="animate-pulse" />

						<!-- Data stream ring -->
						<circle cx="50" cy="50" r="42" fill="none" stroke="#10b981" stroke-width="0.5" stroke-dasharray="10 5" opacity="0.2">
							<animateTransform attributeName="transform" type="rotate" from="360 50 50" to="0 50 50" dur="20s" repeatCount="indefinite" />
						</circle>
					</svg>

					<!-- Quantum Storage Cylinder Stack -->
					<div
						class="relative z-10 w-16 h-16 flex items-center justify-center filter drop-shadow-[0_0_15px_rgba(16,185,129,0.8)]"
					>
						<svg viewBox="0 0 100 100" class="w-full h-full">
							<!-- Bottom Plate -->
							<ellipse cx="50" cy="75" rx="30" ry="10" fill="#064e3b" stroke="#10b981" />

							<!-- Middle Plate 1 -->
							<ellipse
								cx="50"
								cy="60"
								rx="30"
								ry="10"
								fill="#065f46"
								stroke="#10b981"
								class="animate-pulse"
							/>

							<!-- Middle Plate 2 -->
							<ellipse
								cx="50"
								cy="45"
								rx="30"
								ry="10"
								fill="#065f46"
								stroke="#10b981"
								style="animation-delay: 0.5s"
								class="animate-pulse"
							/>

							<!-- Top Plate -->
							<ellipse cx="50" cy="30" rx="30" ry="10" fill="#059669" stroke="#34d399" />

							<!-- Connecting Columns -->
							<rect x="25" y="30" width="2" height="45" fill="#10b981" opacity="0.6" />
							<rect x="73" y="30" width="2" height="45" fill="#10b981" opacity="0.6" />

							<!-- Central Data Stream -->
							<rect x="48" y="20" width="4" height="60" fill="url(#dataStreamGradient)" rx="2">
								<animate
									attributeName="opacity"
									values="0.3;1;0.3"
									dur="2s"
									repeatCount="indefinite"
								/>
							</rect>

							<!-- Orbital Read/Write Heads -->
							<g class="animate-spin" style="animation-duration: 4s">
								<circle cx="85" cy="50" r="3" fill="#6ee7b7" />
								<circle cx="15" cy="50" r="3" fill="#6ee7b7" />
							</g>
						</svg>
					</div>

					<!-- Resonance Ring -->
					<div
						class="absolute inset-0 border border-emerald-500/20 scale-150 animate-ping-slower opacity-10"
					></div>
				</div>
				<div class="mt-6 flex flex-col items-center">
					<div class="flex items-center gap-2 mb-1">
						<span
							class="text-[10px] font-black text-emerald-500 tracking-[0.4em] uppercase opacity-90"
							>QUANTUM_STORAGE</span
						>
					</div>
					<div class="flex gap-1">
						{#each Array(4) as _, i}
							<div class="w-3 h-1 bg-emerald-900/60 rounded-full overflow-hidden">
								<div
									class="w-full h-full bg-emerald-400 animate-pulse"
									style="animation-delay: {i * 0.2}s"
								></div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Spawner Nodes -->
			{#each $spawners as spawner, i (spawner.id)}
				{@const pos = getPosition(i, $spawners.length)}
				{@const isActive = spawner.status === 'online' || spawner.status === 'Online'}
				{@const utilization =
					spawner.max_instances > 0 ? (spawner.current_instances / spawner.max_instances) * 100 : 0}

				<div
					class="absolute z-10 flex flex-col items-center group cursor-pointer transition-all duration-300 hover:scale-110 hover:z-30"
					style="top: {pos.y - 28}px; left: {pos.x - 28}px;"
					in:scale={{ duration: 400, delay: i * 100 }}
					onmouseenter={() => (hoveredSpawnerId = spawner.id)}
					onmouseleave={() => (hoveredSpawnerId = null)}
					role="button"
					tabindex="0"
				>
					<!-- Node circle with utilization ring -->
					<div class="relative">
						<!-- Utilization ring background -->
						<svg
							class="absolute -inset-2 w-16 h-16 -rotate-90"
							style="filter: drop-shadow(0 0 8px rgba(100, 116, 139, 0.3));"
						>
							<circle
								cx="32"
								cy="32"
								r="30"
								fill="none"
								stroke="rgba(100, 116, 139, 0.2)"
								stroke-width="2"
							/>
							{#if isActive}
								<circle
									cx="32"
									cy="32"
									r="30"
									fill="none"
									stroke={utilization > 80 ? '#f59e0b' : utilization > 50 ? '#3b82f6' : '#10b981'}
									stroke-width="2"
									stroke-dasharray={`${(utilization / 100) * 188.4} 188.4`}
									style="transition: stroke-dasharray 0.5s ease, stroke 0.5s ease;"
								/>
							{/if}
						</svg>

						<!-- Main node -->
						<div
							class={`
						relative w-12 h-12 rounded-full flex items-center justify-center border-2 shadow-lg backdrop-blur-md
						${isActive ? 'bg-slate-800/90 border-slate-600 shadow-slate-500/30' : 'bg-slate-900/95 border-red-500/70 shadow-red-500/40'}
						${pulsingSpawners.has(spawner.id) ? 'node-pulse border-emerald-400 shadow-emerald-500/60' : ''}
						transition-all duration-300
					`}
						>
							{#if isActive}
								<Cpu
									class={`w-6 h-6 text-slate-700 dark:text-slate-300 transition-colors duration-800`}
								/>
							{:else}
								<Skull
									class={`w-6 h-6 text-red-400 animate-pulse-slow transition-colors duration-800`}
								/>
							{/if}

							<!-- Status indicator - only show for online nodes -->
							{#if isActive}
								<div
									class={`absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full border-2 border-slate-900
								bg-emerald-500 status-pulse`}
								></div>
							{/if}
						</div>
					</div>

					<!-- Enhanced tooltip -->
					<div
						class="absolute top-16 flex flex-col items-center bg-white/95 dark:bg-slate-950/95 backdrop-blur-sm px-4 py-2 rounded-xl border border-slate-300/50 dark:border-slate-700/50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 whitespace-nowrap z-40 pointer-events-none shadow-2xl"
					>
						<span class="text-sm font-bold text-slate-900 dark:text-white mb-1"
							>Spawner #{spawner.id}</span
						>
						<div
							class="w-full h-px bg-gradient-to-r from-transparent via-slate-600 to-transparent mb-1"
						></div>
						<div class="flex items-center gap-2 text-xs text-slate-700 dark:text-slate-300">
							<Activity class="w-3 h-3" />
							<span>{spawner.region}</span>
						</div>
						<div class="text-xs text-slate-500 dark:text-slate-400 mt-1">
							Instances: <span class="font-mono text-slate-800 dark:text-slate-200"
								>{spawner.current_instances}</span
							>/<span class="font-mono">{spawner.max_instances}</span>
							<span class="ml-2 text-[10px]">({utilization.toFixed(0)}%)</span>
						</div>
						<span
							class={`text-xs font-mono mt-1 font-semibold ${isActive ? 'text-emerald-400' : 'text-red-400'}`}
						>
							● {spawner.status?.toUpperCase() || 'UNKNOWN'}
						</span>
					</div>
				</div>
			{/each}

			<!-- Shatter animations for removed spawners -->
			{#each Array.from(removingSpawners) as removedId (removedId)}
				{@const spawnerData = previousSpawners.get(removedId)}
				{#if spawnerData}
					{@const allSpawnerIds = Array.from(previousSpawners.keys())}
					{@const spawnerIndex = allSpawnerIds.indexOf(removedId)}
					{@const totalSpawners = Math.max(allSpawnerIds.length, 1)}
					{@const pos = getPosition(spawnerIndex >= 0 ? spawnerIndex : 0, totalSpawners)}
					{@const pieces = generateShatterPieces(spawnerIndex >= 0 ? spawnerIndex : 0)}

					<!-- Shatter effect container -->
					<div class="absolute z-50 pointer-events-none" style="top: {pos.y}px; left: {pos.x}px;">
						<!-- Explosion flash -->
						<div
							class="absolute -inset-16 bg-red-500/30 rounded-full animate-explosion-flash"
						></div>

						<!-- Shockwave ring -->
						<div
							class="absolute -inset-12 border-2 border-red-400/50 rounded-full animate-shockwave"
						></div>

						<!-- Shatter pieces -->
						{#each pieces as piece (piece.id)}
							<div
								class="absolute w-2 h-2 bg-gradient-to-br from-slate-600 to-slate-800 rounded-sm shadow-lg animate-shatter-piece"
								style="
							animation-delay: {piece.delay}s;
							--dx: {piece.dx}px;
							--dy: {piece.dy}px;
							--rotation: {piece.rotation}deg;
							width: {piece.size}px;
							height: {piece.size}px;
						"
							></div>
						{/each}

						<!-- Smoke particles -->
						<div class="absolute -inset-8">
							{#each Array(8) as _, smokeIndex}
								<div
									class="absolute w-6 h-6 bg-slate-600/40 rounded-full blur-sm animate-smoke-rise"
									style="
								left: {Math.cos((smokeIndex / 8) * Math.PI * 2) * 20}px;
								top: {Math.sin((smokeIndex / 8) * Math.PI * 2) * 20}px;
								animation-delay: {smokeIndex * 0.05}s;
							"
								></div>
							{/each}
						</div>

						<!-- Spark particles -->
						{#each Array(16) as _, sparkIndex}
							<div
								class="absolute w-1 h-1 bg-orange-400 rounded-full shadow-[0_0_4px_rgba(251,146,60,0.8)] animate-spark-fly"
								style="
							animation-delay: {sparkIndex * 0.02}s;
							--spark-angle: {(sparkIndex / 16) * 360}deg;
							--spark-distance: {30 + Math.random() * 40}px;
						"
							></div>
						{/each}
					</div>
				{/if}
			{/each}
		</div>
	</div>
</div>

<!-- Stats overlay (outside zoom wrapper so it stays fixed) -->
<div class="absolute top-4 left-4 flex flex-col gap-2 text-xs font-mono z-30">
	<div class="backdrop-blur-sm px-3 py-2 rounded-lg">
		<span class="text-slate-500/30 dark:text-slate-400/30">Active Spawners:</span>
		<span class="text-emerald-400/30 ml-2 font-bold">
			{$spawners.filter((s) => s.status === 'online' || s.status === 'Online')
				.length}/{$spawners.length}
		</span>
	</div>
	<div class="bg-transparent backdrop-blur-sm px-3 py-2 rounded-lg">
		<span class="text-slate-500/30 dark:text-slate-400/30">Total Instances:</span>
		<span class="text-blue-400/30 ml-2 font-bold">
			{$spawners.reduce((sum, s) => sum + s.current_instances, 0)}/{$spawners.reduce(
				(sum, s) => sum + s.max_instances,
				0
			)}
		</span>
	</div>
</div>

<style>
	@keyframes gridMove {
		0% {
			background-position: 0 0;
		}
		100% {
			background-position: 250px 80px;
		}
	}

	/* Animated gradient blobs */
	.gradient-blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(60px);
		opacity: 0.6;
	}

	.blob-1 {
		width: 400px;
		height: 400px;
		top: -100px;
		right: -100px;
		animation: float1 12s ease-in-out infinite;
	}

	.blob-2 {
		width: 350px;
		height: 350px;
		bottom: -80px;
		left: -80px;
		animation: float2 15s ease-in-out infinite;
	}

	.blob-3 {
		width: 300px;
		height: 300px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		animation: float3 10s ease-in-out infinite;
	}

	.blob-4 {
		width: 320px;
		height: 320px;
		top: 20%;
		right: 30%;
		animation: float4 13s ease-in-out infinite;
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
			transform: translate(-50%, -50%) scale(1.25);
		}
	}

	@keyframes float4 {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		40% {
			transform: translate(-30px, 35px) scale(1.12);
		}
		80% {
			transform: translate(25px, -25px) scale(0.92);
		}
	}

	/* Background particles */
	.background-particle {
		position: absolute;
		border-radius: 50%;
		pointer-events: none;
		animation: backgroundParticleFloat infinite ease-in-out;
		box-shadow: 0 0 8px currentColor;
		opacity: 0;
	}

	@keyframes backgroundParticleFloat {
		0%,
		100% {
			transform: translateY(0) translateX(0) scale(0);
			opacity: 0;
		}
		10% {
			opacity: 0.6;
		}
		50% {
			transform: translateY(-50px) translateX(20px) scale(1);
			opacity: 0.4;
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
			20deg,
			transparent 30%,
			rgba(56, 189, 248, 0.12) 50%,
			transparent 90%
		);
		background-size: 100% 100%;
		animation: shimmer 50s ease-in-out infinite;
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

	/* Dynamic border gradient colors */
	.master-border-top {
		stop-color: #3b82f6;
		stop-opacity: 1;
	}

	.master-border-bottom {
		stop-color: #1d4ed8;
		stop-opacity: 1;
	}

	.master-shield-ignite .master-border-top {
		stop-color: #10b981;
		stop-opacity: 1;
	}

	.master-shield-ignite .master-border-bottom {
		stop-color: #059669;
		stop-opacity: 1;
	}

	/* Shield shape animations */
	.master-shield {
		filter: drop-shadow(0 0 20px rgba(59, 130, 246, 0.35));
		transition: all 0.5s ease-out;
	}

	.master-shield-ignite {
		filter: drop-shadow(0 0 35px rgba(16, 185, 129, 0.7))
			drop-shadow(0 0 15px rgba(52, 211, 153, 0.5));
		animation: shieldIgnite 0.8s ease-out;
	}

	@keyframes shieldIgnite {
		0% {
			filter: drop-shadow(0 0 20px rgba(16, 185, 129, 0.4));
		}
		40% {
			filter: drop-shadow(0 0 45px rgba(16, 185, 129, 0.9))
				drop-shadow(0 0 25px rgba(52, 211, 153, 0.7));
		}
		100% {
			filter: drop-shadow(0 0 35px rgba(16, 185, 129, 0.7))
				drop-shadow(0 0 15px rgba(52, 211, 153, 0.5));
		}
	}

	/* Icon ignition - subtle rotation and glow */
	.master-icon-ignite {
		animation: iconIgnite 0.6s ease-out;
	}

	@keyframes iconIgnite {
		0% {
			transform: rotate(0deg);
			filter: drop-shadow(0 0 8px currentColor) brightness(1);
		}
		50% {
			transform: rotate(1deg);
			filter: drop-shadow(0 0 15px currentColor) brightness(1.3);
		}
		100% {
			transform: rotate(0deg);
			filter: drop-shadow(0 0 10px currentColor) brightness(1.1);
		}
	}

	/* Green particles flying in from all directions */
	@keyframes particleFlyTop {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(0, 60px) scale(0.2);
		}
	}

	@keyframes particleFlyBottom {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(0, -50px) scale(0.2);
		}
	}

	@keyframes particleFlyLeft {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(60px, 0) scale(0.2);
		}
	}

	@keyframes particleFlyRight {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(-60px, 0) scale(0.2);
		}
	}

	@keyframes particleFlyDiagonalTL {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(50px, 50px) scale(0.2);
		}
	}

	@keyframes particleFlyDiagonalTR {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(-50px, 50px) scale(0.2);
		}
	}

	@keyframes particleFlyDiagonalBL {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(50px, -45px) scale(0.2);
		}
	}

	@keyframes particleFlyDiagonalBR {
		0% {
			opacity: 0;
			transform: translate(0, 0) scale(0.5);
		}
		30% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate(-50px, -45px) scale(0.2);
		}
	}

	.animate-particle-fly-top {
		animation: particleFlyTop 0.8s ease-out;
	}

	.animate-particle-fly-bottom {
		animation: particleFlyBottom 0.8s ease-out;
	}

	.animate-particle-fly-left {
		animation: particleFlyLeft 0.8s ease-out;
	}

	.animate-particle-fly-right {
		animation: particleFlyRight 0.8s ease-out;
	}

	.animate-particle-fly-diagonal-tl {
		animation: particleFlyDiagonalTL 0.85s ease-out;
	}

	.animate-particle-fly-diagonal-tr {
		animation: particleFlyDiagonalTR 0.85s ease-out;
	}

	.animate-particle-fly-diagonal-bl {
		animation: particleFlyDiagonalBL 0.85s ease-out;
	}

	.animate-particle-fly-diagonal-br {
		animation: particleFlyDiagonalBR 0.85s ease-out;
	}

	@keyframes nodePulse {
		0%,
		100% {
			box-shadow: 0 0 20px rgba(16, 185, 129, 0.4);
		}
		50% {
			box-shadow: 0 0 30px rgba(16, 185, 129, 0.6);
		}
	}

	.node-pulse {
		animation: nodePulse 0.8s ease-in-out;
	}

	@keyframes statusPulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.6;
		}
	}

	.status-pulse {
		animation: statusPulse 2s ease-in-out infinite;
	}

	@keyframes ping {
		0% {
			transform: scale(1);
			opacity: 0.5;
		}
		50% {
			opacity: 0.3;
		}
		100% {
			transform: scale(1.5);
			opacity: 0;
		}
	}

	.animate-ping-slow {
		animation: ping 3s cubic-bezier(0, 0, 0.2, 1) infinite;
	}

	.animate-ping-slower {
		animation: ping 5s cubic-bezier(0, 0, 0.2, 1) infinite;
	}

	@keyframes pulse-slow {
		0%,
		100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.7;
			transform: scale(1.05);
		}
	}

	.animate-pulse-slow {
		animation: pulse-slow 2s ease-in-out infinite;
	}

	/* Shatter animation styles */
	@keyframes explosionFlash {
		0% {
			opacity: 0;
			transform: scale(0);
		}
		20% {
			opacity: 0.8;
			transform: scale(1);
		}
		100% {
			opacity: 0;
			transform: scale(2);
		}
	}

	.animate-explosion-flash {
		animation: explosionFlash 0.6s ease-out;
	}

	@keyframes shockwave {
		0% {
			opacity: 0.8;
			transform: scale(0);
		}
		100% {
			opacity: 0;
			transform: scale(3);
		}
	}

	.animate-shockwave {
		animation: shockwave 0.8s ease-out;
	}

	@keyframes shatterPiece {
		0% {
			opacity: 1;
			transform: translate(0, 0) rotate(0deg) scale(1);
		}
		100% {
			opacity: 0;
			transform: translate(var(--dx), var(--dy)) rotate(var(--rotation)) scale(0.3);
		}
	}

	.animate-shatter-piece {
		animation: shatterPiece 1s ease-out forwards;
	}

	@keyframes smokeRise {
		0% {
			opacity: 0.6;
			transform: translateY(0) scale(0.5);
		}
		100% {
			opacity: 0;
			transform: translateY(-60px) scale(1.5);
		}
	}

	.animate-smoke-rise {
		animation: smokeRise 1.2s ease-out forwards;
	}

	@keyframes sparkFly {
		0% {
			opacity: 1;
			transform: rotate(var(--spark-angle)) translateX(0) scale(1);
		}
		100% {
			opacity: 0;
			transform: rotate(var(--spark-angle)) translateX(var(--spark-distance)) scale(0);
		}
	}

	.animate-spark-fly {
		animation: sparkFly 0.8s ease-out forwards;
	}

	@keyframes scanLine {
		0% {
			top: 0%;
			opacity: 0;
		}
		10% {
			opacity: 1;
		}
		90% {
			opacity: 1;
		}
		100% {
			top: 100%;
			opacity: 0;
		}
	}

	.animate-scan-line {
		animation: scanLine 3s linear infinite;
	}

	@keyframes redeyeFlicker {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.3;
		}
		55% {
			opacity: 1;
		}
		60% {
			opacity: 0.4;
		}
	}

	.animate-redeye-flicker {
		animation: redeyeFlicker 0.2s infinite;
	}

	@keyframes missileFlame {
		0%, 100% { transform: scaleY(1); opacity: 0.8; }
		50% { transform: scaleY(1.5); opacity: 1; }
	}

	.missile-flame {
		animation: missileFlame 0.1s ease-in-out infinite;
		transform-origin: top;
	}

	.rocket-ship {
		filter: drop-shadow(0 0 5px #ef4444);
	}

	@keyframes energySurge {
		0% {
			stroke-dashoffset: 100;
			opacity: 0.2;
		}
		50% {
			opacity: 1;
		}
		100% {
			stroke-dashoffset: 0;
			opacity: 0.2;
		}
	}

	.animate-energy-surge {
		stroke-dasharray: 20, 80;
		animation: energySurge 0.8s linear infinite;
	}

	@keyframes threatAppear {
		0% {
			transform: scale(0) rotate(45deg);
			opacity: 0;
		}
		10% {
			transform: scale(1) rotate(45deg);
			opacity: 1;
		}
		80% {
			transform: scale(1) rotate(45deg);
			opacity: 1;
		}
		100% {
			transform: scale(0) rotate(45deg);
			opacity: 0;
		}
	}

	@keyframes refinedBlast {
		0%,
		70% {
			transform: scale(0);
			opacity: 0;
		}
		75% {
			transform: scale(1);
			opacity: 0.8;
		}
		100% {
			transform: scale(3);
			opacity: 0;
		}
	}

	.animate-threat-appear {
		animation: threatAppear 3s ease-out forwards;
	}

	.animate-refined-blast {
		animation: refinedBlast 3s ease-out forwards;
		transform-box: fill-box;
		transform-origin: center;
	}

	@keyframes scanLaser {
		0%,
		100% {
			transform: translateY(-30px);
			opacity: 0;
		}
		50% {
			transform: translateY(30px);
			opacity: 0.8;
		}
	}

	.animate-scan-laser {
		animation: scanLaser 2s ease-in-out infinite;
	}

	@keyframes spin-slow {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.animate-spin-slow {
		animation: spin-slow 10s linear infinite;
	}
</style>
