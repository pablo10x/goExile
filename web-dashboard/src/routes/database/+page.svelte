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
			rust: {
				bg: 'bg-rust/10',
				text: 'text-rust-light',
				hover: 'hover:bg-rust/20 hover:text-rust',
				active: 'bg-rust/20 text-rust border-rust/50'
			},
			amber: {
				bg: 'bg-amber-500/10',
				text: 'text-amber-400',
				hover: 'hover:bg-amber-500/20 hover:text-amber-300',
				active: 'bg-amber-500/20 text-amber-300 border-amber-500/50'
			},
			red: {
				bg: 'bg-red-500/10',
				text: 'text-red-400',
				hover: 'hover:bg-red-500/20 hover:text-red-300',
				active: 'bg-red-500/20 text-red-300 border-red-500/50'
			},
			slate: {
				bg: 'bg-slate-500/10',
				text: 'text-slate-400',
				hover: 'hover:bg-slate-500/20 hover:text-slate-300',
				active: 'bg-slate-500/20 text-slate-300 border-slate-500/50'
			}
		};
		return colors[color] || colors.rust;
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
	class="flex flex-col lg:flex-row h-[calc(100vh-120px)] md:h-[calc(100vh-140px)] overflow-hidden bg-[#050505] text-slate-200 border border-stone-800 shadow-2xl"
