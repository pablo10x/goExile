<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { serverVersions } from '$lib/stores';
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
		TerminalSquare
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
	class="border border-slate-300/50 dark:border-slate-700/50 rounded-lg bg-slate-800/30 overflow-hidden mb-2"
>
	<!-- Header / Collapsed View -->
	<div
		class="flex flex-col sm:flex-row sm:items-center gap-3 px-3 sm:px-4 py-3 cursor-pointer hover:bg-slate-700/30 transition-colors"
		onclick={toggle}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Enter' && toggle()}
	>
		<!-- Top Row (Mobile): Chevron + Name -->
		<div class="flex items-center gap-2 w-full sm:w-auto">
			<div
				class="text-slate-500 transform transition-transform duration-200 {expanded
					? 'rotate-90'
					: ''}"
			>
				<ChevronRight class="w-4 h-4 sm:w-5 sm:h-5" />
			</div>

			<div
				class="font-mono text-sm sm:text-base text-slate-700 dark:text-slate-300 truncate flex-1 sm:flex-initial"
				title={`Exile Gameserver : #${instance.port}`}
			>
				<span class="hidden md:inline text-slate-500">Exile Gameserver : </span>#{instance.port}
			</div>
		</div>

		<!-- Details Grid/Flex -->
		<div
			class="flex-1 flex flex-wrap sm:grid sm:grid-cols-9 gap-2 sm:gap-4 items-center text-xs sm:text-sm pl-6 sm:pl-0"
		>
			<!-- Version -->
			<div class="col-span-3 text-slate-500 dark:text-slate-400 truncate flex items-center gap-2">
				<span class="sm:hidden text-slate-500">Ver:</span>
				{instance.version || 'Unknown'}
				{#if isOutdated}
					<span
						class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-yellow-500/20 text-yellow-400 border border-yellow-500/30 animate-pulse"
					>
						UPDATE
					</span>
				{/if}
			</div>

			<!-- Port -->
			<div class="col-span-2 font-mono text-slate-500 dark:text-slate-400">
				<span class="sm:hidden text-slate-500">Port:</span>
				{instance.port}
			</div>

			<!-- Status -->
			<div class="col-span-3">
				{#if instance.status === 'Running'}
					<span
						class="inline-flex items-center px-2 py-0.5 rounded text-[10px] sm:text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
					>
						Running
					</span>
				{:else if instance.status === 'Provisioning'}
					<span
						class="inline-flex items-center px-2 py-0.5 rounded text-[10px] sm:text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 animate-pulse"
					>
						Provisioning
					</span>
				{:else if instance.status === 'Error'}
					<span
						class="inline-flex items-center px-2 py-0.5 rounded text-[10px] sm:text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20"
					>
						Error
					</span>
				{:else}
					<span
						class="inline-flex items-center px-2 py-0.5 rounded text-[10px] sm:text-xs font-medium bg-slate-700 text-slate-500 dark:text-slate-400"
					>
						{instance.status}
					</span>
				{/if}
			</div>

			<!-- PID -->
			<div class="col-span-1 font-mono text-slate-500 text-[10px] sm:text-xs hidden sm:block">
				{instance.pid || '-'}
			</div>
		</div>

		<!-- Quick Actions (Right aligned on desktop, bottom row on mobile) -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="flex items-center gap-1 ml-auto pl-0 sm:pl-4 border-l-0 sm:border-l border-slate-300/50 dark:border-slate-700/50 w-full sm:w-auto justify-end mt-2 sm:mt-0"
			role="group"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<QuickActionsTooltip
				placement="bottom"
				title="Quick Actions"
				actions={[
					{
						label: 'Start',
						icon: Play,
						onClick: () => dispatch('start', { spawnerId, instanceId: instance.id }),
						disabled: instance.status === 'Running',
						variant: 'success'
					},
					{
						label: 'Stop',
						icon: Square,
						onClick: () => dispatch('stop', { spawnerId, instanceId: instance.id }),
						disabled: instance.status !== 'Running',
						variant: 'warning'
					},
					{
						label: 'Restart',
						icon: RotateCw,
						onClick: () => dispatch('restart', { spawnerId, instanceId: instance.id }),
						disabled: instance.status !== 'Running'
					}
				]}
			>
				<button
					onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
					class="p-1.5 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 rounded transition-colors"
					title="Manage"
				>
					<Settings class="w-4 h-4" />
				</button>
			</QuickActionsTooltip>

			<button
				onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}
				disabled={instance.status === 'Running'}
				class="p-1.5 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
				title="Start"
			>
				<Play class="w-4 h-4" />
			</button>

			<button
				onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}
				disabled={instance.status !== 'Running'}
				class="p-1.5 text-yellow-400 hover:text-yellow-300 hover:bg-yellow-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
				title="Stop"
			>
				<Square class="w-4 h-4" />
			</button>

			<button
				onclick={() => dispatch('restart', { spawnerId, instanceId: instance.id })}
				disabled={instance.status !== 'Running'}
				class="p-1.5 text-blue-400 hover:text-blue-300 hover:bg-blue-400/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
				title="Restart"
			>
				<RotateCw class="w-4 h-4" />
			</button>
		</div>
	</div>

	<!-- Expanded Details -->
	{#if expanded}
		<div
			transition:slide={{ duration: 200 }}
			class="bg-slate-900/50 border-t border-slate-300/50 dark:border-slate-700/50 p-3 sm:p-4"
		>
			<!-- Toolbar -->
			<div
				class="flex flex-wrap gap-2 mb-4 sm:mb-6 pb-3 sm:pb-4 border-b border-slate-300/50 dark:border-slate-700/50"
			>
				<button
					onclick={() => dispatch('tail', { spawnerId, instanceId: instance.id })}
					class="btn-toolbar bg-slate-700 hover:bg-slate-600 text-slate-800 dark:text-slate-200"
				>
					<Settings class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
					Manage
				</button>

				<button
					onclick={() => dispatch('start', { spawnerId, instanceId: instance.id })}
					disabled={instance.status === 'Running'}
					class="btn-toolbar bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-600/30 disabled:opacity-50"
				>
					Start
				</button>

				<button
					onclick={() => dispatch('stop', { spawnerId, instanceId: instance.id })}
					disabled={instance.status !== 'Running'}
					class="btn-toolbar bg-yellow-600/20 hover:bg-yellow-600/30 text-yellow-400 border border-yellow-600/30 disabled:opacity-50"
				>
					Stop
				</button>

				<button
					onclick={() => dispatch('update', { spawnerId, instanceId: instance.id })}
					disabled={versionDiff === 0}
					class={`btn-toolbar disabled:opacity-50 disabled:cursor-not-allowed ${versionDiff > 0 ? 'bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-600/30' : 'bg-orange-600/20 hover:bg-orange-600/30 text-orange-400 border border-orange-600/30'}`}
					title={versionDiff > 0
						? `Update to ${activeVersion?.version}`
						: versionDiff < 0
							? `Downgrade to ${activeVersion?.version}`
							: 'Server is up to date'}
				>
					{versionDiff > 0 ? 'Update' : 'Downgrade'}
				</button>

				<button
					onclick={() => dispatch('delete', { spawnerId, instanceId: instance.id })}
					disabled={instance.status === 'Running'}
					class="btn-toolbar bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-600/30 disabled:opacity-50 ml-auto"
				>
					Delete
				</button>
			</div>

			<!-- Stats Chart -->
			<div
				class="mb-4 sm:mb-6 bg-white/30 dark:bg-slate-950/30 rounded-lg border border-slate-300/50 dark:border-slate-700/50 p-3 sm:p-4"
			>
				<div class="flex justify-between items-end mb-2">
					<h4
						class="text-[10px] sm:text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
					>
						Player Activity (24h)
					</h4>
					<div class="text-xs sm:text-sm font-mono text-blue-400">
						{instance.player_count || 0} active
					</div>
				</div>
				<PlayersChart data={chartData} height={100} color="#3b82f6" />
			</div>

			<!-- Settings Form -->
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
				<div>
					<label
						for={'name-' + instance.id}
						class="block text-[10px] sm:text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2"
					>
						Game Server Name (ID)
					</label>
					<div class="flex gap-2">
						<input
							id={'name-' + instance.id}
							type="text"
							bind:value={renameValue}
							class="flex-1 px-3 py-1.5 bg-slate-900/50 border border-slate-600 rounded text-xs sm:text-sm text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors placeholder-slate-600"
							placeholder={instance.id}
						/>
						<button
							onclick={handleRename}
							disabled={renameValue === instance.id || !renameValue.trim()}
							class="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-slate-900 dark:text-white rounded text-xs font-bold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Rename
						</button>
					</div>
				</div>

				<div>
					<label
						for={'port-' + instance.id}
						class="block text-[10px] sm:text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2"
					>
						Assigned Port
					</label>
					<input
						id={'port-' + instance.id}
						type="text"
						value={instance.port || 'Auto'}
						disabled
						class="w-full px-3 py-1.5 bg-slate-900/30 border border-slate-300/50 dark:border-slate-700/50 rounded text-xs sm:text-sm text-slate-500 dark:text-slate-400 cursor-not-allowed"
					/>
				</div>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	.btn-toolbar {
		@apply gap-1.5 sm:gap-2 px-2.5 sm:px-3 py-1.5 rounded sm:text-xs font-semibold flex items-center text-[10px] transition-all;
	}
</style>
