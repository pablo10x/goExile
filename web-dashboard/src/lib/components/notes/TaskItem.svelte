<script lang="ts">
	import {
		CheckSquare,
		Square,
		Trash2,
		Clock,
		Calendar,
		ChevronDown,
		ChevronRight,
		Plus,
		MessageSquare,
		Send,
		User
	} from 'lucide-svelte';
	import { slide, fade } from 'svelte/transition';
	import type { Todo, TodoComment } from '$lib/stores';
	import { todos } from '$lib/stores';
	import TaskItem from './TaskItem.svelte';

	let { todo, onToggle, onDelete, onToggleProgress } = $props<{
		todo: Todo;
		onToggle: (todo: Todo) => void;
		onDelete: (id: number) => void;
		onToggleProgress: (todo: Todo) => void;
	}>();

	let expanded = $state(false);
	let showComments = $state(false);
	let showAddSubTask = $state(false);
	let newSubTaskContent = $state('');
	let newCommentContent = $state('');

	function formatTime(dateStr: string) {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric'
		});
	}

	async function addSubTask() {
		if (!newSubTaskContent.trim()) return;
		try {
			const res = await fetch('/api/todos', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					content: newSubTaskContent,
					parent_id: todo.id,
					done: false,
					in_progress: false
				})
			});
			if (res.ok) {
				const saved = await res.json();
				todos.update((all) => {
					// Need to find and update the parent in the tree
					const updateTree = (nodes: Todo[]): Todo[] => {
						return nodes.map((n) => {
							if (n.id === todo.id) {
								return { ...n, sub_tasks: [...(n.sub_tasks || []), saved] };
							}
							if (n.sub_tasks) {
								return { ...n, sub_tasks: updateTree(n.sub_tasks) };
							}
							return n;
						});
					};
					return updateTree(all);
				});
				newSubTaskContent = '';
				showAddSubTask = false;
				expanded = true;
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function addComment() {
		if (!newCommentContent.trim()) return;
		try {
			const res = await fetch(`/api/todos/${todo.id}/comments`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ content: newCommentContent, author: 'Admin' })
			});
			if (res.ok) {
				const saved = await res.json();
				todos.update((all) => {
					const updateTree = (nodes: Todo[]): Todo[] => {
						return nodes.map((n) => {
							if (n.id === todo.id) {
								return { ...n, comments: [...(n.comments || []), saved] };
							}
							if (n.sub_tasks) {
								return { ...n, sub_tasks: updateTree(n.sub_tasks) };
							}
							return n;
						});
					};
					return updateTree(all);
				});
				newCommentContent = '';
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteComment(commentId: number) {
		try {
			const res = await fetch(`/api/todos/comments/${commentId}`, { method: 'DELETE' });
			if (res.ok) {
				todos.update((all) => {
					const updateTree = (nodes: Todo[]): Todo[] => {
						return nodes.map((n) => {
							if (n.id === todo.id) {
								return { ...n, comments: (n.comments || []).filter((c) => c.id !== commentId) };
							}
							if (n.sub_tasks) {
								return { ...n, sub_tasks: updateTree(n.sub_tasks) };
							}
							return n;
						});
					};
					return updateTree(all);
				});
			}
		} catch (e) {
			console.error(e);
		}
	}
</script>

<div
	class="flex flex-col gap-1 p-2 rounded-xl transition-all group border border-transparent {todo.done
		? 'bg-slate-800/20 opacity-70'
		: 'bg-slate-800/60 hover:bg-slate-800/80 hover:border-slate-300 dark:border-slate-700'}"
	transition:slide|local
