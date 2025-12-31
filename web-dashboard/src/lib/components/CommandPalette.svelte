<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { 
		Search, Command, Terminal, Gauge, Activity, 
		Shield, Settings, FileText, Database, Users, 
		RefreshCw, Trash2, Cpu, Moon, Sun, X
	} from 'lucide-svelte';
	import { theme, notifications, isAuthenticated } from '$lib/stores';
	import Icon from './theme/Icon.svelte';

	let { isOpen = $bindable(false) } = $props<{ isOpen: boolean }>();

	let query = $state('');
	let selectedIndex = $state(0);
	let inputElement: HTMLInputElement;

	const actions = [
		{ id: 'dash', label: 'Go to Dashboard', icon: Gauge, category: 'Navigation', shortcut: 'G D', action: () => goto('/dashboard') },
		{ id: 'perf', label: 'System Performance', icon: Activity, category: 'Navigation', shortcut: 'G P', action: () => goto('/performance') },
		{ id: 'logs', label: 'Kernel Logs', icon: FileText, category: 'Navigation', shortcut: 'G L', action: () => goto('/logs') },
		{ id: 'nodes', label: 'Node Management', icon: Cpu, category: 'Navigation', shortcut: 'G N', action: () => goto('/server') },
		{ id: 'db', label: 'Database Explorer', icon: Database, category: 'Navigation', shortcut: 'G B', action: () => goto('/database') },
		{ id: 'theme', label: 'Theme Calibration', icon: Settings, category: 'Navigation', shortcut: 'G T', action: () => goto('/config/theme') },
		
		{ id: 'restart', label: 'Restart Master Server', icon: RefreshCw, category: 'System', shortcut: 'S R', action: () => triggerRestart() },
		{ id: 'theme_toggle', label: 'Toggle Dark/Light Mode', icon: Moon, category: 'System', shortcut: 'T M', action: () => theme.update(t => t === 'dark' ? 'light' : 'dark') },
		
		{ id: 'help', label: 'Documentation', icon: Shield, category: 'Support', shortcut: '?', action: () => window.open('https://github.com', '_blank') }
	];

	let filteredActions = $derived(
		actions.filter(a => 
			a.label.toLowerCase().includes(query.toLowerCase()) || 
			a.category.toLowerCase().includes(query.toLowerCase())
		)
	);

	function triggerRestart() {
		if (confirm('Initiate system reboot sequence?')) {
			fetch('/api/restart', { method: 'POST' });
			notifications.add({ type: 'warning', message: 'REBOOT_SEQUENCE_INITIATED' });
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = (selectedIndex + 1) % filteredActions.length;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = (selectedIndex - 1 + filteredActions.length) % filteredActions.length;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (filteredActions[selectedIndex]) {
				filteredActions[selectedIndex].action();
				close();
			}
		} else if (e.key === 'Escape') {
			close();
		}
	}

	function close() {
		isOpen = false;
		query = '';
		selectedIndex = 0;
	}

	$effect(() => {
		if (isOpen) {
			tick().then(() => inputElement?.focus());
		}
	});
</script>

