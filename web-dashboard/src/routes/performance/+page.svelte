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
    subtitle="Telemetry Hub" 
    icon="ph:chart-line-up-bold"
>
    {#snippet actions()}
        <div class="flex bg-neutral-900 border-2 border-neutral-800 p-1 rounded-none shadow-2xl">
            <Button
                variant={autoRefresh ? 'primary' : 'secondary'}
                size="md"
                onclick={toggleAutoRefresh}
                class="min-w-[140px] !rounded-none"
            >
                {autoRefresh ? 'STREAM_LIVE' : 'LINK_PAUSED'}
            </Button>
            <Button
                variant="ghost"
                size="md"
                onclick={fetchMetrics}
                disabled={loading}
                loading={loading}
                icon="ph:arrows-clockwise-bold"
                class="!rounded-none"
            />
        </div>
    {/snippet}
</PageHeader>

<div class="space-y-10">
    <!-- Summary Metrics -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        {#if metrics}
            {#each [
                { label: 'Uptime_Persist', val: formatDuration(metrics.master.uptime_ms), icon: Clock, color: 'text-rust-light' },
                { label: 'Routines_Active', val: metrics.master.num_goroutine, icon: Activity, color: 'text-neutral-400' },
                { label: 'Heap_Allocation', val: formatBytes(metrics.master.heap_alloc), icon: MemoryStick, color: 'text-neutral-400' },
                { label: 'Network_Cycle', val: `${metrics.network.requests_per_second?.toFixed(1)} req/s`, icon: Zap, color: 'text-rust-light' },
                { label: 'Threat_Vector', val: formatNumber(metrics.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-red-500' },
                { label: 'Fault_Ratio', val: `${metrics.network.error_rate?.toFixed(2)}%`, icon: AlertCircle, color: metrics.network.error_rate > 5 ? 'text-red-500' : 'text-neutral-600' }
            ] as block}
                <div class="bg-neutral-900/60 border-2 border-neutral-800 rounded-none p-6 shadow-2xl hover:border-rust/30 transition-all flex flex-col justify-between min-h-[130px] group relative overflow-hidden">
                    <div class="corner-bracket-tl opacity-20"></div>
                    <div class="flex justify-between items-start relative z-10">
                        <span class="text-[8px] font-mono font-black text-neutral-500 uppercase tracking-[0.2em]">{block.label}</span>
                        <block.icon size={14} class={block.color} />
                    </div>
                    <div class="text-2xl font-heading font-black text-white tracking-tighter mt-4 italic relative z-10">{block.val}</div>
                </div>
            {/each}
        {/if}
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-12 gap-8">
        <!-- Memory Diagnostics -->
        <div class="xl:col-span-8">
            <Card title="Memory Diagnostics" subtitle="Allocation_Lifecycle" icon="ph:memory-bold">
                {#snippet actions()}
                    <div class="flex gap-2">
                        <Button variant="secondary" size="xs" onclick={forceGC} disabled={gcLoading} loading={gcLoading}>EXECUTE_GC</Button>
                        <Button variant="secondary" size="xs" onclick={freeMemory} disabled={freeMemLoading} loading={freeMemLoading}>RELEASE_MEM</Button>
                    </div>
                {/snippet}

                <div class="p-8 space-y-10">
                    <div class="space-y-4">
                        <div class="flex justify-between text-[9px] font-mono font-black uppercase tracking-[0.3em]">
                            <span class="text-neutral-500">Heap_State_Buffer</span>
                            <span class="text-rust-light">Limit: {formatBytes(metrics?.master.sys || 0)}</span>
                        </div>
                        <div class="h-2 bg-black border border-neutral-800 rounded-none overflow-hidden p-0 shadow-inner">
                            <div
                                class="h-full bg-rust transition-all duration-1000 ease-out shadow-[0_0_15px_#c2410c]"
                                style="width: {(metrics?.master.heap_usage_ratio || 0) * 100}%"
                            ></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                        {#each [
                            { label: 'Allocated', val: formatBytes(metrics?.master.heap_alloc || 0) },
                            { label: 'Stack_Size', val: formatBytes(metrics?.master.stack_sys || 0) },
                            { label: 'Live_Objects', val: formatNumber(metrics?.master.live_objects || 0) },
                            { label: 'Cycle_Rate', val: `${formatBytes(metrics?.master.heap_alloc_rate || 0)}/s` },
                            { label: 'Idle_Segment', val: formatBytes(metrics?.master.heap_idle || 0) },
                            { label: 'GC_Target', val: formatBytes(metrics?.master.next_gc_target || 0) }
                        ] as item}
                            <div class="bg-black/40 border border-neutral-800 p-5 rounded-none group hover:border-rust/20 transition-all shadow-inner relative">
                                <div class="text-[8px] font-mono font-black text-neutral-600 uppercase tracking-widest mb-2 group-hover:text-rust-light transition-colors">{item.label}</div>
                                <div class="text-xl font-heading font-black text-neutral-300 italic">{item.val}</div>
                            </div>
                        {/each}
                    </div>
                </div>
            </Card>
        </div>

        <!-- Network Operations -->
        <div class="xl:col-span-4">
            <Card title="Network Flow" subtitle="Traffic_Error_Vectors" icon="ph:globe-bold">
                <div class="p-8 space-y-8">
                    <div class="bg-black/60 border border-neutral-800 p-6 rounded-none shadow-inner relative overflow-hidden">
                        <div class="corner-bracket-tl opacity-10"></div>
                        <div class="flex justify-between items-end mb-3">
                            <span class="text-[9px] font-mono font-black text-neutral-500 uppercase tracking-[0.2em]">Active_Uplinks</span>
                            <span class="text-3xl font-heading font-black text-white italic">{metrics?.network.active_connections}</span>
                        </div>
                        <div class="w-full h-1 bg-neutral-900 overflow-hidden rounded-none">
                            <div class="h-full bg-rust-light animate-pulse shadow-[0_0_10px_#f97316]" style="width: 100%"></div>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-1">
                            <span class="text-[8px] font-mono font-black text-neutral-600 uppercase tracking-widest block">Data_Ingress</span>
                            <div class="text-xl font-heading font-black text-neutral-300 tabular-nums">{formatBytes(metrics?.network.bytes_received || 0)}</div>
                        </div>
                        <div class="space-y-1 text-right">
                            <span class="text-[8px] font-mono font-black text-neutral-600 uppercase tracking-widest block">Data_Egress</span>
                            <div class="text-xl font-heading font-black text-neutral-300 tabular-nums">{formatBytes(metrics?.network.bytes_sent || 0)}</div>
                        </div>
                    </div>

                    <div class="space-y-4 pt-6 border-t border-neutral-800">
                        <div class="flex justify-between items-center">
                            <span class="text-[9px] font-mono font-black text-neutral-500 uppercase tracking-widest">Fault_Ratio</span>
                            <span class="font-mono font-black text-rust-light">{(metrics?.network.error_rate || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-1.5 bg-black border border-neutral-800 rounded-none overflow-hidden">
                            <div 
                                class="h-full bg-rust shadow-[0_0_8px_#c2410c]" 
                                style="width: {Math.min(metrics?.network.error_rate || 0, 100)}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>

        <!-- Security & Persistence -->
        <div class="xl:col-span-6">
            <Card title="RedEye Sentinel" subtitle="Surveillance_Engine" icon="ph:shield-check-bold">
                {#snippet actions()}
                    <div class="px-3 py-1 bg-red-500/10 text-red-500 text-[9px] font-mono font-black uppercase rounded-none border border-red-500/20 animate-flicker">
                        ALERT_LEVEL: {metrics?.redeye?.threat_level || 'LOW'}
                    </div>
                {/snippet}

                <div class="p-8 space-y-10">
                    <div class="grid grid-cols-2 sm:grid-cols-4 gap-6">
                        {#each [
                            { label: 'Intercepts', val: formatNumber(metrics?.redeye?.total_blocks || 0), icon: ShieldAlert, color: 'text-red-500' },
                            { label: 'Rate_Caps', val: formatNumber(metrics?.redeye?.total_rate_limits || 0), icon: Activity, color: 'text-amber-500' },
                            { label: 'Active_Bans', val: formatNumber(metrics?.redeye?.active_bans || 0), icon: Ban, color: 'text-red-600' },
                            { label: 'Logic_Delay', val: `${metrics?.redeye?.avg_processing_time_ms.toFixed(2)}ms`, icon: Clock, color: 'text-neutral-500' }
                        ] as item}
                            <div class="flex flex-col items-center text-center space-y-3 group">
                                <div class="p-4 bg-black border border-neutral-800 rounded-none shadow-inner relative group-hover:border-rust/40 transition-colors">
                                    <item.icon size={20} class={item.color} />
                                </div>
                                <div class="space-y-1">
                                    <span class="text-[8px] font-mono font-black text-neutral-600 uppercase tracking-widest">{item.label}</span>
                                    <div class="text-xl font-heading font-black text-white italic">{item.val}</div>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <div class="p-6 bg-black border border-neutral-800 rounded-none relative overflow-hidden">
                        <div class="absolute top-0 left-0 w-1 h-full bg-red-600/40"></div>
                        <div class="flex items-center gap-3 mb-3 ml-2">
                            <div class="w-1 h-3 bg-red-600 shadow-[0_0_8px_#dc2626]"></div>
                            <span class="text-[9px] font-mono font-black text-red-500 uppercase tracking-[0.3em]">Neural_Mitigation_Active</span>
                        </div>
                        <p class="text-[10px] font-mono text-neutral-500 leading-relaxed font-black ml-2 uppercase tracking-tight">
                            Last tactical intercept confirmed at: <span class="text-neutral-300 font-bold">{metrics?.redeye?.last_block_at ? new Date(metrics.redeye.last_block_at).toLocaleTimeString() : 'NO_RECENT_EVENT'}</span>
                        </p>
                    </div>
                </div>
            </Card>
        </div>

        <div class="xl:col-span-6">
            <Card title="Persistence Core" subtitle="Database_Cluster" icon="ph:database-bold">
                <div class="p-8 space-y-10">
                    <div class="grid grid-cols-2 gap-6">
                        <div class="space-y-4">
                            <span class="text-[9px] font-mono font-black text-neutral-600 uppercase tracking-[0.2em] block">Pool_Saturation</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-black border border-neutral-800 p-4 rounded-none text-center shadow-inner">
                                    <div class="text-[7px] text-neutral-600 font-mono font-black uppercase mb-1">In_Use</div>
                                    <div class="text-xl font-heading font-black text-emerald-500 tabular-nums">{metrics?.database.in_use}</div>
                                </div>
                                <div class="bg-black border border-neutral-800 p-4 rounded-none text-center shadow-inner">
                                    <div class="text-[7px] text-neutral-600 font-mono font-black uppercase mb-1">Idle</div>
                                    <div class="text-xl font-heading font-black text-neutral-500 tabular-nums">{metrics?.database.idle}</div>
                                </div>
                            </div>
                        </div>
                        <div class="space-y-4">
                            <span class="text-[9px] font-mono font-black text-neutral-600 uppercase tracking-[0.2em] block">Commit_Cycle</span>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="bg-black border border-neutral-800 p-4 rounded-none text-center shadow-inner">
                                    <div class="text-[7px] text-neutral-600 font-mono font-black uppercase mb-1">Success</div>
                                    <div class="text-xl font-heading font-black text-neutral-300 tabular-nums">{formatNumber(metrics?.database.commits || 0)}</div>
                                </div>
                                <div class="bg-black border border-neutral-800 p-4 rounded-none text-center shadow-inner">
                                    <div class="text-[7px] text-neutral-600 font-mono font-black uppercase mb-1">Faults</div>
                                    <div class="text-xl font-heading font-black text-red-500 tabular-nums">{formatNumber(metrics?.database.rollbacks || 0)}</div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-4">
                        <div class="flex justify-between text-[9px] font-mono font-black uppercase tracking-[0.3em]">
                            <span class="text-neutral-600">Cache_Hit_Efficiency</span>
                            <span class="text-emerald-500 font-black italic">{(metrics?.database.cache_hit_ratio || 0).toFixed(2)}%</span>
                        </div>
                        <div class="h-2 bg-black border border-neutral-800 rounded-none overflow-hidden p-0 shadow-inner">
                            <div
                                class="h-full bg-emerald-500 transition-all duration-1000 ease-out shadow-[0_0_10px_#10b981]"
                                style="width: {metrics?.database.cache_hit_ratio || 0}%"
                            ></div>
                        </div>
                    </div>
                </div>
            </Card>
        </div>
    </div>

    <!-- Infrastructure Summary -->
    <div class="bg-neutral-900 border-2 border-neutral-800 p-8 rounded-none flex flex-wrap justify-center gap-12 text-[9px] font-mono font-black uppercase tracking-[0.3em] text-neutral-600 shadow-2xl relative">
        <div class="corner-bracket-tl opacity-20"></div>
        <div class="corner-bracket-br opacity-20"></div>
        
        <div class="flex items-center gap-3 group cursor-default">
            <Cpu size={14} class="text-rust-light opacity-40 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-neutral-300 transition-colors">Nodes: {metrics?.nodes.online_nodes || 0} Online</span>
        </div>
        <div class="flex items-center gap-3 group cursor-default">
            <Database size={14} class="text-rust-light opacity-40 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-neutral-300 transition-colors">Storage: {metrics?.database.size || 'NULL'}</span>
        </div>
        <div class="flex items-center gap-3 group cursor-default">
            <Activity size={14} class="text-rust-light opacity-40 group-hover:opacity-100 transition-opacity" />
            <span class="group-hover:text-neutral-300 transition-colors">IO_Status: NOMINAL</span>
        </div>
        <div class="flex items-center gap-3 text-rust border-b border-rust/20 pb-1">
            <RefreshCw size={14} class="animate-spin" />
            <span class="italic">Tactical_Link_Stable</span>
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
		background: #262626;
		border-radius: 0px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #c2410c;
	}
</style>