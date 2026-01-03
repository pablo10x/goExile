<script lang="ts">
	import { onMount } from 'svelte';
	import JSZip from 'jszip';
	import { serverVersions, nodes, notifications, stats } from '$lib/stores.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import NodeTable from '$lib/components/NodeTable.svelte';
	import LogViewer from '$lib/components/LogViewer.svelte';
	import InstanceManagerModal from '$lib/components/InstanceManagerModal.svelte';
	import AddNodeModal from '$lib/components/AddNodeModal.svelte';
	import FleetCommander from '$lib/components/server/FleetCommander.svelte';
	import { setContext } from 'svelte';
	import { History, Package, Upload, Play, Trash2, CheckCircle, Clock, RefreshCw, ArrowDownToLine, ArrowDown, ArrowUp, AlertCircle, HardDrive, Activity, Search, Cpu, List, Plus, Server, LayoutGrid, LayoutList } from 'lucide-svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import { fade } from 'svelte/transition';

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
		unityVersion?: string;
		platform?: string;
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
	let showAdvancedFilters = $state(false);

	async function loadVersions() {
		try {
			const res = await fetch('/api/versions');
			if (res.ok) {
				serverVersions.set(await res.json());
			}
		} catch (e) {
			console.error('Failed to load versions', e);
		}
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
				if (manifest.version) {
					version = manifest.version;
				}
			}
		} catch (e) {
			console.warn('Failed to read manifest:', e);
		}

		// Simulate enhanced file analysis
		await new Promise((resolve) => setTimeout(resolve, 1500));

		const isUnity =
			file.name.toLowerCase().includes('unity') ||
			file.name.toLowerCase().includes('server') ||
			file.name.toLowerCase().includes('game');

		const fileSizeMB = file.size / (1024 * 1024);
		const estimatedTime = Math.ceil(fileSizeMB / 10); // Rough estimate: 10MB/s

		// Determine compatibility based on file characteristics
		let compatibility: 'excellent' | 'good' | 'fair' | 'poor' = 'good';
		if (isUnity && file.name.includes('2022')) compatibility = 'excellent';
		if (fileSizeMB > 500) compatibility = 'fair';
		if (fileSizeMB > 1000) compatibility = 'poor';

		fileAnalysis = {
			isUnity,
			unityVersion: isUnity ? '2022.3.0f1' : undefined,
			platform: 'Windows',
			size: formatFileSize(file.size),
			fileCount: Math.floor(Math.random() * 2000) + 500,
			estimatedTime: `~${estimatedTime}s`,
			compatibility
		};

		analyzing = false;
		return fileAnalysis;
	}

	function handleDragEnter(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragCounter++;
		if (e.dataTransfer?.items && e.dataTransfer.items.length > 0) {
			isDragging = true;
		}
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragCounter--;
		if (dragCounter === 0) {
			isDragging = false;
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
	}

	async function handleDrop(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDragging = false;
		dragCounter = 0;

		if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
			const file = e.dataTransfer.files[0];
			if (file.name.endsWith('.zip')) {
				selectedFile = file;
				if (fileInput) fileInput.files = e.dataTransfer.files;
				uploadStatus = '';
				await analyzeFile(file);
			} else {
				uploadStatus = 'Only .zip files are allowed.';
				uploadError = true;
			}
		}
	}

	function getCompatibilityColor(compatibility: string) {
		switch (compatibility) {
			case 'excellent':
				return 'text-success border-emerald-500/20 bg-success/5';
			case 'good':
				return 'text-info border-blue-500/20 bg-blue-500/5';
			case 'fair':
				return 'text-yellow-400 border-yellow-500/20 bg-yellow-500/5';
			case 'poor':
				return 'text-danger border-red-500/20 bg-red-500/5';
			default:
				return 'text-text-dim border-slate-500/20 bg-slate-500/5';
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function getFilteredVersions() {
		let filtered = $serverVersions;

		if (searchQuery) {
			filtered = filtered.filter(
				(v) =>
					v.filename.toLowerCase().includes(searchQuery.toLowerCase()) ||
					v.comment?.toLowerCase().includes(searchQuery.toLowerCase()) ||
					v.version.toLowerCase().includes(searchQuery.toLowerCase())
			);
		}

		if (filterStatus !== 'all') {
			filtered = filtered.filter((v) => (filterStatus === 'active' ? v.is_active : !v.is_active));
		}

		return filtered.sort((a, b) => {
			let aVal, bVal;

			switch (sortBy) {
				case 'date':
					aVal = new Date(a.uploaded_at).getTime();
					bVal = new Date(b.uploaded_at).getTime();
					break;
				case 'version':
					aVal = a.version || '0.0.0';
					bVal = b.version || '0.0.0';
					break;
				case 'size':
					aVal = parseInt(a.filename.match(/(\d+)/)?.[1] || '0');
					bVal = parseInt(b.filename.match(/(\d+)/)?.[1] || '0');
					break;
				default:
					aVal = new Date(a.uploaded_at).getTime();
					bVal = new Date(b.uploaded_at).getTime();
			}

			if (sortOrder === 'asc') {
				return aVal > bVal ? 1 : -1;
			} else {
				return aVal < bVal ? 1 : -1;
			}
		});
	}

	async function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			const file = target.files[0];
			selectedFile = file;
			uploadStatus = '';
			await analyzeFile(file);
		}
	}

	onMount(() => {
		loadVersions();
	});

	async function handleUpload() {
		if (!fileInput?.files || fileInput.files.length === 0) {
			uploadStatus = 'Please select a file first.';
			uploadError = true;
			return;
		}

		const file = fileInput.files[0];
		if (!file.name.endsWith('.zip')) {
			uploadStatus = 'Only .zip files are allowed.';
			uploadError = true;
			return;
		}

		// Validate version format
		const semverRegex = /^\d+\.\d+\.\d+$/;
		if (!semverRegex.test(version.trim())) {
			uploadStatus = 'Version must be in semantic versioning format (e.g. 1.2.3).';
			uploadError = true;
			return;
		}

		uploading = true;
		uploadProgress = 0;
		uploadStatus = 'Initializing upload...';
		uploadError = false;

		const formData = new FormData();
		formData.append('file', file);
		formData.append('comment', comment);
		formData.append('version', version);

		try {
			// Simulate upload progress
			const progressInterval = setInterval(() => {
				if (uploadProgress < 90) {
					uploadProgress += Math.random() * 15;
					uploadStatus = `Uploading... ${Math.round(uploadProgress)}%`;
				}
			}, 200);

			const response = await fetch('/api/upload', {
				method: 'POST',
				body: formData
			});

			clearInterval(progressInterval);
			uploadProgress = 100;

			if (response.ok) {
				uploadStatus = 'Upload successful! Processing file...';
				uploadError = false;
				if (fileInput) fileInput.value = '';
				selectedFile = null;
				fileAnalysis = null;
				comment = '';
				version = '';
				await loadVersions();

				setTimeout(() => {
					uploadStatus = '';
					uploadProgress = 0;
					activeTab = 'history';
				}, 1500);
			} else {
				uploadStatus = 'Upload failed. Server responded with ' + response.status;
				uploadError = true;
				uploadProgress = 0;
			}
		} catch (e) {
			uploadStatus = 'Upload failed due to network error.';
			uploadError = true;
			uploadProgress = 0;
		} finally {
			uploading = false;
		}
	}

	function requestActivate(id: number) {
		confirmTitle = 'Activate Version';
		confirmMessage =
			'Are you sure you want to activate this version? Nodes will download it on next check.';
		confirmButtonText = 'Activate';
		confirmIsCritical = false;
		confirmAction = async () => await activateVersion(id);
		isConfirmOpen = true;
	}

	async function activateVersion(id: number) {
		try {
			const res = await fetch(`/api/versions/${id}/active`, { method: 'POST' });
			if (res.ok) {
				loadVersions();
			} else {
				throw new Error('Failed to activate version');
			}
		} catch (e) {
			console.error(e);
			throw e; // Propagate to dialog for error handling
		}
	}

	function requestDelete(id: number) {
		confirmTitle = 'Delete Version';
		confirmMessage = 'Are you sure you want to delete this version? This action cannot be undone.';
		confirmButtonText = 'Delete';
		confirmIsCritical = true;
		confirmAction = async () => await deleteVersion(id);
		isConfirmOpen = true;
	}

	async function deleteVersion(id: number) {
		try {
			const res = await fetch(`/api/versions/${id}`, { method: 'DELETE' });
			if (res.ok) {
				await loadVersions();
			} else {
				const data = await res.json();
				throw new Error(data.error || 'Failed to delete version');
			}
		} catch (e) {
			console.error(e);
			throw e;
		}
	}

	// --- Node Hub Logic ---
	function handleSpawn(event: CustomEvent<number>) {
		spawnTargetNodeId = event.detail;
		isSpawnDialogOpen = true;
	}

	async function executeSpawn() {
		if (!spawnTargetNodeId) return;
		try {
			const res = await fetch(`/api/nodes/${spawnTargetNodeId}/spawn`, { method: 'POST' });
			if (!res.ok) throw new Error('Spawn failed');
			const instance = await res.json();
			consoleNodeId = spawnTargetNodeId;
			consoleInstanceId = instance.id;
			isConsoleOpen = true;
		} catch (e) {
			notifications.add({ type: 'error', message: 'Failed to spawn instance' });
		}
		isSpawnDialogOpen = false;
	}

	function handleViewLogs(event: CustomEvent<number>) {
		selectedNodeId = event.detail;
		isLogViewerOpen = true;
	}

	function handleTail(event: CustomEvent<{ nodeId: number; instanceId: string }>) {
		consoleNodeId = event.detail.nodeId;
		consoleInstanceId = event.detail.instanceId;
		isConsoleOpen = true;
	}

	async function executeInstanceAction() {
		if (!instanceActionNodeId || !instanceActionType) return;
		try {
			let res: Response;
			const baseUrl = `/api/nodes/${instanceActionNodeId}/instances/${instanceActionInstanceId}`;
			
			if (instanceActionType === 'start') res = await fetch(`${baseUrl}/start`, { method: 'POST' });
			else if (instanceActionType === 'stop') res = await fetch(`${baseUrl}/stop`, { method: 'POST' });
			else if (instanceActionType === 'delete') res = await fetch(baseUrl, { method: 'DELETE' });
			else if (instanceActionType === 'update') res = await fetch(`${baseUrl}/update`, { method: 'POST' });
			else if (instanceActionType === 'rename') {
				res = await fetch(`${baseUrl}/rename`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ new_id: instanceActionNewID })
				});
			} else if (instanceActionType === 'restart') {
				await fetch(`${baseUrl}/stop`, { method: 'POST' });
				res = await fetch(`${baseUrl}/start`, { method: 'POST' });
			} else if (instanceActionType.startsWith('bulk_')) {
				const action = instanceActionType.replace('bulk_', '');
				const promises = instanceActionBulkIds.map(async (id) => {
					const url = `/api/nodes/${instanceActionNodeId}/instances/${id}/${action === 'restart' ? 'stop' : action}`;
					await fetch(url, { method: 'POST' });
					if (action === 'restart') await fetch(`/api/nodes/${instanceActionNodeId}/instances/${id}/start`, { method: 'POST' });
				});
				await Promise.all(promises);
				res = { ok: true } as Response;
			} else res = { ok: false } as Response;

			if (res.ok) {
				nodeTableComponent?.refreshNode(instanceActionNodeId);
				notifications.add({ type: 'success', message: 'Action executed successfully' });
			} else {
				throw new Error('Action failed');
			}
		} catch (e) {
			notifications.add({ type: 'error', message: 'Operation failed' });
		}
		isInstanceActionDialogOpen = false;
	}

	function openInstanceActionDialog(type: string, nodeId: number, instanceId: string, title: string, msg: string, confirm: string) {
		instanceActionType = type;
		instanceActionNodeId = nodeId;
		instanceActionInstanceId = instanceId;
		instanceActionDialogTitle = title;
		instanceActionDialogMessage = msg;
		instanceActionConfirmText = confirm;
		isInstanceActionDialogOpen = true;
	}
