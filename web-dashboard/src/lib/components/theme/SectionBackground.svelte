<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/state';
	import { backgroundConfig, siteSettings } from '$lib/stores';
	import * as THREE from 'three';

	let { 
		type = 'architecture', 
		mixBlend = 'screen',
		speed = undefined,
		color = undefined,
		density = undefined,
		size = undefined
	} = $props<{
		type?: string;
		mixBlend?: string;
		speed?: number;
		color?: string;
		density?: number;
		size?: number;
	}>();

	let currentSettings = $derived(($backgroundConfig.settings as any)[type] || { intensity: 0.5, speed: 1, density: 1, color: '#f97316', size: 1.0 });

	// Use prop if provided, else use global config, else fallback defaults
	let activeSpeed = $derived(speed ?? currentSettings.speed ?? 1);
	let activeColor = $derived(color ?? currentSettings.color ?? '#f97316');
	let activeDensity = $derived(density ?? currentSettings.density ?? 1);
	let activeIntensity = $derived(currentSettings.intensity ?? 0.5);
	let activeSize = $derived(size ?? currentSettings.size ?? 1.0);

	let perf = $derived($siteSettings.performance || { high_quality_smoke: true, particle_density: 0.8, low_power_mode: false });
	
	const parseThreeColor = (c: string) => {
		try {
			return new THREE.Color(c);
		} catch (e) {
			return new THREE.Color('#ffffff');
		}
	};

	let container: HTMLDivElement;
	let renderer: THREE.WebGLRenderer | null = null;
	let scene: THREE.Scene | null = null;
	let camera: THREE.PerspectiveCamera | null = null;
	let frameId: number | null = null;
	let mesh: THREE.Object3D | null = null;
	let uniforms: any;
	let scrollY = 0;
	let resizeObserver: ResizeObserver | null = null;
	let materials: (THREE.Material | THREE.ShaderMaterial)[] = [];
	let geometries: THREE.BufferGeometry[] = [];

	// Cached data to avoid flickering on settings change
	let cachedParticles = $state<{id: number, x: number, y: number, baseSize: number, baseDuration: number, delay: number}[]>([]);
	let lastInitType = '';
	let lastInitDensity = 0;

	function init() {
		if (!container || type === 'none' || perf.low_power_mode) return;
		
		const isTypeChange = type !== lastInitType;
		const isDensityChange = activeDensity !== lastInitDensity;

		// Clear cached particles if type changed
		if (isTypeChange) {
			cachedParticles = [];
		}

		if (type === 'particle_network') {
			cleanup();
			const count = Math.floor(40 * activeDensity);
			if (cachedParticles.length !== count) {
				cachedParticles = Array.from({ length: count }, (_, i) => ({
					id: i,
					x: Math.random() * 100,
					y: Math.random() * 100,
					baseSize: Math.random() * 4 + 2,
					baseDuration: Math.random() * 4 + 3,
					delay: Math.random() * 3
				}));
			}
			lastInitType = type;
			lastInitDensity = activeDensity;
			return; 
		}

		// For Three.js engines, only full re-init if structure changes
		if (!isTypeChange && !isDensityChange && renderer) {
			return;
		}

		cleanup();

		const width = container.clientWidth;
		const height = container.clientHeight;
		const isMobile = window.innerWidth < 768;

		scene = new THREE.Scene();
		camera = new THREE.PerspectiveCamera(60, width / height, 0.1, 1000);
		camera.position.z = 30;

		renderer = new THREE.WebGLRenderer({ 
			alpha: true, 
			antialias: width > 1000 && !isMobile,
			precision: 'lowp',
			powerPreference: 'low-power'
		});
		renderer.setSize(width, height);
		renderer.setPixelRatio(isMobile ? 1 : Math.min(window.devicePixelRatio, 2));
		container.appendChild(renderer.domElement);

		const themeColor = parseThreeColor(activeColor);
		const mobileDensityMultiplier = isMobile ? 0.5 : 1;

		// --- ARCHITECTURE: Floating Brutalist Blocks ---
		if (type === 'architecture') {
			const group = new THREE.Group();
			const count = Math.floor(25 * activeDensity * mobileDensityMultiplier);
			const geometry = new THREE.BoxGeometry(1, 1, 1);
			geometries.push(geometry);

			for(let i=0; i<count; i++) {
				const isWireframe = Math.random() > 0.7;
				const material = new THREE.MeshBasicMaterial({ 
					color: themeColor,
					wireframe: isWireframe,
					transparent: true,
					opacity: (isWireframe ? 0.05 : 0.1) * activeIntensity * 2
				});
				materials.push(material);
				const box = new THREE.Mesh(geometry, material);
				
				const scaleX = 0.5 + Math.random() * 15;
				const scaleY = 0.5 + Math.random() * 30;
				const scaleZ = 0.5 + Math.random() * 15;
				box.scale.set(scaleX, scaleY, scaleZ);
				
				box.position.set(
					(Math.random() - 0.5) * 80,
					(Math.random() - 0.5) * 60,
					(Math.random() - 0.5) * 50
				);
				box.userData = { 
					rotX: (Math.random() - 0.5) * 0.001, 
					rotY: (Math.random() - 0.5) * 0.001,
					floatSpeed: 0.001 + Math.random() * 0.002,
					floatOffset: Math.random() * Math.PI * 2
				};
				group.add(box);
			}
			mesh = group;
		}

		// --- TACTICAL_GRID: Radar Scan Line ---
		else if (type === 'tactical_grid') {
			const group = new THREE.Group();
			const geometry = new THREE.PlaneGeometry(100, 100, 1, 1);
			geometries.push(geometry);
			const material = new THREE.MeshBasicMaterial({ 
				color: themeColor, 
				transparent: true, 
				opacity: 0.05 * activeIntensity,
				wireframe: true 
			});
			materials.push(material);
			const grid = new THREE.Mesh(geometry, material);
			grid.rotation.x = -Math.PI / 2;
			group.add(grid);

			const scanGeo = new THREE.PlaneGeometry(100, 2, 1, 1);
			geometries.push(scanGeo);
			const scanMat = new THREE.MeshBasicMaterial({ 
				color: themeColor, 
				transparent: true, 
				opacity: 0.3 * activeIntensity,
				side: THREE.DoubleSide
			});
			materials.push(scanMat);
			const scan = new THREE.Mesh(scanGeo, scanMat);
			scan.rotation.x = -Math.PI / 2;
			group.add(scan);
			group.userData = { scanNode: scan };
			mesh = group;
		}

		// --- NEURAL_NETWORK: Plexus Connection Nodes ---
		else if (type === 'neural_network') {
			const count = Math.floor(150 * activeDensity * perf.particle_density * mobileDensityMultiplier);
			const positions = new Float32Array(count * 3);
			const velocities = [];
			for(let i=0; i<count; i++) {
				positions[i*3] = (Math.random() - 0.5) * 60;
				positions[i*3+1] = (Math.random() - 0.5) * 60;
				positions[i*3+2] = (Math.random() - 0.5) * 60;
				velocities.push(new THREE.Vector3((Math.random()-0.5)*0.015, (Math.random()-0.5)*0.015, (Math.random()-0.5)*0.015));
			}

			const geometry = new THREE.BufferGeometry();
			geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
			geometries.push(geometry);

			const material = new THREE.PointsMaterial({ 
				color: themeColor, 
				size: isMobile ? 0.4 : 0.6, 
				transparent: true, 
				opacity: 0.6 * activeIntensity
			});
			materials.push(material);

			const points = new THREE.Points(geometry, material);
			
			// Lines for connections
			const lineGeo = new THREE.BufferGeometry();
			lineGeo.setAttribute('position', new THREE.BufferAttribute(new Float32Array(count * 10 * 2 * 3), 3));
			geometries.push(lineGeo);
			const lineMat = new THREE.LineBasicMaterial({ 
				color: themeColor, 
				transparent: true, 
				opacity: 0.08 * activeIntensity,
				blending: THREE.AdditiveBlending 
			});
			materials.push(lineMat);
			const lines = new THREE.LineSegments(lineGeo, lineMat);

			const group = new THREE.Group();
			group.add(points);
			group.add(lines);
			group.userData = { velocities, count };
			mesh = group;
		}

		// --- DATA_FLOW: Cascading Information Blocks ---
		else if (type === 'data_flow') {
			const count = Math.floor(200 * activeDensity * mobileDensityMultiplier);
			const geometry = new THREE.BoxGeometry(0.2, 0.2, 0.2);
			geometries.push(geometry);
			const group = new THREE.Group();
			for(let i=0; i<count; i++) {
				const material = new THREE.MeshBasicMaterial({ 
					color: themeColor, 
					transparent: true, 
					opacity: Math.random() * 0.3 * activeIntensity
				});
				materials.push(material);
				const cube = new THREE.Mesh(geometry, material);
				cube.position.set((Math.random()-0.5)*60, (Math.random()-0.5)*60, (Math.random()-0.5)*40);
				cube.scale.y = 1 + Math.random() * 5;
				cube.userData = { speed: (0.05 + Math.random() * 0.2) * activeSpeed };
				group.add(cube);
			}
			mesh = group;
		}

		// --- DIGITAL_HORIZON: Infinite Scrolling Grid ---
		else if (type === 'digital_horizon') {
			const segs = isMobile ? 40 : 100;
			const geometry = new THREE.PlaneGeometry(200, 200, segs, segs);
			geometries.push(geometry);
			
			uniforms = {
				uTime: { value: 0 },
				uColor: { value: themeColor },
				uScroll: { value: 0 },
				uIntensity: { value: activeIntensity }
			};

			const material = new THREE.ShaderMaterial({
				uniforms,
				wireframe: true,
				transparent: true,
				vertexShader: `
					uniform float uTime;
					uniform float uScroll;
					varying float vDist;
					void main() {
						vec3 pos = position;
						float noise = sin(pos.x * 0.05 + uTime * 0.5) * cos(pos.y * 0.05 + uTime * 0.5) * 4.0;
						pos.z += noise;
						vDist = 1.0 - (length(pos.xy) / 140.0);
						gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
					}
				`,
				fragmentShader: `
					uniform vec3 uColor;
					uniform float uIntensity;
					varying float vDist;
					void main() {
						float scanline = sin(gl_FragCoord.y * 0.5) * 0.1 + 0.9;
						gl_FragColor = vec4(uColor, vDist * 0.15 * uIntensity * 2.0 * scanline);
					}
				`
			});
			materials.push(material);
			const plane = new THREE.Mesh(geometry, material);
			plane.rotation.x = -Math.PI / 2.2;
			plane.position.y = -20;
			mesh = plane;
		}

		// --- CYBER_OCEAN: Rolling Waves of Data ---
		else if (type === 'cyber_ocean') {
			const segs = isMobile ? 40 : 80;
			const geometry = new THREE.PlaneGeometry(150, 150, segs, segs);
			geometries.push(geometry);

			uniforms = {
				uTime: { value: 0 },
				uColor: { value: themeColor },
				uIntensity: { value: activeIntensity }
			};

			const material = new THREE.ShaderMaterial({
				uniforms,
				wireframe: true,
				transparent: true,
				vertexShader: `
					uniform float uTime;
					varying float vHeight;
					varying float vDist;
					void main() {
						vec3 pos = position;
						float h = sin(pos.x * 0.1 + uTime) * cos(pos.y * 0.1 + uTime * 0.5) * 5.0;
						h += sin(pos.x * 0.05 - uTime * 0.3) * 2.0;
						pos.z += h;
						vHeight = h;
						vDist = 1.0 - (length(pos.xy) / 100.0);
						gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
					}
				`,
				fragmentShader: `
					uniform vec3 uColor;
					uniform float uIntensity;
					varying float vHeight;
					varying float vDist;
					void main() {
						float alpha = (vHeight + 7.0) / 14.0;
						gl_FragColor = vec4(uColor, alpha * 0.1 * uIntensity * 2.0 * vDist);
					}
				`
			});
			materials.push(material);
			const ocean = new THREE.Mesh(geometry, material);
			ocean.rotation.x = -Math.PI / 2.1;
			ocean.position.y = -15;
			mesh = ocean;
		}
		// --- STATIC_VOID: Noise Field ---
		else if (type === 'static_void') {
			const count = Math.floor(1000 * activeDensity * mobileDensityMultiplier);
			const positions = new Float32Array(count * 3);
			for(let i=0; i<count; i++) {
				positions[i*3] = (Math.random() - 0.5) * 100;
				positions[i*3+1] = (Math.random() - 0.5) * 100;
				positions[i*3+2] = (Math.random() - 0.5) * 100;
			}
			const geometry = new THREE.BufferGeometry();
			geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
			geometries.push(geometry);

			const material = new THREE.PointsMaterial({
				color: themeColor,
				size: 0.1,
				transparent: true,
				opacity: 0.3 * activeIntensity
			});
			materials.push(material);
			mesh = new THREE.Points(geometry, material);
		}

		if (mesh) scene.add(mesh);
		lastInitType = type;
		lastInitDensity = activeDensity;

		let lastTime = 0;
		const targetFPS = isMobile ? 30 : 60;
		const frameInterval = 1000 / targetFPS;

		function animate(time: number) {
			if (type === 'none' || perf.low_power_mode) {
				cleanup();
				return;
			}
			
			if ($siteSettings.aesthetic?.reduced_motion) {
				if (renderer && scene && camera) renderer.render(scene, camera);
				return;
			}

			frameId = requestAnimationFrame(animate);
			
			if (!$siteSettings.aesthetic.animations_enabled) {
				if (renderer && scene && camera) renderer.render(scene, camera);
				return;
			}

			const deltaTime = time - lastTime;
			if (deltaTime < frameInterval) return;
			lastTime = time - (deltaTime % frameInterval);

			const calcTime = (Date.now() * 0.001) * activeSpeed;

			// Live update uniforms/materials without re-init
			const themeColor = parseThreeColor(activeColor);
			if (uniforms) {
				uniforms.uTime.value = calcTime;
				uniforms.uScroll.value = scrollY;
				if (uniforms.uColor) uniforms.uColor.value = themeColor;
				if (uniforms.uIntensity) uniforms.uIntensity.value = activeIntensity;
			}

			// Update non-shader materials
			materials.forEach(mat => {
				if (mat instanceof THREE.MeshBasicMaterial || mat instanceof THREE.PointsMaterial || mat instanceof THREE.LineBasicMaterial) {
					mat.color.copy(themeColor);
					// Logic for opacity update could be added here if needed, 
					// but usually requires knowledge of the base opacity.
				}
			});

			if (type === 'architecture' && mesh instanceof THREE.Group) {
				mesh.rotation.y = calcTime * 0.05;
				mesh.children.forEach(child => {
					child.rotation.x += child.userData.rotX * activeSpeed;
					child.rotation.y += child.userData.rotY * activeSpeed;
				});
			} 
			else if (type === 'tactical_grid' && mesh instanceof THREE.Group) {
				const scan = mesh.userData.scanNode;
				if (scan) {
					scan.position.z = ((calcTime * 10) % 100) - 50;
				}
				mesh.rotation.y = Math.sin(calcTime * 0.1) * 0.1;
			}
			else if (type === 'neural_network' && mesh instanceof THREE.Group) {
				const points = mesh.children[0] as THREE.Points;
				const posAttr = points.geometry.attributes.position;
				const vels = mesh.userData.velocities;
				
				for(let i=0; i<mesh.userData.count; i++) {
					posAttr.setX(i, posAttr.getX(i) + vels[i].x * activeSpeed);
					posAttr.setY(i, posAttr.getY(i) + vels[i].y * activeSpeed);
					posAttr.setZ(i, posAttr.getZ(i) + vels[i].z * activeSpeed);

					if (Math.abs(posAttr.getX(i)) > 25) vels[i].x *= -1;
					if (Math.abs(posAttr.getY(i)) > 25) vels[i].y *= -1;
					if (Math.abs(posAttr.getZ(i)) > 25) vels[i].z *= -1;
				}
				posAttr.needsUpdate = true;
				mesh.rotation.y = calcTime * 0.02;
			}
			else if (type === 'data_flow' && mesh instanceof THREE.Group) {
				mesh.children.forEach(child => {
					child.position.y -= child.userData.speed;
					if (child.position.y < -30) child.position.y = 30;
				});
			}

			if (renderer && scene && camera) renderer.render(scene, camera);
		}
		frameId = requestAnimationFrame(animate);

		resizeObserver = new ResizeObserver(() => {
			if (!container || !camera || !renderer) return;
			const w = container.clientWidth;
			const h = container.clientHeight;
			camera.aspect = w / h;
			camera.updateProjectionMatrix();
			renderer.setSize(w, h);
		});
		resizeObserver.observe(container);
	}

	function cleanup() {
		if (frameId !== null) {
			cancelAnimationFrame(frameId);
			frameId = null;
		}
		if (renderer) {
			renderer.dispose();
			if (renderer.domElement && renderer.domElement.parentElement) {
				renderer.domElement.parentElement.removeChild(renderer.domElement);
			}
			renderer = null;
		}
		geometries.forEach(g => g.dispose());
		materials.forEach(m => m.dispose());
		geometries = [];
		materials = [];
		if (resizeObserver) {
			resizeObserver.disconnect();
			resizeObserver = null;
		}
		scene = null;
		camera = null;
	}

	onMount(() => {
		const handleScroll = () => { scrollY = window.scrollY; };
		window.addEventListener('scroll', handleScroll);
		return () => {
			window.removeEventListener('scroll', handleScroll);
			cleanup();
		};
	});

	$effect(() => {
		const t = type;
		const d = activeDensity;
		const p = perf.low_power_mode;
		untrack(() => {
			// Only re-init if structure changes (type or density)
			init();
		});
	});
