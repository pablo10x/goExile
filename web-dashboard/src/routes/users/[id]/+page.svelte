<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		Users,
		ArrowLeft,
		RefreshCw,
		User,
		Smartphone,
		Calendar,
		Trophy,
		Globe,
		UserCheck,
		UserPlus,
		Clock,
		AlertOctagon,
		Lock,
		ChevronRight
	} from 'lucide-svelte';
	import { fade, slide } from 'svelte/transition';

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
		friends: Player[];
		incoming_friend_requests: Player[];
		outgoing_friend_requests: Player[];
	}

	let player = $state<Player | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const playerId = page.params.id;

	async function fetchPlayerDetails() {
		loading = true;
		error = null;
		try {
			const res = await fetch(`/api/admin/players/${playerId}`);
			if (res.ok) {
				player = await res.json();
			} else {
				error = 'Player not found or server error.';
			}
		} catch (e: any) {
			console.error('Failed to fetch player details', e);
			error = e.message || 'Failed to fetch player details.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchPlayerDetails();
	});
</script>

<div class="w-full space-y-10 pb-32 md:pb-12">
	<!-- Header -->
	<div class="mb-10 flex flex-col md:flex-row md:items-center justify-between gap-8">
		<div class="flex items-center gap-6">
			<button
				onclick={() => goto('/users')}
				class="p-4 bg-stone-900 border border-stone-800 text-stone-500 hover:text-white hover:border-rust transition-all industrial-frame shadow-xl group"
			>
				<ArrowLeft class="w-6 h-6 group-hover:-tranneutral-x-1 transition-transform" />
			</button>
			<div>
				<div class="flex items-center gap-3 mb-1">
					<div class="h-0.5 w-8 bg-rust"></div>
					<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-[0.3em]">Identity_Vault // Deep_Scan</span>
				</div>
				<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter flex items-center gap-4">
					Subject_<span class="text-rust">{player?.name || 'NULL'}</span>
				</h1>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-black mt-2">
					Detailed analysis of subject profile and synchronization graph
				</p>
			</div>
		</div>

		<button
			onclick={fetchPlayerDetails}
			disabled={loading}
			class="p-4 bg-stone-900 border border-stone-800 hover:border-rust text-stone-500 hover:text-rust transition-all industrial-frame shadow-xl active:tranneutral-y-px disabled:opacity-20"
		>
			<RefreshCw class="w-6 h-6 {loading ? 'animate-spin' : ''}" />
		</button>
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-40 gap-6">
			<div class="w-16 h-16 border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"></div>
			<p class="font-heading font-black text-[12px] text-rust animate-pulse uppercase tracking-[0.5em]">Synchronizing_Identity_Buffer...</p>
		</div>
	{:else if error}
		<div
			class="p-12 text-center bg-red-950/10 border border-red-900/30 industrial-frame shadow-2xl"
		>
			<AlertOctagon class="w-16 h-16 text-red-600 mx-auto mb-6 animate-pulse" />
			<h3 class="text-2xl font-heading font-black text-red-500 mb-3 uppercase tracking-widest">Identity_Extraction_Fault</h3>
			<p class="font-jetbrains text-stone-500 font-bold uppercase tracking-tight">{error}</p>
			<button class="mt-10 px-10 py-3 bg-red-600 hover:bg-red-500 text-white font-heading font-black text-[11px] uppercase tracking-widest transition-all shadow-lg" onclick={fetchPlayerDetails}>Retry_Protocol</button>
		</div>
	{:else if player}
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-10" transition:fade>
			<!-- Main Profile Card -->
			<div class="lg:col-span-8 space-y-10">
				<!-- Identity Card -->
				<div
					class="modern-industrial-card glass-panel p-10 relative overflow-hidden shadow-2xl !rounded-none"
				>
					<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
					
					<div class="absolute top-0 right-0 p-10 opacity-[0.03] pointer-events-none">
						<User class="w-64 h-64" />
					</div>

					<div class="relative z-10 flex flex-col sm:flex-row gap-10 items-center sm:items-start text-center sm:text-left">
						<div
							class="w-28 h-28 bg-stone-950 border-2 border-rust flex items-center justify-center text-5xl font-heading font-black italic text-white shadow-2xl shadow-rust/20 industrial-frame"
						>
							{player.name.charAt(0).toUpperCase()}
						</div>

						<div class="flex-1 space-y-6">
							<div>
								<div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-3">
									<h2 class="text-4xl font-heading font-black text-white uppercase tracking-tighter">{player.name}</h2>
									{#if player.online}
										<div
											class="w-fit px-4 py-1.5 bg-emerald-500/5 border border-emerald-500/20 text-emerald-400 text-[10px] font-black font-jetbrains rounded-none flex items-center gap-2.5 uppercase tracking-widest shadow-inner"
										>
											<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_8px_#10b981]"></div>
											SIGNAL_LOCKED
										</div>
									{:else}
										<div
											class="w-fit px-4 py-1.5 bg-stone-900 border border-stone-800 text-stone-600 text-[10px] font-black font-jetbrains rounded-none uppercase tracking-widest"
										>
											SIGNAL_LOST
										</div>
									{/if}
								</div>
								<div class="text-stone-500 font-jetbrains font-black text-[11px] uppercase tracking-[0.4em] italic">
									REGISTRY_SIGNATURE: 0x{player.id.toString(16).toUpperCase()}
								</div>
							</div>

							<div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
								<div
									class="p-5 bg-stone-950/60 border border-stone-800 industrial-frame shadow-inner"
								>
									<div
										class="text-[9px] text-stone-600 uppercase font-black tracking-[0.2em] mb-3 flex items-center gap-3"
									>
										<Smartphone class="w-3.5 h-3.5 text-rust" /> Terminal_ID
									</div>
									<code class="text-xs text-stone-300 font-jetbrains font-bold break-all opacity-80"
										>{player.device_id}</code
									>
								</div>
								<div
									class="p-5 bg-stone-950/60 border border-stone-800 industrial-frame shadow-inner"
								>
									<div
										class="text-[9px] text-stone-600 uppercase font-black tracking-[0.2em] mb-3 flex items-center gap-3"
									>
										<Lock class="w-3.5 h-3.5 text-rust" /> Registry_UID
									</div>
									<code class="text-xs text-stone-300 font-jetbrains font-bold break-all opacity-80"
										>{player.uid || 'NULL_PTR'}</code
									>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Stats Grid -->
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
					<div
						class="modern-industrial-card glass-panel p-8 flex flex-col items-center justify-center text-center group hover:border-rust/30 transition-all !rounded-none"
					>
						<div class="mb-4 p-3 bg-amber-500/5 border border-amber-500/20 industrial-frame text-amber-500 group-hover:bg-amber-500 group-hover:text-black transition-all">
							<Trophy class="w-6 h-6" />
						</div>
						<div class="text-3xl font-heading font-black text-white tabular-nums tracking-tighter mb-1">
							{player.xp.toLocaleString()}
						</div>
						<div
							class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-[0.3em]"
						>
							XP_ACCUMULATION
						</div>
					</div>

					<div
						class="modern-industrial-card glass-panel p-8 flex flex-col items-center justify-center text-center group hover:border-rust/30 transition-all !rounded-none"
					>
						<div class="mb-4 p-3 bg-rust/5 border border-rust/20 industrial-frame text-rust group-hover:bg-rust group-hover:text-white transition-all">
							<Globe class="w-6 h-6" />
						</div>
						<div
							class="text-sm font-heading font-black text-white uppercase tracking-widest truncate w-full mb-1"
							title={player.last_joined_server}
						>
							{player.last_joined_server || 'VOID_MAP'}
						</div>
						<div
							class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-[0.3em]"
						>
							LAST_KNOWN_VECTOR
						</div>
					</div>

					<div
						class="modern-industrial-card glass-panel p-8 flex flex-col items-center justify-center text-center group hover:border-rust/30 transition-all !rounded-none"
					>
						<div class="mb-4 p-3 bg-stone-800 border border-stone-700 industrial-frame text-stone-400 group-hover:bg-white group-hover:text-black transition-all">
							<Calendar class="w-6 h-6" />
						</div>
						<div class="text-xl font-heading font-black text-white uppercase tracking-tighter mb-1">
							{new Date(player.created_at).toLocaleDateString()}
						</div>
						<div
							class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-[0.3em]"
						>
							UPLINK_INITIATED
						</div>
					</div>
				</div>

				<!-- Synchronization Audit -->
				<div
					class="p-6 bg-stone-950/40 border border-stone-800 industrial-frame flex flex-col sm:flex-row sm:items-center justify-between gap-6 shadow-inner"
				>
					<div class="flex items-center gap-4 text-stone-600">
						<Clock class="w-5 h-5 text-rust" />
						<span class="font-jetbrains text-[10px] font-black uppercase tracking-widest">Last_Identity_Synchronization:</span>
					</div>
					<div class="font-jetbrains font-black text-xs text-stone-400 uppercase tracking-tighter flex items-center gap-4">
						<span class="text-white">{new Date(player.updated_at).toLocaleDateString()}</span>
						<span class="w-px h-3 bg-stone-800"></span>
						<span class="text-rust-light">{new Date(player.updated_at).toLocaleTimeString([], { hour12: false })}</span>
					</div>
				</div>
			</div>

			<!-- Sidebar / Social Graph -->
			<div class="lg:col-span-4 space-y-10">
				<!-- Friends List -->
				<div
					class="modern-industrial-card glass-panel flex flex-col max-h-[600px] !rounded-none shadow-2xl"
				>
					<div
						class="p-6 border-b border-stone-800 bg-[#0a0a0a] flex items-center justify-between"
					>
						<h3 class="font-heading font-black text-xs text-white flex items-center gap-4 uppercase tracking-widest">
							<UserCheck class="w-5 h-5 text-emerald-500" />
							Synchronization_Grid
						</h3>
						<span
							class="px-3 py-1 bg-stone-900 border border-stone-800 text-[10px] font-black font-jetbrains text-rust tabular-nums shadow-inner"
						>
							{player.friends?.length || 0}
						</span>
					</div>

					<div class="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar bg-black/20">
						{#if player.friends && player.friends.length > 0}
							{#each player.friends as friend}
								<a
									href={`/users/${friend.id}`}
									class="flex items-center gap-5 p-4 bg-stone-900/40 border border-stone-800 hover:border-rust/40 transition-all group industrial-frame"
								>
									<div
										class="w-10 h-10 bg-stone-950 border border-stone-800 flex items-center justify-center font-heading font-black italic text-stone-500 group-hover:text-rust transition-colors"
									>
										{friend.name.charAt(0).toUpperCase()}
									</div>
									<div class="flex-1 min-w-0">
										<div
											class="text-xs font-black text-stone-200 uppercase tracking-widest truncate group-hover:text-white transition-colors"
										>
											{friend.name}
										</div>
										<div class="text-[9px] text-stone-600 font-jetbrains font-bold uppercase tracking-widest mt-1 italic">SIG: 0x{friend.id.toString(16).toUpperCase()}</div>
									</div>
									<ChevronRight class="w-4 h-4 text-stone-800 group-hover:text-rust group-hover:tranneutral-x-1 transition-all" />
								</a>
							{/each}
						{:else}
							<div class="py-20 text-center opacity-30">
								<Users class="w-12 h-12 text-stone-800 mx-auto mb-4" />
								<p class="text-stone-600 font-jetbrains font-black text-[10px] uppercase tracking-widest">No_Connections_Mapped</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- Pending Requests -->
				{#if (player.incoming_friend_requests && player.incoming_friend_requests.length > 0) || (player.outgoing_friend_requests && player.outgoing_friend_requests.length > 0)}
					<div
						class="modern-industrial-card glass-panel !rounded-none shadow-2xl"
					>
						<div class="p-6 border-b border-stone-800 bg-[#0a0a0a]">
							<h3 class="font-heading font-black text-xs text-white flex items-center gap-4 uppercase tracking-widest">
								<UserPlus class="w-5 h-5 text-amber-500 animate-pulse" />
								Pending_Protocols
							</h3>
						</div>
						<div class="p-4 space-y-6 bg-black/20">
							{#if player.incoming_friend_requests && player.incoming_friend_requests.length > 0}
								<div class="space-y-3">
									<div class="px-2 text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] italic border-l border-stone-800">
										Incoming_Signals
									</div>
									{#each player.incoming_friend_requests as req}
										<div
											class="flex items-center gap-4 p-3 bg-stone-900/60 border border-stone-800 industrial-frame"
										>
											<div
												class="w-8 h-8 bg-stone-950 border border-stone-800 flex items-center justify-center font-heading font-black italic text-stone-600 text-xs"
											>
												{req.name.charAt(0).toUpperCase()}
											</div>
											<div class="min-w-0">
												<div class="text-[10px] font-black text-stone-300 uppercase tracking-widest truncate">
													{req.name}
												</div>
												<div class="text-[8px] text-stone-700 font-jetbrains font-bold uppercase tracking-tight">SRC: 0x{req.id.toString(16).toUpperCase()}</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}

							{#if player.outgoing_friend_requests && player.outgoing_friend_requests.length > 0}
								<div class="space-y-3">
									<div class="px-2 text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] italic border-l border-stone-800">
										Outgoing_Transmissions
									</div>
									{#each player.outgoing_friend_requests as req}
										<div
											class="flex items-center gap-4 p-3 bg-stone-900/20 border border-stone-800 opacity-60 industrial-frame"
										>
											<div
												class="w-8 h-8 bg-stone-950 border border-stone-800 flex items-center justify-center font-heading font-black italic text-stone-700 text-xs"
											>
												{req.name.charAt(0).toUpperCase()}
											</div>
											<div class="min-w-0">
												<div class="text-[10px] font-black text-stone-500 uppercase tracking-widest truncate">
													{req.name}
												</div>
												<div class="text-[8px] text-stone-800 font-jetbrains font-bold uppercase tracking-tight">DST: 0x{req.id.toString(16).toUpperCase()}</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
