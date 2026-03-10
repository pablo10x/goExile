<script lang="ts">
	import { apiFetch } from '$lib/api';
	import {
		Code2,
		Plus,
		Trash2,
		Play,
		Edit3,
		X,
		Search,
		RefreshCw,
		Copy,
		Check,
		Zap,
		FileCode,
		ChevronDown,
		ChevronRight,
		Settings2,
		Shield,
		Database,
		Info,
		AlertOctagon
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores.svelte';
	import { fade, slide } from 'svelte/transition';
	import { portal } from '$lib/actions';

	interface PGFunction {
		oid: number;
		schema: string;
		name: string;
		result_type: string;
		argument_types: string;
		type: string;
		volatility: string;
		language: string;
		source: string;
		owner: string;
		description: string;
	}

	let functions = $state<PGFunction[]>([]);
	let filteredFunctions = $state<PGFunction[]>([]);
	let loading = $state(false);
	let searchQuery = $state('');
	let selectedSchema = $state('public');
	let schemas = $state<string[]>(['public']);

	// Editor state
	let isEditing = $state(false);
	let isCreating = $state(false);
	let selectedFunction = $state<PGFunction | null>(null);
	let expandedFunctions = $state<Set<number>>(new Set());

	// Form state
	let formData = $state({
		schema: 'public',
		name: '',
		arguments: '',
		returns: 'void',
		language: 'plpgsql',
		body: '',
		volatility: 'VOLATILE',
		isStrict: false,
		securityDefiner: false
	});

	// Execute function state
	let executeModalOpen = $state(false);
	let executeArgs = $state<string[]>([]);
	let executeResult = $state<any[]>([]);
	let executeLoading = $state(false);
	let functionToExecute = $state<PGFunction | null>(null);

	let copied = $state(false);

	const languages = ['plpgsql', 'sql', 'plpython3u', 'plperl', 'plv8'];
	const volatilities = ['VOLATILE', 'STABLE', 'IMMUTABLE'];
	const returnTypes = [
		'void',
		'boolean',
		'integer',
		'bigint',
		'numeric',
		'text',
		'varchar',
		'json',
		'jsonb',
		'timestamp',
		'date',
		'uuid',
		'SETOF record',
		'TABLE'
	];

	async function loadSchemas() {
		try {
			const res = await apiFetch('/api/database/schemas');
			if (res.ok) {
				const data = await res.json();
				schemas = data.map((s: any) => s.name || s.schema_name || s);
			}
		} catch (e) {
			console.error('Failed to load schemas', e);
		}
	}

	async function loadFunctions() {
		loading = true;
		try {
			const res = await apiFetch(`/api/database/functions?schema=${selectedSchema}`, {
				credentials: 'include'
			});

			const text = await res.text();

			if (!res.ok) {
				try {
					const err = JSON.parse(text);
					throw new Error(err.error || `Failed to load functions (${res.status})`);
				} catch {
					throw new Error(text || `Failed to load functions (${res.status})`);
				}
			}

			try {
				functions = JSON.parse(text);
				if (!Array.isArray(functions)) {
					functions = [];
				}
				filterFunctions();
			} catch (parseError) {
				console.error('Failed to parse functions response:', text);
				throw new Error('Invalid response from server');
			}
		} catch (e: any) {
			console.error('Load functions error:', e);
			notifications.add({ type: 'error', message: 'Failed to load functions', details: e.message });
			functions = [];
			filteredFunctions = [];
		} finally {
			loading = false;
		}
	}

	function filterFunctions() {
		if (!searchQuery.trim()) {
			filteredFunctions = functions;
		} else {
			const query = searchQuery.toLowerCase();
			filteredFunctions = functions.filter(
				(fn) =>
					fn.name.toLowerCase().includes(query) ||
					fn.argument_types.toLowerCase().includes(query) ||
					fn.result_type.toLowerCase().includes(query)
			);
		}
	}

	function openCreateModal() {
		formData = {
			schema: selectedSchema,
			name: '',
			arguments: '',
			returns: 'void',
			language: 'plpgsql',
			body: 'BEGIN\n  -- Your code here\n  RETURN;\nEND;',
			volatility: 'VOLATILE',
			isStrict: false,
			securityDefiner: false
		};
		isCreating = true;
		isEditing = false;
		selectedFunction = null;
	}

	function openEditModal(fn: PGFunction) {
		selectedFunction = fn;
		let body = fn.source;
		const bodyMatch = fn.source.match(/\$\$\s*([\s\S]*?)\s*\$\$/);
		if (bodyMatch) {
			body = bodyMatch[1].trim();
		}

		formData = {
			schema: fn.schema,
			name: fn.name,
			arguments: fn.argument_types,
			returns: fn.result_type,
			language: fn.language,
			body: body,
			volatility: fn.volatility.toUpperCase(),
			isStrict: false,
			securityDefiner: false
		};
		isEditing = true;
		isCreating = false;
	}

	function closeModal() {
		isCreating = false;
		isEditing = false;
		selectedFunction = null;
	}

	async function saveFunction() {
		if (!formData.name.trim() || !formData.body.trim()) {
			notifications.add({ type: 'error', message: 'Name and body are required' });
			return;
		}

		loading = true;
		try {
			const method = isCreating ? 'POST' : 'PUT';
			const res = await apiFetch('/api/database/functions', {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					oid: selectedFunction?.oid,
					schema: formData.schema,
					name: formData.name,
					arguments: formData.arguments,
					returns: formData.returns,
					language: formData.language,
					body: formData.body,
					volatility: formData.volatility,
					is_strict: formData.isStrict,
					security_definer: formData.securityDefiner
				})
			});

			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Failed to save function');
			}

			notifications.add({
				type: 'success',
				message: isCreating ? 'Function created successfully' : 'Function updated successfully'
			});
			closeModal();
			loadFunctions();
		} catch (e: any) {
			notifications.add({ type: 'error', message: 'Failed to save function', details: e.message });
		} finally {
			loading = false;
		}
	}

	async function deleteFunction(fn: PGFunction) {
		if (!confirm(`Are you sure you want to delete function "${fn.name}"?`)) {
			return;
		}

		loading = true;
		try {
			const res = await apiFetch('/api/database/functions/delete', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					schema: fn.schema,
					name: fn.name,
					arguments: fn.argument_types,
					cascade: false
				})
			});

			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Failed to delete function');
			}

			notifications.add({ type: 'success', message: `Function "${fn.name}" deleted` });
			loadFunctions();
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Failed to delete function',
				details: e.message
			});
		} finally {
			loading = false;
		}
	}

	function openExecuteModal(fn: PGFunction) {
		functionToExecute = fn;
		executeArgs = fn.argument_types ? fn.argument_types.split(',').map(() => '') : [];
		executeResult = [];
		executeModalOpen = true;
	}

	async function executeFunction() {
		if (!functionToExecute) return;

		executeLoading = true;
		try {
			const res = await apiFetch('/api/database/functions/execute', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					schema: functionToExecute.schema,
					name: functionToExecute.name,
					arguments: executeArgs.filter((a) => a.trim() !== '')
				})
			});

			if (!res.ok) {
				const err = await res.json();
				throw new Error(err.error || 'Failed to execute function');
			}

			executeResult = await res.json();
			notifications.add({ type: 'success', message: 'Function executed successfully' });
		} catch (e: any) {
			notifications.add({
				type: 'error',
				message: 'Function execution failed',
				details: e.message
			});
		} finally {
			executeLoading = false;
		}
	}

	function toggleExpand(oid: number) {
		if (expandedFunctions.has(oid)) {
			expandedFunctions.delete(oid);
		} else {
			expandedFunctions.add(oid);
		}
		expandedFunctions = new Set(expandedFunctions);
	}

	async function copySource(source: string) {
		await navigator.clipboard.writeText(source);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function getTypeColor(type: string): string {
		switch (type) {
			case 'function':
				return 'text-sky-400 bg-sky-500/10 border-sky-500/30';
			case 'procedure':
				return 'text-purple-400 bg-purple-500/10 border-purple-500/30';
			case 'aggregate':
				return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
			case 'window':
				return 'text-cyan-400 bg-cyan-500/10 border-cyan-500/30';
			default:
				return 'text-slate-500 bg-neutral-500/10 border-neutral-500/30';
		}
	}

	function getVolatilityColor(vol: string): string {
		switch (vol.toLowerCase()) {
			case 'immutable':
				return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
			case 'stable':
				return 'text-sky-400 bg-sky-500/10 border-sky-500/30';
			case 'volatile':
				return 'text-orange-400 bg-orange-500/10 border-orange-500/30';
			default:
				return 'text-slate-500 bg-neutral-500/10 border-neutral-500/30';
		}
	}

	$effect(() => {
		filterFunctions();
	});

	onMount(() => {
		loadSchemas();
		loadFunctions();
	});
</script>

<div class="h-full flex flex-col bg-transparent font-sans">
	<!-- Header -->
	<div class="p-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md">
		<div class="flex justify-between items-start">
			<div class="flex items-center gap-4">
				<div
					class="p-3 bg-sky-500/10 rounded-xl border border-sky-500/20 shadow-lg shadow-sky-900/10"
				>
					<Code2 class="w-7 h-7 text-sky-400" />
				</div>
				<div>
					<h2 class="text-2xl font-bold text-white uppercase tracking-tight italic">
						Database Functions
					</h2>
					<p
						class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-1"
					>
						Manage stored procedures, functions and triggers
					</p>
				</div>
			</div>
			<button
				onclick={openCreateModal}
				class="relative z-10 px-6 py-3 bg-sky-500 hover:bg-sky-400 text-white font-bold text-[11px] uppercase tracking-widest shadow-lg shadow-sky-900/20 transition-all rounded-xl"
			>
				<Plus class="w-5 h-5" />
				New Function
			</button>
		</div>

		<!-- Filters -->
		<div class="flex gap-4 mt-8">
			<div class="relative flex-1 max-w-md">
				<Search
					class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-600 pointer-events-none"
				/>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search functions by name or type..."
					class="w-full pl-12 pr-4 py-3 bg-black/40 border border-white/10 text-slate-200 text-[11px] placeholder-slate-800 focus:border-sky-500 outline-none transition-all uppercase tracking-widest rounded-xl"
				/>
			</div>

			<select
				bind:value={selectedSchema}
				onchange={() => loadFunctions()}
				class="px-4 py-3 bg-black/40 border border-white/10 text-slate-400 text-[11px] focus:border-sky-500 outline-none cursor-pointer uppercase tracking-widest rounded-xl appearance-none"
			>
				{#each schemas as schema}
					<option value={schema}>{schema}</option>
				{/each}
			</select>

			<button
				onclick={() => loadFunctions()}
				class="p-3 bg-slate-900/40 border border-white/10 text-slate-500 hover:text-sky-400 hover:border-sky-500/30 transition-all rounded-xl"
				title="Refresh"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Functions List -->
	<div class="flex-1 overflow-auto p-8 custom-scrollbar bg-transparent">
		{#if loading && functions.length === 0}
			<div class="flex flex-col items-center justify-center h-64">
				<div
					class="w-10 h-10 border-2 border-sky-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(14,165,233,0.4)]"
				></div>
				<p
					class="mt-6 font-bold text-[11px] text-slate-600 uppercase tracking-widest animate-pulse"
				>
					Loading Functions...
				</p>
			</div>
		{:else if filteredFunctions.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-slate-700">
				<FileCode class="w-16 h-16 opacity-10 mb-6" />
				<p class="font-bold text-xs tracking-widest uppercase">
					No functions found
				</p>
				<p class="text-[9px] mt-2 uppercase font-bold opacity-40 tracking-widest">
					{searchQuery
						? 'Try adjusting your filters'
						: 'Create your first function to begin'}
				</p>
			</div>
		{:else}
			<div class="space-y-4 max-w-6xl mx-auto">
				{#each filteredFunctions as fn (fn.oid)}
					<div
						class="bg-slate-900/40 border border-white/5 overflow-hidden hover:border-sky-500/40 transition-all group rounded-2xl backdrop-blur-sm shadow-lg"
						transition:fade={{ duration: 150 }}
					>
						<!-- Function Header -->
						<div class="p-5">
							<div class="flex items-start justify-between">
								<div class="flex items-start gap-4 flex-1 min-w-0">
									<button
										onclick={() => toggleExpand(fn.oid)}
										class="p-1 mt-1 text-slate-600 hover:text-sky-400 transition-colors"
									>
										{#if expandedFunctions.has(fn.oid)}
											<ChevronDown class="w-5 h-5" />
										{:else}
											<ChevronRight class="w-5 h-5" />
										{/if}
									</button>

									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-4 flex-wrap">
											<h3
												class="font-bold text-white text-lg tracking-tight uppercase italic"
											>
												{fn.name}
											</h3>
											<span
												class="px-2 py-0.5 text-[9px] font-bold border uppercase tracking-widest rounded {getTypeColor(
													fn.type
												)}"
											>
												{fn.type}
											</span>
											<span
												class="px-2 py-0.5 text-[9px] font-bold border uppercase tracking-widest rounded {getVolatilityColor(
													fn.volatility
												)}"
											>
												{fn.volatility}
											</span>
											<span
												class="px-2 py-0.5 text-[9px] font-bold bg-black text-slate-600 border border-white/5 uppercase tracking-widest italic rounded"
											>
												{fn.language}
											</span>
										</div>

										<div class="mt-3 text-xs font-bold uppercase tracking-tight">
											<span class="text-sky-400">{fn.name}</span>
											<span class="text-slate-700">(</span>
											<span class="text-sky-500/70">{fn.argument_types || ''}</span>
											<span class="text-slate-700">)</span>
											<span class="text-slate-700 mx-3">::</span>
											<span class="text-emerald-500">{fn.result_type}</span>
										</div>

										{#if fn.description}
											<p
												class="mt-3 text-[10px] text-slate-500 uppercase tracking-tight font-bold"
											>
												{fn.description}
											</p>
										{/if}
									</div>
								</div>

								<!-- Actions -->
								<div
									class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
								>
									{#if fn.type === 'function'}
										<button
											onclick={() => openExecuteModal(fn)}
											class="p-2 text-slate-600 hover:text-emerald-400 hover:bg-emerald-500/10 transition-all rounded-lg"
											title="Execute"
										>
											<Play class="w-4 h-4" />
										</button>
									{/if}
									<button
										onclick={() => openEditModal(fn)}
										class="p-2 text-slate-600 hover:text-sky-400 hover:bg-sky-500/10 transition-all rounded-lg"
										title="Edit"
									>
										<Edit3 class="w-4 h-4" />
									</button>
									<button
										onclick={() => deleteFunction(fn)}
										class="p-2 text-slate-600 hover:text-red-400 hover:bg-red-500/10 transition-all rounded-lg"
										title="Delete"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								</div>
							</div>
						</div>

						<!-- Expanded Source Code -->
						{#if expandedFunctions.has(fn.oid)}
							<div
								class="border-t border-white/5 bg-black/20"
								transition:slide={{ duration: 200 }}
							>
								<div class="p-6">
									<div class="flex items-center justify-between mb-4">
										<span
											class="text-[10px] font-bold text-slate-600 uppercase tracking-widest italic"
											>Source Code</span
										>
										<button
											onclick={() => copySource(fn.source)}
											class="flex items-center gap-3 px-4 py-1.5 text-[9px] font-bold text-slate-500 hover:text-sky-400 hover:bg-sky-500/5 border border-white/5 transition-all uppercase tracking-widest rounded-lg"
										>
											{#if copied}
												<Check class="w-3 h-3" />
												Copied
											{:else}
												<Copy class="w-3 h-3" />
												Copy Source
											{/if}
										</button>
									</div>
									<pre
										class="p-6 bg-black border border-white/5 text-[11px] text-slate-400 font-mono overflow-x-auto max-h-[500px] custom-scrollbar shadow-inner rounded-xl"><code>{fn.source}</code></pre>

									<div
										class="flex items-center gap-8 mt-6 text-[9px] font-bold text-slate-600 uppercase tracking-widest italic"
									>
										<span class="flex items-center gap-2">
											<Shield class="w-3.5 h-3.5" />
											Owner: {fn.owner}
										</span>
										<span class="flex items-center gap-2">
											<Database class="w-3.5 h-3.5" />
											OID: {fn.oid}
										</span>
									</div>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Create/Edit Modal -->
{#if isCreating || isEditing}
	<div
		use:portal
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		transition:fade={{ duration: 150 }}
	>
		<button
			class="absolute inset-0 bg-black/60 backdrop-blur-md cursor-default"
			onclick={closeModal}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-4xl max-h-[90vh] bg-slate-900/90 backdrop-blur-2xl border border-white/10 rounded-3xl shadow-2xl overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div class="p-8 border-b border-white/5 bg-black/40">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-5">
						<div class="p-3 bg-sky-500/10 border border-sky-500/20 rounded-2xl shadow-lg">
							{#if isCreating}
								<Plus class="w-6 h-6 text-sky-400" />
							{:else}
								<Edit3 class="w-6 h-6 text-sky-400" />
							{/if}
						</div>
						<div>
							<h3
								class="text-2xl font-bold text-white uppercase tracking-tight italic"
							>
								{isCreating ? 'New Function' : 'Edit Function'}
							</h3>
							<p
								class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-1"
							>
								{isCreating
									? 'Define a new database function'
									: `Modifying ${selectedFunction?.name}`}
							</p>
						</div>
					</div>
					<button
						onclick={closeModal}
						class="p-2 text-slate-500 hover:text-white transition-all rounded-lg"
					>
						<X class="w-6 h-6" />
					</button>
				</div>
			</div>

			<!-- Modal Body -->
			<div class="flex-1 overflow-y-auto p-10 space-y-10 custom-scrollbar bg-transparent">
				<!-- Basic Info -->
				<div class="grid grid-cols-2 gap-8">
					<div class="space-y-3">
						<label
							for="fnSchema"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Database Schema</label
						>
						<select
							id="fnSchema"
							bind:value={formData.schema}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-300 text-xs focus:border-sky-500 outline-none transition-all uppercase appearance-none rounded-xl"
						>
							{#each schemas as schema}
								<option value={schema}>{schema}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-3">
						<label
							for="fnName"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Function Name</label
						>
						<input
							id="fnName"
							type="text"
							bind:value={formData.name}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-200 text-xs focus:border-sky-500 outline-none transition-all uppercase rounded-xl"
							placeholder="function_name"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-8">
					<div class="space-y-3">
						<label
							for="fnArgs"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Arguments</label
						>
						<input
							id="fnArgs"
							type="text"
							bind:value={formData.arguments}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-200 text-xs focus:border-sky-500 outline-none transition-all rounded-xl"
							placeholder="arg1 integer, arg2 text"
						/>
					</div>
					<div class="space-y-3">
						<label
							for="fnReturns"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Return Type</label
						>
						<select
							id="fnReturns"
							bind:value={formData.returns}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-300 text-xs focus:border-sky-500 outline-none transition-all appearance-none rounded-xl"
						>
							{#each returnTypes as type}
								<option value={type}>{type}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="grid grid-cols-3 gap-8">
					<div class="space-y-3">
						<label
							for="fnLang"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Language</label
						>
						<select
							id="fnLang"
							bind:value={formData.language}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-300 text-xs focus:border-sky-500 outline-none transition-all appearance-none rounded-xl"
						>
							{#each languages as lang}
								<option value={lang}>{lang}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-3">
						<label
							for="fnVol"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
							>Volatility</label
						>
						<select
							id="fnVol"
							bind:value={formData.volatility}
							class="w-full px-4 py-3 bg-black/40 border border-white/10 text-slate-300 text-xs focus:border-sky-500 outline-none transition-all appearance-none rounded-xl"
						>
							{#each volatilities as vol}
								<option value={vol}>{vol}</option>
							{/each}
						</select>
					</div>
					<div class="flex flex-col justify-end gap-3">
						<label
							class="flex items-center gap-3 p-3 bg-black/40 border border-white/10 cursor-pointer hover:border-sky-500/50 transition-colors rounded-xl"
						>
							<input
								type="checkbox"
								bind:checked={formData.isStrict}
								class="rounded text-sky-500 focus:ring-sky-500 bg-slate-900 border-white/10"
							/>
							<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
								>Strict Check</span
							>
						</label>
						<label
							class="flex items-center gap-3 p-3 bg-black/40 border border-white/10 cursor-pointer hover:border-sky-500/50 transition-colors rounded-xl"
						>
							<input
								type="checkbox"
								bind:checked={formData.securityDefiner}
								class="rounded text-sky-500 focus:ring-sky-500 bg-slate-900 border-white/10"
							/>
							<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
								>Security Definer</span
							>
						</label>
					</div>
				</div>

				<!-- Function Body -->
				<div class="space-y-3">
					<label
						for="fnBody"
						class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest italic"
						>Execution Logic</label
					>
					<div class="relative">
						<div
							class="absolute top-4 left-4 text-sky-500 font-bold text-xs opacity-20 pointer-events-none tracking-widest"
						>
							BEGIN
						</div>
						<textarea
							id="fnBody"
							bind:value={formData.body}
							rows="15"
							class="w-full px-10 py-8 bg-black/60 border border-white/10 text-slate-200 focus:border-sky-500 outline-none font-mono text-sm resize-none leading-relaxed shadow-inner rounded-2xl"
							placeholder="-- Your code here"
						></textarea>
						<div
							class="absolute bottom-4 right-4 text-sky-500 font-bold text-xs opacity-20 pointer-events-none tracking-widest"
						>
							END;
						</div>
					</div>
				</div>
			</div>

			<!-- Modal Footer -->
			<div
				class="p-8 border-t border-white/5 bg-black/40 flex justify-end items-center gap-6"
			>
				<button
					onclick={closeModal}
					class="px-8 py-3 text-[11px] font-bold text-slate-500 hover:text-white uppercase tracking-widest transition-all rounded-lg"
				>
					Cancel
				</button>
				<button
					onclick={saveFunction}
					disabled={loading}
					class="px-10 py-4 bg-sky-500 hover:bg-sky-400 text-white font-bold text-[11px] uppercase tracking-widest shadow-lg shadow-sky-900/20 transition-all disabled:opacity-20 active:translate-y-px rounded-xl"
				>
					{#if loading}
						<RefreshCw class="w-4 h-4 animate-spin inline mr-3" />
						Saving...
					{:else}
						<Check class="w-4 h-4 inline mr-3" />
						Save Function
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Execute Function Modal -->
{#if executeModalOpen && functionToExecute}
	<div
		use:portal
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		transition:fade={{ duration: 150 }}
	>
		<button
			class="absolute inset-0 bg-black/60 backdrop-blur-sm cursor-default"
			onclick={() => (executeModalOpen = false)}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-2xl max-h-[80vh] bg-slate-900/90 backdrop-blur-2xl border border-white/10 rounded-3xl shadow-2xl overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div class="p-8 border-b border-white/5 bg-black/40">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-5">
						<div class="p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-2xl shadow-lg">
							<Play class="w-6 h-6 text-emerald-400" />
						</div>
						<div>
							<h3
								class="text-2xl font-bold text-white uppercase tracking-tight italic"
							>
								Execute Function
							</h3>
							<p
								class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-1 italic"
							>
								{functionToExecute.name}()
							</p>
						</div>
					</div>
					<button
						onclick={() => (executeModalOpen = false)}
						class="p-2 text-slate-500 hover:text-white transition-all rounded-lg"
					>
						<X class="w-6 h-6" />
					</button>
				</div>
			</div>

			<!-- Modal Body -->
			<div class="flex-1 overflow-y-auto p-10 space-y-10 custom-scrollbar bg-transparent">
				{#if functionToExecute.argument_types}
					<div role="group" aria-labelledby="args-label">
						<div
							id="args-label"
							class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-6 italic border-l border-white/10 pl-4"
						>
							Input Arguments
						</div>
						<div class="space-y-4">
							{#each functionToExecute.argument_types.split(',') as arg, i}
								<div class="flex items-center gap-6 group">
									<label
										for={`execArg-${i}`}
										class="text-[10px] font-bold text-slate-500 uppercase tracking-widest min-w-[140px] italic group-hover:text-sky-400 transition-colors"
										>{arg.trim()}</label
									>
									<input
										id={`execArg-${i}`}
										type="text"
										bind:value={executeArgs[i]}
										class="flex-1 px-4 py-3 bg-black/40 border border-white/10 text-slate-200 focus:border-sky-500 outline-none font-mono text-xs transition-all shadow-inner rounded-xl"
										placeholder="Value"
									/>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center py-10 opacity-40">
						<Info class="w-10 h-10 text-slate-700 mb-4" />
						<p
							class="font-bold text-[10px] text-slate-500 uppercase tracking-widest"
						>
							This function requires no input arguments.
						</p>
					</div>
				{/if}

				{#if executeResult.length > 0}
					<div role="group" aria-labelledby="result-label" transition:slide>
						<div
							id="result-label"
							class="block text-[10px] font-bold text-emerald-500 uppercase tracking-widest mb-6 italic border-l border-emerald-900/50 pl-4"
						>
							Results
						</div>
						<div
							class="bg-black/40 border border-white/10 rounded-2xl overflow-hidden shadow-inner"
						>
							<div class="overflow-x-auto custom-scrollbar">
								<table class="w-full text-xs">
									<thead class="bg-black text-slate-500 border-b border-white/10">
										<tr>
											{#each Object.keys(executeResult[0]) as key}
												<th
													class="px-6 py-4 text-left font-bold uppercase tracking-widest border-r border-white/5 italic"
													>{key}</th
												>
											{/each}
										</tr>
									</thead>
									<tbody class="divide-y divide-white/5">
										{#each executeResult as row}
											<tr class="hover:bg-sky-500/5 transition-colors">
												{#each Object.values(row) as val}
													<td class="px-6 py-4 text-slate-400 font-medium tracking-tight">
														{val === null ? 'NULL' : String(val)}
													</td>
												{/each}
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Modal Footer -->
			<div class="p-8 border-t border-white/5 bg-black/40 flex justify-end gap-6">
				<button
					onclick={() => (executeModalOpen = false)}
					class="px-8 py-3 text-[11px] font-bold text-slate-500 hover:text-white uppercase tracking-widest transition-all rounded-lg"
				>
					Close
				</button>
				<button
					onclick={executeFunction}
					disabled={executeLoading}
					class="px-10 py-4 bg-emerald-500 hover:bg-emerald-400 text-white font-bold text-[11px] uppercase tracking-widest shadow-lg shadow-emerald-900/20 transition-all active:translate-y-px rounded-xl"
				>
					{#if executeLoading}
						<RefreshCw class="w-4 h-4 animate-spin inline mr-3" />
						Executing...
					{:else}
						<Play class="w-4 h-4 inline mr-3" />
						Run Function
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
		height: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
