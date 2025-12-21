<script lang="ts">
	import { Users, UserPlus, Trash2, Key, Shield, Check, X } from 'lucide-svelte';
	import { onMount } from 'svelte';
    import { notifications } from '$lib/stores';
    import { slide } from 'svelte/transition';

	let roles = $state<any[]>([]);
	let loading = $state(false);
    let isCreating = $state(false);
    
    // New Role Form
    let newRoleName = $state('');
    let newRolePass = $state('');
    let roleOptions = $state({
        superuser: false,
        createdb: false,
        createrole: false,
        login: true,
        replication: false,
        bypassrls: false
    });

	async function loadRoles() {
		loading = true;
		try {
			const res = await fetch('/api/database/roles');
			if (res.ok) {
				roles = await res.json();
			}
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to load roles', details: e.message });
		} finally {
			loading = false;
		}
	}

    async function createRole() {
        if (!newRoleName.trim()) return;
        try {
            const options = [];
            if (roleOptions.superuser) options.push('SUPERUSER');
            if (roleOptions.createdb) options.push('CREATEDB');
            if (roleOptions.createrole) options.push('CREATEROLE');
            if (roleOptions.login) options.push('LOGIN');
            if (roleOptions.replication) options.push('REPLICATION');
            if (roleOptions.bypassrls) options.push('BYPASSRLS');

            const res = await fetch('/api/database/roles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name: newRoleName,
                    password: newRolePass,
                    options: options
                })
            });
            
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Failed to create role');
            }
            
            notifications.add({ type: 'success', message: `Role '${newRoleName}' created` });
            isCreating = false;
            newRoleName = '';
            newRolePass = '';
            loadRoles();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Creation failed', details: e.message });
        }
    }

    async function deleteRole(name: string) {
        if (!confirm(`Delete role '${name}'?`)) return;
        try {
            const res = await fetch(`/api/database/roles/${name}`, { method: 'DELETE' });
            if (!res.ok) throw new Error('Failed to delete role');
            notifications.add({ type: 'success', message: 'Role deleted' });
            loadRoles();
        } catch (e: any) {
            notifications.add({ type: 'error', message: 'Delete failed', details: e.message });
        }
    }

	onMount(() => {
		loadRoles();
	});
</script>

