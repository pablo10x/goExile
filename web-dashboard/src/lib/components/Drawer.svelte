<script lang="ts">
	export let isOpen: boolean;
	export let title: string;
	export let onClose: () => void;
</script>

{#if isOpen}
	<div class="fixed inset-0 z-50 flex justify-end">
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
			onclick={onClose}
			onkeydown={(e) => e.key === 'Escape' && onClose()}
			role="button"
			tabindex="0"
			aria-label="Close drawer"
		></div>

		<!-- Drawer Content -->
		<div
			class="relative w-full max-w-2xl bg-slate-900 border-l border-slate-300 dark:border-slate-700 shadow-2xl h-full flex flex-col transform transition-transform duration-300 ease-in-out"
		>
			<div
				class="flex items-center justify-between px-6 py-4 border-b border-slate-300 dark:border-slate-700 bg-slate-800/50"
			>
				<h2 class="text-xl font-bold text-slate-100">{title}</h2>
				<button
					onclick={onClose}
					class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white transition-colors rounded-lg hover:bg-slate-700"
				>
					✕
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-6">
				<slot />
			</div>
		</div>
	</div>
{/if}
