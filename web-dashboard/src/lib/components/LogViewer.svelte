<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { onMount, tick, createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Button from './Button.svelte';
	import { formatBytes } from '$lib/utils';
	import IconComponent from '$lib/components/theme/Icon.svelte';
	import { BarChart3, AlertTriangle, Info, XCircle } from 'lucide-svelte';

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
		raw: any;
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

	let stats = $state({
		all: 0,
		info: 0,
		warn: 0,
		error: 0
	});

	let isAutoRefreshing = $state(false);
	let shouldAutoScroll = $state(true);
	const refreshInterval = 5000;
	let refreshTimer: ReturnType<typeof setInterval> | null = null;

	let logContainer = $state<HTMLElement | null>(null);

	let isConfirmOpen = $state(false);

	async function handleClearLogs() {
		try {
			await apiFetch(`/api/nodes/${nodeId}/logs`, { method: 'DELETE' });
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
		{ id: 'all', label: 'All', icon: BarChart3, color: 'text-slate-500' },
		{ id: 'info', label: 'Info', icon: Info, color: 'text-sky-400' },
		{ id: 'warn', label: 'Warn', icon: AlertTriangle, color: 'text-amber-400' },
		{ id: 'error', label: 'Error', icon: XCircle, color: 'text-rose-400' }
	];

	function parseLogLine(line: string, index: number): ParsedLogEntry {
		try {
			const json = JSON.parse(line);
			const date = new Date(json.time);

			let level: LogLevel = 'INF';
			const l = (json.level || '').toUpperCase();
			if (l === 'DEBUG') level = 'DBG';
			else if (l === 'INFO') level = 'INF';
			else if (l === 'WARN') level = 'WRN';
			else if (l === 'ERROR') level = 'ERR';
			else if (l === 'FATAL') level = 'FTL';
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

		if (selectedTab !== 'all') {
			out = out.filter((l) => {
				if (selectedTab === 'info') return ['INF', 'DBG'].includes(l.level);
				if (selectedTab === 'warn') return l.level === 'WRN';
				return ['ERR', 'FTL', 'PANIC'].includes(l.level);
			});
		}

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
				return 'text-sky-400';
			case 'WRN':
				return 'text-amber-400';
			case 'ERR':
				return 'text-rose-500 font-bold';
			case 'FTL':
				return 'text-purple-500 font-bold';
			case 'PANIC':
				return 'text-purple-600 font-bold bg-purple-950/30';
			default:
				return 'text-slate-500';
		}
	}

	async function fetchLogs() {
		if (!isOpen) return;
		loading = parsedLogs.length === 0;

		try {
			const r = await apiFetch(`/api/nodes/${nodeId}/logs`);
			if (!r.ok) throw new Error('Failed to fetch logs');

			const j = await r.json();
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
		shouldAutoScroll = scrollHeight - scrollTop - clientHeight < 50;
	}

	$effect(() => {
		if (filteredLogs.length > 0 && shouldAutoScroll && logContainer) {
			queueMicrotask(() => {
				if (logContainer) {
					logContainer.scrollTop = logContainer.scrollHeight;
				}
			});
		}
	});

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

	$effect(() => {
		if (isOpen && parsedLogs.length === 0) {
			fetchLogs();
		}
	});
</script>

{#if isOpen}
	{#if !embedded}
		<div
			class="fixed inset-0 z-[150] flex items-center justify-center sm:p-4 bg-slate-950/80 backdrop-blur-md"
			transition:fade={{ duration: 200 }}
		>
			<div
				class="absolute inset-0"
				onclick={onClose}
				onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && onClose()}
				role="button"
				tabindex="0"
				aria-label="Close"
			></div>

			<div
				class="relative w-full h-full sm:h-[90vh] sm:max-w-7xl bg-slate-900/90 border border-white/10 shadow-2xl flex flex-col overflow-hidden rounded-3xl backdrop-blur-2xl"
				transition:scale={{ start: 0.98, duration: 300, easing: cubicOut }}
			>
				<div class="contents">
					{@render content()}
				</div>
			</div>
		</div>
	{:else}
		<div
			class="h-full flex flex-col bg-slate-950/20 overflow-hidden border border-white/5 relative rounded-2xl"
		>
			{@render content()}
		</div>
	{/if}
{/if}

{#snippet content()}
	<div class="px-6 py-5 border-b border-white/5 flex justify-between items-center bg-black/20">
		<div class="flex items-center gap-5 overflow-hidden">
			<div class="p-2.5 bg-sky-500/10 border border-sky-500/20 rounded-xl shadow-lg">
				<IconComponent name="ph:activity-bold" size="1.25rem" class="text-sky-400" />
			</div>
			<div class="flex flex-col">
				<h2 class="text-white font-bold text-xl tracking-tight leading-none uppercase italic font-heading">
					Node Logs : <span class="text-sky-400">ID {nodeId}</span>
				</h2>
				<div class="flex items-center gap-4 mt-1.5">
					<div class="flex items-center gap-2">
						<div
							class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.5)]"
						></div>
						<span class="text-[10px] font-bold tracking-wider uppercase text-slate-500 font-sans"
							>Live Stream</span
						>
					</div>
					{#if fileSize > 0}
						<div class="w-px h-3 bg-white/10"></div>
						<span class="text-[10px] font-bold uppercase tracking-wider text-slate-500 font-sans"
							>{formatBytes(fileSize)} Buffer</span
						>
					{/if}
				</div>
			</div>
			{#if loading && parsedLogs.length > 0}
				<IconComponent
					name="ph:arrows-clockwise-bold"
					size="1rem"
					class="text-sky-400 animate-spin ml-2"
				/>
			{/if}
		</div>

		<div class="flex items-center gap-4">
			<div class="hidden sm:flex items-center gap-6 mr-6 text-[10px] font-bold text-slate-500">
				<label class="flex items-center gap-3 cursor-pointer group">
					<input type="checkbox" bind:checked={shouldAutoScroll} class="sr-only peer" />
					<div
						class="w-3.5 h-3.5 bg-black border border-white/10 rounded peer-checked:bg-sky-500 peer-checked:border-sky-500 transition-all shadow-inner"
					></div>
					<span class="group-hover:text-slate-300 transition-colors uppercase tracking-widest font-sans"
						>Auto Scroll</span
					>
				</label>
				<label class="flex items-center gap-3 cursor-pointer group">
					<input type="checkbox" bind:checked={isAutoRefreshing} class="sr-only peer" />
					<div
						class="w-3.5 h-3.5 bg-black border border-white/10 rounded peer-checked:bg-sky-500 peer-checked:border-sky-500 transition-all shadow-inner"
					></div>
					<span class="group-hover:text-slate-300 transition-colors uppercase tracking-widest font-sans"
						>Live Sync</span
					>
				</label>
			</div>

			<Button
				onclick={fetchLogs}
				variant="secondary"
				size="sm"
				icon="ph:arrows-clockwise-bold"
				{loading}
				class="!rounded-xl"
			/>

			<Button
				onclick={() => (isConfirmOpen = true)}
				variant="secondary"
				size="sm"
				icon="ph:trash-bold"
				class="!text-slate-500 hover:!text-rose-500 !rounded-xl"
			/>

			{#if !embedded}
				<div class="h-10 w-px bg-white/10 mx-2"></div>
				<Button onclick={onClose} variant="danger" size="md" icon="ph:x-bold" class="!rounded-xl" />
			{/if}
		</div>
	</div>

	<div
		class="px-6 py-5 bg-black/40 border-b border-white/5 flex flex-col md:flex-row gap-8 md:items-center"
	>
		<div class="flex gap-1 bg-black p-1 border border-white/5 rounded-xl shadow-inner">
			{#each [{ id: 'all', label: 'All', iconName: 'ph:chart-bar-bold' }, { id: 'info', label: 'Info', iconName: 'ph:info-bold' }, { id: 'warn', label: 'Warn', iconName: 'ph:warning-bold' }, { id: 'error', label: 'Error', iconName: 'ph:x-circle-bold' }] as tab}
				<Button
					onclick={() => (selectedTab = tab.id as TabId)}
					variant={selectedTab === tab.id ? 'primary' : 'ghost'}
					size="xs"
					icon={tab.iconName}
					class="!rounded-lg px-4"
				>
					<span class="hidden sm:inline uppercase font-bold tracking-tight text-[10px]">{tab.label}</span>
					<span class="ml-1 opacity-40 text-[9px]">[{stats[tab.id as TabId]}]</span>
				</Button>
			{/each}
		</div>

		<div class="flex-1 relative group">
			<IconComponent
				name="ph:magnifying-glass-bold"
				size="1rem"
				class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-sky-400 transition-colors"
			/>
			<input
				type="text"
				placeholder="Search log messages..."
				bind:value={searchTerm}
				class="w-full pl-12 pr-4 py-3 bg-black border border-white/10 focus:border-sky-500/50 rounded-xl text-white text-xs font-medium outline-none transition-all placeholder:text-slate-800 shadow-inner"
			/>
		</div>
	</div>

	<div class="flex-1 relative bg-black min-h-0">
		{#if loading && parsedLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center gap-6 text-slate-500">
				<div
					class="w-16 h-16 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"
				></div>
				<span class="text-xs font-bold uppercase tracking-widest animate-pulse text-sky-500 font-sans"
					>Connecting to log stream...</span
				>
			</div>
		{:else if error}
			<div
				class="absolute inset-0 flex flex-col items-center justify-center text-rose-500 gap-8 p-10 text-center"
			>
				<div class="p-6 bg-rose-500/5 border border-rose-500/20 rounded-3xl shadow-2xl">
					<IconComponent name="ph:warning-bold" size="4rem" class="opacity-80" />
				</div>
				<div class="space-y-3">
					<span class="text-lg font-bold uppercase tracking-widest block font-heading italic">Connection Failed</span>
					<p class="text-xs font-medium opacity-60 max-w-md mx-auto leading-relaxed font-sans">{error}</p>
				</div>
				<Button
					onclick={fetchLogs}
					variant="danger"
					size="lg"
					icon="ph:arrows-clockwise-bold"
					class="!rounded-xl">Retry Connection</Button
				>
			</div>
		{:else if filteredLogs.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center gap-6 text-slate-600">
				<div
					class="p-8 border-2 border-dashed border-white/5 rounded-3xl opacity-20 bg-white/[0.02]"
				>
					<IconComponent name="ph:magnifying-glass-bold" size="4rem" />
				</div>
				<span class="text-xs font-bold uppercase tracking-widest font-sans"
					>No logs found matching criteria</span
				>
			</div>
		{:else}
			<div
				bind:this={logContainer}
				onscroll={handleScroll}
				class="absolute inset-0 overflow-y-auto overflow-x-auto p-6 font-mono text-[11px] space-y-1 custom-scrollbar bg-black"
			>
				{#each filteredLogs as l (l.id)}
					<div
						class="flex items-start gap-6 hover:bg-white/[0.02] px-4 py-1.5 border-l-2 border-transparent hover:border-sky-500/50 transition-all select-text group relative rounded-r-md"
					>
						<span
							class="shrink-0 w-28 tabular-nums select-none opacity-40 text-slate-500 font-medium"
							>{l.time}</span
						>

						<span
							class="shrink-0 w-14 font-bold select-none text-[10px] {getLevelClass(
								l.level
							)} uppercase tracking-tighter text-center border border-current/20 rounded px-1 py-0.5 leading-none mt-0.5 bg-black/40"
						>
							{l.level}
						</span>

						<div
							class="flex-1 min-w-0 break-all sm:break-words text-slate-300 leading-relaxed font-medium"
						>
							<span class="font-sans">{l.message}</span>

							{#if l.raw && Object.keys(l.raw).length > 3}
								<div
									class="mt-2 ml-4 space-y-1.5 border-l border-white/5 bg-black/40 p-4 opacity-60 group-hover:opacity-100 transition-opacity rounded-xl shadow-inner"
								>
									{#each Object.entries(l.raw) as [k, v]}
										{#if !['time', 'level', 'msg', 'message'].includes(k)}
											<div class="flex gap-4 flex-wrap">
												<span class="text-sky-500/50 font-bold uppercase text-[9px] tracking-widest font-sans"
													>{k}:</span
												>
												<span
													class="text-slate-400 font-bold text-[10px] whitespace-pre-wrap break-all font-mono"
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
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
		height: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1e293b;
		border-radius: 99px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #334155;
	}
</style>
