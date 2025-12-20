<script lang="ts">
    import { onMount } from 'svelte';
    import { fade, fly } from 'svelte/transition';
    import { 
        Eye, 
        Plus, 
        Trash2, 
        Edit2, 
        RefreshCw, 
        ShieldAlert, 
        Activity,
        Zap,
        Users,
        Ban
    } from 'lucide-svelte';
    
    // Types would usually be imported, defining locally for speed in prototype
    interface RedEyeRule {
        id: number;
        name: string;
        cidr: string;
        port: string;
        protocol: string;
        action: 'ALLOW' | 'DENY' | 'RATE_LIMIT';
        rate_limit: number;
        burst: number;
        enabled: boolean;
    }

    interface RedEyeLog {
        id: number;
        source_ip: string;
        action: string;
        timestamp: string;
    }

    interface AnticheatEvent {
        id: number;
        event_type: string;
        client_ip: string;
        severity: number;
        timestamp: string;
    }

    let activeTab = $state<'rules' | 'logs' | 'anticheat'>('rules');
    let rules = $state<RedEyeRule[]>([]);
    let logs = $state<RedEyeLog[]>([]);
    let events = $state<AnticheatEvent[]>([]);
    let loading = $state(true);

    // Modal
    let showModal = $state(false);
    let editingRule = $state<RedEyeRule | null>(null);
    let form = $state({
        name: '',
        cidr: '',
        port: '*',
        protocol: 'ANY',
        action: 'DENY',
        rate_limit: 0,
        burst: 0,
        enabled: true
    });

    async function fetchRules() {
        loading = true;
        try {
            const res = await fetch('/api/redeye/rules');
            if (res.ok) rules = await res.json();
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function fetchLogs() {
        loading = true;
        try {
            const res = await fetch('/api/redeye/logs?limit=100');
            if (res.ok) {
                const data = await res.json();
                logs = data.logs;
            }
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function fetchEvents() {
        loading = true;
        try {
            const res = await fetch('/api/redeye/anticheat/events?limit=50');
            if (res.ok) {
                const data = await res.json();
                events = data.events;
            }
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function saveRule() {
        try {
            const url = editingRule ? `/api/redeye/rules/${editingRule.id}` : '/api/redeye/rules';
            const method = editingRule ? 'PUT' : 'POST';
            
            // Convert strings to ints if needed (svelte binding usually handles this but safety first)
            const payload = { ...form };
            
            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                showModal = false;
                fetchRules();
            } else {
                alert('Failed to save rule');
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function deleteRule(id: number) {
        if (!confirm('Delete this rule?')) return;
        try {
            const res = await fetch(`/api/redeye/rules/${id}`, { method: 'DELETE' });
            if (res.ok) {
                rules = rules.filter(r => r.id !== id);
            }
        } catch (e) {
            console.error(e);
        }
    }

    function openModal(rule: RedEyeRule | null = null) {
        editingRule = rule;
        if (rule) {
            form = { ...rule };
        } else {
            form = {
                name: '',
                cidr: '',
                port: '*',
                protocol: 'ANY',
                action: 'DENY',
                rate_limit: 10,
                burst: 20,
                enabled: true
            };
        }
        showModal = true;
    }

    function getActionColor(action: string) {
        switch(action) {
            case 'ALLOW': return 'text-green-400 bg-green-900/30 border-green-800';
            case 'DENY': return 'text-red-400 bg-red-900/30 border-red-800';
            case 'RATE_LIMIT': return 'text-yellow-400 bg-yellow-900/30 border-yellow-800';
            default: return 'text-slate-400 bg-slate-800 border-slate-700';
        }
    }

    onMount(() => {
        fetchRules();
    });
</script>

<div class="p-6 h-full flex flex-col overflow-hidden">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6 shrink-0">
        <div>
            <h1 class="text-2xl font-bold text-white flex items-center gap-2">
                <Eye class="w-6 h-6 text-red-500" />
                <span class="bg-gradient-to-r from-red-500 to-rose-500 bg-clip-text text-transparent">RedEye</span> Guardian
            </h1>
            <p class="text-slate-400 text-sm mt-1">Advanced threat detection and access control</p>
        </div>
        <div class="flex gap-2">
            <button onclick={() => activeTab === 'rules' ? fetchRules() : activeTab === 'logs' ? fetchLogs() : fetchEvents()} class="p-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg transition-colors">
                <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
            </button>
        </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 mb-4 shrink-0">
        <button 
            onclick={() => { activeTab = 'rules'; fetchRules(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'rules' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Rules
        </button>
        <button 
            onclick={() => { activeTab = 'anticheat'; fetchEvents(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'anticheat' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Anti-Cheat
        </button>
        <button 
            onclick={() => { activeTab = 'logs'; fetchLogs(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'logs' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Logs
        </button>
    </div>

    <!-- Content -->
    <div class="flex-1 bg-slate-900/50 border border-slate-800 rounded-xl overflow-hidden flex flex-col relative">
        {#if activeTab === 'rules'}
            <div class="p-4 border-b border-slate-800 flex justify-end">
                <button onclick={() => openModal()} class="px-3 py-1.5 bg-red-600 hover:bg-red-500 text-white rounded-lg text-sm font-medium flex items-center gap-2 transition-colors shadow-lg shadow-red-900/20">
                    <Plus class="w-4 h-4" />
                    Add Rule
                </button>
            </div>
            <div class="flex-1 overflow-auto custom-scrollbar">
                <table class="w-full text-left text-sm text-slate-400">
                    <thead class="bg-slate-900 text-slate-200 sticky top-0 z-10">
                        <tr>
                            <th class="px-4 py-3 font-semibold">Name</th>
                            <th class="px-4 py-3 font-semibold">CIDR / IP</th>
                            <th class="px-4 py-3 font-semibold">Action</th>
                            <th class="px-4 py-3 font-semibold">Rate Limit</th>
                            <th class="px-4 py-3 font-semibold text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each rules as rule (rule.id)}
                            <tr class="hover:bg-slate-800/50 transition-colors">
                                <td class="px-4 py-3 font-medium text-slate-200">{rule.name}</td>
                                <td class="px-4 py-3 font-mono text-xs">{rule.cidr}</td>
                                <td class="px-4 py-3">
                                    <span class="px-2 py-0.5 text-xs font-bold rounded border {getActionColor(rule.action)}">
                                        {rule.action}
                                    </span>
                                </td>
                                <td class="px-4 py-3 font-mono text-xs">
                                    {#if rule.action === 'RATE_LIMIT'}
                                        {rule.rate_limit}/s (Burst: {rule.burst})
                                    {:else}
                                        -
                                    {/if}
                                </td>
                                <td class="px-4 py-3 text-right flex justify-end gap-2">
                                    <button onclick={() => openModal(rule)} class="p-1.5 hover:bg-slate-700 rounded text-slate-400 hover:text-white transition-colors">
                                        <Edit2 class="w-4 h-4" />
                                    </button>
                                    <button onclick={() => deleteRule(rule.id)} class="p-1.5 hover:bg-red-900/30 rounded text-slate-400 hover:text-red-400 transition-colors">
                                        <Trash2 class="w-4 h-4" />
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {:else if activeTab === 'anticheat'}
            <div class="flex-1 overflow-auto custom-scrollbar">
                <table class="w-full text-left text-sm text-slate-400">
                    <thead class="bg-slate-900 text-slate-200 sticky top-0 z-10">
                        <tr>
                            <th class="px-4 py-3 font-semibold">Time</th>
                            <th class="px-4 py-3 font-semibold">Event</th>
                            <th class="px-4 py-3 font-semibold">Severity</th>
                            <th class="px-4 py-3 font-semibold">Client IP</th>
                            <th class="px-4 py-3 font-semibold">Player ID</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each events as event}
                            <tr class="hover:bg-slate-800/50 transition-colors">
                                <td class="px-4 py-3 font-mono text-xs">{new Date(event.timestamp).toLocaleString()}</td>
                                <td class="px-4 py-3 text-white font-bold">{event.event_type}</td>
                                <td class="px-4 py-3">
                                    <div class="w-24 h-2 bg-slate-700 rounded-full overflow-hidden">
                                        <div class="h-full {event.severity > 80 ? 'bg-red-500' : event.severity > 50 ? 'bg-yellow-500' : 'bg-blue-500'}" style="width: {event.severity}%"></div>
                                    </div>
                                </td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-300">{event.client_ip}</td>
                                <td class="px-4 py-3 font-mono text-xs">{event.player_id || '-'}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {:else}
            <!-- Logs -->
            <div class="flex-1 overflow-auto custom-scrollbar">
                <table class="w-full text-left text-sm text-slate-400">
                    <thead class="bg-slate-900 text-slate-200 sticky top-0 z-10">
                        <tr>
                            <th class="px-4 py-3 font-semibold">Time</th>
                            <th class="px-4 py-3 font-semibold">Action</th>
                            <th class="px-4 py-3 font-semibold">Source IP</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each logs as log}
                            <tr class="hover:bg-slate-800/50 transition-colors">
                                <td class="px-4 py-3 font-mono text-xs">{new Date(log.timestamp).toLocaleString()}</td>
                                <td class="px-4 py-3">
                                    <span class="px-2 py-0.5 text-xs font-bold rounded {getActionColor(log.action)}">
                                        {log.action}
                                    </span>
                                </td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-300">{log.source_ip}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>

<!-- Modal -->
{#if showModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" transition:fade>
        <div class="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-lg overflow-hidden" onclick={(e) => e.stopPropagation()}>
            <div class="p-6 border-b border-slate-800 flex justify-between items-center">
                <h3 class="text-lg font-bold text-white">{editingRule ? 'Edit Rule' : 'New Rule'}</h3>
                <button onclick={() => showModal = false} class="text-slate-500 hover:text-white">✕</button>
            </div>
            <div class="p-6 space-y-4">
                <div>
                    <label class="block text-sm font-medium text-slate-400 mb-1">Name</label>
                    <input type="text" bind:value={form.name} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                </div>
                <div>
                    <label class="block text-sm font-medium text-slate-400 mb-1">CIDR / IP</label>
                    <input type="text" bind:value={form.cidr} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                </div>
                <div>
                    <label class="block text-sm font-medium text-slate-400 mb-1">Action</label>
                    <div class="flex gap-4">
                        <label class="flex items-center gap-2 cursor-pointer">
                            <input type="radio" bind:group={form.action} value="ALLOW" class="text-green-500 focus:ring-green-500" />
                            <span class="text-slate-300">Allow</span>
                        </label>
                        <label class="flex items-center gap-2 cursor-pointer">
                            <input type="radio" bind:group={form.action} value="DENY" class="text-red-500 focus:ring-red-500" />
                            <span class="text-slate-300">Deny</span>
                        </label>
                        <label class="flex items-center gap-2 cursor-pointer">
                            <input type="radio" bind:group={form.action} value="RATE_LIMIT" class="text-yellow-500 focus:ring-yellow-500" />
                            <span class="text-slate-300">Rate Limit</span>
                        </label>
                    </div>
                </div>
                {#if form.action === 'RATE_LIMIT'}
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label class="block text-sm font-medium text-slate-400 mb-1">Limit (req/s)</label>
                            <input type="number" bind:value={form.rate_limit} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-slate-400 mb-1">Burst</label>
                            <input type="number" bind:value={form.burst} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                        </div>
                    </div>
                {/if}
            </div>
            <div class="p-4 border-t border-slate-800 bg-slate-950 flex justify-end gap-2">
                <button onclick={() => showModal = false} class="px-4 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">Cancel</button>
                <button onclick={saveRule} class="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white font-medium transition-colors shadow-lg shadow-red-900/20">Save Rule</button>
            </div>
        </div>
    </div>
{/if}