<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { Trash2, Edit2, Clock, AlertTriangle, ShieldAlert, Eye, X } from 'lucide-svelte';
	import { scale, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { Note } from '$lib/stores.svelte';
	import Button from '../Button.svelte';

	let { note, onDelete, onEdit } = $props<{
		note: Note;
		onDelete: (id: number) => void;
		onEdit: (note: Note) => void;
	}>();

	let showModal = $state(false);

	const MAX_CONTENT_LENGTH = 120;

	let truncatedContent = $derived(
		note.content.length > MAX_CONTENT_LENGTH
			? note.content.substring(0, MAX_CONTENT_LENGTH) + '...'
			: note.content
	);

	let isContentTruncated = $derived(note.content.length > MAX_CONTENT_LENGTH);

	function handleDelete() {
		onDelete(note.id);
	}

	function handleEdit() {
		onEdit(note);
	}

	function toggleModal() {
		showModal = !showModal;
	}

	const colorMap: Record<string, { bg: string; border: string; icon: string; shadow: string }> = {
		yellow: {
			bg: 'bg-amber-500/10',
			border: 'border-amber-500/20',
			icon: 'text-amber-400',
			shadow: 'shadow-amber-500/5'
		},
		blue: {
			bg: 'bg-sky-500/10',
			border: 'border-sky-500/20',
			icon: 'text-sky-400',
			shadow: 'shadow-sky-500/5'
		},
		green: {
			bg: 'bg-emerald-500/10',
			border: 'border-emerald-500/20',
			icon: 'text-emerald-400',
			shadow: 'shadow-emerald-500/5'
		},
		purple: {
			bg: 'bg-indigo-500/10',
			border: 'border-indigo-500/20',
			icon: 'text-indigo-400',
			shadow: 'shadow-indigo-500/5'
		},
		orange: {
			bg: 'bg-orange-500/10',
			border: 'border-orange-500/20',
			icon: 'text-orange-400',
			shadow: 'shadow-orange-500/5'
		},
		pink: {
			bg: 'bg-rose-500/10',
			border: 'border-rose-500/20',
			icon: 'text-rose-400',
			shadow: 'shadow-rose-500/5'
		},
		cyan: {
			bg: 'bg-cyan-500/10',
			border: 'border-cyan-500/20',
			icon: 'text-cyan-400',
			shadow: 'shadow-cyan-500/5'
		}
	};

	let theme = $derived(colorMap[note.color] || colorMap.blue);
</script>

<div
	class="modern-card group relative flex flex-col p-6 h-full min-h-[200px] {theme.bg} {theme.border} {theme.shadow} border rounded-2xl hover:border-white/20 transition-all duration-300 font-sans"
	in:scale={{ duration: 400, easing: cubicOut }}
>
	<!-- Header -->
	<div class="flex justify-between items-start mb-4">
		<div class="flex items-center gap-2">
			{#if note.status === 'warn'}
				<div class="p-1.5 bg-amber-500/20 rounded-lg text-amber-400 border border-amber-500/30">
					<AlertTriangle class="w-3.5 h-3.5" />
				</div>
			{:else if note.status === 'critical'}
				<div class="p-1.5 bg-rose-500/20 rounded-lg text-rose-400 border border-rose-500/30">
					<ShieldAlert class="w-3.5 h-3.5" />
				</div>
			{:else}
				<div class="p-1.5 {theme.bg} rounded-lg {theme.icon} border {theme.border}">
					<Clock class="w-3.5 h-3.5" />
				</div>
			{/if}
		</div>

		<div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
			<button
				onclick={handleEdit}
				class="p-2 text-slate-400 hover:text-white hover:bg-white/5 rounded-lg transition-all"
				title="Edit"
			>
				<Edit2 class="w-3.5 h-3.5" />
			</button>
			<button
				onclick={handleDelete}
				class="p-2 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all"
				title="Delete"
			>
				<Trash2 class="w-3.5 h-3.5" />
			</button>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 cursor-pointer" onclick={toggleModal} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && toggleModal()}>
		{#if note.title}
			<h3 class="text-white font-bold text-base mb-2 tracking-tight line-clamp-1">{note.title}</h3>
		{/if}
		<p class="text-sm text-slate-400 leading-relaxed line-clamp-4">{truncatedContent}</p>
	</div>

	<!-- Footer -->
	<div class="mt-6 pt-4 border-t border-white/5 flex items-center justify-between">
		<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
			{new Date(note.updated_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
		</span>
		{#if isContentTruncated}
			<button onclick={toggleModal} class="text-[10px] font-bold text-sky-400 hover:text-sky-300 uppercase tracking-widest transition-colors flex items-center gap-1">
				Read More <Eye class="w-3 h-3" />
			</button>
		{/if}
	</div>
</div>

<!-- Modal -->
{#if showModal}
	<div
		class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md"
		transition:fade={{ duration: 150 }}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="relative w-full max-w-2xl bg-slate-900 border border-white/10 rounded-3xl shadow-2xl overflow-hidden flex flex-col font-sans"
			transition:scale={{ duration: 200, start: 0.95 }}
			onclick={e => e.stopPropagation()}
		>
			<div class="px-8 py-6 border-b border-white/5 flex justify-between items-center bg-black/20">
				<div class="flex items-center gap-4">
					<div class="p-2.5 {theme.bg} {theme.border} border rounded-xl">
						<Eye class="w-5 h-5 {theme.icon}" />
					</div>
					<h2 class="text-xl font-bold text-white tracking-tight uppercase italic font-heading">
						{note.title || 'Note Details'}
					</h2>
				</div>
				<button onclick={toggleModal} class="text-slate-500 hover:text-white transition-all p-2 rounded-lg hover:bg-white/5">
					<X class="w-6 h-6" />
				</button>
			</div>

			<div class="p-8 max-h-[60vh] overflow-y-auto custom-scrollbar bg-black/40">
				<p class="text-slate-200 leading-relaxed whitespace-pre-wrap">{note.content}</p>
			</div>

			<div class="px-8 py-6 border-t border-white/5 bg-black/20 flex justify-between items-center">
				<div class="flex items-center gap-2 text-[10px] font-bold text-slate-500 uppercase tracking-widest">
					<Clock class="w-3.5 h-3.5" />
					Modified: {new Date(note.updated_at).toLocaleString()}
				</div>
				<Button onclick={toggleModal} variant="secondary" size="md" class="!rounded-xl">Close</Button>
			</div>
		</div>
	</div>
{/if}

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
