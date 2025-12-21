<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { Users, Search, RefreshCw, Circle, FileText, AlertOctagon, Server } from 'lucide-svelte';

	interface Player {
		id: number;
		uid: string;
		name: string;
		device_id: string;
		xp: number;
		last_joined_server: string;
		created_at: string;
		updated_at: string;
		online: boolean;
	}

	interface Report {
		id: number;
		reporter_id: number;
		reported_user_id: number;
		reason: string;
		game_server_instance_id: string;
		timestamp: string;
		reporter_name?: string;
		reported_user_name?: string;
	}

	let activeTab = $state<'players' | 'reports'>('players');
	
	// Players State
	let players = $state<Player[]>([]);
	let playersLoading = $state(true);
	let playerSearchQuery = $state('');

	// Reports State
	let reports = $state<Report[]>([]);
	let reportsLoading = $state(true);
	let reportSearchQuery = $state('');

	let filteredPlayers = $derived.by(() => {
		if (!playerSearchQuery.trim()) return players;
		const query = playerSearchQuery.toLowerCase();
		return players.filter(
			(p) =>
				p.name.toLowerCase().includes(query) ||
				p.uid.toLowerCase().includes(query) ||
				p.device_id.toLowerCase().includes(query)
		);
	});

	let filteredReports = $derived.by(() => {
		if (!reportSearchQuery.trim()) return reports;
		const query = reportSearchQuery.toLowerCase();
		return reports.filter(
			(r) =>
				r.reason.toLowerCase().includes(query) ||
				(r.reporter_name?.toLowerCase() || '').includes(query) ||
				(r.reported_user_name?.toLowerCase() || '').includes(query) ||
				r.game_server_instance_id.toLowerCase().includes(query)
		);
	});

	async function fetchPlayers() {
		playersLoading = true;
		try {
			const res = await fetch('/api/game/players');
			if (res.ok) {
				players = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch players', e);
		} finally {
			playersLoading = false;
		}
	}

	async function fetchReports() {
		reportsLoading = true;
		try {
			const res = await fetch('/api/reports');
			if (res.ok) {
				reports = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch reports', e);
		} finally {
			reportsLoading = false;
		}
	}

	function refreshCurrentTab() {
		if (activeTab === 'players') {
			fetchPlayers();
		} else {
			fetchReports();
		}
	}

	onMount(() => {
		fetchPlayers();
		fetchReports();
		// Poll for online status updates every 10 seconds
		const interval = setInterval(() => {
			if (activeTab === 'players') fetchPlayers();
		}, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
	<!-- Header -->
	<div class="mb-8 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
		<div>
			<h1 class="text-3xl font-bold text-white flex items-center gap-3">
				<Users class="w-8 h-8 text-blue-400" />
				User Management
			</h1>
			<p class="text-slate-400 mt-1">Manage players and review reports</p>
		</div>

		<div class="flex items-center gap-3 w-full md:w-auto">
			<!-- Tab Switcher -->
			<div class="flex bg-slate-800/50 p-1 rounded-xl border border-slate-700">
				<button
					onclick={() => (activeTab = 'players')}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2 {activeTab ===
					'players'
						? 'bg-blue-600 text-white shadow-lg'
						: 'text-slate-400 hover:text-white'}"
				>
					<Users class="w-4 h-4" />
					Players
				</button>
				<button
					onclick={() => (activeTab = 'reports')}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2 {activeTab ===
					'reports'
						? 'bg-red-600 text-white shadow-lg'
						: 'text-slate-400 hover:text-white'}"
				>
					<AlertOctagon class="w-4 h-4" />
					Reports
				</button>
			</div>

			<button
				onclick={refreshCurrentTab}
				disabled={activeTab === 'players' ? playersLoading : reportsLoading}
				class="p-2.5 bg-slate-800 border border-slate-700 rounded-xl text-slate-400 hover:text-white hover:bg-slate-700 transition-all disabled:opacity-50"
			>
				<RefreshCw class="w-5 h-5 {(activeTab === 'players' ? playersLoading : reportsLoading) ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Content Area -->
	<div class="space-y-4">
		<!-- Search Bar -->
		<div class="relative group max-w-md">
			<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-blue-400 transition-colors" />
			{#if activeTab === 'players'}
				<input
					type="text"
					bind:value={playerSearchQuery}
					placeholder="Search players by name, UID, or device..."
					class="w-full pl-10 pr-4 py-3 bg-slate-800/50 border border-slate-700 rounded-xl text-sm text-slate-200 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
				/>
			{:else}
				<input
					type="text"
					bind:value={reportSearchQuery}
					placeholder="Search reports by reason, user, or server..."
					class="w-full pl-10 pr-4 py-3 bg-slate-800/50 border border-slate-700 rounded-xl text-sm text-slate-200 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
				/>
			{/if}
		</div>

		<!-- Players View -->
		{#if activeTab === 'players'}
			<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden" transition:fade={{ duration: 200 }}>
				<div class="overflow-x-auto">
					<table class="w-full text-left">
						<thead>
							<tr class="border-b border-slate-700/50 bg-slate-900/50">
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Status</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Player</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Identifiers</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Stats</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Last Seen</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-700/50">
							{#if playersLoading && players.length === 0}
								<tr>
									<td colspan="5" class="px-6 py-12 text-center text-slate-500">
										<RefreshCw class="w-8 h-8 mx-auto mb-3 animate-spin opacity-50" />
										Loading players...
									</td>
								</tr>
							{:else if filteredPlayers.length === 0}
								<tr>
									<td colspan="5" class="px-6 py-12 text-center text-slate-500">
										No players found matching your search.
									</td>
								</tr>
							{:else}
								{#each filteredPlayers as player (player.id)}
									<tr class="hover:bg-slate-700/30 transition-colors group">
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="flex items-center gap-2">
												<div class="relative">
													<div class={`w-2.5 h-2.5 rounded-full ${player.online ? 'bg-emerald-500' : 'bg-slate-600'}`}></div>
													{#if player.online}
														<div class="absolute inset-0 rounded-full bg-emerald-500 animate-ping opacity-75"></div>
													{/if}
												</div>
												<span class={`text-xs font-medium ${player.online ? 'text-emerald-400' : 'text-slate-500'}`}>
													{player.online ? 'Online' : 'Offline'}
												</span>
											</div>
										</td>
										<td class="px-6 py-4">
											<div class="flex items-center gap-3">
												<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center font-bold text-white text-sm shadow-lg shadow-blue-900/20">
													{player.name.charAt(0).toUpperCase()}
												</div>
												<div>
													<div class="font-medium text-white group-hover:text-blue-400 transition-colors">{player.name}</div>
													<div class="text-xs text-slate-500 font-mono">ID: {player.id}</div>
												</div>
											</div>
										</td>
										<td class="px-6 py-4">
											<div class="space-y-1.5">
												<div class="flex items-center gap-2 text-xs text-slate-400">
													<span class="w-12 font-mono text-slate-600 uppercase text-[10px]">UID</span>
													<code class="px-1.5 py-0.5 bg-slate-900/80 rounded text-slate-300 font-mono text-[10px]">{player.uid || 'N/A'}</code>
												</div>
												<div class="flex items-center gap-2 text-xs text-slate-400">
													<span class="w-12 font-mono text-slate-600 uppercase text-[10px]">Device</span>
													<code class="px-1.5 py-0.5 bg-slate-900/80 rounded text-slate-300 font-mono truncate max-w-[140px] text-[10px]" title={player.device_id}>{player.device_id}</code>
												</div>
											</div>
										</td>
										<td class="px-6 py-4">
											<div class="flex flex-col gap-1">
												<span class="text-xs text-slate-300">
													<span class="text-amber-400 font-bold text-sm">{player.xp.toLocaleString()}</span> XP
												</span>
											</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-400">
											<div class="flex flex-col">
												<span>{new Date(player.updated_at).toLocaleDateString()}</span>
												<span class="text-xs text-slate-600">{new Date(player.updated_at).toLocaleTimeString()}</span>
											</div>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Reports View -->
		{#if activeTab === 'reports'}
			<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden" transition:fade={{ duration: 200 }}>
				<div class="overflow-x-auto">
					<table class="w-full text-left">
						<thead>
							<tr class="border-b border-slate-700/50 bg-slate-900/50">
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Report Info</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Reporter</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Reported User</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Context</th>
								<th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Time</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-700/50">
							{#if reportsLoading && reports.length === 0}
								<tr>
									<td colspan="5" class="px-6 py-12 text-center text-slate-500">
										<RefreshCw class="w-8 h-8 mx-auto mb-3 animate-spin opacity-50" />
										Loading reports...
									</td>
								</tr>
							{:else if filteredReports.length === 0}
								<tr>
									<td colspan="5" class="px-6 py-12 text-center text-slate-500">
										No reports found.
									</td>
								</tr>
							{:else}
								{#each filteredReports as report (report.id)}
									<tr class="hover:bg-slate-700/30 transition-colors group">
										<td class="px-6 py-4">
											<div class="flex items-start gap-3">
												<div class="mt-1 p-2 bg-red-500/10 rounded-lg text-red-400">
													<AlertOctagon class="w-4 h-4" />
												</div>
												<div>
													<div class="font-medium text-white mb-1">Report #{report.id}</div>
													<div class="text-sm text-slate-300 bg-slate-900/50 p-2 rounded-lg border border-slate-700/50 max-w-xs break-words">
														"{report.reason}"
													</div>
												</div>
											</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="flex items-center gap-2">
												<div class="w-6 h-6 rounded bg-slate-700 flex items-center justify-center text-xs font-bold text-slate-300">
													{(report.reporter_name || 'U').charAt(0).toUpperCase()}
												</div>
												<div class="text-sm text-slate-300">{report.reporter_name || 'Unknown'}</div>
											</div>
											<div class="text-xs text-slate-500 mt-0.5 ml-8">ID: {report.reporter_id}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="flex items-center gap-2">
												<div class="w-6 h-6 rounded bg-red-900/30 flex items-center justify-center text-xs font-bold text-red-400 border border-red-500/20">
													{(report.reported_user_name || 'U').charAt(0).toUpperCase()}
												</div>
												<div class="text-sm text-red-300 font-medium">{report.reported_user_name || 'Unknown'}</div>
											</div>
											<div class="text-xs text-slate-500 mt-0.5 ml-8">ID: {report.reported_user_id}</div>
										</td>
										<td class="px-6 py-4">
											{#if report.game_server_instance_id}
												<div class="flex items-center gap-2 text-xs text-slate-400">
													<Server class="w-3 h-3 text-slate-500" />
													<code class="bg-slate-900 px-1.5 py-0.5 rounded text-slate-300 font-mono">{report.game_server_instance_id}</code>
												</div>
											{:else}
												<span class="text-xs text-slate-600 italic">No server context</span>
											{/if}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-400">
											<div>{new Date(report.timestamp).toLocaleDateString()}</div>
											<div class="text-xs text-slate-600">{new Date(report.timestamp).toLocaleTimeString()}</div>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	</div>
</div>
