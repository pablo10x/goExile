<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { serverVersions, siteSettings } from '$lib/stores';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from './PlayersChart.svelte';
	import Icon from './theme/Icon.svelte';
	import {
		ChevronRight,
		Play,
		Square,
		RotateCw,
		ArrowDownToLine,
		ArrowUpToLine,
		Trash2,
		TerminalSquare,
		Activity,
		Clock
	} from 'lucide-svelte';

	let { nodeId, instance }: { nodeId: number; instance: any } = $props();

	let expanded = $state(false);
	let isHovered = $state(false);
	let renameValue = $state(instance.id);
	let chartData = $state<any[]>([]);

    // Static binary content for visual effect (no GC overhead)
    const binaryLines = Array.from({ length: 8 }, () => 
        Array.from({ length: 16 }, () => Math.random() > 0.5 ? '1' : '0').join('')
    );

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
			const res = await fetch(`/api/nodes/${nodeId}/instances/${instance.id}/stats/history`);
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
	class={`border border-stone-800 ${$siteSettings.aesthetic.industrial_styling ? 'rounded-none' : 'rounded-xl'} glass-panel overflow-hidden mb-2 hover:border-rust/40 transition-all duration-500 shadow-lg group/row relative`}
	style="background-color: rgba(12, 10, 9, var(--card-alpha));"
	class:heartbeat-pulse={instance.status === 'Running' && !$siteSettings.aesthetic.reduced_motion}
	onmouseenter={() => isHovered = true}
	onmouseleave={() => isHovered = false}
	role="region"
	aria-label={`Instance ${instance.id}`}
