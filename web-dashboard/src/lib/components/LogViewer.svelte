<script lang="ts">
	import { onMount, tick, createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
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
		XCircle
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

	let logContainer: HTMLElement; // No need for $state

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
		{ id: 'all', label: 'All', icon: BarChart3, color: 'text-slate-400' },
		{ id: 'info', label: 'Info', icon: Info, color: 'text-blue-400' },
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
				return 'text-blue-400';
			case 'WRN':
				return 'text-yellow-400';
			case 'ERR':
				return 'text-red-500 font-bold';
			case 'FTL':
				return 'text-purple-500 font-bold';
			case 'PANIC':
				return 'text-purple-600 font-bold bg-purple-950/30';
			default:
				return 'text-slate-400';
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
				logContainer.scrollTop = logContainer.scrollHeight;
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
			class="fixed inset-0 z-50 flex items-center justify-center sm:p-4 bg-black/60 backdrop-blur-sm"
			transition:fade={{ duration: 200 }}
		>
			<!-- Backdrop click to close -->
			<div
				class="absolute inset-0"
				on:click={onClose}
				on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && onClose()}
				role="button"
				tabindex="0"
				aria-label="Close"
			></div>

			<!-- Modal Container - Full screen on mobile, limited on desktop -->
			<div
				class="relative w-full h-full sm:h-[85vh] sm:max-w-6xl bg-slate-900 sm:rounded-xl border-0 sm:border border-slate-700 shadow-2xl flex flex-col overflow-hidden"
				transition:scale={{ start: 0.95, duration: 200, easing: cubicOut }}
			>
				<div class="contents">
					<!-- Shared Content -->
					{@render content()}
				</div>
			</div>
		</div>
	{:else}
		<div class="h-full flex flex-col bg-slate-900 overflow-hidden">
			{@render content()}
		</div>
	{/if}
{/if}

