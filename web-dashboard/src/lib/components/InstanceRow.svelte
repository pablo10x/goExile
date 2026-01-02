<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { serverVersions } from '$lib/stores';
	import { compareVersions } from '$lib/semver';
	import PlayersChart from './PlayersChart.svelte';
	import Icon from './theme/Icon.svelte';
	import Button from './Button.svelte';

	let { nodeId, instance }: { nodeId: number; instance: any } = $props();

	let expanded = $state(false);
	let isHovered = $state(false);
	let renameValue = $state(instance.id);
	let chartData = $state<any[]>([]);

    const binaryLines = Array.from({ length: 8 }, () => 
        Array.from({ length: 16 }, () => Math.random() > 0.5 ? '1' : '0').join('')
    );

	$effect(() => {
		renameValue = instance.id;
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
	class="border border-slate-700 rounded-xl glass-panel overflow-hidden mb-2 hover:border-blue-500/40 transition-all duration-500 shadow-lg group/row relative bg-slate-800/40"
	class:heartbeat-pulse={instance.status === 'Running'}
	onmouseenter={() => isHovered = true}
	onmouseleave={() => isHovered = false}
	role="region"
	aria-label={`Instance ${instance.id}`}
>
	<!-- Binary Animation Overlays -->
	{#if instance.status === 'Running' && isHovered}
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
				<span class="text-[9px] font-jetbrains font-black uppercase tracking-tighter text-stone-500">INSTANCE</span>
				<span class="font-heading font-black text-sm text-white tracking-widest uppercase">{instance.id.split('-').pop() || instance.port}</span>
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
			<Button
				onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
				variant="ghost"
				size="xs"
				icon="ph:terminal-window-bold"
				title="Console"
			/>

			<div class="w-px h-5 bg-stone-800 mx-1"></div>

			{#if instance.status !== 'Running'}
				<Button
					onclick={() => dispatch('start', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:play-bold"
					title="Start"
					class="!text-emerald-600 hover:!text-success hover:bg-emerald-500/5"
				/>
			{:else}
				<Button
					onclick={() => dispatch('stop', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:stop-bold"
					title="Stop"
					class="hover:!text-yellow-500 hover:bg-yellow-500/5"
				/>
				<Button
					onclick={() => dispatch('restart', { nodeId, instanceId: instance.id })}
					variant="ghost"
					size="xs"
					icon="ph:arrows-clockwise-bold"
					title="Restart"
					class="hover:!text-rust-light hover:bg-rust/5"
				/>
			{/if}
		</div>
	</div>

	<!-- Expanded Details -->
	{#if expanded}
		<div
			transition:slide={{ duration: 300 }}
			class="bg-slate-900/60 border-t border-slate-700 p-8 space-y-8 relative z-10"
		>
			<div class="grid grid-cols-1 xl:grid-cols-12 gap-10">
				<!-- Left: Technical Readouts -->
				<div class="xl:col-span-7 space-y-8">
					<!-- Primary Toolbar -->
					<div class="flex flex-wrap gap-3 pb-8 border-b border-stone-800">
						<Button
							onclick={() => dispatch('tail', { nodeId, instanceId: instance.id })}
							variant="secondary"
							size="xs"
							icon="ph:terminal-window-bold"
						>
							Console
						</Button>
						
						<!-- Version Controls -->
						<div class="flex gap-1 bg-black/40 p-1 border border-stone-800">
							<Button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								disabled={!isOutdated}
								variant={isOutdated ? 'success' : 'secondary'}
								size="xs"
								icon="ph:arrow-up-to-line-bold"
								title="Upgrade to latest version"
							>
								Upgrade
							</Button>
							<Button
								onclick={() => dispatch('update', { nodeId, instanceId: instance.id })}
								variant="secondary"
								size="xs"
								icon="ph:arrow-down-to-line-bold"
								title="Downgrade version (Select manually in node config)"
							>
								Downgrade
							</Button>
						</div>

						<Button
							onclick={() => dispatch('delete', { nodeId, instanceId: instance.id })}
							disabled={instance.status === 'Running'}
							variant="danger"
							size="xs"
							icon="ph:trash-bold"
							class="ml-auto"
						>
							Delete
						</Button>
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
						<div class="bg-stone-950/60 border border-stone-800 p-6 industrial-sharp shadow-inner">
							<PlayersChart data={chartData} height={160} color="#c2410c" />
						</div>
					</div>
				</div>

				<!-- Right: Node Configuration -->
				<div class="xl:col-span-5 space-y-6">
					<div class="bg-stone-900/40 border border-stone-800 p-6 space-y-6 industrial-sharp">
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
								<Button
									onclick={handleRename}
									disabled={renameValue === instance.id || !renameValue.trim()}
									variant="secondary"
									size="xs"
								>
									Rename
								</Button>
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

					<div class="bg-amber-500/5 border border-amber-500/20 p-5 flex gap-5 items-start industrial-sharp">
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