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
					label: 'Dashboard',
					type: 'info',
					icon: LayoutDashboard,
					color: 'blue',
					description: 'Database overview & stats'
				},
				{
					id: 'browser',
					label: 'Browser',
					type: 'browser',
					icon: FolderTree,
					color: 'indigo',
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
					icon: Terminal,
					color: 'amber',
					description: 'Execute SQL queries'
				},
				{
					id: 'functions',
					label: 'Functions',
					type: 'functions',
					icon: Code2,
					color: 'violet',
					description: 'Manage functions & procedures'
				}
			]
		},
		{
			name: 'Administration',
			items: [
				{
					id: 'roles',
					label: 'Roles',
					type: 'roles',
					icon: Users,
					color: 'emerald',
					description: 'User & role management'
				},
				{
					id: 'backups',
					label: 'Backups',
					type: 'backups',
					icon: HardDrive,
					color: 'orange',
					description: 'Backup & restore'
				},
				{
					id: 'config',
					label: 'Config',
					type: 'config',
					icon: Settings,
					color: 'purple',
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

	function getColorClasses(color: string, isActive: boolean = false) {
		const colors: Record<string, { bg: string; text: string; hover: string; active: string }> = {
			blue: {
				bg: 'bg-rust/10',
				text: 'text-rust-light',
				hover: 'hover:bg-rust/20 hover:text-rust',
				active: 'bg-rust/20 text-rust border-rust/50'
			},
			indigo: {
				bg: 'bg-amber-500/10',
				text: 'text-amber-400',
				hover: 'hover:bg-amber-500/20 hover:text-amber-300',
				active: 'bg-amber-500/20 text-amber-300 border-amber-500/50'
			},
			amber: {
				bg: 'bg-amber-500/10',
				text: 'text-amber-400',
				hover: 'hover:bg-amber-500/20 hover:text-amber-300',
				active: 'bg-amber-500/20 text-amber-300 border-amber-500/50'
			},
			violet: {
				bg: 'bg-violet-500/10',
				text: 'text-violet-400',
				hover: 'hover:bg-violet-500/20 hover:text-violet-300',
				active: 'bg-violet-500/20 text-violet-300 border-violet-500/50'
			},
			emerald: {
				bg: 'bg-emerald-500/10',
				text: 'text-emerald-400',
				hover: 'hover:bg-emerald-500/20 hover:text-emerald-300',
				active: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/50'
			},
			orange: {
				bg: 'bg-orange-500/10',
				text: 'text-orange-400',
				hover: 'hover:bg-orange-500/20 hover:text-orange-300',
				active: 'bg-orange-500/20 text-orange-300 border-orange-500/50'
			},
			purple: {
				bg: 'bg-purple-500/10',
				text: 'text-purple-400',
				hover: 'hover:bg-purple-500/20 hover:text-purple-300',
				active: 'bg-purple-500/20 text-purple-300 border-purple-500/50'
			}
		};
		return colors[color] || colors.blue;
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
	class="flex flex-col md:flex-row h-[calc(100vh-64px)] overflow-hidden bg-white dark:bg-slate-950 text-slate-800 dark:text-slate-200"
>
	<!-- Mobile Top Nav (Horizontal Scroll) -->
	<div
		class="md:hidden border-b border-slate-200/80 dark:border-slate-800/80 bg-slate-900/50 overflow-x-auto no-scrollbar"
	>
		<div class="flex items-center gap-2 p-2 min-w-max">
			{#each allMenuItems as item}
				{@const isActive = activeTabId === item.id}
				{@const colors = getColorClasses(item.color, isActive)}
				<button
					onclick={() => openTab(item.id, item.label, item.type)}
					class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all
					{isActive ? colors.active : 'text-slate-500 dark:text-slate-400 ' + colors.hover}"
				>
					<item.icon class="w-4 h-4" />
					{item.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- Desktop Sidebar -->
	<div
		class="hidden md:flex flex-col border-r border-slate-200/80 dark:border-slate-800/80 transition-all duration-300 bg-gradient-to-b from-slate-900 via-slate-900 to-slate-950 {isSidebarOpen
			? 'w-64'
			: 'w-16'}"
	>
		<!-- Sidebar Header -->
		<div
			class="p-4 border-b border-slate-200/80 dark:border-slate-800/80 bg-slate-900/50 flex items-center justify-between"
		>
			{#if isSidebarOpen}
				<div class="flex items-center gap-3" transition:fade={{ duration: 150 }}>
					<div class="p-2 bg-gradient-to-br from-rust/20 to-rust-light/20 rounded-lg">
						<Database class="w-5 h-5 text-rust-light" />
					</div>
					<div>
						<h2 class="font-bold text-slate-100">Database</h2>
						<p class="text-xs text-slate-500">PostgreSQL Manager</p>
					</div>
				</div>
			{/if}
			<button
				onclick={() => (isSidebarOpen = !isSidebarOpen)}
				class="p-2 rounded-lg text-slate-500 hover:text-slate-700 dark:text-slate-300 hover:bg-slate-800/50 transition-all"
			>
				{#if isSidebarOpen}
					<ChevronLeft class="w-5 h-5" />
				{:else}
					<ChevronRight class="w-5 h-5" />
				{/if}
			</button>
		</div>

		<!-- Sidebar Menu -->
		<div class="flex-1 overflow-y-auto py-4 px-2 space-y-6">
			{#each menuCategories as category}
				<div>
					{#if isSidebarOpen}
						<h3
							class="px-3 mb-2 text-xs font-bold text-slate-500 uppercase tracking-wider"
							transition:fade={{ duration: 100 }}
						>
							{category.name}
						</h3>
					{/if}
					<div class="space-y-1">
						{#each category.items as item}
							{@const isActive = activeTabId === item.id}
							{@const colors = getColorClasses(item.color, isActive)}
							<button
								onclick={() => openTab(item.id, item.label, item.type)}
								class="w-full flex items-center gap-3 p-2.5 rounded-xl transition-all duration-200 border border-transparent
								{isActive
									? colors.active + ' shadow-lg'
									: 'text-slate-500 dark:text-slate-400 ' +
										colors.hover +
										' hover:border-slate-300/50 dark:border-slate-700/50'}"
								title={item.description}
							>
								<div
									class="p-1.5 rounded-lg transition-colors {isActive
										? colors.bg
										: 'bg-slate-800/50'}"
								>
									<item.icon class="w-4 h-4 {isActive ? colors.text : ''}" />
								</div>
								{#if isSidebarOpen}
									<div class="flex-1 text-left" transition:fade={{ duration: 100 }}>
										<span class="text-sm font-medium">{item.label}</span>
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
				class="p-4 border-t border-slate-200/80 dark:border-slate-800/80 bg-white/50 dark:bg-slate-950/50"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex items-center gap-3 text-xs text-slate-500">
					<div class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div>
					<span>Connected</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Main Content Area -->
	<div class="flex-1 flex flex-col min-w-0 bg-slate-900">
		<QueryTabs {tabs} {activeTabId} onSelect={(id) => (activeTabId = id)} onClose={closeTab} />

		<div class="flex-1 overflow-auto relative">
			{#each tabs as tab (tab.id)}
				<div
					class="absolute inset-0 bg-slate-900 {activeTabId === tab.id
						? 'z-10 block'
						: 'z-0 hidden'}"
				>
					{#if tab.type === 'table'}
						<TableTab schema={tab.data.schema} table={tab.data.table} />
					{:else if tab.type === 'browser'}
						<DatabaseBrowserTab onSelectTable={handleSelectTable} />
					{:else if tab.id === 'overview'}
						<!-- Enhanced Overview Content -->
						<div
							class="h-full overflow-auto bg-gradient-to-br from-slate-900 via-slate-900 to-slate-950"
						>
							<!-- Header -->
							<div
								class="p-4 sm:p-8 border-b border-[var(--border-color)] bg-gradient-to-r from-rust/5 via-transparent to-rust-light/5"
							>
								<div class="flex items-center gap-4 mb-2">
									<div
										class="p-3 bg-gradient-to-br from-rust/20 to-rust-light/20 rounded-xl border border-rust/30 shadow-lg shadow-rust/10"
									>
										<Database class="w-6 h-6 sm:w-8 sm:h-8 text-rust-light" />
									</div>
									<div>
										<h1 class="text-2xl sm:text-3xl font-bold text-slate-100">Database Overview</h1>
										<p class="text-sm text-slate-500 mt-1 hidden sm:block">
											Monitor your PostgreSQL database performance and health
										</p>
									</div>
								</div>
							</div>

							<!-- Stats Grid -->
							<div class="p-4 sm:p-8">
								<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-8">
									<!-- Database Size Card -->
									<div
										class="relative overflow-hidden bg-gradient-to-br from-slate-800/80 to-slate-800/40 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-6 hover:border-rust-light/30 transition-all group"
									>
										<div
											class="absolute inset-0 bg-gradient-to-br from-rust/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
										></div>
										<div class="relative">
											<div class="flex items-center justify-between mb-4">
												<div class="p-2.5 bg-rust/10 rounded-xl">
													<HardDrive class="w-6 h-6 text-rust-light" />
												</div>
												<span
													class="text-xs font-medium text-rust-light bg-rust/10 px-2 py-1 rounded-full"
													>Storage</span
												>
											</div>
											<div class="text-3xl font-bold text-slate-100 mb-1">
												{formatBytes(dbStats.size_bytes)}
											</div>
											<div class="text-sm text-slate-500">Database Size</div>
										</div>
									</div>

									<!-- Connections Card -->
									<div
										class="relative overflow-hidden bg-gradient-to-br from-slate-800/80 to-slate-800/40 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-6 hover:border-emerald-500/30 transition-all group"
									>
										<div
											class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
										></div>
										<div class="relative">
											<div class="flex items-center justify-between mb-4">
												<div class="p-2.5 bg-emerald-500/10 rounded-xl">
													<Activity class="w-6 h-6 text-emerald-400" />
												</div>
												<span
													class="text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2 py-1 rounded-full"
													>Active</span
												>
											</div>
											<div class="text-3xl font-bold text-slate-100 mb-1">
												{dbStats.connections}
											</div>
											<div class="text-sm text-slate-500">Connections</div>
										</div>
									</div>

									<!-- Uptime Card -->
									<div
										class="relative overflow-hidden bg-gradient-to-br from-slate-800/80 to-slate-800/40 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-6 hover:border-purple-500/30 transition-all group"
									>
										<div
											class="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
										></div>
										<div class="relative">
											<div class="flex items-center justify-between mb-4">
												<div class="p-2.5 bg-purple-500/10 rounded-xl">
													<Clock class="w-6 h-6 text-purple-400" />
												</div>
												<span
													class="text-xs font-medium text-purple-400 bg-purple-500/10 px-2 py-1 rounded-full"
													>Uptime</span
												>
											</div>
											<div class="text-3xl font-bold text-slate-100 mb-1">
												{Math.floor(dbStats.uptime_seconds / 3600)}h
											</div>
											<div class="text-sm text-slate-500">
												{Math.floor((dbStats.uptime_seconds % 3600) / 60)}m running
											</div>
										</div>
									</div>

									<!-- Version Card -->
									<div
										class="relative overflow-hidden bg-gradient-to-br from-slate-800/80 to-slate-800/40 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl p-6 hover:border-orange-500/30 transition-all group"
									>
										<div
											class="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
										></div>
										<div class="relative">
											<div class="flex items-center justify-between mb-4">
												<div class="p-2.5 bg-orange-500/10 rounded-xl">
													<Server class="w-6 h-6 text-orange-400" />
												</div>
												<span
													class="text-xs font-medium text-orange-400 bg-orange-500/10 px-2 py-1 rounded-full"
													>Version</span
												>
											</div>
											<div class="text-3xl font-bold text-slate-100 mb-1 truncate">
												{dbStats.version.split(' ')[0] || 'PostgreSQL'}
											</div>
											<div class="text-sm text-slate-500 truncate">
												{dbStats.version.split(',')[0] || 'Database Server'}
											</div>
										</div>
									</div>
								</div>

								<!-- Quick Actions -->
								<div class="mb-8">
									<h2 class="text-lg font-bold text-slate-100 mb-4">Quick Actions</h2>
									<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
										<button
											onclick={() => openTab('sql', 'SQL Editor', 'sql')}
											class="flex flex-col items-center gap-3 p-6 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl hover:border-amber-500/30 hover:bg-slate-800 transition-all group"
										>
											<div
												class="p-3 bg-amber-500/10 rounded-xl group-hover:bg-amber-500/20 transition-colors"
											>
												<Terminal class="w-6 h-6 text-amber-400" />
											</div>
											<span class="text-sm font-medium text-slate-700 dark:text-slate-300"
												>Run Query</span
											>
										</button>

										<button
											onclick={() => openTab('functions', 'Functions', 'functions')}
											class="flex flex-col items-center gap-3 p-6 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl hover:border-violet-500/30 hover:bg-slate-800 transition-all group"
										>
											<div
												class="p-3 bg-violet-500/10 rounded-xl group-hover:bg-violet-500/20 transition-colors"
											>
												<Code2 class="w-6 h-6 text-violet-400" />
											</div>
											<span class="text-sm font-medium text-slate-700 dark:text-slate-300"
												>Functions</span
											>
										</button>

										<button
											onclick={() => openTab('backups', 'Backups', 'backups')}
											class="flex flex-col items-center gap-3 p-6 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl hover:border-orange-500/30 hover:bg-slate-800 transition-all group"
										>
											<div
												class="p-3 bg-orange-500/10 rounded-xl group-hover:bg-orange-500/20 transition-colors"
											>
												<HardDrive class="w-6 h-6 text-orange-400" />
											</div>
											<span class="text-sm font-medium text-slate-700 dark:text-slate-300"
												>Backup DB</span
											>
										</button>

										<button
											onclick={() => openTab('roles', 'Roles', 'roles')}
											class="flex flex-col items-center gap-3 p-6 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl hover:border-emerald-500/30 hover:bg-slate-800 transition-all group"
										>
											<div
												class="p-3 bg-emerald-500/10 rounded-xl group-hover:bg-emerald-500/20 transition-colors"
											>
												<Shield class="w-6 h-6 text-emerald-400" />
											</div>
											<span class="text-sm font-medium text-slate-700 dark:text-slate-300"
												>Manage Roles</span
											>
										</button>
									</div>
								</div>

								<!-- Table Stats -->
								{#if tableCounts.length > 0}
									<div>
										<h2 class="text-lg font-bold text-slate-100 mb-4">Table Statistics</h2>
										<div
											class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden"
										>
											<div class="overflow-x-auto">
												<table class="w-full">
													<thead class="bg-slate-900/50">
														<tr>
															<th
																class="px-6 py-3 text-left text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
																>Table</th
															>
															<th
																class="px-6 py-3 text-right text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
																>Row Count</th
															>
														</tr>
													</thead>
													<tbody class="divide-y divide-slate-700/50">
														{#each tableCounts.slice(0, 10) as table}
															<tr class="hover:bg-slate-800/50 transition-colors">
																<td class="px-6 py-4">
																	<div class="flex items-center gap-3">
																		<Table class="w-4 h-4 text-slate-500" />
																		<span class="font-medium text-slate-800 dark:text-slate-200"
																			>{table.name}</span
																		>
																	</div>
																</td>
																<td class="px-6 py-4 text-right">
																	<span class="font-mono text-slate-700 dark:text-slate-300"
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
						<div class="p-8 text-slate-500 flex flex-col items-center justify-center h-full">
							<FileText class="w-16 h-16 opacity-20 mb-4" />
							<p>Feature {tab.label} is currently being upgraded.</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	/* Custom scrollbar for sidebar */
	.overflow-y-auto::-webkit-scrollbar {
		width: 4px;
	}

	.overflow-y-auto::-webkit-scrollbar-track {
		background: transparent;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb {
		background: rgb(51 65 85 / 0.5);
		border-radius: 2px;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb:hover {
		background: rgb(71 85 105 / 0.5);
	}

	/* Hide scrollbar for mobile horizontal nav */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