>
	<div class="flex items-center gap-2">
		{#if todo.sub_tasks && todo.sub_tasks.length > 0}
			<button
				onclick={() => (expanded = !expanded)}
				class="p-1 rounded hover:bg-slate-700 text-slate-500 transition-colors"
			>
				{#if expanded}
					<ChevronDown class="w-3 h-3" />
				{:else}
					<ChevronRight class="w-3 h-3" />
				{/if}
			</button>
		{:else}
			<div class="w-5"></div>
		{/if}

		<button
			onclick={() => onToggle(todo)}
			class="text-slate-500 dark:text-slate-400 hover:text-emerald-400 transition-colors shrink-0"
		>
			{#if todo.done}
				<CheckSquare class="w-5 h-5 text-emerald-500" />
			{:else}
				<Square class="w-5 h-5" />
			{/if}
		</button>

		<div class="flex-1 min-w-0">
			<div class="flex items-center gap-2">
				<span
					class="text-sm font-medium truncate {todo.done
						? 'line-through text-slate-500'
						: 'text-slate-800 dark:text-slate-200'}"
				>
					{todo.content}
				</span>
				{#if !todo.done && todo.in_progress}
					<span
																		class="px-1.5 py-0.5 rounded-full bg-rust/20 text-rust-light text-[9px] font-bold uppercase tracking-wider animate-pulse"					>
						In Progress
					</span>
				{/if}
			</div>

			<div class="flex items-center gap-3 mt-0.5">
				<div class="flex items-center gap-1 text-[9px] text-slate-500">
					<Clock class="w-2.5 h-2.5" />
					<span>{formatTime(todo.created_at)}</span>
				</div>

				{#if todo.deadline}
					<div
						class="flex items-center gap-1 text-[9px] {new Date(todo.deadline) < new Date() &&
						!todo.done
							? 'text-red-400'
							: 'text-slate-500'}"
					>
						<Calendar class="w-2.5 h-2.5" />
						<span>{formatTime(todo.deadline)}</span>
					</div>
				{/if}

				{#if todo.comments && todo.comments.length > 0}
					<button
						onclick={() => (showComments = !showComments)}
						class="flex items-center gap-1 text-[9px] text-indigo-400 hover:text-indigo-300"
					>
						<MessageSquare class="w-2.5 h-2.5" />
						<span>{todo.comments.length}</span>
					</button>
				{/if}
			</div>
		</div>

		<div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
			{#if !todo.done}
				<button
					onclick={() => (showAddSubTask = !showAddSubTask)}
					class="p-1.5 rounded hover:bg-slate-700 text-slate-500 hover:text-emerald-400 transition-colors"
					title="Add sub-task"
				>
					<Plus class="w-3.5 h-3.5" />
				</button>
				<button
					onclick={() => (showComments = !showComments)}
					class="p-1.5 rounded hover:bg-slate-700 text-slate-500 hover:text-indigo-400 transition-colors"
					title="Comments"
				>
					<MessageSquare class="w-3.5 h-3.5" />
				</button>
				<button
					onclick={() => onToggleProgress(todo)}
					class="p-1.5 rounded hover:bg-slate-700 transition-colors {todo.in_progress
						? 'text-blue-400'
						: 'text-slate-500'}"
					title={todo.in_progress ? 'Pause progress' : 'Start progress'}
				>
					<Clock class="w-3.5 h-3.5" />
				</button>
			{/if}
			<button
				onclick={() => onDelete(todo.id)}
				class="text-slate-500 hover:text-red-400 p-1.5 rounded hover:bg-slate-700 transition-colors"
			>
				<Trash2 class="w-3.5 h-3.5" />
			</button>
		</div>
	</div>

	<!-- Add Sub-task form -->
	{#if showAddSubTask}
		<div class="ml-8 mt-1 pr-2" transition:slide>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					addSubTask();
				}}
				class="flex gap-2"
			>
				<input
					type="text"
					bind:value={newSubTaskContent}
					placeholder="New sub-task..."
					class="flex-1 bg-slate-900 border border-[var(--border-color)] rounded-lg px-2 py-1 text-xs text-slate-200 outline-none focus:border-rust transition-all"
				/>
				<button
					type="submit"
					disabled={!newSubTaskContent.trim()}
					class="p-1 bg-slate-700 hover:bg-rust text-white rounded transition-colors disabled:opacity-50"
				>
					<Plus class="w-3 h-3" />
				</button>
			</form>
		</div>
	{/if}

	<!-- Comments Section -->
	{#if showComments}
		<div class="ml-8 mt-2 space-y-2 pr-2 border-l-2 border-rust/30 pl-3" transition:slide>
			<div class="space-y-2">
				{#each todo.comments || [] as comment (comment.id)}
					<div class="flex flex-col gap-1 p-2 bg-slate-900/50 rounded-lg group/comment">
						<div class="flex justify-between items-center">
							<div class="flex items-center gap-1.5 text-[10px] font-bold text-rust-light">
								<User class="w-2.5 h-2.5" />
								<span>{comment.author || 'User'}</span>
								<span class="text-slate-600 font-normal ml-1"
									>{new Date(comment.created_at).toLocaleTimeString([], {
										hour: '2-digit',
										minute: '2-digit'
									})}</span
								>
							</div>
							<button
								onclick={() => deleteComment(comment.id)}
								class="opacity-0 group-hover/comment:opacity-100 text-slate-600 hover:text-red-400 transition-all"
							>
								<Trash2 class="w-2.5 h-2.5" />
							</button>
						</div>
						<p class="text-xs text-slate-700 dark:text-slate-300 leading-relaxed">
							{comment.content}
						</p>
					</div>
				{/each}
			</div>

			<form
				onsubmit={(e) => {
					e.preventDefault();
					addComment();
				}}
				class="flex gap-2 pt-1"
			>
				<input
					type="text"
					bind:value={newCommentContent}
					placeholder="Write a comment..."
					class="flex-1 bg-slate-900 border border-[var(--border-color)] rounded-lg px-2 py-1.5 text-xs text-slate-200 outline-none focus:border-rust transition-all"
				/>
				<button
					type="submit"
					disabled={!newCommentContent.trim()}
					class="p-1.5 bg-rust hover:bg-rust-light text-white rounded-lg transition-colors disabled:opacity-50"
				>
					<Send class="w-3 h-3" />
				</button>
			</form>
		</div>
	{/if}

	<!-- Recursive Sub-tasks -->
	{#if expanded && todo.sub_tasks && todo.sub_tasks.length > 0}
		<div
			class="ml-4 mt-1 flex flex-col gap-1 border-l border-slate-300/50 dark:border-slate-700/50 pl-2"
			transition:slide
		>
			{#each todo.sub_tasks as sub (sub.id)}
				<TaskItem todo={sub} {onToggle} {onDelete} {onToggleProgress} />
			{/each}
		</div>
	{/if}
</div>
