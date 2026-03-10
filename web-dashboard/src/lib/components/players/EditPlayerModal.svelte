<script lang="ts">
	import { X, Save, User, Smartphone, Hash, Trophy, RefreshCw, ChevronRight, Terminal } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { apiFetch, notify } from '$lib/api';
	import { notifications } from '$lib/stores.svelte';
	import type { Player } from '$lib/stores.svelte';
	import Button from '../Button.svelte';

	let { player, isOpen = $bindable(), onSave = () => {}, onClose = () => {} } = $props<{
		player: Player | null;
		isOpen: boolean;
		onSave: (player: Player) => void;
		onClose?: () => void;
	}>();

	let isSaving = $state(false);
	let formData = $state({
		name: '',
		uid: '',
		device_id: '',
		xp: 0
	});

	$effect(() => {
		if (player && isOpen) {
			formData = {
				name: player.name || '',
				uid: player.uid || '',
				device_id: player.device_id || '',
				xp: player.xp || 0
			};
		}
	});

	function handleClose() {
		isOpen = false;
		onClose();
	}

	async function handleSave() {
		if (!player) return;
		isSaving = true;
		try {
			const res = await apiFetch(`/api/admin/players/${player.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formData)
			});

			if (res.ok) {
				const updated = await res.json();
				onSave(updated);
				notifications.add({ type: 'success', message: 'User updated successfully' });
				onClose();
			} else {
				const e = await res.json();
				notifications.add({ type: 'error', message: `Error: ${e.error || 'Update failed'}` });
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: `Error: ${e.message}` });
		} finally {
			isSaving = false;
		}
	}
</script>

{#if isOpen && player}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
		onclick={handleClose}
		transition:fade={{ duration: 150 }}
	>
		<div
			class="w-full max-w-lg bg-neutral-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden relative"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<!-- Header -->
			<div
				class="px-8 py-6 border-b border-white/5 flex justify-between items-center bg-neutral-950/40 relative overflow-hidden group"
			>
				<h2
					class="text-2xl font-bold tracking-tight text-white uppercase flex items-center gap-4 relative z-10"
				>
					<User class="w-6 h-6 text-indigo-400" />
					Edit User <span class="text-indigo-400 font-mono">#{player.id}</span>
				</h2>
				<button
					onclick={handleClose}
					class="text-neutral-500 hover:text-white transition-all p-2 rounded-lg hover:bg-white/5 relative z-10"
				>
					<X class="w-6 h-6" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 space-y-6 relative z-10">
				<!-- Form Fields -->
				<div class="space-y-6">
					<!-- Name -->
					<div class="space-y-2">
						<label class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest flex items-center gap-2">
							<User class="w-3 h-3" /> Display Name
						</label>
						<input
							type="text"
							bind:value={formData.name}
							class="w-full bg-black border border-white/10 px-4 py-3 text-white font-bold rounded-xl focus:border-indigo-500 outline-none transition-all placeholder:text-neutral-800"
							placeholder="Username"
						/>
					</div>

					<!-- Firebase UID -->
					<div class="space-y-2">
						<label class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest flex items-center gap-2">
							<Hash class="w-3 h-3" /> Account UID
						</label>
						<input
							type="text"
							bind:value={formData.uid}
							class="w-full bg-black border border-white/10 px-4 py-3 text-neutral-400 font-mono text-sm rounded-xl focus:border-indigo-500 outline-none transition-all"
							placeholder="Firebase UID"
						/>
					</div>

					<!-- Device ID -->
					<div class="space-y-2">
						<label class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest flex items-center gap-2">
							<Smartphone class="w-3 h-3" /> Device ID
						</label>
						<input
							type="text"
							bind:value={formData.device_id}
							class="w-full bg-black border border-white/10 px-4 py-3 text-neutral-400 font-mono text-sm rounded-xl focus:border-indigo-500 outline-none transition-all"
							placeholder="Hardware ID"
						/>
					</div>

					<!-- XP -->
					<div class="space-y-2">
						<label class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest flex items-center gap-2">
							<Trophy class="w-3 h-3 text-amber-500" /> Player Experience (XP)
						</label>
						<input
							type="number"
							bind:value={formData.xp}
							class="w-full bg-black border border-white/10 px-4 py-3 text-amber-500 font-mono font-bold text-xl rounded-xl focus:border-indigo-500 outline-none transition-all"
						/>
					</div>
				</div>

				<div class="p-4 bg-amber-500/5 border border-amber-500/10 rounded-xl text-[10px] text-amber-500/80 font-bold uppercase tracking-widest leading-relaxed">
					Warning: Changes are permanent and logged for administrative review.
				</div>
			</div>

			<!-- Footer -->
			<div class="p-8 border-t border-white/5 bg-neutral-950/40 flex justify-end gap-4">
				<button
					onclick={onClose}
					class="px-6 py-2.5 text-xs font-bold text-neutral-500 hover:text-white uppercase tracking-widest transition-all"
				>
					Cancel
				</button>
				<button
					onclick={handleSave}
					disabled={isSaving}
					class="px-8 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold uppercase tracking-widest rounded-xl transition-all disabled:opacity-50 shadow-lg shadow-indigo-500/20"
				>
					{#if isSaving}
						<div class="flex items-center gap-2">
							<RefreshCw class="w-4 h-4 animate-spin" />
							Saving...
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<Save class="w-4 h-4" />
							Save Changes
						</div>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
