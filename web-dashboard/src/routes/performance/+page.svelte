<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Zap,
		Activity,
		Cpu,
		MemoryStick,
		HardDrive,
		Server,
		Database,
		Network,
		RefreshCw,
		AlertTriangle,
		CheckCircle,
		XCircle,
		Gauge,
		Timer,
		Layers,
		Trash2,
		Clock,
		BarChart3,
		ArrowUp,
		ArrowDown,
		Eye,
		ShieldAlert,
		Lock,
		ShieldCheck,
		Info,
		Terminal,
		Radio,
		Dna,
		AlertOctagon,
		Signal
	} from 'lucide-svelte';
	import { fade, fly, scale } from 'svelte/transition';
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
		heap_alloc_history?: number[];
		goroutine_history?: number[];
		gc_pause_history?: number[];
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
		last_seen: number;
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
		max_lifetime_closed: number;
		max_idle_closed: number;
		size: string;
		commits: number;
		rollbacks: number;
		cache_hit_ratio: number;
		tup_fetched: number;
		tup_inserted: number;
		tup_updated: number;
		tup_deleted: number;
	}

	interface NetworkMetrics {
		total_requests: number;
		total_errors: number;
		error_rate: number;
		bytes_sent: number;
		bytes_received: number;
		requests_per_second: number;
		active_connections: number;
		redeye_total_blocks: number;
		redeye_total_rate_limit: number;
		redeye_active_bans: number;
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
	let showRedEyeModal = $state(false);
	let isWarmingUp = $state(true);

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

	function getHealthStatus(metrics: CombinedMetrics): { status: string; color: string; icon: any; glow: string } {
		const heapRatio = metrics.master.heap_usage_ratio;
		const gcCpu = metrics.master.gc_cpu_fraction;
		const errorRate = metrics.network.error_rate;
		const dbConnected = metrics.database.connected;

		if (!dbConnected || heapRatio > 0.9 || errorRate > 20) {
			return { status: 'FAULT', color: '#ef4444', icon: AlertOctagon, glow: 'rgba(239, 68, 68, 0.4)' };
		}
		if (heapRatio > 0.75 || gcCpu > 0.05 || errorRate > 10) {
			return { status: 'DEGRADED', color: '#f97316', icon: AlertTriangle, glow: 'rgba(249, 115, 22, 0.4)' };
		}
		return { status: 'OPTIMAL', color: '#fbbf24', icon: CheckCircle, glow: 'rgba(251, 191, 36, 0.4)' };
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
			const res = await fetch('/api/metrics/gc', { method: 'POST' });
			if (!res.ok) throw new Error('CMD_FAIL');
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
			const res = await fetch('/api/metrics/memory/free', { method: 'POST' });
			if (!res.ok) throw new Error('CMD_FAIL');
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
		setTimeout(() => {
			isWarmingUp = false;
		}, 1000);
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});

	// Reactive health status
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

