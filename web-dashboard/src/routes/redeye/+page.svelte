<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, slide, scale } from 'svelte/transition';
	import Icon from '$lib/components/theme/Icon.svelte';
	import {
		Eye,
		Plus,
		Trash2,
		Edit2,
		RefreshCw,
		ShieldAlert,
		Activity,
		Zap,
		Ban,
		ShieldCheck,
		Save,
		Unlock,
		Terminal,
		Cpu,
		Globe,
		Lock,
		AlertTriangle,
		ChevronRight,
		Monitor,
		BarChart3,
		Settings,
		X
	} from 'lucide-svelte';

	interface RedEyeRule {
		id: number;
		name: string;
		cidr: string;
		port: string;
		path_pattern: string;
		protocol: string;
		action: 'ALLOW' | 'DENY' | 'RATE_LIMIT';
		rate_limit: number;
		burst: number;
		enabled: boolean;
	}

	interface RedEyeLog {
		id: number;
		source_ip: string;
		action: string;
		timestamp: string;
	}

	interface AnticheatEvent {
		id: number;
		event_type: string;
		client_ip: string;
		severity: number;
		timestamp: string;
		player_id: string;
	}

	interface RedEyeStats {
		total_rules: number;
		active_bans: number;
		events_24h: number;
		logs_24h: number;
		reputation_count: number;
		entropy: number;
		threat_level: string;
		uptime: string;
		system_active: boolean;
		system_error: string;
		node_id: string;
		crc: string;
	}

	interface RedEyeConfig {
		'redeye.auto_ban_enabled': boolean;
		'redeye.auto_ban_threshold': number;
		'redeye.alert_enabled': boolean;
	}

	interface BannedIP {
		ip: string;
		reputation_score: number;
		ban_reason: string;
		ban_expires_at: string | null;
		last_seen: string;
	}

	let activeTab = $state<'overview' | 'rules' | 'bans' | 'logs' | 'anticheat' | 'config'>(
		'overview'
	);
	let rules = $state<RedEyeRule[]>([]);
	let logs = $state<RedEyeLog[]>([]);
	let events = $state<AnticheatEvent[]>([]);
	let bans = $state<BannedIP[]>([]);
	let stats = $state<RedEyeStats>({
		total_rules: 0,
		active_bans: 0,
		events_24h: 0,
		logs_24h: 0,
		reputation_count: 0,
		entropy: 0.0042,
		threat_level: 'Low',
		uptime: '99.99%',
		system_active: false,
		system_error: '',
		node_id: 'MASTER_RE_01',
		crc: '0x8F2A11'
	});
	let config = $state<RedEyeConfig>({
		'redeye.auto_ban_enabled': true,
		'redeye.auto_ban_threshold': 100,
		'redeye.alert_enabled': true
	});
	let loading = $state(true);

	// Modal
	let showModal = $state(false);
	let editingRule = $state<RedEyeRule | null>(null);
	let form = $state({
		name: '',
		cidr: '',
		port: '*',
		path_pattern: '',
		protocol: 'ANY',
		action: 'DENY' as 'ALLOW' | 'DENY' | 'RATE_LIMIT',
		rate_limit: 0,
		burst: 0,
		enabled: true
	});

	async function fetchStats() {
		try {
			const res = await fetch('/api/redeye/stats');
			if (res.ok) stats = await res.json();
		} catch (e) {
			console.error(e);
		}
	}

	async function fetchConfig() {
		try {
			const res = await fetch('/api/redeye/config');
			if (res.ok) config = await res.json();
		} catch (e) {
			console.error(e);
		}
	}

	async function updateConfig() {
		try {
			const res = await fetch('/api/redeye/config', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(config)
			});
			if (res.ok) {
				// alert('Configuration updated successfully');
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function fetchRules() {
		loading = true;
		try {
			const res = await fetch('/api/redeye/rules');
			if (res.ok) rules = await res.json();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function fetchLogs() {
		loading = true;
		try {
			const res = await fetch('/api/redeye/logs?limit=100');
			if (res.ok) {
				const data = await res.json();
				logs = data.logs;
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function fetchEvents() {
		loading = true;
		try {
			const res = await fetch('/api/redeye/anticheat/events?limit=50');
			if (res.ok) {
				const data = await res.json();
				events = data.events;
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function fetchBans() {
		loading = true;
		try {
			const res = await fetch('/api/redeye/bans');
			if (res.ok) bans = await res.json();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function unbanIP(ip: string) {
		try {
			const res = await fetch(`/api/redeye/bans/${ip}`, { method: 'DELETE' });
			if (res.ok) {
				bans = bans.filter((b) => b.ip !== ip);
				fetchStats();
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function refreshAll() {
		fetchStats();
		if (activeTab === 'rules') fetchRules();
		else if (activeTab === 'logs') fetchLogs();
		else if (activeTab === 'anticheat') fetchEvents();
		else if (activeTab === 'bans') fetchBans();
		else if (activeTab === 'config') fetchConfig();
	}

	async function saveRule() {
		try {
			const url = editingRule ? `/api/redeye/rules/${editingRule.id}` : '/api/redeye/rules';
			const method = editingRule ? 'PUT' : 'POST';

			const payload = { ...form };

			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});

			if (res.ok) {
				showModal = false;
				fetchRules();
				fetchStats();
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteRule(id: number) {
		try {
			const res = await fetch(`/api/redeye/rules/${id}`, { method: 'DELETE' });
			if (res.ok) {
				rules = rules.filter((r) => r.id !== id);
				fetchStats();
			}
		} catch (e) {
			console.error(e);
		}
	}

	function openModal(rule: RedEyeRule | null = null) {
		editingRule = rule;
		if (rule) {
			form.name = rule.name;
			form.cidr = rule.cidr;
			form.port = rule.port;
			form.path_pattern = rule.path_pattern;
			form.protocol = rule.protocol;
			form.action = rule.action;
			form.rate_limit = rule.rate_limit;
			form.burst = rule.burst;
			form.enabled = rule.enabled;
		} else {
			form = {
				name: '',
				cidr: '',
				port: '*',
				path_pattern: '',
				protocol: 'ANY',
				action: 'DENY',
				rate_limit: 10,
				burst: 20,
				enabled: true
			};
		}
		showModal = true;
	}

	function getActionColor(action: string) {
		switch (action) {
			case 'ALLOW':
				return 'text-success bg-success/10 border-success/30';
			case 'DENY':
				return 'text-danger bg-danger/10 border-danger/30';
			case 'RATE_LIMIT':
				return 'text-warning bg-warning/10 border-warning/30';
			default:
				return 'text-text-dim bg-stone-800 border-stone-700';
		}
	}

	onMount(() => {
		fetchStats();
	});
</script>

<div class="w-full h-full flex flex-col overflow-hidden relative font-sans">
	<!-- Tech Background elements -->
	<div class="absolute inset-0 pointer-events-none opacity-5">
		<div class="absolute top-0 left-0 w-full h-full bg-[url('/grid.svg')]"></div>
		<div
			class="absolute top-0 left-0 w-full h-[1px] bg-danger animate-[scan_4s_linear_infinite]"
		></div>
	</div>

	<!-- Header -->
	<div class="flex flex-col lg:flex-row lg:justify-between lg:items-center mb-8 lg:mb-10 gap-8 shrink-0 relative z-10">
		<div class="flex items-center gap-4 sm:gap-6">
			<div class="relative">
				<div class="absolute inset-0 bg-danger blur-2xl opacity-20 animate-pulse"></div>
				<div class="relative bg-stone-950 border border-red-500/50 p-3 sm:p-4 rounded-none industrial-frame shadow-2xl">
					<Icon name="ph:eye-bold" size="2rem" class="text-danger animate-flicker" />
				</div>
			</div>
			<div class="min-w-0">
				<div class="flex flex-wrap items-center gap-3 sm:gap-4">
					<h1 class="text-2xl sm:text-4xl font-heading font-black tracking-tighter text-white uppercase leading-none">
						RedEye_<span class="text-danger">Security</span>
					</h1>
					<div
						class="px-2 py-0.5 sm:px-3 sm:py-1 {stats.system_active
							? 'bg-danger text-black shadow-red-500/50 shadow-lg'
							: 'bg-stone-800 text-stone-400'} text-[8px] sm:text-[10px] font-black uppercase tracking-widest"
					>
						{stats.system_active ? 'Active' : 'Offline'}
					</div>
				</div>
				<p class="font-jetbrains text-text-dim text-[8px] sm:text-[10px] tracking-[0.2em] sm:tracking-[0.3em] uppercase mt-2 font-black truncate">
					Threat Mitigation Core // {stats.system_error
						? `Fault: ${stats.system_error}`
						: 'Status: Nominal'}
				</p>
			</div>
		</div>

		<div class="flex items-center justify-between sm:justify-end gap-6 w-full lg:w-auto">
			<div class="flex flex-col items-start sm:items-end sm:mr-6 text-left sm:text-right">
				<span class="font-jetbrains text-[8px] sm:text-[9px] text-text-dim font-black uppercase tracking-[0.3em]"
					>SYSTEM_ENTROPY_DELTA</span
				>
				<span class="text-base sm:text-lg text-danger font-jetbrains font-black tracking-widest">{stats.entropy?.toFixed(4) || '0.0000'}%</span>
			</div>
			<button
				onclick={refreshAll}
				class="p-2.5 sm:p-3 bg-stone-900 border border-stone-800 hover:border-red-500/50 text-text-dim hover:text-danger transition-all shadow-xl active:tranneutral-y-px"
				aria-label="Refresh security data"
			>
				<RefreshCw class="w-5 h-5 sm:w-6 sm:h-6 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- System Status Banner -->
	<div
		class="mb-8 p-4 sm:p-6 bg-[var(--header-bg)]/80 backdrop-blur-md border-y border-stone-800 flex flex-nowrap overflow-x-auto no-scrollbar gap-8 sm:gap-12 items-center relative"
	>
		<div
			class="absolute inset-y-0 left-0 w-1 {stats.system_active ? 'bg-danger shadow-[0_0_15px_#ef4444]' : 'bg-stone-700'} shrink-0"
		></div>
		<div class="flex items-center gap-3 sm:gap-4 shrink-0">
			<div
				class="w-2 sm:w-2.5 h-2 sm:h-2.5 rounded-full {stats.system_active
					? 'bg-danger animate-pulse shadow-[0_0_10px_#ef4444]'
					: 'bg-stone-700'}"
			></div>
			<span class="font-jetbrains text-[9px] sm:text-[10px] font-black text-text-dim uppercase tracking-[0.3em] whitespace-nowrap"
				>LIVE_MONITORING: <span
					class={stats.system_active ? 'text-white' : 'text-stone-700'}
					>{stats.system_active ? 'ENGAGED' : 'OFFLINE'}</span
				></span
			>
		</div>
		<div class="flex items-center gap-3 font-jetbrains text-[9px] sm:text-[10px] font-black uppercase tracking-widest shrink-0 whitespace-nowrap">
			<span class="text-text-dim">UPTIME_STABILITY:</span>
			<span class="text-danger">{stats.uptime}</span>
		</div>
		<div class="flex items-center gap-3 font-jetbrains text-[9px] sm:text-[10px] font-black uppercase tracking-widest shrink-0 whitespace-nowrap">
			<span class="text-text-dim">THREAT_SPECTRUM:</span>
			<span
				class={`px-2 py-0.5 border ${stats.threat_level === 'Low'
					? 'text-success border-success/20 bg-success/5'
					: stats.threat_level === 'Medium'
						? 'text-warning border-warning/20 bg-warning/5'
						: 'text-danger border-red-500/20 bg-danger/5'}`}>{stats.threat_level?.toUpperCase()}</span>
		</div>
		<div class="ml-auto font-jetbrains text-[8px] sm:text-[9px] font-black text-stone-700 hidden lg:flex gap-6 uppercase tracking-widest shrink-0 whitespace-nowrap">
			<span>CRC_HASH: {stats.crc}</span>
			<div class="w-px h-3 bg-stone-800"></div>
			<span>NODE_UPLINK: {stats.node_id}</span>
		</div>
	</div>

	<!-- Navigation Controls -->
	<div
		class="flex overflow-x-auto no-scrollbar gap-1.5 mb-8 shrink-0 bg-[var(--header-bg)] p-1.5 border border-stone-800 industrial-frame"
	>
		{#each [['overview', BarChart3, 'Overview'], ['rules', ShieldCheck, 'Rules'], ['bans', Ban, 'Bans'], ['anticheat', AlertTriangle, 'Intel'], ['logs', Terminal, 'Logs'], ['config', Settings, 'Settings']] as [id, icon, label]}
			{@const Icon = icon as any}
			<button
				onclick={() => {
					activeTab = id as any;
					refreshAll();
				}}
				class="flex-1 min-w-[80px] sm:min-w-[120px] flex items-center justify-center gap-2 sm:gap-3 py-3 sm:py-3.5 transition-all font-heading font-black text-[9px] sm:text-[10px] uppercase tracking-widest {activeTab ===
				id
					? 'bg-danger text-white shadow-xl shadow-red-900/30'
					: 'text-text-dim hover:text-stone-300 hover:bg-stone-900'}"
			>
				<Icon class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
				<span class="inline-block">{label}</span>
			</button>
		{/each}
	</div>

	<!-- Main Display -->
	<div class="flex-1 min-h-0 relative z-10">
		{#if activeTab === 'overview'}
			<div class="h-full overflow-auto space-y-8 custom-scrollbar pr-0 sm:pr-4">
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
					{#each [{ icon: ShieldCheck, val: stats.total_rules, label: 'Protocols', color: 'text-success' }, { icon: Ban, val: stats.active_bans, label: 'Quarantined', color: 'text-danger' }, { icon: ShieldAlert, val: stats.events_24h, label: 'Anomalies', color: 'text-warning' }, { icon: Activity, val: stats.logs_24h, label: 'Neural_Scans', color: 'text-sky-500' }] as card}
						{@const CardIcon = card.icon as any}
						<div
							class="modern-industrial-card glass-panel p-4 sm:p-6 flex items-center gap-4 sm:gap-6 group !rounded-none"
						>
							<div
								class="p-3 sm:p-4 bg-stone-950 border border-stone-800 group-hover:border-stone-600 transition-all shadow-lg industrial-frame"
							>
								<CardIcon class="w-6 h-6 sm:w-7 sm:h-7 {card.color}" />
							</div>
							<div class="flex flex-col gap-1">
								<span
									class="text-2xl sm:text-3xl font-heading font-black text-white tabular-nums leading-none tracking-tighter"
									>{card.val}</span
								>
								<span class="font-jetbrains text-[8px] sm:text-[9px] font-black text-text-dim uppercase tracking-widest mt-1"
									>{card.label}</span
								>
							</div>
						</div>
					{/each}
				</div>

				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8">
					<!-- Visual Threat Monitor -->
					<div
						class="lg:col-span-2 modern-industrial-card glass-panel p-6 sm:p-12 relative overflow-hidden flex flex-col items-center justify-center min-h-[350px] sm:min-h-[450px] !rounded-none"
					>
						<div class="absolute inset-0 bg-[url('/grid.svg')] opacity-[0.03]"></div>
						<div
							class="absolute inset-0 bg-gradient-to-t from-red-900/5 via-transparent to-transparent"
						></div>

						<div class="relative z-10 w-full max-w-2xl text-center">
							<div
								class="inline-flex items-center gap-3 px-3 py-1 sm:px-4 sm:py-1.5 bg-danger/10 border border-danger/30 rounded-none mb-6 sm:mb-10 industrial-frame"
							>
								<div class="w-1.5 sm:w-2 h-1.5 sm:h-2 rounded-full bg-danger animate-pulse shadow-red-500/50 shadow-lg"></div>
								<span class="font-jetbrains text-[8px] sm:text-[10px] font-black text-danger uppercase tracking-[0.2em] sm:tracking-[0.3em]"
									>Neural_Surveillance_Active</span>
							</div>

							<div class="relative mx-auto w-48 h-48 sm:w-64 sm:h-64 mb-8 sm:mb-12">
								<div
									class="absolute inset-0 border-[2px] sm:border-[3px] border-stone-800 rounded-full"
								></div>
								<div
									class="absolute inset-0 border-t-[2px] sm:border-t-[3px] border-red-500 rounded-full animate-[spin_4s_linear_infinite] shadow-[0_0_20px_rgba(239,68,68,0.3)]"
								></div>
								<div
									class="absolute inset-4 sm:inset-6 border border-stone-800 rounded-full border-dashed opacity-40"
								></div>
								<div class="absolute inset-0 flex items-center justify-center">
									<Activity class="w-12 h-12 sm:w-20 sm:h-20 text-danger animate-pulse drop-shadow-[0_0_15px_rgba(239,68,68,0.5)]" />
								</div>
							</div>

							<h3
								class="text-xl sm:text-3xl font-heading font-black text-white uppercase tracking-tighter"
							>
								AUTONOMOUS_NEURAL_DEFENSE
							</h3>
							<p class="font-jetbrains text-text-dim text-[9px] sm:text-[11px] font-bold mt-4 leading-relaxed uppercase tracking-tight max-w-xl mx-auto">
								System patrols regional ingress points. Latency: <span
									class="text-danger font-black">1.2ms_delta</span
								>. Monitoring
								<span class="text-white font-black"
									>{stats.reputation_count}</span
								> identity signatures global.
							</p>
						</div>
					</div>

					<!-- Quick Tech Stats -->
					<div class="space-y-6">
						<div
							class="modern-industrial-card glass-panel p-6 sm:p-8 !rounded-none h-full flex flex-col justify-between"
						>
							<div class="space-y-6 sm:space-y-8">
								<h4
									class="font-heading font-black text-[9px] sm:text-[11px] text-text-dim uppercase tracking-[0.3em] mb-4 sm:mb-6 flex items-center gap-3"
								>
									<Cpu class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-danger" /> LOAD_METRICS
								</h4>
								<div class="space-y-6 sm:space-y-8">
									{#each [['Heuristic', Number(Math.min(100, 5 + stats.logs_24h / 100)), 'text-success', 'bg-success'], ['Sequencer', Number(Math.min(100, 2 + stats.active_bans * 2)), 'text-danger', 'bg-danger'], ['Indexer', Number(Math.min(100, 1 + stats.reputation_count / 10)), 'text-rust', 'bg-rust']] as [label, perc, colorText, colorBg]}
										<div class="space-y-2 sm:space-y-3">
											<div class="flex justify-between font-jetbrains text-[8px] sm:text-[9px] font-black uppercase tracking-widest">
												<span class="text-text-dim">{label}</span>
												<span class={colorText as string}
													>{Math.round(perc as number)}%</span>
											</div>
											<div class="h-1 bg-stone-950 border border-stone-800 p-0 relative overflow-hidden">
												<div
													class="h-full {colorBg} transition-all duration-1000 shadow-lg"
													style="width: {perc}%;"
												></div>
											</div>
										</div>
									{/each}
								</div>
							</div>

							<div class="mt-8 sm:mt-10 pt-6 sm:pt-8 border-t border-stone-800">
								<h4
									class="font-heading font-black text-[9px] sm:text-[11px] text-text-dim uppercase tracking-[0.3em] mb-4 sm:mb-6 flex items-center gap-3"
								>
									<Globe class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-rust" /> TOPOLOGY
								</h4>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-4 sm:gap-5">
										<div class="text-3xl sm:text-5xl font-heading font-black text-white tabular-nums tracking-tighter">08</div>
										<div class="font-jetbrains text-[8px] sm:text-[9px] font-black text-text-dim leading-tight uppercase tracking-widest">
											ACTIVE<br />Neural_Zones
										</div>
									</div>
									<div class="flex gap-1 sm:gap-1.5">
										{#each Array(8) as _, i}
											<div
												class="h-4 sm:h-6 w-0.5 sm:w-1 bg-danger shadow-[0_0_8px_rgba(239,68,68,0.3)]"
												style="opacity: {0.2 + i * 0.1}"
											></div>
										{/each}
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else if activeTab === 'rules'}
			<div
				class="h-full flex flex-col bg-[var(--header-bg)]/60 backdrop-blur-md border border-stone-800 overflow-hidden industrial-frame"
			>
				<div
					class="p-6 border-b border-stone-800 flex justify-between items-center bg-[var(--header-bg)]"
				>
					<div class="flex items-center gap-4">
						<div class="w-2.5 h-2.5 bg-success shadow-[0_0_10px_#10b981] animate-pulse"></div>
						<h3 class="font-heading font-black text-xs text-white uppercase tracking-[0.3em]">
							DEPLOYED_PROTOCOL_SET
						</h3>
					</div>
					<button
						onclick={() => openModal()}
						class="px-6 py-2.5 bg-danger hover:bg-danger/80 text-white font-heading font-black text-[10px] uppercase tracking-widest flex items-center gap-3 transition-all shadow-lg shadow-red-900/20 active:tranneutral-y-px"
					>
						<Plus class="w-4 h-4" /> Initialize_Protocol
					</button>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left font-jetbrains text-[11px] border-collapse">
						<thead class="bg-[var(--header-bg)] text-text-dim sticky top-0 z-10 border-b border-stone-800">
							<tr class="uppercase font-black tracking-widest">
								<th class="px-8 py-5 border-r border-stone-800/30"
									>Protocol_ID</th
								>
								<th class="px-8 py-5 border-r border-stone-800/30"
									>Target_Path</th
								>
								<th class="px-8 py-5 border-r border-stone-800/30 text-center"
									>Action_Directive</th
								>
								<th class="px-8 py-5 text-right"
									>Sequence_Op</th
								>
							</tr>
						</thead>
						<tbody class="divide-y divide-stone-900">
							{#each rules as rule (rule.id)}
								<tr class="hover:bg-danger/5 transition-all group">
									<td class="px-8 py-5 border-r border-stone-800/20">
										<div class="flex flex-col gap-1">
											<span
												class="font-black text-stone-200 group-hover:text-danger transition-colors uppercase tracking-tight"
												>{rule.name}</span
											>
											<span class="text-[9px] text-text-dim tracking-widest font-bold"
												>{rule.cidr}</span
											>
										</div>
									</td>
									<td class="px-8 py-5 border-r border-stone-800/20">
										<div class="flex items-center gap-3">
											<span
												class="text-text-dim bg-stone-950 px-2 py-1 border border-stone-800 font-bold"
												>PORT:{rule.port}</span
											>
											{#if rule.path_pattern}
												<span class="text-text-dim truncate max-w-[200px] uppercase font-bold"
													>{rule.path_pattern}</span
												>
											{/if}
										</div>
									</td>
									<td class="px-8 py-5 border-r border-stone-800/20 text-center">
										<span
											class={`px-3 py-1 text-[9px] font-black rounded-none border tracking-[0.2em] uppercase ${getActionColor(
												rule.action
											)}`}
										>
											{rule.action}
										</span>
									</td>
									<td class="px-8 py-5">
										<div
											class="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300"
										>
											<button
												onclick={() => openModal(rule)}
												class="p-2 bg-stone-900 border border-stone-800 hover:border-rust hover:text-rust transition-all"
											>
												<Edit2 class="w-4 h-4" />
											</button>
											<button
												onclick={() => deleteRule(rule.id)}
												class="p-2 bg-stone-900 border border-stone-800 hover:border-red-500 hover:text-danger transition-all"
											>
												<Trash2 class="w-4 h-4" />
											</button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{:else if activeTab === 'bans'}
			<div class="h-full overflow-auto grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-3 gap-6 custom-scrollbar pr-4">
				{#each bans as ban}
					<div
						class="modern-industrial-card glass-panel p-6 group hover:border-red-500/40 transition-all !rounded-none"
					>
						<div class="flex justify-between items-start mb-6">
							<div class="flex items-center gap-5">
								<div
									class="bg-stone-950 p-4 border border-stone-800 text-danger group-hover:border-red-500/50 transition-all industrial-frame shadow-lg shadow-red-900/10"
								>
									<Lock class="w-8 h-8" />
								</div>
								<div>
									<h4
										class="text-xl font-jetbrains font-black text-white leading-none tracking-tighter"
									>
										{ban.ip}
									</h4>
									<div class="flex items-center gap-3 mt-3">
										<div class="w-24 h-1 bg-stone-900 border border-stone-800 p-0 relative overflow-hidden">
											<div
												class="h-full bg-danger shadow-[0_0_10px_rgba(239,68,68,0.5)]"
												style="width: {ban.reputation_score}%"
											></div>
										</div>
										<span class="font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-widest"
											>Threat: {ban.reputation_score}%</span>
									</div>
								</div>
							</div>
							<button
								onclick={() => unbanIP(ban.ip)}
								class="px-4 py-2 bg-stone-900 border border-stone-800 hover:border-emerald-500 hover:text-success font-heading font-black text-[9px] uppercase tracking-widest transition-all shadow-lg active:tranneutral-y-px"
							>
								DE_QUARANTINE
							</button>
						</div>
						<div
							class="pt-6 border-t border-stone-800 grid grid-cols-2 gap-6"
						>
							<div class="flex flex-col gap-2">
								<span class="font-jetbrains text-[8px] text-text-dim uppercase font-black tracking-[0.2em]"
									>INCIDENT_REPORT</span
								>
								<p class="font-jetbrains text-[10px] text-stone-400 line-clamp-2 italic uppercase leading-tight">
									"{ban.ban_reason}"
								</p>
							</div>
							<div class="flex flex-col gap-2 text-right">
								<span class="font-jetbrains text-[8px] text-text-dim uppercase font-black tracking-[0.2em]"
									>RESTORATION_T</span
								>
								<span class="font-jetbrains text-[10px] text-danger/80 font-black tracking-widest"
									>{ban.ban_expires_at
										? new Date(ban.ban_expires_at).toLocaleDateString()
										: 'PERPETUAL_LOCK'}</span
								>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'anticheat'}
			<div class="h-full overflow-auto space-y-3 custom-scrollbar pr-4">
				{#each events as event}
					<div
						class="bg-stone-900/40 border border-stone-800 border-l-4 {event.severity > 80
							? 'border-l-red-600 bg-red-950/10'
							: 'border-l-stone-700'} p-6 rounded-none flex items-center gap-10 group transition-all hover:bg-stone-900 hover:border-stone-700 industrial-frame"
					>
						<div class="text-center w-28 shrink-0 border-r border-stone-800 pr-10">
							<div class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest">
								{new Date(event.timestamp).toLocaleDateString()}
							</div>
							<div class="font-jetbrains text-[11px] font-black text-stone-300 mt-1 tabular-nums tracking-widest">
								{new Date(event.timestamp).toLocaleTimeString([], {
									hour: '2-digit',
									minute: '2-digit',
									second: '2-digit',
									hour12: false
								})}
							</div>
						</div>

						<div class="flex-1">
							<div class="flex items-center gap-5">
								<span
									class="text-lg font-heading font-black text-white uppercase tracking-tighter"
									>{event.event_type}</span
								>
								<div
									class="px-3 py-1 bg-stone-950 border border-stone-800 text-[10px] font-jetbrains font-black text-danger/70 tracking-widest uppercase shadow-inner"
								>
									IP::{event.client_ip}
								</div>
							</div>
							<p class="font-jetbrains text-[10px] font-bold text-text-dim mt-2 uppercase tracking-widest opacity-60">
								TARGET_IDENTITY: {event.player_id || 'NULL_REFERENCE'}
							</p>
						</div>

						<div class="flex flex-col items-end gap-3 pr-4">
							<div class="flex gap-1">
								{#each Array(5) as _, i}
									<div
										class={`w-4 h-1.5 rounded-none ${i < Math.ceil(event.severity / 20)
											? event.severity > 80
												? 'bg-danger shadow-[0_0_8px_#ef4444]'
												: 'bg-amber-500 shadow-[0_0_8px_#f59e0b]'
											: 'bg-stone-800'}`}
									></div>
								{/each}
							</div>
							<span
								class={`font-heading font-black text-[10px] uppercase tracking-[0.2em] ${event.severity > 80
									? 'text-danger'
									: 'text-text-dim'}`}>THREAT_LVL: {event.severity}</span>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'config'}
			<div class="h-full flex items-center justify-center p-8 overflow-auto">
				<div
					class="w-full max-w-3xl modern-industrial-card glass-panel p-12 shadow-[0_0_100px_rgba(239,68,68,0.05)] relative overflow-hidden !rounded-none"
				>
					<div class="absolute top-0 right-0 p-6 opacity-[0.02] pointer-events-none">
						<Settings class="w-48 h-48" />
					</div>

					<h3
						class="text-2xl font-heading font-black text-white uppercase tracking-tighter mb-12 flex items-center gap-4"
					>
						<Zap class="w-8 h-8 text-danger animate-pulse" /> Security Configuration
					</h3>

					<div class="space-y-12">
						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-base font-heading font-black text-stone-200 uppercase tracking-widest group-hover:text-danger transition-colors"
								>
									Auto-Ban System
								</h4>
								<p class="font-jetbrains text-[11px] text-text-dim mt-2 uppercase font-bold leading-relaxed tracking-tight">
									Automatically quarantine players who exceed the reputation risk threshold.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.auto_ban_enabled'] = !config['redeye.auto_ban_enabled'];
									updateConfig();
								}}
								class={`w-16 h-8 rounded-none transition-all relative border-2 ${config[
									'redeye.auto_ban_enabled'
								]
									? 'bg-danger border-red-400 shadow-[0_0_20px_rgba(239,68,68,0.4)]'
									: 'bg-stone-900 border-stone-800'}`}
								aria-label="Toggle autonomous neutralization"
							>
								<div
									class={`absolute top-1 left-1 w-4.5 h-4.5 bg-white transition-all ${config[
										'redeye.auto_ban_enabled'
									]
										? 'tranneutral-x-8 bg-white shadow-lg'
										: 'bg-stone-700'}`}
								></div>
							</button>
						</div>

						<div class="space-y-6">
							<div class="flex justify-between items-end">
								<h4
									class="font-heading font-black text-base text-stone-200 uppercase tracking-widest"
								>
									THREAT_SENSITIVITY_TUNING
								</h4>
								<span class="text-3xl font-heading font-black text-danger tracking-widest tabular-nums"
									>{config['redeye.auto_ban_threshold']}</span>
							</div>
							<div class="relative py-4">
								<input
									type="range"
									min="10"
									max="200"
									step="10"
									bind:value={config['redeye.auto_ban_threshold']}
									onchange={updateConfig}
									class="w-full h-1.5 bg-stone-950 border border-stone-800 appearance-none cursor-pointer accent-red-600"
									aria-label="Threat sensitivity threshold"
								/>
							</div>
							<div class="flex justify-between font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-[0.3em]">
								<span>Passive_Mitigation</span>
								<span>Aggressive_Purge</span>
							</div>
						</div>

						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-base font-heading font-black text-stone-200 uppercase tracking-widest"
								>
									Neural_Notification_Feed
								</h4>
								<p class="font-jetbrains text-[11px] text-text-dim mt-2 uppercase font-bold leading-relaxed tracking-tight">
									Broadcast real-time intercept telemetry to decentralized command terminals.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.alert_enabled'] = !config['redeye.alert_enabled'];
									updateConfig();
								}}
								class={`w-16 h-8 rounded-none transition-all relative border-2 ${config['redeye.alert_enabled']
									? 'bg-success border-emerald-400 shadow-[0_0_20px_rgba(16,185,129,0.4)]'
									: 'bg-stone-900 border-stone-800'}`}
								aria-label="Toggle alert notifications"
							>
								<div
									class={`absolute top-1 left-1 w-4.5 h-4.5 bg-white transition-all ${config[
										'redeye.alert_enabled'
									]
										? 'tranneutral-x-8 bg-white shadow-lg'
										: 'bg-stone-700'}`}
								></div>
							</button>
						</div>
					</div>

					<div class="mt-16 pt-10 border-t border-stone-800 text-center">
						<div class="font-jetbrains text-[10px] font-black text-stone-700 uppercase tracking-[0.5em] animate-pulse">
							Status: Synchronizing_With_Mainframe_Cores...
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Logs -->
			<div
				class="h-full flex flex-col bg-[var(--header-bg)]/60 backdrop-blur-md border border-stone-800 overflow-hidden industrial-frame"
			>
				<div
					class="p-6 border-b border-stone-800 bg-[var(--header-bg)] flex items-center justify-between"
				>
					<div class="flex items-center gap-4">
						<Terminal class="w-5 h-5 text-sky-500" />
						<h3 class="font-heading font-black text-xs text-white uppercase tracking-[0.3em]">
							PACKET_INTERCEPT_STREAM
						</h3>
					</div>
					<div class="font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-widest">
						CHANNEL_0xFF // eth0_UPLINK
					</div>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar font-jetbrains text-[11px] p-6">
					<div class="space-y-2">
						{#each logs as log}
							<div
								class="flex items-center gap-6 py-2 border-b border-stone-800/30 hover:bg-white/[0.02] px-4 transition-colors group"
							>
								<span class="text-text-dim shrink-0 font-bold tabular-nums"
									>[{new Date(log.timestamp).toLocaleTimeString([], { hour12: false })}]</span
								>
								<span class={`font-black uppercase tracking-tighter w-28 shrink-0 ${getActionColor(log.action).split(' ')[0]}`}
									>{log.action}</span
								>
								<span class="text-stone-700 font-black shrink-0 tracking-widest text-[9px]">SOURCE_IP:</span>
								<span
									class="text-stone-300 group-hover:text-danger transition-colors font-black tracking-tight"
									>{log.source_ip}</span
								>
								<span class="font-jetbrains text-[9px] text-stone-800 ml-auto uppercase font-black tracking-[0.2em] hidden xl:block"
									>Intercepted_Verified_0x88</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Modal -->
{#if showModal}
	<div
		class="fixed inset-0 z-[500] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
		transition:fade={{ duration: 150 }}
		onclick={() => (showModal = false)}
		onkeydown={(e) => e.key === 'Escape' && (showModal = false)}
		role="button"
		tabindex="0"
		aria-label="Close modal"
	>
		<div
			class="bg-[var(--terminal-bg)] border border-danger/30 rounded-none shadow-[0_0_100px_rgba(0,0,0,0.8)] w-full max-w-lg overflow-hidden relative industrial-frame"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<div
				class="absolute top-0 left-0 w-full h-1 bg-danger shadow-[0_0_15px_#ef4444]"
			></div>

			<div
				class="p-6 sm:p-8 border-b border-stone-800 flex justify-between items-center bg-[var(--header-bg)]"
			>
				<div>
					<h3 class="text-xl font-heading font-black text-white uppercase tracking-tighter">
						{editingRule ? 'Modify' : 'Init'} Protocol
					</h3>
					<p class="text-[10px] text-text-dim font-jetbrains mt-1 uppercase tracking-widest font-bold">
						Define neural gateway parameters
					</p>
				</div>
				<button
					onclick={() => (showModal = false)}
					class="p-2 text-text-dim hover:text-white transition-all hover:rotate-90">
					<X class="w-6 h-6" />
				</button>
			</div>

			<div class="p-6 sm:p-8 space-y-6 bg-[var(--terminal-bg)]">
				<div class="space-y-2">
					<label
						for="ruleName"
						class="text-[10px] font-black text-text-dim uppercase tracking-widest ml-1"
						>Identity Tag</label
					>
					<input
						id="ruleName"
						type="text"
						bind:value={form.name}
						placeholder="Protocol Identifier"
						class="w-full bg-stone-950 border border-stone-800 px-4 py-3 text-sm text-white focus:outline-none focus:border-red-500 transition-all font-jetbrains uppercase tracking-widest shadow-inner"
					/>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
					<div class="space-y-2">
						<label
							for="ruleCidrIp"
							class="text-[10px] font-black text-text-dim uppercase tracking-widest ml-1"
							>Target IP / CIDR</label
						>
						<input
							id="ruleCidrIp"
							type="text"
							bind:value={form.cidr}
							placeholder="0.0.0.0/0"
							class="w-full bg-stone-950 border border-stone-800 px-4 py-3 text-sm text-white focus:outline-none focus:border-red-500 transition-all font-jetbrains tracking-widest shadow-inner"
						/>
					</div>
					<div class="space-y-2">
						<label
							for="rulePort"
							class="text-[10px] font-black text-text-dim uppercase tracking-widest ml-1"
							>Gateway Port</label
						>
						<input
							id="rulePort"
							type="text"
							bind:value={form.port}
							class="w-full bg-stone-950 border border-stone-800 px-4 py-3 text-sm text-white focus:outline-none focus:border-red-500 transition-all font-jetbrains text-center shadow-inner"
							placeholder="*"
						/>
					</div>
				</div>

				<div class="space-y-2">
					<label
						for="rulePathPattern"
						class="text-[10px] font-black text-text-dim uppercase tracking-widest ml-1"
						>Path Filter (Optional)</label
					>
					<input
						id="rulePathPattern"
						type="text"
						bind:value={form.path_pattern}
						placeholder="/v1/secure/..."
						class="w-full bg-stone-950 border border-stone-800 px-4 py-3 text-sm text-white focus:outline-none focus:border-red-500 transition-all font-jetbrains shadow-inner"
					/>
				</div>

				<div class="space-y-3" role="group" aria-labelledby="action-protocol-label">
					<span id="action-protocol-label" class="text-[10px] font-black text-text-dim uppercase tracking-widest ml-1 block"
						>Action Protocol</span
					>
					<div class="grid grid-cols-3 gap-2">
						{#each ['ALLOW', 'DENY', 'RATE_LIMIT'] as action}
							<button
								onclick={() => (form.action = action as any)}
								class="py-3 border-2 text-[9px] font-black uppercase tracking-widest transition-all {form.action ===
								action
									? action === 'ALLOW'
										? 'bg-success/20 border-success/50 text-success shadow-[0_0_15px_rgba(16,185,129,0.2)]'
										: action === 'DENY'
											? 'bg-danger/20 border-danger/50 text-danger shadow-[0_0_15px_rgba(239,68,68,0.2)]'
											: 'bg-warning/20 border-warning/50 text-warning shadow-[0_0_15px_rgba(245,158,11,0.2)]'
									: 'bg-stone-900 border-stone-800 text-text-dim hover:border-stone-700'}"
							>
								{action.replace('_', ' ')}
							</button>
						{/each}
					</div>
				</div>

				{#if form.action === 'RATE_LIMIT'}
					<div
						class="grid grid-cols-2 gap-6 bg-warning/5 p-4 border border-warning/20 shadow-inner"
						transition:slide
					>
						<div class="space-y-1.5">
							<label
								for="ruleLimit"
								class="text-[9px] font-black text-warning/70 uppercase tracking-widest"
								>Rate (req/sec)</label
							>
							<input
								id="ruleLimit"
								type="number"
								bind:value={form.rate_limit}
								class="w-full bg-black border border-stone-800 px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all font-jetbrains shadow-inner"
							/>
						</div>
						<div class="space-y-1.5">
							<label
								for="ruleBurst"
								class="text-[9px] font-black text-warning/70 uppercase tracking-widest"
								>Burst Capacity</label
							>
							<input
								id="ruleBurst"
								type="number"
								bind:value={form.burst}
								class="w-full bg-black border border-stone-800 px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all font-jetbrains shadow-inner"
							/>
						</div>
					</div>
				{/if}
			</div>

			<div
				class="p-6 sm:p-8 border-t border-stone-800 bg-[var(--header-bg)] flex justify-end gap-4"
			>
				<button
					onclick={() => (showModal = false)}
					class="px-6 py-2.5 text-[10px] font-black text-text-dim hover:text-white uppercase tracking-widest transition-all"
				>
					Abort
				</button>
				<button
					onclick={saveRule}
					class="px-8 py-3 bg-danger hover:bg-danger/80 text-white text-[10px] font-black uppercase tracking-[0.2em] transition-all shadow-xl shadow-red-900/20 active:tranneutral-y-px"
				>
					Initialize
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes scan {
		from {
			transform: translateY(-100%);
		}
		to {
			transform: translateY(100vh);
		}
	}

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
