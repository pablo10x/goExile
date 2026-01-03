<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Save, User, Smartphone, Trophy, Hash, Terminal, ChevronRight, RefreshCw } from 'lucide-svelte';
	import { notifications, siteSettings } from '$lib/stores.svelte';

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
				throw new Error(err.error || 'ERR_WRITE_FAILED');
			}

			const updatedPlayer = await res.json();
			notifications.add({ type: 'success', message: 'UPLINK_SUCCESS: Subject updated' });
			onSave(updatedPlayer);
			onClose();
		} catch (e: any) {
			console.error(e);
			notifications.add({ type: 'error', message: `UPLINK_ERROR: ${e.message}` });
		} finally {
			isSaving = false;
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-[300] flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-md font-['JetBrains_Mono',monospace]"
		transition:fade={{ duration: 200 }}
		onclick={onClose}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Escape' && onClose()}
	>
		        <div
		            class="w-full max-w-lg bg-slate-900/80 backdrop-blur-2xl border border-stone-800 rounded-none shadow-2xl overflow-hidden relative industrial-sharp"
		            onclick={(e) => e.stopPropagation()}
		            transition:scale={{ duration: 200, start: 0.95 }}
		        >			<!-- Tactical Corners -->
			<div class="corner-tl"></div>
			<div class="corner-tr"></div>
			<div class="corner-bl"></div>
			<div class="corner-br"></div>

			<!-- Header -->
			<div
				class="px-8 py-6 border-b border-stone-800 flex justify-between items-center bg-slate-950/40 relative overflow-hidden group"
			>
				<div class="absolute inset-0 bg-rust/5 animate-pulse"></div>
				<h2 class="text-2xl font-heading font-black italic tracking-tighter text-white uppercase flex items-center gap-4 relative z-10">
					<Terminal class="w-6 h-6 text-rust" />
					Mod_Subject_<span class="text-rust font-mono">0x{player?.id.toString(16).toUpperCase()}</span>
				</h2>
				<button
					onclick={onClose}
					class="text-text-dim hover:text-white transition-all p-2 border border-transparent hover:border-rust/30 relative z-10"
				>
					<X class="w-6 h-6" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-10 space-y-8 relative z-10 bg-transparent">
				<!-- Identification Header -->
				<div class="flex items-center gap-4 text-[10px] font-black text-text-dim tracking-[0.4em] uppercase italic border-b border-stone-800 pb-4">
					<ChevronRight class="w-3 h-3 text-rust" />
					Field_Modification_Buffer
				</div>

				<!-- Form Fields -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
					<!-- Name -->
					<div class="space-y-3">
						<label
							class="text-[10px] font-black text-rust/60 uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<User class="w-3 h-3" /> Ident_Subject
						</label>
						<input
							type="text"
							bind:value={formData.name}
							class="w-full bg-slate-950/40 border border-stone-800 px-4 py-3 text-white font-bold italic text-lg focus:border-rust outline-none transition-all placeholder:text-stone-900 industrial-frame"
							placeholder="NULL_PTR"
						/>
					</div>

					<!-- Firebase UID -->
					<div class="space-y-3">
						<label
							class="text-[10px] font-black text-text-dim uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Hash class="w-3 h-3" /> Core_UID
						</label>
						<input
							type="text"
							bind:value={formData.uid}
							class="w-full bg-slate-950/40 border border-stone-800 px-4 py-3 text-stone-400 font-mono font-bold tracking-tighter text-sm focus:border-rust outline-none transition-all placeholder:text-stone-900 industrial-frame"
							placeholder="0x000...000"
						/>
					</div>

					<!-- Device ID -->
					<div class="space-y-3 md:col-span-2">
						<label
							class="text-[10px] font-black text-text-dim uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Smartphone class="w-3 h-3" /> Node_Signature
						</label>
						<input
							type="text"
							bind:value={formData.device_id}
							class="w-full bg-slate-950/40 border border-stone-800 px-4 py-3 text-stone-400 font-mono font-bold tracking-tighter text-sm focus:border-rust outline-none transition-all placeholder:text-stone-900 industrial-frame"
							placeholder="UNK_HARDWARE_SIG"
						/>
					</div>

					<!-- XP -->
					<div class="space-y-3">
						<label
							class="text-[10px] font-black text-text-dim uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Trophy class="w-3 h-3 text-warning" /> Exp_Accumulation
						</label>
						<input
							type="number"
							bind:value={formData.xp}
							class="w-full bg-slate-950/40 border border-stone-800 px-4 py-3 text-warning font-mono font-bold italic text-2xl focus:border-rust outline-none transition-all placeholder:text-stone-900 industrial-frame"
							placeholder="0"
						/>
					</div>
				</div>

				<!-- Warning Area -->
				<div class="mt-4 p-4 bg-danger/5 border-l-4 border-red-600/30 text-[10px] text-red-600/60 font-bold uppercase italic tracking-widest leading-relaxed">
					Warning: Direct database mutation detected. All write operations are logged to the central surveillance engine.
				</div>
			</div>

			<!-- Footer Commands -->
			<div class="p-8 border-t border-stone-800 bg-slate-950/40 flex justify-between items-center relative overflow-hidden">
				<div class="text-[9px] font-black text-stone-800 tracking-[0.5em] uppercase italic">
					Waiting_For_Acknowledge
				</div>
				<div class="flex gap-6">
					<button
						onclick={onClose}
						class="px-8 py-3 text-[11px] font-black text-text-dim hover:text-white uppercase tracking-widest italic transition-all border border-transparent hover:border-stone-800"
					>
						[Abort]
					</button>
					<button
						onclick={handleSave}
						disabled={isSaving}
						class="px-10 py-4 text-[11px] font-black bg-rust hover:bg-rust-light text-white uppercase tracking-[0.3em] italic transition-all disabled:opacity-20 shadow-xl relative group overflow-hidden"
					>
						{#if isSaving}
							<div class="flex items-center gap-3 animate-pulse">
								<RefreshCw class="w-4 h-4 animate-spin" />
								Writing...
							</div>
						{:else}
							<div class="flex items-center gap-3">
								<Save class="w-4 h-4" />
								COMMIT_CHANGES
							</div>
						{/if}
					</button>
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
</style>