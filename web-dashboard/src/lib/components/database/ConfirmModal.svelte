<script lang="ts">
	import { AlertTriangle, Trash2, X, Info } from 'lucide-svelte';
	import { scale, fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';

	type ModalVariant = 'danger' | 'warning' | 'info';

	let {
		isOpen = $bindable(false),
		title = 'Confirm Action',
		message = 'Are you sure you want to proceed?',
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		variant = 'danger' as ModalVariant,
		loading = false,
		onConfirm,
		onCancel
	}: {
		isOpen: boolean;
		title?: string;
		message?: string;
		confirmText?: string;
		cancelText?: string;
		variant?: ModalVariant;
		loading?: boolean;
		onConfirm: () => void | Promise<void>;
		onCancel?: () => void;
	} = $props();

	const variantConfig = {
		danger: {
			iconBg: 'bg-red-500/10',
			iconColor: 'text-danger',
			buttonBg: 'bg-danger hover:bg-red-500',
			buttonShadow: 'shadow-red-900/20'
		},
		warning: {
			iconBg: 'bg-amber-500/10',
			iconColor: 'text-warning',
			buttonBg: 'bg-warning hover:bg-amber-500',
			buttonShadow: 'shadow-amber-900/20'
		},
		info: {
			iconBg: 'bg-rust/10',
			iconColor: 'text-rust',
			buttonBg: 'bg-rust hover:bg-rust-light',
			buttonShadow: 'shadow-rust/20'
		}
	};

	$effect(() => {
		if (isOpen) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}
		return () => {
			document.body.style.overflow = '';
		};
	});

	function handleCancel() {
		if (loading) return;
		isOpen = false;
		onCancel?.();
	}

	async function handleConfirm() {
		if (loading) return;
		await onConfirm();
		isOpen = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!isOpen) return;
		if (e.key === 'Escape') {
			handleCancel();
		} else if (e.key === 'Enter' && !loading) {
			handleConfirm();
		}
	}

	const style = $derived(variantConfig[variant]);
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 bg-black/80 backdrop-blur-sm z-[70] flex items-center justify-center p-4"
		transition:fade={{ duration: 150 }}
		role="dialog"
		aria-modal="true"
		aria-labelledby="confirm-modal-title"
	>
		<!-- Click outside to close -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="fixed inset-0" onclick={handleCancel}></div>

		<!-- Modal -->
		<div
			class="bg-[var(--terminal-bg)] border border-stone-800 rounded-none w-full max-w-md shadow-2xl relative z-10 overflow-hidden industrial-frame"
			transition:scale={{ duration: 200, start: 0.95, easing: quintOut }}
		>
			<!-- Header -->
			<div class="p-8 pb-4">
				<div class="flex items-start gap-5">
					<!-- Icon -->
					<div class="shrink-0 p-3 rounded-none border border-stone-800 bg-stone-900/50 {style.iconColor}">
						{#if variant === 'danger'}
							<Trash2 class="w-6 h-6" />
						{:else if variant === 'warning'}
							<AlertTriangle class="w-6 h-6" />
						{:else}
							<Info class="w-6 h-6" />
						{#if variant === 'info'}
							<span class="sr-only">Info</span>
						{/if}
						{/if}
					</div>

					<!-- Content -->
					<div class="flex-1 min-w-0">
						<h3 id="confirm-modal-title" class="text-xl font-heading font-black text-slate-100 uppercase tracking-tighter mb-2">
							{title}
						</h3>
						<p class="font-jetbrains text-[11px] text-text-dim uppercase tracking-widest leading-relaxed">
							{message}
						</p>
					</div>

					<!-- Close button -->
					<button
						onclick={handleCancel}
						disabled={loading}
						class="shrink-0 p-1.5 text-text-dim hover:text-white transition-all disabled:opacity-20"
					>
						<X class="w-5 h-5" />
					</button>
				</div>
			</div>

			<!-- Actions -->
			<div class="p-8 flex items-center justify-end gap-4 bg-[var(--header-bg)]/50">
				<button
					onclick={handleCancel}
					disabled={loading}
					class="px-6 py-3 font-heading font-black text-[10px] text-text-dim hover:text-white transition-all uppercase tracking-widest disabled:opacity-20"
				>
					{cancelText}_OP
				</button>
				<button
					onclick={handleConfirm}
					disabled={loading}
					class="px-8 py-3 font-heading font-black text-[11px] text-white uppercase tracking-widest shadow-lg transition-all disabled:opacity-20 flex items-center gap-3 {style.buttonBg} {style.buttonShadow}"
				>
					{#if loading}
						<div
							class="w-4 h-4 border-2 border-white/30 border-t-white animate-spin"
						></div>
						<span>PROCESSING...</span>
					{:else}
						{#if variant === 'danger'}
							<Trash2 class="w-4 h-4" />
						{:else if variant === 'warning'}
							<AlertTriangle class="w-4 h-4" />
						{:else}
							<Info class="w-4 h-4" />
						{/if}
						<span>{confirmText}</span>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
