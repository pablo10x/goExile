<script lang="ts">
	import { Users, UserPlus, Trash2, Key, Shield, Check, X } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';
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

<div class="h-full flex flex-col bg-transparent">
	<div
		class="p-6 border-b border-neutral-800 bg-neutral-900/40 backdrop-blur-md flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-indigo-500/10 border border-indigo-500/20 rounded-xl">
				<Shield class="w-6 h-6 text-indigo-400" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-white uppercase tracking-tighter italic">SEC_IDENTITY_MANAGER</h2>
				<p class="font-jetbrains text-[10px] text-neutral-500 uppercase tracking-widest mt-1 italic font-bold">Manage database users and privileges</p>
			</div>
		</div>
		<button
			onclick={() => (isCreating = !isCreating)}
			class="px-6 py-2.5 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 transition-all flex items-center gap-2 rounded-xl"
		>
			{#if isCreating}
				<X class="w-4 h-4" /> Cancel_Op
			{:else}
				<UserPlus class="w-4 h-4" /> Create_Identity
			{/if}
		</button>
	</div>

	<div class="flex-1 overflow-auto p-8 relative custom-scrollbar bg-transparent">
		{#if isCreating}
			<div
				class="mb-10 bg-neutral-900/60 border border-neutral-800 p-8 shadow-2xl rounded-2xl backdrop-blur-md"
			>
				<h3
					class="font-heading font-black text-xs text-indigo-400 mb-6 flex items-center gap-3 uppercase tracking-[0.2em]"
				>
					<UserPlus class="w-5 h-5" /> New_Identity_Definition
				</h3>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
					<div class="space-y-6">
						<div>
							<label
								for="newRoleName"
								class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest mb-2"
								>Subject_ID</label
							>
							<input
								id="newRoleName"
								type="text"
								bind:value={newRoleName}
								class="w-full bg-neutral-950/40 border border-neutral-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-indigo-500 outline-none uppercase tracking-widest transition-all rounded-lg"
								placeholder="e.g. ALPHA_USER"
							/>
						</div>
						<div>
							<label
								for="newRolePass"
								class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest mb-2"
								>Access_Key</label
							>
							<input
								id="newRolePass"
								type="password"
								bind:value={newRolePass}
								class="w-full bg-neutral-950/40 border border-neutral-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all rounded-lg"
								placeholder="••••••••"
							/>
						</div>
					</div>

					<div class="space-y-4">
						<div class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest mb-2 italic">
							Auth_Permissions
						</div>
						<div class="grid grid-cols-2 gap-3">
							<label
								class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-indigo-500/40 transition-colors group rounded-xl"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.login}
									class="w-4 h-4 bg-neutral-900 border-neutral-700 text-indigo-500 focus:ring-indigo-500"
								/>
								<span class="font-heading text-[10px] font-bold text-neutral-500 uppercase group-hover:text-neutral-200 tracking-widest">Can_Login</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-indigo-500/40 transition-colors group rounded-xl"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.createdb}
									class="w-4 h-4 bg-neutral-900 border-neutral-700 text-indigo-500 focus:ring-indigo-500"
								/>
								<span class="font-heading text-[10px] font-bold text-neutral-500 uppercase group-hover:text-neutral-200 tracking-widest">Create_DB</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-indigo-500/40 transition-colors group rounded-xl"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.createrole}
									class="w-4 h-4 bg-neutral-900 border-neutral-700 text-indigo-500 focus:ring-indigo-500"
								/>
								<span class="font-heading text-[10px] font-bold text-neutral-500 uppercase group-hover:text-neutral-200 tracking-widest">Create_Role</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-amber-500/40 transition-colors group rounded-xl"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.superuser}
									class="w-4 h-4 bg-neutral-900 border-neutral-700 text-amber-500 focus:ring-amber-500"
								/>
								<span class="font-heading text-[10px] font-bold text-amber-500 uppercase tracking-widest">Superuser</span>
							</label>
						</div>
					</div>
				</div>

				<div class="flex justify-end pt-6 border-t border-neutral-800">
					<button
						onclick={createRole}
						class="px-8 py-3 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 transition-all active:tranneutral-y-px rounded-xl"
						>Authorize_Identity</button
					>
				</div>
			</div>
		{/if}

		{#if loading}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(99,102,241,0.4)]"
				></div>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each roles as role}
					<div
						class="bg-neutral-900/40 border border-neutral-800 p-6 hover:border-indigo-500/40 transition-all group rounded-2xl backdrop-blur-sm shadow-lg"
					>
						<div class="flex justify-between items-start mb-6">
							<div class="flex items-center gap-4">
								<div
									class="p-2.5 rounded-xl {role.superuser
										? 'bg-amber-500/10 text-amber-500 border border-amber-500/30'
										: 'bg-neutral-950 text-neutral-600 border border-neutral-800'}"
								>
									{#if role.superuser}
										<Key class="w-5 h-5" />
									{:else}
										<Users class="w-5 h-5" />
									{/if}
								</div>
								<div>
									<h3 class="font-heading font-black text-white text-base tracking-tight uppercase italic">{role.name}</h3>
									<div class="font-jetbrains text-[9px] text-neutral-500 font-bold uppercase tracking-widest">ID: {role.oid || 'N/A'}</div>
								</div>
							</div>
							<button
								onclick={() => deleteRole(role.name)}
								class="p-2 text-neutral-600 hover:text-red-400 hover:bg-red-500/10 transition-all opacity-0 group-hover:opacity-100 rounded-lg"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</div>

						<div class="flex flex-wrap gap-2 mt-6">
							{#if role.superuser}
								<span
									class="px-2 py-1 bg-amber-500/10 text-amber-500 text-[9px] border border-amber-500/20 font-bold font-jetbrains uppercase tracking-widest rounded"
									>Superuser</span
								>
							{/if}
							{#if role.can_login}
								<span
									class="px-2 py-1 bg-emerald-500/10 text-emerald-400 text-[9px] border border-emerald-500/20 font-bold font-jetbrains uppercase tracking-widest rounded"
									>Login</span
								>
							{/if}
							{#if role.create_db}
								<span
									class="px-2 py-1 bg-indigo-500/10 text-indigo-400 text-[9px] border border-indigo-500/20 font-bold font-jetbrains uppercase tracking-widest rounded"
									>Create DB</span
								>
							{/if}
							{#if role.create_role}
								<span
									class="px-2 py-1 bg-orange-500/10 text-orange-400 text-[9px] border border-orange-500/20 font-bold font-jetbrains uppercase tracking-widest rounded"
									>Create Role</span
								>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
