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
		Database
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { notifications } from '$lib/stores';
	import { fade, slide } from 'svelte/transition';

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
				return 'text-blue-400 bg-blue-500/10 border-blue-500/30';
			case 'procedure':
				return 'text-purple-400 bg-purple-500/10 border-purple-500/30';
			case 'aggregate':
				return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
			case 'window':
				return 'text-cyan-400 bg-cyan-500/10 border-cyan-500/30';
			default:
				return 'text-slate-500 dark:text-slate-400 bg-slate-500/10 border-slate-500/30';
		}
	}

	function getVolatilityColor(vol: string): string {
		switch (vol.toLowerCase()) {
			case 'immutable':
				return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
			case 'stable':
				return 'text-blue-400 bg-blue-500/10 border-blue-500/30';
			case 'volatile':
				return 'text-orange-400 bg-orange-500/10 border-orange-500/30';
			default:
				return 'text-slate-500 dark:text-slate-400 bg-slate-500/10 border-slate-500/30';
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

<div class="h-full flex flex-col bg-slate-900">
	<!-- Header -->
	<div
		class="p-6 border-b border-slate-200 dark:border-slate-800 bg-gradient-to-r from-slate-950 via-slate-900 to-slate-950"
	>
		<div class="flex justify-between items-start">
			<div class="flex items-center gap-4">
				<div
					class="p-3 bg-gradient-to-br from-violet-500/20 to-purple-500/20 rounded-xl border border-violet-500/30 shadow-lg shadow-violet-500/10"
				>
					<Code2 class="w-7 h-7 text-violet-400" />
				</div>
				<div>
					<h2 class="text-2xl font-bold text-slate-100">Functions & Procedures</h2>
					<p class="text-sm text-slate-500 mt-0.5">
						Manage PostgreSQL functions, procedures, and triggers
					</p>
				</div>
			</div>
			<button
				onclick={openCreateModal}
				class="px-5 py-2.5 bg-gradient-to-r from-violet-600 to-purple-600 hover:from-violet-500 hover:to-purple-500 text-slate-900 dark:text-white rounded-xl font-bold flex items-center gap-2 shadow-lg shadow-violet-900/30 transition-all transform hover:-translate-y-0.5 active:scale-95"
			>
				<Plus class="w-5 h-5" />
				New Function
			</button>
		</div>

		<!-- Filters -->
		<div class="flex gap-4 mt-6">
			<div class="relative flex-1 max-w-md">
				<Search
					class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 pointer-events-none"
				/>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search functions..."
					class="w-full pl-10 pr-4 py-2.5 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl text-slate-800 dark:text-slate-200 placeholder-slate-500 focus:border-violet-500/50 focus:ring-2 focus:ring-violet-500/20 outline-none transition-all"
				/>
			</div>

			<select
				bind:value={selectedSchema}
				onchange={() => loadFunctions()}
				class="px-4 py-2.5 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500/50 outline-none cursor-pointer"
			>
				{#each schemas as schema}
					<option value={schema}>{schema}</option>
				{/each}
			</select>

			<button
				onclick={() => loadFunctions()}
				class="p-2.5 bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl text-slate-500 dark:text-slate-400 hover:text-violet-400 hover:border-violet-500/50 transition-all"
				title="Refresh"
			>
				<RefreshCw class="w-5 h-5 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Functions List -->
	<div class="flex-1 overflow-auto p-6">
		{#if loading && functions.length === 0}
			<div class="flex flex-col items-center justify-center h-64">
				<div
					class="w-12 h-12 border-4 border-violet-500 border-t-transparent rounded-full animate-spin"
				></div>
				<p class="mt-4 text-slate-500">Loading functions...</p>
			</div>
		{:else if filteredFunctions.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-slate-500">
				<FileCode class="w-16 h-16 opacity-20 mb-4" />
				<p class="text-lg font-medium">No functions found</p>
				<p class="text-sm mt-1">
					{searchQuery
						? 'Try a different search term'
						: 'Create your first function to get started'}
				</p>
			</div>
		{:else}
			<div class="space-y-3">
				{#each filteredFunctions as fn (fn.oid)}
					<div
						class="bg-slate-800/50 border border-slate-300/50 dark:border-slate-700/50 rounded-xl overflow-hidden hover:border-slate-600/50 transition-all group"
						transition:fade={{ duration: 150 }}
					>
						<!-- Function Header -->
						<div class="p-4">
							<div class="flex items-start justify-between">
								<div class="flex items-start gap-3 flex-1 min-w-0">
									<button
										onclick={() => toggleExpand(fn.oid)}
										class="p-1 mt-0.5 text-slate-500 hover:text-slate-700 dark:text-slate-300 transition-colors"
									>
										{#if expandedFunctions.has(fn.oid)}
											<ChevronDown class="w-5 h-5" />
										{:else}
											<ChevronRight class="w-5 h-5" />
										{/if}
									</button>

									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-3 flex-wrap">
											<h3 class="text-lg font-bold text-slate-100 font-mono">{fn.name}</h3>
											<span
												class="px-2 py-0.5 text-xs font-semibold rounded-full border {getTypeColor(
													fn.type
												)}"
											>
												{fn.type}
											</span>
											<span
												class="px-2 py-0.5 text-xs font-semibold rounded-full border {getVolatilityColor(
													fn.volatility
												)}"
											>
												{fn.volatility}
											</span>
											<span
												class="px-2 py-0.5 text-xs font-mono rounded-full bg-slate-700/50 text-slate-500 dark:text-slate-400 border border-slate-600/50"
											>
												{fn.language}
											</span>
										</div>

										<div class="mt-2 text-sm text-slate-500 dark:text-slate-400 font-mono">
											<span class="text-violet-400">{fn.name}</span>
											<span class="text-slate-500">(</span>
											<span class="text-blue-300">{fn.argument_types || ''}</span>
											<span class="text-slate-500">)</span>
											<span class="text-slate-500 mx-2">→</span>
											<span class="text-emerald-400">{fn.result_type}</span>
										</div>

										{#if fn.description}
											<p class="mt-2 text-sm text-slate-500">{fn.description}</p>
										{/if}
									</div>
								</div>

								<!-- Actions -->
								<div
									class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
								>
									{#if fn.type === 'function'}
										<button
											onclick={() => openExecuteModal(fn)}
											class="p-2 text-slate-500 dark:text-slate-400 hover:text-emerald-400 hover:bg-emerald-500/10 rounded-lg transition-all"
											title="Execute"
										>
											<Play class="w-4 h-4" />
										</button>
									{/if}
									<button
										onclick={() => openEditModal(fn)}
										class="p-2 text-slate-500 dark:text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
										title="Edit"
									>
										<Edit3 class="w-4 h-4" />
									</button>
									<button
										onclick={() => deleteFunction(fn)}
										class="p-2 text-slate-500 dark:text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
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
								class="border-t border-slate-300/50 dark:border-slate-700/50 bg-white/50 dark:bg-slate-950/50"
								transition:slide={{ duration: 200 }}
							>
								<div class="p-4">
									<div class="flex items-center justify-between mb-3">
										<span class="text-xs font-bold text-slate-500 uppercase tracking-wider"
											>Source Code</span
										>
										<button
											onclick={() => copySource(fn.source)}
											class="flex items-center gap-1.5 px-2 py-1 text-xs text-slate-500 dark:text-slate-400 hover:text-violet-400 hover:bg-violet-500/10 rounded transition-all"
										>
											{#if copied}
												<Check class="w-3.5 h-3.5" />
												Copied!
											{:else}
												<Copy class="w-3.5 h-3.5" />
												Copy
											{/if}
										</button>
									</div>
									<pre
										class="p-4 bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 text-sm text-slate-700 dark:text-slate-300 font-mono overflow-x-auto max-h-96"><code
											>{fn.source}</code
										></pre>

									<div class="flex items-center gap-4 mt-4 text-xs text-slate-500">
										<span class="flex items-center gap-1.5">
											<Shield class="w-3.5 h-3.5" />
											Owner: {fn.owner}
										</span>
										<span class="flex items-center gap-1.5">
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
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		transition:fade={{ duration: 150 }}
	>
		<button
			class="absolute inset-0 bg-white/80 dark:bg-slate-950/80 backdrop-blur-sm cursor-default"
			onclick={closeModal}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-4xl max-h-[90vh] bg-slate-900 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div
				class="p-6 border-b border-slate-200 dark:border-slate-800 bg-gradient-to-r from-violet-500/10 via-transparent to-purple-500/10"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-violet-500/20 rounded-lg">
							{#if isCreating}
								<Plus class="w-5 h-5 text-violet-400" />
							{:else}
								<Edit3 class="w-5 h-5 text-violet-400" />
							{/if}
						</div>
						<div>
							<h3 class="text-xl font-bold text-slate-100">
								{isCreating ? 'Create New Function' : 'Edit Function'}
							</h3>
							<p class="text-sm text-slate-500">
								{isCreating
									? 'Define a new PostgreSQL function'
									: `Editing ${selectedFunction?.name}`}
							</p>
						</div>
					</div>
					<button
						onclick={closeModal}
						class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-800 rounded-lg transition-all"
					>
						<X class="w-5 h-5" />
					</button>
				</div>
			</div>

			<!-- Modal Body -->
			<div class="flex-1 overflow-y-auto p-6 space-y-6">
				<!-- Basic Info -->
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label
							for="fnSchema"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Schema</label
						>
						<select
							id="fnSchema"
							bind:value={formData.schema}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none"
						>
							{#each schemas as schema}
								<option value={schema}>{schema}</option>
							{/each}
						</select>
					</div>
					<div>
						<label
							for="fnName"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Function Name</label
						>
						<input
							id="fnName"
							type="text"
							bind:value={formData.name}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none font-mono"
							placeholder="my_function"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label
							for="fnArgs"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Arguments</label
						>
						<input
							id="fnArgs"
							type="text"
							bind:value={formData.arguments}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none font-mono"
							placeholder="arg1 integer, arg2 text"
						/>
					</div>
					<div>
						<label
							for="fnReturns"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Returns</label
						>
						<select
							id="fnReturns"
							bind:value={formData.returns}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none"
						>
							{#each returnTypes as type}
								<option value={type}>{type}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="grid grid-cols-3 gap-4">
					<div>
						<label
							for="fnLang"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Language</label
						>
						<select
							id="fnLang"
							bind:value={formData.language}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none"
						>
							{#each languages as lang}
								<option value={lang}>{lang}</option>
							{/each}
						</select>
					</div>
					<div>
						<label
							for="fnVol"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
							>Volatility</label
						>
						<select
							id="fnVol"
							bind:value={formData.volatility}
							class="w-full px-4 py-2.5 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none"
						>
							{#each volatilities as vol}
								<option value={vol}>{vol}</option>
							{/each}
						</select>
					</div>
					<div class="flex flex-col justify-end gap-2">
						<label
							class="flex items-center gap-2 p-2 rounded-lg bg-slate-800 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-violet-500/50 transition-colors"
						>
							<input
								type="checkbox"
								bind:checked={formData.isStrict}
								class="rounded text-violet-500 focus:ring-violet-500 bg-slate-700 border-slate-600"
							/>
							<span class="text-sm text-slate-700 dark:text-slate-300">Strict</span>
						</label>
						<label
							class="flex items-center gap-2 p-2 rounded-lg bg-slate-800 border border-slate-300 dark:border-slate-700 cursor-pointer hover:border-violet-500/50 transition-colors"
						>
							<input
								type="checkbox"
								bind:checked={formData.securityDefiner}
								class="rounded text-violet-500 focus:ring-violet-500 bg-slate-700 border-slate-600"
							/>
							<span class="text-sm text-slate-700 dark:text-slate-300">Security Definer</span>
						</label>
					</div>
				</div>

				<!-- Function Body -->
				<div>
					<label
						for="fnBody"
						class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-2"
						>Function Body</label
					>
					<textarea
						id="fnBody"
						bind:value={formData.body}
						rows="15"
						class="w-full px-4 py-3 bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-xl text-slate-800 dark:text-slate-200 focus:border-violet-500 outline-none font-mono text-sm resize-none"
						placeholder="BEGIN
  -- Your code here
  RETURN result;
END;"
					></textarea>
				</div>
			</div>

			<!-- Modal Footer -->
			<div
				class="p-6 border-t border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-950/50 flex justify-end gap-3"
			>
				<button
					onclick={closeModal}
					class="px-5 py-2.5 text-sm font-semibold text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-800 rounded-xl transition-all"
				>
					Cancel
				</button>
				<button
					onclick={saveFunction}
					disabled={loading}
					class="px-6 py-2.5 bg-gradient-to-r from-violet-600 to-purple-600 hover:from-violet-500 hover:to-purple-500 text-slate-900 dark:text-white rounded-xl font-bold shadow-lg disabled:opacity-50 transition-all flex items-center gap-2"
				>
					{#if loading}
						<RefreshCw class="w-4 h-4 animate-spin" />
					{:else}
						<Check class="w-4 h-4" />
					{/if}
					{isCreating ? 'Create Function' : 'Save Changes'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Execute Function Modal -->
{#if executeModalOpen && functionToExecute}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		transition:fade={{ duration: 150 }}
	>
		<button
			class="absolute inset-0 bg-white/80 dark:bg-slate-950/80 backdrop-blur-sm cursor-default"
			onclick={() => (executeModalOpen = false)}
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-2xl max-h-[80vh] bg-slate-900 border border-slate-300/50 dark:border-slate-700/50 rounded-2xl shadow-2xl overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div
				class="p-6 border-b border-slate-200 dark:border-slate-800 bg-gradient-to-r from-emerald-500/10 via-transparent to-cyan-500/10"
			>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-emerald-500/20 rounded-lg">
							<Zap class="w-5 h-5 text-emerald-400" />
						</div>
						<div>
							<h3 class="text-xl font-bold text-slate-100">Execute Function</h3>
							<p class="text-sm text-slate-500 font-mono">{functionToExecute.name}()</p>
						</div>
					</div>
					<button
						onclick={() => (executeModalOpen = false)}
						class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-800 rounded-lg transition-all"
					>
						<X class="w-5 h-5" />
					</button>
				</div>
			</div>

			<!-- Modal Body -->
			<div class="flex-1 overflow-y-auto p-6 space-y-6">
				{#if functionToExecute.argument_types}
					<div role="group" aria-labelledby="args-label">
						<div
							id="args-label"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-3"
						>
							Arguments
						</div>
						<div class="space-y-3">
							{#each functionToExecute.argument_types.split(',') as arg, i}
								<div class="flex items-center gap-3">
									<label for={`execArg-${i}`} class="text-sm text-slate-500 font-mono min-w-[120px]"
										>{arg.trim()}</label
									>
									<input
										id={`execArg-${i}`}
										type="text"
										bind:value={executeArgs[i]}
										class="flex-1 px-4 py-2 bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg text-slate-800 dark:text-slate-200 focus:border-emerald-500 outline-none font-mono"
										placeholder="value"
									/>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<p class="text-slate-500 text-sm">This function takes no arguments.</p>
				{/if}

				{#if executeResult.length > 0}
					<div role="group" aria-labelledby="result-label">
						<div
							id="result-label"
							class="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase mb-3"
						>
							Result
						</div>
						<div
							class="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden"
						>
							<div class="overflow-x-auto">
								<table class="w-full text-sm">
									<thead class="bg-slate-900 text-slate-500 dark:text-slate-400">
										<tr>
											{#each Object.keys(executeResult[0]) as key}
												<th
													class="px-4 py-2 text-left font-medium border-b border-slate-200 dark:border-slate-800"
													>{key}</th
												>
											{/each}
										</tr>
									</thead>
									<tbody class="divide-y divide-slate-800">
										{#each executeResult as row}
											<tr class="hover:bg-slate-800/50">
												{#each Object.values(row) as val}
													<td class="px-4 py-2 text-slate-700 dark:text-slate-300 font-mono">
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
				class="p-6 border-t border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-950/50 flex justify-end gap-3"
			>
				<button
					onclick={() => (executeModalOpen = false)}
					class="px-5 py-2.5 text-sm font-semibold text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200 hover:bg-slate-800 rounded-xl transition-all"
				>
					Close
				</button>
				<button
					onclick={executeFunction}
					disabled={executeLoading}
					class="px-6 py-2.5 bg-gradient-to-r from-emerald-600 to-cyan-600 hover:from-emerald-500 hover:to-cyan-500 text-slate-900 dark:text-white rounded-xl font-bold shadow-lg disabled:opacity-50 transition-all flex items-center gap-2"
				>
					{#if executeLoading}
						<RefreshCw class="w-4 h-4 animate-spin" />
						Executing...
					{:else}
						<Play class="w-4 h-4" />
						Execute
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
