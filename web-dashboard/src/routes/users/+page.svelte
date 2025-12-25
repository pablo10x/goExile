<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
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
		Check
	} from 'lucide-svelte';
	import EditPlayerModal from '$lib/components/players/EditPlayerModal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { notifications } from '$lib/stores';

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
	let selectedPlayer = $state<Player | null>(null);
	let isEditModalOpen = $state(false);
	let isDeleteConfirmOpen = $state(false);
	let playerToDelete = $state<Player | null>(null);

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

	onMount(() => {
		fetchPlayers();
		fetchReports();
		const interval = setInterval(() => {
			if (activeTab === 'players') fetchPlayers();
		}, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="w-full h-full flex flex-col overflow-hidden relative font-sans">
	<!-- Cinematic Overlays -->
	<div class="fixed inset-0 pointer-events-none z-[100] bg-vignette opacity-40"></div>
	
	<!-- Main Content Chassis -->
	<div class="w-full h-full flex flex-col gap-10 relative z-10 pb-32 md:pb-12">
		
		<!-- Intelligence Header (Responsive Scale) -->
		<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-4 border-rust pl-10 py-2 bg-[#0a0a0a]/60 backdrop-blur-xl shadow-2xl relative overflow-hidden industrial-frame">
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
			
			<div class="space-y-4 p-2 relative z-10">
				<div class="flex items-center gap-4">
					<span class="bg-rust text-white px-3 py-1 text-[10px] font-black uppercase tracking-[0.2em] shadow-lg shadow-rust/20">CLASSIFIED_ACCESS</span>
					<div class="w-px h-3 bg-stone-800"></div>
					<span class="font-jetbrains text-[9px] font-black text-stone-500 uppercase tracking-[0.4em] italic">STATION: EXILE_HIVE_CORE</span>
				</div>
				<h1 class="text-5xl sm:text-6xl lg:text-7xl font-heading font-black tracking-tighter text-white uppercase leading-none">
					Player_<span class="text-rust">Registry</span>
				</h1>
				<div class="flex items-center gap-6 pt-2">
					<div class="flex items-center gap-3 font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest">
						<div class="w-2 h-2 bg-emerald-500 shadow-[0_0_10px_#10b981] animate-pulse"></div>
						SIGNAL_LOCK: STABLE
					</div>
					<div class="w-px h-4 bg-stone-800"></div>
					<div class="font-jetbrains text-[10px] font-black text-stone-600 uppercase tracking-widest italic">
						AUTH_VECTOR: ROOT_ADMIN
					</div>
				</div>
			</div>

			<div class="flex flex-wrap items-center gap-8 p-4 relative z-10">
				<!-- Tactical Tab Switcher -->
				<div class="flex gap-2 bg-black/40 p-1.5 border border-stone-800 shadow-inner">
					<button
						onclick={() => (activeTab = 'players')}
						class="flex flex-col items-start px-8 py-4 transition-all duration-500 relative group {activeTab === 'players' ? 'bg-rust text-white shadow-xl shadow-rust/30' : 'text-stone-600 hover:text-stone-300 hover:bg-stone-900'}"
					>
						<span class="font-jetbrains text-[8px] font-black tracking-[0.3em] uppercase mb-1 opacity-50">Identity Base</span>
						<span class="font-heading text-base font-black tracking-widest uppercase">Players</span>
						{#if activeTab === 'players'}
							<div class="absolute -top-1 -right-1 w-2 h-2 bg-rust shadow-[0_0_10px_var(--color-rust)]"></div>
						{/if}
					</button>
					<button
						onclick={() => (activeTab = 'reports')}
						class="flex flex-col items-start px-8 py-4 transition-all duration-500 relative group {activeTab === 'reports' ? 'bg-red-600 text-white shadow-xl shadow-red-900/30' : 'text-stone-600 hover:text-stone-300 hover:bg-stone-900'}"
					>
						<span class="font-jetbrains text-[8px] font-black tracking-[0.3em] uppercase mb-1 opacity-50">Violation Logs</span>
						<span class="font-heading text-base font-black tracking-widest uppercase">Reports</span>
						{#if activeTab === 'reports'}
							<div class="absolute -top-1 -right-1 w-2 h-2 bg-red-600 shadow-[0_0_100px_#ef4444]"></div>
						{/if}
					</button>
				</div>

				<button
					onclick={refreshCurrentTab}
					disabled={activeTab === 'players' ? playersLoading : reportsLoading}
					class="p-5 border border-stone-800 bg-stone-950/60 hover:bg-rust hover:text-white hover:border-rust transition-all shadow-xl active:translate-y-px disabled:opacity-20"
				>
					<RefreshCw class="w-6 h-6 {(activeTab === 'players' ? playersLoading : reportsLoading) ? 'animate-spin' : ''}" />
				</button>
			</div>
		</div>

		<!-- Command & Query Line -->
		<div class="grid grid-cols-1 xl:grid-cols-12 gap-8 items-center px-2">
			<div class="xl:col-span-10 relative group">
				<div class="absolute left-6 top-1/2 -translate-y-1/2 text-rust font-black text-xl pointer-events-none opacity-50 group-focus-within:opacity-100 transition-opacity">$</div>
				{#if activeTab === 'players'}
					<input
						type="text"
						bind:value={playerSearchQuery}
						placeholder="FILTER_BY_NEURAL_SIGNATURE..."
						class="w-full bg-stone-950 border border-stone-800 pl-14 pr-10 py-5 font-jetbrains text-sm text-stone-200 font-bold tracking-[0.2em] focus:border-rust outline-none transition-all uppercase placeholder:text-stone-900 italic shadow-inner industrial-frame"
					/>
				{:else}
					<input
						type="text"
						bind:value={reportSearchQuery}
						placeholder="FILTER_BY_FAULT_LOG_VECTOR..."
						class="w-full bg-stone-950 border border-stone-800 pl-14 pr-10 py-5 font-jetbrains text-sm text-stone-200 font-bold tracking-[0.2em] focus:border-rust outline-none transition-all uppercase placeholder:text-stone-900 italic shadow-inner industrial-frame"
					/>
				{/if}
			</div>
			<div class="xl:col-span-2 flex items-center justify-end gap-5 font-jetbrains text-[10px] font-black text-stone-600 uppercase tracking-[0.3em] italic shrink-0">
				<div class="flex items-center gap-3">
					<div class="w-1.5 h-1.5 bg-rust animate-ping"></div>
					SYNC_ACTIVE
				</div>
				<div class="w-px h-4 bg-stone-800"></div>
				<Activity class="w-4 h-4 text-rust" />
			</div>
		</div>

		<!-- Data Briefing Cards -->
		<div class="flex-1 overflow-y-auto custom-scrollbar px-2 space-y-4">
			{#if activeTab === 'players'}
				{#if playersLoading && players.length === 0}
					<div class="py-40 flex flex-col items-center gap-6">
						<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
						<span class="font-heading font-black text-[12px] font-black tracking-[0.6em] uppercase text-rust animate-pulse">Uplink_Synchronizing...</span>
					</div>
				{:else if filteredPlayers.length === 0}
					<div class="py-40 flex flex-col items-center gap-6 opacity-40">
						<div class="p-8 bg-stone-900/20 border border-stone-800 industrial-frame">
							<Search class="w-16 h-16 opacity-10" />
						</div>
						<span class="font-heading font-black text-xl tracking-[0.4em] uppercase italic text-stone-700">>> ZERO_RECORDS_MAPPED</span>
					</div>
				{:else}
					<div class="grid grid-cols-1 gap-4">
						{#each filteredPlayers as player (player.id)}
							<div 
								class="modern-industrial-card glass-panel group relative flex flex-col xl:flex-row items-stretch overflow-hidden shadow-2xl !rounded-none"
								in:fade={{ duration: 200 }}
							>
								<!-- Subject Signal Marker -->
								<div class={`w-2 ${player.online ? 'bg-emerald-500 shadow-[0_0_15px_#10b981]' : 'bg-stone-800'} shrink-0 relative transition-colors duration-500`}>
									{#if player.online}
										<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
									{/if}
								</div>

								<!-- Identification Block -->
								<div class="flex-1 p-8 grid grid-cols-1 lg:grid-cols-12 gap-10 items-center">
									<div class="lg:col-span-4 flex items-center gap-8">
										<div class="w-16 h-16 bg-stone-950 border border-stone-800 flex items-center justify-center text-white font-heading font-black italic text-2xl group-hover:border-rust group-hover:text-rust transition-all duration-500 industrial-frame shadow-inner">
											{player.name.charAt(0).toUpperCase()}
										</div>
										<div class="min-w-0">
											<h3 class="text-3xl font-heading font-black tracking-tighter text-white uppercase leading-none group-hover:text-rust transition-all duration-500 truncate">{player.name}</h3>
											<p class="font-jetbrains text-[10px] font-black text-stone-600 tracking-[0.3em] uppercase mt-2 italic">Player ID: 0x{player.id.toString(16).toUpperCase()}</p>
										</div>
									</div>

									<!-- Metadata Briefing -->
									<div class="lg:col-span-5 grid grid-cols-2 gap-10 border-l border-stone-800/50 pl-10 font-jetbrains">
										<div>
											<div class="text-[9px] font-black text-stone-700 uppercase tracking-[0.2em] mb-2 italic">Registry_UID_Hash</div>
											<div class="text-xs font-black text-stone-500 tracking-tight truncate uppercase italic">{player.uid || 'NULL_PTR'}</div>
										</div>
										<div>
											<div class="text-[9px] font-black text-stone-700 uppercase tracking-[0.2em] mb-2 italic">Accumulated_XP</div>
											<div class="text-2xl font-heading font-black text-amber-500 tracking-tighter leading-none flex items-baseline gap-2">
												{player.xp.toLocaleString()} <span class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-widest">UNITS</span>
											</div>
										</div>
									</div>

									<!-- Uplink Info -->
									<div class="lg:col-span-3 space-y-2 border-l border-stone-800/50 pl-10 hidden xl:block font-jetbrains">
										<div class="flex items-center gap-3">
											<Clock class="w-4 h-4 text-stone-700" />
											<span class="text-[11px] font-black text-white uppercase tracking-widest">{new Date(player.updated_at).toLocaleDateString()}</span>
										</div>
										<div class="font-jetbrains text-[9px] font-black text-rust-light tracking-[0.3em] uppercase pl-7 opacity-60">
											TIMESTAMP: {new Date(player.updated_at).toLocaleTimeString([], { hour12: false })}
										</div>
									</div>
								</div>

								<!-- Tactical Actions -->
								<div class="flex flex-row xl:flex-col divide-x xl:divide-x-0 xl:divide-y divide-stone-800 border-l border-stone-800 bg-stone-900/20">
									<button 
										onclick={() => openEditModal(player)}
										class="flex-1 px-8 py-6 xl:py-0 xl:h-1/2 hover:bg-rust/20 hover:text-rust transition-all flex items-center justify-center group/btn text-stone-600"
										title="Modify_Subject"
									>
										<Pencil class="w-5 h-5 transition-transform duration-500 group-hover/btn:scale-125" />
									</button>
									<button 
										onclick={() => confirmDelete(player)}
										class="flex-1 px-8 py-6 xl:py-0 xl:h-1/2 hover:bg-red-600/20 hover:text-red-500 transition-all flex items-center justify-center group/btn text-stone-600"
										title="Purge_Subject"
									>
										<Trash2 class="w-5 h-5 transition-transform duration-500 group-hover/btn:scale-125" />
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{:else}
				<!-- Incident Briefing Cards -->
				{#if reportsLoading && reports.length === 0}
					<div class="py-40 flex flex-col items-center gap-6">
						<div class="w-16 h-[1px] bg-red-600 animate-pulse"></div>
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
								class="modern-industrial-card glass-panel group relative flex flex-col xl:flex-row items-stretch overflow-hidden shadow-2xl !rounded-none"
								in:fade={{ duration: 200 }}
							>
								<!-- Fault Level Marker -->
								<div class="w-2 bg-red-600 shrink-0 relative overflow-hidden shadow-[0_0_20px_#ef4444]">
									<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
								</div>

								<div class="flex-1 p-8 grid grid-cols-1 xl:grid-cols-12 gap-10 items-start">
									<!-- Violation Detail -->
									<div class="xl:col-span-5 space-y-4">
										<div class="flex items-center gap-5">
											<div class="p-3 border border-red-600/30 bg-red-600/10 text-red-500 industrial-frame shrink-0">
												<ShieldAlert class="w-6 h-6 animate-pulse" />
											</div>
											<div class="min-w-0">
												<h3 class="text-2xl font-heading font-black tracking-tighter text-white uppercase leading-none truncate">Report ID: {report.id}</h3>
												<p class="font-jetbrains text-[9px] font-black text-red-900 uppercase tracking-[0.3em] mt-2 italic font-bold">Severity: High</p>
											</div>
										</div>
										<div class="font-jetbrains text-xs font-bold text-red-500 bg-red-950/10 border-l-2 border-red-600 p-5 tracking-widest uppercase leading-relaxed shadow-inner">
											&gt;&gt; "{report.reason.toUpperCase()}"
										</div>
									</div>

									<!-- Origin & Target Nodes -->
									<div class="xl:col-span-4 grid grid-cols-1 sm:grid-cols-2 gap-10 border-l border-stone-800/50 pl-10">
										<div class="space-y-3">
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] italic">Origin_Node_Sig</div>
											<div class="flex items-center gap-4">
												<div class="w-10 h-10 bg-stone-950 border border-stone-800 flex items-center justify-center text-xs font-heading font-black italic text-stone-500 shrink-0 industrial-frame">{ (report.reporter_name || 'U').charAt(0).toUpperCase() }</div>
												<div class="min-w-0">
													<div class="font-heading text-sm font-black text-white italic uppercase truncate tracking-tight">{report.reporter_name || 'ANON_USER'}</div>
													<div class="font-jetbrains text-[9px] text-stone-700 font-bold uppercase tracking-widest mt-1">HEX: 0x{report.reporter_id.toString(16).toUpperCase()}</div>
												</div>
											</div>
										</div>
										<div class="space-y-3">
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] italic">Target_Node_Sig</div>
											<div class="flex items-center gap-4">
												<div class="w-10 h-10 bg-red-950/30 border border-red-600/30 flex items-center justify-center text-xs font-heading font-black italic text-red-500 shrink-0 industrial-frame shadow-lg shadow-red-900/10">{ (report.reported_user_name || 'U').charAt(0).toUpperCase() }</div>
												<div class="min-w-0">
													<div class="font-heading text-sm font-black text-red-500 italic uppercase truncate underline decoration-1 decoration-red-600/20 underline-offset-4 tracking-tight">{report.reported_user_name || 'VOID_RECO'}</div>
													<div class="font-jetbrains text-[9px] text-stone-700 font-bold uppercase tracking-widest mt-1">HEX: 0x{report.reported_user_id.toString(16).toUpperCase()}</div>
												</div>
											</div>
										</div>
									</div>

									<!-- Context & Time -->
									<div class="xl:col-span-3 space-y-6 border-l border-stone-800/50 pl-10">
										<div>
											<div class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-[0.3em] mb-3 italic">Sector_Linkage</div>
											{#if report.game_server_instance_id}
												<div class="flex items-center gap-3 font-jetbrains text-[10px] font-black text-rust-light italic uppercase bg-stone-950 px-4 py-2 border border-stone-800 shadow-inner group-hover:border-rust/30 transition-all">
													<Server class="w-4 h-4 text-rust shadow-rust/50" />
													<span class="tracking-widest">{report.game_server_instance_id.slice(0, 12)}...</span>
												</div>
											{:else}
												<div class="font-jetbrains text-[10px] text-stone-800 font-black italic tracking-[0.3em] uppercase opacity-30">&gt;&gt; NULL_MAP_PTR</div>
											{/if}
										</div>
										<div class="flex flex-col gap-1.5">
											<div class="font-heading text-base font-black text-white italic tracking-tighter uppercase tabular-nums">{new Date(report.timestamp).toLocaleDateString()}</div>
											<div class="font-jetbrains text-[10px] text-stone-700 font-black italic tracking-[0.4em] uppercase opacity-80 tabular-nums">{new Date(report.timestamp).toLocaleTimeString([], { hour12: false })}</div>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Tactical Intel Footer -->
		<div class="border-t border-stone-800 bg-[#0a0a0a]/60 p-8 flex flex-col md:flex-row justify-between items-center gap-10 font-jetbrains text-[10px] font-black tracking-[0.4em] uppercase text-stone-700 italic industrial-frame">
			<div class="flex flex-wrap justify-center gap-16">
				<div class="flex items-center gap-4 group/item cursor-default">
					<Cpu class="w-5 h-5 text-rust group-hover/item:scale-125 transition-all duration-500 opacity-40 group-hover:opacity-100" />
					<span class="group-hover/item:text-stone-400 transition-colors">Core_Load: Nominal</span>
				</div>
				<div class="flex items-center gap-4 group/item cursor-default">
					<Database class="w-5 h-5 text-rust group-hover/item:scale-125 transition-all duration-500 opacity-40 group-hover:opacity-100" />
					<span class="group-hover/item:text-stone-400 transition-colors">Registry: Synchronized</span>
				</div>
				<div class="flex items-center gap-4 group/item cursor-default text-white/80 border-b border-rust/30 pb-1">
					<Lock class="w-5 h-5 text-rust shadow-rust/50" />
					<span class="tracking-[0.5em]">Cipher: AES_256_GCM</span>
				</div>
			</div>
			<div class="flex items-center gap-5 text-stone-800 font-black px-6 py-3 border border-stone-800 bg-black/40 shadow-inner industrial-frame">
				<ChevronRight class="w-5 h-5 animate-pulse text-rust" />
				<span class="text-white opacity-40">Ver: 0.9.4-TAC_PRIME_HD</span>
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
	/* Cinematic Command UI Styles (High-Definition Modern Industrial) */
	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.8) 100%);
	}

	@keyframes warmup {
		0% { opacity: 1; filter: contrast(3) brightness(0); }
		15% { opacity: 1; filter: contrast(2) brightness(1.5); }
		30% { opacity: 1; filter: contrast(1.5) brightness(0.5); }
		100% { opacity: 0; filter: contrast(1) brightness(1); visibility: hidden; }
	}
	.animate-warmup {
		animation: warmup 1.5s forwards cubic-bezier(0.23, 1, 0.32, 1);
	}

	/* Industrial High-Res Scrollbar (Modernized) */
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
		height: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: #050505;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
		border: 1px solid #050505;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}

	:global(body) {
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
		background-color: #050505;
	}

	input:focus {
		outline: none;
	}
</style>