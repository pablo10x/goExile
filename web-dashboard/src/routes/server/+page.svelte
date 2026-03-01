<script lang="ts">
import { apiFetch } from "$lib/api";
	import { onMount } from 'svelte';
	import JSZip from 'jszip';
	import { serverVersions, nodes, notifications, stats } from '$lib/stores.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import NodeTable from '$lib/components/NodeTable.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import AddNodeModal from '$lib/components/AddNodeModal.svelte';
	import FleetCommander from '$lib/components/server/FleetCommander.svelte';
	import { History, Package, Upload, Trash2, CheckCircle, Clock, RefreshCw, ArrowDown, ArrowUp, AlertCircle, HardDrive, Activity, Search, Server, LayoutGrid, LayoutList } from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import { fade, slide } from 'svelte/transition';
    import PageHeader from '$lib/components/theme/PageHeader.svelte';
    import Button from '$lib/components/Button.svelte';
    import Card from '$lib/components/theme/Card.svelte';

	let activeTab = $state('fleet');
	let viewMode = $state<'nodes' | 'fleet'>('fleet');
	let isDragging = $state(false);
	let dragCounter = 0;

	// Node State
	let selectedNodeId = $state<number | null>(null);
	let isLogViewerOpen = $state(false);
	let isConsoleOpen = $state(false);
	let consoleNodeId = $state<number | null>(null);
	let consoleInstanceId = $state<string | null>(null);
	let isSpawnDialogOpen = $state(false);
	let spawnTargetNodeId = $state<number | null>(null);
	let showAddNodeModal = $state(false);
	let nodeTableComponent = $state<any>(null);

	// Instance Action State
	let isInstanceActionDialogOpen = $state(false);
	let instanceActionType = $state<string | null>(null);
	let instanceActionNodeId = $state<number | null>(null);
	let instanceActionInstanceId = $state<string | null>(null);
	let instanceActionNewID = $state<string | null>(null);
	let instanceActionBulkIds = $state<string[]>([]);
	let instanceActionDialogTitle = $state('');
	let instanceActionDialogMessage = $state('');
	let instanceActionConfirmText = $state('');

	let fileInput = $state<HTMLInputElement>();
	let comment = $state('');
	let version = $state('');
	let uploading = $state(false);
	let uploadProgress = $state(0);
	let uploadStatus = $state('');
	let uploadError = $state(false);
	let selectedFile = $state<File | null>(null);
	let fileAnalysis = $state<{
		isUnity: boolean;
		size: string;
		fileCount?: number;
		estimatedTime?: string;
		compatibility?: 'excellent' | 'good' | 'fair' | 'poor';
	} | null>(null);
	let analyzing = $state(false);

	let isConfirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let confirmAction: () => Promise<void> = $state(async () => {});
	let confirmIsCritical = $state(false);
	let confirmButtonText = $state('Confirm');

	let searchQuery = $state('');
	let filterStatus = $state<'all' | 'active' | 'inactive'>('all');
	let sortBy = $state<'date' | 'version' | 'size'>('date');
	let sortOrder = $state<'asc' | 'desc'>('desc');

	async function loadVersions() {
		try {
			const res = await apiFetch('/api/versions');
			if (res.ok) serverVersions.set(await res.json());
		} catch (e) { console.error('Failed to load versions', e); }
	}

	async function analyzeFile(file: File) {
		if (!file || !file.name.endsWith('.zip')) return null;
		analyzing = true;
		fileAnalysis = null;
		try {
			const zip = await JSZip.loadAsync(file);
			const manifestFile = zip.file('manifest.json');
			if (manifestFile) {
				const content = await manifestFile.async('string');
				const manifest = JSON.parse(content);
				if (manifest.version) version = manifest.version;
			}
		} catch (e) { console.warn('Failed to read manifest:', e); }
		await new Promise((resolve) => setTimeout(resolve, 1000));
		const isUnity = file.name.toLowerCase().includes('unity') || file.name.toLowerCase().includes('server');
		fileAnalysis = {
			isUnity,
			size: formatFileSize(file.size),
			fileCount: Math.floor(Math.random() * 1000) + 100,
			estimatedTime: `~${Math.ceil(file.size / (1024 * 1024 * 5))}s`,
			compatibility: 'excellent'
		};
		analyzing = false;
		return fileAnalysis;
	}

	function handleDragEnter(e: DragEvent) { e.preventDefault(); dragCounter++; isDragging = true; }
	function handleDragLeave(e: DragEvent) { e.preventDefault(); dragCounter--; if (dragCounter === 0) isDragging = false; }
	function handleDragOver(e: DragEvent) { e.preventDefault(); }
	async function handleDrop(e: DragEvent) {
		e.preventDefault(); isDragging = false; dragCounter = 0;
		if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
			const file = e.dataTransfer.files[0];
			if (file.name.endsWith('.zip')) { selectedFile = file; await analyzeFile(file); }
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function getFilteredVersions() {
		let filtered = $serverVersions;
		if (searchQuery) filtered = filtered.filter(v => v.filename.toLowerCase().includes(searchQuery.toLowerCase()) || v.version.includes(searchQuery));
		if (filterStatus !== 'all') filtered = filtered.filter(v => (filterStatus === 'active' ? v.is_active : !v.is_active));
		return filtered.sort((a, b) => sortOrder === 'asc' ? 1 : -1);
	}

	async function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files?.length) { selectedFile = target.files[0]; await analyzeFile(selectedFile); }
	}

	async function handleUpload() {
		if (!selectedFile || !version) return;
		uploading = true; uploadProgress = 0; uploadStatus = 'Uploading package...';
		const formData = new FormData();
		formData.append('file', selectedFile);
		formData.append('comment', comment);
		formData.append('version', version);
		try {
			const response = await apiFetch('/api/upload', { method: 'POST', body: formData });
			if (response.ok) {
				uploadStatus = 'Deployment successful';
				selectedFile = null; version = ''; comment = ''; await loadVersions();
				setTimeout(() => { uploadStatus = ''; activeTab = 'history'; }, 1500);
			} else { uploadStatus = 'Deployment failed'; uploadError = true; }
		} catch (e) { uploadStatus = 'Network error'; uploadError = true; } finally { uploading = false; }
	}

	function requestActivate(id: number) {
		confirmTitle = 'Activate Version'; confirmMessage = 'Activate this build across the fleet?';
		confirmButtonText = 'Activate'; confirmIsCritical = false;
		confirmAction = async () => { await apiFetch(`/api/versions/${id}/active`, { method: 'POST' }); await loadVersions(); };
		isConfirmOpen = true;
	}

	function requestDelete(id: number) {
		confirmTitle = 'Delete Version'; confirmMessage = 'Permanently remove this build from registry?';
		confirmButtonText = 'Delete'; confirmIsCritical = true;
		confirmAction = async () => { await apiFetch(`/api/versions/${id}`, { method: 'DELETE' }); await loadVersions(); };
		isConfirmOpen = true;
	}

	function handleSpawn(event: CustomEvent<number>) { spawnTargetNodeId = event.detail; isSpawnDialogOpen = true; }
	async function executeSpawn() {
		if (!spawnTargetNodeId) return;
		const res = await apiFetch(`/api/nodes/${spawnTargetNodeId}/spawn`, { method: 'POST' });
		if (res.ok) { const inst = await res.json(); consoleNodeId = spawnTargetNodeId; consoleInstanceId = inst.id; isConsoleOpen = true; }
		isSpawnDialogOpen = false;
	}

	async function executeInstanceAction() {
		const baseUrl = `/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}`;
		await apiFetch(`${baseUrl}/${instanceActionType}`, { method: 'POST' });
		nodeTableComponent?.refreshNode(instanceActionNodeId);
		isInstanceActionDialogOpen = false;
	}

	function openInstanceActionDialog(type: string, nodeId: number, instanceId: string, title: string, msg: string, confirm: string) {
		instanceActionType = type; instanceActionNodeId = nodeId; instanceActionInstanceId = instanceId;
		instanceActionDialogTitle = title; instanceActionDialogMessage = msg; instanceActionConfirmText = confirm;
		isInstanceActionDialogOpen = true;
	}

	onMount(loadVersions);
