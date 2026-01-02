<script lang="ts">
	import { Bell, CheckCircle, Info, XCircle, AlertCircle } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { notifications } from '$lib/stores';
	import ConfirmDialog from './ConfirmDialog.svelte';

	const history = notifications.history;
	let isNotificationPanelOpen = $state(false);

	let isConfirmOpen = $state(false);

	function toggleNotificationPanel() {
		isNotificationPanelOpen = !isNotificationPanelOpen;
	}

	async function requestPurge() {
		isConfirmOpen = true;
	}

	async function executePurge() {
		notifications.clearPermanentHistory();
		isConfirmOpen = false;
	}
</script>

<div class="relative">
	<button
		onclick={toggleNotificationPanel}
		class="p-2 text-stone-500 hover:text-white bg-stone-900/50 hover:bg-rust/10 transition-all active:translate-y-px industrial-sharp border border-stone-800"
	>
		<Bell class="w-5 h-5 transition-transform group-hover:scale-110" />
		{#if $notifications.length > 0}
			<span
				class="absolute top-0 right-0 w-2.5 h-2.5 bg-rust rounded-full border-2 border-stone-950 animate-ping"
			></span>
			<span
				class="absolute top-0 right-0 w-2.5 h-2.5 bg-rust rounded-full border-2 border-stone-950"
			></span>
		{/if}
	</button>

	<!-- Notification Dropdown -->
	{#if isNotificationPanelOpen}
		<div
			class="absolute right-0 top-14 w-96 bg-stone-950/95 backdrop-blur-xl shadow-[0_0_50px_rgba(0,0,0,0.8)] z-50 overflow-hidden border border-stone-800 rounded-none"
			transition:slide={{ duration: 200 }}
		>
			<div
				class="px-5 py-4 border-b border-stone-800 flex justify-between items-center bg-black/40"
			>
				<span class="font-heading font-black text-white text-[10px] uppercase tracking-[0.2em]">Recent_Activity_Buffer</span>
				<button
					onclick={requestPurge}
					class="text-[9px] font-black uppercase tracking-widest text-rust hover:text-white transition-all bg-rust/10 hover:bg-rust px-3 py-1 border border-rust/30 rounded-none"
				>PURGE_LOGS</button
				>
			</div>
			<div class="max-h-80 overflow-y-auto p-3 space-y-2 custom-scrollbar bg-[#050505] relative">
				<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.03] pointer-events-none"></div>
				
				{#if $history && $history.length > 0}
					{#each $history as note (note.id)}
						<div
							class="flex gap-4 p-4 bg-stone-900/30 border border-stone-800 hover:border-stone-700 transition-all group relative overflow-hidden rounded-none"
						>
							<div class="mt-0.5 shrink-0">
								{#if note.type === 'success'}
									<div class="p-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-none">
										<CheckCircle class="w-3.5 h-3.5 text-emerald-500" />
									</div>
								{:else if note.type === 'error'}
									<div class="p-1.5 bg-red-500/10 border border-red-500/20 rounded-none">
										<XCircle class="w-3.5 h-3.5 text-red-500" />
									</div>
								{:else if note.type === 'warning'}
									<div class="p-1.5 bg-amber-500/10 border border-amber-500/20 rounded-none">
										<AlertCircle class="w-3.5 h-3.5 text-amber-500" />
									</div>
								{:else}
									<div class="p-1.5 bg-blue-500/10 border border-blue-500/20 rounded-none">
										<Info class="w-3.5 h-3.5 text-blue-500" />
									</div>
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-[10px] font-bold text-stone-300 leading-tight uppercase tracking-tight group-hover:text-white transition-colors">
									{note.message}
								</p>
								{#if note.details}
									<p class="text-[9px] text-stone-500 mt-1.5 font-jetbrains font-medium leading-relaxed opacity-70 break-all">{note.details}</p>
								{/if}
								<span class="text-[8px] font-mono text-stone-600 mt-2 block uppercase tracking-wider">
									{new Date(note.timestamp || Date.now()).toLocaleTimeString([], { hour12: false })}
								</span>
							</div>
						</div>
					{/each}
				{:else}
					<div class="text-center py-16 opacity-40 flex flex-col items-center">
						<div class="p-4 bg-stone-900/50 border border-stone-800 mb-3 rounded-none">
							<Bell class="w-6 h-6 text-stone-600" />
						</div>
						<p class="text-stone-500 text-[9px] font-jetbrains font-bold uppercase tracking-[0.2em]">Buffer Empty</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<ConfirmDialog
	bind:isOpen={isConfirmOpen}
	title="Purge Activity Buffer"
	message="Are you sure you want to permanently delete all recent activity history? This action is localized and irreversible."
	isCritical={true}
	onConfirm={executePurge}
/>