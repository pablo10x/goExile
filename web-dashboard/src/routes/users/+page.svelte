<script lang="ts">
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Users, Search, RefreshCw, Circle } from 'lucide-svelte';

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

	let players = $state<Player[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');

	let filteredPlayers = $derived.by(() => {
		if (!searchQuery.trim()) return players;
		const query = searchQuery.toLowerCase();
		return players.filter(
			(p) =>
				p.name.toLowerCase().includes(query) ||
				p.uid.toLowerCase().includes(query) ||
				p.device_id.toLowerCase().includes(query)
		);
	});

	async function fetchPlayers() {
		loading = true;
		try {
			const res = await fetch('/api/game/players');
			if (res.ok) {
				players = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch players', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchPlayers();
		// Poll for online status updates every 10 seconds
		const interval = setInterval(fetchPlayers, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 p-6">
	<!-- Header -->
	<div class="mb-8 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
		<div>
			<h1 class="text-3xl font-bold text-white flex items-center gap-3">
				<Users class="w-8 h-8 text-blue-400" />
				Player Management
			</h1>
			<p class="text-slate-400 mt-1">Monitor and manage game players</p>
		</div>

		<div class="flex items-center gap-3 w-full md:w-auto">
			<div class="relative flex-1 md:w-64 group">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-blue-400 transition-colors" />
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search players..."
					class="w-full pl-10 pr-4 py-2 bg-slate-800/50 border border-slate-700 rounded-xl text-sm text-slate-200 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
				/>
			</div>
			<button
				onclick={fetchPlayers}
				disabled={loading}
				class="p-2 bg-slate-800 border border-slate-700 rounded-xl text-slate-400 hover:text-white hover:bg-slate-700 transition-all disabled:opacity-50"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Players Table -->
	<div class="bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl overflow-hidden">
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
					{#if loading && players.length === 0}
						<tr>
							<td colspan="5" class="px-6 py-8 text-center text-slate-500">
								<RefreshCw class="w-6 h-6 mx-auto mb-2 animate-spin opacity-50" />
								Loading players...
							</td>
						</tr>
					{:else if filteredPlayers.length === 0}
						<tr>
							<td colspan="5" class="px-6 py-8 text-center text-slate-500">
								No players found matching your search.
							</td>
						</tr>
					{:else}
						{#each filteredPlayers as player (player.id)}
							<tr class="hover:bg-slate-700/30 transition-colors">
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
										<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center font-bold text-white text-xs">
											{player.name.charAt(0).toUpperCase()}
										</div>
										<div>
											<div class="font-medium text-white">{player.name}</div>
											<div class="text-xs text-slate-500">ID: {player.id}</div>
										</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<div class="space-y-1">
										<div class="flex items-center gap-2 text-xs text-slate-400">
											<span class="w-16 font-mono text-slate-600 uppercase">UID</span>
											<code class="px-1.5 py-0.5 bg-slate-900 rounded text-slate-300 font-mono">{player.uid || 'N/A'}</code>
										</div>
										<div class="flex items-center gap-2 text-xs text-slate-400">
											<span class="w-16 font-mono text-slate-600 uppercase">Device</span>
											<code class="px-1.5 py-0.5 bg-slate-900 rounded text-slate-300 font-mono truncate max-w-[150px]" title={player.device_id}>{player.device_id}</code>
										</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<div class="flex flex-col gap-1">
										<span class="text-xs text-slate-300">
											<span class="text-amber-400 font-bold">{player.xp.toLocaleString()}</span> XP
										</span>
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-400">
									<div>{new Date(player.updated_at).toLocaleDateString()}</div>
									<div class="text-xs text-slate-600">{new Date(player.updated_at).toLocaleTimeString()}</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
