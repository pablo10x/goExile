<script lang="ts">
	import { onMount } from 'svelte';
	import SystemTopology from '$lib/components/SystemTopology.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	let isLoaded = $state(false);

	onMount(() => {
		setTimeout(() => {
			isLoaded = true;
		}, 100);
	});
</script>

<div class="flex justify-between items-center mb-6">
	<div
		class="transform transition-all duration-700 {isLoaded
			? 'translate-x-0 opacity-100'
			: '-translate-x-4 opacity-0'}"
	>
		<h1
			class="text-3xl font-bold text-slate-50 bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent animate-pulse"
		>
			System Topology
		</h1>
		<p class="text-slate-400 text-sm mt-1">Real-time infrastructure visualization</p>
	</div>

	<div class="flex items-center gap-4">
		<NotificationBell />
		<div
			class="text-slate-500 text-sm transform transition-all duration-700 delay-100 {isLoaded
				? 'translate-x-0 opacity-100'
				: 'translate-x-4 opacity-0'}"
		>
			{new Date().toLocaleDateString()}
		</div>
	</div>
</div>

<!-- System Topology Animation (Hero) -->
<div
	class="mb-8 h-[calc(100vh-200px)] transform transition-all duration-1000 {isLoaded
		? 'translate-y-0 opacity-100'
		: 'translate-y-8 opacity-0'}"
>
	<SystemTopology />
</div>
