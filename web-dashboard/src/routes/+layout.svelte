<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { get } from 'svelte/store';
	import { fade, slide } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		stats,
		nodes,
		notifications,
		serverVersions,
		isConnected,
		connectionStatus,
		loadAllSettings,
		theme,
		siteSettings,
		showQuickActions,
		backgroundConfig,
		isAuthenticated,
		notes,
		restartRequired,
		lowPowerMode,
		sysState
	} from '$lib/stores.svelte';
	import type { Note } from '$lib/stores.svelte';
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
		ShieldAlert,
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
		X,
		Gauge,
		FileText,
		Sliders,
		AlertCircle
	} from 'lucide-svelte';
	import QuickActionsTooltip from '$lib/components/QuickActionsTooltip.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';

	import NavbarParticles from '$lib/components/theme/NavbarParticles.svelte';
	import MotherboardBackground from '$lib/components/theme/MotherboardBackground.svelte';
	import ServerStatus from '$lib/components/theme/ServerStatus.svelte';
	import Notifications from '$lib/components/theme/Notifications.svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import Button from '$lib/components/Button.svelte';
	import CommandPalette from '$lib/components/CommandPalette.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import ShortcutHelpModal from '$lib/components/ShortcutHelpModal.svelte';

	let { children, data } = $props();
	let isChecking = $state(true);
	let restarting = $state(false);
	let isCommandPaletteOpen = $state(false);
	let isShortcutHelpOpen = $state(false);
	let eventSource: EventSource | null = null;

	let localBackgroundConfig = $derived($backgroundConfig);
	let localSiteSettings = $derived($siteSettings);

	// Keyboard shortcut orchestration
	onMount(() => {
		const handleGlobalKeydown = (e: KeyboardEvent) => {
			// Ctrl+K or Cmd+K for Command Palette
			if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
				e.preventDefault();
				isCommandPaletteOpen = !isCommandPaletteOpen;
			}

			// Don't trigger shortcuts if user is typing in an input
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

			// G + [key] navigation pattern
			if (e.key === 'g') {
				const nextKey = (ev: KeyboardEvent) => {
					if (ev.key === 'd') goto('/dashboard');
					if (ev.key === 'l') goto('/logs');
					if (ev.key === 'p') goto('/performance');
					if (ev.key === 't') goto('/config/theme');
					if (ev.key === 'n') goto('/server');
					window.removeEventListener('keydown', nextKey);
				};
				window.addEventListener('keydown', nextKey, { once: true });
			}
		};

		window.addEventListener('keydown', handleGlobalKeydown);
		return () => window.removeEventListener('keydown', handleGlobalKeydown);
	});

	// Sync isAuthenticated and stats store with server-side data on load
	$effect.pre(() => {
		if (data?.isAuthenticated !== undefined) {
			isAuthenticated.set(data.isAuthenticated);
		}
		if (data?.stats) {
			stats.set(data.stats);
		}
	});

	// Theme handling
	$effect(() => {
		if (typeof window !== 'undefined') {
			if ($theme === 'dark') {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
			localStorage.setItem('theme', $theme);
		}
	});

	function toggleTheme() {
		// Theme is now hardcoded to dark, this function might be redundant or can be kept for future expansion
		// For now, we'll keep it simple
		// theme.update((t) => (t === 'dark' ? 'light' : 'dark'));
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
				} else if (data.type === 'nodes') {
					const list: any[] = Array.isArray(data.payload)
						? data.payload
						: Object.values(data.payload);
					list.sort((a, b) => a.id - b.id);
					nodes.set(list);
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
			const promises: Promise<any>[] = [
				fetch('/api/nodes', { cache: 'no-store', credentials: 'include' }),
				fetch('/api/versions', { cache: 'no-store', credentials: 'include' })
			];

			// Only fetch stats if not already provided by server load
			const currentData = $state.snapshot(data);
			if (!currentData?.stats) {
				promises.push(fetch('/api/stats', { cache: 'no-store', credentials: 'include' }));
			}

			const results = await Promise.all(promises);
			
			const nodesRes = results[0];
			const versionsRes = results[1];
			
			if (nodesRes.ok) nodes.set(await nodesRes.json());
			if (versionsRes.ok) serverVersions.set(await versionsRes.json());
			
			if (results.length > 2) {
				const statsRes = results[2];
				if (statsRes.ok) stats.set(await statsRes.json());
			} else if (currentData?.stats) {
				stats.set(currentData.stats);
			}
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



	onMount(() => {
		// Visibility-aware background sync
		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') {
				if (get(isAuthenticated)) connectSSE();
			} else {
				if (eventSource) {
					eventSource.close();
					eventSource = null;
					isConnected.set(false);
					connectionStatus.set('Standby (Hidden)');
				}
			}
		};
		document.addEventListener('visibilitychange', handleVisibilityChange);

		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') {
			isChecking = false;
		} else {
			// If server-side load already determined we are authenticated,
			// we can proceed to fetch data immediately.
			if (data?.isAuthenticated) {
				isAuthenticated.set(true);
				isChecking = false;
				connectSSE();
				loadAllSettings().then(() => initialFetch());
			} else {
				// Otherwise, verify or redirect
				checkAuth().then(() => {
					if (get(isAuthenticated)) {
						loadAllSettings().then(() => initialFetch());
					}
				});
			}
		}

		setTimeout(() => {
			sidebarLoaded = true;
		}, 300);

		return () => {
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		};
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

	onMount(() => {
		let rafId: number;
		const handleMouseMove = (e: MouseEvent) => {
			if (rafId) cancelAnimationFrame(rafId);
			rafId = requestAnimationFrame(() => {
				document.documentElement.style.setProperty('--mouse-x-global', `${e.clientX}px`);
				document.documentElement.style.setProperty('--mouse-y-global', `${e.clientY}px`);
			});
		};
		window.addEventListener('mousemove', handleMouseMove, { passive: true });
		return () => {
			window.removeEventListener('mousemove', handleMouseMove);
			if (rafId) cancelAnimationFrame(rafId);
		};
	});
</script>

{#if isChecking}
	<div
		class="flex items-center justify-center min-h-screen bg-black"
	>
		<div class="relative">
			<div
				class="animate-spin h-16 w-16 border-4 border-neutral-900 border-t-rust shadow-[0_0_20px_rgba(194,65,12,0.3)]"
			></div>
			<div
				class="absolute inset-0 bg-rust/10 blur-2xl animate-pulse"
			></div>
		</div>
	</div>
{:else}
				{#if $isAuthenticated && page.url.pathname !== '/login'}
					<div class="relative min-h-screen selection:bg-rust selection:text-white">
						<!-- Global Scanline Overlay -->
						<div class="scanline-overlay"></div>

						<!-- Solid Background Layer -->
						<div class="fixed inset-0 z-[-100] industrial-gradient"></div>
		
					<!-- System Status Bar -->
					<div class="fixed top-0 left-0 right-0 h-6 bg-neutral-900 border-b-2 border-neutral-800 z-[120] flex items-center px-4 overflow-hidden shadow-sm backdrop-blur-md">
						<div class="flex items-center gap-8 whitespace-nowrap w-full">
							<div class="flex items-center gap-2 shrink-0">
								<div class="w-1.5 h-1.5 bg-rust animate-pulse shadow-[0_0_8px_#c2410c]"></div>
								<span class="font-mono font-black text-[9px] text-rust-light uppercase tracking-widest">[KERNEL_UPLINK_ESTABLISHED]</span>
							</div>
							<div class="flex items-center gap-6 text-neutral-500 font-mono font-black text-[8px] uppercase tracking-[0.2em]">
								<span>Network: <span class="text-neutral-300">{$connectionStatus}</span></span>
								<span>Fleet: <span class="text-neutral-300">{$stats.active_nodes}</span></span>
								<span class="hidden sm:inline">Entropy: <span class="text-emerald-500">Optimal</span></span>
								<span>Cycle: <span class="text-neutral-300">{new Date().toLocaleTimeString([], { hour12: false })}</span></span>
							</div>
							<div class="ml-auto flex items-center gap-4 shrink-0">
								<button 
									onclick={() => isShortcutHelpOpen = true}
									class="flex items-center gap-1.5 text-[8px] font-black text-neutral-500 hover:text-rust-light transition-colors uppercase tracking-widest font-mono"
									title="Keyboard Shortcuts"
								>
									<Icon name="ph:question-bold" size="0.7rem" />
									<span>Shortcuts</span>
								</button>
								<div class="w-[1px] h-3 bg-neutral-800"></div>
								<span class="text-[8px] text-rust/60 font-black hidden md:inline font-mono tracking-widest">AES_256_ACTIVE</span>
								<div class="w-[1px] h-3 bg-neutral-800"></div>
								<div class="flex gap-1">
									{#each [1,2,3] as i}<div class="w-1 h-1 bg-neutral-800"></div>{/each}
								</div>
							</div>
						</div>
					</div>
		
											<!-- Background/Atmospheric Overlays (Lower Z-Index) -->
											<div class="fixed inset-0 z-[-50] pointer-events-none overflow-hidden opacity-20 grayscale">
												<MotherboardBackground />
											</div>		
					
		
								<div
		
									class="flex h-screen text-neutral-400 overflow-hidden relative bg-transparent transition-colors duration-300 pt-6"
		
								>			<!-- Global Restart Banner -->
			{#if $restartRequired}
				<div
					class="absolute top-0 md:left-64 left-0 right-0 z-50 bg-rust text-white px-4 py-3 flex justify-between items-center shadow-2xl border-b-2 border-rust-light/50 animate-slide-fade text-xs md:text-sm industrial-frame"
				>
					<div class="flex items-center gap-4">
						<AlertCircle class="w-5 h-5 shrink-0 animate-flicker" />
						<span class="font-mono font-black uppercase tracking-widest truncate">Critical: Kernel reboot required to commit parameter delta</span>
					</div>
					<button
						onclick={restartServer}
						disabled={restarting}
						class="px-6 py-2 bg-white text-rust font-mono font-black text-[10px] uppercase tracking-widest hover:bg-neutral-100 active:scale-95 transition-all shadow-xl disabled:opacity-50"
					>
						{restarting ? 'Executing...' : 'Commence_Reboot'}
					</button>
				</div>
			{/if}

			<!-- Desktop Sidebar -->
			<aside
				class="hidden md:flex relative transition-all duration-500 ease-[cubic-bezier(0.23,1,0.32,1)] bg-[#050505] border-r-2 border-neutral-800 flex-col shrink-0 overflow-hidden z-20 {isSidebarCollapsed
					? 'w-20'
					: 'w-64'} shadow-[10px_0_30px_rgba(0,0,0,0.5)]"
			>
				<!-- Inner Fitted Glow -->
				<div class="absolute inset-y-0 right-0 w-[1px] bg-white/5 pointer-events-none"></div>
				
				{#if true}
					<NavbarParticles />
				{/if}
				
				<div class="relative z-10 flex flex-col h-full shadow-[inset_-1px_0_0_rgba(255,255,255,0.02)]">
					<div
						class="p-6 border-b-2 border-neutral-800/80 bg-black/20 transform transition-all duration-700 {sidebarLoaded
							? 'tranneutral-y-0 opacity-100'
							: '-tranneutral-y-4 opacity-0'} flex items-center {isSidebarCollapsed ? 'justify-center' : 'justify-between'}"
					>
						{#if !isSidebarCollapsed}
							<div class="flex flex-col animate-in fade-in zoom-in duration-500">
								<div class="flex items-center gap-2">
									<div class="w-2 h-2 bg-rust shadow-[0_0_8px_rgba(194,65,12,0.5)]"></div>
									<h1 class="text-xl font-black text-white tracking-tighter uppercase italic leading-none">
										EXILE_<span class="text-rust-light">OS</span>
									</h1>
								</div>
								<span class="text-[8px] font-mono text-neutral-600 mt-2 tracking-[0.4em] font-black">STATION_PRO_4.2</span>
							</div>
						{/if}
						<button
							onclick={toggleSidebar}
							class="p-2 border border-neutral-800 bg-neutral-950 text-neutral-500 hover:text-rust-light hover:border-rust/30 transition-all duration-300"
						>
							<Icon name={isSidebarCollapsed ? 'ph:caret-double-right-bold' : 'ph:caret-double-left-bold'} size="0.9rem" />
						</button>
					</div>

					<nav class="flex-1 p-3 space-y-8 overflow-y-auto overflow-x-hidden no-scrollbar py-8">
						<div class="space-y-8">
							<!-- CATEGORY: DASHBOARD -->
							<div class="space-y-1.5">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-neutral-700 tracking-[0.4em] ml-3 mb-3 block uppercase italic">Core_Interface</span>{/if}
								<a href="/dashboard" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/dashboard') || isRouteActive('/')} title={isSidebarCollapsed ? 'Overview' : ''}>
									<div class="nav-icon-container-light"><Icon name="gauge" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Telemetry</span><span class="nav-subtext-light">System Dashboard</span></div>{/if}
								</a>
								<a href="/performance" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/performance')} title={isSidebarCollapsed ? 'Performance' : ''}>
									<div class="nav-icon-container-light"><Icon name="activity" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Analytics</span><span class="nav-subtext-light">Real-time Metrics</span></div>{/if}
								</a>
							</div>

							<!-- CATEGORY: MANAGEMENT -->
							<div class="space-y-1.5">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-neutral-700 tracking-[0.4em] ml-3 mb-3 block uppercase italic">Infrastructure</span>{/if}
								{#each [
									{ href: '/server', icon: 'cpu', label: 'Fleet', sub: 'Server Operations' },
									{ href: '/users', icon: 'users', label: 'Subjects', sub: 'Access Registry' }
								] as link}
									<a href={link.href} class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive(link.href)} title={isSidebarCollapsed ? link.label : ''}>
										<div class="nav-icon-container-light"><Icon name={link.icon} size="1.1rem" /></div>
										{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">{link.label}</span><span class="nav-subtext-light">{link.sub}</span></div>{/if}
									</a>
								{/each}
							</div>

							<!-- CATEGORY: RESOURCES -->
							<div class="space-y-1.5">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-neutral-700 tracking-[0.4em] ml-3 mb-3 block uppercase italic">Persistence</span>{/if}
								<a href="/database" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/database')} title={isSidebarCollapsed ? 'Database' : ''}>
									<div class="nav-icon-container-light"><Icon name="database" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Archives</span><span class="nav-subtext-light">Neural Storage</span></div>{/if}
								</a>
								<a href="/notes" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/notes')} title={isSidebarCollapsed ? 'Operations' : ''}>
									<div class="nav-icon-container-light"><Icon name="file-text" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Worklog</span><span class="nav-subtext-light">Tactical Journal</span></div>{/if}
								</a>
							</div>

							<!-- CATEGORY: SETTINGS -->
							<div class="space-y-1.5">
								{#if !isSidebarCollapsed}<span class="text-[8px] font-black text-neutral-700 tracking-[0.4em] ml-3 mb-3 block uppercase italic">Command</span>{/if}
								<a href="/config" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/config')} title={isSidebarCollapsed ? 'Settings' : ''}>
									<div class="nav-icon-container-light"><Icon name="sliders" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Parameters</span><span class="nav-subtext-light">Environment</span></div>{/if}
								</a>
								<a href="/redeye" class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive('/redeye')} title={isSidebarCollapsed ? 'Security' : ''}>
									<div class="nav-icon-container-light"><Icon name="shield" size="1.1rem" /></div>
									{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light">Sentinel</span><span class="nav-subtext-light">Network Shield</span></div>{/if}
								</a>
							</div>
						</div>
					</nav>

					<!-- Sidebar Footer -->
					<div
						class="mt-auto p-4 border-t-2 border-neutral-800/50 bg-neutral-900/20 flex flex-col gap-3 transform transition-all duration-700 {sidebarLoaded
							? 'tranneutral-y-0 opacity-100'
							: 'tranneutral-y-8 opacity-0'}"
					>
						<div class="flex items-center gap-2 {isSidebarCollapsed ? 'flex-col' : ''}">
							<button
								onclick={() => lowPowerMode.update(v => !v)}
								class="p-2 border transition-all flex items-center justify-center gap-2 flex-1
								{$lowPowerMode 
									? 'bg-amber-500/10 border-amber-500/30 text-amber-500' 
									: 'bg-neutral-950 border-neutral-800 text-neutral-600 hover:text-white hover:border-neutral-600'}"
								title={$lowPowerMode ? 'Disable Eco Mode' : 'Enable Eco Mode'}
							>
								<Zap class="w-3.5 h-3.5 {$lowPowerMode ? '' : 'opacity-40'}" />
								{#if !isSidebarCollapsed}
									<span class="text-[9px] font-black font-mono tracking-widest uppercase">{$lowPowerMode ? 'ECO_ON' : 'PERF_MAX'}</span>
								{/if}
							</button>

							<button
								onclick={logout}
								class="p-2 border border-red-900/30 bg-red-950/10 text-red-600 hover:bg-red-600 hover:text-white transition-all flex items-center justify-center group shadow-lg"
								title="Terminate Session"
							>
								<Icon name="ph:power-bold" size="1rem" />
							</button>
						</div>
					</div>
				</div>
			</aside>

			<div class="flex-1 flex flex-col h-full overflow-hidden relative bg-transparent">
				<!-- Mobile Top Header -->
				<header
					class="md:hidden h-16 bg-neutral-950 border-b-2 border-neutral-800 flex items-center justify-between px-4 z-[130] shrink-0 relative backdrop-blur-md"
				>
					<div class="flex items-center gap-4">
						<button 
							onclick={() => isMobileMenuOpen = true}
							class="p-2 -ml-2 text-neutral-400 hover:text-rust transition-colors"
						>
							<Icon name="ph:list-bold" size="1.5rem" />
						</button>
						<div class="flex flex-col">
							<div class="flex items-center gap-2">
								<div class="w-1.5 h-1.5 bg-rust shadow-[0_0_8px_var(--color-rust)] animate-pulse"></div>
								<h1 class="text-xl font-black text-white tracking-tighter uppercase italic italic">
									EXILE_<span class="text-rust-light">OS</span>
								</h1>
							</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="hidden sm:flex flex-col items-end mr-2">
							<span class="text-[8px] text-emerald-500 font-mono font-black uppercase tracking-widest">Connected</span>
						</div>
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

							<div class="p-8 border-b border-stone-800 bg-[#050505] flex items-center justify-between">
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
									<Icon name="ph:x-bold" size="1.5rem" />
								</button>
							</div>

							<nav class="flex-1 p-4 space-y-6 overflow-y-auto custom-scrollbar">
								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Command</span>
									<a href="/dashboard" class="nav-link" class:active={isRouteActive('/dashboard') || isRouteActive('/')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="gauge" /></div>
										<div class="flex flex-col"><span class="nav-text">INTERFACE</span><span class="nav-subtext">Core Dashboard</span></div>
									</a>
									<a href="/performance" class="nav-link" class:active={isRouteActive('/performance')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="activity" /></div>
										<div class="flex flex-col"><span class="nav-text">TELEMETRY</span><span class="nav-subtext">Real-time Stream</span></div>
									</a>
								</div>

								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Assets</span>
									<a href="/server" class="nav-link" class:active={isRouteActive('/server')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="cpu" /></div>
										<div class="flex flex-col"><span class="nav-text">FILE_SYS</span><span class="nav-subtext">Binary Storage</span></div>
									</a>
									<a href="/database" class="nav-link" class:active={isRouteActive('/database')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="database" /></div>
										<div class="flex flex-col"><span class="nav-text">PERSISTENCE</span><span class="nav-subtext">Data Archive</span></div>
									</a>
									<a href="/users" class="nav-link" class:active={isRouteActive('/users')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="users" /></div>
										<div class="flex flex-col"><span class="nav-text">SUBJECTS</span><span class="nav-subtext">User Registry</span></div>
									</a>
								</div>

								<div class="space-y-1">
									<span class="text-[8px] font-black text-stone-700 tracking-[0.4em] ml-2 mb-2 block uppercase">Security</span>
									<a href="/redeye" class="nav-link" class:active={isRouteActive('/redeye')} onclick={() => isMobileMenuOpen = false}>
										<div class="nav-icon-container"><Icon name="shield" /></div>
										<div class="flex flex-col"><span class="nav-text">SENTINEL</span><span class="nav-subtext">Network Shield</span></div>
									</a>
								</div>
							</nav>

							<div class="p-8 border-t border-stone-800 bg-[#050505] flex flex-col gap-4">
								<div class="flex items-center justify-between mb-2">
									<div class="flex flex-col">
										<span class="text-[7px] text-stone-700 font-mono uppercase">Session_Token</span>
										<span class="text-[9px] text-stone-500 font-mono">0x{Math.random().toString(16).slice(2, 10).toUpperCase()}</span>
									</div>
									<div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
								</div>
								                        <Button 
															onclick={logout} 
															variant="danger"
															size="md"
															block={true}
															class="!py-4"
														>
															Terminate_Session
														</Button>							</div>
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
					class="md:hidden h-16 bg-[var(--header-bg)] backdrop-blur-xl border-t-2 border-stone-800 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 safe-area-pb"
				>
					<a
						href="/dashboard"
						class="flex flex-col items-center justify-center w-full h-full gap-1 {isRouteActive(
							'/dashboard'
						) || isRouteActive('/')
							? 'text-rust-light'
							: 'text-stone-600'} transition-colors"
					>
						<Icon name="gauge" size="1.25rem" />
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
						<Icon name="activity" size="1.25rem" />
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
						<Icon name="sliders" size="1.25rem" />
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
						<Icon name="shield" size="1.25rem" />
						<span class="font-mono text-[8px] font-black uppercase">SHLD</span>
					</a>
				</nav>
			</div>
		</div>
	</div>

		<ServerStatus 
			status={$isConnected ? 'ONLINE' : 'OFFLINE'} 
			players={$stats.active_nodes * 10} 
			servers={$stats.active_nodes} 
		/>

		<Notifications />
		
		<CommandPalette bind:isOpen={isCommandPaletteOpen} />

		<ShortcutHelpModal bind:isOpen={isShortcutHelpOpen} />

		<InstanceManagerModal
			bind:isOpen={sysState.console.isOpen}
			nodeId={sysState.console.nodeId}
			instanceId={sysState.console.instanceId}
			onClose={() => (sysState.console.isOpen = false)}
		/>
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
