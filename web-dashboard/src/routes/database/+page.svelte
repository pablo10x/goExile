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
	import { notifications } from '$lib/stores';
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
	class="flex flex-col lg:flex-row h-[calc(100vh-120px)] md:h-[calc(100vh-140px)] overflow-hidden bg-slate-900/50 backdrop-blur-xl border border-slate-800 rounded-xl shadow-2xl"
>
	<!-- Mobile Top Nav (Horizontal Scroll) -->
	<div
		class="lg:hidden border-b border-slate-800 bg-slate-900/80 overflow-x-auto no-scrollbar backdrop-blur-md shrink-0"
	>
		<div class="flex items-center gap-2 p-3 min-w-max">
			{#each allMenuItems as item}
				{@const isActive = activeTabId === item.id}
				<button
					onclick={() => openTab(item.id, item.label, item.type)}
					class="flex items-center gap-2.5 px-4 py-2 rounded-lg border border-transparent transition-all
					{isActive ? 'bg-blue-600/20 text-blue-400 border-blue-500/30' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}"
				>
					<Icon name={item.iconName} size="1rem" />
					<span class="text-xs font-bold uppercase tracking-wide">{item.label}</span>
				</button>
			{/each}
		</div>
	</div>

	<!-- Desktop Sidebar -->
	<div
		class="hidden lg:flex flex-col border-r border-slate-800 transition-all duration-500 bg-slate-900/50 {isSidebarOpen
			? 'w-64'
			: 'w-20'}"
	>
		<!-- Sidebar Header -->
		<div
			class="p-5 border-b border-slate-800 flex items-center justify-between"
		>
			{#if isSidebarOpen}
				<div class="flex items-center gap-3" transition:fade={{ duration: 150 }}>
					<div class="p-2 bg-blue-500/10 border border-blue-500/20 rounded-lg shadow-lg">
						<Database class="w-5 h-5 text-blue-400" />
					</div>
					<div>
						<h2 class="font-heading font-bold text-sm text-slate-100">DATABASE</h2>
						<p class="font-mono text-[10px] text-slate-500 mt-0.5">MANAGER V1</p>
					</div>
				</div>
			{/if}
			<button
				onclick={() => (isSidebarOpen = !isSidebarOpen)}
				class="p-2 rounded-lg text-slate-500 hover:text-white hover:bg-slate-800 transition-all {isSidebarOpen ? '' : 'mx-auto'}"
			>
				{#if isSidebarOpen}
					<ChevronLeft class="w-4 h-4" />
				{:else}
					<ChevronRight class="w-4 h-4" />
				{/if}
			</button>
		</div>

		<!-- Sidebar Menu -->
		<div class="flex-1 overflow-y-auto py-6 px-3 space-y-8 custom-scrollbar">
			{#each menuCategories as category}
				<div class="space-y-2">
					{#if isSidebarOpen}
						<h3
							class="px-3 text-[10px] font-bold text-slate-500 uppercase tracking-wider flex items-center gap-2"
							transition:fade={{ duration: 100 }}
						>
							{category.name}
						</h3>
					{/if}
					<div class="space-y-1">
						{#each category.items as item}
							{@const isActive = activeTabId === item.id}
							<button
								onclick={() => openTab(item.id, item.label, item.type)}
								class="w-full flex items-center gap-3 p-2.5 rounded-lg transition-all duration-200 group
								{isActive
									? 'bg-blue-600/10 text-blue-400 border border-blue-500/20 shadow-sm'
									: 'text-slate-400 hover:bg-slate-800 hover:text-slate-200 border border-transparent'}"
								title={item.description}
							>
								<div
									class="{isSidebarOpen ? '' : 'mx-auto'} transition-colors {isActive ? 'text-blue-400' : 'text-slate-500 group-hover:text-slate-300'}"
								>
									<Icon name={item.iconName} size="1.2rem" />
								</div>
								{#if isSidebarOpen}
									<div class="flex-1 text-left" transition:fade={{ duration: 100 }}>
										<span class="font-medium text-sm">{item.label}</span>
									</div>
								{/if}
							</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- Sidebar Footer -->
		{#if isSidebarOpen}
			<div
				class="p-4 border-t border-slate-800 bg-slate-900/30"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex items-center gap-3 text-[10px] font-bold uppercase tracking-wide text-slate-500">
					<div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_#10b981]"></div>
					<span class="font-mono">CLUSTER SYNC: ACTIVE</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Main Content Area -->
	<div class="flex-1 flex flex-col min-w-0 bg-slate-950/30 relative">
		<!-- Tab Bar -->
		<QueryTabs {tabs} {activeTabId} onSelect={(id) => (activeTabId = id)} onClose={closeTab} />

		<div class="flex-1 overflow-hidden relative">
			{#each tabs as tab (tab.id)}
				<div
					class="absolute inset-0 {activeTabId === tab.id
						? 'z-10 block'
						: 'z-0 hidden'}"
				>
					{#if tab.type === 'table'}
						<TableTab schema={tab.data.schema} table={tab.data.table} />
					{:else if tab.type === 'browser'}
						<DatabaseBrowserTab onSelectTable={handleSelectTable} />
					{:else if tab.id === 'overview'}
						<!-- Modern Overview Content -->
						<div
							class="h-full overflow-auto custom-scrollbar p-6 sm:p-8 space-y-8"
						>
							
							<!-- Header Section -->
							<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
								<div class="flex items-center gap-5">
									<div
										class="p-4 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-2xl shadow-xl shadow-blue-900/20"
									>
										<Icon name="database" size="2rem" class="text-white" />
									</div>
									<div>
										<h1 class="text-2xl sm:text-3xl font-heading font-bold text-white tracking-tight">
											Database Overview
										</h1>
										<p class="text-sm text-slate-400 font-medium mt-1">
											System Status & Performance Metrics
										</p>
									</div>
								</div>
								
								<div class="flex items-center gap-2 px-4 py-2 bg-emerald-500/10 border border-emerald-500/20 rounded-full">
									<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
									<span class="text-xs font-bold text-emerald-400 uppercase tracking-wide">Operational</span>
								</div>
							</div>

							<!-- Stats Grid -->
							<div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-6">
								<!-- Size -->
								<div class="bg-slate-800/40 border border-slate-700/50 rounded-2xl p-6 hover:border-blue-500/30 transition-all group">
									<div class="flex justify-between items-start mb-4">
										<div class="p-2.5 bg-blue-500/10 rounded-xl text-blue-400 group-hover:bg-blue-500 group-hover:text-white transition-colors">
											<HardDrive class="w-6 h-6" />
										</div>
										<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Storage</span>
									</div>
									<div class="text-3xl font-heading font-bold text-white mb-1">
										{formatBytes(dbStats.size_bytes)}
									</div>
									<div class="text-xs text-slate-400 font-medium">Total Volume Size</div>
								</div>

								<!-- Connections -->
								<div class="bg-slate-800/40 border border-slate-700/50 rounded-2xl p-6 hover:border-emerald-500/30 transition-all group">
									<div class="flex justify-between items-start mb-4">
										<div class="p-2.5 bg-emerald-500/10 rounded-xl text-emerald-400 group-hover:bg-emerald-500 group-hover:text-white transition-colors">
											<Activity class="w-6 h-6" />
										</div>
										<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Active</span>
									</div>
									<div class="text-3xl font-heading font-bold text-white mb-1">
										{dbStats.connections}
									</div>
									<div class="text-xs text-slate-400 font-medium">Open Connections</div>
								</div>

								<!-- Uptime -->
								<div class="bg-slate-800/40 border border-slate-700/50 rounded-2xl p-6 hover:border-purple-500/30 transition-all group">
									<div class="flex justify-between items-start mb-4">
										<div class="p-2.5 bg-purple-500/10 rounded-xl text-purple-400 group-hover:bg-purple-500 group-hover:text-white transition-colors">
											<Clock class="w-6 h-6" />
										</div>
										<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Uptime</span>
									</div>
									<div class="text-3xl font-heading font-bold text-white mb-1">
										{Math.floor(dbStats.uptime_seconds / 3600)}h
									</div>
									<div class="text-xs text-slate-400 font-medium">
										{Math.floor((dbStats.uptime_seconds % 3600) / 60)}m Since Last Restart
									</div>
								</div>

								<!-- Version -->
								<div class="bg-slate-800/40 border border-slate-700/50 rounded-2xl p-6 hover:border-amber-500/30 transition-all group">
									<div class="flex justify-between items-start mb-4">
										<div class="p-2.5 bg-amber-500/10 rounded-xl text-amber-400 group-hover:bg-amber-500 group-hover:text-white transition-colors">
											<Server class="w-6 h-6" />
										</div>
										<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Engine</span>
									</div>
									<div class="text-3xl font-heading font-bold text-white mb-1 truncate">
										{dbStats.version.split(' ')[0] || 'PGSQL'}
									</div>
									<div class="text-xs text-slate-400 font-medium truncate">
										{dbStats.version.split(',')[0] || 'Build Information'}
									</div>
								</div>
							</div>

							<!-- Quick Actions -->
							<div class="space-y-6">
								<h2 class="text-sm font-bold text-slate-400 uppercase tracking-wider flex items-center gap-3">
									<Zap class="w-4 h-4 text-blue-400" />
									Quick Actions
								</h2>
								<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
									<button
										onclick={() => openTab('sql', 'SQL Editor', 'sql')}
										class="flex items-center gap-4 p-5 bg-slate-800/30 border border-slate-700/50 rounded-xl hover:bg-slate-800 hover:border-blue-500/30 transition-all group text-left"
									>
										<div class="p-3 bg-stone-900 rounded-lg text-slate-400 group-hover:text-blue-400 group-hover:bg-blue-500/10 transition-colors">
											<Terminal class="w-6 h-6" />
										</div>
										<div>
											<div class="font-bold text-slate-200 group-hover:text-white">SQL Editor</div>
											<div class="text-xs text-slate-500">Run queries</div>
										</div>
									</button>

									<button
										onclick={() => openTab('functions', 'Functions', 'functions')}
										class="flex items-center gap-4 p-5 bg-slate-800/30 border border-slate-700/50 rounded-xl hover:bg-slate-800 hover:border-purple-500/30 transition-all group text-left"
									>
										<div class="p-3 bg-stone-900 rounded-lg text-slate-400 group-hover:text-purple-400 group-hover:bg-purple-500/10 transition-colors">
											<Code2 class="w-6 h-6" />
										</div>
										<div>
											<div class="font-bold text-slate-200 group-hover:text-white">Functions</div>
											<div class="text-xs text-slate-500">Procedures</div>
										</div>
									</button>

									<button
										onclick={() => openTab('backups', 'Backups', 'backups')}
										class="flex items-center gap-4 p-5 bg-slate-800/30 border border-slate-700/50 rounded-xl hover:bg-slate-800 hover:border-orange-500/30 transition-all group text-left"
									>
										<div class="p-3 bg-stone-900 rounded-lg text-slate-400 group-hover:text-orange-400 group-hover:bg-orange-500/10 transition-colors">
											<HardDrive class="w-6 h-6" />
										</div>
										<div>
											<div class="font-bold text-slate-200 group-hover:text-white">Backups</div>
											<div class="text-xs text-slate-500">Archives</div>
										</div>
									</button>

									<button
										onclick={() => openTab('roles', 'Roles', 'roles')}
										class="flex items-center gap-4 p-5 bg-slate-800/30 border border-slate-700/50 rounded-xl hover:bg-slate-800 hover:border-emerald-500/30 transition-all group text-left"
									>
										<div class="p-3 bg-stone-900 rounded-lg text-slate-400 group-hover:text-emerald-400 group-hover:bg-emerald-500/10 transition-colors">
											<Shield class="w-6 h-6" />
										</div>
										<div>
											<div class="font-bold text-slate-200 group-hover:text-white">Security</div>
											<div class="text-xs text-slate-500">Roles & Users</div>
										</div>
									</button>
								</div>
							</div>

							<!-- Table Stats -->
							{#if tableCounts.length > 0}
								<div class="space-y-6">
									<h2 class="text-sm font-bold text-slate-400 uppercase tracking-wider flex items-center gap-3">
										<BarChart3 class="w-4 h-4 text-blue-400" />
										Table Statistics
									</h2>
									<div
										class="bg-slate-800/30 border border-slate-700/50 rounded-2xl overflow-hidden"
									>
										<div class="overflow-x-auto custom-scrollbar">
											<table class="w-full text-left">
												<thead class="bg-slate-900/50 border-b border-slate-700/50">
													<tr>
														<th
															class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider"
															>Table Name</th
														>
														<th
															class="px-6 py-4 text-right text-xs font-bold text-slate-400 uppercase tracking-wider"
															>Row Count</th
														>
													</tr>
												</thead>
												<tbody class="divide-y divide-slate-700/30">
													{#each tableCounts as table}
														<tr class="hover:bg-slate-800/50 transition-colors group">
															<td class="px-6 py-4">
																<div class="flex items-center gap-3">
																	<Table class="w-4 h-4 text-slate-500 group-hover:text-blue-400 transition-colors" />
																	<span class="font-mono text-sm font-medium text-slate-300 group-hover:text-white"
																		>{table.name}</span
																	>
																</div>
															</td>
															<td class="px-6 py-4 text-right">
																<span class="font-mono text-sm font-bold text-slate-400 group-hover:text-white"
																	>{table.count?.toLocaleString() ?? '0'}</span
																>
															</td>
														</tr>
													{/each}
												</tbody>
											</table>
										</div>
									</div>
								</div>
							{/if}
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
		background: #334155;
		border-radius: 3px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #475569;
	}

	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>