<script lang="ts">
	import { Trash2, Edit2, Clock, AlertTriangle, ShieldAlert, Eye } from 'lucide-svelte';
	import { scale, fade } from 'svelte/transition';
	import { cubicOut, cubicIn } from 'svelte/easing';
	import type { Note } from '$lib/stores';

	// Explosive shatter animation - creates multiple pieces flying away
	function shatter(
		node: Element,
		{ duration = 1000, delay = 0 }: { duration?: number; delay?: number }
	) {
		const fragments = 12;
		const rect = node.getBoundingClientRect();

		return {
			delay,
			duration,
			easing: cubicIn,
			css: (t: number) => {
				const progress = 1 - t;

				const explosionRadius = progress * 400;
				const rotation = progress * 720;
				const scale = 1 - progress * 0.8;

				return `
                    opacity: ${Math.max(0, 1 - progress * 1.2)};
                    transform: scale(${scale}) rotate(${rotation}deg);
                    filter: blur(${progress * 8}px) brightness(${1 + progress * 0.5});
                `;
			},
			tick: (t: number) => {
				const progress = 1 - t;
				if (progress > 0.1 && progress < 0.9) {
					if (Math.random() > 0.85) {
						createFragment(node, progress);
					}
				}
			}
		};
	}

	function createFragment(node: Element, progress: number) {
		const rect = node.getBoundingClientRect();
		const fragment = document.createElement('div');

		const size = Math.random() * 40 + 20;
		const startX = rect.left + Math.random() * rect.width;
		const startY = rect.top + Math.random() * rect.height;
		const angle = Math.random() * Math.PI * 2;
		const velocity = 200 + Math.random() * 300;
		const rotation = Math.random() * 720 - 360;

		fragment.style.cssText = `
            position: fixed;
            left: ${startX}px;
            top: ${startY}px;
            width: ${size}px;
            height: ${size}px;
            background: inherit;
            border: inherit;
            border-radius: ${Math.random() * 8}px;
            pointer-events: none;
            z-index: 9999;
            transform-origin: center;
        `;

		document.body.appendChild(fragment);

		const animation = fragment.animate(
			[
				{
					transform: `translate(0, 0) rotate(0deg) scale(1)`,
					opacity: 1
				},
				{
					transform: `translate(${Math.cos(angle) * velocity}px, ${Math.sin(angle) * velocity + 300}px) rotate(${rotation}deg) scale(0.2)`,
					opacity: 0
				}
			],
			{
				duration: 800,
				easing: 'cubic-bezier(0.4, 0.0, 1, 1)',
				fill: 'forwards'
			}
		);

		animation.onfinish = () => fragment.remove();
	}

	let { note, onDelete, onEdit } = $props<{
		note: Note;
		onDelete: (id: number) => void;
		onEdit: (note: Note) => void;
	}>();

	let isShredding = $state(false);
	let showModal = $state(false);

	const MAX_CONTENT_LENGTH = 80;

	let truncatedContent = $derived(() => {
		if (note.content.length > MAX_CONTENT_LENGTH) {
			return note.content.substring(0, MAX_CONTENT_LENGTH) + '...';
		}
		return note.content;
	});

	let isContentTruncated = $derived(note.content.length > MAX_CONTENT_LENGTH);

	function handleDelete() {
		if (isShredding) return;
		isShredding = true;

		const card = document.querySelector(`[data-note-id="${note.id}"]`);
		if (card) {
			for (let i = 0; i < 20; i++) {
				setTimeout(() => createFragment(card, 0.5), i * 20);
			}
		}

		setTimeout(() => {
			onDelete(note.id);
		}, 1050);
	}

	function handleEdit() {
		onEdit(note);
	}

	function toggleModal() {
		showModal = !showModal;
	}

	function closeModal(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			showModal = false;
		}
	}

	let rotationStyle = $derived(`transform: rotate(${note.rotation}deg);`);

	const colorMap: Record<string, { base: string; paper: string; edge: string; shadow: string }> = {
		yellow: {
			base: '#fefce8',
			paper: 'linear-gradient(135deg, #fefce8 0%, #fef9c3 50%, #fef08a 100%)',
			edge: 'rgba(234, 179, 8, 0.3)',
			shadow: 'rgba(234, 179, 8, 0.15)'
		},
		blue: {
			base: '#eff6ff',
			paper: 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 50%, #bfdbfe 100%)',
			edge: 'rgba(59, 130, 246, 0.3)',
			shadow: 'rgba(59, 130, 246, 0.15)'
		},
		green: {
			base: '#f0fdf4',
			paper: 'linear-gradient(135deg, #f0fdf4 0%, #dcfce7 50%, #bbf7d0 100%)',
			edge: 'rgba(34, 197, 94, 0.3)',
			shadow: 'rgba(34, 197, 94, 0.15)'
		},
		purple: {
			base: '#faf5ff',
			paper: 'linear-gradient(135deg, #faf5ff 0%, #f3e8ff 50%, #e9d5ff 100%)',
			edge: 'rgba(168, 85, 247, 0.3)',
			shadow: 'rgba(168, 85, 247, 0.15)'
		},
		orange: {
			base: '#fff7ed',
			paper: 'linear-gradient(135deg, #fff7ed 0%, #ffedd5 50%, #fed7aa 100%)',
			edge: 'rgba(249, 115, 22, 0.3)',
			shadow: 'rgba(249, 115, 22, 0.15)'
		},
		pink: {
			base: '#fdf2f8',
			paper: 'linear-gradient(135deg, #fdf2f8 0%, #fce7f3 50%, #fbcfe8 100%)',
			edge: 'rgba(236, 72, 153, 0.3)',
			shadow: 'rgba(236, 72, 153, 0.15)'
		},
		cyan: {
			base: '#ecfeff',
			paper: 'linear-gradient(135deg, #ecfeff 0%, #cffafe 50%, #a5f3fc 100%)',
			edge: 'rgba(6, 182, 212, 0.3)',
			shadow: 'rgba(6, 182, 212, 0.15)'
		}
	};

	function getNoteStyles(color: string) {
		const colors = colorMap[color] || colorMap.yellow;
		return {
			background: colors.paper,
			boxShadow: `
                0 2px 8px ${colors.shadow},
                0 8px 24px ${colors.shadow},
                inset 0 1px 0 rgba(255, 255, 255, 0.9),
                inset 0 -1px 2px rgba(0, 0, 0, 0.05),
                inset 2px 0 2px rgba(0, 0, 0, 0.02),
                inset -2px 0 2px rgba(0, 0, 0, 0.02)
            `,
			borderColor: colors.edge
		};
	}

	function getStatusRing(status: Note['status']) {
		switch (status) {
			case 'warn':
				return 'ring-2 ring-amber-400 ring-offset-4 ring-offset-slate-900';
			case 'critical':
				return 'ring-2 ring-red-500 ring-offset-4 ring-offset-slate-900';
			default:
				return '';
		}
	}

	let noteStyles = $derived(getNoteStyles(note.color));
