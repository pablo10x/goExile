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
		TrendingUp,
		TrendingDown,
		AlertTriangle,
		CheckCircle,
		XCircle,
		Gauge,
		Timer,
		Layers,
		GitBranch,
		Trash2,
		PlayCircle,
		Clock,
		BarChart3,
		PieChart,
		ArrowUp,
		ArrowDown,
		Minus,
		Eye,
		ShieldAlert,
		Lock,
		ShieldCheck,
		Search,
		Info
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
		const gcCpu = metrics.master.gc_cpu_fraction;
		const errorRate = metrics.network.error_rate;
		const dbConnected = metrics.database.connected;

		if (!dbConnected || heapRatio > 0.9 || errorRate > 20) {
			return { status: 'Critical', color: 'red', icon: XCircle };
		}
		if (heapRatio > 0.75 || gcCpu > 0.05 || errorRate > 10) {
			return { status: 'Warning', color: 'amber', icon: AlertTriangle };
		}
		return { status: 'Healthy', color: 'emerald', icon: CheckCircle };
	}

	function getTrend(current: number, previous: number | undefined): 'up' | 'down' | 'stable' {
		if (previous === undefined) return 'stable';
		const diff = current - previous;
		const threshold = Math.abs(previous) * 0.05; // 5% change threshold
		if (diff > threshold) return 'up';
		if (diff < -threshold) return 'down';
		return 'stable';
	}

	// API calls
	async function fetchMetrics() {
		try {
			const res = await fetch('/api/metrics');
			if (!res.ok) throw new Error('Failed to fetch metrics');
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
			if (!res.ok) throw new Error('Failed to trigger GC');
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
			if (!res.ok) throw new Error('Failed to free memory');
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

<div class="space-y-6 p-4 sm:p-6 pb-24 md:pb-6">
	<!-- Header -->
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
		<div>
			<h1
				class="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-white flex items-center gap-3"
			>
				<Zap class="w-6 h-6 sm:w-8 sm:h-8 text-yellow-400" />
				Performance
			</h1>
			<p
				class="text-slate-500 dark:text-slate-400 mt-1 sm:mt-2 flex items-center gap-2 text-xs sm:text-sm"
			>
				<Activity class="w-3 h-3 sm:w-4 sm:h-4 text-emerald-400" />
				Real-time monitoring
			</p>
		</div>

		<div class="flex flex-wrap items-center gap-2 sm:gap-3">
			<!-- Health Status -->
			{#if health}
				{@const HealthIcon = health.icon}
				<div
					class="px-3 py-1.5 sm:px-4 sm:py-2 rounded-xl border flex items-center gap-2 text-xs sm:text-sm {health.color ===
					'emerald'
						? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
						: health.color === 'amber'
							? 'bg-amber-500/10 border-amber-500/30 text-amber-400'
							: 'bg-red-500/10 border-red-500/30 text-red-400'}"
					transition:scale={{ start: 0.9, duration: 200 }}
				>
					<HealthIcon class="w-4 h-4 sm:w-5 sm:h-5" />
					<span class="font-semibold">{health.status}</span>
				</div>
			{/if}
			<!-- Auto Refresh Toggle -->
			<button
				onclick={toggleAutoRefresh}
				class="px-3 py-1.5 sm:px-4 sm:py-2 rounded-xl border transition-all duration-200 flex items-center gap-2 text-xs sm:text-sm {autoRefresh
					? 'bg-blue-500/10 border-blue-500/30 text-blue-400'
					: 'bg-slate-800/50 border-slate-300/50 dark:border-slate-700/50 text-slate-500 dark:text-slate-400'}"
			>
				<RefreshCw
					class="w-3 h-3 sm:w-4 sm:h-4 {autoRefresh ? 'animate-spin' : ''}"
					style="animation-duration: 3s"
				/>
				<span class="font-medium">{autoRefresh ? 'Auto' : 'Paused'}</span>
			</button>

			<!-- Manual Refresh -->
			<button
				onclick={fetchMetrics}
				disabled={loading}
				class="p-1.5 sm:p-2 rounded-xl bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:border-slate-600 transition-all disabled:opacity-50"
			>
				<RefreshCw class="w-4 h-4 sm:w-5 sm:h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Last Update -->
	{#if lastUpdate}
		<div class="text-[10px] sm:text-xs text-slate-500 flex items-center gap-2">
			<Clock class="w-3 h-3" />
			Last updated: {lastUpdate.toLocaleTimeString()}
		</div>
	{/if}

	<!-- Error Banner -->
	{#if error}
		<div
			class="p-4 bg-red-500/10 border border-red-500/30 rounded-xl flex items-center gap-3 text-red-400"
			transition:fly={{ y: -10, duration: 300 }}
		>
			<AlertTriangle class="w-5 h-5" />
			<span class="text-sm">{error}</span>
		</div>
	{/if}

	{#if metrics}
		<!-- Quick Stats Row -->
		<div
			class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4"
			transition:fade={{ duration: 300 }}
		>
			<!-- Uptime -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<Timer class="w-3 h-3 sm:w-4 sm:h-4" />
					UPTIME
				</div>
				<div class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white truncate">
					{formatDuration(metrics.master.uptime_ms)}
				</div>
			</div>

			<!-- Goroutines -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<GitBranch class="w-3 h-3 sm:w-4 sm:h-4" />
					GOROUTINES
				</div>
				<div
					class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2"
				>
					{metrics.master.num_goroutine}
					{#if goroutineTrend === 'up'}
						<ArrowUp class="w-3 h-3 sm:w-4 sm:h-4 text-amber-400" />
					{:else if goroutineTrend === 'down'}
						<ArrowDown class="w-3 h-3 sm:w-4 sm:h-4 text-emerald-400" />
					{:else}
						<Minus class="w-3 h-3 sm:w-4 sm:h-4 text-slate-500" />
					{/if}
				</div>
				<div class="text-[10px] sm:text-xs text-slate-500 mt-1">
					Peak: {metrics.master.peak_goroutines}
				</div>
			</div>

			<!-- Heap Usage -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<MemoryStick class="w-3 h-3 sm:w-4 sm:h-4" />
					HEAP ALLOC
				</div>
				<div
					class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2 truncate"
				>
					{formatBytes(metrics.master.heap_alloc)}
					{#if heapTrend === 'up'}
						<ArrowUp class="w-3 h-3 sm:w-4 sm:h-4 text-amber-400" />
					{:else if heapTrend === 'down'}
						<ArrowDown class="w-3 h-3 sm:w-4 sm:h-4 text-emerald-400" />
					{:else}
						<Minus class="w-3 h-3 sm:w-4 sm:h-4 text-slate-500" />
					{/if}
				</div>
				<div class="text-[10px] sm:text-xs text-slate-500 mt-1">
					{(metrics.master.heap_usage_ratio * 100)?.toFixed(1) || 0}% of sys
				</div>
			</div>

			<!-- GC Cycles -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<Trash2 class="w-3 h-3 sm:w-4 sm:h-4" />
					GC CYCLES
				</div>
				<div class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white">
					{metrics.master.num_gc}
				</div>
				<div class="text-[10px] sm:text-xs text-slate-500 mt-1">
					CPU: {(metrics.master.gc_cpu_fraction * 100)?.toFixed(2) || 0}%
				</div>
			</div>

			<!-- Active Spawners -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<Server class="w-3 h-3 sm:w-4 sm:h-4" />
					SPAWNERS
				</div>
				<div class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white">
					<span class="text-emerald-400">{metrics.spawners.online_spawners}</span>
					<span class="text-slate-500">/{metrics.spawners.total_spawners}</span>
				</div>
				<div class="text-[10px] sm:text-xs text-slate-50">
					{metrics.spawners.running_instances} instances
				</div>
			</div>

			<!-- Requests -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-3 sm:p-4"
			>
				<div
					class="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs font-medium mb-1 sm:mb-2"
				>
					<Network class="w-3 h-3 sm:w-4 sm:h-4" />
					REQUESTS
				</div>
				<div class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-white truncate">
					{formatNumber(metrics.network.total_requests)}
				</div>
				<div class="text-[10px] sm:text-xs text-slate-500 mt-1">
					{metrics.network.requests_per_second?.toFixed(1) || 0}/s
				</div>
			</div>
		</div>

		<!-- Main Grid -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
			<!-- Master Server Memory Card -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 100 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex items-center justify-between"
				>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-500/20 rounded-lg">
							<MemoryStick class="w-5 h-5 text-blue-400" />
						</div>
						<div>
							<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
								Memory Management
							</h3>
							<p class="text-[10px] sm:text-xs text-slate-500">Go Runtime Memory Statistics</p>
						</div>
					</div>
					<div class="flex gap-2">
						<button
							onclick={forceGC}
							disabled={gcLoading}
							class="px-2 py-1 sm:px-3 sm:py-1.5 text-[10px] sm:text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded-lg hover:bg-amber-500/20 transition-colors disabled:opacity-50 flex items-center gap-1.5"
						>
							{#if gcLoading}
								<RefreshCw class="w-3 h-3 animate-spin" />
							{:else}
								<Trash2 class="w-3 h-3" />
							{/if}
							<span class="hidden sm:inline">Force GC</span>
							<span class="sm:hidden">GC</span>
						</button>
						<button
							onclick={freeMemory}
							disabled={freeMemLoading}
							class="px-2 py-1 sm:px-3 sm:py-1.5 text-[10px] sm:text-xs font-medium bg-purple-500/10 text-purple-400 border border-purple-500/30 rounded-lg hover:bg-purple-500/20 transition-colors disabled:opacity-50 flex items-center gap-1.5"
						>
							{#if freeMemLoading}
								<RefreshCw class="w-3 h-3 animate-spin" />
							{:else}
								<ArrowDown class="w-3 h-3" />
							{/if}
							<span class="hidden sm:inline">Free OS Mem</span>
							<span class="sm:hidden">Free</span>
						</button>
					</div>
				</div>

				<div class="p-4 space-y-4">
					<!-- Heap Visual -->
					<div class="relative h-8 bg-slate-900/50 rounded-lg overflow-hidden">
						<div
							class="absolute inset-y-0 left-0 bg-gradient-to-r from-blue-600 to-blue-400 transition-all duration-500"
							style="width: {metrics.master.heap_usage_ratio * 100}%"
						></div>
						<div
							class="absolute inset-0 flex items-center justify-center text-[10px] sm:text-xs font-medium text-slate-900 dark:text-white"
						>
							{formatBytes(metrics.master.heap_inuse)} / {formatBytes(metrics.master.heap_sys)}
						</div>
					</div>

					<!-- Memory Grid -->
					<div class="grid grid-cols-2 gap-3">
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Heap Allocated</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.heap_alloc)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Heap Idle</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.heap_idle)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Stack In Use</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.stack_inuse)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Total System</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.sys)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Live Objects</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNumber(metrics.master.live_objects)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Alloc Rate</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.heap_alloc_rate)}/s
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Garbage Collection Card -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 150 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex items-center gap-3"
				>
					<div class="p-2 bg-amber-500/20 rounded-lg">
						<Trash2 class="w-5 h-5 text-amber-400" />
					</div>
					<div>
						<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
							Garbage Collection
						</h3>
						<p class="text-[10px] sm:text-xs text-slate-500">GC Performance & Statistics</p>
					</div>
				</div>

				<div class="p-4 space-y-4">
					<!-- GC CPU Impact -->
					<div>
						<div class="flex items-center justify-between text-xs sm:text-sm mb-2">
							<span class="text-slate-500 dark:text-slate-400">GC CPU Usage</span>
							<span
								class="font-medium {metrics.master.gc_cpu_fraction > 0.05
									? 'text-red-400'
									: metrics.master.gc_cpu_fraction > 0.02
										? 'text-amber-400'
										: 'text-emerald-400'}"
							>
								{(metrics.master.gc_cpu_fraction * 100).toFixed(3)}%
							</span>
						</div>
						<div class="h-2 bg-slate-900/50 rounded-full overflow-hidden">
							<div
								class="h-full transition-all duration-500 {metrics.master.gc_cpu_fraction > 0.05
									? 'bg-red-500'
									: metrics.master.gc_cpu_fraction > 0.02
										? 'bg-amber-500'
										: 'bg-emerald-500'}"
								style="width: {Math.min(metrics.master.gc_cpu_fraction * 1000, 100)}%"
							></div>
						</div>
					</div>

					<!-- GC Stats Grid -->
					<div class="grid grid-cols-2 gap-3">
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Total GC Cycles</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{metrics.master.num_gc}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Forced GC</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{metrics.master.num_forced_gc}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Last Pause</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNanoseconds(metrics.master.last_gc_pause_ns)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Avg Pause</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNanoseconds(metrics.master.avg_gc_pause_ns)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Max Pause</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNanoseconds(metrics.master.max_gc_pause_ns)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Next GC Target</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatBytes(metrics.master.next_gc_target)}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Database Performance Card -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 200 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex items-center justify-between"
				>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-emerald-500/20 rounded-lg">
							<Database class="w-5 h-5 text-emerald-400" />
						</div>
						<div>
							<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
								Database Performance
							</h3>
							<p class="text-[10px] sm:text-xs text-slate-500">PostgreSQL Connection Pool</p>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div
							class="w-2 h-2 rounded-full {metrics.database.connected
								? 'bg-emerald-500'
								: 'bg-red-500'}"
						></div>
						<span
							class="text-[10px] sm:text-xs {metrics.database.connected
								? 'text-emerald-400'
								: 'text-red-400'}"
						>
							{metrics.database.connected ? 'Connected' : 'Disconnected'}
						</span>
					</div>
				</div>

				<div class="p-4 space-y-4">
					<!-- Connection Pool Visual -->
					<div>
						<div class="flex items-center justify-between text-xs sm:text-sm mb-2">
							<span class="text-slate-500 dark:text-slate-400">Connection Pool</span>
							<span class="text-slate-900 dark:text-white font-medium"
								>{metrics.database.in_use} / {metrics.database.open_connections}</span
							>
						</div>
						<div class="h-3 bg-slate-900/50 rounded-full overflow-hidden flex">
							<div
								class="bg-blue-500 transition-all duration-500"
								style="width: {metrics.database.open_connections > 0
									? (metrics.database.in_use / metrics.database.open_connections) * 100
									: 0}%"
							></div>
							<div
								class="bg-slate-600 transition-all duration-500"
								style="width: {metrics.database.open_connections > 0
									? (metrics.database.idle / metrics.database.open_connections) * 100
									: 0}%"
							></div>
						</div>
						<div class="flex gap-4 mt-2 text-[10px] sm:text-xs">
							<span class="text-blue-400">● In Use: {metrics.database.in_use}</span>
							<span class="text-slate-500 dark:text-slate-400">● Idle: {metrics.database.idle}</span
							>
						</div>
					</div>

					<!-- Cache Hit Ratio -->
					<div>
						<div class="flex items-center justify-between text-xs sm:text-sm mb-2">
							<span class="text-slate-500 dark:text-slate-400">Cache Hit Ratio</span>
							<span
								class="font-medium {metrics.database.cache_hit_ratio > 95
									? 'text-emerald-400'
									: metrics.database.cache_hit_ratio > 80
										? 'text-amber-400'
										: 'text-red-400'}"
							>
								{metrics.database.cache_hit_ratio?.toFixed(2)}%
							</span>
						</div>
						<div class="h-2 bg-slate-900/50 rounded-full overflow-hidden">
							<div
								class="h-full transition-all duration-500 {metrics.database.cache_hit_ratio > 95
									? 'bg-emerald-500'
									: metrics.database.cache_hit_ratio > 80
										? 'bg-amber-500'
										: 'bg-red-500'}"
								style="width: {metrics.database.cache_hit_ratio}%"
							></div>
						</div>
					</div>

					<!-- DB Stats Grid -->
					<div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Size</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{metrics.database.size || 'N/A'}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Commits</div>
							<div class="text-sm sm:text-lg font-semibold text-emerald-400">
								{formatNumber(metrics.database.commits)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Rollbacks</div>
							<div class="text-sm sm:text-lg font-semibold text-red-400">
								{formatNumber(metrics.database.rollbacks)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Rows Fetched</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNumber(metrics.database.tup_fetched)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Rows Inserted</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNumber(metrics.database.tup_inserted)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Wait Count</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNumber(metrics.database.wait_count)}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Network & API Card -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 250 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex items-center gap-3"
				>
					<div class="p-2 bg-purple-500/20 rounded-lg">
						<Network class="w-5 h-5 text-purple-400" />
					</div>
					<div>
						<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
							Network & API
						</h3>
						<p class="text-[10px] sm:text-xs text-slate-500">Request Statistics & Bandwidth</p>
					</div>
				</div>

				<div class="p-4 space-y-4">
					<!-- Error Rate -->
					<div>
						<div class="flex items-center justify-between text-xs sm:text-sm mb-2">
							<span class="text-slate-500 dark:text-slate-400">Error Rate</span>
							<span
								class="font-medium {metrics.network.error_rate < 1
									? 'text-emerald-400'
									: metrics.network.error_rate < 5
										? 'text-amber-400'
										: 'text-red-400'}"
							>
								{metrics.network.error_rate?.toFixed(2)}%
							</span>
						</div>
						<div class="h-2 bg-slate-900/50 rounded-full overflow-hidden">
							<div
								class="h-full transition-all duration-500 {metrics.network.error_rate < 1
									? 'bg-emerald-500'
									: metrics.network.error_rate < 5
										? 'bg-amber-500'
										: 'bg-red-500'}"
								style="width: {Math.min(metrics.network.error_rate * 10, 100)}%"
							></div>
						</div>
					</div>

					<!-- Network Stats Grid -->
					<div class="grid grid-cols-2 gap-3">
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Total Requests</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{formatNumber(metrics.network.total_requests)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Total Errors</div>
							<div class="text-sm sm:text-lg font-semibold text-red-400">
								{formatNumber(metrics.network.total_errors)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Requests/sec</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{metrics.network.requests_per_second?.toFixed(2)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Active WS Connections</div>
							<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
								{metrics.network.active_connections}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Bytes Sent</div>
							<div class="text-sm sm:text-lg font-semibold text-emerald-400">
								{formatBytes(metrics.network.bytes_sent)}
							</div>
						</div>
						<div class="bg-slate-900/30 rounded-lg p-3">
							<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Bytes Received</div>
							<div class="text-sm sm:text-lg font-semibold text-blue-400">
								{formatBytes(metrics.network.bytes_received)}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- RedEye Guardian Performance Card -->
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 275 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex items-center justify-between"
				>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-red-500/20 rounded-lg">
							<Eye class="w-5 h-5 text-red-400" />
						</div>
						<div>
							<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
								RedEye Guardian
							</h3>
							<p class="text-[10px] sm:text-xs text-slate-500">Security & Traffic Shield</p>
						</div>
					</div>
					<button
						onclick={() => (showRedEyeModal = true)}
						class="px-3 py-1.5 text-[10px] sm:text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/30 rounded-lg hover:bg-red-500/20 transition-colors flex items-center gap-1.5"
					>
						<BarChart3 class="w-3 h-3" />
						Details
					</button>
				</div>

				<div class="p-4 space-y-4">
					<!-- Active Bans -->
					<div>
						<div class="flex items-center justify-between text-xs sm:text-sm mb-2">
							<span class="text-slate-500 dark:text-slate-400">Security Shield</span>
							<span class="text-emerald-400 font-medium flex items-center gap-1.5">
								<ShieldCheck class="w-3.5 h-3.5" />
								Active
							</span>
						</div>
						<div class="grid grid-cols-2 gap-3">
							<div class="bg-slate-900/30 rounded-lg p-3">
								<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Total Blocks</div>
								<div class="text-sm sm:text-lg font-semibold text-red-400">
									{formatNumber(metrics.network.redeye_total_blocks)}
								</div>
							</div>
							<div class="bg-slate-900/30 rounded-lg p-3">
								<div class="text-[10px] sm:text-xs text-slate-500 mb-1">Active Bans</div>
								<div class="text-sm sm:text-lg font-semibold text-slate-900 dark:text-white">
									{metrics.network.redeye_active_bans}
								</div>
							</div>
						</div>
					</div>

					<!-- Rate Limiting Impact -->
					<div class="bg-amber-500/5 border border-amber-500/10 rounded-lg p-3">
						<div class="flex items-center justify-between mb-1">
							<span class="text-[10px] text-amber-500/70 uppercase tracking-wider font-bold"
								>Traffic Throttling</span
							>
							<Lock class="w-3 h-3 text-amber-500/50" />
						</div>
						<div class="text-xl font-bold text-amber-400">
							{formatNumber(metrics.network.redeye_total_rate_limit)}
						</div>
						<div class="text-[10px] text-slate-500 mt-1">Total requests rate-limited</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Spawners Detail Section -->
		{#if metrics.spawners.spawner_details && metrics.spawners.spawner_details.length > 0}
			<div
				class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
				transition:fly={{ y: 20, duration: 400, delay: 300 }}
			>
				<div
					class="p-4 border-b border-slate-300/50 dark:border-slate-700/50 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"
				>
					<div class="flex items-center gap-3">
						<div class="p-2 bg-cyan-500/20 rounded-lg">
							<Server class="w-5 h-5 text-cyan-400" />
						</div>
						<div>
							<h3 class="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">
								Spawner Fleet
							</h3>
							<p class="text-[10px] sm:text-xs text-slate-500">
								{metrics.spawners.online_spawners} online of {metrics.spawners.total_spawners} total •
								{metrics.spawners.running_instances} instances running
							</p>
						</div>
					</div>

					<!-- Aggregate Stats -->
					<div class="flex gap-4 text-xs sm:text-sm w-full sm:w-auto justify-around sm:justify-end">
						<div class="text-center">
							<div class="text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs">Avg CPU</div>
							<div class="font-semibold text-slate-900 dark:text-white">
								{metrics.spawners.avg_cpu_usage?.toFixed(1)}%
							</div>
						</div>
						<div class="text-center">
							<div class="text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs">Memory</div>
							<div class="font-semibold text-slate-900 dark:text-white">
								{metrics.spawners.mem_usage_percent?.toFixed(1)}%
							</div>
						</div>
						<div class="text-center">
							<div class="text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs">Disk</div>
							<div class="font-semibold text-slate-900 dark:text-white">
								{metrics.spawners.disk_usage_percent?.toFixed(1)}%
							</div>
						</div>
					</div>
				</div>

				<div class="divide-y divide-slate-700/50">
					{#each metrics.spawners.spawner_details as spawner (spawner.id)}
						<div class="p-4 hover:bg-slate-700/20 transition-colors">
							<div
								class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"
							>
								<!-- Spawner Info -->
								<div class="flex items-center gap-4 w-full sm:w-auto">
									<div
										class="w-2 h-2 rounded-full {spawner.status === 'Online'
											? 'bg-emerald-500'
											: 'bg-red-500'}"
									></div>
									<div class="flex-1 sm:flex-initial">
										<div
											class="font-medium text-slate-900 dark:text-white flex items-center gap-2 text-sm sm:text-base"
										>
											{spawner.region}
											<span class="text-[10px] sm:text-xs text-slate-500">#{spawner.id}</span>
										</div>
										<div class="text-[10px] sm:text-xs text-slate-500">
											{spawner.host}:{spawner.port}
											{#if spawner.game_version}
												<span
													class="ml-2 px-1.5 py-0.5 bg-slate-700 rounded text-slate-500 dark:text-slate-400"
													>v{spawner.game_version}</span
												>
											{/if}
										</div>
									</div>
									<!-- Instances (Mobile: Moved next to name) -->
									<div class="text-center px-2 sm:hidden">
										<div class="text-[10px] text-slate-500">Inst.</div>
										<div class="font-semibold text-slate-900 dark:text-white text-sm">
											{spawner.current_instances}
										</div>
									</div>
								</div>

								<!-- Stats Grid -->
								<div class="grid grid-cols-3 sm:flex items-center gap-4 w-full sm:w-auto">
									<!-- Instances (Desktop) -->
									<div class="text-center px-4 hidden sm:block">
										<div class="text-xs text-slate-500">Instances</div>
										<div class="font-semibold text-slate-900 dark:text-white">
											{spawner.current_instances} / {spawner.max_instances}
										</div>
									</div>

									<!-- CPU -->
									<div class="w-full sm:w-32">
										<div class="flex items-center justify-between text-[10px] sm:text-xs mb-1">
											<span class="text-slate-500">CPU</span>
											<span
												class={spawner.cpu_usage > 80
													? 'text-red-400'
													: spawner.cpu_usage > 60
														? 'text-amber-400'
														: 'text-emerald-400'}>{spawner.cpu_usage?.toFixed(1)}%</span
											>
										</div>
										<div class="h-1.5 bg-slate-900/50 rounded-full overflow-hidden">
											<div
												class="h-full transition-all duration-500 {spawner.cpu_usage > 80
													? 'bg-red-500'
													: spawner.cpu_usage > 60
														? 'bg-amber-500'
														: 'bg-emerald-500'}"
												style="width: {spawner.cpu_usage}%"
											></div>
										</div>
									</div>

									<!-- Memory -->
									<div class="w-full sm:w-32">
										<div class="flex items-center justify-between text-[10px] sm:text-xs mb-1">
											<span class="text-slate-500">MEM</span>
											<span
												class={spawner.mem_percent > 85
													? 'text-red-400'
													: spawner.mem_percent > 70
														? 'text-amber-400'
														: 'text-emerald-400'}>{spawner.mem_percent?.toFixed(1)}%</span
											>
										</div>
										<div class="h-1.5 bg-slate-900/50 rounded-full overflow-hidden">
											<div
												class="h-full transition-all duration-500 {spawner.mem_percent > 85
													? 'bg-red-500'
													: spawner.mem_percent > 70
														? 'bg-amber-500'
														: 'bg-emerald-500'}"
												style="width: {spawner.mem_percent}%"
											></div>
										</div>
									</div>

									<!-- Disk -->
									<div class="w-full sm:w-32">
										<div class="flex items-center justify-between text-[10px] sm:text-xs mb-1">
											<span class="text-slate-500">DISK</span>
											<span
												class={spawner.disk_percent > 90
													? 'text-red-400'
													: spawner.disk_percent > 75
														? 'text-amber-400'
														: 'text-emerald-400'}>{spawner.disk_percent?.toFixed(1)}%</span
											>
										</div>
										<div class="h-1.5 bg-slate-900/50 rounded-full overflow-hidden">
											<div
												class="h-full transition-all duration-500 {spawner.disk_percent > 90
													? 'bg-red-500'
													: spawner.disk_percent > 75
														? 'bg-amber-500'
														: 'bg-emerald-500'}"
												style="width: {spawner.disk_percent}%"
											></div>
										</div>
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- System Info Footer -->
		<div
			class="bg-slate-800/30 border border-slate-300/30 dark:border-slate-700/30 rounded-xl p-4"
			transition:fade={{ duration: 300, delay: 350 }}
		>
			<div
				class="flex flex-wrap items-center justify-center gap-4 sm:gap-6 text-[10px] sm:text-xs text-slate-500"
			>
				<div class="flex items-center gap-2">
					<Cpu class="w-3 h-3 sm:w-4 sm:h-4" />
					<span>{metrics.master.num_cpu} CPUs</span>
				</div>
				<div class="flex items-center gap-2">
					<Layers class="w-3 h-3 sm:w-4 sm:h-4" />
					<span>{metrics.master.go_version}</span>
				</div>
				<div class="flex items-center gap-2">
					<Server class="w-3 h-3 sm:w-4 sm:h-4" />
					<span>{metrics.master.goos}/{metrics.master.goarch}</span>
				</div>
				<div class="flex items-center gap-2">
					<BarChart3 class="w-3 h-3 sm:w-4 sm:h-4" />
					<span>{formatNumber(metrics.master.mallocs)} allocs</span>
				</div>
				<div class="flex items-center gap-2">
					<Trash2 class="w-3 h-3 sm:w-4 sm:h-4" />
					<span>{formatNumber(metrics.master.frees)} frees</span>
				</div>
			</div>
		</div>
	{:else if loading}
		<!-- Loading State -->
		<div class="flex items-center justify-center py-20">
			<div class="flex flex-col items-center gap-4">
				<RefreshCw class="w-10 h-10 text-blue-400 animate-spin" />
				<p class="text-slate-500 dark:text-slate-400">Loading metrics...</p>
			</div>
		</div>
	{/if}
</div>

{#if showRedEyeModal && metrics}
	<div
		class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
		transition:fade={{ duration: 200 }}
	>
		<div
			class="bg-slate-900 border border-red-500/30 rounded-3xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col relative"
			transition:scale={{ duration: 300, start: 0.95, easing: cubicOut }}
		>
			<!-- Neural Grid Background -->
			<div class="absolute inset-0 opacity-10 pointer-events-none">
				<div
					class="absolute inset-0"
					style="background-image: radial-gradient(circle at 2px 2px, #ef4444 1px, transparent 0); background-size: 24px 24px;"
				></div>
			</div>

			<!-- Header -->
			<div
				class="p-6 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center bg-white/50 dark:bg-slate-950/50 relative z-10"
			>
				<div class="flex items-center gap-4">
					<div class="p-3 bg-red-500/20 rounded-2xl relative overflow-hidden group">
						<Eye class="w-6 h-6 text-red-500 relative z-10" />
						<div
							class="absolute inset-0 bg-red-500/20 translate-y-full group-hover:translate-y-0 transition-transform duration-500"
						></div>
					</div>
					<div>
						<h2
							class="text-2xl font-black text-slate-900 dark:text-white uppercase tracking-tighter"
						>
							RedEye Neural Core
						</h2>
						<p class="text-[10px] text-red-500/70 font-mono uppercase tracking-[0.2em]">
							Active Defense Subsystem Metrics
						</p>
					</div>
				</div>
				<button
					onclick={() => (showRedEyeModal = false)}
					class="p-2 hover:bg-white/5 rounded-xl text-slate-500 hover:text-slate-900 dark:text-white transition-all"
				>
					<XCircle class="w-6 h-6" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 overflow-y-auto relative z-10 flex-1 custom-scrollbar">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
					<!-- Threat Level -->
					<div
						class="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 relative group overflow-hidden"
					>
						<div class="absolute top-0 left-0 w-1 h-full bg-red-500"></div>
						<div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-4">
							Threat Level
						</div>
						<div class="flex items-end gap-2">
							<div class="text-4xl font-black text-slate-900 dark:text-white font-mono">
								{metrics.network.redeye_total_blocks > 1000 ? 'ELEVATED' : 'STABLE'}
							</div>
						</div>
						<div class="mt-4 flex gap-1">
							{#each Array(5) as _, i}
								<div
									class="h-1 flex-1 rounded-full {i <
									(metrics.network.redeye_total_blocks > 1000 ? 4 : 2)
										? 'bg-red-500'
										: 'bg-slate-800'}"
								></div>
							{/each}
						</div>
					</div>

					<!-- Block Efficiency -->
					<div
						class="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 relative overflow-hidden"
					>
						<div class="absolute top-0 left-0 w-1 h-full bg-emerald-500"></div>
						<div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-4">
							Shield Integrity
						</div>
						<div class="text-4xl font-black text-slate-900 dark:text-white font-mono">99.9%</div>
						<div class="mt-4 text-[10px] text-emerald-500 font-mono">AUTOMATED DEFENSE ACTIVE</div>
					</div>

					<!-- Active Bans -->
					<div
						class="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 relative overflow-hidden"
					>
						<div class="absolute top-0 left-0 w-1 h-full bg-blue-500"></div>
						<div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-4">
							Neural Bans
						</div>
						<div class="text-4xl font-black text-slate-900 dark:text-white font-mono">
							{metrics.network.redeye_active_bans}
						</div>
						<div class="mt-4 text-[10px] text-blue-500 font-mono">ACTIVE REPUTATION QUARANTINE</div>
					</div>
				</div>

				<!-- Detailed Metrics Section -->
				<div class="space-y-6">
					<h3
						class="text-xs font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.3em] flex items-center gap-3"
					>
						<BarChart3 class="w-4 h-4 text-red-500" />
						Interception Analytics
					</h3>

					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<!-- Total Interceptions -->
						<div
							class="bg-white/50 dark:bg-slate-950/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-6"
						>
							<div class="flex justify-between items-start mb-6">
								<div>
									<h4
										class="text-sm font-bold text-slate-900 dark:text-white uppercase tracking-wider"
									>
										Total Interceptions
									</h4>
									<p class="text-[10px] text-slate-500 font-mono">Cumulative DENY operations</p>
								</div>
								<ShieldAlert class="w-5 h-5 text-red-500" />
							</div>
							<div class="text-3xl font-black text-slate-900 dark:text-white mb-2">
								{formatNumber(metrics.network.redeye_total_blocks)}
							</div>
							<div class="h-1.5 bg-slate-900 rounded-full overflow-hidden">
								<div class="h-full bg-red-500 w-[65%] animate-pulse"></div>
							</div>
						</div>

						<!-- Rate Limiting -->
						<div
							class="bg-white/50 dark:bg-slate-950/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-6"
						>
							<div class="flex justify-between items-start mb-6">
								<div>
									<h4
										class="text-sm font-bold text-slate-900 dark:text-white uppercase tracking-wider"
									>
										Throttling Events
									</h4>
									<p class="text-[10px] text-slate-500 font-mono">Traffic shaping triggers</p>
								</div>
								<Gauge class="w-5 h-5 text-amber-500" />
							</div>
							<div class="text-3xl font-black text-slate-900 dark:text-white mb-2">
								{formatNumber(metrics.network.redeye_total_rate_limit)}
							</div>
							<div class="h-1.5 bg-slate-900 rounded-full overflow-hidden">
								<div class="h-full bg-amber-500 w-[42%]"></div>
							</div>
						</div>
					</div>

					<!-- System Status -->
					<div class="bg-red-500/5 border border-red-500/10 rounded-2xl p-6">
						<div class="flex items-start gap-4">
							<div class="p-3 bg-red-500/10 rounded-xl">
								<Info class="w-6 h-6 text-red-400" />
							</div>
							<div class="flex-1">
								<h4
									class="text-sm font-bold text-slate-900 dark:text-white uppercase tracking-wider mb-2"
								>
									Automated Anomaly Detection
								</h4>
								<p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
									The RedEye neural core is currently monitoring for high-frequency traffic
									anomalies. Autonomous ban protocols are active with a threshold of <span
										class="text-red-400 font-mono">100 events/min</span
									>. Current system entropy is within nominal parameters.
								</p>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="p-6 border-t border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-950/80 flex justify-between items-center"
			>
				<div
					class="flex items-center gap-4 text-[10px] font-mono text-slate-500 uppercase tracking-widest"
				>
					<span class="flex items-center gap-1.5">
						<div class="w-1.5 h-1.5 rounded-full bg-red-500 animate-ping"></div>
						Core Online
					</span>
					<span>CRC: 0x8F22A1</span>
					<span>Lat: 0.24ms</span>
				</div>
				<button
					onclick={() => (showRedEyeModal = false)}
					class="px-8 py-3 rounded-xl bg-red-600 hover:bg-red-500 text-slate-900 dark:text-white text-[10px] font-black uppercase tracking-[0.2em] transition-all shadow-xl shadow-red-900/20"
				>
					Acknowledge
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(239, 68, 68, 0.1);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(239, 68, 68, 0.2);
	}
</style>
