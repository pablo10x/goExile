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

<div class="h-full flex flex-col bg-[var(--terminal-bg)]">
	<div
		class="p-6 border-b border-stone-800 bg-[var(--header-bg)] flex justify-between items-center"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
				<Shield class="w-6 h-6 text-rust-light" />
			</div>
			<div>
				<h2 class="text-xl font-heading font-black text-slate-100 uppercase tracking-tighter">SEC_IDENTITY_MANAGER</h2>
				<p class="font-jetbrains text-[10px] text-text-dim uppercase tracking-widest mt-1">Manage database users and privileges</p>
			</div>
		</div>
		<button
			onclick={() => (isCreating = !isCreating)}
			class="px-6 py-2.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all flex items-center gap-2"
		>
			{#if isCreating}
				<X class="w-4 h-4" /> Cancel_Op
			{:else}
				<UserPlus class="w-4 h-4" /> Create_Identity
			{/if}
		</button>
	</div>

	<div class="flex-1 overflow-auto p-8 relative custom-scrollbar">
		{#if isCreating}
			<div
				class="mb-10 bg-stone-900/40 border border-stone-800 p-8 shadow-2xl industrial-frame"
			>
				<h3
					class="font-heading font-black text-xs text-rust-light mb-6 flex items-center gap-3 uppercase tracking-[0.2em]"
				>
					<UserPlus class="w-5 h-5" /> New_Identity_Definition
				</h3>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
					<div class="space-y-6">
						<div>
							<label
								for="newRoleName"
								class="block font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest mb-2"
								>Subject_ID</label
							>
							<input
								id="newRoleName"
								type="text"
								bind:value={newRoleName}
								class="w-full bg-stone-950 border border-stone-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none uppercase tracking-widest"
								placeholder="e.g. ALPHA_USER"
							/>
						</div>
						<div>
							<label
								for="newRolePass"
								class="block font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest mb-2"
								>Access_Key</label
							>
							<input
								id="newRolePass"
								type="password"
								bind:value={newRolePass}
								class="w-full bg-stone-950 border border-stone-800 py-3 px-4 text-stone-200 font-jetbrains text-xs focus:border-rust outline-none"
								placeholder="••••••••"
							/>
						</div>
					</div>

					<div class="space-y-4">
						<div class="block font-jetbrains text-[10px] font-black text-text-dim uppercase tracking-widest mb-2">
							Auth_Permissions
						</div>
						<div class="grid grid-cols-2 gap-3">
							<label
								class="flex items-center gap-3 p-3 bg-stone-950 border border-stone-800 cursor-pointer hover:border-rust/40 transition-colors group"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.login}
									class="w-4 h-4 bg-stone-900 border-stone-700 text-rust focus:ring-rust"
								/>
								<span class="font-heading text-[10px] font-bold text-stone-400 uppercase group-hover:text-stone-200">Can_Login</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-stone-950 border border-stone-800 cursor-pointer hover:border-rust/40 transition-colors group"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.createdb}
									class="w-4 h-4 bg-stone-900 border-stone-700 text-rust focus:ring-rust"
								/>
								<span class="font-heading text-[10px] font-bold text-stone-400 uppercase group-hover:text-stone-200">Create_DB</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-stone-950 border border-stone-800 cursor-pointer hover:border-rust/40 transition-colors group"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.createrole}
									class="w-4 h-4 bg-stone-900 border-stone-700 text-rust focus:ring-rust"
								/>
								<span class="font-heading text-[10px] font-bold text-stone-400 uppercase group-hover:text-stone-200">Create_Role</span>
							</label>
							<label
								class="flex items-center gap-3 p-3 bg-stone-950 border border-stone-800 cursor-pointer hover:border-amber-500/40 transition-colors group"
							>
								<input
									type="checkbox"
									bind:checked={roleOptions.superuser}
									class="w-4 h-4 bg-stone-900 border-stone-700 text-warning focus:ring-amber-500"
								/>
								<span class="font-heading text-[10px] font-bold text-warning uppercase">Superuser</span>
							</label>
						</div>
					</div>
				</div>

				<div class="flex justify-end pt-6 border-t border-stone-800">
					<button
						onclick={createRole}
						class="px-8 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all"
						>Authorize_Identity</button
					>
				</div>
			</div>
		{/if}

		{#if loading}
			<div class="flex justify-center p-20">
				<div
					class="w-10 h-10 border-2 border-rust border-t-transparent rounded-none animate-spin"
				></div>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{#each roles as role}
					<div
						class="bg-stone-900/40 border border-stone-800 p-6 hover:border-rust/40 transition-all group industrial-frame"
					>
						<div class="flex justify-between items-start mb-6">
							<div class="flex items-center gap-4">
								<div
									class="p-2.5 rounded-none {role.superuser
										? 'bg-amber-500/10 text-warning border border-amber-500/30'
										: 'bg-stone-800 text-stone-400 border border-stone-700'}"
								>
									{#if role.superuser}
										<Key class="w-5 h-5" />
									{:else}
										<Users class="w-5 h-5" />
									{/if}
								</div>
								<div>
									<h3 class="font-heading font-black text-stone-100 text-base tracking-tight uppercase">{role.name}</h3>
									<div class="font-jetbrains text-[9px] text-text-dim font-bold uppercase tracking-widest">ID: {role.oid || 'N/A'}</div>
								</div>
							</div>
							<button
								onclick={() => deleteRole(role.name)}
								class="p-2 text-stone-700 hover:text-danger hover:bg-red-500/10 transition-all opacity-0 group-hover:opacity-100"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</div>

						<div class="flex flex-wrap gap-2 mt-6">
							{#if role.superuser}
								<span
									class="px-2 py-1 bg-amber-500/10 text-warning text-[9px] border border-amber-500/20 font-black font-jetbrains uppercase tracking-widest"
									>Superuser</span
								>
							{/if}
							{#if role.can_login}
								<span
									class="px-2 py-1 bg-success/10 text-success text-[9px] border border-emerald-500/20 font-black font-jetbrains uppercase tracking-widest"
									>Login</span
								>
							{/if}
							{#if role.create_db}
								<span
									class="px-2 py-1 bg-rust/10 text-rust text-[9px] border border-rust/20 font-black font-jetbrains uppercase tracking-widest"
									>Create DB</span
								>
							{/if}
							{#if role.create_role}
								<span
									class="px-2 py-1 bg-purple-500/10 text-purple-500 text-[9px] border border-purple-500/20 font-black font-jetbrains uppercase tracking-widest"
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
