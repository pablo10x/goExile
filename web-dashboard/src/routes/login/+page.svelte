<script lang="ts">
	import { goto } from '$app/navigation';
	import { isAuthenticated } from '$lib/stores';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Shield, Lock, User, ChevronRight, Activity, Terminal, Database, Cpu, Globe } from 'lucide-svelte';

	let email = $state('admin@example.com');
	let password = $state('admin123');
	let error = $state('');
	let loading = $state(false);
	let mounted = $state(false);
	let formShake = $state(false);
	let isSubmitting = $state(false);

		// --- Decorative Data Feed ---

		const randomHex = (len = 6) => `0x${Math.floor(Math.random() * Math.pow(16, len)).toString(16).padStart(len, '0')}`;

		const randomArrayEl = (arr: any[]) => arr[Math.floor(Math.random() * arr.length)];

	

		const protocols = ['Q-LINK', 'PSY-COM', 'MEM-INJ', 'SUB-SPACE', 'ETH-COM'];

		const statuses = {
			'SYNCED': 'text-emerald-400',
			'DEVIANT': 'text-amber-400',
			'PENDING': 'text-cyan-400',
			'REJECTED': 'text-red-500',
			'CORRUPT': 'text-purple-400',
		};

		const nodes = ['QN-A7B-3', 'SN-GAMMA-9', 'QN-C3D-1', 'SN-ALPHA-2', 'QN-F9E-8'];

		const messages = [

			'CORRELATION: 99.8%',

			'PATTERN MISMATCH',

			'INITIATING HANDSHAKE',

			'ECHO RECEIVED',

			`STREAM ${randomHex(8)}`,

			'DECRYPTION FAILED',

			'AUTH OK'

		];

	

		let dataFeed = $state(Array.from({ length: 25 }, () => ({

			id: Math.random(),

			timestamp: new Date().toLocaleTimeString('en-US', { hour12: false }),

			protocol: randomArrayEl(protocols),

			status: randomArrayEl(Object.keys(statuses)),

			source: randomArrayEl(nodes),

			target: randomArrayEl(nodes.filter(n => n.startsWith('SN'))),

			msg: randomArrayEl(messages)

		})));

	

		onMount(() => {

			mounted = true;

			const feedInterval = setInterval(() => {

				dataFeed.shift();

				dataFeed.push({

					id: Math.random(),

					timestamp: new Date().toLocaleTimeString('en-US', { hour12: false }),

					protocol: randomArrayEl(protocols),

					status: randomArrayEl(Object.keys(statuses)),

					source: randomArrayEl(nodes),

					target: randomArrayEl(nodes.filter(n => n.startsWith('SN'))),

					msg: randomArrayEl(messages)

				});

				dataFeed = dataFeed;

			}, 800);

	

			return () => clearInterval(feedInterval);

		});

	

		async function handleLogin(event: Event) {

			event.preventDefault();

			if (isSubmitting) return;

			isSubmitting = true;

			loading = true;

			error = '';

	

			await new Promise(r => setTimeout(r, 800)); // UX delay

	

			const formData = new URLSearchParams();

			formData.append('email', email);

			formData.append('password', password);

	

			try {

				const response = await fetch('/api/auth/login', {

					method: 'POST',

					body: formData,

					headers: { 'Content-Type': 'application/x-www-form-urlencoded' }

				});

	

				if (response.ok) {

					const data = await response.json();

					if (data.next_step === 'totp') {

						await goto('/login/2fa');

					} else {

						isAuthenticated.set(true);

						await goto('/');

					}

				} else {

					error = 'ACCESS_DENIED :: CREDENTIALS_INVALID';

					formShake = true;

					setTimeout(() => formShake = false, 500);

				}

			} catch (e) {

				error = 'CONNECTION_ERROR :: REGISTRY_UNREACHABLE';

				formShake = true;

				setTimeout(() => formShake = false, 500);

			} finally {

				loading = false;

				isSubmitting = false;

			}

		}

	</script>

	

	{#if mounted}

	<div class="min-h-screen w-full bg-black text-stone-300 font-jetbrains flex">

	

		<!-- Left Panel: Decorative Data Feed -->

		<div class="hidden lg:block w-1/3 bg-stone-950/50 border-r border-stone-800/50 p-6 overflow-hidden relative">

			<div class="absolute inset-0 bg-grid opacity-5"></div>

			<div class="flex items-center gap-3 mb-6">

				<Globe class="w-5 h-5 text-rust animate-pulse" />

				<h2 class="font-bold text-white uppercase tracking-widest">Quantum Entanglement Feed</h2>

			</div>

			<div class="h-full overflow-y-auto no-scrollbar font-mono text-xs" transition:fade>

				{#each dataFeed as item (item.id)}

				<div class="grid grid-cols-[auto,auto,1fr,auto] gap-x-3 items-center" transition:fade|local>

					<span class="text-stone-600">{item.timestamp}</span>

					<span class={`${statuses[item.status as keyof typeof statuses]} font-bold`}>[{item.status}]</span>

					<span class="text-stone-400 truncate">

						{item.source} <span class="text-stone-600">-></span> {item.target}

					</span>

					<span class="text-purple-400">{item.protocol}</span>

				</div>

				{/each}

			</div>

		</div>

	

		<!-- Right Panel: Login Form -->

		<div class="w-full lg:w-2/3 flex items-center justify-center p-8 relative">

			<div class="absolute inset-0 bg-grid opacity-[0.02]"></div>

			<div class:animate-shake={formShake} class="w-full max-w-md">

				<div class="text-center mb-12">

					<div class="inline-block p-4 bg-rust/10 border border-rust/20 rounded-lg mb-4">

						<Cpu class="w-10 h-10 text-rust" />

					</div>

					<h1 class="text-4xl font-black text-white uppercase tracking-tighter">Asset Registry</h1>

					<p class="text-stone-500 mt-2">Secure access required to manage system assets.</p>

				</div>

	

				<form onsubmit={handleLogin} class="space-y-6">

					<div class="group">

						<label for="email" class="text-xs font-bold text-stone-400 uppercase tracking-wider">Operator ID</label>

						<div class="relative mt-2">

							<User class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-stone-600 group-focus-within:text-rust transition-colors" />

							<input

								type="email"

								id="email"

								bind:value={email}

								required

								placeholder="operator@system.node"

								class="w-full pl-12 pr-4 py-4 bg-stone-950 border-2 border-stone-800 rounded-md text-white placeholder:text-stone-600 outline-none focus:border-rust transition-all duration-300"

							/>

						</div>

					</div>

	

					<div class="group">

						<label for="password" class="text-xs font-bold text-stone-400 uppercase tracking-wider">Access Key</label>

						<div class="relative mt-2">

							<Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-stone-600 group-focus-within:text-rust transition-colors" />

							<input

								type="password"

								id="password"

								bind:value={password}

								required

								placeholder="••••••••••••"

								class="w-full pl-12 pr-4 py-4 bg-stone-950 border-2 border-stone-800 rounded-md text-white placeholder:text-stone-600 outline-none focus:border-rust transition-all duration-300"

							/>

						</div>

					</div>

	

					{#if error}

					<div in:fade class="bg-red-900/50 border border-red-700 text-red-400 text-sm rounded-md p-3 flex items-center gap-3">

						<Shield class="w-5 h-5" />

						<span>{error}</span>

					</div>

					{/if}

	

					<button

						type="submit"

						disabled={loading || isSubmitting}

						class="w-full py-4 bg-rust hover:bg-opacity-90 text-white font-bold uppercase tracking-wider rounded-md transition-all duration-300 shadow-[0_4px_20px_rgba(249,115,22,0.2)] hover:shadow-[0_6px_30px_rgba(249,115,22,0.3)] active:translate-y-px disabled:opacity-50 disabled:grayscale disabled:cursor-not-allowed flex items-center justify-center gap-3"

					>

						{#if loading}

							<Activity class="w-5 h-5 animate-spin" />

							<span>Authenticating...</span>

						{:else}

							<span>Authorize Access</span>

							<ChevronRight class="w-5 h-5" />

						{/if}

					</button>

				</form>

	

				<div class="text-center mt-12 text-xs text-stone-600 font-mono">

					<p>Node: {`main-registry-${randomHex()}`}</p>

					<p>Security Protocol v3.1 Active</p>

				</div>

			</div>

		</div>

	</div>

	{/if}

	

	<style>

		@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Teko:wght@700&display=swap');

	

		:global(body) {

			font-family: 'JetBrains Mono', monospace;

		}

	

		h1 {

			font-family: 'Teko', sans-serif;

		}

	

		.bg-grid {

	  		background-image:

	    		linear-gradient(rgba(255, 255, 255, 0.07) 1px, transparent 1px),

	    		linear-gradient(90deg, rgba(255, 255, 255, 0.07) 1px, transparent 1px);

	  		background-size: 30px 30px;

		}

	

		.no-scrollbar::-webkit-scrollbar {

	    	display: none;

		}

		.no-scrollbar {

			-ms-overflow-style: none;

			scrollbar-width: none;

		}

	

		@keyframes shake {

			0%, 100% { transform: translateX(0); }

			10%, 30%, 50%, 70%, 90% { transform: translateX(-6px); }

			20%, 40%, 60%, 80% { transform: translateX(6px); }

		}

		.animate-shake {

			animation: shake 0.4s ease-in-out;

		}

	</style>

	