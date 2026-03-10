<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { Plus, CheckSquare, StickyNote, Search, X, Clock, Calendar } from 'lucide-svelte';
	import NoteCard from '$lib/components/notes/NoteCard.svelte';
	import TaskItem from '$lib/components/notes/TaskItem.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte';
	import AIBot from '$lib/components/notes/AIBot.svelte';
	import { notes, todos } from '$lib/stores.svelte';
	import type { Note, Todo } from '$lib/stores.svelte';
	import Card from '$lib/components/theme/Card.svelte';
	import Button from '$lib/components/Button.svelte';

	let loading = $state(true);
	let newTodoContent = $state('');
	let newTodoDeadline = $state('');
	let searchQuery = $state('');
	let showNoteModal = $state(false);
	let editingNote = $state<Note | null>(null);

	async function loadData() {
		loading = true;
		try {
			const [nRes, tRes] = await Promise.all([apiFetch('/api/notes'), apiFetch('/api/todos')]);
			if (nRes.ok) notes.set(await nRes.json());
			if (tRes.ok) todos.set(await tRes.json());
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function handleSaveNote(note: Note) {
		if (note.id === 0) {
			const { id, ...rest } = note;
			try {
				const res = await apiFetch('/api/notes', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(rest)
				});
				if (res.ok) {
					const savedNote = await res.json();
					notes.update((n) => [savedNote, ...n]);
				}
			} catch (e) {
				console.error(e);
			}
		} else {
			try {
				await apiFetch(`/api/notes/${note.id}`, {
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
		if (!confirm('Are you sure you want to delete this note?')) return;
		try {
			const res = await apiFetch(`/api/notes/${id}`, { method: 'DELETE' });
			if (res.ok) {
				notes.update((n) => n.filter((x) => x.id !== id));
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function addTodo() {
		if (!newTodoContent.trim()) return;
		try {
			const res = await apiFetch('/api/todos', {
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
				const savedTodo = await res.json();
				todos.update((t) => [...t, savedTodo]);
				newTodoContent = '';
				newTodoDeadline = '';
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteTodo(id: number) {
		try {
			const res = await apiFetch(`/api/todos/${id}`, { method: 'DELETE' });
			if (res.ok) {
				todos.update((all) => {
					const remove = (nodes: Todo[]): Todo[] =>
						nodes
							.filter((n) => n.id !== id)
							.map((n) => ({ ...n, sub_tasks: n.sub_tasks ? remove(n.sub_tasks) : [] }));
					return remove(all);
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
			await apiFetch(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			todos.update((all) => {
				const update = (nodes: Todo[]): Todo[] =>
					nodes.map((n) =>
						n.id === todo.id ? updated : n.sub_tasks ? { ...n, sub_tasks: update(n.sub_tasks) } : n
					);
				return update(all);
			});
		} catch (e) {
			console.error(e);
		}
	}

	async function toggleProgress(todo: Todo) {
		const updated = { ...todo, in_progress: !todo.in_progress, done: false };
		try {
			await apiFetch(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			todos.update((all) => {
				const update = (nodes: Todo[]): Todo[] =>
					nodes.map((n) =>
						n.id === todo.id ? updated : n.sub_tasks ? { ...n, sub_tasks: update(n.sub_tasks) } : n
					);
				return update(all);
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

<div
	class="w-full min-h-[calc(100vh-8rem)] flex flex-col lg:flex-row gap-8 overflow-hidden relative font-sans"
>
	<!-- Left Column: Tasks -->
	<div
		class="w-full lg:w-[450px] flex-shrink-0 flex flex-col bg-slate-900/40 border border-white/5 rounded-3xl overflow-hidden backdrop-blur-xl h-[600px] lg:h-auto shadow-2xl"
	>
		<div
			class="p-8 border-b border-white/5 bg-slate-950/40 flex justify-between items-center shrink-0"
		>
			<div class="flex items-center gap-4">
				<div class="p-2.5 bg-sky-500/10 rounded-xl border border-sky-500/20">
					<CheckSquare class="w-5 h-5 text-sky-400" />
				</div>
				<h2 class="text-lg font-bold text-white tracking-tight">Task List</h2>
			</div>
			<span
				class="px-3 py-1 bg-white/5 border border-white/10 rounded-lg text-[10px] font-bold text-slate-400 uppercase tracking-widest"
				>{pendingTodos.length} Pending</span
			>
		</div>

		<div class="p-8 border-b border-white/5 bg-slate-900/20 shrink-0">
			<form
				onsubmit={(e) => {
					e.preventDefault();
					addTodo();
				}}
				class="space-y-4"
			>
				<div class="flex gap-2">
					<input
						type="text"
						bind:value={newTodoContent}
						placeholder="Task description..."
						class="flex-1 bg-slate-950 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:border-sky-500/50 outline-none transition-all shadow-inner"
					/>
					<Button
						type="submit"
						disabled={!newTodoContent.trim()}
						variant="primary"
						size="md"
						icon="ph:plus-bold"
					/>
				</div>
				<div class="flex items-center gap-4 pl-1">
					<label
						for="todo-deadline"
						class="text-[10px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2"
						><Calendar size={12} /> Deadline:</label
					>
					<input
						type="date"
						id="todo-deadline"
						bind:value={newTodoDeadline}
						class="bg-slate-950 border border-white/5 rounded-lg px-3 py-1.5 text-[10px] text-sky-400 outline-none focus:border-sky-500/50 transition-all uppercase font-semibold"
					/>
					{#if newTodoDeadline}<button
							type="button"
							onclick={() => (newTodoDeadline = '')}
							class="text-[10px] font-bold text-rose-500 hover:text-white transition-colors uppercase"
							>[Clear]</button
						>{/if}
				</div>
			</form>
		</div>

		<div class="flex-1 overflow-y-auto p-6 space-y-3 no-scrollbar relative">
			{#if loading}
				<div class="flex flex-col items-center justify-center py-20 gap-4">
					<div
						class="w-8 h-8 border-4 border-sky-500/20 border-t-sky-500 rounded-full animate-spin"
					></div>
					<span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Loading...</span>
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
					{#if pendingTodos.length > 0}<div class="flex items-center gap-4 my-8 relative z-10">
							<div class="h-px flex-1 bg-white/5"></div>
							<span class="text-[10px] font-bold text-slate-600 uppercase tracking-widest"
								>Completed</span
							>
							<div class="h-px flex-1 bg-white/5"></div>
						</div>{/if}
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
					<div class="flex flex-col items-center justify-center py-20 text-slate-700 gap-4">
						<div class="p-6 border border-dashed border-white/10 rounded-3xl">
							<CheckSquare class="w-10 h-10 opacity-10" />
						</div>
						<p class="text-[10px] font-bold uppercase tracking-widest">No tasks found</p>
					</div>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Right Column: Notes -->
	<div class="flex-1 flex flex-col min-h-0 bg-transparent">
		<div
			class="flex flex-col xl:flex-row justify-between items-start xl:items-center gap-8 mb-10 shrink-0"
		>
			<div class="flex items-center gap-6">
				<div
					class="p-4 bg-slate-800/40 border border-white/5 rounded-3xl shadow-xl backdrop-blur-md group"
				>
					<StickyNote
						class="w-10 h-10 text-sky-400 group-hover:scale-110 transition-transform duration-500"
					/>
				</div>
				<div>
					<h1 class="text-4xl font-bold text-white tracking-tight">Documentation</h1>
					<p class="text-xs font-medium text-slate-500 uppercase tracking-widest mt-2">
						Technical notes & system resources
					</p>
				</div>
			</div>

			<div class="flex items-center gap-4 w-full xl:w-auto">
				<div class="relative flex-1 xl:w-96 group">
					<Search
						class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-sky-400 transition-colors"
					/>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search documentation..."
						class="w-full bg-slate-900/50 border border-white/5 rounded-2xl py-3.5 pl-14 pr-4 text-sm text-slate-200 focus:border-sky-500/50 outline-none transition-all shadow-inner"
					/>
				</div>
				<Button
					onclick={() => {
						editingNote = null;
						showNoteModal = true;
					}}
					variant="primary"
					size="lg"
					icon="ph:plus-bold">Add Note</Button
				>
			</div>
		</div>

		<div class="flex-1 overflow-y-auto no-scrollbar pb-32">
			{#if loading}
				<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
					{#each [1, 2, 3, 4, 5, 6] as _}<div
							class="h-64 bg-slate-800/20 border border-white/5 rounded-[2rem] animate-pulse"
						></div>{/each}
				</div>
			{:else if filteredNotes.length > 0}
				<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
					{#each filteredNotes as note (note.id)}
						<div animate:flip={{ duration: 300 }}>
							<NoteCard
								{note}
								onEdit={(n) => {
									editingNote = n;
									showNoteModal = true;
								}}
								onDelete={(id) => deleteNote(id)}
							/>
						</div>
					{/each}
				</div>
			{:else}
				<div
					class="h-[500px] flex flex-col items-center justify-center border-2 border-dashed border-white/5 rounded-[3rem] bg-slate-950/20"
				>
					<div
						class="p-10 bg-slate-900/40 border border-white/5 rounded-[2.5rem] mb-8 group shadow-2xl"
					>
						<StickyNote
							class="w-16 h-16 text-slate-800 group-hover:text-sky-400 transition-all duration-500"
						/>
					</div>
					<p class="text-sm font-bold text-slate-500 uppercase tracking-[0.3em] text-center px-6">
						{searchQuery ? 'No results match your search' : 'No documentation found'}
					</p>
				</div>
			{/if}
		</div>
	</div>

	<AIBot />
	<NoteModal
		bind:isOpen={showNoteModal}
		note={editingNote}
		onSave={handleSaveNote}
		onClose={() => (showNoteModal = false)}
	/>
</div>

<style>
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
</style>
