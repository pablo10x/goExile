<script lang="ts">
	import { onMount } from 'svelte';
	import * as THREE from 'three';
	import { siteSettings } from '$lib/stores';

	let container: HTMLDivElement;
	let renderer: THREE.WebGLRenderer | null = null;
	let scene: THREE.Scene | null = null;
	let camera: THREE.PerspectiveCamera | null = null;
	let mesh: THREE.InstancedMesh | null = null;
	let frameId: number | null = null;
	let resizeObserver: ResizeObserver | null = null;

	const count = 300;
	const dummy = new THREE.Object3D();
	
	// Particle data
	const particles = new Array(count).fill(0).map(() => ({
		position: new THREE.Vector3(
			(Math.random() - 0.5) * 50, // X: Wide spread
			(Math.random() - 0.5) * 10, // Y: Height
			(Math.random() - 0.5) * 5   // Z: Depth
		),
		velocity: new THREE.Vector3(
			(Math.random() - 0.5) * 0.05 + 0.05, // Drift Right
			(Math.random() - 0.5) * 0.02,        // Slight vertical drift
			0
		),
		rotation: new THREE.Vector3(
			Math.random() * Math.PI,
			Math.random() * Math.PI,
			Math.random() * Math.PI
		),
		rotationSpeed: new THREE.Vector3(
			(Math.random() - 0.5) * 0.02,
			(Math.random() - 0.5) * 0.02,
			(Math.random() - 0.5) * 0.02
		),
		scale: Math.random() * 0.15 + 0.05
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
		if (!container || $siteSettings.performance?.low_power_mode) return;

		const width = container.clientWidth;
		const height = container.clientHeight;

		scene = new THREE.Scene();
		camera = new THREE.PerspectiveCamera(50, width / height, 0.1, 100);
		camera.position.z = 10;

		renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
		renderer.setSize(width, height);
		renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
		container.appendChild(renderer.domElement);

		// Create irregular shard geometry
		const geometry = new THREE.ConeGeometry(0.5, 1, 3); // Triangle shards
		const material = new THREE.MeshBasicMaterial({ 
			color: 0x888888, 
			transparent: true, 
			opacity: 0.8,
			side: THREE.DoubleSide
		});

		mesh = new THREE.InstancedMesh(geometry, material, count);
		scene.add(mesh);

		function animate() {
			if ($siteSettings.performance?.low_power_mode) {
				cleanup();
				return;
			}
			frameId = requestAnimationFrame(animate);

			if (!$siteSettings.aesthetic.animations_enabled) {
				if (renderer && scene && camera) renderer.render(scene, camera);
				return;
			}

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

		animate();

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
		return () => cleanup();
	});

	$effect(() => {
		if ($siteSettings.performance?.low_power_mode) {
			cleanup();
		} else if (!frameId) {
			init();
		}
	});
</script>

<div bind:this={container} class="absolute inset-0 pointer-events-none opacity-60 mix-blend-screen overflow-hidden"></div>
