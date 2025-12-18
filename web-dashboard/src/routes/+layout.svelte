<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		isAuthenticated,
		restartRequired,
		stats,
		spawners,
		serverVersions,
		isConnected,
		connectionStatus
	} from '$lib/stores';
	import ToastContainer from '$lib/components/ToastContainer.svelte';

	let { children } = $props();
	let isChecking = $state(true);
	let restarting = $state(false);
	let eventSource: EventSource | null = null;

	// Animation states
	let sidebarLoaded = $state(false);
	let mouseX = $state(0);
	let mouseY = $state(0);
	let hoveredItem = $state(-1);
	let isSidebarCollapsed = $state(true);

	function connectSSE() {
		if (typeof window === 'undefined') return;
		if (eventSource) eventSource.close();

		eventSource = new EventSource('/events');

		eventSource.onopen = () => {
			isConnected.set(true);
			connectionStatus.set('Live (SSE)');
		};

		eventSource.onerror = () => {
			isConnected.set(false);
			connectionStatus.set('Reconnecting...');
		};

		eventSource.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				// console.log('SSE Message:', data.type); // Uncomment for verbose debugging
				if (data.type === 'stats') {
					stats.set(data.payload);
				} else if (data.type === 'spawners') {
					const list: any[] = Array.isArray(data.payload)
						? data.payload
						: Object.values(data.payload);
					spawners.set(list);
				}
			} catch (e) {
				console.error('SSE Parse Error', e);
			}
		};
	}

	async function checkAuth() {
		try {
			const res = await fetch('/api/stats', { cache: 'no-store', credentials: 'include' });
			if (res.ok) {
				isAuthenticated.set(true);
				connectSSE(); // Connect SSE on auth success
			} else {
				isAuthenticated.set(false);
				if (window.location.pathname !== '/login') {
					goto('/login');
				}
			}
		} catch (e) {
			isAuthenticated.set(false);
			if (window.location.pathname !== '/login') {
				goto('/login');
			}
		} finally {
			isChecking = false;
		}
	}

	async function initialFetch() {
		try {
			const [statsRes, spawnersRes, versionsRes] = await Promise.all([
				fetch('/api/stats', { cache: 'no-store', credentials: 'include' }),
				fetch('/api/spawners', { cache: 'no-store', credentials: 'include' }),
				fetch('/api/versions', { cache: 'no-store', credentials: 'include' })
			]);
			if (statsRes.ok) stats.set(await statsRes.json());
			if (versionsRes.ok) serverVersions.set(await versionsRes.json());
			if (spawnersRes.ok) spawners.set(await spawnersRes.json());
		} catch (e) {
			console.error('Initial fetch failed', e);
		}
	}

	async function restartServer() {
		if (
			!confirm(
				'Are you sure you want to restart the server? This will interrupt all active connections.'
			)
		)
			return;

		try {
			restarting = true;
			await fetch('/api/restart', { method: 'POST' });
			alert('Server restart triggered. The dashboard will reload in a moment.');
			setTimeout(() => {
				window.location.reload();
			}, 5000);
		} catch (e) {
			alert('Failed to trigger restart: ' + e);
			restarting = false;
		}
	}

	onMount(async () => {
		await checkAuth();
		if ($isAuthenticated) {
			initialFetch();
		}

		// Trigger sidebar animations after mount
		setTimeout(() => {
			sidebarLoaded = true;
		}, 300);
	});

	onDestroy(() => {
		if (eventSource) eventSource.close();
	});

	async function logout() {
		await fetch('/logout');
		isAuthenticated.set(false);
		goto('/login');
	}

	function isRouteActive(path: string) {
		return page.url.pathname === path;
	}

	function toggleSidebar() {
		isSidebarCollapsed = !isSidebarCollapsed;
	}
</script>

