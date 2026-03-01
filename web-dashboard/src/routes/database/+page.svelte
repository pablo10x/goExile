<script lang="ts">
import { apiFetch } from "$lib/api";
	import {
		Database,
		ChevronLeft,
		ChevronRight,
		HardDrive,
		Activity,
		Clock,
		Server,
		Zap,
		Terminal,
		Code2,
		Table,
		BarChart3,
		FileText
	} from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import QueryTabs from '$lib/components/database/QueryTabs.svelte';
	import TableTab from '$lib/components/database/TableTab.svelte';
	import DatabaseBrowserTab from '$lib/components/database/DatabaseBrowserTab.svelte';
	import SQLEditorTab from '$lib/components/database/SQLEditorTab.svelte';
	import RolesTab from '$lib/components/database/RolesTab.svelte';
	import BackupsTab from '$lib/components/database/BackupsTab.svelte';
	import ConfigTab from '$lib/components/database/ConfigTab.svelte';
	import FunctionsTab from '$lib/components/database/FunctionsTab.svelte';
	import { formatBytes } from '$lib/utils';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
    import Card from '$lib/components/theme/Card.svelte';

	let isSidebarOpen = $state(true);
	let tabs = $state<any[]>([{ id: 'overview', label: 'Overview', type: 'info' }]);
	let activeTabId = $state<string>('overview');
	let isLoaded = $state(false);

	let dbStats = $state({
		size_bytes: 0,
		version: '',
		connections: 0,
		uptime_seconds: 0,
		active_queries: 0,
		cache_hit_ratio: 0
	});
	let tableCounts = $state<{ name: string; count: number }[]>([]);

	const menuCategories = [
		{
			name: 'Overview',
			items: [
				{ id: 'overview', label: 'Dashboard', type: 'info', iconName: 'ph:chart-line-up-bold', description: 'Real-time database metrics' },
				{ id: 'browser', label: 'Browser', type: 'browser', iconName: 'ph:folder-open-bold', description: 'Explore tables and data' }
			]
		},
		{
			name: 'Development',
			items: [
				{ id: 'sql', label: 'Query Editor', type: 'sql', iconName: 'ph:terminal-window-bold', description: 'Execute direct SQL queries' },
				{ id: 'functions', label: 'Procedures', type: 'functions', iconName: 'ph:code-bold', description: 'Manage stored functions' }
			]
		},
		{
			name: 'Management',
			items: [
				{ id: 'roles', label: 'Security', type: 'roles', iconName: 'ph:shield-check-bold', description: 'Roles and permissions' },
				{ id: 'backups', label: 'Backups', type: 'backups', iconName: 'ph:archive-bold', description: 'Snapshots and restoration' },
				{ id: 'config', label: 'Settings', type: 'config', iconName: 'ph:sliders-bold', description: 'Database configuration' }
			]
		}
	];

	const allMenuItems = menuCategories.flatMap((c) => c.items);

	function openTab(id: string, label: string, type: any, data: any = {}) {
		const existing = tabs.find((t) => t.id === id);
		if (existing) activeTabId = id;
		else { tabs = [...tabs, { id, label, type, data }]; activeTabId = id; }
	}

	function closeTab(id: string) {
		const idx = tabs.findIndex((t) => t.id === id);
		if (idx === -1) return;
		const newTabs = tabs.filter((t) => t.id !== id);
		tabs = newTabs;
		if (activeTabId === id) {
			if (newTabs.length > 0) activeTabId = newTabs[Math.min(idx, newTabs.length - 1)].id;
			else openTab('overview', 'Dashboard', 'info');
		}
	}

	function handleSelectTable(schema: string, table: string) {
		openTab(`table:${schema}.${table}`, `${table}`, 'table', { schema, table });
	}

	async function loadOverviewData() {
		try {
			const res = await apiFetch('/api/database/overview');
			if (res.ok) dbStats = await res.json();
			const resCounts = await apiFetch('/api/database/counts');
			if (resCounts.ok) tableCounts = await resCounts.json();
		} catch (e) { console.error(e); }
	}

	onMount(() => { loadOverviewData(); setTimeout(() => (isLoaded = true), 100); });
</script>

