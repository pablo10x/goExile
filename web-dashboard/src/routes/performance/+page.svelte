<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Zap,
		Activity,
		MemoryStick,
		HardDrive,
		Server,
		RefreshCw,
		AlertTriangle,
		CheckCircle,
		Timer,
		Trash2,
		Signal,
		Terminal,
		Globe,
		Database,
		Lock,
		Cpu,
		Radio,
		Dna,
		AlertOctagon
	} from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	// Types
	interface RuntimeMetrics {
		heap_alloc: number;
		heap_sys: number;
		heap_idle: number;
		heap_inuse: number;
		heap_released: number;
		heap_objects: number;
		stack_inuse: number;
		stack_sys: number;
		total_alloc: number;
		sys: number;
		mallocs: number;
		frees: number;
		live_objects: number;
		heap_alloc_rate: number;
		heap_usage_ratio: number;
		num_gc: number;
		num_forced_gc: number;
		gc_cpu_fraction: number;
		last_gc_pause_ns: number;
		avg_gc_pause_ns: number;
		max_gc_pause_ns: number;
		total_gc_pause_ns: number;
		next_gc_target: number;
		gc_trigger_ratio: number;
		num_goroutine: number;
		num_cpu: number;
		num_cgo_call: number;
		goroutine_growth: number;
		peak_goroutines: number;
		go_version: string;
		goos: string;
		goarch: string;
		uptime_ms: number;
	}

	interface SpawnerDetail {
		id: number;
		region: string;
		host: string;
		port: number;
		status: string;
		current_instances: number;
		max_instances: number;
		cpu_usage: number;
		mem_used: number;
		mem_total: number;
		mem_percent: number;
		disk_used: number;
		disk_total: number;
		disk_percent: number;
		game_version: string;
	}

	interface SpawnerMetrics {
		total_spawners: number;
		online_spawners: number;
		total_instances: number;
		running_instances: number;
		total_cpu_usage: number;
		avg_cpu_usage: number;
		total_mem_used: number;
		total_mem_total: number;
		total_disk_used: number;
		total_disk_total: number;
		mem_usage_percent: number;
		disk_usage_percent: number;
		spawner_details: SpawnerDetail[];
	}

	interface DatabaseMetrics {
		connected: boolean;
		open_connections: number;
		in_use: number;
		idle: number;
		wait_count: number;
		wait_duration_ms: number;
		size: string;
		commits: number;
		rollbacks: number;
		cache_hit_ratio: number;
	}

	interface NetworkMetrics {
		total_requests: number;
		total_errors: number;
		error_rate: number;
		bytes_sent: number;
		bytes_received: number;
		requests_per_second: number;
		active_connections: number;
	}

	interface CombinedMetrics {
		master: RuntimeMetrics;
		spawners: SpawnerMetrics;
		database: DatabaseMetrics;
		network: NetworkMetrics;
	}

	// State
	let metrics = $state<CombinedMetrics | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let lastUpdate = $state<Date | null>(null);
	let refreshInterval: number | null = null;
	let autoRefresh = $state(true);
	let refreshRate = $state(5000);
	let gcLoading = $state(false);
	let freeMemLoading = $state(false);

	// Previous values for trend calculation
	let prevMetrics = $state<CombinedMetrics | null>(null);

	// Helpers
	function formatBytes(bytes: number, decimals = 2): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
	}

	function formatDuration(ms: number): string {
		if (ms < 1000) return `${ms}ms`;
		if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
		if (ms < 3600000) return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
		const hours = Math.floor(ms / 3600000);
		const minutes = Math.floor((ms % 3600000) / 60000);
		return `${hours}h ${minutes}m`;
	}

	function formatNanoseconds(ns: number): string {
		if (ns < 1000) return `${ns}ns`;
		if (ns < 1000000) return `${(ns / 1000).toFixed(2)}µs`;
		if (ns < 1000000000) return `${(ns / 1000000).toFixed(2)}ms`;
		return `${(ns / 1000000000).toFixed(2)}s`;
	}

	function formatNumber(n: number): string {
		if (n >= 1000000000) return (n / 1000000000).toFixed(1) + 'B';
		if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
		if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
		return n.toString();
	}

	function getHealthStatus(metrics: CombinedMetrics): { status: string; color: string; icon: any } {
		const heapRatio = metrics.master.heap_usage_ratio;
		const errorRate = metrics.network.error_rate;
		const dbConnected = metrics.database.connected;

		if (!dbConnected || heapRatio > 0.9 || errorRate > 20) {
			return { status: 'FAULT', color: '#ef4444', icon: AlertOctagon };
		}
		if (heapRatio > 0.75 || errorRate > 10) {
			return { status: 'DEGRADED', color: '#f97316', icon: AlertTriangle };
		}
		return { status: 'OPTIMAL', color: '#10b981', icon: CheckCircle };
	}

	function getTrend(current: number, previous: number | undefined): 'up' | 'down' | 'stable' {
		if (previous === undefined) return 'stable';
		const diff = current - previous;
		const threshold = Math.abs(previous) * 0.02; // 2% change threshold
		if (diff > threshold) return 'up';
		if (diff < -threshold) return 'down';
		return 'stable';
	}

	// API calls
	async function fetchMetrics() {
		try {
			const res = await fetch('/api/metrics');
			if (!res.ok) throw new Error('COMMS_FAILURE');
			const data = await res.json();
			prevMetrics = metrics;
			metrics = data;
			lastUpdate = new Date();
			error = null;
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function forceGC() {
		gcLoading = true;
		try {
			await fetch('/api/metrics/gc', { method: 'POST' });
			await fetchMetrics();
		} catch (e: any) {
			error = e.message;
		} finally {
			gcLoading = false;
		}
	}

	async function freeMemory() {
		freeMemLoading = true;
		try {
			await fetch('/api/metrics/memory/free', { method: 'POST' });
			await fetchMetrics();
		} catch (e: any) {
			error = e.message;
		} finally {
			freeMemLoading = false;
		}
	}

	function startAutoRefresh() {
		if (refreshInterval) clearInterval(refreshInterval);
		if (autoRefresh) {
			refreshInterval = setInterval(fetchMetrics, refreshRate) as unknown as number;
		}
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		startAutoRefresh();
	}

	onMount(() => {
		fetchMetrics();
		startAutoRefresh();
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});

	// Reactive derived values
	let health = $derived(metrics ? getHealthStatus(metrics) : null);
	let heapTrend = $derived(
		metrics && prevMetrics
			? getTrend(metrics.master.heap_alloc, prevMetrics.master.heap_alloc)
			: 'stable'
	);
	let goroutineTrend = $derived(
		metrics && prevMetrics
			? getTrend(metrics.master.num_goroutine, prevMetrics.master.num_goroutine)
			: 'stable'
	);
</script>

<div class="w-full h-full flex flex-col relative font-jetbrains">
	<!-- Cinematic Overlays -->
	<div class="fixed inset-0 pointer-events-none z-[100] bg-vignette opacity-40"></div>
	
	<!-- Main Content Chassis -->
	<div class="w-full h-full flex flex-col gap-10 relative z-10 pb-32 md:pb-12">
		
		<!-- Intelligence Header -->
		<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-4 border-rust pl-10 py-2 bg-[#0a0a0a]/60 backdrop-blur-xl shadow-2xl relative overflow-hidden industrial-frame">
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
			
			<div class="space-y-4 p-2 relative z-10">
				<div class="flex items-center gap-4">
					<span class="bg-rust text-white px-3 py-1 text-[10px] font-black uppercase tracking-[0.2em] shadow-lg shadow-rust/20">REALTIME_TELEMETRY</span>
					<div class="w-px h-3 bg-stone-800"></div>
					<span class="font-jetbrains text-[9px] font-black text-stone-500 uppercase tracking-[0.4em] italic">STATION: EXILE_HIVE_CORE</span>
				</div>
				<h1 class="text-5xl sm:text-6xl lg:text-7xl font-heading font-black tracking-tighter text-white uppercase leading-none">
					SYSTEM_<span class="text-rust">METRICS</span>
				</h1>
				<div class="flex items-center gap-6 pt-2">
					<div class="flex items-center gap-3 font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest">
						<div class="w-2 h-2 bg-emerald-500 shadow-[0_0_10px_#10b981] animate-pulse"></div>
						LINK_STATUS: ACTIVE
					</div>
					<div class="w-px h-4 bg-stone-800"></div>
					{#if lastUpdate}
						<div class="font-jetbrains text-[10px] font-black text-stone-600 uppercase tracking-widest italic">
							SYNC_TIMESTAMP: {lastUpdate.toLocaleTimeString([], { hour12: false })}
						</div>
					{/if}
				</div>
			</div>

			<div class="flex flex-wrap items-center gap-6 p-4 relative z-10">
				<!-- Health Status Summary -->
				{#if health}
					<div class="flex items-center gap-4 bg-black/40 px-6 py-3 border border-stone-800 shadow-inner">
						<health.icon class="w-5 h-5" style="color: {health.color}" />
						<div class="flex flex-col">
							<span class="text-[8px] font-jetbrains font-black text-stone-600 uppercase tracking-widest">Global Integrity</span>
							<span class="text-sm font-heading font-black uppercase italic" style="color: {health.color}">{health.status}</span>
						</div>
					</div>
				{/if}

				<div class="flex bg-black/40 p-1.5 border border-stone-800">
					<button
						onclick={toggleAutoRefresh}
						class="flex flex-col items-start px-6 py-2 transition-all duration-500 relative group {autoRefresh ? 'bg-rust text-white shadow-xl shadow-rust/30' : 'text-stone-600 hover:text-stone-300'}"
					>
						<span class="font-jetbrains text-[8px] font-black tracking-[0.3em] uppercase mb-1 opacity-50">Stream</span>
						<span class="font-heading text-xs font-black tracking-widest uppercase">{autoRefresh ? 'LIVE' : 'PAUSED'}</span>
					</button>
					<button
						onclick={fetchMetrics}
						disabled={loading}
						class="px-4 hover:text-white transition-all text-stone-600 disabled:opacity-20"
					>
						<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
					</button>
				</div>
			</div>
		</div>

		<!-- Top Level Briefing Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 px-2">
			{#if metrics}
				{#each [
					{ label: 'System Uptime', val: formatDuration(metrics.master.uptime_ms), icon: Timer, color: 'text-white' },
					{ label: 'Routine Clusters', val: metrics.master.num_goroutine, icon: Activity, color: 'text-white', trend: goroutineTrend },
					{ label: 'Memory Pressure', val: formatBytes(metrics.master.heap_alloc), icon: MemoryStick, color: 'text-white', detail: `${(metrics.master.heap_usage_ratio * 100).toFixed(1)}% Usage` },
					{ label: 'Signal Velocity', val: `${metrics.network.requests_per_second?.toFixed(1)}Hz`, icon: Signal, color: 'text-white', detail: `${formatNumber(metrics.network.total_requests)} Total` }
				] as block}
					<div class="modern-industrial-card glass-panel group p-6 !rounded-none shadow-2xl flex flex-col justify-between min-h-[140px]">
						<div class="flex justify-between items-start">
							<span class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-[0.3em] italic">{block.label}</span>
							<block.icon class="w-4 h-4 text-stone-800 group-hover:text-rust transition-colors" />
						</div>
						<div class="flex items-end justify-between mt-4">
							<div class="text-4xl font-heading font-black italic tracking-tighter text-white">{block.val}</div>
							{#if block.trend && block.trend !== 'stable'}
								<div class="text-xl font-black {block.trend === 'up' ? 'text-red-500' : 'text-emerald-500'} animate-pulse">
									{block.trend === 'up' ? '▲' : '▼'}
								</div>
							{/if}
						</div>
						{#if block.detail}
							<div class="text-[8px] font-jetbrains font-black text-rust-light/60 uppercase tracking-widest mt-2">{block.detail}</div>
						{/if}
					</div>
				{/each}
			{/if}
		</div>

		<!-- Detailed Diagnostic Deck -->
		<div class="grid grid-cols-1 xl:grid-cols-12 gap-6 px-2 flex-1">
			<!-- Memory Matrix -->
			<div class="xl:col-span-8 flex flex-col gap-6">
				<div class="modern-industrial-card glass-panel !rounded-none flex-1 flex flex-col">
					<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center justify-between">
						<div class="flex items-center gap-4">
							<div class="p-2 bg-rust/5 border border-rust/20 industrial-frame">
								<MemoryStick class="w-5 h-5 text-rust-light" />
							</div>
							<h3 class="font-heading font-black text-base text-white uppercase tracking-widest">Memory_Core_Diagnostics</h3>
						</div>
						<div class="flex gap-2">
							<button onclick={forceGC} disabled={gcLoading} class="px-4 py-2 bg-stone-900 border border-stone-800 text-[9px] font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all">Flush</button>
							<button onclick={freeMemory} disabled={freeMemLoading} class="px-4 py-2 bg-stone-900 border border-stone-800 text-[9px] font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all">Vacuum</button>
						</div>
					</div>
					
					<div class="p-8 space-y-10 flex-1 overflow-y-auto custom-scrollbar">
						<!-- Visual Heap Load -->
						<div class="space-y-4">
							<div class="flex justify-between font-jetbrains text-[9px] font-black uppercase tracking-[0.2em] text-stone-600 italic">
								<span>Heap_Allocation_Graph</span>
								<span class="text-white">SYS_LIMIT: {formatBytes(metrics?.master.sys || 0)}</span>
							</div>
							<div class="h-12 bg-stone-950 border border-stone-800 p-1 relative overflow-hidden shadow-inner industrial-frame">
								<div
									class="h-full bg-gradient-to-r from-rust/20 via-rust/60 to-rust border-r-2 border-rust transition-all duration-1000 ease-out"
									style="width: {(metrics?.master.heap_usage_ratio || 0) * 100}%"
								></div>
								<div class="absolute inset-0 flex items-center justify-between px-6 font-heading font-black text-xs text-white mix-blend-difference uppercase tracking-widest">
									<span>UTILIZED: {formatBytes(metrics?.master.heap_inuse || 0)}</span>
									<span>{((metrics?.master.heap_usage_ratio || 0) * 100).toFixed(1)}%</span>
								</div>
							</div>
						</div>

						<!-- Diagnostic Grid -->
						<div class="grid grid-cols-2 md:grid-cols-3 gap-4">
							{#each [
								{ label: 'Heap Live', val: formatBytes(metrics?.master.heap_alloc || 0) },
								{ label: 'Stack Size', val: formatBytes(metrics?.master.stack_sys || 0) },
								{ label: 'Object Cnt', val: formatNumber(metrics?.master.live_objects || 0) },
								{ label: 'IO Velocity', val: `${formatBytes(metrics?.master.heap_alloc_rate || 0)}/s` },
								{ label: 'Idle Buffer', val: formatBytes(metrics?.master.heap_idle || 0) },
								{ label: 'Sync Target', val: formatBytes(metrics?.master.next_gc_target || 0) }
							] as item}
								<div class="bg-black/40 border border-stone-800 p-5 industrial-frame group hover:border-rust/30 transition-all">
									<div class="text-[8px] font-jetbrains font-black text-stone-700 tracking-widest uppercase mb-2 group-hover:text-rust transition-colors italic">{item.label}</div>
									<div class="text-xl font-heading font-black text-stone-200 uppercase tracking-tighter">{item.val}</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>

			<!-- Auxiliary Logs & Scrubber -->
			<div class="xl:col-span-4 flex flex-col gap-6">
				<div class="modern-industrial-card glass-panel !rounded-none flex-1 flex flex-col">
					<div class="p-6 border-b border-stone-800 bg-stone-950/40">
						<div class="flex items-center gap-4">
							<div class="p-2 bg-rust/5 border border-rust/20 industrial-frame">
								<Trash2 class="w-5 h-5 text-rust-light" />
							</div>
							<h3 class="font-heading font-black text-base text-white uppercase tracking-widest">Scrubber_Activity</h3>
						</div>
					</div>
					
					<div class="p-8 space-y-8 flex-1 overflow-y-auto custom-scrollbar">
						<div class="space-y-4">
							<div class="flex justify-between text-[9px] font-black text-stone-600 uppercase tracking-[0.2em] italic">
								<span>Interrupt Vector</span>
								<span class="text-rust">{((metrics?.master.gc_cpu_fraction || 0) * 100).toFixed(4)}%</span>
							</div>
							<div class="flex h-8 gap-1 bg-stone-950 border border-stone-800 p-1 industrial-frame">
								{#each Array(10) as _, i}
									{@const isFilled = i < Math.min((metrics?.master.gc_cpu_fraction || 0) * 500, 10)}
									<div class="flex-1 {isFilled ? 'bg-rust' : 'bg-stone-900/40'}"></div>
								{/each}
							</div>
						</div>

						<div class="space-y-4 bg-black/40 p-6 border border-stone-800 industrial-frame">
							{#each [
								{ label: 'Total Cycles', val: metrics?.master.num_gc || 0 },
								{ label: 'Avg Latency', val: formatNanoseconds(metrics?.master.avg_gc_pause_ns || 0) },
								{ label: 'Peak Jitter', val: formatNanoseconds(metrics?.master.max_gc_pause_ns || 0), color: 'text-red-500' }
							] as row}
								<div class="flex justify-between border-b border-stone-800/50 pb-3 last:border-none last:pb-0">
									<span class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-widest italic">{row.label}</span>
									<span class="text-xs font-heading font-black {row.color || 'text-stone-300'}">{row.val}</span>
								</div>
							{/each}
						</div>

						<div class="p-6 bg-stone-950/60 border border-stone-800 industrial-frame italic text-[10px] font-jetbrains font-bold uppercase leading-relaxed text-stone-500">
							<div class="text-rust mb-2 font-black tracking-widest">[AUTO_AUDIT]</div>
							Heuristic engine ensuring memory stability. All reclamation processes logged to persistent buffer.
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Fleet Status Footer -->
		<div class="border-t border-stone-800 bg-[#0a0a0a]/60 p-8 flex flex-wrap justify-center gap-12 font-jetbrains text-[10px] font-black tracking-[0.4em] uppercase text-stone-700 italic industrial-frame mx-2 shadow-2xl">
			<div class="flex items-center gap-4 group cursor-default">
				<Cpu class="w-5 h-5 text-rust opacity-40 group-hover:opacity-100 group-hover:scale-110 transition-all" />
				<span class="group-hover:text-stone-400">Nodes: {metrics?.spawners.online_spawners || 0} Online</span>
			</div>
			<div class="flex items-center gap-4 group cursor-default">
				<Database class="w-5 h-5 text-rust opacity-40 group-hover:opacity-100 group-hover:scale-110 transition-all" />
				<span class="group-hover:text-stone-400">Storage: {metrics?.database.size || 'N/A'}</span>
			</div>
			<div class="flex items-center gap-4 group cursor-default">
				<Signal class="w-5 h-5 text-rust opacity-40 group-hover:opacity-100 group-hover:scale-110 transition-all" />
				<span class="group-hover:text-stone-400">IO: Nominal</span>
			</div>
			<div class="flex items-center gap-4 group cursor-default text-white/80 border-b border-rust/30 pb-1">
				<RefreshCw class="w-5 h-5 text-rust animate-spin-slow" />
				<span class="tracking-[0.5em]">RT_PULSE: OK</span>
			</div>
		</div>
	</div>
</div>

<style>
	@keyframes warmup {
		0% { opacity: 1; filter: contrast(3) brightness(0); }
		15% { opacity: 1; filter: contrast(2) brightness(1.5); }
		30% { opacity: 1; filter: contrast(1.5) brightness(0.5); }
		100% { opacity: 0; filter: contrast(1) brightness(1); visibility: hidden; }
	}
	.animate-warmup {
		animation: warmup 1.2s forwards ease-out;
	}

	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.9) 100%);
	}

	.animate-spin-slow {
		animation: spin 8s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}
</style>