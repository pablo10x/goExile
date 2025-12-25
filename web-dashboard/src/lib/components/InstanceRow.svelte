<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { serverVersions, siteSettings } from '$lib/stores';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from './PlayersChart.svelte';
	import QuickActionsTooltip from './QuickActionsTooltip.svelte';
	import {
		ChevronRight,
		Settings,
		Play,
		Square,
		RotateCw,
		ArrowDownToLine,
		Trash2,
		TerminalSquare,
		Activity
	} from 'lucide-svelte';

	let { spawnerId, instance }: { spawnerId: number; instance: any } = $props();

	let expanded = $state(false);
	let renameValue = $state(instance.id);
	let chartData = $state<any[]>([]);

	$effect(() => {
		renameValue = instance.id;
	});

	const dispatch = createEventDispatcher();

	// Derived state for version checking
	let activeVersion = $derived($serverVersions.find((v) => v.is_active));
	let versionDiff = $derived(
		activeVersion && instance.version ? compareVersions(activeVersion.version, instance.version) : 0
	);
	let isOutdated = $derived(versionDiff > 0);

	function toggle() {
		expanded = !expanded;
		if (expanded) {
			renameValue = instance.id;
			fetchHistory();
		}
	}

	async function fetchHistory() {
		try {
			const res = await fetch(`/api/spawners/${spawnerId}/instances/${instance.id}/stats/history`);
			if (res.ok) {
				const data = await res.json();
				if (data.history) {
					chartData = data.history.map((h: any) => ({
						timestamp: new Date(h.timestamp).getTime(),
						count: h.player_count || 0
					}));
				}
			}
		} catch (e) {
			console.error('Failed to fetch history', e);
		}
	}

	function handleRename() {
		if (renameValue !== instance.id) {
			dispatch('rename', { spawnerId, oldId: instance.id, newId: renameValue });
		}
	}
</script>

<div
	class={`border border-stone-800 ${$siteSettings.aesthetic.industrial_styling ? 'rounded-none' : 'rounded-lg'} bg-stone-950/20 overflow-hidden mb-2 hover:border-rust/40 transition-all shadow-sm`}