>
	<!-- Binary Animation Overlays -->
	{#if instance.status === 'Running' && !$siteSettings.aesthetic.reduced_motion && isHovered}
		<div class="absolute left-0 top-0 bottom-0 flex flex-col justify-center px-2 pointer-events-none overflow-hidden opacity-20">
            <div class="animate-binary-slide flex flex-col gap-0.5 will-change-transform">
                {#each [...binaryLines, ...binaryLines] as line}
                    <div class="font-mono text-[6px] text-emerald-500 whitespace-nowrap leading-tight mask-fade-right">
                        {line}
                    </div>
                {/each}
            </div>
		</div>
	{/if}

	<!-- Header / Collapsed View -->
	<div
		class="flex flex-col sm:flex-row sm:items-center gap-4 px-6 py-4 cursor-pointer hover:bg-rust/5 transition-all relative z-10"
		onclick={toggle}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Enter' && toggle()}
	>
		<!-- Chevron -->
		<div
			class="text-stone-600 transform transition-transform duration-500 {expanded
				? 'rotate-90 text-rust shadow-rust/20'
				: ''}"
		>
			<Icon name="ph:caret-right-bold" size="1rem" />
		</div>

		<!-- Name & Identity -->
		<div class="flex flex-col min-w-[160px] ml-4 sm:ml-8">
			<div class="flex items-center gap-2">
				                <span class="text-[9px] font-jetbrains font-black uppercase tracking-tighter" style="color: var(--text-dim)">INSTANCE</span>				<span class="font-heading font-black text-sm text-white tracking-widest uppercase">{instance.id.split('-').pop() || instance.port}</span>
			</div>
			<div class="flex items-center gap-2 mt-1">
				<span class="text-[8px] font-jetbrains font-bold text-stone-700 uppercase">Port:</span>
				<span class="text-[9px] font-jetbrains font-black text-rust-light tracking-widest">{instance.port}</span>
			</div>
		</div>

		<!-- Status Badge -->
		<div class="sm:w-36">
			{#if instance.status === 'Running'}
				<div class="flex items-center gap-2.5 text-success bg-success/5 px-3 py-1 border border-success/20 rounded-none w-fit shadow-[0_0_15px_rgba(16,185,129,0.05)]">
					<div class="w-1.5 h-1.5 rounded-full bg-success animate-pulse shadow-emerald-500/50 shadow-lg"></div>
					<span class="text-[9px] font-heading font-black uppercase tracking-[0.2em]">Running</span>
				</div>
			{:else if instance.status === 'Provisioning'}
				<div class="flex items-center gap-2.5 text-rust-light bg-rust/5 px-3 py-1 border border-rust/20 rounded-none w-fit">
					<div class="w-1.5 h-1.5 rounded-none bg-rust animate-spin shadow-rust/50 shadow-lg"></div>
					<span class="text-[9px] font-heading font-black uppercase tracking-[0.2em]">Starting</span>
				</div>
			{:else if instance.status === 'Error'}
				<div class="flex items-center gap-2.5 text-danger bg-danger/5 px-3 py-1 border border-danger/20 rounded-none w-fit">
					<div class="w-1.5 h-1.5 rounded-full bg-danger animate-flicker shadow-red-500/50 shadow-lg"></div>
					<span class="text-[9px] font-heading font-black uppercase tracking-[0.2em]">Error</span>
				</div>
			{:else}
				<div class="flex items-center gap-2.5 text-stone-500 bg-stone-900/40 px-3 py-1 border border-stone-800 rounded-none w-fit">
					<div class="w-1.5 h-1.5 rounded-full bg-stone-700"></div>
					<span class="text-[9px] font-heading font-black uppercase tracking-[0.2em]">{instance.status || 'Offline'}</span>
				</div>
			{/if}
		</div>

		<!-- Version Info -->
		<div class="flex flex-col hidden md:flex">
			<span class="text-[8px] font-jetbrains font-black text-stone-700 uppercase tracking-widest">Version</span>
			<div class="flex items-center gap-2">
				<span class="text-[10px] font-jetbrains font-bold text-stone-400">{instance.version || '0.0.0'}</span>
				{#if isOutdated}
					<span class="text-[7px] font-black font-heading bg-rust/20 text-rust px-1.5 py-0.5 border border-rust/30 uppercase animate-pulse shadow-rust/10 shadow-lg">UPDATE</span>
				{/if}
			</div>
		</div>

		<!-- Player Load -->
		<div class="flex flex-col hidden lg:flex ml-6">
			<span class="text-[8px] font-jetbrains font-black text-stone-700 uppercase tracking-widest">Players</span>
			<div class="flex items-center gap-3">
				<div class="w-20 h-1 bg-stone-950 border border-stone-800 rounded-none overflow-hidden relative p-0">
					<div class="h-full bg-rust-light shadow-rust/40 shadow-lg transition-all duration-1000 ease-out" style="width: {Math.min(100, (instance.player_count / 100) * 100)}%"></div>
				</div>
				<span class="text-[10px] font-jetbrains font-black text-stone-400 tracking-tighter">{instance.player_count || 0}</span>
			</div>
		</div>

		<!-- Quick Actions -->
		<div
			class="flex items-center gap-2 ml-auto"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="toolbar"
			aria-label="Instance Actions"
			tabindex="0"
		>
			<button
				onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
				class="p-2.5 text-stone-600 hover:text-white hover:bg-stone-900 transition-all border border-transparent hover:shadow-lg"
				title="Console"
			>
				<Icon name="ph:terminal-window-bold" size="1rem" />
			</button>

			<div class="w-px h-5 bg-stone-800 mx-1"></div>

			{#if instance.status !== 'Running'}
				<button
					onclick={() => dispatch('start', { nodeId, instanceId: instance.id })}
					class="p-2.5 text-emerald-600 hover:text-success hover:bg-emerald-500/5 transition-all hover:shadow-emerald-500/10 hover:shadow-lg"
					title="Start"
				>
					<Icon name="ph:play-bold" size="1rem" />
				</button>
			{:else}
				<button
					onclick={() => dispatch('stop', { nodeId, instanceId: instance.id })}
					class="p-2.5 text-stone-600 hover:text-yellow-500 hover:bg-yellow-500/5 transition-all hover:shadow-yellow-500/10 hover:shadow-lg"
					title="Stop"
				>
					<Icon name="ph:stop-bold" size="1rem" />
				</button>
				<button
					onclick={() => dispatch('restart', { nodeId, instanceId: instance.id })}
					class="p-2.5 text-stone-600 hover:text-rust-light hover:bg-rust/5 transition-all hover:shadow-rust/10 hover:shadow-lg"
					title="Restart"
				>
					<Icon name="ph:arrows-clockwise-bold" size="1rem" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Expanded Details -->
	{#if expanded}
		<div
			transition:slide={{ duration: 300 }}
			class="bg-[var(--terminal-bg)]/60 border-t border-stone-800 p-8 space-y-8 relative z-10"
		>
			<div class="grid grid-cols-1 xl:grid-cols-12 gap-10">
				<!-- Left: Technical Readouts -->
				<div class="xl:col-span-7 space-y-8">
					<!-- Primary Toolbar -->
					<div class="flex flex-wrap gap-3 pb-8 border-b border-stone-800">
						<button
							onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
							class="flex items-center gap-2 px-4 py-2 transition-all font-mono text-[9px] font-black uppercase bg-stone-950 text-stone-400 hover:text-white border border-stone-800 hover:border-rust/50 shadow-lg active:translate-y-px"
						>
							<Icon name="ph:terminal-window-bold" size="0.875rem" />
							<span>Console</span>
						</button>
						
						<!-- Version Controls -->
						<div class="flex gap-1 bg-black/40 p-1 border border-stone-800">
							<button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								disabled={!isOutdated}
								class="flex items-center gap-2 px-4 py-2 transition-all font-mono text-[9px] font-black uppercase {isOutdated ? 'bg-emerald-600 text-white hover:bg-emerald-500' : 'text-stone-600 opacity-50 cursor-not-allowed'}"
								title="Upgrade to latest version"
							>
								<Icon name="ph:arrow-up-to-line-bold" size="0.875rem" />
								<span>Upgrade</span>
							</button>
							<button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								class="flex items-center gap-2 px-4 py-2 transition-all font-mono text-[9px] font-black uppercase text-stone-500 hover:bg-stone-800 hover:text-white"
								title="Downgrade version (Select manually in node config)"
							>
								<Icon name="ph:arrow-down-to-line-bold" size="0.875rem" />
								<span>Downgrade</span>
							</button>
						</div>

						<button
							onclick={() => dispatch('delete', { nodeId, instanceId: instance.id })}
							disabled={instance.status === 'Running'}
							class="flex items-center gap-2 px-4 py-2 transition-all font-mono text-[9px] font-black uppercase bg-red-950/20 text-danger hover:bg-red-600 hover:text-white border border-red-900/30 ml-auto shadow-red-900/10"
						>
							<Icon name="ph:trash-bold" size="0.875rem" />
							<span>Delete</span>
						</button>
					</div>

					<!-- Load Chart -->
					<div class="space-y-4">
						<div class="flex justify-between items-end">
							<div class="flex flex-col gap-1">
								<span class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-[0.2em]">Usage Telemetry</span>
								<h4 class="text-[11px] font-heading font-black text-stone-300 uppercase tracking-[0.1em]">Concurrent Players (24h)</h4>
							</div>
							<div class="text-[10px] font-jetbrains font-black text-rust-light bg-rust/5 px-3 py-1 border border-rust/20 shadow-inner">
								Status: {instance.player_count || 0} Online
							</div>
						</div>
						<div class="bg-stone-950/60 border border-stone-800 p-6 industrial-frame shadow-inner">
							<PlayersChart data={chartData} height={160} color="var(--color-rust)" />
						</div>
					</div>
				</div>

				<!-- Right: Node Configuration -->
				<div class="xl:col-span-5 space-y-6">
					<div class="bg-stone-900/40 border border-stone-800 p-6 space-y-6 industrial-frame">
						<div class="space-y-3">
							<label for={'name-' + instance.id} class="text-[10px] font-jetbrains font-black text-stone-500 uppercase tracking-[0.2em] block">Label</label>
							<div class="flex gap-3">
								<input
									id={'name-' + instance.id}
									type="text"
									bind:value={renameValue}
									class="flex-1 bg-stone-950 border border-stone-800 px-4 py-2.5 text-xs font-jetbrains font-bold text-white focus:border-rust outline-none transition-all uppercase tracking-widest"
									placeholder={instance.id}
								/>
								<button
									onclick={handleRename}
									disabled={renameValue === instance.id || !renameValue.trim()}
									class="bg-stone-800 hover:bg-rust text-stone-400 hover:text-white px-5 py-2.5 text-[10px] font-heading font-black uppercase tracking-widest transition-all disabled:opacity-20 border border-stone-700 hover:border-rust-light shadow-lg"
								>
									Rename
								</button>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4 pt-2">
							<div class="bg-stone-950 border border-stone-800 p-4 flex flex-col gap-1 shadow-inner">
								<span class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-widest">Process ID</span>
								<span class="text-xs font-jetbrains font-black text-stone-300 uppercase tracking-tighter">{instance.pid || 'N/A'}</span>
							</div>
							<div class="bg-stone-950 border border-stone-800 p-4 flex flex-col gap-1 shadow-inner">
								<span class="text-[9px] font-jetbrains font-black text-stone-600 uppercase tracking-widest">Network Port</span>
								<span class="text-xs font-jetbrains font-black text-rust uppercase tracking-tighter">{instance.port || 'Auto'}</span>
							</div>
						</div>
					</div>

					<div class="bg-amber-500/5 border border-amber-500/20 p-5 flex gap-5 items-start industrial-frame">
						<div class="p-2.5 bg-amber-500/10 text-amber-500 border border-amber-500/20">
							<Icon name="activity" size="1rem" />
						</div>
						<div class="space-y-2">
							<span class="text-[10px] font-heading font-black text-amber-600 uppercase tracking-[0.2em] block">Monitoring</span>
							<p class="text-[11px] font-jetbrains font-medium text-stone-500 leading-relaxed uppercase tracking-tight">Standard health checks are active. Manual intervention is only required if heartbeats fail to sync.</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}
	.animate-flicker {
		animation: flicker 0.2s infinite;
	}

	.heartbeat-pulse::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(circle at 0% 50%, rgba(16, 185, 129, 0.05) 0%, transparent 50%);
		pointer-events: none;
		animation: heartbeat 4s ease-in-out infinite;
		z-index: 0;
	}

	@keyframes heartbeat {
		0%, 100% { opacity: 0.2; transform: scaleX(1); }
		10%, 20% { opacity: 0.5; transform: scaleX(1.1); }
	}

	.mask-fade-right {
		mask-image: linear-gradient(to right, white 0%, transparent 100%);
	}

    @keyframes binary-slide {
        0% { transform: translateY(0); }
        100% { transform: translateY(-50%); }
    }
    .animate-binary-slide {
        animation: binary-slide 10s linear infinite;
    }
</style>