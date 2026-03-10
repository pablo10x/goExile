<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { serverVersions } from '$lib/stores.svelte';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from './PlayersChart.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';

	let { nodeId, instance }: { nodeId: number; instance: any } = $props();

	let expanded = $state(false);
	let isHovered = $state(false);
	let renameValue = $state('');
	let chartData = $state<any[]>([]);

	$effect(() => {
		if (instance?.id) renameValue = instance.id;
	});

	const dispatch = createEventDispatcher();

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
			const res = await apiFetch(`/api/nodes/${nodeId}/instances/${instance.id}/stats/history`);
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
			dispatch('rename', { nodeId: nodeId, oldId: instance.id, newId: renameValue });
		}
	}
</script>

<div
	class="border border-white/5 rounded-2xl overflow-hidden mb-2 hover:border-sky-500/30 transition-all duration-300 shadow-md group/row relative bg-black/20 font-sans"
	onmouseenter={() => (isHovered = true)}
	onmouseleave={() => (isHovered = false)}
	role="region"
	aria-label={`Instance ${instance.id}`}
>
	<!-- Header / Collapsed View -->
	<div
		class="flex flex-col sm:flex-row sm:items-center gap-4 cursor-pointer hover:bg-white/[0.02] transition-all relative z-10"
	>
		<button
			class="flex flex-1 items-center gap-4 px-6 py-4 border-none bg-transparent text-left cursor-pointer"
			onclick={toggle}
			aria-expanded={expanded}
			aria-controls={`details-${instance.id}`}
		>
			<!-- Chevron -->
			<div
				class="text-slate-600 transform transition-transform duration-300 {expanded
					? 'rotate-90 text-sky-400'
					: ''}"
			>
				<Icon name="ph:caret-right-bold" size="1rem" />
			</div>

			<!-- Name & Identity -->
			<div class="flex flex-col min-w-[160px] ml-4 sm:ml-8">
				<div class="flex items-center gap-2">
					<span class="text-[9px] font-bold uppercase tracking-wider text-slate-500">Instance</span>
					<span class="font-bold text-sm text-white tracking-tight"
						>{instance.id.split('-').pop() || instance.port}</span
					>
				</div>
				<div class="flex items-center gap-2 mt-1">
					<span class="text-[8px] font-bold text-slate-600 uppercase">Port:</span>
					<span class="text-[9px] font-bold text-sky-500/70">{instance.port}</span>
				</div>
			</div>

			<!-- Status Badge -->
			<div class="sm:w-36">
				{#if instance.status === 'Running'}
					<div
						class="flex items-center gap-2.5 text-emerald-400 bg-emerald-500/5 px-3 py-1 border border-emerald-500/20 rounded-lg w-fit shadow-sm"
					>
						<div
							class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"
						></div>
						<span class="text-[10px] font-bold uppercase tracking-wide">Running</span>
					</div>
				{:else if instance.status === 'Provisioning'}
					<div
						class="flex items-center gap-2.5 text-amber-400 bg-amber-500/5 px-3 py-1 border border-amber-500/20 rounded-lg w-fit"
					>
						<div class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></div>
						<span class="text-[10px] font-bold uppercase tracking-wide">Starting</span>
					</div>
				{:else if instance.status === 'Error'}
					<div
						class="flex items-center gap-2.5 text-rose-400 bg-rose-500/5 px-3 py-1 border border-rose-500/20 rounded-lg w-fit"
					>
						<div
							class="w-1.5 h-1.5 rounded-full bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)]"
						></div>
						<span class="text-[10px] font-bold uppercase tracking-wide">Error</span>
					</div>
				{:else}
					<div
						class="flex items-center gap-2.5 text-slate-500 bg-black/40 px-3 py-1 border border-white/5 rounded-lg w-fit"
					>
						<div class="w-1.5 h-1.5 rounded-full bg-slate-700"></div>
						<span class="text-[10px] font-bold uppercase tracking-wide"
							>{instance.status || 'Offline'}</span
						>
					</div>
				{/if}
			</div>

			<!-- Version Info -->
			<div class="flex flex-col hidden md:flex">
				<span class="text-[8px] font-bold text-slate-600 uppercase tracking-widest font-sans">Build</span>
				<div class="flex items-center gap-2">
					<span class="text-[10px] font-medium text-slate-400 font-mono"
						>{instance.version || '0.0.0'}</span
					>
					{#if isOutdated}
						<span
							class="text-[7px] font-bold bg-sky-500/10 text-sky-400 px-1.5 py-0.5 border border-sky-500/20 rounded uppercase"
							>Update Available</span
						>
					{/if}
				</div>
			</div>

			<!-- Player Load -->
			<div class="flex flex-col hidden lg:flex ml-6">
				<span class="text-[8px] font-bold text-slate-600 uppercase tracking-widest font-sans"
					>Active Sessions</span
				>
				<div class="flex items-center gap-3">
					<div
						class="w-20 h-1 bg-black border border-white/5 rounded-full overflow-hidden relative p-0 shadow-inner"
					>
						<div
							class="h-full bg-sky-500 transition-all duration-1000 ease-out shadow-[0_0_8px_rgba(14,165,233,0.3)]"
							style="width: {Math.min(100, (instance.player_count / 100) * 100)}%"
						></div>
					</div>
					<span class="text-[10px] font-mono font-bold text-slate-400 tabular-nums"
						>{instance.player_count || 0}</span
					>
				</div>
			</div>
		</button>

		<!-- Quick Actions -->
		<div class="flex items-center gap-2 pr-6 ml-auto" role="toolbar" aria-label="Instance Actions">
			<Button
				onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
				variant="ghost"
				size="xs"
				icon="ph:terminal-window-bold"
				title="Console"
				class="!rounded-lg"
			/>

			<div class="w-px h-5 bg-white/5 mx-1"></div>

			{#if instance.status !== 'Running'}
				<Button
					onclick={() => dispatch('start', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:play-bold"
					title="Start"
					class="!text-emerald-400 hover:bg-emerald-500/5 !rounded-lg"
				/>
			{:else}
				<Button
					onclick={() => dispatch('stop', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:stop-bold"
					title="Stop"
					class="hover:!text-rose-400 hover:bg-rose-500/5 !rounded-lg"
				/>
				<Button
					onclick={() => dispatch('restart', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:arrows-clockwise-bold"
					title="Restart"
					class="hover:!text-sky-400 hover:bg-sky-500/5 !rounded-lg"
				/>
			{/if}
		</div>
	</div>

	<!-- Expanded Details -->
	{#if expanded}
		<div
			id={`details-${instance.id}`}
			transition:slide={{ duration: 300 }}
			class="bg-black/40 border-t border-white/5 p-8 space-y-8 relative z-10 shadow-inner"
		>
			<div class="grid grid-cols-1 xl:grid-cols-12 gap-10">
				<!-- Left: Technical Readouts -->
				<div class="xl:col-span-7 space-y-8">
					<!-- Primary Toolbar -->
					<div class="flex flex-wrap gap-3 pb-8 border-b border-white/5">
						<Button
							onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
							variant="secondary"
							size="sm"
							icon="ph:terminal-window-bold"
							class="!rounded-xl"
						>
							View Logs
						</Button>

						<!-- Build Controls -->
						<div class="flex gap-1 bg-black/40 p-1 border border-white/5 rounded-xl">
							<Button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								disabled={!isOutdated}
								variant={isOutdated ? 'primary' : 'secondary'}
								size="sm"
								icon="ph:arrow-up-to-line-bold"
								class="!rounded-lg"
							>
								Upgrade
							</Button>
							<Button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								variant="secondary"
								size="sm"
								icon="ph:arrow-down-to-line-bold"
								class="!rounded-lg"
							>
								Reinstall
							</Button>
						</div>

						<Button
							onclick={() => dispatch('delete', { nodeId, instanceId: instance.id })}
							disabled={instance.status === 'Running'}
							variant="danger"
							size="sm"
							icon="ph:trash-bold"
							class="!rounded-xl ml-auto"
						>
							Delete Instance
						</Button>
					</div>

					<!-- Load Chart -->
					<div class="space-y-4">
						<div class="flex justify-between items-end">
							<div class="flex flex-col gap-1">
								<span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest"
									>History</span
								>
								<h4 class="text-[11px] font-bold text-slate-300 uppercase tracking-wide">
									Active Sessions (24h)
								</h4>
							</div>
							<div
								class="text-[10px] font-bold text-sky-400 bg-sky-500/5 px-3 py-1 border border-sky-500/20 rounded-lg shadow-inner"
							>
								Current: {instance.player_count || 0} Sessions
							</div>
						</div>
						<div class="bg-black/40 border border-white/5 p-6 rounded-2xl shadow-inner">
							<PlayersChart data={chartData} height={160} color="#0ea5e9" />
						</div>
					</div>
				</div>

				<!-- Right: Configuration -->
				<div class="xl:col-span-5 space-y-6">
					<div class="bg-black/40 border border-white/5 p-6 space-y-6 rounded-2xl shadow-md">
						<div class="space-y-3">
							<label
								for={'name-' + instance.id}
								class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block ml-1"
								>Instance Name</label
							>
							<div class="flex gap-3">
								<input
									id={'name-' + instance.id}
									type="text"
									bind:value={renameValue}
									class="flex-1 bg-black border border-white/10 rounded-xl px-4 py-2.5 text-xs font-mono font-medium text-white focus:border-sky-500/50 outline-none transition-all shadow-inner"
									placeholder={instance.id}
								/>
								<Button
									onclick={handleRename}
									disabled={renameValue === instance.id || !renameValue.trim()}
									variant="secondary"
									size="sm"
									class="!rounded-xl"
								>
									Rename
								</Button>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4 pt-2">
							<div
								class="bg-black border border-white/5 p-4 flex flex-col gap-1 rounded-xl shadow-inner"
							>
								<span class="text-[9px] font-bold text-slate-600 uppercase tracking-wider font-sans"
									>Process ID</span
								>
								<span class="text-xs font-mono font-bold text-slate-300"
									>{instance.pid || 'N/A'}</span
								>
							</div>
							<div
								class="bg-black border border-white/5 p-4 flex flex-col gap-1 rounded-xl shadow-inner"
							>
								<span class="text-[9px] font-bold text-slate-600 uppercase tracking-wider font-sans"
									>Public Port</span
								>
								<span class="text-xs font-mono font-bold text-sky-400"
									>{instance.port || 'Auto'}</span
								>
							</div>
						</div>
					</div>

					<div
						class="bg-sky-500/5 border border-sky-500/10 p-5 flex gap-5 items-start rounded-2xl shadow-sm"
					>
						<div class="p-2.5 bg-sky-500/10 text-sky-400 border border-sky-500/20 rounded-xl">
							<Icon name="ph:activity-bold" size="1rem" />
						</div>
						<div class="space-y-2">
							<span class="text-[10px] font-bold text-sky-500 uppercase tracking-wider block font-sans"
								>Health Monitoring</span
							>
							<p class="text-[11px] font-medium text-slate-500 leading-relaxed font-sans">
								System monitoring is active. Health checks are performed automatically. Manual
								intervention is recommended if heartbeats are missed.
							</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
</style>