>
	<!-- Header / Collapsed View -->
	<div
		class="flex flex-col sm:flex-row sm:items-center gap-4 px-5 py-3 cursor-pointer hover:bg-white/[0.02] transition-colors"
		onclick={toggle}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Enter' && toggle()}
	>
		<!-- Chevron -->
		<div
			class="text-stone-600 transform transition-transform duration-300 {expanded
				? 'rotate-90 text-rust-light'
				: ''}"
		>
			<ChevronRight class="w-4 h-4" />
		</div>

		<!-- Name & Identity -->
		<div class="flex flex-col min-w-[140px]">
			<div class="flex items-center gap-2">
				<span class="text-[10px] font-mono text-stone-600 uppercase tracking-tighter">Node_</span>
				<span class="font-bold text-sm text-white tracking-tight uppercase">{instance.id.split('-').pop() || instance.port}</span>
			</div>
			<div class="flex items-center gap-2 mt-0.5">
				<span class="text-[8px] font-mono text-stone-700 uppercase">Port:</span>
				<span class="text-[9px] font-mono text-rust-light/80">{instance.port}</span>
			</div>
		</div>

		<!-- Status Badge -->
		<div class="sm:w-32">
			{#if instance.status === 'Running'}
				<div class="flex items-center gap-2 text-emerald-500 bg-emerald-500/5 px-2 py-1 border border-emerald-500/20 w-fit">
					<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
					<span class="text-[9px] font-black uppercase tracking-widest">Running</span>
				</div>
			{:else if instance.status === 'Provisioning'}
				<div class="flex items-center gap-2 text-rust-light bg-rust/5 px-2 py-1 border border-rust/20 w-fit">
					<div class="w-1.5 h-1.5 rounded-none bg-rust animate-spin"></div>
					<span class="text-[9px] font-black uppercase tracking-widest">Syncing</span>
				</div>
			{:else if instance.status === 'Error'}
				<div class="flex items-center gap-2 text-red-500 bg-red-500/5 px-2 py-1 border border-red-500/20 w-fit">
					<div class="w-1.5 h-1.5 rounded-full bg-red-500 animate-flicker"></div>
					<span class="text-[9px] font-black uppercase tracking-widest">Fault</span>
				</div>
			{:else}
				<div class="flex items-center gap-2 text-stone-500 bg-stone-800/10 px-2 py-1 border border-stone-800/30 w-fit">
					<div class="w-1.5 h-1.5 rounded-full bg-stone-700"></div>
					<span class="text-[9px] font-black uppercase tracking-widest">{instance.status || 'Offline'}</span>
				</div>
			{/if}
		</div>

		<!-- Version Info -->
		<div class="flex flex-col hidden md:flex">
			<span class="text-[8px] font-mono text-stone-700 uppercase">System_Ver</span>
			<div class="flex items-center gap-2">
				<span class="text-[10px] font-mono text-stone-400">{instance.version || '0.0.0'}</span>
				{#if isOutdated}
					<span class="text-[7px] font-black bg-rust/20 text-rust px-1 border border-rust/30 uppercase animate-pulse">Update_Avail</span>
				{/if}
			</div>
		</div>

		<!-- Player Load -->
		<div class="flex flex-col hidden lg:flex ml-4">
			<span class="text-[8px] font-mono text-stone-700 uppercase">Active_Load</span>
			<div class="flex items-center gap-2">
				<div class="w-16 h-1.5 bg-stone-900 border border-stone-800 rounded-none overflow-hidden">
					<div class="h-full bg-rust-light opacity-60" style="width: {(instance.player_count / 100) * 100}%"></div>
				</div>
				<span class="text-[10px] font-mono text-stone-400">{instance.player_count || 0} Clients</span>
			</div>
		</div>

		<!-- Quick Actions -->
		<div
			class="flex items-center gap-1 ml-auto"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="toolbar"
			aria-label="Instance Quick Actions"
		>
			<button
				onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
				class="p-2 text-stone-600 hover:text-white hover:bg-stone-900 transition-all border border-transparent"
				title="Access Terminal"
			>
				<TerminalSquare class="w-4 h-4" />
			</button>

			<div class="w-px h-4 bg-stone-800 mx-1"></div>

			{#if instance.status !== 'Running'}
				<button
					onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}
					class="p-2 text-emerald-600 hover:text-emerald-400 hover:bg-emerald-500/5 transition-all"
					title="Execute Node"
				>
					<Play class="w-4 h-4" />
				</button>
			{:else}
				<button
					onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}
					class="p-2 text-stone-600 hover:text-yellow-500 hover:bg-yellow-500/5 transition-all"
					title="Terminate Node"
				>
					<Square class="w-4 h-4" />
				</button>
				<button
					onclick={() => dispatch('restart', { spawnerId, instanceId: instance.id })}
					class="p-2 text-stone-600 hover:text-rust-light hover:bg-rust/5 transition-all"
					title="Reboot Node"
				>
					<RotateCw class="w-4 h-4" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Expanded Details -->
	{#if expanded}
		<div
			transition:slide={{ duration: 300 }}
			class="bg-black/20 border-t border-stone-800/50 p-6 space-y-6"
		>
			<div class="grid grid-cols-1 xl:grid-cols-12 gap-8">
				<!-- Left: Technical Readouts -->
				<div class="xl:col-span-7 space-y-6">
					<!-- Primary Toolbar -->
					<div class="flex flex-wrap gap-2 pb-6 border-b border-stone-800/50">
						<button
							onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
							class="btn-industrial bg-stone-900 text-stone-400 hover:text-white border border-stone-800 hover:border-stone-600"
						>
							<TerminalSquare class="w-3 h-3" />
							<span>Terminal_Bridge</span>
						</button>
						<button
							onclick={() => dispatch('update', { spawnerId, instanceId: instance.id })}
							disabled={!isOutdated}
							class="btn-industrial {isOutdated ? 'bg-rust/10 text-rust-light border-rust/30 hover:bg-rust hover:text-white' : 'bg-stone-900/50 text-stone-600 border-stone-800 opacity-50 cursor-not-allowed'}"
						>
							<ArrowDownToLine class="w-3 h-3" />
							<span>{isOutdated ? 'Sync_Patch' : 'Registry_Up_to_Date'}</span>
						</button>
						<button
							onclick={() => dispatch('delete', { spawnerId, instanceId: instance.id })}
							disabled={instance.status === 'Running'}
							class="btn-industrial bg-red-950/10 text-red-600/70 hover:text-red-500 border border-red-900/20 hover:border-red-600/40 ml-auto"
						>
							<Trash2 class="w-3 h-3" />
							<span>Decommission_Record</span>
						</button>
					</div>

					<!-- Load Chart -->
					<div class="space-y-3">
						<div class="flex justify-between items-end">
							<div class="flex flex-col">
								<span class="text-[8px] font-mono text-stone-600 uppercase tracking-widest">Stream_Buffer</span>
								<h4 class="text-[10px] font-bold text-stone-300 uppercase tracking-wider">Client_Interaction_Telemetry</h4>
							</div>
							<div class="text-[10px] font-mono text-rust-light bg-rust/5 px-2 py-0.5 border border-rust/20">
								SIGNAL: {instance.player_count || 0}_ACTIVE
							</div>
						</div>
						<div class="bg-stone-950/40 border border-stone-800/50 p-4">
							<PlayersChart data={chartData} height={140} color="var(--color-rust)" />
						</div>
					</div>
				</div>

				<!-- Right: Node Configuration -->
				<div class="xl:col-span-5 space-y-4">
					<div class="bg-stone-900/30 border border-stone-800/50 p-5 space-y-4">
						<div class="space-y-2">
							<label for={'name-' + instance.id} class="text-[9px] font-mono text-stone-600 uppercase tracking-widest block">Instance_Identifier</label>
							<div class="flex gap-2">
								<input
									id={'name-' + instance.id}
									type="text"
									bind:value={renameValue}
									class="flex-1 bg-black border border-stone-800 px-4 py-2 text-xs font-mono text-white focus:border-rust/50 outline-none transition-all"
									placeholder={instance.id}
								/>
								<button
									onclick={handleRename}
									disabled={renameValue === instance.id || !renameValue.trim()}
									class="bg-stone-800 hover:bg-rust text-stone-400 hover:text-white px-4 py-2 text-[9px] font-black uppercase transition-all disabled:opacity-30 border border-stone-700 hover:border-rust"
								>
									Commit
								</button>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4 pt-2">
							<div class="bg-black/40 border border-stone-800/50 p-3 flex flex-col">
								<span class="text-[8px] font-mono text-stone-700 uppercase mb-1">Runtime_PID</span>
								<span class="text-xs font-mono text-stone-300 uppercase">{instance.pid || 'NULL'}</span>
							</div>
							<div class="bg-black/40 border border-stone-800/50 p-3 flex flex-col">
								<span class="text-[8px] font-mono text-stone-700 uppercase mb-1">Network_Port</span>
								<span class="text-xs font-mono text-rust-light uppercase">{instance.port || 'AUTO'}</span>
							</div>
						</div>
					</div>

					<div class="bg-amber-950/5 border border-amber-900/20 p-4 flex gap-4 items-start">
						<div class="p-2 bg-amber-900/20 text-amber-500">
							<Activity class="w-4 h-4" />
						</div>
						<div class="space-y-1">
							<span class="text-[9px] font-black text-amber-600 uppercase tracking-widest block">Maintenance_Protocol</span>
							<p class="text-[10px] font-mono text-stone-500 leading-tight">System ensures node stability through real-time heartbeat monitoring. Manual restarts recommended only for fatal buffer overflows.</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	@reference "../../app.css";
	
	.btn-industrial {
		@apply flex items-center gap-2 px-4 py-2 transition-all font-mono text-[9px] font-black uppercase;
	}

	.btn-toolbar {
		@apply gap-2 px-4 py-2 rounded-none text-[10px] font-black uppercase flex items-center transition-all active:translate-x-[1px] active:translate-y-[1px];
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
	.animate-flicker {
		animation: flicker 0.2s infinite;
	}
</style>