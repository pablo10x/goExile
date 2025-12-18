<script lang="ts">
	import { Save, X, RotateCw, Palette, Bell, TriangleAlert, ShieldAlert, Check } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { Note } from '$lib/stores';
	import { autofocus } from '$lib/actions'; // Reusing the autofocus action

	let {
		isOpen = $bindable(false),
		note: initialNote = null, // Note object for editing, or null for creation
		onSave, // (note: Note) => Promise<void>
		onClose // () => void
	} = $props<{
		isOpen: boolean;
		note?: Note | null;
		onSave: (note: Note) => Promise<void>;
		onClose: () => void;
	}>();

	let currentNote = $state<Note>({
		id: initialNote?.id || 0, // id is 0 for new notes
		title: initialNote?.title || '',
		content: initialNote?.content || '',
		color: initialNote?.color || 'yellow', // Default color
		status: initialNote?.status || 'normal', // Default status
		rotation: initialNote?.rotation || 0,
		created_at: initialNote?.created_at || new Date().toISOString(),
		updated_at: initialNote?.updated_at || new Date().toISOString()
	});

	let loading = $state(false);
	let isEditing = $derived(initialNote !== null && initialNote.id !== 0);

	const noteColors = ['yellow', 'blue', 'green', 'purple', 'orange', 'pink', 'cyan']; // Extended colors
	const noteStatuses: Array<Note['status']> = ['normal', 'warn', 'critical'];

	// Reset form when modal opens or initialNote changes (for editing)
	$effect(() => {
		if (isOpen) {
			currentNote = {
				id: initialNote?.id || 0,
				title: initialNote?.title || '',
				content: initialNote?.content || '',
				color: initialNote?.color || 'yellow',
				status: initialNote?.status || 'normal',
				rotation: initialNote?.rotation || Math.floor(Math.random() * 6) - 3, // Small random rotation for new notes
				created_at: initialNote?.created_at || new Date().toISOString(),
				updated_at: initialNote?.updated_at || new Date().toISOString()
			};
		}
	});

	async function handleSave() {
		if (!currentNote.title.trim() && !currentNote.content.trim()) {
			// Optionally show a notification
			return;
		}
		loading = true;
		try {
			// Ensure currentNote has required fields for the backend if it's a new note
			const noteToSave: Note = {
				...currentNote,
				created_at: currentNote.created_at || new Date().toISOString(),
				updated_at: new Date().toISOString(),
				id: currentNote.id || 0 // Ensure ID is set for type compliance, 0 for new
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

	function getNoteCardClasses(color: string, status: Note['status']) {
		let colorClass = '';
		let statusClass = '';

		// Base colors
		switch (color) {
			case 'yellow': colorClass = 'bg-yellow-100 text-yellow-900 border-yellow-200'; break;
			case 'blue': colorClass = 'bg-blue-100 text-blue-900 border-blue-200'; break;
			case 'green': colorClass = 'bg-green-100 text-green-900 border-green-200'; break;
			case 'purple': colorClass = 'bg-purple-100 text-purple-900 border-purple-200'; break;
			case 'orange': colorClass = 'bg-orange-100 text-orange-900 border-orange-200'; break;
			case 'pink': colorClass = 'bg-pink-100 text-pink-900 border-pink-200'; break;
			case 'cyan': colorClass = 'bg-cyan-100 text-cyan-900 border-cyan-200'; break;
			default: colorClass = 'bg-slate-100 text-slate-900 border-slate-200'; break; // Default to a neutral color
		}

		// Status overlays
		switch (status) {
			case 'warn': statusClass = 'border-amber-400 ring-2 ring-amber-300/50'; break;
			case 'critical': statusClass = 'border-red-500 ring-2 ring-red-400/50'; break;
			case 'normal': // no additional class for normal
			default: break;
		}

		return `${colorClass} ${statusClass}`;
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
		onclick={handleBackdropClick}
		onkeydown={(e) => {
			if (e.key === 'Escape' && !loading) onClose();
			if (e.key === 'Enter' && e.ctrlKey && !loading && (currentNote.title.trim() || currentNote.content.trim())) handleSave();
		}}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<div
			class="relative w-full max-w-lg p-6 rounded-sm shadow-xl transition-all duration-300 flex flex-col items-center justify-between"
			class:animate-pop-in={isOpen}
			style="transform: rotate({currentNote.rotation}deg);"
			transition:scale={{ start: 0.8, duration: 200, easing: cubicOut }}
		>
			<!-- Card Body -->
			<div class="relative w-full h-full min-h-[300px] flex flex-col p-6 rounded-sm border {getNoteCardClasses(currentNote.color, currentNote.status)}">
				<input
					type="text"
					bind:value={currentNote.title}
					class="w-full bg-transparent border-b border-black/10 px-1 py-0.5 font-bold text-xl outline-none placeholder-black/30 mb-2"
					placeholder="Note Title"
					use:autofocus
				/>
				<textarea
					bind:value={currentNote.content}
					class="flex-1 w-full bg-transparent p-1 resize-none outline-none text-base placeholder-black/30"
					placeholder="Write your note here..."
				></textarea>

				<!-- Close Button -->
				<button
					onclick={onClose}
					class="absolute top-2 right-2 p-1 text-slate-600 hover:text-black hover:bg-black/10 rounded-full transition-colors"
					title="Close"
					disabled={loading}
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Controls Bar -->
			<div class="mt-4 flex flex-wrap gap-2 justify-center w-full">
				<!-- Color Palette -->
				<div class="p-2 bg-white/70 backdrop-blur-sm rounded-lg flex gap-1 shadow-md border border-slate-300">
					{#each noteColors as color}
						<button
							onclick={() => (currentNote.color = color)}
							class="w-7 h-7 rounded-full border-2 border-transparent {currentNote.color === color ? 'ring-2 ring-offset-1 ring-slate-800' : ''}"
							class:bg-yellow-400={color === 'yellow'}
							class:bg-blue-400={color === 'blue'}
							class:bg-green-400={color === 'green'}
							class:bg-purple-400={color === 'purple'}
							class:bg-orange-400={color === 'orange'}
							class:bg-pink-400={color === 'pink'}
							class:bg-cyan-400={color === 'cyan'}
							title={color.charAt(0).toUpperCase() + color.slice(1)}
						>
							{#if currentNote.color === color}
								<Check class="w-full h-full text-white p-1" />
							{/if}
						</button>
					{/each}
				</div>

				<!-- Status Selector -->
				<div class="p-2 bg-white/70 backdrop-blur-sm rounded-lg flex gap-1 shadow-md border border-slate-300">
					{#each noteStatuses as status}
						<button
							onclick={() => (currentNote.status = status)}
							class="w-7 h-7 rounded-full border-2 border-transparent {currentNote.status === status ? 'ring-2 ring-offset-1 ring-slate-800' : ''} flex items-center justify-center"
							class:bg-slate-400={status === 'normal'}
							class:bg-amber-400={status === 'warn'}
							class:bg-red-500={status === 'critical'}
							title={status.charAt(0).toUpperCase() + status.slice(1)}
						>
							{#if status === 'normal'}
								<Bell class="w-4 h-4 text-white" />
							{:else if status === 'warn'}
								<TriangleAlert class="w-4 h-4 text-white" />
							{:else if status === 'critical'}
								<ShieldAlert class="w-4 h-4 text-white" />
							{/if}
						</button>
					{/each}
				</div>

				<!-- Save Button -->
				<button
					onclick={handleSave}
					disabled={loading || (!currentNote.title.trim() && !currentNote.content.trim())}
					class="px-5 py-2.5 bg-green-500 hover:bg-green-600 text-white font-semibold rounded-lg shadow-md transition-colors flex items-center gap-2 disabled:opacity-50"
				>
					{#if loading}
						<RotateCw class="w-4 h-4 animate-spin" />
					{:else}
						<Save class="w-4 h-4" />
					{/if}
					{isEditing ? 'Save Changes' : 'Create Note'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Keyframe for modal pop-in */
	@keyframes pop-in {
		0% {
			transform: scale(0.8);
			opacity: 0;
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}
	.animate-pop-in {
		animation: pop-in 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
	}
</style>