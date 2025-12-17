<script lang="ts">
	import { Play, RefreshCw, Terminal, Download, Copy, Check } from 'lucide-svelte';
	import { notifications } from '$lib/stores';
    import { slide } from 'svelte/transition';

	let query = $state('');
	let loading = $state(false);
	let results = $state<any[]>([]);
	let error = $state<string | null>(null);
    let executionTime = $state(0);

	async function runQuery() {
		if (!query.trim()) return;
		loading = true;
		error = null;
        results = [];
        const start = performance.now();

		try {
			const res = await fetch('/api/database/sql', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query })
			});
			const data = await res.json();
			if (!res.ok) throw new Error(data.error || 'Query failed');
			results = data || [];
            if (results.length === 0 && !data.error) {
                notifications.add({ type: 'success', message: 'Query executed successfully (No results)' });
            }
		} catch (e: any) {
			error = e.message;
            notifications.add({ type: 'error', message: 'SQL Error', details: e.message });
		} finally {
			loading = false;
            executionTime = performance.now() - start;
		}
	}

    function downloadCSV() {
        if (!results.length) return;
        const headers = Object.keys(results[0]);
        const csvContent = [
            headers.join(','),
            ...results.map(row => headers.map(fieldName => JSON.stringify(row[fieldName])).join(','))
        ].join('\n');
        
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.setAttribute('href', url);
        link.setAttribute('download', `query_results_${Date.now()}.csv`);
        link.click();
    }

    function copyToClipboard() {
        navigator.clipboard.writeText(JSON.stringify(results, null, 2));
        notifications.add({ type: 'success', message: 'Results copied to clipboard' });
    }
</script>

<div class="h-full flex flex-col bg-slate-950">
    <!-- Toolbar -->
    <div class="p-4 border-b border-slate-800 flex justify-between items-center bg-slate-900/50">
        <div class="flex items-center gap-3">
            <div class="p-2 bg-amber-500/10 rounded-lg">
                <Terminal class="w-5 h-5 text-amber-400" />
            </div>
            <div>
                <h2 class="text-lg font-bold text-slate-100">SQL Console</h2>
                <p class="text-xs text-slate-500">Execute raw SQL queries against the database</p>
            </div>
        </div>
        <div class="flex gap-2">
            <button 
                onclick={runQuery} 
                disabled={loading || !query.trim()}
                class="px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-lg font-bold flex items-center gap-2 shadow-lg shadow-amber-900/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
                {#if loading}
                    <RefreshCw class="w-4 h-4 animate-spin" /> Executing...
                {:else}
                    <Play class="w-4 h-4 fill-current" /> Run
                {/if}
            </button>
        </div>
    </div>

    <!-- Editor & Results Split -->
    <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Editor Area -->
        <div class="h-1/3 min-h-[150px] p-4 border-b border-slate-800 bg-slate-900/30">
            <textarea
                bind:value={query}
                class="w-full h-full bg-slate-950 border border-slate-700 rounded-xl p-4 font-mono text-sm text-blue-300 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 outline-none resize-none shadow-inner"
                placeholder="SELECT * FROM table_name LIMIT 10;"
                onkeydown={(e) => {
                    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                        runQuery();
                    }
                }}
            ></textarea>
            <div class="text-xs text-slate-500 mt-2 flex justify-between">
                <span>Cmd/Ctrl + Enter to execute</span>
                {#if executionTime > 0}
                    <span class="text-emerald-400">Time: {executionTime.toFixed(2)}ms</span>
                {/if}
            </div>
        </div>

        <!-- Results Area -->
        <div class="flex-1 flex flex-col overflow-hidden bg-slate-900/50 relative">
            {#if results.length > 0}
                <div class="p-2 border-b border-slate-800 bg-slate-900 flex justify-between items-center px-4">
                    <span class="text-xs font-bold text-slate-400 uppercase tracking-wider">{results.length} Rows</span>
                    <div class="flex gap-2">
                        <button onclick={copyToClipboard} class="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded" title="Copy JSON">
                            <Copy class="w-4 h-4" />
                        </button>
                        <button onclick={downloadCSV} class="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded" title="Download CSV">
                            <Download class="w-4 h-4" />
                        </button>
                    </div>
                </div>
                <div class="flex-1 overflow-auto">
                    <table class="w-full text-left text-sm border-collapse">
                        <thead class="bg-slate-950 text-slate-400 sticky top-0 shadow-md z-10">
                            <tr>
                                {#each Object.keys(results[0]) as key}
                                    <th class="px-4 py-3 font-medium border-b border-slate-800 whitespace-nowrap">{key}</th>
                                {/each}
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-800/50">
                            {#each results as row}
                                <tr class="hover:bg-slate-800/50 transition-colors">
                                    {#each Object.values(row) as val}
                                        <td class="px-4 py-2 text-slate-300 font-mono whitespace-nowrap max-w-xs truncate" title={String(val)}>
                                            {val === null ? 'NULL' : String(val)}
                                        </td>
                                    {/each}
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {:else if error}
                <div class="flex-1 flex items-center justify-center p-8">
                    <div class="bg-red-500/10 border border-red-500/20 rounded-xl p-6 max-w-2xl text-center">
                        <div class="text-red-400 font-mono text-sm whitespace-pre-wrap">{error}</div>
                    </div>
                </div>
            {:else if !loading}
                <div class="flex-1 flex flex-col items-center justify-center text-slate-500">
                    <Terminal class="w-16 h-16 opacity-20 mb-4" />
                    <p>Enter a query and run it to see results.</p>
                </div>
            {/if}
            
            {#if loading}
                <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center z-20">
                    <div class="flex flex-col items-center gap-3">
                        <div class="w-10 h-10 border-4 border-amber-500 border-t-transparent rounded-full animate-spin"></div>
                        <span class="text-amber-400 font-medium animate-pulse">Running Query...</span>
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>
