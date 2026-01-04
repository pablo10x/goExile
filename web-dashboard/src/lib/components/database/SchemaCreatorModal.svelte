<script lang="ts">
	import { Save, X, Database, User, Shield, Check } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';

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
		class="fixed inset-0 bg-black/80 backdrop-blur-md z-[60] flex items-center justify-center p-4"
		transition:fade={{ duration: 200 }}
	>
		<div
			class="bg-neutral-900/80 backdrop-blur-2xl border border-stone-800 rounded-none w-full max-w-md shadow-2xl flex flex-col overflow-hidden relative group industrial-frame"
			transition:scale={{ duration: 300, start: 0.95, easing: quintOut }}
		>
			<!-- Ambient Background Glow -->
			<div
				class="absolute -top-20 -right-20 w-64 h-64 bg-rust/10 rounded-full blur-[80px] pointer-events-none group-hover:bg-rust/20 transition-all duration-700"
			></div>

			<!-- Header -->
			<div
				class="p-6 border-b border-stone-800 relative z-10 flex justify-between items-start bg-neutral-950/40"
			>
				<div>
					<h3
						class="text-xl font-heading font-black text-neutral-100 uppercase tracking-tighter flex items-center gap-3"
					>
						<div
							class="p-2 bg-rust/10 border border-rust/20 rounded-none industrial-frame"
						>
							<Database class="w-5 h-5 text-rust-light" />
						</div>
						Initialize_Schema
					</h3>
					<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest mt-2">
						Define a new namespace for database objects.
					</p>
				</div>
				<button
					onclick={onClose}
					class="text-text-dim hover:text-white p-2 transition-all hover:rotate-90 duration-300"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 space-y-8 relative z-10">
				<!-- Schema Name Input -->
				<div class="space-y-3">
					<label
						class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest flex items-center gap-2"
					>
						<Database class="w-3.5 h-3.5 text-rust" /> Schema_Identity
					</label>
					<div class="relative">
						<input
							type="text"
							bind:value={schemaName}
							class="w-full bg-stone-950 border border-stone-800 py-3 px-4 pl-12 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none uppercase tracking-widest transition-all"
							placeholder="e.g. CORE_MODULE_V1"
						/>
						<div
							class="absolute left-4 top-1/2 -tranneutral-y-1/2 text-stone-700 group-focus-within:text-rust transition-colors"
						>
							<Shield class="w-4 h-4" />
						</div>
					</div>
				</div>

				<!-- Owner Select -->
				<div class="space-y-3">
					<label
						class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest flex items-center gap-2"
					>
						<User class="w-3.5 h-3.5 text-rust" /> Sector_Owner
					</label>
					<div class="relative">
						<select
							bind:value={owner}
							class="w-full bg-stone-950 border border-stone-800 py-3 px-4 pl-12 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none appearance-none cursor-pointer uppercase tracking-widest"
						>
							{#each roles as role}
								<option value={role} class="bg-stone-900">{role}</option>
							{/each}
						</select>
						<div
							class="absolute left-4 top-1/2 -tranneutral-y-1/2 text-stone-700"
						>
							<User class="w-4 h-4" />
						</div>
						<!-- Custom Chevron -->
						<div class="absolute right-4 top-1/2 -tranneutral-y-1/2 text-stone-700 pointer-events-none">
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
					<p class="font-jetbrains text-[9px] text-text-dim uppercase tracking-widest pl-1">Authorized identity that will govern the schema sector.</p>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="p-6 border-t border-stone-800 bg-neutral-950/40 flex justify-end gap-4 relative z-10"
			>
				<button
					onclick={onClose}
					class="px-6 py-3 font-heading font-black text-[10px] text-text-dim hover:text-white transition-all uppercase tracking-widest"
				>
					Abort_Sequence
				</button>
				<button
					onclick={handleSubmit}
					disabled={loading || !schemaName.trim()}
					class="px-8 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 disabled:opacity-20 transition-all active:tranneutral-y-px"
				>
					{#if loading}
						<div
							class="w-4 h-4 border-2 border-white/30 border-t-white animate-spin"
						></div>
						Initializing...
					{:else}
						<Check class="w-4 h-4" />
						Authorize_Init
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
