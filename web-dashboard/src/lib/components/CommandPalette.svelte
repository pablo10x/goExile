<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { Search, Terminal, X, Zap, Activity, Shield, Users, Database, FileText, Sliders, Gauge, Cpu } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { portal } from '../actions/portal';

	let { isOpen = $bindable(false) } = $props<{ isOpen: boolean }>();

	let query = $state('');
	let selectedIndex = $state(0);
	let inputElement = $state<HTMLInputElement>();

	const actions = [
		{ id: 'dash', label: 'Go to Dashboard', category: 'Navigation', icon: Gauge, action: () => goto('/dashboard'), shortcut: 'G D' },
		{ id: 'perf', label: 'Go to Performance', category: 'Navigation', icon: Activity, action: () => goto('/performance'), shortcut: 'G P' },
		{ id: 'nodes', label: 'Go to Nodes', category: 'Navigation', icon: Cpu, action: () => goto('/server'), shortcut: 'G N' },
		{ id: 'users', label: 'Go to Users', category: 'Navigation', icon: Users, action: () => goto('/users'), shortcut: 'G U' },
		{ id: 'db', label: 'Go to Database', category: 'Navigation', icon: Database, action: () => goto('/database'), shortcut: 'G B' },
		{ id: 'notes', label: 'Go to Notes', category: 'Navigation', icon: FileText, action: () => goto('/notes'), shortcut: 'G O' },
		{ id: 'config', label: 'Go to Settings', category: 'Navigation', icon: Sliders, action: () => goto('/config'), shortcut: 'G S' },
		{ id: 'fw', label: 'Go to Firewall', category: 'Navigation', icon: Shield, action: () => goto('/redeye'), shortcut: 'G F' },
	];

	let filteredActions = $derived(
		actions.filter(a => 
			a.label.toLowerCase().includes(query.toLowerCase()) || 
			a.category.toLowerCase().includes(query.toLowerCase())
		)
	);

	function close() {
		isOpen = false;
		query = '';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = (selectedIndex + 1) % filteredActions.length;
		}
		if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = (selectedIndex - 1 + filteredActions.length) % filteredActions.length;
		}
		if (e.key === 'Enter' && filteredActions[selectedIndex]) {
			filteredActions[selectedIndex].action();
			close();
		}
	}

	$effect(() => {
		if (isOpen) {
			selectedIndex = 0;
			tick().then(() => inputElement?.focus());
		}
	});
</script>

{#if isOpen}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div 
		use:portal
		class="fixed inset-0 z-[1000] flex items-start justify-center pt-[15vh] px-4 bg-black/60 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={close}
	>
		<div 
			class="w-full max-w-2xl bg-neutral-900 border border-neutral-800 shadow-2xl overflow-hidden rounded-xl"
			transition:scale={{ start: 0.98, duration: 200 }}
			onclick={e => e.stopPropagation()}
			role="dialog"
			aria-modal="true"
			aria-labelledby="command-palette-title"
		>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-neutral-800 bg-neutral-950 flex justify-between items-center">
				<div class="flex items-center gap-3">
					<Search class="w-4 h-4 text-indigo-400" />
					<span id="command-palette-title" class="text-[10px] font-bold text-neutral-400 uppercase tracking-widest">Quick Search</span>
				</div>
				<button onclick={close} class="text-neutral-500 hover:text-white transition-all">
					<X class="w-4 h-4" />
				</button>
			</div>

			<!-- Input -->
			<div class="p-4 bg-neutral-950">
				<input
					bind:this={inputElement}
					bind:value={query}
					onkeydown={handleKeydown}
					type="text"
					placeholder="Type a command or search..."
					class="w-full bg-neutral-900 border border-neutral-800 px-4 py-3 text-lg text-white placeholder:text-neutral-600 outline-none focus:border-indigo-500 transition-all rounded-lg"
				/>
			</div>

			<!-- Results -->
			<div class="max-h-[400px] overflow-y-auto p-2 space-y-1 custom-scrollbar">
				{#each filteredActions as action, i}
					<button
						onclick={() => {
							action.action();
							close();
						}}
						onmouseenter={() => (selectedIndex = i)}
						class="w-full flex items-center justify-between p-3 rounded-lg transition-all {selectedIndex === i ? 'bg-indigo-600 text-white' : 'text-neutral-400 hover:bg-neutral-800'}"
					>
						<div class="flex items-center gap-4">
							<div class="p-2 rounded-md {selectedIndex === i ? 'bg-indigo-500' : 'bg-neutral-800'}">
								<action.icon size={18} />
							</div>
							<div class="text-left">
								<div class="text-xs font-bold">{action.label}</div>
								<div class="text-[10px] opacity-60 uppercase tracking-tight">{action.category}</div>
							</div>
						</div>
						{#if action.shortcut}
							<div class="flex gap-1">
								{#each action.shortcut.split(' ') as key}
									<kbd class="px-1.5 py-0.5 bg-black/20 border border-white/10 rounded text-[9px] font-mono">{key}</kbd>
								{/each}
							</div>
						{/if}
					</button>
				{:else}
					<div class="py-12 text-center text-neutral-600">
						<Search class="w-8 h-8 mx-auto mb-3 opacity-20" />
						<p class="text-sm">No results found for "{query}"</p>
					</div>
				{/each}
			</div>

			<!-- Footer -->
			<div class="px-6 py-3 border-t border-neutral-800 bg-neutral-950/50 flex justify-between text-[10px] text-neutral-500 font-medium">
				<div class="flex gap-4">
					<span class="flex items-center gap-1.5"><kbd class="px-1 bg-neutral-800 border border-neutral-700 rounded text-[9px]">↑↓</kbd> to navigate</span>
					<span class="flex items-center gap-1.5"><kbd class="px-1 bg-neutral-800 border border-neutral-700 rounded text-[9px]">↵</kbd> to select</span>
				</div>
				<span class="flex items-center gap-1.5"><kbd class="px-1 bg-neutral-800 border border-neutral-700 rounded text-[9px]">esc</kbd> to close</span>
			</div>
		</div>
	</div>
{/if}
