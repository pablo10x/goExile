<script lang="ts">
    import { onMount } from 'svelte';
    import { serverVersions } from '$lib/stores';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

    let activeTab = 'upload'; // 'upload' | 'history'
    
    // Upload State
    let fileInput: HTMLInputElement;
    let comment = '';
    let version = '';
    let uploading = false;
    let uploadStatus = '';
    let uploadError = false;

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmTitle = '';
    let confirmMessage = '';
    let confirmAction: () => Promise<void> = async () => {};
    let confirmIsCritical = false;
    let confirmButtonText = 'Confirm';

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

<div class="space-y-6">
    <div class="flex justify-between items-center">
        <h2 class="text-2xl font-bold text-slate-50">Game Server Files</h2>
    </div>

    <!-- Tabs -->
    <div class="border-b border-slate-700">
        <nav class="-mb-px flex space-x-8">
            <button 
                onclick={() => activeTab = 'upload'}
                class="{activeTab === 'upload' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
            >
                Upload New Version
            </button>
            <button 
                onclick={() => activeTab = 'history'}
                class="{activeTab === 'history' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'} whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
            >
                Version History
            </button>
        </nav>
    </div>

    {#if activeTab === 'upload'}
        <div class="card p-8 bg-slate-800/50 border border-slate-700/50 rounded-xl">
            <p class="text-slate-400 mb-6">
                Upload a new <code>game_server.zip</code> package here. You can add a comment to describe the changes (e.g., "v1.2.0 - Fixed bugs").
            </p>

            <div class="max-w-xl space-y-6">
                <!-- File Input -->
                <div>
                    <span class="block text-sm font-medium text-slate-300 mb-2">
                        Server Package (.zip)
                    </span>
                    <div class="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-slate-700 border-dashed rounded-md hover:border-blue-400 transition-colors bg-slate-900/30">
                        <div class="space-y-1 text-center">
                            <svg class="mx-auto h-12 w-12 text-slate-500" stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                                <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                            </svg>
                            <div class="flex text-sm text-slate-400 justify-center">
                                <label for="file-upload" class="relative cursor-pointer rounded-md font-medium text-blue-400 hover:text-blue-300 focus-within:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-blue-500">
                                    <span>Upload a file</span>
                                    <input id="file-upload" name="file-upload" type="file" class="sr-only" accept=".zip" bind:this={fileInput} onchange={() => { uploadStatus = ''; }}>
                                </label>
                                <p class="pl-1">or drag and drop</p>
                            </div>
                            <p class="text-xs text-slate-500">
                                ZIP up to 100MB
                            </p>
                        </div>
                    </div>
                    {#if fileInput && fileInput.files && fileInput.files.length > 0}
                        <div class="mt-2 text-sm text-emerald-400 flex items-center gap-2">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                            </svg>
                            Selected: {fileInput.files[0].name}
                        </div>
                    {/if}
                </div>

                <!-- Version Input -->
                <div>
                    <label for="version" class="block text-sm font-medium text-slate-300 mb-2">
                        Version (e.g. 1.0.0)
                    </label>
                    <input 
                        type="text" 
                        id="version" 
                        bind:value={version}
                        class="w-full px-4 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                        placeholder="1.0.0"
                    />
                </div>

                <!-- Comment Input -->
                <div>
                    <label for="comment" class="block text-sm font-medium text-slate-300 mb-2">
                        Version Comment
                    </label>
                    <textarea 
                        id="comment" 
                        bind:value={comment}
                        rows="3" 
                        class="w-full px-4 py-2 bg-slate-900/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                        placeholder="e.g. Added new map, fixed crash on startup..."
                    ></textarea>
                </div>

                <!-- Submit -->
                <div>
                    <button 
                        onclick={handleUpload} 
                        disabled={uploading}
                        class="w-full sm:w-auto px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg shadow-lg shadow-blue-900/20 font-semibold transition-all transform active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none flex items-center justify-center gap-2"
                    >
                        {#if uploading}
                            <div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                            Uploading...
                        {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 5.414V13a1 1 0 11-2 0V5.414L7.707 6.707a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                            </svg>
                            Upload Version
                        {/if}
                    </button>
                </div>

                {#if uploadStatus}
                    <div class={`p-3 rounded-lg text-sm font-medium border flex items-center gap-2 ${uploadError ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'}`}>
                        <span>{uploadError ? '⚠️' : '✓'}</span>
                        {uploadStatus}
                    </div>
                {/if}
            </div>
        </div>
    {:else}
        <div class="card bg-slate-800/30 border border-slate-700/50 rounded-xl overflow-hidden">
            <div class="overflow-x-auto">
                <table class="w-full text-sm">
                    <thead class="bg-slate-800/50 border-b border-slate-700">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Version</th>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Uploaded</th>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Filename</th>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Comment</th>
                            <th class="px-6 py-3 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-700 text-slate-300">
                        {#each $serverVersions as version}
                            <tr class="hover:bg-slate-800/30 transition">
                                <td class="px-6 py-4 whitespace-nowrap">
                                    {#if version.is_active}
                                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-sm shadow-emerald-900/20">
                                            <span class="w-1.5 h-1.5 bg-emerald-400 rounded-full mr-1.5 animate-pulse"></span>
                                            Active
                                        </span>
                                    {:else}
                                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-700 text-slate-400">
                                            Inactive
                                        </span>
                                    {/if}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-slate-300 font-medium">
                                    {version.version || 'Unknown'}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-slate-400">
                                    {new Date(version.uploaded_at).toLocaleString()}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap font-mono text-xs text-slate-400">
                                    {version.filename}
                                </td>
                                <td class="px-6 py-4 text-slate-300 max-w-xs truncate" title={version.comment}>
                                    {version.comment || '-'}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                                    {#if !version.is_active}
                                        <button 
                                            onclick={() => requestActivate(version.id)}
                                            class="px-3 py-1 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-600/30 rounded-md text-xs font-semibold transition-colors"
                                        >
                                            Activate
                                        </button>
                                        <button 
                                            onclick={() => requestDelete(version.id)}
                                            class="px-3 py-1 bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-600/30 rounded-md text-xs font-semibold transition-colors"
                                        >
                                            Delete
                                        </button>
                                    {:else}
                                        <span class="text-slate-600 cursor-not-allowed text-xs font-medium px-2">Active</span>
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                        {#if $serverVersions.length === 0}
                            <tr>
                                <td colspan="5" class="px-6 py-8 text-center text-slate-500">
                                    No versions uploaded yet.
                                </td>
                            </tr>
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title={confirmTitle}
    message={confirmMessage}
    confirmText={confirmButtonText}
    isCritical={confirmIsCritical}
    onConfirm={confirmAction}
/>