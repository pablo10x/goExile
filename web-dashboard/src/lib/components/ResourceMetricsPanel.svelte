<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut, elasticOut } from 'svelte/easing';
	import type { ResourceMetricsPanelProps } from '$lib/types/resource-metrics';
	import { useResourceMetrics } from '$lib/composables/useResourceMetrics';
	import ResourceStatsCard from '$lib/components/ResourceStatsCard.svelte';
	import ResourceHistoryChart from '$lib/components/ResourceHistoryChart.svelte';
	import {
		RefreshCw,
		AlertTriangle,
		Cpu,
		HardDrive,
		Database,
		TrendingUp,
		TrendingDown,
		Activity
	} from 'lucide-svelte';

	let {
		spawnerId,
		instanceId,
		memTotal = 0,
		diskTotal = 0,
		height = 250,
		compact = false,
		showTitle = true,
		autoRefresh = true
	}: ResourceMetricsPanelProps = $props();

	const {
		stats,
		history,
		peakStats,
		memTotal: fetchedMemTotal,
		loading,
		error,
		startPolling,
		stopPolling,
		refresh
	} = useResourceMetrics(spawnerId, instanceId);

	// Use passed total or fetched total
	let effectiveMemTotal = $derived(memTotal || $fetchedMemTotal || 0);

	let timeRange = $state('24h');
	const timeRanges = ['1h', '5h', '12h', '24h'];

	let filteredHistory = $derived(
		$history.filter((point) => {
			if (timeRange === '24h') return true;
			const pointTime = new Date(point.timestamp).getTime();
			const now = Date.now();
			const hours = parseInt(timeRange);
			const cutoff = now - hours * 60 * 60 * 1000;
			return pointTime >= cutoff;
		})
	);

	// Calculate trends
	let cpuTrend = $derived.by(() => {
		if ($history.length < 2) return 0;
		const recent = $history.slice(-10);
		const first = recent[0]?.cpu || 0;
		const last = recent[recent.length - 1]?.cpu || 0;
		return last - first;
	});

	let memTrend = $derived.by(() => {
		if ($history.length < 2) return 0;
		const recent = $history.slice(-10);
		const first = recent[0]?.memory_percent || 0;
		const last = recent[recent.length - 1]?.memory_percent || 0;
		return last - first;
	});

	onMount(() => {
		if (autoRefresh) {
			startPolling();
		}
	});

	onDestroy(() => {
		stopPolling();
	});

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function getHealthStatus(percent: number) {
		if (percent < 50) return { color: 'emerald', label: 'Healthy', emoji: '✓' };
		if (percent < 75) return { color: 'yellow', label: 'Moderate', emoji: '!' };
		if (percent < 90) return { color: 'orange', label: 'High', emoji: '⚠' };
		return { color: 'red', label: 'Critical', emoji: '⨯' };
	}

	let cpuHealth = $derived(getHealthStatus($stats?.cpu_percent ?? 0));

	let memPercent = $derived(
		effectiveMemTotal
			? Math.min(100, (($stats?.memory_usage ?? 0) / (effectiveMemTotal * 1024 * 1024)) * 100)
			: 0
	);
	let memHealth = $derived(getHealthStatus(memPercent));

	let diskPercent = $derived(
		diskTotal
			? Math.min(100, (($stats?.disk_usage ?? 0) / (diskTotal * 1024 * 1024 * 1024)) * 100)
			: 0
	);
	let diskHealth = $derived(getHealthStatus(diskPercent));
</script>

