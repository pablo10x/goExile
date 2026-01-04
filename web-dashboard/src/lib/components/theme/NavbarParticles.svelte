<script lang="ts">
	import { onMount } from 'svelte';
	import * as THREE from 'three';
	import { lowPowerMode } from '$lib/stores.svelte';

	let container: HTMLDivElement;
	let renderer: THREE.WebGLRenderer | null = null;
	let scene: THREE.Scene | null = null;
	let camera: THREE.PerspectiveCamera | null = null;
	let mesh: THREE.InstancedMesh | null = null;
	let frameId: number | null = null;
	let resizeObserver: ResizeObserver | null = null;
	let isVisible = true;

	const count = 120;
	const dummy = new THREE.Object3D();
	
	// Particle data
	const particles = new Array(count).fill(0).map(() => ({
		position: new THREE.Vector3(
			(Math.random() - 0.5) * 50, // X: Wide spread
			(Math.random() - 0.5) * 10, // Y: Height
			(Math.random() - 0.5) * 5   // Z: Depth
		),
		velocity: new THREE.Vector3(
			(Math.random() - 0.5) * 0.04 + 0.04, // Drift Right
			(Math.random() - 0.5) * 0.015,        // Slight vertical drift
			0
		),
		rotation: new THREE.Vector3(
			Math.random() * Math.PI,
			Math.random() * Math.PI,
			Math.random() * Math.PI
		),
		rotationSpeed: new THREE.Vector3(
			(Math.random() - 0.5) * 0.015,
			(Math.random() - 0.5) * 0.015,
			(Math.random() - 0.5) * 0.015
		),
		scale: Math.random() * 0.12 + 0.04
	}));

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
		if (resizeObserver) {
			resizeObserver.disconnect();
			resizeObserver = null;
		}
		scene = null;
		camera = null;
		mesh = null;
	}

	function init() {
		if (!container) return;

		const width = container.clientWidth;
		const height = container.clientHeight;

		scene = new THREE.Scene();
		camera = new THREE.PerspectiveCamera(50, width / height, 0.1, 100);
		camera.position.z = 10;

		renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false, powerPreference: 'low-power' });
		renderer.setSize(width, height);
		renderer.setPixelRatio(1); // Force 1x resolution for performance
		container.appendChild(renderer.domElement);

		// Create irregular shard geometry
		const geometry = new THREE.ConeGeometry(0.4, 0.8, 3); // Slightly smaller Triangle shards
		const material = new THREE.MeshBasicMaterial({ 
			color: 0xc2410c, // Tinted Rust
			transparent: true, 
			opacity: 0.4,
			side: THREE.DoubleSide
		});

		mesh = new THREE.InstancedMesh(geometry, material, count);
		mesh.instanceMatrix.setUsage(THREE.DynamicDrawUsage);
		scene.add(mesh);

		let lastTime = 0;
		const targetFPS = 20; // Lower FPS for background decoration
		const frameInterval = 1000 / targetFPS;

		function animate(time: number) {
			frameId = requestAnimationFrame(animate);

			if (!isVisible || $lowPowerMode) return;

			const deltaTime = time - lastTime;
			if (deltaTime < frameInterval) return;
			lastTime = time - (deltaTime % frameInterval);

			if (!mesh) return;

			for (let i = 0; i < count; i++) {
				const p = particles[i];

				// Update Position
				p.position.add(p.velocity);
				p.rotation.add(p.rotationSpeed);

				// Reset if out of bounds
				if (p.position.x > 25) {
					p.position.x = -25;
					p.position.y = (Math.random() - 0.5) * 10;
				}

				dummy.position.copy(p.position);
				dummy.rotation.setFromVector3(p.rotation);
				dummy.scale.setScalar(p.scale);
				dummy.updateMatrix();

				mesh.setMatrixAt(i, dummy.matrix);
			}
			mesh.instanceMatrix.needsUpdate = true;
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

	onMount(() => {
		init();
		
		const handleVisibility = () => {
			isVisible = document.visibilityState === 'visible';
		};
		document.addEventListener('visibilitychange', handleVisibility);

		return () => {
			document.removeEventListener('visibilitychange', handleVisibility);
			cleanup();
		};
	});
</script>

<div bind:this={container} class="absolute inset-0 pointer-events-none opacity-60 mix-blend-screen overflow-hidden"></div>