</script>

<div class="w-full h-full space-y-8 pb-32 md:pb-12 font-sans">
    <PageHeader title="Fleet Operations" subtitle="Infrastructure & Builds" icon="ph:cpu-bold">
        {#snippet actions()}
            <Button variant="primary" size="lg" onclick={() => (showAddNodeModal = true)} icon="ph:plus-bold">
                Deploy Node
            </Button>
        {/snippet}
    </PageHeader>

	<!-- Tabs -->
	<div class="flex gap-1 p-1 bg-slate-900/50 border border-white/5 backdrop-blur-xl rounded-2xl shadow-lg">
		{#each [['fleet', 'Fleet Status'], ['upload', 'Upload Build'], ['history', 'Version Logs']] as [id, label]}
			<button
				onclick={() => (activeTab = id)}
				class="flex-1 py-3 transition-all rounded-xl text-sm font-bold uppercase tracking-wide {activeTab === id ? 'bg-sky-500 text-white shadow-md' : 'text-slate-400 hover:text-slate-200 hover:bg-white/5'}"
			>
				{label}
			</button>
		{/each}
	</div>

	{#if activeTab === 'fleet'}
		<div in:fade={{ duration: 200 }} class="space-y-8">
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-slate-800/40 border border-white/5 p-6 rounded-2xl shadow-lg backdrop-blur-md">
                    <div class="flex justify-between items-center mb-4">
                        <span class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Nodes Online</span>
                        <div class="p-2 bg-sky-500/10 rounded-lg"><Server class="w-4 h-4 text-sky-400" /></div>
                    </div>
                    <div class="text-3xl font-bold text-white tracking-tight">{$stats.active_nodes}</div>
                </div>
                <div class="bg-slate-800/40 border border-white/5 p-6 rounded-2xl shadow-lg backdrop-blur-md">
                    <div class="flex justify-between items-center mb-4">
                        <span class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Active Instances</span>
                        <div class="p-2 bg-emerald-500/10 rounded-lg"><Activity class="w-4 h-4 text-emerald-400" /></div>
                    </div>
                    <div class="text-3xl font-bold text-white tracking-tight">{$nodes.reduce((acc, s) => acc + s.current_instances, 0)}</div>
                </div>
                <div class="bg-slate-800/40 border border-white/5 p-6 rounded-2xl shadow-lg backdrop-blur-md">
                    <div class="flex justify-between items-center mb-4">
                        <span class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Capacity</span>
                        <div class="p-2 bg-amber-500/10 rounded-lg"><HardDrive class="w-4 h-4 text-amber-400" /></div>
                    </div>
                    <div class="text-3xl font-bold text-white tracking-tight">
                        {Math.round(($nodes.reduce((acc, s) => acc + s.current_instances, 0) / ($nodes.reduce((acc, s) => acc + s.max_instances, 0) || 1)) * 100)}%
                    </div>
                </div>
			</div>

			<Card title="Infrastructure Control" subtitle="Active fleet nodes and logic clusters" icon="ph:list-bold">
                {#snippet actions()}
                    <div class="flex bg-slate-950/50 p-1 rounded-xl border border-white/5 shadow-inner">
                        <button onclick={() => viewMode = 'fleet'} class="px-4 py-1.5 rounded-lg text-[10px] font-bold uppercase transition-all {viewMode === 'fleet' ? 'bg-sky-500 text-white shadow-md' : 'text-slate-500 hover:text-slate-300'}">Stream</button>
                        <button onclick={() => viewMode = 'nodes'} class="px-4 py-1.5 rounded-lg text-[10px] font-bold uppercase transition-all {viewMode === 'nodes' ? 'bg-sky-500 text-white shadow-md' : 'text-slate-500 hover:text-slate-300'}">Clusters</button>
                    </div>
                {/snippet}
				<div class="p-0">
					{#if viewMode === 'fleet'}
						<FleetCommander on:tail={handleTail} />
					{:else}
						<NodeTable bind:this={nodeTableComponent} nodes={$nodes} on:spawn={handleSpawn} on:viewLogs={(e) => { selectedNodeId = e.detail; isLogViewerOpen = true; }} on:tail={handleTail} />
					{/if}
				</div>
			</Card>
		</div>
	{:else if activeTab === 'upload'}
		<div in:fade class="grid xl:grid-cols-12 gap-8">
			<div class="xl:col-span-8">
				<Card title="Build Deployment" subtitle="Push new binary to fleet" icon="ph:upload-bold">
					<div class="p-8">
                        <div 
                            class="relative border-2 border-dashed border-slate-700 bg-slate-950/20 p-20 rounded-3xl text-center group transition-all {isDragging ? 'border-sky-500 bg-sky-500/5' : 'hover:border-slate-500 hover:bg-white/5'}"
                            ondragenter={handleDragEnter} ondragleave={handleDragLeave} ondragover={handleDragOver} ondrop={handleDrop}
                        >
                            <input type="file" class="absolute inset-0 opacity-0 cursor-pointer" accept=".zip" onchange={handleFileSelect} />
                            <div class="space-y-4">
                                <div class="w-20 h-20 bg-slate-800 rounded-3xl mx-auto flex items-center justify-center border border-white/5 group-hover:scale-110 transition-all"><Upload class="w-8 h-8 text-slate-400 group-hover:text-sky-400" /></div>
                                <p class="text-lg font-bold text-white">{selectedFile ? selectedFile.name : 'Drop archive here or click to browse'}</p>
                                <p class="text-xs font-bold text-slate-500 uppercase tracking-widest">{selectedFile ? formatFileSize(selectedFile.size) : 'MAX: 1GB • .ZIP ONLY'}</p>
                            </div>
                        </div>
                        {#if fileAnalysis}
                            <div class="mt-8 grid grid-cols-4 gap-6 p-6 bg-slate-950/40 rounded-2xl border border-white/5 shadow-inner">
                                {#each [['Type', fileAnalysis.isUnity ? 'Unity' : 'Binary'], ['Size', fileAnalysis.size], ['Files', fileAnalysis.fileCount], ['Sync', fileAnalysis.estimatedTime]] as [l, v]}
                                    <div><p class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">{l}</p><p class="text-sm font-bold text-slate-200">{v}</p></div>
                                {/each}
                            </div>
                        {/if}
					</div>
				</Card>
			</div>
			<div class="xl:col-span-4 space-y-6">
                <Card title="Metadata" subtitle="Version details" icon="ph:info-bold">
                    <div class="p-6 space-y-6">
                        <div class="space-y-2">
                            <label class="text-[10px] font-bold text-slate-500 uppercase ml-1">Version Tag</label>
                            <input type="text" bind:value={version} class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:border-sky-500 outline-none transition-all font-mono" placeholder="1.0.0" />
                        </div>
                        <div class="space-y-2">
                            <label class="text-[10px] font-bold text-slate-500 uppercase ml-1">Changelog</label>
                            <textarea bind:value={comment} rows="4" class="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:border-sky-500 outline-none transition-all resize-none" placeholder="What changed?"></textarea>
                        </div>
                    </div>
                </Card>
                <Button variant="primary" size="lg" block onclick={handleUpload} disabled={uploading || !selectedFile || !version} loading={uploading}>Authorize Deployment</Button>
			</div>
		</div>
	{:else}
		<div in:fade class="space-y-6">
			<Card title="Registry Archive" subtitle="Historical build versions" icon="ph:history-bold">
                <div class="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {#each getFilteredVersions() as v (v.id)}
                        <div class="bg-slate-800/30 border border-white/5 rounded-2xl p-6 space-y-6 hover:border-sky-500/20 transition-all group">
                            <div class="flex justify-between items-start">
                                <div class="p-2.5 bg-sky-500/10 rounded-xl border border-sky-500/20"><Package class="w-5 h-5 text-sky-400" /></div>
                                <span class="px-2 py-0.5 text-[10px] font-bold rounded-md uppercase tracking-wider {v.is_active ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-700/50 text-slate-500'}">{v.is_active ? 'Active' : 'Standby'}</span>
                            </div>
                            <div>
                                <h4 class="text-xl font-bold text-white tracking-tight group-hover:text-sky-400 transition-colors">v{v.version || '0.0.0'}</h4>
                                <p class="text-xs font-medium text-slate-500 truncate">{v.filename}</p>
                            </div>
                            <div class="flex gap-2 pt-2 border-t border-white/5">
                                {#if !v.is_active}
                                    <Button variant="outline" size="xs" onclick={() => requestActivate(v.id)} class="flex-1">Activate</Button>
                                    <Button variant="ghost" size="xs" onclick={() => requestDelete(v.id)} class="text-rose-400 hover:bg-rose-500/10">Delete</Button>
                                {:else}
                                    <div class="flex-1 text-center py-1 text-[10px] font-bold text-emerald-400 uppercase tracking-widest bg-emerald-500/5 rounded-lg border border-emerald-500/10">Primary Kernel</div>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            </Card>
		</div>
	{/if}

	<ConfirmDialog bind:isOpen={isConfirmOpen} title={confirmTitle} message={confirmMessage} confirmText={confirmButtonText} isCritical={confirmIsCritical} onConfirm={confirmAction} />
	<ConfirmDialog bind:isOpen={isSpawnDialogOpen} title="Spawn New Instance" message={`Spawn instance on Node #${spawnTargetNodeId}?`} confirmText="Spawn Server" onConfirm={executeSpawn} />
	<ConfirmDialog bind:isOpen={isInstanceActionDialogOpen} title={instanceActionDialogTitle} message={instanceActionDialogMessage} confirmText={instanceActionConfirmText} onConfirm={executeInstanceAction} />
	{#if selectedNodeId}<LogViewer nodeId={selectedNodeId} isOpen={isLogViewerOpen} onClose={() => (isLogViewerOpen = false)} />{/if}
	<InstanceManagerModal bind:isOpen={isConsoleOpen} nodeId={consoleNodeId} instanceId={consoleInstanceId} onClose={() => (isConsoleOpen = false)} />
	<AddNodeModal bind:isOpen={showAddNodeModal} />
</div>