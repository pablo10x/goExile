<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Users, ArrowLeft, RefreshCw, User, Smartphone, Calendar, Trophy, Globe, UserCheck, UserPlus, Clock } from 'lucide-svelte';
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
			const res = await fetch(`/api/game/players/${playerId}`);
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

<div class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
	<!-- Header -->
	<div class="mb-8 flex items-center justify-between">
		<div class="flex items-center gap-4">
			<button
				onclick={() => goto('/users')}
				class="p-2 bg-slate-800/50 border border-slate-700/50 rounded-xl text-slate-400 hover:text-white hover:bg-slate-700 transition-all group"
			>
				<ArrowLeft class="w-5 h-5 group-hover:-translate-x-0.5 transition-transform" />
			</button>
			<div>
				<h1 class="text-3xl font-bold text-white flex items-center gap-3">
					<User class="w-8 h-8 text-blue-400" />
					Player Details
				</h1>
				<p class="text-slate-400 mt-1">Detailed view of player profile and social graph</p>
			</div>
		</div>

		<button
			onclick={fetchPlayerDetails}
			disabled={loading}
			class="p-2 bg-slate-800 border border-slate-700 rounded-xl text-slate-400 hover:text-white hover:bg-slate-700 transition-all disabled:opacity-50"
		>
			<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
		</button>
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-20">
			<RefreshCw class="w-12 h-12 text-blue-500 animate-spin mb-4" />
			<p class="text-slate-400">Loading player profile...</p>
		</div>
	{:else if error}
		<div class="p-6 bg-red-500/10 border border-red-500/30 rounded-2xl flex items-center gap-4 text-red-400">
			<div class="p-3 bg-red-500/20 rounded-xl">
				<Users class="w-6 h-6" />
			</div>
			<div>
				<h3 class="font-bold text-lg">Error Loading Player</h3>
				<p>{error}</p>
			</div>
		</div>
	{:else if player}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6" transition:fade>
			<!-- Main Profile Card -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Identity Card -->
				<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 relative overflow-hidden">
					<div class="absolute top-0 right-0 p-6 opacity-10">
						<User class="w-32 h-32" />
					</div>
					
					<div class="relative z-10 flex flex-col sm:flex-row gap-6 items-start">
						<div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center text-3xl font-bold text-white shadow-xl shadow-blue-900/30">
							{player.name.charAt(0).toUpperCase()}
						</div>
						
						<div class="flex-1 space-y-4">
							<div>
								<div class="flex items-center gap-3">
									<h2 class="text-2xl font-bold text-white">{player.name}</h2>
									{#if player.online}
										<div class="px-2 py-0.5 bg-emerald-500/20 border border-emerald-500/30 text-emerald-400 text-xs font-bold rounded-full flex items-center gap-1.5">
											<div class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse"></div>
											ONLINE
										</div>
									{:else}
										<div class="px-2 py-0.5 bg-slate-700/50 border border-slate-600 text-slate-400 text-xs font-bold rounded-full">
											OFFLINE
										</div>
									{/if}
								</div>
								<div class="text-slate-400 font-mono text-sm mt-1">ID: #{player.id}</div>
							</div>

							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div class="p-3 bg-slate-900/50 rounded-xl border border-slate-700/30">
									<div class="text-xs text-slate-500 uppercase font-bold tracking-wider mb-1 flex items-center gap-2">
										<Smartphone class="w-3 h-3" /> Device ID
									</div>
									<code class="text-sm text-slate-300 font-mono break-all">{player.device_id}</code>
								</div>
								<div class="p-3 bg-slate-900/50 rounded-xl border border-slate-700/30">
									<div class="text-xs text-slate-500 uppercase font-bold tracking-wider mb-1 flex items-center gap-2">
										<Users class="w-3 h-3" /> Firebase UID
									</div>
									<code class="text-sm text-slate-300 font-mono break-all">{player.uid || 'Not Linked'}</code>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Stats Grid -->
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<div class="bg-slate-800/40 border border-slate-700/50 rounded-xl p-4 flex flex-col items-center justify-center text-center hover:bg-slate-800/60 transition-colors">
						<div class="mb-2 p-2 bg-amber-500/10 rounded-lg text-amber-400">
							<Trophy class="w-5 h-5" />
						</div>
						<div class="text-2xl font-bold text-white">{player.xp.toLocaleString()}</div>
						<div class="text-xs text-slate-400 uppercase font-bold tracking-wider">Experience</div>
					</div>
					
					<div class="bg-slate-800/40 border border-slate-700/50 rounded-xl p-4 flex flex-col items-center justify-center text-center hover:bg-slate-800/60 transition-colors">
						<div class="mb-2 p-2 bg-purple-500/10 rounded-lg text-purple-400">
							<Globe class="w-5 h-5" />
						</div>
						<div class="text-sm font-bold text-white truncate w-full" title={player.last_joined_server}>
							{player.last_joined_server || 'N/A'}
						</div>
						<div class="text-xs text-slate-400 uppercase font-bold tracking-wider">Last Server</div>
					</div>

					<div class="bg-slate-800/40 border border-slate-700/50 rounded-xl p-4 flex flex-col items-center justify-center text-center hover:bg-slate-800/60 transition-colors">
						<div class="mb-2 p-2 bg-blue-500/10 rounded-lg text-blue-400">
							<Calendar class="w-5 h-5" />
						</div>
						<div class="text-sm font-bold text-white">
							{new Date(player.created_at).toLocaleDateString()}
						</div>
						<div class="text-xs text-slate-400 uppercase font-bold tracking-wider">Joined</div>
					</div>
				</div>

				<!-- Timestamps -->
				<div class="bg-slate-800/40 border border-slate-700/50 rounded-xl p-4 flex items-center justify-between text-sm">
					<div class="flex items-center gap-2 text-slate-400">
						<Clock class="w-4 h-4" />
						<span>Last Updated:</span>
					</div>
					<div class="font-mono text-slate-300">
						{new Date(player.updated_at).toLocaleString()}
					</div>
				</div>
			</div>

			<!-- Sidebar / Social Graph -->
			<div class="space-y-6">
				<!-- Friends List -->
				<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden flex flex-col h-full max-h-[500px]">
					<div class="p-4 border-b border-slate-700/50 bg-slate-900/30 flex items-center justify-between">
						<h3 class="font-bold text-white flex items-center gap-2">
							<UserCheck class="w-5 h-5 text-emerald-400" />
							Friends
						</h3>
						<span class="px-2 py-0.5 bg-slate-700 rounded-full text-xs font-bold text-slate-300">
							{player.friends?.length || 0}
						</span>
					</div>
					
					<div class="flex-1 overflow-y-auto p-2 space-y-1">
						{#if player.friends && player.friends.length > 0}
							{#each player.friends as friend}
								<a href={`/users/${friend.id}`} class="flex items-center gap-3 p-2 hover:bg-slate-700/30 rounded-xl transition-colors group">
									<div class="w-8 h-8 rounded-lg bg-slate-700 flex items-center justify-center font-bold text-white text-xs">
										{friend.name.charAt(0).toUpperCase()}
									</div>
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-slate-200 truncate group-hover:text-blue-400 transition-colors">{friend.name}</div>
										<div class="text-xs text-slate-500 font-mono">#{friend.id}</div>
									</div>
								</a>
							{/each}
						{:else}
							<div class="p-8 text-center text-slate-500 text-sm">
								No friends found
							</div>
						{/if}
					</div>
				</div>

				<!-- Friend Requests -->
				{#if (player.incoming_friend_requests && player.incoming_friend_requests.length > 0) || (player.outgoing_friend_requests && player.outgoing_friend_requests.length > 0)}
					<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden">
						<div class="p-4 border-b border-slate-700/50 bg-slate-900/30">
							<h3 class="font-bold text-white flex items-center gap-2">
								<UserPlus class="w-5 h-5 text-amber-400" />
								Pending Requests
							</h3>
						</div>
						<div class="p-2 space-y-2">
							{#if player.incoming_friend_requests && player.incoming_friend_requests.length > 0}
								<div class="px-2 py-1 text-xs font-bold text-slate-500 uppercase tracking-wider">Incoming</div>
								{#each player.incoming_friend_requests as req}
									<div class="flex items-center gap-3 p-2 bg-slate-900/30 rounded-xl border border-slate-700/30">
										<div class="w-8 h-8 rounded-lg bg-slate-700 flex items-center justify-center font-bold text-white text-xs">
											{req.name.charAt(0).toUpperCase()}
										</div>
										<div>
											<div class="text-sm font-medium text-slate-200">{req.name}</div>
											<div class="text-xs text-slate-500">From #{req.id}</div>
										</div>
									</div>
								{/each}
							{/if}

							{#if player.outgoing_friend_requests && player.outgoing_friend_requests.length > 0}
								<div class="px-2 py-1 text-xs font-bold text-slate-500 uppercase tracking-wider mt-2">Outgoing</div>
								{#each player.outgoing_friend_requests as req}
									<div class="flex items-center gap-3 p-2 bg-slate-900/30 rounded-xl border border-slate-700/30 opacity-70">
										<div class="w-8 h-8 rounded-lg bg-slate-700 flex items-center justify-center font-bold text-white text-xs">
											{req.name.charAt(0).toUpperCase()}
										</div>
										<div>
											<div class="text-sm font-medium text-slate-200">{req.name}</div>
											<div class="text-xs text-slate-500">To #{req.id}</div>
										</div>
									</div>
								{/each}
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
