<script lang="ts">
	import {
		Database,
		Search,
		Plus,
		X,
		Trash2,
		Table,
		ChevronLeft,
		ChevronRight,
		Settings,
		LayoutDashboard,
		Terminal,
		Users,
		HardDrive,
		FileText,
		Activity,
		Clock,
		Code2,
		Layers,
		Zap,
		Server,
		BarChart3,
		Shield,
		FolderTree,
		Menu
	} from 'lucide-svelte';
	import SchemaBrowser from '$lib/components/SchemaBrowser.svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import QueryTabs from '$lib/components/database/QueryTabs.svelte';
	import TableTab from '$lib/components/database/TableTab.svelte';
	import DatabaseBrowserTab from '$lib/components/database/DatabaseBrowserTab.svelte';
	import SQLEditorTab from '$lib/components/database/SQLEditorTab.svelte';
	import RolesTab from '$lib/components/database/RolesTab.svelte';
	import BackupsTab from '$lib/components/database/BackupsTab.svelte';
	import ConfigTab from '$lib/components/database/ConfigTab.svelte';
	import FunctionsTab from '$lib/components/database/FunctionsTab.svelte';
	import TableCreatorModal from '$lib/components/TableCreatorModal.svelte';
	import ColumnManagerModal from '$lib/components/ColumnManagerModal.svelte';
	import { notifications } from '$lib/stores.svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import { formatBytes } from '$lib/utils';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';

	// State
	let isSidebarOpen = $state(true);
	let tabs = $state<
		{
			id: string;
			label: string;
			type: 'table' | 'sql' | 'info' | 'config' | 'roles' | 'backups' | 'browser' | 'functions';
			data?: any;
		}[]
	>([{ id: 'overview', label: 'Overview', type: 'info' }]);
	let activeTabId = $state<string>('overview');
	let isLoaded = $state(false);

	// Overview Data
	let dbStats = $state({
		size_bytes: 0,
		version: '',
		connections: 0,
		uptime_seconds: 0,
		active_queries: 0,
		cache_hit_ratio: 0
	});
	let tableCounts = $state<{ name: string; count: number }[]>([]);

	// Sidebar menu items with categories
	const menuCategories = [
		{
			name: 'Overview',
			items: [
				{
					id: 'overview',
					label: 'Telemetry',
					type: 'info',
					iconName: 'ph:chart-line-up-bold',
					description: 'Database overview & stats'
				},
				{
					id: 'browser',
					label: 'Browser',
					type: 'browser',
					iconName: 'ph:folder-open-bold',
					description: 'Browse schemas & tables'
				}
			]
		},
		{
			name: 'Development',
			items: [
				{
					id: 'sql',
					label: 'SQL Editor',
					type: 'sql',
					iconName: 'ph:terminal-window-bold',
					description: 'Execute SQL queries'
				},
				{
					id: 'functions',
					label: 'Functions',
					type: 'functions',
					iconName: 'ph:code-bold',
					description: 'Manage functions & procedures'
				}
			]
		},
		{
			name: 'Administration',
			items: [
				{
					id: 'roles',
					label: 'Security',
					type: 'roles',
					iconName: 'ph:shield-check-bold',
					description: 'User & role management'
				},
				{
					id: 'backups',
					label: 'Archives',
					type: 'backups',
					iconName: 'ph:archive-bold',
					description: 'Backup & restore'
				},
				{
					id: 'config',
					label: 'Tuning',
					type: 'config',
					iconName: 'ph:sliders-bold',
					description: 'PostgreSQL settings'
				}
			]
		}
	];

	// Flattened items for mobile nav
	const allMenuItems = menuCategories.flatMap((c) => c.items);

	// --- Tab Management ---

	function openTab(id: string, label: string, type: any, data: any = {}) {
		const existing = tabs.find((t) => t.id === id);
		if (existing) {
			activeTabId = id;
		} else {
			tabs = [...tabs, { id, label, type, data }];
			activeTabId = id;
		}
	}

	function closeTab(id: string) {
		const idx = tabs.findIndex((t) => t.id === id);
		if (idx === -1) return;

		const newTabs = tabs.filter((t) => t.id !== id);
		tabs = newTabs;

		if (activeTabId === id) {
			// Activate neighbor
			if (newTabs.length > 0) {
				const newIdx = Math.min(idx, newTabs.length - 1);
				activeTabId = newTabs[newIdx].id;
			} else {
				activeTabId = ''; // Or reopen overview
				openTab('overview', 'Dashboard', 'info');
			}
		}
	}

	function handleSelectTable(schema: string, table: string) {
		openTab(`table:${schema}.${table}`, `${table}`, 'table', { schema, table });
	}

	// --- Initialization ---
	async function loadOverviewData() {
		try {
			const res = await fetch('/api/database/overview');
			if (res.ok) {
				dbStats = await res.json();
			}
		} catch (e) {
			console.error('Failed to load overview data', e);
		}

		try {
			const res = await fetch('/api/database/counts');
			if (res.ok) {
				tableCounts = await res.json();
			}
		} catch (e) {
			console.error('Failed to load table counts', e);
		}
	}

	onMount(() => {
		loadOverviewData();
		setTimeout(() => (isLoaded = true), 100);
	});
