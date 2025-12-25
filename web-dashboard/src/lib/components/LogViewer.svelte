<script lang="ts">
	import { onMount, tick, createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { siteSettings } from '$lib/stores';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { formatBytes } from '$lib/utils';
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
		spawnerId,
		isOpen = false,
		onClose = () => {},
		embedded = false
	} = $props<{
		spawnerId: number;
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
	const filteredLogs = $derived(filterLogs());
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
			await fetch(`/api/spawners/${spawnerId}/logs`, { method: 'DELETE' });
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
			const r = await fetch(`/api/spawners/${spawnerId}/logs`);
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
				class="relative w-full h-full sm:h-[90vh] sm:max-w-7xl bg-stone-950 border-2 border-stone-800 shadow-[0_0_50px_rgba(0,0,0,0.5)] flex flex-col overflow-hidden"
				transition:scale={{ start: 0.98, duration: 300, easing: cubicOut }}
			>
				<!-- CRT Overlay -->
				<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.03] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
				
				<div class="contents">
					{@render content()}
				</div>
			</div>
		</div>
	{:else}
		<div class="h-full flex flex-col bg-stone-950 overflow-hidden border-2 border-stone-800 relative">
			<!-- CRT Overlay -->
			<div class="absolute inset-0 pointer-events-none z-50 opacity-[0.02] bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] bg-[size:100%_4px,3px_100%]"></div>
			{@render content()}
		</div>
	{/if}
{/if}

