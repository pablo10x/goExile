<script lang="ts">
    import { notifications } from '$lib/stores';
    import Toast from './Toast.svelte';
</script>

<div 
    class="fixed bottom-0 right-0 z-[100] flex flex-col gap-2 p-6 pointer-events-none w-full max-w-sm sm:max-w-md"
    aria-live="assertive"
>
    {#each $notifications as notification (notification.id)}
        <Toast {notification} />
    {/each}
</div>