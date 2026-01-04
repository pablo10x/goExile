<script lang="ts">
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
		Info
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
			const res = await fetch('/api/database/schemas');
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
			const res = await fetch(`/api/database/functions?schema=${selectedSchema}`, {
				credentials: 'include'
			});

			const text = await res.text();

			if (!res.ok) {
				// Try to parse as JSON error, otherwise use text
				try {
					const err = JSON.parse(text);
					throw new Error(err.error || `Failed to load functions (${res.status})`);
				} catch {
					throw new Error(text || `Failed to load functions (${res.status})`);
				}
			}

			// Parse successful response
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

		// Parse the source to extract the body
		let body = fn.source;
		// Try to extract just the body from the full function definition
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
			const res = await fetch('/api/database/functions', {
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
			const res = await fetch('/api/database/functions/delete', {
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
			const res = await fetch('/api/database/functions/execute', {
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
				return 'text-info bg-blue-500/10 border-blue-500/30';
			case 'procedure':
				return 'text-purple-400 bg-purple-500/10 border-purple-500/30';
			case 'aggregate':
				return 'text-warning bg-amber-500/10 border-amber-500/30';
			case 'window':
				return 'text-info bg-cyan-500/10 border-cyan-500/30';
			default:
				return 'text-text-dim dark:text-text-dim bg-neutral-500/10 border-neutral-500/30';
		}
	}

	function getVolatilityColor(vol: string): string {
		switch (vol.toLowerCase()) {
			case 'immutable':
				return 'text-success bg-success/10 border-emerald-500/30';
			case 'stable':
				return 'text-info bg-blue-500/10 border-blue-500/30';
			case 'volatile':
				return 'text-warning bg-orange-500/10 border-orange-500/30';
			default:
				return 'text-text-dim dark:text-text-dim bg-neutral-500/10 border-neutral-500/30';
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

<div class="h-full flex flex-col bg-transparent">
	<!-- Header -->
	<div
		class="p-6 border-b border-neutral-800 bg-neutral-900/40 backdrop-blur-md"
	>
		<div class="flex justify-between items-start">
			<div class="flex items-center gap-4">
				<div
					class="p-3 bg-indigo-500/10 rounded-xl border border-indigo-500/20 shadow-lg shadow-indigo-900/10"
				>
					<Code2 class="w-7 h-7 text-indigo-400" />
				</div>
				<div>
					<h2 class="text-2xl font-heading font-black text-white uppercase tracking-tighter italic">PROCEDURAL_KERNEL_v2</h2>
					<p class="font-jetbrains text-[10px] text-neutral-500 uppercase tracking-widest mt-1 italic font-bold">
						Manage sector functions, procedures, and triggers
					</p>
				</div>
			</div>
			<button
				onclick={openCreateModal}
				class="relative z-10 px-6 py-3 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 transition-all active:tranneutral-y-px rounded-xl"
			>
				<Plus class="w-5 h-5" />
				New_Logic_Unit
			</button>
		</div>

		<!-- Filters -->
		<div class="flex gap-4 mt-8">
			<div class="relative flex-1 max-w-md">
				<Search
					class="absolute left-4 top-1/2 -tranneutral-y-1/2 w-4 h-4 text-neutral-600 pointer-events-none"
				/>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="SEARCH_LOGIC_IDENTIFIERS..."
					class="w-full pl-12 pr-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-200 font-jetbrains text-[11px] placeholder-neutral-800 focus:border-indigo-500 outline-none transition-all uppercase tracking-widest rounded-xl"
				/>
			</div>

			<select
				bind:value={selectedSchema}
				onchange={() => loadFunctions()}
				class="px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-neutral-400 font-jetbrains text-[11px] focus:border-indigo-500 outline-none cursor-pointer uppercase tracking-widest rounded-xl appearance-none"
			>
				{#each schemas as schema}
					<option value={schema}>{schema}</option>
				{/each}
			</select>

			<button
				onclick={() => loadFunctions()}
				class="p-3 bg-neutral-900/40 border border-neutral-800 text-neutral-500 hover:text-indigo-400 hover:border-indigo-500/30 transition-all rounded-xl"
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
					class="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin shadow-[0_0_15px_rgba(99,102,241,0.4)]"
				></div>
				<p class="mt-6 font-heading font-black text-[11px] text-neutral-600 uppercase tracking-[0.2em] animate-pulse">Synchronizing_Kernel...</p>
			</div>
		{:else if filteredFunctions.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-neutral-700">
				<FileCode class="w-16 h-16 opacity-10 mb-6" />
				<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">No logical units found</p>
				<p class="font-jetbrains text-[9px] mt-2 uppercase font-bold opacity-40 tracking-widest">
					{searchQuery
						? 'Neural filters returned zero results'
						: 'Initialize first procedural unit to begin'}
				</p>
			</div>
		{:else}
			<div class="space-y-4 max-w-6xl mx-auto">
				{#each filteredFunctions as fn (fn.oid)}
					<div
						class="bg-neutral-900/40 border border-neutral-800 overflow-hidden hover:border-indigo-500/40 transition-all group rounded-2xl backdrop-blur-sm shadow-lg"
						transition:fade={{ duration: 150 }}
					>
						<!-- Function Header -->
						<div class="p-5">
							<div class="flex items-start justify-between">
								<div class="flex items-start gap-4 flex-1 min-w-0">
									<button
										onclick={() => toggleExpand(fn.oid)}
										class="p-1 mt-1 text-neutral-600 hover:text-indigo-400 transition-colors"
									>
										{#if expandedFunctions.has(fn.oid)}
											<ChevronDown class="w-5 h-5" />
										{:else}
											<ChevronRight class="w-5 h-5" />
										{/if}
									</button>

									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-4 flex-wrap">
											<h3 class="font-heading font-black text-white text-lg tracking-tight uppercase italic">{fn.name}</h3>
											<span
												class="px-2 py-0.5 text-[9px] font-black font-jetbrains border uppercase tracking-widest rounded {getTypeColor(
													fn.type
												)}"
											>
												{fn.type}
											</span>
											<span
												class="px-2 py-0.5 text-[9px] font-black font-jetbrains border uppercase tracking-widest rounded {getVolatilityColor(
													fn.volatility
												)}"
											>
												{fn.volatility}
											</span>
											<span
												class="px-2 py-0.5 text-[9px] font-black font-jetbrains bg-neutral-950 text-neutral-600 border border-neutral-800 uppercase tracking-widest italic rounded"
											>
												{fn.language}
											</span>
										</div>

										<div class="mt-3 font-jetbrains text-xs font-bold uppercase tracking-tight">
											<span class="text-indigo-400">{fn.name}</span>
											<span class="text-neutral-700">(</span>
											<span class="text-indigo-500/70">{fn.argument_types || ''}</span>
											<span class="text-neutral-700">)</span>
											<span class="text-neutral-700 mx-3">::</span>
											<span class="text-emerald-500">{fn.result_type}</span>
										</div>

										{#if fn.description}
											<p class="mt-3 font-jetbrains text-[10px] text-neutral-500 uppercase tracking-tight font-bold">{fn.description}</p>
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
											class="p-2 text-neutral-600 hover:text-emerald-400 hover:bg-emerald-500/10 transition-all rounded-lg"
											title="Execute"
										>
											<Play class="w-4 h-4" />
										</button>
									{/if}
									<button
										onclick={() => openEditModal(fn)}
										class="p-2 text-neutral-600 hover:text-indigo-400 hover:bg-indigo-500/10 transition-all rounded-lg"
										title="Edit"
									>
										<Edit3 class="w-4 h-4" />
									</button>
									<button
										onclick={() => deleteFunction(fn)}
										class="p-2 text-neutral-600 hover:text-red-400 hover:bg-red-500/10 transition-all rounded-lg"
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
								class="border-t border-neutral-800 bg-neutral-950/40"
								transition:slide={{ duration: 200 }}
							>
								<div class="p-6">
									<div class="flex items-center justify-between mb-4">
										<span class="font-heading text-[10px] font-black text-neutral-600 uppercase tracking-[0.2em] italic"
											>Logic_Source_Readout</span
										>
										<button
											onclick={() => copySource(fn.source)}
											class="flex items-center gap-3 px-4 py-1.5 font-jetbrains text-[9px] font-black text-neutral-500 hover:text-indigo-400 hover:bg-indigo-500/5 border border-neutral-800 transition-all uppercase tracking-widest rounded-lg"
										>
											{#if copied}
												<Check class="w-3 h-3" />
												COPIED_TO_BUFFER
											{:else}
												<Copy class="w-3 h-3" />
												CLONE_SOURCE
											{/if}
										</button>
									</div>
									<pre
										class="p-6 bg-neutral-950 border border-neutral-800 text-[11px] text-neutral-400 font-jetbrains overflow-x-auto max-h-[500px] custom-scrollbar shadow-inner rounded-xl"><code>{fn.source}</code></pre>

									<div class="flex items-center gap-8 mt-6 font-jetbrains text-[9px] font-bold text-neutral-600 uppercase tracking-widest italic">
										<span class="flex items-center gap-2">
											<Shield class="w-3.5 h-3.5" />
											Authorized_Owner: {fn.owner}
										</span>
										<span class="flex items-center gap-2">
											<Database class="w-3.5 h-3.5" />
											Logic_OID: {fn.oid}
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
			class="absolute inset-0 bg-neutral-950/60 backdrop-blur-md cursor-default"
			onclick={closeModal}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-4xl max-h-[90vh] bg-neutral-900/80 backdrop-blur-2xl border border-neutral-800 rounded-3xl shadow-[0_0_50px_rgba(0,0,0,0.8)] overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div
				class="p-8 border-b border-neutral-800 bg-neutral-950/40"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-5">
						<div class="p-3 bg-indigo-500/10 border border-indigo-500/20 rounded-2xl shadow-lg">
							{#if isCreating}
								<Plus class="w-6 h-6 text-indigo-400" />
							{:else}
								<Edit3 class="w-6 h-6 text-indigo-400" />
							{/if}
						</div>
						<div>
							<h3 class="text-2xl font-heading font-black text-white uppercase tracking-tighter italic">
								{isCreating ? 'Initialize_Logic_Unit' : 'Modify_Logic_Buffer'}
							</h3>
							<p class="font-jetbrains text-[10px] text-neutral-500 uppercase tracking-widest mt-1 font-bold">
								{isCreating
									? 'Define a new procedural PostgreSQL unit'
									: `Extracting and editing ${selectedFunction?.name}`}
							</p>
						</div>
					</div>
					<button
						onclick={closeModal}
						class="p-2 text-neutral-500 hover:text-white transition-all rounded-lg"
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
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Schema_Vector</label
						>
						<select
							id="fnSchema"
							bind:value={formData.schema}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-300 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all uppercase appearance-none rounded-xl"
						>
							{#each schemas as schema}
								<option value={schema}>{schema}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-3">
						<label
							for="fnName"
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Identifier_Tag</label
						>
						<input
							id="fnName"
							type="text"
							bind:value={formData.name}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-200 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all uppercase rounded-xl"
							placeholder="NULL_PTR"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-8">
					<div class="space-y-3">
						<label
							for="fnArgs"
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Argument_Buffer</label
						>
						<input
							id="fnArgs"
							type="text"
							bind:value={formData.arguments}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-200 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all rounded-xl"
							placeholder="arg1 integer, arg2 text"
						/>
					</div>
					<div class="space-y-3">
						<label
							for="fnReturns"
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Return_Signature</label
						>
						<select
							id="fnReturns"
							bind:value={formData.returns}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-300 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all appearance-none rounded-xl"
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
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Language_Engine</label
						>
						<select
							id="fnLang"
							bind:value={formData.language}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-300 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all appearance-none rounded-xl"
						>
							{#each languages as lang}
								<option value={lang}>{lang}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-3">
						<label
							for="fnVol"
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
							>Volatility_State</label
						>
						<select
							id="fnVol"
							bind:value={formData.volatility}
							class="w-full px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-300 font-jetbrains text-xs focus:border-indigo-500 outline-none transition-all appearance-none rounded-xl"
						>
							{#each volatilities as vol}
								<option value={vol}>{vol}</option>
							{/each}
						</select>
					</div>
					<div class="flex flex-col justify-end gap-3">
						<label
							class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-indigo-500/50 transition-colors rounded-xl"
						>
							<input
								type="checkbox"
								bind:checked={formData.isStrict}
								class="rounded text-indigo-500 focus:ring-indigo-500 bg-neutral-900 border-neutral-700"
							/>
							<span class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest">Strict_Check</span>
						</label>
						<label
							class="flex items-center gap-3 p-3 bg-neutral-950/40 border border-neutral-800 cursor-pointer hover:border-indigo-500/50 transition-colors rounded-xl"
						>
							<input
								type="checkbox"
								bind:checked={formData.securityDefiner}
								class="rounded text-indigo-500 focus:ring-indigo-500 bg-neutral-900 border-neutral-700"
							/>
							<span class="text-[10px] font-bold text-neutral-500 uppercase tracking-widest">Sec_Definer</span>
						</label>
					</div>
				</div>

				<!-- Function Body -->
				<div class="space-y-3">
					<label
						for="fnBody"
						class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest italic"
						>Logic_Kernel_Body</label
					>
					<div class="relative">
						<div class="absolute top-4 left-4 text-indigo-500 font-bold text-xs opacity-20 pointer-events-none tracking-widest">BEGIN</div>
						<textarea
							id="fnBody"
							bind:value={formData.body}
							rows="15"
							class="w-full px-10 py-8 bg-neutral-950/60 border border-neutral-800 text-stone-200 focus:border-indigo-500 outline-none font-mono text-sm resize-none leading-relaxed shadow-inner rounded-2xl"
							placeholder="-- Your code here"
						></textarea>
						<div class="absolute bottom-4 right-4 text-indigo-500 font-bold text-xs opacity-20 pointer-events-none tracking-widest">END;</div>
					</div>
				</div>
			</div>

			<!-- Modal Footer -->
			<div
				class="p-8 border-t border-neutral-800 bg-neutral-950/40 flex justify-between items-center"
			>
				<div class="font-jetbrains text-[8px] font-bold text-neutral-700 uppercase tracking-[0.5em] italic">
					Waiting_For_Acknowledge
				</div>
				<div class="flex gap-6">
					<button
						onclick={closeModal}
						class="px-8 py-3 text-[11px] font-bold text-neutral-500 hover:text-white uppercase tracking-widest italic transition-all rounded-lg"
					>
						[Abort]
					</button>
					<button
						onclick={saveFunction}
						disabled={loading}
						class="px-10 py-4 bg-indigo-500 hover:bg-indigo-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-indigo-900/20 transition-all disabled:opacity-20 active:tranneutral-y-px rounded-xl"
					>
						{#if loading}
							<RefreshCw class="w-4 h-4 animate-spin inline mr-3" />
							SYCNING...
						{:else}
							<Check class="w-4 h-4 inline mr-3" />
							COMMIT_LOGIC
						{/if}
					</button>
				</div>
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
			class="absolute inset-0 bg-neutral-950/60 backdrop-blur-sm cursor-default"
			onclick={() => (executeModalOpen = false)}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-2xl max-h-[80vh] bg-neutral-900/80 backdrop-blur-2xl border border-neutral-800 rounded-3xl shadow-[0_0_50px_rgba(0,0,0,0.8)] overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div
				class="p-8 border-b border-neutral-800 bg-neutral-950/40"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-5">
						<div class="p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-2xl shadow-lg">
							<Play class="w-6 h-6 text-emerald-400" />
						</div>
						<div>
							<h3 class="text-2xl font-heading font-black text-white uppercase tracking-tighter italic">Execute_Unit</h3>
							<p class="font-jetbrains text-[10px] text-neutral-500 font-bold uppercase tracking-widest mt-1 italic">{functionToExecute.name}()</p>
						</div>
					</div>
					<button
						onclick={() => (executeModalOpen = false)}
						class="p-2 text-neutral-500 hover:text-white transition-all rounded-lg"
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
							class="block font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest mb-6 italic border-l border-neutral-800 pl-4"
						>
							Input_Arguments
						</div>
						<div class="space-y-4">
							{#each functionToExecute.argument_types.split(',') as arg, i}
								<div class="flex items-center gap-6 group">
									<label for={`execArg-${i}`} class="text-[10px] font-jetbrains font-bold text-neutral-500 uppercase tracking-widest min-w-[140px] italic group-hover:text-indigo-400 transition-colors"
										>{arg.trim()}</label
									>
									<input
										id={`execArg-${i}`}
										type="text"
										bind:value={executeArgs[i]}
										class="flex-1 px-4 py-3 bg-neutral-950/40 border border-neutral-800 text-stone-200 focus:border-indigo-500 outline-none font-mono text-xs transition-all shadow-inner rounded-xl"
										placeholder="NULL_VAL"
									/>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center py-10 opacity-40">
						<Info class="w-10 h-10 text-neutral-700 mb-4" />
						<p class="font-jetbrains text-[10px] font-bold text-neutral-500 uppercase tracking-widest">Procedural unit requires no inputs.</p>
					</div>
				{/if}

				{#if executeResult.length > 0}
					<div role="group" aria-labelledby="result-label" transition:slide>
						<div
							id="result-label"
							class="block font-jetbrains text-[10px] font-bold text-emerald-500 uppercase tracking-widest mb-6 italic border-l border-emerald-900/50 pl-4"
						>
							Output_Result_Buffer
						</div>
						<div
							class="bg-neutral-950/40 border border-neutral-800 rounded-2xl overflow-hidden shadow-inner"
						>
							<div class="overflow-x-auto custom-scrollbar">
								<table class="w-full text-xs font-jetbrains">
									<thead class="bg-neutral-950 text-neutral-500 border-b border-neutral-800">
										<tr>
											{#each Object.keys(executeResult[0]) as key}
												<th
													class="px-6 py-4 text-left font-bold uppercase tracking-widest border-r border-neutral-900 italic"
													>{key}</th
												>
											{/each}
										</tr>
									</thead>
									<tbody class="divide-y divide-neutral-900">
										{#each executeResult as row}
											<tr class="hover:bg-indigo-500/5 transition-colors">
												{#each Object.values(row) as val}
													<td class="px-6 py-4 text-neutral-400 font-medium tracking-tight">
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
			<div
				class="p-8 border-t border-neutral-800 bg-neutral-950/40 flex justify-end gap-6"
			>
				<button
					onclick={() => (executeModalOpen = false)}
					class="px-8 py-3 text-[11px] font-bold text-neutral-500 hover:text-white uppercase tracking-widest italic transition-all rounded-lg"
				>
					[Close]
				</button>
				<button
					onclick={executeFunction}
					disabled={executeLoading}
					class="px-10 py-4 bg-emerald-500 hover:bg-emerald-400 text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-emerald-900/20 transition-all active:tranneutral-y-px rounded-xl"
				>
					{#if executeLoading}
						<RefreshCw class="w-4 h-4 animate-spin inline mr-3" />
						PROCESSING...
					{:else}
						<Play class="w-4 h-4 inline mr-3" />
						Execute_Sequence
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
