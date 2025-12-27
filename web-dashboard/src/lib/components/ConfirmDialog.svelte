<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';
	import DOMPurify from 'dompurify';
	import { siteSettings } from '$lib/stores';
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
		{#if $siteSettings.aesthetic.crt_effect}
			<div class="fixed inset-0 pointer-events-none z-[450] opacity-[0.03] bg-amber-scanlines"></div>
		{/if}

		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/90 {$siteSettings.aesthetic.glassmorphism ? 'backdrop-blur-sm' : ''}"
			onclick={!loading ? close : undefined}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && !loading && close()}
			role="button"
			tabindex="0"
			aria-label="Close dialog"
		></div>

		<!-- Modal Container - Industrial Terminal Style -->
		<div
			class="relative w-full max-w-lg bg-[#050505] shadow-2xl glass-panel overflow-hidden z-[460]
			{isCritical && !$siteSettings.aesthetic.industrial_styling ? 'border-red-900/50' : ''}
			{!isCritical && !$siteSettings.aesthetic.industrial_styling ? 'border-rust/30' : ''}"
			class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
			class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
			transition:modalScale
		>
			<!-- Status Bar -->
			<div class={`h-1 w-full ${isCritical ? 'bg-red-600' : 'bg-rust'} opacity-40 ${$siteSettings.aesthetic.animations_enabled ? 'animate-pulse' : ''}`}></div>

			<!-- Header -->
			<div class="px-8 py-5 border-b border-stone-800 flex justify-between items-center bg-[#0a0a0a]">
				<div class="flex items-center gap-4">
					{#if isCritical}
						<ShieldAlert class="w-5 h-5 text-red-500 {$siteSettings.aesthetic.animations_enabled ? 'animate-flicker' : ''}" />
					{:else}
						<Terminal class="w-5 h-5 text-rust-light" />
					{/if}
					<h3 class="text-xl font-heading font-black tracking-tighter text-white uppercase italic">
						{title}
					</h3>
				</div>
				<button
					onclick={close}
					class="text-stone-600 hover:text-white transition-all p-1"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Content -->
			<div class="p-10 space-y-8 relative overflow-hidden bg-[#050505]">
				<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
				
				<div class="flex items-start gap-6 relative z-10">
					<div class="flex-1 space-y-6">
						<div class="flex items-center gap-3 font-jetbrains text-[9px] font-black text-stone-600 tracking-[0.4em] uppercase italic">
							<ChevronRight class="w-3.5 h-3.5 {isCritical ? 'text-red-600' : 'text-rust'}" />
							SYSTEM_PROMPT_BUFFER
						</div>
						<div class="text-stone-300 font-jetbrains font-bold uppercase tracking-tight leading-relaxed">
							{#if loading && statusMessage}
								<p class="{$siteSettings.aesthetic.animations_enabled ? 'animate-pulse' : ''} text-rust">
									>> {statusMessage}
								</p>
							{:else}
								<p class="opacity-90">&gt;&gt; {message}</p>
							{/if}
						</div>
						{#if $$slots.default}
							<div class="pt-4 border-t border-stone-800/50">
								<slot />
							</div>
						{/if}
					</div>
				</div>

				{#if error}
					<div
						class="p-5 bg-red-950/20 text-red-500 font-jetbrains text-[10px] font-black uppercase tracking-widest"
						class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
						class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
						transition:scale={{ start: 0.98, duration: 200 }}
					>
						<div class="flex items-center gap-4">
							<AlertTriangle class="w-5 h-5 shrink-0" />
							<span>OP_FAULT: {error}</span>
						</div>
					</div>
				{/if}

				{#if loading && progress !== null}
					<!-- Progress -->
					<div class="space-y-4" transition:fade>
						<div class="flex justify-between font-jetbrains text-[9px] font-black text-stone-500 uppercase tracking-widest italic">
							<span>STREAM_PROGRESS</span>
							<span class="text-rust">{Math.round(progress)}%</span>
						</div>
						<div class="w-full h-1.5 bg-stone-950 border border-stone-800 p-0 relative shadow-inner overflow-hidden">
							<div
								class="h-full {isCritical ? 'bg-red-600 shadow-red-900/40' : 'bg-rust shadow-rust/40'} transition-all duration-300 ease-out shadow-lg"
								style="width: {progress}%"
							></div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Commands -->
			<div class="px-8 py-6 bg-[#0a0a0a] border-t border-stone-800 flex justify-between items-center">
				<div class="font-jetbrains text-[8px] font-black text-stone-700 tracking-[0.5em] uppercase italic">
					AWAITING_INPUT
				</div>
				<div class="flex gap-6">
					{#if loading && progress !== null}
						<div class="font-heading font-black text-[11px] text-stone-600 uppercase italic animate-pulse tracking-widest">
							[BUSY]
						</div>
					{:else if loading}
						<div class="flex items-center gap-3 font-heading font-black text-[11px] text-rust uppercase italic animate-pulse tracking-widest">
							<RefreshCw class="w-4 h-4 animate-spin" />
							EXECUTING...
						</div>
					{:else}
						<button
							onclick={close}
							class="px-6 py-2 font-heading font-black text-[11px] text-stone-600 hover:text-white uppercase tracking-widest italic transition-all"
						>
							[Cancel]
						</button>
						<button
							onclick={handleConfirm}
							class="px-8 py-3 font-heading font-black text-[11px] uppercase tracking-[0.2em] italic transition-all {isCritical ? 'bg-red-700 text-white hover:bg-red-600' : 'bg-rust text-white hover:bg-rust-light'} shadow-lg active:translate-y-px active:shadow-none"
							class:industrial-frame={!$siteSettings.aesthetic.industrial_styling}
							class:industrial-sharp={$siteSettings.aesthetic.industrial_styling}
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