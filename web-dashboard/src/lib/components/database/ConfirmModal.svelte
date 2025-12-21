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
			iconColor: 'text-red-400',
			buttonBg: 'bg-red-600 hover:bg-red-500',
			buttonShadow: 'shadow-red-900/20'
		},
		warning: {
			iconBg: 'bg-amber-500/10',
			iconColor: 'text-amber-400',
			buttonBg: 'bg-amber-600 hover:bg-amber-500',
			buttonShadow: 'shadow-amber-900/20'
		},
		info: {
			iconBg: 'bg-blue-500/10',
			iconColor: 'text-blue-400',
			buttonBg: 'bg-blue-600 hover:bg-blue-500',
			buttonShadow: 'shadow-blue-900/20'
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
		class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[70] flex items-center justify-center p-4"
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
			class="bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-2xl w-full max-w-md shadow-2xl relative z-10 overflow-hidden"
			transition:scale={{ duration: 200, start: 0.95, easing: quintOut }}
		>
			<!-- Header -->
			<div class="p-6 pb-4">
				<div class="flex items-start gap-4">
					<!-- Icon -->
					<div class="shrink-0 p-3 rounded-xl {style.iconBg}">
						{#if variant === 'danger'}
							<Trash2 class="w-6 h-6 {style.iconColor}" />
						{:else if variant === 'warning'}
							<AlertTriangle class="w-6 h-6 {style.iconColor}" />
						{:else}
							<Info class="w-6 h-6 {style.iconColor}" />
						{/if}
					</div>

					<!-- Content -->
					<div class="flex-1 min-w-0">
						<h3 id="confirm-modal-title" class="text-lg font-bold text-slate-100 mb-1">
							{title}
						</h3>
						<p class="text-sm text-slate-500 dark:text-slate-400 leading-relaxed">
							{message}
						</p>
					</div>

					<!-- Close button -->
					<button
						onclick={handleCancel}
						disabled={loading}
						class="shrink-0 p-1.5 text-slate-500 hover:text-slate-700 dark:text-slate-300 hover:bg-slate-800 rounded-lg transition-colors disabled:opacity-50"
					>
						<X class="w-5 h-5" />
					</button>
				</div>
			</div>

			<!-- Actions -->
			<div class="px-6 pb-6 flex items-center justify-end gap-3">
				<button
					onclick={handleCancel}
					disabled={loading}
					class="px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:text-white hover:bg-slate-800 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{cancelText}
				</button>
				<button
					onclick={handleConfirm}
					disabled={loading}
					class="px-5 py-2.5 text-sm font-bold text-slate-900 dark:text-white rounded-xl shadow-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 {style.buttonBg} {style.buttonShadow}"
				>
					{#if loading}
						<div
							class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
						></div>
						<span>Processing...</span>
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
