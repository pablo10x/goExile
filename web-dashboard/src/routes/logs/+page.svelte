<script lang="ts">
import { apiFetch } from "$lib/api";
	import { onMount } from 'svelte';
	import { fade, slide, fly, scale } from 'svelte/transition';
	import {
		AlertTriangle,
		Shield,
		Server,
		Activity,
		Search,
		RefreshCw,
		ChevronLeft,
		ChevronRight,
		Info,
		Trash2,
		Check,
		X
	} from 'lucide-svelte';
	import type { SystemLog } from '$lib/types/logs';
	import Icon from '$lib/components/theme/Icon.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import PageHeader from '$lib/components/theme/PageHeader.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import Button from '$lib/components/Button.svelte';

	let logs = $state<SystemLog[]>([]);
	let loading = $state(true);
	let total = $state(0);
	let limit = 50;
	let offset = $state(0);
	let category = $state<'All' | 'Internal' | 'Node' | 'Security'>('All');
	let selectedLog = $state<SystemLog | null>(null);
	let counts = $state<Record<string, number>>({});
	let selectedIds = $state(new Set<number>());

	let isConfirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let isCriticalAction = $state(false);
	let pendingAction = $state<() => Promise<void>>(async () => {});

	const categories = ['All', 'Internal', 'Node', 'Security'];

	async function fetchLogs() {
		loading = true;
		selectedIds = new Set();
		try {
			const q = new URLSearchParams({ limit: limit.toString(), offset: offset.toString(), category: category === 'All' ? '' : category });
			const res = await apiFetch(`/api/logs?${q}`);
			if (res.ok) { const d = await res.json(); logs = d.logs; total = d.total; }
		} catch (e) { console.error(e); }
		finally { loading = false; }
	}

	async function fetchCounts() {
		try {
			const res = await apiFetch('/api/logs/counts');
			if (res.ok) counts = await res.json();
		} catch (e) { console.error(e); }
	}

	function toggleSelection(e: MouseEvent, id: number) {
		e.stopPropagation();
		const n = new Set(selectedIds);
		if (n.has(id)) n.delete(id); else n.add(id);
		selectedIds = n;
	}

	function requestDeleteSelected() {
		confirmTitle = 'Delete Selected Logs';
		confirmMessage = `Delete ${selectedIds.size} selected logs? This action cannot be undone.`;
		isCriticalAction = true;
		pendingAction = async () => {
			try {
				await Promise.all(Array.from(selectedIds).map(id => apiFetch(`/api/logs/${id}`, { method: 'DELETE' })));
				fetchLogs(); fetchCounts();
			} catch (e) { console.error(e); }
		};
		isConfirmOpen = true;
	}

	function requestClearLogs() {
		confirmTitle = 'Clear All Logs';
		confirmMessage = 'Are you sure you want to permanently delete all system logs?';
		isCriticalAction = true;
		pendingAction = async () => {
			const res = await apiFetch('/api/logs', { method: 'DELETE' });
			if (res.ok) { fetchLogs(); fetchCounts(); }
		};
		isConfirmOpen = true;
	}

	function changeCategory(c: string) { category = c as any; offset = 0; fetchLogs(); }
	function nextPage() { if (offset + limit < total) { offset += limit; fetchLogs(); } }
	function prevPage() { if (offset > 0) { offset = Math.max(0, offset - limit); fetchLogs(); } }

	function getLevelColor(level: string) {
		switch (level) {
			case 'ERROR': return 'text-rose-400';
			case 'FATAL': return 'text-rose-500 font-bold';
			case 'WARN': return 'text-amber-400';
			default: return 'text-sky-400';
		}
	}

	onMount(() => { fetchLogs(); fetchCounts(); });
</script>

