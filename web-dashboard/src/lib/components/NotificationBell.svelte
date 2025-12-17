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
		class="p-2 text-slate-400 hover:text-white bg-slate-800/50 hover:bg-slate-700 rounded-lg transition-all border border-slate-700/50 hover:border-blue-500/30"
	>
		<Bell class="w-5 h-5" />
		{#if $notifications.length > 0}
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-slate-900 animate-ping"
			></span>
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-slate-900"
			></span>
		{/if}
	</button>

	<!-- Notification Dropdown -->
	{#if isNotificationPanelOpen}
		<div
			class="absolute right-0 top-12 w-80 bg-slate-900/95 backdrop-blur-xl border border-slate-700 rounded-xl shadow-2xl z-50 overflow-hidden"
			transition:slide={{ duration: 200 }}
		>
			<div
				class="p-3 border-b border-slate-700/50 flex justify-between items-center bg-slate-800/30"
			>
				<span class="font-bold text-slate-200 text-sm">Recent Activity</span>
				<button
					onclick={() => notifications.clearHistory()}
					class="text-xs text-blue-400 hover:text-blue-300">Clear</button
				>
			</div>
			<div class="max-h-80 overflow-y-auto p-2 space-y-2">
				{#if $history && $history.length > 0}
					{#each $history as note (note.id)}
						<div
							class="flex gap-3 p-2 rounded-lg bg-slate-800/30 border border-slate-700/30 hover:bg-slate-800/50 transition-colors"
						>
							<div class="mt-0.5">
								{#if note.type === 'success'}
									<CheckCircle class="w-4 h-4 text-emerald-400" />
								{:else if note.type === 'error'}
									<XCircle class="w-4 h-4 text-red-400" />
								{:else if note.type === 'warning'}
									<AlertCircle class="w-4 h-4 text-orange-400" />
								{:else}
									<Info class="w-4 h-4 text-blue-400" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-xs font-medium text-slate-200 truncate">{note.message}</p>
								{#if note.details}
									<p class="text-[10px] text-slate-500 truncate">{note.details}</p>
								{/if}
								<span class="text-[10px] text-slate-600 mt-1 block">
									{new Date(note.timestamp || Date.now()).toLocaleTimeString()}
								</span>
							</div>
						</div>
					{/each}
				{:else}
					<div class="text-center py-8 text-slate-500 text-xs">No recent notifications</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
