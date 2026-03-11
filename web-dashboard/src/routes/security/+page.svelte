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
	import {
		notifications,
		type SecurityRule,
		type SecurityLog,
		type SecurityEvent,
		type IPReputation
	} from '$lib/stores.svelte';

	let activeTab = $state<'overview' | 'rules' | 'bans' | 'integrity' | 'logs' | 'config'>(
		'overview'
	);
	let loading = $state(false);
	let savingRule = $state(false);
	let rules = $state<SecurityRule[]>([]);
	let bans = $state<IPReputation[]>([]);
	let events = $state<SecurityEvent[]>([]);
	let logs = $state<SecurityLog[]>([]);
	let stats = $state<any>({});
	let metrics = $state<any>(null);
	let config = $state<any>({
		'security.auto_ban_enabled': true,
		'security.auto_ban_threshold': 100,
		'security.alert_enabled': true
	});

	// Modal state
	let showModal = $state(false);
	let editingRule = $state<SecurityRule | null>(null);
	let form = $state<Partial<SecurityRule>>({
		id: 0,
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

	async function safeParseJson(res: Response, label: string) {
		const text = await res.text();
		try {
			return JSON.parse(text);
		} catch (e) {
			console.error(`[Security Hub] Failed to parse JSON from ${label}. Raw data:`, text);
			throw e;
		}
	}

	async function refreshAll() {
		loading = true;
		try {
			const [sRes, rRes, bRes, eRes, lRes, cRes, mRes] = await Promise.all([
				apiFetch('/api/security/stats'),
				apiFetch('/api/security/rules'),
				apiFetch('/api/security/bans'),
				apiFetch('/api/security/integrity/events'),
				apiFetch('/api/security/logs'),
				apiFetch('/api/security/config'),
				apiFetch('/api/metrics')
			]);

			const parse = async (res: Response, name: string) => {
				if (!res.ok) return null;
				const text = await res.text();
				try {
					return JSON.parse(text);
				} catch (e) {
					console.error(`[Security Hub] Failed to parse ${name} JSON. Raw response:`, text);
					// If it's a "position 4" error, let's see exactly what's at index 4
					if (text.length > 4) {
						console.error(`[Security Hub] Char at index 4: '${text[4]}' (Code: ${text.charCodeAt(4)})`);
					}
					return null;
				}
			};

			const sData = await parse(sRes, 'stats'); if (sData) stats = sData;
			const rData = await parse(rRes, 'rules'); if (rData) rules = rData;
			const bData = await parse(bRes, 'bans'); if (bData) bans = bData;
			const eData = await parse(eRes, 'events'); if (eData) events = eData.events || [];
			const lData = await parse(lRes, 'logs'); if (lData) logs = lData.logs || [];
			const cData = await parse(cRes, 'config'); if (cData) config = cData;
			const mData = await parse(mRes, 'metrics'); if (mData) metrics = mData;
		} catch (e) {
			console.error('Failed to refresh security data', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		refreshAll();
		const interval = setInterval(refreshAll, 30000);
		return () => clearInterval(interval);
	});

	function openModal(rule: SecurityRule | null = null) {
		editingRule = rule;
		if (rule) {
			form = { ...rule };
		} else {
			form = {
				id: 0,
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
		savingRule = true;
		try {
			const isEdit = editingRule !== null && editingRule.id > 0;
			const ruleId = editingRule?.id;
			const url = isEdit ? `/api/security/rules/${ruleId}` : '/api/security/rules';
			
			// Clean up form for JSON transmission
			const payload = { ...form };
			if (!isEdit) delete payload.id;
			// created_at is not needed for save/update
			if (payload.created_at) delete payload.created_at;

			const res = await apiFetch(url, {
				method: isEdit ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});

			if (res.ok) {
				notifications.add({ type: 'success', message: 'Security rule saved successfully' });
				showModal = false;
				refreshAll();
			} else {
				let errData = { error: 'Unknown error' };
				const rawText = await res.text();
				// DEBUG LOG
				console.error('[DEBUG] Raw Response Snippet:', rawText.substring(0, 100));
				
				try {
					errData = JSON.parse(rawText);
				} catch (e) {
					errData = { error: rawText || `Server returned ${res.status}` };
				}
				notifications.add({ 
					type: 'error', 
					message: 'Failed to save security rule', 
					details: errData.error 
				});
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Connection error', details: e.message });
		} finally {
			savingRule = false;
		}
	}

	async function deleteRule(id: number) {
		if (!confirm('Are you sure you want to delete this security rule?')) return;
		try {
			const res = await apiFetch(`/api/security/rules/${id}`, { method: 'DELETE' });
			if (res.ok) {
				notifications.add({ type: 'success', message: 'Security rule deleted' });
				refreshAll();
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Failed to delete rule' });
		}
	}

	async function unbanIP(ip: string) {
		try {
			const res = await apiFetch(`/api/security/bans/${ip}`, { method: 'DELETE' });
			if (res.ok) {
				notifications.add({ type: 'success', message: `Access restored for ${ip}` });
				refreshAll();
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Failed to restore access' });
		}
	}

	async function updateConfig() {
		try {
			const res = await apiFetch('/api/security/config', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(config)
			});
			if (res.ok) {
				notifications.add({ type: 'success', message: 'Security configuration updated' });
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Failed to update security configuration' });
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
						System <span class="text-rose-500">Security</span>
					</h1>
					<div
						class="px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wide {stats.system_active
							? 'bg-rose-500/10 text-rose-500 border border-rose-500/20'
							: 'bg-slate-800 text-slate-400 border border-slate-700'}"
					>
						{stats.system_active ? 'Online' : 'Disabled'}
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
					>Risk Assessment</span
				>
				<span class="text-lg font-mono font-semibold text-rose-400"
					>{stats.risk_factor?.toFixed(2) || '0.00'}%</span
				>
			</div>
			<Button
				onclick={refreshAll}
				variant="secondary"
				size="md"
				icon="ph:arrows-clockwise-bold"
				loading={loading}
				class="!p-2.5"
			/>
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
			<span class="text-slate-500">Session Uptime:</span>
			<span class="text-slate-300">{stats.uptime}</span>
		</div>
		<div class="flex items-center gap-2 text-xs font-medium shrink-0">
			<span class="text-slate-500">Threat Evaluation:</span>
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

	<div
		class="flex overflow-x-auto no-scrollbar gap-1 mb-6 shrink-0 bg-slate-900/30 p-1 border border-white/5 rounded-xl"
	>
		{#each [['overview', BarChart3, 'Overview'], ['rules', ShieldCheck, 'Security Rules'], ['bans', Ban, 'Blocked Access'], ['integrity', AlertTriangle, 'Integrity'], ['logs', Terminal, 'Access Logs'], ['config', Settings, 'Settings']] as [id, icon, label]}
			{@const Icon = icon as any}
			<Button
				variant="ghost"
				active={activeTab === id}
				size="sm"
				onclick={() => {
					activeTab = id as any;
					refreshAll();
				}}
				class="flex-1 min-w-[100px] !rounded-lg"
			>
				<Icon class="w-4 h-4" />
				<span>{label}</span>
			</Button>
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
								Network Operations Monitor
							</h3>
							<p class="text-sm text-slate-400 font-medium leading-relaxed">
								System active. Latency: <span class="text-rose-400">{metrics?.network?.avg_response_time_ms?.toFixed(2) || '0.00'}ms</span>. Monitoring
								<span class="text-white">{stats.reputation_count}</span> active reputation profiles.
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
									<Cpu class="w-4 h-4 text-rose-400" /> Resource Allocation
								</h4>
								<div class="space-y-4">
									{#each [['Processing', Number(Math.min(100, (stats.rt_queue_depth || 0) / 10)), 'text-emerald-400', 'bg-emerald-500'], ['Memory', Number(Math.min(100, (metrics?.master?.heap_alloc || 0) / 50000000 * 100)), 'text-rose-400', 'bg-rose-500'], ['Integrity', Number(Math.min(100, stats.risk_factor * 100)), 'text-indigo-400', 'bg-indigo-500']] as [label, perc, colorText, colorBg]}
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
									<Globe class="w-4 h-4 text-indigo-400" /> Security Mesh
								</h4>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-3">
										<div class="text-4xl font-bold text-white tracking-tight">
											{stats.rt_active_trackers?.toString().padStart(2, '0') || '00'}
										</div>
										<div class="text-[10px] font-bold text-slate-500 leading-tight uppercase">
											Active<br />Trackers
										</div>
									</div>
									<div class="flex gap-1">
										{#each Array(8) as _, i}
											<div
												class="h-6 w-1 bg-rose-500/50 rounded-full"
												style="opacity: {(stats.rt_active_trackers > i) ? (0.2 + i * 0.1) : 0.05}"
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
						<h3 class="text-sm font-bold text-white uppercase tracking-wider">Security Rule Registry</h3>
					</div>
					<button
						onclick={() => openModal()}
						class="px-4 py-2 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold uppercase tracking-wide rounded-lg flex items-center gap-2 transition-all shadow-lg shadow-rose-500/20"
					>
						<Plus class="w-4 h-4" /> New Rule
					</button>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left text-xs">
						<thead class="bg-slate-900/80 text-slate-500 sticky top-0 z-10 border-b border-white/5">
							<tr class="uppercase font-bold tracking-wide">
								<th class="px-6 py-4 font-bold">Rule Definition</th>
								<th class="px-6 py-4 font-bold">Target Specification</th>
								<th class="px-6 py-4 text-center font-bold">Action</th>
								<th class="px-6 py-4 text-right font-bold">Operations</th>
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
						<h3 class="text-sm font-bold text-white uppercase tracking-wider">Blocked Access List</h3>
					</div>
					<div class="flex items-center gap-4">
						<div class="relative">
							<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
							<input
								type="text"
								placeholder="Filter entries..."
								class="bg-black/40 border border-white/5 rounded-lg pl-9 pr-4 py-1.5 text-xs text-white focus:border-rose-500 outline-none transition-all w-48"
							/>
						</div>
					</div>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					{#if bans.length === 0}
						<div class="py-20 flex flex-col items-center justify-center opacity-30">
							<Ban class="w-12 h-12 mb-4" />
							<span class="text-xs font-bold uppercase tracking-widest">No active restrictions</span>
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
												Risk Score: {ban.reputation_score} • Observed: {new Date(
													ban.last_seen
												).toLocaleDateString()}
											</div>
										</div>
									</div>
									<button
										onclick={() => unbanIP(ban.ip)}
										class="px-3 py-1.5 bg-slate-800 hover:bg-emerald-500/10 hover:text-emerald-400 border border-white/5 rounded-lg text-[10px] font-bold uppercase tracking-wide transition-all"
									>
										Restore Access
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{:else if activeTab === 'integrity'}
			<div class="h-full flex flex-col gap-6 overflow-auto custom-scrollbar pr-2">
				<Card title="Threat Intelligence" subtitle="Integrity event monitoring" icon="ph:warning-bold">
					<div class="p-0">
						{#if events.length === 0}
							<div class="py-20 flex flex-col items-center justify-center opacity-30">
								<ShieldAlert class="w-12 h-12 mb-4" />
								<span class="text-xs font-bold uppercase tracking-widest"
									>No integrity events detected</span
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
						>Security Access Log</span
					>
					<div class="flex gap-4">
						<div class="flex items-center gap-2 text-[10px] text-slate-600">
							<div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div> AUTHORIZED
						</div>
						<div class="flex items-center gap-2 text-[10px] text-slate-600">
							<div class="w-1.5 h-1.5 rounded-full bg-rose-500"></div> RESTRICTED
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
								class="shrink-0 w-24 font-bold {log.action === 'ALLOW'
									? 'text-emerald-500'
									: 'text-rose-500'}">{log.action === 'ALLOW' ? 'AUTHORIZED' : 'RESTRICTED'}</span
							>
							<span class="text-slate-400 shrink-0">{log.ip}</span>
							<span class="text-slate-500">→</span>
							<span class="text-slate-300 truncate">{log.path}</span>
							<span class="ml-auto text-slate-600 opacity-0 group-hover:opacity-100"
								>RULE:{log.rule_id}</span
							>
						</div>
					{/each}
				</div>
			</div>
		{:else if activeTab === 'config'}
			<div class="h-full overflow-auto pr-2 custom-scrollbar pb-20">
				<div class="grid grid-cols-1 xl:grid-cols-2 gap-6">
					<!-- Core Engine Mode -->
					<Card title="Operation Mode" subtitle="Primary engine behavior" icon="ph:gear-bold">
						<div class="p-6 space-y-6">
							<div class="flex flex-col gap-4">
								{#each ['ENFORCEMENT', 'SIMULATION'] as mode}
									<button
										onclick={() => (config['security.mode'] = mode)}
										class="flex items-center justify-between p-4 rounded-xl border-2 transition-all {config['security.mode'] === mode ? 'bg-sky-500/10 border-sky-500 text-white' : 'bg-slate-900/40 border-white/5 text-slate-500 hover:border-white/10'}"
									>
										<div class="flex items-center gap-4">
											<div class="p-2 rounded-lg {config['security.mode'] === mode ? 'bg-sky-500/20' : 'bg-slate-800'}">
												<Icon name={mode === 'ENFORCEMENT' ? 'ph:shield-check-fill' : 'ph:eye-fill'} size="1.2rem" />
											</div>
											<div class="text-left">
												<div class="text-xs font-bold uppercase tracking-wider">{mode}</div>
												<div class="text-[10px] opacity-60 font-medium">
													{mode === 'ENFORCEMENT' ? 'Active blocking and enforcement' : 'Audit only - no traffic restriction'}
												</div>
											</div>
										</div>
										{#if config['security.mode'] === mode}
											<div class="w-2 h-2 rounded-full bg-sky-500 shadow-[0_0_10px_rgba(14,165,233,0.5)]"></div>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					</Card>

					<!-- Protection Parameters -->
					<Card title="Protection Parameters" subtitle="Sensitivity & decay logic" icon="ph:sliders-horizontal-bold">
						<div class="p-6 space-y-8">
							<div class="flex items-center justify-between">
								<div>
									<div class="text-xs font-bold text-white mb-1 uppercase tracking-wide">Strict Mode</div>
									<p class="text-[10px] text-slate-500 font-medium leading-relaxed">Double reputation penalties for all signals</p>
								</div>
								<button
									onclick={() => (config['security.strict_mode'] = !config['security.strict_mode'])}
									class="w-10 h-5 rounded-full transition-all relative {config['security.strict_mode'] ? 'bg-rose-500' : 'bg-slate-800'}"
								>
									<div class="absolute top-1 w-3 h-3 bg-white rounded-full transition-all {config['security.strict_mode'] ? 'left-6' : 'left-1'}"></div>
								</button>
							</div>

							<div class="space-y-4">
								<div class="flex justify-between items-end">
									<label for="banThreshold" class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Violation Threshold</label>
									<span class="text-xs font-mono text-sky-400 font-bold">{config['security.auto_ban_threshold']} PTS</span>
								</div>
								<input id="banThreshold" type="range" min="10" max="500" step="5" bind:value={config['security.auto_ban_threshold']} class="w-full accent-sky-500" />
							</div>

							<div class="space-y-4">
								<div class="flex justify-between items-end">
									<label for="decayRate" class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Recovery Interval</label>
									<span class="text-xs font-mono text-emerald-400 font-bold">{config['security.decay_rate']} MIN</span>
								</div>
								<input id="decayRate" type="range" min="1" max="60" step="1" bind:value={config['security.decay_rate']} class="w-full accent-emerald-500" />
							</div>
						</div>
					</Card>

					<!-- Trusted Access -->
					<Card title="Trusted Access" subtitle="Whitelist & exclusions" icon="ph:lock-open-bold">
						<div class="p-6 space-y-4">
							<label class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Bypass IP List (CSV)</label>
							<textarea 
								bind:value={config['security.whitelist_ips']}
								placeholder="127.0.0.1, 192.168.1.1..."
								class="w-full h-24 bg-slate-950 border border-white/5 rounded-xl p-4 text-xs font-mono text-sky-300 focus:border-sky-500 outline-none transition-all resize-none shadow-inner"
							></textarea>
							<p class="text-[9px] text-slate-600 italic">Whitelisted IPs bypass all firewall and reputation checks.</p>
						</div>
					</Card>

					<!-- Geographic Boundaries -->
					<Card title="Geographic Boundaries" subtitle="Regional traffic control" icon="ph:globe-hemisphere-west-bold">
						<div class="p-6 space-y-6">
							<div class="flex items-center justify-between mb-4">
								<div>
									<div class="text-xs font-bold text-white mb-1 uppercase tracking-wide">GeoIP Filtering</div>
									<p class="text-[10px] text-slate-500 font-medium">Restrict traffic by country of origin</p>
								</div>
								<button
									onclick={() => (config['security.geoip_enabled'] = !config['security.geoip_enabled'])}
									class="w-10 h-5 rounded-full transition-all relative {config['security.geoip_enabled'] ? 'bg-indigo-500' : 'bg-slate-800'}"
								>
									<div class="absolute top-1 w-3 h-3 bg-white rounded-full transition-all {config['security.geoip_enabled'] ? 'left-6' : 'left-1'}"></div>
								</button>
							</div>

							<div class="space-y-4 {config['security.geoip_enabled'] ? 'opacity-100' : 'opacity-30 pointer-events-none transition-opacity'}">
								<label class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Allowed Country Codes</label>
								<input 
									type="text" 
									bind:value={config['security.allowed_countries']}
									placeholder="US, GB, DE..."
									class="w-full bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-xs font-mono text-indigo-300 focus:border-indigo-500 outline-none transition-all shadow-inner"
								/>
							</div>
						</div>
					</Card>
				</div>

				<!-- Save Action -->
				<div class="mt-8 p-6 bg-sky-500/5 border border-sky-500/20 rounded-2xl flex flex-col sm:flex-row items-center justify-between gap-6">
					<div class="flex items-center gap-4">
						<div class="p-3 bg-sky-500/20 rounded-xl">
							<RefreshCw class="w-6 h-6 text-sky-400" />
						</div>
						<div>
							<div class="text-sm font-bold text-white">Pending Configuration Sync</div>
							<p class="text-xs text-slate-500">Changes are applied immediately to all security processors.</p>
						</div>
					</div>
					<Button variant="primary" size="lg" onclick={updateConfig} class="min-w-[200px]">Synchronize Engine</Button>
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
							{editingRule ? 'Edit' : 'New'} Security Rule
						</h3>
						<p class="text-xs text-slate-500 mt-0.5">Define parameters for traffic filtering and protection</p>
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
							>Rule Designation</label
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
								>Source Address (CIDR)</label
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
							>Resource Pattern (Regex)</label
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
						<span class="text-xs font-bold text-slate-400 ml-1 block uppercase">Access Control Action</span>
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
									>Requests per Second</label
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
									>Burst Tolerance</label
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
					<Button
						variant="ghost"
						onclick={() => (showModal = false)}
						class="!rounded-lg uppercase tracking-widest text-[10px]"
					>
						Cancel
					</Button>
					<Button
						variant="primary"
						onclick={saveRule}
						loading={savingRule}
						class="px-8 !rounded-lg uppercase tracking-widest text-[10px] shadow-lg shadow-rose-500/20"
					>
						Save Security Rule
					</Button>
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
