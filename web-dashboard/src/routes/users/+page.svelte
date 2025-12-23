<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
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
		Signal
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

<div class="min-h-screen bg-[#0a0a0a] text-[#a0a0a0] font-['JetBrains_Mono',monospace] p-0 relative overflow-hidden selection:bg-[#f97316] selection:text-black font-medium">
	<!-- Cinematic Overlays -->
	<div class="fixed inset-0 pointer-events-none z-[100] bg-vignette opacity-40"></div>
	
	<!-- Main Content Chassis -->
	<div class="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-10 space-y-8 relative z-10">
		
		<!-- Intelligence Header (Responsive Scale) -->
		<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-2 border-[#f97316] pl-6 py-1 bg-[#121212]/40 backdrop-blur-md">
			<div class="space-y-3 p-2">
				<div class="flex items-center gap-3">
					<span class="bg-[#f97316] text-black px-1.5 py-0.5 text-[9px] font-black uppercase tracking-widest">Classified</span>
					<span class="text-[9px] font-bold text-slate-600 uppercase tracking-[0.3em] italic font-['JetBrains_Mono']">STATION: EXILE_HIVE</span>
				</div>
				<h1 class="text-4xl sm:text-5xl lg:text-6xl font-black tracking-tighter text-white uppercase font-['Inter',sans-serif] leading-none">
					Subject_<span class="text-[#f97316]">Registry</span>
				</h1>
				<div class="flex items-center gap-4 pt-1">
					<div class="flex items-center gap-2 text-[10px] font-bold text-slate-500 uppercase tracking-[0.1em]">
						<div class="w-1.5 h-1.5 bg-[#f97316] shadow-[0_0_8px_#f97316]"></div>
						Signal_Stable
					</div>
					<div class="w-[1px] h-3 bg-white/10"></div>
					<div class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.1em] italic">
						Role: Super_Admin
					</div>
				</div>
			</div>

			<div class="flex flex-wrap items-center gap-6 p-2">
				<!-- Tactical Tab Switcher -->
				<div class="flex gap-2">
					<button
						onclick={() => (activeTab = 'players')}
						class="flex flex-col items-start px-6 py-3 border transition-all duration-300 relative {activeTab === 'players' ? 'border-[#f97316] bg-[#f97316]/10 text-white shadow-[0_0_15px_rgba(249,115,22,0.1)]' : 'border-white/10 bg-[#1a1a1a]/40 text-slate-600'}"
					>
						<span class="text-[8px] font-black tracking-widest uppercase mb-0.5 opacity-50 font-['JetBrains_Mono']">Identity_Base</span>
						<span class="text-sm font-black tracking-tight uppercase font-['Inter',sans-serif]">01_SUBJECTS</span>
						{#if activeTab === 'players'}
							<div class="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 bg-[#f97316]"></div>
						{/if}
					</button>
					<button
						onclick={() => (activeTab = 'reports')}
						class="flex flex-col items-start px-6 py-3 border transition-all duration-300 relative {activeTab === 'reports' ? 'border-red-600 bg-red-600/10 text-white shadow-[0_0_15px_rgba(239,68,68,0.1)]' : 'border-white/10 bg-[#1a1a1a]/40 text-slate-600'}"
					>
						<span class="text-[8px] font-black tracking-widest uppercase mb-0.5 opacity-50 font-['JetBrains_Mono']">Violation_Logs</span>
						<span class="text-sm font-black tracking-tight uppercase font-['Inter',sans-serif]">02_INCIDENTS</span>
						{#if activeTab === 'reports'}
							<div class="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 bg-red-600"></div>
						{/if}
					</button>
				</div>

				<button
					onclick={refreshCurrentTab}
					disabled={activeTab === 'players' ? playersLoading : reportsLoading}
					class="p-4 border border-white/10 bg-[#1a1a1a]/40 hover:border-white/30 transition-all text-slate-600 disabled:opacity-20"
				>
					<RefreshCw class="w-5 h-5 {(activeTab === 'players' ? playersLoading : reportsLoading) ? 'animate-spin' : ''}" />
				</button>
			</div>
		</div>

		<!-- Command & Query Line -->
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-6 lg:items-center">
			<div class="lg:col-span-9 relative font-['JetBrains_Mono']">
				<div class="absolute left-5 top-1/2 -translate-y-1/2 text-[#f97316] font-black text-lg pointer-events-none">$</div>
				{#if activeTab === 'players'}
					<input
						type="text"
						bind:value={playerSearchQuery}
						placeholder="FILTER_BY_SUBJECT_ID..."
						class="w-full bg-[#121212] border border-white/10 pl-12 pr-10 py-4 text-base text-white font-bold tracking-widest focus:border-[#f97316]/40 transition-all outline-none uppercase placeholder:text-slate-800 italic shadow-inner"
					/>
				{:else}
					<input
						type="text"
						bind:value={reportSearchQuery}
						placeholder="FILTER_BY_FAULT_LOG..."
						class="w-full bg-[#121212] border border-white/10 pl-12 pr-10 py-4 text-base text-white font-bold tracking-widest focus:border-[#f97316]/40 transition-all outline-none uppercase placeholder:text-slate-800 italic shadow-inner"
					/>
				{/if}
			</div>
			<div class="lg:col-span-3 flex items-center justify-end gap-4 text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] italic shrink-0">
				<Activity class="w-3.5 h-3.5 text-[#f97316] animate-pulse" />
				Streaming_Data
			</div>
		</div>

		<!-- Data Briefing Cards -->
		<div class="space-y-3">
			{#if activeTab === 'players'}
				{#if playersLoading && players.length === 0}
					<div class="py-24 flex flex-col items-center gap-4 opacity-40">
						<div class="w-12 h-[1px] bg-[#f97316] animate-pulse"></div>
						<span class="text-[10px] font-black tracking-[0.5em] uppercase text-[#f97316]">Uplink_Sync</span>
					</div>
				{:else if filteredPlayers.length === 0}
					<div class="py-24 flex flex-col items-center gap-4 opacity-20">
						<span class="text-lg font-black tracking-[0.3em] uppercase italic text-[#f97316]">>> NO_RECORDS_LOCATED</span>
					</div>
				{:else}
					{#each filteredPlayers as player (player.id)}
						<div 
							class="group relative flex flex-col lg:flex-row items-stretch bg-[#121212] border border-white/10 hover:border-white/30 transition-all duration-300 overflow-hidden shadow-lg hover:shadow-[#f97316]/5"
							in:fade={{ duration: 200 }}
						>
							<!-- Subject Signal Marker -->
							<div class="w-1.5 {player.online ? 'bg-[#10b981]' : 'bg-slate-800'} shrink-0 relative">
								{#if player.online}
									<div class="absolute inset-0 bg-white/10 animate-pulse"></div>
								{/if}
							</div>

							<!-- Identification Block -->
							<div class="flex-1 p-6 grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
								<div class="md:col-span-4 flex items-center gap-5">
									<div class="w-10 h-10 bg-black border border-white/10 flex items-center justify-center text-white font-black italic text-lg group-hover:border-[#f97316] transition-colors shrink-0">
										{player.name.charAt(0).toUpperCase()}
									</div>
									<div class="min-w-0">
										<h3 class="text-2xl font-black tracking-tight text-white uppercase font-['Inter',sans-serif] leading-none group-hover:text-[#f97316] transition-colors truncate">{player.name}</h3>
										<p class="text-[9px] font-black text-slate-600 tracking-[0.2em] uppercase mt-1.5 italic font-['JetBrains_Mono']">ID: 0x{player.id.toString(16).toUpperCase()}</p>
									</div>
								</div>

								<!-- Metadata Briefing -->
								<div class="md:col-span-5 grid grid-cols-2 gap-6 border-l border-white/10 pl-6 font-['JetBrains_Mono']">
									<div>
										<div class="text-[8px] font-black text-slate-700 uppercase tracking-widest mb-1 italic">Registry_UID</div>
										<div class="text-[11px] font-bold text-slate-500 tracking-tight truncate uppercase italic">{player.uid || 'N/A'}</div>
									</div>
									<div>
										<div class="text-[8px] font-black text-slate-700 uppercase tracking-widest mb-1 italic">Stored_XP</div>
										<div class="text-lg font-black text-[#fbbf24] italic tracking-tight leading-none">
											{player.xp.toLocaleString()} <span class="text-[8px] opacity-30 uppercase font-bold text-[#f97316]">Units</span>
										</div>
									</div>
								</div>

								<!-- Uplink Info -->
								<div class="md:col-span-3 space-y-1.5 border-l border-white/10 pl-6 hidden lg:block font-['JetBrains_Mono']">
									<div class="flex items-center gap-2">
										<Clock class="w-3 h-3 text-slate-600" />
										<span class="text-[10px] font-bold text-white uppercase tracking-tighter">{new Date(player.updated_at).toLocaleDateString()}</span>
									</div>
									<div class="text-[9px] font-black text-[#f97316] tracking-[0.2em] uppercase pl-5 opacity-50">
										{new Date(player.updated_at).toLocaleTimeString()}
									</div>
								</div>
							</div>

							<!-- Tactical Actions -->
							<div class="flex flex-row lg:flex-col divide-x lg:divide-x-0 lg:divide-y divide-white/10 border-l border-white/10 bg-[#1a1a1a]/60">
								<button 
									onclick={() => openEditModal(player)}
									class="flex-1 px-6 py-4 lg:py-0 lg:h-1/2 hover:bg-[#f97316]/20 hover:text-[#f97316] transition-all flex items-center justify-center group/btn text-slate-600"
								>
									<Pencil class="w-4 h-4 transition-transform group-hover/btn:scale-110" />
								</button>
								<button 
									onclick={() => confirmDelete(player)}
									class="flex-1 px-6 py-4 lg:py-0 lg:h-1/2 hover:bg-red-600/20 hover:text-red-600 transition-all flex items-center justify-center group/btn text-slate-600"
								>
									<Trash2 class="w-4 h-4 transition-transform group-hover/btn:scale-110" />
								</button>
							</div>
						</div>
					{/each}
				{/if}
			{:else}
				<!-- Incident Briefing Cards -->
				{#if reportsLoading && reports.length === 0}
					<div class="py-24 flex flex-col items-center gap-4 opacity-40">
						<div class="w-12 h-[1px] bg-red-600 animate-pulse"></div>
						<span class="text-[10px] font-black tracking-[0.5em] uppercase text-red-600">Fault_Log_Sync</span>
					</div>
				{:else if filteredReports.length === 0}
					<div class="py-24 flex flex-col items-center gap-4 opacity-20">
						<span class="text-lg font-black tracking-[0.3em] uppercase italic text-red-600">>> NO_INCIDENTS_LOGGED</span>
					</div>
				{:else}
					{#each filteredReports as report (report.id)}
						<div 
							class="group relative flex flex-col lg:flex-row items-stretch bg-[#121212] border border-white/10 hover:border-red-600/30 transition-all duration-300 overflow-hidden font-['JetBrains_Mono'] shadow-lg"
							in:fade={{ duration: 200 }}
						>
							<!-- Fault Level Marker -->
							<div class="w-1.5 bg-red-600 shrink-0 relative overflow-hidden opacity-60">
								<div class="absolute inset-0 bg-white/10 animate-pulse"></div>
							</div>

							<div class="flex-1 p-6 grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
								<!-- Violation Detail -->
								<div class="lg:col-span-5 space-y-3">
									<div class="flex items-center gap-4">
										<div class="p-2 border border-red-600/20 bg-red-600/10 text-red-600 shrink-0">
											<ShieldAlert class="w-5 h-5 animate-pulse" />
										</div>
										<div class="min-w-0">
											<h3 class="text-xl font-black tracking-tight text-white uppercase font-['Inter',sans-serif] leading-none truncate">Fault_{report.id}</h3>
											<p class="text-[8px] font-black text-slate-600 tracking-[0.2em] uppercase mt-1.5 italic">Threat_Priority: HIGH</p>
										</div>
									</div>
									<div class="text-[11px] font-bold text-red-600/80 italic border-l-2 border-red-600/20 pl-4 py-1.5 bg-white/[0.02] tracking-wide uppercase leading-relaxed">
										&gt;&gt; "{report.reason.toUpperCase()}"
									</div>
								</div>

								<!-- Origin & Target Nodes -->
								<div class="lg:col-span-4 grid grid-cols-1 sm:grid-cols-2 gap-8 border-l border-white/10 pl-8">
									<div class="space-y-2">
										<div class="text-[8px] font-black text-slate-700 uppercase tracking-widest italic">Origin_Node</div>
										<div class="flex items-center gap-3">
											<div class="w-7 h-7 bg-black border border-white/10 flex items-center justify-center text-[10px] font-black italic text-slate-500 shrink-0">{ (report.reporter_name || 'U').charAt(0).toUpperCase() }</div>
											<div class="min-w-0">
												<div class="text-[11px] font-black text-white italic uppercase truncate">{report.reporter_name || 'ANON'}</div>
												<div class="text-[8px] text-slate-700 font-bold uppercase tracking-widest">ID: {report.reporter_id}</div>
											</div>
										</div>
									</div>
									<div class="space-y-2">
										<div class="text-[8px] font-black text-slate-700 uppercase tracking-widest italic">Target_Node</div>
										<div class="flex items-center gap-3">
											<div class="w-7 h-7 bg-red-950/20 border border-red-600/20 flex items-center justify-center text-[10px] font-black italic text-red-600 shrink-0">{ (report.reported_user_name || 'U').charAt(0).toUpperCase() }</div>
											<div class="min-w-0">
												<div class="text-[11px] font-black text-red-600 italic uppercase truncate underline decoration-1 decoration-red-600/20 underline-offset-2">{report.reported_user_name || 'VOID'}</div>
												<div class="text-[8px] text-slate-700 font-bold uppercase tracking-widest">ID: {report.reported_user_id}</div>
											</div>
										</div>
									</div>
								</div>

								<!-- Context & Time -->
								<div class="lg:col-span-3 space-y-4 border-l border-white/10 pl-8">
									<div>
										<div class="text-[8px] font-black text-slate-700 uppercase tracking-widest mb-1.5 italic">Sector_Link</div>
										{#if report.game_server_instance_id}
											<div class="flex items-center gap-2.5 text-[9px] font-black text-slate-500 italic uppercase bg-black/40 px-3 py-1.5 border border-white/10">
												<Server class="w-3 h-3 text-[#f97316]" />
												<span class="tracking-widest">{report.game_server_instance_id.slice(0, 8)}...</span>
											</div>
										{:else}
											<span class="text-[8px] text-slate-800 font-black italic tracking-[0.2em] uppercase opacity-40 italic">&gt;&gt; NULL_LINK</span>
										{/if}
									</div>
									<div class="flex flex-col gap-1">
										<div class="text-[11px] font-black text-white italic tracking-tighter uppercase font-['Inter',sans-serif]">{new Date(report.timestamp).toLocaleDateString()}</div>
										<div class="text-[9px] text-slate-700 font-black italic tracking-[0.2em] uppercase opacity-80">{new Date(report.timestamp).toLocaleTimeString()}</div>
									</div>
								</div>
							</div>
						</div>
					{/each}
				{/if}
			{/if}
		</div>

		<!-- Tactical Intel Footer -->
		<div class="border-t border-white/10 bg-[#121212]/40 pt-10 flex flex-col md:flex-row justify-between items-center gap-10 text-[10px] font-black tracking-[0.4em] uppercase text-slate-700 italic font-['JetBrains_Mono']">
			<div class="flex flex-wrap justify-center gap-10">
				<div class="flex items-center gap-3 group/item cursor-default">
					<Cpu class="w-4 h-4 text-[#f97316] group-hover/item:scale-110 transition-transform opacity-60" />
					<span class="group-hover/item:text-slate-500 transition-colors">Core_Load: Nominal</span>
				</div>
				<div class="flex items-center gap-3 group/item cursor-default">
					<Database class="w-4 h-4 text-[#f97316] group-hover/item:scale-110 transition-transform opacity-60" />
					<span class="group-hover/item:text-slate-500 transition-colors">Registry: Active</span>
				</div>
				<div class="flex items-center gap-3 group/item cursor-default text-white opacity-80">
					<Lock class="w-4 h-4 text-[#f97316]" />
					<span>Encryption: AES_256</span>
				</div>
			</div>
			<div class="flex items-center gap-4 text-slate-800 font-bold px-4 py-2 border border-white/5 bg-black/20">
				<ChevronRight class="w-4 h-4 animate-pulse text-[#f97316]" />
				<span>Ver: 0.9.1-TAC_HD</span>
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
	/* Cinematic Command UI Styles (Lighter Contrast) */
	.bg-vignette {
		background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.7) 100%);
	}

	.bg-amber-scanlines {
		background: linear-gradient(
			rgba(18, 16, 16, 0) 50%,
			rgba(0, 0, 0, 0.1) 50%
		),
		linear-gradient(
			90deg,
			rgba(255, 0, 0, 0.02),
			rgba(0, 255, 0, 0.01),
			rgba(0, 0, 255, 0.02)
		);
		background-size: 100% 4px, 4px 100%;
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.9; }
		55% { opacity: 0.95; }
		60% { opacity: 0.85; }
	}
	.animate-flicker {
		animation: flicker 0.3s infinite;
	}

	@keyframes warmup {
		0% { opacity: 1; filter: contrast(2) brightness(0); }
		20% { opacity: 1; filter: contrast(1.5) brightness(1.2); }
		100% { opacity: 0; filter: contrast(1) brightness(1); visibility: hidden; }
	}
	.animate-warmup {
		animation: warmup 0.8s forwards ease-out;
	}

	.animate-spin-slow {
		animation: rotate 15s linear infinite;
	}
	@keyframes rotate {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	/* Industrial High-Res Scrollbar (Clean) */
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
		height: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: #0a0a0a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #222;
		border: 1px solid #111;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #f97316;
	}

	:global(body) {
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
		background-color: #0a0a0a;
	}

	input:focus {
		outline: none;
	}
</style>