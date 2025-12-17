<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>3D Game Server Dashboard Background</title>
		<style>
			* {
				margin: 0;
				padding: 0;
				box-sizing: border-box;
			}

			body {
				overflow: hidden;
				background: #0a0a0a;
				font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			}

			#canvas {
				position: fixed;
				top: 0;
				left: 0;
				width: 100%;
				height: 100%;
				z-index: 1;
			}

			.content {
				position: relative;
				z-index: 2;
				padding: 40px;
				color: white;
			}

			.demo-card {
				background: rgba(20, 20, 20, 0.7);
				backdrop-filter: blur(10px);
				border: 1px solid rgba(80, 80, 80, 0.3);
				border-radius: 16px;
				padding: 30px;
				max-width: 400px;
				box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
			}

			h1 {
				font-size: 28px;
				margin-bottom: 10px;
				background: linear-gradient(135deg, #ffffff 0%, #888888 100%);
				-webkit-background-clip: text;
				-webkit-text-fill-color: transparent;
				background-clip: text;
			}

			p {
				color: #a0a0a0;
				line-height: 1.6;
			}
		</style>
	</head>
	<body>
		<canvas id="canvas"></canvas>

		<div class="content">
			<div class="demo-card">
				<h1>Game Server Dashboard</h1>
				<p>
					Advanced 3D animated background with mountain range, floating particles, and spark effects
					from bottom right.
				</p>
			</div>
		</div>

		<script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
		<script>
			let scene,
				camera,
				renderer,
				particles,
				geometries = [],
				sparks = [];
			let mouseX = 0,
				mouseY = 0;
			let targetX = 0,
				targetY = 0;

			function init() {
				scene = new THREE.Scene();
				scene.fog = new THREE.FogExp2(0x0a0a0a, 0.0012);

				camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
				camera.position.set(0, 15, 50);

				renderer = new THREE.WebGLRenderer({
					canvas: document.getElementById('canvas'),
					antialias: true,
					alpha: true
				});
				renderer.setSize(window.innerWidth, window.innerHeight);
				renderer.setPixelRatio(window.devicePixelRatio);

				// Create mountain range in background
				createMountains();

				// Particle system (ambient particles)
				const particleGeometry = new THREE.BufferGeometry();
				const particleCount = 1500;
				const positions = new Float32Array(particleCount * 3);
				const colors = new Float32Array(particleCount * 3);
				const sizes = new Float32Array(particleCount);

				const color1 = new THREE.Color(0xffffff);
				const color2 = new THREE.Color(0x888888);
				const color3 = new THREE.Color(0x555555);

				for (let i = 0; i < particleCount; i++) {
					const i3 = i * 3;
					positions[i3] = (Math.random() - 0.5) * 200;
					positions[i3 + 1] = (Math.random() - 0.5) * 100 + 20;
					positions[i3 + 2] = (Math.random() - 0.5) * 200;

					const colorChoice = Math.random();
					let color;
					if (colorChoice < 0.3) color = color1;
					else if (colorChoice < 0.7) color = color2;
					else color = color3;

					colors[i3] = color.r;
					colors[i3 + 1] = color.g;
					colors[i3 + 2] = color.b;

					sizes[i] = Math.random() * 1.5 + 0.3;
				}

				particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
				particleGeometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
				particleGeometry.setAttribute('size', new THREE.BufferAttribute(sizes, 1));

				const particleMaterial = new THREE.PointsMaterial({
					size: 1,
					vertexColors: true,
					transparent: true,
					opacity: 0.4,
					blending: THREE.AdditiveBlending,
					sizeAttenuation: true
				});

				particles = new THREE.Points(particleGeometry, particleMaterial);
				scene.add(particles);

				// Create spark particles from bottom right
				createSparks();

				// Floating geometries
				createFloatingGeometries();

				// Ambient light
				const ambientLight = new THREE.AmbientLight(0x333333, 0.6);
				scene.add(ambientLight);

				// Point lights with grey/white tones
				const light1 = new THREE.PointLight(0xffffff, 1.2, 100);
				light1.position.set(20, 20, 20);
				scene.add(light1);

				const light2 = new THREE.PointLight(0x777777, 1.5, 100);
				light2.position.set(-20, -20, 20);
				scene.add(light2);

				// Subtle orange light for spark area
				const sparkLight = new THREE.PointLight(0xff8844, 0.8, 60);
				sparkLight.position.set(40, -10, 30);
				scene.add(sparkLight);

				// Grid helper
				const gridHelper = new THREE.GridHelper(150, 30, 0x222222, 0x111111);
				gridHelper.position.y = -15;
				scene.add(gridHelper);

				window.addEventListener('resize', onWindowResize);
				document.addEventListener('mousemove', onMouseMove);

				animate();
			}

			function createMountains() {
				// Create multiple mountain peaks
				const mountainCount = 5;

				for (let m = 0; m < mountainCount; m++) {
					const geometry = new THREE.ConeGeometry(
						15 + Math.random() * 10,
						30 + Math.random() * 20,
						4,
						1
					);

					const material = new THREE.MeshPhongMaterial({
						color: 0x1a1a1a,
						flatShading: true,
						transparent: true,
						opacity: 0.7,
						emissive: 0x0a0a0a,
						emissiveIntensity: 0.2
					});

					const mountain = new THREE.Mesh(geometry, material);
					mountain.position.x = (m - 2) * 30 + (Math.random() - 0.5) * 15;
					mountain.position.y = -15;
					mountain.position.z = -80 - Math.random() * 40;
					mountain.rotation.y = Math.random() * Math.PI * 2;

					scene.add(mountain);
				}
			}

			function createSparks() {
				const sparkGeometry = new THREE.BufferGeometry();
				const sparkCount = 150;
				const positions = new Float32Array(sparkCount * 3);
				const colors = new Float32Array(sparkCount * 3);
				const sizes = new Float32Array(sparkCount);
				const velocities = [];
				const lifetimes = [];

				// Spawn point in bottom right
				const spawnX = 40;
				const spawnY = -15;
				const spawnZ = 30;

				for (let i = 0; i < sparkCount; i++) {
					const i3 = i * 3;

					// Start at spawn point with small variation
					positions[i3] = spawnX + (Math.random() - 0.5) * 2;
					positions[i3 + 1] = spawnY + Math.random() * 2;
					positions[i3 + 2] = spawnZ + (Math.random() - 0.5) * 2;

					// Subtle orange/grey spark colors
					const sparkType = Math.random();
					let r, g, b;
					if (sparkType < 0.4) {
						// Dim orange
						r = 0.8;
						g = 0.4;
						b = 0.2;
					} else if (sparkType < 0.7) {
						// Grey
						r = 0.6;
						g = 0.6;
						b = 0.6;
					} else {
						// Dark orange
						r = 0.6;
						g = 0.3;
						b = 0.1;
					}

					colors[i3] = r;
					colors[i3 + 1] = g;
					colors[i3 + 2] = b;

					sizes[i] = Math.random() * 1.5 + 0.5;

					// Velocity - upward and slightly outward
					velocities.push({
						x: (Math.random() - 0.7) * 0.3,
						y: Math.random() * 0.4 + 0.3,
						z: (Math.random() - 0.5) * 0.2
					});

					lifetimes.push(Math.random());
				}

				sparkGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
				sparkGeometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
				sparkGeometry.setAttribute('size', new THREE.BufferAttribute(sizes, 1));

				const sparkMaterial = new THREE.PointsMaterial({
					size: 2,
					vertexColors: true,
					transparent: true,
					opacity: 0.6,
					blending: THREE.AdditiveBlending,
					sizeAttenuation: true
				});

				const sparkSystem = new THREE.Points(sparkGeometry, sparkMaterial);
				sparkSystem.userData = {
					velocities: velocities,
					lifetimes: lifetimes,
					spawnPoint: { x: spawnX, y: spawnY, z: spawnZ }
				};

				scene.add(sparkSystem);
				sparks.push(sparkSystem);
			}

			function createFloatingGeometries() {
				const geometryTypes = [
					new THREE.OctahedronGeometry(2, 0),
					new THREE.TetrahedronGeometry(2.5, 0),
					new THREE.IcosahedronGeometry(2, 0),
					new THREE.BoxGeometry(3, 3, 3)
				];

				for (let i = 0; i < 12; i++) {
					const geometry = geometryTypes[Math.floor(Math.random() * geometryTypes.length)].clone();

					const colorValue = Math.random();
					let meshColor;
					if (colorValue < 0.3) meshColor = 0x888888;
					else if (colorValue < 0.6) meshColor = 0x555555;
					else meshColor = 0x333333;

					const material = new THREE.MeshPhongMaterial({
						color: meshColor,
						wireframe: Math.random() < 0.7,
						transparent: true,
						opacity: Math.random() * 0.25 + 0.15,
						emissive: meshColor,
						emissiveIntensity: 0.15,
						shininess: 20
					});

					const mesh = new THREE.Mesh(geometry, material);

					mesh.position.x = (Math.random() - 0.5) * 80;
					mesh.position.y = (Math.random() - 0.3) * 60;
					mesh.position.z = (Math.random() - 0.5) * 80;

					mesh.rotation.x = Math.random() * Math.PI;
					mesh.rotation.y = Math.random() * Math.PI;

					mesh.userData = {
						rotationSpeed: {
							x: (Math.random() - 0.5) * 0.015,
							y: (Math.random() - 0.5) * 0.015,
							z: (Math.random() - 0.5) * 0.015
						},
						floatSpeed: Math.random() * 0.015 + 0.008,
						floatOffset: Math.random() * Math.PI * 2
					};

					scene.add(mesh);
					geometries.push(mesh);
				}
			}

			function onMouseMove(event) {
				mouseX = (event.clientX / window.innerWidth) * 2 - 1;
				mouseY = -(event.clientY / window.innerHeight) * 2 + 1;
			}

			function onWindowResize() {
				camera.aspect = window.innerWidth / window.innerHeight;
				camera.updateProjectionMatrix();
				renderer.setSize(window.innerWidth, window.innerHeight);
			}

			function animate() {
				requestAnimationFrame(animate);

				const time = Date.now() * 0.001;

				// Smooth camera movement
				targetX = mouseX * 3;
				targetY = mouseY * 3;
				camera.position.x += (targetX - camera.position.x) * 0.03;
				camera.position.y += (15 + targetY - camera.position.y) * 0.03;
				camera.lookAt(0, 0, 0);

				// Rotate ambient particles slowly
				particles.rotation.y += 0.0003;

				// Animate ambient particle positions
				const positions = particles.geometry.attributes.position.array;
				for (let i = 0; i < positions.length; i += 3) {
					positions[i + 1] += Math.sin(time + positions[i]) * 0.008;
				}
				particles.geometry.attributes.position.needsUpdate = true;

				// Animate sparks
				sparks.forEach((sparkSystem) => {
					const pos = sparkSystem.geometry.attributes.position.array;
					const velocities = sparkSystem.userData.velocities;
					const lifetimes = sparkSystem.userData.lifetimes;
					const spawn = sparkSystem.userData.spawnPoint;

					for (let i = 0; i < pos.length / 3; i++) {
						const i3 = i * 3;

						// Update lifetime
						lifetimes[i] += 0.01;

						if (lifetimes[i] > 1) {
							// Reset spark to spawn point
							pos[i3] = spawn.x + (Math.random() - 0.5) * 2;
							pos[i3 + 1] = spawn.y + Math.random() * 2;
							pos[i3 + 2] = spawn.z + (Math.random() - 0.5) * 2;

							velocities[i].x = (Math.random() - 0.7) * 0.3;
							velocities[i].y = Math.random() * 0.4 + 0.3;
							velocities[i].z = (Math.random() - 0.5) * 0.2;

							lifetimes[i] = 0;
						} else {
							// Update position
							pos[i3] += velocities[i].x;
							pos[i3 + 1] += velocities[i].y;
							pos[i3 + 2] += velocities[i].z;

							// Apply gravity
							velocities[i].y -= 0.008;
						}
					}

					sparkSystem.geometry.attributes.position.needsUpdate = true;
				});

				// Animate floating geometries
				geometries.forEach((mesh, i) => {
					mesh.rotation.x += mesh.userData.rotationSpeed.x;
					mesh.rotation.y += mesh.userData.rotationSpeed.y;
					mesh.rotation.z += mesh.userData.rotationSpeed.z;

					mesh.position.y +=
						Math.sin(time * mesh.userData.floatSpeed + mesh.userData.floatOffset) * 0.04;
				});

				renderer.render(scene, camera);
			}

			init();
		</script>
	</body>
</html>
