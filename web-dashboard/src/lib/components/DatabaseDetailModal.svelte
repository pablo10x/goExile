<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { Database, HardDrive, Activity, Archive, Server, Clock } from 'lucide-svelte';

	export let stats: any;
	export let isOpen: boolean = false;
	export let x: number = 0;
	export let y: number = 0;
</script>

{#if isOpen}
	<div
		class="fixed z-50 w-96 p-4 bg-slate-900/95 backdrop-blur-md border border-slate-700/50 rounded-xl shadow-2xl pointer-events-none"
		style="top: {y + 20}px; left: {x - 192}px;"
		transition:scale={{ duration: 200, start: 0.95 }}
	>
		<div class="flex items-center gap-3 mb-4 border-b border-slate-700/50 pb-3">
			<div class="p-2 bg-emerald-500/10 rounded-lg">
				<Database class="w-5 h-5 text-emerald-400" />
			</div>
			<div>
				<h3 class="font-bold text-slate-100">PostgreSQL Metrics</h3>
				<div class="text-xs text-slate-400 flex items-center gap-2">
					<span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
					Live Connection Pool
				</div>
			</div>
		</div>

		<div class="space-y-4">
			<!-- Pool Stats -->
			<div class="grid grid-cols-2 gap-3">
				<div class="bg-slate-800/50 p-3 rounded-lg border border-slate-700/30">
					<div class="text-xs text-slate-400 mb-1">Open / Max</div>
					<div class="text-lg font-bold text-slate-100 font-mono">
						{stats.db_open_connections} <span class="text-slate-500 text-sm">/ 10</span>
					</div>
				</div>
				<div class="bg-slate-800/50 p-3 rounded-lg border border-slate-700/30">
					<div class="text-xs text-slate-400 mb-1">In Use / Idle</div>
					<div class="text-lg font-bold text-emerald-400 font-mono">
						{stats.db_in_use} <span class="text-slate-500 text-sm">/ {stats.db_idle}</span>
					</div>
				</div>
			</div>

			<!-- Wait Stats -->
			<div
				class="bg-slate-800/50 p-3 rounded-lg border border-slate-700/30 flex items-center justify-between"
			>
				<div>
					<div class="text-xs text-slate-400 mb-1">Wait Duration</div>
					<div class="text-sm font-bold text-slate-200">{stats.db_wait_duration}</div>
				</div>
				<div class="text-right">
					<div class="text-xs text-slate-400 mb-1">Wait Count</div>
					<div class="text-sm font-bold text-slate-200">{stats.db_wait_count}</div>
				</div>
			</div>

			<div class="h-px bg-slate-700/50 my-2"></div>

			<!-- Advanced Stats -->
			<div class="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
				<div class="flex justify-between items-center">
					<span class="text-slate-400 flex items-center gap-1.5"
						><HardDrive class="w-3 h-3" /> DB Size</span
					>
					<span class="font-mono text-slate-200">{stats.db_size || 'N/A'}</span>
				</div>
				<div class="flex justify-between items-center">
					<span class="text-slate-400 flex items-center gap-1.5"
						><Activity class="w-3 h-3" /> Cache Hit</span
					>
					<span class="font-mono text-emerald-400"
						>{stats.db_cache_hit ? stats.db_cache_hit.toFixed(2) : 0}%</span
					>
				</div>
				<div class="flex justify-between items-center">
					<span class="text-slate-400 flex items-center gap-1.5"
						><Server class="w-3 h-3" /> Commits</span
					>
					<span class="font-mono text-slate-200">{stats.db_commits || 0}</span>
				</div>
				<div class="flex justify-between items-center">
					<span class="text-slate-400 flex items-center gap-1.5"
						><Archive class="w-3 h-3" /> Rollbacks</span
					>
					<span class="font-mono text-red-400">{stats.db_rollbacks || 0}</span>
				</div>
			</div>

			<!-- Tuple Stats -->
			<div
				class="bg-slate-800/30 p-2 rounded border border-slate-700/30 grid grid-cols-4 gap-1 text-center text-xs mt-2"
			>
				<div>
					<div class="text-slate-500 mb-0.5">Fetch</div>
					<div class="font-mono text-slate-300">{stats.db_tup_fetched || 0}</div>
				</div>
				<div>
					<div class="text-slate-500 mb-0.5">Ins</div>
					<div class="font-mono text-green-400">{stats.db_tup_inserted || 0}</div>
				</div>
				<div>
					<div class="text-slate-500 mb-0.5">Upd</div>
					<div class="font-mono text-blue-400">{stats.db_tup_updated || 0}</div>
				</div>
				<div>
					<div class="text-slate-500 mb-0.5">Del</div>
					<div class="font-mono text-red-400">{stats.db_tup_deleted || 0}</div>
				</div>
			</div>
		</div>
	</div>
{/if}
