<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { Save, X, RotateCw, Bell, TriangleAlert, ShieldAlert, ChevronRight, FileText } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { Note } from '$lib/stores.svelte';
	import { autofocus } from '$lib/actions';
	import Button from '../Button.svelte';

	let {
		isOpen = $bindable(false),
		note: initialNote = null,
		onSave,
		onClose
	} = $props<{
		isOpen: boolean;
		note?: Note | null;
		onSave: (note: Note) => Promise<void>;
		onClose: () => void;
	}>();

	let currentNote = $state<Note>({
		id: 0,
		title: '',
		content: '',
		color: 'yellow',
		status: 'normal',
		rotation: 0,
		created_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	});

	let loading = $state(false);
	let isEditing = $derived(
		initialNote !== null &&
			(initialNote.id !== 0 || initialNote.title !== '' || initialNote.content !== '')
	);

	const noteColors = ['yellow', 'blue', 'green', 'purple', 'orange', 'pink', 'cyan'];
	const noteStatuses: Array<Note['status']> = ['normal', 'warn', 'critical'];

	$effect(() => {
		if (isOpen && initialNote) {
			currentNote = {
				id: initialNote.id || 0,
				title: initialNote.title || '',
				content: initialNote.content || '',
				color: initialNote.color || 'yellow',
				status: initialNote.status || 'normal',
				rotation: initialNote.rotation || 0,
				created_at: initialNote.created_at || new Date().toISOString(),
				updated_at: initialNote.updated_at || new Date().toISOString()
			};
		}
	});

	async function handleSave() {
		if (!currentNote.title.trim() && !currentNote.content.trim()) {
			return;
		}
		loading = true;
		try {
			const noteToSave: Note = {
				...currentNote,
				created_at: currentNote.created_at || new Date().toISOString(),
				updated_at: new Date().toISOString(),
				id: currentNote.id || 0
			};

			await onSave(noteToSave);
			onClose();
		} finally {
			loading = false;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget && !loading) {
			onClose();
		}
	}

	const colorThemeMap: Record<string, { bg: string; border: string; icon: string }> = {
		yellow: { bg: 'bg-amber-500/10', border: 'border-amber-500/20', icon: 'text-amber-400' },
		blue: { bg: 'bg-sky-500/10', border: 'border-sky-500/20', icon: 'text-sky-400' },
		green: { bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', icon: 'text-emerald-400' },
		purple: { bg: 'bg-indigo-500/10', border: 'border-indigo-500/20', icon: 'text-indigo-400' },
		orange: { bg: 'bg-orange-500/10', border: 'border-orange-500/20', icon: 'text-orange-400' },
		pink: { bg: 'bg-rose-500/10', border: 'border-rose-500/20', icon: 'text-rose-400' },
		cyan: { bg: 'bg-cyan-500/10', border: 'border-cyan-500/20', icon: 'text-cyan-400' }
	};

	let activeTheme = $derived(colorThemeMap[currentNote.color] || colorThemeMap.blue);
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="relative w-full max-w-2xl bg-slate-900 border border-white/10 rounded-3xl shadow-2xl overflow-hidden flex flex-col font-sans"
			transition:scale={{ duration: 200, start: 0.95 }}
			onclick={e => e.stopPropagation()}
		>
			<!-- Header -->
			<div class="px-8 py-6 border-b border-white/5 flex justify-between items-center bg-black/20">
				<div class="flex items-center gap-4">
					<div class="p-2.5 {activeTheme.bg} {activeTheme.border} border rounded-xl">
						<FileText class="w-5 h-5 {activeTheme.icon}" />
					</div>
					<h2 class="text-xl font-bold text-white tracking-tight uppercase italic font-heading">
						{isEditing ? 'Edit Note' : 'Create New Note'}
					</h2>
				</div>
				<button onclick={onClose} class="text-slate-500 hover:text-white transition-all p-2 rounded-lg hover:bg-white/5">
					<X class="w-6 h-6" />
				</button>
			</div>

			<!-- Body -->
			<div class="p-8 space-y-8 bg-black/40 flex-1">
				<div class="space-y-6">
					<div class="space-y-2">
						<label class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Title</label>
						<input
							type="text"
							bind:value={currentNote.title}
							class="w-full bg-black border border-white/10 px-4 py-3 text-white font-bold rounded-xl focus:border-sky-500 outline-none transition-all placeholder:text-slate-800"
							placeholder="Enter note title..."
							use:autofocus
						/>
					</div>

					<div class="space-y-2">
						<label class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Content</label>
						<textarea
							bind:value={currentNote.content}
							rows="8"
							class="w-full bg-black border border-white/10 p-4 text-slate-200 rounded-xl focus:border-sky-500 outline-none transition-all placeholder:text-slate-800 resize-none leading-relaxed"
							placeholder="Write your note here..."
						></textarea>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
					<!-- Color Selection -->
					<div class="space-y-3">
						<label class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Color Theme</label>
						<div class="flex flex-wrap gap-2 p-3 bg-black/40 border border-white/5 rounded-xl shadow-inner">
							{#each noteColors as color}
								<button
									onclick={() => (currentNote.color = color)}
									class="w-7 h-7 rounded-lg border-2 transition-all {currentNote.color === color
										? 'border-white scale-110 shadow-lg'
										: 'border-transparent hover:border-white/20'}"
									style="background-color: {color === 'yellow' ? '#fbbf24' : color === 'blue' ? '#0ea5e9' : color === 'green' ? '#10b981' : color === 'purple' ? '#8b5cf6' : color === 'orange' ? '#f59e0b' : color === 'pink' ? '#ec4899' : '#06b6d4'}"
									title={color}
								></button>
							{/each}
						</div>
					</div>

					<!-- Status Selection -->
					<div class="space-y-3">
						<label class="text-[10px] font-bold text-slate-500 uppercase tracking-widest ml-1">Status Priority</label>
						<div class="flex gap-2 p-1.5 bg-black/40 border border-white/5 rounded-xl shadow-inner">
							{#each noteStatuses as status}
								<button
									onclick={() => (currentNote.status = status)}
									class="flex-1 py-2.5 flex items-center justify-center rounded-lg border transition-all {currentNote.status === status
										? 'bg-sky-500/10 border-sky-500/40 text-sky-400 shadow-lg'
										: 'border-transparent text-slate-600 hover:text-slate-300'}"
								>
									{#if status === 'normal'}
										<Bell class="w-4 h-4" />
									{:else if status === 'warn'}
										<TriangleAlert class="w-4 h-4" />
									{:else}
										<ShieldAlert class="w-4 h-4" />
									{/if}
								</button>
							{/each}
						</div>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="px-8 py-6 border-t border-white/5 bg-black/20 flex justify-end items-center gap-4">
				<button
					onclick={onClose}
					class="px-6 py-2.5 text-xs font-bold text-slate-500 hover:text-white uppercase tracking-widest transition-all"
				>
					Cancel
				</button>
				<button
					onclick={handleSave}
					disabled={loading || (!currentNote.title.trim() && !currentNote.content.trim())}
					class="px-8 py-3 bg-sky-600 hover:bg-sky-500 text-white font-bold text-xs uppercase tracking-widest rounded-xl shadow-lg shadow-sky-500/20 transition-all disabled:opacity-20 flex items-center gap-3"
				>
					{#if loading}
						<RotateCw class="w-4 h-4 animate-spin" />
						<span>Saving...</span>
					{:else}
						<Save class="w-4 h-4" />
						<span>{isEditing ? 'Update Note' : 'Create Note'}</span>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
</style>
