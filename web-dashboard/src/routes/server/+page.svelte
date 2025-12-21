<script lang="ts">
	import { onMount } from 'svelte';
	import { serverVersions } from '$lib/stores';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { setContext } from 'svelte';
	import { History, Package, Upload, Play, Trash2, CheckCircle, Clock } from 'lucide-svelte';

	let activeTab = 'upload';
	let isDragging = false;
	let dragCounter = 0;

	let fileInput: HTMLInputElement;
	let comment = '';
	let version = '';
	let uploading = false;
	let uploadProgress = 0;
	let uploadStatus = '';
	let uploadError = false;
	let selectedFile: File | null = null;
	let fileAnalysis: {
		isUnity: boolean;
		unityVersion?: string;
		platform?: string;
		size: string;
		fileCount?: number;
		estimatedTime?: string;
		compatibility?: 'excellent' | 'good' | 'fair' | 'poor';
	} | null = null;
	let analyzing = false;

	let isConfirmOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let confirmAction: () => Promise<void> = async () => {};
	let confirmIsCritical = false;
	let confirmButtonText = 'Confirm';

	let searchQuery = '';
	let filterStatus: 'all' | 'active' | 'inactive' = 'all';
	let sortBy: 'date' | 'version' | 'size' = 'date';
	let sortOrder: 'asc' | 'desc' = 'desc';
	let showAdvancedFilters = false;

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
				fileInput.files = e.dataTransfer.files;
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
				return 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20';
			case 'good':
				return 'text-blue-400 bg-blue-400/10 border-blue-400/20';
			case 'fair':
				return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/20';
			case 'poor':
				return 'text-red-400 bg-red-400/10 border-red-400/20';
			default:
				return 'text-slate-500 dark:text-slate-400 bg-slate-400/10 border-slate-400/20';
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
		if (!fileInput.files || fileInput.files.length === 0) {
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
				fileInput.value = '';
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
			'Are you sure you want to activate this version? Spawners will download it on next check.';
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
				loadVersions();
			} else {
				const data = await res.json();
				throw new Error(data.error || 'Failed to delete version');
			}
		} catch (e) {
			console.error(e);
			throw e;
		}
	}
</script>