>
	<!-- Mobile Top Nav (Horizontal Scroll) -->
	<div
		class="lg:hidden border-b border-stone-800 bg-[#0a0a0a]/80 overflow-x-auto no-scrollbar backdrop-blur-md shrink-0"
	>
		<div class="flex items-center gap-2 p-3 min-w-max">
			{#each allMenuItems as item}
				{@const isActive = activeTabId === item.id}
				{@const colors = getColorClasses(isActive ? 'rust' : 'slate')}
				<button
					onclick={() => openTab(item.id, item.label, item.type)}
					class="flex items-center gap-2.5 px-4 py-2 rounded-none border border-transparent transition-all
					{isActive ? colors.active + ' border-rust/50 shadow-lg shadow-rust/10' : 'text-stone-500 ' + colors.hover}"
				>
					<item.icon class="w-4 h-4" />
					<span class="text-[10px] font-heading font-black uppercase tracking-widest">{item.label}</span>
				</button>
			{/each}
		</div>
	</div>

	<!-- Desktop Sidebar -->
	<div
		class="hidden lg:flex flex-col border-r border-stone-800 transition-all duration-500 bg-[#0a0a0a] {isSidebarOpen
			? 'w-72'
			: 'w-20'}"
	>
		<!-- Sidebar Header -->
		<div
			class="p-6 border-b border-stone-800 bg-stone-900/30 flex items-center justify-between"
		>
			{#if isSidebarOpen}
				<div class="flex items-center gap-4" transition:fade={{ duration: 150 }}>
					<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame shadow-lg">
						<Database class="w-6 h-6 text-rust-light" />
					</div>
					<div>
						<h2 class="font-heading font-black text-sm tracking-tighter text-slate-100">DATA_CORE</h2>
						<p class="font-jetbrains text-[9px] text-rust/60 uppercase tracking-widest font-black mt-0.5">DB_MANAGER_V1</p>
					</div>
				</div>
			{/if}
			<button
				onclick={() => (isSidebarOpen = !isSidebarOpen)}
				class="p-2.5 rounded-none text-stone-600 hover:text-rust hover:bg-rust/5 border border-transparent hover:border-rust/20 transition-all {isSidebarOpen ? '' : 'mx-auto'}"
			>
				{#if isSidebarOpen}
					<ChevronLeft class="w-5 h-5" />
				{:else}
					<ChevronRight class="w-5 h-5" />
				{/if}
			</button>
		</div>

		<!-- Sidebar Menu -->
		<div class="flex-1 overflow-y-auto py-8 px-4 space-y-10 custom-scrollbar">
			{#each menuCategories as category}
				<div class="space-y-3">
					{#if isSidebarOpen}
						<h3
							class="px-3 text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] flex items-center gap-2"
							transition:fade={{ duration: 100 }}
						>
							<div class="w-1 h-1 bg-stone-800"></div>
							{category.name}
						</h3>
					{/if}
					<div class="space-y-1.5">
						{#each category.items as item}
							{@const isActive = activeTabId === item.id}
							<button
								onclick={() => openTab(item.id, item.label, item.type)}
								class="w-full flex items-center gap-4 p-3.5 transition-all duration-300 group
								{isActive
									? 'bg-rust text-white shadow-xl shadow-rust/20 border-l-2 border-rust-light'
									: 'text-stone-500 hover:bg-stone-900/50 hover:text-stone-200 border-l-2 border-transparent'}"
								title={item.description}
							>
								<div
									class="transition-all duration-500 {isActive ? 'scale-110 shadow-rust/50' : 'group-hover:rotate-12 group-hover:scale-110 group-hover:text-rust'}"
								>
									<item.icon class="w-4.5 h-4.5" />
								</div>
								{#if isSidebarOpen}
									<div class="flex-1 text-left" transition:fade={{ duration: 100 }}>
										<span class="font-heading text-[11px] font-black tracking-[0.1em] uppercase">{item.label}</span>
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
				class="p-6 border-t border-stone-800 bg-stone-900/20"
				transition:fade={{ duration: 150 }}
			>
				<div class="flex items-center gap-4 text-[9px] font-black uppercase tracking-[0.2em] text-stone-600">
					<div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-emerald-500/50 shadow-lg"></div>
					<span class="font-jetbrains">CLUSTER_SYNC: ACTIVE</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Main Content Area -->
	<div class="flex-1 flex flex-col min-w-0 bg-[#050505]">
		<QueryTabs {tabs} {activeTabId} onSelect={(id) => (activeTabId = id)} onClose={closeTab} />

		<div class="flex-1 overflow-hidden relative">
			{#each tabs as tab (tab.id)}
				<div
					class="absolute inset-0 bg-[#050505] {activeTabId === tab.id
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
							class="h-full overflow-auto bg-[#050505] relative custom-scrollbar"
						>
							<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
							
							<!-- Header -->
							<div
								class="p-8 sm:p-12 border-b border-stone-800 bg-[#0a0a0a]"
							>
								<div class="flex items-center gap-8 mb-2">
									<div
										class="p-5 bg-rust/10 border border-rust/30 shadow-2xl shadow-rust/10 industrial-frame"
									>
										<Database class="w-10 h-10 text-rust-light" />
									</div>
									<div>
										<h1 class="text-4xl sm:text-5xl font-heading font-black tracking-tighter text-white uppercase">
											CLUSTER_STORAGE_OVERVIEW
										</h1>
										<div class="font-jetbrains text-[11px] text-stone-500 mt-2 uppercase tracking-[0.3em] font-black hidden sm:flex items-center gap-3">
											<div class="w-1.5 h-1.5 bg-rust animate-pulse shadow-rust/50 shadow-lg"></div>
											System_Persistence // Global_Data_Cores // Real-time
										</div>
									</div>
								</div>
							</div>

							<!-- Stats Grid -->
							<div class="p-8 sm:p-12">
								<div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-8 mb-12">
									<!-- Database Size Card -->
									<div class="modern-industrial-card glass-panel p-8 !rounded-none">
										<div class="relative">
											<div class="flex items-center justify-between mb-8">
												<div class="p-3 bg-rust/10 border border-rust/20 industrial-frame">
													<HardDrive class="w-7 h-7 text-rust-light" />
												</div>
												<span class="font-jetbrains text-[10px] font-black text-rust-light bg-rust/10 px-3 py-1.5 uppercase tracking-[0.2em] border border-rust/30">Storage</span>
											</div>
											<div class="text-4xl font-heading font-black text-white mb-2 tracking-tighter">
												{formatBytes(dbStats.size_bytes)}
											</div>
											<div class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-bold">Allocated_Buffer_Space</div>
										</div>
									</div>

									<!-- Connections Card -->
									<div class="modern-industrial-card glass-panel p-8 !rounded-none">
										<div class="relative">
											<div class="flex items-center justify-between mb-8">
												<div class="p-3 bg-emerald-500/10 border border-emerald-500/20 industrial-frame">
													<Activity class="w-7 h-7 text-emerald-400" />
												</div>
												<span class="font-jetbrains text-[10px] font-black text-emerald-400 bg-emerald-500/10 px-3 py-1.5 uppercase tracking-[0.2em] border border-emerald-500/30">Active</span>
											</div>
											<div class="text-4xl font-heading font-black text-white mb-2 tracking-tighter">
												{dbStats.connections}
											</div>
											<div class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-bold">Total_Net_Sockets</div>
										</div>
									</div>

									<!-- Uptime Card -->
									<div class="modern-industrial-card glass-panel p-8 !rounded-none">
										<div class="relative">
											<div class="flex items-center justify-between mb-8">
												<div class="p-3 bg-purple-500/10 border border-purple-500/20 industrial-frame">
													<Clock class="w-7 h-7 text-purple-400" />
												</div>
												<span class="font-jetbrains text-[10px] font-black text-purple-400 bg-purple-500/10 px-3 py-1.5 uppercase tracking-[0.2em] border border-purple-500/30">Uptime</span>
											</div>
											<div class="text-4xl font-heading font-black text-white mb-2 tracking-tighter">
												{Math.floor(dbStats.uptime_seconds / 3600)}h
											</div>
											<div class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-bold">
												{Math.floor((dbStats.uptime_seconds % 3600) / 60)}m Persistent_State
											</div>
										</div>
									</div>

									<!-- Version Card -->
									<div class="modern-industrial-card glass-panel p-8 !rounded-none">
										<div class="relative">
											<div class="flex items-center justify-between mb-8">
												<div class="p-3 bg-amber-500/10 border border-amber-500/20 industrial-frame">
													<Server class="w-7 h-7 text-amber-400" />
												</div>
												<span class="font-jetbrains text-[10px] font-black text-amber-400 bg-amber-500/10 px-3 py-1.5 uppercase tracking-[0.2em] border border-amber-500/30">Kernel</span>
											</div>
											<div class="text-4xl font-heading font-black text-white mb-2 tracking-tighter truncate">
												{dbStats.version.split(' ')[0] || 'PGSQL'}
											</div>
											<div class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-bold truncate">
												{dbStats.version.split(',')[0] || 'Build_Active'}
											</div>
										</div>
									</div>
								</div>

								<!-- Quick Actions -->
								<div class="mb-16">
									<h2 class="font-heading font-black text-sm tracking-[0.3em] text-stone-500 mb-8 flex items-center gap-4">
										<Zap class="w-5 h-5 text-rust animate-pulse" />
										CORE_DIRECTIVES
									</h2>
									<div class="grid grid-cols-2 lg:grid-cols-4 gap-6">
										<button
											onclick={() => openTab('sql', 'SQL Editor', 'sql')}
											class="flex flex-col items-center gap-6 p-10 bg-stone-900/40 border border-stone-800 hover:border-rust hover:bg-rust/5 transition-all duration-500 group industrial-frame"
										>
											<div class="p-5 bg-stone-950 border border-stone-800 group-hover:border-rust/40 group-hover:text-rust transition-all group-hover:scale-110 shadow-lg">
												<Terminal class="w-8 h-8 text-stone-600 group-hover:text-rust" />
											</div>
											<span class="font-heading text-[12px] font-black tracking-[0.2em] text-stone-500 group-hover:text-white uppercase transition-colors">Neural_Query</span>
										</button>

										<button
											onclick={() => openTab('functions', 'Functions', 'functions')}
											class="flex flex-col items-center gap-6 p-10 bg-stone-900/40 border border-stone-800 hover:border-rust hover:bg-rust/5 transition-all duration-500 group industrial-frame"
										>
											<div class="p-5 bg-stone-950 border border-stone-800 group-hover:border-rust/40 group-hover:text-rust transition-all group-hover:scale-110 shadow-lg">
												<Code2 class="w-8 h-8 text-stone-600 group-hover:text-rust" />
											</div>
											<span class="font-heading text-[12px] font-black tracking-[0.2em] text-stone-500 group-hover:text-white uppercase transition-colors">Logic_Units</span>
										</button>

										<button
											onclick={() => openTab('backups', 'Backups', 'backups')}
											class="flex flex-col items-center gap-6 p-10 bg-stone-900/40 border border-stone-800 hover:border-rust hover:bg-rust/5 transition-all duration-500 group industrial-frame"
										>
											<div class="p-5 bg-stone-950 border border-stone-800 group-hover:border-rust/40 group-hover:text-rust transition-all group-hover:scale-110 shadow-lg">
												<HardDrive class="w-8 h-8 text-stone-600 group-hover:text-rust" />
											</div>
											<span class="font-heading text-[12px] font-black tracking-[0.2em] text-stone-500 group-hover:text-white uppercase transition-colors">Snapshot_IO</span>
										</button>

										<button
											onclick={() => openTab('roles', 'Roles', 'roles')}
											class="flex flex-col items-center gap-6 p-10 bg-stone-900/40 border border-stone-800 hover:border-rust hover:bg-rust/5 transition-all duration-500 group industrial-frame"
										>
											<div class="p-5 bg-stone-950 border border-stone-800 group-hover:border-rust/40 group-hover:text-rust transition-all group-hover:scale-110 shadow-lg">
												<Shield class="w-8 h-8 text-stone-600 group-hover:text-rust" />
											</div>
											<span class="font-heading text-[12px] font-black tracking-[0.2em] text-stone-500 group-hover:text-white uppercase transition-colors">Auth_Sector</span>
										</button>
									</div>
								</div>

								<!-- Table Stats -->
								{#if tableCounts.length > 0}
									<div class="w-full">
										<h2 class="font-heading font-black text-sm tracking-[0.3em] text-stone-500 mb-8 flex items-center gap-4">
											<BarChart3 class="w-5 h-5 text-rust animate-pulse" />
											SECTOR_DATA_METRICS
										</h2>
										<div
											class="bg-stone-900/30 border border-stone-800 overflow-hidden industrial-frame"
										>
											<div class="overflow-x-auto custom-scrollbar">
												<table class="w-full">
													<thead class="bg-stone-950 border-b border-stone-800">
														<tr>
															<th
																class="px-8 py-5 text-left text-[11px] font-black text-stone-500 uppercase tracking-[0.3em]"
																>Subject_ID</th
															>
															<th
																class="px-8 py-5 text-right text-[11px] font-black text-stone-500 uppercase tracking-[0.3em]"
																>Entity_Count</th
															>
														</tr>
													</thead>
													<tbody class="divide-y divide-stone-800/50">
														{#each tableCounts as table}
															<tr class="hover:bg-rust/5 transition-all duration-300 group">
																<td class="px-8 py-5">
																	<div class="flex items-center gap-5">
																		<Table class="w-4.5 h-4.5 text-stone-700 group-hover:text-rust transition-colors" />
																		<span class="font-jetbrains text-xs font-black text-stone-400 uppercase tracking-widest group-hover:text-stone-200"
																			>{table.name}</span
																		>
																	</div>
																</td>
																<td class="px-8 py-5 text-right">
																	<span class="font-jetbrains text-xs text-rust font-black tracking-widest shadow-rust/20"
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
						<div class="p-12 text-stone-700 flex flex-col items-center justify-center h-full gap-6">
							<div class="p-8 bg-stone-900/20 border border-stone-800 industrial-frame">
								<FileText class="w-20 h-20 opacity-10" />
							</div>
							<p class="font-heading font-black text-sm tracking-[0.4em] uppercase animate-pulse">Neural_Buffer_Upgrading</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	/* Custom scrollbar for sidebar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #2a2a2a;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
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
