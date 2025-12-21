<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, slide, scale } from 'svelte/transition';
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
		Settings
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
				return 'text-emerald-400 bg-emerald-950/20 border-emerald-500/30';
			case 'DENY':
				return 'text-rose-400 bg-rose-950/20 border-rose-500/30';
			case 'RATE_LIMIT':
				return 'text-amber-400 bg-amber-950/20 border-amber-500/30';
			default:
				return 'text-slate-500 dark:text-slate-400 bg-slate-800 border-slate-300 dark:border-slate-700';
		}
	}

	onMount(() => {
		fetchStats();
	});
</script>

<div class="p-6 h-full flex flex-col overflow-hidden relative font-sans">
	<!-- Tech Background elements -->
	<div class="absolute inset-0 pointer-events-none opacity-5">
		<div class="absolute top-0 left-0 w-full h-full bg-[url('/grid.svg')]"></div>
		<div
			class="absolute top-0 left-0 w-full h-[1px] bg-red-500 animate-[scan_4s_linear_infinite]"
		></div>
	</div>

	<!-- Header -->
	<div class="flex justify-between items-center mb-8 shrink-0 relative z-10">
		<div class="flex items-center gap-4">
			<div class="relative">
				<div class="absolute inset-0 bg-red-500 blur-xl opacity-20 animate-pulse"></div>
				<div class="relative bg-slate-900 border border-red-500/50 p-3 rounded-xl">
					<Eye class="w-8 h-8 text-red-500" />
				</div>
			</div>
			<div>
				<div class="flex items-center gap-2">
					<h1 class="text-3xl font-black tracking-tighter text-slate-900 dark:text-white uppercase">
						RedEye <span class="text-red-500">Protocol</span>
					</h1>
					<div
						class="px-2 py-0.5 {stats.system_active
							? 'bg-red-500 text-black'
							: 'bg-slate-700 text-slate-300'} text-[10px] font-bold rounded uppercase"
					>
						{stats.system_active ? 'Active' : 'Inactive'}
					</div>
				</div>
				<p class="text-slate-500 text-xs font-mono tracking-widest mt-1">
					NEURAL THREAT MITIGATION SYSTEM // {stats.system_error
						? `ERR: ${stats.system_error}`
						: 'REV 4.0.2'}
				</p>
			</div>
		</div>

		<div class="flex items-center gap-3">
			<div class="hidden md:flex flex-col items-end mr-4 text-right">
				<span class="text-[10px] text-slate-500 font-mono uppercase tracking-tighter"
					>System Entropy</span
				>
				<span class="text-xs text-red-400 font-mono">{stats.entropy?.toFixed(4) || '0.0000'}%</span>
			</div>
			<button
				onclick={refreshAll}
				class="p-2.5 bg-slate-900 border border-slate-200 dark:border-slate-800 hover:border-red-500/50 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white rounded-xl transition-all shadow-lg"
				aria-label="Refresh security data"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- System Status Banner -->
	<div
		class="mb-6 p-4 bg-white/50 dark:bg-slate-950/50 border-y border-slate-200 dark:border-slate-800 flex flex-wrap gap-8 items-center relative overflow-hidden"
	>
		<div
			class="absolute inset-y-0 left-0 w-1 {stats.system_active ? 'bg-red-500' : 'bg-slate-600'}"
		></div>
		<div class="flex items-center gap-3">
			<div
				class="w-2 h-2 rounded-full {stats.system_active
					? 'bg-red-500 animate-ping'
					: 'bg-slate-600'}"
			></div>
			<span class="text-xs font-mono text-slate-500 dark:text-slate-400 uppercase tracking-widest"
				>Live Monitoring: <span
					class={stats.system_active ? 'text-slate-900 dark:text-white' : 'text-slate-600'}
					>{stats.system_active ? 'Enabled' : 'Disabled'}</span
				></span
			>
		</div>
		<div class="flex items-center gap-2 text-xs font-mono">
			<span class="text-slate-500">Uptime:</span>
			<span class="text-red-400">{stats.uptime}</span>
		</div>
		<div class="flex items-center gap-2 text-xs font-mono">
			<span class="text-slate-500">Threat Level:</span>
			<span
				class={stats.threat_level === 'Low'
					? 'text-emerald-400'
					: stats.threat_level === 'Medium'
						? 'text-amber-400'
						: 'text-red-500'}>{stats.threat_level}</span
			>
		</div>
		<div class="ml-auto text-[10px] font-mono text-slate-600 hidden lg:block">
			CRC: {stats.crc} - NODE: {stats.node_id}
		</div>
	</div>

	<!-- Navigation Controls -->
	<div
		class="flex gap-1 mb-6 shrink-0 bg-white/80 dark:bg-slate-950/80 p-1.5 rounded-2xl border border-slate-200 dark:border-slate-800"
	>
		{#each [['overview', BarChart3, 'Status'], ['rules', ShieldCheck, 'Rules'], ['bans', Ban, 'Bans'], ['anticheat', AlertTriangle, 'Anti-Cheat'], ['logs', Terminal, 'Logs'], ['config', Settings, 'System']] as [id, icon, label]}
			{@const Icon = icon as any}
			<button
				onclick={() => {
					activeTab = id as any;
					refreshAll();
				}}
				class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl text-xs font-bold uppercase tracking-wider transition-all {activeTab ===
				id
					? 'bg-red-600 text-slate-900 dark:text-white shadow-xl shadow-red-900/20'
					: 'text-slate-500 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-900'}"
			>
				<Icon class="w-4 h-4" />
				<span class="hidden sm:inline">{label}</span>
			</button>
		{/each}
	</div>

	<!-- Main Display -->
	<div class="flex-1 min-h-0 relative z-10">
		{#if activeTab === 'overview'}
			<div class="h-full overflow-auto space-y-6 custom-scrollbar pr-2">
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					{#each [{ icon: ShieldCheck, val: stats.total_rules, label: 'Protocols', color: 'text-emerald-400' }, { icon: Ban, val: stats.active_bans, label: 'Quarantined', color: 'text-red-500' }, { icon: ShieldAlert, val: stats.events_24h, label: 'Incidents', color: 'text-amber-400' }, { icon: Activity, val: stats.logs_24h, label: 'Scans', color: 'text-sky-400' }] as card}
						{@const CardIcon = card.icon as any}
						<div
							class="bg-slate-900/40 border border-slate-200 dark:border-slate-800 p-5 rounded-2xl flex items-center gap-4 hover:border-slate-300 dark:border-slate-700 transition-colors group"
						>
							<div
								class="p-3 bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 group-hover:border-slate-600 transition-colors"
							>
								<CardIcon class="w-6 h-6 {card.color}" />
							</div>
							<div class="flex flex-col">
								<span
									class="text-2xl font-black text-slate-900 dark:text-white font-mono leading-none"
									>{card.val}</span
								>
								<span class="text-[10px] text-slate-500 uppercase tracking-widest font-bold mt-1"
									>{card.label}</span
								>
							</div>
						</div>
					{/each}
				</div>

				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Visual Threat Monitor -->
					<div
						class="lg:col-span-2 bg-slate-900/40 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 relative overflow-hidden flex flex-col items-center justify-center min-h-[350px]"
					>
						<div class="absolute inset-0 bg-[url('/grid.svg')] opacity-10"></div>
						<div
							class="absolute inset-0 bg-gradient-to-t from-slate-950/80 via-transparent to-transparent"
						></div>

						<div class="relative z-10 w-full max-w-lg text-center">
							<div
								class="inline-flex items-center gap-2 px-3 py-1 bg-red-500/10 border border-red-500/20 rounded-full mb-6"
							>
								<div class="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse"></div>
								<span class="text-[10px] font-mono text-red-400 uppercase font-bold"
									>Neural Core Scan Active</span
								>
							</div>

							<div class="relative mx-auto w-48 h-48 mb-8">
								<div
									class="absolute inset-0 border-4 border-slate-200 dark:border-slate-800 rounded-full"
								></div>
								<div
									class="absolute inset-0 border-t-4 border-red-500 rounded-full animate-[spin_3s_linear_infinite]"
								></div>
								<div
									class="absolute inset-4 border border-slate-200 dark:border-slate-800 rounded-full border-dashed"
								></div>
								<div class="absolute inset-0 flex items-center justify-center">
									<Activity class="w-16 h-16 text-red-500 animate-pulse" />
								</div>
							</div>

							<h3
								class="text-2xl font-black text-slate-900 dark:text-white uppercase tracking-tight"
							>
								Active Neural Defense
							</h3>
							<p class="text-slate-500 dark:text-slate-400 text-sm mt-3 leading-relaxed">
								Autonomous traffic analysis is patrolling all endpoints. Current processing latency: <span
									class="text-red-400 font-mono">1.2ms</span
								>. Targeting
								<span class="text-slate-900 dark:text-white font-mono"
									>{stats.reputation_count}</span
								> known entities across all regions.
							</p>
						</div>
					</div>

					<!-- Quick Tech Stats -->
					<div class="space-y-4">
						<div
							class="bg-slate-900/40 border border-slate-200 dark:border-slate-800 rounded-2xl p-6"
						>
							<h4
								class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mb-4 flex items-center gap-2"
							>
								<Cpu class="w-3 h-3" /> System Load
							</h4>
							<div class="space-y-4">
								{#each [['Core Analyzer', Number(Math.min(100, 5 + stats.logs_24h / 100)), 'text-emerald-400'], ['Ban Sequencer', Number(Math.min(100, 2 + stats.active_bans * 2)), 'text-red-400'], ['Log Indexer', Number(Math.min(100, 1 + stats.reputation_count / 10)), 'text-blue-400']] as [label, perc, color]}
									<div class="space-y-1.5">
										<div class="flex justify-between text-[10px] font-mono uppercase">
											<span class="text-slate-500 dark:text-slate-400">{label}</span>
											<span class="text-slate-900 dark:text-white"
												>{Math.round(perc as number)}%</span
											>
										</div>
										<div class="h-1 bg-white dark:bg-slate-950 rounded-full overflow-hidden">
											<div
												class="h-full bg-slate-700 transition-all duration-1000"
												style="width: {perc}%"
											></div>
										</div>
									</div>
								{/each}
							</div>
						</div>

						<div
							class="bg-slate-900/40 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 flex flex-col justify-between"
						>
							<h4
								class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mb-4 flex items-center gap-2"
							>
								<Globe class="w-3 h-3" /> Global Reach
							</h4>
							<div class="flex items-center gap-4">
								<div class="text-3xl font-black text-slate-900 dark:text-white font-mono">08</div>
								<div class="text-[10px] text-slate-500 font-mono leading-tight">
									ACTIVE<br />ZONES
								</div>
							</div>
							<div class="mt-4 flex gap-1">
								{#each Array(8) as _, i}
									<div
										class="h-4 w-1 bg-red-500 rounded-full"
										style="opacity: {0.2 + i * 0.1}"
									></div>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else if activeTab === 'rules'}
			<div
				class="h-full flex flex-col bg-white/30 dark:bg-slate-950/30 rounded-2xl border border-slate-200 dark:border-slate-800 overflow-hidden"
			>
				<div
					class="p-4 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center bg-slate-900/50"
				>
					<div class="flex items-center gap-2">
						<div class="w-2 h-2 rounded-full bg-emerald-500"></div>
						<h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-widest">
							Active Rule Set
						</h3>
					</div>
					<button
						onclick={() => openModal()}
						class="px-4 py-2 bg-red-600 hover:bg-red-500 text-slate-900 dark:text-white rounded-xl text-[10px] font-bold uppercase tracking-widest flex items-center gap-2 transition-all shadow-lg shadow-red-900/20"
					>
						<Plus class="w-3 h-3" /> Initialize Protocol
					</button>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left text-sm border-separate border-spacing-0">
						<thead class="bg-slate-900/80 text-slate-500 sticky top-0 z-10">
							<tr class="text-[10px] uppercase font-mono tracking-tighter">
								<th class="px-6 py-4 border-b border-slate-200 dark:border-slate-800"
									>Protocol Name</th
								>
								<th class="px-6 py-4 border-b border-slate-200 dark:border-slate-800"
									>Destination/Path</th
								>
								<th class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 text-center"
									>Action</th
								>
								<th class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 text-right"
									>Sequence</th
								>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-800/50">
							{#each rules as rule (rule.id)}
								<tr class="hover:bg-slate-800/30 transition-all group">
									<td class="px-6 py-4">
										<div class="flex flex-col">
											<span
												class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-slate-900 dark:text-white transition-colors"
												>{rule.name}</span
											>
											<span class="text-[10px] font-mono text-slate-500 mt-0.5 tracking-tighter"
												>{rule.cidr}</span
											>
										</div>
									</td>
									<td class="px-6 py-4 font-mono text-xs">
										<div class="flex items-center gap-2">
											<span
												class="text-slate-500 dark:text-slate-400 bg-slate-900 px-1.5 py-0.5 rounded border border-slate-200 dark:border-slate-800"
												>PORT:{rule.port}</span
											>
											{#if rule.path_pattern}
												<span class="text-slate-500 truncate max-w-[150px]"
													>{rule.path_pattern}</span
												>
											{/if}
										</div>
									</td>
									<td class="px-6 py-4 text-center">
										<span
											class="px-2.5 py-1 text-[10px] font-black rounded uppercase border tracking-tighter {getActionColor(
												rule.action
											)}"
										>
											{rule.action}
										</span>
									</td>
									<td class="px-6 py-4">
										<div
											class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
										>
											<button
												onclick={() => openModal(rule)}
												class="p-2 hover:bg-slate-700 rounded-lg text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white transition-all"
											>
												<Edit2 class="w-3.5 h-3.5" />
											</button>
											<button
												onclick={() => deleteRule(rule.id)}
												class="p-2 hover:bg-rose-950/30 rounded-lg text-slate-500 dark:text-slate-400 hover:text-rose-400 transition-all"
											>
												<Trash2 class="w-3.5 h-3.5" />
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
			<div class="h-full overflow-auto grid grid-cols-1 xl:grid-cols-2 gap-4 custom-scrollbar">
				{#each bans as ban}
					<div
						class="bg-slate-900/40 border border-slate-200 dark:border-slate-800 rounded-2xl p-5 group hover:border-red-500/30 transition-all"
					>
						<div class="flex justify-between items-start">
							<div class="flex items-center gap-4">
								<div
									class="bg-white dark:bg-slate-950 p-3 rounded-xl border border-slate-200 dark:border-slate-800 text-red-500 group-hover:border-red-500/50 transition-all"
								>
									<Lock class="w-6 h-6" />
								</div>
								<div>
									<h4
										class="text-lg font-mono font-bold text-slate-900 dark:text-white leading-none"
									>
										{ban.ip}
									</h4>
									<div class="flex items-center gap-2 mt-2">
										<div class="w-20 h-1 bg-slate-800 rounded-full overflow-hidden">
											<div
												class="h-full bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]"
												style="width: {ban.reputation_score}%"
											></div>
										</div>
										<span class="text-[10px] font-mono text-slate-500 uppercase tracking-tighter"
											>Threat: {ban.reputation_score}%</span
										>
									</div>
								</div>
							</div>
							<button
								onclick={() => unbanIP(ban.ip)}
								class="px-3 py-1.5 bg-white dark:bg-slate-950 hover:bg-emerald-600 border border-slate-200 dark:border-slate-800 hover:border-emerald-500 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white rounded-lg text-[10px] font-bold uppercase tracking-widest transition-all"
							>
								De-Quarantine
							</button>
						</div>
						<div
							class="mt-4 pt-4 border-t border-slate-200/50 dark:border-slate-800/50 grid grid-cols-2 gap-4"
						>
							<div class="flex flex-col gap-1">
								<span class="text-[9px] text-slate-500 uppercase font-bold tracking-widest"
									>Incident Cause</span
								>
								<p class="text-xs text-slate-700 dark:text-slate-300 line-clamp-1 italic">
									"{ban.ban_reason}"
								</p>
							</div>
							<div class="flex flex-col gap-1 text-right">
								<span class="text-[9px] text-slate-500 uppercase font-bold tracking-widest"
									>Expiration</span
								>
								<span class="text-xs text-slate-500 dark:text-slate-400 font-mono"
									>{ban.ban_expires_at
										? new Date(ban.ban_expires_at).toLocaleDateString()
										: 'PERPETUAL'}</span
								>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'anticheat'}
			<div class="h-full overflow-auto space-y-2 custom-scrollbar">
				{#each events as event}
					<div
						class="bg-white/50 dark:bg-slate-950/50 border-l-2 {event.severity > 80
							? 'border-red-500 bg-red-500/5'
							: 'border-slate-200 dark:border-slate-800'} p-4 rounded-r-xl flex items-center gap-6 group transition-all hover:bg-slate-900/50"
					>
						<div class="text-center w-20 shrink-0">
							<div class="text-[10px] font-mono text-slate-500 leading-none">
								{new Date(event.timestamp).toLocaleDateString()}
							</div>
							<div class="text-xs font-mono text-slate-700 dark:text-slate-300 font-bold mt-1">
								{new Date(event.timestamp).toLocaleTimeString([], {
									hour: '2-digit',
									minute: '2-digit',
									second: '2-digit'
								})}
							</div>
						</div>

						<div class="flex-1">
							<div class="flex items-center gap-3">
								<span
									class="text-sm font-black text-slate-900 dark:text-white uppercase tracking-tighter"
									>{event.event_type}</span
								>
								<div
									class="px-1.5 py-0.5 bg-slate-900 border border-slate-200 dark:border-slate-800 text-[9px] font-mono text-slate-500 dark:text-slate-400 rounded"
								>
									IP:{event.client_ip}
								</div>
							</div>
							<p class="text-xs text-slate-500 mt-1 font-mono">
								{event.player_id || 'UNKNOWN_ENTITY'}
							</p>
						</div>

						<div class="flex flex-col items-end gap-2 pr-2">
							<div class="flex gap-0.5">
								{#each Array(5) as _, i}
									<div
										class="w-3 h-1 rounded-full {i < Math.ceil(event.severity / 20)
											? event.severity > 80
												? 'bg-red-500'
												: 'bg-amber-500'
											: 'bg-slate-800'}"
									></div>
								{/each}
							</div>
							<span
								class="text-[10px] font-bold uppercase tracking-widest {event.severity > 80
									? 'text-red-500'
									: 'text-slate-500'}">Level {event.severity}</span
							>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'config'}
			<div class="h-full flex items-center justify-center p-6 overflow-auto">
				<div
					class="w-full max-w-2xl bg-white/50 dark:bg-slate-950/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 shadow-2xl relative overflow-hidden"
				>
					<div class="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
						<Settings class="w-32 h-32" />
					</div>

					<h3
						class="text-xl font-black text-slate-900 dark:text-white uppercase tracking-tighter mb-8 flex items-center gap-3"
					>
						<Zap class="w-6 h-6 text-red-500" /> System Protocols
					</h3>

					<div class="space-y-10">
						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-sm font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wide group-hover:text-red-400 transition-colors"
								>
									Autonomous Neutralization
								</h4>
								<p class="text-xs text-slate-500 mt-1">
									Engage automatic quarantine protocols for entities exceeding risk thresholds.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.auto_ban_enabled'] = !config['redeye.auto_ban_enabled'];
									updateConfig();
								}}
								class="w-14 h-7 rounded-full transition-all relative {config[
									'redeye.auto_ban_enabled'
								]
									? 'bg-red-600 shadow-[0_0_15px_rgba(220,38,38,0.4)]'
									: 'bg-slate-800'}"
								aria-label="Toggle autonomous neutralization"
							>
								<div
									class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full transition-all {config[
										'redeye.auto_ban_enabled'
									]
										? 'translate-x-7'
										: ''}"
								></div>
							</button>
						</div>

						<div class="space-y-4">
							<div class="flex justify-between items-end">
								<h4
									class="text-sm font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wide"
								>
									Threat Sensitivity
								</h4>
								<span class="text-xl font-black font-mono text-red-500"
									>{config['redeye.auto_ban_threshold']}</span
								>
							</div>
							<div class="relative py-2">
								<input
									type="range"
									min="10"
									max="200"
									step="10"
									bind:value={config['redeye.auto_ban_threshold']}
									onchange={updateConfig}
									class="w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-red-500"
									aria-label="Threat sensitivity threshold"
								/>
							</div>
							<div class="flex justify-between text-[10px] font-mono text-slate-600 uppercase">
								<span>High Tolerance</span>
								<span>Aggressive Mitigation</span>
							</div>
						</div>

						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-sm font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wide"
								>
									Global Notification Feed
								</h4>
								<p class="text-xs text-slate-500 mt-1">
									Broadcast critical security events to the neural command center.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.alert_enabled'] = !config['redeye.alert_enabled'];
									updateConfig();
								}}
								class="w-14 h-7 rounded-full transition-all relative {config['redeye.alert_enabled']
									? 'bg-emerald-600 shadow-[0_0_15px_rgba(5,150,105,0.4)]'
									: 'bg-slate-800'}"
								aria-label="Toggle alert notifications"
							>
								<div
									class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full transition-all {config[
										'redeye.alert_enabled'
									]
										? 'translate-x-7'
										: ''}"
								></div>
							</button>
						</div>
					</div>

					<div class="mt-12 pt-8 border-t border-slate-200 dark:border-slate-800 text-center">
						<div class="text-[10px] font-mono text-slate-600 uppercase">
							Status: Syncing with Mainframe...
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Logs -->
			<div
				class="h-full flex flex-col bg-white/30 dark:bg-slate-950/30 rounded-2xl border border-slate-200 dark:border-slate-800 overflow-hidden"
			>
				<div
					class="p-4 border-b border-slate-200 dark:border-slate-800 bg-slate-900/50 flex items-center gap-2"
				>
					<Terminal class="w-4 h-4 text-sky-400" />
					<h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-widest">
						Packet Intercept Feed
					</h3>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar font-mono text-[11px]">
					<div class="p-4 space-y-1">
						{#each logs as log}
							<div
								class="flex items-center gap-4 py-1 border-b border-slate-200/30 dark:border-slate-800/30 hover:bg-slate-800/20 px-2 rounded transition-colors group"
							>
								<span class="text-slate-600 shrink-0"
									>[{new Date(log.timestamp).toLocaleTimeString()}]</span
								>
								<span class="font-bold {getActionColor(log.action).split(' ')[0]} w-24 shrink-0"
									>{log.action}</span
								>
								<span class="text-slate-500 dark:text-slate-400 shrink-0">SOURCE:</span>
								<span
									class="text-slate-900 dark:text-white group-hover:text-red-400 transition-colors"
									>{log.source_ip}</span
								>
								<span class="text-[9px] text-slate-700 ml-auto uppercase hidden sm:block"
									>Intercepted @ eth0</span
								>
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
		class="fixed inset-0 z-[999] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
		transition:fade
		onclick={() => (showModal = false)}
		role="button"
		tabindex="0"
		onkeydown={(e) => {
			if (e.key === 'Escape') showModal = false;
		}}
	>
		<div
			class="bg-slate-900 border border-red-500/30 rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden relative"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 300, start: 0.95 }}
		>
			<div
				class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-red-500 to-transparent"
			></div>

			<div
				class="p-6 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center bg-white/50 dark:bg-slate-950/50"
			>
				<div>
					<h3 class="text-xl font-black text-slate-900 dark:text-white uppercase tracking-tighter">
						{editingRule ? 'Edit' : 'New'} Protocol
					</h3>
					<p class="text-[10px] text-slate-500 font-mono mt-1 uppercase">
						Define neural gateway parameters
					</p>
				</div>
				<button
					onclick={() => (showModal = false)}
					class="text-slate-500 hover:text-slate-900 dark:text-white transition-colors">✕</button
				>
			</div>

			<div class="p-8 space-y-6">
				<div class="space-y-1.5">
					<label
						for="ruleName"
						class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
						>Identity Tag</label
					>
					<input
						id="ruleName"
						type="text"
						bind:value={form.name}
						placeholder="Protocol Identifier"
						class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-3 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-red-500 transition-all font-mono"
					/>
				</div>

				<div class="grid grid-cols-2 gap-6">
					<div class="space-y-1.5">
						<label
							for="ruleCidrIp"
							class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
							>Target IP / CIDR</label
						>
						<input
							id="ruleCidrIp"
							type="text"
							bind:value={form.cidr}
							placeholder="0.0.0.0/0"
							class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-3 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-red-500 transition-all font-mono"
						/>
					</div>
					<div class="space-y-1.5">
						<label
							for="rulePort"
							class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
							>Gateway Port</label
						>
						<input
							id="rulePort"
							type="text"
							bind:value={form.port}
							class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-3 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-red-500 transition-all font-mono text-center"
							placeholder="*"
						/>
					</div>
				</div>

				<div class="space-y-1.5">
					<label
						for="rulePathPattern"
						class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
						>Path Filter (Optional)</label
					>
					<input
						id="rulePathPattern"
						type="text"
						bind:value={form.path_pattern}
						placeholder="/v1/secure/..."
						class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-3 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-red-500 transition-all font-mono"
					/>
				</div>

				<div class="space-y-3">
					<label class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1"
						>Action Protocol</label
					>
					<div class="grid grid-cols-3 gap-2">
						{#each ['ALLOW', 'DENY', 'RATE_LIMIT'] as action}
							<button
								onclick={() => (form.action = action as any)}
								class="py-3 rounded-xl border-2 text-[10px] font-black uppercase tracking-tighter transition-all {form.action ===
								action
									? action === 'ALLOW'
										? 'bg-emerald-600/20 border-emerald-500 text-emerald-400'
										: action === 'DENY'
											? 'bg-red-600/20 border-red-500 text-red-400'
											: 'bg-amber-600/20 border-amber-500 text-amber-400'
									: 'bg-white dark:bg-slate-950 border-slate-200 dark:border-slate-800 text-slate-600 hover:border-slate-300 dark:border-slate-700'}"
							>
								{action.replace('_', ' ')}
							</button>
						{/each}
					</div>
				</div>

				{#if form.action === 'RATE_LIMIT'}
					<div
						class="grid grid-cols-2 gap-6 bg-amber-500/5 p-4 rounded-2xl border border-amber-500/10"
						transition:slide
					>
						<div class="space-y-1.5">
							<label
								for="ruleLimit"
								class="text-[9px] font-bold text-amber-500/70 uppercase tracking-widest"
								>Rate (req/sec)</label
							>
							<input
								id="ruleLimit"
								type="number"
								bind:value={form.rate_limit}
								class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-amber-500 transition-all font-mono"
							/>
						</div>
						<div class="space-y-1.5">
							<label
								for="ruleBurst"
								class="text-[9px] font-bold text-amber-500/70 uppercase tracking-widest"
								>Burst Capacity</label
							>
							<input
								id="ruleBurst"
								type="number"
								bind:value={form.burst}
								class="w-full bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:border-amber-500 transition-all font-mono"
							/>
						</div>
					</div>
				{/if}
			</div>

			<div
				class="p-6 border-t border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-950/80 flex justify-end gap-3"
			>
				<button
					onclick={() => (showModal = false)}
					class="px-6 py-3 rounded-xl text-[10px] font-bold text-slate-500 hover:text-slate-900 dark:text-white uppercase tracking-widest transition-all"
				>
					Abort
				</button>
				<button
					onclick={saveRule}
					class="px-8 py-3 rounded-xl bg-red-600 hover:bg-red-500 text-slate-900 dark:text-white text-[10px] font-black uppercase tracking-[0.2em] transition-all shadow-xl shadow-red-900/20"
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