<div class="h-full flex flex-col bg-slate-900">
    <div class="p-6 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 flex justify-between items-center">
        <div class="flex items-center gap-3">
            <div class="p-2 bg-emerald-500/10 rounded-lg">
                <Shield class="w-6 h-6 text-emerald-400" />
            </div>
            <div>
                <h2 class="text-xl font-bold text-slate-100">Role Management</h2>
                <p class="text-sm text-slate-500">Manage database users and privileges</p>
            </div>
        </div>
        <button 
            onclick={() => isCreating = !isCreating}
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 dark:text-white rounded-lg font-bold flex items-center gap-2 shadow-lg shadow-emerald-900/20 transition-all"
        >
            {#if isCreating}
                <X class="w-4 h-4" /> Cancel
            {:else}
                <UserPlus class="w-4 h-4" /> Create Role
            {/if}
        </button>
    </div>

    <div class="flex-1 overflow-auto p-6 relative">
        {#if isCreating}
            <div class="mb-8 bg-slate-800/50 border border-slate-300 dark:border-slate-700 rounded-xl p-6 animate-in slide-in-from-top-4 shadow-xl">
                <h3 class="text-lg font-bold text-slate-800 dark:text-slate-200 mb-4 flex items-center gap-2">
                    <UserPlus class="w-5 h-5 text-emerald-400" /> New Role Definition
                </h3>
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                    <div class="space-y-4">
                        <div>
                            <label for="newRoleName" class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-1">Role Name</label>
                            <input id="newRoleName" type="text" bind:value={newRoleName} class="w-full bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 focus:border-emerald-500 outline-none" placeholder="e.g. app_user" />
                        </div>
                        <div>
                            <label for="newRolePass" class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-1">Password</label>
                            <input id="newRolePass" type="password" bind:value={newRolePass} class="w-full bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 focus:border-emerald-500 outline-none" placeholder="••••••••" />
                        </div>
                    </div>
                    
                    <div class="space-y-3">
                        <div class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-1">Privileges</div>
                        <div class="grid grid-cols-2 gap-2">
                            <label class="flex items-center gap-2 p-2 rounded bg-slate-900 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-emerald-500/50 transition-colors">
                                <input type="checkbox" bind:checked={roleOptions.login} class="rounded text-emerald-500 focus:ring-emerald-500 bg-slate-800 border-slate-600" />
                                <span class="text-sm text-slate-700 dark:text-slate-300">Can Login</span>
                            </label>
                            <label class="flex items-center gap-2 p-2 rounded bg-slate-900 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-blue-500/50 transition-colors">
                                <input type="checkbox" bind:checked={roleOptions.createdb} class="rounded text-blue-500 focus:ring-blue-500 bg-slate-800 border-slate-600" />
                                <span class="text-sm text-slate-700 dark:text-slate-300">Create DB</span>
                            </label>
                            <label class="flex items-center gap-2 p-2 rounded bg-slate-900 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-blue-500/50 transition-colors">
                                <input type="checkbox" bind:checked={roleOptions.createrole} class="rounded text-blue-500 focus:ring-blue-500 bg-slate-800 border-slate-600" />
                                <span class="text-sm text-slate-700 dark:text-slate-300">Create Role</span>
                            </label>
                            <label class="flex items-center gap-2 p-2 rounded bg-slate-900 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-amber-500/50 transition-colors">
                                <input type="checkbox" bind:checked={roleOptions.superuser} class="rounded text-amber-500 focus:ring-amber-500 bg-slate-800 border-slate-600" />
                                <span class="text-sm text-amber-400 font-bold">Superuser</span>
                            </label>
                        </div>
                    </div>
                </div>

                <div class="flex justify-end pt-4 border-t border-slate-300 dark:border-slate-700">
                    <button onclick={createRole} class="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-slate-900 dark:text-white rounded-lg font-bold shadow-lg">Create Role</button>
                </div>
            </div>
        {/if}

        {#if loading}
            <div class="flex justify-center p-12">
                <div class="w-8 h-8 border-4 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each roles as role}
                    <div class="bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-all group">
                        <div class="flex justify-between items-start mb-3">
                            <div class="flex items-center gap-3">
                                <div class="p-2 rounded-lg {role.superuser ? 'bg-amber-500/10 text-amber-400' : 'bg-blue-500/10 text-blue-400'}">
                                    {#if role.superuser}
                                        <Key class="w-5 h-5" />
                                    {:else}
                                        <Users class="w-5 h-5" />
                                    {/if}
                                </div>
                                <div>
                                    <h3 class="font-bold text-slate-100 text-lg">{role.name}</h3>
                                    <div class="text-xs text-slate-500 font-mono">ID: {role.oid || 'N/A'}</div>
                                </div>
                            </div>
                            <button onclick={() => deleteRole(role.name)} class="p-2 text-slate-600 hover:text-red-400 hover:bg-red-500/10 rounded-lg opacity-0 group-hover:opacity-100 transition-all">
                                <Trash2 class="w-4 h-4" />
                            </button>
                        </div>
                        
                        <div class="flex flex-wrap gap-2 mt-4">
                            {#if role.superuser}
                                <span class="px-2 py-1 rounded bg-amber-500/10 text-amber-400 text-xs border border-amber-500/20 font-bold">Superuser</span>
                            {/if}
                            {#if role.can_login}
                                <span class="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 text-xs border border-emerald-500/20">Login</span>
                            {/if}
                            {#if role.create_db}
                                <span class="px-2 py-1 rounded bg-blue-500/10 text-blue-400 text-xs border border-blue-500/20">Create DB</span>
                            {/if}
                            {#if role.create_role}
                                <span class="px-2 py-1 rounded bg-purple-500/10 text-purple-400 text-xs border border-purple-500/20">Create Role</span>
                            {/if}
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
