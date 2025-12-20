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
		connectionStatus,
		notes,
		showQuickActions,
		notifications
	} from '$lib/stores';
	import type { Note } from '$lib/stores';
	import {
		FolderTree,
		Menu,
		StickyNote,
		LayoutDashboard,
		Activity,
		Settings,
		Server,
		Database,
		Code2,
		Users,
		HardDrive,
		RefreshCw,
		History,
		Trash2,
		Download,
		Upload,
		ShieldCheck,
		Plus,
		Zap,
		ZapOff,
		Eye
	} from 'lucide-svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import QuickActionsTooltip from '$lib/components/QuickActionsTooltip.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';

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

	// Particle system
	let particleCanvas: HTMLCanvasElement | null = null;
	let particleCtx: CanvasRenderingContext2D | null = null;
	let particles: Particle[] = [];
	let animationFrameId: number | null = null;

	class Particle {
		x: number;
		y: number;
		vx: number;
		vy: number;
		radius: number;
		opacity: number;
		color: string;
		pulseSpeed: number;
		pulsePhase: number;

		constructor(width: number, height: number) {
			this.x = Math.random() * width;
			this.y = Math.random() * height;
			this.vx = (Math.random() - 0.5) * 0.3;
			this.vy = (Math.random() - 0.5) * 0.3;
			this.radius = Math.random() * 2.5 + 0.5;
			this.opacity = Math.random() * 0.5 + 0.2;
			this.pulseSpeed = Math.random() * 0.02 + 0.01;
			this.pulsePhase = Math.random() * Math.PI * 2;

			const colors = [
				'rgba(59, 130, 246',
				'rgba(96, 165, 250',
				'rgba(34, 211, 238',
				'rgba(147, 197, 253'
			];
			this.color = colors[Math.floor(Math.random() * colors.length)];
		}

		update(width: number, height: number) {
			this.x += this.vx;
			this.y += this.vy;

			if (this.x < 0 || this.x > width) this.vx *= -1;
			if (this.y < 0 || this.y > height) this.vy *= -1;

			this.pulsePhase += this.pulseSpeed;
		}

		draw(ctx: CanvasRenderingContext2D) {
			const pulse = Math.sin(this.pulsePhase) * 0.3 + 0.7;
			const currentRadius = this.radius * pulse;
			const currentOpacity = this.opacity * pulse;

			ctx.beginPath();
			ctx.arc(this.x, this.y, currentRadius, 0, Math.PI * 2);
			ctx.fillStyle = `${this.color}, ${currentOpacity})`;
			ctx.fill();

			// Glow effect
			const gradient = ctx.createRadialGradient(
				this.x,
				this.y,
				0,
				this.x,
				this.y,
				currentRadius * 3
			);
			gradient.addColorStop(0, `${this.color}, ${currentOpacity * 0.3})`);
			gradient.addColorStop(1, `${this.color}, 0)`);
			ctx.fillStyle = gradient;
			ctx.beginPath();
			ctx.arc(this.x, this.y, currentRadius * 3, 0, Math.PI * 2);
			ctx.fill();
		}
	}

	function initParticles() {
		if (!particleCanvas) return;
		const width = particleCanvas.width;
		const height = particleCanvas.height;
		particles = [];

		// Create fewer particles for better performance
		const particleCount = Math.min(80, Math.floor((width * height) / 15000));
		for (let i = 0; i < particleCount; i++) {
			particles.push(new Particle(width, height));
		}
	}

	function connectParticles() {
		if (!particleCtx) return;
		const maxDistance = 120;

		for (let i = 0; i < particles.length; i++) {
			for (let j = i + 1; j < particles.length; j++) {
				const dx = particles[i].x - particles[j].x;
				const dy = particles[i].y - particles[j].y;
				const distance = Math.sqrt(dx * dx + dy * dy);

				if (distance < maxDistance) {
					const opacity = (1 - distance / maxDistance) * 0.15;
					particleCtx.beginPath();
					particleCtx.strokeStyle = `rgba(96, 165, 250, ${opacity})`;
					particleCtx.lineWidth = 0.5;
					particleCtx.moveTo(particles[i].x, particles[i].y);
					particleCtx.lineTo(particles[j].x, particles[j].y);
					particleCtx.stroke();
				}
			}
		}
	}

	function animateParticles() {
		if (!particleCanvas || !particleCtx) return;

		const width = particleCanvas.width;
		const height = particleCanvas.height;

		particleCtx.clearRect(0, 0, width, height);

		particles.forEach((particle) => {
			particle.update(width, height);
			particle.draw(particleCtx!);
		});

		connectParticles();

		animationFrameId = requestAnimationFrame(animateParticles);
	}

	function handleResize() {
		if (typeof window === 'undefined' || !particleCanvas) return;
		particleCanvas.width = window.innerWidth;
		particleCanvas.height = window.innerHeight;
		initParticles();
	}

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
				if (data.type === 'stats') {
					stats.set(data.payload);
				} else if (data.type === 'spawners') {
					const list: any[] = Array.isArray(data.payload)
						? data.payload
						: Object.values(data.payload);
					list.sort((a, b) => a.id - b.id);
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
				connectSSE();
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

		// Initialize particle system (browser only)
		if (typeof window !== 'undefined' && particleCanvas) {
			particleCtx = particleCanvas.getContext('2d', { alpha: true });
			handleResize();
			animateParticles();
			window.addEventListener('resize', handleResize);
		}

		setTimeout(() => {
			sidebarLoaded = true;
		}, 300);
	});

	onDestroy(() => {
		if (eventSource) eventSource.close();
		if (animationFrameId) cancelAnimationFrame(animationFrameId);
		if (typeof window !== 'undefined') {
			window.removeEventListener('resize', handleResize);
		}
	});

	async function logout() {
		await fetch('/api/auth/logout', { method: 'POST' });
		isAuthenticated.set(false);
		goto('/login');
	}

	function isRouteActive(path: string) {
		return page.url.pathname === path;
	}

	function toggleSidebar() {
		isSidebarCollapsed = !isSidebarCollapsed;
	}

	let showGlobalNoteModal = $state(false);

	async function handleGlobalSaveNote(note: Note) {
		// Quick Note is always a new note
		const { id, ...noteWithoutId } = note;
		try {
			const res = await fetch('/api/notes', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(noteWithoutId)
			});
			if (res.ok) {
				const saved = await res.json();
				notes.update((n) => [saved, ...n]);
			}
		} catch (e) {
			console.error(e);
		}
	}

	function toggleQuickActions() {
		showQuickActions.update((v) => !v);
	}

	async function handleForceGC() {
		try {
			await fetch('/api/metrics/gc', { method: 'POST' });
			notifications.add({ type: 'success', message: 'Garbage Collection triggered' });
		} catch (e) {
			console.error(e);
		}
	}

	async function handleExportConfig() {
		try {
			const res = await fetch('/api/config');
			const data = await res.json();
			const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `server_config_${new Date().toISOString().split('T')[0]}.json`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			window.URL.revokeObjectURL(url);
			notifications.add({ type: 'success', message: 'Configuration exported' });
		} catch (e) {
			console.error(e);
			notifications.add({ type: 'error', message: 'Failed to export configuration' });
		}
	}

	function handleDownloadFiles() {
		// Trigger download of game_server.zip
		const link = document.createElement('a');
		link.href = '/api/spawners/download';
		link.download = 'game_server.zip';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		notifications.add({ type: 'info', message: 'Download started' });
	}

	async function handleBackupDB() {
		try {
			const res = await fetch('/api/database/backup', { method: 'POST' });
			if (res.ok) {
				notifications.add({ type: 'success', message: 'Database backup started' });
			} else {
				notifications.add({ type: 'error', message: 'Failed to start backup' });
			}
		} catch (e) {
			console.error(e);
			notifications.add({ type: 'error', message: 'Backup request failed' });
		}
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
			<!-- Global Restart Banner -->
			{#if $restartRequired}
				<div
					class="absolute top-0 md:left-64 left-0 right-0 z-50 bg-gradient-to-r from-orange-600/95 via-amber-600/95 to-orange-600/95 backdrop-blur-md text-white px-4 py-2.5 flex justify-between items-center shadow-2xl border-b-2 border-orange-400/50 animate-slide-fade text-xs md:text-sm"
				>
					<div class="flex items-center gap-3">
						<div class="relative">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-5 h-5 shrink-0 animate-bounce"
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
							<div
								class="absolute inset-0 bg-yellow-300/30 rounded-full blur-md animate-pulse"
							></div>
						</div>
						<span class="font-semibold truncate">⚡ Server restart required to apply changes</span>
					</div>
					<button
						onclick={restartServer}
						disabled={restarting}
						class="relative px-4 py-1.5 bg-white text-orange-600 rounded-lg font-bold hover:bg-orange-50 active:scale-95 transition-all shadow-lg hover:shadow-xl text-xs disabled:opacity-50 whitespace-nowrap overflow-hidden group border border-orange-200"
					>
						<div
							class="absolute inset-0 bg-gradient-to-r from-orange-100/0 via-orange-100/50 to-orange-100/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-500"
						></div>
						<span class="relative z-10 flex items-center gap-1.5">
							{#if restarting}
								<svg
									class="animate-spin w-3 h-3"
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
								>
									<circle
										class="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										stroke-width="4"
									></circle>
									<path
										class="opacity-75"
										fill="currentColor"
										d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
									></path>
								</svg>
								Restarting...
							{:else}
								<svg
									xmlns="http://www.w3.org/2000/svg"
									class="w-3 h-3"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
								>
									<polyline points="23 4 23 10 17 10"></polyline>
									<path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
								</svg>
								Restart Now
							{/if}
						</span>
					</button>
				</div>
			{/if}

			<!-- Desktop Sidebar -->
			<aside
				class="hidden md:flex relative transition-all duration-300 ease-in-out bg-slate-950 border-r border-slate-800 flex-col shrink-0 overflow-hidden shadow-2xl z-20 {isSidebarCollapsed
					? 'w-20'
					: 'w-64'}"
			>
				<div
					class="absolute inset-0 bg-slate-900/60 backdrop-blur-[1px] border-r border-white/5"
				></div>

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
							class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/10 transition-all duration-300 active:scale-90 border border-transparent hover:border-white/10 group {isSidebarCollapsed
								? 'mx-auto'
								: ''}"
							title={isSidebarCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-5 h-5 transition-transform duration-300 group-hover:scale-110"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
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
						<QuickActionsTooltip
							placement="right"
							title="Dashboard"
							class="w-full"
							enabled={$showQuickActions}
							actions={[{ label: 'Refresh Data', icon: RefreshCw, onClick: initialFetch }]}
						>
							<a
								href="/dashboard"
								class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
								class:nav-active={isRouteActive('/dashboard') || isRouteActive('/')}
								style="animation-delay: 0.1s;"
								title={isSidebarCollapsed ? 'Dashboard' : ''}
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
						</QuickActionsTooltip>

						<QuickActionsTooltip
							placement="right"
							title="Performance"
							class="w-full"
							enabled={$showQuickActions}
							actions={[{ label: 'Force GC', icon: Trash2, onClick: handleForceGC, variant: 'warning' }]}
						>
							<a
								href="/performance"
								class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
								class:nav-active={isRouteActive('/performance')}
								style="animation-delay: 0.2s;"
								title={isSidebarCollapsed ? 'Performance' : ''}
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
									<span class="animate-in fade-in slide-in-from-left-2 duration-300">Performance</span
									>
								{/if}
							</a>
						</QuickActionsTooltip>

						<QuickActionsTooltip
							placement="right"
							title="Configuration"
							class="w-full"
							enabled={$showQuickActions}
							actions={[
								{ label: 'Export Config', icon: Download, onClick: handleExportConfig },
								{ label: 'Reload Server', icon: RefreshCw, onClick: restartServer, variant: 'danger' }
							]}
						>
							<a
								href="/config"
								class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
								class:nav-active={isRouteActive('/config') || isRouteActive('/config/')}
								style="animation-delay: 0.3s;"
								title={isSidebarCollapsed ? 'Configuration' : ''}
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
									<span class="animate-in fade-in slide-in-from-left-2 duration-300"
										>Configuration</span
									>
								{/if}
							</a>
						</QuickActionsTooltip>

						<QuickActionsTooltip
							placement="right"
							title="Server Files"
							class="w-full"
							enabled={$showQuickActions}
							actions={[{ label: 'Download Core', icon: Download, onClick: handleDownloadFiles }]}
						>
							<a
								href="/server"
								class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
								class:nav-active={isRouteActive('/server')}
								style="animation-delay: 0.4s;"
								title={isSidebarCollapsed ? 'Server Files' : ''}
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
																	<span class="animate-in fade-in slide-in-from-left-2 duration-300"
																		>Server Files</span
																	>
																{/if}
															</a>
														</QuickActionsTooltip>
								
														<QuickActionsTooltip
															placement="right"
															title="Databases"
															class="w-full"
															enabled={$showQuickActions}
															actions={[{ label: 'Backup All', icon: HardDrive, onClick: handleBackupDB }]}
														>
															<a
																href="/database"
																class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
																class:nav-active={isRouteActive('/database')}
																style="animation-delay: 0.5s;"
																title={isSidebarCollapsed ? 'Databases' : ''}
															>
																{#if isRouteActive('/database')}
																	<div class="nav-indicator"></div>
																{/if}
																<Database class="nav-icon" />
																{#if !isSidebarCollapsed}
																	<span class="animate-in fade-in slide-in-from-left-2 duration-300">Databases</span>
																{/if}
															</a>
														</QuickActionsTooltip>
								
														<QuickActionsTooltip
															placement="right"
															title="RedEye"
															class="w-full"
															enabled={$showQuickActions}
															actions={[]}
														>													<a
														href="/redeye"
														class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
														class:nav-active={isRouteActive('/redeye')}
														style="animation-delay: 0.6s;"
														title={isSidebarCollapsed ? 'RedEye' : ''}
													>
														{#if isRouteActive('/redeye')}
															<div class="nav-indicator"></div>
														{/if}
														<Eye class="nav-icon" />
														{#if !isSidebarCollapsed}
															<span class="animate-in fade-in slide-in-from-left-2 duration-300">RedEye</span>
														{/if}
													</a>
												</QuickActionsTooltip>
						
												<QuickActionsTooltip
													placement="right"
													title="Notes"
													class="w-full"
													enabled={$showQuickActions}
													actions={[{ label: 'New Note', icon: Plus, onClick: () => (showGlobalNoteModal = true) }]}
												>
													<a
														href="/notes"
														class="nav-link w-full {isSidebarCollapsed ? 'justify-center px-2' : ''}"
														class:nav-active={isRouteActive('/notes')}
														style="animation-delay: 0.7s;"
														title={isSidebarCollapsed ? 'Notes' : ''}
													>
														{#if isRouteActive('/notes')}
															<div class="nav-indicator"></div>
														{/if}
														<StickyNote class="nav-icon" />
														{#if !isSidebarCollapsed}
															<span class="animate-in fade-in slide-in-from-left-2 duration-300">Notes</span>
														{/if}
													</a>
												</QuickActionsTooltip>					</nav>

					<div
						class="p-4 border-t border-white/5 bg-black/20 backdrop-blur-md flex flex-col gap-2 transform transition-all duration-700 {sidebarLoaded
							? 'translate-y-0 opacity-100'
							: 'translate-y-8 opacity-0'}"
						style="animation-delay: 0.6s;"
					>
						{#if !isSidebarCollapsed}
							<button
								onmouseenter={() => (hoveredItem = 6)}
								onmouseleave={() => (hoveredItem = -1)}
								class="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl text-slate-400 hover:text-white transition-all duration-300 group relative overflow-hidden animate-in fade-in zoom-in
								{hoveredItem === 6
									? 'bg-gradient-to-r from-blue-500/20 to-indigo-500/20 border-blue-400/40 shadow-lg shadow-blue-500/20 scale-[1.02]'
									: 'hover:bg-slate-800/50'}
								border border-transparent hover:border-blue-400/20"
							>
								<div
									class="absolute inset-0 bg-gradient-to-r from-blue-600/0 via-blue-400/10 to-blue-600/0 translate-x-[-100%] {hoveredItem ===
									6
										? 'translate-x-[100%]'
										: ''} transition-transform duration-700"
								></div>

								<div class="relative z-10">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="w-5 h-5 transition-all duration-300 {hoveredItem === 6
											? 'rotate-12 scale-110'
											: ''}"
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
									<span
										class="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full animate-ping"
									></span>
									<span class="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full"></span>
								</div>

								<span
									class="relative z-10 text-sm font-medium transition-all duration-300 {hoveredItem ===
									6
										? 'translate-x-1'
										: ''}">Notifications</span
								>
							</button>

							<button
								onclick={toggleQuickActions}
								onmouseenter={() => (hoveredItem = 7)}
								onmouseleave={() => (hoveredItem = -1)}
								class="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl text-slate-400 hover:text-white transition-all duration-300 group relative overflow-hidden animate-in fade-in zoom-in
								{hoveredItem === 7
									? 'bg-gradient-to-r from-yellow-500/20 to-amber-500/20 border-yellow-400/40 shadow-lg shadow-yellow-500/20 scale-[1.02]'
									: 'hover:bg-slate-800/50'}
								border border-transparent hover:border-yellow-400/20"
								title={$showQuickActions ? 'Disable Quick Actions' : 'Enable Quick Actions'}
							>
								<div
									class="absolute inset-0 bg-gradient-to-r from-yellow-600/0 via-yellow-400/10 to-yellow-600/0 translate-x-[-100%] {hoveredItem ===
									7
										? 'translate-x-[100%]'
										: ''} transition-transform duration-700"
								></div>

								<div class="relative z-10">
									{#if $showQuickActions}
										<Zap
											class="w-5 h-5 transition-all duration-300 text-yellow-400 {hoveredItem === 7
												? 'scale-110 fill-yellow-400/20'
												: ''}"
										/>
									{:else}
										<ZapOff
											class="w-5 h-5 transition-all duration-300 {hoveredItem === 7
												? 'scale-110'
												: ''}"
										/>
									{/if}
								</div>

								<span
									class="relative z-10 text-sm font-medium transition-all duration-300 {hoveredItem ===
									7
										? 'translate-x-1'
										: ''}">Quick Actions</span
								>
							</button>
						{/if}

						<button
							onclick={logout}
							onmouseenter={() => (hoveredItem = 5)}
							onmouseleave={() => (hoveredItem = -1)}
							class="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl text-slate-400 hover:text-white border transition-all duration-300 group relative overflow-hidden
							{hoveredItem === 5
								? 'bg-gradient-to-r from-red-500/20 to-rose-600/20 border-red-400/50 shadow-lg shadow-red-500/30 scale-[1.02]'
								: 'border-transparent hover:bg-red-500/10 hover:border-red-500/20'}"
							title={isSidebarCollapsed ? 'Logout' : ''}
						>
							<!-- Animated shimmer effect -->
							<div
								class="absolute inset-0 bg-gradient-to-r from-red-600/0 via-red-400/20 to-red-600/0 translate-x-[-100%] {hoveredItem ===
								5
									? 'translate-x-[100%]'
									: ''} transition-transform duration-700"
							></div>

							<!-- Pulsing danger background -->
							<div
								class="absolute inset-0 {hoveredItem === 5
									? 'animate-pulse'
									: ''} bg-red-500/5 rounded-xl"
							></div>

							<!-- Warning particles effect -->
							{#if hoveredItem === 5}
								<div
									class="absolute top-0 left-1/4 w-1 h-1 bg-red-400 rounded-full animate-ping"
									style="animation-delay: 0s;"
								></div>
								<div
									class="absolute bottom-1 right-1/3 w-1 h-1 bg-red-400 rounded-full animate-ping"
									style="animation-delay: 0.2s;"
								></div>
								<div
									class="absolute top-1 right-1/4 w-1 h-1 bg-rose-400 rounded-full animate-ping"
									style="animation-delay: 0.4s;"
								></div>
							{/if}

							<div class="relative z-10 flex items-center justify-center gap-2">
								<div
									class="w-4 h-4 flex items-center justify-center transition-all duration-300 {hoveredItem ===
									5
										? 'rotate-12 scale-110'
										: ''}"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="w-4 h-4 transition-all duration-300 group-hover:-translate-x-0.5"
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
										class="text-sm font-semibold transition-all duration-300 {hoveredItem === 5
											? 'translate-x-1'
											: ''}">Logout</span
									>
								{/if}
							</div>
						</button>
					</div>
				</div>
			</aside>

			<div class="flex-1 flex flex-col h-full overflow-hidden relative">
				<!-- Mobile Top Header -->
				<header
					class="md:hidden h-14 bg-slate-950/80 backdrop-blur-md border-b border-white/5 flex items-center justify-between px-4 z-30 shrink-0"
				>
					<div class="flex items-center gap-2">
						<div
							class="w-2.5 h-2.5 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full animate-pulse shadow-lg shadow-blue-500/50"
						></div>
						<h1
							class="text-lg font-bold text-slate-50 bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent"
						>
							GoExile
						</h1>
					</div>
					<div class="flex items-center gap-2">
						<button
							class="relative p-2 text-slate-400 hover:text-white transition-all duration-300 rounded-lg hover:bg-white/5 active:scale-90 group"
							aria-label="Notifications"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-5 h-5 transition-transform duration-300 group-hover:rotate-12"
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
							<span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full animate-ping"
							></span>
							<span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full"></span>
						</button>
						<button
							onclick={logout}
							class="p-2 text-slate-400 hover:text-red-400 transition-all duration-300 rounded-lg hover:bg-red-500/10 active:scale-90 group"
							aria-label="Logout"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="w-5 h-5 transition-transform duration-300 group-hover:-translate-x-0.5"
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
				<nav
					class="md:hidden h-16 bg-slate-950/90 backdrop-blur-xl border-t border-white/10 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 safe-area-pb"
				>
					<a
						href="/dashboard"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/dashboard'
						) || isRouteActive('/')
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<svg
							class="w-5 h-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<rect x="3" y="3" width="7" height="7"></rect>
							<rect x="14" y="3" width="7" height="7"></rect>
							<rect x="14" y="14" width="7" height="7"></rect>
							<rect x="3" y="14" width="7" height="7"></rect>
						</svg>
						<span class="text-[10px] font-medium">Dash</span>
					</a>
					<a
						href="/performance"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/performance'
						)
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<svg
							class="w-5 h-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
						</svg>
						<span class="text-[10px] font-medium">Perf</span>
					</a>
					<a
						href="/config"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/config'
						) || isRouteActive('/config/')
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<svg
							class="w-5 h-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<circle cx="12" cy="12" r="3"></circle>
							<path
								d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"
							></path>
						</svg>
						<span class="text-[10px] font-medium">Config</span>
					</a>
					<a
						href="/server"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/server'
						)
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<svg
							class="w-5 h-5"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7 10 12 15 17 10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						<span class="text-[10px] font-medium">Files</span>
					</a>
					<a
						href="/database"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/database'
						)
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<svg
							class="w-5 h-5"
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
						<span class="text-[10px] font-medium">DB</span>
					</a>
					<a
						href="/redeye"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/redeye'
						)
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<Eye class="w-5 h-5" />
						<span class="text-[10px] font-medium">RedEye</span>
					</a>
					<a
						href="/notes"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/notes'
						)
							? 'text-blue-400'
							: 'text-slate-500 hover:text-slate-300'} transition-colors"
					>
						<StickyNote class="w-5 h-5" />
						<span class="text-[10px] font-medium">Notes</span>
					</a>
				</nav>
			</div>
		</div>

		<!-- Spectacular Animated Particle Background -->
		{#if $isAuthenticated && !isRouteActive('/login')}
			<canvas
				bind:this={particleCanvas}
				class="fixed inset-0 -z-50 pointer-events-none"
				style="background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);"
			></canvas>

			<!-- Additional atmospheric layers -->
			<div class="fixed inset-0 -z-40 pointer-events-none">
				<!-- Radial gradient overlay -->
				<div
					class="absolute inset-0 bg-gradient-radial from-transparent via-slate-900/30 to-slate-950/60"
				></div>

				<!-- Animated gradient orbs -->
				<div
					class="absolute top-0 left-0 w-[500px] h-[500px] bg-blue-500/10 rounded-full blur-[120px] animate-float-slow"
				></div>
				<div
					class="absolute bottom-0 right-0 w-[600px] h-[600px] bg-cyan-500/10 rounded-full blur-[140px] animate-float-slower"
				></div>
				<div
					class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] h-[400px] bg-indigo-500/10 rounded-full blur-[100px] animate-pulse-slow"
				></div>
			</div>
		{/if}
	{:else}
		{@render children()}
	{/if}

	<NoteModal 
		bind:isOpen={showGlobalNoteModal} 
		note={null} 
		onSave={handleGlobalSaveNote} 
		onClose={() => (showGlobalNoteModal = false)} 
	/>

	<ToastContainer />
{/if}

<style>
	.safe-area-pb {
		padding-bottom: env(safe-area-inset-bottom);
	}

	@keyframes float-slow {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(50px, -30px) scale(1.05);
		}
		66% {
			transform: translate(-30px, 30px) scale(0.95);
		}
	}

	@keyframes float-slower {
		0%,
		100% {
			transform: translate(0, 0) scale(1);
		}
		50% {
			transform: translate(-40px, -40px) scale(1.08);
		}
	}

	@keyframes pulse-slow {
		0%,
		100% {
			opacity: 0.5;
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			opacity: 0.7;
			transform: translate(-50%, -50%) scale(1.1);
		}
	}

	.animate-float-slow {
		animation: float-slow 20s ease-in-out infinite;
	}

	.animate-float-slower {
		animation: float-slower 25s ease-in-out infinite;
	}

	.animate-pulse-slow {
		animation: pulse-slow 15s ease-in-out infinite;
	}

	.bg-gradient-radial {
		background: radial-gradient(
			circle at center,
			transparent 0%,
			rgba(15, 23, 42, 0.3) 50%,
			rgba(15, 23, 42, 0.6) 100%
		);
	}

	.animate-gradient {
		background-size: 200% 200%;
		animation: gradient-shift 3s ease-in-out infinite;
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
