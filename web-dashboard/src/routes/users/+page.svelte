<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { onMount, tick } from 'svelte';
	import { fade, fly, scale, slide } from 'svelte/transition';
	import {
		Users,
		Search,
		RefreshCw,
		CheckCircle,
		Server,
		Pencil,
		Trash2,
		Terminal,
		Activity,
		ShieldAlert,
		Clock,
		ChevronRight,
		Cpu,
		Database,
		Lock,
		Hash,
		Dna,
		AlertOctagon,
		Signal,
		Check,
		Ban,
		Shield,
		ChevronDown,
		History,
		Eye,
		MapPin,
		HardDrive,
		Smartphone,
		Fingerprint
	} from 'lucide-svelte';
	import EditPlayerModal from '$lib/components/players/EditPlayerModal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import Button from '$lib/components/Button.svelte';
	import { notifications, siteSettings } from '$lib/stores.svelte';
	import type { Player } from '$lib/stores.svelte';

	interface Report {
		id: number;
		reporter_id: number;
		reported_user_id: number;
		reason: string;
		game_server_instance_id: string;
		created_at: string;
		reporter_name?: string;
		reported_user_name?: string;
	}

	let activeTab = $state<'players' | 'reports'>('players');

	// Players State
	let players = $state<Player[]>([]);
	let expandedPlayerId = $state<number | null>(null);

	// Summary Derived
	let totalXP = $derived(players.reduce((sum, p) => sum + p.xp, 0));
	let onlineCount = $derived(players.filter((p) => p.online).length);
	let bannedCount = $derived(players.filter((p) => p.banned).length);

	let playersLoading = $state(true);
	let playerSearchQuery = $state('');
	let playerSortBy = $state<'id' | 'name' | 'xp' | 'updated_at'>('id');
	let selectedPlayer = $state<Player | null>(null);
	let isEditModalOpen = $state(false);
	let isDeleteConfirmOpen = $state(false);
	let playerToDelete = $state<Player | null>(null);

	// Reports State
	let reports = $state<Report[]>([]);
	let reportsLoading = $state(true);
	let reportSearchQuery = $state('');

	let filteredPlayers = $derived.by(() => {
		let result = players;
		if (playerSearchQuery.trim()) {
			const query = playerSearchQuery.toLowerCase();
			result = players.filter(
				(p) =>
					p.name.toLowerCase().includes(query) ||
					p.uid.toLowerCase().includes(query) ||
					p.device_id.toLowerCase().includes(query)
			);
		}

		return [...result].sort((a, b) => {
			if (playerSortBy === 'xp') return b.xp - a.xp;
			if (playerSortBy === 'updated_at')
				return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
			if (playerSortBy === 'name') return a.name.localeCompare(b.name);
			return b.id - a.id;
		});
	});

	let filteredReports = $derived.by(() => {
		if (!reportSearchQuery.trim()) return reports;
		const query = reportSearchQuery.toLowerCase();
		return reports.filter(
			(r) =>
				r.reason.toLowerCase().includes(query) ||
				(r.reporter_name?.toLowerCase() || '').includes(query) ||
				(r.reported_user_name?.toLowerCase() || '').includes(query) ||
				(r.game_server_instance_id?.toLowerCase() || '').includes(query)
		);
	});

	async function fetchPlayers() {
		playersLoading = true;
		try {
			const res = await apiFetch('/api/admin/players');
			if (res.ok) {
				players = (await res.json()) || [];
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
			const res = await apiFetch('/api/reports');
			if (res.ok) {
				reports = (await res.json()) || [];
			}
		} catch (e) {
			console.error('Failed to fetch reports', e);
		} finally {
			reportsLoading = false;
		}
	}

	function toggleExpand(id: number) {
		expandedPlayerId = expandedPlayerId === id ? null : id;
	}

	function refreshCurrentTab() {
		if (activeTab === 'players') {
			fetchPlayers();
		} else {
			fetchReports();
		}
	}

	function openEditModal(player: Player) {
		selectedPlayer = player;
		isEditModalOpen = true;
	}

	function handlePlayerUpdated(updatedPlayer: Player) {
		players = players.map((p) => (p.id === updatedPlayer.id ? { ...p, ...updatedPlayer } : p));
	}

	function confirmDelete(player: Player) {
		playerToDelete = player;
		isDeleteConfirmOpen = true;
	}

	async function handleDeletePlayer() {
		if (!playerToDelete) return;
		try {
			const res = await apiFetch(`/api/admin/players/${playerToDelete.id}`, { method: 'DELETE' });
			if (res.ok) {
				players = players.filter((p) => p.id !== playerToDelete!.id);
				notifications.add({ type: 'success', message: 'RECORD_DELETED' });
			} else {
				throw new Error('Failed to delete player');
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'TERMINAL_ERROR' });
		} finally {
			isDeleteConfirmOpen = false;
			playerToDelete = null;
		}
	}

	async function toggleBan(player: Player) {
		const newStatus = !player.banned;
		try {
			const res = await apiFetch(`/api/admin/players/${player.id}/ban`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ banned: newStatus })
			});
			if (res.ok) {
				const updated = await res.json();
				players = players.map((p) => (p.id === updated.id ? { ...p, ...updated } : p));
				notifications.add({
					type: newStatus ? 'error' : 'success',
					message: newStatus ? 'User banned' : 'User restored'
				});
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Connection error' });
		}
	}

	onMount(() => {
		fetchPlayers();
		fetchReports();
		const interval = setInterval(() => {
			if (activeTab === 'players') fetchPlayers();
		}, 10000);
		return () => clearInterval(interval);
	});