<div class="space-y-10 font-sans">
	<PageHeader title="System Logs" subtitle="Audit & Events" icon="ph:activity-bold">
		{#snippet actions()}
			<div class="flex flex-wrap gap-3 items-center">
				{#if selectedIds.size > 0}<Button variant="danger" size="md" onclick={requestDeleteSelected} icon="ph:trash-bold">Delete Selected ({selectedIds.size})</Button>{/if}
				<Button variant="secondary" size="md" onclick={requestClearLogs} icon="ph:trash-bold">Clear All</Button>
				<Button variant="secondary" size="md" onclick={() => { fetchLogs(); fetchCounts(); }} loading={loading} icon="ph:arrows-clockwise-bold" />
			</div>
		{/snippet}
	</PageHeader>

	<div class="space-y-6">
		<div class="flex gap-2 overflow-x-auto no-scrollbar pb-2">
			{#each categories as cat}
				<Button variant={category === cat ? 'primary' : 'secondary'} size="sm" onclick={() => changeCategory(cat)} class="whitespace-nowrap rounded-xl">
					{cat} <span class="ml-2 opacity-50 font-sans text-[10px]">{counts[cat] || 0}</span>
				</Button>
			{/each}
		</div>

		<Card>
			<div class="overflow-x-auto overflow-y-auto max-h-[65vh] no-scrollbar">
				<table class="w-full text-left font-sans text-xs">
					<thead class="bg-slate-900/50 text-slate-500 sticky top-0 z-10 border-b border-white/5 backdrop-blur-md">
						<tr class="uppercase font-bold tracking-widest text-[10px]">
							<th class="px-6 py-5 w-16 text-center"><button onclick={() => { if(selectedIds.size === logs.length) selectedIds = new Set(); else selectedIds = new Set(logs.map(l => l.id)); }} class="w-5 h-5 mx-auto border-2 border-slate-700 rounded-lg flex items-center justify-center transition-all {selectedIds.size === logs.length && logs.length > 0 ? 'bg-sky-500 border-sky-400 shadow-lg' : 'hover:border-slate-500'}">{#if selectedIds.size === logs.length && logs.length > 0}<Check size={14} class="text-white" />{/if}</button></th>
							<th class="px-6 py-5">Timestamp</th>
							<th class="px-6 py-5">Level</th>
							<th class="px-6 py-5">Category</th>
							<th class="px-6 py-5">Message</th>
							<th class="px-6 py-5">Path</th>
							<th class="px-6 py-5 w-16"></th>
						</tr>
					</thead>
					<tbody class="divide-y divide-white/5">
						{#if loading && logs.length === 0}
							<tr><td colspan="7" class="py-32"><div class="flex flex-col items-center justify-center gap-4"><div class="w-10 h-10 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"></div><span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Loading Logs...</span></div></td></tr>
						{:else if logs.length === 0}
							<tr><td colspan="7" class="py-32"><div class="flex flex-col items-center justify-center text-slate-600 gap-4"><div class="p-6 bg-slate-900/50 rounded-3xl border border-dashed border-white/5"><Info size={40} class="opacity-20 text-sky-400" /></div><span class="text-xs font-bold uppercase tracking-widest">No records found</span></div></td></tr>
						{:else}
							{#each logs as log (log.id)}
								<tr class="hover:bg-white/5 transition-all cursor-pointer group {selectedIds.has(log.id) ? 'bg-sky-500/10' : ''}" onclick={() => selectedLog = log}>
									<td class="px-6 py-4 text-center" onclick={e => toggleSelection(e, log.id)}><div class="w-5 h-5 mx-auto border-2 border-slate-800 rounded-lg flex items-center justify-center transition-all {selectedIds.has(log.id) ? 'bg-sky-500 border-sky-400' : 'group-hover:border-slate-600 bg-slate-950/40'}">{#if selectedIds.has(log.id)}<Check size={14} class="text-white" />{/if}</div></td>
									<td class="px-6 py-4 whitespace-nowrap text-slate-400 group-hover:text-slate-200 tabular-nums">{log.timestamp ? new Date(log.timestamp).toLocaleString([], { hour12: false }) : 'N/A'}</td>
									<td class="px-6 py-4 font-bold {getLevelColor(log.level)}">{log.level}</td>
									<td class="px-6 py-4"><span class="px-2 py-0.5 font-bold text-[9px] bg-white/5 border border-white/5 text-slate-500 rounded-lg uppercase tracking-wider">{log.category}</span></td>
									<td class="px-6 py-4 text-slate-300 group-hover:text-white max-w-xl truncate font-medium">{log.message}</td>
									<td class="px-6 py-4 text-slate-500 uppercase text-[10px] font-bold">{log.path || '/'}</td>
									<td class="px-6 py-4 text-right"><button onclick={e => { e.stopPropagation(); apiFetch(`/api/logs/${log.id}`, { method: 'DELETE' }).then(r => { if(r.ok) { logs = logs.filter(l => l.id !== log.id); fetchCounts(); } }); }} class="p-2 text-slate-600 hover:text-rose-500 transition-all opacity-0 group-hover:opacity-100"><Trash2 size={16} /></button></td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
			<div class="p-6 border-t border-white/5 bg-slate-900/20 flex items-center justify-between">
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Displaying <span class="text-sky-400">{offset + 1}-{Math.min(offset + limit, total)}</span> of <span class="text-white">{total}</span></span>
				<div class="flex gap-2"><Button variant="secondary" size="sm" onclick={prevPage} disabled={offset === 0} icon="ph:caret-left-bold" /><Button variant="secondary" size="sm" onclick={nextPage} disabled={offset + limit >= total} icon="ph:caret-right-bold" /></div>
			</div>
		</Card>
	</div>
</div>

{#if selectedLog}
	<div class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-slate-950/90 backdrop-blur-md" transition:fade onclick={() => selectedLog = null} role="button" tabindex="0" onkeydown={null}>
		<div class="bg-slate-900 border border-white/10 rounded-[2.5rem] shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col overflow-hidden" onclick={e => e.stopPropagation()} role="document" tabindex="0">
			<div class="p-8 border-b border-white/5 bg-slate-950/40 flex justify-between items-start">
				<div class="flex items-center gap-6">
					<div class="p-4 bg-slate-900 border border-white/5 rounded-2xl shadow-inner"><div class={`font-bold text-2xl tracking-tight ${getLevelColor(selectedLog.level)}`}>{selectedLog.level}</div></div>
					<div><h3 class="text-2xl font-bold text-white tracking-tight">Log Details</h3><p class="text-[10px] text-slate-500 font-bold mt-1 uppercase tracking-widest">Timestamp: {selectedLog.timestamp ? new Date(selectedLog.timestamp).toLocaleString([], { hour12: false }) : 'N/A'}</p></div>
				</div>
				<button onclick={() => selectedLog = null} class="p-2 text-slate-500 hover:text-white transition-all"><X size={24} /></button>
			</div>
			<div class="p-10 overflow-y-auto space-y-10 no-scrollbar bg-slate-900">
				<div class="grid grid-cols-2 md:grid-cols-3 gap-6">
					{#each [
						{ label: 'Category', val: selectedLog.category },
						{ label: 'Source', val: selectedLog.source },
						{ label: 'Path', val: selectedLog.path || 'Root' },
						{ label: 'Method', val: selectedLog.method || 'System' },
						{ label: 'Client IP', val: selectedLog.client_ip || 'Internal' },
						{ label: 'Record ID', val: `#${selectedLog.id}` }
					] as m}
						<div class="p-4 bg-slate-950 border border-white/5 rounded-2xl shadow-sm"><span class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">{m.label}</span><span class="text-xs font-bold text-slate-200 uppercase tracking-tight break-all">{m.val}</span></div>
					{/each}
				</div>
				<div class="space-y-3"><span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Message</span><div class="bg-slate-950 p-6 border border-white/5 text-slate-300 font-mono text-xs whitespace-pre-wrap leading-relaxed rounded-2xl shadow-inner">{selectedLog.message}</div></div>
				{#if selectedLog.details}<div class="space-y-3"><span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Extended Data</span><div class="bg-slate-950 p-6 border border-white/5 text-slate-400 font-mono text-xs whitespace-pre-wrap overflow-x-auto shadow-inner leading-relaxed rounded-2xl">{selectedLog.details}</div></div>{/if}
			</div>
			<div class="p-8 bg-slate-950/40 border-t border-white/5 flex justify-end"><Button onclick={() => selectedLog = null} variant="primary" size="md">Close Details</Button></div>
		</div>
	</div>
{/if}

<ConfirmDialog bind:isOpen={isConfirmOpen} title={confirmTitle} message={confirmMessage} isCritical={isCriticalAction} onConfirm={pendingAction} />

<style>
	.no-scrollbar::-webkit-scrollbar { display: none; }
</style>