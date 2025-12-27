<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { fade, slide } from 'svelte/transition';
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
		notifications,
		theme,
		backgroundConfig,
		siteSettings,
		loadAllSettings
	} from '$lib/stores';
	import type { Note } from '$lib/stores';
	import {
		FolderTree,
		Menu,
		StickyNote,
		LayoutDashboard,
		Activity,
		Settings as SettingsIcon,
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
		Eye,
		Sun,
		Moon,
		Palette
	} from 'lucide-svelte';
	import QuickActionsTooltip from '$lib/components/QuickActionsTooltip.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';

	import NavbarParticles from '$lib/components/theme/NavbarParticles.svelte';
	import GlobalSmoke from '$lib/components/theme/GlobalSmoke.svelte';
	import SectionBackground from '$lib/components/theme/SectionBackground.svelte';
	import ServerStatus from '$lib/components/theme/ServerStatus.svelte';
	import Notifications from '$lib/components/theme/Notifications.svelte';

	let { children } = $props();
	let isChecking = $state(true);
	let restarting = $state(false);
	let eventSource: EventSource | null = null;

	let localBackgroundConfig = $derived($backgroundConfig);
	let localSiteSettings = $derived($siteSettings);

	// Theme handling
	$effect(() => {
		if (typeof window !== 'undefined') {
			if ($theme === 'dark') {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
			localStorage.setItem('theme', $theme);

			// Sync aesthetic variables
			const root = document.documentElement;
			if ($siteSettings.aesthetic) {
				root.style.setProperty('--card-alpha', ($siteSettings.aesthetic.glass_strength ?? $siteSettings.aesthetic.card_alpha ?? 0.6).toString());
				root.style.setProperty('--backdrop-blur', ($siteSettings.aesthetic.backdrop_blur ?? 16) + 'px');
				root.style.setProperty('--sidebar-alpha', ($siteSettings.aesthetic.sidebar_alpha ?? 0.7).toString());
				root.style.setProperty('--bg-opacity', ($siteSettings.aesthetic.bg_opacity ?? 1.0).toString());
				root.style.setProperty('--bg-color', $siteSettings.aesthetic.bg_color || '#050505');
				
				// Fonts
				root.style.setProperty('--font-primary', ($siteSettings.aesthetic.font_body || 'Inter') + ', sans-serif');
				root.style.setProperty('--font-header', ($siteSettings.aesthetic.font_header || 'Kanit') + ', sans-serif');
				root.style.setProperty('--font-body', ($siteSettings.aesthetic.font_body || 'Inter') + ', sans-serif');
				root.style.setProperty('--font-mono', ($siteSettings.aesthetic.font_mono || 'JetBrains Mono') + ', monospace');

				// Geometry
				root.style.setProperty('--radius-sm', ($siteSettings.aesthetic.border_radius_sm ?? 2) + 'px');
				root.style.setProperty('--radius-md', ($siteSettings.aesthetic.border_radius_md ?? 4) + 'px');
				root.style.setProperty('--radius-lg', ($siteSettings.aesthetic.border_radius_lg ?? 8) + 'px');

				root.style.setProperty('--card-border-width', ($siteSettings.aesthetic.card_border_width ?? 1) + 'px');
				root.style.setProperty('--card-shadow-size', ($siteSettings.aesthetic.card_shadow_size ?? 4) + 'px');
				
				// Granular Colors
				root.style.setProperty('--primary-color', $siteSettings.aesthetic.primary_color || '#d97706');
				root.style.setProperty('--secondary-color', $siteSettings.aesthetic.secondary_color || '#1e293b');
				root.style.setProperty('--card-bg-color', $siteSettings.aesthetic.card_bg_color || '#0f172a');
				root.style.setProperty('--hover-color', $siteSettings.aesthetic.hover_color || '#1e293b');
				root.style.setProperty('--text-primary', $siteSettings.aesthetic.text_color_primary || '#e2e2e2');
				root.style.setProperty('--text-secondary', $siteSettings.aesthetic.text_color_secondary || '#888888');
				root.style.setProperty('--border-color', $siteSettings.aesthetic.border_color || '#1e293b');

				// New settings
				root.style.setProperty('--scanline-speed', ($siteSettings.aesthetic.scanline_speed ?? 4) + 's');
				root.style.setProperty('--scanline-density', ($siteSettings.aesthetic.scanline_density ?? 3) + 'px');
				root.style.setProperty('--vignette-intensity', ($siteSettings.aesthetic.vignette_intensity ?? 0.5).toString());
				root.style.setProperty('--glow-intensity', ($siteSettings.aesthetic.border_glow_intensity ?? 0.4).toString());
				root.style.setProperty('--sidebar-width', ($siteSettings.aesthetic.sidebar_width ?? 280) + 'px');
				root.style.setProperty('--font-size-base', ($siteSettings.aesthetic.font_size_base ?? 14) + 'px');
				root.style.setProperty('--glow-color', $siteSettings.aesthetic.card_glow_color || '#d97706');
				
				if ($siteSettings.aesthetic.reduced_motion) {
					root.classList.add('reduced-motion');
				} else {
					root.classList.remove('reduced-motion');
				}

				if ($siteSettings.aesthetic.text_glow) {
					root.classList.add('text-glow-enabled');
				} else {
					root.classList.remove('text-glow-enabled');
				}

				let accent = $siteSettings.aesthetic.primary_color || '#d97706';
				
				// RED ALERT / PANIC MODE
				if ($siteSettings.aesthetic.panic_mode) {
					accent = '#dc2626'; // Pure red
					root.style.setProperty('--border-color', '#991b1b');
					root.style.setProperty('--card-bg-hex', '#450a0a');
				} else {
					root.style.setProperty('--card-bg-hex', $siteSettings.aesthetic.card_bg_color || '#0f172a');
				}

				root.style.setProperty('--accent-color', accent);
				root.style.setProperty('--accent-color-light', accent + 'dd');
			}
		}
	});

	function toggleTheme() {
		theme.update((t) => (t === 'dark' ? 'light' : 'dark'));
	}

	// Animation states
	let sidebarLoaded = $state(false);
	let mouseX = $state(0);
	let mouseY = $state(0);
	let hoveredItem = $state(-1);
	let isSidebarCollapsed = $state(true);
	let isMobileMenuOpen = $state(false);

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
		// Don't check auth if we are already on the login page
		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') {
			isChecking = false;
			return;
		}

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
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme === 'light' || savedTheme === 'dark') {
			theme.init(savedTheme);
		}

		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') {
			isChecking = false;
		} else {
			await checkAuth();
			
			if ($isAuthenticated) {
				await loadAllSettings();
				initialFetch();
			}
		}

		setTimeout(() => {
			sidebarLoaded = true;
		}, 300);
	});

	onDestroy(() => {
		if (eventSource) eventSource.close();
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
		class="flex items-center justify-center min-h-screen bg-stone-950"
	>
		<div class="relative">
			<div
				class="animate-spin rounded-full h-16 w-16 border-4 border-stone-800 border-t-rust shadow-2xl"
			></div>
			<div
				class="absolute inset-0 rounded-full bg-rust/20 blur-xl animate-pulse"
			></div>
		</div>
	</div>
{:else}
	{#if $isAuthenticated && page.url.pathname !== '/login'}
		<div class="relative min-h-screen {localSiteSettings?.aesthetic?.crt_effect ? 'crt-container' : ''} {localSiteSettings?.aesthetic?.crt_curve ? 'crt-curve' : ''} {localSiteSettings?.aesthetic?.panic_mode ? 'panic-mode' : ''}">
			<!-- System Ticker Header (New) -->
			<div class="fixed top-0 left-0 right-0 h-6 bg-black/80 backdrop-blur-md border-b border-stone-800 z-[110] flex items-center px-4 overflow-hidden">
				<div class="flex items-center gap-6 animate-text-reveal whitespace-nowrap">
					<span class="tactical-code text-rust-light">[SYSTEM_READY]</span>
					<span class="tactical-code text-stone-600">CONNECTION: {$connectionStatus}</span>
					<span class="tactical-code text-stone-600">ACTIVE_NODES: {$stats.active_spawners}</span>
					<span class="tactical-code text-stone-600">TIMESTAMP: {new Date().toISOString()}</span>
					<span class="tactical-code text-rust-light">[BUFFER_OPTIMIZED]</span>
				</div>
			</div>

			<!-- Overlays -->
			<div class="fixed inset-0 z-[100] pointer-events-none overflow-hidden">
				<div class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[length:100%_4px,3px_100%]" style="opacity: {localSiteSettings?.aesthetic?.scanlines_opacity || 0.05}"></div>
				<div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/asfalt-dark.png')] opacity-20 mix-blend-overlay" style="opacity: {localSiteSettings?.aesthetic?.noise_opacity || 0.03}"></div>
			</div>

			{#if localBackgroundConfig?.show_clouds}
				<div class="clouds-overlay" style="opacity: {localBackgroundConfig.clouds_opacity}"></div>
				<div class="clouds-overlay opacity-40" style="animation-direction: reverse; animation-duration: 180s; opacity: {localBackgroundConfig.clouds_opacity * 0.5}"></div>
			{/if}

			{#if localBackgroundConfig?.show_rain}
				<div class="rain-container">
					<div class="rain-layer rain-layer-back"></div>
					<div class="rain-layer rain-layer-mid"></div>
					<div class="rain-layer rain-layer-front" style="opacity: {localBackgroundConfig.rain_opacity * 0.5}"></div>
				</div>
			{/if}

			{#if localBackgroundConfig?.show_smoke}
				<GlobalSmoke />
			{/if}

			{#if localBackgroundConfig?.show_vignette}
				<div class="vignette"></div>
			{/if}

			{#if localBackgroundConfig?.global_type && localBackgroundConfig.global_type !== 'none'}
				<div class="fixed inset-0 z-0 pointer-events-none overflow-hidden opacity-50">
					<SectionBackground type={localBackgroundConfig.global_type} />
				</div>
			{/if}

			{#if localSiteSettings?.site_notice?.enabled}
				<div class="relative z-[60] py-2 px-4 text-center text-[10px] font-heading tracking-[0.3em] uppercase transition-colors duration-500 {localSiteSettings.site_notice.type === 'critical' ? 'bg-red-600 text-white animate-pulse' : localSiteSettings.site_notice.type === 'warning' ? 'bg-rust text-white' : 'bg-stone-800 text-stone-300'}">
					<span class="mr-2">[{localSiteSettings.site_notice.type.toUpperCase()}_BROADCAST]</span>
					{localSiteSettings.site_notice.message}
				</div>
			{/if}

			<div
				class="flex h-screen text-stone-400 overflow-hidden relative bg-transparent transition-colors duration-300"
			>
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
				class="hidden md:flex relative transition-all duration-500 ease-[cubic-bezier(0.23,1,0.32,1)] bg-[var(--sidebar-bg)] backdrop-blur-xl border-r-2 border-stone-800 flex-col shrink-0 overflow-hidden z-20 {isSidebarCollapsed
					? 'w-20'
					: 'w-[var(--sidebar-width)]'} {$siteSettings.aesthetic.panic_mode ? 'border-red-600 shadow-[0_0_30px_rgba(220,38,38,0.2)]' : 'shadow-2xl'}"
			>
				{#if localBackgroundConfig?.show_navbar_particles}
					<NavbarParticles />
				{/if}
				
				<!-- Tactical Overlays -->
				<div class="absolute inset-0 pointer-events-none opacity-20">
					<div class="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_2px_2px,rgba(255,255,255,0.05)_1px,transparent_0)] bg-[size:24px_24px]"></div>
					{#if $siteSettings.aesthetic.crt_effect}
						<div class="absolute top-0 left-0 w-full h-full animate-scanline_sweep bg-gradient-to-b from-transparent via-white/5 to-transparent"></div>
					{/if}
				</div>

				<div class="relative z-10 flex flex-col h-full">
					<div
						class="p-6 border-b-2 border-stone-800 bg-black/40 transform transition-all duration-700 {sidebarLoaded
							? 'translate-y-0 opacity-100'
							: '-translate-y-4 opacity-0'} flex items-center justify-between tactical-border"
					>
						{#if !isSidebarCollapsed}
							<div class="corner-tl"></div>
							<div class="corner-tr"></div>
							<div class="flex flex-col animate-in fade-in zoom-in duration-500">
								<div class="flex items-center gap-2">
									<div class="w-2 h-2 bg-rust shadow-[0_0_8px_var(--color-rust)]"></div>
									<h1 class="text-xl font-black military-label text-white tracking-tighter uppercase">
										EXILE_<span class="text-rust-light">CONTROLLER</span>
									</h1>
								</div>
								<span class="text-[8px] font-mono text-stone-600 mt-1 tracking-[0.3em]">VERSION 0.9.4</span>
							</div>
						{/if}
						<button
							onclick={toggleSidebar}
							class="p-2 rounded-none text-stone-500 hover:text-white bg-stone-900/50 border border-stone-800 hover:border-rust transition-all duration-300 {isSidebarCollapsed
								? 'mx-auto'
								: ''}"
						>
							<Menu class="w-4 h-4" />
						</button>
					</div>

					<nav class="flex-1 p-3 space-y-4 overflow-y-auto overflow-x-hidden no-scrollbar">
						<!-- Terminal Output Simulation -->
						{#if !isSidebarCollapsed}
							<div class="px-2 py-1 mb-4 border-l-2 border-rust/30 opacity-40">
								<div class="tactical-code text-stone-500 animate-pulse">>> SYSTEM READY</div>
								<div class="tactical-code text-stone-600">>> READY FOR INPUT</div>
							</div>
						{/if}

						<div class="space-y-1">
							<a
								href="/dashboard"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/dashboard') || isRouteActive('/')}
							>
								<div class="nav-icon-container">
									<LayoutDashboard class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">DASHBOARD</span>
										<span class="nav-subtext">System Overview</span>
									</div>
								{/if}
							</a>

							<a
								href="/performance"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/performance')}
							>
								<div class="nav-icon-container">
									<Activity class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">PERFORMANCE</span>
										<span class="nav-subtext">Real-time Metrics</span>
									</div>
								{/if}
							</a>

							<a
								href="/config"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/config')}
							>
								<div class="nav-icon-container">
									<SettingsIcon class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">CONFIGURATION</span>
										<span class="nav-subtext">App Settings</span>
									</div>
								{/if}
							</a>

							<a
								href="/config/theme"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/config/theme')}
							>
								<div class="nav-icon-container">
									<Palette class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">THEME LAB</span>
										<span class="nav-subtext">Visual Calibration</span>
									</div>
								{/if}
							</a>

							<a
								href="/server"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/server')}
							>
								<div class="nav-icon-container">
									<HardDrive class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">FILE_STORAGE</span>
										<span class="nav-subtext">Game Binaries</span>
									</div>
								{/if}
							</a>

							<a
								href="/database"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/database')}
							>
								<div class="nav-icon-container">
									<Database class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">DATABASE</span>
										<span class="nav-subtext">Data Explorer</span>
									</div>
								{/if}
							</a>

							<a
								href="/users"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/users')}
							>
								<div class="nav-icon-container">
									<Users class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">PLAYER_ACCOUNTS</span>
										<span class="nav-subtext">User Records</span>
									</div>
								{/if}
							</a>

							<a
								href="/redeye"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/redeye')}
							>
								<div class="nav-icon-container">
									<ShieldCheck class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">FIREWALL</span>
										<span class="nav-subtext">Security Management</span>
									</div>
								{/if}
							</a>

							<a
								href="/notes"
								class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}"
								class:active={isRouteActive('/notes')}
							>
								<div class="nav-icon-container">
									<StickyNote class="w-4 h-4" />
								</div>
								{#if !isSidebarCollapsed}
									<div class="flex flex-col">
										<span class="nav-text">NOTES & TASKS</span>
										<span class="nav-subtext">System Journal</span>
									</div>
								{/if}
							</a>
						</div>
					</nav>

					<!-- Sidebar Footer: Terminal Diagnostics -->
					<div
						class="mt-auto p-4 border-t-2 border-stone-800 bg-black/60 flex flex-col gap-4 transform transition-all duration-700 {sidebarLoaded
							? 'translate-y-0 opacity-100'
							: 'translate-y-8 opacity-0'}"
					>
						{#if !isSidebarCollapsed}
							<div class="grid grid-cols-2 gap-2 mb-2">
								<div class="flex flex-col">
									<span class="text-[7px] text-stone-600 font-mono uppercase">Refresh Rate</span>
									<span class="text-[9px] font-bold text-rust-light font-mono">124.5 Hz</span>
								</div>
								<div class="text-right flex flex-col">
									<span class="text-[7px] text-stone-600 font-mono uppercase">Status</span>
									<span class="text-[9px] font-bold text-emerald-500 font-mono">OK</span>
								</div>
							</div>
						{/if}

						<div class="flex items-center gap-2 {isSidebarCollapsed ? 'flex-col' : ''}">
							<button
								onclick={toggleTheme}
								class="p-2 border border-stone-800 bg-stone-950/50 text-stone-500 hover:text-white hover:border-rust transition-all flex-1 flex justify-center"
								title="Toggle Theme"
							>
								{#if $theme === 'dark'}<Moon class="w-4 h-4"/>{:else}<Sun class="w-4 h-4"/>{/if}
							</button>
							<button
								onclick={logout}
								class="p-2 border border-red-900/30 bg-red-950/10 text-red-600 hover:bg-red-600 hover:text-black transition-all flex-1 flex justify-center group"
								title="Logout"
							>
								<code class="text-[10px] font-black group-hover:animate-pulse">EXIT</code>
							</button>
						</div>
					</div>
				</div>
			</aside>

			<div class="flex-1 flex flex-col h-full overflow-hidden relative">
				<!-- Mobile Top Header -->
				<header
					class="md:hidden h-14 bg-black/80 backdrop-blur-md border-b-2 border-stone-800 flex items-center justify-between px-4 z-30 shrink-0 tactical-border"
				>
					<div class="corner-bl"></div>
					<div class="corner-br"></div>
					<div class="flex items-center gap-3">
						<button 
							onclick={() => isMobileMenuOpen = true}
							class="p-2 -ml-2 text-stone-400 hover:text-white transition-colors"
						>
							<Menu class="w-6 h-6" />
						</button>
						<div class="flex flex-col">
							<div class="flex items-center gap-2">
								<div class="w-1.5 h-1.5 bg-rust shadow-[0_0_5px_var(--color-rust)]"></div>
								<h1 class="text-sm font-black military-label text-white tracking-tighter uppercase">
									EXILE_<span class="text-rust-light">OS</span>
								</h1>
							</div>
							<span class="text-[6px] font-mono text-stone-600 tracking-[0.3em]">MOBILE_INTERFACE_V1</span>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<button
							onclick={toggleTheme}
							class="p-2 border border-stone-800 bg-stone-900/50 text-stone-500 hover:text-white transition-all"
						>
							{#if $theme === 'dark'}<Moon class="w-4 h-4"/>{:else}<Sun class="w-4 h-4"/>{/if}
						</button>
						<button
							onclick={logout}
							class="p-2 border border-red-900/30 bg-red-950/10 text-red-600 hover:bg-red-600 hover:text-black transition-all group"
						>
							<code class="text-[8px] font-black uppercase">Exit_</code>
						</button>
					</div>
				</header>

				<!-- Mobile Sidebar Overlay -->
				{#if isMobileMenuOpen}
					<div 
						class="fixed inset-0 z-[150] bg-black/60 backdrop-blur-sm md:hidden"
						transition:fade={{ duration: 200 }}
						onclick={() => isMobileMenuOpen = false}
						onkeydown={(e) => e.key === 'Escape' && (isMobileMenuOpen = false)}
						role="button"
						tabindex="0"
					>
						<aside 
							class="w-72 h-full bg-black border-r-2 border-stone-800 flex flex-col shadow-[0_0_50px_rgba(0,0,0,0.5)]"
							transition:slide={{ axis: 'x', duration: 300 }}
							onclick={(e) => e.stopPropagation()}
							onkeydown={(e) => e.stopPropagation()}
							role="none"
						>
							<div class="p-6 border-b-2 border-stone-800 bg-black/40 flex items-center justify-between">
								<div class="flex flex-col">
									<div class="flex items-center gap-2">
										<div class="w-2 h-2 bg-rust shadow-[0_0_8px_var(--color-rust)]"></div>
										<h1 class="text-xl font-black military-label text-white tracking-tighter uppercase">
											EXILE_<span class="text-rust-light">OS</span>
										</h1>
									</div>
									<span class="text-[8px] font-mono text-stone-600 mt-1 tracking-[0.3em]">MOBILE_V1.0</span>
								</div>
								<button 
									onclick={() => isMobileMenuOpen = false}
									class="p-2 text-stone-500 hover:text-white"
								>
									<Trash2 class="w-5 h-5 rotate-45" />
								</button>
							</div>

							<nav class="flex-1 p-3 space-y-2 overflow-y-auto">
								<a href="/dashboard" class="nav-link-industrial" class:active={isRouteActive('/dashboard') || isRouteActive('/')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><LayoutDashboard class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">CORE_INTERFACE</span><span class="nav-subtext">Unified_Dash</span></div>
								</a>
								<a href="/performance" class="nav-link-industrial" class:active={isRouteActive('/performance')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><Activity class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">TELEMETRY_BUS</span><span class="nav-subtext">Metric_Stream</span></div>
								</a>
								<a href="/config" class="nav-link-industrial" class:active={isRouteActive('/config')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><SettingsIcon class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">SYS_PARAMETERS</span><span class="nav-subtext">Config_Buffer</span></div>
								</a>
								<a href="/server" class="nav-link-industrial" class:active={isRouteActive('/server')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><HardDrive class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">ASSET_INDEX</span><span class="nav-subtext">Binary_Storage</span></div>
								</a>
								<a href="/database" class="nav-link-industrial" class:active={isRouteActive('/database')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><Database class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">DATA_ARCHIVE</span><span class="nav-subtext">Persistence_Cores</span></div>
								</a>
								<a href="/users" class="nav-link-industrial" class:active={isRouteActive('/users')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><Users class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">IDENTITY_VAULT</span><span class="nav-subtext">Client_Protocols</span></div>
								</a>
								<a href="/redeye" class="nav-link-industrial" class:active={isRouteActive('/redeye')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><ShieldCheck class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">NETWORK_SHIELD</span><span class="nav-subtext">Security_Sentinel</span></div>
								</a>
								<a href="/notes" class="nav-link-industrial" class:active={isRouteActive('/notes')} onclick={() => isMobileMenuOpen = false}>
									<div class="nav-icon-container"><StickyNote class="w-4 h-4" /></div>
									<div class="flex flex-col"><span class="nav-text">SIGNAL_LOGS</span><span class="nav-subtext">Notation_Drive</span></div>
								</a>
							</nav>

							<div class="p-6 border-t-2 border-stone-800 bg-black/60">
								<button onclick={logout} class="w-full p-3 border border-red-900/30 bg-red-950/10 text-red-600 font-black uppercase text-xs tracking-widest hover:bg-red-600 hover:text-black transition-all">
									Deauthenticate_Session
								</button>
							</div>
						</aside>
					</div>
				{/if}

				<!-- Main Content -->
				<main class="flex-1 overflow-auto relative">
					<div class="w-full px-4 sm:px-6 md:px-10 py-8 md:py-12 min-h-full pb-32 md:pb-12">
						{@render children()}
					</div>
				</main>

				<!-- Mobile Bottom Nav -->
				<nav
					class="md:hidden h-16 bg-black/90 backdrop-blur-xl border-t-2 border-stone-800 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 safe-area-pb"
				>
					<a
						href="/dashboard"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/dashboard'
						) || isRouteActive('/')
							? 'text-rust-light'
							: 'text-stone-600'} transition-colors"
					>
						<LayoutDashboard class="w-5 h-5" />
						<span class="font-mono text-[8px] font-black uppercase">CORE</span>
					</a>
					<a
						href="/performance"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/performance'
						)
							? 'text-rust-light'
							: 'text-stone-600'} transition-colors"
					>
						<Activity class="w-5 h-5" />
						<span class="font-mono text-[8px] font-black uppercase">PERF</span>
					</a>
					<a
						href="/config"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/config'
						) || isRouteActive('/config/')
							? 'text-rust-light'
							: 'text-stone-600'} transition-colors"
					>
						<SettingsIcon class="w-5 h-5" />
						<span class="font-mono text-[8px] font-black uppercase">CNFG</span>
					</a>
					<a
						href="/redeye"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/redeye'
						)
							? 'text-rust-light'
							: 'text-stone-600'} transition-colors"
					>
						<ShieldCheck class="w-5 h-5" />
						<span class="font-mono text-[8px] font-black uppercase">SHLD</span>
					</a>
				</nav>
			</div>
		</div>
	</div>

		<ServerStatus 
			status={$isConnected ? 'ONLINE' : 'OFFLINE'} 
			players={$stats.active_spawners * 10} 
			servers={$stats.active_spawners} 
		/>

		<Notifications />
	{:else}
		{@render children()}
	{/if}

	<NoteModal
		bind:isOpen={showGlobalNoteModal}
		note={null}
		onSave={handleGlobalSaveNote}
		onClose={() => (showGlobalNoteModal = false)}
	/>
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
		opacity: 1;
		transform: translateX(0);
	}

	:global(.dark) .nav-link {
		color: rgb(148, 163, 184);
	}
	:global(:not(.dark)) .nav-link {
		color: rgb(100, 116, 139);
	}

	:global(.dark) .nav-link:hover {
		background: linear-gradient(to right, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
		color: rgb(226, 232, 240);
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
		border-color: rgba(255, 255, 255, 0.1);
		transform: scale(1.01);
	}

	:global(:not(.dark)) .nav-link:hover {
		background: linear-gradient(to right, rgba(0, 0, 0, 0.05), rgba(0, 0, 0, 0.02));
		color: rgb(15, 23, 42);
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
		border-color: rgba(0, 0, 0, 0.1);
		transform: scale(1.01);
	}

	.nav-link.nav-active {
		border-color: rgba(146, 64, 14, 0.4);
		transform: scale(1.02);
	}

	:global(.dark) .nav-link.nav-active {
		background: linear-gradient(to right, rgba(120, 53, 15, 0.3), rgba(120, 53, 15, 0.2));
		color: #f97316;
		box-shadow: 0 25px 50px -12px rgba(120, 53, 15, 0.25);
	}

	:global(:not(.dark)) .nav-link.nav-active {
		background: linear-gradient(to right, rgba(120, 53, 15, 0.15), rgba(120, 53, 15, 0.05));
		color: #78350f;
		box-shadow: 0 25px 50px -12px rgba(120, 53, 15, 0.1);
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
		background: linear-gradient(to bottom, #f97316, #92400e);
		border-radius: 0 0.25rem 0.25rem 0;
		box-shadow: 0 0 10px rgba(249, 115, 22, 0.5);
		animation: pulse 2s infinite;
	}
</style>
