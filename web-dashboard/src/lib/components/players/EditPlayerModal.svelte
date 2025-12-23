<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Save, User, Smartphone, Trophy, Hash, Terminal, ChevronRight } from 'lucide-svelte';
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
		class="fixed inset-0 z-[300] flex items-center justify-center p-4 bg-black/95 backdrop-blur-sm font-['JetBrains_Mono',monospace]"
		transition:fade={{ duration: 200 }}
		onclick={onClose}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Escape' && onClose()}
	>
		<div
			class="relative w-full max-w-xl bg-black border-2 border-white/10 shadow-[0_0_50px_rgba(0,0,0,1)] overflow-hidden"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 300, easing: cubicOut, start: 0.98 }}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.stopPropagation()}
		>
			<!-- CRT Scanline Effect -->
			<div class="absolute inset-0 pointer-events-none opacity-[0.03] bg-amber-scanlines z-50"></div>

			<!-- Header -->
			<div
				class="px-8 py-6 border-b-2 border-[#f97316]/30 flex justify-between items-center bg-black relative overflow-hidden group"
			>
				<div class="absolute inset-0 bg-[#f97316]/5 animate-pulse"></div>
				<h2 class="text-2xl font-black text-white italic tracking-tighter uppercase flex items-center gap-4 relative z-10">
					<Terminal class="w-6 h-6 text-[#f97316]" />
					Mod_Subject_<span class="text-[#f97316]">0x{player?.id.toString(16).toUpperCase()}</span>
				</h2>
				<button
					onclick={onClose}
					class="text-slate-600 hover:text-[#f97316] transition-all p-2 border border-transparent hover:border-[#f97316]/30 relative z-10"
				>
					<X class="w-6 h-6" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-10 space-y-8 relative z-10 bg-[#050505]">
				<!-- Identification Header -->
				<div class="flex items-center gap-4 text-[10px] font-black text-slate-600 tracking-[0.4em] uppercase italic border-b border-white/5 pb-4">
					<ChevronRight class="w-3 h-3 text-[#f97316]" />
					Field_Modification_Buffer
				</div>

				<!-- Form Fields -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
					<!-- Name -->
					<div class="space-y-3">
						<label
							class="text-[10px] font-black text-[#f97316]/60 uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<User class="w-3 h-3" /> Ident_Subject
						</label>
						<input
							type="text"
							bind:value={formData.name}
							class="w-full bg-black border-b-2 border-white/5 px-0 py-3 text-white font-bold italic text-lg focus:border-[#f97316] outline-none transition-all placeholder:text-slate-900"
							placeholder="NULL_PTR"
						/>
					</div>

					<!-- Firebase UID -->
					<div class="space-y-3 text-slate-500">
						<label
							class="text-[10px] font-black text-slate-600 uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Hash class="w-3 h-3" /> Core_UID
						</label>
						<input
							type="text"
							bind:value={formData.uid}
							class="w-full bg-black border-b-2 border-white/5 px-0 py-3 text-slate-400 font-bold tracking-tighter text-sm focus:border-[#f97316] outline-none transition-all placeholder:text-slate-900"
							placeholder="0x000...000"
						/>
					</div>

					<!-- Device ID -->
					<div class="space-y-3 md:col-span-2">
						<label
							class="text-[10px] font-black text-slate-600 uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Smartphone class="w-3 h-3" /> Node_Signature
						</label>
						<input
							type="text"
							bind:value={formData.device_id}
							class="w-full bg-black border-b-2 border-white/5 px-0 py-3 text-slate-400 font-bold tracking-tighter text-sm focus:border-[#f97316] outline-none transition-all placeholder:text-slate-900"
							placeholder="UNK_HARDWARE_SIG"
						/>
					</div>

					<!-- XP -->
					<div class="space-y-3">
						<label
							class="text-[10px] font-black text-slate-600 uppercase tracking-[0.3em] flex items-center gap-3 italic"
						>
							<Trophy class="w-3 h-3 text-[#fbbf24]" /> Exp_Accumulation
						</label>
						<input
							type="number"
							bind:value={formData.xp}
							class="w-full bg-black border-b-2 border-white/5 px-0 py-3 text-[#fbbf24] font-black italic text-2xl focus:border-[#f97316] outline-none transition-all placeholder:text-slate-900"
							placeholder="0"
						/>
					</div>
				</div>

				<!-- Warning Area -->
				<div class="mt-4 p-4 bg-[#ef4444]/5 border-l-4 border-[#ef4444]/30 text-[10px] text-[#ef4444]/60 font-bold uppercase italic tracking-widest leading-relaxed">
					Warning: Direct database mutation detected. All write operations are logged to the central surveillance engine.
				</div>
			</div>

			<!-- Footer Commands -->
			<div class="p-8 border-t-2 border-white/5 bg-black flex justify-between items-center relative overflow-hidden">
				<div class="text-[9px] font-black text-slate-800 tracking-[0.5em] uppercase italic">
					Waiting_For_Acknowledge
				</div>
				<div class="flex gap-6">
					<button
						onclick={onClose}
						class="px-8 py-3 text-[11px] font-black text-slate-600 hover:text-white uppercase tracking-widest italic transition-all border border-transparent hover:border-white/10"
					>
						[Abort]
					</button>
					<button
						onclick={handleSave}
						disabled={isSaving}
						class="px-10 py-4 text-[11px] font-black bg-[#f97316] hover:bg-white text-black uppercase tracking-[0.3em] italic transition-all disabled:opacity-20 shadow-[8px_8px_0px_#000] relative group overflow-hidden"
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