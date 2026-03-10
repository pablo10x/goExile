<script lang="ts">
	import { apiFetch } from '$lib/api';
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
			error = e.message || 'Operation failed. Please try again.';
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
		class="fixed inset-0 z-[400] flex items-center justify-center p-4 sm:p-6 font-sans"
		transition:fade={{ duration: 200 }}
	>
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/60 backdrop-blur-sm"
			onclick={!loading ? close : undefined}
			onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && !loading && close()}
			role="button"
			tabindex="0"
			aria-label="Close dialog"
		></div>

		<!-- Modal Container -->
		<div
			class="relative w-full max-w-lg bg-slate-900/95 backdrop-blur-3xl shadow-2xl overflow-hidden z-[460] border border-white/10 rounded-3xl"
			transition:modalScale
		>
			<!-- Header -->
			<div class="px-8 py-6 border-b border-white/5 flex justify-between items-center bg-black/20">
				<div class="flex items-center gap-4">
					<div
						class={`p-2.5 rounded-xl ${isCritical ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20' : 'bg-sky-500/10 text-sky-400 border border-sky-500/20'}`}
					>
						<Icon
							name={isCritical ? 'ph:warning-circle-bold' : 'ph:question-bold'}
							size="1.25rem"
						/>
					</div>
					<h3 class="text-xl font-bold text-white tracking-tight uppercase italic font-heading">
						{title}
					</h3>
				</div>
				<button
					onclick={close}
					class="text-slate-500 hover:text-white transition-all p-2 rounded-lg hover:bg-white/5"
				>
					<Icon name="ph:x-bold" size="1.25rem" />
				</button>
			</div>

			<!-- Content -->
			<div class="p-10 space-y-6 relative overflow-hidden bg-transparent">
				<div class="flex items-start gap-6 relative z-10">
					<div class="flex-1 space-y-4">
						<div class="text-slate-300 font-medium leading-relaxed">
							{#if loading && statusMessage}
								<p class="animate-pulse {isCritical ? 'text-rose-400' : 'text-emerald-400'} font-bold uppercase tracking-wider text-sm">
									{statusMessage}
								</p>
							{:else}
								<p class="text-lg opacity-90 text-slate-200">{message}</p>
							{/if}
						</div>
						{#if children}
							<div class="pt-4 border-t border-white/5">
								{@render children()}
							</div>
						{/if}
					</div>
				</div>

				{#if error}
					<div
						class="p-4 bg-rose-500/10 text-rose-400 text-sm font-medium rounded-xl border border-rose-500/20"
						transition:scale={{ start: 0.98, duration: 200 }}
					>
						<div class="flex items-center gap-3">
							<Icon name="ph:warning-bold" size="1.1rem" class="shrink-0" />
							<span>{error}</span>
						</div>
					</div>
				{/if}

				{#if loading && progress !== null}
					<!-- Progress -->
					<div class="space-y-3" transition:fade>
						<div
							class="flex justify-between text-xs font-bold uppercase tracking-wider text-slate-500"
						>
							<span>Operation Progress</span>
							<span class={isCritical ? 'text-rose-400' : 'text-emerald-400'}>{Math.round(progress)}%</span>
						</div>
						<div class="w-full h-2 bg-black rounded-full p-0.5 border border-white/5 shadow-inner">
							<div
								class="h-full rounded-full {isCritical
									? 'bg-rose-500'
									: 'bg-emerald-500'} transition-all duration-300 ease-out shadow-[0_0_10px_rgba(16,185,129,0.3)]"
								style="width: {progress}%"
							></div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="px-8 py-6 bg-black/20 border-t border-white/5 flex justify-end items-center gap-3">
				{#if loading && progress !== null}
					<div class="text-sm font-bold text-slate-500 uppercase tracking-wider animate-pulse">
						Processing...
					</div>
				{:else if loading}
					<div
						class="flex items-center gap-3 text-sm font-bold {isCritical ? 'text-rose-400' : 'text-sky-400'} uppercase tracking-wider animate-pulse"
					>
						<Icon name="ph:arrows-clockwise-bold" size="1rem" class="animate-spin" />
						Executing...
					</div>
				{:else}
					<Button onclick={close} variant="ghost" size="md" class="!rounded-xl">
						{cancelText}
					</Button>
					<Button
						onclick={handleConfirm}
						variant={isCritical ? 'danger' : 'primary'}
						size="lg"
						class="!rounded-xl"
					>
						{confirmText}
					</Button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes flicker {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.8;
		}
		55% {
			opacity: 0.95;
		}
		60% {
			opacity: 0.7;
		}
	}
	.animate-flicker {
		animation: flicker 0.25s infinite;
	}
</style>