</script>

<div
	data-note-id={note.id}
	class="note-card relative group w-full h-full min-h-[220px] p-6 border-2 transition-all duration-300 hover:scale-[1.03] hover:z-10 flex flex-col font-jetbrains {getStatusRing(
		note.status
	)} rounded-none"
	style="{rotationStyle}
           background: {noteStyles.background};
           box-shadow: {noteStyles.boxShadow};
           border-color: {noteStyles.borderColor};
           {isShredding ? 'pointer-events: none;' : ''}"
	in:scale={{ duration: 400, easing: cubicOut }}
	out:shatter={{ duration: 1000 }}
>
	<!-- Paper texture overlay -->
	<div class="paper-grain"></div>

	<!-- Torn edge effect at top -->
	<div class="torn-edge"></div>

	<!-- Dog-ear corner fold -->
	<div class="dog-ear"></div>

	<!-- Subtle ruled lines -->
	<div class="ruled-lines"></div>

	<!-- Content wrapper -->
	<div class="content-wrapper">
		<!-- Header: Status Icon & Actions -->
		<div class="flex justify-between items-start mb-4">
			<div class="flex gap-2">
				{#if note.status === 'warn'}
					<div class="status-badge status-warn">
						<AlertTriangle class="w-4 h-4" />
					</div>
				{:else if note.status === 'critical'}
					<div class="status-badge status-critical">
						<ShieldAlert class="w-4 h-4" />
					</div>
				{/if}
			</div>

			<div class="action-buttons">
				{#if isContentTruncated}
					<button
						onclick={toggleModal}
						class="action-btn view-btn"
						title="View Full Note"
						disabled={isShredding}
					>
						<Eye class="w-4 h-4" />
					</button>
				{/if}
				<button
					onclick={handleEdit}
					class="action-btn edit-btn"
					title="Edit"
					disabled={isShredding}
				>
					<Edit2 class="w-4 h-4" />
				</button>
				<button
					onclick={handleDelete}
					class="action-btn delete-btn"
					title="Delete"
					disabled={isShredding}
				>
					<Trash2 class="w-4 h-4" />
				</button>
			</div>
		</div>

		<div class="flex-1">
			{#if note.title}
				<h3 class="note-title">{note.title}</h3>
			{/if}
			<p class="note-content">{truncatedContent()}</p>
		</div>

		<div class="note-timestamp">
			<Clock class="w-3 h-3" />
			<span
				>{new Date(note.updated_at).toLocaleDateString(undefined, {
					month: 'short',
					day: 'numeric',
					hour: '2-digit',
					minute: '2-digit'
				})}</span
			>
		</div>
	</div>
</div>

<!-- Beautiful Modal for Full Note View -->
{#if showModal}
	<div
		class="modal-overlay"
		onclick={closeModal}
		transition:fade={{ duration: 200 }}
		role="button"
		tabindex="0"
		onkeydown={(e) => {
			if (e.key === 'Escape') showModal = false;
		}}
	>
		<div
			class="modal-content font-jetbrains rounded-none"
			onclick={(e) => e.stopPropagation()}
			transition:scale={{ duration: 300, easing: cubicOut, start: 0.9 }}
		>
			<!-- Modal Header -->
			<div class="modal-header">
				<div class="flex items-center gap-3">
					<div class="modal-icon">
						<Eye class="w-5 h-5" />
					</div>
					{#if note.title}
						<h2 class="modal-title">{note.title}</h2>
					{:else}
						<h2 class="modal-title">Note Details</h2>
					{/if}
				</div>
				<button onclick={toggleModal} class="modal-close-btn" aria-label="Close">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="w-5 h-5"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<!-- Modal Body -->
			<div class="modal-body">
				<div class="modal-paper">
					<!-- Content -->
					<div class="modal-text">
						{note.content}
					</div>
				</div>
			</div>

			<!-- Modal Footer -->
			<div class="modal-footer">
				<div class="modal-meta">
					<Clock class="w-4 h-4" />
					<span>
						{new Date(note.updated_at).toLocaleDateString(undefined, {
							month: 'long',
							day: 'numeric',
							year: 'numeric',
							hour: '2-digit',
							minute: '2-digit'
						})}
					</span>
				</div>
				<button onclick={toggleModal} class="modal-done-btn"> Done </button>
			</div>
		</div>
	</div>
{/if}

<style>
	.note-card {
		position: relative;
		isolation: isolate;
	}

	.note-card:hover {
		box-shadow:
			0 4px 16px var(--shadow-color, rgba(0, 0, 0, 0.1)),
			0 12px 32px var(--shadow-color, rgba(0, 0, 0, 0.15)),
			inset 0 1px 0 rgba(255, 255, 255, 0.95),
			inset 0 -1px 2px rgba(0, 0, 0, 0.08),
			inset 2px 0 3px rgba(0, 0, 0, 0.03),
			inset -2px 0 3px rgba(0, 0, 0, 0.03) !important;
	}

	.paper-grain {
		position: absolute;
		inset: 0;
		background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 512 512' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.95' numOctaves='5' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.08'/%3E%3C/svg%3E");
		border-radius: inherit;
		pointer-events: none;
		opacity: 0.7;
		mix-blend-mode: overlay;
	}

	.torn-edge {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 8px;
		background: linear-gradient(180deg, rgba(0, 0, 0, 0.03) 0%, transparent 100%);
		border-radius: 16px 16px 0 0;
		pointer-events: none;
		z-index: 1;
	}

	.torn-edge::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: repeating-linear-gradient(
			90deg,
			transparent 0px,
			transparent 3px,
			rgba(0, 0, 0, 0.04) 3px,
			rgba(0, 0, 0, 0.04) 6px
		);
	}

	.dog-ear {
		position: absolute;
		top: 0;
		right: 0;
		width: 0;
		height: 0;
		border-style: solid;
		border-width: 0 40px 40px 0;
		border-color: transparent rgba(0, 0, 0, 0.08) transparent transparent;
		border-top-right-radius: 16px;
		pointer-events: none;
		z-index: 2;
	}

	.dog-ear::before {
		content: '';
		position: absolute;
		top: 2px;
		right: -38px;
		width: 0;
		height: 0;
		border-style: solid;
		border-width: 0 38px 38px 0;
		border-color: transparent rgba(255, 255, 255, 0.6) transparent transparent;
	}

	.ruled-lines {
		position: absolute;
		inset: 50px 20px 20px 20px;
		pointer-events: none;
		z-index: 1;
		opacity: 0.12;
	}

	.ruled-lines::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image: repeating-linear-gradient(
			transparent,
			transparent 27px,
			rgba(0, 0, 0, 0.15) 27px,
			rgba(0, 0, 0, 0.15) 28px
		);
	}

	.ruled-lines::after {
		content: '';
		position: absolute;
		left: 40px;
		top: 0;
		bottom: 0;
		width: 2px;
		background: rgba(220, 38, 38, 0.25);
	}

	.content-wrapper {
		position: relative;
		z-index: 3;
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.status-badge {
		padding: 6px;
		border-radius: 8px;
		backdrop-filter: blur(8px);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: all 0.2s ease;
	}

	.status-warn {
		background: rgba(251, 191, 36, 0.25);
		color: #b45309;
	}

	.status-critical {
		background: rgba(239, 68, 68, 0.25);
		color: #991b1b;
	}

	.action-buttons {
		display: flex;
		gap: 4px;
		opacity: 0;
		transform: translateY(-4px);
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.note-card:hover .action-buttons {
		opacity: 1;
		transform: translateY(0);
	}

	.action-btn {
		padding: 8px;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.6);
		border: 1px solid rgba(0, 0, 0, 0.08);
		color: rgba(0, 0, 0, 0.6);
		transition: all 0.2s ease;
		cursor: pointer;
		backdrop-filter: blur(8px);
	}

	.action-btn:hover {
		transform: scale(1.1) rotate(-5deg);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.action-btn:active {
		transform: scale(0.95) rotate(-2deg);
	}

	.edit-btn:hover {
		background: rgba(59, 130, 246, 0.2);
		border-color: rgba(59, 130, 246, 0.3);
		color: #1e40af;
	}

	.view-btn:hover {
		background: rgba(34, 197, 94, 0.2);
		border-color: rgba(34, 197, 94, 0.3);
		color: #15803d;
	}

	.delete-btn:hover {
		background: rgba(239, 68, 68, 0.2);
		border-color: rgba(239, 68, 68, 0.3);
		color: #991b1b;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none !important;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(15, 23, 42, 0.85);
		backdrop-filter: blur(12px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 9999;
		padding: 1rem;
	}

	.modal-content {
		background: linear-gradient(135deg, #fefce8 0%, #fef9c3 50%, #fef08a 100%);
		border-radius: 0;
		box-shadow:
			0 25px 50px -12px rgba(0, 0, 0, 0.4),
			0 10px 25px rgba(0, 0, 0, 0.2),
			inset 0 1px 0 rgba(255, 255, 255, 0.9),
			inset 0 -1px 2px rgba(0, 0, 0, 0.05),
			inset 2px 0 2px rgba(0, 0, 0, 0.02),
			inset -2px 0 2px rgba(0, 0, 0, 0.02);
		max-width: 700px;
		width: 100%;
		max-height: 85vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		position: relative;
		border: 2px solid rgba(234, 179, 8, 0.3);
	}

	.modal-content::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 512 512' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.95' numOctaves='5' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.08'/%3E%3C/svg%3E");
		pointer-events: none;
		opacity: 0.7;
		mix-blend-mode: overlay;
		z-index: 1;
	}

	.modal-content::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 8px;
		background: repeating-linear-gradient(
			90deg,
			transparent 0px,
			transparent 4px,
			rgba(0, 0, 0, 0.04) 4px,
			rgba(0, 0, 0, 0.04) 8px
		);
		z-index: 10;
		pointer-events: none;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem 2rem;
		border-bottom: 2px dashed rgba(0, 0, 0, 0.15);
		position: relative;
		z-index: 2;
		background: transparent;
	}

	.modal-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 8px;
		background: rgba(34, 197, 94, 0.15);
		color: #15803d;
		border: 1px solid rgba(34, 197, 94, 0.3);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.modal-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: rgba(0, 0, 0, 0.9);
		letter-spacing: -0.02em;
		font-family: 'Georgia', 'Times New Roman', serif;
		text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
	}

	.modal-close-btn {
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.6);
		border: 1px solid rgba(0, 0, 0, 0.1);
		color: rgba(0, 0, 0, 0.6);
		transition: all 0.2s ease;
		cursor: pointer;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
	}

	.modal-close-btn:hover {
		background: rgba(239, 68, 68, 0.2);
		border-color: rgba(239, 68, 68, 0.3);
		color: #991b1b;
		transform: rotate(90deg) scale(1.1);
		box-shadow: 0 4px 8px rgba(239, 68, 68, 0.2);
	}

	.modal-close-btn:active {
		transform: rotate(90deg) scale(0.95);
	}

	.modal-body {
		flex: 1;
		overflow-y: auto;
		padding: 2rem;
		position: relative;
		z-index: 2;
		background: transparent;
	}

	.modal-body::-webkit-scrollbar {
		width: 100px;
	}

	.modal-body::-webkit-scrollbar-track {
		background: rgba(0, 0, 0, 0.05);
		border-radius: 5px;
		margin: 8px 0;
	}

	.modal-body::-webkit-scrollbar-thumb {
		background: rgba(234, 179, 8, 0.3);
		border-radius: 5px;
		border: 2px solid transparent;
		background-clip: padding-box;
	}

	.modal-body::-webkit-scrollbar-thumb:hover {
		background: rgba(234, 179, 8, 0.5);
		background-clip: padding-box;
	}

	.modal-paper {
		background: transparent;
		border-radius: 0;
		padding: 0;
		box-shadow: none;
		position: relative;
		overflow: visible;
		border: none;
	}

	.modal-paper::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image: repeating-linear-gradient(
			transparent,
			transparent 27px,
			rgba(0, 0, 0, 0.08) 27px,
			rgba(0, 0, 0, 0.08) 28px
		);
		pointer-events: none;
		z-index: 0;
		opacity: 0.5;
	}

	.modal-paper::after {
		content: '';
		position: absolute;
		left: 50px;
		top: 0;
		bottom: 0;
		width: 2px;
		background: rgba(220, 38, 38, 0.2);
		pointer-events: none;
		z-index: 0;
	}

	.modal-text {
		position: relative;
		z-index: 1;
		color: rgba(0, 0, 0, 0.85);
		font-size: 1rem;
		line-height: 1.75;
		white-space: pre-wrap;
		word-wrap: break-word;
		font-weight: 500;
		padding-left: 60px;
		min-height: 200px;
	}

	.modal-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem 2rem;
		border-top: 2px dashed rgba(0, 0, 0, 0.15);
		background: transparent;
		position: relative;
		z-index: 2;
	}

	.modal-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: rgba(0, 0, 0, 0.5);
		font-size: 0.75rem;
		font-family: 'Courier New', monospace;
	}

	.modal-done-btn {
		padding: 0.75rem 2rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.7);
		border: 1px solid rgba(0, 0, 0, 0.15);
		color: rgba(0, 0, 0, 0.7);
		font-weight: 600;
		font-size: 0.95rem;
		transition: all 0.2s ease;
		cursor: pointer;
		position: relative;
		overflow: hidden;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
		font-family: 'Georgia', serif;
	}

	.modal-done-btn::before {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, transparent, rgba(0, 0, 0, 0.05), transparent);
		transform: translateX(-100%);
		transition: transform 0.5s ease;
	}

	.modal-done-btn:hover::before {
		transform: translateX(100%);
	}

	.modal-done-btn:hover {
		background: rgba(34, 197, 94, 0.2);
		border-color: rgba(34, 197, 94, 0.3);
		color: #15803d;
		transform: scale(1.05);
		box-shadow: 0 4px 12px rgba(34, 197, 94, 0.2);
	}

	.modal-done-btn:active {
		transform: scale(0.98);
	}

	.note-title {
		font-size: 1.25rem;
		font-weight: 700;
		margin-bottom: 12px;
		line-height: 1.3;
		color: rgba(0, 0, 0, 0.9);
		word-wrap: break-word;
		text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
		font-family: 'Georgia', 'Times New Roman', serif;
		letter-spacing: -0.01em;
	}

	.note-content {
		font-size: 0.9rem;
		line-height: 1.7;
		white-space: pre-wrap;
		word-wrap: break-word;
		color: rgba(0, 0, 0, 0.8);
		font-weight: 500;
	}

	.note-timestamp {
		margin-top: 16px;
		padding-top: 12px;
		border-top: 1px dashed rgba(0, 0, 0, 0.15);
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.7rem;
		color: rgba(0, 0, 0, 0.5);
		letter-spacing: 0.02em;
	}
</style>