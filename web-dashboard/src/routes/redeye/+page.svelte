<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, scale, slide } from 'svelte/transition';
	import {
		ShieldCheck,
		Ban,
		AlertTriangle,
		Terminal,
		Settings,
		ShieldAlert,
		Activity,
		Cpu,
		Globe,
		Plus,
		Edit2,
		Trash2,
		Lock,
		Zap,
		X,
		RefreshCw,
		BarChart3,
		ChevronRight,
		Search,
		AlertOctagon
	} from 'lucide-svelte';
	import { apiFetch, notify } from '$lib/api';
	import Icon from '$lib/components/theme/Icon.svelte';
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import type {
		RedEyeRule,
		RedEyeLog,
		RedEyeAnticheatEvent,
		RedEyeIPReputation
	} from '$lib/stores.svelte';

	let activeTab = $state<'overview' | 'rules' | 'bans' | 'anticheat' | 'logs' | 'config'>(
		'overview'
	);
	let loading = $state(false);
	let rules = $state<RedEyeRule[]>([]);
	let bans = $state<RedEyeIPReputation[]>([]);
	let events = $state<RedEyeAnticheatEvent[]>([]);
	let logs = $state<RedEyeLog[]>([]);
	let stats = $state<any>({});
	let config = $state<any>({
		'redeye.auto_ban_enabled': true,
		'redeye.auto_ban_threshold': 100,
		'redeye.alert_enabled': true
	});

	// Modal state
	let showModal = $state(false);
	let editingRule = $state<RedEyeRule | null>(null);
	let form = $state<Partial<RedEyeRule>>({
		name: '',
		cidr: '',
		port: '*',
		path_pattern: '',
		action: 'DENY',
		protocol: 'ANY',
		rate_limit: 0,
		burst: 0,
		enabled: true
	});

	async function refreshAll() {
		loading = true;
		try {
			const [sRes, rRes, bRes, eRes, lRes, cRes] = await Promise.all([
				apiFetch('/api/redeye/stats'),
				apiFetch('/api/redeye/rules'),
				apiFetch('/api/redeye/bans'),
				apiFetch('/api/redeye/anticheat/events'),
				apiFetch('/api/redeye/logs'),
				apiFetch('/api/redeye/config')
			]);

			if (sRes.ok) stats = await sRes.json();
			if (rRes.ok) rules = await rRes.json();
			if (bRes.ok) bans = await bRes.json();
			if (eRes.ok) {
				const data = await eRes.json();
				events = data.events || [];
			}
			if (lRes.ok) {
				const data = await lRes.json();
				logs = data.logs || [];
			}
			if (cRes.ok) config = await cRes.json();
		} catch (e) {
			console.error('Failed to refresh firewall data', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		refreshAll();
		const interval = setInterval(refreshAll, 30000);
		return () => clearInterval(interval);
	});

	function openModal(rule: RedEyeRule | null = null) {
		editingRule = rule;
		if (rule) {
			form = { ...rule };
		} else {
			form = {
				name: '',
				cidr: '',
				port: '*',
				path_pattern: '',
				action: 'DENY',
				protocol: 'ANY',
				rate_limit: 0,
				burst: 0,
				enabled: true
			};
		}
		showModal = true;
	}

	async function saveRule() {
		try {
			const url = editingRule ? `/api/redeye/rules/${editingRule.id}` : '/api/redeye/rules';
			const res = await apiFetch(url, {
				method: editingRule ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(form)
			});

			if (res.ok) {
				notify('Rule saved successfully', 'success');
				showModal = false;
				refreshAll();
			} else {
				const err = await res.json();
				notify(err.error || 'Failed to save rule', 'error');
			}
		} catch (e) {
			notify('Connection error', 'error');
		}
	}

	async function deleteRule(id: number) {
		if (!confirm('Are you sure you want to delete this rule?')) return;
		try {
			const res = await apiFetch(`/api/redeye/rules/${id}`, { method: 'DELETE' });
			if (res.ok) {
				notify('Rule deleted', 'success');
				refreshAll();
			}
		} catch (e) {
			notify('Failed to delete rule', 'error');
		}
	}

	async function unbanIP(ip: string) {
		try {
			const res = await apiFetch(`/api/redeye/bans/${ip}`, { method: 'DELETE' });
			if (res.ok) {
				notify(`IP ${ip} released`, 'success');
				refreshAll();
			}
		} catch (e) {
			notify('Failed to release IP', 'error');
		}
	}

	async function updateConfig() {
		try {
			const res = await apiFetch('/api/redeye/config', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(config)
			});
			if (res.ok) {
				notify('Configuration updated', 'success');
			}
		} catch (e) {
			notify('Failed to update config', 'error');
		}
	}

	function getActionColor(action: string) {
		switch (action) {
			case 'ALLOW':
				return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20';
			case 'DENY':
				return 'text-rose-400 bg-rose-500/10 border-rose-500/20';
			case 'RATE_LIMIT':
				return 'text-amber-400 bg-amber-500/10 border-amber-500/20';
			default:
				return 'text-slate-400 bg-slate-500/10 border-slate-500/20';
		}
	}