{#snippet content()}
	<!-- Header -->
	<div
		class="px-6 py-4 border-b-2 border-stone-800 flex justify-between items-center bg-stone-900/50"
	>
		<div class="flex items-center gap-4 overflow-hidden">
			<div class="p-2 bg-rust/10 border border-rust/30">
				<Activity class="w-4 h-4 text-rust-light" />
			</div>
			<div class="flex flex-col">
				<h2 class="text-white font-black military-label text-lg tracking-tighter uppercase leading-none">
					LOG_ARCHIVE : <span class="text-rust-light">NODE_{spawnerId}</span>
				</h2>
				<div class="flex items-center gap-3 mt-1">
					<span class="text-[8px] font-mono text-stone-600 tracking-[0.2em] uppercase">Status: Streaming</span>
					{#if fileSize > 0}
						<span class="text-[8px] font-mono text-rust-light/60 bg-rust/5 px-1.5 py-0.5 border border-rust/20 uppercase"
							>{formatBytes(fileSize)} Buffer</span>
					{/if}
				</div>
			</div>
			{#if loading && parsedLogs.length > 0}
				<RefreshCw class="w-3 h-3 text-rust-light animate-spin" />
			{/if}
		</div>

		<div class="flex items-center gap-3">
			<div class="hidden sm:flex items-center gap-4 mr-4 text-[10px] font-mono text-stone-500">
				<label class="flex items-center gap-2 cursor-pointer group">
					<input
						type="checkbox"
						bind:checked={shouldAutoScroll}
						class="sr-only peer"
					/>
					<div class="w-3 h-3 border border-stone-700 peer-checked:bg-rust peer-checked:border-rust transition-all"></div>
					<span class="group-hover:text-stone-300 transition-colors uppercase tracking-widest">Auto_Scroll</span>
				</label>
				<label class="flex items-center gap-2 cursor-pointer group">
					<input
						type="checkbox"
						bind:checked={isAutoRefreshing}
						class="sr-only peer"
					/>
					<div class="w-3 h-3 border border-stone-700 peer-checked:bg-rust peer-checked:border-rust transition-all"></div>
					<span class="group-hover:text-stone-300 transition-colors uppercase tracking-widest">Live_Sync</span>
				</label>
			</div>

			<button
				class="p-2 bg-stone-900 border border-stone-800 text-stone-500 hover:text-white hover:border-rust transition-all"
				title="Manual_Refresh"
				onclick={fetchLogs}
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>

			<button
				class="p-2 bg-stone-900 border border-stone-800 text-stone-500 hover:text-red-500 hover:border-red-900/50 transition-all"
				title="Clear_Registry"
				onclick={() => (isConfirmOpen = true)}
			>
				<Trash2 class="w-4 h-4" />
			</button>

			{#if !embedded}
				<div class="h-8 w-px bg-stone-800 mx-1"></div>
				<button
					onclick={onClose}
					class="p-2 bg-red-950/10 border border-red-900/30 text-red-600 hover:bg-red-600 hover:text-black transition-all"
				>
					<X class="w-5 h-5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar: Tabs & Search -->
	<div
		class="px-6 py-4 bg-stone-950 border-b border-stone-800 flex flex-col md:flex-row gap-6 md:items-center"
	>
		<!-- Tabs -->
		<div class="flex gap-1 bg-black p-1 border border-stone-800">
			{#each tabs as tab}
				{@const Icon = tab.icon}
				<button
					class="px-4 py-2 text-[10px] font-black uppercase tracking-widest transition-all flex items-center gap-2
					                        {selectedTab === tab.id
						? 'bg-rust text-white shadow-[0_0_15px_rgba(120,53,15,0.3)]'
						: 'text-stone-600 hover:text-stone-300 hover:bg-stone-900'}"
					onclick={() => (selectedTab = tab.id)}
				>
					<Icon class="w-3.5 h-3.5" />
					<span class="hidden sm:inline">{tab.label}</span>
					<span class="ml-1 opacity-40 font-mono">[{stats[tab.id]}]</span>
				</button>
			{/each}
		</div>

		<!-- Search -->
		<div class="flex-1 relative group">
			<Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-stone-700 group-focus-within:text-rust transition-colors" />
			<input
				type="text"
				placeholder="FILTER_BUFFER_BY_IDENTIFIER..."
				bind:value={searchTerm}
				class="w-full pl-12 pr-4 py-2.5 bg-black border border-stone-800 focus:border-rust text-white text-[11px] font-mono uppercase tracking-widest outline-none transition-all placeholder:text-stone-800"
			/>
		</div>
	</div>

	<!-- Logs Area -->
	<div class="flex-1 relative bg-stone-950 min-h-0">
		{#if loading && parsedLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-stone-700 gap-4">
				<RefreshCw class="w-10 h-10 animate-spin opacity-20" />
				<span class="text-[10px] font-mono uppercase tracking-[0.5em] animate-pulse">Initializing_Buffer_Link...</span>
			</div>
		{:else if error}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-red-500 gap-4 p-8 text-center">
				<AlertTriangle class="w-12 h-12 opacity-40" />
				<div class="space-y-1">
					<span class="text-xs font-black uppercase tracking-widest block">FATAL_CONNECTION_FAULT</span>
					<p class="text-[10px] font-mono opacity-60 uppercase">{error}</p>
				</div>
				<button class="px-6 py-2 bg-red-950/20 border border-red-900 text-red-500 text-[10px] font-black uppercase tracking-widest hover:bg-red-600 hover:text-white transition-all" onclick={fetchLogs}>Retry_Protocol</button>
			</div>
		{:else if filteredLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-stone-800 gap-4">
				<Search class="w-12 h-12 opacity-10" />
				<span class="text-[10px] font-mono uppercase tracking-[0.3em]">NULL_RECORDS_LOCATED</span>
			</div>
		{:else}
			<div
				bind:this={logContainer}
				onscroll={handleScroll}
				class="absolute inset-0 overflow-y-auto overflow-x-auto p-4 font-mono text-[11px] space-y-1 custom-scrollbar bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.01)_1px,transparent_0)] bg-[size:40px_40px]"
			>
				{#each filteredLogs as l (l.id)}
					<div
						class="flex items-start gap-4 hover:bg-white/[0.03] px-3 py-1.5 border-l-2 border-transparent hover:border-rust/40 transition-all select-text group"
					>
						<!-- Time -->
						<span class="text-stone-600 shrink-0 w-24 tabular-nums select-none opacity-60 font-medium"
							>{l.time}</span
						>

						<!-- Level -->
						<span class="shrink-0 w-12 font-black select-none text-[10px] {getLevelClass(l.level)} uppercase tracking-tighter">
							[{l.level}]
						</span>

						<!-- Message -->
						<div class="flex-1 min-w-0 break-all sm:break-words text-stone-300 leading-relaxed font-medium">
							<span>{l.message}</span>

							<!-- Structured Context -->
							{#if l.raw && Object.keys(l.raw).length > 3}
								<div
									class="mt-2 ml-2 space-y-1 border-l-2 border-stone-800 bg-stone-900/30 p-3 opacity-70 group-hover:opacity-100 transition-opacity"
								>
									{#each Object.entries(l.raw) as [k, v]}
										{#if !['time', 'level', 'msg', 'message'].includes(k)}
											<div class="flex gap-3 flex-wrap">
												<span class="text-rust-light/60 font-bold uppercase text-[9px] tracking-widest">{k}:</span>
												<span
													class="text-stone-400 font-mono text-[10px] whitespace-pre-wrap break-all"
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
					<div class="h-8"></div>
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
		background: #050505;
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
