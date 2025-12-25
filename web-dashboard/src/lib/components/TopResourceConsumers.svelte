<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type {
		TopResourceConsumersProps,
		ResourceConsumer,
		ResourceType
	} from '$lib/types/resource-metrics';
	import { getTopResourceConsumers } from '$lib/composables/useResourceMetrics';
	import { Cpu, MemoryStick, HardDrive, Activity, RefreshCw, AlertCircle } from 'lucide-svelte';

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

<div
	class="bg-black/40 backdrop-blur-md border border-stone-800 rounded-none p-8 industrial-frame shadow-2xl"
>
	<!-- Header -->
	<div class="flex items-center justify-between mb-8">
		<div class="flex items-center gap-5">
			<div
				class="p-3 bg-stone-950 border border-stone-800 industrial-frame text-rust"
			>
				<Icon class="w-6 h-6" />
			</div>
			<div>
				<h3 class="font-heading font-black text-white uppercase tracking-widest">
					Top_Resource_Load
				</h3>
				<p class="font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest mt-1">
					Metric: {resourceType === 'cpu' ? 'PROC_LOAD' : resourceType === 'memory' ? 'MEM_LOAD' : 'IO_DISK'}
				</p>
			</div>
		</div>

		<div class="flex items-center gap-4">
			{#if loading}
				<div class="animate-spin rounded-none h-5 w-5 border-2 border-rust border-t-transparent shadow-lg shadow-rust/20"></div>
			{:else if error}
				<div class="text-red-500 font-jetbrains font-black text-[10px] uppercase tracking-widest border border-red-900/30 px-3 py-1 bg-red-950/10 shadow-lg">FAULT_DETECTED</div>
			{:else}
				<button
					onclick={loadConsumers}
					class="p-2.5 text-stone-500 hover:text-white bg-stone-900/50 hover:bg-rust/10 border border-stone-800 hover:border-rust transition-all active:translate-y-px"
					title="Recalibrate"
				>
					<RefreshCw class="w-4 h-4" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="bg-red-950/20 border border-red-900/40 industrial-frame p-6 text-red-500 mb-6 shadow-2xl">
			<div class="flex items-center gap-4">
				<AlertCircle class="w-6 h-6 shrink-0 animate-pulse" />
				<div class="space-y-1">
					<span class="block font-heading font-black text-xs uppercase tracking-widest">SIGNAL_FAULT_0x09</span>
					<p class="font-jetbrains text-[10px] font-bold uppercase tracking-tight">{error}</p>
				</div>
			</div>
		</div>
	{/if}

	<!-- Consumers List -->
	{#if !loading && !error}
		{#if consumers.length === 0}
			<div class="text-center py-20 opacity-40">
				<div class="inline-block p-6 bg-stone-900/40 border border-dashed border-stone-800 industrial-frame mb-6">
					<Activity class="w-10 h-10 text-stone-700" />
				</div>
				<p class="text-stone-600 font-jetbrains font-black text-[11px] uppercase tracking-[0.3em]">Zero_Nodes_Mapped_To_Buffer</p>
			</div>
		{:else}
			<div class="space-y-4">
				{#each consumers as consumer, index (consumer.id)}
					<div
						class="group flex items-center justify-between p-4 bg-stone-900/20 border border-stone-800 hover:border-rust/30 transition-all duration-500 shadow-lg"
					>
						<!-- Instance Info -->
						<div class="flex items-center gap-5 flex-1 min-w-0">
							<div class="flex-shrink-0 border-r border-stone-800 pr-5">
								<div class="text-xs font-jetbrains font-black text-stone-400 tracking-tighter uppercase mb-1">
									#{consumer.port}
								</div>
								<div class="text-[9px] font-jetbrains font-bold text-stone-600 uppercase tracking-widest">
									{consumer.region}
								</div>
							</div>

							{#if !compact}
								<div class="min-w-0">
									<div class="text-[11px] font-black text-stone-200 uppercase tracking-widest truncate group-hover:text-rust transition-colors duration-500">
										Node_{consumer.instanceId.split('-').pop()}
									</div>
									<div class="text-[8px] font-jetbrains font-black text-stone-700 uppercase tracking-[0.2em] mt-1">{consumer.host}</div>
								</div>
							{/if}
						</div>

						<!-- Resource Usage -->
						<div class="flex items-center gap-6">
							<div class="text-right">
								<div
									class="text-2xl font-heading font-black tabular-nums {getColorForValue(
										consumer.cpu_percent,
										resourceType
									)} tracking-tighter leading-none"
								>
									{consumer.cpu_percent?.toFixed(1)}%
								</div>
								<div class="text-[8px] font-jetbrains font-black text-stone-600 uppercase tracking-widest mt-1">
									LOAD_INDEX
								</div>
							</div>

							<!-- Visual Indicator -->
							<div class="w-20 h-1.5 bg-stone-950 border border-stone-800 rounded-none overflow-hidden relative p-0 shadow-inner">
								<div
									class="h-full {getBackgroundForValue(
										consumer.cpu_percent,
										resourceType
									).replace('/10', '')} transition-all duration-1000 ease-out shadow-lg"
									style="width: {Math.min(consumer.cpu_percent, 100)}%"
								></div>

								<!-- Animated pulse for high usage -->
								{#if consumer.cpu_percent >= 80}
									<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
								{/if}
							</div>
						</div>

						<!-- Status -->
						<div class="flex items-center gap-3 border-l border-stone-800 pl-6 ml-6">
							<div
								class={`w-1.5 h-1.5 rounded-full ${consumer.status === 'Running'
									? 'bg-emerald-500 animate-pulse shadow-[0_0_8px_#10b981]'
									: 'bg-stone-700'}`}
							></div>
							<div class="text-[10px] font-heading font-black text-stone-500 uppercase tracking-widest group-hover:text-stone-300 transition-colors">
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
		<div class="space-y-4">
			{#each Array(limit) as _, index (index)}
				<div
					class="flex items-center justify-between p-4 bg-stone-900/40 border border-stone-800 industrial-frame"
				>
					<div class="flex items-center gap-5 flex-1">
						<div class="w-12 h-4 bg-stone-800 animate-pulse"></div>
						<div class="flex-1">
							<div class="h-3 bg-stone-800 w-3/4 animate-pulse mb-2"></div>
							<div class="h-2 bg-stone-800 w-1/2 animate-pulse opacity-50"></div>
						</div>
					</div>
					<div class="w-20 h-1.5 bg-stone-800 animate-pulse"></div>
				</div>
			{/each}
		</div>
	{/if}
</div>
