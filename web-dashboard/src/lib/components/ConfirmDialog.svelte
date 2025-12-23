<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';
	import DOMPurify from 'dompurify';
	import { Terminal, ShieldAlert, AlertTriangle, RefreshCw, X, ChevronRight } from 'lucide-svelte';

	export let isOpen: boolean = false;
	export let title: string = 'Confirm Action';
	export let message: string = 'Are you sure you want to proceed?';
	export let confirmText: string = 'Confirm';
	export let cancelText: string = 'Cancel';
	export let isCritical: boolean = false; // If true, use red colors
	export let progress: number | null = null; // 0-100, if set shows progress bar
	export let statusMessage: string | null = null; // Message during loading/progress

	export let onConfirm: () => Promise<void>;

	const dispatch = createEventDispatcher();

	let loading = false;
	let error: string | null = null;

	async function handleConfirm() {
		loading = true;
		error = null;
		try {
			await onConfirm();
			dispatch('success');
			close();
		} catch (e: any) {
			error = e.message || 'ERR_OP_FAILED: Unexpected interrupt';
			loading = false;
		}
	}

	function close() {
		if (loading && progress !== null) return;
		isOpen = false;
		error = null;
		loading = false;
		dispatch('close');
	}

	// Industrial entrance
	function modalScale(node: HTMLElement, params: { duration?: number } = {}) {
		const { duration = 300 } = params;
		return {
			duration,
			css: (t: number) => {
				const eased = cubicOut(t);
				return `
					transform: scale(${0.98 + eased * 0.02}) translateY(${(1 - eased) * 10}px);
					opacity: ${eased};
				`;
			}
		};
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-[400] flex items-center justify-center p-4 sm:p-6 font-['JetBrains_Mono',monospace]"
		transition:fade={{ duration: 200 }}
	>
		<!-- CRT Overlay Effect (Dialog Scope) -->
		<div class="fixed inset-0 pointer-events-none z-[450] opacity-[0.03] bg-amber-scanlines"></div>

		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/90 backdrop-blur-sm"
			onclick={!loading ? close : undefined}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && !loading && close()}
			role="button"
			tabindex="0"
			aria-label="Close dialog"
		></div>

		<!-- Modal Container - Industrial Terminal Style -->
		<div
			class="relative w-full max-w-lg bg-black border-2 {isCritical ? 'border-red-900 shadow-[0_0_40px_rgba(153,27,27,0.2)]' : 'border-[#f97316]/30 shadow-[0_0_40px_rgba(249,115,22,0.1)]'} overflow-hidden z-[460]"
			transition:modalScale
		>
			<!-- Status Bar -->
			<div class="h-1 w-full {isCritical ? 'bg-red-600' : 'bg-[#f97316]'} opacity-50 animate-pulse"></div>

			<!-- Header -->
			<div class="px-8 py-5 border-b border-white/5 flex justify-between items-center bg-black">
				<div class="flex items-center gap-4">
					{#if isCritical}
						<ShieldAlert class="w-5 h-5 text-red-600 animate-flicker" />
					{:else}
						<Terminal class="w-5 h-5 text-[#f97316]" />
					{/if}
					<h3 class="text-lg font-black italic tracking-tighter text-white uppercase">
						{title}
					</h3>
				</div>
				<button
					onclick={close}
					class="text-slate-700 hover:text-white transition-all p-1"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Content -->
			<div class="p-8 space-y-6 relative overflow-hidden bg-[#050505]">
				<div class="flex items-start gap-6 relative z-10">
					<div class="flex-1 space-y-4">
						<div class="flex items-center gap-3 text-[10px] font-black text-slate-600 tracking-[0.4em] uppercase italic">
							<ChevronRight class="w-3 h-3 {isCritical ? 'text-red-600' : 'text-[#f97316]'}" />
							System_Prompt
						</div>
						<div class="text-slate-400 text-sm leading-relaxed italic font-bold uppercase tracking-tight">
							{#if loading && statusMessage}
								<p class="animate-pulse text-[#f97316]">
									>> {statusMessage}
								</p>
							{:else}
								<p>&gt;&gt; {message}</p>
							{/if}
						</div>
						<div class="pt-2 border-t border-white/5">
							<slot />
						</div>
					</div>
				</div>

				{#if error}
					<div
						class="p-4 bg-red-950/20 border-l-4 border-red-600 text-red-500 text-xs italic font-black animate-flicker"
						transition:scale={{ start: 0.98, duration: 200 }}
					>
						<div class="flex items-center gap-3">
							<AlertTriangle class="w-4 h-4" />
							<span>ERROR: {error}</span>
						</div>
					</div>
				{/if}

				{#if loading && progress !== null}
					<!-- Progress -->
					<div class="space-y-3" transition:fade>
						<div class="flex justify-between text-[9px] font-black text-slate-600 uppercase tracking-widest italic">
							<span>Stream_Progress</span>
							<span>{Math.round(progress)}%</span>
						</div>
						<div class="w-full h-3 bg-black border border-white/10 p-[2px] shadow-inner">
							<div
								class="h-full {isCritical ? 'bg-red-600 shadow-[0_0_10px_rgba(220,38,38,0.5)]' : 'bg-[#f97316] shadow-[0_0_10px_rgba(249,115,22,0.5)]'} transition-all duration-300 ease-out relative"
								style="width: {progress}%"
							>
								<div class="absolute inset-0 bg-white/20 animate-pulse"></div>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Commands -->
			<div class="px-8 py-6 bg-black border-t border-white/5 flex justify-between items-center">
				<div class="text-[8px] font-black text-slate-800 tracking-[0.5em] uppercase italic">
					Waiting_For_Instruction
				</div>
				<div class="flex gap-6">
					{#if loading && progress !== null}
						<div class="text-[10px] font-black text-slate-600 uppercase italic animate-pulse">
							[BUSY]
						</div>
					{:else if loading}
						<div class="flex items-center gap-3 text-[10px] font-black text-[#f97316] uppercase italic animate-pulse">
							<RefreshCw class="w-4 h-4 animate-spin" />
							EXECUTING...
						</div>
					{:else}
						<button
							onclick={close}
							class="px-6 py-2 text-[11px] font-black text-slate-600 hover:text-white uppercase tracking-widest italic transition-all"
						>
							[Cancel]
						</button>
						<button
							onclick={handleConfirm}
							class="px-8 py-3 text-[11px] font-black uppercase tracking-[0.3em] italic transition-all {isCritical ? 'bg-red-700 text-black hover:bg-red-600' : 'bg-[#f97316] text-black hover:bg-white'} shadow-[6px_6px_0px_#000] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none"
						>
							{confirmText.toUpperCase()}
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.bg-amber-scanlines {
		background: linear-gradient(
			rgba(18, 16, 16, 0) 50%,
			rgba(0, 0, 0, 0.25) 50%
		),
		linear-gradient(
			90deg,
			rgba(255, 0, 0, 0.03),
			rgba(0, 255, 0, 0.01),
			rgba(0, 0, 255, 0.03)
		);
		background-size: 100% 4px, 4px 100%;
	}

	@keyframes flicker {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.8; }
		55% { opacity: 0.95; }
		60% { opacity: 0.7; }
	}
	.animate-flicker {
		animation: flicker 0.25s infinite;
	}
</style>