<div class="flex flex-col lg:flex-row h-[calc(100vh-6rem)] -mt-6 -mx-4 sm:-mx-6 md:-mx-10 overflow-hidden relative z-10 font-sans">
	<!-- Mobile Nav -->
	<div class="lg:hidden border-b border-white/5 bg-slate-950/80 overflow-x-auto no-scrollbar backdrop-blur-md shrink-0">
		<div class="flex items-center gap-2 p-3 min-w-max">
			{#each allMenuItems as item}
				<button onclick={() => openTab(item.id, item.label, item.type)} class="flex items-center gap-3 px-5 py-2 transition-all rounded-lg {activeTabId === item.id ? 'bg-sky-500/10 text-sky-400 border border-sky-500/20 shadow-sm' : 'text-slate-500 hover:text-white hover:bg-slate-800'}">
					<Icon name={item.iconName} size="1rem" /><span class="text-xs font-bold uppercase tracking-wide">{item.label}</span>
				</button>
			{/each}
		</div>
	</div>

	<!-- Workbench Sidebar -->
	<div class="hidden lg:flex flex-col border-r border-white/5 transition-all duration-500 bg-slate-900/40 backdrop-blur-xl {isSidebarOpen ? 'w-72' : 'w-24'} rounded-tr-3xl rounded-br-3xl my-4 ml-4 shadow-2xl border-t border-b border-l">
		<div class="p-8 border-b border-white/5 flex items-center justify-between">
			{#if isSidebarOpen}
				<div class="flex items-center gap-4" transition:fade>
					<div class="p-3 bg-sky-500/10 rounded-2xl border border-sky-500/20 shadow-sm"><Database class="w-5 h-5 text-sky-400" /></div>
					<div class="flex flex-col"><h2 class="text-sm font-bold text-white tracking-tight leading-none">Database</h2><span class="text-[10px] text-slate-500 font-bold uppercase mt-1 tracking-wider">Storage</span></div>
				</div>
			{/if}
			<button onclick={() => (isSidebarOpen = !isSidebarOpen)} class="p-2 rounded-xl text-slate-500 hover:text-sky-400 hover:bg-white/5 transition-all {isSidebarOpen ? '' : 'mx-auto'}">
				{#if isSidebarOpen}<ChevronLeft class="w-4 h-4" />{:else}<ChevronRight class="w-4 h-4" />{/if}
			</button>
		</div>

		<div class="flex-1 overflow-y-auto py-6 px-4 space-y-8 no-scrollbar">
			{#each menuCategories as category}
				<div class="space-y-3">
					{#if isSidebarOpen}<span class="text-[11px] font-bold text-slate-600 uppercase tracking-widest ml-3">{category.name}</span>{/if}
					<div class="space-y-1">
						{#each category.items as item}
							<button onclick={() => openTab(item.id, item.label, item.type)} class="w-full flex items-center gap-4 p-3.5 transition-all rounded-2xl {activeTabId === item.id ? 'bg-white/10 text-sky-400 shadow-xl' : 'text-slate-500 hover:bg-white/5 hover:text-slate-200'}">
								<Icon name={item.iconName} size="1.25rem" class={activeTabId === item.id ? 'scale-110' : ''} />
								{#if isSidebarOpen}<div class="flex-1 text-left flex flex-col"><span class="font-bold text-xs uppercase tracking-wide leading-none mb-1">{item.label}</span><span class="text-[10px] text-slate-500 font-medium truncate">{item.description}</span></div>{/if}
							</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		{#if isSidebarOpen}
			<div class="p-8 border-t border-white/5 bg-white/[0.02]">
				<div class="space-y-4">
					<div class="flex justify-between items-center text-[10px] font-bold uppercase tracking-wider text-slate-500"><span>Sync Status</span><span class="text-sky-400">Active</span></div>
					<div class="h-1 bg-slate-800 rounded-full overflow-hidden"><div class="h-full bg-sky-500 w-[92%]"></div></div>
					<div class="flex items-center gap-2 text-[10px] font-bold text-slate-600 uppercase tracking-wide"><div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div><span>Live Connection</span></div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Main Area -->
	<div class="flex-1 flex flex-col min-w-0 bg-transparent relative">
		<div class="relative z-20 px-6 pt-6"><QueryTabs {tabs} {activeTabId} onSelect={(id) => (activeTabId = id)} onClose={closeTab} /></div>
		<div class="flex-1 overflow-hidden relative">
			{#each tabs as tab (tab.id)}
				<div class="absolute inset-0 {activeTabId === tab.id ? 'z-10 block' : 'z-0 hidden'}">
					{#if tab.type === 'table'}<TableTab schema={tab.data.schema} table={tab.data.table} />
					{:else if tab.type === 'browser'}<DatabaseBrowserTab onSelectTable={handleSelectTable} />
					{:else if tab.id === 'overview'}
						<div class="h-full overflow-auto no-scrollbar p-6 lg:p-10 space-y-10">
							<div class="flex flex-col xl:flex-row items-start xl:items-center justify-between gap-10">
								<div class="flex items-center gap-8">
									<div class="p-6 bg-slate-800/40 border border-white/5 rounded-3xl shadow-2xl backdrop-blur-md group"><Icon name="database" size="2.5rem" class="text-sky-500 transition-transform group-hover:scale-110" /></div>
									<div>
										<div class="flex items-center gap-4 mb-2"><span class="px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-[10px] font-bold uppercase tracking-wide rounded-lg">Operational</span><span class="text-slate-500 text-xs font-mono">ID: 0x4F2A</span></div>
										<h1 class="text-4xl font-bold text-white tracking-tight">Database Status</h1>
										<p class="text-xs font-medium text-slate-500 mt-2 uppercase tracking-widest">PostgreSQL Management System</p>
									</div>
								</div>
								<div class="grid grid-cols-2 gap-4 w-full xl:w-auto">
									<div class="p-6 bg-slate-800/40 border border-white/5 rounded-2xl shadow-xl backdrop-blur-sm"><span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block mb-2">Connections</span><div class="flex items-center gap-2">{#each [1,2,3,4,5] as i}<div class="w-2 h-5 rounded-full {i < 5 ? 'bg-sky-500 shadow-lg shadow-sky-500/30' : 'bg-slate-800'}"></div>{/each}</div></div>
									<div class="p-6 bg-slate-800/40 border border-white/5 rounded-2xl shadow-xl backdrop-blur-sm"><span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest block mb-2">Health</span><span class="text-2xl font-bold text-emerald-400">OPTIMAL</span></div>
								</div>
							</div>

							<div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-6">
								{#each [
									{ icon: HardDrive, label: 'Storage Usage', val: formatBytes(dbStats.size_bytes), sub: 'Primary Disk', color: 'text-sky-400' },
									{ icon: Activity, label: 'Active Links', val: dbStats.connections, sub: 'Verified Auth', color: 'text-sky-400' },
									{ icon: Clock, label: 'System Uptime', val: `${Math.floor(dbStats.uptime_seconds / 3600)}h`, sub: 'Active Session', color: 'text-sky-400' },
									{ icon: Server, label: 'Engine Build', val: dbStats.version.split(' ')[0] || 'Postgres', sub: 'Core Release', color: 'text-sky-400' }
								] as m}
									<div class="modern-card p-8 bg-slate-800/40 border-white/5 rounded-3xl flex flex-col group relative overflow-hidden transition-all hover:border-sky-500/30">
										<m.icon class="absolute -top-10 -right-10 w-40 h-40 opacity-5 group-hover:scale-110 transition-transform duration-1000" />
										<div class="flex justify-between items-start mb-10 relative z-10">
											<div class="p-3 bg-sky-500/10 rounded-2xl border border-sky-500/20 transition-all group-hover:bg-sky-500 group-hover:text-white"><m.icon class="w-6 h-6" /></div>
											<div class="text-right flex flex-col"><span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">{m.label}</span><span class="text-[9px] text-slate-600 mt-1 font-bold">{m.sub}</span></div>
										</div>
										<div class="mt-auto relative z-10"><div class="text-3xl font-bold text-white mb-1">{m.val}</div><div class="flex items-center gap-2 text-[10px] font-bold text-slate-500 uppercase tracking-wide"><div class="w-1 h-1 rounded-full bg-sky-500"></div>System Monitor</div></div>
									</div>
								{/each}
							</div>

							<div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
								<div class="lg:col-span-5 space-y-6">
									<h2 class="text-[11px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-3"><div class="p-1.5 bg-amber-500/10 rounded-lg"><Zap class="w-4 h-4 text-amber-500" /></div>Quick Actions</h2>
									<div class="space-y-4">
										{#each [
											{ id: 'sql', type: 'sql', label: 'Query Editor', desc: 'Execute SQL queries', icon: Terminal, accent: 'amber' },
											{ id: 'functions', type: 'functions', label: 'Stored Procedures', desc: 'Manage functions', icon: Code2, accent: 'sky' },
											{ id: 'backups', type: 'backups', label: 'Backups', desc: 'Snapshots & Restoration', icon: HardDrive, accent: 'emerald' }
										] as a}
											<button onclick={() => openTab(a.id, a.label, a.type)} class="w-full flex items-center gap-6 p-6 bg-slate-800/40 border border-white/5 rounded-[2rem] hover:border-{a.accent}-500/50 transition-all group shadow-xl backdrop-blur-md">
												<div class="p-4 bg-slate-900 border border-white/5 text-slate-500 group-hover:text-{a.accent}-400 rounded-2xl transition-all shadow-inner"><a.icon class="w-7 h-7" /></div>
												<div><div class="font-bold text-base text-slate-200 group-hover:text-white transition-colors">{a.label}</div><div class="text-[11px] text-slate-500 font-bold uppercase tracking-wider mt-1">{a.desc}</div></div>
											</button>
										{/each}
									</div>
								</div>
								<div class="lg:col-span-7 space-y-6">
									<div class="flex items-center justify-between px-2"><h2 class="text-[11px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-3"><div class="p-1.5 bg-amber-500/10 rounded-lg"><BarChart3 class="w-4 h-4 text-amber-500" /></div>Table Metrics</h2><span class="text-[10px] font-bold text-slate-600 uppercase tracking-wider">{tableCounts.length} Records</span></div>
									<div class="bg-slate-800/40 border border-white/5 rounded-[2rem] overflow-hidden shadow-2xl backdrop-blur-md">
										<table class="w-full text-left">
											<thead class="bg-slate-950/50 border-b border-white/5"><tr><th class="px-8 py-5 text-[11px] font-bold text-slate-500 uppercase tracking-wider">Table Name</th><th class="px-8 py-5 text-right text-[11px] font-bold text-slate-500 uppercase tracking-wider">Record Count</th></tr></thead>
											<tbody class="divide-y divide-white/5">
												{#each tableCounts as table}
													<tr class="hover:bg-white/5 transition-all group">
														<td class="px-8 py-5"><div class="flex items-center gap-5"><div class="w-10 h-10 bg-slate-900 border border-white/5 rounded-xl flex items-center justify-center text-slate-500 group-hover:text-amber-400 transition-all"><Table class="w-5 h-5" /></div><span class="font-bold text-sm text-slate-200 group-hover:text-white transition-colors">{table.name}</span></div></td>
														<td class="px-8 py-4 text-right"><div class="flex flex-col items-end gap-2"><span class="text-base font-bold text-slate-400 group-hover:text-amber-400 tabular-nums">{table.count?.toLocaleString() ?? '0'}</span><div class="w-28 h-1.5 bg-slate-900 overflow-hidden rounded-full border border-white/5 shadow-inner"><div class="h-full bg-amber-500/20 group-hover:bg-amber-500 transition-all duration-1000" style="width: {Math.min(100, (table.count / 1000) * 100)}%"></div></div></div></td>
													</tr>
												{/each}
											</tbody>
										</table>
									</div>
								</div>
							</div>
						</div>
					{:else if tab.type === 'sql'}<SQLEditorTab />
					{:else if tab.type === 'roles'}<RolesTab />
					{:else if tab.type === 'backups'}<BackupsTab />
					{:else if tab.type === 'config'}<ConfigTab />
					{:else if tab.type === 'functions'}<FunctionsTab />
					{:else}
						<div class="p-12 text-slate-500 flex flex-col items-center justify-center h-full gap-4"><div class="p-6 bg-slate-800/50 rounded-full"><FileText class="w-12 h-12 opacity-50" /></div><p class="font-bold uppercase tracking-widest text-xs">Loading Session...</p></div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar { width: 6px; }
	.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
	.custom-scrollbar::-webkit-scrollbar-thumb { background: #334155; border-radius: 99px; }
	.no-scrollbar::-webkit-scrollbar { display: none; }
</style>