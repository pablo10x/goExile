<script lang="ts">
	import { apiFetch, notify, API_BASE } from '$lib/api';
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
	import ConnectionModal from '$lib/components/ConnectionModal.svelte';
	import AppBackground from '$lib/components/theme/AppBackground.svelte';
	import { isNative, checkConnection } from '$lib/api';

	let { children, data } = $props();
	let isChecking = $state(true);
	let isCommandPaletteOpen = $state(false);
	let isShortcutHelpOpen = $state(false);
	let isConnectionModalOpen = $state(false);
	let eventSource: EventSource | null = null;
	let reconnectTimer: any = null;
	let reconnectAttempts = 0;

	let isSidebarCollapsed = $state(false);
	let isMobileMenuOpen = $state(false);

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
					if (ev.key === 't') goto('/config');
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
		if (
			$isAuthenticated &&
			!isChecking &&
			page.url.pathname !== '/login' &&
			page.url.pathname !== '/login/2fa'
		) {
			loadAllSettings().then(() => {
				if ($nodes.length === 0) initialFetch();
			});
			if (!eventSource) connectSSE();
		}
	});

	function connectSSE() {
		if (typeof window === 'undefined') return;
		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}

		const token = localStorage.getItem('exile_session');
		const url = token ? `${API_BASE()}/events?token=${token}` : `${API_BASE()}/events`;

		try {
			eventSource = new EventSource(url, { withCredentials: true });

			eventSource.onopen = () => {
				isConnected.set(true);
				connectionStatus.set('Connected');
				reconnectAttempts = 0;
				if (reconnectTimer) clearTimeout(reconnectTimer);
			};

			eventSource.onerror = () => {
				isConnected.set(false);
				connectionStatus.set('Disconnected');
				eventSource?.close();
				eventSource = null;

				const timeout = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
				reconnectAttempts++;

				if (reconnectTimer) clearTimeout(reconnectTimer);
				reconnectTimer = setTimeout(() => {
					if (get(isAuthenticated)) connectSSE();
				}, timeout);
			};

			eventSource.onmessage = (event) => {
				try {
					const data = JSON.parse(event.data);
					if (data.type === 'stats') stats.set(data.payload);
					else if (data.type === 'nodes') {
						const list: any[] = Array.isArray(data.payload)
							? data.payload
							: Object.values(data.payload);
						list.sort((a, b) => a.id - b.id);
						nodes.set(list);
					}
				} catch (e) {
					console.error('SSE Error', e);
				}
			};
		} catch (e) {
			console.error('Failed to create EventSource', e);
		}
	}

	async function checkAuth() {
		if (page.url.pathname === '/login' || page.url.pathname === '/login/2fa') {
			isChecking = false;
			return;
		}
		
		try {
			const res = await apiFetch('/api/stats', { cache: 'no-store' });
			if (res.ok) {
				isAuthenticated.set(true);
			} else {
				throw new Error('Auth failed');
			}
		} catch (e) {
			isAuthenticated.set(false);
			if (page.url.pathname !== '/login') {
				goto('/login');
			}
		} finally {
			isChecking = false;
		}
	}

	async function initialFetch() {
		try {
			const promises = [
				apiFetch('/api/nodes', { cache: 'no-store' }),
				apiFetch('/api/versions', { cache: 'no-store' }),
				apiFetch('/api/stats', { cache: 'no-store' })
			];
			const results = await Promise.all(promises);
			if (results[0].ok) nodes.set(await results[0].json());
			if (results[1].ok) serverVersions.set(await results[1].json());
			if (results[2].ok) stats.set(await results[2].json());
		} catch (e) {
			console.error('Initial fetch failed', e);
		}
	}

	async function logout() {
		try {
			await apiFetch('/api/auth/logout', { method: 'POST' });
		} catch (e) {}
		isAuthenticated.set(false);
		localStorage.removeItem('exile_session');
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
		} catch (e) {
			console.error(e);
		}
	}

	onMount(() => {
		checkAuth();
	});
</script>