<div class="min-h-screen bg-[#050505] text-[#e2e8f0] font-['Inter',sans-serif] p-4 sm:p-8 space-y-8 selection:bg-[#f97316] selection:text-black relative overflow-hidden">
	<!-- Industrial UI Overlays -->
	<div class="fixed inset-0 pointer-events-none z-[100] opacity-[0.02] bg-amber-scanlines"></div>
	<div class="fixed inset-0 pointer-events-none z-[101] opacity-[0.15] bg-vignette"></div>

	<!-- Warm-up Animation -->
	{#if isWarmingUp}
		<div class="fixed inset-0 z-[200] bg-[#050505] flex flex-col items-center justify-center animate-warmup font-['JetBrains_Mono']">
			<div class="text-2xl font-extrabold italic tracking-[0.8em] animate-pulse text-[#f97316]">LOAD_CORE_OS</div>
		</div>
	{/if}

	<!-- Header Unit -->
	<div class="border-b-[1px] border-white/10 bg-[#0a0a0a] p-8 relative overflow-hidden shadow-2xl">
		<div class="absolute top-0 right-0 w-96 h-96 bg-[#f97316]/5 rounded-full blur-[100px] -translate-x-1/2 -translate-y-1/2"></div>
		
		<div class="flex flex-col md:flex-row md:items-end justify-between gap-10 relative z-10">
			<div class="space-y-8">
				<div class="flex items-center gap-8">
					<div class="w-24 h-24 border border-white/10 flex items-center justify-center bg-black shadow-[8px_8px_0px_#000] relative group hover:border-[#f97316] transition-all duration-500 overflow-hidden">
						<Terminal class="w-12 h-12 text-[#f97316]" />
						<div class="absolute inset-0 bg-[#f97316]/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
						<div class="absolute bottom-0 left-0 w-full h-[2px] bg-[#f97316] translate-y-full group-hover:translate-y-0 transition-transform"></div>
					</div>
					<div class="space-y-2">
						<div class="flex items-center gap-4">
							<span class="text-[11px] font-black text-[#f97316] tracking-[0.5em] uppercase font-['JetBrains_Mono']">STATION_CMD_01</span>
							<div class="h-[1px] w-16 bg-white/10"></div>
							<span class="text-[11px] font-bold text-slate-500 uppercase tracking-widest italic font-['JetBrains_Mono']">Encrypted_Uplink</span>
						</div>
						<h1 class="text-6xl sm:text-8xl font-black tracking-tight italic leading-none text-white uppercase font-['Inter']">
							Tele<span class="text-[#f97316] underline decoration-4 underline-offset-8">metry</span>
						</h1>
					</div>
				</div>
				<div class="flex flex-wrap gap-10 text-[11px] font-black text-slate-500 tracking-[0.3em] uppercase italic border-t border-white/5 pt-6 font-['JetBrains_Mono']">
					<div class="flex items-center gap-3"><div class="w-2 h-2 bg-[#f97316] shadow-[0_0_10px_#f97316]"></div> STATION_88-FF</div>
					<div class="flex items-center gap-3"><div class="w-2 h-2 bg-white/10"></div> LOC_SECTOR_07</div>
					<div class="flex items-center gap-3 text-[#fbbf24]"><Dna class="w-4 h-4 animate-spin-slow" /> CORE_SYNC_ACTIVE</div>
				</div>
			</div>

			<div class="flex flex-wrap items-center gap-8">
				<!-- Status Badge -->
				{#if health}
					{@const HealthIcon = health.icon}
					<div class="px-10 py-5 bg-black border border-white/10 flex items-center gap-8 shadow-2xl relative overflow-hidden group hover:border-[#f97316] transition-all duration-500">
						<div class="relative">
							<HealthIcon class="w-10 h-10" style="color: {health.color}" />
							<div class="absolute inset-0 blur-2xl opacity-30 animate-pulse" style="background-color: {health.color}"></div>
						</div>
						<div class="flex flex-col relative z-10">
							<span class="text-[10px] font-black text-slate-500 mb-1 tracking-[0.4em] font-['JetBrains_Mono']">INTEGRITY</span>
							<span class="text-3xl font-black tracking-tight italic uppercase font-['Inter']" style="color: {health.color}">{health.status}</span>
						</div>
					</div>
				{/if}

				<!-- Controls -->
				<div class="flex bg-[#0a0a0a] border border-white/10 shadow-2xl p-1.5">
					<button
						onclick={toggleAutoRefresh}
						class="px-8 py-4 hover:bg-[#f97316] hover:text-black transition-all flex items-center gap-4 border border-transparent font-['JetBrains_Mono'] font-extrabold {autoRefresh ? 'text-[#f97316]' : 'text-slate-600'}"
					>
						<Radio class="w-6 h-6 {autoRefresh ? 'animate-pulse' : ''}" />
						<span class="text-sm tracking-widest">{autoRefresh ? 'STREAM_ON' : 'STREAM_OFF'}</span>
					</button>
					<button
						onclick={fetchMetrics}
						disabled={loading}
						class="px-8 py-4 hover:bg-[#f97316] hover:text-black transition-all text-slate-400 border-l border-white/5 disabled:opacity-20"
					>
						<RefreshCw class="w-6 h-6 {loading ? 'animate-spin' : ''}" />
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Primary Metrics Grid -->
	<div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
		<div class="lg:col-span-3 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
			{#if metrics}
				{#each [
					{ label: 'UPTIME_ENGINE', val: formatDuration(metrics.master.uptime_ms), icon: Timer, color: 'text-white' },
					{ label: 'PROCESS_FLOW', val: metrics.master.num_goroutine, icon: Activity, color: 'text-white', trend: goroutineTrend },
					{ label: 'MEMORY_LOAD', val: formatBytes(metrics.master.heap_alloc), icon: MemoryStick, color: 'text-white', detail: `Usage: ${(metrics.master.heap_usage_ratio * 100).toFixed(1)}%` },
					{ label: 'DATA_VELOCITY', val: formatNumber(metrics.network.total_requests), icon: Signal, color: 'text-white', detail: `Rate: ${metrics.network.requests_per_second?.toFixed(1)}Hz` }
				] as block}
					{@const BlockIcon = block.icon}
					<div class="bg-[#0a0a0a] border border-white/5 p-8 group hover:border-[#f97316]/50 transition-all duration-500 relative overflow-hidden shadow-xl">
						<div class="absolute -right-6 -bottom-6 opacity-[0.02] group-hover:opacity-[0.08] transition-opacity duration-1000">
							<BlockIcon size={120} />
						</div>
						<div class="text-[10px] font-black text-slate-500 tracking-[0.4em] mb-4 italic uppercase font-['JetBrains_Mono']">{block.label}</div>
						<div class="flex items-end justify-between relative z-10">
							<div class="text-4xl font-extrabold italic tracking-tighter {block.color} font-['JetBrains_Mono']">{block.val}</div>
							{#if block.trend && block.trend !== 'stable'}
								<div class="text-xl font-black {block.trend === 'up' ? 'text-[#ef4444]' : 'text-[#10b981]'} ml-2 mb-1">
									{block.trend === 'up' ? '▲' : '▼'}
								</div>
							{/if}
						</div>
						{#if block.detail}
							<div class="text-[10px] font-extrabold text-[#f97316] mt-4 uppercase tracking-[0.2em] font-['JetBrains_Mono'] opacity-80">{block.detail}</div>
						{/if}
					</div>
				{/each}
			{/if}
		</div>

		<!-- Station Log (Modern High-Res Mono) -->
		<div class="bg-black border border-white/5 p-8 font-['JetBrains_Mono'] text-[11px] shadow-2xl relative overflow-hidden flex flex-col group hover:border-white/10 transition-colors">
			<div class="flex items-center justify-between mb-6 text-[#f97316] border-b border-white/5 pb-4">
				<div class="flex items-center gap-3">
					<Terminal class="w-4 h-4" />
					<span class="font-black tracking-[0.4em]">STATION_LOG</span>
				</div>
				<div class="w-2.5 h-2.5 rounded-full bg-[#f97316] animate-pulse shadow-[0_0_8px_#f97316]"></div>
			</div>
			<div class="space-y-4 italic text-slate-400 uppercase font-bold flex-1 custom-scrollbar overflow-y-auto">
				<div class="flex items-start gap-4 hover:text-white transition-colors cursor-default group/log">
					<span class="text-slate-600 font-normal">[{lastUpdate?.toLocaleTimeString().split(' ')[0]}]</span>
					<span class="text-white group-hover/log:translate-x-1 transition-transform">&gt;&gt; SYNC_STABLE</span>
				</div>
				{#if metrics}
					<div class="flex items-start gap-4 hover:text-white transition-colors cursor-default group/log">
						<span class="text-slate-600 font-normal">[{lastUpdate?.toLocaleTimeString().split(' ')[0]}]</span>
						<span class="group-hover/log:translate-x-1 transition-transform">&gt;&gt; NODE_HEALTH: <span class="text-[#f97316] underline underline-offset-4 decoration-1 font-extrabold">{health?.status}</span></span>
					</div>
					<div class="flex items-start gap-4 hover:text-white transition-colors cursor-default group/log">
						<span class="text-slate-600 font-normal">[{lastUpdate?.toLocaleTimeString().split(' ')[0]}]</span>
						<span class="group-hover/log:translate-x-1 transition-transform">&gt;&gt; FLEET_LINK: <span class="text-white font-extrabold">{metrics.spawners.online_spawners}</span> UNIT(S)</span>
					</div>
				{/if}
				{#if error}
					<div class="text-[#ef4444] flex items-start gap-4 border-l-2 border-[#ef4444] pl-4 animate-flicker">
						<span class="font-extrabold">&gt;&gt; CRIT_FAULT: {error}</span>
					</div>
				{/if}
			</div>
		</div>
	</div>

	{#if metrics}
		<!-- Industrial Diagnostic Floor -->
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-8 font-['JetBrains_Mono']">
			
			<!-- Memory Engine -->
			<div class="lg:col-span-7 bg-[#0a0a0a] border border-white/5 shadow-2xl relative overflow-hidden group hover:border-[#f97316]/20 transition-all duration-700">
				<div class="p-8 border-b border-white/5 flex items-center justify-between bg-black/60 backdrop-blur-xl">
					<div class="flex items-center gap-6">
						<div class="w-14 h-14 border border-white/10 flex items-center justify-center group-hover:border-[#f97316] transition-colors duration-500 shadow-xl relative overflow-hidden">
							<MemoryStick class="w-8 h-8 text-[#f97316]" />
							<div class="absolute inset-0 bg-[#f97316]/5 animate-pulse"></div>
						</div>
						<div>
							<h3 class="font-black text-2xl tracking-tighter italic uppercase text-white leading-none font-['Inter']">Memory_Core</h3>
							<span class="text-[10px] font-black tracking-[0.5em] text-slate-500 uppercase mt-2 block opacity-80">Phys_Addr_Mapping: Nominal</span>
						</div>
					</div>
					<div class="flex bg-black border border-white/10 p-1.5 gap-1 shadow-inner">
						<button
							onclick={forceGC}
							disabled={gcLoading}
							class="px-6 py-3 text-[11px] font-black hover:bg-[#f97316] hover:text-black transition-all duration-300 disabled:opacity-20 uppercase tracking-[0.2em]"
						>
							{#if gcLoading}
								<RefreshCw class="w-4 h-4 animate-spin" />
							{:else}
								FLUSH
							{/if}
						</button>
						<button
							onclick={freeMemory}
							disabled={freeMemLoading}
							class="px-6 py-3 text-[11px] font-black hover:bg-[#f97316] hover:text-black transition-all duration-300 disabled:opacity-20 uppercase tracking-[0.2em] border-l border-white/5"
						>
							{#if freeMemLoading}
								<RefreshCw class="w-4 h-4 animate-spin" />
							{:else}
								VACUUM
							{/if}
						</button>
					</div>
				</div>

				<div class="p-12 space-y-16">
					<!-- Visual Mapping -->
					<div class="space-y-6">
						<div class="flex justify-between text-[11px] font-black tracking-[0.4em] uppercase italic text-slate-500">
							<span class="flex items-center gap-3"><div class="w-2 h-2 bg-[#f97316] shadow-[0_0_8px_#f97316]"></div> V_MEMORY_LOAD</span>
							<span class="text-white">CAP: {formatBytes(metrics.master.sys)}</span>
						</div>
						<div class="h-20 bg-black border border-white/10 p-2 relative overflow-hidden shadow-inner">
							<div
								class="h-full bg-gradient-to-r from-[#f97316]/20 via-[#f97316]/60 to-[#f97316] border-r-[3px] border-[#f97316] transition-all duration-1000 ease-out relative group-hover:brightness-110"
								style="width: {metrics.master.heap_usage_ratio * 100}%"
							>
								<div class="absolute inset-0 bg-amber-scanlines opacity-20"></div>
								<div class="absolute inset-0 bg-gradient-to-b from-white/10 to-transparent"></div>
							</div>
							<div class="absolute inset-0 flex items-center justify-between px-10 font-black text-lg italic tracking-tighter text-white mix-blend-difference uppercase">
								<span>Engaged: {formatBytes(metrics.master.heap_inuse)}</span>
								<span>{(metrics.master.heap_usage_ratio * 100).toFixed(1)}%</span>
							</div>
						</div>
					</div>

					<!-- Matrix Diagnostics -->
					<div class="grid grid-cols-2 md:grid-cols-3 gap-10">
						{#each [
							{ label: 'HEAP_LIVE', val: formatBytes(metrics.master.heap_alloc) },
							{ label: 'STACK_SYS', val: formatBytes(metrics.master.stack_sys) },
							{ label: 'OBJ_TOTAL', val: formatNumber(metrics.master.live_objects) },
							{ label: 'WRITE_Hz', val: `${formatBytes(metrics.master.heap_alloc_rate)}/s` },
							{ label: 'IDLE_BUF', val: formatBytes(metrics.master.heap_idle) },
							{ label: 'GC_TARGET', val: formatBytes(metrics.master.next_gc_target) }
						] as item}
							<div class="border-l-[1px] border-white/5 bg-white/[0.01] p-6 hover:border-[#f97316] transition-all duration-500 group/stat">
								<div class="text-[10px] font-black text-slate-500 tracking-[0.3em] mb-3 uppercase group-hover/stat:text-[#f97316] transition-colors italic font-['JetBrains_Mono']">{item.label}</div>
								<div class="text-3xl font-extrabold italic tracking-tighter text-white font-['JetBrains_Mono']">{item.val}</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Scrubber Unit -->
			<div class="lg:col-span-5 bg-[#0a0a0a] border border-white/5 shadow-2xl relative overflow-hidden group hover:border-[#f97316]/20 transition-all duration-700">
				<div class="p-8 border-b border-white/5 flex items-center gap-6 bg-black/60 backdrop-blur-xl">
					<div class="w-14 h-14 border border-white/10 flex items-center justify-center">
						<Trash2 class="w-8 h-8 text-[#f97316]" />
					</div>
					<div>
						<h3 class="font-black text-2xl tracking-tighter italic uppercase text-white leading-none font-['Inter']">Scrubber_Log</h3>
						<span class="text-[10px] font-black tracking-[0.5em] text-slate-500 uppercase mt-2 block opacity-80">Logic: Auto_Reclaim</span>
					</div>
				</div>

				<div class="p-12 space-y-12">
					<!-- Load Gauges -->
					<div class="space-y-6">
						<div class="flex justify-between text-[11px] font-black tracking-[0.4em] uppercase italic text-slate-500">
							<span class="flex items-center gap-3">CPU_INTERRUPT_VECTOR</span>
							<span class="text-[#f97316]">{(metrics.master.gc_cpu_fraction * 100).toFixed(4)}%</span>
						</div>
						<div class="flex h-12 gap-1.5 bg-black p-2 border border-white/5 shadow-inner">
							{#each Array(15) as _, i}
								{@const isFilled = i < Math.min(metrics.master.gc_cpu_fraction * 750, 15)}
								{@const isPulsing = i === Math.floor(metrics.master.gc_cpu_fraction * 750)}
								<div 
									class="flex-1 transition-all duration-500 {isFilled ? 'bg-[#f97316] shadow-[0_0_10px_#f97316]' : 'bg-white/[0.02]'} {isPulsing ? 'animate-pulse' : ''}"
								></div>
							{/each}
						</div>
					</div>

					<!-- Trace Table -->
					<div class="bg-black border border-white/5 p-10 space-y-6 shadow-2xl relative group overflow-hidden">
						<div class="absolute top-0 right-0 w-16 h-16 bg-white/[0.02] border-b border-l border-white/5 rotate-0 transition-transform duration-700 group-hover:scale-110"></div>
						{#each [
							{ label: 'Total_Cycles', val: metrics.master.num_gc },
							{ label: 'Avg_Lat', val: formatNanoseconds(metrics.master.avg_gc_pause_ns) },
							{ label: 'Peak_Spike', val: formatNanoseconds(metrics.master.max_gc_pause_ns), color: 'text-[#ef4444]' },
							{ label: 'Forced_Cnt', val: metrics.master.num_forced_gc }
						] as trace}
							<div class="flex justify-between border-b border-white/5 pb-4 text-[12px] font-extrabold italic tracking-tighter">
								<span class="text-slate-500 uppercase tracking-[0.3em] font-['JetBrains_Mono']">{trace.label}</span>
								<span class="{trace.color || 'text-white'} uppercase font-['JetBrains_Mono'] text-lg">{trace.val}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Spawner Array Deck -->
			{#if metrics.spawners.spawner_details && metrics.spawners.spawner_details.length > 0}
				<div class="lg:col-span-12 bg-[#0a0a0a] border border-white/5 shadow-2xl relative overflow-hidden group/fleet">
					<div class="p-10 border-b border-white/5 bg-black/60 backdrop-blur-xl flex flex-col md:flex-row md:items-center justify-between gap-10">
						<div class="flex items-center gap-10">
							<div class="w-24 h-24 border border-[#f97316] flex items-center justify-center bg-[#f97316]/5 shadow-[0_0_40px_rgba(249,115,22,0.15)] relative overflow-hidden">
								<Server class="w-12 h-12 text-[#f97316]" />
								<div class="absolute inset-0 bg-amber-scanlines opacity-10 animate-flicker"></div>
							</div>
							<div>
								<h3 class="font-black text-5xl tracking-tight italic uppercase text-white leading-none font-['Inter']">Fleet_Deployment</h3>
								<div class="flex items-center gap-8 mt-6 font-['JetBrains_Mono']">
									<div class="flex items-center gap-3">
										<div class="w-3 h-3 rounded-full bg-[#10b981] shadow-[0_0_15px_#10b981] animate-pulse"></div>
										<span class="text-[11px] font-black uppercase tracking-[0.4em] text-[#10b981]">{metrics.spawners.online_spawners} ACTIVE_UNITS</span>
									</div>
									<div class="w-[1px] h-6 bg-white/10"></div>
									<span class="text-[11px] font-black uppercase tracking-[0.4em] text-slate-500">{metrics.spawners.running_instances} TOTAL_PROC_NODES</span>
								</div>
							</div>
						</div>

						<!-- Aggregate Telemetry Matrix -->
						<div class="flex bg-black border border-white/10 p-1.5 shadow-2xl">
							{#each [
								{ label: 'AVG_CPU', val: `${metrics.spawners.avg_cpu_usage?.toFixed(1)}%` },
								{ label: 'AVG_MEM', val: `${metrics.spawners.mem_usage_percent?.toFixed(1)}%` },
								{ label: 'AVG_DSK', val: `${metrics.spawners.disk_usage_percent?.toFixed(1)}%` }
							] as stat}
								<div class="px-12 py-6 text-center hover:bg-white/5 transition-all duration-500 cursor-default border-r border-white/5 last:border-none">
									<div class="text-[10px] font-black text-[#f97316] tracking-[0.4em] uppercase mb-2 italic font-['JetBrains_Mono']">{stat.label}</div>
									<div class="text-3xl font-extrabold italic tracking-tighter text-white uppercase font-['JetBrains_Mono']">{stat.val}</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Spawner Units -->
					<div class="divide-y border-t border-white/5">
						{#each metrics.spawners.spawner_details as spawner (spawner.id)}
							<div class="p-12 bg-[#0a0a0a] hover:bg-black transition-all duration-700 group/unit relative overflow-hidden border-white/5">
								<div class="flex flex-col lg:flex-row lg:items-center gap-16 relative z-10">
									<!-- ID & Detail -->
									<div class="flex items-center gap-10 min-w-[380px]">
										<div class="w-20 h-20 border border-white/10 flex items-center justify-center transition-all duration-500 shadow-2xl relative overflow-hidden group-hover/unit:border-[#f97316]">
											<Radio class="w-10 h-10 {spawner.status === 'Online' ? 'text-[#10b981]' : 'text-[#ef4444] animate-pulse'}" />
											<div class="absolute bottom-0 left-0 w-full h-1 {spawner.status === 'Online' ? 'bg-[#10b981]' : 'bg-[#ef4444]'}"></div>
										</div>
										<div>
											<div class="flex items-center gap-5 mb-3">
												<span class="text-4xl font-black italic tracking-tight uppercase text-white group-hover/unit:text-[#f97316] transition-colors font-['Inter']">{spawner.region}</span>
												<span class="text-[11px] bg-[#1a1a1a] text-slate-400 px-4 py-2 font-black italic tracking-[0.3em] border border-white/5 font-['JetBrains_Mono']">UNIT_{spawner.id}</span>
											</div>
											<div class="text-[11px] text-slate-500 font-extrabold italic tracking-[0.2em] flex items-center gap-6 uppercase font-['JetBrains_Mono']">
												<span>Uplink: <span class="text-white">{spawner.host}:{spawner.port}</span></span>
												<span class="w-1.5 h-1.5 bg-white/10 rounded-full"></span>
												<span>v{spawner.game_version}</span>
											</div>
										</div>
									</div>

									<!-- Telemetry Matrix -->
									<div class="flex-1 grid grid-cols-2 md:grid-cols-4 gap-12 font-['JetBrains_Mono']">
										<div class="flex flex-col justify-center border-l-[1px] border-white/5 pl-10">
											<span class="text-[10px] font-black text-slate-500 tracking-[0.3em] uppercase mb-3 italic">Process_Load</span>
											<div class="text-4xl font-extrabold italic tracking-tighter text-white uppercase leading-none">
												{spawner.current_instances} <span class="text-sm text-slate-700 mx-2 font-normal">/</span> {spawner.max_instances}
											</div>
										</div>

										{#each [
											{ label: 'CPU_V', val: spawner.cpu_usage },
											{ label: 'MEM_V', val: spawner.mem_percent },
											{ label: 'DSK_V', val: spawner.disk_percent }
										] as res}
											<div class="flex flex-col justify-center group/gauge">
												<div class="flex justify-between text-[10px] font-black tracking-[0.3em] uppercase mb-4 text-slate-500 italic">
													<span class="group-hover/gauge:text-white transition-colors">{res.label}</span>
													<span class="text-[#f97316] font-extrabold">{res.val?.toFixed(1)}%</span>
												</div>
												<div class="h-2.5 bg-black border border-white/10 p-[2px] shadow-2xl relative overflow-hidden group-hover/unit:border-white/20 transition-colors">
													<div class="h-full bg-gradient-to-r from-[#f97316]/40 to-[#f97316] transition-all duration-1000 shadow-[0_0_15px_rgba(249,115,22,0.3)] relative" style="width: {res.val}%">
														<div class="absolute inset-0 bg-amber-scanlines opacity-20"></div>
													</div>
												</div>
											</div>
										{/each}
									</div>
									
									<div class="px-10 py-5 bg-black border border-white/10 text-[11px] font-black italic text-[#f97316] uppercase tracking-[0.4em] group-hover/unit:border-[#f97316] group-hover:text-black group-hover:bg-[#f97316] transition-all duration-500 cursor-default shadow-2xl font-['JetBrains_Mono']">
										STABLE
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="border-t border-white/5 bg-[#050505] p-12 shadow-2xl relative overflow-hidden group font-['JetBrains_Mono']">
			<div class="absolute inset-0 bg-gradient-to-b from-white/[0.01] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000"></div>
			<div class="flex flex-wrap items-center justify-center gap-16 text-[11px] font-black tracking-[0.5em] uppercase text-slate-600 relative z-10 italic">
				<div class="flex items-center gap-4 hover:text-[#f97316] transition-all duration-300 cursor-default group/fitem">
					<Cpu class="w-5 h-5 text-[#f97316] group-hover/fitem:scale-110 transition-transform" />
					<span>Cores: <span class="text-white">{metrics.master.num_cpu}</span></span>
				</div>
				<div class="flex items-center gap-4 hover:text-[#f97316] transition-all duration-300 cursor-default group/fitem">
					<Layers class="w-5 h-5 text-[#f97316] group-hover/fitem:scale-110 transition-transform" />
					<span>Runtime: <span class="text-white">{metrics.master.go_version}</span></span>
				</div>
				<div class="flex items-center gap-4 hover:text-[#f97316] transition-all duration-300 cursor-default group/fitem">
					<Server class="w-5 h-5 text-[#f97316] group-hover/fitem:scale-110 transition-transform" />
					<span>Arch: <span class="text-white">{metrics.master.goos}/{metrics.master.goarch}</span></span>
				</div>
				<div class="flex items-center gap-4 hover:text-[#f97316] transition-all duration-300 cursor-default group/fitem">
					<BarChart3 class="w-5 h-5 text-[#f97316] group-hover/fitem:scale-110 transition-transform" />
					<span>Allocs: <span class="text-white">{formatNumber(metrics.master.mallocs)}</span></span>
				</div>
			</div>
		</div>
	{:else if loading}
		<div class="flex items-center justify-center py-60 font-['JetBrains_Mono']">
			<div class="flex flex-col items-center gap-12 animate-pulse">
				<div class="w-32 h-32 border-[1px] border-white/5 border-t-[#f97316] rounded-full animate-spin-slow shadow-[0_0_50px_rgba(249,115,22,0.1)]"></div>
				<div class="text-2xl font-black italic tracking-[0.6em] text-[#f97316] uppercase drop-shadow-[0_0_10px_rgba(249,115,22,0.3)]">Initializing_Link</div>
			</div>
		</div>
	{/if}
</div>

{#if showRedEyeModal && metrics}
	<!-- RedEye Modal - High-Definition Modern Industrial -->
	<div
		class="fixed inset-0 z-[300] flex items-center justify-center p-4 bg-[#050505]/99 backdrop-blur-xl font-['Inter',sans-serif]"
		transition:fade={{ duration: 200 }}
	>
		<div
			class="bg-black border border-white/10 shadow-[0_0_150px_rgba(239,68,68,0.15)] w-full max-w-7xl max-h-[95vh] overflow-hidden flex flex-col relative"
			transition:scale={{ duration: 400, start: 0.98, easing: cubicOut }}
		>
			<!-- Header -->
			<div class="p-12 border-b-2 border-[#ef4444] flex justify-between items-center bg-black relative z-10 shadow-2xl">
				<div class="flex items-center gap-10">
					<div class="w-20 h-20 border-[1px] border-[#ef4444] flex items-center justify-center bg-[#ef4444]/5 shadow-[0_0_30px_rgba(239,68,68,0.2)] group hover:bg-[#ef4444]/10 transition-all duration-500">
						<Eye class="w-12 h-12 text-[#ef4444] animate-flicker" />
					</div>
					<div class="space-y-3">
						<div class="flex items-center gap-5">
							<span class="text-[11px] font-black text-[#ef4444] tracking-[0.6em] uppercase font-['JetBrains_Mono'] italic leading-none">SENTINEL_PROTOCOL_V4</span>
							<div class="h-[1px] w-20 bg-red-950"></div>
						</div>
						<h2 class="text-7xl font-black italic uppercase tracking-tighter leading-none text-white font-['Inter']">
							RedEye_<span class="text-[#ef4444] underline decoration-4 underline-offset-12">Shield</span>
						</h2>
					</div>
				</div>
				<button
					onclick={() => (showRedEyeModal = false)}
					class="p-6 border border-white/10 hover:border-[#ef4444] hover:bg-[#ef4444] hover:text-black transition-all duration-500 group relative"
				>
					<XCircle class="w-14 h-14 group-hover:rotate-180 transition-transform duration-700" />
					<div class="absolute inset-0 bg-white/5 opacity-0 group-hover:opacity-100"></div>
				</button>
			</div>

			<!-- Body -->
			<div class="p-16 overflow-y-auto relative z-10 flex-1 custom-scrollbar space-y-20 bg-[#050505]">
				<!-- Top Analytics -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-12 font-['JetBrains_Mono']">
					<div class="lg:col-span-2 border border-white/5 bg-[#0a0a0a] p-16 relative overflow-hidden shadow-2xl group hover:border-[#ef4444]/20 transition-all duration-700">
						<div class="absolute inset-0 bg-gradient-to-br from-red-600/[0.03] to-transparent"></div>
						<div class="flex items-center justify-between mb-16 relative z-10">
							<h4 class="text-md font-black text-slate-500 uppercase tracking-[0.8em] italic leading-none">Threat_Spectrum</h4>
							<AlertOctagon class="w-12 h-12 text-[#ef4444] animate-pulse" />
						</div>
						<div class="flex items-end gap-16 relative z-10">
							<div class="text-9xl font-black italic tracking-tighter leading-none uppercase text-white drop-shadow-[0_0_40px_rgba(255,255,255,0.1)]">
								{metrics.network.redeye_total_blocks > 1000 ? 'Elevated' : 'Nominal'}
							</div>
							<div class="flex-1 space-y-6 pb-4">
								<div class="h-6 bg-black border border-white/10 relative overflow-hidden shadow-inner">
									<div 
										class="h-full bg-gradient-to-r from-red-900 via-red-600 to-[#ef4444] transition-all duration-1000 shadow-[0_0_30px_rgba(239,68,68,0.4)] relative"
										style="width: {metrics.network.redeye_total_blocks > 1000 ? '94%' : '22%'}"
									>
										<div class="absolute inset-0 bg-amber-scanlines opacity-20"></div>
									</div>
								</div>
								<div class="flex justify-between text-[13px] font-black text-slate-600 tracking-[0.5em] uppercase italic">
									<span>Signal_Variance</span>
									<span class="text-red-950 font-extrabold">KEY: 0xCF9</span>
								</div>
							</div>
						</div>
					</div>

					<div class="border border-white/5 bg-[#0a0a0a] p-12 flex flex-col justify-center items-center text-center shadow-2xl relative group hover:border-[#10b981]/20 transition-all duration-700">
						<div class="relative mb-10 animate-pulse">
							<ShieldCheck class="w-32 h-32 text-[#10b981] group-hover:scale-110 transition-transform duration-700" />
							<div class="absolute inset-0 blur-[60px] bg-[#10b981]/20"></div>
						</div>
						<div class="text-[12px] font-black text-slate-500 uppercase tracking-[0.6em] mb-4 italic font-['JetBrains_Mono']">Integrity_Index</div>
						<div class="text-7xl font-black text-white italic tracking-tighter leading-none uppercase font-['JetBrains_Mono']">99.9%</div>
					</div>
				</div>

				<!-- Metric Blocks -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-12 font-['JetBrains_Mono']">
									{#each [
										{ label: 'AGGREGATE_INTERCEPT', val: formatNumber(metrics.network.redeye_total_blocks), color: 'text-white', icon: Lock, barColor: 'bg-red-600', percent: '65%', glow: 'shadow-[0_0_20px_rgba(239,68,68,0.3)]' },
										{ label: 'THROTTLE_THOROUGHPUT', val: formatNumber(metrics.network.redeye_total_rate_limit), color: 'text-white', icon: Gauge, barColor: 'bg-[#f97316]', percent: '45%', glow: 'shadow-[0_0_20px_rgba(249,115,22,0.3)]' }
									] as stat}
										{@const StatIcon = stat.icon}
										<div class="border border-white/5 bg-[#0a0a0a] p-12 relative overflow-hidden shadow-2xl group hover:border-white/10 transition-all duration-500">
											<div class="flex justify-between items-start mb-16 relative z-10">
												<div>
													<h5 class="text-[12px] font-black text-slate-500 uppercase tracking-[0.5em] mb-4 italic leading-none">{stat.label}</h5>
													<div class="text-7xl font-extrabold italic tracking-tighter {stat.color} group-hover:text-[#f97316] transition-colors uppercase leading-none">{stat.val}</div>
												</div>
												<StatIcon class="w-20 h-20 text-white/[0.02] group-hover:text-white/[0.06] transition-all duration-700 group-hover:scale-110" />
											</div>							<div class="h-3 bg-black border border-white/10 relative overflow-hidden group-hover:border-white/20 transition-colors">
								<div class="h-full {stat.barColor} transition-all duration-1000 opacity-40 group-hover:opacity-100 {stat.glow}" style="width: {stat.percent}"></div>
							</div>
						</div>
					{/each}
				</div>

				<!-- Command Logs -->
				<div class="bg-black border border-white/5 p-12 font-['JetBrains_Mono'] text-sm leading-relaxed italic text-slate-400 shadow-2xl relative overflow-hidden group hover:border-white/10 transition-all duration-700">
					<div class="absolute inset-0 bg-gradient-to-r from-red-600/[0.02] to-transparent"></div>
					<div class="flex items-center gap-8 mb-10 text-[#ef4444] font-black tracking-[0.6em] uppercase border-b border-white/5 pb-8 relative z-10">
						<div class="w-12 h-12 border border-[#ef4444] flex items-center justify-center bg-[#ef4444]/5">
							<Info class="w-6 h-6" />
						</div>
						<span class="text-xl">Tactical_Report_012 // Neural_Matrix</span>
					</div>
					<div class="relative z-10 text-lg leading-loose tracking-tight font-medium">
						STATUS: <span class="text-white font-black uppercase tracking-widest bg-white/5 px-3 py-1">Active_Containment</span>. Heuristic engine matching verified threat signatures. Automated quarantine engaged across <span class="text-white font-bold underline decoration-[#ef4444] decoration-2 underline-offset-8">All Sectors</span>. Reputation variance stable (<span class="text-[#f97316] font-extrabold">0.024%</span>). Node synchronization complete. Defense hot-load operational.
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="p-16 border-t border-white/5 bg-black flex flex-col md:flex-row justify-between items-center gap-12 font-['JetBrains_Mono']">
				<div class="flex items-center gap-16 text-[14px] font-black text-slate-600 tracking-[0.6em] uppercase italic">
					<span class="flex items-center gap-5">
						<div class="w-4 h-4 rounded-full bg-[#ef4444] animate-pulse shadow-[0_0_15px_#ef4444]"></div>
						UPLINK_HOT
					</span>
					<span class="hidden lg:block border-l border-white/5 pl-16">CIPHER: AES_256_RSA_OLD</span>
				</div>
				<button
					onclick={() => (showRedEyeModal = false)}
					class="w-full md:w-auto px-24 py-8 bg-[#f97316] hover:bg-white text-black font-black uppercase tracking-[0.6em] italic transition-all shadow-2xl hover:shadow-[0_0_50px_rgba(255,255,255,0.25)] active:scale-95 text-lg font-['Inter']"
				>
					ACKNOWLEDGE
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* CRT & Industrial Styles */
	.bg-amber-scanlines {
		background: linear-gradient(
			rgba(18, 16, 16, 0) 50%,
			rgba(0, 0, 0, 0.2) 50%
		),
		linear-gradient(
			90deg,
			rgba(255, 0, 0, 0.03),
			rgba(0, 255, 0, 0.01),
			rgba(0, 0, 255, 0.03)
		);
		background-size: 100% 4px, 4px 100%;
	}

	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.9) 100%);
	}

	@keyframes warmup {
		0% { opacity: 1; filter: contrast(3) brightness(0); }
		15% { opacity: 1; filter: contrast(2) brightness(1.5); }
		30% { opacity: 1; filter: contrast(1.5) brightness(0.5); }
		100% { opacity: 0; filter: contrast(1) brightness(1); visibility: hidden; }
	}
	.animate-warmup {
		animation: warmup 1.2s forwards ease-out;
	}

	@keyframes loadingBar {
		0% { transform: translateX(-100%); }
		100% { transform: translateX(100%); }
	}
	.animate-loading-bar {
		animation: loadingBar 1.5s linear infinite;
	}

	.animate-spin-slow {
		animation: rotate 12s linear infinite;
	}
	@keyframes rotate {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.8; }
		55% { opacity: 0.95; }
		60% { opacity: 0.7; }
	}
	.animate-flicker {
		animation: flicker 0.25s infinite;
	}

	/* Refined Industrial Scrollbar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: #050505;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
		border: 1px solid #050505;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #f97316;
	}

	/* Typography Overrides */
	:global(body) {
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
	}
</style>