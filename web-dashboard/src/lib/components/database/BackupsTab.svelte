<script lang="ts">
	import { HardDrive, Download, RotateCcw, Trash2, FileText, Plus } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';
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

<div class="h-full flex flex-col bg-transparent">
	<div
		class="p-6 border-b border-neutral-800 bg-neutral-900/40 backdrop-blur-md flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-indigo-500/10 border border-indigo-500/20 rounded-xl">
				<HardDrive class="w-6 h-6 text-indigo-400" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-white uppercase tracking-tighter italic">ARCHIVE_STORAGE_CORE</h2>
				<p class="font-jetbrains text-[10px] text-neutral-500 uppercase tracking-widest mt-1 italic font-bold">Create and restore internal database snapshots</p>
			</div>
		</div>
					<button
						onclick={createBackup}
						disabled={loading}
						class="px-6 py-3 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 disabled:opacity-20 transition-all active:tranneutral-y-px rounded-xl"
					>			<Plus class="w-4 h-4" /> Generate_Snapshot
		</button>
	</div>

	<div class="flex-1 overflow-auto p-8 custom-scrollbar bg-transparent">
		{#if loading && backups.length === 0}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(99,102,241,0.4)]"
				></div>
			</div>
		{:else if backups.length === 0}
			<div
				class="flex flex-col items-center justify-center h-64 text-neutral-700 border border-dashed border-neutral-800 rounded-2xl bg-neutral-900/20"
			>
				<HardDrive class="w-16 h-16 opacity-10 mb-6" />
				<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">No archives detected</p>
				<p class="font-jetbrains text-[9px] mt-2 uppercase tracking-widest opacity-60">Generate first snapshot to secure sector data</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-4 max-w-5xl mx-auto">
				{#each backups as backup}
					<div
						class="flex items-center justify-between p-6 bg-neutral-900/40 border border-neutral-800 hover:border-indigo-500/40 transition-all group rounded-2xl backdrop-blur-sm shadow-lg"
					>
						<div class="flex items-center gap-6">
							<div
								class="p-3 bg-neutral-950 border border-neutral-800 text-neutral-600 group-hover:text-indigo-400 transition-colors rounded-xl shadow-inner"
							>
								<FileText class="w-7 h-7" />
							</div>
							<div>
								<div class="font-jetbrains font-black text-neutral-200 text-base uppercase tracking-tighter italic">
									{backup.name}
								</div>
								<div class="font-jetbrains text-[10px] text-neutral-500 flex gap-4 mt-2 uppercase tracking-widest font-bold">
									<span class="text-indigo-400 font-black">{formatBytes(backup.size)}</span>
									<span class="w-1 h-1 bg-neutral-800 self-center rounded-full"></span>
									<span>{new Date(backup.created_at).toLocaleString()}</span>
								</div>
							</div>
						</div>
						<div class="flex gap-3">
							<button
								onclick={() => downloadFile(backup.name)}
								class="p-2.5 text-neutral-500 hover:text-white hover:bg-neutral-800 transition-all rounded-lg"
								title="Download"
							>
								<Download class="w-5 h-5" />
							</button>
							<button
								onclick={() => restoreBackup(backup.name)}
								class="p-2.5 text-neutral-500 hover:text-indigo-400 hover:bg-indigo-500/10 transition-all rounded-lg"
								title="Restore"
							>
								<RotateCcw class="w-5 h-5" />
							</button>
							<button
								onclick={() => deleteBackup(backup.name)}
								class="p-2.5 text-neutral-500 hover:text-red-400 hover:bg-red-500/10 transition-all rounded-lg"
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
