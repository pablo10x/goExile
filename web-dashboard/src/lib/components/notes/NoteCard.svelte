<script lang="ts">
	import { Trash2, Edit2, Clock, AlertTriangle, ShieldAlert } from 'lucide-svelte';
	import { scale } from 'svelte/transition';
	import type { Note } from '$lib/stores';

    let { note, onDelete, onEdit } = $props<{ 
        note: Note; 
        onDelete: (id: number) => void;
        onEdit: (note: Note) => void;
    }>();

	function handleDelete() {
		if (confirm('Delete this note?')) {
			onDelete(note.id);
		}
	}

	function handleEdit() {
		onEdit(note);
	}

	// Reduced rotation for a cleaner look, kept mainly for "pinned" vibe if desired, or remove if wanting strict grid.
    // Let's keep it subtle.
	let rotationStyle = $derived(`transform: rotate(${note.rotation}deg);`);

    const colorMap: Record<string, string> = {
        yellow: 'bg-yellow-50 text-yellow-900 border-yellow-200 shadow-sm shadow-yellow-100/50',
        blue: 'bg-blue-50 text-blue-900 border-blue-200 shadow-sm shadow-blue-100/50',
        green: 'bg-green-50 text-green-900 border-green-200 shadow-sm shadow-green-100/50',
        purple: 'bg-purple-50 text-purple-900 border-purple-200 shadow-sm shadow-purple-100/50',
        orange: 'bg-orange-50 text-orange-900 border-orange-200 shadow-sm shadow-orange-100/50',
        pink: 'bg-pink-50 text-pink-900 border-pink-200 shadow-sm shadow-pink-100/50',
        cyan: 'bg-cyan-50 text-cyan-900 border-cyan-200 shadow-sm shadow-cyan-100/50',
    };

	function getNoteCardClasses(color: string, status: Note['status']) {
        const base = colorMap[color] || 'bg-slate-50 text-slate-900 border-slate-200 shadow-sm shadow-slate-100/50';
        let statusClass = '';
		switch (status) {
			case 'warn': statusClass = 'ring-2 ring-amber-400 ring-offset-2 ring-offset-slate-900'; break;
			case 'critical': statusClass = 'ring-2 ring-red-500 ring-offset-2 ring-offset-slate-900'; break;
		}
		return `${base} ${statusClass}`;
	}
</script>

<div
	class="relative group w-full h-full min-h-[220px] p-5 rounded-xl border transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:z-10 flex flex-col {getNoteCardClasses(note.color, note.status)}"
	style={rotationStyle}
	transition:scale={{ duration: 300 }}
>
    <!-- Header: Status Icon & Actions -->
    <div class="flex justify-between items-start mb-3">
        <div class="flex gap-2">
            {#if note.status === 'warn'}
                <div class="text-amber-600 bg-amber-100/50 p-1 rounded-md" title="Warning">
                    <AlertTriangle class="w-4 h-4" />
                </div>
            {:else if note.status === 'critical'}
                <div class="text-red-600 bg-red-100/50 p-1 rounded-md" title="Critical">
                    <ShieldAlert class="w-4 h-4" />
                </div>
            {/if}
        </div>

        <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
			<button onclick={handleEdit} class="p-1.5 hover:bg-black/5 rounded-lg transition-colors text-current/70 hover:text-current" title="Edit">
				<Edit2 class="w-4 h-4" />
			</button>
			<button onclick={handleDelete} class="p-1.5 hover:bg-red-500/10 hover:text-red-600 rounded-lg transition-colors text-current/70" title="Delete">
				<Trash2 class="w-4 h-4" />
			</button>
		</div>
    </div>

	<div class="flex-1">
        {#if note.title}
		    <h3 class="font-bold text-lg mb-2 break-words leading-tight">{note.title}</h3>
        {/if}
		<p class="text-sm whitespace-pre-wrap break-words opacity-90 leading-relaxed font-medium">{note.content}</p>
	</div>
    
    <div class="mt-4 pt-3 border-t border-black/5 flex items-center gap-1.5 text-[10px] opacity-60 font-mono">
        <Clock class="w-3 h-3" />
        {new Date(note.updated_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
    </div>
</div>