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
		class="fixed z-50 w-96 p-6 bg-slate-900/80 backdrop-blur-2xl border border-stone-800 rounded-none shadow-[0_0_50px_rgba(0,0,0,0.8)] pointer-events-none industrial-frame"
		style="top: {y + 20}px; left: {x - 192}px;"
		transition:scale={{ duration: 200, start: 0.95 }}
	>
		<div
			class="flex items-center gap-4 mb-6 border-b border-stone-800 pb-4"
		>
			<div class="p-2.5 bg-success/10 border border-success/30 industrial-frame">
				<Database class="w-5 h-5 text-success" />
			</div>
			<div>
				<h3 class="font-heading font-black text-white uppercase tracking-widest text-sm">Persistence_Cores</h3>
				<div class="text-[9px] text-text-dim font-jetbrains font-bold flex items-center gap-2 uppercase tracking-tight">
					<span class="w-1.5 h-1.5 rounded-full bg-success animate-pulse shadow-[0_0_8px_var(--color-success)]"></span>
					Live_Pool_Synchronized
				</div>
			</div>
		</div>

		<div class="space-y-6">
			<!-- Pool Stats -->
			<div class="grid grid-cols-2 gap-4">
				<div
					class="bg-stone-900/40 p-4 border border-stone-800 shadow-inner"
				>
					<div class="text-[8px] font-jetbrains font-black text-text-dim mb-2 uppercase tracking-widest">Active / Cap</div>
					<div class="text-xl font-heading font-black text-white tracking-tighter">
						{stats.db_open_connections} <span class="text-stone-700 text-xs font-jetbrains">/ 10</span>
					</div>
				</div>
				<div
					class="bg-stone-900/40 p-4 border border-stone-800 shadow-inner"
				>
					<div class="text-[8px] font-jetbrains font-black text-text-dim mb-2 uppercase tracking-widest">Load / Idle</div>
					<div class="text-xl font-heading font-black text-success tracking-tighter">
						{stats.db_in_use} <span class="text-stone-700 text-xs font-jetbrains">/ {stats.db_idle}</span>
					</div>
				</div>
			</div>

			<!-- Wait Stats -->
			<div
				class="bg-stone-900/40 p-4 border border-stone-800 flex items-center justify-between shadow-inner"
			>
				<div>
					<div class="text-[8px] font-jetbrains font-black text-text-dim mb-1 uppercase tracking-widest">Buffer_Latency</div>
					<div class="text-xs font-jetbrains font-black text-rust-light tracking-tight">
						{stats.db_wait_duration}
					</div>
				</div>
				<div class="text-right">
					<div class="text-[8px] font-jetbrains font-black text-text-dim mb-1 uppercase tracking-widest">Queue_Wait</div>
					<div class="text-xs font-jetbrains font-black text-stone-300">
						{stats.db_wait_count}
					</div>
				</div>
			</div>

			<div class="h-px bg-stone-800/50 my-2"></div>

			<!-- Advanced Stats -->
			<div class="grid grid-cols-2 gap-x-6 gap-y-4 text-[10px]">
				<div class="flex justify-between items-center group">
					<span class="text-text-dim flex items-center gap-2 uppercase font-jetbrains font-bold"
						><HardDrive class="w-3.5 h-3.5" /> Storage</span>
					<span class="font-jetbrains font-black text-stone-300">{stats.db_size || 'NULL'}</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-text-dim flex items-center gap-2 uppercase font-jetbrains font-bold"
						><Activity class="w-3.5 h-3.5 text-rust" /> Hit_Rate</span>
					<span class="font-jetbrains font-black text-success"
						>{stats.db_cache_hit ? stats.db_cache_hit?.toFixed(2) : 0}%</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-text-dim flex items-center gap-2 uppercase font-jetbrains font-bold"
						><Server class="w-3.5 h-3.5" /> Commits</span>
					<span class="font-jetbrains font-black text-stone-300">{stats.db_commits || 0}</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-text-dim flex items-center gap-2 uppercase font-jetbrains font-bold"
						><Archive class="w-3.5 h-3.5 text-danger" /> Rejects</span>
					<span class="font-jetbrains font-black text-danger">{stats.db_rollbacks || 0}</span>
				</div>
			</div>

			<!-- Tuple Stats -->
			<div
				class="bg-stone-950 p-3 border border-stone-800 grid grid-cols-4 gap-2 text-center text-[8px] font-jetbrains font-black uppercase mt-2 shadow-inner"
			>
				<div>
					<div class="text-stone-700 mb-1">In</div>
					<div class="text-stone-400">
						{stats.db_tup_fetched || 0}
					</div>
				</div>
				<div>
					<div class="text-stone-700 mb-1">Add</div>
					<div class="text-success">{stats.db_tup_inserted || 0}</div>
				</div>
				<div>
					<div class="text-stone-700 mb-1">Mod</div>
					<div class="text-rust">{stats.db_tup_updated || 0}</div>
				</div>
				<div>
					<div class="text-stone-700 mb-1">Del</div>
					<div class="text-danger">{stats.db_tup_deleted || 0}</div>
				</div>
			</div>
		</div>
	</div>
{/if}
