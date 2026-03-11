<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { Bot, X, Send, Sparkles, Loader2 } from 'lucide-svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { notes, todos } from '$lib/stores.svelte';

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
			const response = await apiFetch('/api/ai/chat', {
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

				if (data.suggested_todo) {
					messages = [
						...messages,
						{ role: 'ai', text: `Suggested task: "${data.suggested_todo}"` }
					];
				}
			} else {
				messages = [
					...messages,
					{ role: 'ai', text: 'Service currently unavailable. Please try again later.' }
				];
			}
		} catch (e) {
			messages = [...messages, { role: 'ai', text: 'Network connection failed.' }];
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="fixed bottom-24 right-6 z-50 flex flex-col items-end gap-4 pointer-events-none">
	{#if isOpen}
		<div
			class="pointer-events-auto w-80 sm:w-96 bg-slate-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[600px] backdrop-blur-xl"
			transition:scale={{ duration: 300 }}
		>
			<!-- Header -->
			<div
				class="p-4 bg-gradient-to-r from-sky-600 to-teal-600 flex items-center justify-between"
			>
				<div class="flex items-center gap-2 text-white font-semibold">
					<Bot class="w-5 h-5" />
					<span class="text-sm uppercase tracking-wider">System Assistant</span>
				</div>
				<button
					onclick={toggle}
					class="text-white/80 hover:text-white transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Chat Area -->
			<div
				class="flex-1 p-4 overflow-y-auto space-y-4 bg-slate-950/50 min-h-[300px]"
			>
				{#each messages as msg}
					<div class="flex {msg.role === 'user' ? 'justify-end' : 'justify-start'}">
						<div
							class="max-w-[80%] rounded-2xl px-4 py-2 text-sm {msg.role === 'user'
								? 'bg-sky-600 text-white rounded-br-none'
								: 'bg-slate-800 text-slate-200 rounded-bl-none border border-white/5'}"
						>
							{msg.text}
						</div>
					</div>
				{/each}
				{#if isLoading}
					<div class="flex justify-start">
						<div
							class="bg-slate-800 rounded-2xl rounded-bl-none px-4 py-2 flex items-center gap-2 border border-white/5"
						>
							<Loader2 class="w-4 h-4 text-sky-400 animate-spin" />
							<span class="text-xs text-slate-400">Processing...</span>
						</div>
					</div>
				{/if}
			</div>

			<!-- Input -->
			<div class="p-3 bg-slate-900 border-t border-white/5">
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
						placeholder="Search or ask for assistance..."
						class="flex-1 bg-slate-950 border border-white/5 rounded-xl px-4 py-2 text-sm text-white focus:border-sky-500 outline-none transition-all"
					/>
					<button
						type="submit"
						disabled={isLoading || !input.trim()}
						class="p-2 bg-sky-600 text-white rounded-xl hover:bg-sky-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shadow-lg shadow-sky-900/20"
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
		class="pointer-events-auto w-14 h-14 bg-gradient-to-r from-sky-500 to-teal-500 rounded-full shadow-xl shadow-sky-900/40 flex items-center justify-center text-white hover:scale-110 transition-transform active:scale-95 group"
	>
		{#if isOpen}
			<X class="w-6 h-6" />
		{:else}
			<Bot class="w-7 h-7 group-hover:animate-bounce" />
			<div class="absolute -top-1 -right-1">
				<span class="relative flex h-3 w-3">
					<span
						class="animate-ping absolute inline-flex h-full w-full rounded-full bg-sky-400 opacity-75"
					></span>
					<span class="relative inline-flex rounded-full h-3 w-3 bg-sky-500"></span>
				</span>
			</div>
		{/if}
	</button>
</div>
