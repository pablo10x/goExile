<script lang="ts">
import { apiFetch, notify, API_BASE } from "$lib/api";
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
		Activity,
		Settings as SettingsIcon,
		Plus,
		Zap,
		Eye,
		ChevronRight,
		X,
		Gauge,
		FileText,
		Sliders,
		AlertCircle
	} from 'lucide-svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';

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
			if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
				e.preventDefault();
				isCommandPaletteOpen = !isCommandPaletteOpen;
			}
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
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

	$effect(() => {
		if (data && 'isAuthenticated' in data) {
			const auth = (data as any).isAuthenticated;
			if (auth !== get(isAuthenticated)) {
				isAuthenticated.set(auth);
			}
		}
	});

	$effect(() => {
		if ($isAuthenticated && !isChecking && page.url.pathname !== '/login' && page.url.pathname !== '/login/2fa') {
			loadAllSettings().then(() => {
				if ($nodes.length === 0) initialFetch();
			});
			if (!eventSource) connectSSE();
		}
	});

	$effect(() => {
		if (typeof window !== 'undefined') {
			if ($theme === 'dark') document.documentElement.classList.add('dark');
			else document.documentElement.classList.remove('dark');
			localStorage.setItem('theme', $theme);
		}
	});

	let sidebarLoaded = $state(false);
	let isSidebarCollapsed = $state(false);
	let isMobileMenuOpen = $state(false);

	function connectSSE() {
		if (typeof window === 'undefined') return;
		if (eventSource) eventSource.close();
		const token = localStorage.getItem('exile_session');
		const url = token ? `${API_BASE}/events?token=${token}` : `${API_BASE}/events`;
		eventSource = new EventSource(url, { withCredentials: true });
		eventSource.onopen = () => {
			isConnected.set(true);
			connectionStatus.set('Live');
		};
		eventSource.onerror = () => {
			isConnected.set(false);
			connectionStatus.set('Disconnected');
		};
		eventSource.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (data.type === 'stats') stats.set(data.payload);
				else if (data.type === 'nodes') {
					const list: any[] = Array.isArray(data.payload) ? data.payload : Object.values(data.payload);
					list.sort((a, b) => a.id - b.id);
					nodes.set(list);
				}
			} catch (e) { console.error('SSE Error', e); }
		};
	}

	async function checkAuth() {
		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') {
			isChecking = false;
			return;
		}
		try {
			const res = await apiFetch('/api/stats', { cache: 'no-store', credentials: 'include' });
			if (res.ok) {
				isAuthenticated.set(true);
				connectSSE();
			} else throw new Error('Auth failed');
		} catch (e) {
			isAuthenticated.set(false);
			if (window.location.pathname !== '/login') goto('/login');
		} finally { isChecking = false; }
	}

	async function initialFetch() {
		try {
			const promises = [
				apiFetch('/api/nodes', { cache: 'no-store', credentials: 'include' }),
				apiFetch('/api/versions', { cache: 'no-store', credentials: 'include' }),
				apiFetch('/api/stats', { cache: 'no-store', credentials: 'include' })
			];
			const results = await Promise.all(promises);
			if (results[0].ok) nodes.set(await results[0].json());
			if (results[1].ok) serverVersions.set(await results[1].json());
			if (results[2].ok) stats.set(await results[2].json());
		} catch (e) { console.error('Initial fetch failed', e); }
	}

	async function restartServer() {
		if (!confirm('Restart server and interrupt connections?')) return;
		try {
			restarting = true;
			await apiFetch('/api/restart', { method: 'POST' });
			setTimeout(() => window.location.reload(), 5000);
		} catch (e) { alert(e); restarting = false; }
	}

	onMount(() => {
		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') { if (get(isAuthenticated)) connectSSE(); }
			else if (eventSource) { eventSource.close(); eventSource = null; isConnected.set(false); }
		};
		document.addEventListener('visibilitychange', handleVisibilityChange);
		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') isChecking = false;
		else checkAuth().then(() => { if (get(isAuthenticated)) loadAllSettings().then(() => initialFetch()); });
		setTimeout(() => { sidebarLoaded = true; }, 300);
		return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
	});

	onDestroy(() => { if (eventSource) eventSource.close(); });

	async function logout() {
		await apiFetch('/api/auth/logout', { method: 'POST' });
		isAuthenticated.set(false);
		goto('/login');
	}

	function isRouteActive(path: string) { return page.url.pathname === path; }
	function toggleSidebar() { isSidebarCollapsed = !isSidebarCollapsed; }

	let showGlobalNoteModal = $state(false);
	async function handleGlobalSaveNote(note: Note) {
		const { id, ...noteWithoutId } = note;
		try {
			const res = await apiFetch('/api/notes', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(noteWithoutId)
			});
			if (res.ok) {
				const saved = await res.json();
				notes.update((n) => [saved, ...n]);
			}
		} catch (e) { console.error(e); }
	}
