<script lang="ts">
    import { onMount } from 'svelte';
    import { serverVersions } from '$lib/stores';

    let activeTab = 'upload'; // 'upload' | 'history'
    
    // Upload State
    let fileInput: HTMLInputElement;
    let comment = '';
    let uploading = false;
    let uploadStatus = '';
    let uploadError = false;

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

    async function activateVersion(id: number) {
        if (!confirm('Activate this version? Spawners will download it on next check.')) return;
        try {
            const res = await fetch(`/api/versions/${id}/active`, { method: 'POST' });
            if (res.ok) {
                loadVersions();
            } else {
                alert('Failed to activate version');
            }
        } catch (e) {
            alert('Error activating version');
        }
    }

    async function deleteVersion(id: number) {
        if (!confirm('Delete this version? This cannot be undone.')) return;
        try {
            const res = await fetch(`/api/versions/${id}`, { method: 'DELETE' });
            if (res.ok) {
                loadVersions();
            } else {
                const data = await res.json();
                alert('Failed to delete: ' + (data.error || 'Unknown error'));
            }
        } catch (e) {
            alert('Error deleting version');
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
        <div class="card p-8">
            <p class="text-slate-400 mb-6">
                Upload a new <code>game_server.zip</code> package here. You can add a comment to describe the changes (e.g., "v1.2.0 - Fixed bugs").
            </p>

            <div class="max-w-xl space-y-6">
                <!-- File Input -->
                <div>
                    <label class="block text-sm font-medium text-slate-300 mb-2">
                        Server Package (.zip)
                    </label>
                    <div class="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-slate-700 border-dashed rounded-md hover:border-blue-400 transition-colors bg-slate-800/30">
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
                        <div class="mt-2 text-sm text-slate-300">Selected: {fileInput.files[0].name}</div>
                    {/if}
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
                        class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                        placeholder="e.g. Added new map, fixed crash on startup..."
                    ></textarea>
                </div>

                <!-- Submit -->
                <div>
                    <button 
                        onclick={handleUpload} 
                        disabled={uploading}
                        class="btn-primary w-full sm:w-auto justify-center disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {uploading ? 'Uploading...' : 'Upload Version'}
                    </button>
                </div>

                {#if uploadStatus}
                    <div class={`p-3 rounded text-sm font-semibold border ${uploadError ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'}`}>
                        {uploadStatus}
                    </div>
                {/if}
            </div>
        </div>
    {:else}
        <div class="card overflow-hidden">
            <div class="overflow-x-auto">
                <table class="w-full text-sm">
                    <thead class="bg-slate-800/50 border-b border-slate-700">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
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
                                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                            Active
                                        </span>
                                    {:else}
                                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-700 text-slate-400">
                                            Inactive
                                        </span>
                                    {/if}
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
                                            onclick={() => activateVersion(version.id)}
                                            class="text-blue-400 hover:text-blue-300"
                                        >
                                            Activate
                                        </button>
                                        <button 
                                            onclick={() => deleteVersion(version.id)}
                                            class="text-red-400 hover:text-red-300"
                                        >
                                            Delete
                                        </button>
                                    {:else}
                                        <span class="text-slate-600 cursor-not-allowed">Active</span>
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