<script lang="ts">
	import { Bot, X, Send, Sparkles, Loader2 } from 'lucide-svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { notes, todos } from '$lib/stores';

	let isOpen = false;
	let messages: { role: 'user' | 'ai'; text: string }[] = [];
	let input = '';
	let isLoading = false;

	function toggle() {
		isOpen = !isOpen;
		if (isOpen && messages.length === 0) {
			messages = [
				{
					role: 'ai',
					text: 'Hello! I can help you organize your notes and tasks. What needs to be done?'
				}
			];
		}
	}

	async function sendMessage() {
		if (!input.trim() || isLoading) return;

		const userMsg = input.trim();
		messages = [...messages, { role: 'user', text: userMsg }];
		input = '';
		isLoading = true;

		try {
			// In a real app, we'd pass actual note content context here
			const response = await fetch('/api/ai/chat', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					message: userMsg,
					context: 'general' // Could be dynamic
				})
			});

			if (response.ok) {
				const data = await response.json();
				messages = [...messages, { role: 'ai', text: data.response }];

				// If AI suggested a todo, add it automatically or prompt (for now auto-add logic mock)
				if (data.suggested_todo) {
					// In a real app, maybe ask user first or show a special UI element
					messages = [
						...messages,
						{ role: 'ai', text: `I've suggested a new task: "${data.suggested_todo}"` }
					];
					// Trigger todo creation via event or store?
					// For simplicity in this component, we'll just dispatch an event or rely on the user to add it.
					// Let's dispatch custom event if needed, but for now just chat.
				}
			} else {
				messages = [
					...messages,
					{ role: 'ai', text: 'Sorry, I encountered an error connecting to the neural network.' }
				];
			}
		} catch (e) {
			messages = [...messages, { role: 'ai', text: 'Connection failed.' }];
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="fixed bottom-24 right-6 z-50 flex flex-col items-end gap-4 pointer-events-none">
	{#if isOpen}
		<div
			class="pointer-events-auto w-80 sm:w-96 bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[600px]"
			transition:scale={{ duration: 300 }}
		>
			<!-- Header -->
			<div
				class="p-4 bg-gradient-to-r from-violet-600 to-indigo-600 flex items-center justify-between"
			>
				<div class="flex items-center gap-2 text-slate-900 dark:text-white font-semibold">
					<Bot class="w-5 h-5" />
					<span>AI Assistant</span>
				</div>
				<button
					onclick={toggle}
					class="text-slate-900/80 dark:text-white/80 hover:text-slate-900 dark:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Chat Area -->
			<div
				class="flex-1 p-4 overflow-y-auto space-y-4 bg-white/95 dark:bg-slate-950/95 min-h-[300px]"
			>
				{#each messages as msg}
					<div class="flex {msg.role === 'user' ? 'justify-end' : 'justify-start'}">
						<div
							class="max-w-[80%] rounded-2xl px-4 py-2 text-sm {msg.role === 'user'
								? 'bg-blue-600 text-slate-900 dark:text-white rounded-br-none'
								: 'bg-slate-800 text-slate-800 dark:text-slate-200 rounded-bl-none'}"
						>
							{msg.text}
						</div>
					</div>
				{/each}
				{#if isLoading}
					<div class="flex justify-start">
						<div class="bg-slate-800 rounded-2xl rounded-bl-none px-4 py-2 flex items-center gap-2">
							<Loader2 class="w-4 h-4 text-violet-400 animate-spin" />
							<span class="text-xs text-text-dim dark:text-text-dim">Thinking...</span>
						</div>
					</div>
				{/if}
			</div>

			<!-- Input -->
			<div class="p-3 bg-slate-900 border-t border-slate-200 dark:border-slate-800">
				<form
					class="flex items-center gap-2"
					onsubmit={(e) => {
						e.preventDefault();
						sendMessage();
					}}
				>
					<input
						type="text"
						bind:value={input}
						placeholder="Ask AI to organize..."
						class="flex-1 bg-slate-800 border-none rounded-xl px-4 py-2 text-sm text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none"
					/>
					<button
						type="submit"
						disabled={isLoading || !input.trim()}
						class="p-2 bg-violet-600 text-slate-900 dark:text-white rounded-xl hover:bg-violet-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						<Send class="w-4 h-4" />
					</button>
				</form>
			</div>
		</div>
	{/if}

	<!-- FAB -->
	<button
		onclick={toggle}
		class="pointer-events-auto w-14 h-14 bg-gradient-to-r from-violet-600 to-indigo-600 rounded-full shadow-lg shadow-violet-900/40 flex items-center justify-center text-slate-900 dark:text-white hover:scale-110 transition-transform active:scale-95 group"
	>
		{#if isOpen}
			<X class="w-6 h-6" />
		{:else}
			<Bot class="w-7 h-7 group-hover:animate-bounce" />
			<div class="absolute -top-1 -right-1">
				<span class="relative flex h-3 w-3">
					<span
						class="animate-ping absolute inline-flex h-full w-full rounded-full bg-violet-400 opacity-75"
					></span>
					<span class="relative inline-flex rounded-full h-3 w-3 bg-violet-500"></span>
				</span>
			</div>
		{/if}
	</button>
</div>
