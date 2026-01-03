<script lang="ts">
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

	interface Player {
		id: number;
		uid: string;
		name: string;
		device_id: string;
		xp: number;
		banned: boolean;
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
	let expandedPlayerId = $state<number | null>(null);

	// Summary Derived
	let totalXP = $derived(players.reduce((sum, p) => sum + p.xp, 0));
	let onlineCount = $derived(players.filter(p => p.online).length);
	let bannedCount = $derived(players.filter(p => p.banned).length);

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
			if (playerSortBy === 'updated_at') return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
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
			const res = await fetch('/api/admin/players');
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
			const res = await fetch('/api/reports');
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
			const res = await fetch(`/api/admin/players/${playerToDelete.id}`, { method: 'DELETE' });
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
			const res = await fetch(`/api/admin/players/${player.id}/ban`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ banned: newStatus })
			});
			if (res.ok) {
				const updated = await res.json();
				players = players.map(p => p.id === updated.id ? { ...p, ...updated } : p);
				notifications.add({ 
					type: newStatus ? 'error' : 'success', 
					message: newStatus ? 'SUBJECT_BANNED' : 'SUBJECT_RESTORED' 
				});
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'UPLINK_FAILURE' });
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

<div class="w-full min-h-[calc(100vh-140px)] md:min-h-[calc(100vh-160px)] flex flex-col overflow-hidden relative font-sans">
	<!-- Cinematic Overlays -->
	<div class="fixed inset-0 pointer-events-none z-[100] bg-vignette opacity-40"></div>
	
	<!-- Main Content Chassis -->
	<div class="w-full h-full flex flex-col gap-8 relative z-10 pb-32 md:pb-12">
		
		<!-- Intelligence Header -->
		<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-4 border-rust pl-4 sm:pl-10 py-4 bg-[var(--header-bg)] shadow-2xl relative overflow-hidden industrial-frame">
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
			
			<div class="space-y-4 p-2 relative z-10">
				<div class="flex items-center gap-4">
					<div class="flex items-center gap-2">
						<Fingerprint class="w-4 h-4 text-rust animate-pulse" />
						<span class="text-rust text-[9px] font-black uppercase tracking-[0.3em]">Neural_Registry_v2.0</span>
					</div>
					<div class="w-px h-3 bg-stone-800"></div>
					<span class="font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-[0.4em] italic hidden sm:inline">STATION: EXILE_HIVE_CORE</span>
				</div>
				<h1 class="text-3xl sm:text-5xl lg:text-7xl font-heading font-black tracking-tighter text-white uppercase leading-none">
					Subject_<span class="text-rust">Dossiers</span>
				</h1>
			</div>

			<div class="flex flex-wrap items-center gap-4 sm:gap-8 p-4 relative z-10">
				<!-- Tactical Tab Switcher -->
				<div class="flex gap-2 bg-black/40 p-1 border border-stone-800 shadow-inner flex-1 sm:flex-initial">
					<Button
						onclick={() => (activeTab = 'players')}
						variant={activeTab === 'players' ? 'primary' : 'ghost'}
						size="md"
						class="flex-1 sm:flex-initial !flex-col !items-start px-4 sm:px-8 py-2 sm:py-4"
					>
						<span class="font-jetbrains text-[8px] font-black tracking-[0.3em] uppercase mb-1 opacity-50">Operational Base</span>
						<span class="font-heading text-xs sm:text-base font-black tracking-widest uppercase">Subjects</span>
						{#if activeTab === 'players'}
							<div class="absolute -top-1 -right-1 w-2 h-2 bg-white shadow-[0_0_10px_rgba(255,255,255,0.5)]"></div>
						{/if}
					</Button>
					<Button
						onclick={() => (activeTab = 'reports')}
						variant={activeTab === 'reports' ? 'danger' : 'ghost'}
						size="md"
						class="flex-1 sm:flex-initial !flex-col !items-start px-4 sm:px-8 py-2 sm:py-4"
					>
						<span class="font-jetbrains text-[8px] font-black tracking-[0.3em] uppercase mb-1 opacity-50">Anomaly Logs</span>
						<span class="font-heading text-xs sm:text-base font-black tracking-widest uppercase">Incidents</span>
						{#if activeTab === 'reports'}
							<div class="absolute -top-1 -right-1 w-2 h-2 bg-white shadow-[0_0_10px_rgba(255,255,255,0.5)]"></div>
						{/if}
					</Button>
				</div>

				<Button
					onclick={refreshCurrentTab}
					disabled={activeTab === 'players' ? playersLoading : reportsLoading}
					loading={activeTab === 'players' ? playersLoading : reportsLoading}
					variant="secondary"
					size="lg"
					icon="ph:arrows-clockwise-bold"
					class="!p-3 sm:!p-5"
				/>
			</div>
		</div>

		<!-- Strategic Summary -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 px-2">
			<StatsCard 
				title="Total Subjects" 
				value={players.length} 
				iconName="ph:users-bold" 
				color="rust"
				subValue="Neural Registry Size"
			/>
			<StatsCard 
				title="Uplink Active" 
				value={onlineCount} 
				iconName="activity" 
				color="emerald"
				subValue={`${((onlineCount / (players.length || 1)) * 100).toFixed(1)}% Saturation`}
			/>
			<StatsCard 
				title="Cumulative XP" 
				value={totalXP.toLocaleString()} 
				iconName="ph:dna-bold" 
				color="orange"
				subValue="Total Biomass Growth"
			/>
			<StatsCard 
				title="Active Reports" 
				value={reports.length} 
				iconName="ph:shield-warning-bold" 
				color="red"
				subValue={`${bannedCount} Banned Entities`}
			/>
		</div>

		<!-- Neural Scanner HUD -->
		<div class="grid grid-cols-1 xl:grid-cols-12 gap-6 items-center px-2">
			<div class="xl:col-span-8 relative group">
				<div class="absolute left-6 top-1/2 -translate-y-1/2 flex items-center gap-3 pointer-events-none opacity-50 group-focus-within:opacity-100 transition-opacity">
					<Search class="w-5 h-5 text-rust" />
					<div class="w-px h-4 bg-stone-800"></div>
				</div>
				{#if activeTab === 'players'}
					<input
						type="text"
						bind:value={playerSearchQuery}
						placeholder="SCAN_FOR_NEURAL_SIGNATURE (NAME / ID / DEVICE)..."
						class="w-full bg-stone-950/80 backdrop-blur-md border border-stone-800 pl-16 pr-10 py-5 font-jetbrains text-sm text-stone-200 font-bold tracking-[0.2em] focus:border-rust outline-none transition-all uppercase placeholder:text-stone-900 italic shadow-2xl industrial-frame"
					/>
				{:else}
					<input
						type="text"
						bind:value={reportSearchQuery}
						placeholder="SCAN_FOR_FAULT_VECTORS (REASON / SERVER)..."
						class="w-full bg-stone-950/80 backdrop-blur-md border border-stone-800 pl-16 pr-10 py-5 font-jetbrains text-sm text-stone-200 font-bold tracking-[0.2em] focus:border-rust outline-none transition-all uppercase placeholder:text-stone-900 italic shadow-2xl industrial-frame"
					/>
				{/if}
				<!-- Scanning Line Effect -->
				<div class="absolute inset-0 border border-rust/0 group-focus-within:border-rust/20 pointer-events-none transition-colors"></div>
			</div>
			
			<div class="xl:col-span-4 flex items-center gap-4 bg-black/40 p-1.5 border border-stone-800 shadow-inner industrial-frame">
				<span class="font-jetbrains text-[8px] font-black text-text-dim uppercase tracking-widest pl-4 shrink-0 italic">Sorting_Kernel:</span>
				<div class="flex gap-1 flex-1">
					{#each [
						{ id: 'id', label: 'Hex_ID' },
						{ id: 'name', label: 'Alpha' },
						{ id: 'xp', label: 'Biomass' },
						{ id: 'updated_at', label: 'Last_Seen' }
					] as sort}
						<Button 
							onclick={() => playerSortBy = sort.id as any}
							variant={playerSortBy === sort.id ? 'primary' : 'ghost'}
							size="xs"
							class="flex-1"
						>
							{sort.label}
						</Button>
					{/each}
				</div>
			</div>
		</div>

		<!-- Strategic Intelligence Grid -->
		<div class="flex-1 overflow-y-auto custom-scrollbar px-2">
			{#if activeTab === 'players'}
				{#if playersLoading && players.length === 0}
					<div class="py-40 flex flex-col items-center gap-6">
						<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
						<span class="font-heading font-black text-[12px] font-black tracking-[0.6em] uppercase text-rust animate-pulse">Neural_Sync_Active...</span>
					</div>
				{:else if filteredPlayers.length === 0}
					<div class="py-40 flex flex-col items-center gap-6 opacity-40">
						<div class="p-10 bg-stone-900/40 border border-stone-800 industrial-frame">
							<AlertOctagon class="w-16 h-16 text-stone-800" />
						</div>
						<span class="font-heading font-black text-xl tracking-[0.4em] uppercase italic text-stone-700">>> SUBJECT_NOT_FOUND_IN_SECTOR</span>
					</div>
				{:else}
					<div class="grid grid-cols-1 xl:grid-cols-2 gap-6 pb-20">
						{#each filteredPlayers as player (player.id)}
							{@const isExpanded = expandedPlayerId === player.id}
							<div 
								class="modern-industrial-card glass-panel group relative flex flex-col overflow-hidden shadow-2xl transition-all duration-500 industrial-sharp {isExpanded ? 'border-rust shadow-rust/20 ring-1 ring-rust/30' : 'hover:border-rust/40'}"
								in:fade={{ duration: 200 }}
							>
								<!-- Top Status Indicator -->
								<div class={`h-1.5 w-full relative transition-colors duration-500 z-10 ${player.banned ? 'bg-danger' : (player.online ? 'bg-success' : 'bg-stone-800')}`}>
									{#if player.online && !player.banned}
										<div class="absolute inset-0 bg-white/30 animate-glow-slide"></div>
									{/if}
								</div>

								<div class="p-6 flex flex-col gap-6 relative z-10">
									<!-- Subject Primary Identity -->
									<div class="flex items-start justify-between gap-4">
										<div class="flex items-center gap-5">
											<!-- Bio-Metric Icon -->
											<div 
												class="w-16 h-16 bg-stone-950 border border-stone-800 flex items-center justify-center text-white font-heading font-black italic text-2xl group-hover:border-rust transition-all duration-500 industrial-frame shadow-inner shrink-0 relative overflow-hidden"
												onclick={() => toggleExpand(player.id)}
												role="button"
												tabindex="0"
												onkeydown={(e) => e.key === 'Enter' && toggleExpand(player.id)}
											>
												<div class="absolute inset-0 opacity-10 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')]"></div>
												{#if player.banned}
													<Ban class="w-8 h-8 text-danger/40" />
												{:else if player.online}
													<div class="absolute inset-0 flex items-center justify-center opacity-20 animate-pulse">
														<Dna class="w-10 h-10 text-success" />
													</div>
													<span class="relative z-10">{player.name.charAt(0).toUpperCase()}</span>
												{:else}
													{player.name.charAt(0).toUpperCase()}
												{/if}
											</div>

											<div class="min-w-0 flex-1">
												<div class="flex items-center gap-3 flex-wrap">
													<h3>
														<button 
															class="text-2xl sm:text-3xl font-heading font-black tracking-tighter text-white uppercase leading-none group-hover:text-rust transition-all duration-500 truncate cursor-pointer bg-transparent border-none p-0 text-left"
															onclick={() => toggleExpand(player.id)}
														>
															{player.name}
														</button>
													</h3>
													{#if player.banned}
														<div class="bg-danger text-white text-[7px] font-black px-2 py-0.5 tracking-[0.2em] flex items-center gap-1 shadow-lg shadow-red-900/40 border border-red-400/30">
															TERMINATED
														</div>
													{/if}
													{#if player.online}
														<div class="bg-success/10 text-success text-[7px] font-black px-2 py-0.5 tracking-[0.2em] border border-emerald-500/30">
															UPLINK_LIVE
														</div>
													{/if}
												</div>
												<div class="flex items-center gap-4 mt-3">
													<span class="font-mono text-[9px] font-black text-text-dim uppercase tracking-[0.2em] italic">Sig: 0x{player.id.toString(16).toUpperCase()}</span>
													<div class="w-px h-2 bg-stone-800"></div>
													<div class="flex items-center gap-2">
														<div class={`w-1.5 h-1.5 rounded-full ${player.online ? 'bg-success shadow-[0_0_8px_var(--color-success)] animate-pulse' : 'bg-stone-800'}`}></div>
														<span class="text-[8px] font-black text-stone-600 uppercase tracking-widest">{player.online ? 'Active' : 'Standby'}</span>
													</div>
												</div>
											</div>
										</div>

										<div class="flex flex-col gap-1">
											<Button 
												onclick={() => openEditModal(player)}
												variant="secondary"
												size="xs"
												icon="ph:pencil-bold"
												title="Edit Dossier"
											/>
											<Button 
												onclick={() => toggleExpand(player.id)}
												variant="secondary"
												size="xs"
												icon="ph:chevron-down-bold"
												class="{isExpanded ? 'rotate-180' : ''} transition-transform duration-500"
											/>
										</div>
									</div>

									<!-- Bio-Metric Data Rack -->
									<div class="grid grid-cols-3 gap-4">
										<div class="col-span-2 bg-black/40 p-4 border border-stone-800 industrial-frame relative overflow-hidden group/rack">
											<div class="absolute top-0 left-0 w-full h-[1px] bg-warning/20"></div>
											<span class="text-[7px] font-black text-stone-600 uppercase tracking-widest block italic mb-2">XP_Yield (Biomass)</span>
											<div class="flex items-end justify-between">
												<div class="text-2xl font-mono font-bold text-warning tracking-tighter flex items-center gap-2">
													{player.xp.toLocaleString()}
													<Dna class="w-4 h-4 opacity-20" />
												</div>
												<!-- Mini Sparkline Simulation -->
												<div class="flex items-end gap-0.5 h-6 mb-1">
													{#each Array(8) as _, i}
														<div class="w-1 bg-warning/20 group-hover/rack:bg-warning/40 transition-all" style="height: {20 + Math.random() * 80}%; transition-delay: {i * 50}ms"></div>
													{/each}
												</div>
											</div>
										</div>
										<div class="bg-black/40 p-4 border border-stone-800 industrial-frame flex flex-col justify-between">
											<span class="text-[7px] font-black text-stone-600 uppercase tracking-widest block italic">Clearance</span>
											<div class="text-xl font-heading font-black text-rust-light tracking-tighter text-right">LV_{Math.floor(player.xp / 1000) + 1}</div>
										</div>
									</div>

									<!-- Expandable Intel Dossier -->
									{#if isExpanded}
										<div class="space-y-6 pt-4 border-t border-stone-800/50" transition:slide>
											<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
												<div class="space-y-4">
													<div class="flex items-center gap-2 text-stone-500">
														<Fingerprint class="w-3.5 h-3.5" />
														<span class="text-[9px] font-black uppercase tracking-widest">Neural_Uplink_Signature</span>
													</div>
													<div class="bg-stone-950 p-3 border border-stone-800 text-[10px] font-mono text-text-dim break-all uppercase italic shadow-inner">
														{player.uid || 'NO_SIGNATURE_DETECTED'}
													</div>
												</div>
												<div class="space-y-4">
													<div class="flex items-center gap-2 text-stone-500">
														<Smartphone class="w-3.5 h-3.5" />
														<span class="text-[9px] font-black uppercase tracking-widest">Hardware_Hash_Vector</span>
													</div>
													<div class="bg-stone-950 p-3 border border-stone-800 text-[10px] font-mono text-text-dim break-all uppercase italic shadow-inner">
														{player.device_id || 'LOCAL_SIMULATION_NODE'}
													</div>
												</div>
											</div>

											<div class="grid grid-cols-3 gap-4">
												<div class="bg-stone-900/40 p-4 border border-stone-800 flex flex-col gap-2">
													<div class="flex items-center gap-2 text-stone-600">
														<MapPin class="w-3 h-3" />
														<span class="text-[8px] font-black uppercase tracking-widest">Last_Geo</span>
													</div>
													<span class="text-[10px] font-black text-stone-300 uppercase truncate">{player.last_joined_server?.split('-').pop() || 'Unknown'}</span>
												</div>
												<div class="bg-stone-900/40 p-4 border border-stone-800 flex flex-col gap-2">
													<div class="flex items-center gap-2 text-stone-600">
														<Clock class="w-3 h-3" />
														<span class="text-[8px] font-black uppercase tracking-widest">Enlisted</span>
													</div>
													<span class="text-[10px] font-black text-stone-300 uppercase">{new Date(player.created_at).toLocaleDateString()}</span>
												</div>
												<div class="bg-stone-900/40 p-4 border border-stone-800 flex flex-col gap-2">
													<div class="flex items-center gap-2 text-stone-600">
														<Shield class="w-3 h-3" />
														<span class="text-[8px] font-black uppercase tracking-widest">Status</span>
													</div>
													<span class={`text-[10px] font-black uppercase ${player.banned ? 'text-danger' : 'text-success'}`}>{player.banned ? 'Terminated' : 'Clear'}</span>
												</div>
											</div>
										</div>
									{/if}

									<!-- Interaction Footer -->
									<div class="flex items-center justify-between border-t border-stone-800/30 pt-4">
										<div class="flex items-center gap-4 text-text-dim">
											<div class="flex items-center gap-2">
												<Clock class="w-3 h-3 opacity-40" />
												<span class="text-[8px] font-black uppercase tracking-widest">Last_Uplink: {new Date(player.updated_at).toLocaleTimeString([], { hour12: false })}</span>
											</div>
										</div>
										
										<div class="flex gap-2">
											<Button 
												onclick={() => toggleBan(player)}
												variant={player.banned ? 'success' : 'danger'}
												size="xs"
											>
												{player.banned ? '[RESTORE_ACCESS]' : '[TERMINATE_UPLINK]'}
											</Button>
											<Button 
												onclick={() => confirmDelete(player)}
												variant="secondary"
												size="xs"
												class="!bg-red-950/10 hover:!bg-red-600 !text-stone-700 hover:!text-white"
											>
												[PURGE_FILE]
											</Button>
										</div>
									</div>
								</div>

								<!-- Background Aesthetics -->
								<div class="absolute inset-0 pointer-events-none opacity-[0.03] bg-[url('/grid.svg')] bg-center group-hover:opacity-[0.05] transition-opacity"></div>
								{#if isExpanded}
									<div class="absolute top-0 right-0 p-10 opacity-[0.02] pointer-events-none">
										<Dna class="w-40 h-40 text-rust" />
									</div>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			{:else}
				<!-- Incident Briefing Cards -->
				{#if reportsLoading && reports.length === 0}
					<div class="py-40 flex flex-col items-center gap-6">
						<div class="w-16 h-[1px] bg-danger animate-pulse"></div>
						<span class="font-heading font-black text-[12px] font-black tracking-[0.6em] uppercase text-red-600">INCIDENT_BUFFER_SYNC...</span>
					</div>
				{:else if filteredReports.length === 0}
					<div class="py-40 flex flex-col items-center gap-6 opacity-40">
						<div class="p-8 bg-stone-900/20 border border-stone-800 industrial-frame">
							<AlertOctagon class="w-16 h-16 opacity-10" />
						</div>
						<span class="font-heading font-black text-xl tracking-[0.4em] uppercase italic text-red-900">>> ZERO_INCIDENTS_MAPPED</span>
					</div>
				{:else}
					<div class="grid grid-cols-1 gap-4">
						{#each filteredReports as report (report.id)}
							<div 
								class="modern-industrial-card glass-panel group relative flex flex-col xl:flex-row items-stretch overflow-hidden shadow-2xl industrial-sharp"
								in:fade={{ duration: 200 }}
							>
								<!-- Tactical Corners -->
								<div class="corner-tl"></div>
								<div class="corner-tr"></div>
								<div class="corner-bl"></div>
								<div class="corner-br"></div>

								<!-- Fault Level Marker -->
								<div class="w-2 bg-danger shrink-0 relative overflow-hidden shadow-[0_0_20px_#ef4444] z-10">
									<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
								</div>

								<div class="flex-1 p-8 grid grid-cols-1 xl:grid-cols-12 gap-10 items-start relative z-10">
									<!-- Violation Detail -->
									<div class="xl:col-span-5 space-y-4">
										<div class="flex items-center gap-5">
											<div class="p-3 border border-red-600/30 bg-danger/10 text-danger industrial-frame shrink-0">
												<Icon name="ph:shield-warning-bold" size="1.5rem" class="animate-pulse" />
											</div>
											<div class="min-w-0">
												<h3 class="text-2xl font-heading font-black tracking-tighter text-white uppercase leading-none truncate">Report ID: {report.id}</h3>
												<p class="font-jetbrains text-[9px] font-black text-red-900 uppercase tracking-[0.3em] mt-2 italic font-bold">Severity: High</p>
											</div>
										</div>
										<div class="font-mono text-xs font-bold text-danger bg-red-950/10 border-l-2 border-red-600 p-5 tracking-widest uppercase leading-relaxed shadow-inner">
											&gt;&gt; "{report.reason.toUpperCase()}"
										</div>
									</div>

									<!-- Origin & Target Nodes -->
									<div class="xl:col-span-4 grid grid-cols-1 sm:grid-cols-2 gap-10 border-l border-stone-800/50 pl-10">
										<div class="space-y-3">
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] italic">Origin_Node_Sig</div>
											<div class="flex items-center gap-4">
												<div class="w-10 h-10 bg-stone-950 border border-stone-800 flex items-center justify-center text-xs font-heading font-black italic text-text-dim shrink-0 industrial-frame">{ (report.reporter_name || 'U').charAt(0).toUpperCase() }</div>
												<div class="min-w-0">
													<div class="font-heading text-sm font-black text-white italic uppercase truncate tracking-tight">{report.reporter_name || 'ANON_USER'}</div>
													<div class="font-mono text-[9px] text-stone-700 font-bold uppercase tracking-widest mt-1">HEX: 0x{report.reporter_id.toString(16).toUpperCase()}</div>
												</div>
											</div>
										</div>
										<div class="space-y-3">
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] italic">Target_Node_Sig</div>
											<div class="flex items-center gap-4">
												<div class="w-10 h-10 bg-red-950/30 border border-red-600/30 flex items-center justify-center text-xs font-heading font-black italic text-danger shrink-0 industrial-frame shadow-lg shadow-red-900/10">{ (report.reported_user_name || 'U').charAt(0).toUpperCase() }</div>
												<div class="min-w-0">
													<div class="font-heading text-sm font-black text-danger italic uppercase truncate underline decoration-1 decoration-red-600/20 underline-offset-4 tracking-tight">{report.reported_user_name || 'VOID_RECO'}</div>
													<div class="font-mono text-[9px] text-stone-700 font-bold uppercase tracking-widest mt-1">HEX: 0x{report.reported_user_id.toString(16).toUpperCase()}</div>
												</div>
											</div>
										</div>
									</div>

									<!-- Context & Time -->
									<div class="xl:col-span-3 space-y-6 border-l border-stone-800/50 pl-10">
										<div>
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] mb-3 italic">Sector_Linkage</div>
											{#if report.game_server_instance_id}
												<div class="flex items-center gap-3 font-mono text-[10px] font-black text-rust-light italic uppercase bg-stone-950 px-4 py-2 border border-stone-800 shadow-inner group-hover:border-rust/30 transition-all">
													<Icon name="server" size="1rem" class="text-rust shadow-rust/50" />
													<span class="tracking-widest">{report.game_server_instance_id.slice(0, 12)}...</span>
												</div>
											{:else}
												<div class="font-jetbrains text-[10px] text-stone-800 font-black italic tracking-[0.3em] uppercase opacity-30">&gt;&gt; NULL_MAP_PTR</div>
											{/if}
										</div>
										<div class="flex flex-col gap-1.5">
											<div class="font-mono text-base font-bold text-white italic tracking-tighter uppercase tabular-nums">{new Date(report.timestamp).toLocaleDateString()}</div>
											<div class="font-mono text-[10px] text-stone-700 font-bold italic tracking-[0.4em] uppercase opacity-80 tabular-nums">{new Date(report.timestamp).toLocaleTimeString([], { hour12: false })}</div>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Tactical Status Rail -->
		<div class="border-t border-stone-800 bg-[var(--header-bg)]/80 backdrop-blur-xl p-6 flex flex-col md:flex-row justify-between items-center gap-10 font-jetbrains text-[9px] font-black tracking-[0.4em] uppercase text-stone-600 italic industrial-frame shadow-[0_-10px_50px_rgba(0,0,0,0.5)]">
			<div class="flex flex-wrap justify-center gap-12 sm:gap-20">
				<div class="flex items-center gap-4 group/item cursor-default">
					<Activity class="w-4 h-4 text-emerald-500 animate-pulse" />
					<span class="group-hover/item:text-stone-400 transition-colors">Uplink_Signal: Optimized</span>
				</div>
				<div class="flex items-center gap-4 group/item cursor-default">
					<Shield class="w-4 h-4 text-rust" />
					<span class="group-hover/item:text-stone-400 transition-colors">Neural_Armor: Active</span>
				</div>
				<div class="flex items-center gap-4 group/item cursor-default">
					<Database class="w-4 h-4 text-rust opacity-40" />
					<span class="group-hover/item:text-stone-400 transition-colors">Data_Integrity: 100%</span>
				</div>
			</div>
			<div class="flex items-center gap-5 text-stone-800 font-black px-6 py-2 border border-stone-800 bg-black/40 shadow-inner industrial-frame group">
				<div class="w-2 h-2 rounded-full bg-rust shadow-[0_0_10px_var(--color-rust)] group-hover:scale-125 transition-transform"></div>
				<span class="text-white opacity-30 text-[10px]">OS_KERN: 0.9.4_X_HD</span>
			</div>
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
		title="TERMINATE_SUBJECT_RECORD"
		isCritical={true}
		message={`CONFIRM PERMANENT TERMINATION OF SUBJECT "${playerToDelete?.name.toUpperCase()}". DATA WIPE WILL COMMENCE ACROSS ALL SECTORS.`}
		onConfirm={handleDeletePlayer}
		on:close={() => (isDeleteConfirmOpen = false)}
	/>
</div>

<style>
	/* Cinematic Intelligence Interface Styles */
	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.8) 100%);
	}

	@keyframes glow-slide {
		0% { transform: translateX(-100%); }
		100% { transform: translateX(100%); }
	}
	.animate-glow-slide {
		animation: glow-slide 2s infinite linear;
	}

	/* Tactical High-Res Scrollbar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
		height: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: var(--terminal-bg);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
		border: 1px solid var(--terminal-bg);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}

	:global(body) {
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
		background-color: var(--terminal-bg);
	}

	input:focus {
		outline: none;
	}
</style>