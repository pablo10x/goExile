<script lang="ts">
    import { onMount } from 'svelte';
    import { fade } from 'svelte/transition';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import { stats } from '$lib/stores';

    interface ErrorLog {
        timestamp: string;
        path: string;
        status: number;
        message: string;
        client_ip: string;
    }

    let errors: ErrorLog[] = [];
    let loading = true;
    let clearing = false;

    // Confirm Dialog State
    let isConfirmOpen = false;
    let confirmAction: () => Promise<void> = async () => {};

    async function fetchErrors() {
        try {
            const res = await fetch('/api/errors');
            if (res.ok) {
                errors = await res.json();
            }
        } catch (e) {
            console.error('Failed to fetch errors', e);
        } finally {
            loading = false;
        }
    }

    function requestClearErrors() {
        confirmAction = async () => await clearErrors();
        isConfirmOpen = true;
    }

    async function refreshStats() {
        try {
            const res = await fetch('/api/stats');
            if (res.ok) {
                const statsData = await res.json();
                stats.set(statsData);
            }
        } catch (e) {
            console.error('Failed to refresh stats:', e);
        }
    }

    async function clearErrors() {
        clearing = true;
        try {
            const res = await fetch('/api/errors', { method: 'DELETE' });
            if (res.ok) {
                errors = [];
                // Refresh dashboard stats after clearing errors
                await refreshStats();
            } else {
                alert('Failed to clear errors');
            }
        } catch (e) {
            alert('Error clearing logs');
        } finally {
            clearing = false;
        }
    }

    function getStatusColor(status: number) {
        if (status >= 500) return 'text-red-400 bg-red-400/10 border-red-400/20'; // Critical
        if (status >= 400) return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/20'; // Warning
        return 'text-slate-300 bg-slate-700/50 border-slate-600'; // Info/Default
    }

    function formatDate(ts: string) {
        return new Date(ts).toLocaleString();
    }

    onMount(() => {
        fetchErrors();
        const interval = setInterval(fetchErrors, 5000); // Poll every 5s
        return () => clearInterval(interval);
    });
</script>

<div class="max-w-7xl mx-auto">
    <div class="flex justify-between items-center mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-50">Error Logs</h1>
            <p class="text-slate-400 text-sm mt-1">Recent application errors and warnings</p>
        </div>
        
        <div class="flex gap-4">
            <a href="/" class="px-4 py-2 bg-slate-700 text-slate-300 hover:bg-slate-600 rounded font-semibold transition-colors">
                Back to Dashboard
            </a>
            <button 
                onclick={requestClearErrors} 
                disabled={clearing || errors.length === 0}
                class="px-4 py-2 bg-red-600 text-white hover:bg-red-500 rounded font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-red-900/20"
            >
                {clearing ? 'Clearing...' : 'Clear All Errors'}
            </button>
        </div>
    </div>

    <div class="card bg-slate-900/50 border border-slate-700/50 shadow-xl overflow-hidden">
        {#if loading}
            <div class="p-12 text-center text-slate-500 animate-pulse">
                Loading error logs...
            </div>
        {:else if errors.length === 0}
            <div class="p-16 text-center text-slate-500 flex flex-col items-center gap-4">
                <div class="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center text-3xl">✓</div>
                <div>
                    <h3 class="text-lg font-semibold text-slate-200">No Errors Found</h3>
                    <p class="text-sm">The system is running smoothly.</p>
                </div>
            </div>
        {:else}
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="border-b border-slate-700 bg-slate-800/50 text-xs uppercase tracking-wider text-slate-400 font-semibold">
                            <th class="px-6 py-4">Time</th>
                            <th class="px-6 py-4">Status</th>
                            <th class="px-6 py-4">Message</th>
                            <th class="px-6 py-4">Path</th>
                            <th class="px-6 py-4">Client IP</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each errors as error (error.timestamp + error.message)}
                            <tr transition:fade class="hover:bg-slate-800/30 transition-colors group">
                                <td class="px-6 py-4 text-sm text-slate-400 font-mono whitespace-nowrap">
                                    {formatDate(error.timestamp)}
                                </td>
                                <td class="px-6 py-4">
                                    <span class={`inline-flex items-center px-2.5 py-0.5 rounded text-xs font-medium border ${getStatusColor(error.status)}`}>
                                        {error.status}
                                    </span>
                                </td>
                                <td class="px-6 py-4 text-sm text-slate-200 font-medium">
                                    {error.message}
                                </td>
                                <td class="px-6 py-4 text-sm text-slate-400 font-mono">
                                    {error.path}
                                </td>
                                <td class="px-6 py-4 text-sm text-slate-500 font-mono">
                                    {error.client_ip}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>

<ConfirmDialog
    bind:isOpen={isConfirmOpen}
    title="Clear All Errors"
    message="Are you sure you want to clear all error logs? This cannot be undone."
    confirmText="Clear Errors"
    isCritical={true}
    onConfirm={confirmAction}
/>