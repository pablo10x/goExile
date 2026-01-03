<script lang="ts">
	import {
		Save,
		X,
		RotateCw,
		Bell,
		TriangleAlert,
		ShieldAlert,
		ChevronRight
	} from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { Note } from '$lib/stores.svelte';
	import { autofocus } from '$lib/actions';

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
	let isEditing = $derived(initialNote !== null && (initialNote.id !== 0 || initialNote.title !== '' || initialNote.content !== ''));

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
				rotation: initialNote.rotation || Math.floor(Math.random() * 6) - 3,
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

	function getNoteCardClasses(color: string, status: Note['status']) {
		let colorClass = '';
		let statusClass = '';

		switch (color) {
			case 'yellow':
				colorClass = 'bg-yellow-100/80 text-yellow-900 border-yellow-200/50';
				break;
			case 'blue':
				colorClass = 'bg-orange-100/80 text-orange-900 border-orange-200/50';
				break;
			case 'green':
				colorClass = 'bg-green-100/80 text-green-900 border-green-200/50';
				break;
			case 'purple':
				colorClass = 'bg-purple-100/80 text-purple-900 border-purple-200/50';
				break;
			case 'orange':
				colorClass = 'bg-orange-100/80 text-orange-900 border-orange-200/50';
				break;
			case 'pink':
				colorClass = 'bg-pink-100/80 text-pink-900 border-pink-200/50';
				break;
			case 'cyan':
				colorClass = 'bg-cyan-100/80 text-cyan-900 border-cyan-200/50';
				break;
			default:
				colorClass = 'bg-slate-100/80 text-slate-900 border-slate-200/50';
				break;
		}

		switch (status) {
			case 'warn':
				statusClass = 'border-amber-400 ring-2 ring-amber-300/50';
				break;
			case 'critical':
				statusClass = 'border-red-500 ring-2 ring-red-400/50';
				break;
			case 'normal':
			default:
				break;
		}

		return `${colorClass} ${statusClass}`;
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
		onclick={handleBackdropClick}
		onkeydown={(e) => {
			if (e.key === 'Escape' && !loading) onClose();
			if (
				e.key === 'Enter' &&
				e.ctrlKey &&
				!loading &&
				(currentNote.title.trim() || currentNote.content.trim())
			)
				handleSave();
		}}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="relative w-full max-w-lg p-0 transition-all duration-300 flex flex-col items-center"
			style="transform: rotate({currentNote.rotation}deg);"
			transition:scale={{ start: 0.8, duration: 200, easing: cubicOut }}
		>
			<!-- Card Body -->
			<div
				class="relative w-full min-h-[350px] flex flex-col p-8 rounded-none border-2 shadow-2xl {getNoteCardClasses(
					currentNote.color,
					currentNote.status
				)} industrial-sharp backdrop-blur-md"
			>
				<div class="absolute inset-0 pointer-events-none opacity-[0.03] bg-[url('https://www.transparenttextures.com/patterns/pinstriped-suit.png')]"></div>

				<div class="relative z-10 flex flex-col h-full flex-1">
					<div class="flex items-center gap-2 mb-4 opacity-40">
						<ChevronRight class="w-3 h-3" />
						<span class="text-[8px] font-black font-jetbrains uppercase tracking-widest">SIGNAL_MEMO_INPUT</span>
					</div>

					<input
						type="text"
						bind:value={currentNote.title}
						class="w-full bg-transparent border-b-2 border-black/10 px-0 py-2 font-black font-heading text-2xl outline-none placeholder-black/20 mb-4 uppercase tracking-tighter"
						placeholder="MEMO_IDENTIFIER"
						use:autofocus
					/>
					<textarea
						bind:value={currentNote.content}
						class="flex-1 w-full bg-transparent p-0 resize-none outline-none text-base font-jetbrains font-bold placeholder-black/20 leading-relaxed uppercase"
						placeholder="ENTER_SIGNAL_DATA..."
					></textarea>
				</div>

				<button
					onclick={onClose}
					class="absolute top-4 right-4 p-1.5 text-black/40 hover:text-black hover:bg-black/5 transition-all rounded-none"
					title="Discard"
					disabled={loading}
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Controls Bar -->
			<div class="mt-8 flex flex-col gap-4 w-full max-w-md" style="transform: rotate({-currentNote.rotation}deg);">
				<div class="flex flex-wrap gap-4 justify-center">
					<!-- Color Palette -->
					<div
						class="p-2 bg-black/60 backdrop-blur-md border border-stone-800 flex gap-2 shadow-2xl industrial-sharp"
					>
						{#each noteColors as color}
							<button
								onclick={() => (currentNote.color = color)}
								class="w-6 h-6 border transition-all {currentNote.color === color
									? 'border-white scale-110 shadow-lg'
									: 'border-white/10 hover:border-white/40'}"
								style="background-color: {
									color === 'yellow' ? '#facc15' : 
									color === 'blue' ? '#fb923c' : 
									color === 'green' ? '#4ade80' : 
									color === 'purple' ? '#c084fc' : 
									color === 'orange' ? '#f97316' : 
									color === 'pink' ? '#f472b6' : 
									'#22d3ee'
								}"
								title={color.toUpperCase()}
							>
								{#if currentNote.color === color}
									<div class="w-full h-full flex items-center justify-center">
										<div class="w-1.5 h-1.5 bg-black rounded-full"></div>
									</div>
								{/if}
							</button>
						{/each}
					</div>

					<!-- Status Selector -->
					<div
						class="p-2 bg-black/60 backdrop-blur-md border border-stone-800 flex gap-2 shadow-2xl industrial-sharp"
					>
						{#each noteStatuses as status}
							<button
								onclick={() => (currentNote.status = status)}
								class="w-8 h-8 flex items-center justify-center border transition-all {currentNote.status === status
									? 'bg-rust/20 border-rust text-white shadow-lg shadow-rust/20'
									: 'border-white/5 text-stone-500 hover:text-stone-300 hover:border-white/20'}"
								title={status.toUpperCase()}
							>
								{#if status === 'normal'}
									<Bell class="w-4 h-4" />
								{:else if status === 'warn'}
									<TriangleAlert class="w-4 h-4 text-warning" />
								{:else if status === 'critical'}
									<ShieldAlert class="w-4 h-4 text-danger" />
								{/if}
							</button>
						{/each}
					</div>
				</div>

				<!-- Save Button -->
				<button
					onclick={handleSave}
					disabled={loading || (!currentNote.title.trim() && !currentNote.content.trim())}
					class="w-full py-4 bg-rust hover:bg-rust-light text-white font-heading font-black text-xs uppercase tracking-[0.3em] shadow-2xl transition-all flex items-center justify-center gap-4 disabled:opacity-30 industrial-sharp"
				>
					{#if loading}
						<RotateCw class="w-5 h-5 animate-spin" />
						<span>Committing_Buffer...</span>
					{:else}
						<Save class="w-5 h-5" />
						<span>{isEditing ? 'Sync_Changes' : 'Initialize_Memo'}</span>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}