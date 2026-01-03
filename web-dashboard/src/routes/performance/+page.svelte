<script lang="ts">
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
		AlertTriangle
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
</script>

<PageHeader 
    title="Analytics" 
    subtitle="Performance Metrics" 
    icon="ph:chart-line-up-bold"
>
    {#snippet actions()}
        <div class="flex bg-slate-900/40 p-1.5 border border-slate-800 rounded-2xl shadow-inner backdrop-blur-md">
            <Button
                variant={autoRefresh ? 'primary' : 'secondary'}
                size="md"
                onclick={toggleAutoRefresh}
                class="min-w-[120px]"
            >
                {autoRefresh ? 'LIVE STREAM' : 'PAUSED'}
            </Button>
            <Button
                variant="ghost"
                size="md"
                onclick={fetchMetrics}
                disabled={loading}
                loading={loading}
                icon="ph:arrows-clockwise-bold"
            />
        </div>
    {/snippet}
</PageHeader>

<div class="space-y-10">
    <!-- Summary Metrics -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        {#if metrics}
            {#each [
                { label: 'System Uptime', val: formatDuration(metrics.master.uptime_ms), icon: Clock, color: 'text-white' },
                { label: 'Active Routines', val: metrics.master.num_goroutine, icon: Activity, color: 'text-white' },
                { label: 'Heap Allocation', val: formatBytes(metrics.master.heap_alloc), icon: MemoryStick, color: 'text-white' },
                { label: 'Request Rate', val: `${metrics.network.requests_per_second?.toFixed(1)} req/s`, icon: Zap, color: 'text-white' },
                { label: 'Threat Blocks', val: formatNumber(metrics.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-red-400' },
                { label: 'Error Rate', val: `${metrics.network.error_rate?.toFixed(2)}%`, icon: AlertCircle, color: metrics.network.error_rate > 5 ? 'text-red-400' : 'text-slate-400' }
            ] as block}
                <div class="bg-slate-900/40 border border-slate-800/50 rounded-2xl p-6 shadow-xl hover:border-indigo-500/30 transition-all flex flex-col justify-between min-h-[130px]">
                    <div class="flex justify-between items-start">
                        <span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest">{block.label}</span>
                        <block.icon size={16} class={block.color} />
                    </div>
                    <div class="text-2xl font-heading font-black text-white tracking-tight mt-4">{block.val}</div>
                </div>
            {/each}
        {/if}
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-12 gap-8">
        <!-- Memory Diagnostics -->
        <div class="xl:col-span-8">
            <Card title="Memory Diagnostics" subtitle="Allocation & Lifecycle" icon="ph:memory-bold">
                {#snippet actions()}
                    <div class="flex gap-2">
                        <Button variant="secondary" size="xs" onclick={forceGC} disabled={gcLoading} loading={gcLoading}>RUN GC</Button>
                        <Button variant="secondary" size="xs" onclick={freeMemory} disabled={freeMemLoading} loading={freeMemLoading}>FREE MEM</Button>
                    </div>
                {/snippet}

                <div class="p-8 space-y-10">
                    <div class="space-y-4">
                        <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest">
                            <span class="text-slate-500">Heap Usage</span>
                            <span class="text-white">Limit: {formatBytes(metrics?.master.sys || 0)}</span>
                        </div>
                        <div class="h-4 bg-slate-950 border border-slate-800 rounded-full overflow-hidden p-0.5 shadow-inner">
                            <div
                                class="h-full bg-indigo-600 rounded-full transition-all duration-1000 ease-out shadow-[0_0_10px_rgba(99,102,241,0.4)]"
                                style="width: {(metrics?.master.heap_usage_ratio || 0) * 100}%"
                            ></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                        {#each [
                            { label: 'Allocated', val: formatBytes(metrics?.master.heap_alloc || 0) },
                            { label: 'Stack size', val: formatBytes(metrics?.master.stack_sys || 0) },
                            { label: 'Live Objects', val: formatNumber(metrics?.master.live_objects || 0) },
                            { label: 'Allocation rate', val: `${formatBytes(metrics?.master.heap_alloc_rate || 0)}/s` },
                            { label: 'Idle heap', val: formatBytes(metrics?.master.heap_idle || 0) },
                            { label: 'GC target', val: formatBytes(metrics?.master.next_gc_target || 0) }
                        ] as item}
                            <div class="bg-slate-950/20 border border-slate-800/50 p-5 rounded-2xl group hover:border-indigo-500/20 transition-all shadow-sm">
                                <div class="text-[9px] font-bold text-slate-500 uppercase tracking-widest mb-2 group-hover:text-indigo-400 transition-colors">{item.label}</div>
                                <div class="text-xl font-heading font-black text-slate-200">{item.val}</div>
                            </div>
                        {/each}
                    </div>
                </div>
            </Card>
        </div>

        <!-- Network Operations -->
        <div class="xl:col-span-4">
            <Card title="Network Operations" subtitle="Traffic & Error Vectors" icon="ph:globe-bold">
                <div class="p-8 space-y-8">
                    <div class="bg-slate-950/40 border border-slate-800 p-6 rounded-2xl shadow-inner">
                        <div class="flex justify-between items-end mb-3">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Active Connections</span>
                            <span class="text-3xl font-heading font-black text-white">{metrics?.network.active_connections}</span>
                        </div>
                        <div class="w-full h-1 bg-slate-800 rounded-full overflow-hidden">
                            <div class="h-full bg-indigo-500 animate-pulse shadow-[0_0_8px_#6366f1]" style="width: 100%"></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-1">
                            <span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest block">Data Inbound</span>
                            <div class="text-xl font-heading font-black text-white">{formatBytes(metrics?.network.bytes_received || 0)}</div>
                        </div>
                        <div class="space-y-1 text-right">
                            <span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest block">Data Outbound</span>
                            <div class="text-xl font-heading font-black text-white">{formatBytes(metrics?.network.bytes_sent || 0)}</div>
                        </div>
                    </div>

                    <div class="space-y-4 pt-6 border-t border-slate-800/50">
                        <div class="flex justify-between items-center">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Error Rate</span>
                            <span class="font-black text-indigo-400">{(metrics?.network.error_rate || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-2 bg-slate-950 border border-slate-800 rounded-full overflow-hidden p-0.5">
                            <div 
                                class="h-full bg-indigo-600 rounded-full shadow-[0_0_8px_#6366f1]" 
                                style="width: {Math.min(metrics?.network.error_rate || 0, 100)}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>

        <!-- Security & Persistence -->
        <div class="xl:col-span-6">
            <Card title="Security Core" subtitle="Threat Monitoring" icon="ph:shield-check-bold">
                {#snippet actions()}
                    <div class="px-3 py-1 bg-red-500/10 text-red-400 text-[9px] font-black uppercase rounded-lg border border-red-500/20 animate-pulse">
                        Threat Level: {metrics?.redeye?.threat_level || 'LOW'}
                    </div>
                {/snippet}

                <div class="p-8 space-y-10">
                    <div class="grid grid-cols-2 sm:grid-cols-4 gap-6">
                        {#each [
                            { label: 'Intercepts', val: formatNumber(metrics?.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-red-400' },
                            { label: 'Rate Limits', val: formatNumber(metrics?.redeye?.total_rate_limits || 0), icon: Activity, color: 'text-amber-400' },
                            { label: 'Active Bans', val: formatNumber(metrics?.redeye?.active_bans || 0), icon: Ban, color: 'text-red-500' },
                            { label: 'Rule Latency', val: `${metrics?.redeye?.avg_processing_time_ms.toFixed(2)}ms`, icon: Clock, color: 'text-slate-400' }
                        ] as item}
                            <div class="flex flex-col items-center text-center space-y-3">
                                <div class="p-4 bg-slate-950 border border-slate-800 rounded-2xl shadow-inner relative group">
                                    <item.icon size={24} class={item.color} />
                                </div>
                                <div class="space-y-1">
                                    <span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest">{item.label}</span>
                                    <div class="text-xl font-heading font-black text-white">{item.val}</div>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <div class="p-6 bg-slate-950/40 border border-slate-800 rounded-2xl">
                        <div class="flex items-center gap-3 mb-3">
                            <div class="w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse"></div>
                            <span class="text-[10px] font-bold text-red-400 uppercase tracking-widest">Active Surveillance</span>
                        </div>
                        <p class="text-[11px] text-slate-400 leading-relaxed font-medium">
                            System-wide security protocols are active. Last block incident mitigated at: <span class="text-slate-200 font-bold">{metrics?.redeye?.last_block_at ? new Date(metrics.redeye.last_block_at).toLocaleTimeString() : 'None'}</span>
                        </p>
                    </div>
                </div>
            </Card>
        </div>

        <div class="xl:col-span-6">
            <Card title="Database Health" subtitle="Persistence & Pool" icon="ph:database-bold">
                <div class="p-8 space-y-10">
                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-4">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block">Connection Pool</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-slate-950/40 border border-slate-800 p-4 rounded-xl text-center">
                                    <div class="text-[8px] text-slate-500 font-bold uppercase mb-1">In Use</div>
                                    <div class="text-xl font-heading font-black text-emerald-400">{metrics?.database.in_use}</div>
                                </div>
                                <div class="bg-slate-950/40 border border-slate-800 p-4 rounded-xl text-center">
                                    <div class="text-[8px] text-slate-500 font-bold uppercase mb-1">Idle</div>
                                    <div class="text-xl font-heading font-black text-slate-400">{metrics?.database.idle}</div>
                                </div>
                            </div>
                        </div>
                        <div class="space-y-4">
                            <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block">Session Stats</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-slate-950/40 border border-slate-800 p-4 rounded-xl text-center">
                                    <div class="text-[8px] text-slate-500 font-bold uppercase mb-1">Commits</div>
                                    <div class="text-xl font-heading font-black text-white">{formatNumber(metrics?.database.commits || 0)}</div>
                                </div>
                                <div class="bg-slate-950/40 border border-slate-800 p-4 rounded-xl text-center">
                                    <div class="text-[8px] text-slate-500 font-bold uppercase mb-1">Rollbacks</div>
                                    <div class="text-xl font-heading font-black text-red-400">{formatNumber(metrics?.database.rollbacks || 0)}</div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-4">
                        <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest">
                            <span class="text-slate-500">Memory Cache Hit Ratio</span>
                            <span class="text-emerald-400 font-black">{(metrics?.database.cache_hit_ratio || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-4 bg-slate-950 border border-slate-800 rounded-full overflow-hidden p-0.5 shadow-inner">
                            <div
                                class="h-full bg-emerald-500 rounded-full transition-all duration-1000 ease-out shadow-[0_0_10px_rgba(16,185,129,0.4)]"
                                style="width: {metrics?.database.cache_hit_ratio || 0}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>
    </div>

    <!-- Infrastructure Summary -->
    <div class="bg-slate-900/40 border border-slate-800 p-8 rounded-3xl flex flex-wrap justify-center gap-12 text-[10px] font-bold uppercase tracking-widest text-slate-500 shadow-xl">
        <div class="flex items-center gap-3 group cursor-default">
            <Cpu size={16} class="text-indigo-400 opacity-50 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300">Nodes: {metrics?.nodes.online_nodes || 0} Online</span>
        </div>
        <div class="flex items-center gap-3 group cursor-default">
            <Database size={16} class="text-indigo-400 opacity-50 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300">Storage: {metrics?.database.size || 'N/A'}</span>
        </div>
        <div class="flex items-center gap-3 group cursor-default">
            <Activity size={16} class="text-indigo-400 opacity-50 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-slate-300">IO State: Nominal</span>
        </div>
        <div class="flex items-center gap-3 text-indigo-400 border-b border-indigo-500/20 pb-1">
            <RefreshCw size={16} class="animate-spin" />
            <span>Real-time Link Active</span>
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
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #334155;
	}
</style>
