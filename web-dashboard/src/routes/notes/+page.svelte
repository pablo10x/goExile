<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { Plus, CheckSquare, Square, Trash2, StickyNote, RefreshCw } from 'lucide-svelte';
	import NoteCard from '$lib/components/notes/NoteCard.svelte';
	import NoteModal from '$lib/components/notes/NoteModal.svelte'; // Import the new modal
	import AIBot from '$lib/components/notes/AIBot.svelte';
	import { notes, todos } from '$lib/stores';
	import type { Note, Todo } from '$lib/stores';

	let loading = true;
	let newTodoContent = '';

	// Modal state
	let showNoteModal = false;
	let editingNote: Note | null = null; // Stores the note being edited, or null for creation

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
		editingNote = null; // Ensure we're in creation mode
		showNoteModal = true;
	}

	function openEditNoteModal(e: CustomEvent<Note>) {
		editingNote = e.detail;
		showNoteModal = true;
	}

	async function handleSaveNote(note: Note) {
		// Determine if it's a new note or an update
		if (note.id === 0) {
			// New note - remove the ID as the backend should generate it
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
			// Update existing note
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

	async function deleteNote(e: CustomEvent<number>) {
		const id = e.detail;
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
				body: JSON.stringify({ content: newTodoContent, done: false })
			});
			if (res.ok) {
				const saved = await res.json();
				todos.update((t) => [...t, saved]);
				newTodoContent = '';
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function toggleTodo(todo: Todo) {
		const updated = { ...todo, done: !todo.done };
		try {
			await fetch(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updated)
			});
			todos.update((t) => t.map((x) => (x.id === todo.id ? updated : x)));
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteTodo(id: number) {
		try {
			await fetch(`/api/todos/${id}`, { method: 'DELETE' });
			todos.update((t) => t.filter((x) => x.id !== id));
		} catch (e) {
			console.error(e);
		}
	}

	onMount(loadData);
</script>

<div class="min-h-full pb-24">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
		<div>
			<h1 class="text-3xl font-bold text-white flex items-center gap-3">
				<StickyNote class="w-8 h-8 text-yellow-400" />
				Notes & Tasks
			</h1>
			<p class="text-slate-400 mt-1">Organize your thoughts and track your progress</p>
		</div>
		<button
			onclick={openCreateNoteModal}
			class="flex items-center gap-2 px-6 py-2.5 bg-gradient-to-r from-yellow-500 to-orange-500 hover:from-yellow-400 hover:to-orange-400 text-white font-semibold rounded-xl shadow-lg shadow-orange-900/20 transition-all transform hover:-translate-y-0.5 active:scale-95"
		>
			<Plus class="w-5 h-5" />
			Add Note
		</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<RefreshCw class="w-10 h-10 text-slate-500 animate-spin" />
		</div>
	{:else}
		<!-- Notes Board -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-12">
			{#each $notes as note (note.id)}
				<div animate:flip={{ duration: 300 }}>
					<NoteCard {note} on:edit={openEditNoteModal} on:delete={deleteNote} />
				</div>
			{/each}
			{#if $notes.length === 0}
				<div class="col-span-full py-12 text-center border-2 border-dashed border-slate-700 rounded-2xl">
					<div class="inline-flex p-4 bg-slate-800/50 rounded-full mb-4">
						<StickyNote class="w-8 h-8 text-slate-500" />
					</div>
					<p class="text-slate-400">No notes yet. Click "Add Note" to get started!</p>
				</div>
			{/if}
		</div>

		<!-- Todo List -->
		<div class="max-w-2xl mx-auto bg-slate-800/40 border border-slate-700/50 rounded-2xl overflow-hidden backdrop-blur-sm">
			<div class="p-6 border-b border-slate-700/50 bg-slate-900/50 flex items-center gap-3">
				<CheckSquare class="w-6 h-6 text-emerald-400" />
				<h2 class="text-xl font-bold text-white">Tasks</h2>
			</div>
			
			<div class="p-6 space-y-4">
				<form
					onsubmit={(e) => {
						e.preventDefault();
						addTodo();
					}}
					class="flex gap-3"
				>
					<input
						type="text"
						bind:value={newTodoContent}
						placeholder="Add a new task..."
						class="flex-1 bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 text-slate-200 placeholder:text-slate-500 focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 outline-none transition-all"
					/>
					<button
						type="submit"
						disabled={!newTodoContent.trim()}
						class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-xl font-medium transition-colors disabled:opacity-50"
					>
						Add
					</button>
				</form>

				<div class="space-y-2">
					{#each $todos as todo (todo.id)}
						<div
							class="flex items-center gap-3 p-3 rounded-xl transition-all group {todo.done
								? 'bg-slate-800/30 opacity-60'
								: 'bg-slate-800/80 hover:bg-slate-700/80'}"
							transition:slide|local
						>
							<button
								onclick={() => toggleTodo(todo)}
								class="text-slate-400 hover:text-emerald-400 transition-colors"
							>
								{#if todo.done}
									<CheckSquare class="w-5 h-5 text-emerald-500" />
								{:else}
									<Square class="w-5 h-5" />
								{/if}
							</button>
							<span class="flex-1 text-sm {todo.done ? 'line-through text-slate-500' : 'text-slate-200'}">
								{todo.content}
							</span>
							<button
								onclick={() => deleteTodo(todo.id)}
								class="text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity p-1"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</div>
					{/each}
					{#if $todos.length === 0}
						<p class="text-center text-slate-500 text-sm py-4">No tasks pending.</p>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<AIBot />

	<!-- Note Modal -->
	<NoteModal bind:isOpen={showNoteModal} note={editingNote} onSave={handleSaveNote} onClose={() => (showNoteModal = false)} />
</div>