</script>

<div
	class="flex flex-col lg:flex-row h-[calc(100vh-6rem)] -mt-6 -mx-4 sm:-mx-6 md:-mx-10 overflow-hidden relative z-10"
>
	<!-- Mobile Top Nav -->
	<div
		class="lg:hidden border-b border-slate-800 bg-slate-900/90 overflow-x-auto no-scrollbar backdrop-blur-md shrink-0"
	>
		<div class="flex items-center gap-2 p-3 min-w-max">
			{#each allMenuItems as item}
				{@const isActive = activeTabId === item.id}
				<button
					onclick={() => openTab(item.id, item.label, item.type)}
					class="flex items-center gap-3 px-5 py-2 transition-all rounded-lg
					{isActive ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm' : 'text-slate-500 hover:text-white hover:bg-slate-800'}"
				>
					<Icon name={item.iconName} size="1rem" />
					<span class="text-[10px] font-bold uppercase tracking-widest">{item.label}</span>
				</button>
			{/each}
		</div>
	</div>

	<!-- Desktop Workbench Sidebar -->
	<div
		class="hidden lg:flex flex-col border-r border-slate-800/50 transition-[width] duration-500 bg-slate-900/40 backdrop-blur-xl will-change-[width] {isSidebarOpen
			? 'w-72'
			: 'w-20'} rounded-tr-2xl rounded-br-2xl my-4 ml-4 border-t border-b border-l shadow-2xl"
		style="contain: layout paint;"
	>
		<!-- Sidebar Identity Area -->
		<div
			class="p-6 border-b border-slate-800/50 flex items-center justify-between relative overflow-hidden"
		>
			{#if isSidebarOpen}
				<div class="flex items-center gap-4 relative z-10" transition:fade={{ duration: 150 }}>
					<div class="p-2.5 bg-indigo-500/10 rounded-xl border border-indigo-500/20 shadow-inner">
						<Database class="w-5 h-5 text-indigo-400" />
					</div>
					<div class="flex flex-col">
						<h2 class="font-heading font-black text-[11px] text-white tracking-[0.1em] uppercase">Data_Sector</h2>
						<span class="font-mono text-[8px] text-indigo-400 mt-0.5 font-bold tracking-widest">STATION_PRO_4.2</span>
					</div>
				</div>
			{/if}
			<button
				onclick={() => (isSidebarOpen = !isSidebarOpen)}
				class="p-2 rounded-lg text-slate-500 hover:text-indigo-400 hover:bg-indigo-500/10 transition-all relative z-10 {isSidebarOpen ? '' : 'mx-auto'}"
			>
				{#if isSidebarOpen}
					<ChevronLeft class="w-4 h-4" />
				{:else}
					<ChevronRight class="w-4 h-4" />
				{/if}
			</button>
		</div>

		<!-- Sidebar Navigation Hub -->
		<div class="flex-1 overflow-y-auto py-6 px-4 space-y-8 custom-scrollbar" style="contain: content;">
			{#each menuCategories as category}
				<div class="space-y-3">
					{#if isSidebarOpen}
						<div class="flex items-center gap-3 px-2" transition:fade={{ duration: 100 }}>
							<span class="text-[9px] font-bold text-slate-500 uppercase tracking-widest whitespace-nowrap">{category.name}</span>
							<div class="h-px w-full bg-slate-800/50"></div>
						</div>
					{/if}
					<div class="space-y-1">
						{#each category.items as item}
							{@const isActive = activeTabId === item.id}
							<button
								onclick={() => openTab(item.id, item.label, item.type)}
								class="w-full flex items-center gap-4 p-3 transition-all duration-300 group rounded-xl relative
								{isActive
									? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm'
									: 'text-slate-500 hover:bg-slate-800/50 hover:text-white border border-transparent'}"
								title={item.description}
								style="transform: translateZ(0);"
							>
								<div
									class="transition-transform duration-300 {isActive ? 'scale-110 text-indigo-400' : 'text-slate-500 group-hover:text-slate-300'}"
								>
									<Icon name={item.iconName} size="1.15rem" />
								</div>
								{#if isSidebarOpen}
									<div class="flex-1 text-left flex flex-col" transition:fade={{ duration: 100 }}>
										<span class="font-bold text-[11px] uppercase tracking-wide leading-none mb-1">{item.label}</span>
										<span class="text-[8px] text-slate-500 font-medium tracking-tight truncate group-hover:text-slate-400 transition-colors">{item.description}</span>
									</div>
								{/if}
							</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- Sidebar Status Terminal -->
		{#if isSidebarOpen}
			<div
				class="p-6 border-t border-slate-800/50 bg-slate-900/20"
				transition:fade={{ duration: 150 }}
				style="contain: content;"
			>
				<div class="space-y-4">
					<div class="flex justify-between items-center text-[8px] font-bold uppercase tracking-widest text-slate-500">
						<span>Kernel_Sync</span>
						<span class="text-indigo-400 font-black">ACTIVE</span>
					</div>
					<div class="h-1 bg-slate-800 rounded-full relative overflow-hidden">
						<div class="h-full bg-indigo-500 w-[92%] shadow-[0_0_10px_rgba(99,102,241,0.2)]"></div>
					</div>
					<div class="flex items-center gap-3 text-[9px] font-bold uppercase tracking-widest text-slate-500 italic">
						<div class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_#10b981] animate-pulse" style="transform: translateZ(0);"></div>
						<span>Uplink Verified</span>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Workbench Main Content Area -->
	<div class="flex-1 flex flex-col min-w-0 bg-transparent relative">
		<!-- Integrated Tab Bar -->
		<div class="relative z-20 px-4 pt-4">
			<QueryTabs {tabs} {activeTabId} onSelect={(id) => (activeTabId = id)} onClose={closeTab} />
		</div>

		<div class="flex-1 overflow-hidden relative">
			{#each tabs as tab (tab.id)}
				<div
					class="absolute inset-0 {activeTabId === tab.id
						? 'z-10 block'
						: 'z-0 hidden'} transition-opacity duration-300"
				>
					{#if tab.type === 'table'}
						<TableTab schema={tab.data.schema} table={tab.data.table} />
					{:else if tab.type === 'browser'}
						<DatabaseBrowserTab onSelectTable={handleSelectTable} />
					                    {:else if tab.id === 'overview'}
											<!-- Modern Overview Content -->
											<div
												class="h-full overflow-auto custom-scrollbar p-6 lg:p-10 space-y-8"
												style="contain: content;"
											>
												
												<!-- Workbench Header -->
												<div class="flex flex-col xl:flex-row items-start xl:items-center justify-between gap-10">
													<div class="flex items-center gap-8">
														<div
															class="p-6 bg-slate-900/40 border border-slate-800 rounded-2xl shadow-xl relative group"
															style="transform: translateZ(0);"
														>
															<Icon name="database" size="2rem" class="text-indigo-500 group-hover:scale-110 transition-transform duration-500" />
															<div class="absolute -bottom-1 -right-1 p-1 bg-slate-900 border border-indigo-500/30 rounded-lg shadow-sm">
																<Zap class="w-3.5 h-3.5 text-indigo-400" />
															</div>
														</div>
														<div>
															<div class="flex items-center gap-4 mb-2">
																<span class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-[9px] font-black uppercase tracking-widest italic rounded-lg shadow-sm">Operational</span>
																<span class="text-slate-500 font-mono text-[10px]">// node_id: 0x4F2A</span>
															</div>
															<h1 class="text-4xl sm:text-5xl font-heading font-black text-white tracking-tighter uppercase leading-none">
																Database Engine
															</h1>
															<div class="text-[11px] text-slate-500 font-bold mt-4 uppercase tracking-[0.4em] italic flex items-center gap-3">
																<div class="w-8 h-px bg-slate-800"></div>
																Database Management System
															</div>
														</div>
													</div>
													
													<div class="grid grid-cols-2 gap-4 w-full xl:w-auto">
														<div class="p-5 bg-slate-900/40 border border-slate-800 rounded-2xl flex flex-col gap-2 shadow-sm">
															<span class="text-[8px] font-black text-slate-500 uppercase tracking-widest italic">Connection Status</span>
															<div class="flex items-center gap-1.5">
																{#each [1,2,3,4,5] as i}
																	<div class="w-1.5 h-4 rounded-full {i < 5 ? 'bg-indigo-500 shadow-[0_0_8px_#6366f1]' : 'bg-slate-800'}" style="transform: translateZ(0);"></div>
																{/each}
															</div>
														</div>
														<div class="p-5 bg-slate-900/40 border border-slate-800 rounded-2xl flex flex-col gap-2 shadow-sm">
															<span class="text-[8px] font-black text-slate-500 uppercase tracking-widest italic">System Health</span>
															<span class="text-lg font-heading font-black text-emerald-500 tabular-nums">OPTIMAL</span>
														</div>
													</div>
												</div>
							<!-- Metrics Panel -->
							<div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-6">
								<!-- Size -->
								<div class="modern-card p-8 hover:border-indigo-500/40 transition-all group relative overflow-hidden flex flex-col" style="transform: translateZ(0); contain: content;">
									<div class="absolute -top-10 -right-10 opacity-5 group-hover:scale-110 transition-transform duration-1000 will-change-transform">
										<HardDrive class="w-40 h-40 text-indigo-500" />
									</div>
									<div class="flex justify-between items-start mb-10 relative z-10">
										<div class="p-3 bg-indigo-500/10 rounded-2xl text-indigo-400 border border-indigo-500/20 group-hover:bg-indigo-500 group-hover:text-white transition-all duration-500 shadow-inner">
											<HardDrive class="w-6 h-6" />
										</div>
										<div class="flex flex-col items-end">
											<span class="text-[10px] font-black text-slate-500 uppercase tracking-widest italic">Storage Usage</span>
											<span class="text-[8px] font-mono text-slate-600 mt-1">VOL_01_PRIMARY</span>
										</div>
									</div>
									<div class="mt-auto relative z-10">
										<div class="text-4xl font-heading font-black text-white mb-2 tabular-nums tracking-tighter">
											{formatBytes(dbStats.size_bytes)}
										</div>
										<div class="text-[9px] text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2">
											<span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
											Segmented Data Matrix
										</div>
									</div>
								</div>

								<!-- Connections -->
								<div class="modern-card p-8 hover:border-indigo-500/40 transition-all group relative overflow-hidden flex flex-col" style="transform: translateZ(0); contain: content;">
									<div class="absolute -top-10 -right-10 opacity-5 group-hover:scale-110 transition-transform duration-1000 will-change-transform">
										<Activity class="w-40 h-40 text-indigo-500" />
									</div>
									<div class="flex justify-between items-start mb-10 relative z-10">
										<div class="p-3 bg-indigo-500/10 rounded-2xl text-indigo-400 border border-indigo-500/20 group-hover:bg-indigo-500 group-hover:text-white transition-all duration-500 shadow-inner">
											<Activity class="w-6 h-6" />
										</div>
										<div class="flex flex-col items-end">
											<span class="text-[10px] font-black text-slate-500 uppercase tracking-widest italic">Active Connections</span>
											<span class="text-[8px] font-mono text-slate-600 mt-1">STREAM_AUTH</span>
										</div>
									</div>
									<div class="mt-auto relative z-10">
										<div class="text-4xl font-heading font-black text-white mb-2 tabular-nums tracking-tighter">
											{dbStats.connections}
										</div>
										<div class="text-[9px] text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2">
											<span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
											Verified Client Uplinks
										</div>
									</div>
								</div>

								<!-- Uptime -->
								<div class="modern-card p-8 hover:border-indigo-500/40 transition-all group relative overflow-hidden flex flex-col" style="transform: translateZ(0); contain: content;">
									<div class="absolute -top-10 -right-10 opacity-5 group-hover:scale-110 transition-transform duration-1000 will-change-transform">
										<Clock class="w-40 h-40 text-indigo-500" />
									</div>
									<div class="flex justify-between items-start mb-10 relative z-10">
										<div class="p-3 bg-indigo-500/10 rounded-2xl text-indigo-400 border border-indigo-500/20 group-hover:bg-indigo-500 group-hover:text-white transition-all duration-500 shadow-inner">
											<Clock class="w-6 h-6" />
										</div>
										<div class="flex flex-col items-end">
											<span class="text-[10px] font-black text-slate-500 uppercase tracking-widest italic">System Uptime</span>
											<span class="text-[8px] font-mono text-slate-600 mt-1">SESSION_PERSIST</span>
										</div>
									</div>
									<div class="mt-auto relative z-10">
										<div class="text-4xl font-heading font-black text-white mb-2 tabular-nums tracking-tighter">
											{Math.floor(dbStats.uptime_seconds / 3600)}<span class="text-slate-600 text-2xl">H</span>
										</div>
										<div class="text-[9px] text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2">
											<span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
											Persistent System Uptime
										</div>
									</div>
								</div>

								<!-- Version -->
								<div class="modern-card p-8 hover:border-indigo-500/40 transition-all group relative overflow-hidden flex flex-col" style="transform: translateZ(0); contain: content;">
									<div class="absolute -top-10 -right-10 opacity-5 group-hover:scale-110 transition-transform duration-1000 will-change-transform">
										<Server class="w-40 h-40 text-indigo-500" />
									</div>
									<div class="flex justify-between items-start mb-10 relative z-10">
										<div class="p-3 bg-indigo-500/10 rounded-2xl text-indigo-400 border border-indigo-500/20 group-hover:bg-indigo-500 group-hover:text-white transition-all duration-500 shadow-inner">
											<Server class="w-6 h-6" />
										</div>
										<div class="flex flex-col items-end">
											<span class="text-[10px] font-black text-slate-500 uppercase tracking-widest italic">Engine Version</span>
											<span class="text-[8px] font-mono text-slate-600 mt-1">CORE_REVISION</span>
										</div>
									</div>
									<div class="mt-auto relative z-10">
										<div class="text-3xl font-heading font-black text-white mb-2 truncate uppercase tracking-tighter">
											{dbStats.version.split(' ')[0] || 'PGSQL_PRO'}
										</div>
										<div class="text-[9px] text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2">
											<span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
											Kernel Engine Revision
										</div>
									</div>
								</div>
							</div>

							<div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
								<!-- Quick Command Hub -->
								<div class="lg:col-span-5 space-y-6">
									<h2 class="text-[11px] font-black text-slate-500 uppercase tracking-[0.5em] flex items-center gap-5 italic">
										<Zap class="w-4 h-4 text-amber-500 shadow-[0_0_10px_#f59e0b]" />
										Uplink_Directives
									</h2>
									
									<div class="grid grid-cols-1 gap-4">
										<button
											onclick={() => openTab('sql', 'SQL Editor', 'sql')}
											class="flex items-center gap-6 p-6 bg-slate-900/40 border border-slate-800 rounded-2xl hover:border-amber-500/50 transition-all group text-left shadow-lg overflow-hidden relative"
										>
											<div class="absolute inset-0 bg-amber-500/0 group-hover:bg-amber-500/5 transition-colors duration-500"></div>
											<div class="p-4 bg-slate-950 border border-slate-800 text-slate-500 group-hover:text-amber-400 group-hover:border-amber-500/30 rounded-xl transition-all duration-500 relative z-10">
												<Terminal class="w-7 h-7" />
											</div>
											<div class="relative z-10">
												<div class="font-black text-sm text-slate-300 group-hover:text-white uppercase tracking-widest italic transition-colors">SQL_Manual_Override</div>
												<div class="text-[9px] text-slate-500 font-bold uppercase tracking-[0.2em] mt-2 group-hover:text-slate-400 transition-colors">Execute direct neural queries</div>
											</div>
										</button>

										<button
											onclick={() => openTab('functions', 'Functions', 'functions')}
											class="flex items-center gap-6 p-6 bg-slate-900/40 border border-slate-800 rounded-2xl hover:border-amber-500/50 transition-all group text-left shadow-lg overflow-hidden relative"
										>
											<div class="absolute inset-0 bg-amber-500/0 group-hover:bg-amber-500/5 transition-colors duration-500"></div>
											<div class="p-4 bg-slate-950 border border-slate-800 text-slate-500 group-hover:text-amber-400 group-hover:border-amber-500/30 rounded-xl transition-all duration-500 relative z-10">
												<Code2 class="w-7 h-7" />
											</div>
											<div class="relative z-10">
												<div class="font-black text-sm text-slate-300 group-hover:text-white uppercase tracking-widest italic transition-colors">Logic_Unit_Control</div>
												<div class="text-[9px] text-slate-500 font-bold uppercase tracking-[0.2em] mt-2 group-hover:text-slate-400 transition-colors">Manage stored procedural units</div>
											</div>
										</button>

										<button
											onclick={() => openTab('backups', 'Backups', 'backups')}
											class="flex items-center gap-6 p-6 bg-slate-900/40 border border-slate-800 rounded-2xl hover:border-amber-500/50 transition-all group text-left shadow-lg overflow-hidden relative"
										>
											<div class="absolute inset-0 bg-amber-500/0 group-hover:bg-amber-500/5 transition-colors duration-500"></div>
											<div class="p-4 bg-slate-950 border border-slate-800 text-slate-500 group-hover:text-amber-400 group-hover:border-amber-500/30 rounded-xl transition-all duration-500 relative z-10">
												<HardDrive class="w-7 h-7" />
											</div>
											<div class="relative z-10">
												<div class="font-black text-sm text-slate-300 group-hover:text-white uppercase tracking-widest italic transition-colors">Archive_Sequencing</div>
												<div class="text-[9px] text-slate-500 font-bold uppercase tracking-[0.2em] mt-2 group-hover:text-slate-400 transition-colors">Generate and restore snapshots</div>
											</div>
										</button>
									</div>
								</div>

								<!-- Entity Inventory -->
								<div class="lg:col-span-7 space-y-6">
									<div class="flex items-center justify-between">
										<h2 class="text-[11px] font-black text-slate-500 uppercase tracking-[0.5em] flex items-center gap-5 italic">
											<BarChart3 class="w-4 h-4 text-amber-500" />
											Sector_Inventory_Metrics
										</h2>
										<span class="text-[9px] font-black text-slate-600 uppercase tracking-widest">{tableCounts.length} ENTITIES_MAPPED</span>
									</div>
									
									<div
										class="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden shadow-xl"
									>
										<div class="overflow-x-auto custom-scrollbar">
											<table class="w-full text-left font-jetbrains">
												<thead class="bg-slate-950/50 border-b border-slate-800">
													<tr>
														<th
															class="px-8 py-5 text-[9px] font-black text-slate-500 uppercase tracking-[0.3em] italic"
															>Identifier_Tag</th
														>
														<th
															class="px-8 py-5 text-right text-[9px] font-black text-slate-500 uppercase tracking-[0.3em] italic"
															>Density_Buffer</th
														>
													</tr>
												</thead>
												<tbody class="divide-y divide-slate-800/50">
													{#each tableCounts as table}
														<tr class="hover:bg-slate-800/30 transition-all group">
															<td class="px-8 py-5">
																<div class="flex items-center gap-5">
																	<div class="w-8 h-8 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-center text-slate-600 group-hover:text-amber-500/80 group-hover:border-amber-500/20 transition-all duration-500">
																		<Table class="w-4 h-4" />
																	</div>
																	<span class="font-black text-xs text-slate-400 group-hover:text-white tracking-[0.1em] transition-colors"
																		>{table.name}</span
																	>
																</div>
															</td>
															<td class="px-8 py-4 text-right">
																<div class="flex flex-col items-end gap-1">
																	<span class="font-mono text-sm font-black text-slate-500 group-hover:text-amber-400 tabular-nums transition-colors"
																		>{table.count?.toLocaleString() ?? '0'}</span
																	>
																	<div class="w-24 h-1 bg-slate-800 overflow-hidden rounded-full">
																		<div class="h-full bg-amber-500/20 group-hover:bg-amber-500 transition-all duration-1000" style="width: {Math.min(100, (table.count / 1000) * 100)}%"></div>
																	</div>
																</div>
															</td>
														</tr>
													{/each}
												</tbody>
											</table>
										</div>
									</div>
								</div>
							</div>
						</div>
					{:else if tab.type === 'sql'}
						<SQLEditorTab />
					{:else if tab.type === 'roles'}
						<RolesTab />
					{:else if tab.type === 'backups'}
						<BackupsTab />
					{:else if tab.type === 'config'}
						<ConfigTab />
					{:else if tab.type === 'functions'}
						<FunctionsTab />
					{:else}
						<div class="p-12 text-slate-500 flex flex-col items-center justify-center h-full gap-4">
							<div class="p-6 bg-slate-800/50 rounded-full">
								<FileText class="w-12 h-12 opacity-50" />
							</div>
							<p class="font-medium">Loading Module...</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	/* Custom scrollbar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
		height: 6px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1e293b;
		border-radius: 3px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #334155;
	}

	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>