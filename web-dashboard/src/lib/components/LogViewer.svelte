<script lang="ts">
	import { onMount, tick, createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { siteSettings } from '$lib/stores';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { formatBytes } from '$lib/utils';
	import IconComponent from '$lib/components/theme/Icon.svelte';
	import {
		Search,
		Download,
		Copy,
		RefreshCw,
		X,
		ChevronDown,
		ChevronUp,
		Clock,
		BarChart3,
		AlertTriangle,
		Info,
		Bug,
		XCircle,
		Activity,
		Trash2
	} from 'lucide-svelte';

	const {
		nodeId,
		isOpen = false,
		onClose = () => {},
		embedded = false
	} = $props<{
		nodeId: number;
		isOpen?: boolean;
		onClose?: () => void;
		embedded?: boolean;
	}>();

	type LogLevel = 'DBG' | 'INF' | 'WRN' | 'ERR' | 'FTL' | 'PANIC';

	interface ParsedLogEntry {
		id: number;
		time: string;
		level: LogLevel;
		message: string;
		raw: any; // The full parsed JSON object
		originalLine: string;
		timestamp: number;
	}

	type TabId = 'all' | 'info' | 'warn' | 'error';
	interface TabDef {
		id: TabId;
		label: string;
		icon: any;
		color: string;
	}

	const dispatch = createEventDispatcher();

	let logsRaw = $state('');
	let parsedLogs = $state<ParsedLogEntry[]>([]);
	const filteredLogs = $derived(filterLogs().slice(-200));
	let loading = $state(false);
	let error = $state('');
	let fileSize = $state(0);

	let selectedTab = $state<TabId>('all');
	let searchTerm = $state('');

	// Stats for tabs
	let stats = $state({
		all: 0,
		info: 0,
		warn: 0,
		error: 0
	});

	let isAutoRefreshing = $state(false); // New variable to control auto-refreshing
	let shouldAutoScroll = $state(true); // Initial value for auto-scroll
	const refreshInterval = 5000;
	let refreshTimer: ReturnType<typeof setInterval> | null = null;

	let logContainer = $state<HTMLElement | null>(null);

	let isConfirmOpen = $state(false);

	async function handleClearLogs() {
		try {
			await fetch(`/api/nodes/${nodeId}/logs`, { method: 'DELETE' });
			parsedLogs = [];
			fileSize = 0;
			updateStats();
		} catch (e) {
			console.error(e);
		}
		isConfirmOpen = false;
	}

	let confirmAction = handleClearLogs;

	const tabs: TabDef[] = [
		{ id: 'all', label: 'All', icon: BarChart3, color: 'text-slate-500 dark:text-slate-400' },
		{ id: 'info', label: 'Info', icon: Info, color: 'text-rust-light' },
		{ id: 'warn', label: 'Warn', icon: AlertTriangle, color: 'text-yellow-400' },
		{ id: 'error', label: 'Error', icon: XCircle, color: 'text-red-400' }
	];

	function parseLogLine(line: string, index: number): ParsedLogEntry {
		try {
			const json = JSON.parse(line);
			const date = new Date(json.time);

			// Map slog levels to our types if needed, though usually they match
			// slog default is DEBUG, INFO, WARN, ERROR
			let level: LogLevel = 'INF';
			const l = (json.level || '').toUpperCase();
			if (l === 'DEBUG') level = 'DBG';
			else if (l === 'INFO') level = 'INF';
			else if (l === 'WARN') level = 'WRN';
			else if (l === 'ERROR') level = 'ERR';
			else if (l === 'FATAL')
				level = 'FTL'; // Not standard slog but good to handle
			else if (l === 'PANIC') level = 'PANIC';

			return {
				id: index,
				time:
					date.toLocaleTimeString([], { hour12: false }) +
					'.' +
					date.getMilliseconds().toString().padStart(3, '0'),
				level: level,
				message: json.msg || json.message || '',
				raw: json,
				originalLine: line,
				timestamp: date.getTime()
			};
		} catch (e) {
			// Fallback for non-JSON lines
			return {
				id: index,
				time: '-',
				level: 'INF',
				message: line,
				raw: {},
				originalLine: line,
				timestamp: Date.now()
			};
		}
	}

	function updateStats() {
		const s = { all: 0, info: 0, warn: 0, error: 0 };
		parsedLogs.forEach((l) => {
			s.all++;
			if (l.level === 'WRN') s.warn++;
			else if (['ERR', 'FTL', 'PANIC'].includes(l.level)) s.error++;
			else s.info++;
		});
		stats = s;
	}

	function filterLogs() {
		let out = [...parsedLogs];

		// 1. Tab Filter
		if (selectedTab !== 'all') {
			out = out.filter((l) => {
				if (selectedTab === 'info') return ['INF', 'DBG'].includes(l.level);
				if (selectedTab === 'warn') return l.level === 'WRN';
				return ['ERR', 'FTL', 'PANIC'].includes(l.level);
			});
		}

		// 2. Search Filter
		if (searchTerm.trim()) {
			const lowerTerm = searchTerm.toLowerCase();
			out = out.filter(
				(l) =>
					l.message.toLowerCase().includes(lowerTerm) ||
					l.time.includes(lowerTerm) ||
					(l.raw.error && String(l.raw.error).toLowerCase().includes(lowerTerm))
			);
		}

		return out;
	}

	function getLevelClass(level: LogLevel) {
		switch (level) {
			case 'DBG':
				return 'text-slate-500';
			case 'INF':
				return 'text-rust-light';
			case 'WRN':
				return 'text-yellow-400';
			case 'ERR':
				return 'text-red-500 font-bold';
			case 'FTL':
				return 'text-purple-500 font-bold';
			case 'PANIC':
				return 'text-purple-600 font-bold bg-purple-950/30';
			default:
				return 'text-slate-500 dark:text-slate-400';
		}
	}

	async function fetchLogs() {
		if (!isOpen) return; // Don't fetch if closed
		loading = parsedLogs.length === 0; // Only show loading on initial fetch

		try {
			const r = await fetch(`/api/nodes/${nodeId}/logs`);
			if (!r.ok) throw new Error('Failed to fetch logs');

			const j = await r.json();
			// Handle case where logs might be empty or null
			logsRaw = j.logs || '';
			fileSize = j.size || 0;

			parsedLogs = logsRaw
				.split('\n')
				.filter((line) => line.trim().length > 0)
				.map(parseLogLine);

			updateStats();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function handleScroll() {
		if (!logContainer) return;
		const { scrollTop, scrollHeight, clientHeight } = logContainer;
		// Check if user is near bottom (within 50px)
		shouldAutoScroll = scrollHeight - scrollTop - clientHeight < 50;
	}

	// Auto-scroll logic: use $effect for side effects
	$effect(() => {
		if (filteredLogs.length > 0 && shouldAutoScroll && logContainer) {
			// Use a microtask to ensure DOM is updated before scrolling
			queueMicrotask(() => {
				if (logContainer) {
					// Additional null check for safety
					logContainer.scrollTop = logContainer.scrollHeight;
				}
			});
		}
	});

	// Auto Refresh Logic
	$effect(() => {
		if (isAutoRefreshing && isOpen) {
			if (!refreshTimer) {
				refreshTimer = setInterval(fetchLogs, refreshInterval);
			}
		} else {
			if (refreshTimer) {
				clearInterval(refreshTimer);
				refreshTimer = null;
			}
		}
	});

	onMount(() => {
		if (isOpen) fetchLogs();
		return () => {
			if (refreshTimer) clearInterval(refreshTimer);
		};
	});

	// Watch for open prop changes to trigger fetch
	$effect(() => {
		if (isOpen && parsedLogs.length === 0) {
			fetchLogs();
		}
	});
</script>

{#if isOpen}
	{#if !embedded}
		<div
			class="fixed inset-0 z-[150] flex items-center justify-center sm:p-4 bg-black/80 backdrop-blur-md"
			transition:fade={{ duration: 200 }}
		>
			<!-- Backdrop click to close -->
			<div
				class="absolute inset-0"
				onclick={onClose}
				onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && onClose()}
				role="button"
				tabindex="0"
				aria-label="Close"
			></div>

			<!-- Modal Container -->
			<div
				class="relative w-full h-full sm:h-[90vh] sm:max-w-7xl bg-[var(--terminal-bg)] border border-zinc-800 shadow-2xl flex flex-col overflow-hidden glass-panel industrial-frame"
				transition:scale={{ start: 0.98, duration: 300, easing: cubicOut }}
			>
				<!-- CRT Overlay -->
				<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.02] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
				
				<div class="contents">
					{@render content()}
				</div>
			</div>
		</div>
	{:else}
		<div class="h-full flex flex-col bg-[var(--terminal-bg)] overflow-hidden border border-zinc-800 relative glass-panel industrial-frame">
			<!-- CRT Overlay -->
			<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.01] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
			{@render content()}
		</div>
	{/if}
{/if}

{#snippet content()}
	<!-- Header -->
	<div
		class="px-6 py-5 border-b border-zinc-800 flex justify-between items-center bg-[var(--header-bg)]"
	>
		<div class="flex items-center gap-5 overflow-hidden">
			<div class="p-2.5 bg-rust/10 border border-rust/30 rounded-none industrial-frame shadow-lg">
				<IconComponent name="activity" size="1.25rem" class="text-rust-light" />
			</div>
			<div class="flex flex-col">
				<h2 class="text-white font-heading font-black text-xl tracking-tighter uppercase leading-none">
					LOG_ARCHIVE : <span class="text-rust">NODE_{nodeId}</span>
				</h2>
				<div class="flex items-center gap-4 mt-1.5">
					<div class="flex items-center gap-2">
						<div class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse shadow-emerald-500/50 shadow-lg"></div>
						<span class="font-jetbrains text-[9px] font-black tracking-[0.2em] uppercase" style="color: var(--text-dim)">STATUS: STREAMING</span>
					</div>
					{#if fileSize > 0}
						<div class="w-px h-3 bg-stone-800"></div>
						<span class="font-jetbrains text-[9px] font-black uppercase tracking-widest" style="color: var(--text-dim)"
							>{formatBytes(fileSize)} BUFFER_ACTIVE</span>
					{/if}
				</div>
			</div>
			{#if loading && parsedLogs.length > 0}
				<IconComponent name="ph:arrows-clockwise-bold" size="1rem" class="text-rust animate-spin ml-2" />
			{/if}
		</div>

		<div class="flex items-center gap-4">
			<div class="hidden sm:flex items-center gap-6 mr-6 font-jetbrains text-[9px] font-black" style="color: var(--text-dim)">
				<label class="flex items-center gap-3 cursor-pointer group">
					<input
						type="checkbox"
						bind:checked={shouldAutoScroll}
						class="sr-only peer"
					/>
					<div class="w-3.5 h-3.5 bg-stone-950 border border-zinc-800 peer-checked:bg-rust peer-checked:border-rust transition-all shadow-inner"></div>
					<span class="group-hover:text-stone-300 transition-colors uppercase tracking-widest">AUTO_SCROLL</span>
				</label>
				<label class="flex items-center gap-3 cursor-pointer group">
					<input
						type="checkbox"
						bind:checked={isAutoRefreshing}
						class="sr-only peer"
					/>
					<div class="w-3.5 h-3.5 bg-stone-950 border border-zinc-800 peer-checked:bg-rust peer-checked:border-rust transition-all shadow-inner"></div>
					<span class="group-hover:text-stone-300 transition-colors uppercase tracking-widest">LIVE_SYNC</span>
				</label>
			</div>

			<button
				class="p-2.5 bg-stone-900 border border-zinc-800 text-text-dim hover:text-rust hover:border-rust transition-all active:translate-y-px shadow-lg"
				title="Manual_Refresh"
				onclick={fetchLogs}
			>
				<IconComponent name="ph:arrows-clockwise-bold" size="1rem" class="{loading ? 'animate-spin' : ''}" />
			</button>

			<button
				class="p-2.5 bg-stone-900 border border-zinc-800 text-text-dim hover:text-red-500 hover:border-red-900/50 transition-all active:translate-y-px shadow-lg"
				title="Clear_Registry"
				onclick={() => (isConfirmOpen = true)}
			>
				<IconComponent name="ph:trash-bold" size="1rem" />
			</button>

			{#if !embedded}
				<div class="h-10 w-px bg-stone-800 mx-2"></div>
				<button
					onclick={onClose}
					class="p-2.5 bg-red-950/20 border border-red-900/30 text-red-600 hover:bg-red-600 hover:text-white transition-all active:scale-95 shadow-lg"
				>
					<IconComponent name="ph:x-bold" size="1.25rem" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar: Tabs & Search -->
	<div
		class="px-6 py-5 bg-[var(--header-bg)] border-b border-zinc-800 flex flex-col md:flex-row gap-8 md:items-center"
	>
		<!-- Tabs -->
		<div class="flex gap-1.5 bg-stone-950 p-1 border border-zinc-800 shadow-inner industrial-frame">
			{#each [
				{ id: 'all', label: 'All', iconName: 'ph:chart-bar-bold' },
				{ id: 'info', label: 'Info', iconName: 'ph:info-bold' },
				{ id: 'warn', label: 'Warn', iconName: 'ph:warning-bold' },
				{ id: 'error', label: 'Error', iconName: 'ph:x-circle-bold' }
			] as tab}
				<button
					class="px-5 py-2.5 font-jetbrains text-[9px] font-black uppercase tracking-[0.2em] transition-all flex items-center gap-3
					                        {selectedTab === tab.id
						? 'bg-rust text-white shadow-lg shadow-rust/20'
						: 'text-text-dim hover:text-stone-300 hover:bg-stone-900'}"
					onclick={() => (selectedTab = tab.id as TabId)}
				>
					<IconComponent name={tab.iconName} size="0.875rem" />
					<span class="hidden sm:inline">{tab.label}</span>
					<span class="ml-1 opacity-40">[{stats[tab.id as TabId]}]</span>
				</button>
			{/each}
		</div>

		<!-- Search -->
		<div class="flex-1 relative group">
			<IconComponent name="ph:magnifying-glass-bold" size="1rem" class="absolute left-4 top-1/2 -translate-y-1/2 text-stone-700 group-focus-within:text-rust transition-colors" />
			<input
				type="text"
				placeholder="FILTER_BUFFER_BY_IDENTIFIER..."
				bind:value={searchTerm}
				class="w-full pl-12 pr-4 py-3 bg-stone-950 border border-zinc-800 focus:border-rust text-white font-jetbrains text-[10px] font-bold uppercase tracking-[0.2em] outline-none transition-all placeholder:text-stone-800 shadow-inner"
			/>
		</div>
	</div>

	<!-- Logs Area -->
	<div class="flex-1 relative bg-[var(--terminal-bg)] min-h-0">
		<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
		
		{#if loading && parsedLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center gap-6" style="color: var(--text-dim)">
				<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
				<span class="font-heading font-black text-[11px] uppercase tracking-[0.5em] animate-pulse text-rust">Initializing_Buffer_Link...</span>
			</div>
		{:else if error}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-red-500 gap-8 p-10 text-center">
				<div class="p-6 bg-red-950/10 border border-red-900/30 industrial-frame shadow-2xl">
					<IconComponent name="alert" size="4rem" class="opacity-80 animate-pulse" />
				</div>
				<div class="space-y-3">
					<span class="font-heading font-black text-lg uppercase tracking-[0.3em] block">FATAL_CONNECTION_FAULT</span>
					<p class="font-jetbrains text-[11px] font-bold opacity-60 uppercase tracking-widest max-w-md mx-auto leading-relaxed">{error}</p>
				</div>
				<button class="px-12 py-4 bg-red-600 border border-red-400 text-white font-heading font-black text-[11px] uppercase tracking-widest hover:bg-red-500 transition-all shadow-lg active:translate-y-px shadow-red-900/30" onclick={fetchLogs}>Retry_Protocol</button>
			</div>
		{:else if filteredLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center gap-6" style="color: var(--text-dim)">
				<div class="p-8 bg-stone-900/40 border border-zinc-800 industrial-frame shadow-inner">
					<IconComponent name="ph:magnifying-glass-bold" size="4rem" class="opacity-20" />
				</div>
				<span class="font-jetbrains text-[11px] font-black uppercase tracking-[0.4em]">Null_Records_Located_In_Buffer</span>
			</div>
		{:else}
			<div
				bind:this={logContainer}
				onscroll={handleScroll}
				class="absolute inset-0 overflow-y-auto overflow-x-auto p-6 font-jetbrains text-[11px] space-y-1.5 custom-scrollbar"
			>
				{#each filteredLogs as l (l.id)}
					<div
						class="flex items-start gap-6 hover:bg-rust/5 px-4 py-2 border-l-2 border-transparent hover:border-rust/60 transition-all select-text group relative"
					>
						<!-- Time -->
						<span class="font-jetbrains font-black shrink-0 w-28 tabular-nums select-none opacity-50" style="color: var(--text-dim)"
							>{l.time}</span
						>

						<!-- Level -->
						<span class="shrink-0 w-14 font-black select-none text-[10px] {getLevelClass(l.level)} uppercase tracking-tighter">
							[{l.level}]
						</span>

						<!-- Message -->
						<div class="flex-1 min-w-0 break-all sm:break-words text-stone-300 leading-relaxed font-bold uppercase tracking-tight">
							<span>{l.message}</span>

							<!-- Structured Context -->
							{#if l.raw && Object.keys(l.raw).length > 3}
								<div
									class="mt-3 ml-4 space-y-2 border-l border-zinc-800 bg-stone-900/40 p-4 opacity-60 group-hover:opacity-100 transition-opacity industrial-frame shadow-inner"
								>
									{#each Object.entries(l.raw) as [k, v]}
										{#if !['time', 'level', 'msg', 'message'].includes(k)}
											<div class="flex gap-4 flex-wrap">
												<span class="text-rust-light/60 font-black uppercase text-[9px] tracking-widest">{k}:</span>
												<span
													class="text-stone-400 font-jetbrains font-bold text-[10px] whitespace-pre-wrap break-all"
													>{JSON.stringify(v)}</span
												>
											</div>
										{/if}
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/each}

				{#if shouldAutoScroll}
					<div class="h-12"></div>
				{/if}
			</div>
		{/if}
	</div>
{/snippet}

<ConfirmDialog
	bind:isOpen={isConfirmOpen}
	title="Clear Logs"
	message="Are you sure you want to clear all logs? This action cannot be undone."
	onConfirm={confirmAction}
/>

<style>
	/* Custom Scrollbar for industrial terminal */
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
		height: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: var(--terminal-bg);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #222;
		border: 1px solid #111;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}

	/* Hide scrollbar for tab nav */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
	.animate-flicker {
		animation: flicker 0.2s infinite;
	}
</style>
