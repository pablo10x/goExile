<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { ResourceMetricsPanelProps } from '$lib/types/resource-metrics';
	import { useResourceMetrics } from '$lib/composables/useResourceMetrics';
	import ResourceStatsCard from '$lib/components/ResourceStatsCard.svelte';
	import ResourceHistoryChart from '$lib/components/ResourceHistoryChart.svelte';
	import { siteSettings } from '$lib/stores';
	import {
		RefreshCw,
		AlertTriangle,
		Cpu,
		HardDrive,
		Database,
		Activity,
		ChevronRight
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

	// Reactivity fix: Re-initialize metrics when IDs change
	let metrics = $derived(useResourceMetrics(spawnerId, instanceId));
	
	let stats = $derived(metrics.stats);
	let history = $derived(metrics.history);
	let peakStats = $derived(metrics.peakStats);
	let fetchedMemTotal = $derived(metrics.memTotal);
	let loading = $derived(metrics.loading);
	let error = $derived(metrics.error);

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

	$effect(() => {
		if (autoRefresh) {
			metrics.startPolling();
		}
		return () => metrics.stopPolling();
	});

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function getHealthStatus(percent: number) {
		if (percent < 50) return { color: 'emerald', label: 'NOMINAL', emoji: '✓' };
		if (percent < 75) return { color: 'yellow', label: 'MODERATE', emoji: '!' };
		if (percent < 90) return { color: 'orange', label: 'HIGH_LOAD', emoji: '⚠' };
		return { color: 'red', label: 'CRITICAL', emoji: '⨯' };
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

<div class="space-y-10 relative font-jetbrains">
	<!-- Background Grid -->
	<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>

	<!-- Enhanced Header -->
	{#if showTitle}
		<div class="relative" in:fly={{ y: -10, duration: 400 }}>
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
				<div class="flex items-center gap-5">
					<div class="relative">
						<div
							class="relative p-3 bg-rust/10 rounded-none border border-rust/30 shadow-lg shadow-rust/10"
							class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
							class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
						>
							<Activity class="w-6 h-6 text-rust" />
						</div>
					</div>
					<div>
						<h3
							class="text-2xl font-heading font-black text-white uppercase tracking-tighter flex items-center gap-4"
						>
							RESOURCE_TELEMETRY
							{#if $loading}
								<RefreshCw class="w-4 h-4 text-rust animate-spin" />
							{/if}
						</h3>
						<div class="flex items-center gap-3 mt-1.5">
							<ChevronRight class="w-3 h-3 text-rust" />
							<span class="text-[9px] font-black text-stone-500 uppercase tracking-[0.3em]">
								VECTOR: NODE_{spawnerId} // INSTANCE_{instanceId}
							</span>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-4">
					<!-- Time Range Selector -->
					<div class="flex bg-stone-950 p-1 border border-stone-800 shadow-inner"
						 class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
						 class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
					>
						{#each timeRanges as range}
							<button
								class="px-4 py-1.5 text-[9px] font-black uppercase tracking-widest transition-all
								{timeRange === range ? 'bg-rust text-white shadow-lg' : 'text-stone-600 hover:text-stone-300'}"
								onclick={() => (timeRange = range)}
							>
								{range}
							</button>
						{/each}
					</div>

					<!-- Refresh Button -->
					<button
						onclick={metrics.refresh}
						disabled={$loading}
						class="p-2.5 bg-stone-900 border border-stone-800 text-stone-500 hover:text-rust hover:border-rust transition-all active:translate-y-px shadow-lg"
						title="Refresh metrics"
					>
						<RefreshCw
							class="w-4 h-4 {$loading ? 'animate-spin' : ''}"
						/>
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Enhanced Error State -->
	{#if $error}
		<div class="p-6 bg-red-950/10 border border-red-900/30 rounded-none shadow-2xl flex items-start gap-5 industrial-frame" in:fade>
			<div class="p-2 bg-red-900/20 border border-red-500/30">
				<AlertTriangle class="w-6 h-6 text-red-500" />
			</div>
			<div>
				<h4 class="text-[11px] font-heading font-black text-red-500 uppercase tracking-widest mb-1">BUFFER_READ_FAULT</h4>
				<p class="text-[10px] text-stone-500 uppercase font-bold">{$error}</p>
			</div>
		</div>
	{/if}

	{#if !$error}
		<!-- Stats Cards Grid -->
		{#if $stats}
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				<ResourceStatsCard
					title="CORE_PROCESSOR_LOAD"
					current={$stats.cpu_percent}
					peak={$peakStats.peakCpu}
					unit="%"
					icon="CPU"
					color="orange"
					trend={cpuTrend > 0 ? 'up' : 'down'}
				/>
				<ResourceStatsCard
					title="MEMORY_RESERVATION"
					current={$stats.memory_usage / (1024 * 1024)}
					peak={$peakStats.peakMemory / (1024 * 1024)}
					unit="MB"
					icon="MEM"
					color="purple"
					trend={memTrend > 0 ? 'up' : 'down'}
				/>
				<ResourceStatsCard
					title="STORAGE_PERSISTENCE"
					current={$stats.disk_usage / (1024 * 1024)}
					peak={$peakStats.peakDisk / (1024 * 1024)}
					unit="MB"
					icon="DSK"
					color="emerald"
					trend="stable"
				/>
			</div>
		{:else if $loading}
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				{#each [0, 1, 2] as i}
					<div class="h-32 bg-stone-900/40 border border-stone-800 border-dashed industrial-frame animate-pulse"></div>
				{/each}
			</div>
		{/if}

		<!-- Performance Timeline Chart -->
		{#if !compact}
			<div class="space-y-6" in:fade={{ delay: 200 }}>
				<div class="flex justify-between items-end px-2">
					<div class="space-y-2">
						<div class="flex items-center gap-3 font-jetbrains text-[9px] font-black text-stone-600 uppercase tracking-[0.4em] italic">
							<Activity class="w-3.5 h-3.5 text-rust" />
							TEMPORAL_METRIC_BUFFER
						</div>
						<h3 class="text-[11px] font-heading font-black text-stone-300 uppercase tracking-widest">
							Stream_Analysis_{timeRange}
						</h3>
					</div>
					<div class="flex gap-6">
						<div class="flex items-center gap-3">
							<div class="w-2 h-2 bg-rust shadow-[0_0_8px_rgba(249,115,22,0.4)]"></div>
							<span class="text-[9px] font-black text-stone-500 uppercase tracking-widest">CORE</span>
						</div>
						<div class="flex items-center gap-3">
							<div class="w-2 h-2 bg-rust-dark shadow-[0_0_8px_rgba(124,45,18,0.4)]"></div>
							<span class="text-[9px] font-black text-stone-500 uppercase tracking-widest">MEMORY</span>
						</div>
					</div>
				</div>
				
				<div class="bg-stone-950 border border-stone-800 industrial-frame overflow-hidden relative shadow-inner p-8">
					<div class="absolute inset-0 bg-[url('/grid.svg')] opacity-[0.02] pointer-events-none"></div>
					<ResourceHistoryChart data={filteredHistory} {height} />
				</div>
			</div>
		{/if}
	{/if}
</div>