{#if isChecking}
	<div
		class="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950"
	>
		<div class="relative">
			<div
				class="animate-spin rounded-full h-16 w-16 border-4 border-slate-700 border-t-indigo-500 shadow-2xl"
			></div>
			<div
				class="absolute inset-0 rounded-full bg-gradient-to-r from-indigo-500/20 to-purple-500/20 blur-xl animate-pulse"
			></div>
		</div>
	</div>
{:else}
	{#if $isAuthenticated && page.url.pathname !== '/login'}
		<div class="flex h-screen text-slate-300 overflow-hidden relative">
			<!-- Global Restart Banner (Mobile adjusted) -->
			{#if $restartRequired}
				<div
					class="absolute top-0 md:left-64 left-0 right-0 z-50 bg-orange-600/90 backdrop-blur-md text-white px-4 py-2 flex justify-between items-center shadow-lg border-b border-orange-500/50 animate-slide-fade text-xs md:text-sm"
				>
					<div class="flex items-center gap-2">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="w-4 h-4 animate-pulse shrink-0"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path
								d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
							></path>
							<line x1="12" y1="9" x2="12" y2="13"></line>
							<line x1="12" y1="17" x2="12.01" y2="17"></line>
						</svg>
						<span class="font-medium truncate">Restart required.</span>
					</div>
					<button
						onclick={restartServer}
						disabled={restarting}
						class="px-3 py-1 bg-white text-orange-600 rounded font-bold hover:bg-orange-50 transition-colors shadow-sm text-xs disabled:opacity-50 whitespace-nowrap"
					>
						{restarting ? '...' : 'Restart'}
					</button>
				</div>
			{/if}

			<!-- Desktop Sidebar (Hidden on Mobile) -->
			<aside
				class="hidden md:flex relative transition-all duration-300 ease-in-out bg-slate-950 border-r border-slate-800 flex-col shrink-0 overflow-hidden shadow-2xl z-20 {isSidebarCollapsed ? 'w-20' : 'w-64'}"
			>
				<!-- Glass Surface -->
				<div
					class="absolute inset-0 bg-slate-900/60 backdrop-blur-[1px] border-r border-white/5"
				></div>

				<!-- Content Container -->
				<div class="relative z-10 flex flex-col h-full">
					<div
						class="p-6 border-b border-white/5 bg-white/5 backdrop-blur-md transform transition-all duration-700 {sidebarLoaded
							? 'translate-y-0 opacity-100'
							: '-translate-y-4 opacity-0'} flex items-center justify-between"
					>
						{#if !isSidebarCollapsed}
							<div class="flex items-center gap-3 animate-in fade-in zoom-in duration-300">
								<div
									class="w-3 h-3 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full animate-pulse shadow-lg shadow-blue-500/50"
								></div>
								<h1
									class="text-2xl font-bold text-slate-50 -tracking-wide bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent filter drop-shadow-lg animate-gradient bg-size-200"
								>
									GoExile
								</h1>
							</div>
						{/if}
						<button 
							onclick={toggleSidebar}
							class="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors {isSidebarCollapsed ? 'mx-auto' : ''}"
							title={isSidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								{#if isSidebarCollapsed}
									<polyline points="13 17 18 12 13 7"></polyline>
									<polyline points="6 17 11 12 6 7"></polyline>
								{:else}
									<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
									<line x1="9" y1="3" x2="9" y2="21"></line>
								{/if}
							</svg>
						</button>
					</div>

					<nav class="flex-1 p-4 space-y-2 overflow-y-auto overflow-x-hidden">
						<!-- Dashboard (Management) -->
						<a
							href="/dashboard"
							class="nav-link {isSidebarCollapsed ? 'justify-center px-2' : ''}"
							class:nav-active={isRouteActive('/dashboard') || isRouteActive('/')}
							onmouseenter={() => (hoveredItem = 6)}
							onmouseleave={() => (hoveredItem = -1)}
							style="animation-delay: 0.1s;"
							title={isSidebarCollapsed ? "Dashboard" : ""}
						>
							{#if isRouteActive('/dashboard') || isRouteActive('/')}
								<div class="nav-indicator"></div>
							{/if}
							<svg
								class="nav-icon"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<rect x="3" y="3" width="7" height="7"></rect>
								<rect x="14" y="3" width="7" height="7"></rect>
								<rect x="14" y="14" width="7" height="7"></rect>
								<rect x="3" y="14" width="7" height="7"></rect>
							</svg>
							{#if !isSidebarCollapsed}
								<span class="animate-in fade-in slide-in-from-left-2 duration-300">Dashboard</span>
							{/if}
						</a>

						<!-- Performance -->
						<a
							href="/performance"
							class="nav-link {isSidebarCollapsed ? 'justify-center px-2' : ''}"
							class:nav-active={isRouteActive('/performance')}
							onmouseenter={() => (hoveredItem = 1)}
							onmouseleave={() => (hoveredItem = -1)}
							style="animation-delay: 0.2s;"
							title={isSidebarCollapsed ? "Performance" : ""}
						>
							{#if isRouteActive('/performance')}
								<div class="nav-indicator"></div>
							{/if}
							<svg
								class="nav-icon"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
							</svg>
							{#if !isSidebarCollapsed}
								<span class="animate-in fade-in slide-in-from-left-2 duration-300">Performance</span>
							{/if}
						</a>

						<!-- Configuration -->
						<a
							href="/config"
							class="nav-link {isSidebarCollapsed ? 'justify-center px-2' : ''}"
							class:nav-active={isRouteActive('/config') || isRouteActive('/config/')}
							onmouseenter={() => (hoveredItem = 2)}
							onmouseleave={() => (hoveredItem = -1)}
							style="animation-delay: 0.3s;"
							title={isSidebarCollapsed ? "Configuration" : ""}
						>
							{#if isRouteActive('/config') || isRouteActive('/config/')}
								<div class="nav-indicator"></div>
							{/if}
							<svg
								class="nav-icon"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<circle cx="12" cy="12" r="3"></circle>
								<path
									d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"
								></path>
							</svg>
							{#if !isSidebarCollapsed}
								<span class="animate-in fade-in slide-in-from-left-2 duration-300">Configuration</span>
							{/if}
						</a>

						<!-- Server Files -->
						<a
							href="/server"
							class="nav-link {isSidebarCollapsed ? 'justify-center px-2' : ''}"
							class:nav-active={isRouteActive('/server')}
							onmouseenter={() => (hoveredItem = 3)}
							onmouseleave={() => (hoveredItem = -1)}
							style="animation-delay: 0.4s;"
							title={isSidebarCollapsed ? "Server Files" : ""}
						>
							{#if isRouteActive('/server')}
								<div class="nav-indicator"></div>
							{/if}
							<svg
								class="nav-icon"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
								<polyline points="7 10 12 15 17 10"></polyline>
								<line x1="12" y1="15" x2="12" y2="3"></line>
							</svg>
							{#if !isSidebarCollapsed}
								<span class="animate-in fade-in slide-in-from-left-2 duration-300">Server Files</span>
							{/if}
						</a>

						<!-- Database -->
						<a
							href="/database"
							class="nav-link {isSidebarCollapsed ? 'justify-center px-2' : ''}"
							class:nav-active={isRouteActive('/database')}
							onmouseenter={() => (hoveredItem = 4)}
							onmouseleave={() => (hoveredItem = -1)}
							style="animation-delay: 0.5s;"
							title={isSidebarCollapsed ? "Databases" : ""}
						>
							{#if isRouteActive('/database')}
								<div class="nav-indicator"></div>
							{/if}
							<svg
								class="nav-icon"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
								<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
								<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
								<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
							</svg>
							{#if !isSidebarCollapsed}
								<span class="animate-in fade-in slide-in-from-left-2 duration-300">Databases</span>
							{/if}
						</a>
					</nav>

					<div
						class="p-4 border-t border-white/5 bg-black/20 backdrop-blur-md flex flex-col gap-2 transform transition-all duration-700 {sidebarLoaded
							? 'translate-y-0 opacity-100'
							: 'translate-y-8 opacity-0'}"
						style="animation-delay: 0.6s;"
					>
						{#if !isSidebarCollapsed}
							<!-- Notifications (Future panel trigger) -->
							<button
								class="w-full flex items-center gap-3 px-4 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors animate-in fade-in zoom-in duration-300"
							>
								<div class="relative">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="w-5 h-5"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										stroke-linecap="round"
										stroke-linejoin="round"
									>
										<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
										<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
									</svg>
									<!-- Example Badge -->
									<span class="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full animate-pulse"
									></span>
								</div>
								<span class="text-sm font-medium">Notifications</span>
							</button>
						{/if}

						<button
							onclick={logout}
							onmouseenter={() => (hoveredItem = 5)}
							onmouseleave={() => (hoveredItem = -1)}
							class="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg text-slate-400 hover:text-white border border-transparent transition-all duration-500 group relative overflow-hidden backdrop-blur-sm
                            {hoveredItem === 5
								? 'bg-gradient-to-r from-red-500/20 to-red-600/20 border-red-500/40 shadow-lg shadow-red-500/20 scale-[1.02]'
								: 'hover:bg-red-500/10 hover:border-red-500/30'}"
							title={isSidebarCollapsed ? "Logout" : ""}
						>
							<!-- Animated Background -->
							<div
								class="absolute inset-0 bg-gradient-to-r from-red-600/0 via-red-600/15 to-red-600/0 translate-x-[-100%] {hoveredItem ===
								5
									? 'translate-x-[100%]'
									: ''} transition-transform duration-700"
							></div>

							<!-- Warning Pulse on Hover -->
							<div
								class="absolute inset-0 {hoveredItem === 5
									? 'animate-pulse'
									: ''} bg-red-500/5 rounded-lg"
							></div>

							<div class="relative z-10 flex items-center justify-center gap-2">
								<div
									class="w-4 h-4 flex items-center justify-center {hoveredItem === 5
										? 'animate-bounce'
										: ''}"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="w-4 h-4 transition-all duration-500 group-hover:-translate-x-1 group-hover:scale-110"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										stroke-linecap="round"
										stroke-linejoin="round"
									>
										<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
										<polyline points="16 17 21 12 16 7"></polyline>
										<line x1="21" y1="12" x2="9" y2="12"></line>
									</svg>
								</div>
								{#if !isSidebarCollapsed}
									<span
										class="text-sm font-medium transition-all duration-300 group-hover:translate-x-1 animate-in fade-in slide-in-from-left-2"
										>Logout</span
									>
								{/if}
							</div>
						</button>
					</div>
				</div>
			</aside>

			<div class="flex-1 flex flex-col h-full overflow-hidden relative">
				<!-- Mobile Top Header -->
				<header class="md:hidden h-14 bg-slate-950/80 backdrop-blur-md border-b border-white/5 flex items-center justify-between px-4 z-30 shrink-0">
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full animate-pulse"></div>
						<h1 class="text-lg font-bold text-slate-50 bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent">
							GoExile
						</h1>
					</div>
					<div class="flex items-center gap-3">
						<button class="relative text-slate-400 hover:text-white transition-colors">
							<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
								<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
							</svg>
							<span class="absolute top-0 right-0 w-1.5 h-1.5 bg-red-500 rounded-full"></span>
						</button>
						<button onclick={logout} class="text-slate-400 hover:text-red-400 transition-colors">
							<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
								<polyline points="16 17 21 12 16 7"></polyline>
								<line x1="21" y1="12" x2="9" y2="12"></line>
							</svg>
						</button>
					</div>
				</header>

				<!-- Main Content -->
				<main class="flex-1 overflow-auto relative">
					<div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-6 md:py-8 min-h-full pb-24 md:pb-8">
						{@render children()}
					</div>
				</main>

				<!-- Mobile Bottom Nav -->
				<nav class="md:hidden h-16 bg-slate-950/90 backdrop-blur-xl border-t border-white/10 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 safe-area-pb">
					<a href="/dashboard" class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive('/dashboard') || isRouteActive('/') ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'} transition-colors">
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<rect x="3" y="3" width="7" height="7"></rect>
							<rect x="14" y="3" width="7" height="7"></rect>
							<rect x="14" y="14" width="7" height="7"></rect>
							<rect x="3" y="14" width="7" height="7"></rect>
						</svg>
						<span class="text-[10px] font-medium">Dash</span>
					</a>
					<a href="/performance" class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive('/performance') ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'} transition-colors">
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
						</svg>
						<span class="text-[10px] font-medium">Perf</span>
					</a>
					<a href="/config" class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive('/config') || isRouteActive('/config/') ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'} transition-colors">
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<circle cx="12" cy="12" r="3"></circle>
							<path d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"></path>
						</svg>
						<span class="text-[10px] font-medium">Config</span>
					</a>
					<a href="/server" class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive('/server') ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'} transition-colors">
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7 10 12 15 17 10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						<span class="text-[10px] font-medium">Files</span>
					</a>
					<a href="/database" class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive('/database') ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'} transition-colors">
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
							<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
							<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
						</svg>
						<span class="text-[10px] font-medium">DB</span>
					</a>
				</nav>
			</div>
		</div>

		<!-- Spectacular Background for authenticated pages -->
		{#if $isAuthenticated && !isRouteActive('/login')}
			<div class="fixed inset-0 -z-50 overflow-hidden">
				<!-- Animated Gradient Background -->
				<div
					class="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 animate-gradient-shift"
				></div>

				<!-- Moving Nebula Clouds -->
				<div class="absolute inset-0">
					<div
						class="absolute top-0 left-0 w-[60%] h-[40%] bg-slate-700/20 rounded-full blur-[150px] animate-blob"
						style="animation-delay: 0s;"
					></div>
					<div
						class="absolute top-1/4 right-0 w-[50%] h-[35%] bg-gray-600/20 rounded-full blur-[150px] animate-blob"
						style="animation-delay: 3s;"
					></div>
					<div
						class="absolute bottom-0 left-1/3 w-[45%] h-[30%] bg-zinc-700/20 rounded-full blur-[120px] animate-blob"
						style="animation-delay: 5s;"
					></div>
				</div>

				<!-- Grid Pattern Overlay -->
				<div
					class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPgogICAgPGcgZmlsbD0iIzY0NzQ4YiIgZmlsbC1vcGFjaXR5PSIwLjAzIj4KICAgICAgPHBhdGggZD0iTTM2IDM0di00aC0ydjRoLTR2Mmg0djRoMnYtNGg0di0yaC00em0wLTMwVjBoLTJ2NGgtNHYyaDR2NGgyVjZoNFY0aC00ek02IDM0di00SDR2NEgwdjJoNHY0aDJ2LTRoNHYtMkg2ek02IDRWMEg0djRIMHZoaDR2NGgyVjZoNFY0SDZ6Ii8+CiAgICA8L2c+CiAgPC9nPgo8L3N2Zz4=')] opacity-[0.03]"
				></div>

				<!-- Vignette Effect -->
				<div
					class="absolute inset-0 bg-radial-gradient from-transparent via-black/20 to-black/60"
				></div>
			</div>

			<!-- Floating Elements -->
			<div class="fixed inset-0 pointer-events-none z-5">
				<div
					class="absolute top-20 left-20 w-16 h-16 border-2 border-slate-400/20 rounded-lg animate-float"
					style="animation-delay: 0s;"
				></div>
				<div
					class="absolute top-40 right-32 w-12 h-12 border-2 border-gray-400/20 rounded-full animate-pulse"
					style="animation-delay: 1s;"
				></div>
				<div
					class="absolute bottom-32 left-1/3 w-20 h-20 border-2 border-zinc-400/20 transform rotate-45 animate-spin"
					style="animation-delay: 2s; animation-duration: 15s;"
				></div>
			</div>
		{/if}
	{:else}
		{@render children()}
	{/if}

	<ToastContainer />
{/if}

<style>
	/* Safe Area for Mobile Bottom Nav */
	.safe-area-pb {
		padding-bottom: env(safe-area-inset-bottom);
	}

	@keyframes blob {
		0% {
			transform: translate(0px, 0px) scale(1);
		}
		33% {
			transform: translate(30px, -20px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
		100% {
			transform: translate(0px, 0px) scale(1);
		}
	}

	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-10px);
		}
	}

	@keyframes glow-pulse {
		0%,
		100% {
			opacity: 0.3;
		}
		50% {
			opacity: 0.8;
		}
	}

	@keyframes slide-in-nav {
		from {
			opacity: 0;
			transform: translateX(-30px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes neon-flicker {
		0%,
		100% {
			opacity: 1;
		}
		41.99% {
			opacity: 1;
		}
		42% {
			opacity: 0.8;
		}
		43% {
			opacity: 1;
		}
		47.99% {
			opacity: 1;
		}
		48% {
			opacity: 0.9;
		}
		49% {
			opacity: 1;
		}
	}

	.animate-blob {
		animation: blob 10s infinite;
	}

	.animation-delay-4000 {
		animation-delay: 4s;
	}

	.animate-gradient {
		background-size: 200% 200%;
		animation: gradient-shift 3s ease-in-out infinite;
	}

	.animate-float {
		animation: float 3s ease-in-out infinite;
	}

	.animate-glow-pulse {
		animation: glow-pulse 2s ease-in-out infinite;
	}

	.animate-slide-nav {
		animation: slide-in-nav 0.5s ease-out forwards;
	}

	.animate-neon-flicker {
		animation: neon-flicker 2s infinite;
	}

	.bg-size-200 {
		background-size: 200% 200%;
	}

	/* Enhanced hover effects */
	.hover-lift:hover {
		transform: translateY(-2px) scale(1.02);
		transition: all 0.3s ease;
	}

	.nav-item-active {
		position: relative;
		overflow: hidden;
	}

	.nav-item-active::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.1), transparent);
		animation: slide-shine 3s infinite;
	}

	@keyframes slide-shine {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	/* Sidebar glow effects */
	.sidebar-glow {
		box-shadow:
			inset 0 0 30px rgba(59, 130, 246, 0.1),
			0 0 20px rgba(59, 130, 246, 0.2);
	}

	/* Background Animations */
	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
	}

	@keyframes blob {
		0% {
			transform: translate(0px, 0px) scale(1);
		}
		33% {
			transform: translate(30px, -50px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
		100% {
			transform: translate(0px, 0px) scale(1);
		}
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0px) rotate(0deg);
		}
		50% {
			transform: translateY(-20px) rotate(5deg);
		}
	}

	.animate-gradient-shift {
		background-size: 200% 200%;
		animation: gradient-shift 8s ease infinite;
	}

	.animate-blob {
		animation: blob 7s infinite;
	}

	.animate-float {
		animation: float 3s ease-in-out infinite;
	}

	.bg-radial-gradient {
		background: radial-gradient(
			circle at center,
			transparent 0%,
			rgba(0, 0, 0, 0.2) 50%,
			rgba(0, 0, 0, 0.6) 100%
		);
	}

	/* Reusable Page Container */
	.page-container {
		/* Let pages handle their own styling - this is just for global page wrapper if needed */
	}

	/* Navigation Links */
	.nav-link {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		border-radius: 0.75rem;
		transition: all 0.5s;
		position: relative;
		overflow: hidden;
		backdrop-filter: blur(8px);
		border: 1px solid transparent;
		color: rgb(148, 163, 184);
		opacity: 1;
		transform: translateX(0);
	}

	.nav-link:hover {
		background: linear-gradient(to right, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
		color: rgb(226, 232, 240);
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
		border-color: rgba(255, 255, 255, 0.1);
		transform: scale(1.01);
	}

	.nav-link.nav-active {
		background: linear-gradient(to right, rgba(37, 99, 235, 0.3), rgba(37, 99, 235, 0.2));
		color: rgb(147, 197, 253);
		box-shadow: 0 25px 50px -12px rgba(37, 99, 235, 0.25);
		border-color: rgba(37, 99, 235, 0.4);
		transform: scale(1.02);
	}

	.nav-icon {
		width: 1.25rem;
		height: 1.25rem;
		transition: all 0.5s;
	}

	.nav-link:hover .nav-icon {
		transform: scale(1.25);
	}

	.nav-link:nth-child(1):hover .nav-icon {
		transform: scale(1.25) rotate(12deg);
	}

	.nav-link:nth-child(2):hover .nav-icon {
		transform: scale(1.25) rotate(180deg);
	}

	.nav-link:nth-child(3):hover .nav-icon {
		transform: scale(1.25) rotate(90deg);
	}

	.nav-link:nth-child(4):hover .nav-icon {
		transform: scale(1.25) translateY(-0.125rem);
	}

	.nav-link span {
		font-weight: 500;
		letter-spacing: 0.025em;
		transition: all 0.3s;
	}

	.nav-link:hover span {
		transform: translateX(0.25rem);
	}

	.nav-indicator {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		width: 0.375rem;
		background: linear-gradient(to bottom, rgb(96, 165, 250), rgb(34, 211, 238));
		border-radius: 0 0.25rem 0.25rem 0;
		box-shadow: 0 0 10px rgba(96, 165, 250, 0.5);
		animation: pulse 2s infinite;
	}
</style>