</script>

{#if isChecking}
	<div class="flex items-center justify-center min-h-screen bg-slate-950">
		<div class="relative flex flex-col items-center gap-8">
			<div class="animate-spin h-14 w-14 border-4 border-sky-500/20 border-t-sky-500 rounded-full"></div>
			<span class="text-slate-400 font-sans text-sm font-semibold tracking-wide animate-pulse">Initializing Interface...</span>
		</div>
	</div>
{:else}
	{#if $isAuthenticated && page.url.pathname !== '/login' && page.url.pathname !== '/login/2fa'}
		<div class="relative min-h-screen selection:bg-sky-500 selection:text-white bg-transparent font-sans">
			<!-- Professional Animated Gradient Mesh -->
			<div class="fixed inset-0 z-[-1] animate-gradient-mesh opacity-100"></div>

			<!-- System Status Bar -->
			<div class="fixed top-0 left-0 right-0 h-16 bg-slate-950/40 border-b border-white/5 z-[120] flex items-center px-8 backdrop-blur-2xl">
				<div class="flex items-center gap-8 whitespace-nowrap w-full">
					<div class="flex items-center gap-3 shrink-0">
						<div class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_12px_rgba(16,185,129,0.4)]"></div>
						<span class="font-bold text-xs text-slate-200">System Online</span>
					</div>
					<div class="flex items-center gap-8 text-slate-500 text-xs font-semibold">
						<span class="hidden sm:inline">Uplink: <span class="text-slate-300">{$connectionStatus}</span></span>
						<span>Nodes: <span class="text-slate-300">{$stats.active_nodes} Active</span></span>
					</div>
					<div class="ml-auto flex items-center gap-6 shrink-0">
						<button onclick={() => isShortcutHelpOpen = true} class="flex items-center gap-2 text-xs font-bold text-slate-400 hover:text-sky-400 transition-all">
							<Icon name="ph:question-bold" size="1.1rem" />
							<span class="hidden md:inline">Help</span>
						</button>
					</div>
				</div>
			</div>

			<div class="flex h-screen text-slate-300 overflow-hidden relative pt-16">
				<!-- Desktop Sidebar -->
				<aside class="hidden md:flex relative transition-all duration-500 ease-[cubic-bezier(0.23,1,0.32,1)] bg-slate-950/20 border-r border-white/5 flex-col shrink-0 overflow-hidden z-20 {isSidebarCollapsed ? 'w-24' : 'w-72'}">
					<div class="relative z-10 flex flex-col h-full">
						<div class="p-8 border-b border-white/5 flex items-center {isSidebarCollapsed ? 'justify-center' : 'justify-between'}">
							{#if !isSidebarCollapsed}
								<div class="flex items-center gap-4 animate-in fade-in slide-in-from-left-2 duration-500">
									<div class="w-10 h-10 rounded-2xl bg-gradient-to-br from-sky-500 to-sky-700 flex items-center justify-center shadow-2xl shadow-sky-500/20">
										<span class="font-bold text-white text-lg">E</span>
									</div>
									<h1 class="text-2xl font-bold text-white tracking-tight">Exile<span class="text-sky-400">OS</span></h1>
								</div>
							{/if}
							<button onclick={toggleSidebar} class="p-2.5 rounded-2xl text-slate-500 hover:text-sky-400 hover:bg-white/5 transition-all">
								<Icon name={isSidebarCollapsed ? 'ph:caret-double-right-bold' : 'ph:caret-double-left-bold'} size="1.2rem" />
							</button>
						</div>

						<nav class="flex-1 p-6 space-y-10 overflow-y-auto no-scrollbar py-10">
							<div class="space-y-10">
								{#each [
									{ title: 'Dashboard', items: [{ href: '/dashboard', icon: 'gauge', label: 'Overview', sub: 'Control Center' }, { href: '/performance', icon: 'activity', label: 'Analytics', sub: 'Real-time Metrics' }] },
									{ title: 'Fleet', items: [{ href: '/server', icon: 'cpu', label: 'Nodes', sub: 'Infrastructure' }, { href: '/users', icon: 'users', label: 'Users', sub: 'Access Control' }] },
									{ title: 'Resources', items: [{ href: '/database', icon: 'database', label: 'Database', sub: 'Global Storage' }, { href: '/notes', icon: 'file-text', label: 'Journal', sub: 'Internal Notes' }] },
									{ title: 'Protection', items: [{ href: '/config', icon: 'sliders', label: 'Settings', sub: 'Environment' }, { href: '/redeye', icon: 'shield', label: 'Security', sub: 'Sentinel Guard' }] }
								] as section}
									<div class="space-y-2">
										{#if !isSidebarCollapsed}<span class="text-xs font-bold text-slate-600 uppercase tracking-wider ml-4 mb-4 block">{section.title}</span>{/if}
										{#each section.items as link}
											<a href={link.href} class="nav-link-light {isSidebarCollapsed ? 'justify-center !px-0' : ''}" class:active={isRouteActive(link.href)} title={isSidebarCollapsed ? link.label : ''}>
												<div class="nav-icon-container-light"><Icon name={link.icon} size="1.3rem" /></div>
												{#if !isSidebarCollapsed}<div class="flex flex-col"><span class="nav-text-light text-base">{link.label}</span><span class="nav-subtext-light">{link.sub}</span></div>{/if}
											</a>
										{/each}
									</div>
								{/each}
							</div>
						</nav>

						<div class="mt-auto p-8 border-t border-white/5 bg-white/[0.03] flex flex-col gap-4 rounded-t-[3rem]">
							<div class="flex items-center gap-4 {isSidebarCollapsed ? 'flex-col' : ''}">
								<button onclick={() => lowPowerMode.update(v => !v)} class="p-4 rounded-2xl transition-all flex items-center justify-center gap-3 flex-1 {$lowPowerMode ? 'bg-amber-500/10 border border-amber-500/20 text-amber-500' : 'bg-white/5 border border-white/5 text-slate-400 hover:text-white hover:bg-white/10'}">
									<Zap class="w-5 h-5 {$lowPowerMode ? '' : 'opacity-60'}" />
									{#if !isSidebarCollapsed}<span class="text-xs font-bold tracking-wide">{$lowPowerMode ? 'Power Saving' : 'Max Perf'}</span>{/if}
								</button>
								<button onclick={logout} class="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400 hover:bg-rose-500 hover:text-white transition-all flex items-center justify-center group shadow-2xl shadow-rose-500/10">
									<Icon name="ph:power-bold" size="1.2rem" />
								</button>
							</div>
						</div>
					</div>
				</aside>

				<div class="flex-1 flex flex-col h-full overflow-hidden relative bg-transparent">
					<!-- Mobile Header -->
					<header class="md:hidden h-16 bg-slate-950/60 border-b border-white/5 flex items-center justify-between px-6 z-[130] shrink-0 backdrop-blur-2xl">
						<div class="flex items-center gap-4">
							<button onclick={() => isMobileMenuOpen = true} class="p-2 -ml-2 text-slate-400 hover:text-sky-400 transition-all"><Icon name="ph:list-bold" size="1.5rem" /></button>
							<div class="flex items-center gap-3">
								<div class="w-1.5 h-1.5 rounded-full bg-sky-500 shadow-[0_0_8px_rgba(14,165,233,0.5)]"></div>
								<h1 class="text-xl font-bold text-white tracking-tight">Exile<span class="text-sky-400">OS</span></h1>
							</div>
						</div>
						<span class="text-[10px] text-emerald-500 font-bold uppercase tracking-wider">Online</span>
					</header>

					<!-- Content -->
					<main class="flex-1 overflow-auto relative">
						<div class="w-full px-6 sm:px-10 py-10 md:py-14 min-h-full pb-32 md:pb-12">
							{@render children()}
						</div>
					</main>

					<!-- Mobile Nav -->
					<nav class="md:hidden h-16 bg-slate-950/80 backdrop-blur-2xl border-t border-white/5 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 pb-safe">
						{#each [['/dashboard', 'gauge'], ['/performance', 'activity'], ['/config', 'sliders'], ['/redeye', 'shield']] as [href, icon]}
							<a href={href} class="flex flex-col items-center justify-center w-full h-full {isRouteActive(href) ? 'text-sky-400' : 'text-slate-500'} transition-all">
								<Icon name={icon} size="1.5rem" />
							</a>
						{/each}
					</nav>
				</div>
			</div>
		</div>

		<ServerStatus status={$isConnected ? 'ONLINE' : 'OFFLINE'} players={$stats.active_nodes * 10} servers={$stats.active_nodes} />
		<Notifications />
		<CommandPalette bind:isOpen={isCommandPaletteOpen} />
		<ShortcutHelpModal bind:isOpen={isShortcutHelpOpen} />
		<InstanceManagerModal bind:isOpen={sysState.console.isOpen} nodeId={sysState.console.nodeId} instanceId={sysState.console.instanceId} onClose={() => (sysState.console.isOpen = false)} />
	{:else}
		{@render children()}
	{/if}
	<NoteModal bind:isOpen={showGlobalNoteModal} note={null} onSave={handleGlobalSaveNote} onClose={() => (showGlobalNoteModal = false)} />
{/if}

<style>
	:global(.pb-safe) { padding-bottom: env(safe-area-inset-bottom); }
</style>