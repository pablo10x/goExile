<script lang="ts">
    import { onMount } from 'svelte';
    import { fade, fly, slide } from 'svelte/transition';
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
        Ban,
        BarChart3,
        ShieldCheck,
        Settings,
        Save,
        Unlock
    } from 'lucide-svelte';
    
    interface RedEyeRule {
        id: number;
        name: string;
        cidr: string;
        port: string;
        path_pattern: string;
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
        player_id: string;
    }

    interface RedEyeStats {
        total_rules: number;
        active_bans: number;
        events_24h: number;
        logs_24h: number;
        reputation_count: number;
    }

    interface RedEyeConfig {
        "redeye.auto_ban_enabled": boolean;
        "redeye.auto_ban_threshold": number;
        "redeye.alert_enabled": boolean;
    }

    interface BannedIP {
        ip: string;
        reputation_score: number;
        ban_reason: string;
        ban_expires_at: string | null;
        last_seen: string;
    }

    let activeTab = $state<'overview' | 'rules' | 'bans' | 'logs' | 'anticheat' | 'config'>('overview');
    let rules = $state<RedEyeRule[]>([]);
    let logs = $state<RedEyeLog[]>([]);
    let events = $state<AnticheatEvent[]>([]);
    let bans = $state<BannedIP[]>([]);
    let stats = $state<RedEyeStats>({ total_rules: 0, active_bans: 0, events_24h: 0, logs_24h: 0, reputation_count: 0 });
    let config = $state<RedEyeConfig>({ "redeye.auto_ban_enabled": true, "redeye.auto_ban_threshold": 100, "redeye.alert_enabled": true });
    let loading = $state(true);

    // Modal
    let showModal = $state(false);
    let editingRule = $state<RedEyeRule | null>(null);
    let form = $state({
        name: '',
        cidr: '',
        port: '*',
        path_pattern: '',
        protocol: 'ANY',
        action: 'DENY',
        rate_limit: 0,
        burst: 0,
        enabled: true
    });

    async function fetchStats() {
        try {
            const res = await fetch('/api/redeye/stats');
            if (res.ok) stats = await res.json();
        } catch (e) {
            console.error(e);
        }
    }

    async function fetchConfig() {
        try {
            const res = await fetch('/api/redeye/config');
            if (res.ok) config = await res.json();
        } catch (e) {
            console.error(e);
        }
    }

    async function updateConfig() {
        try {
            const res = await fetch('/api/redeye/config', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            if (res.ok) {
                alert('Configuration updated successfully');
            } else {
                alert('Failed to update configuration');
            }
        } catch (e) {
            console.error(e);
        }
    }

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

    async function fetchBans() {
        loading = true;
        try {
            const res = await fetch('/api/redeye/bans');
            if (res.ok) bans = await res.json();
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function unbanIP(ip: string) {
        if (!confirm(`Unban IP ${ip}?`)) return;
        try {
            const res = await fetch(`/api/redeye/bans/${ip}`, { method: 'DELETE' });
            if (res.ok) {
                bans = bans.filter(b => b.ip !== ip);
                fetchStats();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function refreshAll() {
        fetchStats();
        if (activeTab === 'rules') fetchRules();
        else if (activeTab === 'logs') fetchLogs();
        else if (activeTab === 'anticheat') fetchEvents();
        else if (activeTab === 'bans') fetchBans();
        else if (activeTab === 'config') fetchConfig();
    }

    async function saveRule() {
        try {
            const url = editingRule ? `/api/redeye/rules/${editingRule.id}` : '/api/redeye/rules';
            const method = editingRule ? 'PUT' : 'POST';
            
            const payload = { ...form };
            
            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                showModal = false;
                fetchRules();
                fetchStats();
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
                fetchStats();
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
                path_pattern: '',
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
        fetchStats();
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
            <button onclick={refreshAll} class="p-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg transition-colors">
                <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
            </button>
        </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 mb-4 shrink-0 overflow-x-auto">
        <button 
            onclick={() => { activeTab = 'overview'; fetchStats(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'overview' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Overview
        </button>
        <button 
            onclick={() => { activeTab = 'rules'; fetchRules(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'rules' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Rules
        </button>
        <button 
            onclick={() => { activeTab = 'bans'; fetchBans(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'bans' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Bans
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
        <button 
            onclick={() => { activeTab = 'config'; fetchConfig(); }}
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all {activeTab === 'config' ? 'bg-red-600 text-white shadow-lg shadow-red-900/20' : 'bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700'}"
        >
            Configuration
        </button>
    </div>

    <!-- Content -->
    <div class="flex-1 bg-slate-900/50 border border-slate-800 rounded-xl overflow-hidden flex flex-col relative">
        {#if activeTab === 'overview'}
            <div class="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 overflow-auto">
                <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700 flex flex-col items-center justify-center text-center">
                    <ShieldCheck class="w-8 h-8 text-green-400 mb-2" />
                    <span class="text-3xl font-bold text-white">{stats.total_rules}</span>
                    <span class="text-xs text-slate-400 uppercase tracking-wider">Active Rules</span>
                </div>
                <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700 flex flex-col items-center justify-center text-center">
                    <Ban class="w-8 h-8 text-red-500 mb-2" />
                    <span class="text-3xl font-bold text-white">{stats.active_bans}</span>
                    <span class="text-xs text-slate-400 uppercase tracking-wider">Banned IPs</span>
                </div>
                <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700 flex flex-col items-center justify-center text-center">
                    <ShieldAlert class="w-8 h-8 text-yellow-400 mb-2" />
                    <span class="text-3xl font-bold text-white">{stats.events_24h}</span>
                    <span class="text-xs text-slate-400 uppercase tracking-wider">Events (24h)</span>
                </div>
                <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700 flex flex-col items-center justify-center text-center">
                    <Activity class="w-8 h-8 text-blue-400 mb-2" />
                    <span class="text-3xl font-bold text-white">{stats.logs_24h}</span>
                    <span class="text-xs text-slate-400 uppercase tracking-wider">Traffic Logs (24h)</span>
                </div>

                <!-- Threat Map Placeholder -->
                <div class="col-span-1 md:col-span-2 lg:col-span-4 bg-slate-950/50 border border-slate-800 rounded-xl p-6 flex flex-col items-center justify-center h-64 mt-4 relative overflow-hidden group">
                    <div class="absolute inset-0 bg-[url('/grid.svg')] opacity-10"></div>
                    <div class="absolute inset-0 bg-gradient-to-t from-slate-950 via-transparent to-transparent"></div>
                    
                    <div class="relative z-10 text-center">
                        <Activity class="w-12 h-12 text-red-500 mx-auto mb-4 animate-pulse" />
                        <h3 class="text-xl font-bold text-white">Live Threat Monitor</h3>
                        <p class="text-slate-400 text-sm max-w-md mx-auto mt-2">
                            System is actively monitoring traffic and enforcing {stats.total_rules} rules.
                            Reputation system tracking {stats.reputation_count} IPs.
                        </p>
                    </div>
                </div>
            </div>
        {:else if activeTab === 'rules'}
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
                            <th class="px-4 py-3 font-semibold">Port</th>
                            <th class="px-4 py-3 font-semibold">Path</th>
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
                                <td class="px-4 py-3 font-mono text-xs">{rule.port}</td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-300">{rule.path_pattern || '-'}</td>
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
        {:else if activeTab === 'bans'}
            <div class="flex-1 overflow-auto custom-scrollbar">
                <table class="w-full text-left text-sm text-slate-400">
                    <thead class="bg-slate-900 text-slate-200 sticky top-0 z-10">
                        <tr>
                            <th class="px-4 py-3 font-semibold">IP Address</th>
                            <th class="px-4 py-3 font-semibold">Reputation</th>
                            <th class="px-4 py-3 font-semibold">Reason</th>
                            <th class="px-4 py-3 font-semibold">Last Seen</th>
                            <th class="px-4 py-3 font-semibold">Expires</th>
                            <th class="px-4 py-3 font-semibold text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each bans as ban}
                            <tr class="hover:bg-slate-800/50 transition-colors">
                                <td class="px-4 py-3 font-mono text-white">{ban.ip}</td>
                                <td class="px-4 py-3">
                                    <div class="flex items-center gap-2">
                                        <span class="text-xs font-bold text-red-400">{ban.reputation_score}</span>
                                        <div class="w-16 h-1.5 bg-slate-700 rounded-full overflow-hidden">
                                            <div class="h-full bg-red-500" style="width: {ban.reputation_score}%"></div>
                                        </div>
                                    </div>
                                </td>
                                <td class="px-4 py-3 text-slate-300 max-w-xs truncate" title={ban.ban_reason}>{ban.ban_reason}</td>
                                <td class="px-4 py-3 font-mono text-xs">{new Date(ban.last_seen).toLocaleString()}</td>
                                <td class="px-4 py-3 font-mono text-xs text-slate-400">{ban.ban_expires_at ? new Date(ban.ban_expires_at).toLocaleString() : 'Permanent'}</td>
                                <td class="px-4 py-3 text-right">
                                    <button onclick={() => unbanIP(ban.ip)} class="px-2 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 rounded text-xs transition-colors flex items-center gap-1 ml-auto">
                                        <Unlock class="w-3 h-3" />
                                        Unban
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
        {:else if activeTab === 'config'}
            <div class="p-6 max-w-3xl mx-auto space-y-8">
                <div class="bg-slate-800/50 p-6 rounded-xl border border-slate-700">
                    <h3 class="text-lg font-bold text-white mb-4 flex items-center gap-2">
                        <Zap class="w-5 h-5 text-yellow-400" />
                        Automated Response
                    </h3>
                    <div class="space-y-6">
                        <div class="flex items-center justify-between">
                            <div>
                                <span class="block text-slate-200 font-medium">Auto-Ban High Risk IPs</span>
                                <span class="block text-xs text-slate-400 mt-1">Automatically create DENY rules for IPs that exceed reputation threshold</span>
                            </div>
                            <label class="relative inline-flex items-center cursor-pointer">
                                <input type="checkbox" bind:checked={config['redeye.auto_ban_enabled']} class="sr-only peer">
                                <div class="w-11 h-6 bg-slate-700 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-red-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-red-600"></div>
                            </label>
                        </div>

                        <div>
                            <label for="autoBanThreshold" class="block text-slate-200 font-medium mb-2">Reputation Threshold</label>
                            <div class="flex gap-4 items-center">
                                <input id="autoBanThreshold" type="range" min="10" max="200" step="10" bind:value={config['redeye.auto_ban_threshold']} class="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer accent-red-500">
                                <span class="w-12 text-center font-mono bg-slate-900 rounded px-2 py-1 border border-slate-700 text-white">{config['redeye.auto_ban_threshold']}</span>
                            </div>
                            <span class="block text-xs text-slate-400 mt-1">Score > {config['redeye.auto_ban_threshold']} triggers ban</span>
                        </div>
                    </div>
                </div>

                <div class="bg-slate-800/50 p-6 rounded-xl border border-slate-700">
                    <h3 class="text-lg font-bold text-white mb-4 flex items-center gap-2">
                        <ShieldAlert class="w-5 h-5 text-blue-400" />
                        Alerts
                    </h3>
                    <div class="flex items-center justify-between">
                        <div>
                            <span class="block text-slate-200 font-medium">Enable Alerts</span>
                            <span class="block text-xs text-slate-400 mt-1">Send notifications for critical events</span>
                        </div>
                        <label class="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" bind:checked={config['redeye.alert_enabled']} class="sr-only peer">
                            <div class="w-11 h-6 bg-slate-700 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                        </label>
                    </div>
                </div>

                <div class="flex justify-end">
                    <button onclick={updateConfig} class="px-6 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-medium flex items-center gap-2 transition-colors shadow-lg shadow-red-900/20">
                        <Save class="w-4 h-4" />
                        Save Configuration
                    </button>
                </div>
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
    <div class="fixed inset-0 z-[999] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" transition:fade onclick={() => showModal = false} role="button" tabindex="0" onkeydown={(e) => { if (e.key === 'Escape') showModal = false; }}>
        <div class="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-lg overflow-hidden" onclick={(e) => e.stopPropagation()}>
            <div class="p-6 border-b border-slate-800 flex justify-between items-center">
                <h3 class="text-lg font-bold text-white">{editingRule ? 'Edit Rule' : 'New Rule'}</h3>
                <button onclick={() => showModal = false} class="text-slate-500 hover:text-white">✕</button>
            </div>
            <div class="p-6 space-y-4">
                <div>
                    <label for="ruleName" class="block text-sm font-medium text-slate-400 mb-1">Name</label>
                    <input id="ruleName" type="text" bind:value={form.name} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label for="ruleCidrIp" class="block text-sm font-medium text-slate-400 mb-1">CIDR / IP</label>
                        <input id="ruleCidrIp" type="text" bind:value={form.cidr} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                    </div>
                    <div>
                        <label for="rulePort" class="block text-sm font-medium text-slate-400 mb-1">Port</label>
                        <input id="rulePort" type="text" bind:value={form.port} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" placeholder="*" />
                    </div>
                </div>
                <div>
                    <label for="rulePathPattern" class="block text-sm font-medium text-slate-400 mb-1">Path Pattern (Optional)</label>
                    <input id="rulePathPattern" type="text" bind:value={form.path_pattern} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" placeholder="/api/v1/..." />
                </div>
                <fieldset>
                    <legend class="block text-sm font-medium text-slate-400 mb-1">Action</legend>
                    <div class="flex gap-4">
                        <label for="actionAllow" class="flex items-center gap-2 cursor-pointer">
                            <input id="actionAllow" type="radio" bind:group={form.action} value="ALLOW" class="text-green-500 focus:ring-green-500" />
                            <span class="text-slate-300">Allow</span>
                        </label>
                        <label for="actionDeny" class="flex items-center gap-2 cursor-pointer">
                            <input id="actionDeny" type="radio" bind:group={form.action} value="DENY" class="text-red-500 focus:ring-red-500" />
                            <span class="text-slate-300">Deny</span>
                        </label>
                        <label for="actionRateLimit" class="flex items-center gap-2 cursor-pointer">
                            <input id="actionRateLimit" type="radio" bind:group={form.action} value="RATE_LIMIT" class="text-yellow-500 focus:ring-yellow-500" />
                            <span class="text-slate-300">Rate Limit</span>
                        </label>
                    </div>
                </fieldset>
                {#if form.action === 'RATE_LIMIT'}
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label for="ruleLimit" class="block text-sm font-medium text-slate-400 mb-1">Limit (req/s)</label>
                            <input id="ruleLimit" type="number" bind:value={form.rate_limit} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
                        </div>
                        <div>
                            <label for="ruleBurst" class="block text-sm font-medium text-slate-400 mb-1">Burst</label>
                            <input id="ruleBurst" type="number" bind:value={form.burst} class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-red-500" />
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
