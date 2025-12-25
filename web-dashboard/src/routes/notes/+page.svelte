<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { Plus, CheckSquare, Square, Trash2, StickyNote, RefreshCw, Search, X } from 'lucide-svelte';
	import NoteCard from '$lib/components/notes/NoteCard.svelte';
	import TaskItem from '$lib/components/notes/TaskItem.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';
	import AIBot from '$lib/components/notes/AIBot.svelte';
	import { notes, todos } from '$lib/stores';
	import type { Note, Todo } from '$lib/stores';

	let loading = $state(true);
	let newTodoContent = $state('');
	let newTodoDeadline = $state('');
	let searchQuery = $state('');
	// Modal state
	let showNoteModal = $state(false);
	let editingNote = $state<Note | null>(null);

	async function loadData() {
		loading = true;
		try {
			const [nRes, tRes] = await Promise.all([fetch('/api/notes'), fetch('/api/todos')]);
			if (nRes.ok) notes.set(await nRes.json());
			if (tRes.ok) todos.set(await tRes.json());
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function openCreateNoteModal() {
		editingNote = null;
		showNoteModal = true;
	}

	function openEditNoteModal(note: Note) {
		editingNote = note;
		showNoteModal = true;
	}

	async function handleSaveNote(note: Note) {
		if (note.id === 0) {
			const { id, ...noteWithoutId } = note;
			try {
				const res = await fetch('/api/notes', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(noteWithoutId)
				});
				if (res.ok) {
					const saved = await res.json();
					notes.update((n) => [saved, ...n]);
				}
			} catch (e) {
				console.error(e);
			}
		} else {
			try {
				await fetch(`/api/notes/${note.id}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(note)
				});
				notes.update((n) => n.map((x) => (x.id === note.id ? note : x)));
			} catch (err) {
				console.error(err);
			}
		}
	}

	async function deleteNote(id: number) {
		try {
			await fetch(`/api/notes/${id}`, { method: 'DELETE' });
			notes.update((n) => n.filter((x) => x.id !== id));
		} catch (err) {
			console.error(err);
		}
	}

	async function addTodo() {
		if (!newTodoContent.trim()) return;
		try {
			const res = await fetch('/api/todos', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					content: newTodoContent,
					done: false,
					in_progress: false,
					deadline: newTodoDeadline || null
				})
			});
			if (res.ok) {
				const saved = await res.json();
				todos.update((t) => [...t, saved]);
				newTodoContent = '';
				newTodoDeadline = '';
			}
		} catch (e) {
			console.error(e);
		}
	}
	async function deleteTodo(id: number) {
		try {
			const res = await fetch(`/api/todos/${id}`, { method: 'DELETE' });
			if (res.ok) {
				todos.update((all) => {
					const removeFromTree = (nodes: Todo[]): Todo[] => {
						return nodes
							.filter((n) => n.id !== id)
							.map((n) => ({
								...n,
								sub_tasks: n.sub_tasks ? removeFromTree(n.sub_tasks) : []
							}));
					};
					return removeFromTree(all);
				});
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function toggleTodo(todo: Todo) {
		const updated = {
			...todo,
			done: !todo.done,
			in_progress: !todo.done ? false : todo.in_progress
		};
		try {
			await fetch(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			todos.update((all) => {
				const updateTree = (nodes: Todo[]): Todo[] => {
					return nodes.map((n) => {
						if (n.id === todo.id) return updated;
						if (n.sub_tasks) return { ...n, sub_tasks: updateTree(n.sub_tasks) };
						return n;
					});
				};
				return updateTree(all);
			});
		} catch (e) {
			console.error(e);
		}
	}

	async function toggleProgress(todo: Todo) {
		const updated = { ...todo, in_progress: !todo.in_progress, done: false };
		try {
			await fetch(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			todos.update((all) => {
				const updateTree = (nodes: Todo[]): Todo[] => {
					return nodes.map((n) => {
						if (n.id === todo.id) return updated;
						if (n.sub_tasks) return { ...n, sub_tasks: updateTree(n.sub_tasks) };
						return n;
					});
				};
				return updateTree(all);
			});
		} catch (e) {
			console.error(e);
		}
	}

	let filteredNotes = $derived(
		($notes || []).filter(
			(n) =>
				searchQuery === '' ||
				n.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				n.content.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	let pendingTodos = $derived(($todos || []).filter((t) => !t.done && !t.parent_id));
	let completedTodos = $derived(($todos || []).filter((t) => t.done && !t.parent_id));
	onMount(loadData);
</script>

<div class="w-full h-[calc(100vh-140px)] md:h-[calc(100vh-160px)] flex flex-col lg:flex-row gap-8 overflow-hidden relative border border-stone-800 bg-[#050505] shadow-2xl">
	<!-- Left Column: Tasks -->
	<div
		class="w-full lg:w-[450px] flex-shrink-0 flex flex-col bg-[#0a0a0a]/60 border-r border-stone-800 overflow-hidden backdrop-blur-xl h-[500px] lg:h-full"
	>
		<!-- Tasks Header -->
		<div
			class="p-6 border-b border-stone-800 bg-[#0a0a0a] flex justify-between items-center shrink-0"
		>
			<h2 class="text-xl font-heading font-black text-white flex items-center gap-3 uppercase tracking-tighter">
				<CheckSquare class="w-6 h-6 text-emerald-500" />
				TASK_REGISTRY
			</h2>
			<div class="flex items-center gap-3">
				<span
					class="px-3 py-1 bg-stone-900 border border-stone-800 font-jetbrains text-[10px] font-black text-stone-500 uppercase tracking-widest shadow-inner"
				>
					{pendingTodos.length} PENDING
				</span>
			</div>
		</div>

		<div class="p-6 border-b border-stone-800 bg-stone-900/20 shrink-0">
			<form
				onsubmit={(e) => {
					e.preventDefault();
					addTodo();
				}}
				class="flex flex-col gap-4"
			>
				<div class="flex gap-3">
					<input
						type="text"
						bind:value={newTodoContent}
						placeholder="INIT_NEW_DIRECTIVE..."
						class="flex-1 bg-black border border-stone-800 px-4 py-3 font-jetbrains text-xs text-white placeholder:text-stone-800 focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
					/>
					<button
						type="submit"
						disabled={!newTodoContent.trim()}
						class="p-3 bg-rust hover:bg-rust-light text-white transition-all disabled:opacity-20 shadow-lg shadow-rust/20 active:translate-y-px"
					>
						<Plus class="w-5 h-5" />
					</button>
				</div>
				<div class="flex items-center gap-4">
					<label
						for="todo-deadline"
						class="font-jetbrains text-[9px] font-black text-stone-600 uppercase tracking-[0.3em]">EXECUTION_DEADLINE:</label
					>
					<input
						type="date"
						id="todo-deadline"
						bind:value={newTodoDeadline}
						class="bg-black border border-stone-800 px-3 py-1.5 font-jetbrains text-[10px] text-rust-light outline-none focus:border-rust transition-all uppercase tracking-widest shadow-inner"
					/>
					{#if newTodoDeadline}
						<button
							type="button"
							onclick={() => (newTodoDeadline = '')}
							class="font-jetbrains text-[9px] font-black text-red-500 hover:text-white transition-colors uppercase tracking-widest"
						>
							[CLEAR]
						</button>
					{/if}
				</div>
			</form>
		</div>

		<!-- Task List -->
		<div class="flex-1 overflow-y-auto p-6 space-y-3 custom-scrollbar bg-black/20 relative">
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
			{#if loading}
				<div class="flex flex-col items-center justify-center py-20 gap-4">
					<div class="w-8 h-8 border-2 border-rust border-t-transparent rounded-none animate-spin"></div>
					<span class="font-jetbrains text-[9px] font-black text-stone-600 uppercase tracking-[0.4em]">Syncing_Buffer...</span>
				</div>
			{:else}
				{#if pendingTodos.length > 0}
					<div class="space-y-3 relative z-10">
						{#each pendingTodos as todo (todo.id)}
							<div animate:flip={{ duration: 300 }}>
								<TaskItem
									{todo}
									onToggle={toggleTodo}
									onDelete={deleteTodo}
									onToggleProgress={toggleProgress}
								/>
							</div>
						{/each}
					</div>
				{/if}

				{#if completedTodos.length > 0}
					{#if pendingTodos.length > 0}
						<div class="h-px bg-stone-800 my-8 relative z-10">
							<div class="absolute inset-0 flex items-center justify-center">
								<div class="bg-[#0a0a0a] px-4 font-jetbrains text-[8px] font-black text-stone-700 tracking-[0.5em] uppercase">DEFRAGMENTED_UNITS</div>
							</div>
						</div>
					{/if}
					<div class="space-y-3 opacity-40 relative z-10">
						{#each completedTodos as todo (todo.id)}
							<div animate:flip={{ duration: 300 }}>
								<TaskItem
									{todo}
									onToggle={toggleTodo}
									onDelete={deleteTodo}
									onToggleProgress={toggleProgress}
								/>
							</div>
						{/each}
					</div>
				{/if}

				{#if $todos.length === 0}
					<div class="flex flex-col items-center justify-center py-20 text-stone-800 gap-4">
						<div class="p-6 border border-dashed border-stone-800 industrial-frame">
							<CheckSquare class="w-10 h-10 opacity-10" />
						</div>
						<p class="font-jetbrains text-[10px] font-black uppercase tracking-[0.3em]">Protocol_Buffer_Empty</p>
					</div>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Right Column: Notes -->
	<div class="flex-1 flex flex-col min-h-0 bg-transparent p-6 lg:p-10">
		<!-- Notes Toolbar -->
		<div
			class="flex flex-col xl:flex-row justify-between items-start xl:items-center gap-8 mb-12 shrink-0"
		>
			<div class="flex items-center gap-6">
				<div class="p-4 bg-stone-900 border border-stone-800 shadow-xl industrial-frame">
					<StickyNote class="w-10 h-10 text-rust" />
				</div>
				<div>
					<h1 class="text-4xl font-heading font-black text-white uppercase tracking-tighter">
						NEURAL_NOTATIONS
					</h1>
					<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-widest font-black mt-2">
						Buffer asynchronous concepts and tactical intelligence
					</p>
				</div>
			</div>

			<div class="flex items-center gap-4 w-full xl:w-auto">
				<div class="relative flex-1 xl:w-96 group">
					<Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-stone-600 group-focus-within:text-rust transition-colors" />
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="FILTER_SIGNAL_LOGS..."
						class="w-full bg-stone-950 border border-stone-800 py-3.5 pl-14 pr-4 font-jetbrains text-xs text-stone-300 focus:border-rust outline-none transition-all uppercase tracking-widest shadow-inner"
					/>
				</div>
				<button
					onclick={openCreateNoteModal}
					class="flex items-center gap-3 px-10 py-3.5 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-widest shadow-lg shadow-rust/20 transition-all active:translate-y-px"
				>					<Plus class="w-5 h-5" />
					<span>New_Entry</span>
				</button>
			</div>
		</div>

		<!-- Notes Grid -->
		<div class="flex-1 overflow-y-auto custom-scrollbar pb-32">
			<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.01] pointer-events-none"></div>
			<!-- pb-20 for AI bot clearance -->
			{#if loading}
				<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 3xl:grid-cols-5 gap-8">
					{#each [1, 2, 3, 4, 5] as _}
						<div
							class="h-72 bg-stone-900/40 border border-stone-800 animate-pulse industrial-frame"
						></div>
					{/each}
				</div>
			{:else if filteredNotes.length > 0}
				<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 3xl:grid-cols-5 gap-8">
					{#each filteredNotes as note (note.id)}
						<div animate:flip={{ duration: 300 }}>
							<NoteCard {note} onEdit={openEditNoteModal} onDelete={deleteNote} />
						</div>
					{/each}
				</div>
			{:else}
				<div
					class="h-[500px] flex flex-col items-center justify-center border-2 border-dashed border-stone-800 bg-stone-950/20 industrial-frame"
				>
					<div class="p-10 bg-stone-900/40 border border-stone-800 mb-8 group">
						<StickyNote class="w-16 h-16 text-stone-800 group-hover:text-rust transition-colors duration-500" />
					</div>
					<p class="font-heading font-black text-base text-stone-600 uppercase tracking-[0.4em]">
						{searchQuery ? 'Zero_Signals_Matching_Query' : 'Registry_Clean_Initialize_Buffer'}
					</p>
				</div>
			{/if}
		</div>
	</div>

	<AIBot />

	<!-- Note Modal -->
	<NoteModal
		bind:isOpen={showNoteModal}
		note={editingNote}
		onSave={handleSaveNote}
		onClose={() => (showNoteModal = false)}
	/>
</div>

<style>
	/* Custom Scrollbar */
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
		height: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: rgba(15, 23, 42, 0.1);
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(71, 85, 105, 0.4);
		border-radius: 3px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(71, 85, 105, 0.6);
	}
</style>