<div class="space-y-8 relative">
	<!-- Ambient background effects -->
	<div
		class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-transparent to-purple-500/5 rounded-3xl blur-3xl pointer-events-none"
	></div>

	<!-- Enhanced Header -->
	{#if showTitle}
		<div class="relative" in:fly={{ y: -20, duration: 600, easing: cubicOut }}>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-4">
					<div class="relative">
						<div
							class="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-2xl blur-xl"
						></div>
						<div
							class="relative p-3 bg-gradient-to-br from-slate-900 to-slate-800 rounded-2xl border border-slate-300/50 dark:border-slate-700/50 shadow-xl"
						>
							<Activity class="w-6 h-6 text-blue-400" />
						</div>
					</div>
					<div>
						<h3
							class="text-2xl font-bold bg-gradient-to-r from-slate-100 via-slate-200 to-slate-300 bg-clip-text text-transparent flex items-center gap-3"
						>
							Resource Monitor
							{#if $loading}
								<div class="relative">
									<div
										class="w-4 h-4 rounded-full border-2 border-slate-300 dark:border-slate-700 border-t-blue-500 animate-spin"
									></div>
									<div
										class="absolute inset-0 w-4 h-4 rounded-full bg-blue-500/20 blur-sm animate-pulse"
									></div>
								</div>
							{/if}
						</h3>
						<p class="text-sm text-slate-500 font-mono mt-1 flex items-center gap-2">
							<span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
							Instance {instanceId}
						</p>
					</div>
				</div>

				<div class="flex items-center gap-3">
					<!-- Time Range Selector -->
					<div class="relative group">
						<div
							class="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-xl blur-lg opacity-0 group-hover:opacity-100 transition-opacity"
						></div>
						<div
							class="relative flex bg-slate-900/80 backdrop-blur-xl rounded-xl p-1 border border-slate-300/50 dark:border-slate-700/50 shadow-lg"
						>
							{#each timeRanges as range, i}
								<button
									class="relative px-4 py-2 text-[11px] font-bold uppercase tracking-wider rounded-lg transition-all overflow-hidden {timeRange ===
									range
										? 'text-slate-900 dark:text-white'
										: 'text-slate-500 hover:text-slate-700 dark:text-slate-300'}"
									onclick={() => (timeRange = range)}
									in:scale={{ duration: 300, delay: i * 50, easing: elasticOut }}
								>
									{#if timeRange === range}
										<div
											class="absolute inset-0 bg-gradient-to-r from-blue-600 to-cyan-600 shadow-lg"
											transition:scale={{ duration: 200 }}
										></div>
										<div
											class="absolute inset-0 bg-gradient-to-r from-blue-400 to-cyan-400 opacity-50 blur-sm"
										></div>
									{/if}
									<span class="relative z-10">{range}</span>
								</button>
							{/each}
						</div>
					</div>

					<!-- Refresh Button -->
					<button
						onclick={refresh}
						disabled={$loading}
						class="relative group p-3 bg-slate-900/80 backdrop-blur-xl hover:bg-slate-800 rounded-xl transition-all disabled:opacity-50 border border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600 shadow-lg overflow-hidden"
						title="Refresh metrics"
					>
						<div
							class="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-purple-500/10 opacity-0 group-hover:opacity-100 transition-opacity"
						></div>
						<RefreshCw
							class="relative z-10 w-4 h-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-800 dark:text-slate-200 transition-colors {$loading
								? 'animate-spin'
								: ''}"
						/>
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Enhanced Error State -->
	{#if $error}
		<div class="relative group" in:fly={{ y: 20, duration: 400 }}>
			<div
				class="absolute inset-0 bg-gradient-to-r from-red-500/20 to-rose-500/20 rounded-2xl blur-xl"
			></div>
			<div
				class="relative bg-red-500/10 backdrop-blur-xl border border-red-500/30 rounded-2xl p-6 flex items-start gap-4 shadow-lg"
			>
				<div class="p-2 rounded-xl bg-red-500/20 border border-red-500/30">
					<AlertTriangle class="w-6 h-6 text-red-400" />
				</div>
				<div class="flex-1">
					<h4 class="text-sm font-bold text-red-300 mb-1">Error Fetching Metrics</h4>
					<p class="text-sm text-red-400/80">{$error}</p>
				</div>
			</div>
		</div>
	{/if}

	{#if !$error}
		<!-- Enhanced Stats Cards Grid with Custom Cards -->
		{#if $stats}
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				<!-- CPU Card -->
				<div class="relative group" in:fly={{ y: 30, duration: 500, delay: 0, easing: cubicOut }}>
					<div
						class="absolute inset-0 bg-gradient-to-br from-orange-500/20 to-red-500/20 rounded-2xl blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
					></div>
					<div
						class="relative bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-xl rounded-2xl p-6 border border-slate-300/50 dark:border-slate-700/50 hover:border-orange-500/50 transition-all duration-300 shadow-xl overflow-hidden"
					>
						<!-- Animated background gradient -->
						<div
							class="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-red-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
						></div>

						<!-- Header -->
						<div class="relative z-10 flex items-start justify-between mb-4">
							<div class="flex items-center gap-3">
								<div class="relative">
									<div class="absolute inset-0 bg-orange-500/20 rounded-xl blur-lg"></div>
									<div
										class="relative p-2.5 bg-orange-500/10 rounded-xl border border-orange-500/30"
									>
										<Cpu class="w-5 h-5 text-orange-400" />
									</div>
								</div>
								<div>
									<div
										class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
									>
										CPU Usage
									</div>
									<div class="flex items-center gap-2 mt-1">
										{#if cpuTrend > 1}
											<TrendingUp class="w-3 h-3 text-red-400" />
											<span class="text-[10px] text-red-400 font-mono">+{cpuTrend.toFixed(1)}%</span
											>
										{:else if cpuTrend < -1}
											<TrendingDown class="w-3 h-3 text-emerald-400" />
											<span class="text-[10px] text-emerald-400 font-mono"
												>{cpuTrend.toFixed(1)}%</span
											>
										{/if}
									</div>
								</div>
							</div>
							<div
								class="px-2 py-1 rounded-lg bg-{cpuHealth.color}-500/10 border border-{cpuHealth.color}-500/30 text-{cpuHealth.color}-400 text-[10px] font-bold"
							>
								{cpuHealth.emoji}
								{cpuHealth.label}
							</div>
						</div>

						<!-- Current Value -->
						<div class="relative z-10 mb-4">
							<div
								class="text-4xl font-bold bg-gradient-to-r from-orange-400 to-red-400 bg-clip-text text-transparent font-mono tracking-tight"
							>
								{$stats.cpu_percent?.toFixed(1)}<span class="text-2xl">%</span>
							</div>
							<div class="text-xs text-slate-500 mt-1">Peak: {$peakStats.peakCpu?.toFixed(1)}%</div>
						</div>

						<!-- Enhanced Progress Bar -->
						<div class="relative z-10">
							<div
								class="relative w-full h-2.5 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
							>
								<!-- Animated background shimmer -->
								<div
									class="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent animate-shimmer"
									style="background-size: 200% 100%;"
								></div>

								<!-- Progress -->
								<div
									class="absolute inset-y-0 left-0 bg-gradient-to-r from-orange-500 via-orange-400 to-red-500 rounded-full transition-all duration-700 ease-out shadow-lg"
									style="width: {$stats.cpu_percent}%; box-shadow: 0 0 20px rgba(249, 115, 22, 0.5), 0 0 40px rgba(249, 115, 22, 0.2);"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-white/20 to-transparent rounded-full"
									></div>
								</div>
							</div>
							<!-- Markers -->
							<div class="flex justify-between text-[9px] text-slate-600 font-mono mt-1.5">
								<span>0%</span>
								<span>50%</span>
								<span>100%</span>
							</div>
						</div>

						<!-- Sparkline mini chart indicator -->
						<div class="absolute bottom-0 right-0 w-24 h-16 opacity-10 pointer-events-none">
							<svg viewBox="0 0 100 50" class="w-full h-full">
								<path
									d="M 0 50 L 10 45 L 20 40 L 30 35 L 40 38 L 50 30 L 60 25 L 70 28 L 80 20 L 90 15 L 100 10"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="text-orange-500"
								/>
							</svg>
						</div>
					</div>
				</div>

				<!-- Memory Card -->
				<div class="relative group" in:fly={{ y: 30, duration: 500, delay: 100, easing: cubicOut }}>
					<div
						class="absolute inset-0 bg-gradient-to-br from-purple-500/20 to-pink-500/20 rounded-2xl blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
					></div>
					<div
						class="relative bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-xl rounded-2xl p-6 border border-slate-300/50 dark:border-slate-700/50 hover:border-purple-500/50 transition-all duration-300 shadow-xl overflow-hidden"
					>
						<!-- Animated background gradient -->
						<div
							class="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-pink-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
						></div>

						<!-- Header -->
						<div class="relative z-10 flex items-start justify-between mb-4">
							<div class="flex items-center gap-3">
								<div class="relative">
									<div class="absolute inset-0 bg-purple-500/20 rounded-xl blur-lg"></div>
									<div
										class="relative p-2.5 bg-purple-500/10 rounded-xl border border-purple-500/30"
									>
										<Database class="w-5 h-5 text-purple-400" />
									</div>
								</div>
								<div>
									<div
										class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
									>
										Memory
									</div>
									<div class="flex items-center gap-2 mt-1">
										{#if memTrend > 1}
											<TrendingUp class="w-3 h-3 text-red-400" />
											<span class="text-[10px] text-red-400 font-mono">+{memTrend.toFixed(1)}%</span
											>
										{:else if memTrend < -1}
											<TrendingDown class="w-3 h-3 text-emerald-400" />
											<span class="text-[10px] text-emerald-400 font-mono"
												>{memTrend.toFixed(1)}%</span
											>
										{/if}
									</div>
								</div>
							</div>
							<div
								class="px-2 py-1 rounded-lg bg-{memHealth.color}-500/10 border border-{memHealth.color}-500/30 text-{memHealth.color}-400 text-[10px] font-bold"
							>
								{memHealth.emoji}
								{memHealth.label}
							</div>
						</div>

						<!-- Current Value -->
						<div class="relative z-10 mb-4">
							<div
								class="text-4xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent font-mono tracking-tight"
							>
								{formatBytes($stats.memory_usage)}
							</div>
							<div class="text-xs text-slate-500 mt-1">
								Peak: {formatBytes($peakStats.peakMemory)}
								{#if effectiveMemTotal}
									<span class="text-slate-600 mx-1">•</span>
									<span>{memPercent.toFixed(1)}%</span>
								{/if}
							</div>
						</div>

						<!-- Enhanced Progress Bar -->
						<div class="relative z-10">
							<div
								class="relative w-full h-2.5 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
							>
								<div
									class="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent animate-shimmer"
									style="background-size: 200% 100%;"
								></div>

								<div
									class="absolute inset-y-0 left-0 bg-gradient-to-r from-purple-500 via-purple-400 to-pink-500 rounded-full transition-all duration-700 ease-out shadow-lg"
									style="width: {memPercent}%; box-shadow: 0 0 20px rgba(168, 85, 247, 0.5), 0 0 40px rgba(168, 85, 247, 0.2);"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-white/20 to-transparent rounded-full"
									></div>
								</div>
							</div>
							<div class="flex justify-between text-[9px] text-slate-600 font-mono mt-1.5">
								<span>0%</span>
								<span>50%</span>
								<span>100%</span>
							</div>
						</div>

						<div class="absolute bottom-0 right-0 w-24 h-16 opacity-10 pointer-events-none">
							<svg viewBox="0 0 100 50" class="w-full h-full">
								<path
									d="M 0 40 L 10 38 L 20 35 L 30 30 L 40 32 L 50 28 L 60 25 L 70 22 L 80 20 L 90 18 L 100 15"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="text-purple-500"
								/>
							</svg>
						</div>
					</div>
				</div>

				<!-- Disk Card -->
				<div class="relative group" in:fly={{ y: 30, duration: 500, delay: 200, easing: cubicOut }}>
					<div
						class="absolute inset-0 bg-gradient-to-br from-emerald-500/20 to-teal-500/20 rounded-2xl blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
					></div>
					<div
						class="relative bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-xl rounded-2xl p-6 border border-slate-300/50 dark:border-slate-700/50 hover:border-emerald-500/50 transition-all duration-300 shadow-xl overflow-hidden"
					>
						<div
							class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-teal-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
						></div>

						<!-- Header -->
						<div class="relative z-10 flex items-start justify-between mb-4">
							<div class="flex items-center gap-3">
								<div class="relative">
									<div class="absolute inset-0 bg-emerald-500/20 rounded-xl blur-lg"></div>
									<div
										class="relative p-2.5 bg-emerald-500/10 rounded-xl border border-emerald-500/30"
									>
										<HardDrive class="w-5 h-5 text-emerald-400" />
									</div>
								</div>
								<div>
									<div
										class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest"
									>
										Disk Space
									</div>
									<div class="text-[10px] text-slate-600 mt-1 font-mono">Storage</div>
								</div>
							</div>
							<div
								class="px-2 py-1 rounded-lg bg-{diskHealth.color}-500/10 border border-{diskHealth.color}-500/30 text-{diskHealth.color}-400 text-[10px] font-bold"
							>
								{diskHealth.emoji}
								{diskHealth.label}
							</div>
						</div>

						<!-- Current Value -->
						<div class="relative z-10 mb-4">
							<div
								class="text-4xl font-bold bg-gradient-to-r from-emerald-400 to-teal-400 bg-clip-text text-transparent font-mono tracking-tight"
							>
								{formatBytes($stats.disk_usage)}
							</div>
							<div class="text-xs text-slate-500 mt-1">
								Peak: {formatBytes($peakStats.peakDisk)}
								{#if diskTotal}
									<span class="text-slate-600 mx-1">•</span>
									<span>{diskPercent.toFixed(1)}%</span>
								{/if}
							</div>
						</div>

						<!-- Enhanced Progress Bar -->
						<div class="relative z-10">
							<div
								class="relative w-full h-2.5 bg-slate-800/80 rounded-full overflow-hidden backdrop-blur-sm"
							>
								<div
									class="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent animate-shimmer"
									style="background-size: 200% 100%;"
								></div>

								<div
									class="absolute inset-y-0 left-0 bg-gradient-to-r from-emerald-500 via-emerald-400 to-teal-500 rounded-full transition-all duration-700 ease-out shadow-lg"
									style="width: {diskPercent}%; box-shadow: 0 0 20px rgba(16, 185, 129, 0.5), 0 0 40px rgba(16, 185, 129, 0.2);"
								>
									<div
										class="absolute inset-0 bg-gradient-to-r from-white/20 to-transparent rounded-full"
									></div>
								</div>
							</div>
							<div class="flex justify-between text-[9px] text-slate-600 font-mono mt-1.5">
								<span>0%</span>
								<span>50%</span>
								<span>100%</span>
							</div>
						</div>

						<div class="absolute bottom-0 right-0 w-24 h-16 opacity-10 pointer-events-none">
							<svg viewBox="0 0 100 50" class="w-full h-full">
								<path
									d="M 0 45 L 10 44 L 20 43 L 30 42 L 40 41 L 50 40 L 60 39 L 70 38 L 80 37 L 90 36 L 100 35"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="text-emerald-500"
								/>
							</svg>
						</div>
					</div>
				</div>
			</div>
		{:else if $loading}
			<!-- Enhanced Skeleton Loading -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				{#each [0, 1, 2] as i}
					<div class="relative" in:scale={{ duration: 400, delay: i * 100, easing: elasticOut }}>
						<div
							class="absolute inset-0 bg-gradient-to-r from-slate-800/30 to-slate-700/30 rounded-2xl blur-xl animate-pulse"
						></div>
						<div
							class="relative h-48 bg-gradient-to-br from-slate-900/50 to-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-300/50 dark:border-slate-700/50 overflow-hidden"
						>
							<div
								class="absolute inset-0 bg-gradient-to-r from-transparent via-slate-700/20 to-transparent animate-shimmer"
								style="background-size: 200% 100%;"
							></div>
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Enhanced Performance Timeline Chart -->
		{#if !compact}
			<div class="relative group" in:fly={{ y: 30, duration: 600, delay: 300, easing: cubicOut }}>
				<div
					class="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
				></div>
				<div
					class="relative bg-gradient-to-br from-slate-900/90 to-slate-800/90 backdrop-blur-xl rounded-2xl border border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600/50 p-6 shadow-xl transition-all duration-300 overflow-hidden"
				>
					<!-- Animated background -->
					<div
						class="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-transparent to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
					></div>

					<!-- Header -->
					<div class="relative z-10 flex items-center justify-between mb-6">
						<div class="flex items-center gap-3">
							<div class="relative">
								<div class="absolute inset-0 bg-blue-500/20 rounded-xl blur-lg"></div>
								<div class="relative p-2 bg-blue-500/10 rounded-xl border border-blue-500/30">
									<Activity class="w-4 h-4 text-blue-400" />
								</div>
							</div>
							<div>
								<h4
									class="text-sm font-bold text-slate-800 dark:text-slate-200 uppercase tracking-widest"
								>
									Performance Timeline
								</h4>
								<p class="text-xs text-slate-500 mt-0.5">
									Real-time resource metrics over {timeRange}
								</p>
							</div>
						</div>

						<!-- Legend -->
						<div
							class="flex items-center gap-4 px-4 py-2 bg-slate-900/50 backdrop-blur-sm rounded-xl border border-slate-300/50 dark:border-slate-700/50"
						>
							<div class="flex items-center gap-2">
								<div
									class="w-3 h-3 rounded-full bg-gradient-to-r from-orange-500 to-red-500 shadow-lg shadow-orange-500/50"
								></div>
								<span class="text-xs font-mono text-slate-500 dark:text-slate-400">CPU</span>
							</div>
							<div class="flex items-center gap-2">
								<div
									class="w-3 h-3 rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 shadow-lg shadow-blue-500/50"
								></div>
								<span class="text-xs font-mono text-slate-500 dark:text-slate-400">Memory</span>
							</div>
						</div>
					</div>

					<!-- Chart Container -->
					<div class="relative z-10">
						<ResourceHistoryChart data={filteredHistory} {height} />
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	@keyframes shimmer {
		0% {
			background-position: -200% center;
		}
		100% {
			background-position: 200% center;
		}
	}

	.animate-shimmer {
		animation: shimmer 3s linear infinite;
	}
</style>
