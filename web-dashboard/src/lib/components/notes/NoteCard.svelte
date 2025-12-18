<script lang="ts">
	import { Trash2, Edit2, Bell, TriangleAlert, ShieldAlert } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import type { Note } from '$lib/stores';
	import { createEventDispatcher } from 'svelte';

	export let note: Note;

	const dispatch = createEventDispatcher();

	function deleteNote() {
		if (confirm('Delete this note?')) {
			dispatch('delete', note.id);
		}
	}

	function editNote() {
		dispatch('edit', note);
	}

	// Random rotation for "sticky note" effect
	$: rotationStyle = `transform: rotate(${note.rotation}deg);`;

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

<div
	class="relative group w-full sm:w-64 min-h-[200px] p-4 rounded-sm shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-xl hover:z-10 {getNoteCardClasses(note.color, note.status)}"
	style={rotationStyle}
	transition:scale={{ duration: 300 }}
>
	<div class="flex flex-col h-full">
		<h3 class="font-bold text-lg mb-2 break-words">{note.title}</h3>
		<p class="text-sm whitespace-pre-wrap flex-1 break-words opacity-90">{note.content}</p>

		<div class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
			<button onclick={editNote} class="p-1 hover:bg-black/10 rounded-full transition-colors" title="Edit">
				<Edit2 class="w-3.5 h-3.5" />
			</button>
			<button onclick={deleteNote} class="p-1 hover:bg-black/10 rounded-full transition-colors hover:text-red-700" title="Delete">
				<Trash2 class="w-3.5 h-3.5" />
			</button>
		</div>
		
		<div class="mt-4 text-[10px] opacity-50 text-right">
			{new Date(note.updated_at).toLocaleDateString()}
		</div>
		
		{#if note.status === 'warn'}
			<div class="absolute -top-3 left-1/2 -translate-x-1/2 w-3 h-3 rounded-full bg-amber-500 shadow-sm border border-amber-600 z-20" title="Warning Note"></div>
		{:else if note.status === 'critical'}
			<div class="absolute -top-3 left-1/2 -translate-x-1/2 w-3 h-3 rounded-full bg-red-600 shadow-sm border border-red-700 z-20" title="Critical Note"></div>
		{:else}
			<!-- Pin effect for normal notes (default) -->
			<div class="absolute -top-3 left-1/2 -translate-x-1/2 w-3 h-3 rounded-full bg-slate-500 shadow-sm border border-slate-600 z-20"></div>
		{/if}
	</div>
</div>
