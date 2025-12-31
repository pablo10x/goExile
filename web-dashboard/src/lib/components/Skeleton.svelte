<script lang="ts">
	let { width = '100%', height = '1rem', class: className = '' } = $props<{
		width?: string;
		height?: string;
		class?: string;
	}>();
</script>

<div 
	class="skeleton-frame {className}" 
	style="width: {width}; height: {height};"
>
	<div class="skeleton-shimmer"></div>
</div>

<style>
	@reference "../../app.css";

	.skeleton-frame {
		@apply relative overflow-hidden bg-stone-900/50 border border-stone-800/30;
		/* Use containment to isolate paint */
		contain: paint layout;
	}

	.skeleton-shimmer {
		@apply absolute inset-0;
		background: linear-gradient(
			90deg, 
			transparent 0%, 
			rgba(249, 115, 22, 0.05) 50%, 
			transparent 100%
		);
		animation: shimmer 1.5s infinite linear;
		will-change: transform;
	}

	@keyframes shimmer {
		0% { transform: translateX(-100%); }
		100% { transform: translateX(100%); }
	}
</style>