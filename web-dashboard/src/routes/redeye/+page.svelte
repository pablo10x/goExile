<div class="w-full h-full flex flex-col overflow-hidden relative font-sans">
	
	<!-- Header -->
	<div class="flex flex-col lg:flex-row lg:justify-between lg:items-center mb-8 gap-6 shrink-0 relative z-10">
		<div class="flex items-center gap-4">
			<div class="p-3 bg-rose-500/10 rounded-xl border border-rose-500/20">
				<Icon name="ph:eye-bold" size="1.8rem" class="text-rose-500" />
			</div>
			<div>
				<div class="flex items-center gap-3">
					<h1 class="text-3xl font-bold text-white tracking-tight">
						RedEye <span class="text-rose-500">Security</span>
					</h1>
					<div
						class="px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wide {stats.system_active
							? 'bg-rose-500/10 text-rose-500 border border-rose-500/20'
							: 'bg-slate-800 text-slate-400 border border-slate-700'}"
					>
						{stats.system_active ? 'Active' : 'Offline'}
					</div>
				</div>
				<p class="text-xs font-medium text-slate-400 mt-1">
					Autonomous Threat Mitigation System
				</p>
			</div>
		</div>

		<div class="flex items-center gap-6">
			<div class="flex flex-col items-end">
				<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Entropy Delta</span>
				<span class="text-lg font-mono font-semibold text-rose-400">{stats.entropy?.toFixed(4) || '0.0000'}%</span>
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
				>Monitoring: <span
					class={stats.system_active ? 'text-slate-200' : 'text-slate-500'}
					>{stats.system_active ? 'Engaged' : 'Offline'}</span
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
				class={`px-2 py-0.5 rounded-md ${stats.threat_level === 'Low'
					? 'text-emerald-400 bg-emerald-500/10'
					: stats.threat_level === 'Medium'
						? 'text-amber-400 bg-amber-500/10'
						: 'text-rose-400 bg-rose-500/10'}`}>{stats.threat_level?.toUpperCase()}</span>
		</div>
		<div class="ml-auto text-xs font-mono text-slate-600 hidden lg:flex gap-4 shrink-0">
			<span>CRC: {stats.crc}</span>
			<span>NODE: {stats.node_id}</span>
		</div>
	</div>

	<!-- Navigation Controls -->
	<div
		class="flex overflow-x-auto no-scrollbar gap-1 mb-6 shrink-0 bg-slate-900/30 p-1 border border-white/5 rounded-xl"
	>
		{#each [['overview', BarChart3, 'Overview'], ['rules', ShieldCheck, 'Rules'], ['bans', Ban, 'Bans'], ['anticheat', AlertTriangle, 'Intel'], ['logs', Terminal, 'Logs'], ['config', Settings, 'Settings']] as [id, icon, label]}
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
					{#each [{ icon: ShieldCheck, val: stats.total_rules, label: 'Protocols', color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20' }, { icon: Ban, val: stats.active_bans, label: 'Quarantined', color: 'text-rose-400', bg: 'bg-rose-500/10', border: 'border-rose-500/20' }, { icon: ShieldAlert, val: stats.events_24h, label: 'Anomalies', color: 'text-amber-400', bg: 'bg-amber-500/10', border: 'border-amber-500/20' }, { icon: Activity, val: stats.logs_24h, label: 'Scans', color: 'text-sky-400', bg: 'bg-sky-500/10', border: 'border-sky-500/20' }] as card}
						{@const CardIcon = card.icon as any}
						<div
							class="bg-slate-900/40 border border-white/5 rounded-xl p-5 flex items-center gap-4 hover:border-white/10 transition-all"
						>
							<div
								class="p-3 rounded-lg {card.bg} {card.border} border"
							>
								<CardIcon class="w-6 h-6 {card.color}" />
							</div>
							<div class="flex flex-col">
								<span
									class="text-2xl font-bold text-white tabular-nums leading-none"
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
						<div class="absolute inset-0 bg-gradient-to-br from-rose-500/5 via-transparent to-slate-900/50 pointer-events-none"></div>

						<div class="relative z-10 w-full max-w-lg text-center">
							<div
								class="inline-flex items-center gap-2 px-3 py-1 bg-rose-500/10 border border-rose-500/20 rounded-full mb-8"
							>
								<div class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse"></div>
								<span class="text-[10px] font-bold text-rose-400 uppercase tracking-wide"
									>Live Surveillance</span>
							</div>

							<div class="relative mx-auto w-48 h-48 mb-8">
								<!-- Clean Radar Animation -->
								<div class="absolute inset-0 border border-slate-700/30 rounded-full"></div>
								<div class="absolute inset-8 border border-slate-700/30 rounded-full"></div>
								<div class="absolute inset-16 border border-slate-700/30 rounded-full"></div>
								
								<!-- Scanning Line -->
								<div class="absolute inset-0 rounded-full bg-gradient-to-t from-rose-500/10 to-transparent animate-[spin_3s_linear_infinite]"></div>
								
								<div class="absolute inset-0 flex items-center justify-center">
									<Activity class="w-12 h-12 text-rose-500/80 drop-shadow-[0_0_15px_rgba(244,63,94,0.5)]" />
								</div>
							</div>

							<h3
								class="text-2xl font-bold text-white tracking-tight mb-2"
							>
								Neural Defense Grid
							</h3>
							<p class="text-sm text-slate-400 font-medium leading-relaxed">
								System active. Latency: <span
									class="text-rose-400">1.2ms</span>. Monitoring
								<span class="text-white"
									>{stats.reputation_count}</span
								> identity signatures.
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
									<Cpu class="w-4 h-4 text-rose-400" /> Load Metrics
								</h4>
								<div class="space-y-4">
									{#each [['Heuristic', Number(Math.min(100, 5 + stats.logs_24h / 100)), 'text-emerald-400', 'bg-emerald-500'], ['Sequencer', Number(Math.min(100, 2 + stats.active_bans * 2)), 'text-rose-400', 'bg-rose-500'], ['Indexer', Number(Math.min(100, 1 + stats.reputation_count / 10)), 'text-indigo-400', 'bg-indigo-500']] as [label, perc, colorText, colorBg]}
										<div class="space-y-2">
											<div class="flex justify-between text-[10px] font-bold uppercase tracking-wide">
												<span class="text-slate-400">{label}</span>
												<span class={colorText as string}
													>{Math.round(perc as number)}%</span>
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
									<Globe class="w-4 h-4 text-indigo-400" /> Topology
								</h4>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-3">
										<div class="text-4xl font-bold text-white tracking-tight">08</div>
										<div class="text-[10px] font-bold text-slate-500 leading-tight uppercase">
											Active<br />Zones
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
				<div
					class="p-6 border-b border-white/5 flex justify-between items-center bg-slate-900/50"
				>
					<div class="flex items-center gap-3">
						<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
						<h3 class="text-sm font-bold text-white uppercase tracking-wider">
							Active Protocols
						</h3>
					</div>
					<button
						onclick={() => openModal()}
						class="px-4 py-2 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold uppercase tracking-wide rounded-lg flex items-center gap-2 transition-all shadow-lg shadow-rose-500/20"
					>
						<Plus class="w-4 h-4" /> New Protocol
					</button>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left text-xs">
						<thead class="bg-slate-900/80 text-slate-500 sticky top-0 z-10 border-b border-white/5">
							<tr class="uppercase font-bold tracking-wide">
								<th class="px-6 py-4 font-bold">ID / Name</th>
								<th class="px-6 py-4 font-bold">Target</th>
								<th class="px-6 py-4 text-center font-bold">Action</th>
								<th class="px-6 py-4 text-right font-bold">Controls</th>
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
											<span class="text-[10px] text-slate-500 font-mono"
												>{rule.cidr}</span
											>
										</div>
									</td>
									<td class="px-6 py-4">
										<div class="flex items-center gap-2">
											<span
												class="text-slate-400 bg-slate-800 px-2 py-0.5 rounded text-[10px] font-mono border border-white/5"
												>PORT:{rule.port}</span
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
												class="p-1.5 rounded-lg text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/10 transition-all"
											>
												<Edit2 class="w-4 h-4" />
											</button>
											<button
												onclick={() => deleteRule(rule.id)}
												class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-all"
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
			<div class="h-full overflow-auto grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-3 gap-4 custom-scrollbar pr-2">
				{#each bans as ban}
					<div
						class="bg-slate-900/40 border border-white/5 rounded-xl p-5 hover:border-rose-500/30 transition-all group"
					>
						<div class="flex justify-between items-start mb-4">
							<div class="flex items-center gap-4">
								<div
									class="p-3 bg-rose-500/10 rounded-lg text-rose-500 border border-rose-500/20"
								>
									<Lock class="w-5 h-5" />
								</div>
								<div>
									<h4
										class="text-lg font-mono font-bold text-slate-200"
									>
										{ban.ip}
									</h4>
									<div class="flex items-center gap-2 mt-1">
										<div class="w-16 h-1 bg-slate-800 rounded-full overflow-hidden">
											<div
												class="h-full bg-rose-500"
												style="width: {ban.reputation_score}%"
											></div>
										</div>
										<span class="text-[10px] font-bold text-slate-500 uppercase"
											>Risk: {ban.reputation_score}%</span>
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
						<div
							class="pt-4 border-t border-white/5 grid grid-cols-2 gap-4"
						>
							<div class="flex flex-col gap-1">
								<span class="text-[9px] text-slate-500 uppercase font-bold tracking-wider"
									>Reason</span
								>
								<p class="text-xs text-slate-300 line-clamp-1 italic">
									"{ban.ban_reason}"
								</p>
							</div>
							<div class="flex flex-col gap-1 text-right">
								<span class="text-[9px] text-slate-500 uppercase font-bold tracking-wider"
									>Expires</span
								>
								<span class="text-xs text-rose-400 font-medium"
									>{ban.ban_expires_at
										? new Date(ban.ban_expires_at).toLocaleDateString()
										: 'Permanent'}</span
								>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'anticheat'}
			<div class="h-full overflow-auto space-y-3 custom-scrollbar pr-2">
				{#each events as event}
					<div
						class="bg-slate-900/40 border border-white/5 rounded-xl p-4 flex items-center gap-6 hover:bg-slate-800/40 transition-colors group relative overflow-hidden"
					>
						<div class="absolute left-0 top-0 bottom-0 w-1 {event.severity > 80 ? 'bg-rose-500' : 'bg-slate-700'}"></div>

						<div class="text-center w-24 shrink-0 border-r border-white/5 pr-6">
							<div class="text-[10px] font-bold text-slate-500 uppercase">
								{new Date(event.timestamp).toLocaleDateString()}
							</div>
							<div class="text-xs font-mono font-medium text-slate-300 mt-0.5">
								{new Date(event.timestamp).toLocaleTimeString([], {
									hour: '2-digit',
									minute: '2-digit',
									second: '2-digit',
									hour12: false
								})}
							</div>
						</div>

						<div class="flex-1">
							<div class="flex items-center gap-3">
								<span
									class="text-sm font-bold text-white uppercase tracking-tight"
									>{event.event_type}</span
								>
								<div
									class="px-2 py-0.5 rounded bg-slate-950 border border-white/10 text-[10px] font-mono text-slate-400"
								>
									IP::{event.client_ip}
								</div>
							</div>
							<p class="text-xs font-medium text-slate-500 mt-1">
								Target: <span class="text-slate-300">{event.player_id || 'Unknown'}</span>
							</p>
						</div>

						<div class="flex flex-col items-end gap-1 pr-2">
							<span
								class={`text-[10px] font-bold uppercase tracking-wide ${event.severity > 80
									? 'text-rose-400'
									: 'text-slate-500'}`}>Threat Lvl: {event.severity}</span>
							<div class="flex gap-0.5">
								{#each Array(5) as _, i}
									<div
										class={`w-3 h-1 rounded-sm ${i < Math.ceil(event.severity / 20)
											? event.severity > 80
												? 'bg-rose-500'
												: 'bg-amber-500'
											: 'bg-slate-800'}`}
									></div>
								{/each}
							</div>
						</div>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'config'}
			<div class="h-full flex items-center justify-center p-4 overflow-auto">
				<div
					class="w-full max-w-2xl bg-slate-900/60 border border-white/5 rounded-2xl p-10 shadow-2xl relative overflow-hidden backdrop-blur-xl"
				>
					<div class="absolute top-0 right-0 p-6 opacity-[0.03] pointer-events-none">
						<Settings class="w-32 h-32" />
					</div>

					<h3
						class="text-xl font-bold text-white mb-8 flex items-center gap-3"
					>
						<div class="p-2 bg-rose-500/10 rounded-lg">
							<Zap class="w-5 h-5 text-rose-500" />
						</div>
						Security Configuration
					</h3>

					<div class="space-y-8">
						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-sm font-bold text-slate-200 group-hover:text-rose-400 transition-colors"
								>
									Auto-Ban System
								</h4>
								<p class="text-xs text-slate-500 mt-1 leading-relaxed">
									Automatically quarantine players who exceed the reputation risk threshold.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.auto_ban_enabled'] = !config['redeye.auto_ban_enabled'];
									updateConfig();
								}}
								class={`w-12 h-6 rounded-full transition-all relative border ${config[
									'redeye.auto_ban_enabled'
								]
									? 'bg-rose-500 border-rose-400'
									: 'bg-slate-800 border-slate-700'}`}
								aria-label="Toggle auto-ban"
							>
								<div
									class={`absolute top-0.5 left-0.5 w-4.5 h-4.5 rounded-full bg-white transition-all shadow-sm ${config[
										'redeye.auto_ban_enabled'
									]
										? 'translate-x-6'
										: ''}`}
								></div>
							</button>
						</div>

						<div class="space-y-4">
							<div class="flex justify-between items-end">
								<h4
									class="text-sm font-bold text-slate-200"
								>
									Sensitivity Threshold
								</h4>
								<span class="text-xl font-mono font-bold text-rose-400"
									>{config['redeye.auto_ban_threshold']}</span>
							</div>
							<div class="relative py-2">
								<input
									type="range"
									min="10"
									max="200"
									step="10"
									bind:value={config['redeye.auto_ban_threshold']}
									onchange={updateConfig}
									class="w-full h-1.5 bg-slate-800 rounded-full appearance-none cursor-pointer accent-rose-500"
									aria-label="Threshold slider"
								/>
							</div>
							<div class="flex justify-between text-[10px] font-bold text-slate-600 uppercase tracking-wide">
								<span>Strict</span>
								<span>Lenient</span>
							</div>
						</div>

						<div class="flex items-center justify-between group">
							<div class="max-w-[70%]">
								<h4
									class="text-sm font-bold text-slate-200"
								>
									Notification Feed
								</h4>
								<p class="text-xs text-slate-500 mt-1 leading-relaxed">
									Broadcast real-time intercept telemetry to decentralized command terminals.
								</p>
							</div>
							<button
								onclick={() => {
									config['redeye.alert_enabled'] = !config['redeye.alert_enabled'];
									updateConfig();
								}}
								class={`w-12 h-6 rounded-full transition-all relative border ${config['redeye.alert_enabled']
									? 'bg-emerald-500 border-emerald-400'
									: 'bg-slate-800 border-slate-700'}`}
								aria-label="Toggle alerts"
							>
								<div
									class={`absolute top-0.5 left-0.5 w-4.5 h-4.5 rounded-full bg-white transition-all shadow-sm ${config[
										'redeye.alert_enabled'
									]
										? 'translate-x-6'
										: ''}`}
								></div>
							</button>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Logs -->
			<div
				class="h-full flex flex-col bg-slate-900/40 border border-white/5 rounded-2xl overflow-hidden"
			>
				<div
					class="p-5 border-b border-white/5 bg-slate-900/50 flex items-center justify-between"
				>
					<div class="flex items-center gap-3">
						<Terminal class="w-4 h-4 text-sky-400" />
						<h3 class="text-sm font-bold text-white uppercase tracking-wide">
							Event Log
						</h3>
					</div>
					<div class="text-[10px] font-mono text-slate-500">
						LIVE_STREAM::ETH0
					</div>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar font-mono text-xs p-2">
					<div class="space-y-1">
						{#each logs as log}
							<div
								class="flex items-center gap-4 py-2 px-3 rounded-lg hover:bg-white/5 transition-colors group text-slate-400"
							>
								<span class="text-slate-600 shrink-0"
									>{new Date(log.timestamp).toLocaleTimeString([], { hour12: false })}</span
								>
								<span class={`font-bold w-24 shrink-0 ${getActionColor(log.action).split(' ')[0]}`}
									>{log.action}</span
								>
								<span class="text-slate-500 font-medium shrink-0">IP:</span>
								<span
									class="text-slate-300 group-hover:text-white transition-colors"
									>{log.source_ip}</span
								>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{:/if}
	</div>
</div>

<!-- Modal -->
{#if showModal}
	<div
		class="fixed inset-0 z-[500] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={() => (showModal = false)}
		onkeydown={(e) => e.key === 'Escape' && (showModal = false)}
		role="button"
		tabindex="0"
		aria-label="Close modal"
	>
		<div
			class="bg-slate-900 border border-white/10 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden relative"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<div
				class="p-6 border-b border-white/5 flex justify-between items-center bg-slate-800/50"
			>
				<div>
					<h3 class="text-lg font-bold text-white">
						{editingRule ? 'Edit' : 'New'} Protocol
					</h3>
					<p class="text-xs text-slate-500 mt-0.5">
						Define firewall rule parameters
					</p>
				</div>
				<button
					onclick={() => (showModal = false)}
					class="p-2 text-slate-500 hover:text-white transition-all bg-white/5 rounded-lg hover:bg-white/10">
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-6 space-y-5">
				<div class="space-y-1.5">
					<label
						for="ruleName"
						class="text-xs font-bold text-slate-400 ml-1"
						>Rule Name</label
					>
					<input
						id="ruleName"
						type="text"
						bind:value={form.name}
						placeholder="e.g. Block Malicious Subnet"
						class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all placeholder:text-slate-600"
					/>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
					<div class="space-y-1.5">
						<label
							for="ruleCidrIp"
							class="text-xs font-bold text-slate-400 ml-1"
							>Target IP / CIDR</label
						>
						<input
							id="ruleCidrIp"
							type="text"
							bind:value={form.cidr}
							placeholder="0.0.0.0/0"
							class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all placeholder:text-slate-600"
						/>
					</div>
					<div class="space-y-1.5">
						<label
							for="rulePort"
							class="text-xs font-bold text-slate-400 ml-1"
							>Port</label
						>
						<input
							id="rulePort"
							type="text"
							bind:value={form.port}
							class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all text-center placeholder:text-slate-600"
							placeholder="*"
						/>
					</div>
				</div>

				<div class="space-y-1.5">
					<label
						for="rulePathPattern"
						class="text-xs font-bold text-slate-400 ml-1"
						>Path Pattern (Optional)</label
					>
					<input
						id="rulePathPattern"
						type="text"
						bind:value={form.path_pattern}
						placeholder="/v1/api/..."
						class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-indigo-500 transition-all placeholder:text-slate-600"
					/>
				</div>

				<div class="space-y-2">
					<span class="text-xs font-bold text-slate-400 ml-1 block"
						>Action</span
					>
					<div class="grid grid-cols-3 gap-2">
						{#each ['ALLOW', 'DENY', 'RATE_LIMIT'] as action}
							<button
								onclick={() => (form.action = action as any)}
								class="py-2.5 border text-xs font-bold rounded-lg transition-all {form.action ===
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
							<label
								for="ruleLimit"
								class="text-[10px] font-bold text-amber-500/80 uppercase"
								>Rate (req/s)</label
							>
							<input
								id="ruleLimit"
								type="number"
								bind:value={form.rate_limit}
								class="w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all"
							/>
						</div>
						<div class="space-y-1.5">
							<label
								for="ruleBurst"
								class="text-[10px] font-bold text-amber-500/80 uppercase"
								>Burst</label
							>
							<input
								id="ruleBurst"
								type="number"
								bind:value={form.burst}
								class="w-full bg-slate-950 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-amber-500 transition-all"
							/>
						</div>
					</div>
				{/if}
			</div>

			<div
				class="p-6 border-t border-white/5 bg-slate-800/30 flex justify-end gap-3"
			>
				<button
					onclick={() => (showModal = false)}
					class="px-5 py-2.5 text-xs font-bold text-slate-400 hover:text-white transition-all"
				>
					Cancel
				</button>
				<button
					onclick={saveRule}
					class="px-6 py-2.5 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold rounded-lg transition-all shadow-lg shadow-rose-500/20"
				>
					Save Protocol
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
		background: #334155;
		border-radius: 99px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569;
	}
</style>