{#if isOpen}
	<div 
		class="fixed inset-0 z-[500] flex items-start justify-center pt-[15vh] px-4 bg-black/80 backdrop-blur-md font-jetbrains"
		transition:fade={{ duration: 150 }}
		onclick={close}
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		aria-labelledby="command-palette-title"
	>
		<div 
			class="w-full max-w-2xl bg-stone-950 border border-stone-800 shadow-[0_0_100px_rgba(0,0,0,1)] overflow-hidden industrial-frame"
			transition:scale={{ start: 0.98, duration: 200, easing: cubicOut }}
			onclick={e => e.stopPropagation()}
			role="document"
		>
			<!-- Status Header -->
			<div class="px-6 py-3 border-b border-stone-800 bg-stone-900/50 flex justify-between items-center">
				<div class="flex items-center gap-3">
					<div class="w-1.5 h-1.5 rounded-full bg-rust animate-pulse"></div>
					<span id="command-palette-title" class="text-[9px] font-black text-stone-500 uppercase tracking-[0.3em]">Smart_Uplink :: Command_Center</span>
				</div>
				<div class="flex items-center gap-4">
					<span class="text-[8px] text-stone-700 uppercase">Input_Mode: KBD_GLOBAL</span>
					<button onclick={close} class="text-stone-600 hover:text-white transition-all">
						<X class="w-4 h-4" />
					</button>
				</div>
			</div>

			<!-- Search Input -->
			<div class="relative group p-6 bg-black/40">
				<Search class="absolute left-10 top-1/2 -translate-y-1/2 w-5 h-5 text-stone-700 group-focus-within:text-rust transition-colors" />
				<input
					bind:this={inputElement}
					bind:value={query}
					type="text"
					placeholder="INITIATE_COMMAND_QUERY..."
					class="w-full bg-stone-900/50 border border-stone-800 pl-14 pr-6 py-5 text-lg font-black text-white placeholder:text-stone-800 outline-none focus:border-rust transition-all industrial-frame"
				/>
			</div>

			<!-- Action List -->
			<div class="max-h-[450px] overflow-y-auto p-4 space-y-1 custom-scrollbar bg-black/20">
				{#each filteredActions as action, i}
					<button
						onclick={() => { action.action(); close(); }}
						onmouseenter={() => selectedIndex = i}
						class="w-full flex items-center justify-between p-4 transition-all relative group/item {selectedIndex === i ? 'bg-rust text-white translate-x-1' : 'text-stone-500 hover:text-stone-300'}"
					>
						<div class="flex items-center gap-5">
							<div class="p-2 bg-black/40 border border-stone-800 group-hover/item:border-white/20">
								<action.icon class="w-5 h-5 {selectedIndex === i ? 'text-white' : 'text-stone-700'}" />
							</div>
							<div class="text-left">
								<span class="text-[8px] font-black uppercase opacity-40 block mb-0.5 tracking-widest">{action.category}</span>
								<span class="text-xs font-black uppercase tracking-widest">{action.label}</span>
							</div>
						</div>
						
						{#if action.shortcut}
							<div class="flex items-center gap-1.5 opacity-40">
								{#each action.shortcut.split(' ') as key}
									<kbd class="px-2 py-1 bg-black border border-stone-800 rounded text-[9px] font-black">{key}</kbd>
								{/each}
							</div>
						{/if}

						{#if selectedIndex === i}
							<div class="absolute left-0 top-0 w-1 h-full bg-white animate-pulse"></div>
						{/if}
					</button>
				{:else}
					<div class="py-20 text-center opacity-20">
						<Terminal class="w-12 h-12 mx-auto mb-4" />
						<p class="text-xs font-black uppercase tracking-[0.3em]">No_Matching_Commands_Identified</p>
					</div>
				{/each}
			</div>

			<!-- Footer Tips -->
			<div class="p-4 border-t border-stone-800 bg-stone-950/80 flex justify-between items-center">
				<div class="flex gap-6 text-[8px] font-black text-stone-600 uppercase tracking-widest">
					<span class="flex items-center gap-2"><kbd class="px-1.5 py-0.5 bg-black border border-stone-800">↑↓</kbd> Navigate</span>
					<span class="flex items-center gap-2"><kbd class="px-1.5 py-0.5 bg-black border border-stone-800">ENTER</kbd> Execute</span>
					<span class="flex items-center gap-2"><kbd class="px-1.5 py-0.5 bg-black border border-stone-800">ESC</kbd> Abort</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-1 h-1 bg-emerald-500 rounded-full animate-flicker"></div>
					<span class="text-[8px] text-stone-700 font-black uppercase tracking-widest">A11y_Optimized_Kernel</span>
				</div>
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
		background: #1a1a1a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}
</style>