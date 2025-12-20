<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { Plus, CheckSquare, Square, Trash2, StickyNote, RefreshCw, Search } from 'lucide-svelte';
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
			} catch (e) { console.error(e); }
		} else {
			try {
				await fetch(`/api/notes/${note.id}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(note)
				});
				notes.update((n) => n.map((x) => (x.id === note.id ? note : x)));
			} catch (err) { console.error(err); }
		}
	}

	async function deleteNote(id: number) {
		try {
			await fetch(`/api/notes/${id}`, { method: 'DELETE' });
			notes.update((n) => n.filter((x) => x.id !== id));
		} catch (err) { console.error(err); }
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
	                } catch (e) { console.error(e); }
	        }
	async function deleteTodo(id: number) {
		try {
			const res = await fetch(`/api/todos/${id}`, { method: 'DELETE' });
			if (res.ok) {
				todos.update((all) => {
					const removeFromTree = (nodes: Todo[]): Todo[] => {
						return nodes.filter((n) => n.id !== id).map((n) => ({
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

    	let filteredNotes = $derived(($notes || []).filter((n) => 
    		searchQuery === '' || 
    		n.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
    		n.content.toLowerCase().includes(searchQuery.toLowerCase())
    	));
    
    	let pendingTodos = $derived(($todos || []).filter((t) => !t.done && !t.parent_id));
    	let completedTodos = $derived(($todos || []).filter((t) => t.done && !t.parent_id));
	onMount(loadData);
</script>

<div class="h-full flex flex-col lg:flex-row gap-6 p-4 sm:p-6 overflow-hidden relative">
    
    <!-- Left Column: Tasks (Fixed width on desktop, stacked on mobile) -->
    <div class="w-full lg:w-96 flex-shrink-0 flex flex-col bg-slate-900/50 border border-slate-700/50 rounded-2xl overflow-hidden backdrop-blur-xl shadow-xl h-[500px] lg:h-auto">
        <!-- Tasks Header -->
        <div class="p-5 border-b border-slate-700/50 bg-slate-800/30 flex justify-between items-center shrink-0">
            <h2 class="text-lg font-bold text-white flex items-center gap-2">
                <CheckSquare class="w-5 h-5 text-emerald-400" />
                Tasks
            </h2>
            <div class="flex items-center gap-2">
                <span class="px-2 py-0.5 rounded-md bg-slate-800 border border-slate-700 text-xs font-mono text-slate-400">
                    {pendingTodos.length} Pending
                </span>
            </div>
        </div>

        <div class="p-4 border-b border-slate-700/50 bg-slate-900/30 shrink-0">
            <form
                onsubmit={(e) => { e.preventDefault(); addTodo(); }}
                class="flex flex-col gap-2"
            >
                <div class="flex gap-2">
                    <input
                        type="text"
                        bind:value={newTodoContent}
                        placeholder="New task..."
                        class="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-200 placeholder:text-slate-500 focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 outline-none transition-all"
                    />
                    <button
                        type="submit"
                        disabled={!newTodoContent.trim()}
                        class="p-2 bg-slate-700 hover:bg-emerald-600 text-white rounded-lg transition-colors disabled:opacity-50"
                    >
                        <Plus class="w-4 h-4" />
                    </button>
                </div>
                <div class="flex items-center gap-2">
                    <label for="todo-deadline" class="text-[10px] uppercase font-bold text-slate-500 tracking-wider">Deadline:</label>
                    <input
                        type="date"
                        id="todo-deadline"
                        bind:value={newTodoDeadline}
                        class="bg-slate-800 border border-slate-700 rounded px-2 py-1 text-[10px] text-slate-300 outline-none focus:border-slate-500 transition-colors"
                    />
                    {#if newTodoDeadline}
                        <button 
                            type="button" 
                            onclick={() => newTodoDeadline = ''}
                            class="text-[10px] text-red-400 hover:text-red-300"
                        >
                            Clear
                        </button>
                    {/if}
                </div>
            </form>
        </div>

        <!-- Task List -->
        <div class="flex-1 overflow-y-auto p-4 space-y-2 custom-scrollbar">
            {#if loading}
                <div class="flex justify-center py-4"><RefreshCw class="w-5 h-5 animate-spin text-slate-600"/></div>
            {:else}
                {#if pendingTodos.length > 0}
                    <div class="space-y-2">
                        {#each pendingTodos as todo (todo.id)}
                            <div animate:flip={{ duration: 300 }}>
                                								<TaskItem {todo} onToggle={toggleTodo} onDelete={deleteTodo} onToggleProgress={toggleProgress} />                            </div>
                        {/each}
                    </div>
                {/if}

                {#if completedTodos.length > 0}
                    {#if pendingTodos.length > 0}
                        <div class="border-t border-slate-800/50 my-4"></div>
                    {/if}
                    <div class="text-xs font-bold text-slate-600 uppercase tracking-wider mb-2 px-1">Completed</div>
                    <div class="space-y-2 opacity-60">
                        {#each completedTodos as todo (todo.id)}
                            <div animate:flip={{ duration: 300 }}>
                                								<TaskItem {todo} onToggle={toggleTodo} onDelete={deleteTodo} onToggleProgress={toggleProgress} />                            </div>
                        {/each}
                    </div>
                {/if}

                {#if $todos.length === 0}
                    <div class="text-center py-8">
                        <div class="inline-flex p-3 rounded-full bg-slate-800/50 mb-3">
                            <CheckSquare class="w-6 h-6 text-slate-600" />
                        </div>
                        <p class="text-sm text-slate-500">No tasks yet.</p>
                    </div>
                {/if}
            {/if}
        </div>
    </div>

    <!-- Right Column: Notes -->
    <div class="flex-1 flex flex-col min-h-0 bg-transparent">
        <!-- Notes Toolbar -->
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6 shrink-0">
            <div>
                <h1 class="text-2xl font-bold text-white flex items-center gap-2">
                    <StickyNote class="w-6 h-6 text-yellow-400" />
                    Notes
                </h1>
                <p class="text-slate-400 text-sm mt-0.5">Capture ideas and important info</p>
            </div>

            <div class="flex items-center gap-3 w-full sm:w-auto">
                <div class="relative flex-1 sm:w-64">
                    <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                    <input 
                        type="text" 
                        bind:value={searchQuery}
                        placeholder="Search notes..." 
                        class="w-full bg-slate-900/50 border border-slate-700/50 rounded-xl pl-9 pr-4 py-2 text-sm text-slate-200 focus:ring-2 focus:ring-blue-500/50 outline-none backdrop-blur-sm"
                    />
                </div>
                <button
                    onclick={openCreateNoteModal}
                    class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold rounded-xl shadow-lg shadow-blue-900/20 transition-all hover:-translate-y-0.5"
                >
                    <Plus class="w-4 h-4" />
                    <span class="hidden sm:inline">New Note</span>
                </button>
            </div>
        </div>

        <!-- Notes Grid -->
        <div class="flex-1 overflow-y-auto custom-scrollbar pb-20"> <!-- pb-20 for AI bot clearance -->
            {#if loading}
                <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                    {#each [1,2,3] as _}
                        <div class="h-48 bg-slate-800/30 rounded-xl animate-pulse border border-slate-700/30"></div>
                    {/each}
                </div>
            {:else if filteredNotes.length > 0}
                <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                    {#each filteredNotes as note (note.id)}
                        <div animate:flip={{ duration: 300 }}>
                            <NoteCard 
                                {note} 
                                onEdit={openEditNoteModal} 
                                onDelete={deleteNote} 
                            />
                        </div>
                    {/each}
                </div>
            {:else}
                <div class="h-64 flex flex-col items-center justify-center border-2 border-dashed border-slate-800 rounded-2xl bg-slate-900/20">
                    <div class="p-4 bg-slate-800/50 rounded-full mb-3">
                        <StickyNote class="w-8 h-8 text-slate-600" />
                    </div>
                    <p class="text-slate-500 font-medium">
                        {searchQuery ? 'No matching notes found.' : 'No notes yet. Create one!'}
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