{#snippet content()}
	<!-- Header -->
	<div
		class="px-4 py-3 border-b border-slate-700 flex justify-between items-center bg-slate-800/50"
	>
		<div class="flex items-center gap-3 overflow-hidden">
			<h2 class="text-white font-semibold text-base sm:text-lg flex items-center gap-2 truncate">
				<span class="text-slate-400">#{spawnerId}</span> Logs
				{#if fileSize > 0}
					<span
						class="text-[10px] text-slate-500 font-mono border border-slate-700 rounded px-1.5 py-0.5 bg-slate-900/50 hidden sm:inline-block"
						>{formatBytes(fileSize)}</span
					>
				{/if}
			</h2>
			{#if loading && parsedLogs.length > 0}
				<RefreshCw class="w-3.5 h-3.5 text-slate-400 animate-spin shrink-0" />
			{/if}
		</div>

		<div class="flex items-center gap-1 sm:gap-2">
			<button
				class="p-2 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-colors"
				title="Refresh Now"
				on:click={fetchLogs}
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>

			<button
				class="p-2 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-colors"
				title="Clear Logs"
				on:click={() => (isConfirmOpen = true)}
			>
				<XCircle class="w-4 h-4" />
			</button>

			{#if !embedded}
				<div class="h-6 w-px bg-slate-700 mx-1 sm:mx-2"></div>

				<button
					on:click={onClose}
					class="p-2 hover:bg-red-500/20 hover:text-red-400 rounded-lg text-slate-400 transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Toolbar: Tabs & Search -->
	<div
		class="px-3 sm:px-4 py-3 bg-slate-900 border-b border-slate-800 flex flex-col md:flex-row gap-3 md:items-center"
	>
		<!-- Tabs (Scrollable on mobile) -->
		<div class="flex p-1 bg-slate-950 rounded-lg border border-slate-800 overflow-x-auto no-scrollbar">
			{#each tabs as tab}
					<button
						class="px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-all flex items-center gap-1.5 whitespace-nowrap
                        {selectedTab === tab.id
						? 'bg-slate-800 text-white shadow-sm'
						: 'text-slate-500 hover:text-slate-300'}"
						on:click={() => (selectedTab = tab.id)}
					>
						<svelte:component this={tab.icon} class="w-3.5 h-3.5" />
						{tab.label}
						<span class="ml-0.5 text-[10px] opacity-60 bg-slate-900/50 px-1.5 rounded-full">
							{stats[tab.id]}
						</span>
					</button>
			{/each}
		</div>

		<!-- Search & Options -->
		<div class="flex flex-col sm:flex-row sm:items-center gap-3 w-full md:w-auto">
			<div class="relative w-full sm:w-56">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
				<input
					type="text"
					placeholder="Search logs..."
					bind:value={searchTerm}
					class="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-slate-800 rounded-lg text-xs sm:text-sm text-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
				/>
			</div>

			<div class="flex items-center gap-4 text-xs text-slate-400 select-none">
				<label class="flex items-center gap-2 cursor-pointer">
					<input
						type="checkbox"
						bind:checked={shouldAutoScroll}
						class="rounded bg-slate-800 border-slate-700 text-blue-500 focus:ring-0 w-3.5 h-3.5"
					/>
					<span>Auto-scroll</span>
				</label>
				<label class="flex items-center gap-2 cursor-pointer">
					<input
						type="checkbox"
						bind:checked={isAutoRefreshing}
						class="rounded bg-slate-800 border-slate-700 text-blue-500 focus:ring-0 w-3.5 h-3.5"
					/>
					<span>Auto-refresh</span>
				</label>
			</div>
		</div>
	</div>

	<!-- Logs Area -->
	<div class="flex-1 relative bg-slate-950 min-h-0">
		{#if loading && parsedLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-slate-500 gap-3">
				<RefreshCw class="w-8 h-8 animate-spin opacity-50" />
				<span class="text-sm">Loading logs...</span>
			</div>
		{:else if error}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-red-400 gap-2">
				<AlertTriangle class="w-8 h-8 opacity-50" />
				<span class="text-sm">{error}</span>
				<button class="text-xs underline hover:text-red-300" on:click={fetchLogs}>Try Again</button>
			</div>
		{:else if filteredLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center text-slate-600 gap-2">
				<Search class="w-8 h-8 opacity-20" />
				<span class="text-sm">No logs found</span>
			</div>
		{:else}
			<div
				bind:this={logContainer}
				on:scroll={handleScroll}
				class="absolute inset-0 overflow-y-auto overflow-x-auto p-2 font-mono text-[10px] sm:text-xs space-y-0.5 custom-scrollbar"
			>
				{#each filteredLogs as l (l.id)}
					<div
						class="flex items-start gap-2 sm:gap-3 hover:bg-slate-900/50 px-1.5 sm:px-2 py-1 rounded select-text group transition-colors"
					>
						<!-- Time -->
						<span class="text-slate-600 shrink-0 w-16 sm:w-24 tabular-nums select-none truncate">{l.time}</span>

						<!-- Level -->
						<span class="shrink-0 w-8 sm:w-12 font-bold select-none {getLevelClass(l.level)}">
							{l.level}
						</span>

						<!-- Message -->
						<div class="flex-1 min-w-0 break-all sm:break-words text-slate-300">
							<span>{l.message}</span>

							<!-- Structured Context -->
							{#if l.raw && Object.keys(l.raw).length > 3}
								<div
									class="mt-1 ml-1 sm:ml-2 text-slate-500 text-[9px] sm:text-[10px] space-y-1 border-l-2 border-slate-800 pl-2 opacity-80 group-hover:opacity-100 transition-opacity"
								>
									{#each Object.entries(l.raw) as [k, v]}
										{#if !['time', 'level', 'msg', 'message'].includes(k)}
											<div class="flex gap-2 flex-wrap">
												<span class="text-slate-600">{k}:</span>
												<span class="text-slate-400 font-mono whitespace-pre-wrap break-all"
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
					<div class="h-4"></div>
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
	/* Custom Scrollbar for the log container */
	.custom-scrollbar::-webkit-scrollbar {
		width: 10px;
		height: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: #0f172a; /* slate-950 */
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #334155; /* slate-700 */
		border-radius: 5px;
		border: 2px solid #0f172a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569; /* slate-600 */
	}
	
	/* Hide scrollbar for tab nav */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none; /* IE and Edge */
		scrollbar-width: none; /* Firefox */
	}
</style>