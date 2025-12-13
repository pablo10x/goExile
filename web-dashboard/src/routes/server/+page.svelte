<script lang="ts">
    import { onMount } from 'svelte';
    import { serverVersions } from '$lib/stores';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import { Upload, History, Package, CheckCircle, XCircle, Clock, Trash2, Play } from 'lucide-svelte';

    let activeTab = 'upload';
    
    let fileInput: HTMLInputElement;
    let comment = '';
    let version = '';
    let uploading = false;
    let uploadStatus = '';
    let uploadError = false;
    let selectedFile: File | null = null;
    let fileAnalysis: {
        isUnity: boolean;
        unityVersion?: string;
        platform?: string;
        size: string;
        fileCount?: number;
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
    let sortBy: 'date' | 'version' = 'date';
    let sortOrder: 'asc' | 'desc' = 'desc';

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

        // Simulate file analysis (in real app, this would be server-side)
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        const isUnity = file.name.toLowerCase().includes('unity') || 
                       file.name.toLowerCase().includes('server') ||
                       file.name.toLowerCase().includes('game');
        
        fileAnalysis = {
            isUnity,
            unityVersion: isUnity ? '2022.3.0f1' : undefined,
            platform: 'Windows',
            size: formatFileSize(file.size),
            fileCount: Math.floor(Math.random() * 1000) + 100
        };
        
        analyzing = false;
        return fileAnalysis;
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
            filtered = filtered.filter(v => 
                v.filename.toLowerCase().includes(searchQuery.toLowerCase()) ||
                v.comment?.toLowerCase().includes(searchQuery.toLowerCase()) ||
                v.version.toLowerCase().includes(searchQuery.toLowerCase())
            );
        }
        
        if (filterStatus !== 'all') {
            filtered = filtered.filter(v => 
                filterStatus === 'active' ? v.is_active : !v.is_active
            );
        }
        
        return filtered.sort((a, b) => {
            const aVal = sortBy === 'date' ? new Date(a.uploaded_at).getTime() : a.version;
            const bVal = sortBy === 'date' ? new Date(b.uploaded_at).getTime() : b.version;
            
            if (sortOrder === 'asc') {
                return aVal > bVal ? 1 : -1;
            } else {
                return aVal < bVal ? 1 : -1;
            }
        });
    }

    onMount(() => {
        loadVersions();
    });

    // Drag and drop handlers
    function handleDragOver(event: DragEvent) {
        event.preventDefault();
        dragOver = true;
    }

    function handleDragLeave(event: DragEvent) {
        event.preventDefault();
        dragOver = false;
    }

    function handleDrop(event: DragEvent) {
        event.preventDefault();
        dragOver = false;

        const files = event.dataTransfer?.files;
        if (files && files.length > 0) {
            const file = files[0];
            if (file.name.endsWith('.zip')) {
                // Create a new FileList-like object
                const dt = new DataTransfer();
                dt.items.add(file);
                fileInput.files = dt.files;
                fileInput.dispatchEvent(new Event('change'));
            } else {
                uploadStatus = 'Only .zip files are allowed.';
                uploadError = true;
            }
        }
    }

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
        uploadStatus = 'Uploading...';
        uploadError = false;

        const formData = new FormData();
        formData.append('file', file);
        formData.append('comment', comment);
        formData.append('version', version);

        try {
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                uploadStatus = 'Upload successful!';
                uploadError = false;
                fileInput.value = '';
                comment = '';
                version = '';
                loadVersions();
                setTimeout(() => {
                    uploadStatus = '';
                    activeTab = 'history';
                }, 1000);
            } else {
                uploadStatus = 'Upload failed. Server responded with ' + response.status;
                uploadError = true;
            }
        } catch (e) {
            uploadStatus = 'Upload failed due to network error.';
            uploadError = true;
        } finally {
            uploading = false;
        }
    }

    function requestActivate(id: number) {
        confirmTitle = 'Activate Version';
        confirmMessage = 'Are you sure you want to activate this version? Spawners will download it on next check.';
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

<!-- Modern Header -->
<div class="relative mb-8">
    <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div class="space-y-3">
            <div class="flex items-center gap-4">
                <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                    <Package class="w-6 h-6 text-white" />
                </div>
                <div>
                    <h1 class="text-3xl font-bold bg-gradient-to-r from-slate-100 via-slate-200 to-slate-100 bg-clip-text text-transparent">Server Files</h1>
                    <p class="text-slate-400 text-sm font-medium">Manage game server versions and deployments</p>
                </div>
            </div>

            <!-- Stats Overview -->
            <div class="flex items-center gap-4">
                <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-800/50 border border-slate-700/50 backdrop-blur-sm">
                    <Package class="w-4 h-4 text-indigo-400" />
                    <span class="text-xs font-semibold text-slate-300">{$serverVersions.length} versions</span>
                </div>
                {#if $serverVersions.some(v => v.is_active)}
                    <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
                        <div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
                        <span class="text-xs font-semibold text-emerald-400">Active version deployed</span>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</div>

<!-- Enhanced Tabs -->
<div class="relative mb-8">
    <div class="flex gap-1 p-1.5 bg-slate-800/40 border border-slate-700/50 rounded-2xl backdrop-blur-sm shadow-lg">
        <button
            onclick={() => activeTab = 'upload'}
            class="group relative flex items-center gap-3 px-6 py-3 rounded-xl transition-all duration-300 overflow-hidden
            {activeTab === 'upload' ?
                'bg-gradient-to-r from-indigo-600/20 to-purple-600/20 text-indigo-300 shadow-lg shadow-indigo-500/10 border border-indigo-500/30 backdrop-blur-sm' :
                'hover:bg-slate-700/50 text-slate-400 hover:text-slate-200 border border-transparent'
            }"
        >
            {#if activeTab === 'upload'}
                <div class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-indigo-400 to-purple-400 rounded-r-full shadow-lg"></div>
                <div class="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 rounded-xl"></div>
            {/if}
            <Upload class="w-5 h-5 transition-all duration-300 group-hover:scale-110 {activeTab === 'upload' ? 'text-indigo-400' : ''}" />
            <span class="font-semibold tracking-wide relative z-10">Upload Version</span>
            {#if activeTab === 'upload'}
                <div class="ml-auto w-2 h-2 bg-indigo-400 rounded-full animate-pulse relative z-10"></div>
            {/if}
        </button>

        <button
            onclick={() => activeTab = 'history'}
            class="group relative flex items-center gap-3 px-6 py-3 rounded-xl transition-all duration-300 overflow-hidden
            {activeTab === 'history' ?
                'bg-gradient-to-r from-indigo-600/20 to-purple-600/20 text-indigo-300 shadow-lg shadow-indigo-500/10 border border-indigo-500/30 backdrop-blur-sm' :
                'hover:bg-slate-700/50 text-slate-400 hover:text-slate-200 border border-transparent'
            }"
        >
            {#if activeTab === 'history'}
                <div class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-indigo-400 to-purple-400 rounded-r-full shadow-lg"></div>
                <div class="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 rounded-xl"></div>
            {/if}
            <History class="w-5 h-5 transition-all duration-300 group-hover:scale-110 {activeTab === 'history' ? 'text-indigo-400' : ''}" />
            <span class="font-semibold tracking-wide relative z-10">Version History</span>
            {#if activeTab === 'history'}
                <div class="ml-auto w-2 h-2 bg-indigo-400 rounded-full animate-pulse relative z-10"></div>
            {/if}
        </button>
    </div>
</div>

    {#if activeTab === 'upload'}
        <!-- Modern Upload Section -->
        <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-800/40 via-slate-900/40 to-slate-800/40 border border-slate-700/50 backdrop-blur-sm shadow-2xl">
            <!-- Section Header -->
            <div class="relative px-8 py-6 border-b border-slate-700/50 bg-gradient-to-r from-slate-800/60 to-slate-900/60 backdrop-blur-sm">
                <div class="flex items-center gap-4">
                    <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                        <Upload class="w-5 h-5 text-white" />
                    </div>
                    <div>
                        <h2 class="text-xl font-bold text-slate-100">Upload New Version</h2>
                        <p class="text-sm text-slate-400">Deploy a new game server package to the registry</p>
                    </div>
                </div>
            </div>

            <!-- Upload Form -->
            <div class="p-8">
                <div class="max-w-2xl mx-auto space-y-8">
                    <!-- Info Banner -->
                    <div class="bg-gradient-to-r from-indigo-500/10 to-purple-500/10 border border-indigo-500/20 rounded-2xl p-6">
                        <div class="flex items-start gap-4">
                            <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg flex-shrink-0">
                                <Package class="w-5 h-5 text-white" />
                            </div>
                            <div class="space-y-2">
                                <h3 class="text-lg font-semibold text-slate-100">Package Requirements</h3>
                                <p class="text-slate-300 text-sm leading-relaxed">
                                    Upload a <code class="bg-slate-800/50 px-2 py-1 rounded text-xs font-mono">game_server.zip</code> package containing your game server files.
                                    Include version information and change notes for better tracking.
                                </p>
                            </div>
                        </div>
                    </div>

                    <!-- Enhanced File Upload Area -->
                    <div class="space-y-4">
                        <label class="block text-sm font-bold text-slate-200 uppercase tracking-wider">Server Package</label>

                        <div
                            class="relative group"
                            ondragover={handleDragOver}
                            ondragleave={handleDragLeave}
                            ondrop={handleDrop}
                        >
                            <div class={`relative overflow-hidden rounded-2xl border-2 border-dashed transition-all duration-300 p-8 text-center ${
                                dragOver
                                    ? 'border-indigo-400 bg-gradient-to-br from-indigo-500/10 to-purple-500/10 scale-[1.02] shadow-2xl shadow-indigo-500/20'
                                    : 'border-slate-600/50 bg-gradient-to-br from-slate-800/50 to-slate-900/50 hover:border-slate-500/50 hover:bg-gradient-to-br hover:from-slate-800/60 hover:to-slate-900/60'
                            }`}>

                                <!-- Background Effects -->
                                <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

                                <div class="relative z-10 space-y-4">
                                    <div class={`w-16 h-16 mx-auto rounded-2xl flex items-center justify-center shadow-lg transition-all duration-300 ${
                                        dragOver
                                            ? 'bg-gradient-to-br from-indigo-500 to-purple-600 shadow-indigo-500/30 scale-110'
                                            : 'bg-gradient-to-br from-slate-700 to-slate-800 shadow-slate-900/20 group-hover:scale-105'
                                    }`}>
                                        <Upload class={`w-8 h-8 transition-colors duration-300 ${
                                            dragOver ? 'text-white' : 'text-slate-400 group-hover:text-slate-300'
                                        }`} />
                                    </div>

                                    <div class="space-y-2">
                                        <h3 class="text-lg font-semibold text-slate-200">
                                            {dragOver ? 'Drop your package here' : 'Upload Game Server Package'}
                                        </h3>
                                        <p class="text-slate-400 text-sm">
                                            {dragOver ? 'Release to upload the file' : 'Drag & drop your .zip file here, or click to browse'}
                                        </p>
                                    </div>

                                    <div class="flex justify-center">
                                        <label class="group/btn relative">
                                            <input
                                                type="file"
                                                class="sr-only"
                                                accept=".zip"
                                                bind:this={fileInput}
                                                onchange={() => { uploadStatus = ''; }}
                                            />
                                            <div class="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300 cursor-pointer group-hover/btn:scale-105 group-active/btn:scale-95">
                                                Choose File
                                            </div>
                                        </label>
                                    </div>

                                    <p class="text-xs text-slate-500">
                                        Supports ZIP files up to 100MB
                                    </p>
                                </div>
                            </div>
                        </div>

                        <!-- File Selection Status -->
                        {#if fileInput && fileInput.files && fileInput.files.length > 0}
                            <div class="flex items-center gap-3 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl">
                                <div class="w-8 h-8 bg-emerald-500 rounded-lg flex items-center justify-center">
                                    <CheckCircle class="w-5 h-5 text-white" />
                                </div>
                                <div class="flex-1">
                                    <p class="text-sm font-semibold text-emerald-300">File Selected</p>
                                    <p class="text-xs text-emerald-400 font-mono">{fileInput.files[0].name}</p>
                                </div>
                                <button
                                    onclick={() => { fileInput.value = ''; uploadStatus = ''; }}
                                    class="p-2 text-emerald-400 hover:text-emerald-300 transition-colors"
                                >
                                    <XCircle class="w-4 h-4" />
                                </button>
                            </div>
                        {/if}
                    </div>

                    <!-- Version and Comment Inputs -->
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <!-- Version Input -->
                        <div class="space-y-3">
                            <label for="version" class="block text-sm font-bold text-slate-200 uppercase tracking-wider">
                                Version Tag
                            </label>
                            <input
                                type="text"
                                id="version"
                                bind:value={version}
                                pattern="^\d+\.\d+\.\d+$"
                                title="Version must be in semantic versioning format (e.g. 1.2.3)"
                                class="w-full px-4 py-3 bg-slate-900/50 border border-slate-700/50 rounded-xl text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 focus:ring-2 focus:ring-indigo-500/20 transition-all duration-300 backdrop-blur-sm"
                                placeholder="e.g. 1.2.0"
                                oninput={(e) => {
                                    const target = e.target as HTMLInputElement;
                                    let value = target.value;
                                    
                                    // Only allow digits and periods
                                    value = value.replace(/[^\d.]/g, '');
                                    
                                    // Prevent consecutive periods
                                    value = value.replace(/\.+/g, '.');
                                    
                                    // Prevent starting or ending with period
                                    value = value.replace(/^\.+|\.+$/g, '');
                                    
                                    // Limit to semantic versioning format (max 2 periods)
                                    const parts = value.split('.');
                                    if (parts.length > 3) {
                                        value = parts.slice(0, 3).join('.');
                                    }
                                    
                                    target.value = value;
                                    version = value;
                                }}
                            />
                        </div>

                        <!-- Comment Input -->
                        <div class="space-y-3 md:col-span-2">
                            <label for="comment" class="block text-sm font-bold text-slate-200 uppercase tracking-wider">
                                Release Notes
                            </label>
                            <textarea
                                id="comment"
                                bind:value={comment}
                                rows="4"
                                class="w-full px-4 py-3 bg-slate-900/50 border border-slate-700/50 rounded-xl text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50 focus:ring-2 focus:ring-indigo-500/20 transition-all duration-300 backdrop-blur-sm resize-none"
                                placeholder="Describe the changes, fixes, and new features in this version..."
                            ></textarea>
                        </div>
                    </div>

                    <!-- Upload Button and Status -->
                    <div class="space-y-4">
                        <button
                            onclick={handleUpload}
                            disabled={uploading || !fileInput?.files?.length}
                            class="w-full group relative overflow-hidden bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 disabled:from-slate-700 disabled:to-slate-800 text-white font-bold py-4 rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
                        >
                            <div class="absolute inset-0 bg-gradient-to-r from-indigo-400 to-purple-400 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                            {#if uploading}
                                <div class="relative z-10 w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                                <span class="relative z-10 font-semibold">Deploying Version...</span>
                            {:else}
                                <Upload class="relative z-10 w-5 h-5" />
                                <span class="relative z-10 font-semibold">Deploy Version</span>
                            {/if}
                        </button>

                        <!-- Upload Status -->
                        {#if uploadStatus}
                            <div class={`p-4 rounded-xl border flex items-center gap-3 transition-all duration-300 ${
                                uploadError
                                    ? 'bg-red-500/10 border-red-500/20 text-red-300'
                                    : 'bg-emerald-500/10 border-emerald-500/20 text-emerald-300'
                            }`}>
                                <div class={`w-8 h-8 rounded-lg flex items-center justify-center ${
                                    uploadError ? 'bg-red-500/20' : 'bg-emerald-500/20'
                                }`}>
                                    {#if uploadError}
                                        <XCircle class="w-5 h-5" />
                                    {:else}
                                        <CheckCircle class="w-5 h-5" />
                                    {/if}
                                </div>
                                <div class="flex-1">
                                    <p class="font-semibold">{uploadError ? 'Upload Failed' : 'Upload Successful'}</p>
                                    <p class="text-sm opacity-90">{uploadStatus}</p>
                                </div>
                            </div>
                        {/if}
                    </div>
                </div>
            </div>
        </div>
    {/if}

    {#if activeTab === 'history'}
        <!-- Modern Version History -->
        <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-800/40 via-slate-900/40 to-slate-800/40 border border-slate-700/50 backdrop-blur-sm shadow-2xl">
            <!-- Section Header -->
            <div class="relative px-8 py-6 border-b border-slate-700/50 bg-gradient-to-r from-slate-800/60 to-slate-900/60 backdrop-blur-sm">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-4">
                        <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/25">
                            <History class="w-5 h-5 text-white" />
                        </div>
                        <div>
                            <h2 class="text-xl font-bold text-slate-100">Version History</h2>
                            <p class="text-sm text-slate-400">Manage deployed server versions and deployments</p>
                        </div>
                    </div>
                    <div class="text-xs text-slate-500 bg-slate-800/50 px-3 py-1 rounded-md border border-slate-700/50">
                        {$serverVersions.length} total versions
                    </div>
                </div>
            </div>

            <!-- Version List -->
            <div class="p-6">
                {#if $serverVersions.length === 0}
                    <!-- Empty State -->
                    <div class="flex flex-col items-center justify-center py-16 text-center">
                        <div class="w-20 h-20 bg-slate-800/30 rounded-full flex items-center justify-center border border-slate-700/30 mb-6">
                            <Package class="w-10 h-10 opacity-50 text-slate-500" />
                        </div>
                        <h3 class="text-xl font-semibold text-slate-400 mb-2">No Versions Yet</h3>
                        <p class="text-slate-600 max-w-md">Upload your first game server package to get started with version management.</p>
                        <button
                            onclick={() => activeTab = 'upload'}
                            class="mt-6 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 transition-all duration-300 flex items-center gap-2"
                        >
                            <Upload class="w-5 h-5" />
                            Upload First Version
                        </button>
                    </div>
                {:else}
                    <!-- Version Cards Grid -->
                    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
                        {#each $serverVersions as version}
                            <div class="group relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800/60 to-slate-900/60 border border-slate-700/50 backdrop-blur-sm transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl hover:shadow-indigo-500/10">
                                <!-- Status Indicator -->
                                {#if version.is_active}
                                    <div class="absolute top-4 right-4 z-20">
                                        <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 backdrop-blur-sm">
                                            <div class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></div>
                                            <span class="text-xs font-semibold text-emerald-400">Active</span>
                                        </div>
                                    </div>
                                {/if}

                                <!-- Background Effects -->
                                <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

                                <div class="relative z-10 p-6">
                                    <!-- Version Header -->
                                    <div class="flex items-start justify-between mb-4">
                                        <div class="space-y-1">
                                            <h3 class="text-lg font-bold text-slate-100">
                                                v{version.version || 'Unknown'}
                                            </h3>
                                            <p class="text-xs text-slate-400 font-mono">
                                                {version.filename}
                                            </p>
                                        </div>
                                        {#if !version.is_active}
                                            <div class="px-2 py-1 rounded-md bg-slate-700/50 text-xs text-slate-400 border border-slate-600/30">
                                                Inactive
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Upload Info -->
                                    <div class="space-y-3 mb-6">
                                        <div class="flex items-center gap-2 text-sm text-slate-400">
                                            <Clock class="w-4 h-4" />
                                            <span>Uploaded {new Date(version.uploaded_at).toLocaleDateString()}</span>
                                        </div>
                                        {#if version.comment}
                                            <div class="p-3 bg-slate-800/50 rounded-lg border border-slate-700/30">
                                                <p class="text-sm text-slate-300 leading-relaxed">{version.comment}</p>
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Actions -->
                                    <div class="flex gap-2">
                                        {#if !version.is_active}
                                            <button
                                                onclick={() => requestActivate(version.id)}
                                                class="flex-1 group/btn relative overflow-hidden bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold py-2.5 px-4 rounded-lg shadow-lg shadow-indigo-500/25 transition-all duration-300 flex items-center justify-center gap-2"
                                            >
                                                <div class="absolute inset-0 bg-gradient-to-r from-indigo-400 to-purple-400 translate-x-[-100%] group-hover/btn:translate-x-[100%] transition-transform duration-700"></div>
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
                                            <div class="flex-1 py-2.5 px-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center gap-2">
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