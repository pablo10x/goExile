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
		Palette,
		Radio,
		Compass,
		Cpu,
		Terminal,
		Lock,
		Shield,
		Layers,
		BarChart3,
		ChevronRight,
		X
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
				
				// Advanced Atmospheric
				root.style.setProperty('--glitch-intensity', ($siteSettings.aesthetic.glitch_intensity ?? 0.05).toString());
				root.style.setProperty('--chromatic-aberration', ($siteSettings.aesthetic.chromatic_aberration ?? 0.02).toString());
				root.style.setProperty('--flicker-intensity', ($siteSettings.aesthetic.flicker_intensity ?? 0.01).toString());
				
				// Specialized Colors
				root.style.setProperty('--color-success', $siteSettings.aesthetic.success_color || '#10b981');
				root.style.setProperty('--color-warning', $siteSettings.aesthetic.warning_color || '#f59e0b');
				root.style.setProperty('--color-danger', $siteSettings.aesthetic.danger_color || '#ef4444');
				root.style.setProperty('--color-info', $siteSettings.aesthetic.info_color || '#06b6d4');

				// Geometry & Grid
				root.style.setProperty('--grid-opacity', ($siteSettings.aesthetic.grid_opacity ?? 0.05).toString());
				root.style.setProperty('--border-opacity', ($siteSettings.aesthetic.border_opacity ?? 0.3).toString());

				// Micro-Typography
				root.style.setProperty('--text-letter-spacing', ($siteSettings.aesthetic.letter_spacing ?? 0.05) + 'em');
				root.style.setProperty('--text-line-height', ($siteSettings.aesthetic.line_height ?? 1.5).toString());
				root.style.setProperty('--text-transform', $siteSettings.aesthetic.text_transform || 'uppercase');
				root.style.setProperty('--heading-weight', $siteSettings.aesthetic.heading_weight || '900');
				root.style.setProperty('--base-weight', $siteSettings.aesthetic.font_weight_base || '400');
				root.style.setProperty('--mono-weight', $siteSettings.aesthetic.font_weight_mono || '400');
				root.style.setProperty('--paragraph-spacing', ($siteSettings.aesthetic.paragraph_spacing ?? 1.0) + 'em');
				root.style.setProperty('--text-glow-intensity', ($siteSettings.aesthetic.text_glow_intensity ?? 0.5).toString());

				// Kinetic Physics
				root.style.setProperty('--global-transition', ($siteSettings.aesthetic.global_transition_speed ?? 300) + 'ms');
				root.style.setProperty('--hover-scale', ($siteSettings.aesthetic.hover_scale_factor ?? 1.02).toString());
				root.style.setProperty('--press-depth', ($siteSettings.aesthetic.button_press_depth ?? 2) + 'px');

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
			<!-- System Ticker Header -->
			<div class="fixed top-0 left-0 right-0 h-7 bg-black border-b border-stone-800 z-[120] flex items-center px-4 overflow-hidden shadow-[0_4px_20px_rgba(0,0,0,0.8)]">
				<div class="flex items-center gap-8 animate-text-reveal whitespace-nowrap w-full">
					<div class="flex items-center gap-2 shrink-0">
						<div class="w-1 h-1 rounded-full bg-rust animate-ping"></div>
						<span class="tactical-code text-rust font-black">[SYSTEM_UPLINK_ESTABLISHED]</span>
					</div>
					<div class="flex items-center gap-6 text-stone-600">
						<span class="tactical-code">NET_STATUS: <span class="text-stone-400">{$connectionStatus}</span></span>
						<span class="tactical-code">NODES_ACTIVE: <span class="text-stone-400">{$stats.active_spawners}</span></span>
						<span class="tactical-code hidden sm:inline">ENTROPY: <span class="text-stone-400">0.0042</span></span>
						<span class="tactical-code">TIMESTAMP: <span class="text-stone-400 font-mono tracking-tighter">{new Date().toISOString()}</span></span>
					</div>
					<div class="ml-auto flex items-center gap-4 shrink-0">
						<span class="tactical-code text-rust-light italic hidden md:inline">SECURITY_LEVEL: 4_OMNI</span>
						<div class="w-[2px] h-3 bg-stone-800"></div>
						<div class="flex gap-1">
							{#each [1,2,3] as i}<div class="w-1 h-1 bg-stone-800 rounded-full"></div>{/each}
						</div>
					</div>
				</div>
			</div>

			<!-- Background/Atmospheric Overlays (Lower Z-Index) -->
			<div class="fixed inset-0 z-[10] pointer-events-none overflow-hidden">
				<div class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[length:100%_4px,3px_100%]" style="opacity: {localSiteSettings?.aesthetic?.scanlines_opacity || 0.05}"></div>
				<div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/asfalt-dark.png')] opacity-20 mix-blend-overlay" style="opacity: {localSiteSettings?.aesthetic?.noise_opacity || 0.03}"></div>
			</div>

			{#if localBackgroundConfig?.show_clouds}
				<div class="fixed inset-0 z-[5] clouds-overlay" style="opacity: {localBackgroundConfig.clouds_opacity}"></div>
			{/if}

			{#if localBackgroundConfig?.show_rain}
				<div class="fixed inset-0 z-[5] rain-container">
					<div class="rain-layer rain-layer-back"></div>
					<div class="rain-layer rain-layer-mid"></div>
					<div class="rain-layer rain-layer-front" style="opacity: {localBackgroundConfig.rain_opacity * 0.5}"></div>
				</div>
			{/if}

			{#if localBackgroundConfig?.show_smoke}
				<div class="fixed inset-0 z-[5] pointer-events-none"><GlobalSmoke /></div>
			{/if}

			{#if localBackgroundConfig?.show_vignette}
				<div class="vignette z-[100]"></div>
			{/if}

			{#if localBackgroundConfig?.global_type && localBackgroundConfig.global_type !== 'none'}
				<div class="fixed inset-0 z-0 pointer-events-none overflow-hidden opacity-50">
					<SectionBackground type={localBackgroundConfig.global_type} />
				</div>
			{/if}

			<div
				class="flex h-screen text-stone-400 overflow-hidden relative bg-transparent transition-colors duration-300 pt-7"
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

					<nav class="flex-1 p-3 space-y-6 overflow-y-auto overflow-x-hidden no-scrollbar">
						<!-- Terminal Output Simulation -->
						{#if !isSidebarCollapsed}
							<div class="px-2 py-1 mb-4 border-l-2 border-rust/30 opacity-40">
								<div class="tactical-code text-stone-500 animate-pulse">>> COMMAND_MODULE_ACTIVE</div>
								<div class="tactical-code text-stone-600">>> HANDSHAKE_OK</div>
							</div>
						{/if}

						<div class="space-y-6">
							<!-- CATEGORY: CORE -->
							<div class="space-y-1">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Core_Systems</span>{/if}
								<a href="/dashboard" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/dashboard') || isRouteActive('/')}>
									<div class="nav-icon-container"><LayoutDashboard class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">CORE_DASH</span><span class="nav-subtext">Unified Interface</span></div>{/if}
								</a>
								<a href="/performance" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/performance')}>
									<div class="nav-icon-container"><Activity class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">PERFORMANCE</span><span class="nav-subtext">RT_Telemetry</span></div>{/if}
								</a>
							</div>

							<!-- CATEGORY: FLEET -->
							<div class="space-y-1">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Fleet_Ops</span>{/if}
								<a href="/server" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/server')}>
									<div class="nav-icon-container"><Server class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">NODE_FLEET</span><span class="nav-subtext">Spawner_Matrix</span></div>{/if}
								</a>
								<a href="/users" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/users')}>
									<div class="nav-icon-container"><Users class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">IDENTITIES</span><span class="nav-subtext">Subject Registry</span></div>{/if}
								</a>
							</div>

							<!-- CATEGORY: LOGISTICS -->
							<div class="space-y-1">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Logistics</span>{/if}
								<a href="/database" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/database')}>
									<div class="nav-icon-container"><Database class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">DATABASE</span><span class="nav-subtext">Data_Explorer</span></div>{/if}
								</a>
								<a href="/notes" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/notes')}>
									<div class="nav-icon-container"><StickyNote class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">JOURNAL</span><span class="nav-subtext">Task_Buffer</span></div>{/if}
								</a>
							</div>

							<!-- CATEGORY: CONFIGURATION -->
							<div class="space-y-1">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Calibrations</span>{/if}
								<a href="/config" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/config')}>
									<div class="nav-icon-container"><SettingsIcon class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">SYSTEM_CONFIG</span><span class="nav-subtext">Kernel_Params</span></div>{/if}
								</a>
								<a href="/config/theme" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/config/theme')}>
									<div class="nav-icon-container"><Palette class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">THEME_LAB</span><span class="nav-subtext">Visual_Sync</span></div>{/if}
								</a>
								<a href="/redeye" class="nav-link-industrial {isSidebarCollapsed ? 'justify-center' : ''}" class:active={isRouteActive('/redeye')}>
									<div class="nav-icon-container"><ShieldCheck class="w-4 h-4" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text">SENTINEL</span><span class="nav-subtext">Security_Shield</span></div>{/if}
								</a>
							</div>
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
					class="md:hidden h-14 bg-black border-b border-stone-800 flex items-center justify-between px-4 z-[130] shrink-0 relative"
				>
					<div class="flex items-center gap-4">
						<button 
							onclick={() => isMobileMenuOpen = true}
							class="p-2 -ml-2 text-stone-500 hover:text-white transition-colors border border-transparent hover:border-stone-800"
						>
							<Menu class="w-6 h-6" />
						</button>
						<div class="flex flex-col">
							<div class="flex items-center gap-2">
								<div class="w-1.5 h-1.5 bg-rust shadow-[0_0_5px_var(--color-rust)] animate-pulse"></div>
								<h1 class="text-sm font-black military-label text-white tracking-tighter uppercase">
									EXILE_<span class="text-rust-light">OS</span>
								</h1>
							</div>
							<span class="text-[6px] font-mono text-stone-600 tracking-[0.3em]">NODE_MOBILE_V1</span>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="hidden sm:flex flex-col items-end mr-2">
							<span class="text-[6px] text-stone-700 uppercase font-mono">Status</span>
							<span class="text-[8px] text-emerald-500 font-mono font-bold uppercase">Linked</span>
						</div>
						<button
							onclick={toggleTheme}
							class="p-2 border border-stone-800 bg-stone-900/50 text-stone-500 hover:text-white transition-all"
						>
							{#if $theme === 'dark'}<Moon class="w-4 h-4"/>{:else}<Sun class="w-4 h-4"/>{/if}
						</button>
					</div>
				</header>

				<!-- Mobile Sidebar Overlay -->
				{#if isMobileMenuOpen}
					<div 
						class="fixed inset-0 z-[200] bg-black/80 backdrop-blur-md md:hidden"
						transition:fade={{ duration: 200 }}
						onclick={() => isMobileMenuOpen = false}
						onkeydown={(e) => e.key === 'Escape' && (isMobileMenuOpen = false)}
						role="button"
						tabindex="0"
					>
						<aside 
							class="w-80 h-full bg-black border-r border-stone-800 flex flex-col shadow-[0_0_100px_rgba(0,0,0,1)] relative"
							transition:slide={{ axis: 'x', duration: 400, easing: cubicOut }}
							onclick={(e) => e.stopPropagation()}
							onkeydown={(e) => e.stopPropagation()}
							role="none"
						>
							<!-- Mobile Sidebar Tactical Corners -->
							<div class="corner-tr opacity-20"></div>
							<div class="corner-br opacity-20"></div>

							<div class="p-8 border-b border-stone-800 bg-stone-950 flex items-center justify-between">
								<div class="flex flex-col">
									<div class="flex items-center gap-3">
										<div class="w-2 h-2 bg-rust shadow-[0_0_8px_var(--color-rust)]"></div>
										<h1 class="text-2xl font-black military-label text-white tracking-tighter uppercase leading-none">
											EXILE_<span class="text-rust-light">CORE</span>
										</h1>
									</div>
									<span class="text-[8px] font-mono text-stone-600 mt-2 tracking-[0.4em]">MOBILE_AUTH_TERMINAL</span>
								</div>
								<button 
									onclick={() => isMobileMenuOpen = false}
									class="p-2 text-stone-600 hover:text-white border border-stone-800 hover:border-rust transition-all"
								>
									<X class="w-6 h-6" />
								</button>
							</div>

							<nav class="flex-1 p-4 space-y-6 overflow-y-auto custom-scrollbar">
								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Command</span>
									<a href="/dashboard" class="nav-link-industrial" class:active={isRouteActive('/dashboard') || isRouteActive('/')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><LayoutDashboard class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">INTERFACE</span><span class="nav-subtext">Core Dashboard</span></div>
									</a>
									<a href="/performance" class="nav-link-industrial" class:active={isRouteActive('/performance')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Activity class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">TELEMETRY</span><span class="nav-subtext">Real-time Stream</span></div>
									</a>
								</div>

								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Assets</span>
									<a href="/server" class="nav-link-industrial" class:active={isRouteActive('/server')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><HardDrive class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">FILE_SYS</span><span class="nav-subtext">Binary Storage</span></div>
									</a>
									<a href="/database" class="nav-link-industrial" class:active={isRouteActive('/database')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Database class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">PERSISTENCE</span><span class="nav-subtext">Data Archive</span></div>
									</a>
									<a href="/users" class="nav-link-industrial" class:active={isRouteActive('/users')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Users class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">SUBJECTS</span><span class="nav-subtext">User Registry</span></div>
									</a>
								</div>

								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Security</span>
									<a href="/redeye" class="nav-link-industrial" class:active={isRouteActive('/redeye')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><ShieldCheck class="w-4 h-4" /></div>
										<div class="flex flex-col"><span class="nav-text">SENTINEL</span><span class="nav-subtext">Network Shield</span></div>
									</a>
								</div>
							</nav>

							<div class="p-8 border-t border-stone-800 bg-stone-950 flex flex-col gap-4">
								<div class="flex items-center justify-between mb-2">
									<div class="flex flex-col">
										<span class="text-[7px] text-stone-700 font-mono uppercase">Session_Token</span>
										<span class="text-[9px] text-stone-500 font-mono">0x{Math.random().toString(16).slice(2, 10).toUpperCase()}</span>
									</div>
									<div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
								</div>
								<button onclick={logout} class="w-full py-4 bg-red-900/10 border border-red-900/30 text-red-600 font-heading font-black uppercase text-xs tracking-widest hover:bg-red-600 hover:text-white transition-all shadow-lg active:translate-y-px">
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
</style>
