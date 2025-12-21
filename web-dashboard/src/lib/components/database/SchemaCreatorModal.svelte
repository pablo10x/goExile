<script lang="ts">
	import { Save, X, Database, User, Shield, Check } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores';

	let {
		isOpen = $bindable(false),
		onClose,
		onSave
	} = $props<{
		isOpen: boolean;
		onClose: () => void;
		onSave: (data: { name: string; owner: string }) => Promise<void>;
	}>();

	let schemaName = $state('');
	let owner = $state('');
	let loading = $state(false);
	let roles = $state<string[]>([]);

	async function loadRoles() {
		try {
			const res = await fetch('/api/database/roles');
			if (res.ok) {
				const data = await res.json();
				roles = data.map((r: any) => r.name);
				if (roles.includes('postgres')) owner = 'postgres';
				else if (roles.length > 0) owner = roles[0];
			}
		} catch (e) {
			console.error('Failed to load roles', e);
		}
	}

	$effect(() => {
		if (isOpen) {
			loadRoles();
			schemaName = '';
		}
	});

	async function handleSubmit() {
		if (!schemaName.trim()) return;
		loading = true;
		try {
			await onSave({ name: schemaName, owner });
			onClose();
		} finally {
			loading = false;
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 bg-black/60 backdrop-blur-md z-[60] flex items-center justify-center p-4 perspective-1000"
		transition:fade={{ duration: 200 }}
	>
		<div
			class="bg-slate-900/90 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl w-full max-w-md shadow-2xl flex flex-col overflow-hidden relative group"
			transition:scale={{ duration: 300, start: 0.95, easing: quintOut }}
		>
			<!-- Ambient Background Glow -->
			<div
				class="absolute -top-20 -right-20 w-64 h-64 bg-emerald-500/10 rounded-full blur-[80px] pointer-events-none group-hover:bg-emerald-500/20 transition-all duration-700"
			></div>
			<div
				class="absolute -bottom-20 -left-20 w-64 h-64 bg-blue-500/10 rounded-full blur-[80px] pointer-events-none group-hover:bg-blue-500/20 transition-all duration-700"
			></div>

			<!-- Header -->
			<div
				class="p-6 pb-4 border-b border-slate-200/50 dark:border-slate-800/50 relative z-10 flex justify-between items-start"
			>
				<div>
					<h3
						class="text-xl font-bold text-slate-900 dark:text-white tracking-tight flex items-center gap-2"
					>
						<div
							class="p-2 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-lg shadow-lg shadow-emerald-500/20"
						>
							<Database class="w-5 h-5 text-slate-900 dark:text-white" />
						</div>
						Create New Schema
					</h3>
					<p class="text-slate-500 dark:text-slate-400 text-sm mt-2">
						Define a new namespace for your database objects.
					</p>
				</div>
				<button
					onclick={onClose}
					class="text-slate-500 hover:text-slate-900 dark:text-white hover:bg-slate-800/50 p-2 rounded-full transition-all hover:rotate-90 duration-300"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-6 space-y-6 relative z-10">
				<!-- Schema Name Input -->
				<div class="space-y-2 group/input">
					<label
						class="text-xs font-bold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5"
					>
						<Database class="w-3.5 h-3.5" /> Schema Name
					</label>
					<div class="relative">
						<input
							type="text"
							bind:value={schemaName}
							class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300 dark:border-slate-700 rounded-xl px-4 py-3 pl-11 text-slate-800 dark:text-slate-200 outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500/50 transition-all placeholder:text-slate-600"
							placeholder="e.g. analytics_v2"
						/>
						<div
							class="absolute left-3.5 top-3.5 text-slate-500 group-focus-within/input:text-emerald-500 transition-colors"
						>
							<Shield class="w-4 h-4" />
						</div>
					</div>
				</div>

				<!-- Owner Select -->
				<div class="space-y-2 group/select">
					<label
						class="text-xs font-bold text-blue-400 uppercase tracking-wider flex items-center gap-1.5"
					>
						<User class="w-3.5 h-3.5" /> Owner
					</label>
					<div class="relative">
						<select
							bind:value={owner}
							class="w-full bg-white/50 dark:bg-slate-950/50 border border-slate-300 dark:border-slate-700 rounded-xl px-4 py-3 pl-11 text-slate-800 dark:text-slate-200 outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 transition-all appearance-none cursor-pointer hover:bg-slate-900"
						>
							{#each roles as role}
								<option value={role} class="bg-slate-900">{role}</option>
							{/each}
						</select>
						<div
							class="absolute left-3.5 top-3.5 text-slate-500 group-focus-within/select:text-blue-500 transition-colors"
						>
							<User class="w-4 h-4" />
						</div>
						<!-- Custom Chevron -->
						<div class="absolute right-4 top-3.5 text-slate-500 pointer-events-none">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								width="16"
								height="16"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"><path d="m6 9 6 6 6-6" /></svg
							>
						</div>
					</div>
					<p class="text-xs text-slate-500 pl-1">Role that will own the schema and its objects.</p>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="p-6 pt-2 border-t border-slate-200/50 dark:border-slate-800/50 bg-white/30 dark:bg-slate-950/30 flex justify-end gap-3 relative z-10"
			>
				<button
					onclick={onClose}
					class="px-5 py-2.5 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-800/80 rounded-xl transition-all font-medium text-sm"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={loading || !schemaName.trim()}
					class="px-6 py-2.5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-slate-900 dark:text-white rounded-xl font-bold flex items-center gap-2 shadow-lg shadow-emerald-900/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all transform hover:scale-[1.02] active:scale-[0.98]"
				>
					{#if loading}
						<div
							class="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin"
						></div>
						Creating...
					{:else}
						<Check class="w-4 h-4" />
						Create Schema
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.perspective-1000 {
		perspective: 1000px;
	}
</style>