</script>

<div
	class="w-full min-h-[calc(100vh-140px)] md:min-h-[calc(100vh-160px)] flex flex-col overflow-hidden relative font-sans"
>
	<!-- Main Content Layout -->
	<div class="w-full h-full flex flex-col gap-6 relative z-10 pb-32 md:pb-12">
		<!-- Header -->
		<div
			class="flex flex-col xl:flex-row xl:items-end justify-between gap-6 p-6 bg-slate-800/40 border border-white/5 rounded-2xl shadow-lg backdrop-blur-md"
		>
			<div class="space-y-2 relative z-10">
				<div class="flex items-center gap-3 text-slate-400">
					<div class="flex items-center gap-2">
						<Users class="w-4 h-4 text-indigo-400" />
						<span class="text-xs font-bold uppercase tracking-wider">User Management</span>
					</div>
				</div>
				<h1 class="text-3xl font-bold text-white tracking-tight">
					User <span class="text-indigo-400">Registry</span>
				</h1>
			</div>

			<div class="flex flex-wrap items-center gap-4 relative z-10">
				<!-- Tab Switcher -->
				<div class="flex gap-1.5 bg-black/20 p-1.5 rounded-2xl border border-white/5 backdrop-blur-md">
					<button
						onclick={() => (activeTab = 'players')}
						class="px-6 py-2.5 rounded-xl text-[13px] font-bold uppercase tracking-tight transition-all duration-300 {activeTab === 'players'
							? 'bg-sky-500/10 text-sky-400 shadow-lg shadow-black/20 border border-sky-500/20'
							: 'text-slate-500 hover:text-slate-200 hover:bg-white/5 border border-transparent'}"
					>
						Users
					</button>
					<button
						onclick={() => (activeTab = 'reports')}
						class="px-6 py-2.5 rounded-xl text-[13px] font-bold uppercase tracking-tight transition-all duration-300 {activeTab === 'reports'
							? 'bg-rose-500/10 text-rose-400 shadow-lg shadow-black/20 border border-rose-500/20'
							: 'text-slate-500 hover:text-slate-200 hover:bg-white/5 border border-transparent'}"
					>
						Reports
					</button>
				</div>

				<Button
					onclick={refreshCurrentTab}
					disabled={activeTab === 'players' ? playersLoading : reportsLoading}
					loading={activeTab === 'players' ? playersLoading : reportsLoading}
					variant="secondary"
					size="md"
					icon="ph:arrows-clockwise-bold"
					class="!rounded-2xl !p-4 shadow-xl"
				/>
			</div>
		</div>

		<!-- Summary Stats -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<StatsCard
				title="Total Users"
				value={players.length}
				iconName="ph:users-bold"
				color="rust"
				subValue="Registered Accounts"
			/>
			<StatsCard
				title="Active Sessions"
				value={onlineCount}
				iconName="activity"
				color="emerald"
				subValue={`${((onlineCount / (players.length || 1)) * 100).toFixed(1)}% Online`}
			/>
			<StatsCard
				title="Total XP"
				value={totalXP.toLocaleString()}
				iconName="ph:dna-bold"
				color="orange"
				subValue="Player Experience"
			/>
			<StatsCard
				title="Active Bans"
				value={reports.length}
				iconName="ph:shield-warning-bold"
				color="red"
				subValue={`${bannedCount} Suspended`}
			/>
		</div>

		<!-- Search & Filter -->
		<div class="grid grid-cols-1 xl:grid-cols-12 gap-6 items-center">
			<div class="xl:col-span-8 relative group">
				<div class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none">
					<Search class="w-4 h-4" />
				</div>
				{#if activeTab === 'players'}
					<input
						type="text"
						bind:value={playerSearchQuery}
						placeholder="Search users by name, ID, or device..."
						class="w-full bg-slate-900/50 border border-white/5 rounded-xl pl-11 pr-4 py-3 text-sm text-slate-200 font-medium focus:border-indigo-500/50 focus:ring-1 focus:ring-indigo-500/20 outline-none transition-all placeholder:text-slate-500"
					/>
				{:else}
					<input
						type="text"
						bind:value={reportSearchQuery}
						placeholder="Search reports..."
						class="w-full bg-slate-900/50 border border-white/5 rounded-xl pl-11 pr-4 py-3 text-sm text-slate-200 font-medium focus:border-rose-500/50 focus:ring-1 focus:ring-rose-500/20 outline-none transition-all placeholder:text-slate-500"
					/>
				{/if}
			</div>

			<div
				class="xl:col-span-4 flex items-center gap-3 bg-slate-900/50 p-2 rounded-xl border border-white/5 overflow-x-auto"
			>
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider pl-2 shrink-0"
					>Sort By:</span
				>
				<div class="flex gap-1 flex-1">
					{#each [{ id: 'id', label: 'ID' }, { id: 'name', label: 'Name' }, { id: 'xp', label: 'XP' }, { id: 'updated_at', label: 'Recent' }] as sort}
						<Button
							onclick={() => (playerSortBy = sort.id as any)}
							variant={playerSortBy === sort.id ? 'primary' : 'ghost'}
							size="xs"
							class="flex-1 !rounded-lg !text-xs"
						>
							{sort.label}
						</Button>
					{/each}
				</div>
			</div>
		</div>

		<!-- Content Grid -->
		<div class="flex-1 overflow-y-auto custom-scrollbar">
			{#if activeTab === 'players'}
				{#if playersLoading && players.length === 0}
					<div class="py-20 flex flex-col items-center gap-4">
						<div
							class="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"
						></div>
						<span class="text-xs font-medium text-slate-500 uppercase tracking-wide"
							>Loading users...</span
						>
					</div>
				{:else if filteredPlayers.length === 0}
					<div class="py-20 flex flex-col items-center gap-4 opacity-60">
						<div class="p-6 bg-slate-900/50 rounded-2xl border border-white/5">
							<Users class="w-10 h-10 text-slate-600" />
						</div>
						<span class="text-sm font-medium text-slate-500">No users found matching criteria</span>
					</div>
				{:else}
					<div class="grid grid-cols-1 xl:grid-cols-2 gap-6 pb-20">
						{#each filteredPlayers as player (player.id)}
							{@const isExpanded = expandedPlayerId === player.id}
							<div
								class="modern-card bg-slate-900/50 border border-white/5 rounded-2xl hover:border-sky-500/20 transition-all duration-300 relative group overflow-visible"
								in:fade={{ duration: 200 }}
							>
								<!-- Status Line -->
								<div
									class={`absolute left-0 top-0 bottom-0 w-1.5 rounded-l-2xl ${player.banned ? 'bg-rose-500' : player.online ? 'bg-emerald-500' : 'bg-slate-700'}`}
								></div>

								<div class="p-6 pl-8 flex flex-col gap-6">
									<!-- Header -->
									<div class="flex items-start justify-between gap-4">
										<div class="flex items-center gap-5">
											<!-- Avatar -->
											<div
												class="w-14 h-14 bg-slate-800 rounded-xl border border-white/5 flex items-center justify-center text-xl font-bold text-slate-200 cursor-pointer hover:bg-slate-700 transition-colors shadow-sm"
												onclick={() => toggleExpand(player.id)}
												role="button"
												tabindex="0"
												onkeydown={(e) => e.key === 'Enter' && toggleExpand(player.id)}
											>
												{#if player.banned}
													<Ban class="w-6 h-6 text-rose-500/50" />
												{:else if player.online}
													<span class="relative z-10">{player.name.charAt(0).toUpperCase()}</span>
													<div
														class="absolute -bottom-1 -right-1 w-4 h-4 bg-emerald-500 rounded-full border-2 border-slate-900 shadow-sm"
													></div>
												{:else}
													{player.name.charAt(0).toUpperCase()}
												{/if}
											</div>

											<div>
												<div class="flex items-center gap-3 flex-wrap">
													<button
														class="text-xl font-bold text-white hover:text-sky-400 transition-colors text-left tracking-tight"
														onclick={() => toggleExpand(player.id)}
													>
														{player.name}
													</button>
													{#if player.banned}
														<span
															class="px-2 py-0.5 rounded-lg bg-rose-500/10 text-rose-400 text-[10px] font-bold uppercase tracking-wide border border-rose-500/20"
															>Banned</span
														>
													{/if}
													{#if player.online}
														<span
															class="px-2 py-0.5 rounded-lg bg-emerald-500/10 text-emerald-400 text-[10px] font-bold uppercase tracking-wide border border-emerald-500/20"
															>Active</span
														>
													{/if}
												</div>
												<div
													class="flex items-center gap-4 mt-1.5 text-xs text-slate-500 font-semibold"
												>
													<span>ID: {player.id}</span>
													<span class="w-1 h-1 rounded-full bg-slate-700"></span>
													<span>{player.online ? 'Online Now' : 'Offline'}</span>
												</div>
											</div>
										</div>

										<div class="flex gap-2">
											<Button
												onclick={() => openEditModal(player)}
												variant="ghost"
												size="sm"
												icon="ph:pencil-bold"
												class="!p-2.5 !rounded-xl text-slate-400 hover:text-sky-400 hover:bg-white/5"
											/>
											<Button
												onclick={() => toggleExpand(player.id)}
												variant="ghost"
												size="sm"
												icon="ph:caret-down-bold"
												class="{isExpanded
													? 'rotate-180'
													: ''} transition-transform duration-300 !p-2.5 !rounded-xl text-slate-400 hover:bg-white/5"
											/>
										</div>
									</div>

									<!-- Metrics -->
									<div
										class="grid grid-cols-2 gap-6 bg-slate-950/40 rounded-xl p-4 border border-white/5 shadow-inner"
									>
										<div class="flex flex-col gap-1">
											<span class="text-[10px] uppercase font-bold text-slate-500 tracking-widest"
												>Experience Points</span
											>
											<span class="text-lg font-bold text-amber-400 tabular-nums"
												>{player.xp.toLocaleString()}</span
											>
										</div>
										<div class="flex flex-col text-right gap-1">
											<span class="text-[10px] uppercase font-bold text-slate-500 tracking-widest"
												>Account Level</span
											>
											<span class="text-lg font-bold text-sky-400"
												>LV {Math.floor(player.xp / 1000) + 1}</span
											>
										</div>
									</div>

									<!-- Expandable Details -->
									{#if isExpanded}
										<div class="space-y-5 pt-2 border-t border-white/5" transition:slide>
											<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
												<div class="space-y-2">
													<span
														class="text-[10px] uppercase font-bold text-slate-500 tracking-wider"
														>Account UUID</span
													>
													<div
														class="bg-slate-950/60 p-3 rounded-xl border border-white/5 text-xs font-mono text-slate-400 break-all shadow-inner"
													>
														{player.uid || 'N/A'}
													</div>
												</div>
												<div class="space-y-2">
													<span
														class="text-[10px] uppercase font-bold text-slate-500 tracking-wider"
														>Hardware ID</span
													>
													<div
														class="bg-slate-950/60 p-3 rounded-xl border border-white/5 text-xs font-mono text-slate-400 break-all shadow-inner"
													>
														{player.device_id || 'Virtual'}
													</div>
												</div>
											</div>

											<div class="grid grid-cols-3 gap-3">
												<div
													class="bg-slate-800/30 p-3.5 rounded-xl border border-white/5 flex flex-col gap-1 shadow-sm"
												>
													<div
														class="text-[10px] text-slate-500 font-bold uppercase tracking-wider"
													>
														Last Sync
													</div>
													<div class="text-xs text-slate-200 font-semibold truncate">
														{player.last_joined_server?.split('-').pop() || 'Unknown'}
													</div>
												</div>
												<div
													class="bg-slate-800/30 p-3.5 rounded-xl border border-white/5 flex flex-col gap-1 shadow-sm"
												>
													<div
														class="text-[10px] text-slate-500 font-bold uppercase tracking-wider"
													>
														Registered
													</div>
													<div class="text-xs text-slate-200 font-semibold">
														{new Date(player.created_at).toLocaleDateString()}
													</div>
												</div>
												<div
													class="bg-slate-800/30 p-3.5 rounded-xl border border-white/5 flex flex-col gap-1 shadow-sm"
												>
													<div
														class="text-[10px] text-slate-500 font-bold uppercase tracking-wider"
													>
														Status
													</div>
													<div
														class={`text-xs font-bold uppercase tracking-wide ${player.banned ? 'text-rose-400' : 'text-emerald-400'}`}
													>
														{player.banned ? 'Suspended' : 'Active'}
													</div>
												</div>
											</div>
										</div>
									{/if}

									<!-- Footer Actions -->
									<div class="flex items-center justify-between pt-2">
										<span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider"
											>Last Active: {new Date(player.updated_at).toLocaleTimeString([], {
												hour12: false
											})}</span
										>

										<div class="flex gap-2">
											<Button
												onclick={() => toggleBan(player)}
												variant={player.banned ? 'success' : 'danger'}
												size="xs"
												class="!rounded-lg !px-4"
											>
												{player.banned ? 'Restore Account' : 'Suspend Account'}
											</Button>
											<Button
												onclick={() => confirmDelete(player)}
												variant="ghost"
												size="xs"
												class="!text-rose-400 hover:!bg-rose-500/10 !rounded-lg !px-4"
											>
												Delete
											</Button>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{:else}
				<!-- Reports List -->
				{#if reportsLoading && reports.length === 0}
					<div class="py-20 flex flex-col items-center gap-4">
						<div
							class="w-8 h-8 border-2 border-rose-500 border-t-transparent rounded-full animate-spin"
						></div>
						<span class="text-xs font-medium text-slate-500 uppercase tracking-wide"
							>Syncing Reports...</span
						>
					</div>
				{:else if filteredReports.length === 0}
					<div class="py-20 flex flex-col items-center gap-4 opacity-60">
						<div class="p-8 bg-slate-900/50 rounded-[2.5rem] border border-white/5 shadow-xl">
							<ShieldAlert class="w-12 h-12 text-slate-600" />
						</div>
						<span class="text-base font-semibold text-slate-500">No active incidents</span>
					</div>
				{:else}
					<div class="grid grid-cols-1 gap-6 pb-20">
						{#each filteredReports as report (report.id)}
							<div
								class="modern-card bg-slate-900/50 border border-white/5 rounded-2xl overflow-hidden hover:border-rose-500/30 transition-all duration-300 shadow-xl group"
								in:fade={{ duration: 200 }}
							>
								<div class="flex flex-col md:flex-row">
									<!-- Sidebar Indicator -->
									<div
										class="w-full md:w-1.5 bg-rose-500/50 shadow-[0_0_15px_rgba(244,63,94,0.3)]"
									></div>

									<div class="flex-1 p-8 grid grid-cols-1 md:grid-cols-12 gap-10 items-start">
										<!-- Main Info -->
										<div class="md:col-span-5 space-y-4">
											<div class="flex items-center gap-4">
												<div
													class="p-3 bg-rose-500/10 rounded-xl text-rose-500 border border-rose-500/20"
												>
													<ShieldAlert size={24} />
												</div>
												<div>
													<h3 class="text-xl font-bold text-white tracking-tight">
														Report #{report.id}
													</h3>
													<span class="text-xs font-bold text-rose-400 uppercase tracking-widest"
														>High Severity</span
													>
												</div>
											</div>
											<div
												class="p-4 bg-rose-500/5 border border-rose-500/10 rounded-xl text-sm font-medium text-rose-200 leading-relaxed shadow-inner"
											>
												"{report.reason}"
											</div>
										</div>

										<!-- Actors -->
										<div class="md:col-span-4 grid grid-cols-2 gap-6 border-l border-white/5 pl-10">
											<div class="space-y-2">
												<span class="text-[10px] font-bold uppercase text-slate-500 tracking-widest"
													>Reporter</span
												>
												<div class="flex items-center gap-3">
													<div
														class="w-8 h-8 rounded-lg bg-slate-800 border border-white/5 flex items-center justify-center text-sm font-bold text-slate-300"
													>
														{(report.reporter_name || 'U').charAt(0)}
													</div>
													<div class="text-sm font-bold text-slate-100 truncate">
														{report.reporter_name || 'Anon'}
													</div>
												</div>
											</div>
											<div class="space-y-2">
												<span class="text-[10px] font-bold uppercase text-slate-500 tracking-widest"
													>Reported</span
												>
												<div class="flex items-center gap-3">
													<div
														class="w-8 h-8 rounded-lg bg-rose-900/30 border border-rose-500/20 flex items-center justify-center text-sm font-bold text-rose-400"
													>
														{(report.reported_user_name || 'U').charAt(0)}
													</div>
													<div class="text-sm font-bold text-rose-200 truncate">
														{report.reported_user_name || 'Unknown'}
													</div>
												</div>
											</div>
										</div>

										<!-- Metadata -->
										<div class="md:col-span-3 space-y-6 border-l border-white/5 pl-10">
											<div>
												<span
													class="text-[10px] font-bold uppercase text-slate-500 tracking-widest block mb-2"
													>Source Server</span
												>
												{#if report.game_server_instance_id}
													<div
														class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-950 border border-white/5 text-xs font-mono text-sky-400 shadow-inner"
													>
														<Server size={14} />
														<span>{report.game_server_instance_id.slice(0, 12)}</span>
													</div>
												{:else}
													<span class="text-xs text-slate-600 font-medium italic">Unknown Node</span
													>
												{/if}
											</div>
											<div>
												<div class="text-sm font-bold text-slate-200">
													{new Date(report.created_at).toLocaleDateString()}
												</div>
												<div class="text-xs font-medium text-slate-500 mt-0.5">
													{new Date(report.created_at).toLocaleTimeString()}
												</div>
											</div>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Footer Status -->
		<div
			class="bg-slate-900/80 border-t border-white/5 p-4 flex justify-between items-center text-xs font-medium text-slate-500 rounded-2xl mx-2"
		>
			<div class="flex gap-6">
				<div class="flex items-center gap-2">
					<Activity class="w-3.5 h-3.5 text-emerald-500" />
					<span>System Stable</span>
				</div>
				<div class="flex items-center gap-2">
					<Shield class="w-3.5 h-3.5 text-indigo-500" />
					<span>Database Online</span>
				</div>
			</div>
			<div>Exile Dashboard v0.9.4</div>
		</div>
	</div>

	<EditPlayerModal
		isOpen={isEditModalOpen}
		player={selectedPlayer}
		onClose={() => (isEditModalOpen = false)}
		onSave={handlePlayerUpdated}
	/>

	<ConfirmDialog
		bind:isOpen={isDeleteConfirmOpen}
		title="Delete User Record"
		isCritical={true}
		message={`Are you sure you want to permanently delete user "${playerToDelete?.name}"? This action cannot be undone.`}
		onConfirm={handleDeletePlayer}
		on:close={() => (isDeleteConfirmOpen = false)}
	/>
</div>

<style>
	/* Cinematic Intelligence Interface Styles */
	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0, 0, 0, 0.8) 100%);
	}

	/* Elegant Scrollbar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
		height: 4px;
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

	input:focus {
		outline: none;
	}
</style>
