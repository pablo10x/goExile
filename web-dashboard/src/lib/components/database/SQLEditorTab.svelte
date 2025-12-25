<script lang="ts">
	import { Play, RefreshCw, Terminal, Download, Copy, Check, Database } from 'lucide-svelte';
	import { notifications } from '$lib/stores';
	import { slide } from 'svelte/transition';
	import { onMount } from 'svelte';

	let query = $state('');
	let loading = $state(false);
	let results = $state<any[]>([]);
	let error = $state<string | null>(null);
	let executionTime = $state(0);

	// Autocomplete state
	let textareaEl: HTMLTextAreaElement;
	let showAutocomplete = $state(false);
	let suggestions = $state<string[]>([]);
	let selectedIndex = $state(0);
	let autocompletePosition = $state({ top: 0, left: 0 });
	let currentWord = $state('');
	let wordStartPos = $state(0);

	// Database metadata for autocomplete
	let tables = $state<{ schema: string; table: string }[]>([]);
	let tableColumns = $state<Record<string, string[]>>({});
	let loadingMetadata = $state(false);

	// SQL Keywords for autocomplete
	const SQL_KEYWORDS = [
		'SELECT',
		'FROM',
		'WHERE',
		'AND',
		'OR',
		'NOT',
		'IN',
		'LIKE',
		'ILIKE',
		'INSERT',
		'INTO',
		'VALUES',
		'UPDATE',
		'SET',
		'DELETE',
		'CREATE',
		'TABLE',
		'INDEX',
		'VIEW',
		'SCHEMA',
		'DATABASE',
		'DROP',
		'ALTER',
		'ADD',
		'COLUMN',
		'CONSTRAINT',
		'JOIN',
		'INNER',
		'LEFT',
		'RIGHT',
		'OUTER',
		'FULL',
		'CROSS',
		'ON',
		'GROUP',
		'BY',
		'ORDER',
		'ASC',
		'DESC',
		'HAVING',
		'LIMIT',
		'OFFSET',
		'DISTINCT',
		'AS',
		'CASE',
		'WHEN',
		'THEN',
		'ELSE',
		'END',
		'NULL',
		'IS',
		'TRUE',
		'FALSE',
		'BETWEEN',
		'EXISTS',
		'COUNT',
		'SUM',
		'AVG',
		'MIN',
		'MAX',
		'COALESCE',
		'NULLIF',
		'CAST',
		'CONVERT',
		'SUBSTRING',
		'CONCAT',
		'LENGTH',
		'TRIM',
		'UPPER',
		'LOWER',
		'NOW',
		'CURRENT_DATE',
		'CURRENT_TIME',
		'CURRENT_TIMESTAMP',
		'PRIMARY',
		'KEY',
		'FOREIGN',
		'REFERENCES',
		'UNIQUE',
		'CHECK',
		'DEFAULT',
		'CASCADE',
		'RESTRICT',
		'SET NULL',
		'NO ACTION',
		'BEGIN',
		'COMMIT',
		'ROLLBACK',
		'TRANSACTION',
		'SAVEPOINT',
		'GRANT',
		'REVOKE',
		'ALL',
		'PRIVILEGES',
		'TO',
		'PUBLIC',
		'UNION',
		'INTERSECT',
		'EXCEPT',
		'WITH',
		'RECURSIVE',
		'RETURNING',
		'EXPLAIN',
		'ANALYZE',
		'VACUUM',
		'TRUNCATE',
		'SERIAL',
		'BIGSERIAL',
		'INTEGER',
		'BIGINT',
		'SMALLINT',
		'NUMERIC',
		'DECIMAL',
		'REAL',
		'DOUBLE',
		'PRECISION',
		'VARCHAR',
		'CHAR',
		'TEXT',
		'BOOLEAN',
		'DATE',
		'TIME',
		'TIMESTAMP',
		'TIMESTAMPTZ',
		'INTERVAL',
		'UUID',
		'JSON',
		'JSONB',
		'ARRAY'
	];

	// Load database metadata on mount
	onMount(() => {
		loadMetadata();
	});

	async function loadMetadata() {
		loadingMetadata = true;
		try {
			// Load all tables
			const tablesRes = await fetch('/api/database/all-tables');
			if (tablesRes.ok) {
				tables = await tablesRes.json();
			}
		} catch (e) {
			console.error('Failed to load table metadata:', e);
		} finally {
			loadingMetadata = false;
		}
	}

	async function loadColumnsForTable(schema: string, table: string) {
		const key = `${schema}.${table}`;
		if (tableColumns[key]) return; // Already loaded

		try {
			const res = await fetch(
				`/api/database/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`
			);
			if (res.ok) {
				const cols = await res.json();
				tableColumns[key] = cols.map((c: any) => c.column_name || c.name);
			}
		} catch (e) {
			console.error(`Failed to load columns for ${key}:`, e);
		}
	}

	// Preload columns for frequently used tables
	$effect(() => {
		if (tables.length > 0) {
			// Load columns for first 10 tables
			tables.slice(0, 10).forEach((t) => {
				loadColumnsForTable(t.schema, t.table);
			});
		}
	});

	function getWordAtCursor(): { word: string; start: number; end: number } {
		if (!textareaEl) return { word: '', start: 0, end: 0 };

		const cursorPos = textareaEl.selectionStart;
		const text = query;

		// Find word boundaries
		let start = cursorPos;
		let end = cursorPos;

		// Move start back to find word beginning
		while (start > 0 && /[\w.]/.test(text[start - 1])) {
			start--;
		}

		// Move end forward to find word end
		while (end < text.length && /[\w.]/.test(text[end])) {
			end++;
		}

		return {
			word: text.substring(start, cursorPos),
			start,
			end
		};
	}

	function getContextKeyword(): string | null {
		if (!textareaEl) return null;

		const cursorPos = textareaEl.selectionStart;
		const textBefore = query.substring(0, cursorPos).toUpperCase();

		// Check recent keywords to determine context
		const keywords = textBefore.split(/\s+/).filter(Boolean);
		const recentKeywords = keywords.slice(-5);

		if (
			recentKeywords.includes('FROM') ||
			recentKeywords.includes('JOIN') ||
			recentKeywords.includes('INTO') ||
			recentKeywords.includes('UPDATE') ||
			recentKeywords.includes('TABLE')
		) {
			return 'TABLE';
		}
		if (
			recentKeywords.includes('SELECT') ||
			recentKeywords.includes('WHERE') ||
			recentKeywords.includes('SET') ||
			recentKeywords.includes('ON') ||
			recentKeywords.includes('BY')
		) {
			return 'COLUMN';
		}
		return null;
	}

	function getSuggestions(word: string): string[] {
		if (!word) return [];

		const upperWord = word.toUpperCase();
		const lowerWord = word.toLowerCase();
		const context = getContextKeyword();
		let items: string[] = [];

		// Add SQL keywords
		items = items.concat(SQL_KEYWORDS.filter((k) => k.startsWith(upperWord)));

		// Add table names (with schema prefix for non-public schemas)
		const tableNames = tables.map((t) =>
			t.schema === 'public' ? t.table : `${t.schema}.${t.table}`
		);
		items = items.concat(tableNames.filter((t) => t.toLowerCase().startsWith(lowerWord)));

		// Add column names from all loaded tables
		for (const [tableKey, cols] of Object.entries(tableColumns)) {
			const matchingCols = cols.filter((c) => c.toLowerCase().startsWith(lowerWord));
			items = items.concat(matchingCols);
		}

		// Check if word contains a dot (table.column reference)
		if (word.includes('.')) {
			const [tableRef, colPrefix] = word.split('.');
			const lowerTableRef = tableRef.toLowerCase();

			// Find matching table
			const matchingTable = tables.find(
				(t) =>
					t.table.toLowerCase() === lowerTableRef ||
					`${t.schema}.${t.table}`.toLowerCase() === lowerTableRef
			);

			if (matchingTable) {
				const key = `${matchingTable.schema}.${matchingTable.table}`;
				// Load columns if not loaded
				if (!tableColumns[key]) {
					loadColumnsForTable(matchingTable.schema, matchingTable.table);
				}
				const cols = tableColumns[key] || [];
				const colSuggestions = cols
					.filter((c) => !colPrefix || c.toLowerCase().startsWith(colPrefix.toLowerCase()))
					.map((c) => `${tableRef}.${c}`);
				items = colSuggestions;
			}
		}

		// Prioritize based on context
		if (context === 'TABLE') {
			// Put table names first
			items.sort((a, b) => {
				const aIsTable = tableNames.includes(a);
				const bIsTable = tableNames.includes(b);
				if (aIsTable && !bIsTable) return -1;
				if (!aIsTable && bIsTable) return 1;
				return 0;
			});
		} else if (context === 'COLUMN') {
			// Put column names first
			const allCols = new Set(Object.values(tableColumns).flat());
			items.sort((a, b) => {
				const aIsCol = allCols.has(a);
				const bIsCol = allCols.has(b);
				if (aIsCol && !bIsCol) return -1;
				if (!aIsCol && bIsCol) return 1;
				return 0;
			});
		}

		// Remove duplicates and limit results
		return [...new Set(items)].slice(0, 15);
	}

	function updateAutocomplete() {
		const { word, start } = getWordAtCursor();
		currentWord = word;
		wordStartPos = start;

		if (word.length < 1) {
			showAutocomplete = false;
			return;
		}

		const newSuggestions = getSuggestions(word);
		if (newSuggestions.length === 0) {
			showAutocomplete = false;
			return;
		}

		suggestions = newSuggestions;
		selectedIndex = 0;
		showAutocomplete = true;

		// Calculate position
		updateAutocompletePosition();
	}

	function updateAutocompletePosition() {
		if (!textareaEl) return;

		// Create a mirror div to measure text position
		const mirror = document.createElement('div');
		const computed = window.getComputedStyle(textareaEl);

		// Copy styles
		mirror.style.cssText = `
			position: absolute;
			visibility: hidden;
			white-space: pre-wrap;
			word-wrap: break-word;
			font-family: ${computed.fontFamily};
			font-size: ${computed.fontSize};
			line-height: ${computed.lineHeight};
			padding: ${computed.padding};
			border: ${computed.border};
			width: ${textareaEl.clientWidth}px;
		`;

		const textBefore = query.substring(0, textareaEl.selectionStart);
		mirror.textContent = textBefore;

		const span = document.createElement('span');
		span.textContent = '|';
		mirror.appendChild(span);

		document.body.appendChild(mirror);

		const rect = textareaEl.getBoundingClientRect();
		const spanRect = span.getBoundingClientRect();
		const mirrorRect = mirror.getBoundingClientRect();

		// Calculate relative position within textarea
		const relativeTop = spanRect.top - mirrorRect.top;
		const relativeLeft = spanRect.left - mirrorRect.left;

		// Get textarea's scroll position
		const scrollTop = textareaEl.scrollTop;

		autocompletePosition = {
			top: Math.min(relativeTop - scrollTop + 24, textareaEl.clientHeight - 100),
			left: Math.min(relativeLeft, textareaEl.clientWidth - 250)
		};

		document.body.removeChild(mirror);
	}

	function applySuggestion(suggestion: string) {
		const before = query.substring(0, wordStartPos);
		const { end } = getWordAtCursor();
		const after = query.substring(end);

		// Add space after keywords, nothing after table.column references
		const suffix = SQL_KEYWORDS.includes(suggestion.toUpperCase()) ? ' ' : '';

		query = before + suggestion + suffix + after;
		showAutocomplete = false;

		// Move cursor to end of inserted text
		const newPos = wordStartPos + suggestion.length + suffix.length;
		setTimeout(() => {
			textareaEl?.setSelectionRange(newPos, newPos);
			textareaEl?.focus();
		}, 0);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (showAutocomplete) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				selectedIndex = (selectedIndex + 1) % suggestions.length;
				return;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				selectedIndex = (selectedIndex - 1 + suggestions.length) % suggestions.length;
				return;
			}
			if (e.key === 'Tab' || e.key === 'Enter') {
				if (suggestions.length > 0) {
					e.preventDefault();
					applySuggestion(suggestions[selectedIndex]);
					return;
				}
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				showAutocomplete = false;
				return;
			}
		}

		// Run query shortcut
		if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			runQuery();
		}
	}

	function handleInput() {
		// Small delay to let the value update
		setTimeout(updateAutocomplete, 10);
	}

	function handleBlur(e: FocusEvent) {
		// Delay hiding to allow click on suggestion
		setTimeout(() => {
			showAutocomplete = false;
		}, 200);
	}

	async function runQuery() {
		if (!query.trim()) return;
		loading = true;
		error = null;
		results = [];
		showAutocomplete = false;
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
			// Refresh metadata after write operations
			const upperQuery = query.toUpperCase().trim();
			if (
				upperQuery.startsWith('CREATE') ||
				upperQuery.startsWith('DROP') ||
				upperQuery.startsWith('ALTER')
			) {
				loadMetadata();
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
			...results.map((row) => headers.map((fieldName) => JSON.stringify(row[fieldName])).join(','))
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

<div class="h-full flex flex-col bg-[#050505]">
	<!-- Toolbar -->
	<div
		class="p-4 border-b border-stone-800 flex justify-between items-center bg-[#0a0a0a]"
	>
		<div class="flex items-center gap-4">
			<div class="p-2.5 bg-rust/10 border border-rust/20 rounded-none industrial-frame">
				<Terminal class="w-5 h-5 text-rust-light" />
			</div>
			<div>
				<h2 class="text-lg font-heading font-black text-slate-100 uppercase tracking-tighter">SQL_CONSOLE_CORE</h2>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest mt-1">Execute raw neural queries against the database</p>
			</div>
		</div>
		<div class="flex gap-4 items-center">
			{#if loadingMetadata}
				<span class="font-jetbrains text-[10px] text-stone-600 flex items-center gap-2 uppercase tracking-widest">
					<RefreshCw class="w-3.5 h-3.5 animate-spin" />
					Syncing_Schema...
				</span>
			{:else if tables.length > 0}
				<span class="font-jetbrains text-[10px] text-stone-500 flex items-center gap-2 uppercase tracking-widest">
					<Database class="w-3.5 h-3.5 text-rust/60" />
					{tables.length} Sectors_Mapped
				</span>
			{/if}
			<button
				onclick={runQuery}
				disabled={loading || !query.trim()}
				class="px-6 py-2.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 disabled:opacity-20 transition-all active:translate-y-px"
			>
				{#if loading}
					<RefreshCw class="w-4 h-4 animate-spin" /> Executing...
				{:else}
					<Play class="w-4 h-4 fill-current" /> Run_Op
				{/if}
			</button>
		</div>
	</div>

	<!-- Editor & Results Split -->
	<div class="flex-1 flex flex-col overflow-hidden">
		<!-- Editor Area -->
		<div
			class="h-1/3 min-h-[200px] p-6 border-b border-stone-800 bg-[#050505] relative"
		>
			<div class="relative w-full h-full">
				<textarea
					bind:this={textareaEl}
					bind:value={query}
					class="w-full h-full bg-stone-950 border border-stone-800 p-6 font-jetbrains text-xs text-stone-200 focus:border-rust outline-none resize-none shadow-inner uppercase tracking-widest"
					placeholder="SELECT * FROM table_name LIMIT 10;"
					onkeydown={handleKeydown}
					oninput={handleInput}
					onblur={handleBlur}
					spellcheck="false"
					autocomplete="off"
				></textarea>

				<!-- Autocomplete Dropdown -->
				{#if showAutocomplete && suggestions.length > 0}
					<div
						class="absolute z-50 bg-[#0a0a0a] border border-stone-800 shadow-2xl overflow-hidden min-w-[240px] max-w-[350px] industrial-frame"
						style="top: {autocompletePosition.top}px; left: {autocompletePosition.left}px;"
						transition:slide={{ duration: 100 }}
					>
						<div class="max-h-[250px] overflow-y-auto custom-scrollbar">
							{#each suggestions as suggestion, i}
								<button
									class="w-full px-4 py-2.5 text-left font-jetbrains text-[10px] font-black uppercase tracking-widest flex items-center gap-3 transition-all {i ===
									selectedIndex
										? 'bg-rust text-white shadow-[0_0_15px_rgba(120,53,15,0.4)]'
										: 'text-stone-500 hover:bg-stone-900 hover:text-stone-300'}"
									onmousedown={() => applySuggestion(suggestion)}
									onmouseenter={() => (selectedIndex = i)}
								>
									{#if SQL_KEYWORDS.includes(suggestion.toUpperCase())}
										<span class="w-1.5 h-1.5 bg-rust-light flex-shrink-0"></span>
									{:else if suggestion.includes('.')}
										<span class="w-1.5 h-1.5 bg-emerald-500 flex-shrink-0"></span>
									{:else if tables.some((t) => t.table === suggestion || `${t.schema}.${t.table}` === suggestion)}
										<span class="w-1.5 h-1.5 bg-amber-500 flex-shrink-0"></span>
									{:else}
										<span class="w-1.5 h-1.5 bg-stone-700 flex-shrink-0"></span>
									{/if}
									<span class="truncate">{suggestion}</span>
								</button>
							{/each}
						</div>
						<div
							class="px-4 py-2 bg-[#050505] border-t border-stone-800 font-jetbrains text-[8px] font-black text-stone-600 flex gap-4 uppercase tracking-[0.2em]"
						>
							<span
								><kbd class="px-1 bg-stone-900 border border-stone-800 rounded-none text-stone-500"
									>↑↓</kbd
								> NAV</span>
							<span
								><kbd class="px-1 bg-stone-900 border border-stone-800 rounded-none text-stone-500"
									>TAB</kbd
								> SELECT</span>
							<span
								><kbd class="px-1 bg-stone-900 border border-stone-800 rounded-none text-stone-500"
									>ESC</kbd
								> EXIT</span>
						</div>
					</div>
				{/if}
			</div>
			<div class="font-jetbrains text-[9px] font-black text-stone-600 mt-3 flex justify-between uppercase tracking-widest">
				<div class="flex gap-6">
					<span class="flex items-center gap-2"><div class="w-1 h-1 bg-stone-800"></div> Cmd/Ctrl + Enter to execute</span>
					<span class="flex items-center gap-2"><div class="w-1 h-1 bg-stone-800"></div> Type to autocomplete</span>
				</div>
				{#if executionTime > 0}
					<span class="text-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.2)]">Time: {executionTime.toFixed(2)}ms</span>
				{/if}
			</div>
		</div>

		<!-- Results Area -->
		<div class="flex-1 flex flex-col overflow-hidden bg-[#050505] relative">
			{#if results.length > 0}
				<div
					class="p-3 border-b border-stone-800 bg-[#0a0a0a] flex justify-between items-center px-6"
				>
					<span
						class="font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest"
						>{results.length} Entities_Detected</span>
					<div class="flex gap-3">
						<button
							onclick={copyToClipboard}
							class="p-2 text-stone-600 hover:text-rust transition-all"
							title="Copy JSON"
						>
							<Copy class="w-4 h-4" />
						</button>
						<button
							onclick={downloadCSV}
							class="p-2 text-stone-600 hover:text-rust transition-all"
							title="Download CSV"
						>
							<Download class="w-4 h-4" />
						</button>
					</div>
				</div>
				<div class="flex-1 overflow-auto custom-scrollbar">
					<table class="w-full text-left font-jetbrains text-[11px] border-collapse">
						<thead
							class="bg-[#0a0a0a] text-stone-500 sticky top-0 z-10 border-b border-stone-800"
						>
							<tr>
								{#each Object.keys(results[0]) as key}
									<th
										class="px-5 py-4 font-black uppercase tracking-widest border-r border-stone-800/30 whitespace-nowrap"
										>{key}</th
									>
								{/each}
							</tr>
						</thead>
						<tbody class="divide-y divide-stone-900">
							{#each results as row}
								<tr class="hover:bg-rust/5 transition-colors group">
									{#each Object.values(row) as val}
										<td
											class="px-5 py-3 text-stone-400 group-hover:text-stone-200 border-r border-stone-800/20 whitespace-nowrap max-w-xs truncate"
											title={String(val)}
										>
											{val === null ? 'NULL' : String(val)}
										</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else if error}
				<div class="flex-1 flex items-center justify-center p-10">
					<div class="bg-red-500/5 border border-red-500/20 p-8 max-w-3xl industrial-frame">
						<div class="font-jetbrains text-xs text-red-500 uppercase tracking-widest font-black mb-4 flex items-center gap-3">
							<div class="w-2 h-2 bg-red-500 animate-pulse"></div>
							EXECUTION_FAULT_DETECTED
						</div>
						<div class="text-red-400/80 font-jetbrains text-[11px] whitespace-pre-wrap leading-relaxed">{error}</div>
					</div>
				</div>
			{:else if !loading}
				<div class="flex-1 flex flex-col items-center justify-center text-stone-700">
					<div class="p-8 bg-stone-900/40 border border-stone-800 industrial-frame mb-6 shadow-inner">
						<Terminal class="w-16 h-16 opacity-10" />
					</div>
					<p class="font-heading font-black text-xs tracking-[0.2em] uppercase">Ready_For_Neural_Input</p>
					<p class="font-jetbrains text-[9px] mt-2 uppercase tracking-widest opacity-60">Enter query and execute to view data sectors</p>
				</div>
			{/if}

			{#if loading}
				<div
					class="absolute inset-0 bg-[#050505]/80 backdrop-blur-sm flex items-center justify-center z-20"
				>
					<div class="flex flex-col items-center gap-4">
						<div
							class="w-10 h-10 border-2 border-rust border-t-transparent rounded-none animate-spin"
						></div>
						<span class="font-heading font-black text-[11px] text-rust uppercase tracking-[0.2em] animate-pulse">Running_Query_Sequence...</span>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>
