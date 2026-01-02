<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import Icon from '$lib/components/theme/Icon.svelte';
	import Button from './Button.svelte';

	let {
		isOpen = $bindable(false),
		title = 'Confirm Action',
		message = 'Are you sure you want to proceed?',
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		isCritical = false,
		progress = null,
		statusMessage = null,
		onConfirm,
		children
	}: {
		isOpen: boolean;
		title?: string;
		message?: string;
		confirmText?: string;
		cancelText?: string;
		isCritical?: boolean;
		progress?: number | null;
		statusMessage?: string | null;
		onConfirm: () => Promise<void>;
		children?: any;
	} = $props();

	const dispatch = createEventDispatcher();

	let loading = $state(false);
	let error = $state<string | null>(null);

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
		class="fixed inset-0 z-[400] flex items-center justify-center p-4 sm:p-6 font-mono"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/90 backdrop-blur-sm"
			onclick={!loading ? close : undefined}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && !loading && close()}
			role="button"
			tabindex="0"
			aria-label="Close dialog"
		></div>

		<!-- Modal Container -->
		<div
			class="relative w-full max-w-lg bg-slate-900 shadow-2xl overflow-hidden z-[460] border border-slate-700 rounded-xl"
			transition:modalScale
		>
			<!-- Status Bar -->
			<div class={`h-1 w-full ${isCritical ? 'bg-red-600' : 'bg-rust'} opacity-40 animate-pulse`}></div>

			<!-- Header -->
			<div class="px-8 py-5 border-b border-slate-700 flex justify-between items-center bg-slate-800/50">
				<div class="flex items-center gap-4">
					{#if isCritical}
						<Icon name="shield" size="1.25rem" class="text-red-500 animate-flicker" />
					{:else}
						<Icon name="file-text" size="1.25rem" class="text-rust-light" />
					{/if}
					<h3 class="text-xl font-heading font-black tracking-tighter text-white uppercase italic">
						{title}
					</h3>
				</div>
				<button
					onclick={close}
					class="text-stone-600 hover:text-white transition-all p-1"
				>
					<Icon name="ph:x-bold" size="1.25rem" />
				</button>
			</div>

			<!-- Content -->
			<div class="p-10 space-y-8 relative overflow-hidden bg-slate-900">
				<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
				
				<div class="flex items-start gap-6 relative z-10">
					<div class="flex-1 space-y-6">
						<div class="flex items-center gap-3 font-jetbrains text-[10px] font-black tracking-[0.4em] uppercase italic text-stone-500">
							<Icon name="ph:caret-right-bold" size="0.875rem" class="{isCritical ? 'text-red-600' : 'text-rust'}" />
							SYSTEM_PROMPT_BUFFER
						</div>
						<div class="text-stone-300 font-jetbrains font-bold uppercase tracking-tight leading-relaxed">
							{#if loading && statusMessage}
								<p class="animate-pulse text-rust">
									>> {statusMessage}
								</p>
							{:else}
								<p class="opacity-90">&gt;&gt; {message}</p>
							{/if}
						</div>
						{#if children}
							<div class="pt-4 border-t border-stone-800/50">
								{@render children()}
							</div>
						{/if}
					</div>
				</div>

				{#if error}
					<div
						class="p-5 bg-red-950/20 text-red-500 font-jetbrains text-[10px] font-black uppercase tracking-widest border border-red-900/30"
						transition:scale={{ start: 0.98, duration: 200 }}
					>
						<div class="flex items-center gap-4">
							<Icon name="alert" size="1.25rem" class="shrink-0" />
							<span>OP_FAULT: {error}</span>
						</div>
					</div>
				{/if}

				{#if loading && progress !== null}
					<!-- Progress -->
					<div class="space-y-4" transition:fade>
						<div class="flex justify-between font-jetbrains text-[9px] font-black uppercase tracking-widest italic text-stone-500">
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
			<div class="px-8 py-6 bg-slate-800/50 border-t border-slate-700 flex justify-between items-center">
				<div class="font-jetbrains text-[10px] font-black tracking-[0.5em] uppercase italic text-slate-500">
					AWAITING_INPUT
				</div>
				<div class="flex gap-4">
					{#if loading && progress !== null}
						<div class="font-heading font-black text-[11px] text-stone-600 uppercase italic animate-pulse tracking-widest">
							[BUSY]
						</div>
					{:else if loading}
						<div class="flex items-center gap-3 font-heading font-black text-[11px] text-rust uppercase italic animate-pulse tracking-widest">
							<Icon name="ph:arrows-clockwise-bold" size="1rem" class="animate-spin" />
							EXECUTING...
						</div>
					{:else}
						<Button
							onclick={close}
							variant="ghost"
							size="sm"
						>
							{cancelText}
						</Button>
						<Button
							onclick={handleConfirm}
							variant={isCritical ? 'danger' : 'primary'}
							size="md"
						>
							{confirmText}
						</Button>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
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