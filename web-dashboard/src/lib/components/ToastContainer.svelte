<script lang="ts">
	import { notifications } from '$lib/stores';
	import Toast from '$lib/components/Toast.svelte';
	import { flip } from 'svelte/animate';
</script>

<div class="fixed bottom-4 right-4 z-[9999] w-full max-w-sm space-y-3 pointer-events-none">
	{#each $notifications as notification (notification.id)}
		<div class="pointer-events-auto" animate:flip={{ duration: 300 }}>
			<Toast {notification} onDismiss={notifications.remove} />
		</div>
	{/each}
</div>
