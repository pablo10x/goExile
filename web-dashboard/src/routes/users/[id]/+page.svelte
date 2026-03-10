<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api';
	import {
		User,
		Activity,
		Shield,
		Clock,
		MapPin,
		Smartphone,
		Hash,
		Trophy,
		Calendar,
		ArrowLeft,
		MoreVertical,
		Ban,
		RefreshCw,
		CheckCircle,
		XCircle,
		Send,
		MessageSquare,
		ExternalLink
	} from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import { notifications } from '$lib/stores.svelte';
	import type { Player } from '$lib/stores.svelte';

	let player = $state<Player | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function fetchPlayer() {
		loading = true;
		try {
			const id = page.params.id;
			const res = await apiFetch(`/api/game/players/${id}`);
			if (res.ok) {
				player = await res.json();
			} else {
				error = 'User not found';
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function toggleBan() {
		if (!player) return;
		const newStatus = !player.banned;
		try {
			const res = await apiFetch(`/api/admin/players/${player.id}/ban`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ banned: newStatus })
			});
			if (res.ok) {
				const updated = await res.json();
				player = { ...player, ...updated };
				notifications.add({
					type: newStatus ? 'error' : 'success',
					message: newStatus ? 'User banned' : 'User restored'
				});
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Failed to update user status' });
		}
	}

	onMount(fetchPlayer);

	function formatDate(dateStr: string | undefined) {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleString();
	}
</script>

<div class="w-full h-full space-y-8 pb-32 md:pb-12 font-sans">
	<!-- Top Navigation -->
	<div class="flex items-center justify-between relative z-10">
		<Button variant="ghost" size="sm" href="/users" icon="ph:arrow-left-bold">
			Back to Users
		</Button>
		<div class="flex gap-3">
			<Button variant="secondary" size="sm" onclick={fetchPlayer} icon="ph:arrows-clockwise-bold" />
			<Button
				variant={player?.banned ? 'success' : 'danger'}
				size="sm"
				onclick={toggleBan}
				icon={player?.banned ? 'ph:check-circle-bold' : 'ph:prohibit-bold'}
			>
				{player?.banned ? 'Unban User' : 'Ban User'}
			</Button>
		</div>
	</div>

	{#if loading}
		<div class="flex flex-col items-center justify-center py-20 gap-4">
			<RefreshCw class="w-10 h-10 text-indigo-500 animate-spin" />
			<p class="text-xs font-bold text-slate-500 uppercase tracking-widest">Loading User Profile...</p>
		</div>
	{:else if error || !player}
		<div
			class="flex flex-col items-center justify-center py-20 bg-slate-900/40 border border-white/5 rounded-3xl"
		>
			<XCircle class="w-16 h-16 text-rose-500 mb-6" />
			<h2 class="text-2xl font-bold text-white mb-2">Error Loading Profile</h2>
			<p class="text-slate-400">{error || 'Unknown Error'}</p>
		</div>
	{:else}
		<!-- Profile Header -->
		<div class="relative group">
			<div
				class="p-8 bg-slate-800/40 border border-white/5 rounded-[2.5rem] shadow-2xl backdrop-blur-xl flex flex-col md:flex-row items-center gap-10 relative overflow-hidden"
			>
				<div
					class="absolute top-0 right-0 w-64 h-64 bg-indigo-500/5 rounded-full blur-[80px] -mr-32 -mt-32"
				></div>

				<!-- Avatar -->
				<div class="relative shrink-0">
					<div
						class="w-32 h-32 md:w-40 md:h-40 bg-slate-900 rounded-[2.5rem] border-2 border-white/5 flex items-center justify-center shadow-inner relative z-10 overflow-hidden"
					>
						{#if player.banned}
							<div class="absolute inset-0 bg-rose-500/20 backdrop-blur-[2px] z-20"></div>
						{/if}
						<User
							class="w-16 h-16 md:w-20 md:h-20 {player.banned ? 'text-rose-500' : 'text-indigo-400'}"
						/>
					</div>
					{#if player.online}
						<div
							class="absolute -bottom-2 -right-2 w-10 h-10 bg-emerald-500 border-4 border-slate-800 rounded-2xl flex items-center justify-center shadow-lg z-30"
						>
							<Activity class="w-5 h-5 text-white" />
						</div>
					{/if}
				</div>

				<div class="flex-1 text-center md:text-left relative z-10">
					<div class="flex flex-wrap items-center justify-center md:justify-start gap-4 mb-4">
						<h1 class="text-4xl md:text-5xl font-bold text-white tracking-tight">
							{player.name || 'Anonymous User'}
						</h1>
						{#if player.banned}
							<span
								class="px-4 py-1 bg-rose-500 text-white text-[10px] font-black uppercase tracking-widest rounded-full shadow-lg shadow-rose-500/20"
								>Banned</span
							>
						{:else if player.online}
							<span
								class="px-4 py-1 bg-emerald-500 text-white text-[10px] font-black uppercase tracking-widest rounded-full shadow-lg shadow-emerald-500/20"
								>Online</span
							>
						{/if}
					</div>
					<p class="text-sm text-slate-400 max-w-2xl leading-relaxed">
						User profile and system activity analysis for account 
						<span class="text-white font-mono">#{player.id}</span>
					</p>

					<div
						class="flex flex-wrap items-center justify-center md:justify-start gap-6 mt-8 text-xs font-bold text-slate-500 uppercase tracking-widest"
					>
						<div class="flex items-center gap-2">
							<Clock class="w-4 h-4" /> Joined: {formatDate(player.created_at)}
						</div>
						<div class="flex items-center gap-2">
							<Activity class="w-4 h-4" /> Last Active: {formatDate(player.updated_at)}
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
			<StatsCard
				title="Account Rank"
				value="LVL {Math.floor((player.xp || 0) / 1000) + 1}"
				iconName="ph:crown-bold"
				color="indigo"
				subValue="System Level"
			/>
			<StatsCard
				title="Total XP"
				value={player.xp?.toLocaleString() || '0'}
				iconName="ph:dna-bold"
				color="orange"
				subValue="Account Points"
			/>
			<StatsCard
				title="Session Time"
				value="12.4h"
				iconName="ph:timer-bold"
				color="sky"
				subValue="Total Activity"
			/>
			<StatsCard
				title="Device Count"
				value="1"
				iconName="ph:smartphone-bold"
				color="emerald"
				subValue="Linked Hardware"
			/>
		</div>

		<div class="grid grid-cols-1 xl:grid-cols-12 gap-8">
			<!-- Identity Details -->
			<div class="xl:col-span-5 space-y-8">
				<Card title="Account Metadata" subtitle="System identification data" icon="ph:identification-card-bold">
					<div class="p-8 space-y-8">
						<div class="grid grid-cols-1 gap-6">
							<div class="space-y-2">
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em]"
									>Account Identifier</span
								>
								<div
									class="flex items-center justify-between p-4 bg-black/40 border border-white/5 rounded-2xl group hover:border-indigo-500/20 transition-all"
								>
									<span class="text-sm font-mono text-indigo-400">ID_{player.id}</span>
									<div class="flex gap-2">
										<span class="text-[9px] font-bold text-slate-600 uppercase">Registered</span>
										<CheckCircle class="w-3 h-3 text-emerald-500" />
									</div>
								</div>
							</div>

							<div class="space-y-2">
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em]"
									>User UID</span
								>
								<div
									class="p-4 bg-black/40 border border-white/5 rounded-2xl font-mono text-xs text-slate-400 break-all"
								>
									{player.uid || 'NONE'}
								</div>
							</div>

							<div class="space-y-2">
								<span class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em]"
									>Device ID</span
								>
								<div
									class="p-4 bg-black/40 border border-white/5 rounded-2xl flex items-center gap-4"
								>
									<Smartphone class="w-5 h-5 text-slate-600" />
									<span class="text-sm font-mono text-slate-300"
										>{player.device_id || 'UNKNOWN'}</span
									>
								</div>
							</div>
						</div>
					</div>
				</Card>

				<Card title="Account History" subtitle="System access logs" icon="ph:clock-counter-clockwise-bold">
					<div class="p-0">
						<div class="divide-y divide-white/5">
							{#each [
								{ type: 'LOGIN', time: player.updated_at, loc: '127.0.0.1' },
								{ type: 'UPDATE', time: player.created_at, loc: 'ADMIN_CONSOLE' }
							] as log}
								<div class="p-6 flex items-center justify-between hover:bg-white/5 transition-all">
									<div class="flex items-center gap-4">
										<div class="p-2 bg-slate-900 rounded-lg">
											<Clock class="w-4 h-4 text-slate-500" />
										</div>
										<div>
											<div class="text-xs font-bold text-white uppercase tracking-wider">{log.type}</div>
											<div class="text-[10px] text-slate-500 font-mono mt-0.5">{log.loc}</div>
										</div>
									</div>
									<div class="text-[10px] font-bold text-slate-600">{formatDate(log.time)}</div>
								</div>
							{/each}
						</div>
					</div>
				</Card>
			</div>

			<!-- Activity Graph & More -->
			<div class="xl:col-span-7 space-y-8">
				<Card title="Activity Analysis" subtitle="Performance over time" icon="ph:chart-line-up-bold">
					<div class="p-8">
						<div
							class="h-80 w-full bg-black/40 border border-white/5 rounded-3xl flex items-center justify-center text-slate-700 font-bold uppercase tracking-[0.3em] text-[10px] relative overflow-hidden"
						>
							<Activity class="w-12 h-12 opacity-5 mb-4" />
							<div class="absolute inset-0 flex items-center justify-center">
								Activity visualization placeholder
							</div>
						</div>
					</div>
				</Card>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
					<Card title="Communications" subtitle="Recent messages" icon="ph:chat-circle-dots-bold">
						<div class="p-8 text-center py-12">
							<MessageSquare class="w-10 h-10 text-slate-800 mx-auto mb-4" />
							<p class="text-[10px] font-bold text-slate-600 uppercase tracking-widest">
								No message logs available
							</p>
						</div>
					</Card>
					<Card title="Linked Items" subtitle="Inventory data" icon="ph:package-bold">
						<div class="p-8 text-center py-12">
							<Send class="w-10 h-10 text-slate-800 mx-auto mb-4" />
							<p class="text-[10px] font-bold text-slate-600 uppercase tracking-widest">
								No inventory items found
							</p>
						</div>
					</Card>
				</div>
			</div>
		</div>
	{/if}
</div>
