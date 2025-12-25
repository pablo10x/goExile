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
		class="p-2 text-stone-500 hover:text-white bg-stone-900/50 hover:bg-rust/10 rounded-none transition-all border border-stone-800 hover:border-rust shadow-xl active:translate-y-px"
	>
		<Bell class="w-5 h-5 transition-transform group-hover:scale-110" />
		{#if $notifications.length > 0}
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-rust rounded-full border-2 border-stone-950 animate-ping"
			></span>
			<span
				class="absolute top-1 right-1 w-2.5 h-2.5 bg-rust rounded-full border-2 border-stone-950"
			></span>
		{/if}
	</button>

	<!-- Notification Dropdown -->
	{#if isNotificationPanelOpen}
		<div
			class="absolute right-0 top-12 w-80 bg-black/90 backdrop-blur-xl border border-stone-800 rounded-none shadow-[0_0_50px_rgba(0,0,0,0.8)] z-50 overflow-hidden industrial-frame"
			transition:slide={{ duration: 200 }}
		>
			<div
				class="px-5 py-4 border-b border-stone-800 flex justify-between items-center bg-[#0a0a0a]"
			>
				<span class="font-heading font-black text-white text-[10px] uppercase tracking-[0.2em]">Recent_Activity_Buffer</span>
				<button
					onclick={() => notifications.clearHistory()}
					class="text-[9px] font-black uppercase tracking-widest text-rust hover:text-rust-light transition-all italic border border-stone-800 px-3 py-1 bg-stone-950 hover:border-rust shadow-inner">PURGE</button
				>
			</div>
			<div class="max-h-80 overflow-y-auto p-3 space-y-3 no-scrollbar bg-[#050505] relative">
				<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
				
				{#if $history && $history.length > 0}
					{#each $history as note (note.id)}
						<div
							class="flex gap-4 p-4 rounded-none bg-stone-900/40 border border-stone-800 hover:border-rust/30 transition-all group relative overflow-hidden"
						>
							<div class="mt-0.5 shrink-0">
								{#if note.type === 'success'}
									<div class="p-1.5 bg-emerald-500/10 border border-emerald-500/30 industrial-frame">
										<CheckCircle class="w-3.5 h-3.5 text-emerald-500" />
									</div>
								{:else if note.type === 'error'}
									<div class="p-1.5 bg-red-500/10 border border-red-500/30 industrial-frame">
										<XCircle class="w-3.5 h-3.5 text-red-500" />
									</div>
								{:else if note.type === 'warning'}
									<div class="p-1.5 bg-rust/10 border border-rust/30 industrial-frame">
										<AlertCircle class="w-3.5 h-3.5 text-rust" />
									</div>
								{:else}
									<div class="p-1.5 bg-stone-800 border border-stone-700 industrial-frame">
										<Info class="w-3.5 h-3.5 text-stone-400" />
									</div>
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-[11px] font-black text-stone-200 leading-tight uppercase tracking-tight">
									{note.message}
								</p>
								{#if note.details}
									<p class="text-[9px] text-stone-500 mt-2 font-jetbrains font-bold leading-relaxed uppercase opacity-60 italic border-l border-stone-800 pl-3">{note.details}</p>
								{/if}
								<span class="text-[8px] font-jetbrains font-black text-stone-700 mt-3 block uppercase tracking-widest">
									Captured: {new Date(note.timestamp || Date.now()).toLocaleTimeString([], { hour12: false })}
								</span>
							</div>
						</div>
					{/each}
				{:else}
					<div class="text-center py-16 opacity-40">
						<div class="inline-block p-5 bg-stone-900/40 border border-stone-800 industrial-frame mb-4">
							<Bell class="w-8 h-8 text-stone-700" />
						</div>
						<p class="text-stone-600 text-[10px] font-jetbrains font-black uppercase tracking-[0.3em]">No_Signals_Detected</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
