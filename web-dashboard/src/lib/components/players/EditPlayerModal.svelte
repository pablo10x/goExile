<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Save, User, Smartphone, Trophy, Hash } from 'lucide-svelte';
	import { notifications } from '$lib/stores';

	let { isOpen, player, onClose, onSave } = $props();

	let formData = $state({
		name: '',
		uid: '',
		device_id: '',
		xp: 0
	});

	let isSaving = $state(false);

	$effect(() => {
		if (player) {
			formData = {
				name: player.name || '',
				uid: player.uid || '',
				device_id: player.device_id || '',
				xp: player.xp || 0
			};
		}
	});

	async function handleSave() {
		if (!player) return;
		isSaving = true;
		try {
			const res = await fetch(`/api/admin/players/${player.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formData)
			});

			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Failed to update player');
			}

			const updatedPlayer = await res.json();
			notifications.add({ type: 'success', message: 'Player updated successfully' });
			onSave(updatedPlayer);
			onClose();
		} catch (e: any) {
			console.error(e);
			notifications.add({ type: 'error', message: e.message });
		} finally {
			isSaving = false;
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		transition:fade={{ duration: 200 }}
		onclick={onClose}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Escape' && onClose()}
	>
		<div
			class="relative w-full max-w-lg bg-slate-900/90 border border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden backdrop-blur-xl"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 300, easing: cubicOut, start: 0.95 }}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.stopPropagation()}
		>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-slate-700/50 flex justify-between items-center bg-slate-800/30">
				<h2 class="text-xl font-bold text-white flex items-center gap-2">
					<User class="w-5 h-5 text-blue-400" />
					Edit Player
				</h2>
				<button
					onclick={onClose}
					class="text-slate-400 hover:text-white transition-colors p-1 rounded-lg hover:bg-white/10"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-6 space-y-5">
				<!-- Internal ID (Read Only) -->
				<div class="space-y-1.5 opacity-60">
					<label class="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
						<Hash class="w-3 h-3" /> Internal ID
					</label>
					<div class="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-3 text-slate-400 font-mono text-sm">
						UID: {player?.id}
					</div>
				</div>

				<!-- Name -->
				<div class="space-y-1.5">
					<label class="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
						<User class="w-3 h-3" /> Display Name
					</label>
					<input
						type="text"
						bind:value={formData.name}
						class="w-full bg-slate-950/50 border border-slate-700 rounded-xl px-4 py-3 text-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 outline-none transition-all placeholder:text-slate-600"
						placeholder="Player Name"
					/>
				</div>

				<!-- Firebase UID -->
				<div class="space-y-1.5">
					<label class="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
						<Hash class="w-3 h-3" /> Firebase UID
					</label>
					<input
						type="text"
						bind:value={formData.uid}
						class="w-full bg-slate-950/50 border border-slate-700 rounded-xl px-4 py-3 text-slate-200 font-mono text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 outline-none transition-all placeholder:text-slate-600"
						placeholder="Firebase UID"
					/>
				</div>

				<!-- Device ID -->
				<div class="space-y-1.5">
					<label class="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
						<Smartphone class="w-3 h-3" /> Device ID
					</label>
					<input
						type="text"
						bind:value={formData.device_id}
						class="w-full bg-slate-950/50 border border-slate-700 rounded-xl px-4 py-3 text-slate-200 font-mono text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 outline-none transition-all placeholder:text-slate-600"
						placeholder="Device ID"
					/>
				</div>

				<!-- XP -->
				<div class="space-y-1.5">
					<label class="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
						<Trophy class="w-3 h-3" /> Experience Points
					</label>
					<input
						type="number"
						bind:value={formData.xp}
						class="w-full bg-slate-950/50 border border-slate-700 rounded-xl px-4 py-3 text-slate-200 font-mono text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500/50 outline-none transition-all placeholder:text-slate-600"
						placeholder="0"
					/>
				</div>
			</div>

			<!-- Footer -->
			<div class="p-6 border-t border-slate-700/50 bg-slate-800/30 flex justify-end gap-3">
				<button
					onclick={onClose}
					class="px-5 py-2.5 rounded-xl text-sm font-medium text-slate-400 hover:text-white hover:bg-white/5 transition-all"
				>
					Cancel
				</button>
				<button
					onclick={handleSave}
					disabled={isSaving}
					class="px-6 py-2.5 rounded-xl text-sm font-bold bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
				>
					{#if isSaving}
						<div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
						Saving...
					{:else}
						<Save class="w-4 h-4" />
						Save Changes
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
