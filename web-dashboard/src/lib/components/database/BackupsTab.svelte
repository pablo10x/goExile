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

<div class="h-full flex flex-col bg-slate-900">
	<div
		class="p-6 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 flex justify-between items-center"
	>
		<div class="flex items-center gap-3">
			<div class="p-2 bg-orange-500/10 rounded-lg">
				<HardDrive class="w-6 h-6 text-orange-400" />
			</div>
			<div>
				<h2 class="text-xl font-bold text-slate-100">Backups & Snapshots</h2>
				<p class="text-sm text-slate-500">Create and restore internal database snapshots</p>
			</div>
		</div>
		<button
			onclick={createBackup}
			disabled={loading}
			class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-slate-900 dark:text-white rounded-lg font-bold flex items-center gap-2 shadow-lg shadow-blue-900/20 disabled:opacity-50"
		>
			<Plus class="w-4 h-4" /> Create Backup
		</button>
	</div>

	<div class="flex-1 overflow-auto p-6">
		{#if loading && backups.length === 0}
			<div class="flex justify-center p-12">
				<div
					class="w-8 h-8 border-4 border-orange-500 border-t-transparent rounded-full animate-spin"
				></div>
			</div>
		{:else if backups.length === 0}
			<div
				class="flex flex-col items-center justify-center h-64 text-slate-500 border-2 border-dashed border-slate-200 dark:border-slate-800 rounded-xl"
			>
				<HardDrive class="w-12 h-12 opacity-20 mb-4" />
				<p>No backups found. Create one to get started.</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-3">
				{#each backups as backup}
					<div
						class="flex items-center justify-between p-4 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl hover:bg-slate-800 hover:border-slate-600 transition-all group"
					>
						<div class="flex items-center gap-4">
							<div
								class="p-3 bg-slate-900 rounded-lg text-slate-500 dark:text-slate-400 group-hover:text-blue-400 transition-colors"
							>
								<FileText class="w-6 h-6" />
							</div>
							<div>
								<div class="font-mono font-bold text-slate-800 dark:text-slate-200 text-lg">
									{backup.name}
								</div>
								<div class="text-xs text-slate-500 flex gap-3 mt-1">
									<span class="bg-slate-900 px-2 py-0.5 rounded text-slate-500 dark:text-slate-400"
										>{formatBytes(backup.size)}</span
									>
									<span>{new Date(backup.created_at).toLocaleString()}</span>
								</div>
							</div>
						</div>
						<div class="flex gap-2">
							<button
								onclick={() => downloadFile(backup.name)}
								class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:text-white hover:bg-slate-700 rounded-lg"
								title="Download"
							>
								<Download class="w-5 h-5" />
							</button>
							<button
								onclick={() => restoreBackup(backup.name)}
								class="p-2 text-orange-400 hover:text-orange-300 hover:bg-orange-500/10 rounded-lg"
								title="Restore"
							>
								<RotateCcw class="w-5 h-5" />
							</button>
							<button
								onclick={() => deleteBackup(backup.name)}
								class="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg"
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
