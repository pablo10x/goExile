<script lang="ts">
import { apiFetch } from "$lib/api";
	import { onMount, onDestroy } from 'svelte';
	import {
		Zap,
		Activity,
		MemoryStick,
		HardDrive,
		Server,
		RefreshCw,
		AlertCircle,
		CheckCircle,
		Clock,
		Trash2,
		Signal,
		Terminal,
		Globe,
		Database,
		Lock,
		Cpu,
		Radio,
		Dna,
		AlertOctagon,
		ShieldAlert,
		Ban,
		AlertTriangle,
		ChevronRight
	} from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { siteSettings } from '$lib/stores.svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import Button from '$lib/components/Button.svelte';

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

	interface NodeDetail {
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

	interface NodeMetrics {
		total_nodes: number;
		online_nodes: number;
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
		node_details: NodeDetail[];
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

	interface RedEyeMetrics {
		total_blocks: number;
		total_rate_limits: number;
		active_bans: number;
		total_rules: number;
		avg_processing_time_ms: number;
		threat_level: 'low' | 'moderate' | 'high' | 'critical';
		last_block_at: string;
	}

	interface CombinedMetrics {
		master: RuntimeMetrics;
		nodes: NodeMetrics;
		database: DatabaseMetrics;
		network: NetworkMetrics;
		redeye: RedEyeMetrics;
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

	// API calls
	async function fetchMetrics() {
		try {
			const res = await apiFetch('/api/metrics');
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
			await apiFetch('/api/metrics/gc', { method: 'POST' });
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
			await apiFetch('/api/metrics/memory/free', { method: 'POST' });
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
</script>

<PageHeader 
    title="Analytics" 
    subtitle="System Monitoring" 
    icon="ph:chart-line-up-bold"
>
    {#snippet actions()}
        <div class="flex bg-slate-900/50 border border-white/5 p-1 rounded-xl shadow-lg backdrop-blur-md">
            <button
                onclick={toggleAutoRefresh}
                class="px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wide transition-all {autoRefresh ? 'bg-sky-500 text-white shadow-md' : 'text-slate-400 hover:text-slate-200'}"
            >
                {autoRefresh ? 'Live Updates' : 'Updates Paused'}
            </button>
            <Button
                variant="ghost"
                size="sm"
                onclick={fetchMetrics}
                disabled={loading}
                loading={loading}
                icon="ph:arrows-clockwise-bold"
                class="!rounded-lg"
            />
        </div>
    {/snippet}
</PageHeader>

<div class="space-y-8">
    <!-- Summary Metrics -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        {#if metrics}
            {#each [
                { label: 'System Uptime', val: formatDuration(metrics.master.uptime_ms), icon: Clock, color: 'text-sky-400' },
                { label: 'Active Routines', val: metrics.master.num_goroutine, icon: Activity, color: 'text-slate-400' },
                { label: 'Memory Allocated', val: formatBytes(metrics.master.heap_alloc), icon: MemoryStick, color: 'text-slate-400' },
                { label: 'Request Rate', val: `${metrics.network.requests_per_second?.toFixed(1)}/s`, icon: Zap, color: 'text-sky-400' },
                { label: 'Security Blocks', val: formatNumber(metrics.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-rose-500' },
                { label: 'Error Rate', val: `${metrics.network.error_rate?.toFixed(2)}%`, icon: AlertCircle, color: metrics.network.error_rate > 5 ? 'text-rose-500' : 'text-slate-500' }
            ] as block}
                <div class="bg-slate-800/40 border border-white/5 rounded-2xl p-5 shadow-lg backdrop-blur-md hover:border-sky-500/30 transition-all flex flex-col justify-between min-h-[120px] group relative overflow-hidden">
                    <div class="flex justify-between items-start relative z-10">
                        <span class="text-[11px] font-bold uppercase tracking-wider text-slate-500">{block.label}</span>
                        <block.icon size={16} class={block.color} />
                    </div>
                    <div class="text-2xl font-sans font-semibold text-slate-100 tracking-tight mt-2 relative z-10">{block.val}</div>
                </div>
            {/each}
        {/if}
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-12 gap-6 lg:gap-8">
        <!-- Memory -->
        <div class="xl:col-span-8">
            <Card title="Memory Usage" subtitle="Resource Details" icon="ph:memory-bold">
                {#snippet actions()}
                    <div class="flex gap-2">
                        <Button variant="secondary" size="xs" onclick={forceGC} disabled={gcLoading} loading={gcLoading}>Run GC</Button>
                        <Button variant="secondary" size="xs" onclick={freeMemory} disabled={freeMemLoading} loading={freeMemLoading}>Free Memory</Button>
                    </div>
                {/snippet}

                <div class="p-6 lg:p-8 space-y-8">
                    <div class="space-y-3">
                        <div class="flex justify-between text-xs font-medium text-slate-400">
                            <span>Memory Allocation</span>
                            <span class="text-sky-400">Total: {formatBytes(metrics?.master.sys || 0)}</span>
                        </div>
                        <div class="h-2.5 bg-slate-900/50 border border-white/5 rounded-full overflow-hidden p-0">
                            <div
                                class="h-full bg-gradient-to-r from-sky-500 to-indigo-500 transition-all duration-1000 ease-out rounded-full shadow-lg shadow-sky-500/20"
                                style="width: {(metrics?.master.heap_usage_ratio || 0) * 100}%"
                            ></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                        {#each [
                            { label: 'Heap Allocated', val: formatBytes(metrics?.master.heap_alloc || 0) },
                            { label: 'Stack Size', val: formatBytes(metrics?.master.stack_sys || 0) },
                            { label: 'Live Objects', val: formatNumber(metrics?.master.live_objects || 0) },
                            { label: 'Allocation Rate', val: `${formatBytes(metrics?.master.heap_alloc_rate || 0)}/s` },
                            { label: 'Idle Memory', val: formatBytes(metrics?.master.heap_idle || 0) },
                            { label: 'GC Threshold', val: formatBytes(metrics?.master.next_gc_target || 0) }
                        ] as item}
                            <div class="bg-slate-900/50 border border-white/5 p-4 rounded-xl group hover:border-sky-500/20 transition-all relative">
                                <div class="text-[10px] font-bold uppercase tracking-wider text-slate-500 mb-1 group-hover:text-sky-400 transition-colors">{item.label}</div>
                                <div class="text-lg font-semibold text-slate-200">{item.val}</div>
                            </div>
                        {/each}
                    </div>
                </div>
            </Card>
        </div>

        <!-- Network -->
        <div class="xl:col-span-4">
            <Card title="Network Traffic" subtitle="Data Overview" icon="ph:globe-bold">
                <div class="p-6 lg:p-8 space-y-8">
                    <div class="bg-slate-900/50 border border-white/5 p-6 rounded-2xl relative overflow-hidden">
                        <div class="flex justify-between items-end mb-4">
                            <span class="text-[11px] font-bold uppercase tracking-wider text-slate-500">Active Connections</span>
                            <span class="text-3xl font-semibold text-white">{metrics?.network.active_connections}</span>
                        </div>
                        <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                            <div class="h-full bg-emerald-500 animate-pulse rounded-full" style="width: 100%"></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-1">
                            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500 block">Data Received</span>
                            <div class="text-xl font-semibold text-slate-200 tabular-nums">{formatBytes(metrics?.network.bytes_received || 0)}</div>
                        </div>
                        <div class="space-y-1 text-right">
                            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500 block">Data Sent</span>
                            <div class="text-xl font-semibold text-slate-200 tabular-nums">{formatBytes(metrics?.network.bytes_sent || 0)}</div>
                        </div>
                    </div>

                    <div class="space-y-3 pt-6 border-t border-white/5">
                        <div class="flex justify-between items-center">
                            <span class="text-[11px] font-bold uppercase tracking-wider text-slate-500">Error Rate</span>
                            <span class="font-medium text-rose-400">{(metrics?.network.error_rate || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-1.5 bg-slate-900/50 border border-white/5 rounded-full overflow-hidden">
                            <div 
                                class="h-full bg-rose-500 rounded-full" 
                                style="width: {Math.min(metrics?.network.error_rate || 0, 100)}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>

        <!-- Security -->
        <div class="xl:col-span-6">
            <Card title="Security Monitor" subtitle="Active Protection" icon="ph:shield-check-bold">
                {#snippet actions()}
                    <div class="px-3 py-1 bg-rose-500/10 text-rose-400 text-[10px] font-bold uppercase rounded-lg border border-rose-500/20">
                        Status: {metrics?.redeye?.threat_level || 'Normal'}
                    </div>
                {/snippet}

                <div class="p-6 lg:p-8 space-y-8">
                    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                        {#each [
                            { label: 'Blocked', val: formatNumber(metrics?.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-rose-500' },
                            { label: 'Rate Limited', val: formatNumber(metrics?.redeye?.total_rate_limits || 0), icon: Activity, color: 'text-amber-500' },
                            { label: 'Active Bans', val: formatNumber(metrics?.redeye?.active_bans || 0), icon: Ban, color: 'text-rose-600' },
                            { label: 'Avg Latency', val: `${metrics?.redeye?.avg_processing_time_ms.toFixed(2)}ms`, icon: Clock, color: 'text-slate-500' }
                        ] as item}
                            <div class="flex flex-col items-center text-center space-y-3 group">
                                <div class="p-3 bg-slate-900/50 border border-white/5 rounded-xl group-hover:border-sky-500/30 transition-colors">
                                    <item.icon size={20} class={item.color} />
                                </div>
                                <div class="space-y-0.5">
                                    <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">{item.label}</span>
                                    <div class="text-lg font-semibold text-slate-200">{item.val}</div>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <div class="p-5 bg-slate-900/50 border border-white/5 rounded-xl relative overflow-hidden">
                        <div class="absolute top-0 left-0 w-1 h-full bg-rose-500/50"></div>
                        <div class="flex items-center gap-2 mb-2 ml-2">
                            <div class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"></div>
                            <span class="text-[10px] font-bold text-rose-400 uppercase tracking-wider">Protection Enabled</span>
                        </div>
                        <p class="text-xs font-medium text-slate-400 ml-2">
                            Last security event: <span class="text-slate-200">{metrics?.redeye?.last_block_at ? new Date(metrics.redeye.last_block_at).toLocaleTimeString() : 'None'}</span>
                        </p>
                    </div>
                </div>
            </Card>
        </div>

        <div class="xl:col-span-6">
            <Card title="Database Status" subtitle="Connection Details" icon="ph:database-bold">
                <div class="p-6 lg:p-8 space-y-8">
                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-3">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider block">Connection Pool</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-slate-900/50 border border-white/5 p-3 rounded-xl text-center">
                                    <div class="text-[9px] text-slate-500 font-bold uppercase mb-1">In Use</div>
                                    <div class="text-lg font-semibold text-emerald-400 tabular-nums">{metrics?.database.in_use}</div>
                                </div>
                                <div class="bg-slate-900/50 border border-white/5 p-3 rounded-xl text-center">
                                    <div class="text-[9px] text-slate-500 font-bold uppercase mb-1">Idle</div>
                                    <div class="text-lg font-semibold text-slate-400 tabular-nums">{metrics?.database.idle}</div>
                                </div>
                            </div>
                        </div>
                        <div class="space-y-3">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider block">Transactions</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-slate-900/50 border border-white/5 p-3 rounded-xl text-center">
                                    <div class="text-[9px] text-slate-500 font-bold uppercase mb-1">Success</div>
                                    <div class="text-lg font-semibold text-slate-200 tabular-nums">{formatNumber(metrics?.database.commits || 0)}</div>
                                </div>
                                <div class="bg-slate-900/50 border border-white/5 p-3 rounded-xl text-center">
                                    <div class="text-[9px] text-slate-500 font-bold uppercase mb-1">Errors</div>
                                    <div class="text-lg font-semibold text-rose-500 tabular-nums">{formatNumber(metrics?.database.rollbacks || 0)}</div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-3">
                        <div class="flex justify-between text-xs font-medium">
                            <span class="text-slate-400">Cache Hit Efficiency</span>
                            <span class="text-emerald-400">{(metrics?.database.cache_hit_ratio || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-2.5 bg-slate-900/50 border border-white/5 rounded-full overflow-hidden p-0">
                            <div
                                class="h-full bg-emerald-500 transition-all duration-1000 ease-out rounded-full shadow-lg shadow-emerald-500/20"
                                style="width: {metrics?.database.cache_hit_ratio || 0}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>
    </div>

    <!-- System Summary -->
    <div class="bg-slate-900/50 border border-white/5 p-6 rounded-2xl flex flex-wrap justify-center gap-8 text-xs font-medium text-slate-500 shadow-lg backdrop-blur-md">
        <div class="flex items-center gap-2.5 group cursor-default">
            <Cpu size={16} class="text-sky-400 opacity-60 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300 transition-colors">Nodes: {metrics?.nodes.online_nodes || 0} Online</span>
        </div>
        <div class="flex items-center gap-2.5 group cursor-default">
            <Database size={16} class="text-sky-400 opacity-60 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300 transition-colors">Storage: {metrics?.database.size || 'N/A'}</span>
        </div>
        <div class="flex items-center gap-2.5 group cursor-default">
            <Activity size={16} class="text-sky-400 opacity-60 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300 transition-colors">Status: Normal</span>
        </div>
        <div class="flex items-center gap-2.5 text-sky-400 border-b border-sky-500/20 pb-0.5">
            <RefreshCw size={14} class="animate-spin" />
            <span>Connection Stable</span>
        </div>
    </div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1e293b;
		border-radius: 99px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569;
	}
</style>