</script>

<div class="w-full h-full flex flex-col overflow-hidden relative font-sans">
	<!-- Header -->
	<div
		class="flex flex-col lg:flex-row lg:justify-between lg:items-center mb-8 gap-6 shrink-0 relative z-10"
	>
		<div class="flex items-center gap-4">
			<div class="p-3 bg-rose-500/10 rounded-xl border border-rose-500/20">
				<Icon name="ph:shield-check-bold" size="1.8rem" class="text-rose-500" />
			</div>
			<div>
				<div class="flex items-center gap-3">
					<h1 class="text-3xl font-bold text-white tracking-tight">
						Network <span class="text-rose-500">Firewall</span>
					</h1>
					<div
						class="px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wide {stats.system_active
							? 'bg-rose-500/10 text-rose-500 border border-rose-500/20'
							: 'bg-slate-800 text-slate-400 border border-slate-700'}"
					>
						{stats.system_active ? 'Active' : 'Disabled'}
					</div>
				</div>
				<p class="text-xs font-medium text-slate-400 mt-1">
					Security Rules & Traffic Monitoring
				</p>
			</div>
		</div>

		<div class="flex items-center gap-6">
			<div class="flex flex-col items-end">
				<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider"
					>Risk Factor</span
				>
				<span class="text-lg font-mono font-semibold text-rose-400"
					>{stats.risk_factor?.toFixed(2) || '0.00'}%</span
				>
			</div>
			<button
				onclick={refreshAll}
				class="p-2.5 rounded-lg bg-slate-800/50 border border-white/5 text-slate-400 hover:text-white hover:bg-slate-800 transition-all"
				aria-label="Refresh data"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- System Status Banner -->
	<div
		class="mb-8 p-4 bg-slate-900/50 backdrop-blur-md border border-white/5 rounded-xl flex flex-nowrap overflow-x-auto no-scrollbar gap-8 items-center"
	>
		<div class="flex items-center gap-3 shrink-0">
			<div
				class="w-2 h-2 rounded-full {stats.system_active
					? 'bg-rose-500 animate-pulse'
					: 'bg-slate-600'}"
			></div>
			<span class="text-xs font-bold text-slate-400 uppercase tracking-wider"
				>Monitoring: <span class={stats.system_active ? 'text-slate-200' : 'text-slate-500'}
					>{stats.system_active ? 'Active' : 'Offline'}</span
				></span
			>
		</div>
		<div class="flex items-center gap-2 text-xs font-medium shrink-0">
			<span class="text-slate-500">Uptime:</span>
			<span class="text-slate-300">{stats.uptime}</span>
		</div>
		<div class="flex items-center gap-2 text-xs font-medium shrink-0">
			<span class="text-slate-500">Threat Level:</span>
			<span
				class={`px-2 py-0.5 rounded-md ${
					stats.threat_level === 'Low'
						? 'text-emerald-400 bg-emerald-500/10'
						: stats.threat_level === 'Medium'
							? 'text-amber-400 bg-amber-500/10'
							: 'text-rose-400 bg-rose-500/10'
				}`}>{stats.threat_level?.toUpperCase()}</span
			>
		</div>
		<div class="ml-auto text-xs font-mono text-slate-600 hidden lg:flex gap-4 shrink-0">
			<span>SYSTEM_ID: {stats.crc}</span>
			<span>NODE: {stats.node_id}</span>
		</div>
	</div>

	<!-- Navigation Controls -->
	<div
		class="flex overflow-x-auto no-scrollbar gap-1 mb-6 shrink-0 bg-slate-900/30 p-1 border border-white/5 rounded-xl"
	>
		{#each [['overview', BarChart3, 'Overview'], ['rules', ShieldCheck, 'Rules'], ['bans', Ban, 'Blocked IPs'], ['anticheat', AlertTriangle, 'Threats'], ['logs', Terminal, 'Logs'], ['config', Settings, 'Settings']] as [id, icon, label]}
			{@const Icon = icon as any}
			<button
				onclick={() => {
					activeTab = id as any;
					refreshAll();
				}}
				class="flex-1 min-w-[100px] flex items-center justify-center gap-2 py-2.5 transition-all text-xs font-bold uppercase tracking-wide rounded-lg {activeTab ===
				id
					? 'bg-rose-500 text-white shadow-lg shadow-rose-500/20'
					: 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'}"
			>
				<Icon class="w-4 h-4" />
				<span>{label}</span>
			</button>
		{/each}
	</div>

	<!-- Main Display -->
	<div class="flex-1 min-h-0 relative z-10">
		{#if activeTab === 'overview'}
			<div class="h-full overflow-auto space-y-6 custom-scrollbar pr-0 sm:pr-2">
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					{#each [{ icon: ShieldCheck, val: stats.total_rules, label: 'Active Rules', color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20' }, { icon: Ban, val: stats.active_bans, label: 'Blocked IPs', color: 'text-rose-400', bg: 'bg-rose-500/10', border: 'border-rose-500/20' }, { icon: ShieldAlert, val: stats.events_24h, label: 'Alerts (24h)', color: 'text-amber-400', bg: 'bg-amber-500/10', border: 'border-amber-500/20' }, { icon: Activity, val: stats.logs_24h, label: 'Requests (24h)', color: 'text-sky-400', bg: 'bg-sky-500/10', border: 'border-sky-500/20' }] as card}
						{@const CardIcon = card.icon as any}
						<div
							class="bg-slate-900/40 border border-white/5 rounded-xl p-5 flex items-center gap-4 hover:border-white/10 transition-all"
						>
							<div class="p-3 rounded-lg {card.bg} {card.border} border">
								<CardIcon class="w-6 h-6 {card.color}" />
							</div>
							<div class="flex flex-col">
								<span class="text-2xl font-bold text-white tabular-nums leading-none"
									>{card.val}</span
								>
								<span class="text-xs font-medium text-slate-500 uppercase tracking-wide mt-1"
									>{card.label}</span
								>
							</div>
						</div>
					{/each}
				</div>

				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Visual Threat Monitor -->
					<div
						class="lg:col-span-2 bg-slate-900/40 border border-white/5 rounded-2xl p-8 flex flex-col items-center justify-center min-h-[400px] relative overflow-hidden"
					>
						<!-- Soft Gradient Background -->
						<div
							class="absolute inset-0 bg-gradient-to-br from-rose-500/5 via-transparent to-slate-900/50 pointer-events-none"
						></div>

						<div class="relative z-10 w-full max-w-lg text-center">
							<div
								class="inline-flex items-center gap-2 px-3 py-1 bg-rose-500/10 border border-rose-500/20 rounded-full mb-8"
							>
								<div class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"></div>
								<span class="text-[10px] font-bold text-rose-400 uppercase tracking-wide"
									>Real-time Monitoring</span
								>
							</div>

							<div class="relative mx-auto w-48 h-48 mb-8">
								<!-- Clean Radar Animation -->
								<div class="absolute inset-0 border border-slate-700/30 rounded-full"></div>
								<div class="absolute inset-8 border border-slate-700/30 rounded-full"></div>
								<div class="absolute inset-16 border border-slate-700/30 rounded-full"></div>

								<!-- Scanning Line -->
								<div
									class="absolute inset-0 rounded-full bg-gradient-to-t from-rose-500/10 to-transparent animate-[spin_3s_linear_infinite]"
								></div>

								<div class="absolute inset-0 flex items-center justify-center">
									<Activity
										class="w-12 h-12 text-rose-500/80 drop-shadow-[0_0_15px_rgba(244,63,94,0.5)]"
									/>
								</div>
							</div>

							<h3 class="text-2xl font-bold text-white tracking-tight mb-2">
								Network Activity Monitor
							</h3>
							<p class="text-sm text-slate-400 font-medium leading-relaxed">
								System active. Latency: <span class="text-rose-400">1.2ms</span>. Monitoring
								<span class="text-white">{stats.reputation_count}</span> active security rules.
							</p>
						</div>
					</div>

					<!-- Quick Tech Stats -->
					<div class="space-y-6">
						<div
							class="bg-slate-900/40 border border-white/5 rounded-2xl p-6 h-full flex flex-col justify-between"
						>
							<div class="space-y-6">
								<h4
									class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-4 flex items-center gap-2"
								>
									<Cpu class="w-4 h-4 text-rose-400" /> Resource Usage
								</h4>
								<div class="space-y-4">
									{#each [['Processing', Number(Math.min(100, 5 + stats.logs_24h / 100)), 'text-emerald-400', 'bg-emerald-500'], ['Memory', Number(Math.min(100, 2 + stats.active_bans * 2)), 'text-rose-400', 'bg-rose-500'], ['Database', Number(Math.min(100, 1 + stats.reputation_count / 10)), 'text-indigo-400', 'bg-indigo-500']] as [label, perc, colorText, colorBg]}
										<div class="space-y-2">
											<div
												class="flex justify-between text-[10px] font-bold uppercase tracking-wide"
											>
												<span class="text-slate-400">{label}</span>
												<span class={colorText as string}>{Math.round(perc as number)}%</span>
											</div>
											<div class="h-1.5 bg-slate-800 rounded-full overflow-hidden">
												<div
													class="h-full {colorBg} rounded-full transition-all duration-1000"
													style="width: {perc}%;"
												></div>
											</div>
										</div>
									{/each}
								</div>
							</div>

							<div class="mt-8 pt-6 border-t border-white/5">
								<h4
									class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-4 flex items-center gap-2"
								>
									<Globe class="w-4 h-4 text-indigo-400" /> Network Status
								</h4>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-3">
										<div class="text-4xl font-bold text-white tracking-tight">08</div>
										<div class="text-[10px] font-bold text-slate-500 leading-tight uppercase">
											Active<br />Clusters
										</div>
									</div>
									<div class="flex gap-1">
										{#each Array(8) as _, i}
											<div
												class="h-6 w-1 bg-rose-500/50 rounded-full"
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
				class="h-full flex flex-col bg-slate-900/40 border border-white/5 rounded-2xl overflow-hidden"
			>
				<div class="p-6 border-b border-white/5 flex justify-between items-center bg-slate-900/50">
					<div class="flex items-center gap-3">
						<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
						<h3 class="text-sm font-bold text-white uppercase tracking-wider">Rule Registry</h3>
					</div>
					<button
						onclick={() => openModal()}
						class="px-4 py-2 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold uppercase tracking-wide rounded-lg flex items-center gap-2 transition-all shadow-lg shadow-rose-500/20"
					>
						<Plus class="w-4 h-4" /> Add Rule
					</button>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left text-xs">
						<thead class="bg-slate-900/80 text-slate-500 sticky top-0 z-10 border-b border-white/5">
							<tr class="uppercase font-bold tracking-wide">
								<th class="px-6 py-4 font-bold">Rule Name</th>
								<th class="px-6 py-4 font-bold">Target Specification</th>
								<th class="px-6 py-4 text-center font-bold">Action</th>
								<th class="px-6 py-4 text-right font-bold">Manage</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-white/5">
							{#each rules as rule (rule.id)}
								<tr class="hover:bg-white/5 transition-colors group">
									<td class="px-6 py-4">
										<div class="flex flex-col gap-0.5">
											<span
												class="font-bold text-slate-200 group-hover:text-rose-400 transition-colors"
												>{rule.name}</span
											>
											<span class="text-[10px] text-slate-500 font-mono">CIDR: {rule.cidr}</span>
										</div>
									</td>
									<td class="px-6 py-4">
										<div class="flex items-center gap-2">
											<span
												class="text-slate-400 bg-slate-800 px-2 py-0.5 rounded text-[10px] font-mono border border-white/5"
												>PORT: {rule.port}</span
											>
											{#if rule.path_pattern}
												<span class="text-slate-400 text-[10px] font-mono truncate max-w-[150px]"
													>{rule.path_pattern}</span
												>
											{/if}
										</div>
									</td>
									<td class="px-6 py-4 text-center">
										<span
											class={`px-2.5 py-1 text-[10px] font-bold rounded-md uppercase tracking-wide border ${getActionColor(
												rule.action
											)}`}
										>
											{rule.action}
										</span>
									</td>
									<td class="px-6 py-4">
										<div
											class="flex justify-end gap-2 opacity-60 group-hover:opacity-100 transition-all"
										>
											<button
												onclick={() => openModal(rule)}
												class="p-2 bg-slate-800 hover:bg-white/10 border border-white/5 rounded-lg text-slate-400 hover:text-white transition-all"
												aria-label="Edit rule"
											>
												<Edit2 class="w-3.5 h-3.5" />
											</button>
											<button
												onclick={() => deleteRule(rule.id)}
												class="p-2 bg-slate-800 hover:bg-rose-500/20 border border-white/5 rounded-lg text-slate-400 hover:text-rose-500 transition-all"
												aria-label="Delete rule"
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
			<div
				class="h-full flex flex-col bg-slate-900/40 border border-white/5 rounded-2xl overflow-hidden"
			>
				<div class="p-6 border-b border-white/5 flex justify-between items-center bg-slate-900/50">
					<div class="flex items-center gap-3">
						<Ban class="w-4 h-4 text-rose-500" />
						<h3 class="text-sm font-bold text-white uppercase tracking-wider">Blocked IPs</h3>
					</div>
					<div class="flex items-center gap-4">
						<div class="relative">
							<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
							<input
								type="text"
								placeholder="Search bans..."
								class="bg-black/40 border border-white/5 rounded-lg pl-9 pr-4 py-1.5 text-xs text-white focus:border-rose-500 outline-none transition-all w-48"
							/>
						</div>
					</div>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					{#if bans.length === 0}
						<div class="py-20 flex flex-col items-center justify-center opacity-30">
							<Ban class="w-12 h-12 mb-4" />
							<span class="text-xs font-bold uppercase tracking-widest">No active bans</span>
						</div>
					{:else}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4 p-6">
							{#each bans as ban}
								<div
									class="bg-black/20 border border-white/5 rounded-xl p-4 flex items-center justify-between group hover:border-rose-500/30 transition-all"
								>
									<div class="flex items-center gap-4">
										<div class="p-2 bg-rose-500/10 rounded-lg">
											<Globe class="w-4 h-4 text-rose-500" />
										</div>
										<div>
											<div class="text-sm font-mono text-white">{ban.ip}</div>
											<div class="text-[10px] text-slate-500 uppercase tracking-wide">
												Score: {ban.reputation_score} • Last Seen: {new Date(
													ban.last_seen
												).toLocaleDateString()}
											</div>
										</div>
									</div>
									<button
										onclick={() => unbanIP(ban.ip)}
										class="px-3 py-1.5 bg-slate-800 hover:bg-emerald-500/10 hover:text-emerald-400 border border-white/5 rounded-lg text-[10px] font-bold uppercase tracking-wide transition-all"
									>
										Release
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{:else if activeTab === 'anticheat'}
			<div class="h-full flex flex-col gap-6 overflow-auto custom-scrollbar pr-2">
				<Card title="Threat Intelligence" subtitle="Security event monitoring" icon="ph:warning-bold">
					<div class="p-0">
						{#if events.length === 0}
							<div class="py-20 flex flex-col items-center justify-center opacity-30">
								<ShieldAlert class="w-12 h-12 mb-4" />
								<span class="text-xs font-bold uppercase tracking-widest"
									>No security events detected</span
								>
							</div>
						{:else}
							<div class="divide-y divide-white/5">
								{#each events as event}
									<div class="p-6 flex items-center justify-between hover:bg-white/5 transition-all">
										<div class="flex items-center gap-4">
											<div
												class="p-2 rounded-lg {event.severity === 'high'
													? 'bg-rose-500/10 text-rose-500'
													: 'bg-amber-500/10 text-amber-500'}"
											>
												<AlertOctagon class="w-4 h-4" />
											</div>
											<div>
												<div class="text-xs font-bold text-white uppercase tracking-wider">
													{event.type}
												</div>
												<div class="text-[10px] text-slate-500 font-mono mt-0.5">
													User: {event.player_id} • Node: {event.node_id}
												</div>
											</div>
										</div>
										<div class="text-right">
											<div class="text-[10px] font-bold text-slate-400 uppercase">
												{new Date(event.created_at).toLocaleTimeString()}
											</div>
											<div
												class="text-[9px] font-mono mt-0.5 {event.severity === 'high'
													? 'text-rose-400'
													: 'text-amber-400'}"
											>
												{event.severity.toUpperCase()}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</Card>
			</div>
		{:else if activeTab === 'logs'}
			<div
				class="h-full bg-black/40 border border-white/5 rounded-2xl overflow-hidden flex flex-col font-mono"
			>
				<div class="p-4 border-b border-white/5 bg-slate-900/50 flex justify-between items-center">
					<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
						>Firewall Event Log</span
					>
					<div class="flex gap-4">
						<div class="flex items-center gap-2 text-[10px] text-slate-600">
							<div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div> ALLOW
						</div>
						<div class="flex items-center gap-2 text-[10px] text-slate-600">
							<div class="w-1.5 h-1.5 rounded-full bg-rose-500"></div> DENY
						</div>
					</div>
				</div>
				<div class="flex-1 overflow-auto p-4 space-y-1.5 custom-scrollbar text-[11px]">
					{#each logs as log}
						<div class="flex gap-4 group">
							<span class="text-slate-600 shrink-0"
								>[{new Date(log.created_at).toLocaleTimeString()}]</span
							>
							<span
								class="shrink-0 w-16 font-bold {log.action === 'ALLOW'
									? 'text-emerald-500'
									: 'text-rose-500'}">{log.action}</span
							>
							<span class="text-slate-400 shrink-0">{log.ip}</span>
							<span class="text-slate-500">→</span>
							<span class="text-slate-300 truncate">{log.path}</span>
							<span class="ml-auto text-slate-600 opacity-0 group-hover:opacity-100"
								>ID:{log.rule_id}</span
							>
						</div>
					{/each}
				</div>
			</div>
		{:else if activeTab === 'config'}
			<div class="h-full overflow-auto pr-2 custom-scrollbar">
				<div class="max-w-2xl space-y-8">
					<Card title="Global Configuration" subtitle="System-wide security parameters" icon="ph:sliders-bold">
						<div class="p-8 space-y-8">
							<div class="flex items-center justify-between">
								<div>
									<div class="text-sm font-bold text-white mb-1">Automated Banning</div>
									<p class="text-xs text-slate-500">Automatically block IPs exceeding threshold</p>
								</div>
								<button
									onclick={() =>
										(config['redeye.auto_ban_enabled'] = !config['redeye.auto_ban_enabled'])}
									class="w-12 h-6 rounded-full transition-all relative {config[
										'redeye.auto_ban_enabled'
									]
										? 'bg-rose-500'
										: 'bg-slate-800'}"
								>
									<div
										class="absolute top-1 w-4 h-4 bg-white rounded-full transition-all {config[
											'redeye.auto_ban_enabled'
										]
											? 'left-7'
											: 'left-1'}"
									></div>
								</button>
							</div>

							<div class="space-y-4">
								<div class="flex justify-between items-end">
									<label for="banThreshold" class="text-xs font-bold text-slate-400 uppercase"
										>Violation Threshold</label
									>
									<span class="text-sm font-mono text-rose-400"
										>{config['redeye.auto_ban_threshold']} req/min</span
									>
								</div>
								<input
									id="banThreshold"
									type="range"
									min="10"
									max="1000"
									step="10"
									bind:value={config['redeye.auto_ban_threshold']}
									class="w-full accent-rose-500"
								/>
							</div>

							<div class="pt-8 border-t border-white/5 flex justify-end">
								<Button variant="primary" onclick={updateConfig}>Save Configuration</Button>
							</div>
						</div>
					</Card>

					<div class="p-6 bg-slate-900/40 border border-white/5 rounded-2xl">
						<div class="flex items-center gap-4 mb-4">
							<div class="p-2 bg-indigo-500/10 rounded-lg">
								<RefreshCw class="w-4 h-4 text-indigo-400" />
							</div>
							<h4 class="text-xs font-bold text-white uppercase tracking-wider">Log Management</h4>
						</div>
						<p class="text-xs text-slate-500 leading-relaxed mb-6">
							Security logs are retained for 30 days by default. Performance metrics are aggregated
							daily.
						</p>
						<Button variant="secondary" size="sm">Download Security Audit</Button>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Rule Modal -->
	{#if showModal}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
			onclick={() => (showModal = false)}
			transition:fade={{ duration: 150 }}
		>
			<div
				class="bg-neutral-900 border border-white/10 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden relative"
				onclick={(e) => e.stopPropagation()}
				transition:scale={{ duration: 200, start: 0.95 }}
			>
				<div class="p-6 border-b border-white/5 flex justify-between items-center bg-neutral-950">
					<div>
						<h3 class="text-lg font-bold text-white uppercase tracking-tight">
							{editingRule ? 'Edit' : 'New'} Firewall Rule
						</h3>
						<p class="text-xs text-slate-500 mt-0.5">Define security parameters for traffic filtering</p>
					</div>
					<button
						onclick={() => (showModal = false)}
						class="p-2 text-slate-500 hover:text-white transition-all bg-white/5 rounded-lg hover:bg-white/10"
					>
						<X class="w-5 h-5" />
					</button>
				</div>

				<div class="p-6 space-y-5">
					<div class="space-y-1.5">
						<label for="ruleName" class="text-xs font-bold text-slate-400 ml-1 uppercase"
							>Rule Name</label
						>
						<input
							id="ruleName"
							type="text"
							bind:value={form.name}
							placeholder="e.g. Block API Scraping"
							class="w-full bg-black border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all placeholder:text-slate-800"
						/>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
						<div class="space-y-1.5">
							<label for="ruleCidrIp" class="text-xs font-bold text-slate-400 ml-1 uppercase"
								>Source CIDR/IP</label
							>
							<input
								id="ruleCidrIp"
								type="text"
								bind:value={form.cidr}
								placeholder="0.0.0.0/0"
								class="w-full bg-black border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all font-mono"
							/>
						</div>
						<div class="space-y-1.5">
							<label for="rulePort" class="text-xs font-bold text-slate-400 ml-1 uppercase">Target Port</label>
							<input
								id="rulePort"
								type="text"
								bind:value={form.port}
								class="w-full bg-black border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all text-center font-mono"
								placeholder="*"
							/>
						</div>
					</div>

					<div class="space-y-1.5">
						<label for="rulePathPattern" class="text-xs font-bold text-slate-400 ml-1 uppercase"
							>Path Pattern (Regex)</label
						>
						<input
							id="rulePathPattern"
							type="text"
							bind:value={form.path_pattern}
							placeholder="/api/v1/.*"
							class="w-full bg-black border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all font-mono"
						/>
					</div>

					<div class="space-y-2">
						<span class="text-xs font-bold text-slate-400 ml-1 block uppercase">Action Enforcement</span>
						<div class="grid grid-cols-3 gap-2">
							{#each ['ALLOW', 'DENY', 'RATE_LIMIT'] as action}
								<button
									onclick={() => (form.action = action as any)}
									class="py-2.5 border text-[10px] font-black rounded-lg transition-all {form.action ===
									action
										? action === 'ALLOW'
											? 'bg-emerald-500/10 border-emerald-500 text-emerald-500'
											: action === 'DENY'
												? 'bg-rose-500/10 border-rose-500 text-rose-500'
												: 'bg-amber-500/10 border-amber-500 text-amber-500'
										: 'bg-slate-800 border-slate-700 text-slate-400 hover:bg-slate-700'}"
								>
									{action.replace('_', ' ')}
								</button>
							{/each}
						</div>
					</div>

					{#if form.action === 'RATE_LIMIT'}
						<div
							class="grid grid-cols-2 gap-5 bg-amber-500/5 p-4 rounded-xl border border-amber-500/10"
							transition:slide
						>
							<div class="space-y-1.5">
								<label for="ruleLimit" class="text-[10px] font-bold text-amber-500/80 uppercase"
									>Requests/sec</label
								>
								<input
									id="ruleLimit"
									type="number"
									bind:value={form.rate_limit}
									class="w-full bg-black border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all"
								/>
							</div>
							<div class="space-y-1.5">
								<label for="ruleBurst" class="text-[10px] font-bold text-amber-500/80 uppercase"
									>Burst Capacity</label
								>
								<input
									id="ruleBurst"
									type="number"
									bind:value={form.burst}
									class="w-full bg-black border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all"
								/>
							</div>
						</div>
					{/if}
				</div>

				<div class="p-6 border-t border-white/5 bg-neutral-950 flex justify-end gap-3">
					<button
						onclick={() => (showModal = false)}
						class="px-5 py-2.5 text-xs font-bold text-slate-400 hover:text-white transition-all uppercase tracking-widest"
					>
						Cancel
					</button>
					<button
						onclick={saveRule}
						class="px-6 py-2.5 bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold rounded-lg transition-all shadow-lg shadow-rose-500/20 uppercase tracking-widest"
					>
						Save Rule
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