</script>

{#if type === 'particle_network'}
	<div class="absolute inset-0 pointer-events-none overflow-hidden" style="opacity: {activeIntensity}">
		<!-- Gradient Blobs -->
		<div class="absolute inset-0 overflow-hidden">
			<div 
				class="absolute rounded-full blur-[60px] opacity-60 animate-blob-float-1"
				style="width: 400px; height: 400px; top: -100px; right: -100px; background: radial-gradient(circle, {activeColor}20, transparent 70%);"
			></div>
			<div 
				class="absolute rounded-full blur-[60px] opacity-60 animate-blob-float-2"
				style="width: 350px; height: 350px; bottom: -80px; left: -80px; background: radial-gradient(circle, {activeColor}15, transparent 70%);"
			></div>
			<div 
				class="absolute rounded-full blur-[60px] opacity-60 animate-blob-float-3"
				style="width: 300px; height: 300px; top: 50%; left: 50%; transform: translate(-50%, -50%); background: radial-gradient(circle, {activeColor}10, transparent 70%);"
			></div>
		</div>

		<!-- Particles -->
		<div class="absolute inset-0">
			{#each cachedParticles as particle (particle.id)}
				<div
					class="absolute rounded-full pointer-events-none animate-particle-float"
					style="
						background-color: {activeColor};
						left: {particle.x}%;
						top: {particle.y}%;
						width: {particle.baseSize * activeSize}px;
						height: {particle.baseSize * activeSize}px;
						animation-duration: {particle.baseDuration / activeSpeed}s;
						animation-delay: {particle.delay}s;
						box-shadow: 0 0 10px {activeColor};
					"
				></div>
			{/each}
		</div>
	</div>
{/if}

<div bind:this={container} class="absolute inset-0 pointer-events-none -z-10 overflow-hidden" style="mix-blend-mode: {mixBlend}"></div>