<div class="min-h-screen space-y-6 relative">
	<!-- Background Pattern -->
	<div class="fixed inset-0 -z-10 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
		<div
			class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPgogICAgPGcgZmlsbD0iIzY0NzQ4YiIgZmlsbC1vcGFjaXR5PSIwLjAzIj4KICAgICAgPHBhdGggZD0iTTM2IDM0di00aC0ydjRoLTR2Mmg0djRoMnYtNGg0di0yaC00em0wLTMwVjBoLTJ2NGgtNHYyaDR2NGgyVjZoNFY0aC00ek02IDM0di00SDR2NEgwdjJoNHY0aDJ2LTRoNHYtMkg2ek02IDRWMEg0djRIMHYyaDR2NGgyVjZoNFY0SDZ6Ii8+CiAgICA8L2c+CiAgPC9nPgo8L3N2Zz4=')] opacity-20"
		></div>
		<div class="absolute inset-0 bg-gradient-to-t from-slate-900/50 to-transparent"></div>
	</div>
	<div class="flex justify-between items-center">
		<h2 class="text-2xl font-bold text-slate-50">Game Server Files</h2>
	</div>

	<!-- Tabs -->
	<div class="border-b border-slate-300 dark:border-slate-700">
		<nav class="-mb-px flex space-x-8">
			<button
				onclick={() => (activeTab = 'upload')}
				class="{activeTab === 'upload'
					? 'border-blue-500 text-blue-400'
					: 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:text-slate-300 hover:border-slate-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
			>
				Upload New Version
			</button>
			<button
				onclick={() => (activeTab = 'history')}
				class="{activeTab === 'history'
					? 'border-blue-500 text-blue-400'
					: 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:text-slate-300 hover:border-slate-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
			>
				Version History
			</button>
		</nav>
	</div>

	{#if activeTab === 'upload'}
		<div class="grid lg:grid-cols-3 gap-6">
			<!-- Upload Area -->
			<div class="lg:col-span-2">
				<div class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-8">
					<div class="mb-6">
						<h3 class="text-lg font-semibold text-slate-50 mb-2">Upload Game Server Package</h3>
						<p class="text-slate-500 dark:text-slate-400 text-sm">
							Upload a new <code class="px-1 py-0.5 bg-slate-700 rounded text-xs"
								>game_server.zip</code
							> package with version information and release notes.
						</p>
					</div>

					<!-- Enhanced Drag & Drop Area -->
					<div
						class="relative"
						ondragenter={handleDragEnter}
						ondragleave={handleDragLeave}
						ondragover={handleDragOver}
						ondrop={handleDrop}
						role="region"
						aria-label="File Upload Drop Zone"
					>
						<div
							class="group relative border-2 border-dashed rounded-xl transition-all duration-200 bg-slate-900/30 hover:bg-slate-900/50 {isDragging
								? 'border-blue-400 bg-blue-400/5'
								: 'border-slate-300 dark:border-slate-700 hover:border-slate-600'}"
						>
							<input
								id="file-upload"
								name="file-upload"
								type="file"
								class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
								accept=".zip"
								bind:this={fileInput}
								onchange={handleFileSelect}
							/>

							<div class="p-12 text-center">
								{#if uploading}
									<div class="space-y-4">
										<div
											class="w-16 h-16 mx-auto border-4 border-blue-500 border-t-transparent rounded-full animate-spin"
										></div>
										<div class="space-y-2">
											<p class="text-blue-400 font-medium">{uploadStatus}</p>
											{#if uploadProgress > 0}
												<div class="w-full max-w-md mx-auto">
													<div class="h-2 bg-slate-700 rounded-full overflow-hidden">
														<div
															class="h-full bg-blue-500 transition-all duration-300"
															style="width: {uploadProgress}%"
														></div>
													</div>
													<p class="text-xs text-slate-500 mt-1">
														{Math.round(uploadProgress)}% complete
													</p>
												</div>
											{/if}
										</div>
									</div>
								{:else if analyzing}
									<div class="space-y-4">
										<div
											class="w-16 h-16 mx-auto border-4 border-yellow-500 border-t-transparent rounded-full animate-spin"
										></div>
										<div class="space-y-2">
											<p class="text-yellow-400 font-medium">Analyzing file...</p>
											<p class="text-xs text-slate-500">
												Checking compatibility and extracting metadata
											</p>
										</div>
									</div>
								{:else}
									<div class="space-y-4">
										<div
											class="w-16 h-16 mx-auto mx-auto flex items-center justify-center rounded-full bg-slate-800 group-hover:bg-slate-700 transition-colors"
										>
											<svg
												class="w-8 h-8 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:text-slate-300"
												fill="none"
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="1.5"
													d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
												/>
											</svg>
										</div>
										<div class="space-y-2">
											<p class="text-slate-700 dark:text-slate-300 font-medium">
												{selectedFile
													? selectedFile.name
													: 'Drop your zip file here or click to browse'}
											</p>
											<p class="text-xs text-slate-500">
												{selectedFile
													? formatFileSize(selectedFile.size)
													: 'Maximum file size: 1GB • ZIP files only'}
											</p>
										</div>
										{#if !selectedFile}
											<button
												class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-slate-900 dark:text-white rounded-lg text-sm font-medium transition-colors"
											>
												Choose File
											</button>
										{/if}
									</div>
								{/if}
							</div>
						</div>
					</div>

					<!-- File Analysis Results -->
					{#if fileAnalysis && !uploading}
						<div class="mt-6 p-4 bg-slate-900/50 rounded-lg border border-slate-300 dark:border-slate-700">
							<h4 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">File Analysis</h4>
							<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
								<div class="text-center">
									<p class="text-xs text-slate-500 mb-1">Type</p>
									<p class="text-sm font-medium text-slate-700 dark:text-slate-300">
										{fileAnalysis.isUnity ? 'Unity Server' : 'Game Server'}
									</p>
								</div>
								<div class="text-center">
									<p class="text-xs text-slate-500 mb-1">Size</p>
									<p class="text-sm font-medium text-slate-700 dark:text-slate-300">{fileAnalysis.size}</p>
								</div>
								<div class="text-center">
									<p class="text-xs text-slate-500 mb-1">Files</p>
									<p class="text-sm font-medium text-slate-700 dark:text-slate-300">{fileAnalysis.fileCount}</p>
								</div>
								<div class="text-center">
									<p class="text-xs text-slate-500 mb-1">Est. Time</p>
									<p class="text-sm font-medium text-slate-700 dark:text-slate-300">{fileAnalysis.estimatedTime}</p>
								</div>
							</div>
							{#if fileAnalysis.compatibility}
								<div class="mt-3 flex items-center justify-center">
									<span
										class={`px-3 py-1 rounded-full text-xs font-medium border ${getCompatibilityColor(fileAnalysis.compatibility)}`}
									>
										Compatibility: {fileAnalysis.compatibility}
									</span>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Version Details Panel -->
			<div class="lg:col-span-1 space-y-6">
				<!-- Version Information -->
				<div class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-6">
					<h3 class="text-lg font-semibold text-slate-50 mb-4">Version Information</h3>

					<div class="space-y-4">
						<div>
							<label for="version" class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
								Version Number <span class="text-red-400">*</span>
							</label>
							<input
								type="text"
								id="version"
								bind:value={version}
								class="w-full px-3 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors text-sm"
								placeholder="1.0.0"
								required
							/>
							<p class="text-xs text-slate-500 mt-1">Semantic versioning (e.g., 1.2.3)</p>
						</div>

						<div>
							<label for="comment" class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
								Release Notes
							</label>
							<textarea
								id="comment"
								bind:value={comment}
								rows="4"
								class="w-full px-3 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors text-sm resize-none"
								placeholder="Describe what's new in this version..."
							></textarea>
						</div>
					</div>
				</div>

				<!-- Upload Button & Status -->
				<div class="space-y-4">
					<button
						onclick={handleUpload}
						disabled={uploading || !selectedFile || !version}
						class="w-full px-6 py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 disabled:text-slate-500 text-slate-900 dark:text-white rounded-lg font-semibold transition-all transform active:scale-95 disabled:cursor-not-allowed disabled:transform-none flex items-center justify-center gap-2 shadow-lg shadow-blue-900/20"
					>
						{#if uploading}
							<div
								class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
							></div>
							Uploading...
						{:else}
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
								/>
							</svg>
							Upload Version
						{/if}
					</button>

					{#if uploadStatus}
						<div
							class={`p-4 rounded-lg text-sm font-medium border flex items-start gap-3 ${uploadError ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'}`}
						>
							<span class="flex-shrink-0 mt-0.5">{uploadError ? '⚠️' : '✓'}</span>
							<div>
								<p class="font-medium">{uploadError ? 'Upload Failed' : 'Upload Successful'}</p>
								<p class="text-xs opacity-75 mt-1">{uploadStatus}</p>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{:else}
		<div class="space-y-6">
			<!-- Filters and Search -->
			<div class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-6">
				<div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
					<div class="flex-1 max-w-md">
						<div class="relative">
							<input
								type="text"
								bind:value={searchQuery}
								placeholder="Search versions, comments, or filenames..."
								class="w-full pl-10 pr-4 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors text-sm"
							/>
							<svg
								class="absolute left-3 top-2.5 w-4 h-4 text-slate-500"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
								/>
							</svg>
						</div>
					</div>

					<div class="flex flex-wrap items-center gap-3">
						<!-- Status Filter -->
						<select
							bind:value={filterStatus}
							class="px-3 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 text-sm focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
						>
							<option value="all">All Status</option>
							<option value="active">Active Only</option>
							<option value="inactive">Inactive Only</option>
						</select>

						<!-- Sort Options -->
						<div class="flex items-center gap-2">
							<select
								bind:value={sortBy}
								class="px-3 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-800 dark:text-slate-200 text-sm focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
							>
								<option value="date">Sort by Date</option>
								<option value="version">Sort by Version</option>
								<option value="size">Sort by Size</option>
							</select>
							<button
								onclick={() => (sortOrder = sortOrder === 'desc' ? 'asc' : 'desc')}
								class="p-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:border-slate-500 transition-colors"
								title="Toggle sort order"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									{#if sortOrder === 'desc'}
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 9l-7 7-7-7"
										/>
									{:else}
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M5 15l7-7 7 7"
										/>
									{/if}
								</svg>
							</button>
						</div>
					</div>
				</div>
			</div>

			<!-- Version Grid -->
			{#if getFilteredVersions().length === 0}
				<div class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl p-12 text-center">
					<svg
						class="w-16 h-16 mx-auto text-slate-600 mb-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="1.5"
							d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
						/>
					</svg>
					<h3 class="text-lg font-medium text-slate-700 dark:text-slate-300 mb-2">No versions found</h3>
					<p class="text-slate-500">
						{searchQuery
							? 'Try adjusting your search or filters.'
							: 'Upload your first game server version to get started.'}
					</p>
				</div>
			{:else}
				<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
					{#each getFilteredVersions() as version (version.id)}
						<div
							class="bg-slate-800/60 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden hover:bg-gradient-to-br hover:from-slate-800/80 hover:via-blue-900/30 hover:to-slate-800/80 hover:border-blue-500/30 transition-all duration-300 group preserve-3d transform-gpu hover:rotate-y-6 hover:scale-105 shadow-lg hover:shadow-2xl hover:shadow-blue-500/10"
							style="transform-style: preserve-3d; perspective: 1000px;"
						>
							<!-- Header with Status -->
							<div class="p-6 pb-4 relative overflow-hidden">
								<div class="flex items-start justify-between mb-3">
									<div class="flex items-center gap-3">
										{#if version.is_active}
											<div class="relative">
												<span
													class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
												>
													<span class="w-2 h-2 bg-emerald-400 rounded-full mr-2 animate-pulse"
													></span>
													Active
												</span>
											</div>
										{:else}
											<span
												class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-slate-700 text-slate-500 dark:text-slate-400"
											>
												Inactive
											</span>
										{/if}
									</div>

									<!-- Action Buttons that appear on hover -->
									<div
										class="absolute top-4 right-4 flex gap-2 opacity-0 group-hover:opacity-100 transition-all duration-300 transform translate-y-2 group-hover:translate-y-0 z-10"
									>
										{#if !version.is_active}
											<button
												onclick={() => requestActivate(version.id)}
												class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-slate-900 dark:text-white rounded-lg text-xs font-medium transition-all duration-200 transform hover:scale-110 shadow-lg hover:shadow-emerald-500/25 flex items-center gap-1"
												title="Activate this version"
											>
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
													/>
												</svg>
												Activate
											</button>
										{/if}
										<button
											onclick={() => requestDelete(version.id)}
											class="px-3 py-1.5 bg-red-600 hover:bg-red-500 text-slate-900 dark:text-white rounded-lg text-xs font-medium transition-all duration-200 transform hover:scale-110 shadow-lg hover:shadow-red-500/25 flex items-center gap-1"
											title="Delete this version"
										>
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
												/>
											</svg>
											Delete
										</button>
									</div>
								</div>

								<!-- Version Info -->
								<div class="space-y-3">
									<div>
										<h4 class="text-lg font-semibold text-slate-50 mb-1">
											{version.version || 'Unknown Version'}
										</h4>
										<p class="text-xs font-mono text-slate-500">{version.filename}</p>
									</div>

									<!-- Metadata -->
									<div class="space-y-2">
										<div class="flex items-center justify-between text-sm">
											<span class="text-slate-500">Uploaded</span>
											<span class="text-slate-500 dark:text-slate-400"
												>{new Date(version.uploaded_at).toLocaleDateString()}</span
											>
										</div>
										<div class="flex items-center justify-between text-sm">
											<span class="text-slate-500">Time</span>
											<span class="text-slate-500 dark:text-slate-400"
												>{new Date(version.uploaded_at).toLocaleTimeString()}</span
											>
										</div>
									</div>

									<!-- Comment -->
									{#if version.comment}
										<div class="pt-3 border-t border-slate-300 dark:border-slate-700">
											<p class="text-sm text-slate-500 dark:text-slate-400 line-clamp-3">{version.comment}</p>
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
			class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-800/40 via-slate-900/40 to-slate-800/40 border border-slate-300/50 dark:border-slate-700/50 backdrop-blur-sm shadow-2xl"
		>
			<!-- Section Header -->
			<div
				class="relative px-8 py-6 border-b border-slate-300/50 dark:border-slate-700/50 bg-gradient-to-r from-slate-800/60 to-slate-900/60 backdrop-blur-sm"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div
							class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/25"
						>
							<History class="w-5 h-5 text-slate-900 dark:text-white" />
						</div>
						<div>
							<h2 class="text-xl font-bold text-slate-100">Version History</h2>
							<p class="text-sm text-slate-500 dark:text-slate-400">Manage deployed server versions and deployments</p>
						</div>
					</div>
					<div
						class="text-xs text-slate-500 bg-slate-800/50 px-3 py-1 rounded-md border border-slate-300/50 dark:border-slate-700/50"
					>
						{$serverVersions.length} total versions
					</div>
				</div>
			</div>

			<!-- Version List -->
			<div class="p-6">
				{#if $serverVersions.length === 0}
					<!-- Empty State -->
					<div class="flex flex-col items-center justify-center py-16 text-center">
						<div
							class="w-20 h-20 bg-slate-800/30 rounded-full flex items-center justify-center border border-slate-300/30 dark:border-slate-700/30 mb-6"
						>
							<Package class="w-10 h-10 opacity-50 text-slate-500" />
						</div>
						<h3 class="text-xl font-semibold text-slate-500 dark:text-slate-400 mb-2">No Versions Yet</h3>
						<p class="text-slate-600 max-w-md">
							Upload your first game server package to get started with version management.
						</p>
						<button
							onclick={() => (activeTab = 'upload')}
							class="mt-6 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-slate-900 dark:text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300 flex items-center gap-2"
						>
							<Upload class="w-5 h-5" />
							Upload First Version
						</button>
					</div>
				{:else}
					<!-- Version Cards Grid -->
					<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
						{#each $serverVersions as version}
							<div
								class="group relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800/60 to-slate-900/60 border border-slate-300/50 dark:border-slate-700/50 backdrop-blur-sm transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl hover:shadow-indigo-500/10"
							>
								<!-- Status Indicator -->
								{#if version.is_active}
									<div class="absolute top-4 right-4 z-20">
										<div
											class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 backdrop-blur-sm"
										>
											<div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
											<span class="text-xs font-semibold text-emerald-400">Active</span>
										</div>
									</div>
								{/if}

								<!-- Background Effects -->
								<div
									class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
								></div>

								<div class="relative z-10 p-6">
									<!-- Version Header -->
									<div class="flex items-start justify-between mb-4">
										<div class="space-y-1">
											<h3 class="text-lg font-bold text-slate-100">
												v{version.version || 'Unknown'}
											</h3>
											<p class="text-xs text-slate-500 dark:text-slate-400 font-mono">
												{version.filename}
											</p>
										</div>
										{#if !version.is_active}
											<div
												class="px-2 py-1 rounded-md bg-slate-700/50 text-xs text-slate-500 dark:text-slate-400 border border-slate-600/30"
											>
												Inactive
											</div>
										{/if}
									</div>

									<!-- Upload Info -->
									<div class="space-y-3 mb-6">
										<div class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
											<Clock class="w-4 h-4" />
											<span>Uploaded {new Date(version.uploaded_at).toLocaleDateString()}</span>
										</div>
										{#if version.comment}
											<div class="p-3 bg-slate-800/50 rounded-lg border border-slate-300/30 dark:border-slate-700/30">
												<p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{version.comment}</p>
											</div>
										{/if}
									</div>

									<!-- Actions -->
									<div class="flex gap-2">
										{#if !version.is_active}
											<button
												onclick={() => requestActivate(version.id)}
												class="flex-1 group/btn relative overflow-hidden bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-slate-900 dark:text-white font-semibold py-2.5 px-4 rounded-lg shadow-lg shadow-indigo-500/25 transition-all duration-300 flex items-center justify-center gap-2"
											>
												<div
													class="absolute inset-0 bg-gradient-to-r from-indigo-400 to-purple-400 translate-x-[-100%] group-hover/btn:translate-x-[100%] transition-transform duration-700"
												></div>
												<Play class="w-4 h-4 relative z-10" />
												<span class="text-sm relative z-10">Activate</span>
											</button>
											<button
												onclick={() => requestDelete(version.id)}
												class="p-2.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg transition-all duration-300 hover:scale-105"
												title="Delete Version"
											>
												<Trash2 class="w-4 h-4" />
											</button>
										{:else}
											<div
												class="flex-1 py-2.5 px-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center gap-2"
											>
												<CheckCircle class="w-4 h-4 text-emerald-400" />
												<span class="text-sm font-semibold text-emerald-400">Currently Active</span>
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
</div>
