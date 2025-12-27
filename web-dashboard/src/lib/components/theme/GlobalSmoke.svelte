<script lang="ts">
	import { onMount } from 'svelte';
	import * as THREE from 'three';
	import { siteSettings } from '$lib/stores';

	let container: HTMLDivElement;
	let renderer: THREE.WebGLRenderer | null = null;
	let scene: THREE.Scene | null = null;
	let camera: THREE.PerspectiveCamera | null = null;
	let frameId: number | null = null;
	let smokeParticles: THREE.Group | null = null;
	let resizeObserver: ResizeObserver | null = null;

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
		smokeParticles = null;
	}

	function init() {
		if (!container || $siteSettings.performance?.low_power_mode) return;

		const width = window.innerWidth;
		const height = window.innerHeight;
		const isMobile = width < 768;

		scene = new THREE.Scene();
		// Subtle blue-ish fog for depth
		scene.fog = new THREE.FogExp2(0x000000, 0.001);

		camera = new THREE.PerspectiveCamera(75, width / height, 1, 1000);
		camera.position.z = 20;

		renderer = new THREE.WebGLRenderer({ 
			alpha: true,
			antialias: false,
			precision: 'lowp',
			powerPreference: 'low-power'
		});
		renderer.setPixelRatio(isMobile ? 1 : Math.min(window.devicePixelRatio, 2));
		renderer.setSize(window.innerWidth, window.innerHeight);
		container.appendChild(renderer.domElement);

		// Smoke Texture Generation
		const canvas = document.createElement('canvas');
		canvas.width = 32; canvas.height = 32;
		const context = canvas.getContext('2d');
		if (context) {
			const gradient = context.createRadialGradient(16, 16, 0, 16, 16, 16);
			gradient.addColorStop(0, 'rgba(200,200,210,0.15)'); // Lighter center
			gradient.addColorStop(1, 'rgba(0,0,0,0)');
			context.fillStyle = gradient;
			context.fillRect(0, 0, 32, 32);
		}
		const texture = new THREE.CanvasTexture(canvas);

		const material = new THREE.MeshBasicMaterial({
			map: texture,
			transparent: true,
			opacity: 0.4,
			depthWrite: false,
			blending: THREE.AdditiveBlending,
			side: THREE.DoubleSide
		});

		const geometry = new THREE.PlaneGeometry(15, 15);
		smokeParticles = new THREE.Group();

		// Reduced count for mobile
		const particleCount = isMobile ? 20 : 80;

		for (let i = 0; i < particleCount; i++) {
			const particle = new THREE.Mesh(geometry, material);
			particle.position.set(
				(Math.random() - 0.5) * 40,
				(Math.random() - 0.5) * 20,
				(Math.random() - 0.5) * 40
			);
			particle.rotation.z = Math.random() * Math.PI;
			particle.userData = {
				velocity: (Math.random() - 0.5) * 0.02,
				drift: (Math.random() - 0.5) * 0.01
			};
			smokeParticles.add(particle);
		}

		scene.add(smokeParticles);

		// Scroll interaction
		let targetScrollY = 0;
		const updateScroll = () => {
			targetScrollY = window.scrollY * 0.01;
		};
		window.addEventListener('scroll', updateScroll, { passive: true });

		let lastTime = 0;
		const targetFPS = isMobile ? 30 : 60;
		const frameInterval = 1000 / targetFPS;

		function animate(time: number) {
			if ($siteSettings.performance?.low_power_mode) {
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

			if (smokeParticles) {
				smokeParticles.children.forEach(p => {
					p.rotation.z += p.userData.velocity;
					p.position.x += p.userData.drift;
					
					// Reset position if too far
					if (p.position.x > 20) p.position.x = -20;
					if (p.position.x < -20) p.position.x = 20;
				});
			}

			// Gentle camera movement based on scroll
			if (camera) camera.position.y = THREE.MathUtils.lerp(camera.position.y, -targetScrollY, 0.05);

			if (renderer && scene && camera) renderer.render(scene, camera);
		}
		frameId = requestAnimationFrame(animate);

		resizeObserver = new ResizeObserver(() => {
			if (!container || !camera || !renderer) return;
			const w = window.innerWidth;
			const h = window.innerHeight;
			camera.aspect = w / h;
			camera.updateProjectionMatrix();
			renderer.setSize(w, h);
		});
		resizeObserver.observe(document.body);

		return () => {
			window.removeEventListener('scroll', updateScroll);
			cleanup();
		};
	}

	onMount(() => {
		const unsub = init();
		return () => {
			if (typeof unsub === 'function') unsub();
			cleanup();
		};
	});

	$effect(() => {
		if ($siteSettings.performance?.low_power_mode) {
			cleanup();
		} else if (!frameId) {
			init();
		}
	});
</script>

<div bind:this={container} class="fixed inset-0 pointer-events-none z-0 overflow-hidden mix-blend-screen"></div>
