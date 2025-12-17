<script lang="ts">
	import { onMount } from 'svelte';
	import { spawners } from '$lib/stores';
	import { Server, Activity, Cpu } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';

	// Animation state
	let center = { x: 300, y: 300 }; // Master node position
	let spawnerRadius = 250; // Radius for spawner placement around the master

	// Track heartbeats for pulse animation
	let lastHeartbeats: Record<number, number> = $state({});
	let pulsingSpawners: Set<number> = $state(new Set());
	let masterReceiving = $state(false); // Track when master receives spark
	let masterIgniteTimeout: ReturnType<typeof setTimeout> | null = null;

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

	// Resize observer to make visualization responsive
	onMount(() => {
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
			containerElement.addEventListener('mouseleave', handleMouseLeave); // New mouseleave to disable zoom
			containerElement.addEventListener('dblclick', handleDblClick);
			document.addEventListener('mousemove', handleMouseMove);
			document.addEventListener('mouseup', handleMouseUp);

			return () => {
				resizeObserver.disconnect();
				if (containerElement) {
					containerElement.removeEventListener('wheel', handleWheel);
					containerElement.removeEventListener('mousedown', handleMouseDown);
					containerElement.removeEventListener('mouseleave', handleMouseLeave); // Clean up new listener
					containerElement.removeEventListener('dblclick', handleDblClick);
				}
				document.removeEventListener('mousemove', handleMouseMove);
				document.removeEventListener('mouseup', handleMouseUp);
			};
		}
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

	function getStatusColor(spawner: any, isHovered: boolean) {
		const isActive = spawner.status === 'online' || spawner.status === 'Online';
		
		if (!isActive) return { stroke: '#ef4444', glow: 'rgba(239, 68, 68, 0.3)' };
		if (isHovered) return { stroke: '#38bdf8', glow: 'rgba(56, 189, 248, 0.5)' };
		return { stroke: '#64748b', glow: 'rgba(100, 116, 139, 0.2)' };
	}
</script>

			<div
				bind:this={containerElement}
				class="relative w-full h-[600px] bg-slate-950 rounded-2xl border border-slate-800/50 overflow-hidden flex items-center justify-center shadow-2xl transition-colors duration-300"
			>



	<!-- Animated gradient blobs in background -->


	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<!-- Primary gradient blob -->
		<div class="gradient-blob blob-1 bg-gradient-to-br from-blue-600/15 via-cyan-600/10 to-transparent"></div>
		
		<!-- Secondary gradient blob -->
		<div class="gradient-blob blob-2 bg-gradient-to-tl from-purple-600/12 via-blue-600/8 to-transparent"></div>
		
		<!-- Tertiary gradient blob -->
		<div class="gradient-blob blob-3 bg-gradient-to-tr from-cyan-600/10 via-blue-500/8 to-transparent"></div>
		
		<!-- Quaternary blob for more depth -->
		<div class="gradient-blob blob-4 bg-gradient-to-bl from-indigo-600/12 via-blue-700/6 to-transparent"></div>
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
		class="absolute inset-0 opacity-20"
		style="background-image: 
			linear-gradient(rgba(56, 189, 248, 0.03) 1px, transparent 1px),
			linear-gradient(90deg, rgba(56, 189, 248, 0.03) 1px, transparent 1px);
			background-size: 40px 40px;
			animation: gridMove 20s linear infinite;"
	></div>

	<!-- Radial gradient overlay for depth -->
	<div
		class="absolute inset-0 bg-radial-gradient opacity-40"
		style="background: radial-gradient(circle at 50% 50%, transparent 0%, rgba(0, 0, 0, 0.4) 100%);"
	></div>

	<!-- Shimmer overlay -->
	<div class="absolute inset-0 opacity-10 pointer-events-none gradient-overlay"></div>

	<!-- Edge Fade Overlays -->
	<div class="absolute inset-0 pointer-events-none fade-overlay-top"></div>
	<div class="absolute inset-0 pointer-events-none fade-overlay-bottom"></div>
	<div class="absolute inset-0 pointer-events-none fade-overlay-left"></div>
	<div class="absolute inset-0 pointer-events-none fade-overlay-right"></div>

	<!-- Zoomable and pannable content wrapper -->
	<div 
		class="absolute inset-0 transition-transform duration-200 ease-out pointer-events-none"
		style="transform: scale({zoom}) translate({panOffset.x / zoom}px, {panOffset.y / zoom}px); transform-origin: center center;"
	>
		<div class="pointer-events-auto">

	<svg class="w-full h-full pointer-events-none absolute inset-0" viewBox="0 0 {svgDimensions.width} {svgDimensions.height}">
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
		</defs>

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
				stroke-dasharray={isActive ? '10, 5' : '5, 5'}
				opacity={hoveredSpawnerId === spawner.id ? 0.8 : 0.5}
				fill="none"
				filter={hoveredSpawnerId === spawner.id ? 'url(#glow)' : ''}
				style="transition: all 0.3s ease;"
			>
				{#if isActive}
					<animate
						attributeName="stroke-dashoffset"
						from="0"
						to="-15"
						dur="1.5s"
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
					<circle r="4" fill="#10b981">
						<animateMotion
							dur="1s"
							repeatCount="1"
							path={connectionPathD}
							fill="freeze"
						/>
						<animate
							attributeName="r"
							values="4;5;4"
							dur="0.3s"
							repeatCount="3"
						/>
					</circle>
					<!-- Trail Effect -->
					<circle r="3" fill="#34d399" opacity="0.7">
						<animateMotion
							dur="1s"
							repeatCount="1"
							path={connectionPathD}
							fill="freeze"
							begin="0.05s"
						/>
					</circle>
					<circle r="2" fill="#6ee7b7" opacity="0.5">
						<animateMotion
							dur="1s"
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
						<animateMotion 
							dur="2.5s" 
							repeatCount="indefinite" 
							path={connectionPathD}
						/>
						<animate
							attributeName="fill"
							values="#ff6b35;#ffd93d;#ff6b35"
							dur="0.4s"
							repeatCount="indefinite"
						/>
						<animate
							attributeName="r"
							values="3;3.5;3"
							dur="0.3s"
							repeatCount="indefinite"
						/>
					</circle>
					
					<!-- Inner glow (bright yellow) -->
					<circle r="2" fill="#ffeb3b">
						<animateMotion 
							dur="2.5s" 
							repeatCount="indefinite" 
							path={connectionPathD}
						/>
						<animate
							attributeName="opacity"
							values="0.9;1;0.9"
							dur="0.2s"
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
						<animate
							attributeName="r"
							values="2;1.2;0.3"
							dur="0.7s"
							repeatCount="indefinite"
						/>
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
						<animate
							attributeName="r"
							values="1;0.5;0.1"
							dur="0.5s"
							repeatCount="indefinite"
						/>
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
	<div
		class="absolute z-20"
		style="top: {center.y - 65}px; left: {center.x - 55}px;"
	>
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
						masterReceiving 
							? 'stroke-emerald-500/30' 
							: 'stroke-blue-500/20'
					}`}
					stroke-width="1"
				/>
				
				
				<!-- Center vertical accent -->
				<line 
					x1="55" y1="18" 
					x2="55" y2="100" 
					class={`transition-all duration-500 ${
						masterReceiving 
							? 'stroke-emerald-400/40' 
							: 'stroke-blue-400/25'
					}`}
					stroke-width="0.8"
				/>
			</svg>
			
			<!-- Content layer -->
			<div class="relative z-10 flex flex-col items-center justify-center pt-3">
				<!-- Icon stack - just the main server icon, no badges -->
				<div class="relative mb-2">
					<!-- Main server rack icon only -->
					<Server class={`w-9 h-9 transition-all duration-500 ${
						masterReceiving 
							? 'text-emerald-400 master-icon-ignite' 
							: 'text-blue-400'
					}`} 
					style="filter: drop-shadow(0 0 8px currentColor);" />
				</div>
				
				<!-- Labels -->
				<div class="flex flex-col items-center gap-0.5">
					<span class={` text-[8px] font-bold tracking-widest transition-colors duration-500 ${
						masterReceiving ? 'text-emerald-200' : 'text-blue-200'
					}`}>MASTER</span>
					<div class="flex items-center gap-1">
						<div class={`w-1.5 h-1.5 rounded-full transition-colors duration-800  ${
							masterReceiving ? 'bg-emerald-400 animate-pulse' : 'bg-emerald-200'
						}`}></div>
				<!--<span class="text-[6px] text-emerald-400 font-mono font-semibold">ONLINE</span>-->		
					</div>
				</div>
			</div>

			<!-- Green particles flying in when receiving data -->
			{#if masterReceiving}
				<!-- Particles from top -->
				<div class="absolute -top-16 left-1/2 -translate-x-1/2 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-top shadow-[0_0_10px_rgba(52,211,153,0.8)]"></div>
				<div class="absolute -top-14 left-1/3 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-top" style="animation-delay: 0.05s;"></div>
				<div class="absolute -top-14 right-1/3 w-1.5 h-1.5 rounded-full bg-emerald-300 animate-particle-fly-top" style="animation-delay: 0.1s;"></div>

				<!-- Particles from sides -->
				<div class="absolute top-1/4 -left-16 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-left shadow-[0_0_10px_rgba(74,222,128,0.8)]"></div>
				<div class="absolute top-1/3 -left-14 w-1.5 h-1.5 rounded-full bg-slate-300-400 animate-particle-fly-left" style="animation-delay: 0.07s;"></div>
				<div class="absolute top-1/4 -right-16 w-2 h-2 rounded-full bg-emerald-500 animate-particle-fly-right shadow-[0_0_10px_rgba(16,185,129,0.8)]"></div>
				<div class="absolute top-1/3 -right-14 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-right" style="animation-delay: 0.08s;"></div>
				
				<!-- Particles from bottom -->
				<div class="absolute -bottom-12 left-1/2 -translate-x-1/2 w-2 h-2 rounded-full bg-emerald-400 animate-particle-fly-bottom shadow-[0_0_10px_rgba(52,211,153,0.8)]"></div>
				<div class="absolute -bottom-10 left-1/3 w-1.5 h-1.5 rounded-full bg-green-300 animate-particle-fly-bottom" style="animation-delay: 0.06s;"></div>
				
				<!-- Diagonal particles -->
				<div class="absolute -top-12 -left-12 w-1.5 h-1.5 rounded-full bg-emerald-300 animate-particle-fly-diagonal-tl" style="animation-delay: 0.04s;"></div>
				<div class="absolute -top-12 -right-12 w-1.5 h-1.5 rounded-full bg-green-400 animate-particle-fly-diagonal-tr" style="animation-delay: 0.03s;"></div>
				<div class="absolute -bottom-8 -left-10 w-1.5 h-1.5 rounded-full bg-emerald-400 animate-particle-fly-diagonal-bl" style="animation-delay: 0.09s;"></div>
				<div class="absolute -bottom-8 -right-10 w-1.5 h-1.5 rounded-full bg-green-300 animate-particle-fly-diagonal-br" style="animation-delay: 0.05s;"></div>
			{/if}
		</div>
	</div>

	<!-- Spawner Nodes -->
	{#each $spawners as spawner, i (spawner.id)}
		{@const pos = getPosition(i, $spawners.length)}
		{@const isActive = spawner.status === 'online' || spawner.status === 'Online'}
		{@const utilization = spawner.max_instances > 0 ? (spawner.current_instances / spawner.max_instances) * 100 : 0}

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
				<svg class="absolute -inset-2 w-16 h-16 -rotate-90" style="filter: drop-shadow(0 0 8px rgba(100, 116, 139, 0.3));">
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
						${isActive ? 'bg-slate-800/90 border-slate-600 shadow-slate-500/30' : 'bg-slate-800/90 border-red-500/50 shadow-red-red-500/30'}
						${pulsingSpawners.has(spawner.id) ? 'node-pulse border-emerald-400 shadow-emerald-500/60' : ''}
						transition-all duration-300
					`}
				>
					<Cpu class={`w-6 h-6 ${isActive ? 'text-slate-300' : 'text-slate-500/50'} transition-colors duration-800`} />

					<!-- Status indicator -->
					<div
						class={`absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full border-2 border-slate-900 
							${isActive ? 'bg-emerald-500 status-pulse' : 'bg-red-500/50'}`}
					></div>
				</div>
			</div>

			<!-- Enhanced tooltip -->
			<div
				class="absolute top-16 flex flex-col items-center bg-slate-950/95 backdrop-blur-sm px-4 py-2 rounded-xl border border-slate-700/50 opacity-0 group-hover:opacity-100 transition-opacity duration-200 whitespace-nowrap z-40 pointer-events-none shadow-2xl"
			>
				<span class="text-sm font-bold text-white mb-1">Spawner #{spawner.id}</span>
				<div class="w-full h-px bg-gradient-to-r from-transparent via-slate-600 to-transparent mb-1"></div>
				<div class="flex items-center gap-2 text-xs text-slate-300">
					<Activity class="w-3 h-3" />
					<span>{spawner.region}</span>
				</div>
				<div class="text-xs text-slate-400 mt-1">
					Instances: <span class="font-mono text-slate-200">{spawner.current_instances}</span>/<span class="font-mono">{spawner.max_instances}</span>
					<span class="ml-2 text-[10px]">({utilization.toFixed(0)}%)</span>
				</div>
				<span class={`text-xs font-mono mt-1 font-semibold ${isActive ? 'text-emerald-400' : 'text-red-400'}`}>
					● {spawner.status?.toUpperCase() || 'UNKNOWN'}
				</span>
			</div>
		</div>
	{/each}
		</div>
	</div>
