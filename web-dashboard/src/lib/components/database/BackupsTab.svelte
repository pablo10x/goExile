<script lang="ts">
	import { HardDrive, Download, RotateCcw, Trash2, FileText, Plus } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores';
	import { formatBytes } from '$lib/utils';

	let backups = $state<any[]>([]);
	let loading = $state(false);

	async function loadBackups() {
		loading = true;
		try {
			const res = await fetch('/api/database/backups');
			if (res.ok) {
				backups = await res.json();
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to load backups', details: e.message });
		} finally {
			loading = false;
		}
	}

	async function createBackup() {
		loading = true;
		try {
			const res = await fetch('/api/database/backups', { method: 'POST' });
			if (!res.ok) throw new Error('Backup failed');
			notifications.add({ type: 'success', message: 'Backup created' });
			loadBackups();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Backup failed', details: e.message });
			loading = false;
		}
	}

	async function restoreBackup(filename: string) {
		if (
			!confirm(
				`WARNING: This will overwrite the current database with '${filename}'. Are you sure?`
			)
		)
			return;
		loading = true;
		try {
			const res = await fetch('/api/database/restore', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ filename })
			});
			if (!res.ok) throw new Error('Restore failed');
			notifications.add({ type: 'success', message: 'Database restored. Reloading...' });
			setTimeout(() => window.location.reload(), 2000);
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Restore failed', details: e.message });
			loading = false;
		}
	}

	async function deleteBackup(filename: string) {
		if (!confirm(`Delete backup '${filename}'?`)) return;
		try {
			const res = await fetch(`/api/database/backups/${filename}`, { method: 'DELETE' });
			if (!res.ok) throw new Error('Delete failed');
			notifications.add({ type: 'success', message: 'Backup deleted' });
			loadBackups();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Delete failed', details: e.message });
		}
	}

	function downloadFile(filename: string) {
		window.open(`/api/database/backups/${filename}`, '_blank');
	}

	onMount(() => {
		loadBackups();
	});
</script>

<div class="h-full flex flex-col bg-[#050505]">
	<div
		class="p-6 border-b border-stone-800 bg-[#0a0a0a] flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
				<HardDrive class="w-6 h-6 text-rust-light" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-slate-100 uppercase tracking-tighter">ARCHIVE_STORAGE_CORE</h2>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest mt-1">Create and restore internal database snapshots</p>
			</div>
		</div>
					<button
						onclick={createBackup}
						disabled={loading}
						class="px-6 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 disabled:opacity-20 transition-all active:translate-y-px"
					>			<Plus class="w-4 h-4" /> Generate_Snapshot
		</button>
	</div>

	<div class="flex-1 overflow-auto p-8 custom-scrollbar">
		{#if loading && backups.length === 0}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-rust border-t-transparent rounded-none animate-spin"
				></div>
			</div>
		{:else if backups.length === 0}
			<div
				class="flex flex-col items-center justify-center h-64 text-stone-600 border border-dashed border-stone-800 industrial-frame bg-stone-900/20"
			>
				<HardDrive class="w-16 h-16 opacity-10 mb-6" />
				<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">No archives detected</p>
				<p class="font-jetbrains text-[9px] mt-2 uppercase tracking-widest opacity-60">Generate first snapshot to secure sector data</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-4 max-w-5xl mx-auto">
				{#each backups as backup}
					<div
						class="flex items-center justify-between p-6 bg-stone-900/40 border border-stone-800 hover:border-rust/40 transition-all group industrial-frame"
					>
						<div class="flex items-center gap-6">
							<div
								class="p-3 bg-stone-950 border border-stone-800 text-stone-600 group-hover:text-rust transition-colors"
							>
								<FileText class="w-7 h-7" />
							</div>
							<div>
								<div class="font-jetbrains font-black text-stone-200 text-base uppercase tracking-tighter">
									{backup.name}
								</div>
								<div class="font-jetbrains text-[10px] text-stone-500 flex gap-4 mt-2 uppercase tracking-widest">
									<span class="text-rust/60 font-black">{formatBytes(backup.size)}</span>
									<span class="w-1 h-1 bg-stone-800 self-center"></span>
									<span>{new Date(backup.created_at).toLocaleString()}</span>
								</div>
							</div>
						</div>
						<div class="flex gap-3">
							<button
								onclick={() => downloadFile(backup.name)}
								class="p-2.5 text-stone-600 hover:text-white hover:bg-stone-800 transition-all"
								title="Download"
							>
								<Download class="w-5 h-5" />
							</button>
							<button
								onclick={() => restoreBackup(backup.name)}
								class="p-2.5 text-stone-600 hover:text-amber-500 hover:bg-amber-500/5 transition-all"
								title="Restore"
							>
								<RotateCcw class="w-5 h-5" />
							</button>
							<button
								onclick={() => deleteBackup(backup.name)}
								class="p-2.5 text-stone-600 hover:text-red-500 hover:bg-red-500/5 transition-all"
								title="Delete"
							>
								<Trash2 class="w-5 h-5" />
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