{#if isChecking}
	<div class="flex items-center justify-center min-h-screen bg-[#020617]">
		<div class="relative flex flex-col items-center gap-8">
			<div
				class="animate-spin h-14 w-14 border-4 border-sky-500/20 border-t-sky-500 rounded-full shadow-[0_0_20px_rgba(14,165,233,0.3)]"
			></div>
			<span class="text-slate-400 font-sans text-sm font-bold uppercase tracking-[0.3em] animate-pulse"
				>Initialising System...</span
			>
		</div>
	</div>
{:else}
	{#if $isAuthenticated && page.url.pathname !== '/login' && page.url.pathname !== '/login/2fa'}
		<div
			class="flex h-screen w-screen overflow-hidden bg-[#020617] text-slate-300 font-sans selection:bg-sky-500/30 selection:text-sky-200"
		>
			<!-- Main App Background -->
			<AppBackground />

			<!-- Desktop Sidebar -->
			<aside
				class="hidden md:flex relative transition-all duration-500 ease-[cubic-bezier(0.23,1,0.32,1)] bg-slate-900/40 border-r border-white/5 flex-col shrink-0 overflow-hidden z-[100] backdrop-blur-3xl {isSidebarCollapsed
					? 'w-24'
					: 'w-72'}"
			>
				<div class="relative z-10 flex flex-col h-full">
					<!-- Logo -->
					<div
						class="p-8 border-b border-white/5 flex items-center {isSidebarCollapsed
							? 'justify-center'
							: 'justify-between'}"
					>
						{#if !isSidebarCollapsed}
							<div class="flex items-center gap-4 animate-reveal">
								<div class="w-10 h-10 rounded-2xl bg-gradient-to-br from-sky-500 to-teal-500 flex items-center justify-center shadow-2xl shadow-sky-500/20">
									<span class="font-bold text-white text-lg">E</span>
								</div>
								<h1 class="text-xl font-bold text-white tracking-tighter uppercase italic font-heading">
									Exile <span class="text-sky-400">Admin</span>
								</h1>
							</div>
						{/if}
						<button
							onclick={toggleSidebar}
							class="p-2.5 rounded-xl text-slate-500 hover:text-sky-400 hover:bg-white/5 transition-all"
						>
							<Icon
								name={isSidebarCollapsed
									? 'ph:caret-double-right-bold'
									: 'ph:caret-double-left-bold'}
								size="1.2rem"
							/>
						</button>
					</div>

					<!-- Navigation -->
					<nav class="flex-1 space-y-2 overflow-y-auto no-scrollbar py-6">
						{#each [{ title: 'Overview', items: [{ href: '/dashboard', icon: 'gauge', label: 'Dashboard', sub: 'System Status' }, { href: '/performance', icon: 'activity', label: 'Performance', sub: 'Network Stats' }] }, { title: 'Infrastructure', items: [{ href: '/server', icon: 'cpu', label: 'Nodes', sub: 'Resource Hub' }, { href: '/users', icon: 'users', label: 'Users', sub: 'Accounts' }] }, { title: 'Resources', items: [{ href: '/database', icon: 'database', label: 'Database', sub: 'Schema' }, { href: '/notes', icon: 'file-text', label: 'Notes', sub: 'Memos' }] }, { title: 'Settings', items: [{ href: '/config', icon: 'sliders', label: 'Settings', sub: 'Global' }, { href: '/redeye', icon: 'shield', label: 'Firewall', sub: 'Security' }] }] as section}
							<div class="mb-6">
								{#if !isSidebarCollapsed}
									<span class="nav-section-title">{section.title}</span>
								{/if}
								{#each section.items as link}
									<a
										href={link.href}
										class="nav-link-premium group {isSidebarCollapsed ? 'justify-center !px-0 !mx-2' : ''}"
										class:active={isRouteActive(link.href)}
										title={isSidebarCollapsed ? link.label : ''}
									>
										<div class="nav-icon-wrapper">
											<Icon name={link.icon} size="1.4rem" />
										</div>
										{#if !isSidebarCollapsed}
											<div class="nav-label-container">
												<span class="nav-label-premium">{link.label}</span>
												<span class="nav-sublabel-premium">{link.sub}</span>
											</div>
										{/if}
									</a>
								{/each}
							</div>
						{/each}
					</nav>

					<!-- Bottom Actions -->
					<div class="mt-auto p-4 border-t border-white/5 bg-black/20 flex flex-col gap-2 rounded-t-[2rem]">
						<button
							onclick={() => lowPowerMode.update((v) => !v)}
							class="nav-link-premium group {isSidebarCollapsed ? 'justify-center !px-0 !mx-2' : ''} {$lowPowerMode
								? 'bg-amber-500/10 border border-amber-500/20 text-amber-500'
								: 'bg-transparent border border-transparent'}"
						>
							<div class="nav-icon-wrapper {$lowPowerMode ? 'border-amber-500/30' : ''}">
								<Zap class="w-5 h-5 {$lowPowerMode ? 'text-amber-500' : 'text-slate-500 opacity-60'}" />
							</div>
							{#if !isSidebarCollapsed}
								<div class="nav-label-container">
									<span class="nav-label-premium">{$lowPowerMode ? 'Eco Mode' : 'Performance'}</span>
									<span class="nav-sublabel-premium">System Optimization</span>
								</div>
							{/if}
						</button>
						<button
							onclick={logout}
							class="nav-link-premium group {isSidebarCollapsed ? 'justify-center !px-0 !mx-2' : ''} hover:bg-rose-500/10 group/logout"
						>
							<div class="nav-icon-wrapper group-hover/logout:border-rose-500/30 group-hover/logout:text-rose-400">
								<Icon name="ph:power-bold" size="1.3rem" />
							</div>
							{#if !isSidebarCollapsed}
								<div class="nav-label-container">
									<span class="nav-label-premium group-hover/logout:text-rose-400">Terminate Session</span>
									<span class="nav-sublabel-premium">Secure Sign Out</span>
								</div>
							{/if}
						</button>
					</div>
				</div>
			</aside>

			<!-- Main Content Area -->
			<div class="flex-1 flex flex-col min-w-0 h-full relative overflow-hidden">
				<!-- Top Bar -->
				<header
					class="h-16 bg-slate-950/40 border-b border-white/5 z-50 flex items-center px-8 backdrop-blur-2xl shrink-0"
				>
					<div class="flex items-center gap-8 whitespace-nowrap w-full">
						<div class="flex items-center gap-3 shrink-0">
							<div
								class="w-2 h-2 rounded-full {$isConnected
									? 'bg-emerald-500 shadow-[0_0_12px_rgba(16,185,129,0.6)]'
									: 'bg-rose-500 shadow-[0_0_12px_rgba(244,63,94,0.6)]'}"
							></div>
							<span class="font-bold text-[10px] uppercase tracking-widest text-white/90"
								>System {$isConnected ? 'Connected' : 'Offline'}</span
							>
						</div>
						<div class="flex items-center gap-8 text-slate-500 text-[10px] font-bold uppercase tracking-wider">
							<button
								onclick={() => (isConnectionModalOpen = true)}
								class="group flex items-center gap-2 hover:text-sky-400 transition-all cursor-pointer"
							>
								<span class="hidden sm:inline"
									>Endpoint: <span class="text-slate-300 group-hover:text-sky-400 transition-colors font-mono"
										>{API_BASE() || 'Default'}</span
									></span
								>
								<Sliders class="w-3 h-3 opacity-50 group-hover:opacity-100" />
							</button>
							<span>Nodes: <span class="text-slate-200">{$stats.active_nodes} Active</span></span>
						</div>
						<div class="ml-auto flex items-center gap-6 shrink-0">
							<button
								onclick={() => (isShortcutHelpOpen = true)}
								class="flex items-center gap-2 text-[10px] font-bold text-slate-400 hover:text-white transition-all uppercase tracking-widest"
							>
								<Icon name="ph:keyboard-bold" size="1.1rem" />
								<span class="hidden md:inline">Shortcuts</span>
							</button>
						</div>
					</div>
				</header>

				<!-- Page Content -->
				<main class="flex-1 overflow-auto relative custom-scrollbar p-6 sm:p-10 md:p-14">
					{@render children()}
				</main>

				<!-- Mobile Bottom Nav -->
				<nav
					class="md:hidden h-16 bg-slate-950/80 backdrop-blur-2xl border-t border-white/5 fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around px-2 pb-safe shadow-[0_-10px_40px_rgba(0,0,0,0.5)]"
				>
					{#each [['/dashboard', 'gauge'], ['/performance', 'activity'], ['/config', 'sliders'], ['/redeye', 'shield']] as [href, icon]}
						<a
							{href}
							class="flex flex-col items-center justify-center w-full h-full {isRouteActive(href)
								? 'text-sky-400'
								: 'text-slate-500'} transition-all"
						>
							<Icon name={icon} size="1.5rem" />
						</a>
					{/each}
				</nav>
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
		<ConnectionModal bind:isOpen={isConnectionModalOpen} />
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
	:global(.pb-safe) {
		padding-bottom: env(safe-area-inset-bottom);
	}
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(14, 165, 233, 0.2);
	}
</style>