</script>

<div class="w-full h-full space-y-10 pb-32 md:pb-12">
	<div class="flex justify-between items-center mb-10">
		<div class="flex items-center gap-6">
			<div class="p-4 bg-rust/10 border border-rust/30 industrial-frame shadow-2xl">
				<Icon name="cpu" size="2.5rem" class="text-rust-light" />
			</div>
			<div>
				<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter">
					Node Fleet Operations
				</h1>
				<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest font-black mt-2">Centralized command for nodes and binaries</p>
			</div>
		</div>
		
		<div class="flex items-center gap-4">
			<button
				onclick={() => (showAddNodeModal = true)}
				class="hidden md:flex items-center gap-3 px-6 py-3 bg-white text-black font-heading font-black text-[10px] uppercase tracking-widest hover:bg-rust hover:text-white transition-all industrial-frame shadow-xl"
			>
				<Icon name="ph:plus-bold" />
				Deploy_Node
			</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1.5 p-1.5 bg-[var(--header-bg)]/80 border border-stone-800 backdrop-blur-xl industrial-frame shadow-2xl">
		<button
			onclick={() => (activeTab = 'fleet')}
			class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'fleet'
				? 'bg-rust text-white shadow-lg'
				: 'text-text-dim hover:text-white hover:bg-stone-900'}"
		>
			<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">Fleet Status</span>
			<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">Active Nodes</span>
		</button>
		<button
			onclick={() => (activeTab = 'upload')}
			class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'upload'
				? 'bg-rust text-white shadow-lg'
				: 'text-text-dim hover:text-white hover:bg-stone-900'}"
		>
			<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">Upload Build</span>
			<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">New Binary</span>
		</button>
		<button
			onclick={() => (activeTab = 'history')}
			class="flex-1 flex flex-col items-center gap-1.5 px-8 py-4 transition-all {activeTab === 'history'
				? 'bg-rust text-white shadow-lg'
				: 'text-text-dim hover:text-white hover:bg-stone-900'}"
		>
			<span class="font-heading font-black text-[12px] uppercase tracking-[0.2em]">Build History</span>
			<span class="font-jetbrains text-[8px] font-black opacity-40 uppercase tracking-widest">Version Logs</span>
		</button>
	</div>

	{#if activeTab === 'fleet'}
		<div in:fade={{ duration: 200 }} class="space-y-8">
			<div class="flex flex-col md:flex-row justify-between items-end gap-6">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6 flex-1 w-full">
					<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl">
						<div class="flex justify-between items-center mb-4">
							<span class="text-[10px] font-black text-text-dim uppercase tracking-widest">Nodes Active</span>
							<Server class="w-4 h-4 text-rust" />
						</div>
						<div class="text-4xl font-heading font-black text-white tracking-tighter">{$stats.active_nodes}</div>
					</div>
					<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl">
						<div class="flex justify-between items-center mb-4">
							<span class="text-[10px] font-black text-text-dim uppercase tracking-widest">Total Instances</span>
							<Activity class="w-4 h-4 text-success" />
						</div>
						<div class="text-4xl font-heading font-black text-white tracking-tighter">{$nodes.reduce((acc, s) => acc + s.current_instances, 0)}</div>
					</div>
					<div class="bg-stone-900/40 border border-stone-800 p-6 industrial-frame shadow-xl">
						<div class="flex justify-between items-center mb-4">
							<span class="text-[10px] font-black text-text-dim uppercase tracking-widest">Capacity Used</span>
							<HardDrive class="w-4 h-4 text-warning" />
						</div>
						<div class="text-4xl font-heading font-black text-white tracking-tighter">
							{Math.round(($nodes.reduce((acc, s) => acc + s.current_instances, 0) / ($nodes.reduce((acc, s) => acc + s.max_instances, 0) || 1)) * 100)}%
						</div>
					</div>
				</div>

				<!-- View Switcher -->
				<div class="flex bg-slate-950 border border-slate-800 p-1 rounded-xl shadow-inner">
					<button 
						onclick={() => viewMode = 'fleet'}
						class="flex items-center gap-2 px-4 py-2 rounded-lg transition-all {viewMode === 'fleet' ? 'bg-indigo-500 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'}"
					>
						<LayoutList class="w-4 h-4" />
						<span class="text-[10px] font-black uppercase tracking-widest">Tactical_Stream</span>
					</button>
					<button 
						onclick={() => viewMode = 'nodes'}
						class="flex items-center gap-2 px-4 py-2 rounded-lg transition-all {viewMode === 'nodes' ? 'bg-indigo-500 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'}"
					>
						<LayoutGrid class="w-4 h-4" />
						<span class="text-[10px] font-black uppercase tracking-widest">Node_Clusters</span>
					</button>
				</div>
			</div>

			{#if viewMode === 'fleet'}
				<div in:fade>
					<FleetCommander 
						on:tail={handleTail}
						on:start={(e) => openInstanceActionDialog('start', e.detail.nodeId, e.detail.instanceId, 'Start Instance', `Initialize execution of ${e.detail.instanceId}?`, 'Confirm Start')}
						on:stop={(e) => openInstanceActionDialog('stop', e.detail.nodeId, e.detail.instanceId, 'Stop Instance', `Terminate execution of ${e.detail.instanceId}?`, 'Confirm Stop')}
						on:restart={(e) => openInstanceActionDialog('restart', e.detail.nodeId, e.detail.instanceId, 'Restart Instance', `Reboot instance ${e.detail.instanceId}?`, 'Confirm Restart')}
					/>
				</div>
			{:else}
				<div class="modern-industrial-card glass-panel !rounded-none overflow-hidden border-stone-800 shadow-2xl">
					<NodeTable
						bind:this={nodeTableComponent}
						nodes={$nodes}
						on:spawn={handleSpawn}
						on:viewLogs={handleViewLogs}
						on:tail={handleTail}
						on:startInstanceRequest={(e) => openInstanceActionDialog('start', e.detail.nodeId, e.detail.instanceId, 'Start Instance', `Initialize execution of ${e.detail.instanceId}?`, 'Confirm Start')}
						on:stopInstanceRequest={(e) => openInstanceActionDialog('stop', e.detail.nodeId, e.detail.instanceId, 'Stop Instance', `Terminate execution of ${e.detail.instanceId}?`, 'Confirm Stop')}
						on:restartInstanceRequest={(e) => openInstanceActionDialog('restart', e.detail.nodeId, e.detail.instanceId, 'Restart Instance', `Reboot instance ${e.detail.instanceId}?`, 'Confirm Restart')}
						on:deleteInstanceRequest={(e) => openInstanceActionDialog('delete', e.detail.nodeId, e.detail.instanceId, 'Delete Instance', `Permanently purge ${e.detail.instanceId}?`, 'Confirm Purge')}
						on:updateInstanceRequest={(e) => openInstanceActionDialog('update', e.detail.nodeId, e.detail.instanceId, 'Update Instance', `Reinstall build for ${e.detail.instanceId}?`, 'Confirm Update')}
						on:bulkInstanceActionRequest={(e) => {
							instanceActionType = `bulk_${e.detail.action}`;
							instanceActionNodeId = e.detail.nodeId;
							instanceActionBulkIds = e.detail.instanceIds;
							instanceActionDialogTitle = 'Bulk Operation';
							instanceActionDialogMessage = `Execute ${e.detail.action} on ${e.detail.instanceIds.length} instances?`;
							instanceActionConfirmText = 'Confirm Bulk';
							isInstanceActionDialogOpen = true;
						}}
					/>
				</div>
			{/if}
		</div>
	{:else if activeTab === 'upload'}
		<div class="grid xl:grid-cols-12 gap-8 items-start">
			<!-- Upload Area -->
			<div class="xl:col-span-8">
				<div
					class="modern-industrial-card glass-panel p-10 !rounded-none"
				>
					<div class="mb-10 flex items-center gap-6">
						<div class="p-3 bg-stone-900 border border-stone-800 industrial-frame">
							<Upload class="w-6 h-6 text-rust" />
						</div>
						<div>
							<h3 class="text-2xl font-heading font-black text-white uppercase tracking-tighter">Build Upload</h3>
							<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest mt-1">
								Upload a new <code class="text-rust">game_server.zip</code> package to central registry.
							</p>
						</div>
					</div>

					<!-- Enhanced Drag & Drop Area -->
					<div
						class="relative group"
						ondragenter={handleDragEnter}
						ondragleave={handleDragLeave}
						ondragover={handleDragOver}
						ondrop={handleDrop}
						role="region"
						aria-label="File Upload Drop Zone"
					>
						<div
							class="relative border border-stone-800 bg-black/40 p-16 transition-all duration-500 overflow-hidden {isDragging
								? 'border-rust bg-rust/5 scale-[0.99]'
								: 'hover:border-rust/30'}"
						>
							<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
							
							<input
								id="file-upload"
								name="file-upload"
								type="file"
								class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-20"
								accept=".zip"
								bind:this={fileInput}
								onchange={handleFileSelect}
							/>

							<div class="relative z-10 text-center space-y-8">
								{#if uploading}
									<div class="space-y-6">
										<div
											class="w-20 h-20 mx-auto border-2 border-rust border-t-transparent rounded-none animate-spin shadow-lg shadow-rust/20"
										></div>
										<div class="space-y-4">
											<p class="font-heading font-black text-[11px] text-rust uppercase tracking-[0.3em] animate-pulse">{uploadStatus}</p>
											{#if uploadProgress > 0}
												<div class="w-full max-w-lg mx-auto">
													<div class="h-1 bg-stone-900 border border-stone-800 rounded-none overflow-hidden p-0 relative">
														<div
															class="h-full bg-rust shadow-[0_0_15px_rgba(249,115,22,0.5)] transition-all duration-300"
															style="width: {uploadProgress}%"
														></div>
													</div>
													<p class="font-jetbrains text-[9px] font-black text-text-dim mt-3 uppercase tracking-widest">
														PROGRESS: {Math.round(uploadProgress)}%_SYNCED
													</p>
												</div>
											{/if}
										</div>
									</div>
								{:else if analyzing}
									<div class="space-y-6">
										<div
											class="w-20 h-20 mx-auto border-2 border-amber-500 border-t-transparent rounded-none animate-spin shadow-lg shadow-amber-900/20"
										></div>
										<div class="space-y-2">
											<p class="font-heading font-black text-[11px] text-warning uppercase tracking-[0.3em] animate-pulse">ANALYZING_PAYLOAD...</p>
											<p class="font-jetbrains text-[9px] font-black text-text-dim uppercase tracking-widest">
												Verifying compatibility and sector mapping
											</p>
										</div>
									</div>
								{:else}
									<div class="space-y-6">
										<div
											class="w-24 h-24 mx-auto flex items-center justify-center bg-stone-950 border border-stone-800 industrial-frame group-hover:border-rust group-hover:scale-110 transition-all duration-500 shadow-xl"
										>
											<Upload
												class="w-10 h-10 text-stone-700 group-hover:text-rust transition-colors"
											/>
										</div>
										<div class="space-y-3">
											<p class="font-heading font-black text-lg text-white uppercase tracking-widest">
												{selectedFile
													? selectedFile.name
													: 'DROP_ARCHIVE_OR_ACTIVATE_SELECTOR'}
											</p>
											<p class="font-jetbrains text-[10px] font-bold text-text-dim uppercase tracking-widest">
												{selectedFile
													? formatFileSize(selectedFile.size)
													: 'MAX_LIMIT: 1GB // FORMAT: .ZIP_ONLY'}
											</p>
										</div>
										{#if !selectedFile}
											<div class="pt-4">
												<span class="px-8 py-3 bg-stone-900 border border-stone-800 text-text-dim font-heading font-black text-[10px] uppercase tracking-widest group-hover:border-rust group-hover:text-white transition-all shadow-lg">
													Browse_Buffer
												</span>
											</div>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					</div>

					<!-- File Analysis Results -->
					{#if fileAnalysis && !uploading}
						<div
							class="mt-10 p-8 bg-stone-950 border border-stone-800 industrial-frame shadow-inner"
						>
							<h4 class="font-heading font-black text-xs text-text-dim uppercase tracking-[0.2em] mb-6 flex items-center gap-3">
								<Activity class="w-4 h-4 text-rust" />
								Build Summary
							</h4>
							<div class="grid grid-cols-2 md:grid-cols-4 gap-8">
								<div class="space-y-2">
									<p class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-widest">DATA_TYPE</p>
									<p class="font-jetbrains text-[11px] font-black text-stone-300 uppercase">
										{fileAnalysis.isUnity ? 'UNITY_KERNEL' : 'GENERIC_BINARY'}
									</p>
								</div>
								<div class="space-y-2">
									<p class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-widest">BUFFER_SIZE</p>
									<p class="font-jetbrains text-[11px] font-black text-stone-300 uppercase">
										{fileAnalysis.size}
									</p>
								</div>
								<div class="space-y-2">
									<p class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-widest">FILE_ENTITIES</p>
									<p class="font-jetbrains text-[11px] font-black text-stone-300 uppercase">
										{fileAnalysis.fileCount}
									</p>
								</div>
								<div class="space-y-2">
									<p class="font-jetbrains text-[9px] font-black text-stone-700 uppercase tracking-widest">SYNC_ESTIMATE</p>
									<p class="font-jetbrains text-[11px] font-black text-stone-300 uppercase">
										{fileAnalysis.estimatedTime}
									</p>
								</div>
							</div>
							{#if fileAnalysis.compatibility}
								<div class="mt-8 pt-6 border-t border-stone-900 flex items-center justify-center">
									<span
										class={`px-4 py-1.5 font-jetbrains text-[10px] font-black border uppercase tracking-widest ${getCompatibilityColor(fileAnalysis.compatibility)}`}
									>
										INTEGRITY_INDEX: {fileAnalysis.compatibility}
									</span>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Version Details Panel -->
			<div class="xl:col-span-4 space-y-8">
				<!-- Version Information -->
				<div
					class="modern-industrial-card glass-panel p-8 !rounded-none"
				>
					<h3 class="font-heading font-black text-base text-white uppercase tracking-widest mb-8 border-b border-stone-800 pb-4">Build Metadata</h3>

					<div class="space-y-8">
						<div class="space-y-3">
							<label
								for="version"
								class="block font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest"
							>
								VERSION_IDENTIFIER <span class="text-danger">*</span>
							</label>
							<input
								type="text"
								id="version"
								bind:value={version}
								class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-xs text-stone-200 focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
								placeholder="e.g. 1.0.0"
								required
							/>
							<p class="font-jetbrains text-[8px] text-stone-700 uppercase tracking-widest">Format: Semantic_Versioning (X.Y.Z)</p>
						</div>

						<div class="space-y-3">
							<label
								for="comment"
								class="block font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest"
							>
								PROTOCOL_CHANGELOG
							</label>
							<textarea
								id="comment"
								bind:value={comment}
								rows="6"
								class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-xs text-stone-200 focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner resize-none leading-relaxed"
								placeholder="Describe binary modifications..."
							></textarea>
						</div>
					</div>
				</div>

				<!-- Upload Button & Status -->
				<div class="space-y-6">
					<button
						onclick={handleUpload}
						disabled={uploading || !selectedFile || !version}
						class="w-full px-8 py-4 bg-rust hover:bg-rust-light text-white font-heading font-black text-xs uppercase tracking-[0.2em] shadow-xl shadow-rust/20 transition-all active:translate-y-px disabled:opacity-20 flex items-center justify-center gap-4 industrial-frame"
					>
						{#if uploading}
							<div
								class="w-5 h-5 border-2 border-white border-t-transparent rounded-none animate-spin"
							></div>
							SYNCHRONIZING...
						{:else}
							<Upload class="w-5 h-5" />
							AUTHORIZE_DEPLOYMENT
						{/if}
					</button>

					{#if uploadStatus}
						<div
							class={`p-6 border flex items-start gap-5 industrial-frame ${uploadError ? 'bg-red-950/20 text-danger border-red-900/40 shadow-red-900/10' : 'bg-emerald-950/20 text-success border-emerald-900/40 shadow-emerald-900/10'}`}
						>
							<div class="shrink-0 mt-1">
								{#if uploadError}
									<AlertCircle class="w-6 h-6" />
								{:else}
									<CheckCircle class="w-6 h-6" />
								{/if}
							</div>
							<div>
								<p class="font-heading font-black text-xs uppercase tracking-widest">{uploadError ? 'PROTOCOL_FAULT' : 'BUFFER_SYNC_COMPLETE'}</p>
								<p class="font-jetbrains text-[10px] font-bold opacity-75 mt-2 uppercase leading-relaxed tracking-tight">{uploadStatus}</p>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{:else}
		<div class="space-y-8" transition:fade={{ duration: 200 }}>
			<!-- Filters and Search -->
			<div
				class="modern-industrial-card glass-panel p-8 !rounded-none"
			>
				<div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-8">
					<div class="flex-1 max-w-2xl relative group">
						<Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-text-dim group-focus-within:text-rust transition-colors" />
						<input
							type="text"
							bind:value={searchQuery}
							placeholder="FILTER_ARCHIVE_BY_METADATA..."
							class="w-full bg-stone-950 border border-stone-800 pl-14 pr-4 py-3.5 font-jetbrains text-xs text-stone-200 focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
						/>
					</div>

					<div class="flex flex-wrap items-center gap-4">
						<!-- Status Filter -->
						<select
							bind:value={filterStatus}
							class="bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] font-black text-stone-400 focus:border-rust outline-none cursor-pointer uppercase tracking-widest appearance-none min-w-[160px]"
						>
							<option value="all">ALL_RECORDS</option>
							<option value="all">ACTIVE_ONLY</option>
							<option value="inactive">HALTED_ONLY</option>
						</select>

						<!-- Sort Options -->
						<div class="flex items-center gap-3">
							<select
								bind:value={sortBy}
								class="bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] font-black text-stone-400 focus:border-rust outline-none cursor-pointer uppercase tracking-widest appearance-none min-w-[160px]"
							>
								<option value="date">SORT_BY_TIME</option>
								<option value="version">SORT_BY_REV</option>
								<option value="size">SORT_BY_SIZE</option>
							</select>
							<button
								onclick={() => (sortOrder = sortOrder === 'desc' ? 'asc' : 'desc')}
								class="p-3 bg-stone-950 border border-stone-800 text-text-dim hover:text-rust transition-all active:translate-y-px shadow-lg"
								title="Toggle sort order"
							>
								{#if sortOrder === 'desc'}
									<ArrowDown class="w-5 h-5" />
								{:else}
									<ArrowUp class="w-5 h-5" />
								{/if}
							</button>
						</div>
					</div>
				</div>
			</div>

			<!-- Version Grid -->
			{#if getFilteredVersions().length === 0}
				<div
					class="modern-industrial-card glass-panel p-32 text-center border-dashed !rounded-none bg-stone-950/20"
				>
					<Package class="w-20 h-20 mx-auto text-stone-800 mb-8 opacity-20" />
					<h3 class="font-heading font-black text-xl text-stone-700 uppercase tracking-[0.3em] mb-3">
						Archive_Registry_Empty
					</h3>
					<p class="font-jetbrains text-[10px] font-bold text-text-dim uppercase tracking-widest">
						{searchQuery
							? 'Neural filters returned zero logical matches.'
							: 'Initial deployment binary pending synchronization.'}
					</p>
				</div>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-3 3xl:grid-cols-4 gap-8">
					{#each getFilteredVersions() as version (version.id)}
						<div
							class="modern-industrial-card glass-panel group relative transition-all duration-500 hover:border-rust/40 shadow-xl !rounded-none"
						>
							<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
							
							<div class="p-8 relative z-10 space-y-8">
								<!-- Header with Status -->
								<div class="flex items-start justify-between">
									<div class="p-3 bg-rust/5 border border-rust/20 industrial-frame group-hover:border-rust/40 transition-all shadow-inner">
										<Package class="w-6 h-6 text-rust group-hover:text-rust-light group-hover:scale-110 transition-all" />
									</div>
									
									<div class="flex items-center gap-3">
										{#if version.is_active}
											<span
												class="px-3 py-1 font-jetbrains text-[9px] font-black bg-rust text-white uppercase tracking-[0.2em] shadow-lg shadow-rust/30"
											>
												CORE_ACTIVE
											</span>
										{:else}
											<span
												class="px-3 py-1 font-jetbrains text-[9px] font-black bg-stone-900 border border-stone-800 text-text-dim uppercase tracking-[0.2em]"
											>
												STANDBY
											</span>
										{/if}
									</div>
								</div>

								<!-- Version Info -->
								<div class="space-y-2">
									<h4 class="text-3xl font-heading font-black text-white uppercase tracking-tighter group-hover:text-rust transition-colors duration-500">
										REV_{version.version || 'UNKNOWN'}
									</h4>
									<p class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest truncate">{version.filename}</p>
								</div>

								<!-- Metadata Table -->
								<div class="grid grid-cols-2 gap-6 pt-6 border-t border-stone-800/50">
									<div class="space-y-1.5">
										<span class="block font-jetbrains text-[8px] font-black text-stone-700 uppercase tracking-widest">DEPLOY_DATE</span>
										<span class="font-jetbrains text-[10px] font-black text-stone-400 uppercase tracking-tight">{new Date(version.uploaded_at).toLocaleDateString()}</span>
									</div>
									<div class="space-y-1.5 text-right">
										<span class="block font-jetbrains text-[8px] font-black text-stone-700 uppercase tracking-widest">TIMESTAMP</span>
										<span class="font-jetbrains text-[10px] font-black text-stone-400 uppercase tracking-tight">{new Date(version.uploaded_at).toLocaleTimeString([], { hour12: false })}</span>
									</div>
								</div>

								<!-- Release Notes -->
								{#if version.comment}
									<div class="bg-black/40 p-4 border-l-2 border-stone-800 group-hover:border-rust/30 transition-all">
										<p class="font-jetbrains text-[10px] font-bold text-text-dim leading-relaxed uppercase italic">
											&gt;&gt; "{version.comment}"
										</p>
									</div>
								{/if}

								<!-- Tactical Actions -->
								<div class="flex gap-3 pt-4 border-t border-stone-800/50">
									{#if !version.is_active}
										<button
											onclick={() => requestActivate(version.id)}
											class="flex-1 px-6 py-3 bg-stone-900 hover:bg-rust border border-stone-800 hover:border-rust-light text-text-dim hover:text-white font-heading font-black text-[10px] uppercase tracking-widest transition-all shadow-lg active:translate-y-px"
										>
											Execute_Activate
										</button>
										<button
											onclick={() => requestDelete(version.id)}
											class="p-3 bg-red-950/20 border border-red-900/30 text-red-600 hover:bg-danger hover:text-white transition-all shadow-lg active:scale-95"
											title="Delete Version"
										>
											<Trash2 class="w-4 h-4" />
										</button>
									{:else}
										<div
											class="flex-1 py-3 bg-rust/5 border border-rust/20 text-rust font-heading font-black text-[10px] uppercase tracking-widest flex items-center justify-center gap-3 shadow-inner"
										>
											<CheckCircle class="w-4 h-4 animate-pulse" />
											Primary_Kernel_Active
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	{#if activeTab === 'history'}
		<!-- Modern Version History -->
		<div
			class="relative overflow-hidden bg-[var(--header-bg)]/60 border border-stone-800 shadow-2xl industrial-frame"
		>
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
			
			<!-- Section Header -->
			<div
				class="relative px-10 py-8 border-b border-stone-800 bg-[var(--header-bg)] backdrop-blur-xl"
			>
				<div class="flex flex-col md:flex-row md:items-center justify-between gap-8">
					<div class="flex items-center gap-6">
						<div
							class="p-4 bg-rust/10 border border-rust/30 industrial-frame shadow-xl"
						>
							<History class="w-8 h-8 text-rust-light" />
						</div>
						<div>
							<h2 class="text-2xl font-heading font-black text-white uppercase tracking-tighter">VERSION_DEPLOYMENT_ARCHIVE</h2>
							<p class="font-jetbrains text-[10px] text-text-dim font-black uppercase tracking-widest mt-1">
								Audit server-side binary transitions and state activations
							</p>
						</div>
					</div>
					<div
						class="px-5 py-2 font-jetbrains text-[10px] font-black text-text-dim bg-stone-950 border border-stone-800 uppercase tracking-[0.2em] shadow-inner"
					>
						{$serverVersions.length}_TOTAL_RECORDS_MAPPED
					</div>
				</div>
			</div>

			<!-- Version List -->
			<div class="p-8">
				{#if $serverVersions.length === 0}
					<!-- Empty State -->
					<div class="flex flex-col items-center justify-center py-32 text-center">
						<div
							class="w-24 h-24 bg-stone-900/40 border border-dashed border-stone-800 flex items-center justify-center industrial-frame mb-8 opacity-40"
						>
							<Package class="w-12 h-12 text-text-dim" />
						</div>
						<h3 class="font-heading font-black text-xl text-stone-700 uppercase tracking-[0.3em] mb-3">
							Registry_Empty
						</h3>
						<p class="font-jetbrains text-[10px] font-bold text-text-dim uppercase tracking-widest max-w-lg mx-auto">
							Initialize first binary synchronization protocol to populate history buffer.
						</p>
						<button
							onclick={() => (activeTab = 'upload')}
							class="mt-10 px-8 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all active:translate-y-px"
						>
							<Upload class="w-4 h-4 inline mr-2" />
							Initialize_Ingress
						</button>
					</div>
				{:else}
					<!-- Version Cards Grid -->
					<div class="grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-3 3xl:grid-cols-4 gap-8">
						{#each $serverVersions as version}
							<div
								class="group modern-industrial-card glass-panel relative overflow-hidden transition-all duration-500 hover:border-rust/40 shadow-xl !rounded-none"
							>
								<!-- Status Indicator -->
								{#if version.is_active}
									<div class="absolute top-6 right-6 z-20">
										<div
											class="flex items-center gap-2.5 px-3 py-1 bg-rust text-white shadow-lg shadow-rust/20"
										>
											<div class="w-1.5 h-1.5 bg-white rounded-full animate-pulse shadow-lg"></div>
											<span class="font-jetbrains text-[9px] font-black uppercase tracking-widest">ACTIVE</span>
										</div>
									</div>
								{/if}

								<div class="relative z-10 p-8 space-y-8">
									<!-- Version Header -->
									<div class="flex items-start justify-between">
										<div class="p-3 bg-rust/5 border border-rust/20 industrial-frame group-hover:border-rust/40 transition-all shadow-inner mr-4">
											<Package class="w-6 h-6 text-rust group-hover:text-rust-light group-hover:scale-110 transition-all" />
										</div>
										<div class="space-y-2 flex-1 min-w-0">
											<h3 class="text-3xl font-heading font-black text-white uppercase tracking-tighter group-hover:text-rust transition-colors duration-500">
												REV_{version.version || '0.0.0'}
											</h3>
											<p class="font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest truncate max-w-[200px]">
												{version.filename}
											</p>
										</div>
										{#if !version.is_active}
											<div
												class="px-3 py-1 bg-stone-900 border border-stone-800 text-[9px] font-jetbrains font-black text-text-dim uppercase tracking-widest"
											>
												HALTED
											</div>
										{/if}
									</div>

									<!-- Upload Info -->
									<div class="space-y-6">
										<div class="flex items-center gap-3 font-jetbrains text-[10px] font-bold text-text-dim uppercase tracking-widest">
											<Clock class="w-4 h-4 text-stone-700" />
											<span>DEPLOYED: {new Date(version.uploaded_at).toLocaleDateString()}</span>
										</div>
										{#if version.comment}
											<div
												class="p-5 bg-stone-950 border-l-2 border-stone-800 group-hover:border-rust/30 transition-all shadow-inner"
											>
												<p class="font-jetbrains text-[10px] font-bold text-text-dim leading-relaxed uppercase italic">
													&gt;&gt; "{version.comment}"
												</p>
											</div>
										{/if}
									</div>

									<!-- Actions -->
									<div class="flex gap-3 pt-4 border-t border-stone-800/50">
										{#if !version.is_active}
											<button
												onclick={() => requestActivate(version.id)}
												class="flex-1 px-6 py-3 bg-stone-900 hover:bg-rust border border-stone-800 hover:border-rust-light text-text-dim hover:text-white font-heading font-black text-[10px] uppercase tracking-widest transition-all shadow-lg active:translate-y-px"
											>
												Execute_Activate
											</button>
											<button
												onclick={() => requestDelete(version.id)}
												class="p-3 bg-red-950/20 border border-red-900/30 text-red-600 hover:bg-danger hover:text-white transition-all shadow-lg active:scale-95"
												title="Delete Version"
											>
												<Trash2 class="w-4 h-4" />
											</button>
										{:else}
											<div
												class="flex-1 py-3 bg-rust/5 border border-rust/20 text-rust font-heading font-black text-[10px] uppercase tracking-widest flex items-center justify-center gap-3 shadow-inner"
											>
												<CheckCircle class="w-4 h-4 animate-pulse" />
												Primary_Kernel_Active
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<ConfirmDialog
		bind:isOpen={isConfirmOpen}
		title={confirmTitle}
		message={confirmMessage}
		confirmText={confirmButtonText}
		isCritical={confirmIsCritical}
		onConfirm={confirmAction}
	/>

	<ConfirmDialog
		bind:isOpen={isSpawnDialogOpen}
		title="Spawn New Instance"
		message={`Are you sure you want to spawn a new game server instance on Node #${spawnTargetNodeId}?`}
		confirmText="Spawn Server"
		onConfirm={executeSpawn}
	/>

	<ConfirmDialog
		bind:isOpen={isInstanceActionDialogOpen}
		title={instanceActionDialogTitle}
		message={instanceActionDialogMessage}
		confirmText={instanceActionConfirmText}
		onConfirm={executeInstanceAction}
	/>

	{#if selectedNodeId}
		<LogViewer
			nodeId={selectedNodeId}
			isOpen={isLogViewerOpen}
			onClose={() => (isLogViewerOpen = false)}
		/>
	{/if}

	<InstanceManagerModal
		bind:isOpen={isConsoleOpen}
		nodeId={consoleNodeId}
		instanceId={consoleInstanceId}
		onClose={() => (isConsoleOpen = false)}
	/>

	<AddNodeModal bind:isOpen={showAddNodeModal} />
</div>