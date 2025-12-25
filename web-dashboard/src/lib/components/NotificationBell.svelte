<script lang="ts">
	import { Bell, CheckCircle, Info, XCircle, AlertCircle } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { notifications } from '$lib/stores';

	const history = notifications.history;
	let isNotificationPanelOpen = $state(false);

	function toggleNotificationPanel() {
		isNotificationPanelOpen = !isNotificationPanelOpen;
	}
</script>

<div class="relative">
	<button
		onclick={toggleNotificationPanel}
		class="p-2 text-slate-500 dark:text-slate-400 hover:text-white bg-slate-800/50 hover:bg-slate-700 rounded-lg transition-all border border-slate-300/50 dark:border-slate-700/50 hover:border-rust/30 shadow-lg shadow-black/50"
	>
		<Bell class="w-5 h-5" />
		{#if $notifications.length > 0}
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-rust rounded-full border-2 border-slate-900 animate-ping"
			></span>
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-rust rounded-full border-2 border-slate-900"
			></span>
		{/if}
	</button>

	<!-- Notification Dropdown -->
	{#if isNotificationPanelOpen}
		<div
			class="absolute right-0 top-12 w-80 bg-[var(--card-bg)] backdrop-blur-xl border border-[var(--border-color)] rounded-xl shadow-2xl z-50 overflow-hidden"
			transition:slide={{ duration: 200 }}
		>
			<div
				class="px-4 py-3 border-b border-[var(--border-color)] flex justify-between items-center bg-black/40"
			>
				<span class="font-black text-slate-200 text-[10px] uppercase tracking-[0.2em] font-mono">Recent Activity</span>
				<button
					onclick={() => notifications.clearHistory()}
					class="text-[9px] font-bold uppercase tracking-widest text-rust-light hover:text-rust transition-colors italic">Clear_Log</button
				>
			</div>
			<div class="max-h-80 overflow-y-auto p-2 space-y-2 no-scrollbar">
				{#if $history && $history.length > 0}
					{#each $history as note (note.id)}
						<div
							class="flex gap-3 p-3 rounded-lg bg-black/40 border border-white/5 hover:border-rust/20 transition-all group"
						>
							<div class="mt-0.5">
								{#if note.type === 'success'}
									<CheckCircle class="w-4 h-4 text-emerald-500" />
								{:else if note.type === 'error'}
									<XCircle class="w-4 h-4 text-red-500" />
								{:else if note.type === 'warning'}
									<AlertCircle class="w-4 h-4 text-rust" />
								{:else}
									<Info class="w-4 h-4 text-rust-light" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-[11px] font-bold text-slate-200 leading-tight uppercase tracking-tight">
									{note.message}
								</p>
								{#if note.details}
									<p class="text-[9px] text-slate-500 mt-1 font-mono italic">{note.details}</p>
								{/if}
								<span class="text-[8px] text-slate-600 mt-2 block font-mono">
									{new Date(note.timestamp || Date.now()).toLocaleTimeString()}
								</span>
							</div>
						</div>
					{/each}
				{:else}
					<div class="text-center py-10">
						<div class="inline-block p-3 rounded-full bg-white/5 mb-2">
							<Bell class="w-6 h-6 text-stone-700" />
						</div>
						<p class="text-stone-600 text-[10px] font-mono uppercase tracking-widest">No signals detected</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
