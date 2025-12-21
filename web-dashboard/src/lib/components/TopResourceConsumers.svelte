<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type {
		TopResourceConsumersProps,
		ResourceConsumer,
		ResourceType
	} from '$lib/types/resource-metrics';
	import { getTopResourceConsumers } from '$lib/composables/useResourceMetrics';
	import { Cpu, MemoryStick, HardDrive, Activity, RefreshCw } from 'lucide-svelte';

	let {
		limit = 10,
		refreshInterval = 5000,
		resourceType = 'cpu',
		compact = false
	}: TopResourceConsumersProps = $props();

	let consumers = $state<ResourceConsumer[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let refreshTimer: number | null = null;

	let Icon = $derived(getIconForType(resourceType));

	function getIconForType(type: ResourceType) {
		const iconMap = {
			cpu: Cpu,
			memory: MemoryStick,
			disk: HardDrive
		};
		return iconMap[type] || Activity;
	}

	function getColorForValue(value: number, type: ResourceType) {
		const thresholds: Record<ResourceType, { high: number; medium: number }> = {
			cpu: { high: 80, medium: 60 },
			memory: { high: 85, medium: 70 },
			disk: { high: 90, medium: 75 }
		};

		const threshold = thresholds[type] || thresholds.cpu;
		if (value >= threshold.high) return 'text-red-400';
		if (value >= threshold.medium) return 'text-yellow-400';
		return 'text-green-400';
	}

	function getBackgroundForValue(value: number, type: ResourceType) {
		const thresholds: Record<ResourceType, { high: number; medium: number }> = {
			cpu: { high: 80, medium: 60 },
			memory: { high: 85, medium: 70 },
			disk: { high: 90, medium: 75 }
		};

		const threshold = thresholds[type] || thresholds.cpu;
		if (value >= threshold.high) return 'bg-red-500/10';
		if (value >= threshold.medium) return 'bg-yellow-500/10';
		return 'bg-green-500/10';
	}

	async function loadConsumers() {
		try {
			loading = true;
			error = null;
			const data = await getTopResourceConsumers(limit, resourceType);
			consumers = data;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load resource consumers';
		} finally {
			loading = false;
		}
	}

	function startAutoRefresh() {
		if (refreshInterval > 0) {
			refreshTimer = setInterval(loadConsumers, refreshInterval) as unknown as number;
		}
	}

	function stopAutoRefresh() {
		if (refreshTimer) {
			clearInterval(refreshTimer);
			refreshTimer = null;
		}
	}

	onMount(() => {
		loadConsumers();
		startAutoRefresh();
	});

	onDestroy(() => {
		stopAutoRefresh();
	});
</script>

<div class="bg-slate-800/50 backdrop-blur-md border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center gap-3">
			<div class="p-2 bg-slate-900/50 rounded-lg border border-slate-300/50 dark:border-slate-700/50 text-slate-500 dark:text-slate-400">
				<Icon class="w-6 h-6" />
			</div>
			<div>
				<h3 class="text-lg font-semibold text-slate-800 dark:text-slate-200">Top Resource Consumers</h3>
				<p class="text-sm text-slate-500 dark:text-slate-400">
					{resourceType === 'cpu' ? 'CPU' : resourceType === 'memory' ? 'Memory' : 'Disk'} Usage
				</p>
			</div>
		</div>

		<div class="flex items-center gap-2">
			{#if loading}
				<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
			{:else if error}
				<div class="text-red-400 text-sm">Error</div>
			{:else}
				<button
					onclick={loadConsumers}
					class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 rounded-lg transition-colors"
					title="Refresh"
				>
					<RefreshCw class="w-4 h-4" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-300 mb-4">
			<div class="flex items-center gap-3">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="w-5 h-5"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="8" x2="12" y2="12"></line>
					<line x1="12" y1="16" x2="12.01" y2="16"></line>
				</svg>
				<span>{error}</span>
			</div>
		</div>
	{/if}

	<!-- Consumers List -->
	{#if !loading && !error}
		{#if consumers.length === 0}
			<div class="text-center py-8 text-slate-500 dark:text-slate-400">
				<div class="text-lg mb-2">📊</div>
				<div>No active instances found</div>
			</div>
		{:else}
			<div class="space-y-3">
				{#each consumers as consumer, index (consumer.id)}
					<div
						class="group flex items-center justify-between p-3 bg-slate-900/30 rounded-lg border border-slate-300/50 dark:border-slate-700/50 hover:border-slate-600/50 transition-all duration-300"
					>
						<!-- Instance Info -->
						<div class="flex items-center gap-3 flex-1 min-w-0">
							<div class="flex-shrink-0">
								<div class="text-sm font-mono text-slate-500 dark:text-slate-400">
									#{consumer.port}
								</div>
								<div class="text-xs text-slate-500">
									{consumer.region} • {consumer.host}
								</div>
							</div>

							{#if !compact}
								<div class="ml-3">
									<div class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">
										Instance-{consumer.instanceId}
									</div>
								</div>
							{/if}
						</div>

						<!-- Resource Usage -->
						<div class="flex items-center gap-4">
							<div class="text-right">
								<div
									class="text-lg font-bold tabular-nums {getColorForValue(
										consumer.cpu_percent,
										resourceType
									)}"
								>
									{(consumer.cpu_percent)?.toFixed(1)}%
								</div>
								<div class="text-xs text-slate-500 dark:text-slate-400">
									{resourceType === 'cpu' ? 'CPU' : resourceType === 'memory' ? 'Memory' : 'Disk'}
								</div>
							</div>

							<!-- Visual Indicator -->
							<div class="w-16 h-8 bg-slate-700/50 rounded-full overflow-hidden relative">
								<div
									class="h-full {getBackgroundForValue(
										consumer.cpu_percent,
										resourceType
									)} transition-all duration-500"
									style="width: {Math.min(consumer.cpu_percent, 100)}%"
								></div>

								<!-- Animated pulse for high usage -->
								{#if consumer.cpu_percent >= 80}
									<div class="absolute inset-0 bg-red-500/20 animate-pulse"></div>
								{/if}
							</div>
						</div>

						<!-- Status -->
						<div class="flex items-center gap-2">
							<div
								class="w-2 h-2 rounded-full {consumer.status === 'Running'
									? 'bg-emerald-500'
									: 'bg-slate-600'}"
							></div>
							<div class="text-xs text-slate-500 dark:text-slate-400">
								{consumer.status}
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="space-y-3">
			{#each Array(limit) as _, index (index)}
				<div
					class="flex items-center justify-between p-3 bg-slate-900/30 rounded-lg border border-slate-300/50 dark:border-slate-700/50"
				>
					<div class="flex items-center gap-3 flex-1">
						<div class="w-12 h-4 bg-slate-700/50 rounded animate-pulse"></div>
						<div class="flex-1">
							<div class="h-3 bg-slate-700/50 rounded w-3/4 animate-pulse mb-1"></div>
							<div class="h-3 bg-slate-700/50 rounded w-1/2 animate-pulse"></div>
						</div>
					</div>
					<div class="w-16 h-8 bg-slate-700/50 rounded-full animate-pulse"></div>
				</div>
			{/each}
		</div>
	{/if}
</div>
