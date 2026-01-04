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
		class="fixed z-50 w-96 p-6 bg-neutral-900 border-2 border-neutral-800 rounded-none shadow-[0_0_100px_rgba(0,0,0,1)] pointer-events-none industrial-frame"
		style="top: {y + 20}px; left: {x - 192}px;"
		transition:scale={{ duration: 200, start: 0.95 }}
	>
		<div
			class="flex items-center gap-4 mb-6 border-b-2 border-neutral-800 pb-4"
		>
			<div class="p-2.5 bg-emerald-500/10 border border-emerald-500/30 rounded-none">
				<Database class="w-5 h-5 text-emerald-500" />
			</div>
			<div>
				<h3 class="font-heading font-black text-white uppercase tracking-widest text-sm italic">Persistence_Cores</h3>
				<div class="text-[9px] text-neutral-500 font-mono font-black flex items-center gap-2 uppercase tracking-tight">
					<span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_#10b981]"></span>
					Live_Pool_Synchronized
				</div>
			</div>
		</div>

		<div class="space-y-6">
			<!-- Pool Stats -->
			<div class="grid grid-cols-2 gap-4">
				<div
					class="bg-black border border-neutral-800 p-4 shadow-inner"
				>
					<div class="text-[8px] font-mono font-black text-neutral-600 mb-2 uppercase tracking-widest">Active / Cap</div>
					<div class="text-xl font-heading font-black text-white tracking-tighter">
						{stats.db_open_connections} <span class="text-neutral-700 text-xs font-mono">/ 10</span>
					</div>
				</div>
				<div
					class="bg-black border border-neutral-800 p-4 shadow-inner"
				>
					<div class="text-[8px] font-mono font-black text-neutral-600 mb-2 uppercase tracking-widest">Load / Idle</div>
					<div class="text-xl font-heading font-black text-emerald-500 tracking-tighter">
						{stats.db_in_use} <span class="text-neutral-700 text-xs font-mono">/ {stats.db_idle}</span>
					</div>
				</div>
			</div>

			<!-- Wait Stats -->
			<div
				class="bg-black border border-neutral-800 p-4 flex items-center justify-between shadow-inner"
			>
				<div>
					<div class="text-[8px] font-mono font-black text-neutral-600 mb-1 uppercase tracking-widest">Buffer_Latency</div>
					<div class="text-xs font-mono font-black text-rust-light tracking-tight">
						{stats.db_wait_duration}
					</div>
				</div>
				<div class="text-right">
					<div class="text-[8px] font-mono font-black text-neutral-600 mb-1 uppercase tracking-widest">Queue_Wait</div>
					<div class="text-xs font-mono font-black text-neutral-400">
						{stats.db_wait_count}
					</div>
				</div>
			</div>

			<div class="h-px bg-neutral-800 my-2"></div>

			<!-- Advanced Stats -->
			<div class="grid grid-cols-2 gap-x-6 gap-y-4 text-[10px]">
				<div class="flex justify-between items-center group">
					<span class="text-neutral-500 flex items-center gap-2 uppercase font-mono font-black"
						><HardDrive class="w-3.5 h-3.5" /> Storage</span>
					<span class="font-mono font-black text-neutral-300">{stats.db_size || 'NULL'}</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-neutral-500 flex items-center gap-2 uppercase font-mono font-black"
						><Activity class="w-3.5 h-3.5 text-rust" /> Hit_Rate</span>
					<span class="font-mono font-black text-emerald-500"
						>{stats.db_cache_hit ? stats.db_cache_hit?.toFixed(2) : 0}%</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-neutral-500 flex items-center gap-2 uppercase font-mono font-black"
						><Server class="w-3.5 h-3.5" /> Commits</span>
					<span class="font-mono font-black text-neutral-300">{stats.db_commits || 0}</span>
				</div>
				<div class="flex justify-between items-center group">
					<span class="text-neutral-500 flex items-center gap-2 uppercase font-mono font-black"
						><Archive class="w-3.5 h-3.5 text-red-500" /> Rejects</span>
					<span class="font-mono font-black text-red-500">{stats.db_rollbacks || 0}</span>
				</div>
			</div>

			<!-- Tuple Stats -->
			<div
				class="bg-black p-3 border border-neutral-800 grid grid-cols-4 gap-2 text-center text-[8px] font-mono font-black uppercase mt-2 shadow-inner"
			>
				<div>
					<div class="text-neutral-700 mb-1">In</div>
					<div class="text-neutral-500">
						{stats.db_tup_fetched || 0}
					</div>
				</div>
				<div>
					<div class="text-neutral-700 mb-1">Add</div>
					<div class="text-emerald-500">{stats.db_tup_inserted || 0}</div>
				</div>
				<div>
					<div class="text-neutral-700 mb-1">Mod</div>
					<div class="text-rust">{stats.db_tup_updated || 0}</div>
				</div>
				<div>
					<div class="text-neutral-700 mb-1">Del</div>
					<div class="text-red-500">{stats.db_tup_deleted || 0}</div>
				</div>
			</div>
		</div>
	</div>
{/if}