</div>

<!-- Stats overlay (outside zoom wrapper so it stays fixed) -->
<div class="absolute top-4 left-4 flex flex-col gap-2 text-xs font-mono z-30">
	<div class="bg-slate-950/80 backdrop-blur-sm px-3 py-2 rounded-lg">
		<span class="text-slate-400">Active Spawners:</span>
		<span class="text-emerald-400 ml-2 font-bold">
			{$spawners.filter(s => s.status === 'online' || s.status === 'Online').length}/{$spawners.length}
		</span>
	</div>
	<div class="bg-transparent backdrop-blur-sm px-3 py-2 rounded-lg">
		<span class="text-slate-400">Total Instances:</span>
		<span class="text-blue-400 ml-2 font-bold">
			{$spawners.reduce((sum, s) => sum + s.current_instances, 0)}/{$spawners.reduce((sum, s) => sum + s.max_instances, 0)}
		</span>
	</div>
</div>


<style>
	@keyframes gridMove {
		0% {
			background-position: 0 0;
		}
		100% {
			background-position: 40px 40px;
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
		0%, 100% {
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
		0%, 100% {
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
		0%, 100% {
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			transform: translate(-50%, -50%) scale(1.25);
		}
	}

	@keyframes float4 {
		0%, 100% {
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
		0%, 100% {
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
		background: linear-gradient(45deg, transparent 30%, rgba(56, 189, 248, 0.12) 50%, transparent 70%);
		background-size: 200% 200%;
		animation: shimmer 4s ease-in-out infinite;
	}

	@keyframes shimmer {
		0%, 100% {
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
		filter: drop-shadow(0 0 35px rgba(16, 185, 129, 0.7)) drop-shadow(0 0 15px rgba(52, 211, 153, 0.5));
		animation: shieldIgnite 0.8s ease-out;
	}

	@keyframes shieldIgnite {
		0% {
			filter: drop-shadow(0 0 20px rgba(16, 185, 129, 0.4));
		}
		40% {
			filter: drop-shadow(0 0 45px rgba(16, 185, 129, 0.9)) drop-shadow(0 0 25px rgba(52, 211, 153, 0.7));
		}
		100% {
			filter: drop-shadow(0 0 35px rgba(16, 185, 129, 0.7)) drop-shadow(0 0 15px rgba(52, 211, 153, 0.5));
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
		0%, 100% {
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
		0%, 100% {
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

	.fade-overlay-top {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 15%; /* Adjust height for desired fade amount */
		background: linear-gradient(to bottom, rgb(15 23 42 / 1) 0%, rgb(15 23 42 / 0) 100%);
		z-index: 40; /* Ensure it's above other elements but below hints */
	}
	.fade-overlay-bottom {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 15%; /* Adjust height for desired fade amount */
		background: linear-gradient(to top, rgb(15 23 42 / 1) 0%, rgb(15 23 42 / 0) 100%);
		z-index: 40;
	}
	.fade-overlay-left {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		width: 15%; /* Adjust width for desired fade amount */
		background: linear-gradient(to right, rgb(15 23 42 / 1) 0%, rgb(15 23 42 / 0) 100%);
		z-index: 40;
	}
	.fade-overlay-right {
		position: absolute;
		top: 0;
		bottom: 0;
		right: 0;
		width: 15%; /* Adjust width for desired fade amount */
		background: linear-gradient(to left, rgb(15 23 42 / 1) 0%, rgb(15 23 42 / 0) 100%);
		z-index: 40;
	}
</style>