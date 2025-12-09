<script lang="ts">
    import { onMount } from 'svelte';
    
    export let spawnerId: number;

    let logs: string = '';
    let loading: boolean = true;
    let error: string = '';

    async function fetchLogs() {
        loading = true;
        error = '';
        try {
            const res = await fetch(`/api/spawners/${spawnerId}/logs`);
            if (res.ok) {
                const data = await res.json();
                logs = data.logs || 'No logs available.';
            } else {
                error = `Failed to fetch logs: ${res.statusText}`;
            }
        } catch (e: any) {
            error = `Error: ${e.message}`;
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        fetchLogs();
    });
</script>

<div class="h-full flex flex-col">
    <div class="flex justify-between items-center mb-4">
        <div class="text-sm text-slate-400">
            Viewing logs for Spawner #{spawnerId}
        </div>
        <button 
            onclick={fetchLogs} 
            disabled={loading}
            class="px-3 py-1 bg-blue-600/20 text-blue-400 hover:bg-blue-600/30 rounded text-xs font-semibold transition-colors disabled:opacity-50"
        >
            {loading ? 'Refreshing...' : 'Refresh Logs'}
        </button>
    </div>

    <div class="flex-1 bg-slate-950 rounded-lg border border-slate-800 p-4 overflow-hidden flex flex-col relative">
        {#if loading && !logs}
            <div class="absolute inset-0 flex items-center justify-center text-slate-500">
                Loading...
            </div>
        {:else if error}
            <div class="absolute inset-0 flex items-center justify-center text-red-400 p-4 text-center">
                {error}
            </div>
        {:else}
            <pre class="font-mono text-xs text-slate-300 whitespace-pre-wrap overflow-y-auto h-full scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-slate-900 pr-2">{logs}</pre>
        {/if}
    </div>
</div>